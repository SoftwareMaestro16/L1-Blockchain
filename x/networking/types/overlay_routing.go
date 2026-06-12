package types

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type OverlayMembershipMode string

const (
	OverlayMembershipModeDeterministicRule	OverlayMembershipMode	= "DETERMINISTIC_RULE"
	OverlayMembershipModeCryptographicAuth	OverlayMembershipMode	= "CRYPTOGRAPHIC_AUTHORIZATION"
	OverlayMembershipModeStakeBased		OverlayMembershipMode	= "STAKE_BASED_INCLUSION"
	OverlayMembershipModeServiceRegistry	OverlayMembershipMode	= "SERVICE_REGISTRY_MEMBERSHIP"
	OverlayMembershipModeZoneAssignment	OverlayMembershipMode	= "ZONE_ASSIGNMENT"
	OverlayMembershipModeDynamicRouting	OverlayMembershipMode	= "DYNAMIC_ROUTING_ASSIGNMENT"
)

type MembershipProofType string

const (
	MembershipProofValidatorSet		MembershipProofType	= "VALIDATOR_SET_PROOF"
	MembershipProofZoneAssignment		MembershipProofType	= "ZONE_ASSIGNMENT_PROOF"
	MembershipProofServiceRegistration	MembershipProofType	= "SERVICE_REGISTRATION_PROOF"
	MembershipProofProviderStake		MembershipProofType	= "PROVIDER_STAKE_PROOF"
	MembershipProofSignedAuthorization	MembershipProofType	= "SIGNED_AUTHORIZATION_RECORD"
)

type OverlayMembershipProof struct {
	ProofID		string
	OverlayID	string
	NodeID		string
	ProofType	MembershipProofType
	Mode		OverlayMembershipMode
	Membership	OverlayMembershipRule
	ProofHash	string
	AuthorityHash	string
	ZoneID		string
	ServiceID	string
	StakeAmount	uint64
	Committed	bool
	Deterministic	bool
	ExpiresHeight	uint64
	AuthPubKey	[]byte
	Signature	[]byte
}

type OverlayMembershipRecord struct {
	OverlayID	string
	NodeID		string
	ProofID		string
	Membership	OverlayMembershipRule
	Mode		OverlayMembershipMode
	JoinedHeight	uint64
	ExpiresHeight	uint64
}

type RouteHint struct {
	ZoneID			string
	ShardID			string
	ServiceID		string
	StorageKeyHash		string
	DeterministicHintHash	string
}

type RoutingGraph struct {
	OverlayID		string
	GraphHash		string
	Version			uint64
	Committed		bool
	DeterministicHintHash	string
	Edges			[]RoutingEdge
}

type RoutingEdge struct {
	FromNodeID	string
	ToNodeID	string
	LatencyMillis	uint64
	Weight		uint32
	Priority	uint32
	ZoneID		string
}

type OverlayRoutingRequest struct {
	Message			NetworkMessage
	SourceNodeID		string
	CandidatePeers		[]NodeRecord
	MembershipProofs	[]OverlayMembershipProof
	Graph			RoutingGraph
	Hint			RouteHint
	CurrentHeight		uint64
}

type OverlayRoutePlan struct {
	MessageID			string
	OverlayID			string
	OverlayType			OverlayType
	Strategy			RoutingStrategy
	TargetNodeIDs			[]string
	UsesCommittedRoutingTable	bool
	UsesDeterministicHint		bool
	UsesNodeLocalAdaptation		bool
	FallbackUsed			bool
}

func NewOverlayMembershipProof(proof OverlayMembershipProof) (OverlayMembershipProof, error) {
	proof = NormalizeOverlayMembershipProof(proof)
	if proof.ProofID == "" {
		proof.ProofID = ComputeOverlayMembershipProofID(proof)
	}
	if err := proof.ValidateBasic(0); err != nil {
		return OverlayMembershipProof{}, err
	}
	return proof, nil
}

