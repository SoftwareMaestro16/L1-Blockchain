package types

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeRecordCanonicalFieldsAndSignature(t *testing.T) {
	salt := []byte("aetra-test-network")
	node := testDNLNodeRecord(t, 0x41, salt, []string{"financial"}, []string{"svc-pay"}, 100)

	require.Equal(t, DefaultNodeRecordVersion, node.RecordVersion)
	require.Equal(t, node.NodePubKey, node.PublicKey)
	require.Equal(t, []string{"svc-pay"}, node.Services)
	require.Equal(t, node.Services, node.ServiceIDs)
	require.Equal(t, node.ProtocolVersions, node.SupportedProtocols)
	require.NoError(t, node.Validate(salt, 20))
	require.NotEmpty(t, ComputeNodeRecordCommitmentHash(node))

	tampered := node
	tampered.ServicesSupported = append(tampered.ServicesSupported, "svc-extra")
	require.ErrorContains(t, tampered.Validate(salt, 20), "signature")
}

func TestDNLRoutingStateKeysMatchSpec(t *testing.T) {
	nodeID := HashParts("node")
	nodeKey, err := RoutingNodeKey(nodeID)
	require.NoError(t, err)
	require.Equal(t, "routing/nodes/"+nodeID, nodeKey)

	zoneKey, err := RoutingZoneKey("financial", nodeID)
	require.NoError(t, err)
	require.Equal(t, "routing/zones/financial/"+nodeID, zoneKey)

	serviceKey, err := RoutingServiceKey("svc-pay", nodeID)
	require.NoError(t, err)
	require.Equal(t, "routing/services/svc-pay/"+nodeID, serviceKey)

	reputationKey, err := RoutingReputationKey(nodeID)
	require.NoError(t, err)
	require.Equal(t, "routing/reputation/"+nodeID, reputationKey)

	cacheKey, err := RoutingCacheKey(HashParts("lookup"))
	require.NoError(t, err)
	require.Equal(t, "routing/cache/"+HashParts("lookup"), cacheKey)

	tableKey, err := RoutingTableKey(7)
	require.NoError(t, err)
	require.Equal(t, "routing/table/00000000000000000007", tableKey)
}

func TestDNLRoutingStateBuildsDeterministicIndexesAndRoots(t *testing.T) {
	salt := []byte("aetra-test-network")
	first := testDNLNodeRecord(t, 0x51, salt, []string{"financial", "identity"}, []string{"svc-pay"}, 100)
	second := testDNLNodeRecord(t, 0x52, salt, []string{"contract"}, []string{"svc-contract"}, 100)
	firstReputation := testDNLReputation(t, first.NodeID, 20)
	secondReputation := testDNLReputation(t, second.NodeID, 20)
	cache := testLookupCacheRecord(t, 90)
	table := testRoutingTable(t, 3, "financial", "svc-pay", first.NodeID)

	left, err := BuildDNLRoutingState(
		[]NodeRecord{second, first},
		[]ReputationCommitment{secondReputation, firstReputation},
		[]LookupCacheRecord{cache},
		[]RoutingTable{table},
		30,
	)
	require.NoError(t, err)
	right, err := BuildDNLRoutingState(
		[]NodeRecord{first, second},
		[]ReputationCommitment{firstReputation, secondReputation},
		[]LookupCacheRecord{cache},
		[]RoutingTable{table},
		30,
	)
	require.NoError(t, err)
	require.Equal(t, left.StateRoot, right.StateRoot)
	require.Len(t, left.ZoneIndex, 3)
	require.Len(t, left.ServiceIndex, 2)

	got, found := QueryRoutingNode(left, first.NodeID)
	require.True(t, found)
	require.Equal(t, first.NodeID, got.NodeID)
	require.Len(t, QueryRoutingNodesByZone(left, "financial"), 1)
	require.Len(t, QueryRoutingNodesByService(left, "svc-pay"), 1)
	gotTable, found := LookupRoutingTable(left, 3)
	require.True(t, found)
	require.Equal(t, table.TableRoot, gotTable.TableRoot)
}

