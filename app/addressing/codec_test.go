package addressing_test

import (
	"encoding/hex"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestRawAddressFormat(t *testing.T) {
	addr := sdk.AccAddress(bytes20(0x11))

	text := addressing.Format(addr)

	require.Len(t, text, addressing.RawAddressLength)
	require.True(t, strings.HasPrefix(text, "4:"))
	require.Equal(t, strings.ToLower(text), text)
	require.Regexp(t, `^4:[0-9a-f]{64}$`, text)

	parsed, err := addressing.ParseAccAddress(text)
	require.NoError(t, err)
	require.Equal(t, addr, parsed)
}

func TestUserFacingAddressFormats(t *testing.T) {
	addr := sdk.AccAddress(bytes20(0x22))

	text := addressing.FormatAccAddress(addr)
	requireUserFriendlyAddress(t, text)

	parsed, err := addressing.ParseAccAddress(text)
	require.NoError(t, err)
	require.Equal(t, addr, parsed)

	requireUserFriendlyAddress(t, addressing.FormatValAddress(sdk.ValAddress(addr)))
	requireUserFriendlyAddress(t, addressing.FormatConsAddress(sdk.ConsAddress(addr)))
}

func TestRawLongAddressRoundTrip(t *testing.T) {
	raw, err := hex.DecodeString("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	text := addressing.Format(raw)
	require.Equal(t, "4:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", text)

	parsed, err := addressing.Parse(text)
	require.NoError(t, err)
	require.Equal(t, raw, parsed)
}

func TestSystemRawAddressRoundTrip(t *testing.T) {
	raw, err := hex.DecodeString("01041041041041041041041041041041041041041041041041041042c4093391")
	require.NoError(t, err)

	text := addressing.FormatSystemRawAddress(raw)
	require.Equal(t, "-7:01041041041041041041041041041041041041041041041041041042c4093391", text)
	require.True(t, addressing.IsSystemRawAddress(text))

	parsed, err := addressing.ParseSystemRawAddress(text)
	require.NoError(t, err)
	require.Equal(t, raw, parsed)

	parsedGeneric, err := addressing.Parse(text)
	require.NoError(t, err)
	require.Equal(t, raw, parsedGeneric)
}

func TestZeroAddressFormats(t *testing.T) {
	zero := sdk.AccAddress(bytes20(0))

	require.Equal(t, addressing.ZeroRawAddress, addressing.Format(zero))
	require.Equal(t, addressing.ZeroUserFriendly, addressing.FormatAccAddress(zero))
	require.True(t, addressing.IsZeroAccAddress(zero))

	userFriendly, err := addressing.FormatUserFriendly(zero)
	require.NoError(t, err)
	require.Equal(t, addressing.ZeroUserFriendly, userFriendly)

	rawParsed, err := addressing.ParseAccAddress(addressing.ZeroRawAddress)
	require.NoError(t, err)
	require.True(t, addressing.IsZeroAccAddress(rawParsed))

	friendlyParsed, err := addressing.ParseAccAddress(addressing.ZeroUserFriendly)
	require.NoError(t, err)
	require.True(t, addressing.IsZeroAccAddress(friendlyParsed))
}

func TestZeroAddressValidationPolicy(t *testing.T) {
	valid := sdk.AccAddress(bytes20(0x33))
	validText := addressing.FormatAccAddress(valid)

	require.NoError(t, addressing.ValidateUserAddress("recipient", validText))
	require.NoError(t, addressing.ValidateAuthorityAddress("authority", validText))
	require.NoError(t, addressing.ValidateContractAddress("contract", validText))
	require.NoError(t, addressing.RejectZeroAddress("signer", valid.Bytes()))

	require.ErrorContains(t, addressing.ValidateUserAddress("recipient", addressing.ZeroRawAddress), "must not be zero address")
	require.ErrorContains(t, addressing.ValidateUserAddress("recipient", addressing.ZeroUserFriendly), "must not be zero address")
	require.ErrorContains(t, addressing.ValidateAuthorityAddress("authority", addressing.ZeroRawAddress), "must not be zero address")
	require.ErrorContains(t, addressing.ValidateContractAddress("contract", addressing.ZeroRawAddress), "must not be zero address")
	require.ErrorContains(t, addressing.RejectZeroAddress("signer", sdk.AccAddress(bytes20(0)).Bytes()), "must not be zero address")

	_, present, err := addressing.ParseOptionalAdminAddress("admin", "")
	require.NoError(t, err)
	require.False(t, present)
	require.ErrorContains(t, addressing.ValidateOptionalAdminAddress("admin", addressing.ZeroRawAddress), "must not be zero address")
}

func TestAddressValidationRejectsEmptyMalformedAndLegacyFormats(t *testing.T) {
	validLegacy, err := sdk.Bech32ifyAddressBytes("orb", bytes20(0x44))
	require.NoError(t, err)

	validFriendly, err := addressing.FormatUserFriendly(sdk.AccAddress(bytes20(0x46)))
	require.NoError(t, err)

	tests := map[string]string{
		"empty":                      "",
		"blank":                      "   ",
		"malformed bech32":           "ae1notvalid",
		"foreign bech32":             "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp2n8k9",
		"old raw prefix":             "0:0000000000000000000000000000000000000000000000000000000000000000",
		"mixed case raw":             "4:ABCDEFabcdef0000000000000000000000000000000000000000000000000000",
		"mixed case system raw":      "-7:ABCDEFabcdef0000000000000000000000000000000000000000000000000000",
		"wrong system raw length":    "-7:00000000000000000000000000000000000000000000000000000000000000",
		"wrong length raw":           "4:00000000000000000000000000000000000000000000000000000000000000",
		"old userfriendly prefix":    "ORBAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"wrong userfriendly prefix":  "AF" + validFriendly[2:],
		"non base64url userfriendly": "AE+/" + validFriendly[4:],
		"old bech32 account prefix":  validLegacy,
	}
	for name, text := range tests {
		t.Run(name, func(t *testing.T) {
			require.Error(t, addressing.ValidateUserAddress("sender", text))
		})
	}
}

func TestAddressValidationAcceptsCurrentSDKBech32Compatibility(t *testing.T) {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("ae", "aepub")

	valid, err := sdk.Bech32ifyAddressBytes("ae", bytes20(0x45))
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(valid, "ae1"))
	require.NoError(t, addressing.ValidateUserAddress("sender", valid))
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}

func requireUserFriendlyAddress(t *testing.T, text string) {
	t.Helper()

	require.Len(t, text, addressing.UserFriendlyLength)
	require.True(t, strings.HasPrefix(text, addressing.UserFriendlyPrefix))
	require.Regexp(t, `^[A-Za-z0-9_-]{48}$`, text)
	require.NotRegexp(t, `^[a-z]+1`, text)
}
