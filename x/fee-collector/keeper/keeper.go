package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	corestore "cosmossdk.io/core/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/fee-collector/types"
)

type Keeper struct {
	cdc		codec.BinaryCodec
	storeService	corestore.KVStoreService
	accountKeeper	types.AccountKeeper
	bankKeeper	types.BankKeeper
	authority	string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestore.KVStoreService,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:		cdc,
		storeService:	storeService,
		accountKeeper:	accountKeeper,
		bankKeeper:	bankKeeper,
		authority:	authority,
	}
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

func (k Keeper) SetFeeBalances(ctx context.Context, balances types.FeeBalances) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := balances.Validate(params.BaseDenom); err != nil {
		return types.ErrAccounting.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&balances)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.FeeBalancesKey, bz)
}

func (k Keeper) GetFeeBalances(ctx context.Context) (types.FeeBalances, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.FeeBalancesKey)
	if err != nil || bz == nil {
		return types.DefaultFeeBalances(), err
	}
	var balances types.FeeBalances
	if err := k.cdc.Unmarshal(bz, &balances); err != nil {
		return types.FeeBalances{}, err
	}
	return balances, nil
}

func (k Keeper) SetPendingDistribution(ctx context.Context, pending types.PendingDistribution) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := pending.Validate(params.BaseDenom); err != nil {
		return types.ErrAccounting.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&pending)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.PendingDistributionKey, bz)
}

func (k Keeper) GetPendingDistribution(ctx context.Context) (types.PendingDistribution, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.PendingDistributionKey)
	if err != nil || bz == nil {
		return types.DefaultPendingDistribution(), err
	}
	var pending types.PendingDistribution
	if err := k.cdc.Unmarshal(bz, &pending); err != nil {
		return types.PendingDistribution{}, err
	}
	return pending, nil
}

func (k Keeper) SetProtocolIncomePolicy(ctx context.Context, policy types.ProtocolIncomePolicy) error {
	policy = types.NormalizeProtocolIncomePolicy(policy)
	if err := policy.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	bz, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.ProtocolIncomePolicyKey, bz)
}

func (k Keeper) GetProtocolIncomePolicy(ctx context.Context) (types.ProtocolIncomePolicy, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.ProtocolIncomePolicyKey)
	if err != nil || bz == nil {
		return types.DefaultProtocolIncomePolicy(), err
	}
	var policy types.ProtocolIncomePolicy
	if err := json.Unmarshal(bz, &policy); err != nil {
		return types.ProtocolIncomePolicy{}, err
	}
	policy = types.NormalizeProtocolIncomePolicy(policy)
	if err := policy.Validate(); err != nil {
		return types.ProtocolIncomePolicy{}, types.ErrInvalidParams.Wrap(err.Error())
	}
	return policy, nil
}

func (k Keeper) CollectAndDistributeProtocolIncomeFromAccount(ctx context.Context, sender sdk.AccAddress, fees sdk.Coins) ([]types.ProtocolIncomeAllocation, sdk.Coins, error) {
	if len(sender) == 0 || aetraaddress.IsZeroAccAddress(sender) {
		return nil, nil, types.ErrInvalidFee.Wrap("fee sender must not be empty or zero")
	}
	policy, err := k.GetProtocolIncomePolicy(ctx)
	if err != nil {
		return nil, nil, err
	}
	allocations, roundingRemainder, err := types.SplitProtocolIncome(policy, fees)
	if err != nil {
		return nil, nil, err
	}
	if err := k.assertProtocolIncomeModuleAccounts(policy); err != nil {
		return nil, nil, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	if err := k.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, sender, types.CollectorModuleName, fees); err != nil {
		return nil, nil, err
	}
	for _, allocation := range allocations {
		if allocation.Amount.Empty() {
			continue
		}
		if allocation.Burn {
			if err := k.bankKeeper.BurnCoins(cacheCtx, types.CollectorModuleName, allocation.Amount); err != nil {
				return nil, nil, err
			}
			continue
		}
		if err := k.bankKeeper.SendCoinsFromModuleToModule(cacheCtx, types.CollectorModuleName, allocation.ModuleAccount, allocation.Amount); err != nil {
			return nil, nil, err
		}
	}
	collectorBalance := k.bankKeeper.GetAllBalances(cacheCtx, k.accountKeeper.GetModuleAddress(types.CollectorModuleName))
	if !collectorBalance.Empty() {
		return nil, nil, types.ErrAccounting.Wrapf("collector balance must be empty after protocol income distribution: %s", collectorBalance)
	}
	write()
	return allocations, roundingRemainder, nil
}

func (k Keeper) CollectFeesFromAccount(ctx context.Context, sender sdk.AccAddress, fees sdk.Coins, feeType string) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := types.ValidateFeeCoins(params.BaseDenom, fees); err != nil {
		return err
	}
	if len(sender) == 0 || aetraaddress.IsZeroAccAddress(sender) {
		return types.ErrInvalidFee.Wrap("fee sender must not be empty or zero")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	if err := k.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, sender, types.CollectorModuleName, fees); err != nil {
		return err
	}
	if err := k.recordCollectedFees(cacheCtx, fees, feeType); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) RecordCollectedFees(ctx context.Context, fees sdk.Coins, feeType string) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := types.ValidateFeeCoins(params.BaseDenom, fees); err != nil {
		return err
	}
	return k.recordCollectedFees(ctx, fees, feeType)
}

