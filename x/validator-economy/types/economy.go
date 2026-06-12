package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	ModuleName		= "validator_economy"
	DefaultScoreVersion	= uint32(1)

	SaturationStatusNone		= "none"
	SaturationStatusSaturated	= "saturated"

	DefaultRewardAdjustmentLimitBps		= uint32(2_500)
	DefaultBootstrapBandMaxStakeBps		= uint32(500)
	DefaultBootstrapBonusBps		= uint32(500)
	DefaultBootstrapMaxEpochs		= uint64(4)
	DefaultMinimumFullRewardPerformanceBps	= uint32(9_500)
	DefaultOperatingCostMarginBps		= uint32(1_000)
)

type ValidatorScoreRecord struct {
	EpochID			uint64
	ValidatorAddress	string
	RawStake		sdkmath.Int
	EffectiveStake		sdkmath.Int
	StakeWeight		sdkmath.Int
	PerformanceFactor	uint32
	UptimeFactor		uint32
	LatencyFactor		uint32
	ReliabilityIndex	uint32
	ValidatorScore		sdkmath.Int
	SaturationStatus	string
	ScoreVersion		uint32
}

type ValidatorEconomyGovernanceParams struct {
	MinSelfDelegationNaet		sdkmath.Int
	MaxActiveValidators		uint32
	ConcentrationSoftCapBps		uint32
	StakeWeightBps			uint32
	SelfDelegationWeightBps		uint32
	PerformanceWeightBps		uint32
	UptimeWeightBps			uint32
	MissedBlockWeightBps		uint32
	SlashHistoryWeightBps		uint32
	CommissionWeightBps		uint32
	MetadataWeightBps		uint32
	EpochLengthSeconds		uint64
	MaxValidatorSetChangeRateBps	uint32
	RewardAdjustmentLimitBps	uint32
	BootstrapBandMaxStakeBps	uint32
	BootstrapBonusBps		uint32
	BootstrapMaxEpochs		uint64
	MinimumFullRewardPerformanceBps	uint32
}

type ValidatorSelectionInput struct {
	EpochID			uint64
	ValidatorAddress	string
	BondedStakeNaet		sdkmath.Int
	SelfDelegationNaet	sdkmath.Int
	UptimeBps		uint32
	MissedBlockRateBps	uint32
	SlashHistoryCount	uint32
	CommissionBps		uint32
	StakeConcentrationBps	uint32
	MetadataCompletenessBps	uint32
}

type ValidatorEligibilityScore struct {
	EpochID				uint64
	ValidatorAddress		string
	Eligible			bool
	ScoreBps			uint32
	BondedStakeComponentBps		uint32
	SelfDelegationComponentBps	uint32
	UptimeComponentBps		uint32
	MissedBlockComponentBps		uint32
	SlashHistoryComponentBps	uint32
	CommissionComponentBps		uint32
	ConcentrationComponentBps	uint32
	MetadataComponentBps		uint32
	Reasons				[]string
}

type EpochSelectionEvent struct {
	EpochID			uint64
	ValidatorAddress	string
	Selected		bool
	Score			ValidatorEligibilityScore
	ScoreVersion		uint32
}

type ScoreComponentState struct {
	Records []ValidatorScoreRecord
}

type ElectionRanking struct {
	EpochID			uint64
	Records			[]ValidatorScoreRecord
	Rejected		[]RejectedScoreCandidate
	MaxValidatorSetChanges	uint32
	TransitionLimited	bool
	RequestedValidatorCount	uint32
}

type RejectedScoreCandidate struct {
	ValidatorAddress	string
	Reason			string
}

type ScoreSimulationInput struct {
	EpochID		uint64
	Params		postypes.Params
	Candidates	[]postypes.Candidate
	PreviousActive	[]string
	TargetActive	uint32
}

type ScoreSimulationResult struct {
	Ranking			ElectionRanking
	ActiveValidatorIDs	[]string
	TotalRawStakeNaet	sdkmath.Int
	TotalEffectiveNaet	sdkmath.Int
	MaxRawStakeShareBps	uint32
	MaxEffectiveShareBps	uint32
	CentralizationWarning	bool
}

type ChurnScenario string

const (
	ChurnScenarioNormal		ChurnScenario	= "normal"
	ChurnScenarioAdversarial	ChurnScenario	= "adversarial"
	ChurnScenarioLowParticipation	ChurnScenario	= "low_participation"
)

