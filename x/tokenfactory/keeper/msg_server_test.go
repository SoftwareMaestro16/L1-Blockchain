package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestBurnRejectsBurnFromUnsignedAccount(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsIncremental(app, ctx, 2, sdkmath.NewInt(1_000_000))
	admin, holder := addrs[0], addrs[1]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "gold",
	})
	require.NoError(t, err)

	denom := createRes.NewTokenDenom
	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 100),
		MintToAddress: holder.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.Burn(ctx, &types.MsgBurn{
		Sender:          admin.String(),
		Amount:          sdk.NewInt64Coin(denom, 1),
		BurnFromAddress: holder.String(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn_from_address must match sender")

	balance := app.BankKeeper.GetBalance(ctx, holder, denom)
	require.Equal(t, "100", balance.Amount.String())
}

func TestAdminCanBurnOwnFactoryTokens(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "silver",
	})
	require.NoError(t, err)

	denom := createRes.NewTokenDenom
	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 100),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.Burn(ctx, &types.MsgBurn{
		Sender:          admin.String(),
		Amount:          sdk.NewInt64Coin(denom, 40),
		BurnFromAddress: admin.String(),
	})
	require.NoError(t, err)

	balance := app.BankKeeper.GetBalance(ctx, admin, denom)
	require.Equal(t, "60", balance.Amount.String())
}
