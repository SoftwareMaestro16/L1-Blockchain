package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAetherMeshRoutingGraphSetConnectsAllGraphRootsDeterministically(t *testing.T) {
	zoneA, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"CONTRACT_ZONE",
		Enabled:		true,
		CommittedGasCost:	100,
		CongestionWeightBps:	2500,
		ForwardingFeeWeight:	7,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(132), zoneA.EdgeWeight)

	zoneB, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:		"CONTRACT_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		Enabled:		true,
		CommittedGasCost:	80,
	})
	require.NoError(t, err)

	service, err := NewAetherMeshServiceEdge(AetherMeshServiceEdge{
		SourceService:		"svc.payments",
		DependencyService:	"svc.storage",
		InterfaceHash:		HashParts("iface", "payments-storage"),
		InterfaceCompatible:	true,
		AvailabilityCommitment:	HashParts("availability", "svc.storage"),
		AvailabilityWeightBps:	9500,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(501), service.EdgeWeight)

	message, err := NewAetherMeshMessageEdge(AetherMeshMessageEdge{
		SourceQueue:		HashParts("queue", "contract-outbox"),
		DestinationQueue:	HashParts("queue", "financial-inbox"),
		DeliveryLane:		"cross-zone/settlement",
		QueueBacklog:		13,
		ForwardingFee:		5,
		PriorityWeightBps:	9000,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1018), message.EdgeWeight)

	paymentRoot := HashParts("payment-route-graph")
	storageRoot := HashParts("storage-retrieval-graph")
	left, err := BuildAetherMeshRoutingGraphSet(7, []AetherMeshZoneEdge{zoneA, zoneB}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, paymentRoot, storageRoot)
	require.NoError(t, err)
	require.NoError(t, left.Validate())
	require.Equal(t, paymentRoot, left.PaymentRoot)
	require.Equal(t, storageRoot, left.StorageRoot)
	require.Len(t, left.Graphs, 5)

	right, err := BuildAetherMeshRoutingGraphSet(7, []AetherMeshZoneEdge{zoneB, zoneA}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, paymentRoot, storageRoot)
	require.NoError(t, err)
	require.Equal(t, left.GraphSetRoot, right.GraphSetRoot)
	require.Equal(t, left.Graphs, right.Graphs)
}

func TestAetherMeshRoutingGraphsRejectInvalidWeightsAndUncommittedDependencies(t *testing.T) {
	_, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"CONTRACT_ZONE",
		Enabled:		false,
		CommittedGasCost:	100,
	})
	require.ErrorContains(t, err, "enabled")

	_, err = NewAetherMeshServiceEdge(AetherMeshServiceEdge{
		SourceService:		"svc.a",
		DependencyService:	"svc.b",
		InterfaceHash:		HashParts("iface"),
		InterfaceCompatible:	false,
		AvailabilityCommitment:	HashParts("availability"),
		AvailabilityWeightBps:	9000,
	})
	require.ErrorContains(t, err, "interface compatible")

	_, err = NewAetherMeshMessageEdge(AetherMeshMessageEdge{
		SourceQueue:		HashParts("queue", "a"),
		DestinationQueue:	HashParts("queue", "b"),
		DeliveryLane:		"lane.primary",
		QueueBacklog:		1,
		ForwardingFee:		0,
		PriorityWeightBps:	1000,
	})
	require.ErrorContains(t, err, "forwarding fee")
}

func TestAetherMeshGraphSetRejectsRootTampering(t *testing.T) {
	zone, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"CONTRACT_ZONE",
		Enabled:		true,
		CommittedGasCost:	10,
	})
	require.NoError(t, err)
	graphSet, err := BuildAetherMeshRoutingGraphSet(9, []AetherMeshZoneEdge{zone}, nil, nil, "", "")
	require.NoError(t, err)

	tampered := graphSet
	tampered.Graphs = append([]AetherMeshCommittedGraph(nil), graphSet.Graphs...)
	tampered.Graphs[0].RootHash = HashParts("wrong-root")
	tampered.GraphSetRoot = ComputeAetherMeshRoutingGraphSetRoot(tampered)
	require.ErrorContains(t, tampered.Validate(), "committed graph root mismatch")

	tamperedSetRoot := graphSet
	tamperedSetRoot.GraphSetRoot = HashParts("wrong-set-root")
	require.ErrorContains(t, tamperedSetRoot.Validate(), "graph set root mismatch")
}

