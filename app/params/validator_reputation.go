package params

import (
	"fmt"
	"sort"
)

const (
	ValidatorReputationScoringVersionV1	= "validator-reputation/v1"

	ReputationComponentConsensusSafe	= "consensus_safe"
	ReputationComponentAdvisoryOnly		= "advisory_only"

	DefaultReputationMissingDataPenaltyBps		= int64(250)
	DefaultCaptureDelegationFlowThresholdBps	= int64(2_000)
	DefaultCaptureCommissionChangeThresholdBps	= int64(500)
	DefaultCaptureSelfDelegationExitThresholdBps	= int64(1_000)
	DefaultRewardPerformanceTargetBps		= int64(10_000)
)

type ValidatorReputationWeights struct {
	ReliabilityBps		int64
	SlashHistoryBps		int64
	CommissionStabilityBps	int64
	SelfDelegationBps	int64
	MetadataBps		int64
	DelegationFlowBps	int64
	RewardPerformanceBps	int64
	ConcentrationBps	int64
}

type ValidatorReputationParams struct {
	ScoringVersion				string
	ConcentrationSoftCapBps			int64
	MissingDataPenaltyBps			int64
	CaptureDelegationFlowThresholdBps	int64
	CaptureCommissionChangeThresholdBps	int64
	CaptureSelfDelegationExitThresholdBps	int64
	RewardPerformanceTargetBps		int64
	AllowAdvisoryInputsInConsensus		bool
	Weights					ValidatorReputationWeights
}

type ValidatorReputationInput struct {
	ValidatorID			string
	UptimeHistoryBps		[]int64
	MissedBlockRateHistoryBps	[]int64
	SlashEvents			uint64
	SlashSeverityBps		int64
	CommissionChangeHistoryBps	[]int64
	SelfDelegationChangeHistoryBps	[]int64
	MetadataChangeCount		uint64
	MetadataComplete		bool
	DelegationInflowBps		int64
	DelegationOutflowBps		int64
	HistoricalRewardPerformanceBps	[]int64
	VotingPowerBps			int64
	UseInConsensus			bool
	AdvisoryInputsDeterministic	bool
}

type ValidatorReputationComponent struct {
	Name		string
	Source		string
	ScoreBps	int64
	WeightBps	int64
	Explanation	string
	Missing		bool
}

type ValidatorDelegatorMetadata struct {
	ValidatorID			string
	ScoringVersion			string
	RiskScoreBps			int64
	ReliabilityScoreBps		int64
	CommissionStabilityScoreBps	int64
	ConcentrationWarning		bool
	CaptureRiskWarning		bool
	AdvisoryOnly			bool
	Explanation			[]string
}

type ValidatorReputationReport struct {
	ValidatorID			string
	ScoringVersion			string
	RiskScoreBps			int64
	ReliabilityScoreBps		int64
	CommissionStabilityScoreBps	int64
	ConsensusSafeScoreBps		int64
	AdvisoryScoreBps		int64
	ConcentrationWarning		bool
	CaptureRiskWarning		bool
	ConsensusSafeForSelection	bool
	ConsensusComponents		[]ValidatorReputationComponent
	AdvisoryComponents		[]ValidatorReputationComponent
	DelegatorMetadata		ValidatorDelegatorMetadata
	ScoreExplanation		[]string
	Failed				[]string
}

func DefaultValidatorReputationParams() ValidatorReputationParams {
	return ValidatorReputationParams{
		ScoringVersion:				ValidatorReputationScoringVersionV1,
		ConcentrationSoftCapBps:		MaxTopValidatorConcentrationBps,
		MissingDataPenaltyBps:			DefaultReputationMissingDataPenaltyBps,
		CaptureDelegationFlowThresholdBps:	DefaultCaptureDelegationFlowThresholdBps,
		CaptureCommissionChangeThresholdBps:	DefaultCaptureCommissionChangeThresholdBps,
		CaptureSelfDelegationExitThresholdBps:	DefaultCaptureSelfDelegationExitThresholdBps,
		RewardPerformanceTargetBps:		DefaultRewardPerformanceTargetBps,
		Weights: ValidatorReputationWeights{
			ReliabilityBps:		2_500,
			SlashHistoryBps:	1_500,
			CommissionStabilityBps:	1_500,
			SelfDelegationBps:	1_000,
			MetadataBps:		750,
			DelegationFlowBps:	1_000,
			RewardPerformanceBps:	1_000,
			ConcentrationBps:	750,
		},
	}
}

