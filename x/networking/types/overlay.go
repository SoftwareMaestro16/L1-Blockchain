package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const MaxOverlayDescriptors = 64

type OverlayType string

const (
	OverlayTypeValidator	OverlayType	= "VALIDATOR_OVERLAY"
	OverlayTypeZone		OverlayType	= "ZONE_OVERLAY"
	OverlayTypeExecution	OverlayType	= "EXECUTION_OVERLAY"
	OverlayTypeData		OverlayType	= "DATA_OVERLAY"
	OverlayTypeService	OverlayType	= "SERVICE_OVERLAY"
	OverlayTypeDiscovery	OverlayType	= "DISCOVERY_OVERLAY"
	OverlayTypeStorage	OverlayType	= "STORAGE_OVERLAY"
	OverlayTypeRouting	OverlayType	= "ROUTING_OVERLAY"
)

type OverlayMembershipRule string

const (
	OverlayMembershipValidatorSet		OverlayMembershipRule	= "VALIDATOR_SET"
	OverlayMembershipZoneSupported		OverlayMembershipRule	= "ZONE_SUPPORTED"
	OverlayMembershipExecutionRole		OverlayMembershipRule	= "EXECUTION_ROLE"
	OverlayMembershipDataProvider		OverlayMembershipRule	= "DATA_PROVIDER"
	OverlayMembershipServiceAdvertisement	OverlayMembershipRule	= "SERVICE_ADVERTISEMENT"
	OverlayMembershipSignedDiscovery	OverlayMembershipRule	= "SIGNED_DISCOVERY"
	OverlayMembershipStorageProvider	OverlayMembershipRule	= "STORAGE_PROVIDER"
	OverlayMembershipRoutingRole		OverlayMembershipRule	= "ROUTING_ROLE"
)

type RoutingStrategy string

const (
	RoutingStrategyDeterministicRoundRobin	RoutingStrategy	= "DETERMINISTIC_ROUND_ROBIN"
	RoutingStrategyKBucket			RoutingStrategy	= "K_BUCKET"
	RoutingStrategyFanoutGossip		RoutingStrategy	= "FANOUT_GOSSIP"
	RoutingStrategyBroadcast		RoutingStrategy	= "BROADCAST"
	RoutingStrategyLowLatencyAdvisory	RoutingStrategy	= "LOW_LATENCY_ADVISORY"
	RoutingStrategyRandomWalkAdvisory	RoutingStrategy	= "RANDOM_WALK_ADVISORY"
	RoutingStrategyShortestLatencyPath	RoutingStrategy	= "SHORTEST_LATENCY_PATH"
	RoutingStrategyZoneLocal		RoutingStrategy	= "ZONE_LOCAL"
	RoutingStrategyProbabilisticGossip	RoutingStrategy	= "PROBABILISTIC_GOSSIP_FALLBACK"
	RoutingStrategyDeterministicShard	RoutingStrategy	= "DETERMINISTIC_SHARD"
	RoutingStrategyPriorityBroadcastTree	RoutingStrategy	= "PRIORITY_BROADCAST_TREE"
	RoutingStrategyServiceProvider		RoutingStrategy	= "SERVICE_PROVIDER"
	RoutingStrategyStorageProvider		RoutingStrategy	= "STORAGE_PROVIDER"
)

type OverlayDescriptor struct {
	OverlayID	string
	OverlayType	OverlayType
	PolicyHash	string
	Membership	OverlayMembershipRule
	Routing		RoutingStrategy
	MinPeers	uint32
	MaxPeers	uint32
	Fanout		uint32
	QoSClass	QoSClass
	ExpiresHeight	uint64
	Version		uint64
}

func NewOverlayDescriptor(desc OverlayDescriptor) (OverlayDescriptor, error) {
	desc = NormalizeOverlayDescriptor(desc)
	if desc.OverlayID == "" {
		desc.OverlayID = ComputeOverlayID(desc)
	}
	if err := desc.ValidateBasic(); err != nil {
		return OverlayDescriptor{}, err
	}
	return desc, nil
}

func NormalizeOverlayDescriptor(desc OverlayDescriptor) OverlayDescriptor {
	desc.OverlayID = normalizeHashText(desc.OverlayID)
	desc.OverlayType = OverlayType(strings.ToUpper(strings.TrimSpace(string(desc.OverlayType))))
	desc.PolicyHash = normalizeHashText(desc.PolicyHash)
	desc.Membership = OverlayMembershipRule(strings.ToUpper(strings.TrimSpace(string(desc.Membership))))
	desc.Routing = RoutingStrategy(strings.ToUpper(strings.TrimSpace(string(desc.Routing))))
	desc.QoSClass = QoSClass(strings.ToLower(strings.TrimSpace(string(desc.QoSClass))))
	return desc
}

