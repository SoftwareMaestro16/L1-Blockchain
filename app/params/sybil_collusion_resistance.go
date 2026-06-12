package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	CorrelationSignalCommission	= "synchronized_commission_change"
	CorrelationSignalDowntime	= "correlated_downtime"
	CorrelationSignalOperator	= "deterministic_operator_group"

	EvidenceRouteEquivocation	= "equivocation_evidence_route"
	EvidenceRouteDowntime		= "downtime_evidence_route"
	EvidenceRouteDuplicate		= "duplicate_evidence_rejected"

	DefaultSybilGroupSoftCapBps			= int64(2_500)
	DefaultSybilGroupMaxRewardDampeningBps		= int64(3_000)
	DefaultSybilDeterministicCorrelationBps		= int64(9_000)
	DefaultSybilAdvisoryCorrelationBps		= int64(6_000)
	DefaultBootstrapEligibilityWindowEpochs		= uint64(12)
	DefaultTopNConcentrationLimitBps		= int64(6_700)
	DefaultSynchronizedCommissionThresholdBps	= int64(300)
	DefaultCorrelatedDowntimeThresholdBps		= int64(7_500)
	DefaultEvidenceRoutingRewardCapNaet		= int64(1_000)
)

type SybilCollusionResistanceParams struct {
	GroupConcentrationSoftCapBps		int64
	GroupMaxRewardDampeningBps		int64
	DeterministicCorrelationThresholdBps	int64
	AdvisoryCorrelationThresholdBps		int64
	MinSelfDelegationNaet			sdkmath.Int
	MinSelfDelegationRatioBps		int64
	MinPerformanceBps			int64
	BootstrapEligibilityWindowEpochs	uint64
	BootstrapMaxStakeBps			int64
	BootstrapBonusBps			int64
	TopNConcentrationLimitBps		int64
	SynchronizedCommissionThresholdBps	int64
	CorrelatedDowntimeThresholdBps		int64
	EvidenceRoutingRewardCapNaet		sdkmath.Int
}

type ValidatorSybilRecord struct {
	ValidatorID			string
	EconomicGroupID			string
	StakeNaet			sdkmath.Int
	VotingPowerBps			int64
	SelfDelegationNaet		sdkmath.Int
	SelfDelegationRatioBps		int64
	PerformanceBps			int64
	CommissionBps			int64
	CommissionChangeBps		int64
	DowntimeObserved		bool
	BootstrapWindowID		uint64
	BootstrapRewardAlreadyUsed	bool
	CorrelationScoreBps		int64
	CorrelationProofDeterministic	bool
}

type StakeSplitSimulationInput struct {
	RewardPoolNaet	sdkmath.Int
	Original	ValidatorSybilRecord
	Split		[]ValidatorSybilRecord
	Params		SybilCollusionResistanceParams
}

type StakeSplitSimulationReport struct {
	OriginalGroupPowerBps		int64
	SplitGroupPowerBps		int64
	OriginalRewardNaet		sdkmath.Int
	SplitRewardNaet			sdkmath.Int
	OriginalDampeningBps		int64
	SplitDampeningBps		int64
	BypassPrevented			bool
	ConsensusDampeningApplied	bool
	AdvisoryOnly			bool
}

type ValidatorEligibilityReport struct {
	ValidatorID		string
	Eligible		bool
	BootstrapEligible	bool
	BootstrapBonusBps	int64
	RejectReasons		[]string
	AdvisoryCorrelation	bool
	ConsensusDampeningBps	int64
}

type CorrelationTelemetryReport struct {
	GroupID			string
	Validators		[]string
	CorrelationScoreBps	int64
	DeterministicProof	bool
	AdvisoryOnly		bool
	ConsensusAffecting	bool
	RewardDampeningBps	int64
	Signals			[]string
}

type CollusionScenarioInput struct {
	Validators	[]ValidatorSybilRecord
	TopN		int
	RewardPoolNaet	sdkmath.Int
	Params		SybilCollusionResistanceParams
}

