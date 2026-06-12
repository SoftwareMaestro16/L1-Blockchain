package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultDiscoveryBranchingFactor	= uint32(16)
	DefaultMaxDiscoveryHops		= uint32(64)
	DefaultMaxZoneLatencyMillis	= uint64(250)
)

type PerformanceOptimizationGoal string

const (
	PerformanceGoalParallelPropagation	PerformanceOptimizationGoal	= "parallel_message_propagation"
	PerformanceGoalMultiOverlayConcurrency	PerformanceOptimizationGoal	= "multi_overlay_concurrency"
	PerformanceGoalNoGlobalBroadcastOnly	PerformanceOptimizationGoal	= "no_global_broadcast_for_all_traffic"
	PerformanceGoalShardLocalScaling	PerformanceOptimizationGoal	= "shard_local_execution_scaling"
	PerformanceGoalZoneIsolation		PerformanceOptimizationGoal	= "zone_isolation"
	PerformanceGoalLargePayloadStreaming	PerformanceOptimizationGoal	= "streaming_large_payloads"
	PerformanceGoalBoundedGossipFanout	PerformanceOptimizationGoal	= "bounded_gossip_fanout"
	PerformanceGoalPredictableZoneLatency	PerformanceOptimizationGoal	= "predictable_latency_per_zone"
)

type PerformanceTargetProperty string

const (
	PerformanceTargetLogDiscovery		PerformanceTargetProperty	= "o_log_n_discovery"
	PerformanceTargetNeighborPropagation	PerformanceTargetProperty	= "o_k_neighbor_propagation"
	PerformanceTargetBoundedGossip		PerformanceTargetProperty	= "bounded_gossip_fanout"
	PerformanceTargetHeaderFirstBlocks	PerformanceTargetProperty	= "header_first_block_propagation"
	PerformanceTargetParallelChunks		PerformanceTargetProperty	= "parallel_chunk_streaming"
	PerformanceTargetZoneLocalLatency	PerformanceTargetProperty	= "zone_local_low_latency_paths"
	PerformanceTargetServiceQoSIsolation	PerformanceTargetProperty	= "service_traffic_isolated_from_consensus"
	PerformanceTargetMultiOverlay		PerformanceTargetProperty	= "multi_overlay_concurrency"
	PerformanceTargetShardLocalExecution	PerformanceTargetProperty	= "shard_local_execution_scaling"
	PerformanceTargetNoGlobalBroadcastOnly	PerformanceTargetProperty	= "no_global_broadcast_dependency"
)

type PerformanceModelInput struct {
	PeerCount			uint32
	DiscoveryBranchingFactor	uint32
	OverlayDescriptors		[]OverlayDescriptor
	RoutingGraphs			[]RoutingGraph
	BroadcastPlans			[]BroadcastPlan
	BlockSession			BlockPropagationSession
	StreamPlan			StreamParallelFetchPlan
	QoSPolicies			[]QoSClassPolicy
	ZoneID				string
	MaxZoneLatencyMillis		uint64
}

type PerformanceModelPlan struct {
	PeerCount			uint32
	DiscoveryHops			uint32
	OverlayConcurrency		uint32
	MaxNeighborFanout		uint32
	GlobalBroadcastOnly		bool
	ShardLocalExecution		bool
	ZoneIsolated			bool
	HeaderFirstBlockPropagation	bool
	ParallelChunkStreaming		bool
	ZoneLatencyMillis		uint64
	ServiceTrafficIsolated		bool
	SatisfiedOptimizationGoals	[]PerformanceOptimizationGoal
	SatisfiedTargetProperties	[]PerformanceTargetProperty
}

type PeerRoleCountMetric struct {
	Role	NodeRole
	Count	uint64
}

type OverlayPerformanceMetric struct {
	OverlayID			string
	MembershipSize			uint64
	MessagePropagationLatency	PerformanceLatencySummary
	RouteFailureRateBps		uint32
}

type PerformanceLatencySummary struct {
	Count		uint64
	MinMillis	uint64
	MaxMillis	uint64
	AverageMillis	uint64
}