func TestAetherMeshRouteSelectionCommitsLowestScoreWithLexicographicTieBreak(t *testing.T) {
	edges := []AetherMeshRouteEdge{
		testMeshRouteEdge(t, "aa-disabled-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", false, 200, 1, 0, 0, 0),
		testMeshRouteEdge(t, "aa-expired-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 99, 1, 0, 0, 0),
		testMeshRouteEdge(t, "b-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 10, 2, 3, 4),
		testMeshRouteEdge(t, "a-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 10, 2, 3, 4),
	}
	req := testMeshRouteRequest(edges, 100, 3)

	selected, err := SelectAetherMeshRoute(req, []AetherMeshRouteEdge{edges[2], edges[0], edges[3], edges[1]})
	require.NoError(t, err)
	require.Equal(t, "a-direct", selected.RouteID)
	require.Equal(t, uint64(20), selected.Score)
	require.Equal(t, uint32(1), selected.HopCount)
	require.Equal(t, req.RoutingTableRoot, selected.RoutingTableRoot)
	require.Equal(t, req.CongestionSnapshotRoot, selected.CongestionSnapshotRoot)
	require.NoError(t, selected.Validate())
}

func TestAetherMeshRouteSelectionBuildsMultiHopCandidateDeterministically(t *testing.T) {
	edges := []AetherMeshRouteEdge{
		testMeshRouteEdge(t, "z-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 100, 0, 0, 0),
		testMeshRouteEdge(t, "a-hop1", "APPLICATION_ZONE", "CONTRACT_ZONE", true, 200, 8, 1, 1, 1),
		testMeshRouteEdge(t, "a-hop2", "CONTRACT_ZONE", "FINANCIAL_ZONE", true, 200, 9, 1, 1, 1),
	}
	req := testMeshRouteRequest(edges, 100, 2)

	selected, err := SelectAetherMeshRoute(req, []AetherMeshRouteEdge{edges[2], edges[0], edges[1]})
	require.NoError(t, err)
	require.Equal(t, "a-hop1/a-hop2", selected.RouteID)
	require.Equal(t, uint64(25), selected.Score)
	require.Equal(t, uint32(2), selected.HopCount)

	tooShallow := req
	tooShallow.MaxHops = 1
	selected, err = SelectAetherMeshRoute(tooShallow, edges)
	require.NoError(t, err)
	require.Equal(t, "z-direct", selected.RouteID)
}

func TestAetherMeshRouteSelectionRejectsUncommittedInputs(t *testing.T) {
	_, err := NewAetherMeshRouteEdge(AetherMeshRouteEdge{
		RouteID:		"bad-reliability",
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		Enabled:		true,
		ExpiresHeight:		200,
	})
	require.ErrorContains(t, err, "reliability commitment")

	edges := []AetherMeshRouteEdge{
		testMeshRouteEdge(t, "direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 10, 0, 0, 0),
	}
	req := testMeshRouteRequest(edges, 100, 1)
	req.RoutingTableRoot = HashParts("wrong-routing-root")
	_, err = SelectAetherMeshRoute(req, edges)
	require.ErrorContains(t, err, "routing table root mismatch")

	req = testMeshRouteRequest(edges, 100, 0)
	_, err = SelectAetherMeshRoute(req, edges)
	require.ErrorContains(t, err, "max hops")
}

func TestAetherMeshRoutingEpochStateCommitsGraphsAndRouteProof(t *testing.T) {
	zone, service, message := testMeshGraphEdges(t)
	routes := []AetherMeshRouteEdge{
		testMeshRouteEdge(t, "z-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 100, 0, 0, 0),
		testMeshRouteEdge(t, "a-hop1", "APPLICATION_ZONE", "CONTRACT_ZONE", true, 200, 8, 1, 1, 1),
		testMeshRouteEdge(t, "a-hop2", "CONTRACT_ZONE", "FINANCIAL_ZONE", true, 200, 9, 1, 1, 1),
	}
	state, err := BuildAetherMeshRoutingEpochState(11, 100, 200, []AetherMeshZoneEdge{zone}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, routes, DefaultAetherMeshRoutingCostParams())
	require.NoError(t, err)
	require.NoError(t, state.Validate())
	require.Equal(t, ComputeAetherMeshZoneGraphRoot(11, []AetherMeshZoneEdge{zone}), state.ZoneGraph.RootHash)
	require.Equal(t, ComputeAetherMeshServiceGraphRoot(11, []AetherMeshServiceEdge{service}), state.ServiceGraph.RootHash)
	require.Equal(t, ComputeAetherMeshMessageRouteGraphRoot(11, []AetherMeshMessageEdge{message}, routes), state.MessageRouteGraph.RootHash)
	require.Equal(t, state.CongestionSnapshot.SnapshotRoot, state.CongestionSnapshotRoot)

	proofReq := testMeshRouteProofRequest(120, 2)
	proof, err := QueryAetherMeshRouteProof(state, proofReq)
	require.NoError(t, err)
	require.NoError(t, proof.Validate())
	require.Equal(t, "a-hop1/a-hop2", proof.SelectedRoute.RouteID)
	require.Equal(t, uint64(25), proof.SelectedRoute.Score)
	require.Len(t, proof.HopEdgeHashes, 2)

	simulation, err := SimulateAetherMeshRouting(state, proofReq)
	require.NoError(t, err)
	require.Len(t, simulation.Candidates, 2)
	require.Equal(t, proof.ProofHash, simulation.Proof.ProofHash)
	require.Equal(t, proof.SelectedRoute.MetadataHash, simulation.Selected.MetadataHash)
}

func TestAetherMeshRoutingSimulationIsDeterministicAcrossInputOrder(t *testing.T) {
	zone, service, message := testMeshGraphEdges(t)
	routes := []AetherMeshRouteEdge{
		testMeshRouteEdge(t, "b-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 10, 2, 3, 4),
		testMeshRouteEdge(t, "a-direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 10, 2, 3, 4),
		testMeshRouteEdge(t, "expired", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 110, 1, 0, 0, 0),
	}
	left, err := BuildAetherMeshRoutingEpochState(12, 100, 200, []AetherMeshZoneEdge{zone}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, routes, DefaultAetherMeshRoutingCostParams())
	require.NoError(t, err)
	right, err := BuildAetherMeshRoutingEpochState(12, 100, 200, []AetherMeshZoneEdge{zone}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, []AetherMeshRouteEdge{routes[2], routes[1], routes[0]}, DefaultAetherMeshRoutingCostParams())
	require.NoError(t, err)
	require.Equal(t, left.EpochRoot, right.EpochRoot)

	req := testMeshRouteProofRequest(120, 1)
	leftSim, err := SimulateAetherMeshRouting(left, req)
	require.NoError(t, err)
	rightSim, err := SimulateAetherMeshRouting(right, req)
	require.NoError(t, err)
	require.Equal(t, "a-direct", leftSim.Selected.RouteID)
	require.Equal(t, leftSim.Selected.MetadataHash, rightSim.Selected.MetadataHash)
	require.Equal(t, leftSim.Proof.ProofHash, rightSim.Proof.ProofHash)
}

func TestAetherMeshRoutingEpochStateRejectsTamperingAndOutOfEpochProofs(t *testing.T) {
	zone, service, message := testMeshGraphEdges(t)
	routes := []AetherMeshRouteEdge{
		testMeshRouteEdge(t, "direct", "APPLICATION_ZONE", "FINANCIAL_ZONE", true, 200, 10, 0, 0, 0),
	}
	state, err := BuildAetherMeshRoutingEpochState(13, 100, 200, []AetherMeshZoneEdge{zone}, []AetherMeshServiceEdge{service}, []AetherMeshMessageEdge{message}, routes, DefaultAetherMeshRoutingCostParams())
	require.NoError(t, err)

	tampered := state
	tampered.MessageRouteGraph.RootHash = HashParts("wrong-message-route-root")
	tampered.EpochRoot = ComputeAetherMeshRoutingEpochRoot(tampered)
	require.ErrorContains(t, tampered.Validate(), "message route graph state root mismatch")

	tampered = state
	tampered.CongestionSnapshot.SnapshotRoot = HashParts("wrong-congestion-root")
	tampered.CongestionSnapshotRoot = tampered.CongestionSnapshot.SnapshotRoot
	tampered.EpochRoot = ComputeAetherMeshRoutingEpochRoot(tampered)
	require.ErrorContains(t, tampered.Validate(), "congestion snapshot root mismatch")

	_, err = QueryAetherMeshRouteProof(state, testMeshRouteProofRequest(200, 1))
	require.ErrorContains(t, err, "outside routing epoch")
}

func testMeshRouteEdge(t *testing.T, routeID, source, destination string, enabled bool, expiresHeight, gasCost, congestion, inverseReliability, latency uint64) AetherMeshRouteEdge {
	t.Helper()
	edge, err := NewAetherMeshRouteEdge(AetherMeshRouteEdge{
		RouteID:			routeID,
		SourceZone:			source,
		DestinationZone:		destination,
		Enabled:			enabled,
		ExpiresHeight:			expiresHeight,
		CommittedGasCost:		gasCost,
		CommittedCongestion:		congestion,
		InverseReliabilityScore:	inverseReliability,
		CommittedLatencyBucket:		latency,
		ReliabilityCommitment:		HashParts("reliability", routeID),
	})
	require.NoError(t, err)
	return edge
}

func testMeshGraphEdges(t *testing.T) (AetherMeshZoneEdge, AetherMeshServiceEdge, AetherMeshMessageEdge) {
	t.Helper()
	zone, err := NewAetherMeshZoneEdge(AetherMeshZoneEdge{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		Enabled:		true,
		CommittedGasCost:	10,
	})
	require.NoError(t, err)
	service, err := NewAetherMeshServiceEdge(AetherMeshServiceEdge{
		SourceService:		"svc.payments",
		DependencyService:	"svc.storage",
		InterfaceHash:		HashParts("iface", "epoch"),
		InterfaceCompatible:	true,
		AvailabilityCommitment:	HashParts("availability", "epoch"),
		AvailabilityWeightBps:	9500,
	})
	require.NoError(t, err)
	message, err := NewAetherMeshMessageEdge(AetherMeshMessageEdge{
		SourceQueue:		HashParts("queue", "epoch-outbox"),
		DestinationQueue:	HashParts("queue", "epoch-inbox"),
		DeliveryLane:		"epoch.cross-zone",
		ForwardingFee:		1,
	})
	require.NoError(t, err)
	return zone, service, message
}

func testMeshRouteRequest(edges []AetherMeshRouteEdge, currentHeight uint64, maxHops uint32) AetherMeshRouteSelectionRequest {
	return AetherMeshRouteSelectionRequest{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		Sender:			"aether1sender",
		Recipient:		"aether1recipient",
		Opcode:			"cross_zone.transfer",
		RoutingTableRoot:	ComputeAetherMeshCommittedRoutingTableRoot(edges),
		CongestionSnapshotRoot:	ComputeAetherMeshCommittedCongestionSnapshotRoot(edges),
		MaxHops:		maxHops,
		CurrentHeight:		currentHeight,
		CostParams:		DefaultAetherMeshRoutingCostParams(),
	}
}

func testMeshRouteProofRequest(currentHeight uint64, maxHops uint32) AetherMeshRouteProofRequest {
	return AetherMeshRouteProofRequest{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		Sender:			"aether1sender",
		Recipient:		"aether1recipient",
		Opcode:			"cross_zone.transfer",
		MaxHops:		maxHops,
		CurrentHeight:		currentHeight,
	}
}
