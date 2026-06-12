package types

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultSecurityReplayHorizon		= uint64(10_000)
	DefaultMaxInvalidMessagesPerEpoch	= uint64(16)
	DefaultMaxBytesPerEpoch			= uint64(64 << 20)
	DefaultMaxDelayedBlockHeights		= uint64(2)
	DefaultMinPeerDiversityBps		= uint32(4_000)
	DefaultSybilScoreThresholdBps		= uint32(2_500)
	DefaultMaxPeersPerIdentityCluster	= uint32(2)
	DefaultReputationDecayBps		= uint32(500)
	DefaultMaxPeerMessagesPerWindow		= uint64(128)
	DefaultMaxHandshakeCostUnits		= uint64(10_000)
	DefaultMaxChunkRequestsPerWindow	= uint32(256)
	DefaultMinServiceStakeWeight		= uint64(1_000)
)

type NetworkThreat string

const (
	ThreatMaliciousPeer		NetworkThreat	= "malicious_peer"
	ThreatEclipseAttack		NetworkThreat	= "eclipse_attack"
	ThreatRoutingManipulation	NetworkThreat	= "routing_manipulation"
	ThreatSpamFlood			NetworkThreat	= "spam_flood"
	ThreatDiscoveryPoisoning	NetworkThreat	= "discovery_poisoning"
	ThreatServiceAdvertisementForge	NetworkThreat	= "service_advertisement_forgery"
	ThreatChunkCorruption		NetworkThreat	= "chunk_corruption"
	ThreatBandwidthExhaustion	NetworkThreat	= "bandwidth_exhaustion"
	ThreatSybilPeers		NetworkThreat	= "sybil_peers"
	ThreatWithheldBlockChunks	NetworkThreat	= "withheld_block_chunks"
	ThreatCrossZoneReplay		NetworkThreat	= "cross_zone_message_replay"
)

type SecurityControl string

const (
	ControlPeerReputation		SecurityControl	= "peer_reputation_scoring"
	ControlAdaptivePeerRotation	SecurityControl	= "adaptive_peer_rotation"
	ControlChannelBinding		SecurityControl	= "cryptographic_channel_binding"
	ControlMessageAuthentication	SecurityControl	= "message_authentication"
	ControlReplayProtection		SecurityControl	= "deterministic_replay_protection"
	ControlOverlayIsolation		SecurityControl	= "overlay_isolation"
	ControlSignedDiscovery		SecurityControl	= "signed_discovery_records"
	ControlExpiringAds		SecurityControl	= "expiring_advertisements"
	ControlHashDedup		SecurityControl	= "hash_based_deduplication"
	ControlChunkMerkle		SecurityControl	= "chunk_merkle_verification"
	ControlRateLimits		SecurityControl	= "per_channel_rate_limits"
	ControlQoSIsolation		SecurityControl	= "qos_isolation"
)

type PeerSecurityObservation struct {
	PeerNodeID		string
	InvalidMessages		uint64
	DuplicateMessages	uint64
	ConflictingBroadcasts	uint64
	CorruptChunks		uint64
	ExpiredAdvertisements	uint64
	ForgedAdvertisements	uint64
	BytesThisEpoch		uint64
	SybilClusterPeers	uint32
	DelayedBlockChunkCount	uint32
	CrossZoneReplayCount	uint64
	LastObservedHeight	uint64
}

type NetworkSecurityPolicy struct {
	ReplayHorizon			uint64
	MaxInvalidMessages		uint64
	MaxBytesPerEpoch		uint64
	MaxDelayedBlockHeights		uint64
	MaxPeerMessagesPerWindow	uint64
	MaxHandshakeCostUnits		uint64
	MaxChunkRequestsPerWindow	uint32
	MaxPayloadBytes			uint64
	MinServiceStakeWeight		uint64
	MinPeerDiversityBps		uint32
	SybilScoreThresholdBps		uint32
	ChannelRateLimits		[]ChannelRateLimit
	RequiredControls		[]SecurityControl
	ConsensusChannelRequired	bool
}

type ChannelRateLimit struct {
	Channel		ChannelClass
	MaxBytes	uint64
	WindowHeight	uint64
	DropOnExceeded	bool
}

type ChannelRateUsage struct {
	Channel		ChannelClass
	Bytes		uint64
	WindowStart	uint64
	WindowEnd	uint64
}

