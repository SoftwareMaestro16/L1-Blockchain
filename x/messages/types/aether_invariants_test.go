package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestAetherMsgBusStateStoresRootsAndProofs(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 200, 1))
	receipt, err := AetherReceiptFromMessage(msg, ReceiptStatusExecuted, 100, 10, sdkmath.OneInt(), []byte("ok"), "", EmptyHash(), testMessageHash("state-writes"))
	require.NoError(t, err)
	escrow := lockedEscrow(t, msg)

	state, err := NewAetherMsgBusState([]AetherMessage{msg}, nil, []AetherMessageReceipt{receipt}, []AetherValueEscrow{escrow}, []string{msg.MsgID})
	require.NoError(t, err)
	require.Equal(t, ComputeAetherGlobalMessageRoot(state.Outbox, state.Inbox), state.MessageRoot)
	require.NoError(t, state.Validate())

	msgProof, err := BuildAetherMessageInclusionProof(state.Outbox, msg.MsgID, ComputeAetherMessageListRoot(state.Outbox))
	require.NoError(t, err)
	require.Equal(t, AetherProofMessageInclusion, msgProof.Kind)
	require.NoError(t, msgProof.Validate())

	receiptProof, err := BuildAetherReceiptInclusionProof(state.Receipts, msg.MsgID, state.ReceiptRoot)
	require.NoError(t, err)
	require.Equal(t, receipt.ReceiptHash, receiptProof.ValueHash)
	require.NoError(t, receiptProof.Validate())

	envelopeKey, err := MsgBusEnvelopeKey(msg.MsgID)
	require.NoError(t, err)
	require.Equal(t, "msgbus/envelopes/"+msg.MsgID, envelopeKey)
	outboxKey, err := MsgBusOutboxKey(string(msg.SenderZoneID), msg.SenderShardID, msg.MsgID)
	require.NoError(t, err)
	require.Contains(t, outboxKey, "msgbus/outbox/FINANCIAL_ZONE/financial-1/")
}

func TestAetherMessageInvariantsRejectDuplicateAndUnescrowedValue(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 210, 1))
	escrow := lockedEscrow(t, msg)

	_, err := NewAetherMsgBusState([]AetherMessage{msg}, []AetherMessage{msg}, nil, []AetherValueEscrow{escrow}, nil)
	require.ErrorContains(t, err, "globally unique")

	_, err = NewAetherMsgBusState([]AetherMessage{msg}, nil, nil, nil, nil)
	require.ErrorContains(t, err, "value escrow")

	underfunded := escrow
	underfunded.ValueLocked = sdkmath.ZeroInt()
	underfunded.EscrowHash = ComputeAetherValueEscrowHash(underfunded)
	_, err = NewAetherMsgBusState([]AetherMessage{msg}, nil, nil, []AetherValueEscrow{underfunded}, nil)
	require.ErrorContains(t, err, "cover")
}

func TestAetherApplyInboundExpiredConvertsToReceipt(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 220, 1))
	receipt, err := ApplyAetherInboundMessage(msg, msg.ExpiryHeight+1, func(AetherMessage) (AetherMessageReceipt, error) {
		t.Fatal("expired message must not execute")
		return AetherMessageReceipt{}, nil
	})
	require.NoError(t, err)
	require.Equal(t, ReceiptStatusExpired, receipt.Status)
	require.Equal(t, "ERR_MESSAGE_EXPIRED", receipt.ErrorCode)
	require.NoError(t, receipt.Validate())
}

func TestAetherPayloadPolicyAndRouteInvariant(t *testing.T) {
	draft := testAetherMessageDraft(testAetherRoute(t, 230, 1))
	table := testRoutingTable(t)
	msg, plan, err := CommitAetherMessageDeterministicRoute(draft, table, nil, nil, testRoutingParams())
	require.NoError(t, err)
	require.NoError(t, ValidateAetherRouteReproducible(msg, plan))

	bad := plan
	bad.RouteCommitment = testMessageHash("wrong-route")
	require.ErrorContains(t, ValidateAetherRouteReproducible(msg, bad), "route")

	policy := AetherPayloadExecutionPolicy{NoExternalAPIs: true, NoWallClockTime: true, MeteredIteration: true}
	policy.PolicyHash = ComputeAetherPayloadExecutionPolicyHash(policy)
	require.NoError(t, policy.Validate())
	policy.NoWallClockTime = false
	policy.PolicyHash = ComputeAetherPayloadExecutionPolicyHash(policy)
	require.ErrorContains(t, policy.Validate(), "wall-clock")
}

func TestAetherMsgBusLoadRootsAreDeterministic(t *testing.T) {
	outbox := make([]AetherMessage, 0, 64)
	escrows := make([]AetherValueEscrow, 0, 64)
	for i := 0; i < 64; i++ {
		route := testAetherRoute(t, uint64(240+i), 1)
		msg, err := NewAetherMessage(AetherMessage{
			Sender:			"account/alice",
			SenderZoneID:		route.SourceZoneID,
			SenderShardID:		route.SourceShardID,
			Receiver:		"contract/vault",
			ReceiverZoneID:		route.DestinationZoneID,
			ReceiverShardID:	route.DestinationShardID,
			ValueNAET:		sdkmath.NewInt(1),
			Payload:		[]byte(fmt.Sprintf("payload-%d", i)),
			PayloadType:		"contract.execute",
			GasLimit:		100,
			GasPrice:		sdkmath.OneInt(),
			ForwardingFee:		sdkmath.OneInt(),
			ExpiryHeight:		uint64(400 + i),
			Bounce:			true,
			ExecutionMode:		ExecutionModeAsync,
			OrderingClass:		OrderingClassSenderOrdered,
			RouteCommitment:	route.RouteCommitment,
			CreatedAtHeight:	uint64(240 + i),
			Nonce:			uint64(i + 1),
		})
		require.NoError(t, err)
		outbox = append(outbox, msg)
		escrows = append(escrows, lockedEscrow(t, msg))
	}
	stateA, err := NewAetherMsgBusState(outbox, nil, nil, escrows, nil)
	require.NoError(t, err)
	reversedOutbox := append([]AetherMessage(nil), outbox...)
	reversedEscrows := append([]AetherValueEscrow(nil), escrows...)
	for i, j := 0, len(reversedOutbox)-1; i < j; i, j = i+1, j-1 {
		reversedOutbox[i], reversedOutbox[j] = reversedOutbox[j], reversedOutbox[i]
		reversedEscrows[i], reversedEscrows[j] = reversedEscrows[j], reversedEscrows[i]
	}
	stateB, err := NewAetherMsgBusState(reversedOutbox, nil, nil, reversedEscrows, nil)
	require.NoError(t, err)
	require.Equal(t, stateA.MessageRoot, stateB.MessageRoot)
	require.Equal(t, stateA.StateRoot, stateB.StateRoot)
}

func lockedEscrow(t *testing.T, msg AetherMessage) AetherValueEscrow {
	t.Helper()
	escrow, err := NewAetherValueEscrow(AetherValueEscrow{
		MsgID:		msg.MsgID,
		ValueLocked:	msg.ValueNAET,
		FeeLocked:	msg.ForwardingFee,
		Status:		AetherEscrowLocked,
	})
	require.NoError(t, err)
	return escrow
}
