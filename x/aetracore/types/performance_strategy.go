package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type BlockSTMAccessMode string

const (
	BlockSTMAccessRead	BlockSTMAccessMode	= "read"
	BlockSTMAccessWrite	BlockSTMAccessMode	= "write"
)

type BlockSTMStateAccess struct {
	ActorZoneID	ZoneID
	ActorShardID	ShardID
	StateZoneID	ZoneID
	StateShardID	ShardID
	StateKey	string
	Mode		BlockSTMAccessMode
	ViaMessage	bool
}

type BlockSTMMessageBatch struct {
	SourceZoneID		ZoneID
	SourceShardID		ShardID
	DestinationZoneID	ZoneID
	DestinationShardID	ShardID
	MessageCount		uint32
	BatchHash		string
}

type BlockSTMConflictSet struct {
	ConflictKey	string
	WorkloadIDs	[]string
}

type BlockSTMZoneWorkload struct {
	WorkloadID		string
	ZoneID			ZoneID
	ShardID			ShardID
	Items			[]ProposalItem
	StateAccesses		[]BlockSTMStateAccess
	MessageBatches		[]BlockSTMMessageBatch
	ConflictKeyRoot		string
	MessageBatchRoot	string
}

type BlockSTMZonePerformancePlan struct {
	Height			uint64
	Workloads		[]BlockSTMZoneWorkload
	ConflictSets		[]BlockSTMConflictSet
	ParallelWorkloads	uint32
	GlobalWriteLocks	uint32
	CrossZoneWrites		uint32
	PlanHash		string
}

func BuildBlockSTMZonePerformancePlan(schedule ProposalSchedule, accesses []BlockSTMStateAccess, batches []BlockSTMMessageBatch) (BlockSTMZonePerformancePlan, error) {
	if err := schedule.Validate(); err != nil {
		return BlockSTMZonePerformancePlan{}, err
	}
	accesses = normalizeBlockSTMStateAccesses(accesses)
	batches = normalizeBlockSTMMessageBatches(batches)
	for _, access := range accesses {
		if err := access.Validate(); err != nil {
			return BlockSTMZonePerformancePlan{}, err
		}
		if access.Mode == BlockSTMAccessWrite && access.StateZoneID != access.ActorZoneID && !access.ViaMessage {
			return BlockSTMZonePerformancePlan{}, errors.New("aetracore BlockSTM cross-zone writes must be converted into output messages")
		}
		if access.IsGlobalWriteLock() {
			return BlockSTMZonePerformancePlan{}, errors.New("aetracore BlockSTM workload cannot take global write lock")
		}
	}
	for _, batch := range batches {
		if err := batch.Validate(); err != nil {
			return BlockSTMZonePerformancePlan{}, err
		}
	}

	workloads := make([]BlockSTMZoneWorkload, 0, len(schedule.Groups))
	for _, group := range schedule.Groups {
		workload := BlockSTMZoneWorkload{
			ZoneID:		group.ZoneID,
			ShardID:	group.ShardID,
			Items:		append([]ProposalItem(nil), group.Items...),
		}
		sortProposalItems(workload.Items)
		for _, access := range accesses {
			if access.ActorZoneID == group.ZoneID && access.ActorShardID == group.ShardID {
				workload.StateAccesses = append(workload.StateAccesses, access)
			}
		}
		for _, batch := range batches {
			if batch.SourceZoneID == group.ZoneID && batch.SourceShardID == group.ShardID {
				workload.MessageBatches = append(workload.MessageBatches, batch)
			}
		}
		workload = workload.Normalize()
		workload.WorkloadID = ComputeBlockSTMWorkloadID(workload)
		workload.ConflictKeyRoot = ComputeBlockSTMConflictKeyRoot(workload.StateAccesses)
		workload.MessageBatchRoot = ComputeBlockSTMMessageBatchRoot(workload.MessageBatches)
		workloads = append(workloads, workload)
	}
	plan := BlockSTMZonePerformancePlan{
		Height:			schedule.Height,
		Workloads:		normalizeBlockSTMZoneWorkloads(workloads),
		ParallelWorkloads:	uint32(len(workloads)),
	}
	plan.ConflictSets = computeBlockSTMConflictSets(plan.Workloads)
	plan.CrossZoneWrites = countBlockSTMCrossZoneMessageWrites(accesses)
	plan.PlanHash = ComputeBlockSTMZonePerformancePlanHash(plan)
	return plan, plan.Validate()
}

