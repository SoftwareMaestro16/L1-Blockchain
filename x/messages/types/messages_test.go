package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestMessageIDUsesReplayProtectionFormula(t *testing.T) {
	params := testMessageParams()
	msg := testMessage(t, params, 1, 1, []byte("payload"))

	expected := DeriveMessageID(params.ChainID, msg.SourceZone, msg.Sender, msg.Nonce, msg.PayloadHash)
	require.Equal(t, expected, msg.MessageID)
	require.Len(t, msg.MessageID, MessageIDBytes)

	changedPayload := testMessage(t, params, 1, 1, []byte("changed"))
	require.NotEqual(t, msg.MessageID, changedPayload.MessageID)

	changedChain := params
	changedChain.ChainID = "aetra-test-b"
	changedChain.ParamsHash = EmptyHash()
	msgOnOtherChain := testMessage(t, changedChain, 1, 1, []byte("payload"))
	require.NotEqual(t, msg.MessageID, msgOnOtherChain.MessageID)
}

func TestMessageStateKeysAndCanonicalEncoding(t *testing.T) {
	params := testMessageParams()
	msg := testMessage(t, params, 7, 9, []byte("payload"))
	sender := hex.EncodeToString(msg.Sender)
	msgID := hex.EncodeToString(msg.MessageID)

	outboxKey, err := OutboxKey(msg.SourceZone, msg.Sender, msg.SourceSequence)
	require.NoError(t, err)
	require.Equal(t, "messages/outbox/FINANCIAL_ZONE/"+sender+"/9", outboxKey)

	inboxKey, err := InboxKey(msg.DestinationZone, msg.Sender, msg.SourceSequence)
	require.NoError(t, err)
	require.Equal(t, "messages/inbox/CONTRACT_ZONE/"+sender+"/9", inboxKey)

	receiptKey, err := ReceiptKey(msg.MessageID)
	require.NoError(t, err)
	require.Equal(t, "messages/receipts/"+msgID, receiptKey)

	nonceKey, err := NonceKey(msg.SourceZone, msg.Sender)
	require.NoError(t, err)
	require.Equal(t, "messages/nonces/FINANCIAL_ZONE/"+sender, nonceKey)

	replayKey, err := ReplayKey(msg.MessageID)
	require.NoError(t, err)
	require.Equal(t, "messages/replay/"+msgID, replayKey)

	expiryKey, err := ExpiryKey(msg.Deadline, msg.MessageID)
	require.NoError(t, err)
	require.Equal(t, "messages/expiry/100/"+msgID, expiryKey)

	encodedA, err := CanonicalMessageBinary(msg)
	require.NoError(t, err)
	encodedB, err := CanonicalMessageBinary(msg)
	require.NoError(t, err)
	require.Equal(t, encodedA, encodedB)
}

func TestMessageKeeperSubmitRejectsDuplicateIDAndNonceReuse(t *testing.T) {
	params := testMessageParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)

	msg := testMessage(t, params, 1, 1, []byte("payload"))
	keeper, response, err := keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(msg)})
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, response.MessageID)
	require.Len(t, keeper.State().Outbox, 1)
	require.Equal(t, uint64(1), keeper.State().Nonces[0].Nonce)

	_, _, err = keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(msg)})
	require.ErrorContains(t, err, "message id")

	reusedNonce := testMessage(t, params, 1, 2, []byte("other"))
	_, _, err = keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(reusedNonce)})
	require.ErrorContains(t, err, "nonce")
}

func TestMsgServerAndQueryServerAdapters(t *testing.T) {
	params := testMessageParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)
	msgServer, err := NewMsgServer(&keeper)
	require.NoError(t, err)
	queryServer := NewQueryServer(keeper)
	require.NotNil(t, queryServer)

	msg := testMessage(t, params, 1, 1, []byte("payload"))
	resp, err := msgServer.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(msg)})
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, resp.MessageID)

	queryServer = NewQueryServer(keeper)
	found, err := queryServer.Message(QueryMessageRequest{MessageID: msg.MessageID})
	require.NoError(t, err)
	require.True(t, found.Found)
}

var _ MsgServer = KeeperMsgServer{}
var _ QueryServer = MessageKeeper{}

