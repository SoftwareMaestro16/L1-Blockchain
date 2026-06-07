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

func TestDefaultAetraStakeCentralizationRewardsThreatCoversSection292(t *testing.T) {
	evidence := DefaultAetraStakeCentralizationRewardsThreatEvidence()

	report := BuildAetraStakeCentralizationRewardsThreatReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraThreatModelModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 9, report.Required)
	require.Contains(t, evidence.Threats, AetraThreatStakeCentralizationThroughRewards)
	for _, control := range []string{
		AetraThreatControlOverflowRewardsReduced,
		AetraThreatControlOverCapWarnings,
		AetraThreatControlCommissionFloor,
		AetraThreatControlConcentrationMetrics,
		AetraThreatControlRewardMultiplierBasedOnCap,
	} {
		require.Contains(t, evidence.Controls, control)
	}
	for _, testName := range []string{
		AetraThreatTestOverCapRewardsLower,
		AetraThreatTestDelegatorAPROverflowPenalty,
		AetraThreatTestCapChangeAccountingSafe,
	} {
		require.Contains(t, evidence.Tests, testName)
	}
	require.NoError(t, ValidateAetraStakeCentralizationRewardsThreat(evidence))
}

func TestAetraStakeCentralizationRewardsThreatRejectsMissingControlsAndTests(t *testing.T) {
	evidence := DefaultAetraStakeCentralizationRewardsThreatEvidence()
	evidence.ModuleName = ""
	evidence.Threats = nil
	evidence.Controls = removeString(evidence.Controls,
		AetraThreatControlOverflowRewardsReduced,
		AetraThreatControlRewardMultiplierBasedOnCap,
	)
	evidence.Tests = removeString(evidence.Tests,
		AetraThreatTestDelegatorAPROverflowPenalty,
		AetraThreatTestCapChangeAccountingSafe,
	)
	evidence.Controls = append(evidence.Controls, AetraThreatControlCommissionFloor, "apr_boost_for_largest_validator")

	report := BuildAetraStakeCentralizationRewardsThreatReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, "threats."+AetraThreatStakeCentralizationThroughRewards+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlOverflowRewardsReduced+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlRewardMultiplierBasedOnCap+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlCommissionFloor+":duplicate")
	require.Contains(t, report.Failed, "controls.apr_boost_for_largest_validator:unexpected")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestDelegatorAPROverflowPenalty+":missing")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestCapChangeAccountingSafe+":missing")
	require.Error(t, ValidateAetraStakeCentralizationRewardsThreat(evidence))
}

func TestDefaultAetraDowntimeWeakOperatorsThreatCoversSection293(t *testing.T) {
	evidence := DefaultAetraDowntimeWeakOperatorsThreatEvidence()

	report := BuildAetraDowntimeWeakOperatorsThreatReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraThreatModelModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 11, report.Required)
	require.Contains(t, evidence.Threats, AetraThreatDowntimeWeakOperators)
	for _, control := range []string{
		AetraThreatControlMinimumSelfBond,
		AetraThreatControlValidatorScore,
		AetraThreatControlDowntimeSlashing,
		AetraThreatControlJail,
		AetraThreatControlPublicMetrics,
		AetraThreatControlGradualValidatorSetGrowth,
	} {
		require.Contains(t, evidence.Controls, control)
	}
	for _, testName := range []string{
		AetraThreatTestLivenessUnderOneThirdOffline,
		AetraThreatTestHaltOverOneThirdOfflineDoc,
		AetraThreatTestRecoveryAfterValidatorsReturn,
		AetraThreatTestDowntimePenaltiesApplied,
	} {
		require.Contains(t, evidence.Tests, testName)
	}
	require.NoError(t, ValidateAetraDowntimeWeakOperatorsThreat(evidence))
}

