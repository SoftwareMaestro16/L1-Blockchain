package types

import (
	"errors"
	"strings"
)

const (
	DefaultDiscoveryBranchingFactor = uint32(16)
	DefaultMaxDiscoveryHops         = uint32(64)
	DefaultMaxZoneLatencyMillis     = uint64(250)
)

type PerformanceOptimizationGoal string

const (
	PerformanceGoalParallelPropagation     PerformanceOptimizationGoal = "parallel_message_propagation"
	PerformanceGoalMultiOverlayConcurrency PerformanceOptimizationGoal = "multi_overlay_concurrency"
	PerformanceGoalNoGlobalBroadcastOnly   PerformanceOptimizationGoal = "no_global_broadcast_for_all_traffic"
	PerformanceGoalShardLocalScaling       PerformanceOptimizationGoal = "shard_local_execution_scaling"
	PerformanceGoalZoneIsolation           PerformanceOptimizationGoal = "zone_isolation"
	PerformanceGoalLargePayloadStreaming   PerformanceOptimizationGoal = "streaming_large_payloads"
	PerformanceGoalBoundedGossipFanout     PerformanceOptimizationGoal = "bounded_gossip_fanout"
	PerformanceGoalPredictableZoneLatency  PerformanceOptimizationGoal = "predictable_latency_per_zone"
)

type PerformanceTargetProperty string

const (
	PerformanceTargetLogDiscovery          PerformanceTargetProperty = "o_log_n_discovery"
	PerformanceTargetNeighborPropagation   PerformanceTargetProperty = "o_k_neighbor_propagation"
	PerformanceTargetBoundedGossip         PerformanceTargetProperty = "bounded_gossip_fanout"
	PerformanceTargetHeaderFirstBlocks     PerformanceTargetProperty = "header_first_block_propagation"
	PerformanceTargetParallelChunks        PerformanceTargetProperty = "parallel_chunk_streaming"
	PerformanceTargetZoneLocalLatency      PerformanceTargetProperty = "zone_local_low_latency_paths"
	PerformanceTargetServiceQoSIsolation   PerformanceTargetProperty = "service_traffic_isolated_from_consensus"
	PerformanceTargetMultiOverlay          PerformanceTargetProperty = "multi_overlay_concurrency"
	PerformanceTargetShardLocalExecution   PerformanceTargetProperty = "shard_local_execution_scaling"
	PerformanceTargetNoGlobalBroadcastOnly PerformanceTargetProperty = "no_global_broadcast_dependency"
)

type PerformanceModelInput struct {
	PeerCount                uint32
	DiscoveryBranchingFactor uint32
	OverlayDescriptors       []OverlayDescriptor
	RoutingGraphs            []RoutingGraph
	BroadcastPlans           []BroadcastPlan
	BlockSession             BlockPropagationSession
	StreamPlan               StreamParallelFetchPlan
	QoSPolicies              []QoSClassPolicy
	ZoneID                   string
	MaxZoneLatencyMillis     uint64
}

type PerformanceModelPlan struct {
	PeerCount                   uint32
	DiscoveryHops               uint32
	OverlayConcurrency          uint32
	MaxNeighborFanout           uint32
	GlobalBroadcastOnly         bool
	ShardLocalExecution         bool
	ZoneIsolated                bool
	HeaderFirstBlockPropagation bool
	ParallelChunkStreaming      bool
	ZoneLatencyMillis           uint64
	ServiceTrafficIsolated      bool
	SatisfiedOptimizationGoals  []PerformanceOptimizationGoal
	SatisfiedTargetProperties   []PerformanceTargetProperty
}