type CollusionScenarioReport struct {
	TopNVotingPowerBps		int64
	TopNLimitBps			int64
	ConcentrationExceeded		bool
	ConcentrationDampeningBps	int64
	RewardBeforeDampeningNaet	sdkmath.Int
	RewardAfterDampeningNaet	sdkmath.Int
	CommissionAlerts		[]string
	DowntimeAlerts			[]string
	CorrelationReports		[]CorrelationTelemetryReport
	GovernanceSummary		string
}

type EvidenceRoutingInput struct {
	ValidatorID			string
	EvidenceType			string
	Duplicate			bool
	Accepted			bool
	RequestedReporterRewardNaet	sdkmath.Int
	Params				SybilCollusionResistanceParams
}

type EvidenceRoutingReport struct {
	Route			string
	Accepted		bool
	ReporterRewardNaet	sdkmath.Int
	RejectedReason		string
	Auditable		bool
}

func DefaultSybilCollusionResistanceParams() SybilCollusionResistanceParams {
	return SybilCollusionResistanceParams{
		GroupConcentrationSoftCapBps:		DefaultSybilGroupSoftCapBps,
		GroupMaxRewardDampeningBps:		DefaultSybilGroupMaxRewardDampeningBps,
		DeterministicCorrelationThresholdBps:	DefaultSybilDeterministicCorrelationBps,
		AdvisoryCorrelationThresholdBps:	DefaultSybilAdvisoryCorrelationBps,
		MinSelfDelegationNaet:			sdkmath.NewInt(1_000),
		MinSelfDelegationRatioBps:		DefaultMinSelfDelegationBps,
		MinPerformanceBps:			DefaultValidatorReliabilityTargetBps,
		BootstrapEligibilityWindowEpochs:	DefaultBootstrapEligibilityWindowEpochs,
		BootstrapMaxStakeBps:			DefaultMaxBootstrapStakeBps,
		BootstrapBonusBps:			DefaultValidatorBootstrapBonusBps,
		TopNConcentrationLimitBps:		DefaultTopNConcentrationLimitBps,
		SynchronizedCommissionThresholdBps:	DefaultSynchronizedCommissionThresholdBps,
		CorrelatedDowntimeThresholdBps:		DefaultCorrelatedDowntimeThresholdBps,
		EvidenceRoutingRewardCapNaet:		sdkmath.NewInt(DefaultEvidenceRoutingRewardCapNaet),
	}
}

func SimulateStakeSplitting(input StakeSplitSimulationInput) (StakeSplitSimulationReport, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return StakeSplitSimulationReport{}, err
	}
	if err := input.Original.Validate(); err != nil {
		return StakeSplitSimulationReport{}, err
	}
	if len(input.Split) == 0 {
		return StakeSplitSimulationReport{}, fmt.Errorf("split validators are required")
	}
	rewardPool := normalizeInt(input.RewardPoolNaet)
	if rewardPool.IsNegative() {
		return StakeSplitSimulationReport{}, fmt.Errorf("reward_pool_naet must not be negative")
	}
	originalPower := input.Original.VotingPowerBps
	originalDampening, originalConsensus, originalAdvisory := groupDampening(originalPower, input.Original.CorrelationScoreBps, input.Original.CorrelationProofDeterministic, params)
	originalReward := ApplyBps(ApplyBps(rewardPool, originalPower), BasisPoints-originalDampening)

	splitPower := int64(0)
	deterministic := false
	advisory := false
	maxCorrelation := int64(0)
	for _, validator := range input.Split {
		if err := validator.Validate(); err != nil {
			return StakeSplitSimulationReport{}, err
		}
		splitPower += validator.VotingPowerBps
		deterministic = deterministic || validator.CorrelationProofDeterministic
		advisory = advisory || validator.CorrelationScoreBps >= params.AdvisoryCorrelationThresholdBps
		if validator.CorrelationScoreBps > maxCorrelation {
			maxCorrelation = validator.CorrelationScoreBps
		}
	}
	splitDampening, splitConsensus, splitAdvisory := groupDampening(splitPower, maxCorrelation, deterministic, params)
	splitReward := ApplyBps(ApplyBps(rewardPool, splitPower), BasisPoints-splitDampening)
	return StakeSplitSimulationReport{
		OriginalGroupPowerBps:		originalPower,
		SplitGroupPowerBps:		splitPower,
		OriginalRewardNaet:		originalReward,
		SplitRewardNaet:		splitReward,
		OriginalDampeningBps:		originalDampening,
		SplitDampeningBps:		splitDampening,
		BypassPrevented:		splitPower == originalPower && splitReward.LTE(originalReward),
		ConsensusDampeningApplied:	originalConsensus || splitConsensus,
		AdvisoryOnly:			(originalAdvisory || splitAdvisory || advisory) && !(originalConsensus || splitConsensus),
	}, nil
}

