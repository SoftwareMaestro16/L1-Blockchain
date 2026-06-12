package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	DefaultPeerRotationLimitBps	= uint32(2_500)
	DefaultStalePeerEpochs		= uint64(4)
)

type OverlayMembershipManager struct {
	Descriptors	[]OverlayDescriptor
	Records		[]OverlayMembershipRecord
}

type PeerRotationPolicy struct {
	MaxRotatedPeersBps	uint32
	StaleAfterEpochs	uint64
}

type PeerSetManager struct {
	Descriptor	OverlayDescriptor
	Graph		AdaptiveOverlayGraph
	DecayPolicy	PeerScoreDecayPolicy
	RotationPolicy	PeerRotationPolicy
}

func NewOverlayMembershipManager(descriptors []OverlayDescriptor) (OverlayMembershipManager, error) {
	descriptors = cloneOverlayDescriptors(descriptors)
	sortOverlayDescriptors(descriptors)
	if err := ValidateOverlayDescriptors(descriptors, 0); err != nil {
		return OverlayMembershipManager{}, err
	}
	return OverlayMembershipManager{Descriptors: descriptors}, nil
}

func (m OverlayMembershipManager) Join(record NodeRecord, proof OverlayMembershipProof, currentHeight uint64) (OverlayMembershipManager, OverlayMembershipRecord, error) {
	desc, found := m.descriptorByID(proof.OverlayID)
	if !found {
		return OverlayMembershipManager{}, OverlayMembershipRecord{}, errors.New("networking overlay membership manager missing descriptor")
	}
	membership, err := AuthorizeOverlayMembership(record, desc, proof, currentHeight)
	if err != nil {
		return OverlayMembershipManager{}, OverlayMembershipRecord{}, err
	}
	next := m.Clone()
	replaced := false
	for i, existing := range next.Records {
		if existing.OverlayID == membership.OverlayID && existing.NodeID == membership.NodeID {
			next.Records[i] = membership
			replaced = true
			break
		}
	}
	if !replaced {
		next.Records = append(next.Records, membership)
	}
	sortOverlayMembershipRecords(next.Records)
	return next, membership, next.Validate(currentHeight)
}

func (m OverlayMembershipManager) Members(overlayID string, currentHeight uint64) []string {
	overlayID = normalizeHashText(overlayID)
	out := make([]string, 0)
	for _, record := range m.Records {
		record = NormalizeOverlayMembershipRecord(record)
		if record.OverlayID != overlayID {
			continue
		}
		if currentHeight > 0 && currentHeight > record.ExpiresHeight {
			continue
		}
		out = append(out, record.NodeID)
	}
	sortStrings(out)
	return out
}