type PropagationLatencySample struct {
	OverlayID	string
	MessageID	string
	LatencyMillis	uint64
}

type RouteFailureSample struct {
	OverlayID	string
	Attempts	uint64
	Failures	uint64
}

type CrossZoneDeliverySample struct {
	SourceZone	string
	DestinationZone	string
	Sequence	uint64
	LatencyMillis	uint64
}

type ChannelBandwidthMetric struct {
	Channel		ChannelClass
	BytesEnqueued	uint64
	BytesSent	uint64
	BytesDropped	uint64
	UsageBps	uint32
}

type PeerScoreDistributionMetric struct {
	Count		uint64
	MinBps		uint32
	MaxBps		uint32
	AverageBps	uint32
	LowCount	uint64
	MidCount	uint64
	HighCount	uint64
}

type BlockPropagationBenchmark struct {
	HeaderLatencyMillis		uint64
	ReconstructionMillis		uint64
	ChunkCount			uint32
	HeaderFirst			bool
	ReconstructionThroughput	uint64
}

type ChunkStreamingBenchmark struct {
	StreamID		string
	ThroughputBytesBps	uint64
	RetryRateBps		uint32
	StallCount		uint64
	ParallelRequests	uint32
}

type PerformanceMetricsInput struct {
	NodeRecords			[]NodeRecord
	OverlayMemberships		[]OverlayMembershipRecord
	MessageLatencies		[]PropagationLatencySample
	RouteFailures			[]RouteFailureSample
	BlockSession			BlockPropagationSession
	BlockHeaderLatencyMillis	uint64
	BlockReconstructionMillis	uint64
	BlockBytes			uint64
	ChunkAttempts			uint64
	ChunkRetries			uint64
	StreamMetrics			[]StreamMetrics
	StreamPlans			[]StreamParallelFetchPlan
	DiscoveryLatencies		[]uint64
	CrossZoneDeliveries		[]CrossZoneDeliverySample
	ChannelMetrics			[]L0ChannelMetrics
	PeerScores			[]PeerScore
}

type PerformanceMetricsSnapshot struct {
	PeerCountByRole			[]PeerRoleCountMetric
	OverlayMetrics			[]OverlayPerformanceMetric
	BlockBenchmark			BlockPropagationBenchmark
	ChunkBenchmarks			[]ChunkStreamingBenchmark
	DiscoveryQueryLatency		PerformanceLatencySummary
	CrossZoneDeliveryLatency	PerformanceLatencySummary
	ChannelBandwidth		[]ChannelBandwidthMetric
	PeerScoreDistribution		PeerScoreDistributionMetric
	ServiceTrafficIsolated		bool
	RouteFailureRateBps		uint32
	MessagePropagationLatency	PerformanceLatencySummary
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
		PeerCount:			normalized.PeerCount,
		DiscoveryHops:			hops,
		OverlayConcurrency:		uint32(len(descriptors)),
		MaxNeighborFanout:		maxFanout,
		GlobalBroadcastOnly:		globalOnly,
		ShardLocalExecution:		shardLocal,
		ZoneIsolated:			zoneIsolated,
		HeaderFirstBlockPropagation:	headerFirst,
		ParallelChunkStreaming:		parallelChunks,
		ZoneLatencyMillis:		zoneLatency,
		ServiceTrafficIsolated:		serviceIsolated,
	}
	plan.SatisfiedOptimizationGoals = performanceGoals(plan)
	plan.SatisfiedTargetProperties = performanceTargets(plan)
	return plan, nil
}

