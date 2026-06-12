package types

import (
	"errors"
	"fmt"
	"sort"
)

type KernelMessageKind string

const (
	KernelMessageLocalTx		KernelMessageKind	= "LOCAL_TX"
	KernelMessageRoutedInbound	KernelMessageKind	= "ROUTED_INBOUND"
)

type KernelGasLimits struct {
	MaxBlockGas	uint64
	MaxZoneGas	uint64
}

type KernelMessageEnvelope struct {
	Kind			KernelMessageKind
	TxHash			string
	SourceZone		ZoneID
	SourceShard		ShardID
	DestinationZone		ZoneID
	DestinationShard	ShardID
	Sender			string
	Nonce			uint64
	GasLimit		uint64
	PriorityClass		uint32
	AdmissionHeight		uint64
	TxIndex			uint32
	MessageIndex		uint32
	CommittedHeight		uint64
	EligibleHeight		uint64
}

type KernelZoneWorkload struct {
	ZoneID		ZoneID
	ShardID		ShardID
	GasLimit	uint64
	Items		[]ProposalItem
}

type KernelABCIProposal struct {
	Plan			KernelBlockPlan
	Workloads		[]KernelZoneWorkload
	RoutedMessageRoot	string
	EnvelopeRoot		string
	BlockGas		uint64
	ZoneGas			[]KernelZoneGas
	ProposalHash		string
}

type KernelZoneGas struct {
	ZoneID		ZoneID
	GasLimit	uint64
}

type KernelCleanupItem struct {
	QueueID		string
	ItemID		string
	HeightDue	uint64
	DeleteRoot	string
}

type KernelCleanupResult struct {
	Height		uint64
	Processed	[]KernelCleanupItem
	CleanupRoot	string
}

type KernelABCICommitRecord struct {
	Height		uint64
	AppHash		string
	HeaderHash	string
	GlobalRoot	string
	MessageRoot	string
	ReceiptsRoot	string
	ProofRootCount	uint64
	CommitmentCount	uint64
	CommitHash	string
}

func PrepareKernelABCIProposal(ctx KernelConsensusContext, state CoreState, localTxs, routedMessages []KernelMessageEnvelope, limits KernelGasLimits) (KernelABCIProposal, error) {
	if err := validateKernelGasLimits(limits); err != nil {
		return KernelABCIProposal{}, err
	}
	envelopes := append([]KernelMessageEnvelope(nil), localTxs...)
	envelopes = append(envelopes, routedMessages...)
	envelopes = normalizeKernelEnvelopes(envelopes)
	if err := validateKernelEnvelopes(ctx, state, envelopes, limits); err != nil {
		return KernelABCIProposal{}, err
	}
	items := make([]ProposalItem, len(envelopes))
	for i, envelope := range envelopes {
		items[i] = envelope.ProposalItem()
	}
	plan, err := PrepareKernelProposal(ctx, state, items)
	if err != nil {
		return KernelABCIProposal{}, err
	}
	proposal := KernelABCIProposal{
		Plan:			plan,
		Workloads:		buildKernelWorkloads(envelopes),
		RoutedMessageRoot:	ComputeKernelRoutedMessageRoot(routedMessages),
		EnvelopeRoot:		ComputeKernelEnvelopeRoot(envelopes),
		BlockGas:		sumKernelGas(envelopes),
		ZoneGas:		buildKernelZoneGas(envelopes),
	}
	proposal.ProposalHash = ComputeKernelABCIProposalHash(proposal)
	return proposal, ProcessKernelABCIProposal(ctx, state, proposal, envelopes, limits)
}

