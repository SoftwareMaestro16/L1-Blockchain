package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingRoadmapPhase string

const (
	RoadmapPhaseBaselineInstrumentation NetworkingRoadmapPhase = "phase_0_baseline_and_instrumentation"
	RoadmapPhaseAetherNetworkingAdapter NetworkingRoadmapPhase = "phase_1_aether_networking_adapter"
	RoadmapPhaseNodeIdentitySessions    NetworkingRoadmapPhase = "phase_2_node_identity_and_sessions"
)

type NetworkingRoadmapTask string

const (
	RoadmapTaskInventoryCometBFTP2P        NetworkingRoadmapTask = "inventory_current_cometbft_p2p_configuration"
	RoadmapTaskPeerMetricsCollection       NetworkingRoadmapTask = "add_peer_metrics_collection"
	RoadmapTaskChannelBandwidthMetrics     NetworkingRoadmapTask = "add_channel_bandwidth_metrics"
	RoadmapTaskBlockPropagationLatency     NetworkingRoadmapTask = "add_block_propagation_latency_metrics"
	RoadmapTaskMempoolPropagationMetrics   NetworkingRoadmapTask = "add_mempool_propagation_metrics"
	RoadmapTaskNetworkParameterSchema      NetworkingRoadmapTask = "define_network_parameter_schema"
	RoadmapTaskANAWrapper                  NetworkingRoadmapTask = "implement_ana_wrapper"
	RoadmapTaskLogicalChannelClasses       NetworkingRoadmapTask = "add_logical_channel_classes"
	RoadmapTaskPeerScoreModel              NetworkingRoadmapTask = "add_peer_score_model"
	RoadmapTaskAdaptiveFanoutConfiguration NetworkingRoadmapTask = "add_adaptive_fanout_configuration"
	RoadmapTaskQoSPolicy                   NetworkingRoadmapTask = "add_qos_policy"
	RoadmapTaskServiceTrafficIsolation     NetworkingRoadmapTask = "add_service_traffic_isolation"
	RoadmapTaskNodeIDDerivation            NetworkingRoadmapTask = "define_node_id_derivation"
	RoadmapTaskNodeRecord                  NetworkingRoadmapTask = "define_node_record"
	RoadmapTaskSignedNodeAdvertisements    NetworkingRoadmapTask = "implement_signed_node_advertisements"
	RoadmapTaskSessionHandshake            NetworkingRoadmapTask = "implement_session_handshake"
	RoadmapTaskMultiplexedStreams          NetworkingRoadmapTask = "implement_multiplexed_streams"
	RoadmapTaskSessionKeyRotation          NetworkingRoadmapTask = "implement_session_key_rotation"
)

type NetworkingExitCriterion string

const (
	ExitCurrentBehaviorMeasurable    NetworkingExitCriterion = "current_network_behavior_is_measurable"
	ExitBaselineMetricsExist         NetworkingExitCriterion = "baseline_propagation_and_peer_quality_metrics_exist"
	ExitConsensusProtectedPriority   NetworkingExitCriterion = "consensus_traffic_has_protected_priority"
	ExitPeerScoringChannelMetrics    NetworkingExitCriterion = "peer_scoring_and_channel_metrics_are_available"
	ExitServiceCannotStarveConsensus NetworkingExitCriterion = "service_traffic_cannot_starve_consensus"
	ExitCryptographicNodeAuth        NetworkingExitCriterion = "nodes_authenticate_through_cryptographic_identity"
	ExitLogicalStreamsShareSession   NetworkingExitCriterion = "logical_streams_share_one_peer_session"
	ExitExpiredForgedRecordsRejected NetworkingExitCriterion = "expired_or_forged_node_records_are_rejected"
)

type RoadmapTaskStatus string

const (
	RoadmapTaskPending  RoadmapTaskStatus = "pending"
	RoadmapTaskComplete RoadmapTaskStatus = "complete"
)

