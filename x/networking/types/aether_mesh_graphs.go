package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxAetherMeshGraphEdges	= 4096
	MaxAetherMeshGraphID	= 128
)

type AetherMeshGraphKind string

const (
	AetherMeshGraphZone	AetherMeshGraphKind	= "zone"
	AetherMeshGraphService	AetherMeshGraphKind	= "service"
	AetherMeshGraphMessage	AetherMeshGraphKind	= "message"
	AetherMeshGraphPayment	AetherMeshGraphKind	= "payment"
	AetherMeshGraphStorage	AetherMeshGraphKind	= "storage"
)

type AetherMeshZoneEdge struct {
	SourceZone		string
	DestinationZone		string
	Enabled			bool
	CommittedGasCost	uint64
	CongestionWeightBps	uint32
	ForwardingFeeWeight	uint64
	EdgeWeight		uint64
	EdgeHash		string
}

type AetherMeshServiceEdge struct {
	SourceService		string
	DependencyService	string
	InterfaceHash		string
	InterfaceCompatible	bool
	AvailabilityCommitment	string
	AvailabilityWeightBps	uint32
	EdgeWeight		uint64
	EdgeHash		string
}

type AetherMeshMessageEdge struct {
	SourceQueue		string
	DestinationQueue	string
	DeliveryLane		string
	QueueBacklog		uint64
	ForwardingFee		uint64
	PriorityWeightBps	uint32
	EdgeWeight		uint64
	EdgeHash		string
}

type AetherMeshCommittedGraph struct {
	Kind		AetherMeshGraphKind
	Epoch		uint64
	RootHash	string
	EdgeCount	uint32
}

type AetherMeshRoutingGraphSet struct {
	Epoch		uint64
	ZoneEdges	[]AetherMeshZoneEdge
	ServiceEdges	[]AetherMeshServiceEdge
	MessageEdges	[]AetherMeshMessageEdge
	PaymentRoot	string
	StorageRoot	string
	Graphs		[]AetherMeshCommittedGraph
	GraphSetRoot	string
}

func NewAetherMeshZoneEdge(edge AetherMeshZoneEdge) (AetherMeshZoneEdge, error) {
	edge = normalizeAetherMeshZoneEdge(edge)
	if edge.EdgeWeight == 0 {
		edge.EdgeWeight = ComputeAetherMeshZoneEdgeWeight(edge)
	}
	if edge.EdgeHash == "" {
		edge.EdgeHash = ComputeAetherMeshZoneEdgeHash(edge)
	}
	return edge, edge.Validate()
}

func NewAetherMeshServiceEdge(edge AetherMeshServiceEdge) (AetherMeshServiceEdge, error) {
	edge = normalizeAetherMeshServiceEdge(edge)
	if edge.EdgeWeight == 0 {
		edge.EdgeWeight = ComputeAetherMeshServiceEdgeWeight(edge)
	}
	if edge.EdgeHash == "" {
		edge.EdgeHash = ComputeAetherMeshServiceEdgeHash(edge)
	}
	return edge, edge.Validate()
}

func NewAetherMeshMessageEdge(edge AetherMeshMessageEdge) (AetherMeshMessageEdge, error) {
	edge = normalizeAetherMeshMessageEdge(edge)
	if edge.EdgeWeight == 0 {
		edge.EdgeWeight = ComputeAetherMeshMessageEdgeWeight(edge)
	}
	if edge.EdgeHash == "" {
		edge.EdgeHash = ComputeAetherMeshMessageEdgeHash(edge)
	}
	return edge, edge.Validate()
}

