package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraEconomicsSpecCoversModulePurpose(t *testing.T) {
	evidence := DefaultAetraEconomicsSpecEvidence()

	report := BuildAetraEconomicsSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 5, report.Required)
	require.NoError(t, ValidateAetraEconomicsSpec(evidence))
}

func TestAetraEconomicsSpecRejectsMissingPurposeComponents(t *testing.T) {
	evidence := DefaultAetraEconomicsSpecEvidence()
	evidence.LowModerateInflation = false
	evidence.FeeBurn = false
	evidence.TreasuryAllocation = false
	evidence.RewardSmoothing = false
	evidence.TransparentAPRModel = false

	report := BuildAetraEconomicsSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraEconomicsPurposeLowModerateInflation)
	require.Contains(t, report.Failed, AetraEconomicsPurposeFeeBurn)
	require.Contains(t, report.Failed, AetraEconomicsPurposeTreasuryAllocation)
	require.Contains(t, report.Failed, AetraEconomicsPurposeRewardSmoothing)
	require.Contains(t, report.Failed, AetraEconomicsPurposeTransparentAPRModel)
	require.Error(t, ValidateAetraEconomicsSpec(evidence))
}

func TestAetraEconomicsSpecRejectsWrongModuleIdentity(t *testing.T) {
	evidence := DefaultAetraEconomicsSpecEvidence()
	evidence.ModuleName = ""

	report := BuildAetraEconomicsSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")

	evidence.ModuleName = "x/economics"
	report = BuildAetraEconomicsSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraEconomicsModuleName)
}
