package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKernelABCIProposalLifecycleIncludesRoutedMessagesAndGasLimits(t *testing.T) {
	state := abciLifecycleState(t)
	ctx := KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: 1_700_000_021}
	local := KernelMessageEnvelope{
		Kind:			KernelMessageLocalTx,
		TxHash:			testHash("abci/local"),
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDFinancial,
		DestinationShard:	"0",
		Sender:			"sender.local",
		Nonce:			1,
		GasLimit:		100,
		PriorityClass:		2,
		AdmissionHeight:	21,
	}
	routed := KernelMessageEnvelope{
		Kind:			KernelMessageRoutedInbound,
		TxHash:			testHash("abci/routed"),
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDContract,
		DestinationShard:	"0",
		Sender:			"sender.routed",
		Nonce:			1,
		GasLimit:		150,
		PriorityClass:		1,
		AdmissionHeight:	21,
		CommittedHeight:	20,
		EligibleHeight:		21,
	}
	limits := KernelGasLimits{MaxBlockGas: 1_000, MaxZoneGas: 500}

	proposalA, err := PrepareKernelABCIProposal(ctx, state, []KernelMessageEnvelope{local}, []KernelMessageEnvelope{routed}, limits)
	require.NoError(t, err)
	proposalB, err := PrepareKernelABCIProposal(ctx, state, []KernelMessageEnvelope{local}, []KernelMessageEnvelope{routed}, limits)
	require.NoError(t, err)
	require.Equal(t, proposalA, proposalB)
	require.Equal(t, uint64(250), proposalA.BlockGas)
	require.Len(t, proposalA.Workloads, 2)
	require.Len(t, proposalA.ZoneGas, 2)
	require.NoError(t, ProcessKernelABCIProposal(ctx, state, proposalA, []KernelMessageEnvelope{routed, local}, limits))

	tampered := proposalA
	tampered.RoutedMessageRoot = testHash("wrong-routed-root")
	require.ErrorContains(t, ProcessKernelABCIProposal(ctx, state, tampered, []KernelMessageEnvelope{local, routed}, limits), "routed message root")

	duplicateNonce := routed
	duplicateNonce.Sender = local.Sender
	require.ErrorContains(t, ProcessKernelABCIProposal(ctx, state, proposalA, []KernelMessageEnvelope{local, duplicateNonce}, limits), "duplicate")

	_, err = PrepareKernelABCIProposal(ctx, state, []KernelMessageEnvelope{local}, []KernelMessageEnvelope{routed}, KernelGasLimits{MaxBlockGas: 200, MaxZoneGas: 200})
	require.ErrorContains(t, err, "block gas")
}

func TestKernelABCIProcessProposalDeterministicAcceptRejectAndTimestampBounds(t *testing.T) {
	state := abciLifecycleState(t)
	ctx := KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: 1_700_000_021}
	local := KernelMessageEnvelope{
		Kind:			KernelMessageLocalTx,
		TxHash:			testHash("abci/deterministic/local"),
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDFinancial,
		DestinationShard:	"0",
		Sender:			"sender.deterministic",
		Nonce:			1,
		GasLimit:		100,
		PriorityClass:		1,
		AdmissionHeight:	21,
	}
	limits := KernelGasLimits{MaxBlockGas: 1_000, MaxZoneGas: 500}
	bounds := KernelTimestampBounds{PreviousBlockTimeUnix: 1_700_000_015, MaxForwardDriftSeconds: 120}
	proposal, err := PrepareKernelABCIProposal(ctx, state, []KernelMessageEnvelope{local}, nil, limits)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, ProcessKernelABCIProposalWithTimestampBounds(ctx, state, proposal, []KernelMessageEnvelope{local}, limits, bounds))
	}

	tampered := proposal
	tampered.BlockGas++
	errA := ProcessKernelABCIProposalWithTimestampBounds(ctx, state, tampered, []KernelMessageEnvelope{local}, limits, bounds)
	errB := ProcessKernelABCIProposalWithTimestampBounds(ctx, state, tampered, []KernelMessageEnvelope{local}, limits, bounds)
	require.Error(t, errA)
	require.Equal(t, errA.Error(), errB.Error())
	require.Contains(t, errA.Error(), "block gas mismatch")

	farFuture := ctx
	farFuture.BlockTimeUnix = bounds.PreviousBlockTimeUnix + bounds.MaxForwardDriftSeconds + 1
	errA = ProcessKernelABCIProposalWithTimestampBounds(farFuture, state, proposal, []KernelMessageEnvelope{local}, limits, bounds)
	errB = ProcessKernelABCIProposalWithTimestampBounds(farFuture, state, proposal, []KernelMessageEnvelope{local}, limits, bounds)
	require.Error(t, errA)
	require.Equal(t, errA.Error(), errB.Error())
	require.Contains(t, errA.Error(), "outside allowed consensus bounds")

	wrongHeight := proposal
	wrongHeight.Plan.Height = ctx.Height + 1
	errA = ProcessKernelABCIProposalWithTimestampBounds(ctx, state, wrongHeight, []KernelMessageEnvelope{local}, limits, bounds)
	errB = ProcessKernelABCIProposalWithTimestampBounds(ctx, state, wrongHeight, []KernelMessageEnvelope{local}, limits, bounds)
	require.Error(t, errA)
	require.Equal(t, errA.Error(), errB.Error())
	require.Contains(t, errA.Error(), "consensus context mismatch")
}

