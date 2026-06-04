package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestCalcSwapOutAppliesFee(t *testing.T) {
	out := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(100), 30)
	if !out.Equal(sdkmath.NewInt(90)) {
		t.Fatalf("unexpected output amount: %s", out)
	}
}

func TestCalcSwapOutRespondsToGovernedFee(t *testing.T) {
	zeroFee := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(500), 0)
	maxFee := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(500), 1000)
	if !zeroFee.GT(maxFee) {
		t.Fatalf("expected higher fee to reduce output: zero_fee=%s max_fee=%s", zeroFee, maxFee)
	}
}
