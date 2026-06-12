package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMUpgradeStateRegistryTracksScheduledStateAndRuntimeVersions(t *testing.T) {
	runtimeVersions := testAVMRuntimeVersionSet(t, "v1")
	state, err := NewAVMScheduledUpgradeState(AVMScheduledUpgradeState{
		UpgradeID:		"upgrade-153",
		Component:		AVMUpgradeComponentSchedulerRules,
		FromVersion:		"scheduler-v1",
		ToVersion:		"scheduler-v2",
		ActivationHeight:	150,
		MigrationRequired:	true,
		CompatibilityMode:	AVMUpgradeCompatibilityVersionedPolicy,
		Status:			AVMUpgradeStatusScheduled,
	})
	require.NoError(t, err)
	gasTable := testAVMVersionedGasTable(t)

	registry, err := NewAVMUpgradeStateRegistry(AVMUpgradeStateRegistry{
		RuntimeVersions:	runtimeVersions,
		States:			[]AVMScheduledUpgradeState{state},
		GasTable:		gasTable,
	})
	require.NoError(t, err)
	require.NoError(t, registry.Validate())
	require.Equal(t, ComputeAVMUpgradeStateRegistryHash(registry), registry.RegistryHash)
	require.Equal(t, "scheduler-v2", registry.States[0].ToVersion)
}

func TestAVMUpgradeMigrationHandlersForQueuesAndContinuations(t *testing.T) {
	msg := testAVMUpgradeMessage(t, "queue-migrate", 1, 90)
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, msg, 91, 16)
	require.NoError(t, err)
	queueMigration, err := BuildAVMQueueUpgradeMigration("upgrade-153", "queue-v1", "queue-v2", queue)
	require.NoError(t, err)
	require.Equal(t, AVMUpgradeMigrationQueue, queueMigration.Kind)
	require.Equal(t, uint32(1), queueMigration.MigratedCount)
	require.True(t, queueMigration.Bounded)

	continuations := []ContinuationRecord{{
		ContinuationID:		"continuation-153",
		ActorID:		"actor-153",
		StepIndex:		1,
		PartialStateHash:	engineHash("partial-153"),
		PartialStateBytes:	64,
		ResumeHeight:		160,
		ExpiryHeight:		200,
		GasReserved:		1000,
		Status:			ContinuationStatusScheduled,
		ResumeBy:		ContinuationResumeByScheduler,
	}}
	continuationMigration, err := BuildAVMContinuationUpgradeMigration("upgrade-153", "runtime-v1", "runtime-v2", continuations)
	require.NoError(t, err)
	require.Equal(t, AVMUpgradeMigrationContinuation, continuationMigration.Kind)
	require.Equal(t, uint32(1), continuationMigration.MigratedCount)
	require.NotEqual(t, continuationMigration.SourceRoot, continuationMigration.TargetRoot)
}

func TestAVMVersionedGasTableSelectsActivationByHeight(t *testing.T) {
	table := testAVMVersionedGasTable(t)
	before, err := table.ActiveAt(99)
	require.NoError(t, err)
	require.Equal(t, "gas-v1", before.PolicyVersion)

	after, err := table.ActiveAt(100)
	require.NoError(t, err)
	require.Equal(t, "gas-v2", after.PolicyVersion)
	require.Greater(t, after.Policy.QueueInsertGas, before.Policy.QueueInsertGas)

	_, err = table.ActiveAt(0)
	require.ErrorContains(t, err, "height")
}

func TestAVMPendingMessagesKeepCompatibilityAcrossUpgrade(t *testing.T) {
	pre := testAVMUpgradeMessage(t, "pending-pre", 1, 90)
	post := testAVMUpgradeMessage(t, "pending-post", 2, 101)
	policies, err := BuildAVMPendingMessageCompatibilityPolicies(
		[]AVMAsyncMessage{post, pre},
		100,
		120,
		"runtime-v1",
		"runtime-v2",
		"scheduler-v1",
		"scheduler-v2",
		"gas-v1",
		"gas-v2",
	)
	require.NoError(t, err)
	require.Len(t, policies, 2)

	byID := map[string]AVMVersionedMessageExecutionPolicy{}
	for _, policy := range policies {
		byID[policy.MessageID] = policy
	}
	require.Equal(t, "runtime-v1", byID[pre.ID].RuntimeVersion)
	require.Equal(t, "scheduler-v1", byID[pre.ID].SchedulerVersion)
	require.Equal(t, "gas-v1", byID[pre.ID].GasPolicyVersion)
	require.Equal(t, "runtime-v2", byID[post.ID].RuntimeVersion)
	require.Equal(t, "scheduler-v2", byID[post.ID].SchedulerVersion)
	require.Equal(t, "gas-v2", byID[post.ID].GasPolicyVersion)
}

