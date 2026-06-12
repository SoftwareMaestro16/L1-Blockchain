package types

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	HashHexLength			= 64
	BasisPoints			= uint32(10_000)
	MaxRolesPerNode			= 16
	MaxZonesPerNode			= 128
	MaxServicesPerNode		= 128
	MaxProtocolsPerNode		= 32
	MaxNetworkAddressBytes		= 256
	MaxProtocolIDBytes		= 64
	MaxServiceIDBytes		= 96
	MaxZoneIDBytes			= 64
	MaxNonceBytes			= 64
	MaxChannelPolicies		= 32
	MaxStreamsPerSession		= 32
	SessionEphemeralKeyBytes	= 32
	MaxEncryptionContextBytes	= 128
	MaxQoSClassPolicies		= 16
	MaxPayloadChunks		= 65_536
	MaxChunkBytes			= 4 << 20
	MaxStreamMessageBytes		= 64 << 20
	DefaultMaxMessageBytes		= 1 << 20
	DefaultFlowWindowBytes		= 4 << 20
	DefaultHandshakeVersion		= uint32(1)
	DefaultProtocolVersion		= "aetra-networking-v1"
	DefaultCipherSuite		= CipherSuiteEd25519X25519ChaCha20Poly1305
	DefaultCompressionMode		= CompressionModeNone
	DefaultQoSPolicy		= QoSPolicyBalanced
)

type NodeRole string

const (
	NodeRoleValidator	NodeRole	= "VALIDATOR"
	NodeRoleFull		NodeRole	= "FULL_NODE"
	NodeRoleArchive		NodeRole	= "ARCHIVE_NODE"
	NodeRoleStateSync	NodeRole	= "STATE_SYNC_NODE"
	NodeRoleZoneExecution	NodeRole	= "ZONE_EXECUTION_NODE"
	NodeRoleService		NodeRole	= "SERVICE_NODE"
	NodeRoleStorageProvider	NodeRole	= "STORAGE_PROVIDER_NODE"
	NodeRoleRouting		NodeRole	= "ROUTING_NODE"
	NodeRoleIndex		NodeRole	= "INDEX_NODE"
	NodeRoleLightGateway	NodeRole	= "LIGHT_CLIENT_GATEWAY"
)

type ChannelClass string

const (
	ChannelConsensus	ChannelClass	= "CONSENSUS_CHANNEL"
	ChannelMempool		ChannelClass	= "MEMPOOL_CHANNEL"
	ChannelBlock		ChannelClass	= "BLOCK_CHANNEL"
	ChannelStateSync	ChannelClass	= "STATE_SYNC_CHANNEL"
	ChannelData		ChannelClass	= "DATA_CHANNEL"
	ChannelExecution	ChannelClass	= "EXECUTION_CHANNEL"
	ChannelService		ChannelClass	= "SERVICE_CHANNEL"
	ChannelRouting		ChannelClass	= "ROUTING_CHANNEL"
	ChannelDiscovery	ChannelClass	= "DISCOVERY_CHANNEL"
)

type CipherSuite string

const (
	CipherSuiteEd25519X25519ChaCha20Poly1305 CipherSuite = "ED25519_X25519_CHACHA20_POLY1305"
)

type CompressionMode string

const (
	CompressionModeNone	CompressionMode	= "NONE"
	CompressionModeZstd	CompressionMode	= "ZSTD"
)

type QoSPolicy string

const (
	QoSPolicyConsensusFirst	QoSPolicy	= "CONSENSUS_FIRST"
	QoSPolicyBalanced	QoSPolicy	= "BALANCED"
	QoSPolicyBulk		QoSPolicy	= "BULK"
)

type QoSClass string

const (
	QoSClassCriticalConsensus	QoSClass	= "critical_consensus"
	QoSClassBlockPropagation	QoSClass	= "block_propagation"
	QoSClassStateSync		QoSClass	= "state_sync"
	QoSClassExecutionMessage	QoSClass	= "execution_message"
	QoSClassServiceCall		QoSClass	= "service_call"
	QoSClassDiscovery		QoSClass	= "discovery"
	QoSClassBulkData		QoSClass	= "bulk_data"
)

type NodeRecord struct {
	NodeID			string
	NodePubKey		[]byte
	PublicKey		[]byte
	ValidatorPubKey		[]byte
	OperatorAddress		string
	Roles			[]NodeRole
	NetworkAddressesHash	string
	ZonesSupported		[]string
	Services		[]string
	ServicesSupported	[]string
	ServiceIDs		[]string
	ProtocolVersions	[]string
	SupportedProtocols	[]string
	Reputation		ReputationCommitment
	LatencyVector		[]NodeLatencyVectorEntry
	RecordVersion		uint64
	ExpiresHeight		uint64
	Signature		[]byte
}

type ChannelPolicy struct {
	Channel		ChannelClass
	Priority	uint32
	MaxMessageBytes	uint64
	BandwidthWeight	uint32
	BurstBytes	uint64
}

type StreamSpec struct {
	StreamID		string
	Channel			ChannelClass
	Priority		uint32
	FlowControlWindow	uint64
	MaxMessageBytes		uint64
	Compression		CompressionMode
	EncryptionContext	string
}

type SessionRequest struct {
	LocalNodeID			string
	RemoteNodeID			string
	HandshakeVersion		uint32
	CipherSuites			[]CipherSuite
	ProtocolVersions		[]string
	ChannelClasses			[]ChannelClass
	LocalEphemeralPubKey		[]byte
	RemoteEphemeralPubKey		[]byte
	SessionSecretCommitmentHash	string
	OpenedHeight			uint64
	ExpiresHeight			uint64
	Nonce				[]byte
	QOSPolicy			QoSPolicy
}

type SessionChannel struct {
	LocalNodeID		string
	RemoteNodeID		string
	SessionID		string
	HandshakeVersion	uint32
	CipherSuite		CipherSuite
	ProtocolVersions	[]string
	OpenedHeight		uint64
	ExpiresHeight		uint64
	SessionKeys		SessionKeySet
	Streams			[]StreamSpec
	QOSPolicy		QoSPolicy
}

type SessionKeySet struct {
	KeyID			string
	CipherSuite		CipherSuite
	LocalEphemeralPubKey	[]byte
	RemoteEphemeralPubKey	[]byte
	TranscriptHash		string
	SecretCommitmentHash	string
	EstablishedHeight	uint64
	ExpiresHeight		uint64
}

type TransportEnvelope struct {
	ID		string
	Channel		ChannelClass
	SizeBytes	uint64
	EnqueuedHeight	uint64
	Sequence	uint64
	PayloadHash	string
}

type PayloadChunk struct {
	PayloadID	string
	PayloadHash	string
	Index		uint32
	Total		uint32
	ChunkHash	string
	Bytes		[]byte
}

type PeerMetrics struct {
	LatencyMillis		uint64
	ReliabilityBps		uint32
	ThroughputBytesPerSec	uint64
	InvalidMessageCount	uint64
}

