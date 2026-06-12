package keeper

import (
	"context"
	"encoding/binary"

	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/sovereign-l1/l1/x/dynamic-commission/types"
)

type Keeper struct {
	cdc			codec.BinaryCodec
	storeService		corestore.KVStoreService
	authority		string
	reputationKeeper	types.ReputationKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, authority: authority}
}

func (k Keeper) WithReputationKeeper(rk types.ReputationKeeper) Keeper {
	k.reputationKeeper = rk
	return k
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

func (k Keeper) SetBaseCommission(ctx context.Context, validator string, baseBps uint32, height uint64) (types.ValidatorCommission, error) {
	if height == 0 {
		return types.ValidatorCommission{}, types.ErrInvalidCommission.Wrap("height must be positive")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.ValidatorCommission{}, err
	}
	previous, found, err := k.GetValidatorCommission(ctx, validator)
	if err != nil {
		return types.ValidatorCommission{}, err
	}
	if !found {
		previous = types.DefaultCommission(validator, params)
	}
	next := previous
	next.BaseCommissionBps = baseBps
	next.CommissionFloorBps = params.CommissionFloorBps
	next.CommissionCeilingBps = params.CommissionCeilingBps
	next.EffectiveCommissionBps = types.EffectiveCommission(params, baseBps, next.PerformanceModifierBps, next.ReputationModifierBps)
	next.LastUpdateHeight = height
	if err := next.Validate(params); err != nil {
		return types.ValidatorCommission{}, types.ErrInvalidCommission.Wrap(err.Error())
	}
	if found && types.RateLimitExceeded(previous.EffectiveCommissionBps, next.EffectiveCommissionBps, params.MaxRateChangeBps) {
		return types.ValidatorCommission{}, types.ErrRateLimited.Wrap("base commission change exceeds max_rate_change_bps")
	}
	return k.commitCommission(ctx, next, height, "base_commission_set")
}

func (k Keeper) RecomputeEffectiveCommission(ctx context.Context, validator string, performanceScoreBps, reputationScoreBps uint32, jailed bool, height uint64) (types.ValidatorCommission, error) {
	if height == 0 {
		return types.ValidatorCommission{}, types.ErrInvalidCommission.Wrap("height must be positive")
	}
	if k.reputationKeeper != nil {
		score, isJailed, err := k.reputationKeeper.GetValidatorTotalScore(ctx, validator)
		if err == nil {
			reputationScoreBps = score
			jailed = jailed || isJailed
		}
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.ValidatorCommission{}, err
	}
	previous, found, err := k.GetValidatorCommission(ctx, validator)
	if err != nil {
		return types.ValidatorCommission{}, err
	}
	if !found {
		previous = types.DefaultCommission(validator, params)
	}
	performanceModifier, reputationModifier, err := types.ComputeModifiers(params, performanceScoreBps, reputationScoreBps, jailed)
	if err != nil {
		return types.ValidatorCommission{}, types.ErrInvalidCommission.Wrap(err.Error())
	}
	next := previous
	next.PerformanceModifierBps = performanceModifier
	next.ReputationModifierBps = reputationModifier
	next.CommissionFloorBps = params.CommissionFloorBps
	next.CommissionCeilingBps = params.CommissionCeilingBps
	next.EffectiveCommissionBps = types.EffectiveCommission(params, next.BaseCommissionBps, performanceModifier, reputationModifier)
	next.LastUpdateHeight = height
	next.Jailed = jailed
	if err := next.Validate(params); err != nil {
		return types.ValidatorCommission{}, types.ErrInvalidCommission.Wrap(err.Error())
	}
	if found && types.RateLimitExceeded(previous.EffectiveCommissionBps, next.EffectiveCommissionBps, params.MaxRateChangeBps) {
		return types.ValidatorCommission{}, types.ErrRateLimited.Wrap("effective commission change exceeds max_rate_change_bps")
	}
	return k.commitCommission(ctx, next, height, "effective_commission_recomputed")
}

func (k Keeper) commitCommission(ctx context.Context, commission types.ValidatorCommission, height uint64, reason string) (types.ValidatorCommission, error) {
	if err := k.SetValidatorCommission(ctx, commission); err != nil {
		return types.ValidatorCommission{}, err
	}
	entry := types.CommissionHistoryEntry{
		ValidatorAddress:	commission.ValidatorAddress,
		Height:			height,
		BaseCommissionBps:	commission.BaseCommissionBps,
		EffectiveCommissionBps:	commission.EffectiveCommissionBps,
		PerformanceModifierBps:	commission.PerformanceModifierBps,
		ReputationModifierBps:	commission.ReputationModifierBps,
		Jailed:			commission.Jailed,
		Reason:			reason,
	}
	if err := k.SetCommissionHistory(ctx, entry); err != nil {
		return types.ValidatorCommission{}, err
	}
	return commission, nil
}

func (k Keeper) SetValidatorCommission(ctx context.Context, commission types.ValidatorCommission) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := commission.Validate(params); err != nil {
		return types.ErrInvalidCommission.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&commission)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(commissionKey(commission.ValidatorAddress), bz)
}

