package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/fees/types"
)

// testFormulaParams returns a formula param set with predictable fixed values for
// golden tests. All values in naet.
func testFormulaParams() types.FeeFormulaParams {
	return types.FeeFormulaParams{
		TargetTransferFeeNaet:		"10000000",
		BaseFeePerGasNaet:		"1",
		ByteFeeNaet:			"1",
		MessageFeeNaet:			"1000",
		MaxCongestionSurchargeNaet:	"2000000",
		LowReputationPremiumCapNaet:	"500000",
		HighReputationDiscountCapNaet:	"500000",
		StorageRentSideEffectsNaet:	"0",
	}
}

// testBaseParams returns Params with simple, predictable values.
func testBaseParams() types.Params {
	p := types.DefaultParams()
	p.MinFeeAmount = "1"
	p.BaseFeeAmount = "100"
	p.MaxFeeAmount = "1000000000000000000"
	p.TargetBlockUtilizationBps = 5_000
	p.CongestionThresholdBps = 8_000
	return p
}

// TestGoldenFeeFixtureLowLoad verifies the complete formula at low block utilization
// (below congestion threshold → zero congestion surcharge).
func TestGoldenFeeFixtureLowLoad(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	gasUsed := uint64(50_000)
	txBytes := uint64(200)
	msgCount := uint64(1)
	utilizationBps := uint32(0)

	expected := sdkmath.NewInt(51_300)

	got, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		types.ReputationNeutralScore, false,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)
	require.Equal(t, expected, got, "low load golden fee mismatch")
}

// TestGoldenFeeFixtureMediumLoad verifies the formula at medium utilization
// (above target but below congestion threshold → still zero congestion surcharge).
func TestGoldenFeeFixtureMediumLoad(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	gasUsed := uint64(50_000)
	txBytes := uint64(200)
	msgCount := uint64(1)
	utilizationBps := uint32(6_000)

	expected := sdkmath.NewInt(51_300)

	got, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		types.ReputationNeutralScore, false,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)
	require.Equal(t, expected, got, "medium load golden fee mismatch")
}

// TestGoldenFeeFixtureHighLoad verifies the formula at high block utilization
// (above congestion threshold → non-zero congestion surcharge).
func TestGoldenFeeFixtureHighLoad(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	gasUsed := uint64(50_000)
	txBytes := uint64(200)
	msgCount := uint64(1)
	utilizationBps := uint32(9_000)

	expected := sdkmath.NewInt(1_051_300)

	got, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		types.ReputationNeutralScore, false,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)
	require.Equal(t, expected, got, "high load golden fee mismatch")
}

// TestLowReputationPaysMoreThanNeutralWithinCap verifies that a low-reputation sender
// pays more than neutral, and the premium never exceeds the cap.
func TestLowReputationPaysMoreThanNeutralWithinCap(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	gasUsed := uint64(50_000)
	txBytes := uint64(200)
	msgCount := uint64(1)
	utilizationBps := uint32(0)

	neutralFee, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		types.ReputationNeutralScore, false,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)

	lowScore := uint32(1_000)
	lowFee, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		lowScore, true,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)

	require.True(t, lowFee.GT(neutralFee),
		"low-reputation fee %s must exceed neutral fee %s", lowFee, neutralFee)

	capNaet := sdkmath.NewInt(500_000)
	premium := lowFee.Sub(neutralFee)
	require.True(t, premium.LTE(capNaet),
		"reputation premium %s must not exceed cap %s", premium, capNaet)
}

// TestHighReputationNeverPaysBelowMinTxFee verifies that even with maximum discount,
// the fee never drops below min_tx_fee_naet (Requirement 1.5).
func TestHighReputationNeverPaysBelowMinTxFee(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	gasUsed := uint64(1)
	txBytes := uint64(1)
	msgCount := uint64(1)
	utilizationBps := uint32(0)

	maxScore := uint32(10_000)

	fee, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		maxScore, true,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)

	minFee, err := bp.MinFeeInt()
	require.NoError(t, err)

	require.True(t, fee.GTE(minFee),
		"high-reputation fee %s must be >= min_tx_fee_naet %s", fee, minFee)
}

// TestZeroReputationDoesNotBlockTx verifies that score=0 still returns a valid
// non-negative fee (never blocks the transaction, Requirement 1.4, 7.2).
func TestZeroReputationDoesNotBlockTx(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	fee, err := types.ComputeFullTransferFee(
		bp, fp,
		50_000, 200, 1,
		0,
		0, true,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)
	require.True(t, fee.IsPositive(), "score=0 must still produce a positive fee (tx not blocked)")
}

// TestHighReputationDiscountBoundedByCap verifies that the discount for the highest
// possible reputation score stays within the discount cap.
func TestHighReputationDiscountBoundedByCap(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	gasUsed := uint64(50_000)
	txBytes := uint64(200)
	msgCount := uint64(1)
	utilizationBps := uint32(0)

	neutralFee, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		types.ReputationNeutralScore, false,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)

	highScore := uint32(10_000)
	highFee, err := types.ComputeFullTransferFee(
		bp, fp,
		gasUsed, txBytes, msgCount,
		utilizationBps,
		highScore, true,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)

	require.True(t, highFee.LTE(neutralFee),
		"high-reputation fee %s must be <= neutral fee %s", highFee, neutralFee)

	discountCap := sdkmath.NewInt(500_000)
	discount := neutralFee.Sub(highFee)
	require.True(t, discount.LTE(discountCap),
		"reputation discount %s must not exceed cap %s", discount, discountCap)
}

// TestWrongDenomRejected verifies that a fee paid in wrong denom is rejected.
func TestWrongDenomRejected(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(1)

	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}

	wrongDenomTx := feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000))}
	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, wrongDenomTx, false)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidFee)
	require.Contains(t, err.Error(), "not accepted")
}

