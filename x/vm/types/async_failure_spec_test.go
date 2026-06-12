package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMAsyncFailureRecordCoversSection12FailureClasses(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	classes := []AVMAsyncFailureClass{
		AVMAsyncFailureInvalidPayload,
		AVMAsyncFailureInsufficientGas,
		AVMAsyncFailureDestinationNotFound,
		AVMAsyncFailureDestinationDisabled,
		AVMAsyncFailureExpiredMessage,
		AVMAsyncFailureHandlerFailure,
		AVMAsyncFailureStorageLimitExceeded,
		AVMAsyncFailureProofVerificationFailure,
		AVMAsyncFailureRetryExhausted,
	}

	for i, class := range classes {
		record, err := NewAVMAsyncFailureRecord(AVMAsyncFailureRecord{
			MessageID:	msg.ID,
			ZoneID:		msg.DestinationZone,
			FailureClass:	class,
			FailedHeight:	14,
			Attempt:	uint32(i + 1),
			GasUsed:	uint64(i),
			RetryExhausted:	class == AVMAsyncFailureRetryExhausted,
		})
		require.NoError(t, err)
		require.NoError(t, record.Validate())
		code, ok := AVMAsyncFailureErrorCode(class)
		require.True(t, ok)
		require.Equal(t, string(code), record.ErrorCode)
		require.Equal(t, ComputeAVMAsyncFailureHash(record), record.FailureHash)
	}

	badClass := AVMAsyncFailureRecord{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		FailureClass:	"network_timeout",
		ErrorCode:	string(AVMAsyncErrorHandlerFailure),
		FailedHeight:	14,
		Attempt:	1,
	}
	_, err := NewAVMAsyncFailureRecord(badClass)
	require.ErrorContains(t, err, "invalid AVM async failure class")

	retryMismatch := badClass
	retryMismatch.FailureClass = AVMAsyncFailureRetryExhausted
	_, err = NewAVMAsyncFailureRecord(retryMismatch)
	require.ErrorContains(t, err, "retry exhausted")

	badCode := badClass
	badCode.FailureClass = AVMAsyncFailureHandlerFailure
	badCode.ErrorCode = "ERR_UNKNOWN"
	_, err = NewAVMAsyncFailureRecord(badCode)
	require.ErrorContains(t, err, "invalid AVM async error code")
}

func TestAVMAsyncBounceCreatesMessageReceiptAndConsumesBoundedGas(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	failure := testAVMFailureRecord(t, msg, AVMAsyncFailureHandlerFailure)
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)

	outcome, err := NewAVMAsyncBounceOutcome(msg, failure, schedule, 7, 99, 14)
	require.NoError(t, err)
	require.NoError(t, outcome.Validate())
	require.Equal(t, msg.Destination, outcome.BounceMessage.Source)
	require.Equal(t, msg.Source, outcome.BounceMessage.Destination)
	require.Equal(t, msg.DestinationZone, outcome.BounceMessage.SourceZone)
	require.Equal(t, msg.SourceZone, outcome.BounceMessage.DestinationZone)
	require.Equal(t, uint64(7), outcome.BounceMessage.ValueNAET)
	require.Equal(t, schedule.BounceGas, outcome.BounceGasUsed)
	require.Equal(t, schedule.BounceGas, outcome.BoundedBounceGas)
	require.Equal(t, AVMReceiptStatusBounced, outcome.BounceReceipt.Status)
	require.Equal(t, msg.ID, outcome.BounceReceipt.MessageID)
	require.Contains(t, string(outcome.BounceMessage.Payload), msg.ID)
	require.Contains(t, string(outcome.BounceMessage.Payload), string(AVMAsyncErrorHandlerFailure))
	require.Contains(t, outcome.BounceMessage.RouteHintOptional, msg.ID)
	require.Equal(t, ComputeAVMAsyncBounceOutcomeHash(outcome), outcome.OutcomeHash)
}

