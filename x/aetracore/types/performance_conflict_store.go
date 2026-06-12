package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ConflictReductionSpecVersion	= uint64(1)
	StoreV2UsageSpecVersion		= uint64(1)
)

type ConflictReductionRuleID string
type StoreV2LayoutKind string
type StoreV2BenchmarkID string

const (
	ConflictRuleAvoidGlobalCounters	ConflictReductionRuleID	= "avoid-global-counters"
	ConflictRuleShardFeeAccumulator	ConflictReductionRuleID	= "per-shard-fee-accumulators"
	ConflictRuleShardMessageQueues	ConflictReductionRuleID	= "per-shard-message-queues"
	ConflictRuleVersionedObjects	ConflictReductionRuleID	= "versioned-object-updates"
	ConflictRuleObjectLocalLocks	ConflictReductionRuleID	= "object-local-locks"
	ConflictRuleAsyncRemoteWrites	ConflictReductionRuleID	= "async-remote-writes"

	StoreV2LayoutObjectStore	StoreV2LayoutKind	= "object-store"
	StoreV2LayoutKVHybrid		StoreV2LayoutKind	= "kv-hybrid"
	StoreV2LayoutPrefixProof	StoreV2LayoutKind	= "prefix-proof"
	StoreV2LayoutCompactRoot	StoreV2LayoutKind	= "compact-root"
	StoreV2LayoutBoundedRangeScan	StoreV2LayoutKind	= "bounded-range-scan"

	StoreV2BenchmarkDirectBalanceRead	StoreV2BenchmarkID	= "direct-balance-read"
	StoreV2BenchmarkDirectIdentityResolve	StoreV2BenchmarkID	= "direct-identity-resolution"
	StoreV2BenchmarkRecursiveIdentity	StoreV2BenchmarkID	= "recursive-identity-resolution"
	StoreV2BenchmarkContractStorageRW	StoreV2BenchmarkID	= "contract-storage-read-write"
	StoreV2BenchmarkMessageEnqueueDequeue	StoreV2BenchmarkID	= "message-enqueue-dequeue"
	StoreV2BenchmarkPaymentChannelSettle	StoreV2BenchmarkID	= "payment-channel-settle"
	StoreV2BenchmarkDEXPoolUpdate		StoreV2BenchmarkID	= "dex-pool-update"
	StoreV2BenchmarkProofGeneration		StoreV2BenchmarkID	= "proof-generation"
)

type ConflictReductionRule struct {
	RuleID		ConflictReductionRuleID
	Rule		string
	Enforcement	string
	Evidence	string
	DescriptorHash	string
}

type ConflictReductionSpec struct {
	Version	uint64
	Rules	[]ConflictReductionRule
	Root	string
}

type StoreV2LayoutDescriptor struct {
	LayoutID	string
	Kind		StoreV2LayoutKind
	StateClasses	[]string
	KeyPattern	string
	ProofSurface	string
	BoundedScan	bool
	DescriptorHash	string
}

type StoreV2BenchmarkRequirement struct {
	BenchmarkID	StoreV2BenchmarkID
	Operation	string
	LayoutID	string
	Metric		string
	MaxRangeItems	uint32
	Required	bool
	DescriptorHash	string
}

type StoreV2UsageSpec struct {
	Version		uint64
	Layouts		[]StoreV2LayoutDescriptor
	Benchmarks	[]StoreV2BenchmarkRequirement
	Root		string
}

func DefaultConflictReductionSpec() (ConflictReductionSpec, error) {
	return BuildConflictReductionSpec(ConflictReductionRules())
}

