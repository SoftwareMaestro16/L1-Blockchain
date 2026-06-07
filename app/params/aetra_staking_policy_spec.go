package params

import (
	"fmt"
	"sort"
)

const (
	AetraStakingPolicyModuleName = "x/aetra-staking-policy"

	AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration = "control_effective_voting_power_delegation_overflow_commission_policy_and_anti_concentration_incentives"
	AetraStakingPolicyCentralAntiCentralizationModule                          = "central_anti_centralization_module"

	AetraStakingPolicyResponsibilityRawStake                       = "calculate_raw_validator_stake"
	AetraStakingPolicyResponsibilityEffectiveStake                 = "calculate_effective_validator_stake"
	AetraStakingPolicyResponsibilityOverflowStake                  = "calculate_overflow_stake"
	AetraStakingPolicyResponsibilityEffectiveVotingPowerCap        = "enforce_or_expose_effective_voting_power_cap"
	AetraStakingPolicyResponsibilityOverflowRewardMultiplier       = "calculate_reward_multiplier_for_overflow_stake"
	AetraStakingPolicyResponsibilityDelegationConcentrationWarning = "expose_delegation_concentration_warnings"
	AetraStakingPolicyResponsibilityCommissionFloor                = "enforce_commission_floor"
	AetraStakingPolicyResponsibilityMaxCommission                  = "enforce_max_commission"
	AetraStakingPolicyResponsibilityMaxCommissionChangeRate        = "enforce_max_commission_change_rate"
	AetraStakingPolicyResponsibilityTopNConcentrationMetrics       = "expose_top_n_concentration_metrics"
	AetraStakingPolicyResponsibilityGovernanceParamValidation      = "validate_governance_param_changes"
	AetraStakingPolicyResponsibilityPolicyChangeEvents             = "emit_events_for_cap_overflow_commission_policy_changes"
	AetraStakingPolicyResponsibilityDeterministicExportImport      = "remain_deterministic_and_export_import_safe"

	AetraStakingPolicyStateParams                = "Params"
	AetraStakingPolicyStateValidatorPolicy       = "ValidatorPolicy"
	AetraStakingPolicyStateConcentrationSnapshot = "ConcentrationSnapshot"

	AetraStakingPolicyStateParamMaxValidatorsSoftTarget        = "MaxValidatorsSoftTarget"
	AetraStakingPolicyStateParamValidatorPowerCapBps           = "ValidatorPowerCapBps"
	AetraStakingPolicyStateParamValidatorPowerCapSchedule      = "ValidatorPowerCapSchedule"
	AetraStakingPolicyStateParamOverflowRewardMultiplierBps    = "OverflowRewardMultiplierBps"
	AetraStakingPolicyStateParamCommissionFloorBps             = "CommissionFloorBps"
	AetraStakingPolicyStateParamCommissionMaxBps               = "CommissionMaxBps"
	AetraStakingPolicyStateParamCommissionMaxDailyChangeBps    = "CommissionMaxDailyChangeBps"
	AetraStakingPolicyStateParamTop10TargetBps                 = "Top10TargetBps"
	AetraStakingPolicyStateParamTop20TargetBps                 = "Top20TargetBps"
	AetraStakingPolicyStateParamTop33TargetBps                 = "Top33TargetBps"
	AetraStakingPolicyStateParamMinSelfBond                    = "MinSelfBond"
	AetraStakingPolicyStateParamMinValidatorBond               = "MinValidatorBond"
	AetraStakingPolicyStateParamWarningThresholdBps            = "WarningThresholdBps"
	AetraStakingPolicyStateValidatorOperatorAddress            = "OperatorAddress"
	AetraStakingPolicyStateValidatorRawBondedTokens            = "RawBondedTokens"
	AetraStakingPolicyStateValidatorEffectiveBondedTokens      = "EffectiveBondedTokens"
	AetraStakingPolicyStateValidatorOverflowBondedTokens       = "OverflowBondedTokens"
	AetraStakingPolicyStateValidatorEffectivePowerBps          = "EffectivePowerBps"
	AetraStakingPolicyStateValidatorIsOverCap                  = "IsOverCap"
	AetraStakingPolicyStateValidatorRewardMultiplierBps        = "RewardMultiplierBps"
	AetraStakingPolicyStateValidatorLastCommissionChangeTime   = "LastCommissionChangeTime"
	AetraStakingPolicyStateValidatorLastCommissionRateBps      = "LastCommissionRateBps"
	AetraStakingPolicyStateSnapshotHeight                      = "Height"
	AetraStakingPolicyStateSnapshotBondedRatio                 = "BondedRatio"
	AetraStakingPolicyStateSnapshotActiveValidators            = "ActiveValidators"
	AetraStakingPolicyStateSnapshotTop10Bps                    = "Top10Bps"
	AetraStakingPolicyStateSnapshotTop20Bps                    = "Top20Bps"
	AetraStakingPolicyStateSnapshotTop33Bps                    = "Top33Bps"
	AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate = "NakamotoCoefficientEstimate"
	AetraStakingPolicyStateIntegerBpsOrSDKDecimal              = "integer_basis_points_or_sdk_decimal_accounting"
	AetraStakingPolicyStateNoFloatingPoint                     = "avoid_floating_point_accounting"

	AetraStakingPolicyParamValidatorPowerCapBps         = "ValidatorPowerCapBps"
	AetraStakingPolicyParamOverflowRewardMultiplierBps  = "OverflowRewardMultiplierBps"
	AetraStakingPolicyParamCommissionFloorBps           = "CommissionFloorBps"
	AetraStakingPolicyParamCommissionMaxBps             = "CommissionMaxBps"
	AetraStakingPolicyParamCommissionMaxDailyChangeBps  = "CommissionMaxDailyChangeBps"
	AetraStakingPolicyParamTop10TargetBps               = "Top10TargetBps"
	AetraStakingPolicyParamTop20TargetBps               = "Top20TargetBps"
	AetraStakingPolicyParamTop33TargetBps               = "Top33TargetBps"
	AetraStakingPolicyParamMaxValidatorsSoftTarget      = "MaxValidatorsSoftTarget"
	AetraStakingPolicyParamRejectNegativeOrOverflowMath = "reject_negative_or_overflowing_math_values"

	AetraStakingPolicyEffectivePowerStage1 = "stage_1_rewards_and_delegation_warnings"
	AetraStakingPolicyEffectivePowerStage2 = "stage_2_capped_validator_updates"

	AetraStakingPolicyEffectivePowerDefinesCapScope            = "define_whether_cap_affects_rewards_voting_power_or_both"
	AetraStakingPolicyEffectivePowerRewards                    = "cap_affects_reward_calculation"
	AetraStakingPolicyEffectivePowerDelegationWarnings         = "cap_affects_delegation_warnings"
	AetraStakingPolicyEffectivePowerCometBFTVotingPower        = "cap_affects_actual_cometbft_voting_power"
	AetraStakingPolicyEffectivePowerStage1LowConsensusRisk     = "stage_1_low_consensus_risk"
	AetraStakingPolicyEffectivePowerStage2DeepIntegrationTests = "stage_2_requires_deeper_integration_and_heavy_tests"
	AetraStakingPolicyEffectivePowerValidatorUpdatesCapped     = "validator_updates_sent_to_cometbft_use_capped_power"
	AetraStakingPolicyEffectivePowerTotalVotingConsistent      = "total_voting_power_remains_consistent"
	AetraStakingPolicyEffectivePowerNoValidatorExceedsCap      = "no_validator_can_exceed_cap"
	AetraStakingPolicyEffectivePowerSharesCorrect              = "delegation_and_unbonding_shares_remain_correct"
	AetraStakingPolicyEffectivePowerSlashingRawStake           = "slashing_can_still_slash_underlying_raw_stake"
	AetraStakingPolicyEffectivePowerEvidenceHandlingCorrect    = "evidence_handling_remains_correct"
)

