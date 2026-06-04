package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	"github.com/sovereign-l1/l1/x/dex/types"
)

func fundAccount(t *testing.T, app *l1app.L1App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
}

func setupDexPool(t *testing.T) (*l1app.L1App, sdk.Context, types.MsgServer, sdk.AccAddress, uint64) {
	t.Helper()

	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	creator := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	fundAccount(t, app, ctx, creator, sdk.NewCoins(sdk.NewInt64Coin("uatom", 10_000)))
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	res, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin("uatom", 1_000),
		TokenB:  sdk.NewInt64Coin("uorb", 1_000),
	})
	require.NoError(t, err)

	return app, ctx, msgServer, creator, res.PoolId
}

func TestAddLiquidityRejectsMalformedMinShares(t *testing.T) {
	app, ctx, msgServer, depositor, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, depositor, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))

	_, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: depositor.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin("uorb", 100),
		MinShares: "not-an-int",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "min_shares")
}

func TestAddLiquidityRejectsCorruptedPoolStateWithoutPanic(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	depositor := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	fundAccount(t, app, ctx, depositor, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
		Id:          99,
		Denom0:      "uatom",
		Denom1:      "uorb",
		Reserve0:    "not-an-int",
		Reserve1:    "100",
		TotalShares: "100",
		LpDenom:     "lp/99",
	}))

	var err error
	require.NotPanics(t, func() {
		_, err = msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
			Depositor: depositor.String(),
			PoolId:    99,
			TokenA:    sdk.NewInt64Coin("uatom", 100),
			TokenB:    sdk.NewInt64Coin("uorb", 100),
			MinShares: "1",
		})
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool state")
}
