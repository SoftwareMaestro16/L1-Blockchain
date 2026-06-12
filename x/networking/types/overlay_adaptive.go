package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	MaxAdaptivePeersPerSet		= 128
	DefaultPeerScoreDecayBps	= uint32(1_000)
	DefaultPeerScoreDecayFloor	= uint32(500)
	DefaultRandomDiversityBucket	= 2
)

type PeerSetClass string

const (
	PeerSetFast	PeerSetClass	= "fast_set"
	PeerSetStable	PeerSetClass	= "stable_set"
	PeerSetRandom	PeerSetClass	= "random_set"
	PeerSetZone	PeerSetClass	= "zone_set"
	PeerSetService	PeerSetClass	= "service_set"
	PeerSetFallback	PeerSetClass	= "fallback_set"
)

type AdaptivePeer struct {
	NodeID		string
	ScoreBps	uint32
	LatencyMillis	uint64
	ReliabilityBps	uint32
	Roles		[]NodeRole
	ZonesSupported	[]string
	Services	[]string
	CommittedScore	bool
	LastSeenHeight	uint64
	LastScoreHeight	uint64
}

type AdaptivePeerSet struct {
	Class	PeerSetClass
	Peers	[]AdaptivePeer
}

type AdaptiveOverlayGraph struct {
	OverlayID		string
	LocalNodeID		string
	RoutingEpoch		uint64
	FastSet			[]AdaptivePeer
	StableSet		[]AdaptivePeer
	RandomSet		[]AdaptivePeer
	ZoneSet			[]AdaptivePeer
	ServiceSet		[]AdaptivePeer
	FallbackSet		[]AdaptivePeer
	PolicyHash		string
	LivePeerScoresCommitted	bool
}

type PeerScoreDecayPolicy struct {
	MaxDecayBpsPerEpoch	uint32
	MinScoreBps		uint32
}

type OverlayRouteRoot struct {
	OverlayID	string
	RootHash	string
}

type RoutingTableCommitment struct {
	RoutingEpoch		uint64
	OverlayRoots		[]OverlayRouteRoot
	ZoneRouteRoot		string
	ServiceRouteRoot	string
	PeerClassRoot		string
	CongestionSnapshotRoot	string
	PolicyHash		string
}

type RoutingTableUse struct {
	Commitment			RoutingTableCommitment
	Committed			bool
	UsedForExecutionScheduling	bool
	UsedForPhysicalForwarding	bool
	DeterministicRouteHintHash	string
	AllowsNodeLocalOptimization	bool
}

func BuildAdaptiveOverlayGraph(desc OverlayDescriptor, localNodeID string, peers []AdaptivePeer, routingEpoch uint64, policyHash string) (AdaptiveOverlayGraph, error) {
	desc = NormalizeOverlayDescriptor(desc)
	if err := desc.ValidateBasic(); err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	localNodeID = normalizeHashText(localNodeID)
	if err := ValidateHash("networking adaptive overlay local node id", localNodeID); err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	policyHash = normalizeHashText(policyHash)
	if err := ValidateHash("networking adaptive overlay policy hash", policyHash); err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	normalized, err := normalizeAdaptivePeers(peers)
	if err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	limit := int(desc.Fanout)
	if limit < int(desc.MinPeers) {
		limit = int(desc.MinPeers)
	}
	if limit > MaxAdaptivePeersPerSet {
		limit = MaxAdaptivePeersPerSet
	}
	graph := AdaptiveOverlayGraph{
		OverlayID:	desc.OverlayID,
		LocalNodeID:	localNodeID,
		RoutingEpoch:	routingEpoch,
		FastSet:	takeAdaptivePeers(sortAdaptivePeersByLatency(normalized), limit),
		StableSet:	takeAdaptivePeers(sortAdaptivePeersByReliability(normalized), limit),
		RandomSet:	takeAdaptivePeers(sortAdaptivePeersBySeed(normalized, HashParts("adaptive-random", desc.OverlayID, fmt.Sprintf("%d", routingEpoch), policyHash)), limit),
		ZoneSet:	takeAdaptivePeers(filterAdaptivePeersWithZones(normalized), limit),
		ServiceSet:	takeAdaptivePeers(filterAdaptivePeersWithServices(normalized), limit),
		FallbackSet:	takeAdaptivePeers(sortAdaptivePeersBySeed(normalized, HashParts("adaptive-fallback", desc.OverlayID, policyHash)), limit),
		PolicyHash:	policyHash,
	}
	if err := graph.Validate(desc); err != nil {
		return AdaptiveOverlayGraph{}, err
	}
	return graph, nil
}

