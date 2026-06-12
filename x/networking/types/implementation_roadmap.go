package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingRoadmapPhase string

const (
	RoadmapPhaseBaselineInstrumentation	NetworkingRoadmapPhase	= "phase_0_baseline_and_instrumentation"
	RoadmapPhaseAetherNetworkingAdapter	NetworkingRoadmapPhase	= "phase_1_aether_networking_adapter"
	RoadmapPhaseNodeIdentitySessions	NetworkingRoadmapPhase	= "phase_2_node_identity_and_sessions"
	RoadmapPhaseOverlayRouting		NetworkingRoadmapPhase	= "phase_3_overlay_routing"
	RoadmapPhaseRL2Streaming		NetworkingRoadmapPhase	= "phase_4_rl2_streaming"
	RoadmapPhaseDiscoveryLayer		NetworkingRoadmapPhase	= "phase_5_discovery_layer"
	RoadmapPhaseHybridBroadcast		NetworkingRoadmapPhase	= "phase_6_hybrid_broadcast"
	RoadmapPhaseAetherMesh			NetworkingRoadmapPhase	= "phase_7_aether_mesh"
	RoadmapPhaseSecurityLoadHardening	NetworkingRoadmapPhase	= "phase_8_security_and_load_hardening"
)

type NetworkingRoadmapTask string

const (
	RoadmapTaskInventoryCometBFTP2P		NetworkingRoadmapTask	= "inventory_current_cometbft_p2p_configuration"
	RoadmapTaskPeerMetricsCollection	NetworkingRoadmapTask	= "add_peer_metrics_collection"
	RoadmapTaskChannelBandwidthMetrics	NetworkingRoadmapTask	= "add_channel_bandwidth_metrics"
	RoadmapTaskBlockPropagationLatency	NetworkingRoadmapTask	= "add_block_propagation_latency_metrics"
	RoadmapTaskMempoolPropagationMetrics	NetworkingRoadmapTask	= "add_mempool_propagation_metrics"
	RoadmapTaskNetworkParameterSchema	NetworkingRoadmapTask	= "define_network_parameter_schema"
	RoadmapTaskANAWrapper			NetworkingRoadmapTask	= "implement_ana_wrapper"
	RoadmapTaskLogicalChannelClasses	NetworkingRoadmapTask	= "add_logical_channel_classes"
	RoadmapTaskPeerScoreModel		NetworkingRoadmapTask	= "add_peer_score_model"
	RoadmapTaskAdaptiveFanoutConfiguration	NetworkingRoadmapTask	= "add_adaptive_fanout_configuration"
	RoadmapTaskQoSPolicy			NetworkingRoadmapTask	= "add_qos_policy"
	RoadmapTaskServiceTrafficIsolation	NetworkingRoadmapTask	= "add_service_traffic_isolation"
	RoadmapTaskNodeIDDerivation		NetworkingRoadmapTask	= "define_node_id_derivation"
	RoadmapTaskNodeRecord			NetworkingRoadmapTask	= "define_node_record"
	RoadmapTaskSignedNodeAdvertisements	NetworkingRoadmapTask	= "implement_signed_node_advertisements"
	RoadmapTaskSessionHandshake		NetworkingRoadmapTask	= "implement_session_handshake"
	RoadmapTaskMultiplexedStreams		NetworkingRoadmapTask	= "implement_multiplexed_streams"
	RoadmapTaskSessionKeyRotation		NetworkingRoadmapTask	= "implement_session_key_rotation"
	RoadmapTaskOverlayDescriptor		NetworkingRoadmapTask	= "define_overlay_descriptor"
	RoadmapTaskOverlayMembership		NetworkingRoadmapTask	= "implement_overlay_membership"
	RoadmapTaskOverlayPeerSets		NetworkingRoadmapTask	= "implement_overlay_peer_sets"
	RoadmapTaskRouteGraph			NetworkingRoadmapTask	= "implement_route_graph"
	RoadmapTaskRoutingTableCommitment	NetworkingRoadmapTask	= "implement_routing_table_commitment"
	RoadmapTaskZoneServiceOverlays		NetworkingRoadmapTask	= "add_zone_and_service_overlays"
	RoadmapTaskTransferOfferProtocol	NetworkingRoadmapTask	= "implement_transfer_offer_protocol"
	RoadmapTaskChunkDescriptors		NetworkingRoadmapTask	= "implement_chunk_descriptors"
	RoadmapTaskChunkMerkleVerification	NetworkingRoadmapTask	= "implement_chunk_merkle_verification"
	RoadmapTaskResumableTransfer		NetworkingRoadmapTask	= "implement_resumable_transfer"
	RoadmapTaskAdaptiveChunkSizingRL2	NetworkingRoadmapTask	= "implement_adaptive_chunk_sizing"
	RoadmapTaskRL2Backpressure		NetworkingRoadmapTask	= "implement_backpressure"
	RoadmapTaskDRT				NetworkingRoadmapTask	= "implement_drt"
	RoadmapTaskDiscoveryRecords		NetworkingRoadmapTask	= "add_discovery_records"
	RoadmapTaskLeaseRenewal			NetworkingRoadmapTask	= "add_lease_renewal"
	RoadmapTaskProofAttachedLookupResponse	NetworkingRoadmapTask	= "add_proof_attached_lookup_responses"
	RoadmapTaskServiceZoneIndexes		NetworkingRoadmapTask	= "add_service_and_zone_indexes"
	RoadmapTaskSignedAdvertisementValid	NetworkingRoadmapTask	= "add_signed_advertisement_validation"
	RoadmapTaskTreeBroadcast		NetworkingRoadmapTask	= "implement_tree_broadcast"
	RoadmapTaskGossipFallback		NetworkingRoadmapTask	= "implement_gossip_fallback"
	RoadmapTaskHashDeduplication		NetworkingRoadmapTask	= "implement_hash_deduplication"
	RoadmapTaskHeaderFirstBlockPropagation	NetworkingRoadmapTask	= "implement_header_first_block_propagation"
	RoadmapTaskParallelChunkFetch		NetworkingRoadmapTask	= "implement_parallel_chunk_fetch"
	RoadmapTaskApplicationMessageEnvelope	NetworkingRoadmapTask	= "implement_application_network_message_envelope"
	RoadmapTaskExecutionZoneMessageRouting	NetworkingRoadmapTask	= "add_execution_zone_message_routing"
	RoadmapTaskCrossZoneSequenceHandling	NetworkingRoadmapTask	= "add_cross_zone_sequence_handling"
	RoadmapTaskReceiptDeliveryProtocol	NetworkingRoadmapTask	= "add_receipt_delivery_protocol"
	RoadmapTaskQueryResponseProofAttach	NetworkingRoadmapTask	= "add_query_response_proof_attachment"
	RoadmapTaskServiceTrafficPath		NetworkingRoadmapTask	= "add_service_network_traffic_path"
	RoadmapTaskPeerReputationHardening	NetworkingRoadmapTask	= "add_peer_reputation_hardening"
	RoadmapTaskEclipseResistanceTests	NetworkingRoadmapTask	= "add_eclipse_resistance_tests"
	RoadmapTaskSpamFloodSimulations		NetworkingRoadmapTask	= "add_spam_flood_simulations"
	RoadmapTaskRoutingManipulationSims	NetworkingRoadmapTask	= "add_routing_manipulation_simulations"
	RoadmapTaskBandwidthExhaustionTests	NetworkingRoadmapTask	= "add_bandwidth_exhaustion_tests"
	RoadmapTaskChunkCorruptionTests		NetworkingRoadmapTask	= "add_chunk_corruption_tests"
)

