package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	ValidatorEconomicEventScored		= "validator_economic_scored"
	ValidatorEconomicEventSelected		= "validator_economic_selected"
	ValidatorEconomicEventConcentration	= "validator_concentration_dampened"
	ValidatorEconomicEventEpochTransition	= "validator_epoch_transition_recommended"

	DefaultValidatorScoreWeightStakeBps		= int64(2_000)
	DefaultValidatorScoreWeightSelfDelegationBps	= int64(1_500)
	DefaultValidatorScoreWeightUptimeBps		= int64(2_000)
	DefaultValidatorScoreWeightMissedBlocksBps	= int64(1_000)
	DefaultValidatorScoreWeightSlashHistoryBps	= int64(1_500)
	DefaultValidatorScoreWeightCommissionBps	= int64(1_000)
	DefaultValidatorScoreWeightConcentrationBps	= int64(750)
	DefaultValidatorScoreWeightMetadataBps		= int64(250)
	DefaultValidatorMinEligibilityScoreBps		= int64(7_000)
	DefaultValidatorMinUptimeBps			= int64(9_500)
	DefaultValidatorSlashPenaltyStepBps		= int64(1_000)
	DefaultValidatorMaxSlashPenaltyBps		= int64(5_000)
	DefaultValidatorMaxCommissionPenaltyBps		= int64(3_000)
	DefaultValidatorMaxEpochChurnBps		= int64(2_500)
	DefaultValidatorMinRewardAdjustmentBps		= int64(7_000)
	DefaultValidatorMaxRewardAdjustmentBps		= int64(10_500)
)

type StakingEnhancementParams struct {
	MinSelfDelegationRatioBps	int64
	MaxCommissionBps		int64
	MaxActiveSet			uint32
	ConcentrationSoftCapBps		int64
	MaxConcentrationDampeningBps	int64
	MinUptimeBps			int64
	MinEligibilityScoreBps		int64
	MaxEpochChurnBps		int64
	EpochLengthBlocks		uint64
	RequireMetadataComplete		bool
	MinRewardAdjustmentBps		int64
	MaxRewardAdjustmentBps		int64
	SlashPenaltyStepBps		int64
	MaxSlashPenaltyBps		int64
	MaxCommissionScorePenaltyBps	int64
	ScoreWeights			ValidatorScoreWeights
}

type ValidatorScoreWeights struct {
	StakeBps		int64
	SelfDelegationBps	int64
	UptimeBps		int64
	MissedBlocksBps		int64
	SlashHistoryBps		int64
	CommissionBps		int64
	ConcentrationBps	int64
	MetadataBps		int64
}

type ValidatorEconomicRecord struct {
	ValidatorID		string
	BondedStakeNaet		sdkmath.Int
	SelfDelegationNaet	sdkmath.Int
	CommissionBps		int64
	UptimeBps		int64
	MissedBlockRateBps	int64
	SlashEvents		uint64
	RecentSlashSeverityBps	int64
	MetadataComplete	bool
	CurrentlyActive		bool
}

type ValidatorEconomicScore struct {
	ValidatorID			string
	Eligible			bool
	RejectReasons			[]string
	BondedStakeNaet			sdkmath.Int
	SelfDelegationRatioBps		int64
	VotingPowerBps			int64
	StakeScoreBps			int64
	SelfDelegationScoreBps		int64
	UptimeScoreBps			int64
	MissedBlockScoreBps		int64
	SlashHistoryScoreBps		int64
	CommissionScoreBps		int64
	ConcentrationScoreBps		int64
	MetadataScoreBps		int64
	ScoreBps			int64
	ConcentrationPenaltyBps		int64
	ReliabilityPenaltyBps		int64
	SlashPenaltyBps			int64
	RewardAdjustmentFactorBps	int64
}

type ValidatorEconomicEvent struct {
	Type				string
	EpochID				uint64
	ValidatorID			string
	Eligible			bool
	ScoreBps			int64
	ConcentrationScoreBps		int64
	RewardAdjustmentFactorBps	int64
	RejectReasons			[]string
}

type ActiveSetRecommendationInput struct {
	EpochID				uint64
	CurrentActiveValidatorIDs	[]string
	Candidates			[]ValidatorEconomicRecord
	Params				StakingEnhancementParams
}

