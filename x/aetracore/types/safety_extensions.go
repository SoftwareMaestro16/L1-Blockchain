package types

import (
	"errors"
	"fmt"
	"sort"
)

const ExtendedSafetyRulesSpecVersion = uint64(1)

type ShardSafetyRuleID string
type VMSafetyRuleID string
type ProofSafetyRuleID string

const (
	ShardSafetyEpochBoundary	ShardSafetyRuleID	= "layout-changes-at-epoch-boundaries"
	ShardSafetyReproducibleDecision	ShardSafetyRuleID	= "reproducible-split-merge-decisions"
	ShardSafetyDeliveryEpoch	ShardSafetyRuleID	= "in-flight-messages-include-delivery-epoch"
	ShardSafetyMigrationRoot	ShardSafetyRuleID	= "state-migration-emits-migration-root"
	ShardSafetyProofHorizon		ShardSafetyRuleID	= "old-layout-queryable-for-proof-horizon"

	VMSafetyGasMetering		VMSafetyRuleID	= "mandatory-gas-metering"
	VMSafetyBoundedIteration	VMSafetyRuleID	= "bounded-storage-iteration"
	VMSafetyMeteredProofs		VMSafetyRuleID	= "metered-proof-verification"
	VMSafetyForwardingFeeReserve	VMSafetyRuleID	= "message-forwarding-fee-reserved"
	VMSafetyNoRemoteMutation	VMSafetyRuleID	= "no-synchronous-remote-zone-mutation"
	VMSafetyDeterministicTimeout	VMSafetyRuleID	= "deterministic-promise-timeouts"

	ProofSafetyTrustedHeight	ProofSafetyRuleID	= "trusted-header-height-binding"
	ProofSafetyZoneShardIDs		ProofSafetyRuleID	= "zone-and-shard-identifiers"
	ProofSafetyObjectKeyRootType	ProofSafetyRuleID	= "object-key-and-root-type"
	ProofSafetyExplicitAbsence	ProofSafetyRuleID	= "explicit-non-existence-proofs"
	ProofSafetySupportedVersion	ProofSafetyRuleID	= "reject-unsupported-proof-versions"
)

type ShardSafetyRule struct {
	RuleID		ShardSafetyRuleID
	Rule		string
	Enforcement	string
	Evidence	string
	DescriptorHash	string
}

type VMSafetyRule struct {
	RuleID		VMSafetyRuleID
	Rule		string
	Enforcement	string
	Evidence	string
	DescriptorHash	string
}

type ProofSafetyRule struct {
	RuleID		ProofSafetyRuleID
	Rule		string
	Enforcement	string
	Evidence	string
	DescriptorHash	string
}

type ExtendedSafetyRulesSpec struct {
	Version		uint64
	ShardSafety	[]ShardSafetyRule
	VMSafety	[]VMSafetyRule
	ProofSafety	[]ProofSafetyRule
	ShardSafetyRoot	string
	VMSafetyRoot	string
	ProofSafetyRoot	string
	Root		string
}

type ShardSafetyEvidence struct {
	SourceLayoutEpoch	uint64
	TargetLayoutEpoch	uint64
	ActivationHeight	uint64
	DecisionHeight		uint64
	DecisionHash		string
	CommittedMetricsRoot	string
	DeliveryEpoch		uint64
	MigrationRoot		string
	OldLayoutHash		string
	ProofHorizonUntil	uint64
	EvidenceHash		string
}

type VMSafetyEvidence struct {
	GasTableHash			string
	GasLimit			uint64
	GasUsed				uint64
	MaxStorageIterationItems	uint32
	ProofVerificationGas		uint64
	ForwardingFeeReserved		uint64
	CreatedMessageCount		uint32
	SynchronousRemoteMutation	bool
	PromiseTimeoutHeight		uint64
	ConsensusHeight			uint64
	EvidenceHash			string
}

