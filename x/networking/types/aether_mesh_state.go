package types

import (
	"errors"
	"fmt"
	"strings"
)

type AetherMeshZoneGraphState struct {
	Epoch		uint64
	Edges		[]AetherMeshZoneEdge
	RootHash	string
}

type AetherMeshServiceGraphState struct {
	Epoch		uint64
	Edges		[]AetherMeshServiceEdge
	RootHash	string
}

type AetherMeshMessageRouteGraphState struct {
	Epoch		uint64
	MessageEdges	[]AetherMeshMessageEdge
	RouteEdges	[]AetherMeshRouteEdge
	RootHash	string
}

type AetherMeshCongestionSnapshot struct {
	Epoch		uint64
	Height		uint64
	RoutingRoot	string
	SnapshotRoot	string
}

type AetherMeshRoutingEpochState struct {
	Epoch			uint64
	StartsHeight		uint64
	ExpiresHeight		uint64
	ZoneGraph		AetherMeshZoneGraphState
	ServiceGraph		AetherMeshServiceGraphState
	MessageRouteGraph	AetherMeshMessageRouteGraphState
	CongestionSnapshot	AetherMeshCongestionSnapshot
	CostParams		AetherMeshRoutingCostParams
	CostParamsHash		string
	RoutingTableRoot	string
	CongestionSnapshotRoot	string
	EpochRoot		string
}

type AetherMeshRouteProofRequest struct {
	SourceZone	string
	DestinationZone	string
	Sender		string
	Recipient	string
	Opcode		string
	MaxHops		uint32
	CurrentHeight	uint64
}

type AetherMeshRouteProof struct {
	Epoch			uint64
	QueryHash		string
	SelectedRoute		AetherMeshSelectedRouteMetadata
	CandidateRoute		AetherMeshRouteCandidate
	HopEdgeHashes		[]string
	RoutingTableRoot	string
	CongestionSnapshotRoot	string
	EpochRoot		string
	ProofHash		string
}

type AetherMeshRoutingSimulation struct {
	Epoch		uint64
	Request		AetherMeshRouteSelectionRequest
	Candidates	[]AetherMeshRouteCandidate
	Selected	AetherMeshSelectedRouteMetadata
	Proof		AetherMeshRouteProof
}

func BuildAetherMeshRoutingEpochState(epoch, startsHeight, expiresHeight uint64, zoneEdges []AetherMeshZoneEdge, serviceEdges []AetherMeshServiceEdge, messageEdges []AetherMeshMessageEdge, routeEdges []AetherMeshRouteEdge, params AetherMeshRoutingCostParams) (AetherMeshRoutingEpochState, error) {
	if epoch == 0 {
		return AetherMeshRoutingEpochState{}, errors.New("aether mesh routing epoch must be positive")
	}
	if startsHeight == 0 || expiresHeight <= startsHeight {
		return AetherMeshRoutingEpochState{}, errors.New("aether mesh routing epoch height range is invalid")
	}
	zoneState, err := BuildAetherMeshZoneGraphState(epoch, zoneEdges)
	if err != nil {
		return AetherMeshRoutingEpochState{}, err
	}
	serviceState, err := BuildAetherMeshServiceGraphState(epoch, serviceEdges)
	if err != nil {
		return AetherMeshRoutingEpochState{}, err
	}
	messageState, err := BuildAetherMeshMessageRouteGraphState(epoch, messageEdges, routeEdges)
	if err != nil {
		return AetherMeshRoutingEpochState{}, err
	}
	params = normalizeAetherMeshRoutingCostParams(params)
	if err := params.Validate(); err != nil {
		return AetherMeshRoutingEpochState{}, err
	}
	routingRoot := ComputeAetherMeshCommittedRoutingTableRoot(routeEdges)
	snapshotRoot := ComputeAetherMeshCommittedCongestionSnapshotRoot(routeEdges)
	state := AetherMeshRoutingEpochState{
		Epoch:			epoch,
		StartsHeight:		startsHeight,
		ExpiresHeight:		expiresHeight,
		ZoneGraph:		zoneState,
		ServiceGraph:		serviceState,
		MessageRouteGraph:	messageState,
		CongestionSnapshot:	BuildAetherMeshCongestionSnapshot(epoch, startsHeight, routeEdges),
		CostParams:		params,
		CostParamsHash:		ComputeAetherMeshRoutingCostParamsHash(params),
		RoutingTableRoot:	routingRoot,
		CongestionSnapshotRoot:	snapshotRoot,
	}
	state.EpochRoot = ComputeAetherMeshRoutingEpochRoot(state)
	return state, state.Validate()
}

