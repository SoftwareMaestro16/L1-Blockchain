package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImplementationBacklogReportPassesHighAndMediumPriorityTasks(t *testing.T) {
	report := BuildImplementationBacklogReport(validImplementationBacklogInput())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
	require.NotEmpty(t, report.ReportHash)
}

func TestImplementationBacklogReportRequiresAllHighPriorityTasks(t *testing.T) {
	input := validImplementationBacklogInput()
	input.HighPriority = input.HighPriority[:len(input.HighPriority)-1]

	report := BuildImplementationBacklogReport(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "high_priority_backlog")
	require.NoError(t, report.Validate())
}

func TestImplementationBacklogReportRejectsNondeterministicMediumTask(t *testing.T) {
	input := validImplementationBacklogInput()
	input.MediumPriority[3].Deterministic = false

	report := BuildImplementationBacklogReport(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "medium_priority_backlog")
}

func TestImplementationBacklogReportRejectsUnexpectedTask(t *testing.T) {
	input := validImplementationBacklogInput()
	input.HighPriority[0].TaskID = "external_oracle_route"

	report := BuildImplementationBacklogReport(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "high_priority_backlog")
}

func validImplementationBacklogInput() ImplementationBacklogInput {
	return ImplementationBacklogInput{
		BacklogVersion: "backlog_16",
		HighPriority: []ImplementationBacklogTaskCheck{
			backlogTask(BacklogTaskZoneDescriptors, ImplementationBacklogHigh, "x_aethercore_types"),
			backlogTask(BacklogTaskAetherCoreSkeleton, ImplementationBacklogHigh, "x_aethercore_module"),
			backlogTask(BacklogTaskGlobalRootHierarchy, ImplementationBacklogHigh, "x_aethercore_roots"),
			backlogTask(BacklogTaskMsgBusMessageEncoding, ImplementationBacklogHigh, "x_msgbus_encoding"),
			backlogTask(BacklogTaskLocalMessageStores, ImplementationBacklogHigh, "x_msgbus_stores"),
			backlogTask(BacklogTaskStoreV2KeyPrefixPlan, ImplementationBacklogHigh, "store_v2_prefixes"),
			backlogTask(BacklogTaskBlockSTMZoneBatchConflictTests, ImplementationBacklogHigh, "blockstm_zone_batches"),
			backlogTask(BacklogTaskDeterministicRoutingTable, ImplementationBacklogHigh, "x_routing_tables"),
			backlogTask(BacklogTaskProofRegistrySchema, ImplementationBacklogHigh, "x_aethercore_proofs"),
		},
		MediumPriority: []ImplementationBacklogTaskCheck{
			backlogTask(BacklogTaskFinancialZoneExtraction, ImplementationBacklogMedium, "x_zones_financial"),
			backlogTask(BacklogTaskIdentityZoneActivation, ImplementationBacklogMedium, "x_identity_zone"),
			backlogTask(BacklogTaskPerZoneMempoolLanes, ImplementationBacklogMedium, "mempool_lanes"),
			backlogTask(BacklogTaskPerShardFeeAccumulators, ImplementationBacklogMedium, "fees_shards"),
			backlogTask(BacklogTaskShardSplitMergeScheduler, ImplementationBacklogMedium, "x_shards_scheduler"),
			backlogTask(BacklogTaskAVMBytecodeGasTable, ImplementationBacklogMedium, "x_aetherisvm_avm"),
			backlogTask(BacklogTaskPaymentSettlementState, ImplementationBacklogMedium, "x_payments_state"),
			backlogTask(BacklogTaskCrossZoneIdentityLookup, ImplementationBacklogMedium, "x_identity_lookup"),
		},
	}
}

func backlogTask(taskID string, priority ImplementationBacklogPriority, component string) ImplementationBacklogTaskCheck {
	return ImplementationBacklogTaskCheck{
		TaskID:        taskID,
		Priority:      priority,
		Component:     component,
		EvidenceHash:  hashStrings("implementation-backlog-evidence", taskID),
		TestHash:      hashStrings("implementation-backlog-test", taskID),
		Implemented:   true,
		Deterministic: true,
	}
}
