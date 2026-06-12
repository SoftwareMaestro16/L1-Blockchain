package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	AttackClassStakeConcentration	= "stake_concentration"
	AttackClassCommissionBaitSwitch	= "commission_bait_and_switch"
	AttackClassRewardManipulation	= "reward_manipulation"
	AttackClassFeeSpam		= "fee_spam"
	AttackClassStateBloat		= "state_bloat"
	AttackClassEvidenceSpam		= "evidence_spam"
	AttackClassDelegationCapture	= "delegation_capture"

	DefaultAttackMinCostToProfitBps		= int64(15_000)
	DefaultRewardManipulationThresholdBps	= int64(500)
	DefaultDelegationInflowThresholdBps	= int64(1_000)
	DefaultStakeMovementThresholdBps	= int64(750)
	DefaultCartelVotingPowerThresholdBps	= int64(3_334)
	DefaultCartelPenaltyBps			= int64(20_000)
)

type EconomicAttackHardeningParams struct {
	MinCostToProfitBps		int64
	StakeConcentrationThresholdBps	int64
	CommissionJumpThresholdBps	int64
	RewardManipulationThresholdBps	int64
	FeeSpamFailedTxThresholdBps	int64
	StateGrowthThresholdBytes	int64
	EvidenceSpamSubmissionThreshold	uint64
	DelegationInflowThresholdBps	int64
	StakeMovementThresholdBps	int64
	CartelVotingPowerThresholdBps	int64
	CartelPenaltyBps		int64
	CircuitBreakerParams		EconomicCircuitBreakerParams
}

type EconomicAttackPreventionInput struct {
	ExpectedAttackProfitNaet	sdkmath.Int
	TotalStakeNaet			sdkmath.Int
	ValidatorStakeNaet		sdkmath.Int
	ValidatorStakeBps		int64
	CommissionChangeBps		int64
	RewardDeviationBps		int64
	SlashingPenaltyNaet		sdkmath.Int
	FeeSpamTxCount			uint64
	FeePerSpamTxNaet		sdkmath.Int
	FailedTxRateBps			int64
	StateGrowthBytes		int64
	StateExpansionFeeNaet		sdkmath.Int
	EvidenceSubmissions		uint64
	EvidenceDepositNaet		sdkmath.Int
	ReporterRewardCapNaet		sdkmath.Int
	DelegationInflowBps		int64
	ControllerInput			EconomicCircuitBreakerInput
}

type EconomicAttackAssessment struct {
	Class		string
	Triggered	bool
	CostNaet	sdkmath.Int
	ExpectedProfit	sdkmath.Int
	CostToProfitBps	int64
	Profitable	bool
	Mitigation	string
	Signals		[]string
}

type EconomicAttackPreventionReport struct {
	Assessments	[]EconomicAttackAssessment
	CircuitBreaker	EconomicCircuitBreakerOutput
	Passed		bool
	Failed		[]string
}

type EconomicAttackInvariantReport struct {
	Passed	bool
	Failed	[]string
}

type CartelSimulationInput struct {
	ValidatorPowerBps	[]int64
	ColludingIndices	[]int
	RewardPoolNaet		sdkmath.Int
	ExpectedMEVNaet		sdkmath.Int
	Epochs			uint64
	Params			EconomicAttackHardeningParams
}

type CartelSimulationReport struct {
	CartelVotingPowerBps	int64
	ExpectedGainNaet	sdkmath.Int
	ExpectedPenaltyNaet	sdkmath.Int
	Profitable		bool
	Mitigation		string
	Triggered		bool
}

type StakeMovementSnapshot struct {
	ValidatorID	string
	StakeBps	int64
}

type StakeMovementAlert struct {
	ValidatorID	string
	PreviousBps	int64
	CurrentBps	int64
	DeltaBps	int64
	Reason		string
}

type StakeMovementMonitorInput struct {
	Previous	[]StakeMovementSnapshot
	Current		[]StakeMovementSnapshot
	Params		EconomicAttackHardeningParams
}

type StakeMovementMonitorReport struct {
	Alerts		[]StakeMovementAlert
	TotalInflowBps	int64
	TotalOutflowBps	int64
	Abnormal	bool
}

