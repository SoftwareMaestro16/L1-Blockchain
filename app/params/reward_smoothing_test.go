package params

import (
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestSmoothEpochRewardsBoundsShortTermVariance(t *testing.T) {
	out, err := SmoothEpochRewards(RewardSmoothingInput{
		EpochID:			7,
		GrossRewardsNaet:		sdkmath.NewInt(20_000),
		PreviousEpochRewardsNaet:	sdkmath.NewInt(10_000),
		Validators: []ValidatorRewardParticipant{
			rewardValidator("val-b", 40, 500, rewardDelegator("del-b", 1_000)),
			rewardValidator("val-a", 60, 1_000, rewardDelegator("del-a", 3_000), rewardDelegator("del-c", 1_000)),
		},
	})

	require.NoError(t, err)
	require.True(t, out.BoundApplied)
	require.Equal(t, sdkmath.NewInt(12_000), out.SmoothedRewardsNaet)
	require.Equal(t, uint64(100), out.TotalVotingPower)
	require.Len(t, out.ValidatorRewards, 2)
	require.Equal(t, "val-a", out.ValidatorRewards[0].ValidatorID)
	require.Equal(t, "val-b", out.ValidatorRewards[1].ValidatorID)
	require.Contains(t, rewardSmoothingEventTypes(out.Events), RewardSmoothingEventBoundApplied)
	require.NoError(t, out.State.Validate())
}

func TestSmoothEpochRewardsCalculatesCommissionAndDelegatorRewards(t *testing.T) {
	out, err := SmoothEpochRewards(RewardSmoothingInput{
		EpochID:			8,
		GrossRewardsNaet:		sdkmath.NewInt(10_000),
		PreviousEpochRewardsNaet:	sdkmath.NewInt(10_000),
		Validators: []ValidatorRewardParticipant{
			rewardValidator("val-a", 100, 1_000, rewardDelegator("del-a", 3_000), rewardDelegator("del-b", 1_000)),
		},
	})

	require.NoError(t, err)
	require.False(t, out.BoundApplied)
	require.Equal(t, sdkmath.NewInt(10_000), out.ValidatorRewards[0].GrossRewardNaet)
	require.Equal(t, sdkmath.NewInt(1_000), out.ValidatorRewards[0].CommissionNaet)
	require.Equal(t, sdkmath.NewInt(9_000), out.ValidatorRewards[0].DelegatorRewardPoolNaet)
	require.Equal(t, []DelegatorRewardAllocation{
		{DelegatorID: "del-a", RewardNaet: sdkmath.NewInt(6_750)},
		{DelegatorID: "del-b", RewardNaet: sdkmath.NewInt(2_250)},
	}, out.ValidatorRewards[0].DelegatorRewards)
	require.NoError(t, out.State.Validate())
}

func TestRewardSmoothingStateExportImportSafe(t *testing.T) {
	out, err := SmoothEpochRewards(RewardSmoothingInput{
		EpochID:			9,
		GrossRewardsNaet:		sdkmath.NewInt(1_001),
		PreviousEpochRewardsNaet:	sdkmath.NewInt(1_000),
		Validators: []ValidatorRewardParticipant{
			rewardValidator("val-b", 1, 500, rewardDelegator("del-b", 1)),
			rewardValidator("val-a", 2, 500, rewardDelegator("del-a", 1)),
		},
	})
	require.NoError(t, err)

	bz, err := json.Marshal(out.State)
	require.NoError(t, err)
	var imported RewardSmoothingState
	require.NoError(t, json.Unmarshal(bz, &imported))
	require.Equal(t, out.State, imported)
	require.NoError(t, imported.Validate())
}

func TestRewardSmoothingStateRejectsAccountingDrift(t *testing.T) {
	state := RewardSmoothingState{
		EpochID:		10,
		EpochLengthBlocks:	DefaultRewardSmoothingEpochLengthBlocks,
		TotalRewardsNaet:	sdkmath.NewInt(100),
		ValidatorRewards: []ValidatorRewardAllocation{{
			ValidatorID:			"val-a",
			GrossRewardNaet:		sdkmath.NewInt(100),
			CommissionNaet:			sdkmath.NewInt(10),
			DelegatorRewardPoolNaet:	sdkmath.NewInt(80),
			DelegatorRewards:		[]DelegatorRewardAllocation{{DelegatorID: "del-a", RewardNaet: sdkmath.NewInt(80)}},
		}},
	}

	require.ErrorContains(t, state.Validate(), "validator reward allocation")
}

func rewardValidator(id string, power uint64, commissionBps int64, delegators ...DelegatorRewardParticipant) ValidatorRewardParticipant {
	return ValidatorRewardParticipant{
		ValidatorID:	id,
		VotingPower:	power,
		CommissionBps:	commissionBps,
		DelegatorStake:	delegators,
	}
}

func rewardDelegator(id string, stake int64) DelegatorRewardParticipant {
	return DelegatorRewardParticipant{DelegatorID: id, StakeNaet: sdkmath.NewInt(stake)}
}

func rewardSmoothingEventTypes(events []RewardSmoothingEvent) []string {
	out := make([]string, 0, len(events))
	for _, event := range events {
		out = append(out, event.Type)
	}
	return out
}
