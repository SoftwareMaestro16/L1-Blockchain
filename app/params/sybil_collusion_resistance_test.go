package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestStakeSplittingDoesNotBypassConcentrationPolicy(t *testing.T) {
	params := DefaultSybilCollusionResistanceParams()
	params.GroupConcentrationSoftCapBps = 2_500
	params.GroupMaxRewardDampeningBps = 3_000
	params.DeterministicCorrelationThresholdBps = 9_000

	report, err := SimulateStakeSplitting(StakeSplitSimulationInput{
		RewardPoolNaet:	sdkmath.NewInt(100_000),
		Original: ValidatorSybilRecord{
			ValidatorID:			"val-original",
			EconomicGroupID:		"operator-a",
			VotingPowerBps:			4_000,
			PerformanceBps:			BasisPoints,
			SelfDelegationRatioBps:		1_000,
			SelfDelegationNaet:		sdkmath.NewInt(10_000),
			CorrelationScoreBps:		9_500,
			CorrelationProofDeterministic:	true,
		},
		Split: []ValidatorSybilRecord{
			{
				ValidatorID:			"val-a1",
				EconomicGroupID:		"operator-a",
				VotingPowerBps:			2_000,
				PerformanceBps:			BasisPoints,
				SelfDelegationRatioBps:		1_000,
				SelfDelegationNaet:		sdkmath.NewInt(5_000),
				CorrelationScoreBps:		9_500,
				CorrelationProofDeterministic:	true,
			},
			{
				ValidatorID:			"val-a2",
				EconomicGroupID:		"operator-a",
				VotingPowerBps:			2_000,
				PerformanceBps:			BasisPoints,
				SelfDelegationRatioBps:		1_000,
				SelfDelegationNaet:		sdkmath.NewInt(5_000),
				CorrelationScoreBps:		9_500,
				CorrelationProofDeterministic:	true,
			},
		},
		Params:	params,
	})
	require.NoError(t, err)
	require.Equal(t, int64(4_000), report.OriginalGroupPowerBps)
	require.Equal(t, int64(4_000), report.SplitGroupPowerBps)
	require.Equal(t, report.OriginalDampeningBps, report.SplitDampeningBps)
	require.Equal(t, report.OriginalRewardNaet, report.SplitRewardNaet)
	require.True(t, report.BypassPrevented)
	require.True(t, report.ConsensusDampeningApplied)
	require.False(t, report.AdvisoryOnly)
}

func TestAdvisoryCorrelationDoesNotAffectConsensusWithoutDeterministicProof(t *testing.T) {
	params := DefaultSybilCollusionResistanceParams()
	report, err := EvaluateCorrelationTelemetry("operator-advisory", []ValidatorSybilRecord{
		{
			ValidatorID:		"val-b1",
			EconomicGroupID:	"operator-advisory",
			VotingPowerBps:		2_000,
			PerformanceBps:		BasisPoints,
			SelfDelegationRatioBps:	1_000,
			SelfDelegationNaet:	sdkmath.NewInt(5_000),
			CorrelationScoreBps:	7_000,
			CommissionChangeBps:	400,
		},
		{
			ValidatorID:		"val-b2",
			EconomicGroupID:	"operator-advisory",
			VotingPowerBps:		1_500,
			PerformanceBps:		BasisPoints,
			SelfDelegationRatioBps:	1_000,
			SelfDelegationNaet:	sdkmath.NewInt(5_000),
			CorrelationScoreBps:	7_000,
			CommissionChangeBps:	400,
		},
	}, params)
	require.NoError(t, err)
	require.True(t, report.AdvisoryOnly)
	require.False(t, report.ConsensusAffecting)
	require.Zero(t, report.RewardDampeningBps)
	require.Contains(t, report.Signals, CorrelationSignalCommission)
}

func TestLegitimateNewValidatorRetainsBootstrapPathOncePerWindow(t *testing.T) {
	params := DefaultSybilCollusionResistanceParams()
	params.BootstrapMaxStakeBps = 500
	params.BootstrapBonusBps = 400
	validator := ValidatorSybilRecord{
		ValidatorID:			"val-new",
		VotingPowerBps:			300,
		PerformanceBps:			BasisPoints,
		SelfDelegationRatioBps:		1_000,
		SelfDelegationNaet:		sdkmath.NewInt(2_000),
		BootstrapWindowID:		12,
		BootstrapRewardAlreadyUsed:	false,
	}
	report, err := EvaluateValidatorActiveSetEligibility(validator, params)
	require.NoError(t, err)
	require.True(t, report.Eligible)
	require.True(t, report.BootstrapEligible)
	require.Equal(t, int64(400), report.BootstrapBonusBps)
	require.Empty(t, report.RejectReasons)

	validator.BootstrapRewardAlreadyUsed = true
	used, err := EvaluateValidatorActiveSetEligibility(validator, params)
	require.NoError(t, err)
	require.True(t, used.Eligible)
	require.False(t, used.BootstrapEligible)
	require.Zero(t, used.BootstrapBonusBps)
}