func ComputeOverlayID(desc OverlayDescriptor) string {
	desc = NormalizeOverlayDescriptor(desc)
	return HashParts(
		"overlay-descriptor",
		string(desc.OverlayType),
		desc.PolicyHash,
		string(desc.Membership),
		string(desc.Routing),
		fmt.Sprintf("%d", desc.MinPeers),
		fmt.Sprintf("%d", desc.MaxPeers),
		fmt.Sprintf("%d", desc.Fanout),
		string(desc.QoSClass),
		fmt.Sprintf("%d", desc.ExpiresHeight),
		fmt.Sprintf("%d", desc.Version),
	)
}

func (d OverlayDescriptor) ValidateBasic() error {
	desc := NormalizeOverlayDescriptor(d)
	if err := ValidateHash("networking overlay id", desc.OverlayID); err != nil {
		return err
	}
	if desc.OverlayID != ComputeOverlayID(desc) {
		return errors.New("networking overlay id does not match descriptor")
	}
	if !IsOverlayType(desc.OverlayType) {
		return fmt.Errorf("unknown networking overlay type %q", desc.OverlayType)
	}
	if err := ValidateHash("networking overlay policy hash", desc.PolicyHash); err != nil {
		return err
	}
	if !IsOverlayMembershipRule(desc.Membership) {
		return fmt.Errorf("unknown networking overlay membership rule %q", desc.Membership)
	}
	if !IsRoutingStrategy(desc.Routing) {
		return fmt.Errorf("unknown networking overlay routing strategy %q", desc.Routing)
	}
	if desc.MinPeers == 0 {
		return errors.New("networking overlay min peers must be positive")
	}
	if desc.MaxPeers < desc.MinPeers {
		return errors.New("networking overlay max peers must be >= min peers")
	}
	if desc.Fanout == 0 || desc.Fanout > desc.MaxPeers {
		return errors.New("networking overlay fanout must be positive and <= max peers")
	}
	if !IsQoSClass(desc.QoSClass) {
		return fmt.Errorf("unknown networking overlay qos class %q", desc.QoSClass)
	}
	if desc.Version == 0 {
		return errors.New("networking overlay version must be positive")
	}
	return validateOverlayCompatibility(desc)
}

func DefaultOverlayDescriptors() []OverlayDescriptor {
	descriptors := []OverlayDescriptor{
		mustOverlayDescriptor(OverlayTypeValidator, OverlayMembershipValidatorSet, RoutingStrategyDeterministicRoundRobin, QoSClassCriticalConsensus, 4, 128, 8),
		mustOverlayDescriptor(OverlayTypeZone, OverlayMembershipZoneSupported, RoutingStrategyKBucket, QoSClassExecutionMessage, 2, 96, 6),
		mustOverlayDescriptor(OverlayTypeExecution, OverlayMembershipExecutionRole, RoutingStrategyFanoutGossip, QoSClassExecutionMessage, 2, 96, 8),
		mustOverlayDescriptor(OverlayTypeData, OverlayMembershipDataProvider, RoutingStrategyKBucket, QoSClassBulkData, 2, 128, 8),
		mustOverlayDescriptor(OverlayTypeService, OverlayMembershipServiceAdvertisement, RoutingStrategyLowLatencyAdvisory, QoSClassServiceCall, 2, 64, 6),
		mustOverlayDescriptor(OverlayTypeDiscovery, OverlayMembershipSignedDiscovery, RoutingStrategyRandomWalkAdvisory, QoSClassDiscovery, 2, 128, 8),
		mustOverlayDescriptor(OverlayTypeStorage, OverlayMembershipStorageProvider, RoutingStrategyKBucket, QoSClassBulkData, 2, 128, 8),
		mustOverlayDescriptor(OverlayTypeRouting, OverlayMembershipRoutingRole, RoutingStrategyFanoutGossip, QoSClassServiceCall, 2, 96, 8),
	}
	sortOverlayDescriptors(descriptors)
	return descriptors
}

