package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	DefaultMaxValidatorFeeShareBps		= int64(9_000)
	DefaultMaxNativeFeePressureBps		= int64(1_000)
	DefaultMinFeeNormalizationBps		= int64(500)
	DefaultMaxFeeNormalizationBps		= int64(20_000)
	DefaultMaxHardwareUtilizationBps	= int64(8_000)
	DefaultMinSimulationSampleCount		= 4
	DefaultCongestedMaintenanceCoverage	= appparams.BasisPoints
)

type FeeModelEfficiencyInput struct {
	Params				Params
	SimulationLoadsBps		[]uint32
	CurrentBlockLoadBps		uint32
	ValidatorRewardCoverageBps	int64
	MaintenanceFundingCoverageBps	int64
	NativeLiquidityDepthNaet	sdkmath.Int
	DailyFeePressureNaet		sdkmath.Int
	AntiSpamMultiplierBps		int64
	BurnIntegratedWithIssuance	bool
	FeeBurnRatioBps			int64
	BlockTxFeeNaet			sdkmath.Int
	StorageFeeNaet			sdkmath.Int
	ExecutionFeeNaet		sdkmath.Int
	ExpectedBlockProcessingMs	uint64
	ObservedBlockProcessingMs	uint64
	ExpectedValidatorMemoryMB	uint64
	ObservedValidatorMemoryMB	uint64
	MaxValidatorFeeShareBps		int64
	MaxNativeFeePressureBps		int64
	MinFeeNormalizationBps		int64
	MaxFeeNormalizationBps		int64
	MaxHardwareUtilizationBps	int64
	MinSimulationSampleCount	int
}

type FeeModelEfficiencyReport struct {
	Healthy			bool
	ValidatorFeeShareBps	int64
	CommunityFeeShareBps	int64
	NativeFeePressureBps	int64
	StorageExecutionFeeBps	int64
	BlockProcessingUsageBps	int64
	ValidatorMemoryUsageBps	int64
	SimulationSampleCount	int
	Risks			[]string
}

type FeeModelSimulationReport struct {
	Passed		bool
	SampleCount	int
	Risks		[]string
}

func EvaluateFeeModelEfficiency(input FeeModelEfficiencyInput) (FeeModelEfficiencyReport, error) {
	applyFeeEfficiencyDefaults(&input)
	params := NormalizeParams(input.Params)
	if err := params.Validate(); err != nil {
		return FeeModelEfficiencyReport{}, err
	}
	input.Params = params
	if err := validateFeeEfficiencyInput(input); err != nil {
		return FeeModelEfficiencyReport{}, err
	}
	validatorShare, communityShare, err := feeSplitBps(params)
	if err != nil {
		return FeeModelEfficiencyReport{}, err
	}
	simulation, err := SimulateFeeModel(params, input.SimulationLoadsBps, input.MinSimulationSampleCount)
	if err != nil {
		return FeeModelEfficiencyReport{}, err
	}

	risks := make([]string, 0, 8)
	if input.CurrentBlockLoadBps >= params.CongestionThresholdBps &&
		validatorShare > input.MaxValidatorFeeShareBps &&
		input.MaintenanceFundingCoverageBps < DefaultCongestedMaintenanceCoverage {
		risks = append(risks, "static_fee_split_overpays_validators_underfunds_maintenance")
	}
	if !simulation.Passed {
		risks = append(risks, simulation.Risks...)
	}

	feePressureBps := feePressureBps(input.DailyFeePressureNaet, input.NativeLiquidityDepthNaet)
	if len(params.AllowedFeeDenoms) == 1 && params.AllowedFeeDenoms[0] == BondDenom && feePressureBps > input.MaxNativeFeePressureBps {
		risks = append(risks, "native_denom_fee_pressure_concentrated")
	}
	if input.CurrentBlockLoadBps >= params.CongestionThresholdBps && input.AntiSpamMultiplierBps < appparams.MinSpamCostMultiplierBps {
		risks = append(risks, "anti_spam_multiplier_too_weak")
	}
	if input.CurrentBlockLoadBps >= params.CongestionThresholdBps && (!input.BurnIntegratedWithIssuance || input.FeeBurnRatioBps == 0) {
		risks = append(risks, "fee_burn_not_integrated_with_issuance")
	}

	storageExecutionBps := storageExecutionFeeBps(input.StorageFeeNaet, input.ExecutionFeeNaet, input.BlockTxFeeNaet)
	if storageExecutionBps < input.MinFeeNormalizationBps || storageExecutionBps > input.MaxFeeNormalizationBps {
		risks = append(risks, "storage_execution_fees_not_normalized_to_block_fees")
	}

	blockProcessingUsage := usageBps(input.ObservedBlockProcessingMs, input.ExpectedBlockProcessingMs)
	memoryUsage := usageBps(input.ObservedValidatorMemoryMB, input.ExpectedValidatorMemoryMB)
	if input.ExpectedBlockProcessingMs == 0 || input.ExpectedValidatorMemoryMB == 0 {
		risks = append(risks, "validator_hardware_calibration_missing")
	} else if blockProcessingUsage > input.MaxHardwareUtilizationBps || memoryUsage > input.MaxHardwareUtilizationBps {
		risks = append(risks, "gas_limits_exceed_validator_hardware_budget")
	}

	return FeeModelEfficiencyReport{
		Healthy:			len(risks) == 0,
		ValidatorFeeShareBps:		validatorShare,
		CommunityFeeShareBps:		communityShare,
		NativeFeePressureBps:		feePressureBps,
		StorageExecutionFeeBps:		storageExecutionBps,
		BlockProcessingUsageBps:	blockProcessingUsage,
		ValidatorMemoryUsageBps:	memoryUsage,
		SimulationSampleCount:		simulation.SampleCount,
		Risks:				risks,
	}, nil
}

