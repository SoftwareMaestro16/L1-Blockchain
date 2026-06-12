package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServiceReceiptCanonicalViewMatchesCallReceiptFields(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 30}
	descriptor := testInterfaceSystemDescriptor()
	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "receipt/payload")
	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("receipt/response"),
		ProofHash:	testInterfaceHash("receipt/proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	44,
		ProviderID:	"provider.storage",
		ExecutedHeight:	31,
		AnchoredHeight:	32,
	})
	require.NoError(t, err)

	view, err := NewServiceReceiptCanonicalView(receipt)
	require.NoError(t, err)
	require.Equal(t, call.CallID, view.CallID)
	require.Equal(t, descriptor.ServiceID, view.ServiceID)
	require.Equal(t, "submit", view.MethodID)
	require.Equal(t, coretypes.DefaultAuthority, view.Caller)
	require.Equal(t, ServiceReceiptStatusExecuted, view.Status)
	require.Equal(t, call.PayloadHash, view.RequestHash)
	require.Equal(t, "settled", view.PaymentStatus)
	require.Equal(t, uint64(44), view.GasUsed)
	require.Equal(t, "provider.storage", view.ProviderID)
	require.Equal(t, receipt.ReceiptHash, view.ReceiptHash)
	require.NoError(t, view.Validate())

	for _, tc := range []struct {
		status	coretypes.ServiceCallStatus
		text	ServiceReceiptStatusText
	}{
		{coretypes.ServiceCallStatusAccepted, ServiceReceiptStatusAccepted},
		{coretypes.ServiceCallStatusExecuted, ServiceReceiptStatusExecuted},
		{coretypes.ServiceCallStatusFailed, ServiceReceiptStatusFailed},
		{coretypes.ServiceCallStatusExpired, ServiceReceiptStatusExpired},
		{coretypes.ServiceCallStatusChallenged, ServiceReceiptStatusChallenged},
		{coretypes.ServiceCallStatusSettled, ServiceReceiptStatusSettled},
		{coretypes.ServiceCallStatusReverted, ServiceReceiptStatusReverted},
	} {
		text, err := CanonicalServiceReceiptStatus(tc.status)
		require.NoError(t, err)
		require.Equal(t, tc.text, text)
	}
}

func TestServiceReplayIndexRejectsSameCallerNonceForService(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 30}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)

	call := testReceiptUnifiedCall(t, ctx, descriptor, 1, "replay/payload-a")
	require.Equal(t, ComputeUnifiedServiceCallID(ctx, descriptor.ServiceID, coretypes.DefaultAuthority, 1, call.IdempotencyKey, call.PayloadHash), call.CallID)
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)
	require.True(t, index.ContainsCallID(call.CallID))

	duplicateNonce := testReceiptUnifiedCall(t, ctx, descriptor, 1, "replay/payload-b")
	_, err = AcceptUnifiedServiceCall(ctx, descriptor, index, duplicateNonce)
	require.ErrorContains(t, err, "nonce already used")

	nextNonce := testReceiptUnifiedCall(t, ctx, descriptor, 2, "replay/payload-c")
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, nextNonce)
	require.NoError(t, err)
	require.Len(t, index.Entries, 2)
	require.NoError(t, index.Validate())
}

func TestServiceRetryReferencesOriginalAndTombstoneProofHorizon(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 40}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(10)
	require.NoError(t, err)

	original := testReceiptUnifiedCall(t, ctx, descriptor, 1, "retry/original")
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, original)
	require.NoError(t, err)
	receipt, err := coretypes.NewServiceCallReceipt(original.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		original.CallID,
		Status:		coretypes.ServiceCallStatusFailed,
		ResponseHash:	testInterfaceHash("retry/failed-response"),
		PaymentStatus:	coretypes.ServicePaymentStatusRefunded,
		GasUsed:	5,
		ProviderID:	"provider.storage",
		ExecutedHeight:	42,
		AnchoredHeight:	43,
		ErrorCode:	"temporary-failure",
	})
	require.NoError(t, err)
	index, tombstone, err := TombstoneServiceReceipt(ctx, index, original, receipt)
	require.NoError(t, err)
	require.Equal(t, uint64(53), tombstone.RetainUntilHeight)
	require.True(t, index.ContainsCallID(original.CallID))

	retained, err := PruneExpiredReceiptTombstones(index, tombstone.RetainUntilHeight)
	require.NoError(t, err)
	_, found := retained.TombstoneByCallID(original.CallID)
	require.True(t, found)

	pruned, err := PruneExpiredReceiptTombstones(index, tombstone.RetainUntilHeight+1)
	require.NoError(t, err)
	_, found = pruned.TombstoneByCallID(original.CallID)
	require.False(t, found)

	retry := testReceiptUnifiedCall(t, ctx, descriptor, 2, "retry/attempt")
	retry.Kind = coretypes.ServiceCallKindRetry
	retry.RetryOf = original.CallID
	retry.UnifiedCallHash = ComputeUnifiedServiceCallHash(retry)
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, retry)
	require.NoError(t, err)
	require.Equal(t, original.CallID, retry.ToServiceCallEnvelope().RetryOf)

	missingOriginal := testReceiptUnifiedCall(t, ctx, descriptor, 3, "retry/missing")
	missingOriginal.Kind = coretypes.ServiceCallKindRetry
	missingOriginal.RetryOf = testInterfaceHash("missing/original")
	missingOriginal.UnifiedCallHash = ComputeUnifiedServiceCallHash(missingOriginal)
	empty, err := NewServiceCallReplayIndex(10)
	require.NoError(t, err)
	_, err = AcceptUnifiedServiceCall(ctx, descriptor, empty, missingOriginal)
	require.ErrorContains(t, err, "unknown original")

	noIdempotency := retry
	noIdempotency.Nonce = 4
	noIdempotency.IdempotencyKey = ""
	noIdempotency.CallID = coretypes.NormalizeServiceCall(ctx, noIdempotency.ToServiceCallEnvelope()).CallID
	noIdempotency.UnifiedCallHash = ComputeUnifiedServiceCallHash(noIdempotency)
	_, err = AcceptUnifiedServiceCall(ctx, descriptor, index, noIdempotency)
	require.ErrorContains(t, err, "idempotency")
}

func testReceiptUnifiedCall(t *testing.T, ctx coretypes.ServiceConsensusContext, descriptor coretypes.ServiceDescriptor, nonce uint64, payloadSeed string) UnifiedServiceCall {
	t.Helper()
	call, err := NewUnifiedServiceCall(
		ctx,
		descriptor,
		"submit",
		coretypes.DefaultAuthority,
		nonce,
		testInterfaceHash(payloadSeed),
		"9",
		testInterfaceHash(payloadSeed+"/signature"),
		12,
		descriptor.ServiceID+"-submit-idem-"+payloadSeed,
		"",
	)
	require.NoError(t, err)
	return call
}