func EvaluateValidatorReputation(input ValidatorReputationInput, params ValidatorReputationParams) (ValidatorReputationReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return ValidatorReputationReport{}, err
	}
	if err := input.Validate(); err != nil {
		return ValidatorReputationReport{}, err
	}

	reliabilityScore, reliabilityMissing := reliabilityReputationScore(input, params)
	slashScore := BasisPoints - clampInt64(int64(input.SlashEvents)*DefaultValidatorSlashPenaltyStepBps+input.SlashSeverityBps, 0, BasisPoints)
	commissionScore, commissionMissing := stabilityScore(input.CommissionChangeHistoryBps, params.MissingDataPenaltyBps)
	selfDelegationScore, selfDelegationMissing := stabilityScore(input.SelfDelegationChangeHistoryBps, params.MissingDataPenaltyBps)
	metadataScoreValue := reputationMetadataScore(input.MetadataComplete, input.MetadataChangeCount, params)
	delegationFlowScore := BasisPoints - clampInt64(maxInt64(input.DelegationInflowBps, input.DelegationOutflowBps), 0, BasisPoints)
	rewardPerformanceScore, rewardMissing := rewardPerformanceScore(input.HistoricalRewardPerformanceBps, params)
	concentrationScore := BasisPoints - concentrationDampeningBps(input.VotingPowerBps, StakingEnhancementParams{
		ConcentrationSoftCapBps:	params.ConcentrationSoftCapBps,
		MaxConcentrationDampeningBps:	MaxValidatorRewardDampeningBps,
	})

	consensusComponents := []ValidatorReputationComponent{
		reputationComponent("reliability", ReputationComponentConsensusSafe, reliabilityScore, params.Weights.ReliabilityBps, "uptime history, missed blocks, and slash penalties are deterministic validator performance signals", reliabilityMissing),
		reputationComponent("slash_history", ReputationComponentConsensusSafe, slashScore, params.Weights.SlashHistoryBps, "accepted slash events reduce reliability confidence", false),
	}
	advisoryComponents := []ValidatorReputationComponent{
		reputationComponent("commission_stability", ReputationComponentAdvisoryOnly, commissionScore, params.Weights.CommissionStabilityBps, "large commission changes are surfaced as delegator risk", commissionMissing),
		reputationComponent("self_delegation_stability", ReputationComponentAdvisoryOnly, selfDelegationScore, params.Weights.SelfDelegationBps, "self-delegation withdrawals can signal weaker operator alignment", selfDelegationMissing),
		reputationComponent("metadata_stability", ReputationComponentAdvisoryOnly, metadataScoreValue, params.Weights.MetadataBps, "metadata completeness and churn are wallet and explorer risk signals", false),
		reputationComponent("delegation_flow", ReputationComponentAdvisoryOnly, delegationFlowScore, params.Weights.DelegationFlowBps, "sudden delegation inflow or outflow is treated as capture-risk telemetry", false),
		reputationComponent("reward_performance", ReputationComponentAdvisoryOnly, rewardPerformanceScore, params.Weights.RewardPerformanceBps, "historical reward performance helps delegators compare realized outcomes", rewardMissing),
		reputationComponent("concentration", ReputationComponentAdvisoryOnly, concentrationScore, params.Weights.ConcentrationBps, "voting-power concentration is visible before delegation", false),
	}
	sortReputationComponents(consensusComponents)
	sortReputationComponents(advisoryComponents)

	consensusScore := weightedReputationScore(consensusComponents)
	advisoryScore := weightedReputationScore(advisoryComponents)
	overallScore := weightedReputationScore(append(append([]ValidatorReputationComponent{}, consensusComponents...), advisoryComponents...))
	riskScore := BasisPoints - overallScore
	concentrationWarning := input.VotingPowerBps > params.ConcentrationSoftCapBps
	captureWarning := captureRiskWarning(input, params)

	failed := make([]string, 0)
	consensusSafeForSelection := true
	if input.UseInConsensus && len(advisoryComponents) > 0 && (!params.AllowAdvisoryInputsInConsensus || !input.AdvisoryInputsDeterministic) {
		consensusSafeForSelection = false
		failed = append(failed, "advisory_reputation_inputs_not_consensus_safe")
	}

	explanations := reputationExplanations(consensusComponents, advisoryComponents, concentrationWarning, captureWarning, input.UseInConsensus)
	metadata := ValidatorDelegatorMetadata{
		ValidatorID:			input.ValidatorID,
		ScoringVersion:			params.ScoringVersion,
		RiskScoreBps:			riskScore,
		ReliabilityScoreBps:		reliabilityScore,
		CommissionStabilityScoreBps:	commissionScore,
		ConcentrationWarning:		concentrationWarning,
		CaptureRiskWarning:		captureWarning,
		AdvisoryOnly:			!consensusSafeForSelection || !input.UseInConsensus,
		Explanation:			explanations,
	}

	return ValidatorReputationReport{
		ValidatorID:			input.ValidatorID,
		ScoringVersion:			params.ScoringVersion,
		RiskScoreBps:			riskScore,
		ReliabilityScoreBps:		reliabilityScore,
		CommissionStabilityScoreBps:	commissionScore,
		ConsensusSafeScoreBps:		consensusScore,
		AdvisoryScoreBps:		advisoryScore,
		ConcentrationWarning:		concentrationWarning,
		CaptureRiskWarning:		captureWarning,
		ConsensusSafeForSelection:	consensusSafeForSelection,
		ConsensusComponents:		consensusComponents,
		AdvisoryComponents:		advisoryComponents,
		DelegatorMetadata:		metadata,
		ScoreExplanation:		explanations,
		Failed:				failed,
	}, nil
}