type ValidatorChurnSimulationInput struct {
	Scenario	ChurnScenario
	ScoreInput	ScoreSimulationInput
	ExpectedMaxNew	uint32
}

type ValidatorChurnSimulationReport struct {
	Scenario		ChurnScenario
	Passed			bool
	ActiveValidatorIDs	[]string
	NewValidatorCount	uint32
	TransitionLimited	bool
	Warnings		[]string
}

type ValidatorRewardAdjustmentInput struct {
	Record			ValidatorScoreRecord
	TotalActiveStakeNaet	sdkmath.Int
	OperatingCostNaet	sdkmath.Int
	ExpectedRewardNaet	sdkmath.Int
	ValidatorAgeEpochs	uint64
	BootstrapQualified	bool
	Governance		ValidatorEconomyGovernanceParams
}

type ValidatorRewardAdjustment struct {
	ValidatorAddress		string
	BaseRewardWeightNaet		sdkmath.Int
	AdjustedRewardWeightNaet	sdkmath.Int
	RewardMultiplierBps		uint32
	ReliabilityAdjustmentBps	uint32
	ConcentrationDampeningBps	uint32
	BootstrapBonusBps		uint32
	BootstrapExpired		bool
	FullRewardEligible		bool
	RewardPerVotingPowerNaet	sdkmath.Int
	ProfitabilityMarginBps		int32
	VisibleBeforeDelegation		bool
}

func DefaultValidatorEconomyGovernanceParams(params postypes.Params) ValidatorEconomyGovernanceParams {
	return ValidatorEconomyGovernanceParams{
		MinSelfDelegationNaet:			params.MinStakeNaet,
		MaxActiveValidators:			params.MaxActiveValidators,
		ConcentrationSoftCapBps:		params.MaxVotingPowerBps,
		StakeWeightBps:				1_500,
		SelfDelegationWeightBps:		1_000,
		PerformanceWeightBps:			1_500,
		UptimeWeightBps:			2_000,
		MissedBlockWeightBps:			1_000,
		SlashHistoryWeightBps:			1_000,
		CommissionWeightBps:			1_000,
		MetadataWeightBps:			1_000,
		EpochLengthSeconds:			params.EpochDurationSeconds,
		MaxValidatorSetChangeRateBps:		params.MaxValidatorSetChangeRateBps,
		RewardAdjustmentLimitBps:		DefaultRewardAdjustmentLimitBps,
		BootstrapBandMaxStakeBps:		DefaultBootstrapBandMaxStakeBps,
		BootstrapBonusBps:			DefaultBootstrapBonusBps,
		BootstrapMaxEpochs:			DefaultBootstrapMaxEpochs,
		MinimumFullRewardPerformanceBps:	DefaultMinimumFullRewardPerformanceBps,
	}
}

func (p ValidatorEconomyGovernanceParams) Validate(posParams postypes.Params) error {
	if err := posParams.Validate(); err != nil {
		return err
	}
	if !p.MinSelfDelegationNaet.IsPositive() {
		return errors.New("minimum self delegation must be positive")
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "max_active_validators", value: p.MaxActiveValidators},
		{name: "concentration_soft_cap_bps", value: p.ConcentrationSoftCapBps},
		{name: "stake_weight_bps", value: p.StakeWeightBps},
		{name: "self_delegation_weight_bps", value: p.SelfDelegationWeightBps},
		{name: "performance_weight_bps", value: p.PerformanceWeightBps},
		{name: "uptime_weight_bps", value: p.UptimeWeightBps},
		{name: "missed_block_weight_bps", value: p.MissedBlockWeightBps},
		{name: "slash_history_weight_bps", value: p.SlashHistoryWeightBps},
		{name: "commission_weight_bps", value: p.CommissionWeightBps},
		{name: "metadata_weight_bps", value: p.MetadataWeightBps},
		{name: "max_validator_set_change_rate_bps", value: p.MaxValidatorSetChangeRateBps},
		{name: "reward_adjustment_limit_bps", value: p.RewardAdjustmentLimitBps},
		{name: "bootstrap_band_max_stake_bps", value: p.BootstrapBandMaxStakeBps},
		{name: "bootstrap_bonus_bps", value: p.BootstrapBonusBps},
		{name: "minimum_full_reward_performance_bps", value: p.MinimumFullRewardPerformanceBps},
	} {
		if item.value > postypes.BasisPoints {
			return fmt.Errorf("%s must be <= %d bps", item.name, postypes.BasisPoints)
		}
	}
	if p.MaxActiveValidators == 0 || p.MaxActiveValidators > posParams.MaxActiveValidators {
		return fmt.Errorf("max_active_validators must be within 1..%d", posParams.MaxActiveValidators)
	}
	weightTotal := uint64(p.StakeWeightBps) + uint64(p.SelfDelegationWeightBps) + uint64(p.PerformanceWeightBps) +
		uint64(p.UptimeWeightBps) + uint64(p.MissedBlockWeightBps) + uint64(p.SlashHistoryWeightBps) +
		uint64(p.CommissionWeightBps) + uint64(p.MetadataWeightBps)
	if weightTotal != uint64(postypes.BasisPoints) {
		return fmt.Errorf("validator economy score weights must sum to %d bps", postypes.BasisPoints)
	}
	if p.EpochLengthSeconds < postypes.MinEpochDurationSeconds || p.EpochLengthSeconds > postypes.MaxEpochDurationSeconds {
		return fmt.Errorf("epoch length must be within %d..%d seconds", postypes.MinEpochDurationSeconds, postypes.MaxEpochDurationSeconds)
	}
	return nil
}

