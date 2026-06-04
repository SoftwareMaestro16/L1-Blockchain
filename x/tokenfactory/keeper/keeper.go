package keeper

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	corestore "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

var subdenomRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9/:._-]{2,63}$`)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	bankKeeper   types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, bankKeeper types.BankKeeper) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, bankKeeper: bankKeeper}
}

func denomKey(denom string) []byte {
	return append(types.DenomPrefix, []byte(denom)...)
}

func (k Keeper) SetDenom(ctx context.Context, meta types.DenomAuthorityMetadata) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(denomKey(meta.Denom), k.cdc.MustMarshal(&meta))
}

func (k Keeper) GetDenom(ctx context.Context, denom string) (types.DenomAuthorityMetadata, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(denomKey(denom))
	if err != nil || bz == nil {
		return types.DenomAuthorityMetadata{}, false, err
	}
	var meta types.DenomAuthorityMetadata
	k.cdc.MustUnmarshal(bz, &meta)
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
		k.cdc.MustUnmarshal(iter.Value(), &meta)
		out = append(out, meta)
	}
	return out, nil
}

func (k Keeper) FullDenom(creator, subdenom string) (string, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return "", err
	}
	if !subdenomRe.MatchString(subdenom) || strings.Contains(subdenom, "//") {
		return "", types.ErrInvalidDenom.Wrap("subdenom must be 3-64 chars and start with a letter")
	}
	return fmt.Sprintf("%s/%s/%s", types.FactoryDenomPrefix, creatorAddr.String(), subdenom), nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) {
	for _, meta := range gs.Denoms {
		if err := k.SetDenom(ctx, meta); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	denoms, err := k.GetAllDenoms(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{Denoms: denoms}
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
