package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type CompatibilitySurface string

const (
	CompatibilityCometBFTP2P		CompatibilitySurface	= "cometbft_p2p_transport"
	CompatibilityCometBFTConsensusMessages	CompatibilitySurface	= "cometbft_consensus_messages"
	CompatibilityCosmosSDKTxFlow		CompatibilitySurface	= "cosmos_sdk_transaction_flow"
	CompatibilityABCIPlusPlus		CompatibilitySurface	= "abci_plus_plus_lifecycle"
	CompatibilityBlockSTM			CompatibilitySurface	= "blockstm_execution_grouping"
	CompatibilityGRPC			CompatibilitySurface	= "grpc_external_api"
	CompatibilityREST			CompatibilitySurface	= "rest_external_api"
	CompatibilityRPC			CompatibilitySurface	= "rpc_external_api"
	CompatibilityStateSyncSnapshots		CompatibilitySurface	= "state_sync_snapshots"
)

type ABCILifecyclePhase string

const (
	ABCIPrepareProposal	ABCILifecyclePhase	= "PrepareProposal"
	ABCIProcessProposal	ABCILifecyclePhase	= "ProcessProposal"
	ABCIFinalizeBlock	ABCILifecyclePhase	= "FinalizeBlock"
)

type CosmosCometBFTCompatibilityPlan struct {
	Adapter					AetherNetworkingAdapter
	Surfaces				[]CompatibilitySurface
	Phases					[]ABCILifecyclePhase
	PreserveCometBFTTransport		bool
	PreserveCometBFTConsensusMessages	bool
	PreserveCosmosSDKTransactionFlow	bool
	SupportsBlockSTMExecutionGrouping	bool
	SupportsGRPCExternalAPI			bool
	SupportsRESTExternalAPI			bool
	SupportsRPCExternalAPI			bool
	SupportsStateSyncAndSnapshotMechanism	bool
}

type ANAProposalHint struct {
	HintID			string
	ScheduleID		string
	ScheduleHash		string
	ZoneID			string
	ShardID			string
	DeterminismSource	DeterminismSource
	CommittedStateDerived	bool
	PeerLocal		bool
	AdvisoryOnly		bool
	UsedForValidity		bool
	UsedForOrdering		bool
	Priority		uint32
	BlockSTMGroupID		string
	ExecutionGroupID	string
	DeterministicHintProof	string
}

type ABCIProposalGroup struct {
	GroupID		string
	ZoneID		string
	ShardID		string
	ScheduleID	string
	ScheduleHash	string
	TransactionIDs	[]string
	MessageIDs	[]string
	HintIDs		[]string
	BlockSTMGroupID	string
	Deterministic	bool
}

type ABCIPrepareProposalInput struct {
	Height		uint64
	Adapter		AetherNetworkingAdapter
	Schedules	[]ExecutionMessageSchedule
	Hints		[]ANAProposalHint
}

type ABCIProposalPlan struct {
	Height				uint64
	Phase				ABCILifecyclePhase
	Groups				[]ABCIProposalGroup
	ScheduleRoot			string
	OrderingCommitment		string
	UsesPeerLocalValidityInput	bool
	LiveNetworkStateRead		bool
	TransactionCount		uint64
	MessageCount			uint64
}

type ABCIProcessProposalInput struct {
	Height				uint64
	Proposal			ABCIProposalPlan
	ExpectedScheduleRoot		string
	ExpectedOrderingCommitment	string
	DependsOnPeerLocalData		bool
	VerifiesOrderingCommitment	bool
}

type ABCIFinalizeBlockInput struct {
	Height			uint64
	Proposal		ABCIProposalPlan
	ExecutionMessages	[]ExecutionZoneMessage
	Receipts		[]CrossZoneReceipt
	LiveNetworkStateRead	bool
}

type ABCIFinalizeBlockResult struct {
	Height			uint64
	ExecutedMessageIDs	[]string
	ExecutionRoot		string
	ReceiptsRoot		string
	ScheduleRoot		string
}

