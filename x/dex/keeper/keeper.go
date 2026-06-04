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
	authority    string
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, bankKeeper types.BankKeeper, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, bankKeeper: bankKeeper, authority: authority}
}

func (k Keeper) Authority() string { return k.authority }

func poolKey(id uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.PoolPrefix[0]
	binary.BigEndian.PutUint64(key[1:], id)
	return key
}

func pairKey(denom0, denom1 string) []byte {
	key := make([]byte, 1, 1+len(denom0)+1+len(denom1))
	key[0] = types.PoolPairPrefix[0]
	key = append(key, denom0...)
	key = append(key, 0x00)
	key = append(key, denom1...)
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
	bz, err := k.cdc.Marshal(&pool)
	if err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	if err := store.Set(poolKey(pool.Id), bz); err != nil {
		return err
	}
	idBz := make([]byte, 8)
	binary.BigEndian.PutUint64(idBz, pool.Id)
	return store.Set(pairKey(pool.Denom0, pool.Denom1), idBz)
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

func (k Keeper) GetPoolIDByPair(ctx context.Context, denom0, denom1 string) (uint64, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(pairKey(denom0, denom1))
	if err != nil || bz == nil {
		return 0, false, err
	}
	if len(bz) != 8 {
		return 0, false, types.ErrInvalidPool.Wrap("corrupted pair index")
	}
	return binary.BigEndian.Uint64(bz), true, nil
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

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	if gs.NextPoolId == 0 {
		gs.NextPoolId = types.DefaultNextPoolID
	}
	if err := k.SetNextPoolID(ctx, gs.NextPoolId); err != nil {
		return err
	}
	for _, pool := range gs.Pools {
		if err := k.SetPool(ctx, pool); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	next, err := k.GetNextPoolID(ctx)
	if err != nil {
		return nil, err
	}
	pools, err := k.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{NextPoolId: next, Pools: pools, Params: params}
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

func calcSwapOut(reserveIn, reserveOut, amountIn sdkmath.Int, feeBps uint32) sdkmath.Int {
	amountInAfterFee := amountIn.MulRaw(types.BpsDenominator - int64(feeBps)).QuoRaw(types.BpsDenominator)
	return reserveOut.Mul(amountInAfterFee).Quo(reserveIn.Add(amountInAfterFee))
}
