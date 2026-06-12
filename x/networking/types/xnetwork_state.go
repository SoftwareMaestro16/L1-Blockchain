package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	XNetworkParamsKeyPrefix		= "network/params"
	XNetworkNodesKeyPrefix		= "network/nodes"
	XNetworkRolesKeyPrefix		= "network/roles"
	XNetworkOverlaysKeyPrefix	= "network/overlays"
	XNetworkDiscoveryKeyPrefix	= "network/discovery"
	XNetworkReputationKeyPrefix	= "network/reputation"
	XNetworkEvidenceKeyPrefix	= "network/evidence"
)

type XNetworkMsgType string

const (
	MsgRegisterNode			XNetworkMsgType	= "MsgRegisterNode"
	MsgUpdateNode			XNetworkMsgType	= "MsgUpdateNode"
	MsgRenewNode			XNetworkMsgType	= "MsgRenewNode"
	MsgRevokeNode			XNetworkMsgType	= "MsgRevokeNode"
	MsgSubmitNetworkEvidence	XNetworkMsgType	= "MsgSubmitNetworkEvidence"
)

type XNetworkQueryType string

const (
	QueryNode		XNetworkQueryType	= "QueryNode"
	QueryNodesByRole	XNetworkQueryType	= "QueryNodesByRole"
	QueryOverlay		XNetworkQueryType	= "QueryOverlay"
	QueryDiscoveryRecord	XNetworkQueryType	= "QueryDiscoveryRecord"
	QueryNetworkParams	XNetworkQueryType	= "QueryNetworkParams"
	QueryNetworkEvidence	XNetworkQueryType	= "QueryNetworkEvidence"
)

type NetworkEvidenceType string

const (
	NetworkEvidenceInvalidMessage		NetworkEvidenceType	= "invalid_message"
	NetworkEvidenceConflictingBroadcast	NetworkEvidenceType	= "conflicting_broadcast"
	NetworkEvidenceDiscoveryForgery		NetworkEvidenceType	= "discovery_forgery"
	NetworkEvidenceChunkCorruption		NetworkEvidenceType	= "chunk_corruption"
	NetworkEvidenceRoutingManipulation	NetworkEvidenceType	= "routing_manipulation"
	NetworkEvidenceCrossZoneReplay		NetworkEvidenceType	= "cross_zone_replay"
	NetworkEvidenceBandwidthExhaustion	NetworkEvidenceType	= "bandwidth_exhaustion"
)

type XNetworkParams struct {
	NetworkSaltHash		string
	MaxNodeRecordTTL	uint64
	MaxDiscoveryRecordTTL	uint64
	EvidenceHorizon		uint64
	ReputationDecayBps	uint32
	MaxEvidenceBytes	uint64
}

type XNetworkState struct {
	Params			XNetworkParams
	Network			NetworkingState
	DiscoveryRecords	[]DiscoveryRecord
	Reputation		[]NetworkReputationRecord
	Evidence		[]NetworkEvidenceRecord
	StateRoot		string
}

type NetworkReputationRecord struct {
	NodeID			string
	Score			PeerScore
	LastUpdatedHeight	uint64
	EvidenceHash		string
	ConsensusEligible	bool
}

type NetworkEvidenceRecord struct {
	EvidenceID	string
	EvidenceType	NetworkEvidenceType
	ReporterNodeID	string
	SubjectNodeID	string
	EvidenceHash	string
	EvidenceHeight	uint64
	PayloadBytes	uint64
	Committed	bool
}

type XNetworkStateKeys struct {
	ParamsKey	string
	NodeKeys	[]string
	RoleKeys	[]string
	OverlayKeys	[]string
	DiscoveryKeys	[]string
	ReputationKeys	[]string
	EvidenceKeys	[]string
	StateKeysRoot	string
}

type MsgRegisterNodeRequest struct {
	SignerNodeID	string
	Record		NodeRecord
	NetworkSalt	[]byte
	CurrentHeight	uint64
}