type NetworkSecurityDecision struct {
	PeerNodeID		string
	Accepted		bool
	Quarantine		bool
	RotatePeer		bool
	DropMessage		bool
	DowngradeQoS		bool
	ConsensusIsolated	bool
	Score			PeerScore
	Threats			[]NetworkThreat
	Controls		[]SecurityControl
	Reason			string
}

type ReplayProtectionEntry struct {
	ReplayID	string
	Height		uint64
}

type ReplayProtectionCache struct {
	Horizon	uint64
	Entries	[]ReplayProtectionEntry
}

type PeerDiversityReport struct {
	TotalPeers		uint32
	UniqueNodeIDs		uint32
	UniqueAddressHashes	uint32
	UniqueZoneSets		uint32
	DiversityBps		uint32
	EclipseRisk		bool
	SybilRisk		bool
}

type PeerReputationInput struct {
	PeerNodeID			string
	ValidMessages			uint64
	InvalidMessages			uint64
	LatencyMillis			uint64
	ThroughputBytesPerSec		uint64
	CorrectChunks			uint64
	CorruptChunks			uint64
	ValidDiscoveryResponses		uint64
	InvalidDiscoveryResponses	uint64
	ValidServiceResponses		uint64
	InvalidServiceResponses		uint64
	Timeouts			uint64
	DuplicateBroadcasts		uint64
	ConflictingBroadcasts		uint64
	EvidenceHash			string
	EvidenceHeight			uint64
	CommittedEvidence		bool
	UsedForConsensus		bool
	ElapsedEpochs			uint64
	DecayPolicy			PeerScoreDecayPolicy
}

type PeerReputationDecision struct {
	PeerNodeID		string
	Score			PeerScore
	LocalAdvisory		bool
	ConsensusEligible	bool
	PenaltyBps		uint32
	DecayAppliedBps		uint32
	EvidenceHash		string
	Reason			string
}

type EclipseResistancePolicy struct {
	MinRandomPeers			uint32
	MinValidatorPeers		uint32
	MinZonePeers			uint32
	MaxPeersPerIdentityCluster	uint32
	PreferProofBackedRecords	bool
	RotateDiscoverySources		bool
}

type EclipseResistancePlan struct {
	RandomSetMaintained		bool
	ValidatorDiversity		bool
	ZoneDiversity			bool
	DiscoverySourcesRotated		bool
	IdentityClusterLimited		bool
	ProofBackedCriticalRoute	bool
	PeersToDrop			[]string
	DiscoverySources		[]string
	CriticalRoutingPeers		[]string
}

type SignedTransportEnvelope struct {
	Envelope	TransportEnvelope
	SignerNodeID	string
	Height		uint64
	Signature	[]byte
}

type PeerRateUsage struct {
	PeerNodeID	string
	Channel		ChannelClass
	Messages	uint64
	Bytes		uint64
	WindowStart	uint64
	WindowEnd	uint64
}

type PeerRateLimitDecision struct {
	Allowed			bool
	DropMessage		bool
	ThrottlePeer		bool
	ExceededMessages	bool
	ExceededBytes		bool
	Reason			string
}

type HandshakeCostReport struct {
	PeerNodeID		string
	EphemeralKeyBytes	uint64
	ProtocolCount		uint32
	ChannelCount		uint32
	NonceBytes		uint64
	CostUnits		uint64
	Accepted		bool
}

type SpamSimulationResult struct {
	TotalMessages	uint64
	Accepted	uint64
	Dropped		uint64
	Throttled	bool
	Threats		[]NetworkThreat
}

type RoutingManipulationSimulationResult struct {
	ConflictingBroadcasts	uint64
	FaultsDetected		uint64
	Threats			[]NetworkThreat
}

func DefaultNetworkSecurityPolicy() NetworkSecurityPolicy {
	return NetworkSecurityPolicy{
		ReplayHorizon:			DefaultSecurityReplayHorizon,
		MaxInvalidMessages:		DefaultMaxInvalidMessagesPerEpoch,
		MaxBytesPerEpoch:		DefaultMaxBytesPerEpoch,
		MaxDelayedBlockHeights:		DefaultMaxDelayedBlockHeights,
		MaxPeerMessagesPerWindow:	DefaultMaxPeerMessagesPerWindow,
		MaxHandshakeCostUnits:		DefaultMaxHandshakeCostUnits,
		MaxChunkRequestsPerWindow:	DefaultMaxChunkRequestsPerWindow,
		MaxPayloadBytes:		MaxStreamMessageBytes,
		MinServiceStakeWeight:		DefaultMinServiceStakeWeight,
		MinPeerDiversityBps:		DefaultMinPeerDiversityBps,
		SybilScoreThresholdBps:		DefaultSybilScoreThresholdBps,
		ChannelRateLimits:		DefaultChannelRateLimits(),
		RequiredControls:		DefaultSecurityControls(),
		ConsensusChannelRequired:	true,
	}
}