func DefaultEconomicAttackHardeningParams() EconomicAttackHardeningParams {
	return EconomicAttackHardeningParams{
		MinCostToProfitBps:			DefaultAttackMinCostToProfitBps,
		StakeConcentrationThresholdBps:		MaxTopValidatorConcentrationBps,
		CommissionJumpThresholdBps:		MaxDailyCommissionChangeBps,
		RewardManipulationThresholdBps:		DefaultRewardManipulationThresholdBps,
		FeeSpamFailedTxThresholdBps:		DefaultCircuitBreakerFailedTxRateBps,
		StateGrowthThresholdBytes:		DefaultStateGrowthSurchargeThresholdBytes,
		EvidenceSpamSubmissionThreshold:	10,
		DelegationInflowThresholdBps:		DefaultDelegationInflowThresholdBps,
		StakeMovementThresholdBps:		DefaultStakeMovementThresholdBps,
		CartelVotingPowerThresholdBps:		DefaultCartelVotingPowerThresholdBps,
		CartelPenaltyBps:			DefaultCartelPenaltyBps,
		CircuitBreakerParams:			DefaultEconomicCircuitBreakerParams(),
	}
}

func EvaluateEconomicAttackPrevention(input EconomicAttackPreventionInput, params EconomicAttackHardeningParams) (EconomicAttackPreventionReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return EconomicAttackPreventionReport{}, err
	}
	if err := input.Validate(); err != nil {
		return EconomicAttackPreventionReport{}, err
	}
	profit := normalizeInt(input.ExpectedAttackProfitNaet)
	assessments := []EconomicAttackAssessment{
		attackAssessment(AttackClassStakeConcentration, input.ValidatorStakeBps > params.StakeConcentrationThresholdBps, input.ValidatorStakeNaet.Add(normalizeInt(input.SlashingPenaltyNaet)), profit, "concentration_soft_cap_reward_dampening", "validator_stake_above_threshold"),
		attackAssessment(AttackClassCommissionBaitSwitch, input.CommissionChangeBps > params.CommissionJumpThresholdBps, ApplyBps(normalizeInt(input.ValidatorStakeNaet), input.CommissionChangeBps).Add(normalizeInt(input.SlashingPenaltyNaet)), profit, "commission_change_delay_and_risk_flag", "commission_jump_above_governed_limit"),
		attackAssessment(AttackClassRewardManipulation, input.RewardDeviationBps > params.RewardManipulationThresholdBps, normalizeInt(input.SlashingPenaltyNaet).Add(normalizeInt(input.ReporterRewardCapNaet)), profit, "reward_invariant_checks_and_reporter_caps", "reward_deviation_abnormal"),
		attackAssessment(AttackClassFeeSpam, input.FailedTxRateBps > params.FeeSpamFailedTxThresholdBps || input.FeeSpamTxCount > 0, normalizeInt(input.FeePerSpamTxNaet).MulRaw(int64(input.FeeSpamTxCount)), profit, "sender_local_surcharge_and_mempool_admission", "failed_transaction_pressure"),
		attackAssessment(AttackClassStateBloat, input.StateGrowthBytes > params.StateGrowthThresholdBytes, stateBloatAttackCost(input, params), profit, "state_growth_surcharge_and_rent", "state_growth_above_threshold"),
		attackAssessment(AttackClassEvidenceSpam, input.EvidenceSubmissions > params.EvidenceSpamSubmissionThreshold, normalizeInt(input.EvidenceDepositNaet).MulRaw(int64(input.EvidenceSubmissions)).Add(normalizeInt(input.ReporterRewardCapNaet)), profit, "evidence_deposit_duplicate_rejection_and_reward_cap", "evidence_submission_spike"),
		attackAssessment(AttackClassDelegationCapture, input.DelegationInflowBps > params.DelegationInflowThresholdBps, ProportionalShare(normalizeInt(input.TotalStakeNaet), sdkmath.NewInt(input.DelegationInflowBps), sdkmath.NewInt(BasisPoints)), profit, "delegation_inflow_alerts_and_capture_risk_flag", "delegation_inflow_abnormal"),
	}
	circuit, err := EvaluateEconomicCircuitBreaker(input.ControllerInput, params.CircuitBreakerParams)
	if err != nil {
		return EconomicAttackPreventionReport{}, err
	}
	report := EconomicAttackPreventionReport{Assessments: assessments, CircuitBreaker: circuit}
	invariants := ValidateEconomicAttackInvariants(report, params)
	report.Passed = invariants.Passed
	report.Failed = invariants.Failed
	return report, nil
}

