package keeper

import (
	"context"
	"encoding/binary"

	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/treasury/types"
)

type Keeper struct {
	cdc		codec.BinaryCodec
	storeService	corestore.KVStoreService
	accountKeeper	types.AccountKeeper
	bankKeeper	types.BankKeeper
	authority	string
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, accountKeeper: accountKeeper, bankKeeper: bankKeeper, authority: authority}
}

func (k Keeper) Authority() string	{ return k.authority }

func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	params = types.NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.ParamsKey, bz)
}

func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.ParamsKey)
	if err != nil || bz == nil {
		return types.DefaultParams(), err
	}
	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return types.Params{}, err
	}
	params = types.NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return types.Params{}, types.ErrInvalidParams.Wrap(err.Error())
	}
	return params, nil
}

func (k Keeper) SetAllocations(ctx context.Context, allocations types.TreasuryAllocations) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := allocations.Validate(params); err != nil {
		return types.ErrAccounting.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&allocations)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.AllocationsKey, bz)
}

func (k Keeper) GetAllocations(ctx context.Context) (types.TreasuryAllocations, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.AllocationsKey)
	if err != nil || bz == nil {
		return types.DefaultAllocations(), err
	}
	var allocations types.TreasuryAllocations
	if err := k.cdc.Unmarshal(bz, &allocations); err != nil {
		return types.TreasuryAllocations{}, err
	}
	return allocations, nil
}

func (k Keeper) SyncIncomingFunds(ctx context.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	allocations, err := k.GetAllocations(ctx)
	if err != nil {
		return err
	}
	moduleAddr := k.accountKeeper.GetModuleAddress(params.TreasuryModule)
	if moduleAddr == nil {
		return types.ErrAccounting.Wrapf("module account %s is not configured", params.TreasuryModule)
	}
	bankBalance := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	accountingBalance := allocations.AccountingBalance()
	if bankBalance.Equal(accountingBalance) {
		return nil
	}
	if !bankBalance.IsAllGTE(accountingBalance) {
		return types.ErrAccounting.Wrapf("module bank balance %s is below accounting state %s", bankBalance, accountingBalance)
	}
	delta := bankBalance.Sub(accountingBalance...)
	allocations = addIncomingByPolicy(allocations, params, delta)
	return k.SetAllocations(ctx, allocations)
}

