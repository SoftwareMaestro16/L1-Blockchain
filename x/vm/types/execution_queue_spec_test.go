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
		ZoneID:        zonestypes.ZoneIDContract,
		PriorityQueue: []AVMZoneQueueEntry{low, high},
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
		ZoneID:        zonestypes.ZoneIDContract,
		PriorityQueue: entries,
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
		ZoneID:        zonestypes.ZoneIDContract,
		PriorityQueue: []AVMZoneQueueEntry{badSort},
	})
	require.ErrorContains(t, err, "sort key")

	failed, err := NewAVMZoneQueueEntry(AVMQueueLaneFailed, msg, 0)
	require.NoError(t, err)
	queue, err := NewAVMZoneQueue(AVMZoneQueue{
		ZoneID:      zonestypes.ZoneIDContract,
		FailedQueue: []AVMZoneQueueEntry{failed},
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

func testAVMQueueMessage(source string, nonce, createdHeight uint64, priority uint8, delayHeight, expiryHeight, gasLimit uint64) AVMAsyncMessage {
	msg := testAVMAsyncMessage(source, zonestypes.ZoneIDApplication, "contract", zonestypes.ZoneIDContract, nonce, createdHeight)
	msg.Priority = priority
	msg.DelayHeight = delayHeight
	msg.ExpiryHeight = expiryHeight
	msg.RetryPolicy = DefaultAVMRetryPolicy(expiryHeight)
	msg.GasLimit = gasLimit
	return msg
}