func ValidateEconomicAttackInvariants(report EconomicAttackPreventionReport, params EconomicAttackHardeningParams) EconomicAttackInvariantReport {
	params = params.withDefaults()
	seen := make(map[string]bool)
	failed := make([]string, 0)
	for _, assessment := range report.Assessments {
		seen[assessment.Class] = true
		if assessment.Class == "" {
			failed = append(failed, "attack_class_missing")
		}
		if normalizeInt(assessment.CostNaet).IsNegative() {
			failed = append(failed, assessment.Class+"_negative_cost")
		}
		if normalizeInt(assessment.ExpectedProfit).IsNegative() {
			failed = append(failed, assessment.Class+"_negative_profit")
		}
		if assessment.Mitigation == "" {
			failed = append(failed, assessment.Class+"_mitigation_missing")
		}
		if assessment.CostToProfitBps < 0 {
			failed = append(failed, assessment.Class+"_cost_to_profit_missing")
		}
		if assessment.Triggered && assessment.Profitable && assessment.CostToProfitBps < params.MinCostToProfitBps {
			failed = append(failed, assessment.Class+"_attack_profitable_below_cost_floor")
		}
	}
	for _, class := range requiredAttackClasses() {
		if !seen[class] {
			failed = append(failed, class+"_assessment_missing")
		}
	}
	if report.CircuitBreaker.Active && report.CircuitBreaker.CooldownBlocks == 0 {
		failed = append(failed, "circuit_breaker_cooldown_missing")
	}
	return EconomicAttackInvariantReport{Passed: len(failed) == 0, Failed: failed}
}

func SimulateValidatorCartel(input CartelSimulationInput) (CartelSimulationReport, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return CartelSimulationReport{}, err
	}
	if len(input.ValidatorPowerBps) == 0 || len(input.ColludingIndices) == 0 {
		return CartelSimulationReport{}, fmt.Errorf("validator powers and colluding indices are required")
	}
	if input.Epochs == 0 {
		return CartelSimulationReport{}, fmt.Errorf("epochs must be positive")
	}
	for i, power := range input.ValidatorPowerBps {
		if err := validateBps(fmt.Sprintf("validator_power_bps[%d]", i), power, 0, BasisPoints); err != nil {
			return CartelSimulationReport{}, err
		}
	}
	cartelPower := int64(0)
	seen := make(map[int]bool)
	for _, idx := range input.ColludingIndices {
		if idx < 0 || idx >= len(input.ValidatorPowerBps) {
			return CartelSimulationReport{}, fmt.Errorf("colluding index out of range")
		}
		if seen[idx] {
			continue
		}
		seen[idx] = true
		cartelPower += input.ValidatorPowerBps[idx]
	}
	rewardPool := normalizeInt(input.RewardPoolNaet)
	mev := normalizeInt(input.ExpectedMEVNaet)
	if rewardPool.IsNegative() || mev.IsNegative() {
		return CartelSimulationReport{}, fmt.Errorf("reward_pool_naet and expected_mev_naet must not be negative")
	}
	gain := ApplyBps(rewardPool, cartelPower).Add(mev).MulRaw(int64(input.Epochs))
	penalty := ApplyBps(gain, params.CartelPenaltyBps)
	triggered := cartelPower >= params.CartelVotingPowerThresholdBps
	return CartelSimulationReport{
		CartelVotingPowerBps:	cartelPower,
		ExpectedGainNaet:	gain,
		ExpectedPenaltyNaet:	penalty,
		Profitable:		triggered && gain.GT(penalty),
		Mitigation:		"cartel_power_alert_slashing_risk_and_reward_dampening",
		Triggered:		triggered,
	}, nil
}