func (m OverlayMembershipManager) Validate(currentHeight uint64) error {
	if err := ValidateOverlayDescriptors(m.Descriptors, currentHeight); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(m.Records))
	var previous string
	for i, record := range m.Records {
		record = NormalizeOverlayMembershipRecord(record)
		if err := record.Validate(currentHeight); err != nil {
			return err
		}
		if _, found := m.descriptorByID(record.OverlayID); !found {
			return errors.New("networking overlay membership record references unknown overlay")
		}
		key := record.OverlayID + "/" + record.NodeID
		if _, found := seen[key]; found {
			return errors.New("networking duplicate overlay membership record")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("networking overlay membership records must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func (m OverlayMembershipManager) Clone() OverlayMembershipManager {
	out := OverlayMembershipManager{
		Descriptors:	cloneOverlayDescriptors(m.Descriptors),
		Records:	make([]OverlayMembershipRecord, len(m.Records)),
	}
	for i, record := range m.Records {
		out.Records[i] = NormalizeOverlayMembershipRecord(record)
	}
	sortOverlayDescriptors(out.Descriptors)
	sortOverlayMembershipRecords(out.Records)
	return out
}

func (m OverlayMembershipManager) descriptorByID(overlayID string) (OverlayDescriptor, bool) {
	overlayID = normalizeHashText(overlayID)
	for _, desc := range m.Descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if desc.OverlayID == overlayID {
			return desc, true
		}
	}
	return OverlayDescriptor{}, false
}

func NormalizeOverlayMembershipRecord(record OverlayMembershipRecord) OverlayMembershipRecord {
	record.OverlayID = normalizeHashText(record.OverlayID)
	record.NodeID = normalizeHashText(record.NodeID)
	record.ProofID = normalizeHashText(record.ProofID)
	record.Membership = OverlayMembershipRule(string(record.Membership))
	record.Mode = OverlayMembershipMode(string(record.Mode))
	return record
}

func (r OverlayMembershipRecord) Validate(currentHeight uint64) error {
	record := NormalizeOverlayMembershipRecord(r)
	if err := ValidateHash("networking overlay membership record overlay id", record.OverlayID); err != nil {
		return err
	}
	if err := ValidateHash("networking overlay membership record node id", record.NodeID); err != nil {
		return err
	}
	if err := ValidateHash("networking overlay membership record proof id", record.ProofID); err != nil {
		return err
	}
	if !IsOverlayMembershipRule(record.Membership) {
		return fmt.Errorf("unknown networking overlay membership record rule %q", record.Membership)
	}
	if !IsOverlayMembershipMode(record.Mode) {
		return fmt.Errorf("unknown networking overlay membership record mode %q", record.Mode)
	}
	if record.JoinedHeight == 0 || record.ExpiresHeight == 0 {
		return errors.New("networking overlay membership record heights must be positive")
	}
	if record.JoinedHeight > record.ExpiresHeight {
		return errors.New("networking overlay membership record joined height cannot exceed expiry")
	}
	if currentHeight > 0 && currentHeight > record.ExpiresHeight {
		return errors.New("networking overlay membership record is expired")
	}
	return nil
}

func DefaultPeerRotationPolicy() PeerRotationPolicy {
	return PeerRotationPolicy{
		MaxRotatedPeersBps:	DefaultPeerRotationLimitBps,
		StaleAfterEpochs:	DefaultStalePeerEpochs,
	}
}

func NewPeerSetManager(desc OverlayDescriptor, localNodeID string, peers []AdaptivePeer, routingEpoch uint64, policyHash string) (PeerSetManager, error) {
	graph, err := BuildAdaptiveOverlayGraph(desc, localNodeID, peers, routingEpoch, policyHash)
	if err != nil {
		return PeerSetManager{}, err
	}
	return PeerSetManager{
		Descriptor:	NormalizeOverlayDescriptor(desc),
		Graph:		graph,
		DecayPolicy:	DefaultPeerScoreDecayPolicy(),
		RotationPolicy:	DefaultPeerRotationPolicy(),
	}, nil
}

func (m PeerSetManager) Rotate(candidates []AdaptivePeer, routingEpoch uint64) (PeerSetManager, error) {
	if err := m.Validate(); err != nil {
		return PeerSetManager{}, err
	}
	graph, err := RotateAdaptivePeerSets(m.Descriptor, m.Graph, candidates, routingEpoch, m.RotationPolicy)
	if err != nil {
		return PeerSetManager{}, err
	}
	next := m
	next.Graph = graph
	return next, next.Validate()
}

func (m PeerSetManager) RoutingGraph(committed bool, deterministicHintHash string) (RoutingGraph, error) {
	if err := m.Validate(); err != nil {
		return RoutingGraph{}, err
	}
	return BuildRoutingGraphFromAdaptiveGraph(m.Descriptor, m.Graph, committed, deterministicHintHash)
}

func (m PeerSetManager) Validate() error {
	if err := m.Descriptor.ValidateBasic(); err != nil {
		return err
	}
	if err := m.Graph.Validate(m.Descriptor); err != nil {
		return err
	}
	if err := m.DecayPolicy.Validate(); err != nil {
		return err
	}
	return m.RotationPolicy.Validate()
}

func RotateAdaptivePeerSets(desc OverlayDescriptor, graph AdaptiveOverlayGraph, candidates []AdaptivePeer, routingEpoch uint64, policy PeerRotationPolicy) (AdaptiveOverlayGraph, error) {
	desc = NormalizeOverlayDescriptor(desc)
	if err := desc.ValidateBasic(); err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	graph = NormalizeAdaptiveOverlayGraph(graph)
	if err := graph.Validate(desc); err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	if err := policy.Validate(); err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	if routingEpoch <= graph.RoutingEpoch {
		return AdaptiveOverlayGraph{}, errors.New("networking adaptive peer rotation epoch must increase")
	}
	existing, err := normalizeAdaptivePeers(uniqueAdaptivePeers(graph.FastSet, graph.StableSet, graph.RandomSet, graph.ZoneSet, graph.ServiceSet, graph.FallbackSet))
	if err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	nextCandidates, err := normalizeAdaptivePeers(candidates)
	if err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	maxRotate := int(uint64(len(existing)) * uint64(policy.MaxRotatedPeersBps) / uint64(BasisPoints))
	if maxRotate == 0 && len(nextCandidates) > 0 {
		maxRotate = 1
	}
	retained := make([]AdaptivePeer, 0, len(existing))
	rotated := 0
	for _, peer := range existing {
		if rotated < maxRotate && routingEpoch > peer.LastSeenHeight && routingEpoch-peer.LastSeenHeight > policy.StaleAfterEpochs {
			rotated++
			continue
		}
		retained = append(retained, peer)
	}
	seen := adaptivePeerIndex(retained)
	added := 0
	for _, peer := range sortAdaptivePeersBySeed(nextCandidates, HashParts("adaptive-rotation", desc.OverlayID, fmt.Sprintf("%d", routingEpoch))) {
		if added >= maxRotate {
			break
		}
		if _, found := seen[peer.NodeID]; found {
			continue
		}
		retained = append(retained, peer)
		seen[peer.NodeID] = struct{}{}
		added++
	}
	return BuildAdaptiveOverlayGraph(desc, graph.LocalNodeID, retained, routingEpoch, graph.PolicyHash)
}

func BuildRoutingGraphFromAdaptiveGraph(desc OverlayDescriptor, adaptive AdaptiveOverlayGraph, committed bool, deterministicHintHash string) (RoutingGraph, error) {
	desc = NormalizeOverlayDescriptor(desc)
	adaptive = NormalizeAdaptiveOverlayGraph(adaptive)
	if err := adaptive.Validate(desc); err != nil {
		return RoutingGraph{}, err
	}
	deterministicHintHash = normalizeHashText(deterministicHintHash)
	if deterministicHintHash != "" {
		if err := ValidateHash("networking routing graph deterministic hint", deterministicHintHash); err != nil {
			return RoutingGraph{}, err
		}
	}
	edges := make([]RoutingEdge, 0)
	addAdaptiveEdges := func(priority uint32, peers []AdaptivePeer) {
		for _, peer := range peers {
			edges = append(edges, RoutingEdge{
				FromNodeID:	adaptive.LocalNodeID,
				ToNodeID:	peer.NodeID,
				LatencyMillis:	peer.LatencyMillis,
				Weight:		peer.ScoreBps,
				Priority:	priority,
				ZoneID:		firstString(peer.ZonesSupported),
			})
		}
	}
	addAdaptiveEdges(0, adaptive.FastSet)
	addAdaptiveEdges(1, adaptive.StableSet)
	addAdaptiveEdges(2, adaptive.RandomSet)
	addAdaptiveEdges(3, adaptive.ZoneSet)
	addAdaptiveEdges(4, adaptive.ServiceSet)
	addAdaptiveEdges(5, adaptive.FallbackSet)
	graph := NormalizeRoutingGraph(RoutingGraph{
		OverlayID:		desc.OverlayID,
		Version:		adaptive.RoutingEpoch,
		Committed:		committed,
		DeterministicHintHash:	deterministicHintHash,
		Edges:			dedupeRoutingEdges(edges),
	})
	if err := graph.Validate(desc); err != nil {
		return RoutingGraph{}, err
	}
	return graph, nil
}

func BuildOverlayRouteWithFallback(req OverlayRoutingRequest, descriptors []OverlayDescriptor, fallbackGraph AdaptiveOverlayGraph) (OverlayRoutePlan, error) {
	plan, err := BuildOverlayRoute(req, descriptors)
	if err == nil {
		return plan, nil
	}
	msg := req.Message.Normalize()
	if msg.ReplaySafeID == "" {
		msg.ReplaySafeID = ComputeNetworkMessageID(msg)
	}
	if msg.ConsensusEffect {
		return OverlayRoutePlan{}, err
	}
	desc, found, classifyErr := ClassifyOverlayForMessage(msg, descriptors)
	if classifyErr != nil {
		return OverlayRoutePlan{}, classifyErr
	}
	if !found {
		return OverlayRoutePlan{}, err
	}
	fallbackGraph = NormalizeAdaptiveOverlayGraph(fallbackGraph)
	if fallbackGraph.OverlayID != desc.OverlayID {
		return OverlayRoutePlan{}, errors.New("networking fallback graph overlay mismatch")
	}
	if len(fallbackGraph.FallbackSet) == 0 {
		return OverlayRoutePlan{}, err
	}
	targets := make([]string, 0, len(fallbackGraph.FallbackSet))
	for _, peer := range fallbackGraph.FallbackSet {
		targets = append(targets, peer.NodeID)
	}
	sortStrings(targets)
	fanout := int(desc.Fanout)
	if fanout > 0 && len(targets) > fanout {
		targets = targets[:fanout]
	}
	return OverlayRoutePlan{
		MessageID:			msg.ReplaySafeID,
		OverlayID:			desc.OverlayID,
		OverlayType:			desc.OverlayType,
		Strategy:			RoutingStrategyProbabilisticGossip,
		TargetNodeIDs:			targets,
		UsesNodeLocalAdaptation:	true,
		FallbackUsed:			true,
	}, nil
}

func (p PeerRotationPolicy) Validate() error {
	if p.MaxRotatedPeersBps == 0 || p.MaxRotatedPeersBps > BasisPoints {
		return fmt.Errorf("networking peer rotation limit must be between 1 and %d bps", BasisPoints)
	}
	if p.StaleAfterEpochs == 0 {
		return errors.New("networking stale peer epochs must be positive")
	}
	return nil
}

func sortOverlayMembershipRecords(records []OverlayMembershipRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := NormalizeOverlayMembershipRecord(records[i])
		right := NormalizeOverlayMembershipRecord(records[j])
		if left.OverlayID != right.OverlayID {
			return left.OverlayID < right.OverlayID
		}
		return left.NodeID < right.NodeID
	})
}

