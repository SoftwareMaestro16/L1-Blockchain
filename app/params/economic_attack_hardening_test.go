package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestEconomicAttackPreventionModelsEveryAttackClass(t *testing.T) {
	params := DefaultEconomicAttackHardeningParams()
	report, err := EvaluateEconomicAttackPrevention(EconomicAttackPreventionInput{
		ExpectedAttackProfitNaet:	sdkmath.NewInt(1_000),
		TotalStakeNaet:			sdkmath.NewInt(1_000_000),
		ValidatorStakeNaet:		sdkmath.NewInt(400_000),
		ValidatorStakeBps:		4_000,
		CommissionChangeBps:		500,
		RewardDeviationBps:		900,
		SlashingPenaltyNaet:		sdkmath.NewInt(20_000),
		FeeSpamTxCount:			500,
		FeePerSpamTxNaet:		sdkmath.NewInt(3),
		FailedTxRateBps:		2_500,
		StateGrowthBytes:		25_000,
		StateExpansionFeeNaet:		sdkmath.NewInt(5_000),
		EvidenceSubmissions:		25,
		EvidenceDepositNaet:		sdkmath.NewInt(100),
		ReporterRewardCapNaet:		sdkmath.NewInt(500),
		DelegationInflowBps:		1_500,
		ControllerInput: EconomicCircuitBreakerInput{
			BlockLoadBps:		BasisPoints,
			FeeSpikeBps:		DefaultCircuitBreakerFeeSpikeBps + 1,
			ControllerDriftBps:	DefaultCircuitBreakerControllerDriftBps + 1,
			FailedTxRateBps:	DefaultCircuitBreakerFailedTxRateBps + 1,
			BurnToMintBps:		DeflationGuardBurnToMintBps + 1,
		},
	}, params)
	require.NoError(t, err)
	require.True(t, report.Passed)
	require.Empty(t, report.Failed)
	require.Len(t, report.Assessments, 7)
	require.True(t, report.CircuitBreaker.Active)
	require.NotZero(t, report.CircuitBreaker.CooldownBlocks)

	seen := map[string]EconomicAttackAssessment{}
	for _, assessment := range report.Assessments {
		seen[assessment.Class] = assessment
		require.False(t, assessment.CostNaet.IsNegative())
		require.NotEmpty(t, assessment.Mitigation)
		require.GreaterOrEqual(t, assessment.CostToProfitBps, int64(0))
	}
	for _, class := range requiredAttackClasses() {
		require.Contains(t, seen, class)
		require.True(t, seen[class].Triggered, class)
	}
	require.Equal(t, sdkmath.NewInt(1_500), seen[AttackClassFeeSpam].CostNaet)
	require.Equal(t, sdkmath.NewInt(150_000), seen[AttackClassDelegationCapture].CostNaet)
}

func TestEconomicAttackInvariantFailsMissingMitigationOrClass(t *testing.T) {
	report := ValidateEconomicAttackInvariants(EconomicAttackPreventionReport{
		Assessments: []EconomicAttackAssessment{
			{Class: AttackClassFeeSpam, CostNaet: sdkmath.NewInt(1), ExpectedProfit: sdkmath.NewInt(10), CostToProfitBps: 1, Triggered: true, Profitable: true},
		},
	}, DefaultEconomicAttackHardeningParams())
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, AttackClassFeeSpam+"_mitigation_missing")
	require.Contains(t, report.Failed, AttackClassStakeConcentration+"_assessment_missing")
	require.Contains(t, report.Failed, AttackClassFeeSpam+"_attack_profitable_below_cost_floor")
}

func TestValidatorCartelSimulationPricesCollusion(t *testing.T) {
	params := DefaultEconomicAttackHardeningParams()
	params.CartelPenaltyBps = 20_000
	report, err := SimulateValidatorCartel(CartelSimulationInput{
		ValidatorPowerBps:	[]int64{2_000, 1_600, 800, 600},
		ColludingIndices:	[]int{0, 1},
		RewardPoolNaet:		sdkmath.NewInt(10_000),
		ExpectedMEVNaet:	sdkmath.NewInt(500),
		Epochs:			2,
		Params:			params,
	})
	require.NoError(t, err)
	require.True(t, report.Triggered)
	require.Equal(t, int64(3_600), report.CartelVotingPowerBps)
	require.Equal(t, sdkmath.NewInt(8_200), report.ExpectedGainNaet)
	require.Equal(t, sdkmath.NewInt(16_400), report.ExpectedPenaltyNaet)
	require.False(t, report.Profitable)
	require.NotEmpty(t, report.Mitigation)
}

func TestStakeMovementMonitorFlagsAbnormalMovements(t *testing.T) {
	params := DefaultEconomicAttackHardeningParams()
	params.StakeMovementThresholdBps = 500
	report, err := MonitorStakeMovements(StakeMovementMonitorInput{
		Previous: []StakeMovementSnapshot{
			{ValidatorID: "val1", StakeBps: 1_000},
			{ValidatorID: "val2", StakeBps: 1_200},
			{ValidatorID: "val3", StakeBps: 800},
		},
		Current: []StakeMovementSnapshot{
			{ValidatorID: "val1", StakeBps: 1_700},
			{ValidatorID: "val2", StakeBps: 600},
			{ValidatorID: "val3", StakeBps: 900},
		},
		Params:	params,
	})
	require.NoError(t, err)
	require.True(t, report.Abnormal)
	require.Equal(t, int64(800), report.TotalInflowBps)
	require.Equal(t, int64(600), report.TotalOutflowBps)
	require.Len(t, report.Alerts, 2)
	require.Equal(t, "val1", report.Alerts[0].ValidatorID)
	require.Equal(t, "stake_inflow_abnormal", report.Alerts[0].Reason)
	require.Equal(t, "val2", report.Alerts[1].ValidatorID)
	require.Equal(t, "stake_outflow_abnormal", report.Alerts[1].Reason)
}

func TestGovernedCircuitBreakerActivationIsDeterministic(t *testing.T) {
	params := DefaultEconomicAttackHardeningParams()
	params.CircuitBreakerParams.MaxFeeSpikeBps = 1_000
	params.CircuitBreakerParams.MaxControllerDriftBps = 100
	params.CircuitBreakerParams.MaxFailedTxRateBps = 100
	params.CircuitBreakerParams.MinCooldownBlocks = 9

	first, err := EvaluateGovernedEconomicCircuitBreaker(EconomicCircuitBreakerInput{
		BlockLoadBps:		BasisPoints,
		FeeSpikeBps:		1_001,
		ControllerDriftBps:	101,
		FailedTxRateBps:	101,
		BurnToMintBps:		DeflationGuardBurnToMintBps + 1,
	}, params)
	require.NoError(t, err)
	second, err := EvaluateGovernedEconomicCircuitBreaker(EconomicCircuitBreakerInput{
		BlockLoadBps:		BasisPoints,
		FeeSpikeBps:		1_001,
		ControllerDriftBps:	101,
		FailedTxRateBps:	101,
		BurnToMintBps:		DeflationGuardBurnToMintBps + 1,
	}, params)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.True(t, first.Active)
	require.Equal(t, uint64(9), first.CooldownBlocks)
	require.ElementsMatch(t, []string{
		"block_load_abnormal",
		"fee_spike_abnormal",
		"controller_drift_abnormal",
		"failed_tx_rate_abnormal",
		"burn_pressure_abnormal",
	}, first.Reasons)
}
