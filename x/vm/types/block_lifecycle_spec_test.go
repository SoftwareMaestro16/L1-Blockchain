package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMBlockLifecycleModelsABCIOrderAndCommittedRoots(t *testing.T) {
	height := uint64(20)
	proposal := testAVMProposalPlan(t, height)
	finalize := testAVMFinalizePlan(t, height)
	cleanup := testAVMEndBlockCleanup(t, height, nil)

	plan, err := NewAVMBlockLifecyclePlan(AVMBlockLifecyclePlan{
		Height:			height,
		PrepareProposal:	withAVMProposalPhase(proposal, AVMABCIPrepareProposal),
		ProcessProposal:	withAVMProposalPhase(proposal, AVMABCIProcessProposal),
		FinalizeBlock:		finalize,
		EndBlock:		cleanup,
	})
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Equal(t, ComputeAVMBlockLifecycleRoot(plan), plan.PlanRoot)
	require.Equal(t, proposal.ProposalRoot, plan.PrepareProposal.ProposalRoot)
	require.Equal(t, proposal.ProposalRoot, plan.ProcessProposal.ProposalRoot)
	require.Equal(t, AVMBlockStageBeginBlock, plan.FinalizeBlock.Steps[0].Stage)
	require.Equal(t, AVMBlockStageFinalizeStateRoots, plan.FinalizeBlock.Steps[5].Stage)

	mutated := plan
	mutated.FinalizeBlock.FinalizeRoot = engineHash("finalize-mutated")
	require.NotEqual(t, plan.PlanRoot, ComputeAVMBlockLifecycleRoot(mutated))
}

func TestAVMProcessProposalRejectsExpiredIneligibleAndBudgetOverflow(t *testing.T) {
	height := uint64(20)
	eligibleMsg, err := NewAVMAsyncMessage(testAVMQueueMessage("eligible", 1, 10, 5, 0, 50, 20))
	require.NoError(t, err)
	eligible, err := NewAVMBlockProposalMessage(eligibleMsg, AVMQueueLanePriority, "actor-a")
	require.NoError(t, err)

	expired := eligible
	expired.MessageID = engineHash("expired")
	expired.ExpiryHeight = 19
	expired.SenderHash = AVMQueueSenderHash(zonestypes.ZoneIDApplication, "expired")
	expired.Nonce = 2
	expiredPlan, err := NewAVMABCIProposalPlan(AVMABCIProposalPlan{
		Height:		height,
		Phase:		AVMABCIProcessProposal,
		Messages:	[]AVMBlockProposalMessage{expired},
		ZoneBudgets: []AVMBlockZoneBudget{{
			ZoneID:	zonestypes.ZoneIDContract,
			Budget:	zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
		}},
	})
	require.ErrorContains(t, err, "expired")
	require.NotEmpty(t, expiredPlan.ProposalRoot)

	ineligible := eligible
	ineligible.MessageID = engineHash("ineligible")
	ineligible.ScheduledHeight = 21
	ineligible.ExpiryHeight = 50
	ineligible.SenderHash = AVMQueueSenderHash(zonestypes.ZoneIDApplication, "ineligible")
	ineligible.Nonce = 3
	_, err = NewAVMABCIProposalPlan(AVMABCIProposalPlan{
		Height:		height,
		Phase:		AVMABCIProcessProposal,
		Messages:	[]AVMBlockProposalMessage{ineligible},
		ZoneBudgets: []AVMBlockZoneBudget{{
			ZoneID:	zonestypes.ZoneIDContract,
			Budget:	zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
		}},
	})
	require.ErrorContains(t, err, "below scheduled height")

	_, err = NewAVMABCIProposalPlan(AVMABCIProposalPlan{
		Height:		height,
		Phase:		AVMABCIProcessProposal,
		Messages:	[]AVMBlockProposalMessage{eligible},
		ZoneBudgets: []AVMBlockZoneBudget{{
			ZoneID:	zonestypes.ZoneIDContract,
			Budget:	zonestypes.ZoneExecutionBudget{MaxGas: 10, MaxMessages: 10},
		}},
	})
	require.ErrorContains(t, err, "gas used exceeds")
}

