package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func TestMsgUpdateParamsRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	_, err := msgServer.UpdateParams(ctx, nil)
	require.Error(t, err)
}

func TestMsgUpdateParamsRejectsUnauthorizedAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: "orb1unauthorized",
		Params:    types.DefaultParams(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestMsgUpdateParamsRejectsMalformedParams(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)
	params := types.DefaultParams()
	params.MinFeeAmount = "0"

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.FeesKeeper.Authority(),
		Params:    params,
	})
	require.Error(t, err)
}

func TestMsgUpdateParamsSyncsDistributionCommunityTax(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)
	params := types.DefaultParams()
	params.ValidatorRewardsRatio = "0.90"
	params.CommunityPoolRatio = "0.10"
	params.MinFeeAmount = "10"

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.FeesKeeper.Authority(),
		Params:    params,
	})
	require.NoError(t, err)

	stored, err := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, "10", stored.MinFeeAmount)

	distrParams, err := app.DistrKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.True(t, distrParams.CommunityTax.Equal(sdkmath.LegacyMustNewDecFromStr("0.10")))
}
