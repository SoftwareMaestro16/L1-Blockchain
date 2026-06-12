package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	burnkeeper "github.com/sovereign-l1/l1/x/burn/keeper"
	"github.com/sovereign-l1/l1/x/burn/types"
)

func TestUserBurnReducesBalanceSupplyAndRecordsCounters(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(1_000)))[0]
	msgServer := burnkeeper.NewMsgServerImpl(app.BurnKeeper)
	supplyBefore := app.BankKeeper.GetSupply(ctx, types.BaseDenom)
	userBefore := app.BankKeeper.GetBalance(ctx, user, types.BaseDenom)

	res, err := msgServer.BurnUserCoins(ctx, &types.MsgBurnUserCoins{
		Burner:	aetraaddress.FormatAccAddress(user),
		Amount:	sdk.NewCoins(coin(125)),
		Epoch:	7,
		Reason:	"user-opt-in",
	})
	require.NoError(t, err)

	require.Equal(t, uint64(1), res.Burn.Id)
	require.False(t, res.Burn.Protocol)
	require.Equal(t, userBefore.Amount.Sub(sdkmath.NewInt(125)), app.BankKeeper.GetBalance(ctx, user, types.BaseDenom).Amount)
	require.Equal(t, supplyBefore.Amount.Sub(sdkmath.NewInt(125)), app.BankKeeper.GetSupply(ctx, types.BaseDenom).Amount)
	require.Equal(t, sdk.NewCoins(), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName)))

	byDenom, found, err := app.BurnKeeper.GetBurnedDenomEntry(ctx, types.BaseDenom)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoins(coin(125)), byDenom.Amount)

	byEpoch, found, err := app.BurnKeeper.GetBurnedEpochEntry(ctx, 7)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoins(coin(125)), byEpoch.Amount)
}

func TestProtocolBurnReducesModuleBalanceSupplyAndRecordsReason(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(coin(500))))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, sdk.NewCoins(coin(500))))
	msgServer := burnkeeper.NewMsgServerImpl(app.BurnKeeper)
	supplyBefore := app.BankKeeper.GetSupply(ctx, types.BaseDenom)
	moduleAddr := app.AccountKeeper.GetModuleAddress(types.ModuleName)
	moduleBefore := app.BankKeeper.GetBalance(ctx, moduleAddr, types.BaseDenom)

	res, err := msgServer.BurnProtocolCoins(ctx, &types.MsgBurnProtocolCoins{
		Authority:	app.BurnKeeper.Authority(),
		SourceModule:	types.ModuleName,
		Amount:		sdk.NewCoins(coin(200)),
		Epoch:		9,
		Reason:		"protocol-deflation",
	})
	require.NoError(t, err)

	require.True(t, res.Burn.Protocol)
	require.Equal(t, types.ModuleName, res.Burn.SourceModule)
	require.Equal(t, moduleBefore.Amount.Sub(sdkmath.NewInt(200)), app.BankKeeper.GetBalance(ctx, moduleAddr, types.BaseDenom).Amount)
	require.Equal(t, supplyBefore.Amount.Sub(sdkmath.NewInt(200)), app.BankKeeper.GetSupply(ctx, types.BaseDenom).Amount)

	reasons, err := app.BurnKeeper.GetAllBurnReasons(ctx)
	require.NoError(t, err)
	require.Len(t, reasons, 1)
	require.Equal(t, "protocol-deflation", reasons[0].Reason)
}

func TestZeroAndNegativeBurnRejectedWithoutMutation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(100)))[0]
	msgServer := burnkeeper.NewMsgServerImpl(app.BurnKeeper)
	supplyBefore := app.BankKeeper.GetSupply(ctx, types.BaseDenom)

	_, err := msgServer.BurnUserCoins(ctx, &types.MsgBurnUserCoins{
		Burner:	aetraaddress.FormatAccAddress(user),
		Amount:	sdk.NewCoins(),
		Epoch:	1,
	})
	require.ErrorIs(t, err, types.ErrInvalidBurn)

	negative := sdk.Coins{{Denom: types.BaseDenom, Amount: sdkmath.NewInt(-1)}}
	_, err = app.BurnKeeper.BurnUserCoins(ctx, user, negative, 1, "negative")
	require.ErrorIs(t, err, types.ErrInvalidBurn)
	require.Equal(t, supplyBefore, app.BankKeeper.GetSupply(ctx, types.BaseDenom))
	require.Equal(t, sdk.NewCoins(), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName)))
}

func TestUnauthorizedProtocolBurnRejectedWithoutSupplyMutation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(coin(100))))
	supplyBefore := app.BankKeeper.GetSupply(ctx, types.BaseDenom)
	msgServer := burnkeeper.NewMsgServerImpl(app.BurnKeeper)

	_, err := msgServer.BurnProtocolCoins(ctx, &types.MsgBurnProtocolCoins{
		Authority:	app.BurnKeeper.Authority(),
		SourceModule:	minttypes.ModuleName,
		Amount:		sdk.NewCoins(coin(10)),
		Epoch:		2,
		Reason:		"not-allowed",
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.Equal(t, supplyBefore, app.BankKeeper.GetSupply(ctx, types.BaseDenom))

	_, err = msgServer.BurnProtocolCoins(ctx, &types.MsgBurnProtocolCoins{
		Authority:	aetraaddress.FormatAccAddress(sdk.AccAddress(bytes20(1))),
		SourceModule:	types.ModuleName,
		Amount:		sdk.NewCoins(coin(10)),
		Epoch:		2,
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestExportImportPreservesBurnCounters(t *testing.T) {
	source := l1app.Setup(t, false)
	sourceCtx := source.NewContext(false)
	user := l1app.AddTestAddrsWithCoins(t, source, sourceCtx, 1, sdk.NewCoins(coin(1_000)))[0]
	_, err := source.BurnKeeper.BurnUserCoins(sourceCtx, user, sdk.NewCoins(coin(123)), 11, "round-trip")
	require.NoError(t, err)

	exported, err := source.BurnKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := l1app.Setup(t, false)
	targetCtx := target.NewContext(false)
	require.NoError(t, target.BurnKeeper.InitGenesis(targetCtx, *exported))
	imported, err := target.BurnKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func coin(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(types.BaseDenom, amount)
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