func (k Keeper) recordCollectedFees(ctx context.Context, fees sdk.Coins, feeType string) error {
	if feeType == "" {
		feeType = types.FeeTypeProtocol
	}
	balances, err := k.GetFeeBalances(ctx)
	if err != nil {
		return err
	}
	switch feeType {
	case types.FeeTypeGas:
		balances.GasFees = balances.GasFees.Add(fees...)
	case types.FeeTypeForwarding:
		balances.ForwardingFees = balances.ForwardingFees.Add(fees...)
	case types.FeeTypeProtocol:
		balances.ProtocolFees = balances.ProtocolFees.Add(fees...)
	default:
		return types.ErrInvalidFee.Wrapf("unknown fee type %s", feeType)
	}
	balances.TotalCollected = balances.TotalCollected.Add(fees...)

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	split, _, err := types.SplitFees(params, fees)
	if err != nil {
		return err
	}
	pending, err := k.GetPendingDistribution(ctx)
	if err != nil {
		return err
	}
	pending.Treasury = pending.Treasury.Add(split.Treasury...)
	pending.Protection = pending.Protection.Add(split.Protection...)
	pending.Validators = pending.Validators.Add(split.Validators...)
	pending.Burn = pending.Burn.Add(split.Burn...)

	if err := k.SetFeeBalances(ctx, balances); err != nil {
		return err
	}
	if err := k.SetPendingDistribution(ctx, pending); err != nil {
		return err
	}
	if err := k.AssertModuleAccountingInvariant(ctx); err != nil {
		return err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCollectFees,
		sdk.NewAttribute(types.AttributeKeyFeeType, feeType),
		sdk.NewAttribute(types.AttributeKeyAmount, fees.String()),
		sdk.NewAttribute(types.AttributeKeyBurn, split.Burn.String()),
	))
	return nil
}

func (k Keeper) DistributeFees(ctx context.Context, epoch uint64) (types.FeeHistoryEntry, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	history, err := k.distributeFees(cacheCtx, epoch)
	if err != nil {
		return types.FeeHistoryEntry{}, err
	}
	write()
	return history, nil
}

func (k Keeper) distributeFees(ctx context.Context, epoch uint64) (types.FeeHistoryEntry, error) {
	if _, found, err := k.GetFeeHistory(ctx, epoch); err != nil {
		return types.FeeHistoryEntry{}, err
	} else if found {
		return types.FeeHistoryEntry{}, types.ErrDuplicateHistory.Wrapf("epoch %d", epoch)
	}
	if err := k.AssertModuleAccountingInvariant(ctx); err != nil {
		return types.FeeHistoryEntry{}, err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.FeeHistoryEntry{}, err
	}
	pending, err := k.GetPendingDistribution(ctx)
	if err != nil {
		return types.FeeHistoryEntry{}, err
	}
	collected := pending.Total()
	if collected.Empty() {
		return types.FeeHistoryEntry{}, types.ErrEmptyDistribution.Wrap("no pending fees")
	}

	if !pending.Treasury.Empty() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, params.CollectorModule, params.TreasuryModule, pending.Treasury); err != nil {
			return types.FeeHistoryEntry{}, err
		}
	}
	if !pending.Protection.Empty() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, params.CollectorModule, params.ProtectionModule, pending.Protection); err != nil {
			return types.FeeHistoryEntry{}, err
		}
	}
	if !pending.Validators.Empty() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, params.CollectorModule, params.ValidatorsModule, pending.Validators); err != nil {
			return types.FeeHistoryEntry{}, err
		}
	}
	if !pending.Burn.Empty() {
		if err := k.bankKeeper.BurnCoins(ctx, params.CollectorModule, pending.Burn); err != nil {
			return types.FeeHistoryEntry{}, err
		}
	}

	balances, err := k.GetFeeBalances(ctx)
	if err != nil {
		return types.FeeHistoryEntry{}, err
	}
	if !balances.AccountingBalance().Equal(collected) {
		return types.FeeHistoryEntry{}, types.ErrAccounting.Wrapf("accounting balance %s does not match pending %s", balances.AccountingBalance(), collected)
	}
	balances.GasFees = sdk.NewCoins()
	balances.ForwardingFees = sdk.NewCoins()
	balances.ProtocolFees = sdk.NewCoins()
	balances.TotalDistributed = balances.TotalDistributed.Add(pending.Treasury...).Add(pending.Protection...).Add(pending.Validators...)
	balances.TotalBurned = balances.TotalBurned.Add(pending.Burn...)
	totalBurned := balances.TotalBurned
	if err := k.SetFeeBalances(ctx, balances); err != nil {
		return types.FeeHistoryEntry{}, err
	}

	_, remainder, err := types.SplitFees(params, collected)
	if err != nil {
		return types.FeeHistoryEntry{}, err
	}
	history := types.FeeHistoryEntry{
		Epoch:			epoch,
		Collected:		collected,
		Treasury:		pending.Treasury,
		Protection:		pending.Protection,
		Validators:		pending.Validators,
		Burn:			pending.Burn,
		RoundingRemainder:	remainder,
		DistributedAtHeight:	sdk.UnwrapSDKContext(ctx).BlockHeight(),
	}
	if err := k.SetFeeHistory(ctx, history); err != nil {
		return types.FeeHistoryEntry{}, err
	}
	if err := k.SetPendingDistribution(ctx, types.DefaultPendingDistribution()); err != nil {
		return types.FeeHistoryEntry{}, err
	}
	if err := k.AssertModuleAccountingInvariant(ctx); err != nil {
		return types.FeeHistoryEntry{}, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDistributeFees,
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute(types.AttributeKeyAmount, collected.String()),
		sdk.NewAttribute(types.AttributeKeyBurn, pending.Burn.String()),
		sdk.NewAttribute(types.AttributeKeyTotalBurn, totalBurned.String()),
	))
	return history, nil
}

