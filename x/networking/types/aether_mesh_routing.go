package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const MaxAetherMeshRouteHops = 16

type AetherMeshRoutingCostParams struct {
	BaseHopCost		uint64
	GasCostWeight		uint64
	CongestionWeight	uint64
	ReliabilityWeight	uint64
	LatencyWeight		uint64
}

type AetherMeshRouteEdge struct {
	RouteID			string
	SourceZone		string
	DestinationZone		string
	Enabled			bool
	ExpiresHeight		uint64
	CommittedGasCost	uint64
	CommittedCongestion	uint64
	InverseReliabilityScore	uint64
	CommittedLatencyBucket	uint64
	ReliabilityCommitment	string
	EdgeHash		string
}

type AetherMeshRouteSelectionRequest struct {
	SourceZone		string
	DestinationZone		string
	Sender			string
	Recipient		string
	Opcode			string
	RoutingTableRoot	string
	CongestionSnapshotRoot	string
	MaxHops			uint32
	CurrentHeight		uint64
	CostParams		AetherMeshRoutingCostParams
}

type AetherMeshRouteCandidate struct {
	RouteID		string
	Hops		[]AetherMeshRouteEdge
	Score		uint64
	RouteHash	string
}

type AetherMeshSelectedRouteMetadata struct {
	RouteID			string
	SourceZone		string
	DestinationZone		string
	Sender			string
	Recipient		string
	Opcode			string
	RoutingTableRoot	string
	CongestionSnapshotRoot	string
	Score			uint64
	HopCount		uint32
	RouteHash		string
	MetadataHash		string
}

func DefaultAetherMeshRoutingCostParams() AetherMeshRoutingCostParams {
	return AetherMeshRoutingCostParams{
		BaseHopCost:		1,
		GasCostWeight:		1,
		CongestionWeight:	1,
		ReliabilityWeight:	1,
		LatencyWeight:		1,
	}
}

func NewAetherMeshRouteEdge(edge AetherMeshRouteEdge) (AetherMeshRouteEdge, error) {
	edge = normalizeAetherMeshRouteEdge(edge)
	if edge.EdgeHash == "" {
		edge.EdgeHash = ComputeAetherMeshRouteEdgeHash(edge)
	}
	return edge, edge.Validate()
}

func SelectAetherMeshRoute(req AetherMeshRouteSelectionRequest, routingTable []AetherMeshRouteEdge) (AetherMeshSelectedRouteMetadata, error) {
	req = normalizeAetherMeshRouteSelectionRequest(req)
	if err := req.Validate(); err != nil {
		return AetherMeshSelectedRouteMetadata{}, err
	}
	edges := normalizeAetherMeshRouteEdges(routingTable)
	for _, edge := range edges {
		if err := edge.Validate(); err != nil {
			return AetherMeshSelectedRouteMetadata{}, err
		}
	}
	if req.RoutingTableRoot != ComputeAetherMeshCommittedRoutingTableRoot(edges) {
		return AetherMeshSelectedRouteMetadata{}, errors.New("aether mesh committed routing table root mismatch")
	}
	if req.CongestionSnapshotRoot != ComputeAetherMeshCommittedCongestionSnapshotRoot(edges) {
		return AetherMeshSelectedRouteMetadata{}, errors.New("aether mesh committed congestion snapshot root mismatch")
	}
	candidates, err := BuildAetherMeshRouteCandidates(req, edges)
	if err != nil {
		return AetherMeshSelectedRouteMetadata{}, err
	}
	if len(candidates) == 0 {
		return AetherMeshSelectedRouteMetadata{}, errors.New("aether mesh route selection found no candidate path")
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score < candidates[j].Score
		}
		return candidates[i].RouteID < candidates[j].RouteID
	})
	selected := candidates[0]
	metadata := AetherMeshSelectedRouteMetadata{
		RouteID:		selected.RouteID,
		SourceZone:		req.SourceZone,
		DestinationZone:	req.DestinationZone,
		Sender:			req.Sender,
		Recipient:		req.Recipient,
		Opcode:			req.Opcode,
		RoutingTableRoot:	req.RoutingTableRoot,
		CongestionSnapshotRoot:	req.CongestionSnapshotRoot,
		Score:			selected.Score,
		HopCount:		uint32(len(selected.Hops)),
		RouteHash:		selected.RouteHash,
	}
	metadata.MetadataHash = ComputeAetherMeshSelectedRouteMetadataHash(metadata)
	return metadata, metadata.Validate()
}

