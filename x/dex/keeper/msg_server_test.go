package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	"github.com/sovereign-l1/l1/x/dex/types"
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

func fundAccount(t testing.TB, app *l1app.L1App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
}

func mustIntFromString(t *testing.T, value string) sdkmath.Int {
	t.Helper()
	out, ok := sdkmath.NewIntFromString(value)
	require.True(t, ok)
	return out
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
		TokenB:  sdk.NewInt64Coin(appparams.BaseDenom, 1_000),
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
		TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 100),
		MinShares: "not-an-int",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "min_shares")
}

func TestDexRejectsZeroActorAddress(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	_, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: aetherisaddress.ZeroRawAddress,
		TokenA:  sdk.NewInt64Coin("uatom", 1_000),
		TokenB:  sdk.NewInt64Coin(appparams.BaseDenom, 1_000),
	})
	require.ErrorIs(t, err, types.ErrInvalidAddress)
	require.Contains(t, err.Error(), "creator must not be zero address")

	app, ctx, msgServer, _, poolID := setupDexPool(t)
	pool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)

	_, err = msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: aetherisaddress.ZeroRawAddress,
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 1),
		TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 1),
		MinShares: "1",
	})
	require.ErrorIs(t, err, types.ErrInvalidAddress)
	require.Contains(t, err.Error(), "depositor must not be zero address")

	_, err = msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: aetherisaddress.ZeroUserFriendly,
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin(pool.LpDenom, 1),
	})
	require.ErrorIs(t, err, types.ErrInvalidAddress)
	require.Contains(t, err.Error(), "withdrawer must not be zero address")

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        aetherisaddress.ZeroRawAddress,
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 1),
		TokenOutDenom: appparams.BaseDenom,
		MinAmountOut:  "1",
	})
	require.ErrorIs(t, err, types.ErrInvalidAddress)
	require.Contains(t, err.Error(), "trader must not be zero address")
}

func TestCreatePoolRejectsDuplicatePair(t *testing.T) {
	app, ctx, msgServer, creator, poolID := setupDexPool(t)

	existingID, found, err := app.DexKeeper.GetPoolIDByPair(ctx, appparams.BaseDenom, "uatom")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, poolID, existingID)

	_, err = msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin(appparams.BaseDenom, 10),
		TokenB:  sdk.NewInt64Coin("uatom", 10),
	})
	require.ErrorIs(t, err, types.ErrInvalidPool)
	require.Contains(t, err.Error(), "pool already exists")

	nextID, err := app.DexKeeper.GetNextPoolID(ctx)
	require.NoError(t, err)
	require.Equal(t, poolID+1, nextID)
}