func EvaluateValidatorActiveSetEligibility(validator ValidatorSybilRecord, params SybilCollusionResistanceParams) (ValidatorEligibilityReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return ValidatorEligibilityReport{}, err
	}
	if err := validator.Validate(); err != nil {
		return ValidatorEligibilityReport{}, err
	}
	reasons := make([]string, 0)
	if normalizeInt(validator.SelfDelegationNaet).LT(normalizeInt(params.MinSelfDelegationNaet)) {
		reasons = append(reasons, "self_delegation_below_minimum")
	}
	if validator.SelfDelegationRatioBps < params.MinSelfDelegationRatioBps {
		reasons = append(reasons, "self_delegation_ratio_below_minimum")
	}
	if validator.PerformanceBps < params.MinPerformanceBps {
		reasons = append(reasons, "performance_below_minimum")
	}
	bootstrap := len(reasons) == 0 &&
		!validator.BootstrapRewardAlreadyUsed &&
		validator.BootstrapWindowID > 0 &&
		validator.VotingPowerBps <= params.BootstrapMaxStakeBps
	dampening, consensus, advisory := groupDampening(validator.VotingPowerBps, validator.CorrelationScoreBps, validator.CorrelationProofDeterministic, params)
	return ValidatorEligibilityReport{
		ValidatorID:		validator.ValidatorID,
		Eligible:		len(reasons) == 0,
		BootstrapEligible:	bootstrap,
		BootstrapBonusBps:	bootstrapBps(bootstrap, params),
		RejectReasons:		reasons,
		AdvisoryCorrelation:	advisory && !consensus,
		ConsensusDampeningBps:	dampening,
	}, nil
}

func EvaluateCorrelationTelemetry(groupID string, validators []ValidatorSybilRecord, params SybilCollusionResistanceParams) (CorrelationTelemetryReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return CorrelationTelemetryReport{}, err
	}
	if groupID == "" {
		return CorrelationTelemetryReport{}, fmt.Errorf("group_id is required")
	}
	ids := make([]string, 0, len(validators))
	groupPower := int64(0)
	maxScore := int64(0)
	deterministic := false
	signals := make([]string, 0)
	commissionChanges := 0
	downtime := 0
	for _, validator := range validators {
		if err := validator.Validate(); err != nil {
			return CorrelationTelemetryReport{}, err
		}
		ids = append(ids, validator.ValidatorID)
		groupPower += validator.VotingPowerBps
		if validator.CorrelationScoreBps > maxScore {
			maxScore = validator.CorrelationScoreBps
		}
		deterministic = deterministic || validator.CorrelationProofDeterministic
		if validator.CommissionChangeBps >= params.SynchronizedCommissionThresholdBps {
			commissionChanges++
		}
		if validator.DowntimeObserved {
			downtime++
		}
	}
	if commissionChanges >= 2 {
		signals = append(signals, CorrelationSignalCommission)
	}
	if len(validators) > 0 && int64(downtime)*BasisPoints/int64(len(validators)) >= params.CorrelatedDowntimeThresholdBps {
		signals = append(signals, CorrelationSignalDowntime)
	}
	if deterministic {
		signals = append(signals, CorrelationSignalOperator)
	}
	dampening, consensus, advisory := groupDampening(groupPower, maxScore, deterministic, params)
	sort.Strings(ids)
	return CorrelationTelemetryReport{
		GroupID:		groupID,
		Validators:		ids,
		CorrelationScoreBps:	maxScore,
		DeterministicProof:	deterministic,
		AdvisoryOnly:		advisory && !consensus,
		ConsensusAffecting:	consensus,
		RewardDampeningBps:	dampening,
		Signals:		signals,
	}, nil
}

