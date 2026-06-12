package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	stakeconcentrationkeeper "github.com/sovereign-l1/l1/x/stake-concentration/keeper"
	"github.com/sovereign-l1/l1/x/stake-concentration/types"
)

func TestValidatorAboveCapRejectsNewDelegation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)

	res, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority:	app.StakeConcentrationKeeper.Authority(),
		Epoch:		1,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x33), VotingPower: 25},
			{OperatorAddress: operator(0x11), VotingPower: 50},
			{OperatorAddress: operator(0x22), VotingPower: 25},
		},
	})
	require.NoError(t, err)

	require.Equal(t, operator(0x11), res.Network.Validators[0].OperatorAddress)
	require.Equal(t, uint32(5_000), res.Network.Validators[0].RawVotingPowerBps)
	require.Equal(t, uint32(300), res.Network.Validators[0].EffectiveVotingPowerBps)
	require.Equal(t, uint32(4_700), res.Network.Validators[0].OverflowVotingPowerBps())
	require.Equal(t, uint64(3), res.Network.Validators[0].RewardableVotingPower(res.Network.TotalVotingPower))
	require.Equal(t, uint64(47), res.Network.Validators[0].OverflowVotingPower(res.Network.TotalVotingPower))
	require.True(t, res.Network.Validators[0].AboveHardCap)
	require.False(t, res.Network.Validators[0].DelegationAllowed)

	allowed, err := app.StakeConcentrationKeeper.CanAcceptDelegation(ctx, operator(0x11))
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestAetraPowerCapExampleTracksExcessStake(t *testing.T) {
	params := types.DefaultParams()
	validatorSet := []types.ValidatorPower{
		{OperatorAddress: operator(0x11), VotingPower: 50},
		{OperatorAddress: operator(0x22), VotingPower: 950},
	}

	network, err := types.ComputeNetworkConcentration(params, 1, validatorSet, 10)
	require.NoError(t, err)

	validator, found := findConcentration(network.Validators, operator(0x11))
	require.True(t, found)
	require.Equal(t, uint64(1_000), network.TotalVotingPower)
	require.Equal(t, uint32(500), validator.RawVotingPowerBps)
	require.Equal(t, uint32(300), validator.EffectiveVotingPowerBps)
	require.Equal(t, uint32(200), validator.OverflowVotingPowerBps())
	require.Equal(t, uint64(30), validator.RewardableVotingPower(network.TotalVotingPower))
	require.Equal(t, uint64(20), validator.OverflowVotingPower(network.TotalVotingPower))
	require.True(t, validator.AboveHardCap)
	require.False(t, validator.DelegationAllowed)
	require.Equal(t, "hard_cap_exceeded", validator.Warning)
}

func TestRewardModifierApplies(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)

	res, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority:	app.StakeConcentrationKeeper.Authority(),
		Epoch:		1,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x11), VotingPower: 40},
			{OperatorAddress: operator(0x22), VotingPower: 30},
			{OperatorAddress: operator(0x33), VotingPower: 30},
		},
	})
	require.NoError(t, err)

	metric := res.Network.Validators[0]
	require.Equal(t, uint32(4_000), metric.RawVotingPowerBps)
	require.Equal(t, uint32(8_847), metric.RewardModifierBps)
	require.Less(t, metric.RewardModifierBps, types.BasisPoints)
	require.GreaterOrEqual(t, metric.RewardModifierBps, uint32(0))
}

