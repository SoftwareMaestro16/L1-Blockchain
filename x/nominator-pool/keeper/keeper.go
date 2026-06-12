package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.Params
	State	types.State
}

type OperationCounters struct {
	PoolLookups			uint64
	DelegatorLookups		uint64
	DelegatorRewardUpdates		uint64
	ValidatorAllocationReads	uint64
	ProofQueries			uint64
}

const (
	accountStatusActive	= "active"
	accountStatusInactive	= "inactive"
	accountStatusFrozen	= "frozen"
)

type AccountStatusReader interface {
	AccountStatus(address string) (string, bool)
}

type poolIndexEntry struct {
	index		int
	delegator	map[string]int
}

type Keeper struct {
	genesis			GenesisState
	storeService		corestore.KVStoreService
	runtimeCtx		context.Context
	accountStatusReader	AccountStatusReader
	indexes			map[string]poolIndexEntry
	counters		OperationCounters
}

func NewKeeper() Keeper {
	k := Keeper{genesis: DefaultGenesis()}
	k.rebuildIndexes()
	return k
}

func NewKeeperWithAccountStatus(reader AccountStatusReader) Keeper {
	k := NewKeeper()
	k.accountStatusReader = reader
	return k
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	k := Keeper{genesis: DefaultGenesis(), storeService: storeService}
	k.rebuildIndexes()
	return k
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{Version: prototype.CurrentGenesisVersion, Params: params, State: types.State{}.Normalize(params)}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("nominator pool unsupported genesis version")
	}
	return gs.State.Validate(gs.Params)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	k.rebuildIndexes()
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	k.runtimeCtx = ctx
	k.rebuildIndexes()
	if k.storeService == nil {
		return nil
	}
	return k.writeGenesisState(ctx, k.genesis)
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

func (k *Keeper) saveGenesis(next GenesisState) error {
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(next)
	k.rebuildIndexes()
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return k.writeGenesisState(k.runtimeCtx, k.genesis)
}

func (k Keeper) writeGenesisState(ctx context.Context, gs GenesisState) error {
	if k.storeService == nil {
		return nil
	}
	if err := prefixgenesis.Save(ctx, k.storeService, genesisKey, cloneGenesis(gs)); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	for _, pool := range gs.State.Pools {
		bz, err := json.Marshal(pool)
		if err != nil {
			return err
		}
		if err := store.Set(types.PoolKey(pool.PoolID), bz); err != nil {
			return err
		}
	}
	for _, share := range gs.State.PoolShares {
		bz, err := json.Marshal(share)
		if err != nil {
			return err
		}
		if err := store.Set(types.PoolShareKey(share.PoolID, share.Owner), bz); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) OperationCounters() OperationCounters {
	return k.counters
}

func (k *Keeper) ResetOperationCounters() {
	k.counters = OperationCounters{}
}

func (k *Keeper) UpdateParams(msg types.MsgUpdateParams) (types.Params, error) {
	if msg.Height == 0 {
		return types.Params{}, errors.New("nominator pool params update height must be positive")
	}
	next := msg.Params
	if err := k.genesis.Params.ValidateParamsUpdate(msg.Authority, next); err != nil {
		return types.Params{}, err
	}
	next.Authority = k.genesis.Params.Authority
	updated := cloneGenesis(k.genesis)
	updated.Params = next
	if err := k.saveGenesis(updated); err != nil {
		return types.Params{}, err
	}
	return k.genesis.Params, nil
}

func (k *Keeper) UpdateStakingParams(msg types.MsgUpdateStakingParams) (types.Params, error) {
	return k.UpdateParams(types.MsgUpdateParams{
		Authority:	msg.Authority,
		Params:		msg.Params,
		Height:		msg.Height,
	})
}

func (k *Keeper) RegisterValidator(msg types.MsgRegisterValidator) (types.ValidatorRegistrationReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("validator registration signer", msg.SignerAddress); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if err := types.ValidateUserFacingAEAddress("validator registration validator", msg.ValidatorAddress); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if msg.Height == 0 {
		return types.ValidatorRegistrationReceipt{}, errors.New("validator registration height must be positive")
	}
	if err := k.ensureActiveWallet(msg.SignerAddress, "validator registration"); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if _, _, found := findValidator(k.genesis.State.Validators, msg.ValidatorAddress); found {
		return types.ValidatorRegistrationReceipt{}, errors.New("staking validator already registered")
	}
	mode := types.ValidatorFundingPoolBacked
	if msg.NominatorStake == 0 {
		mode = types.ValidatorFundingSolo
	}
	if err := k.genesis.Params.ValidateValidatorFunding(types.ValidatorFunding{Mode: mode, SelfStake: msg.SelfStake, NominatorStake: msg.NominatorStake}); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if err := k.genesis.Params.ValidateCommission(msg.CommissionBps, k.genesis.Params.DefaultValidatorCommissionBps, 0); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	validator := types.Validator{
		Address:		msg.ValidatorAddress,
		SelfStake:		msg.SelfStake,
		NominatorStake:		msg.NominatorStake,
		Status:			types.StateValidatorStatusActive,
		PerformanceScore:	types.MaxBasisPoints,
		CommissionBps:		msg.CommissionBps,
		SlashingRiskBps:	0,
		AllocationLimitBps:	k.genesis.Params.MaxPoolValidatorAllocationBps,
		UpdatedHeight:		msg.Height,
	}
	next := cloneGenesis(k.genesis)
	next.State.Validators = append(next.State.Validators, validator)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	return types.ValidatorRegistrationReceipt{
		Validator:	msg.ValidatorAddress,
		Status:		validator.Status,
		SelfStake:	validator.SelfStake,
		PoolStake:	validator.NominatorStake,
		TouchedKeys:	[]string{string(types.ValidatorKey(msg.ValidatorAddress))},
	}, nil
}

func (k *Keeper) UpdateValidator(msg types.MsgUpdateValidator) (types.ValidatorRegistrationReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("validator update signer", msg.SignerAddress); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if err := types.ValidateUserFacingAEAddress("validator update validator", msg.ValidatorAddress); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if msg.Height == 0 {
		return types.ValidatorRegistrationReceipt{}, errors.New("validator update height must be positive")
	}
	if msg.SignerAddress != msg.ValidatorAddress {
		return types.ValidatorRegistrationReceipt{}, errors.New("validator update signer must match validator address")
	}
	if err := k.ensureActiveWallet(msg.SignerAddress, "validator update"); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	idx, validator, found := findValidator(k.genesis.State.Validators, msg.ValidatorAddress)
	if !found {
		return types.ValidatorRegistrationReceipt{}, errors.New("staking validator not found")
	}
	if msg.SelfStake > 0 {
		validator.SelfStake = msg.SelfStake
	}
	if msg.NominatorStake > 0 || validator.NominatorStake > 0 {
		validator.NominatorStake = msg.NominatorStake
	}
	if msg.PerformanceScore > 0 {
		validator.PerformanceScore = msg.PerformanceScore
	}
	if msg.CommissionBps > 0 {
		dailyChange := validator.CommissionBps - msg.CommissionBps
		if msg.CommissionBps > validator.CommissionBps {
			dailyChange = msg.CommissionBps - validator.CommissionBps
		}
		if err := k.genesis.Params.ValidateCommission(msg.CommissionBps, validator.CommissionBps, dailyChange); err != nil {
			return types.ValidatorRegistrationReceipt{}, err
		}
		validator.CommissionBps = msg.CommissionBps
	}
	validator.SlashingRiskBps = msg.SlashingRiskBps
	if msg.AllocationLimitBps > 0 {
		validator.AllocationLimitBps = msg.AllocationLimitBps
	}
	if msg.Status != "" {
		validator.Status = msg.Status
	}
	validator.UpdatedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Validators[idx] = validator
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ValidatorRegistrationReceipt{}, err
	}
	return types.ValidatorRegistrationReceipt{
		Validator:	validator.Address,
		Status:		validator.Status,
		SelfStake:	validator.SelfStake,
		PoolStake:	validator.NominatorStake,
		TouchedKeys:	[]string{string(types.ValidatorKey(validator.Address))},
	}, nil
}

func (k *Keeper) rebuildIndexes() {
	k.indexes = make(map[string]poolIndexEntry, len(k.genesis.State.Pools))
	for poolIdx, pool := range k.genesis.State.Pools {
		entry := poolIndexEntry{
			index:		poolIdx,
			delegator:	make(map[string]int, len(pool.DelegatorShares)),
		}
		for delegatorIdx, share := range pool.DelegatorShares {
			entry.delegator[share.Delegator] = delegatorIdx
		}
		k.indexes[pool.PoolID] = entry
	}
}

func (k *Keeper) ensureIndexes() {
	if k.indexes == nil || len(k.indexes) != len(k.genesis.State.Pools) {
		k.rebuildIndexes()
	}
}

func (k *Keeper) lookupPool(poolID string) (int, types.NominatorPool, bool) {
	k.ensureIndexes()
	k.counters.PoolLookups++
	entry, found := k.indexes[poolID]
	if !found || entry.index < 0 || entry.index >= len(k.genesis.State.Pools) {
		return -1, types.NominatorPool{}, false
	}
	return entry.index, k.genesis.State.Pools[entry.index], true
}

