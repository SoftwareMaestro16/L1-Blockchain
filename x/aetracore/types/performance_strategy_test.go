package types

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockSTMZonePerformancePlanGroupsDisjointZonesAndShards(t *testing.T) {
	items := []ProposalItem{
		testProposalItem(ZoneIDContract, "1", "contract-call", 2, 12, 1),
		testProposalItem(ZoneIDFinancial, "0", "financial-transfer", 1, 12, 0),
	}
	schedule, err := BuildProposalSchedule(25, items, TestnetParams())
	require.NoError(t, err)

	accesses := []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDContract, "1", ZoneIDContract, "1", "contract/storage/abc", BlockSTMAccessWrite, false),
		blockSTMAccess(ZoneIDFinancial, "0", ZoneIDFinancial, "0", "financial/balances/alice", BlockSTMAccessWrite, false),
		blockSTMAccess(ZoneIDFinancial, "0", ZoneIDFinancial, "0", "financial/balances/bob", BlockSTMAccessRead, false),
	}
	batches := []BlockSTMMessageBatch{
		blockSTMBatch(ZoneIDFinancial, "0", ZoneIDContract, "1", 2),
	}

	plan, err := BuildBlockSTMZonePerformancePlan(schedule, accesses, batches)
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Equal(t, uint64(25), plan.Height)
	require.Equal(t, uint32(2), plan.ParallelWorkloads)
	require.Zero(t, plan.GlobalWriteLocks)
	require.Empty(t, plan.ConflictSets)
	require.NotEmpty(t, plan.PlanHash)

	byScope := map[string]BlockSTMZoneWorkload{}
	for _, workload := range plan.Workloads {
		byScope[string(workload.ZoneID)+"/"+string(workload.ShardID)] = workload
	}
	require.Len(t, byScope["FINANCIAL_ZONE/0"].Items, 1)
	require.Len(t, byScope["CONTRACT_ZONE/1"].Items, 1)
	require.Len(t, byScope["FINANCIAL_ZONE/0"].MessageBatches, 1)
}

func TestBlockSTMZonePerformancePlanRootIsCanonicalAcrossInputOrder(t *testing.T) {
	items := []ProposalItem{
		testProposalItem(ZoneIDFinancial, "1", "right", 2, 13, 1),
		testProposalItem(ZoneIDFinancial, "0", "left", 1, 13, 0),
	}
	scheduleA, err := BuildProposalSchedule(26, items, TestnetParams())
	require.NoError(t, err)
	slices.Reverse(items)
	scheduleB, err := BuildProposalSchedule(26, items, TestnetParams())
	require.NoError(t, err)

	accesses := []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDFinancial, "1", ZoneIDFinancial, "1", "financial/dex/pools/2", BlockSTMAccessWrite, false),
		blockSTMAccess(ZoneIDFinancial, "0", ZoneIDFinancial, "0", "financial/balances/alice", BlockSTMAccessWrite, false),
	}
	batches := []BlockSTMMessageBatch{
		blockSTMBatch(ZoneIDFinancial, "1", ZoneIDIdentity, "0", 1),
		blockSTMBatch(ZoneIDFinancial, "0", ZoneIDContract, "2", 3),
	}

	planA, err := BuildBlockSTMZonePerformancePlan(scheduleA, accesses, batches)
	require.NoError(t, err)
	slices.Reverse(accesses)
	slices.Reverse(batches)
	planB, err := BuildBlockSTMZonePerformancePlan(scheduleB, accesses, batches)
	require.NoError(t, err)

	require.Equal(t, planA.PlanHash, planB.PlanHash)
	require.Equal(t, planA.Workloads, planB.Workloads)
}

func TestBlockSTMZonePerformancePlanRejectsDirectCrossZoneWrites(t *testing.T) {
	schedule, err := BuildProposalSchedule(27, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "0", "transfer", 1, 27, 0),
	}, TestnetParams())
	require.NoError(t, err)

	_, err = BuildBlockSTMZonePerformancePlan(schedule, []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDFinancial, "0", ZoneIDContract, "0", "contract/storage/remote", BlockSTMAccessWrite, false),
	}, nil)
	require.ErrorContains(t, err, "cross-zone writes")

	plan, err := BuildBlockSTMZonePerformancePlan(schedule, []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDFinancial, "0", ZoneIDContract, "0", "contract/storage/remote", BlockSTMAccessWrite, true),
	}, []BlockSTMMessageBatch{
		blockSTMBatch(ZoneIDFinancial, "0", ZoneIDContract, "0", 1),
	})
	require.NoError(t, err)
	require.Equal(t, uint32(1), plan.CrossZoneWrites)
}

func TestBlockSTMZonePerformancePlanRejectsGlobalWriteLocks(t *testing.T) {
	schedule, err := BuildProposalSchedule(28, []ProposalItem{
		testProposalItem(ZoneIDAetraCore, "0", "global", 1, 28, 0),
	}, TestnetParams())
	require.NoError(t, err)

	_, err = BuildBlockSTMZonePerformancePlan(schedule, []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDAetraCore, "0", ZoneIDAetraCore, "0", "core/global-lock", BlockSTMAccessWrite, false),
	}, nil)
	require.ErrorContains(t, err, "global write lock")
}

func TestBlockSTMZonePerformancePlanDetectsSameObjectConflicts(t *testing.T) {
	schedule, err := BuildProposalSchedule(29, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "0", "left", 1, 29, 0),
		testProposalItem(ZoneIDFinancial, "1", "right", 1, 29, 1),
	}, TestnetParams())
	require.NoError(t, err)

	plan, err := BuildBlockSTMZonePerformancePlan(schedule, []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDFinancial, "0", ZoneIDFinancial, "0", "financial/balances/shared", BlockSTMAccessWrite, false),
		blockSTMAccess(ZoneIDFinancial, "1", ZoneIDFinancial, "0", "financial/balances/shared", BlockSTMAccessWrite, false),
	}, nil)
	require.NoError(t, err)
	require.Len(t, plan.ConflictSets, 1)
	require.Equal(t, "FINANCIAL_ZONE/0/financial/balances/shared", plan.ConflictSets[0].ConflictKey)
	require.Len(t, plan.ConflictSets[0].WorkloadIDs, 2)
}

func blockSTMAccess(actorZone ZoneID, actorShard ShardID, stateZone ZoneID, stateShard ShardID, key string, mode BlockSTMAccessMode, viaMessage bool) BlockSTMStateAccess {
	return BlockSTMStateAccess{
		ActorZoneID:	actorZone,
		ActorShardID:	actorShard,
		StateZoneID:	stateZone,
		StateShardID:	stateShard,
		StateKey:	key,
		Mode:		mode,
		ViaMessage:	viaMessage,
	}
}

func blockSTMBatch(sourceZone ZoneID, sourceShard ShardID, destinationZone ZoneID, destinationShard ShardID, count uint32) BlockSTMMessageBatch {
	batch := BlockSTMMessageBatch{
		SourceZoneID:		sourceZone,
		SourceShardID:		sourceShard,
		DestinationZoneID:	destinationZone,
		DestinationShardID:	destinationShard,
		MessageCount:		count,
	}
	batch.BatchHash = ComputeBlockSTMMessageBatchHash(batch)
	return batch
}
