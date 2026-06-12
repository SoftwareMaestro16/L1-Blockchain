package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultMinGossipFanout		= uint32(2)
	DefaultMaxGossipFanout		= uint32(16)
	DefaultConsensusReserveBps	= uint32(6_000)
	DefaultPeerScoreFloorBps	= uint32(2_500)
	DefaultMaxOutboundBytesPerBlock	= uint64(128 << 20)
)

type BaseTransportCapability string

const (
	BaseCapabilityConsensusGossip		BaseTransportCapability	= "CONSENSUS_MESSAGE_GOSSIP"
	BaseCapabilityProposalVotes		BaseTransportCapability	= "PROPOSAL_AND_VOTE_PROPAGATION"
	BaseCapabilityBlockPropagation		BaseTransportCapability	= "BLOCK_PROPAGATION"
	BaseCapabilityMempoolPropagation	BaseTransportCapability	= "BASELINE_MEMPOOL_TX_PROPAGATION"
	BaseCapabilityValidatorCoordination	BaseTransportCapability	= "VALIDATOR_COORDINATION"
	BaseCapabilityPeerTransport		BaseTransportCapability	= "PEER_TRANSPORT_PRIMITIVES"
)

type ANAResponsibility string

const (
	ANAResponsibilityPeerScoring		ANAResponsibility	= "PEER_SCORING"
	ANAResponsibilityConnectionMultiplex	ANAResponsibility	= "CONNECTION_MULTIPLEXING"
	ANAResponsibilityAdaptiveFanout		ANAResponsibility	= "ADAPTIVE_GOSSIP_FANOUT"
	ANAResponsibilityBandwidthAware		ANAResponsibility	= "BANDWIDTH_AWARE_PROPAGATION"
	ANAResponsibilityZoneHints		ANAResponsibility	= "ZONE_AWARE_ROUTING_HINTS"
	ANAResponsibilityMessagePriority	ANAResponsibility	= "MESSAGE_CLASS_PRIORITY"
	ANAResponsibilityStreamingNegotiation	ANAResponsibility	= "STREAMING_PAYLOAD_NEGOTIATION"
	ANAResponsibilityRoleValidation		ANAResponsibility	= "PEER_ROLE_ADVERTISEMENT_VALIDATION"
)

type MultiplexBinding struct {
	Channel			ChannelClass
	PhysicalTransport	string
	CometBFTPassthrough	bool
	Priority		uint32
}

type AdaptiveFanoutPolicy struct {
	MinFanout		uint32
	MaxFanout		uint32
	PeerScoreFloorBps	uint32
}

type BandwidthPolicy struct {
	MaxOutboundBytesPerBlock	uint64
	ConsensusReserveBps		uint32
}

type ZoneRoutingHint struct {
	ZoneID		string
	Channel		ChannelClass
	AdvisoryOnly	bool
}

type AetherNetworkingAdapter struct {
	BaselineTransport		string
	BaseCapabilities		[]BaseTransportCapability
	Responsibilities		[]ANAResponsibility
	ChannelBindings			[]MultiplexBinding
	Fanout				AdaptiveFanoutPolicy
	Bandwidth			BandwidthPolicy
	ZoneHints			[]ZoneRoutingHint
	ValidateRoleAdvertisements	bool
	ChangesConsensusValidity	bool
	HidesConsensusMessages		bool
	ReplacesCometBFTConsensusGossip	bool
	PeerMetricsAffectCommittedState	bool
}

type PropagationPlan struct {
	Envelope		TransportEnvelope
	HandledByCometBFT	bool
	AdapterFanout		uint32
	Priority		uint32
	ConsensusReserveBps	uint32
	UsesAdvisoryPeerMetric	bool
}

func DefaultAetherNetworkingAdapter() AetherNetworkingAdapter {
	policies := DefaultChannelPolicies()
	return AetherNetworkingAdapter{
		BaselineTransport:	BaseTransportCometBFTP2P,
		BaseCapabilities: []BaseTransportCapability{
			BaseCapabilityConsensusGossip,
			BaseCapabilityProposalVotes,
			BaseCapabilityBlockPropagation,
			BaseCapabilityMempoolPropagation,
			BaseCapabilityValidatorCoordination,
			BaseCapabilityPeerTransport,
		},
		Responsibilities: []ANAResponsibility{
			ANAResponsibilityPeerScoring,
			ANAResponsibilityConnectionMultiplex,
			ANAResponsibilityAdaptiveFanout,
			ANAResponsibilityBandwidthAware,
			ANAResponsibilityZoneHints,
			ANAResponsibilityMessagePriority,
			ANAResponsibilityStreamingNegotiation,
			ANAResponsibilityRoleValidation,
		},
		ChannelBindings:		defaultMultiplexBindings(policies),
		Fanout:				AdaptiveFanoutPolicy{MinFanout: DefaultMinGossipFanout, MaxFanout: DefaultMaxGossipFanout, PeerScoreFloorBps: DefaultPeerScoreFloorBps},
		Bandwidth:			BandwidthPolicy{MaxOutboundBytesPerBlock: DefaultMaxOutboundBytesPerBlock, ConsensusReserveBps: DefaultConsensusReserveBps},
		ValidateRoleAdvertisements:	true,
	}
}