func ProcessKernelABCIProposal(ctx KernelConsensusContext, state CoreState, proposal KernelABCIProposal, envelopes []KernelMessageEnvelope, limits KernelGasLimits) error {
	if err := validateKernelGasLimits(limits); err != nil {
		return err
	}
	envelopes = normalizeKernelEnvelopes(envelopes)
	if err := validateKernelEnvelopes(ctx, state, envelopes, limits); err != nil {
		return err
	}
	if err := ProcessKernelProposal(ctx, state, proposal.Plan); err != nil {
		return err
	}
	if expected := ComputeKernelEnvelopeRoot(envelopes); expected != proposal.EnvelopeRoot {
		return errors.New("aetracore ABCI proposal envelope root mismatch")
	}
	routed := make([]KernelMessageEnvelope, 0)
	for _, envelope := range envelopes {
		if envelope.Kind == KernelMessageRoutedInbound {
			routed = append(routed, envelope)
		}
	}
	if expected := ComputeKernelRoutedMessageRoot(routed); expected != proposal.RoutedMessageRoot {
		return errors.New("aetracore ABCI proposal routed message root mismatch")
	}
	workloads := buildKernelWorkloads(envelopes)
	if !sameKernelWorkloads(workloads, proposal.Workloads) {
		return errors.New("aetracore ABCI proposal workload grouping mismatch")
	}
	if proposal.BlockGas != sumKernelGas(envelopes) {
		return errors.New("aetracore ABCI proposal block gas mismatch")
	}
	if !sameKernelZoneGas(buildKernelZoneGas(envelopes), proposal.ZoneGas) {
		return errors.New("aetracore ABCI proposal zone gas mismatch")
	}
	if expected := ComputeKernelABCIProposalHash(proposal); expected != proposal.ProposalHash {
		return fmt.Errorf("aetracore ABCI proposal hash mismatch: expected %s", expected)
	}
	return nil
}

func ProcessKernelABCIProposalWithTimestampBounds(ctx KernelConsensusContext, state CoreState, proposal KernelABCIProposal, envelopes []KernelMessageEnvelope, limits KernelGasLimits, bounds KernelTimestampBounds) error {
	if err := ValidateKernelTimestampBounds(ctx, bounds); err != nil {
		return err
	}
	return ProcessKernelABCIProposal(ctx, state, proposal, envelopes, limits)
}

func FinalizeKernelABCIBlock(ctx KernelConsensusContext, state CoreState, proposal KernelABCIProposal, envelopes []KernelMessageEnvelope, input KernelFinalizationInput, cleanupQueue []KernelCleanupItem, cleanupLimit uint64) (CoreState, KernelFinalization, KernelCleanupResult, error) {
	if cleanupLimit == 0 {
		return CoreState{}, KernelFinalization{}, KernelCleanupResult{}, errors.New("aetracore ABCI cleanup limit must be positive")
	}
	if err := ProcessKernelABCIProposal(ctx, state, proposal, envelopes, KernelGasLimits{MaxBlockGas: proposal.BlockGas, MaxZoneGas: maxKernelZoneGas(proposal.ZoneGas)}); err != nil {
		return CoreState{}, KernelFinalization{}, KernelCleanupResult{}, err
	}
	next, finalization, err := FinalizeKernelBlock(ctx, state, proposal.Plan, input)
	if err != nil {
		return CoreState{}, KernelFinalization{}, KernelCleanupResult{}, err
	}
	cleanup, err := ProcessKernelCleanupQueue(ctx.Height, cleanupQueue, cleanupLimit)
	if err != nil {
		return CoreState{}, KernelFinalization{}, KernelCleanupResult{}, err
	}
	return next, finalization, cleanup, nil
}

func CommitKernelABCIBlock(finalization KernelFinalization, appHash string) (KernelABCICommitRecord, error) {
	if err := finalization.Validate(); err != nil {
		return KernelABCICommitRecord{}, err
	}
	if err := ValidateHash("aetracore ABCI app hash", appHash); err != nil {
		return KernelABCICommitRecord{}, err
	}
	record := KernelABCICommitRecord{
		Height:			finalization.Height,
		AppHash:		appHash,
		HeaderHash:		finalization.Header.HeaderHash,
		GlobalRoot:		finalization.GlobalRoot.GlobalRoot,
		MessageRoot:		finalization.RootSnapshot.Finality.GlobalMessageRoot,
		ReceiptsRoot:		finalization.RootSnapshot.Finality.ExecutionReceiptRoot,
		ProofRootCount:		uint64(len(finalization.RootSnapshot.ProofRoots)),
		CommitmentCount:	finalization.CommitmentCount,
	}
	record.CommitHash = ComputeKernelABCICommitHash(record)
	return record, record.Validate()
}

