package observability

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAPISurfaceCoversSection30RequiredModules(t *testing.T) {
	report := BuildAPISurfaceReadinessReport(nil)

	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, len(requiredAPIModules()), report.RequiredCount)
	require.Equal(t, report.RequiredCount, report.ReadyCount)
	require.NoError(t, ValidateAPISurfaceReadiness(nil))

	for _, module := range report.Modules {
		require.True(t, module.GRPCQuery)
		require.True(t, module.RESTQuery)
		require.True(t, module.Events)
		require.True(t, module.BoundedAttrs)
		require.True(t, module.StableResponses)
		require.True(t, module.ExamplesInDocs)
		require.Len(t, module.CLICommands, 2)
	}
}

func TestAPISurfaceRequiresQueryAndTxCommands(t *testing.T) {
	modules := DefaultAPISurfaceModuleSpecs()
	modules[0].CLICommands = modules[0].CLICommands[:1]

	report := BuildAPISurfaceReadinessReport(modules)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, modules[0].Module+":cli_tx:missing")
	require.Error(t, ValidateAPISurfaceReadiness(modules))
}

func TestAPISurfaceRejectsMissingCLIBehavior(t *testing.T) {
	modules := DefaultAPISurfaceModuleSpecs()
	modules[0].CLICommands[0].JSONOutput = false
	modules[0].CLICommands[0].HeightQuery = false
	modules[0].CLICommands[0].Pagination = false
	modules[0].CLICommands[0].ClearErrors = false
	modules[0].CLICommands[0].ExamplesInDocs = false

	report := BuildAPISurfaceReadinessReport(modules)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, modules[0].Module+":query:"+RequiredAPISurfaceJSONOutput+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":query:"+RequiredAPISurfaceHeightQuery+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":query:"+RequiredAPISurfacePagination+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":query:"+RequiredAPISurfaceClearErrors+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":query:"+RequiredAPISurfaceExamplesInDocs+":missing")
}

func TestAPISurfaceRejectsMissingTxValidation(t *testing.T) {
	modules := DefaultAPISurfaceModuleSpecs()
	modules[0].CLICommands[1].SignerValidation = false
	modules[0].CLICommands[1].AuthorityValidation = false

	report := BuildAPISurfaceReadinessReport(modules)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, modules[0].Module+":tx:signer_validation:missing")
	require.Contains(t, report.Failed, modules[0].Module+":tx:authority_validation:missing")
}

func TestAPISurfaceRejectsMissingGRPCRestEventsAndDocs(t *testing.T) {
	modules := DefaultAPISurfaceModuleSpecs()
	modules[0].GRPCQuery = false
	modules[0].RESTQuery = false
	modules[0].Events = false
	modules[0].BoundedAttrs = false
	modules[0].StableResponses = false
	modules[0].ExamplesInDocs = false

	report := BuildAPISurfaceReadinessReport(modules)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, modules[0].Module+":"+RequiredAPISurfaceGRPCQuery+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":"+RequiredAPISurfaceRESTQuery+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":"+RequiredAPISurfaceEvents+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":"+RequiredAPISurfaceBoundedAttrs+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":"+RequiredAPISurfaceStableResponses+":missing")
	require.Contains(t, report.Failed, modules[0].Module+":"+RequiredAPISurfaceExamplesInDocs+":missing")
}

func TestAPISurfaceRejectsMissingRequiredModule(t *testing.T) {
	modules := DefaultAPISurfaceModuleSpecs()
	modules = modules[:len(modules)-1]

	report := BuildAPISurfaceReadinessReport(modules)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, RequiredAPIModuleValidatorScore+":missing_module")
}