// TestZeroFeeRejected verifies that a zero-fee transaction is rejected before any
// state mutation (Requirement 1.7, no fee bypass).
func TestZeroFeeRejected(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(1)

	stateMutated := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		stateMutated = true
		return ctx, nil
	}

	zeroFeeTx := feeTx{fees: sdk.Coins{}}
	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, zeroFeeTx, false)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidFee)
	require.False(t, stateMutated, "state must not be mutated when fee=0")
}

// TestTargetTransferFeeIsGovernanceParam verifies that target_transfer_fee_naet is
// stored as a KV-backed governance parameter, not a compile-time constant.
func TestTargetTransferFeeIsGovernanceParam(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	fp, err := app.FeesKeeper.GetFeeFormulaParams(ctx)
	require.NoError(t, err)

	target, err := fp.TargetTransferFeeInt()
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(types.TargetTransferFeeNaet), target,
		"target_transfer_fee_naet must default to 10_000_000 naet (Requirement 1.2)")

	fp.TargetTransferFeeNaet = "20000000"
	require.NoError(t, app.FeesKeeper.SetFeeFormulaParams(ctx, fp))

	updated, err := app.FeesKeeper.GetFeeFormulaParams(ctx)
	require.NoError(t, err)
	updatedTarget, err := updated.TargetTransferFeeInt()
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(20_000_000), updatedTarget,
		"target_transfer_fee_naet must be updatable via governance")
}

// TestMinTxFeeIsGovernanceParam verifies that min_tx_fee_naet is governance-controlled.
func TestMinTxFeeIsGovernanceParam(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	params := types.DefaultParams()
	params.MinFeeAmount = "5000"
	params.BaseFeeAmount = "5000"
	params.MaxFeeAmount = "10000"
	require.NoError(t, app.FeesKeeper.SetParams(ctx, params))

	stored, err := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, "5000", stored.MinFeeAmount,
		"min_tx_fee_naet must be updatable via governance")
}

// TestFeeFormulaParamsDefaultsAreValid verifies that default formula params pass validation.
func TestFeeFormulaParamsDefaultsAreValid(t *testing.T) {
	fp := types.DefaultFeeFormulaParams()
	require.NoError(t, fp.Validate())

	target, err := fp.TargetTransferFeeInt()
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(types.TargetTransferFeeNaet), target)
}

// TestFeeFormulaParamsRejectNegativeValues ensures negative fee components are rejected.
func TestFeeFormulaParamsRejectNegativeValues(t *testing.T) {
	tests := []struct {
		name	string
		mutate	func(*types.FeeFormulaParams)
	}{
		{"negative target fee", func(p *types.FeeFormulaParams) { p.TargetTransferFeeNaet = "-1" }},
		{"negative gas fee", func(p *types.FeeFormulaParams) { p.BaseFeePerGasNaet = "-1" }},
		{"negative byte fee", func(p *types.FeeFormulaParams) { p.ByteFeeNaet = "-1" }},
		{"negative message fee", func(p *types.FeeFormulaParams) { p.MessageFeeNaet = "-1" }},
		{"negative congestion cap", func(p *types.FeeFormulaParams) { p.MaxCongestionSurchargeNaet = "-1" }},
		{"negative premium cap", func(p *types.FeeFormulaParams) { p.LowReputationPremiumCapNaet = "-1" }},
		{"negative discount cap", func(p *types.FeeFormulaParams) { p.HighReputationDiscountCapNaet = "-1" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fp := types.DefaultFeeFormulaParams()
			tc.mutate(&fp)
			require.Error(t, fp.Validate(), "negative formula param must be rejected")
		})
	}
}

// TestCongestionSurchargeIsZeroBelowThreshold ensures surcharge=0 when utilization
// is at or below the congestion threshold (Requirement 1.3).
func TestCongestionSurchargeIsZeroBelowThreshold(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	for _, bps := range []uint32{0, 1000, 5000, 7999, 8000} {
		feeAtThreshold, err := types.ComputeFullTransferFee(
			bp, fp,
			1, 1, 1,
			bps,
			types.ReputationNeutralScore, false,
			sdkmath.ZeroInt(),
		)
		require.NoError(t, err)

		feeNoSurcharge, err := types.ComputeFullTransferFee(
			bp, fp,
			1, 1, 1,
			0,
			types.ReputationNeutralScore, false,
			sdkmath.ZeroInt(),
		)
		require.NoError(t, err)

		require.Equal(t, feeNoSurcharge, feeAtThreshold,
			"congestion surcharge must be zero at utilization bps=%d (threshold=8000)", bps)
	}
}

// TestStorageRentSideEffectsIncludedInFee verifies that storage rent side effects
// are added to the required fee budget (Requirement 6.6).
func TestStorageRentSideEffectsIncludedInFee(t *testing.T) {
	bp := testBaseParams()
	fp := testFormulaParams()

	feeWithoutRent, err := types.ComputeFullTransferFee(
		bp, fp,
		1, 1, 1, 0,
		types.ReputationNeutralScore, false,
		sdkmath.ZeroInt(),
	)
	require.NoError(t, err)

	rentAmount := sdkmath.NewInt(10_000)
	feeWithRent, err := types.ComputeFullTransferFee(
		bp, fp,
		1, 1, 1, 0,
		types.ReputationNeutralScore, false,
		rentAmount,
	)
	require.NoError(t, err)

	diff := feeWithRent.Sub(feeWithoutRent)
	require.Equal(t, rentAmount, diff,
		"storage rent side effects must be included 1:1 in the required fee budget")
}
