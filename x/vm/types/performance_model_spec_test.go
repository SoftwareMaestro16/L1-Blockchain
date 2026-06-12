package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMPerformanceModelRequiresTargetProperties(t *testing.T) {
	model, err := DefaultAVMPerformanceModel()
	require.NoError(t, err)
	require.NoError(t, model.Validate())
	require.Contains(t, model.Targets, AVMPerformanceParallelZoneExecution)
	require.Contains(t, model.Targets, AVMPerformancePipelinedAsyncQueues)
	require.Contains(t, model.Targets, AVMPerformanceBatchedMessageExecution)
	require.Contains(t, model.Targets, AVMPerformanceMinimizedStoreV2Writes)
	require.Contains(t, model.Targets, AVMPerformanceLazyStateLoading)
	require.Contains(t, model.Targets, AVMPerformanceBoundedReceiptGeneration)
	require.Contains(t, model.Targets, AVMPerformanceActorLocalConflictIsolation)

	missing := model
	missing.Targets = missing.Targets[:len(missing.Targets)-1]
	missing.ModelHash = ComputeAVMPerformanceModelHash(missing)
	require.ErrorContains(t, missing.Validate(), "every target property")

	bad := model
	bad.Targets[0] = "unbounded_parallelism"
	bad.ModelHash = ComputeAVMPerformanceModelHash(bad)
	require.ErrorContains(t, bad.Validate(), "invalid AVM performance target")
}

func TestAVMBlockSTMPlanPartitionsParallelSafeWorkloads(t *testing.T) {
	workloads := []AVMBlockSTMWorkload{
		testAVMBlockSTMWorkload(t, "zone-a-actor-a", zonestypes.ZoneIDApplication, "actor-a", "", "queue-a", AVMBlockSTMConflictActorMailbox, "mailbox", 7),
		testAVMBlockSTMWorkload(t, "zone-b-actor-b", zonestypes.ZoneIDContract, "actor-b", "", "queue-b", AVMBlockSTMConflictActorMailbox, "mailbox", 9),
		testAVMBlockSTMWorkload(t, "zone-c-contract", zonestypes.ZoneIDFinancial, "", "contract-a", "queue-c", AVMBlockSTMConflictContractStorage, "balance", 11),
	}
	plan, err := NewAVMBlockSTMExecutionPlan(workloads)
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Len(t, plan.Partitions, 3)
	require.Len(t, plan.Accumulators, 3)
	require.Equal(t, ComputeAVMBlockSTMExecutionPlanHash(plan), plan.PlanHash)
}

func TestAVMBlockSTMRejectsConflictProneWorkloads(t *testing.T) {
	sameMailboxA := testAVMBlockSTMWorkload(t, "mailbox-a", zonestypes.ZoneIDContract, "actor-a", "", "actor-a-mailbox", AVMBlockSTMConflictActorMailbox, "mailbox", 1)
	sameMailboxB := testAVMBlockSTMWorkload(t, "mailbox-b", zonestypes.ZoneIDContract, "actor-a", "", "actor-a-mailbox", AVMBlockSTMConflictActorMailbox, "mailbox", 2)
	_, err := NewAVMBlockSTMExecutionPlan([]AVMBlockSTMWorkload{sameMailboxA, sameMailboxB})
	require.ErrorContains(t, err, "conflict key")

	storageA := testAVMBlockSTMWorkload(t, "storage-a", zonestypes.ZoneIDContract, "", "contract-a", "queue-a", AVMBlockSTMConflictContractStorage, "same-key", 3)
	storageB := testAVMBlockSTMWorkload(t, "storage-b", zonestypes.ZoneIDContract, "", "contract-a", "queue-b", AVMBlockSTMConflictContractStorage, "same-key", 4)
	_, err = NewAVMBlockSTMExecutionPlan([]AVMBlockSTMWorkload{storageA, storageB})
	require.ErrorContains(t, err, "conflict key")

	nonceA := testAVMBlockSTMWorkload(t, "nonce-a", zonestypes.ZoneIDApplication, "", "", "queue-a", AVMBlockSTMConflictSenderNonce, "app/alice/0001", 5)
	nonceB := testAVMBlockSTMWorkload(t, "nonce-b", zonestypes.ZoneIDApplication, "", "", "queue-b", AVMBlockSTMConflictSenderNonce, "app/alice/0001", 6)
	_, err = NewAVMBlockSTMExecutionPlan([]AVMBlockSTMWorkload{nonceA, nonceB})
	require.ErrorContains(t, err, "conflict key")
}