type PeerScore struct {
	ScoreBps	uint32
	LatencyBps	uint32
	ReliabilityBps	uint32
	ThroughputBps	uint32
	PenaltyBps	uint32
}

type QoSClassPolicy struct {
	Class			QoSClass
	Channel			ChannelClass
	Priority		uint32
	BandwidthFloorBps	uint32
	BandwidthCeilBps	uint32
	ReservedCapacity	bool
	Backpressure		bool
}

type StreamResetPolicy string

const (
	StreamResetKeepSession	StreamResetPolicy	= "KEEP_SESSION"
	StreamResetCloseSession	StreamResetPolicy	= "CLOSE_SESSION"
)

type StreamResetDecision struct {
	StreamID		string
	SessionClosed		bool
	RemainingStreams	[]StreamSpec
}

type PeerQoSDecision struct {
	Class			QoSClass
	DowngradeServiceTraffic	bool
	DisconnectConsensus	bool
	Reason			string
}

func DefaultChannelPolicies() []ChannelPolicy {
	return []ChannelPolicy{
		{Channel: ChannelConsensus, Priority: 0, MaxMessageBytes: DefaultMaxMessageBytes, BandwidthWeight: 3_000, BurstBytes: 2 << 20},
		{Channel: ChannelBlock, Priority: 1, MaxMessageBytes: 4 << 20, BandwidthWeight: 2_000, BurstBytes: 8 << 20},
		{Channel: ChannelStateSync, Priority: 2, MaxMessageBytes: 4 << 20, BandwidthWeight: 1_250, BurstBytes: 16 << 20},
		{Channel: ChannelExecution, Priority: 3, MaxMessageBytes: DefaultMaxMessageBytes, BandwidthWeight: 1_000, BurstBytes: 4 << 20},
		{Channel: ChannelMempool, Priority: 4, MaxMessageBytes: DefaultMaxMessageBytes, BandwidthWeight: 1_000, BurstBytes: 4 << 20},
		{Channel: ChannelService, Priority: 5, MaxMessageBytes: DefaultMaxMessageBytes, BandwidthWeight: 350, BurstBytes: 2 << 20},
		{Channel: ChannelRouting, Priority: 5, MaxMessageBytes: DefaultMaxMessageBytes, BandwidthWeight: 750, BurstBytes: 2 << 20},
		{Channel: ChannelDiscovery, Priority: 5, MaxMessageBytes: 256 << 10, BandwidthWeight: 500, BurstBytes: 1 << 20},
		{Channel: ChannelData, Priority: 6, MaxMessageBytes: MaxStreamMessageBytes, BandwidthWeight: 150, BurstBytes: MaxStreamMessageBytes},
	}
}

func DefaultQoSClassPolicies() []QoSClassPolicy {
	return []QoSClassPolicy{
		{Class: QoSClassCriticalConsensus, Channel: ChannelConsensus, Priority: 0, BandwidthFloorBps: 3_000, BandwidthCeilBps: BasisPoints, ReservedCapacity: true},
		{Class: QoSClassBlockPropagation, Channel: ChannelBlock, Priority: 1, BandwidthFloorBps: 1_500, BandwidthCeilBps: 8_000, ReservedCapacity: true},
		{Class: QoSClassStateSync, Channel: ChannelStateSync, Priority: 2, BandwidthFloorBps: 1_000, BandwidthCeilBps: 7_000, Backpressure: true},
		{Class: QoSClassExecutionMessage, Channel: ChannelExecution, Priority: 3, BandwidthFloorBps: 750, BandwidthCeilBps: 6_000},
		{Class: QoSClassServiceCall, Channel: ChannelService, Priority: 5, BandwidthFloorBps: 250, BandwidthCeilBps: 4_000},
		{Class: QoSClassDiscovery, Channel: ChannelDiscovery, Priority: 5, BandwidthFloorBps: 100, BandwidthCeilBps: 2_000},
		{Class: QoSClassBulkData, Channel: ChannelData, Priority: 6, BandwidthFloorBps: 0, BandwidthCeilBps: 3_000, Backpressure: true},
	}
}

func PriorityForChannel(channel ChannelClass) uint32 {
	for _, policy := range DefaultChannelPolicies() {
		if policy.Channel == channel {
			return policy.Priority
		}
	}
	return 100
}

func NormalizeNodeRecord(record NodeRecord) NodeRecord {
	record.NodeID = strings.ToLower(strings.TrimSpace(record.NodeID))
	record.NodePubKey = cloneBytes(record.NodePubKey)
	record.PublicKey = cloneBytes(record.PublicKey)
	if len(record.NodePubKey) == 0 && len(record.PublicKey) > 0 {
		record.NodePubKey = cloneBytes(record.PublicKey)
	}
	if len(record.PublicKey) == 0 && len(record.NodePubKey) > 0 {
		record.PublicKey = cloneBytes(record.NodePubKey)
	}
	record.ValidatorPubKey = cloneBytes(record.ValidatorPubKey)
	record.OperatorAddress = strings.TrimSpace(record.OperatorAddress)
	record.NetworkAddressesHash = strings.ToLower(strings.TrimSpace(record.NetworkAddressesHash))
	record.Roles = normalizeRoles(record.Roles)
	record.ZonesSupported, _ = normalizeStringSet("zone", record.ZonesSupported, MaxZoneIDBytes)
	services := normalizePreferredStringSet(MaxServiceIDBytes, record.ServicesSupported, record.Services, record.ServiceIDs)
	record.Services = append([]string(nil), services...)
	record.ServicesSupported = append([]string(nil), services...)
	record.ServiceIDs = append([]string(nil), services...)
	protocols := normalizePreferredStringSet(MaxProtocolIDBytes, record.ProtocolVersions, record.SupportedProtocols)
	record.ProtocolVersions = append([]string(nil), protocols...)
	record.SupportedProtocols = append([]string(nil), protocols...)
	record.Reputation = NormalizeReputationCommitment(record.Reputation)
	record.LatencyVector = normalizeNodeLatencyVector(record.LatencyVector)
	if record.RecordVersion == 0 {
		record.RecordVersion = DefaultNodeRecordVersion
	}
	record.Signature = cloneBytes(record.Signature)
	return record
}

func (r NodeRecord) SigningPayload() ([]byte, error) {
	normalized := NormalizeNodeRecord(r)
	normalized.Signature = nil
	bz, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func SignNodeRecord(record NodeRecord, privateKey ed25519.PrivateKey, networkSalt []byte) (NodeRecord, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return NodeRecord{}, errors.New("networking node private key must be ed25519")
	}
	if len(networkSalt) == 0 {
		return NodeRecord{}, errors.New("networking network salt is required")
	}
	pubKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return NodeRecord{}, errors.New("networking node public key must be ed25519")
	}
	record.NodePubKey = cloneBytes(pubKey)
	identityPubKey := record.NodePubKey
	if len(record.ValidatorPubKey) > 0 {
		identityPubKey = record.ValidatorPubKey
	}
	record.NodeID = ComputeNodeID(identityPubKey, networkSalt)
	normalized := NormalizeNodeRecord(record)
	payload, err := normalized.SigningPayload()
	if err != nil {
		return NodeRecord{}, err
	}
	normalized.Signature = ed25519.Sign(privateKey, payload)
	if err := normalized.Validate(networkSalt, 0); err != nil {
		return NodeRecord{}, err
	}
	return normalized, nil
}