func BuildPerformanceModelPlan(input PerformanceModelInput) (PerformanceModelPlan, error) {
	normalized, err := normalizePerformanceInput(input)
	if err != nil {
		return PerformanceModelPlan{}, err
	}
	hops, err := EstimateDiscoveryHops(normalized.PeerCount, normalized.DiscoveryBranchingFactor)
	if err != nil {
		return PerformanceModelPlan{}, err
	}
	descriptors, err := normalizePerformanceDescriptors(normalized.OverlayDescriptors)
	if err != nil {
		return PerformanceModelPlan{}, err
	}
	maxFanout, globalOnly := performanceFanoutProfile(descriptors)
	zoneLatency, zoneIsolated, err := EvaluateZoneLocalLatency(normalized.RoutingGraphs, normalized.ZoneID, normalized.MaxZoneLatencyMillis)
	if err != nil {
		return PerformanceModelPlan{}, err
	}
	serviceIsolated := ValidatePerformanceQoSIsolation(normalized.QoSPolicies) == nil
	headerFirst := ValidateHeaderFirstPerformance(normalized.BlockSession) == nil
	parallelChunks := ValidateParallelChunkPerformance(normalized.StreamPlan) == nil
	shardLocal := performanceHasOverlayType(descriptors, OverlayTypeZone) && performanceHasOverlayType(descriptors, OverlayTypeExecution)
	plan := PerformanceModelPlan{
		PeerCount:                   normalized.PeerCount,
		DiscoveryHops:               hops,
		OverlayConcurrency:          uint32(len(descriptors)),
		MaxNeighborFanout:           maxFanout,
		GlobalBroadcastOnly:         globalOnly,
		ShardLocalExecution:         shardLocal,
		ZoneIsolated:                zoneIsolated,
		HeaderFirstBlockPropagation: headerFirst,
		ParallelChunkStreaming:      parallelChunks,
		ZoneLatencyMillis:           zoneLatency,
		ServiceTrafficIsolated:      serviceIsolated,
	}
	plan.SatisfiedOptimizationGoals = performanceGoals(plan)
	plan.SatisfiedTargetProperties = performanceTargets(plan)
	return plan, nil
}

func ValidatePerformanceModelPlan(plan PerformanceModelPlan) error {
	if plan.PeerCount == 0 {
		return errors.New("networking performance peer count is required")
	}
	if plan.DiscoveryHops == 0 || plan.DiscoveryHops > DefaultMaxDiscoveryHops {
		return errors.New("networking performance discovery must be bounded")
	}
	if plan.OverlayConcurrency < 2 {
		return errors.New("networking performance requires multi-overlay concurrency")
	}
	if plan.GlobalBroadcastOnly {
		return errors.New("networking performance must not depend on global broadcast for all traffic")
	}
	if plan.MaxNeighborFanout == 0 || plan.MaxNeighborFanout > MaxBroadcastFanout {
		return errors.New("networking performance neighbor propagation fanout is invalid")
	}
	if !plan.ShardLocalExecution {
		return errors.New("networking performance requires shard-local execution scaling")
	}
	if !plan.ZoneIsolated {
		return errors.New("networking performance requires zone isolation")
	}
	if !plan.HeaderFirstBlockPropagation {
		return errors.New("networking performance requires header-first block propagation")
	}
	if !plan.ParallelChunkStreaming {
		return errors.New("networking performance requires parallel chunk streaming")
	}
	if !plan.ServiceTrafficIsolated {
		return errors.New("networking performance requires service traffic isolation from consensus")
	}
	return nil
}

func EstimateDiscoveryHops(peerCount, branchingFactor uint32) (uint32, error) {
	if peerCount == 0 {
		return 0, errors.New("networking performance peer count is required")
	}
	if branchingFactor == 0 {
		branchingFactor = DefaultDiscoveryBranchingFactor
	}
	if branchingFactor < 2 {
		return 0, errors.New("networking performance discovery branching factor must be >= 2")
	}
	hops := uint32(1)
	reach := uint64(branchingFactor)
	for reach < uint64(peerCount) {
		hops++
		reach *= uint64(branchingFactor)
		if hops > DefaultMaxDiscoveryHops {
			return 0, errors.New("networking performance discovery hop bound exceeded")
		}
	}
	return hops, nil
}

func ValidateBoundedOverlayFanout(desc OverlayDescriptor, candidatePeers uint32) (uint32, error) {
	desc = NormalizeOverlayDescriptor(desc)
	if err := desc.ValidateBasic(); err != nil {
		return 0, err
	}
	fanout, err := PlanOverlayFanout(desc, candidatePeers)
	if err != nil {
		return 0, err
	}
	if fanout > desc.Fanout || fanout > desc.MaxPeers {
		return 0, errors.New("networking performance overlay fanout exceeds policy")
	}
	return fanout, nil
}

