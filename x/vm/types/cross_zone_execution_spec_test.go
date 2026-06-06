package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMCrossZoneRouteAdmitsZoneAToZoneBQueueAndCommitsReceipt(t *testing.T) {
	msg := testAVMCrossZoneMessage(t, true, true, 25)
	policy := testAVMCrossZonePolicy(t, AVMCrossZoneProofAuthAndState, AVMCrossZoneValueEscrow, AVMCrossZoneBounceAllowed)
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: msg.DestinationZone})
	require.NoError(t, err)

	nextQueue, entry, err := AdmitAVMCrossZoneMessage(queue, msg, msg.CreatedHeight, 10, policy)
	require.NoError(t, err)
	require.Equal(t, msg.DestinationZone, entry.ZoneID)
	require.Equal(t, msg.ID, entry.MessageID)
	require.Len(t, nextQueue.DelayedQueue, 1)

	receipt := testAVMCrossZoneReceipt(t, msg, AVMReceiptStatusExecuted, 14)
	execution, err := NewAVMCrossZoneExecution(AVMCrossZoneExecution{
		Message:                  msg,
		RoutePolicy:              policy,
		DestinationQueueEntry:    entry,
		DestinationReceipt:       receipt,
		SourceOutputMessagesRoot: engineHash("source-output"),
		DestinationReceiptRoot:   engineHash("destination-receipt"),
		ValueEscrowedNAET:        msg.ValueNAET,
	})
	require.NoError(t, err)
	require.NoError(t, execution.Validate())
	require.Equal(t, ComputeAVMCrossZoneExecutionHash(execution), execution.ExecutionHash)
}

func TestAVMCrossZoneRouteRejectsDirectWriteFilterProofAndValueDrift(t *testing.T) {
	msg := testAVMCrossZoneMessage(t, true, false, 5)
	policy := testAVMCrossZonePolicy(t, AVMCrossZoneProofAuthAndState, AVMCrossZoneValueEscrow, AVMCrossZoneBounceAllowed)

	require.ErrorContains(t, ValidateAVMCrossZoneRoute(msg, policy), "auth and state proofs")

	withProofs := testAVMCrossZoneMessage(t, true, true, 5)
	disallowed := policy
	disallowed.AllowedOpcodes = []string{"contract.other"}
	disallowed.PolicyHash = ComputeAVMCrossZoneRoutePolicyHash(disallowed)
	require.ErrorContains(t, ValidateAVMCrossZoneRoute(withProofs, disallowed), "opcode")

	noValueAccounting := policy
	noValueAccounting.ValueAccounting = AVMCrossZoneValueNone
	noValueAccounting.PolicyHash = ComputeAVMCrossZoneRoutePolicyHash(noValueAccounting)
	require.ErrorContains(t, ValidateAVMCrossZoneRoute(withProofs, noValueAccounting), "value transfer")

	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: withProofs.DestinationZone})
	require.NoError(t, err)
	_, entry, err := AdmitAVMCrossZoneMessage(queue, withProofs, withProofs.CreatedHeight, 10, policy)
	require.NoError(t, err)
	receipt := testAVMCrossZoneReceipt(t, withProofs, AVMReceiptStatusExecuted, 14)
	execution := AVMCrossZoneExecution{
		Message:                   withProofs,
		RoutePolicy:               policy,
		DestinationQueueEntry:     entry,
		DestinationReceipt:        receipt,
		SourceOutputMessagesRoot:  engineHash("source-output"),
		DestinationReceiptRoot:    engineHash("destination-receipt"),
		DirectStateWriteAttempted: true,
		ValueEscrowedNAET:         withProofs.ValueNAET,
	}
	_, err = NewAVMCrossZoneExecution(execution)
	require.ErrorContains(t, err, "direct state writes")

	execution.DirectStateWriteAttempted = false
	execution.ValueEscrowedNAET = withProofs.ValueNAET - 1
	_, err = NewAVMCrossZoneExecution(execution)
	require.ErrorContains(t, err, "escrow value accounting")
}

func TestAVMCrossZoneFailureMustBounceOrDeadLetter(t *testing.T) {
	msg := testAVMCrossZoneMessage(t, true, true, 1)
	policy := testAVMCrossZonePolicy(t, AVMCrossZoneProofAuthAndState, AVMCrossZoneValueEscrow, AVMCrossZoneBounceRequired)
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: msg.DestinationZone})
	require.NoError(t, err)
	_, entry, err := AdmitAVMCrossZoneMessage(queue, msg, msg.CreatedHeight, 10, policy)
	require.NoError(t, err)
	failed := testAVMCrossZoneReceipt(t, msg, AVMReceiptStatusFailed, 14)

	_, err = NewAVMCrossZoneExecution(AVMCrossZoneExecution{
		Message:                  msg,
		RoutePolicy:              policy,
		DestinationQueueEntry:    entry,
		DestinationReceipt:       failed,
		SourceOutputMessagesRoot: engineHash("source-output"),
		DestinationReceiptRoot:   engineHash("destination-receipt"),
		ValueEscrowedNAET:        msg.ValueNAET,
	})
	require.ErrorContains(t, err, "must bounce or dead-letter")

	bounce := testAVMCrossZoneBounceMessage(t, msg)
	execution, err := NewAVMCrossZoneExecution(AVMCrossZoneExecution{
		Message:                  msg,
		RoutePolicy:              policy,
		DestinationQueueEntry:    entry,
		DestinationReceipt:       failed,
		SourceOutputMessagesRoot: engineHash("source-output"),
		DestinationReceiptRoot:   engineHash("destination-receipt"),
		ValueEscrowedNAET:        msg.ValueNAET,
		FailureResolution:        AVMCrossZoneFailureBounced,
		BounceMessageOptional:    bounce,
	})
	require.NoError(t, err)
	require.NoError(t, execution.Validate())

	badBounce := bounce
	badBounce.RouteHintOptional = "missing-reference"
	badBounce, err = NewAVMAsyncMessage(badBounce)
	require.NoError(t, err)
	_, err = NewAVMCrossZoneExecution(AVMCrossZoneExecution{
		Message:                  msg,
		RoutePolicy:              policy,
		DestinationQueueEntry:    entry,
		DestinationReceipt:       failed,
		SourceOutputMessagesRoot: engineHash("source-output"),
		DestinationReceiptRoot:   engineHash("destination-receipt"),
		ValueEscrowedNAET:        msg.ValueNAET,
		FailureResolution:        AVMCrossZoneFailureBounced,
		BounceMessageOptional:    badBounce,
	})
	require.ErrorContains(t, err, "reference original")
}

