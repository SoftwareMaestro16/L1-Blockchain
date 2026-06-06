package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultRequiredEconomicTestCoverageIsComplete(t *testing.T) {
	report := BuildRequiredEconomicTestCoverageReport(nil, nil)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.InvariantCases, 9)
	require.Len(t, report.SimulationCases, 14)
	require.Equal(t, 9, report.RequiredInvariants)
	require.Equal(t, 14, report.RequiredSimulations)
	require.Equal(t, 9, report.CoveredInvariants)
	require.Equal(t, 14, report.CoveredSimulations)
	require.Equal(t, BasisPoints, report.InvariantCoverageBps)
	require.Equal(t, BasisPoints, report.SimulationCoverageBps)
	require.Contains(t, report.GovernanceSummary, "required_invariants=9/9")
	require.Contains(t, report.GovernanceSummary, "required_simulations=14/14")

	for _, item := range report.InvariantCases {
		require.Equal(t, EconomicTestCoverageKindInvariant, item.Kind)
		require.True(t, item.Required)
		require.True(t, item.Deterministic)
		require.True(t, item.CIEnabled)
		require.NotEmpty(t, item.Evidence)
	}
	for _, item := range report.SimulationCases {
		require.Equal(t, EconomicTestCoverageKindSimulation, item.Kind)
		require.True(t, item.Required)
		require.True(t, item.Deterministic)
		require.True(t, item.CIEnabled)
		require.NotEmpty(t, item.Evidence)
	}
}

func TestRequiredEconomicTestCoverageRejectsMissingInvariantEvidence(t *testing.T) {
	invariants := DefaultRequiredEconomicInvariantCoverageCases()
	for i := range invariants {
		if invariants[i].ID == EconomicInvariantFeeBucketsSum {
			invariants[i].Evidence = nil
		}
	}

	report := BuildRequiredEconomicTestCoverageReport(invariants, nil)
	require.False(t, report.Passed)
	require.Equal(t, int64(8_888), report.InvariantCoverageBps)
	require.Contains(t, report.Failed, EconomicInvariantFeeBucketsSum+":evidence_missing")
}

func TestRequiredEconomicTestCoverageRejectsMissingAndDuplicateSimulationCases(t *testing.T) {
	simulations := DefaultRequiredEconomicSimulationCoverageCases()
	simulations = append(simulations[:1], simulations[2:]...)
	simulations = append(simulations, simulations[0])

	report := BuildRequiredEconomicTestCoverageReport(nil, simulations)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicSimulationNormalTargetStake+":missing_required_coverage")
	require.Contains(t, report.Failed, EconomicSimulationLowActivityLowFees+":duplicate_coverage_case")
	require.Less(t, report.SimulationCoverageBps, BasisPoints)
}

func TestRequiredEconomicTestCoverageRequiresDeterministicCIEnabledCases(t *testing.T) {
	invariants := DefaultRequiredEconomicInvariantCoverageCases()
	invariants[0].Deterministic = false
	invariants[1].CIEnabled = false

	simulations := DefaultRequiredEconomicSimulationCoverageCases()
	simulations[0].Kind = EconomicTestCoverageKindInvariant

	report := BuildRequiredEconomicTestCoverageReport(invariants, simulations)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicInvariantMintedSupplyReconciles+":not_deterministic")
	require.Contains(t, report.Failed, EconomicInvariantBurnRemovedFromSpendable+":not_ci_enabled")
	require.Contains(t, report.Failed, EconomicSimulationLowActivityLowFees+":wrong_coverage_kind")
	require.Less(t, report.InvariantCoverageBps, BasisPoints)
	require.Less(t, report.SimulationCoverageBps, BasisPoints)
}

func TestRequiredEconomicTestCoverageRejectsBlankRequiredFields(t *testing.T) {
	invariants := DefaultRequiredEconomicInvariantCoverageCases()
	invariants[0].Description = " "
	invariants[0].Evidence = []string{" "}
	invariants = append(invariants, EconomicTestCoverageCase{Kind: EconomicTestCoverageKindInvariant, Required: true})

	report := BuildRequiredEconomicTestCoverageReport(invariants, nil)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicInvariantMintedSupplyReconciles+":description_missing")
	require.Contains(t, report.Failed, EconomicInvariantMintedSupplyReconciles+":evidence_0_blank")
	require.Contains(t, report.Failed, EconomicTestCoverageKindInvariant+":case_id_required")
}