func uniqueAdaptivePeers(sets ...[]AdaptivePeer) []AdaptivePeer {
	out := make([]AdaptivePeer, 0)
	seen := make(map[string]struct{})
	for _, set := range sets {
		for _, peer := range set {
			peer.NodeID = normalizeHashText(peer.NodeID)
			if _, found := seen[peer.NodeID]; found {
				continue
			}
			seen[peer.NodeID] = struct{}{}
			out = append(out, peer)
		}
	}
	return out
}

func adaptivePeerIndex(peers []AdaptivePeer) map[string]struct{} {
	out := make(map[string]struct{}, len(peers))
	for _, peer := range peers {
		out[normalizeHashText(peer.NodeID)] = struct{}{}
	}
	return out
}

func dedupeRoutingEdges(edges []RoutingEdge) []RoutingEdge {
	out := make([]RoutingEdge, 0, len(edges))
	seen := make(map[string]struct{}, len(edges))
	sort.SliceStable(edges, func(i, j int) bool {
		if edges[i].Priority != edges[j].Priority {
			return edges[i].Priority < edges[j].Priority
		}
		if edges[i].FromNodeID != edges[j].FromNodeID {
			return edges[i].FromNodeID < edges[j].FromNodeID
		}
		return edges[i].ToNodeID < edges[j].ToNodeID
	})
	for _, edge := range edges {
		key := edge.FromNodeID + "/" + edge.ToNodeID
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, edge)
	}
	return out
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
