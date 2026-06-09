package types

import (
	"errors"
	"fmt"
	"sort"
)

const AcceptanceCriteriaSpecVersion = uint64(1)

type AcceptanceCriterionID string

const (
	AcceptanceCoreCommitsRoots           AcceptanceCriterionID = "aether-core-commits-zone-and-message-roots"
	AcceptanceZoneAdapterExecution       AcceptanceCriterionID = "zone-executes-through-zone-adapter"
	AcceptanceMessageRoutingProofs       AcceptanceCriterionID = "messages-routing-inclusion-proofs-receipts"
	AcceptanceStoreV2ProofLayout         AcceptanceCriterionID = "store-v2-zone-shard-proof-layout"
	AcceptanceBlockSTMParallelShards     AcceptanceCriterionID = "blockstm-independent-shard-parallelism"
	AcceptanceShardSplitMergeDeterminism AcceptanceCriterionID = "shard-split-merge-deterministic-from-committed-state"
	AcceptanceAVMInstructionGasSpec      AcceptanceCriterionID = "avm-2.0-instruction-set-gas-table-specified"
	AcceptanceIdentityProofCrossZone     AcceptanceCriterionID = "identity-proof-backed-cross-zone-callable"
	AcceptancePaymentTrustlessProof      AcceptanceCriterionID = "payment-settlement-trustless-proof-verifiable"
	AcceptanceMigrationPreservesState    AcceptanceCriterionID = "migration-preserves-module-state-invariants"
)

type AcceptanceCriterion struct {
	CriterionID    AcceptanceCriterionID
	Criterion      string
	Target         string
	Evidence       string
	DescriptorHash string
}

type AcceptanceCriteriaSpec struct {
	Version  uint64
	Criteria []AcceptanceCriterion
	Root     string
}

type AcceptanceCriteriaEvidence struct {
	AcceptanceRoot          string
	CoreRootEvidence        string
	ZoneAdapterEvidence     string
	MessageEvidence         string
	StoreV2Evidence         string
	BlockSTMEvidence        string
	ShardMigrationEvidence  string
	AVMEvidence             string
	IdentityEvidence        string
	PaymentEvidence         string
	MigrationEvidence       string
	CoreRootsCommitted      bool
	ZoneAdapterExecutable   bool
	MessagesProofable       bool
	StoreV2Proofable        bool
	BlockSTMParallel        bool
	ShardRulesDeterministic bool
	AVMSpecified            bool
	IdentityProofBacked     bool
	PaymentTrustless        bool
	MigrationPreservesState bool
	EvidenceHash            string
}

func DefaultAcceptanceCriteriaSpec() (AcceptanceCriteriaSpec, error) {
	return BuildAcceptanceCriteriaSpec(AcceptanceCriteria())
}

func BuildAcceptanceCriteriaSpec(criteria []AcceptanceCriterion) (AcceptanceCriteriaSpec, error) {
	spec := AcceptanceCriteriaSpec{
		Version:  AcceptanceCriteriaSpecVersion,
		Criteria: normalizeAcceptanceCriteria(criteria),
	}
	if err := spec.ValidateFormat(); err != nil {
		return AcceptanceCriteriaSpec{}, err
	}
	spec.Root = ComputeAcceptanceCriteriaRoot(spec.Criteria)
	return spec, spec.Validate()
}

