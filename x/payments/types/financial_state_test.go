package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinancialPaymentStateKeysMatchSectionNineSeven(t *testing.T) {
	paymentID := HashParts("financial-state-payment")
	channelID := HashParts("financial-state-channel")
	conditionID := HashParts("financial-state-condition")
	routeID := HashParts("financial-state-route")
	disputeID := HashParts("financial-state-dispute")

	intentKey, err := FinancialPaymentIntentStateKey(paymentID)
	require.NoError(t, err)
	require.Equal(t, "financial/payments/intents/"+paymentID, intentKey)
	channelKey, err := FinancialPaymentChannelStateKey(channelID)
	require.NoError(t, err)
	require.Equal(t, "financial/payments/channels/"+channelID, channelKey)
	conditionKey, err := FinancialPaymentConditionStateKey(conditionID)
	require.NoError(t, err)
	require.Equal(t, "financial/payments/conditions/"+conditionID, conditionKey)
	routeKey, err := FinancialPaymentRouteStateKey(routeID)
	require.NoError(t, err)
	require.Equal(t, "financial/payments/routes/"+routeID, routeKey)
	settlementKey, err := FinancialPaymentSettlementStateKey(paymentID)
	require.NoError(t, err)
	require.Equal(t, "financial/payments/settlements/"+paymentID, settlementKey)
	disputeKey, err := FinancialPaymentDisputeStateKey(disputeID)
	require.NoError(t, err)
	require.Equal(t, "financial/payments/disputes/"+disputeID, disputeKey)
}

func TestPaymentImplementationTasksCoverSectionNineSix(t *testing.T) {
	require.NoError(t, ValidatePaymentImplementationTasks())
	require.Len(t, PaymentImplementationTasks(), 13)

	seen := map[PaymentImplementationTaskID]PaymentImplementationTask{}
	for _, task := range PaymentImplementationTasks() {
		seen[task.TaskID] = task
		require.NotEmpty(t, task.AcceptanceCriteria)
	}
	require.Equal(t, PaymentImplementationP0, seen[PaymentTaskFinancialZoneModule].Priority)
	require.Contains(t, seen[PaymentTaskCanonicalEnvelope].AcceptanceCriteria, "hash-settlements")
	require.Contains(t, seen[PaymentTaskChannelCollateralEscrow].AcceptanceCriteria, "preserve-value-conservation")
	require.Contains(t, seen[PaymentTaskCrossZoneMessages].AcceptanceCriteria, "bounce-message")
	require.Contains(t, seen[PaymentTaskAdversarialTests].AcceptanceCriteria, "preserve-deterministic-replay")
}

func TestFinancialZonePaymentStateCommitsAllPaymentRootsDeterministically(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	state, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:			90,
		Intents:		[]PaymentIntent{fixture.Intent},
		Channels:		[]PaymentChannel{fixture.Channel},
		Conditions:		[]NativeConditionalPayment{fixture.Condition},
		Routes:			[]PaymentRouteCommitment{fixture.RouteCommitment},
		Settlements:		[]PaymentSettlement{fixture.Settlement},
		Disputes:		[]PaymentDispute{fixture.Dispute},
		Receipts:		[]PaymentReceipt{fixture.Receipt},
		Proofs:			[]SettlementProof{fixture.Proof},
		Fees:			[]PaymentFeeAccountingRecord{fixture.Fee},
		Messages:		[]CrossZonePaymentMessage{fixture.Message},
		CanonicalEnvelopes:	[]PaymentEnvelopeCanonicalRecord{fixture.CanonicalIntent, fixture.CanonicalSettlement},
	})
	require.NoError(t, err)
	require.NoError(t, state.Validate())
	require.NotEmpty(t, state.PaymentRoot)
	require.Equal(t, ComputePaymentIntentSetRoot(state.Intents), state.IntentRoot)
	require.Equal(t, ComputePaymentSettlementSetRoot(state.Settlements), state.SettlementRoot)
	require.Equal(t, ComputeNativePaymentReceiptSetRoot(state.Receipts), state.ReceiptRoot)

	reordered, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:			90,
		CanonicalEnvelopes:	[]PaymentEnvelopeCanonicalRecord{fixture.CanonicalSettlement, fixture.CanonicalIntent},
		Messages:		[]CrossZonePaymentMessage{fixture.Message},
		Fees:			[]PaymentFeeAccountingRecord{fixture.Fee},
		Proofs:			[]SettlementProof{fixture.Proof},
		Receipts:		[]PaymentReceipt{fixture.Receipt},
		Disputes:		[]PaymentDispute{fixture.Dispute},
		Settlements:		[]PaymentSettlement{fixture.Settlement},
		Routes:			[]PaymentRouteCommitment{fixture.RouteCommitment},
		Conditions:		[]NativeConditionalPayment{fixture.Condition},
		Channels:		[]PaymentChannel{fixture.Channel},
		Intents:		[]PaymentIntent{fixture.Intent},
	})
	require.NoError(t, err)
	require.Equal(t, state.PaymentRoot, reordered.PaymentRoot)
}

