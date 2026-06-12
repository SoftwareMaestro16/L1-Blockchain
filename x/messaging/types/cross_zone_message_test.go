package types

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestCrossZoneMessageEnvelopeCanonicalEncodingAndID(t *testing.T) {
	params := testCrossZoneParams()
	msg := testCrossZoneMessage(t, params, 1, 1, []byte("payload"))
	require.Len(t, msg.MessageID, MessageIDBytes)
	require.Len(t, msg.PayloadHash, MessageIDBytes)
	require.Equal(t, ComputeCrossZonePayloadHash([]byte("payload")), msg.PayloadHash)

	derived, err := DeriveCrossZoneMessageID(msg, params)
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, derived)

	encodedA, err := CanonicalCrossZoneMessageBinary(msg)
	require.NoError(t, err)
	encodedB, err := CanonicalCrossZoneMessageBinary(msg)
	require.NoError(t, err)
	require.Equal(t, encodedA, encodedB)

	tampered := msg
	tampered.Payload = []byte("different")
	require.ErrorContains(t, tampered.Validate(params), "payload hash mismatch")
}

func TestCrossZoneReplayRejectsDuplicateIDsNoncesAndSequences(t *testing.T) {
	params := testCrossZoneParams()
	state := NewCrossZoneReplayState()
	first := testCrossZoneMessage(t, params, 1, 1, []byte("payload-1"))

	next, err := state.CheckAndRecord(first, params)
	require.NoError(t, err)
	_, err = next.CheckAndRecord(first, params)
	require.ErrorContains(t, err, "message_id")

	lowerNonce := testCrossZoneMessage(t, params, 1, 2, []byte("payload-2"))
	_, err = next.CheckAndRecord(lowerNonce, params)
	require.ErrorContains(t, err, "nonce")

	lowerSequence := testCrossZoneMessage(t, params, 2, 1, []byte("payload-3"))
	_, err = next.CheckAndRecord(lowerSequence, params)
	require.ErrorContains(t, err, "source sequence")

	otherScope := testCrossZoneMessage(t, params, 1, 1, []byte("payload-4"))
	otherScope.Sender = addr(3)
	otherScope.MessageID = nil
	otherScope.PayloadHash = nil
	otherScope, err = NewCrossZoneMessageEnvelope(otherScope, params)
	require.NoError(t, err)
	_, err = next.CheckAndRecord(otherScope, params)
	require.NoError(t, err)
}

func TestCrossZoneZoneMessageAndRootAreDeterministic(t *testing.T) {
	params := testCrossZoneParams()
	first := testCrossZoneMessage(t, params, 1, 1, []byte("payload-1"))
	second := testCrossZoneMessage(t, params, 2, 2, []byte("payload-2"))

	zoneMsg, err := first.ZoneMessage()
	require.NoError(t, err)
	require.Equal(t, zonestypes.ZoneIDContract, zoneMsg.ZoneID)
	require.Equal(t, "contract.execute", zoneMsg.MessageType)
	require.Equal(t, uint64(1), zoneMsg.Sequence)
	require.Equal(t, bytes.ToLower([]byte(zoneMsg.PayloadHash)), []byte(zoneMsg.PayloadHash))
	require.NoError(t, zoneMsg.Validate(zonestypes.ZoneIDContract))

	rootA, err := ComputeCrossZoneMessageRoot([]CrossZoneMessageEnvelope{second, first}, params)
	require.NoError(t, err)
	rootB, err := ComputeCrossZoneMessageRoot([]CrossZoneMessageEnvelope{first, second}, params)
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)
	require.NoError(t, zonestypes.ValidateHash("cross-zone message root", rootA))
}

func TestCrossZoneEnvelopeRejectsInvalidDeadlineAndFee(t *testing.T) {
	params := testCrossZoneParams()
	msg := testCrossZoneMessage(t, params, 1, 1, []byte("payload"))

	lowFee := msg
	lowFee.FeeLimit = sdkmath.ZeroInt()
	lowFee.MessageID = nil
	lowFee.PayloadHash = nil
	_, err := NewCrossZoneMessageEnvelope(lowFee, params)
	require.ErrorContains(t, err, "fee limit")

	expired := msg
	expired.Deadline = expired.CreatedHeight - 1
	expired.MessageID = nil
	expired.PayloadHash = nil
	_, err = NewCrossZoneMessageEnvelope(expired, params)
	require.ErrorContains(t, err, "deadline")
}

func testCrossZoneParams() CrossZoneMessageParams {
	return CrossZoneMessageParams{
		MaxPayloadSize:	256,
		MinGasLimit:	1,
		MaxGasLimit:	1_000,
		MinFeeLimit:	sdkmath.NewInt(3),
	}
}

func testCrossZoneMessage(t *testing.T, params CrossZoneMessageParams, nonce uint64, sequence uint64, payload []byte) CrossZoneMessageEnvelope {
	t.Helper()
	msg, err := NewCrossZoneMessageEnvelope(CrossZoneMessageEnvelope{
		SourceZone:		zonestypes.ZoneIDFinancial,
		DestinationZone:	zonestypes.ZoneIDContract,
		Sender:			addr(1),
		Recipient:		addr(2),
		Value:			sdkmath.NewInt(10),
		Opcode:			"contract.execute",
		Payload:		payload,
		GasLimit:		10,
		Deadline:		100,
		Nonce:			nonce,
		SourceSequence:		sequence,
		RouteID:		"route-fin-contract",
		Bounce:			true,
		FeeLimit:		sdkmath.NewInt(3),
		CreatedHeight:		20,
		AuthScope:		"owner",
	}, params)
	require.NoError(t, err)
	return msg
}
