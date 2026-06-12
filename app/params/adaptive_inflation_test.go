package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestAdaptiveInflationRaisesUnderLowStakeLowFeesWithinWindowLimit(t *testing.T) {
	params := DefaultAdaptiveInflationParams()
	params.PerWindowAdjustmentLimitBps = 20
	params.GovernanceAllowsBelowRewardFloor = true
	report, err := ComputeAdaptiveInflationEpoch(baseAdaptiveInflationInput(params, AdaptiveInflationInput{
		EpochID:			31,
		BondedStakeRatioBps:		5_800,
		FeeRevenueNaet:			sdkmath.NewInt(100_000_000),
		ValidatorCount:			55,
		ValidatorRewardFloorNaet:	sdkmath.NewInt(40_000),
		NetworkActivitySamplesBps:	[]int64{2_000, 2_200, 2_100},
		TreasuryReserveHealthBps:	7_000,
		SecurityReserveHealthBps:	7_500,
	}))
	require.NoError(t, err)
	require.True(t, report.Reconciled, report.Failed)
	require.Equal(t, int64(320), report.InflationRateNextEpochBps)
	require.Equal(t, int64(20), report.ControllerState.AppliedDeltaBps)
	require.True(t, report.ControllerState.ChangeLimited)
	require.Equal(t, sdkmath.NewInt(32_000), report.MintAmountNaet)
	require.True(t, report.DeflationGuard.Active)
	require.Contains(t, report.DeflationGuard.Reasons, "security_reward_floor_pressure")
}

func TestAdaptiveInflationFallsAfterSecurityNeedDeclines(t *testing.T) {
	params := DefaultAdaptiveInflationParams()
	params.PerWindowAdjustmentLimitBps = 25
	report, err := ComputeAdaptiveInflationEpoch(baseAdaptiveInflationInput(params, AdaptiveInflationInput{
		EpochID:			32,
		BondedStakeRatioBps:		7_500,
		FeeRevenueNaet:			sdkmath.NewInt(3_000_000_000),
		ValidatorCount:			90,
		ValidatorRewardFloorNaet:	sdkmath.NewInt(5_000),
		NetworkActivitySamplesBps:	[]int64{8_900, 9_100, 9_200},
		TreasuryReserveHealthBps:	12_000,
		SecurityReserveHealthBps:	11_500,
		RecentInflationBps:		[]int64{300, 295, 290},
	}))
	require.NoError(t, err)
	require.True(t, report.Reconciled, report.Failed)
	require.Less(t, report.InflationRateNextEpochBps, int64(300))
	require.GreaterOrEqual(t, report.InflationRateNextEpochBps, params.MinInflationBps)
	require.False(t, report.DeflationGuard.Active)
}

func TestAdaptiveInflationUsesManipulationResistantActivityScore(t *testing.T) {
	params := DefaultAdaptiveInflationParams()
	params.ActivityClampDeltaBps = 500
	report, err := ComputeAdaptiveInflationEpoch(baseAdaptiveInflationInput(params, AdaptiveInflationInput{
		EpochID:			33,
		BondedStakeRatioBps:		DefaultTargetStakeBps,
		FeeRevenueNaet:			sdkmath.NewInt(DefaultFeeRevenueTargetNaet),
		ValidatorCount:			DefaultActiveValidatorTarget,
		ValidatorRewardFloorNaet:	sdkmath.NewInt(1_000),
		NetworkActivitySamplesBps:	[]int64{8_800, 8_900, 9_000, 100},
		TreasuryReserveHealthBps:	BasisPoints,
		SecurityReserveHealthBps:	BasisPoints,
	}))
	require.NoError(t, err)
	require.True(t, report.ControllerState.ActivityManipulationClamped)
	require.Equal(t, int64(8_762), report.ControllerState.ManipulationResistantActivityBps)
	require.True(t, report.Reconciled, report.Failed)
}

