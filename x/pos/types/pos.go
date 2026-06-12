package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	BasisPoints	= uint32(10_000)

	DefaultMinActiveValidators	= uint32(75)
	DefaultMaxActiveValidators	= uint32(400)

	DefaultEpochDurationSeconds	= uint64(43_200)
	MinEpochDurationSeconds		= uint64(43_200)
	MaxEpochDurationSeconds		= uint64(86_400)

	DefaultMaxCommissionBps			= uint32(2_000)
	DefaultMaxVotingPowerBps		= uint32(1_500)
	DefaultMinUptimeBps			= uint32(9_500)
	DefaultInactiveAfterEpochs		= uint64(2)
	DefaultStakeDecayBps			= uint32(100)
	DefaultStakeSaturationThresholdNaet	= int64(10_000_000_000)
	DefaultStakeSaturationCapFactorBps	= uint32(100_000)
	MaxStakeSaturationCapFactorBps		= uint32(1_000_000)
	DefaultStakeSaturationNaet		= int64(100_000_000_000)
	DefaultSaturatedStakeRewardBps		= uint32(2_500)
	DefaultUnbondingSeconds			= uint64(1_209_600)
	MinUnbondingSeconds			= uint64(604_800)
	MaxUnbondingSeconds			= uint64(1_814_400)
	DefaultTargetCommitMillis		= uint32(1_500)
	MaxTargetCommitMillis			= uint32(2_000)

	DefaultMaxValidatorSetChangeRateBps	= uint32(1_000)

	DefaultPerformanceUptimeWeightBps	= uint32(4_000)
	DefaultPerformanceLatencyWeightBps	= uint32(3_000)
	DefaultPerformanceCorrectnessWeightBps	= uint32(3_000)

	MisbehaviorDowntime	= "downtime"
	MisbehaviorDoubleSign	= "double_sign"
	MisbehaviorInvalidBlock	= "invalid_block"
)

type Params struct {
	MinActiveValidators		uint32
	MaxActiveValidators		uint32
	EpochDurationSeconds		uint64
	PhaseDurations			EpochPhaseDurations
	EpochSeedSource			EpochSeedSource
	MaxValidatorSetChangeRateBps	uint32
	DelegationActivationEpochs	uint64
	EvidenceWindowEpochs		uint64
	MinStakeNaet			sdkmath.Int
	MaxCommissionBps		uint32
	MaxVotingPowerBps		uint32
	MinUptimeBps			uint32
	InactiveAfterEpochs		uint64
	StakeDecayBps			uint32
	StakeSaturationThresholdNaet	sdkmath.Int
	StakeSaturationCapFactorBps	uint32
	StakeSaturationNaet		sdkmath.Int
	SaturatedStakeRewardBps		uint32
	UnbondingSeconds		uint64
	TargetCommitMillis		uint32
	MinTaskGroupValidators		uint32
	MaxTaskGroupValidators		uint32
	ReporterRewardBps		uint32
	PerformanceWeights		PerformanceWeights
}

type PerformanceSignals struct {
	UptimeBps	uint32
	LatencyBps	uint32
	CorrectnessBps	uint32
}

type PerformanceWeights struct {
	UptimeWeightBps		uint32
	LatencyWeightBps	uint32
	CorrectnessWeightBps	uint32
}

type Candidate struct {
	ValidatorID		string
	SelfStakeNaet		sdkmath.Int
	DelegatedStakeNaet	sdkmath.Int
	PerformanceScoreBps	uint32
	UptimeFactorBps		uint32
	LatencyFactorBps	uint32
	ReliabilityIndexBps	uint32
	CommissionBps		uint32
	InactiveEpochs		uint64
	Jailed			bool
	Tombstoned		bool
	Roles			[]ValidatorRole
	Capacity		ValidatorCapacity
	Nominations		[]Nomination
}

type Nomination struct {
	NominatorID	string
	StakeNaet	sdkmath.Int
}

type ValidatorScoreComponents struct {
	StakeWeightNaet		sdkmath.Int
	StakeSaturationCapNaet	sdkmath.Int
	SaturatedStakeNaet	sdkmath.Int
	RewardWeightNaet	sdkmath.Int
	PerformanceFactorBps	uint32
	UptimeFactorBps		uint32
	LatencyFactorBps	uint32
	ReliabilityIndexBps	uint32
	Score			sdkmath.Int
}

