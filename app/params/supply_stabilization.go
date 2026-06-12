package params

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	SupplyPolicyLowerNetIssuance	= "lower_net_issuance"
	SupplyPolicyHigherNetIssuance	= "higher_net_issuance"
	SupplyPolicyHoldNetIssuance	= "hold_net_issuance"

	DefaultTargetNetIssuanceMinBps		= int64(50)
	DefaultTargetNetIssuanceMaxBps		= int64(300)
	DefaultLowerNetIssuanceStepBps		= int64(50)
	DefaultHigherNetIssuanceStepBps		= int64(75)
	DefaultFeeRevenueToMintThresholdBps	= int64(5_000)
	DefaultAdequateReserveCoverageBps	= int64(10_000)
	DefaultMaxLowSlashingRateBps		= int64(50)
	DefaultMaxHealthyValidatorAttritionBps	= int64(100)
	DefaultMinSupplyProjectionYears		= uint32(1)
	DefaultMaxSupplyProjectionYears		= uint32(5)
)

type SupplyStabilizationParams struct {
	TargetNetIssuanceMinBps		int64
	TargetNetIssuanceMaxBps		int64
	LowerNetIssuanceStepBps		int64
	HigherNetIssuanceStepBps	int64
	FeeRevenueToMintThresholdBps	int64
	StableBondedStakeMinBps		int64
	HealthyValidatorMinCount	uint64
	LowSlashingRateMaxBps		int64
	AdequateReserveCoverageBps	int64
	ValidatorAttritionMaxBps	int64
	MinProjectionYears		uint32
	MaxProjectionYears		uint32
}

type SupplyStabilizationInput struct {
	CurrentSupplyNaet			sdkmath.Int
	RecentAnnualGrossMintedNaet		sdkmath.Int
	RecentAnnualBurnedNaet			sdkmath.Int
	RecentAnnualFeeRevenueNaet		sdkmath.Int
	RecentAnnualValidatorRewardsNaet	sdkmath.Int
	BondedStakeRatioBps			int64
	ActiveValidatorCount			uint64
	SlashingRateBps				int64
	ReserveCoverageBps			int64
	ValidatorAttritionBps			int64
	SecurityRiskBps				int64
	ProjectionYears				uint32
	ConsensusRewardAccountingPreserved	bool
}

type SupplyStabilizationCondition struct {
	Name	string
	Met	bool
	Value	int64
	Target	int64
}

type SupplyProjectionYear struct {
	Year				uint32
	StartingSupplyNaet		sdkmath.Int
	ProjectedGrossMintedNaet	sdkmath.Int
	ProjectedBurnedNaet		sdkmath.Int
	ProjectedNetIssuanceNaet	sdkmath.Int
	ProjectedEndingSupplyNaet	sdkmath.Int
	ProjectedNetIssuanceBps		int64
	ValidatorSecuritySpendNaet	sdkmath.Int
}

type SupplyStabilizationReport struct {
	PolicyDirection				string
	CurrentNetIssuanceBps			int64
	TargetNetIssuanceBps			int64
	TargetNetIssuanceMinBps			int64
	TargetNetIssuanceMaxBps			int64
	LowerIssuanceConditions			[]SupplyStabilizationCondition
	HigherIssuanceConditions		[]SupplyStabilizationCondition
	ProjectionYears				[]SupplyProjectionYear
	GovernanceSummary			string
	ConsensusRewardAccountingPreserved	bool
	Passed					bool
	Failed					[]string
}

func DefaultSupplyStabilizationParams() SupplyStabilizationParams {
	return SupplyStabilizationParams{
		TargetNetIssuanceMinBps:	DefaultTargetNetIssuanceMinBps,
		TargetNetIssuanceMaxBps:	DefaultTargetNetIssuanceMaxBps,
		LowerNetIssuanceStepBps:	DefaultLowerNetIssuanceStepBps,
		HigherNetIssuanceStepBps:	DefaultHigherNetIssuanceStepBps,
		FeeRevenueToMintThresholdBps:	DefaultFeeRevenueToMintThresholdBps,
		StableBondedStakeMinBps:	DefaultTargetStakeBps - DefaultStakeTargetToleranceBps,
		HealthyValidatorMinCount:	DefaultActiveValidatorTarget,
		LowSlashingRateMaxBps:		DefaultMaxLowSlashingRateBps,
		AdequateReserveCoverageBps:	DefaultAdequateReserveCoverageBps,
		ValidatorAttritionMaxBps:	DefaultMaxHealthyValidatorAttritionBps,
		MinProjectionYears:		DefaultMinSupplyProjectionYears,
		MaxProjectionYears:		DefaultMaxSupplyProjectionYears,
	}
}