func SignOverlayMembershipProof(proof OverlayMembershipProof, privateKey ed25519.PrivateKey) (OverlayMembershipProof, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return OverlayMembershipProof{}, errors.New("networking overlay membership private key must be ed25519")
	}
	pubKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return OverlayMembershipProof{}, errors.New("networking overlay membership public key must be ed25519")
	}
	proof.AuthPubKey = cloneBytes(pubKey)
	proof.Signature = nil
	proof = NormalizeOverlayMembershipProof(proof)
	proof.ProofID = ComputeOverlayMembershipProofID(proof)
	payload, err := proof.SigningPayload()
	if err != nil {
		return OverlayMembershipProof{}, err
	}
	proof.Signature = ed25519.Sign(privateKey, payload)
	if err := proof.ValidateBasic(0); err != nil {
		return OverlayMembershipProof{}, err
	}
	return proof, nil
}

func NormalizeOverlayMembershipProof(proof OverlayMembershipProof) OverlayMembershipProof {
	proof.ProofID = normalizeHashText(proof.ProofID)
	proof.OverlayID = normalizeHashText(proof.OverlayID)
	proof.NodeID = normalizeHashText(proof.NodeID)
	proof.ProofType = MembershipProofType(strings.ToUpper(strings.TrimSpace(string(proof.ProofType))))
	proof.Mode = OverlayMembershipMode(strings.ToUpper(strings.TrimSpace(string(proof.Mode))))
	proof.Membership = OverlayMembershipRule(strings.ToUpper(strings.TrimSpace(string(proof.Membership))))
	proof.ProofHash = normalizeHashText(proof.ProofHash)
	proof.AuthorityHash = normalizeHashText(proof.AuthorityHash)
	proof.ZoneID = strings.TrimSpace(proof.ZoneID)
	proof.ServiceID = strings.TrimSpace(proof.ServiceID)
	proof.AuthPubKey = cloneBytes(proof.AuthPubKey)
	proof.Signature = cloneBytes(proof.Signature)
	return proof
}

func ComputeOverlayMembershipProofID(proof OverlayMembershipProof) string {
	proof = NormalizeOverlayMembershipProof(proof)
	return HashParts(
		"overlay-membership-proof",
		proof.OverlayID,
		proof.NodeID,
		string(proof.ProofType),
		string(proof.Mode),
		string(proof.Membership),
		proof.ProofHash,
		proof.AuthorityHash,
		proof.ZoneID,
		proof.ServiceID,
		fmt.Sprintf("%d", proof.StakeAmount),
		fmt.Sprintf("%t", proof.Committed),
		fmt.Sprintf("%t", proof.Deterministic),
		fmt.Sprintf("%d", proof.ExpiresHeight),
		hashBytes("aetra-overlay-membership-auth-pub-key-v1", proof.AuthPubKey),
	)
}

func (p OverlayMembershipProof) SigningPayload() ([]byte, error) {
	proof := NormalizeOverlayMembershipProof(p)
	proof.Signature = nil
	return json.Marshal(proof)
}

