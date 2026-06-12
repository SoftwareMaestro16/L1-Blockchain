package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestDefaultXServicePaymentsModuleBreakdownCoversSection154(t *testing.T) {
	breakdown, err := DefaultXServicePaymentsModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, ServiceModulePayments, breakdown.ModulePath)
	require.NotEmpty(t, breakdown.BreakdownHash)

	require.Contains(t, breakdown.StateObjects, XServicePaymentsStatePaymentModel)
	require.Contains(t, breakdown.StateObjects, XServicePaymentsStatePaymentEnvelope)
	require.Contains(t, breakdown.StateObjects, XServicePaymentsStateServiceEscrow)
	require.Contains(t, breakdown.StateObjects, XServicePaymentsStatePaymentStream)
	require.Contains(t, breakdown.StateObjects, XServicePaymentsStateMeteredUsage)
	require.Contains(t, breakdown.StateObjects, XServicePaymentsStatePaymentSettlement)

	require.Contains(t, breakdown.Messages, XServicePaymentsMsgSetServicePaymentModel)
	require.Contains(t, breakdown.Messages, XServicePaymentsMsgCreateServiceEscrow)
	require.Contains(t, breakdown.Messages, XServicePaymentsMsgSettleServiceEscrow)
	require.Contains(t, breakdown.Messages, XServicePaymentsMsgOpenPaymentStream)
	require.Contains(t, breakdown.Messages, XServicePaymentsMsgClosePaymentStream)
	require.Contains(t, breakdown.Messages, XServicePaymentsMsgSubmitMeteredUsage)

	require.Contains(t, breakdown.Queries, XServicePaymentsQueryPaymentModel)
	require.Contains(t, breakdown.Queries, XServicePaymentsQueryServiceEscrow)
	require.Contains(t, breakdown.Queries, XServicePaymentsQueryPaymentStream)
	require.Contains(t, breakdown.Queries, XServicePaymentsQueryMeteredUsage)
	require.Contains(t, breakdown.Queries, XServicePaymentsQueryPaymentSettlement)

	require.Contains(t, breakdown.IntegrationPoints, XServicePaymentsIntegrationBankOrFinancialZone)
	require.Contains(t, breakdown.IntegrationPoints, XServicePaymentsIntegrationServices)
	require.Contains(t, breakdown.IntegrationPoints, XServicePaymentsIntegrationServiceCalls)
	require.Contains(t, breakdown.IntegrationPoints, XServicePaymentsIntegrationPayments)
}

func TestXServicePaymentsModuleBreakdownRejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultXServicePaymentsModuleBreakdown()
	require.NoError(t, err)
	breakdown.Messages = removeXServicePaymentsMessageForTest(breakdown.Messages, XServicePaymentsMsgClosePaymentStream)
	breakdown.BreakdownHash = ComputeXServicePaymentsModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "message")

	breakdown, err = DefaultXServicePaymentsModuleBreakdown()
	require.NoError(t, err)
	breakdown.FailureModes = breakdown.FailureModes[1:]
	breakdown.BreakdownHash = ComputeXServicePaymentsModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "failure")
}

func TestXServicePaymentsModelEscrowMessagesAndUnderfundedGuard(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	model, err := NewServicePaymentModelFromDescriptor(descriptor)
	require.NoError(t, err)
	setModel, err := NewMsgSetServicePaymentModel(coretypes.DefaultAuthority, model)
	require.NoError(t, err)
	require.Equal(t, ComputeMsgSetServicePaymentModelHash(setModel), setModel.MessageHash)
	require.NoError(t, setModel.ValidateBasic())

	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentEscrow, "9")
	create, err := NewMsgCreateServiceEscrow(coretypes.DefaultAuthority, envelope, 12)
	require.NoError(t, err)
	require.Equal(t, envelope.EscrowIDOptional, create.Escrow.EscrowID)
	require.NoError(t, ValidateServiceEscrowFunding(create.Escrow, "9"))
	require.ErrorContains(t, ValidateServiceEscrowFunding(create.Escrow, "10"), "underfunded")

	quote, err := QuotePerCallPayment(envelope, coretypes.ServiceTrustHybridChallengeable)
	require.NoError(t, err)
	settlement, err := SettlePaymentFromQuote(testInterfaceHash("servicepayments/escrow/call"), quote, coretypes.ServicePaymentStatusEscrowed, coretypes.ServiceFailureChallenge, 15)
	require.NoError(t, err)
	settle, err := NewMsgSettleServiceEscrow(coretypes.DefaultAuthority, create.Escrow, settlement)
	require.NoError(t, err)
	require.NoError(t, settle.ValidateForEscrow(create.Escrow))
}

func TestXServicePaymentsUsageReceiptInvalidGuard(t *testing.T) {
	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerComputeUnit, coretypes.ServicePaymentMetered, "3")
	receipt, err := NewPaymentUsageReceipt(PaymentUsageReceipt{
		ServiceID:	envelope.PayeeService,
		CallID:		testInterfaceHash("servicepayments/metered/call"),
		ProviderID:	"provider-1",
		ComputeUnits:	5,
		ReceiptHeight:	30,
		SignedBy:	"provider-1",
		SignatureHash:	testInterfaceHash("servicepayments/metered/signature"),
	})
	require.NoError(t, err)
	quote, err := QuotePerComputeUnitPayment(envelope, receipt)
	require.NoError(t, err)
	usage, err := RecordMeteredUsageFromQuote(envelope.MeterIDOptional, quote, receipt, 31)
	require.NoError(t, err)
	msg, err := NewMsgSubmitMeteredUsage(coretypes.DefaultAuthority, usage)
	require.NoError(t, err)
	require.Equal(t, ComputeMsgSubmitMeteredUsageHash(msg), msg.MessageHash)

	invalid := usage
	invalid.CallID = testInterfaceHash("servicepayments/metered/wrong-call")
	invalid.UsageHash = ComputeMeteredUsageHash(invalid)
	_, err = NewMsgSubmitMeteredUsage(coretypes.DefaultAuthority, invalid)
	require.ErrorContains(t, err, "receipt mismatch")

	unsigned := receipt
	unsigned.SignedBy = ""
	unsigned.SignatureHash = ""
	unsigned.ProofHash = ""
	unsigned.ReceiptHash = ""
	_, err = NewPaymentUsageReceipt(unsigned)
	require.ErrorContains(t, err, "signature or proof")
}

