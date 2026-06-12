package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestExecutionGuaranteesForOnChainService(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 40}
	descriptor := testUnifiedOnChainDescriptor()
	input := testExecutionGuaranteeInput(t, ctx, descriptor, 1, "guarantee/onchain", true, false)

	report, err := NewExecutionGuaranteeReport(input)
	require.NoError(t, err)
	require.True(t, report.OnChainPathsDeterministic)
	require.True(t, report.PaymentRulesKnownBeforeSigning)
	require.True(t, report.InterfaceHashKnownBeforeConstruction)
	require.True(t, report.AnchoredReceiptsDeterministic)
	require.False(t, report.EndpointAvailabilityBackedByCommitment)
	require.True(t, report.UIGenerationClientResponsibility)
	require.True(t, report.OffChainResultCorrectnessVerifiedOrSettled)
	require.NoError(t, report.Validate())
}

func TestExecutionGuaranteesForOffChainSignedReplayProtectedCall(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 40}
	descriptor := testUnifiedOffChainDescriptor()
	input := testExecutionGuaranteeInput(t, ctx, descriptor, 1, "guarantee/offchain", true, false)
	input.ResultProofVerified = true
	cache, err := NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		descriptor.ServiceID,
		DescriptorHash:		coretypes.ComputeServiceDescriptorHash(descriptor),
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		Source:			ServiceResolutionSignedCache,
		SignatureOptional:	testInterfaceHash("cache/signature"),
		ExpiresHeight:		90,
		FetchedAtHeight:	40,
		Trust:			ServiceDiscoveryCacheVerified,
	}, ServiceDiscoveryCacheConstraints{RegistryExpiryHeight: descriptor.ExpiryHeight, CurrentHeight: 40})
	require.NoError(t, err)
	input.DiscoveryCache = cache
	input.AvailabilityCommitmentHash = testInterfaceHash("provider/availability")

	report, err := NewExecutionGuaranteeReport(input)
	require.NoError(t, err)
	require.True(t, report.OffChainCallsSignedAndReplayProtected)
	require.True(t, report.OffChainResultCorrectnessVerifiedOrSettled)
	require.True(t, report.CachedDiscoveryAuthoritative)
	require.True(t, report.EndpointAvailabilityBackedByCommitment)

	tampered := input
	tampered.ReplayProof.PayloadHash = testInterfaceHash("wrong/payload")
	tampered.ReplayProof.ControlHash = ComputeServiceReplayProtectionProofHash(tampered.ReplayProof)
	_, err = NewExecutionGuaranteeReport(tampered)
	require.ErrorContains(t, err, "signed replay-protected")
}

func TestExecutionGuaranteesRejectOffChainCorrectnessBeforeProofOrChallengeSettlement(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 40}
	descriptor := testUnifiedOffChainDescriptor()
	input := testExecutionGuaranteeInput(t, ctx, descriptor, 1, "guarantee/offchain-unverified", true, false)

	_, err := NewExecutionGuaranteeReport(input)
	require.ErrorContains(t, err, "off-chain correctness")

	input.ChallengePeriodElapsed = true
	report, err := NewExecutionGuaranteeReport(input)
	require.NoError(t, err)
	require.True(t, report.OffChainResultCorrectnessVerifiedOrSettled)
}

func TestExecutionGuaranteesForMixedChallengeableServiceWithPenaltyRoute(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 40}
	descriptor := testInterfaceSystemDescriptor()
	input := testExecutionGuaranteeInput(t, ctx, descriptor, 1, "guarantee/mixed", true, true)
	input.ChallengePeriodElapsed = true

	report, err := NewExecutionGuaranteeReport(input)
	require.NoError(t, err)
	require.True(t, report.HybridResultsVerifiableOrChallengeable)
	require.True(t, report.ProviderMisbehaviorEconomicallyPenalized)
	require.NotEmpty(t, report.PenaltyRouteHash)
}

