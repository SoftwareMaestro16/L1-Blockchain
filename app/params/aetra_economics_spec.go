package params

import (
	"fmt"
	"sort"
)

const (
	AetraEconomicsModuleName	= "x/aetra-economics"

	AetraEconomicsPurposeLowModerateInflation	= "low_moderate_inflation"
	AetraEconomicsPurposeFeeBurn			= "fee_burn"
	AetraEconomicsPurposeTreasuryAllocation		= "treasury_allocation"
	AetraEconomicsPurposeRewardSmoothing		= "reward_smoothing"
	AetraEconomicsPurposeTransparentAPRModel	= "transparent_apr_model"

	AetraEconomicsResponsibilityDynamicInflation	= "calculate_dynamic_inflation"
	AetraEconomicsResponsibilityBondedRatio		= "track_bonded_ratio"
	AetraEconomicsResponsibilityStakingAPR		= "estimate_staking_apr"
	AetraEconomicsResponsibilityFeeSplit		= "split_fees"
	AetraEconomicsResponsibilityBurnFeeShare	= "burn_configured_fee_share"
	AetraEconomicsResponsibilityRewardsShare	= "send_configured_share_to_distribution_rewards"
	AetraEconomicsResponsibilityTreasuryShare	= "send_configured_share_to_treasury"
	AetraEconomicsResponsibilityRewardSmoothing	= "smooth_reward_changes"
	AetraEconomicsResponsibilityEconomicMetrics	= "expose_economic_metrics"
	AetraEconomicsResponsibilitySupplyInvariants	= "protect_supply_invariants"

	AetraEconomicsStateParams		= "Params"
	AetraEconomicsStateEpochEconomics	= "EpochEconomics"
	AetraEconomicsStateSupplyStats		= "SupplyStats"

	AetraEconomicsStateParamInflationMinBps		= "InflationMinBps"
	AetraEconomicsStateParamInflationMaxBps		= "InflationMaxBps"
	AetraEconomicsStateParamInflationChangeRateBps	= "InflationChangeRateBps"
	AetraEconomicsStateParamTargetBondedRatioBps	= "TargetBondedRatioBps"
	AetraEconomicsStateParamBurnFeeShareBps		= "BurnFeeShareBps"
	AetraEconomicsStateParamRewardFeeShareBps	= "RewardFeeShareBps"
	AetraEconomicsStateParamTreasuryFeeShareBps	= "TreasuryFeeShareBps"
	AetraEconomicsStateParamRewardSmoothingEpochs	= "RewardSmoothingEpochs"
	AetraEconomicsStateParamAprTargetMinBps		= "AprTargetMinBps"
	AetraEconomicsStateParamAprTargetMaxBps		= "AprTargetMaxBps"

	AetraEconomicsStateEpochNumber		= "EpochNumber"
	AetraEconomicsStateEpochStartHeight	= "StartHeight"
	AetraEconomicsStateEpochEndHeight	= "EndHeight"
	AetraEconomicsStateEpochBondedRatioBps	= "BondedRatioBps"
	AetraEconomicsStateEpochInflationBps	= "InflationBps"
	AetraEconomicsStateEpochEstimatedAprBps	= "EstimatedAprBps"
	AetraEconomicsStateEpochFeesCollected	= "FeesCollected"
	AetraEconomicsStateEpochFeesBurned	= "FeesBurned"
	AetraEconomicsStateEpochFeesToRewards	= "FeesToRewards"
	AetraEconomicsStateEpochFeesToTreasury	= "FeesToTreasury"
	AetraEconomicsStateEpochMintedRewards	= "MintedRewards"

	AetraEconomicsStateSupplyTotalMinted	= "TotalMinted"
	AetraEconomicsStateSupplyTotalBurned	= "TotalBurned"
	AetraEconomicsStateSupplyNetIssuance	= "NetIssuance"

	AetraEconomicsInflationCurveBelowTargetIncreases	= "bonded_ratio_below_target_increases_inflation"
	AetraEconomicsInflationCurveAboveTargetDecreases	= "bonded_ratio_above_target_decreases_inflation"
	AetraEconomicsInflationCurveNeverBelowMin		= "inflation_never_below_min"
	AetraEconomicsInflationCurveNeverAboveMax		= "inflation_never_above_max"
	AetraEconomicsInflationCurveEpochChangeBounded		= "inflation_change_per_epoch_bounded"
	AetraEconomicsInflationCurveNoFloatingPoint		= "no_floating_point"
	AetraEconomicsInflationCurveNoPerBlockInstability	= "no_per_block_instability"
	AetraEconomicsInflationCurveDeterministic		= "all_calculations_deterministic"

	AetraEconomicsFeeSplitSumToBasisPoints			= "fee_split_sums_to_10000_bps"
	AetraEconomicsFeeSplitRecommendedBurnRange		= "burn_fee_share_bps_3000_6000"
	AetraEconomicsFeeSplitRecommendedRewardRange		= "reward_fee_share_bps_2000_4000"
	AetraEconomicsFeeSplitRecommendedTreasuryRange		= "treasury_fee_share_bps_1000_2000"
	AetraEconomicsFeeSplitRejectsInvalidSum			= "reject_sum_not_10000_bps"
	AetraEconomicsFeeSplitRejectsNegativeShares		= "reject_negative_shares"
	AetraEconomicsFeeSplitRejectsBurnAboveGovernanceMax	= "reject_burn_share_above_governance_max"
	AetraEconomicsFeeSplitRejectsTreasuryAboveMax		= "reject_treasury_share_above_governance_max"
	AetraEconomicsFeeSplitRejectsZeroRewards		= "reject_zero_rewards_without_emergency_governance"

	AetraEconomicsAPRQueryInflationOnly	= "inflation_only_apr"
	AetraEconomicsAPRQueryFeeAdjusted	= "fee_adjusted_apr"
	AetraEconomicsAPRQueryCommissionImpact	= "validator_commission_impact"
	AetraEconomicsAPRQueryDelegatorEstimate	= "estimated_delegator_apr"
	AetraEconomicsAPRQueryValidatorGross	= "estimated_validator_gross_apr"
	AetraEconomicsAPRQueryValidatorNet	= "estimated_validator_net_apr"
	AetraEconomicsAPRQueryLabeledAsEstimate	= "apr_labeled_as_estimate_not_guaranteed"

	AetraEconomicsRequiredTestInflationBelowTarget		= "inflation_increases_when_bonded_ratio_below_target"
	AetraEconomicsRequiredTestInflationAboveTarget		= "inflation_decreases_when_bonded_ratio_above_target"
	AetraEconomicsRequiredTestInflationMinMax		= "inflation_remains_within_min_max"
	AetraEconomicsRequiredTestInflationChangeBounded	= "inflation_change_rate_bounded"
	AetraEconomicsRequiredTestFeeSplitAccounting		= "fee_split_exact_accounting"
	AetraEconomicsRequiredTestBurnAccounting		= "burn_accounting"
	AetraEconomicsRequiredTestTreasuryAccounting		= "treasury_accounting"
	AetraEconomicsRequiredTestRewardsAccounting		= "rewards_accounting"
	AetraEconomicsRequiredTestAPRMath			= "apr_estimate_math"
	AetraEconomicsRequiredTestZeroFeeBlock			= "zero_fee_block_handling"
	AetraEconomicsRequiredTestHighFeeBlock			= "high_fee_block_handling"
	AetraEconomicsRequiredTestExportImportState		= "export_import_economics_state"
	AetraEconomicsRequiredTestSupplyInvariantManyEpochs	= "supply_invariant_after_many_epochs"
	AetraEconomicsRequiredTestGovernanceInvalidParams	= "governance_invalid_params_rejected"
)

