package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentSettlementRulesAndMessageDescriptorsCoverSectionsNineEightNine(t *testing.T) {
	require.NoError(t, ValidatePaymentSettlementRulesAndMessages())
	require.Len(t, PaymentSettlementRules(), 6)
	require.Len(t, PaymentMessageDescriptors(), 9)

	byMessage := map[PaymentMessageType]PaymentMessageDescriptor{}
	for _, descriptor := range PaymentMessageDescriptors() {
		byMessage[descriptor.Message] = descriptor
	}
	require.Contains(t, byMessage[PaymentMessageCreateIntent].RequiredValidation, "idempotency-key")
	require.Contains(t, byMessage[PaymentMessageOpenChannel].RequiredValidation, "collateral-availability")
	require.Contains(t, byMessage[PaymentMessageUpdateChannel].RequiredValidation, "nonce-monotonicity")
	require.Contains(t, byMessage[PaymentMessageDisputeChannel].RequiredValidation, "active-challenge-period")
	require.Contains(t, byMessage[PaymentMessageSettlePayment].RequiredValidation, "payment-state-root-consistency")
}

func TestPaymentMessagesValidateCreateOpenUpdateCloseAndDispute(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	createMsg, intent, err := BuildMsgCreatePaymentIntent(MsgCreatePaymentIntent{
		PaymentID:	HashParts("msg-create-payment-intent"),
		Payer:		fixture.Intent.Payer,
		Payee:		fixture.Intent.Payee,
		Amount:		"100",
		Denom:		NativeDenom,
		MaxFee:		"5",
		ExpiryHeight:	100,
		IdempotencyKey:	HashParts("msg-create-payment-intent-idempotency"),
	})
	require.NoError(t, err)
	require.NoError(t, createMsg.Validate())
	require.Equal(t, createMsg.PaymentID, intent.PaymentID)

	_, _, err = BuildMsgCreatePaymentIntent(MsgCreatePaymentIntent{
		PaymentID:	HashParts("bad-denom-intent"),
		Payer:		fixture.Intent.Payer,
		Payee:		fixture.Intent.Payee,
		Amount:		"100",
		Denom:		"uatom",
		MaxFee:		"5",
		ExpiryHeight:	100,
		IdempotencyKey:	HashParts("bad-denom-idempotency"),
	})
	require.ErrorContains(t, err, "denom")

	openMsg, err := BuildMsgOpenPaymentChannel(MsgOpenPaymentChannel{
		Channel:		fixture.Channel,
		CollateralAvailable:	fixture.Channel.Collateral,
		ParticipantSignatures:	fixture.Channel.Participants,
		IdempotencyKey:		HashParts("msg-open-channel-idempotency"),
	}, nil)
	require.NoError(t, err)
	require.NoError(t, openMsg.ValidateWithExisting(nil))

	_, err = BuildMsgOpenPaymentChannel(openMsg, []PaymentChannel{fixture.Channel})
	require.ErrorContains(t, err, "empty before construction")

	updateMsg, err := BuildMsgUpdatePaymentChannel(MsgUpdatePaymentChannel{
		ChannelID:		fixture.Channel.ChannelID,
		Submitter:		fixture.Channel.Participants[0],
		PreviousNonce:		fixture.Channel.LatestNonce,
		NewNonce:		fixture.Channel.LatestNonce + 1,
		SignedStateHash:	HashParts("msg-update-signed-state"),
		BalanceRoot:		HashParts("msg-update-balance-root"),
		ConditionRoot:		HashParts("msg-update-condition-root"),
	}, fixture.Channel)
	require.NoError(t, err)
	require.NoError(t, updateMsg.ValidateForChannel(fixture.Channel))

	badUpdate := updateMsg
	badUpdate.MessageHash = ""
	badUpdate.NewNonce = badUpdate.PreviousNonce
	_, err = BuildMsgUpdatePaymentChannel(badUpdate, fixture.Channel)
	require.ErrorContains(t, err, "nonce")

	closeMsg, err := BuildMsgClosePaymentChannel(MsgClosePaymentChannel{
		PaymentID:		fixture.Intent.PaymentID,
		ChannelID:		fixture.Channel.ChannelID,
		LatestStateHash:	fixture.Channel.LatestStateHash,
		ChallengeStart:		90,
		ChallengeEnd:		106,
		SettlementStatus:	NativePaymentSettlementClosing,
		CollateralRoot:		fixture.Channel.ChannelRoot,
	}, fixture.Channel)
	require.NoError(t, err)
	require.NoError(t, closeMsg.ValidateForChannel(fixture.Channel))

	disputeMsg, err := BuildMsgDisputePaymentChannel(MsgDisputePaymentChannel{
		Dispute:	fixture.Dispute,
		StaleNonce:	1,
		NewerNonce:	2,
		CurrentHeight:	90,
	})
	require.NoError(t, err)
	require.NoError(t, disputeMsg.Validate())

	lateDispute := disputeMsg
	lateDispute.MessageHash = ""
	lateDispute.CurrentHeight = fixture.Dispute.ChallengeEnd + 1
	_, err = BuildMsgDisputePaymentChannel(lateDispute)
	require.ErrorContains(t, err, "challenge")
}