type ActiveSetRecommendation struct {
	EpochID		uint64
	Selected	[]ValidatorEconomicScore
	Standby		[]ValidatorEconomicScore
	AllScores	[]ValidatorEconomicScore
	ChurnCount	uint32
	AllowedChurn	uint32
	Deterministic	bool
	Events		[]ValidatorEconomicEvent
	Failed		[]string
}

type ValidatorDistributionScenario struct {
	Name				string
	CurrentActiveValidatorIDs	[]string
	Candidates			[]ValidatorEconomicRecord
}

type ValidatorDistributionScenarioReport struct {
	Name				string
	Recommendation			ActiveSetRecommendation
	RewardAdjustmentBoundsPassed	bool
	ChurnBounded			bool
	ConcentrationWarnings		uint32
	Warnings			[]string
}

type ValidatorDistributionSimulationInput struct {
	EpochID		uint64
	Scenarios	[]ValidatorDistributionScenario
	Params		StakingEnhancementParams
}

type ValidatorDistributionSimulationReport struct {
	Scenarios	[]ValidatorDistributionScenarioReport
	Passed		bool
	Failed		[]string
}

type StakingEnhancementInvariantReport struct {
	Passed	bool
	Failed	[]string
}

func DefaultStakingEnhancementParams() StakingEnhancementParams {
	return StakingEnhancementParams{
		MinSelfDelegationRatioBps:	DefaultMinSelfDelegationBps,
		MaxCommissionBps:		MaxCommissionBps,
		MaxActiveSet:			uint32(DefaultActiveValidatorTarget),
		ConcentrationSoftCapBps:	MaxTopValidatorConcentrationBps,
		MaxConcentrationDampeningBps:	MaxValidatorRewardDampeningBps,
		MinUptimeBps:			DefaultValidatorMinUptimeBps,
		MinEligibilityScoreBps:		DefaultValidatorMinEligibilityScoreBps,
		MaxEpochChurnBps:		DefaultValidatorMaxEpochChurnBps,
		EpochLengthBlocks:		10_000,
		RequireMetadataComplete:	true,
		MinRewardAdjustmentBps:		DefaultValidatorMinRewardAdjustmentBps,
		MaxRewardAdjustmentBps:		DefaultValidatorMaxRewardAdjustmentBps,
		SlashPenaltyStepBps:		DefaultValidatorSlashPenaltyStepBps,
		MaxSlashPenaltyBps:		DefaultValidatorMaxSlashPenaltyBps,
		MaxCommissionScorePenaltyBps:	DefaultValidatorMaxCommissionPenaltyBps,
		ScoreWeights: ValidatorScoreWeights{
			StakeBps:		DefaultValidatorScoreWeightStakeBps,
			SelfDelegationBps:	DefaultValidatorScoreWeightSelfDelegationBps,
			UptimeBps:		DefaultValidatorScoreWeightUptimeBps,
			MissedBlocksBps:	DefaultValidatorScoreWeightMissedBlocksBps,
			SlashHistoryBps:	DefaultValidatorScoreWeightSlashHistoryBps,
			CommissionBps:		DefaultValidatorScoreWeightCommissionBps,
			ConcentrationBps:	DefaultValidatorScoreWeightConcentrationBps,
			MetadataBps:		DefaultValidatorScoreWeightMetadataBps,
		},
	}
}

