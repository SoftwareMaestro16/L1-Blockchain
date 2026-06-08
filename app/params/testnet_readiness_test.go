package params

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultTestnetReadinessChecklistIsComplete(t *testing.T) {
	checklist := DefaultTestnetReadinessChecklist()

	require.NoError(t, ValidateTestnetReadinessChecklist(checklist))
	require.Len(t, checklist, len(RequiredTestnetReadinessGateIDs()))

	seenJobs := map[string]struct{}{}
	seenCommands := map[string]struct{}{}
	for _, gate := range checklist {
		require.True(t, gate.Required, gate.ID)
		require.NotContains(t, gate.Command, "TODO", gate.ID)
		require.NotContains(t, gate.Description, "TODO", gate.ID)
		if _, found := seenJobs[gate.CIJob]; found {
			t.Fatalf("duplicate CI job %s", gate.CIJob)
		}
		seenJobs[gate.CIJob] = struct{}{}
		if _, found := seenCommands[gate.Command]; found {
			t.Fatalf("duplicate command %s", gate.Command)
		}
		seenCommands[gate.Command] = struct{}{}
	}
}

func TestTestnetReadinessEvidenceRequiresEveryGate(t *testing.T) {
	passed := map[string]bool{}
	for _, id := range RequiredTestnetReadinessGateIDs() {
		passed[id] = true
	}
	require.NoError(t, ValidateTestnetReadinessEvidence(passed))

	passed[TestnetReadinessExportImportRoundtrip] = false
	report := BuildTestnetReadinessReport(passed)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, TestnetReadinessExportImportRoundtrip)
	require.ErrorContains(t, ValidateTestnetReadinessEvidence(passed), TestnetReadinessExportImportRoundtrip)
}

func TestTestnetReadinessWorkflowContainsEveryRequiredJob(t *testing.T) {
	workflow := readRepoFileForParamsTest(t, ".github/workflows/testnet-readiness.yml")
	for _, gate := range DefaultTestnetReadinessChecklist() {
		require.Contains(t, workflow, gate.CIJob, gate.ID)
		require.Contains(t, workflow, gate.Command, gate.ID)
	}
	require.Contains(t, workflow, "go test ./...")
	require.Contains(t, workflow, "genesis validate")
	require.Contains(t, workflow, "localnet smoke")
	require.Contains(t, workflow, "export/import roundtrip")
	require.Contains(t, workflow, "release artifact")
}

func TestValidatorDocumentationCoversMinimumOperatorRunbook(t *testing.T) {
	docs := readRepoFileForParamsTest(t, "docs/validator-onboarding.md")
	for _, term := range []string{
		"version --long --output json",
		"chain-id",
		"genesis validate-genesis",
		"keyring-backend os",
		"state sync",
		"snapshot",
		"create-validator",
		"query staking validators",
		"missed block",
		"export/import",
		"upgrade",
	} {
		require.Contains(t, strings.ToLower(docs), strings.ToLower(term), term)
	}
}

func readRepoFileForParamsTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile("../../" + path)
	require.NoError(t, err)
	return string(data)
}