func DefaultCosmosCometBFTCompatibilityPlan() CosmosCometBFTCompatibilityPlan {
	return CosmosCometBFTCompatibilityPlan{
		Adapter:	DefaultAetherNetworkingAdapter(),
		Surfaces: []CompatibilitySurface{
			CompatibilityCometBFTP2P,
			CompatibilityCometBFTConsensusMessages,
			CompatibilityCosmosSDKTxFlow,
			CompatibilityABCIPlusPlus,
			CompatibilityBlockSTM,
			CompatibilityGRPC,
			CompatibilityREST,
			CompatibilityRPC,
			CompatibilityStateSyncSnapshots,
		},
		Phases: []ABCILifecyclePhase{
			ABCIPrepareProposal,
			ABCIProcessProposal,
			ABCIFinalizeBlock,
		},
		PreserveCometBFTTransport:		true,
		PreserveCometBFTConsensusMessages:	true,
		PreserveCosmosSDKTransactionFlow:	true,
		SupportsBlockSTMExecutionGrouping:	true,
		SupportsGRPCExternalAPI:		true,
		SupportsRESTExternalAPI:		true,
		SupportsRPCExternalAPI:			true,
		SupportsStateSyncAndSnapshotMechanism:	true,
	}
}

func ValidateCosmosCometBFTCompatibilityPlan(plan CosmosCometBFTCompatibilityPlan) error {
	if err := ValidateAetherNetworkingAdapter(plan.Adapter); err != nil {
		return err
	}
	requiredSurfaces := []CompatibilitySurface{
		CompatibilityCometBFTP2P,
		CompatibilityCometBFTConsensusMessages,
		CompatibilityCosmosSDKTxFlow,
		CompatibilityABCIPlusPlus,
		CompatibilityBlockSTM,
		CompatibilityGRPC,
		CompatibilityREST,
		CompatibilityRPC,
		CompatibilityStateSyncSnapshots,
	}
	for _, surface := range requiredSurfaces {
		if !hasCompatibilitySurface(plan.Surfaces, surface) {
			return fmt.Errorf("networking compatibility missing surface %s", surface)
		}
	}
	for _, phase := range []ABCILifecyclePhase{ABCIPrepareProposal, ABCIProcessProposal, ABCIFinalizeBlock} {
		if !hasABCIPhase(plan.Phases, phase) {
			return fmt.Errorf("networking compatibility missing ABCI++ phase %s", phase)
		}
	}
	if !plan.PreserveCometBFTTransport {
		return errors.New("networking compatibility must preserve CometBFT P2P transport")
	}
	if !plan.PreserveCometBFTConsensusMessages {
		return errors.New("networking compatibility must preserve CometBFT consensus messages")
	}
	if !plan.PreserveCosmosSDKTransactionFlow {
		return errors.New("networking compatibility must preserve Cosmos SDK transaction flow")
	}
	if !plan.SupportsBlockSTMExecutionGrouping {
		return errors.New("networking compatibility must support BlockSTM execution grouping")
	}
	if !plan.SupportsGRPCExternalAPI || !plan.SupportsRESTExternalAPI || !plan.SupportsRPCExternalAPI {
		return errors.New("networking compatibility must preserve gRPC, REST, and RPC APIs")
	}
	if !plan.SupportsStateSyncAndSnapshotMechanism {
		return errors.New("networking compatibility must preserve state sync and snapshots")
	}
	return nil
}

func BuildPrepareProposalCompatibility(input ABCIPrepareProposalInput) (ABCIProposalPlan, error) {
	if input.Height == 0 {
		return ABCIProposalPlan{}, errors.New("networking PrepareProposal height must be positive")
	}
	if err := ValidateAetherNetworkingAdapter(input.Adapter); err != nil {
		return ABCIProposalPlan{}, err
	}
	hintsBySchedule, err := normalizeAndIndexANAProposalHints(input.Hints)
	if err != nil {
		return ABCIProposalPlan{}, err
	}
	schedules, err := normalizeCommittedExecutionSchedules(input.Schedules)
	if err != nil {
		return ABCIProposalPlan{}, err
	}
	if len(schedules) == 0 {
		return ABCIProposalPlan{}, errors.New("networking PrepareProposal requires deterministic execution schedules")
	}
	groups := make([]ABCIProposalGroup, 0, len(schedules))
	for _, schedule := range schedules {
		group := proposalGroupFromSchedule(schedule, hintsBySchedule[schedule.ScheduleID])
		groups = append(groups, group)
	}
	sortProposalGroups(groups)
	plan := ABCIProposalPlan{
		Height:		input.Height,
		Phase:		ABCIPrepareProposal,
		Groups:		groups,
		ScheduleRoot:	ComputeABCIProposalScheduleRoot(groups),
	}
	plan.OrderingCommitment = ComputeABCIOrderingCommitment(groups)
	plan.TransactionCount, plan.MessageCount = countProposalItems(groups)
	return plan, plan.Validate()
}

