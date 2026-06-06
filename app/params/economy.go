package params

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	BasisPoints = int64(10_000)

	MinCommissionBps            = int64(100)
	MaxCommissionBps            = int64(2_000)
	MaxDailyCommissionChangeBps = int64(100)

	MinInflationBps             = int64(100)
	MaxInflationBps             = int64(500)
	DefaultTargetInflationBps   = int64(300)
	DefaultTargetStakeBps       = int64(6_700)
	DefaultResponsivenessBps    = int64(800)
	DefaultActivityCouplingBps  = int64(100)
	NormalBurnRatioBps          = int64(3_000)
	CongestedBurnRatioBps       = int64(4_000)
	MinBurnRatioBps             = int64(1_000)
	MaxBurnRatioBps             = int64(5_000)
	TreasuryFeeRatioBps         = int64(1_000)
	DefaultTargetLoadBps        = int64(7_000)
	HighCongestionLoadBps       = int64(9_000)
	DeflationGuardBurnToMintBps = int64(12_500)
	DeflationGuardStepBps       = int64(500)
	RateLimitFailedTxBps        = int64(1_000)
	DefaultMaxLoadMultiplierBps = int64(40_000)

	DefaultStakeTargetToleranceBps    = int64(500)
	MaxTopValidatorConcentrationBps   = int64(3_334)
	MinValidatorRewardCoverageBps     = BasisPoints
	MinDelegatorRiskSignalCoverageBps = BasisPoints
	MinFeeResponseBps                 = BasisPoints
	MinSpamCostMultiplierBps          = BasisPoints
	MinStorageCostCoverageBps         = BasisPoints
	MinSlashingPenaltyCoverageBps     = BasisPoints
	MinTreasuryFundingCoverageBps     = BasisPoints
)

type ValidatorIncomeInput struct {
	TotalMintRewards sdkmath.Int
	TotalFeeRewards  sdkmath.Int
	ValidatorPower   sdkmath.Int
	TotalPower       sdkmath.Int
	CommissionBps    int64
}

type ValidatorIncome struct {
	RewardWeightBps       int64
	MintRewardShare       sdkmath.Int
	FeeRewardShare        sdkmath.Int
	ValidatorCommission   sdkmath.Int
	ValidatorIncome       sdkmath.Int
	DelegatorIncome       sdkmath.Int
	DelegatorGrossRewards sdkmath.Int
}

type BalanceControllerInput struct {
	CurrentInflationBps int64
	StakeRatioBps       int64
	BlockLoadBps        int64
	AnnualMint          sdkmath.Int
	AnnualBurn          sdkmath.Int
	Activity            ProtocolEconomicActivity
	AsyncQueueDepth     uint64
	FailedTxRateBps     int64
}

type BalanceControllerParams struct {
	MinInflationBps              int64
	MaxInflationBps              int64
	TargetStakeBps               int64
	InflationResponsivenessBps   int64
	ActivityInflationCouplingBps int64
	NormalBurnRatioBps           int64
	CongestedBurnRatioBps        int64
	MinBurnRatioBps              int64
	MaxBurnRatioBps              int64
	TreasuryFeeRatioBps          int64
	TargetLoadBps                int64
	HighCongestionLoadBps        int64
	DeflationGuardBurnToMintBps  int64
	DeflationGuardStepBps        int64
	RateLimitFailedTxBps         int64
}

type BalanceControllerOutput struct {
	InflationBps              int64
	StakeInflationDeltaBps    int64
	ActivityInflationDeltaBps int64
	BurnRatioBps              int64
	ValidatorFeeRatioBps      int64
	Congested                 bool
	DeflationGuardActive      bool
	QueueLimited              bool
	RateLimited               bool
}

type ProtocolEconomicActivity struct {
	TxFeeNaet             sdkmath.Int
	AVMStorageFeeNaet     sdkmath.Int
	AVMForwardingFeeNaet  sdkmath.Int
	AVMDeploymentCostNaet sdkmath.Int
}

type ProtocolEconomicFlowInput struct {
	Activity         ProtocolEconomicActivity
	BurnRatioBps     int64
	TreasuryRatioBps int64
}

