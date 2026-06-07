package aft

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestCreateTokenMasterMintAndDeriveHolderWallet(t *testing.T) {
	admin := testAddr(1)
	holder := testAddr(2)
	state := newTestState(t, admin)

	walletAddr, err := state.WalletAddress(holder)
	require.NoError(t, err)
	again, err := DeriveWalletAddress(state.Master.Address, holder, state.Master.WalletCodeHash)
	require.NoError(t, err)
	require.Equal(t, walletAddr, again)

	require.NoError(t, state.Mint(admin, holder, sdkmath.NewInt(1_000)))
	wallet, ok, err := state.Wallet(holder)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, walletAddr, wallet.Address)
	require.Equal(t, sdkmath.NewInt(1_000), wallet.Balance)
	require.Equal(t, sdkmath.NewInt(1_000), state.Master.TotalSupply)
	require.NoError(t, state.ValidateAccounting())
}

func TestTransferDeploysMissingRecipientWalletAndBurnDecrementsSupply(t *testing.T) {
	admin := testAddr(1)
	alice := testAddr(2)
	bob := testAddr(3)
	state := newTestState(t, admin)

	require.NoError(t, state.Mint(admin, alice, sdkmath.NewInt(1_000)))
	require.NoError(t, state.Transfer(alice, bob, sdkmath.NewInt(400), 10))

	aliceWallet, ok, err := state.Wallet(alice)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, sdkmath.NewInt(600), aliceWallet.Balance)

	bobWallet, ok, err := state.Wallet(bob)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, sdkmath.NewInt(400), bobWallet.Balance)
	require.Equal(t, sdkmath.NewInt(1_000), state.Master.TotalSupply)

	require.NoError(t, state.Burn(bob, sdkmath.NewInt(100), 11))
	bobWallet, ok, err = state.Wallet(bob)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, sdkmath.NewInt(300), bobWallet.Balance)
	require.Equal(t, sdkmath.NewInt(900), state.Master.TotalSupply)
	require.NoError(t, state.ValidateAccounting())
}

func TestAdminTransferAndRenounce(t *testing.T) {
	admin := testAddr(1)
	nextAdmin := testAddr(2)
	holder := testAddr(3)
	state := newTestState(t, admin)

	require.NoError(t, state.ChangeAdmin(admin, nextAdmin))
	require.ErrorContains(t, state.Mint(nextAdmin, holder, sdkmath.NewInt(1)), "only token admin")
	require.NoError(t, state.AcceptAdmin(nextAdmin))
	require.ErrorContains(t, state.Mint(admin, holder, sdkmath.NewInt(1)), "only token admin")
	require.NoError(t, state.Mint(nextAdmin, holder, sdkmath.NewInt(1)))
	require.NoError(t, state.RenounceAdmin(nextAdmin))
	require.True(t, state.Master.AdminRenounced)
	require.False(t, state.Master.Mintable)
	require.ErrorContains(t, state.Mint(nextAdmin, holder, sdkmath.NewInt(1)), "token admin is renounced")
}

func TestAdminControlsMetadata(t *testing.T) {
	admin := testAddr(1)
	attacker := testAddr(2)
	state := newTestState(t, admin)
	metadata := TokenMetadata{Name: "Updated USD", Symbol: "UUSD", Decimals: 6, ContentRef: "ipfs://updated"}

	require.ErrorContains(t, state.ChangeMetadata(attacker, metadata), "only token admin")
	require.NoError(t, state.ChangeMetadata(admin, metadata))
	require.Equal(t, metadata, state.Master.Metadata)

	require.ErrorContains(t, state.ChangeMetadata(admin, TokenMetadata{Name: "Aetra", Symbol: "UUSD"}), "must not spoof")
	require.Equal(t, metadata, state.Master.Metadata)
}

func TestNonAdminMintRejected(t *testing.T) {
	admin := testAddr(1)
	attacker := testAddr(2)
	holder := testAddr(3)
	state := newTestState(t, admin)

	require.ErrorContains(t, state.Mint(attacker, holder, sdkmath.NewInt(1)), "only token admin")
	require.Equal(t, sdkmath.ZeroInt(), state.Master.TotalSupply)
	require.NoError(t, state.ValidateAccounting())
}