type ScoredValidator struct {
	Candidate
	TotalStakeNaet		sdkmath.Int
	EffectiveStakeNaet	sdkmath.Int
	Score			sdkmath.Int
	VotingPowerNaet		sdkmath.Int
	VotingPowerCap		VotingPowerCapStatus
	ScoreComponents		ValidatorScoreComponents
}

type StakeSaturationPreview struct {
	ValidatorID		string
	BondedStakeNaet		sdkmath.Int
	SaturationThresholdNaet	sdkmath.Int
	CapFactorBps		uint32
	SaturationCapNaet	sdkmath.Int
	EffectiveStakeNaet	sdkmath.Int
	SaturatedStakeNaet	sdkmath.Int
	RewardWeightNaet	sdkmath.Int
	SaturatedStakeRewardBps	uint32
	Saturated		bool
	Warning			string
}

type VotingPowerCapStatus struct {
	CapNaet			sdkmath.Int
	PreCapVotingPowerNaet	sdkmath.Int
	FinalVotingPowerNaet	sdkmath.Int
	ExcessVotingPowerNaet	sdkmath.Int
	MaxVotingPowerBps	uint32
	SoftCapped		bool
	Warning			string
}

type Selection struct {
	Active			[]ScoredValidator
	Rejected		[]RejectedCandidate
	TargetActive		uint32
	InsufficientActive	bool
}

type RejectedCandidate struct {
	Candidate	Candidate
	Reason		string
}

type RewardInput struct {
	ValidatorID		string
	TotalRewardsNaet	sdkmath.Int
	CommissionBps		uint32
	SelfStakeNaet		sdkmath.Int
	Nominations		[]Nomination
}

type RewardDistribution struct {
	ValidatorID		string
	ValidatorCommissionNaet	sdkmath.Int
	ValidatorSelfShareNaet	sdkmath.Int
	NominatorRewards	[]NominatorReward
	RemainderNaet		sdkmath.Int
	TotalDistributedNaet	sdkmath.Int
}

type NominatorReward struct {
	NominatorID	string
	RewardNaet	sdkmath.Int
}

type SlashInput struct {
	ValidatorID		string
	Misbehavior		string
	SlashFractionBps	uint32
	SelfStakeNaet		sdkmath.Int
	Nominations		[]Nomination
	EvidenceHeight		int64
	EvidenceFinalized	bool
}

type SlashDistribution struct {
	ValidatorID		string
	Misbehavior		string
	SelfSlashedNaet		sdkmath.Int
	NominatorSlashes	[]NominatorSlash
	TotalSlashedNaet	sdkmath.Int
	EvidenceHeight		int64
	EvidenceFinalized	bool
}

type NominatorSlash struct {
	NominatorID	string
	SlashedNaet	sdkmath.Int
}

func DefaultParams() Params {
	return Params{
		MinActiveValidators:		DefaultMinActiveValidators,
		MaxActiveValidators:		DefaultMaxActiveValidators,
		EpochDurationSeconds:		DefaultEpochDurationSeconds,
		PhaseDurations:			DefaultEpochPhaseDurations(DefaultEpochDurationSeconds),
		EpochSeedSource:		EpochSeedSourcePreviousSeedValidatorSet,
		MaxValidatorSetChangeRateBps:	DefaultMaxValidatorSetChangeRateBps,
		DelegationActivationEpochs:	DefaultDelegationActivationEpochs,
		EvidenceWindowEpochs:		DefaultEvidenceWindowEpochs,
		MinStakeNaet:			sdkmath.NewInt(1_000_000_000),
		MaxCommissionBps:		DefaultMaxCommissionBps,
		MaxVotingPowerBps:		DefaultMaxVotingPowerBps,
		MinUptimeBps:			DefaultMinUptimeBps,
		InactiveAfterEpochs:		DefaultInactiveAfterEpochs,
		StakeDecayBps:			DefaultStakeDecayBps,
		StakeSaturationThresholdNaet:	sdkmath.NewInt(DefaultStakeSaturationThresholdNaet),
		StakeSaturationCapFactorBps:	DefaultStakeSaturationCapFactorBps,
		StakeSaturationNaet:		sdkmath.NewInt(DefaultStakeSaturationNaet),
		SaturatedStakeRewardBps:	DefaultSaturatedStakeRewardBps,
		UnbondingSeconds:		DefaultUnbondingSeconds,
		TargetCommitMillis:		DefaultTargetCommitMillis,
		MinTaskGroupValidators:		DefaultMinTaskGroupValidators,
		MaxTaskGroupValidators:		DefaultMaxTaskGroupValidators,
		ReporterRewardBps:		DefaultReporterRewardBps,
		PerformanceWeights: PerformanceWeights{
			UptimeWeightBps:	DefaultPerformanceUptimeWeightBps,
			LatencyWeightBps:	DefaultPerformanceLatencyWeightBps,
			CorrectnessWeightBps:	DefaultPerformanceCorrectnessWeightBps,
		},
	}
}