func BuildAetherMeshRouteCandidates(req AetherMeshRouteSelectionRequest, routingTable []AetherMeshRouteEdge) ([]AetherMeshRouteCandidate, error) {
	req = normalizeAetherMeshRouteSelectionRequest(req)
	if err := req.Validate(); err != nil {
		return nil, err
	}
	edges := normalizeAetherMeshRouteEdges(routingTable)
	adjacency := make(map[string][]AetherMeshRouteEdge)
	for _, edge := range edges {
		if err := edge.Validate(); err != nil {
			return nil, err
		}
		if !edge.Enabled || edge.ExpiresHeight <= req.CurrentHeight {
			continue
		}
		adjacency[edge.SourceZone] = append(adjacency[edge.SourceZone], edge)
	}
	for source := range adjacency {
		sort.SliceStable(adjacency[source], func(i, j int) bool {
			return aetherMeshRouteEdgeKey(adjacency[source][i]) < aetherMeshRouteEdgeKey(adjacency[source][j])
		})
	}

	var out []AetherMeshRouteCandidate
	var walk func(zone string, path []AetherMeshRouteEdge, visited map[string]struct{}) error
	walk = func(zone string, path []AetherMeshRouteEdge, visited map[string]struct{}) error {
		if uint32(len(path)) > req.MaxHops {
			return nil
		}
		if zone == req.DestinationZone && len(path) > 0 {
			candidate, err := BuildAetherMeshRouteCandidate(req, path)
			if err != nil {
				return err
			}
			out = append(out, candidate)
			return nil
		}
		if uint32(len(path)) == req.MaxHops {
			return nil
		}
		for _, edge := range adjacency[zone] {
			if _, seen := visited[edge.DestinationZone]; seen {
				continue
			}
			nextVisited := make(map[string]struct{}, len(visited)+1)
			for key := range visited {
				nextVisited[key] = struct{}{}
			}
			nextVisited[edge.DestinationZone] = struct{}{}
			nextPath := append(append([]AetherMeshRouteEdge(nil), path...), edge)
			if err := walk(edge.DestinationZone, nextPath, nextVisited); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(req.SourceZone, nil, map[string]struct{}{req.SourceZone: {}}); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score < out[j].Score
		}
		return out[i].RouteID < out[j].RouteID
	})
	return out, nil
}

func BuildAetherMeshRouteCandidate(req AetherMeshRouteSelectionRequest, hops []AetherMeshRouteEdge) (AetherMeshRouteCandidate, error) {
	req = normalizeAetherMeshRouteSelectionRequest(req)
	if err := req.Validate(); err != nil {
		return AetherMeshRouteCandidate{}, err
	}
	if len(hops) == 0 || uint32(len(hops)) > req.MaxHops {
		return AetherMeshRouteCandidate{}, errors.New("aether mesh route candidate hop count is invalid")
	}
	normalized := normalizeAetherMeshRoutePathEdges(hops)
	if normalized[0].SourceZone != req.SourceZone || normalized[len(normalized)-1].DestinationZone != req.DestinationZone {
		return AetherMeshRouteCandidate{}, errors.New("aether mesh route candidate endpoints mismatch")
	}
	score := uint64(0)
	routeIDs := make([]string, 0, len(normalized))
	for i, edge := range normalized {
		if err := edge.Validate(); err != nil {
			return AetherMeshRouteCandidate{}, err
		}
		if i > 0 && normalized[i-1].DestinationZone != edge.SourceZone {
			return AetherMeshRouteCandidate{}, errors.New("aether mesh route candidate path is disconnected")
		}
		edgeCost, err := ComputeAetherMeshRouteEdgeCost(edge, req.CostParams)
		if err != nil {
			return AetherMeshRouteCandidate{}, err
		}
		score, err = checkedAetherMeshAdd(score, edgeCost)
		if err != nil {
			return AetherMeshRouteCandidate{}, err
		}
		routeIDs = append(routeIDs, edge.RouteID)
	}
	candidate := AetherMeshRouteCandidate{
		RouteID:	strings.Join(routeIDs, "/"),
		Hops:		normalized,
		Score:		score,
	}
	candidate.RouteHash = ComputeAetherMeshRouteCandidateHash(req, candidate)
	return candidate, nil
}

