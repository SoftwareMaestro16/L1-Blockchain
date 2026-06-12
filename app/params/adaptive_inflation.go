package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	AdaptiveInflationEventController	= "adaptive_inflation_controller"
	AdaptiveInflationEventMint		= "adaptive_inflation_mint_accounted"
	AdaptiveInflationEventGuard		= "adaptive_inflation_deflation_guard"
	AdaptiveInflationEventReconcile		= "adaptive_inflation_epoch_reconciled"

	DefaultAdaptiveActivityTargetBps	= DefaultTargetLoadBps
	DefaultAdaptiveActivityClampDeltaBps	= int64(1_500)
	DefaultAdaptiveReserveHealthTargetBps	= BasisPoints
)

type AdaptiveInflationParams struct {
	MinInflationBps				int64
	MaxInflationBps				int64
	TargetStakeRatioBps			int64
	TargetActivityBps			int64
	TargetFeeRevenueNaet			sdkmath.Int
	TargetValidatorCount			uint64
	TargetReserveHealthBps			int64
	PerWindowAdjustmentLimitBps		int64
	SmoothingWindow				uint32
	ActivityClampDeltaBps			int64
	StakeWeightBps				int64
	FeeRevenueWeightBps			int64
	ValidatorCountWeightBps			int64
	RewardFloorWeightBps			int64
	ActivityWeightBps			int64
	ReserveHealthWeightBps			int64
	EmergencyFreeze				bool
	GovernanceAllowsBelowRewardFloor	bool
}

type AdaptiveInflationInput struct {
	EpochID				uint64
	AccountingPeriod		string
	BlocksInEpoch			uint64
	CurrentSupplyNaet		sdkmath.Int
	EndingSupplyNaet		sdkmath.Int
	CurrentInflationBps		int64
	BondedStakeRatioBps		int64
	TargetStakeRatioBps		int64
	FeeRevenueNaet			sdkmath.Int
	BurnAmountNaet			sdkmath.Int
	ValidatorCount			uint64
	ValidatorRewardFloorNaet	sdkmath.Int
	NetworkActivitySamplesBps	[]int64
	TreasuryReserveHealthBps	int64
	SecurityReserveHealthBps	int64
	RecentInflationBps		[]int64
	Params				AdaptiveInflationParams
}

type AdaptiveInflationControllerState struct {
	PreviousInflationBps			int64
	RawTargetInflationBps			int64
	SmoothedInflationBps			int64
	InflationRateNextEpochBps		int64
	AppliedDeltaBps				int64
	ManipulationResistantActivityBps	int64
	ActivityManipulationClamped		bool
	BoundedByMin				bool
	BoundedByMax				bool
	ChangeLimited				bool
	EmergencyFrozen				bool
	Components				[]InflationAdjustmentComponent
}

type AdaptiveDeflationGuardStatus struct {
	Active				bool
	Reasons				[]string
	SecurityRewardFloorNaet		sdkmath.Int
	SecurityRewardFloorPreserved	bool
	BurnAmountNaet			sdkmath.Int
	NetIssuanceNaet			sdkmath.Int
}

type AdaptiveInflationAccountingEvent struct {
	Type			string
	EpochID			uint64
	InflationBps		int64
	AmountNaet		sdkmath.Int
	NetSupplyChangeNaet	sdkmath.Int
	Reconciled		bool
	Reasons			[]string
}

type AdaptiveInflationEpochReport struct {
	EpochID				uint64
	InflationRateNextEpochBps	int64
	MintAmountNaet			sdkmath.Int
	ExpectedEndingSupplyNaet	sdkmath.Int
	EndingSupplyNaet		sdkmath.Int
	NetIssuance			NetIssuanceReport
	ControllerState			AdaptiveInflationControllerState
	DeflationGuard			AdaptiveDeflationGuardStatus
	Events				[]AdaptiveInflationAccountingEvent
	Reconciled			bool
	Failed				[]string
}

type AdaptiveInflationStressScenario struct {
	Name	string
	Input	AdaptiveInflationInput
}

