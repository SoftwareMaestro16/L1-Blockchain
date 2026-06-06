package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultEconomicPriorityMatrixIsCompleteAndRanked(t *testing.T) {
	report, err := BuildEconomicPriorityMatrixReport(nil, EconomicPriorityParams{})
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Ranked, 30)
	require.Equal(t, 16, report.HighUrgencyCount)
	require.Equal(t, int64(5), report.AverageComplexity)
	require.Equal(t, "Add validator concentration metrics", report.HighestPriority.Improvement)
	require.Equal(t, PriorityWaveCritical, report.HighestPriority.ExecutionWave)
	require.Contains(t, report.GovernanceSummary, "items=30")
	require.NotEmpty(t, report.Waves)

	for i := 1; i < len(report.Ranked); i++ {
		require.GreaterOrEqual(t, report.Ranked[i-1].PriorityScore, report.Ranked[i].PriorityScore)
		require.NotEmpty(t, report.Ranked[i].GovernanceRationale)
	}
}

func TestEconomicPriorityMatrixSortsTiesDeterministically(t *testing.T) {
	params := DefaultEconomicPriorityParams()
	params.RequireKnownMatrixComplete = false
	report, err := BuildEconomicPriorityMatrixReport([]EconomicPriorityItem{
		{Improvement: "b-option", SecurityImpact: 8, DecentralizationImpact: 5, ImplementationComplexity: 5, Urgency: PriorityUrgencyHigh},
		{Improvement: "a-option", SecurityImpact: 8, DecentralizationImpact: 5, ImplementationComplexity: 5, Urgency: PriorityUrgencyHigh},
	}, params)
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Equal(t, "a-option", report.Ranked[0].Improvement)
	require.Equal(t, "b-option", report.Ranked[1].Improvement)
}

func TestEconomicPriorityMatrixRejectsInvalidRows(t *testing.T) {
	params := DefaultEconomicPriorityParams()
	params.RequireKnownMatrixComplete = false
	report, err := BuildEconomicPriorityMatrixReport([]EconomicPriorityItem{
		{Improvement: "bad-score", SecurityImpact: 11, DecentralizationImpact: 1, ImplementationComplexity: 1, Urgency: PriorityUrgencyHigh},
		{Improvement: "bad-score", SecurityImpact: 1, DecentralizationImpact: 1, ImplementationComplexity: 0, Urgency: "Soon"},
	}, params)
	require.NoError(t, err)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "bad-score:security_impact_out_of_range")
	require.Contains(t, report.Failed, "bad-score:implementation_complexity_out_of_range")
	require.Contains(t, report.Failed, "bad-score:urgency_invalid")
	require.Contains(t, report.Failed, "duplicate_improvement:bad-score")
}

func TestEconomicPriorityMatrixRequiresKnownMatrixWhenConfigured(t *testing.T) {
	report, err := BuildEconomicPriorityMatrixReport([]EconomicPriorityItem{
		{Improvement: "only-one", SecurityImpact: 5, DecentralizationImpact: 5, ImplementationComplexity: 5, Urgency: PriorityUrgencyLow},
	}, DefaultEconomicPriorityParams())
	require.NoError(t, err)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "prioritization_matrix_size_mismatch")
	require.Contains(t, report.Failed, "missing_default_improvement:Add economic invariant tests")
}

func TestEconomicPriorityMatrixGovernanceWeightsCanReprioritize(t *testing.T) {
	params := DefaultEconomicPriorityParams()
	params.SecurityWeight = 300
	params.DecentralizationWeight = 600
	report, err := BuildEconomicPriorityMatrixReport(nil, params)
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Equal(t, "Add validator concentration metrics", report.HighestPriority.Improvement)
	require.Greater(t, report.HighestPriority.DecentralizationImpact, int64(8))

	waveCounts := map[string]int{}
	for _, wave := range report.Waves {
		waveCounts[wave.Wave] = wave.Count
		require.LessOrEqual(t, wave.ScoreMin, wave.ScoreMax)
		require.NotEmpty(t, wave.Items)
	}
	require.Greater(t, waveCounts[PriorityWaveCritical], 0)
}

func TestEconomicPriorityParamsRejectBadThresholdOrder(t *testing.T) {
	params := DefaultEconomicPriorityParams()
	params.HighPriorityScore = params.CriticalPriorityScore + 1
	_, err := BuildEconomicPriorityMatrixReport(nil, params)
	require.ErrorContains(t, err, "priority score thresholds must be descending")
}
