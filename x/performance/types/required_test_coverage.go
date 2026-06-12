package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type RequiredCoverageClass string

const (
	RequiredCoverageDeterminism	RequiredCoverageClass	= "DETERMINISM"
	RequiredCoverageInvariant	RequiredCoverageClass	= "INVARIANT"
	RequiredCoverageSimulation	RequiredCoverageClass	= "SIMULATION"
	RequiredCoveragePerformance	RequiredCoverageClass	= "PERFORMANCE"
)

const (
	RequiredDeterminismZoneRoots	= "same_block_identical_zone_roots"
	RequiredDeterminismMessageRoots	= "same_block_identical_message_roots"
	RequiredDeterminismRoutingPaths	= "same_routing_table_identical_paths"
	RequiredDeterminismShardIDs	= "same_shard_layout_identical_shard_ids"
	RequiredDeterminismVMOutput	= "same_vm_bytecode_identical_output"

	RequiredInvariantZoneRootIncludesShardRoots	= "zone_root_includes_all_shard_roots"
	RequiredInvariantOutboxReceiptOrPending		= "message_outbox_inclusion_receipt_or_pending"
	RequiredInvariantCrossZoneValueConservation	= "cross_zone_value_transfer_conserves_naet"
	RequiredInvariantPaymentCollateralOverpay	= "payment_settlement_cannot_overpay_collateral"
	RequiredInvariantIdentityProofRootMatch		= "identity_resolver_proof_matches_identity_zone_root"
	RequiredInvariantContractProofRootMatch		= "contract_state_proof_matches_contract_zone_root"
	RequiredInvariantShardSplitPreservesKeys	= "shard_split_preserves_all_state_keys"
	RequiredInvariantShardMergePreservesKeys	= "shard_merge_preserves_all_state_keys"

	RequiredSimulationHighVolumeBankTransfers	= "high_volume_bank_transfers_across_shards"
	RequiredSimulationIdentityUpdateBursts		= "identity_resolver_update_bursts"
	RequiredSimulationContractAsyncChains		= "contract_async_call_chains"
	RequiredSimulationPaymentTimeoutBounce		= "payment_route_timeout_and_bounce"
	RequiredSimulationCrossZoneCongestion		= "cross_zone_congestion"
	RequiredSimulationShardSplitLoad		= "shard_split_under_sustained_load"
	RequiredSimulationAdaptiveSyncQueues		= "node_recovery_adaptivesync_active_message_queues"

	RequiredPerformanceLocalZoneTPS			= "local_zone_tps"
	RequiredPerformanceCrossShardThroughput		= "cross_shard_message_throughput"
	RequiredPerformanceCrossZoneThroughput		= "cross_zone_message_throughput"
	RequiredPerformanceAVMInstructionRate		= "avm_instruction_throughput"
	RequiredPerformanceStoreV2ProofLatency		= "store_v2_proof_generation_latency"
	RequiredPerformanceBlockSTMConflictRate		= "blockstm_conflict_rate_by_workload"
	RequiredPerformanceMempoolGrouping		= "mempool_grouping_effectiveness"
	RequiredPerformanceStateSyncMultipleZones	= "state_sync_time_with_multiple_zones"
)

var requiredDeterminismCoverage = map[string]struct{}{
	RequiredDeterminismZoneRoots:		{},
	RequiredDeterminismMessageRoots:	{},
	RequiredDeterminismRoutingPaths:	{},
	RequiredDeterminismShardIDs:		{},
	RequiredDeterminismVMOutput:		{},
}

var requiredInvariantCoverage = map[string]struct{}{
	RequiredInvariantZoneRootIncludesShardRoots:	{},
	RequiredInvariantOutboxReceiptOrPending:	{},
	RequiredInvariantCrossZoneValueConservation:	{},
	RequiredInvariantPaymentCollateralOverpay:	{},
	RequiredInvariantIdentityProofRootMatch:	{},
	RequiredInvariantContractProofRootMatch:	{},
	RequiredInvariantShardSplitPreservesKeys:	{},
	RequiredInvariantShardMergePreservesKeys:	{},
}

