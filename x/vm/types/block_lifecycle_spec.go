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
	AVMBlockStageBeginBlock			AVMBlockStage	= "begin_block"
	AVMBlockStageExecuteSyncTx		AVMBlockStage	= "execute_sync_tx"
	AVMBlockStageProcessAsyncQueue		AVMBlockStage	= "process_async_queue"
	AVMBlockStageExecuteScheduledMessages	AVMBlockStage	= "execute_scheduled_messages"
	AVMBlockStageProcessContinuations	AVMBlockStage	= "process_continuations"
	AVMBlockStageFinalizeStateRoots		AVMBlockStage	= "finalize_state_roots"
	AVMBlockStageEndBlockCleanup		AVMBlockStage	= "end_block_cleanup"

	AVMABCIPrepareProposal	AVMABCIPhase	= "prepare_proposal"
	AVMABCIProcessProposal	AVMABCIPhase	= "process_proposal"
	AVMABCIFinalizeBlock	AVMABCIPhase	= "finalize_block"
	AVMABCIEndBlock		AVMABCIPhase	= "end_block"
)

type AVMBlockStage string
type AVMABCIPhase string

type AVMBlockProposalMessage struct {
	MessageID	string
	ZoneID		zonestypes.ZoneID
	TargetActor	string
	Lane		AVMQueueLane
	Priority	uint8
	ScheduledHeight	uint64
	ExpiryHeight	uint64
	SenderHash	string
	Nonce		uint64
	GasLimit	uint64
}

type AVMBlockZoneBudget struct {
	ZoneID	zonestypes.ZoneID
	Budget	zonestypes.ZoneExecutionBudget
}

type AVMABCIProposalPlan struct {
	Height		uint64
	Phase		AVMABCIPhase
	Messages	[]AVMBlockProposalMessage
	ZoneBudgets	[]AVMBlockZoneBudget
	ProposalRoot	string
}

type AVMBlockLifecycleStep struct {
	Stage		AVMBlockStage
	Height		uint64
	Sequence	uint32
	Root		string
}

type AVMFinalizeBlockPlan struct {
	Height			uint64
	RouterRoot		string
	SyncRoot		string
	AsyncRoot		string
	ScheduledRoot		string
	ActorRoot		string
	ContinuationRoot	string
	ReceiptRoot		string
	AVMRoot			AVMRoot
	ZoneRoots		[]AVMZoneStateRoot
	Receipts		[]AVMExecutionReceipt
	Steps			[]AVMBlockLifecycleStep
	FinalizeRoot		string
}

type AVMEndBlockCleanupPlan struct {
	Height			uint64
	ExpiredMessages		[]AVMBlockProposalMessage
	PrunedTombstones	[]AVMAsyncReplayTombstone
	ProofHorizon		uint64
	ZoneSummaries		[]zonestypes.ZoneExecutionSummary
	CleanupRoot		string
}

type AVMBlockLifecyclePlan struct {
	Height		uint64
	PrepareProposal	AVMABCIProposalPlan
	ProcessProposal	AVMABCIProposalPlan
	FinalizeBlock	AVMFinalizeBlockPlan
	EndBlock	AVMEndBlockCleanupPlan
	PlanRoot	string
}

func NewAVMBlockProposalMessage(msg AVMAsyncMessage, lane AVMQueueLane, targetActor string) (AVMBlockProposalMessage, error) {
	msg = canonicalAVMAsyncMessage(msg)
	proposed := AVMBlockProposalMessage{
		MessageID:		msg.ID,
		ZoneID:			msg.DestinationZone,
		TargetActor:		strings.TrimSpace(targetActor),
		Lane:			lane,
		Priority:		msg.Priority,
		ScheduledHeight:	AVMMessageScheduledHeight(msg),
		ExpiryHeight:		msg.ExpiryHeight,
		SenderHash:		AVMQueueSenderHash(msg.SourceZone, msg.Source),
		Nonce:			msg.SenderNonce,
		GasLimit:		msg.GasLimit,
	}
	return proposed, proposed.Validate()
}