func BuildAetherMeshRoutingGraphSet(epoch uint64, zoneEdges []AetherMeshZoneEdge, serviceEdges []AetherMeshServiceEdge, messageEdges []AetherMeshMessageEdge, paymentRoot, storageRoot string) (AetherMeshRoutingGraphSet, error) {
	graphSet := AetherMeshRoutingGraphSet{
		Epoch:		epoch,
		ZoneEdges:	normalizeAetherMeshZoneEdges(zoneEdges),
		ServiceEdges:	normalizeAetherMeshServiceEdges(serviceEdges),
		MessageEdges:	normalizeAetherMeshMessageEdges(messageEdges),
		PaymentRoot:	normalizeHashText(paymentRoot),
		StorageRoot:	normalizeHashText(storageRoot),
	}
	if graphSet.PaymentRoot == "" {
		graphSet.PaymentRoot = HashParts("aether-mesh-empty-payment-route-graph", fmt.Sprintf("%020d", epoch))
	}
	if graphSet.StorageRoot == "" {
		graphSet.StorageRoot = HashParts("aether-mesh-empty-storage-retrieval-graph", fmt.Sprintf("%020d", epoch))
	}
	if err := graphSet.ValidateFormat(); err != nil {
		return AetherMeshRoutingGraphSet{}, err
	}
	zoneRoot := ComputeAetherMeshZoneGraphRoot(epoch, graphSet.ZoneEdges)
	serviceRoot := ComputeAetherMeshServiceGraphRoot(epoch, graphSet.ServiceEdges)
	messageRoot := ComputeAetherMeshMessageGraphRoot(epoch, graphSet.MessageEdges)
	graphSet.Graphs = []AetherMeshCommittedGraph{
		{Kind: AetherMeshGraphMessage, Epoch: epoch, RootHash: messageRoot, EdgeCount: uint32(len(graphSet.MessageEdges))},
		{Kind: AetherMeshGraphPayment, Epoch: epoch, RootHash: graphSet.PaymentRoot, EdgeCount: 0},
		{Kind: AetherMeshGraphService, Epoch: epoch, RootHash: serviceRoot, EdgeCount: uint32(len(graphSet.ServiceEdges))},
		{Kind: AetherMeshGraphStorage, Epoch: epoch, RootHash: graphSet.StorageRoot, EdgeCount: 0},
		{Kind: AetherMeshGraphZone, Epoch: epoch, RootHash: zoneRoot, EdgeCount: uint32(len(graphSet.ZoneEdges))},
	}
	graphSet.GraphSetRoot = ComputeAetherMeshRoutingGraphSetRoot(graphSet)
	return graphSet, graphSet.Validate()
}