func ValidateBoundedBroadcastFanout(msg BroadcastMessage, desc OverlayDescriptor) error {
	msg = NormalizeBroadcastMessage(msg)
	desc = NormalizeOverlayDescriptor(desc)
	if err := msg.FanoutPolicy.Validate(); err != nil {
		return err
	}
	if err := desc.ValidateBasic(); err != nil {
		return err
	}
	tree := clampBroadcastFanout(msg.FanoutPolicy.TreeFanout, desc.Fanout)
	gossip := clampBroadcastFanout(msg.FanoutPolicy.GossipFanout, desc.Fanout)
	if tree > desc.Fanout || gossip > desc.Fanout {
		return errors.New("networking performance broadcast fanout exceeds overlay policy")
	}
	return nil
}

func ValidateHeaderFirstPerformance(session BlockPropagationSession) error {
	if err := session.Header.Validate(0); err != nil {
		return err
	}
	if session.Header.BlockID == "" || session.Header.ChunkCount == 0 {
		return errors.New("networking performance block propagation requires header before chunks")
	}
	if len(session.VerifiedBitmap) > 0 && len(session.VerifiedBitmap) != int(session.Header.ChunkCount) {
		return errors.New("networking performance block bitmap must match header chunk count")
	}
	return nil
}

func ValidateParallelChunkPerformance(plan StreamParallelFetchPlan) error {
	if err := ValidateHash("networking performance stream id", normalizeHashText(plan.StreamID)); err != nil {
		return err
	}
	if plan.PayloadBytes == 0 || plan.ChunkSize == 0 || plan.TotalChunks == 0 {
		return errors.New("networking performance stream plan requires payload and chunk metadata")
	}
	if len(plan.Requests) < 2 {
		return errors.New("networking performance stream plan requires parallel chunk requests")
	}
	seenPeers := make(map[string]struct{}, len(plan.Requests))
	for _, request := range plan.Requests {
		if request.StreamID != plan.StreamID {
			return errors.New("networking performance stream request id mismatch")
		}
		if request.ChunkIndex >= plan.TotalChunks {
			return errors.New("networking performance stream request index out of range")
		}
		if request.ChunkSize == 0 || request.RangeEnd <= request.RangeStart {
			return errors.New("networking performance stream request range is invalid")
		}
		if request.AssignedPeer != "" {
			seenPeers[request.AssignedPeer] = struct{}{}
		}
	}
	if len(seenPeers) < 2 {
		return errors.New("networking performance stream plan requires peer-level parallelism")
	}
	return nil
}

func EvaluateZoneLocalLatency(graphs []RoutingGraph, zoneID string, maxLatencyMillis uint64) (uint64, bool, error) {
	zoneID = normalizeZoneText(zoneID)
	if zoneID == "" {
		return 0, false, errors.New("networking performance zone id is required")
	}
	if maxLatencyMillis == 0 {
		maxLatencyMillis = DefaultMaxZoneLatencyMillis
	}
	var total uint64
	var count uint64
	for _, graph := range graphs {
		graph = NormalizeRoutingGraph(graph)
		for _, edge := range graph.Edges {
			if normalizeZoneText(edge.ZoneID) != zoneID {
				continue
			}
			total += edge.LatencyMillis
			count++
			if edge.LatencyMillis > maxLatencyMillis {
				return edge.LatencyMillis, false, nil
			}
		}
	}
	if count == 0 {
		return 0, false, errors.New("networking performance zone-local path is required")
	}
	return total / count, true, nil
}

func ValidatePerformanceQoSIsolation(policies []QoSClassPolicy) error {
	if len(policies) == 0 {
		policies = DefaultQoSClassPolicies()
	}
	if err := ValidateQoSClassPolicies(policies); err != nil {
		return err
	}
	consensus := priorityForQoSClass(policies, QoSClassCriticalConsensus)
	service := priorityForQoSClass(policies, QoSClassServiceCall)
	execution := priorityForQoSClass(policies, QoSClassExecutionMessage)
	if consensus >= service || execution >= service {
		return errors.New("networking performance service traffic must not outrank consensus or execution")
	}
	return nil
}

func normalizePerformanceInput(input PerformanceModelInput) (PerformanceModelInput, error) {
	if input.PeerCount == 0 {
		return PerformanceModelInput{}, errors.New("networking performance peer count is required")
	}
	if input.DiscoveryBranchingFactor == 0 {
		input.DiscoveryBranchingFactor = DefaultDiscoveryBranchingFactor
	}
	if input.MaxZoneLatencyMillis == 0 {
		input.MaxZoneLatencyMillis = DefaultMaxZoneLatencyMillis
	}
	if len(input.QoSPolicies) == 0 {
		input.QoSPolicies = DefaultQoSClassPolicies()
	}
	return input, nil
}

