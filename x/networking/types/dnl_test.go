package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDNLStateDeterministicRootsAndZoneAwareLookup(t *testing.T) {
	routeA := testDNLRoute(t, "financial", "svc-pay", "node-a", 1, 8_000, 90)
	routeB := testDNLRoute(t, "identity", "svc-name", "node-b", 2, 7_000, 80)
	serviceA := testDNLEntry(t, "svc-pay", "financial", routeA.RouteID, 90)
	serviceB := testDNLEntry(t, "svc-name", "identity", routeB.RouteID, 80)

	left, err := BuildDNLState(
		[]DNLServiceDiscoveryEntry{serviceB, serviceA},
		[]DNLRoutingTableEntry{routeB, routeA},
		nil,
		20,
	)
	require.NoError(t, err)
	right, err := BuildDNLState(
		[]DNLServiceDiscoveryEntry{serviceA, serviceB},
		[]DNLRoutingTableEntry{routeA, routeB},
		nil,
		20,
	)
	require.NoError(t, err)
	require.Equal(t, left.StateRoot, right.StateRoot)
	require.Equal(t, left.LookupRoot, right.LookupRoot)

	response, err := QueryDNL(left, DNLQuery{
		ServiceID:	"svc-pay",
		ZoneID:		"financial",
		CurrentHeight:	30,
		Limit:		10,
		RequireProof:	true,
	})
	require.NoError(t, err)
	require.Len(t, response.Entries, 1)
	require.Len(t, response.Routes, 1)
	require.Equal(t, serviceA.EntryID, response.Entries[0].EntryID)
	require.Equal(t, routeA.RouteID, response.Routes[0].RouteID)
	require.NoError(t, response.Proof.Validate())
	require.Equal(t, left.StateRoot, response.Proof.StateRoot)
}

func TestDNLProofAttachedCacheEntryCommitsDiscoveryResponse(t *testing.T) {
	route := testDNLRoute(t, "contract", "svc-contract", "node-c", 1, 9_000, 100)
	service := testDNLEntry(t, "svc-contract", "contract", route.RouteID, 100)
	state, err := BuildDNLState([]DNLServiceDiscoveryEntry{service}, []DNLRoutingTableEntry{route}, nil, 20)
	require.NoError(t, err)

	response, err := QueryDNL(state, DNLQuery{ServiceID: "svc-contract", CurrentHeight: 21, RequireProof: true})
	require.NoError(t, err)
	cache, err := BuildDNLCacheEntryFromResponse(response)
	require.NoError(t, err)
	require.Equal(t, response.ResponseHash, cache.ResponseHash)
	require.Equal(t, response.Proof.ProofHash, cache.ProofHash)

	cachedState, err := BuildDNLState([]DNLServiceDiscoveryEntry{service}, []DNLRoutingTableEntry{route}, []DNLCacheEntry{cache}, 22)
	require.NoError(t, err)
	key, err := DNLCacheKey(cache.CacheKey)
	require.NoError(t, err)
	proof, err := QueryDNLProof(cachedState, key)
	require.NoError(t, err)
	require.Equal(t, cache.EntryHash, proof.ValueHash)
}

func TestDNLRejectsExpiredCacheAndUncommittedRoutes(t *testing.T) {
	route := testDNLRoute(t, "apps", "svc-app", "node-d", 1, 5_000, 10)
	service := testDNLEntry(t, "svc-app", "apps", route.RouteID, 10)
	response := DNLDiscoveryResponse{
		QueryHash:	HashParts("query"),
		Entries:	[]DNLServiceDiscoveryEntry{service},
		Routes:		[]DNLRoutingTableEntry{route},
		Proof:		DNLProof{ProofHash: HashParts("proof")},
		ExpiryHeight:	10,
	}
	response.ResponseHash = ComputeDNLDiscoveryResponseHash(response)
	cache, err := NewDNLCacheEntry(DNLCacheEntry{
		QueryHash:	response.QueryHash,
		ResponseHash:	response.ResponseHash,
		ExpiryHeight:	10,
		ProofHash:	response.Proof.ProofHash,
	})
	require.NoError(t, err)
	_, err = BuildDNLState([]DNLServiceDiscoveryEntry{service}, []DNLRoutingTableEntry{route}, []DNLCacheEntry{cache}, 11)
	require.ErrorContains(t, err, "expired")

	orphan := testDNLRoute(t, "apps", "svc-other", "node-e", 1, 5_000, 30)
	_, err = BuildDNLState([]DNLServiceDiscoveryEntry{service}, []DNLRoutingTableEntry{orphan}, nil, 20)
	require.ErrorContains(t, err, "committed service")
}

func TestDNLNodeLocalObservationIsAdvisoryOnly(t *testing.T) {
	observation, err := NewDNLAdvisoryObservation(DNLAdvisoryObservation{
		ObservedNodeID:	HashParts("node-local"),
		ServiceID:	"svc-pay",
		ZoneID:		"financial",
		EndpointHash:	HashParts("observed-endpoint"),
		ObservedHeight:	25,
	})
	require.NoError(t, err)
	require.NoError(t, observation.Validate())
	require.ErrorContains(t, RejectDNLAdvisoryObservationForConsensus(observation), "advisory")
}

func TestDNLQuerySkipsExpiredEntriesAndRequiresProofWhenRequested(t *testing.T) {
	route := testDNLRoute(t, "storage", "svc-store", "node-f", 1, 8_000, 30)
	service := testDNLEntry(t, "svc-store", "storage", route.RouteID, 30)
	state, err := BuildDNLState([]DNLServiceDiscoveryEntry{service}, []DNLRoutingTableEntry{route}, nil, 20)
	require.NoError(t, err)

	response, err := QueryDNL(state, DNLQuery{ServiceID: "svc-store", CurrentHeight: 31, RequireProof: false})
	require.NoError(t, err)
	require.Empty(t, response.Entries)
	require.Empty(t, response.Routes)

	_, err = QueryDNL(state, DNLQuery{ServiceID: "missing", CurrentHeight: 21, RequireProof: true})
	require.ErrorContains(t, err, "not found")
}

func testDNLEntry(t *testing.T, serviceID, zoneID, routeID string, expiry uint64) DNLServiceDiscoveryEntry {
	t.Helper()
	entry, err := NewDNLServiceDiscoveryEntry(DNLServiceDiscoveryEntry{
		ServiceID:	serviceID,
		ZoneID:		zoneID,
		InterfaceHash:	HashParts(serviceID, "interface"),
		EndpointHash:	HashParts(serviceID, "endpoint"),
		RouteID:	routeID,
		ExpiryHeight:	expiry,
		ProofHash:	HashParts(serviceID, "proof"),
	})
	require.NoError(t, err)
	return entry
}

func testDNLRoute(t *testing.T, zoneID, serviceID, nodeSeed string, priority, weight uint32, expiry uint64) DNLRoutingTableEntry {
	t.Helper()
	route, err := NewDNLRoutingTableEntry(DNLRoutingTableEntry{
		ZoneID:		zoneID,
		ServiceID:	serviceID,
		NextHopNodeID:	HashParts(nodeSeed, "node"),
		OverlayID:	HashParts(zoneID, "overlay"),
		Priority:	priority,
		WeightBps:	weight,
		ExpiryHeight:	expiry,
	})
	require.NoError(t, err)
	return route
}