type MsgUpdateNodeRequest struct {
	SignerNodeID	string
	Record		NodeRecord
	NetworkSalt	[]byte
	CurrentHeight	uint64
}

type MsgRenewNodeRequest struct {
	SignerNodeID	string
	Record		NodeRecord
	NetworkSalt	[]byte
	CurrentHeight	uint64
}

type MsgRevokeNodeRequest struct {
	SignerNodeID	string
	NodeID		string
	ReasonHash	string
	CurrentHeight	uint64
}

type MsgSubmitNetworkEvidenceRequest struct {
	SignerNodeID	string
	Evidence	NetworkEvidenceRecord
}

type QueryNodeRequest struct {
	NodeID string
}

type QueryNodesByRoleRequest struct {
	Role		NodeRole
	CurrentHeight	uint64
}

type QueryOverlayRequest struct {
	OverlayID string
}

type QueryDiscoveryRecordRequest struct {
	RecordID string
}

type QueryNetworkParamsRequest struct{}

type QueryNetworkEvidenceRequest struct {
	EvidenceID string
}

func DefaultXNetworkParams(networkSalt []byte) XNetworkParams {
	return XNetworkParams{
		NetworkSaltHash:	hashBytes("aetra-x-network-salt-v1", networkSalt),
		MaxNodeRecordTTL:	100_000,
		MaxDiscoveryRecordTTL:	25_000,
		EvidenceHorizon:	DefaultSecurityReplayHorizon,
		ReputationDecayBps:	DefaultReputationDecayBps,
		MaxEvidenceBytes:	DefaultMaxMessageBytes,
	}
}

func NewXNetworkState(params XNetworkParams, network NetworkingState, discovery []DiscoveryRecord, reputation []NetworkReputationRecord, evidence []NetworkEvidenceRecord) (XNetworkState, error) {
	state := XNetworkState{
		Params:			params,
		Network:		network.Export(),
		DiscoveryRecords:	cloneDiscoveryRecords(discovery),
		Reputation:		cloneNetworkReputationRecords(reputation),
		Evidence:		cloneNetworkEvidenceRecords(evidence),
	}
	sortDiscoveryRecords(state.DiscoveryRecords)
	sortNetworkReputationRecords(state.Reputation)
	sortNetworkEvidenceRecords(state.Evidence)
	state.StateRoot = ComputeXNetworkStateRoot(state)
	return state, state.Validate()
}

func (s XNetworkState) Validate() error {
	state := NormalizeXNetworkState(s)
	if err := state.Params.Validate(); err != nil {
		return err
	}
	if err := state.Network.Validate(); err != nil {
		return err
	}
	for _, record := range state.DiscoveryRecords {
		record = NormalizeDiscoveryRecord(record)
		if err := ValidateHash("network discovery record id", record.RecordID); err != nil {
			return err
		}
	}
	if err := validateNetworkReputationRecords(state.Reputation); err != nil {
		return err
	}
	if err := validateNetworkEvidenceRecords(state.Evidence, state.Params); err != nil {
		return err
	}
	if state.StateRoot != ComputeXNetworkStateRoot(state) {
		return errors.New("network state root mismatch")
	}
	return nil
}

func NormalizeXNetworkState(state XNetworkState) XNetworkState {
	state.Params = NormalizeXNetworkParams(state.Params)
	state.Network = state.Network.Export()
	state.DiscoveryRecords = cloneDiscoveryRecords(state.DiscoveryRecords)
	sortDiscoveryRecords(state.DiscoveryRecords)
	state.Reputation = cloneNetworkReputationRecords(state.Reputation)
	sortNetworkReputationRecords(state.Reputation)
	state.Evidence = cloneNetworkEvidenceRecords(state.Evidence)
	sortNetworkEvidenceRecords(state.Evidence)
	state.StateRoot = normalizeHashText(state.StateRoot)
	return state
}

func NormalizeXNetworkParams(params XNetworkParams) XNetworkParams {
	params.NetworkSaltHash = normalizeHashText(params.NetworkSaltHash)
	return params
}