type AetraStakingPolicySpecEvidence struct {
	ModuleName string

	PurposeEffectivePowerOverflowCommissionAntiConcentration bool
	CentralAntiCentralizationModule                          bool

	CalculatesRawValidatorStake              bool
	CalculatesEffectiveValidatorStake        bool
	CalculatesOverflowStake                  bool
	EnforcesOrExposesEffectiveVotingPowerCap bool
	CalculatesOverflowRewardMultiplier       bool
	ExposesDelegationConcentrationWarnings   bool
	EnforcesCommissionFloor                  bool
	EnforcesMaxCommission                    bool
	EnforcesMaxCommissionChangeRate          bool
	ExposesTopNConcentrationMetrics          bool
	ValidatesGovernanceParamChanges          bool
	EmitsCapOverflowCommissionPolicyEvents   bool
	RemainsDeterministicAndExportImportSafe  bool
}

type AetraStakingPolicySpecReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

type AetraStakingPolicyStateSpecEvidence struct {
	ModuleName string

	ParamsFields                []string
	ValidatorPolicyFields       []string
	ConcentrationSnapshotFields []string

	IntegerBasisPointsOrSDKDecimals bool
	AvoidsFloatingPoint             bool
}

type AetraStakingPolicyStateSpecReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

type AetraStakingPolicyBpsRule struct {
	Name           string
	MinBps         int64
	MaxBps         int64
	RecommendedMin int64
	RecommendedMax int64
}

