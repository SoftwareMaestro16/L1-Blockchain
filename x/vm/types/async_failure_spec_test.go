package types

import (
	"strings"
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
			MessageID:      msg.ID,
			ZoneID:         msg.DestinationZone,
			FailureClass:   class,
			ErrorCode:      "ERR_" + strings.ToUpper(string(class)),
			FailedHeight:   14,
			Attempt:        uint32(i + 1),
			GasUsed:        uint64(i),
			RetryExhausted: class == AVMAsyncFailureRetryExhausted,
		})
		require.NoError(t, err)
		require.NoError(t, record.Validate())
		require.Equal(t, ComputeAVMAsyncFailureHash(record), record.FailureHash)
	}

	badClass := AVMAsyncFailureRecord{
		MessageID:    msg.ID,
		ZoneID:       msg.DestinationZone,
		FailureClass: "network_timeout",
		ErrorCode:    "ERR_UNKNOWN",
		FailedHeight: 14,
		Attempt:      1,
	}
	_, err := NewAVMAsyncFailureRecord(badClass)
	require.ErrorContains(t, err, "invalid AVM async failure class")

	retryMismatch := badClass
	retryMismatch.FailureClass = AVMAsyncFailureRetryExhausted
	_, err = NewAVMAsyncFailureRecord(retryMismatch)
	require.ErrorContains(t, err, "retry exhausted")
}

func TestAVMAsyncBounceCreatesMessageReceiptAndConsumesBoundedGas(t *testing.T) {
	msg := testAVMFailureMessage(t, 25)
	failure := testAVMFailureRecord(t, msg, AVMAsyncFailureHandlerFailure, "ERR_HANDLER")
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
	require.Contains(t, string(outcome.BounceMessage.Payload), "ERR_HANDLER")
	require.Contains(t, outcome.BounceMessage.RouteHintOptional, msg.ID)
	require.Equal(t, ComputeAVMAsyncBounceOutcomeHash(outcome), outcome.OutcomeHash)
}

func TestAVMAsyncBounceRejectsDisabledInflatedOversizedAndUnboundedCases(t *testing.T) {
	msg := testAVMFailureMessage(t, 10)
	failure := testAVMFailureRecord(t, msg, AVMAsyncFailureProofVerificationFailure, "ERR_PROOF")
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
	failure := testAVMFailureRecord(t, msg, AVMAsyncFailureRetryExhausted, "ERR_RETRY")
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

func testAVMFailureRecord(t *testing.T, msg AVMAsyncMessage, class AVMAsyncFailureClass, errorCode string) AVMAsyncFailureRecord {
	t.Helper()
	record, err := NewAVMAsyncFailureRecord(AVMAsyncFailureRecord{
		MessageID:      msg.ID,
		ZoneID:         msg.DestinationZone,
		FailureClass:   class,
		ErrorCode:      errorCode,
		FailedHeight:   14,
		Attempt:        1,
		GasUsed:        11,
		RetryExhausted: class == AVMAsyncFailureRetryExhausted,
	})
	require.NoError(t, err)
	return record
}