func (k *Keeper) lookupDelegator(poolID string, delegator string) (int, types.DelegatorShare, bool) {
	k.ensureIndexes()
	k.counters.DelegatorLookups++
	entry, found := k.indexes[poolID]
	if !found {
		return -1, types.DelegatorShare{}, false
	}
	pool := k.genesis.State.Pools[entry.index]
	delegatorIdx, found := entry.delegator[delegator]
	if !found || delegatorIdx < 0 || delegatorIdx >= len(pool.DelegatorShares) {
		return -1, types.DelegatorShare{}, false
	}
	return delegatorIdx, pool.DelegatorShares[delegatorIdx], true
}

func (k *Keeper) CreateNominatorPool(msg types.MsgCreateNominatorPool) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.NominatorPool{}, errors.New("nominator pool creation height must be positive")
	}
	if types.IsJailedValidatorStatus(msg.ValidatorStatus) {
		return types.NominatorPool{}, errors.New("nominator pool cannot delegate to jailed validator")
	}
	if _, _, found := findPool(k.genesis.State.Pools, msg.PoolID); found {
		return types.NominatorPool{}, errors.New("nominator pool already exists")
	}
	pool := types.NominatorPool{
		PoolID:			msg.PoolID,
		PoolOperator:		msg.PoolOperator,
		ValidatorTarget:	msg.ValidatorTarget,
		PoolCommissionBps:	msg.PoolCommissionBps,
		Status:			types.PoolStatusActive,
	}
	if err := pool.Validate(k.genesis.Params); err != nil {
		return types.NominatorPool{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Pools = append(next.State.Pools, pool)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.NominatorPool{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.NominatorPool{}, err
	}
	return pool, nil
}

func (k *Keeper) CreateOfficialLiquidStakingPool(msg types.MsgCreateOfficialLiquidStakingPool) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.NominatorPool{}, errors.New("official liquid staking pool creation height must be positive")
	}
	if _, _, found := findPool(k.genesis.State.Pools, msg.PoolID); found {
		return types.NominatorPool{}, errors.New("official liquid staking pool already exists")
	}
	pool := types.NominatorPool{
		PoolID:			msg.PoolID,
		ContractAddressUser:	msg.ContractAddressUser,
		ContractAddressRaw:	msg.ContractAddressRaw,
		OfficialLiquidStaking:	true,
		PoolOperator:		msg.PoolOperator,
		PoolCommissionBps:	msg.PoolCommissionBps,
		Status:			types.PoolStatusActive,
	}
	if err := pool.Validate(k.genesis.Params); err != nil {
		return types.NominatorPool{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Pools = append(next.State.Pools, pool)
	next.State.LiquidStakingPools = append(next.State.LiquidStakingPools, types.LiquidStakingPool{
		PoolID:				msg.PoolID,
		ContractAddressUser:		msg.ContractAddressUser,
		ContractAddressRaw:		msg.ContractAddressRaw,
		ReceiptToken:			next.Params.PoolReceiptDenomOrCodeID,
		RentPayerPolicy:		types.RentPayerPolicyPoolReserve,
		Status:				types.PoolStatusActive,
		LastStorageChargeHeight:	msg.Height,
	})
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.NominatorPool{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.NominatorPool{}, err
	}
	return pool, nil
}

func (k *Keeper) DepositToPool(msg types.MsgDepositToPool) (types.DelegatorShare, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.DelegatorShare{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.DelegatorShare{}, errors.New("nominator pool deposit amount and height must be positive")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.DelegatorShare{}, errors.New("nominator pool not found")
	}
	if pool.Status != types.PoolStatusActive {
		return types.DelegatorShare{}, errors.New("nominator pool must be active for deposits")
	}
	shareAmount, err := types.SharesForDepositChecked(pool, msg.Amount)
	if err != nil {
		return types.DelegatorShare{}, err
	}
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, msg.Delegator)
	if found {
		delegator.PendingRewards = types.AccruedReward(delegator, pool.RewardIndex)
		delegator.Shares += shareAmount
		delegator.RewardIndexCheckpoint = pool.RewardIndex
		delegator.SlashIndexCheckpoint = pool.SlashIndex
		pool.DelegatorShares[delegatorIdx] = delegator
	} else {
		delegator = types.DelegatorShare{
			Delegator:		msg.Delegator,
			Shares:			shareAmount,
			RewardIndexCheckpoint:	pool.RewardIndex,
			SlashIndexCheckpoint:	pool.SlashIndex,
		}
		pool.DelegatorShares = append(pool.DelegatorShares, delegator)
	}
	pool.TotalShares += shareAmount
	pool.TotalBondedStake += msg.Amount
	pool.PendingDeposits = append(pool.PendingDeposits, types.PendingDeposit{Delegator: msg.Delegator, Amount: msg.Amount, Height: msg.Height})
	return k.savePool(idx, pool, delegator)
}

func (k *Keeper) DepositToOfficialLiquidStaking(msg types.MsgDepositToOfficialLiquidStaking) (types.DelegatorShare, error) {
	if err := types.ValidateOfficialLiquidStakingDeposit(msg, k.genesis.Params); err != nil {
		return types.DelegatorShare{}, err
	}
	rawUserAddress, err := types.RawAddressForUserAddress(msg.UserAddress)
	if err != nil {
		return types.DelegatorShare{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.DelegatorShare{}, errors.New("official liquid staking pool not found")
	}
	if !pool.OfficialLiquidStaking {
		return types.DelegatorShare{}, errors.New("pool is not an official liquid staking pool")
	}
	if pool.Status != types.PoolStatusActive {
		return types.DelegatorShare{}, errors.New("official liquid staking pool must be active for deposits")
	}
	shareAmount, err := types.SharesForDepositChecked(pool, msg.Amount)
	if err != nil {
		return types.DelegatorShare{}, err
	}
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, rawUserAddress)
	if found {
		delegator.PendingRewards = types.AccruedReward(delegator, pool.RewardIndex)
		delegator.Shares += shareAmount
		delegator.RewardIndexCheckpoint = pool.RewardIndex
		delegator.SlashIndexCheckpoint = pool.SlashIndex
		pool.DelegatorShares[delegatorIdx] = delegator
	} else {
		delegator = types.DelegatorShare{
			Delegator:		rawUserAddress,
			Shares:			shareAmount,
			RewardIndexCheckpoint:	pool.RewardIndex,
			SlashIndexCheckpoint:	pool.SlashIndex,
		}
		pool.DelegatorShares = append(pool.DelegatorShares, delegator)
	}
	pool.TotalShares += shareAmount
	pool.TotalBondedStake += msg.Amount
	pool.PendingDeposits = append(pool.PendingDeposits, types.PendingDeposit{Delegator: rawUserAddress, Amount: msg.Amount, Height: msg.Height})
	return k.savePool(idx, pool, delegator)
}

func (k *Keeper) DepositToStakingPool(msg types.MsgDepositToStakingPool) (types.StakingPoolDepositReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("staking pool depositor", msg.WalletAddress); err != nil {
		return types.StakingPoolDepositReceipt{}, err
	}
	if msg.ReservedRouting != "" {
		return types.StakingPoolDepositReceipt{}, errors.New("staking pool deposit must not include a routing field")
	}
	if err := k.ensureActiveWallet(msg.WalletAddress, "staking pool deposit"); err != nil {
		return types.StakingPoolDepositReceipt{}, err
	}

	poolID := msg.PoolID

	if poolID != "" {
		if _, err := addressing.Parse(poolID); err == nil {
			return types.StakingPoolDepositReceipt{}, errors.New("pool id must not be an address")
		} else {
		}
	}
	if msg.OfficialContract != "" {
		found := false
		resolvedID := ""

		for _, p := range k.genesis.State.Pools {
			if p.ContractAddressUser == msg.OfficialContract {
				resolvedID = p.PoolID
				found = true
				break
			}
		}

		if !found {
			for _, lp := range k.genesis.State.LiquidStakingPools {
				if lp.ContractAddressUser == msg.OfficialContract {
					resolvedID = lp.PoolID
					found = true
					break
				}
			}
		}
		if !found {
			return types.StakingPoolDepositReceipt{}, errors.New("official liquid staking pool not found")
		}

		if msg.PoolID != "" && msg.PoolID != resolvedID {
			return types.StakingPoolDepositReceipt{}, errors.New("pool id does not match official contract")
		}
		poolID = resolvedID
	}
	share, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	k.genesis.Params.Authority,
		PoolID:		poolID,
		UserAddress:	msg.WalletAddress,
		Amount:		msg.Amount,
		Height:		msg.Height,
	})
	if err != nil {
		return types.StakingPoolDepositReceipt{}, err
	}
	rawUserAddress, err := types.RawAddressForUserAddress(msg.WalletAddress)
	if err != nil {
		return types.StakingPoolDepositReceipt{}, err
	}
	_, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return types.StakingPoolDepositReceipt{}, errors.New("official liquid staking pool not found")
	}
	if err := k.upsertLiquidPoolAfterPoolMutation(pool, msg.Height); err != nil {
		return types.StakingPoolDepositReceipt{}, err
	}
	if err := k.upsertPoolShare(poolID, msg.WalletAddress, share, msg.Amount, msg.Height); err != nil {
		return types.StakingPoolDepositReceipt{}, err
	}
	return types.StakingPoolDepositReceipt{
		PoolID:				poolID,
		OwnerAddress:			msg.WalletAddress,
		PoolContractAddressUser:	pool.ContractAddressUser,
		ReceiptToken:			k.genesis.Params.PoolReceiptDenomOrCodeID,
		Amount:				msg.Amount,
		Shares:				share.Shares,
		Height:				msg.Height,
		InternalMetadata: types.PoolStateMetadata{
			OwnerRaw:		rawUserAddress,
			PoolContractAddressRaw:	pool.ContractAddressRaw,
			TouchedKeys: []string{
				string(types.PoolKey(poolID)),
				string(types.PoolShareKey(poolID, msg.WalletAddress)),
			},
		},
	}, nil
}

