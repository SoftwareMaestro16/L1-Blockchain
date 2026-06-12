package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidCrossZoneMessageSucceeds(t *testing.T) {
	state, msg := validMeshFixture(t)

	next, receipt, err := ApplyMessage(state, msg, successResult(), 100)
	require.NoError(t, err)
	require.Equal(t, ReceiptStatusSuccess, receipt.Status)
	require.True(t, receipt.Reason == FailureReasonNone)
	require.Len(t, next.ReplayMarkers, 1)
	require.Len(t, next.Receipts, 1)
	require.Empty(t, next.BounceReceipts)
	require.Empty(t, next.RefundReceipts)
}

func TestDuplicateMessageRejected(t *testing.T) {
	state, msg := validMeshFixture(t)
	next, _, err := ApplyMessage(state, msg, successResult(), 100)
	require.NoError(t, err)

	_, _, err = ApplyMessage(next, msg, successResult(), 101)
	require.ErrorContains(t, err, "replay")
}

func TestDuplicateReceiptRejected(t *testing.T) {
	state, msg := validMeshFixture(t)
	next, receipt, err := ApplyMessage(state, msg, successResult(), 100)
	require.NoError(t, err)

	_, err = CommitReceipt(next, receipt)
	require.ErrorContains(t, err, "duplicate receipt")
}

func TestMissingSourceProofRejected(t *testing.T) {
	state, msg := validMeshFixture(t)
	msg.Proof = MeshProof{}

	_, _, err := ApplyMessage(state, msg, successResult(), 100)
	require.ErrorContains(t, err, "source proof")
}

func TestStaleSourceFinalityRejected(t *testing.T) {
	state, msg := validMeshFixture(t)
	state.Params = MeshParams{MaxFinalityAge: 5}

	_, _, err := ApplyMessage(state, msg, successResult(), 100)
	require.ErrorContains(t, err, "stale")
}

func TestWrongDestinationShardBounces(t *testing.T) {
	state, msg := validMeshFixture(t)
	commitment := state.FinalizedCommitments[0]
	msg.DestinationShard = ShardID("0:wrong")
	msg.MessageID = ""
	msg.Proof = MeshProof{}
	msg = mustMessage(t, msg)
	msg.Proof = BuildProof(msg, commitment)

	next, receipt, err := ApplyMessage(state, msg, successResult(), 100)
	require.NoError(t, err)
	require.Equal(t, ReceiptStatusBounced, receipt.Status)
	require.Equal(t, FailureReasonInvalidDestination, receipt.Reason)
	require.Len(t, next.ReplayMarkers, 1)
	require.Len(t, next.BounceReceipts, 1)
	require.Empty(t, next.RefundReceipts)
}

func TestExpiredMessageBounces(t *testing.T) {
	state, msg := validMeshFixture(t)
	commitment := state.FinalizedCommitments[0]
	msg.TimeoutHeight = 99
	msg.MessageID = ""
	msg.Proof = MeshProof{}
	msg = mustMessage(t, msg)
	msg.Proof = BuildProof(msg, commitment)

	next, receipt, err := ApplyMessage(state, msg, successResult(), 100)
	require.NoError(t, err)
	require.Equal(t, ReceiptStatusBounced, receipt.Status)
	require.Equal(t, FailureReasonExpired, receipt.Reason)
	require.Len(t, next.BounceReceipts, 1)
	require.Equal(t, msg.MessageID, next.BounceReceipts[0].SourceMessageID)
}

func TestFailedExecutionRefunds(t *testing.T) {
	state, msg := validMeshFixture(t)

	next, receipt, err := ApplyMessage(state, msg, failedResult(), 100)
	require.NoError(t, err)
	require.Equal(t, ReceiptStatusRefunded, receipt.Status)
	require.Equal(t, FailureReasonExecutionFailed, receipt.Reason)
	require.Len(t, next.ReplayMarkers, 1)
	require.Len(t, next.RefundReceipts, 1)
	require.Equal(t, msg.Sender, next.RefundReceipts[0].Recipient)
	require.Empty(t, next.BounceReceipts)
}