func ComputeValidatorEligibilityScore(posParams postypes.Params, gov ValidatorEconomyGovernanceParams, input ValidatorSelectionInput) (ValidatorEligibilityScore, error) {
	if gov.MinSelfDelegationNaet.IsNil() {
		gov = DefaultValidatorEconomyGovernanceParams(posParams)
	}
	if err := gov.Validate(posParams); err != nil {
		return ValidatorEligibilityScore{}, err
	}
	if input.EpochID == 0 {
		return ValidatorEligibilityScore{}, errors.New("selection score epoch id is required")
	}
	if err := validateEconomyToken("selection validator", input.ValidatorAddress); err != nil {
		return ValidatorEligibilityScore{}, err
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "uptime_bps", value: input.UptimeBps},
		{name: "missed_block_rate_bps", value: input.MissedBlockRateBps},
		{name: "commission_bps", value: input.CommissionBps},
		{name: "stake_concentration_bps", value: input.StakeConcentrationBps},
		{name: "metadata_completeness_bps", value: input.MetadataCompletenessBps},
	} {
		if item.value > postypes.BasisPoints {
			return ValidatorEligibilityScore{}, fmt.Errorf("%s must be <= %d bps", item.name, postypes.BasisPoints)
		}
	}
	if input.CommissionBps > posParams.MaxCommissionBps {
		return ValidatorEligibilityScore{}, fmt.Errorf("commission must be <= %d bps", posParams.MaxCommissionBps)
	}
	if input.BondedStakeNaet.IsNegative() || input.SelfDelegationNaet.IsNegative() {
		return ValidatorEligibilityScore{}, errors.New("selection stake amounts cannot be negative")
	}

	reasons := make([]string, 0, 4)
	if input.BondedStakeNaet.LT(posParams.MinStakeNaet) {
		reasons = append(reasons, "bonded_stake_below_minimum")
	}
	if input.SelfDelegationNaet.LT(gov.MinSelfDelegationNaet) {
		reasons = append(reasons, "self_delegation_below_minimum")
	}
	if input.UptimeBps < posParams.MinUptimeBps {
		reasons = append(reasons, "uptime_below_minimum")
	}
	if input.MetadataCompletenessBps < postypes.BasisPoints {
		reasons = append(reasons, "metadata_incomplete")
	}

	stakeComponent := uint32(0)
	if posParams.StakeSaturationNaet.IsPositive() {
		stakeComponent = uint32(minInt(input.BondedStakeNaet.MulRaw(int64(postypes.BasisPoints)).Quo(posParams.StakeSaturationNaet), int64(postypes.BasisPoints)))
	}
	selfDelegationComponent := uint32(0)
	if gov.MinSelfDelegationNaet.IsPositive() {
		selfDelegationComponent = uint32(minInt(input.SelfDelegationNaet.MulRaw(int64(postypes.BasisPoints)).Quo(gov.MinSelfDelegationNaet), int64(postypes.BasisPoints)))
	}
	missedBlockComponent := postypes.BasisPoints - input.MissedBlockRateBps
	slashComponent := slashHistoryComponentBps(input.SlashHistoryCount)
	commissionComponent := postypes.BasisPoints - uint32(uint64(input.CommissionBps)*uint64(postypes.BasisPoints)/uint64(posParams.MaxCommissionBps))
	concentrationComponent := postypes.BasisPoints
	if input.StakeConcentrationBps > gov.ConcentrationSoftCapBps {
		over := input.StakeConcentrationBps - gov.ConcentrationSoftCapBps
		denom := postypes.BasisPoints - gov.ConcentrationSoftCapBps
		if denom == 0 || over >= denom {
			concentrationComponent = 0
		} else {
			concentrationComponent = postypes.BasisPoints - uint32(uint64(over)*uint64(postypes.BasisPoints)/uint64(denom))
		}
	}
	performanceComponent := input.UptimeBps
	score := weightedAverageBps([]weightedScoreComponent{
		{value: stakeComponent, weight: gov.StakeWeightBps},
		{value: selfDelegationComponent, weight: gov.SelfDelegationWeightBps},
		{value: performanceComponent, weight: gov.PerformanceWeightBps},
		{value: input.UptimeBps, weight: gov.UptimeWeightBps},
		{value: missedBlockComponent, weight: gov.MissedBlockWeightBps},
		{value: slashComponent, weight: gov.SlashHistoryWeightBps},
		{value: commissionComponent, weight: gov.CommissionWeightBps},
		{value: input.MetadataCompletenessBps, weight: gov.MetadataWeightBps},
	})
	score = uint32(uint64(score) * uint64(concentrationComponent) / uint64(postypes.BasisPoints))

	return ValidatorEligibilityScore{
		EpochID:			input.EpochID,
		ValidatorAddress:		strings.TrimSpace(input.ValidatorAddress),
		Eligible:			len(reasons) == 0,
		ScoreBps:			score,
		BondedStakeComponentBps:	stakeComponent,
		SelfDelegationComponentBps:	selfDelegationComponent,
		UptimeComponentBps:		input.UptimeBps,
		MissedBlockComponentBps:	missedBlockComponent,
		SlashHistoryComponentBps:	slashComponent,
		CommissionComponentBps:		commissionComponent,
		ConcentrationComponentBps:	concentrationComponent,
		MetadataComponentBps:		input.MetadataCompletenessBps,
		Reasons:			reasons,
	}, nil
}

