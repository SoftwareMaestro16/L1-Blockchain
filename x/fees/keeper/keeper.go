package keeper

import (
	"context"

	corestore "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sovereign-l1/l1/x/fees/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	authority    string
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, authority: authority}
}

func (k Keeper) Authority() string { return k.authority }

func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.ParamsKey, k.cdc.MustMarshal(&params))
}

func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.ParamsKey)
	if err != nil {
		return types.Params{}, err
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params, nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) {
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{Params: params}
}

func (k Keeper) IsAllowedFeeDenom(ctx context.Context, denom string) (bool, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}
	for _, allowed := range params.AllowedFeeDenoms {
		if denom == allowed {
			return true, nil
		}
	}
	return false, nil
}