func BuildPerformanceMetricsSnapshot(input PerformanceMetricsInput) (PerformanceMetricsSnapshot, error) {
	peerCounts, err := ComputePeerCountByRole(input.NodeRecords)
	if err != nil {
		return PerformanceMetricsSnapshot{}, err
	}
	overlayMetrics, err := ComputeOverlayPerformanceMetrics(input.OverlayMemberships, input.MessageLatencies, input.RouteFailures)
	if err != nil {
		return PerformanceMetricsSnapshot{}, err
	}
	blockBenchmark, err := BenchmarkBlockPropagation(input.BlockSession, input.BlockHeaderLatencyMillis, input.BlockReconstructionMillis, input.BlockBytes)
	if err != nil {
		return PerformanceMetricsSnapshot{}, err
	}
	chunkBenchmarks, err := BenchmarkChunkStreaming(input.StreamMetrics, input.StreamPlans, input.ChunkAttempts, input.ChunkRetries)
	if err != nil {
		return PerformanceMetricsSnapshot{}, err
	}
	channelBandwidth, err := ComputeChannelBandwidthMetrics(input.ChannelMetrics)
	if err != nil {
		return PerformanceMetricsSnapshot{}, err
	}
	scoreDistribution, err := ComputePeerScoreDistribution(input.PeerScores)
	if err != nil {
		return PerformanceMetricsSnapshot{}, err
	}
	routeFailureRate, err := ComputeRouteFailureRate(input.RouteFailures)
	if err != nil {
		return PerformanceMetricsSnapshot{}, err
	}
	serviceIsolated := ValidateServiceTrafficIsolationFromMetrics(input.ChannelMetrics) == nil
	return PerformanceMetricsSnapshot{
		PeerCountByRole:		peerCounts,
		OverlayMetrics:			overlayMetrics,
		BlockBenchmark:			blockBenchmark,
		ChunkBenchmarks:		chunkBenchmarks,
		DiscoveryQueryLatency:		SummarizeLatency(input.DiscoveryLatencies),
		CrossZoneDeliveryLatency:	SummarizeCrossZoneLatency(input.CrossZoneDeliveries),
		ChannelBandwidth:		channelBandwidth,
		PeerScoreDistribution:		scoreDistribution,
		ServiceTrafficIsolated:		serviceIsolated,
		RouteFailureRateBps:		routeFailureRate,
		MessagePropagationLatency:	SummarizePropagationLatency(input.MessageLatencies),
	}, nil
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

func ComputePeerCountByRole(records []NodeRecord) ([]PeerRoleCountMetric, error) {
	counts := make(map[NodeRole]uint64)
	for _, record := range records {
		record = NormalizeNodeRecord(record)
		if err := record.ValidateBasic(); err != nil {
			return nil, err
		}
		for _, role := range record.Roles {
			counts[role]++
		}
	}
	out := make([]PeerRoleCountMetric, 0, len(counts))
	for role, count := range counts {
		out = append(out, PeerRoleCountMetric{Role: role, Count: count})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Role < out[j].Role
	})
	return out, nil
}

func ComputeOverlayPerformanceMetrics(memberships []OverlayMembershipRecord, latencies []PropagationLatencySample, failures []RouteFailureSample) ([]OverlayPerformanceMetric, error) {
	byOverlay := make(map[string]OverlayPerformanceMetric)
	for _, membership := range memberships {
		overlayID := normalizeHashText(membership.OverlayID)
		if err := ValidateHash("networking performance overlay membership id", overlayID); err != nil {
			return nil, err
		}
		metric := byOverlay[overlayID]
		metric.OverlayID = overlayID
		metric.MembershipSize++
		byOverlay[overlayID] = metric
	}
	latencyByOverlay := make(map[string][]uint64)
	for _, sample := range latencies {
		overlayID := normalizeHashText(sample.OverlayID)
		if err := ValidateHash("networking performance latency overlay id", overlayID); err != nil {
			return nil, err
		}
		if sample.LatencyMillis == 0 {
			return nil, errors.New("networking performance propagation latency must be positive")
		}
		latencyByOverlay[overlayID] = append(latencyByOverlay[overlayID], sample.LatencyMillis)
		metric := byOverlay[overlayID]
		metric.OverlayID = overlayID
		byOverlay[overlayID] = metric
	}
	failuresByOverlay := make(map[string]RouteFailureSample)
	for _, failure := range failures {
		overlayID := normalizeHashText(failure.OverlayID)
		if err := ValidateHash("networking performance route failure overlay id", overlayID); err != nil {
			return nil, err
		}
		if failure.Failures > failure.Attempts {
			return nil, errors.New("networking performance route failures exceed attempts")
		}
		merged := failuresByOverlay[overlayID]
		merged.OverlayID = overlayID
		merged.Attempts += failure.Attempts
		merged.Failures += failure.Failures
		failuresByOverlay[overlayID] = merged
		metric := byOverlay[overlayID]
		metric.OverlayID = overlayID
		byOverlay[overlayID] = metric
	}
	out := make([]OverlayPerformanceMetric, 0, len(byOverlay))
	for overlayID, metric := range byOverlay {
		metric.MessagePropagationLatency = SummarizeLatency(latencyByOverlay[overlayID])
		if failure := failuresByOverlay[overlayID]; failure.Attempts > 0 {
			metric.RouteFailureRateBps = uint32(failure.Failures * uint64(BasisPoints) / failure.Attempts)
		}
		out = append(out, metric)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].OverlayID < out[j].OverlayID
	})
	return out, nil
}