func BuildAetherMeshZoneGraphState(epoch uint64, edges []AetherMeshZoneEdge) (AetherMeshZoneGraphState, error) {
	state := AetherMeshZoneGraphState{
		Epoch:		epoch,
		Edges:		normalizeAetherMeshZoneEdges(edges),
		RootHash:	ComputeAetherMeshZoneGraphRoot(epoch, edges),
	}
	return state, state.Validate()
}

func BuildAetherMeshServiceGraphState(epoch uint64, edges []AetherMeshServiceEdge) (AetherMeshServiceGraphState, error) {
	state := AetherMeshServiceGraphState{
		Epoch:		epoch,
		Edges:		normalizeAetherMeshServiceEdges(edges),
		RootHash:	ComputeAetherMeshServiceGraphRoot(epoch, edges),
	}
	return state, state.Validate()
}

func BuildAetherMeshMessageRouteGraphState(epoch uint64, messageEdges []AetherMeshMessageEdge, routeEdges []AetherMeshRouteEdge) (AetherMeshMessageRouteGraphState, error) {
	state := AetherMeshMessageRouteGraphState{
		Epoch:		epoch,
		MessageEdges:	normalizeAetherMeshMessageEdges(messageEdges),
		RouteEdges:	normalizeAetherMeshRouteEdges(routeEdges),
		RootHash:	ComputeAetherMeshMessageRouteGraphRoot(epoch, messageEdges, routeEdges),
	}
	return state, state.Validate()
}

func BuildAetherMeshCongestionSnapshot(epoch, height uint64, routeEdges []AetherMeshRouteEdge) AetherMeshCongestionSnapshot {
	return AetherMeshCongestionSnapshot{
		Epoch:		epoch,
		Height:		height,
		RoutingRoot:	ComputeAetherMeshCommittedRoutingTableRoot(routeEdges),
		SnapshotRoot:	ComputeAetherMeshCommittedCongestionSnapshotRoot(routeEdges),
	}
}

func QueryAetherMeshRouteProof(state AetherMeshRoutingEpochState, req AetherMeshRouteProofRequest) (AetherMeshRouteProof, error) {
	state = normalizeAetherMeshRoutingEpochState(state)
	if err := state.Validate(); err != nil {
		return AetherMeshRouteProof{}, err
	}
	req = normalizeAetherMeshRouteProofRequest(req)
	if err := req.Validate(); err != nil {
		return AetherMeshRouteProof{}, err
	}
	if req.CurrentHeight < state.StartsHeight || req.CurrentHeight >= state.ExpiresHeight {
		return AetherMeshRouteProof{}, errors.New("aether mesh route proof request height outside routing epoch")
	}
	selectionReq := AetherMeshRouteSelectionRequest{
		SourceZone:		req.SourceZone,
		DestinationZone:	req.DestinationZone,
		Sender:			req.Sender,
		Recipient:		req.Recipient,
		Opcode:			req.Opcode,
		RoutingTableRoot:	state.RoutingTableRoot,
		CongestionSnapshotRoot:	state.CongestionSnapshotRoot,
		MaxHops:		req.MaxHops,
		CurrentHeight:		req.CurrentHeight,
		CostParams:		state.CostParams,
	}
	selected, err := SelectAetherMeshRoute(selectionReq, state.MessageRouteGraph.RouteEdges)
	if err != nil {
		return AetherMeshRouteProof{}, err
	}
	candidates, err := BuildAetherMeshRouteCandidates(selectionReq, state.MessageRouteGraph.RouteEdges)
	if err != nil {
		return AetherMeshRouteProof{}, err
	}
	var candidate AetherMeshRouteCandidate
	for _, item := range candidates {
		if item.RouteHash == selected.RouteHash {
			candidate = item
			break
		}
	}
	if candidate.RouteHash == "" {
		return AetherMeshRouteProof{}, errors.New("aether mesh selected route proof candidate missing")
	}
	hopHashes := make([]string, 0, len(candidate.Hops))
	for _, hop := range candidate.Hops {
		hopHashes = append(hopHashes, hop.EdgeHash)
	}
	proof := AetherMeshRouteProof{
		Epoch:			state.Epoch,
		QueryHash:		ComputeAetherMeshRouteProofQueryHash(req),
		SelectedRoute:		selected,
		CandidateRoute:		candidate,
		HopEdgeHashes:		hopHashes,
		RoutingTableRoot:	state.RoutingTableRoot,
		CongestionSnapshotRoot:	state.CongestionSnapshotRoot,
		EpochRoot:		state.EpochRoot,
	}
	proof.ProofHash = ComputeAetherMeshRouteProofHash(proof)
	return proof, proof.Validate()
}

