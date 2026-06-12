package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMSecurityModelRequiresSection13Assumptions(t *testing.T) {
	model, err := DefaultAVMSecurityModel(engineHash("store-root"))
	require.NoError(t, err)
	require.NoError(t, model.Validate())
	require.Equal(t, ComputeAVMSecurityModelHash(model), model.ModelHash)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionCometBFTFinality)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionDeterministicExecution)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionNoExternalCalls)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionGasBounding)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionReplayProtection)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionZoneStateIsolation)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionActorStateIsolation)
	require.Contains(t, model.Assumptions, AVMSecurityAssumptionStoreRootCommitments)

	missing := model
	missing.Assumptions = missing.Assumptions[:len(missing.Assumptions)-1]
	missing.ModelHash = ComputeAVMSecurityModelHash(missing)
	require.ErrorContains(t, missing.Validate(), "every security assumption")

	bad := model
	bad.Assumptions[0] = "wall_clock_time"
	bad.ModelHash = ComputeAVMSecurityModelHash(bad)
	require.ErrorContains(t, bad.Validate(), "invalid AVM security assumption")
}

func TestAVMReplayProtectionRecordCommitsReplayFields(t *testing.T) {
	msg := testAVMSecurityMessage(t, "alice", 7, 10)
	record, err := NewAVMReplayProtectionRecord(msg)
	require.NoError(t, err)
	require.NoError(t, record.Validate())
	require.Equal(t, msg.ChainID, record.ChainID)
	require.Equal(t, msg.SourceZone, record.SourceZone)
	require.Equal(t, msg.Source, record.Sender)
	require.Equal(t, msg.SenderNonce, record.SenderNonce)
	require.Equal(t, msg.ID, record.MessageID)
	require.Equal(t, msg.CreatedHeight, record.CreatedHeight)
	require.Equal(t, msg.ExpiryHeight, record.ExpiryHeight)
	require.Equal(t, ComputeAVMReplayProtectionRecordHash(record), record.RecordHash)

	mutated := record
	mutated.ExpiryHeight++
	require.ErrorContains(t, mutated.Validate(), "hash mismatch")
}

func TestAVMReplayNonceMustIncreaseAndConsumedCreatesTombstone(t *testing.T) {
	state := testAVMReplayState(t, "alice", 6)
	msg := testAVMSecurityMessage(t, "alice", 7, 10)

	next, tombstone, err := ConsumeAVMReplayMessage(state, msg, 12)
	require.NoError(t, err)
	require.Equal(t, msg.ID, tombstone.MessageID)
	require.Equal(t, uint64(7), next.LastNonce)
	require.Len(t, next.ConsumedTombstones, 1)
	require.NoError(t, next.Validate())

	_, _, err = ConsumeAVMReplayMessage(next, msg, 13)
	require.ErrorContains(t, err, "consumed message tombstone")

	lower := testAVMSecurityMessage(t, "alice", 6, 10)
	require.ErrorContains(t, ValidateAVMReplaySubmission(next, lower, 12), "nonce must increase")
}

func TestAVMReplayExpiredNonceCannotBeResubmitted(t *testing.T) {
	state := testAVMReplayState(t, "alice", 4)
	expired := testAVMSecurityMessage(t, "alice", 5, 10)

	next, nonceTombstone, err := ExpireAVMReplayMessage(state, expired, expired.ExpiryHeight+1)
	require.NoError(t, err)
	require.NoError(t, nonceTombstone.Validate())
	require.NoError(t, next.Validate())

	resubmit := testAVMSecurityMessage(t, "alice", 5, 20)
	resubmit.Destination = "contract-b"
	resubmit.Payload = []byte("different")
	resubmit.PayloadHash = ""
	resubmit, err = NewAVMAsyncMessage(resubmit)
	require.NoError(t, err)
	require.NotEqual(t, expired.ID, resubmit.ID)
	require.ErrorContains(t, ValidateAVMReplaySubmission(next, resubmit, 21), "expired nonce tombstone")
}

func TestAVMCrossZoneReplayBindsSourceAndDestinationZones(t *testing.T) {
	msg := testAVMSecurityMessage(t, "alice", 9, 10)
	require.NotEqual(t, msg.SourceZone, msg.DestinationZone)
	require.NoError(t, ValidateAVMCrossZoneReplayBinding(msg))

	otherDestination := testAVMSecurityMessage(t, "alice", 9, 10)
	otherDestination.DestinationZone = zonestypes.ZoneIDIdentity
	otherDestination.PayloadHash = ""
	otherDestination, err := NewAVMAsyncMessage(otherDestination)
	require.NoError(t, err)
	require.NotEqual(t, msg.ID, otherDestination.ID)

	tampered := msg
	tampered.DestinationZone = zonestypes.ZoneIDIdentity
	require.ErrorContains(t, ValidateAVMCrossZoneReplayBinding(tampered), "id mismatch")

	sameEndpoint := msg
	sameEndpoint.Destination = sameEndpoint.Source
	sameEndpoint.PayloadHash = ""
	sameEndpoint, err = NewAVMAsyncMessage(sameEndpoint)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateAVMCrossZoneReplayBinding(sameEndpoint), "endpoints must differ")
}

func TestAVMReplaySubmissionRejectsScopeMismatchAndExpiredHeight(t *testing.T) {
	state := testAVMReplayState(t, "alice", 1)
	msg := testAVMSecurityMessage(t, "bob", 2, 10)
	require.ErrorContains(t, ValidateAVMReplaySubmission(state, msg, 11), "scope mismatch")

	msg = testAVMSecurityMessage(t, "alice", 2, 10)
	require.ErrorContains(t, ValidateAVMReplaySubmission(state, msg, msg.ExpiryHeight+1), "message is expired")
}

func testAVMReplayState(t *testing.T, sender string, lastNonce uint64) AVMReplayNonceState {
	t.Helper()
	state, err := NewAVMReplayNonceState(AVMReplayNonceState{
		ChainID:	"aetra-1",
		SourceZone:	zonestypes.ZoneIDApplication,
		Sender:		sender,
		LastNonce:	lastNonce,
	})
	require.NoError(t, err)
	return state
}

func testAVMSecurityMessage(t *testing.T, sender string, nonce uint64, createdHeight uint64) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage(sender, zonestypes.ZoneIDApplication, "contract-a", zonestypes.ZoneIDContract, nonce, createdHeight)
	msg.ValueNAET = 3
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}
