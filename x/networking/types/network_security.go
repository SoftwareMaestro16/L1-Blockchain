package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultSecurityReplayHorizon      = uint64(10_000)
	DefaultMaxInvalidMessagesPerEpoch = uint64(16)
	DefaultMaxBytesPerEpoch           = uint64(64 << 20)
	DefaultMaxDelayedBlockHeights     = uint64(2)
	DefaultMinPeerDiversityBps        = uint32(4_000)
	DefaultSybilScoreThresholdBps     = uint32(2_500)
)

type NetworkThreat string

const (
	ThreatMaliciousPeer             NetworkThreat = "malicious_peer"
	ThreatEclipseAttack             NetworkThreat = "eclipse_attack"
	ThreatRoutingManipulation       NetworkThreat = "routing_manipulation"
	ThreatSpamFlood                 NetworkThreat = "spam_flood"
	ThreatDiscoveryPoisoning        NetworkThreat = "discovery_poisoning"
	ThreatServiceAdvertisementForge NetworkThreat = "service_advertisement_forgery"
	ThreatChunkCorruption           NetworkThreat = "chunk_corruption"
	ThreatBandwidthExhaustion       NetworkThreat = "bandwidth_exhaustion"
	ThreatSybilPeers                NetworkThreat = "sybil_peers"
	ThreatWithheldBlockChunks       NetworkThreat = "withheld_block_chunks"
	ThreatCrossZoneReplay           NetworkThreat = "cross_zone_message_replay"
)

type SecurityControl string

const (
	ControlPeerReputation        SecurityControl = "peer_reputation_scoring"
	ControlAdaptivePeerRotation  SecurityControl = "adaptive_peer_rotation"
	ControlChannelBinding        SecurityControl = "cryptographic_channel_binding"
	ControlMessageAuthentication SecurityControl = "message_authentication"
	ControlReplayProtection      SecurityControl = "deterministic_replay_protection"
	ControlOverlayIsolation      SecurityControl = "overlay_isolation"
	ControlSignedDiscovery       SecurityControl = "signed_discovery_records"
	ControlExpiringAds           SecurityControl = "expiring_advertisements"
	ControlHashDedup             SecurityControl = "hash_based_deduplication"
	ControlChunkMerkle           SecurityControl = "chunk_merkle_verification"
	ControlRateLimits            SecurityControl = "per_channel_rate_limits"
	ControlQoSIsolation          SecurityControl = "qos_isolation"
)

type PeerSecurityObservation struct {
	PeerNodeID             string
	InvalidMessages        uint64
	DuplicateMessages      uint64
	ConflictingBroadcasts  uint64
	CorruptChunks          uint64
	ExpiredAdvertisements  uint64
	ForgedAdvertisements   uint64
	BytesThisEpoch         uint64
	SybilClusterPeers      uint32
	DelayedBlockChunkCount uint32
	CrossZoneReplayCount   uint64
	LastObservedHeight     uint64
}

type NetworkSecurityPolicy struct {
	ReplayHorizon            uint64
	MaxInvalidMessages       uint64
	MaxBytesPerEpoch         uint64
	MaxDelayedBlockHeights   uint64
	MinPeerDiversityBps      uint32
	SybilScoreThresholdBps   uint32
	ChannelRateLimits        []ChannelRateLimit
	RequiredControls         []SecurityControl
	ConsensusChannelRequired bool
}

type ChannelRateLimit struct {
	Channel        ChannelClass
	MaxBytes       uint64
	WindowHeight   uint64
	DropOnExceeded bool
}

type ChannelRateUsage struct {
	Channel     ChannelClass
	Bytes       uint64
	WindowStart uint64
	WindowEnd   uint64
}

type NetworkSecurityDecision struct {
	PeerNodeID        string
	Accepted          bool
	Quarantine        bool
	RotatePeer        bool
	DropMessage       bool
	DowngradeQoS      bool
	ConsensusIsolated bool
	Score             PeerScore
	Threats           []NetworkThreat
	Controls          []SecurityControl
	Reason            string
}

