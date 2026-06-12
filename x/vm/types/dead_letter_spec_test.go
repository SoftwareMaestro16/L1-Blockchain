package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMDeadLetterRecordIsProofQueryableTerminalRecord(t *testing.T) {
	msg, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 3, 10))
	require.NoError(t, err)
	receipt := testAVMDeadLetterReceipt(t, msg, 14)

	record, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:		msg.ID,
		ZoneID:			msg.DestinationZone,
		Reason:			"retry budget exhausted",
		FailedAttempts:		3,
		LastErrorCode:		"handler_failed",
		FinalHeight:		receipt.CreatedHeight,
		RefundAmountOptional:	7,
		ReceiptID:		receipt.ReceiptID,
	})
	require.NoError(t, err)
	require.NoError(t, record.Validate())
	require.NoError(t, record.ValidateWithReceipt(receipt))
	require.Equal(t, AVMAsyncDeadLetterKey(msg.DestinationZone, msg.ID), record.ProofKey())
	require.Equal(t, record.ProofKey(), AVMDeadLetterProofKey(msg.DestinationZone, msg.ID))
	require.Equal(t, ComputeAVMDeadLetterRecordHash(record), record.RecordHash)
	require.True(t, record.CanTriggerBounce(msg))

	mutated := record
	mutated.FailedAttempts++
	require.NotEqual(t, record.RecordHash, ComputeAVMDeadLetterRecordHash(mutated))
}

func TestAVMDeadLetterRecordRejectsNonTerminalOrMismatchedReceipt(t *testing.T) {
	msg, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 4, 10))
	require.NoError(t, err)
	receipt := testAVMDeadLetterReceipt(t, msg, 14)

	record, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		Reason:		"expired retry window",
		FailedAttempts:	2,
		LastErrorCode:	"expired",
		FinalHeight:	receipt.CreatedHeight,
		ReceiptID:	receipt.ReceiptID,
	})
	require.NoError(t, err)

	executed := receipt
	executed.Status = AVMReceiptStatusExecuted
	executed.ReceiptHash = ComputeAVMExecutionReceiptHash(executed)
	require.ErrorContains(t, record.ValidateWithReceipt(executed), "dead_lettered")

	wrongHeight := record
	wrongHeight.FinalHeight++
	wrongHeight.RecordHash = ComputeAVMDeadLetterRecordHash(wrongHeight)
	require.ErrorContains(t, wrongHeight.ValidateWithReceipt(receipt), "final height")

	wrongReceiptID := receipt
	wrongReceiptID.ReceiptID = engineHash("wrong-receipt")
	wrongReceiptID.ReceiptHash = ComputeAVMExecutionReceiptHash(wrongReceiptID)
	require.ErrorContains(t, record.ValidateWithReceipt(wrongReceiptID), "receipt id")
}

func TestAVMDeadLetterRecordRejectsIncompleteFieldsAndDisabledBounce(t *testing.T) {
	msg, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 5, 10))
	require.NoError(t, err)
	receipt := testAVMDeadLetterReceipt(t, msg, 14)

	_, err = NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		Reason:		"",
		FailedAttempts:	1,
		LastErrorCode:	"failed",
		FinalHeight:	receipt.CreatedHeight,
		ReceiptID:	receipt.ReceiptID,
	})
	require.ErrorContains(t, err, "reason")

	_, err = NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		Reason:		"failed",
		FailedAttempts:	0,
		LastErrorCode:	"failed",
		FinalHeight:	receipt.CreatedHeight,
		ReceiptID:	receipt.ReceiptID,
	})
	require.ErrorContains(t, err, "failed attempts")

	record, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		Reason:		"failed",
		FailedAttempts:	1,
		LastErrorCode:	"failed",
		FinalHeight:	receipt.CreatedHeight,
		ReceiptID:	receipt.ReceiptID,
	})
	require.NoError(t, err)

	disabledBounce := msg
	disabledBounce.BounceFlag = false
	require.False(t, record.CanTriggerBounce(disabledBounce))
}

func testAVMDeadLetterReceipt(t *testing.T, msg AVMAsyncMessage, height uint64) AVMExecutionReceipt {
	t.Helper()
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		msg.ID,
		ZoneID:			msg.DestinationZone,
		Executor:		"async-engine",
		Status:			AVMReceiptStatusDeadLettered,
		GasUsed:		9,
		EventsHash:		engineHash("dead-letter-events"),
		OutputMessagesRoot:	engineHash("dead-letter-output"),
		ErrorCodeOptional:	"handler_failed",
		CreatedHeight:		height,
	})
	require.NoError(t, err)
	return receipt
}
