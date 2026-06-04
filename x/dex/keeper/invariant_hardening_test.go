package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	"github.com/sovereign-l1/l1/x/dex/types"
)

func assertPoolAccounting(t *testing.T, app *l1app.L1App, ctx sdk.Context, pool types.Pool) {
	t.Helper()

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

func currentPool(t *testing.T, app *l1app.L1App, ctx sdk.Context, poolID uint64) types.Pool {
	t.Helper()

	pool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	return pool
}

func TestCreatePoolRejectsDuplicatePairInEitherOrder(t *testing.T) {
	app, ctx, msgServer, creator, _ := setupDexPool(t)
	fundAccount(t, app, ctx, creator, sdk.NewCoins(sdk.NewInt64Coin("uatom", 2_000)))

	_, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin("norb", 1_000),
		TokenB:  sdk.NewInt64Coin("uatom", 1_000),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool already exists")
}

func TestCreatePoolRejectsSameDenomAndPreservesNextID(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	creator := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	nextBefore, err := app.DexKeeper.GetNextPoolID(ctx)
	require.NoError(t, err)
	_, err = msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin("norb", 1_000),
		TokenB:  sdk.NewInt64Coin("norb", 1_000),
	})
	require.Error(t, err)

	nextAfter, err := app.DexKeeper.GetNextPoolID(ctx)
	require.NoError(t, err)
	require.Equal(t, nextBefore, nextAfter)
}

func TestAddLiquidityRejectsUnbalancedDeposit(t *testing.T) {
	app, ctx, msgServer, depositor, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, depositor, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000), sdk.NewInt64Coin("norb", 1_000)))
	before := currentPool(t, app, ctx, poolID)

	_, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: depositor.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin("norb", 101),
		MinShares: "1",
	})
	require.Error(t, err)

	after := currentPool(t, app, ctx, poolID)
	require.Equal(t, before, after)
	assertPoolAccounting(t, app, ctx, after)
}

func TestOperationsRejectReserveBalanceDesync(t *testing.T) {
	app, ctx, msgServer, user, poolID := setupDexPool(t)
	pool := currentPool(t, app, ctx, poolID)
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user, sdk.NewCoins(sdk.NewInt64Coin(pool.Denom0, 1))))
	fundAccount(t, app, ctx, user, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000), sdk.NewInt64Coin("norb", 1_000)))

	_, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: user.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin("norb", 100),
		MinShares: "1",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "module balance")

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        user.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: "norb",
		MinAmountOut:  "1",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "module balance")
}

func TestOperationsRejectLPSupplyDesync(t *testing.T) {
	app, ctx, msgServer, user, poolID := setupDexPool(t)
	pool := currentPool(t, app, ctx, poolID)
	require.NoError(t, app.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(pool.LpDenom, 1))))

	_, err := msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: user.String(),
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin(pool.LpDenom, 1),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "LP supply")
}

func TestRemoveLiquidityRejectsFullPoolDrain(t *testing.T) {
	app, ctx, msgServer, user, poolID := setupDexPool(t)
	pool := currentPool(t, app, ctx, poolID)

	_, err := msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: user.String(),
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin(pool.LpDenom, 1_000),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot remove all liquidity")

	assertPoolAccounting(t, app, ctx, currentPool(t, app, ctx, poolID))
}

func TestSwapRejectsTinyInputThatRoundsToZero(t *testing.T) {
	app, ctx, msgServer, user, poolID := setupDexPool(t)
	before := currentPool(t, app, ctx, poolID)

	_, err := msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        user.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 1),
		TokenOutDenom: "norb",
		MinAmountOut:  "0",
	})
	require.Error(t, err)

	after := currentPool(t, app, ctx, poolID)
	require.Equal(t, before, after)
	assertPoolAccounting(t, app, ctx, after)
}

func TestSwapKeepsConstantProductAndAccounting(t *testing.T) {
	app, ctx, msgServer, trader, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("uatom", 10_000)))

	before := currentPool(t, app, ctx, poolID)
	beforeReserve0, ok := sdkmath.NewIntFromString(before.Reserve0)
	require.True(t, ok)
	beforeReserve1, ok := sdkmath.NewIntFromString(before.Reserve1)
	require.True(t, ok)
	beforeK := beforeReserve0.Mul(beforeReserve1)

	res, err := msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 500),
		TokenOutDenom: "norb",
		MinAmountOut:  "1",
	})
	require.NoError(t, err)
	require.True(t, res.TokenOut.IsPositive())

	after := currentPool(t, app, ctx, poolID)
	afterReserve0, ok := sdkmath.NewIntFromString(after.Reserve0)
	require.True(t, ok)
	afterReserve1, ok := sdkmath.NewIntFromString(after.Reserve1)
	require.True(t, ok)
	require.True(t, afterReserve0.Mul(afterReserve1).GTE(beforeK))
	assertPoolAccounting(t, app, ctx, after)
}

func TestInsufficientFundsDoesNotMutatePool(t *testing.T) {
	app, ctx, msgServer, user, poolID := setupDexPool(t)
	before := currentPool(t, app, ctx, poolID)

	_, err := msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        user.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 1_000_000_000),
		TokenOutDenom: "norb",
		MinAmountOut:  "1",
	})
	require.Error(t, err)
	require.Equal(t, before, currentPool(t, app, ctx, poolID))
}

func TestOperationSequencePreservesPoolAccounting(t *testing.T) {
	app, ctx, msgServer, user, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, user, sdk.NewCoins(sdk.NewInt64Coin("uatom", 10_000), sdk.NewInt64Coin("norb", 10_000)))
	assertPoolAccounting(t, app, ctx, currentPool(t, app, ctx, poolID))

	_, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: user.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 500),
		TokenB:    sdk.NewInt64Coin("norb", 500),
		MinShares: "1",
	})
	require.NoError(t, err)
	assertPoolAccounting(t, app, ctx, currentPool(t, app, ctx, poolID))

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        user.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 250),
		TokenOutDenom: "norb",
		MinAmountOut:  "1",
	})
	require.NoError(t, err)
	assertPoolAccounting(t, app, ctx, currentPool(t, app, ctx, poolID))

	pool := currentPool(t, app, ctx, poolID)
	_, err = msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: user.String(),
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin(pool.LpDenom, 100),
	})
	require.NoError(t, err)
	assertPoolAccounting(t, app, ctx, currentPool(t, app, ctx, poolID))
}
