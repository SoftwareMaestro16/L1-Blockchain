package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingNonGoal string

const (
	NetworkingNonGoalApplicationLogic		NetworkingNonGoal	= "implement_application_logic"
	NetworkingNonGoalReplaceCometBFTConsensus	NetworkingNonGoal	= "replace_cometbft_consensus"
	NetworkingNonGoalCentralizedRouting		NetworkingNonGoal	= "assume_centralized_routing"
	NetworkingNonGoalExternalDiscoveryServices	NetworkingNonGoal	= "rely_on_external_discovery_services"
	NetworkingNonGoalMessagingOrSocialNetworkLayer	NetworkingNonGoal	= "introduce_messaging_or_social_network_layer"
	NetworkingNonGoalLiveMetricsConsensusAuthority	NetworkingNonGoal	= "make_live_network_metrics_consensus_authoritative"
	NetworkingNonGoalOffChainServiceConsensusLogic	NetworkingNonGoal	= "execute_offchain_service_logic_inside_consensus"
)

type NetworkingNonGoalSpec struct {
	NonGoals	[]NetworkingNonGoal
	SpecRoot	string
}

type NetworkingScopeBoundary struct {
	NonGoals				NetworkingNonGoalSpec
	ImplementsApplicationLogic		bool
	ReplacesCometBFTConsensus		bool
	RequiresCentralizedRouting		bool
	RequiresExternalDiscoveryServices	bool
	IntroducesMessagingSocialLayer		bool
	LiveMetricsConsensusAuthoritative	bool
	OffChainServiceLogicInConsensus		bool
	BoundaryRoot				string
}

func DefaultNetworkingNonGoalSpec() NetworkingNonGoalSpec {
	spec := NetworkingNonGoalSpec{
		NonGoals: []NetworkingNonGoal{
			NetworkingNonGoalApplicationLogic,
			NetworkingNonGoalReplaceCometBFTConsensus,
			NetworkingNonGoalCentralizedRouting,
			NetworkingNonGoalExternalDiscoveryServices,
			NetworkingNonGoalMessagingOrSocialNetworkLayer,
			NetworkingNonGoalLiveMetricsConsensusAuthority,
			NetworkingNonGoalOffChainServiceConsensusLogic,
		},
	}
	spec = NormalizeNetworkingNonGoalSpec(spec)
	spec.SpecRoot = ComputeNetworkingNonGoalSpecRoot(spec)
	return spec
}

func DefaultNetworkingScopeBoundary() NetworkingScopeBoundary {
	boundary := NetworkingScopeBoundary{
		NonGoals: DefaultNetworkingNonGoalSpec(),
	}
	boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
	return boundary
}

func ValidateNetworkingNonGoalSpec(spec NetworkingNonGoalSpec) error {
	normalized := NormalizeNetworkingNonGoalSpec(spec)
	required := DefaultNetworkingNonGoalSpec()
	if len(normalized.NonGoals) != len(required.NonGoals) {
		return fmt.Errorf("networking non-goal spec must define %d non-goals", len(required.NonGoals))
	}
	seen := make(map[NetworkingNonGoal]struct{}, len(normalized.NonGoals))
	for _, nonGoal := range normalized.NonGoals {
		if !IsNetworkingNonGoal(nonGoal) {
			return fmt.Errorf("unknown networking non-goal %q", nonGoal)
		}
		if _, found := seen[nonGoal]; found {
			return errors.New("networking non-goal spec duplicate non-goal")
		}
		seen[nonGoal] = struct{}{}
	}
	for _, nonGoal := range required.NonGoals {
		if _, found := seen[nonGoal]; !found {
			return fmt.Errorf("networking non-goal spec missing %s", nonGoal)
		}
	}
	if normalized.SpecRoot == "" {
		return errors.New("networking non-goal spec root is required")
	}
	if normalized.SpecRoot != ComputeNetworkingNonGoalSpecRoot(normalized) {
		return errors.New("networking non-goal spec root mismatch")
	}
	return nil
}