func TestAVMCrossZoneDeadLetterResolution(t *testing.T) {
	msg := testAVMCrossZoneMessage(t, true, true, 1)
	policy := testAVMCrossZonePolicy(t, AVMCrossZoneProofAuthAndState, AVMCrossZoneValueEscrow, AVMCrossZoneBounceAllowed)
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: msg.DestinationZone})
	require.NoError(t, err)
	_, entry, err := AdmitAVMCrossZoneMessage(queue, msg, msg.CreatedHeight, 10, policy)
	require.NoError(t, err)
	receipt := testAVMCrossZoneReceipt(t, msg, AVMReceiptStatusDeadLettered, 14)
	dead, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:      msg.ID,
		ZoneID:         msg.DestinationZone,
		Reason:         "cross-zone handler failed",
		FailedAttempts: 3,
		LastErrorCode:  receipt.ErrorCodeOptional,
		FinalHeight:    receipt.CreatedHeight,
		ReceiptID:      receipt.ReceiptID,
	})
	require.NoError(t, err)

	execution, err := NewAVMCrossZoneExecution(AVMCrossZoneExecution{
		Message:                  msg,
		RoutePolicy:              policy,
		DestinationQueueEntry:    entry,
		DestinationReceipt:       receipt,
		SourceOutputMessagesRoot: engineHash("source-output"),
		DestinationReceiptRoot:   engineHash("destination-receipt"),
		ValueEscrowedNAET:        msg.ValueNAET,
		FailureResolution:        AVMCrossZoneFailureDeadLettered,
		DeadLetterOptional:       dead,
	})
	require.NoError(t, err)
	require.NoError(t, execution.Validate())
}

func testAVMCrossZonePolicy(t *testing.T, proof AVMCrossZoneProofRequirement, value AVMCrossZoneValueAccounting, bounce AVMCrossZoneBounceBehavior) AVMCrossZoneRoutePolicy {
	t.Helper()
	policy, err := NewAVMCrossZoneRoutePolicy(AVMCrossZoneRoutePolicy{
		SourceZone:       zonestypes.ZoneIDFinancial,
		DestinationZone:  zonestypes.ZoneIDContract,
		GasPolicy:        zonestypes.DefaultZoneGasPolicy(),
		ExecutionBudget:  zonestypes.ZoneExecutionBudget{MaxGas: 1_000, MaxMessages: 10},
		MessageFilter:    zonestypes.ZoneMessageFilter{AllowedMessageTypes: []string{"contract.call"}},
		AllowedOpcodes:   []string{"contract.call"},
		BounceBehavior:   bounce,
		ProofRequirement: proof,
		ValueAccounting:  value,
	})
	require.NoError(t, err)
	return policy
}

func testAVMCrossZoneMessage(t *testing.T, authProof bool, stateProof bool, value uint64) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage("financial-escrow", zonestypes.ZoneIDFinancial, "contract-service", zonestypes.ZoneIDContract, 17, 10)
	msg.ValueNAET = value
	if authProof {
		msg.AuthProofOptional = engineHash("cross-zone-auth")
	}
	if stateProof {
		msg.StateProofOptional = engineHash("cross-zone-state")
	}
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}

func testAVMCrossZoneBounceMessage(t *testing.T, original AVMAsyncMessage) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage("contract-service", original.DestinationZone, "financial-escrow", original.SourceZone, 18, 11)
	msg.PayloadType = "contract.call"
	msg.ValueNAET = 0
	msg.RouteHintOptional = "bounce:" + original.ID
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}

func testAVMCrossZoneReceipt(t *testing.T, msg AVMAsyncMessage, status AVMReceiptStatus, gasUsed uint64) AVMExecutionReceipt {
	t.Helper()
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:          msg.ID,
		ZoneID:             msg.DestinationZone,
		Executor:           "cross-zone-executor",
		Status:             status,
		GasUsed:            gasUsed,
		StorageWritten:     1,
		EventsHash:         engineHash("cross-zone-events"),
		OutputMessagesRoot: engineHash("cross-zone-output"),
		ErrorCodeOptional:  crossZoneErrorForStatus(status),
		CreatedHeight:      14,
	})
	require.NoError(t, err)
	return receipt
}

func crossZoneErrorForStatus(status AVMReceiptStatus) string {
	switch status {
	case AVMReceiptStatusFailed, AVMReceiptStatusBounced, AVMReceiptStatusDeadLettered:
		return "cross_zone_failed"
	default:
		return ""
	}
}
