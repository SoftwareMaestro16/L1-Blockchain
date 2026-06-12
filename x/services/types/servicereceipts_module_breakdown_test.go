package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestDefaultXServiceReceiptsModuleBreakdownCoversSection156(t *testing.T) {
	breakdown, err := DefaultXServiceReceiptsModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, ServiceModuleReceipts, breakdown.ModulePath)
	require.NotEmpty(t, breakdown.BreakdownHash)

	require.Contains(t, breakdown.StateObjects, XServiceReceiptsStateReceiptRecord)
	require.Contains(t, breakdown.StateObjects, XServiceReceiptsStateReceiptRoot)
	require.Contains(t, breakdown.StateObjects, XServiceReceiptsStateReceiptTombstone)
	require.Contains(t, breakdown.StateObjects, XServiceReceiptsStateReceiptParams)

	require.Contains(t, breakdown.Messages, XServiceReceiptsMsgAnchorReceipt)
	require.Contains(t, breakdown.Messages, XServiceReceiptsMsgPruneReceipt)

	require.Contains(t, breakdown.Queries, XServiceReceiptsQueryReceipt)
	require.Contains(t, breakdown.Queries, XServiceReceiptsQueryReceiptsByService)
	require.Contains(t, breakdown.Queries, XServiceReceiptsQueryReceiptRoot)
	require.Contains(t, breakdown.Queries, XServiceReceiptsQueryReceiptProof)

	require.Contains(t, breakdown.IntegrationPoints, XServiceReceiptsIntegrationAllServiceModules)
	require.Contains(t, breakdown.IntegrationPoints, XServiceReceiptsIntegrationProofRegistry)
	require.Contains(t, breakdown.IntegrationPoints, XServiceReceiptsIntegrationStoreV2)
}

func TestXServiceReceiptsBreakdownRejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultXServiceReceiptsModuleBreakdown()
	require.NoError(t, err)
	breakdown.Messages = removeXServiceReceiptsMessageForTest(breakdown.Messages, XServiceReceiptsMsgPruneReceipt)
	breakdown.BreakdownHash = ComputeXServiceReceiptsModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "message")

	breakdown, err = DefaultXServiceReceiptsModuleBreakdown()
	require.NoError(t, err)
	breakdown.FailureModes = breakdown.FailureModes[1:]
	breakdown.BreakdownHash = ComputeXServiceReceiptsModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "failure")
}

func TestXServiceReceiptsAnchorDuplicateAndHashMismatchGuards(t *testing.T) {
	receipt, _ := testServiceReceiptForReceipts(t, 21)
	msg, err := NewMsgAnchorReceipt(coretypes.DefaultAuthority, receipt, receipt.ReceiptHash)
	require.NoError(t, err)
	records, record, err := AnchorReceiptRecord(nil, msg, ServiceReceiptRecordKindService, 25)
	require.NoError(t, err)
	require.NoError(t, record.Validate())
	require.Len(t, records, 1)

	_, _, err = AnchorReceiptRecord(records, msg, ServiceReceiptRecordKindService, 25)
	require.ErrorContains(t, err, "duplicate receipt")

	_, err = NewMsgAnchorReceipt(coretypes.DefaultAuthority, receipt, testInterfaceHash("servicereceipts/wrong-hash"))
	require.ErrorContains(t, err, "hash mismatch")
}

func TestXServiceReceiptsPruneProofHorizonGuard(t *testing.T) {
	receipt, _ := testServiceReceiptForReceipts(t, 30)
	msg, err := NewMsgAnchorReceipt(coretypes.DefaultAuthority, receipt, receipt.ReceiptHash)
	require.NoError(t, err)
	records, record, err := AnchorReceiptRecord(nil, msg, ServiceReceiptRecordKindService, 10)
	require.NoError(t, err)
	params, err := NewReceiptParams(10, 100)
	require.NoError(t, err)

	early, err := NewMsgPruneReceipt(coretypes.DefaultAuthority, record, record.RetainUntilHeight-1)
	require.NoError(t, err)
	_, err = PruneReceiptRecord(records, params, early)
	require.ErrorContains(t, err, "before proof horizon")

	late, err := NewMsgPruneReceipt(coretypes.DefaultAuthority, record, record.RetainUntilHeight)
	require.NoError(t, err)
	next, err := PruneReceiptRecord(records, params, late)
	require.NoError(t, err)
	require.Empty(t, next)
}

