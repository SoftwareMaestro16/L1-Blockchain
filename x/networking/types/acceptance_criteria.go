package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingAcceptanceCriterion string

const (
	AcceptanceCriterionL0CometBFTProtected		NetworkingAcceptanceCriterion	= "l0_cometbft_transport_consensus_critical_protected"
	AcceptanceCriterionANAChannelQoS		NetworkingAcceptanceCriterion	= "ana_channel_classification_peer_scoring_adaptive_fanout_qos"
	AcceptanceCriterionL1IdentitySessions		NetworkingAcceptanceCriterion	= "l1_node_identity_session_channels_defined_testable"
	AcceptanceCriterionL2Overlays			NetworkingAcceptanceCriterion	= "l2_overlays_validator_zone_execution_data_service_discovery"
	AcceptanceCriterionL3AetherMesh			NetworkingAcceptanceCriterion	= "l3_aether_mesh_execution_query_service_storage_cross_zone"
	AcceptanceCriterionRL2Streaming			NetworkingAcceptanceCriterion	= "rl2_chunked_verified_resumable_bandwidth_adaptive_streaming"
	AcceptanceCriterionDRTDiscovery			NetworkingAcceptanceCriterion	= "drt_signed_expiring_proof_attached_discovery_records"
	AcceptanceCriterionHybridBroadcast		NetworkingAcceptanceCriterion	= "hybrid_broadcast_tree_gossip_fallback_hash_dedup"
	AcceptanceCriterionSecurityControls		NetworkingAcceptanceCriterion	= "security_peer_reputation_eclipse_spam_replay_chunk_verification"
	AcceptanceCriterionCosmosABCIIntegration	NetworkingAcceptanceCriterion	= "cosmos_sdk_abci_integration_rules_explicit"
	AcceptanceCriterionRequiredTestCoverage		NetworkingAcceptanceCriterion	= "tests_unit_integration_security_performance"
)

type NetworkingAcceptanceCriteriaSpec struct {
	Criteria	[]NetworkingAcceptanceCriterion
	SpecRoot	string
}

type NetworkingAcceptanceEvidence struct {
	Criterion	NetworkingAcceptanceCriterion
	Evidence	[]string
	Accepted	bool
}

type NetworkingImplementationPlanningReport struct {
	Criteria	NetworkingAcceptanceCriteriaSpec
	Evidence	[]NetworkingAcceptanceEvidence
	Missing		[]NetworkingAcceptanceCriterion
	Rejected	[]NetworkingAcceptanceCriterion
	Ready		bool
	ReadinessHash	string
}

func DefaultNetworkingAcceptanceCriteriaSpec() NetworkingAcceptanceCriteriaSpec {
	spec := NetworkingAcceptanceCriteriaSpec{
		Criteria: []NetworkingAcceptanceCriterion{
			AcceptanceCriterionL0CometBFTProtected,
			AcceptanceCriterionANAChannelQoS,
			AcceptanceCriterionL1IdentitySessions,
			AcceptanceCriterionL2Overlays,
			AcceptanceCriterionL3AetherMesh,
			AcceptanceCriterionRL2Streaming,
			AcceptanceCriterionDRTDiscovery,
			AcceptanceCriterionHybridBroadcast,
			AcceptanceCriterionSecurityControls,
			AcceptanceCriterionCosmosABCIIntegration,
			AcceptanceCriterionRequiredTestCoverage,
		},
	}
	spec = NormalizeNetworkingAcceptanceCriteriaSpec(spec)
	spec.SpecRoot = ComputeNetworkingAcceptanceCriteriaRoot(spec)
	return spec
}

func ValidateNetworkingAcceptanceCriteriaSpec(spec NetworkingAcceptanceCriteriaSpec) error {
	normalized := NormalizeNetworkingAcceptanceCriteriaSpec(spec)
	required := DefaultNetworkingAcceptanceCriteriaSpec()
	if len(normalized.Criteria) != len(required.Criteria) {
		return fmt.Errorf("networking acceptance criteria must define %d criteria", len(required.Criteria))
	}
	seen := make(map[NetworkingAcceptanceCriterion]struct{}, len(normalized.Criteria))
	for _, criterion := range normalized.Criteria {
		if !IsNetworkingAcceptanceCriterion(criterion) {
			return fmt.Errorf("unknown networking acceptance criterion %q", criterion)
		}
		if _, found := seen[criterion]; found {
			return errors.New("networking acceptance criteria duplicate criterion")
		}
		seen[criterion] = struct{}{}
	}
	for _, criterion := range required.Criteria {
		if _, found := seen[criterion]; !found {
			return fmt.Errorf("networking acceptance criteria missing %s", criterion)
		}
	}
	if normalized.SpecRoot == "" {
		return errors.New("networking acceptance criteria root is required")
	}
	if normalized.SpecRoot != ComputeNetworkingAcceptanceCriteriaRoot(normalized) {
		return errors.New("networking acceptance criteria root mismatch")
	}
	return nil
}