func TestActiveSetEligibilityRequiresSelfDelegationAndPerformance(t *testing.T) {
	report, err := EvaluateValidatorActiveSetEligibility(ValidatorSybilRecord{
		ValidatorID:		"val-weak",
		VotingPowerBps:		300,
		PerformanceBps:		9_000,
		SelfDelegationRatioBps:	10,
		SelfDelegationNaet:	sdkmath.NewInt(10),
		BootstrapWindowID:	1,
	}, DefaultSybilCollusionResistanceParams())
	require.NoError(t, err)
	require.False(t, report.Eligible)
	require.False(t, report.BootstrapEligible)
	require.ElementsMatch(t, []string{
		"self_delegation_below_minimum",
		"self_delegation_ratio_below_minimum",
		"performance_below_minimum",
	}, report.RejectReasons)
}

func TestValidatorCollusionSimulationDampensTopNAndReportsCorrelation(t *testing.T) {
	params := DefaultSybilCollusionResistanceParams()
	params.TopNConcentrationLimitBps = 5_000
	params.GroupMaxRewardDampeningBps = 3_000
	params.CorrelatedDowntimeThresholdBps = 5_000
	report, err := SimulateValidatorCollusion(CollusionScenarioInput{
		TopN:		3,
		RewardPoolNaet:	sdkmath.NewInt(100_000),
		Params:		params,
		Validators: []ValidatorSybilRecord{
			collusionValidator("val1", "group-a", 2_500, 400, true, true),
			collusionValidator("val2", "group-a", 2_000, 400, true, true),
			collusionValidator("val3", "group-b", 1_000, 0, false, false),
			collusionValidator("val4", "group-c", 500, 0, false, false),
		},
	})
	require.NoError(t, err)
	require.True(t, report.ConcentrationExceeded)
	require.Equal(t, int64(5_500), report.TopNVotingPowerBps)
	require.Greater(t, report.ConcentrationDampeningBps, int64(0))
	require.True(t, report.RewardAfterDampeningNaet.LT(report.RewardBeforeDampeningNaet))
	require.Contains(t, report.CommissionAlerts, CorrelationSignalCommission)
	require.Contains(t, report.DowntimeAlerts, CorrelationSignalDowntime)
	require.Len(t, report.CorrelationReports, 1)
	require.True(t, report.CorrelationReports[0].ConsensusAffecting)
	require.NotEmpty(t, report.GovernanceSummary)
}

func TestEvidenceRoutingCapsReporterRewardsAndRejectsDuplicates(t *testing.T) {
	params := DefaultSybilCollusionResistanceParams()
	params.EvidenceRoutingRewardCapNaet = sdkmath.NewInt(500)
	accepted, err := RouteConsensusFaultEvidence(EvidenceRoutingInput{
		ValidatorID:			"val1",
		EvidenceType:			"equivocation",
		Accepted:			true,
		RequestedReporterRewardNaet:	sdkmath.NewInt(900),
		Params:				params,
	})
	require.NoError(t, err)
	require.True(t, accepted.Accepted)
	require.Equal(t, EvidenceRouteEquivocation, accepted.Route)
	require.Equal(t, sdkmath.NewInt(500), accepted.ReporterRewardNaet)
	require.True(t, accepted.Auditable)

	duplicate, err := RouteConsensusFaultEvidence(EvidenceRoutingInput{
		ValidatorID:			"val1",
		EvidenceType:			"downtime",
		Duplicate:			true,
		RequestedReporterRewardNaet:	sdkmath.NewInt(100),
		Params:				params,
	})
	require.NoError(t, err)
	require.False(t, duplicate.Accepted)
	require.Equal(t, EvidenceRouteDuplicate, duplicate.Route)
	require.True(t, duplicate.ReporterRewardNaet.IsZero())
	require.Equal(t, "duplicate_evidence", duplicate.RejectedReason)
}

func collusionValidator(id, group string, power int64, commissionChange int64, downtime bool, deterministic bool) ValidatorSybilRecord {
	return ValidatorSybilRecord{
		ValidatorID:			id,
		EconomicGroupID:		group,
		VotingPowerBps:			power,
		PerformanceBps:			BasisPoints,
		SelfDelegationRatioBps:		1_000,
		SelfDelegationNaet:		sdkmath.NewInt(5_000),
		CommissionChangeBps:		commissionChange,
		DowntimeObserved:		downtime,
		CorrelationScoreBps:		9_500,
		CorrelationProofDeterministic:	deterministic,
	}
}