type ReplayProtectionEntry struct {
	ReplayID string
	Height   uint64
}

type ReplayProtectionCache struct {
	Horizon uint64
	Entries []ReplayProtectionEntry
}

type PeerDiversityReport struct {
	TotalPeers          uint32
	UniqueNodeIDs       uint32
	UniqueAddressHashes uint32
	UniqueZoneSets      uint32
	DiversityBps        uint32
	EclipseRisk         bool
	SybilRisk           bool
}

func DefaultNetworkSecurityPolicy() NetworkSecurityPolicy {
	return NetworkSecurityPolicy{
		ReplayHorizon:            DefaultSecurityReplayHorizon,
		MaxInvalidMessages:       DefaultMaxInvalidMessagesPerEpoch,
		MaxBytesPerEpoch:         DefaultMaxBytesPerEpoch,
		MaxDelayedBlockHeights:   DefaultMaxDelayedBlockHeights,
		MinPeerDiversityBps:      DefaultMinPeerDiversityBps,
		SybilScoreThresholdBps:   DefaultSybilScoreThresholdBps,
		ChannelRateLimits:        DefaultChannelRateLimits(),
		RequiredControls:         DefaultSecurityControls(),
		ConsensusChannelRequired: true,
	}
}

func DefaultSecurityControls() []SecurityControl {
	return []SecurityControl{
		ControlPeerReputation,
		ControlAdaptivePeerRotation,
		ControlChannelBinding,
		ControlMessageAuthentication,
		ControlReplayProtection,
		ControlOverlayIsolation,
		ControlSignedDiscovery,
		ControlExpiringAds,
		ControlHashDedup,
		ControlChunkMerkle,
		ControlRateLimits,
		ControlQoSIsolation,
	}
}

func DefaultChannelRateLimits() []ChannelRateLimit {
	return []ChannelRateLimit{
		{Channel: ChannelConsensus, MaxBytes: 16 << 20, WindowHeight: 1},
		{Channel: ChannelBlock, MaxBytes: 64 << 20, WindowHeight: 1},
		{Channel: ChannelExecution, MaxBytes: 16 << 20, WindowHeight: 1},
		{Channel: ChannelDiscovery, MaxBytes: 2 << 20, WindowHeight: 1, DropOnExceeded: true},
		{Channel: ChannelService, MaxBytes: 8 << 20, WindowHeight: 1, DropOnExceeded: true},
		{Channel: ChannelData, MaxBytes: 128 << 20, WindowHeight: 1},
	}
}

func (p NetworkSecurityPolicy) Normalize() NetworkSecurityPolicy {
	if p.ReplayHorizon == 0 {
		p.ReplayHorizon = DefaultSecurityReplayHorizon
	}
	if p.MaxInvalidMessages == 0 {
		p.MaxInvalidMessages = DefaultMaxInvalidMessagesPerEpoch
	}
	if p.MaxBytesPerEpoch == 0 {
		p.MaxBytesPerEpoch = DefaultMaxBytesPerEpoch
	}
	if p.MaxDelayedBlockHeights == 0 {
		p.MaxDelayedBlockHeights = DefaultMaxDelayedBlockHeights
	}
	if p.MinPeerDiversityBps == 0 {
		p.MinPeerDiversityBps = DefaultMinPeerDiversityBps
	}
	if p.SybilScoreThresholdBps == 0 {
		p.SybilScoreThresholdBps = DefaultSybilScoreThresholdBps
	}
	if len(p.ChannelRateLimits) == 0 {
		p.ChannelRateLimits = DefaultChannelRateLimits()
	}
	if len(p.RequiredControls) == 0 {
		p.RequiredControls = DefaultSecurityControls()
	}
	p.RequiredControls = normalizeSecurityControls(p.RequiredControls)
	return p
}

