package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestSDKServiceCallBuilderProducesEnvelopeRoutingAndSchema(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 50}
	descriptor := testInterfaceSystemDescriptor()

	result, err := BuildSDKServiceCall(SDKServiceCallBuildRequest{
		Context:	ctx,
		Descriptor:	descriptor,
		MethodID:	"submit",
		Caller:		coretypes.DefaultAuthority,
		Nonce:		1,
		PayloadHash:	testInterfaceHash("sdk/payload"),
		MaxFeeAmount:	"9",
		SignatureHash:	testInterfaceHash("sdk/signature"),
		TimeoutDelta:	12,
		IdempotencyKey:	"sdk-submit-idem",
		CallbackTarget:	"portable-service/callback",
	})
	require.NoError(t, err)
	require.Equal(t, result.Call.CallID, result.Envelope.CallID)
	require.Equal(t, result.Call.UnifiedCallHash, ComputeUnifiedServiceCallHash(result.Call))
	require.Equal(t, "submit", result.Schema.MethodID)
	require.Contains(t, result.Routing.Routes, UnifiedRouteServiceNetwork)
	require.NoError(t, result.ValidateForContext(ctx))
}

func TestServiceCallAnteValidationRejectsReplayAndUnknownRetry(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 50}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(20)
	require.NoError(t, err)

	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "ante/payload")
	ante, err := ValidateServiceCallAnte(ctx, descriptor, index, call)
	require.NoError(t, err)
	require.True(t, ante.Accept)
	require.True(t, ante.ReservePayment)
	require.True(t, ante.RequiresProof)
	require.NoError(t, ante.Validate())

	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)
	_, err = ValidateServiceCallAnte(ctx, descriptor, index, call)
	require.ErrorContains(t, err, "nonce already used")

	retry := testReceiptUnifiedCall(t, ctx, descriptor, 2, "ante/retry")
	retry.Kind = coretypes.ServiceCallKindRetry
	retry.RetryOf = testInterfaceHash("unknown/original")
	retry.UnifiedCallHash = ComputeUnifiedServiceCallHash(retry)
	_, err = ValidateServiceCallAnte(ctx, descriptor, index, retry)
	require.ErrorContains(t, err, "unknown original")
}

func TestReceiptAnchoringAndCallProofQuery(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 60}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(30)
	require.NoError(t, err)
	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "anchor/payload")

	anchor, err := AnchorUnifiedServiceReceipt(ctx, descriptor, index, call, coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("anchor/response"),
		ProofHash:	testInterfaceHash("anchor/proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	88,
		ProviderID:	"provider.storage",
	}, coretypes.DefaultAuthority)
	require.NoError(t, err)
	require.Equal(t, call.CallID, anchor.Receipt.CallID)
	require.Equal(t, anchor.AnchorHash, anchor.Msg.AnchorHash)
	require.NoError(t, anchor.Validate())
	require.True(t, anchor.ReplayIndex.ContainsCallID(call.CallID))

	proof, err := QueryServiceCallProofFromReplayIndex(anchor.ReplayIndex, []ServiceReceipt{anchor.Receipt}, QueryServiceCallProof{
		ServiceID:	descriptor.ServiceID,
		CallID:		call.CallID,
	}, ctx.Height)
	require.NoError(t, err)
	require.True(t, proof.Found)
	require.Equal(t, anchor.Receipt.ReceiptHash, proof.Proof.ReceiptHash)
	require.Equal(t, anchor.Tombstone.TombstoneHash, proof.Proof.TombstoneHash)
	require.Equal(t, anchor.ReplayIndex.IndexHash, proof.Proof.ReplayIndexHash)
	require.NoError(t, proof.Proof.Validate())

	missing, err := QueryServiceCallProofFromReplayIndex(anchor.ReplayIndex, nil, QueryServiceCallProof{
		ServiceID:	descriptor.ServiceID,
		CallID:		testInterfaceHash("missing/call"),
	}, ctx.Height)
	require.NoError(t, err)
	require.False(t, missing.Found)
}
