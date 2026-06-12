package types

import (
	"errors"
	"fmt"
	"sort"
)

type KernelABCIPhase string

const (
	KernelPhasePrepareProposal	KernelABCIPhase	= "PREPARE_PROPOSAL"
	KernelPhaseProcessProposal	KernelABCIPhase	= "PROCESS_PROPOSAL"
	KernelPhaseFinalizeBlock	KernelABCIPhase	= "FINALIZE_BLOCK"
	KernelPhaseCommit		KernelABCIPhase	= "COMMIT"
)

type KernelConsensusContext struct {
	ChainID		string
	Height		uint64
	BlockTimeUnix	int64
}

type KernelTimestampBounds struct {
	PreviousBlockTimeUnix	int64
	MaxForwardDriftSeconds	int64
}

type KernelBlockPlan struct {
	Height			uint64
	ChainID			string
	Phase			KernelABCIPhase
	Schedule		ProposalSchedule
	ProposalRoot		string
	RoutingRoot		string
	PreviousGlobalRoot	string
	ParamsHash		string
	PlanHash		string
}

type KernelBlockHeaderCommitment struct {
	Height		uint64
	TimeUnix	int64
	PreviousAppHash	string
	ZonesRoot	string
	MessagesRoot	string
	ReceiptsRoot	string
	HeaderHash	string
}

type KernelFinalizationInput struct {
	ZoneCommitments	[]ZoneCommitment
	Receipts	[]ExecutionReceipt
	Contributions	RootContributions
}

type KernelFinalization struct {
	Height		uint64
	ChainID		string
	Phase		KernelABCIPhase
	PlanHash	string
	Header		KernelBlockHeaderCommitment
	GlobalRoot	GlobalStateRoot
	RootSnapshot	RootSnapshot
	ReceiptsRoot	string
	ReceiptCount	uint64
	CommitmentCount	uint64
	FinalityHash	string
}

func PrepareKernelProposal(ctx KernelConsensusContext, state CoreState, items []ProposalItem) (KernelBlockPlan, error) {
	if err := ctx.Validate(); err != nil {
		return KernelBlockPlan{}, err
	}
	if err := state.Validate(); err != nil {
		return KernelBlockPlan{}, err
	}
	schedule, err := BuildProposalSchedule(ctx.Height, items, state.Params)
	if err != nil {
		return KernelBlockPlan{}, err
	}
	if err := ValidateProposalScheduleForState(schedule, state); err != nil {
		return KernelBlockPlan{}, err
	}
	proposalRoot, err := ComputeProposalScheduleRoot(schedule)
	if err != nil {
		return KernelBlockPlan{}, err
	}
	plan := KernelBlockPlan{
		Height:		ctx.Height,
		ChainID:	ctx.ChainID,
		Phase:		KernelPhasePrepareProposal,
		Schedule:	schedule,
		ProposalRoot:	proposalRoot,
		RoutingRoot:	EmptyRootHash,
		ParamsHash:	ComputeAetraCoreParamsHash(state.Params),
	}
	if table, found := state.LatestRoutingTableAtHeight(ctx.Height); found {
		plan.RoutingRoot = table.TableHash
	}
	if root, found := latestGlobalRootBefore(state.GlobalRoots, ctx.Height); found {
		plan.PreviousGlobalRoot = root.GlobalRoot
	} else {
		plan.PreviousGlobalRoot = EmptyRootHash
	}
	plan.PlanHash = ComputeKernelBlockPlanHash(plan)
	return plan, ValidateKernelBlockPlanForState(ctx, state, plan)
}

func ProcessKernelProposal(ctx KernelConsensusContext, state CoreState, plan KernelBlockPlan) error {
	if err := ValidateKernelBlockPlanForState(ctx, state, plan); err != nil {
		return err
	}
	if plan.Phase != KernelPhasePrepareProposal && plan.Phase != KernelPhaseProcessProposal {
		return errors.New("aetracore kernel proposal is not in a processable phase")
	}
	return nil
}

func ProcessKernelProposalWithTimestampBounds(ctx KernelConsensusContext, state CoreState, plan KernelBlockPlan, bounds KernelTimestampBounds) error {
	if err := ValidateKernelTimestampBounds(ctx, bounds); err != nil {
		return err
	}
	return ProcessKernelProposal(ctx, state, plan)
}

