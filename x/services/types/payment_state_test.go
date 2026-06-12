package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServicePaymentStateKeysMatchSpecification(t *testing.T) {
	modelKey, err := PaymentModelStateKey("svc")
	require.NoError(t, err)
	require.Equal(t, "services/payments/models/svc", modelKey)

	escrowKey, err := ServiceEscrowStateKey("escrow-1")
	require.NoError(t, err)
	require.Equal(t, "services/payments/escrow/escrow-1", escrowKey)

	streamKey, err := PaymentStreamStateKey("stream-1")
	require.NoError(t, err)
	require.Equal(t, "services/payments/streams/stream-1", streamKey)

	meterKey, err := MeteredUsageStateKey("meter-1")
	require.NoError(t, err)
	require.Equal(t, "services/payments/meters/meter-1", meterKey)

	settlementKey, err := PaymentSettlementStateKey(testInterfaceHash("call/settle"))
	require.NoError(t, err)
	require.Equal(t, "services/payments/settlements/"+testInterfaceHash("call/settle"), settlementKey)
}

func TestServicePaymentModelQueryAndEnvelopeRules(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.ZoneID = coretypes.ZoneIDFinancial
	descriptor.Payment.Denom = "uatom"

	model, err := NewServicePaymentModelFromDescriptor(descriptor)
	require.NoError(t, err)
	require.True(t, model.KnownBeforeSigning)
	require.Contains(t, model.SupportedDenoms, "uatom")
	require.Contains(t, model.SupportedDenoms, coretypes.NativeFeePolicyID)

	envelope, err := NewPaymentEnvelope(PaymentEnvelope{
		Asset:		"uatom",
		Payer:		testPaymentPayer(),
		PayeeService:	descriptor.ServiceID,
		Denom:		"uatom",
		Amount:		"1",
		PricingUnit:	coretypes.ServicePricingPerCall,
		SettlementMode:	coretypes.ServicePaymentPrepaid,
		ExpiryHeight:	90,
	})
	require.NoError(t, err)
	require.NoError(t, EnsureEnvelopeMatchesPaymentModel(envelope, model))

	badDenom := envelope
	badDenom.Asset = "factory/alice/token"
	badDenom.Denom = "factory/alice/token"
	badDenom.EnvelopeHash = ComputePaymentEnvelopeHash(badDenom)
	require.ErrorContains(t, EnsureEnvelopeMatchesPaymentModel(badDenom, model), "denom")

	state, err := BuildServicePaymentState([]ServicePaymentModel{model}, nil, nil, nil, nil, 10)
	require.NoError(t, err)
	query, err := QueryServicePaymentModelFromState(state, QueryPaymentModel{ServiceID: descriptor.ServiceID})
	require.NoError(t, err)
	require.True(t, query.Found)
	require.Equal(t, model.ModelHash, query.Model.ModelHash)
}

func TestServicePaymentEscrowSettlementAndProof(t *testing.T) {
	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentEscrow, "9")
	escrow, err := CreateServiceEscrowFromEnvelope(envelope, 12)
	require.NoError(t, err)

	escrowPlan, err := NewPaymentEscrowSettlement(PaymentEscrowSettlement{
		EscrowID:		envelope.EscrowIDOptional,
		ServiceID:		envelope.PayeeService,
		ReceiptHeight:		15,
		ProofHeight:		16,
		ChallengeWindow:	4,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(20), escrowPlan.SettleAfterHeight)

	quote, err := PlanEscrowPaymentSettlement(envelope, escrowPlan)
	require.NoError(t, err)
	settlement, err := SettlePaymentFromQuote(testInterfaceHash("call/escrow"), quote, coretypes.ServicePaymentStatusEscrowed, coretypes.ServiceFailureChallenge, 21)
	require.NoError(t, err)

	state, err := BuildServicePaymentState(nil, []ServiceEscrow{escrow}, nil, nil, []PaymentSettlement{settlement}, 22)
	require.NoError(t, err)
	require.NotEmpty(t, state.StateRoot)

	escrowKey, err := ServiceEscrowStateKey(escrow.EscrowID)
	require.NoError(t, err)
	proof, err := QueryServicePaymentProofFromState(state, QueryPaymentProof{Key: escrowKey})
	require.NoError(t, err)
	require.True(t, proof.Found)
	require.Equal(t, escrow.LockHash, proof.Proof.ValueHash)
	require.NoError(t, proof.Proof.Validate())
}