func (k Keeper) AssertTreasuryAccountingInvariant(ctx context.Context) error {
	if err := k.SyncIncomingFunds(ctx); err != nil {
		return err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	allocations, err := k.GetAllocations(ctx)
	if err != nil {
		return err
	}
	moduleAddr := k.accountKeeper.GetModuleAddress(params.TreasuryModule)
	if moduleAddr == nil {
		return types.ErrAccounting.Wrapf("module account %s is not configured", params.TreasuryModule)
	}
	bankBalance := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	accountingBalance := allocations.AccountingBalance()
	if !bankBalance.Equal(accountingBalance) {
		return types.ErrAccounting.Wrapf("module bank balance %s != accounting state %s", bankBalance, accountingBalance)
	}
	return nil
}

func (k Keeper) SubmitSpend(ctx context.Context, proposer, recipient string, amount sdk.Coins, bucket string, epoch, vestingStart, vestingEnd uint64, metadata string) (types.TreasurySpend, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	if err := aetraaddress.ValidateUserAddress("proposer", proposer); err != nil {
		return types.TreasurySpend{}, types.ErrUnauthorized.Wrap(err.Error())
	}
	if err := aetraaddress.ValidateUserAddress("recipient", recipient); err != nil {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap(err.Error())
	}
	if !types.IsRecipientAllowed(params, recipient) {
		return types.TreasurySpend{}, types.ErrUnauthorized.Wrap("recipient is not allowlisted")
	}
	if epoch == 0 {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap("epoch must be positive")
	}
	if len(metadata) > int(params.MaxMetadataBytes) {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap("metadata exceeds max_metadata_bytes")
	}
	if err := types.ValidateTreasuryCoins(params, amount, false); err != nil {
		return types.TreasurySpend{}, err
	}
	id, err := k.nextSpendID(ctx)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	spend := types.TreasurySpend{
		Id:			id,
		Proposer:		proposer,
		Recipient:		recipient,
		Amount:			amount,
		Bucket:			bucket,
		Status:			types.StatusPending,
		Epoch:			epoch,
		VestingStartEpoch:	vestingStart,
		VestingEndEpoch:	vestingEnd,
		Metadata:		metadata,
		CreatedHeight:		sdk.UnwrapSDKContext(ctx).BlockHeight(),
		UpdatedHeight:		sdk.UnwrapSDKContext(ctx).BlockHeight(),
	}
	if err := spend.Validate(params); err != nil {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap(err.Error())
	}
	if err := k.SetSpend(ctx, spend); err != nil {
		return types.TreasurySpend{}, err
	}
	if err := k.setNextSpendID(ctx, id+1); err != nil {
		return types.TreasurySpend{}, err
	}
	return spend, nil
}

func (k Keeper) ApproveSpend(ctx context.Context, id uint64, metadata string) (types.TreasurySpend, error) {
	return k.transitionSpend(ctx, id, types.StatusPending, types.StatusApproved, metadata)
}

func (k Keeper) RejectSpend(ctx context.Context, id uint64, metadata string) (types.TreasurySpend, error) {
	return k.transitionSpend(ctx, id, types.StatusPending, types.StatusRejected, metadata)
}

func (k Keeper) CancelSpend(ctx context.Context, id uint64, actor, metadata string) (types.TreasurySpend, error) {
	spend, found, err := k.GetSpend(ctx, id)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	if !found {
		return types.TreasurySpend{}, types.ErrNotFound.Wrapf("spend %d", id)
	}
	if actor != spend.Proposer && actor != k.Authority() {
		return types.TreasurySpend{}, types.ErrUnauthorized.Wrap("only proposer or authority can cancel spend")
	}
	if types.SpendIsTerminal(spend.Status) {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrapf("spend %d is already terminal", id)
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	spend.Status = types.StatusCanceled
	spend.Metadata = metadata
	spend.UpdatedHeight = sdk.UnwrapSDKContext(ctx).BlockHeight()
	if err := spend.Validate(params); err != nil {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap(err.Error())
	}
	return spend, k.SetSpend(ctx, spend)
}

func (k Keeper) ExecuteSpend(ctx context.Context, id, epoch uint64) (types.TreasurySpend, error) {
	if epoch == 0 {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap("epoch must be positive")
	}
	if err := k.SyncIncomingFunds(ctx); err != nil {
		return types.TreasurySpend{}, err
	}
	spend, found, err := k.GetSpend(ctx, id)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	if !found {
		return types.TreasurySpend{}, types.ErrNotFound.Wrapf("spend %d", id)
	}
	if spend.Status != types.StatusApproved {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrapf("spend %d is not approved", id)
	}
	if spend.VestingEndEpoch != 0 && epoch < spend.VestingEndEpoch {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap("vesting cannot release early")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	if !types.IsRecipientAllowed(params, spend.Recipient) {
		return types.TreasurySpend{}, types.ErrUnauthorized.Wrap("recipient is not allowlisted")
	}
	recipient, err := aetraaddress.ParseAccAddress(spend.Recipient)
	if err != nil {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap(err.Error())
	}
	allocations, err := k.GetAllocations(ctx)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	available := bucketBalance(allocations, spend.Bucket)
	if !available.IsAllGTE(spend.Amount) {
		return types.TreasurySpend{}, types.ErrInsufficientFunds.Wrapf("bucket %s has %s, needs %s", spend.Bucket, available, spend.Amount)
	}
	epochSpend, _, err := k.GetEpochSpend(ctx, epoch)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	nextEpochSpent := epochSpend.Spent.Add(spend.Amount...)
	if !capAllows(params, nextEpochSpent) {
		return types.TreasurySpend{}, types.ErrSpendCapExceeded.Wrapf("epoch %d spend %s exceeds cap %s", epoch, nextEpochSpent, params.PerEpochSpendCap)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, params.TreasuryModule, recipient, spend.Amount); err != nil {
		return types.TreasurySpend{}, err
	}
	allocations = subtractFromBucket(allocations, spend.Bucket, spend.Amount)
	allocations.TotalSpent = allocations.TotalSpent.Add(spend.Amount...)
	if err := k.SetAllocations(cacheCtx, allocations); err != nil {
		return types.TreasurySpend{}, err
	}
	epochSpend.Epoch = epoch
	epochSpend.Spent = nextEpochSpent
	if err := k.SetEpochSpend(cacheCtx, epochSpend); err != nil {
		return types.TreasurySpend{}, err
	}
	spend.Status = types.StatusExecuted
	spend.UpdatedHeight = sdkCtx.BlockHeight()
	spend.ExecutedHeight = sdkCtx.BlockHeight()
	if err := k.SetSpend(cacheCtx, spend); err != nil {
		return types.TreasurySpend{}, err
	}
	write()
	return spend, nil
}

func (k Keeper) transitionSpend(ctx context.Context, id uint64, from, to, metadata string) (types.TreasurySpend, error) {
	spend, found, err := k.GetSpend(ctx, id)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	if !found {
		return types.TreasurySpend{}, types.ErrNotFound.Wrapf("spend %d", id)
	}
	if spend.Status != from {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrapf("spend %d has status %s", id, spend.Status)
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.TreasurySpend{}, err
	}
	if len(metadata) > int(params.MaxMetadataBytes) {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap("metadata exceeds max_metadata_bytes")
	}
	spend.Status = to
	spend.Metadata = metadata
	spend.UpdatedHeight = sdk.UnwrapSDKContext(ctx).BlockHeight()
	if err := spend.Validate(params); err != nil {
		return types.TreasurySpend{}, types.ErrInvalidSpend.Wrap(err.Error())
	}
	return spend, k.SetSpend(ctx, spend)
}

func (k Keeper) SetSpend(ctx context.Context, spend types.TreasurySpend) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := spend.Validate(params); err != nil {
		return types.ErrInvalidSpend.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&spend)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(spendKey(spend.Id), bz)
}

func (k Keeper) GetSpend(ctx context.Context, id uint64) (types.TreasurySpend, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(spendKey(id))
	if err != nil || bz == nil {
		return types.TreasurySpend{}, false, err
	}
	var spend types.TreasurySpend
	if err := k.cdc.Unmarshal(bz, &spend); err != nil {
		return types.TreasurySpend{}, false, err
	}
	return spend, true, nil
}

func (k Keeper) GetAllSpends(ctx context.Context, status string) ([]types.TreasurySpend, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.SpendPrefix, storetypes.PrefixEndBytes(types.SpendPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.TreasurySpend{}
	for ; iter.Valid(); iter.Next() {
		var spend types.TreasurySpend
		if err := k.cdc.Unmarshal(iter.Value(), &spend); err != nil {
			return nil, err
		}
		if status == "" || spend.Status == status {
			out = append(out, spend)
		}
	}
	return types.SortSpends(out), nil
}

func (k Keeper) SetEpochSpend(ctx context.Context, epochSpend types.EpochSpend) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := epochSpend.Validate(params); err != nil {
		return types.ErrInvalidSpend.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&epochSpend)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(epochSpendKey(epochSpend.Epoch), bz)
}

func (k Keeper) GetEpochSpend(ctx context.Context, epoch uint64) (types.EpochSpend, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(epochSpendKey(epoch))
	if err != nil || bz == nil {
		return types.EpochSpend{Epoch: epoch, Spent: sdk.NewCoins()}, false, err
	}
	var epochSpend types.EpochSpend
	if err := k.cdc.Unmarshal(bz, &epochSpend); err != nil {
		return types.EpochSpend{}, false, err
	}
	return epochSpend, true, nil
}

func (k Keeper) GetAllEpochSpends(ctx context.Context) ([]types.EpochSpend, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.EpochSpendPrefix, storetypes.PrefixEndBytes(types.EpochSpendPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.EpochSpend{}
	for ; iter.Valid(); iter.Next() {
		var epochSpend types.EpochSpend
		if err := k.cdc.Unmarshal(iter.Value(), &epochSpend); err != nil {
			return nil, err
		}
		out = append(out, epochSpend)
	}
	return types.SortEpochSpends(out), nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	gs.Params = types.NormalizeParams(gs.Params)
	if gs.NextSpendId == 0 {
		gs.NextSpendId = 1
	}
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	if err := k.SetAllocations(ctx, gs.Allocations); err != nil {
		return err
	}
	for _, spend := range types.SortSpends(gs.Spends) {
		if err := k.SetSpend(ctx, spend); err != nil {
			return err
		}
	}
	for _, epochSpend := range types.SortEpochSpends(gs.EpochSpends) {
		if err := k.SetEpochSpend(ctx, epochSpend); err != nil {
			return err
		}
	}
	return k.setNextSpendID(ctx, gs.NextSpendId)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	if err := k.SyncIncomingFunds(ctx); err != nil {
		return nil, err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	allocations, err := k.GetAllocations(ctx)
	if err != nil {
		return nil, err
	}
	spends, err := k.GetAllSpends(ctx, "")
	if err != nil {
		return nil, err
	}
	epochSpends, err := k.GetAllEpochSpends(ctx)
	if err != nil {
		return nil, err
	}
	nextID, err := k.nextSpendID(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{Params: params, Allocations: allocations, Spends: spends, EpochSpends: epochSpends, NextSpendId: nextID}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

func (k Keeper) nextSpendID(ctx context.Context) (uint64, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.NextSpendIDKey)
	if err != nil || bz == nil {
		return 1, err
	}
	return binary.BigEndian.Uint64(bz), nil
}

func (k Keeper) setNextSpendID(ctx context.Context, id uint64) error {
	if id == 0 {
		return types.ErrInvalidSpend.Wrap("next spend id must be positive")
	}
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return k.storeService.OpenKVStore(ctx).Set(types.NextSpendIDKey, bz)
}

func addIncomingByPolicy(allocations types.TreasuryAllocations, params types.Params, delta sdk.Coins) types.TreasuryAllocations {
	for _, coin := range delta {
		reserve := types.BpsAmount(coin.Amount, params.ReserveBps)
		ecosystem := types.BpsAmount(coin.Amount, params.EcosystemBps)
		validators := types.BpsAmount(coin.Amount, params.ValidatorIncentivesBps)
		burn := types.BpsAmount(coin.Amount, params.BurnBps)
		allocated := reserve.Add(ecosystem).Add(validators).Add(burn)
		remainder := coin.Amount.Sub(allocated)
		allocations.ReserveBalance = addPositiveCoin(allocations.ReserveBalance, coin.Denom, reserve.Add(remainder))
		allocations.EcosystemBalance = addPositiveCoin(allocations.EcosystemBalance, coin.Denom, ecosystem)
		allocations.ValidatorIncentiveBalance = addPositiveCoin(allocations.ValidatorIncentiveBalance, coin.Denom, validators)
		allocations.BurnBalance = addPositiveCoin(allocations.BurnBalance, coin.Denom, burn)
	}
	allocations.TotalReceived = allocations.TotalReceived.Add(delta...)
	return allocations
}

func bucketBalance(allocations types.TreasuryAllocations, bucket string) sdk.Coins {
	switch bucket {
	case types.BucketReserve:
		return allocations.ReserveBalance
	case types.BucketEcosystem:
		return allocations.EcosystemBalance
	case types.BucketValidatorIncentives:
		return allocations.ValidatorIncentiveBalance
	default:
		return sdk.NewCoins()
	}
}

func subtractFromBucket(allocations types.TreasuryAllocations, bucket string, amount sdk.Coins) types.TreasuryAllocations {
	switch bucket {
	case types.BucketReserve:
		allocations.ReserveBalance = allocations.ReserveBalance.Sub(amount...)
	case types.BucketEcosystem:
		allocations.EcosystemBalance = allocations.EcosystemBalance.Sub(amount...)
	case types.BucketValidatorIncentives:
		allocations.ValidatorIncentiveBalance = allocations.ValidatorIncentiveBalance.Sub(amount...)
	}
	return allocations
}

func capAllows(params types.Params, spent sdk.Coins) bool {
	if spent.Empty() {
		return true
	}
	cap := sdk.NewCoins(params.PerEpochSpendCap)
	return cap.IsAllGTE(spent)
}

func addPositiveCoin(coins sdk.Coins, denom string, amount sdkmath.Int) sdk.Coins {
	if amount.IsNil() || !amount.IsPositive() {
		return coins
	}
	return coins.Add(sdk.NewCoin(denom, amount))
}

func spendKey(id uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.SpendPrefix[0]
	binary.BigEndian.PutUint64(key[1:], id)
	return key
}

func epochSpendKey(epoch uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.EpochSpendPrefix[0]
	binary.BigEndian.PutUint64(key[1:], epoch)
	return key
}