func SimulateAetherMeshRouting(state AetherMeshRoutingEpochState, req AetherMeshRouteProofRequest) (AetherMeshRoutingSimulation, error) {
	proof, err := QueryAetherMeshRouteProof(state, req)
	if err != nil {
		return AetherMeshRoutingSimulation{}, err
	}
	selectionReq := AetherMeshRouteSelectionRequest{
		SourceZone:		req.SourceZone,
		DestinationZone:	req.DestinationZone,
		Sender:			req.Sender,
		Recipient:		req.Recipient,
		Opcode:			req.Opcode,
		RoutingTableRoot:	state.RoutingTableRoot,
		CongestionSnapshotRoot:	state.CongestionSnapshotRoot,
		MaxHops:		req.MaxHops,
		CurrentHeight:		req.CurrentHeight,
		CostParams:		state.CostParams,
	}
	candidates, err := BuildAetherMeshRouteCandidates(selectionReq, state.MessageRouteGraph.RouteEdges)
	if err != nil {
		return AetherMeshRoutingSimulation{}, err
	}
	return AetherMeshRoutingSimulation{
		Epoch:		state.Epoch,
		Request:	normalizeAetherMeshRouteSelectionRequest(selectionReq),
		Candidates:	candidates,
		Selected:	proof.SelectedRoute,
		Proof:		proof,
	}, nil
}

func (state AetherMeshZoneGraphState) Validate() error {
	state = normalizeAetherMeshZoneGraphState(state)
	if state.Epoch == 0 {
		return errors.New("aether mesh zone graph epoch must be positive")
	}
	for _, edge := range state.Edges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("aether mesh zone graph state root", state.RootHash); err != nil {
		return err
	}
	if state.RootHash != ComputeAetherMeshZoneGraphRoot(state.Epoch, state.Edges) {
		return errors.New("aether mesh zone graph state root mismatch")
	}
	return nil
}

func (state AetherMeshServiceGraphState) Validate() error {
	state = normalizeAetherMeshServiceGraphState(state)
	if state.Epoch == 0 {
		return errors.New("aether mesh service graph epoch must be positive")
	}
	for _, edge := range state.Edges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("aether mesh service graph state root", state.RootHash); err != nil {
		return err
	}
	if state.RootHash != ComputeAetherMeshServiceGraphRoot(state.Epoch, state.Edges) {
		return errors.New("aether mesh service graph state root mismatch")
	}
	return nil
}

func (state AetherMeshMessageRouteGraphState) Validate() error {
	state = normalizeAetherMeshMessageRouteGraphState(state)
	if state.Epoch == 0 {
		return errors.New("aether mesh message route graph epoch must be positive")
	}
	for _, edge := range state.MessageEdges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	for _, edge := range state.RouteEdges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("aether mesh message route graph state root", state.RootHash); err != nil {
		return err
	}
	if state.RootHash != ComputeAetherMeshMessageRouteGraphRoot(state.Epoch, state.MessageEdges, state.RouteEdges) {
		return errors.New("aether mesh message route graph state root mismatch")
	}
	return nil
}