type ProofSafetyEvidence struct {
	ProofVersion		uint64
	TrustedHeaderHeight	uint64
	ProofHeight		uint64
	ZoneID			ZoneID
	ShardID			ShardID
	RootType		RootType
	ObjectKey		[]byte
	NonExistenceProof	bool
	AbsenceMarker		[]byte
	EvidenceHash		string
}

func DefaultExtendedSafetyRulesSpec() (ExtendedSafetyRulesSpec, error) {
	return BuildExtendedSafetyRulesSpec(ShardSafetyRules(), VMSafetyRules(), ProofSafetyRules())
}

func ShardSafetyRules() []ShardSafetyRule {
	return []ShardSafetyRule{
		shardSafetyRule(ShardSafetyEpochBoundary, "Shard layout changes only at epoch boundaries.", "Target layout epochs must be future epochs and activation heights must be after decision height.", "ShardRebalanceDecision target epoch and activation height"),
		shardSafetyRule(ShardSafetyReproducibleDecision, "Split and merge decisions are reproducible.", "Decisions bind source layout hash and committed shard metrics root before computing decision hash.", "ShardRebalanceDecision.SourceMetricsHash; DecisionHash"),
		shardSafetyRule(ShardSafetyDeliveryEpoch, "In-flight messages include delivery epoch.", "Migration tasks and shard migration receipts carry delivery epoch not earlier than target layout epoch.", "ShardMigrationTask.DeliveryEpoch"),
		shardSafetyRule(ShardSafetyMigrationRoot, "State migration emits migration root.", "Migration execution evidence must commit a migration root for proofable state movement.", "migration root hash"),
		shardSafetyRule(ShardSafetyProofHorizon, "Old shard layout remains queryable for proof horizon.", "Old layout hash remains available until the configured proof horizon height.", "old layout hash; proof horizon"),
	}
}

func VMSafetyRules() []VMSafetyRule {
	return []VMSafetyRule{
		vmSafetyRule(VMSafetyGasMetering, "Gas metering is mandatory.", "VM execution requires a gas table, positive gas limit, and gas used not exceeding the limit.", "AVMGasTable; AVMExecutionKeeper"),
		vmSafetyRule(VMSafetyBoundedIteration, "Storage iteration is bounded.", "KV range scans require a positive maximum item bound.", "KV_RANGE_BOUNDED; Store v2 range bounds"),
		vmSafetyRule(VMSafetyMeteredProofs, "Proof verification cost is metered.", "Proof verification syscalls require positive proof-verification gas.", "VERIFY_* gas metering"),
		vmSafetyRule(VMSafetyForwardingFeeReserve, "Message creation requires reserved forwarding fee.", "Message-emitting VM execution must reserve forwarding fee before MSG_SEND.", "MSG_SEND forwarding fee escrow"),
		vmSafetyRule(VMSafetyNoRemoteMutation, "Contract cannot synchronously mutate remote zone state.", "Remote mutations must be emitted as asynchronous messages.", "ContractAsyncCall; unified message layer"),
		vmSafetyRule(VMSafetyDeterministicTimeout, "Promise timeouts are deterministic.", "Promise timeout heights must be committed heights greater than current consensus height.", "PROMISE_TIMEOUT; scheduled timeout height"),
	}
}

func ProofSafetyRules() []ProofSafetyRule {
	return []ProofSafetyRule{
		proofSafetyRule(ProofSafetyTrustedHeight, "Proofs bind to trusted header height.", "Proof height must equal the trusted header height being verified.", "UniversalTrustedHeader.Height; UniversalProofEnvelope.Height"),
		proofSafetyRule(ProofSafetyZoneShardIDs, "Proofs include zone and shard identifiers.", "Zone-scoped and shard-scoped proofs must carry explicit identifiers.", "UniversalProofEnvelope.ZoneID; ShardID"),
		proofSafetyRule(ProofSafetyObjectKeyRootType, "Proofs include object key and root type.", "Every proof envelope must bind a non-empty key to a declared root type.", "UniversalProofEnvelope.Key; RootType"),
		proofSafetyRule(ProofSafetyExplicitAbsence, "Non-existence proofs are explicit.", "Absence proofs must use NonExistenceProof and a non-empty absence marker.", "ProofTypeNonExistence; AbsenceMarker"),
		proofSafetyRule(ProofSafetySupportedVersion, "Proof versions are rejected if unsupported.", "Verifier accepts only supported proof versions.", "UniversalProofVersionV1"),
	}
}