func TestAdaptiveInflationDetectsEpochAccountingMismatch(t *testing.T) {
	params := DefaultAdaptiveInflationParams()
	input := baseAdaptiveInflationInput(params, AdaptiveInflationInput{
		EpochID:			34,
		BondedStakeRatioBps:		DefaultTargetStakeBps,
		FeeRevenueNaet:			sdkmath.NewInt(DefaultFeeRevenueTargetNaet),
		ValidatorCount:			DefaultActiveValidatorTarget,
		ValidatorRewardFloorNaet:	sdkmath.NewInt(1_000),
		NetworkActivitySamplesBps:	[]int64{DefaultTargetLoadBps},
		TreasuryReserveHealthBps:	BasisPoints,
		SecurityReserveHealthBps:	BasisPoints,
	})
	input.EndingSupplyNaet = sdkmath.NewInt(1)

	report, err := ComputeAdaptiveInflationEpoch(input)
	require.NoError(t, err)
	require.False(t, report.Reconciled)
	require.Contains(t, report.Failed, "epoch_accounting_mismatch")
	require.Equal(t, AdaptiveInflationEventReconcile, report.Events[2].Type)
	require.False(t, report.Events[2].Reconciled)
}

func TestAdaptiveInflationStressSupplyAndSecurityBudget(t *testing.T) {
	params := DefaultAdaptiveInflationParams()
	params.GovernanceAllowsBelowRewardFloor = true
	report, err := RunAdaptiveInflationStressTest([]AdaptiveInflationStressScenario{
		{
			Name:	"low_activity_security_budget",
			Input: baseAdaptiveInflationInput(params, AdaptiveInflationInput{
				EpochID:			40,
				BondedStakeRatioBps:		5_500,
				FeeRevenueNaet:			sdkmath.NewInt(100_000_000),
				ValidatorCount:			50,
				ValidatorRewardFloorNaet:	sdkmath.NewInt(45_000),
				NetworkActivitySamplesBps:	[]int64{2_000, 2_100, 2_200},
				TreasuryReserveHealthBps:	6_000,
				SecurityReserveHealthBps:	6_500,
			}),
		},
		{
			Name:	"normal",
			Input: baseAdaptiveInflationInput(params, AdaptiveInflationInput{
				EpochID:			41,
				BondedStakeRatioBps:		DefaultTargetStakeBps,
				FeeRevenueNaet:			sdkmath.NewInt(DefaultFeeRevenueTargetNaet),
				ValidatorCount:			DefaultActiveValidatorTarget,
				ValidatorRewardFloorNaet:	sdkmath.NewInt(20_000),
				NetworkActivitySamplesBps:	[]int64{DefaultTargetLoadBps, DefaultTargetLoadBps + 100},
				TreasuryReserveHealthBps:	BasisPoints,
				SecurityReserveHealthBps:	BasisPoints,
			}),
		},
		{
			Name:	"high_activity_fee_rich",
			Input: baseAdaptiveInflationInput(params, AdaptiveInflationInput{
				EpochID:			42,
				BondedStakeRatioBps:		7_600,
				FeeRevenueNaet:			sdkmath.NewInt(3_000_000_000),
				ValidatorCount:			95,
				ValidatorRewardFloorNaet:	sdkmath.NewInt(5_000),
				NetworkActivitySamplesBps:	[]int64{9_000, 9_100, 9_200},
				TreasuryReserveHealthBps:	13_000,
				SecurityReserveHealthBps:	12_000,
			}),
		},
		{
			Name:	"adversarial_activity_spike",
			Input: baseAdaptiveInflationInput(params, AdaptiveInflationInput{
				EpochID:			43,
				BondedStakeRatioBps:		6_500,
				FeeRevenueNaet:			sdkmath.NewInt(800_000_000),
				ValidatorCount:			70,
				ValidatorRewardFloorNaet:	sdkmath.NewInt(20_000),
				NetworkActivitySamplesBps:	[]int64{7_000, 7_100, 10_000, 0},
				TreasuryReserveHealthBps:	9_000,
				SecurityReserveHealthBps:	9_500,
			}),
		},
	}, params)
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Scenarios, 4)
	require.GreaterOrEqual(t, report.MinInflationObservedBps, params.MinInflationBps)
	require.LessOrEqual(t, report.MaxInflationObservedBps, params.MaxInflationBps)
	for _, scenario := range report.Scenarios {
		require.True(t, scenario.Reconciled, scenario.Failed)
		require.NotEmpty(t, scenario.Events)
		require.Equal(t, scenario.MintAmountNaet.Sub(scenario.DeflationGuard.BurnAmountNaet), scenario.NetIssuance.NetSupplyChangeNaet)
	}
}