func TestServicePaymentStreamingState(t *testing.T) {
	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentStreaming, "2")
	stream, err := CreatePaymentStreamFromEnvelope(envelope, 10, 20)
	require.NoError(t, err)
	require.Equal(t, uint64(10), stream.PaidThrough)
	require.Equal(t, "2", stream.RatePerHeight)

	state, err := BuildServicePaymentState(nil, nil, []PaymentStream{stream}, nil, nil, 21)
	require.NoError(t, err)
	require.NoError(t, state.Validate())
}

func TestServicePaymentMeteredUsageRequiresReceiptSignatureOrProof(t *testing.T) {
	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerComputeUnit, coretypes.ServicePaymentMetered, "3")
	receipt, err := NewPaymentUsageReceipt(PaymentUsageReceipt{
		ServiceID:	envelope.PayeeService,
		CallID:		testInterfaceHash("call/metered"),
		ProviderID:	"provider-1",
		ComputeUnits:	5,
		ReceiptHeight:	30,
		SignedBy:	"provider-1",
		SignatureHash:	testInterfaceHash("usage/signature"),
	})
	require.NoError(t, err)

	quote, err := QuotePerComputeUnitPayment(envelope, receipt)
	require.NoError(t, err)
	require.Equal(t, "15", quote.AmountDue)
	usage, err := RecordMeteredUsageFromQuote(envelope.MeterIDOptional, quote, receipt, 31)
	require.NoError(t, err)
	require.Equal(t, receipt.ReceiptHash, usage.UsageReceipt.ReceiptHash)

	unsigned := receipt
	unsigned.SignedBy = ""
	unsigned.SignatureHash = ""
	unsigned.ProofHash = ""
	unsigned.ReceiptHash = ""
	_, err = NewPaymentUsageReceipt(unsigned)
	require.ErrorContains(t, err, "signature or proof")
}

func TestServicePaymentFailurePolicyAndFinancialRoute(t *testing.T) {
	require.Equal(t, coretypes.ServicePaymentStatusRefunded, PaymentStatusForFailurePolicy(coretypes.ServiceFailureRefund))
	require.Equal(t, coretypes.ServicePaymentStatusEscrowed, PaymentStatusForFailurePolicy(coretypes.ServiceFailureSlashProvider))
	require.Equal(t, coretypes.ServicePaymentStatusReserved, PaymentStatusForFailurePolicy(coretypes.ServiceFailureRetry))

	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentOnChain, "4")
	route, err := BuildFinancialZonePaymentRoute(envelope, "bank", "FINANCIAL_ZONE")
	require.NoError(t, err)
	require.Equal(t, envelope.PayeeService, route.ServiceID)
	require.Equal(t, envelope.Amount, route.Amount)
	require.NoError(t, route.Validate())
}

func testPaymentStateEnvelope(t *testing.T, unit coretypes.ServicePricingUnit, mode coretypes.ServicePaymentSettlementMode, amount string) PaymentEnvelope {
	t.Helper()
	envelope := PaymentEnvelope{
		Asset:		coretypes.NativeFeePolicyID,
		Payer:		testPaymentPayer(),
		PayeeService:	"portable-service",
		Denom:		coretypes.NativeFeePolicyID,
		Amount:		amount,
		PricingUnit:	unit,
		SettlementMode:	mode,
		ExpiryHeight:	100,
	}
	switch mode {
	case coretypes.ServicePaymentEscrow:
		envelope.EscrowIDOptional = "escrow-1"
	case coretypes.ServicePaymentStreaming:
		envelope.StreamIDOptional = "stream-1"
	case coretypes.ServicePaymentMetered:
		envelope.MeterIDOptional = "meter-1"
		envelope.MaxAmountOptional = "100"
	}
	out, err := NewPaymentEnvelope(envelope)
	require.NoError(t, err)
	return out
}
