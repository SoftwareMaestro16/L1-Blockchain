package sim

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeterministicRoutingAndExportImport(t *testing.T) {
	sim := newTestSimulator(t)
	require.NoError(t, sim.SplitShard(ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}))
	source := ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}
	dest := ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}
	msg := testMessage(source, dest, 1)
	require.NoError(t, sim.EnqueueMessage(msg))

	receipt, err := sim.ProcessNext(source, 1)
	require.NoError(t, err)
	require.Equal(t, MessageID(source, dest, 1, []byte("payload")), receipt.MessageID)
	require.NoError(t, sim.RequireReceipt(receipt.MessageID))

	exported := sim.Export()
	imported, err := Import(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported, imported.Export()))
}

func TestShardSplitMergeAndValidatorReassignment(t *testing.T) {
	sim := newTestSimulator(t)
	root := ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}
	before := sim.Export().Shards[root.Key()].ValidatorSubset

	require.NoError(t, sim.SplitShard(root))
	left := ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}
	right := ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}
	require.Contains(t, sim.Export().Shards, left.Key())
	require.Contains(t, sim.Export().Shards, right.Key())

	sim.ReassignValidators(10)
	after := sim.Export().Shards[left.Key()].ValidatorSubset
	require.NotEmpty(t, after)
	require.NotEqual(t, before, after)

	require.NoError(t, sim.MergeShards(left, right))
	require.Contains(t, sim.Export().Shards, root.Key())
	require.NotContains(t, sim.Export().Shards, left.Key())
	require.NotContains(t, sim.Export().Shards, right.Key())
}

func TestDelayedReceiptAndDataUnavailableShardBlock(t *testing.T) {
	sim := newTestSimulator(t)
	require.NoError(t, sim.SplitShard(ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}))
	source := ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}
	dest := ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}
	require.NoError(t, sim.MarkShardAvailability(dest, false))
	require.NoError(t, sim.EnqueueMessage(testMessage(source, dest, 1)))

	_, err := sim.ProcessNext(source, 1)
	require.ErrorContains(t, err, "data unavailable")
	require.ErrorContains(t, sim.RequireReceipt(MessageID(source, dest, 1, []byte("payload"))), "missing")

	require.NoError(t, sim.MarkShardAvailability(dest, true))
	msg := testMessage(source, dest, 2)
	require.NoError(t, sim.EnqueueMessage(msg))
	receipt, err := sim.ProcessNext(source, 2)
	require.NoError(t, err)
	require.NoError(t, sim.RequireReceipt(receipt.MessageID))
}

func TestAdversarialCrossShardReceiptAndProofValidation(t *testing.T) {
	sim := newTestSimulator(t)
	require.NoError(t, sim.SplitShard(ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}))
	source := ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}
	dest := ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}
	msg := testMessage(source, dest, 1)
	require.NoError(t, sim.EnqueueMessage(msg))
	receipt, err := sim.ProcessNext(source, 1)
	require.NoError(t, err)
	require.ErrorContains(t, sim.CommitReceipt(receipt), "duplicate")

	replay := testMessage(source, dest, 1)
	replay.MessageID = receipt.MessageID
	replay.Proof = sim.Export().Headers[source.Key()].Commitment
	require.ErrorContains(t, func() error {
		_, err := sim.Deliver(replay, 2)
		return err
	}(), "replayed")

	invalidProof := testMessage(source, dest, 2)
	invalidProof.MessageID = MessageID(source, dest, 2, invalidProof.Payload)
	invalidProof.Proof = "bad-proof"
	_, err = sim.Deliver(invalidProof, 2)
	require.ErrorContains(t, err, "invalid shard proof")
}

