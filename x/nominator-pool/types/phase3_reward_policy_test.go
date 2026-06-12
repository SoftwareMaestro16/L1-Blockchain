package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPhase33DefaultRewardPolicyV1CapturesLaunchRules(t *testing.T) {
	policy := DefaultRewardPolicyV1()
	require.NoError(t, policy.Validate())
	require.Equal(t, RewardSourceFeesAndInflation, policy.RewardSource)
	require.Equal(t, RewardDistributionByPoolShares, policy.Distribution)
	require.Equal(t, RewardValidatorCommissionBeforePool, policy.CommissionRule)
	require.Equal(t, RewardRoundingFloorWithRemainder, policy.RoundingRule)
	require.True(t, policy.LazyRewardIndex)
	require.True(t, policy.CapByCollectedRewards)
	require.False(t, policy.ManualValidatorChoice)

	policy.ManualValidatorChoice = true
	require.ErrorContains(t, policy.Validate(), "manual validator choice")
}

func TestPhase33RewardIndexDeterministicAndCappedByFeesPlusInflation(t *testing.T) {
	params := DefaultParams()
	pool := NominatorPool{
		PoolID:			"phase33-rewards",
		TotalShares:		3,
		TotalBondedStake:	3,
		PoolCommissionBps:	0,
		Status:			PoolStatusActive,
	}
	msg := MsgSyncPoolRewards{
		Authority:		params.Authority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		Height:			10,
		RewardRateBps:		3_334,
		EmissionsAllocated:	0,
		FeesAllocated:		1,
		Allocations: []ValidatorRewardAllocation{{
			Validator:		chat3AEAddress(0xc1),
			PoolAllocatedStake:	3,
			PerformanceBps:		MaxBasisPoints,
		}},
	}

	first, firstSummary, err := SyncPoolRewards(params, pool, msg)
	require.NoError(t, err)
	second, secondSummary, err := SyncPoolRewards(params, pool, msg)
	require.NoError(t, err)
	require.Equal(t, first.RewardIndex, second.RewardIndex)
	require.Equal(t, first.RewardRemainder, second.RewardRemainder)
	require.Equal(t, firstSummary, secondSummary)
	require.Equal(t, uint64(333_333_333), firstSummary.RewardIndexAfter)
	require.Equal(t, uint64(1), firstSummary.RewardRemainder)

	msg.FeesAllocated = 0
	_, _, err = SyncPoolRewards(params, pool, msg)
	require.ErrorContains(t, err, "exceed emissions and fee allocation cap")
}

func TestPhase33SlashedValidatorDoesNotProducePositiveOperatorBonus(t *testing.T) {
	params := DefaultParams()
	pool := NominatorPool{
		PoolID:			"phase33-slashed",
		TotalShares:		1_000,
		TotalBondedStake:	100_000,
		PoolCommissionBps:	0,
		Status:			PoolStatusActive,
	}
	next, summary, err := SyncPoolRewards(params, pool, MsgSyncPoolRewards{
		Authority:		params.Authority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		Height:			10,
		RewardRateBps:		1_000,
		EmissionsAllocated:	30_000,
		Allocations: []ValidatorRewardAllocation{{
			Validator:			chat3AEAddress(0xc2),
			PoolAllocatedStake:		100_000,
			ValidatorSelfStake:		50_000,
			PerformanceBps:			MaxBasisPoints,
			CommissionBps:			500,
			OperatorPerformanceBonusBps:	500,
			SlashingLoss:			1,
		}},
	})
	require.NoError(t, err)
	require.Positive(t, summary.GrossPoolRewards)
	require.Zero(t, summary.OperatorPerformanceBonus)
	require.Len(t, next.ValidatorAllocations, 1)
	require.Zero(t, next.ValidatorAllocations[0].OperatorPerformanceBonus)
	require.Zero(t, next.ValidatorAllocations[0].OperatorPerformanceBonusBps)
}