func BenchmarkBlockPropagation(session BlockPropagationSession, headerLatencyMillis, reconstructionMillis, blockBytes uint64) (BlockPropagationBenchmark, error) {
	if err := ValidateHeaderFirstPerformance(session); err != nil {
		return BlockPropagationBenchmark{}, err
	}
	if headerLatencyMillis == 0 {
		return BlockPropagationBenchmark{}, errors.New("networking performance block header latency is required")
	}
	if reconstructionMillis == 0 {
		return BlockPropagationBenchmark{}, errors.New("networking performance block reconstruction time is required")
	}
	throughput := uint64(0)
	if blockBytes > 0 {
		throughput = blockBytes * 1_000 / reconstructionMillis
	}
	return BlockPropagationBenchmark{
		HeaderLatencyMillis:		headerLatencyMillis,
		ReconstructionMillis:		reconstructionMillis,
		ChunkCount:			session.Header.ChunkCount,
		HeaderFirst:			true,
		ReconstructionThroughput:	throughput,
	}, nil
}

func BenchmarkChunkStreaming(metrics []StreamMetrics, plans []StreamParallelFetchPlan, attempts, retries uint64) ([]ChunkStreamingBenchmark, error) {
	if retries > attempts {
		return nil, errors.New("networking performance chunk retries exceed attempts")
	}
	retryRate := uint32(0)
	if attempts > 0 {
		retryRate = uint32(retries * uint64(BasisPoints) / attempts)
	}
	planParallelism := make(map[string]uint32, len(plans))
	for _, plan := range plans {
		if err := ValidateParallelChunkPerformance(plan); err != nil {
			return nil, err
		}
		planParallelism[normalizeHashText(plan.StreamID)] = uint32(len(plan.Requests))
	}
	out := make([]ChunkStreamingBenchmark, 0, len(metrics))
	for _, metric := range metrics {
		if err := ValidateHash("networking performance stream metric id", normalizeHashText(metric.StreamID)); err != nil {
			return nil, err
		}
		out = append(out, ChunkStreamingBenchmark{
			StreamID:		normalizeHashText(metric.StreamID),
			ThroughputBytesBps:	metric.ThroughputBytesBps,
			RetryRateBps:		retryRate,
			StallCount:		metric.StallCount,
			ParallelRequests:	planParallelism[normalizeHashText(metric.StreamID)],
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].StreamID < out[j].StreamID
	})
	return out, nil
}

func ComputeChannelBandwidthMetrics(metrics []L0ChannelMetrics) ([]ChannelBandwidthMetric, error) {
	out := make([]ChannelBandwidthMetric, 0, len(metrics))
	for _, metric := range sortL0Metrics(metrics) {
		if !IsChannelClass(metric.Channel) {
			return nil, fmt.Errorf("unknown networking performance channel %q", metric.Channel)
		}
		droppedBytes := uint64(0)
		if metric.EnqueuedCount > 0 && metric.DroppedCount > 0 {
			average := metric.BytesEnqueued / metric.EnqueuedCount
			droppedBytes = average * metric.DroppedCount
		}
		usage := uint32(0)
		if metric.BytesEnqueued > 0 {
			usage = uint32(metric.BytesSent * uint64(BasisPoints) / metric.BytesEnqueued)
		}
		out = append(out, ChannelBandwidthMetric{
			Channel:	metric.Channel,
			BytesEnqueued:	metric.BytesEnqueued,
			BytesSent:	metric.BytesSent,
			BytesDropped:	droppedBytes,
			UsageBps:	usage,
		})
	}
	return out, nil
}