func (k *Keeper) DelegateUserToValidator(msg types.MsgDelegateToValidator) error {
	return types.ValidateDirectUserDelegation(msg, k.genesis.Params)
}

func (k *Keeper) InjectPooledStake(msg types.MsgInjectPooledStake) (types.NominatorPool, error) {
	if err := types.ValidateUserFacingAEAddress("pooled stake caller contract", msg.CallerContractUser); err != nil {
		return types.NominatorPool{}, err
	}
	if err := types.ValidateUserFacingAEAddress("pooled stake validator address", msg.ValidatorAddress); err != nil {
		return types.NominatorPool{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.NominatorPool{}, errors.New("pooled stake injection amount and height must be positive")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.NominatorPool{}, errors.New("official liquid staking pool not found")
	}
	if !pool.OfficialLiquidStaking || pool.ContractAddressUser != msg.CallerContractUser {
		return types.NominatorPool{}, errors.New("pooled stake injection requires official liquid staking contract")
	}
	if pool.Status != types.PoolStatusActive {
		return types.NominatorPool{}, errors.New("official liquid staking pool must be active for stake injection")
	}
	currentAllocated := totalAllocated(pool.Allocations)
	if msg.Amount > pool.TotalBondedStake-currentAllocated {
		return types.NominatorPool{}, errors.New("pooled stake injection exceeds unallocated pool stake")
	}
	allocationIdx, allocation, found := findAllocation(pool.Allocations, msg.ValidatorAddress)
	if found {
		allocation.Amount += msg.Amount
		allocation.Height = msg.Height
		pool.Allocations[allocationIdx] = allocation
	} else {
		pool.Allocations = append(pool.Allocations, types.PoolAllocation{
			ValidatorAddress:	msg.ValidatorAddress,
			Amount:			msg.Amount,
			Height:			msg.Height,
		})
	}
	savedPool, err := k.savePoolOnly(idx, pool)
	if err != nil {
		return types.NominatorPool{}, err
	}
	if savedPool.OfficialLiquidStaking {
		if err := k.upsertLiquidPoolAfterPoolMutation(savedPool, msg.Height); err != nil {
			return types.NominatorPool{}, err
		}
	}
	return savedPool, nil
}

func (k *Keeper) InjectPoolStake(msg types.MsgInjectPoolStake) (types.PoolRebalanceReceipt, error) {
	if len(msg.Allocations) == 0 {
		return types.PoolRebalanceReceipt{}, errors.New("pool stake injection requires allocations")
	}
	var updated types.NominatorPool
	for _, allocation := range types.SortAllocations(msg.Allocations) {
		pool, err := k.InjectPooledStake(types.MsgInjectPooledStake{
			CallerContractUser:	msg.CallerContractUser,
			PoolID:			msg.PoolID,
			ValidatorAddress:	allocation.ValidatorAddress,
			Amount:			allocation.Amount,
			Height:			msg.Height,
		})
		if err != nil {
			return types.PoolRebalanceReceipt{}, err
		}
		updated = pool
		if err := k.upsertPoolValidatorAllocation(msg.PoolID, allocation.ValidatorAddress, allocation.Amount, msg.Height); err != nil {
			return types.PoolRebalanceReceipt{}, err
		}
	}
	return k.poolAllocationReceipt(updated, 0, msg.Height)
}

func (k *Keeper) RebalancePoolAllocations(msg types.MsgRebalancePoolAllocations) (types.PoolRebalanceReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("pool rebalance caller contract", msg.CallerContractUser); err != nil {
		return types.PoolRebalanceReceipt{}, err
	}
	if msg.Epoch == 0 || msg.Height == 0 {
		return types.PoolRebalanceReceipt{}, errors.New("pool rebalance epoch and height must be positive")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PoolRebalanceReceipt{}, errors.New("official liquid staking pool not found")
	}
	if !pool.OfficialLiquidStaking || pool.ContractAddressUser != msg.CallerContractUser {
		return types.PoolRebalanceReceipt{}, errors.New("pool rebalance requires official liquid staking contract")
	}
	if pool.Status != types.PoolStatusActive {
		return types.PoolRebalanceReceipt{}, errors.New("official liquid staking pool must be active for rebalance")
	}
	weights, err := k.genesis.Params.AllocationWeights(msg.Candidates)
	if err != nil {
		return types.PoolRebalanceReceipt{}, err
	}
	nextAllocations := make([]types.PoolAllocation, 0, len(weights))
	allocated := uint64(0)
	lastPositive := -1
	for idx := range weights {
		if weights[idx].WeightBps > 0 {
			lastPositive = idx
		}
	}
	for idx, weight := range weights {
		if weight.WeightBps == 0 {
			continue
		}
		amount, err := types.MulDivUint64(pool.TotalBondedStake, uint64(weight.WeightBps), uint64(types.MaxBasisPoints))
		if err != nil {
			return types.PoolRebalanceReceipt{}, err
		}
		if idx == lastPositive {
			amount = pool.TotalBondedStake - allocated
		}
		allocated += amount
		nextAllocations = append(nextAllocations, types.PoolAllocation{
			ValidatorAddress:	weight.ValidatorAddress,
			Amount:			amount,
			Height:			msg.Height,
		})
		if err := k.upsertPoolValidatorAllocation(msg.PoolID, weight.ValidatorAddress, amount, msg.Height); err != nil {
			return types.PoolRebalanceReceipt{}, err
		}
	}
	pool.Allocations = types.SortAllocations(nextAllocations)
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	for allocationIdx := range next.State.PoolValidatorAllocations {
		if next.State.PoolValidatorAllocations[allocationIdx].PoolID == msg.PoolID {
			next.State.PoolValidatorAllocations[allocationIdx].UpdatedHeight = msg.Height
		}
	}
	if liquidIdx, liquid, found := findLiquidPool(next.State.LiquidStakingPools, msg.PoolID); found {
		if err := k.accrueOfficialPoolRent(&liquid, pool, msg.Height); err != nil {
			return types.PoolRebalanceReceipt{}, err
		}
		liquid.TotalActiveStake = allocated
		liquid.AllocationEpoch = msg.Epoch
		next.State.LiquidStakingPools[liquidIdx] = liquid
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PoolRebalanceReceipt{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.PoolRebalanceReceipt{}, err
	}
	return k.poolAllocationReceipt(pool, msg.Epoch, msg.Height)
}

func (k *Keeper) SetOfficialLiquidStakingContract(msg types.MsgSetOfficialLiquidStakingContract) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.NominatorPool{}, errors.New("official liquid staking contract update height must be positive")
	}
	if err := types.ValidateAddressPair("official liquid staking contract", msg.ContractAddressUser, msg.ContractAddressRaw); err != nil {
		return types.NominatorPool{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.NominatorPool{}, errors.New("official liquid staking pool not found")
	}
	pool.ContractAddressUser = msg.ContractAddressUser
	pool.ContractAddressRaw = msg.ContractAddressRaw
	pool.OfficialLiquidStaking = true
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	if liquidIdx, liquid, found := findLiquidPool(next.State.LiquidStakingPools, msg.PoolID); found {
		liquid.ContractAddressUser = msg.ContractAddressUser
		liquid.ContractAddressRaw = msg.ContractAddressRaw
		liquid.LastStorageChargeHeight = msg.Height
		next.State.LiquidStakingPools[liquidIdx] = liquid
	} else {
		next.State.LiquidStakingPools = append(next.State.LiquidStakingPools, types.LiquidStakingPool{
			PoolID:				msg.PoolID,
			ContractAddressUser:		msg.ContractAddressUser,
			ContractAddressRaw:		msg.ContractAddressRaw,
			ReceiptToken:			next.Params.PoolReceiptDenomOrCodeID,
			RentPayerPolicy:		types.RentPayerPolicyPoolReserve,
			Status:				pool.Status,
			LastStorageChargeHeight:	msg.Height,
		})
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.NominatorPool{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.NominatorPool{}, err
	}
	return pool, nil
}

func (k *Keeper) RequestPoolWithdrawal(msg types.MsgRequestPoolWithdrawal) (types.PendingWithdrawal, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingWithdrawal{}, err
	}
	if msg.Shares == 0 || msg.Height == 0 {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal shares and height must be positive")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool not found")
	}
	if _, _, found := findWithdrawal(pool.PendingWithdrawals, msg.WithdrawalID); found {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal already exists")
	}
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, msg.Delegator)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool delegator not found")
	}
	if msg.Shares > delegator.Shares || msg.Shares > pool.TotalShares {
		return types.PendingWithdrawal{}, errors.New("nominator pool cannot withdraw more than total stake")
	}
	reward := types.AccruedReward(delegator, pool.RewardIndex)
	amount := types.ShareValue(pool, msg.Shares)
	if amount == 0 || amount > pool.TotalBondedStake {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal amount exceeds bonded stake")
	}
	delegator.Shares -= msg.Shares
	delegator.PendingRewards = reward
	delegator.RewardIndexCheckpoint = pool.RewardIndex
	delegator.SlashIndexCheckpoint = pool.SlashIndex
	pool.TotalShares -= msg.Shares
	pool.TotalBondedStake -= amount
	if delegator.Shares == 0 && delegator.PendingRewards == 0 {
		pool.DelegatorShares = append(pool.DelegatorShares[:delegatorIdx], pool.DelegatorShares[delegatorIdx+1:]...)
	} else {
		pool.DelegatorShares[delegatorIdx] = delegator
	}
	withdrawal := types.PendingWithdrawal{
		WithdrawalID:	msg.WithdrawalID,
		Delegator:	msg.Delegator,
		Shares:		msg.Shares,
		Amount:		amount,
		RequestHeight:	msg.Height,
		CompleteHeight:	msg.Height + k.genesis.Params.UnbondingBlocks,
		Status:		types.WithdrawalStatusPending,
	}
	pool.PendingWithdrawals = append(pool.PendingWithdrawals, withdrawal)
	pool.UnbondingQueue = append(pool.UnbondingQueue, types.UnbondingEntry{
		WithdrawalID:	withdrawal.WithdrawalID,
		Delegator:	withdrawal.Delegator,
		Amount:		withdrawal.Amount,
		CompleteHeight:	withdrawal.CompleteHeight,
		Status:		withdrawal.Status,
	})
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PendingWithdrawal{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.PendingWithdrawal{}, err
	}
	return withdrawal, nil
}