func ValidateNetworkingScopeBoundary(boundary NetworkingScopeBoundary) error {
	boundary = NormalizeNetworkingScopeBoundary(boundary)
	if err := ValidateNetworkingNonGoalSpec(boundary.NonGoals); err != nil {
		return err
	}
	if boundary.ImplementsApplicationLogic {
		return errors.New("networking layer must not implement application logic")
	}
	if boundary.ReplacesCometBFTConsensus {
		return errors.New("networking layer must not replace CometBFT consensus")
	}
	if boundary.RequiresCentralizedRouting {
		return errors.New("networking layer must not assume centralized routing")
	}
	if boundary.RequiresExternalDiscoveryServices {
		return errors.New("networking layer must not rely on external discovery services")
	}
	if boundary.IntroducesMessagingSocialLayer {
		return errors.New("networking layer must not introduce a messaging or social network layer")
	}
	if boundary.LiveMetricsConsensusAuthoritative {
		return errors.New("networking layer must not make live network metrics consensus-authoritative")
	}
	if boundary.OffChainServiceLogicInConsensus {
		return errors.New("networking layer must not execute off-chain service logic inside consensus")
	}
	if boundary.BoundaryRoot == "" {
		return errors.New("networking scope boundary root is required")
	}
	if boundary.BoundaryRoot != ComputeNetworkingScopeBoundaryRoot(boundary) {
		return errors.New("networking scope boundary root mismatch")
	}
	return nil
}

func ComputeNetworkingNonGoalSpecRoot(spec NetworkingNonGoalSpec) string {
	spec = NormalizeNetworkingNonGoalSpec(spec)
	parts := []string{"networking-non-goals"}
	for _, nonGoal := range spec.NonGoals {
		parts = append(parts, string(nonGoal))
	}
	return HashParts(parts...)
}

func ComputeNetworkingScopeBoundaryRoot(boundary NetworkingScopeBoundary) string {
	boundary = NormalizeNetworkingScopeBoundary(boundary)
	return HashParts(
		"networking-scope-boundary",
		boundary.NonGoals.SpecRoot,
		fmt.Sprintf("%t", boundary.ImplementsApplicationLogic),
		fmt.Sprintf("%t", boundary.ReplacesCometBFTConsensus),
		fmt.Sprintf("%t", boundary.RequiresCentralizedRouting),
		fmt.Sprintf("%t", boundary.RequiresExternalDiscoveryServices),
		fmt.Sprintf("%t", boundary.IntroducesMessagingSocialLayer),
		fmt.Sprintf("%t", boundary.LiveMetricsConsensusAuthoritative),
		fmt.Sprintf("%t", boundary.OffChainServiceLogicInConsensus),
	)
}

func IsNetworkingNonGoal(nonGoal NetworkingNonGoal) bool {
	switch nonGoal {
	case NetworkingNonGoalApplicationLogic,
		NetworkingNonGoalReplaceCometBFTConsensus,
		NetworkingNonGoalCentralizedRouting,
		NetworkingNonGoalExternalDiscoveryServices,
		NetworkingNonGoalMessagingOrSocialNetworkLayer,
		NetworkingNonGoalLiveMetricsConsensusAuthority,
		NetworkingNonGoalOffChainServiceConsensusLogic:
		return true
	default:
		return false
	}
}

func NormalizeNetworkingScopeBoundary(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
	boundary.NonGoals = NormalizeNetworkingNonGoalSpec(boundary.NonGoals)
	boundary.BoundaryRoot = normalizeHashText(boundary.BoundaryRoot)
	return boundary
}

func NormalizeNetworkingNonGoalSpec(spec NetworkingNonGoalSpec) NetworkingNonGoalSpec {
	spec.NonGoals = normalizeNetworkingNonGoals(spec.NonGoals)
	spec.SpecRoot = normalizeHashText(spec.SpecRoot)
	return spec
}

func normalizeNetworkingNonGoals(nonGoals []NetworkingNonGoal) []NetworkingNonGoal {
	out := make([]NetworkingNonGoal, 0, len(nonGoals))
	seen := make(map[NetworkingNonGoal]struct{}, len(nonGoals))
	for _, nonGoal := range nonGoals {
		nonGoal = NetworkingNonGoal(strings.ToLower(strings.TrimSpace(string(nonGoal))))
		if nonGoal == "" {
			continue
		}
		if _, found := seen[nonGoal]; found {
			continue
		}
		seen[nonGoal] = struct{}{}
		out = append(out, nonGoal)
	}
	sortNetworkingNonGoals(out)
	return out
}

func sortNetworkingNonGoals(nonGoals []NetworkingNonGoal) {
	sort.SliceStable(nonGoals, func(i, j int) bool { return nonGoals[i] < nonGoals[j] })
}
