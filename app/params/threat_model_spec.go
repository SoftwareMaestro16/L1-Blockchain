package params

import (
	"fmt"
	"sort"
)

const (
	AetraThreatModelModuleName	= "aetra-threat-model"

	AetraThreatValidatorCartel			= "several_validators_coordinate_censorship_or_governance_capture"
	AetraThreatStakeCentralizationThroughRewards	= "large_validators_grow_faster_because_delegators_chase_apparent_safety_apr"
	AetraThreatDowntimeWeakOperators		= "too_many_low_quality_validators_reduce_liveness"
	AetraThreatGovernanceAttack			= "malicious_proposal_changes_economics_slashing_cap_or_vm_params_dangerously"
	AetraThreatContractAttack			= "malicious_cosmwasm_contract_consumes_gas_storage_exploits_permissions_or_causes_state_bloat"

	AetraThreatControlValidatorSetTarget			= "100_300_validator_target"
	AetraThreatControlValidatorPowerCap			= "validator_power_cap"
	AetraThreatControlTopNMonitoring			= "top_n_monitoring"
	AetraThreatControlCommissionFloor			= "commission_floor"
	AetraThreatControlIdentityTransparency			= "identity_transparency"
	AetraThreatControlGovernanceParticipationMetrics	= "governance_participation_metrics"
	AetraThreatControlDelegationWarnings			= "delegation_warnings"
	AetraThreatControlOverflowRewardsReduced		= "overflow_rewards_reduced"
	AetraThreatControlOverCapWarnings			= "over_cap_warnings"
	AetraThreatControlConcentrationMetrics			= "concentration_metrics"
	AetraThreatControlRewardMultiplierBasedOnCap		= "reward_multiplier_based_on_cap"
	AetraThreatControlMinimumSelfBond			= "minimum_self_bond"
	AetraThreatControlValidatorScore			= "validator_score"
	AetraThreatControlDowntimeSlashing			= "downtime_slashing"
	AetraThreatControlJail					= "jail"
	AetraThreatControlPublicMetrics				= "public_metrics"
	AetraThreatControlGradualValidatorSetGrowth		= "gradual_validator_set_growth"
	AetraThreatControlParamBounds				= "param_bounds"
	AetraThreatControlDelayedActivation			= "delayed_activation"
	AetraThreatControlEmergencyReviewWindow			= "emergency_review_window_for_critical_params"
	AetraThreatControlExplicitAuthorityChecks		= "explicit_authority_checks"
	AetraThreatControlEventMonitoring			= "event_monitoring"
	AetraThreatControlGasLimits				= "gas_limits"
	AetraThreatControlStoragePricing			= "storage_pricing"
	AetraThreatControlUploadPolicy				= "upload_policy"
	AetraThreatControlMigrationControls			= "migration_controls"
	AetraThreatControlContractSizeLimit			= "contract_size_limit"
	AetraThreatControlMaliciousContractTestSuite		= "malicious_contract_test_suite"

	AetraThreatSimulationTop10Concentration		= "top_10_concentration_simulation"
	AetraThreatSimulationSplitIdentityValidator	= "split_identity_validator_simulation"
	AetraThreatSimulationDelegationOverflow		= "delegation_overflow_simulation"
	AetraThreatSimulationGovernanceCaptureThreshold	= "governance_capture_threshold_analysis"

	AetraThreatTestOverCapRewardsLower		= "rewards_for_over_cap_validator_lower_than_normal"
	AetraThreatTestDelegatorAPROverflowPenalty	= "delegator_apr_estimate_reflects_overflow_penalty"
	AetraThreatTestCapChangeAccountingSafe		= "cap_changes_do_not_create_accounting_corruption"
	AetraThreatTestLivenessUnderOneThirdOffline	= "liveness_with_less_than_one_third_voting_power_offline"
	AetraThreatTestHaltOverOneThirdOfflineDoc	= "halt_behavior_with_more_than_one_third_offline_documented"
	AetraThreatTestRecoveryAfterValidatorsReturn	= "recovery_after_validators_return"
	AetraThreatTestDowntimePenaltiesApplied		= "downtime_penalties_applied"
	AetraThreatTestMaliciousParamProposalRejected	= "malicious_param_proposal_rejected"
	AetraThreatTestOutOfRangeValuesRejected		= "out_of_range_values_rejected"
	AetraThreatTestAuthoritySpoofingRejected	= "authority_spoofing_rejected"
	AetraThreatTestDelayedActivationWorks		= "delayed_activation_works"
	AetraThreatTestContractGasExhaustion		= "gas_exhaustion"
	AetraThreatTestContractStorageAbuse		= "storage_abuse"
	AetraThreatTestUnauthorizedMigration		= "unauthorized_migration"
	AetraThreatTestInvalidInstantiate		= "invalid_instantiate"
	AetraThreatTestMaliciousContainedExportImport	= "export_import_with_malicious_but_contained_contract_state"
)

