package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	orbitaladdress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
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

func TestTokenfactoryLifecycleEmitsStableEventsAndStateQueries(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsIncremental(app, ctx, 2, sdkmath.NewInt(1_000_000))
	admin, newAdmin := addrs[0], addrs[1]
	adminText := orbitaladdress.FormatAccAddress(admin)
	newAdminText := orbitaladdress.FormatAccAddress(newAdmin)
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "events",
	})
	require.NoError(t, err)
	denom := createRes.NewTokenDenom
	requireEvent(t, ctx, types.EventTypeCreateDenom, map[string]string{
		types.AttributeKeyDenom:   denom,
		types.AttributeKeyCreator: adminText,
		types.AttributeKeyAdmin:   adminText,
	})
	meta, found, err := app.TokenFactoryKeeper.GetDenom(ctx, denom)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, adminText, meta.Admin)

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 100),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)
	requireEvent(t, ctx, types.EventTypeMint, map[string]string{
		types.AttributeKeyDenom:         denom,
		types.AttributeKeySender:        adminText,
		types.AttributeKeyAmount:        "100",
		types.AttributeKeyMintToAddress: adminText,
	})
	require.Equal(t, "100", app.BankKeeper.GetBalance(ctx, admin, denom).Amount.String())
	require.Equal(t, "100", app.BankKeeper.GetSupply(ctx, denom).Amount.String())

	_, err = msgServer.Burn(ctx, &types.MsgBurn{
		Sender:          admin.String(),
		Amount:          sdk.NewInt64Coin(denom, 40),
		BurnFromAddress: admin.String(),
	})
	require.NoError(t, err)
	requireEvent(t, ctx, types.EventTypeBurn, map[string]string{
		types.AttributeKeyDenom:           denom,
		types.AttributeKeySender:          adminText,
		types.AttributeKeyAmount:          "40",
		types.AttributeKeyBurnFromAddress: adminText,
	})
	require.Equal(t, "60", app.BankKeeper.GetBalance(ctx, admin, denom).Amount.String())
	require.Equal(t, "60", app.BankKeeper.GetSupply(ctx, denom).Amount.String())

	_, err = msgServer.ChangeAdmin(ctx, &types.MsgChangeAdmin{
		Sender:   admin.String(),
		Denom:    denom,
		NewAdmin: newAdmin.String(),
	})
	require.NoError(t, err)
	requireEvent(t, ctx, types.EventTypeChangeAdmin, map[string]string{
		types.AttributeKeyDenom:    denom,
		types.AttributeKeySender:   adminText,
		types.AttributeKeyNewAdmin: newAdminText,
	})
	meta, found, err = app.TokenFactoryKeeper.GetDenom(ctx, denom)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, newAdminText, meta.Admin)
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

func TestMintToBlockedModuleAddressDoesNotLeakSupply(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "blockedmint",
	})
	require.NoError(t, err)

	blockedRecipient := app.AccountKeeper.GetModuleAddress(dextypes.ModuleName)
	require.NotNil(t, blockedRecipient)

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(createRes.NewTokenDenom, 100),
		MintToAddress: blockedRecipient.String(),
	})
	require.Error(t, err)

	require.True(t, app.BankKeeper.GetSupply(ctx, createRes.NewTokenDenom).Amount.IsZero())
	require.True(t, app.BankKeeper.GetBalance(ctx, blockedRecipient, createRes.NewTokenDenom).Amount.IsZero())
	requireNoEvent(t, ctx, types.EventTypeMint)
}

func TestCreateDenomRejectsNativeTokenSpoofing(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	for _, subdenom := range []string{appparams.BaseDenom, appparams.DisplayDenom, "orb", appparams.TokenName} {
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

func TestAdminTransferRejectsOldAdminAndPreservesSupply(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsIncremental(app, ctx, 2, sdkmath.NewInt(1_000_000))
	admin, newAdmin := addrs[0], addrs[1]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "adminflow",
	})
	require.NoError(t, err)
	denom := createRes.NewTokenDenom

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 1_000),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.ChangeAdmin(ctx, &types.MsgChangeAdmin{
		Sender:   admin.String(),
		Denom:    denom,
		NewAdmin: newAdmin.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 1),
		MintToAddress: admin.String(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        newAdmin.String(),
		Amount:        sdk.NewInt64Coin(denom, 100),
		MintToAddress: newAdmin.String(),
	})
	require.NoError(t, err)

	supply := app.BankKeeper.GetSupply(ctx, denom)
	require.Equal(t, "1100", supply.Amount.String())
	require.Equal(t, "1000", app.BankKeeper.GetBalance(ctx, admin, denom).Amount.String())
	require.Equal(t, "100", app.BankKeeper.GetBalance(ctx, newAdmin, denom).Amount.String())
}
