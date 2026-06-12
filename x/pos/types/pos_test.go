package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestTargetActiveValidatorsAdaptsToLoad(t *testing.T) {
	params := DefaultParams()

	require.Equal(t, uint32(75), TargetActiveValidators(params, 0))
	require.Equal(t, uint32(400), TargetActiveValidators(params, BasisPoints))
	require.Equal(t, uint32(238), TargetActiveValidators(params, 5_000))
}

func TestSelectActiveValidatorsRanksByStakePerformanceAndUptime(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MaxVotingPowerBps = 0
	candidates := makeCandidates(80, 1_000)
	candidates[10].SelfStakeNaet = sdkmath.NewInt(10_000)
	candidates[10].PerformanceScoreBps = 9_000
	candidates[10].UptimeFactorBps = 10_000
	candidates[11].SelfStakeNaet = sdkmath.NewInt(10_000)
	candidates[11].PerformanceScoreBps = 8_000
	candidates[11].UptimeFactorBps = 10_000

	selection, err := SelectActiveValidators(params, candidates, 0)
	require.NoError(t, err)
	require.False(t, selection.InsufficientActive)
	require.Len(t, selection.Active, 75)
	require.Equal(t, "val-010", selection.Active[0].ValidatorID)
	require.Equal(t, "val-011", selection.Active[1].ValidatorID)
}

func TestSelectActiveValidatorsRejectsIneligibleCandidates(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(1_000)

	candidates := []Candidate{
		candidate("low-stake", 1, 0),
		candidate("bad-uptime", 1_000, 0),
		candidate("bad-commission", 1_000, 0),
		candidate("jailed", 1_000, 0),
	}
	candidates[1].UptimeFactorBps = params.MinUptimeBps - 1
	candidates[2].CommissionBps = params.MaxCommissionBps + 1
	candidates[3].Jailed = true

	selection, err := SelectActiveValidators(params, candidates, 0)
	require.NoError(t, err)
	require.Empty(t, selection.Active)
	require.True(t, selection.InsufficientActive)
	require.Len(t, selection.Rejected, 4)
	require.Contains(t, selection.Rejected[0].Reason, "below election minimum")
	require.Contains(t, selection.Rejected[1].Reason, "uptime below threshold")
	require.Contains(t, selection.Rejected[2].Reason, "commission exceeds cap")
	require.Contains(t, selection.Rejected[3].Reason, "jailed validator")
}

func TestVotingPowerCapBoundsWhaleInfluence(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MaxVotingPowerBps = 1_000
	candidates := makeCandidates(75, 1_000)
	candidates[0].SelfStakeNaet = sdkmath.NewInt(1_000_000)

	selection, err := SelectActiveValidators(params, candidates, 0)
	require.NoError(t, err)
	require.Len(t, selection.Active, 75)

	totalEffective := sdkmath.ZeroInt()
	for _, validator := range selection.Active {
		totalEffective = totalEffective.Add(validator.EffectiveStakeNaet)
	}
	cap := totalEffective.MulRaw(int64(params.MaxVotingPowerBps)).QuoRaw(int64(BasisPoints))
	require.True(t, selection.Active[0].VotingPowerNaet.LTE(cap))
	require.True(t, selection.Active[0].VotingPowerCap.SoftCapped)
	require.Equal(t, sdkmath.NewInt(1_000_000), selection.Active[0].VotingPowerCap.PreCapVotingPowerNaet)
	require.Equal(t, cap, selection.Active[0].VotingPowerCap.FinalVotingPowerNaet)
	require.Equal(t, cap, selection.Active[0].VotingPowerCap.CapNaet)
	require.Contains(t, selection.Active[0].VotingPowerCap.Warning, "voting power target")
}

func TestStakeDecayReducesInactiveWeightOnly(t *testing.T) {
	params := DefaultParams()
	params.InactiveAfterEpochs = 2
	params.StakeDecayBps = 100
	stake := sdkmath.NewInt(10_000)

	require.Equal(t, stake, ApplyStakeDecay(stake, 2, params))
	require.Equal(t, sdkmath.NewInt(9_800), ApplyStakeDecay(stake, 4, params))
	require.Equal(t, sdkmath.ZeroInt(), ApplyStakeDecay(stake, 200, params))
}

