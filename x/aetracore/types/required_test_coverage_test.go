package types

import "testing"

func TestRequiredTestCoverageCoversSectionSeventeen(t *testing.T) {
	spec, err := DefaultRequiredTestCoverageSpec()
	if err != nil {
		t.Fatalf("default required test coverage: %v", err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("validate required test coverage: %v", err)
	}
	if err := ValidateRequiredTestCoverage(); err != nil {
		t.Fatalf("required test coverage completeness: %v", err)
	}

	counts := map[RequiredCoverageKind]int{}
	byID := map[RequiredCoverageID]RequiredTestCase{}
	for _, test := range spec.Tests {
		counts[test.Kind]++
		byID[test.TestID] = test
	}
	if counts[RequiredCoverageDeterminism] != 5 {
		t.Fatalf("expected 5 determinism tests, got %d", counts[RequiredCoverageDeterminism])
	}
	if counts[RequiredCoverageInvariant] != 8 {
		t.Fatalf("expected 8 invariant tests, got %d", counts[RequiredCoverageInvariant])
	}
	if counts[RequiredCoverageSimulation] != 8 {
		t.Fatalf("expected 8 simulation tests, got %d", counts[RequiredCoverageSimulation])
	}
	if counts[RequiredCoveragePerformance] != 8 {
		t.Fatalf("expected 8 performance tests, got %d", counts[RequiredCoveragePerformance])
	}
	if byID[RequiredCoverageSameBlockZoneRoots].Target != "block replay harness" {
		t.Fatalf("unexpected zone root determinism target: %s", byID[RequiredCoverageSameBlockZoneRoots].Target)
	}
	if byID[RequiredCoverageOutboxReceiptOrPending].Kind != RequiredCoverageInvariant {
		t.Fatal("expected outbox receipt coverage to be an invariant")
	}
	if byID[RequiredCoverageShardSplitPreservesKeys].Target != "shard split migration invariant" {
		t.Fatalf("unexpected shard split target: %s", byID[RequiredCoverageShardSplitPreservesKeys].Target)
	}
	if byID[RequiredCoverageShardMergePreservesKeys].Target != "shard merge migration invariant" {
		t.Fatalf("unexpected shard merge target: %s", byID[RequiredCoverageShardMergePreservesKeys].Target)
	}
	if byID[RequiredCoverageAdaptiveSyncActiveQueues].Target != "AdaptiveSync queue recovery simulator" {
		t.Fatalf("unexpected AdaptiveSync simulation target: %s", byID[RequiredCoverageAdaptiveSyncActiveQueues].Target)
	}
	if byID[RequiredCoverageStoreV2ProofLatency].Target != "Store v2 proof benchmark" {
		t.Fatalf("unexpected Store v2 benchmark target: %s", byID[RequiredCoverageStoreV2ProofLatency].Target)
	}
}

func TestRequiredTestCoverageRootCanonicalAndRejectsTamper(t *testing.T) {
	defaultSpec, err := DefaultRequiredTestCoverageSpec()
	if err != nil {
		t.Fatalf("default required test coverage: %v", err)
	}

	reordered := append([]RequiredTestCase{}, InvariantTestCases()...)
	reordered = append(reordered, PerformanceTestCases()...)
	reordered = append(reordered, DeterminismTestCases()...)
	reordered = append(reordered, SimulationTestCases()...)
	reorderedSpec, err := BuildRequiredTestCoverageSpec(reordered)
	if err != nil {
		t.Fatalf("reordered required test coverage: %v", err)
	}
	if reorderedSpec.Root != defaultSpec.Root {
		t.Fatalf("canonical required coverage root mismatch: %s != %s", reorderedSpec.Root, defaultSpec.Root)
	}

	if _, err := BuildRequiredTestCoverageSpec([]RequiredTestCase{defaultSpec.Tests[0], defaultSpec.Tests[0]}); err == nil {
		t.Fatal("expected duplicate required coverage item to fail")
	}

	tampered := defaultSpec
	tampered.Tests[0].DescriptorHash = hashParts("tampered required coverage test")
	if err := tampered.Validate(); err == nil {
		t.Fatal("expected tampered required coverage descriptor hash to fail")
	}
}

func TestRequiredTestCoverageRejectsWrongKindAndMissingAssertion(t *testing.T) {
	wrongKind := requiredTest(RequiredCoverageInvariant, RequiredCoverageSameVMOutput, "Same VM bytecode produces identical output.", "AVM determinism tests", "Compare outputs.")
	if err := wrongKind.Validate(); err == nil {
		t.Fatal("expected wrong kind for determinism coverage to fail")
	}

	missingAssertion := requiredTest(RequiredCoverageDeterminism, RequiredCoverageSameVMOutput, "Same VM bytecode produces identical output.", "AVM determinism tests", "Compare outputs.")
	missingAssertion.Assertion = ""
	missingAssertion.DescriptorHash = ComputeRequiredTestCaseHash(missingAssertion)
	if err := missingAssertion.Validate(); err == nil {
		t.Fatal("expected missing assertion to fail")
	}

	badHash := requiredTest(RequiredCoverageDeterminism, RequiredCoverageSameVMOutput, "Same VM bytecode produces identical output.", "AVM determinism tests", "Compare outputs.")
	badHash.DescriptorHash = hashParts("wrong required coverage hash")
	if err := badHash.Validate(); err == nil {
		t.Fatal("expected wrong descriptor hash to fail")
	}
}

func TestRequiredTestCoverageEvidenceRequiresAllCoverageGroupsToPass(t *testing.T) {
	evidence := validRequiredTestCoverageEvidence(t)
	if err := evidence.Validate(); err != nil {
		t.Fatalf("required coverage evidence should validate: %v", err)
	}

	noDeterminism := evidence
	noDeterminism.DeterminismTestsPassed = false
	noDeterminism.EvidenceHash = ComputeRequiredTestCoverageEvidenceHash(noDeterminism)
	if err := noDeterminism.Validate(); err == nil {
		t.Fatal("expected evidence without determinism pass to fail")
	}

	noInvariants := evidence
	noInvariants.InvariantTestsPassed = false
	noInvariants.EvidenceHash = ComputeRequiredTestCoverageEvidenceHash(noInvariants)
	if err := noInvariants.Validate(); err == nil {
		t.Fatal("expected evidence without invariant pass to fail")
	}

	noSimulation := evidence
	noSimulation.SimulationTestsPassed = false
	noSimulation.EvidenceHash = ComputeRequiredTestCoverageEvidenceHash(noSimulation)
	if err := noSimulation.Validate(); err == nil {
		t.Fatal("expected evidence without simulation pass to fail")
	}

	noPerformance := evidence
	noPerformance.PerformanceTestsPassed = false
	noPerformance.EvidenceHash = ComputeRequiredTestCoverageEvidenceHash(noPerformance)
	if err := noPerformance.Validate(); err == nil {
		t.Fatal("expected evidence without performance pass to fail")
	}

	tampered := evidence
	tampered.EvidenceHash = hashParts("tampered required coverage evidence")
	if err := tampered.Validate(); err == nil {
		t.Fatal("expected tampered required coverage evidence hash to fail")
	}
}

func validRequiredTestCoverageEvidence(t *testing.T) RequiredTestCoverageEvidence {
	t.Helper()
	spec, err := DefaultRequiredTestCoverageSpec()
	if err != nil {
		t.Fatalf("default required test coverage: %v", err)
	}
	evidence := RequiredTestCoverageEvidence{
		CoverageRoot:           spec.Root,
		DeterminismVectorRoot:  hashParts("determinism vectors"),
		InvariantVectorRoot:    hashParts("invariant vectors"),
		SimulationVectorRoot:   hashParts("simulation vectors"),
		PerformanceVectorRoot:  hashParts("performance vectors"),
		ReplayHarnessRoot:      hashParts("replay harness"),
		DeterminismTestsPassed: true,
		InvariantTestsPassed:   true,
		SimulationTestsPassed:  true,
		PerformanceTestsPassed: true,
	}
	evidence.EvidenceHash = ComputeRequiredTestCoverageEvidenceHash(evidence)
	return evidence
}