func NewAVMABCIProposalPlan(plan AVMABCIProposalPlan) (AVMABCIProposalPlan, error) {
	plan = canonicalAVMABCIProposalPlan(plan)
	plan.ProposalRoot = ComputeAVMABCIProposalRoot(plan)
	return plan, plan.ValidateProcessProposal()
}

func NewAVMFinalizeBlockPlan(plan AVMFinalizeBlockPlan) (AVMFinalizeBlockPlan, error) {
	plan = canonicalAVMFinalizeBlockPlan(plan)
	if len(plan.Steps) == 0 {
		plan.Steps = defaultAVMBlockLifecycleSteps(plan.Height, plan.RouterRoot, plan.SyncRoot, plan.AsyncRoot, plan.ScheduledRoot, plan.ActorRoot, plan.ContinuationRoot, plan.AVMRoot.RootHash)
	}
	plan.FinalizeRoot = ComputeAVMFinalizeBlockRoot(plan)
	return plan, plan.Validate()
}

func NewAVMEndBlockCleanupPlan(plan AVMEndBlockCleanupPlan) (AVMEndBlockCleanupPlan, error) {
	plan = canonicalAVMEndBlockCleanupPlan(plan)
	plan.CleanupRoot = ComputeAVMEndBlockCleanupRoot(plan)
	return plan, plan.Validate()
}

func NewAVMBlockLifecyclePlan(plan AVMBlockLifecyclePlan) (AVMBlockLifecyclePlan, error) {
	plan = canonicalAVMBlockLifecyclePlan(plan)
	plan.PlanRoot = ComputeAVMBlockLifecycleRoot(plan)
	return plan, plan.Validate()
}

func (m AVMBlockProposalMessage) Validate() error {
	m = canonicalAVMBlockProposalMessage(m)
	if err := zonestypes.ValidateHash("AVM proposal message id", m.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM proposal target actor", m.TargetActor, MaxRouterTargetLength); err != nil {
		return err
	}
	if !IsAVMQueueLane(m.Lane) || m.Lane == AVMQueueLaneFailed {
		return fmt.Errorf("invalid executable AVM proposal queue lane %q", m.Lane)
	}
	if m.ScheduledHeight == 0 {
		return errors.New("AVM proposal scheduled height must be positive")
	}
	if m.ExpiryHeight == 0 {
		return errors.New("AVM proposal expiry height must be positive")
	}
	if m.ScheduledHeight > m.ExpiryHeight {
		return errors.New("AVM proposal scheduled height must not exceed expiry")
	}
	if err := zonestypes.ValidateHash("AVM proposal sender hash", m.SenderHash); err != nil {
		return err
	}
	if m.GasLimit == 0 {
		return errors.New("AVM proposal gas limit must be positive")
	}
	return nil
}

func (m AVMBlockProposalMessage) IsEligible(height uint64) bool {
	return height >= m.ScheduledHeight && height <= m.ExpiryHeight
}

func (m AVMBlockProposalMessage) SortKey() string {
	return AVMQueueSortKey(m.Priority, m.ScheduledHeight, m.SenderHash, m.Nonce, m.MessageID)
}

func (p AVMABCIProposalPlan) ValidatePrepareProposal() error {
	return p.validateProposal(AVMABCIPrepareProposal)
}

func (p AVMABCIProposalPlan) ValidateProcessProposal() error {
	if err := p.validateProposal(AVMABCIProcessProposal); err != nil {
		return err
	}
	for _, msg := range p.Messages {
		if p.Height > msg.ExpiryHeight {
			return fmt.Errorf("expired AVM proposal message %q must not be proposed for execution", msg.MessageID)
		}
		if p.Height < msg.ScheduledHeight {
			return fmt.Errorf("ineligible AVM proposal message %q is below scheduled height", msg.MessageID)
		}
	}
	return nil
}