func (p OverlayMembershipProof) ValidateBasic(currentHeight uint64) error {
	proof := NormalizeOverlayMembershipProof(p)
	if err := ValidateHash("networking overlay membership proof id", proof.ProofID); err != nil {
		return err
	}
	if proof.ProofID != ComputeOverlayMembershipProofID(proof) {
		return errors.New("networking overlay membership proof id does not match payload")
	}
	if err := ValidateHash("networking overlay membership overlay id", proof.OverlayID); err != nil {
		return err
	}
	if err := ValidateHash("networking overlay membership node id", proof.NodeID); err != nil {
		return err
	}
	if !IsMembershipProofType(proof.ProofType) {
		return fmt.Errorf("unknown networking overlay membership proof type %q", proof.ProofType)
	}
	if !IsOverlayMembershipMode(proof.Mode) {
		return fmt.Errorf("unknown networking overlay membership mode %q", proof.Mode)
	}
	if !IsOverlayMembershipRule(proof.Membership) {
		return fmt.Errorf("unknown networking overlay membership rule %q", proof.Membership)
	}
	if err := ValidateHash("networking overlay membership proof hash", proof.ProofHash); err != nil {
		return err
	}
	if err := ValidateHash("networking overlay membership authority hash", proof.AuthorityHash); err != nil {
		return err
	}
	if proof.ExpiresHeight == 0 {
		return errors.New("networking overlay membership proof expires height must be positive")
	}
	if currentHeight > 0 && currentHeight > proof.ExpiresHeight {
		return errors.New("networking overlay membership proof is expired")
	}
	if proof.ProofType == MembershipProofSignedAuthorization {
		if len(proof.AuthPubKey) != ed25519.PublicKeySize {
			return fmt.Errorf("networking overlay membership auth pub key must be %d bytes", ed25519.PublicKeySize)
		}
		if len(proof.Signature) != ed25519.SignatureSize {
			return fmt.Errorf("networking overlay membership signature must be %d bytes", ed25519.SignatureSize)
		}
		payload, err := proof.SigningPayload()
		if err != nil {
			return err
		}
		if !ed25519.Verify(proof.AuthPubKey, payload, proof.Signature) {
			return errors.New("networking overlay membership signature verification failed")
		}
	}
	if proof.ProofType == MembershipProofProviderStake && (!proof.Committed || proof.StakeAmount == 0) {
		return errors.New("networking provider stake proof must be committed and positive")
	}
	if proof.ProofType == MembershipProofValidatorSet && (!proof.Committed || !proof.Deterministic) {
		return errors.New("networking validator set proof must be committed and deterministic")
	}
	return nil
}

func AuthorizeOverlayMembership(record NodeRecord, desc OverlayDescriptor, proof OverlayMembershipProof, currentHeight uint64) (OverlayMembershipRecord, error) {
	record = NormalizeNodeRecord(record)
	desc = NormalizeOverlayDescriptor(desc)
	proof = NormalizeOverlayMembershipProof(proof)
	if err := record.ValidateBasic(); err != nil {
		return OverlayMembershipRecord{}, err
	}
	if err := desc.ValidateBasic(); err != nil {
		return OverlayMembershipRecord{}, err
	}
	if err := proof.ValidateBasic(currentHeight); err != nil {
		return OverlayMembershipRecord{}, err
	}
	if proof.OverlayID != desc.OverlayID || proof.NodeID != record.NodeID {
		return OverlayMembershipRecord{}, errors.New("networking overlay membership proof subject mismatch")
	}
	if proof.Membership != desc.Membership {
		return OverlayMembershipRecord{}, errors.New("networking overlay membership proof rule mismatch")
	}
	if !proofTypeMatchesMembership(proof.ProofType, desc.Membership) {
		return OverlayMembershipRecord{}, errors.New("networking overlay membership proof type does not satisfy membership rule")
	}
	matches, err := NodeSatisfiesOverlayMembership(record, desc)
	if err != nil {
		return OverlayMembershipRecord{}, err
	}
	if !matches {
		return OverlayMembershipRecord{}, errors.New("networking node record does not satisfy overlay membership rule")
	}
	if proof.ProofType == MembershipProofZoneAssignment && !containsString(record.ZonesSupported, proof.ZoneID) {
		return OverlayMembershipRecord{}, errors.New("networking zone assignment proof must match advertised zone")
	}
	if proof.ProofType == MembershipProofServiceRegistration && !containsString(record.ServicesSupported, proof.ServiceID) {
		return OverlayMembershipRecord{}, errors.New("networking service registration proof must match advertised service")
	}
	if proof.ExpiresHeight > record.ExpiresHeight {
		return OverlayMembershipRecord{}, errors.New("networking overlay membership proof cannot outlive node record")
	}
	return OverlayMembershipRecord{
		OverlayID:	desc.OverlayID,
		NodeID:		record.NodeID,
		ProofID:	proof.ProofID,
		Membership:	desc.Membership,
		Mode:		proof.Mode,
		JoinedHeight:	currentHeight,
		ExpiresHeight:	proof.ExpiresHeight,
	}, nil
}