func (p BlockSTMZonePerformancePlan) Validate() error {
	p = p.Normalize()
	if p.Height == 0 {
		return errors.New("aetracore BlockSTM performance height must be positive")
	}
	if len(p.Workloads) == 0 {
		return errors.New("aetracore BlockSTM performance plan requires workloads")
	}
	if p.GlobalWriteLocks != 0 {
		return errors.New("aetracore BlockSTM performance plan contains global write locks")
	}
	if p.ParallelWorkloads != uint32(len(p.Workloads)) {
		return errors.New("aetracore BlockSTM parallel workload count mismatch")
	}
	var previous string
	seen := make(map[string]struct{}, len(p.Workloads))
	for i, workload := range p.Workloads {
		if err := workload.Validate(); err != nil {
			return err
		}
		if _, found := seen[workload.WorkloadID]; found {
			return fmt.Errorf("duplicate aetracore BlockSTM workload %s", workload.WorkloadID)
		}
		seen[workload.WorkloadID] = struct{}{}
		if i > 0 && previous >= workload.WorkloadID {
			return errors.New("aetracore BlockSTM workloads must be sorted canonically")
		}
		previous = workload.WorkloadID
	}
	for _, conflict := range p.ConflictSets {
		if err := conflict.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore BlockSTM performance plan hash", p.PlanHash); err != nil {
		return err
	}
	if expected := ComputeBlockSTMZonePerformancePlanHash(p); p.PlanHash != expected {
		return fmt.Errorf("aetracore BlockSTM performance plan hash mismatch: expected %s", expected)
	}
	return nil
}

func (p BlockSTMZonePerformancePlan) Normalize() BlockSTMZonePerformancePlan {
	p.Workloads = normalizeBlockSTMZoneWorkloads(p.Workloads)
	p.ConflictSets = normalizeBlockSTMConflictSets(p.ConflictSets)
	p.PlanHash = strings.ToLower(strings.TrimSpace(p.PlanHash))
	return p
}

