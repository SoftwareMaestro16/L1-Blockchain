package params

import (
	"fmt"
	"sort"
)

const (
	AetraNearTermTaskAuditExistingModules		= "audit_existing_validator_economics_modules_and_map_to_spec"
	AetraNearTermTaskPowerCapCommissionParams	= "add_missing_params_validation_for_validator_power_cap_and_commission_policy"
	AetraNearTermTaskEffectivePowerQueries		= "add_effective_power_and_overflow_stake_queries"
	AetraNearTermTaskConcentrationSnapshotQuery	= "add_concentration_snapshot_query"
	AetraNearTermTaskCapMathTests			= "add_cap_math_tests_for_100_150_200_300_validator_scenarios"
	AetraNearTermTaskFeeSplitAccountingTests	= "add_fee_split_accounting_tests"
	AetraNearTermTaskInflationBoundsTests		= "add_inflation_bounds_tests"
	AetraNearTermTaskSupplyInvariantTests		= "add_supply_invariant_tests"
	AetraNearTermTaskValidatorScoreStateQueries	= "add_validator_score_state_and_query_tests"
	AetraNearTermTaskProgressiveDowntimeDecision	= "add_progressive_downtime_design_or_document_standard_slashing_v1"
	AetraNearTermTaskNominationPoolAccounting	= "add_nomination_pool_accounting_tests"
	AetraNearTermTaskAVMSmokeMalicious		= "add_avm_smoke_and_malicious_contract_tests"
	AetraNearTermTaskFinalityMeasurementScript	= "add_public_testnet_finality_measurement_script"
	AetraNearTermTaskValidatorDelegatorDocs		= "add_documentation_for_validators_and_delegators"
	AetraNearTermTaskCriticalCIGate			= "add_ci_gate_for_critical_unit_integration_tests"
)

const (
	AetraNearTermChecklistConsensusEconomicsBehavior	= "what_consensus_economics_behavior_changes"
	AetraNearTermChecklistParamsAddedChanged		= "what_params_are_added_or_changed"
	AetraNearTermChecklistTestsAdded			= "what_tests_were_added"
	AetraNearTermChecklistMigrationRisk			= "what_migration_risk_exists"
	AetraNearTermChecklistPublicDocs			= "whether_public_docs_need_updates"
)

type AetraNearTermTaskListEvidence struct {
	Tasks		[]string
	Checklist	[]string
}

type AetraNearTermTaskListReport struct {
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraNearTermTaskListEvidence() AetraNearTermTaskListEvidence {
	return AetraNearTermTaskListEvidence{
		Tasks:		RequiredAetraNearTermTasks(),
		Checklist:	RequiredAetraNearTermChecklist(),
	}
}

func ValidateAetraNearTermTaskList(evidence AetraNearTermTaskListEvidence) error {
	report := BuildAetraNearTermTaskListReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra near-term task list failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraNearTermTaskListReport(evidence AetraNearTermTaskListEvidence) AetraNearTermTaskListReport {
	passedTasks, failedTasks := validateNearTermCatalog("tasks", evidence.Tasks, RequiredAetraNearTermTasks())
	passedChecklist, failedChecklist := validateNearTermCatalog("checklist", evidence.Checklist, RequiredAetraNearTermChecklist())
	failed := append(failedTasks, failedChecklist...)
	sort.Strings(failed)
	return AetraNearTermTaskListReport{
		Required:	len(RequiredAetraNearTermTasks()) + len(RequiredAetraNearTermChecklist()),
		Passed:		passedTasks + passedChecklist,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func RequiredAetraNearTermTasks() []string {
	return []string{
		AetraNearTermTaskAuditExistingModules,
		AetraNearTermTaskPowerCapCommissionParams,
		AetraNearTermTaskEffectivePowerQueries,
		AetraNearTermTaskConcentrationSnapshotQuery,
		AetraNearTermTaskCapMathTests,
		AetraNearTermTaskFeeSplitAccountingTests,
		AetraNearTermTaskInflationBoundsTests,
		AetraNearTermTaskSupplyInvariantTests,
		AetraNearTermTaskValidatorScoreStateQueries,
		AetraNearTermTaskProgressiveDowntimeDecision,
		AetraNearTermTaskNominationPoolAccounting,
		AetraNearTermTaskAVMSmokeMalicious,
		AetraNearTermTaskFinalityMeasurementScript,
		AetraNearTermTaskValidatorDelegatorDocs,
		AetraNearTermTaskCriticalCIGate,
	}
}

func RequiredAetraNearTermChecklist() []string {
	return []string{
		AetraNearTermChecklistConsensusEconomicsBehavior,
		AetraNearTermChecklistParamsAddedChanged,
		AetraNearTermChecklistTestsAdded,
		AetraNearTermChecklistMigrationRisk,
		AetraNearTermChecklistPublicDocs,
	}
}

func validateNearTermCatalog(kind string, actual []string, required []string) (int, []string) {
	requiredSet := map[string]bool{}
	actualCounts := map[string]int{}
	for _, item := range required {
		requiredSet[item] = true
	}
	for _, item := range actual {
		actualCounts[item]++
	}

	failed := make([]string, 0)
	passed := 0
	for _, item := range required {
		switch actualCounts[item] {
		case 0:
			failed = append(failed, kind+"."+item+":missing")
		case 1:
			passed++
		default:
			failed = append(failed, kind+"."+item+":duplicate")
		}
	}
	for item := range actualCounts {
		if !requiredSet[item] {
			failed = append(failed, kind+"."+item+":unexpected")
		}
	}
	return passed, failed
}
