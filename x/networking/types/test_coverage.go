package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingTestCoverageCategory string

const (
	NetworkingTestCoverageUnit		NetworkingTestCoverageCategory	= "unit"
	NetworkingTestCoverageIntegration	NetworkingTestCoverageCategory	= "integration"
	NetworkingTestCoverageSecurity		NetworkingTestCoverageCategory	= "security"
	NetworkingTestCoveragePerformance	NetworkingTestCoverageCategory	= "performance"
)

type NetworkingRequiredTest string

const (
	RequiredTestNodeIDDerivation		NetworkingRequiredTest	= "node_id_derivation"
	RequiredTestNodeRecordSignature		NetworkingRequiredTest	= "node_record_signature_verification"
	RequiredTestSessionHandshake		NetworkingRequiredTest	= "session_handshake_validation"
	RequiredTestStreamPriority		NetworkingRequiredTest	= "stream_priority_classification"
	RequiredTestOverlayMembership		NetworkingRequiredTest	= "overlay_membership_validation"
	RequiredTestRouteCost			NetworkingRequiredTest	= "route_cost_calculation"
	RequiredTestNetworkMessageID		NetworkingRequiredTest	= "network_message_id_derivation"
	RequiredTestDiscoveryRecordExpiry	NetworkingRequiredTest	= "discovery_record_expiry"
	RequiredTestChunkHashVerification	NetworkingRequiredTest	= "chunk_hash_verification"
	RequiredTestBroadcastDeduplication	NetworkingRequiredTest	= "broadcast_deduplication"
	RequiredTestCometBFTANAConsensus	NetworkingRequiredTest	= "cometbft_consensus_traffic_with_ana_enabled"
	RequiredTestMultiplexedSessionStreams	NetworkingRequiredTest	= "multiplexed_streams_over_one_session"
	RequiredTestZoneOverlayFormation	NetworkingRequiredTest	= "zone_overlay_formation"
	RequiredTestServiceOverlayFormation	NetworkingRequiredTest	= "service_overlay_formation"
	RequiredTestCrossZoneDelivery		NetworkingRequiredTest	= "cross_zone_message_delivery"
	RequiredTestRL2BlockChunkTransfer	NetworkingRequiredTest	= "rl2_block_chunk_transfer"
	RequiredTestResumableStateSnapshot	NetworkingRequiredTest	= "resumable_state_snapshot_transfer"
	RequiredTestDiscoveryProofLookup	NetworkingRequiredTest	= "discovery_lookup_with_proof_attached_response"
	RequiredTestHeaderFirstPropagation	NetworkingRequiredTest	= "header_first_block_propagation"
	RequiredTestReplayedHandshake		NetworkingRequiredTest	= "replayed_handshake"
	RequiredTestForgedNodeAdvertisement	NetworkingRequiredTest	= "forged_node_advertisement"
	RequiredTestExpiredDiscoverySecurity	NetworkingRequiredTest	= "expired_discovery_record"
	RequiredTestConflictingBroadcast	NetworkingRequiredTest	= "conflicting_broadcast_payload"
	RequiredTestInvalidChunkSecurity	NetworkingRequiredTest	= "invalid_chunk"
	RequiredTestEclipsePeerSetSimulation	NetworkingRequiredTest	= "eclipse_peer_set_simulation"
	RequiredTestServiceSpamFlood		NetworkingRequiredTest	= "spam_flood_on_service_channel"
	RequiredTestConsensusUnderBulkLoad	NetworkingRequiredTest	= "consensus_traffic_under_bulk_data_load"
	RequiredTestCrossZoneReplaySecurity	NetworkingRequiredTest	= "cross_zone_message_replay"
	RequiredTestBlockHeaderLatency		NetworkingRequiredTest	= "block_header_propagation_latency"
	RequiredTestBlockReconstructionTime	NetworkingRequiredTest	= "block_reconstruction_time"
	RequiredTestChunkStreamingThroughput	NetworkingRequiredTest	= "chunk_streaming_throughput"
	RequiredTestDiscoveryQueryLatency	NetworkingRequiredTest	= "discovery_query_latency"
	RequiredTestOverlayJoinLatency		NetworkingRequiredTest	= "overlay_join_latency"
	RequiredTestCrossZonePropagation	NetworkingRequiredTest	= "cross_zone_message_propagation_latency"
	RequiredTestServiceTrafficThroughput	NetworkingRequiredTest	= "service_traffic_throughput"
	RequiredTestConsensusMixedLoadLatency	NetworkingRequiredTest	= "consensus_traffic_latency_under_mixed_load"
	RequiredTestPeerRotationStability	NetworkingRequiredTest	= "peer_rotation_stability"
)

type NetworkingTestCoverageSpec struct {
	Test		NetworkingRequiredTest
	Category	NetworkingTestCoverageCategory
	Title		string
}

