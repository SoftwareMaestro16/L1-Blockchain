package addressing_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/sovereign-l1/l1/app/addressing"
)

type policyTx struct{ msgs []sdk.Msg }

func (tx policyTx) GetMsgs() []sdk.Msg	{ return tx.msgs }
func (tx policyTx) GetMsgsV2() ([]protov2.Message, error) {
	out := make([]protov2.Message, 0, len(tx.msgs))
	for _, msg := range tx.msgs {
		msgV2, ok := msg.(protov2.Message)
		if !ok {
			return nil, nil
		}
		out = append(out, msgV2)
	}
	return out, nil
}

func TestReservedSystemAddressesParseAndMatch(t *testing.T) {
	require.NoError(t, addressing.ValidateReservedSystemAddressCatalog())

	seenNames := map[string]struct{}{}
	seenBytes := map[string]string{}

	for _, address := range addressing.AllSystemAddresses() {
		t.Run(address.Name, func(t *testing.T) {
			require.Equal(t, addressing.SystemAddressStatusActive, address.Status)
			require.Equal(t, strings.ToUpper(address.UserFriendly), address.UserFriendly)

			rawBytes, err := addressing.Parse(address.Raw)
			require.NoError(t, err)
			ufBytes, err := addressing.Parse(address.UserFriendly)
			require.NoError(t, err)

			rawKey, err := addressing.AddressTextBytesKey(address.Raw)
			require.NoError(t, err)
			ufKey, err := addressing.AddressTextBytesKey(address.UserFriendly)
			require.NoError(t, err)
			require.Equal(t, rawKey, ufKey)
			require.Equal(t, rawBytes, ufBytes)
			require.True(t, addressing.IsReservedSystemAddressBytes(rawBytes))
			require.True(t, addressing.IsReservedSystemAddressText(address.Raw))
			require.True(t, addressing.IsReservedSystemAddressText(address.UserFriendly))

			_, duplicateName := seenNames[address.Name]
			require.False(t, duplicateName, "duplicate reserved system address name %s", address.Name)
			seenNames[address.Name] = struct{}{}

			if other, duplicateBytes := seenBytes[rawKey]; duplicateBytes {
				t.Fatalf("duplicate reserved system address bytes used by %s and %s", other, address.Name)
			}
			seenBytes[rawKey] = address.Name
		})
	}
}

func TestReservedSystemAddressVanitySuffixes(t *testing.T) {
	require.Equal(t, 4, addressing.ReservedUserWorkchain)
	require.Equal(t, -7, addressing.ReservedSystemWorkchain)
	require.True(t, addressing.IsSystemRawAddress(addressing.SystemAddressAETElectorRaw))

	elector, found := addressing.SystemAddressByName(addressing.SystemAddressAETElectorName)
	require.True(t, found)
	require.Equal(t, addressing.SystemAddressAETElectorUserFriendly, elector.UserFriendly)
	require.True(t, strings.HasSuffix(elector.UserFriendly, "ELECTOR"))

	config, found := addressing.SystemAddressByName(addressing.SystemAddressAETConfigName)
	require.True(t, found)
	require.Equal(t, addressing.SystemAddressAETConfigUserFriendly, config.UserFriendly)
	require.True(t, strings.HasSuffix(config.UserFriendly, "CONFIG"))

	mint, found := addressing.SystemAddressByName(addressing.SystemAddressAETMintName)
	require.True(t, found)
	require.Equal(t, addressing.SystemAddressAETMintUserFriendly, mint.UserFriendly)
	require.True(t, strings.HasSuffix(mint.UserFriendly, "MINT"))

	burn, found := addressing.SystemAddressByName(addressing.SystemAddressAETBurnName)
	require.True(t, found)
	require.Equal(t, addressing.SystemAddressAETBurnUserFriendly, burn.UserFriendly)
	require.True(t, strings.HasSuffix(burn.UserFriendly, "BURN"))
}

func TestReservedSystemAddressSignerAndRecipientPolicy(t *testing.T) {
	mint, found := addressing.SystemAddressByName(addressing.SystemAddressAETMintName)
	require.True(t, found)
	require.ErrorContains(t, addressing.ValidateUserSignerAddress(mint.Raw), "reserved system address")
	require.ErrorContains(t, addressing.ValidateUserRecipientAddress(mint.Raw), "cannot receive user funds")

	burn, found := addressing.SystemAddressByName(addressing.SystemAddressAETBurnName)
	require.True(t, found)
	require.ErrorContains(t, addressing.ValidateUserSignerAddress(burn.UserFriendly), "reserved system address")
	require.NoError(t, addressing.ValidateUserRecipientAddress(burn.UserFriendly))
}

func TestZeroAddressPolicyRejectsSignerRecipientAdminAuthority(t *testing.T) {
	require.ErrorContains(t, addressing.ValidateUserSignerAddress(addressing.ZeroRawAddress), "zero address")
	require.ErrorContains(t, addressing.ValidateUserRecipientAddress(addressing.ZeroUserFriendly), "zero address")
	require.ErrorContains(t, addressing.ValidateUserAdminAddress("admin", addressing.ZeroUserFriendly), "zero address")
	require.ErrorContains(t, addressing.ValidateTxAuthorityAddress("authority", addressing.ZeroRawAddress), "zero address")
	require.ErrorContains(t, addressing.ValidateNewUserAccountAddress("activation", addressing.ZeroUserFriendly), "zero address")
}

func TestReservedAddressPolicyRejectsUserCreationAndAnteRoles(t *testing.T) {
	mint, found := addressing.SystemAddressByName(addressing.SystemAddressAETMintName)
	require.True(t, found)
	require.ErrorContains(t, addressing.ValidateNewUserAccountAddress("activation", mint.UserFriendly), "reserved system address")
	require.ErrorContains(t, addressing.ValidateUserAdminAddress("admin", mint.UserFriendly), "reserved system address")
	require.ErrorContains(t, addressing.ValidateTxAuthorityAddress("authority", mint.Raw), "reserved system address")

	err := addressing.ValidateAnteAddressPolicy(policyTx{msgs: []sdk.Msg{&banktypes.MsgSend{
		FromAddress:	mint.UserFriendly,
		ToAddress:	addressing.SystemAddressAETBurnUserFriendly,
		Amount:		sdk.NewCoins(sdk.NewInt64Coin("naet", 1)),
	}}})
	require.ErrorContains(t, err, "reserved system address")
}

func TestReservedSystemAddressCatalogRejectsDuplicateAndZeroFixtures(t *testing.T) {
	addresses := addressing.AllSystemAddresses()
	duplicate := append([]addressing.SystemAddress(nil), addresses...)
	duplicate = append(duplicate, addresses[0])
	require.ErrorContains(t, addressing.ValidateSystemAddressCatalog(duplicate), "duplicate reserved system address")

	zero := append([]addressing.SystemAddress(nil), addresses...)
	zero[0].Raw = addressing.ZeroRawAddress
	zero[0].UserFriendly = addressing.ZeroUserFriendly
	require.ErrorContains(t, addressing.ValidateSystemAddressCatalog(zero), "zero address")
}