func BuildExtendedSafetyRulesSpec(shard []ShardSafetyRule, vm []VMSafetyRule, proof []ProofSafetyRule) (ExtendedSafetyRulesSpec, error) {
	spec := ExtendedSafetyRulesSpec{
		Version:	ExtendedSafetyRulesSpecVersion,
		ShardSafety:	normalizeShardSafetyRules(shard),
		VMSafety:	normalizeVMSafetyRules(vm),
		ProofSafety:	normalizeProofSafetyRules(proof),
	}
	if err := spec.ValidateFormat(); err != nil {
		return ExtendedSafetyRulesSpec{}, err
	}
	spec.ShardSafetyRoot = ComputeShardSafetyRulesRoot(spec.ShardSafety)
	spec.VMSafetyRoot = ComputeVMSafetyRulesRoot(spec.VMSafety)
	spec.ProofSafetyRoot = ComputeProofSafetyRulesRoot(spec.ProofSafety)
	spec.Root = ComputeExtendedSafetyRulesSpecRoot(spec)
	return spec, spec.Validate()
}

func (s ExtendedSafetyRulesSpec) Normalize() ExtendedSafetyRulesSpec {
	if s.Version == 0 {
		s.Version = ExtendedSafetyRulesSpecVersion
	}
	s.ShardSafety = normalizeShardSafetyRules(s.ShardSafety)
	s.VMSafety = normalizeVMSafetyRules(s.VMSafety)
	s.ProofSafety = normalizeProofSafetyRules(s.ProofSafety)
	s.ShardSafetyRoot = normalizePerformanceHash(s.ShardSafetyRoot)
	s.VMSafetyRoot = normalizePerformanceHash(s.VMSafetyRoot)
	s.ProofSafetyRoot = normalizePerformanceHash(s.ProofSafetyRoot)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s ExtendedSafetyRulesSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != ExtendedSafetyRulesSpecVersion {
		return fmt.Errorf("aetracore extended safety spec version must be %d", ExtendedSafetyRulesSpecVersion)
	}
	if len(s.ShardSafety) == 0 || len(s.VMSafety) == 0 || len(s.ProofSafety) == 0 {
		return errors.New("aetracore extended safety spec requires shard, VM, and proof rules")
	}
	if err := validateShardSafetyRules(s.ShardSafety); err != nil {
		return err
	}
	if err := validateVMSafetyRules(s.VMSafety); err != nil {
		return err
	}
	return validateProofSafetyRules(s.ProofSafety)
}

func (s ExtendedSafetyRulesSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	for _, item := range []struct {
		name		string
		value		string
		expected	string
	}{
		{"aetracore shard safety root", s.ShardSafetyRoot, ComputeShardSafetyRulesRoot(s.ShardSafety)},
		{"aetracore VM safety root", s.VMSafetyRoot, ComputeVMSafetyRulesRoot(s.VMSafety)},
		{"aetracore proof safety root", s.ProofSafetyRoot, ComputeProofSafetyRulesRoot(s.ProofSafety)},
		{"aetracore extended safety root", s.Root, ComputeExtendedSafetyRulesSpecRoot(s)},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
		if item.value != item.expected {
			return fmt.Errorf("%s mismatch", item.name)
		}
	}
	return nil
}