func (req AetherMeshRouteSelectionRequest) Validate() error {
	req = normalizeAetherMeshRouteSelectionRequest(req)
	if err := validateIdentifierSet("aether mesh route source zone", []string{req.SourceZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh route destination zone", []string{req.DestinationZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if req.SourceZone == req.DestinationZone {
		return errors.New("aether mesh route endpoints must differ")
	}
	if err := validateMeshGraphToken("aether mesh route sender", req.Sender); err != nil {
		return err
	}
	if err := validateMeshGraphToken("aether mesh route recipient", req.Recipient); err != nil {
		return err
	}
	if err := validateMeshGraphToken("aether mesh route opcode", req.Opcode); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh committed routing table root", req.RoutingTableRoot); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh committed congestion snapshot root", req.CongestionSnapshotRoot); err != nil {
		return err
	}
	if req.MaxHops == 0 || req.MaxHops > MaxAetherMeshRouteHops {
		return fmt.Errorf("aether mesh route max hops must be between 1 and %d", MaxAetherMeshRouteHops)
	}
	if req.CurrentHeight == 0 {
		return errors.New("aether mesh route current height must be positive")
	}
	return req.CostParams.Validate()
}

func (params AetherMeshRoutingCostParams) Validate() error {
	params = normalizeAetherMeshRoutingCostParams(params)
	if params.BaseHopCost == 0 {
		return errors.New("aether mesh route base hop cost must be positive")
	}
	return nil
}

func (edge AetherMeshRouteEdge) Validate() error {
	edge = normalizeAetherMeshRouteEdge(edge)
	if err := validateMeshGraphToken("aether mesh route id", edge.RouteID); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh route source zone", []string{edge.SourceZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh route destination zone", []string{edge.DestinationZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if edge.SourceZone == edge.DestinationZone {
		return errors.New("aether mesh route edge endpoints must differ")
	}
	if edge.ExpiresHeight == 0 {
		return errors.New("aether mesh route edge expiry height must be positive")
	}
	if err := ValidateHash("aether mesh route reliability commitment", edge.ReliabilityCommitment); err != nil {
		return err
	}
	return validateAetherMeshEdgeHash("aether mesh route edge", edge.EdgeHash, ComputeAetherMeshRouteEdgeHash(edge))
}

func (metadata AetherMeshSelectedRouteMetadata) Validate() error {
	metadata = normalizeAetherMeshSelectedRouteMetadata(metadata)
	if err := validateAetherMeshRoutePathID("aether mesh selected route id", metadata.RouteID); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh selected source zone", []string{metadata.SourceZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh selected destination zone", []string{metadata.DestinationZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if metadata.HopCount == 0 || metadata.HopCount > MaxAetherMeshRouteHops {
		return errors.New("aether mesh selected route hop count is invalid")
	}
	if err := ValidateHash("aether mesh selected routing table root", metadata.RoutingTableRoot); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh selected congestion snapshot root", metadata.CongestionSnapshotRoot); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh selected route hash", metadata.RouteHash); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh selected route metadata hash", metadata.MetadataHash); err != nil {
		return err
	}
	if metadata.MetadataHash != ComputeAetherMeshSelectedRouteMetadataHash(metadata) {
		return errors.New("aether mesh selected route metadata hash mismatch")
	}
	return nil
}

func ComputeAetherMeshRouteEdgeCost(edge AetherMeshRouteEdge, params AetherMeshRoutingCostParams) (uint64, error) {
	edge = normalizeAetherMeshRouteEdge(edge)
	params = normalizeAetherMeshRoutingCostParams(params)
	total := params.BaseHopCost
	for _, term := range []struct {
		weight	uint64
		value	uint64
	}{
		{params.GasCostWeight, edge.CommittedGasCost},
		{params.CongestionWeight, edge.CommittedCongestion},
		{params.ReliabilityWeight, edge.InverseReliabilityScore},
		{params.LatencyWeight, edge.CommittedLatencyBucket},
	} {
		product, err := checkedAetherMeshMul(term.weight, term.value)
		if err != nil {
			return 0, err
		}
		total, err = checkedAetherMeshAdd(total, product)
		if err != nil {
			return 0, err
		}
	}
	return total, nil
}

