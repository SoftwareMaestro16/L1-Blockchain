package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServicePipelinePrepareProcessFinalizeAndEndBlock(t *testing.T) {
	ctx := ServiceConsensusContext{ChainID: "aetra-test-1", Height: 10}
	state := servicePipelineState(t)

	identity, found := state.ServiceByID("identity-resolver")
	require.True(t, found)
	indexer, found := state.ServiceByID("indexer-feed")
	require.True(t, found)
	hybrid, found := state.ServiceByID("hybrid-storage")
	require.True(t, found)

	onchain := servicePipelineCall(ctx, identity, "resolve", ServiceCallKindOnChain, 1, "identity/accounts/caller", "1")
	offchain := servicePipelineCall(ctx, indexer, "query", ServiceCallKindOffChainReceipt, 1, "indexer/receipts/caller", "9")
	mixedDispute := servicePipelineCall(ctx, hybrid, "put", ServiceCallKindMixedDispute, 1, "storage/disputes/1", "5")
	mixedDispute.Dispute = true

	plan, err := PrepareServiceProposal(ctx, state, []ServiceCallEnvelope{offchain, onchain, mixedDispute})
	require.NoError(t, err)
	require.NoError(t, ProcessServiceProposal(ctx, state, plan))
	require.NoError(t, ValidateHash("plan hash", plan.PlanHash))
	require.NoError(t, ValidateHash("registry root", plan.RegistryRoot))
	require.NoError(t, ValidateHash("interface root", plan.InterfaceRoot))

	ordered := plan.OrderedCalls()
	require.Len(t, ordered, 3)
	require.Equal(t, mixedDispute.CallID, ordered[0].CallID)

	finalization, err := FinalizeServiceProposal(ctx, state, plan, []ServiceExecutionOutcome{
		{
			CallID:		onchain.CallID,
			Status:		ServiceCallStatusExecuted,
			ResponseHash:	testHash("identity-response"),
			PaymentStatus:	ServicePaymentStatusSettled,
			GasUsed:	100,
		},
		{
			CallID:		offchain.CallID,
			Status:		ServiceCallStatusAccepted,
			ResponseHash:	testHash("indexer-response"),
			PaymentStatus:	ServicePaymentStatusReserved,
			ProviderID:	"provider-indexer",
		},
		{
			CallID:		mixedDispute.CallID,
			Status:		ServiceCallStatusChallenged,
			ResponseHash:	testHash("challenge-response"),
			ProofHash:	testHash("challenge-proof"),
			PaymentStatus:	ServicePaymentStatusEscrowed,
			ProviderID:	"provider-storage",
		},
	})
	require.NoError(t, err)
	require.Len(t, finalization.Receipts, 3)
	require.NoError(t, ValidateHash("service receipts root", finalization.ServiceReceiptsRoot))

	maintenance, err := EndBlockServiceMaintenance(state, 125, finalization.Receipts, 10)
	require.NoError(t, err)
	require.Equal(t, []string{"identity-resolver", "indexer-feed"}, maintenance.ExpiredServiceIDs)
	require.Equal(t, uint64(2), maintenance.Metrics.ProviderReceiptCount)
	require.Len(t, maintenance.ReputationDeltas, 2)
	require.NoError(t, ValidateHash("maintenance hash", maintenance.MaintenanceHash))
}

func TestServicePipelineRejectsMalformedOrExpiredCalls(t *testing.T) {
	ctx := ServiceConsensusContext{ChainID: "aetra-test-1", Height: 10}
	state := servicePipelineState(t)
	identity, found := state.ServiceByID("identity-resolver")
	require.True(t, found)

	expired := servicePipelineCall(ctx, identity, "resolve", ServiceCallKindOnChain, 1, "identity/accounts/caller", "1")
	expired.DeadlineHeight = 9
	expired.CallID = ComputeServiceCallID(ctx, expired)
	_, err := PrepareServiceProposal(ctx, state, []ServiceCallEnvelope{expired})
	require.ErrorContains(t, err, "expired")

	missingID := servicePipelineCall(ctx, identity, "resolve", ServiceCallKindOnChain, 2, "identity/accounts/caller", "1")
	missingID.IdempotencyKey = ""
	missingID.CallID = ComputeServiceCallID(ctx, missingID)
	_, err = PrepareServiceProposal(ctx, state, []ServiceCallEnvelope{missingID})
	require.ErrorContains(t, err, "idempotency")

	badMethod := servicePipelineCall(ctx, identity, "resolve", ServiceCallKindOnChain, 3, "identity/accounts/caller", "1")
	badMethod.MethodID = "missing"
	badMethod.CallID = ComputeServiceCallID(ctx, badMethod)
	_, err = PrepareServiceProposal(ctx, state, []ServiceCallEnvelope{badMethod})
	require.ErrorContains(t, err, "not registered")
}