func normalizePerformanceDescriptors(descriptors []OverlayDescriptor) ([]OverlayDescriptor, error) {
	if len(descriptors) == 0 {
		return nil, errors.New("networking performance overlay descriptors are required")
	}
	out := make([]OverlayDescriptor, len(descriptors))
	for i, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if err := desc.ValidateBasic(); err != nil {
			return nil, err
		}
		out[i] = desc
	}
	sortOverlayDescriptors(out)
	return out, nil
}

func performanceFanoutProfile(descriptors []OverlayDescriptor) (uint32, bool) {
	maxFanout := uint32(0)
	globalOnly := true
	for _, desc := range descriptors {
		if desc.Fanout > maxFanout {
			maxFanout = desc.Fanout
		}
		if desc.Routing != RoutingStrategyBroadcast {
			globalOnly = false
		}
	}
	return maxFanout, globalOnly
}

func performanceHasOverlayType(descriptors []OverlayDescriptor, overlayType OverlayType) bool {
	for _, desc := range descriptors {
		if desc.OverlayType == overlayType {
			return true
		}
	}
	return false
}

func performanceGoals(plan PerformanceModelPlan) []PerformanceOptimizationGoal {
	goals := make([]PerformanceOptimizationGoal, 0, 8)
	if plan.MaxNeighborFanout > 1 {
		goals = append(goals, PerformanceGoalParallelPropagation)
	}
	if plan.OverlayConcurrency >= 2 {
		goals = append(goals, PerformanceGoalMultiOverlayConcurrency)
	}
	if !plan.GlobalBroadcastOnly {
		goals = append(goals, PerformanceGoalNoGlobalBroadcastOnly)
	}
	if plan.ShardLocalExecution {
		goals = append(goals, PerformanceGoalShardLocalScaling)
	}
	if plan.ZoneIsolated {
		goals = append(goals, PerformanceGoalZoneIsolation)
	}
	if plan.ParallelChunkStreaming {
		goals = append(goals, PerformanceGoalLargePayloadStreaming)
	}
	if plan.MaxNeighborFanout <= MaxBroadcastFanout {
		goals = append(goals, PerformanceGoalBoundedGossipFanout)
	}
	if plan.ZoneLatencyMillis > 0 && plan.ZoneLatencyMillis <= DefaultMaxZoneLatencyMillis {
		goals = append(goals, PerformanceGoalPredictableZoneLatency)
	}
	return goals
}

func performanceTargets(plan PerformanceModelPlan) []PerformanceTargetProperty {
	targets := make([]PerformanceTargetProperty, 0, 10)
	if plan.DiscoveryHops > 0 && plan.DiscoveryHops <= DefaultMaxDiscoveryHops {
		targets = append(targets, PerformanceTargetLogDiscovery)
	}
	if plan.MaxNeighborFanout > 0 {
		targets = append(targets, PerformanceTargetNeighborPropagation)
	}
	if plan.MaxNeighborFanout <= MaxBroadcastFanout {
		targets = append(targets, PerformanceTargetBoundedGossip)
	}
	if plan.HeaderFirstBlockPropagation {
		targets = append(targets, PerformanceTargetHeaderFirstBlocks)
	}
	if plan.ParallelChunkStreaming {
		targets = append(targets, PerformanceTargetParallelChunks)
	}
	if plan.ZoneIsolated {
		targets = append(targets, PerformanceTargetZoneLocalLatency)
	}
	if plan.ServiceTrafficIsolated {
		targets = append(targets, PerformanceTargetServiceQoSIsolation)
	}
	if plan.OverlayConcurrency >= 2 {
		targets = append(targets, PerformanceTargetMultiOverlay)
	}
	if plan.ShardLocalExecution {
		targets = append(targets, PerformanceTargetShardLocalExecution)
	}
	if !plan.GlobalBroadcastOnly {
		targets = append(targets, PerformanceTargetNoGlobalBroadcastOnly)
	}
	return targets
}

func normalizeZoneText(value string) string {
	return strings.TrimSpace(value)
}
