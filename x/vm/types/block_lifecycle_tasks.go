package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMHookPrepareClassifyTransactions	AVMAppPipelineHookName	= "prepare.classify_transactions"
	AVMHookPrepareGroupByZoneActor		AVMAppPipelineHookName	= "prepare.group_by_zone_actor"
	AVMHookPrepareIncludeEligibleScheduled	AVMAppPipelineHookName	= "prepare.include_eligible_scheduled"
	AVMHookPrepareReserveZoneBudgets	AVMAppPipelineHookName	= "prepare.reserve_zone_budgets"
	AVMHookProcessVerifyQueueOrdering	AVMAppPipelineHookName	= "process.verify_queue_ordering"
	AVMHookProcessVerifyBudgetBounds	AVMAppPipelineHookName	= "process.verify_budget_bounds"
	AVMHookProcessVerifyMessageEligibility	AVMAppPipelineHookName	= "process.verify_message_eligibility"
	AVMHookProcessRejectExpiredMessages	AVMAppPipelineHookName	= "process.reject_expired_messages"
	AVMHookFinalizeExecuteSyncTransactions	AVMAppPipelineHookName	= "finalize.execute_sync_transactions"
	AVMHookFinalizeDrainAsyncQueues		AVMAppPipelineHookName	= "finalize.drain_async_queues"
	AVMHookFinalizeExecuteActorHandlers	AVMAppPipelineHookName	= "finalize.execute_actor_handlers"
	AVMHookFinalizeResumeContinuations	AVMAppPipelineHookName	= "finalize.resume_continuations"
	AVMHookFinalizeEmitReceipts		AVMAppPipelineHookName	= "finalize.emit_receipts"
	AVMHookFinalizeCommitRoots		AVMAppPipelineHookName	= "finalize.commit_roots"
	AVMHookEndBlockBoundedCleanup		AVMAppPipelineHookName	= "end_block.bounded_cleanup"
	AVMHookEndBlockMarkExpiredMessages	AVMAppPipelineHookName	= "end_block.mark_expired_messages"
	AVMHookEndBlockPruneTombstones		AVMAppPipelineHookName	= "end_block.prune_tombstones"
	AVMHookEndBlockEmitZoneSummaries	AVMAppPipelineHookName	= "end_block.emit_zone_summaries"
)

type AVMAppPipelineHookName string

type AVMAppPipelineHook struct {
	Name		AVMAppPipelineHookName
	Phase		AVMABCIPhase
	Stage		AVMBlockStage
	Sequence	uint32
	Root		string
}

type AVMAppExecutionPipeline struct {
	Hooks	[]AVMAppPipelineHook
	Root	string
}

type AVMZoneBudgetAccounting struct {
	ZoneID	zonestypes.ZoneID
	Before	zonestypes.ZoneExecutionBudget
	After	zonestypes.ZoneExecutionBudget
}