func BuildOverlayRoute(req OverlayRoutingRequest, descriptors []OverlayDescriptor) (OverlayRoutePlan, error) {
	msg := req.Message.Normalize()
	if msg.ReplaySafeID == "" {
		msg.ReplaySafeID = ComputeNetworkMessageID(msg)
	}
	if err := msg.ValidateHardRules(); err != nil {
		return OverlayRoutePlan{}, err
	}
	sourceNodeID := normalizeHashText(req.SourceNodeID)
	if err := ValidateHash("networking overlay route source node id", sourceNodeID); err != nil {
		return OverlayRoutePlan{}, err
	}
	desc, found, err := ClassifyOverlayForMessage(msg, descriptors)
	if err != nil {
		return OverlayRoutePlan{}, err
	}
	if !found {
		return OverlayRoutePlan{}, errors.New("networking no overlay descriptor for message")
	}
	graph := NormalizeRoutingGraph(req.Graph)
	if graph.OverlayID == "" {
		graph.OverlayID = desc.OverlayID
	}
	if err := graph.Validate(desc); err != nil {
		return OverlayRoutePlan{}, err
	}
	if err := validateConsensusRouteSafety(msg, desc, graph, req.Hint); err != nil {
		return OverlayRoutePlan{}, err
	}
	proofByNode := make(map[string]OverlayMembershipProof, len(req.MembershipProofs))
	for _, proof := range req.MembershipProofs {
		proof = NormalizeOverlayMembershipProof(proof)
		proofByNode[proof.NodeID] = proof
	}
	eligible := make([]NodeRecord, 0, len(req.CandidatePeers))
	for _, peer := range req.CandidatePeers {
		peer = NormalizeNodeRecord(peer)
		proof, ok := proofByNode[peer.NodeID]
		if !ok {
			continue
		}
		if _, err := AuthorizeOverlayMembership(peer, desc, proof, req.CurrentHeight); err != nil {
			continue
		}
		eligible = append(eligible, peer)
	}
	fanout, err := PlanOverlayFanout(desc, uint32(len(eligible)))
	if err != nil {
		return OverlayRoutePlan{}, err
	}
	targets := selectOverlayTargets(desc, msg, graph, req.Hint, sourceNodeID, eligible)
	if len(targets) > int(fanout) {
		targets = targets[:fanout]
	}
	if len(targets) == 0 {
		return OverlayRoutePlan{}, errors.New("networking overlay route has no target peers")
	}
	return OverlayRoutePlan{
		MessageID:			msg.ReplaySafeID,
		OverlayID:			desc.OverlayID,
		OverlayType:			desc.OverlayType,
		Strategy:			desc.Routing,
		TargetNodeIDs:			targets,
		UsesCommittedRoutingTable:	graph.Committed,
		UsesDeterministicHint:		graph.DeterministicHintHash != "" || normalizeHashText(req.Hint.DeterministicHintHash) != "",
		UsesNodeLocalAdaptation:	strategyUsesNodeLocalAdaptation(desc.Routing) && !graph.Committed,
		FallbackUsed:			desc.Routing == RoutingStrategyProbabilisticGossip || len(graph.Edges) == 0,
	}, nil
}

func ClassifyOverlayForMessage(msg NetworkMessage, descriptors []OverlayDescriptor) (OverlayDescriptor, bool, error) {
	overlayType := overlayTypeForMessage(msg)
	for _, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if err := desc.ValidateBasic(); err != nil {
			return OverlayDescriptor{}, false, err
		}
		if desc.OverlayType == overlayType {
			return desc, true, nil
		}
	}
	return OverlayDescriptor{}, false, nil
}

