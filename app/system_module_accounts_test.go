package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
)

func TestAppBootsWithReservedSystemModuleAccounts(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.NoError(t, ValidateReservedSystemModuleAccountWiring(BlockedAddresses()))

	permissions := GetMaccPerms()
	for _, account := range ReservedSystemModuleAccounts() {
		require.Contains(t, permissions, account.ModuleAccountName, account.Name)

		addr, found, err := ReservedSystemModuleAccountAddress(account.ModuleAccountName)
		require.NoError(t, err)
		require.True(t, found, account.ModuleAccountName)

		storedAccount := app.AccountKeeper.GetAccount(ctx, addr)
		if storedAccount != nil {
			require.Nil(t, storedAccount.GetPubKey(), account.Name)
		}
	}
}

func TestReservedSystemModuleAccountAddressesMatchConstants(t *testing.T) {
	for _, account := range ReservedSystemModuleAccounts() {
		addr, found, err := ReservedSystemModuleAccountAddress(account.ModuleAccountName)
		require.NoError(t, err)
		require.True(t, found, account.ModuleAccountName)

		rawBytes, err := aetraaddress.Parse(account.Raw)
		require.NoError(t, err)
		require.Equal(t, rawBytes, []byte(addr), account.Name)

		catalogAddress, found := aetraaddress.SystemAddressByName(account.Name)
		require.True(t, found, account.Name)
		require.Equal(t, catalogAddress.Raw, account.Raw, account.Name)
		require.Equal(t, catalogAddress.UserFriendly, account.UserFriendly, account.Name)
	}

	mint, found := ReservedSystemModuleAccountByName("AETMint")
	require.True(t, found)
	require.Equal(t, mintauthoritytypes.DefaultMintAuthorityModuleAccount, mint.ModuleAccountName)

	burn, found := ReservedSystemModuleAccountByName("AETBurn")
	require.True(t, found)
	require.Equal(t, "burn", burn.ModuleAccountName)
}

func TestBankBlockedAddressesIncludeNonReceivableSystemAccounts(t *testing.T) {
	blocked := BlockedAddresses()

	for _, address := range aetraaddress.AllSystemAddresses() {
		bz, err := aetraaddress.Parse(address.Raw)
		require.NoError(t, err)

		key := sdk.AccAddress(bz).String()
		require.Equal(t, !address.CanReceiveUserFunds, blocked[key], address.Name)
	}

	for _, account := range ReservedSystemModuleAccounts() {
		bz, err := aetraaddress.Parse(account.Raw)
		require.NoError(t, err)

		require.Equal(t, !account.CanReceiveUserFunds, blocked[sdk.AccAddress(bz).String()], account.Name)
	}
}

func TestReservedSystemBankSendPolicy(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	sender := AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 100)))[0]
	msgServer := bankkeeper.NewMsgServerImpl(app.BankKeeper)

	mint, found := ReservedSystemModuleAccountByName("AETMint")
	require.True(t, found)
	mintAddr, found, err := ReservedSystemModuleAccountAddress(mint.ModuleAccountName)
	require.NoError(t, err)
	require.True(t, found)
	_, err = msgServer.Send(ctx, &banktypes.MsgSend{
		FromAddress:	sender.String(),
		ToAddress:	mint.Raw,
		Amount:		sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1)),
	})
	require.ErrorContains(t, err, "not allowed to receive funds")

	burn, found := ReservedSystemModuleAccountByName("AETBurn")
	require.True(t, found)
	_, err = msgServer.Send(ctx, &banktypes.MsgSend{
		FromAddress:	sender.String(),
		ToAddress:	burn.Raw,
		Amount:		sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 2)),
	})
	require.NoError(t, err)

	burnAddr, found, err := ReservedSystemModuleAccountAddress(burn.ModuleAccountName)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 2), app.BankKeeper.GetBalance(ctx, burnAddr, appparams.BaseDenom))

	require.True(t, app.BankKeeper.BlockedAddr(mintAddr))
	require.False(t, app.BankKeeper.BlockedAddr(burnAddr))
}