func (p NetworkSecurityPolicy) Validate() error {
	policy := p.Normalize()
	if policy.ReplayHorizon == 0 {
		return errors.New("networking security replay horizon is required")
	}
	if policy.MaxBytesPerEpoch == 0 {
		return errors.New("networking security max bytes per epoch is required")
	}
	if policy.MinPeerDiversityBps > BasisPoints || policy.SybilScoreThresholdBps > BasisPoints {
		return fmt.Errorf("networking security bps policy values must be <= %d", BasisPoints)
	}
	if len(policy.RequiredControls) == 0 {
		return errors.New("networking security controls are required")
	}
	seenLimits := make(map[ChannelClass]struct{}, len(policy.ChannelRateLimits))
	for _, limit := range policy.ChannelRateLimits {
		if err := limit.Validate(); err != nil {
			return err
		}
		if _, found := seenLimits[limit.Channel]; found {
			return errors.New("networking security duplicate channel rate limit")
		}
		seenLimits[limit.Channel] = struct{}{}
	}
	required := securityControlSet(policy.RequiredControls)
	for _, control := range DefaultSecurityControls() {
		if _, found := required[control]; !found {
			return fmt.Errorf("networking security missing required control %q", control)
		}
	}
	return nil
}

func (l ChannelRateLimit) Validate() error {
	if !IsChannelClass(l.Channel) {
		return fmt.Errorf("unknown networking security rate limit channel %q", l.Channel)
	}
	if l.MaxBytes == 0 {
		return errors.New("networking security channel rate limit bytes are required")
	}
	if l.WindowHeight == 0 {
		return errors.New("networking security channel rate limit window is required")
	}
	return nil
}

func EvaluateNetworkSecurity(score PeerScore, observation PeerSecurityObservation, policy NetworkSecurityPolicy) (NetworkSecurityDecision, error) {
	policy = policy.Normalize()
	if err := policy.Validate(); err != nil {
		return NetworkSecurityDecision{}, err
	}
	if err := ValidateHash("networking security peer node id", normalizeHashText(observation.PeerNodeID)); err != nil {
		return NetworkSecurityDecision{}, err
	}
	if score.ScoreBps > BasisPoints {
		return NetworkSecurityDecision{}, fmt.Errorf("networking security peer score must be <= %d bps", BasisPoints)
	}
	threats := DetectNetworkThreats(observation, policy)
	nextScore := ApplySecurityReputation(score, observation, policy)
	decision := NetworkSecurityDecision{
		PeerNodeID:        normalizeHashText(observation.PeerNodeID),
		Accepted:          true,
		ConsensusIsolated: true,
		Score:             nextScore,
		Threats:           threats,
		Controls:          policy.RequiredControls,
	}
	if hasThreat(threats, ThreatSpamFlood) || hasThreat(threats, ThreatBandwidthExhaustion) {
		decision.DropMessage = true
		decision.DowngradeQoS = true
		decision.Reason = appendSecurityReason(decision.Reason, "rate limit exceeded")
	}
	if hasThreat(threats, ThreatChunkCorruption) || hasThreat(threats, ThreatServiceAdvertisementForge) || hasThreat(threats, ThreatCrossZoneReplay) {
		decision.Quarantine = true
		decision.Accepted = false
		decision.Reason = appendSecurityReason(decision.Reason, "cryptographic fault evidence")
	}
	if hasThreat(threats, ThreatSybilPeers) || hasThreat(threats, ThreatEclipseAttack) || nextScore.ScoreBps < policy.SybilScoreThresholdBps {
		decision.RotatePeer = true
		decision.DowngradeQoS = true
		decision.Reason = appendSecurityReason(decision.Reason, "peer rotation required")
	}
	sortThreats(decision.Threats)
	return decision, nil
}