func TestXServicePaymentsStreamSettlementLimit(t *testing.T) {
	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentStreaming, "2")
	open, err := NewMsgOpenPaymentStream(coretypes.DefaultAuthority, envelope, 10, 20)
	require.NoError(t, err)
	require.NoError(t, open.ValidateBasic())

	closeMsg, err := NewMsgClosePaymentStream(coretypes.DefaultAuthority, open.Stream, 15, "10")
	require.NoError(t, err)
	require.NoError(t, closeMsg.ValidateForStream(open.Stream))

	_, err = NewMsgClosePaymentStream(coretypes.DefaultAuthority, open.Stream, 15, "11")
	require.ErrorContains(t, err, "exceeds maximum")
}

func TestXServicePaymentsModelSnapshotRejectsChangedAfterSigning(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	descriptor := testInterfaceSystemDescriptor()
	call := testInteractionCall(t, ctx, descriptor, "submit", 1, "servicepayments/model")
	model, err := NewServicePaymentModelFromDescriptor(descriptor)
	require.NoError(t, err)
	snapshot, err := NewServicePaymentSignedModelSnapshot(model, call, ctx.Height)
	require.NoError(t, err)
	require.NoError(t, ValidatePaymentModelSnapshotForCall(snapshot, model))

	changed := model
	changed.UnitAmount = "2"
	changed.MaxAmountOptional = "2"
	changed.ModelHash = ""
	changed, err = NewServicePaymentModel(changed)
	require.NoError(t, err)
	require.ErrorContains(t, ValidatePaymentModelSnapshotForCall(snapshot, changed), "changed after call signing")
}

func TestXServicePaymentsQueriesAndFinancialIntegration(t *testing.T) {
	envelope := testPaymentStateEnvelope(t, coretypes.ServicePricingPerComputeUnit, coretypes.ServicePaymentMetered, "3")
	model, err := NewServicePaymentModel(ServicePaymentModel{
		ServiceID:		envelope.PayeeService,
		SupportedDenoms:	[]string{envelope.Denom},
		DefaultDenom:		envelope.Denom,
		PricingUnit:		envelope.PricingUnit,
		SettlementMode:		envelope.SettlementMode,
		UnitAmount:		envelope.Amount,
		MaxAmountOptional:	envelope.MaxAmountOptional,
		FailurePolicy:		coretypes.ServiceFailureRefund,
		KnownBeforeSigning:	true,
		UpdatedHeight:		10,
	})
	require.NoError(t, err)
	receipt, err := NewPaymentUsageReceipt(PaymentUsageReceipt{
		ServiceID:	envelope.PayeeService,
		CallID:		testInterfaceHash("servicepayments/query/call"),
		ProviderID:	"provider-1",
		ComputeUnits:	2,
		ReceiptHeight:	30,
		SignedBy:	"provider-1",
		SignatureHash:	testInterfaceHash("servicepayments/query/signature"),
	})
	require.NoError(t, err)
	quote, err := QuotePerComputeUnitPayment(envelope, receipt)
	require.NoError(t, err)
	usage, err := RecordMeteredUsageFromQuote(envelope.MeterIDOptional, quote, receipt, 31)
	require.NoError(t, err)
	settlement, err := SettlePaymentFromQuote(receipt.CallID, quote, coretypes.ServicePaymentStatusSettled, coretypes.ServiceFailureRefund, 32)
	require.NoError(t, err)
	state, err := BuildServicePaymentState([]ServicePaymentModel{model}, nil, nil, []MeteredUsage{usage}, []PaymentSettlement{settlement}, 33)
	require.NoError(t, err)

	modelResp, err := QueryServicePaymentModelFromState(state, QueryPaymentModel{ServiceID: envelope.PayeeService})
	require.NoError(t, err)
	require.True(t, modelResp.Found)
	meterResp, err := QueryMeteredUsageFromState(state, QueryMeteredUsage{MeterID: envelope.MeterIDOptional})
	require.NoError(t, err)
	require.True(t, meterResp.Found)
	settlementResp, err := QueryPaymentSettlementFromState(state, QueryPaymentSettlement{CallID: receipt.CallID})
	require.NoError(t, err)
	require.True(t, settlementResp.Found)

	route, err := BuildFinancialZonePaymentRoute(testPaymentStateEnvelope(t, coretypes.ServicePricingPerCall, coretypes.ServicePaymentOnChain, "4"), "bank", "FINANCIAL_ZONE")
	require.NoError(t, err)
	require.NoError(t, route.Validate())
}

func removeXServicePaymentsMessageForTest(messages []XServicePaymentsMessageName, target XServicePaymentsMessageName) []XServicePaymentsMessageName {
	out := make([]XServicePaymentsMessageName, 0, len(messages))
	for _, message := range messages {
		if message != target {
			out = append(out, message)
		}
	}
	return out
}