func ConflictReductionRules() []ConflictReductionRule {
	return []ConflictReductionRule{
		conflictRule(ConflictRuleAvoidGlobalCounters, "Avoid global counters in hot paths.", "Reject hot-path write scopes that use core/global counters or wildcard write locks.", "BlockSTMStateAccess.IsGlobalWriteLock;BlockSTMZonePerformancePlan.GlobalWriteLocks"),
		conflictRule(ConflictRuleShardFeeAccumulator, "Use per-shard fee accumulators.", "Fee writes use zone/{zone_id}/shard/{shard_id}/metrics or fee accumulator keys instead of global counters.", "zone/{zone_id}/shard/{shard_id}/metrics/*;per-shard fee roots"),
		conflictRule(ConflictRuleShardMessageQueues, "Use per-shard message queues.", "Message batches are keyed by source and destination shard and committed through shard-local inbox/outbox roots.", "BlockSTMMessageBatch;ComputeBlockSTMMessageBatchRoot"),
		conflictRule(ConflictRuleVersionedObjects, "Use versioned object updates.", "Hot objects carry deterministic version keys so same-object updates conflict predictably.", "BlockSTMStateAccess.ConflictKey;versioned object key suffixes"),
		conflictRule(ConflictRuleObjectLocalLocks, "Use object-local locks only for multi-step local operations.", "Locks must target one object key inside the actor zone and shard, never a global or remote namespace.", "object-local lock conflict keys"),
		conflictRule(ConflictRuleAsyncRemoteWrites, "Use async messages instead of synchronous remote writes.", "Direct cross-zone writes are rejected unless represented by ViaMessage and a committed message batch.", "BuildBlockSTMZonePerformancePlan cross-zone write validation"),
	}
}

func BuildConflictReductionSpec(rules []ConflictReductionRule) (ConflictReductionSpec, error) {
	spec := ConflictReductionSpec{
		Version:	ConflictReductionSpecVersion,
		Rules:		normalizeConflictReductionRules(rules),
	}
	if err := spec.ValidateFormat(); err != nil {
		return ConflictReductionSpec{}, err
	}
	spec.Root = ComputeConflictReductionSpecRoot(spec.Rules)
	return spec, spec.Validate()
}

func (s ConflictReductionSpec) Normalize() ConflictReductionSpec {
	if s.Version == 0 {
		s.Version = ConflictReductionSpecVersion
	}
	s.Rules = normalizeConflictReductionRules(s.Rules)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s ConflictReductionSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != ConflictReductionSpecVersion {
		return fmt.Errorf("aetracore conflict reduction spec version must be %d", ConflictReductionSpecVersion)
	}
	if len(s.Rules) == 0 {
		return errors.New("aetracore conflict reduction spec requires rules")
	}
	seen := make(map[ConflictReductionRuleID]struct{}, len(s.Rules))
	var previous ConflictReductionRuleID
	for i, rule := range s.Rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.RuleID]; found {
			return fmt.Errorf("duplicate aetracore conflict reduction rule %s", rule.RuleID)
		}
		seen[rule.RuleID] = struct{}{}
		if i > 0 && previous >= rule.RuleID {
			return errors.New("aetracore conflict reduction rules must be sorted canonically")
		}
		previous = rule.RuleID
	}
	if s.Root != "" {
		if err := ValidateHash("aetracore conflict reduction spec root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s ConflictReductionSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("aetracore conflict reduction spec root is required")
	}
	expected := ComputeConflictReductionSpecRoot(s.Rules)
	if s.Root != expected {
		return fmt.Errorf("aetracore conflict reduction spec root mismatch: expected %s", expected)
	}
	return nil
}

func (r ConflictReductionRule) Normalize() ConflictReductionRule {
	r.Rule = compactPerformanceText(r.Rule)
	r.Enforcement = compactPerformanceText(r.Enforcement)
	r.Evidence = compactPerformanceText(r.Evidence)
	r.DescriptorHash = normalizePerformanceHash(r.DescriptorHash)
	return r
}