func ScoreValidatorEconomics(record ValidatorEconomicRecord, totalBondedStake sdkmath.Int, params StakingEnhancementParams) (ValidatorEconomicScore, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return ValidatorEconomicScore{}, err
	}
	if err := record.Validate(params); err != nil {
		return ValidatorEconomicScore{}, err
	}
	totalStake := normalizeInt(totalBondedStake)
	if !totalStake.IsPositive() {
		return ValidatorEconomicScore{}, fmt.Errorf("total_bonded_stake must be positive")
	}

	bonded := normalizeInt(record.BondedStakeNaet)
	selfDelegation := normalizeInt(record.SelfDelegationNaet)
	votingPowerBps := bonded.MulRaw(BasisPoints).Quo(totalStake).Int64()
	selfRatioBps := int64(0)
	if bonded.IsPositive() {
		selfRatioBps = selfDelegation.MulRaw(BasisPoints).Quo(bonded).Int64()
	}

	concentrationPenalty := concentrationDampeningBps(votingPowerBps, params)
	reliabilityPenalty := clampInt64(maxInt64(params.MinUptimeBps-record.UptimeBps, 0)+record.MissedBlockRateBps/2, 0, params.MaxConcentrationDampeningBps)
	slashPenalty := clampInt64(int64(record.SlashEvents)*params.SlashPenaltyStepBps+record.RecentSlashSeverityBps, 0, params.MaxSlashPenaltyBps)

	score := ValidatorEconomicScore{
		ValidatorID:			record.ValidatorID,
		BondedStakeNaet:		bonded,
		SelfDelegationRatioBps:		selfRatioBps,
		VotingPowerBps:			votingPowerBps,
		StakeScoreBps:			validatorStakeScore(votingPowerBps, params.ConcentrationSoftCapBps),
		SelfDelegationScoreBps:		ratioScore(selfRatioBps, params.MinSelfDelegationRatioBps),
		UptimeScoreBps:			clampInt64(record.UptimeBps, 0, BasisPoints),
		MissedBlockScoreBps:		BasisPoints - clampInt64(record.MissedBlockRateBps, 0, BasisPoints),
		SlashHistoryScoreBps:		BasisPoints - slashPenalty,
		CommissionScoreBps:		BasisPoints - ApplyBps(sdkmath.NewInt(params.MaxCommissionScorePenaltyBps), commissionPressureBps(record.CommissionBps, params.MaxCommissionBps)).Int64(),
		ConcentrationScoreBps:		BasisPoints - concentrationPenalty,
		MetadataScoreBps:		metadataScore(record.MetadataComplete),
		ConcentrationPenaltyBps:	concentrationPenalty,
		ReliabilityPenaltyBps:		reliabilityPenalty,
		SlashPenaltyBps:		slashPenalty,
		RewardAdjustmentFactorBps:	rewardAdjustmentBps(concentrationPenalty, reliabilityPenalty, slashPenalty, params),
	}
	score.ScoreBps = weightedValidatorScore(score, params.ScoreWeights)
	score.RejectReasons = validatorRejectReasons(record, score, params)
	score.Eligible = len(score.RejectReasons) == 0
	return score, nil
}

func RecommendActiveSetForEpoch(input ActiveSetRecommendationInput) (ActiveSetRecommendation, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return ActiveSetRecommendation{}, err
	}
	if input.EpochID == 0 {
		return ActiveSetRecommendation{}, fmt.Errorf("epoch_id must be positive")
	}
	if len(input.Candidates) == 0 {
		return ActiveSetRecommendation{}, fmt.Errorf("candidates are required")
	}

	totalStake := sdkmath.ZeroInt()
	for _, candidate := range input.Candidates {
		totalStake = totalStake.Add(normalizeInt(candidate.BondedStakeNaet))
	}
	if !totalStake.IsPositive() {
		return ActiveSetRecommendation{}, fmt.Errorf("total candidate bonded stake must be positive")
	}

	scores := make([]ValidatorEconomicScore, 0, len(input.Candidates))
	for _, candidate := range input.Candidates {
		score, err := ScoreValidatorEconomics(candidate, totalStake, params)
		if err != nil {
			return ActiveSetRecommendation{}, err
		}
		scores = append(scores, score)
	}
	sortValidatorEconomicScores(scores)

	selected := topEligibleScores(scores, int(params.MaxActiveSet))
	allowedChurn := allowedEpochChurn(input.CurrentActiveValidatorIDs, params)
	if len(input.CurrentActiveValidatorIDs) > 0 {
		selected = boundEpochChurn(selected, scores, input.CurrentActiveValidatorIDs, allowedChurn, int(params.MaxActiveSet))
	}
	sortValidatorEconomicScores(selected)
	selectedIDs := validatorScoreIDSet(selected)

	standby := make([]ValidatorEconomicScore, 0, len(scores)-len(selected))
	for _, score := range scores {
		if !selectedIDs[score.ValidatorID] {
			standby = append(standby, score)
		}
	}

	churn := countNewValidators(selected, input.CurrentActiveValidatorIDs)
	failed := make([]string, 0)
	if uint32(len(selected)) < params.MaxActiveSet && eligibleScoreCount(scores) >= int(params.MaxActiveSet) {
		failed = append(failed, "active_set_below_max")
	}
	if churn > allowedChurn && len(input.CurrentActiveValidatorIDs) > 0 {
		failed = append(failed, "epoch_churn_exceeds_bound")
	}

	events := validatorEconomicEvents(input.EpochID, scores, selectedIDs)
	events = append(events, ValidatorEconomicEvent{
		Type:		ValidatorEconomicEventEpochTransition,
		EpochID:	input.EpochID,
		Eligible:	len(failed) == 0,
		ScoreBps:	averageSelectedScore(selected),
		RejectReasons:	failed,
	})

	return ActiveSetRecommendation{
		EpochID:	input.EpochID,
		Selected:	selected,
		Standby:	standby,
		AllScores:	scores,
		ChurnCount:	churn,
		AllowedChurn:	allowedChurn,
		Deterministic:	true,
		Events:		events,
		Failed:		failed,
	}, nil
}