func TestPowerCapEnforcedAcrossEpochTransition(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)

	first, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority:	app.StakeConcentrationKeeper.Authority(),
		Epoch:		1,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x11), VotingPower: 34},
			{OperatorAddress: operator(0x22), VotingPower: 33},
			{OperatorAddress: operator(0x33), VotingPower: 33},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(300), first.Network.MaxValidatorPowerBps)

	second, err := msgServer.RecomputeConcentration(ctx, &types.MsgRecomputeConcentration{
		Authority:	app.StakeConcentrationKeeper.Authority(),
		Epoch:		2,
		ValidatorSet: []types.ValidatorPower{
			{OperatorAddress: operator(0x11), VotingPower: 70},
			{OperatorAddress: operator(0x22), VotingPower: 20},
			{OperatorAddress: operator(0x33), VotingPower: 10},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(2), second.Network.Epoch)
	require.Equal(t, uint32(300), second.Network.MaxValidatorPowerBps)
	require.Equal(t, uint32(300), second.Network.Validators[0].EffectiveVotingPowerBps)
}

func TestAetraPowerCapSchedule(t *testing.T) {
	params := types.DefaultParams()

	require.Equal(t, uint32(300), types.EffectiveMaxVotingPowerBps(params, 100))
	require.Equal(t, uint32(300), types.EffectiveMaxVotingPowerBps(params, 150))
	require.Equal(t, uint32(250), types.EffectiveMaxVotingPowerBps(params, 151))
	require.Equal(t, uint32(250), types.EffectiveMaxVotingPowerBps(params, 250))
	require.Equal(t, uint32(200), types.EffectiveMaxVotingPowerBps(params, 251))
	require.Equal(t, uint32(200), types.EffectiveMaxVotingPowerBps(params, 300))

	params.MaxVotingPowerBps = 175
	require.Equal(t, uint32(175), types.EffectiveMaxVotingPowerBps(params, 100))
	require.Equal(t, uint32(175), types.EffectiveMaxVotingPowerBps(params, 300))
}

func TestNetworkConcentrationUsesScheduledPowerCap(t *testing.T) {
	params := types.DefaultParams()
	validatorSet := makeValidatorSet(251)
	validatorSet[0].VotingPower = 10_000

	network, err := types.ComputeNetworkConcentration(params, 1, validatorSet, 10)
	require.NoError(t, err)

	heavy, found := findConcentration(network.Validators, validatorSet[0].OperatorAddress)
	require.True(t, found)
	require.Greater(t, heavy.RawVotingPowerBps, uint32(200))
	require.Equal(t, uint32(200), heavy.EffectiveVotingPowerBps)
	require.Equal(t, uint32(200), network.MaxValidatorPowerBps)
	require.True(t, heavy.AboveHardCap)
	require.False(t, heavy.DelegationAllowed)
}

func TestConcentrationTargetsEmitEconomicSignalsWithoutHalting(t *testing.T) {
	params := types.DefaultParams()
	validatorSet := makeValidatorSet(40)
	for i := 0; i < 10; i++ {
		validatorSet[i].VotingPower = 3
	}

	network, err := types.ComputeNetworkConcentration(params, 1, validatorSet, 10)
	require.NoError(t, err)

	assessment := network.AssessConcentrationTargets()
	require.Equal(t, uint32(5_000), assessment.Top10VotingPowerBps)
	require.Equal(t, uint32(6_660), assessment.Top20VotingPowerBps)
	require.Equal(t, uint32(8_818), assessment.Top33VotingPowerBps)
	require.True(t, assessment.Top10Exceeded)
	require.True(t, assessment.Top20Exceeded)
	require.True(t, assessment.Top33Exceeded)
	require.ElementsMatch(t, []string{
		types.ConcentrationSignalLowerRewardMultiplier,
		types.ConcentrationSignalDelegationWarning,
		types.ConcentrationSignalProtocolMetric,
		types.ConcentrationSignalGovernanceAlert,
		types.ConcentrationSignalParameterProposal,
	}, assessment.Signals)
}

func TestConcentrationTargetsStayQuietWhenBelowThresholds(t *testing.T) {
	params := types.DefaultParams()
	validatorSet := makeValidatorSet(100)

	network, err := types.ComputeNetworkConcentration(params, 1, validatorSet, 10)
	require.NoError(t, err)

	assessment := network.AssessConcentrationTargets()
	require.Equal(t, uint32(1_000), assessment.Top10VotingPowerBps)
	require.Equal(t, uint32(2_000), assessment.Top20VotingPowerBps)
	require.Equal(t, uint32(3_300), assessment.Top33VotingPowerBps)
	require.False(t, assessment.Top10Exceeded)
	require.False(t, assessment.Top20Exceeded)
	require.False(t, assessment.Top33Exceeded)
	require.Empty(t, assessment.Signals)
}

func TestInvalidParamsRejected(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := stakeconcentrationkeeper.NewMsgServerImpl(app.StakeConcentrationKeeper)
	params := types.DefaultParams()
	params.SoftVotingPowerBps = params.MaxVotingPowerBps + 1

	_, err := msgServer.UpdateConcentrationParams(ctx, &types.MsgUpdateConcentrationParams{
		Authority:	app.StakeConcentrationKeeper.Authority(),
		Params:		params,
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
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bytes20(fill)))
}

func operatorN(n int) string {
	out := make([]byte, 20)
	out[18] = byte(n >> 8)
	out[19] = byte(n)
	return aetraaddress.FormatAccAddress(sdk.AccAddress(out))
}

func makeValidatorSet(count int) []types.ValidatorPower {
	out := make([]types.ValidatorPower, count)
	for i := range out {
		out[i] = types.ValidatorPower{OperatorAddress: operatorN(i + 1), VotingPower: 1}
	}
	return out
}

func findConcentration(validators []types.ValidatorConcentration, operatorAddress string) (types.ValidatorConcentration, bool) {
	for _, validator := range validators {
		if validator.OperatorAddress == operatorAddress {
			return validator, true
		}
	}
	return types.ValidatorConcentration{}, false
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