func RegisterOverlayDescriptor(state NetworkingState, desc OverlayDescriptor, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	if currentHeight == 0 {
		return NetworkingState{}, errors.New("networking current height must be positive")
	}
	desc, err := NewOverlayDescriptor(desc)
	if err != nil {
		return NetworkingState{}, err
	}
	if desc.ExpiresHeight > 0 && currentHeight > desc.ExpiresHeight {
		return NetworkingState{}, errors.New("networking overlay descriptor is expired")
	}
	next := state.Clone()
	replaced := false
	for i, existing := range next.OverlayDescriptors {
		if existing.OverlayID == desc.OverlayID {
			next.OverlayDescriptors[i] = desc
			replaced = true
			break
		}
	}
	if !replaced {
		next.OverlayDescriptors = append(next.OverlayDescriptors, desc)
	}
	sortOverlayDescriptors(next.OverlayDescriptors)
	return next, next.Validate()
}

func NodeSatisfiesOverlayMembership(record NodeRecord, desc OverlayDescriptor) (bool, error) {
	record = NormalizeNodeRecord(record)
	if err := record.ValidateBasic(); err != nil {
		return false, err
	}
	desc = NormalizeOverlayDescriptor(desc)
	if err := desc.ValidateBasic(); err != nil {
		return false, err
	}
	switch desc.Membership {
	case OverlayMembershipValidatorSet:
		return hasRole(record.Roles, NodeRoleValidator), nil
	case OverlayMembershipZoneSupported:
		return len(record.ZonesSupported) > 0, nil
	case OverlayMembershipExecutionRole:
		return hasRole(record.Roles, NodeRoleZoneExecution), nil
	case OverlayMembershipDataProvider:
		return hasRole(record.Roles, NodeRoleFull) || hasRole(record.Roles, NodeRoleArchive) || hasRole(record.Roles, NodeRoleStateSync) || hasRole(record.Roles, NodeRoleStorageProvider), nil
	case OverlayMembershipServiceAdvertisement:
		return hasRole(record.Roles, NodeRoleService) && len(record.ServicesSupported) > 0, nil
	case OverlayMembershipSignedDiscovery:
		return true, nil
	case OverlayMembershipStorageProvider:
		return hasRole(record.Roles, NodeRoleStorageProvider), nil
	case OverlayMembershipRoutingRole:
		return hasRole(record.Roles, NodeRoleRouting), nil
	default:
		return false, fmt.Errorf("unknown networking overlay membership rule %q", desc.Membership)
	}
}

func PlanOverlayFanout(desc OverlayDescriptor, eligiblePeers uint32) (uint32, error) {
	desc = NormalizeOverlayDescriptor(desc)
	if err := desc.ValidateBasic(); err != nil {
		return 0, err
	}
	if eligiblePeers < desc.MinPeers {
		return 0, errors.New("networking overlay has insufficient eligible peers")
	}
	limit := eligiblePeers
	if limit > desc.MaxPeers {
		limit = desc.MaxPeers
	}
	if desc.Fanout < limit {
		return desc.Fanout, nil
	}
	return limit, nil
}

func ValidateOverlayDescriptors(descriptors []OverlayDescriptor, currentHeight uint64) error {
	return validateOverlayDescriptors(descriptors, currentHeight)
}

func IsOverlayType(overlayType OverlayType) bool {
	switch overlayType {
	case OverlayTypeValidator, OverlayTypeZone, OverlayTypeExecution, OverlayTypeData, OverlayTypeService, OverlayTypeDiscovery, OverlayTypeStorage, OverlayTypeRouting:
		return true
	default:
		return false
	}
}

func IsOverlayMembershipRule(rule OverlayMembershipRule) bool {
	switch rule {
	case OverlayMembershipValidatorSet, OverlayMembershipZoneSupported, OverlayMembershipExecutionRole, OverlayMembershipDataProvider, OverlayMembershipServiceAdvertisement, OverlayMembershipSignedDiscovery, OverlayMembershipStorageProvider, OverlayMembershipRoutingRole:
		return true
	default:
		return false
	}
}

func IsRoutingStrategy(strategy RoutingStrategy) bool {
	switch strategy {
	case RoutingStrategyDeterministicRoundRobin,
		RoutingStrategyKBucket,
		RoutingStrategyFanoutGossip,
		RoutingStrategyBroadcast,
		RoutingStrategyLowLatencyAdvisory,
		RoutingStrategyRandomWalkAdvisory,
		RoutingStrategyShortestLatencyPath,
		RoutingStrategyZoneLocal,
		RoutingStrategyProbabilisticGossip,
		RoutingStrategyDeterministicShard,
		RoutingStrategyPriorityBroadcastTree,
		RoutingStrategyServiceProvider,
		RoutingStrategyStorageProvider:
		return true
	default:
		return false
	}
}

