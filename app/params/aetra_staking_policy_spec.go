package params

import (
	"fmt"
	"sort"
)

const (
	AetraStakingPolicyModuleName	= "x/aetra-staking-policy"

	AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration	= "control_effective_voting_power_delegation_overflow_commission_policy_and_anti_concentration_incentives"
	AetraStakingPolicyCentralAntiCentralizationModule				= "central_anti_centralization_module"

	AetraStakingPolicyResponsibilityRawStake			= "calculate_raw_validator_stake"
	AetraStakingPolicyResponsibilityEffectiveStake			= "calculate_effective_validator_stake"
	AetraStakingPolicyResponsibilityOverflowStake			= "calculate_overflow_stake"
	AetraStakingPolicyResponsibilityEffectiveVotingPowerCap		= "enforce_or_expose_effective_voting_power_cap"
	AetraStakingPolicyResponsibilityOverflowRewardMultiplier	= "calculate_reward_multiplier_for_overflow_stake"
	AetraStakingPolicyResponsibilityDelegationConcentrationWarning	= "expose_delegation_concentration_warnings"
	AetraStakingPolicyResponsibilityCommissionFloor			= "enforce_commission_floor"
	AetraStakingPolicyResponsibilityMaxCommission			= "enforce_max_commission"
	AetraStakingPolicyResponsibilityMaxCommissionChangeRate		= "enforce_max_commission_change_rate"
	AetraStakingPolicyResponsibilityTopNConcentrationMetrics	= "expose_top_n_concentration_metrics"
	AetraStakingPolicyResponsibilityGovernanceParamValidation	= "validate_governance_param_changes"
	AetraStakingPolicyResponsibilityPolicyChangeEvents		= "emit_events_for_cap_overflow_commission_policy_changes"
	AetraStakingPolicyResponsibilityDeterministicExportImport	= "remain_deterministic_and_export_import_safe"

	AetraStakingPolicyStateParams			= "Params"
	AetraStakingPolicyStateValidatorPolicy		= "ValidatorPolicy"
	AetraStakingPolicyStateConcentrationSnapshot	= "ConcentrationSnapshot"

	AetraStakingPolicyStateParamMaxValidatorsSoftTarget		= "MaxValidatorsSoftTarget"
	AetraStakingPolicyStateParamValidatorPowerCapBps		= "ValidatorPowerCapBps"
	AetraStakingPolicyStateParamValidatorPowerCapSchedule		= "ValidatorPowerCapSchedule"
	AetraStakingPolicyStateParamOverflowRewardMultiplierBps		= "OverflowRewardMultiplierBps"
	AetraStakingPolicyStateParamCommissionFloorBps			= "CommissionFloorBps"
	AetraStakingPolicyStateParamCommissionMaxBps			= "CommissionMaxBps"
	AetraStakingPolicyStateParamCommissionMaxDailyChangeBps		= "CommissionMaxDailyChangeBps"
	AetraStakingPolicyStateParamTop10TargetBps			= "Top10TargetBps"
	AetraStakingPolicyStateParamTop20TargetBps			= "Top20TargetBps"
	AetraStakingPolicyStateParamTop33TargetBps			= "Top33TargetBps"
	AetraStakingPolicyStateParamMinSelfBond				= "MinSelfBond"
	AetraStakingPolicyStateParamMinValidatorBond			= "MinValidatorBond"
	AetraStakingPolicyStateParamWarningThresholdBps			= "WarningThresholdBps"
	AetraStakingPolicyStateValidatorOperatorAddress			= "OperatorAddress"
	AetraStakingPolicyStateValidatorRawBondedTokens			= "RawBondedTokens"
	AetraStakingPolicyStateValidatorEffectiveBondedTokens		= "EffectiveBondedTokens"
	AetraStakingPolicyStateValidatorOverflowBondedTokens		= "OverflowBondedTokens"
	AetraStakingPolicyStateValidatorEffectivePowerBps		= "EffectivePowerBps"
	AetraStakingPolicyStateValidatorIsOverCap			= "IsOverCap"
	AetraStakingPolicyStateValidatorRewardMultiplierBps		= "RewardMultiplierBps"
	AetraStakingPolicyStateValidatorLastCommissionChangeTime	= "LastCommissionChangeTime"
	AetraStakingPolicyStateValidatorLastCommissionRateBps		= "LastCommissionRateBps"
	AetraStakingPolicyStateSnapshotHeight				= "Height"
	AetraStakingPolicyStateSnapshotBondedRatio			= "BondedRatio"
	AetraStakingPolicyStateSnapshotActiveValidators			= "ActiveValidators"
	AetraStakingPolicyStateSnapshotTop10Bps				= "Top10Bps"
	AetraStakingPolicyStateSnapshotTop20Bps				= "Top20Bps"
	AetraStakingPolicyStateSnapshotTop33Bps				= "Top33Bps"
	AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate	= "NakamotoCoefficientEstimate"
	AetraStakingPolicyStateIntegerBpsOrSDKDecimal			= "integer_basis_points_or_sdk_decimal_accounting"
	AetraStakingPolicyStateNoFloatingPoint				= "avoid_floating_point_accounting"

	AetraStakingPolicyParamValidatorPowerCapBps		= "ValidatorPowerCapBps"
	AetraStakingPolicyParamOverflowRewardMultiplierBps	= "OverflowRewardMultiplierBps"
	AetraStakingPolicyParamCommissionFloorBps		= "CommissionFloorBps"
	AetraStakingPolicyParamCommissionMaxBps			= "CommissionMaxBps"
	AetraStakingPolicyParamCommissionMaxDailyChangeBps	= "CommissionMaxDailyChangeBps"
	AetraStakingPolicyParamTop10TargetBps			= "Top10TargetBps"
	AetraStakingPolicyParamTop20TargetBps			= "Top20TargetBps"
	AetraStakingPolicyParamTop33TargetBps			= "Top33TargetBps"
	AetraStakingPolicyParamMaxValidatorsSoftTarget		= "MaxValidatorsSoftTarget"
	AetraStakingPolicyParamRejectNegativeOrOverflowMath	= "reject_negative_or_overflowing_math_values"

	AetraStakingPolicyEffectivePowerStage1	= "stage_1_rewards_and_delegation_warnings"
	AetraStakingPolicyEffectivePowerStage2	= "stage_2_capped_validator_updates"

	AetraStakingPolicyEffectivePowerDefinesCapScope			= "define_whether_cap_affects_rewards_voting_power_or_both"
	AetraStakingPolicyEffectivePowerRewards				= "cap_affects_reward_calculation"
	AetraStakingPolicyEffectivePowerDelegationWarnings		= "cap_affects_delegation_warnings"
	AetraStakingPolicyEffectivePowerCometBFTVotingPower		= "cap_affects_actual_cometbft_voting_power"
	AetraStakingPolicyEffectivePowerStage1LowConsensusRisk		= "stage_1_low_consensus_risk"
	AetraStakingPolicyEffectivePowerStage2DeepIntegrationTests	= "stage_2_requires_deeper_integration_and_heavy_tests"
	AetraStakingPolicyEffectivePowerValidatorUpdatesCapped		= "validator_updates_sent_to_cometbft_use_capped_power"
	AetraStakingPolicyEffectivePowerTotalVotingConsistent		= "total_voting_power_remains_consistent"
	AetraStakingPolicyEffectivePowerNoValidatorExceedsCap		= "no_validator_can_exceed_cap"
	AetraStakingPolicyEffectivePowerSharesCorrect			= "delegation_and_unbonding_shares_remain_correct"
	AetraStakingPolicyEffectivePowerSlashingRawStake		= "slashing_can_still_slash_underlying_raw_stake"
	AetraStakingPolicyEffectivePowerEvidenceHandlingCorrect		= "evidence_handling_remains_correct"

	AetraStakingPolicyMessageMsgUpdateStakingPolicyParams		= "MsgUpdateStakingPolicyParams"
	AetraStakingPolicyMessageMsgUpdateValidatorPowerCapSchedule	= "MsgUpdateValidatorPowerCapSchedule"
	AetraStakingPolicyMessageMsgSetCommissionPolicy			= "MsgSetCommissionPolicy"
	AetraStakingPolicyMessageMsgRegisterValidatorIdentity		= "MsgRegisterValidatorIdentity"
	AetraStakingPolicyMessageMsgUpdateValidatorIdentity		= "MsgUpdateValidatorIdentity"
	AetraStakingPolicyMessageMsgAcknowledgeOverCapWarning		= "MsgAcknowledgeOverCapWarning"

	AetraStakingPolicyMessageGovernanceOrAuthorityOnly	= "governance_or_authority_only_messages"
	AetraStakingPolicyMessageOptionalValidatorMessages	= "optional_validator_messages"
	AetraStakingPolicyMessageValidateAuthority		= "validate_authority"
	AetraStakingPolicyMessageValidateSigner			= "validate_signer"
	AetraStakingPolicyMessageRejectMalformedAddresses	= "reject_malformed_addresses"
	AetraStakingPolicyMessageRejectInvalidParams		= "reject_invalid_params"
	AetraStakingPolicyMessageEmitEvents			= "emit_events"
	AetraStakingPolicyMessageCoveredByTests			= "covered_by_tests"

	AetraStakingPolicyQueryParams			= "Query/Params"
	AetraStakingPolicyQueryValidatorPolicy		= "Query/ValidatorPolicy"
	AetraStakingPolicyQueryValidatorEffectivePower	= "Query/ValidatorEffectivePower"
	AetraStakingPolicyQueryValidatorOverflow	= "Query/ValidatorOverflow"
	AetraStakingPolicyQueryTopNConcentration	= "Query/TopNConcentration"
	AetraStakingPolicyQueryDelegationWarning	= "Query/DelegationWarning"
	AetraStakingPolicyQueryCommissionPolicy		= "Query/CommissionPolicy"
	AetraStakingPolicyQueryConcentrationSnapshot	= "Query/ConcentrationSnapshot"
	AetraStakingPolicyQueryNakamotoCoefficient	= "Query/NakamotoCoefficient"

	AetraStakingPolicyQueryStableResponses		= "query_responses_stable"
	AetraStakingPolicyQueryIndexerFriendlyResponses	= "query_responses_indexer_friendly"

	AetraStakingPolicyEventParamsUpdated		= "aetra.staking_policy.params_updated"
	AetraStakingPolicyEventValidatorOverCap		= "aetra.staking_policy.validator_over_cap"
	AetraStakingPolicyEventValidatorBackUnderCap	= "aetra.staking_policy.validator_back_under_cap"
	AetraStakingPolicyEventCommissionRejected	= "aetra.staking_policy.commission_rejected"
	AetraStakingPolicyEventConcentrationSnapshot	= "aetra.staking_policy.concentration_snapshot"
	AetraStakingPolicyEventRewardMultiplierChanged	= "aetra.staking_policy.reward_multiplier_changed"
	AetraStakingPolicyEventStableNames		= "event_names_are_stable"
	AetraStakingPolicyEventIndexerFriendlyAttrs	= "event_attributes_are_indexer_friendly"

	AetraStakingPolicyInvariantEffectivePowerCap		= "effective_power_never_exceeds_configured_cap"
	AetraStakingPolicyInvariantOverflowNonNegative		= "overflow_stake_is_never_negative"
	AetraStakingPolicyInvariantRawStakeConservation		= "raw_stake_equals_effective_stake_plus_overflow_stake"
	AetraStakingPolicyInvariantCommissionBounds		= "commission_floor_lte_commission_lte_commission_max"
	AetraStakingPolicyInvariantCommissionDailyChange	= "commission_change_lte_max_daily_change"
	AetraStakingPolicyInvariantTopNMaxHundredPercent	= "top_n_calculations_do_not_exceed_100_percent"
	AetraStakingPolicyInvariantExportImportPreserves	= "state_export_import_preserves_policy_state"
	AetraStakingPolicyInvariantCoveredByTests		= "invariants_are_covered_by_tests"

	AetraStakingPolicyTestCapMath100Validators			= "cap_math_for_100_validators"
	AetraStakingPolicyTestCapMath150Validators			= "cap_math_for_150_validators"
	AetraStakingPolicyTestCapMath250Validators			= "cap_math_for_250_validators"
	AetraStakingPolicyTestCapMath300Validators			= "cap_math_for_300_validators"
	AetraStakingPolicyTestValidatorCrossingCapUpward		= "validator_crossing_cap_upward"
	AetraStakingPolicyTestValidatorCrossingCapDownward		= "validator_crossing_cap_downward"
	AetraStakingPolicyTestDelegationToOverCapValidator		= "delegation_to_over_cap_validator"
	AetraStakingPolicyTestRedelegationFromOverCapValidator		= "redelegation_from_over_cap_validator"
	AetraStakingPolicyTestUnbondingFromOverCapValidator		= "unbonding_from_over_cap_validator"
	AetraStakingPolicyTestSlashingOverCapValidator			= "slashing_over_cap_validator"
	AetraStakingPolicyTestCommissionBelowFloorRejected		= "commission_below_floor_rejected"
	AetraStakingPolicyTestCommissionAboveMaxRejected		= "commission_above_max_rejected"
	AetraStakingPolicyTestCommissionDailyJumpRejected		= "commission_daily_jump_rejected"
	AetraStakingPolicyTestGovernanceParamUpdateAccepted		= "governance_param_update_accepted_within_bounds"
	AetraStakingPolicyTestGovernanceParamUpdateRejected		= "governance_param_update_rejected_outside_bounds"
	AetraStakingPolicyTestExportImportOverCapValidators		= "export_import_with_over_cap_validators"
	AetraStakingPolicyTestDeterministicConcentrationSnapshot	= "deterministic_concentration_snapshot"
)

