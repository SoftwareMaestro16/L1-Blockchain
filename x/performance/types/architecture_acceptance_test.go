package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchitectureAcceptanceReportPassesAllPlanningCriteria(t *testing.T) {
	report := BuildArchitectureAcceptanceReport(validArchitectureAcceptanceInput())
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
	require.NotEmpty(t, report.ReportHash)
}

func TestArchitectureAcceptanceReportRequiresAllCriteria(t *testing.T) {
	input := validArchitectureAcceptanceInput()
	input.Criteria = input.Criteria[:len(input.Criteria)-1]

	report := BuildArchitectureAcceptanceReport(input)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "architecture_acceptance_criteria")
	require.NoError(t, report.Validate())
}

func TestArchitectureAcceptanceReportRejectsUnreadyCoreRoots(t *testing.T) {
	input := validArchitectureAcceptanceInput()
	input.Criteria[0].Ready = false

	report := BuildArchitectureAcceptanceReport(input)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "architecture_acceptance_criteria")
}

func TestArchitectureAcceptanceReportRejectsUnbackedPaymentSettlement(t *testing.T) {
	input := validArchitectureAcceptanceInput()
	input.Criteria[8].ProofBacked = false

	report := BuildArchitectureAcceptanceReport(input)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "architecture_acceptance_criteria")
}

func TestArchitectureAcceptanceReportRejectsNondeterministicShardRules(t *testing.T) {
	input := validArchitectureAcceptanceInput()
	input.Criteria[5].Deterministic = false

	report := BuildArchitectureAcceptanceReport(input)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "architecture_acceptance_criteria")
}

func TestArchitectureAcceptanceReportRejectsUnexpectedCriterion(t *testing.T) {
	input := validArchitectureAcceptanceInput()
	input.Criteria[0].CriterionID = "external_route_oracle"

	report := BuildArchitectureAcceptanceReport(input)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "architecture_acceptance_criteria")
}

func validArchitectureAcceptanceInput() ArchitectureAcceptanceInput {
	return ArchitectureAcceptanceInput{
		AcceptanceVersion:	"acceptance_18",
		Criteria: []ArchitectureAcceptanceCriterion{
			architectureAcceptance(AcceptanceAetraCoreRoots, "x_aetracore_roots"),
			architectureAcceptance(AcceptanceZoneAdapterExecution, "x_zones_adapter"),
			architectureAcceptance(AcceptanceMessageProofReceipts, "x_msgbus_receipts"),
			architectureAcceptance(AcceptanceStoreV2ZoneShardProof, "store_v2_proofs"),
			architectureAcceptance(AcceptanceBlockSTMParallelism, "blockstm_shards"),
			architectureAcceptance(AcceptanceShardSplitMergeRules, "x_shards_scheduler"),
			architectureAcceptance(AcceptanceAVM20Spec, "x_aetravm_avm"),
			architectureAcceptance(AcceptanceIdentityProofLookup, "x_identity_zone"),
			architectureAcceptance(AcceptancePaymentSettlement, "x_payments_settlement"),
			architectureAcceptance(AcceptanceMigrationPreservation, "x_migration_invariants"),
		},
	}
}

func architectureAcceptance(criterionID, component string) ArchitectureAcceptanceCriterion {
	return ArchitectureAcceptanceCriterion{
		CriterionID:	criterionID,
		Component:	component,
		EvidenceHash:	hashStrings("architecture-acceptance-evidence", criterionID),
		TestHash:	hashStrings("architecture-acceptance-test", criterionID),
		Ready:		true,
		Deterministic:	true,
		ProofBacked:	true,
	}
}