func (p AVMABCIProposalPlan) validateProposal(defaultPhase AVMABCIPhase) error {
	p.Phase = canonicalAVMABCIPhase(p.Phase, defaultPhase)
	if p.Height == 0 {
		return errors.New("AVM proposal height must be positive")
	}
	if !IsAVMABCIPhase(p.Phase) {
		return fmt.Errorf("invalid AVM ABCI phase %q", p.Phase)
	}
	if len(p.ZoneBudgets) == 0 {
		return errors.New("AVM proposal zone budgets are required")
	}
	budgets := make(map[zonestypes.ZoneID]zonestypes.ZoneExecutionBudget, len(p.ZoneBudgets))
	for i, item := range p.ZoneBudgets {
		if err := zonestypes.ValidateZoneID(item.ZoneID); err != nil {
			return err
		}
		if err := item.Budget.Validate(); err != nil {
			return err
		}
		if _, found := budgets[item.ZoneID]; found {
			return fmt.Errorf("duplicate AVM proposal zone budget %s", item.ZoneID)
		}
		budgets[item.ZoneID] = item.Budget
		if i > 0 && p.ZoneBudgets[i-1].ZoneID >= item.ZoneID {
			return errors.New("AVM proposal zone budgets must be sorted canonically")
		}
	}
	seenMessages := make(map[string]struct{}, len(p.Messages))
	nextBudgets := make(map[zonestypes.ZoneID]zonestypes.ZoneExecutionBudget, len(budgets))
	for zoneID, budget := range budgets {
		nextBudgets[zoneID] = budget
	}
	for i, msg := range p.Messages {
		if err := msg.Validate(); err != nil {
			return err
		}
		if _, found := seenMessages[msg.MessageID]; found {
			return fmt.Errorf("duplicate AVM proposal message %q", msg.MessageID)
		}
		seenMessages[msg.MessageID] = struct{}{}
		if i > 0 && compareAVMProposalMessages(p.Messages[i-1], msg) >= 0 {
			return errors.New("AVM proposal messages must be sorted deterministically")
		}
		budget, found := nextBudgets[msg.ZoneID]
		if !found {
			return fmt.Errorf("AVM proposal message %q has no zone budget", msg.MessageID)
		}
		consumed, err := budget.Consume(msg.GasLimit, 1)
		if err != nil {
			return err
		}
		nextBudgets[msg.ZoneID] = consumed
	}
	if p.ProposalRoot == "" {
		return errors.New("AVM proposal root is required")
	}
	if err := zonestypes.ValidateHash("AVM proposal root", p.ProposalRoot); err != nil {
		return err
	}
	if p.ProposalRoot != ComputeAVMABCIProposalRoot(p) {
		return errors.New("AVM proposal root mismatch")
	}
	return nil
}

func (p AVMFinalizeBlockPlan) Validate() error {
	p = canonicalAVMFinalizeBlockPlan(p)
	if p.Height == 0 {
		return errors.New("AVM FinalizeBlock height must be positive")
	}
	if err := validateFinalizeRoot("AVM FinalizeBlock router root", p.RouterRoot); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM FinalizeBlock sync root", value: p.SyncRoot},
		{name: "AVM FinalizeBlock async root", value: p.AsyncRoot},
		{name: "AVM FinalizeBlock scheduled root", value: p.ScheduledRoot},
		{name: "AVM FinalizeBlock actor root", value: p.ActorRoot},
		{name: "AVM FinalizeBlock continuation root", value: p.ContinuationRoot},
		{name: "AVM FinalizeBlock receipt root", value: p.ReceiptRoot},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if err := p.AVMRoot.Validate(); err != nil {
		return err
	}
	if p.AVMRoot.Height != p.Height {
		return errors.New("AVM FinalizeBlock AVM root height drift")
	}
	if p.AVMRoot.RouterRoot != p.RouterRoot ||
		p.AVMRoot.AsyncMessageRoot != p.AsyncRoot ||
		p.AVMRoot.ActorRoot != p.ActorRoot ||
		p.AVMRoot.ContinuationRoot != p.ContinuationRoot ||
		p.AVMRoot.ReceiptRoot != p.ReceiptRoot {
		return errors.New("AVM FinalizeBlock committed root drift")
	}
	if err := validateAVMFinalizeZoneRoots(p.Height, p.ZoneRoots); err != nil {
		return err
	}
	if err := validateAVMFinalizeReceipts(p.Receipts); err != nil {
		return err
	}
	if err := validateAVMBlockLifecycleSteps(p.Height, p.Steps); err != nil {
		return err
	}
	if p.FinalizeRoot == "" {
		return errors.New("AVM FinalizeBlock root is required")
	}
	if p.FinalizeRoot != ComputeAVMFinalizeBlockRoot(p) {
		return errors.New("AVM FinalizeBlock root mismatch")
	}
	return nil
}

