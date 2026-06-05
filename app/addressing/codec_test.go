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

	text := addressing.FormatAccAddress(addr)

	require.Len(t, text, addressing.RawAddressLength)
	require.True(t, strings.HasPrefix(text, "0:"))
	require.Equal(t, strings.ToLower(text), text)
	require.Regexp(t, `^0:[0-9a-f]{64}$`, text)

	parsed, err := addressing.ParseAccAddress(text)
	require.NoError(t, err)
	require.Equal(t, addr, parsed)
}

func TestUserFriendlyAddressFormat(t *testing.T) {
	addr := sdk.AccAddress(bytes20(0x22))

	text, err := addressing.FormatUserFriendly(addr)
	require.NoError(t, err)

	require.Len(t, text, addressing.UserFriendlyLength)
	require.True(t, strings.HasPrefix(text, addressing.UserFriendlyPrefix))
	require.Regexp(t, `^[A-Za-z0-9_-]{48}$`, text)

	parsed, err := addressing.ParseAccAddress(text)
	require.NoError(t, err)
	require.Equal(t, addr, parsed)
}

func TestRawLongAddressRoundTrip(t *testing.T) {
	raw, err := hex.DecodeString("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	text := addressing.Format(raw)
	require.Equal(t, "0:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", text)

	parsed, err := addressing.Parse(text)
	require.NoError(t, err)
	require.Equal(t, raw, parsed)
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
