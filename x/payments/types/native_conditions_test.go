package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativeConditionalPaymentPreimageResolvesLinkedActiveChain(t *testing.T) {
	alice := testAddress(0xa1)
	router := testAddress(0xa2)
	bob := testAddress(0xa3)
	routeID := HashParts("native-condition-route")
	preimage := "native-condition-secret"
	hashLock := HashParts(preimage)
	firstID := HashParts("native-condition-first")
	secondID := HashParts("native-condition-second")
	paymentStateRoot := HashParts("native-condition-payment-state")

	first, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:			firstID,
		Payer:				alice,
		Payee:				router,
		Amount:				"100",
		HashLock:			hashLock,
		TimeoutHeight:			100,
		RouteID:			routeID,
		NextConditionIDOptional:	secondID,
		Status:				NativeConditionalPaymentPending,
	})
	require.NoError(t, err)
	second, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:			secondID,
		Payer:				router,
		Payee:				bob,
		Amount:				"95",
		HashLock:			hashLock,
		TimeoutHeight:			80,
		RouteID:			routeID,
		PreviousConditionIDOptional:	firstID,
		Status:				NativeConditionalPaymentPending,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateNativeConditionTimeoutOrdering([]NativeConditionalPayment{second, first}, 10))

	resolved, outcomes, err := ResolveNativeConditionalPaymentChain([]NativeConditionalPayment{second, first}, preimage, 70, paymentStateRoot)
	require.NoError(t, err)
	require.Len(t, resolved, 2)
	require.Len(t, outcomes, 2)
	require.Equal(t, firstID, resolved[0].ConditionID)
	require.Equal(t, secondID, resolved[1].ConditionID)
	require.Equal(t, NativeConditionalPaymentResolved, resolved[0].Status)
	require.Equal(t, NativeConditionalPaymentResolved, resolved[1].Status)
	require.Equal(t, router, outcomes[0].Recipient)
	require.Equal(t, bob, outcomes[1].Recipient)
	require.Equal(t, hashLock, outcomes[0].PreimageHash)
	require.Equal(t, paymentStateRoot, outcomes[1].PaymentStateRoot)

	_, _, err = ResolveNativeConditionalPaymentChain([]NativeConditionalPayment{second, first}, "wrong", 70, paymentStateRoot)
	require.ErrorContains(t, err, "preimage")
}

func TestNativeConditionalPaymentTimeoutOrderingProtectsIntermediaries(t *testing.T) {
	alice := testAddress(0xa4)
	router := testAddress(0xa5)
	bob := testAddress(0xa6)
	routeID := HashParts("native-condition-timeout-route")
	firstID := HashParts("native-condition-timeout-first")
	secondID := HashParts("native-condition-timeout-second")

	upstream, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:			firstID,
		Payer:				alice,
		Payee:				router,
		Amount:				"50",
		TimeoutHeight:			100,
		RouteID:			routeID,
		NextConditionIDOptional:	secondID,
	})
	require.NoError(t, err)
	downstream, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:			secondID,
		Payer:				router,
		Payee:				bob,
		Amount:				"45",
		TimeoutHeight:			95,
		RouteID:			routeID,
		PreviousConditionIDOptional:	firstID,
	})
	require.NoError(t, err)
	require.ErrorContains(t, ValidateNativeConditionTimeoutOrdering([]NativeConditionalPayment{upstream, downstream}, 10), "protect")

	downstream.TimeoutHeight = 80
	downstream.ConditionRoot = ""
	downstream, err = BuildNativeConditionalPayment(downstream)
	require.NoError(t, err)
	require.NoError(t, ValidateNativeConditionTimeoutOrdering([]NativeConditionalPayment{downstream, upstream}, 10))
}

