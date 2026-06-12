package params

import (
	"fmt"
	"sort"
)

const (
	AetraRepoAreaProto	= "proto/"
	AetraRepoAreaX		= "x/"
	AetraRepoAreaApp	= "app/"
	AetraRepoAreaTests	= "tests/"
)

const (
	AetraRepoProtoTaskDefineMessages	= "define_protobuf_messages_for_new_modules"
	AetraRepoProtoTaskDefineQueryServices	= "define_query_services"
	AetraRepoProtoTaskDefineTxServices	= "define_tx_services"
	AetraRepoProtoTaskDefineGenesis		= "define_genesis_messages"
	AetraRepoProtoTaskDefineParams		= "define_params_messages"
	AetraRepoProtoTaskRunCodeGeneration	= "run_code_generation"
	AetraRepoProtoTaskBreakingChangeChecks	= "add_proto_breaking_change_checks_if_available"
)

const (
	AetraRepoProtoTestGeneratedCodeCompiles	= "generated_code_compiles"
	AetraRepoProtoTestLintPasses		= "proto_lint_passes_if_configured"
	AetraRepoProtoTestServiceRegistration	= "query_tx_service_registration_tested"
)

const (
	AetraRepoXTaskImplementKeepers		= "implement_keepers"
	AetraRepoXTaskImplementMsgServers	= "implement_message_servers"
	AetraRepoXTaskImplementQueryServers	= "implement_query_servers"
	AetraRepoXTaskImplementGenesis		= "implement_genesis"
	AetraRepoXTaskImplementParamsValidation	= "implement_params_validation"
	AetraRepoXTaskImplementInvariants	= "implement_invariants"
	AetraRepoXTaskImplementHooks		= "implement_hooks_where_needed"
	AetraRepoXTaskImplementEvents		= "implement_events"
	AetraRepoXTaskImplementModuleInterfaces	= "implement_module_interfaces"
)

const (
	AetraRepoXTestKeeperUnit	= "keeper_unit_tests"
	AetraRepoXTestMsgServer		= "msg_server_tests"
	AetraRepoXTestQueryServer	= "query_server_tests"
	AetraRepoXTestGenesis		= "genesis_tests"
	AetraRepoXTestInvariant		= "invariant_tests"
	AetraRepoXTestFuzzPropertyMath	= "fuzz_property_tests_for_math"
)

const (
	AetraRepoAppTaskWireKeepers			= "wire_keepers"
	AetraRepoAppTaskWireModules			= "wire_modules"
	AetraRepoAppTaskWireModuleAccountPermissions	= "wire_module_account_permissions"
	AetraRepoAppTaskWireBeginEndPreblockOrder	= "wire_begin_end_preblock_order"
	AetraRepoAppTaskWireSimulationManager		= "wire_simulation_manager_if_used"
	AetraRepoAppTaskWireAPIRoutes			= "wire_api_routes"
	AetraRepoAppTaskWireAutoCLI			= "wire_autocli_if_used"
	AetraRepoAppTaskValidateStartup			= "validate_startup"
)

const (
	AetraRepoAppTestStartup				= "app_startup"
	AetraRepoAppTestModuleAccountPermissions	= "module_account_permissions"
	AetraRepoAppTestBeginEndOrder			= "begin_end_order"
	AetraRepoAppTestExportImport			= "export_import"
	AetraRepoAppTestDeterministicRestart		= "deterministic_restart"
	AetraRepoAppTestAPIServiceRegistration		= "api_service_registration"
)

const (
	AetraRepoTestsTaskIntegrationSuites	= "integration_test_suites"
	AetraRepoTestsTaskE2ELocalnetSmoke	= "e2e_localnet_smoke_tests"
	AetraRepoTestsTaskAdversarial		= "adversarial_tests"
	AetraRepoTestsTaskLoadProfiles		= "load_profile_tests"
	AetraRepoTestsTaskDocumentationPath	= "documentation_path_tests"
	AetraRepoTestsTaskCIScripts		= "ci_scripts"
)

const (
	AetraRepoTestsRequirementDocumentedCommands	= "tests_runnable_from_documented_commands"
	AetraRepoTestsRequirementWindowsPowerShell	= "windows_powershell_local_scripts_remain_usable_if_supported"
	AetraRepoTestsRequirementLinuxCIPrimary		= "linux_ci_path_primary_for_production_confidence"
)

type AetraRepoWorkAreaEvidence struct {
	Area	string
	Tasks	[]string
	Tests	[]string
}