func TestBounceRefundCannotDoubleSpend(t *testing.T) {
	state, msg := validMeshFixture(t)
	next, _, err := ApplyMessage(state, msg, failedResult(), 100)
	require.NoError(t, err)
	require.Len(t, next.RefundReceipts, 1)

	_, _, err = ApplyMessage(next, msg, failedResult(), 101)
	require.ErrorContains(t, err, "replay")

	state, msg = validMeshFixture(t)
	commitment := state.FinalizedCommitments[0]
	msg.Kind = MessageKindRefund
	msg.ParentMessageID = HashParts("parent-message")
	msg.MessageID = ""
	msg.Proof = MeshProof{}
	msg = mustMessage(t, msg)
	msg.Proof = BuildProof(msg, commitment)

	next, receipt, err := ApplyMessage(state, msg, failedResult(), 100)
	require.NoError(t, err)
	require.Equal(t, ReceiptStatusTerminalFailure, receipt.Status)
	require.Empty(t, next.BounceReceipts)
	require.Empty(t, next.RefundReceipts)
}

func TestExportImportPreservesReplayMarkers(t *testing.T) {
	state, msg := validMeshFixture(t)
	next, _, err := ApplyMessage(state, msg, successResult(), 100)
	require.NoError(t, err)

	exported := next.Export()
	imported, err := ImportState(exported)
	require.NoError(t, err)

	require.Equal(t, exported.ReplayMarkers, imported.ReplayMarkers)
	require.True(t, reflect.DeepEqual(exported, imported))
}

func TestDeterministicMessageOrdering(t *testing.T) {
	state, msg := validMeshFixture(t)
	commitment := state.FinalizedCommitments[0]
	earlier := msg
	earlier.Sequence = 1
	earlier.MessageID = ""
	earlier.Proof = MeshProof{}
	earlier = mustMessage(t, earlier)
	earlier.Proof = BuildProof(earlier, commitment)

	later := msg
	later.Sequence = 9
	later.Nonce++
	later.MessageID = ""
	later.Proof = MeshProof{}
	later = mustMessage(t, later)
	later.Proof = BuildProof(later, commitment)

	ordered := SortMessages([]MeshMessage{later, earlier})
	require.Len(t, ordered, 2)
	require.LessOrEqual(t, CompareMessages(ordered[0], ordered[1]), 0)
	require.Equal(t, ordered, SortMessages([]MeshMessage{ordered[1], ordered[0]}))
}

func validMeshFixture(t *testing.T) (MeshState, MeshMessage) {
	t.Helper()

	state := EmptyState(DefaultParams())
	var err error
	state, err = RegisterDestination(state, MeshDestination{
		ZoneID:		ZoneID("FINANCIAL_ZONE"),
		ShardID:	ShardID("0:0"),
		Active:		true,
	})
	require.NoError(t, err)
	state, err = RegisterDestination(state, MeshDestination{
		ZoneID:		ZoneID("CONTRACT_ZONE"),
		ShardID:	ShardID("0:1"),
		Active:		true,
	})
	require.NoError(t, err)

	commitment := FinalizedCommitment{
		ZoneID:		ZoneID("FINANCIAL_ZONE"),
		ShardID:	ShardID("0:0"),
		Height:		90,
		CommitmentHash:	HashParts("source-commitment", "financial", "0:0", "90"),
		MessageRoot:	HashParts("message-root", "financial", "90"),
		ReceiptRoot:	HashParts("receipt-root", "financial", "90"),
	}
	state, err = AddFinalizedCommitment(state, commitment)
	require.NoError(t, err)

	msg := mustMessage(t, MeshMessage{
		SourceZone:		ZoneID("FINANCIAL_ZONE"),
		SourceShard:		ShardID("0:0"),
		DestinationZone:	ZoneID("CONTRACT_ZONE"),
		DestinationShard:	ShardID("0:1"),
		Nonce:			7,
		Sender:			[]byte("orb1sender"),
		Recipient:		[]byte("contract1recipient"),
		AssetCommitment:	HashParts("asset", "100norb"),
		PayloadHash:		HashParts("payload", "execute"),
		TimeoutHeight:		150,
		Finality:		FinalityReference{Height: commitment.Height, CommitmentHash: commitment.CommitmentHash},
		Sequence:		3,
		SourceLogicalTime:	88,
	})
	msg.Proof = BuildProof(msg, commitment)
	return state, msg
}

func mustMessage(t *testing.T, msg MeshMessage) MeshMessage {
	t.Helper()
	out, err := NewMessage(msg)
	require.NoError(t, err)
	return out
}

func successResult() ExecutionResult {
	return ExecutionResult{
		Success:	true,
		Code:		0,
		ResultHash:	HashParts("execution", "success"),
	}
}

func failedResult() ExecutionResult {
	return ExecutionResult{
		Success:	false,
		Code:		42,
		ResultHash:	HashParts("execution", "failed"),
	}
}