type NetworkingRoadmapPhaseSpec struct {
	Phase        NetworkingRoadmapPhase
	Title        string
	Tasks        []NetworkingRoadmapTask
	ExitCriteria []NetworkingExitCriterion
	DependsOn    []NetworkingRoadmapPhase
}

type NetworkingRoadmapTaskEvidence struct {
	Task     NetworkingRoadmapTask
	Status   RoadmapTaskStatus
	Evidence string
}

type NetworkingRoadmapEvidence struct {
	CometBFTInventory       AetherNetworkingAdapter
	PerformanceSnapshot     PerformanceMetricsSnapshot
	L0Schedule              L0Schedule
	XNetworkParams          XNetworkParams
	Session                 SessionChannel
	NodeRecords             []NodeRecord
	SignedDiscoveryRecords  []DiscoveryRecord
	HandshakeReplayRejected bool
	KeyRotationAvailable    bool
}

type NetworkingRoadmapPhaseReport struct {
	Phase                 NetworkingRoadmapPhase
	Tasks                 []NetworkingRoadmapTaskEvidence
	SatisfiedExitCriteria []NetworkingExitCriterion
	Ready                 bool
	ReportHash            string
}

type NetworkingImplementationRoadmap struct {
	Phases      []NetworkingRoadmapPhaseSpec
	RoadmapRoot string
}

func DefaultNetworkingImplementationRoadmap() NetworkingImplementationRoadmap {
	roadmap := NetworkingImplementationRoadmap{
		Phases: []NetworkingRoadmapPhaseSpec{
			{
				Phase: RoadmapPhaseBaselineInstrumentation,
				Title: "Baseline and Instrumentation",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskInventoryCometBFTP2P,
					RoadmapTaskPeerMetricsCollection,
					RoadmapTaskChannelBandwidthMetrics,
					RoadmapTaskBlockPropagationLatency,
					RoadmapTaskMempoolPropagationMetrics,
					RoadmapTaskNetworkParameterSchema,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitCurrentBehaviorMeasurable,
					ExitBaselineMetricsExist,
				},
			},
			{
				Phase: RoadmapPhaseAetherNetworkingAdapter,
				Title: "Aether Networking Adapter",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskANAWrapper,
					RoadmapTaskLogicalChannelClasses,
					RoadmapTaskPeerScoreModel,
					RoadmapTaskAdaptiveFanoutConfiguration,
					RoadmapTaskQoSPolicy,
					RoadmapTaskServiceTrafficIsolation,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitConsensusProtectedPriority,
					ExitPeerScoringChannelMetrics,
					ExitServiceCannotStarveConsensus,
				},
				DependsOn: []NetworkingRoadmapPhase{RoadmapPhaseBaselineInstrumentation},
			},
			{
				Phase: RoadmapPhaseNodeIdentitySessions,
				Title: "Node Identity and Sessions",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskNodeIDDerivation,
					RoadmapTaskNodeRecord,
					RoadmapTaskSignedNodeAdvertisements,
					RoadmapTaskSessionHandshake,
					RoadmapTaskMultiplexedStreams,
					RoadmapTaskSessionKeyRotation,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitCryptographicNodeAuth,
					ExitLogicalStreamsShareSession,
					ExitExpiredForgedRecordsRejected,
				},
				DependsOn: []NetworkingRoadmapPhase{RoadmapPhaseAetherNetworkingAdapter},
			},
		},
	}
	roadmap.RoadmapRoot = ComputeNetworkingRoadmapRoot(roadmap)
	return roadmap
}

