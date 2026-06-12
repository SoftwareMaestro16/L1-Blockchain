package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAetherMessageReceiptCommitsOutcomeFields(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 100, 2))
	receipt, err := AetherReceiptFromMessage(
		msg,
		ReceiptStatusExecuted,
		104,
		77,
		sdkmath.NewInt(11),
		[]byte("ok"),
		"",
		testMessageHash("output-messages"),
		testMessageHash("state-writes"),
	)
	require.NoError(t, err)
	require.Equal(t, msg.MsgID, receipt.MsgID)
	require.Equal(t, msg.ReceiverZoneID, receipt.ReceiverZoneID)
	require.Equal(t, msg.ReceiverShardID, receipt.ReceiverShardID)
	require.Equal(t, ReceiptStatusExecuted, receipt.Status)
	require.Equal(t, ComputeAetherMessageReceiptHash(receipt), receipt.ReceiptHash)
	require.NoError(t, receipt.Validate())

	encodedA, err := CanonicalAetherMessageReceiptBinary(receipt)
	require.NoError(t, err)
	encodedB, err := CanonicalAetherMessageReceiptBinary(receipt)
	require.NoError(t, err)
	require.Equal(t, encodedA, encodedB)
}

func TestAetherMessageReceiptStatusesAndRootAreCanonical(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 110, 1))
	statuses := []ReceiptStatus{
		ReceiptStatusAccepted,
		ReceiptStatusExecuted,
		ReceiptStatusFailed,
		ReceiptStatusExpired,
		ReceiptStatusBounced,
		ReceiptStatusRejected,
		ReceiptStatusDeferred,
	}
	receipts := make([]AetherMessageReceipt, 0, len(statuses))
	for i, status := range statuses {
		errorCode := ""
		if status == ReceiptStatusFailed || status == ReceiptStatusRejected {
			errorCode = "ERR_DETERMINISTIC_" + string(status)
		}
		receipt, err := NewAetherMessageReceipt(AetherMessageReceipt{
			MsgID:			msg.MsgID,
			Height:			120 + uint64(i),
			ReceiverZoneID:		msg.ReceiverZoneID,
			ReceiverShardID:	msg.ReceiverShardID,
			Status:			status,
			GasUsed:		uint64(i + 1),
			FeeCharged:		sdkmath.NewInt(int64(i)),
			ReturnPayload:		[]byte("return-" + string(status)),
			ErrorCode:		errorCode,
			OutputMessagesRoot:	testMessageHash("out-" + string(status)),
			StateWriteSummaryHash:	testMessageHash("writes-" + string(status)),
		})
		require.NoError(t, err)
		receipts = append(receipts, receipt)
		require.True(t, IsReceiptStatus(status))
	}
	rootA, err := ComputeAetherReceiptRoot(receipts)
	require.NoError(t, err)
	rootB, err := ComputeAetherReceiptRoot([]AetherMessageReceipt{receipts[6], receipts[4], receipts[2], receipts[0], receipts[1], receipts[3], receipts[5]})
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)
	require.NoError(t, zonestypes.ValidateHash("aether receipt root", rootA))
}

func TestAetherMessageReceiptRejectsInvalidConsensusFields(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 130, 1))
	receipt, err := AetherReceiptFromMessage(
		msg,
		ReceiptStatusAccepted,
		131,
		0,
		sdkmath.ZeroInt(),
		nil,
		"",
		EmptyHash(),
		testMessageHash("accepted-state-writes"),
	)
	require.NoError(t, err)

	badStatus := receipt.Clone()
	badStatus.Status = "unknown"
	badStatus.ReceiptHash = ComputeAetherMessageReceiptHash(badStatus)
	require.ErrorContains(t, badStatus.Validate(), "status")

	badMsgID := receipt.Clone()
	badMsgID.MsgID = "bad"
	badMsgID.ReceiptHash = ComputeAetherMessageReceiptHash(badMsgID)
	require.ErrorContains(t, badMsgID.Validate(), "message id")

	badRoot := receipt.Clone()
	badRoot.OutputMessagesRoot = "bad"
	badRoot.ReceiptHash = ComputeAetherMessageReceiptHash(badRoot)
	require.ErrorContains(t, badRoot.Validate(), "output messages root")

	badFee := receipt.Clone()
	badFee.FeeCharged = sdkmath.NewInt(-1)
	badFee.ReceiptHash = ComputeAetherMessageReceiptHash(badFee)
	require.ErrorContains(t, badFee.Validate(), "fee")

	badError := receipt.Clone()
	badError.ErrorCode = "bad code with spaces"
	badError.ReceiptHash = ComputeAetherMessageReceiptHash(badError)
	require.ErrorContains(t, badError.Validate(), "error code")
}
