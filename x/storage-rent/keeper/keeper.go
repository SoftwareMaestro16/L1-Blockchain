package keeper

import (
	"context"
	"errors"
	"math"

	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/storage-rent/types"
)

const storageRentReserveModule = "feecollector_storage_rent_reserve"

var storageRentBaseDenom = "naet"

// BankKeeper defines the subset of bank functionality needed by the storage rent keeper.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	RentParams	types.StorageRentParams
	State		types.StorageRentState
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
	runtimeCtx	context.Context
	bankKeeper	BankKeeper
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func (k Keeper) WithBankKeeper(bk BankKeeper) Keeper {
	k.bankKeeper = bk
	return k
}

func (k Keeper) StorageRentRatePerByteBlock() uint64 {
	return k.genesis.RentParams.RentRatePerByteBlock
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		prototype.DefaultParams(),
		RentParams:	types.DefaultStorageRentParams(),
		State:		types.EmptyStorageRentState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("storage rent prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.RentParams)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	k.runtimeCtx = ctx
	if k.storeService == nil {
		return nil
	}
	return prefixgenesis.Save(ctx, k.storeService, genesisKey, k.genesis)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	gs, _, err := prefixgenesis.Load(ctx, k.storeService, genesisKey, DefaultGenesis())
	if err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k Keeper) SystemRentStatus() (types.SystemRentResult, error) {
	if err := k.genesis.State.SystemReserve.Validate(); err != nil {
		return types.SystemRentResult{}, err
	}
	return k.genesis.State.SystemReserve.Evaluate(), nil
}

func (k *Keeper) UpdateSystemRentReserve(authority string, reserve types.SystemRentReserve, height uint64) (types.SystemRentResult, error) {
	if err := k.requireAuthority(authority); err != nil {
		return types.SystemRentResult{}, err
	}
	if height == 0 {
		return types.SystemRentResult{}, errors.New("storage rent system reserve height must be positive")
	}
	if err := reserve.Validate(); err != nil {
		return types.SystemRentResult{}, err
	}
	result := reserve.Evaluate()
	reserve = reserve.WithResult(height, result)
	next := cloneGenesis(k.genesis)
	next.State.SystemReserve = reserve
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.SystemRentResult{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.SystemRentResult{}, err
	}
	return result, nil
}

func (k *Keeper) TrackContractStorageUsage(authority, contractAddress, actorID string, storageBytes, observedStorageBytes, height uint64) (types.ContractRentRecord, error) {
	if err := k.requireAuthority(authority); err != nil {
		return types.ContractRentRecord{}, err
	}
	if height == 0 {
		return types.ContractRentRecord{}, errors.New("storage rent accounting height must be positive")
	}
	if storageBytes != observedStorageBytes {
		return types.ContractRentRecord{}, errors.New("storage rent usage must match actor registry/storage accounting")
	}
	next := cloneGenesis(k.genesis)
	if index, contract, found := contractIndex(next.State.Contracts, contractAddress); found {
		accrued, _, err := types.AccrueRent(contract, next.RentParams, height)
		if err != nil {
			return types.ContractRentRecord{}, err
		}
		accrued.StorageBytes = storageBytes
		next.State.Contracts[index] = accrued.Normalize()
	} else {
		contract := types.ContractRentRecord{
			ContractAddress:	contractAddress,
			ActorID:		actorID,
			StorageBytes:		storageBytes,
			LastChargedHeight:	height,
			Status:			types.ContractStatusActive,
			ArchivalProofRoot:	types.DefaultProofRoot,
			Exempt:			isExempt(next.State.Exemptions, contractAddress),
		}
		if err := contract.Validate(); err != nil {
			return types.ContractRentRecord{}, err
		}
		next.State.Contracts = append(next.State.Contracts, contract.Normalize())
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ContractRentRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ContractRentRecord{}, err
	}
	_, contract, _ := contractIndex(k.genesis.State.Contracts, contractAddress)
	return contract, nil
}

func (k *Keeper) PayStorageRent(ctx context.Context, msg types.MsgPayStorageRent) (types.ContractRentRecord, types.RentDistributionRecord, error) {
	if msg.Payer == "" {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent payer is required")
	}
	record, distribution, err := k.pay(msg.ContractAddress, msg.Amount, msg.Height, false)
	if err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	if err := k.collectRentPayment(ctx, msg.Payer, msg.Amount); err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	return record, distribution, nil
}

func (k *Keeper) WithdrawExcessRent(msg types.MsgWithdrawExcessRent) (types.ContractRentRecord, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ContractRentRecord{}, err
	}
	if msg.Amount == 0 {
		return types.ContractRentRecord{}, errors.New("storage rent withdraw amount must be positive")
	}
	index, contract, found := contractIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractRentRecord{}, errors.New("storage rent contract not found")
	}
	if msg.Height == 0 || msg.Height < contract.LastChargedHeight {
		return types.ContractRentRecord{}, errors.New("storage rent withdraw height must be monotonic")
	}
	contract, _, err := types.AccrueRent(contract, k.genesis.RentParams, msg.Height)
	if err != nil {
		return types.ContractRentRecord{}, err
	}
	if contract.RentDebt != 0 {
		return types.ContractRentRecord{}, errors.New("storage rent cannot withdraw while debt is outstanding")
	}
	if msg.Amount > contract.PrepaidRentBalance {
		return types.ContractRentRecord{}, errors.New("storage rent withdraw exceeds prepaid balance")
	}
	contract.PrepaidRentBalance -= msg.Amount
	next := cloneGenesis(k.genesis)
	next.State.Contracts[index] = contract.Normalize()
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ContractRentRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ContractRentRecord{}, err
	}
	return contract.Normalize(), nil
}