type ProtocolEconomicFlowOutput struct {
	TotalChargesNaet     sdkmath.Int
	BurnNaet             sdkmath.Int
	TreasuryNaet         sdkmath.Int
	ValidatorRewardsNaet sdkmath.Int
}

type OptimalEconomicStateInput struct {
	StakeRatioBps                  int64
	StakeTargetToleranceBps        int64
	InflationBps                   int64
	ValidatorRewardCoverageBps     int64
	DelegatorRiskSignalCoverageBps int64
	ActiveValidatorCount           uint64
	MinActiveValidatorCount        uint64
	TopValidatorStakeBps           int64
	BlockLoadBps                   int64
	FeeResponseBps                 int64
	SpamCostMultiplierBps          int64
	StorageCostCoverageBps         int64
	BurnToMintBps                  int64
	SlashingPenaltyCoverageBps     int64
	TreasuryFundingCoverageBps     int64
}

type OptimalEconomicState struct {
	Optimal          bool
	FailedConditions []string
}

func DefaultBalanceControllerParams() BalanceControllerParams {
	return BalanceControllerParams{
		MinInflationBps:              MinInflationBps,
		MaxInflationBps:              MaxInflationBps,
		TargetStakeBps:               DefaultTargetStakeBps,
		InflationResponsivenessBps:   DefaultResponsivenessBps,
		ActivityInflationCouplingBps: DefaultActivityCouplingBps,
		NormalBurnRatioBps:           NormalBurnRatioBps,
		CongestedBurnRatioBps:        CongestedBurnRatioBps,
		MinBurnRatioBps:              MinBurnRatioBps,
		MaxBurnRatioBps:              MaxBurnRatioBps,
		TreasuryFeeRatioBps:          TreasuryFeeRatioBps,
		TargetLoadBps:                DefaultTargetLoadBps,
		HighCongestionLoadBps:        HighCongestionLoadBps,
		DeflationGuardBurnToMintBps:  DeflationGuardBurnToMintBps,
		DeflationGuardStepBps:        DeflationGuardStepBps,
		RateLimitFailedTxBps:         RateLimitFailedTxBps,
	}
}

func EvaluateOptimalEconomicState(input OptimalEconomicStateInput) (OptimalEconomicState, error) {
	if input.StakeTargetToleranceBps == 0 {
		input.StakeTargetToleranceBps = DefaultStakeTargetToleranceBps
	}
	if input.MinActiveValidatorCount == 0 {
		input.MinActiveValidatorCount = 1
	}
	if err := validateBps("stake_ratio_bps", input.StakeRatioBps, 0, BasisPoints); err != nil {
		return OptimalEconomicState{}, err
	}
	if err := validateBps("stake_target_tolerance_bps", input.StakeTargetToleranceBps, 1, BasisPoints); err != nil {
		return OptimalEconomicState{}, err
	}
	if err := validateBps("inflation_bps", input.InflationBps, MinInflationBps, MaxInflationBps); err != nil {
		return OptimalEconomicState{}, err
	}
	for _, item := range []struct {
		name  string
		value int64
	}{
		{name: "validator_reward_coverage_bps", value: input.ValidatorRewardCoverageBps},
		{name: "delegator_risk_signal_coverage_bps", value: input.DelegatorRiskSignalCoverageBps},
		{name: "fee_response_bps", value: input.FeeResponseBps},
		{name: "spam_cost_multiplier_bps", value: input.SpamCostMultiplierBps},
		{name: "storage_cost_coverage_bps", value: input.StorageCostCoverageBps},
		{name: "burn_to_mint_bps", value: input.BurnToMintBps},
		{name: "slashing_penalty_coverage_bps", value: input.SlashingPenaltyCoverageBps},
		{name: "treasury_funding_coverage_bps", value: input.TreasuryFundingCoverageBps},
	} {
		if err := validateBps(item.name, item.value, 0, DefaultMaxLoadMultiplierBps); err != nil {
			return OptimalEconomicState{}, err
		}
	}
	if err := validateBps("top_validator_stake_bps", input.TopValidatorStakeBps, 0, BasisPoints); err != nil {
		return OptimalEconomicState{}, err
	}
	if err := validateBps("block_load_bps", input.BlockLoadBps, 0, BasisPoints); err != nil {
		return OptimalEconomicState{}, err
	}

	failed := make([]string, 0, 12)
	if absInt64(input.StakeRatioBps-DefaultTargetStakeBps) > input.StakeTargetToleranceBps {
		failed = append(failed, "stake_ratio_outside_target_band")
	}
	if input.ValidatorRewardCoverageBps < MinValidatorRewardCoverageBps {
		failed = append(failed, "validator_rewards_below_operating_cost")
	}
	if input.DelegatorRiskSignalCoverageBps < MinDelegatorRiskSignalCoverageBps {
		failed = append(failed, "delegator_risk_signals_incomplete")
	}
	if input.ActiveValidatorCount < input.MinActiveValidatorCount {
		failed = append(failed, "active_validator_set_too_small")
	}
	if input.TopValidatorStakeBps > MaxTopValidatorConcentrationBps {
		failed = append(failed, "validator_stake_too_concentrated")
	}
	if input.FeeResponseBps < MinFeeResponseBps || input.FeeResponseBps > DefaultMaxLoadMultiplierBps {
		failed = append(failed, "fee_response_outside_predictable_bounds")
	}
	if input.SpamCostMultiplierBps < MinSpamCostMultiplierBps {
		failed = append(failed, "spam_cost_not_escalating")
	}
	if input.StorageCostCoverageBps < MinStorageCostCoverageBps {
		failed = append(failed, "storage_cost_not_accountable")
	}
	if input.BurnToMintBps > DeflationGuardBurnToMintBps {
		failed = append(failed, "burn_pressure_exceeds_deflation_guard")
	}
	if input.SlashingPenaltyCoverageBps < MinSlashingPenaltyCoverageBps {
		failed = append(failed, "slashing_penalties_under_security_damage")
	}
	if input.TreasuryFundingCoverageBps < MinTreasuryFundingCoverageBps {
		failed = append(failed, "treasury_funding_below_maintenance_need")
	}

	return OptimalEconomicState{
		Optimal:          len(failed) == 0,
		FailedConditions: failed,
	}, nil
}

