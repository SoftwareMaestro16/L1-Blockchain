package types

import (
	"errors"
	"fmt"
	"sort"
)

const RequiredTestCoverageSpecVersion = uint64(1)

type RequiredCoverageKind string
type RequiredCoverageID string

const (
	RequiredCoverageDeterminism RequiredCoverageKind = "determinism"
	RequiredCoverageInvariant   RequiredCoverageKind = "invariant"
	RequiredCoverageSimulation  RequiredCoverageKind = "simulation"
	RequiredCoveragePerformance RequiredCoverageKind = "performance"

	RequiredCoverageSameBlockZoneRoots    RequiredCoverageID = "same-block-identical-zone-roots"
	RequiredCoverageSameBlockMessageRoots RequiredCoverageID = "same-block-identical-message-roots"
	RequiredCoverageSameRoutingPaths      RequiredCoverageID = "same-routing-table-identical-paths"
	RequiredCoverageSameShardIDs          RequiredCoverageID = "same-shard-layout-identical-shard-ids"
	RequiredCoverageSameVMOutput          RequiredCoverageID = "same-vm-bytecode-identical-output"

	RequiredCoverageZoneRootIncludesShardRoots RequiredCoverageID = "zone-root-includes-all-shard-roots"
	RequiredCoverageOutboxReceiptOrPending     RequiredCoverageID = "message-outbox-inclusion-receipt-or-pending"
	RequiredCoverageCrossZoneValueConservation RequiredCoverageID = "cross-zone-value-transfer-conserves-naet"
	RequiredCoveragePaymentCollateralBound     RequiredCoverageID = "payment-settlement-cannot-overpay-collateral"
	RequiredCoverageIdentityResolverProofRoot  RequiredCoverageID = "identity-resolver-proof-matches-zone-root"
	RequiredCoverageContractStateProofRoot     RequiredCoverageID = "contract-state-proof-matches-zone-root"
	RequiredCoverageShardSplitPreservesKeys    RequiredCoverageID = "shard-split-preserves-state-keys"
	RequiredCoverageShardMergePreservesKeys    RequiredCoverageID = "shard-merge-preserves-state-keys"

	RequiredCoverageHighVolumeBankTransfers  RequiredCoverageID = "high-volume-bank-transfers-across-shards"
	RequiredCoverageIdentityResolverBursts   RequiredCoverageID = "identity-resolver-update-bursts"
	RequiredCoverageContractAsyncCallChains  RequiredCoverageID = "contract-async-call-chains"
	RequiredCoveragePaymentTimeoutBounce     RequiredCoverageID = "payment-route-timeout-and-bounce"
	RequiredCoverageDEXShardConflictUpdates  RequiredCoverageID = "dex-pool-updates-under-shard-conflict"
	RequiredCoverageCrossZoneCongestion      RequiredCoverageID = "cross-zone-congestion"
	RequiredCoverageShardSplitSustainedLoad  RequiredCoverageID = "shard-split-under-sustained-load"
	RequiredCoverageAdaptiveSyncActiveQueues RequiredCoverageID = "node-recovery-adaptivesync-active-message-queues"

	RequiredCoverageLocalZoneTPS                 RequiredCoverageID = "local-zone-tps"
	RequiredCoverageCrossShardMessageThroughput  RequiredCoverageID = "cross-shard-message-throughput"
	RequiredCoverageCrossZoneMessageThroughput   RequiredCoverageID = "cross-zone-message-throughput"
	RequiredCoverageAVMInstructionThroughput     RequiredCoverageID = "avm-instruction-throughput"
	RequiredCoverageStoreV2ProofLatency          RequiredCoverageID = "store-v2-proof-generation-latency"
	RequiredCoverageBlockSTMConflictRate         RequiredCoverageID = "blockstm-conflict-rate-by-workload"
	RequiredCoverageMempoolGroupingEffectiveness RequiredCoverageID = "mempool-grouping-effectiveness"
	RequiredCoverageMultiZoneStateSyncTime       RequiredCoverageID = "state-sync-time-with-multiple-zones"
)