func ValidateShardSafetyEvidence(e ShardSafetyEvidence) error {
	e = normalizeShardSafetyEvidence(e)
	if e.SourceLayoutEpoch == 0 || e.TargetLayoutEpoch == 0 {
		return errors.New("aetracore shard safety layout epochs must be positive")
	}
	if e.TargetLayoutEpoch <= e.SourceLayoutEpoch {
		return errors.New("aetracore shard safety target epoch must be future")
	}
	if e.DecisionHeight == 0 || e.ActivationHeight == 0 || e.ActivationHeight <= e.DecisionHeight {
		return errors.New("aetracore shard safety activation height must be after decision height")
	}
	if e.DeliveryEpoch < e.TargetLayoutEpoch {
		return errors.New("aetracore shard safety delivery epoch must include target layout epoch")
	}
	if e.ProofHorizonUntil < e.ActivationHeight {
		return errors.New("aetracore shard safety old layout proof horizon is too short")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"aetracore shard safety decision hash", e.DecisionHash},
		{"aetracore shard safety committed metrics root", e.CommittedMetricsRoot},
		{"aetracore shard safety migration root", e.MigrationRoot},
		{"aetracore shard safety old layout hash", e.OldLayoutHash},
		{"aetracore shard safety evidence hash", e.EvidenceHash},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if e.EvidenceHash != ComputeShardSafetyEvidenceHash(e) {
		return errors.New("aetracore shard safety evidence hash mismatch")
	}
	return nil
}

func ValidateVMSafetyEvidence(e VMSafetyEvidence) error {
	e = normalizeVMSafetyEvidence(e)
	if err := ValidateHash("aetracore VM safety gas table hash", e.GasTableHash); err != nil {
		return err
	}
	if e.GasLimit == 0 {
		return errors.New("aetracore VM safety gas metering is required")
	}
	if e.GasUsed > e.GasLimit {
		return errors.New("aetracore VM safety gas used exceeds gas limit")
	}
	if e.MaxStorageIterationItems == 0 {
		return errors.New("aetracore VM safety storage iteration must be bounded")
	}
	if e.ProofVerificationGas == 0 {
		return errors.New("aetracore VM safety proof verification gas is required")
	}
	if e.CreatedMessageCount > 0 && e.ForwardingFeeReserved == 0 {
		return errors.New("aetracore VM safety message creation requires reserved forwarding fee")
	}
	if e.SynchronousRemoteMutation {
		return errors.New("aetracore VM safety forbids synchronous remote zone mutation")
	}
	if e.PromiseTimeoutHeight != 0 && e.PromiseTimeoutHeight <= e.ConsensusHeight {
		return errors.New("aetracore VM safety promise timeout must be a future consensus height")
	}
	if err := ValidateHash("aetracore VM safety evidence hash", e.EvidenceHash); err != nil {
		return err
	}
	if e.EvidenceHash != ComputeVMSafetyEvidenceHash(e) {
		return errors.New("aetracore VM safety evidence hash mismatch")
	}
	return nil
}