func FinalizeKernelBlock(ctx KernelConsensusContext, state CoreState, plan KernelBlockPlan, input KernelFinalizationInput) (CoreState, KernelFinalization, error) {
	if err := ProcessKernelProposal(ctx, state, plan); err != nil {
		return CoreState{}, KernelFinalization{}, err
	}
	receiptsRoot, err := ComputeExecutionReceiptsRoot(input.Receipts)
	if err != nil {
		return CoreState{}, KernelFinalization{}, err
	}
	if err := input.Contributions.Validate(); err != nil {
		return CoreState{}, KernelFinalization{}, err
	}
	if input.Contributions.ReceiptsRoot != receiptsRoot {
		return CoreState{}, KernelFinalization{}, errors.New("aetracore kernel receipts root contribution mismatch")
	}
	next := state.Export()
	commitments := append([]ZoneCommitment(nil), input.ZoneCommitments...)
	sortZoneCommitments(commitments)
	for _, commitment := range commitments {
		if commitment.Height != ctx.Height {
			return CoreState{}, KernelFinalization{}, errors.New("aetracore kernel commitment height mismatch")
		}
		next, err = AppendZoneCommitment(next, commitment)
		if err != nil {
			return CoreState{}, KernelFinalization{}, err
		}
	}
	next, snapshot, err := CommitBlockRootsWithContributions(next, ctx.Height, input.Contributions)
	if err != nil {
		return CoreState{}, KernelFinalization{}, err
	}
	globalRoot, found := next.GlobalRootByHeight(ctx.Height)
	if !found {
		return CoreState{}, KernelFinalization{}, errors.New("aetracore kernel global root was not committed")
	}
	header, err := NewKernelBlockHeaderCommitment(ctx, plan.PreviousGlobalRoot, globalRoot, snapshot)
	if err != nil {
		return CoreState{}, KernelFinalization{}, err
	}
	finalization := KernelFinalization{
		Height:			ctx.Height,
		ChainID:		ctx.ChainID,
		Phase:			KernelPhaseFinalizeBlock,
		PlanHash:		plan.PlanHash,
		Header:			header,
		GlobalRoot:		globalRoot,
		RootSnapshot:		snapshot,
		ReceiptsRoot:		receiptsRoot,
		ReceiptCount:		uint64(len(input.Receipts)),
		CommitmentCount:	uint64(len(commitments)),
	}
	finalization.FinalityHash = ComputeKernelFinalizationHash(finalization)
	return next.Export(), finalization, finalization.Validate()
}

func CommitKernelBlock(finalization KernelFinalization) (FinalityRoots, error) {
	if err := finalization.Validate(); err != nil {
		return FinalityRoots{}, err
	}
	return finalization.RootSnapshot.Finality, nil
}

func ValidateKernelBlockPlanForState(ctx KernelConsensusContext, state CoreState, plan KernelBlockPlan) error {
	if err := ctx.Validate(); err != nil {
		return err
	}
	if err := state.Validate(); err != nil {
		return err
	}
	if plan.Height != ctx.Height || plan.ChainID != ctx.ChainID {
		return errors.New("aetracore kernel plan consensus context mismatch")
	}
	if plan.Phase != KernelPhasePrepareProposal && plan.Phase != KernelPhaseProcessProposal {
		return fmt.Errorf("unknown aetracore kernel proposal phase %q", plan.Phase)
	}
	if err := ValidateHash("aetracore kernel proposal root", plan.ProposalRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel routing root", plan.RoutingRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel previous global root", plan.PreviousGlobalRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel params hash", plan.ParamsHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel plan hash", plan.PlanHash); err != nil {
		return err
	}
	if err := ValidateProposalScheduleForState(plan.Schedule, state); err != nil {
		return err
	}
	proposalRoot, err := ComputeProposalScheduleRoot(plan.Schedule)
	if err != nil {
		return err
	}
	if proposalRoot != plan.ProposalRoot {
		return errors.New("aetracore kernel proposal root mismatch")
	}
	routingRoot := EmptyRootHash
	if table, found := state.LatestRoutingTableAtHeight(ctx.Height); found {
		routingRoot = table.TableHash
	}
	if routingRoot != plan.RoutingRoot {
		return errors.New("aetracore kernel routing root mismatch")
	}
	previousRoot := EmptyRootHash
	if root, found := latestGlobalRootBefore(state.GlobalRoots, ctx.Height); found {
		previousRoot = root.GlobalRoot
	}
	if previousRoot != plan.PreviousGlobalRoot {
		return errors.New("aetracore kernel previous global root mismatch")
	}
	if paramsHash := ComputeAetraCoreParamsHash(state.Params); paramsHash != plan.ParamsHash {
		return errors.New("aetracore kernel params hash mismatch")
	}
	if expected := ComputeKernelBlockPlanHash(plan); expected != plan.PlanHash {
		return fmt.Errorf("aetracore kernel plan hash mismatch: expected %s", expected)
	}
	return nil
}