func SimulateFeeModel(params Params, loads []uint32, minSamples int) (FeeModelSimulationReport, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return FeeModelSimulationReport{}, err
	}
	if minSamples == 0 {
		minSamples = DefaultMinSimulationSampleCount
	}
	if minSamples < 0 {
		return FeeModelSimulationReport{}, fmt.Errorf("min simulation sample count must not be negative")
	}
	baseFee, err := params.BaseFeeInt()
	if err != nil {
		return FeeModelSimulationReport{}, err
	}
	maxFee, err := params.MaxFeeInt()
	if err != nil {
		return FeeModelSimulationReport{}, err
	}
	risks := make([]string, 0, 3)
	if len(loads) < minSamples {
		risks = append(risks, "fee_simulation_coverage_too_low")
	}
	previous := sdkmath.ZeroInt()
	for i, load := range loads {
		if load > uint32(BasisPoints) {
			return FeeModelSimulationReport{}, fmt.Errorf("simulation load %d exceeds 10000 bps", i)
		}
		fee := DynamicFeeAmount(baseFee, maxFee, params.TargetBlockUtilizationBps, load)
		if fee.LT(baseFee) || fee.GT(maxFee) {
			risks = append(risks, "fee_simulation_outside_bounds")
			break
		}
		if i > 0 && fee.LT(previous) {
			risks = append(risks, "fee_simulation_not_monotonic")
			break
		}
		previous = fee
	}
	return FeeModelSimulationReport{
		Passed:		len(risks) == 0,
		SampleCount:	len(loads),
		Risks:		risks,
	}, nil
}

func applyFeeEfficiencyDefaults(input *FeeModelEfficiencyInput) {
	if input.Params.AllowedFeeDenoms == nil {
		input.Params = DefaultParams()
	}
	if input.MaxValidatorFeeShareBps == 0 {
		input.MaxValidatorFeeShareBps = DefaultMaxValidatorFeeShareBps
	}
	if input.MaxNativeFeePressureBps == 0 {
		input.MaxNativeFeePressureBps = DefaultMaxNativeFeePressureBps
	}
	if input.MinFeeNormalizationBps == 0 {
		input.MinFeeNormalizationBps = DefaultMinFeeNormalizationBps
	}
	if input.MaxFeeNormalizationBps == 0 {
		input.MaxFeeNormalizationBps = DefaultMaxFeeNormalizationBps
	}
	if input.MaxHardwareUtilizationBps == 0 {
		input.MaxHardwareUtilizationBps = DefaultMaxHardwareUtilizationBps
	}
	if input.MinSimulationSampleCount == 0 {
		input.MinSimulationSampleCount = DefaultMinSimulationSampleCount
	}
}

