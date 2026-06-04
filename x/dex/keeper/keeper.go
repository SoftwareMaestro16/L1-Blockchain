package keeper

import (
	"context"
	"encoding/binary"

	corestore "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/sovereign-l1/l1/x/dex/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	bankKeeper   types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, bankKeeper types.BankKeeper) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, bankKeeper: bankKeeper}
}

func poolKey(id uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.PoolPrefix[0]
	binary.BigEndian.PutUint64(key[1:], id)
	return key
}

func pairKey(denom0, denom1 string) []byte {
	key := make([]byte, 0, len(types.PairPrefix)+len(denom0)+1+len(denom1))
	key = append(key, types.PairPrefix...)
	key = append(key, []byte(denom0)...)
	key = append(key, 0)
	return append(key, []byte(denom1)...)
}

func (k Keeper) GetNextPoolID(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.NextPoolIDKey)
	if err != nil || bz == nil {
		return types.DefaultNextPoolID, err
	}
	return binary.BigEndian.Uint64(bz), nil
}

func (k Keeper) SetNextPoolID(ctx context.Context, id uint64) error {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return k.storeService.OpenKVStore(ctx).Set(types.NextPoolIDKey, bz)
}

func (k Keeper) SetPool(ctx context.Context, pool types.Pool) error {
	bz, err := k.cdc.Marshal(&pool)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(poolKey(pool.Id), bz)
}

func (k Keeper) GetPool(ctx context.Context, id uint64) (types.Pool, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(poolKey(id))
	if err != nil || bz == nil {
		return types.Pool{}, false, err
	}
	var pool types.Pool
	if err := k.cdc.Unmarshal(bz, &pool); err != nil {
		return types.Pool{}, false, err
	}
	return pool, true, nil
}

func (k Keeper) GetAllPools(ctx context.Context) ([]types.Pool, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.PoolPrefix, storetypes.PrefixEndBytes(types.PoolPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var pools []types.Pool
	for ; iter.Valid(); iter.Next() {
		var pool types.Pool
		if err := k.cdc.Unmarshal(iter.Value(), &pool); err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}
	return pools, nil
}

func (k Keeper) GetPoolIDByPair(ctx context.Context, denom0, denom1 string) (uint64, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(pairKey(denom0, denom1))
	if err != nil || bz == nil {
		return 0, false, err
	}
	if len(bz) != 8 {
		return 0, false, types.ErrInvariant.Wrap("invalid pair index value")
	}
	return binary.BigEndian.Uint64(bz), true, nil
}

func (k Keeper) SetPoolPairIndex(ctx context.Context, pool types.Pool) error {
	if _, _, _, err := validatePoolState(pool); err != nil {
		return err
	}
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, pool.Id)
	return k.storeService.OpenKVStore(ctx).Set(pairKey(pool.Denom0, pool.Denom1), bz)
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) {
	if gs.NextPoolId == 0 {
		gs.NextPoolId = types.DefaultNextPoolID
	}
	if err := gs.Validate(); err != nil {
		panic(err)
	}
	if err := k.SetNextPoolID(ctx, gs.NextPoolId); err != nil {
		panic(err)
	}
	for _, pool := range gs.Pools {
		if err := k.SetPool(ctx, pool); err != nil {
			panic(err)
		}
		if err := k.SetPoolPairIndex(ctx, pool); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	next, err := k.GetNextPoolID(ctx)
	if err != nil {
		panic(err)
	}
	pools, err := k.GetAllPools(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{NextPoolId: next, Pools: pools}
}
