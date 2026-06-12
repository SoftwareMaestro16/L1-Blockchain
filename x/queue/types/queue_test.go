package types

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestQueueOrderingAndProcessingLimit(t *testing.T) {
	params := DefaultParams()
	params.MaxPerBlock = 2
	q, err := NewQueue(params)
	require.NoError(t, err)

	low, err := q.Enqueue(testItem(10, ClassRestricted, 1, 0, 0, 2))
	require.NoError(t, err)
	high, err := q.Enqueue(testItem(10, ClassElite, 1, 1, 0, 1))
	require.NoError(t, err)
	earlier, err := q.Enqueue(testItem(9, ClassRestricted, 2, 0, 0, 3))
	require.NoError(t, err)

	ready := q.PopReady(10)
	require.Len(t, ready, 2)
	require.Equal(t, earlier.Sequence, ready[0].Sequence)
	require.Equal(t, high.Sequence, ready[1].Sequence)
	require.Len(t, q.Items(), 1)
	require.Equal(t, low.Sequence, q.Items()[0].Sequence)
	require.Equal(t, uint64(2), q.Metrics(10).Processed)
}

func TestQueueLimitsAndRetryFailureMetrics(t *testing.T) {
	params := DefaultParams()
	params.MaxPerAccountQueued = 1
	params.MaxPerContractQueued = 2
	q, err := NewQueue(params)
	require.NoError(t, err)
	item := testItem(1, ClassNormal, 1, 0, 0, 1)
	_, err = q.Enqueue(item)
	require.NoError(t, err)
	_, err = q.Enqueue(item)
	require.ErrorContains(t, err, "account queued")

	ready := q.PopReady(1)
	require.Len(t, ready, 1)
	retry, err := q.Retry(ready[0], 5, "temporary")
	require.NoError(t, err)
	require.Equal(t, uint32(1), retry.Attempts)
	require.Equal(t, "temporary", retry.LastError)
	failed := q.Fail(retry, "permanent")
	require.Equal(t, uint32(2), failed.Attempts)
	require.Equal(t, uint64(1), q.Metrics(5).Failed)
}

func TestStarvationProtectionAndReputationClass(t *testing.T) {
	params := DefaultParams()
	params.StarvationWindowHeights = 3
	item := testItem(10, ClassRestricted, 1, 0, 0, 1)
	require.Equal(t, ClassRestricted, EffectiveReputationClass(item, 12, params))
	require.Equal(t, ClassElite, EffectiveReputationClass(item, 13, params))
	require.Equal(t, ClassElite, ReputationClassForScore(95))
	require.Equal(t, ClassRestricted, ReputationClassForScore(1))
}

func TestQueueValidation(t *testing.T) {
	params := DefaultParams()
	q, err := NewQueue(params)
	require.NoError(t, err)
	item := testItem(1, ClassRestricted+1, 1, 0, 0, 1)
	_, err = q.Enqueue(item)
	require.ErrorContains(t, err, "reputation class")
	item = testItem(1, ClassNormal, 1, 0, 0, 1)
	item.Account = make([]byte, 20)
	_, err = q.Enqueue(item)
	require.ErrorContains(t, err, "queue account")
}

func testItem(scheduled uint64, class uint8, txHeight uint64, txIndex uint32, msgIndex uint32, sourceLT uint64) QueueItem {
	return QueueItem{
		ScheduledHeight:	scheduled,
		ReputationClass:	class,
		TxHeight:		txHeight,
		TxIndex:		txIndex,
		MessageIndex:		msgIndex,
		SourceLogicalTime:	sourceLT,
		Account:		addr(1),
		Contract:		addr(2),
		Payload:		[]byte("payload"),
	}
}

func addr(seed byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20))
}
