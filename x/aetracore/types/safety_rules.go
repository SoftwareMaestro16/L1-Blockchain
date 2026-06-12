package types

import (
	"errors"
	"fmt"
	"sort"
)

const SafetyRulesSpecVersion = uint64(1)

type DeterminismRuleID string
type RoutingSafetyRuleID string

const (
	DeterminismNoExternalAPIs	DeterminismRuleID	= "no-external-api-calls"
	DeterminismNoLocalClock		DeterminismRuleID	= "no-local-clock"
	DeterminismNoRandomShard	DeterminismRuleID	= "no-random-shard-placement"
	DeterminismNoMapIteration	DeterminismRuleID	= "no-nondeterministic-map-iteration"
	DeterminismNoMempoolOnlyOutput	DeterminismRuleID	= "no-mempool-only-execution-results"
	DeterminismNoFloatingPoint	DeterminismRuleID	= "no-floating-point-consensus-math"

	RoutingSafetyCommittedTable		RoutingSafetyRuleID	= "routing-table-committed"
	RoutingSafetyCommittedMetrics		RoutingSafetyRuleID	= "routing-metrics-committed"
	RoutingSafetyDeterministicTieBreak	RoutingSafetyRuleID	= "deterministic-path-tie-breaks"
	RoutingSafetyRouteFailureReceipt	RoutingSafetyRuleID	= "failed-route-receipt"
	RoutingSafetyBounceValueCap		RoutingSafetyRuleID	= "bounce-value-cap"
)

type DeterminismRule struct {
	RuleID		DeterminismRuleID
	Rule		string
	Enforcement	string
	Evidence	string
	DescriptorHash	string
}

type RoutingSafetyRule struct {
	RuleID		RoutingSafetyRuleID
	Rule		string
	Enforcement	string
	Evidence	string
	DescriptorHash	string
}

type SafetyRulesSpec struct {
	Version		uint64
	Determinism	[]DeterminismRule
	RoutingSafety	[]RoutingSafetyRule
	DeterminismRoot	string
	RoutingRoot	string
	Root		string
}

type ConsensusDeterminismEvidence struct {
	ExternalAPICalls	bool
	LocalClockUsage		bool
	RandomShardPlacement	bool
	MapIterationOutput	bool
	MempoolOnlyResult	bool
	FloatingPointMath	bool
	EvidenceHash		string
}

type RoutingSafetyEvidence struct {
	RoutingTableHash		string
	RoutingMetricsRoot		string
	DeterministicTieBreakHash	string
	FailureReceiptRoot		string
	BounceValueCapHash		string
	FailedRouteValueNAET		uint64
	ReceiptedValueNAET		uint64
	OriginalBounceValueNAET		uint64
	BounceValueNAET			uint64
	EvidenceHash			string
}

func DefaultSafetyRulesSpec() (SafetyRulesSpec, error) {
	return BuildSafetyRulesSpec(DeterminismRules(), RoutingSafetyRules())
}

func DeterminismRules() []DeterminismRule {
	return []DeterminismRule{
		determinismRule(DeterminismNoExternalAPIs, "No external API calls in consensus execution.", "Consensus execution adapters must not expose network, HTTP, RPC, or external service calls.", "module boundary rule; payload policy prohibits external APIs"),
		determinismRule(DeterminismNoLocalClock, "No local clock usage outside consensus time.", "Execution inputs use committed block height/time and reject wall-clock or process-local timestamps.", "KernelConsensusContext consensus time"),
		determinismRule(DeterminismNoRandomShard, "No random shard placement.", "Shard routing uses committed ShardLayout and deterministic route-key functions only.", "RouteKeyToShard; ShardLayout layout hash"),
		determinismRule(DeterminismNoMapIteration, "No nondeterministic map iteration in state transitions.", "State roots and schedules sort canonical slices before hashing or execution.", "normalize functions; canonical root builders"),
		determinismRule(DeterminismNoMempoolOnlyOutput, "No mempool-only data in execution results.", "Mempool ordering affects consensus only when encoded into proposal schedule and validated by deterministic bounds.", "ProposalSchedule root; MempoolSeparationPlan is node-side"),
		determinismRule(DeterminismNoFloatingPoint, "No floating point math in consensus.", "Consensus cost, fee, routing, and settlement math uses integers and fixed bps weights.", "integer fee policies; routing uint64 cost model"),
	}
}

