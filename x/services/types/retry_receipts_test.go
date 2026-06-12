package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServiceRetryPolicyBoundsAttemptsDeadlineAndPayment(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	descriptor := testInterfaceSystemDescriptor()
	original := testRetryCall(t, ctx, descriptor, 1, "original")
	retry := testRetryAttemptCall(t, ctx, descriptor, original, 2, "retry")
	policy, err := NewServiceRetryPolicy(ServiceRetryPolicy{
		MaxAttempts:		1,
		MaxDeadlineDelta:	20,
		PaymentPolicy:		RetryPaymentOriginalOnly,
	})
	require.NoError(t, err)

	attempt, err := NewServiceRetryAttempt(ctx, policy, original, retry, nil)
	require.NoError(t, err)
	require.False(t, attempt.ChargeAttempt)
	require.Equal(t, uint32(1), attempt.AttemptNumber)
	require.NoError(t, attempt.Validate())

	_, err = NewServiceRetryAttempt(ctx, policy, original, retry, []ServiceRetryAttempt{attempt})
	require.ErrorContains(t, err, "count")

	tooLate := retry
	tooLate.DeadlineHeight = original.CreatedHeight + policy.MaxDeadlineDelta + 1
	tooLate.TimeoutHeight = tooLate.DeadlineHeight - tooLate.CreatedHeight
	tooLate.CallID = ""
	tooLate.CallID = coretypes.NormalizeServiceCall(ctx, tooLate.ToServiceCallEnvelope()).CallID
	tooLate.UnifiedCallHash = ComputeUnifiedServiceCallHash(tooLate)
	_, err = NewServiceRetryAttempt(ctx, policy, original, tooLate, nil)
	require.ErrorContains(t, err, "deadline")

	noID := retry
	noID.IdempotencyKey = ""
	noID.CallID = ""
	noID.CallID = coretypes.NormalizeServiceCall(ctx, noID.ToServiceCallEnvelope()).CallID
	noID.UnifiedCallHash = ComputeUnifiedServiceCallHash(noID)
	_, err = NewServiceRetryAttempt(ctx, policy, original, noID, nil)
	require.ErrorContains(t, err, "idempotency")
}

func TestServiceRetryCanChargeAttemptsWhenPolicyAllows(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	descriptor := testInterfaceSystemDescriptor()
	original := testRetryCall(t, ctx, descriptor, 1, "original")
	retry := testRetryAttemptCall(t, ctx, descriptor, original, 2, "retry")
	policy, err := NewServiceRetryPolicy(ServiceRetryPolicy{
		MaxAttempts:		2,
		MaxDeadlineDelta:	20,
		PaymentPolicy:		RetryPaymentChargeAttempts,
	})
	require.NoError(t, err)

	attempt, err := NewServiceRetryAttempt(ctx, policy, original, retry, nil)
	require.NoError(t, err)
	require.True(t, attempt.ChargeAttempt)
	require.Equal(t, ComputeServiceRetryPaymentChargeHash(policy, original, retry, true), attempt.PaymentChargeHash)
}

func TestServiceRetryReceiptLinkReferencesOriginalCall(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	descriptor := testInterfaceSystemDescriptor()
	original := testRetryCall(t, ctx, descriptor, 1, "original")
	retry := testRetryAttemptCall(t, ctx, descriptor, original, 2, "retry")
	receipt := testReceiptForCall(t, ctx, retry, coretypes.ServiceCallStatusExecuted)

	link, err := NewServiceRetryReceiptLink(original, retry, receipt)
	require.NoError(t, err)
	require.Equal(t, original.CallID, link.OriginalCallID)
	require.Equal(t, retry.CallID, link.RetryCallID)
	require.NoError(t, link.Validate())

	wrongRetry := retry
	wrongRetry.RetryOf = testInterfaceHash("other/original")
	_, err = NewServiceRetryReceiptLink(original, wrongRetry, receipt)
	require.ErrorContains(t, err, "original")
}