func RunValidatorDistributionSimulation(input ValidatorDistributionSimulationInput) (ValidatorDistributionSimulationReport, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return ValidatorDistributionSimulationReport{}, err
	}
	if input.EpochID == 0 {
		return ValidatorDistributionSimulationReport{}, fmt.Errorf("epoch_id must be positive")
	}
	if len(input.Scenarios) == 0 {
		return ValidatorDistributionSimulationReport{}, fmt.Errorf("scenarios are required")
	}

	reports := make([]ValidatorDistributionScenarioReport, 0, len(input.Scenarios))
	failed := make([]string, 0)
	for i, scenario := range input.Scenarios {
		if scenario.Name == "" {
			return ValidatorDistributionSimulationReport{}, fmt.Errorf("scenario name is required")
		}
		recommendation, err := RecommendActiveSetForEpoch(ActiveSetRecommendationInput{
			EpochID:			input.EpochID + uint64(i),
			CurrentActiveValidatorIDs:	scenario.CurrentActiveValidatorIDs,
			Candidates:			scenario.Candidates,
			Params:				params,
		})
		if err != nil {
			return ValidatorDistributionSimulationReport{}, err
		}
		invariants := ValidateStakingEnhancementInvariants(recommendation, params)
		warnings := append([]string(nil), recommendation.Failed...)
		if len(recommendation.Selected) < int(params.MaxActiveSet) {
			warnings = append(warnings, "low_participation_active_set_not_full")
		}
		if !invariants.Passed {
			for _, failure := range invariants.Failed {
				failed = append(failed, scenario.Name+":"+failure)
			}
		}
		reports = append(reports, ValidatorDistributionScenarioReport{
			Name:				scenario.Name,
			Recommendation:			recommendation,
			RewardAdjustmentBoundsPassed:	!containsString(invariants.Failed, "reward_adjustment_out_of_bounds"),
			ChurnBounded:			!containsString(invariants.Failed, "epoch_churn_exceeds_bound"),
			ConcentrationWarnings:		concentrationWarningCount(recommendation.Events),
			Warnings:			warnings,
		})
	}

	return ValidatorDistributionSimulationReport{
		Scenarios:	reports,
		Passed:		len(failed) == 0,
		Failed:		failed,
	}, nil
}

func ValidateStakingEnhancementInvariants(recommendation ActiveSetRecommendation, params StakingEnhancementParams) StakingEnhancementInvariantReport {
	params = params.withDefaults()
	failed := make([]string, 0)
	if err := params.Validate(); err != nil {
		return StakingEnhancementInvariantReport{Passed: false, Failed: []string{err.Error()}}
	}
	if uint32(len(recommendation.Selected)) > params.MaxActiveSet {
		failed = append(failed, "active_set_exceeds_max")
	}
	if recommendation.ChurnCount > recommendation.AllowedChurn && recommendation.AllowedChurn > 0 {
		failed = append(failed, "epoch_churn_exceeds_bound")
	}
	for _, score := range recommendation.Selected {
		if !score.Eligible {
			failed = append(failed, "ineligible_validator_selected")
		}
		if score.RewardAdjustmentFactorBps < params.MinRewardAdjustmentBps || score.RewardAdjustmentFactorBps > params.MaxRewardAdjustmentBps {
			failed = append(failed, "reward_adjustment_out_of_bounds")
		}
	}
	if !scoresSortedDeterministically(recommendation.Selected) {
		failed = append(failed, "selected_validators_not_deterministically_sorted")
	}
	if len(recommendation.Events) == 0 {
		failed = append(failed, "validator_economic_events_missing")
	}
	return StakingEnhancementInvariantReport{Passed: len(failed) == 0, Failed: uniqueStrings(failed)}
}