func RoutingSafetyRules() []RoutingSafetyRule {
	return []RoutingSafetyRule{
		routingSafetyRule(RoutingSafetyCommittedTable, "Routing table is committed.", "Route selection requires a validated RoutingTableCommitment hash.", "RoutingTableCommitment.ValidateHash"),
		routingSafetyRule(RoutingSafetyCommittedMetrics, "Routing metrics are committed before use.", "Routing metrics require committed heights and metric hashes before path scoring.", "AetherRoutingMetric.CommittedHeight; ComputeAetherRoutingMetricHash"),
		routingSafetyRule(RoutingSafetyDeterministicTieBreak, "Path selection tie-breaks are deterministic.", "Candidates sort by total cost, hop count, path commitment, and destination shard ID.", "sortAetherRouteCandidates"),
		routingSafetyRule(RoutingSafetyRouteFailureReceipt, "Failed route cannot burn value without receipt.", "Failure, expiry, refund, and bounce outcomes commit receipt roots before value finalization.", "receipt root; state invariant cross-zone value conservation"),
		routingSafetyRule(RoutingSafetyBounceValueCap, "Bounce cannot create extra value.", "Bounce value and forwarding fee are capped by original escrowed value and fee budget.", "BuildAetherBounce; ValidatePaymentRouteBounceConservation"),
	}
}

func BuildSafetyRulesSpec(determinism []DeterminismRule, routing []RoutingSafetyRule) (SafetyRulesSpec, error) {
	spec := SafetyRulesSpec{
		Version:	SafetyRulesSpecVersion,
		Determinism:	normalizeDeterminismRules(determinism),
		RoutingSafety:	normalizeRoutingSafetyRules(routing),
	}
	if err := spec.ValidateFormat(); err != nil {
		return SafetyRulesSpec{}, err
	}
	spec.DeterminismRoot = ComputeDeterminismRulesRoot(spec.Determinism)
	spec.RoutingRoot = ComputeRoutingSafetyRulesRoot(spec.RoutingSafety)
	spec.Root = ComputeSafetyRulesSpecRoot(spec)
	return spec, spec.Validate()
}

func (s SafetyRulesSpec) Normalize() SafetyRulesSpec {
	if s.Version == 0 {
		s.Version = SafetyRulesSpecVersion
	}
	s.Determinism = normalizeDeterminismRules(s.Determinism)
	s.RoutingSafety = normalizeRoutingSafetyRules(s.RoutingSafety)
	s.DeterminismRoot = normalizePerformanceHash(s.DeterminismRoot)
	s.RoutingRoot = normalizePerformanceHash(s.RoutingRoot)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s SafetyRulesSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != SafetyRulesSpecVersion {
		return fmt.Errorf("aetracore safety rules spec version must be %d", SafetyRulesSpecVersion)
	}
	if len(s.Determinism) == 0 {
		return errors.New("aetracore safety rules spec requires determinism rules")
	}
	if len(s.RoutingSafety) == 0 {
		return errors.New("aetracore safety rules spec requires routing safety rules")
	}
	if err := validateDeterminismRules(s.Determinism); err != nil {
		return err
	}
	if err := validateRoutingSafetyRules(s.RoutingSafety); err != nil {
		return err
	}
	return nil
}

func (s SafetyRulesSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if err := ValidateHash("aetracore determinism safety root", s.DeterminismRoot); err != nil {
		return err
	}
	if s.DeterminismRoot != ComputeDeterminismRulesRoot(s.Determinism) {
		return errors.New("aetracore determinism safety root mismatch")
	}
	if err := ValidateHash("aetracore routing safety root", s.RoutingRoot); err != nil {
		return err
	}
	if s.RoutingRoot != ComputeRoutingSafetyRulesRoot(s.RoutingSafety) {
		return errors.New("aetracore routing safety root mismatch")
	}
	if err := ValidateHash("aetracore safety rules root", s.Root); err != nil {
		return err
	}
	if s.Root != ComputeSafetyRulesSpecRoot(s) {
		return errors.New("aetracore safety rules root mismatch")
	}
	return nil
}

