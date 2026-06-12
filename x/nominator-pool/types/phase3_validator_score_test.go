package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPhase35ValidatorScoreDeterministicAndDrivesWeights(t *testing.T) {
	params := DefaultParams()
	valA := chat3AEAddress(0xd1)
	valB := chat3AEAddress(0xd2)
	candidates := []ValidatorPolicyCandidate{
		phase35Candidate(valB, func(c *ValidatorPolicyCandidate) {
			c.UptimeWindow = 100
			c.MissedBlocks = 10
			c.CommissionBps = params.ValidatorCommissionCeilingBps
		}),
		phase35Candidate(valA, nil),
	}

	first, err := params.AllocationWeights(candidates)
	require.NoError(t, err)
	second, err := params.AllocationWeights([]ValidatorPolicyCandidate{candidates[1], candidates[0]})
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Greater(t, first[0].Score, first[1].Score)
	require.Greater(t, first[0].WeightBps, first[1].WeightBps)
}

func TestPhase35ValidatorScoreChangesWithInputsAndRejectsIneligible(t *testing.T) {
	params := DefaultParams()
	base, err := params.ComputeValidatorScoreV1(phase35Candidate(chat3AEAddress(0xd3), nil))
	require.NoError(t, err)
	require.True(t, base.Eligible)

	lowUptime, err := params.ComputeValidatorScoreV1(phase35Candidate(chat3AEAddress(0xd3), func(c *ValidatorPolicyCandidate) {
		c.UptimeWindow = 100
		c.MissedBlocks = 30
	}))
	require.NoError(t, err)
	require.Less(t, lowUptime.OverallScoreBps, base.OverallScoreBps)

	highCommission, err := params.ComputeValidatorScoreV1(phase35Candidate(chat3AEAddress(0xd3), func(c *ValidatorPolicyCandidate) {
		c.CommissionBps = params.ValidatorCommissionCeilingBps
	}))
	require.NoError(t, err)
	require.Less(t, highCommission.OverallScoreBps, base.OverallScoreBps)

	slashed, err := params.ComputeValidatorScoreV1(phase35Candidate(chat3AEAddress(0xd3), func(c *ValidatorPolicyCandidate) {
		c.SlashingRiskBps = 2_000
		c.Slashed = true
	}))
	require.NoError(t, err)
	require.False(t, slashed.Eligible)
	require.Less(t, slashed.SlashingRiskScoreBps, base.SlashingRiskScoreBps)

	_, err = params.ComputeValidatorScoreV1(phase35Candidate(chat3AEAddress(0xd3), func(c *ValidatorPolicyCandidate) {
		c.UptimeWindow = 10
		c.MissedBlocks = 11
	}))
	require.ErrorContains(t, err, "missed blocks exceed uptime window")
}

func TestPhase35ExportImportPreservesScoresAndSnapshots(t *testing.T) {
	params := DefaultParams()
	validator := chat3AEAddress(0xd4)
	state := State{
		ValidatorPerformanceScores: []ValidatorPerformanceScore{{
			Validator:	validator,
			Epoch:		7,
			ScoreBps:	8_765,
		}},
		EpochStakingSnapshots: []EpochStakingSnapshot{{
			Epoch:			7,
			TotalActiveStake:	100,
			TotalPools:		1,
			ValidatorCount:		1,
			SnapshotHash:		"epoch-hash",
		}},
		ValidatorSetSnapshots: []ValidatorSetSnapshot{{
			HeightOrEpoch:	7,
			Validators:	[]string{validator},
			TotalPower:	100,
			SnapshotHash:	"set-hash",
		}},
	}
	normalized := state.Normalize(params)
	require.NoError(t, normalized.Validate(params))

	imported := normalized.Normalize(params)
	require.Equal(t, normalized.ValidatorPerformanceScores, imported.ValidatorPerformanceScores)
	require.Equal(t, normalized.EpochStakingSnapshots, imported.EpochStakingSnapshots)
	require.Equal(t, normalized.ValidatorSetSnapshots, imported.ValidatorSetSnapshots)
}

func phase35Candidate(validator string, mutate func(*ValidatorPolicyCandidate)) ValidatorPolicyCandidate {
	candidate := ValidatorPolicyCandidate{
		ValidatorAddress:	validator,
		ReputationScore:	MaxBasisPoints,
		UptimeBps:		MaxBasisPoints,
		UptimeWindow:		100,
		MissedBlocks:		0,
		CommissionBps:		DefaultParams().ValidatorCommissionFloorBps,
		StakeEfficiencyBps:	MaxBasisPoints,
		SlashingRiskBps:	0,
		NetworkLoadBps:		0,
		AllocationLimitBps:	DefaultParams().MaxPoolValidatorAllocationBps,
		OperationalHistoryBps:	MaxBasisPoints,
		CurrentAllocationBps:	0,
	}
	if mutate != nil {
		mutate(&candidate)
	}
	return candidate
}
