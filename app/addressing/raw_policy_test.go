package addressing_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestValidateRawAddressPolicyGoldenVectors(t *testing.T) {

	systemAddrs := []string{
		addressing.SystemAddressAETMintRaw,
		addressing.SystemAddressAETBurnRaw,
		addressing.SystemAddressAETElectorRaw,
		addressing.SystemAddressAETConfigRaw,
	}
	for _, raw := range systemAddrs {
		bz, err := addressing.Parse(raw)
		require.NoError(t, err, "parse system address %s", raw)
		err = addressing.ValidateRawAddressPolicy(bz, addressing.RawAddressPolicyVersionLegacyPadded)
		require.NoError(t, err, "legacy padded policy should accept system address %s", raw)
		err = addressing.ValidateRawAddressPolicy(bz, addressing.RawAddressPolicyVersionV2)
		require.NoError(t, err, "v2 policy should accept system address %s", raw)
	}

	legacyCosmos, err := hex.DecodeString("00000000000000000000000000112233445566778899aabbccddeeff00112233")
	require.NoError(t, err)
	require.Len(t, legacyCosmos, 32)
	err = addressing.ValidateRawAddressPolicy(legacyCosmos, addressing.RawAddressPolicyVersionLegacyPadded)
	require.NoError(t, err, "legacy padded policy must accept legacy padded address")
	err = addressing.ValidateRawAddressPolicy(legacyCosmos, addressing.RawAddressPolicyVersionV2)
	require.ErrorContains(t, err, "avoid legacy padding", "v2 policy must reject legacy padded address")

	v2Addr := make([]byte, 32)
	for i := range v2Addr {
		v2Addr[i] = byte(i*17 + 0xa3)
	}
	err = addressing.ValidateRawAddressPolicy(v2Addr, addressing.RawAddressPolicyVersionV2)
	require.NoError(t, err, "v2 policy must accept true 256-bit address")
	err = addressing.ValidateRawAddressPolicy(v2Addr, addressing.RawAddressPolicyVersionLegacyPadded)
	require.ErrorContains(t, err, "must use legacy padded", "legacy padded policy must reject v2 address")
}

func TestLegacyPaddedRejectedAfterV2Gate(t *testing.T) {

	legacyBytes := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
		0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00,
	}
	require.Len(t, legacyBytes, 32)
	require.True(t, addressing.IsLegacyPaddedRawAddress(legacyBytes))

	err := addressing.ValidateRawAddressPolicy(legacyBytes, addressing.RawAddressPolicyVersionV2)
	require.ErrorContains(t, err, "avoid legacy padding",
		"V2 policy must reject addresses with leading zero padding")

	err = addressing.ValidateRawAddressPolicy(legacyBytes, addressing.RawAddressPolicyVersionLegacyPadded)
	require.NoError(t, err,
		"LegacyPadded policy must accept addresses with leading zero padding")
}

func TestSystemAddressesAcceptedByBothPolicies(t *testing.T) {
	for _, addr := range addressing.AllSystemAddresses() {
		bz, err := addressing.Parse(addr.Raw)
		require.NoError(t, err, "parse %s", addr.Name)

		err = addressing.ValidateRawAddressPolicy(bz, addressing.RawAddressPolicyVersionLegacyPadded)
		require.NoError(t, err, "legacy padded policy must accept system address %s", addr.Name)

		err = addressing.ValidateRawAddressPolicy(bz, addressing.RawAddressPolicyVersionV2)
		require.NoError(t, err, "v2 policy must accept system address %s", addr.Name)

		class := addressing.ClassifyRawAddressBytes(bz)
		require.Equal(t, addressing.RawAddressClassSystemFixed, class,
			"system address %s must classify as system_fixed", addr.Name)
	}
}