func (p Params) Validate() error {
	if p.MinActiveValidators == 0 {
		return errors.New("minimum active validators must be positive")
	}
	if p.MinActiveValidators > p.MaxActiveValidators {
		return errors.New("minimum active validators cannot exceed maximum active validators")
	}
	if p.MinActiveValidators < DefaultMinActiveValidators {
		return fmt.Errorf("minimum active validators must be at least %d", DefaultMinActiveValidators)
	}
	if p.MaxActiveValidators > DefaultMaxActiveValidators {
		return fmt.Errorf("maximum active validators must not exceed %d", DefaultMaxActiveValidators)
	}
	if p.EpochDurationSeconds < MinEpochDurationSeconds || p.EpochDurationSeconds > MaxEpochDurationSeconds {
		return fmt.Errorf("epoch duration must be within %d..%d seconds", MinEpochDurationSeconds, MaxEpochDurationSeconds)
	}
	if err := p.EffectivePhaseDurations().Validate(p.EpochDurationSeconds); err != nil {
		return err
	}
	if err := p.EffectiveEpochSeedSource().Validate(); err != nil {
		return err
	}
	if p.MaxValidatorSetChangeRateBps == 0 || p.MaxValidatorSetChangeRateBps > BasisPoints {
		return fmt.Errorf("max validator set change rate must be within 1..%d bps", BasisPoints)
	}
	if p.DelegationActivationEpochs == 0 {
		return errors.New("delegation activation delay must be positive")
	}
	if p.EvidenceWindowEpochs == 0 {
		return errors.New("evidence slashable window must be positive")
	}
	if !p.MinStakeNaet.IsPositive() {
		return errors.New("minimum validator stake must be positive")
	}
	if p.MaxCommissionBps > 2_000 {
		return errors.New("validator commission cap cannot exceed 20%")
	}
	if p.MaxVotingPowerBps > BasisPoints {
		return fmt.Errorf("max voting power cap must be <= %d bps", BasisPoints)
	}
	if p.MinUptimeBps > BasisPoints {
		return fmt.Errorf("minimum uptime must be <= %d bps", BasisPoints)
	}
	if p.StakeDecayBps > BasisPoints {
		return fmt.Errorf("stake decay must be <= %d bps", BasisPoints)
	}
	if !p.StakeSaturationThresholdNaet.IsPositive() {
		return errors.New("stake saturation threshold must be positive")
	}
	if p.StakeSaturationCapFactorBps == 0 || p.StakeSaturationCapFactorBps > MaxStakeSaturationCapFactorBps {
		return fmt.Errorf("stake saturation cap factor must be within 1..%d bps", MaxStakeSaturationCapFactorBps)
	}
	if !p.StakeSaturationNaet.IsPositive() {
		return errors.New("stake saturation must be positive")
	}
	if p.SaturatedStakeRewardBps > BasisPoints {
		return fmt.Errorf("saturated stake reward weight must be <= %d bps", BasisPoints)
	}
	if p.UnbondingSeconds < MinUnbondingSeconds || p.UnbondingSeconds > MaxUnbondingSeconds {
		return fmt.Errorf("unbonding period must be within %d..%d seconds", MinUnbondingSeconds, MaxUnbondingSeconds)
	}
	if p.TargetCommitMillis == 0 || p.TargetCommitMillis > MaxTargetCommitMillis {
		return fmt.Errorf("target commit latency must be within 1..%d milliseconds", MaxTargetCommitMillis)
	}
	if p.MinTaskGroupValidators == 0 {
		return errors.New("minimum task group validators must be positive")
	}
	if p.MinTaskGroupValidators > p.MaxTaskGroupValidators {
		return errors.New("minimum task group validators cannot exceed maximum task group validators")
	}
	if p.MaxTaskGroupValidators > p.MaxActiveValidators {
		return errors.New("maximum task group validators cannot exceed maximum active validators")
	}
	if p.ReporterRewardBps > MaxReporterRewardBps {
		return fmt.Errorf("reporter reward must be <= %d bps", MaxReporterRewardBps)
	}
	if err := p.PerformanceWeights.Validate(); err != nil {
		return err
	}
	return nil
}

