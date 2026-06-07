package keeper

import (
	"context"
	"encoding/json"

	corestore "cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/reputation/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	authority    string
}

func NewKeeper(storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{storeService: storeService, authority: authority}
}

func (k Keeper) Authority() string { return k.authority }

func (k Keeper) DefaultGenesis() types.ReputationState {
	params := types.DefaultReputationParams()
	params.Authority = k.authority
	state, err := types.NewReputationState(params)
	if err != nil {
		panic(err)
	}
	return state
}

func (k Keeper) GetState(ctx context.Context) (types.ReputationState, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.StateKey)
	if err != nil || bz == nil {
		return k.DefaultGenesis(), err
	}
	var state types.ReputationState
	if err := json.Unmarshal(bz, &state); err != nil {
		return types.ReputationState{}, err
	}
	state = types.NormalizeReputationState(state)
	if err := state.Validate(); err != nil {
		return types.ReputationState{}, err
	}
	return state, nil
}

func (k Keeper) SetState(ctx context.Context, state types.ReputationState) error {
	state = types.NormalizeReputationState(state)
	if err := state.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.StateKey, bz)
}

func (k Keeper) InitGenesis(ctx context.Context, state types.ReputationState) error {
	if state.Params.Authority == "" || state.Params.Authority == types.DefaultReputationAuthority {
		state.Params.Authority = k.authority
	}
	return k.SetState(ctx, state)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.ReputationState, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	exported, err := types.ExportReputationState(state)
	if err != nil {
		return nil, err
	}
	return &exported, nil
}

func parseAddress(text string) (sdk.AccAddress, error) {
	return aetraaddress.ParseAccAddress(text)
}

func mustJSON(v any) (string, error) {
	bz, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