func (k *Keeper) RequestPoolUnbond(msg types.MsgRequestPoolUnbond) (types.PoolUnbondReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("pool unbond owner", msg.OwnerAddress); err != nil {
		return types.PoolUnbondReceipt{}, err
	}
	if err := k.ensureActiveWallet(msg.OwnerAddress, "pool unbond request"); err != nil {
		return types.PoolUnbondReceipt{}, err
	}
	rawOwner, err := types.RawAddressForUserAddress(msg.OwnerAddress)
	if err != nil {
		return types.PoolUnbondReceipt{}, err
	}
	withdrawal, err := k.RequestPoolWithdrawal(types.MsgRequestPoolWithdrawal{
		Authority:	k.genesis.Params.Authority,
		PoolID:		msg.PoolID,
		WithdrawalID:	msg.RequestID,
		Delegator:	rawOwner,
		Shares:		msg.Shares,
		Height:		msg.Height,
	})
	if err != nil {
		return types.PoolUnbondReceipt{}, err
	}
	_, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PoolUnbondReceipt{}, errors.New("nominator pool not found")
	}
	if err := k.upsertLiquidPoolAfterPoolMutation(pool, msg.Height); err != nil {
		return types.PoolUnbondReceipt{}, err
	}
	if err := k.upsertPoolUnbonding(msg.PoolID, msg.OwnerAddress, withdrawal); err != nil {
		return types.PoolUnbondReceipt{}, err
	}
	if err := k.updatePoolShareAfterUnbond(msg.PoolID, msg.OwnerAddress, withdrawal, msg.Height); err != nil {
		return types.PoolUnbondReceipt{}, err
	}
	return types.PoolUnbondReceipt{
		PoolID:		msg.PoolID,
		OwnerAddress:	msg.OwnerAddress,
		RequestID:	msg.RequestID,
		Shares:		withdrawal.Shares,
		Amount:		withdrawal.Amount,
		RequestHeight:	withdrawal.RequestHeight,
		CompleteHeight:	withdrawal.CompleteHeight,
		InternalMetadata: types.PoolStateMetadata{
			OwnerRaw:		rawOwner,
			PoolContractAddressRaw:	pool.ContractAddressRaw,
			TouchedKeys: []string{
				string(types.PoolKey(msg.PoolID)),
				string(types.PoolShareKey(msg.PoolID, msg.OwnerAddress)),
				string(types.PoolUnbondingKey(msg.PoolID, msg.OwnerAddress, msg.RequestID)),
			},
		},
	}, nil
}

func (k *Keeper) WithdrawPoolStake(msg types.MsgWithdrawPoolStake) (types.PoolWithdrawalReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("pool withdrawal caller contract", msg.CallerContractUser); err != nil {
		return types.PoolWithdrawalReceipt{}, err
	}
	if err := types.ValidateUserFacingAEAddress("pool withdrawal owner", msg.OwnerAddress); err != nil {
		return types.PoolWithdrawalReceipt{}, err
	}
	rawOwner, err := types.RawAddressForUserAddress(msg.OwnerAddress)
	if err != nil {
		return types.PoolWithdrawalReceipt{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PoolWithdrawalReceipt{}, errors.New("official liquid staking pool not found")
	}
	if !pool.OfficialLiquidStaking || pool.ContractAddressUser != msg.CallerContractUser {
		return types.PoolWithdrawalReceipt{}, errors.New("pool withdrawal requires official liquid staking contract")
	}
	withdrawalIdx, withdrawal, found := findWithdrawal(pool.PendingWithdrawals, msg.RequestID)
	if !found {
		return types.PoolWithdrawalReceipt{}, errors.New("pool withdrawal request not found")
	}
	if withdrawal.Delegator != rawOwner {
		return types.PoolWithdrawalReceipt{}, errors.New("pool withdrawal owner mismatch")
	}
	if withdrawal.Status != types.WithdrawalStatusPending {
		return types.PoolWithdrawalReceipt{}, errors.New("pool withdrawal is not pending")
	}
	if msg.Height < withdrawal.CompleteHeight {
		return types.PoolWithdrawalReceipt{}, errors.New("pool withdrawal cannot release before unbonding period")
	}
	withdrawal.Status = types.WithdrawalStatusCompleted
	pool.PendingWithdrawals[withdrawalIdx] = withdrawal
	for entryIdx, entry := range pool.UnbondingQueue {
		if entry.WithdrawalID == msg.RequestID {
			entry.Status = types.WithdrawalStatusCompleted
			pool.UnbondingQueue[entryIdx] = entry
		}
	}
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PoolWithdrawalReceipt{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.PoolWithdrawalReceipt{}, err
	}
	if err := k.upsertLiquidPoolAfterPoolMutation(pool, msg.Height); err != nil {
		return types.PoolWithdrawalReceipt{}, err
	}
	if err := k.upsertPoolUnbonding(msg.PoolID, msg.OwnerAddress, withdrawal); err != nil {
		return types.PoolWithdrawalReceipt{}, err
	}
	return types.PoolWithdrawalReceipt{
		PoolID:		msg.PoolID,
		OwnerAddress:	msg.OwnerAddress,
		RequestID:	msg.RequestID,
		Amount:		withdrawal.Amount,
		Height:		msg.Height,
		InternalMetadata: types.PoolStateMetadata{
			OwnerRaw:		rawOwner,
			PoolContractAddressRaw:	pool.ContractAddressRaw,
			TouchedKeys: []string{
				string(types.PoolKey(msg.PoolID)),
				string(types.PoolUnbondingKey(msg.PoolID, msg.OwnerAddress, msg.RequestID)),
			},
		},
	}, nil
}

func (k *Keeper) TopUpPoolReserve(msg types.MsgTopUpPoolReserve) (types.PoolTopUpReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("pool top-up payer", msg.PayerAddress); err != nil {
		return types.PoolTopUpReceipt{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.PoolTopUpReceipt{}, errors.New("pool top-up amount and height must be positive")
	}
	if err := k.ensureActiveWallet(msg.PayerAddress, "pool top-up"); err != nil {
		return types.PoolTopUpReceipt{}, err
	}
	rawPayer, err := types.RawAddressForUserAddress(msg.PayerAddress)
	if err != nil {
		return types.PoolTopUpReceipt{}, err
	}
	next := cloneGenesis(k.genesis)
	_, pool, found := findPool(next.State.Pools, msg.PoolID)
	if !found {
		return types.PoolTopUpReceipt{}, errors.New("official liquid staking pool not found")
	}
	if !pool.OfficialLiquidStaking {
		return types.PoolTopUpReceipt{}, errors.New("pool top-up requires official liquid staking pool")
	}
	if pool.Status == types.PoolStatusClosed {
		return types.PoolTopUpReceipt{}, errors.New("closed pool reserve cannot be topped up")
	}
	liquidIdx, liquid, found := findLiquidPool(next.State.LiquidStakingPools, msg.PoolID)
	if !found {
		return types.PoolTopUpReceipt{}, errors.New("liquid staking pool state not found")
	}
	if err := k.accrueOfficialPoolRent(&liquid, pool, msg.Height); err != nil {
		return types.PoolTopUpReceipt{}, err
	}
	debtPaid := msg.Amount
	if debtPaid > liquid.StorageRentDebt {
		debtPaid = liquid.StorageRentDebt
	}
	liquid.StorageRentDebt -= debtPaid
	liquid.ContractAddressUser = pool.ContractAddressUser
	liquid.ContractAddressRaw = pool.ContractAddressRaw
	liquid.Status = pool.Status
	if liquid.StorageRentDebt > 0 && pool.Status == types.PoolStatusActive {
		liquid.Status = types.PoolStatusFrozenLimited
	}
	next.State.LiquidStakingPools[liquidIdx] = liquid
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PoolTopUpReceipt{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.PoolTopUpReceipt{}, err
	}
	return types.PoolTopUpReceipt{
		PoolID:			msg.PoolID,
		PayerAddress:		msg.PayerAddress,
		Amount:			msg.Amount,
		StorageDebtPaid:	debtPaid,
		Height:			msg.Height,
		InternalMetadata: types.PoolStateMetadata{
			OwnerRaw:		rawPayer,
			PoolContractAddressRaw:	pool.ContractAddressRaw,
			TouchedKeys: []string{
				string(types.PoolKey(msg.PoolID)),
				string(types.PoolStorageDebtKey(msg.PoolID)),
			},
		},
	}, nil
}

func (k *Keeper) CancelPoolWithdrawal(msg types.MsgCancelPoolWithdrawal) (types.PendingWithdrawal, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingWithdrawal{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool not found")
	}
	withdrawalIdx, withdrawal, found := findWithdrawal(pool.PendingWithdrawals, msg.WithdrawalID)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal not found")
	}
	if withdrawal.Delegator != msg.Delegator {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal delegator mismatch")
	}
	if withdrawal.Status != types.WithdrawalStatusPending {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal is not pending")
	}
	withdrawal.Status = types.WithdrawalStatusCancelled
	pool.PendingWithdrawals[withdrawalIdx] = withdrawal
	for entryIdx, entry := range pool.UnbondingQueue {
		if entry.WithdrawalID == msg.WithdrawalID {
			entry.Status = types.WithdrawalStatusCancelled
			pool.UnbondingQueue[entryIdx] = entry
		}
	}
	shares, err := types.SharesForDepositChecked(pool, withdrawal.Amount)
	if err != nil {
		return types.PendingWithdrawal{}, err
	}
	if shares < withdrawal.Shares {
		shares = withdrawal.Shares
	}
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, msg.Delegator)
	if found {
		delegator.PendingRewards = types.AccruedReward(delegator, pool.RewardIndex)
		delegator.Shares += shares
		delegator.RewardIndexCheckpoint = pool.RewardIndex
		pool.DelegatorShares[delegatorIdx] = delegator
	} else {
		pool.DelegatorShares = append(pool.DelegatorShares, types.DelegatorShare{
			Delegator:		msg.Delegator,
			Shares:			shares,
			RewardIndexCheckpoint:	pool.RewardIndex,
			SlashIndexCheckpoint:	pool.SlashIndex,
		})
	}
	pool.TotalShares += shares
	pool.TotalBondedStake += withdrawal.Amount
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PendingWithdrawal{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.PendingWithdrawal{}, err
	}
	return withdrawal, nil
}

