package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMAppExecutionPipelineDeclaresRequiredHooks(t *testing.T) {
	pipeline, err := DefaultAVMAppExecutionPipeline()
	require.NoError(t, err)
	require.NoError(t, pipeline.Validate())
	require.Len(t, pipeline.Hooks, 18)
	require.Equal(t, AVMHookPrepareClassifyTransactions, pipeline.Hooks[0].Name)
	require.Equal(t, AVMABCIPrepareProposal, pipeline.Hooks[0].Phase)
	require.Equal(t, AVMHookFinalizeCommitRoots, pipeline.Hooks[13].Name)
	require.Equal(t, AVMHookEndBlockEmitZoneSummaries, pipeline.Hooks[17].Name)
	require.Equal(t, ComputeAVMAppExecutionPipelineRoot(pipeline), pipeline.Root)

	missing := pipeline
	missing.Hooks = missing.Hooks[:len(missing.Hooks)-1]
	missing.Root = ComputeAVMAppExecutionPipelineRoot(missing)
	require.ErrorContains(t, missing.Validate(), "every required hook")

	drift := pipeline
	drift.Hooks = append([]AVMAppPipelineHook(nil), pipeline.Hooks...)
	drift.Hooks[0].Name = AVMHookPrepareGroupByZoneActor
	drift.Root = ComputeAVMAppExecutionPipelineRoot(drift)
	require.ErrorContains(t, drift.Validate(), "out of order")
}

func TestAVMEligibleMessageSelectionBuildsPrepareProposalAndBudgetAccounting(t *testing.T) {
	height := uint64(15)
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	readyHigh, err := NewAVMAsyncMessage(testAVMQueueMessage("ready-high", 1, 10, 9, 0, 50, 20))
	require.NoError(t, err)
	readyLow, err := NewAVMAsyncMessage(testAVMQueueMessage("ready-low", 1, 10, 3, 0, 50, 10))
	require.NoError(t, err)
	delayed, err := NewAVMAsyncMessage(testAVMQueueMessage("delayed", 1, 10, 10, 10, 50, 10))
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, readyLow, height, 10)
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, readyHigh, height, 10)
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, delayed, height, 10)
	require.NoError(t, err)

	proposal, selection, err := SelectEligibleAVMProposalMessages(
		height,
		queue,
		[]AVMAsyncMessage{readyLow, readyHigh, delayed},
		zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
	)
	require.NoError(t, err)
	require.Len(t, selection.Ready, 2)
	require.Len(t, proposal.Messages, 2)
	require.Equal(t, AVMABCIPrepareProposal, proposal.Phase)
	require.Equal(t, readyHigh.ID, proposal.Messages[0].MessageID)
	require.Equal(t, readyLow.ID, proposal.Messages[1].MessageID)

	accounting, err := ComputeAVMProposalBudgetAccounting(withAVMProposalPhase(proposal, AVMABCIProcessProposal))
	require.NoError(t, err)
	require.Len(t, accounting, 1)
	require.Equal(t, uint64(30), accounting[0].After.GasUsed)
	require.Equal(t, uint32(2), accounting[0].After.MessagesUsed)
}

func TestAVMProcessProposalVerificationCatchesQueueOrderingAndRootMismatch(t *testing.T) {
	expected := testAVMProposalPlan(t, 20)
	require.NoError(t, VerifyAVMProcessProposal(withAVMProposalPhase(expected, AVMABCIProcessProposal), expected))

	reordered := withAVMProposalPhase(expected, AVMABCIProcessProposal)
	reordered.Messages = append([]AVMBlockProposalMessage(nil), expected.Messages...)
	reordered.Messages[0], reordered.Messages[1] = reordered.Messages[1], reordered.Messages[0]
	reordered.ProposalRoot = ComputeAVMABCIProposalRoot(reordered)
	require.ErrorContains(t, VerifyAVMProcessProposal(reordered, expected), "sorted deterministically")

	wrongRoot := withAVMProposalPhase(expected, AVMABCIProcessProposal)
	wrongRoot.ProposalRoot = engineHash("wrong-proposal")
	require.ErrorContains(t, VerifyAVMProcessProposal(wrongRoot, expected), "root mismatch")
}

func TestAVMRootFinalizationAndReplayDeterminism(t *testing.T) {
	height := uint64(20)
	finalize := testAVMFinalizePlan(t, height)
	root, err := FinalizeAVMBlockRoots(
		height,
		finalize.RouterRoot,
		finalize.AsyncRoot,
		finalize.ActorRoot,
		finalize.AVMRoot.ContractRoot,
		finalize.ContinuationRoot,
		finalize.AVMRoot.InterfaceRoot,
		finalize.ReceiptRoot,
		finalize.ZoneRoots,
	)
	require.NoError(t, err)
	require.Equal(t, finalize.AVMRoot.RootHash, root.RootHash)

	proposal := testAVMProposalPlan(t, height)
	cleanup := testAVMEndBlockCleanup(t, height, nil)
	left, err := NewAVMBlockLifecyclePlan(AVMBlockLifecyclePlan{
		Height:			height,
		PrepareProposal:	withAVMProposalPhase(proposal, AVMABCIPrepareProposal),
		ProcessProposal:	withAVMProposalPhase(proposal, AVMABCIProcessProposal),
		FinalizeBlock:		finalize,
		EndBlock:		cleanup,
	})
	require.NoError(t, err)
	right, err := NewAVMBlockLifecyclePlan(AVMBlockLifecyclePlan{
		Height:			height,
		PrepareProposal:	withAVMProposalPhase(proposal, AVMABCIPrepareProposal),
		ProcessProposal:	withAVMProposalPhase(proposal, AVMABCIProcessProposal),
		FinalizeBlock:		finalize,
		EndBlock:		cleanup,
	})
	require.NoError(t, err)
	require.NoError(t, VerifyAVMBlockReplayDeterminism(left, right))

	right.EndBlock.CleanupRoot = engineHash("cleanup-drift")
	right.PlanRoot = ComputeAVMBlockLifecycleRoot(right)
	require.ErrorContains(t, VerifyAVMBlockReplayDeterminism(left, right), "root mismatch")
}

func TestAVMRootFinalizationRejectsZoneRootHeightDrift(t *testing.T) {
	finalize := testAVMFinalizePlan(t, 20)
	badZoneRoots := append([]AVMZoneStateRoot(nil), finalize.ZoneRoots...)
	badZoneRoots[0].Height++
	badZoneRoots[0].RootHash = ComputeAVMZoneStateRootHash(badZoneRoots[0])
	_, err := FinalizeAVMBlockRoots(
		20,
		finalize.RouterRoot,
		finalize.AsyncRoot,
		finalize.ActorRoot,
		finalize.AVMRoot.ContractRoot,
		finalize.ContinuationRoot,
		finalize.AVMRoot.InterfaceRoot,
		finalize.ReceiptRoot,
		badZoneRoots,
	)
	require.ErrorContains(t, err, "height drift")
}