func NormalizeRoutingGraph(graph RoutingGraph) RoutingGraph {
	graph.OverlayID = normalizeHashText(graph.OverlayID)
	graph.GraphHash = normalizeHashText(graph.GraphHash)
	graph.DeterministicHintHash = normalizeHashText(graph.DeterministicHintHash)
	for i := range graph.Edges {
		graph.Edges[i].FromNodeID = normalizeHashText(graph.Edges[i].FromNodeID)
		graph.Edges[i].ToNodeID = normalizeHashText(graph.Edges[i].ToNodeID)
		graph.Edges[i].ZoneID = strings.TrimSpace(graph.Edges[i].ZoneID)
	}
	sort.SliceStable(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].FromNodeID != graph.Edges[j].FromNodeID {
			return graph.Edges[i].FromNodeID < graph.Edges[j].FromNodeID
		}
		return graph.Edges[i].ToNodeID < graph.Edges[j].ToNodeID
	})
	if graph.GraphHash == "" {
		graph.GraphHash = ComputeRoutingGraphHash(graph)
	}
	return graph
}

func ComputeRoutingGraphHash(graph RoutingGraph) string {
	graph = NormalizeRoutingGraphNoHash(graph)
	parts := []string{
		"routing-graph",
		graph.OverlayID,
		fmt.Sprintf("%d", graph.Version),
		fmt.Sprintf("%t", graph.Committed),
		graph.DeterministicHintHash,
	}
	for _, edge := range graph.Edges {
		parts = append(parts,
			edge.FromNodeID,
			edge.ToNodeID,
			fmt.Sprintf("%d", edge.LatencyMillis),
			fmt.Sprintf("%d", edge.Weight),
			fmt.Sprintf("%d", edge.Priority),
			edge.ZoneID,
		)
	}
	return HashParts(parts...)
}

func NormalizeRoutingGraphNoHash(graph RoutingGraph) RoutingGraph {
	graph.GraphHash = ""
	graph.OverlayID = normalizeHashText(graph.OverlayID)
	graph.DeterministicHintHash = normalizeHashText(graph.DeterministicHintHash)
	for i := range graph.Edges {
		graph.Edges[i].FromNodeID = normalizeHashText(graph.Edges[i].FromNodeID)
		graph.Edges[i].ToNodeID = normalizeHashText(graph.Edges[i].ToNodeID)
		graph.Edges[i].ZoneID = strings.TrimSpace(graph.Edges[i].ZoneID)
	}
	sort.SliceStable(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].FromNodeID != graph.Edges[j].FromNodeID {
			return graph.Edges[i].FromNodeID < graph.Edges[j].FromNodeID
		}
		return graph.Edges[i].ToNodeID < graph.Edges[j].ToNodeID
	})
	return graph
}

func (g RoutingGraph) Validate(desc OverlayDescriptor) error {
	graph := NormalizeRoutingGraph(g)
	desc = NormalizeOverlayDescriptor(desc)
	if graph.OverlayID != desc.OverlayID {
		return errors.New("networking routing graph overlay mismatch")
	}
	if graph.Version == 0 {
		return errors.New("networking routing graph version must be positive")
	}
	if err := ValidateHash("networking routing graph hash", graph.GraphHash); err != nil {
		return err
	}
	if graph.GraphHash != ComputeRoutingGraphHash(graph) {
		return errors.New("networking routing graph hash does not match graph")
	}
	if graph.DeterministicHintHash != "" {
		if err := ValidateHash("networking routing deterministic hint hash", graph.DeterministicHintHash); err != nil {
			return err
		}
	}
	for _, edge := range graph.Edges {
		if err := ValidateHash("networking routing graph from node id", edge.FromNodeID); err != nil {
			return err
		}
		if err := ValidateHash("networking routing graph to node id", edge.ToNodeID); err != nil {
			return err
		}
	}
	return nil
}

