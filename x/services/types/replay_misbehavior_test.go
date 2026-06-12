package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServiceReplayProtectionProofBindsAllControlsAndTombstone(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 30}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "replay-proof/payload")

	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)
	proof, err := NewServiceReplayProtectionProof(ctx, descriptor, index, call, false)
	require.NoError(t, err)
	require.False(t, proof.ReceiptTombstone)
	require.Equal(t, call.Nonce, proof.Nonce)
	require.Equal(t, call.TargetService, proof.ServiceID)
	require.Equal(t, call.MethodID, proof.MethodID)
	require.Equal(t, call.IdempotencyKey, proof.IdempotencyKey)
	require.Equal(t, call.PayloadHash, proof.PayloadHash)
	require.Equal(t, call.DeadlineHeight, proof.DeadlineHeight)
	require.NoError(t, proof.Validate())

	_, err = NewServiceReplayProtectionProof(ctx, descriptor, index, call, true)
	require.ErrorContains(t, err, "tombstone")

	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("replay-proof/response"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	3,
		ProviderID:	"provider.storage",
		ExecutedHeight:	31,
		AnchoredHeight:	32,
	})
	require.NoError(t, err)
	index, tombstone, err := TombstoneServiceReceipt(ctx, index, call, receipt)
	require.NoError(t, err)
	proof, err = NewServiceReplayProtectionProof(ctx, descriptor, index, call, true)
	require.NoError(t, err)
	require.True(t, proof.ReceiptTombstone)
	require.Equal(t, tombstone.TombstoneHash, proof.TombstoneHash)
	require.Equal(t, index.IndexHash, proof.ReplayIndexHash)
	require.NoError(t, proof.Validate())

	tampered := proof
	tampered.PayloadHash = testInterfaceHash("other/payload")
	require.ErrorContains(t, tampered.Validate(), "control hash mismatch")
}

func TestServiceReplayProtectionProofRejectsMismatchedReplayEntry(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 30}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "replay-proof/mismatch")
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)

	mismatched := call
	mismatched.IdempotencyKey = "different-idem"
	mismatched.CallID = ""
	mismatched.CallID = coretypes.NormalizeServiceCall(ctx, mismatched.ToServiceCallEnvelope()).CallID
	mismatched.UnifiedCallHash = ComputeUnifiedServiceCallHash(mismatched)
	_, err = NewServiceReplayProtectionProof(ctx, descriptor, index, mismatched, false)
	require.ErrorContains(t, err, "does not match")
}

func TestProviderMisbehaviorReportInvalidResultPenalties(t *testing.T) {
	report, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		"portable-service",
		ProviderID:		"provider.storage",
		CallID:			testInterfaceHash("fault/invalid-result/call"),
		FaultClass:		ProviderFaultInvalidResult,
		EvidenceHash:		testInterfaceHash("fault/invalid-result/evidence"),
		ObservedHeight:		44,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyReputationScore, ProviderPenaltyCollateral},
		CollateralSlashAmount:	"25",
		ReputationDelta:	-15,
	})
	require.NoError(t, err)
	require.Equal(t, []ProviderPenaltySource{ProviderPenaltyCollateral, ProviderPenaltyReputationScore}, report.PenaltySources)
	require.NoError(t, report.Validate())

	noSlash := report
	noSlash.ReportHash = ""
	noSlash.CollateralSlashAmount = ""
	_, err = NewProviderMisbehaviorReport(noSlash)
	require.ErrorContains(t, err, "collateral slash")
}

func TestProviderMisbehaviorReportDeadlineFaults(t *testing.T) {
	_, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		"portable-service",
		ProviderID:		"provider.storage",
		CallID:			testInterfaceHash("fault/late/call"),
		FaultClass:		ProviderFaultLateResult,
		EvidenceHash:		testInterfaceHash("fault/late/evidence"),
		ObservedHeight:		50,
		DeadlineHeight:		50,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyEscrowedPayment},
		EscrowForfeitAmount:	"4",
	})
	require.ErrorContains(t, err, "after deadline")

	report, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		"portable-service",
		ProviderID:		"provider.storage",
		CallID:			testInterfaceHash("fault/missing/call"),
		FaultClass:		ProviderFaultMissingResult,
		EvidenceHash:		testInterfaceHash("fault/missing/evidence"),
		ObservedHeight:		51,
		DeadlineHeight:		50,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyEscrowedPayment},
		EscrowForfeitAmount:	"4",
	})
	require.NoError(t, err)
	require.NoError(t, report.Validate())
}

func TestProviderMisbehaviorReportInterfaceProofAndAvailabilityRules(t *testing.T) {
	_, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:			"portable-service",
		ProviderID:			"provider.storage",
		FaultClass:			ProviderFaultWrongInterfaceVersion,
		EvidenceHash:			testInterfaceHash("fault/interface/evidence"),
		ExpectedInterfaceHash:		testInterfaceHash("same/interface"),
		ObservedInterfaceHash:		testInterfaceHash("same/interface"),
		ObservedHeight:			61,
		PenaltySources:			[]ProviderPenaltySource{ProviderPenaltyServiceStake},
		ServiceStakeSlashAmount:	"10",
	})
	require.ErrorContains(t, err, "mismatched interface")

	proofReport, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		"portable-service",
		ProviderID:		"provider.storage",
		CallID:			testInterfaceHash("fault/proof/call"),
		FaultClass:		ProviderFaultInvalidProof,
		EvidenceHash:		testInterfaceHash("fault/proof/evidence"),
		ProofHash:		testInterfaceHash("fault/proof/proof"),
		ObservedHeight:		62,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyCollateral},
		CollateralSlashAmount:	"12",
	})
	require.NoError(t, err)
	require.NoError(t, proofReport.Validate())

	_, err = NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		"portable-service",
		ProviderID:		"provider.storage",
		FaultClass:		ProviderFaultAvailabilityViolation,
		EvidenceHash:		testInterfaceHash("fault/availability/evidence"),
		ObservedHeight:		63,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyCollateral},
		CollateralSlashAmount:	"1",
	})
	require.ErrorContains(t, err, "reputation")

	availability, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
		ServiceID:		"portable-service",
		ProviderID:		"provider.storage",
		FaultClass:		ProviderFaultAvailabilityViolation,
		EvidenceHash:		testInterfaceHash("fault/availability/evidence"),
		ObservedHeight:		63,
		PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyReputationScore},
		ReputationDelta:	-5,
	})
	require.NoError(t, err)
	require.NoError(t, availability.Validate())
}