func DetectNetworkThreats(observation PeerSecurityObservation, policy NetworkSecurityPolicy) []NetworkThreat {
	policy = policy.Normalize()
	threats := make([]NetworkThreat, 0, 8)
	if observation.InvalidMessages > policy.MaxInvalidMessages {
		threats = append(threats, ThreatMaliciousPeer)
	}
	if observation.DuplicateMessages > policy.MaxInvalidMessages {
		threats = append(threats, ThreatSpamFlood)
	}
	if observation.ConflictingBroadcasts > 0 {
		threats = append(threats, ThreatRoutingManipulation)
	}
	if observation.CorruptChunks > 0 {
		threats = append(threats, ThreatChunkCorruption)
	}
	if observation.ExpiredAdvertisements > policy.MaxInvalidMessages {
		threats = append(threats, ThreatDiscoveryPoisoning)
	}
	if observation.ForgedAdvertisements > 0 {
		threats = append(threats, ThreatServiceAdvertisementForge)
	}
	if observation.BytesThisEpoch > policy.MaxBytesPerEpoch {
		threats = append(threats, ThreatBandwidthExhaustion)
	}
	if observation.SybilClusterPeers > 0 {
		threats = append(threats, ThreatSybilPeers)
	}
	if observation.DelayedBlockChunkCount > 0 {
		threats = append(threats, ThreatWithheldBlockChunks)
	}
	if observation.CrossZoneReplayCount > 0 {
		threats = append(threats, ThreatCrossZoneReplay)
	}
	return uniqueThreats(threats)
}

func ApplySecurityReputation(score PeerScore, observation PeerSecurityObservation, policy NetworkSecurityPolicy) PeerScore {
	policy = policy.Normalize()
	penalty := uint64(score.PenaltyBps)
	penalty += minSecurityUint64(observation.InvalidMessages*250, uint64(BasisPoints))
	penalty += minSecurityUint64(observation.DuplicateMessages*100, uint64(BasisPoints))
	penalty += minSecurityUint64(observation.ConflictingBroadcasts*1_000, uint64(BasisPoints))
	penalty += minSecurityUint64(observation.CorruptChunks*2_000, uint64(BasisPoints))
	penalty += minSecurityUint64(observation.ForgedAdvertisements*2_000, uint64(BasisPoints))
	penalty += minSecurityUint64(observation.CrossZoneReplayCount*2_000, uint64(BasisPoints))
	penalty += minSecurityUint64(uint64(observation.SybilClusterPeers)*500, uint64(BasisPoints))
	if observation.BytesThisEpoch > policy.MaxBytesPerEpoch {
		penalty += 1_000
	}
	if penalty > uint64(BasisPoints) {
		penalty = uint64(BasisPoints)
	}
	out := score
	out.PenaltyBps = uint32(penalty)
	if out.ScoreBps > out.PenaltyBps {
		out.ScoreBps -= out.PenaltyBps
	} else {
		out.ScoreBps = 0
	}
	return out
}

func NewReplayProtectionCache(horizon uint64) ReplayProtectionCache {
	if horizon == 0 {
		horizon = DefaultSecurityReplayHorizon
	}
	return ReplayProtectionCache{Horizon: horizon}
}

func (c ReplayProtectionCache) Accept(replayID string, height uint64) (ReplayProtectionCache, bool, error) {
	replayID = normalizeHashText(replayID)
	if err := ValidateHash("networking replay id", replayID); err != nil {
		return ReplayProtectionCache{}, false, err
	}
	if height == 0 {
		return ReplayProtectionCache{}, false, errors.New("networking replay height is required")
	}
	next := c.Prune(height)
	if next.Horizon == 0 {
		next.Horizon = DefaultSecurityReplayHorizon
	}
	for _, entry := range next.Entries {
		if entry.ReplayID == replayID {
			return next, false, nil
		}
	}
	next.Entries = append(next.Entries, ReplayProtectionEntry{ReplayID: replayID, Height: height})
	sort.SliceStable(next.Entries, func(i, j int) bool {
		if next.Entries[i].Height != next.Entries[j].Height {
			return next.Entries[i].Height < next.Entries[j].Height
		}
		return next.Entries[i].ReplayID < next.Entries[j].ReplayID
	})
	return next, true, nil
}

func (c ReplayProtectionCache) Prune(currentHeight uint64) ReplayProtectionCache {
	next := ReplayProtectionCache{Horizon: c.Horizon}
	if next.Horizon == 0 {
		next.Horizon = DefaultSecurityReplayHorizon
	}
	if currentHeight == 0 {
		next.Entries = append([]ReplayProtectionEntry(nil), c.Entries...)
		return next
	}
	minHeight := uint64(0)
	if currentHeight > next.Horizon {
		minHeight = currentHeight - next.Horizon
	}
	for _, entry := range c.Entries {
		if entry.Height >= minHeight {
			next.Entries = append(next.Entries, entry)
		}
	}
	return next
}

