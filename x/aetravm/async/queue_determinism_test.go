package async

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDeterministicQueueOrderingStableRegardlessInsertionOrder(t *testing.T) {
	destA := testAddr(0x0a)
	destB := testAddr(0x0b)
	source := testAddr(0x01)
	records := []QueuedMessage{
		buildQueuedMessage(testQueueMessage(source, destB, 3, 5, 7), 10, 2, 0, 3),
		buildQueuedMessage(testQueueMessage(source, destA, 1, 0, 5), 10, 1, 1, 1),
		buildQueuedMessage(testQueueMessage(source, destB, 2, 0, 5), 10, 1, 0, 2),
		buildQueuedMessage(testQueueMessage(source, destA, 4, 5, 7), 10, 2, 0, 4),
	}
	for i := range records {
		require.NoError(t, validateQueuedMessage(records[i], DefaultParams()))
	}

	forward := append([]QueuedMessage(nil), records...)
	reverse := []QueuedMessage{records[3], records[2], records[1], records[0]}
	sort.SliceStable(forward, func(i, j int) bool { return queuedMessageLess(forward[i], forward[j]) })
	sort.SliceStable(reverse, func(i, j int) bool { return queuedMessageLess(reverse[i], reverse[j]) })

	require.Equal(t, queueQueryIDs(forward), queueQueryIDs(reverse))
	require.Equal(t, []uint64{2, 1, 3, 4}, queueQueryIDs(forward))
}

func TestQueuePerBlockAndPerContractLimits(t *testing.T) {
	t.Run("per block", func(t *testing.T) {
		params := DefaultParams()
		params.MaxMessagesPerBlock = 1
		executor, err := NewExecutor(params)
		require.NoError(t, err)
		dest := deployTestContract(t, executor, testAddr(1), []byte("block-limit"))
		require.NoError(t, executor.RegisterHandler(dest, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
			return ExecutionResult{NewState: contract.State, ResultCode: ResultOK}
		}))
		require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{
			testMessage(testAddr(9), dest, 1),
			testMessage(testAddr(9), dest, 2),
		}))

		receipts, err := executor.ProcessBlock(1)
		require.NoError(t, err)
		require.Len(t, receipts, 1)
		require.Equal(t, QueueStatusExecuted, receipts[0].QueueStatus)
		require.Len(t, executor.Queue(), 1)
	})

	t.Run("per contract", func(t *testing.T) {
		params := DefaultParams()
		params.MaxQueuedMessagesPerContract = 1
		executor, err := NewExecutor(params)
		require.NoError(t, err)
		dest := deployTestContract(t, executor, testAddr(1), []byte("contract-limit"))
		require.NoError(t, executor.EnqueueMessage(testMessage(testAddr(9), dest, 1)))
		require.ErrorContains(t, executor.EnqueueMessage(testMessage(testAddr(9), dest, 2)), "queued messages per contract")
	})
}

func TestExpiredMessageMarkedDeterministically(t *testing.T) {
	executor := newTestExecutor(t)
	dest := deployTestContract(t, executor, testAddr(1), []byte("expired-queue"))
	msg := testMessage(testAddr(9), dest, 1)
	msg.Bounce = false
	msg.Value = naetCoin(0)
	msg.DeadlineBlock = 1
	require.NoError(t, executor.EnqueueMessage(msg))

	receipts, err := executor.ProcessBlock(2)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultExpired, receipts[0].ResultCode)
	require.Equal(t, QueueStatusExpired, receipts[0].QueueStatus)
	require.Empty(t, executor.Queue())
}