func TestAVMActivatedUpgradePreventsRollback(t *testing.T) {
	active, err := NewAVMScheduledUpgradeState(AVMScheduledUpgradeState{
		UpgradeID:		"upgrade-forward",
		Component:		AVMUpgradeComponentVMInterpreter,
		FromVersion:		"vm-v1",
		ToVersion:		"vm-v2",
		ActivationHeight:	100,
		MigrationRequired:	false,
		CompatibilityMode:	AVMUpgradeCompatibilityNone,
		Status:			AVMUpgradeStatusCompleted,
	})
	require.NoError(t, err)
	rollback, err := NewAVMScheduledUpgradeState(AVMScheduledUpgradeState{
		UpgradeID:		"upgrade-rollback",
		Component:		AVMUpgradeComponentVMInterpreter,
		FromVersion:		"vm-v2",
		ToVersion:		"vm-v1",
		ActivationHeight:	200,
		MigrationRequired:	false,
		CompatibilityMode:	AVMUpgradeCompatibilityNone,
		Status:			AVMUpgradeStatusActive,
	})
	require.NoError(t, err)
	require.ErrorContains(t, ValidateAVMUpgradeRollbackPrevention([]AVMScheduledUpgradeState{active, rollback}), "cannot roll back")

	registry, err := NewAVMUpgradeStateRegistry(AVMUpgradeStateRegistry{
		RuntimeVersions:	testAVMRuntimeVersionSet(t, "v2"),
		States:			[]AVMScheduledUpgradeState{active, rollback},
		GasTable:		testAVMVersionedGasTable(t),
	})
	require.ErrorContains(t, err, "cannot roll back")
	require.NotEmpty(t, registry.RegistryHash)
}

func testAVMRuntimeVersionSet(t *testing.T, suffix string) AVMRuntimeVersionSet {
	t.Helper()
	versions, err := NewAVMRuntimeVersionSet(AVMRuntimeVersionSet{
		VMInterpreterVersion:	"vm-" + suffix,
		SchedulerVersion:	"scheduler-" + suffix,
		GasPolicyVersion:	"gas-" + suffix,
		ZoneConfigVersion:	"zone-" + suffix,
		BackendAdapterVersion:	"backend-" + suffix,
		InterfaceSchemaVersion:	"interface-" + suffix,
		RetryPolicyVersion:	"retry-" + suffix,
		QueueLimitVersion:	"queue-" + suffix,
	})
	require.NoError(t, err)
	return versions
}

func testAVMVersionedGasTable(t *testing.T) AVMVersionedGasTable {
	t.Helper()
	policyV1, err := DefaultAVMGasPolicy()
	require.NoError(t, err)
	scheduleV1, err := AVMGasScheduleFromPolicy(policyV1, true, 1_000_000)
	require.NoError(t, err)
	tableV1, err := NewAVMGasTableActivation(AVMGasTableActivation{
		ActivationHeight:	1,
		PolicyVersion:		"gas-v1",
		Policy:			policyV1,
		Schedule:		scheduleV1,
	})
	require.NoError(t, err)

	policyV2 := policyV1
	policyV2.QueueInsertGas += 25
	policyV2.PolicyHash = ComputeAVMGasPolicyHash(policyV2)
	scheduleV2, err := AVMGasScheduleFromPolicy(policyV2, true, 1_000_000)
	require.NoError(t, err)
	tableV2, err := NewAVMGasTableActivation(AVMGasTableActivation{
		ActivationHeight:	100,
		PolicyVersion:		"gas-v2",
		Policy:			policyV2,
		Schedule:		scheduleV2,
	})
	require.NoError(t, err)

	table, err := NewAVMVersionedGasTable([]AVMGasTableActivation{tableV2, tableV1})
	require.NoError(t, err)
	return table
}
