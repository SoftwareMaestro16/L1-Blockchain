package keeper

import (
	"context"
	"encoding/json"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/aetra-validator-score/types"
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

func (k *Keeper) UpdateScores(epoch uint64, metrics []types.ValidatorMetricInput) ([]types.ValidatorScore, error) {
	canonicalMetrics, err := types.CanonicalMetricInputs(metrics)
	if err != nil {
		return nil, err
	}
	scores, err := types.ComputeValidatorScores(k.state.Params, epoch, canonicalMetrics)
	if err != nil {
		return nil, err
	}
	next := k.state
	next.Epoch = epoch
	next.Metrics = canonicalMetrics
	next.Scores = scores
	if err := next.Validate(); err != nil {
		return nil, types.ErrInvalidScore.Wrap(err.Error())
	}
	k.state = next
	return scores, k.save()
}

func (k Keeper) QueryParams(req types.QueryParamsRequest) (types.QueryParamsResponse, error) {
	return types.QueryParamsResponse{Params: k.state.Params}, nil
}

func (k Keeper) QueryValidatorScore(req types.QueryValidatorScoreRequest) (types.QueryValidatorScoreResponse, error) {
	score, found := k.findScore(req.OperatorAddress)
	if !found {
		return types.QueryValidatorScoreResponse{}, types.ErrNotFound
	}
	return types.QueryValidatorScoreResponse{Score: score}, nil
}

func (k Keeper) QueryPublicValidatorMetrics(req types.QueryPublicValidatorMetricsRequest) (types.QueryPublicValidatorMetricsResponse, error) {
	score, found := k.findScore(req.OperatorAddress)
	if !found {
		return types.QueryPublicValidatorMetricsResponse{}, types.ErrNotFound
	}
	return types.QueryPublicValidatorMetricsResponse{Metrics: types.PublicMetricsFromScore(score)}, nil
}

func (k Keeper) QueryAllValidatorScores(req types.QueryAllValidatorScoresRequest) (types.QueryAllValidatorScoresResponse, error) {
	return types.QueryAllValidatorScoresResponse{Scores: append([]types.ValidatorScore(nil), k.state.Scores...)}, nil
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

func (k Keeper) findScore(operator string) (types.ValidatorScore, bool) {
	for _, score := range k.state.Scores {
		if score.OperatorAddress == operator {
			return score, true
		}
	}
	return types.ValidatorScore{}, false
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
