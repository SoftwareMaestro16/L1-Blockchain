package keeper

import (
	"context"
	"encoding/binary"

	corestore "cosmossdk.io/core/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/emissions/types"
)

type Keeper struct {
	cdc		codec.BinaryCodec
	storeService	corestore.KVStoreService
	authority	string
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, authority: authority}
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

func (k Keeper) FinalizeEmissionEpoch(ctx context.Context, epoch uint64, stakingRatioBps uint32) (types.EmissionEpoch, error) {
	if epoch == 0 {
		return types.EmissionEpoch{}, types.ErrInvalidEpoch.Wrap("epoch must be positive")
	}
	if _, found, err := k.GetEmissionEpoch(ctx, epoch); err != nil {
		return types.EmissionEpoch{}, err
	} else if found {
		return types.EmissionEpoch{}, types.ErrDuplicateEpoch.Wrapf("epoch %d", epoch)
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.EmissionEpoch{}, err
	}
	record, err := types.ComputeEpochEmission(params, epoch, uint64(stakingRatioBps), sdk.UnwrapSDKContext(ctx).BlockHeight())
	if err != nil {
		return types.EmissionEpoch{}, err
	}
	if err := k.SetEmissionEpoch(ctx, record); err != nil {
		return types.EmissionEpoch{}, err
	}
	params.CurrentInflationBps = record.InflationBps
	if err := k.SetParams(ctx, params); err != nil {
		return types.EmissionEpoch{}, err
	}
	total, err := k.GetTotalMintedAccounting(ctx)
	if err != nil {
		return types.EmissionEpoch{}, err
	}
	total = sdk.NewCoin(params.BaseDenom, total.Amount.Add(record.EmissionAmount.Amount))
	if err := k.SetTotalMintedAccounting(ctx, total); err != nil {
		return types.EmissionEpoch{}, err
	}
	return record, nil
}

func (k Keeper) SetEmissionEpoch(ctx context.Context, record types.EmissionEpoch) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := record.Validate(params); err != nil {
		return types.ErrInvalidEpoch.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(epochKey(record.Epoch), bz)
}

func (k Keeper) GetEmissionEpoch(ctx context.Context, epoch uint64) (types.EmissionEpoch, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(epochKey(epoch))
	if err != nil || bz == nil {
		return types.EmissionEpoch{}, false, err
	}
	var record types.EmissionEpoch
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return types.EmissionEpoch{}, false, err
	}
	return record, true, nil
}

func (k Keeper) GetAllEmissionEpochs(ctx context.Context) ([]types.EmissionEpoch, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.EpochPrefix, storetypes.PrefixEndBytes(types.EpochPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.EmissionEpoch{}
	for ; iter.Valid(); iter.Next() {
		var record types.EmissionEpoch
		if err := k.cdc.Unmarshal(iter.Value(), &record); err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return types.SortEmissionEpochs(out), nil
}

func (k Keeper) SetTotalMintedAccounting(ctx context.Context, coin sdk.Coin) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if coin.Denom != params.BaseDenom {
		return types.ErrAccounting.Wrapf("total minted denom must be %s", params.BaseDenom)
	}
	if coin.Amount.IsNil() || coin.Amount.IsNegative() {
		return types.ErrAccounting.Wrap("total minted accounting cannot be negative")
	}
	bz, err := k.cdc.Marshal(&coin)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.TotalMintedAccountingKey, bz)
}

func (k Keeper) GetTotalMintedAccounting(ctx context.Context) (sdk.Coin, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.TotalMintedAccountingKey)
	if err != nil || bz == nil {
		return sdk.NewInt64Coin(types.BaseDenom, 0), err
	}
	var coin sdk.Coin
	if err := k.cdc.Unmarshal(bz, &coin); err != nil {
		return sdk.Coin{}, err
	}
	return coin, nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	gs.Params = types.NormalizeParams(gs.Params)
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	for _, record := range types.SortEmissionEpochs(gs.EpochHistory) {
		if err := k.SetEmissionEpoch(ctx, record); err != nil {
			return err
		}
	}
	return k.SetTotalMintedAccounting(ctx, gs.TotalMintedAccounting)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	history, err := k.GetAllEmissionEpochs(ctx)
	if err != nil {
		return nil, err
	}
	total, err := k.GetTotalMintedAccounting(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{Params: params, EpochHistory: history, TotalMintedAccounting: total}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

func epochKey(epoch uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.EpochPrefix[0]
	binary.BigEndian.PutUint64(key[1:], epoch)
	return key
}