type AetraValidatorCartelThreatEvidence struct {
	ModuleName	string

	Threats		[]string
	Controls	[]string
	Simulations	[]string

	UsesObjectiveChainData		bool
	UsesEconomicSignals		bool
	AvoidsMandatoryValidatorKYC	bool
	DoesNotHaltStakingOnWarning	bool
}

type AetraValidatorCartelThreatReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakeCentralizationRewardsThreatEvidence struct {
	ModuleName	string

	Threats		[]string
	Controls	[]string
	Tests		[]string
}

type AetraStakeCentralizationRewardsThreatReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraDowntimeWeakOperatorsThreatEvidence struct {
	ModuleName	string

	Threats		[]string
	Controls	[]string
	Tests		[]string
}

type AetraDowntimeWeakOperatorsThreatReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraGovernanceAttackThreatEvidence struct {
	ModuleName	string

	Threats		[]string
	Controls	[]string
	Tests		[]string
}

type AetraGovernanceAttackThreatReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraContractAttackThreatEvidence struct {
	ModuleName	string

	Threats		[]string
	Controls	[]string
	Tests		[]string
}

type AetraContractAttackThreatReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraValidatorCartelThreatEvidence() AetraValidatorCartelThreatEvidence {
	return AetraValidatorCartelThreatEvidence{
		ModuleName:	AetraThreatModelModuleName,
		Threats: []string{
			AetraThreatValidatorCartel,
		},
		Controls:	requiredAetraValidatorCartelControls(),
		Simulations:	requiredAetraValidatorCartelSimulations(),

		UsesObjectiveChainData:		true,
		UsesEconomicSignals:		true,
		AvoidsMandatoryValidatorKYC:	true,
		DoesNotHaltStakingOnWarning:	true,
	}
}