type AetraEconomicsSpecEvidence struct {
	ModuleName	string

	LowModerateInflation	bool
	FeeBurn			bool
	TreasuryAllocation	bool
	RewardSmoothing		bool
	TransparentAPRModel	bool
}

type AetraEconomicsSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraEconomicsResponsibilitiesEvidence struct {
	ModuleName	string

	CalculatesDynamicInflation			bool
	TracksBondedRatio				bool
	EstimatesStakingAPR				bool
	SplitsFees					bool
	BurnsConfiguredFeeShare				bool
	SendsConfiguredShareToDistributionRewards	bool
	SendsConfiguredShareToTreasury			bool
	SmoothsRewardChanges				bool
	ExposesEconomicMetrics				bool
	ProtectsSupplyInvariants			bool
}

type AetraEconomicsResponsibilitiesReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraEconomicsStateSpecEvidence struct {
	ModuleName	string

	ParamsFields		[]string
	EpochEconomicsFields	[]string
	SupplyStatsFields	[]string
}

type AetraEconomicsStateSpecReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraEconomicsInflationCurveEvidence struct {
	ModuleName	string

	BondedRatioBelowTargetIncreasesInflation	bool
	BondedRatioAboveTargetDecreasesInflation	bool
	InflationNeverBelowMin				bool
	InflationNeverAboveMax				bool
	InflationChangePerEpochBounded			bool
	NoFloatingPoint					bool
	NoPerBlockInstability				bool
	AllCalculationsDeterministic			bool
}

type AetraEconomicsInflationCurveReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraEconomicsFeeSplitRulesEvidence struct {
	ModuleName	string

	FeeSplitSumsToBasisPoints		bool
	RecommendedBurnRange			bool
	RecommendedRewardRange			bool
	RecommendedTreasuryRange		bool
	RejectsInvalidSum			bool
	RejectsNegativeShares			bool
	RejectsBurnAboveGovernanceMax		bool
	RejectsTreasuryAboveGovernanceMax	bool
	RejectsZeroRewardsWithoutEmergency	bool
}

type AetraEconomicsFeeSplitRulesReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraEconomicsAPRQueryEvidence struct {
	ModuleName	string

	InflationOnlyAPR		bool
	FeeAdjustedAPR			bool
	ValidatorCommissionImpact	bool
	EstimatedDelegatorAPR		bool
	EstimatedValidatorGrossAPR	bool
	EstimatedValidatorNetAPR	bool
	LabeledAsEstimate		bool
}

type AetraEconomicsAPRQueryReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraEconomicsTestingRequirementsEvidence struct {
	ModuleName	string

	InflationIncreasesBelowTarget	bool
	InflationDecreasesAboveTarget	bool
	InflationWithinMinMax		bool
	InflationChangeRateBounded	bool
	FeeSplitExactAccounting		bool
	BurnAccounting			bool
	TreasuryAccounting		bool
	RewardsAccounting		bool
	APRMath				bool
	ZeroFeeBlockHandling		bool
	HighFeeBlockHandling		bool
	ExportImportEconomicsState	bool
	SupplyInvariantAfterManyEpochs	bool
	GovernanceInvalidParamsRejected	bool
}

type AetraEconomicsTestingRequirementsReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraEconomicsSpecEvidence() AetraEconomicsSpecEvidence {
	return AetraEconomicsSpecEvidence{
		ModuleName:	AetraEconomicsModuleName,

		LowModerateInflation:	true,
		FeeBurn:		true,
		TreasuryAllocation:	true,
		RewardSmoothing:	true,
		TransparentAPRModel:	true,
	}
}