func (snapshot AetherMeshCongestionSnapshot) Validate(routeEdges []AetherMeshRouteEdge) error {
	snapshot = normalizeAetherMeshCongestionSnapshot(snapshot)
	if snapshot.Epoch == 0 || snapshot.Height == 0 {
		return errors.New("aether mesh congestion snapshot epoch and height must be positive")
	}
	if snapshot.RoutingRoot != ComputeAetherMeshCommittedRoutingTableRoot(routeEdges) {
		return errors.New("aether mesh congestion snapshot routing root mismatch")
	}
	if snapshot.SnapshotRoot != ComputeAetherMeshCommittedCongestionSnapshotRoot(routeEdges) {
		return errors.New("aether mesh congestion snapshot root mismatch")
	}
	return nil
}

func (state AetherMeshRoutingEpochState) Validate() error {
	state = normalizeAetherMeshRoutingEpochState(state)
	if state.Epoch == 0 {
		return errors.New("aether mesh routing epoch must be positive")
	}
	if state.StartsHeight == 0 || state.ExpiresHeight <= state.StartsHeight {
		return errors.New("aether mesh routing epoch height range is invalid")
	}
	if state.ZoneGraph.Epoch != state.Epoch || state.ServiceGraph.Epoch != state.Epoch || state.MessageRouteGraph.Epoch != state.Epoch || state.CongestionSnapshot.Epoch != state.Epoch {
		return errors.New("aether mesh routing epoch graph epochs must match")
	}
	if err := state.ZoneGraph.Validate(); err != nil {
		return err
	}
	if err := state.ServiceGraph.Validate(); err != nil {
		return err
	}
	if err := state.MessageRouteGraph.Validate(); err != nil {
		return err
	}
	if err := state.CongestionSnapshot.Validate(state.MessageRouteGraph.RouteEdges); err != nil {
		return err
	}
	if err := state.CostParams.Validate(); err != nil {
		return err
	}
	if state.CostParamsHash != ComputeAetherMeshRoutingCostParamsHash(state.CostParams) {
		return errors.New("aether mesh routing epoch cost params hash mismatch")
	}
	if state.RoutingTableRoot != ComputeAetherMeshCommittedRoutingTableRoot(state.MessageRouteGraph.RouteEdges) {
		return errors.New("aether mesh routing epoch table root mismatch")
	}
	if state.CongestionSnapshotRoot != ComputeAetherMeshCommittedCongestionSnapshotRoot(state.MessageRouteGraph.RouteEdges) {
		return errors.New("aether mesh routing epoch congestion snapshot root mismatch")
	}
	if err := ValidateHash("aether mesh routing epoch root", state.EpochRoot); err != nil {
		return err
	}
	if state.EpochRoot != ComputeAetherMeshRoutingEpochRoot(state) {
		return errors.New("aether mesh routing epoch root mismatch")
	}
	return nil
}