type weightedScoreComponent struct {
	value	uint32
	weight	uint32
}

func BuildValidatorScoreRecord(epochID uint64, params postypes.Params, candidate postypes.Candidate) (ValidatorScoreRecord, error) {
	scored, err := postypes.ScoreCandidate(params, candidate)
	if err != nil {
		return ValidatorScoreRecord{}, err
	}
	status := SaturationStatusNone
	if scored.ScoreComponents.SaturatedStakeNaet.IsPositive() {
		status = SaturationStatusSaturated
	}
	record := ValidatorScoreRecord{
		EpochID:		epochID,
		ValidatorAddress:	strings.TrimSpace(scored.ValidatorID),
		RawStake:		scored.TotalStakeNaet,
		EffectiveStake:		scored.EffectiveStakeNaet,
		StakeWeight:		scored.ScoreComponents.StakeWeightNaet,
		PerformanceFactor:	scored.ScoreComponents.PerformanceFactorBps,
		UptimeFactor:		scored.ScoreComponents.UptimeFactorBps,
		LatencyFactor:		scored.ScoreComponents.LatencyFactorBps,
		ReliabilityIndex:	scored.ScoreComponents.ReliabilityIndexBps,
		ValidatorScore:		scored.Score,
		SaturationStatus:	status,
		ScoreVersion:		DefaultScoreVersion,
	}
	return record, record.Validate()
}

