package types

import (
	"errors"
	"fmt"
	"sort"
)

type ShardLayoutChangeKind string

const (
	ShardLayoutChangeNone	ShardLayoutChangeKind	= "none"
	ShardLayoutChangeSplit	ShardLayoutChangeKind	= "split"
	ShardLayoutChangeMerge	ShardLayoutChangeKind	= "merge"
)

type ShardRebalanceThresholds struct {
	GasLimitPerShard		uint64
	SplitGasUtilization		uint64
	SplitStateSizeBytes		uint64
	SplitWriteConflictCount		uint64
	SplitQueueBacklog		uint64
	SplitProofLatencyMicros		uint64
	MergeGasUtilization		uint64
	MergeStateSizeBytes		uint64
	MergeQueueBacklog		uint64
	MergeWriteConflictCount		uint64
	DecisionWindow			uint64
	FutureLayoutEpochDelta		uint64
	FutureActivationHeightGap	uint64
}

type ShardRebalanceDecision struct {
	ZoneID			ZoneID
	SourceLayoutEpoch	uint64
	TargetLayoutEpoch	uint64
	DecisionHeight		uint64
	ActivationHeight	uint64
	ChangeKind		ShardLayoutChangeKind
	Reason			string
	SourceShardIDs		[]ShardID
	TargetShardIDs		[]ShardID
	SourceMetricsHash	string
	SourceLayoutHash	string
	MigrationTasks		[]ShardMigrationTask
	DecisionHash		string
}

type ShardMigrationTask struct {
	TaskID			string
	ZoneID			ZoneID
	SourceShardID		ShardID
	DestinationShardID	ShardID
	SourceLayoutEpoch	uint64
	TargetLayoutEpoch	uint64
	KeyPrefix		string
	ObjectKey		string
	HashRangeStart		uint64
	HashRangeEnd		uint64
	DeliveryEpoch		uint64
	TaskHash		string
}

type ObjectLock struct {
	ZoneID		ZoneID
	ShardID		ShardID
	ObjectID	string
	LockID		string
	Reason		string
	OwnerTaskID	string
	CreatedHeight	uint64
	ExpiryHeight	uint64
	LockHash	string
}

type ShardRoot struct {
	ZoneID		ZoneID
	ShardID		ShardID
	Height		uint64
	StateRoot	string
	InboxRoot	string
	OutboxRoot	string
	ReceiptsRoot	string
	MetricsHash	string
	RootHash	string
}