func ValidateProofSafetyEvidence(e ProofSafetyEvidence) error {
	e = normalizeProofSafetyEvidence(e)
	if e.ProofVersion != UniversalProofVersionV1 {
		return errors.New("aetracore proof safety unsupported proof version")
	}
	if e.TrustedHeaderHeight == 0 || e.ProofHeight == 0 || e.ProofHeight != e.TrustedHeaderHeight {
		return errors.New("aetracore proof safety proof height must bind to trusted header height")
	}
	if err := ValidateZoneID(e.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(e.ShardID); err != nil {
		return err
	}
	if e.RootType == "" {
		return errors.New("aetracore proof safety root type is required")
	}
	if len(e.ObjectKey) == 0 {
		return errors.New("aetracore proof safety object key is required")
	}
	if e.NonExistenceProof && len(e.AbsenceMarker) == 0 {
		return errors.New("aetracore proof safety non-existence proof requires explicit marker")
	}
	if err := ValidateHash("aetracore proof safety evidence hash", e.EvidenceHash); err != nil {
		return err
	}
	if e.EvidenceHash != ComputeProofSafetyEvidenceHash(e) {
		return errors.New("aetracore proof safety evidence hash mismatch")
	}
	return nil
}

func ValidateExtendedSafetyRulesCoverage() error {
	spec, err := DefaultExtendedSafetyRulesSpec()
	if err != nil {
		return err
	}
	if err := requireShardSafetyCoverage(spec.ShardSafety); err != nil {
		return err
	}
	if err := requireVMSafetyCoverage(spec.VMSafety); err != nil {
		return err
	}
	return requireProofSafetyCoverage(spec.ProofSafety)
}

func IsShardSafetyRuleID(ruleID ShardSafetyRuleID) bool {
	switch ruleID {
	case ShardSafetyEpochBoundary, ShardSafetyReproducibleDecision, ShardSafetyDeliveryEpoch, ShardSafetyMigrationRoot, ShardSafetyProofHorizon:
		return true
	default:
		return false
	}
}

func IsVMSafetyRuleID(ruleID VMSafetyRuleID) bool {
	switch ruleID {
	case VMSafetyGasMetering, VMSafetyBoundedIteration, VMSafetyMeteredProofs, VMSafetyForwardingFeeReserve, VMSafetyNoRemoteMutation, VMSafetyDeterministicTimeout:
		return true
	default:
		return false
	}
}

func IsProofSafetyRuleID(ruleID ProofSafetyRuleID) bool {
	switch ruleID {
	case ProofSafetyTrustedHeight, ProofSafetyZoneShardIDs, ProofSafetyObjectKeyRootType, ProofSafetyExplicitAbsence, ProofSafetySupportedVersion:
		return true
	default:
		return false
	}
}

func (r ShardSafetyRule) Normalize() ShardSafetyRule {
	r.Rule = compactPerformanceText(r.Rule)
	r.Enforcement = compactPerformanceText(r.Enforcement)
	r.Evidence = compactPerformanceText(r.Evidence)
	r.DescriptorHash = normalizePerformanceHash(r.DescriptorHash)
	return r
}

func (r ShardSafetyRule) Validate() error {
	r = r.Normalize()
	if !IsShardSafetyRuleID(r.RuleID) {
		return fmt.Errorf("unknown aetracore shard safety rule %q", r.RuleID)
	}
	if r.Rule == "" || r.Enforcement == "" || r.Evidence == "" {
		return errors.New("aetracore shard safety rule requires rule, enforcement, and evidence")
	}
	if err := ValidateHash("aetracore shard safety rule hash", r.DescriptorHash); err != nil {
		return err
	}
	if r.DescriptorHash != ComputeShardSafetyRuleHash(r) {
		return errors.New("aetracore shard safety rule hash mismatch")
	}
	return nil
}

func (r VMSafetyRule) Normalize() VMSafetyRule {
	r.Rule = compactPerformanceText(r.Rule)
	r.Enforcement = compactPerformanceText(r.Enforcement)
	r.Evidence = compactPerformanceText(r.Evidence)
	r.DescriptorHash = normalizePerformanceHash(r.DescriptorHash)
	return r
}

func (r VMSafetyRule) Validate() error {
	r = r.Normalize()
	if !IsVMSafetyRuleID(r.RuleID) {
		return fmt.Errorf("unknown aetracore VM safety rule %q", r.RuleID)
	}
	if r.Rule == "" || r.Enforcement == "" || r.Evidence == "" {
		return errors.New("aetracore VM safety rule requires rule, enforcement, and evidence")
	}
	if err := ValidateHash("aetracore VM safety rule hash", r.DescriptorHash); err != nil {
		return err
	}
	if r.DescriptorHash != ComputeVMSafetyRuleHash(r) {
		return errors.New("aetracore VM safety rule hash mismatch")
	}
	return nil
}

func (r ProofSafetyRule) Normalize() ProofSafetyRule {
	r.Rule = compactPerformanceText(r.Rule)
	r.Enforcement = compactPerformanceText(r.Enforcement)
	r.Evidence = compactPerformanceText(r.Evidence)
	r.DescriptorHash = normalizePerformanceHash(r.DescriptorHash)
	return r
}

func (r ProofSafetyRule) Validate() error {
	r = r.Normalize()
	if !IsProofSafetyRuleID(r.RuleID) {
		return fmt.Errorf("unknown aetracore proof safety rule %q", r.RuleID)
	}
	if r.Rule == "" || r.Enforcement == "" || r.Evidence == "" {
		return errors.New("aetracore proof safety rule requires rule, enforcement, and evidence")
	}
	if err := ValidateHash("aetracore proof safety rule hash", r.DescriptorHash); err != nil {
		return err
	}
	if r.DescriptorHash != ComputeProofSafetyRuleHash(r) {
		return errors.New("aetracore proof safety rule hash mismatch")
	}
	return nil
}

func ComputeShardSafetyRuleHash(rule ShardSafetyRule) string {
	rule = rule.Normalize()
	return hashParts("aetra-shard-safety-rule-v1", string(rule.RuleID), rule.Rule, rule.Enforcement, rule.Evidence)
}

func ComputeVMSafetyRuleHash(rule VMSafetyRule) string {
	rule = rule.Normalize()
	return hashParts("aetra-vm-safety-rule-v1", string(rule.RuleID), rule.Rule, rule.Enforcement, rule.Evidence)
}

func ComputeProofSafetyRuleHash(rule ProofSafetyRule) string {
	rule = rule.Normalize()
	return hashParts("aetra-proof-safety-rule-v1", string(rule.RuleID), rule.Rule, rule.Enforcement, rule.Evidence)
}

func ComputeShardSafetyRulesRoot(rules []ShardSafetyRule) string {
	ordered := normalizeShardSafetyRules(rules)
	parts := []string{"aetra-shard-safety-rules-root-v1", fmt.Sprintf("%020d", ExtendedSafetyRulesSpecVersion)}
	for _, rule := range ordered {
		parts = append(parts, string(rule.RuleID), rule.DescriptorHash)
	}
	return hashParts(parts...)
}

func ComputeVMSafetyRulesRoot(rules []VMSafetyRule) string {
	ordered := normalizeVMSafetyRules(rules)
	parts := []string{"aetra-vm-safety-rules-root-v1", fmt.Sprintf("%020d", ExtendedSafetyRulesSpecVersion)}
	for _, rule := range ordered {
		parts = append(parts, string(rule.RuleID), rule.DescriptorHash)
	}
	return hashParts(parts...)
}

func ComputeProofSafetyRulesRoot(rules []ProofSafetyRule) string {
	ordered := normalizeProofSafetyRules(rules)
	parts := []string{"aetra-proof-safety-rules-root-v1", fmt.Sprintf("%020d", ExtendedSafetyRulesSpecVersion)}
	for _, rule := range ordered {
		parts = append(parts, string(rule.RuleID), rule.DescriptorHash)
	}
	return hashParts(parts...)
}

func ComputeExtendedSafetyRulesSpecRoot(spec ExtendedSafetyRulesSpec) string {
	spec = spec.Normalize()
	return hashParts("aetra-extended-safety-rules-spec-v1", fmt.Sprintf("%020d", spec.Version), spec.ShardSafetyRoot, spec.VMSafetyRoot, spec.ProofSafetyRoot)
}

func ComputeShardSafetyEvidenceHash(e ShardSafetyEvidence) string {
	e = normalizeShardSafetyEvidence(e)
	e.EvidenceHash = ""
	return hashParts("aetra-shard-safety-evidence-v1", fmt.Sprint(e.SourceLayoutEpoch), fmt.Sprint(e.TargetLayoutEpoch), fmt.Sprint(e.ActivationHeight), fmt.Sprint(e.DecisionHeight), e.DecisionHash, e.CommittedMetricsRoot, fmt.Sprint(e.DeliveryEpoch), e.MigrationRoot, e.OldLayoutHash, fmt.Sprint(e.ProofHorizonUntil))
}

func ComputeVMSafetyEvidenceHash(e VMSafetyEvidence) string {
	e = normalizeVMSafetyEvidence(e)
	e.EvidenceHash = ""
	return hashParts("aetra-vm-safety-evidence-v1", e.GasTableHash, fmt.Sprint(e.GasLimit), fmt.Sprint(e.GasUsed), fmt.Sprint(e.MaxStorageIterationItems), fmt.Sprint(e.ProofVerificationGas), fmt.Sprint(e.ForwardingFeeReserved), fmt.Sprint(e.CreatedMessageCount), fmt.Sprint(e.SynchronousRemoteMutation), fmt.Sprint(e.PromiseTimeoutHeight), fmt.Sprint(e.ConsensusHeight))
}

func ComputeProofSafetyEvidenceHash(e ProofSafetyEvidence) string {
	e = normalizeProofSafetyEvidence(e)
	e.EvidenceHash = ""
	return hashParts("aetra-proof-safety-evidence-v1", fmt.Sprint(e.ProofVersion), fmt.Sprint(e.TrustedHeaderHeight), fmt.Sprint(e.ProofHeight), string(e.ZoneID), string(e.ShardID), string(e.RootType), string(e.ObjectKey), fmt.Sprint(e.NonExistenceProof), string(e.AbsenceMarker))
}

func shardSafetyRule(ruleID ShardSafetyRuleID, rule string, enforcement string, evidence string) ShardSafetyRule {
	out := ShardSafetyRule{RuleID: ruleID, Rule: rule, Enforcement: enforcement, Evidence: evidence}.Normalize()
	out.DescriptorHash = ComputeShardSafetyRuleHash(out)
	return out
}

func vmSafetyRule(ruleID VMSafetyRuleID, rule string, enforcement string, evidence string) VMSafetyRule {
	out := VMSafetyRule{RuleID: ruleID, Rule: rule, Enforcement: enforcement, Evidence: evidence}.Normalize()
	out.DescriptorHash = ComputeVMSafetyRuleHash(out)
	return out
}

func proofSafetyRule(ruleID ProofSafetyRuleID, rule string, enforcement string, evidence string) ProofSafetyRule {
	out := ProofSafetyRule{RuleID: ruleID, Rule: rule, Enforcement: enforcement, Evidence: evidence}.Normalize()
	out.DescriptorHash = ComputeProofSafetyRuleHash(out)
	return out
}

func normalizeShardSafetyRules(values []ShardSafetyRule) []ShardSafetyRule {
	out := make([]ShardSafetyRule, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeShardSafetyRuleHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out
}

func normalizeVMSafetyRules(values []VMSafetyRule) []VMSafetyRule {
	out := make([]VMSafetyRule, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeVMSafetyRuleHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out
}

func normalizeProofSafetyRules(values []ProofSafetyRule) []ProofSafetyRule {
	out := make([]ProofSafetyRule, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeProofSafetyRuleHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out
}

func validateShardSafetyRules(rules []ShardSafetyRule) error {
	seen := make(map[ShardSafetyRuleID]struct{}, len(rules))
	var previous ShardSafetyRuleID
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.RuleID]; found {
			return fmt.Errorf("duplicate aetracore shard safety rule %s", rule.RuleID)
		}
		seen[rule.RuleID] = struct{}{}
		if i > 0 && previous >= rule.RuleID {
			return errors.New("aetracore shard safety rules must be sorted canonically")
		}
		previous = rule.RuleID
	}
	return nil
}

func validateVMSafetyRules(rules []VMSafetyRule) error {
	seen := make(map[VMSafetyRuleID]struct{}, len(rules))
	var previous VMSafetyRuleID
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.RuleID]; found {
			return fmt.Errorf("duplicate aetracore VM safety rule %s", rule.RuleID)
		}
		seen[rule.RuleID] = struct{}{}
		if i > 0 && previous >= rule.RuleID {
			return errors.New("aetracore VM safety rules must be sorted canonically")
		}
		previous = rule.RuleID
	}
	return nil
}