func BuildEpochSelectionEvents(epochID uint64, posParams postypes.Params, gov ValidatorEconomyGovernanceParams, ranking ElectionRanking, candidates []postypes.Candidate) ([]EpochSelectionEvent, error) {
	if ranking.EpochID != epochID {
		return nil, errors.New("selection event epoch must match ranking epoch")
	}
	selected := make(map[string]struct{}, len(ranking.Records))
	for _, record := range ranking.Records {
		selected[record.ValidatorAddress] = struct{}{}
	}
	totalStake := sdkmath.ZeroInt()
	for _, candidate := range candidates {
		totalStake = totalStake.Add(candidate.SelfStakeNaet).Add(candidate.DelegatedStakeNaet)
	}
	events := make([]EpochSelectionEvent, 0, len(candidates))
	for _, candidate := range candidates {
		validatorID := strings.TrimSpace(candidate.ValidatorID)
		bonded := candidate.SelfStakeNaet.Add(candidate.DelegatedStakeNaet)
		score, err := ComputeValidatorEligibilityScore(posParams, gov, ValidatorSelectionInput{
			EpochID:			epochID,
			ValidatorAddress:		validatorID,
			BondedStakeNaet:		bonded,
			SelfDelegationNaet:		candidate.SelfStakeNaet,
			UptimeBps:			candidate.UptimeFactorBps,
			MissedBlockRateBps:		postypes.BasisPoints - normalizeOptionalBps(candidate.UptimeFactorBps),
			SlashHistoryCount:		0,
			CommissionBps:			candidate.CommissionBps,
			StakeConcentrationBps:		shareBps(bonded, totalStake),
			MetadataCompletenessBps:	metadataCompletenessBps(candidate),
		})
		if err != nil {
			return nil, err
		}
		_, isSelected := selected[validatorID]
		events = append(events, EpochSelectionEvent{
			EpochID:		epochID,
			ValidatorAddress:	validatorID,
			Selected:		isSelected,
			Score:			score,
			ScoreVersion:		DefaultScoreVersion,
		})
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].EpochID != events[j].EpochID {
			return events[i].EpochID < events[j].EpochID
		}
		return events[i].ValidatorAddress < events[j].ValidatorAddress
	})
	return events, nil
}

func SimulateValidatorChurn(input ValidatorChurnSimulationInput) (ValidatorChurnSimulationReport, error) {
	if err := input.Scenario.Validate(); err != nil {
		return ValidatorChurnSimulationReport{}, err
	}
	result, err := SimulateScores(input.ScoreInput)
	if err != nil {
		return ValidatorChurnSimulationReport{}, err
	}
	previous := make(map[string]struct{}, len(input.ScoreInput.PreviousActive))
	for _, validatorID := range input.ScoreInput.PreviousActive {
		validatorID = strings.TrimSpace(validatorID)
		if validatorID != "" {
			previous[validatorID] = struct{}{}
		}
	}
	newCount := uint32(0)
	for _, validatorID := range result.ActiveValidatorIDs {
		if _, found := previous[validatorID]; !found {
			newCount++
		}
	}
	expectedMax := input.ExpectedMaxNew
	if expectedMax == 0 && len(previous) > 0 {
		maxChanges, err := MaxValidatorSetChanges(input.ScoreInput.Params, uint32(len(previous)))
		if err != nil {
			return ValidatorChurnSimulationReport{}, err
		}
		expectedMax = maxChanges
	}
	warnings := make([]string, 0, 3)
	if len(result.ActiveValidatorIDs) < int(input.ScoreInput.Params.MinActiveValidators) {
		warnings = append(warnings, "low_participation_active_set_below_minimum")
	}
	if expectedMax > 0 && newCount > expectedMax {
		warnings = append(warnings, "validator_churn_exceeds_epoch_bound")
	}
	if input.Scenario == ChurnScenarioAdversarial && result.CentralizationWarning {
		warnings = append(warnings, "adversarial_concentration_warning")
	}
	return ValidatorChurnSimulationReport{
		Scenario:		input.Scenario,
		Passed:			len(warnings) == 0,
		ActiveValidatorIDs:	result.ActiveValidatorIDs,
		NewValidatorCount:	newCount,
		TransitionLimited:	result.Ranking.TransitionLimited,
		Warnings:		warnings,
	}, nil
}

func (s ChurnScenario) Validate() error {
	switch s {
	case ChurnScenarioNormal, ChurnScenarioAdversarial, ChurnScenarioLowParticipation:
		return nil
	default:
		return fmt.Errorf("unsupported churn scenario %q", s)
	}
}