func GenerateSupplyStabilizationReport(input SupplyStabilizationInput, params SupplyStabilizationParams) (SupplyStabilizationReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return SupplyStabilizationReport{}, err
	}
	if err := input.Validate(params); err != nil {
		return SupplyStabilizationReport{}, err
	}

	supply := normalizeInt(input.CurrentSupplyNaet)
	grossMinted := normalizeInt(input.RecentAnnualGrossMintedNaet)
	burned := normalizeInt(input.RecentAnnualBurnedNaet)
	currentNet := grossMinted.Sub(burned)
	currentNetBps := int64(0)
	if supply.IsPositive() {
		currentNetBps = currentNet.MulRaw(BasisPoints).Quo(supply).Int64()
	}

	lowerConditions := lowerNetIssuanceConditions(input, params)
	higherConditions := higherNetIssuanceConditions(input, params)
	lowerReady := allSupplyConditionsMet(lowerConditions)
	higherNeeded := anySupplyConditionMet(higherConditions)

	targetBps := clampInt64(currentNetBps, params.TargetNetIssuanceMinBps, params.TargetNetIssuanceMaxBps)
	direction := SupplyPolicyHoldNetIssuance
	if higherNeeded {
		direction = SupplyPolicyHigherNetIssuance
		targetBps = clampInt64(targetBps+params.HigherNetIssuanceStepBps, params.TargetNetIssuanceMinBps, params.TargetNetIssuanceMaxBps)
	} else if lowerReady {
		direction = SupplyPolicyLowerNetIssuance
		targetBps = clampInt64(targetBps-params.LowerNetIssuanceStepBps, params.TargetNetIssuanceMinBps, params.TargetNetIssuanceMaxBps)
	}

	projections := projectSupply(input, targetBps)
	failed := make([]string, 0)
	if !input.ConsensusRewardAccountingPreserved {
		failed = append(failed, "consensus_reward_accounting_not_preserved")
	}
	if len(projections) == 0 {
		failed = append(failed, "supply_projection_missing")
	}
	return SupplyStabilizationReport{
		PolicyDirection:			direction,
		CurrentNetIssuanceBps:			currentNetBps,
		TargetNetIssuanceBps:			targetBps,
		TargetNetIssuanceMinBps:		params.TargetNetIssuanceMinBps,
		TargetNetIssuanceMaxBps:		params.TargetNetIssuanceMaxBps,
		LowerIssuanceConditions:		lowerConditions,
		HigherIssuanceConditions:		higherConditions,
		ProjectionYears:			projections,
		GovernanceSummary:			governanceSupplySummary(direction, currentNetBps, targetBps, input.ProjectionYears),
		ConsensusRewardAccountingPreserved:	input.ConsensusRewardAccountingPreserved,
		Passed:					len(failed) == 0,
		Failed:					failed,
	}, nil
}