func TestNativeConditionalPaymentExpiryRefundsReservedLiquidity(t *testing.T) {
	alice := testAddress(0xa7)
	bob := testAddress(0xa8)
	stateRoot := HashParts("native-condition-expiry-state")
	condition, err := BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:	HashParts("native-condition-expiry"),
		Payer:		alice,
		Payee:		bob,
		Amount:		"33",
		TimeoutHeight:	40,
		RouteID:	HashParts("native-condition-expiry-route"),
	})
	require.NoError(t, err)

	refunded, outcome, err := ExpireNativeConditionalPayment(condition, bob, 41, stateRoot)
	require.NoError(t, err)
	require.Equal(t, NativeConditionalPaymentRefunded, refunded.Status)
	require.Equal(t, NativeConditionalPaymentRefunded, outcome.Status)
	require.Equal(t, alice, outcome.Recipient)
	require.Equal(t, "33", outcome.Amount)
	require.Equal(t, stateRoot, outcome.PaymentStateRoot)

	_, _, err = ExpireNativeConditionalPayment(condition, bob, 40, stateRoot)
	require.ErrorContains(t, err, "has not passed")
}

func TestCrossZonePaymentRoutingEnforcesCommittedRouteOrReservation(t *testing.T) {
	route := testMsgPaymentRoute()
	commitment := PaymentRouteCommitment{
		RouteID:	route.RouteID,
		Committer:	route.Payer,
		CommitmentHash:	ComputePaymentRouteCommitmentHash(route),
		Signed:		true,
		ExpiresHeight:	90,
	}
	input, err := BuildCrossZonePaymentRoutingInput(CrossZonePaymentRoutingInput{
		SourceAccount:		route.Payer,
		TargetAccount:		route.Payee,
		Amount:			route.Amount,
		MaxFee:			route.MaxFee,
		ExpiryHeight:		route.ExpiryHeight,
		RoutePolicy:		DefaultRoutePolicy(),
		LiquidityHints:		[]PaymentRouteBalance{{Participant: route.Payer, Available: "105"}},
		RouteCommitment:	commitment,
		UnifiedMessageRoot:	HashParts("cross-zone-unified-message-root"),
		FinancialFallbackRoot:	HashParts("cross-zone-financial-fallback-root"),
	})
	require.NoError(t, err)
	require.NoError(t, ValidateCrossZonePaymentRouteSettlement(input, route, 50))

	msg, err := BuildCrossZonePaymentMessage(CrossZonePaymentMessage{
		SourceZoneID:		"financial",
		DestinationZoneID:	"contract",
		SourceShardID:		1,
		DestinationShardID:	2,
		PayloadType:		"MsgPaymentRoute",
		RouteID:		route.RouteID,
		RouteCommitmentHash:	commitment.CommitmentHash,
		PaymentStateRoot:	input.RoutingRoot,
		UnifiedMessageRoot:	input.UnifiedMessageRoot,
		ExpiryHeight:		route.ExpiryHeight,
	})
	require.NoError(t, err)
	require.NoError(t, msg.Validate())
	require.NotEmpty(t, msg.MessageID)

	tampered := route
	tampered.Amount = "99"
	require.ErrorContains(t, ValidateCrossZonePaymentRouteSettlement(input, tampered, 50), "value")
}

func TestCrossZonePaymentReservedRouteRequiresOnChainReservationRoot(t *testing.T) {
	route := testMsgPaymentRoute()
	input, err := BuildCrossZonePaymentRoutingInput(CrossZonePaymentRoutingInput{
		SourceAccount:	route.Payer,
		TargetIdentity:	"merchant.aet",
		Amount:		route.Amount,
		MaxFee:		route.MaxFee,
		ExpiryHeight:	route.ExpiryHeight,
		RoutePolicy:	DefaultRoutePolicy(),
		RouteCommitment: PaymentRouteCommitment{
			RouteID:	route.RouteID,
			Committer:	route.Payer,
			CommitmentHash:	ComputePaymentRouteCommitmentHash(route),
			Reserved:	true,
			ExpiresHeight:	90,
		},
		UnifiedMessageRoot:	HashParts("cross-zone-reserved-message-root"),
		FinancialFallbackRoot:	HashParts("cross-zone-reserved-fallback-root"),
	})
	require.NoError(t, err)
	require.ErrorContains(t, ValidateCrossZonePaymentRouteSettlement(input, route, 50), "reservation root")

	input.ReservationRoot = HashParts("cross-zone-reservation-root")
	input.RoutingRoot = ""
	input, err = BuildCrossZonePaymentRoutingInput(input)
	require.NoError(t, err)
	require.NoError(t, ValidateCrossZonePaymentRouteSettlement(input, route, 50))
}