func ComputeValidatorRewardAdjustment(input ValidatorRewardAdjustmentInput) (ValidatorRewardAdjustment, error) {
	if err := input.Record.Validate(); err != nil {
		return ValidatorRewardAdjustment{}, err
	}
	posParams := postypes.DefaultParams()
	gov := input.Governance
	if gov.MinSelfDelegationNaet.IsNil() {
		gov = DefaultValidatorEconomyGovernanceParams(posParams)
	}
	if err := gov.Validate(posParams); err != nil {
		return ValidatorRewardAdjustment{}, err
	}
	totalActiveStake := normalizeEconomyInt(input.TotalActiveStakeNaet)
	operatingCost := normalizeEconomyInt(input.OperatingCostNaet)
	expectedReward := normalizeEconomyInt(input.ExpectedRewardNaet)
	if totalActiveStake.IsNegative() || operatingCost.IsNegative() || expectedReward.IsNegative() {
		return ValidatorRewardAdjustment{}, errors.New("reward adjustment amounts cannot be negative")
	}
	reliabilityAdjustment := uint32(0)
	fullRewardEligible := input.Record.PerformanceFactor >= gov.MinimumFullRewardPerformanceBps &&
		input.Record.UptimeFactor >= gov.MinimumFullRewardPerformanceBps &&
		input.Record.ReliabilityIndex >= gov.MinimumFullRewardPerformanceBps
	if !fullRewardEligible {
		worst := minUint32(input.Record.PerformanceFactor, minUint32(input.Record.UptimeFactor, input.Record.ReliabilityIndex))
		gap := gov.MinimumFullRewardPerformanceBps - worst
		reliabilityAdjustment = minUint32(gap, gov.RewardAdjustmentLimitBps)
	}
	concentrationDampening := uint32(0)
	share := shareBps(input.Record.RawStake, totalActiveStake)
	if share > gov.ConcentrationSoftCapBps {
		over := share - gov.ConcentrationSoftCapBps
		denom := postypes.BasisPoints - gov.ConcentrationSoftCapBps
		if denom > 0 {
			concentrationDampening = minUint32(uint32(uint64(over)*uint64(gov.RewardAdjustmentLimitBps)/uint64(denom)), gov.RewardAdjustmentLimitBps)
		}
	}
	bootstrapExpired := input.ValidatorAgeEpochs >= gov.BootstrapMaxEpochs
	bootstrapBonus := uint32(0)
	if input.BootstrapQualified && !bootstrapExpired && share <= gov.BootstrapBandMaxStakeBps {
		bootstrapBonus = minUint32(gov.BootstrapBonusBps, gov.RewardAdjustmentLimitBps)
	}
	totalDampening := minUint32(reliabilityAdjustment+concentrationDampening, gov.RewardAdjustmentLimitBps)
	multiplier := postypes.BasisPoints - totalDampening + bootstrapBonus
	adjusted := input.Record.StakeWeight.MulRaw(int64(multiplier)).QuoRaw(int64(postypes.BasisPoints))
	rewardPerPower := sdkmath.ZeroInt()
	if input.Record.EffectiveStake.IsPositive() {
		rewardPerPower = expectedReward.MulRaw(1_000_000_000).Quo(input.Record.EffectiveStake)
	}
	margin := int32(0)
	if operatingCost.IsPositive() {
		net := expectedReward.Sub(operatingCost)
		margin = int32(net.MulRaw(int64(postypes.BasisPoints)).Quo(operatingCost).Int64())
	}
	return ValidatorRewardAdjustment{
		ValidatorAddress:		input.Record.ValidatorAddress,
		BaseRewardWeightNaet:		input.Record.StakeWeight,
		AdjustedRewardWeightNaet:	adjusted,
		RewardMultiplierBps:		multiplier,
		ReliabilityAdjustmentBps:	reliabilityAdjustment,
		ConcentrationDampeningBps:	concentrationDampening,
		BootstrapBonusBps:		bootstrapBonus,
		BootstrapExpired:		bootstrapExpired,
		FullRewardEligible:		fullRewardEligible,
		RewardPerVotingPowerNaet:	rewardPerPower,
		ProfitabilityMarginBps:		margin,
		VisibleBeforeDelegation:	true,
	}, nil
}

func (r ValidatorScoreRecord) Validate() error {
	if r.EpochID == 0 {
		return errors.New("score record epoch id is required")
	}
	if strings.TrimSpace(r.ValidatorAddress) == "" {
		return errors.New("score record validator address is required")
	}
	if r.RawStake.IsNegative() {
		return errors.New("score record raw stake cannot be negative")
	}
	if r.EffectiveStake.IsNegative() {
		return errors.New("score record effective stake cannot be negative")
	}
	if r.StakeWeight.IsNegative() {
		return errors.New("score record stake weight cannot be negative")
	}
	if r.EffectiveStake.GT(r.RawStake) {
		return errors.New("score record effective stake cannot exceed raw stake")
	}
	if !r.EffectiveStake.Equal(r.StakeWeight) {
		return errors.New("score record stake weight must equal effective stake")
	}
	if r.PerformanceFactor > postypes.BasisPoints {
		return fmt.Errorf("performance factor must be <= %d bps", postypes.BasisPoints)
	}
	if r.UptimeFactor > postypes.BasisPoints {
		return fmt.Errorf("uptime factor must be <= %d bps", postypes.BasisPoints)
	}
	if r.LatencyFactor > postypes.BasisPoints {
		return fmt.Errorf("latency factor must be <= %d bps", postypes.BasisPoints)
	}
	if r.ReliabilityIndex > postypes.BasisPoints {
		return fmt.Errorf("reliability index must be <= %d bps", postypes.BasisPoints)
	}
	if r.ValidatorScore.IsNegative() {
		return errors.New("validator score cannot be negative")
	}
	if r.SaturationStatus != SaturationStatusNone && r.SaturationStatus != SaturationStatusSaturated {
		return fmt.Errorf("unsupported saturation status %q", r.SaturationStatus)
	}
	if r.ScoreVersion == 0 {
		return errors.New("score version is required")
	}
	return nil
}

