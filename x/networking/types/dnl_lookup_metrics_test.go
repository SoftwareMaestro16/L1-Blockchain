package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDNLRecursiveLookupUsesProofsAndIgnoresExpiredNodes(t *testing.T) {
	salt := []byte("aetra-test-network")
	node := testDNLNodeRecord(t, 0x71, salt, []string{"financial"}, []string{"svc-pay"}, 100)
	expired := testDNLNodeRecord(t, 0x72, salt, []string{"financial"}, []string{"svc-pay"}, 20)
	reputation := testDNLReputation(t, node.NodeID, 30)
	expiredReputation := testDNLReputation(t, expired.NodeID, 30)

	validRoute := testDNLRouteForNode(t, "financial", "svc-pay", node.NodeID, 1, 8_000, 100)
	expiredRoute := testDNLRouteForNode(t, "financial", "svc-pay", expired.NodeID, 2, 7_000, 100)
	validEntry := testDNLEntry(t, "svc-pay", "financial", validRoute.RouteID, 100)
	expiredEntry := testDNLEntry(t, "svc-pay", "financial", expiredRoute.RouteID, 100)
	dnl, err := BuildDNLState(
		[]DNLServiceDiscoveryEntry{expiredEntry, validEntry},
		[]DNLRoutingTableEntry{expiredRoute, validRoute},
		nil,
		30,
	)
	require.NoError(t, err)
	routing, err := BuildDNLRoutingState(
		[]NodeRecord{expired, node},
		[]ReputationCommitment{expiredReputation, reputation},
		nil,
		[]RoutingTable{testRoutingTableForRoutes(t, 3, validRoute, expiredRoute)},
		30,
	)
	require.NoError(t, err)

	response, err := RecursiveDNLLookup(dnl, routing, DNLRecursiveLookupRequest{
		ServiceID:		"svc-pay",
		CurrentHeight:		30,
		MaxDepth:		2,
		Limit:			10,
		ConsensusRoutingOnly:	true,
	})
	require.NoError(t, err)
	require.Len(t, response.Hops, 1)
	require.Equal(t, node.NodeID, response.Hops[0].NodeID)
	require.NotEmpty(t, response.Proofs)
	require.NoError(t, response.Validate())
}

func TestDNLConsensusRouteSelectionUsesCommittedMetricsOnly(t *testing.T) {
	salt := []byte("aetra-test-network")
	fast := testDNLNodeRecord(t, 0x81, salt, []string{"contract"}, []string{"svc-contract"}, 100)
	slow := testDNLNodeRecord(t, 0x82, salt, []string{"contract"}, []string{"svc-contract"}, 100)
	fastRoute := testDNLRouteForNode(t, "contract", "svc-contract", fast.NodeID, 1, 8_000, 100)
	slowRoute := testDNLRouteForNode(t, "contract", "svc-contract", slow.NodeID, 1, 8_000, 100)
	table := testRoutingTableForRoutes(t, 7, slowRoute, fastRoute)
	state, err := BuildDNLRoutingState(
		[]NodeRecord{slow, fast},
		[]ReputationCommitment{testDNLReputation(t, slow.NodeID, 40), testDNLReputation(t, fast.NodeID, 40)},
		nil,
		[]RoutingTable{table},
		40,
	)
	require.NoError(t, err)
	fastMetric := testDNLMetric(t, fastRoute, 10, 10, 9_000, 500, true, 40)
	slowMetric := testDNLMetric(t, slowRoute, 200, 10, 9_000, 500, true, 40)

	selection, err := SelectConsensusRouteFromCommittedMetrics(state, 7, []DNLRoutingMetricSnapshot{slowMetric, fastMetric}, DNLRecursiveLookupRequest{
		ServiceID:		"svc-contract",
		ZoneID:			"contract",
		CurrentHeight:		41,
		ConsensusRoutingOnly:	true,
	})
	require.NoError(t, err)
	require.Equal(t, fastRoute.RouteID, selection.Route.RouteID)
	require.True(t, selection.UsedCommittedTable)
	require.False(t, selection.UsedLiveHint)
	require.NoError(t, selection.Validate(true))

	uncommitted := fastMetric
	uncommitted.Committed = false
	uncommitted.SnapshotID = ""
	uncommitted.SnapshotHash = ""
	uncommitted, err = NewDNLRoutingMetricSnapshot(uncommitted)
	require.NoError(t, err)
	_, err = SelectConsensusRouteFromCommittedMetrics(state, 7, []DNLRoutingMetricSnapshot{uncommitted, slowMetric}, DNLRecursiveLookupRequest{
		ServiceID:		"svc-contract",
		ZoneID:			"contract",
		CurrentHeight:		41,
		ConsensusRoutingOnly:	true,
	})
	require.ErrorContains(t, err, "committed")
}

func TestDNLLiveLatencyHintsAreAdvisoryUntilCommitted(t *testing.T) {
	hint, err := NewDNLLiveRoutingHint(DNLLiveRoutingHint{
		RouteID:	HashParts("route"),
		NodeID:		HashParts("node"),
		LatencyMillis:	5,
		ObservedHeight:	42,
	})
	require.NoError(t, err)
	require.NoError(t, hint.Validate())
	require.ErrorContains(t, RejectLiveRoutingHintForConsensus(hint), "advisory")

	route := testDNLRouteForNode(t, "storage", "svc-store", hint.NodeID, 1, 8_000, 100)
	metric := testDNLMetric(t, route, hint.LatencyMillis, 20, 8_500, 100, true, 42)
	require.NoError(t, ValidateDNLMetricConsensusUse(metric, true))
	require.NotEmpty(t, ComputeDNLMetricRoot([]DNLRoutingMetricSnapshot{metric}))
}

func testDNLRouteForNode(t *testing.T, zoneID, serviceID, nodeID string, priority, weight uint32, expiry uint64) DNLRoutingTableEntry {
	t.Helper()
	route, err := NewDNLRoutingTableEntry(DNLRoutingTableEntry{
		ZoneID:		zoneID,
		ServiceID:	serviceID,
		NextHopNodeID:	nodeID,
		OverlayID:	HashParts("lookup-metrics-overlay", zoneID, serviceID),
		Priority:	priority,
		WeightBps:	weight,
		ExpiryHeight:	expiry,
	})
	require.NoError(t, err)
	return route
}

func testRoutingTableForRoutes(t *testing.T, epoch uint64, routes ...DNLRoutingTableEntry) RoutingTable {
	t.Helper()
	table, err := NewRoutingTable(RoutingTable{Epoch: epoch, Routes: routes})
	require.NoError(t, err)
	return table
}

func testDNLMetric(t *testing.T, route DNLRoutingTableEntry, latency, gas uint64, reliability, congestion uint32, committed bool, height uint64) DNLRoutingMetricSnapshot {
	t.Helper()
	metric, err := NewDNLRoutingMetricSnapshot(DNLRoutingMetricSnapshot{
		RouteID:		route.RouteID,
		NodeID:			route.NextHopNodeID,
		ZoneID:			route.ZoneID,
		ServiceID:		route.ServiceID,
		LatencyMillis:		latency,
		GasCost:		gas,
		ReliabilityScoreBps:	reliability,
		CongestionWeightBps:	congestion,
		ZoneSupport:		true,
		ServiceSupport:		true,
		Committed:		committed,
		SnapshotHeight:		height,
	})
	require.NoError(t, err)
	return metric
}
