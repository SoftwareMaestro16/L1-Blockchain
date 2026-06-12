package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestActivityInflationControllerRaisesInflationWithExplainableInputs(t *testing.T) {
	params := DefaultActivityInflationControllerParams()
	out, err := ActivityInflationControllerWithParams(ActivityInflationControllerInput{
		CurrentInflationBps:		DefaultTargetInflationBps,
		BondedStakeRatioBps:		5_000,
		ValidatorOperatingCostIndexBps:	8_000,
		FeeRevenueNaet:			sdkmath.NewInt(100_000_000),
		ActiveValidatorCount:		50,
		SlashingRiskEvents:		2,
		NetworkActivityScoreBps:	3_000,
		TreasuryReserveHealthBps:	7_000,
		RecentInflationBps:		[]int64{290, 295, DefaultTargetInflationBps},
	}, params)
	require.NoError(t, err)
	require.Equal(t, DefaultTargetInflationBps+params.PerWindowChangeLimitBps, out.InflationBps)
	require.Equal(t, params.MaxInflationBps, out.RawTargetInflationBps)
	require.True(t, out.ChangeLimited)
	require.False(t, out.EmergencyFrozen)
	require.Len(t, out.Components, 7)
	require.Equal(t, int64(5_000), out.QueryableInputs.BondedStakeRatioBps)
	require.Equal(t, sdkmath.NewInt(100_000_000), out.QueryableInputs.FeeRevenueNaet)

	componentNames := map[string]bool{}
	for _, component := range out.Components {
		componentNames[component.Name] = true
	}
	for _, name := range []string{
		"bonded_stake_ratio",
		"validator_operating_cost_index",
		"fee_revenue",
		"active_validator_count",
		"slashing_risk_events",
		"network_activity_score",
		"treasury_reserve_health",
	} {
		require.True(t, componentNames[name])
	}
}

func TestActivityInflationControllerEmergencyFreezeHoldsCurrentInflation(t *testing.T) {
	params := DefaultActivityInflationControllerParams()
	params.EmergencyFreeze = true

	out, err := ActivityInflationControllerWithParams(ActivityInflationControllerInput{
		CurrentInflationBps:		DefaultTargetInflationBps,
		BondedStakeRatioBps:		1_000,
		ValidatorOperatingCostIndexBps:	1_000,
		FeeRevenueNaet:			sdkmath.ZeroInt(),
		ActiveValidatorCount:		5,
		SlashingRiskEvents:		10,
		NetworkActivityScoreBps:	0,
		TreasuryReserveHealthBps:	1_000,
		RecentInflationBps:		[]int64{MaxInflationBps, MaxInflationBps},
	}, params)
	require.NoError(t, err)
	require.Equal(t, DefaultTargetInflationBps, out.InflationBps)
	require.Zero(t, out.AppliedDeltaBps)
	require.True(t, out.EmergencyFrozen)
	require.False(t, out.ChangeLimited)
	require.Len(t, out.Components, 7)
}

func TestActivityInflationControllerHighActivityReducesIssuanceWithinLimit(t *testing.T) {
	params := DefaultActivityInflationControllerParams()
	out, err := ActivityInflationControllerWithParams(ActivityInflationControllerInput{
		CurrentInflationBps:		DefaultTargetInflationBps,
		BondedStakeRatioBps:		7_500,
		ValidatorOperatingCostIndexBps:	12_000,
		FeeRevenueNaet:			sdkmath.NewInt(DefaultFeeRevenueTargetNaet * 3),
		ActiveValidatorCount:		90,
		NetworkActivityScoreBps:	9_500,
		TreasuryReserveHealthBps:	12_000,
		RecentInflationBps:		[]int64{310, 305, DefaultTargetInflationBps},
	}, params)
	require.NoError(t, err)
	require.Equal(t, -params.PerWindowChangeLimitBps, out.AppliedDeltaBps)
	require.Equal(t, DefaultTargetInflationBps-params.PerWindowChangeLimitBps, out.InflationBps)
	require.GreaterOrEqual(t, out.InflationBps, params.MinInflationBps)
	require.LessOrEqual(t, out.InflationBps, params.MaxInflationBps)
	require.True(t, out.ChangeLimited)
	require.Less(t, out.RawTargetInflationBps, DefaultTargetInflationBps)
}

func TestActivityInflationControllerClampsToConfiguredBounds(t *testing.T) {
	params := DefaultActivityInflationControllerParams()
	out, err := ActivityInflationControllerWithParams(ActivityInflationControllerInput{
		CurrentInflationBps:		MaxInflationBps,
		BondedStakeRatioBps:		0,
		ValidatorOperatingCostIndexBps:	0,
		FeeRevenueNaet:			sdkmath.ZeroInt(),
		ActiveValidatorCount:		1,
		SlashingRiskEvents:		100,
		NetworkActivityScoreBps:	0,
		TreasuryReserveHealthBps:	0,
		RecentInflationBps:		[]int64{MaxInflationBps, MaxInflationBps},
	}, params)
	require.NoError(t, err)
	require.Equal(t, MaxInflationBps, out.InflationBps)
	require.LessOrEqual(t, out.InflationBps, params.MaxInflationBps)
}

