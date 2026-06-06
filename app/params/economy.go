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

	DefaultCircuitBreakerFeeSpikeBps        = int64(20_000)
	DefaultCircuitBreakerControllerDriftBps = int64(1_000)
	DefaultCircuitBreakerFailedTxRateBps    = int64(2_000)
	DefaultInflationNoiseToleranceBps       = int64(500)
	DefaultCleanupRewardBps                 = int64(2_000)
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

type EconomicInvariantInput struct {
	StakingDenom                  string
	FeeDenom                      string
	RewardDenom                   string
	SlashingDenom                 string
	ExecutionChargeDenom          string
	CirculatingSupply             sdkmath.Int
	AnnualMint                    sdkmath.Int
	AnnualBurn                    sdkmath.Int
	MaxNetIssuanceBps             int64
	MaxNetBurnBps                 int64
	ControllerParams              BalanceControllerParams
	ControllerOutput              BalanceControllerOutput
	FeeFlow                       ProtocolEconomicFlowOutput
	MaxBlockFeeNaet               sdkmath.Int
	BlockFeeNaet                  sdkmath.Int
	ValidatorRewardsDeterministic bool
	FeeComputationDeterministic   bool
	SlashingDeterministic         bool
	SlashingAuditable             bool
	SlashingRewardSafe            bool
	ControllerParamsExposed       bool
	ControllerStateExposed        bool
	ControllerEventsExposed       bool
	StorageFeePerByteNaet         sdkmath.Int
	LongLivedStorageBytes         int64
	StorageRetentionPeriods       int64
	TransientExecutionChargeNaet  sdkmath.Int
}

type EconomicInvariantReport struct {
	Passed           bool
	FailedInvariants []string
}

type EconomicWeaknessControlInput struct {
	BurnControllerWired               bool
	InflationUsesNetworkActivity      bool
	DeflationGuardEnforced            bool
	SlashingFlowIntegrated            bool
	EpochValidatorSelectionProduction bool
	AVMFeesInGlobalFeeMarket          bool
	StateRentOrCleanupIncentive       bool
	ValidatorReputationInDelegation   bool
	StakeConcentrationDampening       bool
	EconomicCircuitBreakerEnabled     bool
}

type EconomicWeaknessControlReport struct {
	ProductionReady bool
	MissingControls []string
}

type InflationRiskInput struct {
	CirculatingSupply              sdkmath.Int
	AnnualMint                     sdkmath.Int
	AnnualBurn                     sdkmath.Int
	ValidatorRewardPoolNaet        sdkmath.Int
	ValidatorOperatingCostNaet     sdkmath.Int
	CurrentInflationBps            int64
	StakeRatioBps                  int64
	TopValidatorStakeBps           int64
	DelegatorRiskSignalCoverageBps int64
	ActivitySamplesBps             []int64
	BurnIntegratedWithIssuance     bool
	NetIssuanceTargetConfigured    bool
	MaxNetIssuanceBps              int64
	ActivityNoiseToleranceBps      int64
}

type InflationRiskReport struct {
	Stable                bool
	NetIssuanceBps        int64
	RewardCoverageBps     int64
	ActivityVolatilityBps int64
	Risks                 []string
}

type EconomicCircuitBreakerParams struct {
	MaxBlockLoadBps       int64
	MaxFeeSpikeBps        int64
	MaxControllerDriftBps int64
	MaxFailedTxRateBps    int64
	MaxBurnToMintBps      int64
	MinCooldownBlocks     uint64
}

type EconomicCircuitBreakerInput struct {
	BlockLoadBps       int64
	FeeSpikeBps        int64
	ControllerDriftBps int64
	FailedTxRateBps    int64
	BurnToMintBps      int64
}

type EconomicCircuitBreakerOutput struct {
	Active         bool
	Reasons        []string
	CooldownBlocks uint64
}

type SlashingEconomyFlowInput struct {
	PenaltyNaet       sdkmath.Int
	BurnRatioBps      int64
	TreasuryRatioBps  int64
	ReporterRewardBps int64
}