type NetworkingExitCriterion string

const (
	ExitCurrentBehaviorMeasurable		NetworkingExitCriterion	= "current_network_behavior_is_measurable"
	ExitBaselineMetricsExist		NetworkingExitCriterion	= "baseline_propagation_and_peer_quality_metrics_exist"
	ExitConsensusProtectedPriority		NetworkingExitCriterion	= "consensus_traffic_has_protected_priority"
	ExitPeerScoringChannelMetrics		NetworkingExitCriterion	= "peer_scoring_and_channel_metrics_are_available"
	ExitServiceCannotStarveConsensus	NetworkingExitCriterion	= "service_traffic_cannot_starve_consensus"
	ExitCryptographicNodeAuth		NetworkingExitCriterion	= "nodes_authenticate_through_cryptographic_identity"
	ExitLogicalStreamsShareSession		NetworkingExitCriterion	= "logical_streams_share_one_peer_session"
	ExitExpiredForgedRecordsRejected	NetworkingExitCriterion	= "expired_or_forged_node_records_are_rejected"
	ExitOverlayJoinSupported		NetworkingExitCriterion	= "nodes_can_join_validator_zone_service_data_and_discovery_overlays"
	ExitCommittedRoutesReproducible		NetworkingExitCriterion	= "overlay_routing_decisions_are_reproducible_when_committed"
	ExitPeerRotationConnectivity		NetworkingExitCriterion	= "peer_rotation_preserves_connectivity"
	ExitChunkedStreamingPayloads		NetworkingExitCriterion	= "blocks_state_snapshots_and_proof_bundles_stream_in_chunks"
	ExitInterruptedTransfersResume		NetworkingExitCriterion	= "interrupted_transfers_can_resume"
	ExitInvalidChunksRejected		NetworkingExitCriterion	= "invalid_chunks_are_rejected"
	ExitDiscoveryObjectsDiscoverable	NetworkingExitCriterion	= "nodes_zones_services_endpoints_and_storage_providers_are_discoverable"
	ExitDiscoveryRecordsExpireVerify	NetworkingExitCriterion	= "discovery_records_expire_and_can_be_verified"
	ExitForgedExpiredRecordsRejected	NetworkingExitCriterion	= "forged_or_expired_discovery_records_are_rejected"
	ExitBlocksHeaderChunksProofSet		NetworkingExitCriterion	= "blocks_propagate_as_header_chunks_and_proof_set"
	ExitDuplicateConflictingHandled		NetworkingExitCriterion	= "duplicate_and_conflicting_broadcasts_are_handled"
	ExitFallbackGossipResilient		NetworkingExitCriterion	= "fallback_gossip_preserves_resilience"
	ExitL3MessageClassesSupported		NetworkingExitCriterion	= "l3_messages_support_execution_service_query_storage_and_cross_zone_classes"
	ExitCrossZoneDeliverySemantics		NetworkingExitCriterion	= "cross_zone_delivery_is_at_least_once_network_and_exactly_once_execution"
	ExitReceiptsVisibleProofQueryable	NetworkingExitCriterion	= "receipts_are_delivery_visible_and_proof_queryable_where_committed"
	ExitMaliciousPeersIsolated		NetworkingExitCriterion	= "malicious_peers_are_isolated_locally"
	ExitCriticalChannelsUnderFlood		NetworkingExitCriterion	= "critical_channels_remain_available_under_service_flood"
	ExitDiscoveryPoisoningDetected		NetworkingExitCriterion	= "discovery_poisoning_is_detected_by_signature_and_proof_checks"
)

type RoadmapTaskStatus string

const (
	RoadmapTaskPending	RoadmapTaskStatus	= "pending"
	RoadmapTaskComplete	RoadmapTaskStatus	= "complete"
)

type NetworkingRoadmapPhaseSpec struct {
	Phase		NetworkingRoadmapPhase
	Title		string
	Tasks		[]NetworkingRoadmapTask
	ExitCriteria	[]NetworkingExitCriterion
	DependsOn	[]NetworkingRoadmapPhase
}