func (p BalanceControllerParams) Validate() error {
	if err := validateBps("min_inflation_bps", p.MinInflationBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_inflation_bps", p.MaxInflationBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MinInflationBps > p.MaxInflationBps {
		return fmt.Errorf("min_inflation_bps must be <= max_inflation_bps")
	}
	if err := validateBps("target_stake_bps", p.TargetStakeBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("inflation_responsiveness_bps", p.InflationResponsivenessBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("activity_inflation_coupling_bps", p.ActivityInflationCouplingBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("normal_burn_ratio_bps", p.NormalBurnRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("congested_burn_ratio_bps", p.CongestedBurnRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("min_burn_ratio_bps", p.MinBurnRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_burn_ratio_bps", p.MaxBurnRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MinBurnRatioBps > p.MaxBurnRatioBps {
		return fmt.Errorf("min_burn_ratio_bps must be <= max_burn_ratio_bps")
	}
	if err := validateBps("treasury_fee_ratio_bps", p.TreasuryFeeRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MaxBurnRatioBps+p.TreasuryFeeRatioBps > BasisPoints {
		return fmt.Errorf("max burn and treasury ratios exceed 100%%")
	}
	if err := validateBps("target_load_bps", p.TargetLoadBps, 0, BasisPoints-1); err != nil {
		return err
	}
	if err := validateBps("high_congestion_load_bps", p.HighCongestionLoadBps, 1, BasisPoints); err != nil {
		return err
	}
	if p.HighCongestionLoadBps <= p.TargetLoadBps {
		return fmt.Errorf("high_congestion_load_bps must be greater than target_load_bps")
	}
	if p.DeflationGuardBurnToMintBps < BasisPoints {
		return fmt.Errorf("deflation_guard_burn_to_mint_bps must be >= 10000")
	}
	if err := validateBps("deflation_guard_step_bps", p.DeflationGuardStepBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("rate_limit_failed_tx_bps", p.RateLimitFailedTxBps, 0, BasisPoints); err != nil {
		return err
	}
	return nil
}

func ComputeValidatorIncome(input ValidatorIncomeInput) (ValidatorIncome, error) {
	totalMintRewards := normalizeInt(input.TotalMintRewards)
	totalFeeRewards := normalizeInt(input.TotalFeeRewards)
	validatorPower := normalizeInt(input.ValidatorPower)
	totalPower := normalizeInt(input.TotalPower)
	if totalMintRewards.IsNegative() || totalFeeRewards.IsNegative() {
		return ValidatorIncome{}, fmt.Errorf("validator rewards must not be negative")
	}
	if !validatorPower.IsPositive() || !totalPower.IsPositive() {
		return ValidatorIncome{}, fmt.Errorf("validator and total power must be positive")
	}
	if validatorPower.GT(totalPower) {
		return ValidatorIncome{}, fmt.Errorf("validator power must be <= total power")
	}
	if err := ValidateCommissionBounds(input.CommissionBps, 0); err != nil {
		return ValidatorIncome{}, err
	}

	rewardWeightBps := validatorPower.MulRaw(BasisPoints).Quo(totalPower).Int64()
	mintShare := ProportionalShare(totalMintRewards, validatorPower, totalPower)
	feeShare := ProportionalShare(totalFeeRewards, validatorPower, totalPower)
	gross := mintShare.Add(feeShare)
	commission := ApplyBps(gross, input.CommissionBps)
	delegatorIncome := gross.Sub(commission)

	return ValidatorIncome{
		RewardWeightBps:       rewardWeightBps,
		MintRewardShare:       mintShare,
		FeeRewardShare:        feeShare,
		ValidatorCommission:   commission,
		ValidatorIncome:       gross.Add(commission),
		DelegatorIncome:       delegatorIncome,
		DelegatorGrossRewards: gross,
	}, nil
}

func BalanceController(input BalanceControllerInput) (BalanceControllerOutput, error) {
	return BalanceControllerWithParams(input, DefaultBalanceControllerParams())
}

func BalanceControllerWithParams(input BalanceControllerInput, params BalanceControllerParams) (BalanceControllerOutput, error) {
	if err := params.Validate(); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := validateBps("current_inflation_bps", input.CurrentInflationBps, params.MinInflationBps, params.MaxInflationBps); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := validateBps("stake_ratio_bps", input.StakeRatioBps, 0, BasisPoints); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := validateBps("block_load_bps", input.BlockLoadBps, 0, BasisPoints); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := validateBps("failed_tx_rate_bps", input.FailedTxRateBps, 0, BasisPoints); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := input.Activity.Validate(); err != nil {
		return BalanceControllerOutput{}, err
	}
	annualMint := normalizeInt(input.AnnualMint)
	annualBurn := normalizeInt(input.AnnualBurn).Add(input.Activity.TotalCharges())
	if annualMint.IsNegative() || annualBurn.IsNegative() {
		return BalanceControllerOutput{}, fmt.Errorf("annual mint and burn must not be negative")
	}

	stakeDelta := params.InflationResponsivenessBps * (params.TargetStakeBps - input.StakeRatioBps) / BasisPoints
	activityDelta := int64(0)
	if input.BlockLoadBps > params.TargetLoadBps && params.ActivityInflationCouplingBps > 0 {
		activityDelta = -(params.ActivityInflationCouplingBps * (input.BlockLoadBps - params.TargetLoadBps) / (BasisPoints - params.TargetLoadBps))
	}
	inflation := clampInt64(input.CurrentInflationBps+stakeDelta+activityDelta, params.MinInflationBps, params.MaxInflationBps)

	burnRatio := params.NormalBurnRatioBps
	congested := input.BlockLoadBps >= params.HighCongestionLoadBps
	if congested {
		burnRatio = params.CongestedBurnRatioBps
	}

	deflationGuard := false
	maxBurn := sdkmath.ZeroInt()
	if annualMint.IsPositive() {
		maxBurn = annualMint.MulRaw(params.DeflationGuardBurnToMintBps).QuoRaw(BasisPoints)
	}
	if annualBurn.IsPositive() && (!annualMint.IsPositive() || annualBurn.GT(maxBurn)) {
		deflationGuard = true
		burnRatio = clampInt64(burnRatio-params.DeflationGuardStepBps, params.MinBurnRatioBps, params.MaxBurnRatioBps)
	}
	burnRatio = clampInt64(burnRatio, params.MinBurnRatioBps, params.MaxBurnRatioBps)
	validatorFeeRatio := BasisPoints - burnRatio - params.TreasuryFeeRatioBps
	if validatorFeeRatio < 0 {
		return BalanceControllerOutput{}, fmt.Errorf("burn and treasury ratios exceed 100%%")
	}

	return BalanceControllerOutput{
		InflationBps:              inflation,
		StakeInflationDeltaBps:    stakeDelta,
		ActivityInflationDeltaBps: activityDelta,
		BurnRatioBps:              burnRatio,
		ValidatorFeeRatioBps:      validatorFeeRatio,
		Congested:                 congested,
		DeflationGuardActive:      deflationGuard,
		QueueLimited:              input.AsyncQueueDepth > 0 && input.BlockLoadBps > params.TargetLoadBps,
		RateLimited:               input.FailedTxRateBps > params.RateLimitFailedTxBps || input.BlockLoadBps >= params.HighCongestionLoadBps,
	}, nil
}

func (a ProtocolEconomicActivity) Validate() error {
	for _, item := range []struct {
		name  string
		value sdkmath.Int
	}{
		{name: "tx_fee_naet", value: a.TxFeeNaet},
		{name: "avm_storage_fee_naet", value: a.AVMStorageFeeNaet},
		{name: "avm_forwarding_fee_naet", value: a.AVMForwardingFeeNaet},
		{name: "avm_deployment_cost_naet", value: a.AVMDeploymentCostNaet},
	} {
		value := normalizeInt(item.value)
		if value.IsNegative() {
			return fmt.Errorf("%s must not be negative", item.name)
		}
	}
	return nil
}

func (a ProtocolEconomicActivity) TotalCharges() sdkmath.Int {
	return normalizeInt(a.TxFeeNaet).
		Add(normalizeInt(a.AVMStorageFeeNaet)).
		Add(normalizeInt(a.AVMForwardingFeeNaet)).
		Add(normalizeInt(a.AVMDeploymentCostNaet))
}

func ComputeProtocolEconomicFlow(input ProtocolEconomicFlowInput) (ProtocolEconomicFlowOutput, error) {
	if err := input.Activity.Validate(); err != nil {
		return ProtocolEconomicFlowOutput{}, err
	}
	if err := validateBps("burn_ratio_bps", input.BurnRatioBps, 0, BasisPoints); err != nil {
		return ProtocolEconomicFlowOutput{}, err
	}
	if err := validateBps("treasury_ratio_bps", input.TreasuryRatioBps, 0, BasisPoints); err != nil {
		return ProtocolEconomicFlowOutput{}, err
	}
	if input.BurnRatioBps+input.TreasuryRatioBps > BasisPoints {
		return ProtocolEconomicFlowOutput{}, fmt.Errorf("burn and treasury ratios exceed 100%%")
	}
	total := input.Activity.TotalCharges()
	burn := ApplyBps(total, input.BurnRatioBps)
	treasury := ApplyBps(total, input.TreasuryRatioBps)
	validator := total.Sub(burn).Sub(treasury)
	return ProtocolEconomicFlowOutput{
		TotalChargesNaet:     total,
		BurnNaet:             burn,
		TreasuryNaet:         treasury,
		ValidatorRewardsNaet: validator,
	}, nil
}

func ValidateCommissionBounds(commissionBps, dailyChangeBps int64) error {
	if err := validateBps("commission_bps", commissionBps, MinCommissionBps, MaxCommissionBps); err != nil {
		return err
	}
	return validateBps("daily_commission_change_bps", dailyChangeBps, 0, MaxDailyCommissionChangeBps)
}

func ProportionalShare(total, numerator, denominator sdkmath.Int) sdkmath.Int {
	total = normalizeInt(total)
	numerator = normalizeInt(numerator)
	denominator = normalizeInt(denominator)
	if total.IsZero() || numerator.IsZero() || !denominator.IsPositive() {
		return sdkmath.ZeroInt()
	}
	return total.Mul(numerator).Quo(denominator)
}

func ApplyBps(amount sdkmath.Int, bps int64) sdkmath.Int {
	amount = normalizeInt(amount)
	if amount.IsZero() || bps == 0 {
		return sdkmath.ZeroInt()
	}
	return amount.MulRaw(bps).QuoRaw(BasisPoints)
}

func normalizeInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}

func validateBps(name string, value, min, max int64) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return nil
}

func clampInt64(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
