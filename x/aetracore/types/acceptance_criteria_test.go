package types

import "testing"

func TestAcceptanceCriteriaCoversSectionEighteen(t *testing.T) {
	spec, err := DefaultAcceptanceCriteriaSpec()
	if err != nil {
		t.Fatalf("default acceptance criteria: %v", err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("validate acceptance criteria: %v", err)
	}
	if err := ValidateAcceptanceCriteriaCoverage(); err != nil {
		t.Fatalf("acceptance criteria coverage: %v", err)
	}
	if len(spec.Criteria) != 10 {
		t.Fatalf("expected 10 acceptance criteria, got %d", len(spec.Criteria))
	}
	byID := map[AcceptanceCriterionID]AcceptanceCriterion{}
	for _, criterion := range spec.Criteria {
		byID[criterion.CriterionID] = criterion
	}
	if byID[AcceptanceCoreCommitsRoots].Target != "x/aethercore root aggregation" {
		t.Fatalf("unexpected core root target: %s", byID[AcceptanceCoreCommitsRoots].Target)
	}
	if byID[AcceptanceMessageRoutingProofs].Target != "x/msgbus and routing" {
		t.Fatalf("unexpected message target: %s", byID[AcceptanceMessageRoutingProofs].Target)
	}
	if byID[AcceptanceAVMInstructionGasSpec].Target != "x/avm specification" {
		t.Fatalf("unexpected AVM target: %s", byID[AcceptanceAVMInstructionGasSpec].Target)
	}
	if byID[AcceptanceMigrationPreservesState].Target != "migration path evidence" {
		t.Fatalf("unexpected migration target: %s", byID[AcceptanceMigrationPreservesState].Target)
	}
}

func TestAcceptanceCriteriaRootCanonicalAndRejectsTamper(t *testing.T) {
	defaultSpec, err := DefaultAcceptanceCriteriaSpec()
	if err != nil {
		t.Fatalf("default acceptance criteria: %v", err)
	}

	reordered := append([]AcceptanceCriterion{}, AcceptanceCriteria()...)
	for i, j := 0, len(reordered)-1; i < j; i, j = i+1, j-1 {
		reordered[i], reordered[j] = reordered[j], reordered[i]
	}
	reorderedSpec, err := BuildAcceptanceCriteriaSpec(reordered)
	if err != nil {
		t.Fatalf("reordered acceptance criteria: %v", err)
	}
	if reorderedSpec.Root != defaultSpec.Root {
		t.Fatalf("canonical acceptance root mismatch: %s != %s", reorderedSpec.Root, defaultSpec.Root)
	}

	if _, err := BuildAcceptanceCriteriaSpec([]AcceptanceCriterion{defaultSpec.Criteria[0], defaultSpec.Criteria[0]}); err == nil {
		t.Fatal("expected duplicate acceptance criterion to fail")
	}

	tampered := defaultSpec
	tampered.Criteria[0].DescriptorHash = hashParts("tampered acceptance criterion")
	if err := tampered.Validate(); err == nil {
		t.Fatal("expected tampered acceptance criterion hash to fail")
	}
}

func TestAcceptanceCriteriaRejectsUnknownAndMissingEvidence(t *testing.T) {
	unknown := acceptanceCriterion(AcceptanceCriterionID("unknown"), "Unknown criterion.", "unknown target", "unknown evidence")
	if err := unknown.Validate(); err == nil {
		t.Fatal("expected unknown acceptance criterion to fail")
	}

	missingEvidence := acceptanceCriterion(AcceptanceCoreCommitsRoots, "Aether Core can commit zone roots and message roots.", "x/aethercore root aggregation", "root evidence")
	missingEvidence.Evidence = ""
	missingEvidence.DescriptorHash = ComputeAcceptanceCriterionHash(missingEvidence)
	if err := missingEvidence.Validate(); err == nil {
		t.Fatal("expected missing acceptance evidence to fail")
	}

	badHash := acceptanceCriterion(AcceptanceCoreCommitsRoots, "Aether Core can commit zone roots and message roots.", "x/aethercore root aggregation", "root evidence")
	badHash.DescriptorHash = hashParts("wrong acceptance criterion hash")
	if err := badHash.Validate(); err == nil {
		t.Fatal("expected wrong acceptance descriptor hash to fail")
	}
}

func TestAcceptanceCriteriaEvidenceRequiresAllReadinessGates(t *testing.T) {
	evidence := validAcceptanceCriteriaEvidence(t)
	if err := evidence.Validate(); err != nil {
		t.Fatalf("acceptance criteria evidence should validate: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*AcceptanceCriteriaEvidence)
	}{
		{name: "core roots", mutate: func(e *AcceptanceCriteriaEvidence) { e.CoreRootsCommitted = false }},
		{name: "zone adapter", mutate: func(e *AcceptanceCriteriaEvidence) { e.ZoneAdapterExecutable = false }},
		{name: "messages", mutate: func(e *AcceptanceCriteriaEvidence) { e.MessagesProofable = false }},
		{name: "Store v2", mutate: func(e *AcceptanceCriteriaEvidence) { e.StoreV2Proofable = false }},
		{name: "BlockSTM", mutate: func(e *AcceptanceCriteriaEvidence) { e.BlockSTMParallel = false }},
		{name: "shards", mutate: func(e *AcceptanceCriteriaEvidence) { e.ShardRulesDeterministic = false }},
		{name: "AVM", mutate: func(e *AcceptanceCriteriaEvidence) { e.AVMSpecified = false }},
		{name: "identity", mutate: func(e *AcceptanceCriteriaEvidence) { e.IdentityProofBacked = false }},
		{name: "payment", mutate: func(e *AcceptanceCriteriaEvidence) { e.PaymentTrustless = false }},
		{name: "migration", mutate: func(e *AcceptanceCriteriaEvidence) { e.MigrationPreservesState = false }},
	}

	for _, tc := range cases {
		broken := evidence
		tc.mutate(&broken)
		broken.EvidenceHash = ComputeAcceptanceCriteriaEvidenceHash(broken)
		if err := broken.Validate(); err == nil {
			t.Fatalf("expected missing %s readiness gate to fail", tc.name)
		}
	}

	tampered := evidence
	tampered.EvidenceHash = hashParts("tampered acceptance evidence")
	if err := tampered.Validate(); err == nil {
		t.Fatal("expected tampered acceptance evidence hash to fail")
	}
}

func validAcceptanceCriteriaEvidence(t *testing.T) AcceptanceCriteriaEvidence {
	t.Helper()
	spec, err := DefaultAcceptanceCriteriaSpec()
	if err != nil {
		t.Fatalf("default acceptance criteria: %v", err)
	}
	evidence := AcceptanceCriteriaEvidence{
		AcceptanceRoot:          spec.Root,
		CoreRootEvidence:        hashParts("core root evidence"),
		ZoneAdapterEvidence:     hashParts("zone adapter evidence"),
		MessageEvidence:         hashParts("message evidence"),
		StoreV2Evidence:         hashParts("store v2 evidence"),
		BlockSTMEvidence:        hashParts("blockstm evidence"),
		ShardMigrationEvidence:  hashParts("shard migration evidence"),
		AVMEvidence:             hashParts("avm evidence"),
		IdentityEvidence:        hashParts("identity evidence"),
		PaymentEvidence:         hashParts("payment evidence"),
		MigrationEvidence:       hashParts("migration evidence"),
		CoreRootsCommitted:      true,
		ZoneAdapterExecutable:   true,
		MessagesProofable:       true,
		StoreV2Proofable:        true,
		BlockSTMParallel:        true,
		ShardRulesDeterministic: true,
		AVMSpecified:            true,
		IdentityProofBacked:     true,
		PaymentTrustless:        true,
		MigrationPreservesState: true,
	}
	evidence.EvidenceHash = ComputeAcceptanceCriteriaEvidenceHash(evidence)
	return evidence
}