type AetraRepoWorkAreaReport struct {
	Area		string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraRepoProtoWorkEvidence() AetraRepoWorkAreaEvidence {
	return AetraRepoWorkAreaEvidence{
		Area:	AetraRepoAreaProto,
		Tasks:	RequiredAetraRepoProtoTasks(),
		Tests:	RequiredAetraRepoProtoTests(),
	}
}

func DefaultAetraRepoXWorkEvidence() AetraRepoWorkAreaEvidence {
	return AetraRepoWorkAreaEvidence{
		Area:	AetraRepoAreaX,
		Tasks:	RequiredAetraRepoXTasks(),
		Tests:	RequiredAetraRepoXTests(),
	}
}

func DefaultAetraRepoAppWorkEvidence() AetraRepoWorkAreaEvidence {
	return AetraRepoWorkAreaEvidence{
		Area:	AetraRepoAreaApp,
		Tasks:	RequiredAetraRepoAppTasks(),
		Tests:	RequiredAetraRepoAppTests(),
	}
}

func DefaultAetraRepoTestsWorkEvidence() AetraRepoWorkAreaEvidence {
	return AetraRepoWorkAreaEvidence{
		Area:	AetraRepoAreaTests,
		Tasks:	RequiredAetraRepoTestsTasks(),
		Tests:	RequiredAetraRepoTestsRequirements(),
	}
}

func ValidateAetraRepoProtoWork(evidence AetraRepoWorkAreaEvidence) error {
	report := BuildAetraRepoProtoWorkReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra repository proto work breakdown failed: %v", report.Failed)
	}
	return nil
}

func ValidateAetraRepoXWork(evidence AetraRepoWorkAreaEvidence) error {
	report := BuildAetraRepoXWorkReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra repository x work breakdown failed: %v", report.Failed)
	}
	return nil
}

func ValidateAetraRepoAppWork(evidence AetraRepoWorkAreaEvidence) error {
	report := BuildAetraRepoAppWorkReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra repository app work breakdown failed: %v", report.Failed)
	}
	return nil
}