func ComputeAetherMeshCommittedRoutingTableRoot(edges []AetherMeshRouteEdge) string {
	out := normalizeAetherMeshRouteEdges(edges)
	parts := []string{"aether-mesh-committed-routing-table"}
	for _, edge := range out {
		parts = append(parts, edge.EdgeHash)
	}
	return HashParts(parts...)
}

func ComputeAetherMeshCommittedCongestionSnapshotRoot(edges []AetherMeshRouteEdge) string {
	out := normalizeAetherMeshRouteEdges(edges)
	parts := []string{"aether-mesh-committed-congestion-snapshot"}
	for _, edge := range out {
		parts = append(parts, edge.RouteID, fmt.Sprintf("%020d", edge.CommittedCongestion), fmt.Sprintf("%020d", edge.InverseReliabilityScore), fmt.Sprintf("%020d", edge.CommittedLatencyBucket), edge.ReliabilityCommitment)
	}
	return HashParts(parts...)
}

func ComputeAetherMeshRouteEdgeHash(edge AetherMeshRouteEdge) string {
	edge = normalizeAetherMeshRouteEdge(edge)
	return HashParts("aether-mesh-route-edge", edge.RouteID, edge.SourceZone, edge.DestinationZone, fmt.Sprintf("%t", edge.Enabled), fmt.Sprintf("%020d", edge.ExpiresHeight), fmt.Sprintf("%020d", edge.CommittedGasCost), fmt.Sprintf("%020d", edge.CommittedCongestion), fmt.Sprintf("%020d", edge.InverseReliabilityScore), fmt.Sprintf("%020d", edge.CommittedLatencyBucket), edge.ReliabilityCommitment)
}

func ComputeAetherMeshRouteCandidateHash(req AetherMeshRouteSelectionRequest, candidate AetherMeshRouteCandidate) string {
	req = normalizeAetherMeshRouteSelectionRequest(req)
	candidate = normalizeAetherMeshRouteCandidate(candidate)
	parts := []string{"aether-mesh-route-candidate", req.RoutingTableRoot, req.CongestionSnapshotRoot, candidate.RouteID, fmt.Sprintf("%020d", candidate.Score)}
	for _, hop := range candidate.Hops {
		parts = append(parts, hop.EdgeHash)
	}
	return HashParts(parts...)
}

func ComputeAetherMeshSelectedRouteMetadataHash(metadata AetherMeshSelectedRouteMetadata) string {
	metadata = normalizeAetherMeshSelectedRouteMetadata(metadata)
	return HashParts("aether-mesh-selected-route-metadata", metadata.RouteID, metadata.SourceZone, metadata.DestinationZone, metadata.Sender, metadata.Recipient, metadata.Opcode, metadata.RoutingTableRoot, metadata.CongestionSnapshotRoot, fmt.Sprintf("%020d", metadata.Score), fmt.Sprintf("%010d", metadata.HopCount), metadata.RouteHash)
}

func normalizeAetherMeshRouteSelectionRequest(req AetherMeshRouteSelectionRequest) AetherMeshRouteSelectionRequest {
	req.SourceZone = strings.TrimSpace(req.SourceZone)
	req.DestinationZone = strings.TrimSpace(req.DestinationZone)
	req.Sender = strings.TrimSpace(req.Sender)
	req.Recipient = strings.TrimSpace(req.Recipient)
	req.Opcode = strings.TrimSpace(req.Opcode)
	req.RoutingTableRoot = normalizeHashText(req.RoutingTableRoot)
	req.CongestionSnapshotRoot = normalizeHashText(req.CongestionSnapshotRoot)
	req.CostParams = normalizeAetherMeshRoutingCostParams(req.CostParams)
	return req
}

