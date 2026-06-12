package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestBuildMemoStoreRecordStoresFullMemoOrHashOnly(t *testing.T) {
	params := DefaultMemoParams()
	base := testMemoRecordBase()
	metadata := TxMetadata{Memo: "pay alice", MemoVisible: true}

	record, event, err := BuildMemoStoreRecord(metadata, params, DefaultMemoStoragePolicy(params), base)
	require.NoError(t, err)
	require.Equal(t, "pay alice", record.Memo)
	require.Equal(t, MemoHash("pay alice"), record.MemoHash)
	require.Equal(t, record.Memo, event.Memo)
	require.Equal(t, record.MemoHash, event.MemoHash)

	record, event, err = BuildMemoStoreRecord(metadata, params, MemoStoragePolicy{Mode: StoragePolicyHashOnlyOnchain, MaxOnChainBytes: params.MaxMemoBytes}, base)
	require.NoError(t, err)
	require.Empty(t, record.Memo)
	require.Equal(t, MemoHash("pay alice"), record.MemoHash)
	require.Empty(t, event.Memo)
	require.Equal(t, record.MemoHash, event.MemoHash)
}

func TestBuildMemoStoreRecordRejectsInvalidInputs(t *testing.T) {
	params := DefaultMemoParams()
	base := testMemoRecordBase()
	_, _, err := BuildMemoStoreRecord(TxMetadata{Memo: string([]byte{0xff})}, params, DefaultMemoStoragePolicy(params), base)
	require.ErrorContains(t, err, "UTF-8")

	oversized := TxMetadata{Memo: strings.Repeat("a", int(params.MaxMemoBytes)+1)}
	_, _, err = BuildMemoStoreRecord(oversized, params, DefaultMemoStoragePolicy(params), base)
	require.Error(t, err)

	base.Sender = make([]byte, 20)
	_, _, err = BuildMemoStoreRecord(TxMetadata{}, params, DefaultMemoStoragePolicy(params), base)
	require.ErrorContains(t, err, "memo sender")
}

func TestIndexMemoRecords(t *testing.T) {
	params := DefaultMemoParams()
	base := testMemoRecordBase()
	record, _, err := BuildMemoStoreRecord(TxMetadata{Memo: "domain payment"}, params, DefaultMemoStoragePolicy(params), base)
	require.NoError(t, err)
	other := record
	other.TxHash = []byte("tx2")
	other.Sender = addr(3)
	other.Receiver = addr(4)
	other.RelatedDomain = "bob.aet"
	other.BlockHeight = 8

	index, err := IndexMemoRecords([]MemoStoreRecord{other, record})
	require.NoError(t, err)
	require.Len(t, index.ByTxHash[string(record.TxHash)], 1)
	require.Equal(t, record.TxHash, index.ByTxHash[string(record.TxHash)][0].TxHash)
	require.Len(t, index.BySender[string(record.Sender)], 1)
	require.Len(t, index.ByReceiver[string(record.Receiver)], 1)
	require.Len(t, index.ByDomain["alice.aet"], 1)
	require.Len(t, index.ByAsset[AssetTypeNative], 2)
	require.Len(t, index.ByEventType[EventTypeResolverPayment], 2)
	require.Equal(t, other.TxHash, index.ByAsset[AssetTypeNative][0].TxHash)
	require.Equal(t, record.TxHash, index.ByAsset[AssetTypeNative][1].TxHash)
	require.False(t, MemoSearchIndexAffectsConsensus())
}

func TestDeterministicMemoEvent(t *testing.T) {
	params := DefaultMemoParams()
	record, event, err := BuildMemoStoreRecord(TxMetadata{Memo: "hello"}, params, DefaultMemoStoragePolicy(params), testMemoRecordBase())
	require.NoError(t, err)
	require.Equal(t, event, DeterministicMemoEvent(record))
	require.Equal(t, []byte("tx1"), event.TxHash)
	require.Equal(t, addr(1), event.From)
	require.Equal(t, addr(2), event.To)
	require.Equal(t, "alice.aet", event.Domain)
	require.Equal(t, "hello", event.Memo)
}

func TestMemoFeeReputationCostOrdering(t *testing.T) {
	params := DefaultMemoParams()
	metadata := TxMetadata{Memo: "hello"}
	low, _, err := MemoFee(metadata, params, 10, DefaultCongestionBps)
	require.NoError(t, err)
	high, _, err := MemoFee(metadata, params, 80, DefaultCongestionBps)
	require.NoError(t, err)
	require.True(t, low.GT(high))
}

func testMemoRecordBase() MemoStoreRecord {
	return MemoStoreRecord{
		TxHash:		[]byte("tx1"),
		Sender:		addr(1),
		Receiver:	addr(2),
		AssetType:	AssetTypeNative,
		RelatedDomain:	"alice.aet",
		EventType:	EventTypeResolverPayment,
		BlockHeight:	9,
		TimestampUnix:	100,
	}
}

func addr(seed byte) sdk.AccAddress {
	out := make([]byte, 20)
	out[19] = seed
	return sdk.AccAddress(out)
}