func (p ValidatorReputationParams) Validate() error {
	if p.ScoringVersion == "" {
		return fmt.Errorf("scoring_version is required")
	}
	if err := validateBps("concentration_soft_cap_bps", p.ConcentrationSoftCapBps, 1, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("missing_data_penalty_bps", p.MissingDataPenaltyBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("capture_delegation_flow_threshold_bps", p.CaptureDelegationFlowThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("capture_commission_change_threshold_bps", p.CaptureCommissionChangeThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("capture_self_delegation_exit_threshold_bps", p.CaptureSelfDelegationExitThresholdBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("reward_performance_target_bps", p.RewardPerformanceTargetBps, 1, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	return p.Weights.Validate()
}

func (p ValidatorReputationParams) withDefaults() ValidatorReputationParams {
	defaults := DefaultValidatorReputationParams()
	if p.ScoringVersion == "" {
		p.ScoringVersion = defaults.ScoringVersion
	}
	if p.ConcentrationSoftCapBps == 0 {
		p.ConcentrationSoftCapBps = defaults.ConcentrationSoftCapBps
	}
	if p.MissingDataPenaltyBps == 0 {
		p.MissingDataPenaltyBps = defaults.MissingDataPenaltyBps
	}
	if p.CaptureDelegationFlowThresholdBps == 0 {
		p.CaptureDelegationFlowThresholdBps = defaults.CaptureDelegationFlowThresholdBps
	}
	if p.CaptureCommissionChangeThresholdBps == 0 {
		p.CaptureCommissionChangeThresholdBps = defaults.CaptureCommissionChangeThresholdBps
	}
	if p.CaptureSelfDelegationExitThresholdBps == 0 {
		p.CaptureSelfDelegationExitThresholdBps = defaults.CaptureSelfDelegationExitThresholdBps
	}
	if p.RewardPerformanceTargetBps == 0 {
		p.RewardPerformanceTargetBps = defaults.RewardPerformanceTargetBps
	}
	if p.Weights == (ValidatorReputationWeights{}) {
		p.Weights = defaults.Weights
	}
	return p
}

func (w ValidatorReputationWeights) Validate() error {
	for _, field := range []struct {
		name	string
		value	int64
	}{
		{name: "reliability_weight_bps", value: w.ReliabilityBps},
		{name: "slash_history_weight_bps", value: w.SlashHistoryBps},
		{name: "commission_stability_weight_bps", value: w.CommissionStabilityBps},
		{name: "self_delegation_weight_bps", value: w.SelfDelegationBps},
		{name: "metadata_weight_bps", value: w.MetadataBps},
		{name: "delegation_flow_weight_bps", value: w.DelegationFlowBps},
		{name: "reward_performance_weight_bps", value: w.RewardPerformanceBps},
		{name: "concentration_weight_bps", value: w.ConcentrationBps},
	} {
		if field.value < 0 {
			return fmt.Errorf("%s must not be negative", field.name)
		}
		if field.value > DefaultMaxLoadMultiplierBps {
			return fmt.Errorf("%s exceeds maximum", field.name)
		}
	}
	if w.total() <= 0 {
		return fmt.Errorf("reputation weights must be positive")
	}
	return nil
}

func (w ValidatorReputationWeights) total() int64 {
	return w.ReliabilityBps +
		w.SlashHistoryBps +
		w.CommissionStabilityBps +
		w.SelfDelegationBps +
		w.MetadataBps +
		w.DelegationFlowBps +
		w.RewardPerformanceBps +
		w.ConcentrationBps
}

func (input ValidatorReputationInput) Validate() error {
	if input.ValidatorID == "" {
		return fmt.Errorf("validator_id is required")
	}
	for name, values := range map[string][]int64{
		"uptime_history_bps":			input.UptimeHistoryBps,
		"missed_block_rate_history_bps":	input.MissedBlockRateHistoryBps,
		"commission_change_history_bps":	input.CommissionChangeHistoryBps,
		"historical_reward_performance_bps":	input.HistoricalRewardPerformanceBps,
	} {
		for i, value := range values {
			if err := validateBps(fmt.Sprintf("%s[%d]", name, i), value, 0, DefaultMaxLoadMultiplierBps); err != nil {
				return err
			}
		}
	}
	for i, value := range input.SelfDelegationChangeHistoryBps {
		if absInt64(value) > DefaultMaxLoadMultiplierBps {
			return fmt.Errorf("self_delegation_change_history_bps[%d] exceeds maximum", i)
		}
	}
	if err := validateBps("slash_severity_bps", input.SlashSeverityBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("delegation_inflow_bps", input.DelegationInflowBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("delegation_outflow_bps", input.DelegationOutflowBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("voting_power_bps", input.VotingPowerBps, 0, BasisPoints); err != nil {
		return err
	}
	return nil
}

func reliabilityReputationScore(input ValidatorReputationInput, params ValidatorReputationParams) (int64, bool) {
	missing := len(input.UptimeHistoryBps) == 0 || len(input.MissedBlockRateHistoryBps) == 0
	uptime := averageBpsWithDefault(input.UptimeHistoryBps, DefaultValidatorReliabilityTargetBps)
	missed := averageBpsWithDefault(input.MissedBlockRateHistoryBps, 0)
	slashPenalty := clampInt64(int64(input.SlashEvents)*params.MissingDataPenaltyBps+input.SlashSeverityBps, 0, BasisPoints)
	score := uptime - missed/2 - slashPenalty
	if missing {
		score -= params.MissingDataPenaltyBps
	}
	return clampInt64(score, 0, BasisPoints), missing
}

func stabilityScore(changes []int64, missingPenaltyBps int64) (int64, bool) {
	if len(changes) == 0 {
		return clampInt64(BasisPoints-missingPenaltyBps, 0, BasisPoints), true
	}
	total := int64(0)
	for _, change := range changes {
		total += absInt64(change)
	}
	averageChange := total / int64(len(changes))
	return BasisPoints - clampInt64(averageChange, 0, BasisPoints), false
}

func rewardPerformanceScore(values []int64, params ValidatorReputationParams) (int64, bool) {
	if len(values) == 0 {
		return clampInt64(BasisPoints-params.MissingDataPenaltyBps, 0, BasisPoints), true
	}
	average := averageBpsWithDefault(values, params.RewardPerformanceTargetBps)
	return clampInt64(average*BasisPoints/params.RewardPerformanceTargetBps, 0, BasisPoints), false
}

func reputationMetadataScore(complete bool, changes uint64, params ValidatorReputationParams) int64 {
	score := int64(BasisPoints)
	if !complete {
		score -= params.MissingDataPenaltyBps * 2
	}
	changePenalty := int64(changes) * params.MissingDataPenaltyBps
	if changePenalty > BasisPoints {
		changePenalty = BasisPoints
	}
	return clampInt64(score-changePenalty, 0, BasisPoints)
}

func reputationComponent(name, source string, score, weight int64, explanation string, missing bool) ValidatorReputationComponent {
	return ValidatorReputationComponent{
		Name:		name,
		Source:		source,
		ScoreBps:	clampInt64(score, 0, BasisPoints),
		WeightBps:	weight,
		Explanation:	explanation,
		Missing:	missing,
	}
}

func weightedReputationScore(components []ValidatorReputationComponent) int64 {
	totalWeight := int64(0)
	total := int64(0)
	for _, component := range components {
		if component.WeightBps <= 0 {
			continue
		}
		totalWeight += component.WeightBps
		total += component.ScoreBps * component.WeightBps
	}
	if totalWeight == 0 {
		return 0
	}
	return clampInt64(total/totalWeight, 0, BasisPoints)
}

func captureRiskWarning(input ValidatorReputationInput, params ValidatorReputationParams) bool {
	if input.DelegationInflowBps >= params.CaptureDelegationFlowThresholdBps || input.DelegationOutflowBps >= params.CaptureDelegationFlowThresholdBps {
		return true
	}
	if maxAbs(input.CommissionChangeHistoryBps) >= params.CaptureCommissionChangeThresholdBps {
		return true
	}
	if maxNegativeAbs(input.SelfDelegationChangeHistoryBps) >= params.CaptureSelfDelegationExitThresholdBps {
		return true
	}
	if input.MetadataChangeCount > 0 || input.SlashEvents > 0 {
		return true
	}
	return false
}

func reputationExplanations(consensus, advisory []ValidatorReputationComponent, concentrationWarning, captureWarning, useInConsensus bool) []string {
	explanations := make([]string, 0, len(consensus)+len(advisory)+3)
	for _, component := range append(append([]ValidatorReputationComponent{}, consensus...), advisory...) {
		text := component.Name + ":" + component.Source
		if component.Missing {
			text += ":missing_data_penalized"
		}
		explanations = append(explanations, text)
	}
	if concentrationWarning {
		explanations = append(explanations, "concentration_warning")
	}
	if captureWarning {
		explanations = append(explanations, "capture_risk_warning")
	}
	if useInConsensus {
		explanations = append(explanations, "consensus_use_requires_consensus_safe_components_only")
	}
	sort.Strings(explanations)
	return explanations
}

func sortReputationComponents(components []ValidatorReputationComponent) {
	sort.SliceStable(components, func(i, j int) bool {
		return components[i].Name < components[j].Name
	})
}

func averageBpsWithDefault(values []int64, fallback int64) int64 {
	if len(values) == 0 {
		return fallback
	}
	total := int64(0)
	for _, value := range values {
		total += value
	}
	return total / int64(len(values))
}

func maxAbs(values []int64) int64 {
	maxValue := int64(0)
	for _, value := range values {
		if absInt64(value) > maxValue {
			maxValue = absInt64(value)
		}
	}
	return maxValue
}

func maxNegativeAbs(values []int64) int64 {
	maxValue := int64(0)
	for _, value := range values {
		if value < 0 && absInt64(value) > maxValue {
			maxValue = absInt64(value)
		}
	}
	return maxValue
}