func TestPaymentEnvelopeCanonicalEncodingRejectsWrongPrefixAndVersion(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	require.NoError(t, fixture.CanonicalIntent.Validate())

	wrongPrefix := fixture.CanonicalIntent
	wrongPrefix.StateKey = "zone/financial/payments/intents/" + fixture.Intent.PaymentID
	wrongPrefix.EnvelopeHash = ""
	_, err := BuildPaymentEnvelopeCanonicalRecord(wrongPrefix)
	require.ErrorContains(t, err, "financial payments prefix")

	wrongVersion := fixture.CanonicalIntent
	wrongVersion.EncodingVersion = 99
	wrongVersion.EnvelopeHash = ""
	_, err = BuildPaymentEnvelopeCanonicalRecord(wrongVersion)
	require.ErrorContains(t, err, "encoding version")
}

func TestPaymentSettlementProofQueryFindsProofUnderPaymentRoot(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	state, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:		91,
		Settlements:	[]PaymentSettlement{fixture.Settlement},
		Proofs:		[]SettlementProof{fixture.Proof},
	})
	require.NoError(t, err)
	resp, err := QueryPaymentSettlementProofFromState(state, PaymentSettlementProofQuery{
		PaymentID:	fixture.Settlement.PaymentID,
		ProofType:	fixture.Proof.ProofType,
	})
	require.NoError(t, err)
	require.True(t, resp.Found)
	require.Equal(t, fixture.Proof.ProofRoot, resp.Proof.ProofRoot)

	missing, err := QueryPaymentSettlementProofFromState(state, PaymentSettlementProofQuery{
		PaymentID:	HashParts("missing-payment"),
		ProofType:	SettlementProofLatestState,
	})
	require.NoError(t, err)
	require.False(t, missing.Found)
}

func TestFinancialZonePaymentStateRejectsDuplicateReceiptsAndUnsignedRoutes(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	_, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:		92,
		Receipts:	[]PaymentReceipt{fixture.Receipt, fixture.Receipt},
	})
	require.ErrorContains(t, err, "duplicate native receipt")

	route := fixture.RouteCommitment
	route.Signed = false
	route.Reserved = false
	_, err = BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:	93,
		Routes:	[]PaymentRouteCommitment{route},
	})
	require.ErrorContains(t, err, "signed or reserved")
}

type financialPaymentStateFixtureSet struct {
	Intent			PaymentIntent
	Channel			PaymentChannel
	Condition		NativeConditionalPayment
	RouteCommitment		PaymentRouteCommitment
	Settlement		PaymentSettlement
	Dispute			PaymentDispute
	Receipt			PaymentReceipt
	Proof			SettlementProof
	Fee			PaymentFeeAccountingRecord
	Message			CrossZonePaymentMessage
	CanonicalIntent		PaymentEnvelopeCanonicalRecord
	CanonicalSettlement	PaymentEnvelopeCanonicalRecord
}