func TestAVMFinalizeBlockRejectsStageOrderAndRootDrift(t *testing.T) {
	height := uint64(20)
	finalize := testAVMFinalizePlan(t, height)

	reordered := finalize
	reordered.Steps = append([]AVMBlockLifecycleStep(nil), finalize.Steps...)
	reordered.Steps[1].Stage = AVMBlockStageProcessAsyncQueue
	reordered.Steps[2].Stage = AVMBlockStageExecuteSyncTx
	reordered.FinalizeRoot = ComputeAVMFinalizeBlockRoot(reordered)
	require.ErrorContains(t, reordered.Validate(), "out of order")

	drift := finalize
	drift.AVMRoot.ReceiptRoot = engineHash("wrong-receipt")
	drift.AVMRoot.RootHash = ComputeAVMRootHash(drift.AVMRoot)
	drift.FinalizeRoot = ComputeAVMFinalizeBlockRoot(drift)
	require.ErrorContains(t, drift.Validate(), "committed root drift")
}

func TestAVMEndBlockCleanupBoundsProofHorizonAndSummaries(t *testing.T) {
	height := uint64(20)
	msg, err := NewAVMAsyncMessage(testAVMQueueMessage("expired-cleanup", 1, 5, 1, 0, 10, 10))
	require.NoError(t, err)
	expired, err := NewAVMBlockProposalMessage(msg, AVMQueueLanePriority, "")
	require.NoError(t, err)

	cleanup := testAVMEndBlockCleanup(t, height, []AVMBlockProposalMessage{expired})
	require.NoError(t, cleanup.Validate())
	require.Equal(t, ComputeAVMEndBlockCleanupRoot(cleanup), cleanup.CleanupRoot)

	insideHorizon := cleanup
	insideHorizon.PrunedTombstones = []AVMAsyncReplayTombstone{{
		MessageID:	engineHash("recent-tombstone"),
		ConsumedHeight:	16,
	}}
	insideHorizon.CleanupRoot = ComputeAVMEndBlockCleanupRoot(insideHorizon)
	require.ErrorContains(t, insideHorizon.Validate(), "proof horizon")

	badSummary := cleanup
	badSummary.ZoneSummaries = append([]zonestypes.ZoneExecutionSummary(nil), cleanup.ZoneSummaries...)
	badSummary.ZoneSummaries[0].Height++
	badSummary.ZoneSummaries[0].SummaryHash = zonestypes.ComputeZoneExecutionSummaryHash(badSummary.ZoneSummaries[0])
	badSummary.CleanupRoot = ComputeAVMEndBlockCleanupRoot(badSummary)
	require.ErrorContains(t, badSummary.Validate(), "height drift")
}

func testAVMProposalPlan(t *testing.T, height uint64) AVMABCIProposalPlan {
	t.Helper()
	lowMsg, err := NewAVMAsyncMessage(testAVMQueueMessage("wallet-low", 2, 10, 3, 0, 50, 10))
	require.NoError(t, err)
	highMsg, err := NewAVMAsyncMessage(testAVMQueueMessage("wallet-high", 1, 10, 8, 0, 50, 20))
	require.NoError(t, err)
	low, err := NewAVMBlockProposalMessage(lowMsg, AVMQueueLanePriority, "actor-b")
	require.NoError(t, err)
	high, err := NewAVMBlockProposalMessage(highMsg, AVMQueueLanePriority, "actor-a")
	require.NoError(t, err)
	proposal, err := NewAVMABCIProposalPlan(AVMABCIProposalPlan{
		Height:		height,
		Phase:		AVMABCIProcessProposal,
		Messages:	[]AVMBlockProposalMessage{low, high},
		ZoneBudgets: []AVMBlockZoneBudget{{
			ZoneID:	zonestypes.ZoneIDContract,
			Budget:	zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
		}},
	})
	require.NoError(t, err)
	return proposal
}

