package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestUnifiedInteractionClassesCoverServiceExecutionModes(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}

	onChain := testUnifiedOnChainDescriptor()
	onChainCall := testInteractionCall(t, ctx, onChain, "query", 1, "onchain")
	onChainPlan, err := BuildUnifiedInteractionPlan(ctx, onChain, onChainCall)
	require.NoError(t, err)
	require.Equal(t, InteractionOnChainTransaction, onChainPlan.InteractionClass)
	require.Contains(t, onChainPlan.Routes, UnifiedRouteDeliverTx)

	offChain := testUnifiedOffChainDescriptor()
	offChainCall := testInteractionCall(t, ctx, offChain, "submit", 2, "offchain")
	offChainPlan, err := BuildUnifiedInteractionPlan(ctx, offChain, offChainCall)
	require.NoError(t, err)
	require.Equal(t, InteractionOffChainServiceCall, offChainPlan.InteractionClass)
	require.Contains(t, offChainPlan.Routes, UnifiedRouteServiceNetwork)

	mixed := testInterfaceSystemDescriptor()
	mixedCall := testInteractionCall(t, ctx, mixed, "submit", 3, "mixed")
	mixedPlan, err := BuildUnifiedInteractionPlan(ctx, mixed, mixedCall)
	require.NoError(t, err)
	require.Equal(t, InteractionHybridExecutionFlow, mixedPlan.InteractionClass)
	require.Contains(t, mixedPlan.Routes, UnifiedRouteOnChainCommitment)

	eventedCall := testInteractionCall(t, ctx, mixed, "stream", 4, "evented")
	eventedPlan, err := BuildUnifiedInteractionPlan(ctx, mixed, eventedCall)
	require.NoError(t, err)
	require.Equal(t, InteractionEventedSubscription, eventedPlan.InteractionClass)

	retry := mixedCall
	retry.Kind = coretypes.ServiceCallKindRetry
	retry.RetryOf = mixedCall.CallID
	retry.Nonce = 5
	retry.IdempotencyKey = "mixed-retry-idem"
	retry.CallID = ""
	retry.CallID = coretypes.NormalizeServiceCall(ctx, retry.ToServiceCallEnvelope()).CallID
	retry.UnifiedCallHash = ComputeUnifiedServiceCallHash(retry)
	retryPlan, err := BuildUnifiedInteractionPlan(ctx, mixed, retry)
	require.NoError(t, err)
	require.Equal(t, InteractionRetry, retryPlan.InteractionClass)
}

func TestUnifiedServiceCallbackValidatesTargetInterfaceAndReplaySafety(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	originalDescriptor := testInterfaceSystemDescriptor()
	original := testInteractionCall(t, ctx, originalDescriptor, "submit", 7, "original")
	target := testCallbackTargetDescriptor()

	callback, err := NewUnifiedServiceCallback(
		ctx,
		original,
		target,
		"notify",
		testInterfaceHash("callback/payload"),
		"callback-prepaid:naet:0",
		35,
		8,
		"callback-idem",
	)
	require.NoError(t, err)
	require.Equal(t, original.CallID, callback.OriginalCallID)
	require.Equal(t, target.ServiceID, callback.CallbackTarget)
	require.Equal(t, "notify", callback.CallbackMethod)
	require.NoError(t, callback.ValidateForTarget(ctx, target))

	envelope := coretypes.NormalizeServiceCall(ctx, callback.ToServiceCallEnvelope(target))
	require.Equal(t, callback.CallbackCallID, envelope.CallID)
	require.Equal(t, coretypes.ServiceCallKindCallback, envelope.Kind)
	require.True(t, envelope.Callback)
	require.Equal(t, original.CallID, envelope.RetryOf)

	noID := callback
	noID.IdempotencyKey = ""
	noID.CallbackHash = ComputeUnifiedServiceCallbackHash(noID)
	require.ErrorContains(t, noID.ValidateForTarget(ctx, target), "replay-safe")

	wrongMethod := callback
	wrongMethod.CallbackMethod = "query"
	wrongMethod.CallbackCallID = coretypes.NormalizeServiceCall(ctx, wrongMethod.ToServiceCallEnvelope(target)).CallID
	wrongMethod.CallbackHash = ComputeUnifiedServiceCallbackHash(wrongMethod)
	require.ErrorContains(t, wrongMethod.ValidateForTarget(ctx, target), "does not support callbacks")
}

func TestUnifiedServiceCallbackExecutionEmitsReceipt(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	original := testInteractionCall(t, ctx, testInterfaceSystemDescriptor(), "submit", 9, "original")
	target := testCallbackTargetDescriptor()
	callback, err := NewUnifiedServiceCallback(ctx, original, target, "notify", testInterfaceHash("callback/payload"), "callback-prepaid:naet:0", 35, 10, "callback-idem")
	require.NoError(t, err)

	outcome := coretypes.ServiceExecutionOutcome{
		CallID:		callback.CallbackCallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("callback/response"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	11,
	}
	emission, err := EmitServiceCallbackReceipt(ctx, target, callback, outcome)
	require.NoError(t, err)
	require.Equal(t, callback.CallbackCallID, emission.Receipt.CallID)
	require.Equal(t, target.ServiceID, emission.Receipt.ServiceID)
	require.Equal(t, "notify", emission.Receipt.MethodID)
	require.Equal(t, coretypes.ServiceCallStatusExecuted, emission.Receipt.Status)
	require.NoError(t, emission.Validate())
}

func testInteractionCall(t *testing.T, ctx coretypes.ServiceConsensusContext, descriptor coretypes.ServiceDescriptor, methodID string, nonce uint64, seed string) UnifiedServiceCall {
	t.Helper()
	call, err := NewUnifiedServiceCall(
		ctx,
		descriptor,
		methodID,
		coretypes.DefaultAuthority,
		nonce,
		testInterfaceHash(seed+"/payload"),
		"9",
		testInterfaceHash(seed+"/signature"),
		10,
		seed+"-idem",
		"",
	)
	require.NoError(t, err)
	return call
}

func testCallbackTargetDescriptor() coretypes.ServiceDescriptor {
	descriptor := testUnifiedOnChainDescriptor()
	descriptor.ServiceID = "callback-target"
	descriptor.EndpointKey = "callback-target.endpoint"
	descriptor.Interface.InterfaceID = "l1.services.v1.CallbackTarget"
	descriptor.Interface.InterfaceName = "l1.services.v1.CallbackTarget"
	descriptor.Interface.Methods = []coretypes.ServiceMethodDescriptor{
		testInterfaceMethod("notify", coretypes.ServiceMethodAsync, coretypes.ServiceVerificationConsensusReceipt, coretypes.DefaultGasPolicy),
		testInterfaceMethod("query", coretypes.ServiceMethodSync, coretypes.ServiceVerificationConsensusReceipt, coretypes.DefaultGasPolicy),
	}
	descriptor.Interface.Methods[0].CallbackSupported = true
	descriptor.Interface.Methods[1].CallbackSupported = false
	descriptor.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(descriptor.Interface)
	descriptor.InterfaceID = descriptor.Interface.InterfaceID
	return coretypes.CanonicalServiceDescriptor(descriptor)
}