func ValidateAetraRepoTestsWork(evidence AetraRepoWorkAreaEvidence) error {
	report := BuildAetraRepoTestsWorkReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra repository tests work breakdown failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraRepoProtoWorkReport(evidence AetraRepoWorkAreaEvidence) AetraRepoWorkAreaReport {
	failed := make([]string, 0)
	if evidence.Area != AetraRepoAreaProto {
		failed = append(failed, "area_must_be_"+AetraRepoAreaProto)
	}
	passedTasks, failedTasks := validateRepoWorkCatalog("tasks", evidence.Tasks, RequiredAetraRepoProtoTasks())
	passedTests, failedTests := validateRepoWorkCatalog("tests", evidence.Tests, RequiredAetraRepoProtoTests())
	failed = append(failed, failedTasks...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraRepoWorkAreaReport{
		Area:		evidence.Area,
		Required:	len(RequiredAetraRepoProtoTasks()) + len(RequiredAetraRepoProtoTests()),
		Passed:		passedTasks + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func BuildAetraRepoXWorkReport(evidence AetraRepoWorkAreaEvidence) AetraRepoWorkAreaReport {
	failed := make([]string, 0)
	if evidence.Area != AetraRepoAreaX {
		failed = append(failed, "area_must_be_"+AetraRepoAreaX)
	}
	passedTasks, failedTasks := validateRepoWorkCatalog("tasks", evidence.Tasks, RequiredAetraRepoXTasks())
	passedTests, failedTests := validateRepoWorkCatalog("tests", evidence.Tests, RequiredAetraRepoXTests())
	failed = append(failed, failedTasks...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraRepoWorkAreaReport{
		Area:		evidence.Area,
		Required:	len(RequiredAetraRepoXTasks()) + len(RequiredAetraRepoXTests()),
		Passed:		passedTasks + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func BuildAetraRepoAppWorkReport(evidence AetraRepoWorkAreaEvidence) AetraRepoWorkAreaReport {
	failed := make([]string, 0)
	if evidence.Area != AetraRepoAreaApp {
		failed = append(failed, "area_must_be_"+AetraRepoAreaApp)
	}
	passedTasks, failedTasks := validateRepoWorkCatalog("tasks", evidence.Tasks, RequiredAetraRepoAppTasks())
	passedTests, failedTests := validateRepoWorkCatalog("tests", evidence.Tests, RequiredAetraRepoAppTests())
	failed = append(failed, failedTasks...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraRepoWorkAreaReport{
		Area:		evidence.Area,
		Required:	len(RequiredAetraRepoAppTasks()) + len(RequiredAetraRepoAppTests()),
		Passed:		passedTasks + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func BuildAetraRepoTestsWorkReport(evidence AetraRepoWorkAreaEvidence) AetraRepoWorkAreaReport {
	failed := make([]string, 0)
	if evidence.Area != AetraRepoAreaTests {
		failed = append(failed, "area_must_be_"+AetraRepoAreaTests)
	}
	passedTasks, failedTasks := validateRepoWorkCatalog("tasks", evidence.Tasks, RequiredAetraRepoTestsTasks())
	passedTests, failedTests := validateRepoWorkCatalog("requirements", evidence.Tests, RequiredAetraRepoTestsRequirements())
	failed = append(failed, failedTasks...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraRepoWorkAreaReport{
		Area:		evidence.Area,
		Required:	len(RequiredAetraRepoTestsTasks()) + len(RequiredAetraRepoTestsRequirements()),
		Passed:		passedTasks + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func RequiredAetraRepoProtoTasks() []string {
	return []string{
		AetraRepoProtoTaskDefineMessages,
		AetraRepoProtoTaskDefineQueryServices,
		AetraRepoProtoTaskDefineTxServices,
		AetraRepoProtoTaskDefineGenesis,
		AetraRepoProtoTaskDefineParams,
		AetraRepoProtoTaskRunCodeGeneration,
		AetraRepoProtoTaskBreakingChangeChecks,
	}
}

func RequiredAetraRepoXTasks() []string {
	return []string{
		AetraRepoXTaskImplementKeepers,
		AetraRepoXTaskImplementMsgServers,
		AetraRepoXTaskImplementQueryServers,
		AetraRepoXTaskImplementGenesis,
		AetraRepoXTaskImplementParamsValidation,
		AetraRepoXTaskImplementInvariants,
		AetraRepoXTaskImplementHooks,
		AetraRepoXTaskImplementEvents,
		AetraRepoXTaskImplementModuleInterfaces,
	}
}

func RequiredAetraRepoAppTasks() []string {
	return []string{
		AetraRepoAppTaskWireKeepers,
		AetraRepoAppTaskWireModules,
		AetraRepoAppTaskWireModuleAccountPermissions,
		AetraRepoAppTaskWireBeginEndPreblockOrder,
		AetraRepoAppTaskWireSimulationManager,
		AetraRepoAppTaskWireAPIRoutes,
		AetraRepoAppTaskWireAutoCLI,
		AetraRepoAppTaskValidateStartup,
	}
}

func RequiredAetraRepoTestsTasks() []string {
	return []string{
		AetraRepoTestsTaskIntegrationSuites,
		AetraRepoTestsTaskE2ELocalnetSmoke,
		AetraRepoTestsTaskAdversarial,
		AetraRepoTestsTaskLoadProfiles,
		AetraRepoTestsTaskDocumentationPath,
		AetraRepoTestsTaskCIScripts,
	}
}

func RequiredAetraRepoProtoTests() []string {
	return []string{
		AetraRepoProtoTestGeneratedCodeCompiles,
		AetraRepoProtoTestLintPasses,
		AetraRepoProtoTestServiceRegistration,
	}
}

func RequiredAetraRepoXTests() []string {
	return []string{
		AetraRepoXTestKeeperUnit,
		AetraRepoXTestMsgServer,
		AetraRepoXTestQueryServer,
		AetraRepoXTestGenesis,
		AetraRepoXTestInvariant,
		AetraRepoXTestFuzzPropertyMath,
	}
}

func RequiredAetraRepoAppTests() []string {
	return []string{
		AetraRepoAppTestStartup,
		AetraRepoAppTestModuleAccountPermissions,
		AetraRepoAppTestBeginEndOrder,
		AetraRepoAppTestExportImport,
		AetraRepoAppTestDeterministicRestart,
		AetraRepoAppTestAPIServiceRegistration,
	}
}

func RequiredAetraRepoTestsRequirements() []string {
	return []string{
		AetraRepoTestsRequirementDocumentedCommands,
		AetraRepoTestsRequirementWindowsPowerShell,
		AetraRepoTestsRequirementLinuxCIPrimary,
	}
}

func validateRepoWorkCatalog(kind string, actual []string, required []string) (int, []string) {
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