func (edge AetherMeshZoneEdge) Validate() error {
	edge = normalizeAetherMeshZoneEdge(edge)
	if err := validateIdentifierSet("aether mesh source zone", []string{edge.SourceZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh destination zone", []string{edge.DestinationZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if edge.SourceZone == edge.DestinationZone {
		return errors.New("aether mesh zone edge endpoints must differ")
	}
	if !edge.Enabled {
		return errors.New("aether mesh zone edge must be enabled")
	}
	if edge.CommittedGasCost == 0 {
		return errors.New("aether mesh zone edge committed gas cost must be positive")
	}
	if edge.CongestionWeightBps > BasisPoints {
		return errors.New("aether mesh zone edge congestion bps exceeds maximum")
	}
	if edge.EdgeWeight != ComputeAetherMeshZoneEdgeWeight(edge) {
		return errors.New("aether mesh zone edge weight mismatch")
	}
	return validateAetherMeshEdgeHash("aether mesh zone edge", edge.EdgeHash, ComputeAetherMeshZoneEdgeHash(edge))
}

func (edge AetherMeshServiceEdge) Validate() error {
	edge = normalizeAetherMeshServiceEdge(edge)
	if err := validateIdentifierSet("aether mesh source service", []string{edge.SourceService}, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh dependency service", []string{edge.DependencyService}, MaxServiceIDBytes); err != nil {
		return err
	}
	if edge.SourceService == edge.DependencyService {
		return errors.New("aether mesh service dependency endpoints must differ")
	}
	if err := ValidateHash("aether mesh service interface hash", edge.InterfaceHash); err != nil {
		return err
	}
	if !edge.InterfaceCompatible {
		return errors.New("aether mesh service dependency must be interface compatible")
	}
	if err := ValidateHash("aether mesh service availability commitment", edge.AvailabilityCommitment); err != nil {
		return err
	}
	if edge.AvailabilityWeightBps > BasisPoints {
		return errors.New("aether mesh service availability bps exceeds maximum")
	}
	if edge.EdgeWeight != ComputeAetherMeshServiceEdgeWeight(edge) {
		return errors.New("aether mesh service edge weight mismatch")
	}
	return validateAetherMeshEdgeHash("aether mesh service edge", edge.EdgeHash, ComputeAetherMeshServiceEdgeHash(edge))
}

func (edge AetherMeshMessageEdge) Validate() error {
	edge = normalizeAetherMeshMessageEdge(edge)
	if err := ValidateHash("aether mesh source queue", edge.SourceQueue); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh destination queue", edge.DestinationQueue); err != nil {
		return err
	}
	if edge.SourceQueue == edge.DestinationQueue {
		return errors.New("aether mesh message queues must differ")
	}
	if err := validateMeshGraphToken("aether mesh delivery lane", edge.DeliveryLane); err != nil {
		return err
	}
	if edge.ForwardingFee == 0 {
		return errors.New("aether mesh message edge forwarding fee must be positive")
	}
	if edge.PriorityWeightBps > BasisPoints {
		return errors.New("aether mesh message priority bps exceeds maximum")
	}
	if edge.EdgeWeight != ComputeAetherMeshMessageEdgeWeight(edge) {
		return errors.New("aether mesh message edge weight mismatch")
	}
	return validateAetherMeshEdgeHash("aether mesh message edge", edge.EdgeHash, ComputeAetherMeshMessageEdgeHash(edge))
}

func (graph AetherMeshCommittedGraph) Validate() error {
	if !IsAetherMeshGraphKind(graph.Kind) {
		return fmt.Errorf("unknown aether mesh graph kind %q", graph.Kind)
	}
	if graph.Epoch == 0 {
		return errors.New("aether mesh graph epoch must be positive")
	}
	return ValidateHash("aether mesh graph root", graph.RootHash)
}

func (graphSet AetherMeshRoutingGraphSet) ValidateFormat() error {
	if graphSet.Epoch == 0 {
		return errors.New("aether mesh graph set epoch must be positive")
	}
	if len(graphSet.ZoneEdges)+len(graphSet.ServiceEdges)+len(graphSet.MessageEdges) > MaxAetherMeshGraphEdges {
		return fmt.Errorf("aether mesh graph edges must be <= %d", MaxAetherMeshGraphEdges)
	}
	for _, edge := range graphSet.ZoneEdges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	for _, edge := range graphSet.ServiceEdges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	for _, edge := range graphSet.MessageEdges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("aether mesh payment route graph root", graphSet.PaymentRoot); err != nil {
		return err
	}
	return ValidateHash("aether mesh storage retrieval graph root", graphSet.StorageRoot)
}

