package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServiceSecurityImplementationBundleLinksCollateralChallengeFaultPenaltyAndReplay(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 30}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "security-impl/payload")
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)

	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusChallenged,
		ResponseHash:	testInterfaceHash("security-impl/response"),
		ProofHash:	testInterfaceHash("security-impl/invalid-proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusEscrowed,
		GasUsed:	9,
		ProviderID:	"provider.storage",
		ExecutedHeight:	31,
		AnchoredHeight:	32,
		ErrorCode:	"invalid-proof",
	})
	require.NoError(t, err)
	index, tombstone, err := TombstoneServiceReceipt(ctx, index, call, receipt)
	require.NoError(t, err)
	replayProof, err := NewServiceReplayProtectionProof(ctx, descriptor, index, call, true)
	require.NoError(t, err)
	freshness, err := NewServiceReceiptFreshnessProof(tombstone, 40, true)
	require.NoError(t, err)
	require.False(t, freshness.Stale)

	report, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		descriptor.ServiceID,
		ProviderID:		"provider.storage",
		CallID:			call.CallID,
		FaultClass:		ProviderFaultInvalidProof,
		EvidenceHash:		testInterfaceHash("security-impl/evidence"),
		ProofHash:		testInterfaceHash("security-impl/invalid-proof"),
		ObservedHeight:		33,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyCollateral, ProviderPenaltyReputationScore},
		CollateralSlashAmount:	"25",
		ReputationDelta:	-20,
	})
	require.NoError(t, err)
	faultProof, err := NewServiceFaultProof(report, receipt, tombstone, 34)
	require.NoError(t, err)
	require.Equal(t, report.ReportHash, faultProof.ReportHash)

	route, err := NewProviderPenaltyRoute(report, coretypes.NativeFeePolicyID, "treasury.services")
	require.NoError(t, err)
	require.Len(t, route.Entries, 2)

	dispute, err := coretypes.NewMsgSubmitServiceDispute(descriptor.Owner, descriptor.ServiceID, call.CallID, "provider.storage", report.EvidenceHash, string(report.FaultClass), 35)
	require.NoError(t, err)
	flow, err := NewServiceChallengeFlow(descriptor, dispute, faultProof, route)
	require.NoError(t, err)
	require.Equal(t, uint64(45), flow.ChallengeEndHeight)

	collateral, err := coretypes.NewMsgStakeProviderCollateral(descriptor.Owner, descriptor.ServiceID, "provider.storage", coretypes.NativeFeePolicyID, "100", 29)
	require.NoError(t, err)
	bundle, err := NewServiceSecurityImplementationBundle(descriptor, collateral, flow, faultProof, route, replayProof, freshness)
	require.NoError(t, err)
	require.Equal(t, ServiceTrustLabelHybridChallengeable, bundle.TrustModelLabel)
	require.Equal(t, ServiceFailureLabelChallenge, bundle.ExecutionFailureBehaviorLabel)
	require.NoError(t, bundle.Validate())
}

func TestServiceFaultProofRejectsInvalidProofWithoutProofHash(t *testing.T) {
	_, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		"portable-service",
		ProviderID:		"provider.storage",
		CallID:			testInterfaceHash("security-impl/invalid-proof/call"),
		FaultClass:		ProviderFaultInvalidProof,
		EvidenceHash:		testInterfaceHash("security-impl/invalid-proof/evidence"),
		ObservedHeight:		44,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyCollateral},
		CollateralSlashAmount:	"10",
	})
	require.ErrorContains(t, err, "proof hash")
}

func TestServiceSecurityImplementationRejectsDuplicateCalls(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 30}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "security-impl/duplicate-a")
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)

	duplicate := testReceiptUnifiedCall(t, ctx, descriptor, 1, "security-impl/duplicate-b")
	_, err = AcceptUnifiedServiceCall(ctx, descriptor, index, duplicate)
	require.ErrorContains(t, err, "nonce already used")
}

func TestServiceReceiptFreshnessProofRejectsStaleReceipts(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 30}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(5)
	require.NoError(t, err)
	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "security-impl/stale")
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)
	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("security-impl/stale-response"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	1,
		ProviderID:	"provider.storage",
		ExecutedHeight:	31,
		AnchoredHeight:	32,
	})
	require.NoError(t, err)
	_, tombstone, err := TombstoneServiceReceipt(ctx, index, call, receipt)
	require.NoError(t, err)

	_, err = NewServiceReceiptFreshnessProof(tombstone, tombstone.RetainUntilHeight+1, true)
	require.ErrorContains(t, err, "stale")

	stale, err := NewServiceReceiptFreshnessProof(tombstone, tombstone.RetainUntilHeight+1, false)
	require.NoError(t, err)
	require.True(t, stale.Stale)
	require.NoError(t, stale.Validate())
}
