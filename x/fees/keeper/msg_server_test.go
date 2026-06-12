package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	l1testutil "github.com/sovereign-l1/l1/tests/testutil"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func requireEvent(t *testing.T, ctx sdk.Context, eventType string, attrs map[string]string) {
	t.Helper()
	for _, event := range ctx.EventManager().Events() {
		if event.Type != eventType {
			continue
		}
		for key, expected := range attrs {
			attr, found := event.GetAttribute(key)
			require.Truef(t, found, "event %s missing attribute %s", eventType, key)
			require.Equal(t, expected, attr.Value)
		}
		return
	}
	require.Failf(t, "missing event", "event type %s not emitted", eventType)
}

func requireNoEvent(t *testing.T, ctx sdk.Context, eventType string) {
	t.Helper()
	for _, event := range ctx.EventManager().Events() {
		require.NotEqual(t, eventType, event.Type)
	}
}

func TestUpdateParamsRequiresAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority:	"ae1notgov",
		Params:		types.DefaultParams(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, types.DefaultParams(), params)
	requireNoEvent(t, ctx, types.EventTypeUpdateParams)
}

func TestUpdateParamsRejectsZeroAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority:	aetraaddress.ZeroRawAddress,
		Params:		types.DefaultParams(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.Contains(t, err.Error(), "authority must not be zero address")
}

func TestUpdateParamsAcceptsGovernanceAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority:	app.FeesKeeper.Authority(),
		Params:		types.DefaultParams(),
	})
	require.NoError(t, err)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, types.DefaultParams(), params)
	requireEvent(t, ctx, types.EventTypeUpdateParams, map[string]string{
		types.AttributeKeyAuthority:			app.FeesKeeper.Authority(),
		types.AttributeKeyAllowedFeeDenom:		types.DefaultParams().AllowedFeeDenoms[0],
		types.AttributeKeyValidatorRewardsRatio:	types.DefaultParams().ValidatorRewardsRatio,
		types.AttributeKeyCommunityPoolRatio:		types.DefaultParams().CommunityPoolRatio,
	})
}

func TestUpdateParamsAcceptsBoundedGovernanceFeePolicyAndSyncsDistribution(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	next := types.DefaultParams()
	next.MinFeeAmount = "42"
	next.BaseFeeAmount = "42"
	next.ValidatorRewardsRatio = "0.90"
	next.CommunityPoolRatio = "0.10"

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority:	app.FeesKeeper.Authority(),
		Params:		next,
	})
	require.NoError(t, err)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, next, params)

	distrParams, getErr := app.DistrKeeper.Params.Get(ctx)
	require.NoError(t, getErr)
	require.Equal(t, sdkmath.LegacyMustNewDecFromStr("0.10"), distrParams.CommunityTax)
}

func TestUpdateParamsRejectsInvalidParamsWithoutMutatingState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	invalid := types.DefaultParams()
	invalid.AllowedFeeDenoms = []string{l1testutil.TestAssetDenom}

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority:	app.FeesKeeper.Authority(),
		Params:		invalid,
	})
	require.ErrorIs(t, err, types.ErrInvalidParams)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, types.DefaultParams(), params)
	requireNoEvent(t, ctx, types.EventTypeUpdateParams)
}

func TestUpdateParamsRejectsOutOfBoundsMinFeeWithoutMutatingState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	invalid := types.DefaultParams()
	invalid.MinFeeAmount = "1000000000000000001"

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority:	app.FeesKeeper.Authority(),
		Params:		invalid,
	})
	require.ErrorIs(t, err, types.ErrInvalidParams)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, types.DefaultParams(), params)
	requireNoEvent(t, ctx, types.EventTypeUpdateParams)
}
