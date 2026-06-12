package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShardRebalanceSchedulesSplitFromCommittedMetrics(t *testing.T) {
	layout := routingTestLayout(t, ZoneIDFinancial, 11, ShardAssignmentConsistentHash)
	metrics := []ShardMetrics{
		shardMetric(t, ZoneIDFinancial, "0", 38, 10, 1, 1, 100, 10),
		shardMetric(t, ZoneIDFinancial, "0", 39, 10, 1, 1, 100, 10),
		shardMetric(t, ZoneIDFinancial, "0", 40, 10, 1, 1, 100, 10),
		shardMetric(t, ZoneIDFinancial, "1", 38, 950, 2, 3, 512, 20),
		shardMetric(t, ZoneIDFinancial, "1", 39, 930, 2, 3, 512, 20),
		shardMetric(t, ZoneIDFinancial, "1", 40, 920, 2, 3, 512, 20),
		shardMetric(t, ZoneIDFinancial, "2", 38, 20, 1, 1, 100, 10),
		shardMetric(t, ZoneIDFinancial, "2", 39, 20, 1, 1, 100, 10),
		shardMetric(t, ZoneIDFinancial, "2", 40, 20, 1, 1, 100, 10),
	}

	decision, err := NewShardRebalanceDecision(layout, metrics, shardRebalanceThresholds(), 40)
	require.NoError(t, err)
	require.Equal(t, ShardLayoutChangeSplit, decision.ChangeKind)
	require.Equal(t, uint64(11), decision.SourceLayoutEpoch)
	require.Equal(t, uint64(13), decision.TargetLayoutEpoch)
	require.Equal(t, uint64(45), decision.ActivationHeight)
	require.Equal(t, []ShardID{ShardID("1")}, decision.SourceShardIDs)
	require.Len(t, decision.MigrationTasks, 1)
	require.Equal(t, uint64(13), decision.MigrationTasks[0].DeliveryEpoch)
	require.NoError(t, decision.ValidateHash())

	reordered, err := NewShardRebalanceDecision(layout, []ShardMetrics{metrics[8], metrics[2], metrics[0], metrics[6], metrics[4], metrics[1], metrics[5], metrics[3], metrics[7]}, shardRebalanceThresholds(), 40)
	require.NoError(t, err)
	require.Equal(t, decision.DecisionHash, reordered.DecisionHash)
}

func TestShardRebalanceSchedulesMergeAndRejectsMidBlockEpoch(t *testing.T) {
	layout := routingTestLayout(t, ZoneIDApplication, 5, ShardAssignmentConsistentHash)
	metrics := []ShardMetrics{
		shardMetric(t, ZoneIDApplication, "0", 48, 10, 0, 0, 64, 5),
		shardMetric(t, ZoneIDApplication, "0", 49, 10, 0, 0, 64, 5),
		shardMetric(t, ZoneIDApplication, "0", 50, 10, 0, 0, 64, 5),
		shardMetric(t, ZoneIDApplication, "1", 48, 15, 0, 0, 64, 5),
		shardMetric(t, ZoneIDApplication, "1", 49, 15, 0, 0, 64, 5),
		shardMetric(t, ZoneIDApplication, "1", 50, 15, 0, 0, 64, 5),
		shardMetric(t, ZoneIDApplication, "2", 48, 600, 0, 0, 128, 5),
		shardMetric(t, ZoneIDApplication, "2", 49, 600, 0, 0, 128, 5),
		shardMetric(t, ZoneIDApplication, "2", 50, 600, 0, 0, 128, 5),
	}

	decision, err := NewShardRebalanceDecision(layout, metrics, shardRebalanceThresholds(), 50)
	require.NoError(t, err)
	require.Equal(t, ShardLayoutChangeMerge, decision.ChangeKind)
	require.Equal(t, []ShardID{ShardID("0"), ShardID("1")}, decision.SourceShardIDs)
	require.Equal(t, []ShardID{ShardID("0")}, decision.TargetShardIDs)
	require.NoError(t, decision.ValidateHash())

	midBlock := decision
	midBlock.TargetLayoutEpoch = midBlock.SourceLayoutEpoch
	midBlock.DecisionHash = ComputeShardRebalanceDecisionHash(midBlock)
	require.ErrorContains(t, midBlock.ValidateHash(), "future")
}