func TestExecutionGuaranteesRejectUnknownPaymentRulesAndInterfaceMismatch(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 40}
	descriptor := testUnifiedOnChainDescriptor()
	input := testExecutionGuaranteeInput(t, ctx, descriptor, 1, "guarantee/payment", true, false)
	input.PaymentModel.KnownBeforeSigning = false
	input.PaymentModel.ModelHash = ComputeServicePaymentModelHash(input.PaymentModel)

	_, err := NewExecutionGuaranteeReport(input)
	require.ErrorContains(t, err, "known before call signing")

	input = testExecutionGuaranteeInput(t, ctx, descriptor, 2, "guarantee/interface", true, false)
	input.Call.InterfaceHash = testInterfaceHash("wrong/interface")
	input.Call.UnifiedCallHash = ComputeUnifiedServiceCallHash(input.Call)
	_, err = NewExecutionGuaranteeReport(input)
	require.ErrorContains(t, err, "interface hash mismatch")
}

func TestExecutionNonGuaranteesKeepAdvisoryCacheAndAvailabilityExplicit(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 40}
	descriptor := testUnifiedOffChainDescriptor()
	input := testExecutionGuaranteeInput(t, ctx, descriptor, 1, "guarantee/cache-advisory", true, false)
	input.ResultProofVerified = true
	cache, err := NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		descriptor.ServiceID,
		DescriptorHash:		coretypes.ComputeServiceDescriptorHash(descriptor),
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		Source:			ServiceResolutionDistributedMesh,
		ExpiresHeight:		90,
		FetchedAtHeight:	40,
		Trust:			ServiceDiscoveryCacheAdvisory,
	}, ServiceDiscoveryCacheConstraints{RegistryExpiryHeight: descriptor.ExpiryHeight, CurrentHeight: 40})
	require.NoError(t, err)
	input.DiscoveryCache = cache

	report, err := NewExecutionGuaranteeReport(input)
	require.NoError(t, err)
	require.False(t, report.CachedDiscoveryAuthoritative)
	require.False(t, report.EndpointAvailabilityBackedByCommitment)
	require.True(t, report.UIGenerationClientResponsibility)
}

func testExecutionGuaranteeInput(t *testing.T, ctx coretypes.ServiceConsensusContext, descriptor coretypes.ServiceDescriptor, nonce uint64, seed string, tombstoneReceipt bool, withPenalty bool) ExecutionGuaranteeInput {
	t.Helper()
	call, err := NewUnifiedServiceCall(ctx, descriptor, descriptor.Interface.Methods[0].MethodID, coretypes.DefaultAuthority, nonce, testInterfaceHash(seed+"/payload"), "9", testInterfaceHash(seed+"/signature"), 10, seed+"-idem", "")
	require.NoError(t, err)
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)
	receipt := testExecutionGuaranteeReceipt(t, call, seed)
	if tombstoneReceipt {
		index, _, err = TombstoneServiceReceipt(ctx, index, call, receipt)
		require.NoError(t, err)
	}
	replayProof, err := NewServiceReplayProtectionProof(ctx, descriptor, index, call, tombstoneReceipt)
	require.NoError(t, err)
	paymentModel, err := NewServicePaymentModelFromDescriptor(descriptor)
	require.NoError(t, err)
	input := ExecutionGuaranteeInput{
		Context:	ctx,
		Descriptor:	descriptor,
		Call:		call,
		ReplayProof:	replayProof,
		Receipt:	receipt,
		PaymentModel:	paymentModel,
	}
	if withPenalty {
		report, err := NewProviderMisbehaviorReport(ProviderMisbehaviorReport{
			ServiceID:		descriptor.ServiceID,
			ProviderID:		"provider.storage",
			CallID:			call.CallID,
			FaultClass:		ProviderFaultInvalidProof,
			EvidenceHash:		testInterfaceHash(seed + "/fault-evidence"),
			ProofHash:		testInterfaceHash(seed + "/fault-proof"),
			ObservedHeight:		ctx.Height,
			PenaltySources:		[]ProviderPenaltySource{ProviderPenaltyCollateral},
			CollateralSlashAmount:	"10",
		})
		require.NoError(t, err)
		route, err := NewProviderPenaltyRoute(report, coretypes.NativeFeePolicyID, "treasury.services")
		require.NoError(t, err)
		input.PenaltyRoute = route
	}
	return input
}

func testExecutionGuaranteeReceipt(t *testing.T, call UnifiedServiceCall, seed string) ServiceReceipt {
	t.Helper()
	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash(seed + "/response"),
		ProofHash:	testInterfaceHash(seed + "/proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	3,
		ProviderID:	"provider.storage",
		ExecutedHeight:	call.CreatedHeight + 1,
		AnchoredHeight:	call.CreatedHeight + 2,
	})
	require.NoError(t, err)
	return receipt
}
