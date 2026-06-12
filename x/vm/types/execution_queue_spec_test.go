package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMZoneQueueKeysAndCanonicalSort(t *testing.T) {
	msgLow, err := NewAVMAsyncMessage(testAVMQueueMessage("alice", 1, 10, 4, 0, 100, 10))
	require.NoError(t, err)
	msgHigh, err := NewAVMAsyncMessage(testAVMQueueMessage("alice", 2, 10, 9, 0, 100, 10))
	require.NoError(t, err)

	low, err := NewAVMZoneQueueEntry(AVMQueueLanePriority, msgLow, 0)
	require.NoError(t, err)
	high, err := NewAVMZoneQueueEntry(AVMQueueLanePriority, msgHigh, 0)
	require.NoError(t, err)
	delayed, err := NewAVMZoneQueueEntry(AVMQueueLaneDelayed, msgLow, 25)
	require.NoError(t, err)
	retry, err := NewAVMZoneQueueEntry(AVMQueueLaneRetry, msgHigh, 30)
	require.NoError(t, err)
	failed, err := NewAVMZoneQueueEntry(AVMQueueLaneFailed, msgHigh, 0)
	require.NoError(t, err)

	require.Equal(t, AVMQueuePriorityKey(msgLow.DestinationZone, low.SortKey), low.StateKey())
	require.Equal(t, AVMQueueDelayedKey(msgLow.DestinationZone, 25, delayed.SortKey), delayed.StateKey())
	require.Equal(t, AVMQueueRetryKey(msgHigh.DestinationZone, 30, retry.SortKey), retry.StateKey())
	require.Equal(t, AVMQueueFailedKey(msgHigh.DestinationZone, failed.SortKey), failed.StateKey())
	require.Contains(t, low.SortKey, msgLow.ID)
	require.Less(t, high.SortKey, low.SortKey)

	queue, err := NewAVMZoneQueue(AVMZoneQueue{
		ZoneID:		zonestypes.ZoneIDContract,
		PriorityQueue:	[]AVMZoneQueueEntry{low, high},
	})
	require.NoError(t, err)
	require.Equal(t, msgHigh.ID, queue.PriorityQueue[0].MessageID)
	require.Equal(t, ComputeAVMZoneQueueRoot(queue), queue.QueueRoot)

	mutated := queue
	mutated.PriorityQueue[0].Priority--
	require.NotEqual(t, queue.QueueRoot, ComputeAVMZoneQueueRoot(mutated))
}

func TestAVMZoneQueueSelectsBySchedulingRuleAndBudget(t *testing.T) {
	nonceTwo, err := NewAVMAsyncMessage(testAVMQueueMessage("same-sender", 2, 10, 5, 0, 100, 30))
	require.NoError(t, err)
	nonceOne, err := NewAVMAsyncMessage(testAVMQueueMessage("same-sender", 1, 10, 5, 0, 100, 30))
	require.NoError(t, err)
	expired, err := NewAVMAsyncMessage(testAVMQueueMessage("expired", 1, 10, 9, 0, 12, 30))
	require.NoError(t, err)
	delayed, err := NewAVMAsyncMessage(testAVMQueueMessage("delayed", 1, 10, 10, 10, 100, 30))
	require.NoError(t, err)
	overBudget, err := NewAVMAsyncMessage(testAVMQueueMessage("over-budget", 1, 10, 4, 0, 100, 80))
	require.NoError(t, err)

	entries := make([]AVMZoneQueueEntry, 0, 5)
	for _, msg := range []AVMAsyncMessage{nonceTwo, nonceOne, expired, delayed, overBudget} {
		entry, err := NewAVMZoneQueueEntry(AVMQueueLanePriority, msg, 0)
		require.NoError(t, err)
		entries = append(entries, entry)
	}
	queue, err := NewAVMZoneQueue(AVMZoneQueue{
		ZoneID:		zonestypes.ZoneIDContract,
		PriorityQueue:	entries,
	})
	require.NoError(t, err)

	selection, err := SelectAVMZoneQueueWork(
		queue,
		[]AVMAsyncMessage{nonceTwo, nonceOne, expired, delayed, overBudget},
		15,
		zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 3},
	)
	require.NoError(t, err)
	require.Len(t, selection.Expired, 1)
	require.Equal(t, expired.ID, selection.Expired[0].ID)
	require.Len(t, selection.Ready, 2)
	require.Equal(t, nonceOne.ID, selection.Ready[0].ID)
	require.Equal(t, nonceTwo.ID, selection.Ready[1].ID)
	require.Equal(t, uint64(60), selection.Budget.GasUsed)
	require.Equal(t, uint32(3), selection.Budget.MessagesUsed)
	require.Len(t, selection.Remaining.PriorityQueue, 2)
	require.Contains(t, []string{selection.Remaining.PriorityQueue[0].MessageID, selection.Remaining.PriorityQueue[1].MessageID}, delayed.ID)
	require.Contains(t, []string{selection.Remaining.PriorityQueue[0].MessageID, selection.Remaining.PriorityQueue[1].MessageID}, overBudget.ID)
}