type NetworkingTestCoverageEvidence struct {
	Test		NetworkingRequiredTest
	Category	NetworkingTestCoverageCategory
	TestNames	[]string
	Passed		bool
}

type NetworkingTestCoverageReport struct {
	Required	[]NetworkingTestCoverageSpec
	Evidence	[]NetworkingTestCoverageEvidence
	Missing		[]NetworkingRequiredTest
	Failed		[]NetworkingRequiredTest
	Ready		bool
	ReportHash	string
}

func DefaultRequiredNetworkingTestCoverage() []NetworkingTestCoverageSpec {
	return []NetworkingTestCoverageSpec{
		{Test: RequiredTestNodeIDDerivation, Category: NetworkingTestCoverageUnit, Title: "NodeID derivation"},
		{Test: RequiredTestNodeRecordSignature, Category: NetworkingTestCoverageUnit, Title: "NodeRecord signature verification"},
		{Test: RequiredTestSessionHandshake, Category: NetworkingTestCoverageUnit, Title: "Session handshake validation"},
		{Test: RequiredTestStreamPriority, Category: NetworkingTestCoverageUnit, Title: "Stream priority classification"},
		{Test: RequiredTestOverlayMembership, Category: NetworkingTestCoverageUnit, Title: "Overlay membership validation"},
		{Test: RequiredTestRouteCost, Category: NetworkingTestCoverageUnit, Title: "Route cost calculation"},
		{Test: RequiredTestNetworkMessageID, Category: NetworkingTestCoverageUnit, Title: "NetworkMessage ID derivation"},
		{Test: RequiredTestDiscoveryRecordExpiry, Category: NetworkingTestCoverageUnit, Title: "DiscoveryRecord expiry"},
		{Test: RequiredTestChunkHashVerification, Category: NetworkingTestCoverageUnit, Title: "Chunk hash verification"},
		{Test: RequiredTestBroadcastDeduplication, Category: NetworkingTestCoverageUnit, Title: "Broadcast deduplication"},
		{Test: RequiredTestCometBFTANAConsensus, Category: NetworkingTestCoverageIntegration, Title: "CometBFT consensus traffic with ANA enabled"},
		{Test: RequiredTestMultiplexedSessionStreams, Category: NetworkingTestCoverageIntegration, Title: "Multiplexed streams over one session"},
		{Test: RequiredTestZoneOverlayFormation, Category: NetworkingTestCoverageIntegration, Title: "Zone overlay formation"},
		{Test: RequiredTestServiceOverlayFormation, Category: NetworkingTestCoverageIntegration, Title: "Service overlay formation"},
		{Test: RequiredTestCrossZoneDelivery, Category: NetworkingTestCoverageIntegration, Title: "Cross-zone message delivery"},
		{Test: RequiredTestRL2BlockChunkTransfer, Category: NetworkingTestCoverageIntegration, Title: "RL2 block chunk transfer"},
		{Test: RequiredTestResumableStateSnapshot, Category: NetworkingTestCoverageIntegration, Title: "Resumable state snapshot transfer"},
		{Test: RequiredTestDiscoveryProofLookup, Category: NetworkingTestCoverageIntegration, Title: "Discovery lookup with proof-attached response"},
		{Test: RequiredTestHeaderFirstPropagation, Category: NetworkingTestCoverageIntegration, Title: "Header-first block propagation"},
		{Test: RequiredTestReplayedHandshake, Category: NetworkingTestCoverageSecurity, Title: "Replayed handshake"},
		{Test: RequiredTestForgedNodeAdvertisement, Category: NetworkingTestCoverageSecurity, Title: "Forged node advertisement"},
		{Test: RequiredTestExpiredDiscoverySecurity, Category: NetworkingTestCoverageSecurity, Title: "Expired discovery record"},
		{Test: RequiredTestConflictingBroadcast, Category: NetworkingTestCoverageSecurity, Title: "Conflicting broadcast payload"},
		{Test: RequiredTestInvalidChunkSecurity, Category: NetworkingTestCoverageSecurity, Title: "Invalid chunk"},
		{Test: RequiredTestEclipsePeerSetSimulation, Category: NetworkingTestCoverageSecurity, Title: "Eclipse peer set simulation"},
		{Test: RequiredTestServiceSpamFlood, Category: NetworkingTestCoverageSecurity, Title: "Spam flood on service channel"},
		{Test: RequiredTestConsensusUnderBulkLoad, Category: NetworkingTestCoverageSecurity, Title: "Consensus traffic under bulk data load"},
		{Test: RequiredTestCrossZoneReplaySecurity, Category: NetworkingTestCoverageSecurity, Title: "Cross-zone message replay"},
		{Test: RequiredTestBlockHeaderLatency, Category: NetworkingTestCoveragePerformance, Title: "Block header propagation latency"},
		{Test: RequiredTestBlockReconstructionTime, Category: NetworkingTestCoveragePerformance, Title: "Block reconstruction time"},
		{Test: RequiredTestChunkStreamingThroughput, Category: NetworkingTestCoveragePerformance, Title: "Chunk streaming throughput"},
		{Test: RequiredTestDiscoveryQueryLatency, Category: NetworkingTestCoveragePerformance, Title: "Discovery query latency"},
		{Test: RequiredTestOverlayJoinLatency, Category: NetworkingTestCoveragePerformance, Title: "Overlay join latency"},
		{Test: RequiredTestCrossZonePropagation, Category: NetworkingTestCoveragePerformance, Title: "Cross-zone message propagation latency"},
		{Test: RequiredTestServiceTrafficThroughput, Category: NetworkingTestCoveragePerformance, Title: "Service traffic throughput"},
		{Test: RequiredTestConsensusMixedLoadLatency, Category: NetworkingTestCoveragePerformance, Title: "Consensus traffic latency under mixed load"},
		{Test: RequiredTestPeerRotationStability, Category: NetworkingTestCoveragePerformance, Title: "Peer rotation stability"},
	}
}