func ValidateAetraValidatorCartelThreat(evidence AetraValidatorCartelThreatEvidence) error {
	report := BuildAetraValidatorCartelThreatReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator cartel threat model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorCartelThreatReport(evidence AetraValidatorCartelThreatEvidence) AetraValidatorCartelThreatReport {
	failed := validateAetraThreatModelModuleName(evidence.ModuleName)
	passedThreats, failedThreats := validateAetraThreatModelCatalog("threats", evidence.Threats, []string{AetraThreatValidatorCartel})
	passedControls, failedControls := validateAetraThreatModelCatalog("controls", evidence.Controls, requiredAetraValidatorCartelControls())
	passedSimulations, failedSimulations := validateAetraThreatModelCatalog("simulations", evidence.Simulations, requiredAetraValidatorCartelSimulations())
	failed = append(failed, failedThreats...)
	failed = append(failed, failedControls...)
	failed = append(failed, failedSimulations...)

	passed := passedThreats + passedControls + passedSimulations
	for _, check := range []requirementCheck{
		{"uses_objective_chain_data", evidence.UsesObjectiveChainData},
		{"uses_economic_signals", evidence.UsesEconomicSignals},
		{"avoids_mandatory_validator_kyc", evidence.AvoidsMandatoryValidatorKYC},
		{"does_not_halt_staking_on_warning", evidence.DoesNotHaltStakingOnWarning},
	} {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorCartelThreatReport{
		ModuleName:	evidence.ModuleName,
		Required:	16,
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraStakeCentralizationRewardsThreatEvidence() AetraStakeCentralizationRewardsThreatEvidence {
	return AetraStakeCentralizationRewardsThreatEvidence{
		ModuleName:	AetraThreatModelModuleName,
		Threats: []string{
			AetraThreatStakeCentralizationThroughRewards,
		},
		Controls:	requiredAetraStakeCentralizationRewardsControls(),
		Tests:		requiredAetraStakeCentralizationRewardsTests(),
	}
}

func ValidateAetraStakeCentralizationRewardsThreat(evidence AetraStakeCentralizationRewardsThreatEvidence) error {
	report := BuildAetraStakeCentralizationRewardsThreatReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra stake centralization rewards threat model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakeCentralizationRewardsThreatReport(evidence AetraStakeCentralizationRewardsThreatEvidence) AetraStakeCentralizationRewardsThreatReport {
	failed := validateAetraThreatModelModuleName(evidence.ModuleName)
	passedThreats, failedThreats := validateAetraThreatModelCatalog("threats", evidence.Threats, []string{AetraThreatStakeCentralizationThroughRewards})
	passedControls, failedControls := validateAetraThreatModelCatalog("controls", evidence.Controls, requiredAetraStakeCentralizationRewardsControls())
	passedTests, failedTests := validateAetraThreatModelCatalog("tests", evidence.Tests, requiredAetraStakeCentralizationRewardsTests())
	failed = append(failed, failedThreats...)
	failed = append(failed, failedControls...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraStakeCentralizationRewardsThreatReport{
		ModuleName:	evidence.ModuleName,
		Required:	9,
		Passed:		passedThreats + passedControls + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraDowntimeWeakOperatorsThreatEvidence() AetraDowntimeWeakOperatorsThreatEvidence {
	return AetraDowntimeWeakOperatorsThreatEvidence{
		ModuleName:	AetraThreatModelModuleName,
		Threats: []string{
			AetraThreatDowntimeWeakOperators,
		},
		Controls:	requiredAetraDowntimeWeakOperatorsControls(),
		Tests:		requiredAetraDowntimeWeakOperatorsTests(),
	}
}

func ValidateAetraDowntimeWeakOperatorsThreat(evidence AetraDowntimeWeakOperatorsThreatEvidence) error {
	report := BuildAetraDowntimeWeakOperatorsThreatReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra downtime weak operators threat model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraDowntimeWeakOperatorsThreatReport(evidence AetraDowntimeWeakOperatorsThreatEvidence) AetraDowntimeWeakOperatorsThreatReport {
	failed := validateAetraThreatModelModuleName(evidence.ModuleName)
	passedThreats, failedThreats := validateAetraThreatModelCatalog("threats", evidence.Threats, []string{AetraThreatDowntimeWeakOperators})
	passedControls, failedControls := validateAetraThreatModelCatalog("controls", evidence.Controls, requiredAetraDowntimeWeakOperatorsControls())
	passedTests, failedTests := validateAetraThreatModelCatalog("tests", evidence.Tests, requiredAetraDowntimeWeakOperatorsTests())
	failed = append(failed, failedThreats...)
	failed = append(failed, failedControls...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraDowntimeWeakOperatorsThreatReport{
		ModuleName:	evidence.ModuleName,
		Required:	11,
		Passed:		passedThreats + passedControls + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraGovernanceAttackThreatEvidence() AetraGovernanceAttackThreatEvidence {
	return AetraGovernanceAttackThreatEvidence{
		ModuleName:	AetraThreatModelModuleName,
		Threats: []string{
			AetraThreatGovernanceAttack,
		},
		Controls:	requiredAetraGovernanceAttackControls(),
		Tests:		requiredAetraGovernanceAttackTests(),
	}
}

func ValidateAetraGovernanceAttackThreat(evidence AetraGovernanceAttackThreatEvidence) error {
	report := BuildAetraGovernanceAttackThreatReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra governance attack threat model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraGovernanceAttackThreatReport(evidence AetraGovernanceAttackThreatEvidence) AetraGovernanceAttackThreatReport {
	failed := validateAetraThreatModelModuleName(evidence.ModuleName)
	passedThreats, failedThreats := validateAetraThreatModelCatalog("threats", evidence.Threats, []string{AetraThreatGovernanceAttack})
	passedControls, failedControls := validateAetraThreatModelCatalog("controls", evidence.Controls, requiredAetraGovernanceAttackControls())
	passedTests, failedTests := validateAetraThreatModelCatalog("tests", evidence.Tests, requiredAetraGovernanceAttackTests())
	failed = append(failed, failedThreats...)
	failed = append(failed, failedControls...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraGovernanceAttackThreatReport{
		ModuleName:	evidence.ModuleName,
		Required:	10,
		Passed:		passedThreats + passedControls + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraContractAttackThreatEvidence() AetraContractAttackThreatEvidence {
	return AetraContractAttackThreatEvidence{
		ModuleName:	AetraThreatModelModuleName,
		Threats: []string{
			AetraThreatContractAttack,
		},
		Controls:	requiredAetraContractAttackControls(),
		Tests:		requiredAetraContractAttackTests(),
	}
}

func ValidateAetraContractAttackThreat(evidence AetraContractAttackThreatEvidence) error {
	report := BuildAetraContractAttackThreatReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra contract attack threat model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraContractAttackThreatReport(evidence AetraContractAttackThreatEvidence) AetraContractAttackThreatReport {
	failed := validateAetraThreatModelModuleName(evidence.ModuleName)
	passedThreats, failedThreats := validateAetraThreatModelCatalog("threats", evidence.Threats, []string{AetraThreatContractAttack})
	passedControls, failedControls := validateAetraThreatModelCatalog("controls", evidence.Controls, requiredAetraContractAttackControls())
	passedTests, failedTests := validateAetraThreatModelCatalog("tests", evidence.Tests, requiredAetraContractAttackTests())
	failed = append(failed, failedThreats...)
	failed = append(failed, failedControls...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraContractAttackThreatReport{
		ModuleName:	evidence.ModuleName,
		Required:	12,
		Passed:		passedThreats + passedControls + passedTests,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraValidatorCartelControls() []string {
	return []string{
		AetraThreatControlValidatorSetTarget,
		AetraThreatControlValidatorPowerCap,
		AetraThreatControlTopNMonitoring,
		AetraThreatControlCommissionFloor,
		AetraThreatControlIdentityTransparency,
		AetraThreatControlGovernanceParticipationMetrics,
		AetraThreatControlDelegationWarnings,
	}
}

func requiredAetraValidatorCartelSimulations() []string {
	return []string{
		AetraThreatSimulationTop10Concentration,
		AetraThreatSimulationSplitIdentityValidator,
		AetraThreatSimulationDelegationOverflow,
		AetraThreatSimulationGovernanceCaptureThreshold,
	}
}

func requiredAetraStakeCentralizationRewardsControls() []string {
	return []string{
		AetraThreatControlOverflowRewardsReduced,
		AetraThreatControlOverCapWarnings,
		AetraThreatControlCommissionFloor,
		AetraThreatControlConcentrationMetrics,
		AetraThreatControlRewardMultiplierBasedOnCap,
	}
}

func requiredAetraStakeCentralizationRewardsTests() []string {
	return []string{
		AetraThreatTestOverCapRewardsLower,
		AetraThreatTestDelegatorAPROverflowPenalty,
		AetraThreatTestCapChangeAccountingSafe,
	}
}

func requiredAetraDowntimeWeakOperatorsControls() []string {
	return []string{
		AetraThreatControlMinimumSelfBond,
		AetraThreatControlValidatorScore,
		AetraThreatControlDowntimeSlashing,
		AetraThreatControlJail,
		AetraThreatControlPublicMetrics,
		AetraThreatControlGradualValidatorSetGrowth,
	}
}

func requiredAetraDowntimeWeakOperatorsTests() []string {
	return []string{
		AetraThreatTestLivenessUnderOneThirdOffline,
		AetraThreatTestHaltOverOneThirdOfflineDoc,
		AetraThreatTestRecoveryAfterValidatorsReturn,
		AetraThreatTestDowntimePenaltiesApplied,
	}
}

func requiredAetraGovernanceAttackControls() []string {
	return []string{
		AetraThreatControlParamBounds,
		AetraThreatControlDelayedActivation,
		AetraThreatControlEmergencyReviewWindow,
		AetraThreatControlExplicitAuthorityChecks,
		AetraThreatControlEventMonitoring,
	}
}

func requiredAetraGovernanceAttackTests() []string {
	return []string{
		AetraThreatTestMaliciousParamProposalRejected,
		AetraThreatTestOutOfRangeValuesRejected,
		AetraThreatTestAuthoritySpoofingRejected,
		AetraThreatTestDelayedActivationWorks,
	}
}

func requiredAetraContractAttackControls() []string {
	return []string{
		AetraThreatControlGasLimits,
		AetraThreatControlStoragePricing,
		AetraThreatControlUploadPolicy,
		AetraThreatControlMigrationControls,
		AetraThreatControlContractSizeLimit,
		AetraThreatControlMaliciousContractTestSuite,
	}
}

func requiredAetraContractAttackTests() []string {
	return []string{
		AetraThreatTestContractGasExhaustion,
		AetraThreatTestContractStorageAbuse,
		AetraThreatTestUnauthorizedMigration,
		AetraThreatTestInvalidInstantiate,
		AetraThreatTestMaliciousContainedExportImport,
	}
}

func validateAetraThreatModelModuleName(moduleName string) []string {
	if moduleName == "" {
		return []string{"module_name_required"}
	}
	if moduleName != AetraThreatModelModuleName {
		return []string{"module_name_must_be_" + AetraThreatModelModuleName}
	}
	return nil
}

func validateAetraThreatModelCatalog(group string, actual []string, required []string) (int, []string) {
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