func TestDexLifecycleEmitsStableEventsAndStateQueries(t *testing.T) {
	app, ctx, msgServer, trader, poolID := setupDexPool(t)
	traderText := aetherisaddress.FormatAccAddress(trader)
	pool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	requireEvent(t, ctx, types.EventTypeCreatePool, map[string]string{
		types.AttributeKeyPoolID:       "1",
		types.AttributeKeyCreator:      traderText,
		types.AttributeKeyDenom0:       pool.Denom0,
		types.AttributeKeyDenom1:       pool.Denom1,
		types.AttributeKeyAmount0:      "1000",
		types.AttributeKeyAmount1:      "1000",
		types.AttributeKeyLPDenom:      pool.LpDenom,
		types.AttributeKeyMintedShares: "1000",
	})
	require.Equal(t, sdk.NewInt64Coin(pool.LpDenom, 1_000), app.BankKeeper.GetBalance(ctx, trader, pool.LpDenom))

	fundAccount(t, app, ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))
	addRes, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: trader.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 100),
		MinShares: "1",
	})
	require.NoError(t, err)
	requireEvent(t, ctx, types.EventTypeAddLiquidity, map[string]string{
		types.AttributeKeyPoolID:       "1",
		types.AttributeKeyDepositor:    traderText,
		types.AttributeKeyDenom0:       pool.Denom0,
		types.AttributeKeyDenom1:       pool.Denom1,
		types.AttributeKeyAmount0:      "100",
		types.AttributeKeyAmount1:      "100",
		types.AttributeKeyLPDenom:      pool.LpDenom,
		types.AttributeKeyMintedShares: addRes.MintedShares.Amount.String(),
	})
	pool, found, err = app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "1100", pool.TotalShares)

	swapRes, err := msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: appparams.BaseDenom,
		MinAmountOut:  "1",
	})
	require.NoError(t, err)
	requireEvent(t, ctx, types.EventTypeSwapExactAmountIn, map[string]string{
		types.AttributeKeyPoolID:   "1",
		types.AttributeKeyTrader:   traderText,
		types.AttributeKeyTokenIn:  "10uatom",
		types.AttributeKeyTokenOut: swapRes.TokenOut.String(),
	})
	require.True(t, app.BankKeeper.GetBalance(ctx, trader, appparams.BaseDenom).Amount.IsPositive())

	removeRes, err := msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: trader.String(),
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin(pool.LpDenom, 50),
	})
	require.NoError(t, err)
	requireEvent(t, ctx, types.EventTypeRemoveLiquidity, map[string]string{
		types.AttributeKeyPoolID:     "1",
		types.AttributeKeyWithdrawer: traderText,
		types.AttributeKeyLPDenom:    pool.LpDenom,
		types.AttributeKeyShares:     "50",
		types.AttributeKeyDenom0:     removeRes.TokenA.Denom,
		types.AttributeKeyDenom1:     removeRes.TokenB.Denom,
		types.AttributeKeyAmount0:    removeRes.TokenA.Amount.String(),
		types.AttributeKeyAmount1:    removeRes.TokenB.Amount.String(),
	})
	updatedPool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoin(updatedPool.LpDenom, mustIntFromString(t, updatedPool.TotalShares)), app.BankKeeper.GetSupply(ctx, updatedPool.LpDenom))
}

func TestInitGenesisRestoresPairIndex(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	app.DexKeeper.InitGenesis(ctx, types.GenesisState{
		NextPoolId: 2,
		Pools: []types.Pool{
			{
				Id:          1,
				Denom0:      appparams.BaseDenom,
				Denom1:      "uatom",
				Reserve0:    "100",
				Reserve1:    "100",
				TotalShares: "100",
				LpDenom:     "lp/1",
			},
		},
	})

	poolID, found, err := app.DexKeeper.GetPoolIDByPair(ctx, appparams.BaseDenom, "uatom")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(1), poolID)
}

func TestAddLiquidityRejectsCorruptedPoolStateWithoutPanic(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	depositor := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	fundAccount(t, app, ctx, depositor, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
		Id:          99,
		Denom0:      appparams.BaseDenom,
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
			TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 100),
			MinShares: "1",
		})
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool state")
}

func TestSwapRejectsCorruptedPoolStateWithoutPanic(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	trader := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	fundAccount(t, app, ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
		Id:          99,
		Denom0:      appparams.BaseDenom,
		Denom1:      "uatom",
		Reserve0:    "100",
		Reserve1:    "100",
		TotalShares: "100",
		LpDenom:     "lp/2",
	}))

	var err error
	require.NotPanics(t, func() {
		_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
			Trader:        trader.String(),
			PoolId:        99,
			TokenIn:       sdk.NewInt64Coin("uatom", 10),
			TokenOutDenom: appparams.BaseDenom,
			MinAmountOut:  "1",
		})
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid lp denom")
}

func TestPoolAccountingMatchesBankBalancesAndLPSupply(t *testing.T) {
	app, ctx, msgServer, creator, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, creator, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))

	_, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: creator.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 100),
		MinShares: "1",
	})
	require.NoError(t, err)

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        creator.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: appparams.BaseDenom,
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

func TestAddLiquidityInsufficientFundsLeavesPoolAndBalancesUnchanged(t *testing.T) {
	app, ctx, msgServer, depositor, poolID := setupDexPool(t)

	poolBefore, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	lpBefore := app.BankKeeper.GetBalance(ctx, depositor, poolBefore.LpDenom)

	_, err = msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: depositor.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 1_000_000),
		TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 100),
		MinShares: "1",
	})
	require.Error(t, err)

	poolAfter, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, poolBefore, poolAfter)
	require.Equal(t, lpBefore, app.BankKeeper.GetBalance(ctx, depositor, poolBefore.LpDenom))
}

