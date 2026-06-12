package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectZoneExecutionSummaryBuildsCommitmentHash(t *testing.T) {
	state := abciLifecycleState(t)
	envelopes := []KernelMessageEnvelope{
		{
			Kind:			KernelMessageLocalTx,
			TxHash:			testHash("summary/local"),
			SourceZone:		ZoneIDFinancial,
			SourceShard:		"0",
			DestinationZone:	ZoneIDFinancial,
			DestinationShard:	"0",
			Sender:			"summary.sender",
			Nonce:			1,
			GasLimit:		100,
			PriorityClass:		1,
			AdmissionHeight:	21,
		},
		{
			Kind:			KernelMessageRoutedInbound,
			TxHash:			testHash("summary/outbox"),
			SourceZone:		ZoneIDFinancial,
			SourceShard:		"0",
			DestinationZone:	ZoneIDContract,
			DestinationShard:	"0",
			Sender:			"summary.sender",
			Nonce:			2,
			GasLimit:		120,
			PriorityClass:		1,
			AdmissionHeight:	21,
			CommittedHeight:	20,
			EligibleHeight:		21,
		},
	}
	classified, err := ClassifyTransaction(state, ClassificationInput{
		Height:			21,
		TxHash:			envelopes[0].TxHash,
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDFinancial,
		DestinationShard:	"0",
		AdmissionHeight:	21,
	})
	require.NoError(t, err)
	receipt, err := ExecuteSync(classified, ExecutionResult{Success: true, ResultHash: testHash("summary/result")}, 21, 1)
	require.NoError(t, err)

	summary, err := CollectZoneExecutionSummary(21, ZoneIDFinancial, envelopes, []ExecutionReceipt{receipt}, 220, testHash("summary/events"))
	require.NoError(t, err)
	require.Equal(t, uint64(1), summary.LocalTxCount)
	require.Equal(t, uint64(1), summary.OutboxMessageCount)
	require.Equal(t, uint64(1), summary.ReceiptCount)
	require.Equal(t, uint64(0), summary.FailedReceiptCount)
	require.NoError(t, summary.Validate())

	commitment, err := NewZoneCommitmentFromSummary(summary, testHash("summary/state"), testHash("summary/inbox"), testHash("summary/params"))
	require.NoError(t, err)
	require.Equal(t, summary.SummaryHash, commitment.ExecutionSummaryHash)
	require.Equal(t, summary.MessageRoot, commitment.OutboxRoot)
	require.NoError(t, commitment.ValidateHash())
}

func TestRootAggregationInvariantsAndReplayProof(t *testing.T) {
	nodeA := populatedState(t, []ZoneID{ZoneIDFinancial, ZoneIDContract})
	nodeB := populatedState(t, []ZoneID{ZoneIDContract, ZoneIDFinancial})
	var err error
	nodeA, _, err = CommitBlockRootsWithContributions(nodeA, 7, testContributions(7))
	require.NoError(t, err)
	nodeB, _, err = CommitBlockRootsWithContributions(nodeB, 7, testContributions(7))
	require.NoError(t, err)

	require.NoError(t, ValidateRootAggregationInvariants(nodeA))
	require.NoError(t, ValidateRootAggregationInvariants(nodeB))
	require.NoError(t, AssertReplayIdenticalRoots(nodeA, nodeB))

	broken := nodeA.Export()
	broken.GlobalRoots[0].ZonesRoot = testHash("wrong-zones-root")
	broken.GlobalRoots[0].GlobalRoot = ComputeGlobalStateRootHash(broken.GlobalRoots[0])
	require.ErrorContains(t, ValidateRootAggregationInvariants(broken), "zones root mismatch")
}