func DefaultAVMAppExecutionPipeline() (AVMAppExecutionPipeline, error) {
	pipeline := AVMAppExecutionPipeline{Hooks: []AVMAppPipelineHook{
		{Name: AVMHookPrepareClassifyTransactions, Phase: AVMABCIPrepareProposal, Stage: AVMBlockStageBeginBlock, Sequence: 1},
		{Name: AVMHookPrepareGroupByZoneActor, Phase: AVMABCIPrepareProposal, Stage: AVMBlockStageBeginBlock, Sequence: 2},
		{Name: AVMHookPrepareIncludeEligibleScheduled, Phase: AVMABCIPrepareProposal, Stage: AVMBlockStageProcessAsyncQueue, Sequence: 3},
		{Name: AVMHookPrepareReserveZoneBudgets, Phase: AVMABCIPrepareProposal, Stage: AVMBlockStageProcessAsyncQueue, Sequence: 4},
		{Name: AVMHookProcessVerifyQueueOrdering, Phase: AVMABCIProcessProposal, Stage: AVMBlockStageProcessAsyncQueue, Sequence: 5},
		{Name: AVMHookProcessVerifyBudgetBounds, Phase: AVMABCIProcessProposal, Stage: AVMBlockStageProcessAsyncQueue, Sequence: 6},
		{Name: AVMHookProcessVerifyMessageEligibility, Phase: AVMABCIProcessProposal, Stage: AVMBlockStageExecuteScheduledMessages, Sequence: 7},
		{Name: AVMHookProcessRejectExpiredMessages, Phase: AVMABCIProcessProposal, Stage: AVMBlockStageExecuteScheduledMessages, Sequence: 8},
		{Name: AVMHookFinalizeExecuteSyncTransactions, Phase: AVMABCIFinalizeBlock, Stage: AVMBlockStageExecuteSyncTx, Sequence: 9},
		{Name: AVMHookFinalizeDrainAsyncQueues, Phase: AVMABCIFinalizeBlock, Stage: AVMBlockStageProcessAsyncQueue, Sequence: 10},
		{Name: AVMHookFinalizeExecuteActorHandlers, Phase: AVMABCIFinalizeBlock, Stage: AVMBlockStageExecuteScheduledMessages, Sequence: 11},
		{Name: AVMHookFinalizeResumeContinuations, Phase: AVMABCIFinalizeBlock, Stage: AVMBlockStageProcessContinuations, Sequence: 12},
		{Name: AVMHookFinalizeEmitReceipts, Phase: AVMABCIFinalizeBlock, Stage: AVMBlockStageFinalizeStateRoots, Sequence: 13},
		{Name: AVMHookFinalizeCommitRoots, Phase: AVMABCIFinalizeBlock, Stage: AVMBlockStageFinalizeStateRoots, Sequence: 14},
		{Name: AVMHookEndBlockBoundedCleanup, Phase: AVMABCIEndBlock, Stage: AVMBlockStageEndBlockCleanup, Sequence: 15},
		{Name: AVMHookEndBlockMarkExpiredMessages, Phase: AVMABCIEndBlock, Stage: AVMBlockStageEndBlockCleanup, Sequence: 16},
		{Name: AVMHookEndBlockPruneTombstones, Phase: AVMABCIEndBlock, Stage: AVMBlockStageEndBlockCleanup, Sequence: 17},
		{Name: AVMHookEndBlockEmitZoneSummaries, Phase: AVMABCIEndBlock, Stage: AVMBlockStageEndBlockCleanup, Sequence: 18},
	}}
	pipeline = canonicalAVMAppExecutionPipeline(pipeline)
	pipeline.Root = ComputeAVMAppExecutionPipelineRoot(pipeline)
	return pipeline, pipeline.Validate()
}

func (p AVMAppExecutionPipeline) Validate() error {
	p = canonicalAVMAppExecutionPipeline(p)
	required := defaultAVMPipelineHookNames()
	if len(p.Hooks) != len(required) {
		return errors.New("AVM app execution pipeline must declare every required hook")
	}
	seen := make(map[AVMAppPipelineHookName]struct{}, len(p.Hooks))
	for i, hook := range p.Hooks {
		if err := hook.Validate(); err != nil {
			return err
		}
		if hook.Name != required[i] {
			return errors.New("AVM app execution pipeline hooks are out of order")
		}
		if _, found := seen[hook.Name]; found {
			return fmt.Errorf("duplicate AVM app pipeline hook %q", hook.Name)
		}
		seen[hook.Name] = struct{}{}
		if hook.Sequence != uint32(i+1) {
			return errors.New("AVM app execution pipeline hook sequence drift")
		}
	}
	if p.Root == "" {
		return errors.New("AVM app execution pipeline root is required")
	}
	if err := zonestypes.ValidateHash("AVM app execution pipeline root", p.Root); err != nil {
		return err
	}
	if p.Root != ComputeAVMAppExecutionPipelineRoot(p) {
		return errors.New("AVM app execution pipeline root mismatch")
	}
	return nil
}

func (h AVMAppPipelineHook) Validate() error {
	if !IsAVMAppPipelineHookName(h.Name) {
		return fmt.Errorf("invalid AVM app pipeline hook %q", h.Name)
	}
	if !IsAVMABCIPhase(h.Phase) {
		return fmt.Errorf("invalid AVM app pipeline phase %q", h.Phase)
	}
	if !IsAVMBlockStage(h.Stage) {
		return fmt.Errorf("invalid AVM app pipeline stage %q", h.Stage)
	}
	if h.Sequence == 0 {
		return errors.New("AVM app pipeline hook sequence must be positive")
	}
	if h.Root != "" {
		return zonestypes.ValidateHash("AVM app pipeline hook root", h.Root)
	}
	return nil
}