func normalizeAetherMeshRoutingCostParams(params AetherMeshRoutingCostParams) AetherMeshRoutingCostParams {
	if params == (AetherMeshRoutingCostParams{}) {
		return DefaultAetherMeshRoutingCostParams()
	}
	return params
}

func normalizeAetherMeshRouteEdge(edge AetherMeshRouteEdge) AetherMeshRouteEdge {
	edge.RouteID = strings.TrimSpace(edge.RouteID)
	edge.SourceZone = strings.TrimSpace(edge.SourceZone)
	edge.DestinationZone = strings.TrimSpace(edge.DestinationZone)
	edge.ReliabilityCommitment = normalizeHashText(edge.ReliabilityCommitment)
	edge.EdgeHash = normalizeHashText(edge.EdgeHash)
	return edge
}

func normalizeAetherMeshRouteEdges(edges []AetherMeshRouteEdge) []AetherMeshRouteEdge {
	out := make([]AetherMeshRouteEdge, len(edges))
	for i, edge := range edges {
		edge = normalizeAetherMeshRouteEdge(edge)
		if edge.EdgeHash == "" {
			edge.EdgeHash = ComputeAetherMeshRouteEdgeHash(edge)
		}
		out[i] = edge
	}
	sort.SliceStable(out, func(i, j int) bool {
		return aetherMeshRouteEdgeKey(out[i]) < aetherMeshRouteEdgeKey(out[j])
	})
	return out
}

func normalizeAetherMeshRoutePathEdges(edges []AetherMeshRouteEdge) []AetherMeshRouteEdge {
	out := make([]AetherMeshRouteEdge, len(edges))
	for i, edge := range edges {
		edge = normalizeAetherMeshRouteEdge(edge)
		if edge.EdgeHash == "" {
			edge.EdgeHash = ComputeAetherMeshRouteEdgeHash(edge)
		}
		out[i] = edge
	}
	return out
}

func normalizeAetherMeshRouteCandidate(candidate AetherMeshRouteCandidate) AetherMeshRouteCandidate {
	candidate.RouteID = strings.TrimSpace(candidate.RouteID)
	candidate.Hops = normalizeAetherMeshRoutePathEdges(candidate.Hops)
	candidate.RouteHash = normalizeHashText(candidate.RouteHash)
	return candidate
}

func normalizeAetherMeshSelectedRouteMetadata(metadata AetherMeshSelectedRouteMetadata) AetherMeshSelectedRouteMetadata {
	metadata.RouteID = strings.TrimSpace(metadata.RouteID)
	metadata.SourceZone = strings.TrimSpace(metadata.SourceZone)
	metadata.DestinationZone = strings.TrimSpace(metadata.DestinationZone)
	metadata.Sender = strings.TrimSpace(metadata.Sender)
	metadata.Recipient = strings.TrimSpace(metadata.Recipient)
	metadata.Opcode = strings.TrimSpace(metadata.Opcode)
	metadata.RoutingTableRoot = normalizeHashText(metadata.RoutingTableRoot)
	metadata.CongestionSnapshotRoot = normalizeHashText(metadata.CongestionSnapshotRoot)
	metadata.RouteHash = normalizeHashText(metadata.RouteHash)
	metadata.MetadataHash = normalizeHashText(metadata.MetadataHash)
	return metadata
}

func aetherMeshRouteEdgeKey(edge AetherMeshRouteEdge) string {
	edge = normalizeAetherMeshRouteEdge(edge)
	return edge.SourceZone + "/" + edge.DestinationZone + "/" + edge.RouteID
}

func validateAetherMeshRoutePathID(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	maxBytes := MaxAetherMeshGraphID * MaxAetherMeshRouteHops
	if len(value) > maxBytes {
		return fmt.Errorf("%s must be <= %d bytes", field, maxBytes)
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character %q", field, r)
	}
	return nil
}

func checkedAetherMeshAdd(left, right uint64) (uint64, error) {
	if right > ^uint64(0)-left {
		return 0, errors.New("aether mesh route cost overflow")
	}
	return left + right, nil
}

func checkedAetherMeshMul(left, right uint64) (uint64, error) {
	if left != 0 && right > ^uint64(0)/left {
		return 0, errors.New("aether mesh route cost overflow")
	}
	return left * right, nil
}