type NetworkingRoadmapTaskEvidence struct {
	Task		NetworkingRoadmapTask
	Status		RoadmapTaskStatus
	Evidence	string
}

type NetworkingRoadmapEvidence struct {
	CometBFTInventory		AetherNetworkingAdapter
	PerformanceSnapshot		PerformanceMetricsSnapshot
	L0Schedule			L0Schedule
	XNetworkParams			XNetworkParams
	Session				SessionChannel
	NodeRecords			[]NodeRecord
	SignedDiscoveryRecords		[]DiscoveryRecord
	HandshakeReplayRejected		bool
	KeyRotationAvailable		bool
	OverlayDescriptors		[]OverlayDescriptor
	OverlayMemberships		[]OverlayMembershipRecord
	AdaptiveGraph			AdaptiveOverlayGraph
	RoutingGraph			RoutingGraph
	RoutingTableUse			RoutingTableUse
	PeerRotationPreserved		bool
	RL2Offer			RL2TransferOffer
	RL2ChunkDescriptors		[]RL2ChunkDescriptor
	RL2Session			RL2TransferSession
	RL2StreamingPlan		RL2StreamingPlan
	RL2PayloadTypes			[]RL2PayloadType
	RL2BackpressureSignal		RL2BackpressureSignal
	RL2InvalidChunkRejected		bool
	RL2InterruptedResumed		bool
	DiscoveryTable			DistributedRoutingTable
	DiscoveryResponse		DiscoveryResponse
	DiscoveryObjectTypes		[]DRTObjectType
	DiscoveryLeaseRenewed		bool
	DiscoveryForgedRejected		bool
	DiscoveryExpiredRejected	bool
	BroadcastMessage		BroadcastMessage
	BroadcastPlan			BroadcastPlan
	BroadcastDedupCache		BroadcastDedupCache
	BroadcastDuplicateHandled	bool
	BroadcastConflictHandled	bool
	BlockSession			BlockPropagationSession
	ParallelChunkPlan		StreamParallelFetchPlan
	GossipFallbackUsed		bool
	MeshMessages			[]AetherMeshMessage
	MeshDeliveries			[]AetherMeshDelivery
	CrossZoneTracker		CrossZoneSequenceTracker
	CrossZoneReceipt		CrossZoneReceipt
	ReceiptDelivery			ReceiptDelivery
	QueryResponseProof		QueryResponseProof
	L3Metrics			[]L3OverlayMetrics
	CrossZoneAtLeastOnce		bool
	CrossZoneExactlyOnce		bool
	SecurityPolicy			NetworkSecurityPolicy
	SecurityDecision		NetworkSecurityDecision
	ReputationDecision		PeerReputationDecision
	EclipsePlan			EclipseResistancePlan
	EclipseThreats			[]NetworkThreat
	SpamSimulation			SpamSimulationResult
	RoutingManipulation		RoutingManipulationSimulationResult
	BandwidthExhaustionDetected	bool
	ChunkCorruptionDetected		bool
	DiscoveryPoisoningDetected	bool
	CriticalChannelsAvailable	bool
}

type NetworkingRoadmapPhaseReport struct {
	Phase			NetworkingRoadmapPhase
	Tasks			[]NetworkingRoadmapTaskEvidence
	SatisfiedExitCriteria	[]NetworkingExitCriterion
	Ready			bool
	ReportHash		string
}

type NetworkingImplementationRoadmap struct {
	Phases		[]NetworkingRoadmapPhaseSpec
	RoadmapRoot	string
}

func DefaultNetworkingImplementationRoadmap() NetworkingImplementationRoadmap {
	roadmap := NetworkingImplementationRoadmap{
		Phases: []NetworkingRoadmapPhaseSpec{
			{
				Phase:	RoadmapPhaseBaselineInstrumentation,
				Title:	"Baseline and Instrumentation",
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
				Phase:	RoadmapPhaseAetherNetworkingAdapter,
				Title:	"Aether Networking Adapter",
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
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseBaselineInstrumentation},
			},
			{
				Phase:	RoadmapPhaseNodeIdentitySessions,
				Title:	"Node Identity and Sessions",
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
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseAetherNetworkingAdapter},
			},
			{
				Phase:	RoadmapPhaseOverlayRouting,
				Title:	"Overlay Routing",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskOverlayDescriptor,
					RoadmapTaskOverlayMembership,
					RoadmapTaskOverlayPeerSets,
					RoadmapTaskRouteGraph,
					RoadmapTaskRoutingTableCommitment,
					RoadmapTaskZoneServiceOverlays,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitOverlayJoinSupported,
					ExitCommittedRoutesReproducible,
					ExitPeerRotationConnectivity,
				},
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseNodeIdentitySessions},
			},
			{
				Phase:	RoadmapPhaseRL2Streaming,
				Title:	"RL2 Streaming",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskTransferOfferProtocol,
					RoadmapTaskChunkDescriptors,
					RoadmapTaskChunkMerkleVerification,
					RoadmapTaskResumableTransfer,
					RoadmapTaskAdaptiveChunkSizingRL2,
					RoadmapTaskRL2Backpressure,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitChunkedStreamingPayloads,
					ExitInterruptedTransfersResume,
					ExitInvalidChunksRejected,
				},
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseOverlayRouting},
			},
			{
				Phase:	RoadmapPhaseDiscoveryLayer,
				Title:	"Discovery Layer",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskDRT,
					RoadmapTaskDiscoveryRecords,
					RoadmapTaskLeaseRenewal,
					RoadmapTaskProofAttachedLookupResponse,
					RoadmapTaskServiceZoneIndexes,
					RoadmapTaskSignedAdvertisementValid,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitDiscoveryObjectsDiscoverable,
					ExitDiscoveryRecordsExpireVerify,
					ExitForgedExpiredRecordsRejected,
				},
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseRL2Streaming},
			},
			{
				Phase:	RoadmapPhaseHybridBroadcast,
				Title:	"Hybrid Broadcast",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskTreeBroadcast,
					RoadmapTaskGossipFallback,
					RoadmapTaskHashDeduplication,
					RoadmapTaskHeaderFirstBlockPropagation,
					RoadmapTaskParallelChunkFetch,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitBlocksHeaderChunksProofSet,
					ExitDuplicateConflictingHandled,
					ExitFallbackGossipResilient,
				},
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseDiscoveryLayer},
			},
			{
				Phase:	RoadmapPhaseAetherMesh,
				Title:	"Aether Mesh",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskApplicationMessageEnvelope,
					RoadmapTaskExecutionZoneMessageRouting,
					RoadmapTaskCrossZoneSequenceHandling,
					RoadmapTaskReceiptDeliveryProtocol,
					RoadmapTaskQueryResponseProofAttach,
					RoadmapTaskServiceTrafficPath,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitL3MessageClassesSupported,
					ExitCrossZoneDeliverySemantics,
					ExitReceiptsVisibleProofQueryable,
				},
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseHybridBroadcast},
			},
			{
				Phase:	RoadmapPhaseSecurityLoadHardening,
				Title:	"Security and Load Hardening",
				Tasks: []NetworkingRoadmapTask{
					RoadmapTaskPeerReputationHardening,
					RoadmapTaskEclipseResistanceTests,
					RoadmapTaskSpamFloodSimulations,
					RoadmapTaskRoutingManipulationSims,
					RoadmapTaskBandwidthExhaustionTests,
					RoadmapTaskChunkCorruptionTests,
				},
				ExitCriteria: []NetworkingExitCriterion{
					ExitMaliciousPeersIsolated,
					ExitCriticalChannelsUnderFlood,
					ExitDiscoveryPoisoningDetected,
				},
				DependsOn:	[]NetworkingRoadmapPhase{RoadmapPhaseAetherMesh},
			},
		},
	}
	roadmap.RoadmapRoot = ComputeNetworkingRoadmapRoot(roadmap)
	return roadmap
}