func (k *Keeper) FreezeExpiredContract(msg types.MsgFreezeExpiredContract) (types.ContractRentRecord, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ContractRentRecord{}, err
	}
	if msg.Height == 0 {
		return types.ContractRentRecord{}, errors.New("storage rent freeze height must be positive")
	}
	index, contract, found := contractIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractRentRecord{}, errors.New("storage rent contract not found")
	}
	contract, _, err := types.AccrueRent(contract, k.genesis.RentParams, msg.Height)
	if err != nil {
		return types.ContractRentRecord{}, err
	}
	if contract.RentDebt == 0 {
		return types.ContractRentRecord{}, errors.New("storage rent contract is not expired")
	}
	if contract.Status != types.ContractStatusActive {
		return types.ContractRentRecord{}, errors.New("storage rent only active contract can be frozen")
	}
	if msg.Height > math.MaxUint64-k.genesis.RentParams.RetentionBlocks {
		return types.ContractRentRecord{}, errors.New("storage rent deletion eligibility height overflow")
	}
	contract.Status = types.ContractStatusFrozen
	contract.FreezeHeight = msg.Height
	contract.DeletionEligibilityHeight = msg.Height + k.genesis.RentParams.RetentionBlocks
	next := cloneGenesis(k.genesis)
	next.State.Contracts[index] = contract.Normalize()
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ContractRentRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ContractRentRecord{}, err
	}
	return contract.Normalize(), nil
}

func (k *Keeper) UnfreezeContract(ctx context.Context, msg types.MsgUnfreezeContract) (types.ContractRentRecord, types.RentDistributionRecord, error) {
	if msg.Payer == "" {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent payer is required")
	}
	if msg.Height == 0 {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent unfreeze height must be positive")
	}
	index, contract, found := contractIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent contract not found")
	}
	if contract.Status != types.ContractStatusFrozen {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent only frozen contract can be unfrozen")
	}
	contract, _, err := types.AccrueRent(contract, k.genesis.RentParams, msg.Height)
	if err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	required, err := types.RequiredUnfreezePayment(contract, k.genesis.RentParams)
	if err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	if msg.Amount < required {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent unfreeze requires full debt plus configured buffer")
	}
	contract, err = applyPaymentChecked(contract, msg.Amount)
	if err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	contract.Status = types.ContractStatusActive
	contract.FreezeHeight = 0
	contract.DeletionEligibilityHeight = 0
	distribution := types.BuildDistribution(contract.ContractAddress, msg.Height, msg.Amount, k.genesis.RentParams)
	next := cloneGenesis(k.genesis)
	next.State.Contracts[index] = contract.Normalize()
	next.State.Distributions = append(next.State.Distributions, distribution)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	if err := k.collectRentPayment(ctx, msg.Payer, msg.Amount); err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	return contract.Normalize(), distribution, nil
}

func (k *Keeper) DeleteExpiredContract(msg types.MsgDeleteExpiredContract) (types.ContractRentRecord, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ContractRentRecord{}, err
	}
	if msg.Height == 0 {
		return types.ContractRentRecord{}, errors.New("storage rent delete height must be positive")
	}
	index, contract, found := contractIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractRentRecord{}, errors.New("storage rent contract not found")
	}
	if contract.Status != types.ContractStatusFrozen {
		return types.ContractRentRecord{}, errors.New("storage rent only frozen contract can be deleted")
	}
	if msg.Height < contract.DeletionEligibilityHeight {
		return types.ContractRentRecord{}, errors.New("storage rent deletion cannot happen before retention period")
	}
	contract.Status = types.ContractStatusDeleted
	if msg.ArchivalProofRoot != "" {
		contract.ArchivalProofRoot = msg.ArchivalProofRoot
	}
	next := cloneGenesis(k.genesis)
	next.State.Contracts[index] = contract.Normalize()
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ContractRentRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ContractRentRecord{}, err
	}
	return contract.Normalize(), nil
}