func TestDNLRoutingStateRejectsMissingReputationAndExpiredCache(t *testing.T) {
	salt := []byte("aetra-test-network")
	node := testDNLNodeRecord(t, 0x61, salt, []string{"apps"}, []string{"svc-app"}, 100)
	_, err := BuildDNLRoutingState([]NodeRecord{node}, nil, nil, nil, 20)
	require.ErrorContains(t, err, "reputation")

	reputation := testDNLReputation(t, node.NodeID, 20)
	cache := testLookupCacheRecord(t, 19)
	_, err = BuildDNLRoutingState([]NodeRecord{node}, []ReputationCommitment{reputation}, []LookupCacheRecord{cache}, nil, 20)
	require.ErrorContains(t, err, "expired")
}

func testDNLNodeRecord(t *testing.T, seed byte, salt []byte, zones, services []string, expires uint64) NodeRecord {
	t.Helper()
	privateKey := deterministicPrivateKey(seed)
	addressHash, err := HashNetworkAddresses([]string{"tcp://10.0.0." + string(rune(seed)) + ":26656"})
	require.NoError(t, err)
	peerKey := ed25519.NewKeyFromSeed([]byte{
		seed, seed, seed, seed, seed, seed, seed, seed,
		seed, seed, seed, seed, seed, seed, seed, seed,
		seed, seed, seed, seed, seed, seed, seed, seed,
		seed, seed, seed, seed, seed, seed, seed, seed,
	}).Public().(ed25519.PublicKey)
	peerID := ComputeNodeID(peerKey, salt)
	nodeID := ComputeNodeID(privateKey.Public().(ed25519.PublicKey), salt)
	latency, err := NewNodeLatencyVectorEntry(NodeLatencyVectorEntry{
		NodeID:		nodeID,
		PeerNodeID:	peerID,
		ZoneID:		zones[0],
		LatencyMillis:	25,
		SampleHeight:	12,
	})
	require.NoError(t, err)
	record, err := SignNodeRecord(NodeRecord{
		OperatorAddress:	"operator-" + HashParts("operator", string([]byte{seed}))[:8],
		Roles:			[]NodeRole{NodeRoleService, NodeRoleRouting},
		NetworkAddressesHash:	addressHash,
		ZonesSupported:		zones,
		Services:		services,
		SupportedProtocols:	[]string{DefaultProtocolVersion},
		LatencyVector:		[]NodeLatencyVectorEntry{latency},
		ExpiresHeight:		expires,
	}, privateKey, salt)
	require.NoError(t, err)
	return record
}

func testDNLReputation(t *testing.T, nodeID string, height uint64) ReputationCommitment {
	t.Helper()
	reputation, err := NewReputationCommitment(ReputationCommitment{
		NodeID:	nodeID,
		Reputation: PeerScore{
			ScoreBps:	9_000,
			LatencyBps:	8_000,
			ReliabilityBps:	9_500,
			ThroughputBps:	7_500,
			PenaltyBps:	100,
		},
		EvidenceRoot:	HashParts("evidence", nodeID),
		UpdatedHeight:	height,
	})
	require.NoError(t, err)
	return reputation
}

func testLookupCacheRecord(t *testing.T, expiry uint64) LookupCacheRecord {
	t.Helper()
	record, err := NewLookupCacheRecord(LookupCacheRecord{
		QueryHash:	HashParts("query"),
		ResponseHash:	HashParts("response"),
		ExpiryHeight:	expiry,
		ProofHash:	HashParts("proof"),
	})
	require.NoError(t, err)
	return record
}

func testRoutingTable(t *testing.T, epoch uint64, zoneID, serviceID, nodeID string) RoutingTable {
	t.Helper()
	route, err := NewDNLRoutingTableEntry(DNLRoutingTableEntry{
		ZoneID:		zoneID,
		ServiceID:	serviceID,
		NextHopNodeID:	nodeID,
		OverlayID:	HashParts("routing-table-overlay", zoneID),
		Priority:	1,
		WeightBps:	8_000,
		ExpiryHeight:	100,
	})
	require.NoError(t, err)
	table, err := NewRoutingTable(RoutingTable{Epoch: epoch, Routes: []DNLRoutingTableEntry{route}})
	require.NoError(t, err)
	return table
}