func ValidateNetworkingImplementationRoadmap(roadmap NetworkingImplementationRoadmap) error {
	roadmap = NormalizeNetworkingImplementationRoadmap(roadmap)
	if len(roadmap.Phases) != 3 {
		return errors.New("networking roadmap must define phases 0-2")
	}
	if roadmap.RoadmapRoot != ComputeNetworkingRoadmapRoot(roadmap) {
		return errors.New("networking roadmap root mismatch")
	}
	seen := make(map[NetworkingRoadmapPhase]struct{}, len(roadmap.Phases))
	for _, phase := range roadmap.Phases {
		if err := phase.Validate(); err != nil {
			return err
		}
		if _, found := seen[phase.Phase]; found {
			return errors.New("networking roadmap duplicate phase")
		}
		seen[phase.Phase] = struct{}{}
	}
	for _, required := range []NetworkingRoadmapPhase{RoadmapPhaseBaselineInstrumentation, RoadmapPhaseAetherNetworkingAdapter, RoadmapPhaseNodeIdentitySessions} {
		if _, found := seen[required]; !found {
			return fmt.Errorf("networking roadmap missing phase %s", required)
		}
	}
	for _, phase := range roadmap.Phases {
		for _, dep := range phase.DependsOn {
			if _, found := seen[dep]; !found {
				return fmt.Errorf("networking roadmap phase %s missing dependency %s", phase.Phase, dep)
			}
		}
	}
	return nil
}

func NormalizeNetworkingImplementationRoadmap(roadmap NetworkingImplementationRoadmap) NetworkingImplementationRoadmap {
	for i := range roadmap.Phases {
		roadmap.Phases[i] = NormalizeNetworkingRoadmapPhaseSpec(roadmap.Phases[i])
	}
	sortRoadmapPhaseSpecs(roadmap.Phases)
	roadmap.RoadmapRoot = normalizeHashText(roadmap.RoadmapRoot)
	return roadmap
}

func NormalizeNetworkingRoadmapPhaseSpec(phase NetworkingRoadmapPhaseSpec) NetworkingRoadmapPhaseSpec {
	phase.Phase = NetworkingRoadmapPhase(strings.ToLower(strings.TrimSpace(string(phase.Phase))))
	phase.Title = strings.TrimSpace(phase.Title)
	phase.Tasks = normalizeRoadmapTaskSet(phase.Tasks)
	phase.ExitCriteria = normalizeExitCriterionSet(phase.ExitCriteria)
	phase.DependsOn = normalizeRoadmapPhaseSet(phase.DependsOn)
	return phase
}

func (p NetworkingRoadmapPhaseSpec) Validate() error {
	phase := NormalizeNetworkingRoadmapPhaseSpec(p)
	if !IsNetworkingRoadmapPhase(phase.Phase) {
		return fmt.Errorf("unknown networking roadmap phase %q", phase.Phase)
	}
	if phase.Title == "" {
		return errors.New("networking roadmap phase requires title")
	}
	requiredTasks, requiredCriteria := roadmapRequirementsForPhase(phase.Phase)
	for _, task := range requiredTasks {
		if !containsRoadmapTask(phase.Tasks, task) {
			return fmt.Errorf("networking roadmap phase %s missing task %s", phase.Phase, task)
		}
	}
	for _, criterion := range requiredCriteria {
		if !containsExitCriterion(phase.ExitCriteria, criterion) {
			return fmt.Errorf("networking roadmap phase %s missing exit criterion %s", phase.Phase, criterion)
		}
	}
	for _, task := range phase.Tasks {
		if !IsNetworkingRoadmapTask(task) {
			return fmt.Errorf("unknown networking roadmap task %q", task)
		}
	}
	for _, criterion := range phase.ExitCriteria {
		if !IsNetworkingExitCriterion(criterion) {
			return fmt.Errorf("unknown networking exit criterion %q", criterion)
		}
	}
	return nil
}

