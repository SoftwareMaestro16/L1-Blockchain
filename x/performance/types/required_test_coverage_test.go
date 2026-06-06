package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredTestCoverageReportPassesDeterminismAndInvariants(t *testing.T) {
	report := BuildRequiredTestCoverageReport(validRequiredTestCoverageInput())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
	require.NotEmpty(t, report.ReportHash)
}

func TestRequiredTestCoverageReportRequiresAllDeterminismCases(t *testing.T) {
	input := validRequiredTestCoverageInput()
	input.Determinism = input.Determinism[:len(input.Determinism)-1]

	report := BuildRequiredTestCoverageReport(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "determinism_tests")
	require.NoError(t, report.Validate())
}

func TestRequiredTestCoverageReportRejectsUncoveredInvariant(t *testing.T) {
	input := validRequiredTestCoverageInput()
	input.Invariants[2].Covered = false

	report := BuildRequiredTestCoverageReport(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "invariant_tests")
}

func TestRequiredTestCoverageReportRejectsMismatchedClass(t *testing.T) {
	input := validRequiredTestCoverageInput()
	input.Determinism[0].Class = RequiredCoverageInvariant

	report := BuildRequiredTestCoverageReport(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "determinism_tests")
}

func TestRequiredTestCoverageReportRejectsDuplicateInvariant(t *testing.T) {
	input := validRequiredTestCoverageInput()
	input.Invariants[1].CaseID = input.Invariants[0].CaseID

	report := BuildRequiredTestCoverageReport(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "invariant_tests")
}

func validRequiredTestCoverageInput() RequiredTestCoverageInput {
	return RequiredTestCoverageInput{
		CoverageVersion: "required_17",
		Determinism: []RequiredTestCoverageCase{
			requiredCoverage(RequiredDeterminismZoneRoots, RequiredCoverageDeterminism, "zone_roots"),
			requiredCoverage(RequiredDeterminismMessageRoots, RequiredCoverageDeterminism, "message_roots"),
			requiredCoverage(RequiredDeterminismRoutingPaths, RequiredCoverageDeterminism, "routing_paths"),
			requiredCoverage(RequiredDeterminismShardIDs, RequiredCoverageDeterminism, "shard_ids"),
			requiredCoverage(RequiredDeterminismVMOutput, RequiredCoverageDeterminism, "vm_output"),
		},
		Invariants: []RequiredTestCoverageCase{
			requiredCoverage(RequiredInvariantZoneRootIncludesShardRoots, RequiredCoverageInvariant, "zone_root"),
			requiredCoverage(RequiredInvariantOutboxReceiptOrPending, RequiredCoverageInvariant, "message_outbox"),
			requiredCoverage(RequiredInvariantCrossZoneValueConservation, RequiredCoverageInvariant, "cross_zone_value"),
			requiredCoverage(RequiredInvariantPaymentCollateralOverpay, RequiredCoverageInvariant, "payment_collateral"),
			requiredCoverage(RequiredInvariantIdentityProofRootMatch, RequiredCoverageInvariant, "identity_proof"),
			requiredCoverage(RequiredInvariantContractProofRootMatch, RequiredCoverageInvariant, "contract_proof"),
			requiredCoverage(RequiredInvariantShardSplitPreservesKeys, RequiredCoverageInvariant, "shard_split"),
			requiredCoverage(RequiredInvariantShardMergePreservesKeys, RequiredCoverageInvariant, "shard_merge"),
		},
	}
}

func requiredCoverage(caseID string, class RequiredCoverageClass, target string) RequiredTestCoverageCase {
	return RequiredTestCoverageCase{
		CaseID:        caseID,
		Class:         class,
		Target:        target,
		EvidenceHash:  hashStrings("required-test-coverage", caseID),
		Deterministic: true,
		Covered:       true,
	}
}