func (graphSet AetherMeshRoutingGraphSet) Validate() error {
	graphSet = normalizeAetherMeshRoutingGraphSet(graphSet)
	if err := graphSet.ValidateFormat(); err != nil {
		return err
	}
	expected := []AetherMeshCommittedGraph{
		{Kind: AetherMeshGraphMessage, Epoch: graphSet.Epoch, RootHash: ComputeAetherMeshMessageGraphRoot(graphSet.Epoch, graphSet.MessageEdges), EdgeCount: uint32(len(graphSet.MessageEdges))},
		{Kind: AetherMeshGraphPayment, Epoch: graphSet.Epoch, RootHash: graphSet.PaymentRoot, EdgeCount: 0},
		{Kind: AetherMeshGraphService, Epoch: graphSet.Epoch, RootHash: ComputeAetherMeshServiceGraphRoot(graphSet.Epoch, graphSet.ServiceEdges), EdgeCount: uint32(len(graphSet.ServiceEdges))},
		{Kind: AetherMeshGraphStorage, Epoch: graphSet.Epoch, RootHash: graphSet.StorageRoot, EdgeCount: 0},
		{Kind: AetherMeshGraphZone, Epoch: graphSet.Epoch, RootHash: ComputeAetherMeshZoneGraphRoot(graphSet.Epoch, graphSet.ZoneEdges), EdgeCount: uint32(len(graphSet.ZoneEdges))},
	}
	if len(graphSet.Graphs) != len(expected) {
		return errors.New("aether mesh graph set must commit all graph roots")
	}
	for i, graph := range graphSet.Graphs {
		if err := graph.Validate(); err != nil {
			return err
		}
		if graph != expected[i] {
			return errors.New("aether mesh graph set committed graph root mismatch")
		}
	}
	if err := ValidateHash("aether mesh graph set root", graphSet.GraphSetRoot); err != nil {
		return err
	}
	if graphSet.GraphSetRoot != ComputeAetherMeshRoutingGraphSetRoot(graphSet) {
		return errors.New("aether mesh graph set root mismatch")
	}
	return nil
}

func ComputeAetherMeshZoneEdgeWeight(edge AetherMeshZoneEdge) uint64 {
	edge = normalizeAetherMeshZoneEdge(edge)
	return edge.CommittedGasCost + (edge.CommittedGasCost*uint64(edge.CongestionWeightBps))/uint64(BasisPoints) + edge.ForwardingFeeWeight
}

func ComputeAetherMeshServiceEdgeWeight(edge AetherMeshServiceEdge) uint64 {
	edge = normalizeAetherMeshServiceEdge(edge)
	availabilityPenalty := uint64(BasisPoints - edge.AvailabilityWeightBps)
	if edge.InterfaceCompatible {
		return 1 + availabilityPenalty
	}
	return uint64(BasisPoints) + availabilityPenalty
}

func ComputeAetherMeshMessageEdgeWeight(edge AetherMeshMessageEdge) uint64 {
	edge = normalizeAetherMeshMessageEdge(edge)
	priorityPenalty := uint64(BasisPoints - edge.PriorityWeightBps)
	return edge.QueueBacklog + edge.ForwardingFee + priorityPenalty
}

func ComputeAetherMeshZoneEdgeHash(edge AetherMeshZoneEdge) string {
	edge = normalizeAetherMeshZoneEdge(edge)
	return HashParts("aether-mesh-zone-edge", edge.SourceZone, edge.DestinationZone, fmt.Sprintf("%t", edge.Enabled), fmt.Sprintf("%020d", edge.CommittedGasCost), fmt.Sprintf("%010d", edge.CongestionWeightBps), fmt.Sprintf("%020d", edge.ForwardingFeeWeight), fmt.Sprintf("%020d", edge.EdgeWeight))
}

func ComputeAetherMeshServiceEdgeHash(edge AetherMeshServiceEdge) string {
	edge = normalizeAetherMeshServiceEdge(edge)
	return HashParts("aether-mesh-service-edge", edge.SourceService, edge.DependencyService, edge.InterfaceHash, fmt.Sprintf("%t", edge.InterfaceCompatible), edge.AvailabilityCommitment, fmt.Sprintf("%010d", edge.AvailabilityWeightBps), fmt.Sprintf("%020d", edge.EdgeWeight))
}

func ComputeAetherMeshMessageEdgeHash(edge AetherMeshMessageEdge) string {
	edge = normalizeAetherMeshMessageEdge(edge)
	return HashParts("aether-mesh-message-edge", edge.SourceQueue, edge.DestinationQueue, edge.DeliveryLane, fmt.Sprintf("%020d", edge.QueueBacklog), fmt.Sprintf("%020d", edge.ForwardingFee), fmt.Sprintf("%010d", edge.PriorityWeightBps), fmt.Sprintf("%020d", edge.EdgeWeight))
}

