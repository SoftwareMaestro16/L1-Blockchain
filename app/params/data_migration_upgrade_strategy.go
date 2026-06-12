package params

import (
	"fmt"
	"sort"
)

const (
	AetraUpgradeRequirementStoreKeyDecision		= "store_key_decision"
	AetraUpgradeRequirementGenesisImportExport	= "genesis_import_export"
	AetraUpgradeRequirementMigrationHandler		= "migration_handler"
	AetraUpgradeRequirementVersionMapUpdate		= "version_map_update"
	AetraUpgradeRequirementUpgradeTest		= "upgrade_test"
	AetraUpgradeRequirementRollbackNotes		= "rollback_notes_where_possible"
	AetraUpgradeRequirementOperatorInstructions	= "operator_instructions"
)

const (
	AetraMigrationTestOldGenesisImports		= "old_genesis_imports_into_new_binary"
	AetraMigrationTestInitializesParams		= "migration_initializes_params"
	AetraMigrationTestPreservesBalances		= "migration_preserves_balances"
	AetraMigrationTestPreservesStakingState		= "migration_preserves_staking_state"
	AetraMigrationTestPreservesSlashingState	= "migration_preserves_slashing_state"
	AetraMigrationTestPreservesContractState	= "migration_preserves_contract_state_if_applicable"
	AetraMigrationTestDeterministicAppHash		= "app_hash_after_migration_is_deterministic"
)

type AetraUpgradeStrategyEvidence struct {
	ModuleName	string
	Requirements	[]string
	Tests		[]string
}

type AetraUpgradeStrategyReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraUpgradeStrategyEvidence(moduleName string) AetraUpgradeStrategyEvidence {
	return AetraUpgradeStrategyEvidence{
		ModuleName:	moduleName,
		Requirements:	RequiredAetraUpgradeRequirements(),
		Tests:		RequiredAetraMigrationTests(),
	}
}

func ValidateAetraUpgradeStrategy(evidence AetraUpgradeStrategyEvidence) error {
	report := BuildAetraUpgradeStrategyReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra upgrade strategy failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraUpgradeStrategyReport(evidence AetraUpgradeStrategyEvidence) AetraUpgradeStrategyReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	}
	passedRequirements, failedRequirements := validateUpgradeCatalog("requirements", evidence.Requirements, RequiredAetraUpgradeRequirements())
	passedTests, failedTests := validateUpgradeCatalog("tests", evidence.Tests, RequiredAetraMigrationTests())
	failed = append(failed, failedRequirements...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraUpgradeStrategyReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(RequiredAetraUpgradeRequirements()) + len(RequiredAetraMigrationTests()),
		Passed:		passedRequirements + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func RequiredAetraUpgradeRequirements() []string {
	return []string{
		AetraUpgradeRequirementStoreKeyDecision,
		AetraUpgradeRequirementGenesisImportExport,
		AetraUpgradeRequirementMigrationHandler,
		AetraUpgradeRequirementVersionMapUpdate,
		AetraUpgradeRequirementUpgradeTest,
		AetraUpgradeRequirementRollbackNotes,
		AetraUpgradeRequirementOperatorInstructions,
	}
}

func RequiredAetraMigrationTests() []string {
	return []string{
		AetraMigrationTestOldGenesisImports,
		AetraMigrationTestInitializesParams,
		AetraMigrationTestPreservesBalances,
		AetraMigrationTestPreservesStakingState,
		AetraMigrationTestPreservesSlashingState,
		AetraMigrationTestPreservesContractState,
		AetraMigrationTestDeterministicAppHash,
	}
}

func validateUpgradeCatalog(kind string, actual []string, required []string) (int, []string) {
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