func TestScoreCandidateUsesSaturatedCompositeIntegerModel(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.StakeSaturationNaet = sdkmath.NewInt(1_000)
	candidate := candidate("val-score-model", 10_000, 0)
	candidate.PerformanceScoreBps = 9_000
	candidate.UptimeFactorBps = 9_500
	candidate.LatencyFactorBps = 8_000
	candidate.ReliabilityIndexBps = 7_000

	scored, err := ScoreCandidate(params, candidate)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000), scored.EffectiveStakeNaet)
	require.Equal(t, sdkmath.NewInt(478), scored.Score)
	require.Equal(t, ValidatorScoreComponents{
		StakeWeightNaet:	sdkmath.NewInt(1_000),
		StakeSaturationCapNaet:	sdkmath.NewInt(1_000),
		SaturatedStakeNaet:	sdkmath.NewInt(9_000),
		RewardWeightNaet:	sdkmath.NewInt(3_250),
		PerformanceFactorBps:	9_000,
		UptimeFactorBps:	9_500,
		LatencyFactorBps:	8_000,
		ReliabilityIndexBps:	7_000,
		Score:			sdkmath.NewInt(478),
	}, scored.ScoreComponents)
}

func TestStakeSaturationPreviewUsesCapFactorAndPreservesBondedBalance(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(2_000)
	params.StakeSaturationCapFactorBps = 15_000
	params.StakeSaturationNaet = sdkmath.NewInt(9_999)
	params.SaturatedStakeRewardBps = 2_000
	saturatedCandidate := candidate("val-saturated", 2_000, 8_000)

	preview, err := PreviewStakeSaturation(params, saturatedCandidate)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(10_000), preview.BondedStakeNaet)
	require.Equal(t, sdkmath.NewInt(3_000), preview.SaturationCapNaet)
	require.Equal(t, sdkmath.NewInt(3_000), preview.EffectiveStakeNaet)
	require.Equal(t, sdkmath.NewInt(7_000), preview.SaturatedStakeNaet)
	require.Equal(t, sdkmath.NewInt(4_400), preview.RewardWeightNaet)
	require.True(t, preview.Saturated)
	require.Contains(t, preview.Warning, "excess stake")

	scored, err := ScoreCandidate(params, saturatedCandidate)
	require.NoError(t, err)
	require.Equal(t, preview.BondedStakeNaet, scored.TotalStakeNaet)
	require.Equal(t, preview.EffectiveStakeNaet, scored.EffectiveStakeNaet)
	require.Equal(t, preview.RewardWeightNaet, scored.ScoreComponents.RewardWeightNaet)

	pending, err := PreviewDelegationSaturation(params, candidate("val-saturated", 2_000, 0), sdkmath.NewInt(8_000))
	require.NoError(t, err)
	require.Equal(t, preview, pending)

	_, err = PreviewDelegationSaturation(params, saturatedCandidate, sdkmath.NewInt(-1))
	require.ErrorContains(t, err, "additional delegation")
}