func validateProofSafetyRules(rules []ProofSafetyRule) error {
	seen := make(map[ProofSafetyRuleID]struct{}, len(rules))
	var previous ProofSafetyRuleID
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.RuleID]; found {
			return fmt.Errorf("duplicate aetracore proof safety rule %s", rule.RuleID)
		}
		seen[rule.RuleID] = struct{}{}
		if i > 0 && previous >= rule.RuleID {
			return errors.New("aetracore proof safety rules must be sorted canonically")
		}
		previous = rule.RuleID
	}
	return nil
}

func requireShardSafetyCoverage(rules []ShardSafetyRule) error {
	seen := make(map[ShardSafetyRuleID]struct{}, len(rules))
	for _, rule := range rules {
		seen[rule.RuleID] = struct{}{}
	}
	for _, ruleID := range []ShardSafetyRuleID{ShardSafetyEpochBoundary, ShardSafetyReproducibleDecision, ShardSafetyDeliveryEpoch, ShardSafetyMigrationRoot, ShardSafetyProofHorizon} {
		if _, found := seen[ruleID]; !found {
			return fmt.Errorf("aetracore extended safety missing shard rule %s", ruleID)
		}
	}
	return nil
}

func requireVMSafetyCoverage(rules []VMSafetyRule) error {
	seen := make(map[VMSafetyRuleID]struct{}, len(rules))
	for _, rule := range rules {
		seen[rule.RuleID] = struct{}{}
	}
	for _, ruleID := range []VMSafetyRuleID{VMSafetyGasMetering, VMSafetyBoundedIteration, VMSafetyMeteredProofs, VMSafetyForwardingFeeReserve, VMSafetyNoRemoteMutation, VMSafetyDeterministicTimeout} {
		if _, found := seen[ruleID]; !found {
			return fmt.Errorf("aetracore extended safety missing VM rule %s", ruleID)
		}
	}
	return nil
}