func NewShardRebalanceDecision(layout ShardLayout, metrics []ShardMetrics, thresholds ShardRebalanceThresholds, decisionHeight uint64) (ShardRebalanceDecision, error) {
	if err := layout.ValidateHash(); err != nil {
		return ShardRebalanceDecision{}, err
	}
	if decisionHeight == 0 {
		return ShardRebalanceDecision{}, errors.New("aetracore shard rebalance decision height must be positive")
	}
	if err := thresholds.Validate(); err != nil {
		return ShardRebalanceDecision{}, err
	}
	ordered := cloneShardMetrics(metrics)
	sortShardMetrics(ordered)
	if len(ordered) == 0 {
		return ShardRebalanceDecision{}, errors.New("aetracore shard rebalance requires committed metrics")
	}
	for _, metric := range ordered {
		if err := metric.ValidateHash(); err != nil {
			return ShardRebalanceDecision{}, err
		}
		if metric.ZoneID != layout.ZoneID {
			return ShardRebalanceDecision{}, errors.New("aetracore shard rebalance metrics zone mismatch")
		}
		if !layout.HasActiveShard(metric.ShardID) {
			return ShardRebalanceDecision{}, errors.New("aetracore shard rebalance metrics shard is not active")
		}
		if metric.Height > decisionHeight {
			return ShardRebalanceDecision{}, errors.New("aetracore shard rebalance cannot use future metrics")
		}
	}
	decision := ShardRebalanceDecision{
		ZoneID:			layout.ZoneID,
		SourceLayoutEpoch:	layout.LayoutEpoch,
		TargetLayoutEpoch:	layout.LayoutEpoch + thresholds.FutureLayoutEpochDelta,
		DecisionHeight:		decisionHeight,
		ActivationHeight:	decisionHeight + thresholds.FutureActivationHeightGap,
		ChangeKind:		ShardLayoutChangeNone,
		SourceLayoutHash:	layout.LayoutHash,
		SourceMetricsHash:	ComputeShardMetricsWindowHash(ordered),
	}
	if splitShard, reason, ok := selectSplitShard(ordered, thresholds); ok {
		destination := nextShardID(layout)
		task, err := NewShardMigrationTask(ShardMigrationTask{
			ZoneID:			layout.ZoneID,
			SourceShardID:		splitShard,
			DestinationShardID:	destination,
			SourceLayoutEpoch:	layout.LayoutEpoch,
			TargetLayoutEpoch:	decision.TargetLayoutEpoch,
			DeliveryEpoch:		decision.TargetLayoutEpoch,
		})
		if err != nil {
			return ShardRebalanceDecision{}, err
		}
		decision.ChangeKind = ShardLayoutChangeSplit
		decision.Reason = reason
		decision.SourceShardIDs = []ShardID{splitShard}
		decision.TargetShardIDs = []ShardID{splitShard, destination}
		decision.MigrationTasks = []ShardMigrationTask{task}
	} else if mergeA, mergeB, reason, ok := selectMergeShards(ordered, thresholds); ok {
		task, err := NewShardMigrationTask(ShardMigrationTask{
			ZoneID:			layout.ZoneID,
			SourceShardID:		mergeB,
			DestinationShardID:	mergeA,
			SourceLayoutEpoch:	layout.LayoutEpoch,
			TargetLayoutEpoch:	decision.TargetLayoutEpoch,
			DeliveryEpoch:		decision.TargetLayoutEpoch,
		})
		if err != nil {
			return ShardRebalanceDecision{}, err
		}
		decision.ChangeKind = ShardLayoutChangeMerge
		decision.Reason = reason
		decision.SourceShardIDs = []ShardID{mergeA, mergeB}
		decision.TargetShardIDs = []ShardID{mergeA}
		decision.MigrationTasks = []ShardMigrationTask{task}
	}
	sortShardIDs(decision.SourceShardIDs)
	sortShardIDs(decision.TargetShardIDs)
	sortShardMigrationTasks(decision.MigrationTasks)
	decision.DecisionHash = ComputeShardRebalanceDecisionHash(decision)
	return decision, decision.ValidateHash()
}

func (t ShardRebalanceThresholds) Validate() error {
	if t.GasLimitPerShard == 0 {
		return errors.New("aetracore shard rebalance gas limit must be positive")
	}
	if t.SplitGasUtilization == 0 {
		return errors.New("aetracore shard split gas utilization threshold must be positive")
	}
	if t.MergeGasUtilization > t.SplitGasUtilization {
		return errors.New("aetracore shard merge gas threshold must not exceed split threshold")
	}
	if t.DecisionWindow == 0 {
		return errors.New("aetracore shard rebalance decision window must be positive")
	}
	if t.FutureLayoutEpochDelta == 0 {
		return errors.New("aetracore shard rebalance future layout epoch delta must be positive")
	}
	if t.FutureActivationHeightGap == 0 {
		return errors.New("aetracore shard rebalance activation height gap must be positive")
	}
	return nil
}

func (d ShardRebalanceDecision) ValidateHash() error {
	if err := d.ValidateFormat(); err != nil {
		return err
	}
	if d.DecisionHash != ComputeShardRebalanceDecisionHash(d) {
		return errors.New("aetracore shard rebalance decision hash mismatch")
	}
	return nil
}

func (d ShardRebalanceDecision) ValidateFormat() error {
	if err := ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	if d.SourceLayoutEpoch == 0 || d.TargetLayoutEpoch == 0 {
		return errors.New("aetracore shard rebalance layout epochs must be positive")
	}
	if d.TargetLayoutEpoch <= d.SourceLayoutEpoch {
		return errors.New("aetracore shard rebalance target epoch must be future")
	}
	if d.DecisionHeight == 0 || d.ActivationHeight == 0 {
		return errors.New("aetracore shard rebalance heights must be positive")
	}
	if d.ActivationHeight <= d.DecisionHeight {
		return errors.New("aetracore shard rebalance activation height must be future")
	}
	if !IsShardLayoutChangeKind(d.ChangeKind) {
		return fmt.Errorf("unknown aetracore shard layout change kind %q", d.ChangeKind)
	}
	if d.ChangeKind != ShardLayoutChangeNone {
		if err := validateToken("aetracore shard rebalance reason", d.Reason, MaxScopeLength); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore shard rebalance metrics hash", d.SourceMetricsHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore shard rebalance source layout hash", d.SourceLayoutHash); err != nil {
		return err
	}
	if err := validateShardIDList("aetracore shard rebalance source shard", d.SourceShardIDs); err != nil {
		return err
	}
	if err := validateShardIDList("aetracore shard rebalance target shard", d.TargetShardIDs); err != nil {
		return err
	}
	if err := validateShardMigrationTasks(d.MigrationTasks, d); err != nil {
		return err
	}
	if d.DecisionHash != "" {
		return ValidateHash("aetracore shard rebalance decision hash", d.DecisionHash)
	}
	return nil
}