func (req AetherMeshRouteProofRequest) Validate() error {
	req = normalizeAetherMeshRouteProofRequest(req)
	if err := validateIdentifierSet("aether mesh route proof source zone", []string{req.SourceZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("aether mesh route proof destination zone", []string{req.DestinationZone}, MaxZoneIDBytes); err != nil {
		return err
	}
	if req.SourceZone == req.DestinationZone {
		return errors.New("aether mesh route proof endpoints must differ")
	}
	if err := validateMeshGraphToken("aether mesh route proof sender", req.Sender); err != nil {
		return err
	}
	if err := validateMeshGraphToken("aether mesh route proof recipient", req.Recipient); err != nil {
		return err
	}
	if err := validateMeshGraphToken("aether mesh route proof opcode", req.Opcode); err != nil {
		return err
	}
	if req.MaxHops == 0 || req.MaxHops > MaxAetherMeshRouteHops {
		return fmt.Errorf("aether mesh route proof max hops must be between 1 and %d", MaxAetherMeshRouteHops)
	}
	if req.CurrentHeight == 0 {
		return errors.New("aether mesh route proof current height must be positive")
	}
	return nil
}

func (proof AetherMeshRouteProof) Validate() error {
	proof = normalizeAetherMeshRouteProof(proof)
	if proof.Epoch == 0 {
		return errors.New("aether mesh route proof epoch must be positive")
	}
	if err := ValidateHash("aether mesh route proof query hash", proof.QueryHash); err != nil {
		return err
	}
	if err := proof.SelectedRoute.Validate(); err != nil {
		return err
	}
	if proof.SelectedRoute.RouteHash != proof.CandidateRoute.RouteHash {
		return errors.New("aether mesh route proof candidate mismatch")
	}
	if len(proof.HopEdgeHashes) != len(proof.CandidateRoute.Hops) {
		return errors.New("aether mesh route proof hop hash count mismatch")
	}
	for i, hop := range proof.CandidateRoute.Hops {
		if proof.HopEdgeHashes[i] != hop.EdgeHash {
			return errors.New("aether mesh route proof hop hash mismatch")
		}
	}
	if err := ValidateHash("aether mesh route proof routing root", proof.RoutingTableRoot); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh route proof congestion root", proof.CongestionSnapshotRoot); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh route proof epoch root", proof.EpochRoot); err != nil {
		return err
	}
	if err := ValidateHash("aether mesh route proof hash", proof.ProofHash); err != nil {
		return err
	}
	if proof.ProofHash != ComputeAetherMeshRouteProofHash(proof) {
		return errors.New("aether mesh route proof hash mismatch")
	}
	return nil
}

func ComputeAetherMeshMessageRouteGraphRoot(epoch uint64, messageEdges []AetherMeshMessageEdge, routeEdges []AetherMeshRouteEdge) string {
	messageRoot := ComputeAetherMeshMessageGraphRoot(epoch, messageEdges)
	routingRoot := ComputeAetherMeshCommittedRoutingTableRoot(routeEdges)
	congestionRoot := ComputeAetherMeshCommittedCongestionSnapshotRoot(routeEdges)
	return HashParts("aether-mesh-message-route-graph", fmt.Sprintf("%020d", epoch), messageRoot, routingRoot, congestionRoot)
}

func ComputeAetherMeshRoutingCostParamsHash(params AetherMeshRoutingCostParams) string {
	params = normalizeAetherMeshRoutingCostParams(params)
	return HashParts("aether-mesh-routing-cost-params", fmt.Sprintf("%020d", params.BaseHopCost), fmt.Sprintf("%020d", params.GasCostWeight), fmt.Sprintf("%020d", params.CongestionWeight), fmt.Sprintf("%020d", params.ReliabilityWeight), fmt.Sprintf("%020d", params.LatencyWeight))
}

func ComputeAetherMeshRoutingEpochRoot(state AetherMeshRoutingEpochState) string {
	state = normalizeAetherMeshRoutingEpochState(state)
	return HashParts("aether-mesh-routing-epoch", fmt.Sprintf("%020d", state.Epoch), fmt.Sprintf("%020d", state.StartsHeight), fmt.Sprintf("%020d", state.ExpiresHeight), state.ZoneGraph.RootHash, state.ServiceGraph.RootHash, state.MessageRouteGraph.RootHash, state.CostParamsHash, state.RoutingTableRoot, state.CongestionSnapshotRoot)
}

func ComputeAetherMeshRouteProofQueryHash(req AetherMeshRouteProofRequest) string {
	req = normalizeAetherMeshRouteProofRequest(req)
	return HashParts("aether-mesh-route-proof-query", req.SourceZone, req.DestinationZone, req.Sender, req.Recipient, req.Opcode, fmt.Sprintf("%010d", req.MaxHops), fmt.Sprintf("%020d", req.CurrentHeight))
}

func ComputeAetherMeshRouteProofHash(proof AetherMeshRouteProof) string {
	proof = normalizeAetherMeshRouteProof(proof)
	parts := []string{"aether-mesh-route-proof", fmt.Sprintf("%020d", proof.Epoch), proof.QueryHash, proof.SelectedRoute.MetadataHash, proof.CandidateRoute.RouteHash, proof.RoutingTableRoot, proof.CongestionSnapshotRoot, proof.EpochRoot}
	for _, hopHash := range proof.HopEdgeHashes {
		parts = append(parts, hopHash)
	}
	return HashParts(parts...)
}

func normalizeAetherMeshRoutingEpochState(state AetherMeshRoutingEpochState) AetherMeshRoutingEpochState {
	state.ZoneGraph = normalizeAetherMeshZoneGraphState(state.ZoneGraph)
	state.ServiceGraph = normalizeAetherMeshServiceGraphState(state.ServiceGraph)
	state.MessageRouteGraph = normalizeAetherMeshMessageRouteGraphState(state.MessageRouteGraph)
	state.CongestionSnapshot = normalizeAetherMeshCongestionSnapshot(state.CongestionSnapshot)
	state.CostParams = normalizeAetherMeshRoutingCostParams(state.CostParams)
	state.CostParamsHash = normalizeHashText(state.CostParamsHash)
	state.RoutingTableRoot = normalizeHashText(state.RoutingTableRoot)
	state.CongestionSnapshotRoot = normalizeHashText(state.CongestionSnapshotRoot)
	state.EpochRoot = normalizeHashText(state.EpochRoot)
	return state
}

func normalizeAetherMeshZoneGraphState(state AetherMeshZoneGraphState) AetherMeshZoneGraphState {
	state.Edges = normalizeAetherMeshZoneEdges(state.Edges)
	state.RootHash = normalizeHashText(state.RootHash)
	return state
}

func normalizeAetherMeshServiceGraphState(state AetherMeshServiceGraphState) AetherMeshServiceGraphState {
	state.Edges = normalizeAetherMeshServiceEdges(state.Edges)
	state.RootHash = normalizeHashText(state.RootHash)
	return state
}

func normalizeAetherMeshMessageRouteGraphState(state AetherMeshMessageRouteGraphState) AetherMeshMessageRouteGraphState {
	state.MessageEdges = normalizeAetherMeshMessageEdges(state.MessageEdges)
	state.RouteEdges = normalizeAetherMeshRouteEdges(state.RouteEdges)
	state.RootHash = normalizeHashText(state.RootHash)
	return state
}

func normalizeAetherMeshCongestionSnapshot(snapshot AetherMeshCongestionSnapshot) AetherMeshCongestionSnapshot {
	snapshot.RoutingRoot = normalizeHashText(snapshot.RoutingRoot)
	snapshot.SnapshotRoot = normalizeHashText(snapshot.SnapshotRoot)
	return snapshot
}

func normalizeAetherMeshRouteProofRequest(req AetherMeshRouteProofRequest) AetherMeshRouteProofRequest {
	req.SourceZone = normalizeMeshRouteProofToken(req.SourceZone)
	req.DestinationZone = normalizeMeshRouteProofToken(req.DestinationZone)
	req.Sender = normalizeMeshRouteProofToken(req.Sender)
	req.Recipient = normalizeMeshRouteProofToken(req.Recipient)
	req.Opcode = normalizeMeshRouteProofToken(req.Opcode)
	return req
}

func normalizeAetherMeshRouteProof(proof AetherMeshRouteProof) AetherMeshRouteProof {
	proof.QueryHash = normalizeHashText(proof.QueryHash)
	proof.SelectedRoute = normalizeAetherMeshSelectedRouteMetadata(proof.SelectedRoute)
	proof.CandidateRoute = normalizeAetherMeshRouteCandidate(proof.CandidateRoute)
	proof.HopEdgeHashes = append([]string(nil), proof.HopEdgeHashes...)
	for i, hash := range proof.HopEdgeHashes {
		proof.HopEdgeHashes[i] = normalizeHashText(hash)
	}
	proof.RoutingTableRoot = normalizeHashText(proof.RoutingTableRoot)
	proof.CongestionSnapshotRoot = normalizeHashText(proof.CongestionSnapshotRoot)
	proof.EpochRoot = normalizeHashText(proof.EpochRoot)
	proof.ProofHash = normalizeHashText(proof.ProofHash)
	return proof
}

func normalizeMeshRouteProofToken(value string) string {
	return strings.TrimSpace(value)
}