func TestShardStateKeysRootsAndLocksAreCanonical(t *testing.T) {
	metaKey, err := ShardMetaKey(ZoneIDContract, "2")
	require.NoError(t, err)
	require.Equal(t, "zones/CONTRACT_ZONE/shards/2/meta", metaKey)

	metricsKey, err := ShardMetricsKey(ZoneIDContract, "2", 99)
	require.NoError(t, err)
	require.Equal(t, "zones/CONTRACT_ZONE/shards/2/metrics/00000000000000000099", metricsKey)

	inboxKey, err := ShardInboxKey(ZoneIDContract, "2", "msg-7")
	require.NoError(t, err)
	require.Equal(t, "zones/CONTRACT_ZONE/shards/2/inbox/msg-7", inboxKey)

	outboxKey, err := ShardOutboxKey(ZoneIDContract, "2", "msg-7")
	require.NoError(t, err)
	require.Equal(t, "zones/CONTRACT_ZONE/shards/2/outbox/msg-7", outboxKey)

	receiptKey, err := ShardReceiptKey(ZoneIDContract, "2", "msg-7")
	require.NoError(t, err)
	require.Equal(t, "zones/CONTRACT_ZONE/shards/2/receipts/msg-7", receiptKey)

	rootKey, err := ShardRootKey(ZoneIDContract, "2", 99)
	require.NoError(t, err)
	require.Equal(t, "zones/CONTRACT_ZONE/shards/2/root/00000000000000000099", rootKey)

	task, err := NewShardMigrationTask(ShardMigrationTask{
		ZoneID:			ZoneIDContract,
		SourceShardID:		"2",
		DestinationShardID:	"3",
		SourceLayoutEpoch:	9,
		TargetLayoutEpoch:	10,
		KeyPrefix:		"contract/storage",
		DeliveryEpoch:		10,
	})
	require.NoError(t, err)
	lock, err := NewObjectLock(ObjectLock{
		ZoneID:		ZoneIDContract,
		ShardID:	"2",
		ObjectID:	"contract/storage/a",
		Reason:		"migration",
		OwnerTaskID:	task.TaskID,
		CreatedHeight:	100,
		ExpiryHeight:	120,
	})
	require.NoError(t, err)
	lockKey, err := ShardLockKey(ZoneIDContract, "2", lock.ObjectID)
	require.NoError(t, err)
	require.Equal(t, "zones/CONTRACT_ZONE/shards/2/locks/contract/storage/a", lockKey)
	require.NoError(t, lock.ValidateHash())

	rootA := shardRoot(t, ZoneIDContract, "2", 99)
	rootB := shardRoot(t, ZoneIDContract, "1", 99)
	rootAB, err := ComputeShardRootsRoot([]ShardRoot{rootA, rootB})
	require.NoError(t, err)
	rootBA, err := ComputeShardRootsRoot([]ShardRoot{rootB, rootA})
	require.NoError(t, err)
	require.Equal(t, rootAB, rootBA)
}

func shardRebalanceThresholds() ShardRebalanceThresholds {
	return ShardRebalanceThresholds{
		GasLimitPerShard:		1_000,
		SplitGasUtilization:		90,
		SplitStateSizeBytes:		1_000,
		SplitWriteConflictCount:	10,
		SplitQueueBacklog:		20,
		SplitProofLatencyMicros:	1_000,
		MergeGasUtilization:		10,
		MergeStateSizeBytes:		100,
		MergeQueueBacklog:		2,
		MergeWriteConflictCount:	1,
		DecisionWindow:			3,
		FutureLayoutEpochDelta:		2,
		FutureActivationHeightGap:	5,
	}
}

func shardMetric(t *testing.T, zoneID ZoneID, shardID ShardID, height uint64, gas uint64, inbox uint64, outbox uint64, size uint64, proofLatency uint64) ShardMetrics {
	t.Helper()
	metric, err := NewShardMetrics(ShardMetrics{
		ZoneID:			zoneID,
		ShardID:		shardID,
		Height:			height,
		GasUsed:		gas,
		InboxBacklog:		inbox,
		OutboxBacklog:		outbox,
		StateSizeBytes:		size,
		ProofLatencyMicros:	proofLatency,
	})
	require.NoError(t, err)
	return metric
}

func shardRoot(t *testing.T, zoneID ZoneID, shardID ShardID, height uint64) ShardRoot {
	t.Helper()
	root, err := NewShardRoot(ShardRoot{
		ZoneID:		zoneID,
		ShardID:	shardID,
		Height:		height,
		StateRoot:	testHash(string(zoneID) + "/" + string(shardID) + "/state"),
		InboxRoot:	testHash(string(zoneID) + "/" + string(shardID) + "/inbox"),
		OutboxRoot:	testHash(string(zoneID) + "/" + string(shardID) + "/outbox"),
		ReceiptsRoot:	testHash(string(zoneID) + "/" + string(shardID) + "/receipts"),
		MetricsHash:	testHash(string(zoneID) + "/" + string(shardID) + "/metrics"),
	})
	require.NoError(t, err)
	return root
}
