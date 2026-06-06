package params

import (
	"fmt"
	"sort"
	"strings"
)

const (
	EconomicGovernanceCategoryInflation = "inflation"
	EconomicGovernanceCategoryBurn      = "burn"
	EconomicGovernanceCategoryFee       = "fee"

	GovernanceParamMinimumInflation            = "minimum_inflation"
	GovernanceParamTargetInflation             = "target_inflation"
	GovernanceParamMaximumInflation            = "maximum_inflation"
	GovernanceParamPerWindowAdjustmentLimit    = "per_window_adjustment_limit"
	GovernanceParamSmoothingWindow             = "smoothing_window"
	GovernanceParamTargetStakeRatio            = "target_stake_ratio"
	GovernanceParamValidatorRewardFloor        = "validator_reward_floor"
	GovernanceParamNetIssuanceFloor            = "net_issuance_floor"
	GovernanceParamFeeBurnAllocation           = "fee_burn_allocation"
	GovernanceParamSlashingBurnAllocation      = "slashing_burn_allocation"
	GovernanceParamBurnCapPerEpoch             = "burn_cap_per_epoch"
	GovernanceParamBurnActivationThreshold     = "burn_activation_threshold"
	GovernanceParamDeflationGuardThreshold     = "deflation_guard_threshold"
	GovernanceParamMinimumBaseFee              = "minimum_base_fee"
	GovernanceParamMaximumBaseFee              = "maximum_base_fee"
	GovernanceParamTargetBlockUtilization      = "target_block_utilization"
	GovernanceParamMaxFeeAdjustmentPerWindow   = "maximum_fee_adjustment_per_window"
	GovernanceParamCongestionMultiplierBounds  = "congestion_multiplier_bounds"
	GovernanceParamSenderLocalSurcharge        = "sender_local_surcharge_parameters"
	GovernanceParamResourceSpecificMultipliers = "resource_specific_multipliers"
	GovernanceParamFeeAllocationBucketWeights  = "fee_allocation_bucket_weights"
)

type EconomicGovernanceParameter struct {
	ID               string
	Category         string
	Source           string
	Required         bool
	Queryable        bool
	Bounded          bool
	ImpactReport     bool
	ChangeControlled bool
	MinBps           int64
	MaxBps           int64
	Unit             string
	DependsOn        []string
}

type EconomicGovernanceSurfaceReport struct {
	Parameters           []EconomicGovernanceParameter
	RequiredInflation    int
	RequiredBurn         int
	RequiredFee          int
	CoveredInflation     int
	CoveredBurn          int
	CoveredFee           int
	InflationCoverageBps int64
	BurnCoverageBps      int64
	FeeCoverageBps       int64
	Passed               bool
	Failed               []string
	GovernanceSummary    string
}