func testAVMFinalizePlan(t *testing.T, height uint64) AVMFinalizeBlockPlan {
	t.Helper()
	receiptMsg, err := NewAVMAsyncMessage(testAVMQueueMessage("receipt-msg", 1, 10, 5, 0, 50, 10))
	require.NoError(t, err)
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		receiptMsg.ID,
		ZoneID:			zonestypes.ZoneIDContract,
		Executor:		"finalize-executor",
		Status:			AVMReceiptStatusExecuted,
		GasUsed:		10,
		StorageWritten:		1,
		EventsHash:		engineHash("events"),
		OutputMessagesRoot:	engineHash("output"),
		CreatedHeight:		height,
	})
	require.NoError(t, err)
	zoneRoot, err := NewAVMZoneStateRoot(AVMZoneStateRoot{
		ZoneID:			zonestypes.ZoneIDContract,
		Height:			height,
		StateRoot:		engineHash("zone-state"),
		MessageRoot:		engineHash("zone-message"),
		ExecutionRoot:		engineHash("zone-execution"),
		ContinuationRoot:	engineHash("zone-continuation"),
	})
	require.NoError(t, err)
	avmRoot, err := NewAVMRoot(AVMRoot{
		Height:			height,
		RouterRoot:		engineHash("router"),
		AsyncMessageRoot:	engineHash("async"),
		ActorRoot:		engineHash("actor"),
		ContractRoot:		engineHash("contract"),
		ContinuationRoot:	engineHash("continuation"),
		InterfaceRoot:		engineHash("interface"),
		ReceiptRoot:		engineHash("receipt"),
	})
	require.NoError(t, err)
	finalize, err := NewAVMFinalizeBlockPlan(AVMFinalizeBlockPlan{
		Height:			height,
		RouterRoot:		avmRoot.RouterRoot,
		SyncRoot:		engineHash("sync"),
		AsyncRoot:		avmRoot.AsyncMessageRoot,
		ScheduledRoot:		engineHash("scheduled"),
		ActorRoot:		avmRoot.ActorRoot,
		ContinuationRoot:	avmRoot.ContinuationRoot,
		ReceiptRoot:		avmRoot.ReceiptRoot,
		AVMRoot:		avmRoot,
		ZoneRoots:		[]AVMZoneStateRoot{zoneRoot},
		Receipts:		[]AVMExecutionReceipt{receipt},
	})
	require.NoError(t, err)
	return finalize
}

func testAVMEndBlockCleanup(t *testing.T, height uint64, expired []AVMBlockProposalMessage) AVMEndBlockCleanupPlan {
	t.Helper()
	summary, err := zonestypes.NewZoneExecutionSummary(zonestypes.ZoneExecutionSummary{
		ZoneID:			zonestypes.ZoneIDContract,
		Height:			height,
		TransactionsExecuted:	1,
		InboundMessagesApplied:	1,
		ReceiptsProduced:	1,
		GasConsumed:		10,
		StateRoot:		engineHash("summary-state"),
		InboxRoot:		engineHash("summary-inbox"),
		OutboxRoot:		engineHash("summary-outbox"),
		ReceiptRoot:		engineHash("summary-receipt"),
		ExecutionResultRoot:	engineHash("summary-execution"),
	})
	require.NoError(t, err)
	cleanup, err := NewAVMEndBlockCleanupPlan(AVMEndBlockCleanupPlan{
		Height:			height,
		ExpiredMessages:	expired,
		PrunedTombstones: []AVMAsyncReplayTombstone{{
			MessageID:	engineHash("old-tombstone"),
			ConsumedHeight:	1,
		}},
		ProofHorizon:	10,
		ZoneSummaries:	[]zonestypes.ZoneExecutionSummary{summary},
	})
	require.NoError(t, err)
	return cleanup
}

func withAVMProposalPhase(plan AVMABCIProposalPlan, phase AVMABCIPhase) AVMABCIProposalPlan {
	plan.Phase = phase
	plan.ProposalRoot = ComputeAVMABCIProposalRoot(plan)
	return plan
}