func ValidateRequiredNetworkingTestCoverage(specs []NetworkingTestCoverageSpec) error {
	specs = NormalizeNetworkingTestCoverageSpecs(specs)
	required := DefaultRequiredNetworkingTestCoverage()
	if len(specs) != len(required) {
		return fmt.Errorf("networking required test coverage must define %d areas", len(required))
	}
	seen := make(map[NetworkingRequiredTest]NetworkingTestCoverageSpec, len(specs))
	for _, spec := range specs {
		if !IsNetworkingRequiredTest(spec.Test) {
			return fmt.Errorf("unknown networking required test %q", spec.Test)
		}
		if !IsNetworkingTestCoverageCategory(spec.Category) {
			return fmt.Errorf("unknown networking test coverage category %q", spec.Category)
		}
		if spec.Title == "" {
			return errors.New("networking test coverage title is required")
		}
		if _, found := seen[spec.Test]; found {
			return errors.New("networking test coverage duplicate required test")
		}
		seen[spec.Test] = spec
	}
	for _, requiredSpec := range required {
		spec, found := seen[requiredSpec.Test]
		if !found {
			return fmt.Errorf("networking test coverage missing %s", requiredSpec.Test)
		}
		if spec.Category != requiredSpec.Category {
			return fmt.Errorf("networking test coverage %s must be %s", requiredSpec.Test, requiredSpec.Category)
		}
	}
	return nil
}

func EvaluateNetworkingTestCoverage(evidence []NetworkingTestCoverageEvidence) (NetworkingTestCoverageReport, error) {
	required := DefaultRequiredNetworkingTestCoverage()
	if err := ValidateRequiredNetworkingTestCoverage(required); err != nil {
		return NetworkingTestCoverageReport{}, err
	}
	normalized := NormalizeNetworkingTestCoverageEvidence(evidence)
	report := NetworkingTestCoverageReport{
		Required:	required,
		Evidence:	normalized,
	}
	byTest := make(map[NetworkingRequiredTest]NetworkingTestCoverageEvidence, len(normalized))
	for _, item := range normalized {
		if !IsNetworkingRequiredTest(item.Test) {
			return NetworkingTestCoverageReport{}, fmt.Errorf("unknown networking required test evidence %q", item.Test)
		}
		if !IsNetworkingTestCoverageCategory(item.Category) {
			return NetworkingTestCoverageReport{}, fmt.Errorf("unknown networking test coverage evidence category %q", item.Category)
		}
		if _, found := byTest[item.Test]; found {
			return NetworkingTestCoverageReport{}, errors.New("networking test coverage evidence duplicate required test")
		}
		byTest[item.Test] = item
	}
	for _, spec := range required {
		item, found := byTest[spec.Test]
		if !found {
			report.Missing = append(report.Missing, spec.Test)
			continue
		}
		if item.Category != spec.Category {
			return NetworkingTestCoverageReport{}, fmt.Errorf("networking test coverage evidence %s must be %s", spec.Test, spec.Category)
		}
		if !item.Passed || len(item.TestNames) == 0 {
			report.Failed = append(report.Failed, spec.Test)
		}
	}
	sortRequiredTests(report.Missing)
	sortRequiredTests(report.Failed)
	report.Ready = len(report.Missing) == 0 && len(report.Failed) == 0
	report.ReportHash = ComputeNetworkingTestCoverageReportHash(report)
	return report, nil
}

