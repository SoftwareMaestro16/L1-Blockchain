package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraEngineeringPrioritiesCoverSection35(t *testing.T) {
	report := BuildAetraEngineeringPrioritiesReport(nil)

	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.True(t, report.P3Allowed)
	require.Equal(t, 21, report.Required)
	require.Equal(t, report.Required, report.Passed)
	require.NoError(t, ValidateAetraEngineeringPriorities(nil))

	byPriority := map[string][]string{}
	for _, priority := range report.Priorities {
		byPriority[priority.Priority] = priority.Items
	}
	require.Contains(t, byPriority[AetraEngineeringPriorityP0], AetraEngineeringP0ConsensusSafety)
	require.Contains(t, byPriority[AetraEngineeringPriorityP0], AetraEngineeringP0ExportImport)
	require.Contains(t, byPriority[AetraEngineeringPriorityP1], AetraEngineeringP1ValidatorPowerCap)
	require.Contains(t, byPriority[AetraEngineeringPriorityP1], AetraEngineeringP1GovernanceBounds)
	require.Contains(t, byPriority[AetraEngineeringPriorityP2], AetraEngineeringP2AVMHardening)
	require.Contains(t, byPriority[AetraEngineeringPriorityP2], AetraEngineeringP2PublicTestnetDocs)
	require.Contains(t, byPriority[AetraEngineeringPriorityP3], AetraEngineeringP3EncryptedMempoolResearch)
	require.Contains(t, byPriority[AetraEngineeringPriorityP3], AetraEngineeringP3HigherValidatorCapExperiments)
}

func TestAetraEngineeringPrioritiesRejectP3BeforeP0P1Stable(t *testing.T) {
	evidence := DefaultAetraEngineeringPrioritiesEvidence()
	evidence[0].Stable = false
	evidence[1].Stable = true

	report := BuildAetraEngineeringPrioritiesReport(evidence)
	require.False(t, report.Ready)
	require.False(t, report.P3Allowed)
	require.Contains(t, report.Failed, "p3_requires_p0_and_p1_stable")
	require.Error(t, ValidateAetraEngineeringPriorities(evidence))

	evidence[0].Stable = true
	evidence[1].Stable = false
	report = BuildAetraEngineeringPrioritiesReport(evidence)
	require.False(t, report.Ready)
	require.False(t, report.P3Allowed)
	require.Contains(t, report.Failed, "p3_requires_p0_and_p1_stable")
}

func TestAetraEngineeringPrioritiesRejectMissingDuplicateUnexpectedItems(t *testing.T) {
	evidence := DefaultAetraEngineeringPrioritiesEvidence()
	evidence[0].Items = removeEngineeringPriorityItem(evidence[0].Items,
		AetraEngineeringP0ConsensusSafety,
		AetraEngineeringP0SupplyInvariants,
	)
	evidence[0].Items = append(evidence[0].Items, AetraEngineeringP0ExportImport, "manual_foundation_note")
	evidence[1].Priority = "PX"

	report := BuildAetraEngineeringPrioritiesReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraEngineeringPriorityP0+"."+AetraEngineeringP0ConsensusSafety+":missing")
	require.Contains(t, report.Failed, AetraEngineeringPriorityP0+"."+AetraEngineeringP0SupplyInvariants+":missing")
	require.Contains(t, report.Failed, AetraEngineeringPriorityP0+"."+AetraEngineeringP0ExportImport+":duplicate")
	require.Contains(t, report.Failed, AetraEngineeringPriorityP0+".manual_foundation_note:unexpected")
	require.Contains(t, report.Failed, "PX:unknown_priority")
	require.Contains(t, report.Failed, AetraEngineeringPriorityP1+":missing_priority")
}

func TestAetraEngineeringPrioritiesRejectDuplicatePriority(t *testing.T) {
	evidence := DefaultAetraEngineeringPrioritiesEvidence()
	evidence[1].Priority = evidence[0].Priority
	evidence[1].Items = RequiredAetraEngineeringP0Items()

	report := BuildAetraEngineeringPrioritiesReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraEngineeringPriorityP0+":duplicate_priority")
	require.Contains(t, report.Failed, AetraEngineeringPriorityP1+":missing_priority")
}

func removeEngineeringPriorityItem(items []string, targets ...string) []string {
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
