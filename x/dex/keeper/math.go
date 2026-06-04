package keeper

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/dex/types"
)

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

func calcAmountInAfterFee(amountIn sdkmath.Int) sdkmath.Int {
	return amountIn.MulRaw(types.BpsDenominator - types.PoolFeeBps).QuoRaw(types.BpsDenominator)
}

func calcSwapOut(reserveIn, reserveOut, amountIn sdkmath.Int) sdkmath.Int {
	amountInAfterFee := calcAmountInAfterFee(amountIn)
	return reserveOut.Mul(amountInAfterFee).Quo(reserveIn.Add(amountInAfterFee))
}

func calcLiquidityShares(reserve0, reserve1, totalShares, amount0, amount1 sdkmath.Int) (sdkmath.Int, error) {
	if !reserve0.IsPositive() || !reserve1.IsPositive() || !totalShares.IsPositive() || !amount0.IsPositive() || !amount1.IsPositive() {
		return sdkmath.Int{}, types.ErrInvalidLiquidity.Wrap("liquidity inputs must be positive")
	}
	if !amount0.Mul(reserve1).Equal(amount1.Mul(reserve0)) {
		return sdkmath.Int{}, types.ErrInvalidLiquidity.Wrap("liquidity must match pool ratio")
	}

	shares0 := amount0.Mul(totalShares).Quo(reserve0)
	shares1 := amount1.Mul(totalShares).Quo(reserve1)
	shares := minInt(shares0, shares1)
	if !shares.IsPositive() {
		return sdkmath.Int{}, types.ErrInvalidLiquidity.Wrap("minted shares round to zero")
	}
	return shares, nil
}

func assertConstantProductNotDecreased(before0, before1, after0, after1 sdkmath.Int) error {
	if !after0.IsPositive() || !after1.IsPositive() {
		return types.ErrInvariant.Wrap("pool reserves must remain positive")
	}
	if after0.Mul(after1).LT(before0.Mul(before1)) {
		return types.ErrInvariant.Wrap("constant product decreased")
	}
	return nil
}
