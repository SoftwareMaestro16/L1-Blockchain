package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	scorekeeper "github.com/sovereign-l1/l1/x/aetra-validator-score/keeper"
	"github.com/sovereign-l1/l1/x/aetra-validator-score/types"
)

const authority = "ae1scoregov"

func TestKeeperExportImportPreservesScores(t *testing.T) {
	source := scorekeeper.NewKeeper(authority)
	_, err := source.UpdateScores(11, []types.ValidatorMetricInput{
		validatorMetrics("val-b", 9_800, 200),
		validatorMetrics("val-a", 10_000, 0),
	})
	require.NoError(t, err)

	exported, err := source.ExportGenesis()
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := scorekeeper.NewKeeper(authority)
	require.NoError(t, target.InitGenesis(exported))
	imported, err := target.ExportGenesis()
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func TestMarshalUnmarshalGenesisRoundTrip(t *testing.T) {
	source := scorekeeper.NewKeeper(authority)
	_, err := source.UpdateScores(12, []types.ValidatorMetricInput{validatorMetrics("val-a", 9_900, 100)})
	require.NoError(t, err)

	bz, err := source.MarshalGenesis()
	require.NoError(t, err)

	target := scorekeeper.NewKeeper(authority)
	require.NoError(t, target.UnmarshalGenesis(bz))
	imported, err := target.ExportGenesis()
	require.NoError(t, err)
	exported, err := source.ExportGenesis()
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func TestGovernanceAuthorityRequiredForMessages(t *testing.T) {
	k := scorekeeper.NewKeeper(authority)
	msgServer := scorekeeper.NewMsgServerImpl(&k)
	params := types.DefaultParams(authority)
	params.MinRewardMultiplierBps = 8_000

	err := msgServer.UpdateValidatorScoreParams(types.MsgUpdateValidatorScoreParams{
		Authority:	"ae1notgov",
		Params:		params,
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	require.NoError(t, msgServer.UpdateValidatorScoreParams(types.MsgUpdateValidatorScoreParams{
		Authority:	authority,
		Params:		params,
	}))
	res, err := k.QueryParams(types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint32(8_000), res.Params.MinRewardMultiplierBps)
}

func TestPublicValidatorMetricsQuery(t *testing.T) {
	k := scorekeeper.NewKeeper(authority)
	require.NoError(t, scorekeeper.NewMsgServerImpl(&k).UpdateValidatorScores(types.MsgUpdateValidatorScores{
		Authority:	authority,
		Epoch:		20,
		Metrics: []types.ValidatorMetricInput{
			validatorMetrics("val-a", 9_950, 50),
		},
	}))

	score, err := k.QueryValidatorScore(types.QueryValidatorScoreRequest{OperatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, uint64(20), score.Score.Epoch)
	require.False(t, score.Score.ConsensusOverrideAllowed)

	public, err := k.QueryPublicValidatorMetrics(types.QueryPublicValidatorMetricsRequest{OperatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, "val-a", public.Metrics.OperatorAddress)
	require.Equal(t, score.Score.OverallScoreBps, public.Metrics.OverallScoreBps)
	require.Equal(t, score.Score.RewardMultiplierBps, public.Metrics.RewardMultiplierBps)
}

func TestAllScoresQueryIsDeterministic(t *testing.T) {
	k := scorekeeper.NewKeeper(authority)
	_, err := k.UpdateScores(30, []types.ValidatorMetricInput{
		validatorMetrics("val-c", 9_000, 1_000),
		validatorMetrics("val-a", 10_000, 0),
		validatorMetrics("val-b", 9_900, 100),
	})
	require.NoError(t, err)

	all, err := k.QueryAllValidatorScores(types.QueryAllValidatorScoresRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{"val-a", "val-b", "val-c"}, []string{all.Scores[0].OperatorAddress, all.Scores[1].OperatorAddress, all.Scores[2].OperatorAddress})
}

func validatorMetrics(operator string, signed, missed uint64) types.ValidatorMetricInput {
	return types.ValidatorMetricInput{
		OperatorAddress:		operator,
		SignedBlocks:			signed,
		MissedBlocks:			missed,
		UptimeWindow:			10_000,
		SelfBond:			100,
		TotalBond:			1_000,
		CommissionHistory:		[]types.CommissionPoint{{Epoch: 1, CommissionBps: 500}, {Epoch: 2, CommissionBps: 500}},
		GovernanceVotes:		4,
		GovernanceProposals:		4,
		ConcentrationBps:		250,
		ConcentrationStatus:		types.ConcentrationStatusNormal,
		IdentityMetadataComplete:	true,
	}
}