func (w PerformanceWeights) Validate() error {
	total := uint64(w.UptimeWeightBps) + uint64(w.LatencyWeightBps) + uint64(w.CorrectnessWeightBps)
	if total != uint64(BasisPoints) {
		return fmt.Errorf("performance weights must sum to %d bps", BasisPoints)
	}
	return nil
}

func (s PerformanceSignals) Validate() error {
	if s.UptimeBps > BasisPoints {
		return fmt.Errorf("performance uptime must be <= %d bps", BasisPoints)
	}
	if s.LatencyBps > BasisPoints {
		return fmt.Errorf("performance latency must be <= %d bps", BasisPoints)
	}
	if s.CorrectnessBps > BasisPoints {
		return fmt.Errorf("performance correctness must be <= %d bps", BasisPoints)
	}
	return nil
}

func ComputePerformanceScore(weights PerformanceWeights, signals PerformanceSignals) (uint32, error) {
	if err := weights.Validate(); err != nil {
		return 0, err
	}
	if err := signals.Validate(); err != nil {
		return 0, err
	}
	score := uint64(weights.UptimeWeightBps)*uint64(signals.UptimeBps) +
		uint64(weights.LatencyWeightBps)*uint64(signals.LatencyBps) +
		uint64(weights.CorrectnessWeightBps)*uint64(signals.CorrectnessBps)
	return uint32(score / uint64(BasisPoints)), nil
}

func (c Candidate) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(c.ValidatorID) == "" {
		return errors.New("validator id is required")
	}
	if c.SelfStakeNaet.IsNegative() {
		return errors.New("validator self-stake cannot be negative")
	}
	if c.DelegatedStakeNaet.IsNegative() {
		return errors.New("validator delegated stake cannot be negative")
	}
	if c.PerformanceScoreBps > BasisPoints {
		return fmt.Errorf("performance score must be <= %d bps", BasisPoints)
	}
	if c.UptimeFactorBps > BasisPoints {
		return fmt.Errorf("uptime factor must be <= %d bps", BasisPoints)
	}
	if c.LatencyFactorBps > BasisPoints {
		return fmt.Errorf("latency factor must be <= %d bps", BasisPoints)
	}
	if c.ReliabilityIndexBps > BasisPoints {
		return fmt.Errorf("reliability index must be <= %d bps", BasisPoints)
	}
	if c.CommissionBps > params.MaxCommissionBps {
		return fmt.Errorf("validator commission exceeds cap %d bps", params.MaxCommissionBps)
	}
	if err := validateValidatorRoles(c.Roles); err != nil {
		return err
	}
	if err := c.Capacity.Validate(); err != nil {
		return err
	}
	if err := validateNominations(c.Nominations); err != nil {
		return err
	}
	if !sumNominations(c.Nominations).Equal(c.DelegatedStakeNaet) {
		return errors.New("delegated stake must equal sum of nominations")
	}
	return nil
}