type AdaptiveInflationStressReport struct {
	Scenarios		[]AdaptiveInflationEpochReport
	MinInflationObservedBps	int64
	MaxInflationObservedBps	int64
	Passed			bool
	Failed			[]string
}

func DefaultAdaptiveInflationParams() AdaptiveInflationParams {
	return AdaptiveInflationParams{
		MinInflationBps:		MinInflationBps,
		MaxInflationBps:		MaxInflationBps,
		TargetStakeRatioBps:		DefaultTargetStakeBps,
		TargetActivityBps:		DefaultAdaptiveActivityTargetBps,
		TargetFeeRevenueNaet:		sdkmath.NewInt(DefaultFeeRevenueTargetNaet),
		TargetValidatorCount:		DefaultActiveValidatorTarget,
		TargetReserveHealthBps:		DefaultAdaptiveReserveHealthTargetBps,
		PerWindowAdjustmentLimitBps:	DefaultInflationPerWindowChangeLimitBps,
		SmoothingWindow:		DefaultInflationSmoothingWindow,
		ActivityClampDeltaBps:		DefaultAdaptiveActivityClampDeltaBps,
		StakeWeightBps:			2_500,
		FeeRevenueWeightBps:		1_750,
		ValidatorCountWeightBps:	1_000,
		RewardFloorWeightBps:		1_500,
		ActivityWeightBps:		1_500,
		ReserveHealthWeightBps:		750,
	}
}

func ComputeAdaptiveInflationEpoch(input AdaptiveInflationInput) (AdaptiveInflationEpochReport, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return AdaptiveInflationEpochReport{}, err
	}
	if err := input.Validate(params); err != nil {
		return AdaptiveInflationEpochReport{}, err
	}
	if input.TargetStakeRatioBps == 0 {
		input.TargetStakeRatioBps = params.TargetStakeRatioBps
	}

	activityScore, activityClamped := manipulationResistantActivityScore(input.NetworkActivitySamplesBps, params)
	components := adaptiveInflationComponents(input, params, activityScore)
	weightedPressure := int64(0)
	for _, component := range components {
		weightedPressure += component.ContributionBps
	}

	rawTarget := clampInt64(input.CurrentInflationBps+weightedPressure/10, params.MinInflationBps, params.MaxInflationBps)
	smoothed := smoothInflationTarget(rawTarget, input.RecentInflationBps, params.SmoothingWindow)
	delta := smoothed - input.CurrentInflationBps
	limitedDelta := clampInt64(delta, -params.PerWindowAdjustmentLimitBps, params.PerWindowAdjustmentLimitBps)
	nextInflation := clampInt64(input.CurrentInflationBps+limitedDelta, params.MinInflationBps, params.MaxInflationBps)

	state := AdaptiveInflationControllerState{
		PreviousInflationBps:			input.CurrentInflationBps,
		RawTargetInflationBps:			rawTarget,
		SmoothedInflationBps:			smoothed,
		InflationRateNextEpochBps:		nextInflation,
		AppliedDeltaBps:			nextInflation - input.CurrentInflationBps,
		ManipulationResistantActivityBps:	activityScore,
		ActivityManipulationClamped:		activityClamped,
		BoundedByMin:				nextInflation == params.MinInflationBps && input.CurrentInflationBps+limitedDelta < params.MinInflationBps,
		BoundedByMax:				nextInflation == params.MaxInflationBps && input.CurrentInflationBps+limitedDelta > params.MaxInflationBps,
		ChangeLimited:				delta != limitedDelta,
		Components:				components,
	}
	if params.EmergencyFreeze {
		state.RawTargetInflationBps = input.CurrentInflationBps
		state.SmoothedInflationBps = input.CurrentInflationBps
		state.InflationRateNextEpochBps = input.CurrentInflationBps
		state.AppliedDeltaBps = 0
		state.BoundedByMin = false
		state.BoundedByMax = false
		state.ChangeLimited = false
		state.EmergencyFrozen = true
		nextInflation = input.CurrentInflationBps
	}

	mint := ApplyBps(normalizeInt(input.CurrentSupplyNaet), nextInflation)
	netIssuance, err := ReportNetIssuance(NetIssuanceInput{
		EpochID:			input.EpochID,
		AccountingPeriod:		input.AccountingPeriod,
		Blocks:				input.BlocksInEpoch,
		GrossMintedNaet:		mint,
		BurnedNaet:			input.BurnAmountNaet,
		FeeRevenueNaet:			input.FeeRevenueNaet,
		ValidatorSecuritySpendNaet:	maxSdkInt(mint, normalizeInt(input.ValidatorRewardFloorNaet)),
	})
	if err != nil {
		return AdaptiveInflationEpochReport{}, err
	}

	guard := adaptiveDeflationGuard(input, mint, netIssuance.NetSupplyChangeNaet, params)
	startingSupply := normalizeInt(input.CurrentSupplyNaet)
	expectedEnding := startingSupply.Add(mint).Sub(normalizeInt(input.BurnAmountNaet))
	endingSupply := normalizeInt(input.EndingSupplyNaet)
	if endingSupply.IsZero() {
		endingSupply = expectedEnding
	}

	failed := make([]string, 0)
	if !endingSupply.Equal(expectedEnding) {
		failed = append(failed, "epoch_accounting_mismatch")
	}
	if !params.GovernanceAllowsBelowRewardFloor && !guard.SecurityRewardFloorPreserved {
		failed = append(failed, "security_reward_floor_unmet")
	}
	if absInt64(state.AppliedDeltaBps) > params.PerWindowAdjustmentLimitBps {
		failed = append(failed, "per_window_adjustment_limit_exceeded")
	}
	if state.InflationRateNextEpochBps < params.MinInflationBps || state.InflationRateNextEpochBps > params.MaxInflationBps {
		failed = append(failed, "inflation_bounds_exceeded")
	}

	reconciled := len(failed) == 0
	events := adaptiveInflationEvents(input.EpochID, state, mint, netIssuance.NetSupplyChangeNaet, guard, reconciled)
	return AdaptiveInflationEpochReport{
		EpochID:			input.EpochID,
		InflationRateNextEpochBps:	state.InflationRateNextEpochBps,
		MintAmountNaet:			mint,
		ExpectedEndingSupplyNaet:	expectedEnding,
		EndingSupplyNaet:		endingSupply,
		NetIssuance:			netIssuance,
		ControllerState:		state,
		DeflationGuard:			guard,
		Events:				events,
		Reconciled:			reconciled,
		Failed:				failed,
	}, nil
}