func ProcessProposalCompatibility(input ABCIProcessProposalInput) (ABCIProposalPlan, error) {
	if input.Height == 0 {
		return ABCIProposalPlan{}, errors.New("networking ProcessProposal height must be positive")
	}
	proposal := NormalizeABCIProposalPlan(input.Proposal)
	if proposal.Height != input.Height {
		return ABCIProposalPlan{}, errors.New("networking ProcessProposal height mismatch")
	}
	if err := proposal.Validate(); err != nil {
		return ABCIProposalPlan{}, err
	}
	if input.DependsOnPeerLocalData || proposal.UsesPeerLocalValidityInput {
		return ABCIProposalPlan{}, errors.New("networking ProcessProposal must not depend on peer-local data")
	}
	if proposal.LiveNetworkStateRead {
		return ABCIProposalPlan{}, errors.New("networking ProcessProposal must not read live network state")
	}
	if !input.VerifiesOrderingCommitment {
		return ABCIProposalPlan{}, errors.New("networking ProcessProposal must verify message ordering commitments")
	}
	if input.ExpectedScheduleRoot != "" && normalizeHashText(input.ExpectedScheduleRoot) != proposal.ScheduleRoot {
		return ABCIProposalPlan{}, errors.New("networking ProcessProposal schedule root mismatch")
	}
	if input.ExpectedOrderingCommitment != "" && normalizeHashText(input.ExpectedOrderingCommitment) != proposal.OrderingCommitment {
		return ABCIProposalPlan{}, errors.New("networking ProcessProposal ordering commitment mismatch")
	}
	proposal.Phase = ABCIProcessProposal
	return proposal, nil
}

func FinalizeBlockCompatibility(input ABCIFinalizeBlockInput) (ABCIFinalizeBlockResult, error) {
	if input.Height == 0 {
		return ABCIFinalizeBlockResult{}, errors.New("networking FinalizeBlock height must be positive")
	}
	proposal := NormalizeABCIProposalPlan(input.Proposal)
	if proposal.Height != input.Height {
		return ABCIFinalizeBlockResult{}, errors.New("networking FinalizeBlock height mismatch")
	}
	if err := proposal.Validate(); err != nil {
		return ABCIFinalizeBlockResult{}, err
	}
	if input.LiveNetworkStateRead || proposal.LiveNetworkStateRead {
		return ABCIFinalizeBlockResult{}, errors.New("networking FinalizeBlock must ignore live network state")
	}
	allowed := proposalMessageIndex(proposal.Groups)
	executed := make([]string, 0, len(input.ExecutionMessages))
	for _, msg := range input.ExecutionMessages {
		msg = NormalizeExecutionZoneMessage(msg)
		if msg.ConsensusScheduleID == "" {
			return ABCIFinalizeBlockResult{}, errors.New("networking FinalizeBlock executes committed messages only")
		}
		messages, found := allowed[msg.ConsensusScheduleID]
		if !found {
			return ABCIFinalizeBlockResult{}, errors.New("networking FinalizeBlock executes committed messages only")
		}
		if msg.Message.MessageID == "" {
			return ABCIFinalizeBlockResult{}, errors.New("networking FinalizeBlock execution message requires mesh message id")
		}
		if _, found := messages[msg.Message.MessageID]; !found {
			return ABCIFinalizeBlockResult{}, errors.New("networking FinalizeBlock message is not in committed schedule")
		}
		executed = append(executed, msg.Message.MessageID)
	}
	sort.Strings(executed)
	receiptIDs := make([]string, 0, len(input.Receipts))
	for _, receipt := range input.Receipts {
		receipt = NormalizeCrossZoneReceipt(receipt)
		if err := receipt.Validate(); err != nil {
			return ABCIFinalizeBlockResult{}, err
		}
		receiptIDs = append(receiptIDs, receipt.ReceiptID)
	}
	sort.Strings(receiptIDs)
	return ABCIFinalizeBlockResult{
		Height:			input.Height,
		ExecutedMessageIDs:	executed,
		ExecutionRoot:		HashParts(append([]string{"abci-finalize-execution-root"}, executed...)...),
		ReceiptsRoot:		HashParts(append([]string{"abci-finalize-receipts-root"}, receiptIDs...)...),
		ScheduleRoot:		proposal.ScheduleRoot,
	}, nil
}