func EvaluateRoadmapPhaseReadiness(phase NetworkingRoadmapPhase, evidence NetworkingRoadmapEvidence) (NetworkingRoadmapPhaseReport, error) {
	if !IsNetworkingRoadmapPhase(phase) {
		return NetworkingRoadmapPhaseReport{}, fmt.Errorf("unknown networking roadmap phase %q", phase)
	}
	var tasks []NetworkingRoadmapTaskEvidence
	var criteria []NetworkingExitCriterion
	var err error
	switch phase {
	case RoadmapPhaseBaselineInstrumentation:
		tasks, criteria, err = evaluatePhase0(evidence)
	case RoadmapPhaseAetherNetworkingAdapter:
		tasks, criteria, err = evaluatePhase1(evidence)
	case RoadmapPhaseNodeIdentitySessions:
		tasks, criteria, err = evaluatePhase2(evidence)
	}
	if err != nil {
		return NetworkingRoadmapPhaseReport{}, err
	}
	report := NetworkingRoadmapPhaseReport{
		Phase:                 phase,
		Tasks:                 tasks,
		SatisfiedExitCriteria: criteria,
		Ready:                 allRoadmapTasksComplete(tasks) && allExitCriteriaSatisfied(phase, criteria),
	}
	report.ReportHash = ComputeRoadmapPhaseReportHash(report)
	return report, nil
}

func ComputeNetworkingRoadmapRoot(roadmap NetworkingImplementationRoadmap) string {
	roadmap = NormalizeNetworkingImplementationRoadmap(roadmap)
	parts := []string{"networking-implementation-roadmap"}
	for _, phase := range roadmap.Phases {
		parts = append(parts, string(phase.Phase), phase.Title)
		for _, task := range phase.Tasks {
			parts = append(parts, string(task))
		}
		for _, criterion := range phase.ExitCriteria {
			parts = append(parts, string(criterion))
		}
		for _, dep := range phase.DependsOn {
			parts = append(parts, string(dep))
		}
	}
	return HashParts(parts...)
}

func ComputeRoadmapPhaseReportHash(report NetworkingRoadmapPhaseReport) string {
	parts := []string{"networking-roadmap-phase-report", string(report.Phase), fmt.Sprintf("%t", report.Ready)}
	for _, task := range normalizeTaskEvidence(report.Tasks) {
		parts = append(parts, string(task.Task), string(task.Status), strings.TrimSpace(task.Evidence))
	}
	for _, criterion := range normalizeExitCriterionSet(report.SatisfiedExitCriteria) {
		parts = append(parts, string(criterion))
	}
	return HashParts(parts...)
}

func IsNetworkingRoadmapPhase(phase NetworkingRoadmapPhase) bool {
	switch phase {
	case RoadmapPhaseBaselineInstrumentation, RoadmapPhaseAetherNetworkingAdapter, RoadmapPhaseNodeIdentitySessions:
		return true
	default:
		return false
	}
}

func IsNetworkingRoadmapTask(task NetworkingRoadmapTask) bool {
	switch task {
	case RoadmapTaskInventoryCometBFTP2P,
		RoadmapTaskPeerMetricsCollection,
		RoadmapTaskChannelBandwidthMetrics,
		RoadmapTaskBlockPropagationLatency,
		RoadmapTaskMempoolPropagationMetrics,
		RoadmapTaskNetworkParameterSchema,
		RoadmapTaskANAWrapper,
		RoadmapTaskLogicalChannelClasses,
		RoadmapTaskPeerScoreModel,
		RoadmapTaskAdaptiveFanoutConfiguration,
		RoadmapTaskQoSPolicy,
		RoadmapTaskServiceTrafficIsolation,
		RoadmapTaskNodeIDDerivation,
		RoadmapTaskNodeRecord,
		RoadmapTaskSignedNodeAdvertisements,
		RoadmapTaskSessionHandshake,
		RoadmapTaskMultiplexedStreams,
		RoadmapTaskSessionKeyRotation:
		return true
	default:
		return false
	}
}

func IsNetworkingExitCriterion(criterion NetworkingExitCriterion) bool {
	switch criterion {
	case ExitCurrentBehaviorMeasurable,
		ExitBaselineMetricsExist,
		ExitConsensusProtectedPriority,
		ExitPeerScoringChannelMetrics,
		ExitServiceCannotStarveConsensus,
		ExitCryptographicNodeAuth,
		ExitLogicalStreamsShareSession,
		ExitExpiredForgedRecordsRejected:
		return true
	default:
		return false
	}
}

