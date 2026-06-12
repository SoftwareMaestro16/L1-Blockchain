package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestScoreValidatorEconomicsDeterministicAndBounded(t *testing.T) {
	params := DefaultStakingEnhancementParams()
	params.MaxActiveSet = 3
	totalStake := sdkmath.NewInt(10_000)

	score, err := ScoreValidatorEconomics(ValidatorEconomicRecord{
		ValidatorID:		"val-a",
		BondedStakeNaet:	sdkmath.NewInt(4_000),
		SelfDelegationNaet:	sdkmath.NewInt(400),
		CommissionBps:		500,
		UptimeBps:		9_900,
		MissedBlockRateBps:	20,
		SlashEvents:		1,
		RecentSlashSeverityBps:	100,
		MetadataComplete:	true,
	}, totalStake, params)
	require.NoError(t, err)
	require.True(t, score.Eligible)
	require.Equal(t, int64(4_000), score.VotingPowerBps)
	require.Equal(t, int64(1_000), score.SelfDelegationRatioBps)
	require.Positive(t, score.ConcentrationPenaltyBps)
	require.GreaterOrEqual(t, score.RewardAdjustmentFactorBps, params.MinRewardAdjustmentBps)
	require.LessOrEqual(t, score.RewardAdjustmentFactorBps, params.MaxRewardAdjustmentBps)

	again, err := ScoreValidatorEconomics(ValidatorEconomicRecord{
		ValidatorID:		"val-a",
		BondedStakeNaet:	sdkmath.NewInt(4_000),
		SelfDelegationNaet:	sdkmath.NewInt(400),
		CommissionBps:		500,
		UptimeBps:		9_900,
		MissedBlockRateBps:	20,
		SlashEvents:		1,
		RecentSlashSeverityBps:	100,
		MetadataComplete:	true,
	}, totalStake, params)
	require.NoError(t, err)
	require.Equal(t, score, again)
}

func TestRecommendActiveSetBoundsEpochChurnDeterministically(t *testing.T) {
	params := DefaultStakingEnhancementParams()
	params.MaxActiveSet = 3
	params.MaxEpochChurnBps = 2_500
	current := []string{"val-a", "val-b", "val-c"}
	candidates := []ValidatorEconomicRecord{
		stakingCandidate("val-e", 4_400, 440, 9_980),
		stakingCandidate("val-d", 4_500, 450, 9_990),
		stakingCandidate("val-a", 2_500, 250, 9_900),
		stakingCandidate("val-b", 2_000, 200, 9_850),
		stakingCandidate("val-c", 1_500, 150, 9_800),
	}

	first, err := RecommendActiveSetForEpoch(ActiveSetRecommendationInput{
		EpochID:			11,
		CurrentActiveValidatorIDs:	current,
		Candidates:			candidates,
		Params:				params,
	})
	require.NoError(t, err)
	require.Empty(t, first.Failed)
	require.Equal(t, uint32(1), first.AllowedChurn)
	require.Equal(t, uint32(1), first.ChurnCount)
	require.Len(t, first.Selected, 3)
	require.True(t, first.Deterministic)
	require.True(t, scoresSortedDeterministically(first.Selected))
	require.Contains(t, selectedValidatorIDs(first.Selected), "val-d")

	reversed := append([]ValidatorEconomicRecord(nil), candidates...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	second, err := RecommendActiveSetForEpoch(ActiveSetRecommendationInput{
		EpochID:			11,
		CurrentActiveValidatorIDs:	current,
		Candidates:			reversed,
		Params:				params,
	})
	require.NoError(t, err)
	require.Equal(t, selectedValidatorIDs(first.Selected), selectedValidatorIDs(second.Selected))

	invariants := ValidateStakingEnhancementInvariants(first, params)
	require.True(t, invariants.Passed, invariants.Failed)
}

func TestValidatorDistributionSimulationCoversNormalAdversarialAndLowParticipation(t *testing.T) {
	params := DefaultStakingEnhancementParams()
	params.MaxActiveSet = 3
	report, err := RunValidatorDistributionSimulation(ValidatorDistributionSimulationInput{
		EpochID:	20,
		Params:		params,
		Scenarios: []ValidatorDistributionScenario{
			{
				Name:	"normal",
				Candidates: []ValidatorEconomicRecord{
					stakingCandidate("val-a", 3_000, 300, 9_950),
					stakingCandidate("val-b", 2_800, 280, 9_930),
					stakingCandidate("val-c", 2_500, 250, 9_910),
					stakingCandidate("val-d", 1_700, 170, 9_900),
				},
			},
			{
				Name:	"adversarial_concentration",
				Candidates: []ValidatorEconomicRecord{
					stakingCandidate("val-whale", 8_000, 800, 9_990),
					stakingCandidate("val-b", 800, 80, 9_950),
					stakingCandidate("val-c", 700, 70, 9_940),
					stakingCandidate("val-d", 500, 50, 9_930),
				},
			},
			{
				Name:	"low_participation",
				Candidates: []ValidatorEconomicRecord{
					stakingCandidate("val-a", 4_000, 400, 9_950),
					{
						ValidatorID:		"val-no-metadata",
						BondedStakeNaet:	sdkmath.NewInt(3_000),
						SelfDelegationNaet:	sdkmath.NewInt(300),
						CommissionBps:		500,
						UptimeBps:		9_900,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Scenarios, 3)
	require.True(t, report.Scenarios[0].RewardAdjustmentBoundsPassed)
	require.True(t, report.Scenarios[1].ConcentrationWarnings > 0)
	require.Contains(t, report.Scenarios[2].Warnings, "low_participation_active_set_not_full")
}

func TestStakingEnhancementInvariantRejectsRewardAdjustmentOutOfBounds(t *testing.T) {
	params := DefaultStakingEnhancementParams()
	recommendation := ActiveSetRecommendation{
		EpochID:	3,
		AllowedChurn:	1,
		Selected: []ValidatorEconomicScore{
			{
				ValidatorID:			"val-a",
				Eligible:			true,
				BondedStakeNaet:		sdkmath.NewInt(1_000),
				ScoreBps:			9_000,
				RewardAdjustmentFactorBps:	params.MaxRewardAdjustmentBps + 1,
			},
		},
		Events:	[]ValidatorEconomicEvent{{Type: ValidatorEconomicEventSelected, EpochID: 3, ValidatorID: "val-a"}},
	}

	invariants := ValidateStakingEnhancementInvariants(recommendation, params)
	require.False(t, invariants.Passed)
	require.Contains(t, invariants.Failed, "reward_adjustment_out_of_bounds")
}

func stakingCandidate(id string, stake, self, uptime int64) ValidatorEconomicRecord {
	return ValidatorEconomicRecord{
		ValidatorID:		id,
		BondedStakeNaet:	sdkmath.NewInt(stake),
		SelfDelegationNaet:	sdkmath.NewInt(self),
		CommissionBps:		500,
		UptimeBps:		uptime,
		MetadataComplete:	true,
	}
}

func selectedValidatorIDs(scores []ValidatorEconomicScore) []string {
	ids := make([]string, 0, len(scores))
	for _, score := range scores {
		ids = append(ids, score.ValidatorID)
	}
	return ids
}