func (r NodeRecord) ValidateBasic() error {
	record := NormalizeNodeRecord(r)
	if err := ValidateHash("networking node id", record.NodeID); err != nil {
		return err
	}
	if len(record.NodePubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking node pub key must be %d bytes", ed25519.PublicKeySize)
	}
	if len(record.PublicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking public key must be %d bytes", ed25519.PublicKeySize)
	}
	if string(record.PublicKey) != string(record.NodePubKey) {
		return errors.New("networking public key must match node pub key")
	}
	if len(record.ValidatorPubKey) > 0 && len(record.ValidatorPubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking validator pub key must be %d bytes", ed25519.PublicKeySize)
	}
	if len(record.Roles) == 0 {
		return errors.New("networking node record requires at least one role")
	}
	if len(record.Roles) > MaxRolesPerNode {
		return fmt.Errorf("networking node record roles must be <= %d", MaxRolesPerNode)
	}
	if hasRole(record.Roles, NodeRoleValidator) && len(record.ValidatorPubKey) == 0 {
		return errors.New("networking validator role requires validator pub key")
	}
	if record.NetworkAddressesHash == "" {
		return errors.New("networking network addresses hash is required")
	}
	if err := ValidateHash("networking network addresses hash", record.NetworkAddressesHash); err != nil {
		return err
	}
	if len(record.ZonesSupported) > MaxZonesPerNode {
		return fmt.Errorf("networking zones supported must be <= %d", MaxZonesPerNode)
	}
	if len(record.ServicesSupported) > MaxServicesPerNode {
		return fmt.Errorf("networking services supported must be <= %d", MaxServicesPerNode)
	}
	if len(record.Services) > MaxServicesPerNode || len(record.ServiceIDs) > MaxServicesPerNode {
		return fmt.Errorf("networking service ids must be <= %d", MaxServicesPerNode)
	}
	if len(record.ProtocolVersions) == 0 {
		return errors.New("networking node record requires at least one protocol")
	}
	if len(record.ProtocolVersions) > MaxProtocolsPerNode {
		return fmt.Errorf("networking protocol versions must be <= %d", MaxProtocolsPerNode)
	}
	if len(record.SupportedProtocols) > MaxProtocolsPerNode {
		return fmt.Errorf("networking supported protocols must be <= %d", MaxProtocolsPerNode)
	}
	if record.RecordVersion == 0 {
		return errors.New("networking node record version must be positive")
	}
	if record.ExpiresHeight == 0 {
		return errors.New("networking node record expires height must be positive")
	}
	if len(record.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("networking node record signature must be %d bytes", ed25519.SignatureSize)
	}
	for _, role := range record.Roles {
		if !IsNodeRole(role) {
			return fmt.Errorf("unknown networking node role %q", role)
		}
	}
	if err := validateIdentifierSet("zone", record.ZonesSupported, MaxZoneIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("service", record.ServicesSupported, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("service", record.Services, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("service", record.ServiceIDs, MaxServiceIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("protocol", record.ProtocolVersions, MaxProtocolIDBytes); err != nil {
		return err
	}
	if err := validateIdentifierSet("protocol", record.SupportedProtocols, MaxProtocolIDBytes); err != nil {
		return err
	}
	if IsReputationCommitmentSet(record.Reputation) {
		if err := record.Reputation.Validate(); err != nil {
			return err
		}
		if record.Reputation.NodeID != record.NodeID {
			return errors.New("networking node reputation must reference node")
		}
	}
	return validateNodeLatencyVector(record.LatencyVector, record.NodeID)
}

func (r NodeRecord) Validate(networkSalt []byte, currentHeight uint64) error {
	record := NormalizeNodeRecord(r)
	if err := record.ValidateBasic(); err != nil {
		return err
	}
	if len(networkSalt) == 0 {
		return errors.New("networking network salt is required")
	}
	if currentHeight > 0 && currentHeight > record.ExpiresHeight {
		return errors.New("networking node record is expired")
	}
	identityPubKey := record.NodePubKey
	if len(record.ValidatorPubKey) > 0 {
		identityPubKey = record.ValidatorPubKey
	}
	expectedNodeID := ComputeNodeID(identityPubKey, networkSalt)
	if record.NodeID != expectedNodeID {
		return errors.New("networking node id does not match identity key")
	}
	payload, err := record.SigningPayload()
	if err != nil {
		return err
	}
	if !ed25519.Verify(record.NodePubKey, payload, record.Signature) {
		return errors.New("networking node record signature verification failed")
	}
	return nil
}

func VerifyNodeRecordAddresses(record NodeRecord, addresses []string) error {
	record = NormalizeNodeRecord(record)
	if err := record.ValidateBasic(); err != nil {
		return err
	}
	addressHash, err := HashNetworkAddresses(addresses)
	if err != nil {
		return err
	}
	if addressHash != record.NetworkAddressesHash {
		return errors.New("networking network address list does not match signed node record hash")
	}
	return nil
}

func ValidateChannelPolicies(policies []ChannelPolicy) error {
	if len(policies) == 0 {
		return errors.New("networking channel policies are required")
	}
	if len(policies) > MaxChannelPolicies {
		return fmt.Errorf("networking channel policies must be <= %d", MaxChannelPolicies)
	}
	seen := make(map[ChannelClass]struct{}, len(policies))
	for _, policy := range policies {
		if err := policy.Validate(); err != nil {
			return err
		}
		if _, found := seen[policy.Channel]; found {
			return errors.New("networking duplicate channel policy")
		}
		seen[policy.Channel] = struct{}{}
	}
	if priorityForPolicy(policies, ChannelConsensus) >= priorityForPolicy(policies, ChannelService) {
		return errors.New("networking consensus channel must outrank service channel")
	}
	if priorityForPolicy(policies, ChannelConsensus) >= priorityForPolicy(policies, ChannelData) {
		return errors.New("networking consensus channel must outrank data channel")
	}
	return nil
}

func (p ChannelPolicy) Validate() error {
	if !IsChannelClass(p.Channel) {
		return fmt.Errorf("unknown networking channel %q", p.Channel)
	}
	if p.MaxMessageBytes == 0 || p.MaxMessageBytes > MaxStreamMessageBytes {
		return fmt.Errorf("networking channel max message bytes must be between 1 and %d", MaxStreamMessageBytes)
	}
	if p.BandwidthWeight > BasisPoints {
		return fmt.Errorf("networking channel bandwidth weight must be <= %d bps", BasisPoints)
	}
	if p.BurstBytes < p.MaxMessageBytes {
		return errors.New("networking channel burst bytes must cover max message bytes")
	}
	return nil
}

func (s StreamSpec) Validate() error {
	if strings.TrimSpace(s.StreamID) != s.StreamID || s.StreamID == "" {
		return errors.New("networking stream id is required and must not have surrounding whitespace")
	}
	if len(s.StreamID) > MaxProtocolIDBytes {
		return fmt.Errorf("networking stream id must be <= %d bytes", MaxProtocolIDBytes)
	}
	if !IsChannelClass(s.Channel) {
		return fmt.Errorf("unknown networking stream channel %q", s.Channel)
	}
	if s.FlowControlWindow == 0 || s.FlowControlWindow > MaxStreamMessageBytes*2 {
		return errors.New("networking stream flow control window out of bounds")
	}
	if s.MaxMessageBytes == 0 || s.MaxMessageBytes > MaxStreamMessageBytes {
		return fmt.Errorf("networking stream max message bytes must be between 1 and %d", MaxStreamMessageBytes)
	}
	if s.MaxMessageBytes > s.FlowControlWindow {
		return errors.New("networking stream max message bytes must fit flow control window")
	}
	if !IsCompressionMode(s.Compression) {
		return fmt.Errorf("unknown networking compression mode %q", s.Compression)
	}
	if strings.TrimSpace(s.EncryptionContext) != s.EncryptionContext || s.EncryptionContext == "" {
		return errors.New("networking stream encryption context is required and must not have surrounding whitespace")
	}
	if len(s.EncryptionContext) > MaxEncryptionContextBytes {
		return fmt.Errorf("networking stream encryption context must be <= %d bytes", MaxEncryptionContextBytes)
	}
	if s.Channel == ChannelConsensus && s.FlowControlWindow < DefaultFlowWindowBytes {
		return errors.New("networking consensus stream requires reserved flow control capacity")
	}
	if s.Channel == ChannelData && s.FlowControlWindow < s.MaxMessageBytes {
		return errors.New("networking bulk data stream requires backpressure-capable flow window")
	}
	return nil
}

func (req SessionRequest) Normalize() SessionRequest {
	req.LocalNodeID = strings.ToLower(strings.TrimSpace(req.LocalNodeID))
	req.RemoteNodeID = strings.ToLower(strings.TrimSpace(req.RemoteNodeID))
	if req.HandshakeVersion == 0 {
		req.HandshakeVersion = DefaultHandshakeVersion
	}
	if len(req.CipherSuites) == 0 {
		req.CipherSuites = []CipherSuite{DefaultCipherSuite}
	}
	req.ProtocolVersions, _ = normalizeStringSet("protocol", req.ProtocolVersions, MaxProtocolIDBytes)
	req.ChannelClasses = normalizeChannels(req.ChannelClasses)
	req.LocalEphemeralPubKey = cloneBytes(req.LocalEphemeralPubKey)
	req.RemoteEphemeralPubKey = cloneBytes(req.RemoteEphemeralPubKey)
	req.SessionSecretCommitmentHash = strings.ToLower(strings.TrimSpace(req.SessionSecretCommitmentHash))
	req.Nonce = cloneBytes(req.Nonce)
	if req.QOSPolicy == "" {
		req.QOSPolicy = DefaultQoSPolicy
	}
	return req
}

func NegotiateSession(local, remote NodeRecord, req SessionRequest) (SessionChannel, error) {
	local = NormalizeNodeRecord(local)
	remote = NormalizeNodeRecord(remote)
	req = req.Normalize()
	if err := local.ValidateBasic(); err != nil {
		return SessionChannel{}, err
	}
	if err := remote.ValidateBasic(); err != nil {
		return SessionChannel{}, err
	}
	if req.LocalNodeID != local.NodeID {
		return SessionChannel{}, errors.New("networking session local node mismatch")
	}
	if req.RemoteNodeID != remote.NodeID {
		return SessionChannel{}, errors.New("networking session remote node mismatch")
	}
	if req.HandshakeVersion != DefaultHandshakeVersion {
		return SessionChannel{}, errors.New("networking unsupported handshake version")
	}
	if req.OpenedHeight == 0 {
		return SessionChannel{}, errors.New("networking session opened height must be positive")
	}
	if req.ExpiresHeight <= req.OpenedHeight {
		return SessionChannel{}, errors.New("networking session expires height must exceed opened height")
	}
	if req.ExpiresHeight > local.ExpiresHeight || req.ExpiresHeight > remote.ExpiresHeight {
		return SessionChannel{}, errors.New("networking session cannot outlive node records")
	}
	if len(req.Nonce) == 0 || len(req.Nonce) > MaxNonceBytes {
		return SessionChannel{}, fmt.Errorf("networking session nonce must be between 1 and %d bytes", MaxNonceBytes)
	}
	if !IsQoSPolicy(req.QOSPolicy) {
		return SessionChannel{}, fmt.Errorf("unknown networking qos policy %q", req.QOSPolicy)
	}
	cipher, err := chooseCipher(req.CipherSuites)
	if err != nil {
		return SessionChannel{}, err
	}
	protocols := intersectStrings(local.ProtocolVersions, remote.ProtocolVersions, req.ProtocolVersions)
	if len(protocols) == 0 {
		return SessionChannel{}, errors.New("networking session has no mutually supported protocol")
	}
	channels := req.ChannelClasses
	if len(channels) == 0 {
		channels = normalizeChannels([]ChannelClass{ChannelConsensus, ChannelBlock, ChannelStateSync, ChannelExecution, ChannelMempool, ChannelService, ChannelRouting, ChannelDiscovery, ChannelData})
	}
	if err := validateChannels(channels); err != nil {
		return SessionChannel{}, err
	}
	sessionKeys, err := BuildSessionKeySet(req, cipher, protocols, channels)
	if err != nil {
		return SessionChannel{}, err
	}
	session := SessionChannel{
		LocalNodeID:		req.LocalNodeID,
		RemoteNodeID:		req.RemoteNodeID,
		HandshakeVersion:	req.HandshakeVersion,
		CipherSuite:		cipher,
		ProtocolVersions:	protocols,
		OpenedHeight:		req.OpenedHeight,
		ExpiresHeight:		req.ExpiresHeight,
		SessionKeys:		sessionKeys,
		Streams:		defaultStreams(channels, sessionKeys.KeyID),
		QOSPolicy:		req.QOSPolicy,
	}
	session.SessionID = ComputeSessionID(req, cipher, protocols, channels)
	if err := session.Validate(); err != nil {
		return SessionChannel{}, err
	}
	return session, nil
}

func (s SessionChannel) Validate() error {
	s.LocalNodeID = strings.ToLower(strings.TrimSpace(s.LocalNodeID))
	s.RemoteNodeID = strings.ToLower(strings.TrimSpace(s.RemoteNodeID))
	if err := ValidateHash("networking session local node id", s.LocalNodeID); err != nil {
		return err
	}
	if err := ValidateHash("networking session remote node id", s.RemoteNodeID); err != nil {
		return err
	}
	if s.LocalNodeID == s.RemoteNodeID {
		return errors.New("networking session endpoints must differ")
	}
	if err := ValidateHash("networking session id", strings.ToLower(strings.TrimSpace(s.SessionID))); err != nil {
		return err
	}
	if s.HandshakeVersion != DefaultHandshakeVersion {
		return errors.New("networking unsupported session handshake version")
	}
	if !IsCipherSuite(s.CipherSuite) {
		return fmt.Errorf("unknown networking cipher suite %q", s.CipherSuite)
	}
	if s.OpenedHeight == 0 || s.ExpiresHeight <= s.OpenedHeight {
		return errors.New("networking session height range is invalid")
	}
	if err := s.SessionKeys.Validate(); err != nil {
		return err
	}
	if s.SessionKeys.CipherSuite != s.CipherSuite {
		return errors.New("networking session key cipher suite mismatch")
	}
	if s.SessionKeys.EstablishedHeight < s.OpenedHeight || s.SessionKeys.EstablishedHeight > s.ExpiresHeight || s.SessionKeys.ExpiresHeight > s.ExpiresHeight {
		return errors.New("networking session key height range mismatch")
	}
	if len(s.ProtocolVersions) == 0 {
		return errors.New("networking session requires protocol versions")
	}
	if err := validateIdentifierSet("protocol", s.ProtocolVersions, MaxProtocolIDBytes); err != nil {
		return err
	}
	if len(s.Streams) == 0 || len(s.Streams) > MaxStreamsPerSession {
		return fmt.Errorf("networking session streams must be between 1 and %d", MaxStreamsPerSession)
	}
	seen := make(map[string]struct{}, len(s.Streams))
	for _, stream := range s.Streams {
		if err := stream.Validate(); err != nil {
			return err
		}
		if stream.EncryptionContext != streamEncryptionContext(s.SessionKeys.KeyID, stream.StreamID) {
			return errors.New("networking stream encryption context does not match session key")
		}
		if _, found := seen[stream.StreamID]; found {
			return errors.New("networking duplicate stream id")
		}
		seen[stream.StreamID] = struct{}{}
	}
	if err := ValidateStreamSet(s.Streams, DefaultQoSClassPolicies()); err != nil {
		return err
	}
	if !IsQoSPolicy(s.QOSPolicy) {
		return fmt.Errorf("unknown networking qos policy %q", s.QOSPolicy)
	}
	return nil
}

func BuildSessionKeySet(req SessionRequest, cipher CipherSuite, protocols []string, channels []ChannelClass) (SessionKeySet, error) {
	req = req.Normalize()
	if !IsCipherSuite(cipher) {
		return SessionKeySet{}, fmt.Errorf("unknown networking cipher suite %q", cipher)
	}
	if len(req.LocalEphemeralPubKey) != SessionEphemeralKeyBytes || len(req.RemoteEphemeralPubKey) != SessionEphemeralKeyBytes {
		return SessionKeySet{}, fmt.Errorf("networking session ephemeral public keys must be %d bytes", SessionEphemeralKeyBytes)
	}
	if err := ValidateHash("networking session secret commitment hash", req.SessionSecretCommitmentHash); err != nil {
		return SessionKeySet{}, err
	}
	transcriptHash := ComputeSessionTranscriptHash(req, cipher, protocols, channels)
	keySet := SessionKeySet{
		KeyID:			HashParts("session-key", transcriptHash, req.SessionSecretCommitmentHash),
		CipherSuite:		cipher,
		LocalEphemeralPubKey:	cloneBytes(req.LocalEphemeralPubKey),
		RemoteEphemeralPubKey:	cloneBytes(req.RemoteEphemeralPubKey),
		TranscriptHash:		transcriptHash,
		SecretCommitmentHash:	req.SessionSecretCommitmentHash,
		EstablishedHeight:	req.OpenedHeight,
		ExpiresHeight:		req.ExpiresHeight,
	}
	return keySet, keySet.Validate()
}

func (s SessionKeySet) Validate() error {
	keyID := strings.ToLower(strings.TrimSpace(s.KeyID))
	transcriptHash := strings.ToLower(strings.TrimSpace(s.TranscriptHash))
	secretCommitmentHash := strings.ToLower(strings.TrimSpace(s.SecretCommitmentHash))
	if err := ValidateHash("networking session key id", keyID); err != nil {
		return err
	}
	if !IsCipherSuite(s.CipherSuite) {
		return fmt.Errorf("unknown networking session key cipher suite %q", s.CipherSuite)
	}
	if len(s.LocalEphemeralPubKey) != SessionEphemeralKeyBytes || len(s.RemoteEphemeralPubKey) != SessionEphemeralKeyBytes {
		return fmt.Errorf("networking session ephemeral public keys must be %d bytes", SessionEphemeralKeyBytes)
	}
	if bytes.Equal(s.LocalEphemeralPubKey, s.RemoteEphemeralPubKey) {
		return errors.New("networking session ephemeral public keys must differ")
	}
	if err := ValidateHash("networking session transcript hash", transcriptHash); err != nil {
		return err
	}
	if err := ValidateHash("networking session secret commitment hash", secretCommitmentHash); err != nil {
		return err
	}
	if HashParts("session-key", transcriptHash, secretCommitmentHash) != keyID {
		return errors.New("networking session key id mismatch")
	}
	if s.EstablishedHeight == 0 || s.ExpiresHeight <= s.EstablishedHeight {
		return errors.New("networking session key height range is invalid")
	}
	return nil
}

func NegotiateVerifiedSession(local, remote NodeRecord, req SessionRequest, networkSalt []byte, currentHeight uint64) (SessionChannel, error) {
	local = NormalizeNodeRecord(local)
	remote = NormalizeNodeRecord(remote)
	if err := local.Validate(networkSalt, currentHeight); err != nil {
		return SessionChannel{}, err
	}
	if err := remote.Validate(networkSalt, currentHeight); err != nil {
		return SessionChannel{}, err
	}
	return NegotiateSession(local, remote, req)
}

func ValidateQoSClassPolicies(policies []QoSClassPolicy) error {
	if len(policies) == 0 || len(policies) > MaxQoSClassPolicies {
		return fmt.Errorf("networking qos class policies must be between 1 and %d", MaxQoSClassPolicies)
	}
	seen := make(map[QoSClass]struct{}, len(policies))
	for _, policy := range policies {
		if err := policy.Validate(); err != nil {
			return err
		}
		if _, found := seen[policy.Class]; found {
			return errors.New("networking duplicate qos class policy")
		}
		seen[policy.Class] = struct{}{}
	}
	if priorityForQoSClass(policies, QoSClassCriticalConsensus) >= priorityForQoSClass(policies, QoSClassServiceCall) {
		return errors.New("networking qos priority inversion for consensus and service traffic")
	}
	if priorityForQoSClass(policies, QoSClassCriticalConsensus) >= priorityForQoSClass(policies, QoSClassBulkData) {
		return errors.New("networking qos priority inversion for consensus and bulk data")
	}
	consensus, found := findQoSClassPolicy(policies, QoSClassCriticalConsensus)
	if !found || !consensus.ReservedCapacity || consensus.BandwidthFloorBps == 0 {
		return errors.New("networking consensus qos class requires reserved bandwidth floor")
	}
	bulk, found := findQoSClassPolicy(policies, QoSClassBulkData)
	if !found || !bulk.Backpressure {
		return errors.New("networking bulk data qos class requires backpressure")
	}
	return nil
}

func (p QoSClassPolicy) Validate() error {
	if !IsQoSClass(p.Class) {
		return fmt.Errorf("unknown networking qos class %q", p.Class)
	}
	if !IsChannelClass(p.Channel) {
		return fmt.Errorf("unknown networking qos channel %q", p.Channel)
	}
	if p.BandwidthFloorBps > BasisPoints || p.BandwidthCeilBps > BasisPoints {
		return fmt.Errorf("networking qos bandwidth floor and ceiling must be <= %d bps", BasisPoints)
	}
	if p.BandwidthCeilBps == 0 || p.BandwidthFloorBps > p.BandwidthCeilBps {
		return errors.New("networking qos bandwidth floor must be <= positive ceiling")
	}
	if QoSClassForChannel(p.Channel) != p.Class {
		return errors.New("networking qos class does not match channel")
	}
	return nil
}

func ValidateStreamSet(streams []StreamSpec, policies []QoSClassPolicy) error {
	if err := ValidateQoSClassPolicies(policies); err != nil {
		return err
	}
	byChannel := make(map[ChannelClass]StreamSpec, len(streams))
	for _, stream := range streams {
		byChannel[stream.Channel] = stream
	}
	consensus, found := byChannel[ChannelConsensus]
	if !found {
		return errors.New("networking stream set requires consensus stream")
	}
	if consensus.Priority != priorityForQoSClass(policies, QoSClassCriticalConsensus) {
		return errors.New("networking consensus stream priority must match qos class")
	}
	for _, channel := range []ChannelClass{ChannelService, ChannelData} {
		stream, found := byChannel[channel]
		if !found {
			continue
		}
		if stream.Priority <= consensus.Priority {
			return errors.New("networking service or bulk stream cannot outrank consensus stream")
		}
	}
	if data, found := byChannel[ChannelData]; found && data.FlowControlWindow < data.MaxMessageBytes {
		return errors.New("networking bulk data stream must support backpressure")
	}
	return nil
}

func ResetStream(session SessionChannel, streamID string, policy StreamResetPolicy) (StreamResetDecision, error) {
	if err := session.Validate(); err != nil {
		return StreamResetDecision{}, err
	}
	streamID = strings.TrimSpace(streamID)
	if streamID == "" {
		return StreamResetDecision{}, errors.New("networking stream reset id is required")
	}
	switch policy {
	case StreamResetKeepSession, StreamResetCloseSession:
	default:
		return StreamResetDecision{}, fmt.Errorf("unknown networking stream reset policy %q", policy)
	}
	if policy == StreamResetCloseSession {
		return StreamResetDecision{StreamID: streamID, SessionClosed: true}, nil
	}
	remaining := make([]StreamSpec, 0, len(session.Streams))
	found := false
	for _, stream := range session.Streams {
		if stream.StreamID == streamID {
			found = true
			continue
		}
		remaining = append(remaining, stream)
	}
	if !found {
		return StreamResetDecision{}, errors.New("networking stream reset target not found")
	}
	return StreamResetDecision{StreamID: streamID, RemainingStreams: remaining}, nil
}

func EvaluatePeerServiceQuota(usedBytes, quotaBytes uint64) PeerQoSDecision {
	if quotaBytes == 0 || usedBytes <= quotaBytes {
		return PeerQoSDecision{Class: QoSClassServiceCall}
	}
	return PeerQoSDecision{
		Class:				QoSClassServiceCall,
		DowngradeServiceTraffic:	true,
		DisconnectConsensus:		false,
		Reason:				"service quota exceeded",
	}
}

func (e TransportEnvelope) Normalize() TransportEnvelope {
	e.ID = strings.ToLower(strings.TrimSpace(e.ID))
	e.PayloadHash = strings.ToLower(strings.TrimSpace(e.PayloadHash))
	if e.ID == "" {
		e.ID = ComputeTransportEnvelopeID(e)
	}
	return e
}

func (e TransportEnvelope) Validate(policies []ChannelPolicy) error {
	envelope := e.Normalize()
	if err := ValidateHash("networking transport envelope id", envelope.ID); err != nil {
		return err
	}
	if !IsChannelClass(envelope.Channel) {
		return fmt.Errorf("unknown networking transport channel %q", envelope.Channel)
	}
	if envelope.SizeBytes == 0 {
		return errors.New("networking transport envelope size must be positive")
	}
	if envelope.EnqueuedHeight == 0 {
		return errors.New("networking transport envelope height must be positive")
	}
	if err := ValidateHash("networking transport payload hash", envelope.PayloadHash); err != nil {
		return err
	}
	policy, found := findPolicy(policies, envelope.Channel)
	if !found {
		return errors.New("networking missing channel policy")
	}
	if envelope.SizeBytes > policy.MaxMessageBytes {
		return errors.New("networking transport envelope exceeds channel limit")
	}
	return nil
}

func SortTransportEnvelopes(envelopes []TransportEnvelope, policies []ChannelPolicy) []TransportEnvelope {
	out := make([]TransportEnvelope, len(envelopes))
	for i, envelope := range envelopes {
		out[i] = envelope.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		leftPriority := priorityForPolicy(policies, out[i].Channel)
		rightPriority := priorityForPolicy(policies, out[j].Channel)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if out[i].EnqueuedHeight != out[j].EnqueuedHeight {
			return out[i].EnqueuedHeight < out[j].EnqueuedHeight
		}
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func SelectNextEnvelope(envelopes []TransportEnvelope, policies []ChannelPolicy) (TransportEnvelope, bool, error) {
	if len(envelopes) == 0 {
		return TransportEnvelope{}, false, nil
	}
	if err := ValidateChannelPolicies(policies); err != nil {
		return TransportEnvelope{}, false, err
	}
	for _, envelope := range envelopes {
		if err := envelope.Validate(policies); err != nil {
			return TransportEnvelope{}, false, err
		}
	}
	ordered := SortTransportEnvelopes(envelopes, policies)
	return ordered[0], true, nil
}

func ChunkPayload(payload []byte, maxChunkBytes uint64) ([]PayloadChunk, error) {
	if maxChunkBytes == 0 || maxChunkBytes > MaxChunkBytes {
		return nil, fmt.Errorf("networking max chunk bytes must be between 1 and %d", MaxChunkBytes)
	}
	if len(payload) == 0 {
		return nil, errors.New("networking payload is required")
	}
	total := (uint64(len(payload)) + maxChunkBytes - 1) / maxChunkBytes
	if total == 0 || total > MaxPayloadChunks {
		return nil, fmt.Errorf("networking payload chunks must be between 1 and %d", MaxPayloadChunks)
	}
	payloadHash := hashBytes("aetra-networking-payload-v1", payload)
	payloadID := HashParts("payload", payloadHash, fmt.Sprintf("%d", total))
	chunks := make([]PayloadChunk, 0, total)
	for offset, index := uint64(0), uint32(0); offset < uint64(len(payload)); index++ {
		end := offset + maxChunkBytes
		if end > uint64(len(payload)) {
			end = uint64(len(payload))
		}
		body := cloneBytes(payload[offset:end])
		chunk := PayloadChunk{
			PayloadID:	payloadID,
			PayloadHash:	payloadHash,
			Index:		index,
			Total:		uint32(total),
			Bytes:		body,
		}
		chunk.ChunkHash = ComputeChunkHash(chunk)
		chunks = append(chunks, chunk)
		offset = end
	}
	return chunks, nil
}

func ComputeChunkHash(chunk PayloadChunk) string {
	h := sha256Digest("aetra-networking-chunk-v1", chunk.PayloadID, chunk.PayloadHash, uint64(chunk.Index), uint64(chunk.Total), chunk.Bytes)
	return h
}

func ReassemblePayload(chunks []PayloadChunk) ([]byte, error) {
	if len(chunks) == 0 || len(chunks) > MaxPayloadChunks {
		return nil, fmt.Errorf("networking payload chunk count must be between 1 and %d", MaxPayloadChunks)
	}
	ordered := make([]PayloadChunk, len(chunks))
	copy(ordered, chunks)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Index < ordered[j].Index
	})
	first := ordered[0]
	if err := first.Validate(); err != nil {
		return nil, err
	}
	if int(first.Total) != len(ordered) {
		return nil, errors.New("networking payload chunk total mismatch")
	}
	payload := make([]byte, 0)
	for i, chunk := range ordered {
		if err := chunk.Validate(); err != nil {
			return nil, err
		}
		if chunk.PayloadID != first.PayloadID || chunk.PayloadHash != first.PayloadHash || chunk.Total != first.Total {
			return nil, errors.New("networking payload chunk set mismatch")
		}
		if int(chunk.Index) != i {
			return nil, errors.New("networking payload chunk sequence gap")
		}
		payload = append(payload, chunk.Bytes...)
	}
	if hashBytes("aetra-networking-payload-v1", payload) != first.PayloadHash {
		return nil, errors.New("networking payload hash mismatch")
	}
	return payload, nil
}

func (c PayloadChunk) Validate() error {
	if err := ValidateHash("networking payload id", strings.ToLower(strings.TrimSpace(c.PayloadID))); err != nil {
		return err
	}
	if err := ValidateHash("networking payload hash", strings.ToLower(strings.TrimSpace(c.PayloadHash))); err != nil {
		return err
	}
	if c.Total == 0 || c.Total > MaxPayloadChunks {
		return fmt.Errorf("networking payload chunk total must be between 1 and %d", MaxPayloadChunks)
	}
	if c.Index >= c.Total {
		return errors.New("networking payload chunk index out of range")
	}
	if len(c.Bytes) == 0 || len(c.Bytes) > MaxChunkBytes {
		return fmt.Errorf("networking payload chunk bytes must be between 1 and %d", MaxChunkBytes)
	}
	if err := ValidateHash("networking chunk hash", strings.ToLower(strings.TrimSpace(c.ChunkHash))); err != nil {
		return err
	}
	if ComputeChunkHash(c) != c.ChunkHash {
		return errors.New("networking chunk hash mismatch")
	}
	return nil
}

func ComputePeerScore(metrics PeerMetrics) (PeerScore, error) {
	if metrics.ReliabilityBps > BasisPoints {
		return PeerScore{}, fmt.Errorf("networking peer reliability must be <= %d bps", BasisPoints)
	}
	latency := latencyScore(metrics.LatencyMillis)
	throughput := throughputScore(metrics.ThroughputBytesPerSec)
	penalty := invalidMessagePenalty(metrics.InvalidMessageCount)
	raw := (uint64(latency)*2_000 + uint64(metrics.ReliabilityBps)*5_000 + uint64(throughput)*3_000) / uint64(BasisPoints)
	if raw > uint64(penalty) {
		raw -= uint64(penalty)
	} else {
		raw = 0
	}
	if raw > uint64(BasisPoints) {
		raw = uint64(BasisPoints)
	}
	return PeerScore{
		ScoreBps:	uint32(raw),
		LatencyBps:	latency,
		ReliabilityBps:	metrics.ReliabilityBps,
		ThroughputBps:	throughput,
		PenaltyBps:	penalty,
	}, nil
}

func ValidateHash(fieldName, value string) error {
	if len(value) != HashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	return nil
}

func IsNodeRole(role NodeRole) bool {
	switch role {
	case NodeRoleValidator, NodeRoleFull, NodeRoleArchive, NodeRoleStateSync, NodeRoleZoneExecution, NodeRoleService, NodeRoleStorageProvider, NodeRoleRouting, NodeRoleIndex, NodeRoleLightGateway:
		return true
	default:
		return false
	}
}

func IsChannelClass(channel ChannelClass) bool {
	switch channel {
	case ChannelConsensus, ChannelMempool, ChannelBlock, ChannelStateSync, ChannelData, ChannelExecution, ChannelService, ChannelRouting, ChannelDiscovery:
		return true
	default:
		return false
	}
}

func IsCipherSuite(cipher CipherSuite) bool {
	return cipher == CipherSuiteEd25519X25519ChaCha20Poly1305
}

func IsCompressionMode(mode CompressionMode) bool {
	switch mode {
	case CompressionModeNone, CompressionModeZstd:
		return true
	default:
		return false
	}
}

func IsQoSPolicy(policy QoSPolicy) bool {
	switch policy {
	case QoSPolicyConsensusFirst, QoSPolicyBalanced, QoSPolicyBulk:
		return true
	default:
		return false
	}
}

func IsQoSClass(class QoSClass) bool {
	switch class {
	case QoSClassCriticalConsensus, QoSClassBlockPropagation, QoSClassStateSync, QoSClassExecutionMessage, QoSClassServiceCall, QoSClassDiscovery, QoSClassBulkData:
		return true
	default:
		return false
	}
}

func QoSClassForChannel(channel ChannelClass) QoSClass {
	switch channel {
	case ChannelConsensus, ChannelMempool:
		return QoSClassCriticalConsensus
	case ChannelBlock:
		return QoSClassBlockPropagation
	case ChannelStateSync:
		return QoSClassStateSync
	case ChannelExecution:
		return QoSClassExecutionMessage
	case ChannelService, ChannelRouting:
		return QoSClassServiceCall
	case ChannelDiscovery:
		return QoSClassDiscovery
	case ChannelData:
		return QoSClassBulkData
	default:
		return ""
	}
}

func defaultStreams(channels []ChannelClass, sessionKeyID string) []StreamSpec {
	streams := make([]StreamSpec, len(channels))
	for i, channel := range channels {
		maxMessageBytes := uint64(DefaultMaxMessageBytes)
		if channel == ChannelData {
			maxMessageBytes = uint64(MaxStreamMessageBytes)
		}
		streamID := strings.ToLower(strings.TrimSuffix(string(channel), "_CHANNEL"))
		streams[i] = StreamSpec{
			StreamID:		streamID,
			Channel:		channel,
			Priority:		PriorityForChannel(channel),
			FlowControlWindow:	maxUint64(DefaultFlowWindowBytes, maxMessageBytes),
			MaxMessageBytes:	maxMessageBytes,
			Compression:		DefaultCompressionMode,
			EncryptionContext:	streamEncryptionContext(sessionKeyID, streamID),
		}
	}
	return streams
}

func streamEncryptionContext(sessionKeyID, streamID string) string {
	return HashParts("stream-encryption-context", sessionKeyID, streamID)[:MaxEncryptionContextBytes/2]
}

func chooseCipher(cipherSuites []CipherSuite) (CipherSuite, error) {
	for _, cipher := range cipherSuites {
		if IsCipherSuite(cipher) {
			return cipher, nil
		}
	}
	return "", errors.New("networking session has no supported cipher suite")
}

func findPolicy(policies []ChannelPolicy, channel ChannelClass) (ChannelPolicy, bool) {
	for _, policy := range policies {
		if policy.Channel == channel {
			return policy, true
		}
	}
	return ChannelPolicy{}, false
}

func priorityForPolicy(policies []ChannelPolicy, channel ChannelClass) uint32 {
	if policy, found := findPolicy(policies, channel); found {
		return policy.Priority
	}
	return PriorityForChannel(channel)
}

func findQoSClassPolicy(policies []QoSClassPolicy, class QoSClass) (QoSClassPolicy, bool) {
	for _, policy := range policies {
		if policy.Class == class {
			return policy, true
		}
	}
	return QoSClassPolicy{}, false
}

func priorityForQoSClass(policies []QoSClassPolicy, class QoSClass) uint32 {
	if policy, found := findQoSClassPolicy(policies, class); found {
		return policy.Priority
	}
	switch class {
	case QoSClassCriticalConsensus:
		return 0
	case QoSClassBlockPropagation:
		return 1
	case QoSClassStateSync:
		return 2
	case QoSClassExecutionMessage:
		return 3
	case QoSClassServiceCall, QoSClassDiscovery:
		return 5
	case QoSClassBulkData:
		return 6
	default:
		return 100
	}
}

func channelSortRank(channel ChannelClass) uint32 {
	id, err := ChannelIDForClass(channel)
	if err != nil {
		return uint32(MaxChannelPolicies) + uint32(len(channel))
	}
	return uint32(id)
}

func normalizeRoles(roles []NodeRole) []NodeRole {
	out := make([]NodeRole, 0, len(roles))
	seen := make(map[NodeRole]struct{}, len(roles))
	for _, role := range roles {
		normalized := NodeRole(strings.ToUpper(strings.TrimSpace(string(role))))
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

func normalizeChannels(channels []ChannelClass) []ChannelClass {
	out := make([]ChannelClass, 0, len(channels))
	seen := make(map[ChannelClass]struct{}, len(channels))
	for _, channel := range channels {
		normalized := ChannelClass(strings.ToUpper(strings.TrimSpace(string(channel))))
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.SliceStable(out, func(i, j int) bool {
		leftPriority := PriorityForChannel(out[i])
		rightPriority := PriorityForChannel(out[j])
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return channelSortRank(out[i]) < channelSortRank(out[j])
	})
	return out
}

func validateChannels(channels []ChannelClass) error {
	if len(channels) == 0 || len(channels) > MaxChannelPolicies {
		return fmt.Errorf("networking channels must be between 1 and %d", MaxChannelPolicies)
	}
	seen := make(map[ChannelClass]struct{}, len(channels))
	for _, channel := range channels {
		if !IsChannelClass(channel) {
			return fmt.Errorf("unknown networking channel %q", channel)
		}
		if _, found := seen[channel]; found {
			return errors.New("networking duplicate channel")
		}
		seen[channel] = struct{}{}
	}
	return nil
}

func normalizeStringSet(fieldName string, values []string, maxBytes int) ([]string, error) {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if len(normalized) > maxBytes {
			return nil, fmt.Errorf("networking %s must be <= %d bytes", fieldName, maxBytes)
		}
		if hasControl(normalized) {
			return nil, fmt.Errorf("networking %s must not contain control characters", fieldName)
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sortStrings(out)
	return out, nil
}

func validateIdentifierSet(fieldName string, values []string, maxBytes int) error {
	normalized, err := normalizeStringSet(fieldName, values, maxBytes)
	if err != nil {
		return err
	}
	if len(normalized) != len(values) {
		return fmt.Errorf("networking %s set must be canonical and duplicate-free", fieldName)
	}
	for i := range values {
		if values[i] != normalized[i] {
			return fmt.Errorf("networking %s set must be sorted canonically", fieldName)
		}
	}
	for _, value := range values {
		for _, r := range value {
			if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
				continue
			}
			return fmt.Errorf("networking %s contains invalid character", fieldName)
		}
	}
	return nil
}

func hasControl(value string) bool {
	for _, r := range value {
		if r < ' ' || r == 0x7f {
			return true
		}
	}
	return false
}

func hasRole(roles []NodeRole, role NodeRole) bool {
	for _, candidate := range roles {
		if candidate == role {
			return true
		}
	}
	return false
}

func intersectStrings(sets ...[]string) []string {
	if len(sets) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, set := range sets {
		normalized, _ := normalizeStringSet("protocol", set, MaxProtocolIDBytes)
		for _, value := range normalized {
			counts[value]++
		}
	}
	out := make([]string, 0)
	for value, count := range counts {
		if count == len(sets) {
			out = append(out, value)
		}
	}
	sortStrings(out)
	return out
}

func latencyScore(latencyMillis uint64) uint32 {
	switch {
	case latencyMillis == 0:
		return BasisPoints
	case latencyMillis >= 10_000:
		return 0
	default:
		return uint32(uint64(BasisPoints) - (latencyMillis * uint64(BasisPoints) / 10_000))
	}
}

func throughputScore(bytesPerSecond uint64) uint32 {
	const target = uint64(64 << 20)
	if bytesPerSecond >= target {
		return BasisPoints
	}
	return uint32(bytesPerSecond * uint64(BasisPoints) / target)
}

func invalidMessagePenalty(count uint64) uint32 {
	penalty := count * 500
	if penalty > uint64(BasisPoints) {
		return BasisPoints
	}
	return uint32(penalty)
}

func sha256Digest(domain, left, right string, index, total uint64, bytes []byte) string {
	h := sha256.New()
	writeString(h, domain)
	writeString(h, left)
	writeString(h, right)
	writeUint64(h, index)
	writeUint64(h, total)
	writeBytes(h, bytes)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func maxUint64(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}