type RequiredTestCase struct {
	Kind           RequiredCoverageKind
	TestID         RequiredCoverageID
	Requirement    string
	Target         string
	Assertion      string
	DescriptorHash string
}

type RequiredTestCoverageSpec struct {
	Version uint64
	Tests   []RequiredTestCase
	Root    string
}

type RequiredTestCoverageEvidence struct {
	CoverageRoot           string
	DeterminismVectorRoot  string
	InvariantVectorRoot    string
	SimulationVectorRoot   string
	PerformanceVectorRoot  string
	ReplayHarnessRoot      string
	DeterminismTestsPassed bool
	InvariantTestsPassed   bool
	SimulationTestsPassed  bool
	PerformanceTestsPassed bool
	EvidenceHash           string
}

func DefaultRequiredTestCoverageSpec() (RequiredTestCoverageSpec, error) {
	return BuildRequiredTestCoverageSpec(RequiredTestCases())
}

func BuildRequiredTestCoverageSpec(tests []RequiredTestCase) (RequiredTestCoverageSpec, error) {
	spec := RequiredTestCoverageSpec{
		Version: RequiredTestCoverageSpecVersion,
		Tests:   normalizeRequiredTestCases(tests),
	}
	if err := spec.ValidateFormat(); err != nil {
		return RequiredTestCoverageSpec{}, err
	}
	spec.Root = ComputeRequiredTestCoverageRoot(spec.Tests)
	return spec, spec.Validate()
}

func RequiredTestCases() []RequiredTestCase {
	tests := make([]RequiredTestCase, 0, 29)
	tests = append(tests, DeterminismTestCases()...)
	tests = append(tests, InvariantTestCases()...)
	tests = append(tests, SimulationTestCases()...)
	tests = append(tests, PerformanceTestCases()...)
	return tests
}

func DeterminismTestCases() []RequiredTestCase {
	return []RequiredTestCase{
		requiredTest(RequiredCoverageDeterminism, RequiredCoverageSameBlockZoneRoots, "Same block produces identical zone roots across nodes.", "block replay harness", "Execute identical block inputs on independent nodes and compare every ZoneCommitment root."),
		requiredTest(RequiredCoverageDeterminism, RequiredCoverageSameBlockMessageRoots, "Same block produces identical message roots across nodes.", "message root replay harness", "Replay identical local and inbound message batches and compare GlobalMessageRoot outputs."),
		requiredTest(RequiredCoverageDeterminism, RequiredCoverageSameRoutingPaths, "Same routing table produces identical paths.", "routing table tests", "Recompute route selection from the same committed table, metrics, and tie-breaks and compare path commitments."),
		requiredTest(RequiredCoverageDeterminism, RequiredCoverageSameShardIDs, "Same shard layout produces identical shard IDs.", "shard routing tests", "Route identical zone, key, and layout epoch inputs and compare shard IDs on every node."),
		requiredTest(RequiredCoverageDeterminism, RequiredCoverageSameVMOutput, "Same VM bytecode produces identical output.", "AVM determinism tests", "Execute identical bytecode, state root, context, and message input and compare outputs, gas, events, messages, and receipt roots."),
	}
}