func ComputeAetherMeshZoneGraphRoot(epoch uint64, edges []AetherMeshZoneEdge) string {
	out := normalizeAetherMeshZoneEdges(edges)
	parts := []string{"aether-mesh-zone-graph", fmt.Sprintf("%020d", epoch)}
	for _, edge := range out {
		parts = append(parts, edge.EdgeHash)
	}
	return HashParts(parts...)
}

func ComputeAetherMeshServiceGraphRoot(epoch uint64, edges []AetherMeshServiceEdge) string {
	out := normalizeAetherMeshServiceEdges(edges)
	parts := []string{"aether-mesh-service-graph", fmt.Sprintf("%020d", epoch)}
	for _, edge := range out {
		parts = append(parts, edge.EdgeHash)
	}
	return HashParts(parts...)
}

func ComputeAetherMeshMessageGraphRoot(epoch uint64, edges []AetherMeshMessageEdge) string {
	out := normalizeAetherMeshMessageEdges(edges)
	parts := []string{"aether-mesh-message-graph", fmt.Sprintf("%020d", epoch)}
	for _, edge := range out {
		parts = append(parts, edge.EdgeHash)
	}
	return HashParts(parts...)
}

func ComputeAetherMeshRoutingGraphSetRoot(graphSet AetherMeshRoutingGraphSet) string {
	graphSet = normalizeAetherMeshRoutingGraphSet(graphSet)
	parts := []string{"aether-mesh-routing-graph-set", fmt.Sprintf("%020d", graphSet.Epoch)}
	for _, graph := range graphSet.Graphs {
		parts = append(parts, string(graph.Kind), graph.RootHash, fmt.Sprintf("%010d", graph.EdgeCount))
	}
	return HashParts(parts...)
}

func IsAetherMeshGraphKind(kind AetherMeshGraphKind) bool {
	switch kind {
	case AetherMeshGraphZone, AetherMeshGraphService, AetherMeshGraphMessage, AetherMeshGraphPayment, AetherMeshGraphStorage:
		return true
	default:
		return false
	}
}

func normalizeAetherMeshRoutingGraphSet(graphSet AetherMeshRoutingGraphSet) AetherMeshRoutingGraphSet {
	graphSet.ZoneEdges = normalizeAetherMeshZoneEdges(graphSet.ZoneEdges)
	graphSet.ServiceEdges = normalizeAetherMeshServiceEdges(graphSet.ServiceEdges)
	graphSet.MessageEdges = normalizeAetherMeshMessageEdges(graphSet.MessageEdges)
	graphSet.PaymentRoot = normalizeHashText(graphSet.PaymentRoot)
	graphSet.StorageRoot = normalizeHashText(graphSet.StorageRoot)
	graphSet.Graphs = append([]AetherMeshCommittedGraph(nil), graphSet.Graphs...)
	sort.SliceStable(graphSet.Graphs, func(i, j int) bool {
		return graphSet.Graphs[i].Kind < graphSet.Graphs[j].Kind
	})
	graphSet.GraphSetRoot = normalizeHashText(graphSet.GraphSetRoot)
	return graphSet
}

func normalizeAetherMeshZoneEdge(edge AetherMeshZoneEdge) AetherMeshZoneEdge {
	edge.SourceZone = strings.TrimSpace(edge.SourceZone)
	edge.DestinationZone = strings.TrimSpace(edge.DestinationZone)
	edge.EdgeHash = normalizeHashText(edge.EdgeHash)
	return edge
}

func normalizeAetherMeshServiceEdge(edge AetherMeshServiceEdge) AetherMeshServiceEdge {
	edge.SourceService = strings.TrimSpace(edge.SourceService)
	edge.DependencyService = strings.TrimSpace(edge.DependencyService)
	edge.InterfaceHash = normalizeHashText(edge.InterfaceHash)
	edge.AvailabilityCommitment = normalizeHashText(edge.AvailabilityCommitment)
	edge.EdgeHash = normalizeHashText(edge.EdgeHash)
	return edge
}