func (k Keeper) AssertModuleAccountingInvariant(ctx context.Context) error {
	balances, err := k.GetFeeBalances(ctx)
	if err != nil {
		return err
	}
	pending, err := k.GetPendingDistribution(ctx)
	if err != nil {
		return err
	}
	accountingBalance := balances.AccountingBalance()
	if !accountingBalance.Equal(pending.Total()) {
		return types.ErrAccounting.Wrapf("accounting balance %s != pending %s", accountingBalance, pending.Total())
	}
	moduleAddr := k.accountKeeper.GetModuleAddress(types.CollectorModuleName)
	if moduleAddr == nil {
		return types.ErrAccounting.Wrapf("module account %s is not configured", types.CollectorModuleName)
	}
	bankBalance := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	if !bankBalance.Equal(accountingBalance) {
		return types.ErrAccounting.Wrapf("module bank balance %s != accounting state %s", bankBalance, accountingBalance)
	}
	return nil
}

func (k Keeper) assertProtocolIncomeModuleAccounts(policy types.ProtocolIncomePolicy) error {
	if k.accountKeeper.GetModuleAddress(types.CollectorModuleName) == nil {
		return types.ErrAccounting.Wrapf("module account %s is not configured", types.CollectorModuleName)
	}
	for _, bucket := range types.NormalizeProtocolIncomePolicy(policy).Buckets {
		if k.accountKeeper.GetModuleAddress(bucket.ModuleAccount) == nil {
			return types.ErrAccounting.Wrapf("module account %s for bucket %s is not configured", bucket.ModuleAccount, bucket.Bucket)
		}
	}
	return nil
}

func (k Keeper) ModuleAccountAddress() string {
	addr := k.accountKeeper.GetModuleAddress(types.CollectorModuleName)
	if addr == nil {
		return ""
	}
	return aetraaddress.FormatAccAddress(addr)
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	gs.Params = types.NormalizeParams(gs.Params)
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	if err := k.SetFeeBalances(ctx, gs.Balances); err != nil {
		return err
	}
	if err := k.SetPendingDistribution(ctx, gs.PendingDistribution); err != nil {
		return err
	}
	for _, entry := range gs.FeeHistory {
		if err := k.SetFeeHistory(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	balances, err := k.GetFeeBalances(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := k.GetPendingDistribution(ctx)
	if err != nil {
		return nil, err
	}
	history, err := k.GetAllFeeHistory(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{Params: params, Balances: balances, PendingDistribution: pending, FeeHistory: history}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

func (k Keeper) SetFeeHistory(ctx context.Context, entry types.FeeHistoryEntry) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := entry.Validate(params.BaseDenom); err != nil {
		return err
	}
	bz, err := k.cdc.Marshal(&entry)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(feeHistoryKey(entry.Epoch), bz)
}

func (k Keeper) GetFeeHistory(ctx context.Context, epoch uint64) (types.FeeHistoryEntry, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(feeHistoryKey(epoch))
	if err != nil || bz == nil {
		return types.FeeHistoryEntry{}, false, err
	}
	var entry types.FeeHistoryEntry
	if err := k.cdc.Unmarshal(bz, &entry); err != nil {
		return types.FeeHistoryEntry{}, false, err
	}
	return entry, true, nil
}

func (k Keeper) GetAllFeeHistory(ctx context.Context) ([]types.FeeHistoryEntry, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.FeeHistoryPrefix, storetypes.PrefixEndBytes(types.FeeHistoryPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.FeeHistoryEntry{}
	for ; iter.Valid(); iter.Next() {
		var entry types.FeeHistoryEntry
		if err := k.cdc.Unmarshal(iter.Value(), &entry); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, nil
}

func feeHistoryKey(epoch uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.FeeHistoryPrefix[0]
	binary.BigEndian.PutUint64(key[1:], epoch)
	return key
}