func SelectActiveValidators(params Params, candidates []Candidate, loadScoreBps uint32) (Selection, error) {
	if err := params.Validate(); err != nil {
		return Selection{}, err
	}
	if loadScoreBps > BasisPoints {
		return Selection{}, fmt.Errorf("load score must be <= %d bps", BasisPoints)
	}
	target := TargetActiveValidators(params, loadScoreBps)
	eligible := make([]ScoredValidator, 0, len(candidates))
	selection := Selection{TargetActive: target}

	for _, candidate := range candidates {
		scored, err := ScoreCandidate(params, candidate)
		if err != nil {
			selection.Rejected = append(selection.Rejected, RejectedCandidate{Candidate: cloneCandidate(candidate), Reason: err.Error()})
			continue
		}
		eligible = append(eligible, scored)
	}

	sortScoredValidators(eligible)
	limit := minUint32(target, uint32(len(eligible)))
	selection.Active = cloneScoredValidators(eligible[:limit])
	if uint32(len(selection.Active)) < params.MinActiveValidators {
		selection.InsufficientActive = true
	}
	applyVotingPowerCap(params, selection.Active)
	return selection, nil
}

func TargetActiveValidators(params Params, loadScoreBps uint32) uint32 {
	if loadScoreBps > BasisPoints {
		loadScoreBps = BasisPoints
	}
	if params.MaxActiveValidators <= params.MinActiveValidators {
		return params.MinActiveValidators
	}
	spread := params.MaxActiveValidators - params.MinActiveValidators
	return params.MinActiveValidators + uint32((uint64(spread)*uint64(loadScoreBps)+uint64(BasisPoints)-1)/uint64(BasisPoints))
}

func EpochNumber(params Params, unixSeconds uint64) (uint64, error) {
	if err := params.Validate(); err != nil {
		return 0, err
	}
	return unixSeconds / params.EpochDurationSeconds, nil
}

func ScoreCandidate(params Params, candidate Candidate) (ScoredValidator, error) {
	if err := candidate.Validate(params); err != nil {
		return ScoredValidator{}, err
	}
	totalStake := candidate.SelfStakeNaet.Add(candidate.DelegatedStakeNaet)
	if totalStake.LT(params.MinStakeNaet) {
		return ScoredValidator{}, errors.New("validator stake below election minimum")
	}
	if candidate.Jailed {
		return ScoredValidator{}, errors.New("jailed validator is not eligible")
	}
	if candidate.Tombstoned {
		return ScoredValidator{}, errors.New("tombstoned validator is not eligible")
	}
	if candidate.UptimeFactorBps < params.MinUptimeBps {
		return ScoredValidator{}, errors.New("validator uptime below threshold")
	}
	components := ComputeValidatorScoreComponents(params, totalStake, candidate)
	return ScoredValidator{
		Candidate:		cloneCandidate(candidate),
		TotalStakeNaet:		totalStake,
		EffectiveStakeNaet:	components.StakeWeightNaet,
		Score:			components.Score,
		VotingPowerNaet:	components.StakeWeightNaet,
		ScoreComponents:	components,
	}, nil
}

func ComputeValidatorScoreComponents(params Params, totalStake sdkmath.Int, candidate Candidate) ValidatorScoreComponents {
	decayedStake := ApplyStakeDecay(totalStake, candidate.InactiveEpochs, params)
	stakeWeight := ApplyStakeSaturation(decayedStake, params)
	saturationCap := StakeSaturationCap(params)
	saturatedStake := SaturatedStake(decayedStake, params)
	score := mulIntBps(stakeWeight, candidate.PerformanceScoreBps)
	score = mulIntBps(score, candidate.UptimeFactorBps)
	latencyFactor := normalizeOptionalFactorBps(candidate.LatencyFactorBps)
	reliabilityIndex := normalizeOptionalFactorBps(candidate.ReliabilityIndexBps)
	score = mulIntBps(score, latencyFactor)
	score = mulIntBps(score, reliabilityIndex)
	return ValidatorScoreComponents{
		StakeWeightNaet:	stakeWeight,
		StakeSaturationCapNaet:	saturationCap,
		SaturatedStakeNaet:	saturatedStake,
		RewardWeightNaet:	RewardCurveStakeWeight(decayedStake, params),
		PerformanceFactorBps:	candidate.PerformanceScoreBps,
		UptimeFactorBps:	candidate.UptimeFactorBps,
		LatencyFactorBps:	latencyFactor,
		ReliabilityIndexBps:	reliabilityIndex,
		Score:			score,
	}
}