type AetraStakingPolicyParameterRuleSet struct {
	ValidatorPowerCapBps        AetraStakingPolicyBpsRule
	OverflowRewardMultiplierBps AetraStakingPolicyBpsRule
	CommissionFloorBps          AetraStakingPolicyBpsRule
	CommissionMaxBps            AetraStakingPolicyBpsRule
	CommissionMaxDailyChangeBps AetraStakingPolicyBpsRule
}

type AetraStakingPolicyParameterValues struct {
	ValidatorPowerCapBps        int64
	OverflowRewardMultiplierBps int64
	CommissionFloorBps          int64
	CommissionMaxBps            int64
	CommissionMaxDailyChangeBps int64
	Top10TargetBps              int64
	Top20TargetBps              int64
	Top33TargetBps              int64
	MaxValidatorsSoftTarget     int64
}

type AetraStakingPolicyParameterReport struct {
	Required int
	Passed   int
	Failed   []string
	Ready    bool
}

type AetraStakingPolicyEffectivePowerEvidence struct {
	ModuleName string
	Stage      string

	DefinesCapScope               bool
	CapAffectsRewardCalculation   bool
	CapAffectsDelegationWarnings  bool
	CapAffectsCometBFTVotingPower bool
	Stage1LowConsensusRisk        bool
	Stage2DeepIntegrationTests    bool

	ValidatorUpdatesUseCappedPower   bool
	TotalVotingPowerConsistent       bool
	NoValidatorCanExceedCap          bool
	DelegationUnbondingSharesCorrect bool
	SlashingUsesUnderlyingRawStake   bool
	EvidenceHandlingCorrect          bool
}