func (k *Keeper) ClaimPoolRewards(msg types.MsgClaimPoolRewards) (uint64, error) {
	ownerAddress := msg.OwnerAddress
	delegator := msg.Delegator
	if ownerAddress != "" {
		if msg.Height == 0 {
			return 0, errors.New("pool reward claim height must be positive")
		}
		if err := types.ValidateUserFacingAEAddress("pool reward claim owner", ownerAddress); err != nil {
			return 0, err
		}
		if err := k.ensureActiveWallet(ownerAddress, "pool reward claim"); err != nil {
			return 0, err
		}
		rawOwner, err := types.RawAddressForUserAddress(ownerAddress)
		if err != nil {
			return 0, err
		}
		delegator = rawOwner
	} else if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return 0, err
	}
	idx, pool, found := k.lookupPool(msg.PoolID)
	if !found {
		return 0, errors.New("nominator pool not found")
	}
	delegatorIdx, share, found := k.lookupDelegator(msg.PoolID, delegator)
	if !found {
		return 0, errors.New("nominator pool delegator not found")
	}
	reward := types.AccruedReward(share, pool.RewardIndex)
	share.PendingRewards = 0
	share.RewardIndexCheckpoint = pool.RewardIndex
	if err := share.Validate(); err != nil {
		return 0, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx].DelegatorShares[delegatorIdx] = share
	if err := k.saveGenesis(next); err != nil {
		return 0, err
	}
	if pool.OfficialLiquidStaking && msg.Height > 0 {
		if err := k.upsertLiquidPoolAfterPoolMutation(next.State.Pools[idx], msg.Height); err != nil {
			return 0, err
		}
	}
	k.counters.DelegatorRewardUpdates++
	if ownerAddress != "" {
		if err := k.upsertRewardClaim(msg.PoolID, ownerAddress, pool.RewardEpoch, reward); err != nil {
			return 0, err
		}
		if err := k.upsertPoolShare(msg.PoolID, ownerAddress, share, 0, msg.Height); err != nil {
			return 0, err
		}
	}
	return reward, nil
}

func (k *Keeper) ClaimPoolRewardsWithReceipt(msg types.MsgClaimPoolRewards) (types.PoolRewardClaimReceipt, error) {
	if msg.OwnerAddress == "" {
		return types.PoolRewardClaimReceipt{}, errors.New("pool reward claim requires AE owner address")
	}
	amount, err := k.ClaimPoolRewards(msg)
	if err != nil {
		return types.PoolRewardClaimReceipt{}, err
	}
	rawOwner, err := types.RawAddressForUserAddress(msg.OwnerAddress)
	if err != nil {
		return types.PoolRewardClaimReceipt{}, err
	}
	_, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PoolRewardClaimReceipt{}, errors.New("nominator pool not found")
	}
	return types.PoolRewardClaimReceipt{
		PoolID:		msg.PoolID,
		OwnerAddress:	msg.OwnerAddress,
		Amount:		amount,
		Epoch:		pool.RewardEpoch,
		Height:		msg.Height,
		InternalMetadata: types.PoolStateMetadata{
			OwnerRaw:		rawOwner,
			PoolContractAddressRaw:	pool.ContractAddressRaw,
			TouchedKeys: []string{
				string(types.PoolShareKey(msg.PoolID, msg.OwnerAddress)),
				string(types.RewardClaimKey(msg.PoolID, msg.OwnerAddress, pool.RewardEpoch)),
			},
		},
	}, nil
}

func (k *Keeper) ClaimStakeReputation(msg types.MsgClaimStakeReputation) (types.StakeReputationClaimReceipt, error) {
	if err := types.ValidateUserFacingAEAddress("stake reputation claim owner", msg.OwnerAddress); err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	if msg.Height == 0 {
		return types.StakeReputationClaimReceipt{}, errors.New("stake reputation claim height must be positive")
	}
	if err := k.ensureActiveWallet(msg.OwnerAddress, "stake reputation claim"); err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	next := cloneGenesis(k.genesis)
	_, pool, found := findPool(next.State.Pools, msg.PoolID)
	if !found {
		return types.StakeReputationClaimReceipt{}, errors.New("nominator pool not found")
	}
	shareIdx, share, found := findPoolShare(next.State.PoolShares, msg.PoolID, msg.OwnerAddress)
	if !found {
		return types.StakeReputationClaimReceipt{}, errors.New("pool share not found for stake reputation claim")
	}
	if msg.Height < share.LastReputationUpdate {
		return types.StakeReputationClaimReceipt{}, errors.New("stake reputation claim height precedes previous update")
	}
	elapsed := msg.Height - share.LastReputationUpdate
	if share.LastReputationUpdate == 0 {
		elapsed = msg.Height - share.CreatedHeight
	}
	effectiveStake, err := poolShareActiveStakeExposure(next.State, pool, share)
	if err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	delta, err := types.MulDivUint64(effectiveStake, elapsed, 1)
	if err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	share.StakeWeightedSeconds, err = types.CheckedAddUint64(share.StakeWeightedSeconds, delta)
	if err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	share.LastReputationUpdate = msg.Height
	share.UpdatedHeight = msg.Height
	next.State.PoolShares[shareIdx] = share

	scoreDelta, err := types.MulDivUint64(delta, uint64(k.genesis.Params.ReputationStakeWeightBps), uint64(types.MaxBasisPoints))
	if err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	rawOwner, err := types.RawAddressForUserAddress(msg.OwnerAddress)
	if err != nil {
		return types.StakeReputationClaimReceipt{}, err
	}
	return types.StakeReputationClaimReceipt{
		Account:		msg.OwnerAddress,
		PoolID:			msg.PoolID,
		ReputationDelta:	scoreDelta,
		Height:			msg.Height,
		InternalMetadata: types.PoolStateMetadata{
			OwnerRaw:	rawOwner,
			TouchedKeys:	[]string{string(types.PoolShareKey(msg.PoolID, msg.OwnerAddress))},
		},
	}, nil
}