func ApplyStakeDecay(stake sdkmath.Int, inactiveEpochs uint64, params Params) sdkmath.Int {
	if !stake.IsPositive() || inactiveEpochs <= params.InactiveAfterEpochs || params.StakeDecayBps == 0 {
		return stake
	}
	decayEpochs := inactiveEpochs - params.InactiveAfterEpochs
	decayBps := decayEpochs * uint64(params.StakeDecayBps)
	if decayBps >= uint64(BasisPoints) {
		return sdkmath.ZeroInt()
	}
	return mulIntBps(stake, uint32(uint64(BasisPoints)-decayBps))
}

func ApplyStakeSaturation(stake sdkmath.Int, params Params) sdkmath.Int {
	if !stake.IsPositive() {
		return sdkmath.ZeroInt()
	}
	cap := StakeSaturationCap(params)
	if cap.IsPositive() && stake.GT(cap) {
		return cap
	}
	return stake
}

func StakeSaturationCap(params Params) sdkmath.Int {
	derivedCap := mulIntBps(params.StakeSaturationThresholdNaet, params.StakeSaturationCapFactorBps)
	if !derivedCap.IsPositive() {
		return params.StakeSaturationNaet
	}
	if params.StakeSaturationNaet.IsPositive() && params.StakeSaturationNaet.LT(derivedCap) {
		return params.StakeSaturationNaet
	}
	return derivedCap
}

func SaturatedStake(stake sdkmath.Int, params Params) sdkmath.Int {
	if !stake.IsPositive() {
		return sdkmath.ZeroInt()
	}
	cap := StakeSaturationCap(params)
	if !cap.IsPositive() || stake.LTE(cap) {
		return sdkmath.ZeroInt()
	}
	return stake.Sub(cap)
}

func RewardCurveStakeWeight(stake sdkmath.Int, params Params) sdkmath.Int {
	if !stake.IsPositive() {
		return sdkmath.ZeroInt()
	}
	cap := StakeSaturationCap(params)
	if !cap.IsPositive() || stake.LTE(cap) {
		return stake
	}
	excessWeight := mulIntBps(stake.Sub(cap), params.SaturatedStakeRewardBps)
	return cap.Add(excessWeight)
}

func PreviewStakeSaturation(params Params, candidate Candidate) (StakeSaturationPreview, error) {
	if err := candidate.Validate(params); err != nil {
		return StakeSaturationPreview{}, err
	}
	bondedStake := candidate.SelfStakeNaet.Add(candidate.DelegatedStakeNaet)
	return buildStakeSaturationPreview(params, strings.TrimSpace(candidate.ValidatorID), bondedStake), nil
}

func PreviewDelegationSaturation(params Params, candidate Candidate, additionalStake sdkmath.Int) (StakeSaturationPreview, error) {
	if additionalStake.IsNegative() {
		return StakeSaturationPreview{}, errors.New("additional delegation stake cannot be negative")
	}
	if err := candidate.Validate(params); err != nil {
		return StakeSaturationPreview{}, err
	}
	bondedStake := candidate.SelfStakeNaet.Add(candidate.DelegatedStakeNaet).Add(additionalStake)
	return buildStakeSaturationPreview(params, strings.TrimSpace(candidate.ValidatorID), bondedStake), nil
}

func buildStakeSaturationPreview(params Params, validatorID string, bondedStake sdkmath.Int) StakeSaturationPreview {
	cap := StakeSaturationCap(params)
	preview := StakeSaturationPreview{
		ValidatorID:			validatorID,
		BondedStakeNaet:		bondedStake,
		SaturationThresholdNaet:	params.StakeSaturationThresholdNaet,
		CapFactorBps:			params.StakeSaturationCapFactorBps,
		SaturationCapNaet:		cap,
		EffectiveStakeNaet:		ApplyStakeSaturation(bondedStake, params),
		SaturatedStakeNaet:		SaturatedStake(bondedStake, params),
		RewardWeightNaet:		RewardCurveStakeWeight(bondedStake, params),
		SaturatedStakeRewardBps:	params.SaturatedStakeRewardBps,
	}
	preview.Saturated = preview.SaturatedStakeNaet.IsPositive()
	if preview.Saturated {
		preview.Warning = "validator stake exceeds saturation cap; excess stake keeps bonded balance but receives reduced election weight"
	}
	return preview
}