type AetraStakingPolicySpecEvidence struct {
	ModuleName	string

	PurposeEffectivePowerOverflowCommissionAntiConcentration	bool
	CentralAntiCentralizationModule					bool

	CalculatesRawValidatorStake			bool
	CalculatesEffectiveValidatorStake		bool
	CalculatesOverflowStake				bool
	EnforcesOrExposesEffectiveVotingPowerCap	bool
	CalculatesOverflowRewardMultiplier		bool
	ExposesDelegationConcentrationWarnings		bool
	EnforcesCommissionFloor				bool
	EnforcesMaxCommission				bool
	EnforcesMaxCommissionChangeRate			bool
	ExposesTopNConcentrationMetrics			bool
	ValidatesGovernanceParamChanges			bool
	EmitsCapOverflowCommissionPolicyEvents		bool
	RemainsDeterministicAndExportImportSafe		bool
}

type AetraStakingPolicySpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyStateSpecEvidence struct {
	ModuleName	string

	ParamsFields			[]string
	ValidatorPolicyFields		[]string
	ConcentrationSnapshotFields	[]string

	IntegerBasisPointsOrSDKDecimals	bool
	AvoidsFloatingPoint		bool
}

type AetraStakingPolicyStateSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyBpsRule struct {
	Name		string
	MinBps		int64
	MaxBps		int64
	RecommendedMin	int64
	RecommendedMax	int64
}