func ValidateAetraEconomicsSpec(evidence AetraEconomicsSpecEvidence) error {
	report := BuildAetraEconomicsSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsSpecReport(evidence AetraEconomicsSpecEvidence) AetraEconomicsSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsPurposeLowModerateInflation, evidence.LowModerateInflation},
		{AetraEconomicsPurposeFeeBurn, evidence.FeeBurn},
		{AetraEconomicsPurposeTreasuryAllocation, evidence.TreasuryAllocation},
		{AetraEconomicsPurposeRewardSmoothing, evidence.RewardSmoothing},
		{AetraEconomicsPurposeTransparentAPRModel, evidence.TransparentAPRModel},
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
	return AetraEconomicsSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraEconomicsResponsibilitiesEvidence() AetraEconomicsResponsibilitiesEvidence {
	return AetraEconomicsResponsibilitiesEvidence{
		ModuleName:	AetraEconomicsModuleName,

		CalculatesDynamicInflation:			true,
		TracksBondedRatio:				true,
		EstimatesStakingAPR:				true,
		SplitsFees:					true,
		BurnsConfiguredFeeShare:			true,
		SendsConfiguredShareToDistributionRewards:	true,
		SendsConfiguredShareToTreasury:			true,
		SmoothsRewardChanges:				true,
		ExposesEconomicMetrics:				true,
		ProtectsSupplyInvariants:			true,
	}
}

func ValidateAetraEconomicsResponsibilities(evidence AetraEconomicsResponsibilitiesEvidence) error {
	report := BuildAetraEconomicsResponsibilitiesReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics responsibilities failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsResponsibilitiesReport(evidence AetraEconomicsResponsibilitiesEvidence) AetraEconomicsResponsibilitiesReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsResponsibilityDynamicInflation, evidence.CalculatesDynamicInflation},
		{AetraEconomicsResponsibilityBondedRatio, evidence.TracksBondedRatio},
		{AetraEconomicsResponsibilityStakingAPR, evidence.EstimatesStakingAPR},
		{AetraEconomicsResponsibilityFeeSplit, evidence.SplitsFees},
		{AetraEconomicsResponsibilityBurnFeeShare, evidence.BurnsConfiguredFeeShare},
		{AetraEconomicsResponsibilityRewardsShare, evidence.SendsConfiguredShareToDistributionRewards},
		{AetraEconomicsResponsibilityTreasuryShare, evidence.SendsConfiguredShareToTreasury},
		{AetraEconomicsResponsibilityRewardSmoothing, evidence.SmoothsRewardChanges},
		{AetraEconomicsResponsibilityEconomicMetrics, evidence.ExposesEconomicMetrics},
		{AetraEconomicsResponsibilitySupplyInvariants, evidence.ProtectsSupplyInvariants},
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
	return AetraEconomicsResponsibilitiesReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraEconomicsStateSpecEvidence() AetraEconomicsStateSpecEvidence {
	return AetraEconomicsStateSpecEvidence{
		ModuleName:	AetraEconomicsModuleName,
		ParamsFields: []string{
			AetraEconomicsStateParamInflationMinBps,
			AetraEconomicsStateParamInflationMaxBps,
			AetraEconomicsStateParamInflationChangeRateBps,
			AetraEconomicsStateParamTargetBondedRatioBps,
			AetraEconomicsStateParamBurnFeeShareBps,
			AetraEconomicsStateParamRewardFeeShareBps,
			AetraEconomicsStateParamTreasuryFeeShareBps,
			AetraEconomicsStateParamRewardSmoothingEpochs,
			AetraEconomicsStateParamAprTargetMinBps,
			AetraEconomicsStateParamAprTargetMaxBps,
		},
		EpochEconomicsFields: []string{
			AetraEconomicsStateEpochNumber,
			AetraEconomicsStateEpochStartHeight,
			AetraEconomicsStateEpochEndHeight,
			AetraEconomicsStateEpochBondedRatioBps,
			AetraEconomicsStateEpochInflationBps,
			AetraEconomicsStateEpochEstimatedAprBps,
			AetraEconomicsStateEpochFeesCollected,
			AetraEconomicsStateEpochFeesBurned,
			AetraEconomicsStateEpochFeesToRewards,
			AetraEconomicsStateEpochFeesToTreasury,
			AetraEconomicsStateEpochMintedRewards,
		},
		SupplyStatsFields: []string{
			AetraEconomicsStateSupplyTotalMinted,
			AetraEconomicsStateSupplyTotalBurned,
			AetraEconomicsStateSupplyNetIssuance,
		},
	}
}