func DefaultEclipseResistancePolicy() EclipseResistancePolicy {
	return EclipseResistancePolicy{
		MinRandomPeers:			DefaultRandomDiversityBucket,
		MinValidatorPeers:		2,
		MinZonePeers:			2,
		MaxPeersPerIdentityCluster:	DefaultMaxPeersPerIdentityCluster,
		PreferProofBackedRecords:	true,
		RotateDiscoverySources:		true,
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
	if p.MaxPeerMessagesPerWindow == 0 {
		p.MaxPeerMessagesPerWindow = DefaultMaxPeerMessagesPerWindow
	}
	if p.MaxHandshakeCostUnits == 0 {
		p.MaxHandshakeCostUnits = DefaultMaxHandshakeCostUnits
	}
	if p.MaxChunkRequestsPerWindow == 0 {
		p.MaxChunkRequestsPerWindow = DefaultMaxChunkRequestsPerWindow
	}
	if p.MaxPayloadBytes == 0 {
		p.MaxPayloadBytes = MaxStreamMessageBytes
	}
	if p.MinServiceStakeWeight == 0 {
		p.MinServiceStakeWeight = DefaultMinServiceStakeWeight
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
	if policy.MaxPayloadBytes == 0 || policy.MaxPayloadBytes > MaxStreamMessageBytes {
		return fmt.Errorf("networking security max payload bytes must be between 1 and %d", MaxStreamMessageBytes)
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

func SignSecurityEnvelope(envelope TransportEnvelope, signer NodeRecord, privateKey ed25519.PrivateKey, height uint64) (SignedTransportEnvelope, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return SignedTransportEnvelope{}, errors.New("networking security envelope private key must be ed25519")
	}
	signer = NormalizeNodeRecord(signer)
	if !ed25519.PrivateKey(privateKey).Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(signer.NodePubKey)) {
		return SignedTransportEnvelope{}, errors.New("networking security envelope signer key mismatch")
	}
	envelope = envelope.Normalize()
	if err := envelope.Validate(DefaultChannelPolicies()); err != nil {
		return SignedTransportEnvelope{}, err
	}
	if height == 0 {
		return SignedTransportEnvelope{}, errors.New("networking security envelope height is required")
	}
	signed := SignedTransportEnvelope{Envelope: envelope, SignerNodeID: signer.NodeID, Height: height}
	payload, err := signed.SigningPayload()
	if err != nil {
		return SignedTransportEnvelope{}, err
	}
	signed.Signature = ed25519.Sign(privateKey, payload)
	return signed, signed.Validate(signer.NodePubKey)
}

func (s SignedTransportEnvelope) Normalize() SignedTransportEnvelope {
	s.Envelope = s.Envelope.Normalize()
	s.SignerNodeID = normalizeHashText(s.SignerNodeID)
	s.Signature = cloneBytes(s.Signature)
	return s
}

func (s SignedTransportEnvelope) SigningPayload() ([]byte, error) {
	signed := s.Normalize()
	signed.Signature = nil
	return json.Marshal(signed)
}

func (s SignedTransportEnvelope) Validate(pubKey ed25519.PublicKey) error {
	signed := s.Normalize()
	if err := signed.Envelope.Validate(DefaultChannelPolicies()); err != nil {
		return err
	}
	if err := ValidateHash("networking security envelope signer", signed.SignerNodeID); err != nil {
		return err
	}
	if signed.Height == 0 {
		return errors.New("networking security envelope height is required")
	}
	if len(pubKey) != ed25519.PublicKeySize {
		return errors.New("networking security envelope public key must be ed25519")
	}
	if len(signed.Signature) != ed25519.SignatureSize {
		return errors.New("networking security envelope signature is required")
	}
	payload, err := signed.SigningPayload()
	if err != nil {
		return err
	}
	if !ed25519.Verify(pubKey, payload, signed.Signature) {
		return errors.New("networking security envelope signature verification failed")
	}
	return nil
}

func EvaluatePeerRateLimit(policy NetworkSecurityPolicy, usage PeerRateUsage) (PeerRateLimitDecision, error) {
	policy = policy.Normalize()
	if err := policy.Validate(); err != nil {
		return PeerRateLimitDecision{}, err
	}
	usage.PeerNodeID = normalizeHashText(usage.PeerNodeID)
	if err := ValidateHash("networking peer rate peer node id", usage.PeerNodeID); err != nil {
		return PeerRateLimitDecision{}, err
	}
	if !IsChannelClass(usage.Channel) {
		return PeerRateLimitDecision{}, fmt.Errorf("unknown networking peer rate channel %q", usage.Channel)
	}
	if usage.WindowEnd < usage.WindowStart {
		return PeerRateLimitDecision{}, errors.New("networking peer rate window is invalid")
	}
	decision := PeerRateLimitDecision{Allowed: true}
	if usage.Messages > policy.MaxPeerMessagesPerWindow {
		decision.Allowed = false
		decision.DropMessage = true
		decision.ThrottlePeer = true
		decision.ExceededMessages = true
		decision.Reason = appendSecurityReason(decision.Reason, "peer message rate exceeded")
	}
	for _, limit := range policy.ChannelRateLimits {
		if limit.Channel != usage.Channel {
			continue
		}
		ok, err := CheckChannelRateLimit(limit, ChannelRateUsage{
			Channel:	usage.Channel,
			Bytes:		usage.Bytes,
			WindowStart:	usage.WindowStart,
			WindowEnd:	usage.WindowEnd,
		})
		if err != nil {
			return PeerRateLimitDecision{}, err
		}
		if !ok {
			decision.Allowed = false
			decision.ExceededBytes = true
			decision.ThrottlePeer = true
			decision.DropMessage = decision.DropMessage || limit.DropOnExceeded
			decision.Reason = appendSecurityReason(decision.Reason, "channel byte limit exceeded")
		}
	}
	return decision, nil
}

func EvaluateHandshakeCost(req SessionRequest, policy NetworkSecurityPolicy) (HandshakeCostReport, error) {
	policy = policy.Normalize()
	req = req.Normalize()
	if err := ValidateHash("networking handshake cost local node", req.LocalNodeID); err != nil {
		return HandshakeCostReport{}, err
	}
	if err := ValidateHash("networking handshake cost remote node", req.RemoteNodeID); err != nil {
		return HandshakeCostReport{}, err
	}
	if req.HandshakeVersion != DefaultHandshakeVersion {
		return HandshakeCostReport{}, errors.New("networking handshake cost unsupported version")
	}
	if len(req.Nonce) == 0 || len(req.Nonce) > MaxNonceBytes {
		return HandshakeCostReport{}, fmt.Errorf("networking handshake cost nonce must be between 1 and %d bytes", MaxNonceBytes)
	}
	if len(req.LocalEphemeralPubKey) != SessionEphemeralKeyBytes || len(req.RemoteEphemeralPubKey) != SessionEphemeralKeyBytes {
		return HandshakeCostReport{}, fmt.Errorf("networking handshake cost ephemeral public keys must be %d bytes", SessionEphemeralKeyBytes)
	}
	cost := uint64(len(req.LocalEphemeralPubKey)+len(req.RemoteEphemeralPubKey)+len(req.Nonce)) +
		uint64(len(req.ProtocolVersions))*128 +
		uint64(len(req.ChannelClasses))*64 +
		uint64(len(req.CipherSuites))*64
	return HandshakeCostReport{
		PeerNodeID:		req.RemoteNodeID,
		EphemeralKeyBytes:	uint64(len(req.LocalEphemeralPubKey) + len(req.RemoteEphemeralPubKey)),
		ProtocolCount:		uint32(len(req.ProtocolVersions)),
		ChannelCount:		uint32(len(req.ChannelClasses)),
		NonceBytes:		uint64(len(req.Nonce)),
		CostUnits:		cost,
		Accepted:		cost <= policy.MaxHandshakeCostUnits,
	}, nil
}

func ValidatePayloadSize(envelope TransportEnvelope, policy NetworkSecurityPolicy) error {
	policy = policy.Normalize()
	envelope = envelope.Normalize()
	if err := envelope.Validate(DefaultChannelPolicies()); err != nil {
		return err
	}
	if envelope.SizeBytes > policy.MaxPayloadBytes {
		return errors.New("networking security payload size limit exceeded")
	}
	return nil
}

func ValidateChunkRequestLimit(request RL2ChunkRequest, transfer RL2Transfer, policy NetworkSecurityPolicy) error {
	policy = policy.Normalize()
	if err := transfer.Validate(0); err != nil {
		return err
	}
	if normalizeHashText(request.TransferID) != transfer.TransferID {
		return errors.New("networking security chunk request transfer mismatch")
	}
	if len(request.MissingIndexes) == 0 {
		return errors.New("networking security chunk request missing indexes are required")
	}
	if uint32(len(request.MissingIndexes)) > policy.MaxChunkRequestsPerWindow {
		return errors.New("networking security chunk request limit exceeded")
	}
	seen := make(map[uint32]struct{}, len(request.MissingIndexes))
	for _, index := range request.MissingIndexes {
		if index >= transfer.ChunkCount {
			return errors.New("networking security chunk request index out of range")
		}
		if _, found := seen[index]; found {
			return errors.New("networking security duplicate chunk request index")
		}
		seen[index] = struct{}{}
	}
	return nil
}

func ValidateResourceBackedAdvertisement(ad DRTAdvertisement, networkSalt []byte, currentHeight uint64, policy NetworkSecurityPolicy, proofRequired bool) error {
	policy = policy.Normalize()
	if err := ad.Validate(networkSalt, currentHeight); err != nil {
		return err
	}
	switch ad.ObjectType {
	case DRTObjectServiceEndpoint, DRTObjectStorageProvider, DRTObjectRoutingEntryPoint, DRTObjectStreamProvider:
		if ad.StakeWeight < policy.MinServiceStakeWeight {
			return errors.New("networking security advertisement requires stake backed resource")
		}
		if proofRequired && (ad.Discovery.ProofHash == "" || ad.Discovery.ProofHeight == 0) {
			return errors.New("networking security advertisement requires proof backed record")
		}
	}
	return nil
}

func SuppressDuplicateBroadcast(cache BroadcastDedupCache, msg BroadcastMessage, peerNodeID string, currentHeight uint64) (BroadcastDedupCache, bool, error) {
	next, decision, err := cache.Accept(msg, peerNodeID, currentHeight)
	if err != nil {
		return BroadcastDedupCache{}, false, err
	}
	return next, decision.Accepted, nil
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
		PeerNodeID:		normalizeHashText(observation.PeerNodeID),
		Accepted:		true,
		ConsensusIsolated:	true,
		Score:			nextScore,
		Threats:		threats,
		Controls:		policy.RequiredControls,
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

func ComputePeerReputation(input PeerReputationInput) (PeerReputationDecision, error) {
	input.PeerNodeID = normalizeHashText(input.PeerNodeID)
	if err := ValidateHash("networking reputation peer node id", input.PeerNodeID); err != nil {
		return PeerReputationDecision{}, err
	}
	if input.UsedForConsensus {
		input.EvidenceHash = normalizeHashText(input.EvidenceHash)
		if !input.CommittedEvidence {
			return PeerReputationDecision{}, errors.New("networking reputation used for consensus requires committed evidence")
		}
		if err := ValidateHash("networking reputation evidence hash", input.EvidenceHash); err != nil {
			return PeerReputationDecision{}, err
		}
		if input.EvidenceHeight == 0 {
			return PeerReputationDecision{}, errors.New("networking reputation evidence height is required")
		}
	}
	if input.DecayPolicy.MaxDecayBpsPerEpoch == 0 {
		input.DecayPolicy = PeerScoreDecayPolicy{
			MaxDecayBpsPerEpoch:	DefaultReputationDecayBps,
			MinScoreBps:		DefaultPeerScoreDecayFloor,
		}
	}
	if err := input.DecayPolicy.Validate(); err != nil {
		return PeerReputationDecision{}, err
	}
	totalMessages := input.ValidMessages + input.InvalidMessages
	validRate := uint32(BasisPoints)
	if totalMessages > 0 {
		validRate = uint32(input.ValidMessages * uint64(BasisPoints) / totalMessages)
	}
	chunkTotal := input.CorrectChunks + input.CorruptChunks
	chunkCorrectness := uint32(BasisPoints)
	if chunkTotal > 0 {
		chunkCorrectness = uint32(input.CorrectChunks * uint64(BasisPoints) / chunkTotal)
	}
	discoveryTotal := input.ValidDiscoveryResponses + input.InvalidDiscoveryResponses
	discoveryValidity := uint32(BasisPoints)
	if discoveryTotal > 0 {
		discoveryValidity = uint32(input.ValidDiscoveryResponses * uint64(BasisPoints) / discoveryTotal)
	}
	serviceTotal := input.ValidServiceResponses + input.InvalidServiceResponses
	serviceValidity := uint32(BasisPoints)
	if serviceTotal > 0 {
		serviceValidity = uint32(input.ValidServiceResponses * uint64(BasisPoints) / serviceTotal)
	}
	latencyBps := uint32(BasisPoints)
	if input.LatencyMillis > 0 {
		if input.LatencyMillis >= 1_000 {
			latencyBps = 1_000
		} else {
			latencyBps = uint32(uint64(BasisPoints) - input.LatencyMillis*9)
		}
	}
	throughputBps := uint32(0)
	if input.ThroughputBytesPerSec >= 64<<20 {
		throughputBps = BasisPoints
	} else {
		throughputBps = uint32(input.ThroughputBytesPerSec * uint64(BasisPoints) / uint64(64<<20))
	}
	timeoutPenalty := minSecurityUint64(input.Timeouts*400, uint64(BasisPoints))
	duplicatePenalty := minSecurityUint64(input.DuplicateBroadcasts*150, uint64(BasisPoints))
	conflictPenalty := minSecurityUint64(input.ConflictingBroadcasts*1_500, uint64(BasisPoints))
	invalidPenalty := minSecurityUint64(input.InvalidMessages*250, uint64(BasisPoints))
	scoreBps := weightedReputationScore(validRate, latencyBps, throughputBps, chunkCorrectness, discoveryValidity, serviceValidity)
	penalty := timeoutPenalty + duplicatePenalty + conflictPenalty + invalidPenalty
	if penalty > uint64(BasisPoints) {
		penalty = uint64(BasisPoints)
	}
	if uint64(scoreBps) > penalty {
		scoreBps -= uint32(penalty)
	} else {
		scoreBps = 0
	}
	score := PeerScore{
		ScoreBps:	scoreBps,
		LatencyBps:	latencyBps,
		ReliabilityBps:	validRate,
		ThroughputBps:	throughputBps,
		PenaltyBps:	uint32(penalty),
	}
	decayed, err := DecayPeerScore(score, input.ElapsedEpochs, input.DecayPolicy)
	if err != nil {
		return PeerReputationDecision{}, err
	}
	decayApplied := uint32(0)
	if score.ScoreBps > decayed.ScoreBps {
		decayApplied = score.ScoreBps - decayed.ScoreBps
	}
	return PeerReputationDecision{
		PeerNodeID:		input.PeerNodeID,
		Score:			decayed,
		LocalAdvisory:		!input.UsedForConsensus,
		ConsensusEligible:	input.UsedForConsensus && input.CommittedEvidence,
		PenaltyBps:		uint32(penalty),
		DecayAppliedBps:	decayApplied,
		EvidenceHash:		input.EvidenceHash,
	}, nil
}

func ValidateReputationConsensusUse(decision PeerReputationDecision, usedForConsensus bool) error {
	if !usedForConsensus {
		return nil
	}
	if !decision.ConsensusEligible {
		return errors.New("networking reputation is advisory until committed evidence exists")
	}
	if err := ValidateHash("networking reputation decision evidence hash", decision.EvidenceHash); err != nil {
		return err
	}
	return nil
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
		TotalPeers:		total,
		UniqueNodeIDs:		uint32(len(nodeIDs)),
		UniqueAddressHashes:	uint32(len(addresses)),
		UniqueZoneSets:		uint32(len(zoneSets)),
		DiversityBps:		diversity,
		EclipseRisk:		diversity < policy.MinPeerDiversityBps,
		SybilRisk:		uint32(len(nodeIDs)) < total,
	}
	return report, nil
}

func BuildEclipseResistancePlan(graph AdaptiveOverlayGraph, records []DiscoveryRecord, policy EclipseResistancePolicy, routingEpoch uint64) (EclipseResistancePlan, error) {
	graph = NormalizeAdaptiveOverlayGraph(graph)
	if err := ValidateHash("networking eclipse graph local node id", graph.LocalNodeID); err != nil {
		return EclipseResistancePlan{}, err
	}
	if policy.MaxPeersPerIdentityCluster == 0 {
		policy = DefaultEclipseResistancePolicy()
	}
	if routingEpoch == 0 {
		return EclipseResistancePlan{}, errors.New("networking eclipse routing epoch is required")
	}
	allPeers := uniqueAdaptivePeers(graph.FastSet, graph.StableSet, graph.RandomSet, graph.ZoneSet, graph.ServiceSet, graph.FallbackSet)
	if len(allPeers) == 0 {
		return EclipseResistancePlan{}, errors.New("networking eclipse peers are required")
	}
	clusterCounts := make(map[string]uint32, len(allPeers))
	peersToDrop := make([]string, 0)
	for _, peer := range allPeers {
		cluster := peerIdentityCluster(peer)
		clusterCounts[cluster]++
		if clusterCounts[cluster] > policy.MaxPeersPerIdentityCluster {
			peersToDrop = append(peersToDrop, peer.NodeID)
		}
	}
	randomSet := sortAdaptivePeersBySeed(graph.RandomSet, HashParts("eclipse-random", graph.OverlayID, fmt.Sprintf("%d", routingEpoch)))
	discoverySources := adaptivePeerIDsForSecurity(sortAdaptivePeersBySeed(allPeers, HashParts("eclipse-discovery", graph.OverlayID, fmt.Sprintf("%d", routingEpoch))))
	criticalPeers := proofBackedCriticalRoutingPeers(records, policy.PreferProofBackedRecords)
	if len(criticalPeers) == 0 && !policy.PreferProofBackedRecords {
		criticalPeers = discoverySources
	}
	sortStrings(peersToDrop)
	return EclipseResistancePlan{
		RandomSetMaintained:		uint32(len(randomSet)) >= policy.MinRandomPeers && distinctAdaptivePeerBuckets(randomSet) >= minUint32(policy.MinRandomPeers, uint32(len(randomSet))),
		ValidatorDiversity:		countAdaptivePeersWithRole(allPeers, NodeRoleValidator) >= policy.MinValidatorPeers,
		ZoneDiversity:			distinctAdaptiveZones(allPeers) >= policy.MinZonePeers,
		DiscoverySourcesRotated:	policy.RotateDiscoverySources && len(discoverySources) > 0,
		IdentityClusterLimited:		len(peersToDrop) == 0,
		ProofBackedCriticalRoute:	!policy.PreferProofBackedRecords || len(criticalPeers) > 0,
		PeersToDrop:			peersToDrop,
		DiscoverySources:		discoverySources,
		CriticalRoutingPeers:		criticalPeers,
	}, nil
}

func ValidateEclipseResistancePlan(plan EclipseResistancePlan) error {
	if !plan.RandomSetMaintained {
		return errors.New("networking eclipse resistance requires maintained random peer set")
	}
	if !plan.ValidatorDiversity {
		return errors.New("networking eclipse resistance requires validator peer diversity")
	}
	if !plan.ZoneDiversity {
		return errors.New("networking eclipse resistance requires zone peer diversity")
	}
	if !plan.DiscoverySourcesRotated {
		return errors.New("networking eclipse resistance requires rotated discovery sources")
	}
	if !plan.IdentityClusterLimited {
		return errors.New("networking eclipse resistance requires identity cluster limits")
	}
	if !plan.ProofBackedCriticalRoute {
		return errors.New("networking eclipse resistance requires proof backed critical routing")
	}
	return nil
}

func SimulateSpamResistance(policy NetworkSecurityPolicy, usage PeerRateUsage, duplicates []BroadcastMessage, peerNodeID string, currentHeight uint64) (SpamSimulationResult, error) {
	policy = policy.Normalize()
	decision, err := EvaluatePeerRateLimit(policy, usage)
	if err != nil {
		return SpamSimulationResult{}, err
	}
	cache := NewBroadcastDedupCache(DefaultBroadcastDedupHorizon)
	accepted := uint64(0)
	dropped := uint64(0)
	for _, msg := range duplicates {
		var ok bool
		cache, ok, err = SuppressDuplicateBroadcast(cache, msg, peerNodeID, currentHeight)
		if err != nil {
			return SpamSimulationResult{}, err
		}
		if ok {
			accepted++
		} else {
			dropped++
		}
	}
	if !decision.Allowed {
		dropped += usage.Messages
	}
	threats := make([]NetworkThreat, 0, 2)
	if decision.ExceededMessages || dropped > accepted {
		threats = append(threats, ThreatSpamFlood)
	}
	if decision.ExceededBytes {
		threats = append(threats, ThreatBandwidthExhaustion)
	}
	return SpamSimulationResult{
		TotalMessages:	usage.Messages + uint64(len(duplicates)),
		Accepted:	accepted,
		Dropped:	dropped,
		Throttled:	decision.ThrottlePeer,
		Threats:	uniqueThreats(threats),
	}, nil
}

func SimulateRoutingManipulation(msgs []BroadcastMessage, peerNodeID string, currentHeight uint64) (RoutingManipulationSimulationResult, error) {
	cache := NewBroadcastDedupCache(DefaultBroadcastDedupHorizon)
	conflicts := uint64(0)
	for _, msg := range msgs {
		var decision BroadcastDedupDecision
		var err error
		cache, decision, err = cache.Accept(msg, peerNodeID, currentHeight)
		if err != nil {
			return RoutingManipulationSimulationResult{}, err
		}
		if decision.FaultEvidence.EvidenceHash != "" {
			conflicts++
		}
	}
	threats := []NetworkThreat{}
	if conflicts > 0 {
		threats = append(threats, ThreatRoutingManipulation)
	}
	return RoutingManipulationSimulationResult{
		ConflictingBroadcasts:	conflicts,
		FaultsDetected:		uint64(len(cache.Faults)),
		Threats:		threats,
	}, nil
}

func SimulateEclipseResistance(graph AdaptiveOverlayGraph, records []DiscoveryRecord, policy EclipseResistancePolicy, routingEpoch uint64) (EclipseResistancePlan, []NetworkThreat, error) {
	plan, err := BuildEclipseResistancePlan(graph, records, policy, routingEpoch)
	if err != nil {
		return EclipseResistancePlan{}, nil, err
	}
	threats := make([]NetworkThreat, 0, 2)
	if err := ValidateEclipseResistancePlan(plan); err != nil {
		threats = append(threats, ThreatEclipseAttack)
		if !plan.IdentityClusterLimited {
			threats = append(threats, ThreatSybilPeers)
		}
	}
	return plan, uniqueThreats(threats), nil
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

func weightedReputationScore(validRate, latencyBps, throughputBps, chunkCorrectness, discoveryValidity, serviceValidity uint32) uint32 {
	total := uint64(validRate)*3_000 +
		uint64(chunkCorrectness)*2_000 +
		uint64(discoveryValidity)*1_250 +
		uint64(serviceValidity)*1_250 +
		uint64(latencyBps)*1_250 +
		uint64(throughputBps)*1_250
	return uint32(total / uint64(BasisPoints))
}

func peerIdentityCluster(peer AdaptivePeer) string {
	roles := make([]string, len(peer.Roles))
	for i, role := range peer.Roles {
		roles[i] = string(role)
	}
	sortStrings(roles)
	zones := append([]string(nil), peer.ZonesSupported...)
	sortStrings(zones)
	return HashParts("peer-identity-cluster", strings.Join(roles, ","), strings.Join(zones, ","))
}

func countAdaptivePeersWithRole(peers []AdaptivePeer, role NodeRole) uint32 {
	count := uint32(0)
	for _, peer := range peers {
		if hasRole(peer.Roles, role) {
			count++
		}
	}
	return count
}

func distinctAdaptiveZones(peers []AdaptivePeer) uint32 {
	seen := make(map[string]struct{})
	for _, peer := range peers {
		for _, zone := range peer.ZonesSupported {
			zone = strings.TrimSpace(zone)
			if zone != "" {
				seen[zone] = struct{}{}
			}
		}
	}
	return uint32(len(seen))
}

func proofBackedCriticalRoutingPeers(records []DiscoveryRecord, proofRequired bool) []string {
	out := make([]string, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		record = NormalizeDiscoveryRecord(record)
		if proofRequired && (record.ProofHash == "" || record.ProofHeight == 0) {
			continue
		}
		if record.RecordType != DRTObjectRoutingEntryPoint && record.RecordType != DRTObjectOverlayMembershipRecord && record.RecordType != DRTObjectNode {
			continue
		}
		if _, found := seen[record.OwnerNodeID]; found {
			continue
		}
		seen[record.OwnerNodeID] = struct{}{}
		out = append(out, record.OwnerNodeID)
	}
	sortStrings(out)
	return out
}

func adaptivePeerIDsForSecurity(peers []AdaptivePeer) []string {
	out := make([]string, 0, len(peers))
	seen := make(map[string]struct{}, len(peers))
	for _, peer := range peers {
		nodeID := normalizeHashText(peer.NodeID)
		if nodeID == "" {
			continue
		}
		if _, found := seen[nodeID]; found {
			continue
		}
		seen[nodeID] = struct{}{}
		out = append(out, nodeID)
	}
	sortStrings(out)
	return out
}
