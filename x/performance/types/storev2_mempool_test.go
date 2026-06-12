package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreV2BuildsBoundedPrefixProofAndZoneRoot(t *testing.T) {
	state := sampleStoreV2State(t)

	accountPrefix := "storev2/object/account-zone/shard-a/account"
	entries, err := StoreV2BoundedRangeScan(state, accountPrefix, "", 1)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Contains(t, entries[0].Key, "alice")

	_, err = StoreV2BoundedRangeScan(state, accountPrefix, "", MaxStoreV2RangeLimit+1)
	require.ErrorContains(t, err, "limit")

	proof, err := GenerateStoreV2PrefixProof(state, "storev2/kv/account-zone/shard-a/identity/alice.aet", "", 8)
	require.NoError(t, err)
	require.Len(t, proof.Entries, 2)
	require.NoError(t, proof.Validate(state.RootHash))

	root, err := BuildStoreV2ZoneRoot("account-zone", []StoreV2ShardState{state})
	require.NoError(t, err)
	require.NotEmpty(t, root.ZoneRootHash)
	require.NoError(t, root.Validate())
}

func TestStoreV2BenchmarkCoverageRequiresAllOperations(t *testing.T) {
	results := make([]StoreV2BenchmarkResult, 0, len(RequiredStoreV2Benchmarks()))
	for _, operation := range RequiredStoreV2Benchmarks() {
		results = append(results, StoreV2BenchmarkResult{
			Operation:		operation,
			Samples:		1,
			Operations:		1,
			MaxRangeLimit:		8,
			ObservedRootHash:	hashStrings("benchmark-root", string(operation)),
		})
	}
	require.NoError(t, ValidateStoreV2BenchmarkCoverage(results))

	require.ErrorContains(t, ValidateStoreV2BenchmarkCoverage(results[:len(results)-1]), "required")
}

func TestSeparatedMempoolUsesZoneShardLanesAndDeterministicOrdering(t *testing.T) {
	txs := []SeparatedMempoolTx{
		sampleMempoolTx("tx-low", "alice", "financial", "shard-a", "acct-a", MempoolClassPayment, "2", 100),
		sampleMempoolTx("tx-high", "bob", "financial", "shard-a", "acct-b", MempoolClassPayment, "9", 100),
		sampleMempoolTx("tx-msg-late", "carol", "identity", "shard-b", "name-a", MempoolClassMessage, "100", 120),
		sampleMempoolTx("tx-msg-early", "dave", "identity", "shard-b", "name-b", MempoolClassMessage, "1", 90),
		{TxID: "tx-unknown", Sender: "erin", FeeAmount: "3", GasWanted: 1, SizeBytes: 128, CreatedHeight: 10, ExpiryHeight: 110},
	}

	snapshot, err := BuildSeparatedMempoolSnapshot(50, txs, MempoolSeparationLimits{MaxPerSender: 8, MaxPerTargetObject: 8})
	require.NoError(t, err)
	require.NoError(t, snapshot.Validate(MempoolSeparationLimits{MaxPerSender: 8, MaxPerTargetObject: 8}.Normalize()))
	require.Len(t, snapshot.Lanes, 3)

	paymentLane := snapshot.Lanes[0]
	require.Equal(t, "financial", paymentLane.ZoneID)
	require.Equal(t, MempoolClassPayment, paymentLane.MessageClass)
	require.Equal(t, "tx-high", paymentLane.Transactions[0].TxID)
	require.Equal(t, "tx-low", paymentLane.Transactions[1].TxID)

	messageLane := snapshot.Lanes[1]
	require.Equal(t, MempoolClassMessage, messageLane.MessageClass)
	require.Equal(t, "tx-msg-early", messageLane.Transactions[0].TxID)
	require.Equal(t, "tx-msg-late", messageLane.Transactions[1].TxID)

	systemLane := snapshot.Lanes[2]
	require.Equal(t, SystemMempoolZoneID, systemLane.ZoneID)
	require.Equal(t, SystemMempoolShardID, systemLane.ShardID)
	require.Equal(t, MempoolClassSystem, systemLane.MessageClass)
}