func financialPaymentStateFixture(t *testing.T) financialPaymentStateFixtureSet {
	t.Helper()

	routing := nativeRoutingFixture(t)
	paymentID := HashParts("financial-payment-fixture")
	condition, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:	routing.Condition.ConditionID,
		Payer:		routing.Condition.Payer,
		Payee:		routing.Condition.Payee,
		Amount:		routing.Condition.Amount,
		HashLock:	routing.Condition.HashLock,
		TimeoutHeight:	routing.Condition.TimeoutHeight,
		RouteID:	routing.Route.RouteID,
		Status:		NativeConditionalPaymentPending,
	})
	require.NoError(t, err)
	intent, err := BuildPaymentIntent(PaymentIntent{
		PaymentID:		paymentID,
		IntentType:		PaymentIntentInitiate,
		Payer:			routing.Route.Payer,
		Payee:			routing.Route.Payee,
		Amount:			routing.Route.Amount,
		MaxFee:			routing.Route.MaxFee,
		RouteIDOptional:	routing.Route.RouteID,
		ExpiryHeight:		routing.Route.ExpiryHeight,
	})
	require.NoError(t, err)

	routeCommitment := PaymentRouteCommitment{
		RouteID:	routing.Route.RouteID,
		Committer:	routing.Route.Payer,
		CommitmentHash:	routing.Route.RouteCommitment,
		Signed:		true,
		ExpiresHeight:	routing.Route.ExpiryHeight,
	}.Normalize()

	settlement, err := BuildPaymentSettlement(PaymentSettlement{
		PaymentID:	paymentID,
		ChannelID:	routing.Channel.ChannelID,
		RouteID:	routing.Route.RouteID,
		FinalStateHash:	routing.Channel.LatestStateHash,
		ReceiptHash:	routing.Receipt.ReceiptHash,
		CloseStatus:	NativePaymentSettlementSettled,
		ProofRoot:	routing.Proof.ProofRoot,
		SettledHeight:	90,
	})
	require.NoError(t, err)

	dispute, err := BuildPaymentDispute(PaymentDispute{
		DisputeID:	HashParts("financial-payment-dispute"),
		PaymentID:	paymentID,
		ChannelID:	routing.Channel.ChannelID,
		FraudProofHash:	HashParts("financial-payment-fraud-proof"),
		StaleStateHash:	HashParts("financial-payment-stale-state"),
		NewerStateHash:	routing.Channel.LatestStateHash,
		SubmittedBy:	routing.Route.Payer,
		OpenedHeight:	80,
		ChallengeEnd:	96,
		Status:		PaymentDisputeAccepted,
	})
	require.NoError(t, err)

	fee, err := BuildPaymentFeeAccountingRecord(PaymentFeeAccountingRecord{
		FeeID:			HashParts("financial-payment-fees"),
		RouteID:		routing.Route.RouteID,
		ForwardingFee:		"2",
		RouteFee:		"3",
		ReserveFee:		"1",
		SettlementGasFee:	"4",
		RecordedHeight:		90,
	})
	require.NoError(t, err)

	message, err := BuildCrossZonePaymentMessage(CrossZonePaymentMessage{
		SourceZoneID:		"financial",
		DestinationZoneID:	"contract",
		SourceShardID:		1,
		DestinationShardID:	2,
		PayloadType:		string(CrossZonePaymentMessageSettle),
		RouteID:		routing.Route.RouteID,
		RouteCommitmentHash:	routing.Route.RouteCommitment,
		PaymentStateRoot:	settlement.SettlementHash,
		UnifiedMessageRoot:	HashParts("financial-payment-message-root"),
		ExpiryHeight:		routing.Route.ExpiryHeight,
	})
	require.NoError(t, err)

	intentKey, err := FinancialPaymentIntentStateKey(intent.PaymentID)
	require.NoError(t, err)
	canonicalIntent, err := BuildPaymentEnvelopeCanonicalRecord(PaymentEnvelopeCanonicalRecord{
		ObjectType:	PaymentEnvelopeIntent,
		ObjectID:	intent.PaymentID,
		StateKey:	intentKey,
		ObjectHash:	intent.IntentHash,
	})
	require.NoError(t, err)
	settlementKey, err := FinancialPaymentSettlementStateKey(settlement.PaymentID)
	require.NoError(t, err)
	canonicalSettlement, err := BuildPaymentEnvelopeCanonicalRecord(PaymentEnvelopeCanonicalRecord{
		ObjectType:	PaymentEnvelopeSettlement,
		ObjectID:	settlement.PaymentID,
		StateKey:	settlementKey,
		ObjectHash:	settlement.SettlementHash,
	})
	require.NoError(t, err)

	return financialPaymentStateFixtureSet{
		Intent:			intent,
		Channel:		routing.Channel,
		Condition:		condition,
		RouteCommitment:	routeCommitment,
		Settlement:		settlement,
		Dispute:		dispute,
		Receipt:		routing.Receipt,
		Proof:			routing.Proof,
		Fee:			fee,
		Message:		message,
		CanonicalIntent:	canonicalIntent,
		CanonicalSettlement:	canonicalSettlement,
	}
}