func DistributeRewards(input RewardInput) (RewardDistribution, error) {
	if strings.TrimSpace(input.ValidatorID) == "" {
		return RewardDistribution{}, errors.New("validator id is required")
	}
	if input.TotalRewardsNaet.IsNegative() {
		return RewardDistribution{}, errors.New("total rewards cannot be negative")
	}
	if input.CommissionBps > DefaultMaxCommissionBps {
		return RewardDistribution{}, errors.New("validator commission cannot exceed 20%")
	}
	if input.SelfStakeNaet.IsNegative() {
		return RewardDistribution{}, errors.New("validator self-stake cannot be negative")
	}
	if err := validateNominations(input.Nominations); err != nil {
		return RewardDistribution{}, err
	}
	totalStake := input.SelfStakeNaet.Add(sumNominations(input.Nominations))
	if !totalStake.IsPositive() {
		return RewardDistribution{}, errors.New("reward distribution requires positive stake")
	}

	commission := mulIntBps(input.TotalRewardsNaet, input.CommissionBps)
	remaining := input.TotalRewardsNaet.Sub(commission)
	validatorSelfShare := shareByStake(remaining, input.SelfStakeNaet, totalStake)
	out := RewardDistribution{
		ValidatorID:			strings.TrimSpace(input.ValidatorID),
		ValidatorCommissionNaet:	commission,
		ValidatorSelfShareNaet:		validatorSelfShare,
		NominatorRewards:		make([]NominatorReward, 0, len(input.Nominations)),
	}
	distributed := commission.Add(validatorSelfShare)
	for _, nomination := range sortNominations(input.Nominations) {
		reward := shareByStake(remaining, nomination.StakeNaet, totalStake)
		out.NominatorRewards = append(out.NominatorRewards, NominatorReward{
			NominatorID:	nomination.NominatorID,
			RewardNaet:	reward,
		})
		distributed = distributed.Add(reward)
	}
	out.RemainderNaet = input.TotalRewardsNaet.Sub(distributed)
	out.TotalDistributedNaet = distributed.Add(out.RemainderNaet)
	return out, nil
}

func ComputeSlash(input SlashInput) (SlashDistribution, error) {
	if strings.TrimSpace(input.ValidatorID) == "" {
		return SlashDistribution{}, errors.New("validator id is required")
	}
	if !IsSlashableMisbehavior(input.Misbehavior) {
		return SlashDistribution{}, fmt.Errorf("unsupported misbehavior %q", input.Misbehavior)
	}
	if input.SlashFractionBps == 0 || input.SlashFractionBps > BasisPoints {
		return SlashDistribution{}, fmt.Errorf("slash fraction must be within 1..%d bps", BasisPoints)
	}
	if !input.EvidenceFinalized {
		return SlashDistribution{}, errors.New("slash evidence must be finalized")
	}
	if input.EvidenceHeight < 0 {
		return SlashDistribution{}, errors.New("slash evidence height cannot be negative")
	}
	if input.SelfStakeNaet.IsNegative() {
		return SlashDistribution{}, errors.New("validator self-stake cannot be negative")
	}
	if err := validateNominations(input.Nominations); err != nil {
		return SlashDistribution{}, err
	}
	out := SlashDistribution{
		ValidatorID:		strings.TrimSpace(input.ValidatorID),
		Misbehavior:		input.Misbehavior,
		SelfSlashedNaet:	mulIntBps(input.SelfStakeNaet, input.SlashFractionBps),
		NominatorSlashes:	make([]NominatorSlash, 0, len(input.Nominations)),
		EvidenceHeight:		input.EvidenceHeight,
		EvidenceFinalized:	input.EvidenceFinalized,
	}
	out.TotalSlashedNaet = out.SelfSlashedNaet
	for _, nomination := range sortNominations(input.Nominations) {
		slashed := mulIntBps(nomination.StakeNaet, input.SlashFractionBps)
		out.NominatorSlashes = append(out.NominatorSlashes, NominatorSlash{
			NominatorID:	nomination.NominatorID,
			SlashedNaet:	slashed,
		})
		out.TotalSlashedNaet = out.TotalSlashedNaet.Add(slashed)
	}
	return out, nil
}