func (p XNetworkParams) Validate() error {
	params := NormalizeXNetworkParams(p)
	if err := ValidateHash("network params salt hash", params.NetworkSaltHash); err != nil {
		return err
	}
	if params.MaxNodeRecordTTL == 0 {
		return errors.New("network params max node record ttl must be positive")
	}
	if params.MaxDiscoveryRecordTTL == 0 {
		return errors.New("network params max discovery record ttl must be positive")
	}
	if params.EvidenceHorizon == 0 {
		return errors.New("network params evidence horizon must be positive")
	}
	if params.ReputationDecayBps > BasisPoints {
		return fmt.Errorf("network params reputation decay must be <= %d bps", BasisPoints)
	}
	if params.MaxEvidenceBytes == 0 || params.MaxEvidenceBytes > MaxAetherMeshPayloadBytes {
		return fmt.Errorf("network params max evidence bytes must be between 1 and %d", uint64(MaxAetherMeshPayloadBytes))
	}
	return nil
}

func BuildXNetworkStateKeys(state XNetworkState) (XNetworkStateKeys, error) {
	state = NormalizeXNetworkState(state)
	if err := state.Validate(); err != nil {
		return XNetworkStateKeys{}, err
	}
	keys := XNetworkStateKeys{ParamsKey: XNetworkParamsStateKey()}
	for _, record := range state.Network.NodeRecords {
		keys.NodeKeys = append(keys.NodeKeys, XNetworkNodeStateKey(record.NodeID))
		for _, role := range record.Roles {
			keys.RoleKeys = append(keys.RoleKeys, XNetworkRoleStateKey(role, record.NodeID))
		}
	}
	for _, overlay := range state.Network.OverlayDescriptors {
		keys.OverlayKeys = append(keys.OverlayKeys, XNetworkOverlayStateKey(overlay.OverlayID))
	}
	for _, record := range state.DiscoveryRecords {
		keys.DiscoveryKeys = append(keys.DiscoveryKeys, XNetworkDiscoveryStateKey(record.RecordID))
	}
	for _, record := range state.Reputation {
		keys.ReputationKeys = append(keys.ReputationKeys, XNetworkReputationStateKey(record.NodeID))
	}
	for _, record := range state.Evidence {
		keys.EvidenceKeys = append(keys.EvidenceKeys, XNetworkEvidenceStateKey(record.EvidenceID))
	}
	sort.Strings(keys.NodeKeys)
	sort.Strings(keys.RoleKeys)
	sort.Strings(keys.OverlayKeys)
	sort.Strings(keys.DiscoveryKeys)
	sort.Strings(keys.ReputationKeys)
	sort.Strings(keys.EvidenceKeys)
	keys.StateKeysRoot = ComputeXNetworkStateKeysRoot(keys)
	return keys, nil
}

func XNetworkParamsStateKey() string {
	return XNetworkParamsKeyPrefix
}

func XNetworkNodeStateKey(nodeID string) string {
	return XNetworkNodesKeyPrefix + "/" + normalizeHashText(nodeID)
}

func XNetworkRoleStateKey(role NodeRole, nodeID string) string {
	return XNetworkRolesKeyPrefix + "/" + string(role) + "/" + normalizeHashText(nodeID)
}

func XNetworkOverlayStateKey(overlayID string) string {
	return XNetworkOverlaysKeyPrefix + "/" + normalizeHashText(overlayID)
}

func XNetworkDiscoveryStateKey(recordID string) string {
	return XNetworkDiscoveryKeyPrefix + "/" + normalizeHashText(recordID)
}

func XNetworkReputationStateKey(nodeID string) string {
	return XNetworkReputationKeyPrefix + "/" + normalizeHashText(nodeID)
}

func XNetworkEvidenceStateKey(evidenceID string) string {
	return XNetworkEvidenceKeyPrefix + "/" + normalizeHashText(evidenceID)
}

