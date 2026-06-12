package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatorReputationSeparatesConsensusSafeAndAdvisoryComponents(t *testing.T) {
	report, err := EvaluateValidatorReputation(ValidatorReputationInput{
		ValidatorID:			"val-a",
		UptimeHistoryBps:		[]int64{9_950, 9_900, 9_980},
		MissedBlockRateHistoryBps:	[]int64{20, 30, 10},
		SlashEvents:			1,
		SlashSeverityBps:		200,
		CommissionChangeHistoryBps:	[]int64{50, 75},
		SelfDelegationChangeHistoryBps:	[]int64{100, -50},
		MetadataComplete:		true,
		DelegationInflowBps:		300,
		HistoricalRewardPerformanceBps:	[]int64{10_100, 9_900},
		VotingPowerBps:			2_500,
	}, ValidatorReputationParams{})
	require.NoError(t, err)
	require.Equal(t, ValidatorReputationScoringVersionV1, report.ScoringVersion)
	require.True(t, report.ConsensusSafeForSelection)
	require.Len(t, report.ConsensusComponents, 2)
	require.Len(t, report.AdvisoryComponents, 6)
	for _, component := range report.ConsensusComponents {
		require.Equal(t, ReputationComponentConsensusSafe, component.Source)
	}
	for _, component := range report.AdvisoryComponents {
		require.Equal(t, ReputationComponentAdvisoryOnly, component.Source)
	}
	require.False(t, report.ConcentrationWarning)
	require.True(t, report.CaptureRiskWarning)
	require.Equal(t, report.RiskScoreBps, report.DelegatorMetadata.RiskScoreBps)
	require.NotEmpty(t, report.ScoreExplanation)
}

func TestValidatorReputationBlocksUnsafeAdvisoryConsensusUse(t *testing.T) {
	report, err := EvaluateValidatorReputation(ValidatorReputationInput{
		ValidatorID:			"val-consensus",
		UptimeHistoryBps:		[]int64{9_980},
		MissedBlockRateHistoryBps:	[]int64{0},
		CommissionChangeHistoryBps:	[]int64{100},
		SelfDelegationChangeHistoryBps:	[]int64{0},
		MetadataComplete:		true,
		HistoricalRewardPerformanceBps:	[]int64{10_000},
		UseInConsensus:			true,
		AdvisoryInputsDeterministic:	false,
	}, ValidatorReputationParams{})
	require.NoError(t, err)
	require.False(t, report.ConsensusSafeForSelection)
	require.Contains(t, report.Failed, "advisory_reputation_inputs_not_consensus_safe")
	require.True(t, report.DelegatorMetadata.AdvisoryOnly)
	require.Contains(t, report.ScoreExplanation, "consensus_use_requires_consensus_safe_components_only")
}

func TestValidatorReputationAllowsConsensusUseWhenAdvisoryInputsAreGovernedDeterministic(t *testing.T) {
	params := DefaultValidatorReputationParams()
	params.AllowAdvisoryInputsInConsensus = true

	report, err := EvaluateValidatorReputation(ValidatorReputationInput{
		ValidatorID:			"val-safe",
		UptimeHistoryBps:		[]int64{9_980},
		MissedBlockRateHistoryBps:	[]int64{0},
		CommissionChangeHistoryBps:	[]int64{0},
		SelfDelegationChangeHistoryBps:	[]int64{0},
		MetadataComplete:		true,
		HistoricalRewardPerformanceBps:	[]int64{10_000},
		UseInConsensus:			true,
		AdvisoryInputsDeterministic:	true,
	}, params)
	require.NoError(t, err)
	require.True(t, report.ConsensusSafeForSelection)
	require.Empty(t, report.Failed)
}

func TestValidatorReputationStableUnderMissingAdvisoryData(t *testing.T) {
	base := ValidatorReputationInput{
		ValidatorID:			"val-missing",
		UptimeHistoryBps:		[]int64{9_900, 9_950},
		MissedBlockRateHistoryBps:	[]int64{10, 20},
		MetadataComplete:		true,
		VotingPowerBps:			2_000,
	}
	withAdvisory := base
	withAdvisory.CommissionChangeHistoryBps = []int64{20}
	withAdvisory.SelfDelegationChangeHistoryBps = []int64{0}
	withAdvisory.HistoricalRewardPerformanceBps = []int64{10_000}

	missing, err := EvaluateValidatorReputation(base, ValidatorReputationParams{})
	require.NoError(t, err)
	full, err := EvaluateValidatorReputation(withAdvisory, ValidatorReputationParams{})
	require.NoError(t, err)

	require.Equal(t, full.ConsensusSafeScoreBps, missing.ConsensusSafeScoreBps)
	require.Equal(t, full.ReliabilityScoreBps, missing.ReliabilityScoreBps)
	require.Less(t, missing.AdvisoryScoreBps, full.AdvisoryScoreBps)
	require.Contains(t, missing.ScoreExplanation, "commission_stability:advisory_only:missing_data_penalized")
	require.Contains(t, missing.ScoreExplanation, "reward_performance:advisory_only:missing_data_penalized")
}

func TestValidatorReputationCaptureAndConcentrationWarnings(t *testing.T) {
	report, err := EvaluateValidatorReputation(ValidatorReputationInput{
		ValidatorID:			"val-whale",
		UptimeHistoryBps:		[]int64{9_950},
		MissedBlockRateHistoryBps:	[]int64{0},
		CommissionChangeHistoryBps:	[]int64{800},
		SelfDelegationChangeHistoryBps:	[]int64{-1_500},
		MetadataChangeCount:		2,
		MetadataComplete:		false,
		DelegationInflowBps:		2_500,
		HistoricalRewardPerformanceBps:	[]int64{9_500},
		VotingPowerBps:			MaxTopValidatorConcentrationBps + 1,
	}, ValidatorReputationParams{})
	require.NoError(t, err)
	require.True(t, report.ConcentrationWarning)
	require.True(t, report.CaptureRiskWarning)
	require.True(t, report.RiskScoreBps > 0)
	require.Equal(t, report.ConcentrationWarning, report.DelegatorMetadata.ConcentrationWarning)
	require.Equal(t, report.CaptureRiskWarning, report.DelegatorMetadata.CaptureRiskWarning)
}
