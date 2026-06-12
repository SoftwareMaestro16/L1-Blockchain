package params

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	DefaultInflationPerWindowChangeLimitBps	= int64(25)
	DefaultInflationSmoothingWindow		= uint32(4)
	DefaultOperatingCostCoverageTargetBps	= BasisPoints
	DefaultTreasuryReserveHealthTargetBps	= BasisPoints
	DefaultFeeRevenueTargetNaet		= int64(1_000_000_000)
	DefaultActiveValidatorTarget		= uint64(75)

	InflationScenarioLowActivity		= "low_activity"
	InflationScenarioNormalActivity		= "normal_activity"
	InflationScenarioHighActivity		= "high_activity"
	InflationScenarioAdversarialActivity	= "adversarial_activity"
)

type ActivityInflationControllerParams struct {
	MinInflationBps			int64
	MaxInflationBps			int64
	TargetStakeBps			int64
	TargetOperatingCostCoverageBps	int64
	TargetFeeRevenueNaet		sdkmath.Int
	TargetActiveValidators		uint64
	TargetTreasuryReserveHealthBps	int64
	StakeWeightBps			int64
	OperatingCostWeightBps		int64
	FeeRevenueWeightBps		int64
	ValidatorCountWeightBps		int64
	SlashingRiskWeightBps		int64
	NetworkActivityWeightBps	int64
	TreasuryReserveWeightBps	int64
	PerWindowChangeLimitBps		int64
	SmoothingWindow			uint32
	EmergencyFreeze			bool
}

type ActivityInflationControllerInput struct {
	CurrentInflationBps		int64
	BondedStakeRatioBps		int64
	ValidatorOperatingCostIndexBps	int64
	FeeRevenueNaet			sdkmath.Int
	ActiveValidatorCount		uint64
	SlashingRiskEvents		uint32
	NetworkActivityScoreBps		int64
	TreasuryReserveHealthBps	int64
	RecentInflationBps		[]int64
}

type InflationAdjustmentComponent struct {
	Name		string
	ValueBps	int64
	WeightBps	int64
	ContributionBps	int64
}

type ActivityInflationControllerOutput struct {
	InflationBps		int64
	PreviousInflationBps	int64
	RawTargetInflationBps	int64
	SmoothedInflationBps	int64
	AppliedDeltaBps		int64
	BoundedByMin		bool
	BoundedByMax		bool
	ChangeLimited		bool
	EmergencyFrozen		bool
	Components		[]InflationAdjustmentComponent
	QueryableInputs		ActivityInflationControllerInput
}

type NetIssuanceInput struct {
	EpochID				uint64
	AccountingPeriod		string
	Blocks				uint64
	GrossMintedNaet			sdkmath.Int
	BurnedNaet			sdkmath.Int
	FeeRevenueNaet			sdkmath.Int
	ValidatorSecuritySpendNaet	sdkmath.Int
}

type NetIssuanceReport struct {
	EpochID				uint64
	AccountingPeriod		string
	Blocks				uint64
	GrossMintedNaet			sdkmath.Int
	BurnedNaet			sdkmath.Int
	NetSupplyChangeNaet		sdkmath.Int
	FeeRevenueNaet			sdkmath.Int
	ValidatorSecuritySpendNaet	sdkmath.Int
	SecuritySpendPerBlockNaet	sdkmath.Int
}

type InflationSimulationStep struct {
	Scenario	string
	Controller	ActivityInflationControllerInput
	NetIssuance	NetIssuanceInput
}

type InflationSimulationReport struct {
	ScenarioCount		int
	FinalInflationBps	int64
	MinObservedInflationBps	int64
	MaxObservedInflationBps	int64
	ControllerOutputs	[]ActivityInflationControllerOutput
	NetIssuanceReports	[]NetIssuanceReport
	Passed			bool
	Risks			[]string
}

func DefaultActivityInflationControllerParams() ActivityInflationControllerParams {
	return ActivityInflationControllerParams{
		MinInflationBps:		MinInflationBps,
		MaxInflationBps:		MaxInflationBps,
		TargetStakeBps:			DefaultTargetStakeBps,
		TargetOperatingCostCoverageBps:	DefaultOperatingCostCoverageTargetBps,
		TargetFeeRevenueNaet:		sdkmath.NewInt(DefaultFeeRevenueTargetNaet),
		TargetActiveValidators:		DefaultActiveValidatorTarget,
		TargetTreasuryReserveHealthBps:	DefaultTreasuryReserveHealthTargetBps,
		StakeWeightBps:			2_500,
		OperatingCostWeightBps:		2_000,
		FeeRevenueWeightBps:		1_500,
		ValidatorCountWeightBps:	1_000,
		SlashingRiskWeightBps:		1_000,
		NetworkActivityWeightBps:	1_000,
		TreasuryReserveWeightBps:	1_000,
		PerWindowChangeLimitBps:	DefaultInflationPerWindowChangeLimitBps,
		SmoothingWindow:		DefaultInflationSmoothingWindow,
	}
}

