package types

import "testing"

func TestImplementationBacklogCoverage(t *testing.T) {
	spec, err := DefaultImplementationBacklogSpec()
	if err != nil {
		t.Fatalf("default implementation backlog: %v", err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("validate implementation backlog: %v", err)
	}
	if err := ValidateImplementationBacklogCoverage(); err != nil {
		t.Fatalf("implementation backlog coverage: %v", err)
	}

	counts := map[BacklogPriority]int{}
	for _, item := range spec.Items {
		counts[item.Priority]++
	}
	if counts[BacklogPriorityHigh] != 9 {
		t.Fatalf("expected 9 high priority items, got %d", counts[BacklogPriorityHigh])
	}
	if counts[BacklogPriorityMedium] != 8 {
		t.Fatalf("expected 8 medium priority items, got %d", counts[BacklogPriorityMedium])
	}
	if counts[BacklogPriorityLower] != 6 {
		t.Fatalf("expected 6 lower priority items, got %d", counts[BacklogPriorityLower])
	}
}

func TestImplementationBacklogRootCanonicalAndRejectsTamper(t *testing.T) {
	defaultSpec, err := DefaultImplementationBacklogSpec()
	if err != nil {
		t.Fatalf("default implementation backlog: %v", err)
	}

	reorderedItems := append([]BacklogItem{}, LowerPriorityBacklogItems()...)
	reorderedItems = append(reorderedItems, HighPriorityBacklogItems()...)
	reorderedItems = append(reorderedItems, MediumPriorityBacklogItems()...)
	reordered, err := BuildImplementationBacklogSpec(reorderedItems)
	if err != nil {
		t.Fatalf("build reordered implementation backlog: %v", err)
	}
	if reordered.Root != defaultSpec.Root {
		t.Fatalf("canonical backlog root mismatch: %s != %s", reordered.Root, defaultSpec.Root)
	}

	if _, err := BuildImplementationBacklogSpec([]BacklogItem{defaultSpec.Items[0], defaultSpec.Items[0]}); err == nil {
		t.Fatal("expected duplicate backlog item to fail")
	}

	tampered := defaultSpec
	tampered.Items[0].DescriptorHash = hashParts("tampered backlog item")
	if err := tampered.Validate(); err == nil {
		t.Fatal("expected tampered backlog descriptor hash to fail")
	}
}

func TestImplementationBacklogRejectsWrongPriorityAndMissingAcceptance(t *testing.T) {
	wrongPriority := backlogItem(BacklogPriorityMedium, BacklogItemAetraCoreSkeleton, "Implement x/aetracore skeleton.", "x/aetracore module", []string{"keeper shell"})
	if err := wrongPriority.Validate(); err == nil {
		t.Fatal("expected backlog item with wrong priority to fail")
	}

	missingAcceptance := backlogItem(BacklogPriorityHigh, BacklogItemAetraCoreSkeleton, "Implement x/aetracore skeleton.", "x/aetracore module", []string{"keeper shell"})
	missingAcceptance.Acceptance = nil
	missingAcceptance.DescriptorHash = ComputeBacklogItemHash(missingAcceptance)
	if err := missingAcceptance.Validate(); err == nil {
		t.Fatal("expected backlog item without acceptance criteria to fail")
	}

	badHash := backlogItem(BacklogPriorityHigh, BacklogItemAetraCoreSkeleton, "Implement x/aetracore skeleton.", "x/aetracore module", []string{"keeper shell"})
	badHash.DescriptorHash = hashParts("wrong backlog item hash")
	if err := badHash.Validate(); err == nil {
		t.Fatal("expected backlog item with wrong hash to fail")
	}
}

func TestImplementationBacklogAcceptanceCanonicalization(t *testing.T) {
	item := BacklogItem{
		Priority:	BacklogPriorityHigh,
		ItemID:		BacklogItemAetraCoreSkeleton,
		Task:		" Implement x/aetracore skeleton. ",
		Target:		" x/aetracore module ",
		Acceptance: []string{
			" keeper shell ",
			"params",
			"keeper shell",
		},
	}
	item = item.Normalize()
	item.DescriptorHash = ComputeBacklogItemHash(item)
	if err := item.Validate(); err != nil {
		t.Fatalf("normalized backlog item should validate: %v", err)
	}
	if len(item.Acceptance) != 2 {
		t.Fatalf("expected duplicate acceptance criteria to collapse, got %d", len(item.Acceptance))
	}
	if item.Acceptance[0] != "keeper shell" || item.Acceptance[1] != "params" {
		t.Fatalf("unexpected canonical acceptance criteria: %#v", item.Acceptance)
	}
}
