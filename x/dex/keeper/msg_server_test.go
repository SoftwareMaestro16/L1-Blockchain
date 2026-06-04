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
		TokenB:  sdk.NewInt64Coin("norb", 1_000),
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
		TokenB:    sdk.NewInt64Coin("norb", 100),
		MinShares: "not-an-int",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "min_shares")
}

func TestMsgUpdateParamsRejectsUnauthorizedAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: "orb1unauthorized",
		Params:    types.DefaultParams(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestMsgUpdateParamsRejectsInvalidParams(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)
	params := types.DefaultParams()
	params.SwapFeeBps = params.MaxSwapFeeBps + 1

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.DexKeeper.Authority(),
		Params:    params,
	})
	require.ErrorIs(t, err, types.ErrInvalidParams)
}

func TestMsgUpdateParamsDisablesPoolCreation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	creator := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)
	params := types.DefaultParams()
	params.PoolCreationEnabled = false

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.DexKeeper.Authority(),
		Params:    params,
	})
	require.NoError(t, err)

	_, err = msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin("uatom", 1_000),
		TokenB:  sdk.NewInt64Coin("norb", 1_000),
	})
	require.ErrorIs(t, err, types.ErrOperationDisabled)
}

func TestMsgUpdateParamsDisablesSwapsAndLiquidity(t *testing.T) {
	app, ctx, msgServer, creator, poolID := setupDexPool(t)
	params := types.DefaultParams()
	params.SwapsEnabled = false

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.DexKeeper.Authority(),
		Params:    params,
	})
	require.NoError(t, err)

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        creator.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: "norb",
		MinAmountOut:  "1",
	})
	require.ErrorIs(t, err, types.ErrOperationDisabled)

	params = types.DefaultParams()
	params.LiquidityEnabled = false
	_, err = msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.DexKeeper.Authority(),
		Params:    params,
	})
	require.NoError(t, err)

	_, err = msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: creator.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 1),
		TokenB:    sdk.NewInt64Coin("norb", 1),
		MinShares: "1",
	})
	require.ErrorIs(t, err, types.ErrOperationDisabled)
}

func TestCreatePoolRejectsDuplicatePairViaIndex(t *testing.T) {
	app, ctx, msgServer, creator, _ := setupDexPool(t)
	fundAccount(t, app, ctx, creator, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))

	_, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin("norb", 100),
		TokenB:  sdk.NewInt64Coin("uatom", 100),
	})
	require.ErrorIs(t, err, types.ErrInvalidPool)
}

func TestParamsQueryReturnsCurrentParamsAndRejectsNil(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	_, err := app.DexKeeper.Params(ctx, nil)
	require.Error(t, err)

	params := types.DefaultParams()
	params.SwapFeeBps = 100
	require.NoError(t, app.DexKeeper.SetParams(ctx, params))

	res, err := app.DexKeeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint32(100), res.Params.SwapFeeBps)
}

func TestAddLiquidityRejectsCorruptedPoolStateWithoutPanic(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	depositor := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	fundAccount(t, app, ctx, depositor, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
		Id:          99,
		Denom0:      "norb",
		Denom1:      "uatom",
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
			TokenB:    sdk.NewInt64Coin("norb", 100),
			MinShares: "1",
		})
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool state")
}

func TestPoolAccountingMatchesBankBalancesAndLPSupply(t *testing.T) {
	app, ctx, msgServer, creator, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, creator, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))

	_, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: creator.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin("norb", 100),
		MinShares: "1",
	})
	require.NoError(t, err)

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        creator.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: "norb",
		MinAmountOut:  "1",
	})
	require.NoError(t, err)

	pool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)

	reserve0, ok := sdkmath.NewIntFromString(pool.Reserve0)
	require.True(t, ok)
	reserve1, ok := sdkmath.NewIntFromString(pool.Reserve1)
	require.True(t, ok)
	totalShares, ok := sdkmath.NewIntFromString(pool.TotalShares)
	require.True(t, ok)

	moduleAddr := app.AccountKeeper.GetModuleAddress(types.ModuleName)
	require.NotNil(t, moduleAddr)
	require.Equal(t, sdk.NewCoin(pool.Denom0, reserve0), app.BankKeeper.GetBalance(ctx, moduleAddr, pool.Denom0))
	require.Equal(t, sdk.NewCoin(pool.Denom1, reserve1), app.BankKeeper.GetBalance(ctx, moduleAddr, pool.Denom1))
	require.Equal(t, sdk.NewCoin(pool.LpDenom, totalShares), app.BankKeeper.GetSupply(ctx, pool.LpDenom))
}