func TestDeterministicReceiptRootsCommitAllReceiptClasses(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 20}
	onChain := testRetryCall(t, ctx, testUnifiedOnChainDescriptor(), 1, "onchain")
	offChain := testRetryCall(t, ctx, testUnifiedOffChainDescriptor(), 2, "offchain")
	mixed := testRetryCall(t, ctx, testInterfaceSystemDescriptor(), 3, "mixed")
	mixed.Kind = coretypes.ServiceCallKindMixedSettlement
	mixed.CallID = ""
	mixed.CallID = coretypes.NormalizeServiceCall(ctx, mixed.ToServiceCallEnvelope()).CallID
	mixed.UnifiedCallHash = ComputeUnifiedServiceCallHash(mixed)

	serviceReceipts := []ServiceReceipt{
		testReceiptForCall(t, ctx, mixed, coretypes.ServiceCallStatusSettled),
		testReceiptForCall(t, ctx, onChain, coretypes.ServiceCallStatusExecuted),
		testReceiptForCall(t, ctx, offChain, coretypes.ServiceCallStatusExecuted),
	}
	payment, err := NewReceiptCommitment("payment/"+mixed.CallID, testInterfaceHash("payment/receipt"))
	require.NoError(t, err)
	storage, err := NewReceiptCommitment("storage/"+mixed.CallID, testInterfaceHash("storage/receipt"))
	require.NoError(t, err)
	requirements := DeterministicReceiptRequirements{
		OnChainCallIDs:		[]string{onChain.CallID},
		OffChainResultCallIDs:	[]string{offChain.CallID},
		MixedSettlementCallIDs:	[]string{mixed.CallID},
	}

	roots, err := BuildDeterministicReceiptRoots(serviceReceipts, []ReceiptCommitment{payment}, []ReceiptCommitment{storage}, requirements)
	require.NoError(t, err)
	require.NotEmpty(t, roots.ServiceReceiptsRoot)
	require.NotEmpty(t, roots.CallReceiptsRoot)
	require.NotEmpty(t, roots.PaymentReceiptsRoot)
	require.NotEmpty(t, roots.StorageReceiptsRoot)
	require.NoError(t, roots.Validate())

	reordered, err := BuildDeterministicReceiptRoots([]ServiceReceipt{serviceReceipts[1], serviceReceipts[2], serviceReceipts[0]}, []ReceiptCommitment{payment}, []ReceiptCommitment{storage}, requirements)
	require.NoError(t, err)
	require.Equal(t, roots, reordered)

	missing := requirements
	missing.OffChainResultCallIDs = []string{testInterfaceHash("missing/offchain")}
	_, err = BuildDeterministicReceiptRoots(serviceReceipts, []ReceiptCommitment{payment}, []ReceiptCommitment{storage}, missing)
	require.ErrorContains(t, err, "missing")
}

func testRetryCall(t *testing.T, ctx coretypes.ServiceConsensusContext, descriptor coretypes.ServiceDescriptor, nonce uint64, seed string) UnifiedServiceCall {
	t.Helper()
	call, err := NewUnifiedServiceCall(ctx, descriptor, descriptor.Interface.Methods[0].MethodID, coretypes.DefaultAuthority, nonce, testInterfaceHash(seed+"/payload"), "9", testInterfaceHash(seed+"/signature"), 10, seed+"-idem", "")
	require.NoError(t, err)
	return call
}

func testRetryAttemptCall(t *testing.T, ctx coretypes.ServiceConsensusContext, descriptor coretypes.ServiceDescriptor, original UnifiedServiceCall, nonce uint64, seed string) UnifiedServiceCall {
	t.Helper()
	retry := testRetryCall(t, ctx, descriptor, nonce, seed)
	retry.Kind = coretypes.ServiceCallKindRetry
	retry.RetryOf = original.CallID
	retry.IdempotencyKey = seed + "-retry-idem"
	retry.CallID = ""
	retry.CallID = coretypes.NormalizeServiceCall(ctx, retry.ToServiceCallEnvelope()).CallID
	retry.UnifiedCallHash = ComputeUnifiedServiceCallHash(retry)
	require.NoError(t, ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, retry))
	return retry
}

func testReceiptForCall(t *testing.T, ctx coretypes.ServiceConsensusContext, call UnifiedServiceCall, status coretypes.ServiceCallStatus) ServiceReceipt {
	t.Helper()
	outcome := coretypes.NormalizeServiceExecutionOutcome(ctx, coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		status,
		ResponseHash:	testInterfaceHash(call.CallID + "/response"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	1,
	})
	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), outcome)
	require.NoError(t, err)
	return receipt
}