func (ctx KernelConsensusContext) Validate() error {
	if err := validatePolicyID("aetracore kernel chain id", ctx.ChainID); err != nil {
		return err
	}
	if ctx.Height == 0 {
		return errors.New("aetracore kernel consensus height must be positive")
	}
	if ctx.BlockTimeUnix < 0 {
		return errors.New("aetracore kernel block time must be consensus supplied")
	}
	return nil
}

func ValidateKernelTimestampBounds(ctx KernelConsensusContext, bounds KernelTimestampBounds) error {
	if err := ctx.Validate(); err != nil {
		return err
	}
	if bounds.PreviousBlockTimeUnix < 0 {
		return errors.New("aetracore kernel previous block time must be consensus supplied")
	}
	if bounds.MaxForwardDriftSeconds <= 0 {
		return errors.New("aetracore kernel timestamp max forward drift must be positive")
	}
	if ctx.BlockTimeUnix <= bounds.PreviousBlockTimeUnix {
		return errors.New("aetracore kernel block time must be after previous consensus time")
	}
	if ctx.BlockTimeUnix-bounds.PreviousBlockTimeUnix > bounds.MaxForwardDriftSeconds {
		return errors.New("aetracore kernel block time is outside allowed consensus bounds")
	}
	return nil
}

func NewKernelBlockHeaderCommitment(ctx KernelConsensusContext, previousAppHash string, root GlobalStateRoot, snapshot RootSnapshot) (KernelBlockHeaderCommitment, error) {
	if err := ctx.Validate(); err != nil {
		return KernelBlockHeaderCommitment{}, err
	}
	if err := ValidateHash("aetracore kernel previous app hash", previousAppHash); err != nil {
		return KernelBlockHeaderCommitment{}, err
	}
	if err := root.ValidateHash(); err != nil {
		return KernelBlockHeaderCommitment{}, err
	}
	if err := snapshot.Validate(); err != nil {
		return KernelBlockHeaderCommitment{}, err
	}
	if root.Height != ctx.Height || snapshot.Height != ctx.Height {
		return KernelBlockHeaderCommitment{}, errors.New("aetracore kernel header height mismatch")
	}
	header := KernelBlockHeaderCommitment{
		Height:			ctx.Height,
		TimeUnix:		ctx.BlockTimeUnix,
		PreviousAppHash:	previousAppHash,
		ZonesRoot:		root.ZonesRoot,
		MessagesRoot:		snapshot.Finality.GlobalMessageRoot,
		ReceiptsRoot:		snapshot.Finality.ExecutionReceiptRoot,
	}
	header.HeaderHash = ComputeKernelBlockHeaderHash(header)
	return header, header.Validate()
}

func (h KernelBlockHeaderCommitment) Validate() error {
	if h.Height == 0 {
		return errors.New("aetracore kernel header height must be positive")
	}
	if h.TimeUnix < 0 {
		return errors.New("aetracore kernel header time must be consensus supplied")
	}
	if err := ValidateHash("aetracore kernel header previous app hash", h.PreviousAppHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel header zones root", h.ZonesRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel header messages root", h.MessagesRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel header receipts root", h.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel header hash", h.HeaderHash); err != nil {
		return err
	}
	if expected := ComputeKernelBlockHeaderHash(h); expected != h.HeaderHash {
		return fmt.Errorf("aetracore kernel header hash mismatch: expected %s", expected)
	}
	return nil
}