func TestNormalizeV2RawAddressForcesHighEntropy(t *testing.T) {

	legacyBytes := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
		0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
	}
	require.True(t, addressing.IsLegacyPaddedRawAddress(legacyBytes))

	v2, err := addressing.NormalizeV2RawAddress("aetra", legacyBytes)
	require.NoError(t, err, "NormalizeV2RawAddress must re-derive legacy padded into v2")
	require.Len(t, v2, 32)
	require.True(t, addressing.IsV2RawAddress(v2),
		"normalized address must classify as v2, got %s", addressing.ClassifyRawAddressBytes(v2))

	v2Direct, err := addressing.NormalizeV2RawAddress("aetra", v2)
	require.NoError(t, err)
	require.Equal(t, v2, v2Direct, "v2 address must roundtrip unchanged through NormalizeV2RawAddress")
}

func TestV2RawAddressClassifiesCorrectly(t *testing.T) {

	nonPadded := make([]byte, 32)
	for i := range nonPadded {
		nonPadded[i] = byte(i*31 + 0x77)
	}
	require.False(t, addressing.IsLegacyPaddedRawAddress(nonPadded))
	require.True(t, addressing.IsV2RawAddress(nonPadded))

	require.Equal(t, addressing.RawAddressClassUnknown, addressing.ClassifyRawAddressBytes(nil))
	require.Equal(t, addressing.RawAddressClassUnknown, addressing.ClassifyRawAddressBytes([]byte{}))

	require.Equal(t, addressing.RawAddressClassUnknown, addressing.ClassifyRawAddressBytes(make([]byte, 16)))
}

func TestV2RawAddressValidationRejectsUnsupportedPolicy(t *testing.T) {
	bz := make([]byte, 32)
	for i := range bz {
		bz[i] = byte(i*19 + 0x99)
	}

	err := addressing.ValidateRawAddressPolicy(bz, 99)
	require.ErrorContains(t, err, "unsupported raw address policy version")
}

func TestDeriveV2RawAddressDeterministic(t *testing.T) {
	seed := []byte("deterministic-seed!!")
	v2a, err := addressing.NormalizeV2RawAddress("aetra", seed)
	require.NoError(t, err)
	require.Len(t, v2a, 32)
	require.True(t, addressing.IsV2RawAddress(v2a))

	v2b, err := addressing.NormalizeV2RawAddress("aetra", seed)
	require.NoError(t, err)
	require.Equal(t, v2a, v2b, "v2 derivation must be deterministic for same domain and seed")
}

func TestNormalizeV2RawAddressRejectsEmptyDomain(t *testing.T) {
	bz := make([]byte, 32)
	copy(bz, []byte("some-20-byte-key!"))
	_, err := addressing.NormalizeV2RawAddress("", bz)
	require.ErrorContains(t, err, "domain is required")
}

func TestLegacyPaddedRoundTripConsistency(t *testing.T) {
	legacyBytes := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x11, 0x22, 0x33, 0x44,
		0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc,
		0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44,
	}
	require.Len(t, legacyBytes, 32)
	require.True(t, addressing.IsLegacyPaddedRawAddress(legacyBytes))

	raw4 := addressing.Format(legacyBytes)
	require.True(t, strings.HasPrefix(raw4, "4:"))

	parsed, err := addressing.Parse(raw4)
	require.NoError(t, err)

	user, err := addressing.FormatUserFriendly(legacyBytes)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(user, "AE"))
	require.Len(t, user, 48)

	parsed2, err := addressing.Parse(user)
	require.NoError(t, err)
	require.Equal(t, parsed, parsed2, "AE and 4: must decode to same bytes")

	reformatted := addressing.Format(parsed2)
	parsed3, err := addressing.Parse(reformatted)
	require.NoError(t, err)
	require.Equal(t, parsed, parsed3, "AE -> 4: -> Parse roundtrip must match")
}

func TestClassifyRawAddressTextInterface(t *testing.T) {

	user := addressing.SystemAddressAETMintUserFriendly
	class, err := addressing.ClassifyRawAddressText(user)
	require.NoError(t, err)
	require.Equal(t, addressing.RawAddressClassSystemFixed, class)

	_, err = addressing.ClassifyRawAddressText(addressing.SystemAddressAETMintRaw)
	require.NoError(t, err)

	_, err = addressing.ClassifyRawAddressText("garbage")
	require.Error(t, err)
}