func requireProofSafetyCoverage(rules []ProofSafetyRule) error {
	seen := make(map[ProofSafetyRuleID]struct{}, len(rules))
	for _, rule := range rules {
		seen[rule.RuleID] = struct{}{}
	}
	for _, ruleID := range []ProofSafetyRuleID{ProofSafetyTrustedHeight, ProofSafetyZoneShardIDs, ProofSafetyObjectKeyRootType, ProofSafetyExplicitAbsence, ProofSafetySupportedVersion} {
		if _, found := seen[ruleID]; !found {
			return fmt.Errorf("aetracore extended safety missing proof rule %s", ruleID)
		}
	}
	return nil
}

func normalizeShardSafetyEvidence(e ShardSafetyEvidence) ShardSafetyEvidence {
	e.DecisionHash = normalizePerformanceHash(e.DecisionHash)
	e.CommittedMetricsRoot = normalizePerformanceHash(e.CommittedMetricsRoot)
	e.MigrationRoot = normalizePerformanceHash(e.MigrationRoot)
	e.OldLayoutHash = normalizePerformanceHash(e.OldLayoutHash)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func normalizeVMSafetyEvidence(e VMSafetyEvidence) VMSafetyEvidence {
	e.GasTableHash = normalizePerformanceHash(e.GasTableHash)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func normalizeProofSafetyEvidence(e ProofSafetyEvidence) ProofSafetyEvidence {
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	e.ObjectKey = append([]byte(nil), e.ObjectKey...)
	e.AbsenceMarker = append([]byte(nil), e.AbsenceMarker...)
	return e
}