func TestMsgCrossZoneCallValidatesEscrowPayloadAndEnqueues(t *testing.T) {
	params := testMessageParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)

	call := testCrossZoneCall()
	admission := testCrossZoneAdmission(call)
	keeper, response, err := keeper.SubmitCrossZoneCall(call, admission)
	require.NoError(t, err)
	require.Len(t, response.MessageID, MessageIDBytes)
	require.Len(t, keeper.State().Outbox, 1)

	queued := keeper.State().Outbox[0].Message
	require.Equal(t, call.SourceZoneID, queued.SourceZone)
	require.Equal(t, call.DestinationZoneID, queued.DestinationZone)
	require.Equal(t, call.PayloadType, queued.Opcode)
	require.Equal(t, call.ValueNAET, queued.Value)
	require.Equal(t, call.ForwardingFee, queued.FeeLimit)
	require.True(t, queued.Bounce)
	require.Equal(t, "cross-zone-call/caller", queued.AuthScope)
}

func TestMsgCrossZoneCallRejectsUnsupportedPayloadReplyAndMissingEscrow(t *testing.T) {
	params := testMessageParams()
	call := testCrossZoneCall()
	admission := testCrossZoneAdmission(call)

	unsupportedPayload := call
	unsupportedPayload.PayloadType = "identity.resolve"
	require.ErrorContains(t, unsupportedPayload.Validate(admission, params), "payload type")

	unsupportedReply := call
	unsupportedReply.ReplyMode = "stream"
	require.ErrorContains(t, unsupportedReply.Validate(admission, params), "reply mode")

	missingEscrow := admission
	missingEscrow.Escrows = nil
	require.ErrorContains(t, call.Validate(missingEscrow, params), "escrow")
}

func TestMessageKeeperRoutesDrainsReceiptsAndTombstones(t *testing.T) {
	params := testMessageParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)
	msg := testMessage(t, params, 1, 1, []byte("payload"))

	keeper, _, err = keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(msg)})
	require.NoError(t, err)
	keeper, err = keeper.RouteOutboxToInbox(msg.MessageID, 21)
	require.NoError(t, err)
	require.Len(t, keeper.State().Inbox, 1)
	require.Len(t, keeper.State().Outbox, 0)

	keeper, receipts, err := keeper.DrainInbox(22, 10, func(message Message) (MessageReceipt, error) {
		return ReceiptFromMessage(message, MessageStatusExecuted, 55, sdkmath.NewInt(3), ComputePayloadHash([]byte("return")), nil, 22), nil
	})
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, MessageStatusExecuted, receipts[0].Status)
	require.Len(t, keeper.State().Receipts, 1)
	require.Len(t, keeper.State().Tombstones, 1)
	require.Equal(t, uint64(22+params.ProofHorizon), keeper.State().Tombstones[0].RetainUntil)

	_, _, err = keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(msg)})
	require.ErrorContains(t, err, "message id")
}

func TestMessageExpiryBounceAndPruneHorizon(t *testing.T) {
	params := testMessageParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)
	msg := testMessage(t, params, 1, 1, []byte("payload"))

	keeper, _, err = keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(msg)})
	require.NoError(t, err)
	keeper, receipts, err := keeper.ProcessExpiry(msg.Deadline+1, 10)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, MessageStatusExpired, receipts[0].Status)
	require.Len(t, keeper.State().Outbox, 0)
	require.Len(t, keeper.State().Tombstones, 1)

	bounce, err := keeper.BuildBounce(msg, receipts[0], 2, 2, msg.Deadline+1)
	require.NoError(t, err)
	require.Equal(t, msg.DestinationZone, bounce.SourceZone)
	require.Equal(t, msg.SourceZone, bounce.DestinationZone)
	require.Equal(t, msg.Value, bounce.Value)
	require.Equal(t, "aether.bounce", bounce.Opcode)
	require.Contains(t, string(bounce.Payload), "status=expired")

	keeper = keeper.PruneTombstones(receipts[0].ExecutedHeight + params.ProofHorizon + 1)
	require.Len(t, keeper.State().Tombstones, 0)
}