func ProcessKernelCleanupQueue(height uint64, queue []KernelCleanupItem, limit uint64) (KernelCleanupResult, error) {
	if height == 0 {
		return KernelCleanupResult{}, errors.New("aetracore cleanup height must be positive")
	}
	if limit == 0 {
		return KernelCleanupResult{}, errors.New("aetracore cleanup limit must be positive")
	}
	ordered := normalizeKernelCleanupItems(queue)
	processed := make([]KernelCleanupItem, 0)
	for _, item := range ordered {
		if err := item.Validate(); err != nil {
			return KernelCleanupResult{}, err
		}
		if item.HeightDue > height {
			continue
		}
		processed = append(processed, item)
		if uint64(len(processed)) == limit {
			break
		}
	}
	result := KernelCleanupResult{Height: height, Processed: processed}
	result.CleanupRoot = ComputeKernelCleanupRoot(result)
	return result, nil
}

func (e KernelMessageEnvelope) ProposalItem() ProposalItem {
	admission := e.AdmissionHeight
	if e.Kind == KernelMessageRoutedInbound && e.EligibleHeight != 0 {
		admission = e.EligibleHeight
	}
	return ProposalItem{
		ZoneID:			e.DestinationZone,
		ShardID:		e.DestinationShard,
		TxHash:			e.TxHash,
		PriorityClass:		e.PriorityClass,
		AdmissionHeight:	admission,
		TxIndex:		e.TxIndex,
		MessageIndex:		e.MessageIndex,
	}
}

func (e KernelMessageEnvelope) ValidateBasic(ctx KernelConsensusContext, state CoreState) error {
	if err := ctx.Validate(); err != nil {
		return err
	}
	if e.Kind != KernelMessageLocalTx && e.Kind != KernelMessageRoutedInbound {
		return fmt.Errorf("unknown aetracore kernel message kind %q", e.Kind)
	}
	if err := ValidateHash("aetracore kernel envelope tx hash", e.TxHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore kernel envelope sender", e.Sender); err != nil {
		return err
	}
	if e.Nonce == 0 {
		return errors.New("aetracore kernel envelope nonce must be positive")
	}
	if e.GasLimit == 0 {
		return errors.New("aetracore kernel envelope gas limit must be positive")
	}
	if e.AdmissionHeight == 0 || e.AdmissionHeight > ctx.Height {
		return errors.New("aetracore kernel envelope admission height is invalid")
	}
	if err := validateZoneAndShard(state, e.SourceZone, e.SourceShard, ctx.Height); err != nil {
		return err
	}
	if err := validateZoneAndShard(state, e.DestinationZone, e.DestinationShard, ctx.Height); err != nil {
		return err
	}
	if e.Kind == KernelMessageLocalTx {
		if e.SourceZone != e.DestinationZone || e.SourceShard != e.DestinationShard {
			return errors.New("aetracore local envelope must stay within one zone shard")
		}
		return nil
	}
	if e.SourceZone == e.DestinationZone {
		return errors.New("aetracore routed envelope requires cross-zone source")
	}
	if e.CommittedHeight == 0 || e.EligibleHeight == 0 {
		return errors.New("aetracore routed envelope requires committed and eligible heights")
	}
	if e.EligibleHeight > ctx.Height {
		return errors.New("aetracore routed envelope is not yet eligible")
	}
	if e.CommittedHeight >= e.EligibleHeight {
		return errors.New("aetracore routed envelope eligible height must follow committed height")
	}
	if _, found := state.RootSnapshotAtHeight(e.CommittedHeight); !found {
		return errors.New("aetracore routed envelope missing committed root")
	}
	if _, found := state.ZoneCommitmentAtHeight(e.CommittedHeight, e.SourceZone); !found {
		return errors.New("aetracore routed envelope missing source zone commitment")
	}
	return nil
}

func (i KernelCleanupItem) Validate() error {
	if err := validatePolicyID("aetracore cleanup queue id", i.QueueID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore cleanup item id", i.ItemID); err != nil {
		return err
	}
	if i.HeightDue == 0 {
		return errors.New("aetracore cleanup item height must be positive")
	}
	return ValidateHash("aetracore cleanup delete root", i.DeleteRoot)
}