func TestNetIssuanceReportPerEpochAndAccountingPeriod(t *testing.T) {
	report, err := ReportNetIssuance(NetIssuanceInput{
		EpochID:			10,
		AccountingPeriod:		"daily",
		Blocks:				100,
		GrossMintedNaet:		sdkmath.NewInt(1_000),
		BurnedNaet:			sdkmath.NewInt(250),
		FeeRevenueNaet:			sdkmath.NewInt(700),
		ValidatorSecuritySpendNaet:	sdkmath.NewInt(500),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(10), report.EpochID)
	require.Equal(t, "daily", report.AccountingPeriod)
	require.Equal(t, sdkmath.NewInt(1_000), report.GrossMintedNaet)
	require.Equal(t, sdkmath.NewInt(250), report.BurnedNaet)
	require.Equal(t, sdkmath.NewInt(750), report.NetSupplyChangeNaet)
	require.Equal(t, sdkmath.NewInt(5), report.SecuritySpendPerBlockNaet)
}

func TestSimulateActivityInflationCoversLowNormalHighAndAdversarialActivity(t *testing.T) {
	params := DefaultActivityInflationControllerParams()
	steps := []InflationSimulationStep{
		{
			Scenario:	InflationScenarioLowActivity,
			Controller: ActivityInflationControllerInput{
				CurrentInflationBps:		DefaultTargetInflationBps,
				BondedStakeRatioBps:		5_500,
				ValidatorOperatingCostIndexBps:	8_500,
				FeeRevenueNaet:			sdkmath.NewInt(100_000_000),
				ActiveValidatorCount:		50,
				NetworkActivityScoreBps:	2_000,
				TreasuryReserveHealthBps:	8_000,
				RecentInflationBps:		[]int64{290, 295, 300},
			},
			NetIssuance:	simulationNetIssuance(1, 10_000, 1_000),
		},
		{
			Scenario:	InflationScenarioNormalActivity,
			Controller: ActivityInflationControllerInput{
				BondedStakeRatioBps:		DefaultTargetStakeBps,
				ValidatorOperatingCostIndexBps:	BasisPoints,
				FeeRevenueNaet:			sdkmath.NewInt(DefaultFeeRevenueTargetNaet),
				ActiveValidatorCount:		DefaultActiveValidatorTarget,
				NetworkActivityScoreBps:	DefaultTargetLoadBps,
				TreasuryReserveHealthBps:	BasisPoints,
				RecentInflationBps:		[]int64{300, 310, 320},
			},
			NetIssuance:	simulationNetIssuance(2, 10_500, 2_000),
		},
		{
			Scenario:	InflationScenarioHighActivity,
			Controller: ActivityInflationControllerInput{
				BondedStakeRatioBps:		7_700,
				ValidatorOperatingCostIndexBps:	12_000,
				FeeRevenueNaet:			sdkmath.NewInt(DefaultFeeRevenueTargetNaet * 4),
				ActiveValidatorCount:		90,
				NetworkActivityScoreBps:	9_500,
				TreasuryReserveHealthBps:	12_000,
				RecentInflationBps:		[]int64{320, 315, 310},
			},
			NetIssuance:	simulationNetIssuance(3, 9_000, 5_000),
		},
		{
			Scenario:	InflationScenarioAdversarialActivity,
			Controller: ActivityInflationControllerInput{
				BondedStakeRatioBps:		4_800,
				ValidatorOperatingCostIndexBps:	7_000,
				FeeRevenueNaet:			sdkmath.NewInt(50_000_000),
				ActiveValidatorCount:		40,
				SlashingRiskEvents:		8,
				NetworkActivityScoreBps:	9_900,
				TreasuryReserveHealthBps:	6_500,
				RecentInflationBps:		[]int64{300, 310, 320},
			},
			NetIssuance:	simulationNetIssuance(4, 12_000, 4_000),
		},
	}

	report, err := SimulateActivityInflation(params, steps)
	require.NoError(t, err)
	require.True(t, report.Passed)
	require.Empty(t, report.Risks)
	require.Equal(t, 4, report.ScenarioCount)
	require.Len(t, report.ControllerOutputs, 4)
	require.Len(t, report.NetIssuanceReports, 4)
	require.GreaterOrEqual(t, report.MinObservedInflationBps, params.MinInflationBps)
	require.LessOrEqual(t, report.MaxObservedInflationBps, params.MaxInflationBps)
	for _, out := range report.ControllerOutputs {
		require.LessOrEqual(t, absInt64(out.AppliedDeltaBps), params.PerWindowChangeLimitBps)
	}
}

func simulationNetIssuance(epochID uint64, grossMinted, burned int64) NetIssuanceInput {
	return NetIssuanceInput{
		EpochID:			epochID,
		AccountingPeriod:		"epoch",
		Blocks:				100,
		GrossMintedNaet:		sdkmath.NewInt(grossMinted),
		BurnedNaet:			sdkmath.NewInt(burned),
		FeeRevenueNaet:			sdkmath.NewInt(grossMinted + burned),
		ValidatorSecuritySpendNaet:	sdkmath.NewInt(grossMinted / 2),
	}
}