func AcceptanceCriteria() []AcceptanceCriterion {
	return []AcceptanceCriterion{
		acceptanceCriterion(AcceptanceCoreCommitsRoots, "Aether Core can commit zone roots and message roots.", "x/aethercore root aggregation", "ZoneCommitment, GlobalMessageRoot, RootSnapshot, and app hash root evidence."),
		acceptanceCriterion(AcceptanceZoneAdapterExecution, "At least one zone can execute through the zone adapter.", "x/zones adapter", "ExecuteZoneBatch, ApplyInboundMessage, ZoneExecutionSummary, export, and import evidence."),
		acceptanceCriterion(AcceptanceMessageRoutingProofs, "Messages have deterministic routing, inclusion proofs, and receipts.", "x/msgbus and routing", "Route commitment, outbox inclusion proof, destination receipt, and bounce or pending evidence."),
		acceptanceCriterion(AcceptanceStoreV2ProofLayout, "Store v2 key layout supports zone and shard proof generation.", "Store v2 prefix layout", "Core, zone, shard, message, identity, payment, and contract proof prefix evidence."),
		acceptanceCriterion(AcceptanceBlockSTMParallelShards, "BlockSTM conflict tests show independent shard execution can run in parallel.", "BlockSTM conflict tests", "Disjoint shard workload evidence, conflict profile root, and parallel execution metrics."),
		acceptanceCriterion(AcceptanceShardSplitMergeDeterminism, "Shard split and merge rules are deterministic from committed state.", "x/shards split merge scheduler", "Committed metrics, future layout epoch, migration root, and in-flight message delivery epoch evidence."),
		acceptanceCriterion(AcceptanceAVMInstructionGasSpec, "AVM 2.0 instruction set and gas table are specified.", "x/avm specification", "Bytecode format, opcode table, gas table, storage adapter, message syscall, proof syscall, and ABI registry evidence."),
		acceptanceCriterion(AcceptanceIdentityProofCrossZone, "Identity resolution is proof-backed and cross-zone callable.", "Identity Zone integration", ".aet resolver proof, reverse proof, MsgResolveIdentity, result receipt, and expiry evidence."),
		acceptanceCriterion(AcceptancePaymentTrustlessProof, "Payment settlement is trustless and proof-verifiable.", "Financial Zone payments", "Channel collateral, conditional payment, dispute, settlement proof, and payment receipt root evidence."),
		acceptanceCriterion(AcceptanceMigrationPreservesState, "Migration preserves current Aetheris module state and invariants.", "migration path evidence", "Export manifest, deterministic genesis import, legacy invariant coverage, and prefix migration evidence."),
	}
}