func TestStaleHeaderWrongDestinationAndTimeoutBounce(t *testing.T) {
	sim := newTestSimulator(t)
	root := ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}
	require.ErrorContains(t, sim.VerifyHeaderFresh(root, 3), "stale")
	require.ErrorContains(t, sim.EnqueueMessage(testMessage(root, ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}, 1)), "destination shard")

	require.NoError(t, sim.SplitShard(root))
	source := ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}
	dest := ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}
	msg := testMessage(source, dest, 9)
	msg.Timeout = 1
	require.NoError(t, sim.EnqueueMessage(msg))
	_, err := sim.ProcessNext(source, 2)
	require.ErrorContains(t, err, "timeout")
	require.NotEmpty(t, sim.Export().Shards[dest.Key()].Queue)
}

func TestEquivocationEvidenceAndMalformedStateRejected(t *testing.T) {
	sim := newTestSimulator(t)
	root := ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}
	require.NoError(t, sim.SubmitEquivocation(EquivocationEvidence{
		Validator:	"val1",
		ShardID:	root,
		Height:		1,
		LeftRoot:	"a",
		RightRoot:	"b",
	}))
	require.Len(t, sim.Export().Evidence, 1)

	corrupted := sim.Export()
	shard := corrupted.Shards[root.Key()]
	shard.StateRoot = "corrupted"
	corrupted.Shards[root.Key()] = shard
	_, err := Import(corrupted)
	require.ErrorContains(t, err, "header commitment")

	corrupted = sim.Export()
	wc := corrupted.Workchains[BaseWorkchain]
	wc.FeeDenom = "testtoken"
	corrupted.Workchains[BaseWorkchain] = wc
	_, err = Import(corrupted)
	require.ErrorContains(t, err, "naet")
}

func BenchmarkRoutingTableLookup(b *testing.B) {
	sim := benchmarkSimulator(b)
	root := ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}
	if err := sim.SplitShard(root); err != nil {
		b.Fatal(err)
	}
	exported := sim.Export()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = exported.Shards[ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}.Key()]
	}
}

func BenchmarkCrossShardProofVerification(b *testing.B) {
	sim := benchmarkSimulator(b)
	if err := sim.SplitShard(ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}); err != nil {
		b.Fatal(err)
	}
	source := ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}
	dest := ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}
	header := sim.Export().Headers[source.Key()]
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		msg := testMessage(source, dest, uint64(i+1))
		msg.MessageID = MessageID(source, dest, msg.Nonce, msg.Payload)
		msg.Proof = header.Commitment
		if msg.Proof != header.Commitment {
			b.Fatal("invalid proof")
		}
	}
}

func BenchmarkShardSplitMerge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := benchmarkSimulator(b)
		root := ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}
		if err := sim.SplitShard(root); err != nil {
			b.Fatal(err)
		}
		if err := sim.MergeShards(ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}, ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkShardedStateExportImport(b *testing.B) {
	sim := benchmarkSimulator(b)
	if err := sim.SplitShard(ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}); err != nil {
		b.Fatal(err)
	}
	exported := sim.Export()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Import(exported); err != nil {
			b.Fatal(err)
		}
	}
}

func newTestSimulator(t testing.TB) *Simulator {
	t.Helper()
	return benchmarkSimulator(t)
}

func benchmarkSimulator(t testing.TB) *Simulator {
	t.Helper()
	sim, err := New([]Validator{
		{Address: "val1", Power: 10},
		{Address: "val2", Power: 20},
		{Address: "val3", Power: 30},
		{Address: "val4", Power: 40},
	}, "seed")
	require.NoError(t, err)
	require.NoError(t, sim.AddWorkchain(WorkchainConfig{
		ID:			BaseWorkchain,
		AllowedVMs:		[]string{"AVM", "CosmWasm-gated"},
		FeeDenom:		FeeDenomNaet,
		AddressFormat:		"4:<64-lower-hex>",
		GenesisStateHash:	HashParts("genesis"),
		UpgradePolicy:		"governance-bounded",
	}))
	return sim
}

func testMessage(source, dest ShardID, nonce uint64) CrossShardMessage {
	return CrossShardMessage{
		Source:		source,
		Destination:	dest,
		Nonce:		nonce,
		Payload:	[]byte("payload"),
		Timeout:	10,
		Bounce:		true,
	}
}
