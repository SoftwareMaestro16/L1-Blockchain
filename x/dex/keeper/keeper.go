package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"sort"

	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

func parseInt(value string) sdkmath.Int {
	out, ok := sdkmath.NewIntFromString(value)
	if !ok {
		return sdkmath.ZeroInt()
	}
	return out
}

func intString(value sdkmath.Int) string { return value.String() }

func parsePositiveInt(field, value string) (sdkmath.Int, error) {
	out, ok := sdkmath.NewIntFromString(value)
	if !ok || !out.IsPositive() {
		return sdkmath.Int{}, types.ErrInvalidPool.Wrapf("%s must be a positive integer", field)
	}
	return out, nil
}

func parseNonNegativeInt(field, value string) (sdkmath.Int, error) {
	out, ok := sdkmath.NewIntFromString(value)
	if !ok || out.IsNegative() {
		return sdkmath.Int{}, types.ErrInvalidPool.Wrapf("%s must be a non-negative integer", field)
	}
	return out, nil
}

func validatePoolState(pool types.Pool) (reserve0, reserve1, totalShares sdkmath.Int, err error) {
	if pool.Id == 0 {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("pool id must be positive")
	}
	if err := sdk.ValidateDenom(pool.Denom0); err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid denom0: %v", err)
	}
	if err := sdk.ValidateDenom(pool.Denom1); err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid denom1: %v", err)
	}
	if pool.Denom0 >= pool.Denom1 {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("pool denoms must be unique and canonical")
	}
	if pool.LpDenom != lpDenom(pool.Id) {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("invalid lp denom")
	}
	reserve0, err = parsePositiveInt("reserve0", pool.Reserve0)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	reserve1, err = parsePositiveInt("reserve1", pool.Reserve1)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	totalShares, err = parsePositiveInt("total_shares", pool.TotalShares)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	return reserve0, reserve1, totalShares, nil
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
	return k.storeService.OpenKVStore(ctx).Set(poolKey(pool.Id), k.cdc.MustMarshal(&pool))
}

func (k Keeper) GetPool(ctx context.Context, id uint64) (types.Pool, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(poolKey(id))
	if err != nil || bz == nil {
		return types.Pool{}, false, err
	}
	var pool types.Pool
	k.cdc.MustUnmarshal(bz, &pool)
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
		k.cdc.MustUnmarshal(iter.Value(), &pool)
		pools = append(pools, pool)
	}
	return pools, nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) {
	if gs.NextPoolId == 0 {
		gs.NextPoolId = types.DefaultNextPoolID
	}
	if err := k.SetNextPoolID(ctx, gs.NextPoolId); err != nil {
		panic(err)
	}
	for _, pool := range gs.Pools {
		if err := k.SetPool(ctx, pool); err != nil {
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

func canonicalPair(a, b sdk.Coin) (sdk.Coin, sdk.Coin, error) {
	if !a.IsValid() || !b.IsValid() || !a.IsPositive() || !b.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, types.ErrInvalidLiquidity.Wrap("tokens must be positive")
	}
	if a.Denom == b.Denom {
		return sdk.Coin{}, sdk.Coin{}, types.ErrInvalidPool.Wrap("pool requires two different denoms")
	}
	coins := []sdk.Coin{a, b}
	sort.Slice(coins, func(i, j int) bool { return coins[i].Denom < coins[j].Denom })
	return coins[0], coins[1], nil
}

func coinsForPool(pool types.Pool, a, b sdk.Coin) (sdk.Coin, sdk.Coin, error) {
	if a.Denom == pool.Denom0 && b.Denom == pool.Denom1 {
		return a, b, nil
	}
	if a.Denom == pool.Denom1 && b.Denom == pool.Denom0 {
		return b, a, nil
	}
	return sdk.Coin{}, sdk.Coin{}, types.ErrInvalidLiquidity.Wrap("liquidity denoms do not match pool")
}

func lpDenom(poolID uint64) string {
	return fmt.Sprintf("%s/%d", types.LPDenomPrefix, poolID)
}

func minInt(a, b sdkmath.Int) sdkmath.Int {
	if a.LT(b) {
		return a
	}
	return b
}

func calcSwapOut(reserveIn, reserveOut, amountIn sdkmath.Int) sdkmath.Int {
	amountInAfterFee := amountIn.MulRaw(types.BpsDenominator - types.PoolFeeBps).QuoRaw(types.BpsDenominator)
	return reserveOut.Mul(amountInAfterFee).Quo(reserveIn.Add(amountInAfterFee))
}