func (k *Keeper) SyncPoolRewards(msg types.MsgSyncPoolRewards) (types.PoolRewardSummary, error) {
	idx, pool, found := k.lookupPool(msg.PoolID)
	if !found {
		return types.PoolRewardSummary{}, errors.New("nominator pool not found")
	}
	nextPool, summary, err := types.SyncPoolRewards(k.genesis.Params, pool, msg)
	if err != nil {
		return types.PoolRewardSummary{}, err
	}
	k.counters.ValidatorAllocationReads += summary.AllocationsTouched
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = nextPool
	if err := k.saveGenesis(next); err != nil {
		return types.PoolRewardSummary{}, err
	}
	if nextPool.OfficialLiquidStaking {
		if err := k.upsertLiquidPoolAfterPoolMutation(nextPool, msg.Height); err != nil {
			return types.PoolRewardSummary{}, err
		}
	}
	return summary, nil
}

func (k *Keeper) ClaimStakingRewards(msg types.MsgClaimStakingRewards) (uint64, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return 0, err
	}
	if !msg.InternalMigration {
		return 0, errors.New("direct staking reward claims are internal migration only; use pool reward claims")
	}
	if msg.Height == 0 {
		return 0, errors.New("staking reward claim height must be positive")
	}
	if err := types.ValidateRawAddress("staking reward delegator", msg.Delegator); err != nil {
		return 0, err
	}
	if err := types.ValidateRawAddress("staking reward validator", msg.Validator); err != nil {
		return 0, err
	}
	return 0, nil
}

func (k *Keeper) StakingProof(req types.StakingProofRequest) (types.StakingProofMetadata, error) {
	k.counters.ProofQueries++
	return types.BuildStakingProofMetadata(req)
}

func (k *Keeper) UpdatePoolCommission(msg types.MsgUpdatePoolCommission) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if pool.PoolOperator != msg.PoolOperator {
		return types.NominatorPool{}, errors.New("nominator pool operator mismatch")
	}
	pool.PoolCommissionBps = msg.PoolCommissionBps
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) ChangePoolValidator(msg types.MsgChangePoolValidator) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	if types.IsJailedValidatorStatus(msg.ValidatorStatus) {
		return types.NominatorPool{}, errors.New("nominator pool cannot delegate to jailed validator")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if pool.PoolOperator != msg.PoolOperator {
		return types.NominatorPool{}, errors.New("nominator pool operator mismatch")
	}
	if msg.Height == 0 {
		return types.NominatorPool{}, errors.New("nominator pool validator change height must be positive")
	}
	if pool.PendingValidatorTarget == msg.ValidatorTarget && msg.Height >= pool.ValidatorChangeHeight {
		pool.ValidatorTarget = msg.ValidatorTarget
		pool.PendingValidatorTarget = ""
		pool.ValidatorChangeHeight = 0
	} else if pool.PendingValidatorTarget == msg.ValidatorTarget {

	} else {
		pool.PendingValidatorTarget = msg.ValidatorTarget
		pool.ValidatorChangeHeight = msg.Height + k.genesis.Params.ValidatorChangeDelay
	}
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) ApplyPoolReward(poolID string, rewardAmount uint64) (types.NominatorPool, error) {
	idx, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if rewardAmount == 0 {
		return pool, nil
	}
	commission := rewardAmount * uint64(pool.PoolCommissionBps) / uint64(types.MaxBasisPoints)
	netReward := rewardAmount - commission
	pool.TotalBondedStake += netReward
	pool.RewardIndex += types.RewardDelta(netReward, pool.TotalShares)
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) ApplyPoolSlash(poolID string, slashAmount uint64) (types.NominatorPool, error) {
	idx, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if slashAmount == 0 {
		return pool, nil
	}
	if slashAmount > pool.TotalBondedStake {
		slashAmount = pool.TotalBondedStake
	}
	pool.TotalBondedStake -= slashAmount
	pool.SlashIndex += types.RewardDelta(slashAmount, pool.TotalShares)
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) ApplyValidatorSlash(msg types.MsgApplyValidatorSlash) ([]types.ValidatorSlashEvent, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return nil, err
	}

	var slashBps uint32
	switch msg.Fault {
	case types.SlashingFaultDowntime:
		slashBps = k.genesis.Params.DowntimeSlashBps
	case types.SlashingFaultDoubleSign:
		slashBps = k.genesis.Params.DoubleSignSlashBps
	default:
		return nil, errors.New("unknown slashing fault")
	}

	// Find and update the validator
	var validatorIdx int
	var validator types.Validator
	var found bool
	for i, v := range k.genesis.State.Validators {
		if v.Address == msg.ValidatorAddress {
			validatorIdx = i
			validator = v
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("validator not found")
	}

	// Determine new status
	var newStatus string
	var tombstoned bool
	switch msg.Fault {
	case types.SlashingFaultDowntime:
		newStatus = types.StateValidatorStatusJailed
	case types.SlashingFaultDoubleSign:
		newStatus = types.StateValidatorStatusSlashed
		if k.genesis.Params.DoubleSignTombstone {
			tombstoned = true
		}
	default:
		newStatus = validator.Status
	}

	validator.Status = newStatus
	validator.SlashingRiskBps = slashBps
	if msg.Fault == types.SlashingFaultDowntime {
		validator.Jailed = true
	}
	if tombstoned {
		validator.Tombstoned = true
	}
	k.genesis.State.Validators[validatorIdx] = validator

	// Create slash events for affected pools
	var events []types.ValidatorSlashEvent

	for poolIdx, pool := range k.genesis.State.Pools {
		slashAmount := uint64(0)

		for allocIdx, alloc := range pool.Allocations {
			if alloc.ValidatorAddress == msg.ValidatorAddress {
				loss := alloc.Amount * uint64(slashBps) / uint64(types.MaxBasisPoints)
				slashAmount += loss
				pool.Allocations[allocIdx].Amount -= loss
			}
		}

		if slashAmount > 0 {

			if pool.TotalBondedStake >= slashAmount {
				pool.TotalBondedStake -= slashAmount
			} else {
				pool.TotalBondedStake = 0
			}

			pool.SlashIndex += types.RewardDelta(slashAmount, pool.TotalShares)

			k.genesis.State.Pools[poolIdx] = pool

			event := types.ValidatorSlashEvent{
				Height:			msg.Height,
				Validator:		msg.ValidatorAddress,
				PoolID:			pool.PoolID,
				Fault:			msg.Fault,
				Epoch:			msg.Epoch,
				SlashingLoss:		slashAmount,
				ValidatorStatus:	newStatus,
				Tombstoned:		tombstoned,
				PoolSlashIndexAfter:	pool.SlashIndex,
			}
			events = append(events, event)

			for pvIdx, pv := range k.genesis.State.PoolValidatorAllocations {
				if pv.PoolID == pool.PoolID && pv.Validator == msg.ValidatorAddress {

					loss := pv.ActiveStake * uint64(slashBps) / uint64(types.MaxBasisPoints)
					if pv.ActiveStake >= loss {
						k.genesis.State.PoolValidatorAllocations[pvIdx].ActiveStake -= loss
					} else {
						k.genesis.State.PoolValidatorAllocations[pvIdx].ActiveStake = 0
					}

					if newStatus != types.StateValidatorStatusActive {
						k.genesis.State.PoolValidatorAllocations[pvIdx].TargetWeightBps = 0
					}
				}
			}
		}
	}

	for _, event := range events {
		k.genesis.State.ValidatorSlashEvents = append(k.genesis.State.ValidatorSlashEvents, event)
	}

	if k.storeService != nil {
		if k.runtimeCtx == nil {
			k.runtimeCtx = context.Background()
		}

		if err := prefixgenesis.Save(k.runtimeCtx, k.storeService, genesisKey, cloneGenesis(k.genesis)); err != nil {
			return nil, err
		}

		store := k.storeService.OpenKVStore(k.runtimeCtx)
		for _, pool := range k.genesis.State.Pools {
			bz, err := json.Marshal(pool)
			if err != nil {
				return nil, err
			}
			if err := store.Set(types.PoolKey(pool.PoolID), bz); err != nil {
				return nil, err
			}
		}
		for _, alloc := range k.genesis.State.PoolValidatorAllocations {
			bz, err := json.Marshal(alloc)
			if err != nil {
				return nil, err
			}
			if err := store.Set(types.PoolAllocationKey(alloc.PoolID, alloc.Validator), bz); err != nil {
				return nil, err
			}
		}
	}

	return events, nil
}

func (k *Keeper) NominatorPool(poolID string) (types.NominatorPool, bool) {
	_, pool, found := k.lookupPool(poolID)
	return pool, found
}