func SelectEligibleAVMProposalMessages(height uint64, queue AVMZoneQueue, messages []AVMAsyncMessage, budget zonestypes.ZoneExecutionBudget) (AVMABCIProposalPlan, AVMZoneQueueSelection, error) {
	selection, err := SelectAVMZoneQueueWork(queue, messages, height, budget)
	if err != nil {
		return AVMABCIProposalPlan{}, AVMZoneQueueSelection{}, err
	}
	proposed := make([]AVMBlockProposalMessage, 0, len(selection.Ready))
	for _, msg := range selection.Ready {
		entry, found := findAVMQueueEntry(queue, msg.ID)
		if !found {
			return AVMABCIProposalPlan{}, AVMZoneQueueSelection{}, fmt.Errorf("eligible AVM message %q has no queue entry", msg.ID)
		}
		proposal, err := NewAVMBlockProposalMessage(msg, entry.Lane, entryLaneTargetActor(entry, msg))
		if err != nil {
			return AVMABCIProposalPlan{}, AVMZoneQueueSelection{}, err
		}
		proposed = append(proposed, proposal)
	}
	plan, err := NewAVMABCIProposalPlan(AVMABCIProposalPlan{
		Height:		height,
		Phase:		AVMABCIPrepareProposal,
		Messages:	proposed,
		ZoneBudgets: []AVMBlockZoneBudget{{
			ZoneID:	queue.ZoneID,
			Budget:	budget,
		}},
	})
	if err != nil {
		return AVMABCIProposalPlan{}, AVMZoneQueueSelection{}, err
	}
	plan.Phase = AVMABCIPrepareProposal
	return plan, selection, nil
}

func VerifyAVMProcessProposal(proposed AVMABCIProposalPlan, expected AVMABCIProposalPlan) error {
	if err := proposed.ValidateProcessProposal(); err != nil {
		return err
	}
	expected = canonicalAVMABCIProposalPlan(expected)
	if err := expected.ValidateProcessProposal(); err != nil {
		return err
	}
	if proposed.Height != expected.Height {
		return errors.New("AVM process proposal height drift")
	}
	if proposed.ProposalRoot != expected.ProposalRoot {
		return errors.New("AVM process proposal root mismatch")
	}
	if len(proposed.Messages) != len(expected.Messages) {
		return errors.New("AVM process proposal message count mismatch")
	}
	for i := range proposed.Messages {
		if proposed.Messages[i].MessageID != expected.Messages[i].MessageID ||
			proposed.Messages[i].SortKey() != expected.Messages[i].SortKey() {
			return errors.New("AVM process proposal queue ordering mismatch")
		}
	}
	return nil
}

func ComputeAVMProposalBudgetAccounting(plan AVMABCIProposalPlan) ([]AVMZoneBudgetAccounting, error) {
	plan = canonicalAVMABCIProposalPlan(plan)
	if err := plan.ValidateProcessProposal(); err != nil {
		return nil, err
	}
	accounting := make([]AVMZoneBudgetAccounting, 0, len(plan.ZoneBudgets))
	byZone := make(map[zonestypes.ZoneID]int, len(plan.ZoneBudgets))
	for _, item := range plan.ZoneBudgets {
		accounting = append(accounting, AVMZoneBudgetAccounting{
			ZoneID:	item.ZoneID,
			Before:	item.Budget,
			After:	item.Budget,
		})
		byZone[item.ZoneID] = len(accounting) - 1
	}
	for _, msg := range plan.Messages {
		idx := byZone[msg.ZoneID]
		after, err := accounting[idx].After.Consume(msg.GasLimit, 1)
		if err != nil {
			return nil, err
		}
		accounting[idx].After = after
	}
	return accounting, nil
}

func FinalizeAVMBlockRoots(height uint64, routerRoot, asyncRoot, actorRoot, contractRoot, continuationRoot, interfaceRoot, receiptRoot string, zoneRoots []AVMZoneStateRoot) (AVMRoot, error) {
	root, err := NewAVMRoot(AVMRoot{
		Height:			height,
		RouterRoot:		routerRoot,
		AsyncMessageRoot:	asyncRoot,
		ActorRoot:		actorRoot,
		ContractRoot:		contractRoot,
		ContinuationRoot:	continuationRoot,
		InterfaceRoot:		interfaceRoot,
		ReceiptRoot:		receiptRoot,
	})
	if err != nil {
		return AVMRoot{}, err
	}
	for _, zoneRoot := range zoneRoots {
		if err := zoneRoot.Validate(); err != nil {
			return AVMRoot{}, err
		}
		if zoneRoot.Height != height {
			return AVMRoot{}, errors.New("AVM root finalization zone root height drift")
		}
	}
	return root, nil
}