func InvariantTestCases() []RequiredTestCase {
	return []RequiredTestCase{
		requiredTest(RequiredCoverageInvariant, RequiredCoverageZoneRootIncludesShardRoots, "Zone root includes all shard roots.", "zone root invariant", "Recompute shard_roots_root from all active shard roots and require it to match the ZoneCommitment."),
		requiredTest(RequiredCoverageInvariant, RequiredCoverageOutboxReceiptOrPending, "Message outbox inclusion has matching receipt or pending status.", "message lifecycle invariant", "For every source outbox entry require one destination receipt, pending delivery record, or deterministic expiry state."),
		requiredTest(RequiredCoverageInvariant, RequiredCoverageCrossZoneValueConservation, "Cross-zone value transfer conserves naet.", "value conservation invariant", "Verify source escrow, destination credit, refunds, fees, and bounces sum to the original amount and fee budget."),
		requiredTest(RequiredCoverageInvariant, RequiredCoveragePaymentCollateralBound, "Payment settlement cannot overpay collateral.", "payment settlement invariant", "Verify channel and conditional settlements never exceed locked Financial Zone collateral."),
		requiredTest(RequiredCoverageInvariant, RequiredCoverageIdentityResolverProofRoot, "Identity resolver proof matches identity zone root.", "identity proof invariant", "Verify resolver and reverse lookup proofs through the committed Identity Zone root."),
		requiredTest(RequiredCoverageInvariant, RequiredCoverageContractStateProofRoot, "Contract state proof matches contract zone root.", "contract proof invariant", "Verify code, instance, storage, ABI, and event proofs through the committed Contract Zone root."),
		requiredTest(RequiredCoverageInvariant, RequiredCoverageShardSplitPreservesKeys, "Shard split preserves all state keys.", "shard split migration invariant", "Compare pre-split key manifest with deterministic post-split shard manifests and require no missing or duplicate keys."),
		requiredTest(RequiredCoverageInvariant, RequiredCoverageShardMergePreservesKeys, "Shard merge preserves all state keys.", "shard merge migration invariant", "Compare pre-merge shard manifests with merged shard manifest and require all keys and values to be preserved."),
	}
}

func SimulationTestCases() []RequiredTestCase {
	return []RequiredTestCase{
		requiredTest(RequiredCoverageSimulation, RequiredCoverageHighVolumeBankTransfers, "High-volume bank transfers across shards.", "financial shard load simulator", "Generate account-address-routed transfers across many shards and require deterministic balances, escrows, receipts, and shard roots."),
		requiredTest(RequiredCoverageSimulation, RequiredCoverageIdentityResolverBursts, "Identity resolver update bursts.", "Identity Zone burst simulator", "Apply high-volume resolver and reverse-record updates and require deterministic resolver roots, receipts, and cache invalidations."),
		requiredTest(RequiredCoverageSimulation, RequiredCoverageContractAsyncCallChains, "Contract async call chains.", "Contract Zone async simulator", "Execute chained promise, callback, timeout, and outbound message flows and require deterministic receipts and contract roots."),
		requiredTest(RequiredCoverageSimulation, RequiredCoveragePaymentTimeoutBounce, "Payment route timeout and bounce.", "payment route simulator", "Drive routed payments through timeout, refund, and bounce paths and require value conservation and proofable receipts."),
		requiredTest(RequiredCoverageSimulation, RequiredCoverageDEXShardConflictUpdates, "DEX pool updates under shard conflict.", "DEX conflict simulator", "Run contending pool updates and swaps against shard-routed pools and require deterministic conflict handling and pool roots."),
		requiredTest(RequiredCoverageSimulation, RequiredCoverageCrossZoneCongestion, "Cross-zone congestion.", "cross-zone congestion simulator", "Saturate message queues across zones and require committed congestion metrics, deterministic delivery order, expiry, and receipt roots."),
		requiredTest(RequiredCoverageSimulation, RequiredCoverageShardSplitSustainedLoad, "Shard split under sustained load.", "shard split load simulator", "Sustain gas, queue, and state-size pressure until a future split epoch is scheduled and migrated deterministically."),
		requiredTest(RequiredCoverageSimulation, RequiredCoverageAdaptiveSyncActiveQueues, "Node recovery with AdaptiveSync during active message queues.", "AdaptiveSync queue recovery simulator", "Recover a node while message queues are active and require identical zone, shard, inbox, outbox, and receipt commitments."),
	}
}

