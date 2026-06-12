package keeper

import (
	"context"
	"encoding/binary"
	"strings"

	corestore "cosmossdk.io/core/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/burn/types"
)

type Keeper struct {
	cdc		codec.BinaryCodec
	storeService	corestore.KVStoreService
	bankKeeper	types.BankKeeper
	authority	string
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, bankKeeper types.BankKeeper, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, bankKeeper: bankKeeper, authority: authority}
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

func (k Keeper) BurnUserCoins(ctx context.Context, burner sdk.AccAddress, amount sdk.Coins, epoch uint64, reason string) (types.BurnReason, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.BurnReason{}, err
	}
	if epoch == 0 {
		return types.BurnReason{}, types.ErrInvalidBurn.Wrap("epoch must be positive")
	}
	if len(reason) > int(params.MaxReasonBytes) {
		return types.BurnReason{}, types.ErrInvalidBurn.Wrap("reason exceeds max_reason_bytes")
	}
	if len(burner) == 0 || aetraaddress.IsZeroAccAddress(burner) {
		return types.BurnReason{}, types.ErrInvalidBurn.Wrap("burner must not be empty or zero")
	}
	if err := types.ValidateBurnCoins(params, amount); err != nil {
		return types.BurnReason{}, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	if err := k.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, burner, types.ModuleName, amount); err != nil {
		return types.BurnReason{}, err
	}
	if err := k.bankKeeper.BurnCoins(cacheCtx, types.ModuleName, amount); err != nil {
		return types.BurnReason{}, err
	}
	record, err := k.recordBurn(cacheCtx, amount, epoch, aetraaddress.FormatAccAddress(burner), "", reason, false)
	if err != nil {
		return types.BurnReason{}, err
	}
	write()
	return record, nil
}

func (k Keeper) BurnProtocolCoins(ctx context.Context, sourceModule string, amount sdk.Coins, epoch uint64, reason string) (types.BurnReason, error) {
	sourceModule = strings.TrimSpace(sourceModule)
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.BurnReason{}, err
	}
	if epoch == 0 {
		return types.BurnReason{}, types.ErrInvalidBurn.Wrap("epoch must be positive")
	}
	if len(reason) > int(params.MaxReasonBytes) {
		return types.BurnReason{}, types.ErrInvalidBurn.Wrap("reason exceeds max_reason_bytes")
	}
	if sourceModule == "" {
		return types.BurnReason{}, types.ErrUnauthorized.Wrap("source module must be set")
	}
	if err := types.ValidateProtocolBurn(params, sourceModule, amount); err != nil {
		return types.BurnReason{}, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	if sourceModule != types.ModuleName {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(cacheCtx, sourceModule, types.ModuleName, amount); err != nil {
			return types.BurnReason{}, err
		}
	}
	if err := k.bankKeeper.BurnCoins(cacheCtx, types.ModuleName, amount); err != nil {
		return types.BurnReason{}, err
	}
	record, err := k.recordBurn(cacheCtx, amount, epoch, "", sourceModule, reason, true)
	if err != nil {
		return types.BurnReason{}, err
	}
	write()
	return record, nil
}

func (k Keeper) recordBurn(ctx context.Context, amount sdk.Coins, epoch uint64, burner, sourceModule, reason string, protocol bool) (types.BurnReason, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.BurnReason{}, err
	}
	id, err := k.nextBurnReasonID(ctx)
	if err != nil {
		return types.BurnReason{}, err
	}
	record := types.BurnReason{
		Id:		id,
		Epoch:		epoch,
		Burner:		burner,
		SourceModule:	sourceModule,
		Amount:		amount,
		Reason:		reason,
		Height:		sdk.UnwrapSDKContext(ctx).BlockHeight(),
		Protocol:	protocol,
	}
	if err := record.Validate(params); err != nil {
		return types.BurnReason{}, types.ErrInvalidBurn.Wrap(err.Error())
	}
	for _, coin := range amount {
		entry, _, err := k.GetBurnedDenomEntry(ctx, coin.Denom)
		if err != nil {
			return types.BurnReason{}, err
		}
		entry.Denom = coin.Denom
		entry.Amount = entry.Amount.Add(coin)
		if err := k.SetBurnedDenomEntry(ctx, entry); err != nil {
			return types.BurnReason{}, err
		}
	}
	epochEntry, _, err := k.GetBurnedEpochEntry(ctx, epoch)
	if err != nil {
		return types.BurnReason{}, err
	}
	epochEntry.Epoch = epoch
	epochEntry.Amount = epochEntry.Amount.Add(amount...)
	if err := k.SetBurnedEpochEntry(ctx, epochEntry); err != nil {
		return types.BurnReason{}, err
	}
	if err := k.SetBurnReason(ctx, record); err != nil {
		return types.BurnReason{}, err
	}
	if err := k.setNextBurnReasonID(ctx, id+1); err != nil {
		return types.BurnReason{}, err
	}
	return record, nil
}

func (k Keeper) SetBurnedDenomEntry(ctx context.Context, entry types.BurnedByDenomEntry) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := entry.Validate(params); err != nil {
		return types.ErrInvalidBurn.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&entry)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(burnedDenomKey(entry.Denom), bz)
}

