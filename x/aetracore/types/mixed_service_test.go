package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMixedServiceAnchorsResultAndSettlesAfterChallengeWindow(t *testing.T) {
	state := testMixedServiceState(t)
	state, anchor, err := AnchorMixedServiceResult(state, testMixedResultAnchor(10))
	require.NoError(t, err)
	require.Equal(t, state.ServiceID, anchor.ServiceID)
	require.Equal(t, uint64(30), anchor.ChallengeEndHeight)
	require.Equal(t, testHash("mixed/request/1"), anchor.RequestCommitment)
	require.Equal(t, testHash("mixed/result/1"), anchor.ResultCommitment)
	require.Equal(t, testHash("mixed/receipt/1"), anchor.ReceiptCommitment)
	require.NotEmpty(t, anchor.AnchorHash)

	_, _, err = SettleMixedServiceResult(state, anchor.AnchorID, 30)
	require.ErrorContains(t, err, "still in challenge window")

	state, settlement, err := SettleMixedServiceResult(state, anchor.AnchorID, 31)
	require.NoError(t, err)
	require.Equal(t, MixedSettlementReleased, settlement.Status)
	require.Equal(t, ServicePaymentStatusSettled, settlement.PaymentStatus)
	require.Equal(t, "5", settlement.PaymentAmount)
	require.NotEmpty(t, settlement.SettlementHash)
	require.NoError(t, state.Validate())
}

func TestMixedServiceChallengeCanPenalizeProvider(t *testing.T) {
	state := testMixedServiceState(t)
	state, anchor, err := AnchorMixedServiceResult(state, testMixedResultAnchor(20))
	require.NoError(t, err)

	state, dispute, err := OpenMixedServiceChallenge(state, MixedChallengeMessage{
		AnchorID:		anchor.AnchorID,
		Challenger:		"challenger.storage.1",
		ChallengeCommitment:	testHash("mixed/challenge/1"),
		OpenedHeight:		25,
		VerificationHook: MixedVerificationHook{
			HookType:		MixedHookProofVerification,
			ProofCommitment:	testHash("mixed/proof/1"),
			ProofMeterGas:		25_000,
			ExpectedResultHash:	testHash("mixed/expected/1"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, MixedDisputeOpen, dispute.Status)
	require.Equal(t, anchor.ChallengeEndHeight, dispute.ResolveByHeight)
	require.Equal(t, state.FallbackServiceID, dispute.VerificationHook.TargetServiceID)
	require.NotEmpty(t, dispute.DisputeHash)

	resolution := MixedDisputeResolution{
		DisputeID:	dispute.DisputeID,
		Resolver:	DefaultAuthority,
		ResolvedHeight:	29,
		ProofAccepted:	true,
	}
	resolution.ResolutionHash = ComputeMixedDisputeResolutionHash(resolution)
	state, settlement, err := ResolveMixedServiceChallenge(state, resolution)
	require.NoError(t, err)
	require.Equal(t, MixedSettlementPenalized, settlement.Status)
	require.Equal(t, ServicePaymentStatusRefunded, settlement.PaymentStatus)
	require.Equal(t, NativeFeePolicyID, settlement.PenaltyDenom)
	require.Equal(t, "25", settlement.PenaltyAmount)
	require.Equal(t, "challenger.storage.1", settlement.PenaltyRecipient)
	require.NoError(t, state.Validate())
}

func TestMixedServiceFallbackHookMustBeDeterministic(t *testing.T) {
	state := testMixedServiceState(t)
	state, anchor, err := AnchorMixedServiceResult(state, testMixedResultAnchor(20))
	require.NoError(t, err)

	_, _, err = OpenMixedServiceChallenge(state, MixedChallengeMessage{
		AnchorID:		anchor.AnchorID,
		Challenger:		"challenger.storage.2",
		ChallengeCommitment:	testHash("mixed/challenge/fallback"),
		OpenedHeight:		21,
		VerificationHook: MixedVerificationHook{
			HookType:	MixedHookFallbackExecution,
			Deterministic:	false,
		},
	})
	require.ErrorContains(t, err, "fallback verification hook must be deterministic")

	state, dispute, err := OpenMixedServiceChallenge(state, MixedChallengeMessage{
		AnchorID:		anchor.AnchorID,
		Challenger:		"challenger.storage.2",
		ChallengeCommitment:	testHash("mixed/challenge/fallback"),
		OpenedHeight:		21,
		VerificationHook: MixedVerificationHook{
			HookType:	MixedHookFallbackExecution,
			Deterministic:	true,
		},
	})
	require.NoError(t, err)

	resolution := MixedDisputeResolution{
		DisputeID:		dispute.DisputeID,
		Resolver:		DefaultAuthority,
		ResolvedHeight:		22,
		FallbackExecuted:	true,
	}
	resolution.ResolutionHash = ComputeMixedDisputeResolutionHash(resolution)
	state, settlement, err := ResolveMixedServiceChallenge(state, resolution)
	require.NoError(t, err)
	require.Equal(t, MixedSettlementFallback, settlement.Status)
	require.Equal(t, ServicePaymentStatusEscrowed, settlement.PaymentStatus)
	require.NoError(t, state.Validate())
}

func TestMixedServiceRejectsInsufficientProviderCollateral(t *testing.T) {
	descriptor := testMixedServiceWithCollateral()
	descriptor.Verification.ProviderCollateralAmount = "10"
	_, err := NewMixedServiceState(descriptor, "provider.storage.1", MixedFaultHigh, "25")
	require.ErrorContains(t, err, "must cover required amount")
}

func testMixedServiceState(t *testing.T) MixedServiceState {
	t.Helper()
	state, err := NewMixedServiceState(testMixedServiceWithCollateral(), "provider.storage.1", MixedFaultHigh, "25")
	require.NoError(t, err)
	require.NotEmpty(t, state.StateHash)
	require.NoError(t, state.Validate())
	return state
}

func testMixedServiceWithCollateral() ServiceDescriptor {
	descriptor := testMixedService("hybrid-storage-v2", ZoneIDApplication)
	descriptor.Verification.ProviderCollateralDenom = NativeFeePolicyID
	descriptor.Verification.ProviderCollateralAmount = "100"
	descriptor.Verification.FallbackServiceID = "storage-fallback"
	descriptor.Execution.FailureBehavior = ServiceFailureFallbackOnChain
	descriptor.Payment.Amount = "5"
	descriptor.Payment.EscrowID = "storage-escrow-v2"
	descriptor = CanonicalServiceDescriptor(descriptor)
	return descriptor
}

func testMixedResultAnchor(height uint64) MixedResultAnchor {
	return MixedResultAnchor{
		CallID:			testHash("mixed/call/1"),
		RequestCommitment:	testHash("mixed/request/1"),
		ResultCommitment:	testHash("mixed/result/1"),
		ReceiptCommitment:	testHash("mixed/receipt/1"),
		Height:			height,
	}
}
