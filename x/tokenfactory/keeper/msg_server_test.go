package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	appparams "github.com/sovereign-l1/l1/app/params"
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

func TestCreateDenomRejectsNativeTokenSpoofing(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	for _, subdenom := range []string{appparams.BaseDenom, appparams.DisplayDenom, appparams.TokenName} {
		t.Run(subdenom, func(t *testing.T) {
			_, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
				Creator:  admin.String(),
				Subdenom: subdenom,
			})
			require.ErrorIs(t, err, types.ErrInvalidDenom)
			require.Contains(t, err.Error(), "native ORB/norb")
		})
	}
}

func TestFactoryTokenMetadataDoesNotReplaceNativeMetadata(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "gold",
	})
	require.NoError(t, err)

	nativeMetadata, found := app.BankKeeper.GetDenomMetaData(ctx, appparams.BaseDenom)
	require.True(t, found)
	require.Equal(t, appparams.BaseDenom, nativeMetadata.Base)
	require.Equal(t, appparams.DisplayDenom, nativeMetadata.Display)
	require.Equal(t, appparams.TokenSymbol, nativeMetadata.Symbol)

	factoryMetadata, found := app.BankKeeper.GetDenomMetaData(ctx, createRes.NewTokenDenom)
	require.True(t, found)
	require.Equal(t, createRes.NewTokenDenom, factoryMetadata.Base)
	require.Equal(t, createRes.NewTokenDenom, factoryMetadata.Display)
	require.Equal(t, createRes.NewTokenDenom, factoryMetadata.Symbol)
	require.NotEqual(t, appparams.BaseDenom, factoryMetadata.Base)
	require.NotEqual(t, appparams.DisplayDenom, factoryMetadata.Display)
}