func ValidateXNetworkStateKey(key string) error {
	key = strings.TrimSpace(key)
	if key == XNetworkParamsKeyPrefix {
		return nil
	}
	parts := strings.Split(key, "/")
	if len(parts) < 3 || parts[0] != "network" {
		return errors.New("network state key must use network prefix")
	}
	switch parts[1] {
	case "nodes", "overlays", "discovery", "reputation", "evidence":
		if len(parts) != 3 {
			return errors.New("network state key has invalid path length")
		}
		return ValidateHash("network state key id", parts[2])
	case "roles":
		if len(parts) != 4 {
			return errors.New("network role state key has invalid path length")
		}
		if !IsNodeRole(NodeRole(parts[2])) {
			return fmt.Errorf("unknown network role state key role %q", parts[2])
		}
		return ValidateHash("network role state key node id", parts[3])
	default:
		return fmt.Errorf("unknown network state key prefix %q", parts[1])
	}
}

func (r NetworkReputationRecord) Validate() error {
	record := NormalizeNetworkReputationRecord(r)
	if err := ValidateHash("network reputation node id", record.NodeID); err != nil {
		return err
	}
	if err := validateNetworkPeerScore(record.Score); err != nil {
		return err
	}
	if record.LastUpdatedHeight == 0 {
		return errors.New("network reputation height must be positive")
	}
	if record.ConsensusEligible {
		if err := ValidateHash("network reputation evidence hash", record.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func NormalizeNetworkReputationRecord(record NetworkReputationRecord) NetworkReputationRecord {
	record.NodeID = normalizeHashText(record.NodeID)
	record.EvidenceHash = normalizeHashText(record.EvidenceHash)
	return record
}

func (r NetworkEvidenceRecord) Validate(params XNetworkParams) error {
	record := NormalizeNetworkEvidenceRecord(r)
	if err := ValidateHash("network evidence id", record.EvidenceID); err != nil {
		return err
	}
	if record.EvidenceID != ComputeNetworkEvidenceID(record) {
		return errors.New("network evidence id mismatch")
	}
	if !IsNetworkEvidenceType(record.EvidenceType) {
		return fmt.Errorf("unknown network evidence type %q", record.EvidenceType)
	}
	if err := ValidateHash("network evidence reporter node id", record.ReporterNodeID); err != nil {
		return err
	}
	if err := ValidateHash("network evidence subject node id", record.SubjectNodeID); err != nil {
		return err
	}
	if err := ValidateHash("network evidence hash", record.EvidenceHash); err != nil {
		return err
	}
	if record.EvidenceHeight == 0 {
		return errors.New("network evidence height must be positive")
	}
	if record.PayloadBytes == 0 || record.PayloadBytes > params.MaxEvidenceBytes {
		return fmt.Errorf("network evidence payload bytes must be between 1 and %d", params.MaxEvidenceBytes)
	}
	return nil
}

func NormalizeNetworkEvidenceRecord(record NetworkEvidenceRecord) NetworkEvidenceRecord {
	record.EvidenceID = normalizeHashText(record.EvidenceID)
	record.ReporterNodeID = normalizeHashText(record.ReporterNodeID)
	record.SubjectNodeID = normalizeHashText(record.SubjectNodeID)
	record.EvidenceHash = normalizeHashText(record.EvidenceHash)
	return record
}

func NewNetworkEvidenceRecord(record NetworkEvidenceRecord) (NetworkEvidenceRecord, error) {
	record = NormalizeNetworkEvidenceRecord(record)
	if record.EvidenceID == "" {
		record.EvidenceID = ComputeNetworkEvidenceID(record)
	}
	params := DefaultXNetworkParams([]byte("network-evidence-default"))
	if err := record.Validate(params); err != nil {
		return NetworkEvidenceRecord{}, err
	}
	return record, nil
}

func ComputeNetworkEvidenceID(record NetworkEvidenceRecord) string {
	record = NormalizeNetworkEvidenceRecord(record)
	return HashParts(
		"network-evidence",
		string(record.EvidenceType),
		record.ReporterNodeID,
		record.SubjectNodeID,
		record.EvidenceHash,
		fmt.Sprintf("%d", record.EvidenceHeight),
	)
}

func (m MsgRegisterNodeRequest) ValidateBasic() error {
	return validateNodeMessage("register", m.SignerNodeID, m.Record, m.NetworkSalt, m.CurrentHeight, false)
}

func (m MsgUpdateNodeRequest) ValidateBasic() error {
	return validateNodeMessage("update", m.SignerNodeID, m.Record, m.NetworkSalt, m.CurrentHeight, false)
}

func (m MsgRenewNodeRequest) ValidateBasic() error {
	return validateNodeMessage("renew", m.SignerNodeID, m.Record, m.NetworkSalt, m.CurrentHeight, true)
}

func (m MsgRevokeNodeRequest) ValidateBasic() error {
	signer := normalizeHashText(m.SignerNodeID)
	nodeID := normalizeHashText(m.NodeID)
	if err := ValidateHash("network revoke signer", signer); err != nil {
		return err
	}
	if err := ValidateHash("network revoke node id", nodeID); err != nil {
		return err
	}
	if signer != nodeID {
		return errors.New("network revoke signer must match node id")
	}
	if err := ValidateHash("network revoke reason hash", normalizeHashText(m.ReasonHash)); err != nil {
		return err
	}
	if m.CurrentHeight == 0 {
		return errors.New("network revoke height must be positive")
	}
	return nil
}

func (m MsgSubmitNetworkEvidenceRequest) ValidateBasic(params XNetworkParams) error {
	signer := normalizeHashText(m.SignerNodeID)
	if err := ValidateHash("network evidence signer", signer); err != nil {
		return err
	}
	evidence := NormalizeNetworkEvidenceRecord(m.Evidence)
	if evidence.ReporterNodeID != signer {
		return errors.New("network evidence signer must match reporter")
	}
	return evidence.Validate(params)
}

func (q QueryNodeRequest) ValidateBasic() error {
	return ValidateHash("network query node id", normalizeHashText(q.NodeID))
}

func (q QueryNodesByRoleRequest) ValidateBasic() error {
	if !IsNodeRole(q.Role) {
		return fmt.Errorf("unknown network query role %q", q.Role)
	}
	return nil
}

func (q QueryOverlayRequest) ValidateBasic() error {
	return ValidateHash("network query overlay id", normalizeHashText(q.OverlayID))
}

func (q QueryDiscoveryRecordRequest) ValidateBasic() error {
	return ValidateHash("network query discovery record id", normalizeHashText(q.RecordID))
}

func (q QueryNetworkParamsRequest) ValidateBasic() error {
	return nil
}

func (q QueryNetworkEvidenceRequest) ValidateBasic() error {
	return ValidateHash("network query evidence id", normalizeHashText(q.EvidenceID))
}

func QueryNodeFromXNetworkState(state XNetworkState, req QueryNodeRequest) (NodeRecord, bool, error) {
	if err := req.ValidateBasic(); err != nil {
		return NodeRecord{}, false, err
	}
	state = NormalizeXNetworkState(state)
	if err := state.Validate(); err != nil {
		return NodeRecord{}, false, err
	}
	nodeID := normalizeHashText(req.NodeID)
	for _, record := range state.Network.NodeRecords {
		if record.NodeID == nodeID {
			return NormalizeNodeRecord(record), true, nil
		}
	}
	return NodeRecord{}, false, nil
}

func QueryNodesByRoleFromXNetworkState(state XNetworkState, req QueryNodesByRoleRequest) ([]NodeRecord, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}
	state = NormalizeXNetworkState(state)
	if err := state.Validate(); err != nil {
		return nil, err
	}
	out := make([]NodeRecord, 0)
	for _, record := range state.Network.NodeRecords {
		if hasRole(record.Roles, req.Role) && (req.CurrentHeight == 0 || req.CurrentHeight <= record.ExpiresHeight) {
			out = append(out, NormalizeNodeRecord(record))
		}
	}
	sortNodeRecords(out)
	return out, nil
}

func QueryOverlayFromXNetworkState(state XNetworkState, req QueryOverlayRequest) (OverlayDescriptor, bool, error) {
	if err := req.ValidateBasic(); err != nil {
		return OverlayDescriptor{}, false, err
	}
	state = NormalizeXNetworkState(state)
	if err := state.Validate(); err != nil {
		return OverlayDescriptor{}, false, err
	}
	overlayID := normalizeHashText(req.OverlayID)
	for _, desc := range state.Network.OverlayDescriptors {
		desc = NormalizeOverlayDescriptor(desc)
		if desc.OverlayID == overlayID {
			return desc, true, nil
		}
	}
	return OverlayDescriptor{}, false, nil
}

func QueryDiscoveryRecordFromXNetworkState(state XNetworkState, req QueryDiscoveryRecordRequest) (DiscoveryRecord, bool, error) {
	if err := req.ValidateBasic(); err != nil {
		return DiscoveryRecord{}, false, err
	}
	state = NormalizeXNetworkState(state)
	if err := state.Validate(); err != nil {
		return DiscoveryRecord{}, false, err
	}
	recordID := normalizeHashText(req.RecordID)
	for _, record := range state.DiscoveryRecords {
		if record.RecordID == recordID {
			return NormalizeDiscoveryRecord(record), true, nil
		}
	}
	return DiscoveryRecord{}, false, nil
}

func QueryNetworkParamsFromXNetworkState(state XNetworkState, _ QueryNetworkParamsRequest) (XNetworkParams, error) {
	state = NormalizeXNetworkState(state)
	if err := state.Validate(); err != nil {
		return XNetworkParams{}, err
	}
	return state.Params, nil
}

func QueryNetworkEvidenceFromXNetworkState(state XNetworkState, req QueryNetworkEvidenceRequest) (NetworkEvidenceRecord, bool, error) {
	if err := req.ValidateBasic(); err != nil {
		return NetworkEvidenceRecord{}, false, err
	}
	state = NormalizeXNetworkState(state)
	if err := state.Validate(); err != nil {
		return NetworkEvidenceRecord{}, false, err
	}
	evidenceID := normalizeHashText(req.EvidenceID)
	for _, evidence := range state.Evidence {
		if evidence.EvidenceID == evidenceID {
			return evidence, true, nil
		}
	}
	return NetworkEvidenceRecord{}, false, nil
}

func ComputeXNetworkStateRoot(state XNetworkState) string {
	state = NormalizeXNetworkState(state)
	parts := []string{
		"x-network-state",
		state.Params.NetworkSaltHash,
		fmt.Sprintf("%d", state.Params.MaxNodeRecordTTL),
		fmt.Sprintf("%d", state.Params.MaxDiscoveryRecordTTL),
		fmt.Sprintf("%d", state.Params.EvidenceHorizon),
		fmt.Sprintf("%d", state.Params.ReputationDecayBps),
		fmt.Sprintf("%d", state.Params.MaxEvidenceBytes),
	}
	for _, record := range state.Network.NodeRecords {
		parts = append(parts, record.NodeID)
	}
	for _, desc := range state.Network.OverlayDescriptors {
		parts = append(parts, desc.OverlayID)
	}
	for _, record := range state.DiscoveryRecords {
		parts = append(parts, record.RecordID)
	}
	for _, record := range state.Reputation {
		parts = append(parts, record.NodeID, fmt.Sprintf("%d", record.Score.ScoreBps), record.EvidenceHash)
	}
	for _, record := range state.Evidence {
		parts = append(parts, record.EvidenceID, record.EvidenceHash)
	}
	return HashParts(parts...)
}

func ComputeXNetworkStateKeysRoot(keys XNetworkStateKeys) string {
	parts := []string{"x-network-state-keys", strings.TrimSpace(keys.ParamsKey)}
	parts = append(parts, sortedStrings(keys.NodeKeys)...)
	parts = append(parts, sortedStrings(keys.RoleKeys)...)
	parts = append(parts, sortedStrings(keys.OverlayKeys)...)
	parts = append(parts, sortedStrings(keys.DiscoveryKeys)...)
	parts = append(parts, sortedStrings(keys.ReputationKeys)...)
	parts = append(parts, sortedStrings(keys.EvidenceKeys)...)
	return HashParts(parts...)
}

func IsXNetworkMsgType(msgType XNetworkMsgType) bool {
	switch msgType {
	case MsgRegisterNode, MsgUpdateNode, MsgRenewNode, MsgRevokeNode, MsgSubmitNetworkEvidence:
		return true
	default:
		return false
	}
}

func IsXNetworkQueryType(queryType XNetworkQueryType) bool {
	switch queryType {
	case QueryNode, QueryNodesByRole, QueryOverlay, QueryDiscoveryRecord, QueryNetworkParams, QueryNetworkEvidence:
		return true
	default:
		return false
	}
}

func IsNetworkEvidenceType(evidenceType NetworkEvidenceType) bool {
	switch evidenceType {
	case NetworkEvidenceInvalidMessage,
		NetworkEvidenceConflictingBroadcast,
		NetworkEvidenceDiscoveryForgery,
		NetworkEvidenceChunkCorruption,
		NetworkEvidenceRoutingManipulation,
		NetworkEvidenceCrossZoneReplay,
		NetworkEvidenceBandwidthExhaustion:
		return true
	default:
		return false
	}
}

func validateNodeMessage(action, signer string, record NodeRecord, networkSalt []byte, currentHeight uint64, renewal bool) error {
	signer = normalizeHashText(signer)
	if err := ValidateHash("network "+action+" signer", signer); err != nil {
		return err
	}
	record = NormalizeNodeRecord(record)
	if err := record.Validate(networkSalt, currentHeight); err != nil {
		return err
	}
	if record.NodeID != signer {
		return fmt.Errorf("network %s signer must match node record", action)
	}
	if renewal && record.ExpiresHeight <= currentHeight {
		return errors.New("network renew node requires future expiry")
	}
	return nil
}

func validateNetworkReputationRecords(records []NetworkReputationRecord) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		record = NormalizeNetworkReputationRecord(record)
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.NodeID]; found {
			return errors.New("network duplicate reputation record")
		}
		seen[record.NodeID] = struct{}{}
	}
	return nil
}