func (k Keeper) NominatorPools() []types.NominatorPool {
	return types.SortPools(k.genesis.State.Pools)
}

func (k *Keeper) PoolDelegator(poolID string, delegator string) (types.DelegatorShare, bool) {
	_, _, found := k.lookupPool(poolID)
	if !found {
		return types.DelegatorShare{}, false
	}
	_, share, found := k.lookupDelegator(poolID, delegator)
	return share, found
}

func (k *Keeper) PoolRewards(poolID string, delegator string) (uint64, bool) {
	_, pool, found := k.lookupPool(poolID)
	if !found {
		return 0, false
	}
	_, share, found := k.lookupDelegator(poolID, delegator)
	if !found {
		return 0, false
	}
	return types.AccruedReward(share, pool.RewardIndex), true
}

func (k *Keeper) PoolShare(req types.QueryPoolShareRequest) (types.QueryPoolShareResponse, bool) {
	_, pool, found := k.lookupPool(req.PoolID)
	if !found {
		return types.QueryPoolShareResponse{}, false
	}
	_, share, found := k.lookupDelegator(req.PoolID, req.Delegator)
	if !found {
		return types.QueryPoolShareResponse{}, false
	}
	return types.QueryPoolShareResponse{
		Share:		share,
		PendingRewards:	types.AccruedReward(share, pool.RewardIndex),
	}, true
}

func (k *Keeper) PoolAllocations(req types.QueryPoolAllocationsRequest) (types.QueryPoolAllocationsResponse, bool) {
	_, pool, found := k.lookupPool(req.PoolID)
	if !found {
		return types.QueryPoolAllocationsResponse{}, false
	}
	return types.QueryPoolAllocationsResponse{Allocations: types.SortValidatorRewardAllocations(pool.ValidatorAllocations)}, true
}

func (k Keeper) StakingRewards(req types.QueryStakingRewardsRequest) (types.QueryStakingRewardsResponse, error) {
	if !req.InternalMigration {
		return types.QueryStakingRewardsResponse{}, errors.New("staking rewards query is internal migration only; use pool rewards")
	}
	return types.QueryStakingRewardsResponse{RewardAmount: 0}, nil
}

func (k Keeper) PoolUnbondingQueue(poolID string) []types.UnbondingEntry {
	_, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return []types.UnbondingEntry{}
	}
	return types.SortUnbonding(pool.UnbondingQueue)
}

type Migrator struct{ keeper *Keeper }

func NewMigrator(k *Keeper) Migrator	{ return Migrator{keeper: k} }
func (m Migrator) Migrate1to2() error	{ return m.keeper.ExportGenesis().Validate() }
func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k *Keeper) savePool(idx int, pool types.NominatorPool, delegator types.DelegatorShare) (types.DelegatorShare, error) {
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.DelegatorShare{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.DelegatorShare{}, err
	}
	return delegator, nil
}

func (k *Keeper) savePoolOnly(idx int, pool types.NominatorPool) (types.NominatorPool, error) {
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.NominatorPool{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.NominatorPool{}, err
	}
	return pool, nil
}

func (k *Keeper) ensureActiveWallet(address string, action string) error {
	if k.accountStatusReader == nil {
		return nil
	}
	status, found := k.accountStatusReader.AccountStatus(address)
	if !found || status == accountStatusInactive {
		return errors.New(action + " requires active wallet")
	}
	if status == accountStatusFrozen {
		return errors.New(action + " rejected for frozen wallet; pay storage debt and unfreeze first")
	}
	if status != accountStatusActive {
		return errors.New(action + " requires active wallet")
	}
	return nil
}

func (k *Keeper) accrueOfficialPoolRent(liquid *types.LiquidStakingPool, pool types.NominatorPool, height uint64) error {
	if liquid == nil {
		return errors.New("official pool storage rent state is required")
	}
	if height == 0 {
		return errors.New("official pool storage rent height must be positive")
	}
	if liquid.LastStorageChargeHeight == 0 {
		liquid.LastStorageChargeHeight = height
		return nil
	}
	if height < liquid.LastStorageChargeHeight {
		return errors.New("official pool storage rent height must be monotonic")
	}
	if height == liquid.LastStorageChargeHeight {
		return nil
	}
	rate := k.genesis.Params.StorageRentRatePerByteSecond
	if rate == 0 {
		liquid.LastStorageChargeHeight = height
		return nil
	}
	elapsed := height - liquid.LastStorageChargeHeight
	footprint := officialPoolStorageFootprintBytes(pool, *liquid)
	charge, err := multiplyUint64Checked(elapsed, footprint, rate)
	if err != nil {
		return err
	}
	if liquid.StorageRentReserve >= charge {
		liquid.StorageRentReserve -= charge
	} else {
		unpaid := charge - liquid.StorageRentReserve
		liquid.StorageRentReserve = 0
		liquid.StorageRentDebt, err = types.CheckedAddUint64(liquid.StorageRentDebt, unpaid)
		if err != nil {
			return err
		}
	}
	liquid.LastStorageChargeHeight = height
	liquid.Status = pool.Status
	if liquid.StorageRentDebt > 0 && pool.Status == types.PoolStatusActive {
		liquid.Status = types.PoolStatusFrozenLimited
	}
	return nil
}

func officialPoolStorageFootprintBytes(pool types.NominatorPool, liquid types.LiquidStakingPool) uint64 {
	base := uint64(160)
	base += uint64(len(pool.PoolID) + len(pool.ContractAddressUser) + len(pool.ContractAddressRaw))
	base += uint64(len(liquid.ReceiptToken) + len(liquid.RentPayerPolicy) + len(liquid.Status))
	base += uint64(len(pool.DelegatorShares)) * 48
	base += uint64(len(pool.PendingWithdrawals)) * 56
	base += uint64(len(pool.Allocations)) * 40
	if base == 0 {
		return 1
	}
	return base
}

func multiplyUint64Checked(factors ...uint64) (uint64, error) {
	acc := big.NewInt(1)
	limit := new(big.Int).SetUint64(math.MaxUint64)
	for _, factor := range factors {
		acc.Mul(acc, new(big.Int).SetUint64(factor))
		if acc.Cmp(limit) > 0 {
			return 0, errors.New("nominator pool uint64 accounting overflow")
		}
	}
	return acc.Uint64(), nil
}

