package keeper

import (
	"context"
	"encoding/json"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/performance/types"
)

type Keeper struct {
	storeService	corestore.KVStoreService
	authority	string
}

func NewKeeper(storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{storeService: storeService, authority: authority}
}

func (k Keeper) Authority() string	{ return k.authority }

func (k Keeper) DefaultGenesis() types.PerformanceOracleState {
	params := types.DefaultPerformanceOracleParams()
	params.Authority = k.authority
	state, err := types.NewPerformanceOracleState(params)
	if err != nil {
		panic(err)
	}
	return state
}

func (k Keeper) GetState(ctx context.Context) (types.PerformanceOracleState, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.StateKey)
	if err != nil || bz == nil {
		return k.DefaultGenesis(), err
	}
	var state types.PerformanceOracleState
	if err := json.Unmarshal(bz, &state); err != nil {
		return types.PerformanceOracleState{}, err
	}
	state = types.NormalizePerformanceOracleState(state)
	if err := state.Validate(); err != nil {
		return types.PerformanceOracleState{}, err
	}
	return state, nil
}

func (k Keeper) SetState(ctx context.Context, state types.PerformanceOracleState) error {
	state = types.NormalizePerformanceOracleState(state)
	if err := state.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.StateKey, bz)
}

func (k Keeper) InitGenesis(ctx context.Context, state types.PerformanceOracleState) error {
	if state.Params.Authority == "" || state.Params.Authority == types.DefaultPerformanceOracleAuthority {
		state.Params.Authority = k.authority
	}
	return k.SetState(ctx, state)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.PerformanceOracleState, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	exported, err := types.ExportPerformanceOracleState(state)
	if err != nil {
		return nil, err
	}
	return &exported, nil
}

func mustJSON(v any) (string, error) {
	bz, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