var requiredSimulationCoverage = map[string]struct{}{
	RequiredSimulationHighVolumeBankTransfers:	{},
	RequiredSimulationIdentityUpdateBursts:		{},
	RequiredSimulationContractAsyncChains:		{},
	RequiredSimulationPaymentTimeoutBounce:		{},
	RequiredSimulationCrossZoneCongestion:		{},
	RequiredSimulationShardSplitLoad:		{},
	RequiredSimulationAdaptiveSyncQueues:		{},
}

var requiredPerformanceCoverage = map[string]struct{}{
	RequiredPerformanceLocalZoneTPS:		{},
	RequiredPerformanceCrossShardThroughput:	{},
	RequiredPerformanceCrossZoneThroughput:		{},
	RequiredPerformanceAVMInstructionRate:		{},
	RequiredPerformanceStoreV2ProofLatency:		{},
	RequiredPerformanceBlockSTMConflictRate:	{},
	RequiredPerformanceMempoolGrouping:		{},
	RequiredPerformanceStateSyncMultipleZones:	{},
}

type RequiredTestCoverageCase struct {
	CaseID		string
	Class		RequiredCoverageClass
	Target		string
	EvidenceHash	string
	Deterministic	bool
	Covered		bool
}

type RequiredTestCoverageInput struct {
	CoverageVersion	string
	Determinism	[]RequiredTestCoverageCase
	Invariants	[]RequiredTestCoverageCase
	Simulations	[]RequiredTestCoverageCase
	Performance	[]RequiredTestCoverageCase
}

type RequiredTestCoverageReport struct {
	CoverageVersion	string
	Passed		bool
	Failed		[]string
	Evidence	[]string
	ReportHash	string
}

func BuildRequiredTestCoverageReport(input RequiredTestCoverageInput) RequiredTestCoverageReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	if err := validateExecutionToken("required test coverage version", input.CoverageVersion); err != nil {
		failed = append(failed, "coverage_version")
	} else {
		evidence = append(evidence, "coverage_version:"+input.CoverageVersion)
	}
	if err := validateRequiredCoverageCases(RequiredCoverageDeterminism, input.Determinism, requiredDeterminismCoverage); err != nil {
		failed = append(failed, "determinism_tests")
	} else {
		evidence = append(evidence, "determinism_tests:"+hashRequiredCoverageCases(RequiredCoverageDeterminism, input.Determinism))
	}
	if err := validateRequiredCoverageCases(RequiredCoverageInvariant, input.Invariants, requiredInvariantCoverage); err != nil {
		failed = append(failed, "invariant_tests")
	} else {
		evidence = append(evidence, "invariant_tests:"+hashRequiredCoverageCases(RequiredCoverageInvariant, input.Invariants))
	}
	if err := validateRequiredCoverageCases(RequiredCoverageSimulation, input.Simulations, requiredSimulationCoverage); err != nil {
		failed = append(failed, "simulation_tests")
	} else {
		evidence = append(evidence, "simulation_tests:"+hashRequiredCoverageCases(RequiredCoverageSimulation, input.Simulations))
	}
	if err := validateRequiredCoverageCases(RequiredCoveragePerformance, input.Performance, requiredPerformanceCoverage); err != nil {
		failed = append(failed, "performance_tests")
	} else {
		evidence = append(evidence, "performance_tests:"+hashRequiredCoverageCases(RequiredCoveragePerformance, input.Performance))
	}
	report := RequiredTestCoverageReport{
		CoverageVersion:	input.CoverageVersion,
		Passed:			len(failed) == 0,
		Failed:			normalizeStringSet(failed),
		Evidence:		normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeRequiredTestCoverageReportHash(report)
	return report
}

func (i RequiredTestCoverageInput) Normalize() RequiredTestCoverageInput {
	i.CoverageVersion = strings.TrimSpace(i.CoverageVersion)
	for idx := range i.Determinism {
		i.Determinism[idx] = i.Determinism[idx].Normalize()
	}
	sort.SliceStable(i.Determinism, func(left, right int) bool {
		return i.Determinism[left].CaseID < i.Determinism[right].CaseID
	})
	for idx := range i.Invariants {
		i.Invariants[idx] = i.Invariants[idx].Normalize()
	}
	sort.SliceStable(i.Invariants, func(left, right int) bool {
		return i.Invariants[left].CaseID < i.Invariants[right].CaseID
	})
	for idx := range i.Simulations {
		i.Simulations[idx] = i.Simulations[idx].Normalize()
	}
	sort.SliceStable(i.Simulations, func(left, right int) bool {
		return i.Simulations[left].CaseID < i.Simulations[right].CaseID
	})
	for idx := range i.Performance {
		i.Performance[idx] = i.Performance[idx].Normalize()
	}
	sort.SliceStable(i.Performance, func(left, right int) bool {
		return i.Performance[left].CaseID < i.Performance[right].CaseID
	})
	return i
}