func (p StakingEnhancementParams) Validate() error {
	if err := validateBps("min_self_delegation_ratio_bps", p.MinSelfDelegationRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_commission_bps", p.MaxCommissionBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MaxActiveSet == 0 {
		return fmt.Errorf("max_active_set must be positive")
	}
	if err := validateBps("concentration_soft_cap_bps", p.ConcentrationSoftCapBps, 1, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_concentration_dampening_bps", p.MaxConcentrationDampeningBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("min_uptime_bps", p.MinUptimeBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("min_eligibility_score_bps", p.MinEligibilityScoreBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_epoch_churn_bps", p.MaxEpochChurnBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.EpochLengthBlocks == 0 {
		return fmt.Errorf("epoch_length_blocks must be positive")
	}
	if err := validateBps("min_reward_adjustment_bps", p.MinRewardAdjustmentBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("max_reward_adjustment_bps", p.MaxRewardAdjustmentBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if p.MinRewardAdjustmentBps > p.MaxRewardAdjustmentBps {
		return fmt.Errorf("min_reward_adjustment_bps must be <= max_reward_adjustment_bps")
	}
	if err := validateBps("slash_penalty_step_bps", p.SlashPenaltyStepBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_slash_penalty_bps", p.MaxSlashPenaltyBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_commission_score_penalty_bps", p.MaxCommissionScorePenaltyBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.ScoreWeights.total() <= 0 {
		return fmt.Errorf("score weights must be positive")
	}
	return p.ScoreWeights.Validate()
}

func (p StakingEnhancementParams) withDefaults() StakingEnhancementParams {
	defaults := DefaultStakingEnhancementParams()
	if p.MinSelfDelegationRatioBps == 0 {
		p.MinSelfDelegationRatioBps = defaults.MinSelfDelegationRatioBps
	}
	if p.MaxCommissionBps == 0 {
		p.MaxCommissionBps = defaults.MaxCommissionBps
	}
	if p.MaxActiveSet == 0 {
		p.MaxActiveSet = defaults.MaxActiveSet
	}
	if p.ConcentrationSoftCapBps == 0 {
		p.ConcentrationSoftCapBps = defaults.ConcentrationSoftCapBps
	}
	if p.MaxConcentrationDampeningBps == 0 {
		p.MaxConcentrationDampeningBps = defaults.MaxConcentrationDampeningBps
	}
	if p.MinUptimeBps == 0 {
		p.MinUptimeBps = defaults.MinUptimeBps
	}
	if p.MinEligibilityScoreBps == 0 {
		p.MinEligibilityScoreBps = defaults.MinEligibilityScoreBps
	}
	if p.MaxEpochChurnBps == 0 {
		p.MaxEpochChurnBps = defaults.MaxEpochChurnBps
	}
	if p.EpochLengthBlocks == 0 {
		p.EpochLengthBlocks = defaults.EpochLengthBlocks
	}
	if p.MinRewardAdjustmentBps == 0 {
		p.MinRewardAdjustmentBps = defaults.MinRewardAdjustmentBps
	}
	if p.MaxRewardAdjustmentBps == 0 {
		p.MaxRewardAdjustmentBps = defaults.MaxRewardAdjustmentBps
	}
	if p.SlashPenaltyStepBps == 0 {
		p.SlashPenaltyStepBps = defaults.SlashPenaltyStepBps
	}
	if p.MaxSlashPenaltyBps == 0 {
		p.MaxSlashPenaltyBps = defaults.MaxSlashPenaltyBps
	}
	if p.MaxCommissionScorePenaltyBps == 0 {
		p.MaxCommissionScorePenaltyBps = defaults.MaxCommissionScorePenaltyBps
	}
	if p.ScoreWeights == (ValidatorScoreWeights{}) {
		p.ScoreWeights = defaults.ScoreWeights
	}
	return p
}

func (w ValidatorScoreWeights) Validate() error {
	for _, field := range []struct {
		name	string
		value	int64
	}{
		{name: "score_weight_stake_bps", value: w.StakeBps},
		{name: "score_weight_self_delegation_bps", value: w.SelfDelegationBps},
		{name: "score_weight_uptime_bps", value: w.UptimeBps},
		{name: "score_weight_missed_blocks_bps", value: w.MissedBlocksBps},
		{name: "score_weight_slash_history_bps", value: w.SlashHistoryBps},
		{name: "score_weight_commission_bps", value: w.CommissionBps},
		{name: "score_weight_concentration_bps", value: w.ConcentrationBps},
		{name: "score_weight_metadata_bps", value: w.MetadataBps},
	} {
		if field.value < 0 {
			return fmt.Errorf("%s must not be negative", field.name)
		}
		if field.value > DefaultMaxLoadMultiplierBps {
			return fmt.Errorf("%s exceeds maximum", field.name)
		}
	}
	return nil
}

func (w ValidatorScoreWeights) total() int64 {
	return w.StakeBps + w.SelfDelegationBps + w.UptimeBps + w.MissedBlocksBps + w.SlashHistoryBps + w.CommissionBps + w.ConcentrationBps + w.MetadataBps
}

func (r ValidatorEconomicRecord) Validate(params StakingEnhancementParams) error {
	if r.ValidatorID == "" {
		return fmt.Errorf("validator_id is required")
	}
	if !normalizeInt(r.BondedStakeNaet).IsPositive() {
		return fmt.Errorf("bonded_stake_naet must be positive")
	}
	if normalizeInt(r.SelfDelegationNaet).IsNegative() {
		return fmt.Errorf("self_delegation_naet must not be negative")
	}
	if normalizeInt(r.SelfDelegationNaet).GT(normalizeInt(r.BondedStakeNaet)) {
		return fmt.Errorf("self_delegation_naet must not exceed bonded_stake_naet")
	}
	if err := validateBps("commission_bps", r.CommissionBps, 0, BasisPoints); err != nil {
		return err
	}
	if r.CommissionBps > params.MaxCommissionBps {
		return fmt.Errorf("commission_bps exceeds module maximum")
	}
	if err := validateBps("uptime_bps", r.UptimeBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("missed_block_rate_bps", r.MissedBlockRateBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("recent_slash_severity_bps", r.RecentSlashSeverityBps, 0, BasisPoints); err != nil {
		return err
	}
	return nil
}

func weightedValidatorScore(score ValidatorEconomicScore, weights ValidatorScoreWeights) int64 {
	total := weights.total()
	if total <= 0 {
		return 0
	}
	sum := score.StakeScoreBps*weights.StakeBps +
		score.SelfDelegationScoreBps*weights.SelfDelegationBps +
		score.UptimeScoreBps*weights.UptimeBps +
		score.MissedBlockScoreBps*weights.MissedBlocksBps +
		score.SlashHistoryScoreBps*weights.SlashHistoryBps +
		score.CommissionScoreBps*weights.CommissionBps +
		score.ConcentrationScoreBps*weights.ConcentrationBps +
		score.MetadataScoreBps*weights.MetadataBps
	return clampInt64(sum/total, 0, BasisPoints)
}

func validatorRejectReasons(record ValidatorEconomicRecord, score ValidatorEconomicScore, params StakingEnhancementParams) []string {
	reasons := make([]string, 0)
	if score.SelfDelegationRatioBps < params.MinSelfDelegationRatioBps {
		reasons = append(reasons, "self_delegation_ratio_below_minimum")
	}
	if record.UptimeBps < params.MinUptimeBps {
		reasons = append(reasons, "uptime_below_minimum")
	}
	if params.RequireMetadataComplete && !record.MetadataComplete {
		reasons = append(reasons, "metadata_incomplete")
	}
	if score.ScoreBps < params.MinEligibilityScoreBps {
		reasons = append(reasons, "score_below_minimum")
	}
	if score.SlashPenaltyBps >= params.MaxSlashPenaltyBps {
		reasons = append(reasons, "slash_history_at_penalty_cap")
	}
	return reasons
}

func validatorStakeScore(votingPowerBps, softCapBps int64) int64 {
	if votingPowerBps <= 0 {
		return 0
	}
	if votingPowerBps >= softCapBps {
		return BasisPoints
	}
	return clampInt64(votingPowerBps*BasisPoints/softCapBps, 0, BasisPoints)
}

func ratioScore(valueBps, targetBps int64) int64 {
	if targetBps <= 0 {
		return BasisPoints
	}
	return clampInt64(valueBps*BasisPoints/targetBps, 0, BasisPoints)
}

func metadataScore(complete bool) int64 {
	if complete {
		return BasisPoints
	}
	return 0
}

func commissionPressureBps(commissionBps, maxCommissionBps int64) int64 {
	if maxCommissionBps <= 0 {
		return 0
	}
	return clampInt64(commissionBps*BasisPoints/maxCommissionBps, 0, BasisPoints)
}

func concentrationDampeningBps(votingPowerBps int64, params StakingEnhancementParams) int64 {
	if votingPowerBps <= params.ConcentrationSoftCapBps {
		return 0
	}
	denom := BasisPoints - params.ConcentrationSoftCapBps
	if denom <= 0 {
		return params.MaxConcentrationDampeningBps
	}
	return clampInt64((votingPowerBps-params.ConcentrationSoftCapBps)*params.MaxConcentrationDampeningBps/denom, 0, params.MaxConcentrationDampeningBps)
}

func rewardAdjustmentBps(concentrationPenalty, reliabilityPenalty, slashPenalty int64, params StakingEnhancementParams) int64 {
	raw := BasisPoints - concentrationPenalty - reliabilityPenalty - slashPenalty
	return clampInt64(raw, params.MinRewardAdjustmentBps, params.MaxRewardAdjustmentBps)
}

func topEligibleScores(scores []ValidatorEconomicScore, limit int) []ValidatorEconomicScore {
	selected := make([]ValidatorEconomicScore, 0, limit)
	for _, score := range scores {
		if !score.Eligible {
			continue
		}
		selected = append(selected, score)
		if len(selected) == limit {
			break
		}
	}
	return selected
}

func boundEpochChurn(selected, allScores []ValidatorEconomicScore, currentActive []string, allowedChurn uint32, maxActive int) []ValidatorEconomicScore {
	current := stringSet(currentActive)
	selectedIDs := validatorScoreIDSet(selected)
	newSelected := make([]ValidatorEconomicScore, 0)
	keptSelected := make([]ValidatorEconomicScore, 0)
	for _, score := range selected {
		if current[score.ValidatorID] {
			keptSelected = append(keptSelected, score)
		} else {
			newSelected = append(newSelected, score)
		}
	}
	if uint32(len(newSelected)) <= allowedChurn {
		return selected
	}

	bounded := append([]ValidatorEconomicScore(nil), keptSelected...)
	for i := 0; i < int(allowedChurn) && i < len(newSelected); i++ {
		bounded = append(bounded, newSelected[i])
	}
	boundedIDs := validatorScoreIDSet(bounded)
	for _, score := range allScores {
		if len(bounded) == maxActive {
			break
		}
		if !score.Eligible || boundedIDs[score.ValidatorID] || selectedIDs[score.ValidatorID] {
			continue
		}
		if current[score.ValidatorID] {
			bounded = append(bounded, score)
			boundedIDs[score.ValidatorID] = true
		}
	}
	for _, score := range selected {
		if len(bounded) == maxActive {
			break
		}
		if !boundedIDs[score.ValidatorID] {
			bounded = append(bounded, score)
			boundedIDs[score.ValidatorID] = true
		}
	}
	return bounded
}

func allowedEpochChurn(currentActive []string, params StakingEnhancementParams) uint32 {
	if len(currentActive) == 0 {
		return uint32(params.MaxActiveSet)
	}
	allowed := (uint64(len(currentActive))*uint64(params.MaxEpochChurnBps) + uint64(BasisPoints) - 1) / uint64(BasisPoints)
	if allowed == 0 && params.MaxEpochChurnBps > 0 {
		allowed = 1
	}
	return uint32(allowed)
}

func sortValidatorEconomicScores(scores []ValidatorEconomicScore) {
	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].ScoreBps != scores[j].ScoreBps {
			return scores[i].ScoreBps > scores[j].ScoreBps
		}
		if !scores[i].BondedStakeNaet.Equal(scores[j].BondedStakeNaet) {
			return scores[i].BondedStakeNaet.GT(scores[j].BondedStakeNaet)
		}
		return scores[i].ValidatorID < scores[j].ValidatorID
	})
}

func scoresSortedDeterministically(scores []ValidatorEconomicScore) bool {
	for i := 1; i < len(scores); i++ {
		left := scores[i-1]
		right := scores[i]
		if left.ScoreBps < right.ScoreBps {
			return false
		}
		if left.ScoreBps == right.ScoreBps && left.BondedStakeNaet.LT(right.BondedStakeNaet) {
			return false
		}
		if left.ScoreBps == right.ScoreBps && left.BondedStakeNaet.Equal(right.BondedStakeNaet) && left.ValidatorID > right.ValidatorID {
			return false
		}
	}
	return true
}

func validatorEconomicEvents(epochID uint64, scores []ValidatorEconomicScore, selectedIDs map[string]bool) []ValidatorEconomicEvent {
	events := make([]ValidatorEconomicEvent, 0, len(scores)*2)
	for _, score := range scores {
		events = append(events, ValidatorEconomicEvent{
			Type:				ValidatorEconomicEventScored,
			EpochID:			epochID,
			ValidatorID:			score.ValidatorID,
			Eligible:			score.Eligible,
			ScoreBps:			score.ScoreBps,
			ConcentrationScoreBps:		score.ConcentrationScoreBps,
			RewardAdjustmentFactorBps:	score.RewardAdjustmentFactorBps,
			RejectReasons:			append([]string(nil), score.RejectReasons...),
		})
		if score.ConcentrationPenaltyBps > 0 {
			events = append(events, ValidatorEconomicEvent{
				Type:				ValidatorEconomicEventConcentration,
				EpochID:			epochID,
				ValidatorID:			score.ValidatorID,
				Eligible:			score.Eligible,
				ScoreBps:			score.ScoreBps,
				ConcentrationScoreBps:		score.ConcentrationScoreBps,
				RewardAdjustmentFactorBps:	score.RewardAdjustmentFactorBps,
			})
		}
		if selectedIDs[score.ValidatorID] {
			events = append(events, ValidatorEconomicEvent{
				Type:				ValidatorEconomicEventSelected,
				EpochID:			epochID,
				ValidatorID:			score.ValidatorID,
				Eligible:			true,
				ScoreBps:			score.ScoreBps,
				ConcentrationScoreBps:		score.ConcentrationScoreBps,
				RewardAdjustmentFactorBps:	score.RewardAdjustmentFactorBps,
			})
		}
	}
	return events
}

func validatorScoreIDSet(scores []ValidatorEconomicScore) map[string]bool {
	set := make(map[string]bool, len(scores))
	for _, score := range scores {
		set[score.ValidatorID] = true
	}
	return set
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func countNewValidators(selected []ValidatorEconomicScore, currentActive []string) uint32 {
	current := stringSet(currentActive)
	count := uint32(0)
	for _, score := range selected {
		if !current[score.ValidatorID] {
			count++
		}
	}
	return count
}

func eligibleScoreCount(scores []ValidatorEconomicScore) int {
	count := 0
	for _, score := range scores {
		if score.Eligible {
			count++
		}
	}
	return count
}

func averageSelectedScore(selected []ValidatorEconomicScore) int64 {
	if len(selected) == 0 {
		return 0
	}
	total := int64(0)
	for _, score := range selected {
		total += score.ScoreBps
	}
	return total / int64(len(selected))
}

func concentrationWarningCount(events []ValidatorEconomicEvent) uint32 {
	count := uint32(0)
	for _, event := range events {
		if event.Type == ValidatorEconomicEventConcentration {
			count++
		}
	}
	return count
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}
