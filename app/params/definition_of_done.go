package params

import (
	"fmt"
	"sort"
)

const (
	AetraDoDRequirementCodeImplemented		= "code_is_implemented"
	AetraDoDRequirementParamsValidated		= "params_are_validated"
	AetraDoDRequirementGenesisImportExport		= "genesis_import_export_works"
	AetraDoDRequirementQuerySurface			= "query_surface_exists"
	AetraDoDRequirementOperationalEvents		= "events_exist_where_operationally_relevant"
	AetraDoDRequirementUnitTests			= "unit_tests_pass"
	AetraDoDRequirementIntegrationTests		= "integration_tests_pass"
	AetraDoDRequirementE2ELocalnetUserFlow		= "e2e_localnet_test_exists_for_user_facing_flow"
	AetraDoDRequirementOperatorUserDocs		= "docs_describe_operator_user_impact"
	AetraDoDRequirementFailureModesDocumented	= "failure_modes_are_documented"
	AetraDoDRequirementSecurityReviewed		= "security_implications_are_reviewed"
)

const (
	AetraDoDCriticalRequirementAdversarialTests	= "adversarial_tests"
	AetraDoDCriticalRequirementInvariantTests	= "invariant_tests"
	AetraDoDCriticalRequirementExportImportTest	= "export_import_test"
	AetraDoDCriticalRequirementDeterministicRestart	= "deterministic_restart_test"
	AetraDoDCriticalRequirementMigrationIfState	= "migration_test_if_state_changed"
)

type AetraDefinitionOfDoneEvidence struct {
	TaskName				string
	ConsensusEconomicsOrStakingChange	bool
	Requirements				[]string
	CriticalRequirements			[]string
}

type AetraDefinitionOfDoneReport struct {
	TaskName	string
	CriticalChange	bool
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraDefinitionOfDoneEvidence(taskName string, criticalChange bool) AetraDefinitionOfDoneEvidence {
	evidence := AetraDefinitionOfDoneEvidence{
		TaskName:				taskName,
		ConsensusEconomicsOrStakingChange:	criticalChange,
		Requirements:				RequiredAetraDoDRequirements(),
	}
	if criticalChange {
		evidence.CriticalRequirements = RequiredAetraDoDCriticalRequirements()
	}
	return evidence
}

func ValidateAetraDefinitionOfDone(evidence AetraDefinitionOfDoneEvidence) error {
	report := BuildAetraDefinitionOfDoneReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra definition of done failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraDefinitionOfDoneReport(evidence AetraDefinitionOfDoneEvidence) AetraDefinitionOfDoneReport {
	failed := make([]string, 0)
	if evidence.TaskName == "" {
		failed = append(failed, "task_name_required")
	}
	passedBase, failedBase := validateDoDCatalog("requirements", evidence.Requirements, RequiredAetraDoDRequirements())
	failed = append(failed, failedBase...)

	required := len(RequiredAetraDoDRequirements())
	passed := passedBase
	if evidence.ConsensusEconomicsOrStakingChange {
		passedCritical, failedCritical := validateDoDCatalog("critical_requirements", evidence.CriticalRequirements, RequiredAetraDoDCriticalRequirements())
		required += len(RequiredAetraDoDCriticalRequirements())
		passed += passedCritical
		failed = append(failed, failedCritical...)
	} else if len(evidence.CriticalRequirements) > 0 {
		failed = append(failed, "critical_requirements_unexpected_for_non_critical_change")
	}

	sort.Strings(failed)
	return AetraDefinitionOfDoneReport{
		TaskName:	evidence.TaskName,
		CriticalChange:	evidence.ConsensusEconomicsOrStakingChange,
		Required:	required,
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func RequiredAetraDoDRequirements() []string {
	return []string{
		AetraDoDRequirementCodeImplemented,
		AetraDoDRequirementParamsValidated,
		AetraDoDRequirementGenesisImportExport,
		AetraDoDRequirementQuerySurface,
		AetraDoDRequirementOperationalEvents,
		AetraDoDRequirementUnitTests,
		AetraDoDRequirementIntegrationTests,
		AetraDoDRequirementE2ELocalnetUserFlow,
		AetraDoDRequirementOperatorUserDocs,
		AetraDoDRequirementFailureModesDocumented,
		AetraDoDRequirementSecurityReviewed,
	}
}

func RequiredAetraDoDCriticalRequirements() []string {
	return []string{
		AetraDoDCriticalRequirementAdversarialTests,
		AetraDoDCriticalRequirementInvariantTests,
		AetraDoDCriticalRequirementExportImportTest,
		AetraDoDCriticalRequirementDeterministicRestart,
		AetraDoDCriticalRequirementMigrationIfState,
	}
}

func validateDoDCatalog(kind string, actual []string, required []string) (int, []string) {
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