type AetraStakingPolicyEffectivePowerReport struct {
	ModuleName string
	Stage      string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

func DefaultAetraStakingPolicySpecEvidence() AetraStakingPolicySpecEvidence {
	return AetraStakingPolicySpecEvidence{
		ModuleName: AetraStakingPolicyModuleName,

		PurposeEffectivePowerOverflowCommissionAntiConcentration: true,
		CentralAntiCentralizationModule:                          true,

		CalculatesRawValidatorStake:              true,
		CalculatesEffectiveValidatorStake:        true,
		CalculatesOverflowStake:                  true,
		EnforcesOrExposesEffectiveVotingPowerCap: true,
		CalculatesOverflowRewardMultiplier:       true,
		ExposesDelegationConcentrationWarnings:   true,
		EnforcesCommissionFloor:                  true,
		EnforcesMaxCommission:                    true,
		EnforcesMaxCommissionChangeRate:          true,
		ExposesTopNConcentrationMetrics:          true,
		ValidatesGovernanceParamChanges:          true,
		EmitsCapOverflowCommissionPolicyEvents:   true,
		RemainsDeterministicAndExportImportSafe:  true,
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
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func DefaultAetraStakingPolicyStateSpecEvidence() AetraStakingPolicyStateSpecEvidence {
	return AetraStakingPolicyStateSpecEvidence{
		ModuleName: AetraStakingPolicyModuleName,
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
		IntegerBasisPointsOrSDKDecimals: true,
		AvoidsFloatingPoint:             true,
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
		ModuleName: evidence.ModuleName,
		Required:   len(requiredParams) + len(requiredValidator) + len(requiredSnapshot) + 2,
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
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
			Name:           AetraStakingPolicyParamValidatorPowerCapBps,
			MinBps:         100,
			MaxBps:         500,
			RecommendedMin: 200,
			RecommendedMax: 300,
		},
		OverflowRewardMultiplierBps: AetraStakingPolicyBpsRule{
			Name:           AetraStakingPolicyParamOverflowRewardMultiplierBps,
			MinBps:         0,
			MaxBps:         10_000,
			RecommendedMin: 0,
			RecommendedMax: 3_000,
		},
		CommissionFloorBps: AetraStakingPolicyBpsRule{
			Name:           AetraStakingPolicyParamCommissionFloorBps,
			MinBps:         0,
			MaxBps:         1_000,
			RecommendedMin: 300,
			RecommendedMax: 500,
		},
		CommissionMaxBps: AetraStakingPolicyBpsRule{
			Name:           AetraStakingPolicyParamCommissionMaxBps,
			MinBps:         0,
			MaxBps:         3_000,
			RecommendedMin: 1_500,
			RecommendedMax: 2_000,
		},
		CommissionMaxDailyChangeBps: AetraStakingPolicyBpsRule{
			Name:           AetraStakingPolicyParamCommissionMaxDailyChangeBps,
			MinBps:         1,
			MaxBps:         500,
			RecommendedMin: 50,
			RecommendedMax: 100,
		},
	}
}

func DefaultAetraStakingPolicyParameterValues() AetraStakingPolicyParameterValues {
	return AetraStakingPolicyParameterValues{
		ValidatorPowerCapBps:        250,
		OverflowRewardMultiplierBps: 0,
		CommissionFloorBps:          300,
		CommissionMaxBps:            2_000,
		CommissionMaxDailyChangeBps: 100,
		Top10TargetBps:              2_500,
		Top20TargetBps:              4_000,
		Top33TargetBps:              5_000,
		MaxValidatorsSoftTarget:     200,
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
		Required: 10,
		Passed:   passed,
		Failed:   failed,
		Ready:    len(failed) == 0,
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
		ModuleName: AetraStakingPolicyModuleName,
		Stage:      AetraStakingPolicyEffectivePowerStage1,

		DefinesCapScope:               true,
		CapAffectsRewardCalculation:   true,
		CapAffectsDelegationWarnings:  true,
		CapAffectsCometBFTVotingPower: false,
		Stage1LowConsensusRisk:        true,
	}
}

func DefaultAetraStakingPolicyEffectivePowerStage2Evidence() AetraStakingPolicyEffectivePowerEvidence {
	return AetraStakingPolicyEffectivePowerEvidence{
		ModuleName: AetraStakingPolicyModuleName,
		Stage:      AetraStakingPolicyEffectivePowerStage2,

		DefinesCapScope:               true,
		CapAffectsRewardCalculation:   true,
		CapAffectsDelegationWarnings:  true,
		CapAffectsCometBFTVotingPower: true,
		Stage2DeepIntegrationTests:    true,

		ValidatorUpdatesUseCappedPower:   true,
		TotalVotingPowerConsistent:       true,
		NoValidatorCanExceedCap:          true,
		DelegationUnbondingSharesCorrect: true,
		SlashingUsesUnderlyingRawStake:   true,
		EvidenceHandlingCorrect:          true,
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
		ModuleName: evidence.ModuleName,
		Stage:      evidence.Stage,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}