func (k *Keeper) UpdateStorageRentParams(msg types.MsgUpdateStorageRentParams) error {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	next.RentParams = msg.Params
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k Keeper) ContractRent(contractAddress string) (types.ContractRentRecord, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.ContractRentRecord{}, false, err
	}
	_, contract, found := contractIndex(k.genesis.State.Contracts, contractAddress)
	return contract, found, nil
}

func (k Keeper) RentDebt(contractAddress string) (uint64, bool, error) {
	contract, found, err := k.ContractRent(contractAddress)
	return contract.RentDebt, found, err
}

func (k Keeper) FrozenContracts() ([]types.ContractRentRecord, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	out := make([]types.ContractRentRecord, 0)
	for _, contract := range k.genesis.State.Export().Contracts {
		if contract.Status == types.ContractStatusFrozen {
			out = append(out, contract)
		}
	}
	types.SortContracts(out)
	return out, nil
}

func (k Keeper) DeletionQueue() ([]types.ContractRentRecord, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	out := make([]types.ContractRentRecord, 0)
	for _, contract := range k.genesis.State.Export().Contracts {
		if contract.Status == types.ContractStatusFrozen && contract.DeletionEligibilityHeight > 0 {
			out = append(out, contract)
		}
	}
	types.SortContracts(out)
	return out, nil
}

func (k Keeper) StorageRentParams() (types.StorageRentParams, error) {
	if err := k.genesis.RentParams.Validate(); err != nil {
		return types.StorageRentParams{}, err
	}
	return k.genesis.RentParams, nil
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	return m.keeper.ExportGenesis().Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k *Keeper) pay(contractAddress string, amount, height uint64, requireFrozen bool) (types.ContractRentRecord, types.RentDistributionRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	if amount == 0 {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent payment amount must be positive")
	}
	if height == 0 {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent payment height must be positive")
	}
	index, contract, found := contractIndex(k.genesis.State.Contracts, contractAddress)
	if !found {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent contract not found")
	}
	if requireFrozen && contract.Status != types.ContractStatusFrozen {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, errors.New("storage rent contract is not frozen")
	}
	contract, _, err := types.AccrueRent(contract, k.genesis.RentParams, height)
	if err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	contract, err = applyPaymentChecked(contract, amount)
	if err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	distribution := types.BuildDistribution(contract.ContractAddress, height, amount, k.genesis.RentParams)
	next := cloneGenesis(k.genesis)
	next.State.Contracts[index] = contract.Normalize()
	next.State.Distributions = append(next.State.Distributions, distribution)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ContractRentRecord{}, types.RentDistributionRecord{}, err
	}
	return contract.Normalize(), distribution, nil
}

func (k *Keeper) saveGenesis(next GenesisState) error {
	next = cloneGenesis(next)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return prefixgenesis.Save(k.runtimeCtx, k.storeService, genesisKey, next)
}

func (k Keeper) requireAuthority(authority string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return k.genesis.Params.Authorize(authority)
}

func contractIndex(contracts []types.ContractRentRecord, contractAddress string) (int, types.ContractRentRecord, bool) {
	for i, contract := range contracts {
		contract = contract.Normalize()
		if contract.ContractAddress == contractAddress {
			return i, contract, true
		}
	}
	return -1, types.ContractRentRecord{}, false
}

func isExempt(exemptions []types.RentExemption, account string) bool {
	for _, exemption := range exemptions {
		if exemption.Account == account {
			return true
		}
	}
	return false
}

func applyPaymentChecked(contract types.ContractRentRecord, amount uint64) (types.ContractRentRecord, error) {
	if amount <= contract.RentDebt {
		contract.RentDebt -= amount
		return contract, nil
	}
	excess := amount - contract.RentDebt
	contract.RentDebt = 0
	if contract.PrepaidRentBalance > math.MaxUint64-excess {
		return types.ContractRentRecord{}, errors.New("storage rent prepaid balance overflow")
	}
	contract.PrepaidRentBalance += excess
	return contract, nil
}

func (k *Keeper) collectRentPayment(ctx context.Context, payer string, amount uint64) error {
	if k.bankKeeper == nil {
		return nil
	}
	payerAddr, err := sdk.AccAddressFromBech32(payer)
	if err != nil {
		return err
	}
	coin := sdk.NewCoins(sdk.NewCoin(storageRentBaseDenom, sdkmath.NewInt(int64(amount))))
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, payerAddr, storageRentReserveModule, coin)
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