func (h ANAProposalHint) Normalize() ANAProposalHint {
	h.HintID = normalizeHashText(h.HintID)
	h.ScheduleID = normalizeHashText(h.ScheduleID)
	h.ScheduleHash = normalizeHashText(h.ScheduleHash)
	h.ZoneID = strings.TrimSpace(h.ZoneID)
	h.ShardID = strings.TrimSpace(h.ShardID)
	h.BlockSTMGroupID = normalizeHashText(h.BlockSTMGroupID)
	h.ExecutionGroupID = normalizeHashText(h.ExecutionGroupID)
	h.DeterministicHintProof = normalizeHashText(h.DeterministicHintProof)
	return h
}

func (h ANAProposalHint) Validate() error {
	hint := h.Normalize()
	if err := ValidateHash("networking ANA proposal hint id", hint.HintID); err != nil {
		return err
	}
	if err := ValidateHash("networking ANA proposal hint schedule id", hint.ScheduleID); err != nil {
		return err
	}
	if err := ValidateHash("networking ANA proposal hint schedule hash", hint.ScheduleHash); err != nil {
		return err
	}
	if hint.ZoneID == "" || hint.ShardID == "" {
		return errors.New("networking ANA proposal hint requires zone and shard")
	}
	if hint.PeerLocal && (hint.UsedForValidity || hint.UsedForOrdering) {
		return errors.New("networking ANA peer-local hints cannot affect proposal validity or ordering")
	}
	if hint.AdvisoryOnly && hint.UsedForValidity {
		return errors.New("networking ANA advisory hints cannot affect proposal validity")
	}
	if hint.UsedForValidity || hint.UsedForOrdering {
		if !hint.CommittedStateDerived {
			return errors.New("networking ANA proposal hint used for validity or ordering must be committed-state derived")
		}
		if !IsConsensusSafeDeterminismSource(hint.DeterminismSource) {
			return fmt.Errorf("networking ANA proposal hint requires deterministic source, got %q", hint.DeterminismSource)
		}
	}
	if hint.DeterministicHintProof != "" && hint.DeterminismSource != DeterminismDeterministicProof {
		return errors.New("networking ANA deterministic hint proof requires deterministic proof source")
	}
	if hint.DeterminismSource == DeterminismDeterministicProof {
		if err := ValidateHash("networking ANA deterministic hint proof", hint.DeterministicHintProof); err != nil {
			return err
		}
	}
	return nil
}

func NormalizeABCIProposalPlan(plan ABCIProposalPlan) ABCIProposalPlan {
	plan.ScheduleRoot = normalizeHashText(plan.ScheduleRoot)
	plan.OrderingCommitment = normalizeHashText(plan.OrderingCommitment)
	for i := range plan.Groups {
		plan.Groups[i] = NormalizeABCIProposalGroup(plan.Groups[i])
	}
	sortProposalGroups(plan.Groups)
	plan.TransactionCount, plan.MessageCount = countProposalItems(plan.Groups)
	return plan
}

func NormalizeABCIProposalGroup(group ABCIProposalGroup) ABCIProposalGroup {
	group.GroupID = normalizeHashText(group.GroupID)
	group.ZoneID = strings.TrimSpace(group.ZoneID)
	group.ShardID = strings.TrimSpace(group.ShardID)
	group.ScheduleID = normalizeHashText(group.ScheduleID)
	group.ScheduleHash = normalizeHashText(group.ScheduleHash)
	group.TransactionIDs = normalizeHashSet(group.TransactionIDs)
	group.MessageIDs = normalizeHashSet(group.MessageIDs)
	group.HintIDs = normalizeHashSet(group.HintIDs)
	group.BlockSTMGroupID = normalizeHashText(group.BlockSTMGroupID)
	return group
}

func (p ABCIProposalPlan) Validate() error {
	plan := NormalizeABCIProposalPlan(p)
	if plan.Height == 0 {
		return errors.New("networking ABCI proposal plan height must be positive")
	}
	if !IsABCILifecyclePhase(plan.Phase) {
		return fmt.Errorf("unknown networking ABCI++ phase %q", plan.Phase)
	}
	if len(plan.Groups) == 0 {
		return errors.New("networking ABCI proposal plan requires groups")
	}
	if plan.UsesPeerLocalValidityInput {
		return errors.New("networking ABCI proposal plan must not use peer-local validity input")
	}
	if plan.LiveNetworkStateRead {
		return errors.New("networking ABCI proposal plan must not read live network state")
	}
	for _, group := range plan.Groups {
		if err := group.Validate(); err != nil {
			return err
		}
	}
	if plan.ScheduleRoot != ComputeABCIProposalScheduleRoot(plan.Groups) {
		return errors.New("networking ABCI proposal schedule root mismatch")
	}
	if plan.OrderingCommitment != ComputeABCIOrderingCommitment(plan.Groups) {
		return errors.New("networking ABCI proposal ordering commitment mismatch")
	}
	return nil
}