func PerformanceTestCases() []RequiredTestCase {
	return []RequiredTestCase{
		requiredTest(RequiredCoveragePerformance, RequiredCoverageLocalZoneTPS, "Local zone TPS.", "local zone benchmark", "Measure same-zone same-shard transaction throughput with deterministic root output and bounded variance reporting."),
		requiredTest(RequiredCoveragePerformance, RequiredCoverageCrossShardMessageThroughput, "Cross-shard message throughput.", "cross-shard throughput benchmark", "Measure shard-local outbox to destination inbox throughput and receipt rate under fixed committed layouts."),
		requiredTest(RequiredCoveragePerformance, RequiredCoverageCrossZoneMessageThroughput, "Cross-zone message throughput.", "cross-zone throughput benchmark", "Measure committed cross-zone message delivery, expiry, and receipt throughput under deterministic scheduling."),
		requiredTest(RequiredCoveragePerformance, RequiredCoverageAVMInstructionThroughput, "AVM instruction throughput.", "AVM runtime benchmark", "Measure instruction execution throughput by opcode class with deterministic gas and output roots."),
		requiredTest(RequiredCoveragePerformance, RequiredCoverageStoreV2ProofLatency, "Store v2 proof generation latency.", "Store v2 proof benchmark", "Measure account, resolver, contract, message, and payment proof generation latency under committed roots."),
		requiredTest(RequiredCoveragePerformance, RequiredCoverageBlockSTMConflictRate, "BlockSTM conflict rate by workload.", "BlockSTM conflict benchmark", "Measure conflict rate for bank, DEX, identity, contract, and payment workloads using canonical conflict keys."),
		requiredTest(RequiredCoveragePerformance, RequiredCoverageMempoolGroupingEffectiveness, "Mempool grouping effectiveness.", "zonemempool grouping benchmark", "Measure proposal grouping quality by zone, shard, priority, expiry, and transaction hash order."),
		requiredTest(RequiredCoveragePerformance, RequiredCoverageMultiZoneStateSyncTime, "State sync time with multiple zones.", "multi-zone state sync benchmark", "Measure AdaptiveSync recovery time for core roots, zone roots, shard roots, message roots, and proof metadata."),
	}
}