func (c RequiredTestCoverageCase) Normalize() RequiredTestCoverageCase {
	c.CaseID = strings.TrimSpace(c.CaseID)
	c.Class = RequiredCoverageClass(strings.TrimSpace(string(c.Class)))
	c.Target = strings.TrimSpace(c.Target)
	c.EvidenceHash = normalizeLowerHex(c.EvidenceHash)
	return c
}

func (c RequiredTestCoverageCase) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("required test coverage case id", check.CaseID); err != nil {
		return err
	}
	if !IsRequiredCoverageClass(check.Class) {
		return errors.New("required test coverage class is unsupported")
	}
	if err := validateExecutionToken("required test coverage target", check.Target); err != nil {
		return err
	}
	if !check.Covered || !check.Deterministic {
		return errors.New("required test coverage case must be covered and deterministic")
	}
	return validateHexHash("required test coverage evidence hash", check.EvidenceHash)
}

func (r RequiredTestCoverageReport) Validate() error {
	if err := validateExecutionToken("required test coverage report version", r.CoverageVersion); err != nil {
		return err
	}
	if r.Passed && len(r.Failed) > 0 {
		return errors.New("required test coverage passed report must not include failures")
	}
	if len(r.Evidence) == 0 {
		return errors.New("required test coverage evidence is required")
	}
	if r.ReportHash != ComputeRequiredTestCoverageReportHash(r) {
		return errors.New("required test coverage report hash mismatch")
	}
	return nil
}

func IsRequiredCoverageClass(class RequiredCoverageClass) bool {
	switch class {
	case RequiredCoverageDeterminism, RequiredCoverageInvariant, RequiredCoverageSimulation, RequiredCoveragePerformance:
		return true
	default:
		return false
	}
}

func validateRequiredCoverageCases(class RequiredCoverageClass, cases []RequiredTestCoverageCase, required map[string]struct{}) error {
	missing := make(map[string]struct{}, len(required))
	for caseID := range required {
		missing[caseID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(cases))
	for _, coverage := range cases {
		if err := coverage.Validate(); err != nil {
			return err
		}
		coverage = coverage.Normalize()
		if coverage.Class != class {
			return fmt.Errorf("required test coverage case %s class mismatch", coverage.CaseID)
		}
		if _, expected := required[coverage.CaseID]; !expected {
			return fmt.Errorf("required test coverage unexpected case: %s", coverage.CaseID)
		}
		if _, exists := seen[coverage.CaseID]; exists {
			return fmt.Errorf("required test coverage duplicate case: %s", coverage.CaseID)
		}
		seen[coverage.CaseID] = struct{}{}
		delete(missing, coverage.CaseID)
	}
	if len(missing) > 0 {
		return fmt.Errorf("required test coverage missing cases: %v", sortedMapKeys(missing))
	}
	return nil
}

func ComputeRequiredTestCoverageReportHash(report RequiredTestCoverageReport) string {
	failed := normalizeStringSet(report.Failed)
	evidence := normalizeStringSet(report.Evidence)
	parts := []string{"aetra-required-test-coverage-report", strings.TrimSpace(report.CoverageVersion), fmt.Sprintf("%t", report.Passed)}
	parts = append(parts, failed...)
	parts = append(parts, evidence...)
	return hashStrings(parts...)
}

func hashRequiredCoverageCases(class RequiredCoverageClass, cases []RequiredTestCoverageCase) string {
	parts := []string{"aetra-required-test-coverage-cases", string(class)}
	for _, coverage := range cases {
		coverage = coverage.Normalize()
		parts = append(parts, coverage.CaseID, string(coverage.Class), coverage.Target, coverage.EvidenceHash, fmt.Sprintf("%t", coverage.Deterministic), fmt.Sprintf("%t", coverage.Covered))
	}
	return hashStrings(parts...)
}