func ValidateAetraEconomicsStateSpec(evidence AetraEconomicsStateSpecEvidence) error {
	report := BuildAetraEconomicsStateSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics state spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsStateSpecReport(evidence AetraEconomicsStateSpecEvidence) AetraEconomicsStateSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	requiredParams := requiredAetraEconomicsParamsFields()
	requiredEpoch := requiredAetraEconomicsEpochEconomicsFields()
	requiredSupply := requiredAetraEconomicsSupplyStatsFields()

	passedParams, failedParams := validateAetraEconomicsCatalog(AetraEconomicsStateParams, evidence.ParamsFields, requiredParams)
	passedEpoch, failedEpoch := validateAetraEconomicsCatalog(AetraEconomicsStateEpochEconomics, evidence.EpochEconomicsFields, requiredEpoch)
	passedSupply, failedSupply := validateAetraEconomicsCatalog(AetraEconomicsStateSupplyStats, evidence.SupplyStatsFields, requiredSupply)

	failed = append(failed, failedParams...)
	failed = append(failed, failedEpoch...)
	failed = append(failed, failedSupply...)
	sort.Strings(failed)
	return AetraEconomicsStateSpecReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredParams) + len(requiredEpoch) + len(requiredSupply),
		Passed:		passedParams + passedEpoch + passedSupply,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraEconomicsInflationCurveEvidence() AetraEconomicsInflationCurveEvidence {
	return AetraEconomicsInflationCurveEvidence{
		ModuleName:	AetraEconomicsModuleName,

		BondedRatioBelowTargetIncreasesInflation:	true,
		BondedRatioAboveTargetDecreasesInflation:	true,
		InflationNeverBelowMin:				true,
		InflationNeverAboveMax:				true,
		InflationChangePerEpochBounded:			true,
		NoFloatingPoint:				true,
		NoPerBlockInstability:				true,
		AllCalculationsDeterministic:			true,
	}
}

func ValidateAetraEconomicsInflationCurve(evidence AetraEconomicsInflationCurveEvidence) error {
	report := BuildAetraEconomicsInflationCurveReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics inflation curve failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsInflationCurveReport(evidence AetraEconomicsInflationCurveEvidence) AetraEconomicsInflationCurveReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsInflationCurveBelowTargetIncreases, evidence.BondedRatioBelowTargetIncreasesInflation},
		{AetraEconomicsInflationCurveAboveTargetDecreases, evidence.BondedRatioAboveTargetDecreasesInflation},
		{AetraEconomicsInflationCurveNeverBelowMin, evidence.InflationNeverBelowMin},
		{AetraEconomicsInflationCurveNeverAboveMax, evidence.InflationNeverAboveMax},
		{AetraEconomicsInflationCurveEpochChangeBounded, evidence.InflationChangePerEpochBounded},
		{AetraEconomicsInflationCurveNoFloatingPoint, evidence.NoFloatingPoint},
		{AetraEconomicsInflationCurveNoPerBlockInstability, evidence.NoPerBlockInstability},
		{AetraEconomicsInflationCurveDeterministic, evidence.AllCalculationsDeterministic},
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
	return AetraEconomicsInflationCurveReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraEconomicsFeeSplitRulesEvidence() AetraEconomicsFeeSplitRulesEvidence {
	return AetraEconomicsFeeSplitRulesEvidence{
		ModuleName:	AetraEconomicsModuleName,

		FeeSplitSumsToBasisPoints:		true,
		RecommendedBurnRange:			true,
		RecommendedRewardRange:			true,
		RecommendedTreasuryRange:		true,
		RejectsInvalidSum:			true,
		RejectsNegativeShares:			true,
		RejectsBurnAboveGovernanceMax:		true,
		RejectsTreasuryAboveGovernanceMax:	true,
		RejectsZeroRewardsWithoutEmergency:	true,
	}
}