func ValidateAetherNetworkingAdapter(adapter AetherNetworkingAdapter) error {
	if strings.TrimSpace(adapter.BaselineTransport) != BaseTransportCometBFTP2P {
		return errors.New("networking ANA must run above CometBFT P2P baseline")
	}
	if err := validateBaseCapabilities(adapter.BaseCapabilities); err != nil {
		return err
	}
	if err := validateANAResponsibilities(adapter.Responsibilities); err != nil {
		return err
	}
	if adapter.ChangesConsensusValidity {
		return errors.New("networking ANA must not change consensus message validity")
	}
	if adapter.HidesConsensusMessages {
		return errors.New("networking ANA must not hide consensus-critical messages from CometBFT")
	}
	if adapter.ReplacesCometBFTConsensusGossip {
		return errors.New("networking ANA must not replace CometBFT consensus gossip")
	}
	if adapter.PeerMetricsAffectCommittedState {
		return errors.New("networking ANA peer metrics are advisory until committed")
	}
	if !adapter.ValidateRoleAdvertisements {
		return errors.New("networking ANA must validate peer role advertisements")
	}
	if err := adapter.Fanout.Validate(); err != nil {
		return err
	}
	if err := adapter.Bandwidth.Validate(); err != nil {
		return err
	}
	return validateMultiplexBindings(adapter.ChannelBindings)
}

func (p AdaptiveFanoutPolicy) Validate() error {
	if p.MinFanout == 0 {
		return errors.New("networking ANA min fanout must be positive")
	}
	if p.MaxFanout < p.MinFanout {
		return errors.New("networking ANA max fanout must be >= min fanout")
	}
	if p.PeerScoreFloorBps > BasisPoints {
		return fmt.Errorf("networking ANA peer score floor must be <= %d bps", BasisPoints)
	}
	return nil
}

func (p BandwidthPolicy) Validate() error {
	if p.MaxOutboundBytesPerBlock == 0 {
		return errors.New("networking ANA max outbound bytes per block must be positive")
	}
	if p.ConsensusReserveBps == 0 || p.ConsensusReserveBps > BasisPoints {
		return fmt.Errorf("networking ANA consensus reserve must be between 1 and %d bps", BasisPoints)
	}
	return nil
}

func PlanPropagation(adapter AetherNetworkingAdapter, envelope TransportEnvelope, peerCount uint32, score PeerScore) (PropagationPlan, error) {
	if err := ValidateAetherNetworkingAdapter(adapter); err != nil {
		return PropagationPlan{}, err
	}
	if peerCount == 0 {
		return PropagationPlan{}, errors.New("networking ANA peer count must be positive")
	}
	if err := envelope.Validate(DefaultChannelPolicies()); err != nil {
		return PropagationPlan{}, err
	}
	envelope = envelope.Normalize()
	priority := priorityForBinding(adapter.ChannelBindings, envelope.Channel)
	plan := PropagationPlan{
		Envelope:		envelope,
		Priority:		priority,
		ConsensusReserveBps:	adapter.Bandwidth.ConsensusReserveBps,
	}
	if isCometBFTPassthroughChannel(envelope.Channel) {
		plan.HandledByCometBFT = true
		return plan, nil
	}
	plan.UsesAdvisoryPeerMetric = true
	plan.AdapterFanout = calculateFanout(adapter.Fanout, peerCount, score.ScoreBps)
	return plan, nil
}

func ValidatePeerRoleAdvertisement(adapter AetherNetworkingAdapter, discovery DiscoveryRecord, commitments []RoleCommitment, networkSalt []byte, currentHeight uint64) ([]RoleScope, error) {
	if err := ValidateAetherNetworkingAdapter(adapter); err != nil {
		return nil, err
	}
	if err := discovery.Validate(networkSalt, currentHeight); err != nil {
		return nil, err
	}
	return RoleScopes(discovery.Record, commitments, currentHeight)
}

func calculateFanout(policy AdaptiveFanoutPolicy, peerCount uint32, scoreBps uint32) uint32 {
	if peerCount <= policy.MinFanout {
		return peerCount
	}
	if scoreBps < policy.PeerScoreFloorBps {
		return policy.MinFanout
	}
	fanout := policy.MinFanout + uint32((uint64(policy.MaxFanout-policy.MinFanout)*uint64(scoreBps))/uint64(BasisPoints))
	if fanout > policy.MaxFanout {
		fanout = policy.MaxFanout
	}
	if fanout > peerCount {
		return peerCount
	}
	return fanout
}

