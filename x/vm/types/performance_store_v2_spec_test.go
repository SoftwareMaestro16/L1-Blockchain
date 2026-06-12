package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMStoreV2MessageRecordsCompactAndPayloadByHash(t *testing.T) {
	inline := testAVMStoreV2Message(t, "inline", 1, []byte("small"))
	record, payload, err := NewAVMStoreV2MessageRecord(inline)
	require.NoError(t, err)
	require.NoError(t, record.Validate())
	require.True(t, record.PayloadInline)
	require.Empty(t, record.PayloadRefKey)
	require.True(t, payload.Inline)
	require.LessOrEqual(t, record.CompactBytes, uint32(AVMStoreV2CompactMessageMaxBytes))

	largePayload := make([]byte, AVMStoreV2PayloadInlineMaxBytes+1)
	for i := range largePayload {
		largePayload[i] = byte(i)
	}
	external := testAVMStoreV2Message(t, "external", 2, largePayload)
	record, payload, err = NewAVMStoreV2MessageRecord(external)
	require.NoError(t, err)
	require.False(t, record.PayloadInline)
	require.Equal(t, AVMStoreV2PayloadKey(external.PayloadHash), record.PayloadRefKey)
	require.False(t, payload.Inline)
	require.Equal(t, uint32(len(largePayload)), payload.PayloadSize)
}

func TestAVMStoreV2PrefixesBucketsAndPruning(t *testing.T) {
	actorPrefix, err := NewAVMStoreV2ActorStatePrefix("actor-a")
	require.NoError(t, err)
	require.Equal(t, ActorStateKeyPrefix("actor-a"), actorPrefix.Prefix)

	contractPrefix, err := NewAVMStoreV2ContractStatePrefix("contract-a")
	require.NoError(t, err)
	require.Equal(t, AVMStatePrefixContractStorage+"/contract-a/", contractPrefix.Prefix)

	msg := testAVMStoreV2Message(t, "delayed", 3, []byte("payload"))
	bucket, err := NewAVMStoreV2DelayedQueueBucket(msg.DestinationZone, AVMMessageScheduledHeight(msg), []string{msg.ID})
	require.NoError(t, err)
	require.Contains(t, bucket.BucketKey, "delayed")
	require.Contains(t, bucket.BucketKey, "0000000000000000011")

	consumed := AVMAsyncReplayTombstone{MessageID: engineHash("old-consumed"), ConsumedHeight: 10}
	expired := AVMExpiredNonceTombstone{
		ChainID:	"aetra-1",
		SourceZone:	zonestypes.ZoneIDApplication,
		Sender:		"alice",
		SenderNonce:	9,
		MessageID:	engineHash("old-expired"),
		ExpiryHeight:	12,
	}
	expired.TombstoneHash = ComputeAVMExpiredNonceTombstoneHash(expired)
	store, err := NewAVMReplayTombstoneStore(AVMReplayTombstoneStore{
		ConsumedTombstones:	[]AVMAsyncReplayTombstone{consumed},
		ExpiredNonces:		[]AVMExpiredNonceTombstone{expired},
	})
	require.NoError(t, err)
	pruning, err := NewAVMStoreV2TombstonePruningPlan(store, 100, 50)
	require.NoError(t, err)
	require.Equal(t, uint64(50), pruning.RetainAfterHeight)
	require.Contains(t, pruning.PrunableConsumedIDs, consumed.MessageID)
	require.Len(t, pruning.PrunableExpiredScopes, 1)
}

func TestAVMStoreV2LayoutRootCommitsStrategy(t *testing.T) {
	msg := testAVMStoreV2Message(t, "layout", 4, []byte("payload"))
	record, payload, err := NewAVMStoreV2MessageRecord(msg)
	require.NoError(t, err)
	actorPrefix, err := NewAVMStoreV2ActorStatePrefix("actor-a")
	require.NoError(t, err)
	contractPrefix, err := NewAVMStoreV2ContractStatePrefix("contract-a")
	require.NoError(t, err)
	bucket, err := NewAVMStoreV2DelayedQueueBucket(msg.DestinationZone, AVMMessageScheduledHeight(msg), []string{msg.ID})
	require.NoError(t, err)
	store, err := NewAVMReplayTombstoneStore(AVMReplayTombstoneStore{})
	require.NoError(t, err)
	pruning, err := NewAVMStoreV2TombstonePruningPlan(store, 20, 10)
	require.NoError(t, err)

	layout, err := NewAVMStoreV2LayoutStrategy(AVMStoreV2LayoutStrategy{
		MessageRecords:		[]AVMStoreV2MessageRecord{record},
		PayloadRecords:		[]AVMStoreV2PayloadRecord{payload},
		ActorPrefixes:		[]AVMStoreV2ActorStatePrefix{actorPrefix},
		ContractPrefixes:	[]AVMStoreV2ContractStatePrefix{contractPrefix},
		DelayedBuckets:		[]AVMStoreV2DelayedQueueBucket{bucket},
		PruningPlan:		pruning,
	})
	require.NoError(t, err)
	require.NoError(t, layout.Validate())
	require.Equal(t, ComputeAVMStoreV2LayoutRoot(layout), layout.LayoutRoot)

	bad := layout
	bad.MessageRecords[0].CompactBytes = AVMStoreV2CompactMessageMaxBytes + 1
	bad.MessageRecords[0].RecordHash = ComputeAVMStoreV2MessageRecordHash(bad.MessageRecords[0])
	bad.LayoutRoot = ComputeAVMStoreV2LayoutRoot(bad)
	require.ErrorContains(t, bad.Validate(), "compact message")
}

func testAVMStoreV2Message(t *testing.T, source string, nonce uint64, payload []byte) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage(source, zonestypes.ZoneIDApplication, "contract", zonestypes.ZoneIDContract, nonce, 10)
	msg.Payload = append([]byte(nil), payload...)
	msg.PayloadHash = ""
	msg.DelayHeight = 1
	msg.ExpiryHeight = 30
	msg.RetryPolicy = DefaultAVMRetryPolicy(30)
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}