func TestAVMZoneQueueRejectsInvalidEntriesAndIgnoresFailedLaneForExecution(t *testing.T) {
	msg, err := NewAVMAsyncMessage(testAVMQueueMessage("alice", 1, 10, 5, 0, 100, 10))
	require.NoError(t, err)
	entry, err := NewAVMZoneQueueEntry(AVMQueueLanePriority, msg, 0)
	require.NoError(t, err)

	badSort := entry
	badSort.SortKey = "wrong"
	_, err = NewAVMZoneQueue(AVMZoneQueue{
		ZoneID:		zonestypes.ZoneIDContract,
		PriorityQueue:	[]AVMZoneQueueEntry{badSort},
	})
	require.ErrorContains(t, err, "sort key")

	failed, err := NewAVMZoneQueueEntry(AVMQueueLaneFailed, msg, 0)
	require.NoError(t, err)
	queue, err := NewAVMZoneQueue(AVMZoneQueue{
		ZoneID:		zonestypes.ZoneIDContract,
		FailedQueue:	[]AVMZoneQueueEntry{failed},
	})
	require.NoError(t, err)
	selection, err := SelectAVMZoneQueueWork(
		queue,
		[]AVMAsyncMessage{msg},
		15,
		zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
	)
	require.NoError(t, err)
	require.Empty(t, selection.Ready)
	require.Empty(t, selection.Expired)
	require.Len(t, selection.Remaining.FailedQueue, 1)
}

func TestAVMZoneQueueAdmissionValidatesDepthDuplicateAndDelayLane(t *testing.T) {
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	delayedMsg, err := NewAVMAsyncMessage(testAVMQueueMessage("alice", 1, 10, 3, 5, 100, 10))
	require.NoError(t, err)

	queue, entry, err := AdmitAVMZoneQueueMessage(queue, delayedMsg, 12, 2)
	require.NoError(t, err)
	require.Equal(t, AVMQueueLaneDelayed, entry.Lane)
	require.Empty(t, queue.PriorityQueue)
	require.Len(t, queue.DelayedQueue, 1)

	_, _, err = AdmitAVMZoneQueueMessage(queue, delayedMsg, 12, 2)
	require.ErrorContains(t, err, "duplicate")

	readyMsg, err := NewAVMAsyncMessage(testAVMQueueMessage("bob", 1, 10, 3, 0, 100, 10))
	require.NoError(t, err)
	queue, entry, err = AdmitAVMZoneQueueMessage(queue, readyMsg, 12, 2)
	require.NoError(t, err)
	require.Equal(t, AVMQueueLanePriority, entry.Lane)
	require.Len(t, queue.PriorityQueue, 1)

	third, err := NewAVMAsyncMessage(testAVMQueueMessage("carol", 1, 10, 3, 0, 100, 10))
	require.NoError(t, err)
	_, _, err = AdmitAVMZoneQueueMessage(queue, third, 12, 2)
	require.ErrorContains(t, err, "max queue depth")
}

func TestAVMZoneQueuePromotesDelayedAndRetryQueuesBounded(t *testing.T) {
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	delayedDue, err := NewAVMAsyncMessage(testAVMQueueMessage("delayed-due", 1, 10, 4, 5, 100, 10))
	require.NoError(t, err)
	delayedFuture, err := NewAVMAsyncMessage(testAVMQueueMessage("delayed-future", 1, 10, 4, 20, 100, 10))
	require.NoError(t, err)
	retryDue, err := NewAVMAsyncMessage(testAVMQueueMessage("retry-due", 1, 10, 5, 0, 100, 10))
	require.NoError(t, err)

	queue, _, err = AdmitAVMZoneQueueMessage(queue, delayedDue, 12, 10)
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, delayedFuture, 12, 10)
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneRetryMessage(queue, retryDue, 15, 10)
	require.NoError(t, err)

	next, promoted, err := PromoteAVMZoneQueue(queue, 15, 1)
	require.NoError(t, err)
	require.Len(t, promoted, 1)
	require.Equal(t, retryDue.ID, promoted[0].MessageID)
	require.Equal(t, AVMQueueLanePriority, promoted[0].Lane)
	require.Len(t, next.PriorityQueue, 1)
	require.Len(t, next.DelayedQueue, 2)
	require.Empty(t, next.RetryQueue)

	next, promoted, err = PromoteAVMZoneQueue(next, 15, 10)
	require.NoError(t, err)
	require.Len(t, promoted, 1)
	require.Len(t, next.PriorityQueue, 2)
	require.Len(t, next.DelayedQueue, 1)
	require.Empty(t, next.RetryQueue)
}

func TestAVMZoneQueueDeadLettersToFailedQueueAndProvidesProof(t *testing.T) {
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	msg, err := NewAVMAsyncMessage(testAVMQueueMessage("alice", 1, 10, 5, 0, 100, 10))
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, msg, 12, 10)
	require.NoError(t, err)
	receipt := testAVMDeadLetterReceipt(t, msg, 14)

	next, record, err := DeadLetterAVMZoneQueueMessage(queue, msg, receipt, "handler failed permanently", 3, 1)
	require.NoError(t, err)
	require.NoError(t, record.ValidateWithReceipt(receipt))
	require.Empty(t, next.PriorityQueue)
	require.Len(t, next.FailedQueue, 1)
	require.Equal(t, msg.ID, next.FailedQueue[0].MessageID)

	proof, err := QueryAVMZoneQueueProof(next, AVMQueueLaneFailed, msg.ID)
	require.NoError(t, err)
	require.NoError(t, proof.Validate())
	require.Equal(t, next.QueueRoot, proof.QueueRoot)
	require.Equal(t, next.FailedQueue[0].StateKey(), proof.StateKey)
	require.Equal(t, AVMDeadLetterProofKey(msg.DestinationZone, msg.ID), record.ProofKey())
}

func testAVMQueueMessage(source string, nonce, createdHeight uint64, priority uint8, delayHeight, expiryHeight, gasLimit uint64) AVMAsyncMessage {
	msg := testAVMAsyncMessage(source, zonestypes.ZoneIDApplication, "contract", zonestypes.ZoneIDContract, nonce, createdHeight)
	msg.Priority = priority
	msg.DelayHeight = delayHeight
	msg.ExpiryHeight = expiryHeight
	msg.RetryPolicy = DefaultAVMRetryPolicy(expiryHeight)
	msg.GasLimit = gasLimit
	return msg
}