func validateNetworkEvidenceRecords(records []NetworkEvidenceRecord, params XNetworkParams) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		record = NormalizeNetworkEvidenceRecord(record)
		if err := record.Validate(params); err != nil {
			return err
		}
		if _, found := seen[record.EvidenceID]; found {
			return errors.New("network duplicate evidence record")
		}
		seen[record.EvidenceID] = struct{}{}
	}
	return nil
}

func validateNetworkPeerScore(score PeerScore) error {
	if score.ScoreBps > BasisPoints ||
		score.LatencyBps > BasisPoints ||
		score.ReliabilityBps > BasisPoints ||
		score.ThroughputBps > BasisPoints ||
		score.PenaltyBps > BasisPoints {
		return fmt.Errorf("network reputation peer score fields must be <= %d bps", BasisPoints)
	}
	return nil
}

func cloneNetworkReputationRecords(records []NetworkReputationRecord) []NetworkReputationRecord {
	out := make([]NetworkReputationRecord, len(records))
	for i, record := range records {
		out[i] = NormalizeNetworkReputationRecord(record)
	}
	return out
}

func cloneNetworkEvidenceRecords(records []NetworkEvidenceRecord) []NetworkEvidenceRecord {
	out := make([]NetworkEvidenceRecord, len(records))
	for i, record := range records {
		out[i] = NormalizeNetworkEvidenceRecord(record)
	}
	return out
}

func sortNetworkReputationRecords(records []NetworkReputationRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].NodeID < records[j].NodeID
	})
}

func sortNetworkEvidenceRecords(records []NetworkEvidenceRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].EvidenceID < records[j].EvidenceID
	})
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