func ComputeNetworkingTestCoverageRoot(specs []NetworkingTestCoverageSpec) string {
	parts := []string{"networking-required-test-coverage"}
	for _, spec := range NormalizeNetworkingTestCoverageSpecs(specs) {
		parts = append(parts, string(spec.Category), string(spec.Test), spec.Title)
	}
	return HashParts(parts...)
}

func ComputeNetworkingTestCoverageReportHash(report NetworkingTestCoverageReport) string {
	parts := []string{"networking-test-coverage-report", fmt.Sprintf("%t", report.Ready)}
	for _, item := range NormalizeNetworkingTestCoverageEvidence(report.Evidence) {
		parts = append(parts, string(item.Category), string(item.Test), fmt.Sprintf("%t", item.Passed))
		parts = append(parts, item.TestNames...)
	}
	for _, missing := range report.Missing {
		parts = append(parts, "missing", string(missing))
	}
	for _, failed := range report.Failed {
		parts = append(parts, "failed", string(failed))
	}
	return HashParts(parts...)
}

func IsNetworkingRequiredTest(test NetworkingRequiredTest) bool {
	switch test {
	case RequiredTestNodeIDDerivation,
		RequiredTestNodeRecordSignature,
		RequiredTestSessionHandshake,
		RequiredTestStreamPriority,
		RequiredTestOverlayMembership,
		RequiredTestRouteCost,
		RequiredTestNetworkMessageID,
		RequiredTestDiscoveryRecordExpiry,
		RequiredTestChunkHashVerification,
		RequiredTestBroadcastDeduplication,
		RequiredTestCometBFTANAConsensus,
		RequiredTestMultiplexedSessionStreams,
		RequiredTestZoneOverlayFormation,
		RequiredTestServiceOverlayFormation,
		RequiredTestCrossZoneDelivery,
		RequiredTestRL2BlockChunkTransfer,
		RequiredTestResumableStateSnapshot,
		RequiredTestDiscoveryProofLookup,
		RequiredTestHeaderFirstPropagation,
		RequiredTestReplayedHandshake,
		RequiredTestForgedNodeAdvertisement,
		RequiredTestExpiredDiscoverySecurity,
		RequiredTestConflictingBroadcast,
		RequiredTestInvalidChunkSecurity,
		RequiredTestEclipsePeerSetSimulation,
		RequiredTestServiceSpamFlood,
		RequiredTestConsensusUnderBulkLoad,
		RequiredTestCrossZoneReplaySecurity,
		RequiredTestBlockHeaderLatency,
		RequiredTestBlockReconstructionTime,
		RequiredTestChunkStreamingThroughput,
		RequiredTestDiscoveryQueryLatency,
		RequiredTestOverlayJoinLatency,
		RequiredTestCrossZonePropagation,
		RequiredTestServiceTrafficThroughput,
		RequiredTestConsensusMixedLoadLatency,
		RequiredTestPeerRotationStability:
		return true
	default:
		return false
	}
}

func IsNetworkingTestCoverageCategory(category NetworkingTestCoverageCategory) bool {
	switch category {
	case NetworkingTestCoverageUnit, NetworkingTestCoverageIntegration, NetworkingTestCoverageSecurity, NetworkingTestCoveragePerformance:
		return true
	default:
		return false
	}
}

func NormalizeNetworkingTestCoverageSpecs(specs []NetworkingTestCoverageSpec) []NetworkingTestCoverageSpec {
	out := make([]NetworkingTestCoverageSpec, 0, len(specs))
	for _, spec := range specs {
		spec.Test = NetworkingRequiredTest(strings.ToLower(strings.TrimSpace(string(spec.Test))))
		spec.Category = NetworkingTestCoverageCategory(strings.ToLower(strings.TrimSpace(string(spec.Category))))
		spec.Title = strings.TrimSpace(spec.Title)
		if spec.Test == "" {
			continue
		}
		out = append(out, spec)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Test < out[j].Test
	})
	return out
}

func NormalizeNetworkingTestCoverageEvidence(evidence []NetworkingTestCoverageEvidence) []NetworkingTestCoverageEvidence {
	out := make([]NetworkingTestCoverageEvidence, 0, len(evidence))
	for _, item := range evidence {
		item.Test = NetworkingRequiredTest(strings.ToLower(strings.TrimSpace(string(item.Test))))
		item.Category = NetworkingTestCoverageCategory(strings.ToLower(strings.TrimSpace(string(item.Category))))
		names := make([]string, 0, len(item.TestNames))
		seen := make(map[string]struct{}, len(item.TestNames))
		for _, name := range item.TestNames {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, found := seen[name]; found {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
		sort.Strings(names)
		item.TestNames = names
		if item.Test == "" {
			continue
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Test < out[j].Test
	})
	return out
}

func sortRequiredTests(values []NetworkingRequiredTest) {
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
}