func VerifyAVMBlockReplayDeterminism(left, right AVMBlockLifecyclePlan) error {
	left = canonicalAVMBlockLifecyclePlan(left)
	right = canonicalAVMBlockLifecyclePlan(right)
	if err := left.Validate(); err != nil {
		return err
	}
	if err := right.Validate(); err != nil {
		return err
	}
	if left.Height != right.Height {
		return errors.New("AVM replay height drift")
	}
	if left.PrepareProposal.ProposalRoot != right.PrepareProposal.ProposalRoot ||
		left.ProcessProposal.ProposalRoot != right.ProcessProposal.ProposalRoot ||
		left.FinalizeBlock.FinalizeRoot != right.FinalizeBlock.FinalizeRoot ||
		left.EndBlock.CleanupRoot != right.EndBlock.CleanupRoot ||
		left.PlanRoot != right.PlanRoot {
		return errors.New("AVM block replay root mismatch")
	}
	return nil
}

func ComputeAVMAppExecutionPipelineRoot(pipeline AVMAppExecutionPipeline) string {
	pipeline = canonicalAVMAppExecutionPipeline(pipeline)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-app-execution-pipeline-v1")
	writeEngineUint64(h, uint64(len(pipeline.Hooks)))
	for _, hook := range pipeline.Hooks {
		writeEnginePart(h, string(hook.Name))
		writeEnginePart(h, string(hook.Phase))
		writeEnginePart(h, string(hook.Stage))
		writeEngineUint64(h, uint64(hook.Sequence))
		writeEnginePart(h, hook.Root)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func IsAVMAppPipelineHookName(name AVMAppPipelineHookName) bool {
	for _, candidate := range defaultAVMPipelineHookNames() {
		if name == candidate {
			return true
		}
	}
	return false
}

func defaultAVMPipelineHookNames() []AVMAppPipelineHookName {
	return []AVMAppPipelineHookName{
		AVMHookPrepareClassifyTransactions,
		AVMHookPrepareGroupByZoneActor,
		AVMHookPrepareIncludeEligibleScheduled,
		AVMHookPrepareReserveZoneBudgets,
		AVMHookProcessVerifyQueueOrdering,
		AVMHookProcessVerifyBudgetBounds,
		AVMHookProcessVerifyMessageEligibility,
		AVMHookProcessRejectExpiredMessages,
		AVMHookFinalizeExecuteSyncTransactions,
		AVMHookFinalizeDrainAsyncQueues,
		AVMHookFinalizeExecuteActorHandlers,
		AVMHookFinalizeResumeContinuations,
		AVMHookFinalizeEmitReceipts,
		AVMHookFinalizeCommitRoots,
		AVMHookEndBlockBoundedCleanup,
		AVMHookEndBlockMarkExpiredMessages,
		AVMHookEndBlockPruneTombstones,
		AVMHookEndBlockEmitZoneSummaries,
	}
}

func canonicalAVMAppExecutionPipeline(pipeline AVMAppExecutionPipeline) AVMAppExecutionPipeline {
	pipeline.Root = strings.TrimSpace(pipeline.Root)
	pipeline.Hooks = append([]AVMAppPipelineHook(nil), pipeline.Hooks...)
	for i := range pipeline.Hooks {
		pipeline.Hooks[i].Root = strings.TrimSpace(pipeline.Hooks[i].Root)
	}
	sort.SliceStable(pipeline.Hooks, func(i, j int) bool {
		return pipeline.Hooks[i].Sequence < pipeline.Hooks[j].Sequence
	})
	return pipeline
}

func entryLaneTargetActor(entry AVMZoneQueueEntry, msg AVMAsyncMessage) string {
	if msg.DestinationActorOptional != "" {
		return msg.DestinationActorOptional
	}
	if entry.Lane == AVMQueueLaneRetry || entry.Lane == AVMQueueLaneDelayed {
		return "scheduler"
	}
	return ""
}