func normalizeAetherMeshMessageEdge(edge AetherMeshMessageEdge) AetherMeshMessageEdge {
	edge.SourceQueue = normalizeHashText(edge.SourceQueue)
	edge.DestinationQueue = normalizeHashText(edge.DestinationQueue)
	edge.DeliveryLane = strings.TrimSpace(edge.DeliveryLane)
	edge.EdgeHash = normalizeHashText(edge.EdgeHash)
	return edge
}

func normalizeAetherMeshZoneEdges(edges []AetherMeshZoneEdge) []AetherMeshZoneEdge {
	out := make([]AetherMeshZoneEdge, len(edges))
	for i, edge := range edges {
		edge = normalizeAetherMeshZoneEdge(edge)
		if edge.EdgeWeight == 0 {
			edge.EdgeWeight = ComputeAetherMeshZoneEdgeWeight(edge)
		}
		if edge.EdgeHash == "" {
			edge.EdgeHash = ComputeAetherMeshZoneEdgeHash(edge)
		}
		out[i] = edge
	}
	sort.SliceStable(out, func(i, j int) bool {
		return aetherMeshZoneEdgeKey(out[i]) < aetherMeshZoneEdgeKey(out[j])
	})
	return out
}

func normalizeAetherMeshServiceEdges(edges []AetherMeshServiceEdge) []AetherMeshServiceEdge {
	out := make([]AetherMeshServiceEdge, len(edges))
	for i, edge := range edges {
		edge = normalizeAetherMeshServiceEdge(edge)
		if edge.EdgeWeight == 0 {
			edge.EdgeWeight = ComputeAetherMeshServiceEdgeWeight(edge)
		}
		if edge.EdgeHash == "" {
			edge.EdgeHash = ComputeAetherMeshServiceEdgeHash(edge)
		}
		out[i] = edge
	}
	sort.SliceStable(out, func(i, j int) bool {
		return aetherMeshServiceEdgeKey(out[i]) < aetherMeshServiceEdgeKey(out[j])
	})
	return out
}

func normalizeAetherMeshMessageEdges(edges []AetherMeshMessageEdge) []AetherMeshMessageEdge {
	out := make([]AetherMeshMessageEdge, len(edges))
	for i, edge := range edges {
		edge = normalizeAetherMeshMessageEdge(edge)
		if edge.EdgeWeight == 0 {
			edge.EdgeWeight = ComputeAetherMeshMessageEdgeWeight(edge)
		}
		if edge.EdgeHash == "" {
			edge.EdgeHash = ComputeAetherMeshMessageEdgeHash(edge)
		}
		out[i] = edge
	}
	sort.SliceStable(out, func(i, j int) bool {
		return aetherMeshMessageEdgeKey(out[i]) < aetherMeshMessageEdgeKey(out[j])
	})
	return out
}

func aetherMeshZoneEdgeKey(edge AetherMeshZoneEdge) string {
	edge = normalizeAetherMeshZoneEdge(edge)
	return edge.SourceZone + "/" + edge.DestinationZone
}

func aetherMeshServiceEdgeKey(edge AetherMeshServiceEdge) string {
	edge = normalizeAetherMeshServiceEdge(edge)
	return edge.SourceService + "/" + edge.DependencyService
}

func aetherMeshMessageEdgeKey(edge AetherMeshMessageEdge) string {
	edge = normalizeAetherMeshMessageEdge(edge)
	return edge.SourceQueue + "/" + edge.DestinationQueue + "/" + edge.DeliveryLane
}

func validateAetherMeshEdgeHash(label, actual, expected string) error {
	if err := ValidateHash(label+" hash", actual); err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("%s hash mismatch", label)
	}
	return nil
}

func validateMeshGraphToken(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > MaxAetherMeshGraphID {
		return fmt.Errorf("%s must be <= %d bytes", field, MaxAetherMeshGraphID)
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character %q", field, r)
	}
	return nil
}
