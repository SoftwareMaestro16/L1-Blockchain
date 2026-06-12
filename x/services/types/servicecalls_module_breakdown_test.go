package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestDefaultXServiceCallsModuleBreakdownCoversSection153(t *testing.T) {
	breakdown, err := DefaultXServiceCallsModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, ServiceModuleCalls, breakdown.ModulePath)
	require.NotEmpty(t, breakdown.BreakdownHash)

	require.Contains(t, breakdown.StateObjects, XServiceCallsStateServiceCall)
	require.Contains(t, breakdown.StateObjects, XServiceCallsStateCallNonce)
	require.Contains(t, breakdown.StateObjects, XServiceCallsStateIdempotencyRecord)
	require.Contains(t, breakdown.StateObjects, XServiceCallsStateCallbackRecord)
	require.Contains(t, breakdown.StateObjects, XServiceCallsStateCallReceipt)

	require.Contains(t, breakdown.Messages, XServiceCallsMsgSubmitServiceCall)
	require.Contains(t, breakdown.Messages, XServiceCallsMsgAnchorServiceResult)
	require.Contains(t, breakdown.Messages, XServiceCallsMsgRetryServiceCall)
	require.Contains(t, breakdown.Messages, XServiceCallsMsgSubmitCallback)
	require.Contains(t, breakdown.Messages, XServiceCallsMsgExpireServiceCall)

	require.Contains(t, breakdown.Queries, XServiceCallsQueryServiceCall)
	require.Contains(t, breakdown.Queries, XServiceCallsQueryCallReceipt)
	require.Contains(t, breakdown.Queries, XServiceCallsQueryCallsByCaller)
	require.Contains(t, breakdown.Queries, XServiceCallsQueryCallProof)

	require.Contains(t, breakdown.IntegrationPoints, XServiceCallsIntegrationServices)
	require.Contains(t, breakdown.IntegrationPoints, XServiceCallsIntegrationServicePayments)
	require.Contains(t, breakdown.IntegrationPoints, XServiceCallsIntegrationServiceReceipts)
	require.Contains(t, breakdown.IntegrationPoints, XServiceCallsIntegrationABCIProposalHandling)
}

func TestXServiceCallsModuleBreakdownRejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultXServiceCallsModuleBreakdown()
	require.NoError(t, err)
	breakdown.Messages = removeXServiceCallsMessageForTest(breakdown.Messages, XServiceCallsMsgExpireServiceCall)
	breakdown.BreakdownHash = ComputeXServiceCallsModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "message")

	breakdown, err = DefaultXServiceCallsModuleBreakdown()
	require.NoError(t, err)
	breakdown.FailureModes = breakdown.FailureModes[1:]
	breakdown.BreakdownHash = ComputeXServiceCallsModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "failure")
}

func TestXServiceCallsMessagesAndNonceReplayGuard(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	descriptor := testInterfaceSystemDescriptor()
	index, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	call := testInteractionCall(t, ctx, descriptor, "submit", 1, "servicecalls/submit")

	submit, err := NewMsgSubmitServiceCall(coretypes.DefaultAuthority, call)
	require.NoError(t, err)
	require.Equal(t, ComputeMsgSubmitServiceCallHash(submit), submit.MessageHash)
	require.NoError(t, submit.ValidateBasic())

	_, err = ValidateServiceCallAnte(ctx, descriptor, index, call)
	require.NoError(t, err)
	index, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	require.NoError(t, err)
	_, err = ValidateServiceCallAnte(ctx, descriptor, index, call)
	require.ErrorContains(t, err, "nonce already used")

	expire, err := NewMsgExpireServiceCall(coretypes.DefaultAuthority, call, call.DeadlineHeight+1)
	require.NoError(t, err)
	require.NoError(t, expire.ValidateForCall(call))
	_, err = NewMsgExpireServiceCall(coretypes.DefaultAuthority, call, call.DeadlineHeight)
	require.ErrorContains(t, err, "after call deadline")
}