func TestMessageRootsAndProofQueries(t *testing.T) {
	params := testMessageParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)
	first := testMessage(t, params, 1, 1, []byte("payload-1"))
	second := testMessage(t, params, 2, 2, []byte("payload-2"))
	second.Sender = addr(3)
	second.MessageID = nil
	second.PayloadHash = nil
	second, err = NewMessage(second, params)
	require.NoError(t, err)

	keeper, _, err = keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(second)})
	require.NoError(t, err)
	keeper, _, err = keeper.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutCommitments(first)})
	require.NoError(t, err)

	roots := ComputeKeeperRoots(keeper.State())
	require.NoError(t, zonestypes.ValidateHash("message root", roots.MessageRoot))
	require.NoError(t, zonestypes.ValidateHash("receipt root", roots.ReceiptRoot))

	proof, err := keeper.MessageProof(QueryProofRequest{Kind: MessageProofInclusion, MessageID: first.MessageID, Root: roots.MessageRoot, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, MessageProofInclusion, proof.Kind)
	require.NotEmpty(t, proof.ValueHash)

	keeper, err = keeper.RouteOutboxToInbox(first.MessageID, 20)
	require.NoError(t, err)
	keeper, receipts, err := keeper.DrainInbox(21, 10, func(message Message) (MessageReceipt, error) {
		return ReceiptFromMessage(message, MessageStatusExecuted, 1, sdkmath.OneInt(), nil, nil, 21), nil
	})
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	roots = ComputeKeeperRoots(keeper.State())
	receiptProof, err := keeper.MessageProof(QueryProofRequest{Kind: MessageProofReceipt, MessageID: first.MessageID, Root: roots.ReceiptRoot, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, MessageProofReceipt, receiptProof.Kind)

	messageResp, err := keeper.Message(QueryMessageRequest{MessageID: second.MessageID})
	require.NoError(t, err)
	require.True(t, messageResp.Found)
	receiptResp, err := keeper.Receipt(QueryReceiptRequest{MessageID: first.MessageID})
	require.NoError(t, err)
	require.True(t, receiptResp.Found)
}

func testMessageParams() MessageParams {
	params := DefaultMessageParams("aetra-test")
	params.MaxPayloadSize = 256
	params.MinGasLimit = 1
	params.MaxGasLimit = 1_000
	params.MinFeeLimit = sdkmath.NewInt(1)
	params.ProofHorizon = 10
	params.MaxDrainPerBlock = 10
	params.BounceGasReserve = 1
	params.ParamsHash = EmptyHash()
	return params
}

func testMessage(t *testing.T, params MessageParams, nonce uint64, sequence uint64, payload []byte) Message {
	t.Helper()
	msg, err := NewMessage(Message{
		SourceZone:		zonestypes.ZoneIDFinancial,
		DestinationZone:	zonestypes.ZoneIDContract,
		Sender:			addr(1),
		Recipient:		addr(2),
		Value:			sdkmath.NewInt(10),
		Opcode:			"contract.execute",
		Payload:		payload,
		GasLimit:		10,
		Deadline:		100,
		Nonce:			nonce,
		SourceSequence:		sequence,
		RouteID:		"financial/contract",
		Bounce:			true,
		FeeLimit:		sdkmath.NewInt(1),
		CreatedHeight:		20,
		AuthScope:		"owner",
	}, params)
	require.NoError(t, err)
	return msg
}

func msgWithoutCommitments(msg Message) Message {
	msg.MessageID = nil
	msg.PayloadHash = nil
	return msg
}

func addr(seed byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20))
}

func testCrossZoneCall() MsgCrossZoneCall {
	return MsgCrossZoneCall{
		Caller:			addr(1),
		Callee:			addr(2),
		SourceZoneID:		zonestypes.ZoneIDFinancial,
		DestinationZoneID:	zonestypes.ZoneIDContract,
		PayloadType:		"contract.execute",
		Payload:		[]byte("call-body"),
		ValueNAET:		sdkmath.NewInt(10),
		GasLimit:		10,
		ForwardingFee:		sdkmath.NewInt(1),
		ReplyMode:		ReplyModeCaller,
		ExpiryHeight:		100,
	}
}

func testCrossZoneAdmission(call MsgCrossZoneCall) CrossZoneCallAdmission {
	return CrossZoneCallAdmission{
		CreatedHeight:	20,
		Nonce:		11,
		SourceSequence:	12,
		SupportedPayloadTypes: []DestinationPayloadTypes{{
			ZoneID:		call.DestinationZoneID,
			PayloadTypes:	[]string{call.PayloadType},
		}},
		SupportedReplyModes:	[]string{ReplyModeNone, ReplyModeCaller},
		Escrows: []CrossZoneCallEscrow{{
			Caller:			call.Caller,
			SourceZoneID:		call.SourceZoneID,
			DestinationZoneID:	call.DestinationZoneID,
			ValueNAET:		call.ValueNAET,
			ForwardingFee:		call.ForwardingFee,
			ExpiryHeight:		call.ExpiryHeight,
			Escrowed:		true,
		}},
	}
}