func (k Keeper) GetValidatorCommission(ctx context.Context, validator string) (types.ValidatorCommission, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(commissionKey(validator))
	if err != nil || bz == nil {
		return types.ValidatorCommission{}, false, err
	}
	var commission types.ValidatorCommission
	if err := k.cdc.Unmarshal(bz, &commission); err != nil {
		return types.ValidatorCommission{}, false, err
	}
	return commission, true, nil
}

func (k Keeper) GetAllValidatorCommissions(ctx context.Context) ([]types.ValidatorCommission, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.CommissionPrefix, storetypes.PrefixEndBytes(types.CommissionPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.ValidatorCommission{}
	for ; iter.Valid(); iter.Next() {
		var commission types.ValidatorCommission
		if err := k.cdc.Unmarshal(iter.Value(), &commission); err != nil {
			return nil, err
		}
		out = append(out, commission)
	}
	return types.SortCommissions(out), nil
}

func (k Keeper) SetCommissionHistory(ctx context.Context, entry types.CommissionHistoryEntry) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := entry.Validate(params); err != nil {
		return types.ErrInvalidCommission.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&entry)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(historyKey(entry.ValidatorAddress, entry.Height, entry.Reason), bz)
}

func (k Keeper) GetCommissionHistory(ctx context.Context, validator string) ([]types.CommissionHistoryEntry, error) {
	store := k.storeService.OpenKVStore(ctx)
	start := historyPrefixForValidator(validator)
	iter, err := store.Iterator(start, storetypes.PrefixEndBytes(start))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.CommissionHistoryEntry{}
	for ; iter.Valid(); iter.Next() {
		var entry types.CommissionHistoryEntry
		if err := k.cdc.Unmarshal(iter.Value(), &entry); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return types.SortHistory(out), nil
}

func (k Keeper) GetAllCommissionHistory(ctx context.Context) ([]types.CommissionHistoryEntry, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.HistoryPrefix, storetypes.PrefixEndBytes(types.HistoryPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.CommissionHistoryEntry{}
	for ; iter.Valid(); iter.Next() {
		var entry types.CommissionHistoryEntry
		if err := k.cdc.Unmarshal(iter.Value(), &entry); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return types.SortHistory(out), nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	gs.Params = types.NormalizeParams(gs.Params)
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	for _, commission := range gs.Commissions {
		if err := k.SetValidatorCommission(ctx, commission); err != nil {
			return err
		}
	}
	for _, entry := range gs.CommissionHistory {
		if err := k.SetCommissionHistory(ctx, entry); err != nil {
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
	commissions, err := k.GetAllValidatorCommissions(ctx)
	if err != nil {
		return nil, err
	}
	history, err := k.GetAllCommissionHistory(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{Params: params, Commissions: commissions, CommissionHistory: history}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

func commissionKey(validator string) []byte {
	key := append([]byte{}, types.CommissionPrefix...)
	key = append(key, []byte(validator)...)
	return key
}

func historyPrefixForValidator(validator string) []byte {
	key := append([]byte{}, types.HistoryPrefix...)
	key = append(key, []byte(validator)...)
	key = append(key, 0x00)
	return key
}

func historyKey(validator string, height uint64, reason string) []byte {
	key := historyPrefixForValidator(validator)
	heightBz := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBz, height)
	key = append(key, heightBz...)
	key = append(key, 0x00)
	key = append(key, []byte(reason)...)
	return key
}
