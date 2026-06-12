package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalPaymentEnvelopeBuildsDeterministicIDAndSignature(t *testing.T) {
	alice := testAddress(0x61)
	bob := testAddress(0x62)
	condition := HashParts("invoice", "stream-1")

	first, err := BuildPayment(Payment{
		From:			" " + alice + " ",
		To:			bob,
		Amount:			"125",
		ConditionHash:		condition,
		Expiry:			100,
		RouteHint:		"route/a",
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	FinancialSettlementZoneID,
		Denom:			NativeDenom,
		FeeLimit:		"7",
		SettlementMode:		PaymentSettlementConditional,
		Nonce:			9,
	})
	require.NoError(t, err)
	require.NoError(t, first.Validate())
	require.Equal(t, first.PaymentID, ComputePaymentID(first))
	require.Equal(t, first.Signature, ComputePaymentSignatureHash(first))

	second, err := BuildPayment(Payment{
		From:			alice,
		To:			bob,
		Amount:			"125",
		ConditionHash:		condition,
		Expiry:			100,
		RouteHint:		"route/a",
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	FinancialSettlementZoneID,
		Denom:			NativeDenom,
		FeeLimit:		"7",
		SettlementMode:		PaymentSettlementConditional,
		Nonce:			9,
	})
	require.NoError(t, err)
	require.Equal(t, first.PaymentID, second.PaymentID)
	require.Equal(t, first.Signature, second.Signature)

	tampered := first
	tampered.Amount = "126"
	require.ErrorContains(t, tampered.Validate(), "payment id mismatch")
}

func TestPaymentAbstractionStateRejectsNonceReplayAndSettlesInFinancialZone(t *testing.T) {
	alice := testAddress(0x63)
	bob := testAddress(0x64)
	payment, err := BuildPayment(Payment{
		From:			alice,
		To:			bob,
		Amount:			"25",
		Expiry:			50,
		RouteHint:		"direct",
		SourceZone:		"IDENTITY_ZONE",
		DestinationZone:	FinancialSettlementZoneID,
		Denom:			NativeDenom,
		FeeLimit:		"2",
		SettlementMode:		PaymentSettlementZoneToZone,
		Nonce:			1,
	})
	require.NoError(t, err)

	state, stored, err := RegisterPayment(EmptyPaymentAbstractionState(10), payment, 10)
	require.NoError(t, err)
	require.Equal(t, payment.PaymentID, stored.PaymentID)
	require.Len(t, state.Nonces, 1)
	require.NoError(t, state.Validate())

	_, _, err = RegisterPayment(state, payment, 11)
	require.ErrorContains(t, err, "already exists")

	replay := payment
	replay.PaymentID = ""
	replay.Signature = ""
	replay.Amount = "26"
	_, _, err = RegisterPayment(state, replay, 11)
	require.ErrorContains(t, err, "sender nonce already used")

	state, receipt, err := SettlePaymentInFinancialZone(state, payment.PaymentID, 44, "1", 12)
	require.NoError(t, err)
	require.Equal(t, PaymentStatusSettled, receipt.Status)
	require.Equal(t, FinancialSettlementZoneID, receipt.SettlementZone)
	require.Equal(t, receipt.ReceiptHash, ComputePaymentAbstractionReceiptHash(receipt))
	require.NoError(t, state.Validate())

	badReceipt := receipt
	badReceipt.SettlementZone = "APPLICATION_ZONE"
	badReceipt.ReceiptHash = ""
	_, err = BuildPaymentAbstractionReceipt(badReceipt)
	require.ErrorContains(t, err, "financial zone")
}