func validateFeeEfficiencyInput(input FeeModelEfficiencyInput) error {
	for _, item := range []struct {
		name	string
		value	int64
	}{
		{name: "validator_reward_coverage_bps", value: input.ValidatorRewardCoverageBps},
		{name: "maintenance_funding_coverage_bps", value: input.MaintenanceFundingCoverageBps},
		{name: "anti_spam_multiplier_bps", value: input.AntiSpamMultiplierBps},
		{name: "fee_burn_ratio_bps", value: input.FeeBurnRatioBps},
		{name: "max_validator_fee_share_bps", value: input.MaxValidatorFeeShareBps},
		{name: "max_native_fee_pressure_bps", value: input.MaxNativeFeePressureBps},
		{name: "min_fee_normalization_bps", value: input.MinFeeNormalizationBps},
		{name: "max_fee_normalization_bps", value: input.MaxFeeNormalizationBps},
		{name: "max_hardware_utilization_bps", value: input.MaxHardwareUtilizationBps},
	} {
		if err := validateIntBps(item.name, item.value, 0, appparams.DefaultMaxLoadMultiplierBps); err != nil {
			return err
		}
	}
	if input.MinFeeNormalizationBps > input.MaxFeeNormalizationBps {
		return fmt.Errorf("min_fee_normalization_bps must be <= max_fee_normalization_bps")
	}
	if input.MinSimulationSampleCount < 0 {
		return fmt.Errorf("min_simulation_sample_count must not be negative")
	}
	for _, item := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "native_liquidity_depth_naet", value: input.NativeLiquidityDepthNaet},
		{name: "daily_fee_pressure_naet", value: input.DailyFeePressureNaet},
		{name: "block_tx_fee_naet", value: input.BlockTxFeeNaet},
		{name: "storage_fee_naet", value: input.StorageFeeNaet},
		{name: "execution_fee_naet", value: input.ExecutionFeeNaet},
	} {
		if normalizeInt(item.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", item.name)
		}
	}
	if input.CurrentBlockLoadBps > uint32(BasisPoints) {
		return fmt.Errorf("current_block_load_bps must be <= 10000")
	}
	return nil
}

func feeSplitBps(params Params) (int64, int64, error) {
	validatorRatio, err := validateRatio("validator_rewards_ratio", params.ValidatorRewardsRatio)
	if err != nil {
		return 0, 0, err
	}
	communityRatio, err := validateRatio("community_pool_ratio", params.CommunityPoolRatio)
	if err != nil {
		return 0, 0, err
	}
	validatorBps := validatorRatio.MulInt64(int64(BasisPoints)).TruncateInt64()
	communityBps := communityRatio.MulInt64(int64(BasisPoints)).TruncateInt64()
	return validatorBps, communityBps, nil
}

func feePressureBps(pressure, liquidity sdkmath.Int) int64 {
	pressure = normalizeInt(pressure)
	liquidity = normalizeInt(liquidity)
	if pressure.IsZero() || !liquidity.IsPositive() {
		return 0
	}
	return pressure.MulRaw(int64(BasisPoints)).Quo(liquidity).Int64()
}

func storageExecutionFeeBps(storageFee, executionFee, blockTxFee sdkmath.Int) int64 {
	total := normalizeInt(storageFee).Add(normalizeInt(executionFee))
	blockTxFee = normalizeInt(blockTxFee)
	if total.IsZero() || !blockTxFee.IsPositive() {
		return 0
	}
	return total.MulRaw(int64(BasisPoints)).Quo(blockTxFee).Int64()
}

func usageBps(observed, expected uint64) int64 {
	if observed == 0 || expected == 0 {
		return 0
	}
	return sdkmath.NewIntFromUint64(observed).MulRaw(int64(BasisPoints)).Quo(sdkmath.NewIntFromUint64(expected)).Int64()
}

func validateIntBps(name string, value, min, max int64) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return nil
}

func normalizeInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}