func IsMembershipProofType(proofType MembershipProofType) bool {
	switch proofType {
	case MembershipProofValidatorSet, MembershipProofZoneAssignment, MembershipProofServiceRegistration, MembershipProofProviderStake, MembershipProofSignedAuthorization:
		return true
	default:
		return false
	}
}

func IsOverlayMembershipMode(mode OverlayMembershipMode) bool {
	switch mode {
	case OverlayMembershipModeDeterministicRule, OverlayMembershipModeCryptographicAuth, OverlayMembershipModeStakeBased, OverlayMembershipModeServiceRegistry, OverlayMembershipModeZoneAssignment, OverlayMembershipModeDynamicRouting:
		return true
	default:
		return false
	}
}

func overlayTypeForMessage(msg NetworkMessage) OverlayType {
	switch msg.Channel {
	case ChannelConsensus, ChannelBlock, ChannelMempool:
		return OverlayTypeValidator
	case ChannelExecution:
		return OverlayTypeExecution
	case ChannelData, ChannelStateSync:
		return OverlayTypeData
	case ChannelService:
		return OverlayTypeService
	case ChannelRouting:
		return OverlayTypeRouting
	case ChannelDiscovery:
		return OverlayTypeDiscovery
	default:
		return OverlayTypeDiscovery
	}
}

func proofTypeMatchesMembership(proofType MembershipProofType, membership OverlayMembershipRule) bool {
	switch membership {
	case OverlayMembershipValidatorSet:
		return proofType == MembershipProofValidatorSet
	case OverlayMembershipZoneSupported, OverlayMembershipExecutionRole:
		return proofType == MembershipProofZoneAssignment || proofType == MembershipProofSignedAuthorization
	case OverlayMembershipDataProvider, OverlayMembershipStorageProvider:
		return proofType == MembershipProofProviderStake
	case OverlayMembershipServiceAdvertisement:
		return proofType == MembershipProofServiceRegistration
	case OverlayMembershipSignedDiscovery, OverlayMembershipRoutingRole:
		return proofType == MembershipProofSignedAuthorization
	default:
		return false
	}
}

func validateConsensusRouteSafety(msg NetworkMessage, desc OverlayDescriptor, graph RoutingGraph, hint RouteHint) error {
	if !msg.ConsensusEffect {
		return nil
	}
	hintHash := normalizeHashText(hint.DeterministicHintHash)
	hasDeterministicHint := graph.DeterministicHintHash != "" || hintHash != ""
	if hintHash != "" {
		if err := ValidateHash("networking overlay route deterministic hint", hintHash); err != nil {
			return err
		}
	}
	if !graph.Committed && !hasDeterministicHint {
		return errors.New("networking consensus-effect routing requires committed routing table or deterministic route hint")
	}
	if strategyUsesNodeLocalAdaptation(desc.Routing) && !graph.Committed {
		return errors.New("networking node-local adaptive routing cannot determine consensus-effect delivery")
	}
	return nil
}

func selectOverlayTargets(desc OverlayDescriptor, msg NetworkMessage, graph RoutingGraph, hint RouteHint, sourceNodeID string, eligible []NodeRecord) []string {
	ids := make([]string, 0, len(eligible))
	for _, peer := range eligible {
		ids = append(ids, NormalizeNodeRecord(peer).NodeID)
	}
	sortStrings(ids)
	switch desc.Routing {
	case RoutingStrategyShortestLatencyPath, RoutingStrategyLowLatencyAdvisory:
		return sortByLatency(sourceNodeID, ids, graph)
	case RoutingStrategyZoneLocal:
		return sortByZone(hint.ZoneID, eligible)
	case RoutingStrategyDeterministicShard:
		return sortByHashSeed(ids, HashParts("deterministic-shard", desc.OverlayID, hint.ShardID, msg.ReplaySafeID))
	case RoutingStrategyPriorityBroadcastTree:
		return sortByPriority(sourceNodeID, ids, graph)
	case RoutingStrategyProbabilisticGossip, RoutingStrategyRandomWalkAdvisory:
		return sortByHashSeed(ids, HashParts("probabilistic-gossip", desc.OverlayID, msg.ReplaySafeID))
	case RoutingStrategyServiceProvider:
		return sortByService(hint.ServiceID, eligible)
	case RoutingStrategyStorageProvider:
		return sortByHashSeed(ids, HashParts("storage-provider", desc.OverlayID, hint.StorageKeyHash, msg.ReplaySafeID))
	default:
		return ids
	}
}

