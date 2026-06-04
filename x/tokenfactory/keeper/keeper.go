package keeper

import (
	"context"
	"fmt"

	corestore "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	bankKeeper   types.BankKeeper
	authority    string
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, bankKeeper types.BankKeeper, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, bankKeeper: bankKeeper, authority: authority}
}

func (k Keeper) Authority() string { return k.authority }

func denomKey(denom string) []byte {
	return append(types.DenomPrefix, []byte(denom)...)
}

func (k Keeper) SetDenom(ctx context.Context, meta types.DenomAuthorityMetadata) error {
	bz, err := k.cdc.Marshal(&meta)
	if err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(denomKey(meta.Denom), bz)
}

func (k Keeper) GetDenom(ctx context.Context, denom string) (types.DenomAuthorityMetadata, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(denomKey(denom))
	if err != nil || bz == nil {
		return types.DenomAuthorityMetadata{}, false, err
	}
	var meta types.DenomAuthorityMetadata
	if err := k.cdc.Unmarshal(bz, &meta); err != nil {
		return types.DenomAuthorityMetadata{}, false, err
	}
	return meta, true, nil
}

func (k Keeper) GetAllDenoms(ctx context.Context) ([]types.DenomAuthorityMetadata, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.DenomPrefix, storetypes.PrefixEndBytes(types.DenomPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var out []types.DenomAuthorityMetadata
	for ; iter.Valid(); iter.Next() {
		var meta types.DenomAuthorityMetadata
		if err := k.cdc.Unmarshal(iter.Value(), &meta); err != nil {
			return nil, err
		}
		out = append(out, meta)
	}
	return out, nil
}

func (k Keeper) FullDenom(ctx context.Context, creator, subdenom string) (string, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return "", err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return "", err
	}
	if err := types.ValidateSubdenom(subdenom, params); err != nil {
		return "", types.ErrInvalidDenom.Wrap(err.Error())
	}
	return fmt.Sprintf("%s/%s/%s", types.FactoryDenomPrefix, creatorAddr.String(), subdenom), nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	for _, meta := range gs.Denoms {
		if err := k.SetDenom(ctx, meta); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	denoms, err := k.GetAllDenoms(ctx)
	if err != nil {
		return nil, err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{Denoms: denoms, Params: params}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

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

func BankMetadata(denom string) banktypes.Metadata {
	return banktypes.Metadata{
		Base:        denom,
		Display:     denom,
		Name:        denom,
		Symbol:      denom,
		Description: "factory token " + denom,
		DenomUnits:  []*banktypes.DenomUnit{{Denom: denom, Exponent: 0}},
	}
}