func DefaultEconomicGovernanceParameters() []EconomicGovernanceParameter {
	return []EconomicGovernanceParameter{
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamMinimumInflation, "adaptive_inflation.min_inflation_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamTargetInflation, "adaptive_inflation.target_inflation_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamMaximumInflation, "adaptive_inflation.max_inflation_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamPerWindowAdjustmentLimit, "adaptive_inflation.per_window_change_limit_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamSmoothingWindow, "adaptive_inflation.smoothing_window", 1, 365, "epochs"),
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamTargetStakeRatio, "adaptive_inflation.target_stake_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamValidatorRewardFloor, "adaptive_inflation.validator_reward_floor_naet", 0, 0, "naet"),
		governanceParam(EconomicGovernanceCategoryInflation, GovernanceParamNetIssuanceFloor, "adaptive_inflation.net_issuance_floor_naet", 0, 0, "naet"),
		governanceParam(EconomicGovernanceCategoryBurn, GovernanceParamFeeBurnAllocation, "burn_deflation.fee_burn_allocation_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryBurn, GovernanceParamSlashingBurnAllocation, "burn_deflation.slashing_burn_allocation_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryBurn, GovernanceParamBurnCapPerEpoch, "burn_deflation.epoch_burn_cap_naet", 0, 0, "naet"),
		governanceParam(EconomicGovernanceCategoryBurn, GovernanceParamBurnActivationThreshold, "burn_deflation.activation_threshold_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryBurn, GovernanceParamDeflationGuardThreshold, "burn_deflation.deflation_guard_threshold_bps", 0, DefaultMaxLoadMultiplierBps, "bps"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamMinimumBaseFee, "fee_market.min_base_fee_naet", 0, 0, "naet"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamMaximumBaseFee, "fee_market.max_base_fee_naet", 0, 0, "naet"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamTargetBlockUtilization, "fee_market.target_block_utilization_bps", 1, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamMaxFeeAdjustmentPerWindow, "fee_market.max_fee_adjustment_bps", 0, BasisPoints, "bps"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamCongestionMultiplierBounds, "fee_market.resource_multiplier_bounds_bps", 0, DefaultMaxLoadMultiplierBps, "bps"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamSenderLocalSurcharge, "fee_market.sender_local_surcharge_bps", 0, DefaultMaxLoadMultiplierBps, "bps"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamResourceSpecificMultipliers, "fee_market.resource_specific_multipliers_bps", 0, DefaultMaxLoadMultiplierBps, "bps"),
		governanceParam(EconomicGovernanceCategoryFee, GovernanceParamFeeAllocationBucketWeights, "fee_market.bucket_weights_bps", 0, BasisPoints, "bps"),
	}
}

func BuildEconomicGovernanceSurfaceReport(parameters []EconomicGovernanceParameter) EconomicGovernanceSurfaceReport {
	if parameters == nil {
		parameters = DefaultEconomicGovernanceParameters()
	}
	params, failed, required, covered := evaluateEconomicGovernanceParameters(parameters)
	sort.Strings(failed)
	inflationCoverage := coverageBps(covered[EconomicGovernanceCategoryInflation], required[EconomicGovernanceCategoryInflation])
	burnCoverage := coverageBps(covered[EconomicGovernanceCategoryBurn], required[EconomicGovernanceCategoryBurn])
	feeCoverage := coverageBps(covered[EconomicGovernanceCategoryFee], required[EconomicGovernanceCategoryFee])
	return EconomicGovernanceSurfaceReport{
		Parameters:           params,
		RequiredInflation:    required[EconomicGovernanceCategoryInflation],
		RequiredBurn:         required[EconomicGovernanceCategoryBurn],
		RequiredFee:          required[EconomicGovernanceCategoryFee],
		CoveredInflation:     covered[EconomicGovernanceCategoryInflation],
		CoveredBurn:          covered[EconomicGovernanceCategoryBurn],
		CoveredFee:           covered[EconomicGovernanceCategoryFee],
		InflationCoverageBps: inflationCoverage,
		BurnCoverageBps:      burnCoverage,
		FeeCoverageBps:       feeCoverage,
		Passed:               len(failed) == 0 && inflationCoverage == BasisPoints && burnCoverage == BasisPoints && feeCoverage == BasisPoints,
		Failed:               failed,
		GovernanceSummary:    fmt.Sprintf("inflation=%d/%d burn=%d/%d fee=%d/%d coverage_bps=%d/%d/%d", covered[EconomicGovernanceCategoryInflation], required[EconomicGovernanceCategoryInflation], covered[EconomicGovernanceCategoryBurn], required[EconomicGovernanceCategoryBurn], covered[EconomicGovernanceCategoryFee], required[EconomicGovernanceCategoryFee], inflationCoverage, burnCoverage, feeCoverage),
	}
}

func governanceParam(category, id, source string, minBps, maxBps int64, unit string, dependsOn ...string) EconomicGovernanceParameter {
	return EconomicGovernanceParameter{
		ID:               id,
		Category:         category,
		Source:           source,
		Required:         true,
		Queryable:        true,
		Bounded:          true,
		ImpactReport:     true,
		ChangeControlled: true,
		MinBps:           minBps,
		MaxBps:           maxBps,
		Unit:             unit,
		DependsOn:        append([]string{}, dependsOn...),
	}
}