func MonitorStakeMovements(input StakeMovementMonitorInput) (StakeMovementMonitorReport, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return StakeMovementMonitorReport{}, err
	}
	previous, err := stakeSnapshotMap(input.Previous, "previous")
	if err != nil {
		return StakeMovementMonitorReport{}, err
	}
	current, err := stakeSnapshotMap(input.Current, "current")
	if err != nil {
		return StakeMovementMonitorReport{}, err
	}
	ids := make(map[string]bool)
	for id := range previous {
		ids[id] = true
	}
	for id := range current {
		ids[id] = true
	}
	alerts := make([]StakeMovementAlert, 0)
	inflow := int64(0)
	outflow := int64(0)
	for id := range ids {
		prev := previous[id]
		cur := current[id]
		delta := cur - prev
		if delta > 0 {
			inflow += delta
		} else {
			outflow += -delta
		}
		if absInt64(delta) >= params.StakeMovementThresholdBps {
			reason := "stake_inflow_abnormal"
			if delta < 0 {
				reason = "stake_outflow_abnormal"
			}
			alerts = append(alerts, StakeMovementAlert{ValidatorID: id, PreviousBps: prev, CurrentBps: cur, DeltaBps: delta, Reason: reason})
		}
	}
	sort.SliceStable(alerts, func(i, j int) bool {
		if absInt64(alerts[i].DeltaBps) == absInt64(alerts[j].DeltaBps) {
			return alerts[i].ValidatorID < alerts[j].ValidatorID
		}
		return absInt64(alerts[i].DeltaBps) > absInt64(alerts[j].DeltaBps)
	})
	return StakeMovementMonitorReport{Alerts: alerts, TotalInflowBps: inflow, TotalOutflowBps: outflow, Abnormal: len(alerts) > 0}, nil
}

func EvaluateGovernedEconomicCircuitBreaker(input EconomicCircuitBreakerInput, params EconomicAttackHardeningParams) (EconomicCircuitBreakerOutput, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return EconomicCircuitBreakerOutput{}, err
	}
	return EvaluateEconomicCircuitBreaker(input, params.CircuitBreakerParams)
}