func baseAdaptiveInflationInput(params AdaptiveInflationParams, override AdaptiveInflationInput) AdaptiveInflationInput {
	input := AdaptiveInflationInput{
		EpochID:			1,
		AccountingPeriod:		"epoch",
		BlocksInEpoch:			100,
		CurrentSupplyNaet:		sdkmath.NewInt(1_000_000),
		CurrentInflationBps:		DefaultTargetInflationBps,
		BondedStakeRatioBps:		DefaultTargetStakeBps,
		TargetStakeRatioBps:		params.TargetStakeRatioBps,
		FeeRevenueNaet:			sdkmath.NewInt(DefaultFeeRevenueTargetNaet),
		BurnAmountNaet:			sdkmath.NewInt(5_000),
		ValidatorCount:			DefaultActiveValidatorTarget,
		ValidatorRewardFloorNaet:	sdkmath.NewInt(10_000),
		NetworkActivitySamplesBps:	[]int64{DefaultTargetLoadBps},
		TreasuryReserveHealthBps:	BasisPoints,
		SecurityReserveHealthBps:	BasisPoints,
		RecentInflationBps:		[]int64{DefaultTargetInflationBps},
		Params:				params,
	}
	if override.EpochID != 0 {
		input.EpochID = override.EpochID
	}
	if override.AccountingPeriod != "" {
		input.AccountingPeriod = override.AccountingPeriod
	}
	if override.BlocksInEpoch != 0 {
		input.BlocksInEpoch = override.BlocksInEpoch
	}
	if !override.CurrentSupplyNaet.IsNil() {
		input.CurrentSupplyNaet = override.CurrentSupplyNaet
	}
	if !override.EndingSupplyNaet.IsNil() {
		input.EndingSupplyNaet = override.EndingSupplyNaet
	}
	if override.CurrentInflationBps != 0 {
		input.CurrentInflationBps = override.CurrentInflationBps
	}
	if override.BondedStakeRatioBps != 0 {
		input.BondedStakeRatioBps = override.BondedStakeRatioBps
	}
	if override.TargetStakeRatioBps != 0 {
		input.TargetStakeRatioBps = override.TargetStakeRatioBps
	}
	if !override.FeeRevenueNaet.IsNil() {
		input.FeeRevenueNaet = override.FeeRevenueNaet
	}
	if !override.BurnAmountNaet.IsNil() {
		input.BurnAmountNaet = override.BurnAmountNaet
	}
	if override.ValidatorCount != 0 {
		input.ValidatorCount = override.ValidatorCount
	}
	if !override.ValidatorRewardFloorNaet.IsNil() {
		input.ValidatorRewardFloorNaet = override.ValidatorRewardFloorNaet
	}
	if override.NetworkActivitySamplesBps != nil {
		input.NetworkActivitySamplesBps = override.NetworkActivitySamplesBps
	}
	if override.TreasuryReserveHealthBps != 0 {
		input.TreasuryReserveHealthBps = override.TreasuryReserveHealthBps
	}
	if override.SecurityReserveHealthBps != 0 {
		input.SecurityReserveHealthBps = override.SecurityReserveHealthBps
	}
	if override.RecentInflationBps != nil {
		input.RecentInflationBps = override.RecentInflationBps
	}
	return input
}