func SimulateValidatorCollusion(input CollusionScenarioInput) (CollusionScenarioReport, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return CollusionScenarioReport{}, err
	}
	if input.TopN <= 0 {
		return CollusionScenarioReport{}, fmt.Errorf("top_n must be positive")
	}
	if len(input.Validators) == 0 {
		return CollusionScenarioReport{}, fmt.Errorf("validators are required")
	}
	validators := append([]ValidatorSybilRecord(nil), input.Validators...)
	for _, validator := range validators {
		if err := validator.Validate(); err != nil {
			return CollusionScenarioReport{}, err
		}
	}
	sort.SliceStable(validators, func(i, j int) bool {
		if validators[i].VotingPowerBps == validators[j].VotingPowerBps {
			return validators[i].ValidatorID < validators[j].ValidatorID
		}
		return validators[i].VotingPowerBps > validators[j].VotingPowerBps
	})
	limit := input.TopN
	if limit > len(validators) {
		limit = len(validators)
	}
	topPower := int64(0)
	for i := 0; i < limit; i++ {
		topPower += validators[i].VotingPowerBps
	}
	dampening := int64(0)
	if topPower > params.TopNConcentrationLimitBps {
		dampening = clampInt64((topPower-params.TopNConcentrationLimitBps)*params.GroupMaxRewardDampeningBps/(BasisPoints-params.TopNConcentrationLimitBps), 0, params.GroupMaxRewardDampeningBps)
	}
	rewardBefore := ApplyBps(normalizeInt(input.RewardPoolNaet), topPower)
	rewardAfter := ApplyBps(rewardBefore, BasisPoints-dampening)
	commissionAlerts := synchronizedCommissionAlerts(validators, params)
	downtimeAlerts := correlatedDowntimeAlerts(validators, params)
	correlationReports := make([]CorrelationTelemetryReport, 0)
	byGroup := validatorsByGroup(validators)
	for groupID, groupValidators := range byGroup {
		report, err := EvaluateCorrelationTelemetry(groupID, groupValidators, params)
		if err != nil {
			return CollusionScenarioReport{}, err
		}
		if len(report.Signals) > 0 || report.ConsensusAffecting {
			correlationReports = append(correlationReports, report)
		}
	}
	sort.SliceStable(correlationReports, func(i, j int) bool {
		return correlationReports[i].GroupID < correlationReports[j].GroupID
	})
	return CollusionScenarioReport{
		TopNVotingPowerBps:		topPower,
		TopNLimitBps:			params.TopNConcentrationLimitBps,
		ConcentrationExceeded:		topPower > params.TopNConcentrationLimitBps,
		ConcentrationDampeningBps:	dampening,
		RewardBeforeDampeningNaet:	rewardBefore,
		RewardAfterDampeningNaet:	rewardAfter,
		CommissionAlerts:		commissionAlerts,
		DowntimeAlerts:			downtimeAlerts,
		CorrelationReports:		correlationReports,
		GovernanceSummary:		fmt.Sprintf("top_n_power_bps=%d limit_bps=%d dampening_bps=%d", topPower, params.TopNConcentrationLimitBps, dampening),
	}, nil
}

func RouteConsensusFaultEvidence(input EvidenceRoutingInput) (EvidenceRoutingReport, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return EvidenceRoutingReport{}, err
	}
	if input.ValidatorID == "" {
		return EvidenceRoutingReport{}, fmt.Errorf("validator_id is required")
	}
	reward := minInt(normalizeInt(input.RequestedReporterRewardNaet), normalizeInt(params.EvidenceRoutingRewardCapNaet))
	if input.Duplicate {
		return EvidenceRoutingReport{Route: EvidenceRouteDuplicate, Accepted: false, ReporterRewardNaet: sdkmath.ZeroInt(), RejectedReason: "duplicate_evidence", Auditable: true}, nil
	}
	if !input.Accepted {
		return EvidenceRoutingReport{Route: evidenceRoute(input.EvidenceType), Accepted: false, ReporterRewardNaet: sdkmath.ZeroInt(), RejectedReason: "evidence_not_accepted", Auditable: true}, nil
	}
	return EvidenceRoutingReport{Route: evidenceRoute(input.EvidenceType), Accepted: true, ReporterRewardNaet: reward, Auditable: true}, nil
}