func TestNativeAETMetadataSpoofRejected(t *testing.T) {
	tests := []TokenMetadata{
		{Name: "Aetra", Symbol: "USDT", Decimals: 6},
		{Name: "Wrapped USD", Symbol: "AET", Decimals: 6},
		{Name: "Wrapped USD", Symbol: "naet", Decimals: 6},
		{Name: "Wrapped USD", Symbol: "USDT", DisplayName: "AET", Decimals: 6},
	}

	for _, metadata := range tests {
		t.Run(metadata.Name+"-"+metadata.Symbol+"-"+metadata.DisplayName, func(t *testing.T) {
			master := testMaster(testAddr(1))
			master.Metadata = metadata
			_, err := NewState(master)
			require.ErrorContains(t, err, "must not spoof native AET/naet")
		})
	}
}

func TestNonNaetFeeRejectedForAFT1Operations(t *testing.T) {
	require.NoError(t, ValidateOperationFees(sdk.NewCoins(sdk.NewInt64Coin("naet", 1))))

	tests := []sdk.Coins{
		sdk.Coins{},
		sdk.NewCoins(sdk.NewInt64Coin("uatom", 1)),
		sdk.NewCoins(sdk.NewInt64Coin("testtoken", 1)),
		sdk.Coins{sdk.NewInt64Coin("naet", 1), sdk.NewInt64Coin("testtoken", 1)},
	}
	for _, fees := range tests {
		t.Run(fees.String(), func(t *testing.T) {
			require.Error(t, ValidateOperationFees(fees))
		})
	}
}

func TestBounceFinalizesPendingQueryDeterministically(t *testing.T) {
	admin := testAddr(1)
	alice := testAddr(2)
	bob := testAddr(3)
	state := newTestState(t, admin)

	require.NoError(t, state.Mint(admin, alice, sdkmath.NewInt(100)))
	require.NoError(t, state.Transfer(alice, bob, sdkmath.NewInt(25), 99))
	require.ErrorContains(t, state.Transfer(alice, bob, sdkmath.NewInt(1), 99), "replayed wallet message")

	require.NoError(t, state.BounceTransfer(alice, sdkmath.NewInt(25), 99))
	aliceWallet, ok, err := state.Wallet(alice)
	require.NoError(t, err)
	require.True(t, ok)
	_, pending := aliceWallet.PendingQueryIDs[99]
	_, processed := aliceWallet.ProcessedQueryIDs[99]
	require.False(t, pending)
	require.True(t, processed)
	require.ErrorContains(t, state.Transfer(alice, bob, sdkmath.NewInt(1), 99), "replayed wallet message")
	require.NoError(t, state.ValidateAccounting())
}

func TestMalformedAFT1MessagesRejected(t *testing.T) {
	admin := testAddr(1)
	alice := testAddr(2)
	state := newTestState(t, admin)

	_, err := DeriveWalletAddress(testAddr(1), alice, []byte{1})
	require.ErrorContains(t, err, "wallet code hash")
	_, err = DeriveWalletAddress(sdk.AccAddress(make([]byte, 20)), alice, testCodeHash())
	require.ErrorContains(t, err, "must not be zero")
	require.ErrorContains(t, state.Mint(admin, alice, sdkmath.ZeroInt()), "mint amount must be positive")
	require.ErrorContains(t, state.Transfer(alice, testAddr(3), sdkmath.NewInt(1), 1), "token wallet does not exist")
	require.NoError(t, state.Mint(admin, alice, sdkmath.NewInt(10)))
	require.ErrorContains(t, state.Transfer(alice, testAddr(3), sdkmath.ZeroInt(), 1), "transfer amount must be positive")
	require.ErrorContains(t, state.Transfer(alice, testAddr(3), sdkmath.NewInt(1), 0), "query id must be non-zero")
	require.ErrorContains(t, state.Burn(alice, sdkmath.ZeroInt(), 1), "burn amount must be positive")
	require.ErrorContains(t, state.Burn(alice, sdkmath.NewInt(1), 0), "query id must be non-zero")
}

func newTestState(t testing.TB, admin sdk.AccAddress) *State {
	t.Helper()
	state, err := NewState(testMaster(admin))
	require.NoError(t, err)
	return state
}

func testMaster(admin sdk.AccAddress) MasterState {
	return MasterState{
		Address:        testAddr(9),
		Admin:          admin,
		TotalSupply:    sdkmath.ZeroInt(),
		Mintable:       true,
		Burnable:       true,
		WalletCodeHash: testCodeHash(),
		Metadata: TokenMetadata{
			Name:       "USD Test Asset",
			Symbol:     "USDT",
			Decimals:   6,
			ContentRef: "ipfs://bafy-aft44-usdt",
		},
	}
}

func testAddr(fill byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{fill}, 20))
}

func testCodeHash() []byte {
	return bytes.Repeat([]byte{7}, WalletCodeHashLength)
}
