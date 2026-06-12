package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestDefaultXServiceProvidersModuleBreakdownCoversSection155(t *testing.T) {
	breakdown, err := DefaultXServiceProvidersModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, ServiceModuleProviders, breakdown.ModulePath)
	require.NotEmpty(t, breakdown.BreakdownHash)

	require.Contains(t, breakdown.StateObjects, XServiceProvidersStateProviderRecord)
	require.Contains(t, breakdown.StateObjects, XServiceProvidersStateProviderCollateral)
	require.Contains(t, breakdown.StateObjects, XServiceProvidersStateProviderReputation)
	require.Contains(t, breakdown.StateObjects, XServiceProvidersStateAvailabilityCommitment)
	require.Contains(t, breakdown.StateObjects, XServiceProvidersStateProviderFault)

	require.Contains(t, breakdown.Messages, XServiceProvidersMsgRegisterProvider)
	require.Contains(t, breakdown.Messages, XServiceProvidersMsgUpdateProvider)
	require.Contains(t, breakdown.Messages, XServiceProvidersMsgStakeProviderCollateral)
	require.Contains(t, breakdown.Messages, XServiceProvidersMsgUnstakeProviderCollateral)
	require.Contains(t, breakdown.Messages, XServiceProvidersMsgSubmitAvailabilityCommitment)
	require.Contains(t, breakdown.Messages, XServiceProvidersMsgSubmitProviderFault)

	require.Contains(t, breakdown.Queries, XServiceProvidersQueryProvider)
	require.Contains(t, breakdown.Queries, XServiceProvidersQueryProvidersByService)
	require.Contains(t, breakdown.Queries, XServiceProvidersQueryProviderCollateral)
	require.Contains(t, breakdown.Queries, XServiceProvidersQueryProviderReputation)
	require.Contains(t, breakdown.Queries, XServiceProvidersQueryAvailabilityCommitment)

	require.Contains(t, breakdown.IntegrationPoints, XServiceProvidersIntegrationServices)
	require.Contains(t, breakdown.IntegrationPoints, XServiceProvidersIntegrationServicePayments)
	require.Contains(t, breakdown.IntegrationPoints, XServiceProvidersIntegrationSlashingPenaltyRoute)
	require.Contains(t, breakdown.IntegrationPoints, XServiceProvidersIntegrationRoutingDiscovery)
}

func TestXServiceProvidersBreakdownRejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultXServiceProvidersModuleBreakdown()
	require.NoError(t, err)
	breakdown.Messages = removeXServiceProvidersMessageForTest(breakdown.Messages, XServiceProvidersMsgSubmitProviderFault)
	breakdown.BreakdownHash = ComputeXServiceProvidersModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "message")

	breakdown, err = DefaultXServiceProvidersModuleBreakdown()
	require.NoError(t, err)
	breakdown.FailureModes = breakdown.FailureModes[1:]
	breakdown.BreakdownHash = ComputeXServiceProvidersModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "failure")
}

func TestXServiceProvidersCollateralAndInterfaceGuards(t *testing.T) {
	provider := testServiceProviderRecord(t)
	collateral, err := NewProviderCollateral(provider, 10)
	require.NoError(t, err)
	require.NoError(t, ValidateProviderCollateralSufficient(collateral, "naet", "100"))
	require.ErrorContains(t, ValidateProviderCollateralSufficient(collateral, "naet", "101"), "insufficient")

	require.NoError(t, ValidateProviderAdvertisesInterface(provider, provider.Provider.SupportedInterfaces[0]))
	require.ErrorContains(t, ValidateProviderAdvertisesInterface(provider, testInterfaceHash("serviceproviders/unsupported-interface")), "unsupported interface")
}

func TestXServiceProvidersAvailabilityCommitmentGuardAndMessage(t *testing.T) {
	provider := testServiceProviderRecord(t)
	commitment, err := NewAvailabilityCommitment(provider, provider.Provider.SupportedInterfaces[0])
	require.NoError(t, err)
	require.NoError(t, ValidateAvailabilityCommitmentActive(commitment, 15))
	require.ErrorContains(t, ValidateAvailabilityCommitmentActive(commitment, 31), "expired")

	msg, err := NewMsgSubmitAvailabilityCommitment(coretypes.DefaultAuthority, commitment)
	require.NoError(t, err)
	require.Equal(t, ComputeMsgSubmitAvailabilityCommitmentHash(msg), msg.MessageHash)
}

func TestXServiceProvidersFaultProofAndReputationGuards(t *testing.T) {
	report, proof := testServiceProviderFaultProof(t)
	msg, err := NewMsgSubmitProviderFault(coretypes.DefaultAuthority, report, proof)
	require.NoError(t, err)
	require.Equal(t, ComputeMsgSubmitProviderFaultHash(msg), msg.MessageHash)
	require.NoError(t, ValidateProviderFaultProof(report, proof))

	tampered := proof
	tampered.ProviderID = "provider-2"
	tampered.FaultProofHash = ComputeServiceFaultProofHash(tampered)
	require.ErrorContains(t, ValidateProviderFaultProof(report, tampered), "invalid")

	reputation := ReputationRecord{ProviderID: report.ProviderID, Score: 100, Successes: 3, Failures: 1, UpdatedHeight: 20}
	reputation.RecordHash = coretypes.ComputeReputationRecordHash(reputation)
	next, err := ApplyDeterministicProviderReputationUpdate(reputation, report, 80, "")
	require.NoError(t, err)
	require.Equal(t, uint64(90), next.Score)
	require.Equal(t, uint64(2), next.Failures)
	require.ErrorContains(t, func() error {
		_, err := ApplyDeterministicProviderReputationUpdate(reputation, report, 80, testInterfaceHash("serviceproviders/wrong-reputation-hash"))
		return err
	}(), "not deterministic")
}

