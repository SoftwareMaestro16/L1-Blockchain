package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMAsyncMessageDerivesDeterministicIDFromSection52Fields(t *testing.T) {
	msg, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 7, 10))
	require.NoError(t, err)
	require.NoError(t, msg.Validate())
	require.Equal(t, ComputeAVMAsyncPayloadHash([]byte("payload")), msg.PayloadHash)
	require.Equal(t, DeriveAVMAsyncMessageID(msg), msg.ID)

	same, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 7, 10))
	require.NoError(t, err)
	require.Equal(t, msg.ID, same.ID)

	changedNonce, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 8, 10))
	require.NoError(t, err)
	require.NotEqual(t, msg.ID, changedNonce.ID)

	changedHeight, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 7, 11))
	require.NoError(t, err)
	require.NotEqual(t, msg.ID, changedHeight.ID)
}

func TestAVMAsyncMessageValidatesExtendedFieldsAndProofs(t *testing.T) {
	msg := testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 1, 10)
	msg.SourceActorOptional = "actor-a"
	msg.DestinationActorOptional = "actor-b"
	msg.RouteHintOptional = "contract.queue"
	msg.AuthProofOptional = engineHash("auth")
	msg.StateProofOptional = engineHash("state")
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	require.NoError(t, built.Validate())

	badHash := built
	badHash.PayloadHash = engineHash("wrong")
	require.ErrorContains(t, badHash.Validate(), "payload hash mismatch")

	badExpiry := testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 2, 10)
	badExpiry.ExpiryHeight = badExpiry.CreatedHeight
	_, err = NewAVMAsyncMessage(badExpiry)
	require.ErrorContains(t, err, "expiry height")

	badProof := built
	badProof.AuthProofOptional = "not-a-root"
	require.ErrorContains(t, badProof.Validate(), "auth proof")
}

func TestAVMAsyncMessageRegistryRejectsDuplicateIDAndNonceScope(t *testing.T) {
	msgA, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 1, 10))
	require.NoError(t, err)
	msgB, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "carol", zonestypes.ZoneIDContract, 1, 10))
	require.NoError(t, err)
	require.NotEqual(t, msgA.ID, msgB.ID)
	require.Equal(t,
		AsyncMessageNonceScope(msgA.SourceZone, msgA.Source, msgA.SenderNonce),
		AsyncMessageNonceScope(msgB.SourceZone, msgB.Source, msgB.SenderNonce),
	)

	registry := AVMAsyncMessageRegistry{Messages: []AVMAsyncMessage{msgA, msgA}}
	require.ErrorContains(t, registry.Validate(), "duplicate async message id")

	registry = AVMAsyncMessageRegistry{Messages: []AVMAsyncMessage{msgA, msgB}}
	require.ErrorContains(t, registry.Validate(), "duplicate async sender nonce scope")

	msgOtherSender, err := NewAVMAsyncMessage(testAVMAsyncMessage("dave", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 1, 10))
	require.NoError(t, err)
	msgOtherZone, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDFinancial, "bob", zonestypes.ZoneIDContract, 1, 10))
	require.NoError(t, err)
	require.NoError(t, (AVMAsyncMessageRegistry{Messages: []AVMAsyncMessage{msgA, msgOtherSender, msgOtherZone}}).Validate())
}

func TestAVMAsyncMessageRegistryRequiresReplayTombstoneForConsumedMessages(t *testing.T) {
	consumed, err := NewAVMAsyncMessage(testAVMAsyncMessage("alice", zonestypes.ZoneIDApplication, "bob", zonestypes.ZoneIDContract, 1, 10))
	require.NoError(t, err)

	registry := AVMAsyncMessageRegistry{
		ConsumedMessageIDs: []string{consumed.ID},
	}
	require.ErrorContains(t, registry.Validate(), "must create replay tombstone")

	registry = AVMAsyncMessageRegistry{
		ConsumedMessageIDs:	[]string{consumed.ID},
		ReplayTombstones: []AVMAsyncReplayTombstone{{
			MessageID:	consumed.ID,
			ConsumedHeight:	12,
		}},
	}
	require.NoError(t, registry.Validate())

	replay := registry
	replay.Messages = []AVMAsyncMessage{consumed}
	require.ErrorContains(t, replay.Validate(), "replay tombstoned")
}

func testAVMAsyncMessage(source string, sourceZone zonestypes.ZoneID, destination string, destinationZone zonestypes.ZoneID, nonce uint64, createdHeight uint64) AVMAsyncMessage {
	return AVMAsyncMessage{
		ChainID:		"aetra-1",
		Source:			source,
		Destination:		destination,
		Payload:		[]byte("payload"),
		GasLimit:		100,
		DelayHeight:		1,
		ExpiryHeight:		createdHeight + 10,
		RetryPolicy:		DefaultAVMRetryPolicy(createdHeight + 10),
		BounceFlag:		true,
		SourceZone:		sourceZone,
		DestinationZone:	destinationZone,
		SenderNonce:		nonce,
		PayloadType:		"contract.call",
		ValueNAET:		1,
		ForwardingFee:		1,
		Priority:		3,
		CreatedHeight:		createdHeight,
	}
}