func (r DeterminismRule) Normalize() DeterminismRule {
	r.Rule = compactPerformanceText(r.Rule)
	r.Enforcement = compactPerformanceText(r.Enforcement)
	r.Evidence = compactPerformanceText(r.Evidence)
	r.DescriptorHash = normalizePerformanceHash(r.DescriptorHash)
	return r
}

func (r DeterminismRule) Validate() error {
	r = r.Normalize()
	if !IsDeterminismRuleID(r.RuleID) {
		return fmt.Errorf("unknown aetracore determinism rule %q", r.RuleID)
	}
	if r.Rule == "" || r.Enforcement == "" || r.Evidence == "" {
		return errors.New("aetracore determinism rule requires rule, enforcement, and evidence")
	}
	if err := ValidateHash("aetracore determinism rule hash", r.DescriptorHash); err != nil {
		return err
	}
	if r.DescriptorHash != ComputeDeterminismRuleHash(r) {
		return errors.New("aetracore determinism rule hash mismatch")
	}
	return nil
}

func (r RoutingSafetyRule) Normalize() RoutingSafetyRule {
	r.Rule = compactPerformanceText(r.Rule)
	r.Enforcement = compactPerformanceText(r.Enforcement)
	r.Evidence = compactPerformanceText(r.Evidence)
	r.DescriptorHash = normalizePerformanceHash(r.DescriptorHash)
	return r
}

func (r RoutingSafetyRule) Validate() error {
	r = r.Normalize()
	if !IsRoutingSafetyRuleID(r.RuleID) {
		return fmt.Errorf("unknown aetracore routing safety rule %q", r.RuleID)
	}
	if r.Rule == "" || r.Enforcement == "" || r.Evidence == "" {
		return errors.New("aetracore routing safety rule requires rule, enforcement, and evidence")
	}
	if err := ValidateHash("aetracore routing safety rule hash", r.DescriptorHash); err != nil {
		return err
	}
	if r.DescriptorHash != ComputeRoutingSafetyRuleHash(r) {
		return errors.New("aetracore routing safety rule hash mismatch")
	}
	return nil
}

func ValidateConsensusDeterminismEvidence(e ConsensusDeterminismEvidence) error {
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	if e.ExternalAPICalls {
		return errors.New("aetracore determinism violation: external API calls")
	}
	if e.LocalClockUsage {
		return errors.New("aetracore determinism violation: local clock usage")
	}
	if e.RandomShardPlacement {
		return errors.New("aetracore determinism violation: random shard placement")
	}
	if e.MapIterationOutput {
		return errors.New("aetracore determinism violation: nondeterministic map iteration")
	}
	if e.MempoolOnlyResult {
		return errors.New("aetracore determinism violation: mempool-only execution result")
	}
	if e.FloatingPointMath {
		return errors.New("aetracore determinism violation: floating point consensus math")
	}
	if err := ValidateHash("aetracore determinism evidence hash", e.EvidenceHash); err != nil {
		return err
	}
	if e.EvidenceHash != ComputeConsensusDeterminismEvidenceHash(e) {
		return errors.New("aetracore determinism evidence hash mismatch")
	}
	return nil
}

func ValidateRoutingSafetyEvidence(e RoutingSafetyEvidence) error {
	e = normalizeRoutingSafetyEvidence(e)
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"aetracore routing safety table hash", e.RoutingTableHash},
		{"aetracore routing safety metrics root", e.RoutingMetricsRoot},
		{"aetracore routing safety tie-break hash", e.DeterministicTieBreakHash},
		{"aetracore routing safety failure receipt root", e.FailureReceiptRoot},
		{"aetracore routing safety bounce cap hash", e.BounceValueCapHash},
		{"aetracore routing safety evidence hash", e.EvidenceHash},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if e.FailedRouteValueNAET > 0 && e.ReceiptedValueNAET == 0 {
		return errors.New("aetracore routing safety violation: failed route value requires receipt")
	}
	if e.ReceiptedValueNAET > e.FailedRouteValueNAET {
		return errors.New("aetracore routing safety violation: receipt value exceeds failed route value")
	}
	if e.BounceValueNAET > e.OriginalBounceValueNAET {
		return errors.New("aetracore routing safety violation: bounce creates extra value")
	}
	if e.EvidenceHash != ComputeRoutingSafetyEvidenceHash(e) {
		return errors.New("aetracore routing safety evidence hash mismatch")
	}
	return nil
}