func TestRemoveLiquidityReserveMismatchDoesNotBurnSharesOrUpdatePool(t *testing.T) {
	app, ctx, msgServer, withdrawer, poolID := setupDexPool(t)

	poolBefore, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
		Id:          poolBefore.Id,
		Denom0:      poolBefore.Denom0,
		Denom1:      poolBefore.Denom1,
		Reserve0:    "1000000000",
		Reserve1:    "1000000000",
		TotalShares: poolBefore.TotalShares,
		LpDenom:     poolBefore.LpDenom,
	}))
	corruptedPool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	lpBefore := app.BankKeeper.GetBalance(ctx, withdrawer, poolBefore.LpDenom)
	supplyBefore := app.BankKeeper.GetSupply(ctx, poolBefore.LpDenom)

	_, err = msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: withdrawer.String(),
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin(poolBefore.LpDenom, 10),
	})
	require.Error(t, err)

	poolAfter, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, corruptedPool, poolAfter)
	require.Equal(t, lpBefore, app.BankKeeper.GetBalance(ctx, withdrawer, poolBefore.LpDenom))
	require.Equal(t, supplyBefore, app.BankKeeper.GetSupply(ctx, poolBefore.LpDenom))
}

func TestRemoveLiquidityRejectsFullPoolWithdrawal(t *testing.T) {
	app, ctx, msgServer, withdrawer, poolID := setupDexPool(t)

	poolBefore, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	lpBefore := app.BankKeeper.GetBalance(ctx, withdrawer, poolBefore.LpDenom)
	supplyBefore := app.BankKeeper.GetSupply(ctx, poolBefore.LpDenom)

	_, err = msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: withdrawer.String(),
		PoolId:     poolID,
		Shares:     sdk.NewCoin(poolBefore.LpDenom, mustIntFromString(t, poolBefore.TotalShares)),
	})
	require.ErrorIs(t, err, types.ErrInvalidLiquidity)
	require.Contains(t, err.Error(), "cannot remove all LP shares")

	poolAfter, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, poolBefore, poolAfter)
	require.Equal(t, lpBefore, app.BankKeeper.GetBalance(ctx, withdrawer, poolBefore.LpDenom))
	require.Equal(t, supplyBefore, app.BankKeeper.GetSupply(ctx, poolBefore.LpDenom))
}

func TestDexLifecycleRejectsSlippageAndWrongDenoms(t *testing.T) {
	app, ctx, msgServer, trader, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))

	pool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)

	_, err = msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: trader.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 100),
		MinShares: "101",
	})
	require.ErrorIs(t, err, types.ErrSlippage)

	unchangedPool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, pool, unchangedPool)

	addRes, err := msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
		Depositor: trader.String(),
		PoolId:    poolID,
		TokenA:    sdk.NewInt64Coin("uatom", 100),
		TokenB:    sdk.NewInt64Coin(appparams.BaseDenom, 100),
		MinShares: "100",
	})
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(pool.LpDenom, 100), addRes.MintedShares)

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 1),
		TokenOutDenom: appparams.BaseDenom,
		MinAmountOut:  "0",
	})
	require.ErrorIs(t, err, types.ErrSlippage)
	require.Contains(t, err.Error(), "amount out below minimum")

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: "bad",
		MinAmountOut:  "1",
	})
	require.ErrorIs(t, err, types.ErrInvalidPool)

	naetBeforeSwap := app.BankKeeper.GetBalance(ctx, trader, appparams.BaseDenom)
	swapRes, err := msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: appparams.BaseDenom,
		MinAmountOut:  "1",
	})
	require.NoError(t, err)
	require.True(t, swapRes.TokenOut.Amount.IsPositive())
	require.True(t, app.BankKeeper.GetBalance(ctx, trader, appparams.BaseDenom).Amount.GT(naetBeforeSwap.Amount))

	_, err = msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: trader.String(),
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin("lp/999", 10),
	})
	require.ErrorIs(t, err, types.ErrInvalidLiquidity)

	lpBeforeRemove := app.BankKeeper.GetBalance(ctx, trader, pool.LpDenom)
	removeRes, err := msgServer.RemoveLiquidity(ctx, &types.MsgRemoveLiquidity{
		Withdrawer: trader.String(),
		PoolId:     poolID,
		Shares:     sdk.NewInt64Coin(pool.LpDenom, 50),
	})
	require.NoError(t, err)
	require.True(t, removeRes.TokenA.IsPositive())
	require.True(t, removeRes.TokenB.IsPositive())
	require.Equal(t, lpBeforeRemove.Amount.SubRaw(50), app.BankKeeper.GetBalance(ctx, trader, pool.LpDenom).Amount)

	updatedPool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoin(updatedPool.LpDenom, mustIntFromString(t, updatedPool.TotalShares)), app.BankKeeper.GetSupply(ctx, updatedPool.LpDenom))
}

