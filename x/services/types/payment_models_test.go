package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestPerCallPaymentQuoteUsesFixedAmountAndTrustTiming(t *testing.T) {
	envelope := testPaymentModelEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentOnChain, "7")

	after, err := QuotePerCallPayment(envelope, coretypes.ServiceTrustConsensusExecuted)
	require.NoError(t, err)
	require.Equal(t, uint64(1), after.Units)
	require.Equal(t, "7", after.AmountDue)
	require.Equal(t, PaymentSettlementAfterExecution, after.SettlementTiming)
	require.NoError(t, after.Validate())

	prepaid := envelope
	prepaid.SettlementMode = coretypes.ServicePaymentPrepaid
	prepaid.EnvelopeHash = ComputePaymentEnvelopeHash(prepaid)
	before, err := QuotePerCallPayment(prepaid, coretypes.ServiceTrustFullyTrusted)
	require.NoError(t, err)
	require.Equal(t, PaymentSettlementBeforeExecution, before.SettlementTiming)
}

func TestPerBytePaymentRequiresDeterministicMeteringRecord(t *testing.T) {
	envelope := testPaymentModelEnvelope(t, coretypes.ServicePricingPerByte, coretypes.ServicePaymentMetered, "2")
	envelope.MeterIDOptional = "bytes-meter"
	envelope.MaxAmountOptional = "100"
	envelope.EnvelopeHash = ComputePaymentEnvelopeHash(envelope)

	meter, err := NewPaymentMeteringRecord(PaymentMeteringRecord{
		ServiceID:	envelope.PayeeService,
		CallID:		testInterfaceHash("call/bytes"),
		RequestBytes:	10,
		ResponseBytes:	20,
		StorageBytes:	5,
		MeterID:	"bytes-meter",
		MeterHeight:	30,
	})
	require.NoError(t, err)
	quote, err := QuotePerBytePayment(envelope, meter)
	require.NoError(t, err)
	require.Equal(t, uint64(35), quote.Units)
	require.Equal(t, "70", quote.AmountDue)
	require.True(t, quote.RequiresDeterministicMeterRecord)
	require.Equal(t, PaymentSettlementAfterExecution, quote.SettlementTiming)

	noBytes := meter
	noBytes.RequestBytes, noBytes.ResponseBytes, noBytes.StorageBytes = 0, 0, 0
	noBytes.RecordHash = ""
	_, err = NewPaymentMeteringRecord(noBytes)
	require.ErrorContains(t, err, "at least one byte")

	tooMuch := envelope
	tooMuch.MaxAmountOptional = "69"
	tooMuch.EnvelopeHash = ComputePaymentEnvelopeHash(tooMuch)
	_, err = QuotePerBytePayment(tooMuch, meter)
	require.ErrorContains(t, err, "exceeds max")
}

func TestPerComputeUnitPaymentRequiresSignedOrProofBackedUsageReceipt(t *testing.T) {
	envelope := testPaymentModelEnvelope(t, coretypes.ServicePricingPerComputeUnit, coretypes.ServicePaymentMetered, "3")
	envelope.MeterIDOptional = "compute-meter"
	envelope.MaxAmountOptional = "100"
	envelope.EnvelopeHash = ComputePaymentEnvelopeHash(envelope)

	_, err := NewPaymentUsageReceipt(PaymentUsageReceipt{
		ServiceID:	envelope.PayeeService,
		CallID:		testInterfaceHash("call/compute"),
		ComputeUnits:	8,
		ReceiptHeight:	40,
	})
	require.ErrorContains(t, err, "signature or proof")

	receipt, err := NewPaymentUsageReceipt(PaymentUsageReceipt{
		ServiceID:	envelope.PayeeService,
		CallID:		testInterfaceHash("call/compute"),
		ProviderID:	"provider-1",
		ComputeUnits:	8,
		ReceiptHeight:	40,
		SignedBy:	"provider-1",
		SignatureHash:	testInterfaceHash("compute/signature"),
	})
	require.NoError(t, err)
	quote, err := QuotePerComputeUnitPayment(envelope, receipt)
	require.NoError(t, err)
	require.Equal(t, uint64(8), quote.Units)
	require.Equal(t, "24", quote.AmountDue)
	require.True(t, quote.RequiresUsageReceipt)

	forged := receipt
	forged.ComputeUnits = 9
	require.ErrorContains(t, forged.Validate(), "hash mismatch")
}