func (k *Keeper) upsertLiquidPoolAfterPoolMutation(pool types.NominatorPool, height uint64) error {
	idx, liquid, found := findLiquidPool(k.genesis.State.LiquidStakingPools, pool.PoolID)
	if !found {
		liquid = types.LiquidStakingPool{
			PoolID:			pool.PoolID,
			ContractAddressUser:	pool.ContractAddressUser,
			ContractAddressRaw:	pool.ContractAddressRaw,
			ReceiptToken:		k.genesis.Params.PoolReceiptDenomOrCodeID,
			RentPayerPolicy:	types.RentPayerPolicyPoolReserve,
			Status:			pool.Status,
		}
	}
	liquid.ContractAddressUser = pool.ContractAddressUser
	liquid.ContractAddressRaw = pool.ContractAddressRaw
	liquid.TotalDeposited = pool.TotalBondedStake + totalPendingWithdrawalAmount(pool.PendingWithdrawals)
	liquid.TotalActiveStake = totalAllocated(pool.Allocations)
	liquid.TotalUnbonding = totalPendingWithdrawalAmount(pool.PendingWithdrawals)
	liquid.TotalShares = pool.TotalShares
	liquid.RewardIndex = pool.RewardIndex
	if err := k.accrueOfficialPoolRent(&liquid, pool, height); err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	if found {
		next.State.LiquidStakingPools[idx] = liquid
	} else {
		next.State.LiquidStakingPools = append(next.State.LiquidStakingPools, liquid)
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k *Keeper) upsertPoolShare(poolID string, owner string, delegator types.DelegatorShare, principalDelta uint64, height uint64) error {
	idx, share, found := findPoolShare(k.genesis.State.PoolShares, poolID, owner)
	if !found {
		share = types.PoolShare{
			Owner:			owner,
			PoolID:			poolID,
			CreatedHeight:		height,
			LastReputationUpdate:	height,
		}
	}
	share.Shares = delegator.Shares
	share.PrincipalAmount += principalDelta
	share.UpdatedHeight = height
	share.LastRewardIndex = delegator.RewardIndexCheckpoint
	share.PendingRewards = delegator.PendingRewards
	next := cloneGenesis(k.genesis)
	if found {
		next.State.PoolShares[idx] = share
	} else {
		next.State.PoolShares = append(next.State.PoolShares, share)
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k *Keeper) upsertPoolUnbonding(poolID string, owner string, withdrawal types.PendingWithdrawal) error {
	idx, request, found := findPoolUnbonding(k.genesis.State.PoolUnbondingRequests, poolID, owner, withdrawal.WithdrawalID)
	if !found {
		request = types.PoolUnbondingRequest{PoolID: poolID, Owner: owner, RequestID: withdrawal.WithdrawalID}
	}
	request.Shares = withdrawal.Shares
	request.Amount = withdrawal.Amount
	request.RequestHeight = withdrawal.RequestHeight
	request.CompleteHeight = withdrawal.CompleteHeight
	request.Status = withdrawal.Status
	next := cloneGenesis(k.genesis)
	if found {
		next.State.PoolUnbondingRequests[idx] = request
	} else {
		next.State.PoolUnbondingRequests = append(next.State.PoolUnbondingRequests, request)
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k *Keeper) updatePoolShareAfterUnbond(poolID string, owner string, withdrawal types.PendingWithdrawal, height uint64) error {
	idx, share, found := findPoolShare(k.genesis.State.PoolShares, poolID, owner)
	if !found {
		return nil
	}
	next := cloneGenesis(k.genesis)
	if withdrawal.Shares >= share.Shares {
		next.State.PoolShares = append(next.State.PoolShares[:idx], next.State.PoolShares[idx+1:]...)
	} else {
		share.Shares -= withdrawal.Shares
		if withdrawal.Amount >= share.PrincipalAmount {
			share.PrincipalAmount = 1
		} else {
			share.PrincipalAmount -= withdrawal.Amount
		}
		share.UpdatedHeight = height
		next.State.PoolShares[idx] = share
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k *Keeper) upsertRewardClaim(poolID string, owner string, epoch uint64, amount uint64) error {
	if epoch == 0 {
		epoch = 1
	}
	idx, claim, found := findRewardClaim(k.genesis.State.RewardClaims, poolID, owner, epoch)
	if !found {
		claim = types.RewardClaim{PoolID: poolID, Owner: owner, Epoch: epoch}
	}
	var err error
	claim.Amount, err = types.CheckedAddUint64(claim.Amount, amount)
	if err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	if found {
		next.State.RewardClaims[idx] = claim
	} else {
		next.State.RewardClaims = append(next.State.RewardClaims, claim)
	}
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k *Keeper) upsertPoolValidatorAllocation(poolID string, validatorAddress string, amount uint64, height uint64) error {
	// perform manual lookup with verbose debugging
	var validator types.Validator
	foundValidator := false
	for _, v := range k.genesis.State.Validators {
		if v.Address == validatorAddress {
			validator = v
			foundValidator = true
			break
		}
	}
	if !foundValidator || validator.Status != types.StateValidatorStatusActive {
		return errors.New("pool allocation requires registered active validator")
	}
	idx, allocation, found := findPoolValidatorAllocation(k.genesis.State.PoolValidatorAllocations, poolID, validatorAddress)
	if !found {
		allocation = types.PoolValidatorAllocation{PoolID: poolID, Validator: validatorAddress}
	}
	_, pool, poolFound := findPool(k.genesis.State.Pools, poolID)
	if !poolFound {
		return errors.New("nominator pool not found")
	}
	targetWeight := uint32(0)
	if pool.TotalBondedStake > 0 {
		weight, err := types.MulDivUint64(amount, uint64(types.MaxBasisPoints), pool.TotalBondedStake)
		if err != nil {
			return err
		}
		targetWeight = uint32(weight)
	}
	allocation.TargetWeightBps = targetWeight
	allocation.ActiveStake = amount
	allocation.PerformanceScore = validator.PerformanceScore
	allocation.CommissionBps = validator.CommissionBps
	allocation.SlashingRiskBps = validator.SlashingRiskBps
	allocation.UpdatedHeight = height

	if found {
		k.genesis.State.PoolValidatorAllocations[idx] = allocation
	} else {
		k.genesis.State.PoolValidatorAllocations = append(k.genesis.State.PoolValidatorAllocations, allocation)
	}
	k.genesis.State = k.genesis.State.Normalize(k.genesis.Params)
	k.rebuildIndexes()

	if k.storeService != nil {
		if k.runtimeCtx == nil {
			k.runtimeCtx = context.Background()
		}
		store := k.storeService.OpenKVStore(k.runtimeCtx)
		bz, err := json.Marshal(allocation)
		if err != nil {
			return err
		}
		if err := store.Set(types.PoolAllocationKey(poolID, validatorAddress), bz); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) poolAllocationReceipt(pool types.NominatorPool, epoch uint64, height uint64) (types.PoolRebalanceReceipt, error) {
	allocations := []types.PoolValidatorAllocation{}
	touched := []string{string(types.PoolKey(pool.PoolID))}
	for _, allocation := range types.SortPoolValidatorAllocations(k.genesis.State.PoolValidatorAllocations) {
		if allocation.PoolID != pool.PoolID {
			continue
		}
		allocations = append(allocations, allocation)
		touched = append(touched, string(types.PoolAllocationKey(pool.PoolID, allocation.Validator)))
	}
	return types.PoolRebalanceReceipt{
		PoolID:		pool.PoolID,
		Epoch:		epoch,
		Height:		height,
		Allocations:	allocations,
		InternalMetadata: types.PoolStateMetadata{
			PoolContractAddressRaw:	pool.ContractAddressRaw,
			TouchedKeys:		touched,
		},
	}, nil
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}

func findPool(pools []types.NominatorPool, poolID string) (int, types.NominatorPool, bool) {
	for idx, pool := range pools {
		if pool.PoolID == poolID {
			return idx, pool, true
		}
	}
	return -1, types.NominatorPool{}, false
}

func findDelegator(delegators []types.DelegatorShare, delegator string) (int, types.DelegatorShare, bool) {
	for idx, share := range delegators {
		if share.Delegator == delegator {
			return idx, share, true
		}
	}
	return -1, types.DelegatorShare{}, false
}

func findWithdrawal(withdrawals []types.PendingWithdrawal, withdrawalID string) (int, types.PendingWithdrawal, bool) {
	for idx, withdrawal := range withdrawals {
		if withdrawal.WithdrawalID == withdrawalID {
			return idx, withdrawal, true
		}
	}
	return -1, types.PendingWithdrawal{}, false
}

func findAllocation(allocations []types.PoolAllocation, validatorAddress string) (int, types.PoolAllocation, bool) {
	for idx, allocation := range allocations {
		if allocation.ValidatorAddress == validatorAddress {
			return idx, allocation, true
		}
	}
	return -1, types.PoolAllocation{}, false
}

func findValidator(validators []types.Validator, validatorAddress string) (int, types.Validator, bool) {
	for idx, validator := range validators {
		if validator.Address == validatorAddress {
			return idx, validator, true
		}
	}
	return -1, types.Validator{}, false
}

func findLiquidPool(pools []types.LiquidStakingPool, poolID string) (int, types.LiquidStakingPool, bool) {
	for idx, pool := range pools {
		if pool.PoolID == poolID {
			return idx, pool, true
		}
	}
	return -1, types.LiquidStakingPool{}, false
}

func findPoolShare(shares []types.PoolShare, poolID string, owner string) (int, types.PoolShare, bool) {
	for idx, share := range shares {
		if share.PoolID == poolID && share.Owner == owner {
			return idx, share, true
		}
	}
	return -1, types.PoolShare{}, false
}

func findPoolUnbonding(requests []types.PoolUnbondingRequest, poolID string, owner string, requestID string) (int, types.PoolUnbondingRequest, bool) {
	for idx, request := range requests {
		if request.PoolID == poolID && request.Owner == owner && request.RequestID == requestID {
			return idx, request, true
		}
	}
	return -1, types.PoolUnbondingRequest{}, false
}

func findPoolValidatorAllocation(allocations []types.PoolValidatorAllocation, poolID string, validator string) (int, types.PoolValidatorAllocation, bool) {
	for idx, allocation := range allocations {
		if allocation.PoolID == poolID && allocation.Validator == validator {
			return idx, allocation, true
		}
	}
	return -1, types.PoolValidatorAllocation{}, false
}

func findRewardClaim(claims []types.RewardClaim, poolID string, owner string, epoch uint64) (int, types.RewardClaim, bool) {
	for idx, claim := range claims {
		if claim.PoolID == poolID && claim.Owner == owner && claim.Epoch == epoch {
			return idx, claim, true
		}
	}
	return -1, types.RewardClaim{}, false
}

func poolShareActiveStakeExposure(state types.State, pool types.NominatorPool, share types.PoolShare) (uint64, error) {
	if share.Shares == 0 || pool.TotalShares == 0 {
		return 0, nil
	}
	activeStake := totalAllocated(pool.Allocations)
	if _, liquid, found := findLiquidPool(state.LiquidStakingPools, pool.PoolID); found {
		activeStake = liquid.TotalActiveStake
	}
	if activeStake == 0 {
		return 0, nil
	}
	return types.MulDivUint64(activeStake, share.Shares, pool.TotalShares)
}

func totalAllocated(allocations []types.PoolAllocation) uint64 {
	total := uint64(0)
	for _, allocation := range allocations {
		total += allocation.Amount
	}
	return total
}

func totalPendingWithdrawalAmount(withdrawals []types.PendingWithdrawal) uint64 {
	total := uint64(0)
	for _, withdrawal := range withdrawals {
		if withdrawal.Status == types.WithdrawalStatusPending {
			total += withdrawal.Amount
		}
	}
	return total
}