func IsSlashableMisbehavior(misbehavior string) bool {
	switch misbehavior {
	case MisbehaviorDowntime, MisbehaviorDoubleSign, MisbehaviorInvalidBlock:
		return true
	default:
		return false
	}
}

func validateNominations(nominations []Nomination) error {
	seen := make(map[string]struct{}, len(nominations))
	for _, nomination := range nominations {
		id := strings.TrimSpace(nomination.NominatorID)
		if id == "" {
			return errors.New("nominator id is required")
		}
		if _, found := seen[id]; found {
			return fmt.Errorf("duplicate nominator %q", id)
		}
		seen[id] = struct{}{}
		if !nomination.StakeNaet.IsPositive() {
			return errors.New("nominator stake must be positive")
		}
	}
	return nil
}

func sumNominations(nominations []Nomination) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, nomination := range nominations {
		total = total.Add(nomination.StakeNaet)
	}
	return total
}

func sortScoredValidators(validators []ScoredValidator) {
	sort.SliceStable(validators, func(i, j int) bool {
		left := validators[i]
		right := validators[j]
		if !left.Score.Equal(right.Score) {
			return left.Score.GT(right.Score)
		}
		if !left.EffectiveStakeNaet.Equal(right.EffectiveStakeNaet) {
			return left.EffectiveStakeNaet.GT(right.EffectiveStakeNaet)
		}
		return left.ValidatorID < right.ValidatorID
	})
}

func applyVotingPowerCap(params Params, validators []ScoredValidator) {
	if params.MaxVotingPowerBps == 0 || len(validators) == 0 {
		return
	}
	total := sdkmath.ZeroInt()
	for _, validator := range validators {
		total = total.Add(validator.EffectiveStakeNaet)
	}
	if !total.IsPositive() {
		return
	}
	cap := mulIntBps(total, params.MaxVotingPowerBps)
	for i := range validators {
		status := VotingPowerCapStatus{
			CapNaet:		cap,
			PreCapVotingPowerNaet:	validators[i].VotingPowerNaet,
			FinalVotingPowerNaet:	validators[i].VotingPowerNaet,
			MaxVotingPowerBps:	params.MaxVotingPowerBps,
		}
		if validators[i].VotingPowerNaet.GT(cap) {
			status.ExcessVotingPowerNaet = validators[i].VotingPowerNaet.Sub(cap)
			status.FinalVotingPowerNaet = cap
			status.SoftCapped = true
			status.Warning = "validator exceeds voting power target; marginal stake has reduced consensus weight"
			validators[i].VotingPowerNaet = cap
		}
		validators[i].VotingPowerCap = status
	}
}

func cloneScoredValidators(validators []ScoredValidator) []ScoredValidator {
	out := make([]ScoredValidator, len(validators))
	for i, validator := range validators {
		out[i] = validator
		out[i].Candidate = cloneCandidate(validator.Candidate)
	}
	return out
}

func cloneCandidate(candidate Candidate) Candidate {
	out := candidate
	out.ValidatorID = strings.TrimSpace(candidate.ValidatorID)
	out.Roles = make([]ValidatorRole, len(candidate.Roles))
	copy(out.Roles, candidate.Roles)
	out.Capacity = cloneValidatorCapacity(candidate.Capacity)
	out.Nominations = make([]Nomination, len(candidate.Nominations))
	copy(out.Nominations, candidate.Nominations)
	return out
}

func sortNominations(nominations []Nomination) []Nomination {
	out := make([]Nomination, len(nominations))
	copy(out, nominations)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].NominatorID < out[j].NominatorID
	})
	return out
}

func shareByStake(total, stake, totalStake sdkmath.Int) sdkmath.Int {
	if !total.IsPositive() || !stake.IsPositive() || !totalStake.IsPositive() {
		return sdkmath.ZeroInt()
	}
	return total.Mul(stake).Quo(totalStake)
}

func mulIntBps(value sdkmath.Int, bps uint32) sdkmath.Int {
	if !value.IsPositive() || bps == 0 {
		return sdkmath.ZeroInt()
	}
	return value.MulRaw(int64(bps)).QuoRaw(int64(BasisPoints))
}

func normalizeOptionalFactorBps(bps uint32) uint32 {
	if bps == 0 {
		return BasisPoints
	}
	return bps
}

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