func NewShardMigrationTask(task ShardMigrationTask) (ShardMigrationTask, error) {
	if task.TaskHash != "" || task.TaskID != "" {
		return ShardMigrationTask{}, errors.New("aetracore shard migration task id and hash must be empty before construction")
	}
	if err := task.ValidateFormat(); err != nil {
		return ShardMigrationTask{}, err
	}
	task.TaskHash = ComputeShardMigrationTaskHash(task)
	task.TaskID = task.TaskHash[:16]
	return task, task.ValidateHash()
}

func (t ShardMigrationTask) ValidateHash() error {
	if err := t.ValidateFormat(); err != nil {
		return err
	}
	if t.TaskHash != ComputeShardMigrationTaskHash(t) {
		return errors.New("aetracore shard migration task hash mismatch")
	}
	if t.TaskID != t.TaskHash[:16] {
		return errors.New("aetracore shard migration task id mismatch")
	}
	return nil
}

func (t ShardMigrationTask) ValidateFormat() error {
	if err := ValidateZoneID(t.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(t.SourceShardID); err != nil {
		return err
	}
	if err := ValidateShardID(t.DestinationShardID); err != nil {
		return err
	}
	if t.SourceShardID == t.DestinationShardID {
		return errors.New("aetracore shard migration source and destination must differ")
	}
	if t.SourceLayoutEpoch == 0 || t.TargetLayoutEpoch == 0 || t.DeliveryEpoch == 0 {
		return errors.New("aetracore shard migration epochs must be positive")
	}
	if t.TargetLayoutEpoch <= t.SourceLayoutEpoch {
		return errors.New("aetracore shard migration target epoch must be future")
	}
	if t.DeliveryEpoch < t.TargetLayoutEpoch {
		return errors.New("aetracore shard migration delivery epoch must not precede target layout epoch")
	}
	if t.KeyPrefix != "" {
		if err := validateToken("aetracore shard migration key prefix", t.KeyPrefix, MaxScopeLength); err != nil {
			return err
		}
	}
	if t.ObjectKey != "" {
		if err := validateToken("aetracore shard migration object key", t.ObjectKey, MaxScopeLength); err != nil {
			return err
		}
	}
	if t.HashRangeEnd < t.HashRangeStart {
		return errors.New("aetracore shard migration hash range end must not precede start")
	}
	if t.TaskID != "" {
		if err := validateToken("aetracore shard migration task id", t.TaskID, MaxIDLength); err != nil {
			return err
		}
	}
	if t.TaskHash != "" {
		return ValidateHash("aetracore shard migration task hash", t.TaskHash)
	}
	return nil
}

func NewObjectLock(lock ObjectLock) (ObjectLock, error) {
	if lock.LockHash != "" || lock.LockID != "" {
		return ObjectLock{}, errors.New("aetracore object lock id and hash must be empty before construction")
	}
	if err := lock.ValidateFormat(); err != nil {
		return ObjectLock{}, err
	}
	lock.LockHash = ComputeObjectLockHash(lock)
	lock.LockID = lock.LockHash[:16]
	return lock, lock.ValidateHash()
}

func (l ObjectLock) ValidateHash() error {
	if err := l.ValidateFormat(); err != nil {
		return err
	}
	if l.LockHash != ComputeObjectLockHash(l) {
		return errors.New("aetracore object lock hash mismatch")
	}
	if l.LockID != l.LockHash[:16] {
		return errors.New("aetracore object lock id mismatch")
	}
	return nil
}

func (l ObjectLock) ValidateFormat() error {
	if err := ValidateZoneID(l.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(l.ShardID); err != nil {
		return err
	}
	if err := validateToken("aetracore object lock object id", l.ObjectID, MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore object lock reason", l.Reason, MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore object lock owner task id", l.OwnerTaskID, MaxIDLength); err != nil {
		return err
	}
	if l.CreatedHeight == 0 || l.ExpiryHeight == 0 {
		return errors.New("aetracore object lock heights must be positive")
	}
	if l.ExpiryHeight <= l.CreatedHeight {
		return errors.New("aetracore object lock expiry height must be future")
	}
	if l.LockID != "" {
		if err := validateToken("aetracore object lock id", l.LockID, MaxIDLength); err != nil {
			return err
		}
	}
	if l.LockHash != "" {
		return ValidateHash("aetracore object lock hash", l.LockHash)
	}
	return nil
}

func NewShardRoot(root ShardRoot) (ShardRoot, error) {
	if root.RootHash != "" {
		return ShardRoot{}, errors.New("aetracore shard root hash must be empty before construction")
	}
	if err := root.ValidateFormat(); err != nil {
		return ShardRoot{}, err
	}
	root.RootHash = ComputeShardRootHash(root)
	return root, root.ValidateHash()
}

func (r ShardRoot) ValidateHash() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.RootHash != ComputeShardRootHash(r) {
		return errors.New("aetracore shard root hash mismatch")
	}
	return nil
}

func (r ShardRoot) ValidateFormat() error {
	if err := ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(r.ShardID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("aetracore shard root height must be positive")
	}
	for field, value := range map[string]string{
		"state":	r.StateRoot,
		"inbox":	r.InboxRoot,
		"outbox":	r.OutboxRoot,
		"receipts":	r.ReceiptsRoot,
		"metrics":	r.MetricsHash,
	} {
		if err := ValidateHash("aetracore shard "+field+" root", value); err != nil {
			return err
		}
	}
	if r.RootHash != "" {
		return ValidateHash("aetracore shard root hash", r.RootHash)
	}
	return nil
}

func ComputeShardMetricsWindowHash(metrics []ShardMetrics) string {
	ordered := cloneShardMetrics(metrics)
	sortShardMetrics(ordered)
	parts := []string{"aetra-aek-shard-metrics-window-v1", fmt.Sprint(len(ordered))}
	for _, metric := range ordered {
		parts = append(parts, string(metric.ZoneID), string(metric.ShardID), fmt.Sprint(metric.Height), metric.MetricsHash)
	}
	return hashParts(parts...)
}

func ComputeShardRebalanceDecisionHash(decision ShardRebalanceDecision) string {
	sourceIDs := append([]ShardID(nil), decision.SourceShardIDs...)
	targetIDs := append([]ShardID(nil), decision.TargetShardIDs...)
	tasks := cloneShardMigrationTasks(decision.MigrationTasks)
	sortShardIDs(sourceIDs)
	sortShardIDs(targetIDs)
	sortShardMigrationTasks(tasks)
	parts := []string{
		"aetra-aek-shard-rebalance-decision-v1",
		string(decision.ZoneID),
		fmt.Sprint(decision.SourceLayoutEpoch),
		fmt.Sprint(decision.TargetLayoutEpoch),
		fmt.Sprint(decision.DecisionHeight),
		fmt.Sprint(decision.ActivationHeight),
		string(decision.ChangeKind),
		decision.Reason,
		decision.SourceMetricsHash,
		decision.SourceLayoutHash,
		fmt.Sprint(len(sourceIDs)),
	}
	for _, id := range sourceIDs {
		parts = append(parts, string(id))
	}
	parts = append(parts, fmt.Sprint(len(targetIDs)))
	for _, id := range targetIDs {
		parts = append(parts, string(id))
	}
	parts = append(parts, fmt.Sprint(len(tasks)))
	for _, task := range tasks {
		parts = append(parts, task.TaskHash)
	}
	return hashParts(parts...)
}

func ComputeShardMigrationTaskHash(task ShardMigrationTask) string {
	return hashParts(
		"aetra-aek-shard-migration-task-v1",
		string(task.ZoneID),
		string(task.SourceShardID),
		string(task.DestinationShardID),
		fmt.Sprint(task.SourceLayoutEpoch),
		fmt.Sprint(task.TargetLayoutEpoch),
		task.KeyPrefix,
		task.ObjectKey,
		fmt.Sprint(task.HashRangeStart),
		fmt.Sprint(task.HashRangeEnd),
		fmt.Sprint(task.DeliveryEpoch),
	)
}

func ComputeObjectLockHash(lock ObjectLock) string {
	return hashParts(
		"aetra-aek-object-lock-v1",
		string(lock.ZoneID),
		string(lock.ShardID),
		lock.ObjectID,
		lock.Reason,
		lock.OwnerTaskID,
		fmt.Sprint(lock.CreatedHeight),
		fmt.Sprint(lock.ExpiryHeight),
	)
}

func ComputeShardRootHash(root ShardRoot) string {
	return hashParts(
		"aetra-aek-shard-root-v1",
		string(root.ZoneID),
		string(root.ShardID),
		fmt.Sprint(root.Height),
		root.StateRoot,
		root.InboxRoot,
		root.OutboxRoot,
		root.ReceiptsRoot,
		root.MetricsHash,
	)
}

func ComputeShardRootsRoot(roots []ShardRoot) (string, error) {
	ordered := cloneShardRoots(roots)
	sortShardRoots(ordered)
	if len(ordered) == 0 {
		return "", errors.New("aetracore shard roots root requires shard roots")
	}
	parts := []string{"aetra-aek-shard-roots-root-v1", fmt.Sprint(len(ordered))}
	for _, root := range ordered {
		if err := root.ValidateHash(); err != nil {
			return "", err
		}
		parts = append(parts, string(root.ZoneID), string(root.ShardID), fmt.Sprint(root.Height), root.RootHash)
	}
	return hashParts(parts...), nil
}

func IsShardLayoutChangeKind(kind ShardLayoutChangeKind) bool {
	switch kind {
	case ShardLayoutChangeNone, ShardLayoutChangeSplit, ShardLayoutChangeMerge:
		return true
	default:
		return false
	}
}

func selectSplitShard(metrics []ShardMetrics, thresholds ShardRebalanceThresholds) (ShardID, string, bool) {
	ordered := cloneShardMetrics(metrics)
	sortShardMetrics(ordered)
	byShard := shardMetricsByShard(ordered)
	shardIDs := make([]ShardID, 0, len(byShard))
	for shardID := range byShard {
		shardIDs = append(shardIDs, shardID)
	}
	sortShardIDs(shardIDs)
	window := int(thresholds.DecisionWindow)
	for _, shardID := range shardIDs {
		history := byShard[shardID]
		if len(history) < window {
			continue
		}
		reasonCounts := make(map[string]int)
		for _, metric := range history[len(history)-window:] {
			for _, reason := range splitReasons(metric, thresholds) {
				reasonCounts[reason]++
			}
		}
		for _, reason := range []string{"split_gas_utilization", "split_state_size", "split_write_conflict_rate", "split_queue_backlog", "split_proof_latency"} {
			if reasonCounts[reason] == window {
				return shardID, reason, true
			}
		}
	}
	return "", "", false
}

func selectMergeShards(metrics []ShardMetrics, thresholds ShardRebalanceThresholds) (ShardID, ShardID, string, bool) {
	ordered := cloneShardMetrics(metrics)
	sortShardMetrics(ordered)
	byShard := shardMetricsByShard(ordered)
	shardIDs := make([]ShardID, 0, len(byShard))
	for shardID := range byShard {
		shardIDs = append(shardIDs, shardID)
	}
	sortShardIDs(shardIDs)
	candidates := make([]ShardID, 0, len(ordered))
	window := int(thresholds.DecisionWindow)
	for _, shardID := range shardIDs {
		history := byShard[shardID]
		if len(history) < window {
			continue
		}
		lowWindow := true
		for _, metric := range history[len(history)-window:] {
			if !mergeEligible(metric, thresholds) {
				lowWindow = false
				break
			}
		}
		if lowWindow {
			candidates = append(candidates, shardID)
		}
		if len(candidates) == 2 {
			return candidates[0], candidates[1], "merge_low_utilization", true
		}
	}
	return "", "", "", false
}

func splitReasons(metric ShardMetrics, thresholds ShardRebalanceThresholds) []string {
	reasons := make([]string, 0, 5)
	if metric.GasUsed >= thresholds.GasLimitPerShard*thresholds.SplitGasUtilization/100 {
		reasons = append(reasons, "split_gas_utilization")
	}
	if thresholds.SplitStateSizeBytes > 0 && metric.StateSizeBytes >= thresholds.SplitStateSizeBytes {
		reasons = append(reasons, "split_state_size")
	}
	if thresholds.SplitWriteConflictCount > 0 && metric.WriteConflictCount >= thresholds.SplitWriteConflictCount {
		reasons = append(reasons, "split_write_conflict_rate")
	}
	if thresholds.SplitQueueBacklog > 0 && metric.InboxBacklog+metric.OutboxBacklog >= thresholds.SplitQueueBacklog {
		reasons = append(reasons, "split_queue_backlog")
	}
	if thresholds.SplitProofLatencyMicros > 0 && metric.ProofLatencyMicros >= thresholds.SplitProofLatencyMicros {
		reasons = append(reasons, "split_proof_latency")
	}
	return reasons
}

func mergeEligible(metric ShardMetrics, thresholds ShardRebalanceThresholds) bool {
	lowGas := metric.GasUsed <= thresholds.GasLimitPerShard*thresholds.MergeGasUtilization/100
	lowSize := thresholds.MergeStateSizeBytes == 0 || metric.StateSizeBytes <= thresholds.MergeStateSizeBytes
	lowQueue := thresholds.MergeQueueBacklog == 0 || metric.InboxBacklog+metric.OutboxBacklog <= thresholds.MergeQueueBacklog
	lowConflict := thresholds.MergeWriteConflictCount == 0 || metric.WriteConflictCount <= thresholds.MergeWriteConflictCount
	return lowGas && lowSize && lowQueue && lowConflict
}

func shardMetricsByShard(metrics []ShardMetrics) map[ShardID][]ShardMetrics {
	out := make(map[ShardID][]ShardMetrics)
	for _, metric := range metrics {
		out[metric.ShardID] = append(out[metric.ShardID], metric)
	}
	for shardID := range out {
		sortShardMetrics(out[shardID])
	}
	return out
}

func nextShardID(layout ShardLayout) ShardID {
	active := availableShardDescriptors(layout.ActiveShards)
	if len(active) == 0 {
		return "0"
	}
	max := active[len(active)-1].ShardID
	return ShardID(fmt.Sprintf("%s_split_%020d", max, layout.LayoutEpoch+1))
}

func validateShardIDList(fieldName string, ids []ShardID) error {
	ordered := append([]ShardID(nil), ids...)
	sortShardIDs(ordered)
	var previous ShardID
	for i, id := range ordered {
		if err := ValidateShardID(id); err != nil {
			return err
		}
		if i > 0 && previous >= id {
			return fmt.Errorf("%s ids must be sorted canonically", fieldName)
		}
		previous = id
	}
	return nil
}

func validateShardMigrationTasks(tasks []ShardMigrationTask, decision ShardRebalanceDecision) error {
	ordered := cloneShardMigrationTasks(tasks)
	sortShardMigrationTasks(ordered)
	for _, task := range ordered {
		if err := task.ValidateHash(); err != nil {
			return err
		}
		if task.ZoneID != decision.ZoneID {
			return errors.New("aetracore shard migration task zone mismatch")
		}
		if task.SourceLayoutEpoch != decision.SourceLayoutEpoch || task.TargetLayoutEpoch != decision.TargetLayoutEpoch {
			return errors.New("aetracore shard migration task epoch mismatch")
		}
	}
	return nil
}

func sortShardIDs(ids []ShardID) {
	sort.SliceStable(ids, func(i, j int) bool { return ids[i] < ids[j] })
}

func sortShardMetrics(metrics []ShardMetrics) {
	sort.SliceStable(metrics, func(i, j int) bool {
		if metrics[i].Height == metrics[j].Height {
			return metrics[i].ShardID < metrics[j].ShardID
		}
		return metrics[i].Height < metrics[j].Height
	})
}

func sortShardMigrationTasks(tasks []ShardMigrationTask) {
	sort.SliceStable(tasks, func(i, j int) bool { return tasks[i].TaskHash < tasks[j].TaskHash })
}

func sortShardRoots(roots []ShardRoot) {
	sort.SliceStable(roots, func(i, j int) bool {
		if roots[i].ZoneID == roots[j].ZoneID {
			if roots[i].Height == roots[j].Height {
				return roots[i].ShardID < roots[j].ShardID
			}
			return roots[i].Height < roots[j].Height
		}
		return roots[i].ZoneID < roots[j].ZoneID
	})
}

func cloneShardMetrics(metrics []ShardMetrics) []ShardMetrics {
	out := make([]ShardMetrics, len(metrics))
	copy(out, metrics)
	return out
}

func cloneShardMigrationTasks(tasks []ShardMigrationTask) []ShardMigrationTask {
	out := make([]ShardMigrationTask, len(tasks))
	copy(out, tasks)
	return out
}

func cloneShardRoots(roots []ShardRoot) []ShardRoot {
	out := make([]ShardRoot, len(roots))
	copy(out, roots)
	return out
}