func defaultMultiplexBindings(policies []ChannelPolicy) []MultiplexBinding {
	out := make([]MultiplexBinding, len(policies))
	for i, policy := range policies {
		out[i] = MultiplexBinding{
			Channel:		policy.Channel,
			PhysicalTransport:	BaseTransportCometBFTP2P,
			CometBFTPassthrough:	isCometBFTPassthroughChannel(policy.Channel),
			Priority:		policy.Priority,
		}
	}
	return out
}

func validateBaseCapabilities(capabilities []BaseTransportCapability) error {
	required := []BaseTransportCapability{
		BaseCapabilityConsensusGossip,
		BaseCapabilityProposalVotes,
		BaseCapabilityBlockPropagation,
		BaseCapabilityMempoolPropagation,
		BaseCapabilityValidatorCoordination,
		BaseCapabilityPeerTransport,
	}
	for _, capability := range required {
		if !hasBaseCapability(capabilities, capability) {
			return fmt.Errorf("networking CometBFT base transport missing capability %s", capability)
		}
	}
	return nil
}

func validateANAResponsibilities(responsibilities []ANAResponsibility) error {
	required := []ANAResponsibility{
		ANAResponsibilityPeerScoring,
		ANAResponsibilityConnectionMultiplex,
		ANAResponsibilityAdaptiveFanout,
		ANAResponsibilityBandwidthAware,
		ANAResponsibilityZoneHints,
		ANAResponsibilityMessagePriority,
		ANAResponsibilityStreamingNegotiation,
		ANAResponsibilityRoleValidation,
	}
	for _, responsibility := range required {
		if !hasANAResponsibility(responsibilities, responsibility) {
			return fmt.Errorf("networking ANA missing responsibility %s", responsibility)
		}
	}
	return nil
}

func validateMultiplexBindings(bindings []MultiplexBinding) error {
	if len(bindings) == 0 {
		return errors.New("networking ANA multiplex bindings are required")
	}
	seen := make(map[ChannelClass]struct{}, len(bindings))
	for _, binding := range bindings {
		if !IsChannelClass(binding.Channel) {
			return fmt.Errorf("unknown networking ANA channel %q", binding.Channel)
		}
		if strings.TrimSpace(binding.PhysicalTransport) != BaseTransportCometBFTP2P {
			return errors.New("networking ANA channel binding must use CometBFT P2P physical transport")
		}
		if _, found := seen[binding.Channel]; found {
			return errors.New("networking ANA duplicate channel binding")
		}
		seen[binding.Channel] = struct{}{}
		if isCometBFTPassthroughChannel(binding.Channel) && !binding.CometBFTPassthrough {
			return errors.New("networking ANA consensus, block, mempool, and state-sync channels must pass through CometBFT")
		}
	}
	for _, policy := range DefaultChannelPolicies() {
		if _, found := seen[policy.Channel]; !found {
			return fmt.Errorf("networking ANA missing channel binding %s", policy.Channel)
		}
	}
	if priorityForBinding(bindings, ChannelConsensus) >= priorityForBinding(bindings, ChannelService) {
		return errors.New("networking ANA consensus priority must outrank service traffic")
	}
	if priorityForBinding(bindings, ChannelConsensus) >= priorityForBinding(bindings, ChannelData) {
		return errors.New("networking ANA consensus priority must outrank bulk data")
	}
	return nil
}

func isCometBFTPassthroughChannel(channel ChannelClass) bool {
	switch channel {
	case ChannelConsensus, ChannelBlock, ChannelMempool, ChannelStateSync:
		return true
	default:
		return false
	}
}

func priorityForBinding(bindings []MultiplexBinding, channel ChannelClass) uint32 {
	for _, binding := range bindings {
		if binding.Channel == channel {
			return binding.Priority
		}
	}
	return PriorityForChannel(channel)
}

func hasBaseCapability(capabilities []BaseTransportCapability, required BaseTransportCapability) bool {
	for _, capability := range capabilities {
		if capability == required {
			return true
		}
	}
	return false
}

func hasANAResponsibility(responsibilities []ANAResponsibility, required ANAResponsibility) bool {
	for _, responsibility := range responsibilities {
		if responsibility == required {
			return true
		}
	}
	return false
}

func cloneAdapter(adapter AetherNetworkingAdapter) AetherNetworkingAdapter {
	adapter.BaseCapabilities = append([]BaseTransportCapability(nil), adapter.BaseCapabilities...)
	adapter.Responsibilities = append([]ANAResponsibility(nil), adapter.Responsibilities...)
	adapter.ChannelBindings = append([]MultiplexBinding(nil), adapter.ChannelBindings...)
	adapter.ZoneHints = append([]ZoneRoutingHint(nil), adapter.ZoneHints...)
	return adapter
}