func (p AVMEndBlockCleanupPlan) Validate() error {
	p = canonicalAVMEndBlockCleanupPlan(p)
	if p.Height == 0 {
		return errors.New("AVM EndBlock height must be positive")
	}
	if p.ProofHorizon == 0 {
		return errors.New("AVM EndBlock proof horizon must be positive")
	}
	for i, msg := range p.ExpiredMessages {
		if err := msg.Validate(); err != nil {
			return err
		}
		if p.Height <= msg.ExpiryHeight {
			return fmt.Errorf("AVM EndBlock message %q is not expired", msg.MessageID)
		}
		if i > 0 && compareAVMProposalMessages(p.ExpiredMessages[i-1], msg) >= 0 {
			return errors.New("AVM EndBlock expired messages must be sorted deterministically")
		}
	}
	for i, tombstone := range p.PrunedTombstones {
		if err := tombstone.Validate(); err != nil {
			return err
		}
		if tombstone.ConsumedHeight > p.Height || p.Height-tombstone.ConsumedHeight < p.ProofHorizon {
			return errors.New("AVM EndBlock tombstone is inside proof horizon")
		}
		if i > 0 && p.PrunedTombstones[i-1].MessageID >= tombstone.MessageID {
			return errors.New("AVM EndBlock tombstones must be sorted deterministically")
		}
	}
	for i, summary := range p.ZoneSummaries {
		if err := summary.Validate(); err != nil {
			return err
		}
		if summary.Height != p.Height {
			return errors.New("AVM EndBlock zone summary height drift")
		}
		if i > 0 && p.ZoneSummaries[i-1].ZoneID >= summary.ZoneID {
			return errors.New("AVM EndBlock zone summaries must be sorted deterministically")
		}
	}
	if p.CleanupRoot == "" {
		return errors.New("AVM EndBlock cleanup root is required")
	}
	if p.CleanupRoot != ComputeAVMEndBlockCleanupRoot(p) {
		return errors.New("AVM EndBlock cleanup root mismatch")
	}
	return nil
}

func (p AVMBlockLifecyclePlan) Validate() error {
	p = canonicalAVMBlockLifecyclePlan(p)
	if p.Height == 0 {
		return errors.New("AVM block lifecycle height must be positive")
	}
	if p.PrepareProposal.Height != p.Height ||
		p.ProcessProposal.Height != p.Height ||
		p.FinalizeBlock.Height != p.Height ||
		p.EndBlock.Height != p.Height {
		return errors.New("AVM block lifecycle height drift")
	}
	if err := p.PrepareProposal.ValidatePrepareProposal(); err != nil {
		return err
	}
	if err := p.ProcessProposal.ValidateProcessProposal(); err != nil {
		return err
	}
	if p.PrepareProposal.ProposalRoot != p.ProcessProposal.ProposalRoot {
		return errors.New("AVM block lifecycle proposal root drift")
	}
	if err := p.FinalizeBlock.Validate(); err != nil {
		return err
	}
	if err := p.EndBlock.Validate(); err != nil {
		return err
	}
	if p.PlanRoot == "" {
		return errors.New("AVM block lifecycle root is required")
	}
	if p.PlanRoot != ComputeAVMBlockLifecycleRoot(p) {
		return errors.New("AVM block lifecycle root mismatch")
	}
	return nil
}

