package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraAcceptanceMatrixCoversSection34(t *testing.T) {
	report := BuildAetraAcceptanceMatrixReport(nil)

	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, 41, report.Required)
	require.Equal(t, report.Required, report.Passed)
	require.NoError(t, ValidateAetraAcceptanceMatrix(nil))

	byCategory := map[string][]string{}
	for _, category := range report.Categories {
		byCategory[category.Category] = category.Scenarios
	}
	require.Contains(t, byCategory[AetraAcceptanceCategoryBaseNode], AetraAcceptanceBaseNodeStateSyncSnapshotRestore)
	require.Contains(t, byCategory[AetraAcceptanceCategoryStaking], AetraAcceptanceStakingValidatorCommissionUpdate)
	require.Contains(t, byCategory[AetraAcceptanceCategoryAntiCentralization], AetraAcceptanceAntiCentralizationCommissionFloor)
	require.Contains(t, byCategory[AetraAcceptanceCategorySlashing], AetraAcceptanceSlashingDelegatorAccounting)
	require.Contains(t, byCategory[AetraAcceptanceCategoryEconomics], AetraAcceptanceEconomicsSupplyInvariant)
	require.Contains(t, byCategory[AetraAcceptanceCategoryAVM], AetraAcceptanceAVMGasExhaustionContained)
	require.Contains(t, byCategory[AetraAcceptanceCategoryGovernance], AetraAcceptanceGovernanceDelayedCriticalActivation)
	require.Contains(t, byCategory[AetraAcceptanceCategoryObservability], AetraAcceptanceObservabilityEventsIndexable)
}

func TestAetraAcceptanceMatrixRejectsMissingCategory(t *testing.T) {
	evidence := DefaultAetraAcceptanceMatrixEvidence()
	evidence = evidence[:len(evidence)-1]

	report := BuildAetraAcceptanceMatrixReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraAcceptanceCategoryObservability+":missing_category")
	require.Error(t, ValidateAetraAcceptanceMatrix(evidence))
}

func TestAetraAcceptanceMatrixRejectsMissingDuplicateAndUnexpectedScenarios(t *testing.T) {
	evidence := DefaultAetraAcceptanceMatrixEvidence()
	evidence[0].Scenarios = removeAcceptanceScenario(evidence[0].Scenarios,
		AetraAcceptanceBaseNodeBootSingle,
		AetraAcceptanceBaseNodeExportImport,
	)
	evidence[0].Scenarios = append(evidence[0].Scenarios, AetraAcceptanceBaseNodeRestart, "manual_node_smoke")
	evidence[1].Category = "staking-copy"

	report := BuildAetraAcceptanceMatrixReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraAcceptanceCategoryBaseNode+"."+AetraAcceptanceBaseNodeBootSingle+":missing")
	require.Contains(t, report.Failed, AetraAcceptanceCategoryBaseNode+"."+AetraAcceptanceBaseNodeExportImport+":missing")
	require.Contains(t, report.Failed, AetraAcceptanceCategoryBaseNode+"."+AetraAcceptanceBaseNodeRestart+":duplicate")
	require.Contains(t, report.Failed, AetraAcceptanceCategoryBaseNode+".manual_node_smoke:unexpected")
	require.Contains(t, report.Failed, "staking-copy:unknown_category")
	require.Contains(t, report.Failed, AetraAcceptanceCategoryStaking+":missing_category")
}

func TestAetraAcceptanceMatrixRejectsDuplicateCategory(t *testing.T) {
	evidence := DefaultAetraAcceptanceMatrixEvidence()
	evidence[1].Category = evidence[0].Category
	evidence[1].Scenarios = RequiredAetraAcceptanceBaseNodeScenarios()

	report := BuildAetraAcceptanceMatrixReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraAcceptanceCategoryBaseNode+":duplicate_category")
	require.Contains(t, report.Failed, AetraAcceptanceCategoryStaking+":missing_category")
}

func removeAcceptanceScenario(items []string, targets ...string) []string {
	targetSet := map[string]bool{}
	for _, target := range targets {
		targetSet[target] = true
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !targetSet[item] {
			out = append(out, item)
		}
	}
	return out
}