func TestPerformanceScoreUsesUptimeLatencyAndCorrectness(t *testing.T) {
	params := DefaultParams()
	score, err := ComputePerformanceScore(params.PerformanceWeights, PerformanceSignals{
		UptimeBps:	10_000,
		LatencyBps:	8_000,
		CorrectnessBps:	9_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(9_100), score)

	_, err = ComputePerformanceScore(PerformanceWeights{UptimeWeightBps: 1}, PerformanceSignals{})
	require.ErrorContains(t, err, "performance weights")
}

func TestEpochNumberUsesConfigurableEpochDuration(t *testing.T) {
	params := DefaultParams()
	params.EpochDurationSeconds = 43_200

	epoch, err := EpochNumber(params, 86_399)
	require.NoError(t, err)
	require.Equal(t, uint64(1), epoch)

	epoch, err = EpochNumber(params, 86_400)
	require.NoError(t, err)
	require.Equal(t, uint64(2), epoch)
}

func TestDistributeRewardsPaysCommissionThenStakeShares(t *testing.T) {
	result, err := DistributeRewards(RewardInput{
		ValidatorID:		"val-a",
		TotalRewardsNaet:	sdkmath.NewInt(1_000),
		CommissionBps:		1_000,
		SelfStakeNaet:		sdkmath.NewInt(100),
		Nominations: []Nomination{
			{NominatorID: "nom-b", StakeNaet: sdkmath.NewInt(300)},
			{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(600)},
		},
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(100), result.ValidatorCommissionNaet)
	require.Equal(t, sdkmath.NewInt(90), result.ValidatorSelfShareNaet)
	require.Equal(t, []NominatorReward{
		{NominatorID: "nom-a", RewardNaet: sdkmath.NewInt(540)},
		{NominatorID: "nom-b", RewardNaet: sdkmath.NewInt(270)},
	}, result.NominatorRewards)
	require.True(t, result.RemainderNaet.IsZero())
	require.True(t, result.TotalDistributedNaet.Equal(sdkmath.NewInt(1_000)))
}

func TestComputeSlashSharesPenaltyWithNominators(t *testing.T) {
	result, err := ComputeSlash(SlashInput{
		ValidatorID:		"val-a",
		Misbehavior:		MisbehaviorDoubleSign,
		SlashFractionBps:	500,
		SelfStakeNaet:		sdkmath.NewInt(1_000),
		EvidenceHeight:		12,
		EvidenceFinalized:	true,
		Nominations: []Nomination{
			{NominatorID: "nom-b", StakeNaet: sdkmath.NewInt(4_000)},
			{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(2_000)},
		},
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(50), result.SelfSlashedNaet)
	require.Equal(t, []NominatorSlash{
		{NominatorID: "nom-a", SlashedNaet: sdkmath.NewInt(100)},
		{NominatorID: "nom-b", SlashedNaet: sdkmath.NewInt(200)},
	}, result.NominatorSlashes)
	require.Equal(t, sdkmath.NewInt(350), result.TotalSlashedNaet)
}

func TestComputeSlashRequiresObjectiveFinalizedEvidence(t *testing.T) {
	_, err := ComputeSlash(SlashInput{
		ValidatorID:		"val-a",
		Misbehavior:		MisbehaviorInvalidBlock,
		SlashFractionBps:	100,
		SelfStakeNaet:		sdkmath.NewInt(1_000),
	})
	require.ErrorContains(t, err, "evidence must be finalized")

	_, err = ComputeSlash(SlashInput{
		ValidatorID:		"val-a",
		Misbehavior:		"subjective",
		SlashFractionBps:	100,
		SelfStakeNaet:		sdkmath.NewInt(1_000),
		EvidenceFinalized:	true,
	})
	require.ErrorContains(t, err, "unsupported misbehavior")
}

func TestParamsEnforceProductionBounds(t *testing.T) {
	params := DefaultParams()
	params.MinActiveValidators = 74
	require.ErrorContains(t, params.Validate(), "at least 75")

	params = DefaultParams()
	params.MaxActiveValidators = 401
	require.ErrorContains(t, params.Validate(), "must not exceed 400")

	params = DefaultParams()
	params.EpochDurationSeconds = 10
	require.ErrorContains(t, params.Validate(), "epoch duration")

	params = DefaultParams()
	params.MaxCommissionBps = 2_001
	require.ErrorContains(t, params.Validate(), "20%")

	params = DefaultParams()
	params.UnbondingSeconds = MinUnbondingSeconds - 1
	require.ErrorContains(t, params.Validate(), "unbonding period")

	params = DefaultParams()
	params.TargetCommitMillis = MaxTargetCommitMillis + 1
	require.ErrorContains(t, params.Validate(), "target commit latency")

	params = DefaultParams()
	params.StakeSaturationCapFactorBps = 0
	require.ErrorContains(t, params.Validate(), "cap factor")

	params = DefaultParams()
	params.SaturatedStakeRewardBps = BasisPoints + 1
	require.ErrorContains(t, params.Validate(), "saturated stake reward")
}

func makeCandidates(count int, stake int64) []Candidate {
	candidates := make([]Candidate, count)
	for i := range candidates {
		candidates[i] = candidate(fmt.Sprintf("val-%03d", i), stake, 0)
	}
	return candidates
}

func candidate(id string, selfStake int64, delegatedStake int64) Candidate {
	return Candidate{
		ValidatorID:		id,
		SelfStakeNaet:		sdkmath.NewInt(selfStake),
		DelegatedStakeNaet:	sdkmath.NewInt(delegatedStake),
		PerformanceScoreBps:	BasisPoints,
		UptimeFactorBps:	BasisPoints,
		CommissionBps:		500,
		Nominations:		nominations(delegatedStake),
	}
}

func nominations(delegatedStake int64) []Nomination {
	if delegatedStake == 0 {
		return nil
	}
	return []Nomination{{NominatorID: "nom-000", StakeNaet: sdkmath.NewInt(delegatedStake)}}
}
