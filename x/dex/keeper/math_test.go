package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestCalcSwapOutAppliesFee(t *testing.T) {
	out := calcSwapOut(sdkmath.NewInt(1000), sdkmath.NewInt(1000), sdkmath.NewInt(100))
	if !out.Equal(sdkmath.NewInt(90)) {
		t.Fatalf("unexpected output amount: %s", out)
	}
}
