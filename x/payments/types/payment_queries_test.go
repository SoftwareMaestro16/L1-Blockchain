package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentQueriesReturnHeightScopedProofFriendlyState(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	state := buildPaymentQueryState(t, fixture)

	intent, err := QueryPaymentIntentFromState(state, QueryPaymentIntent{PaymentID: fixture.Intent.PaymentID, Height: state.Height})
	require.NoError(t, err)
	require.True(t, intent.Found)
	require.Equal(t, fixture.Intent.ExpiryHeight, intent.ExpiryHeight)
	require.Equal(t, fixture.Intent.RouteIDOptional, intent.ProofMetadata.ProofMetadata[1][len("route:"):])
	require.NoError(t, intent.ProofMetadata.Validate())

	channel, err := QueryPaymentChannelFromState(state, QueryPaymentChannel{ChannelID: fixture.Channel.ChannelID, Height: state.Height})
	require.NoError(t, err)
	require.True(t, channel.Found)
	require.Equal(t, fixture.Channel.Participants, channel.Channel.Participants)
	require.Equal(t, fixture.Channel.Collateral, channel.Channel.Collateral)
	require.Equal(t, fixture.Channel.LatestNonce, channel.Channel.LatestNonce)
	require.Equal(t, fixture.Channel.SettlementStatus, channel.Channel.SettlementStatus)
	require.NoError(t, channel.ProofMetadata.Validate())

	condition, err := QueryConditionalPaymentFromState(state, QueryConditionalPayment{ConditionID: fixture.Condition.ConditionID, Height: state.Height})
	require.NoError(t, err)
	require.True(t, condition.Found)
	require.Equal(t, fixture.Condition.HashLock, condition.Condition.HashLock)
	require.Equal(t, fixture.Condition.TimeoutHeight, condition.Condition.TimeoutHeight)
	require.Equal(t, fixture.Condition.RouteID, condition.Condition.RouteID)
	require.Equal(t, fixture.Condition.Status, condition.Condition.Status)
	require.NoError(t, condition.ProofMetadata.Validate())

	route, err := QueryPaymentRouteFromState(state, QueryPaymentRoute{RouteID: fixture.RouteCommitment.RouteID, Height: state.Height})
	require.NoError(t, err)
	require.True(t, route.Found)
	require.Equal(t, fixture.RouteCommitment.CommitmentHash, route.Route.CommitmentHash)
	require.True(t, route.Metadata.Signed)
	require.Equal(t, fixture.RouteCommitment.ExpiresHeight, route.Metadata.ExpiresHeight)
	require.Equal(t, state.RouteRoot, route.Metadata.RouteRoot)
	require.Equal(t, "unreserved", route.Metadata.ReservationStatus)
	require.NoError(t, route.ProofMetadata.Validate())

	settlement, err := QueryPaymentSettlementFromState(state, QueryPaymentSettlement{PaymentID: fixture.Settlement.PaymentID, Height: state.Height})
	require.NoError(t, err)
	require.True(t, settlement.Found)
	require.Equal(t, fixture.Settlement.CloseStatus, settlement.Settlement.CloseStatus)
	require.Equal(t, fixture.Settlement.FinalStateHash, settlement.Settlement.FinalStateHash)
	require.Equal(t, fixture.Settlement.ReceiptHash, settlement.Settlement.ReceiptHash)
	require.NoError(t, settlement.ProofMetadata.Validate())

	dispute, err := QueryPaymentDisputeFromState(state, QueryPaymentDispute{DisputeID: fixture.Dispute.DisputeID, Height: state.Height})
	require.NoError(t, err)
	require.True(t, dispute.Found)
	require.Equal(t, fixture.Dispute.FraudProofHash, dispute.Dispute.FraudProofHash)
	require.Equal(t, fixture.Dispute.ChallengeEnd, dispute.Dispute.ChallengeEnd)
	require.Equal(t, fixture.Dispute.NewerStateHash, dispute.Dispute.NewerStateHash)
	require.Equal(t, fixture.Dispute.Status, dispute.Dispute.Status)
	require.NoError(t, dispute.ProofMetadata.Validate())
}

func TestPaymentProofQueryReturnsTypedObjectEnvelope(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	state := buildPaymentQueryState(t, fixture)

	for _, tc := range []struct {
		objectType	PaymentProofObjectType
		objectID	string
		stateKey	string
	}{
		{PaymentProofObjectIntent, fixture.Intent.PaymentID, FinancialPaymentIntentsPrefix + "/" + fixture.Intent.PaymentID},
		{PaymentProofObjectChannel, fixture.Channel.ChannelID, FinancialPaymentChannelsPrefix + "/" + fixture.Channel.ChannelID},
		{PaymentProofObjectCondition, fixture.Condition.ConditionID, FinancialPaymentConditionsPrefix + "/" + fixture.Condition.ConditionID},
		{PaymentProofObjectRoute, fixture.RouteCommitment.RouteID, FinancialPaymentRoutesPrefix + "/" + fixture.RouteCommitment.RouteID},
		{PaymentProofObjectSettlement, fixture.Settlement.PaymentID, FinancialPaymentSettlementsPrefix + "/" + fixture.Settlement.PaymentID},
		{PaymentProofObjectDispute, fixture.Dispute.DisputeID, FinancialPaymentDisputesPrefix + "/" + fixture.Dispute.DisputeID},
		{PaymentProofObjectReceipt, fixture.Receipt.PaymentID, FinancialPaymentReceiptsPrefix + "/" + fixture.Receipt.PaymentID},
	} {
		resp, err := QueryPaymentProofFromState(state, QueryPaymentProof{
			ObjectType:	tc.objectType,
			ObjectID:	tc.objectID,
			Height:		state.Height,
		})
		require.NoError(t, err)
		require.True(t, resp.Found, tc.objectType)
		require.Equal(t, tc.stateKey, resp.Proof.StateKey)
		require.Equal(t, state.PaymentRoot, resp.Proof.PaymentRoot)
		require.NoError(t, resp.Proof.Validate())
	}
}

func TestPaymentQueriesRejectUnavailableHeightAndMissingObjects(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	state := buildPaymentQueryState(t, fixture)

	_, err := QueryPaymentIntentFromState(state, QueryPaymentIntent{PaymentID: fixture.Intent.PaymentID, Height: state.Height - 1})
	require.ErrorContains(t, err, "outside available root history")

	missing, err := QueryPaymentProofFromState(state, QueryPaymentProof{
		ObjectType:	PaymentProofObjectReceipt,
		ObjectID:	HashParts("missing-receipt"),
		Height:		state.Height,
	})
	require.NoError(t, err)
	require.False(t, missing.Found)
}

func TestPaymentProofEnvelopeRejectsTampering(t *testing.T) {
	fixture := financialPaymentStateFixture(t)
	state := buildPaymentQueryState(t, fixture)
	resp, err := QueryPaymentProofFromState(state, QueryPaymentProof{
		ObjectType:	PaymentProofObjectSettlement,
		ObjectID:	fixture.Settlement.PaymentID,
		Height:		state.Height,
	})
	require.NoError(t, err)
	require.True(t, resp.Found)

	tampered := resp.Proof
	tampered.ObjectHash = HashParts("tampered-settlement")
	require.ErrorContains(t, tampered.Validate(), "proof hash mismatch")
}

func buildPaymentQueryState(t *testing.T, fixture financialPaymentStateFixtureSet) FinancialZonePaymentState {
	t.Helper()
	state, err := BuildFinancialZonePaymentState(FinancialZonePaymentState{
		Height:			121,
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
	return state
}