func (p SupplyStabilizationParams) Validate() error {
	if err := validateBps("target_net_issuance_min_bps", p.TargetNetIssuanceMinBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("target_net_issuance_max_bps", p.TargetNetIssuanceMaxBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if p.TargetNetIssuanceMinBps > p.TargetNetIssuanceMaxBps {
		return fmt.Errorf("target_net_issuance_min_bps must be <= target_net_issuance_max_bps")
	}
	if err := validateBps("lower_net_issuance_step_bps", p.LowerNetIssuanceStepBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("higher_net_issuance_step_bps", p.HigherNetIssuanceStepBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("fee_revenue_to_mint_threshold_bps", p.FeeRevenueToMintThresholdBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("stable_bonded_stake_min_bps", p.StableBondedStakeMinBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.HealthyValidatorMinCount == 0 {
		return fmt.Errorf("healthy_validator_min_count must be positive")
	}
	if err := validateBps("low_slashing_rate_max_bps", p.LowSlashingRateMaxBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("adequate_reserve_coverage_bps", p.AdequateReserveCoverageBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("validator_attrition_max_bps", p.ValidatorAttritionMaxBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MinProjectionYears == 0 {
		return fmt.Errorf("min_projection_years must be positive")
	}
	if p.MaxProjectionYears < p.MinProjectionYears {
		return fmt.Errorf("max_projection_years must be >= min_projection_years")
	}
	return nil
}

func (p SupplyStabilizationParams) withDefaults() SupplyStabilizationParams {
	defaults := DefaultSupplyStabilizationParams()
	if p.TargetNetIssuanceMinBps == 0 {
		p.TargetNetIssuanceMinBps = defaults.TargetNetIssuanceMinBps
	}
	if p.TargetNetIssuanceMaxBps == 0 {
		p.TargetNetIssuanceMaxBps = defaults.TargetNetIssuanceMaxBps
	}
	if p.LowerNetIssuanceStepBps == 0 {
		p.LowerNetIssuanceStepBps = defaults.LowerNetIssuanceStepBps
	}
	if p.HigherNetIssuanceStepBps == 0 {
		p.HigherNetIssuanceStepBps = defaults.HigherNetIssuanceStepBps
	}
	if p.FeeRevenueToMintThresholdBps == 0 {
		p.FeeRevenueToMintThresholdBps = defaults.FeeRevenueToMintThresholdBps
	}
	if p.StableBondedStakeMinBps == 0 {
		p.StableBondedStakeMinBps = defaults.StableBondedStakeMinBps
	}
	if p.HealthyValidatorMinCount == 0 {
		p.HealthyValidatorMinCount = defaults.HealthyValidatorMinCount
	}
	if p.LowSlashingRateMaxBps == 0 {
		p.LowSlashingRateMaxBps = defaults.LowSlashingRateMaxBps
	}
	if p.AdequateReserveCoverageBps == 0 {
		p.AdequateReserveCoverageBps = defaults.AdequateReserveCoverageBps
	}
	if p.ValidatorAttritionMaxBps == 0 {
		p.ValidatorAttritionMaxBps = defaults.ValidatorAttritionMaxBps
	}
	if p.MinProjectionYears == 0 {
		p.MinProjectionYears = defaults.MinProjectionYears
	}
	if p.MaxProjectionYears == 0 {
		p.MaxProjectionYears = defaults.MaxProjectionYears
	}
	return p
}

func (input SupplyStabilizationInput) Validate(params SupplyStabilizationParams) error {
	for _, item := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "current_supply_naet", value: input.CurrentSupplyNaet},
		{name: "recent_annual_gross_minted_naet", value: input.RecentAnnualGrossMintedNaet},
		{name: "recent_annual_burned_naet", value: input.RecentAnnualBurnedNaet},
		{name: "recent_annual_fee_revenue_naet", value: input.RecentAnnualFeeRevenueNaet},
		{name: "recent_annual_validator_rewards_naet", value: input.RecentAnnualValidatorRewardsNaet},
	} {
		value := normalizeInt(item.value)
		if value.IsNegative() {
			return fmt.Errorf("%s must not be negative", item.name)
		}
	}
	if !normalizeInt(input.CurrentSupplyNaet).IsPositive() {
		return fmt.Errorf("current_supply_naet must be positive")
	}
	if err := validateBps("bonded_stake_ratio_bps", input.BondedStakeRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("slashing_rate_bps", input.SlashingRateBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("reserve_coverage_bps", input.ReserveCoverageBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("validator_attrition_bps", input.ValidatorAttritionBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("security_risk_bps", input.SecurityRiskBps, 0, BasisPoints); err != nil {
		return err
	}
	if input.ProjectionYears < params.MinProjectionYears || input.ProjectionYears > params.MaxProjectionYears {
		return fmt.Errorf("projection_years must be between %d and %d", params.MinProjectionYears, params.MaxProjectionYears)
	}
	return nil
}

func lowerNetIssuanceConditions(input SupplyStabilizationInput, params SupplyStabilizationParams) []SupplyStabilizationCondition {
	feeToMintBps := ratioBps(normalizeInt(input.RecentAnnualFeeRevenueNaet), normalizeInt(input.RecentAnnualGrossMintedNaet))
	return []SupplyStabilizationCondition{
		{Name: "sustained_fee_revenue", Met: feeToMintBps >= params.FeeRevenueToMintThresholdBps, Value: feeToMintBps, Target: params.FeeRevenueToMintThresholdBps},
		{Name: "stable_bonded_stake", Met: input.BondedStakeRatioBps >= params.StableBondedStakeMinBps, Value: input.BondedStakeRatioBps, Target: params.StableBondedStakeMinBps},
		{Name: "healthy_validator_set", Met: input.ActiveValidatorCount >= params.HealthyValidatorMinCount, Value: int64(input.ActiveValidatorCount), Target: int64(params.HealthyValidatorMinCount)},
		{Name: "low_slashing_rate", Met: input.SlashingRateBps <= params.LowSlashingRateMaxBps, Value: input.SlashingRateBps, Target: params.LowSlashingRateMaxBps},
		{Name: "adequate_reserves", Met: input.ReserveCoverageBps >= params.AdequateReserveCoverageBps, Value: input.ReserveCoverageBps, Target: params.AdequateReserveCoverageBps},
	}
}

func higherNetIssuanceConditions(input SupplyStabilizationInput, params SupplyStabilizationParams) []SupplyStabilizationCondition {
	feeToMintBps := ratioBps(normalizeInt(input.RecentAnnualFeeRevenueNaet), normalizeInt(input.RecentAnnualGrossMintedNaet))
	return []SupplyStabilizationCondition{
		{Name: "low_bonded_stake", Met: input.BondedStakeRatioBps < params.StableBondedStakeMinBps, Value: input.BondedStakeRatioBps, Target: params.StableBondedStakeMinBps},
		{Name: "validator_attrition", Met: input.ValidatorAttritionBps > params.ValidatorAttritionMaxBps, Value: input.ValidatorAttritionBps, Target: params.ValidatorAttritionMaxBps},
		{Name: "low_fee_revenue", Met: feeToMintBps < params.FeeRevenueToMintThresholdBps, Value: feeToMintBps, Target: params.FeeRevenueToMintThresholdBps},
		{Name: "elevated_security_risk", Met: input.SecurityRiskBps > 0 || input.SlashingRateBps > params.LowSlashingRateMaxBps, Value: maxInt64(input.SecurityRiskBps, input.SlashingRateBps), Target: params.LowSlashingRateMaxBps},
	}
}

func allSupplyConditionsMet(conditions []SupplyStabilizationCondition) bool {
	for _, condition := range conditions {
		if !condition.Met {
			return false
		}
	}
	return len(conditions) > 0
}

func anySupplyConditionMet(conditions []SupplyStabilizationCondition) bool {
	for _, condition := range conditions {
		if condition.Met {
			return true
		}
	}
	return false
}

func projectSupply(input SupplyStabilizationInput, targetNetIssuanceBps int64) []SupplyProjectionYear {
	years := make([]SupplyProjectionYear, 0, input.ProjectionYears)
	startingSupply := normalizeInt(input.CurrentSupplyNaet)
	recentBurned := normalizeInt(input.RecentAnnualBurnedNaet)
	securitySpend := normalizeInt(input.RecentAnnualValidatorRewardsNaet)
	for year := uint32(1); year <= input.ProjectionYears; year++ {
		targetNet := startingSupply.MulRaw(targetNetIssuanceBps).QuoRaw(BasisPoints)
		projectedMint := targetNet.Add(recentBurned)
		endingSupply := startingSupply.Add(targetNet)
		years = append(years, SupplyProjectionYear{
			Year:				year,
			StartingSupplyNaet:		startingSupply,
			ProjectedGrossMintedNaet:	projectedMint,
			ProjectedBurnedNaet:		recentBurned,
			ProjectedNetIssuanceNaet:	targetNet,
			ProjectedEndingSupplyNaet:	endingSupply,
			ProjectedNetIssuanceBps:	targetNetIssuanceBps,
			ValidatorSecuritySpendNaet:	securitySpend,
		})
		startingSupply = endingSupply
	}
	return years
}

func ratioBps(numerator, denominator sdkmath.Int) int64 {
	numerator = normalizeInt(numerator)
	denominator = normalizeInt(denominator)
	if numerator.IsZero() || !denominator.IsPositive() {
		return 0
	}
	return numerator.MulRaw(BasisPoints).Quo(denominator).Int64()
}

func governanceSupplySummary(direction string, currentBps, targetBps int64, years uint32) string {
	return fmt.Sprintf("policy=%s current_net_issuance_bps=%d target_net_issuance_bps=%d projection_years=%d", direction, currentBps, targetBps, years)
}
