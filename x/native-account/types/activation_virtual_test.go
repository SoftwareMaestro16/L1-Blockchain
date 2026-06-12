package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestMsgActivateAccountRequiresDerivedAEAddress(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)

	require.NoError(t, MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey}.ValidateBasic())

	other := addressing.FormatAccAddress(sdk.AccAddress(virtualBytes20(0x99)))
	require.ErrorContains(t, MsgActivateAccount{AddressUser: other, PublicKey: pubKey}.ValidateBasic(), "must equal derived")
}

func TestMsgActivateAccountRejectsLegacyAndForeignAddressFormats(t *testing.T) {
	pubKey := activationTestPubKey()
	foreign, err := sdk.Bech32ifyAddressBytes("cosmos", virtualBytes20(0x13))
	require.NoError(t, err)

	for _, text := range []string{
		foreign,
		"0:0000000000000000000000000000000000000000000000000000000000000000",
		"4:ABCDEFabcdef0000000000000000000000000000000000000000000000000000",
		"AE-not-a-valid-address",
	} {
		require.Error(t, MsgActivateAccount{AddressUser: text, PublicKey: pubKey}.ValidateBasic(), text)
	}
}

func TestMsgActivateAccountRejectsReservedSystemAddress(t *testing.T) {
	pubKey := activationTestPubKey()
	mint, found := addressing.SystemAddressByName(addressing.SystemAddressAETMintName)
	require.True(t, found)

	err := MsgActivateAccount{AddressUser: mint.UserFriendly, PublicKey: pubKey}.ValidateBasic()

	require.ErrorContains(t, err, "reserved system address")
}

func TestQueryUnactivatedAddressReturnsVirtualInactiveWithoutState(t *testing.T) {
	book, err := NewVirtualAccountBook()
	require.NoError(t, err)
	user := addressing.FormatAccAddress(sdk.AccAddress(virtualBytes20(0x21)))

	view, err := book.QueryAccount(user)

	require.NoError(t, err)
	require.Equal(t, user, view.AddressUser)
	require.Equal(t, addressing.Format(sdk.AccAddress(virtualBytes20(0x21))), view.AddressRaw)
	require.Equal(t, VirtualAccountStatusInactive, view.Status)
	require.False(t, view.Persistent)
	require.False(t, view.StorageRentActive)
	require.False(t, StorageRentAccruesForAccount(view))
	require.Empty(t, book.ExportGenesisAccounts())
}

func TestActivatedAccountIsPersistentAndExported(t *testing.T) {
	book, err := NewVirtualAccountBook()
	require.NoError(t, err)
	user := addressing.FormatAccAddress(sdk.AccAddress(virtualBytes20(0x22)))

	require.NoError(t, book.PutPersistentAccount(VirtualAccountView{
		AddressUser:	user,
		Status:		VirtualAccountStatusActive,
	}, PersistentWriteReasonActivation))

	view, err := book.QueryAccount(user)
	require.NoError(t, err)
	require.True(t, view.Persistent)
	require.True(t, view.StorageRentActive)
	require.Equal(t, []VirtualAccountView{view}, book.ExportGenesisAccounts())
}

func TestInactiveAccountCannotBePersistedOrChargedRent(t *testing.T) {
	book, err := NewVirtualAccountBook()
	require.NoError(t, err)
	user := addressing.FormatAccAddress(sdk.AccAddress(virtualBytes20(0x23)))

	err = book.PutPersistentAccount(VirtualAccountView{
		AddressUser:	user,
		Status:		VirtualAccountStatusInactive,
	}, PersistentWriteReasonActivation)

	require.ErrorContains(t, err, "virtual only")
	require.Empty(t, book.ExportGenesisAccounts())
}

func TestPersistentAccountWritesRequireActivationOrControlledMigration(t *testing.T) {
	book, err := NewVirtualAccountBook()
	require.NoError(t, err)
	user := addressing.FormatAccAddress(sdk.AccAddress(virtualBytes20(0x24)))

	err = book.PutPersistentAccount(VirtualAccountView{
		AddressUser:	user,
		Status:		VirtualAccountStatusActive,
	}, "query")

	require.ErrorContains(t, err, "activation or controlled migration")
	require.Empty(t, book.ExportGenesisAccounts())
}

func TestInactiveAccountRejectsNonActivationMessages(t *testing.T) {
	book, err := NewVirtualAccountBook()
	require.NoError(t, err)
	user := addressing.FormatAccAddress(sdk.AccAddress(virtualBytes20(0x25)))
	view, err := book.QueryAccount(user)
	require.NoError(t, err)

	require.NoError(t, ValidateAccountMessage(view, AccountMessageActivate))
	require.ErrorContains(t, ValidateAccountMessage(view, AccountMessageNormal), "inactive account can only send MsgActivateAccount")
}

func activationTestPubKey() *secp256k1.PubKey {
	return &secp256k1.PubKey{Key: []byte{
		0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b, 0x07,
		0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98,
	}}
}

func virtualBytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