func TestSubscriptionPaymentRequiresActiveEntitlement(t *testing.T) {
	envelope := testPaymentModelEnvelope(t, coretypes.ServicePricingSubscription, coretypes.ServicePaymentPrepaid, "50")
	entitlement, err := NewPaymentSubscriptionEntitlement(PaymentSubscriptionEntitlement{
		SubscriptionID:	"sub-1",
		Payer:		envelope.Payer,
		ServiceID:	envelope.PayeeService,
		StartHeight:	10,
		EndHeight:	20,
		StartUnix:	100,
		EndUnix:	200,
		StateBacked:	true,
	})
	require.NoError(t, err)
	quote, err := QuoteSubscriptionPayment(envelope, entitlement, 15, 150)
	require.NoError(t, err)
	require.Equal(t, "50", quote.AmountDue)
	require.True(t, quote.RequiresSubscriptionEntitlement)

	_, err = QuoteSubscriptionPayment(envelope, entitlement, 21, 150)
	require.ErrorContains(t, err, "not active")

	_, err = NewPaymentSubscriptionEntitlement(PaymentSubscriptionEntitlement{
		SubscriptionID:	"sub-2",
		Payer:		envelope.Payer,
		ServiceID:	envelope.PayeeService,
		StartHeight:	10,
		EndHeight:	20,
	})
	require.ErrorContains(t, err, "state or proof")
}

func TestEscrowPaymentSettlesAfterReceiptProofOrChallengeWindow(t *testing.T) {
	envelope := testPaymentModelEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentEscrow, "9")
	envelope.EscrowIDOptional = "escrow-1"
	envelope.EnvelopeHash = ComputePaymentEnvelopeHash(envelope)
	settlement, err := NewPaymentEscrowSettlement(PaymentEscrowSettlement{
		EscrowID:		"escrow-1",
		ServiceID:		envelope.PayeeService,
		ReceiptHeight:		30,
		ProofHeight:		33,
		ChallengeWindow:	7,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(40), settlement.SettleAfterHeight)

	quote, err := PlanEscrowPaymentSettlement(envelope, settlement)
	require.NoError(t, err)
	require.Equal(t, PaymentSettlementAfterChallenge, quote.SettlementTiming)
	require.True(t, quote.RequiresEscrowLock)
	require.Equal(t, uint64(40), quote.SettleAfterHeight)

	early := settlement
	early.SettleAfterHeight = 39
	early.SettlementHash = ComputePaymentEscrowSettlementHash(early)
	require.ErrorContains(t, early.Validate(), "challenge window")
}

func testPaymentModelEnvelope(t *testing.T, unit coretypes.ServicePricingUnit, mode coretypes.ServicePaymentSettlementMode, amount string) PaymentEnvelope {
	t.Helper()
	envelope, err := NewPaymentEnvelope(PaymentEnvelope{
		Asset:		coretypes.NativeFeePolicyID,
		Payer:		testPaymentPayer(),
		PayeeService:	"portable-service",
		Denom:		coretypes.NativeFeePolicyID,
		Amount:		amount,
		PricingUnit:	unit,
		SettlementMode:	mode,
		ExpiryHeight:	100,
		MeterIDOptional: func() string {
			if mode == coretypes.ServicePaymentMetered {
				return "meter-1"
			}
			return ""
		}(),
		EscrowIDOptional: func() string {
			if mode == coretypes.ServicePaymentEscrow {
				return "escrow-1"
			}
			return ""
		}(),
	})
	require.NoError(t, err)
	return envelope
}