func NormalizeAdaptiveOverlayGraph(graph AdaptiveOverlayGraph) AdaptiveOverlayGraph {
	graph.OverlayID = normalizeHashText(graph.OverlayID)
	graph.LocalNodeID = normalizeHashText(graph.LocalNodeID)
	graph.PolicyHash = normalizeHashText(graph.PolicyHash)
	graph.FastSet, _ = normalizeAdaptivePeers(graph.FastSet)
	graph.StableSet, _ = normalizeAdaptivePeers(graph.StableSet)
	graph.RandomSet, _ = normalizeAdaptivePeers(graph.RandomSet)
	graph.ZoneSet, _ = normalizeAdaptivePeers(graph.ZoneSet)
	graph.ServiceSet, _ = normalizeAdaptivePeers(graph.ServiceSet)
	graph.FallbackSet, _ = normalizeAdaptivePeers(graph.FallbackSet)
	return graph
}

func (g AdaptiveOverlayGraph) Validate(desc OverlayDescriptor) error {
	graph := NormalizeAdaptiveOverlayGraph(g)
	desc = NormalizeOverlayDescriptor(desc)
	if err := desc.ValidateBasic(); err != nil {
		return err
	}
	if graph.OverlayID != desc.OverlayID {
		return errors.New("networking adaptive overlay graph descriptor mismatch")
	}
	if err := ValidateHash("networking adaptive overlay local node id", graph.LocalNodeID); err != nil {
		return err
	}
	if graph.RoutingEpoch == 0 {
		return errors.New("networking adaptive overlay routing epoch must be positive")
	}
	if err := ValidateHash("networking adaptive overlay policy hash", graph.PolicyHash); err != nil {
		return err
	}
	sets := []AdaptivePeerSet{
		{Class: PeerSetFast, Peers: graph.FastSet},
		{Class: PeerSetStable, Peers: graph.StableSet},
		{Class: PeerSetRandom, Peers: graph.RandomSet},
		{Class: PeerSetZone, Peers: graph.ZoneSet},
		{Class: PeerSetService, Peers: graph.ServiceSet},
		{Class: PeerSetFallback, Peers: graph.FallbackSet},
	}
	for _, set := range sets {
		if err := validateAdaptivePeerSet(set); err != nil {
			return err
		}
	}
	if len(graph.RandomSet) == 0 || len(graph.FallbackSet) == 0 {
		return errors.New("networking adaptive overlay graph requires random and fallback peer sets for eclipse resistance")
	}
	if distinctAdaptivePeerBuckets(graph.RandomSet) < minUint32(DefaultRandomDiversityBucket, uint32(len(graph.RandomSet))) {
		return errors.New("networking adaptive overlay random set lacks diversity")
	}
	if len(graph.ZoneSet) > 0 && !hasGlobalPeerOutsideZoneSet(graph) {
		return errors.New("networking adaptive overlay zone peers must not replace global peer sets")
	}
	return nil
}

func ValidateAdaptiveOverlayGraphUse(graph AdaptiveOverlayGraph, usedForConsensus bool) error {
	if !usedForConsensus {
		return nil
	}
	if !graph.LivePeerScoresCommitted {
		return errors.New("networking live peer scores are advisory until committed")
	}
	return nil
}

func DefaultPeerScoreDecayPolicy() PeerScoreDecayPolicy {
	return PeerScoreDecayPolicy{
		MaxDecayBpsPerEpoch:	DefaultPeerScoreDecayBps,
		MinScoreBps:		DefaultPeerScoreDecayFloor,
	}
}