func ActivityInflationController(input ActivityInflationControllerInput) (ActivityInflationControllerOutput, error) {
	return ActivityInflationControllerWithParams(input, DefaultActivityInflationControllerParams())
}

func ActivityInflationControllerWithParams(input ActivityInflationControllerInput, params ActivityInflationControllerParams) (ActivityInflationControllerOutput, error) {
	if err := params.Validate(); err != nil {
		return ActivityInflationControllerOutput{}, err
	}
	if err := input.Validate(params); err != nil {
		return ActivityInflationControllerOutput{}, err
	}

	components := inflationAdjustmentComponents(input, params)
	weightedPressureBps := int64(0)
	for _, component := range components {
		weightedPressureBps += component.ContributionBps
	}

	rawTarget := clampInt64(input.CurrentInflationBps+(weightedPressureBps/10), params.MinInflationBps, params.MaxInflationBps)
	smoothed := smoothInflationTarget(rawTarget, input.RecentInflationBps, params.SmoothingWindow)
	delta := smoothed - input.CurrentInflationBps
	limitedDelta := clampInt64(delta, -params.PerWindowChangeLimitBps, params.PerWindowChangeLimitBps)
	inflation := clampInt64(input.CurrentInflationBps+limitedDelta, params.MinInflationBps, params.MaxInflationBps)

	output := ActivityInflationControllerOutput{
		InflationBps:		inflation,
		PreviousInflationBps:	input.CurrentInflationBps,
		RawTargetInflationBps:	rawTarget,
		SmoothedInflationBps:	smoothed,
		AppliedDeltaBps:	inflation - input.CurrentInflationBps,
		BoundedByMin:		inflation == params.MinInflationBps && input.CurrentInflationBps+limitedDelta < params.MinInflationBps,
		BoundedByMax:		inflation == params.MaxInflationBps && input.CurrentInflationBps+limitedDelta > params.MaxInflationBps,
		ChangeLimited:		delta != limitedDelta,
		Components:		components,
		QueryableInputs:	input,
	}
	if params.EmergencyFreeze {
		output.InflationBps = input.CurrentInflationBps
		output.SmoothedInflationBps = input.CurrentInflationBps
		output.AppliedDeltaBps = 0
		output.BoundedByMin = false
		output.BoundedByMax = false
		output.ChangeLimited = false
		output.EmergencyFrozen = true
	}
	return output, nil
}

func ReportNetIssuance(input NetIssuanceInput) (NetIssuanceReport, error) {
	if input.EpochID == 0 {
		return NetIssuanceReport{}, fmt.Errorf("epoch_id must be positive")
	}
	if input.AccountingPeriod == "" {
		return NetIssuanceReport{}, fmt.Errorf("accounting_period must not be empty")
	}
	if input.Blocks == 0 {
		return NetIssuanceReport{}, fmt.Errorf("blocks must be positive")
	}
	grossMinted := normalizeInt(input.GrossMintedNaet)
	burned := normalizeInt(input.BurnedNaet)
	feeRevenue := normalizeInt(input.FeeRevenueNaet)
	securitySpend := normalizeInt(input.ValidatorSecuritySpendNaet)
	for _, item := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "gross_minted_naet", value: grossMinted},
		{name: "burned_naet", value: burned},
		{name: "fee_revenue_naet", value: feeRevenue},
		{name: "validator_security_spend_naet", value: securitySpend},
	} {
		if item.value.IsNegative() {
			return NetIssuanceReport{}, fmt.Errorf("%s must not be negative", item.name)
		}
	}
	return NetIssuanceReport{
		EpochID:			input.EpochID,
		AccountingPeriod:		input.AccountingPeriod,
		Blocks:				input.Blocks,
		GrossMintedNaet:		grossMinted,
		BurnedNaet:			burned,
		NetSupplyChangeNaet:		grossMinted.Sub(burned),
		FeeRevenueNaet:			feeRevenue,
		ValidatorSecuritySpendNaet:	securitySpend,
		SecuritySpendPerBlockNaet:	securitySpend.QuoRaw(int64(input.Blocks)),
	}, nil
}