func TestAVMAsyncBounceRejectsDisabledInflatedOversizedAndUnboundedCases(t *testing.T) {
	msg := testAVMFailureMessage(t, 10)
	failure := testAVMFailureRecord(t, msg, AVMAsyncFailureProofVerificationFailure)
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)

	disabled := msg
	disabled.BounceFlag = false
	disabled, err = NewAVMAsyncMessage(disabled)
	require.NoError(t, err)
	_, err = NewAVMAsyncBounceOutcome(disabled, failure, schedule, 1, 88, 14)
	require.ErrorContains(t, err, "bounce flag")

	_, err = NewAVMAsyncBounceOutcome(msg, failure, schedule, msg.ValueNAET+1, 88, 14)
	require.ErrorContains(t, err, "cannot create more value")

	outcome, err := NewAVMAsyncBounceOutcome(msg, failure, schedule, 1, 88, 14)
	require.NoError(t, err)
	overGas := outcome
	overGas.BounceGasUsed = overGas.BoundedBounceGas + 1
	overGas.OutcomeHash = ComputeAVMAsyncBounceOutcomeHash(overGas)
	require.ErrorContains(t, overGas.Validate(), "exceeds bounded bounce gas")

	oversized := outcome
	oversized.BounceMessage.Payload = make([]byte, MaxAVMBouncePayloadBytes+1)
	oversized.BounceMessage.PayloadHash = ComputeAVMAsyncPayloadHash(oversized.BounceMessage.Payload)
	oversized.BounceMessage.ID = DeriveAVMAsyncMessageID(oversized.BounceMessage)
	oversized.OutcomeHash = ComputeAVMAsyncBounceOutcomeHash(oversized)
	require.ErrorContains(t, oversized.Validate(), "payload")
}

func TestAVMAsyncBounceCanDeadLetterWhenBounceFails(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	failure := testAVMFailureRecord(t, msg, AVMAsyncFailureRetryExhausted)
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)
	outcome, err := NewAVMAsyncBounceOutcome(msg, failure, schedule, 5, 100, 14)
	require.NoError(t, err)

	dead, err := NewAVMAsyncBounceDeadLetterOutcome(outcome, "bounce destination disabled", 2, 15)
	require.NoError(t, err)
	require.NoError(t, dead.Validate())
	require.Equal(t, outcome.BounceMessage.ID, dead.DeadLetter.MessageID)
	require.Equal(t, outcome.BounceMessage.DestinationZone, dead.DeadLetter.ZoneID)
	require.Equal(t, AVMReceiptStatusDeadLettered, dead.DeadLetterReceipt.Status)
	require.Equal(t, uint64(5), dead.DeadLetter.RefundAmountOptional)
	require.Equal(t, ComputeAVMAsyncBounceDeadLetterOutcomeHash(dead), dead.OutcomeHash)
}

func TestAVMAsyncFailedReceiptModelCommitsFailure(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	failure := testAVMFailureRecord(t, msg, AVMAsyncFailureStorageLimitExceeded)

	model, err := NewAVMAsyncFailedReceiptModel(msg, failure, "actor-runtime", 2, "", 14)
	require.NoError(t, err)
	require.NoError(t, model.Validate())
	require.Equal(t, AVMReceiptStatusFailed, model.Receipt.Status)
	require.Equal(t, failure.ErrorCode, model.Receipt.ErrorCodeOptional)
	require.Equal(t, failure.FailureHash, model.Receipt.EventsHash)
	require.Equal(t, ComputeAVMAsyncFailureOutputRoot(failure), model.Receipt.OutputMessagesRoot)

	expired := testAVMFailureRecord(t, msg, AVMAsyncFailureExpiredMessage)
	expiredModel, err := NewAVMAsyncFailedReceiptModel(msg, expired, "async-engine", 0, "", 14)
	require.NoError(t, err)
	require.Equal(t, AVMReceiptStatusExpired, expiredModel.Receipt.Status)
}

func TestAVMZoneDeadLetterQueueTriggersAndRetryExhaustion(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	queue, err := NewAVMZoneDeadLetterQueue(AVMZoneDeadLetterQueue{ZoneID: msg.DestinationZone})
	require.NoError(t, err)

	disabled := testAVMFailureRecord(t, msg, AVMAsyncFailureDestinationDisabled)
	trigger, ok := AVMDeadLetterTriggerForFailure(disabled)
	require.True(t, ok)
	require.Equal(t, AVMDeadLetterTriggerDestinationPermanentlyDisabled, trigger)

	queue, record, receipt, err := DeadLetterAVMAsyncFailure(queue, disabled, trigger, "destination permanently disabled", 7, 3, 15)
	require.NoError(t, err)
	require.NoError(t, record.ValidateWithReceipt(receipt))
	require.NoError(t, queue.Validate())
	require.Len(t, queue.Records, 1)
	require.Equal(t, AVMReceiptStatusDeadLettered, receipt.Status)

	other := testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "contract-b", zonestypes.ZoneIDContract, 45, 10)
	other.ValueNAET = 4
	otherMsg, err := NewAVMAsyncMessage(other)
	require.NoError(t, err)
	retry, err := NewAVMRetryExhaustionOutcome(queue, otherMsg, 3, 9, 16, 4)
	require.NoError(t, err)
	require.NoError(t, retry.Validate())
	require.Len(t, retry.Queue.Records, 2)
	require.Equal(t, AVMAsyncFailureRetryExhausted, retry.Failure.FailureClass)
	retryTrigger, ok := AVMDeadLetterTriggerForFailure(retry.Failure)
	require.True(t, ok)
	require.Equal(t, AVMDeadLetterTriggerRetryExhausted, retryTrigger)
	require.Equal(t, string(AVMAsyncErrorRetryExhausted), retry.Failure.ErrorCode)
}