func ValidateSafetyRulesCoverage() error {
	spec, err := DefaultSafetyRulesSpec()
	if err != nil {
		return err
	}
	determinism := make(map[DeterminismRuleID]struct{}, len(spec.Determinism))
	for _, rule := range spec.Determinism {
		determinism[rule.RuleID] = struct{}{}
	}
	for _, ruleID := range []DeterminismRuleID{
		DeterminismNoExternalAPIs,
		DeterminismNoLocalClock,
		DeterminismNoRandomShard,
		DeterminismNoMapIteration,
		DeterminismNoMempoolOnlyOutput,
		DeterminismNoFloatingPoint,
	} {
		if _, found := determinism[ruleID]; !found {
			return fmt.Errorf("aetracore safety spec missing determinism rule %s", ruleID)
		}
	}
	routing := make(map[RoutingSafetyRuleID]struct{}, len(spec.RoutingSafety))
	for _, rule := range spec.RoutingSafety {
		routing[rule.RuleID] = struct{}{}
	}
	for _, ruleID := range []RoutingSafetyRuleID{
		RoutingSafetyCommittedTable,
		RoutingSafetyCommittedMetrics,
		RoutingSafetyDeterministicTieBreak,
		RoutingSafetyRouteFailureReceipt,
		RoutingSafetyBounceValueCap,
	} {
		if _, found := routing[ruleID]; !found {
			return fmt.Errorf("aetracore safety spec missing routing rule %s", ruleID)
		}
	}
	return nil
}

func IsDeterminismRuleID(ruleID DeterminismRuleID) bool {
	switch ruleID {
	case DeterminismNoExternalAPIs,
		DeterminismNoLocalClock,
		DeterminismNoRandomShard,
		DeterminismNoMapIteration,
		DeterminismNoMempoolOnlyOutput,
		DeterminismNoFloatingPoint:
		return true
	default:
		return false
	}
}

func IsRoutingSafetyRuleID(ruleID RoutingSafetyRuleID) bool {
	switch ruleID {
	case RoutingSafetyCommittedTable,
		RoutingSafetyCommittedMetrics,
		RoutingSafetyDeterministicTieBreak,
		RoutingSafetyRouteFailureReceipt,
		RoutingSafetyBounceValueCap:
		return true
	default:
		return false
	}
}

func ComputeDeterminismRuleHash(rule DeterminismRule) string {
	rule = rule.Normalize()
	return hashParts("aetra-determinism-rule-v1", string(rule.RuleID), rule.Rule, rule.Enforcement, rule.Evidence)
}

func ComputeRoutingSafetyRuleHash(rule RoutingSafetyRule) string {
	rule = rule.Normalize()
	return hashParts("aetra-routing-safety-rule-v1", string(rule.RuleID), rule.Rule, rule.Enforcement, rule.Evidence)
}

func ComputeDeterminismRulesRoot(rules []DeterminismRule) string {
	ordered := normalizeDeterminismRules(rules)
	parts := []string{"aetra-determinism-rules-root-v1", fmt.Sprintf("%020d", SafetyRulesSpecVersion)}
	for _, rule := range ordered {
		parts = append(parts, string(rule.RuleID), rule.DescriptorHash)
	}
	return hashParts(parts...)
}

func ComputeRoutingSafetyRulesRoot(rules []RoutingSafetyRule) string {
	ordered := normalizeRoutingSafetyRules(rules)
	parts := []string{"aetra-routing-safety-rules-root-v1", fmt.Sprintf("%020d", SafetyRulesSpecVersion)}
	for _, rule := range ordered {
		parts = append(parts, string(rule.RuleID), rule.DescriptorHash)
	}
	return hashParts(parts...)
}

func ComputeSafetyRulesSpecRoot(spec SafetyRulesSpec) string {
	spec = spec.Normalize()
	return hashParts("aetra-safety-rules-spec-v1", fmt.Sprintf("%020d", spec.Version), spec.DeterminismRoot, spec.RoutingRoot)
}

