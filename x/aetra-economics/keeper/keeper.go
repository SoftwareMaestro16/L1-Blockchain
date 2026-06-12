package keeper

import (
	"context"
	"encoding/json"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/aetra-economics/types"
)

var genesisKey = []byte{0x01}

type Keeper struct {
	state		types.GenesisState
	storeService	corestore.KVStoreService
	runtimeCtx	context.Context
}

func NewKeeper(authority string) Keeper {
	return Keeper{state: types.DefaultGenesisState(authority)}
}

func NewPersistentKeeper(storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{state: types.DefaultGenesisState(authority), storeService: storeService}
}

func (k Keeper) Authority() string {
	return k.state.Params.Authority
}

func (k Keeper) Params() types.Params {
	return k.state.Params
}

func (k *Keeper) SetParams(params types.Params) error {
	if err := params.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	next := k.state
	next.Params = params
	if err := next.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	k.state = next
	return k.save()
}

func (k *Keeper) ApplyEpoch(input types.EpochEconomicsInput) (types.EpochRewardSummary, error) {
	next, summary, err := types.ApplyEpoch(k.state.Params, k.state.State, input)
	if err != nil {
		return types.EpochRewardSummary{}, err
	}
	k.state.State = next
	return summary, k.save()
}

func (k Keeper) QueryCurrentInflation(req types.QueryCurrentInflationRequest) (types.QueryCurrentInflationResponse, error) {
	return types.QueryCurrentInflationResponse{InflationBps: k.state.State.CurrentInflationBps}, nil
}

func (k Keeper) QueryCurrentBondedRatio(req types.QueryCurrentBondedRatioRequest) (types.QueryCurrentBondedRatioResponse, error) {
	return types.QueryCurrentBondedRatioResponse{BondedRatioBps: k.state.State.CurrentBondedRatioBps}, nil
}

func (k Keeper) QueryEstimatedAPR(req types.QueryEstimatedAPRRequest) (types.QueryEstimatedAPRResponse, error) {
	return types.EstimateAPRBreakdown(k.state.Params, k.state.State, req)
}

func (k Keeper) QueryFeeSplitParams(req types.QueryFeeSplitParamsRequest) (types.QueryFeeSplitParamsResponse, error) {
	params := k.state.Params
	return types.QueryFeeSplitParamsResponse{
		BurnMinBps:			params.BurnMinBps,
		BurnMaxBps:			params.BurnMaxBps,
		BurnCurrentBps:			params.BurnCurrentBps,
		ValidatorRewardMinBps:		params.ValidatorRewardMinBps,
		ValidatorRewardMaxBps:		params.ValidatorRewardMaxBps,
		ValidatorRewardBps:		params.ValidatorRewardBps,
		TreasuryMinBps:			params.TreasuryMinBps,
		TreasuryMaxBps:			params.TreasuryMaxBps,
		TreasuryBps:			params.TreasuryBps,
		EmergencyAllowZeroRewardShare:	params.EmergencyAllowZeroRewardShare,
	}, nil
}

func (k Keeper) QueryBurnedSupply(req types.QueryBurnedSupplyRequest) (types.QueryBurnedSupplyResponse, error) {
	return types.QueryBurnedSupplyResponse{BurnedSupply: k.state.State.BurnedSupply}, nil
}

func (k Keeper) QueryTreasuryBalance(req types.QueryTreasuryBalanceRequest) (types.QueryTreasuryBalanceResponse, error) {
	return types.QueryTreasuryBalanceResponse{TreasuryBalance: k.state.State.TreasuryBalance}, nil
}

func (k Keeper) QueryEpochRewardSummary(req types.QueryEpochRewardSummaryRequest) (types.QueryEpochRewardSummaryResponse, error) {
	for _, summary := range k.state.State.RewardHistory {
		if summary.Epoch == req.Epoch {
			return types.QueryEpochRewardSummaryResponse{Summary: summary}, nil
		}
	}
	return types.QueryEpochRewardSummaryResponse{}, types.ErrNotFound
}

func (k *Keeper) InitGenesis(gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.state = gs
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.state = gs
	k.runtimeCtx = ctx
	if k.storeService == nil {
		return nil
	}
	return k.saveWithCtx(ctx)
}

func (k Keeper) ExportGenesis() (types.GenesisState, error) {
	if err := k.state.Validate(); err != nil {
		return types.GenesisState{}, err
	}
	return k.state, nil
}

func (k Keeper) ExportGenesisState(ctx context.Context) (types.GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis()
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return types.GenesisState{}, err
	}
	if len(bz) == 0 {
		return k.ExportGenesis()
	}
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return types.GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return types.GenesisState{}, err
	}
	return gs, nil
}

func (k Keeper) MarshalGenesis() ([]byte, error) {
	gs, err := k.ExportGenesis()
	if err != nil {
		return nil, err
	}
	return json.Marshal(gs)
}

func (k *Keeper) UnmarshalGenesis(bz []byte) error {
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return err
	}
	return k.InitGenesis(gs)
}

func (k *Keeper) save() error {
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return k.saveWithCtx(k.runtimeCtx)
}

func (k *Keeper) saveWithCtx(ctx context.Context) error {
	bz, err := json.Marshal(k.state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}