func TestXServiceReceiptsMissingExecutedOnChainReceiptGuard(t *testing.T) {
	receipt, call := testServiceReceiptForReceipts(t, 40)
	record, err := NewReceiptRecordFromServiceReceipt(receipt, ServiceReceiptRecordKindService, 10)
	require.NoError(t, err)
	require.NoError(t, ValidateExecutedOnChainCallsHaveReceipts([]UnifiedServiceCall{call}, []ReceiptRecord{record}))
	require.ErrorContains(t, ValidateExecutedOnChainCallsHaveReceipts([]UnifiedServiceCall{call}, nil), "missing receipt")
}

func TestXServiceReceiptsQueriesRootsAndProof(t *testing.T) {
	receipt, _ := testServiceReceiptForReceipts(t, 50)
	msg, err := NewMsgAnchorReceipt(coretypes.DefaultAuthority, receipt, receipt.ReceiptHash)
	require.NoError(t, err)
	records, record, err := AnchorReceiptRecord(nil, msg, ServiceReceiptRecordKindService, 15)
	require.NoError(t, err)

	receiptResp, err := QueryReceiptFromRecords(records, QueryReceipt{ReceiptID: record.ReceiptID})
	require.NoError(t, err)
	require.True(t, receiptResp.Found)

	serviceResp, err := QueryReceiptsByServiceFromRecords(records, QueryReceiptsByService{ServiceID: record.ServiceID})
	require.NoError(t, err)
	require.Len(t, serviceResp.Records, 1)
	require.NotEmpty(t, serviceResp.ResponseHash)

	rootResp, err := QueryReceiptRootFromRecords(records, QueryReceiptRoot{RootKind: ServiceReceiptRecordKindService, Height: record.Height})
	require.NoError(t, err)
	require.True(t, rootResp.Found)
	require.Equal(t, uint64(1), rootResp.Root.RecordCount)

	proofResp, err := QueryReceiptProofFromRecords(records, QueryReceiptProof{ReceiptID: record.ReceiptID})
	require.NoError(t, err)
	require.True(t, proofResp.Found)
	require.NoError(t, proofResp.Proof.Validate())
}

func testServiceReceiptForReceipts(t *testing.T, nonce uint64) (ServiceReceipt, UnifiedServiceCall) {
	t.Helper()
	ctx := coretypes.ServiceConsensusContext{ChainID: "aetra-test-1", Height: 70}
	descriptor := testInterfaceSystemDescriptor()
	call := testInteractionCall(t, ctx, descriptor, "submit", nonce, "servicereceipts/call")
	call.Kind = coretypes.ServiceCallKindOnChain
	call.CallID = coretypes.NormalizeServiceCall(ctx, call.ToServiceCallEnvelope()).CallID
	call.UnifiedCallHash = ComputeUnifiedServiceCallHash(call)
	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), coretypes.ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		coretypes.ServiceCallStatusExecuted,
		ResponseHash:	testInterfaceHash("servicereceipts/response"),
		ProofHash:	testInterfaceHash("servicereceipts/proof"),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	17,
		ExecutedHeight:	ctx.Height,
		AnchoredHeight:	ctx.Height,
	})
	require.NoError(t, err)
	return receipt, call
}

func removeXServiceReceiptsMessageForTest(messages []XServiceReceiptsMessageName, target XServiceReceiptsMessageName) []XServiceReceiptsMessageName {
	out := make([]XServiceReceiptsMessageName, 0, len(messages))
	for _, message := range messages {
		if message != target {
			out = append(out, message)
		}
	}
	return out
}
