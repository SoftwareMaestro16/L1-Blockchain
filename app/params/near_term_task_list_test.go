package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraNearTermTaskListCoversSection36(t *testing.T) {
	evidence := DefaultAetraNearTermTaskListEvidence()

	report := BuildAetraNearTermTaskListReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, 20, report.Required)
	require.Equal(t, report.Required, report.Passed)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskAuditExistingModules)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskPowerCapCommissionParams)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskEffectivePowerQueries)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskConcentrationSnapshotQuery)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskCapMathTests)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskFeeSplitAccountingTests)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskInflationBoundsTests)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskSupplyInvariantTests)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskValidatorScoreStateQueries)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskProgressiveDowntimeDecision)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskNominationPoolAccounting)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskAVMSmokeMalicious)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskFinalityMeasurementScript)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskValidatorDelegatorDocs)
	require.Contains(t, evidence.Tasks, AetraNearTermTaskCriticalCIGate)
	require.Contains(t, evidence.Checklist, AetraNearTermChecklistConsensusEconomicsBehavior)
	require.Contains(t, evidence.Checklist, AetraNearTermChecklistParamsAddedChanged)
	require.Contains(t, evidence.Checklist, AetraNearTermChecklistTestsAdded)
	require.Contains(t, evidence.Checklist, AetraNearTermChecklistMigrationRisk)
	require.Contains(t, evidence.Checklist, AetraNearTermChecklistPublicDocs)
	require.NoError(t, ValidateAetraNearTermTaskList(evidence))
}

func TestAetraNearTermTaskListRejectsMissingDuplicateUnexpectedTasks(t *testing.T) {
	evidence := DefaultAetraNearTermTaskListEvidence()
	evidence.Tasks = removeNearTermItem(evidence.Tasks,
		AetraNearTermTaskAuditExistingModules,
		AetraNearTermTaskCapMathTests,
		AetraNearTermTaskCriticalCIGate,
	)
	evidence.Checklist = removeNearTermItem(evidence.Checklist,
		AetraNearTermChecklistTestsAdded,
		AetraNearTermChecklistMigrationRisk,
	)
	evidence.Tasks = append(evidence.Tasks, AetraNearTermTaskInflationBoundsTests, "manual_backlog_note")
	evidence.Checklist = append(evidence.Checklist, AetraNearTermChecklistPublicDocs, "chat_summary_only")

	report := BuildAetraNearTermTaskListReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "tasks."+AetraNearTermTaskAuditExistingModules+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraNearTermTaskCapMathTests+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraNearTermTaskCriticalCIGate+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraNearTermTaskInflationBoundsTests+":duplicate")
	require.Contains(t, report.Failed, "tasks.manual_backlog_note:unexpected")
	require.Contains(t, report.Failed, "checklist."+AetraNearTermChecklistTestsAdded+":missing")
	require.Contains(t, report.Failed, "checklist."+AetraNearTermChecklistMigrationRisk+":missing")
	require.Contains(t, report.Failed, "checklist."+AetraNearTermChecklistPublicDocs+":duplicate")
	require.Contains(t, report.Failed, "checklist.chat_summary_only:unexpected")
	require.Error(t, ValidateAetraNearTermTaskList(evidence))
}

func removeNearTermItem(items []string, targets ...string) []string {
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
