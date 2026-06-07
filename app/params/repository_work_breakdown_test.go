package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraRepoProtoWorkCoversSection321(t *testing.T) {
	evidence := DefaultAetraRepoProtoWorkEvidence()

	report := BuildAetraRepoProtoWorkReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraRepoAreaProto, report.Area)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 10, report.Required)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineMessages)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineQueryServices)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineTxServices)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineGenesis)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineParams)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskRunCodeGeneration)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskBreakingChangeChecks)
	require.Contains(t, evidence.Tests, AetraRepoProtoTestGeneratedCodeCompiles)
	require.Contains(t, evidence.Tests, AetraRepoProtoTestLintPasses)
	require.Contains(t, evidence.Tests, AetraRepoProtoTestServiceRegistration)
	require.NoError(t, ValidateAetraRepoProtoWork(evidence))
}

func TestAetraRepoProtoWorkRejectsMissingTasksAndTests(t *testing.T) {
	evidence := DefaultAetraRepoProtoWorkEvidence()
	evidence.Area = "x/proto"
	evidence.Tasks = removeRepoWorkItem(evidence.Tasks,
		AetraRepoProtoTaskDefineMessages,
		AetraRepoProtoTaskDefineQueryServices,
		AetraRepoProtoTaskRunCodeGeneration,
	)
	evidence.Tests = removeRepoWorkItem(evidence.Tests,
		AetraRepoProtoTestGeneratedCodeCompiles,
		AetraRepoProtoTestServiceRegistration,
	)
	evidence.Tasks = append(evidence.Tasks, AetraRepoProtoTaskDefineParams, "manual_proto_note")
	evidence.Tests = append(evidence.Tests, AetraRepoProtoTestLintPasses, "manual_buf_review_only")

	report := BuildAetraRepoProtoWorkReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "area_must_be_"+AetraRepoAreaProto)
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskDefineMessages+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskDefineQueryServices+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskRunCodeGeneration+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskDefineParams+":duplicate")
	require.Contains(t, report.Failed, "tasks.manual_proto_note:unexpected")
	require.Contains(t, report.Failed, "tests."+AetraRepoProtoTestGeneratedCodeCompiles+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoProtoTestServiceRegistration+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoProtoTestLintPasses+":duplicate")
	require.Contains(t, report.Failed, "tests.manual_buf_review_only:unexpected")
	require.Error(t, ValidateAetraRepoProtoWork(evidence))
}

func TestDefaultAetraRepoXWorkCoversSection322(t *testing.T) {
	evidence := DefaultAetraRepoXWorkEvidence()

	report := BuildAetraRepoXWorkReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraRepoAreaX, report.Area)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 15, report.Required)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementKeepers)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementMsgServers)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementQueryServers)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementGenesis)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementParamsValidation)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementInvariants)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementHooks)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementEvents)
	require.Contains(t, evidence.Tasks, AetraRepoXTaskImplementModuleInterfaces)
	require.Contains(t, evidence.Tests, AetraRepoXTestKeeperUnit)
	require.Contains(t, evidence.Tests, AetraRepoXTestMsgServer)
	require.Contains(t, evidence.Tests, AetraRepoXTestQueryServer)
	require.Contains(t, evidence.Tests, AetraRepoXTestGenesis)
	require.Contains(t, evidence.Tests, AetraRepoXTestInvariant)
	require.Contains(t, evidence.Tests, AetraRepoXTestFuzzPropertyMath)
	require.NoError(t, ValidateAetraRepoXWork(evidence))
}

func TestAetraRepoXWorkRejectsMissingTasksAndTests(t *testing.T) {
	evidence := DefaultAetraRepoXWorkEvidence()
	evidence.Area = "modules/"
	evidence.Tasks = removeRepoWorkItem(evidence.Tasks,
		AetraRepoXTaskImplementKeepers,
		AetraRepoXTaskImplementQueryServers,
		AetraRepoXTaskImplementInvariants,
	)
	evidence.Tests = removeRepoWorkItem(evidence.Tests,
		AetraRepoXTestKeeperUnit,
		AetraRepoXTestGenesis,
		AetraRepoXTestFuzzPropertyMath,
	)
	evidence.Tasks = append(evidence.Tasks, AetraRepoXTaskImplementEvents, "manual_keeper_note")
	evidence.Tests = append(evidence.Tests, AetraRepoXTestInvariant, "manual_math_review_only")

	report := BuildAetraRepoXWorkReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "area_must_be_"+AetraRepoAreaX)
	require.Contains(t, report.Failed, "tasks."+AetraRepoXTaskImplementKeepers+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoXTaskImplementQueryServers+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoXTaskImplementInvariants+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoXTaskImplementEvents+":duplicate")
	require.Contains(t, report.Failed, "tasks.manual_keeper_note:unexpected")
	require.Contains(t, report.Failed, "tests."+AetraRepoXTestKeeperUnit+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoXTestGenesis+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoXTestFuzzPropertyMath+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoXTestInvariant+":duplicate")
	require.Contains(t, report.Failed, "tests.manual_math_review_only:unexpected")
	require.Error(t, ValidateAetraRepoXWork(evidence))
}