func TestXServiceCallsIdempotencyMisuseGuard(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	descriptor := testInterfaceSystemDescriptor()
	call := testInteractionCall(t, ctx, descriptor, "submit", 1, "servicecalls/idempotency")

	record, err := NewServiceCallIdempotencyRecord(call, nil)
	require.NoError(t, err)
	require.NoError(t, record.Validate())
	same, err := NewServiceCallIdempotencyRecord(call, []ServiceCallIdempotencyRecord{record})
	require.NoError(t, err)
	require.Equal(t, record.RecordHash, same.RecordHash)

	misuse := call
	misuse.Nonce = 2
	misuse.PayloadHash = testInterfaceHash("servicecalls/idempotency/other-payload")
	misuse.CallID = coretypes.NormalizeServiceCall(ctx, misuse.ToServiceCallEnvelope()).CallID
	misuse.UnifiedCallHash = ComputeUnifiedServiceCallHash(misuse)
	_, err = NewServiceCallIdempotencyRecord(misuse, []ServiceCallIdempotencyRecord{record})
	require.ErrorContains(t, err, "duplicate idempotency key misuse")
}

func TestXServiceCallsCallbackMismatchGuardAndMessage(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	original := testInteractionCall(t, ctx, testInterfaceSystemDescriptor(), "submit", 3, "servicecalls/original")
	target := testCallbackTargetDescriptor()
	callback, err := NewUnifiedServiceCallback(ctx, original, target, "notify", testInterfaceHash("servicecalls/callback/payload"), "callback-prepaid:naet:0", 85, 4, "callback-idem")
	require.NoError(t, err)

	msg, err := NewMsgSubmitCallback(coretypes.DefaultAuthority, callback)
	require.NoError(t, err)
	require.Equal(t, ComputeMsgSubmitCallbackHash(msg), msg.MessageHash)
	require.NoError(t, msg.ValidateBasic())

	wrongTarget := target
	wrongTarget.ServiceID = "wrong-callback-target"
	_, err = NewUnifiedServiceCallback(ctx, original, wrongTarget, "missing", testInterfaceHash("servicecalls/callback/wrong"), "callback-prepaid:naet:0", 85, 5, "callback-wrong-idem")
	require.ErrorContains(t, err, "not registered")

	tampered := callback
	tampered.CallbackTarget = "other-target"
	tampered.CallbackHash = ComputeUnifiedServiceCallbackHash(tampered)
	require.ErrorContains(t, tampered.ValidateForTarget(ctx, target), "target service mismatch")
}

func TestXServiceCallsAnchorWindowAndResultHashGuards(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	call := testInteractionCall(t, ctx, testInterfaceSystemDescriptor(), "submit", 5, "servicecalls/result")
	outcome := coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("servicecalls/result/response"),
		ProofHash:	testInterfaceHash("servicecalls/result/proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	10,
	}

	msg, err := NewMsgAnchorServiceResult(coretypes.DefaultAuthority, call, outcome.ResponseHash, outcome)
	require.NoError(t, err)
	require.NoError(t, msg.ValidateForCall(call))
	require.NoError(t, ValidateServiceCallResultHash(call, outcome.ResponseHash, outcome))

	_, err = NewMsgAnchorServiceResult(coretypes.DefaultAuthority, call, testInterfaceHash("servicecalls/result/wrong"), outcome)
	require.ErrorContains(t, err, "result hash mismatch")

	record, err := ValidateServiceCallResultAnchorWindow(call, call.DeadlineHeight)
	require.NoError(t, err)
	require.NoError(t, record.Validate())
	_, err = ValidateServiceCallResultAnchorWindow(call, call.DeadlineHeight+1)
	require.ErrorContains(t, err, "anchored late")
}

