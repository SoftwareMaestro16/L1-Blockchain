package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouteKeyToShardConsistentHashIsDeterministicAcrossNodes(t *testing.T) {
	layoutA := routingTestLayout(t, ZoneIDFinancial, 7, ShardAssignmentConsistentHash)
	layoutB := routingTestLayout(t, ZoneIDFinancial, 7, ShardAssignmentConsistentHash)
	layoutB.ActiveShards = []ShardDescriptor{layoutB.ActiveShards[2], layoutB.ActiveShards[0], layoutB.ActiveShards[1]}
	layoutB.LayoutHash = ComputeShardLayoutHash(layoutB)

	input := ShardRoutingInput{
		ZoneID:			ZoneIDFinancial,
		StateKey:		"zone/financial/balances/alice/naet",
		ShardLayoutEpoch:	7,
	}
	routeA, err := RouteKeyToShard(layoutA, input)
	require.NoError(t, err)
	routeB, err := RouteKeyToShard(layoutB, input)
	require.NoError(t, err)
	require.Equal(t, routeA.ShardID, routeB.ShardID)
	require.Equal(t, routeA.RouteHash, routeB.RouteHash)
	require.Equal(t, ShardAssignmentConsistentHash, routeA.AssignmentMode)
	require.NoError(t, routeA.ValidateHash())
}

func TestRouteKeyToShardUsesCommittedPrefixAndExplicitPlacement(t *testing.T) {
	layout := routingTestLayout(t, ZoneIDApplication, 3, ShardAssignmentKeyPrefix)

	appRoute, err := RouteKeyToShard(layout, ShardRoutingInput{
		ZoneID:			ZoneIDApplication,
		StateKey:		"apps/workflow/workflow-9",
		ShardLayoutEpoch:	3,
	})
	require.NoError(t, err)
	require.Equal(t, ShardID("1"), appRoute.ShardID)
	require.Equal(t, ShardAssignmentKeyPrefix, appRoute.AssignmentMode)

	systemRoute, err := RouteKeyToShard(layout, ShardRoutingInput{
		ZoneID:			ZoneIDApplication,
		StateKey:		"core/params",
		ShardLayoutEpoch:	3,
	})
	require.NoError(t, err)
	require.Equal(t, ShardID("0"), systemRoute.ShardID)
	require.Equal(t, ShardAssignmentExplicit, systemRoute.AssignmentMode)
	require.Equal(t, ShardID("0"), systemRoute.PlacementOverride)

	readOnlyRoute, err := RouteKeyToShard(layout, ShardRoutingInput{
		ZoneID:			ZoneIDApplication,
		StateKey:		"zone/cache/routing/table",
		ShardLayoutEpoch:	3,
	})
	require.NoError(t, err)
	require.Equal(t, ShardID("0"), readOnlyRoute.ShardID)
	require.True(t, readOnlyRoute.ReadOnlyReplicated)
}

func TestRouteKeyToShardRejectsWrongEpochAndMetricsCommit(t *testing.T) {
	layout := routingTestLayout(t, ZoneIDContract, 4, ShardAssignmentConsistentHash)
	_, err := RouteKeyToShard(layout, ShardRoutingInput{
		ZoneID:			ZoneIDContract,
		StateKey:		"contract/instance/contract-1",
		ShardLayoutEpoch:	5,
	})
	require.ErrorContains(t, err, "epoch")

	metrics, err := NewShardMetrics(ShardMetrics{
		ZoneID:			ZoneIDContract,
		ShardID:		"2",
		Height:			99,
		GasUsed:		1_000,
		FeeCollected:		55,
		InboxBacklog:		3,
		OutboxBacklog:		4,
		WriteConflictCount:	2,
		StateSizeBytes:		4096,
		ProofLatencyMicros:	800,
		ExecutionDelayMicros:	900,
		FailedDeliveryCount:	1,
		ExpiredMessageCount:	1,
	})
	require.NoError(t, err)
	require.Equal(t, ComputeShardMetricsHash(metrics), metrics.MetricsHash)
	require.NoError(t, metrics.ValidateHash())
}

func routingTestLayout(t *testing.T, zoneID ZoneID, epoch uint64, mode ShardAssignmentMode) ShardLayout {
	t.Helper()
	shards := []ShardDescriptor{
		{
			ShardID:		"0",
			StatePrefix:		"zone/" + string(zoneID) + "/shard/0",
			KeyPrefix:		"apps/app",
			ActivationHeight:	1,
			ValidatorSetHash:	testHash(string(zoneID) + "/0/validators"),
			Available:		true,
			SystemShard:		true,
		},
		{
			ShardID:		"1",
			StatePrefix:		"zone/" + string(zoneID) + "/shard/1",
			KeyPrefix:		"apps/workflow",
			ActivationHeight:	1,
			ValidatorSetHash:	testHash(string(zoneID) + "/1/validators"),
			Available:		true,
		},
		{
			ShardID:		"2",
			StatePrefix:		"zone/" + string(zoneID) + "/shard/2",
			KeyPrefix:		"contract/instance",
			ActivationHeight:	1,
			ValidatorSetHash:	testHash(string(zoneID) + "/2/validators"),
			Available:		true,
		},
	}
	layout, err := NewShardLayout(zoneID, epoch, 1, testHash(string(zoneID)+"/routing-seed"), shards)
	require.NoError(t, err)
	layout.AssignmentMode = mode
	layout.SystemShardID = "0"
	layout.PlacementOverrides = []ShardPlacementOverride{
		{ObjectKey: "core/params", ShardID: "0"},
	}
	layout.ReadOnlyReplicatedKeys = []string{"zone/cache"}
	layout.LayoutHash = ComputeShardLayoutHash(layout)
	require.NoError(t, layout.ValidateHash())
	return layout
}
