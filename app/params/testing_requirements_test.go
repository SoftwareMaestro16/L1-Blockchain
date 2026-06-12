package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultTestingRequirementsCoverSection16Layers(t *testing.T) {
	report := BuildTestingRequirementsReport(nil)

	require.True(t, report.Passed)
	require.Empty(t, report.Failed)
	require.Greater(t, report.Required, 0)
	require.Equal(t, report.Required, report.Covered)
	require.NoError(t, ValidateTestingRequirements(nil))
}

func TestTestingRequirementsRejectMissingRequiredTest(t *testing.T) {
	requirements := DefaultTestingRequirements()
	requirements[0].HasTest = false

	report := BuildTestingRequirementsReport(requirements)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, requirements[0].Layer+"/"+requirements[0].ScenarioID+":missing_required_test")
	require.Error(t, ValidateTestingRequirements(requirements))
}

func TestTestingRequirementsRejectMissingLayerScenarioAndDuplicates(t *testing.T) {
	requirements := DefaultTestingRequirements()
	requirements = append(requirements, TestingLayerRequirement{Layer: TestLayerUnit, ScenarioID: TestRequirementKeeperLogic, Required: true, Feasible: true, HasTest: true, EvidenceRef: "duplicate"})
	requirements = append(requirements, TestingLayerRequirement{Layer: "manual", ScenarioID: "unknown", Required: true, Feasible: true, HasTest: true, EvidenceRef: "bad"})

	report := BuildTestingRequirementsReport(requirements)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, TestLayerUnit+"/"+TestRequirementKeeperLogic+":duplicate")
	require.Contains(t, report.Failed, "manual/unknown:unknown_layer")
	require.Contains(t, report.Failed, "manual/unknown:unknown_scenario")
}

func TestTestingRequirementsTreatFeasibleOptionalAsRequiredToImplement(t *testing.T) {
	requirements := DefaultTestingRequirements()
	for i := range requirements {
		if requirements[i].ScenarioID == TestRequirementDoubleSignEvidence {
			requirements[i].Feasible = true
			requirements[i].HasTest = false
			break
		}
	}

	report := BuildTestingRequirementsReport(requirements)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, TestLayerE2ELocalnet+"/"+TestRequirementDoubleSignEvidence+":feasible_optional_test_missing")
}

func TestFeatureTestingEvidenceRequiresTestsForCompletion(t *testing.T) {
	evidence := FeatureTestingEvidence{
		FeatureID:		"x/aetra-staking-policy",
		ImplementationReady:	true,
		UnitTests:		true,
		IntegrationTests:	true,
		E2ELocalnetTests:	true,
		AdversarialTests:	true,
		PerformanceTests:	true,
	}
	require.NoError(t, ValidateFeatureTestingEvidence(evidence))

	evidence.UnitTests = false
	require.ErrorContains(t, ValidateFeatureTestingEvidence(evidence), "feature is not complete without tests")

	evidence.UnitTests = true
	evidence.PerformanceTests = false
	evidence.DeferredReason = "not feasible in fast PR gate; scheduled for nightly"
	require.ErrorContains(t, ValidateFeatureTestingEvidence(evidence), "performance tests require explicit deferral")
}

func TestFeatureTestingEvidenceRejectsMissingIdentityOrImplementation(t *testing.T) {
	require.ErrorContains(t, ValidateFeatureTestingEvidence(FeatureTestingEvidence{}), "feature id")

	require.ErrorContains(t, ValidateFeatureTestingEvidence(FeatureTestingEvidence{
		FeatureID: "feature",
	}), "not ready")
}

func TestModuleProductionReadinessRequiresAcceptanceRule(t *testing.T) {
	evidence := ModuleProductionReadinessEvidence{
		ModuleName:		"x/aetra-staking-policy",
		UnitTestsPass:		true,
		IntegrationTestsPass:	true,
		GenesisValidationTests:	true,
		ExportImportTests:	true,
		DeterministicRestart:	true,
		AdversarialTests:	true,
		CriticalCISubset:	true,
	}
	report := BuildModuleProductionReadinessReport(evidence)
	require.True(t, report.Ready)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.NoError(t, ValidateModuleProductionReadiness(evidence))

	evidence.ExportImportTests = false
	report = BuildModuleProductionReadinessReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, ProductionAcceptanceExportImportPass)
	require.Error(t, ValidateModuleProductionReadiness(evidence))
}

func TestModuleProductionReadinessRejectsMissingModuleNameAndCI(t *testing.T) {
	evidence := ModuleProductionReadinessEvidence{
		UnitTestsPass:		true,
		IntegrationTestsPass:	true,
		GenesisValidationTests:	true,
		ExportImportTests:	true,
		DeterministicRestart:	true,
		AdversarialTests:	true,
		CriticalCISubset:	false,
	}

	report := BuildModuleProductionReadinessReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, ProductionAcceptanceCriticalCISubset)
}