func TestConditionalPaymentMessagesResolveAndExpireDeterministically(t *testing.T) {
	alice := testAddress(0xb1)
	bob := testAddress(0xb2)
	preimage := "msg-condition-preimage"
	condition, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:	HashParts("msg-condition"),
		Payer:		alice,
		Payee:		bob,
		Amount:		"44",
		HashLock:	HashParts(preimage),
		TimeoutHeight:	50,
		RouteID:	HashParts("msg-condition-route"),
	})
	require.NoError(t, err)

	createMsg, err := BuildMsgCreateConditionalPayment(MsgCreateConditionalPayment{
		Condition:		condition,
		Payer:			alice,
		ReservedLiquidity:	"44",
	})
	require.NoError(t, err)
	require.NoError(t, createMsg.Validate())

	_, err = BuildMsgCreateConditionalPayment(MsgCreateConditionalPayment{
		Condition:		condition,
		Payer:			bob,
		ReservedLiquidity:	"44",
	})
	require.ErrorContains(t, err, "payer")

	resolveMsg, outcomes, err := BuildMsgResolveConditionalPayment(MsgResolveConditionalPayment{
		Conditions:		[]NativeConditionalPayment{condition},
		Preimage:		preimage,
		PaymentStateRoot:	HashParts("msg-condition-state-root"),
		CurrentHeight:		45,
	})
	require.NoError(t, err)
	require.NoError(t, resolveMsg.ValidateFormat())
	require.Len(t, outcomes, 1)
	require.Equal(t, bob, outcomes[0].Recipient)

	expiring, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:	HashParts("msg-condition-expire"),
		Payer:		alice,
		Payee:		bob,
		Amount:		"12",
		TimeoutHeight:	10,
		RouteID:	HashParts("msg-condition-expire-route"),
	})
	require.NoError(t, err)
	expireMsg, outcome, err := BuildMsgExpireConditionalPayment(MsgExpireConditionalPayment{
		Condition:		expiring,
		Resolver:		bob,
		RefundRouteRoot:	HashParts("msg-condition-refund-route"),
		PaymentStateRoot:	HashParts("msg-condition-expire-state"),
		CurrentHeight:		11,
	})
	require.NoError(t, err)
	require.NoError(t, expireMsg.ValidateFormat())
	require.Equal(t, alice, outcome.Recipient)
	require.Equal(t, NativeConditionalPaymentRefunded, outcome.Status)
}

func TestSettlePaymentMessageRequiresCommittedRouteReceiptAndPaymentRoots(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	state, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:		100,
		Settlements:	[]PaymentSettlement{fixture.Settlement},
		Routes:		[]PaymentRouteCommitment{fixture.RouteCommitment},
		Receipts:	[]PaymentReceipt{fixture.Receipt},
		Proofs:		[]SettlementProof{fixture.Proof},
	})
	require.NoError(t, err)

	msg, err := BuildMsgSettlePayment(MsgSettlePayment{
		Settlement:		fixture.Settlement,
		RouteCommitment:	fixture.RouteCommitment,
		ReceiptRoot:		state.ReceiptRoot,
		PaymentStateRoot:	state.PaymentRoot,
		FinancialStateRoot:	state.PaymentRoot,
	}, state)
	require.NoError(t, err)
	require.NoError(t, msg.ValidateAgainstState(state))

	tampered := msg
	tampered.MessageHash = ""
	tampered.ReceiptRoot = HashParts("wrong-receipt-root")
	_, err = BuildMsgSettlePayment(tampered, state)
	require.ErrorContains(t, err, "receipt root")
}

func TestFinalSettlementWritesFinancialZoneStateAndChangesPaymentRoot(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	before, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:		99,
		Intents:	[]PaymentIntent{fixture.Intent},
	})
	require.NoError(t, err)
	after, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:		100,
		Intents:	[]PaymentIntent{fixture.Intent},
		Settlements:	[]PaymentSettlement{fixture.Settlement},
		Receipts:	[]PaymentReceipt{fixture.Receipt},
	})
	require.NoError(t, err)
	require.NoError(t, ValidateFinalSettlementWritesFinancialZoneState(before, after, fixture.Settlement))

	unchanged := before
	err = ValidateFinalSettlementWritesFinancialZoneState(before, unchanged, fixture.Settlement)
	require.ErrorContains(t, err, "must change")
}