func ComputeConsensusDeterminismEvidenceHash(e ConsensusDeterminismEvidence) string {
	e.EvidenceHash = ""
	return hashParts(
		"aetra-determinism-evidence-v1",
		fmt.Sprint(e.ExternalAPICalls),
		fmt.Sprint(e.LocalClockUsage),
		fmt.Sprint(e.RandomShardPlacement),
		fmt.Sprint(e.MapIterationOutput),
		fmt.Sprint(e.MempoolOnlyResult),
		fmt.Sprint(e.FloatingPointMath),
	)
}

func ComputeRoutingSafetyEvidenceHash(e RoutingSafetyEvidence) string {
	e = normalizeRoutingSafetyEvidence(e)
	e.EvidenceHash = ""
	return hashParts(
		"aetra-routing-safety-evidence-v1",
		e.RoutingTableHash,
		e.RoutingMetricsRoot,
		e.DeterministicTieBreakHash,
		e.FailureReceiptRoot,
		e.BounceValueCapHash,
		fmt.Sprint(e.FailedRouteValueNAET),
		fmt.Sprint(e.ReceiptedValueNAET),
		fmt.Sprint(e.OriginalBounceValueNAET),
		fmt.Sprint(e.BounceValueNAET),
	)
}

func determinismRule(ruleID DeterminismRuleID, rule string, enforcement string, evidence string) DeterminismRule {
	out := DeterminismRule{RuleID: ruleID, Rule: rule, Enforcement: enforcement, Evidence: evidence}.Normalize()
	out.DescriptorHash = ComputeDeterminismRuleHash(out)
	return out
}

func routingSafetyRule(ruleID RoutingSafetyRuleID, rule string, enforcement string, evidence string) RoutingSafetyRule {
	out := RoutingSafetyRule{RuleID: ruleID, Rule: rule, Enforcement: enforcement, Evidence: evidence}.Normalize()
	out.DescriptorHash = ComputeRoutingSafetyRuleHash(out)
	return out
}

func normalizeDeterminismRules(values []DeterminismRule) []DeterminismRule {
	out := make([]DeterminismRule, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeDeterminismRuleHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].RuleID < out[j].RuleID
	})
	return out
}

func normalizeRoutingSafetyRules(values []RoutingSafetyRule) []RoutingSafetyRule {
	out := make([]RoutingSafetyRule, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeRoutingSafetyRuleHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].RuleID < out[j].RuleID
	})
	return out
}

func validateDeterminismRules(rules []DeterminismRule) error {
	seen := make(map[DeterminismRuleID]struct{}, len(rules))
	var previous DeterminismRuleID
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.RuleID]; found {
			return fmt.Errorf("duplicate aetracore determinism rule %s", rule.RuleID)
		}
		seen[rule.RuleID] = struct{}{}
		if i > 0 && previous >= rule.RuleID {
			return errors.New("aetracore determinism rules must be sorted canonically")
		}
		previous = rule.RuleID
	}
	return nil
}

func validateRoutingSafetyRules(rules []RoutingSafetyRule) error {
	seen := make(map[RoutingSafetyRuleID]struct{}, len(rules))
	var previous RoutingSafetyRuleID
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.RuleID]; found {
			return fmt.Errorf("duplicate aetracore routing safety rule %s", rule.RuleID)
		}
		seen[rule.RuleID] = struct{}{}
		if i > 0 && previous >= rule.RuleID {
			return errors.New("aetracore routing safety rules must be sorted canonically")
		}
		previous = rule.RuleID
	}
	return nil
}

func normalizeRoutingSafetyEvidence(e RoutingSafetyEvidence) RoutingSafetyEvidence {
	e.RoutingTableHash = normalizePerformanceHash(e.RoutingTableHash)
	e.RoutingMetricsRoot = normalizePerformanceHash(e.RoutingMetricsRoot)
	e.DeterministicTieBreakHash = normalizePerformanceHash(e.DeterministicTieBreakHash)
	e.FailureReceiptRoot = normalizePerformanceHash(e.FailureReceiptRoot)
	e.BounceValueCapHash = normalizePerformanceHash(e.BounceValueCapHash)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}