func NewScoreComponentState(records []ValidatorScoreRecord) (ScoreComponentState, error) {
	out := make([]ValidatorScoreRecord, len(records))
	seen := make(map[string]struct{}, len(records))
	for i, record := range records {
		record.ValidatorAddress = strings.TrimSpace(record.ValidatorAddress)
		if err := record.Validate(); err != nil {
			return ScoreComponentState{}, err
		}
		key := scoreRecordKey(record.EpochID, record.ValidatorAddress)
		if _, found := seen[key]; found {
			return ScoreComponentState{}, fmt.Errorf("duplicate score record %s", key)
		}
		seen[key] = struct{}{}
		out[i] = record
	}
	sortScoreRecords(out)
	return ScoreComponentState{Records: out}, nil
}

func (s ScoreComponentState) GetScoreRecord(epochID uint64, validatorAddress string) (ValidatorScoreRecord, bool) {
	validatorAddress = strings.TrimSpace(validatorAddress)
	for _, record := range s.Records {
		if record.EpochID == epochID && record.ValidatorAddress == validatorAddress {
			return record, true
		}
	}
	return ValidatorScoreRecord{}, false
}

func (s ScoreComponentState) RecordsForEpoch(epochID uint64) []ValidatorScoreRecord {
	records := make([]ValidatorScoreRecord, 0)
	for _, record := range s.Records {
		if record.EpochID == epochID {
			records = append(records, record)
		}
	}
	sortScoreRecords(records)
	return records
}

func BuildElectionRanking(epochID uint64, params postypes.Params, candidates []postypes.Candidate, targetActive uint32) (ElectionRanking, error) {
	if err := params.Validate(); err != nil {
		return ElectionRanking{}, err
	}
	if epochID == 0 {
		return ElectionRanking{}, errors.New("ranking epoch id is required")
	}
	if targetActive == 0 {
		targetActive = params.MinActiveValidators
	}
	ranking := ElectionRanking{EpochID: epochID, RequestedValidatorCount: targetActive}
	records := make([]ValidatorScoreRecord, 0, len(candidates))
	for _, candidate := range candidates {
		record, err := BuildValidatorScoreRecord(epochID, params, candidate)
		if err != nil {
			ranking.Rejected = append(ranking.Rejected, RejectedScoreCandidate{
				ValidatorAddress:	strings.TrimSpace(candidate.ValidatorID),
				Reason:			err.Error(),
			})
			continue
		}
		records = append(records, record)
	}
	sortScoreRecords(records)
	limit := minUint32(targetActive, uint32(len(records)))
	ranking.Records = records[:limit]
	return ranking, nil
}

func ApplyValidatorSetTransitionLimit(params postypes.Params, previousActive []string, ranking ElectionRanking) (ElectionRanking, error) {
	if len(previousActive) == 0 {
		return ranking, nil
	}
	maxChanges, err := MaxValidatorSetChanges(params, uint32(len(previousActive)))
	if err != nil {
		return ElectionRanking{}, err
	}
	if maxChanges == 0 {
		ranking.MaxValidatorSetChanges = maxChanges
		return ranking, nil
	}
	previous := make(map[string]struct{}, len(previousActive))
	for _, validatorID := range previousActive {
		validatorID = strings.TrimSpace(validatorID)
		if validatorID != "" {
			previous[validatorID] = struct{}{}
		}
	}
	changes := uint32(0)
	limited := make([]ValidatorScoreRecord, 0, len(ranking.Records))
	deferred := make([]ValidatorScoreRecord, 0)
	for _, record := range ranking.Records {
		if _, wasActive := previous[record.ValidatorAddress]; wasActive || changes < maxChanges {
			limited = append(limited, record)
			if !wasActive {
				changes++
			}
			continue
		}
		deferred = append(deferred, record)
	}
	if len(deferred) > 0 {
		ranking.TransitionLimited = true
	}
	ranking.MaxValidatorSetChanges = maxChanges
	ranking.Records = limited
	return ranking, nil
}

