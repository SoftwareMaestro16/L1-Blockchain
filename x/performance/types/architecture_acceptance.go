package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	AcceptanceAetraCoreRoots	= "aetra_core_commits_zone_and_message_roots"
	AcceptanceZoneAdapterExecution	= "zone_executes_through_zone_adapter"
	AcceptanceMessageProofReceipts	= "messages_have_routing_inclusion_proofs_receipts"
	AcceptanceStoreV2ZoneShardProof	= "store_v2_layout_supports_zone_shard_proofs"
	AcceptanceBlockSTMParallelism	= "blockstm_conflict_tests_parallel_shards"
	AcceptanceShardSplitMergeRules	= "shard_split_merge_deterministic_committed_state"
	AcceptanceAVM20Spec		= "avm_2_instruction_set_gas_table_specified"
	AcceptanceIdentityProofLookup	= "identity_resolution_proof_backed_cross_zone_callable"
	AcceptancePaymentSettlement	= "payment_settlement_trustless_proof_verifiable"
	AcceptanceMigrationPreservation	= "migration_preserves_module_state_invariants"
)

var requiredArchitectureAcceptanceCriteria = map[string]struct{}{
	AcceptanceAetraCoreRoots:		{},
	AcceptanceZoneAdapterExecution:		{},
	AcceptanceMessageProofReceipts:		{},
	AcceptanceStoreV2ZoneShardProof:	{},
	AcceptanceBlockSTMParallelism:		{},
	AcceptanceShardSplitMergeRules:		{},
	AcceptanceAVM20Spec:			{},
	AcceptanceIdentityProofLookup:		{},
	AcceptancePaymentSettlement:		{},
	AcceptanceMigrationPreservation:	{},
}

type ArchitectureAcceptanceCriterion struct {
	CriterionID	string
	Component	string
	EvidenceHash	string
	TestHash	string
	Ready		bool
	Deterministic	bool
	ProofBacked	bool
}

type ArchitectureAcceptanceInput struct {
	AcceptanceVersion	string
	Criteria		[]ArchitectureAcceptanceCriterion
}

type ArchitectureAcceptanceReport struct {
	AcceptanceVersion	string
	Ready			bool
	Failed			[]string
	Evidence		[]string
	ReportHash		string
}

func BuildArchitectureAcceptanceReport(input ArchitectureAcceptanceInput) ArchitectureAcceptanceReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	if err := validateExecutionToken("architecture acceptance version", input.AcceptanceVersion); err != nil {
		failed = append(failed, "acceptance_version")
	} else {
		evidence = append(evidence, "acceptance_version:"+input.AcceptanceVersion)
	}
	if err := validateArchitectureAcceptanceCriteria(input.Criteria); err != nil {
		failed = append(failed, "architecture_acceptance_criteria")
	} else {
		evidence = append(evidence, "architecture_acceptance_criteria:"+hashArchitectureAcceptanceCriteria(input.Criteria))
	}
	report := ArchitectureAcceptanceReport{
		AcceptanceVersion:	input.AcceptanceVersion,
		Ready:			len(failed) == 0,
		Failed:			normalizeStringSet(failed),
		Evidence:		normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeArchitectureAcceptanceReportHash(report)
	return report
}

func (i ArchitectureAcceptanceInput) Normalize() ArchitectureAcceptanceInput {
	i.AcceptanceVersion = strings.TrimSpace(i.AcceptanceVersion)
	for idx := range i.Criteria {
		i.Criteria[idx] = i.Criteria[idx].Normalize()
	}
	sort.SliceStable(i.Criteria, func(left, right int) bool {
		return i.Criteria[left].CriterionID < i.Criteria[right].CriterionID
	})
	return i
}

func (c ArchitectureAcceptanceCriterion) Normalize() ArchitectureAcceptanceCriterion {
	c.CriterionID = strings.TrimSpace(c.CriterionID)
	c.Component = strings.TrimSpace(c.Component)
	c.EvidenceHash = normalizeLowerHex(c.EvidenceHash)
	c.TestHash = normalizeLowerHex(c.TestHash)
	return c
}

func (c ArchitectureAcceptanceCriterion) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("architecture acceptance criterion id", check.CriterionID); err != nil {
		return err
	}
	if _, expected := requiredArchitectureAcceptanceCriteria[check.CriterionID]; !expected {
		return fmt.Errorf("architecture acceptance criterion is not required: %s", check.CriterionID)
	}
	if err := validateExecutionToken("architecture acceptance component", check.Component); err != nil {
		return err
	}
	if !check.Ready || !check.Deterministic || !check.ProofBacked {
		return errors.New("architecture acceptance criterion must be ready, deterministic, and proof-backed")
	}
	if err := validateHexHash("architecture acceptance evidence hash", check.EvidenceHash); err != nil {
		return err
	}
	return validateHexHash("architecture acceptance test hash", check.TestHash)
}

func (r ArchitectureAcceptanceReport) Validate() error {
	if err := validateExecutionToken("architecture acceptance report version", r.AcceptanceVersion); err != nil {
		return err
	}
	if r.Ready && len(r.Failed) > 0 {
		return errors.New("architecture acceptance ready report must not include failures")
	}
	if len(r.Evidence) == 0 {
		return errors.New("architecture acceptance evidence is required")
	}
	if r.ReportHash != ComputeArchitectureAcceptanceReportHash(r) {
		return errors.New("architecture acceptance report hash mismatch")
	}
	return nil
}

func validateArchitectureAcceptanceCriteria(criteria []ArchitectureAcceptanceCriterion) error {
	missing := make(map[string]struct{}, len(requiredArchitectureAcceptanceCriteria))
	for criterionID := range requiredArchitectureAcceptanceCriteria {
		missing[criterionID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(criteria))
	for _, criterion := range criteria {
		if err := criterion.Validate(); err != nil {
			return err
		}
		criterion = criterion.Normalize()
		if _, exists := seen[criterion.CriterionID]; exists {
			return fmt.Errorf("architecture acceptance duplicate criterion: %s", criterion.CriterionID)
		}
		seen[criterion.CriterionID] = struct{}{}
		delete(missing, criterion.CriterionID)
	}
	if len(missing) > 0 {
		return fmt.Errorf("architecture acceptance missing criteria: %v", sortedMapKeys(missing))
	}
	return nil
}

func ComputeArchitectureAcceptanceReportHash(report ArchitectureAcceptanceReport) string {
	failed := normalizeStringSet(report.Failed)
	evidence := normalizeStringSet(report.Evidence)
	parts := []string{"aetra-architecture-acceptance-report", strings.TrimSpace(report.AcceptanceVersion), fmt.Sprintf("%t", report.Ready)}
	parts = append(parts, failed...)
	parts = append(parts, evidence...)
	return hashStrings(parts...)
}

func hashArchitectureAcceptanceCriteria(criteria []ArchitectureAcceptanceCriterion) string {
	parts := []string{"aetra-architecture-acceptance-criteria"}
	for _, criterion := range criteria {
		criterion = criterion.Normalize()
		parts = append(parts, criterion.CriterionID, criterion.Component, criterion.EvidenceHash, criterion.TestHash, fmt.Sprintf("%t", criterion.Ready), fmt.Sprintf("%t", criterion.Deterministic), fmt.Sprintf("%t", criterion.ProofBacked))
	}
	return hashStrings(parts...)
}