func SimulateActivityInflation(params ActivityInflationControllerParams, steps []InflationSimulationStep) (InflationSimulationReport, error) {
	if len(steps) == 0 {
		return InflationSimulationReport{}, fmt.Errorf("simulation steps must not be empty")
	}
	if err := params.Validate(); err != nil {
		return InflationSimulationReport{}, err
	}

	seenScenarios := make(map[string]bool, len(steps))
	outputs := make([]ActivityInflationControllerOutput, 0, len(steps))
	reports := make([]NetIssuanceReport, 0, len(steps))
	risks := make([]string, 0)
	minObserved := params.MaxInflationBps
	maxObserved := params.MinInflationBps
	current := steps[0].Controller.CurrentInflationBps

	for i, step := range steps {
		if step.Scenario == "" {
			return InflationSimulationReport{}, fmt.Errorf("scenario must not be empty at step %d", i)
		}
		seenScenarios[step.Scenario] = true
		input := step.Controller
		if i > 0 && input.CurrentInflationBps == 0 {
			input.CurrentInflationBps = current
		}
		output, err := ActivityInflationControllerWithParams(input, params)
		if err != nil {
			return InflationSimulationReport{}, err
		}
		outputs = append(outputs, output)
		current = output.InflationBps
		minObserved = minInt64(minObserved, output.InflationBps)
		maxObserved = maxInt64(maxObserved, output.InflationBps)
		if output.InflationBps < params.MinInflationBps || output.InflationBps > params.MaxInflationBps {
			risks = append(risks, "inflation_outside_bounds")
		}
		if absInt64(output.AppliedDeltaBps) > params.PerWindowChangeLimitBps {
			risks = append(risks, "inflation_change_limit_exceeded")
		}
		if step.NetIssuance.Blocks > 0 {
			report, err := ReportNetIssuance(step.NetIssuance)
			if err != nil {
				return InflationSimulationReport{}, err
			}
			reports = append(reports, report)
		}
	}
	for _, scenario := range []string{
		InflationScenarioLowActivity,
		InflationScenarioNormalActivity,
		InflationScenarioHighActivity,
		InflationScenarioAdversarialActivity,
	} {
		if !seenScenarios[scenario] {
			risks = append(risks, "missing_"+scenario)
		}
	}
	return InflationSimulationReport{
		ScenarioCount:			len(seenScenarios),
		FinalInflationBps:		current,
		MinObservedInflationBps:	minObserved,
		MaxObservedInflationBps:	maxObserved,
		ControllerOutputs:		outputs,
		NetIssuanceReports:		reports,
		Passed:				len(risks) == 0,
		Risks:				risks,
	}, nil
}

