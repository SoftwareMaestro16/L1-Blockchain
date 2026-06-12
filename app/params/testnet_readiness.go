package params

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	TestnetReadinessGoTest			= "ci_go_test_all"
	TestnetReadinessGenesisValidate		= "ci_genesis_validate"
	TestnetReadinessLocalnetSmoke		= "ci_localnet_smoke"
	TestnetReadinessExportImportRoundtrip	= "ci_export_import_roundtrip"
	TestnetReadinessInvariants		= "ci_invariants"
	TestnetReadinessLinter			= "ci_linter"
	TestnetReadinessReleaseArtifact		= "ci_release_artifact_build"
	TestnetReadinessVersionCommand		= "node_binary_version_command"
	TestnetReadinessChainIDValidation	= "chain_id_validation"
	TestnetReadinessValidatorDocs		= "validator_documentation_minimum"
)

type TestnetReadinessGate struct {
	ID		string
	Description	string
	CIJob		string
	Command		string
	Required	bool
}

type TestnetReadinessReport struct {
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultTestnetReadinessChecklist() []TestnetReadinessGate {
	return []TestnetReadinessGate{
		{ID: TestnetReadinessGoTest, Description: "all Go packages compile and tests pass", CIJob: "go-test-all", Command: "go test ./...", Required: true},
		{ID: TestnetReadinessGenesisValidate, Description: "generated localnet genesis validates before boot", CIJob: "genesis-validate", Command: "scripts/localnet/validate-genesis.ps1", Required: true},
		{ID: TestnetReadinessLocalnetSmoke, Description: "localnet reaches height and validator set is live", CIJob: "localnet-smoke", Command: "tests/e2e/localnet_smoke.ps1", Required: true},
		{ID: TestnetReadinessExportImportRoundtrip, Description: "exported state validates and imports into a fresh node flow", CIJob: "export-import-roundtrip", Command: "tests/e2e/export_import_smoke.ps1", Required: true},
		{ID: TestnetReadinessInvariants, Description: "app-level invariants are registered and executable", CIJob: "invariants", Command: "go test ./app -run Invariant", Required: true},
		{ID: TestnetReadinessLinter, Description: "static linter gate for Go and protobuf surfaces", CIJob: "linter", Command: "go vet ./... && buf lint", Required: true},
		{ID: TestnetReadinessReleaseArtifact, Description: "release package builds signed-version node artifacts", CIJob: "release-artifact-build", Command: "scripts/release/prototype-package.ps1", Required: true},
		{ID: TestnetReadinessVersionCommand, Description: "node binary exposes version metadata", CIJob: "version-command", Command: "aetrad version --long --output json", Required: true},
		{ID: TestnetReadinessChainIDValidation, Description: "chain-id naming policy rejects malformed or wrong-network IDs", CIJob: "chain-id-validation", Command: "ValidateAetraTestnetChainID", Required: true},
		{ID: TestnetReadinessValidatorDocs, Description: "validator onboarding docs cover build, genesis, keys, sync, staking, monitoring, and upgrades", CIJob: "validator-docs", Command: "docs/validator-onboarding.md", Required: true},
	}
}

func ValidateTestnetReadinessChecklist(checklist []TestnetReadinessGate) error {
	if len(checklist) == 0 {
		return errors.New("testnet readiness checklist is required")
	}
	seen := map[string]struct{}{}
	for _, gate := range checklist {
		if strings.TrimSpace(gate.ID) == "" || strings.TrimSpace(gate.Description) == "" || strings.TrimSpace(gate.CIJob) == "" || strings.TrimSpace(gate.Command) == "" {
			return errors.New("testnet readiness gate id, description, CI job, and command are required")
		}
		if _, found := seen[gate.ID]; found {
			return fmt.Errorf("duplicate testnet readiness gate %s", gate.ID)
		}
		seen[gate.ID] = struct{}{}
	}
	for _, required := range RequiredTestnetReadinessGateIDs() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("missing testnet readiness gate %s", required)
		}
	}
	return nil
}

func RequiredTestnetReadinessGateIDs() []string {
	ids := []string{
		TestnetReadinessGoTest,
		TestnetReadinessGenesisValidate,
		TestnetReadinessLocalnetSmoke,
		TestnetReadinessExportImportRoundtrip,
		TestnetReadinessInvariants,
		TestnetReadinessLinter,
		TestnetReadinessReleaseArtifact,
		TestnetReadinessVersionCommand,
		TestnetReadinessChainIDValidation,
		TestnetReadinessValidatorDocs,
	}
	sort.Strings(ids)
	return ids
}

func BuildTestnetReadinessReport(passed map[string]bool) TestnetReadinessReport {
	failed := make([]string, 0)
	passedCount := 0
	for _, gate := range DefaultTestnetReadinessChecklist() {
		if !gate.Required {
			continue
		}
		if passed[gate.ID] {
			passedCount++
		} else {
			failed = append(failed, gate.ID)
		}
	}
	sort.Strings(failed)
	return TestnetReadinessReport{
		Required:	len(RequiredTestnetReadinessGateIDs()),
		Passed:		passedCount,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func ValidateTestnetReadinessEvidence(passed map[string]bool) error {
	report := BuildTestnetReadinessReport(passed)
	if !report.Ready {
		return fmt.Errorf("testnet readiness failed: %v", report.Failed)
	}
	return nil
}
