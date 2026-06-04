package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func FuzzCalcSwapOutPreservesConstantProduct(f *testing.F) {
	seeds := [][3]int64{
		{1_000, 1_000, 10},
		{1_000_000, 500_000, 123},
		{10, 10_000, 9},
		{9_223_372_036, 4_000_000_000, 1_000_000},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2])
	}

	f.Fuzz(func(t *testing.T, reserveInRaw, reserveOutRaw, amountInRaw int64) {
		if reserveInRaw <= 0 || reserveOutRaw <= 1 || amountInRaw <= 0 {
			t.Skip()
		}

		reserveIn := sdkmath.NewInt(reserveInRaw)
		reserveOut := sdkmath.NewInt(reserveOutRaw)
		amountIn := sdkmath.NewInt(amountInRaw)
		amountInAfterFee := calcAmountInAfterFee(amountIn)
		if !amountInAfterFee.IsPositive() {
			t.Skip()
		}

		out := calcSwapOut(reserveIn, reserveOut, amountIn)
		if !out.IsPositive() {
			t.Skip()
		}
		if !out.LT(reserveOut) {
			t.Fatalf("invalid out amount: reserve_out=%s amount_in=%s out=%s", reserveOut, amountIn, out)
		}

		beforeK := reserveIn.Mul(reserveOut)
		afterK := reserveIn.Add(amountIn).Mul(reserveOut.Sub(out))
		if afterK.LT(beforeK) {
			t.Fatalf("constant product decreased: before=%s after=%s", beforeK, afterK)
		}
	})
}

func FuzzCalcLiquiditySharesIsBounded(f *testing.F) {
	seeds := [][5]int64{
		{1_000, 1_000, 1_000, 100, 100},
		{1_000, 2_000, 1_000, 50, 100},
		{10_000_000, 5_000_000, 1_000_000, 1_000, 500},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4])
	}

	f.Fuzz(func(t *testing.T, reserve0Raw, reserve1Raw, totalSharesRaw, amount0Raw, amount1Raw int64) {
		if reserve0Raw <= 0 || reserve1Raw <= 0 || totalSharesRaw <= 0 || amount0Raw <= 0 || amount1Raw <= 0 {
			t.Skip()
		}

		reserve0 := sdkmath.NewInt(reserve0Raw)
		reserve1 := sdkmath.NewInt(reserve1Raw)
		totalShares := sdkmath.NewInt(totalSharesRaw)
		amount0 := sdkmath.NewInt(amount0Raw)
		amount1 := sdkmath.NewInt(amount1Raw)

		shares, err := calcLiquidityShares(reserve0, reserve1, totalShares, amount0, amount1)
		if err != nil {
			return
		}
		if !shares.IsPositive() {
			t.Fatalf("positive proportional deposit minted non-positive shares")
		}
		if shares.GT(amount0.Mul(totalShares).Quo(reserve0)) || shares.GT(amount1.Mul(totalShares).Quo(reserve1)) {
			t.Fatalf("minted shares exceed reserve-ratio bound")
		}
	})
}
