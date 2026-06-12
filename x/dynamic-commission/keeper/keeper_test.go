package keeper_test

import (
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	dynamiccommissionkeeper "github.com/sovereign-l1/l1/x/dynamic-commission/keeper"
	"github.com/sovereign-l1/l1/x/dynamic-commission/types"
)

func TestHighPerformanceBonus(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dynamiccommissionkeeper.NewMsgServerImpl(app.DynamicCommissionKeeper)
	validator := validatorAddress()

	_, err := msgServer.SetBaseCommission(ctx, &types.MsgSetBaseCommission{
		ValidatorAddress:	validator,
		BaseCommissionBps:	1_000,
		Height:			1,
	})
	require.NoError(t, err)
	res, err := msgServer.RecomputeEffectiveCommission(ctx, &types.MsgRecomputeEffectiveCommission{
		Authority:		app.DynamicCommissionKeeper.Authority(),
		ValidatorAddress:	validator,
		PerformanceScoreBps:	9_500,
		ReputationScoreBps:	6_000,
		Height:			2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(1_100), res.Commission.EffectiveCommissionBps)
	require.Equal(t, int32(100), res.Commission.PerformanceModifierBps)
	require.Equal(t, int32(0), res.Commission.ReputationModifierBps)
}

func TestDefaultCommissionParamsMatchAetraAntiCartelPolicy(t *testing.T) {
	params := types.DefaultParams()

	require.Equal(t, uint32(300), params.CommissionFloorBps)
	require.Equal(t, uint32(2_000), params.CommissionCeilingBps)
	require.Equal(t, uint32(100), params.MaxRateChangeBps)
	require.NoError(t, params.Validate())
}

func TestLowPerformancePenalty(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dynamiccommissionkeeper.NewMsgServerImpl(app.DynamicCommissionKeeper)
	validator := validatorAddress()

	_, err := msgServer.SetBaseCommission(ctx, &types.MsgSetBaseCommission{
		ValidatorAddress:	validator,
		BaseCommissionBps:	1_000,
		Height:			1,
	})
	require.NoError(t, err)
	res, err := msgServer.RecomputeEffectiveCommission(ctx, &types.MsgRecomputeEffectiveCommission{
		Authority:		app.DynamicCommissionKeeper.Authority(),
		ValidatorAddress:	validator,
		PerformanceScoreBps:	4_000,
		ReputationScoreBps:	6_000,
		Height:			2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(900), res.Commission.EffectiveCommissionBps)
	require.Equal(t, int32(-100), res.Commission.PerformanceModifierBps)
}

func TestFloorCeilingAndJailedBonusRules(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dynamiccommissionkeeper.NewMsgServerImpl(app.DynamicCommissionKeeper)
	validator := validatorAddress()

	params := types.DefaultParams()
	params.CommissionFloorBps = 900
	params.CommissionCeilingBps = 1_100
	_, err := msgServer.UpdateCommissionParams(ctx, &types.MsgUpdateCommissionParams{
		Authority:	app.DynamicCommissionKeeper.Authority(),
		Params:		params,
	})
	require.NoError(t, err)
	_, err = msgServer.SetBaseCommission(ctx, &types.MsgSetBaseCommission{
		ValidatorAddress:	validator,
		BaseCommissionBps:	1_000,
		Height:			1,
	})
	require.NoError(t, err)

	ceiling, err := msgServer.RecomputeEffectiveCommission(ctx, &types.MsgRecomputeEffectiveCommission{
		Authority:		app.DynamicCommissionKeeper.Authority(),
		ValidatorAddress:	validator,
		PerformanceScoreBps:	9_500,
		ReputationScoreBps:	9_000,
		Height:			2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(1_100), ceiling.Commission.EffectiveCommissionBps)

	floorValidator := validatorAddressN(2)
	_, err = msgServer.SetBaseCommission(ctx, &types.MsgSetBaseCommission{
		ValidatorAddress:	floorValidator,
		BaseCommissionBps:	1_000,
		Height:			1,
	})
	require.NoError(t, err)
	floor, err := msgServer.RecomputeEffectiveCommission(ctx, &types.MsgRecomputeEffectiveCommission{
		Authority:		app.DynamicCommissionKeeper.Authority(),
		ValidatorAddress:	floorValidator,
		PerformanceScoreBps:	1_000,
		ReputationScoreBps:	1_000,
		Height:			3,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(900), floor.Commission.EffectiveCommissionBps)

	jailed, err := msgServer.RecomputeEffectiveCommission(ctx, &types.MsgRecomputeEffectiveCommission{
		Authority:		app.DynamicCommissionKeeper.Authority(),
		ValidatorAddress:	floorValidator,
		PerformanceScoreBps:	9_500,
		ReputationScoreBps:	6_000,
		Jailed:			true,
		Height:			4,
	})
	require.NoError(t, err)
	require.Equal(t, int32(0), jailed.Commission.PerformanceModifierBps)
	require.Equal(t, uint32(1_000), jailed.Commission.EffectiveCommissionBps)
}

func TestRateLimitEnforcedWithoutMutation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dynamiccommissionkeeper.NewMsgServerImpl(app.DynamicCommissionKeeper)
	validator := validatorAddress()

	_, err := msgServer.SetBaseCommission(ctx, &types.MsgSetBaseCommission{
		ValidatorAddress:	validator,
		BaseCommissionBps:	1_000,
		Height:			1,
	})
	require.NoError(t, err)
	_, err = msgServer.SetBaseCommission(ctx, &types.MsgSetBaseCommission{
		ValidatorAddress:	validator,
		BaseCommissionBps:	1_800,
		Height:			2,
	})
	require.ErrorIs(t, err, types.ErrRateLimited)

	commission, found, getErr := app.DynamicCommissionKeeper.GetValidatorCommission(ctx, validator)
	require.NoError(t, getErr)
	require.True(t, found)
	require.Equal(t, uint32(1_000), commission.EffectiveCommissionBps)
}

func TestExportImportPreservesEffectiveCommission(t *testing.T) {
	source := l1app.Setup(t, false)
	ctx := source.NewContext(false)
	msgServer := dynamiccommissionkeeper.NewMsgServerImpl(source.DynamicCommissionKeeper)
	validator := validatorAddress()

	_, err := msgServer.SetBaseCommission(ctx, &types.MsgSetBaseCommission{
		ValidatorAddress:	validator,
		BaseCommissionBps:	1_000,
		Height:			1,
	})
	require.NoError(t, err)
	_, err = msgServer.RecomputeEffectiveCommission(ctx, &types.MsgRecomputeEffectiveCommission{
		Authority:		source.DynamicCommissionKeeper.Authority(),
		ValidatorAddress:	validator,
		PerformanceScoreBps:	9_500,
		ReputationScoreBps:	9_000,
		Height:			2,
	})
	require.NoError(t, err)

	exported, err := source.DynamicCommissionKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := l1app.Setup(t, false)
	targetCtx := target.NewContext(false)
	require.NoError(t, target.DynamicCommissionKeeper.InitGenesis(targetCtx, *exported))

	imported, found, err := target.DynamicCommissionKeeper.GetValidatorCommission(targetCtx, validator)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint32(1_100), imported.EffectiveCommissionBps)
	history, err := target.DynamicCommissionKeeper.GetCommissionHistory(targetCtx, validator)
	require.NoError(t, err)
	require.Len(t, history, 2)
}

func validatorAddress() string {
	return aetraaddress.FormatAccAddress(simtestutil.CreateIncrementalAccounts(1)[0])
}

func validatorAddressN(n int) string {
	return aetraaddress.FormatAccAddress(simtestutil.CreateIncrementalAccounts(n)[n-1])
}