func (s RequiredTestCoverageSpec) Normalize() RequiredTestCoverageSpec {
	if s.Version == 0 {
		s.Version = RequiredTestCoverageSpecVersion
	}
	s.Tests = normalizeRequiredTestCases(s.Tests)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s RequiredTestCoverageSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != RequiredTestCoverageSpecVersion {
		return fmt.Errorf("aethercore required test coverage spec version must be %d", RequiredTestCoverageSpecVersion)
	}
	if len(s.Tests) == 0 {
		return errors.New("aethercore required test coverage spec requires tests")
	}
	seen := make(map[RequiredCoverageID]struct{}, len(s.Tests))
	var previousKind RequiredCoverageKind
	var previousID RequiredCoverageID
	for i, test := range s.Tests {
		if err := test.Validate(); err != nil {
			return err
		}
		if _, found := seen[test.TestID]; found {
			return fmt.Errorf("duplicate aethercore required test coverage item %s", test.TestID)
		}
		seen[test.TestID] = struct{}{}
		if i > 0 {
			if requiredCoverageKindRank(previousKind) > requiredCoverageKindRank(test.Kind) {
				return errors.New("aethercore required test coverage kinds must be sorted canonically")
			}
			if previousKind == test.Kind && previousID >= test.TestID {
				return errors.New("aethercore required test coverage IDs must be sorted canonically")
			}
		}
		previousKind = test.Kind
		previousID = test.TestID
	}
	if s.Root != "" {
		if err := ValidateHash("aethercore required test coverage root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s RequiredTestCoverageSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("aethercore required test coverage root is required")
	}
	expected := ComputeRequiredTestCoverageRoot(s.Tests)
	if s.Root != expected {
		return fmt.Errorf("aethercore required test coverage root mismatch: expected %s", expected)
	}
	return nil
}

func (t RequiredTestCase) Normalize() RequiredTestCase {
	t.Requirement = compactPerformanceText(t.Requirement)
	t.Target = compactPerformanceText(t.Target)
	t.Assertion = compactPerformanceText(t.Assertion)
	t.DescriptorHash = normalizePerformanceHash(t.DescriptorHash)
	return t
}

func (t RequiredTestCase) ValidateFormat() error {
	t = t.Normalize()
	if !IsRequiredCoverageKind(t.Kind) {
		return fmt.Errorf("unknown aethercore required test coverage kind %q", t.Kind)
	}
	if !IsRequiredCoverageID(t.Kind, t.TestID) {
		return fmt.Errorf("unknown aethercore required test coverage ID %q for kind %s", t.TestID, t.Kind)
	}
	if t.Requirement == "" || t.Target == "" || t.Assertion == "" {
		return errors.New("aethercore required test coverage item requires requirement, target, and assertion")
	}
	if t.DescriptorHash != "" {
		if err := ValidateHash("aethercore required test coverage descriptor hash", t.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (t RequiredTestCase) Validate() error {
	t = t.Normalize()
	if err := t.ValidateFormat(); err != nil {
		return err
	}
	if t.DescriptorHash == "" {
		return errors.New("aethercore required test coverage descriptor hash is required")
	}
	expected := ComputeRequiredTestCaseHash(t)
	if t.DescriptorHash != expected {
		return fmt.Errorf("aethercore required test coverage descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func (e RequiredTestCoverageEvidence) Normalize() RequiredTestCoverageEvidence {
	e.CoverageRoot = normalizePerformanceHash(e.CoverageRoot)
	e.DeterminismVectorRoot = normalizePerformanceHash(e.DeterminismVectorRoot)
	e.InvariantVectorRoot = normalizePerformanceHash(e.InvariantVectorRoot)
	e.SimulationVectorRoot = normalizePerformanceHash(e.SimulationVectorRoot)
	e.PerformanceVectorRoot = normalizePerformanceHash(e.PerformanceVectorRoot)
	e.ReplayHarnessRoot = normalizePerformanceHash(e.ReplayHarnessRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e RequiredTestCoverageEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name  string
		value string
	}{
		{"aethercore required test coverage root", e.CoverageRoot},
		{"aethercore required determinism vector root", e.DeterminismVectorRoot},
		{"aethercore required invariant vector root", e.InvariantVectorRoot},
		{"aethercore required simulation vector root", e.SimulationVectorRoot},
		{"aethercore required performance vector root", e.PerformanceVectorRoot},
		{"aethercore required replay harness root", e.ReplayHarnessRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.DeterminismTestsPassed {
		return errors.New("aethercore required test coverage evidence requires determinism tests to pass")
	}
	if !e.InvariantTestsPassed {
		return errors.New("aethercore required test coverage evidence requires invariant tests to pass")
	}
	if !e.SimulationTestsPassed {
		return errors.New("aethercore required test coverage evidence requires simulation tests to pass")
	}
	if !e.PerformanceTestsPassed {
		return errors.New("aethercore required test coverage evidence requires performance tests to pass")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aethercore required test coverage evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e RequiredTestCoverageEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aethercore required test coverage evidence hash is required")
	}
	expected := ComputeRequiredTestCoverageEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aethercore required test coverage evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateRequiredTestCoverage() error {
	spec, err := DefaultRequiredTestCoverageSpec()
	if err != nil {
		return err
	}
	required := map[RequiredCoverageKind][]RequiredCoverageID{
		RequiredCoverageDeterminism: {
			RequiredCoverageSameBlockZoneRoots,
			RequiredCoverageSameBlockMessageRoots,
			RequiredCoverageSameRoutingPaths,
			RequiredCoverageSameShardIDs,
			RequiredCoverageSameVMOutput,
		},
		RequiredCoverageInvariant: {
			RequiredCoverageZoneRootIncludesShardRoots,
			RequiredCoverageOutboxReceiptOrPending,
			RequiredCoverageCrossZoneValueConservation,
			RequiredCoveragePaymentCollateralBound,
			RequiredCoverageIdentityResolverProofRoot,
			RequiredCoverageContractStateProofRoot,
			RequiredCoverageShardSplitPreservesKeys,
			RequiredCoverageShardMergePreservesKeys,
		},
		RequiredCoverageSimulation: {
			RequiredCoverageHighVolumeBankTransfers,
			RequiredCoverageIdentityResolverBursts,
			RequiredCoverageContractAsyncCallChains,
			RequiredCoveragePaymentTimeoutBounce,
			RequiredCoverageDEXShardConflictUpdates,
			RequiredCoverageCrossZoneCongestion,
			RequiredCoverageShardSplitSustainedLoad,
			RequiredCoverageAdaptiveSyncActiveQueues,
		},
		RequiredCoveragePerformance: {
			RequiredCoverageLocalZoneTPS,
			RequiredCoverageCrossShardMessageThroughput,
			RequiredCoverageCrossZoneMessageThroughput,
			RequiredCoverageAVMInstructionThroughput,
			RequiredCoverageStoreV2ProofLatency,
			RequiredCoverageBlockSTMConflictRate,
			RequiredCoverageMempoolGroupingEffectiveness,
			RequiredCoverageMultiZoneStateSyncTime,
		},
	}
	seen := make(map[RequiredCoverageKind]map[RequiredCoverageID]struct{}, len(required))
	for _, test := range spec.Tests {
		if _, found := seen[test.Kind]; !found {
			seen[test.Kind] = map[RequiredCoverageID]struct{}{}
		}
		seen[test.Kind][test.TestID] = struct{}{}
	}
	for kind, ids := range required {
		for _, id := range ids {
			if _, found := seen[kind][id]; !found {
				return fmt.Errorf("aethercore required test coverage missing %s test %s", kind, id)
			}
		}
	}
	return nil
}

func ComputeRequiredTestCoverageRoot(tests []RequiredTestCase) string {
	tests = normalizeRequiredTestCases(tests)
	return hashRoot("aetheris-aek-required-test-coverage-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(tests)))
		for _, test := range tests {
			writePart(w, string(test.Kind))
			writePart(w, string(test.TestID))
			writePart(w, test.DescriptorHash)
		}
	})
}

func ComputeRequiredTestCaseHash(test RequiredTestCase) string {
	test = test.Normalize()
	return hashRoot("aetheris-aek-required-test-case-v1", func(w byteWriter) {
		writePart(w, string(test.Kind))
		writePart(w, string(test.TestID))
		writePart(w, test.Requirement)
		writePart(w, test.Target)
		writePart(w, test.Assertion)
	})
}

func ComputeRequiredTestCoverageEvidenceHash(e RequiredTestCoverageEvidence) string {
	e = e.Normalize()
	return hashRoot("aetheris-aek-required-test-coverage-evidence-v1", func(w byteWriter) {
		writePart(w, e.CoverageRoot)
		writePart(w, e.DeterminismVectorRoot)
		writePart(w, e.InvariantVectorRoot)
		writePart(w, e.SimulationVectorRoot)
		writePart(w, e.PerformanceVectorRoot)
		writePart(w, e.ReplayHarnessRoot)
		writeBoolPart(w, e.DeterminismTestsPassed)
		writeBoolPart(w, e.InvariantTestsPassed)
		writeBoolPart(w, e.SimulationTestsPassed)
		writeBoolPart(w, e.PerformanceTestsPassed)
	})
}

func writeStringParts(w byteWriter, values []string) {
	values = normalizeRoadmapStringSet(values)
	writeUint64(w, uint64(len(values)))
	for _, value := range values {
		writePart(w, value)
	}
}

func IsRequiredCoverageKind(kind RequiredCoverageKind) bool {
	return kind == RequiredCoverageDeterminism || kind == RequiredCoverageInvariant || kind == RequiredCoverageSimulation || kind == RequiredCoveragePerformance
}

func IsRequiredCoverageID(kind RequiredCoverageKind, id RequiredCoverageID) bool {
	for _, known := range requiredCoverageIDsForKind(kind) {
		if known == id {
			return true
		}
	}
	return false
}

func requiredTest(kind RequiredCoverageKind, id RequiredCoverageID, requirement, target, assertion string) RequiredTestCase {
	test := RequiredTestCase{
		Kind:        kind,
		TestID:      id,
		Requirement: requirement,
		Target:      target,
		Assertion:   assertion,
	}
	test.DescriptorHash = ComputeRequiredTestCaseHash(test)
	return test
}

func normalizeRequiredTestCases(tests []RequiredTestCase) []RequiredTestCase {
	normalized := make([]RequiredTestCase, len(tests))
	for i, test := range tests {
		normalized[i] = test.Normalize()
	}
	sort.Slice(normalized, func(i, j int) bool {
		left := normalized[i]
		right := normalized[j]
		if requiredCoverageKindRank(left.Kind) != requiredCoverageKindRank(right.Kind) {
			return requiredCoverageKindRank(left.Kind) < requiredCoverageKindRank(right.Kind)
		}
		return left.TestID < right.TestID
	})
	return normalized
}

func requiredCoverageKindRank(kind RequiredCoverageKind) int {
	switch kind {
	case RequiredCoverageDeterminism:
		return 0
	case RequiredCoverageInvariant:
		return 1
	case RequiredCoverageSimulation:
		return 2
	case RequiredCoveragePerformance:
		return 3
	default:
		return 99
	}
}

func requiredCoverageIDsForKind(kind RequiredCoverageKind) []RequiredCoverageID {
	switch kind {
	case RequiredCoverageDeterminism:
		return []RequiredCoverageID{
			RequiredCoverageSameBlockZoneRoots,
			RequiredCoverageSameBlockMessageRoots,
			RequiredCoverageSameRoutingPaths,
			RequiredCoverageSameShardIDs,
			RequiredCoverageSameVMOutput,
		}
	case RequiredCoverageInvariant:
		return []RequiredCoverageID{
			RequiredCoverageZoneRootIncludesShardRoots,
			RequiredCoverageOutboxReceiptOrPending,
			RequiredCoverageCrossZoneValueConservation,
			RequiredCoveragePaymentCollateralBound,
			RequiredCoverageIdentityResolverProofRoot,
			RequiredCoverageContractStateProofRoot,
			RequiredCoverageShardSplitPreservesKeys,
			RequiredCoverageShardMergePreservesKeys,
		}
	case RequiredCoverageSimulation:
		return []RequiredCoverageID{
			RequiredCoverageHighVolumeBankTransfers,
			RequiredCoverageIdentityResolverBursts,
			RequiredCoverageContractAsyncCallChains,
			RequiredCoveragePaymentTimeoutBounce,
			RequiredCoverageDEXShardConflictUpdates,
			RequiredCoverageCrossZoneCongestion,
			RequiredCoverageShardSplitSustainedLoad,
			RequiredCoverageAdaptiveSyncActiveQueues,
		}
	case RequiredCoveragePerformance:
		return []RequiredCoverageID{
			RequiredCoverageLocalZoneTPS,
			RequiredCoverageCrossShardMessageThroughput,
			RequiredCoverageCrossZoneMessageThroughput,
			RequiredCoverageAVMInstructionThroughput,
			RequiredCoverageStoreV2ProofLatency,
			RequiredCoverageBlockSTMConflictRate,
			RequiredCoverageMempoolGroupingEffectiveness,
			RequiredCoverageMultiZoneStateSyncTime,
		}
	default:
		return nil
	}
}
