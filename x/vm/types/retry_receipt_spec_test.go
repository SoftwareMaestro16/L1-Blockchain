package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMRetryPolicyModesAndDeterministicRetryHeights(t *testing.T) {
	none := AVMRetryPolicy{Mode: AVMRetryModeNone, BackoffMode: AVMBackoffModeNone}
	require.NoError(t, none.ValidateForMessage(20))

	fixed := AVMRetryPolicy{
		Mode:		AVMRetryModeFixed,
		MaxAttempts:	3,
		RetryDelay:	2,
		BackoffMode:	AVMBackoffModeNone,
		MaxRetryHeight:	20,
		ChargeRetryGas:	true,
	}
	require.NoError(t, fixed.ValidateForMessage(20))
	next, err := NextAVMRetryHeight(10, 2, fixed)
	require.NoError(t, err)
	require.Equal(t, uint64(12), next)

	backoff := AVMRetryPolicy{
		Mode:		AVMRetryModeBoundedBackoff,
		MaxAttempts:	4,
		RetryDelay:	2,
		BackoffMode:	AVMBackoffModeExponential,
		MaxRetryHeight:	30,
		ChargeRetryGas:	true,
	}
	next, err = NextAVMRetryHeight(10, 3, backoff)
	require.NoError(t, err)
	require.Equal(t, uint64(18), next)
}

func TestAVMRetryPolicyRejectsUnboundedExpiryAndGasDrift(t *testing.T) {
	require.ErrorContains(t, AVMRetryPolicy{
		Mode:		AVMRetryModeFixed,
		RetryDelay:	1,
		BackoffMode:	AVMBackoffModeNone,
		MaxRetryHeight:	20,
		ChargeRetryGas:	true,
	}.ValidateForMessage(20), "bounded")

	require.ErrorContains(t, AVMRetryPolicy{
		Mode:		AVMRetryModeFixed,
		MaxAttempts:	1,
		RetryDelay:	0,
		BackoffMode:	AVMBackoffModeNone,
		MaxRetryHeight:	20,
		ChargeRetryGas:	true,
	}.ValidateForMessage(20), "deterministic")

	require.ErrorContains(t, AVMRetryPolicy{
		Mode:		AVMRetryModeBoundedBackoff,
		MaxAttempts:	1,
		RetryDelay:	1,
		BackoffMode:	AVMBackoffModeLinear,
		MaxRetryHeight:	21,
		ChargeRetryGas:	true,
	}.ValidateForMessage(20), "expiry")

	require.ErrorContains(t, AVMRetryPolicy{
		Mode:		AVMRetryModeFixed,
		MaxAttempts:	1,
		RetryDelay:	1,
		BackoffMode:	AVMBackoffModeNone,
		MaxRetryHeight:	20,
		ChargeRetryGas:	false,
	}.ValidateForMessage(20), "retry gas")
}

func TestAVMExecutionReceiptHashesSection54Fields(t *testing.T) {
	msg, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 1, 10))
	require.NoError(t, err)
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		msg.ID,
		ZoneID:			zonestypes.ZoneIDContract,
		Executor:		"contract-executor",
		Status:			AVMReceiptStatusExecuted,
		GasUsed:		40,
		StorageWritten:		2,
		EventsHash:		engineHash("events"),
		OutputMessagesRoot:	engineHash("out"),
		CreatedHeight:		12,
	})
	require.NoError(t, err)
	require.NoError(t, receipt.Validate())
	require.Equal(t, DeriveAVMReceiptID(receipt), receipt.ReceiptID)
	require.Equal(t, ComputeAVMExecutionReceiptHash(receipt), receipt.ReceiptHash)

	mutated := receipt
	mutated.StorageWritten++
	require.NotEqual(t, receipt.ReceiptHash, ComputeAVMExecutionReceiptHash(mutated))
}

func TestAVMExecutionReceiptRejectsInvalidStatusHashAndNonTerminalWrites(t *testing.T) {
	msg, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 2, 10))
	require.NoError(t, err)
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		msg.ID,
		ZoneID:			zonestypes.ZoneIDContract,
		Executor:		"contract-executor",
		Status:			AVMReceiptStatusFailed,
		GasUsed:		1,
		EventsHash:		engineHash("events"),
		OutputMessagesRoot:	engineHash("out"),
		ErrorCodeOptional:	"handler_failed",
		CreatedHeight:		12,
	})
	require.NoError(t, err)

	badStatus := receipt
	badStatus.Status = AVMReceiptStatus("bad")
	badStatus.ReceiptHash = ComputeAVMExecutionReceiptHash(badStatus)
	require.ErrorContains(t, badStatus.Validate(), "status")

	badHash := receipt
	badHash.ReceiptHash = engineHash("wrong")
	require.ErrorContains(t, badHash.Validate(), "hash mismatch")

	submitted, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		msg.ID,
		ZoneID:			zonestypes.ZoneIDContract,
		Executor:		"scheduler",
		Status:			AVMReceiptStatusSubmitted,
		EventsHash:		engineHash("events"),
		OutputMessagesRoot:	engineHash("out"),
		CreatedHeight:		12,
	})
	require.NoError(t, err)
	submitted.StorageWritten = 1
	submitted.ReceiptHash = ComputeAVMExecutionReceiptHash(submitted)
	require.ErrorContains(t, submitted.Validate(), "must not write storage")
}