func evaluateEconomicGovernanceParameters(parameters []EconomicGovernanceParameter) ([]EconomicGovernanceParameter, []string, map[string]int, map[string]int) {
	out := append([]EconomicGovernanceParameter{}, parameters...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Category == out[j].Category {
			return out[i].ID < out[j].ID
		}
		return out[i].Category < out[j].Category
	})
	expected := requiredEconomicGovernanceParameterIDs()
	required := map[string]int{}
	covered := map[string]int{}
	failed := make([]string, 0)
	seen := make(map[string]EconomicGovernanceParameter, len(out))
	for category, ids := range expected {
		required[category] = len(ids)
	}
	for _, param := range out {
		if param.ID == "" {
			failed = append(failed, "governance_parameter_id_required")
			continue
		}
		if _, ok := expected[param.Category]; !ok {
			failed = append(failed, param.ID+":unknown_governance_category")
		}
		if _, ok := seen[param.ID]; ok {
			failed = append(failed, param.ID+":duplicate_governance_parameter")
		}
		seen[param.ID] = param
		if param.Required {
			if strings.TrimSpace(param.Source) == "" {
				failed = append(failed, param.ID+":source_missing")
			}
			if strings.TrimSpace(param.Unit) == "" {
				failed = append(failed, param.ID+":unit_missing")
			}
			if !param.Queryable {
				failed = append(failed, param.ID+":not_queryable")
			}
			if !param.Bounded {
				failed = append(failed, param.ID+":not_bounded")
			}
			if !param.ImpactReport {
				failed = append(failed, param.ID+":impact_report_missing")
			}
			if !param.ChangeControlled {
				failed = append(failed, param.ID+":change_control_missing")
			}
			if param.MinBps < 0 || param.MaxBps < 0 {
				failed = append(failed, param.ID+":negative_bounds")
			}
			if param.MaxBps > 0 && param.MinBps > param.MaxBps {
				failed = append(failed, param.ID+":invalid_bound_order")
			}
		}
	}
	for category, ids := range expected {
		for _, id := range ids {
			param, ok := seen[id]
			if !ok {
				failed = append(failed, id+":missing_required_governance_parameter")
				continue
			}
			if governanceParameterCovered(category, param) {
				covered[category]++
			}
		}
	}
	return out, failed, required, covered
}

func governanceParameterCovered(category string, param EconomicGovernanceParameter) bool {
	return param.Required &&
		param.Category == category &&
		param.Queryable &&
		param.Bounded &&
		param.ImpactReport &&
		param.ChangeControlled &&
		strings.TrimSpace(param.Source) != "" &&
		strings.TrimSpace(param.Unit) != "" &&
		param.MinBps >= 0 &&
		param.MaxBps >= 0 &&
		(param.MaxBps == 0 || param.MinBps <= param.MaxBps)
}

func requiredEconomicGovernanceParameterIDs() map[string][]string {
	return map[string][]string{
		EconomicGovernanceCategoryInflation: {
			GovernanceParamMinimumInflation,
			GovernanceParamTargetInflation,
			GovernanceParamMaximumInflation,
			GovernanceParamPerWindowAdjustmentLimit,
			GovernanceParamSmoothingWindow,
			GovernanceParamTargetStakeRatio,
			GovernanceParamValidatorRewardFloor,
			GovernanceParamNetIssuanceFloor,
		},
		EconomicGovernanceCategoryBurn: {
			GovernanceParamFeeBurnAllocation,
			GovernanceParamSlashingBurnAllocation,
			GovernanceParamBurnCapPerEpoch,
			GovernanceParamBurnActivationThreshold,
			GovernanceParamDeflationGuardThreshold,
		},
		EconomicGovernanceCategoryFee: {
			GovernanceParamMinimumBaseFee,
			GovernanceParamMaximumBaseFee,
			GovernanceParamTargetBlockUtilization,
			GovernanceParamMaxFeeAdjustmentPerWindow,
			GovernanceParamCongestionMultiplierBounds,
			GovernanceParamSenderLocalSurcharge,
			GovernanceParamResourceSpecificMultipliers,
			GovernanceParamFeeAllocationBucketWeights,
		},
	}
}