func DecayPeerScore(score PeerScore, elapsedEpochs uint64, policy PeerScoreDecayPolicy) (PeerScore, error) {
	if err := policy.Validate(); err != nil {
		return PeerScore{}, err
	}
	if score.ScoreBps > BasisPoints {
		return PeerScore{}, fmt.Errorf("networking peer score must be <= %d bps", BasisPoints)
	}
	decay := uint64(policy.MaxDecayBpsPerEpoch) * elapsedEpochs
	if decay > uint64(score.ScoreBps) {
		decay = uint64(score.ScoreBps)
	}
	next := score
	if uint64(score.ScoreBps)-decay < uint64(policy.MinScoreBps) {
		next.ScoreBps = policy.MinScoreBps
	} else {
		next.ScoreBps = uint32(uint64(score.ScoreBps) - decay)
	}
	return next, nil
}

func (p PeerScoreDecayPolicy) Validate() error {
	if p.MaxDecayBpsPerEpoch == 0 || p.MaxDecayBpsPerEpoch > BasisPoints {
		return fmt.Errorf("networking peer score decay per epoch must be between 1 and %d bps", BasisPoints)
	}
	if p.MinScoreBps > BasisPoints {
		return fmt.Errorf("networking peer score decay floor must be <= %d bps", BasisPoints)
	}
	return nil
}

func NewRoutingTableCommitment(commitment RoutingTableCommitment) (RoutingTableCommitment, error) {
	commitment = NormalizeRoutingTableCommitment(commitment)
	if err := commitment.Validate(); err != nil {
		return RoutingTableCommitment{}, err
	}
	return commitment, nil
}

func NormalizeRoutingTableCommitment(commitment RoutingTableCommitment) RoutingTableCommitment {
	commitment.ZoneRouteRoot = normalizeHashText(commitment.ZoneRouteRoot)
	commitment.ServiceRouteRoot = normalizeHashText(commitment.ServiceRouteRoot)
	commitment.PeerClassRoot = normalizeHashText(commitment.PeerClassRoot)
	commitment.CongestionSnapshotRoot = normalizeHashText(commitment.CongestionSnapshotRoot)
	commitment.PolicyHash = normalizeHashText(commitment.PolicyHash)
	commitment.OverlayRoots = cloneOverlayRouteRoots(commitment.OverlayRoots)
	sortOverlayRouteRoots(commitment.OverlayRoots)
	return commitment
}

func ComputeRoutingTableCommitmentHash(commitment RoutingTableCommitment) string {
	commitment = NormalizeRoutingTableCommitment(commitment)
	parts := []string{
		"routing-table-commitment",
		fmt.Sprintf("%d", commitment.RoutingEpoch),
		commitment.ZoneRouteRoot,
		commitment.ServiceRouteRoot,
		commitment.PeerClassRoot,
		commitment.CongestionSnapshotRoot,
		commitment.PolicyHash,
	}
	for _, root := range commitment.OverlayRoots {
		parts = append(parts, root.OverlayID, root.RootHash)
	}
	return HashParts(parts...)
}

func (c RoutingTableCommitment) Validate() error {
	commitment := NormalizeRoutingTableCommitment(c)
	if commitment.RoutingEpoch == 0 {
		return errors.New("networking routing table commitment epoch must be positive")
	}
	if len(commitment.OverlayRoots) == 0 || len(commitment.OverlayRoots) > MaxOverlayDescriptors {
		return fmt.Errorf("networking routing table overlay roots must be between 1 and %d", MaxOverlayDescriptors)
	}
	for _, field := range []struct {
		name	string
		value	string
	}{
		{"networking routing table zone route root", commitment.ZoneRouteRoot},
		{"networking routing table service route root", commitment.ServiceRouteRoot},
		{"networking routing table peer class root", commitment.PeerClassRoot},
		{"networking routing table congestion snapshot root", commitment.CongestionSnapshotRoot},
		{"networking routing table policy hash", commitment.PolicyHash},
	} {
		if err := ValidateHash(field.name, field.value); err != nil {
			return err
		}
	}
	seen := make(map[string]struct{}, len(commitment.OverlayRoots))
	var previous string
	for _, root := range commitment.OverlayRoots {
		if err := ValidateHash("networking routing table overlay id", root.OverlayID); err != nil {
			return err
		}
		if err := ValidateHash("networking routing table overlay root", root.RootHash); err != nil {
			return err
		}
		if _, found := seen[root.OverlayID]; found {
			return errors.New("networking duplicate routing table overlay root")
		}
		seen[root.OverlayID] = struct{}{}
		if previous != "" && previous >= root.OverlayID {
			return errors.New("networking routing table overlay roots must be sorted canonically")
		}
		previous = root.OverlayID
	}
	return nil
}