func strategyUsesNodeLocalAdaptation(strategy RoutingStrategy) bool {
	return strategy == RoutingStrategyShortestLatencyPath ||
		strategy == RoutingStrategyLowLatencyAdvisory ||
		strategy == RoutingStrategyRandomWalkAdvisory
}

func sortByLatency(sourceNodeID string, ids []string, graph RoutingGraph) []string {
	latency := make(map[string]uint64, len(graph.Edges))
	for _, edge := range graph.Edges {
		if edge.FromNodeID == sourceNodeID {
			latency[edge.ToNodeID] = edge.LatencyMillis
		}
	}
	out := append([]string(nil), ids...)
	sort.SliceStable(out, func(i, j int) bool {
		left, leftFound := latency[out[i]]
		right, rightFound := latency[out[j]]
		if leftFound != rightFound {
			return leftFound
		}
		if left != right {
			return left < right
		}
		return out[i] < out[j]
	})
	return out
}

func sortByPriority(sourceNodeID string, ids []string, graph RoutingGraph) []string {
	priority := make(map[string]uint32, len(graph.Edges))
	for _, edge := range graph.Edges {
		if edge.FromNodeID == sourceNodeID {
			priority[edge.ToNodeID] = edge.Priority
		}
	}
	out := append([]string(nil), ids...)
	sort.SliceStable(out, func(i, j int) bool {
		left, leftFound := priority[out[i]]
		right, rightFound := priority[out[j]]
		if leftFound != rightFound {
			return leftFound
		}
		if left != right {
			return left < right
		}
		return out[i] < out[j]
	})
	return out
}

func sortByZone(zoneID string, peers []NodeRecord) []string {
	out := make([]NodeRecord, len(peers))
	copy(out, peers)
	sort.SliceStable(out, func(i, j int) bool {
		left := containsString(out[i].ZonesSupported, zoneID)
		right := containsString(out[j].ZonesSupported, zoneID)
		if left != right {
			return left
		}
		return out[i].NodeID < out[j].NodeID
	})
	ids := make([]string, len(out))
	for i, peer := range out {
		ids[i] = NormalizeNodeRecord(peer).NodeID
	}
	return ids
}

func sortByService(serviceID string, peers []NodeRecord) []string {
	out := make([]NodeRecord, len(peers))
	copy(out, peers)
	sort.SliceStable(out, func(i, j int) bool {
		left := containsString(out[i].ServicesSupported, serviceID)
		right := containsString(out[j].ServicesSupported, serviceID)
		if left != right {
			return left
		}
		return out[i].NodeID < out[j].NodeID
	})
	ids := make([]string, len(out))
	for i, peer := range out {
		ids[i] = NormalizeNodeRecord(peer).NodeID
	}
	return ids
}

func sortByHashSeed(ids []string, seed string) []string {
	out := append([]string(nil), ids...)
	sort.SliceStable(out, func(i, j int) bool {
		left := HashParts(seed, out[i])
		right := HashParts(seed, out[j])
		if left != right {
			return left < right
		}
		return out[i] < out[j]
	})
	return out
}

func containsString(values []string, value string) bool {
	value = strings.TrimSpace(value)
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