func TestAVMAsyncFailureDeadLetterTriggerRejectsMismatches(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	queue, err := NewAVMZoneDeadLetterQueue(AVMZoneDeadLetterQueue{ZoneID: msg.DestinationZone})
	require.NoError(t, err)
	handler := testAVMFailureRecord(t, msg, AVMAsyncFailureHandlerFailure)
	_, _, _, err = DeadLetterAVMAsyncFailure(queue, handler, AVMDeadLetterTriggerExpiredBeforeExecution, "wrong trigger", 0, 1, 15)
	require.ErrorContains(t, err, "not a dead letter trigger")

	bounceFailed, err := NewAVMAsyncFailureRecord(AVMAsyncFailureRecord{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		FailureClass:	AVMAsyncFailureHandlerFailure,
		ErrorCode:	string(AVMAsyncErrorBounceFailed),
		FailedHeight:	15,
		Attempt:	2,
		GasUsed:	3,
	})
	require.NoError(t, err)
	queue, record, receipt, err := DeadLetterAVMAsyncFailure(queue, bounceFailed, AVMDeadLetterTriggerBounceFailed, "", 0, 3, 15)
	require.NoError(t, err)
	require.Equal(t, "bounce_failed", record.Reason)
	require.NoError(t, record.ValidateWithReceipt(receipt))
	require.NoError(t, queue.Validate())
}

func TestAVMAsyncFailureValueConservation(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	check, err := NewAVMAsyncFailureValueConservationCheck(msg, 7, 0)
	require.NoError(t, err)
	require.NoError(t, check.Validate())

	_, err = NewAVMAsyncFailureValueConservationCheck(msg, msg.ValueNAET+1, 0)
	require.ErrorContains(t, err, "bounce value exceeds")

	_, err = NewAVMAsyncFailureValueConservationCheck(msg, 7, 1)
	require.ErrorContains(t, err, "both bounced and dead-letter refunded")
}

func FuzzAVMAsyncFailureRecordValidation(f *testing.F) {
	f.Add(uint8(0), uint32(1), uint64(0), uint64(14))
	f.Add(uint8(4), uint32(2), uint64(7), uint64(15))
	f.Add(uint8(8), uint32(3), uint64(9), uint64(16))

	classes := []AVMAsyncFailureClass{
		AVMAsyncFailureInvalidPayload,
		AVMAsyncFailureInsufficientGas,
		AVMAsyncFailureDestinationNotFound,
		AVMAsyncFailureDestinationDisabled,
		AVMAsyncFailureExpiredMessage,
		AVMAsyncFailureHandlerFailure,
		AVMAsyncFailureStorageLimitExceeded,
		AVMAsyncFailureProofVerificationFailure,
		AVMAsyncFailureRetryExhausted,
	}

	f.Fuzz(func(t *testing.T, classIndex uint8, attempt uint32, gasUsed uint64, failedHeight uint64) {
		msg := testAVMFailureMessage(t, 25)
		class := classes[int(classIndex)%len(classes)]
		if attempt == 0 {
			attempt = 1
		}
		if failedHeight == 0 {
			failedHeight = 1
		}
		record, err := NewAVMAsyncFailureRecord(AVMAsyncFailureRecord{
			MessageID:	msg.ID,
			ZoneID:		msg.DestinationZone,
			FailureClass:	class,
			FailedHeight:	failedHeight,
			Attempt:	attempt,
			GasUsed:	gasUsed,
			RetryExhausted:	class == AVMAsyncFailureRetryExhausted,
		})
		require.NoError(t, err)
		require.NoError(t, record.Validate())
		require.True(t, IsAVMAsyncErrorCode(AVMAsyncErrorCode(record.ErrorCode)))
	})
}

func testAVMFailureMessage(t *testing.T, valueNAET uint64) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "contract-a", zonestypes.ZoneIDContract, 44, 10)
	msg.ValueNAET = valueNAET
	msg.AuthProofOptional = engineHash("auth-proof")
	msg.StateProofOptional = engineHash("state-proof")
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}

func testAVMFailureRecord(t *testing.T, msg AVMAsyncMessage, class AVMAsyncFailureClass) AVMAsyncFailureRecord {
	t.Helper()
	record, err := NewAVMAsyncFailureRecord(AVMAsyncFailureRecord{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		FailureClass:	class,
		FailedHeight:	14,
		Attempt:	1,
		GasUsed:	11,
		RetryExhausted:	class == AVMAsyncFailureRetryExhausted,
	})
	require.NoError(t, err)
	return record
}
