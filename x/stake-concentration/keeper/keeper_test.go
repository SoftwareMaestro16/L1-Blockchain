package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	stakeconcentrationkeeper "github.com/sovereign-l1/l1/x/stake-concentration/keeper"
	"github.com/sovereign-l1/l1/x/stake-concentration/types"
)

func TestValidatorAboveCapRejectsNewDelegation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)

	res, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority: app.StakeConcentrationKeeper.Authority(),
		Epoch:     1,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x33), VotingPower: 25},
			{OperatorAddress: operator(0x11), VotingPower: 50},
			{OperatorAddress: operator(0x22), VotingPower: 25},
		},
	})
	require.NoError(t, err)

	require.Equal(t, operator(0x11), res.Network.Validators[0].OperatorAddress)
	require.Equal(t, uint32(5_000), res.Network.Validators[0].RawVotingPowerBps)
	require.Equal(t, uint32(3_334), res.Network.Validators[0].EffectiveVotingPowerBps)
	require.True(t, res.Network.Validators[0].AboveHardCap)
	require.False(t, res.Network.Validators[0].DelegationAllowed)

	allowed, err := app.StakeConcentrationKeeper.CanAcceptDelegation(ctx, operator(0x11))
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestRewardModifierApplies(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)

	res, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority: app.StakeConcentrationKeeper.Authority(),
		Epoch:     1,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x11), VotingPower: 40},
			{OperatorAddress: operator(0x22), VotingPower: 30},
			{OperatorAddress: operator(0x33), VotingPower: 30},
		},
	})
	require.NoError(t, err)

	metric := res.Network.Validators[0]
	require.Equal(t, uint32(4_000), metric.RawVotingPowerBps)
	require.Equal(t, uint32(9_400), metric.RewardModifierBps)
	require.Less(t, metric.RewardModifierBps, types.BasisPoints)
	require.GreaterOrEqual(t, metric.RewardModifierBps, uint32(0))
}

func TestPowerCapEnforcedAcrossEpochTransition(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)

	first, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority: app.StakeConcentrationKeeper.Authority(),
		Epoch:     1,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x11), VotingPower: 34},
			{OperatorAddress: operator(0x22), VotingPower: 33},
			{OperatorAddress: operator(0x33), VotingPower: 33},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(3_334), first.Network.MaxValidatorPowerBps)

	second, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority: app.StakeConcentrationKeeper.Authority(),
		Epoch:     2,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x11), VotingPower: 70},
			{OperatorAddress: operator(0x22), VotingPower: 20},
			{OperatorAddress: operator(0x33), VotingPower: 10},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(2), second.Network.Epoch)
	require.Equal(t, uint32(3_334), second.Network.MaxValidatorPowerBps)
	require.Equal(t, uint32(3_334), second.Network.Validators[0].EffectiveVotingPowerBps)
}

func TestInvalidParamsRejected(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)
	params := types.DefaultParams()
	params.SoftVotingPowerBps = params.MaxVotingPowerBps + 1

	_, err := msgServer.UpdateConcentrationParams(ctx, &types.MsgUpdateConcentrationParams{
		Authority: app.StakeConcentrationKeeper.Authority(),
		Params:    params,
	})
	require.ErrorIs(t, err, types.ErrInvalidParams)
}

func TestExportImportPreservesConcentrationMetrics(t *testing.T) {
	source := l1app.Setup(t, false)
	sourceCtx := source.NewContext(false)
	_, err := source.StakeConcentrationKeeper.RecomputeConcentration(sourceCtx, 7, []types.ValidatorPower{
		{OperatorAddress: operator(0x11), VotingPower: 50},
		{OperatorAddress: operator(0x22), VotingPower: 30},
		{OperatorAddress: operator(0x33), VotingPower: 20},
	})
	require.NoError(t, err)

	exported, err := source.StakeConcentrationKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := l1app.Setup(t, false)
	targetCtx := target.NewContext(false)
	require.NoError(t, target.StakeConcentrationKeeper.InitGenesis(targetCtx, *exported))
	imported, err := target.StakeConcentrationKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func TestOverflowVotingPowerRejected(t *testing.T) {
	params := types.DefaultParams()
	_, err := types.ComputeNetworkConcentration(params, 1, []types.ValidatorPower{
		{OperatorAddress: operator(0x11), VotingPower: ^uint64(0)},
		{OperatorAddress: operator(0x22), VotingPower: 1},
	}, 1)
	require.ErrorIs(t, err, types.ErrInvalidConcentration)
}

func operator(fill byte) string {
	return aetherisaddress.FormatAccAddress(sdk.AccAddress(bytes20(fill)))
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