func TestAVMBlockSTMRequiresExpectedVersionsAndRejectsGlobalCounters(t *testing.T) {
	noVersion := AVMBlockSTMWorkload{
		WorkloadID:		"write-no-version",
		ZoneID:			zonestypes.ZoneIDContract,
		QueueIDOptional:	"queue-a",
		ExpectedVersion:	0,
		GasEstimate:		10,
		WritesState:		true,
		ConflictKeyKind:	AVMBlockSTMConflictZoneQueueHead,
	}
	_, err := NewAVMBlockSTMWorkload(noVersion)
	require.ErrorContains(t, err, "expected versions")

	globalCounter := noVersion
	globalCounter.WorkloadID = "global-counter"
	globalCounter.ExpectedVersion = 1
	globalCounter.UsesGlobalCounter = true
	_, err = NewAVMBlockSTMWorkload(globalCounter)
	require.ErrorContains(t, err, "global counters")
}

func TestAVMBlockSTMPerZoneAccumulatorBoundsReceiptsAndWrites(t *testing.T) {
	workloads := []AVMBlockSTMWorkload{
		testAVMBlockSTMWorkload(t, "app-1", zonestypes.ZoneIDApplication, "actor-a", "", "queue-a", AVMBlockSTMConflictActorMailbox, "mailbox-a", 10),
		testAVMBlockSTMWorkload(t, "app-2", zonestypes.ZoneIDApplication, "actor-b", "", "queue-b", AVMBlockSTMConflictActorMailbox, "mailbox-b", 20),
		testAVMBlockSTMWorkload(t, "contract-1", zonestypes.ZoneIDContract, "", "contract-a", "queue-c", AVMBlockSTMConflictContractStorage, "slot-a", 30),
	}
	plan, err := NewAVMBlockSTMExecutionPlan(workloads)
	require.NoError(t, err)
	require.Len(t, plan.Accumulators, 2)
	appAcc := plan.Accumulators[0]
	require.Equal(t, zonestypes.ZoneIDApplication, appAcc.ZoneID)
	require.Equal(t, uint32(2), appAcc.MessageCount)
	require.Equal(t, uint32(2), appAcc.ReceiptCount)
	require.Equal(t, uint32(2), appAcc.StoreWriteCount)
	require.Equal(t, uint64(30), appAcc.GasUsed)

	bad := appAcc
	bad.ReceiptCount = bad.MessageCount + 1
	bad.AccumulatorHash = ComputeAVMZoneExecutionAccumulatorHash(bad)
	require.ErrorContains(t, bad.Validate(), "receipt generation")
}

func testAVMBlockSTMWorkload(t *testing.T, id string, zoneID zonestypes.ZoneID, actorID, contractAddress, queueID string, kind AVMBlockSTMConflictKeyKind, key string, expectedVersion uint64) AVMBlockSTMWorkload {
	t.Helper()
	workload := AVMBlockSTMWorkload{
		WorkloadID:			id,
		ZoneID:				zoneID,
		ActorIDOptional:		actorID,
		ContractAddressOptional:	contractAddress,
		QueueIDOptional:		queueID,
		ExpectedVersion:		expectedVersion,
		GasEstimate:			expectedVersion,
		WritesState:			true,
		ConflictKeyKind:		kind,
	}
	switch kind {
	case AVMBlockSTMConflictContractStorage:
		workload.StorageKeyOptional = key
	case AVMBlockSTMConflictSenderNonce:
		workload.SenderNonceScopeOptional = key
	case AVMBlockSTMConflictPaymentEscrow:
		workload.PaymentEscrowOptional = key
	case AVMBlockSTMConflictContinuation:
		workload.ContinuationIDOptional = key
	case AVMBlockSTMConflictServiceCall:
		workload.ServiceCallOptional = key
	default:
		if workload.ActorIDOptional == "" {
			workload.ActorIDOptional = key
		}
	}
	built, err := NewAVMBlockSTMWorkload(workload)
	require.NoError(t, err)
	return built
}