func TestCreatePoolLPDenomCannotSpoofNativeToken(t *testing.T) {
	app, ctx, _, _, poolID := setupDexPool(t)

	pool, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "lp/1", pool.LpDenom)
	require.NotEqual(t, appparams.BaseDenom, pool.LpDenom)
	require.NotEqual(t, appparams.DisplayDenom, pool.LpDenom)
	require.NotEqual(t, appparams.TokenSymbol, pool.LpDenom)
}

func TestCreatePoolRejectsNativeSpoofingThroughPoolDenoms(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	creator := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	fundAccount(t, app, ctx, creator, sdk.NewCoins(
		sdk.NewInt64Coin("uatom", 10_000),
		sdk.NewInt64Coin(appparams.DisplayDenom, 10_000),
	))
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	tests := []sdk.Coin{
		sdk.NewInt64Coin(appparams.DisplayDenom, 1_000),
		sdk.NewInt64Coin(appparams.TokenName, 1_000),
		sdk.NewInt64Coin("factory/"+creator.String()+"/"+appparams.BaseDenom, 1_000),
		sdk.NewInt64Coin("factory/"+creator.String()+"/"+appparams.DisplayDenom, 1_000),
	}

	for _, token := range tests {
		t.Run(token.Denom, func(t *testing.T) {
			nextIDBefore, err := app.DexKeeper.GetNextPoolID(ctx)
			require.NoError(t, err)
			_, err = msgServer.CreatePool(ctx, &types.MsgCreatePool{
				Creator: creator.String(),
				TokenA:  token,
				TokenB:  sdk.NewInt64Coin("uatom", 1_000),
			})
			require.ErrorIs(t, err, types.ErrInvalidPool)
			require.Contains(t, err.Error(), "native AET/naet")
			nextIDAfter, err := app.DexKeeper.GetNextPoolID(ctx)
			require.NoError(t, err)
			require.Equal(t, nextIDBefore, nextIDAfter)
		})
	}
}

func TestSwapPreservesConstantProductAndAccountingInvariants(t *testing.T) {
	app, ctx, msgServer, trader, poolID := setupDexPool(t)
	fundAccount(t, app, ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000)))

	before, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	beforeReserve0 := mustIntFromString(t, before.Reserve0)
	beforeReserve1 := mustIntFromString(t, before.Reserve1)
	beforeProduct := beforeReserve0.Mul(beforeReserve1)

	_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 100),
		TokenOutDenom: appparams.BaseDenom,
		MinAmountOut:  "1",
	})
	require.NoError(t, err)

	after, found, err := app.DexKeeper.GetPool(ctx, poolID)
	require.NoError(t, err)
	require.True(t, found)
	afterReserve0 := mustIntFromString(t, after.Reserve0)
	afterReserve1 := mustIntFromString(t, after.Reserve1)
	require.True(t, afterReserve0.Mul(afterReserve1).GTE(beforeProduct))

	moduleAddr := app.AccountKeeper.GetModuleAddress(types.ModuleName)
	require.NotNil(t, moduleAddr)
	require.Equal(t, sdk.NewCoin(after.Denom0, afterReserve0), app.BankKeeper.GetBalance(ctx, moduleAddr, after.Denom0))
	require.Equal(t, sdk.NewCoin(after.Denom1, afterReserve1), app.BankKeeper.GetBalance(ctx, moduleAddr, after.Denom1))
	require.Equal(t, sdk.NewCoin(after.LpDenom, mustIntFromString(t, after.TotalShares)), app.BankKeeper.GetSupply(ctx, after.LpDenom))
}

func TestUpdateParamsRejectsZeroAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: aetherisaddress.ZeroRawAddress,
		Params:    types.DefaultParams(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.Contains(t, err.Error(), "authority must not be zero address")
}
