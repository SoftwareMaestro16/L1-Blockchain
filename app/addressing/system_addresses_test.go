package addressing_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestReservedSystemAddressesParseAndMatch(t *testing.T) {
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
	elector, found := addressing.SystemAddressByName("AETElector")
	require.True(t, found)
	require.True(t, strings.HasSuffix(elector.UserFriendly, "ELECTOR"))

	config, found := addressing.SystemAddressByName("AETConfig")
	require.True(t, found)
	require.True(t, strings.HasSuffix(config.UserFriendly, "CONFIG"))

	mint, found := addressing.SystemAddressByName("AETMint")
	require.True(t, found)
	require.True(t, strings.HasSuffix(mint.UserFriendly, "MINT"))

	burn, found := addressing.SystemAddressByName("AETBurn")
	require.True(t, found)
	require.True(t, strings.HasSuffix(burn.UserFriendly, "BURN"))
}