func TestAetraDowntimeWeakOperatorsThreatRejectsMissingControlsAndTests(t *testing.T) {
	evidence := DefaultAetraDowntimeWeakOperatorsThreatEvidence()
	evidence.ModuleName = "x/downtime"
	evidence.Threats = nil
	evidence.Controls = removeString(evidence.Controls,
		AetraThreatControlMinimumSelfBond,
		AetraThreatControlDowntimeSlashing,
		AetraThreatControlGradualValidatorSetGrowth,
	)
	evidence.Tests = removeString(evidence.Tests,
		AetraThreatTestHaltOverOneThirdOfflineDoc,
		AetraThreatTestDowntimePenaltiesApplied,
	)
	evidence.Tests = append(evidence.Tests, AetraThreatTestRecoveryAfterValidatorsReturn, "ignore_downtime_penalties")

	report := BuildAetraDowntimeWeakOperatorsThreatReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraThreatModelModuleName)
	require.Contains(t, report.Failed, "threats."+AetraThreatDowntimeWeakOperators+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlMinimumSelfBond+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlDowntimeSlashing+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlGradualValidatorSetGrowth+":missing")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestHaltOverOneThirdOfflineDoc+":missing")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestDowntimePenaltiesApplied+":missing")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestRecoveryAfterValidatorsReturn+":duplicate")
	require.Contains(t, report.Failed, "tests.ignore_downtime_penalties:unexpected")
	require.Error(t, ValidateAetraDowntimeWeakOperatorsThreat(evidence))
}

func TestDefaultAetraGovernanceAttackThreatCoversSection294(t *testing.T) {
	evidence := DefaultAetraGovernanceAttackThreatEvidence()

	report := BuildAetraGovernanceAttackThreatReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraThreatModelModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 10, report.Required)
	require.Contains(t, evidence.Threats, AetraThreatGovernanceAttack)
	for _, control := range []string{
		AetraThreatControlParamBounds,
		AetraThreatControlDelayedActivation,
		AetraThreatControlEmergencyReviewWindow,
		AetraThreatControlExplicitAuthorityChecks,
		AetraThreatControlEventMonitoring,
	} {
		require.Contains(t, evidence.Controls, control)
	}
	for _, testName := range []string{
		AetraThreatTestMaliciousParamProposalRejected,
		AetraThreatTestOutOfRangeValuesRejected,
		AetraThreatTestAuthoritySpoofingRejected,
		AetraThreatTestDelayedActivationWorks,
	} {
		require.Contains(t, evidence.Tests, testName)
	}
	require.NoError(t, ValidateAetraGovernanceAttackThreat(evidence))
}

func TestAetraGovernanceAttackThreatRejectsMissingControlsAndTests(t *testing.T) {
	evidence := DefaultAetraGovernanceAttackThreatEvidence()
	evidence.ModuleName = ""
	evidence.Threats = nil
	evidence.Controls = removeString(evidence.Controls,
		AetraThreatControlParamBounds,
		AetraThreatControlEmergencyReviewWindow,
		AetraThreatControlExplicitAuthorityChecks,
	)
	evidence.Tests = removeString(evidence.Tests,
		AetraThreatTestOutOfRangeValuesRejected,
		AetraThreatTestAuthoritySpoofingRejected,
	)
	evidence.Controls = append(evidence.Controls, AetraThreatControlDelayedActivation, "unbounded_param_update")
	evidence.Tests = append(evidence.Tests, AetraThreatTestDelayedActivationWorks, "manual_multisig_review_only")

	report := BuildAetraGovernanceAttackThreatReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, "threats."+AetraThreatGovernanceAttack+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlParamBounds+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlEmergencyReviewWindow+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlExplicitAuthorityChecks+":missing")
	require.Contains(t, report.Failed, "controls."+AetraThreatControlDelayedActivation+":duplicate")
	require.Contains(t, report.Failed, "controls.unbounded_param_update:unexpected")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestOutOfRangeValuesRejected+":missing")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestAuthoritySpoofingRejected+":missing")
	require.Contains(t, report.Failed, "tests."+AetraThreatTestDelayedActivationWorks+":duplicate")
	require.Contains(t, report.Failed, "tests.manual_multisig_review_only:unexpected")
	require.Error(t, ValidateAetraGovernanceAttackThreat(evidence))
}