func TestXServiceProvidersQueries(t *testing.T) {
	provider := testServiceProviderRecord(t)
	collateral, err := NewProviderCollateral(provider, 10)
	require.NoError(t, err)
	commitment, err := NewAvailabilityCommitment(provider, provider.Provider.SupportedInterfaces[0])
	require.NoError(t, err)
	reputation := ReputationRecord{ProviderID: provider.Provider.ProviderID, Score: 50, UpdatedHeight: 12}
	reputation.RecordHash = coretypes.ComputeReputationRecordHash(reputation)
	state, err := BuildServiceProviderState([]ProviderRecord{provider}, []ProviderCollateral{collateral}, []ReputationRecord{reputation}, []AvailabilityCommitment{commitment}, 13)
	require.NoError(t, err)

	providerResp, err := QueryProviderFromState(state, QueryProvider{ProviderID: provider.Provider.ProviderID})
	require.NoError(t, err)
	require.True(t, providerResp.Found)
	collateralResp, err := QueryProviderCollateralFromState(state, QueryProviderCollateral{ProviderID: provider.Provider.ProviderID})
	require.NoError(t, err)
	require.True(t, collateralResp.Found)
	reputationResp, err := QueryProviderReputationFromState(state, QueryProviderReputation{ProviderID: provider.Provider.ProviderID})
	require.NoError(t, err)
	require.True(t, reputationResp.Found)
	commitmentResp, err := QueryAvailabilityCommitmentFromState(state, QueryAvailabilityCommitment{ProviderID: provider.Provider.ProviderID, ServiceID: provider.ServiceID})
	require.NoError(t, err)
	require.True(t, commitmentResp.Found)
}

func testServiceProviderRecord(t *testing.T) ProviderRecord {
	t.Helper()
	availability := coretypes.FogAvailabilityCommitment{
		EndpointHash:		testInterfaceHash("serviceproviders/endpoint"),
		WindowStart:		10,
		WindowEnd:		30,
		UptimeTargetBps:	9_500,
		RenewalNonce:		1,
		SignatureHash:		testInterfaceHash("serviceproviders/signature"),
	}
	availability.CommitmentHash = coretypes.ComputeFogAvailabilityCommitmentHash(availability)
	provider := coretypes.FogProviderRecord{
		ProviderID:		"provider-1",
		IdentityKey:		"provider-1-key",
		Category:		coretypes.FogCategoryCompute,
		Pricing:		coretypes.FogProviderPricing{Denom: "naet", Amount: "1", MaxAmount: "10", Unit: coretypes.FogPricingPerRequest},
		ReputationScore:	100,
		CollateralDenom:	"naet",
		CollateralAmount:	"100",
		StakeAmount:		"100",
		AvailabilityCommitment:	availability,
		SupportedInterfaces:	[]string{testInterfaceHash("serviceproviders/interface")},
		Status:			coretypes.FogProviderActive,
		RegisteredHeight:	10,
		UpdatedHeight:		10,
		ExpiryHeight:		40,
	}
	provider.ProviderHash = coretypes.ComputeFogProviderHash(provider)
	record := ProviderRecord{ServiceID: "fog-market", Provider: provider}
	record.RecordHash = coretypes.ComputeProviderRecordHash(record)
	require.NoError(t, record.Validate())
	return record
}

func testServiceProviderFaultProof(t *testing.T) (ProviderMisbehaviorReport, ServiceFaultProof) {
	t.Helper()
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	descriptor := testInterfaceSystemDescriptor()
	call := testInteractionCall(t, ctx, descriptor, "submit", 11, "serviceproviders/fault")
	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusFailed,
		ResponseHash:	testInterfaceHash("serviceproviders/fault/response"),
		ProofHash:	testInterfaceHash("serviceproviders/fault/proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusRefunded,
		ProviderID:	"provider-1",
		ExecutedHeight:	72,
		AnchoredHeight:	72,
		ErrorCode:	"invalid_result",
	})
	require.NoError(t, err)
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)
	_, tombstone, err := TombstoneServiceReceipt(ctx, index, call, receipt)
	require.NoError(t, err)
	report, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		descriptor.ServiceID,
		ProviderID:		"provider-1",
		CallID:			call.CallID,
		FaultClass:		ProviderFaultInvalidResult,
		EvidenceHash:		testInterfaceHash("serviceproviders/fault/evidence"),
		ObservedHeight:		72,
		DeadlineHeight:		80,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyCollateral, ProviderPenaltyReputationScore},
		CollateralSlashAmount:	"10",
		ReputationDelta:	-10,
	})
	require.NoError(t, err)
	proof, err := NewServiceFaultProof(report, receipt, tombstone, 73)
	require.NoError(t, err)
	return report, proof
}

func removeXServiceProvidersMessageForTest(messages []XServiceProvidersMessageName, target XServiceProvidersMessageName) []XServiceProvidersMessageName {
	out := make([]XServiceProvidersMessageName, 0, len(messages))
	for _, message := range messages {
		if message != target {
			out = append(out, message)
		}
	}
	return out
}
