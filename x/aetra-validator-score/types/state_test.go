package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aetra-validator-score/types"
)

const authority = "ae1scoregov"

func TestUptimeAccountingScoresSignedWindow(t *testing.T) {
	scores, err := types.ComputeValidatorScores(types.DefaultParams(authority), 7, []types.ValidatorMetricInput{
		validatorMetrics("val-a", 9_500, 500),
	})
	require.NoError(t, err)
	require.Len(t, scores, 1)
	require.Equal(t, uint32(9_500), scores[0].UptimeScoreBps)
	require.Equal(t, uint32(9_500), scores[0].MissedBlockScoreBps)
	require.False(t, scores[0].ConsensusOverrideAllowed)
	require.Less(t, scores[0].RewardMultiplierBps, types.BasisPoints)
}

func TestMissedBlockWindowRejectsImpossibleCounts(t *testing.T) {
	input := validatorMetrics("val-a", 9_800, 300)
	input.UptimeWindow = 10_000
	_, err := types.ComputeValidatorScores(types.DefaultParams(authority), 1, []types.ValidatorMetricInput{input})
	require.ErrorIs(t, err, types.ErrInvalidScore)
}

func TestScoreUpdateEpochDeterministicAndSorted(t *testing.T) {
	params := types.DefaultParams(authority)
	inputs := []types.ValidatorMetricInput{
		validatorMetrics("val-c", 9_000, 1_000),
		validatorMetrics("val-a", 10_000, 0),
		validatorMetrics("val-b", 9_900, 100),
	}
	scores, err := types.ComputeValidatorScores(params, 42, inputs)
	require.NoError(t, err)
	require.Equal(t, []string{"val-a", "val-b", "val-c"}, []string{scores[0].OperatorAddress, scores[1].OperatorAddress, scores[2].OperatorAddress})
	require.Equal(t, uint64(42), scores[0].Epoch)
	require.Greater(t, scores[0].OverallScoreBps, scores[2].OverallScoreBps)
}

func TestSlashHistoryReducesScore(t *testing.T) {
	clean := validatorMetrics("val-a", 10_000, 0)
	slashed := validatorMetrics("val-b", 10_000, 0)
	slashed.SlashEvents = []types.SlashEvent{
		{Height: 15, FractionBps: 500, Reason: "double_sign"},
		{Height: 30, FractionBps: 100, Reason: "downtime"},
	}
	scores, err := types.ComputeValidatorScores(types.DefaultParams(authority), 3, []types.ValidatorMetricInput{slashed, clean})
	require.NoError(t, err)
	require.Equal(t, uint32(8_400), scores[1].SlashHistoryScoreBps)
	require.Less(t, scores[1].OverallScoreBps, scores[0].OverallScoreBps)
}

func TestJailHistoryReducesScoreAndRewardMultiplier(t *testing.T) {
	clean := validatorMetrics("val-a", 10_000, 0)
	jailed := validatorMetrics("val-b", 10_000, 0)
	jailed.JailEvents = 2

	scores, err := types.ComputeValidatorScores(types.DefaultParams(authority), 4, []types.ValidatorMetricInput{jailed, clean})
	require.NoError(t, err)
	require.Equal(t, uint32(8_000), scores[1].JailScoreBps)
	require.Less(t, scores[1].OverallScoreBps, scores[0].OverallScoreBps)
	require.Less(t, scores[1].RewardMultiplierBps, scores[0].RewardMultiplierBps)
}

func TestGovernanceParticipationScore(t *testing.T) {
	input := validatorMetrics("val-a", 10_000, 0)
	input.GovernanceProposals = 8
	input.GovernanceVotes = 6
	scores, err := types.ComputeValidatorScores(types.DefaultParams(authority), 1, []types.ValidatorMetricInput{input})
	require.NoError(t, err)
	require.Equal(t, uint32(7_500), scores[0].GovernanceScoreBps)
	require.Equal(t, uint32(7_500), scores[0].GovernanceParticipationBps)
}

func TestDeterministicScoringIndependentOfInputOrder(t *testing.T) {
	inputs := []types.ValidatorMetricInput{
		validatorMetrics("val-b", 9_900, 100),
		validatorMetrics("val-a", 9_700, 300),
	}
	reversed := []types.ValidatorMetricInput{inputs[1], inputs[0]}

	first, err := types.ComputeValidatorScores(types.DefaultParams(authority), 9, inputs)
	require.NoError(t, err)
	second, err := types.ComputeValidatorScores(types.DefaultParams(authority), 9, reversed)
	require.NoError(t, err)
	require.Equal(t, first, second)
}

func TestGenesisValidationRejectsConsensusOverrideDrift(t *testing.T) {
	params := types.DefaultParams(authority)
	score := types.ValidatorScore{
		OperatorAddress:          "val-a",
		RewardMultiplierBps:      params.MinRewardMultiplierBps,
		ConsensusOverrideAllowed: true,
	}
	genesis := types.GenesisState{Params: params, Scores: []types.ValidatorScore{score}}
	require.Error(t, genesis.Validate())
}

func validatorMetrics(operator string, signed, missed uint64) types.ValidatorMetricInput {
	return types.ValidatorMetricInput{
		OperatorAddress:          operator,
		SignedBlocks:             signed,
		MissedBlocks:             missed,
		UptimeWindow:             10_000,
		SelfBond:                 100,
		TotalBond:                1_000,
		CommissionHistory:        []types.CommissionPoint{{Epoch: 1, CommissionBps: 500}, {Epoch: 2, CommissionBps: 500}},
		GovernanceVotes:          4,
		GovernanceProposals:      4,
		ConcentrationBps:         250,
		ConcentrationStatus:      types.ConcentrationStatusNormal,
		IdentityMetadataComplete: true,
	}
}