func ValidateNetworkingImplementationRoadmap(roadmap NetworkingImplementationRoadmap) error {
	roadmap = NormalizeNetworkingImplementationRoadmap(roadmap)
	if len(roadmap.Phases) != 9 {
		return errors.New("networking roadmap must define phases 0-8")
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
	for _, required := range []NetworkingRoadmapPhase{RoadmapPhaseBaselineInstrumentation, RoadmapPhaseAetherNetworkingAdapter, RoadmapPhaseNodeIdentitySessions, RoadmapPhaseOverlayRouting, RoadmapPhaseRL2Streaming, RoadmapPhaseDiscoveryLayer, RoadmapPhaseHybridBroadcast, RoadmapPhaseAetherMesh, RoadmapPhaseSecurityLoadHardening} {
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
	case RoadmapPhaseOverlayRouting:
		tasks, criteria, err = evaluatePhase3(evidence)
	case RoadmapPhaseRL2Streaming:
		tasks, criteria, err = evaluatePhase4(evidence)
	case RoadmapPhaseDiscoveryLayer:
		tasks, criteria, err = evaluatePhase5(evidence)
	case RoadmapPhaseHybridBroadcast:
		tasks, criteria, err = evaluatePhase6(evidence)
	case RoadmapPhaseAetherMesh:
		tasks, criteria, err = evaluatePhase7(evidence)
	case RoadmapPhaseSecurityLoadHardening:
		tasks, criteria, err = evaluatePhase8(evidence)
	}
	if err != nil {
		return NetworkingRoadmapPhaseReport{}, err
	}
	report := NetworkingRoadmapPhaseReport{
		Phase:			phase,
		Tasks:			tasks,
		SatisfiedExitCriteria:	criteria,
		Ready:			allRoadmapTasksComplete(tasks) && allExitCriteriaSatisfied(phase, criteria),
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
	case RoadmapPhaseBaselineInstrumentation, RoadmapPhaseAetherNetworkingAdapter, RoadmapPhaseNodeIdentitySessions, RoadmapPhaseOverlayRouting, RoadmapPhaseRL2Streaming, RoadmapPhaseDiscoveryLayer, RoadmapPhaseHybridBroadcast, RoadmapPhaseAetherMesh, RoadmapPhaseSecurityLoadHardening:
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
		RoadmapTaskSessionKeyRotation,
		RoadmapTaskOverlayDescriptor,
		RoadmapTaskOverlayMembership,
		RoadmapTaskOverlayPeerSets,
		RoadmapTaskRouteGraph,
		RoadmapTaskRoutingTableCommitment,
		RoadmapTaskZoneServiceOverlays,
		RoadmapTaskTransferOfferProtocol,
		RoadmapTaskChunkDescriptors,
		RoadmapTaskChunkMerkleVerification,
		RoadmapTaskResumableTransfer,
		RoadmapTaskAdaptiveChunkSizingRL2,
		RoadmapTaskRL2Backpressure,
		RoadmapTaskDRT,
		RoadmapTaskDiscoveryRecords,
		RoadmapTaskLeaseRenewal,
		RoadmapTaskProofAttachedLookupResponse,
		RoadmapTaskServiceZoneIndexes,
		RoadmapTaskSignedAdvertisementValid,
		RoadmapTaskTreeBroadcast,
		RoadmapTaskGossipFallback,
		RoadmapTaskHashDeduplication,
		RoadmapTaskHeaderFirstBlockPropagation,
		RoadmapTaskParallelChunkFetch,
		RoadmapTaskApplicationMessageEnvelope,
		RoadmapTaskExecutionZoneMessageRouting,
		RoadmapTaskCrossZoneSequenceHandling,
		RoadmapTaskReceiptDeliveryProtocol,
		RoadmapTaskQueryResponseProofAttach,
		RoadmapTaskServiceTrafficPath,
		RoadmapTaskPeerReputationHardening,
		RoadmapTaskEclipseResistanceTests,
		RoadmapTaskSpamFloodSimulations,
		RoadmapTaskRoutingManipulationSims,
		RoadmapTaskBandwidthExhaustionTests,
		RoadmapTaskChunkCorruptionTests:
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
		ExitExpiredForgedRecordsRejected,
		ExitOverlayJoinSupported,
		ExitCommittedRoutesReproducible,
		ExitPeerRotationConnectivity,
		ExitChunkedStreamingPayloads,
		ExitInterruptedTransfersResume,
		ExitInvalidChunksRejected,
		ExitDiscoveryObjectsDiscoverable,
		ExitDiscoveryRecordsExpireVerify,
		ExitForgedExpiredRecordsRejected,
		ExitBlocksHeaderChunksProofSet,
		ExitDuplicateConflictingHandled,
		ExitFallbackGossipResilient,
		ExitL3MessageClassesSupported,
		ExitCrossZoneDeliverySemantics,
		ExitReceiptsVisibleProofQueryable,
		ExitMaliciousPeersIsolated,
		ExitCriticalChannelsUnderFlood,
		ExitDiscoveryPoisoningDetected:
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

func evaluatePhase3(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	descriptorsOK := ValidateOverlayDescriptors(evidence.OverlayDescriptors, 0) == nil
	membershipOK := roadmapOverlayMembershipsValid(evidence.OverlayMemberships)
	peerSetsOK := roadmapAdaptiveGraphValid(evidence.AdaptiveGraph, evidence.OverlayDescriptors)
	routeGraphOK := roadmapRoutingGraphValid(evidence.RoutingGraph, evidence.OverlayDescriptors)
	routingCommitmentOK := ValidateRoutingTableUse(evidence.RoutingTableUse) == nil && evidence.RoutingTableUse.Committed
	zoneServiceOK := roadmapHasOverlayTypes(evidence.OverlayDescriptors, OverlayTypeZone, OverlayTypeService)
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskOverlayDescriptor, descriptorsOK, "overlay descriptors validate"),
		taskEvidence(RoadmapTaskOverlayMembership, membershipOK, "overlay membership records validate"),
		taskEvidence(RoadmapTaskOverlayPeerSets, peerSetsOK, "adaptive peer sets validate"),
		taskEvidence(RoadmapTaskRouteGraph, routeGraphOK, "routing graph validates"),
		taskEvidence(RoadmapTaskRoutingTableCommitment, routingCommitmentOK, "routing table commitment validates"),
		taskEvidence(RoadmapTaskZoneServiceOverlays, zoneServiceOK, "zone and service overlays present"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if roadmapMembershipsCoverOverlayTypes(evidence.OverlayDescriptors, evidence.OverlayMemberships, OverlayTypeValidator, OverlayTypeZone, OverlayTypeService, OverlayTypeData, OverlayTypeDiscovery) && membershipOK {
		criteria = append(criteria, ExitOverlayJoinSupported)
	}
	if routingCommitmentOK && evidence.RoutingTableUse.UsedForExecutionScheduling {
		criteria = append(criteria, ExitCommittedRoutesReproducible)
	}
	if peerSetsOK && evidence.PeerRotationPreserved {
		criteria = append(criteria, ExitPeerRotationConnectivity)
	}
	return tasks, criteria, nil
}

func evaluatePhase4(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	offerOK := evidence.RL2Offer.Validate(0) == nil
	descriptorsOK := offerOK && ValidateRL2ChunkDescriptors(evidence.RL2Offer.Transfer, evidence.RL2ChunkDescriptors) == nil
	merkleOK := descriptorsOK && roadmapRL2DescriptorsHaveProofs(evidence.RL2ChunkDescriptors)
	resumeOK := evidence.RL2InterruptedResumed && evidence.RL2Session.Validate() == nil && evidence.RL2Session.Acceptance.ResumeToken != ""
	adaptiveOK := offerOK && evidence.RL2Offer.SuggestedChunkSize > 0 && evidence.RL2Offer.SuggestedChunkSize <= evidence.RL2Offer.Transfer.ChunkSize
	backpressureOK := offerOK && evidence.RL2BackpressureSignal.TransferID == evidence.RL2Offer.Transfer.TransferID && evidence.RL2BackpressureSignal.ResumeToken != ""
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskTransferOfferProtocol, offerOK, "RL2 transfer offer validates"),
		taskEvidence(RoadmapTaskChunkDescriptors, descriptorsOK, "RL2 chunk descriptors validate"),
		taskEvidence(RoadmapTaskChunkMerkleVerification, merkleOK, "RL2 Merkle proof paths present"),
		taskEvidence(RoadmapTaskResumableTransfer, resumeOK, "RL2 resume token/session validate"),
		taskEvidence(RoadmapTaskAdaptiveChunkSizingRL2, adaptiveOK, "RL2 suggested chunk size validates"),
		taskEvidence(RoadmapTaskRL2Backpressure, backpressureOK, "RL2 backpressure signal validates"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if offerOK && descriptorsOK && roadmapHasRL2PayloadTypes(evidence.RL2PayloadTypes, RL2PayloadLargeBlock, RL2PayloadStateSyncStream, RL2PayloadProofSet) {
		criteria = append(criteria, ExitChunkedStreamingPayloads)
	}
	if resumeOK {
		criteria = append(criteria, ExitInterruptedTransfersResume)
	}
	if evidence.RL2InvalidChunkRejected {
		criteria = append(criteria, ExitInvalidChunksRejected)
	}
	return tasks, criteria, nil
}

func evaluatePhase5(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	tableOK := evidence.DiscoveryTable.Validate(nil, 0) == nil && (len(evidence.DiscoveryTable.Records) > 0 || len(evidence.DiscoveryTable.Advertisements) > 0)
	recordsOK := len(evidence.DiscoveryTable.Records) > 0 || len(evidence.SignedDiscoveryRecords) > 0
	responseOK := roadmapDiscoveryResponseHasProof(evidence.DiscoveryResponse)
	indexOK := roadmapHasDiscoveryObjectTypes(evidence.DiscoveryObjectTypes, DRTObjectNode, DRTObjectExecutionZone, DRTObjectServiceEndpoint, DRTObjectRPCEndpoint, DRTObjectStorageProvider)
	signedValidationOK := evidence.DiscoveryForgedRejected && evidence.DiscoveryExpiredRejected
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskDRT, tableOK, "distributed routing table validates"),
		taskEvidence(RoadmapTaskDiscoveryRecords, recordsOK, "signed discovery records available"),
		taskEvidence(RoadmapTaskLeaseRenewal, evidence.DiscoveryLeaseRenewed, "lease renewal accepted"),
		taskEvidence(RoadmapTaskProofAttachedLookupResponse, responseOK, "proof-attached discovery response validates structurally"),
		taskEvidence(RoadmapTaskServiceZoneIndexes, indexOK, "service, zone, endpoint, storage, and node indexes covered"),
		taskEvidence(RoadmapTaskSignedAdvertisementValid, signedValidationOK, "forged and expired advertisements rejected"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if tableOK && indexOK {
		criteria = append(criteria, ExitDiscoveryObjectsDiscoverable)
	}
	if recordsOK && evidence.DiscoveryLeaseRenewed && responseOK {
		criteria = append(criteria, ExitDiscoveryRecordsExpireVerify)
	}
	if signedValidationOK {
		criteria = append(criteria, ExitForgedExpiredRecordsRejected)
	}
	return tasks, criteria, nil
}

func evaluatePhase6(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	treeOK := len(evidence.BroadcastPlan.TreeTargets) > 0
	gossipOK := evidence.GossipFallbackUsed && evidence.BroadcastPlan.FallbackUsed && len(evidence.BroadcastPlan.GossipTargets) > 0
	dedupOK := len(evidence.BroadcastDedupCache.Entries) > 0 && evidence.BroadcastDuplicateHandled && evidence.BroadcastConflictHandled
	headerFirstOK := ValidateHeaderFirstPerformance(evidence.BlockSession) == nil
	parallelOK := ValidateParallelChunkPerformance(evidence.ParallelChunkPlan) == nil
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskTreeBroadcast, treeOK, "tree broadcast targets selected"),
		taskEvidence(RoadmapTaskGossipFallback, gossipOK, "gossip fallback targets selected"),
		taskEvidence(RoadmapTaskHashDeduplication, dedupOK, "hash dedup cache handles duplicate and conflict"),
		taskEvidence(RoadmapTaskHeaderFirstBlockPropagation, headerFirstOK, "header-first block propagation validates"),
		taskEvidence(RoadmapTaskParallelChunkFetch, parallelOK, "parallel chunk fetch plan validates"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if headerFirstOK && parallelOK && len(evidence.BlockSession.ProofSet.ProofHashes) > 0 {
		criteria = append(criteria, ExitBlocksHeaderChunksProofSet)
	}
	if dedupOK && len(evidence.BroadcastDedupCache.Faults) > 0 {
		criteria = append(criteria, ExitDuplicateConflictingHandled)
	}
	if gossipOK {
		criteria = append(criteria, ExitFallbackGossipResilient)
	}
	return tasks, criteria, nil
}

func evaluatePhase7(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	envelopeOK := roadmapMeshMessagesValid(evidence.MeshMessages)
	executionRoutingOK := roadmapHasMeshDelivery(evidence.MeshDeliveries, MeshMessageExecution, ChannelExecution) || roadmapHasMeshDelivery(evidence.MeshDeliveries, MeshMessageCrossZone, ChannelExecution)
	crossZoneOK := evidence.CrossZoneAtLeastOnce && evidence.CrossZoneExactlyOnce && len(evidence.CrossZoneTracker.States) > 0
	receiptOK := evidence.CrossZoneReceipt.Validate() == nil && evidence.ReceiptDelivery.State == ReceiptDeliveryAcknowledged
	queryProofOK := evidence.QueryResponseProof.Validate() == nil
	servicePathOK := roadmapHasMeshDelivery(evidence.MeshDeliveries, MeshMessageService, ChannelService)
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskApplicationMessageEnvelope, envelopeOK, "application mesh message envelopes validate"),
		taskEvidence(RoadmapTaskExecutionZoneMessageRouting, executionRoutingOK, "execution zone messages route through execution channel"),
		taskEvidence(RoadmapTaskCrossZoneSequenceHandling, crossZoneOK, "cross-zone sequence tracker rejects replay and preserves delivery intent"),
		taskEvidence(RoadmapTaskReceiptDeliveryProtocol, receiptOK, "receipt delivery is acknowledged and visible"),
		taskEvidence(RoadmapTaskQueryResponseProofAttach, queryProofOK, "query response proof validates"),
		taskEvidence(RoadmapTaskServiceTrafficPath, servicePathOK, "service traffic routes through service channel"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if envelopeOK && roadmapHasMeshMessageTypes(evidence.MeshMessages, MeshMessageExecution, MeshMessageService, MeshMessageQuery, MeshMessageStorage, MeshMessageCrossZone) {
		criteria = append(criteria, ExitL3MessageClassesSupported)
	}
	if crossZoneOK {
		criteria = append(criteria, ExitCrossZoneDeliverySemantics)
	}
	if receiptOK && evidence.CrossZoneReceipt.ProofQueryable && evidence.CrossZoneReceipt.ProofHash != "" {
		criteria = append(criteria, ExitReceiptsVisibleProofQueryable)
	}
	return tasks, criteria, nil
}

func evaluatePhase8(evidence NetworkingRoadmapEvidence) ([]NetworkingRoadmapTaskEvidence, []NetworkingExitCriterion, error) {
	reputationOK := evidence.SecurityPolicy.Validate() == nil && evidence.ReputationDecision.PeerNodeID != "" && evidence.SecurityDecision.PeerNodeID != ""
	eclipseOK := ValidateEclipseResistancePlan(evidence.EclipsePlan) == nil || containsNetworkThreat(evidence.EclipseThreats, ThreatEclipseAttack)
	spamOK := containsNetworkThreat(evidence.SpamSimulation.Threats, ThreatSpamFlood)
	routingOK := evidence.RoutingManipulation.FaultsDetected > 0 && containsNetworkThreat(evidence.RoutingManipulation.Threats, ThreatRoutingManipulation)
	bandwidthOK := evidence.BandwidthExhaustionDetected && containsNetworkThreat(evidence.SpamSimulation.Threats, ThreatBandwidthExhaustion)
	chunkOK := evidence.ChunkCorruptionDetected && containsNetworkThreat(evidence.SecurityDecision.Threats, ThreatChunkCorruption)
	tasks := []NetworkingRoadmapTaskEvidence{
		taskEvidence(RoadmapTaskPeerReputationHardening, reputationOK, "peer reputation and security decision available"),
		taskEvidence(RoadmapTaskEclipseResistanceTests, eclipseOK, "eclipse resistance plan or simulation detects risk"),
		taskEvidence(RoadmapTaskSpamFloodSimulations, spamOK, "spam flood simulation detects spam"),
		taskEvidence(RoadmapTaskRoutingManipulationSims, routingOK, "routing manipulation simulation detects conflicts"),
		taskEvidence(RoadmapTaskBandwidthExhaustionTests, bandwidthOK, "bandwidth exhaustion is detected"),
		taskEvidence(RoadmapTaskChunkCorruptionTests, chunkOK, "chunk corruption is detected"),
	}
	criteria := make([]NetworkingExitCriterion, 0, 3)
	if reputationOK && evidence.SecurityDecision.ConsensusIsolated && (evidence.SecurityDecision.Quarantine || evidence.SecurityDecision.RotatePeer || evidence.SecurityDecision.DropMessage) {
		criteria = append(criteria, ExitMaliciousPeersIsolated)
	}
	if evidence.CriticalChannelsAvailable && ValidateServiceTrafficIsolationFromMetrics(evidence.L0Schedule.Metrics) == nil {
		criteria = append(criteria, ExitCriticalChannelsUnderFlood)
	}
	if evidence.DiscoveryPoisoningDetected && (evidence.DiscoveryForgedRejected || containsNetworkThreat(evidence.SecurityDecision.Threats, ThreatServiceAdvertisementForge) || containsNetworkThreat(evidence.SecurityDecision.Threats, ThreatDiscoveryPoisoning)) {
		criteria = append(criteria, ExitDiscoveryPoisoningDetected)
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
	case RoadmapPhaseOverlayRouting:
		return []NetworkingRoadmapTask{
				RoadmapTaskOverlayDescriptor,
				RoadmapTaskOverlayMembership,
				RoadmapTaskOverlayPeerSets,
				RoadmapTaskRouteGraph,
				RoadmapTaskRoutingTableCommitment,
				RoadmapTaskZoneServiceOverlays,
			}, []NetworkingExitCriterion{
				ExitOverlayJoinSupported,
				ExitCommittedRoutesReproducible,
				ExitPeerRotationConnectivity,
			}
	case RoadmapPhaseRL2Streaming:
		return []NetworkingRoadmapTask{
				RoadmapTaskTransferOfferProtocol,
				RoadmapTaskChunkDescriptors,
				RoadmapTaskChunkMerkleVerification,
				RoadmapTaskResumableTransfer,
				RoadmapTaskAdaptiveChunkSizingRL2,
				RoadmapTaskRL2Backpressure,
			}, []NetworkingExitCriterion{
				ExitChunkedStreamingPayloads,
				ExitInterruptedTransfersResume,
				ExitInvalidChunksRejected,
			}
	case RoadmapPhaseDiscoveryLayer:
		return []NetworkingRoadmapTask{
				RoadmapTaskDRT,
				RoadmapTaskDiscoveryRecords,
				RoadmapTaskLeaseRenewal,
				RoadmapTaskProofAttachedLookupResponse,
				RoadmapTaskServiceZoneIndexes,
				RoadmapTaskSignedAdvertisementValid,
			}, []NetworkingExitCriterion{
				ExitDiscoveryObjectsDiscoverable,
				ExitDiscoveryRecordsExpireVerify,
				ExitForgedExpiredRecordsRejected,
			}
	case RoadmapPhaseHybridBroadcast:
		return []NetworkingRoadmapTask{
				RoadmapTaskTreeBroadcast,
				RoadmapTaskGossipFallback,
				RoadmapTaskHashDeduplication,
				RoadmapTaskHeaderFirstBlockPropagation,
				RoadmapTaskParallelChunkFetch,
			}, []NetworkingExitCriterion{
				ExitBlocksHeaderChunksProofSet,
				ExitDuplicateConflictingHandled,
				ExitFallbackGossipResilient,
			}
	case RoadmapPhaseAetherMesh:
		return []NetworkingRoadmapTask{
				RoadmapTaskApplicationMessageEnvelope,
				RoadmapTaskExecutionZoneMessageRouting,
				RoadmapTaskCrossZoneSequenceHandling,
				RoadmapTaskReceiptDeliveryProtocol,
				RoadmapTaskQueryResponseProofAttach,
				RoadmapTaskServiceTrafficPath,
			}, []NetworkingExitCriterion{
				ExitL3MessageClassesSupported,
				ExitCrossZoneDeliverySemantics,
				ExitReceiptsVisibleProofQueryable,
			}
	case RoadmapPhaseSecurityLoadHardening:
		return []NetworkingRoadmapTask{
				RoadmapTaskPeerReputationHardening,
				RoadmapTaskEclipseResistanceTests,
				RoadmapTaskSpamFloodSimulations,
				RoadmapTaskRoutingManipulationSims,
				RoadmapTaskBandwidthExhaustionTests,
				RoadmapTaskChunkCorruptionTests,
			}, []NetworkingExitCriterion{
				ExitMaliciousPeersIsolated,
				ExitCriticalChannelsUnderFlood,
				ExitDiscoveryPoisoningDetected,
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

func roadmapOverlayMembershipsValid(records []OverlayMembershipRecord) bool {
	if len(records) == 0 {
		return false
	}
	for _, record := range records {
		if err := record.Validate(0); err != nil {
			return false
		}
	}
	return true
}

func roadmapRoutingGraphValid(graph RoutingGraph, descriptors []OverlayDescriptor) bool {
	graph = NormalizeRoutingGraph(graph)
	for _, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if desc.OverlayID == graph.OverlayID {
			return graph.Validate(desc) == nil
		}
	}
	return false
}

func roadmapAdaptiveGraphValid(graph AdaptiveOverlayGraph, descriptors []OverlayDescriptor) bool {
	graph = NormalizeAdaptiveOverlayGraph(graph)
	for _, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if desc.OverlayID == graph.OverlayID {
			return graph.Validate(desc) == nil
		}
	}
	return false
}

func roadmapHasOverlayTypes(descriptors []OverlayDescriptor, required ...OverlayType) bool {
	seen := make(map[OverlayType]struct{}, len(descriptors))
	for _, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		seen[desc.OverlayType] = struct{}{}
	}
	for _, overlayType := range required {
		if _, found := seen[overlayType]; !found {
			return false
		}
	}
	return true
}

func roadmapMembershipsCoverOverlayTypes(descriptors []OverlayDescriptor, memberships []OverlayMembershipRecord, required ...OverlayType) bool {
	overlayTypes := make(map[string]OverlayType, len(descriptors))
	for _, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		overlayTypes[desc.OverlayID] = desc.OverlayType
	}
	covered := make(map[OverlayType]struct{}, len(required))
	for _, membership := range memberships {
		membership = NormalizeOverlayMembershipRecord(membership)
		if overlayType, found := overlayTypes[membership.OverlayID]; found {
			covered[overlayType] = struct{}{}
		}
	}
	for _, overlayType := range required {
		if _, found := covered[overlayType]; !found {
			return false
		}
	}
	return true
}

func roadmapRL2DescriptorsHaveProofs(descriptors []RL2ChunkDescriptor) bool {
	if len(descriptors) == 0 {
		return false
	}
	for _, descriptor := range descriptors {
		if len(descriptor.ProofPath) == 0 {
			return false
		}
	}
	return true
}

func roadmapHasRL2PayloadTypes(payloadTypes []RL2PayloadType, required ...RL2PayloadType) bool {
	seen := make(map[RL2PayloadType]struct{}, len(payloadTypes))
	for _, payloadType := range payloadTypes {
		seen[payloadType] = struct{}{}
	}
	for _, payloadType := range required {
		if _, found := seen[payloadType]; !found {
			return false
		}
	}
	return true
}

func roadmapDiscoveryResponseHasProof(response DiscoveryResponse) bool {
	response = NormalizeDiscoveryResponse(response)
	if len(response.MatchedRecords) == 0 || response.AdvisoryOnly {
		return false
	}
	if response.ResponseID == "" || response.ResultHash == "" || len(response.SourceSignature) == 0 {
		return false
	}
	if response.OnChainProof.ProofHash == "" || response.OnChainProof.StateRoot == "" || response.OnChainProof.ProofHeight == 0 {
		return false
	}
	return response.ResponseID == ComputeDiscoveryResponseID(response)
}

func roadmapHasDiscoveryObjectTypes(objectTypes []DRTObjectType, required ...DRTObjectType) bool {
	seen := make(map[DRTObjectType]struct{}, len(objectTypes))
	for _, objectType := range objectTypes {
		objectType = DRTObjectType(strings.ToLower(strings.TrimSpace(string(objectType))))
		if objectType != "" {
			seen[objectType] = struct{}{}
		}
	}
	for _, objectType := range required {
		if _, found := seen[objectType]; !found {
			return false
		}
	}
	return true
}

func roadmapMeshMessagesValid(messages []AetherMeshMessage) bool {
	if len(messages) == 0 {
		return false
	}
	for _, msg := range messages {
		if err := msg.ValidateBasic(0); err != nil {
			return false
		}
	}
	return true
}

func roadmapHasMeshMessageTypes(messages []AetherMeshMessage, required ...AetherMeshMessageType) bool {
	seen := make(map[AetherMeshMessageType]struct{}, len(messages))
	for _, msg := range messages {
		msg = NormalizeAetherMeshMessage(msg)
		seen[msg.Type] = struct{}{}
	}
	for _, messageType := range required {
		if _, found := seen[messageType]; !found {
			return false
		}
	}
	return true
}

func roadmapHasMeshDelivery(deliveries []AetherMeshDelivery, messageType AetherMeshMessageType, channel ChannelClass) bool {
	for _, delivery := range deliveries {
		if NormalizeAetherMeshMessage(delivery.Message).Type == messageType && delivery.Channel == channel && len(delivery.Route.TargetNodeIDs) > 0 {
			return true
		}
	}
	return false
}

func containsNetworkThreat(threats []NetworkThreat, needle NetworkThreat) bool {
	for _, threat := range threats {
		if threat == needle {
			return true
		}
	}
	return false
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
			Task:		NetworkingRoadmapTask(strings.ToLower(strings.TrimSpace(string(value.Task)))),
			Status:		RoadmapTaskStatus(strings.ToLower(strings.TrimSpace(string(value.Status)))),
			Evidence:	strings.TrimSpace(value.Evidence),
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