func (p SybilCollusionResistanceParams) Validate() error {
	for _, item := range []struct {
		name	string
		value	int64
	}{
		{name: "group_concentration_soft_cap_bps", value: p.GroupConcentrationSoftCapBps},
		{name: "group_max_reward_dampening_bps", value: p.GroupMaxRewardDampeningBps},
		{name: "deterministic_correlation_threshold_bps", value: p.DeterministicCorrelationThresholdBps},
		{name: "advisory_correlation_threshold_bps", value: p.AdvisoryCorrelationThresholdBps},
		{name: "min_self_delegation_ratio_bps", value: p.MinSelfDelegationRatioBps},
		{name: "min_performance_bps", value: p.MinPerformanceBps},
		{name: "bootstrap_max_stake_bps", value: p.BootstrapMaxStakeBps},
		{name: "bootstrap_bonus_bps", value: p.BootstrapBonusBps},
		{name: "top_n_concentration_limit_bps", value: p.TopNConcentrationLimitBps},
		{name: "synchronized_commission_threshold_bps", value: p.SynchronizedCommissionThresholdBps},
		{name: "correlated_downtime_threshold_bps", value: p.CorrelatedDowntimeThresholdBps},
	} {
		if err := validateBps(item.name, item.value, 0, BasisPoints); err != nil {
			return err
		}
	}
	if normalizeInt(p.MinSelfDelegationNaet).IsNegative() {
		return fmt.Errorf("min_self_delegation_naet must not be negative")
	}
	if normalizeInt(p.EvidenceRoutingRewardCapNaet).IsNegative() {
		return fmt.Errorf("evidence_routing_reward_cap_naet must not be negative")
	}
	if p.BootstrapEligibilityWindowEpochs == 0 {
		return fmt.Errorf("bootstrap_eligibility_window_epochs must be positive")
	}
	return nil
}

func (p SybilCollusionResistanceParams) withDefaults() SybilCollusionResistanceParams {
	defaults := DefaultSybilCollusionResistanceParams()
	if p.GroupConcentrationSoftCapBps == 0 {
		p.GroupConcentrationSoftCapBps = defaults.GroupConcentrationSoftCapBps
	}
	if p.GroupMaxRewardDampeningBps == 0 {
		p.GroupMaxRewardDampeningBps = defaults.GroupMaxRewardDampeningBps
	}
	if p.DeterministicCorrelationThresholdBps == 0 {
		p.DeterministicCorrelationThresholdBps = defaults.DeterministicCorrelationThresholdBps
	}
	if p.AdvisoryCorrelationThresholdBps == 0 {
		p.AdvisoryCorrelationThresholdBps = defaults.AdvisoryCorrelationThresholdBps
	}
	if p.MinSelfDelegationNaet.IsNil() {
		p.MinSelfDelegationNaet = defaults.MinSelfDelegationNaet
	}
	if p.MinSelfDelegationRatioBps == 0 {
		p.MinSelfDelegationRatioBps = defaults.MinSelfDelegationRatioBps
	}
	if p.MinPerformanceBps == 0 {
		p.MinPerformanceBps = defaults.MinPerformanceBps
	}
	if p.BootstrapEligibilityWindowEpochs == 0 {
		p.BootstrapEligibilityWindowEpochs = defaults.BootstrapEligibilityWindowEpochs
	}
	if p.BootstrapMaxStakeBps == 0 {
		p.BootstrapMaxStakeBps = defaults.BootstrapMaxStakeBps
	}
	if p.BootstrapBonusBps == 0 {
		p.BootstrapBonusBps = defaults.BootstrapBonusBps
	}
	if p.TopNConcentrationLimitBps == 0 {
		p.TopNConcentrationLimitBps = defaults.TopNConcentrationLimitBps
	}
	if p.SynchronizedCommissionThresholdBps == 0 {
		p.SynchronizedCommissionThresholdBps = defaults.SynchronizedCommissionThresholdBps
	}
	if p.CorrelatedDowntimeThresholdBps == 0 {
		p.CorrelatedDowntimeThresholdBps = defaults.CorrelatedDowntimeThresholdBps
	}
	if p.EvidenceRoutingRewardCapNaet.IsNil() {
		p.EvidenceRoutingRewardCapNaet = defaults.EvidenceRoutingRewardCapNaet
	}
	return p
}