func (r KernelABCICommitRecord) Validate() error {
	if r.Height == 0 {
		return errors.New("aetracore ABCI commit height must be positive")
	}
	if err := ValidateHash("aetracore ABCI commit app hash", r.AppHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore ABCI commit header hash", r.HeaderHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore ABCI commit global root", r.GlobalRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore ABCI commit message root", r.MessageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore ABCI commit receipts root", r.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore ABCI commit hash", r.CommitHash); err != nil {
		return err
	}
	if expected := ComputeKernelABCICommitHash(r); expected != r.CommitHash {
		return fmt.Errorf("aetracore ABCI commit hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeKernelEnvelopeHash(envelope KernelMessageEnvelope) string {
	envelope = normalizeKernelEnvelope(envelope)
	return hashParts(
		"aetra-aek-kernel-envelope-v1",
		string(envelope.Kind),
		envelope.TxHash,
		string(envelope.SourceZone),
		string(envelope.SourceShard),
		string(envelope.DestinationZone),
		string(envelope.DestinationShard),
		envelope.Sender,
		fmt.Sprint(envelope.Nonce),
		fmt.Sprint(envelope.GasLimit),
		fmt.Sprint(envelope.PriorityClass),
		fmt.Sprint(envelope.AdmissionHeight),
		fmt.Sprint(envelope.TxIndex),
		fmt.Sprint(envelope.MessageIndex),
		fmt.Sprint(envelope.CommittedHeight),
		fmt.Sprint(envelope.EligibleHeight),
	)
}

func ComputeKernelEnvelopeRoot(envelopes []KernelMessageEnvelope) string {
	ordered := normalizeKernelEnvelopes(envelopes)
	parts := []string{"aetra-aek-kernel-envelope-root-v1", fmt.Sprint(len(ordered))}
	for _, envelope := range ordered {
		parts = append(parts, ComputeKernelEnvelopeHash(envelope))
	}
	return hashParts(parts...)
}

func ComputeKernelRoutedMessageRoot(envelopes []KernelMessageEnvelope) string {
	routed := make([]KernelMessageEnvelope, 0, len(envelopes))
	for _, envelope := range envelopes {
		envelope = normalizeKernelEnvelope(envelope)
		if envelope.Kind == KernelMessageRoutedInbound {
			routed = append(routed, envelope)
		}
	}
	sortKernelEnvelopes(routed)
	parts := []string{"aetra-aek-kernel-routed-message-root-v1", fmt.Sprint(len(routed))}
	for _, envelope := range routed {
		parts = append(parts, ComputeKernelEnvelopeHash(envelope))
	}
	return hashParts(parts...)
}

func ComputeKernelABCIProposalHash(proposal KernelABCIProposal) string {
	workloadsRoot := ComputeKernelWorkloadsRoot(proposal.Workloads)
	zoneGasRoot := ComputeKernelZoneGasRoot(proposal.ZoneGas)
	return hashParts(
		"aetra-aek-abci-proposal-v1",
		proposal.Plan.PlanHash,
		proposal.EnvelopeRoot,
		proposal.RoutedMessageRoot,
		workloadsRoot,
		fmt.Sprint(proposal.BlockGas),
		zoneGasRoot,
	)
}

func ComputeKernelWorkloadsRoot(workloads []KernelZoneWorkload) string {
	ordered := cloneKernelWorkloads(workloads)
	sortKernelWorkloads(ordered)
	parts := []string{"aetra-aek-kernel-workloads-root-v1", fmt.Sprint(len(ordered))}
	for _, workload := range ordered {
		parts = append(parts, string(workload.ZoneID), string(workload.ShardID), fmt.Sprint(workload.GasLimit), fmt.Sprint(len(workload.Items)))
		items := append([]ProposalItem(nil), workload.Items...)
		sortProposalItems(items)
		for _, item := range items {
			parts = append(parts, ComputeProposalItemHash(item))
		}
	}
	return hashParts(parts...)
}

func ComputeKernelZoneGasRoot(zoneGas []KernelZoneGas) string {
	ordered := append([]KernelZoneGas(nil), zoneGas...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ZoneID < ordered[j].ZoneID })
	parts := []string{"aetra-aek-kernel-zone-gas-root-v1", fmt.Sprint(len(ordered))}
	for _, gas := range ordered {
		parts = append(parts, string(gas.ZoneID), fmt.Sprint(gas.GasLimit))
	}
	return hashParts(parts...)
}

func ComputeKernelCleanupRoot(result KernelCleanupResult) string {
	items := normalizeKernelCleanupItems(result.Processed)
	parts := []string{"aetra-aek-kernel-cleanup-root-v1", fmt.Sprint(result.Height), fmt.Sprint(len(items))}
	for _, item := range items {
		parts = append(parts, item.QueueID, item.ItemID, fmt.Sprint(item.HeightDue), item.DeleteRoot)
	}
	return hashParts(parts...)
}

func ComputeKernelABCICommitHash(record KernelABCICommitRecord) string {
	return hashParts(
		"aetra-aek-abci-commit-v1",
		fmt.Sprint(record.Height),
		record.AppHash,
		record.HeaderHash,
		record.GlobalRoot,
		record.MessageRoot,
		record.ReceiptsRoot,
		fmt.Sprint(record.ProofRootCount),
		fmt.Sprint(record.CommitmentCount),
	)
}

func validateKernelEnvelopes(ctx KernelConsensusContext, state CoreState, envelopes []KernelMessageEnvelope, limits KernelGasLimits) error {
	var blockGas uint64
	zoneGas := make(map[ZoneID]uint64)
	seen := make(map[string]struct{}, len(envelopes))
	for _, envelope := range normalizeKernelEnvelopes(envelopes) {
		if err := envelope.ValidateBasic(ctx, state); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%020d", envelope.Sender, envelope.Nonce)
		if _, found := seen[key]; found {
			return errors.New("duplicate aetracore kernel envelope nonce")
		}
		seen[key] = struct{}{}
		if limits.MaxBlockGas-blockGas < envelope.GasLimit {
			return errors.New("aetracore kernel proposal exceeds block gas")
		}
		blockGas += envelope.GasLimit
		currentZoneGas := zoneGas[envelope.DestinationZone]
		if limits.MaxZoneGas-currentZoneGas < envelope.GasLimit {
			return errors.New("aetracore kernel proposal exceeds zone gas")
		}
		zoneGas[envelope.DestinationZone] = currentZoneGas + envelope.GasLimit
	}
	return nil
}

func validateKernelGasLimits(limits KernelGasLimits) error {
	if limits.MaxBlockGas == 0 {
		return errors.New("aetracore kernel max block gas must be positive")
	}
	if limits.MaxZoneGas == 0 {
		return errors.New("aetracore kernel max zone gas must be positive")
	}
	if limits.MaxZoneGas > limits.MaxBlockGas {
		return errors.New("aetracore kernel max zone gas must not exceed block gas")
	}
	return nil
}

func normalizeKernelEnvelope(envelope KernelMessageEnvelope) KernelMessageEnvelope {
	if envelope.AdmissionHeight == 0 {
		envelope.AdmissionHeight = envelope.EligibleHeight
	}
	return envelope
}

func normalizeKernelEnvelopes(envelopes []KernelMessageEnvelope) []KernelMessageEnvelope {
	out := make([]KernelMessageEnvelope, len(envelopes))
	for i, envelope := range envelopes {
		out[i] = normalizeKernelEnvelope(envelope)
	}
	sortKernelEnvelopes(out)
	return out
}

func sortKernelEnvelopes(envelopes []KernelMessageEnvelope) {
	sort.SliceStable(envelopes, func(i, j int) bool {
		return compareKernelEnvelopes(envelopes[i], envelopes[j]) < 0
	})
}

func compareKernelEnvelopes(left, right KernelMessageEnvelope) int {
	if left.Kind != right.Kind {
		if left.Kind == KernelMessageRoutedInbound {
			return -1
		}
		return 1
	}
	if left.Kind == KernelMessageRoutedInbound {
		if left.EligibleHeight != right.EligibleHeight {
			if left.EligibleHeight < right.EligibleHeight {
				return -1
			}
			return 1
		}
		if left.CommittedHeight != right.CommittedHeight {
			if left.CommittedHeight < right.CommittedHeight {
				return -1
			}
			return 1
		}
	}
	if left.DestinationZone != right.DestinationZone {
		if left.DestinationZone < right.DestinationZone {
			return -1
		}
		return 1
	}
	if left.DestinationShard != right.DestinationShard {
		if left.DestinationShard < right.DestinationShard {
			return -1
		}
		return 1
	}
	if left.PriorityClass != right.PriorityClass {
		if left.PriorityClass < right.PriorityClass {
			return -1
		}
		return 1
	}
	if left.Sender != right.Sender {
		if left.Sender < right.Sender {
			return -1
		}
		return 1
	}
	if left.Nonce != right.Nonce {
		if left.Nonce < right.Nonce {
			return -1
		}
		return 1
	}
	if left.TxHash < right.TxHash {
		return -1
	}
	if left.TxHash > right.TxHash {
		return 1
	}
	return 0
}

func buildKernelWorkloads(envelopes []KernelMessageEnvelope) []KernelZoneWorkload {
	envelopes = normalizeKernelEnvelopes(envelopes)
	byRoute := make(map[string]int)
	workloads := make([]KernelZoneWorkload, 0)
	for _, envelope := range envelopes {
		item := envelope.ProposalItem()
		key := fmt.Sprintf("%s/%s", item.ZoneID, item.ShardID)
		idx, found := byRoute[key]
		if !found {
			idx = len(workloads)
			byRoute[key] = idx
			workloads = append(workloads, KernelZoneWorkload{ZoneID: item.ZoneID, ShardID: item.ShardID})
		}
		workloads[idx].GasLimit += envelope.GasLimit
		workloads[idx].Items = append(workloads[idx].Items, item)
	}
	sortKernelWorkloads(workloads)
	return workloads
}

func buildKernelZoneGas(envelopes []KernelMessageEnvelope) []KernelZoneGas {
	byZone := make(map[ZoneID]uint64)
	for _, envelope := range normalizeKernelEnvelopes(envelopes) {
		byZone[envelope.DestinationZone] += envelope.GasLimit
	}
	zones := make([]ZoneID, 0, len(byZone))
	for zoneID := range byZone {
		zones = append(zones, zoneID)
	}
	sort.SliceStable(zones, func(i, j int) bool { return zones[i] < zones[j] })
	out := make([]KernelZoneGas, len(zones))
	for i, zoneID := range zones {
		out[i] = KernelZoneGas{ZoneID: zoneID, GasLimit: byZone[zoneID]}
	}
	return out
}

func sumKernelGas(envelopes []KernelMessageEnvelope) uint64 {
	var total uint64
	for _, envelope := range envelopes {
		total += envelope.GasLimit
	}
	return total
}

func maxKernelZoneGas(zoneGas []KernelZoneGas) uint64 {
	var max uint64
	for _, gas := range zoneGas {
		if gas.GasLimit > max {
			max = gas.GasLimit
		}
	}
	if max == 0 {
		return 1
	}
	return max
}

func sortKernelWorkloads(workloads []KernelZoneWorkload) {
	sort.SliceStable(workloads, func(i, j int) bool {
		if workloads[i].ZoneID != workloads[j].ZoneID {
			return workloads[i].ZoneID < workloads[j].ZoneID
		}
		return workloads[i].ShardID < workloads[j].ShardID
	})
	for i := range workloads {
		sortProposalItems(workloads[i].Items)
	}
}

func cloneKernelWorkloads(workloads []KernelZoneWorkload) []KernelZoneWorkload {
	out := make([]KernelZoneWorkload, len(workloads))
	for i, workload := range workloads {
		out[i] = workload
		out[i].Items = append([]ProposalItem(nil), workload.Items...)
	}
	return out
}

func sameKernelWorkloads(left, right []KernelZoneWorkload) bool {
	left = cloneKernelWorkloads(left)
	right = cloneKernelWorkloads(right)
	sortKernelWorkloads(left)
	sortKernelWorkloads(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].ZoneID != right[i].ZoneID || left[i].ShardID != right[i].ShardID || left[i].GasLimit != right[i].GasLimit {
			return false
		}
		if len(left[i].Items) != len(right[i].Items) {
			return false
		}
		for j := range left[i].Items {
			if left[i].Items[j] != right[i].Items[j] {
				return false
			}
		}
	}
	return true
}

func sameKernelZoneGas(left, right []KernelZoneGas) bool {
	left = append([]KernelZoneGas(nil), left...)
	right = append([]KernelZoneGas(nil), right...)
	sort.SliceStable(left, func(i, j int) bool { return left[i].ZoneID < left[j].ZoneID })
	sort.SliceStable(right, func(i, j int) bool { return right[i].ZoneID < right[j].ZoneID })
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func normalizeKernelCleanupItems(items []KernelCleanupItem) []KernelCleanupItem {
	out := append([]KernelCleanupItem(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].HeightDue != out[j].HeightDue {
			return out[i].HeightDue < out[j].HeightDue
		}
		if out[i].QueueID != out[j].QueueID {
			return out[i].QueueID < out[j].QueueID
		}
		return out[i].ItemID < out[j].ItemID
	})
	return out
}