func (g ABCIProposalGroup) Validate() error {
	group := NormalizeABCIProposalGroup(g)
	if err := ValidateHash("networking ABCI proposal group id", group.GroupID); err != nil {
		return err
	}
	if group.GroupID != ComputeABCIProposalGroupID(group) {
		return errors.New("networking ABCI proposal group id mismatch")
	}
	if group.ZoneID == "" || group.ShardID == "" {
		return errors.New("networking ABCI proposal group requires zone and shard")
	}
	if err := ValidateHash("networking ABCI proposal schedule id", group.ScheduleID); err != nil {
		return err
	}
	if err := ValidateHash("networking ABCI proposal schedule hash", group.ScheduleHash); err != nil {
		return err
	}
	if len(group.TransactionIDs) == 0 && len(group.MessageIDs) == 0 {
		return errors.New("networking ABCI proposal group requires transactions or messages")
	}
	for _, id := range append(append([]string(nil), group.TransactionIDs...), group.MessageIDs...) {
		if err := ValidateHash("networking ABCI proposal item id", id); err != nil {
			return err
		}
	}
	for _, id := range group.HintIDs {
		if err := ValidateHash("networking ABCI proposal hint id", id); err != nil {
			return err
		}
	}
	if !group.Deterministic {
		return errors.New("networking ABCI proposal group must be deterministic")
	}
	return nil
}

func ComputeABCIProposalGroupID(group ABCIProposalGroup) string {
	group = NormalizeABCIProposalGroup(group)
	parts := []string{
		"abci-proposal-group",
		group.ZoneID,
		group.ShardID,
		group.ScheduleID,
		group.ScheduleHash,
		group.BlockSTMGroupID,
	}
	parts = append(parts, group.TransactionIDs...)
	parts = append(parts, group.MessageIDs...)
	parts = append(parts, group.HintIDs...)
	return HashParts(parts...)
}

func ComputeABCIProposalScheduleRoot(groups []ABCIProposalGroup) string {
	normalized := normalizeProposalGroups(groups)
	parts := []string{"abci-proposal-schedule-root"}
	for _, group := range normalized {
		parts = append(parts, group.GroupID, group.ScheduleID, group.ScheduleHash)
	}
	return HashParts(parts...)
}

func ComputeABCIOrderingCommitment(groups []ABCIProposalGroup) string {
	normalized := normalizeProposalGroups(groups)
	parts := []string{"abci-proposal-ordering-commitment"}
	for _, group := range normalized {
		parts = append(parts, group.GroupID)
		parts = append(parts, group.TransactionIDs...)
		parts = append(parts, group.MessageIDs...)
	}
	return HashParts(parts...)
}

func IsCompatibilitySurface(surface CompatibilitySurface) bool {
	switch surface {
	case CompatibilityCometBFTP2P,
		CompatibilityCometBFTConsensusMessages,
		CompatibilityCosmosSDKTxFlow,
		CompatibilityABCIPlusPlus,
		CompatibilityBlockSTM,
		CompatibilityGRPC,
		CompatibilityREST,
		CompatibilityRPC,
		CompatibilityStateSyncSnapshots:
		return true
	default:
		return false
	}
}

func IsABCILifecyclePhase(phase ABCILifecyclePhase) bool {
	switch phase {
	case ABCIPrepareProposal, ABCIProcessProposal, ABCIFinalizeBlock:
		return true
	default:
		return false
	}
}

func normalizeAndIndexANAProposalHints(hints []ANAProposalHint) (map[string][]ANAProposalHint, error) {
	bySchedule := make(map[string][]ANAProposalHint)
	seen := make(map[string]struct{}, len(hints))
	for _, hint := range hints {
		hint = hint.Normalize()
		if err := hint.Validate(); err != nil {
			return nil, err
		}
		if _, found := seen[hint.HintID]; found {
			return nil, errors.New("networking ANA duplicate proposal hint")
		}
		seen[hint.HintID] = struct{}{}
		bySchedule[hint.ScheduleID] = append(bySchedule[hint.ScheduleID], hint)
	}
	for scheduleID := range bySchedule {
		sort.SliceStable(bySchedule[scheduleID], func(i, j int) bool {
			left := bySchedule[scheduleID][i]
			right := bySchedule[scheduleID][j]
			if left.Priority != right.Priority {
				return left.Priority < right.Priority
			}
			return left.HintID < right.HintID
		})
	}
	return bySchedule, nil
}

