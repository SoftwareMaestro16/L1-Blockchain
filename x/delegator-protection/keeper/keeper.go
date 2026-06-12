package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/x/delegator-protection/types"
)

type Keeper struct {
	storeService	corestore.KVStoreService
	authority	string
}

func NewKeeper(storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{storeService: storeService, authority: authority}
}

func (k Keeper) Authority() string	{ return k.authority }

func (k Keeper) DefaultGenesis() types.DelegatorProtectionState {
	params := types.DefaultProtectionParams()
	params.Authority = k.authority
	state, err := types.NewDelegatorProtectionState(params)
	if err != nil {
		panic(err)
	}
	return state
}

func (k Keeper) GetState(ctx context.Context) (types.DelegatorProtectionState, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.StateKey)
	if err != nil || bz == nil {
		return k.DefaultGenesis(), err
	}
	var state types.DelegatorProtectionState
	if err := json.Unmarshal(bz, &state); err != nil {
		return types.DelegatorProtectionState{}, err
	}
	state = types.NormalizeDelegatorProtectionState(state)
	if err := state.Validate(); err != nil {
		return types.DelegatorProtectionState{}, err
	}
	return state, nil
}

func (k Keeper) SetState(ctx context.Context, state types.DelegatorProtectionState) error {
	state = types.NormalizeDelegatorProtectionState(state)
	if err := state.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.StateKey, bz)
}

func (k Keeper) InitGenesis(ctx context.Context, state types.DelegatorProtectionState) error {
	if state.Params.Authority == "" || state.Params.Authority == types.DefaultProtectionAuthority {
		state.Params.Authority = k.authority
	}
	return k.SetState(ctx, state)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.DelegatorProtectionState, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	exported, err := types.ExportDelegatorProtectionState(state)
	if err != nil {
		return nil, err
	}
	return &exported, nil
}

func parseInt(value string) (sdkmath.Int, error) {
	if value == "" {
		return sdkmath.ZeroInt(), nil
	}
	out, ok := sdkmath.NewIntFromString(value)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid integer %q", value)
	}
	return out, nil
}

func mustJSON(v any) (string, error) {
	bz, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
