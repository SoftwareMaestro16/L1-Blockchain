package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ImplementationBacklogPriority string

const (
	ImplementationBacklogHigh	ImplementationBacklogPriority	= "HIGH"
	ImplementationBacklogMedium	ImplementationBacklogPriority	= "MEDIUM"
	ImplementationBacklogLower	ImplementationBacklogPriority	= "LOWER"
)

const (
	BacklogTaskZoneDescriptors			= "zone_descriptors"
	BacklogTaskAetraCoreSkeleton			= "aetracore_skeleton"
	BacklogTaskGlobalRootHierarchy			= "global_root_hierarchy"
	BacklogTaskMsgBusMessageEncoding		= "msgbus_message_encoding"
	BacklogTaskLocalMessageStores			= "local_message_stores"
	BacklogTaskStoreV2KeyPrefixPlan			= "store_v2_key_prefix_plan"
	BacklogTaskBlockSTMZoneBatchConflictTests	= "blockstm_zone_batch_conflict_tests"
	BacklogTaskDeterministicRoutingTable		= "deterministic_routing_table_format"
	BacklogTaskProofRegistrySchema			= "proof_registry_schema"

	BacklogTaskFinancialZoneExtraction	= "financial_zone_extraction"
	BacklogTaskIdentityZoneActivation	= "identity_zone_activation"
	BacklogTaskPerZoneMempoolLanes		= "per_zone_mempool_lanes"
	BacklogTaskPerShardFeeAccumulators	= "per_shard_fee_accumulators"
	BacklogTaskShardSplitMergeScheduler	= "shard_split_merge_scheduler"
	BacklogTaskAVMBytecodeGasTable		= "avm_bytecode_gas_table"
	BacklogTaskPaymentSettlementState	= "payment_settlement_state"
	BacklogTaskCrossZoneIdentityLookup	= "cross_zone_identity_lookup"

	BacklogTaskDynamicRouteCapacityScoring	= "dynamic_route_capacity_scoring"
	BacklogTaskVirtualPaymentChannels	= "virtual_payment_channels"
	BacklogTaskAdvancedABIIntrospection	= "advanced_abi_introspection"
	BacklogTaskVMNativeResolverContracts	= "vm_native_resolver_contracts"
	BacklogTaskValidatorServiceMetadata	= "validator_service_metadata"
	BacklogTaskZoneStateRentPolicies	= "zone_state_rent_policies"
)

var requiredHighPriorityBacklogTasks = map[string]struct{}{
	BacklogTaskZoneDescriptors:			{},
	BacklogTaskAetraCoreSkeleton:			{},
	BacklogTaskGlobalRootHierarchy:			{},
	BacklogTaskMsgBusMessageEncoding:		{},
	BacklogTaskLocalMessageStores:			{},
	BacklogTaskStoreV2KeyPrefixPlan:		{},
	BacklogTaskBlockSTMZoneBatchConflictTests:	{},
	BacklogTaskDeterministicRoutingTable:		{},
	BacklogTaskProofRegistrySchema:			{},
}

var requiredMediumPriorityBacklogTasks = map[string]struct{}{
	BacklogTaskFinancialZoneExtraction:	{},
	BacklogTaskIdentityZoneActivation:	{},
	BacklogTaskPerZoneMempoolLanes:		{},
	BacklogTaskPerShardFeeAccumulators:	{},
	BacklogTaskShardSplitMergeScheduler:	{},
	BacklogTaskAVMBytecodeGasTable:		{},
	BacklogTaskPaymentSettlementState:	{},
	BacklogTaskCrossZoneIdentityLookup:	{},
}

var requiredLowerPriorityBacklogTasks = map[string]struct{}{
	BacklogTaskDynamicRouteCapacityScoring:	{},
	BacklogTaskVirtualPaymentChannels:	{},
	BacklogTaskAdvancedABIIntrospection:	{},
	BacklogTaskVMNativeResolverContracts:	{},
	BacklogTaskValidatorServiceMetadata:	{},
	BacklogTaskZoneStateRentPolicies:	{},
}

type ImplementationBacklogTaskCheck struct {
	TaskID		string
	Priority	ImplementationBacklogPriority
	Component	string
	EvidenceHash	string
	TestHash	string
	Implemented	bool
	Deterministic	bool
}

type ImplementationBacklogInput struct {
	BacklogVersion	string
	HighPriority	[]ImplementationBacklogTaskCheck
	MediumPriority	[]ImplementationBacklogTaskCheck
	LowerPriority	[]ImplementationBacklogTaskCheck
}

type ImplementationBacklogReport struct {
	BacklogVersion	string
	Passed		bool
	Failed		[]string
	Evidence	[]string
	ReportHash	string
}

func BuildImplementationBacklogReport(input ImplementationBacklogInput) ImplementationBacklogReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	if err := validateExecutionToken("implementation backlog version", input.BacklogVersion); err != nil {
		failed = append(failed, "backlog_version")
	} else {
		evidence = append(evidence, "backlog_version:"+input.BacklogVersion)
	}
	if err := validateBacklogTaskSet(ImplementationBacklogHigh, input.HighPriority, requiredHighPriorityBacklogTasks); err != nil {
		failed = append(failed, "high_priority_backlog")
	} else {
		evidence = append(evidence, "high_priority_backlog:"+hashBacklogTaskChecks(ImplementationBacklogHigh, input.HighPriority))
	}
	if err := validateBacklogTaskSet(ImplementationBacklogMedium, input.MediumPriority, requiredMediumPriorityBacklogTasks); err != nil {
		failed = append(failed, "medium_priority_backlog")
	} else {
		evidence = append(evidence, "medium_priority_backlog:"+hashBacklogTaskChecks(ImplementationBacklogMedium, input.MediumPriority))
	}
	if err := validateBacklogTaskSet(ImplementationBacklogLower, input.LowerPriority, requiredLowerPriorityBacklogTasks); err != nil {
		failed = append(failed, "lower_priority_backlog")
	} else {
		evidence = append(evidence, "lower_priority_backlog:"+hashBacklogTaskChecks(ImplementationBacklogLower, input.LowerPriority))
	}
	report := ImplementationBacklogReport{
		BacklogVersion:	input.BacklogVersion,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeImplementationBacklogReportHash(report)
	return report
}