func CheckChannelRateLimit(limit ChannelRateLimit, usage ChannelRateUsage) (bool, error) {
	if err := limit.Validate(); err != nil {
		return false, err
	}
	if usage.Channel != limit.Channel {
		return false, errors.New("networking security rate usage channel mismatch")
	}
	if usage.WindowEnd < usage.WindowStart {
		return false, errors.New("networking security rate usage window is invalid")
	}
	if usage.WindowEnd-usage.WindowStart+1 > limit.WindowHeight {
		return false, errors.New("networking security rate usage exceeds accounting window")
	}
	return usage.Bytes <= limit.MaxBytes, nil
}

func ValidateSecurityChannelBinding(session SessionChannel, remote NodeRecord, streamID string) error {
	if err := session.Validate(); err != nil {
		return err
	}
	remote = NormalizeNodeRecord(remote)
	if remote.NodeID != session.RemoteNodeID {
		return errors.New("networking security channel binding remote node mismatch")
	}
	if err := ValidateHash("networking security remote node id", remote.NodeID); err != nil {
		return err
	}
	streamID = strings.TrimSpace(streamID)
	for _, stream := range session.Streams {
		if stream.StreamID == streamID {
			if stream.EncryptionContext != streamEncryptionContext(session.SessionKeys.KeyID, stream.StreamID) {
				return errors.New("networking security stream encryption context mismatch")
			}
			return nil
		}
	}
	return errors.New("networking security stream not bound to session")
}

func ValidateOverlayIsolation(desc OverlayDescriptor, message AetherMeshMessage) error {
	if err := desc.ValidateBasic(); err != nil {
		return err
	}
	message = NormalizeAetherMeshMessage(message)
	if message.OverlayID != desc.OverlayID {
		return errors.New("networking security overlay isolation mismatch")
	}
	switch desc.OverlayType {
	case OverlayTypeZone, OverlayTypeExecution:
		if message.SourceZone == "" && message.DestinationZone == "" {
			return errors.New("networking security zone overlay requires zone-scoped message")
		}
	case OverlayTypeService:
		if message.Type != MeshMessageService && message.Type != MeshMessageQuery {
			return errors.New("networking security service overlay cannot carry non-service traffic")
		}
	case OverlayTypeValidator:
		if message.Type != MeshMessageConsensus {
			return errors.New("networking security validator overlay cannot carry non-consensus traffic")
		}
	}
	return nil
}

func ValidateSecurityDiscoveryRecord(record DiscoveryRecord, networkSalt []byte, currentHeight uint64) error {
	if err := ValidateSignedDiscoveryRecord(record, networkSalt, currentHeight); err != nil {
		return err
	}
	if record.ExpiresHeight <= currentHeight {
		return errors.New("networking security discovery record expired")
	}
	return nil
}

func VerifySecurityChunk(transfer RL2Transfer, descriptor RL2ChunkDescriptor, chunk PayloadChunk) error {
	if err := VerifyRL2Chunk(transfer, descriptor, chunk); err != nil {
		return err
	}
	if err := VerifyRL2ChunkProof(descriptor, transfer.PayloadRoot, transfer.ChunkCount); err != nil {
		return err
	}
	return nil
}

func DetectWithheldBlockChunks(session BlockPropagationSession, currentHeight uint64, policy NetworkSecurityPolicy) (bool, error) {
	policy = policy.Normalize()
	if err := session.Header.Validate(0); err != nil {
		return false, err
	}
	if currentHeight < session.Header.Height {
		return false, errors.New("networking security block chunk height regression")
	}
	if uint32(len(session.ReceivedChunks)) >= session.Header.ChunkCount {
		return false, nil
	}
	return currentHeight-session.Header.Height > policy.MaxDelayedBlockHeights, nil
}