func (f KernelFinalization) Validate() error {
	if f.Height == 0 {
		return errors.New("aetracore kernel finalization height must be positive")
	}
	if err := validatePolicyID("aetracore kernel finalization chain id", f.ChainID); err != nil {
		return err
	}
	if f.Phase != KernelPhaseFinalizeBlock {
		return fmt.Errorf("unknown aetracore kernel finalization phase %q", f.Phase)
	}
	if err := ValidateHash("aetracore kernel finalization plan hash", f.PlanHash); err != nil {
		return err
	}
	if err := f.GlobalRoot.ValidateHash(); err != nil {
		return err
	}
	if err := f.RootSnapshot.Validate(); err != nil {
		return err
	}
	if err := f.Header.Validate(); err != nil {
		return err
	}
	if f.GlobalRoot.Height != f.Height || f.RootSnapshot.Height != f.Height {
		return errors.New("aetracore kernel finalization height mismatch")
	}
	if f.Header.Height != f.Height {
		return errors.New("aetracore kernel finalization header height mismatch")
	}
	if f.Header.ZonesRoot != f.GlobalRoot.ZonesRoot {
		return errors.New("aetracore kernel header zones root mismatch")
	}
	if f.Header.MessagesRoot != f.RootSnapshot.Finality.GlobalMessageRoot {
		return errors.New("aetracore kernel header messages root mismatch")
	}
	if f.Header.ReceiptsRoot != f.RootSnapshot.Finality.ExecutionReceiptRoot {
		return errors.New("aetracore kernel header receipts root mismatch")
	}
	if f.RootSnapshot.Finality.GlobalStateRoot != f.GlobalRoot.GlobalRoot {
		return errors.New("aetracore kernel finality global root mismatch")
	}
	if f.RootSnapshot.Finality.ExecutionReceiptRoot != f.ReceiptsRoot {
		return errors.New("aetracore kernel finality receipt root mismatch")
	}
	if f.GlobalRoot.ReceiptsRoot != f.ReceiptsRoot {
		return errors.New("aetracore kernel global receipt root mismatch")
	}
	if err := ValidateHash("aetracore kernel finalization receipts root", f.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore kernel finality hash", f.FinalityHash); err != nil {
		return err
	}
	if expected := ComputeKernelFinalizationHash(f); expected != f.FinalityHash {
		return fmt.Errorf("aetracore kernel finality hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeProposalItemHash(item ProposalItem) string {
	return hashParts(
		"aetra-aek-proposal-item-v1",
		string(item.ZoneID),
		string(item.ShardID),
		item.TxHash,
		fmt.Sprint(item.PriorityClass),
		fmt.Sprint(item.AdmissionHeight),
		fmt.Sprint(item.TxIndex),
		fmt.Sprint(item.MessageIndex),
	)
}

func ComputeProposalScheduleRoot(schedule ProposalSchedule) (string, error) {
	if err := schedule.Validate(); err != nil {
		return "", err
	}
	groups := append([]ProposalGroup(nil), schedule.Groups...)
	sort.SliceStable(groups, func(i, j int) bool {
		return compareProposalGroupKey(groups[i], groups[j]) < 0
	})
	parts := []string{"aetra-aek-proposal-schedule-root-v1", fmt.Sprint(schedule.Height), fmt.Sprint(len(groups))}
	for _, group := range groups {
		items := append([]ProposalItem(nil), group.Items...)
		sortProposalItems(items)
		parts = append(parts, string(group.ZoneID), string(group.ShardID), fmt.Sprint(len(items)))
		for _, item := range items {
			parts = append(parts, ComputeProposalItemHash(item))
		}
	}
	return hashParts(parts...), nil
}

func ComputeExecutionReceiptsRoot(receipts []ExecutionReceipt) (string, error) {
	ordered := append([]ExecutionReceipt(nil), receipts...)
	sortExecutionReceipts(ordered)
	parts := []string{"aetra-aek-execution-receipts-root-v1", fmt.Sprint(len(ordered))}
	seen := make(map[string]struct{}, len(ordered))
	var previous ExecutionReceipt
	for i, receipt := range ordered {
		if err := receipt.Validate(); err != nil {
			return "", err
		}
		if _, found := seen[receipt.ReceiptHash]; found {
			return "", errors.New("duplicate aetracore execution receipt")
		}
		seen[receipt.ReceiptHash] = struct{}{}
		if i > 0 && compareExecutionReceipts(previous, receipt) >= 0 {
			return "", errors.New("aetracore execution receipts must be sorted canonically")
		}
		parts = append(parts, receipt.ReceiptHash)
		previous = receipt
	}
	return hashParts(parts...), nil
}

func ComputeAetraCoreParamsHash(params AetraCoreParams) string {
	return hashParts(
		"aetra-aek-params-v1",
		fmt.Sprint(params.Enabled),
		params.Authority,
		fmt.Sprint(params.DefaultQueryLimit),
		fmt.Sprint(params.MaxQueryLimit),
		fmt.Sprint(params.MaxZones),
		fmt.Sprint(params.MaxShardsPerZone),
		fmt.Sprint(params.MaxProposalItemsPerBlock),
		fmt.Sprint(params.RootHistoryWindow),
		fmt.Sprint(params.CrossZoneFinalityDelay),
		fmt.Sprint(params.DeterministicProposalGrouping),
		params.ProductionVersionGate,
	)
}

func ComputeKernelBlockPlanHash(plan KernelBlockPlan) string {
	return hashParts(
		"aetra-aek-kernel-block-plan-v1",
		plan.ChainID,
		fmt.Sprint(plan.Height),
		string(plan.Phase),
		plan.ProposalRoot,
		plan.RoutingRoot,
		plan.PreviousGlobalRoot,
		plan.ParamsHash,
	)
}

func ComputeKernelFinalizationHash(finalization KernelFinalization) string {
	return hashParts(
		"aetra-aek-kernel-finalization-v1",
		finalization.ChainID,
		fmt.Sprint(finalization.Height),
		finalization.PlanHash,
		finalization.Header.HeaderHash,
		finalization.GlobalRoot.GlobalRoot,
		finalization.RootSnapshot.Finality.GlobalMessageRoot,
		finalization.ReceiptsRoot,
		fmt.Sprint(finalization.ReceiptCount),
		fmt.Sprint(finalization.CommitmentCount),
	)
}

func ComputeKernelBlockHeaderHash(header KernelBlockHeaderCommitment) string {
	return hashParts(
		"aetra-aek-block-header-v1",
		fmt.Sprint(header.Height),
		fmt.Sprint(header.TimeUnix),
		header.PreviousAppHash,
		header.ZonesRoot,
		header.MessagesRoot,
		header.ReceiptsRoot,
	)
}

func BuildKernelExportManifest(state CoreState, height uint64, appHash string) (ExportManifest, error) {
	if err := state.Validate(); err != nil {
		return ExportManifest{}, err
	}
	root, found := state.GlobalRootByHeight(height)
	if !found {
		return ExportManifest{}, fmt.Errorf("aetracore kernel export missing global root at height %d", height)
	}
	return NewExportManifest(root, appHash, state)
}

func ValidateKernelImport(state CoreState, manifest ExportManifest) error {
	if err := state.Validate(); err != nil {
		return err
	}
	if err := manifest.ValidateHash(); err != nil {
		return err
	}
	root, found := state.GlobalRootByHeight(manifest.Height)
	if !found {
		return fmt.Errorf("aetracore kernel import missing global root at height %d", manifest.Height)
	}
	if manifest.GlobalRoot != root.GlobalRoot {
		return errors.New("aetracore kernel import global root mismatch")
	}
	if err := ValidateExportImportRootChecks(state, manifest); err != nil {
		return err
	}
	if manifest.ZoneCommitmentCount != uint64(len(state.CommitmentsAtHeight(manifest.Height))) {
		return errors.New("aetracore kernel import zone commitment count mismatch")
	}
	if manifest.ServiceDescriptorCount != uint64(len(state.ServiceDescriptors)) {
		return errors.New("aetracore kernel import service descriptor count mismatch")
	}
	return nil
}

func sortExecutionReceipts(receipts []ExecutionReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return compareExecutionReceipts(receipts[i], receipts[j]) < 0
	})
}

func compareExecutionReceipts(left, right ExecutionReceipt) int {
	if left.Height < right.Height {
		return -1
	}
	if left.Height > right.Height {
		return 1
	}
	if left.Sequence < right.Sequence {
		return -1
	}
	if left.Sequence > right.Sequence {
		return 1
	}
	if left.SourceZone < right.SourceZone {
		return -1
	}
	if left.SourceZone > right.SourceZone {
		return 1
	}
	if left.DestinationZone < right.DestinationZone {
		return -1
	}
	if left.DestinationZone > right.DestinationZone {
		return 1
	}
	if left.TxHash < right.TxHash {
		return -1
	}
	if left.TxHash > right.TxHash {
		return 1
	}
	if left.ReceiptHash < right.ReceiptHash {
		return -1
	}
	if left.ReceiptHash > right.ReceiptHash {
		return 1
	}
	return 0
}

func latestGlobalRootBefore(roots []GlobalStateRoot, height uint64) (GlobalStateRoot, bool) {
	var latest GlobalStateRoot
	found := false
	for _, root := range roots {
		if root.Height >= height {
			continue
		}
		if !found || root.Height > latest.Height {
			latest = root
			found = true
		}
	}
	return latest, found
}