func mustOverlayDescriptor(overlayType OverlayType, membership OverlayMembershipRule, routing RoutingStrategy, qos QoSClass, minPeers, maxPeers, fanout uint32) OverlayDescriptor {
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	overlayType,
		PolicyHash:	HashParts("default-overlay-policy", string(overlayType)),
		Membership:	membership,
		Routing:	routing,
		MinPeers:	minPeers,
		MaxPeers:	maxPeers,
		Fanout:		fanout,
		QoSClass:	qos,
		Version:	1,
	})
	if err != nil {
		panic(err)
	}
	return desc
}

func validateOverlayDescriptors(descriptors []OverlayDescriptor, currentHeight uint64) error {
	if len(descriptors) == 0 {
		return errors.New("networking overlay descriptors are required")
	}
	if len(descriptors) > MaxOverlayDescriptors {
		return fmt.Errorf("networking overlay descriptors must be <= %d", MaxOverlayDescriptors)
	}
	seen := make(map[string]struct{}, len(descriptors))
	var previous string
	for i, desc := range descriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if err := desc.ValidateBasic(); err != nil {
			return err
		}
		if desc.ExpiresHeight > 0 && currentHeight > 0 && currentHeight > desc.ExpiresHeight {
			return errors.New("networking overlay descriptor is expired")
		}
		if _, found := seen[desc.OverlayID]; found {
			return errors.New("networking duplicate overlay descriptor")
		}
		seen[desc.OverlayID] = struct{}{}
		if i > 0 && previous >= desc.OverlayID {
			return errors.New("networking overlay descriptors must be sorted canonically")
		}
		previous = desc.OverlayID
	}
	return nil
}

func validateOverlayCompatibility(desc OverlayDescriptor) error {
	expectedMembership, expectedQoS, found := overlayCompatibility(desc.OverlayType)
	if !found {
		return fmt.Errorf("unknown networking overlay type %q", desc.OverlayType)
	}
	if desc.Membership != expectedMembership {
		return errors.New("networking overlay membership rule does not match overlay type")
	}
	if desc.QoSClass != expectedQoS {
		return errors.New("networking overlay qos class does not match overlay type")
	}
	if desc.OverlayType == OverlayTypeValidator && isAdvisoryRoutingStrategy(desc.Routing) {
		return errors.New("networking validator overlay cannot use advisory routing")
	}
	return nil
}

func overlayCompatibility(overlayType OverlayType) (OverlayMembershipRule, QoSClass, bool) {
	switch overlayType {
	case OverlayTypeValidator:
		return OverlayMembershipValidatorSet, QoSClassCriticalConsensus, true
	case OverlayTypeZone:
		return OverlayMembershipZoneSupported, QoSClassExecutionMessage, true
	case OverlayTypeExecution:
		return OverlayMembershipExecutionRole, QoSClassExecutionMessage, true
	case OverlayTypeData:
		return OverlayMembershipDataProvider, QoSClassBulkData, true
	case OverlayTypeService:
		return OverlayMembershipServiceAdvertisement, QoSClassServiceCall, true
	case OverlayTypeDiscovery:
		return OverlayMembershipSignedDiscovery, QoSClassDiscovery, true
	case OverlayTypeStorage:
		return OverlayMembershipStorageProvider, QoSClassBulkData, true
	case OverlayTypeRouting:
		return OverlayMembershipRoutingRole, QoSClassServiceCall, true
	default:
		return "", "", false
	}
}

func isAdvisoryRoutingStrategy(strategy RoutingStrategy) bool {
	return strategy == RoutingStrategyLowLatencyAdvisory ||
		strategy == RoutingStrategyRandomWalkAdvisory ||
		strategy == RoutingStrategyShortestLatencyPath
}

func cloneOverlayDescriptors(descriptors []OverlayDescriptor) []OverlayDescriptor {
	out := make([]OverlayDescriptor, len(descriptors))
	for i, desc := range descriptors {
		out[i] = NormalizeOverlayDescriptor(desc)
	}
	return out
}

func sortOverlayDescriptors(descriptors []OverlayDescriptor) {
	sort.SliceStable(descriptors, func(i, j int) bool {
		return NormalizeOverlayDescriptor(descriptors[i]).OverlayID < NormalizeOverlayDescriptor(descriptors[j]).OverlayID
	})
}