func IsAVMABCIPhase(phase AVMABCIPhase) bool {
	switch phase {
	case AVMABCIPrepareProposal, AVMABCIProcessProposal, AVMABCIFinalizeBlock, AVMABCIEndBlock:
		return true
	default:
		return false
	}
}

func IsAVMBlockStage(stage AVMBlockStage) bool {
	switch stage {
	case AVMBlockStageBeginBlock,
		AVMBlockStageExecuteSyncTx,
		AVMBlockStageProcessAsyncQueue,
		AVMBlockStageExecuteScheduledMessages,
		AVMBlockStageProcessContinuations,
		AVMBlockStageFinalizeStateRoots,
		AVMBlockStageEndBlockCleanup:
		return true
	default:
		return false
	}
}

func ComputeAVMABCIProposalRoot(plan AVMABCIProposalPlan) string {
	plan.Phase = canonicalAVMABCIPhase(plan.Phase, AVMABCIProcessProposal)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-abci-proposal-v1")
	writeEngineUint64(h, plan.Height)
	writeEngineUint64(h, uint64(len(plan.ZoneBudgets)))
	for _, item := range plan.ZoneBudgets {
		writeEnginePart(h, string(item.ZoneID))
		writeEngineUint64(h, item.Budget.MaxGas)
		writeEngineUint64(h, item.Budget.GasUsed)
		writeEngineUint64(h, uint64(item.Budget.MaxMessages))
		writeEngineUint64(h, uint64(item.Budget.MessagesUsed))
	}
	writeEngineUint64(h, uint64(len(plan.Messages)))
	for _, msg := range plan.Messages {
		writeAVMBlockProposalMessage(h, msg)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMFinalizeBlockRoot(plan AVMFinalizeBlockPlan) string {
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-finalize-block-v1")
	writeEngineUint64(h, plan.Height)
	writeEnginePart(h, plan.RouterRoot)
	writeEnginePart(h, plan.SyncRoot)
	writeEnginePart(h, plan.AsyncRoot)
	writeEnginePart(h, plan.ScheduledRoot)
	writeEnginePart(h, plan.ActorRoot)
	writeEnginePart(h, plan.ContinuationRoot)
	writeEnginePart(h, plan.ReceiptRoot)
	writeEnginePart(h, plan.AVMRoot.RootHash)
	writeEngineUint64(h, uint64(len(plan.ZoneRoots)))
	for _, root := range plan.ZoneRoots {
		writeEnginePart(h, root.RootHash)
	}
	writeEngineUint64(h, uint64(len(plan.Receipts)))
	for _, receipt := range plan.Receipts {
		writeEnginePart(h, receipt.ReceiptHash)
	}
	writeEngineUint64(h, uint64(len(plan.Steps)))
	for _, step := range plan.Steps {
		writeEnginePart(h, string(step.Stage))
		writeEngineUint64(h, step.Height)
		writeEngineUint64(h, uint64(step.Sequence))
		writeEnginePart(h, step.Root)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMEndBlockCleanupRoot(plan AVMEndBlockCleanupPlan) string {
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-end-block-cleanup-v1")
	writeEngineUint64(h, plan.Height)
	writeEngineUint64(h, plan.ProofHorizon)
	writeEngineUint64(h, uint64(len(plan.ExpiredMessages)))
	for _, msg := range plan.ExpiredMessages {
		writeAVMBlockProposalMessage(h, msg)
	}
	writeEngineUint64(h, uint64(len(plan.PrunedTombstones)))
	for _, tombstone := range plan.PrunedTombstones {
		writeEnginePart(h, tombstone.MessageID)
		writeEngineUint64(h, tombstone.ConsumedHeight)
	}
	writeEngineUint64(h, uint64(len(plan.ZoneSummaries)))
	for _, summary := range plan.ZoneSummaries {
		writeEnginePart(h, summary.SummaryHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMBlockLifecycleRoot(plan AVMBlockLifecyclePlan) string {
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-block-lifecycle-v1")
	writeEngineUint64(h, plan.Height)
	writeEnginePart(h, plan.PrepareProposal.ProposalRoot)
	writeEnginePart(h, plan.ProcessProposal.ProposalRoot)
	writeEnginePart(h, plan.FinalizeBlock.FinalizeRoot)
	writeEnginePart(h, plan.EndBlock.CleanupRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func defaultAVMBlockLifecycleSteps(height uint64, routerRoot, syncRoot, asyncRoot, scheduledRoot, actorRoot, continuationRoot, finalRoot string) []AVMBlockLifecycleStep {
	return []AVMBlockLifecycleStep{
		{Stage: AVMBlockStageBeginBlock, Height: height, Sequence: 1, Root: routerRoot},
		{Stage: AVMBlockStageExecuteSyncTx, Height: height, Sequence: 2, Root: syncRoot},
		{Stage: AVMBlockStageProcessAsyncQueue, Height: height, Sequence: 3, Root: asyncRoot},
		{Stage: AVMBlockStageExecuteScheduledMessages, Height: height, Sequence: 4, Root: scheduledRoot},
		{Stage: AVMBlockStageProcessContinuations, Height: height, Sequence: 5, Root: continuationRoot},
		{Stage: AVMBlockStageFinalizeStateRoots, Height: height, Sequence: 6, Root: finalRoot},
		{Stage: AVMBlockStageEndBlockCleanup, Height: height, Sequence: 7, Root: actorRoot},
	}
}

func validateAVMBlockLifecycleSteps(height uint64, steps []AVMBlockLifecycleStep) error {
	expected := []AVMBlockStage{
		AVMBlockStageBeginBlock,
		AVMBlockStageExecuteSyncTx,
		AVMBlockStageProcessAsyncQueue,
		AVMBlockStageExecuteScheduledMessages,
		AVMBlockStageProcessContinuations,
		AVMBlockStageFinalizeStateRoots,
		AVMBlockStageEndBlockCleanup,
	}
	if len(steps) != len(expected) {
		return errors.New("AVM block lifecycle must declare all execution stages")
	}
	for i, step := range steps {
		if !IsAVMBlockStage(step.Stage) {
			return fmt.Errorf("invalid AVM block lifecycle stage %q", step.Stage)
		}
		if step.Stage != expected[i] {
			return errors.New("AVM block lifecycle stages are out of order")
		}
		if step.Height != height {
			return errors.New("AVM block lifecycle step height drift")
		}
		if step.Sequence != uint32(i+1) {
			return errors.New("AVM block lifecycle step sequence drift")
		}
		if err := validateFinalizeRoot("AVM block lifecycle step root", step.Root); err != nil {
			return err
		}
	}
	return nil
}

func validateAVMFinalizeZoneRoots(height uint64, roots []AVMZoneStateRoot) error {
	for i, root := range roots {
		if err := root.Validate(); err != nil {
			return err
		}
		if root.Height != height {
			return errors.New("AVM FinalizeBlock zone root height drift")
		}
		if i > 0 && (roots[i-1].ZoneID > root.ZoneID || roots[i-1].ZoneID == root.ZoneID && roots[i-1].Height >= root.Height) {
			return errors.New("AVM FinalizeBlock zone roots must be sorted deterministically")
		}
	}
	return nil
}

func validateAVMFinalizeReceipts(receipts []AVMExecutionReceipt) error {
	for i, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if !requiresReceiptGas(receipt.Status) {
			return errors.New("AVM FinalizeBlock receipts must be terminal committed receipts")
		}
		if i > 0 && receipts[i-1].ReceiptID >= receipt.ReceiptID {
			return errors.New("AVM FinalizeBlock receipts must be sorted deterministically")
		}
	}
	return nil
}

func compareAVMProposalMessages(left, right AVMBlockProposalMessage) int {
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.TargetActor < right.TargetActor {
		return -1
	}
	if left.TargetActor > right.TargetActor {
		return 1
	}
	if left.SortKey() < right.SortKey() {
		return -1
	}
	if left.SortKey() > right.SortKey() {
		return 1
	}
	return strings.Compare(left.MessageID, right.MessageID)
}

func canonicalAVMABCIProposalPlan(plan AVMABCIProposalPlan) AVMABCIProposalPlan {
	plan.Phase = canonicalAVMABCIPhase(plan.Phase, AVMABCIProcessProposal)
	plan.ProposalRoot = strings.TrimSpace(plan.ProposalRoot)
	plan.Messages = append([]AVMBlockProposalMessage(nil), plan.Messages...)
	for i := range plan.Messages {
		plan.Messages[i] = canonicalAVMBlockProposalMessage(plan.Messages[i])
	}
	sort.SliceStable(plan.Messages, func(i, j int) bool {
		return compareAVMProposalMessages(plan.Messages[i], plan.Messages[j]) < 0
	})
	plan.ZoneBudgets = append([]AVMBlockZoneBudget(nil), plan.ZoneBudgets...)
	sort.SliceStable(plan.ZoneBudgets, func(i, j int) bool {
		return plan.ZoneBudgets[i].ZoneID < plan.ZoneBudgets[j].ZoneID
	})
	return plan
}

func canonicalAVMBlockProposalMessage(msg AVMBlockProposalMessage) AVMBlockProposalMessage {
	msg.MessageID = strings.TrimSpace(msg.MessageID)
	msg.TargetActor = strings.TrimSpace(msg.TargetActor)
	msg.SenderHash = strings.TrimSpace(msg.SenderHash)
	return msg
}

func canonicalAVMFinalizeBlockPlan(plan AVMFinalizeBlockPlan) AVMFinalizeBlockPlan {
	plan.RouterRoot = strings.TrimSpace(plan.RouterRoot)
	plan.SyncRoot = strings.TrimSpace(plan.SyncRoot)
	plan.AsyncRoot = strings.TrimSpace(plan.AsyncRoot)
	plan.ScheduledRoot = strings.TrimSpace(plan.ScheduledRoot)
	plan.ActorRoot = strings.TrimSpace(plan.ActorRoot)
	plan.ContinuationRoot = strings.TrimSpace(plan.ContinuationRoot)
	plan.ReceiptRoot = strings.TrimSpace(plan.ReceiptRoot)
	plan.FinalizeRoot = strings.TrimSpace(plan.FinalizeRoot)
	plan.AVMRoot = canonicalAVMRoot(plan.AVMRoot)
	plan.ZoneRoots = append([]AVMZoneStateRoot(nil), plan.ZoneRoots...)
	for i := range plan.ZoneRoots {
		plan.ZoneRoots[i] = canonicalAVMZoneStateRoot(plan.ZoneRoots[i])
	}
	sort.SliceStable(plan.ZoneRoots, func(i, j int) bool {
		if plan.ZoneRoots[i].ZoneID != plan.ZoneRoots[j].ZoneID {
			return plan.ZoneRoots[i].ZoneID < plan.ZoneRoots[j].ZoneID
		}
		return plan.ZoneRoots[i].Height < plan.ZoneRoots[j].Height
	})
	plan.Receipts = append([]AVMExecutionReceipt(nil), plan.Receipts...)
	for i := range plan.Receipts {
		plan.Receipts[i] = canonicalAVMExecutionReceipt(plan.Receipts[i])
	}
	sort.SliceStable(plan.Receipts, func(i, j int) bool {
		return plan.Receipts[i].ReceiptID < plan.Receipts[j].ReceiptID
	})
	plan.Steps = append([]AVMBlockLifecycleStep(nil), plan.Steps...)
	for i := range plan.Steps {
		plan.Steps[i].Root = strings.TrimSpace(plan.Steps[i].Root)
	}
	sort.SliceStable(plan.Steps, func(i, j int) bool {
		return plan.Steps[i].Sequence < plan.Steps[j].Sequence
	})
	return plan
}

func canonicalAVMEndBlockCleanupPlan(plan AVMEndBlockCleanupPlan) AVMEndBlockCleanupPlan {
	plan.CleanupRoot = strings.TrimSpace(plan.CleanupRoot)
	plan.ExpiredMessages = append([]AVMBlockProposalMessage(nil), plan.ExpiredMessages...)
	for i := range plan.ExpiredMessages {
		plan.ExpiredMessages[i] = canonicalAVMBlockProposalMessage(plan.ExpiredMessages[i])
	}
	sort.SliceStable(plan.ExpiredMessages, func(i, j int) bool {
		return compareAVMProposalMessages(plan.ExpiredMessages[i], plan.ExpiredMessages[j]) < 0
	})
	plan.PrunedTombstones = append([]AVMAsyncReplayTombstone(nil), plan.PrunedTombstones...)
	for i := range plan.PrunedTombstones {
		plan.PrunedTombstones[i].MessageID = strings.TrimSpace(plan.PrunedTombstones[i].MessageID)
	}
	sort.SliceStable(plan.PrunedTombstones, func(i, j int) bool {
		return plan.PrunedTombstones[i].MessageID < plan.PrunedTombstones[j].MessageID
	})
	plan.ZoneSummaries = append([]zonestypes.ZoneExecutionSummary(nil), plan.ZoneSummaries...)
	sort.SliceStable(plan.ZoneSummaries, func(i, j int) bool {
		return plan.ZoneSummaries[i].ZoneID < plan.ZoneSummaries[j].ZoneID
	})
	return plan
}

func canonicalAVMBlockLifecyclePlan(plan AVMBlockLifecyclePlan) AVMBlockLifecyclePlan {
	plan.PrepareProposal = canonicalAVMABCIProposalPlan(plan.PrepareProposal)
	plan.ProcessProposal = canonicalAVMABCIProposalPlan(plan.ProcessProposal)
	plan.FinalizeBlock = canonicalAVMFinalizeBlockPlan(plan.FinalizeBlock)
	plan.EndBlock = canonicalAVMEndBlockCleanupPlan(plan.EndBlock)
	plan.PlanRoot = strings.TrimSpace(plan.PlanRoot)
	return plan
}

func canonicalAVMABCIPhase(phase AVMABCIPhase, fallback AVMABCIPhase) AVMABCIPhase {
	if phase == "" {
		return fallback
	}
	return phase
}

func validateFinalizeRoot(fieldName, value string) error {
	if err := zonestypes.ValidateHash(fieldName, value); err == nil {
		return nil
	}
	return validateAVMStatePrefix(fieldName, value)
}

func writeAVMBlockProposalMessage(h engineByteWriter, msg AVMBlockProposalMessage) {
	writeEnginePart(h, msg.MessageID)
	writeEnginePart(h, string(msg.ZoneID))
	writeEnginePart(h, msg.TargetActor)
	writeEnginePart(h, string(msg.Lane))
	writeEngineUint64(h, uint64(msg.Priority))
	writeEngineUint64(h, msg.ScheduledHeight)
	writeEngineUint64(h, msg.ExpiryHeight)
	writeEnginePart(h, msg.SenderHash)
	writeEngineUint64(h, msg.Nonce)
	writeEngineUint64(h, msg.GasLimit)
	writeEnginePart(h, msg.SortKey())
}
