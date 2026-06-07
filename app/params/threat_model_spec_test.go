package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraValidatorCartelThreatCoversSection291(t *testing.T) {
	evidence := DefaultAetraValidatorCartelThreatEvidence()

	report := BuildAetraValidatorCartelThreatReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraThreatModelModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 16, report.Required)
	require.Contains(t, evidence.Threats, AetraThreatValidatorCartel)
	for _, control := range []string{
		AetraThreatControlValidatorSetTarget,
		AetraThreatControlValidatorPowerCap,
		AetraThreatControlTopNMonitoring,
		AetraThreatControlCommissionFloor,
		AetraThreatControlIdentityTransparency,
		AetraThreatControlGovernanceParticipationMetrics,
		AetraThreatControlDelegationWarnings,
	} {
		require.Contains(t, evidence.Controls, control)
	}
	for _, simulation := range []string{
		AetraThreatSimulationTop10Concentration,
		AetraThreatSimulationSplitIdentityValidator,
		AetraThreatSimulationDelegationOverflow,
		AetraThreatSimulationGovernanceCaptureThreshold,
	} {
		require.Contains(t, evidence.Simulations, simulation)
	}
	require.NoError(t, ValidateAetraValidatorCartelThreat(evidence))
}

func TestAetraValidatorCartelThreatRejectsMissingControlsSimulationsAndSafety(t *testing.T) {
	evidence := DefaultAetraValidatorCartelThreatEvidence()
	evidence.ModuleName = "x/other-threat-model"
	evidence.Threats = nil
	evidence.Controls = removeString(evidence.Controls,
		AetraThreatControlValidatorPowerCap,
		AetraThreatControlDelegationWarnings,
	)
	evidence.Simulations = removeString(evidence.Simulations,
		AetraThreatSimulationSplitIdentityValidator,
		AetraThreatSimulationGovernanceCaptureThreshold,
	)
	evidence.Controls = append(evidence.Controls, AetraThreatControlTopNMonitoring, "mandatory_kyc_gate")
	evidence.UsesObjectiveChainData = false
	evidence.UsesEconomicSignals = false
	evidence.AvoidsMandatoryValidatorKYC = false
	evidence.DoesNotHaltStakingOnWarning = false

	report := BuildAetraValidatorCartelThreatReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraThreatModelModuleName)
	require.Contains(t, report.Failed, "threats."+AetraThreatValidatorCartel+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlValidatorPowerCap+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlDelegationWarnings+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlTopNMonitoring+":duplicate")
	require.Contains(t, report.Failed, "controls.mandatory_kyc_gate:unexpected")
	require.Contains(t, report.Failed, "simulations."+AetraThreatSimulationSplitIdentityValidator+":missing")
	require.Contains(t, report.Failed, "simulations."+AetraThreatSimulationGovernanceCaptureThreshold+":missing")
	require.Contains(t, report.Failed, "uses_objective_chain_data")
	require.Contains(t, report.Failed, "uses_economic_signals")
	require.Contains(t, report.Failed, "avoids_mandatory_validator_kyc")
	require.Contains(t, report.Failed, "does_not_halt_staking_on_warning")
	require.Error(t, ValidateAetraValidatorCartelThreat(evidence))
}
