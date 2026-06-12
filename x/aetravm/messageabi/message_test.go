package messageabi

import (
	"bytes"
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestEncodeDecodeRoundTripAndDebugJSON(t *testing.T) {
	msg := validMessage(t)
	encoded, err := Encode(msg, DefaultParams())
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(encoded, []byte(Magic)))

	decoded, err := Decode(encoded, DefaultParams(), 90)
	require.NoError(t, err)
	require.Equal(t, msg, decoded)

	debug, err := DebugJSON(msg, DefaultParams())
	require.NoError(t, err)
	require.Contains(t, string(debug), `"magic":"AVMM"`)
	require.Contains(t, string(debug), `"kind":"internal"`)
	require.Contains(t, string(debug), `"body_hex":"010203"`)
}

func TestCanonicalHashStable(t *testing.T) {
	id, err := MessageIDHex(validMessage(t), DefaultParams())
	require.NoError(t, err)
	require.Equal(t, "1a2a7761383d1681c27591e7c4c89d5b83dbe673d868345a701e2d7c78227563", id)
}

func TestMalformedABIRejected(t *testing.T) {
	msg := validMessage(t)
	encoded, err := Encode(msg, DefaultParams())
	require.NoError(t, err)

	_, err = Decode([]byte("BAD!"), DefaultParams(), 1)
	require.ErrorContains(t, err, "invalid magic")

	_, err = Decode(encoded[:len(encoded)-1], DefaultParams(), 1)
	require.ErrorContains(t, err, "truncated")

	malformedBool := append([]byte(nil), encoded...)
	bodyStart := len(Magic) + 2 + 1 + 8 + 8
	bodyStart += 4 + len(msg.Sender.User)
	bodyStart += 4 + len(msg.Sender.Raw)
	bodyStart += 4 + len(msg.Destination.User)
	bodyStart += 4 + len(msg.Destination.Raw)
	bodyStart += 8
	malformedBool[bodyStart] = 2
	_, err = Decode(malformedBool, DefaultParams(), 1)
	require.ErrorContains(t, err, "invalid bool")
}

func TestOversizedBodyRejected(t *testing.T) {
	msg := validMessage(t)
	msg.Body = []byte{1, 2, 3, 4}
	params := DefaultParams()
	params.MaxBodyBytes = 3

	_, err := Encode(msg, params)
	require.ErrorContains(t, err, "body exceeds maximum")

	msg = validMessage(t)
	msg.Metadata = []byte{1, 2, 3, 4}
	params = DefaultParams()
	params.MaxMetadataBytes = 3
	_, err = Encode(msg, params)
	require.ErrorContains(t, err, "metadata exceeds maximum")
}

func TestZeroSenderAndDestinationRejected(t *testing.T) {
	msg := validMessage(t)
	msg.Sender = zeroPair()
	require.ErrorContains(t, msg.Validate(DefaultParams(), 1), "sender address must not be zero")

	msg = validMessage(t)
	msg.Destination = zeroPair()
	require.ErrorContains(t, msg.Validate(DefaultParams(), 1), "destination address must not be zero")
}

func TestExpiredMessageRejected(t *testing.T) {
	msg := validMessage(t)
	msg.DeadlineBlock = 99
	require.NoError(t, msg.Validate(DefaultParams(), 99))
	require.ErrorContains(t, msg.Validate(DefaultParams(), 100), "expired")

	encoded, err := Encode(msg, DefaultParams())
	require.NoError(t, err)
	_, err = Decode(encoded, DefaultParams(), 100)
	require.ErrorContains(t, err, "expired")
}

func TestInvalidBounceCombinationRejected(t *testing.T) {
	msg := validMessage(t)
	msg.Bounced = true
	require.ErrorContains(t, msg.Validate(DefaultParams(), 1), "invalid bounce combination")

	msg = validMessage(t)
	msg.Kind = KindBounced
	msg.Bounce = false
	msg.Bounced = false
	require.ErrorContains(t, msg.Validate(DefaultParams(), 1), "requires bounced flag")

	msg.Bounced = true
	require.NoError(t, msg.Validate(DefaultParams(), 1))
}

func TestMessageIDDeterminismAndMutations(t *testing.T) {
	msg := validMessage(t)
	id1, err := MessageIDHex(msg, DefaultParams())
	require.NoError(t, err)
	id2, err := MessageIDHex(msg, DefaultParams())
	require.NoError(t, err)
	require.Equal(t, id1, id2)

	mutated := msg
	mutated.Opcode++
	requireDifferentID(t, msg, mutated)

	mutated = msg.Clone()
	mutated.Body = append(mutated.Body, 0xff)
	requireDifferentID(t, msg, mutated)

	mutated = msg
	mutated.DeadlineBlock++
	requireDifferentID(t, msg, mutated)
}

func TestUnsupportedAddressFormatsRejected(t *testing.T) {
	msg := validMessage(t)
	msg.Sender.Raw = "4:ABCDEF0000000000000000000000000000000000000000000000000000000000"
	require.ErrorContains(t, msg.Validate(DefaultParams(), 1), "raw address")

	msg = validMessage(t)
	msg.Sender.User = msg.Sender.Raw
	require.ErrorContains(t, msg.Validate(DefaultParams(), 1), "AE format")
}

func requireDifferentID(t *testing.T, left, right Message) {
	t.Helper()
	leftID, err := MessageIDHex(left, DefaultParams())
	require.NoError(t, err)
	rightID, err := MessageIDHex(right, DefaultParams())
	require.NoError(t, err)
	require.NotEqual(t, leftID, rightID)
}

func validMessage(t *testing.T) Message {
	t.Helper()
	return Message{
		Kind:		KindInternal,
		Opcode:		0x1020_3040_5060_7080,
		QueryID:	42,
		Sender:		pair(t, 0x11),
		Destination:	pair(t, 0x22),
		ValueNAET:	300_000_000,
		Bounce:		true,
		DeadlineBlock:	100,
		GasLimit:	1_000_000,
		Body:		[]byte{1, 2, 3},
		StateInit:	[]byte("state-init-canonical"),
		Metadata:	[]byte("debug-metadata"),
		Signature:	bytes.Repeat([]byte{0x33}, 64),
	}
}

func pair(t *testing.T, fill byte) AddressPair {
	t.Helper()
	addr := sdk.AccAddress(bytes.Repeat([]byte{fill}, 20))
	pair, err := AddressPairFromUser(addressing.FormatAccAddress(addr))
	require.NoError(t, err)
	require.NoError(t, validateAddressPair("test", pair))
	return pair
}

func zeroPair() AddressPair {
	raw, _ := hex.DecodeString("0000000000000000000000000000000000000000")
	user := addressing.FormatAccAddress(sdk.AccAddress(raw))
	return AddressPair{User: user, Raw: addressing.ZeroRawAddress}
}