func EvaluateNetworkingImplementationPlanningReadiness(spec NetworkingAcceptanceCriteriaSpec, evidence []NetworkingAcceptanceEvidence) (NetworkingImplementationPlanningReport, error) {
	spec = NormalizeNetworkingAcceptanceCriteriaSpec(spec)
	if err := ValidateNetworkingAcceptanceCriteriaSpec(spec); err != nil {
		return NetworkingImplementationPlanningReport{}, err
	}
	normalizedEvidence := NormalizeNetworkingAcceptanceEvidence(evidence)
	report := NetworkingImplementationPlanningReport{
		Criteria:	spec,
		Evidence:	normalizedEvidence,
	}
	byCriterion := make(map[NetworkingAcceptanceCriterion]NetworkingAcceptanceEvidence, len(normalizedEvidence))
	for _, item := range normalizedEvidence {
		if !IsNetworkingAcceptanceCriterion(item.Criterion) {
			return NetworkingImplementationPlanningReport{}, fmt.Errorf("unknown networking acceptance evidence %q", item.Criterion)
		}
		if _, found := byCriterion[item.Criterion]; found {
			return NetworkingImplementationPlanningReport{}, errors.New("networking acceptance evidence duplicate criterion")
		}
		byCriterion[item.Criterion] = item
	}
	for _, criterion := range spec.Criteria {
		item, found := byCriterion[criterion]
		if !found {
			report.Missing = append(report.Missing, criterion)
			continue
		}
		if !item.Accepted || len(item.Evidence) == 0 {
			report.Rejected = append(report.Rejected, criterion)
		}
	}
	sortNetworkingAcceptanceCriteria(report.Missing)
	sortNetworkingAcceptanceCriteria(report.Rejected)
	report.Ready = len(report.Missing) == 0 && len(report.Rejected) == 0
	report.ReadinessHash = ComputeNetworkingImplementationPlanningReadinessHash(report)
	return report, nil
}

func ComputeNetworkingAcceptanceCriteriaRoot(spec NetworkingAcceptanceCriteriaSpec) string {
	spec = NormalizeNetworkingAcceptanceCriteriaSpec(spec)
	parts := []string{"networking-acceptance-criteria"}
	for _, criterion := range spec.Criteria {
		parts = append(parts, string(criterion))
	}
	return HashParts(parts...)
}

func ComputeNetworkingImplementationPlanningReadinessHash(report NetworkingImplementationPlanningReport) string {
	parts := []string{"networking-implementation-planning-readiness", report.Criteria.SpecRoot, fmt.Sprintf("%t", report.Ready)}
	for _, item := range NormalizeNetworkingAcceptanceEvidence(report.Evidence) {
		parts = append(parts, string(item.Criterion), fmt.Sprintf("%t", item.Accepted))
		parts = append(parts, item.Evidence...)
	}
	for _, criterion := range report.Missing {
		parts = append(parts, "missing", string(criterion))
	}
	for _, criterion := range report.Rejected {
		parts = append(parts, "rejected", string(criterion))
	}
	return HashParts(parts...)
}

func IsNetworkingAcceptanceCriterion(criterion NetworkingAcceptanceCriterion) bool {
	switch criterion {
	case AcceptanceCriterionL0CometBFTProtected,
		AcceptanceCriterionANAChannelQoS,
		AcceptanceCriterionL1IdentitySessions,
		AcceptanceCriterionL2Overlays,
		AcceptanceCriterionL3AetherMesh,
		AcceptanceCriterionRL2Streaming,
		AcceptanceCriterionDRTDiscovery,
		AcceptanceCriterionHybridBroadcast,
		AcceptanceCriterionSecurityControls,
		AcceptanceCriterionCosmosABCIIntegration,
		AcceptanceCriterionRequiredTestCoverage:
		return true
	default:
		return false
	}
}

func NormalizeNetworkingAcceptanceCriteriaSpec(spec NetworkingAcceptanceCriteriaSpec) NetworkingAcceptanceCriteriaSpec {
	spec.Criteria = normalizeNetworkingAcceptanceCriteria(spec.Criteria)
	spec.SpecRoot = normalizeHashText(spec.SpecRoot)
	return spec
}

func NormalizeNetworkingAcceptanceEvidence(evidence []NetworkingAcceptanceEvidence) []NetworkingAcceptanceEvidence {
	out := make([]NetworkingAcceptanceEvidence, 0, len(evidence))
	for _, item := range evidence {
		item.Criterion = NetworkingAcceptanceCriterion(strings.ToLower(strings.TrimSpace(string(item.Criterion))))
		seen := make(map[string]struct{}, len(item.Evidence))
		normalized := make([]string, 0, len(item.Evidence))
		for _, value := range item.Evidence {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if _, found := seen[value]; found {
				continue
			}
			seen[value] = struct{}{}
			normalized = append(normalized, value)
		}
		sort.Strings(normalized)
		item.Evidence = normalized
		if item.Criterion == "" {
			continue
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Criterion < out[j].Criterion
	})
	return out
}

func normalizeNetworkingAcceptanceCriteria(criteria []NetworkingAcceptanceCriterion) []NetworkingAcceptanceCriterion {
	out := make([]NetworkingAcceptanceCriterion, 0, len(criteria))
	seen := make(map[NetworkingAcceptanceCriterion]struct{}, len(criteria))
	for _, criterion := range criteria {
		criterion = NetworkingAcceptanceCriterion(strings.ToLower(strings.TrimSpace(string(criterion))))
		if criterion == "" {
			continue
		}
		if _, found := seen[criterion]; found {
			continue
		}
		seen[criterion] = struct{}{}
		out = append(out, criterion)
	}
	sortNetworkingAcceptanceCriteria(out)
	return out
}

func sortNetworkingAcceptanceCriteria(criteria []NetworkingAcceptanceCriterion) {
	sort.SliceStable(criteria, func(i, j int) bool { return criteria[i] < criteria[j] })
}