func TestDefaultAetraRepoAppWorkCoversSection323(t *testing.T) {
	evidence := DefaultAetraRepoAppWorkEvidence()

	report := BuildAetraRepoAppWorkReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraRepoAreaApp, report.Area)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 14, report.Required)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskWireKeepers)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskWireModules)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskWireModuleAccountPermissions)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskWireBeginEndPreblockOrder)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskWireSimulationManager)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskWireAPIRoutes)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskWireAutoCLI)
	require.Contains(t, evidence.Tasks, AetraRepoAppTaskValidateStartup)
	require.Contains(t, evidence.Tests, AetraRepoAppTestStartup)
	require.Contains(t, evidence.Tests, AetraRepoAppTestModuleAccountPermissions)
	require.Contains(t, evidence.Tests, AetraRepoAppTestBeginEndOrder)
	require.Contains(t, evidence.Tests, AetraRepoAppTestExportImport)
	require.Contains(t, evidence.Tests, AetraRepoAppTestDeterministicRestart)
	require.Contains(t, evidence.Tests, AetraRepoAppTestAPIServiceRegistration)
	require.NoError(t, ValidateAetraRepoAppWork(evidence))
}

func TestAetraRepoAppWorkRejectsMissingTasksAndTests(t *testing.T) {
	evidence := DefaultAetraRepoAppWorkEvidence()
	evidence.Area = "cmd/"
	evidence.Tasks = removeRepoWorkItem(evidence.Tasks,
		AetraRepoAppTaskWireKeepers,
		AetraRepoAppTaskWireModuleAccountPermissions,
		AetraRepoAppTaskValidateStartup,
	)
	evidence.Tests = removeRepoWorkItem(evidence.Tests,
		AetraRepoAppTestStartup,
		AetraRepoAppTestExportImport,
		AetraRepoAppTestAPIServiceRegistration,
	)
	evidence.Tasks = append(evidence.Tasks, AetraRepoAppTaskWireModules, "manual_app_note")
	evidence.Tests = append(evidence.Tests, AetraRepoAppTestBeginEndOrder, "manual_startup_only")

	report := BuildAetraRepoAppWorkReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "area_must_be_"+AetraRepoAreaApp)
	require.Contains(t, report.Failed, "tasks."+AetraRepoAppTaskWireKeepers+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoAppTaskWireModuleAccountPermissions+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoAppTaskValidateStartup+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoAppTaskWireModules+":duplicate")
	require.Contains(t, report.Failed, "tasks.manual_app_note:unexpected")
	require.Contains(t, report.Failed, "tests."+AetraRepoAppTestStartup+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoAppTestExportImport+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoAppTestAPIServiceRegistration+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoAppTestBeginEndOrder+":duplicate")
	require.Contains(t, report.Failed, "tests.manual_startup_only:unexpected")
	require.Error(t, ValidateAetraRepoAppWork(evidence))
}

func TestDefaultAetraRepoTestsWorkCoversSection324(t *testing.T) {
	evidence := DefaultAetraRepoTestsWorkEvidence()

	report := BuildAetraRepoTestsWorkReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraRepoAreaTests, report.Area)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 9, report.Required)
	require.Contains(t, evidence.Tasks, AetraRepoTestsTaskIntegrationSuites)
	require.Contains(t, evidence.Tasks, AetraRepoTestsTaskE2ELocalnetSmoke)
	require.Contains(t, evidence.Tasks, AetraRepoTestsTaskAdversarial)
	require.Contains(t, evidence.Tasks, AetraRepoTestsTaskLoadProfiles)
	require.Contains(t, evidence.Tasks, AetraRepoTestsTaskDocumentationPath)
	require.Contains(t, evidence.Tasks, AetraRepoTestsTaskCIScripts)
	require.Contains(t, evidence.Tests, AetraRepoTestsRequirementDocumentedCommands)
	require.Contains(t, evidence.Tests, AetraRepoTestsRequirementWindowsPowerShell)
	require.Contains(t, evidence.Tests, AetraRepoTestsRequirementLinuxCIPrimary)
	require.NoError(t, ValidateAetraRepoTestsWork(evidence))
}

func TestAetraRepoTestsWorkRejectsMissingTasksAndRequirements(t *testing.T) {
	evidence := DefaultAetraRepoTestsWorkEvidence()
	evidence.Area = "qa/"
	evidence.Tasks = removeRepoWorkItem(evidence.Tasks,
		AetraRepoTestsTaskIntegrationSuites,
		AetraRepoTestsTaskAdversarial,
		AetraRepoTestsTaskCIScripts,
	)
	evidence.Tests = removeRepoWorkItem(evidence.Tests,
		AetraRepoTestsRequirementDocumentedCommands,
		AetraRepoTestsRequirementLinuxCIPrimary,
	)
	evidence.Tasks = append(evidence.Tasks, AetraRepoTestsTaskDocumentationPath, "manual_test_plan")
	evidence.Tests = append(evidence.Tests, AetraRepoTestsRequirementWindowsPowerShell, "undocumented_local_script")

	report := BuildAetraRepoTestsWorkReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "area_must_be_"+AetraRepoAreaTests)
	require.Contains(t, report.Failed, "tasks."+AetraRepoTestsTaskIntegrationSuites+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoTestsTaskAdversarial+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoTestsTaskCIScripts+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoTestsTaskDocumentationPath+":duplicate")
	require.Contains(t, report.Failed, "tasks.manual_test_plan:unexpected")
	require.Contains(t, report.Failed, "requirements."+AetraRepoTestsRequirementDocumentedCommands+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraRepoTestsRequirementLinuxCIPrimary+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraRepoTestsRequirementWindowsPowerShell+":duplicate")
	require.Contains(t, report.Failed, "requirements.undocumented_local_script:unexpected")
	require.Error(t, ValidateAetraRepoTestsWork(evidence))
}

func removeRepoWorkItem(items []string, targets ...string) []string {
	targetSet := map[string]bool{}
	for _, target := range targets {
		targetSet[target] = true
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !targetSet[item] {
			out = append(out, item)
		}
	}
	return out
}