func (r ConflictReductionRule) ValidateFormat() error {
	r = r.Normalize()
	if !IsConflictReductionRuleID(r.RuleID) {
		return fmt.Errorf("unknown aetracore conflict reduction rule %q", r.RuleID)
	}
	if r.Rule == "" || r.Enforcement == "" || r.Evidence == "" {
		return errors.New("aetracore conflict reduction rule requires rule, enforcement, and evidence")
	}
	if r.DescriptorHash != "" {
		if err := ValidateHash("aetracore conflict reduction descriptor hash", r.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (r ConflictReductionRule) Validate() error {
	r = r.Normalize()
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.DescriptorHash == "" {
		return errors.New("aetracore conflict reduction descriptor hash is required")
	}
	if expected := ComputeConflictReductionRuleHash(r); r.DescriptorHash != expected {
		return fmt.Errorf("aetracore conflict reduction descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateConflictReductionCoverage() error {
	spec, err := DefaultConflictReductionSpec()
	if err != nil {
		return err
	}
	required := []ConflictReductionRuleID{
		ConflictRuleAvoidGlobalCounters,
		ConflictRuleShardFeeAccumulator,
		ConflictRuleShardMessageQueues,
		ConflictRuleVersionedObjects,
		ConflictRuleObjectLocalLocks,
		ConflictRuleAsyncRemoteWrites,
	}
	seen := make(map[ConflictReductionRuleID]struct{}, len(spec.Rules))
	for _, rule := range spec.Rules {
		seen[rule.RuleID] = struct{}{}
	}
	for _, ruleID := range required {
		if _, found := seen[ruleID]; !found {
			return fmt.Errorf("aetracore conflict reduction spec missing %s", ruleID)
		}
	}
	return nil
}

func DefaultStoreV2UsageSpec() (StoreV2UsageSpec, error) {
	return BuildStoreV2UsageSpec(StoreV2LayoutDescriptors(), StoreV2BenchmarkRequirements())
}

func StoreV2LayoutDescriptors() []StoreV2LayoutDescriptor {
	return []StoreV2LayoutDescriptor{
		storeV2Layout("object-store-records", StoreV2LayoutObjectStore, []string{"accounts", "channels", "contracts", "domains", "pools"}, "zone/{zone_id}/shard/{shard_id}/state/{object_type}/{object_id}", "shard state root", true),
		storeV2Layout("contract-storage-kv", StoreV2LayoutKVHybrid, []string{"contract-storage", "resolver-fields"}, "zone/{zone_id}/shard/{shard_id}/state/{owner}/{key_prefix}/{key}", "contract or resolver proof root", true),
		storeV2Layout("shard-prefix-proofs", StoreV2LayoutPrefixProof, []string{"shard-state", "message-queues", "receipts"}, "zone/{zone_id}/shard/{shard_id}/{prefix}/{key}", "shard prefix proof", true),
		storeV2Layout("zone-compact-roots", StoreV2LayoutCompactRoot, []string{"zone-roots", "message-roots", "receipt-roots"}, "core/zone_roots/{height}/{zone_id}", "global zone root", false),
		storeV2Layout("bounded-range-scans", StoreV2LayoutBoundedRangeScan, []string{"contract-range", "resolver-range", "scheduler-bucket"}, "zone/{zone_id}/shard/{shard_id}/state/{prefix}/{start}:{limit}", "bounded range proof", true),
	}
}

func StoreV2BenchmarkRequirements() []StoreV2BenchmarkRequirement {
	return []StoreV2BenchmarkRequirement{
		storeV2Benchmark(StoreV2BenchmarkDirectBalanceRead, "direct balance read", "object-store-records", "read latency and proof bytes", 1),
		storeV2Benchmark(StoreV2BenchmarkDirectIdentityResolve, "direct identity resolution", "object-store-records", "resolver latency and proof bytes", 1),
		storeV2Benchmark(StoreV2BenchmarkRecursiveIdentity, "recursive identity resolution", "bounded-range-scans", "bounded recursion latency and proof bytes", 32),
		storeV2Benchmark(StoreV2BenchmarkContractStorageRW, "contract storage read and write", "contract-storage-kv", "read/write latency and root update cost", 64),
		storeV2Benchmark(StoreV2BenchmarkMessageEnqueueDequeue, "message enqueue and dequeue", "shard-prefix-proofs", "queue mutation latency and receipt root cost", 128),
		storeV2Benchmark(StoreV2BenchmarkPaymentChannelSettle, "payment channel settle", "object-store-records", "settlement latency and payment proof bytes", 8),
		storeV2Benchmark(StoreV2BenchmarkDEXPoolUpdate, "DEX pool update", "object-store-records", "pool mutation latency and root update cost", 8),
		storeV2Benchmark(StoreV2BenchmarkProofGeneration, "proof generation", "shard-prefix-proofs", "proof latency and proof size", 256),
	}
}

func BuildStoreV2UsageSpec(layouts []StoreV2LayoutDescriptor, benchmarks []StoreV2BenchmarkRequirement) (StoreV2UsageSpec, error) {
	spec := StoreV2UsageSpec{
		Version:	StoreV2UsageSpecVersion,
		Layouts:	normalizeStoreV2LayoutDescriptors(layouts),
		Benchmarks:	normalizeStoreV2BenchmarkRequirements(benchmarks),
	}
	if err := spec.ValidateFormat(); err != nil {
		return StoreV2UsageSpec{}, err
	}
	spec.Root = ComputeStoreV2UsageSpecRoot(spec.Layouts, spec.Benchmarks)
	return spec, spec.Validate()
}

func (s StoreV2UsageSpec) Normalize() StoreV2UsageSpec {
	if s.Version == 0 {
		s.Version = StoreV2UsageSpecVersion
	}
	s.Layouts = normalizeStoreV2LayoutDescriptors(s.Layouts)
	s.Benchmarks = normalizeStoreV2BenchmarkRequirements(s.Benchmarks)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s StoreV2UsageSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != StoreV2UsageSpecVersion {
		return fmt.Errorf("aetracore Store v2 usage spec version must be %d", StoreV2UsageSpecVersion)
	}
	if len(s.Layouts) == 0 {
		return errors.New("aetracore Store v2 usage spec requires layouts")
	}
	if len(s.Benchmarks) == 0 {
		return errors.New("aetracore Store v2 usage spec requires benchmarks")
	}
	layoutIDs := make(map[string]struct{}, len(s.Layouts))
	var previousLayout string
	for i, layout := range s.Layouts {
		if err := layout.Validate(); err != nil {
			return err
		}
		if _, found := layoutIDs[layout.LayoutID]; found {
			return fmt.Errorf("duplicate aetracore Store v2 layout %s", layout.LayoutID)
		}
		layoutIDs[layout.LayoutID] = struct{}{}
		if i > 0 && previousLayout >= layout.LayoutID {
			return errors.New("aetracore Store v2 layouts must be sorted canonically")
		}
		previousLayout = layout.LayoutID
	}
	seenBenchmarks := make(map[StoreV2BenchmarkID]struct{}, len(s.Benchmarks))
	var previousBenchmark StoreV2BenchmarkID
	for i, benchmark := range s.Benchmarks {
		if err := benchmark.Validate(); err != nil {
			return err
		}
		if _, found := layoutIDs[benchmark.LayoutID]; !found {
			return fmt.Errorf("aetracore Store v2 benchmark %s references missing layout %s", benchmark.BenchmarkID, benchmark.LayoutID)
		}
		if _, found := seenBenchmarks[benchmark.BenchmarkID]; found {
			return fmt.Errorf("duplicate aetracore Store v2 benchmark %s", benchmark.BenchmarkID)
		}
		seenBenchmarks[benchmark.BenchmarkID] = struct{}{}
		if i > 0 && previousBenchmark >= benchmark.BenchmarkID {
			return errors.New("aetracore Store v2 benchmarks must be sorted canonically")
		}
		previousBenchmark = benchmark.BenchmarkID
	}
	if s.Root != "" {
		if err := ValidateHash("aetracore Store v2 usage spec root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s StoreV2UsageSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("aetracore Store v2 usage spec root is required")
	}
	expected := ComputeStoreV2UsageSpecRoot(s.Layouts, s.Benchmarks)
	if s.Root != expected {
		return fmt.Errorf("aetracore Store v2 usage spec root mismatch: expected %s", expected)
	}
	return nil
}

func (l StoreV2LayoutDescriptor) Normalize() StoreV2LayoutDescriptor {
	l.LayoutID = strings.TrimSpace(l.LayoutID)
	l.StateClasses = normalizePerformanceStrings(l.StateClasses)
	l.KeyPattern = compactPerformanceText(l.KeyPattern)
	l.ProofSurface = compactPerformanceText(l.ProofSurface)
	l.DescriptorHash = normalizePerformanceHash(l.DescriptorHash)
	return l
}

func (l StoreV2LayoutDescriptor) Validate() error {
	l = l.Normalize()
	if err := validateToken("aetracore Store v2 layout id", l.LayoutID, MaxScopeLength); err != nil {
		return err
	}
	if !IsStoreV2LayoutKind(l.Kind) {
		return fmt.Errorf("unknown aetracore Store v2 layout kind %q", l.Kind)
	}
	if len(l.StateClasses) == 0 {
		return errors.New("aetracore Store v2 layout requires state classes")
	}
	if err := validateCapabilitiesForField("aetracore Store v2 state class", l.StateClasses); err != nil {
		return err
	}
	if l.KeyPattern == "" || l.ProofSurface == "" {
		return errors.New("aetracore Store v2 layout requires key pattern and proof surface")
	}
	if err := ValidateHash("aetracore Store v2 layout descriptor hash", l.DescriptorHash); err != nil {
		return err
	}
	if expected := ComputeStoreV2LayoutDescriptorHash(l); l.DescriptorHash != expected {
		return fmt.Errorf("aetracore Store v2 layout descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func (b StoreV2BenchmarkRequirement) Normalize() StoreV2BenchmarkRequirement {
	b.Operation = compactPerformanceText(b.Operation)
	b.LayoutID = strings.TrimSpace(b.LayoutID)
	b.Metric = compactPerformanceText(b.Metric)
	b.DescriptorHash = normalizePerformanceHash(b.DescriptorHash)
	return b
}

func (b StoreV2BenchmarkRequirement) Validate() error {
	b = b.Normalize()
	if !IsStoreV2BenchmarkID(b.BenchmarkID) {
		return fmt.Errorf("unknown aetracore Store v2 benchmark %q", b.BenchmarkID)
	}
	if b.Operation == "" || b.LayoutID == "" || b.Metric == "" {
		return errors.New("aetracore Store v2 benchmark requires operation, layout, and metric")
	}
	if err := validateToken("aetracore Store v2 benchmark layout id", b.LayoutID, MaxScopeLength); err != nil {
		return err
	}
	if b.Required && b.MaxRangeItems == 0 {
		return errors.New("aetracore Store v2 benchmark requires positive range bound")
	}
	if err := ValidateHash("aetracore Store v2 benchmark descriptor hash", b.DescriptorHash); err != nil {
		return err
	}
	if expected := ComputeStoreV2BenchmarkDescriptorHash(b); b.DescriptorHash != expected {
		return fmt.Errorf("aetracore Store v2 benchmark descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateStoreV2UsageCoverage() error {
	spec, err := DefaultStoreV2UsageSpec()
	if err != nil {
		return err
	}
	requiredLayouts := []StoreV2LayoutKind{
		StoreV2LayoutObjectStore,
		StoreV2LayoutKVHybrid,
		StoreV2LayoutPrefixProof,
		StoreV2LayoutCompactRoot,
		StoreV2LayoutBoundedRangeScan,
	}
	seenLayouts := make(map[StoreV2LayoutKind]struct{}, len(spec.Layouts))
	for _, layout := range spec.Layouts {
		seenLayouts[layout.Kind] = struct{}{}
	}
	for _, kind := range requiredLayouts {
		if _, found := seenLayouts[kind]; !found {
			return fmt.Errorf("aetracore Store v2 usage spec missing layout kind %s", kind)
		}
	}
	requiredBenchmarks := []StoreV2BenchmarkID{
		StoreV2BenchmarkDirectBalanceRead,
		StoreV2BenchmarkDirectIdentityResolve,
		StoreV2BenchmarkRecursiveIdentity,
		StoreV2BenchmarkContractStorageRW,
		StoreV2BenchmarkMessageEnqueueDequeue,
		StoreV2BenchmarkPaymentChannelSettle,
		StoreV2BenchmarkDEXPoolUpdate,
		StoreV2BenchmarkProofGeneration,
	}
	seenBenchmarks := make(map[StoreV2BenchmarkID]struct{}, len(spec.Benchmarks))
	for _, benchmark := range spec.Benchmarks {
		seenBenchmarks[benchmark.BenchmarkID] = struct{}{}
	}
	for _, benchmarkID := range requiredBenchmarks {
		if _, found := seenBenchmarks[benchmarkID]; !found {
			return fmt.Errorf("aetracore Store v2 usage spec missing benchmark %s", benchmarkID)
		}
	}
	return nil
}

func IsConflictReductionRuleID(ruleID ConflictReductionRuleID) bool {
	switch ruleID {
	case ConflictRuleAvoidGlobalCounters,
		ConflictRuleShardFeeAccumulator,
		ConflictRuleShardMessageQueues,
		ConflictRuleVersionedObjects,
		ConflictRuleObjectLocalLocks,
		ConflictRuleAsyncRemoteWrites:
		return true
	default:
		return false
	}
}

func IsStoreV2LayoutKind(kind StoreV2LayoutKind) bool {
	switch kind {
	case StoreV2LayoutObjectStore,
		StoreV2LayoutKVHybrid,
		StoreV2LayoutPrefixProof,
		StoreV2LayoutCompactRoot,
		StoreV2LayoutBoundedRangeScan:
		return true
	default:
		return false
	}
}

func IsStoreV2BenchmarkID(benchmarkID StoreV2BenchmarkID) bool {
	switch benchmarkID {
	case StoreV2BenchmarkDirectBalanceRead,
		StoreV2BenchmarkDirectIdentityResolve,
		StoreV2BenchmarkRecursiveIdentity,
		StoreV2BenchmarkContractStorageRW,
		StoreV2BenchmarkMessageEnqueueDequeue,
		StoreV2BenchmarkPaymentChannelSettle,
		StoreV2BenchmarkDEXPoolUpdate,
		StoreV2BenchmarkProofGeneration:
		return true
	default:
		return false
	}
}

func ComputeConflictReductionRuleHash(rule ConflictReductionRule) string {
	rule = rule.Normalize()
	return hashParts("aetra-conflict-reduction-rule-v1", string(rule.RuleID), rule.Rule, rule.Enforcement, rule.Evidence)
}

func ComputeConflictReductionSpecRoot(rules []ConflictReductionRule) string {
	ordered := normalizeConflictReductionRules(rules)
	parts := []string{"aetra-conflict-reduction-spec-v1", fmt.Sprintf("%020d", ConflictReductionSpecVersion)}
	for _, rule := range ordered {
		parts = append(parts, string(rule.RuleID), rule.DescriptorHash)
	}
	return hashParts(parts...)
}

func ComputeStoreV2LayoutDescriptorHash(layout StoreV2LayoutDescriptor) string {
	layout = layout.Normalize()
	parts := []string{"aetra-store-v2-layout-v1", layout.LayoutID, string(layout.Kind), layout.KeyPattern, layout.ProofSurface, fmt.Sprint(layout.BoundedScan)}
	parts = append(parts, layout.StateClasses...)
	return hashParts(parts...)
}

func ComputeStoreV2BenchmarkDescriptorHash(benchmark StoreV2BenchmarkRequirement) string {
	benchmark = benchmark.Normalize()
	return hashParts("aetra-store-v2-benchmark-v1", string(benchmark.BenchmarkID), benchmark.Operation, benchmark.LayoutID, benchmark.Metric, fmt.Sprint(benchmark.MaxRangeItems), fmt.Sprint(benchmark.Required))
}

func ComputeStoreV2UsageSpecRoot(layouts []StoreV2LayoutDescriptor, benchmarks []StoreV2BenchmarkRequirement) string {
	orderedLayouts := normalizeStoreV2LayoutDescriptors(layouts)
	orderedBenchmarks := normalizeStoreV2BenchmarkRequirements(benchmarks)
	parts := []string{"aetra-store-v2-usage-spec-v1", fmt.Sprintf("%020d", StoreV2UsageSpecVersion)}
	for _, layout := range orderedLayouts {
		parts = append(parts, layout.LayoutID, layout.DescriptorHash)
	}
	for _, benchmark := range orderedBenchmarks {
		parts = append(parts, string(benchmark.BenchmarkID), benchmark.DescriptorHash)
	}
	return hashParts(parts...)
}

func conflictRule(ruleID ConflictReductionRuleID, rule string, enforcement string, evidence string) ConflictReductionRule {
	out := ConflictReductionRule{RuleID: ruleID, Rule: rule, Enforcement: enforcement, Evidence: evidence}.Normalize()
	out.DescriptorHash = ComputeConflictReductionRuleHash(out)
	return out
}

func storeV2Layout(layoutID string, kind StoreV2LayoutKind, classes []string, keyPattern string, proofSurface string, boundedScan bool) StoreV2LayoutDescriptor {
	out := StoreV2LayoutDescriptor{LayoutID: layoutID, Kind: kind, StateClasses: classes, KeyPattern: keyPattern, ProofSurface: proofSurface, BoundedScan: boundedScan}.Normalize()
	out.DescriptorHash = ComputeStoreV2LayoutDescriptorHash(out)
	return out
}

func storeV2Benchmark(benchmarkID StoreV2BenchmarkID, operation string, layoutID string, metric string, maxRangeItems uint32) StoreV2BenchmarkRequirement {
	out := StoreV2BenchmarkRequirement{BenchmarkID: benchmarkID, Operation: operation, LayoutID: layoutID, Metric: metric, MaxRangeItems: maxRangeItems, Required: true}.Normalize()
	out.DescriptorHash = ComputeStoreV2BenchmarkDescriptorHash(out)
	return out
}

func normalizeConflictReductionRules(values []ConflictReductionRule) []ConflictReductionRule {
	out := make([]ConflictReductionRule, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeConflictReductionRuleHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].RuleID < out[j].RuleID
	})
	return out
}

func normalizeStoreV2LayoutDescriptors(values []StoreV2LayoutDescriptor) []StoreV2LayoutDescriptor {
	out := make([]StoreV2LayoutDescriptor, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeStoreV2LayoutDescriptorHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].LayoutID < out[j].LayoutID
	})
	return out
}

func normalizeStoreV2BenchmarkRequirements(values []StoreV2BenchmarkRequirement) []StoreV2BenchmarkRequirement {
	out := make([]StoreV2BenchmarkRequirement, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeStoreV2BenchmarkDescriptorHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].BenchmarkID < out[j].BenchmarkID
	})
	return out
}

func normalizePerformanceStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, compactPerformanceText(value))
	}
	sort.Strings(out)
	return out
}

func normalizePerformanceHash(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func compactPerformanceText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