func ValidateRoutingTableUse(use RoutingTableUse) error {
	use.DeterministicRouteHintHash = normalizeHashText(use.DeterministicRouteHintHash)
	if err := use.Commitment.Validate(); err != nil {
		return err
	}
	if use.DeterministicRouteHintHash != "" {
		if err := ValidateHash("networking routing table deterministic route hint", use.DeterministicRouteHintHash); err != nil {
			return err
		}
	}
	if use.UsedForExecutionScheduling && !use.Committed && use.DeterministicRouteHintHash == "" {
		return errors.New("networking execution scheduling requires committed routing table or deterministic route hint")
	}
	if !use.Committed && use.UsedForExecutionScheduling {
		return errors.New("networking non-committed routing table is limited to physical forwarding")
	}
	return nil
}

func AdaptivePeerFromNodeRecord(record NodeRecord, score PeerScore, metrics PeerMetrics, committedScore bool, height uint64) AdaptivePeer {
	record = NormalizeNodeRecord(record)
	return AdaptivePeer{
		NodeID:			record.NodeID,
		ScoreBps:		score.ScoreBps,
		LatencyMillis:		metrics.LatencyMillis,
		ReliabilityBps:		metrics.ReliabilityBps,
		Roles:			append([]NodeRole(nil), record.Roles...),
		ZonesSupported:		append([]string(nil), record.ZonesSupported...),
		Services:		append([]string(nil), record.ServicesSupported...),
		CommittedScore:		committedScore,
		LastSeenHeight:		height,
		LastScoreHeight:	height,
	}
}

func normalizeAdaptivePeers(peers []AdaptivePeer) ([]AdaptivePeer, error) {
	out := make([]AdaptivePeer, 0, len(peers))
	for _, peer := range peers {
		peer.NodeID = normalizeHashText(peer.NodeID)
		if err := ValidateHash("networking adaptive peer node id", peer.NodeID); err != nil {
			return nil, err
		}
		if peer.ScoreBps > BasisPoints || peer.ReliabilityBps > BasisPoints {
			return nil, fmt.Errorf("networking adaptive peer score and reliability must be <= %d bps", BasisPoints)
		}
		peer.Roles = normalizeRoles(peer.Roles)
		peer.ZonesSupported, _ = normalizeStringSet("zone", peer.ZonesSupported, MaxZoneIDBytes)
		peer.Services, _ = normalizeStringSet("service", peer.Services, MaxServiceIDBytes)
		out = append(out, peer)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].NodeID < out[j].NodeID
	})
	return out, nil
}

func validateAdaptivePeerSet(set AdaptivePeerSet) error {
	if !IsPeerSetClass(set.Class) {
		return fmt.Errorf("unknown networking adaptive peer set %q", set.Class)
	}
	if len(set.Peers) > MaxAdaptivePeersPerSet {
		return fmt.Errorf("networking adaptive peer set must be <= %d peers", MaxAdaptivePeersPerSet)
	}
	peers, err := normalizeAdaptivePeers(set.Peers)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(peers))
	for _, peer := range peers {
		if _, found := seen[peer.NodeID]; found {
			return errors.New("networking duplicate adaptive peer in set")
		}
		seen[peer.NodeID] = struct{}{}
	}
	return nil
}

func IsPeerSetClass(class PeerSetClass) bool {
	switch class {
	case PeerSetFast, PeerSetStable, PeerSetRandom, PeerSetZone, PeerSetService, PeerSetFallback:
		return true
	default:
		return false
	}
}