func (p EconomicAttackHardeningParams) Validate() error {
	if err := validateBps("min_cost_to_profit_bps", p.MinCostToProfitBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("stake_concentration_threshold_bps", p.StakeConcentrationThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("commission_jump_threshold_bps", p.CommissionJumpThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("reward_manipulation_threshold_bps", p.RewardManipulationThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("fee_spam_failed_tx_threshold_bps", p.FeeSpamFailedTxThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.StateGrowthThresholdBytes < 0 {
		return fmt.Errorf("state_growth_threshold_bytes must not be negative")
	}
	if err := validateBps("delegation_inflow_threshold_bps", p.DelegationInflowThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("stake_movement_threshold_bps", p.StakeMovementThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("cartel_voting_power_threshold_bps", p.CartelVotingPowerThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("cartel_penalty_bps", p.CartelPenaltyBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	return p.CircuitBreakerParams.Validate()
}

func (p EconomicAttackHardeningParams) withDefaults() EconomicAttackHardeningParams {
	defaults := DefaultEconomicAttackHardeningParams()
	if p.MinCostToProfitBps == 0 {
		p.MinCostToProfitBps = defaults.MinCostToProfitBps
	}
	if p.StakeConcentrationThresholdBps == 0 {
		p.StakeConcentrationThresholdBps = defaults.StakeConcentrationThresholdBps
	}
	if p.CommissionJumpThresholdBps == 0 {
		p.CommissionJumpThresholdBps = defaults.CommissionJumpThresholdBps
	}
	if p.RewardManipulationThresholdBps == 0 {
		p.RewardManipulationThresholdBps = defaults.RewardManipulationThresholdBps
	}
	if p.FeeSpamFailedTxThresholdBps == 0 {
		p.FeeSpamFailedTxThresholdBps = defaults.FeeSpamFailedTxThresholdBps
	}
	if p.StateGrowthThresholdBytes == 0 {
		p.StateGrowthThresholdBytes = defaults.StateGrowthThresholdBytes
	}
	if p.EvidenceSpamSubmissionThreshold == 0 {
		p.EvidenceSpamSubmissionThreshold = defaults.EvidenceSpamSubmissionThreshold
	}
	if p.DelegationInflowThresholdBps == 0 {
		p.DelegationInflowThresholdBps = defaults.DelegationInflowThresholdBps
	}
	if p.StakeMovementThresholdBps == 0 {
		p.StakeMovementThresholdBps = defaults.StakeMovementThresholdBps
	}
	if p.CartelVotingPowerThresholdBps == 0 {
		p.CartelVotingPowerThresholdBps = defaults.CartelVotingPowerThresholdBps
	}
	if p.CartelPenaltyBps == 0 {
		p.CartelPenaltyBps = defaults.CartelPenaltyBps
	}
	if p.CircuitBreakerParams == (EconomicCircuitBreakerParams{}) {
		p.CircuitBreakerParams = defaults.CircuitBreakerParams
	}
	return p
}

func (input EconomicAttackPreventionInput) Validate() error {
	for _, item := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "expected_attack_profit_naet", value: input.ExpectedAttackProfitNaet},
		{name: "total_stake_naet", value: input.TotalStakeNaet},
		{name: "validator_stake_naet", value: input.ValidatorStakeNaet},
		{name: "slashing_penalty_naet", value: input.SlashingPenaltyNaet},
		{name: "fee_per_spam_tx_naet", value: input.FeePerSpamTxNaet},
		{name: "state_expansion_fee_naet", value: input.StateExpansionFeeNaet},
		{name: "evidence_deposit_naet", value: input.EvidenceDepositNaet},
		{name: "reporter_reward_cap_naet", value: input.ReporterRewardCapNaet},
	} {
		if normalizeInt(item.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", item.name)
		}
	}
	if err := validateBps("validator_stake_bps", input.ValidatorStakeBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("commission_change_bps", input.CommissionChangeBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("reward_deviation_bps", input.RewardDeviationBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("failed_tx_rate_bps", input.FailedTxRateBps, 0, BasisPoints); err != nil {
		return err
	}
	if input.StateGrowthBytes < 0 {
		return fmt.Errorf("state_growth_bytes must not be negative")
	}
	return validateBps("delegation_inflow_bps", input.DelegationInflowBps, 0, BasisPoints)
}

func attackAssessment(class string, triggered bool, cost, profit sdkmath.Int, mitigation, signal string) EconomicAttackAssessment {
	cost = normalizeInt(cost)
	profit = normalizeInt(profit)
	costToProfit := int64(0)
	if profit.IsPositive() {
		costToProfit = cost.MulRaw(BasisPoints).Quo(profit).Int64()
	} else if cost.IsPositive() {
		costToProfit = DefaultMaxLoadMultiplierBps
	}
	signals := []string{}
	if triggered {
		signals = append(signals, signal)
	}
	return EconomicAttackAssessment{
		Class:			class,
		Triggered:		triggered,
		CostNaet:		cost,
		ExpectedProfit:		profit,
		CostToProfitBps:	costToProfit,
		Profitable:		triggered && profit.IsPositive() && profit.GT(cost),
		Mitigation:		mitigation,
		Signals:		signals,
	}
}

func stateBloatAttackCost(input EconomicAttackPreventionInput, params EconomicAttackHardeningParams) sdkmath.Int {
	base := normalizeInt(input.StateExpansionFeeNaet)
	if input.StateGrowthBytes <= params.StateGrowthThresholdBytes || params.StateGrowthThresholdBytes == 0 {
		return base
	}
	multiple := (input.StateGrowthBytes - params.StateGrowthThresholdBytes + params.StateGrowthThresholdBytes - 1) / params.StateGrowthThresholdBytes
	surchargeBps := clampInt64(multiple*DefaultStateGrowthSurchargeStepBps, 0, DefaultStateGrowthSurchargeMaxBps)
	return base.Add(ApplyBps(base, surchargeBps))
}

func requiredAttackClasses() []string {
	return []string{
		AttackClassStakeConcentration,
		AttackClassCommissionBaitSwitch,
		AttackClassRewardManipulation,
		AttackClassFeeSpam,
		AttackClassStateBloat,
		AttackClassEvidenceSpam,
		AttackClassDelegationCapture,
	}
}

func stakeSnapshotMap(snapshots []StakeMovementSnapshot, name string) (map[string]int64, error) {
	out := make(map[string]int64, len(snapshots))
	for i, snapshot := range snapshots {
		if snapshot.ValidatorID == "" {
			return nil, fmt.Errorf("%s[%d] validator_id is required", name, i)
		}
		if err := validateBps(fmt.Sprintf("%s[%d].stake_bps", name, i), snapshot.StakeBps, 0, BasisPoints); err != nil {
			return nil, err
		}
		out[snapshot.ValidatorID] = snapshot.StakeBps
	}
	return out, nil
}