func RunAdaptiveInflationStressTest(scenarios []AdaptiveInflationStressScenario, params AdaptiveInflationParams) (AdaptiveInflationStressReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return AdaptiveInflationStressReport{}, err
	}
	if len(scenarios) == 0 {
		return AdaptiveInflationStressReport{}, fmt.Errorf("scenarios are required")
	}

	reports := make([]AdaptiveInflationEpochReport, 0, len(scenarios))
	failed := make([]string, 0)
	minObserved := params.MaxInflationBps
	maxObserved := params.MinInflationBps
	for _, scenario := range scenarios {
		if scenario.Name == "" {
			return AdaptiveInflationStressReport{}, fmt.Errorf("scenario name is required")
		}
		input := scenario.Input
		input.Params = params
		report, err := ComputeAdaptiveInflationEpoch(input)
		if err != nil {
			return AdaptiveInflationStressReport{}, err
		}
		reports = append(reports, report)
		minObserved = minInt64(minObserved, report.InflationRateNextEpochBps)
		maxObserved = maxInt64(maxObserved, report.InflationRateNextEpochBps)
		for _, failure := range report.Failed {
			failed = append(failed, scenario.Name+":"+failure)
		}
	}
	return AdaptiveInflationStressReport{
		Scenarios:			reports,
		MinInflationObservedBps:	minObserved,
		MaxInflationObservedBps:	maxObserved,
		Passed:				len(failed) == 0,
		Failed:				failed,
	}, nil
}

