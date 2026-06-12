package params

import (
	"fmt"
	"sort"
)

const (
	AetraCosmWasmPolicyModuleName	= "app/wasmconfig"

	AetraCosmWasmRoleOptionalCompatibility	= "cosmwasm_optional_gated_compatibility_layer"
	AetraCosmWasmRoleAVMPrimaryRuntime	= "avm_remains_primary_native_contract_runtime"

	AetraCosmWasmLaunchPhaseEarlyTestnet	= "early_testnet"
	AetraCosmWasmLaunchPhaseLaterTestnet	= "later_testnet"
	AetraCosmWasmLaunchPhaseMainnet		= "mainnet"

	AetraCosmWasmUploadPermissionedOrGovernanceGated	= "permissioned_or_governance_gated_upload"
	AetraCosmWasmUploadPermissionlessWithStrongFeesDeposits	= "permissionless_upload_with_strong_fees_deposits"
	AetraCosmWasmMainnetAfterSecurityReview			= "mainnet_policy_decided_after_security_review"

	AetraCosmWasmGasStorageMaxWasmCodeSize			= "max_wasm_code_size"
	AetraCosmWasmGasStorageMaxInstantiateGas		= "max_instantiate_gas"
	AetraCosmWasmGasStorageMaxExecuteGasPerTx		= "max_execute_gas_per_tx"
	AetraCosmWasmGasStorageMaxQueryGas			= "max_query_gas"
	AetraCosmWasmGasStorageStorageRentOrPricing		= "storage_rent_or_storage_pricing"
	AetraCosmWasmGasStorageContractUploadFee		= "contract_upload_fee"
	AetraCosmWasmGasStorageMigrationAuthorityRules		= "contract_migration_authority_rules"
	AetraCosmWasmGasStoragePinnedCodePolicyIfUsed		= "pinned_code_policy_if_used"
	AetraCosmWasmGasStorageGovernanceConfigurable		= "gas_storage_params_governance_configurable_with_bounds"
	AetraCosmWasmGasStorageDeterministicAccounting		= "deterministic_integer_accounting"
	AetraCosmWasmGasStorageSecurityAndBenchmarkGates	= "security_review_and_benchmark_gates_required"

	AetraCosmWasmTestLaunchPolicyCovered		= "launch_policy_tests"
	AetraCosmWasmTestGasStorageCovered		= "gas_storage_limit_tests"
	AetraCosmWasmTestUploadFeeCovered		= "upload_fee_tests"
	AetraCosmWasmTestStoragePricingCovered		= "storage_pricing_tests"
	AetraCosmWasmTestMigrationAuthority		= "migration_authority_tests"
	AetraCosmWasmTestPinnedCodePolicy		= "pinned_code_policy_tests"
	AetraCosmWasmTestAVMCompatibilityBoundary	= "avm_cosmwasm_boundary_tests"

	AetraCosmWasmSecurityTestInfiniteLoopGasLimit		= "infinite_loop_contract_hits_gas_limit"
	AetraCosmWasmSecurityTestLargeStorageWriteBounded	= "large_storage_write_bounded"
	AetraCosmWasmSecurityTestFailedContractStateSafe	= "failed_contract_does_not_corrupt_state"
	AetraCosmWasmSecurityTestReservedModuleFundsDenied	= "contract_cannot_access_reserved_module_funds"
	AetraCosmWasmSecurityTestMigrationAuthorization		= "migration_authorization_enforced"
	AetraCosmWasmSecurityTestReplySubmessageDeterminism	= "reply_submessage_behavior_deterministic"
	AetraCosmWasmSecurityTestStableEvents			= "event_emission_stable"
	AetraCosmWasmSecurityTestExportImportContracts		= "export_import_with_contracts"
	AetraCosmWasmSecurityTestQueryNoStateMutation		= "contract_query_does_not_mutate_state"
)

type AetraCosmWasmLaunchPhasePolicy struct {
	Phase		string
	UploadPolicy	string
}

type AetraCosmWasmLaunchPolicyEvidence struct {
	ModuleName	string

	OptionalGatedCompatibilityLayer	bool
	AVMRemainsPrimaryRuntime	bool
	PhasePolicies			[]AetraCosmWasmLaunchPhasePolicy
	MainnetAfterSecurityReview	bool
}

type AetraCosmWasmLaunchPolicyReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraCosmWasmGasStorageEvidence struct {
	ModuleName	string

	RequiredControls	[]string

	GovernanceConfigurableWithBounds	bool
	DeterministicIntegerAccounting		bool
	SecurityReviewAndBenchmarks		bool
}

type AetraCosmWasmGasStorageReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraCosmWasmTestEvidence struct {
	ModuleName	string

	RequiredTests	[]string
}

type AetraCosmWasmTestReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraCosmWasmContractSecurityTestEvidence struct {
	ModuleName	string

	RequiredTests	[]string
}

type AetraCosmWasmContractSecurityTestReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraCosmWasmLaunchPolicyEvidence() AetraCosmWasmLaunchPolicyEvidence {
	return AetraCosmWasmLaunchPolicyEvidence{
		ModuleName:				AetraCosmWasmPolicyModuleName,
		OptionalGatedCompatibilityLayer:	true,
		AVMRemainsPrimaryRuntime:		true,
		PhasePolicies: []AetraCosmWasmLaunchPhasePolicy{
			{Phase: AetraCosmWasmLaunchPhaseEarlyTestnet, UploadPolicy: AetraCosmWasmUploadPermissionedOrGovernanceGated},
			{Phase: AetraCosmWasmLaunchPhaseLaterTestnet, UploadPolicy: AetraCosmWasmUploadPermissionlessWithStrongFeesDeposits},
			{Phase: AetraCosmWasmLaunchPhaseMainnet, UploadPolicy: AetraCosmWasmMainnetAfterSecurityReview},
		},
		MainnetAfterSecurityReview:	true,
	}
}

func ValidateAetraCosmWasmLaunchPolicy(evidence AetraCosmWasmLaunchPolicyEvidence) error {
	report := BuildAetraCosmWasmLaunchPolicyReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra cosmwasm launch policy failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraCosmWasmLaunchPolicyReport(evidence AetraCosmWasmLaunchPolicyEvidence) AetraCosmWasmLaunchPolicyReport {
	failed := validateAetraCosmWasmModuleName(evidence.ModuleName)
	passed, policyFailures := validateAetraCosmWasmPhasePolicies(evidence.PhasePolicies)
	failed = append(failed, policyFailures...)

	checks := []requirementCheck{
		{AetraCosmWasmRoleOptionalCompatibility, evidence.OptionalGatedCompatibilityLayer},
		{AetraCosmWasmRoleAVMPrimaryRuntime, evidence.AVMRemainsPrimaryRuntime},
		{AetraCosmWasmMainnetAfterSecurityReview, evidence.MainnetAfterSecurityReview},
	}
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraCosmWasmLaunchPolicyReport{
		ModuleName:	evidence.ModuleName,
		Required:	6,
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraCosmWasmGasStorageEvidence() AetraCosmWasmGasStorageEvidence {
	return AetraCosmWasmGasStorageEvidence{
		ModuleName:		AetraCosmWasmPolicyModuleName,
		RequiredControls:	requiredAetraCosmWasmGasStorageControls(),

		GovernanceConfigurableWithBounds:	true,
		DeterministicIntegerAccounting:		true,
		SecurityReviewAndBenchmarks:		true,
	}
}

func ValidateAetraCosmWasmGasStorage(evidence AetraCosmWasmGasStorageEvidence) error {
	report := BuildAetraCosmWasmGasStorageReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra cosmwasm gas/storage policy failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraCosmWasmGasStorageReport(evidence AetraCosmWasmGasStorageEvidence) AetraCosmWasmGasStorageReport {
	failed := validateAetraCosmWasmModuleName(evidence.ModuleName)
	requiredControls := requiredAetraCosmWasmGasStorageControls()
	passed, controlFailures := validateAetraCosmWasmCatalog("gas_storage", evidence.RequiredControls, requiredControls)
	failed = append(failed, controlFailures...)

	checks := []requirementCheck{
		{AetraCosmWasmGasStorageGovernanceConfigurable, evidence.GovernanceConfigurableWithBounds},
		{AetraCosmWasmGasStorageDeterministicAccounting, evidence.DeterministicIntegerAccounting},
		{AetraCosmWasmGasStorageSecurityAndBenchmarkGates, evidence.SecurityReviewAndBenchmarks},
	}
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraCosmWasmGasStorageReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredControls) + len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraCosmWasmTestEvidence() AetraCosmWasmTestEvidence {
	return AetraCosmWasmTestEvidence{
		ModuleName:	AetraCosmWasmPolicyModuleName,
		RequiredTests:	requiredAetraCosmWasmTests(),
	}
}

func ValidateAetraCosmWasmTests(evidence AetraCosmWasmTestEvidence) error {
	report := BuildAetraCosmWasmTestReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra cosmwasm tests failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraCosmWasmTestReport(evidence AetraCosmWasmTestEvidence) AetraCosmWasmTestReport {
	failed := validateAetraCosmWasmModuleName(evidence.ModuleName)
	requiredTests := requiredAetraCosmWasmTests()
	passed, testFailures := validateAetraCosmWasmCatalog("tests", evidence.RequiredTests, requiredTests)
	failed = append(failed, testFailures...)

	sort.Strings(failed)
	return AetraCosmWasmTestReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredTests),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraCosmWasmContractSecurityTestEvidence() AetraCosmWasmContractSecurityTestEvidence {
	return AetraCosmWasmContractSecurityTestEvidence{
		ModuleName:	AetraCosmWasmPolicyModuleName,
		RequiredTests:	requiredAetraCosmWasmContractSecurityTests(),
	}
}

func ValidateAetraCosmWasmContractSecurityTests(evidence AetraCosmWasmContractSecurityTestEvidence) error {
	report := BuildAetraCosmWasmContractSecurityTestReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra cosmwasm contract security tests failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraCosmWasmContractSecurityTestReport(evidence AetraCosmWasmContractSecurityTestEvidence) AetraCosmWasmContractSecurityTestReport {
	failed := validateAetraCosmWasmModuleName(evidence.ModuleName)
	requiredTests := requiredAetraCosmWasmContractSecurityTests()
	passed, testFailures := validateAetraCosmWasmCatalog("contract_security_tests", evidence.RequiredTests, requiredTests)
	failed = append(failed, testFailures...)

	sort.Strings(failed)
	return AetraCosmWasmContractSecurityTestReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredTests),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraCosmWasmGasStorageControls() []string {
	return []string{
		AetraCosmWasmGasStorageMaxWasmCodeSize,
		AetraCosmWasmGasStorageMaxInstantiateGas,
		AetraCosmWasmGasStorageMaxExecuteGasPerTx,
		AetraCosmWasmGasStorageMaxQueryGas,
		AetraCosmWasmGasStorageStorageRentOrPricing,
		AetraCosmWasmGasStorageContractUploadFee,
		AetraCosmWasmGasStorageMigrationAuthorityRules,
		AetraCosmWasmGasStoragePinnedCodePolicyIfUsed,
	}
}

func requiredAetraCosmWasmTests() []string {
	return []string{
		AetraCosmWasmTestLaunchPolicyCovered,
		AetraCosmWasmTestGasStorageCovered,
		AetraCosmWasmTestUploadFeeCovered,
		AetraCosmWasmTestStoragePricingCovered,
		AetraCosmWasmTestMigrationAuthority,
		AetraCosmWasmTestPinnedCodePolicy,
		AetraCosmWasmTestAVMCompatibilityBoundary,
	}
}

func requiredAetraCosmWasmContractSecurityTests() []string {
	return []string{
		AetraCosmWasmSecurityTestInfiniteLoopGasLimit,
		AetraCosmWasmSecurityTestLargeStorageWriteBounded,
		AetraCosmWasmSecurityTestFailedContractStateSafe,
		AetraCosmWasmSecurityTestReservedModuleFundsDenied,
		AetraCosmWasmSecurityTestMigrationAuthorization,
		AetraCosmWasmSecurityTestReplySubmessageDeterminism,
		AetraCosmWasmSecurityTestStableEvents,
		AetraCosmWasmSecurityTestExportImportContracts,
		AetraCosmWasmSecurityTestQueryNoStateMutation,
	}
}

func validateAetraCosmWasmModuleName(moduleName string) []string {
	if moduleName == "" {
		return []string{"module_name_required"}
	}
	if moduleName != AetraCosmWasmPolicyModuleName {
		return []string{"module_name_must_be_" + AetraCosmWasmPolicyModuleName}
	}
	return nil
}

func validateAetraCosmWasmPhasePolicies(policies []AetraCosmWasmLaunchPhasePolicy) (int, []string) {
	required := map[string]string{
		AetraCosmWasmLaunchPhaseEarlyTestnet:	AetraCosmWasmUploadPermissionedOrGovernanceGated,
		AetraCosmWasmLaunchPhaseLaterTestnet:	AetraCosmWasmUploadPermissionlessWithStrongFeesDeposits,
		AetraCosmWasmLaunchPhaseMainnet:	AetraCosmWasmMainnetAfterSecurityReview,
	}
	seen := map[string]string{}
	failed := make([]string, 0)
	for _, policy := range policies {
		if policy.Phase == "" || policy.UploadPolicy == "" {
			failed = append(failed, "launch_policy.item_required")
			continue
		}
		if _, exists := seen[policy.Phase]; exists {
			failed = append(failed, "launch_policy."+policy.Phase+":duplicate")
			continue
		}
		seen[policy.Phase] = policy.UploadPolicy
		if expected, ok := required[policy.Phase]; !ok {
			failed = append(failed, "launch_policy."+policy.Phase+":unexpected")
		} else if policy.UploadPolicy != expected {
			failed = append(failed, "launch_policy."+policy.Phase+":wrong_policy")
		}
	}

	passed := 0
	for phase, expected := range required {
		if seen[phase] == expected {
			passed++
		} else if _, ok := seen[phase]; !ok {
			failed = append(failed, "launch_policy."+phase+":missing")
		}
	}
	return passed, failed
}

func validateAetraCosmWasmCatalog(group string, actual []string, required []string) (int, []string) {
	failed := make([]string, 0)
	requiredSet := map[string]bool{}
	for _, item := range required {
		requiredSet[item] = true
	}
	seen := map[string]bool{}
	for _, item := range actual {
		if item == "" {
			failed = append(failed, group+".item_required")
			continue
		}
		if seen[item] {
			failed = append(failed, group+"."+item+":duplicate")
			continue
		}
		seen[item] = true
		if !requiredSet[item] {
			failed = append(failed, group+"."+item+":unexpected")
		}
	}
	passed := 0
	for _, item := range required {
		if seen[item] {
			passed++
		} else {
			failed = append(failed, group+"."+item+":missing")
		}
	}
	return passed, failed
}