func takeAdaptivePeers(peers []AdaptivePeer, limit int) []AdaptivePeer {
	if limit > len(peers) {
		limit = len(peers)
	}
	out := make([]AdaptivePeer, limit)
	copy(out, peers[:limit])
	return out
}

func sortAdaptivePeersByLatency(peers []AdaptivePeer) []AdaptivePeer {
	out := append([]AdaptivePeer(nil), peers...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LatencyMillis != out[j].LatencyMillis {
			return out[i].LatencyMillis < out[j].LatencyMillis
		}
		if out[i].ScoreBps != out[j].ScoreBps {
			return out[i].ScoreBps > out[j].ScoreBps
		}
		return out[i].NodeID < out[j].NodeID
	})
	return out
}

func sortAdaptivePeersByReliability(peers []AdaptivePeer) []AdaptivePeer {
	out := append([]AdaptivePeer(nil), peers...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ReliabilityBps != out[j].ReliabilityBps {
			return out[i].ReliabilityBps > out[j].ReliabilityBps
		}
		if out[i].ScoreBps != out[j].ScoreBps {
			return out[i].ScoreBps > out[j].ScoreBps
		}
		return out[i].NodeID < out[j].NodeID
	})
	return out
}

func sortAdaptivePeersBySeed(peers []AdaptivePeer, seed string) []AdaptivePeer {
	out := append([]AdaptivePeer(nil), peers...)
	sort.SliceStable(out, func(i, j int) bool {
		left := HashParts(seed, out[i].NodeID)
		right := HashParts(seed, out[j].NodeID)
		if left != right {
			return left < right
		}
		return out[i].NodeID < out[j].NodeID
	})
	return out
}

func filterAdaptivePeersWithZones(peers []AdaptivePeer) []AdaptivePeer {
	out := make([]AdaptivePeer, 0, len(peers))
	for _, peer := range peers {
		if len(peer.ZonesSupported) > 0 {
			out = append(out, peer)
		}
	}
	return sortAdaptivePeersByReliability(out)
}

func filterAdaptivePeersWithServices(peers []AdaptivePeer) []AdaptivePeer {
	out := make([]AdaptivePeer, 0, len(peers))
	for _, peer := range peers {
		if len(peer.Services) > 0 {
			out = append(out, peer)
		}
	}
	return sortAdaptivePeersByLatency(out)
}

func distinctAdaptivePeerBuckets(peers []AdaptivePeer) uint32 {
	buckets := make(map[string]struct{}, len(peers))
	for _, peer := range peers {
		nodeID := normalizeHashText(peer.NodeID)
		bucket := nodeID
		if len(bucket) > 2 {
			bucket = bucket[:2]
		}
		buckets[bucket] = struct{}{}
	}
	return uint32(len(buckets))
}

func hasGlobalPeerOutsideZoneSet(graph AdaptiveOverlayGraph) bool {
	zone := make(map[string]struct{}, len(graph.ZoneSet))
	for _, peer := range graph.ZoneSet {
		zone[peer.NodeID] = struct{}{}
	}
	for _, set := range [][]AdaptivePeer{graph.FastSet, graph.StableSet, graph.RandomSet, graph.FallbackSet} {
		for _, peer := range set {
			if _, found := zone[peer.NodeID]; !found {
				return true
			}
		}
	}
	return false
}

func cloneOverlayRouteRoots(roots []OverlayRouteRoot) []OverlayRouteRoot {
	out := make([]OverlayRouteRoot, len(roots))
	for i, root := range roots {
		out[i] = OverlayRouteRoot{
			OverlayID:	normalizeHashText(root.OverlayID),
			RootHash:	normalizeHashText(root.RootHash),
		}
	}
	return out
}

func sortOverlayRouteRoots(roots []OverlayRouteRoot) {
	sort.SliceStable(roots, func(i, j int) bool {
		return roots[i].OverlayID < roots[j].OverlayID
	})
}

func minUint32(left, right uint32) uint32 {
	if left < right {
		return left
	}
	return right
}