func (i ImplementationBacklogInput) Normalize() ImplementationBacklogInput {
	i.BacklogVersion = strings.TrimSpace(i.BacklogVersion)
	for idx := range i.HighPriority {
		i.HighPriority[idx] = i.HighPriority[idx].Normalize()
	}
	sort.SliceStable(i.HighPriority, func(left, right int) bool {
		return i.HighPriority[left].TaskID < i.HighPriority[right].TaskID
	})
	for idx := range i.MediumPriority {
		i.MediumPriority[idx] = i.MediumPriority[idx].Normalize()
	}
	sort.SliceStable(i.MediumPriority, func(left, right int) bool {
		return i.MediumPriority[left].TaskID < i.MediumPriority[right].TaskID
	})
	for idx := range i.LowerPriority {
		i.LowerPriority[idx] = i.LowerPriority[idx].Normalize()
	}
	sort.SliceStable(i.LowerPriority, func(left, right int) bool {
		return i.LowerPriority[left].TaskID < i.LowerPriority[right].TaskID
	})
	return i
}

func (c ImplementationBacklogTaskCheck) Normalize() ImplementationBacklogTaskCheck {
	c.TaskID = strings.TrimSpace(c.TaskID)
	c.Priority = ImplementationBacklogPriority(strings.TrimSpace(string(c.Priority)))
	c.Component = strings.TrimSpace(c.Component)
	c.EvidenceHash = normalizeLowerHex(c.EvidenceHash)
	c.TestHash = normalizeLowerHex(c.TestHash)
	return c
}

func (c ImplementationBacklogTaskCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("implementation backlog task id", check.TaskID); err != nil {
		return err
	}
	if !IsImplementationBacklogPriority(check.Priority) {
		return errors.New("implementation backlog priority is unsupported")
	}
	if err := validateExecutionToken("implementation backlog component", check.Component); err != nil {
		return err
	}
	if !check.Implemented || !check.Deterministic {
		return errors.New("implementation backlog task must be implemented and deterministic")
	}
	if err := validateHexHash("implementation backlog evidence hash", check.EvidenceHash); err != nil {
		return err
	}
	return validateHexHash("implementation backlog test hash", check.TestHash)
}

func (r ImplementationBacklogReport) Validate() error {
	if err := validateExecutionToken("implementation backlog report version", r.BacklogVersion); err != nil {
		return err
	}
	if r.Passed && len(r.Failed) > 0 {
		return errors.New("implementation backlog passed report must not include failures")
	}
	if len(r.Evidence) == 0 {
		return errors.New("implementation backlog evidence is required")
	}
	if r.ReportHash != ComputeImplementationBacklogReportHash(r) {
		return errors.New("implementation backlog report hash mismatch")
	}
	return nil
}

func IsImplementationBacklogPriority(priority ImplementationBacklogPriority) bool {
	switch priority {
	case ImplementationBacklogHigh, ImplementationBacklogMedium, ImplementationBacklogLower:
		return true
	default:
		return false
	}
}

func validateBacklogTaskSet(priority ImplementationBacklogPriority, checks []ImplementationBacklogTaskCheck, required map[string]struct{}) error {
	missing := make(map[string]struct{}, len(required))
	for taskID := range required {
		missing[taskID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		check = check.Normalize()
		if check.Priority != priority {
			return fmt.Errorf("implementation backlog task %s priority mismatch", check.TaskID)
		}
		if _, expected := required[check.TaskID]; !expected {
			return fmt.Errorf("implementation backlog unexpected task: %s", check.TaskID)
		}
		if _, exists := seen[check.TaskID]; exists {
			return fmt.Errorf("implementation backlog duplicate task: %s", check.TaskID)
		}
		seen[check.TaskID] = struct{}{}
		delete(missing, check.TaskID)
	}
	if len(missing) > 0 {
		return fmt.Errorf("implementation backlog missing tasks: %v", sortedMapKeys(missing))
	}
	return nil
}

func ComputeImplementationBacklogReportHash(report ImplementationBacklogReport) string {
	failed := normalizeStringSet(report.Failed)
	evidence := normalizeStringSet(report.Evidence)
	parts := []string{"implementation-backlog-report", strings.TrimSpace(report.BacklogVersion), fmt.Sprintf("%t", report.Passed)}
	parts = append(parts, failed...)
	parts = append(parts, evidence...)
	return hashStrings(parts...)
}

func hashBacklogTaskChecks(priority ImplementationBacklogPriority, checks []ImplementationBacklogTaskCheck) string {
	parts := []string{"implementation-backlog-task-checks", string(priority)}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.TaskID, string(check.Priority), check.Component, check.EvidenceHash, check.TestHash, fmt.Sprintf("%t", check.Implemented), fmt.Sprintf("%t", check.Deterministic))
	}
	return hashStrings(parts...)
}