func (w BlockSTMZoneWorkload) Validate() error {
	w = w.Normalize()
	if err := ValidateHash("aetracore BlockSTM workload id", w.WorkloadID); err != nil {
		return err
	}
	if err := ValidateZoneID(w.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(w.ShardID); err != nil {
		return err
	}
	if len(w.Items) == 0 {
		return errors.New("aetracore BlockSTM workload requires proposal items")
	}
	for _, item := range w.Items {
		if err := item.Validate(); err != nil {
			return err
		}
		if item.ZoneID != w.ZoneID || item.ShardID != w.ShardID {
			return errors.New("aetracore BlockSTM workload proposal item route mismatch")
		}
	}
	for _, access := range w.StateAccesses {
		if err := access.Validate(); err != nil {
			return err
		}
		if access.ActorZoneID != w.ZoneID || access.ActorShardID != w.ShardID {
			return errors.New("aetracore BlockSTM workload access route mismatch")
		}
	}
	for _, batch := range w.MessageBatches {
		if err := batch.Validate(); err != nil {
			return err
		}
		if batch.SourceZoneID != w.ZoneID || batch.SourceShardID != w.ShardID {
			return errors.New("aetracore BlockSTM workload message batch route mismatch")
		}
	}
	if err := ValidateHash("aetracore BlockSTM workload conflict key root", w.ConflictKeyRoot); err != nil {
		return err
	}
	if w.ConflictKeyRoot != ComputeBlockSTMConflictKeyRoot(w.StateAccesses) {
		return errors.New("aetracore BlockSTM workload conflict key root mismatch")
	}
	if err := ValidateHash("aetracore BlockSTM workload message batch root", w.MessageBatchRoot); err != nil {
		return err
	}
	if w.MessageBatchRoot != ComputeBlockSTMMessageBatchRoot(w.MessageBatches) {
		return errors.New("aetracore BlockSTM workload message batch root mismatch")
	}
	if w.WorkloadID != ComputeBlockSTMWorkloadID(w) {
		return errors.New("aetracore BlockSTM workload id mismatch")
	}
	return nil
}

func (w BlockSTMZoneWorkload) Normalize() BlockSTMZoneWorkload {
	sortProposalItems(w.Items)
	w.StateAccesses = normalizeBlockSTMStateAccesses(w.StateAccesses)
	w.MessageBatches = normalizeBlockSTMMessageBatches(w.MessageBatches)
	w.WorkloadID = strings.ToLower(strings.TrimSpace(w.WorkloadID))
	w.ConflictKeyRoot = strings.ToLower(strings.TrimSpace(w.ConflictKeyRoot))
	w.MessageBatchRoot = strings.ToLower(strings.TrimSpace(w.MessageBatchRoot))
	return w
}

func (a BlockSTMStateAccess) Validate() error {
	if err := ValidateZoneID(a.ActorZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(a.ActorShardID); err != nil {
		return err
	}
	if err := ValidateZoneID(a.StateZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(a.StateShardID); err != nil {
		return err
	}
	if err := validateToken("aetracore BlockSTM state key", a.StateKey, MaxScopeLength); err != nil {
		return err
	}
	if !IsBlockSTMAccessMode(a.Mode) {
		return fmt.Errorf("unknown aetracore BlockSTM access mode %q", a.Mode)
	}
	return nil
}

func (a BlockSTMStateAccess) Normalize() BlockSTMStateAccess {
	a.StateKey = strings.TrimSpace(a.StateKey)
	return a
}

func (a BlockSTMStateAccess) ConflictKey() string {
	a = a.Normalize()
	return string(a.StateZoneID) + "/" + string(a.StateShardID) + "/" + a.StateKey
}

func (a BlockSTMStateAccess) IsGlobalWriteLock() bool {
	a = a.Normalize()
	if a.Mode != BlockSTMAccessWrite {
		return false
	}
	return a.StateZoneID == ZoneIDAetraCore && (a.StateKey == "core/global-lock" || a.StateKey == "core/*" || a.StateKey == "*")
}

func (b BlockSTMMessageBatch) Validate() error {
	if err := ValidateZoneID(b.SourceZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(b.SourceShardID); err != nil {
		return err
	}
	if err := ValidateZoneID(b.DestinationZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(b.DestinationShardID); err != nil {
		return err
	}
	if b.MessageCount == 0 {
		return errors.New("aetracore BlockSTM message batch count must be positive")
	}
	if err := ValidateHash("aetracore BlockSTM message batch hash", b.BatchHash); err != nil {
		return err
	}
	if b.BatchHash != ComputeBlockSTMMessageBatchHash(b) {
		return errors.New("aetracore BlockSTM message batch hash mismatch")
	}
	return nil
}

func (b BlockSTMMessageBatch) Normalize() BlockSTMMessageBatch {
	b.BatchHash = strings.ToLower(strings.TrimSpace(b.BatchHash))
	return b
}

func (c BlockSTMConflictSet) Validate() error {
	if c.ConflictKey == "" {
		return errors.New("aetracore BlockSTM conflict key is required")
	}
	if len(c.WorkloadIDs) < 2 {
		return errors.New("aetracore BlockSTM conflict set requires at least two workloads")
	}
	seen := make(map[string]struct{}, len(c.WorkloadIDs))
	var previous string
	for i, id := range c.WorkloadIDs {
		if err := ValidateHash("aetracore BlockSTM conflict workload id", id); err != nil {
			return err
		}
		if _, found := seen[id]; found {
			return fmt.Errorf("duplicate aetracore BlockSTM conflict workload %s", id)
		}
		seen[id] = struct{}{}
		if i > 0 && previous >= id {
			return errors.New("aetracore BlockSTM conflict workloads must be sorted canonically")
		}
		previous = id
	}
	return nil
}

func IsBlockSTMAccessMode(mode BlockSTMAccessMode) bool {
	switch mode {
	case BlockSTMAccessRead, BlockSTMAccessWrite:
		return true
	default:
		return false
	}
}

func ComputeBlockSTMWorkloadID(workload BlockSTMZoneWorkload) string {
	workload = workload.Normalize()
	parts := []string{"aetra-blockstm-zone-workload-v1", string(workload.ZoneID), string(workload.ShardID), fmt.Sprint(len(workload.Items))}
	for _, item := range workload.Items {
		parts = append(parts, ComputeProposalItemHash(item))
	}
	return hashParts(parts...)
}

func ComputeBlockSTMConflictKeyRoot(accesses []BlockSTMStateAccess) string {
	ordered := normalizeBlockSTMStateAccesses(accesses)
	parts := []string{"aetra-blockstm-conflict-key-root-v1", fmt.Sprint(len(ordered))}
	for _, access := range ordered {
		parts = append(parts, string(access.Mode), access.ConflictKey(), fmt.Sprint(access.ViaMessage))
	}
	return hashParts(parts...)
}

func ComputeBlockSTMMessageBatchHash(batch BlockSTMMessageBatch) string {
	batch = batch.Normalize()
	return hashParts(
		"aetra-blockstm-message-batch-v1",
		string(batch.SourceZoneID),
		string(batch.SourceShardID),
		string(batch.DestinationZoneID),
		string(batch.DestinationShardID),
		fmt.Sprint(batch.MessageCount),
	)
}

func ComputeBlockSTMMessageBatchRoot(batches []BlockSTMMessageBatch) string {
	ordered := normalizeBlockSTMMessageBatches(batches)
	parts := []string{"aetra-blockstm-message-batch-root-v1", fmt.Sprint(len(ordered))}
	for _, batch := range ordered {
		parts = append(parts, batch.BatchHash)
	}
	return hashParts(parts...)
}

func ComputeBlockSTMZonePerformancePlanHash(plan BlockSTMZonePerformancePlan) string {
	plan = plan.Normalize()
	parts := []string{
		"aetra-blockstm-zone-performance-plan-v1",
		fmt.Sprint(plan.Height),
		fmt.Sprint(plan.ParallelWorkloads),
		fmt.Sprint(plan.GlobalWriteLocks),
		fmt.Sprint(plan.CrossZoneWrites),
	}
	for _, workload := range plan.Workloads {
		parts = append(parts, workload.WorkloadID, workload.ConflictKeyRoot, workload.MessageBatchRoot)
	}
	for _, conflict := range plan.ConflictSets {
		parts = append(parts, conflict.ConflictKey)
		parts = append(parts, conflict.WorkloadIDs...)
	}
	return hashParts(parts...)
}

func normalizeBlockSTMStateAccesses(values []BlockSTMStateAccess) []BlockSTMStateAccess {
	out := make([]BlockSTMStateAccess, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := blockSTMAccessSortKey(out[i])
		right := blockSTMAccessSortKey(out[j])
		return left < right
	})
	return out
}

func normalizeBlockSTMMessageBatches(values []BlockSTMMessageBatch) []BlockSTMMessageBatch {
	out := make([]BlockSTMMessageBatch, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.BatchHash == "" {
			normalized.BatchHash = ComputeBlockSTMMessageBatchHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return blockSTMMessageBatchSortKey(out[i]) < blockSTMMessageBatchSortKey(out[j])
	})
	return out
}

func normalizeBlockSTMZoneWorkloads(values []BlockSTMZoneWorkload) []BlockSTMZoneWorkload {
	out := make([]BlockSTMZoneWorkload, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].WorkloadID < out[j].WorkloadID
	})
	return out
}

func normalizeBlockSTMConflictSets(values []BlockSTMConflictSet) []BlockSTMConflictSet {
	out := make([]BlockSTMConflictSet, len(values))
	for i, value := range values {
		value.WorkloadIDs = append([]string(nil), value.WorkloadIDs...)
		sort.Strings(value.WorkloadIDs)
		out[i] = value
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ConflictKey < out[j].ConflictKey
	})
	return out
}

func computeBlockSTMConflictSets(workloads []BlockSTMZoneWorkload) []BlockSTMConflictSet {
	owners := make(map[string]map[string]struct{})
	for _, workload := range workloads {
		for _, access := range workload.StateAccesses {
			if access.Mode != BlockSTMAccessWrite || access.ViaMessage {
				continue
			}
			key := access.ConflictKey()
			if owners[key] == nil {
				owners[key] = make(map[string]struct{})
			}
			owners[key][workload.WorkloadID] = struct{}{}
		}
	}
	out := make([]BlockSTMConflictSet, 0)
	for key, ids := range owners {
		if len(ids) < 2 {
			continue
		}
		conflict := BlockSTMConflictSet{ConflictKey: key}
		for id := range ids {
			conflict.WorkloadIDs = append(conflict.WorkloadIDs, id)
		}
		sort.Strings(conflict.WorkloadIDs)
		out = append(out, conflict)
	}
	return normalizeBlockSTMConflictSets(out)
}

func countBlockSTMCrossZoneMessageWrites(accesses []BlockSTMStateAccess) uint32 {
	var count uint32
	for _, access := range accesses {
		if access.Mode == BlockSTMAccessWrite && access.StateZoneID != access.ActorZoneID && access.ViaMessage {
			count++
		}
	}
	return count
}

func blockSTMAccessSortKey(access BlockSTMStateAccess) string {
	return strings.Join([]string{
		string(access.ActorZoneID),
		string(access.ActorShardID),
		string(access.StateZoneID),
		string(access.StateShardID),
		access.StateKey,
		string(access.Mode),
		fmt.Sprint(access.ViaMessage),
	}, "/")
}

func blockSTMMessageBatchSortKey(batch BlockSTMMessageBatch) string {
	return strings.Join([]string{
		string(batch.SourceZoneID),
		string(batch.SourceShardID),
		string(batch.DestinationZoneID),
		string(batch.DestinationShardID),
		fmt.Sprintf("%020d", batch.MessageCount),
		batch.BatchHash,
	}, "/")
}