func TestXServiceCallsRetryAndQueries(t *testing.T) {
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	descriptor := testInterfaceSystemDescriptor()
	retryIndex, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	original := testInteractionCall(t, ctx, descriptor, "submit", 7, "servicecalls/original")
	_, err = AcceptUnifiedServiceCall(ctx, descriptor, retryIndex, original)
	require.NoError(t, err)

	retry := testInteractionCall(t, ctx, descriptor, "submit", 8, "servicecalls/retry")
	retry.Kind = coretypes.ServiceCallKindRetry
	retry.RetryOf = original.CallID
	retry.IdempotencyKey = "retry-idem"
	retry.CallID = coretypes.NormalizeServiceCall(ctx, retry.ToServiceCallEnvelope()).CallID
	retry.UnifiedCallHash = ComputeUnifiedServiceCallHash(retry)
	retryMsg, err := NewMsgRetryServiceCall(coretypes.DefaultAuthority, original, retry)
	require.NoError(t, err)
	require.NoError(t, retryMsg.ValidateForOriginal(original))

	anchorIndex, err := NewServiceCallReplayIndex(25)
	require.NoError(t, err)
	anchor, err := AnchorUnifiedServiceReceipt(ctx, descriptor, anchorIndex, original, coretypes.ServiceExecutionOutcome{
		CallID:		original.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("servicecalls/query/response"),
		ProofHash:	testInterfaceHash("servicecalls/query/proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	42,
	}, coretypes.DefaultAuthority)
	require.NoError(t, err)

	callResponse, err := QueryServiceCallFromCalls([]UnifiedServiceCall{original, retry}, QueryServiceCall{CallID: original.CallID})
	require.NoError(t, err)
	require.True(t, callResponse.Found)
	require.Equal(t, original.CallID, callResponse.Call.CallID)

	receiptResponse, err := QueryCallReceiptFromReceipts([]ServiceReceipt{anchor.Receipt}, QueryCallReceipt{CallID: original.CallID})
	require.NoError(t, err)
	require.True(t, receiptResponse.Found)
	require.Equal(t, original.CallID, receiptResponse.Receipt.CallID)

	callerResponse, err := QueryCallsByCallerFromCalls([]UnifiedServiceCall{retry, original}, QueryCallsByCaller{Caller: coretypes.DefaultAuthority})
	require.NoError(t, err)
	require.Equal(t, uint64(2), callerResponse.Total)
	require.NoError(t, callerResponse.Validate())

	callProof := QueryCallProof{ServiceID: descriptor.ServiceID, CallID: original.CallID}
	require.NoError(t, callProof.Validate())
	proof, err := QueryServiceCallProofFromReplayIndex(anchor.ReplayIndex, []ServiceReceipt{anchor.Receipt}, callProof.ToServiceCallProofQuery(), ctx.Height)
	require.NoError(t, err)
	require.True(t, proof.Found)
	require.NoError(t, proof.Proof.Validate())
}

func TestXServiceCallsABCIProposalContract(t *testing.T) {
	contract, err := DefaultServiceCallsABCIProposalContract()
	require.NoError(t, err)
	require.NoError(t, contract.Validate())
	require.True(t, contract.ClassifyByTargetService)
	require.True(t, contract.VerifySameSenderOrdering)
	require.True(t, contract.RejectExpiredCalls)
	require.True(t, contract.IncludeCallbacksAndRetries)
	require.True(t, contract.AnchorReceiptsInFinalizeBlock)

	contract.RejectExpiredCalls = false
	contract.ContractHash = ComputeServiceCallsABCIProposalContractHash(contract)
	require.ErrorContains(t, contract.Validate(), "ABCI++")
}

func removeXServiceCallsMessageForTest(messages []XServiceCallsMessageName, target XServiceCallsMessageName) []XServiceCallsMessageName {
	out := make([]XServiceCallsMessageName, 0, len(messages))
	for _, message := range messages {
		if message != target {
			out = append(out, message)
		}
	}
	return out
}