func MaxValidatorSetChanges(params postypes.Params, activeValidatorCount uint32) (uint32, error) {
	return postypes.MaxValidatorSetChanges(params, activeValidatorCount)
}

func SimulateScores(input ScoreSimulationInput) (ScoreSimulationResult, error) {
	ranking, err := BuildElectionRanking(input.EpochID, input.Params, input.Candidates, input.TargetActive)
	if err != nil {
		return ScoreSimulationResult{}, err
	}
	ranking, err = ApplyValidatorSetTransitionLimit(input.Params, input.PreviousActive, ranking)
	if err != nil {
		return ScoreSimulationResult{}, err
	}
	result := ScoreSimulationResult{
		Ranking:		ranking,
		TotalRawStakeNaet:	sdkmath.ZeroInt(),
		TotalEffectiveNaet:	sdkmath.ZeroInt(),
	}
	maxRaw := sdkmath.ZeroInt()
	maxEffective := sdkmath.ZeroInt()
	for _, record := range ranking.Records {
		result.ActiveValidatorIDs = append(result.ActiveValidatorIDs, record.ValidatorAddress)
		result.TotalRawStakeNaet = result.TotalRawStakeNaet.Add(record.RawStake)
		result.TotalEffectiveNaet = result.TotalEffectiveNaet.Add(record.EffectiveStake)
		if record.RawStake.GT(maxRaw) {
			maxRaw = record.RawStake
		}
		if record.EffectiveStake.GT(maxEffective) {
			maxEffective = record.EffectiveStake
		}
	}
	result.MaxRawStakeShareBps = shareBps(maxRaw, result.TotalRawStakeNaet)
	result.MaxEffectiveShareBps = shareBps(maxEffective, result.TotalEffectiveNaet)
	result.CentralizationWarning = result.MaxEffectiveShareBps > input.Params.MaxVotingPowerBps
	return result, nil
}

func sortScoreRecords(records []ValidatorScoreRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := records[i]
		right := records[j]
		if !left.ValidatorScore.Equal(right.ValidatorScore) {
			return left.ValidatorScore.GT(right.ValidatorScore)
		}
		if !left.EffectiveStake.Equal(right.EffectiveStake) {
			return left.EffectiveStake.GT(right.EffectiveStake)
		}
		if left.EpochID != right.EpochID {
			return left.EpochID < right.EpochID
		}
		return left.ValidatorAddress < right.ValidatorAddress
	})
}

func scoreRecordKey(epochID uint64, validatorAddress string) string {
	return fmt.Sprintf("%d/%s", epochID, validatorAddress)
}

func shareBps(part sdkmath.Int, total sdkmath.Int) uint32 {
	if !part.IsPositive() || !total.IsPositive() {
		return 0
	}
	if part.GTE(total) {
		return postypes.BasisPoints
	}
	return uint32(part.MulRaw(int64(postypes.BasisPoints)).Quo(total).Uint64())
}

func weightedAverageBps(components []weightedScoreComponent) uint32 {
	total := uint64(0)
	for _, component := range components {
		total += uint64(component.value) * uint64(component.weight)
	}
	return uint32(total / uint64(postypes.BasisPoints))
}

func slashHistoryComponentBps(count uint32) uint32 {
	penalty := count * 2_000
	if penalty >= postypes.BasisPoints {
		return 0
	}
	return postypes.BasisPoints - penalty
}

func metadataCompletenessBps(candidate postypes.Candidate) uint32 {
	score := uint32(0)
	if strings.TrimSpace(candidate.ValidatorID) != "" {
		score += 2_500
	}
	if len(candidate.Roles) > 0 {
		score += 2_500
	}
	if candidate.Capacity.MaxTaskGroups > 0 || len(candidate.Capacity.SupportedWorkloads) > 0 {
		score += 2_500
	}
	if candidate.CommissionBps > 0 {
		score += 2_500
	}
	return score
}

func normalizeOptionalBps(value uint32) uint32 {
	if value == 0 {
		return postypes.BasisPoints
	}
	return value
}

func normalizeEconomyInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}

func minInt(value sdkmath.Int, max int64) int64 {
	if value.LT(sdkmath.ZeroInt()) {
		return 0
	}
	limit := sdkmath.NewInt(max)
	if value.GT(limit) {
		return max
	}
	return value.Int64()
}

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