func (p ActivityInflationControllerParams) Validate() error {
	if err := validateBps("min_inflation_bps", p.MinInflationBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_inflation_bps", p.MaxInflationBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MinInflationBps > p.MaxInflationBps {
		return fmt.Errorf("min inflation must be <= max inflation")
	}
	if err := validateBps("target_stake_bps", p.TargetStakeBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("target_operating_cost_coverage_bps", p.TargetOperatingCostCoverageBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("target_treasury_reserve_health_bps", p.TargetTreasuryReserveHealthBps, 0, BasisPoints); err != nil {
		return err
	}
	if normalizeInt(p.TargetFeeRevenueNaet).IsNegative() {
		return fmt.Errorf("target_fee_revenue_naet must not be negative")
	}
	if p.TargetActiveValidators == 0 {
		return fmt.Errorf("target_active_validators must be positive")
	}
	if p.SmoothingWindow == 0 {
		return fmt.Errorf("smoothing_window must be positive")
	}
	if p.PerWindowChangeLimitBps < 0 {
		return fmt.Errorf("per_window_change_limit_bps must not be negative")
	}
	totalWeight := p.StakeWeightBps + p.OperatingCostWeightBps + p.FeeRevenueWeightBps +
		p.ValidatorCountWeightBps + p.SlashingRiskWeightBps + p.NetworkActivityWeightBps +
		p.TreasuryReserveWeightBps
	if totalWeight != BasisPoints {
		return fmt.Errorf("inflation controller weights must sum to %d", BasisPoints)
	}
	for name, value := range map[string]int64{
		"stake_weight_bps":		p.StakeWeightBps,
		"operating_cost_weight_bps":	p.OperatingCostWeightBps,
		"fee_revenue_weight_bps":	p.FeeRevenueWeightBps,
		"validator_count_weight_bps":	p.ValidatorCountWeightBps,
		"slashing_risk_weight_bps":	p.SlashingRiskWeightBps,
		"network_activity_weight_bps":	p.NetworkActivityWeightBps,
		"treasury_reserve_weight_bps":	p.TreasuryReserveWeightBps,
	} {
		if err := validateBps(name, value, 0, BasisPoints); err != nil {
			return err
		}
	}
	return nil
}

func (input ActivityInflationControllerInput) Validate(params ActivityInflationControllerParams) error {
	if err := validateBps("current_inflation_bps", input.CurrentInflationBps, params.MinInflationBps, params.MaxInflationBps); err != nil {
		return err
	}
	if err := validateBps("bonded_stake_ratio_bps", input.BondedStakeRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("validator_operating_cost_index_bps", input.ValidatorOperatingCostIndexBps, 0, BasisPoints*2); err != nil {
		return err
	}
	if normalizeInt(input.FeeRevenueNaet).IsNegative() {
		return fmt.Errorf("fee_revenue_naet must not be negative")
	}
	if err := validateBps("network_activity_score_bps", input.NetworkActivityScoreBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("treasury_reserve_health_bps", input.TreasuryReserveHealthBps, 0, BasisPoints*2); err != nil {
		return err
	}
	for i, value := range input.RecentInflationBps {
		if err := validateBps(fmt.Sprintf("recent_inflation_bps[%d]", i), value, params.MinInflationBps, params.MaxInflationBps); err != nil {
			return err
		}
	}
	return nil
}

func inflationAdjustmentComponents(input ActivityInflationControllerInput, params ActivityInflationControllerParams) []InflationAdjustmentComponent {
	targetFeeRevenue := normalizeInt(params.TargetFeeRevenueNaet)
	feeRevenue := normalizeInt(input.FeeRevenueNaet)
	feePressureBps := int64(0)
	if targetFeeRevenue.IsPositive() {
		feePressureBps = targetFeeRevenue.Sub(feeRevenue).MulRaw(BasisPoints).Quo(targetFeeRevenue).Int64()
	}
	validatorPressureBps := (int64(params.TargetActiveValidators) - int64(input.ActiveValidatorCount)) * BasisPoints / int64(params.TargetActiveValidators)
	slashingPressureBps := int64(input.SlashingRiskEvents) * 500
	return buildInflationComponents([]InflationAdjustmentComponent{
		{Name: "bonded_stake_ratio", ValueBps: params.TargetStakeBps - input.BondedStakeRatioBps, WeightBps: params.StakeWeightBps},
		{Name: "validator_operating_cost_index", ValueBps: params.TargetOperatingCostCoverageBps - input.ValidatorOperatingCostIndexBps, WeightBps: params.OperatingCostWeightBps},
		{Name: "fee_revenue", ValueBps: feePressureBps, WeightBps: params.FeeRevenueWeightBps},
		{Name: "active_validator_count", ValueBps: validatorPressureBps, WeightBps: params.ValidatorCountWeightBps},
		{Name: "slashing_risk_events", ValueBps: slashingPressureBps, WeightBps: params.SlashingRiskWeightBps},
		{Name: "network_activity_score", ValueBps: DefaultTargetLoadBps - input.NetworkActivityScoreBps, WeightBps: params.NetworkActivityWeightBps},
		{Name: "treasury_reserve_health", ValueBps: params.TargetTreasuryReserveHealthBps - input.TreasuryReserveHealthBps, WeightBps: params.TreasuryReserveWeightBps},
	})
}

func buildInflationComponents(components []InflationAdjustmentComponent) []InflationAdjustmentComponent {
	for i := range components {
		components[i].ValueBps = clampInt64(components[i].ValueBps, -BasisPoints, BasisPoints)
		components[i].ContributionBps = components[i].ValueBps * components[i].WeightBps / BasisPoints
	}
	return components
}

func smoothInflationTarget(rawTarget int64, recent []int64, window uint32) int64 {
	if window <= 1 {
		return rawTarget
	}
	values := make([]int64, 0, int(window))
	if keep := int(window) - 1; keep > 0 && len(recent) > 0 {
		start := len(recent) - keep
		if start < 0 {
			start = 0
		}
		values = append(values, recent[start:]...)
	}
	values = append(values, rawTarget)
	return averageBps(values)
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
