package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestCollectRolePerformanceMetricsBuildsDeterministicRecord(t *testing.T) {
	record, err := CollectRolePerformanceMetrics(RoleMetricCollectionInput{
		EpochID:			14,
		OperatorAddress:		"val-a",
		Role:				postypes.ValidatorRoleVerifier,
		AssignedTasks:			10,
		CompletedTasks:			7,
		MissedTasks:			2,
		InvalidTasks:			1,
		SignedBlocks:			90,
		TotalBlocks:			100,
		TaskParticipations:		8,
		MissedTaskParticipations:	2,
		CommittedLatencyWindow:		true,
		LatencyTargetMillis:		1_000,
		LatencyP95Millis:		2_000,
		ValidSignatures:		8,
		InvalidSignatures:		1,
		ValidTaskOutputs:		2,
		AcceptedEvidence:		1,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(8_700), record.UptimeScoreBps)
	require.Equal(t, uint32(5_000), record.LatencyScoreBps)
	require.Equal(t, uint32(7_692), record.CorrectnessScoreBps)
	require.Equal(t, uint32(7_000), record.TaskCompletionRateBps)
	require.Equal(t, uint32(2_342), record.RewardMultiplierBps)
	require.NoError(t, record.Validate())
}

func TestPerformanceQueriesReturnCurrentAndHistoricalRecords(t *testing.T) {
	verifier := performanceRecord(t, 14, "val-a", postypes.ValidatorRoleVerifier, 8_000)
	collator := performanceRecord(t, 14, "val-b", postypes.ValidatorRoleCollator, 6_000)
	older := performanceRecord(t, 13, "val-a", postypes.ValidatorRoleVerifier, 7_000)
	snapshot, err := NewPerformanceSnapshot([]postypes.PerformanceRecord{collator, older, verifier})
	require.NoError(t, err)

	current, err := QueryPerformanceRecord(snapshot, QueryPerformanceRecordRequest{
		EpochID:		14,
		OperatorAddress:	"val-a",
		Role:			postypes.ValidatorRoleVerifier,
	})
	require.NoError(t, err)
	require.Equal(t, verifier, current.Record)

	multiplier, err := QueryRewardMultiplier(snapshot, QueryRewardMultiplierRequest{
		EpochID:		14,
		OperatorAddress:	"val-a",
		Role:			postypes.ValidatorRoleVerifier,
	})
	require.NoError(t, err)
	require.Equal(t, verifier.RewardMultiplierBps, multiplier.RewardMultiplierBps)

	history, err := QueryOperatorPerformanceHistory(snapshot, QueryOperatorPerformanceHistoryRequest{
		OperatorAddress:	"val-a",
		Limit:			1,
	})
	require.NoError(t, err)
	require.Len(t, history.Records, 1)
	require.Equal(t, uint64(14), history.Records[0].EpochID)

	role, err := QueryRolePerformance(snapshot, QueryRolePerformanceRequest{
		EpochID:	14,
		Role:		postypes.ValidatorRoleCollator,
	})
	require.NoError(t, err)
	require.Equal(t, []postypes.PerformanceRecord{collator}, role.Records)
}

func TestPerformanceDistributionDampensRewardsBeforeDistribution(t *testing.T) {
	record := performanceRecord(t, 14, "val-a", postypes.ValidatorRoleVerifier, 5_000)
	result, err := SettlePerformanceDistribution(PerformanceDistributionInput{
		Performance:	record,
		Reward: postypes.RewardInput{
			ValidatorID:		"val-a",
			TotalRewardsNaet:	sdkmath.NewInt(1_000),
			CommissionBps:		1_000,
			SelfStakeNaet:		sdkmath.NewInt(1_000),
			Nominations: []postypes.Nomination{
				{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(1_000)},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000), result.OriginalRewardNaet)
	require.Equal(t, sdkmath.NewInt(500), result.DampenedRewardNaet)
	require.Equal(t, sdkmath.NewInt(50), result.Distribution.ValidatorCommissionNaet)
	require.Equal(t, sdkmath.NewInt(225), result.Distribution.ValidatorSelfShareNaet)
	require.Equal(t, sdkmath.NewInt(225), result.Distribution.NominatorRewards[0].RewardNaet)
	require.True(t, result.Distribution.TotalDistributedNaet.LTE(result.OriginalRewardNaet))
}

func TestPerformanceRejectsScoreManipulationAndRewardBounds(t *testing.T) {
	_, err := CollectRolePerformanceMetrics(RoleMetricCollectionInput{
		EpochID:		14,
		OperatorAddress:	"val-a",
		Role:			postypes.ValidatorRoleVerifier,
		AssignedTasks:		2,
		CompletedTasks:		2,
		InvalidTasks:		1,
	})
	require.ErrorContains(t, err, "task counts")

	_, err = CollectRolePerformanceMetrics(RoleMetricCollectionInput{
		EpochID:		14,
		OperatorAddress:	"val-a",
		Role:			postypes.ValidatorRoleVerifier,
		AssignedTasks:		1,
		CompletedTasks:		1,
		SignedBlocks:		1,
		TotalBlocks:		1,
		CommittedLatencyWindow:	false,
		ValidSignatures:	1,
	})
	require.ErrorContains(t, err, "committed measurement")

	record := performanceRecord(t, 14, "val-a", postypes.ValidatorRoleVerifier, 5_000)
	_, err = SettlePerformanceDistribution(PerformanceDistributionInput{
		Performance:	record,
		Reward: postypes.RewardInput{
			ValidatorID:		"val-b",
			TotalRewardsNaet:	sdkmath.NewInt(1_000),
			SelfStakeNaet:		sdkmath.NewInt(1),
		},
	})
	require.ErrorContains(t, err, "validator mismatch")

	tampered := record
	tampered.RewardMultiplierBps = postypes.BasisPoints + 1
	_, err = NewPerformanceSnapshot([]postypes.PerformanceRecord{tampered})
	require.ErrorContains(t, err, "reward multiplier")
}

func performanceRecord(t *testing.T, epoch uint64, operator string, role postypes.ValidatorRole, multiplier uint32) postypes.PerformanceRecord {
	t.Helper()
	record := postypes.PerformanceRecord{
		EpochID:		epoch,
		OperatorAddress:	operator,
		Role:			role,
		AssignedTasks:		1,
		CompletedTasks:		1,
		UptimeScoreBps:		postypes.BasisPoints,
		LatencyScoreBps:	postypes.BasisPoints,
		CorrectnessScoreBps:	multiplier,
		TaskCompletionRateBps:	postypes.BasisPoints,
		RewardMultiplierBps:	multiplier,
	}
	require.NoError(t, record.Validate())
	return record
}