func (s AcceptanceCriteriaSpec) Normalize() AcceptanceCriteriaSpec {
	if s.Version == 0 {
		s.Version = AcceptanceCriteriaSpecVersion
	}
	s.Criteria = normalizeAcceptanceCriteria(s.Criteria)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s AcceptanceCriteriaSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != AcceptanceCriteriaSpecVersion {
		return fmt.Errorf("aethercore acceptance criteria spec version must be %d", AcceptanceCriteriaSpecVersion)
	}
	if len(s.Criteria) == 0 {
		return errors.New("aethercore acceptance criteria spec requires criteria")
	}
	seen := make(map[AcceptanceCriterionID]struct{}, len(s.Criteria))
	var previous AcceptanceCriterionID
	for i, criterion := range s.Criteria {
		if err := criterion.Validate(); err != nil {
			return err
		}
		if _, found := seen[criterion.CriterionID]; found {
			return fmt.Errorf("duplicate aethercore acceptance criterion %s", criterion.CriterionID)
		}
		seen[criterion.CriterionID] = struct{}{}
		if i > 0 && previous >= criterion.CriterionID {
			return errors.New("aethercore acceptance criteria must be sorted canonically")
		}
		previous = criterion.CriterionID
	}
	if s.Root != "" {
		if err := ValidateHash("aethercore acceptance criteria root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s AcceptanceCriteriaSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("aethercore acceptance criteria root is required")
	}
	expected := ComputeAcceptanceCriteriaRoot(s.Criteria)
	if s.Root != expected {
		return fmt.Errorf("aethercore acceptance criteria root mismatch: expected %s", expected)
	}
	return nil
}

func (c AcceptanceCriterion) Normalize() AcceptanceCriterion {
	c.Criterion = compactPerformanceText(c.Criterion)
	c.Target = compactPerformanceText(c.Target)
	c.Evidence = compactPerformanceText(c.Evidence)
	c.DescriptorHash = normalizePerformanceHash(c.DescriptorHash)
	return c
}

func (c AcceptanceCriterion) ValidateFormat() error {
	c = c.Normalize()
	if !IsAcceptanceCriterionID(c.CriterionID) {
		return fmt.Errorf("unknown aethercore acceptance criterion %q", c.CriterionID)
	}
	if c.Criterion == "" || c.Target == "" || c.Evidence == "" {
		return errors.New("aethercore acceptance criterion requires criterion, target, and evidence")
	}
	if c.DescriptorHash != "" {
		if err := ValidateHash("aethercore acceptance criterion descriptor hash", c.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (c AcceptanceCriterion) Validate() error {
	c = c.Normalize()
	if err := c.ValidateFormat(); err != nil {
		return err
	}
	if c.DescriptorHash == "" {
		return errors.New("aethercore acceptance criterion descriptor hash is required")
	}
	expected := ComputeAcceptanceCriterionHash(c)
	if c.DescriptorHash != expected {
		return fmt.Errorf("aethercore acceptance criterion descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func (e AcceptanceCriteriaEvidence) Normalize() AcceptanceCriteriaEvidence {
	e.AcceptanceRoot = normalizePerformanceHash(e.AcceptanceRoot)
	e.CoreRootEvidence = normalizePerformanceHash(e.CoreRootEvidence)
	e.ZoneAdapterEvidence = normalizePerformanceHash(e.ZoneAdapterEvidence)
	e.MessageEvidence = normalizePerformanceHash(e.MessageEvidence)
	e.StoreV2Evidence = normalizePerformanceHash(e.StoreV2Evidence)
	e.BlockSTMEvidence = normalizePerformanceHash(e.BlockSTMEvidence)
	e.ShardMigrationEvidence = normalizePerformanceHash(e.ShardMigrationEvidence)
	e.AVMEvidence = normalizePerformanceHash(e.AVMEvidence)
	e.IdentityEvidence = normalizePerformanceHash(e.IdentityEvidence)
	e.PaymentEvidence = normalizePerformanceHash(e.PaymentEvidence)
	e.MigrationEvidence = normalizePerformanceHash(e.MigrationEvidence)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e AcceptanceCriteriaEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name  string
		value string
	}{
		{"aethercore acceptance root", e.AcceptanceRoot},
		{"aethercore acceptance core root evidence", e.CoreRootEvidence},
		{"aethercore acceptance zone adapter evidence", e.ZoneAdapterEvidence},
		{"aethercore acceptance message evidence", e.MessageEvidence},
		{"aethercore acceptance Store v2 evidence", e.StoreV2Evidence},
		{"aethercore acceptance BlockSTM evidence", e.BlockSTMEvidence},
		{"aethercore acceptance shard migration evidence", e.ShardMigrationEvidence},
		{"aethercore acceptance AVM evidence", e.AVMEvidence},
		{"aethercore acceptance identity evidence", e.IdentityEvidence},
		{"aethercore acceptance payment evidence", e.PaymentEvidence},
		{"aethercore acceptance migration evidence", e.MigrationEvidence},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.CoreRootsCommitted {
		return errors.New("aethercore acceptance evidence requires committed zone and message roots")
	}
	if !e.ZoneAdapterExecutable {
		return errors.New("aethercore acceptance evidence requires zone adapter execution")
	}
	if !e.MessagesProofable {
		return errors.New("aethercore acceptance evidence requires deterministic messages, proofs, and receipts")
	}
	if !e.StoreV2Proofable {
		return errors.New("aethercore acceptance evidence requires Store v2 zone and shard proof layout")
	}
	if !e.BlockSTMParallel {
		return errors.New("aethercore acceptance evidence requires BlockSTM independent shard parallelism")
	}
	if !e.ShardRulesDeterministic {
		return errors.New("aethercore acceptance evidence requires deterministic shard split and merge rules")
	}
	if !e.AVMSpecified {
		return errors.New("aethercore acceptance evidence requires AVM instruction set and gas table")
	}
	if !e.IdentityProofBacked {
		return errors.New("aethercore acceptance evidence requires proof-backed cross-zone identity")
	}
	if !e.PaymentTrustless {
		return errors.New("aethercore acceptance evidence requires trustless proof-verifiable payments")
	}
	if !e.MigrationPreservesState {
		return errors.New("aethercore acceptance evidence requires migration state and invariant preservation")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aethercore acceptance criteria evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e AcceptanceCriteriaEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aethercore acceptance criteria evidence hash is required")
	}
	expected := ComputeAcceptanceCriteriaEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aethercore acceptance criteria evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateAcceptanceCriteriaCoverage() error {
	spec, err := DefaultAcceptanceCriteriaSpec()
	if err != nil {
		return err
	}
	required := []AcceptanceCriterionID{
		AcceptanceCoreCommitsRoots,
		AcceptanceZoneAdapterExecution,
		AcceptanceMessageRoutingProofs,
		AcceptanceStoreV2ProofLayout,
		AcceptanceBlockSTMParallelShards,
		AcceptanceShardSplitMergeDeterminism,
		AcceptanceAVMInstructionGasSpec,
		AcceptanceIdentityProofCrossZone,
		AcceptancePaymentTrustlessProof,
		AcceptanceMigrationPreservesState,
	}
	seen := make(map[AcceptanceCriterionID]struct{}, len(spec.Criteria))
	for _, criterion := range spec.Criteria {
		seen[criterion.CriterionID] = struct{}{}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("aethercore acceptance criteria coverage missing %s", id)
		}
	}
	return nil
}

func ComputeAcceptanceCriteriaRoot(criteria []AcceptanceCriterion) string {
	criteria = normalizeAcceptanceCriteria(criteria)
	return hashRoot("aetheris-aek-acceptance-criteria-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(criteria)))
		for _, criterion := range criteria {
			writePart(w, string(criterion.CriterionID))
			writePart(w, criterion.DescriptorHash)
		}
	})
}

func ComputeAcceptanceCriterionHash(criterion AcceptanceCriterion) string {
	criterion = criterion.Normalize()
	return hashRoot("aetheris-aek-acceptance-criterion-v1", func(w byteWriter) {
		writePart(w, string(criterion.CriterionID))
		writePart(w, criterion.Criterion)
		writePart(w, criterion.Target)
		writePart(w, criterion.Evidence)
	})
}

func ComputeAcceptanceCriteriaEvidenceHash(e AcceptanceCriteriaEvidence) string {
	e = e.Normalize()
	return hashRoot("aetheris-aek-acceptance-criteria-evidence-v1", func(w byteWriter) {
		writePart(w, e.AcceptanceRoot)
		writePart(w, e.CoreRootEvidence)
		writePart(w, e.ZoneAdapterEvidence)
		writePart(w, e.MessageEvidence)
		writePart(w, e.StoreV2Evidence)
		writePart(w, e.BlockSTMEvidence)
		writePart(w, e.ShardMigrationEvidence)
		writePart(w, e.AVMEvidence)
		writePart(w, e.IdentityEvidence)
		writePart(w, e.PaymentEvidence)
		writePart(w, e.MigrationEvidence)
		writeBoolPart(w, e.CoreRootsCommitted)
		writeBoolPart(w, e.ZoneAdapterExecutable)
		writeBoolPart(w, e.MessagesProofable)
		writeBoolPart(w, e.StoreV2Proofable)
		writeBoolPart(w, e.BlockSTMParallel)
		writeBoolPart(w, e.ShardRulesDeterministic)
		writeBoolPart(w, e.AVMSpecified)
		writeBoolPart(w, e.IdentityProofBacked)
		writeBoolPart(w, e.PaymentTrustless)
		writeBoolPart(w, e.MigrationPreservesState)
	})
}

func IsAcceptanceCriterionID(id AcceptanceCriterionID) bool {
	switch id {
	case AcceptanceCoreCommitsRoots,
		AcceptanceZoneAdapterExecution,
		AcceptanceMessageRoutingProofs,
		AcceptanceStoreV2ProofLayout,
		AcceptanceBlockSTMParallelShards,
		AcceptanceShardSplitMergeDeterminism,
		AcceptanceAVMInstructionGasSpec,
		AcceptanceIdentityProofCrossZone,
		AcceptancePaymentTrustlessProof,
		AcceptanceMigrationPreservesState:
		return true
	default:
		return false
	}
}

func acceptanceCriterion(id AcceptanceCriterionID, criterion, target, evidence string) AcceptanceCriterion {
	item := AcceptanceCriterion{
		CriterionID: id,
		Criterion:   criterion,
		Target:      target,
		Evidence:    evidence,
	}
	item.DescriptorHash = ComputeAcceptanceCriterionHash(item)
	return item
}

func normalizeAcceptanceCriteria(criteria []AcceptanceCriterion) []AcceptanceCriterion {
	normalized := make([]AcceptanceCriterion, len(criteria))
	for i, criterion := range criteria {
		normalized[i] = criterion.Normalize()
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].CriterionID < normalized[j].CriterionID
	})
	return normalized
}