func EvaluatePeerDiversity(peers []NodeRecord, policy NetworkSecurityPolicy) (PeerDiversityReport, error) {
	policy = policy.Normalize()
	if len(peers) == 0 {
		return PeerDiversityReport{}, errors.New("networking security peers are required")
	}
	nodeIDs := make(map[string]struct{}, len(peers))
	addresses := make(map[string]struct{}, len(peers))
	zoneSets := make(map[string]struct{}, len(peers))
	for _, peer := range peers {
		peer = NormalizeNodeRecord(peer)
		if err := ValidateHash("networking security peer node id", peer.NodeID); err != nil {
			return PeerDiversityReport{}, err
		}
		nodeIDs[peer.NodeID] = struct{}{}
		if peer.NetworkAddressesHash != "" {
			addresses[peer.NetworkAddressesHash] = struct{}{}
		}
		zoneSets[strings.Join(peer.ZonesSupported, ",")] = struct{}{}
	}
	total := uint32(len(peers))
	diversity := uint32(len(nodeIDs)) * uint32(BasisPoints) / total
	addressDiversity := uint32(len(addresses)) * uint32(BasisPoints) / total
	if len(addresses) > 0 && addressDiversity < diversity {
		diversity = addressDiversity
	}
	report := PeerDiversityReport{
		TotalPeers:          total,
		UniqueNodeIDs:       uint32(len(nodeIDs)),
		UniqueAddressHashes: uint32(len(addresses)),
		UniqueZoneSets:      uint32(len(zoneSets)),
		DiversityBps:        diversity,
		EclipseRisk:         diversity < policy.MinPeerDiversityBps,
		SybilRisk:           uint32(len(nodeIDs)) < total,
	}
	return report, nil
}

func ValidateSecurityQoSIsolation(policies []QoSClassPolicy) error {
	if err := ValidateQoSClassPolicies(policies); err != nil {
		return err
	}
	consensus, found := findQoSClassPolicy(policies, QoSClassCriticalConsensus)
	if !found || !consensus.ReservedCapacity || consensus.BandwidthFloorBps == 0 {
		return errors.New("networking security consensus qos isolation requires reserved capacity")
	}
	execution, found := findQoSClassPolicy(policies, QoSClassExecutionMessage)
	if !found || priorityForQoSClass(policies, QoSClassBulkData) <= execution.Priority {
		return errors.New("networking security bulk traffic cannot outrank execution")
	}
	bulk, found := findQoSClassPolicy(policies, QoSClassBulkData)
	if !found || !bulk.Backpressure {
		return errors.New("networking security bulk qos isolation requires backpressure")
	}
	return nil
}

func securityControlSet(controls []SecurityControl) map[SecurityControl]struct{} {
	out := make(map[SecurityControl]struct{}, len(controls))
	for _, control := range normalizeSecurityControls(controls) {
		out[control] = struct{}{}
	}
	return out
}

func normalizeSecurityControls(controls []SecurityControl) []SecurityControl {
	out := make([]SecurityControl, 0, len(controls))
	seen := make(map[SecurityControl]struct{}, len(controls))
	for _, control := range controls {
		normalized := SecurityControl(strings.ToLower(strings.TrimSpace(string(control))))
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func uniqueThreats(threats []NetworkThreat) []NetworkThreat {
	out := make([]NetworkThreat, 0, len(threats))
	seen := make(map[NetworkThreat]struct{}, len(threats))
	for _, threat := range threats {
		normalized := NetworkThreat(strings.ToLower(strings.TrimSpace(string(threat))))
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sortThreats(out)
	return out
}

func sortThreats(threats []NetworkThreat) {
	sort.SliceStable(threats, func(i, j int) bool {
		return threats[i] < threats[j]
	})
}

func hasThreat(threats []NetworkThreat, target NetworkThreat) bool {
	for _, threat := range threats {
		if threat == target {
			return true
		}
	}
	return false
}

func appendSecurityReason(existing, next string) string {
	if existing == "" {
		return next
	}
	return existing + "; " + next
}

func minSecurityUint64(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}