func ValidateAetraEconomicsFeeSplitRules(evidence AetraEconomicsFeeSplitRulesEvidence) error {
	report := BuildAetraEconomicsFeeSplitRulesReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics fee split rules failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsFeeSplitRulesReport(evidence AetraEconomicsFeeSplitRulesEvidence) AetraEconomicsFeeSplitRulesReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsFeeSplitSumToBasisPoints, evidence.FeeSplitSumsToBasisPoints},
		{AetraEconomicsFeeSplitRecommendedBurnRange, evidence.RecommendedBurnRange},
		{AetraEconomicsFeeSplitRecommendedRewardRange, evidence.RecommendedRewardRange},
		{AetraEconomicsFeeSplitRecommendedTreasuryRange, evidence.RecommendedTreasuryRange},
		{AetraEconomicsFeeSplitRejectsInvalidSum, evidence.RejectsInvalidSum},
		{AetraEconomicsFeeSplitRejectsNegativeShares, evidence.RejectsNegativeShares},
		{AetraEconomicsFeeSplitRejectsBurnAboveGovernanceMax, evidence.RejectsBurnAboveGovernanceMax},
		{AetraEconomicsFeeSplitRejectsTreasuryAboveMax, evidence.RejectsTreasuryAboveGovernanceMax},
		{AetraEconomicsFeeSplitRejectsZeroRewards, evidence.RejectsZeroRewardsWithoutEmergency},
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
	return AetraEconomicsFeeSplitRulesReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraEconomicsAPRQueryEvidence() AetraEconomicsAPRQueryEvidence {
	return AetraEconomicsAPRQueryEvidence{
		ModuleName:	AetraEconomicsModuleName,

		InflationOnlyAPR:		true,
		FeeAdjustedAPR:			true,
		ValidatorCommissionImpact:	true,
		EstimatedDelegatorAPR:		true,
		EstimatedValidatorGrossAPR:	true,
		EstimatedValidatorNetAPR:	true,
		LabeledAsEstimate:		true,
	}
}

func ValidateAetraEconomicsAPRQuery(evidence AetraEconomicsAPRQueryEvidence) error {
	report := BuildAetraEconomicsAPRQueryReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics apr query failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsAPRQueryReport(evidence AetraEconomicsAPRQueryEvidence) AetraEconomicsAPRQueryReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsAPRQueryInflationOnly, evidence.InflationOnlyAPR},
		{AetraEconomicsAPRQueryFeeAdjusted, evidence.FeeAdjustedAPR},
		{AetraEconomicsAPRQueryCommissionImpact, evidence.ValidatorCommissionImpact},
		{AetraEconomicsAPRQueryDelegatorEstimate, evidence.EstimatedDelegatorAPR},
		{AetraEconomicsAPRQueryValidatorGross, evidence.EstimatedValidatorGrossAPR},
		{AetraEconomicsAPRQueryValidatorNet, evidence.EstimatedValidatorNetAPR},
		{AetraEconomicsAPRQueryLabeledAsEstimate, evidence.LabeledAsEstimate},
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
	return AetraEconomicsAPRQueryReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraEconomicsTestingRequirementsEvidence() AetraEconomicsTestingRequirementsEvidence {
	return AetraEconomicsTestingRequirementsEvidence{
		ModuleName:	AetraEconomicsModuleName,

		InflationIncreasesBelowTarget:		true,
		InflationDecreasesAboveTarget:		true,
		InflationWithinMinMax:			true,
		InflationChangeRateBounded:		true,
		FeeSplitExactAccounting:		true,
		BurnAccounting:				true,
		TreasuryAccounting:			true,
		RewardsAccounting:			true,
		APRMath:				true,
		ZeroFeeBlockHandling:			true,
		HighFeeBlockHandling:			true,
		ExportImportEconomicsState:		true,
		SupplyInvariantAfterManyEpochs:		true,
		GovernanceInvalidParamsRejected:	true,
	}
}

