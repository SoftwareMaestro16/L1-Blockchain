package params

import (
	"fmt"
	"sort"
)

const (
	AetraRepoAreaProto = "proto/"
	AetraRepoAreaX     = "x/"
)

const (
	AetraRepoProtoTaskDefineMessages       = "define_protobuf_messages_for_new_modules"
	AetraRepoProtoTaskDefineQueryServices  = "define_query_services"
	AetraRepoProtoTaskDefineTxServices     = "define_tx_services"
	AetraRepoProtoTaskDefineGenesis        = "define_genesis_messages"
	AetraRepoProtoTaskDefineParams         = "define_params_messages"
	AetraRepoProtoTaskRunCodeGeneration    = "run_code_generation"
	AetraRepoProtoTaskBreakingChangeChecks = "add_proto_breaking_change_checks_if_available"
)

const (
	AetraRepoProtoTestGeneratedCodeCompiles = "generated_code_compiles"
	AetraRepoProtoTestLintPasses            = "proto_lint_passes_if_configured"
	AetraRepoProtoTestServiceRegistration   = "query_tx_service_registration_tested"
)

const (
	AetraRepoXTaskImplementKeepers          = "implement_keepers"
	AetraRepoXTaskImplementMsgServers       = "implement_message_servers"
	AetraRepoXTaskImplementQueryServers     = "implement_query_servers"
	AetraRepoXTaskImplementGenesis          = "implement_genesis"
	AetraRepoXTaskImplementParamsValidation = "implement_params_validation"
	AetraRepoXTaskImplementInvariants       = "implement_invariants"
	AetraRepoXTaskImplementHooks            = "implement_hooks_where_needed"
	AetraRepoXTaskImplementEvents           = "implement_events"
	AetraRepoXTaskImplementModuleInterfaces = "implement_module_interfaces"
)

const (
	AetraRepoXTestKeeperUnit       = "keeper_unit_tests"
	AetraRepoXTestMsgServer        = "msg_server_tests"
	AetraRepoXTestQueryServer      = "query_server_tests"
	AetraRepoXTestGenesis          = "genesis_tests"
	AetraRepoXTestInvariant        = "invariant_tests"
	AetraRepoXTestFuzzPropertyMath = "fuzz_property_tests_for_math"
)

type AetraRepoWorkAreaEvidence struct {
	Area  string
	Tasks []string
	Tests []string
}

type AetraRepoWorkAreaReport struct {
	Area     string
	Required int
	Passed   int
	Failed   []string
	Ready    bool
}

func DefaultAetraRepoProtoWorkEvidence() AetraRepoWorkAreaEvidence {
	return AetraRepoWorkAreaEvidence{
		Area:  AetraRepoAreaProto,
		Tasks: RequiredAetraRepoProtoTasks(),
		Tests: RequiredAetraRepoProtoTests(),
	}
}

func DefaultAetraRepoXWorkEvidence() AetraRepoWorkAreaEvidence {
	return AetraRepoWorkAreaEvidence{
		Area:  AetraRepoAreaX,
		Tasks: RequiredAetraRepoXTasks(),
		Tests: RequiredAetraRepoXTests(),
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
		Area:     evidence.Area,
		Required: len(RequiredAetraRepoProtoTasks()) + len(RequiredAetraRepoProtoTests()),
		Passed:   passedTasks + passedTests,
		Failed:   failed,
		Ready:    len(failed) == 0,
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
		Area:     evidence.Area,
		Required: len(RequiredAetraRepoXTasks()) + len(RequiredAetraRepoXTests()),
		Passed:   passedTasks + passedTests,
		Failed:   failed,
		Ready:    len(failed) == 0,
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