func TestServiceSTFRejectsNonDeterministicOrUnmeteredExecution(t *testing.T) {
	ctx := ServiceConsensusContext{ChainID: "aetra-test-1", Height: 10}
	state := servicePipelineState(t)
	hybrid, found := state.ServiceByID("hybrid-storage")
	require.True(t, found)
	call := servicePipelineCall(ctx, hybrid, "put", ServiceCallKindMixedSettlement, 1, "storage/settlements/1", "5")

	transition := ServiceStateTransition{
		CurrentStateRoot:	testHash("current-state"),
		NextStateRoot:		testHash("next-state"),
		Call:			call,
		Context:		ctx,
		StateReadSet:		[]string{"storage/settlements/1"},
		StateWriteSet:		[]string{"storage/settlements/1"},
		ExternalCalls:		[]string{"https://storage.aetra.local"},
		IterationLimit:		100,
		ProofVerificationGas:	50,
	}
	require.ErrorContains(t, transition.Validate(state), "external network calls")

	transition.ExternalCalls = nil
	transition.UsesWallClock = true
	require.ErrorContains(t, transition.Validate(state), "wall-clock")

	transition.UsesWallClock = false
	transition.LiveAvailabilityChecks = true
	require.ErrorContains(t, transition.Validate(state), "availability")

	transition.LiveAvailabilityChecks = false
	transition.ProofVerificationGas = 0
	require.ErrorContains(t, transition.Validate(state), "metered")

	transition.ProofVerificationGas = 50
	transition.DirectCrossZoneWrites = []string{"CONTRACT_ZONE/state"}
	require.ErrorContains(t, transition.Validate(state), "cross-zone")

	transition.DirectCrossZoneWrites = nil
	require.NoError(t, transition.Validate(state))
}

func servicePipelineState(t *testing.T) CoreState {
	t.Helper()
	state := EmptyState(TestnetParams())
	var err error
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
		testDescriptor(ZoneIDApplication, ZoneTypeApplication, "application"),
	} {
		state, err = RegisterZoneDescriptor(state, zone)
		require.NoError(t, err)
	}
	for _, service := range []ServiceDescriptor{
		testService("identity-resolver", ZoneIDIdentity),
		testOffChainService("indexer-feed", ZoneIDApplication),
		testMixedService("hybrid-storage", ZoneIDApplication),
	} {
		state, err = RegisterServiceDescriptor(state, service)
		require.NoError(t, err)
	}
	return state
}

func servicePipelineCall(ctx ServiceConsensusContext, descriptor ServiceDescriptor, methodID string, kind ServiceCallKind, nonce uint64, writeKey string, maxFee string) ServiceCallEnvelope {
	method, _ := descriptor.Interface.MethodByID(methodID)
	call := ServiceCallEnvelope{
		ServiceID:		descriptor.ServiceID,
		Caller:			DefaultAuthority,
		Nonce:			nonce,
		IdempotencyKey:		descriptor.ServiceID + "/" + methodID + "/idem",
		MethodID:		methodID,
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		PayloadHash:		testHash(descriptor.ServiceID + "/" + methodID + "/payload"),
		PaymentDenom:		descriptor.Payment.Denom,
		MaxFeeAmount:		maxFee,
		ProofRequirement:	method.VerificationModel,
		Kind:			kind,
		CreatedHeight:		ctx.Height,
		DeadlineHeight:		ctx.Height + 10,
		PriorityClass:		1,
		StateReadSet:		[]string{writeKey},
		StateWriteSet:		[]string{writeKey},
	}
	call.CallID = ComputeServiceCallID(ctx, call)
	return call
}
