package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockSTMStrategyGroupsDisjointZonesAndShardQueues(t *testing.T) {
	items := []BlockSTMExecutionItem{
		blockSTMItem("tx-b", 1, "financial", "shard-2", "acct-b", 3, "7"),
		blockSTMItem("tx-a", 0, "identity", "shard-1", "name-a", 9, "5"),
	}
	items[1].RemoteWrites = []BlockSTMRemoteWrite{{
		DestinationZoneID:	"contract",
		DestinationShardID:	"shard-9",
		ObjectID:		"contract-a",
		PayloadHash:		hashStrings("payload", "contract-a"),
	}}

	plan, err := BuildBlockSTMStrategyPlan(77, items)
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.NotEmpty(t, plan.PlanHash)
	require.Len(t, plan.Groups, 2)
	require.Equal(t, uint32(1), plan.Groups[0].ParallelBatch)
	require.Equal(t, uint32(1), plan.Groups[1].ParallelBatch)

	require.Len(t, plan.FeeAccumulators, 2)
	require.Equal(t, "financial", plan.FeeAccumulators[0].ZoneID)
	require.Equal(t, "7", plan.FeeAccumulators[0].Amount)
	require.Equal(t, "identity", plan.FeeAccumulators[1].ZoneID)
	require.Equal(t, "5", plan.FeeAccumulators[1].Amount)

	require.Len(t, plan.MessageQueues, 1)
	require.Equal(t, "contract", plan.MessageQueues[0].ZoneID)
	require.Equal(t, "shard-9", plan.MessageQueues[0].ShardID)
	require.Len(t, plan.MessageQueues[0].Messages, 1)
	require.Equal(t, uint64(1), plan.MessageQueues[0].Messages[0].Sequence)
	require.Equal(t, "identity", plan.MessageQueues[0].Messages[0].SourceZoneID)

	require.Len(t, plan.ObjectUpdates, 2)
	require.Equal(t, uint64(10), plan.ObjectUpdates[1].NextVersion)
}

func TestBlockSTMStrategySerializesSameObjectConflictsDeterministically(t *testing.T) {
	first := blockSTMItem("tx-a", 0, "financial", "shard-1", "acct-a", 4, "1")
	second := blockSTMItem("tx-b", 1, "financial", "shard-1", "acct-a", 4, "1")
	second.WriteKeys = []string{"zone/financial/shard/shard-1/object/acct-a/version/00000000000000000004/balance-secondary"}

	plan, err := BuildBlockSTMStrategyPlan(80, []BlockSTMExecutionItem{second, first})
	require.NoError(t, err)
	require.Len(t, plan.Groups, 2)
	require.Equal(t, uint32(1), plan.Groups[0].ParallelBatch)
	require.Equal(t, "tx-a", plan.Groups[0].Items[0].TxID)
	require.Equal(t, uint32(2), plan.Groups[1].ParallelBatch)
	require.Equal(t, "tx-b", plan.Groups[1].Items[0].TxID)
	require.Len(t, plan.ObjectUpdates, 1)
}

func TestBlockSTMStrategyRejectsGlobalHotPathWrites(t *testing.T) {
	item := blockSTMItem("tx-hot", 0, "financial", "shard-1", "acct-hot", 1, "1")
	item.WriteKeys = []string{"zone/financial/shard/shard-1/counter/global-fees"}

	_, err := BuildBlockSTMStrategyPlan(81, []BlockSTMExecutionItem{item})
	require.ErrorContains(t, err, "global counters")
}

func TestBlockSTMStrategyRejectsSynchronousRemoteLocalWrites(t *testing.T) {
	item := blockSTMItem("tx-local-remote", 0, "financial", "shard-1", "acct-local", 1, "1")
	item.RemoteWrites = []BlockSTMRemoteWrite{{
		DestinationZoneID:	"financial",
		DestinationShardID:	"shard-1",
		ObjectID:		"acct-local",
		PayloadHash:		hashStrings("payload", "local"),
	}}

	_, err := BuildBlockSTMStrategyPlan(82, []BlockSTMExecutionItem{item})
	require.ErrorContains(t, err, "local writes")
}

func blockSTMItem(txID string, txIndex uint32, zoneID string, shardID string, objectID string, version uint64, fee string) BlockSTMExecutionItem {
	stateKey := VersionedObjectStateKey(zoneID, shardID, objectID, version)
	return BlockSTMExecutionItem{
		TxID:		txID,
		TxIndex:	txIndex,
		ZoneID:		zoneID,
		ShardID:	shardID,
		ObjectID:	objectID,
		ObjectVersion:	version,
		FeeAmount:	fee,
		ReadKeys:	[]string{stateKey + "/balance"},
		WriteKeys:	[]string{stateKey + "/balance"},
	}
}
