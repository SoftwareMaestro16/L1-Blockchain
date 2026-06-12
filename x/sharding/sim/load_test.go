package sim

import (
	"sort"
	"testing"

	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	"github.com/stretchr/testify/require"
)

func TestLowLoadKeepsOneShard(t *testing.T) {
	sim := newTestSimulator(t)
	transition, err := sim.UpdateLoadAndShards(immediatePolicy(), loadtypes.Metrics{}, 1)
	require.NoError(t, err)

	require.Equal(t, loadtypes.LoadBandLow, transition.LoadResult.Band)
	require.Equal(t, uint32(1), transition.AppliedShardCount)
	require.Equal(t, uint32(1), sim.ActiveShardCount(BaseWorkchain))
}

func TestMediumLoadActivatesPartialSharding(t *testing.T) {
	sim := newTestSimulator(t)
	policy := immediatePolicy()
	metrics := loadtypes.Metrics{
		CanonicalMempoolSize:	policy.LoadParams.TargetMempoolSize,
		UsedBlockGas:		policy.LoadParams.TargetBlockGas,
	}

	transition, err := sim.UpdateLoadAndShards(policy, metrics, 1)
	require.NoError(t, err)

	require.Equal(t, loadtypes.LoadBandMedium, transition.LoadResult.Band)
	require.Equal(t, uint32(2), transition.AppliedShardCount)
	require.Equal(t, []ShardID{
		{WorkchainID: BaseWorkchain, Prefix: "0"},
		{WorkchainID: BaseWorkchain, Prefix: "1"},
	}, sim.WorkchainShardIDs(BaseWorkchain))
}

func TestHighLoadActivatesFullSharding(t *testing.T) {
	sim := newTestSimulator(t)
	policy := immediatePolicy()

	transition, err := sim.UpdateLoadAndShards(policy, saturatedLoadMetrics(policy.LoadParams), 1)
	require.NoError(t, err)

	require.Equal(t, loadtypes.LoadBandHigh, transition.LoadResult.Band)
	require.Equal(t, uint32(4), transition.AppliedShardCount)
	require.Equal(t, []ShardID{
		{WorkchainID: BaseWorkchain, Prefix: "00"},
		{WorkchainID: BaseWorkchain, Prefix: "01"},
		{WorkchainID: BaseWorkchain, Prefix: "10"},
		{WorkchainID: BaseWorkchain, Prefix: "11"},
	}, sim.WorkchainShardIDs(BaseWorkchain))
}

func TestSpikeLoadIsCappedByMaxDelta(t *testing.T) {
	sim := newTestSimulator(t)
	policy := immediatePolicy()
	policy.LoadParams.MaxDeltaBps = loadtypes.DefaultMaxDeltaBps

	transition, err := sim.UpdateLoadAndShards(policy, saturatedLoadMetrics(policy.LoadParams), 1)
	require.NoError(t, err)

	require.Equal(t, uint32(10_000), transition.LoadResult.RawLoadScoreBps)
	require.Equal(t, loadtypes.DefaultMaxDeltaBps, transition.LoadResult.LoadScoreBps)
	require.Equal(t, loadtypes.LoadBandLow, transition.LoadResult.Band)
	require.Equal(t, uint32(1), transition.AppliedShardCount)
}

func TestCooldownDelaysShardDeactivation(t *testing.T) {
	sim := newTestSimulator(t)
	policy := immediatePolicy()
	policy.CooldownBlocks = 3
	_, err := sim.UpdateLoadAndShards(policy, saturatedLoadMetrics(policy.LoadParams), 1)
	require.NoError(t, err)
	require.Equal(t, uint32(4), sim.ActiveShardCount(BaseWorkchain))

	transition, err := sim.UpdateLoadAndShards(policy, loadtypes.Metrics{}, 2)
	require.NoError(t, err)
	require.True(t, transition.CooldownStarted)
	require.False(t, transition.CooldownSatisfied)
	require.Equal(t, uint32(4), transition.AppliedShardCount)

	transition, err = sim.UpdateLoadAndShards(policy, loadtypes.Metrics{}, 5)
	require.NoError(t, err)
	require.True(t, transition.CooldownSatisfied)
	require.Equal(t, uint32(1), transition.AppliedShardCount)
}

func TestShardSplitKeepsEveryMessageExactlyOnce(t *testing.T) {
	sim := newTestSimulator(t)
	root := ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}
	ids := seedRootQueue(t, sim, root, []string{"msg-c", "msg-a", "msg-b", "msg-d"})

	require.NoError(t, sim.SplitShard(root))

	seen := map[string]uint32{}
	for _, shardID := range []ShardID{{WorkchainID: BaseWorkchain, Prefix: "0"}, {WorkchainID: BaseWorkchain, Prefix: "1"}} {
		for _, msg := range sim.Export().Shards[shardID.Key()].Queue {
			seen[msg.MessageID]++
		}
	}
	require.Len(t, seen, len(ids))
	for _, id := range ids {
		require.Equal(t, uint32(1), seen[id], id)
	}
}

func TestShardMergePreservesCanonicalMessageOrdering(t *testing.T) {
	sim := newTestSimulator(t)
	root := ShardID{WorkchainID: BaseWorkchain, Prefix: BaseShardID}
	require.NoError(t, sim.SplitShard(root))
	leftID := ShardID{WorkchainID: BaseWorkchain, Prefix: "0"}
	rightID := ShardID{WorkchainID: BaseWorkchain, Prefix: "1"}
	left := sim.state.Shards[leftID.Key()]
	right := sim.state.Shards[rightID.Key()]
	left.Queue = []CrossShardMessage{{MessageID: "msg-c"}, {MessageID: "msg-a"}}
	right.Queue = []CrossShardMessage{{MessageID: "msg-b"}}
	left.MessageQueueRoot = hashQueue(left.Queue)
	right.MessageQueueRoot = hashQueue(right.Queue)
	sim.state.Shards[leftID.Key()] = left
	sim.state.Shards[rightID.Key()] = right
	sim.commitHeader(left)
	sim.commitHeader(right)

	require.NoError(t, sim.MergeShards(leftID, rightID))

	merged := sim.Export().Shards[root.Key()].Queue
	require.Equal(t, []string{"msg-a", "msg-b", "msg-c"}, queueIDs(merged))
	require.Equal(t, hashQueue(merged), sim.Export().Shards[root.Key()].MessageQueueRoot)
}