func TestQueueExportImportPreservesDeterministicOrderAndRecordFields(t *testing.T) {
	executor := newTestExecutor(t)
	dest := deployTestContract(t, executor, testAddr(1), []byte("queue-export"))
	delayed := testMessage(testAddr(9), dest, 1)
	delayed.DeliverAtBlock = 5
	immediate := testMessage(testAddr(9), dest, 2)
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{delayed, immediate}))

	exported := executor.ExportState()
	require.Len(t, exported.Queue, 2)
	require.Equal(t, uint64(0), exported.Queue[0].ScheduledHeight)
	require.Equal(t, uint64(5), exported.Queue[1].ScheduledHeight)
	require.Equal(t, QueueStatusPending, exported.Queue[0].Status)
	require.Len(t, exported.Queue[0].MessageID, MessageIDLength)
	require.Equal(t, QueueMessageID(exported.Queue[0]), exported.Queue[0].MessageID)
	require.Equal(t, queueAddressKey(dest), exported.Queue[0].DestinationKey)

	imported, err := ImportState(exported)
	require.NoError(t, err)
	require.Equal(t, exported.Queue, imported.Queue())
	require.Equal(t, exported.ExportQueueOrder(), imported.ExportState().ExportQueueOrder())
}

func TestQueueRejectsAttemptsDepthBodyAndMessageExplosion(t *testing.T) {
	t.Run("attempts", func(t *testing.T) {
		params := DefaultParams()
		params.MaxProcessingAttempts = 2
		msg := testMessage(testAddr(1), testAddr(2), 1)
		msg.RetryCount = 2
		msg.MaxRetries = 2
		require.ErrorContains(t, msg.Validate(params), "processing attempts")
	})

	t.Run("depth", func(t *testing.T) {
		params := DefaultParams()
		msg := testMessage(testAddr(1), testAddr(2), 1)
		msg.Depth = params.MaxRecursionDepth + 1
		require.ErrorContains(t, msg.Validate(params), "message depth")
	})

	t.Run("body", func(t *testing.T) {
		params := DefaultParams()
		params.MaxBodySize = 3
		msg := testMessage(testAddr(1), testAddr(2), 1)
		msg.Body = []byte("large")
		require.ErrorContains(t, msg.Validate(params), "message body size")
	})

	t.Run("emitted messages", func(t *testing.T) {
		params := DefaultParams()
		params.MaxEmittedMessagesPerExec = 1
		executor, err := NewExecutor(params)
		require.NoError(t, err)
		source := deployTestContract(t, executor, testAddr(1), []byte("explosion-src"))
		dest := deployTestContract(t, executor, testAddr(1), []byte("explosion-dst"))
		require.NoError(t, executor.RegisterHandler(source, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
			return ExecutionResult{
				NewState:	[]byte("must-not-commit"),
				ResultCode:	ResultOK,
				Outgoing: []MessageEnvelope{
					testMessage(contract.Address, dest, 10),
					testMessage(contract.Address, dest, 11),
				},
			}
		}))
		msg := testMessage(testAddr(9), source, 1)
		msg.Bounce = false
		msg.Value = naetCoin(0)
		require.NoError(t, executor.EnqueueMessage(msg))
		receipts, err := executor.ProcessBlock(2)
		require.NoError(t, err)
		require.Len(t, receipts, 1)
		require.Equal(t, ResultLimitExceeded, receipts[0].ResultCode)
		require.Equal(t, QueueStatusFailed, receipts[0].QueueStatus)
		require.Empty(t, executor.Queue())
		contract, ok := executor.Contract(source)
		require.True(t, ok)
		require.NotEqual(t, []byte("must-not-commit"), contract.State)
	})
}

func queueQueryIDs(messages []QueuedMessage) []uint64 {
	out := make([]uint64, len(messages))
	for i, msg := range messages {
		out[i] = msg.Envelope.QueryID
	}
	return out
}

func testQueueMessage(source, dest sdk.AccAddress, queryID, scheduledHeight, logicalTime uint64) MessageEnvelope {
	msg := testMessage(source, dest, queryID)
	msg.DeliverAtBlock = scheduledHeight
	msg.CreatedLogicalTime = logicalTime
	return msg
}

func (e ExportedState) ExportQueueOrder() []uint64 {
	return queueQueryIDs(e.Queue)
}