func evaluatePhase0(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskInventoryCometBFTP2P, ValidateAetherNetworkingAdapter(evidence.CometBFTInventory) == nil, "CometBFT P2P baseline adapter inventory"),
		taskEvidence(RoadmapTaskPeerMetricsCollection, evidence.PerformanceSnapshot.PeerScoreDistribution.Count > 0, "peer score distribution metrics"),
		taskEvidence(RoadmapTaskChannelBandwidthMetrics, len(evidence.PerformanceSnapshot.ChannelBandwidth) > 0, "channel bandwidth metrics"),
		taskEvidence(RoadmapTaskBlockPropagationLatency, evidence.PerformanceSnapshot.BlockBenchmark.HeaderLatencyMillis > 0, "block propagation latency benchmark"),
		taskEvidence(RoadmapTaskMempoolPropagationMetrics, hasChannelBandwidthMetric(evidence.PerformanceSnapshot.ChannelBandwidth, ChannelMempool), "mempool channel bandwidth metrics"),
		taskEvidence(RoadmapTaskNetworkParameterSchema, evidence.XNetworkParams.Validate() == nil, "x/network parameter schema"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 2)
	if allRoadmapTasksComplete(tasks) {
		criteria = append(criteria, ExitCurrentBehaviorMeasurable)
	}
	if evidence.PerformanceSnapshot.PeerScoreDistribution.Count > 0 &&
		len(evidence.PerformanceSnapshot.ChannelBandwidth) > 0 &&
		evidence.PerformanceSnapshot.MessagePropagationLatency.Count > 0 {
		criteria = append(criteria, ExitBaselineMetricsExist)
	}
	return tasks, criteria, nil
}

func evaluatePhase1(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	adapterOK := ValidateAetherNetworkingAdapter(evidence.CometBFTInventory) == nil
	channelsOK := validateChannelClassesAvailable()
	scoreOK := evidence.PerformanceSnapshot.PeerScoreDistribution.Count > 0
	fanoutOK := evidence.CometBFTInventory.Fanout.Validate() == nil
	qosOK := ValidateQoSClassPolicies(DefaultQoSClassPolicies()) == nil
	isolationOK := evidence.PerformanceSnapshot.ServiceTrafficIsolated && ValidateServiceTrafficIsolationFromMetrics(evidence.L0Schedule.Metrics) == nil
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskANAWrapper, adapterOK, "ANA wrapper validates CometBFT baseline"),
		taskEvidence(RoadmapTaskLogicalChannelClasses, channelsOK, "logical channel classes are defined"),
		taskEvidence(RoadmapTaskPeerScoreModel, scoreOK, "peer score model metrics available"),
		taskEvidence(RoadmapTaskAdaptiveFanoutConfiguration, fanoutOK, "adaptive fanout policy validates"),
		taskEvidence(RoadmapTaskQoSPolicy, qosOK, "QoS policy validates"),
		taskEvidence(RoadmapTaskServiceTrafficIsolation, isolationOK, "service traffic isolation metrics"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if evidence.L0Schedule.Validate() == nil && l0ScheduleConsensusProtected(evidence.L0Schedule) {
		criteria = append(criteria, ExitConsensusProtectedPriority)
	}
	if scoreOK && len(evidence.PerformanceSnapshot.ChannelBandwidth) > 0 {
		criteria = append(criteria, ExitPeerScoringChannelMetrics)
	}
	if isolationOK {
		criteria = append(criteria, ExitServiceCannotStarveConsensus)
	}
	return tasks, criteria, nil
}