func TestKernelTimestampBoundsRejectNonCometBFTCompatibleTimes(t *testing.T) {
	bounds := KernelTimestampBounds{PreviousBlockTimeUnix: 1_700_000_015, MaxForwardDriftSeconds: 120}

	require.NoError(t, ValidateKernelTimestampBounds(
		KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: 1_700_000_021},
		bounds,
	))
	require.ErrorContains(t, ValidateKernelTimestampBounds(
		KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: bounds.PreviousBlockTimeUnix},
		bounds,
	), "after previous consensus time")
	require.ErrorContains(t, ValidateKernelTimestampBounds(
		KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: bounds.PreviousBlockTimeUnix + bounds.MaxForwardDriftSeconds + 1},
		bounds,
	), "outside allowed consensus bounds")
	require.ErrorContains(t, ValidateKernelTimestampBounds(
		KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: 1_700_000_021},
		KernelTimestampBounds{PreviousBlockTimeUnix: -1, MaxForwardDriftSeconds: 120},
	), "consensus supplied")
}

func TestKernelABCIFinalizeCommitAndCleanup(t *testing.T) {
	state := abciLifecycleState(t)
	ctx := KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: 1_700_000_021}
	local := KernelMessageEnvelope{
		Kind:			KernelMessageLocalTx,
		TxHash:			testHash("abci/finalize/local"),
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDFinancial,
		DestinationShard:	"0",
		Sender:			"sender.finalize",
		Nonce:			1,
		GasLimit:		100,
		PriorityClass:		1,
		AdmissionHeight:	21,
	}
	limits := KernelGasLimits{MaxBlockGas: 1_000, MaxZoneGas: 500}
	proposal, err := PrepareKernelABCIProposal(ctx, state, []KernelMessageEnvelope{local}, nil, limits)
	require.NoError(t, err)
	classified, err := ClassifyTransaction(state, ClassificationInput{
		Height:			ctx.Height,
		TxHash:			local.TxHash,
		SourceZone:		local.SourceZone,
		SourceShard:		local.SourceShard,
		DestinationZone:	local.DestinationZone,
		DestinationShard:	local.DestinationShard,
		AdmissionHeight:	local.AdmissionHeight,
	})
	require.NoError(t, err)
	receipt, err := ExecuteSync(classified, ExecutionResult{Success: true, ResultHash: testHash("abci/finalize/result")}, ctx.Height, 1)
	require.NoError(t, err)
	receiptsRoot, err := ComputeExecutionReceiptsRoot([]ExecutionReceipt{receipt})
	require.NoError(t, err)
	contributions := testContributions(ctx.Height)
	contributions.ReceiptsRoot = receiptsRoot

	next, finalization, cleanup, err := FinalizeKernelABCIBlock(ctx, state, proposal, []KernelMessageEnvelope{local}, KernelFinalizationInput{
		ZoneCommitments: []ZoneCommitment{
			testCommitment(t, ctx.Height, ZoneIDFinancial),
			testCommitment(t, ctx.Height, ZoneIDContract),
		},
		Receipts:	[]ExecutionReceipt{receipt},
		Contributions:	contributions,
	}, []KernelCleanupItem{
		{QueueID: "receipts", ItemID: "future", HeightDue: 22, DeleteRoot: testHash("future")},
		{QueueID: "receipts", ItemID: "old", HeightDue: 20, DeleteRoot: testHash("old")},
		{QueueID: "messages", ItemID: "ready", HeightDue: 21, DeleteRoot: testHash("ready")},
	}, 1)
	require.NoError(t, err)
	require.NoError(t, finalization.Validate())
	require.Len(t, cleanup.Processed, 1)
	require.Equal(t, "old", cleanup.Processed[0].ItemID)
	require.NotEmpty(t, next.RootSnapshots)

	record, err := CommitKernelABCIBlock(finalization, testHash("final-app-hash"))
	require.NoError(t, err)
	require.Equal(t, finalization.Height, record.Height)
	require.Equal(t, finalization.Header.HeaderHash, record.HeaderHash)
	require.Equal(t, finalization.GlobalRoot.GlobalRoot, record.GlobalRoot)
	require.Greater(t, record.ProofRootCount, uint64(0))
	require.NoError(t, record.Validate())
}

func TestKernelABCIRejectsMalformedRoutedBatch(t *testing.T) {
	state := abciLifecycleState(t)
	ctx := KernelConsensusContext{ChainID: "aetra-testnet", Height: 21, BlockTimeUnix: 1_700_000_021}
	routed := KernelMessageEnvelope{
		Kind:			KernelMessageRoutedInbound,
		TxHash:			testHash("abci/malformed-routed"),
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDContract,
		DestinationShard:	"0",
		Sender:			"sender.bad",
		Nonce:			1,
		GasLimit:		100,
		PriorityClass:		1,
		AdmissionHeight:	21,
		CommittedHeight:	19,
		EligibleHeight:		21,
	}
	_, err := PrepareKernelABCIProposal(ctx, state, nil, []KernelMessageEnvelope{routed}, KernelGasLimits{MaxBlockGas: 1_000, MaxZoneGas: 500})
	require.ErrorContains(t, err, "committed root")
}

func abciLifecycleState(t *testing.T) CoreState {
	t.Helper()
	state := EmptyState(TestnetParams())
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDContract, ZoneTypeContract, "contract"))
	require.NoError(t, err)
	state, err = RegisterShardLayout(state, testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"0"}))
	require.NoError(t, err)
	state, err = RegisterShardLayout(state, testShardLayout(t, ZoneIDContract, 1, []ShardID{"0"}))
	require.NoError(t, err)
	state, err = AppendZoneCommitment(state, testCommitment(t, 20, ZoneIDFinancial))
	require.NoError(t, err)
	state, _, err = CommitBlockRoots(state, 20)
	require.NoError(t, err)
	return state
}