func TestValidatorReassignmentTriggeredByRoutingEpochIsDeterministic(t *testing.T) {
	left := newTestSimulator(t)
	right := newTestSimulator(t)
	policy := immediatePolicy()
	policy.RoutingEpoch = 7
	require.NoError(t, left.SetActiveShardCount(BaseWorkchain, 4))
	require.NoError(t, right.SetActiveShardCount(BaseWorkchain, 4))

	leftTransition, err := left.UpdateLoadAndShards(policy, saturatedLoadMetrics(policy.LoadParams), 10)
	require.NoError(t, err)
	rightTransition, err := right.UpdateLoadAndShards(policy, saturatedLoadMetrics(policy.LoadParams), 10)
	require.NoError(t, err)

	require.True(t, leftTransition.ValidatorReassigned)
	require.True(t, rightTransition.ValidatorReassigned)
	require.Equal(t, left.Export().Shards, right.Export().Shards)
	require.Equal(t, uint64(10), left.Export().LoadStates[BaseWorkchain].LastValidatorEpochHeight)
}

func TestDataUnavailableShardCannotBeFinalizedForRouting(t *testing.T) {
	sim := newTestSimulator(t)
	require.NoError(t, sim.SetActiveShardCount(BaseWorkchain, 4))
	key := []byte("route-me")
	selected, err := sim.RouteWork(BaseWorkchain, key)
	require.NoError(t, err)

	require.NoError(t, sim.MarkShardAvailability(selected, false))
	_, err = sim.RouteWork(BaseWorkchain, key)
	require.ErrorContains(t, err, "data unavailable")
}

func TestExportImportPreservesLoadWindowsAndShardActivation(t *testing.T) {
	sim := newTestSimulator(t)
	policy := immediatePolicy()
	policy.RoutingEpoch = 3
	first, err := sim.UpdateLoadAndShards(policy, saturatedLoadMetrics(policy.LoadParams), 7)
	require.NoError(t, err)
	second, err := sim.UpdateLoadAndShards(policy, loadtypes.Metrics{}, 8)
	require.NoError(t, err)
	require.NotEqual(t, first.LoadResult.EMA.WindowHeight, uint64(0))
	require.Equal(t, uint64(2), second.LoadResult.EMA.WindowHeight)

	exported := sim.Export()
	imported, err := Import(exported)
	require.NoError(t, err)

	require.Equal(t, exported.LoadStates, imported.Export().LoadStates)
	require.Equal(t, exported.Shards, imported.Export().Shards)
	require.Equal(t, uint32(4), imported.ActiveShardCount(BaseWorkchain))
}

func BenchmarkShardSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := benchmarkSimulator(b)
		if err := sim.SetActiveShardCount(BaseWorkchain, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkShardMerge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := benchmarkSimulator(b)
		if err := sim.SetActiveShardCount(BaseWorkchain, 4); err != nil {
			b.Fatal(err)
		}
		if err := sim.SetActiveShardCount(BaseWorkchain, 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidatorReassignment(b *testing.B) {
	sim := benchmarkSimulator(b)
	if err := sim.SetActiveShardCount(BaseWorkchain, 4); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sim.ReassignValidators(uint64(i + 1))
	}
}

func immediatePolicy() ShardActivationPolicy {
	params := loadtypes.DefaultParams()
	params.AlphaNumerator = 1
	params.AlphaDenominator = 1
	params.MaxDeltaBps = loadtypes.BasisPoints
	return ShardActivationPolicy{
		WorkchainID:		BaseWorkchain,
		LoadParams:		params,
		PartialShardCount:	2,
		MaxShardCount:		4,
		CooldownBlocks:		2,
	}
}

func saturatedLoadMetrics(params loadtypes.Params) loadtypes.Metrics {
	return loadtypes.Metrics{
		CanonicalMempoolSize:		params.TargetMempoolSize,
		UsedBlockGas:			params.TargetBlockGas,
		AverageInclusionDelayBlocks:	params.TargetLatencyBlocks,
		FailedTxCount:			1,
		TotalTxCount:			1,
		ExecutionStepCount:		params.TargetExecutionSteps,
	}
}

func seedRootQueue(t *testing.T, sim *Simulator, root ShardID, ids []string) []string {
	t.Helper()
	shard := sim.state.Shards[root.Key()]
	shard.Queue = make([]CrossShardMessage, 0, len(ids))
	out := append([]string(nil), ids...)
	for _, id := range ids {
		shard.Queue = append(shard.Queue, CrossShardMessage{
			Source:		root,
			Destination:	root,
			MessageID:	id,
			RoutingKey:	[]byte(id),
			Timeout:	10,
		})
	}
	shard.MessageQueueRoot = hashQueue(shard.Queue)
	sim.state.Shards[root.Key()] = shard
	sim.commitHeader(shard)
	sort.Strings(out)
	return out
}

func queueIDs(queue []CrossShardMessage) []string {
	ids := make([]string, len(queue))
	for i, msg := range queue {
		ids[i] = msg.MessageID
	}
	return ids
}