func evaluatePhase2(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	nodeIDOK := len(evidence.NodeRecords) > 0 && evidence.NodeRecords[0].NodeID == ComputeNodeID(evidence.NodeRecords[0].NodePubKey, roadmapNodeIDKey(evidence.NodeRecords[0]))
	recordOK := false
	signedAdsOK := false
	for _, record := range evidence.NodeRecords {
		if record.NodeID != "" && len(record.Signature) > 0 {
			recordOK = true
			break
		}
	}
	for _, record := range evidence.SignedDiscoveryRecords {
		if len(record.Signature) > 0 {
			signedAdsOK = true
			break
		}
	}
	sessionOK := evidence.Session.Validate() == nil
	streamsOK := sessionOK && len(evidence.Session.Streams) > 1
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskNodeIDDerivation, nodeIDOK || len(evidence.NodeRecords) > 0, "NodeID derivation available"),
		taskEvidence(RoadmapTaskNodeRecord, recordOK, "NodeRecord signed identity records"),
		taskEvidence(RoadmapTaskSignedNodeAdvertisements, signedAdsOK, "signed discovery/node advertisements"),
		taskEvidence(RoadmapTaskSessionHandshake, sessionOK, "session handshake result validates"),
		taskEvidence(RoadmapTaskMultiplexedStreams, streamsOK, "multiplexed streams on one session"),
		taskEvidence(RoadmapTaskSessionKeyRotation, evidence.KeyRotationAvailable, "session key rotation support"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if recordOK && sessionOK {
		criteria = append(criteria, ExitCryptographicNodeAuth)
	}
	if streamsOK {
		criteria = append(criteria, ExitLogicalStreamsShareSession)
	}
	if evidence.HandshakeReplayRejected {
		criteria = append(criteria, ExitExpiredForgedRecordsRejected)
	}
	return tasks, criteria, nil
}

func taskEvidence(task NetworkingRoadmapTask, complete bool, evidence string) NetworkingRoadmapTaskEvidence {
	status := RoadmapTaskPending
	if complete {
		status = RoadmapTaskComplete
	}
	return NetworkingRoadmapTaskEvidence{Task: task, Status: status, Evidence: evidence}
}

func roadmapRequirementsForPhase(phase NetworkingRoadmapPhase) ([]NetworkingRoadmapTask, []NetworkingExitCriterion) {
	switch phase {
	case RoadmapPhaseBaselineInstrumentation:
		return []NetworkingRoadmapTask{
				RoadmapTaskInventoryCometBFTP2P,
				RoadmapTaskPeerMetricsCollection,
				RoadmapTaskChannelBandwidthMetrics,
				RoadmapTaskBlockPropagationLatency,
				RoadmapTaskMempoolPropagationMetrics,
				RoadmapTaskNetworkParameterSchema,
			}, []NetworkingExitCriterion{
				ExitCurrentBehaviorMeasurable,
				ExitBaselineMetricsExist,
			}
	case RoadmapPhaseAetherNetworkingAdapter:
		return []NetworkingRoadmapTask{
				RoadmapTaskANAWrapper,
				RoadmapTaskLogicalChannelClasses,
				RoadmapTaskPeerScoreModel,
				RoadmapTaskAdaptiveFanoutConfiguration,
				RoadmapTaskQoSPolicy,
				RoadmapTaskServiceTrafficIsolation,
			}, []NetworkingExitCriterion{
				ExitConsensusProtectedPriority,
				ExitPeerScoringChannelMetrics,
				ExitServiceCannotStarveConsensus,
			}
	case RoadmapPhaseNodeIdentitySessions:
		return []NetworkingRoadmapTask{
				RoadmapTaskNodeIDDerivation,
				RoadmapTaskNodeRecord,
				RoadmapTaskSignedNodeAdvertisements,
				RoadmapTaskSessionHandshake,
				RoadmapTaskMultiplexedStreams,
				RoadmapTaskSessionKeyRotation,
			}, []NetworkingExitCriterion{
				ExitCryptographicNodeAuth,
				ExitLogicalStreamsShareSession,
				ExitExpiredForgedRecordsRejected,
			}
	default:
		return nil, nil
	}
}

func allRoadmapTasksComplete(tasks []NetworkingRoadmapTaskEvidence) bool {
	if len(tasks) == 0 {
		return false
	}
	for _, task := range tasks {
		if task.Status != RoadmapTaskComplete {
			return false
		}
	}
	return true
}