func TestSeparatedMempoolRejectsExpiryAndDoSLimits(t *testing.T) {
	_, err := BuildSeparatedMempoolSnapshot(50, []SeparatedMempoolTx{
		sampleMempoolTx("expired", "alice", "financial", "shard-a", "acct-a", MempoolClassPayment, "1", 49),
	}, MempoolSeparationLimits{})
	require.ErrorContains(t, err, "expired")

	_, err = BuildSeparatedMempoolSnapshot(50, []SeparatedMempoolTx{
		sampleMempoolTx("tx-a", "alice", "financial", "shard-a", "acct-a", MempoolClassPayment, "1", 100),
		sampleMempoolTx("tx-b", "alice", "financial", "shard-a", "acct-b", MempoolClassPayment, "1", 100),
	}, MempoolSeparationLimits{MaxPerSender: 1, MaxPerTargetObject: 8})
	require.ErrorContains(t, err, "per-sender")

	_, err = BuildSeparatedMempoolSnapshot(50, []SeparatedMempoolTx{
		sampleMempoolTx("tx-c", "alice", "financial", "shard-a", "acct-a", MempoolClassPayment, "1", 100),
		sampleMempoolTx("tx-d", "bob", "financial", "shard-a", "acct-a", MempoolClassPayment, "1", 100),
	}, MempoolSeparationLimits{MaxPerSender: 8, MaxPerTargetObject: 1})
	require.ErrorContains(t, err, "per-target")
}

func sampleStoreV2State(t *testing.T) StoreV2ShardState {
	t.Helper()
	records := []StoreV2ObjectRecord{
		storeV2Record("account-zone", "shard-a", StoreV2RecordAccount, "account/alice", "balance:100", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordDomain, "identity/alice.aet", "owner:alice", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordContract, "contract/escrow", "code:escrow", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordChannel, "channel/alice-bob", "settled:false", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordPool, "pool/aet-usd", "x:100:y:200", 1),
	}
	fields := []StoreV2KVField{
		storeV2Field("account-zone", "shard-a", "identity/alice.aet", "resolver/address", "aet1alice", 1),
		storeV2Field("account-zone", "shard-a", "identity/alice.aet", "resolver/parent", "root.aet", 1),
		storeV2Field("account-zone", "shard-a", "contract/escrow", "storage/slot-1", "slot-value", 1),
	}
	state, err := BuildStoreV2ShardState("account-zone", "shard-a", records, fields)
	require.NoError(t, err)
	return state
}

func storeV2Record(zoneID, shardID string, kind StoreV2RecordKind, objectKey, value string, version uint64) StoreV2ObjectRecord {
	record := StoreV2ObjectRecord{
		ZoneID:		zoneID,
		ShardID:	shardID,
		Kind:		kind,
		ObjectKey:	objectKey,
		ValueHash:	hashStrings("storev2-value", value),
		Version:	version,
		UpdatedHeight:	20,
		SizeBytes:	uint32(len(value)),
	}
	record.RecordHash = ComputeStoreV2RecordHash(record)
	return record
}

func storeV2Field(zoneID, shardID, objectKey, fieldPath, value string, version uint64) StoreV2KVField {
	field := StoreV2KVField{
		ZoneID:		zoneID,
		ShardID:	shardID,
		ObjectKey:	objectKey,
		FieldPath:	fieldPath,
		ValueHash:	hashStrings("storev2-field-value", value),
		Version:	version,
		UpdatedHeight:	20,
	}
	field.FieldHash = ComputeStoreV2FieldHash(field)
	return field
}

func sampleMempoolTx(txID, sender, zoneID, shardID, objectID string, class MempoolMessageClass, fee string, expiry uint64) SeparatedMempoolTx {
	tx := SeparatedMempoolTx{
		TxID:		txID,
		Sender:		sender,
		TargetZoneID:	zoneID,
		TargetShardID:	shardID,
		TargetObject:	objectID,
		RouteKey:	objectID,
		MessageClass:	class,
		FeeAmount:	fee,
		GasWanted:	100,
		SizeBytes:	256,
		CreatedHeight:	10,
		ExpiryHeight:	expiry,
		PreResolved:	true,
	}
	tx.TxHash = ComputeSeparatedMempoolTxHash(tx)
	return tx
}