func TestPaymentAbstractionValidationModesExpiryAndStateRoot(t *testing.T) {
	alice := testAddress(0x65)
	bob := testAddress(0x66)
	_, err := BuildPayment(Payment{
		From:			alice,
		To:			bob,
		Amount:			"1",
		Expiry:			20,
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	FinancialSettlementZoneID,
		SettlementMode:		PaymentSettlementConditional,
		Nonce:			1,
	})
	require.ErrorContains(t, err, "condition hash")

	first, err := BuildPayment(Payment{
		From:			alice,
		To:			bob,
		Amount:			"10",
		Expiry:			20,
		RouteHint:		"slow",
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	FinancialSettlementZoneID,
		SettlementMode:		PaymentSettlementStreaming,
		Nonce:			1,
	})
	require.NoError(t, err)
	second, err := BuildPayment(Payment{
		From:			bob,
		To:			alice,
		Amount:			"5",
		Expiry:			30,
		RouteHint:		"fast",
		SourceZone:		"CONTRACT_ZONE",
		DestinationZone:	FinancialSettlementZoneID,
		SettlementMode:		PaymentSettlementDirect,
		Nonce:			1,
	})
	require.NoError(t, err)

	left := EmptyPaymentAbstractionState(10)
	left, _, err = RegisterPayment(left, first, 10)
	require.NoError(t, err)
	left, _, err = RegisterPayment(left, second, 10)
	require.NoError(t, err)

	right := EmptyPaymentAbstractionState(10)
	right, _, err = RegisterPayment(right, second, 10)
	require.NoError(t, err)
	right, _, err = RegisterPayment(right, first, 10)
	require.NoError(t, err)
	require.Equal(t, ComputePaymentAbstractionStateRoot(left), ComputePaymentAbstractionStateRoot(right))

	expired, receipts, err := ExpirePayments(left, 21)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, first.PaymentID, receipts[0].PaymentID)
	require.Equal(t, PaymentStatusExpired, receipts[0].Status)
	require.NoError(t, expired.Validate())
}

func TestPaymentRouteFeeOptimizationIsDeterministicAndBounded(t *testing.T) {
	alice := testAddress(0x67)
	bob := testAddress(0x68)
	payment, err := BuildPayment(Payment{
		From:			alice,
		To:			bob,
		Amount:			"50",
		Expiry:			80,
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	FinancialSettlementZoneID,
		Denom:			NativeDenom,
		FeeLimit:		"5",
		SettlementMode:		PaymentSettlementZoneToZone,
		Nonce:			7,
	})
	require.NoError(t, err)

	quotes := []PaymentRouteQuote{
		{RouteHint: "expensive", SourceZone: "APPLICATION_ZONE", DestinationZone: FinancialSettlementZoneID, FeeAmount: "9", Capacity: "100", ExpiresHeight: 90},
		{RouteHint: "later", SourceZone: "APPLICATION_ZONE", DestinationZone: FinancialSettlementZoneID, FeeAmount: "3", Capacity: "100", ExpiresHeight: 95},
		{RouteHint: "best", SourceZone: "APPLICATION_ZONE", DestinationZone: FinancialSettlementZoneID, FeeAmount: "3", Capacity: "100", ExpiresHeight: 85},
		{RouteHint: "wrong-zone", SourceZone: "IDENTITY_ZONE", DestinationZone: FinancialSettlementZoneID, FeeAmount: "1", Capacity: "100", ExpiresHeight: 90},
		{RouteHint: "underfunded", SourceZone: "APPLICATION_ZONE", DestinationZone: FinancialSettlementZoneID, FeeAmount: "1", Capacity: "10", ExpiresHeight: 90},
	}
	best, err := SelectOptimalPaymentRoute(payment, quotes, 70)
	require.NoError(t, err)
	require.Equal(t, "best", best.RouteHint)
	require.NotEmpty(t, best.RouteScoreHash)
	require.NoError(t, best.Validate())

	payment.FeeLimit = "0"
	payment.PaymentID = ComputePaymentID(payment)
	payment.Signature = ComputePaymentSignatureHash(payment)
	best, err = SelectOptimalPaymentRoute(payment, quotes, 70)
	require.NoError(t, err)
	require.Equal(t, "underfunded", quotes[4].RouteHint)
	require.Equal(t, "best", best.RouteHint)
}