func allExitCriteriaSatisfied(phase NetworkingRoadmapPhase, criteria []NetworkingExitCriterion) bool {
	_, required := roadmapRequirementsForPhase(phase)
	for _, criterion := range required {
		if !containsExitCriterion(criteria, criterion) {
			return false
		}
	}
	return true
}

func validateChannelClassesAvailable() bool {
	for _, channel := range []ChannelClass{ChannelConsensus, ChannelMempool, ChannelBlock, ChannelStateSync, ChannelData, ChannelExecution, ChannelService, ChannelRouting, ChannelDiscovery} {
		if !IsChannelClass(channel) {
			return false
		}
	}
	return true
}

func hasChannelBandwidthMetric(metrics []ChannelBandwidthMetric, channel ChannelClass) bool {
	for _, metric := range metrics {
		if metric.Channel == channel {
			return true
		}
	}
	return false
}

func l0ScheduleConsensusProtected(schedule L0Schedule) bool {
	for _, metric := range schedule.Metrics {
		if metric.Channel == ChannelConsensus && (metric.DroppedCount > 0 || metric.ConsensusDelayBlocks > 0) {
			return false
		}
	}
	for _, plan := range schedule.Plans {
		if plan.Envelope.Channel == ChannelConsensus && !plan.HandledByCometBFT {
			return false
		}
	}
	return true
}

func roadmapNodeIDKey(record NodeRecord) []byte {
	if len(record.ValidatorPubKey) > 0 {
		return record.ValidatorPubKey
	}
	return record.NodePubKey
}

func normalizeRoadmapTaskSet(values []NetworkingRoadmapTask) []NetworkingRoadmapTask {
	seen := make(map[NetworkingRoadmapTask]struct{}, len(values))
	out := make([]NetworkingRoadmapTask, 0, len(values))
	for _, value := range values {
		value = NetworkingRoadmapTask(strings.ToLower(strings.TrimSpace(string(value))))
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func normalizeExitCriterionSet(values []NetworkingExitCriterion) []NetworkingExitCriterion {
	seen := make(map[NetworkingExitCriterion]struct{}, len(values))
	out := make([]NetworkingExitCriterion, 0, len(values))
	for _, value := range values {
		value = NetworkingExitCriterion(strings.ToLower(strings.TrimSpace(string(value))))
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func normalizeRoadmapPhaseSet(values []NetworkingRoadmapPhase) []NetworkingRoadmapPhase {
	seen := make(map[NetworkingRoadmapPhase]struct{}, len(values))
	out := make([]NetworkingRoadmapPhase, 0, len(values))
	for _, value := range values {
		value = NetworkingRoadmapPhase(strings.ToLower(strings.TrimSpace(string(value))))
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func normalizeTaskEvidence(values []NetworkingRoadmapTaskEvidence) []NetworkingRoadmapTaskEvidence {
	out := make([]NetworkingRoadmapTaskEvidence, len(values))
	for i, value := range values {
		out[i] = NetworkingRoadmapTaskEvidence{
			Task:     NetworkingRoadmapTask(strings.ToLower(strings.TrimSpace(string(value.Task)))),
			Status:   RoadmapTaskStatus(strings.ToLower(strings.TrimSpace(string(value.Status)))),
			Evidence: strings.TrimSpace(value.Evidence),
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Task < out[j].Task })
	return out
}

func containsRoadmapTask(tasks []NetworkingRoadmapTask, needle NetworkingRoadmapTask) bool {
	for _, task := range tasks {
		if task == needle {
			return true
		}
	}
	return false
}

func containsExitCriterion(criteria []NetworkingExitCriterion, needle NetworkingExitCriterion) bool {
	for _, criterion := range criteria {
		if criterion == needle {
			return true
		}
	}
	return false
}

func sortRoadmapPhaseSpecs(phases []NetworkingRoadmapPhaseSpec) {
	sort.SliceStable(phases, func(i, j int) bool {
		return phases[i].Phase < phases[j].Phase
	})
}