type SlashingEconomyFlowOutput struct {
	PenaltyNaet        sdkmath.Int
	BurnNaet           sdkmath.Int
	TreasuryNaet       sdkmath.Int
	ReporterRewardNaet sdkmath.Int
	ValidatorPoolNaet  sdkmath.Int
}

type StateRentInput struct {
	StorageBytes          int64
	RetentionPeriods      int64
	FeePerByteNaet        sdkmath.Int
	CleanupRewardRatioBps int64
	CleanupEligible       bool
}

type StateRentOutput struct {
	RentNaet          sdkmath.Int
	CleanupRewardNaet sdkmath.Int
	BurnableRentNaet  sdkmath.Int
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

func DefaultEconomicCircuitBreakerParams() EconomicCircuitBreakerParams {
	return EconomicCircuitBreakerParams{
		MaxBlockLoadBps:       HighCongestionLoadBps,
		MaxFeeSpikeBps:        DefaultCircuitBreakerFeeSpikeBps,
		MaxControllerDriftBps: DefaultCircuitBreakerControllerDriftBps,
		MaxFailedTxRateBps:    DefaultCircuitBreakerFailedTxRateBps,
		MaxBurnToMintBps:      DeflationGuardBurnToMintBps,
		MinCooldownBlocks:     1,
	}
}

func EvaluateEconomicWeaknessControls(input EconomicWeaknessControlInput) EconomicWeaknessControlReport {
	failed := make([]string, 0, 10)
	for _, item := range []struct {
		name string
		ok   bool
	}{
		{name: "burn_controller_not_wired_to_fee_reward_flow", ok: input.BurnControllerWired},
		{name: "inflation_controller_not_activity_coupled", ok: input.InflationUsesNetworkActivity},
		{name: "deflation_guard_not_enforced", ok: input.DeflationGuardEnforced},
		{name: "slashing_flow_not_integrated", ok: input.SlashingFlowIntegrated},
		{name: "epoch_validator_selection_not_productionized", ok: input.EpochValidatorSelectionProduction},
		{name: "avm_fees_not_in_global_market", ok: input.AVMFeesInGlobalFeeMarket},
		{name: "state_rent_or_cleanup_missing", ok: input.StateRentOrCleanupIncentive},
		{name: "validator_reputation_not_in_delegation", ok: input.ValidatorReputationInDelegation},
		{name: "stake_concentration_dampening_missing", ok: input.StakeConcentrationDampening},
		{name: "economic_circuit_breaker_missing", ok: input.EconomicCircuitBreakerEnabled},
	} {
		if !item.ok {
			failed = append(failed, item.name)
		}
	}
	return EconomicWeaknessControlReport{
		ProductionReady: len(failed) == 0,
		MissingControls: failed,
	}
}

func EvaluateInflationRisks(input InflationRiskInput) (InflationRiskReport, error) {
	if input.MaxNetIssuanceBps == 0 {
		input.MaxNetIssuanceBps = MaxInflationBps
	}
	if input.ActivityNoiseToleranceBps == 0 {
		input.ActivityNoiseToleranceBps = DefaultInflationNoiseToleranceBps
	}
	if err := validateBps("current_inflation_bps", input.CurrentInflationBps, MinInflationBps, MaxInflationBps); err != nil {
		return InflationRiskReport{}, err
	}
	if err := validateBps("stake_ratio_bps", input.StakeRatioBps, 0, BasisPoints); err != nil {
		return InflationRiskReport{}, err
	}
	if err := validateBps("top_validator_stake_bps", input.TopValidatorStakeBps, 0, BasisPoints); err != nil {
		return InflationRiskReport{}, err
	}
	if err := validateBps("delegator_risk_signal_coverage_bps", input.DelegatorRiskSignalCoverageBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return InflationRiskReport{}, err
	}
	if err := validateBps("max_net_issuance_bps", input.MaxNetIssuanceBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return InflationRiskReport{}, err
	}
	if err := validateBps("activity_noise_tolerance_bps", input.ActivityNoiseToleranceBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return InflationRiskReport{}, err
	}
	for i, sample := range input.ActivitySamplesBps {
		if err := validateBps(fmt.Sprintf("activity_samples_bps[%d]", i), sample, 0, BasisPoints); err != nil {
			return InflationRiskReport{}, err
		}
	}
	for _, item := range []struct {
		name  string
		value sdkmath.Int
	}{
		{name: "circulating_supply", value: input.CirculatingSupply},
		{name: "annual_mint", value: input.AnnualMint},
		{name: "annual_burn", value: input.AnnualBurn},
		{name: "validator_reward_pool_naet", value: input.ValidatorRewardPoolNaet},
		{name: "validator_operating_cost_naet", value: input.ValidatorOperatingCostNaet},
	} {
		if normalizeInt(item.value).IsNegative() {
			return InflationRiskReport{}, fmt.Errorf("%s must not be negative", item.name)
		}
	}

	supply := normalizeInt(input.CirculatingSupply)
	annualMint := normalizeInt(input.AnnualMint)
	annualBurn := normalizeInt(input.AnnualBurn)
	rewardPool := normalizeInt(input.ValidatorRewardPoolNaet)
	operatingCost := normalizeInt(input.ValidatorOperatingCostNaet)

	netIssuanceBps := int64(0)
	if supply.IsPositive() {
		net := annualMint.Sub(annualBurn)
		if net.IsPositive() {
			netIssuanceBps = net.MulRaw(BasisPoints).Quo(supply).Int64()
		}
	} else if annualMint.IsPositive() || annualBurn.IsPositive() {
		return InflationRiskReport{}, fmt.Errorf("circulating_supply is required when mint or burn is positive")
	}

	rewardCoverageBps := int64(0)
	if operatingCost.IsPositive() {
		rewardCoverageBps = rewardPool.MulRaw(BasisPoints).Quo(operatingCost).Int64()
	}
	volatility := activityVolatilityBps(input.ActivitySamplesBps)

	risks := make([]string, 0, 7)
	if !input.NetIssuanceTargetConfigured {
		risks = append(risks, "net_issuance_target_missing")
	} else if netIssuanceBps > input.MaxNetIssuanceBps {
		risks = append(risks, "net_issuance_outside_target")
	}
	if averageBps(input.ActivitySamplesBps) < DefaultTargetLoadBps/2 &&
		input.StakeRatioBps >= DefaultTargetStakeBps &&
		input.CurrentInflationBps >= DefaultTargetInflationBps {
		risks = append(risks, "security_overpaid_during_low_activity")
	}
	if operatingCost.IsPositive() && rewardCoverageBps < BasisPoints {
		risks = append(risks, "validator_security_underpaid")
	}
	if input.TopValidatorStakeBps > MaxTopValidatorConcentrationBps || input.DelegatorRiskSignalCoverageBps < MinDelegatorRiskSignalCoverageBps {
		risks = append(risks, "stake_target_risk_not_priced")
	}
	if volatility > input.ActivityNoiseToleranceBps {
		risks = append(risks, "inflation_activity_signal_noisy")
	}
	if !input.BurnIntegratedWithIssuance {
		risks = append(risks, "burn_not_integrated_with_issuance")
	}

	return InflationRiskReport{
		Stable:                len(risks) == 0,
		NetIssuanceBps:        netIssuanceBps,
		RewardCoverageBps:     rewardCoverageBps,
		ActivityVolatilityBps: volatility,
		Risks:                 risks,
	}, nil
}

func EvaluateEconomicCircuitBreaker(input EconomicCircuitBreakerInput, params EconomicCircuitBreakerParams) (EconomicCircuitBreakerOutput, error) {
	if params == (EconomicCircuitBreakerParams{}) {
		params = DefaultEconomicCircuitBreakerParams()
	}
	if err := params.Validate(); err != nil {
		return EconomicCircuitBreakerOutput{}, err
	}
	for _, item := range []struct {
		name  string
		value int64
	}{
		{name: "block_load_bps", value: input.BlockLoadBps},
		{name: "failed_tx_rate_bps", value: input.FailedTxRateBps},
		{name: "burn_to_mint_bps", value: input.BurnToMintBps},
	} {
		if err := validateBps(item.name, item.value, 0, DefaultMaxLoadMultiplierBps); err != nil {
			return EconomicCircuitBreakerOutput{}, err
		}
	}
	if input.FeeSpikeBps < 0 {
		return EconomicCircuitBreakerOutput{}, fmt.Errorf("fee_spike_bps must not be negative")
	}
	if input.ControllerDriftBps < 0 {
		return EconomicCircuitBreakerOutput{}, fmt.Errorf("controller_drift_bps must not be negative")
	}

	reasons := make([]string, 0, 5)
	if input.BlockLoadBps >= params.MaxBlockLoadBps {
		reasons = append(reasons, "block_load_abnormal")
	}
	if input.FeeSpikeBps > params.MaxFeeSpikeBps {
		reasons = append(reasons, "fee_spike_abnormal")
	}
	if input.ControllerDriftBps > params.MaxControllerDriftBps {
		reasons = append(reasons, "controller_drift_abnormal")
	}
	if input.FailedTxRateBps > params.MaxFailedTxRateBps {
		reasons = append(reasons, "failed_tx_rate_abnormal")
	}
	if input.BurnToMintBps > params.MaxBurnToMintBps {
		reasons = append(reasons, "burn_pressure_abnormal")
	}
	return EconomicCircuitBreakerOutput{
		Active:         len(reasons) > 0,
		Reasons:        reasons,
		CooldownBlocks: params.MinCooldownBlocks,
	}, nil
}

func (p EconomicCircuitBreakerParams) Validate() error {
	if err := validateBps("max_block_load_bps", p.MaxBlockLoadBps, 1, BasisPoints); err != nil {
		return err
	}
	if p.MaxFeeSpikeBps < 0 {
		return fmt.Errorf("max_fee_spike_bps must not be negative")
	}
	if p.MaxControllerDriftBps < 0 {
		return fmt.Errorf("max_controller_drift_bps must not be negative")
	}
	if err := validateBps("max_failed_tx_rate_bps", p.MaxFailedTxRateBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_burn_to_mint_bps", p.MaxBurnToMintBps, BasisPoints, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if p.MinCooldownBlocks == 0 {
		return fmt.Errorf("min_cooldown_blocks must be positive")
	}
	return nil
}

func ComputeSlashingEconomyFlow(input SlashingEconomyFlowInput) (SlashingEconomyFlowOutput, error) {
	penalty := normalizeInt(input.PenaltyNaet)
	if penalty.IsNegative() {
		return SlashingEconomyFlowOutput{}, fmt.Errorf("penalty_naet must not be negative")
	}
	for _, item := range []struct {
		name  string
		value int64
	}{
		{name: "burn_ratio_bps", value: input.BurnRatioBps},
		{name: "treasury_ratio_bps", value: input.TreasuryRatioBps},
		{name: "reporter_reward_bps", value: input.ReporterRewardBps},
	} {
		if err := validateBps(item.name, item.value, 0, BasisPoints); err != nil {
			return SlashingEconomyFlowOutput{}, err
		}
	}
	if input.BurnRatioBps+input.TreasuryRatioBps+input.ReporterRewardBps > BasisPoints {
		return SlashingEconomyFlowOutput{}, fmt.Errorf("slashing output ratios exceed 100%%")
	}
	burn := ApplyBps(penalty, input.BurnRatioBps)
	treasury := ApplyBps(penalty, input.TreasuryRatioBps)
	reporter := ApplyBps(penalty, input.ReporterRewardBps)
	return SlashingEconomyFlowOutput{
		PenaltyNaet:        penalty,
		BurnNaet:           burn,
		TreasuryNaet:       treasury,
		ReporterRewardNaet: reporter,
		ValidatorPoolNaet:  penalty.Sub(burn).Sub(treasury).Sub(reporter),
	}, nil
}

func ComputeStateRent(input StateRentInput) (StateRentOutput, error) {
	if input.StorageBytes < 0 {
		return StateRentOutput{}, fmt.Errorf("storage_bytes must not be negative")
	}
	if input.RetentionPeriods < 0 {
		return StateRentOutput{}, fmt.Errorf("retention_periods must not be negative")
	}
	if input.CleanupRewardRatioBps == 0 {
		input.CleanupRewardRatioBps = DefaultCleanupRewardBps
	}
	if err := validateBps("cleanup_reward_ratio_bps", input.CleanupRewardRatioBps, 0, BasisPoints); err != nil {
		return StateRentOutput{}, err
	}
	fee := normalizeInt(input.FeePerByteNaet)
	if fee.IsNegative() {
		return StateRentOutput{}, fmt.Errorf("fee_per_byte_naet must not be negative")
	}
	rent := fee.MulRaw(input.StorageBytes).MulRaw(input.RetentionPeriods)
	cleanupReward := sdkmath.ZeroInt()
	if input.CleanupEligible {
		cleanupReward = ApplyBps(rent, input.CleanupRewardRatioBps)
	}
	return StateRentOutput{
		RentNaet:          rent,
		CleanupRewardNaet: cleanupReward,
		BurnableRentNaet:  rent.Sub(cleanupReward),
	}, nil
}

func EvaluateEconomicInvariants(input EconomicInvariantInput) (EconomicInvariantReport, error) {
	if input.ControllerParams == (BalanceControllerParams{}) {
		input.ControllerParams = DefaultBalanceControllerParams()
	}
	if input.MaxNetIssuanceBps == 0 {
		input.MaxNetIssuanceBps = MaxInflationBps
	}
	if input.MaxNetBurnBps == 0 {
		input.MaxNetBurnBps = DeflationGuardBurnToMintBps
	}
	if input.LongLivedStorageBytes == 0 {
		input.LongLivedStorageBytes = 1
	}
	if input.StorageRetentionPeriods == 0 {
		input.StorageRetentionPeriods = 2
	}
	if err := input.ControllerParams.Validate(); err != nil {
		return EconomicInvariantReport{}, err
	}
	if err := validateBps("controller_output_inflation_bps", input.ControllerOutput.InflationBps, input.ControllerParams.MinInflationBps, input.ControllerParams.MaxInflationBps); err != nil {
		return EconomicInvariantReport{}, err
	}
	if err := validateBps("controller_output_burn_ratio_bps", input.ControllerOutput.BurnRatioBps, input.ControllerParams.MinBurnRatioBps, input.ControllerParams.MaxBurnRatioBps); err != nil {
		return EconomicInvariantReport{}, err
	}
	if err := validateBps("controller_output_validator_fee_ratio_bps", input.ControllerOutput.ValidatorFeeRatioBps, 0, BasisPoints); err != nil {
		return EconomicInvariantReport{}, err
	}
	if input.ControllerOutput.BurnRatioBps+input.ControllerParams.TreasuryFeeRatioBps+input.ControllerOutput.ValidatorFeeRatioBps != BasisPoints {
		return EconomicInvariantReport{}, fmt.Errorf("controller fee ratios must sum to 10000 bps")
	}
	if err := validateBps("max_net_issuance_bps", input.MaxNetIssuanceBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return EconomicInvariantReport{}, err
	}
	if err := validateBps("max_net_burn_bps", input.MaxNetBurnBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return EconomicInvariantReport{}, err
	}
	if input.LongLivedStorageBytes < 0 {
		return EconomicInvariantReport{}, fmt.Errorf("long_lived_storage_bytes must not be negative")
	}
	if input.StorageRetentionPeriods < 0 {
		return EconomicInvariantReport{}, fmt.Errorf("storage_retention_periods must not be negative")
	}
	for _, item := range []struct {
		name  string
		value sdkmath.Int
	}{
		{name: "circulating_supply", value: input.CirculatingSupply},
		{name: "annual_mint", value: input.AnnualMint},
		{name: "annual_burn", value: input.AnnualBurn},
		{name: "max_block_fee_naet", value: input.MaxBlockFeeNaet},
		{name: "block_fee_naet", value: input.BlockFeeNaet},
		{name: "storage_fee_per_byte_naet", value: input.StorageFeePerByteNaet},
		{name: "transient_execution_charge_naet", value: input.TransientExecutionChargeNaet},
	} {
		value := normalizeInt(item.value)
		if value.IsNegative() {
			return EconomicInvariantReport{}, fmt.Errorf("%s must not be negative", item.name)
		}
	}

	failed := make([]string, 0, 8)
	for _, item := range []struct {
		name  string
		denom string
	}{
		{name: "staking", denom: input.StakingDenom},
		{name: "fees", denom: input.FeeDenom},
		{name: "rewards", denom: input.RewardDenom},
		{name: "slashing", denom: input.SlashingDenom},
		{name: "execution_charges", denom: input.ExecutionChargeDenom},
	} {
		if item.denom != BaseDenom {
			failed = append(failed, item.name+"_not_aet_primary_asset")
		}
	}

	supply := normalizeInt(input.CirculatingSupply)
	annualMint := normalizeInt(input.AnnualMint)
	annualBurn := normalizeInt(input.AnnualBurn)
	if supply.IsPositive() {
		netIssuance := annualMint.Sub(annualBurn)
		if netIssuance.IsPositive() && netIssuance.MulRaw(BasisPoints).Quo(supply).GT(sdkmath.NewInt(input.MaxNetIssuanceBps)) {
			failed = append(failed, "net_issuance_outside_bounds")
		}
		netBurn := annualBurn.Sub(annualMint)
		if netBurn.IsPositive() && netBurn.MulRaw(BasisPoints).Quo(supply).GT(sdkmath.NewInt(input.MaxNetBurnBps)) {
			failed = append(failed, "net_burn_outside_bounds")
		}
	} else if annualMint.IsPositive() || annualBurn.IsPositive() {
		failed = append(failed, "supply_required_for_net_bounds")
	}

	if !input.ValidatorRewardsDeterministic {
		failed = append(failed, "validator_rewards_not_deterministic")
	}
	if !input.FeeComputationDeterministic {
		failed = append(failed, "fee_computation_not_deterministic")
	}
	if normalizeInt(input.BlockFeeNaet).GT(normalizeInt(input.MaxBlockFeeNaet)) {
		failed = append(failed, "block_fee_exceeds_bound")
	}
	if !input.SlashingDeterministic || !input.SlashingAuditable || !input.SlashingRewardSafe {
		failed = append(failed, "slashing_invariant_not_satisfied")
	}
	if !input.ControllerParamsExposed || !input.ControllerStateExposed || !input.ControllerEventsExposed {
		failed = append(failed, "adaptive_controller_not_observable")
	}
	if err := input.FeeFlow.Validate(); err != nil {
		failed = append(failed, "economic_flow_not_conservative")
	}
	longLivedStorageCost := normalizeInt(input.StorageFeePerByteNaet).MulRaw(input.LongLivedStorageBytes).MulRaw(input.StorageRetentionPeriods)
	if !longLivedStorageCost.GT(normalizeInt(input.TransientExecutionChargeNaet)) {
		failed = append(failed, "storage_pricing_not_above_transient_execution")
	}

	return EconomicInvariantReport{
		Passed:           len(failed) == 0,
		FailedInvariants: failed,
	}, nil
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

func (f ProtocolEconomicFlowOutput) Validate() error {
	for _, item := range []struct {
		name  string
		value sdkmath.Int
	}{
		{name: "total_charges_naet", value: f.TotalChargesNaet},
		{name: "burn_naet", value: f.BurnNaet},
		{name: "treasury_naet", value: f.TreasuryNaet},
		{name: "validator_rewards_naet", value: f.ValidatorRewardsNaet},
	} {
		value := normalizeInt(item.value)
		if value.IsNegative() {
			return fmt.Errorf("%s must not be negative", item.name)
		}
	}
	total := normalizeInt(f.TotalChargesNaet)
	targets := normalizeInt(f.BurnNaet).
		Add(normalizeInt(f.TreasuryNaet)).
		Add(normalizeInt(f.ValidatorRewardsNaet))
	if !total.Equal(targets) {
		return fmt.Errorf("economic flow must conserve charges")
	}
	return nil
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

func averageBps(values []int64) int64 {
	if len(values) == 0 {
		return DefaultTargetLoadBps
	}
	sum := int64(0)
	for _, value := range values {
		sum += value
	}
	return sum / int64(len(values))
}

func activityVolatilityBps(values []int64) int64 {
	if len(values) < 2 {
		return 0
	}
	minValue := values[0]
	maxValue := values[0]
	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue - minValue
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