func ValidateAetraEconomicsTestingRequirements(evidence AetraEconomicsTestingRequirementsEvidence) error {
	report := BuildAetraEconomicsTestingRequirementsReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics testing requirements failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsTestingRequirementsReport(evidence AetraEconomicsTestingRequirementsEvidence) AetraEconomicsTestingRequirementsReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsRequiredTestInflationBelowTarget, evidence.InflationIncreasesBelowTarget},
		{AetraEconomicsRequiredTestInflationAboveTarget, evidence.InflationDecreasesAboveTarget},
		{AetraEconomicsRequiredTestInflationMinMax, evidence.InflationWithinMinMax},
		{AetraEconomicsRequiredTestInflationChangeBounded, evidence.InflationChangeRateBounded},
		{AetraEconomicsRequiredTestFeeSplitAccounting, evidence.FeeSplitExactAccounting},
		{AetraEconomicsRequiredTestBurnAccounting, evidence.BurnAccounting},
		{AetraEconomicsRequiredTestTreasuryAccounting, evidence.TreasuryAccounting},
		{AetraEconomicsRequiredTestRewardsAccounting, evidence.RewardsAccounting},
		{AetraEconomicsRequiredTestAPRMath, evidence.APRMath},
		{AetraEconomicsRequiredTestZeroFeeBlock, evidence.ZeroFeeBlockHandling},
		{AetraEconomicsRequiredTestHighFeeBlock, evidence.HighFeeBlockHandling},
		{AetraEconomicsRequiredTestExportImportState, evidence.ExportImportEconomicsState},
		{AetraEconomicsRequiredTestSupplyInvariantManyEpochs, evidence.SupplyInvariantAfterManyEpochs},
		{AetraEconomicsRequiredTestGovernanceInvalidParams, evidence.GovernanceInvalidParamsRejected},
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
	return AetraEconomicsTestingRequirementsReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraEconomicsParamsFields() []string {
	return []string{
		AetraEconomicsStateParamInflationMinBps,
		AetraEconomicsStateParamInflationMaxBps,
		AetraEconomicsStateParamInflationChangeRateBps,
		AetraEconomicsStateParamTargetBondedRatioBps,
		AetraEconomicsStateParamBurnFeeShareBps,
		AetraEconomicsStateParamRewardFeeShareBps,
		AetraEconomicsStateParamTreasuryFeeShareBps,
		AetraEconomicsStateParamRewardSmoothingEpochs,
		AetraEconomicsStateParamAprTargetMinBps,
		AetraEconomicsStateParamAprTargetMaxBps,
	}
}

func requiredAetraEconomicsEpochEconomicsFields() []string {
	return []string{
		AetraEconomicsStateEpochNumber,
		AetraEconomicsStateEpochStartHeight,
		AetraEconomicsStateEpochEndHeight,
		AetraEconomicsStateEpochBondedRatioBps,
		AetraEconomicsStateEpochInflationBps,
		AetraEconomicsStateEpochEstimatedAprBps,
		AetraEconomicsStateEpochFeesCollected,
		AetraEconomicsStateEpochFeesBurned,
		AetraEconomicsStateEpochFeesToRewards,
		AetraEconomicsStateEpochFeesToTreasury,
		AetraEconomicsStateEpochMintedRewards,
	}
}

func requiredAetraEconomicsSupplyStatsFields() []string {
	return []string{
		AetraEconomicsStateSupplyTotalMinted,
		AetraEconomicsStateSupplyTotalBurned,
		AetraEconomicsStateSupplyNetIssuance,
	}
}

func validateAetraEconomicsCatalog(group string, actual []string, required []string) (int, []string) {
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
