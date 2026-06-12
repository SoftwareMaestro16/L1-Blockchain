package types

import (
	"encoding/json"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func TestCHAT3AllocationWeightsDeterministicGoldenAndRejectJailedSlashedOverCap(t *testing.T) {
	params := DefaultParams()
	valA := chat3AEAddress(0x11)
	valB := chat3AEAddress(0x22)
	valC := chat3AEAddress(0x33)
	valD := chat3AEAddress(0x44)
	candidates := []ValidatorPolicyCandidate{
		chat3Candidate(params, valD, func(c *ValidatorPolicyCandidate) { c.CurrentAllocationBps = params.MaxPoolValidatorAllocationBps }),
		chat3Candidate(params, valB, nil),
		chat3Candidate(params, valC, func(c *ValidatorPolicyCandidate) { c.Jailed = true; c.Slashed = true }),
		chat3Candidate(params, valA, nil),
	}

	weights, err := params.AllocationWeights(candidates)
	require.NoError(t, err)

	require.Equal(t, []AllocationWeight{
		{ValidatorAddress: valA, Score: weights[0].Score, WeightBps: 5_000},
		{ValidatorAddress: valB, Score: weights[1].Score, WeightBps: 5_000},
		{ValidatorAddress: valC, Score: 0, WeightBps: 0},
		{ValidatorAddress: valD, Score: 0, WeightBps: 0},
	}, weights)
	require.NotZero(t, weights[0].Score)
	require.Equal(t, weights[0].Score, weights[1].Score)

	reversed := []ValidatorPolicyCandidate{candidates[3], candidates[2], candidates[1], candidates[0]}
	for i := 0; i < 8; i++ {
		again, err := params.AllocationWeights(reversed)
		require.NoError(t, err)
		require.Equal(t, weights, again)
	}
}

func TestCHAT3AllocationPlanTouchesBoundedPoolAllocationKeysOnly(t *testing.T) {
	valA := chat3AEAddress(0x55)
	valB := chat3AEAddress(0x66)
	plan, err := BuildPoolAllocationPlan(PoolAllocationPlanInput{
		PoolID:			"pool-chat3",
		Epoch:			9,
		Height:			90,
		MaxTouchedAllocations:	1,
		Weights: []AllocationWeight{
			{ValidatorAddress: valB, WeightBps: 4_000},
			{ValidatorAddress: valA, WeightBps: 6_000},
		},
	})

	require.NoError(t, err)
	require.Len(t, plan.Allocations, 1)
	require.Equal(t, valA, plan.Allocations[0].Validator)
	require.Equal(t, uint32(6_000), plan.Allocations[0].TargetWeightBps)
	require.Equal(t, []string{string(PoolAllocationKey("pool-chat3", valA))}, plan.InternalMetadata.TouchedKeys)
}

func TestCHAT3AllocationPlanAppliesDeterministicBoundedStateTransition(t *testing.T) {
	params := DefaultParams()
	valA := chat3AEAddress(0xa1)
	valB := chat3AEAddress(0xa2)
	valOther := chat3AEAddress(0xa3)
	existing := []PoolValidatorAllocation{
		{PoolID: "pool-chat3", Validator: valA, TargetWeightBps: 25, UpdatedHeight: 10},
		{PoolID: "pool-other", Validator: valOther, TargetWeightBps: 25, UpdatedHeight: 10},
	}
	receipt := PoolRebalanceReceipt{
		PoolID:	"pool-chat3",
		Epoch:	2,
		Height:	20,
		Allocations: []PoolValidatorAllocation{
			{PoolID: "pool-chat3", Validator: valB, TargetWeightBps: 100},
			{PoolID: "pool-chat3", Validator: valA, TargetWeightBps: 200},
		},
	}

	next, touched, err := ApplyPoolAllocationPlan(params, existing, receipt)
	require.NoError(t, err)
	require.Equal(t, []string{
		string(PoolAllocationKey("pool-chat3", valA)),
		string(PoolAllocationKey("pool-chat3", valB)),
	}, touched)
	require.Equal(t, []PoolValidatorAllocation{
		{PoolID: "pool-chat3", Validator: valA, TargetWeightBps: 200, UpdatedHeight: 20},
		{PoolID: "pool-chat3", Validator: valB, TargetWeightBps: 100, UpdatedHeight: 20},
		{PoolID: "pool-other", Validator: valOther, TargetWeightBps: 25, UpdatedHeight: 10},
	}, next)

	receipt.Allocations[0].TargetWeightBps = params.MaxPoolValidatorAllocationBps + 1
	_, _, err = ApplyPoolAllocationPlan(params, existing, receipt)
	require.ErrorContains(t, err, "exceeds configured cap")
}

func TestCHAT3ValidatorPowerCapIsStage1RewardsOnlyUntilCometBFTEvidence(t *testing.T) {
	scope := ActiveValidatorPowerCapScope()
	require.NoError(t, scope.Validate())
	require.Equal(t, ValidatorPowerCapStageRewardsOnly, scope.Stage)
	require.True(t, scope.CapsPoolAllocationWeight)
	require.True(t, scope.CapsRewardEffectivePower)
	require.False(t, scope.CapsCometBFTVotingPower)
	require.Equal(t, "x/pos+x/staking+CometBFT", scope.ConsensusVotingPowerOwner)

	allocationType := reflect.TypeOf(PoolValidatorAllocation{})
	_, hasCometBFTPower := allocationType.FieldByName("CometBFTVotingPower")
	_, hasConsensusPower := allocationType.FieldByName("ConsensusVotingPower")
	require.False(t, hasCometBFTPower)
	require.False(t, hasConsensusPower)

	invalidStage1 := scope
	invalidStage1.CapsCometBFTVotingPower = true
	require.ErrorContains(t, invalidStage1.Validate(), "must not claim CometBFT")

	stage2 := ValidatorPowerCapScope{
		Stage:				ValidatorPowerCapStageCometBFT,
		CapsPoolAllocationWeight:	true,
		CapsRewardEffectivePower:	true,
		CapsCometBFTVotingPower:	true,
	}
	require.NoError(t, stage2.Validate())
}

func TestCHAT3PoolRewardsDeductCommissionThenPoolFeeAndCapRewards(t *testing.T) {
	params := DefaultParams()
	pool := chat3RewardPool()
	validator := chat3AEAddress(0x77)
	msg := MsgSyncPoolRewards{
		Authority:		params.Authority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		Height:			10,
		RewardRateBps:		1_000,
		EmissionsAllocated:	20_000,
		Allocations: []ValidatorRewardAllocation{{
			Validator:		validator,
			PoolAllocatedStake:	100_000,
			ValidatorSelfStake:	50_000,
			PerformanceBps:		MaxBasisPoints,
			CommissionBps:		500,
		}},
	}

	next, summary, err := SyncPoolRewards(params, pool, msg)
	require.NoError(t, err)
	require.Equal(t, uint64(10_000), summary.GrossPoolRewards)
	require.Equal(t, uint64(500), summary.ValidatorCommission)
	require.Equal(t, uint64(95), summary.PoolProtocolFee)
	require.Equal(t, uint64(9_405), summary.PoolUserRewards)
	require.Equal(t, uint64(5_000), summary.ValidatorSelfStakeRewards)
	require.Equal(t, uint64(9_405_000_000), summary.RewardIndexAfter)
	require.Equal(t, summary.RewardIndexAfter, next.RewardIndex)
	require.Equal(t, uint64(0), next.RewardRemainder)
	require.Equal(t, summary.PoolProtocolFee, next.ProtocolFeeAccrued)
	require.Equal(t, summary.ValidatorCommission, next.ValidatorCommissionAccrued)

	msg.EmissionsAllocated = 10_000
	_, _, err = SyncPoolRewards(params, pool, msg)
	require.ErrorContains(t, err, "exceed emissions and fee allocation cap")
}

func TestCHAT3PoolRewardClaimUpdatesCallerOnlyAndScalesToMillionUsers(t *testing.T) {
	ownerA := chat3AEAddress(0x88)
	ownerB := chat3AEAddress(0x99)
	rewardIndex := uint64(9_405_000_000)
	shareA := DelegatorShare{Delegator: ownerA, Shares: 250}
	shareB := DelegatorShare{Delegator: ownerB, Shares: 750}

	nextA, receiptA, err := ClaimPoolRewardShare(PoolRewardClaimInput{
		PoolID:		"pool-chat3",
		Share:		shareA,
		RewardIndex:	rewardIndex,
		Epoch:		2,
		Height:		20,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(2_351), receiptA.Amount)
	require.Equal(t, rewardIndex, nextA.RewardIndexCheckpoint)
	require.Zero(t, nextA.PendingRewards)
	require.Equal(t, []string{string(PoolShareKey("pool-chat3", ownerA))}, receiptA.InternalMetadata.TouchedKeys)
	require.Equal(t, shareB, shareB)

	nextB, receiptB, err := ClaimPoolRewardShare(PoolRewardClaimInput{
		PoolID:		"pool-chat3",
		Share:		shareB,
		RewardIndex:	rewardIndex,
		Epoch:		2,
		Height:		20,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(7_053), receiptB.Amount)
	require.Equal(t, rewardIndex, nextB.RewardIndexCheckpoint)
	require.Len(t, receiptB.InternalMetadata.TouchedKeys, 1)

	const simulatedUsers = 1_000_000
	for i := 0; i < 4; i++ {
		next, receipt, err := ClaimPoolRewardShare(PoolRewardClaimInput{
			PoolID:		"pool-chat3",
			Share:		DelegatorShare{Delegator: ownerA, Shares: uint64(simulatedUsers + i), PendingRewards: 1},
			RewardIndex:	rewardIndex,
			Epoch:		3,
			Height:		30,
		})
		require.NoError(t, err)
		require.NotZero(t, receipt.Amount)
		require.Len(t, receipt.InternalMetadata.TouchedKeys, 1)
		require.Equal(t, ownerA, next.Delegator)
	}
}

func TestCHAT3PoolRewardStateRecordsAndExportImportRoundTrip(t *testing.T) {
	params := DefaultParams()
	owner := chat3AEAddress(0xb1)
	pool := chat3RewardPool()
	nextPool, _, err := SyncPoolRewards(params, pool, MsgSyncPoolRewards{
		Authority:		params.Authority,
		PoolID:			pool.PoolID,
		Epoch:			4,
		Height:			40,
		RewardRateBps:		1_000,
		EmissionsAllocated:	20_000,
		Allocations: []ValidatorRewardAllocation{{
			Validator:		chat3AEAddress(0xb2),
			PoolAllocatedStake:	100_000,
			ValidatorSelfStake:	50_000,
			PerformanceBps:		MaxBasisPoints,
			CommissionBps:		500,
		}},
	})
	require.NoError(t, err)
	share, receipt, err := ClaimPoolRewardShare(PoolRewardClaimInput{
		PoolID:		pool.PoolID,
		Share:		DelegatorShare{Delegator: owner, Shares: 1_000},
		RewardIndex:	nextPool.RewardIndex,
		Epoch:		4,
		Height:		41,
	})
	require.NoError(t, err)
	require.Equal(t, nextPool.RewardIndex, share.RewardIndexCheckpoint)

	claims, claim, err := RecordPoolRewardClaim(params, nil, receipt)
	require.NoError(t, err)
	require.Equal(t, RewardClaim{PoolID: pool.PoolID, Owner: owner, Epoch: 4, Amount: receipt.Amount}, claim)
	_, _, err = RecordPoolRewardClaim(params, claims, receipt)
	require.ErrorContains(t, err, "duplicate reward claim")
	_, _, err = RecordPoolRewardClaim(params, nil, PoolRewardClaimReceipt{
		PoolID:		pool.PoolID,
		OwnerAddress:	"notvalid1address",
		Amount:		receipt.Amount,
		Epoch:		4,
		Height:		41,
	})
	require.Error(t, err)
	_, _, err = RecordPoolRewardClaim(params, nil, PoolRewardClaimReceipt{
		PoolID:		pool.PoolID,
		OwnerAddress:	owner,
		Epoch:		4,
		Height:		41,
	})
	require.ErrorContains(t, err, "amount must be positive")

	state := State{
		PoolRewardIndexes:	[]PoolRewardIndex{{PoolID: nextPool.PoolID, RewardIndex: nextPool.RewardIndex, Epoch: nextPool.RewardEpoch}},
		RewardClaims:		claims,
	}
	normalized := state.Normalize(params)
	require.NoError(t, normalized.Validate(params))
	exported, err := json.Marshal(normalized)
	require.NoError(t, err)
	var imported State
	require.NoError(t, json.Unmarshal(exported, &imported))

	roundTrip := imported.Normalize(params)
	require.Equal(t, normalized.PoolRewardIndexes, roundTrip.PoolRewardIndexes)
	require.Equal(t, normalized.RewardClaims, roundTrip.RewardClaims)
	require.Equal(t, nextPool.RewardIndex, roundTrip.PoolRewardIndexes[0].RewardIndex)
	require.Equal(t, receipt.Amount, roundTrip.RewardClaims[0].Amount)
}

func TestCHAT3RewardFormulaFixtureForThreeHundredThousandAET(t *testing.T) {
	const stake = uint64(300_000) * DefaultAETBaseUnits
	reward, err := RewardForStake(stake, 350, 9_500)
	require.NoError(t, err)
	require.Equal(t, uint64(9_975_000_000_000), reward)
}

func chat3Candidate(params Params, validator string, mutate func(*ValidatorPolicyCandidate)) ValidatorPolicyCandidate {
	candidate := ValidatorPolicyCandidate{
		ValidatorAddress:	validator,
		ReputationScore:	10_000,
		UptimeBps:		10_000,
		CommissionBps:		params.ValidatorCommissionFloorBps,
		StakeEfficiencyBps:	10_000,
		SlashingRiskBps:	0,
		NetworkLoadBps:		0,
	}
	if mutate != nil {
		mutate(&candidate)
	}
	return candidate
}

func chat3RewardPool() NominatorPool {
	return NominatorPool{
		PoolID:			"pool-chat3",
		TotalShares:		1_000,
		TotalBondedStake:	100_000,
		PoolCommissionBps:	100,
		Status:			PoolStatusActive,
	}
}

func chat3AEAddress(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bz))
}