func normalizeCommittedExecutionSchedules(schedules []ExecutionMessageSchedule) ([]ExecutionMessageSchedule, error) {
	out := make([]ExecutionMessageSchedule, 0, len(schedules))
	seen := make(map[string]struct{}, len(schedules))
	for _, schedule := range schedules {
		schedule = NormalizeExecutionMessageSchedule(schedule)
		if err := schedule.Validate(); err != nil {
			return nil, err
		}
		if !schedule.Committed {
			return nil, errors.New("networking PrepareProposal requires committed deterministic state schedules")
		}
		if _, found := seen[schedule.ScheduleID]; found {
			return nil, errors.New("networking duplicate execution schedule")
		}
		seen[schedule.ScheduleID] = struct{}{}
		out = append(out, schedule)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ZoneID != out[j].ZoneID {
			return out[i].ZoneID < out[j].ZoneID
		}
		if out[i].ShardID != out[j].ShardID {
			return out[i].ShardID < out[j].ShardID
		}
		return out[i].ScheduleID < out[j].ScheduleID
	})
	return out, nil
}

func proposalGroupFromSchedule(schedule ExecutionMessageSchedule, hints []ANAProposalHint) ABCIProposalGroup {
	group := ABCIProposalGroup{
		ZoneID:		schedule.ZoneID,
		ShardID:	schedule.ShardID,
		ScheduleID:	schedule.ScheduleID,
		ScheduleHash:	schedule.ScheduleHash,
		TransactionIDs:	append([]string(nil), schedule.TransactionIDs...),
		MessageIDs:	append([]string(nil), schedule.MessageIDs...),
		Deterministic:	true,
	}
	for _, hint := range hints {
		if hint.ScheduleHash != schedule.ScheduleHash || hint.ZoneID != schedule.ZoneID || hint.ShardID != schedule.ShardID {
			continue
		}
		group.HintIDs = append(group.HintIDs, hint.HintID)
		if hint.UsedForOrdering && hint.BlockSTMGroupID != "" && group.BlockSTMGroupID == "" {
			group.BlockSTMGroupID = hint.BlockSTMGroupID
		}
	}
	group = NormalizeABCIProposalGroup(group)
	group.GroupID = ComputeABCIProposalGroupID(group)
	return group
}

func normalizeProposalGroups(groups []ABCIProposalGroup) []ABCIProposalGroup {
	out := make([]ABCIProposalGroup, len(groups))
	for i, group := range groups {
		out[i] = NormalizeABCIProposalGroup(group)
		if out[i].GroupID == "" {
			out[i].GroupID = ComputeABCIProposalGroupID(out[i])
		}
	}
	sortProposalGroups(out)
	return out
}

func sortProposalGroups(groups []ABCIProposalGroup) {
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].ZoneID != groups[j].ZoneID {
			return groups[i].ZoneID < groups[j].ZoneID
		}
		if groups[i].ShardID != groups[j].ShardID {
			return groups[i].ShardID < groups[j].ShardID
		}
		return groups[i].ScheduleID < groups[j].ScheduleID
	})
}

func countProposalItems(groups []ABCIProposalGroup) (uint64, uint64) {
	var txs uint64
	var messages uint64
	for _, group := range groups {
		txs += uint64(len(group.TransactionIDs))
		messages += uint64(len(group.MessageIDs))
	}
	return txs, messages
}

func proposalMessageIndex(groups []ABCIProposalGroup) map[string]map[string]struct{} {
	index := make(map[string]map[string]struct{}, len(groups))
	for _, group := range groups {
		group = NormalizeABCIProposalGroup(group)
		messages := make(map[string]struct{}, len(group.MessageIDs))
		for _, id := range group.MessageIDs {
			messages[id] = struct{}{}
		}
		index[group.ScheduleID] = messages
	}
	return index
}

func hasCompatibilitySurface(surfaces []CompatibilitySurface, required CompatibilitySurface) bool {
	for _, surface := range surfaces {
		if !IsCompatibilitySurface(surface) {
			continue
		}
		if surface == required {
			return true
		}
	}
	return false
}

func hasABCIPhase(phases []ABCILifecyclePhase, required ABCILifecyclePhase) bool {
	for _, phase := range phases {
		if phase == required {
			return true
		}
	}
	return false
}
