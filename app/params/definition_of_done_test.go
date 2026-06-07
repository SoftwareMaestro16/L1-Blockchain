package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraDefinitionOfDoneCoversSection33BaseTask(t *testing.T) {
	evidence := DefaultAetraDefinitionOfDoneEvidence("api-surface", false)

	report := BuildAetraDefinitionOfDoneReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, "api-surface", report.TaskName)
	require.False(t, report.CriticalChange)
	require.Equal(t, 11, report.Required)
	require.Equal(t, report.Required, report.Passed)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementCodeImplemented)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementParamsValidated)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementGenesisImportExport)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementQuerySurface)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementOperationalEvents)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementUnitTests)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementIntegrationTests)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementE2ELocalnetUserFlow)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementOperatorUserDocs)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementFailureModesDocumented)
	require.Contains(t, evidence.Requirements, AetraDoDRequirementSecurityReviewed)
	require.Empty(t, evidence.CriticalRequirements)
	require.NoError(t, ValidateAetraDefinitionOfDone(evidence))
}

func TestDefaultAetraDefinitionOfDoneCoversConsensusEconomicsStakingTask(t *testing.T) {
	evidence := DefaultAetraDefinitionOfDoneEvidence("staking-policy", true)

	report := BuildAetraDefinitionOfDoneReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.True(t, report.CriticalChange)
	require.Equal(t, 16, report.Required)
	require.Equal(t, report.Required, report.Passed)
	require.Contains(t, evidence.CriticalRequirements, AetraDoDCriticalRequirementAdversarialTests)
	require.Contains(t, evidence.CriticalRequirements, AetraDoDCriticalRequirementInvariantTests)
	require.Contains(t, evidence.CriticalRequirements, AetraDoDCriticalRequirementExportImportTest)
	require.Contains(t, evidence.CriticalRequirements, AetraDoDCriticalRequirementDeterministicRestart)
	require.Contains(t, evidence.CriticalRequirements, AetraDoDCriticalRequirementMigrationIfState)
	require.NoError(t, ValidateAetraDefinitionOfDone(evidence))
}

func TestAetraDefinitionOfDoneRejectsMissingBaseRequirements(t *testing.T) {
	evidence := DefaultAetraDefinitionOfDoneEvidence("", false)
	evidence.Requirements = removeDoDItem(evidence.Requirements,
		AetraDoDRequirementCodeImplemented,
		AetraDoDRequirementIntegrationTests,
		AetraDoDRequirementSecurityReviewed,
	)
	evidence.Requirements = append(evidence.Requirements, AetraDoDRequirementUnitTests, "manual_checklist_only")
	evidence.CriticalRequirements = append(evidence.CriticalRequirements, AetraDoDCriticalRequirementAdversarialTests)

	report := BuildAetraDefinitionOfDoneReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "task_name_required")
	require.Contains(t, report.Failed, "requirements."+AetraDoDRequirementCodeImplemented+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraDoDRequirementIntegrationTests+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraDoDRequirementSecurityReviewed+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraDoDRequirementUnitTests+":duplicate")
	require.Contains(t, report.Failed, "requirements.manual_checklist_only:unexpected")
	require.Contains(t, report.Failed, "critical_requirements_unexpected_for_non_critical_change")
	require.Error(t, ValidateAetraDefinitionOfDone(evidence))
}

func TestAetraDefinitionOfDoneRejectsMissingCriticalRequirements(t *testing.T) {
	evidence := DefaultAetraDefinitionOfDoneEvidence("economics", true)
	evidence.CriticalRequirements = removeDoDItem(evidence.CriticalRequirements,
		AetraDoDCriticalRequirementAdversarialTests,
		AetraDoDCriticalRequirementExportImportTest,
		AetraDoDCriticalRequirementMigrationIfState,
	)
	evidence.CriticalRequirements = append(evidence.CriticalRequirements, AetraDoDCriticalRequirementInvariantTests, "manual_attack_review")

	report := BuildAetraDefinitionOfDoneReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "critical_requirements."+AetraDoDCriticalRequirementAdversarialTests+":missing")
	require.Contains(t, report.Failed, "critical_requirements."+AetraDoDCriticalRequirementExportImportTest+":missing")
	require.Contains(t, report.Failed, "critical_requirements."+AetraDoDCriticalRequirementMigrationIfState+":missing")
	require.Contains(t, report.Failed, "critical_requirements."+AetraDoDCriticalRequirementInvariantTests+":duplicate")
	require.Contains(t, report.Failed, "critical_requirements.manual_attack_review:unexpected")
	require.Error(t, ValidateAetraDefinitionOfDone(evidence))
}

func removeDoDItem(items []string, targets ...string) []string {
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