func (p AdaptiveInflationParams) Validate() error {
	if err := validateBps("min_inflation_bps", p.MinInflationBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_inflation_bps", p.MaxInflationBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MinInflationBps > p.MaxInflationBps {
		return fmt.Errorf("min_inflation_bps must be <= max_inflation_bps")
	}
	if err := validateBps("target_stake_ratio_bps", p.TargetStakeRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("target_activity_bps", p.TargetActivityBps, 0, BasisPoints); err != nil {
		return err
	}
	if normalizeInt(p.TargetFeeRevenueNaet).IsNegative() {
		return fmt.Errorf("target_fee_revenue_naet must not be negative")
	}
	if p.TargetValidatorCount == 0 {
		return fmt.Errorf("target_validator_count must be positive")
	}
	if err := validateBps("target_reserve_health_bps", p.TargetReserveHealthBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("per_window_adjustment_limit_bps", p.PerWindowAdjustmentLimitBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.SmoothingWindow == 0 {
		return fmt.Errorf("smoothing_window must be positive")
	}
	if err := validateBps("activity_clamp_delta_bps", p.ActivityClampDeltaBps, 0, BasisPoints); err != nil {
		return err
	}
	for name, value := range map[string]int64{
		"stake_weight_bps":		p.StakeWeightBps,
		"fee_revenue_weight_bps":	p.FeeRevenueWeightBps,
		"validator_count_weight_bps":	p.ValidatorCountWeightBps,
		"reward_floor_weight_bps":	p.RewardFloorWeightBps,
		"activity_weight_bps":		p.ActivityWeightBps,
		"reserve_health_weight_bps":	p.ReserveHealthWeightBps,
	} {
		if value < 0 {
			return fmt.Errorf("%s must not be negative", name)
		}
		if value > DefaultMaxLoadMultiplierBps {
			return fmt.Errorf("%s exceeds maximum", name)
		}
	}
	return nil
}

func (p AdaptiveInflationParams) withDefaults() AdaptiveInflationParams {
	defaults := DefaultAdaptiveInflationParams()
	if p.MinInflationBps == 0 {
		p.MinInflationBps = defaults.MinInflationBps
	}
	if p.MaxInflationBps == 0 {
		p.MaxInflationBps = defaults.MaxInflationBps
	}
	if p.TargetStakeRatioBps == 0 {
		p.TargetStakeRatioBps = defaults.TargetStakeRatioBps
	}
	if p.TargetActivityBps == 0 {
		p.TargetActivityBps = defaults.TargetActivityBps
	}
	if p.TargetFeeRevenueNaet.IsNil() {
		p.TargetFeeRevenueNaet = defaults.TargetFeeRevenueNaet
	}
	if p.TargetValidatorCount == 0 {
		p.TargetValidatorCount = defaults.TargetValidatorCount
	}
	if p.TargetReserveHealthBps == 0 {
		p.TargetReserveHealthBps = defaults.TargetReserveHealthBps
	}
	if p.PerWindowAdjustmentLimitBps == 0 {
		p.PerWindowAdjustmentLimitBps = defaults.PerWindowAdjustmentLimitBps
	}
	if p.SmoothingWindow == 0 {
		p.SmoothingWindow = defaults.SmoothingWindow
	}
	if p.ActivityClampDeltaBps == 0 {
		p.ActivityClampDeltaBps = defaults.ActivityClampDeltaBps
	}
	if p.StakeWeightBps == 0 {
		p.StakeWeightBps = defaults.StakeWeightBps
	}
	if p.FeeRevenueWeightBps == 0 {
		p.FeeRevenueWeightBps = defaults.FeeRevenueWeightBps
	}
	if p.ValidatorCountWeightBps == 0 {
		p.ValidatorCountWeightBps = defaults.ValidatorCountWeightBps
	}
	if p.RewardFloorWeightBps == 0 {
		p.RewardFloorWeightBps = defaults.RewardFloorWeightBps
	}
	if p.ActivityWeightBps == 0 {
		p.ActivityWeightBps = defaults.ActivityWeightBps
	}
	if p.ReserveHealthWeightBps == 0 {
		p.ReserveHealthWeightBps = defaults.ReserveHealthWeightBps
	}
	return p
}

func (input AdaptiveInflationInput) Validate(params AdaptiveInflationParams) error {
	if input.EpochID == 0 {
		return fmt.Errorf("epoch_id must be positive")
	}
	if input.AccountingPeriod == "" {
		return fmt.Errorf("accounting_period must not be empty")
	}
	if input.BlocksInEpoch == 0 {
		return fmt.Errorf("blocks_in_epoch must be positive")
	}
	if !normalizeInt(input.CurrentSupplyNaet).IsPositive() {
		return fmt.Errorf("current_supply_naet must be positive")
	}
	for _, field := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "ending_supply_naet", value: input.EndingSupplyNaet},
		{name: "fee_revenue_naet", value: input.FeeRevenueNaet},
		{name: "burn_amount_naet", value: input.BurnAmountNaet},
		{name: "validator_reward_floor_naet", value: input.ValidatorRewardFloorNaet},
	} {
		if normalizeInt(field.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	if err := validateBps("current_inflation_bps", input.CurrentInflationBps, params.MinInflationBps, params.MaxInflationBps); err != nil {
		return err
	}
	if err := validateBps("bonded_stake_ratio_bps", input.BondedStakeRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	targetStake := input.TargetStakeRatioBps
	if targetStake == 0 {
		targetStake = params.TargetStakeRatioBps
	}
	if err := validateBps("target_stake_ratio_bps", targetStake, 0, BasisPoints); err != nil {
		return err
	}
	if input.ValidatorCount == 0 {
		return fmt.Errorf("validator_count must be positive")
	}
	for i, sample := range input.NetworkActivitySamplesBps {
		if err := validateBps(fmt.Sprintf("network_activity_samples_bps[%d]", i), sample, 0, BasisPoints); err != nil {
			return err
		}
	}
	if err := validateBps("treasury_reserve_health_bps", input.TreasuryReserveHealthBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("security_reserve_health_bps", input.SecurityReserveHealthBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	for i, recent := range input.RecentInflationBps {
		if err := validateBps(fmt.Sprintf("recent_inflation_bps[%d]", i), recent, params.MinInflationBps, params.MaxInflationBps); err != nil {
			return err
		}
	}
	return nil
}

func adaptiveInflationComponents(input AdaptiveInflationInput, params AdaptiveInflationParams, activityScoreBps int64) []InflationAdjustmentComponent {
	targetStake := input.TargetStakeRatioBps
	if targetStake == 0 {
		targetStake = params.TargetStakeRatioBps
	}
	feePressure := int64(0)
	if normalizeInt(params.TargetFeeRevenueNaet).IsPositive() {
		feePressure = normalizeInt(params.TargetFeeRevenueNaet).Sub(normalizeInt(input.FeeRevenueNaet)).MulRaw(BasisPoints).Quo(normalizeInt(params.TargetFeeRevenueNaet)).Int64()
	}
	validatorPressure := int64(params.TargetValidatorCount) - int64(input.ValidatorCount)
	validatorPressure = validatorPressure * BasisPoints / int64(params.TargetValidatorCount)

	currentMint := ApplyBps(normalizeInt(input.CurrentSupplyNaet), input.CurrentInflationBps)
	rewardFloorPressure := int64(0)
	if normalizeInt(input.ValidatorRewardFloorNaet).GT(currentMint) && currentMint.IsPositive() {
		rewardFloorPressure = normalizeInt(input.ValidatorRewardFloorNaet).Sub(currentMint).MulRaw(BasisPoints).Quo(currentMint).Int64()
	}
	reserveHealth := minInt64(input.TreasuryReserveHealthBps, input.SecurityReserveHealthBps)
	return buildInflationComponents([]InflationAdjustmentComponent{
		{Name: "bonded_stake_ratio", ValueBps: targetStake - input.BondedStakeRatioBps, WeightBps: params.StakeWeightBps},
		{Name: "fee_revenue", ValueBps: feePressure, WeightBps: params.FeeRevenueWeightBps},
		{Name: "validator_count", ValueBps: validatorPressure, WeightBps: params.ValidatorCountWeightBps},
		{Name: "validator_reward_floor", ValueBps: rewardFloorPressure, WeightBps: params.RewardFloorWeightBps},
		{Name: "network_activity_score", ValueBps: params.TargetActivityBps - activityScoreBps, WeightBps: params.ActivityWeightBps},
		{Name: "reserve_health", ValueBps: params.TargetReserveHealthBps - reserveHealth, WeightBps: params.ReserveHealthWeightBps},
	})
}

func manipulationResistantActivityScore(samples []int64, params AdaptiveInflationParams) (int64, bool) {
	if len(samples) == 0 {
		return params.TargetActivityBps, false
	}
	sorted := append([]int64(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	median := sorted[len(sorted)/2]
	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	}
	lower := clampInt64(median-params.ActivityClampDeltaBps, 0, BasisPoints)
	upper := clampInt64(median+params.ActivityClampDeltaBps, 0, BasisPoints)
	total := int64(0)
	clamped := false
	for _, sample := range samples {
		bounded := clampInt64(sample, lower, upper)
		if bounded != sample {
			clamped = true
		}
		total += bounded
	}
	return clampInt64(total/int64(len(samples)), 0, BasisPoints), clamped
}

func adaptiveDeflationGuard(input AdaptiveInflationInput, mint, netIssuance sdkmath.Int, params AdaptiveInflationParams) AdaptiveDeflationGuardStatus {
	reasons := make([]string, 0)
	floor := normalizeInt(input.ValidatorRewardFloorNaet)
	preserved := mint.GTE(floor)
	if !preserved {
		reasons = append(reasons, "security_reward_floor_pressure")
	}
	if normalizeInt(input.BurnAmountNaet).GT(mint) {
		reasons = append(reasons, "burn_exceeds_mint")
	}
	if input.BondedStakeRatioBps < params.TargetStakeRatioBps-DefaultStakeTargetToleranceBps {
		reasons = append(reasons, "bonded_stake_below_safety_band")
	}
	return AdaptiveDeflationGuardStatus{
		Active:				len(reasons) > 0,
		Reasons:			reasons,
		SecurityRewardFloorNaet:	floor,
		SecurityRewardFloorPreserved:	preserved,
		BurnAmountNaet:			normalizeInt(input.BurnAmountNaet),
		NetIssuanceNaet:		normalizeInt(netIssuance),
	}
}

func adaptiveInflationEvents(epochID uint64, state AdaptiveInflationControllerState, mint, netSupplyChange sdkmath.Int, guard AdaptiveDeflationGuardStatus, reconciled bool) []AdaptiveInflationAccountingEvent {
	events := []AdaptiveInflationAccountingEvent{
		{
			Type:		AdaptiveInflationEventController,
			EpochID:	epochID,
			InflationBps:	state.InflationRateNextEpochBps,
			Reconciled:	true,
		},
		{
			Type:			AdaptiveInflationEventMint,
			EpochID:		epochID,
			InflationBps:		state.InflationRateNextEpochBps,
			AmountNaet:		normalizeInt(mint),
			NetSupplyChangeNaet:	normalizeInt(netSupplyChange),
			Reconciled:		true,
		},
		{
			Type:			AdaptiveInflationEventReconcile,
			EpochID:		epochID,
			InflationBps:		state.InflationRateNextEpochBps,
			AmountNaet:		normalizeInt(mint),
			NetSupplyChangeNaet:	normalizeInt(netSupplyChange),
			Reconciled:		reconciled,
		},
	}
	if guard.Active {
		events = append(events, AdaptiveInflationAccountingEvent{
			Type:			AdaptiveInflationEventGuard,
			EpochID:		epochID,
			InflationBps:		state.InflationRateNextEpochBps,
			AmountNaet:		guard.BurnAmountNaet,
			NetSupplyChangeNaet:	guard.NetIssuanceNaet,
			Reconciled:		reconciled,
			Reasons:		append([]string(nil), guard.Reasons...),
		})
	}
	return events
}

func maxSdkInt(a, b sdkmath.Int) sdkmath.Int {
	if normalizeInt(a).GTE(normalizeInt(b)) {
		return normalizeInt(a)
	}
	return normalizeInt(b)
}