func (v ValidatorSybilRecord) Validate() error {
	if v.ValidatorID == "" {
		return fmt.Errorf("validator_id is required")
	}
	if normalizeInt(v.StakeNaet).IsNegative() {
		return fmt.Errorf("stake_naet must not be negative")
	}
	if normalizeInt(v.SelfDelegationNaet).IsNegative() {
		return fmt.Errorf("self_delegation_naet must not be negative")
	}
	for _, item := range []struct {
		name	string
		value	int64
	}{
		{name: "voting_power_bps", value: v.VotingPowerBps},
		{name: "self_delegation_ratio_bps", value: v.SelfDelegationRatioBps},
		{name: "performance_bps", value: v.PerformanceBps},
		{name: "commission_bps", value: v.CommissionBps},
		{name: "commission_change_bps", value: v.CommissionChangeBps},
		{name: "correlation_score_bps", value: v.CorrelationScoreBps},
	} {
		if err := validateBps(item.name, item.value, 0, BasisPoints); err != nil {
			return err
		}
	}
	return nil
}

func groupDampening(groupPowerBps, correlationBps int64, deterministic bool, params SybilCollusionResistanceParams) (int64, bool, bool) {
	advisory := correlationBps >= params.AdvisoryCorrelationThresholdBps
	consensus := deterministic && correlationBps >= params.DeterministicCorrelationThresholdBps
	if !consensus || groupPowerBps <= params.GroupConcentrationSoftCapBps {
		return 0, consensus, advisory
	}
	denom := BasisPoints - params.GroupConcentrationSoftCapBps
	return clampInt64((groupPowerBps-params.GroupConcentrationSoftCapBps)*params.GroupMaxRewardDampeningBps/denom, 0, params.GroupMaxRewardDampeningBps), true, advisory
}

func bootstrapBps(eligible bool, params SybilCollusionResistanceParams) int64 {
	if !eligible {
		return 0
	}
	return params.BootstrapBonusBps
}

func validatorsByGroup(validators []ValidatorSybilRecord) map[string][]ValidatorSybilRecord {
	out := make(map[string][]ValidatorSybilRecord)
	for _, validator := range validators {
		groupID := validator.EconomicGroupID
		if groupID == "" {
			groupID = validator.ValidatorID
		}
		out[groupID] = append(out[groupID], validator)
	}
	return out
}

func synchronizedCommissionAlerts(validators []ValidatorSybilRecord, params SybilCollusionResistanceParams) []string {
	count := 0
	for _, validator := range validators {
		if validator.CommissionChangeBps >= params.SynchronizedCommissionThresholdBps {
			count++
		}
	}
	if count >= 2 {
		return []string{CorrelationSignalCommission}
	}
	return nil
}

func correlatedDowntimeAlerts(validators []ValidatorSybilRecord, params SybilCollusionResistanceParams) []string {
	count := 0
	for _, validator := range validators {
		if validator.DowntimeObserved {
			count++
		}
	}
	if len(validators) > 0 && int64(count)*BasisPoints/int64(len(validators)) >= params.CorrelatedDowntimeThresholdBps {
		return []string{CorrelationSignalDowntime}
	}
	return nil
}

func evidenceRoute(evidenceType string) string {
	switch evidenceType {
	case "equivocation":
		return EvidenceRouteEquivocation
	case "downtime":
		return EvidenceRouteDowntime
	default:
		return "generic_consensus_fault_evidence_route"
	}
}