func (k Keeper) GetBurnedDenomEntry(ctx context.Context, denom string) (types.BurnedByDenomEntry, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(burnedDenomKey(denom))
	if err != nil || bz == nil {
		return types.BurnedByDenomEntry{Denom: denom, Amount: sdk.NewCoins()}, false, err
	}
	var entry types.BurnedByDenomEntry
	if err := k.cdc.Unmarshal(bz, &entry); err != nil {
		return types.BurnedByDenomEntry{}, false, err
	}
	return entry, true, nil
}

func (k Keeper) GetAllBurnedByDenom(ctx context.Context) ([]types.BurnedByDenomEntry, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.BurnedDenomPrefix, storetypes.PrefixEndBytes(types.BurnedDenomPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.BurnedByDenomEntry{}
	for ; iter.Valid(); iter.Next() {
		var entry types.BurnedByDenomEntry
		if err := k.cdc.Unmarshal(iter.Value(), &entry); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return types.SortBurnedByDenom(out), nil
}

func (k Keeper) SetBurnedEpochEntry(ctx context.Context, entry types.BurnedByEpochEntry) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := entry.Validate(params); err != nil {
		return types.ErrInvalidBurn.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&entry)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(burnedEpochKey(entry.Epoch), bz)
}

func (k Keeper) GetBurnedEpochEntry(ctx context.Context, epoch uint64) (types.BurnedByEpochEntry, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(burnedEpochKey(epoch))
	if err != nil || bz == nil {
		return types.BurnedByEpochEntry{Epoch: epoch, Amount: sdk.NewCoins()}, false, err
	}
	var entry types.BurnedByEpochEntry
	if err := k.cdc.Unmarshal(bz, &entry); err != nil {
		return types.BurnedByEpochEntry{}, false, err
	}
	return entry, true, nil
}

func (k Keeper) GetAllBurnedByEpoch(ctx context.Context) ([]types.BurnedByEpochEntry, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.BurnedEpochPrefix, storetypes.PrefixEndBytes(types.BurnedEpochPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.BurnedByEpochEntry{}
	for ; iter.Valid(); iter.Next() {
		var entry types.BurnedByEpochEntry
		if err := k.cdc.Unmarshal(iter.Value(), &entry); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return types.SortBurnedByEpoch(out), nil
}

func (k Keeper) SetBurnReason(ctx context.Context, record types.BurnReason) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := record.Validate(params); err != nil {
		return types.ErrInvalidBurn.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(burnReasonKey(record.Id), bz)
}

func (k Keeper) GetAllBurnReasons(ctx context.Context) ([]types.BurnReason, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.BurnReasonPrefix, storetypes.PrefixEndBytes(types.BurnReasonPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []types.BurnReason{}
	for ; iter.Valid(); iter.Next() {
		var record types.BurnReason
		if err := k.cdc.Unmarshal(iter.Value(), &record); err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return types.SortBurnReasons(out), nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	gs.Params = types.NormalizeParams(gs.Params)
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	for _, entry := range types.SortBurnedByDenom(gs.BurnedByDenom) {
		if err := k.SetBurnedDenomEntry(ctx, entry); err != nil {
			return err
		}
	}
	for _, entry := range types.SortBurnedByEpoch(gs.BurnedByEpoch) {
		if err := k.SetBurnedEpochEntry(ctx, entry); err != nil {
			return err
		}
	}
	nextID := uint64(1)
	for _, record := range types.SortBurnReasons(gs.BurnReasons) {
		if err := k.SetBurnReason(ctx, record); err != nil {
			return err
		}
		if record.Id >= nextID {
			nextID = record.Id + 1
		}
	}
	return k.setNextBurnReasonID(ctx, nextID)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	byDenom, err := k.GetAllBurnedByDenom(ctx)
	if err != nil {
		return nil, err
	}
	byEpoch, err := k.GetAllBurnedByEpoch(ctx)
	if err != nil {
		return nil, err
	}
	reasons, err := k.GetAllBurnReasons(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{Params: params, BurnedByDenom: byDenom, BurnedByEpoch: byEpoch, BurnReasons: reasons}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

func (k Keeper) nextBurnReasonID(ctx context.Context) (uint64, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.NextBurnReasonIDKey)
	if err != nil || bz == nil {
		return 1, err
	}
	return binary.BigEndian.Uint64(bz), nil
}

func (k Keeper) setNextBurnReasonID(ctx context.Context, id uint64) error {
	if id == 0 {
		return types.ErrInvalidBurn.Wrap("next burn reason id must be positive")
	}
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return k.storeService.OpenKVStore(ctx).Set(types.NextBurnReasonIDKey, bz)
}

func burnedDenomKey(denom string) []byte {
	key := append([]byte{}, types.BurnedDenomPrefix...)
	key = append(key, []byte(denom)...)
	return key
}

func burnedEpochKey(epoch uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.BurnedEpochPrefix[0]
	binary.BigEndian.PutUint64(key[1:], epoch)
	return key
}

func burnReasonKey(id uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.BurnReasonPrefix[0]
	binary.BigEndian.PutUint64(key[1:], id)
	return key
}