func ComputePeerScoreDistribution(scores []PeerScore) (PeerScoreDistributionMetric, error) {
	if len(scores) == 0 {
		return PeerScoreDistributionMetric{}, errors.New("networking performance peer scores are required")
	}
	dist := PeerScoreDistributionMetric{Count: uint64(len(scores)), MinBps: BasisPoints}
	var total uint64
	for _, score := range scores {
		if score.ScoreBps > BasisPoints {
			return PeerScoreDistributionMetric{}, fmt.Errorf("networking performance peer score must be <= %d bps", BasisPoints)
		}
		if score.ScoreBps < dist.MinBps {
			dist.MinBps = score.ScoreBps
		}
		if score.ScoreBps > dist.MaxBps {
			dist.MaxBps = score.ScoreBps
		}
		total += uint64(score.ScoreBps)
		switch {
		case score.ScoreBps < 4_000:
			dist.LowCount++
		case score.ScoreBps < 8_000:
			dist.MidCount++
		default:
			dist.HighCount++
		}
	}
	dist.AverageBps = uint32(total / uint64(len(scores)))
	return dist, nil
}

func ComputeRouteFailureRate(samples []RouteFailureSample) (uint32, error) {
	var attempts uint64
	var failures uint64
	for _, sample := range samples {
		if sample.Failures > sample.Attempts {
			return 0, errors.New("networking performance route failures exceed attempts")
		}
		attempts += sample.Attempts
		failures += sample.Failures
	}
	if attempts == 0 {
		return 0, nil
	}
	return uint32(failures * uint64(BasisPoints) / attempts), nil
}

func ValidateServiceTrafficIsolationFromMetrics(metrics []L0ChannelMetrics) error {
	var consensus L0ChannelMetrics
	var service L0ChannelMetrics
	for _, metric := range metrics {
		switch metric.Channel {
		case ChannelConsensus:
			consensus = metric
		case ChannelService:
			service = metric
		}
	}
	if consensus.DroppedCount > 0 || consensus.ConsensusDelayBlocks > 0 {
		return errors.New("networking performance consensus traffic was delayed or dropped")
	}
	if consensus.SentCount == 0 && service.SentCount > 0 {
		return errors.New("networking performance service traffic cannot progress while consensus is unsent")
	}
	return nil
}

func SummarizePropagationLatency(samples []PropagationLatencySample) PerformanceLatencySummary {
	latencies := make([]uint64, 0, len(samples))
	for _, sample := range samples {
		if sample.LatencyMillis > 0 {
			latencies = append(latencies, sample.LatencyMillis)
		}
	}
	return SummarizeLatency(latencies)
}

func SummarizeCrossZoneLatency(samples []CrossZoneDeliverySample) PerformanceLatencySummary {
	latencies := make([]uint64, 0, len(samples))
	for _, sample := range samples {
		if sample.LatencyMillis > 0 {
			latencies = append(latencies, sample.LatencyMillis)
		}
	}
	return SummarizeLatency(latencies)
}

func SummarizeLatency(latencies []uint64) PerformanceLatencySummary {
	if len(latencies) == 0 {
		return PerformanceLatencySummary{}
	}
	summary := PerformanceLatencySummary{Count: uint64(len(latencies)), MinMillis: ^uint64(0)}
	var total uint64
	for _, latency := range latencies {
		if latency < summary.MinMillis {
			summary.MinMillis = latency
		}
		if latency > summary.MaxMillis {
			summary.MaxMillis = latency
		}
		total += latency
	}
	summary.AverageMillis = total / uint64(len(latencies))
	return summary
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