type AetraStakingPolicyParameterRuleSet struct {
	ValidatorPowerCapBps		AetraStakingPolicyBpsRule
	OverflowRewardMultiplierBps	AetraStakingPolicyBpsRule
	CommissionFloorBps		AetraStakingPolicyBpsRule
	CommissionMaxBps		AetraStakingPolicyBpsRule
	CommissionMaxDailyChangeBps	AetraStakingPolicyBpsRule
}

type AetraStakingPolicyParameterValues struct {
	ValidatorPowerCapBps		int64
	OverflowRewardMultiplierBps	int64
	CommissionFloorBps		int64
	CommissionMaxBps		int64
	CommissionMaxDailyChangeBps	int64
	Top10TargetBps			int64
	Top20TargetBps			int64
	Top33TargetBps			int64
	MaxValidatorsSoftTarget		int64
}

type AetraStakingPolicyParameterReport struct {
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyEffectivePowerEvidence struct {
	ModuleName	string
	Stage		string

	DefinesCapScope			bool
	CapAffectsRewardCalculation	bool
	CapAffectsDelegationWarnings	bool
	CapAffectsCometBFTVotingPower	bool
	Stage1LowConsensusRisk		bool
	Stage2DeepIntegrationTests	bool

	ValidatorUpdatesUseCappedPower		bool
	TotalVotingPowerConsistent		bool
	NoValidatorCanExceedCap			bool
	DelegationUnbondingSharesCorrect	bool
	SlashingUsesUnderlyingRawStake		bool
	EvidenceHandlingCorrect			bool
}

type AetraStakingPolicyEffectivePowerReport struct {
	ModuleName	string
	Stage		string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyMessageSpecEvidence struct {
	ModuleName	string

	GovernanceAuthorityMessages	[]string
	OptionalValidatorMessages	[]string

	ValidateAuthority		bool
	ValidateSigner			bool
	RejectMalformedAddresses	bool
	RejectInvalidParams		bool
	EmitEvents			bool
	CoveredByTests			bool
}

type AetraStakingPolicyMessageSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyQuerySpecEvidence struct {
	ModuleName	string

	RequiredQueries	[]string

	StableResponses			bool
	IndexerFriendlyResponses	bool
}

type AetraStakingPolicyQuerySpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyEventSpecEvidence struct {
	ModuleName	string

	RequiredEvents	[]string

	StableEventNames		bool
	IndexerFriendlyAttributes	bool
}

type AetraStakingPolicyEventSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyInvariantSpecEvidence struct {
	ModuleName	string

	EffectivePowerNeverExceedsCap		bool
	OverflowStakeNeverNegative		bool
	RawStakeConservation			bool
	CommissionWithinFloorAndMax		bool
	CommissionChangeWithinDailyLimit	bool
	TopNDoesNotExceedHundredPercent		bool
	ExportImportPreservesPolicyState	bool
	CoveredByTests				bool
}

type AetraStakingPolicyInvariantSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraStakingPolicyTestSpecEvidence struct {
	ModuleName	string

	RequiredTests	[]string
}

type AetraStakingPolicyTestSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraStakingPolicySpecEvidence() AetraStakingPolicySpecEvidence {
	return AetraStakingPolicySpecEvidence{
		ModuleName:	AetraStakingPolicyModuleName,

		PurposeEffectivePowerOverflowCommissionAntiConcentration:	true,
		CentralAntiCentralizationModule:				true,

		CalculatesRawValidatorStake:			true,
		CalculatesEffectiveValidatorStake:		true,
		CalculatesOverflowStake:			true,
		EnforcesOrExposesEffectiveVotingPowerCap:	true,
		CalculatesOverflowRewardMultiplier:		true,
		ExposesDelegationConcentrationWarnings:		true,
		EnforcesCommissionFloor:			true,
		EnforcesMaxCommission:				true,
		EnforcesMaxCommissionChangeRate:		true,
		ExposesTopNConcentrationMetrics:		true,
		ValidatesGovernanceParamChanges:		true,
		EmitsCapOverflowCommissionPolicyEvents:		true,
		RemainsDeterministicAndExportImportSafe:	true,
	}
}

func ValidateAetraStakingPolicySpec(evidence AetraStakingPolicySpecEvidence) error {
	report := BuildAetraStakingPolicySpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicySpecReport(evidence AetraStakingPolicySpecEvidence) AetraStakingPolicySpecReport {
	checks := []requirementCheck{
		{AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration, evidence.PurposeEffectivePowerOverflowCommissionAntiConcentration},
		{AetraStakingPolicyCentralAntiCentralizationModule, evidence.CentralAntiCentralizationModule},
		{AetraStakingPolicyResponsibilityRawStake, evidence.CalculatesRawValidatorStake},
		{AetraStakingPolicyResponsibilityEffectiveStake, evidence.CalculatesEffectiveValidatorStake},
		{AetraStakingPolicyResponsibilityOverflowStake, evidence.CalculatesOverflowStake},
		{AetraStakingPolicyResponsibilityEffectiveVotingPowerCap, evidence.EnforcesOrExposesEffectiveVotingPowerCap},
		{AetraStakingPolicyResponsibilityOverflowRewardMultiplier, evidence.CalculatesOverflowRewardMultiplier},
		{AetraStakingPolicyResponsibilityDelegationConcentrationWarning, evidence.ExposesDelegationConcentrationWarnings},
		{AetraStakingPolicyResponsibilityCommissionFloor, evidence.EnforcesCommissionFloor},
		{AetraStakingPolicyResponsibilityMaxCommission, evidence.EnforcesMaxCommission},
		{AetraStakingPolicyResponsibilityMaxCommissionChangeRate, evidence.EnforcesMaxCommissionChangeRate},
		{AetraStakingPolicyResponsibilityTopNConcentrationMetrics, evidence.ExposesTopNConcentrationMetrics},
		{AetraStakingPolicyResponsibilityGovernanceParamValidation, evidence.ValidatesGovernanceParamChanges},
		{AetraStakingPolicyResponsibilityPolicyChangeEvents, evidence.EmitsCapOverflowCommissionPolicyEvents},
		{AetraStakingPolicyResponsibilityDeterministicExportImport, evidence.RemainsDeterministicAndExportImportSafe},
	}

	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicySpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraStakingPolicyStateSpecEvidence() AetraStakingPolicyStateSpecEvidence {
	return AetraStakingPolicyStateSpecEvidence{
		ModuleName:	AetraStakingPolicyModuleName,
		ParamsFields: []string{
			AetraStakingPolicyStateParamMaxValidatorsSoftTarget,
			AetraStakingPolicyStateParamValidatorPowerCapBps,
			AetraStakingPolicyStateParamValidatorPowerCapSchedule,
			AetraStakingPolicyStateParamOverflowRewardMultiplierBps,
			AetraStakingPolicyStateParamCommissionFloorBps,
			AetraStakingPolicyStateParamCommissionMaxBps,
			AetraStakingPolicyStateParamCommissionMaxDailyChangeBps,
			AetraStakingPolicyStateParamTop10TargetBps,
			AetraStakingPolicyStateParamTop20TargetBps,
			AetraStakingPolicyStateParamTop33TargetBps,
			AetraStakingPolicyStateParamMinSelfBond,
			AetraStakingPolicyStateParamMinValidatorBond,
			AetraStakingPolicyStateParamWarningThresholdBps,
		},
		ValidatorPolicyFields: []string{
			AetraStakingPolicyStateValidatorOperatorAddress,
			AetraStakingPolicyStateValidatorRawBondedTokens,
			AetraStakingPolicyStateValidatorEffectiveBondedTokens,
			AetraStakingPolicyStateValidatorOverflowBondedTokens,
			AetraStakingPolicyStateValidatorEffectivePowerBps,
			AetraStakingPolicyStateValidatorIsOverCap,
			AetraStakingPolicyStateValidatorRewardMultiplierBps,
			AetraStakingPolicyStateValidatorLastCommissionChangeTime,
			AetraStakingPolicyStateValidatorLastCommissionRateBps,
		},
		ConcentrationSnapshotFields: []string{
			AetraStakingPolicyStateSnapshotHeight,
			AetraStakingPolicyStateSnapshotBondedRatio,
			AetraStakingPolicyStateSnapshotActiveValidators,
			AetraStakingPolicyStateSnapshotTop10Bps,
			AetraStakingPolicyStateSnapshotTop20Bps,
			AetraStakingPolicyStateSnapshotTop33Bps,
			AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate,
		},
		IntegerBasisPointsOrSDKDecimals:	true,
		AvoidsFloatingPoint:			true,
	}
}

func ValidateAetraStakingPolicyStateSpec(evidence AetraStakingPolicyStateSpecEvidence) error {
	report := BuildAetraStakingPolicyStateSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy state spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyStateSpecReport(evidence AetraStakingPolicyStateSpecEvidence) AetraStakingPolicyStateSpecReport {
	requiredParams := requiredAetraStakingPolicyStateParamsFields()
	requiredValidator := requiredAetraStakingPolicyStateValidatorPolicyFields()
	requiredSnapshot := requiredAetraStakingPolicyStateConcentrationSnapshotFields()

	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	passed := 0
	paramsPassed, paramsFailed := validateAetraStakingPolicyStateFields(AetraStakingPolicyStateParams, evidence.ParamsFields, requiredParams)
	passed += paramsPassed
	failed = append(failed, paramsFailed...)

	validatorPassed, validatorFailed := validateAetraStakingPolicyStateFields(AetraStakingPolicyStateValidatorPolicy, evidence.ValidatorPolicyFields, requiredValidator)
	passed += validatorPassed
	failed = append(failed, validatorFailed...)

	snapshotPassed, snapshotFailed := validateAetraStakingPolicyStateFields(AetraStakingPolicyStateConcentrationSnapshot, evidence.ConcentrationSnapshotFields, requiredSnapshot)
	passed += snapshotPassed
	failed = append(failed, snapshotFailed...)

	for _, check := range []requirementCheck{
		{AetraStakingPolicyStateIntegerBpsOrSDKDecimal, evidence.IntegerBasisPointsOrSDKDecimals},
		{AetraStakingPolicyStateNoFloatingPoint, evidence.AvoidsFloatingPoint},
	} {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicyStateSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredParams) + len(requiredValidator) + len(requiredSnapshot) + 2,
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraStakingPolicyStateParamsFields() []string {
	return []string{
		AetraStakingPolicyStateParamMaxValidatorsSoftTarget,
		AetraStakingPolicyStateParamValidatorPowerCapBps,
		AetraStakingPolicyStateParamValidatorPowerCapSchedule,
		AetraStakingPolicyStateParamOverflowRewardMultiplierBps,
		AetraStakingPolicyStateParamCommissionFloorBps,
		AetraStakingPolicyStateParamCommissionMaxBps,
		AetraStakingPolicyStateParamCommissionMaxDailyChangeBps,
		AetraStakingPolicyStateParamTop10TargetBps,
		AetraStakingPolicyStateParamTop20TargetBps,
		AetraStakingPolicyStateParamTop33TargetBps,
		AetraStakingPolicyStateParamMinSelfBond,
		AetraStakingPolicyStateParamMinValidatorBond,
		AetraStakingPolicyStateParamWarningThresholdBps,
	}
}

func requiredAetraStakingPolicyStateValidatorPolicyFields() []string {
	return []string{
		AetraStakingPolicyStateValidatorOperatorAddress,
		AetraStakingPolicyStateValidatorRawBondedTokens,
		AetraStakingPolicyStateValidatorEffectiveBondedTokens,
		AetraStakingPolicyStateValidatorOverflowBondedTokens,
		AetraStakingPolicyStateValidatorEffectivePowerBps,
		AetraStakingPolicyStateValidatorIsOverCap,
		AetraStakingPolicyStateValidatorRewardMultiplierBps,
		AetraStakingPolicyStateValidatorLastCommissionChangeTime,
		AetraStakingPolicyStateValidatorLastCommissionRateBps,
	}
}

func requiredAetraStakingPolicyStateConcentrationSnapshotFields() []string {
	return []string{
		AetraStakingPolicyStateSnapshotHeight,
		AetraStakingPolicyStateSnapshotBondedRatio,
		AetraStakingPolicyStateSnapshotActiveValidators,
		AetraStakingPolicyStateSnapshotTop10Bps,
		AetraStakingPolicyStateSnapshotTop20Bps,
		AetraStakingPolicyStateSnapshotTop33Bps,
		AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate,
	}
}

func validateAetraStakingPolicyStateFields(group string, actual []string, required []string) (int, []string) {
	failed := make([]string, 0)
	requiredSet := map[string]bool{}
	for _, field := range required {
		requiredSet[field] = true
	}
	seen := map[string]bool{}
	for _, field := range actual {
		if field == "" {
			failed = append(failed, group+".field_name_required")
			continue
		}
		if seen[field] {
			failed = append(failed, group+"."+field+":duplicate")
			continue
		}
		seen[field] = true
		if !requiredSet[field] {
			failed = append(failed, group+"."+field+":unexpected")
		}
	}
	passed := 0
	for _, field := range required {
		if seen[field] {
			passed++
		} else {
			failed = append(failed, group+"."+field+":missing")
		}
	}
	return passed, failed
}

func DefaultAetraStakingPolicyParameterRuleSet() AetraStakingPolicyParameterRuleSet {
	return AetraStakingPolicyParameterRuleSet{
		ValidatorPowerCapBps: AetraStakingPolicyBpsRule{
			Name:		AetraStakingPolicyParamValidatorPowerCapBps,
			MinBps:		100,
			MaxBps:		500,
			RecommendedMin:	200,
			RecommendedMax:	300,
		},
		OverflowRewardMultiplierBps: AetraStakingPolicyBpsRule{
			Name:		AetraStakingPolicyParamOverflowRewardMultiplierBps,
			MinBps:		0,
			MaxBps:		10_000,
			RecommendedMin:	0,
			RecommendedMax:	3_000,
		},
		CommissionFloorBps: AetraStakingPolicyBpsRule{
			Name:		AetraStakingPolicyParamCommissionFloorBps,
			MinBps:		0,
			MaxBps:		1_000,
			RecommendedMin:	300,
			RecommendedMax:	500,
		},
		CommissionMaxBps: AetraStakingPolicyBpsRule{
			Name:		AetraStakingPolicyParamCommissionMaxBps,
			MinBps:		0,
			MaxBps:		3_000,
			RecommendedMin:	1_500,
			RecommendedMax:	2_000,
		},
		CommissionMaxDailyChangeBps: AetraStakingPolicyBpsRule{
			Name:		AetraStakingPolicyParamCommissionMaxDailyChangeBps,
			MinBps:		1,
			MaxBps:		500,
			RecommendedMin:	50,
			RecommendedMax:	100,
		},
	}
}

func DefaultAetraStakingPolicyParameterValues() AetraStakingPolicyParameterValues {
	return AetraStakingPolicyParameterValues{
		ValidatorPowerCapBps:		250,
		OverflowRewardMultiplierBps:	0,
		CommissionFloorBps:		300,
		CommissionMaxBps:		2_000,
		CommissionMaxDailyChangeBps:	100,
		Top10TargetBps:			2_500,
		Top20TargetBps:			4_000,
		Top33TargetBps:			5_000,
		MaxValidatorsSoftTarget:	200,
	}
}

func ValidateAetraStakingPolicyParameterValues(values AetraStakingPolicyParameterValues) error {
	report := BuildAetraStakingPolicyParameterReport(values)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy parameter rules failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyParameterReport(values AetraStakingPolicyParameterValues) AetraStakingPolicyParameterReport {
	rules := DefaultAetraStakingPolicyParameterRuleSet()
	failed := make([]string, 0)
	passed := 0

	for _, check := range []requirementCheck{
		{AetraStakingPolicyParamValidatorPowerCapBps, validateAetraStakingPolicyBps(values.ValidatorPowerCapBps, rules.ValidatorPowerCapBps.MinBps, rules.ValidatorPowerCapBps.MaxBps)},
		{AetraStakingPolicyParamOverflowRewardMultiplierBps, validateAetraStakingPolicyBps(values.OverflowRewardMultiplierBps, rules.OverflowRewardMultiplierBps.MinBps, rules.OverflowRewardMultiplierBps.MaxBps)},
		{AetraStakingPolicyParamCommissionFloorBps, validateAetraStakingPolicyBps(values.CommissionFloorBps, rules.CommissionFloorBps.MinBps, rules.CommissionFloorBps.MaxBps)},
		{AetraStakingPolicyParamCommissionMaxBps, validateAetraStakingPolicyBps(values.CommissionMaxBps, values.CommissionFloorBps, rules.CommissionMaxBps.MaxBps)},
		{AetraStakingPolicyParamCommissionMaxDailyChangeBps, validateAetraStakingPolicyBps(values.CommissionMaxDailyChangeBps, rules.CommissionMaxDailyChangeBps.MinBps, rules.CommissionMaxDailyChangeBps.MaxBps)},
		{AetraStakingPolicyParamTop10TargetBps, validateAetraStakingPolicyTopNTargets(values.Top10TargetBps, values.Top20TargetBps, values.Top33TargetBps)},
		{AetraStakingPolicyParamTop20TargetBps, validateAetraStakingPolicyTopNTargets(values.Top10TargetBps, values.Top20TargetBps, values.Top33TargetBps)},
		{AetraStakingPolicyParamTop33TargetBps, validateAetraStakingPolicyTopNTargets(values.Top10TargetBps, values.Top20TargetBps, values.Top33TargetBps)},
		{AetraStakingPolicyParamMaxValidatorsSoftTarget, values.MaxValidatorsSoftTarget > 0},
		{AetraStakingPolicyParamRejectNegativeOrOverflowMath, !hasNegativeAetraStakingPolicyParameter(values)},
	} {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicyParameterReport{
		Required:	10,
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func validateAetraStakingPolicyBps(value, minValue, maxValue int64) bool {
	return value >= minValue && value <= maxValue
}

func validateAetraStakingPolicyTopNTargets(top10, top20, top33 int64) bool {
	return top10 > 0 &&
		top10 <= top20 &&
		top20 <= top33 &&
		top33 <= 10_000
}

func hasNegativeAetraStakingPolicyParameter(values AetraStakingPolicyParameterValues) bool {
	return values.ValidatorPowerCapBps < 0 ||
		values.OverflowRewardMultiplierBps < 0 ||
		values.CommissionFloorBps < 0 ||
		values.CommissionMaxBps < 0 ||
		values.CommissionMaxDailyChangeBps < 0 ||
		values.Top10TargetBps < 0 ||
		values.Top20TargetBps < 0 ||
		values.Top33TargetBps < 0 ||
		values.MaxValidatorsSoftTarget < 0
}

func DefaultAetraStakingPolicyEffectivePowerStage1Evidence() AetraStakingPolicyEffectivePowerEvidence {
	return AetraStakingPolicyEffectivePowerEvidence{
		ModuleName:	AetraStakingPolicyModuleName,
		Stage:		AetraStakingPolicyEffectivePowerStage1,

		DefinesCapScope:		true,
		CapAffectsRewardCalculation:	true,
		CapAffectsDelegationWarnings:	true,
		CapAffectsCometBFTVotingPower:	false,
		Stage1LowConsensusRisk:		true,
	}
}

func DefaultAetraStakingPolicyEffectivePowerStage2Evidence() AetraStakingPolicyEffectivePowerEvidence {
	return AetraStakingPolicyEffectivePowerEvidence{
		ModuleName:	AetraStakingPolicyModuleName,
		Stage:		AetraStakingPolicyEffectivePowerStage2,

		DefinesCapScope:		true,
		CapAffectsRewardCalculation:	true,
		CapAffectsDelegationWarnings:	true,
		CapAffectsCometBFTVotingPower:	true,
		Stage2DeepIntegrationTests:	true,

		ValidatorUpdatesUseCappedPower:		true,
		TotalVotingPowerConsistent:		true,
		NoValidatorCanExceedCap:		true,
		DelegationUnbondingSharesCorrect:	true,
		SlashingUsesUnderlyingRawStake:		true,
		EvidenceHandlingCorrect:		true,
	}
}

func ValidateAetraStakingPolicyEffectivePowerEvidence(evidence AetraStakingPolicyEffectivePowerEvidence) error {
	report := BuildAetraStakingPolicyEffectivePowerReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy effective power rules failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyEffectivePowerReport(evidence AetraStakingPolicyEffectivePowerEvidence) AetraStakingPolicyEffectivePowerReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}
	if evidence.Stage != AetraStakingPolicyEffectivePowerStage1 && evidence.Stage != AetraStakingPolicyEffectivePowerStage2 {
		failed = append(failed, "effective_power_stage_unknown")
	}

	checks := []requirementCheck{
		{AetraStakingPolicyEffectivePowerDefinesCapScope, evidence.DefinesCapScope},
	}
	switch evidence.Stage {
	case AetraStakingPolicyEffectivePowerStage1:
		checks = append(checks,
			requirementCheck{AetraStakingPolicyEffectivePowerRewards, evidence.CapAffectsRewardCalculation},
			requirementCheck{AetraStakingPolicyEffectivePowerDelegationWarnings, evidence.CapAffectsDelegationWarnings},
			requirementCheck{AetraStakingPolicyEffectivePowerStage1LowConsensusRisk, evidence.Stage1LowConsensusRisk},
		)
		if evidence.CapAffectsCometBFTVotingPower {
			failed = append(failed, AetraStakingPolicyEffectivePowerCometBFTVotingPower+":stage_1_must_not_touch_cometbft_power")
		}
	case AetraStakingPolicyEffectivePowerStage2:
		checks = append(checks,
			requirementCheck{AetraStakingPolicyEffectivePowerCometBFTVotingPower, evidence.CapAffectsCometBFTVotingPower},
			requirementCheck{AetraStakingPolicyEffectivePowerStage2DeepIntegrationTests, evidence.Stage2DeepIntegrationTests},
			requirementCheck{AetraStakingPolicyEffectivePowerValidatorUpdatesCapped, evidence.ValidatorUpdatesUseCappedPower},
			requirementCheck{AetraStakingPolicyEffectivePowerTotalVotingConsistent, evidence.TotalVotingPowerConsistent},
			requirementCheck{AetraStakingPolicyEffectivePowerNoValidatorExceedsCap, evidence.NoValidatorCanExceedCap},
			requirementCheck{AetraStakingPolicyEffectivePowerSharesCorrect, evidence.DelegationUnbondingSharesCorrect},
			requirementCheck{AetraStakingPolicyEffectivePowerSlashingRawStake, evidence.SlashingUsesUnderlyingRawStake},
			requirementCheck{AetraStakingPolicyEffectivePowerEvidenceHandlingCorrect, evidence.EvidenceHandlingCorrect},
		)
	}

	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}
	sort.Strings(failed)
	return AetraStakingPolicyEffectivePowerReport{
		ModuleName:	evidence.ModuleName,
		Stage:		evidence.Stage,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraStakingPolicyMessageSpecEvidence() AetraStakingPolicyMessageSpecEvidence {
	return AetraStakingPolicyMessageSpecEvidence{
		ModuleName:	AetraStakingPolicyModuleName,
		GovernanceAuthorityMessages: []string{
			AetraStakingPolicyMessageMsgUpdateStakingPolicyParams,
			AetraStakingPolicyMessageMsgUpdateValidatorPowerCapSchedule,
			AetraStakingPolicyMessageMsgSetCommissionPolicy,
		},
		OptionalValidatorMessages: []string{
			AetraStakingPolicyMessageMsgRegisterValidatorIdentity,
			AetraStakingPolicyMessageMsgUpdateValidatorIdentity,
			AetraStakingPolicyMessageMsgAcknowledgeOverCapWarning,
		},
		ValidateAuthority:		true,
		ValidateSigner:			true,
		RejectMalformedAddresses:	true,
		RejectInvalidParams:		true,
		EmitEvents:			true,
		CoveredByTests:			true,
	}
}

func ValidateAetraStakingPolicyMessageSpec(evidence AetraStakingPolicyMessageSpecEvidence) error {
	report := BuildAetraStakingPolicyMessageSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy message spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyMessageSpecReport(evidence AetraStakingPolicyMessageSpecEvidence) AetraStakingPolicyMessageSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	requiredGovMessages := requiredAetraStakingPolicyGovernanceMessages()
	requiredValidatorMessages := optionalAetraStakingPolicyValidatorMessages()
	passedGov, failedGov := validateAetraStakingPolicyCatalog(AetraStakingPolicyMessageGovernanceOrAuthorityOnly, evidence.GovernanceAuthorityMessages, requiredGovMessages)
	passedValidator, failedValidator := validateAetraStakingPolicyCatalog(AetraStakingPolicyMessageOptionalValidatorMessages, evidence.OptionalValidatorMessages, requiredValidatorMessages)
	failed = append(failed, failedGov...)
	failed = append(failed, failedValidator...)

	checks := []requirementCheck{
		{AetraStakingPolicyMessageValidateAuthority, evidence.ValidateAuthority},
		{AetraStakingPolicyMessageValidateSigner, evidence.ValidateSigner},
		{AetraStakingPolicyMessageRejectMalformedAddresses, evidence.RejectMalformedAddresses},
		{AetraStakingPolicyMessageRejectInvalidParams, evidence.RejectInvalidParams},
		{AetraStakingPolicyMessageEmitEvents, evidence.EmitEvents},
		{AetraStakingPolicyMessageCoveredByTests, evidence.CoveredByTests},
	}
	passed := passedGov + passedValidator
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicyMessageSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredGovMessages) + len(requiredValidatorMessages) + len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraStakingPolicyQuerySpecEvidence() AetraStakingPolicyQuerySpecEvidence {
	return AetraStakingPolicyQuerySpecEvidence{
		ModuleName:	AetraStakingPolicyModuleName,
		RequiredQueries: []string{
			AetraStakingPolicyQueryParams,
			AetraStakingPolicyQueryValidatorPolicy,
			AetraStakingPolicyQueryValidatorEffectivePower,
			AetraStakingPolicyQueryValidatorOverflow,
			AetraStakingPolicyQueryTopNConcentration,
			AetraStakingPolicyQueryDelegationWarning,
			AetraStakingPolicyQueryCommissionPolicy,
			AetraStakingPolicyQueryConcentrationSnapshot,
			AetraStakingPolicyQueryNakamotoCoefficient,
		},
		StableResponses:		true,
		IndexerFriendlyResponses:	true,
	}
}

func ValidateAetraStakingPolicyQuerySpec(evidence AetraStakingPolicyQuerySpecEvidence) error {
	report := BuildAetraStakingPolicyQuerySpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy query spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyQuerySpecReport(evidence AetraStakingPolicyQuerySpecEvidence) AetraStakingPolicyQuerySpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	requiredQueries := requiredAetraStakingPolicyQueries()
	passed, queryFailures := validateAetraStakingPolicyCatalog("queries", evidence.RequiredQueries, requiredQueries)
	failed = append(failed, queryFailures...)
	checks := []requirementCheck{
		{AetraStakingPolicyQueryStableResponses, evidence.StableResponses},
		{AetraStakingPolicyQueryIndexerFriendlyResponses, evidence.IndexerFriendlyResponses},
	}
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicyQuerySpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredQueries) + len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraStakingPolicyGovernanceMessages() []string {
	return []string{
		AetraStakingPolicyMessageMsgUpdateStakingPolicyParams,
		AetraStakingPolicyMessageMsgUpdateValidatorPowerCapSchedule,
		AetraStakingPolicyMessageMsgSetCommissionPolicy,
	}
}

func optionalAetraStakingPolicyValidatorMessages() []string {
	return []string{
		AetraStakingPolicyMessageMsgRegisterValidatorIdentity,
		AetraStakingPolicyMessageMsgUpdateValidatorIdentity,
		AetraStakingPolicyMessageMsgAcknowledgeOverCapWarning,
	}
}

func requiredAetraStakingPolicyQueries() []string {
	return []string{
		AetraStakingPolicyQueryParams,
		AetraStakingPolicyQueryValidatorPolicy,
		AetraStakingPolicyQueryValidatorEffectivePower,
		AetraStakingPolicyQueryValidatorOverflow,
		AetraStakingPolicyQueryTopNConcentration,
		AetraStakingPolicyQueryDelegationWarning,
		AetraStakingPolicyQueryCommissionPolicy,
		AetraStakingPolicyQueryConcentrationSnapshot,
		AetraStakingPolicyQueryNakamotoCoefficient,
	}
}

func validateAetraStakingPolicyCatalog(group string, actual []string, required []string) (int, []string) {
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

func DefaultAetraStakingPolicyEventSpecEvidence() AetraStakingPolicyEventSpecEvidence {
	return AetraStakingPolicyEventSpecEvidence{
		ModuleName:	AetraStakingPolicyModuleName,
		RequiredEvents:	requiredAetraStakingPolicyEvents(),

		StableEventNames:		true,
		IndexerFriendlyAttributes:	true,
	}
}

func ValidateAetraStakingPolicyEventSpec(evidence AetraStakingPolicyEventSpecEvidence) error {
	report := BuildAetraStakingPolicyEventSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy event spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyEventSpecReport(evidence AetraStakingPolicyEventSpecEvidence) AetraStakingPolicyEventSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	requiredEvents := requiredAetraStakingPolicyEvents()
	passed, eventFailures := validateAetraStakingPolicyCatalog("events", evidence.RequiredEvents, requiredEvents)
	failed = append(failed, eventFailures...)
	for _, check := range []requirementCheck{
		{AetraStakingPolicyEventStableNames, evidence.StableEventNames},
		{AetraStakingPolicyEventIndexerFriendlyAttrs, evidence.IndexerFriendlyAttributes},
	} {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicyEventSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredEvents) + 2,
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraStakingPolicyInvariantSpecEvidence() AetraStakingPolicyInvariantSpecEvidence {
	return AetraStakingPolicyInvariantSpecEvidence{
		ModuleName:	AetraStakingPolicyModuleName,

		EffectivePowerNeverExceedsCap:		true,
		OverflowStakeNeverNegative:		true,
		RawStakeConservation:			true,
		CommissionWithinFloorAndMax:		true,
		CommissionChangeWithinDailyLimit:	true,
		TopNDoesNotExceedHundredPercent:	true,
		ExportImportPreservesPolicyState:	true,
		CoveredByTests:				true,
	}
}

func ValidateAetraStakingPolicyInvariantSpec(evidence AetraStakingPolicyInvariantSpecEvidence) error {
	report := BuildAetraStakingPolicyInvariantSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy invariant spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyInvariantSpecReport(evidence AetraStakingPolicyInvariantSpecEvidence) AetraStakingPolicyInvariantSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	checks := []requirementCheck{
		{AetraStakingPolicyInvariantEffectivePowerCap, evidence.EffectivePowerNeverExceedsCap},
		{AetraStakingPolicyInvariantOverflowNonNegative, evidence.OverflowStakeNeverNegative},
		{AetraStakingPolicyInvariantRawStakeConservation, evidence.RawStakeConservation},
		{AetraStakingPolicyInvariantCommissionBounds, evidence.CommissionWithinFloorAndMax},
		{AetraStakingPolicyInvariantCommissionDailyChange, evidence.CommissionChangeWithinDailyLimit},
		{AetraStakingPolicyInvariantTopNMaxHundredPercent, evidence.TopNDoesNotExceedHundredPercent},
		{AetraStakingPolicyInvariantExportImportPreserves, evidence.ExportImportPreservesPolicyState},
		{AetraStakingPolicyInvariantCoveredByTests, evidence.CoveredByTests},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicyInvariantSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraStakingPolicyEvents() []string {
	return []string{
		AetraStakingPolicyEventParamsUpdated,
		AetraStakingPolicyEventValidatorOverCap,
		AetraStakingPolicyEventValidatorBackUnderCap,
		AetraStakingPolicyEventCommissionRejected,
		AetraStakingPolicyEventConcentrationSnapshot,
		AetraStakingPolicyEventRewardMultiplierChanged,
	}
}

func DefaultAetraStakingPolicyTestSpecEvidence() AetraStakingPolicyTestSpecEvidence {
	return AetraStakingPolicyTestSpecEvidence{
		ModuleName:	AetraStakingPolicyModuleName,
		RequiredTests:	requiredAetraStakingPolicyTests(),
	}
}

func ValidateAetraStakingPolicyTestSpec(evidence AetraStakingPolicyTestSpecEvidence) error {
	report := BuildAetraStakingPolicyTestSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy test spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyTestSpecReport(evidence AetraStakingPolicyTestSpecEvidence) AetraStakingPolicyTestSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	requiredTests := requiredAetraStakingPolicyTests()
	passed, testFailures := validateAetraStakingPolicyCatalog("tests", evidence.RequiredTests, requiredTests)
	failed = append(failed, testFailures...)

	sort.Strings(failed)
	return AetraStakingPolicyTestSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredTests),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraStakingPolicyTests() []string {
	return []string{
		AetraStakingPolicyTestCapMath100Validators,
		AetraStakingPolicyTestCapMath150Validators,
		AetraStakingPolicyTestCapMath250Validators,
		AetraStakingPolicyTestCapMath300Validators,
		AetraStakingPolicyTestValidatorCrossingCapUpward,
		AetraStakingPolicyTestValidatorCrossingCapDownward,
		AetraStakingPolicyTestDelegationToOverCapValidator,
		AetraStakingPolicyTestRedelegationFromOverCapValidator,
		AetraStakingPolicyTestUnbondingFromOverCapValidator,
		AetraStakingPolicyTestSlashingOverCapValidator,
		AetraStakingPolicyTestCommissionBelowFloorRejected,
		AetraStakingPolicyTestCommissionAboveMaxRejected,
		AetraStakingPolicyTestCommissionDailyJumpRejected,
		AetraStakingPolicyTestGovernanceParamUpdateAccepted,
		AetraStakingPolicyTestGovernanceParamUpdateRejected,
		AetraStakingPolicyTestExportImportOverCapValidators,
		AetraStakingPolicyTestDeterministicConcentrationSnapshot,
	}
}
