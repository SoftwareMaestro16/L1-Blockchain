package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultNodeRecordVersion	= uint64(1)

	RoutingNodeKeyPrefix		= "routing/nodes"
	RoutingZoneKeyPrefix		= "routing/zones"
	RoutingServiceKeyPrefix		= "routing/services"
	RoutingReputationKeyPrefix	= "routing/reputation"
	RoutingCacheKeyPrefix		= "routing/cache"
	RoutingTableKeyPrefix		= "routing/table"
)

type NodeLatencyVectorEntry struct {
	NodeID		string
	PeerNodeID	string
	ZoneID		string
	LatencyMillis	uint64
	SampleHeight	uint64
	VectorHash	string
}

type ReputationCommitment struct {
	NodeID		string
	Reputation	PeerScore
	EvidenceRoot	string
	UpdatedHeight	uint64
	CommitmentHash	string
}

type LookupCacheRecord struct {
	LookupKey	string
	QueryHash	string
	ResponseHash	string
	ExpiryHeight	uint64
	ProofHash	string
	CacheHash	string
}

type RoutingTable struct {
	Epoch		uint64
	Routes		[]DNLRoutingTableEntry
	TableRoot	string
}

type RoutingIndexRecord struct {
	Key		string
	Value		string
	EntryHash	string
}

type DNLRoutingState struct {
	Nodes		[]NodeRecord
	ZoneIndex	[]RoutingIndexRecord
	ServiceIndex	[]RoutingIndexRecord
	Reputation	[]ReputationCommitment
	Cache		[]LookupCacheRecord
	Tables		[]RoutingTable
	Height		uint64
	NodesRoot	string
	ZonesRoot	string
	ServicesRoot	string
	ReputationRoot	string
	CacheRoot	string
	TablesRoot	string
	StateRoot	string
}

func NewReputationCommitment(commitment ReputationCommitment) (ReputationCommitment, error) {
	commitment = NormalizeReputationCommitment(commitment)
	if commitment.CommitmentHash == "" {
		commitment.CommitmentHash = ComputeReputationCommitmentHash(commitment)
	}
	return commitment, commitment.Validate()
}

func NewNodeLatencyVectorEntry(entry NodeLatencyVectorEntry) (NodeLatencyVectorEntry, error) {
	entry = NormalizeNodeLatencyVectorEntry(entry)
	if entry.VectorHash == "" {
		entry.VectorHash = ComputeNodeLatencyVectorHash(entry)
	}
	return entry, entry.Validate()
}

func NewLookupCacheRecord(record LookupCacheRecord) (LookupCacheRecord, error) {
	record = NormalizeLookupCacheRecord(record)
	if record.LookupKey == "" {
		record.LookupKey = ComputeLookupCacheKey(record.QueryHash, record.ResponseHash)
	}
	if record.CacheHash == "" {
		record.CacheHash = ComputeLookupCacheRecordHash(record)
	}
	return record, record.Validate()
}

func NewRoutingTable(table RoutingTable) (RoutingTable, error) {
	table = NormalizeRoutingTable(table)
	if table.TableRoot == "" {
		table.TableRoot = ComputeRoutingTableRoot(table)
	}
	return table, table.Validate()
}

func BuildDNLRoutingState(nodes []NodeRecord, reputation []ReputationCommitment, cache []LookupCacheRecord, tables []RoutingTable, height uint64) (DNLRoutingState, error) {
	state := DNLRoutingState{
		Nodes:		normalizeRoutingNodeRecords(nodes),
		Reputation:	normalizeReputationCommitments(reputation),
		Cache:		normalizeLookupCacheRecords(cache),
		Tables:		normalizeRoutingTables(tables),
		Height:		height,
	}
	if err := state.populateIndexes(); err != nil {
		return DNLRoutingState{}, err
	}
	if err := state.ValidateFormat(); err != nil {
		return DNLRoutingState{}, err
	}
	state.NodesRoot = ComputeRoutingNodesRoot(state.Nodes)
	state.ZonesRoot = ComputeRoutingIndexRoot(state.ZoneIndex)
	state.ServicesRoot = ComputeRoutingIndexRoot(state.ServiceIndex)
	state.ReputationRoot = ComputeRoutingReputationRoot(state.Reputation)
	state.CacheRoot = ComputeLookupCacheRoot(state.Cache)
	state.TablesRoot = ComputeRoutingTablesRoot(state.Tables)
	state.StateRoot = ComputeDNLRoutingStateRoot(state)
	return state, state.Validate()
}

func RoutingNodeKey(nodeID string) (string, error) {
	nodeID = normalizeHashText(nodeID)
	if err := ValidateHash("networking routing node id", nodeID); err != nil {
		return "", err
	}
	return RoutingNodeKeyPrefix + "/" + nodeID, nil
}

func RoutingZoneKey(zoneID, nodeID string) (string, error) {
	if err := validateIdentifierSet("routing zone id", []string{zoneID}, MaxZoneIDBytes); err != nil {
		return "", err
	}
	nodeID = normalizeHashText(nodeID)
	if err := ValidateHash("networking routing zone node id", nodeID); err != nil {
		return "", err
	}
	return RoutingZoneKeyPrefix + "/" + zoneID + "/" + nodeID, nil
}

func RoutingServiceKey(serviceID, nodeID string) (string, error) {
	if err := validateIdentifierSet("routing service id", []string{serviceID}, MaxServiceIDBytes); err != nil {
		return "", err
	}
	nodeID = normalizeHashText(nodeID)
	if err := ValidateHash("networking routing service node id", nodeID); err != nil {
		return "", err
	}
	return RoutingServiceKeyPrefix + "/" + serviceID + "/" + nodeID, nil
}

func RoutingReputationKey(nodeID string) (string, error) {
	nodeID = normalizeHashText(nodeID)
	if err := ValidateHash("networking routing reputation node id", nodeID); err != nil {
		return "", err
	}
	return RoutingReputationKeyPrefix + "/" + nodeID, nil
}

func RoutingCacheKey(lookupKey string) (string, error) {
	lookupKey = normalizeHashText(lookupKey)
	if err := ValidateHash("networking routing cache lookup key", lookupKey); err != nil {
		return "", err
	}
	return RoutingCacheKeyPrefix + "/" + lookupKey, nil
}

func RoutingTableKey(epoch uint64) (string, error) {
	if epoch == 0 {
		return "", errors.New("networking routing table epoch must be positive")
	}
	return fmt.Sprintf("%s/%020d", RoutingTableKeyPrefix, epoch), nil
}

func QueryRoutingNode(state DNLRoutingState, nodeID string) (NodeRecord, bool) {
	nodeID = normalizeHashText(nodeID)
	for _, node := range state.Nodes {
		node = NormalizeNodeRecord(node)
		if node.NodeID == nodeID {
			return node, true
		}
	}
	return NodeRecord{}, false
}

func QueryRoutingNodesByZone(state DNLRoutingState, zoneID string) []NodeRecord {
	out := make([]NodeRecord, 0)
	for _, node := range state.Nodes {
		node = NormalizeNodeRecord(node)
		if containsString(node.ZonesSupported, zoneID) {
			out = append(out, node)
		}
	}
	return normalizeRoutingNodeRecords(out)
}

func QueryRoutingNodesByService(state DNLRoutingState, serviceID string) []NodeRecord {
	out := make([]NodeRecord, 0)
	for _, node := range state.Nodes {
		node = NormalizeNodeRecord(node)
		if containsString(node.ServiceIDs, serviceID) {
			out = append(out, node)
		}
	}
	return normalizeRoutingNodeRecords(out)
}

func LookupRoutingTable(state DNLRoutingState, epoch uint64) (RoutingTable, bool) {
	for _, table := range state.Tables {
		if table.Epoch == epoch {
			return table, true
		}
	}
	return RoutingTable{}, false
}

func NormalizeReputationCommitment(commitment ReputationCommitment) ReputationCommitment {
	commitment.NodeID = normalizeHashText(commitment.NodeID)
	commitment.EvidenceRoot = normalizeHashText(commitment.EvidenceRoot)
	commitment.CommitmentHash = normalizeHashText(commitment.CommitmentHash)
	return commitment
}

func NormalizeNodeLatencyVectorEntry(entry NodeLatencyVectorEntry) NodeLatencyVectorEntry {
	entry.NodeID = normalizeHashText(entry.NodeID)
	entry.PeerNodeID = normalizeHashText(entry.PeerNodeID)
	entry.ZoneID = strings.TrimSpace(entry.ZoneID)
	entry.VectorHash = normalizeHashText(entry.VectorHash)
	return entry
}

func NormalizeLookupCacheRecord(record LookupCacheRecord) LookupCacheRecord {
	record.LookupKey = normalizeHashText(record.LookupKey)
	record.QueryHash = normalizeHashText(record.QueryHash)
	record.ResponseHash = normalizeHashText(record.ResponseHash)
	record.ProofHash = normalizeHashText(record.ProofHash)
	record.CacheHash = normalizeHashText(record.CacheHash)
	return record
}

func NormalizeRoutingTable(table RoutingTable) RoutingTable {
	table.Routes = normalizeDNLRoutes(table.Routes)
	table.TableRoot = normalizeHashText(table.TableRoot)
	return table
}

func IsReputationCommitmentSet(commitment ReputationCommitment) bool {
	commitment = NormalizeReputationCommitment(commitment)
	return commitment.NodeID != "" ||
		commitment.EvidenceRoot != "" ||
		commitment.CommitmentHash != "" ||
		commitment.UpdatedHeight != 0 ||
		commitment.Reputation.ScoreBps != 0 ||
		commitment.Reputation.LatencyBps != 0 ||
		commitment.Reputation.ReliabilityBps != 0 ||
		commitment.Reputation.ThroughputBps != 0 ||
		commitment.Reputation.PenaltyBps != 0
}

func (entry NodeLatencyVectorEntry) Validate() error {
	entry = NormalizeNodeLatencyVectorEntry(entry)
	if err := ValidateHash("networking latency vector node id", entry.NodeID); err != nil {
		return err
	}
	if err := ValidateHash("networking latency vector peer node id", entry.PeerNodeID); err != nil {
		return err
	}
	if err := validateIdentifierSet("latency vector zone id", []string{entry.ZoneID}, MaxZoneIDBytes); err != nil {
		return err
	}
	if entry.LatencyMillis == 0 {
		return errors.New("networking latency vector latency must be positive")
	}
	if entry.SampleHeight == 0 {
		return errors.New("networking latency vector sample height must be positive")
	}
	if err := ValidateHash("networking latency vector hash", entry.VectorHash); err != nil {
		return err
	}
	if entry.VectorHash != ComputeNodeLatencyVectorHash(entry) {
		return errors.New("networking latency vector hash mismatch")
	}
	return nil
}

func (commitment ReputationCommitment) Validate() error {
	commitment = NormalizeReputationCommitment(commitment)
	if err := ValidateHash("networking reputation node id", commitment.NodeID); err != nil {
		return err
	}
	if err := ValidateHash("networking reputation evidence root", commitment.EvidenceRoot); err != nil {
		return err
	}
	if commitment.UpdatedHeight == 0 {
		return errors.New("networking reputation updated height must be positive")
	}
	if err := validatePeerScore(commitment.Reputation); err != nil {
		return err
	}
	if err := ValidateHash("networking reputation commitment hash", commitment.CommitmentHash); err != nil {
		return err
	}
	if commitment.CommitmentHash != ComputeReputationCommitmentHash(commitment) {
		return errors.New("networking reputation commitment hash mismatch")
	}
	return nil
}

func (record LookupCacheRecord) Validate() error {
	record = NormalizeLookupCacheRecord(record)
	if err := ValidateHash("networking lookup cache key", record.LookupKey); err != nil {
		return err
	}
	if record.LookupKey != ComputeLookupCacheKey(record.QueryHash, record.ResponseHash) {
		return errors.New("networking lookup cache key mismatch")
	}
	if err := ValidateHash("networking lookup cache query hash", record.QueryHash); err != nil {
		return err
	}
	if err := ValidateHash("networking lookup cache response hash", record.ResponseHash); err != nil {
		return err
	}
	if record.ExpiryHeight == 0 {
		return errors.New("networking lookup cache expiry height must be positive")
	}
	if err := ValidateHash("networking lookup cache proof hash", record.ProofHash); err != nil {
		return err
	}
	if err := ValidateHash("networking lookup cache record hash", record.CacheHash); err != nil {
		return err
	}
	if record.CacheHash != ComputeLookupCacheRecordHash(record) {
		return errors.New("networking lookup cache record hash mismatch")
	}
	return nil
}

func (table RoutingTable) Validate() error {
	table = NormalizeRoutingTable(table)
	if table.Epoch == 0 {
		return errors.New("networking routing table epoch must be positive")
	}
	if len(table.Routes) == 0 {
		return errors.New("networking routing table requires routes")
	}
	if err := validateDNLRoutes(table.Routes, nil); err != nil {
		return err
	}
	if err := ValidateHash("networking routing table root", table.TableRoot); err != nil {
		return err
	}
	if table.TableRoot != ComputeRoutingTableRoot(table) {
		return errors.New("networking routing table root mismatch")
	}
	return nil
}

func (state DNLRoutingState) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("networking DNL routing state height must be positive")
	}
	if err := validateRoutingNodes(state.Nodes); err != nil {
		return err
	}
	if err := validateRoutingIndexRecords("zone", state.ZoneIndex); err != nil {
		return err
	}
	if err := validateRoutingIndexRecords("service", state.ServiceIndex); err != nil {
		return err
	}
	if err := validateRoutingReputations(state.Reputation, state.Nodes); err != nil {
		return err
	}
	if err := validateLookupCacheRecords(state.Cache, state.Height); err != nil {
		return err
	}
	if err := validateRoutingTables(state.Tables); err != nil {
		return err
	}
	if state.StateRoot != "" {
		for _, field := range []struct {
			name	string
			value	string
		}{
			{"nodes root", state.NodesRoot},
			{"zones root", state.ZonesRoot},
			{"services root", state.ServicesRoot},
			{"reputation root", state.ReputationRoot},
			{"cache root", state.CacheRoot},
			{"tables root", state.TablesRoot},
			{"state root", state.StateRoot},
		} {
			if err := ValidateHash("networking DNL routing "+field.name, field.value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (state DNLRoutingState) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.StateRoot == "" {
		return errors.New("networking DNL routing state root is required")
	}
	expected := DNLRoutingState{
		Nodes:		state.Nodes,
		ZoneIndex:	state.ZoneIndex,
		ServiceIndex:	state.ServiceIndex,
		Reputation:	state.Reputation,
		Cache:		state.Cache,
		Tables:		state.Tables,
		Height:		state.Height,
		NodesRoot:	ComputeRoutingNodesRoot(state.Nodes),
		ZonesRoot:	ComputeRoutingIndexRoot(state.ZoneIndex),
		ServicesRoot:	ComputeRoutingIndexRoot(state.ServiceIndex),
		ReputationRoot:	ComputeRoutingReputationRoot(state.Reputation),
		CacheRoot:	ComputeLookupCacheRoot(state.Cache),
		TablesRoot:	ComputeRoutingTablesRoot(state.Tables),
	}
	expected.StateRoot = ComputeDNLRoutingStateRoot(expected)
	if state.NodesRoot != expected.NodesRoot ||
		state.ZonesRoot != expected.ZonesRoot ||
		state.ServicesRoot != expected.ServicesRoot ||
		state.ReputationRoot != expected.ReputationRoot ||
		state.CacheRoot != expected.CacheRoot ||
		state.TablesRoot != expected.TablesRoot ||
		state.StateRoot != expected.StateRoot {
		return errors.New("networking DNL routing state root mismatch")
	}
	return nil
}

func ComputeNodeLatencyVectorHash(entry NodeLatencyVectorEntry) string {
	entry = NormalizeNodeLatencyVectorEntry(entry)
	return HashParts("node-latency-vector-entry", entry.NodeID, entry.PeerNodeID, entry.ZoneID, fmt.Sprintf("%d", entry.LatencyMillis), fmt.Sprintf("%d", entry.SampleHeight))
}

func ComputeReputationCommitmentHash(commitment ReputationCommitment) string {
	commitment = NormalizeReputationCommitment(commitment)
	return HashParts(
		"routing-reputation-commitment",
		commitment.NodeID,
		fmt.Sprintf("%d", commitment.Reputation.ScoreBps),
		fmt.Sprintf("%d", commitment.Reputation.LatencyBps),
		fmt.Sprintf("%d", commitment.Reputation.ReliabilityBps),
		fmt.Sprintf("%d", commitment.Reputation.ThroughputBps),
		fmt.Sprintf("%d", commitment.Reputation.PenaltyBps),
		commitment.EvidenceRoot,
		fmt.Sprintf("%d", commitment.UpdatedHeight),
	)
}

func ComputeLookupCacheKey(queryHash, responseHash string) string {
	return HashParts("routing-lookup-cache-key", normalizeHashText(queryHash), normalizeHashText(responseHash))
}

func ComputeLookupCacheRecordHash(record LookupCacheRecord) string {
	record = NormalizeLookupCacheRecord(record)
	return HashParts("routing-lookup-cache-record", record.LookupKey, record.QueryHash, record.ResponseHash, fmt.Sprintf("%d", record.ExpiryHeight), record.ProofHash)
}

func ComputeRoutingTableRoot(table RoutingTable) string {
	table = NormalizeRoutingTable(table)
	parts := []string{"routing-table-root", fmt.Sprintf("%d", table.Epoch), fmt.Sprintf("%d", len(table.Routes))}
	for _, route := range table.Routes {
		parts = append(parts, route.EntryHash)
	}
	return HashParts(parts...)
}

func ComputeNodeRecordCommitmentHash(record NodeRecord) string {
	record = NormalizeNodeRecord(record)
	latencies := normalizeNodeLatencyVector(record.LatencyVector)
	parts := []string{
		"routing-node-record",
		record.NodeID,
		fmt.Sprintf("%x", record.PublicKey),
		record.OperatorAddress,
		record.NetworkAddressesHash,
		fmt.Sprintf("%d", record.RecordVersion),
		fmt.Sprintf("%d", record.ExpiresHeight),
		record.Reputation.CommitmentHash,
		fmt.Sprintf("%x", record.Signature),
		fmt.Sprintf("%d", len(record.ZonesSupported)),
		fmt.Sprintf("%d", len(record.ServiceIDs)),
		fmt.Sprintf("%d", len(record.SupportedProtocols)),
		fmt.Sprintf("%d", len(latencies)),
	}
	parts = append(parts, record.ZonesSupported...)
	parts = append(parts, record.ServiceIDs...)
	parts = append(parts, record.SupportedProtocols...)
	for _, latency := range latencies {
		parts = append(parts, latency.VectorHash)
	}
	return HashParts(parts...)
}

func ComputeRoutingIndexEntryHash(entry RoutingIndexRecord) string {
	return HashParts("routing-index-record", entry.Key, entry.Value)
}

func ComputeRoutingNodesRoot(nodes []NodeRecord) string {
	ordered := normalizeRoutingNodeRecords(nodes)
	parts := []string{"routing-nodes-root", fmt.Sprintf("%d", len(ordered))}
	for _, node := range ordered {
		parts = append(parts, ComputeNodeRecordCommitmentHash(node))
	}
	return HashParts(parts...)
}

func ComputeRoutingIndexRoot(records []RoutingIndexRecord) string {
	ordered := normalizeRoutingIndexRecords(records)
	parts := []string{"routing-index-root", fmt.Sprintf("%d", len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.EntryHash)
	}
	return HashParts(parts...)
}

func ComputeRoutingReputationRoot(reputation []ReputationCommitment) string {
	ordered := normalizeReputationCommitments(reputation)
	parts := []string{"routing-reputation-root", fmt.Sprintf("%d", len(ordered))}
	for _, commitment := range ordered {
		parts = append(parts, commitment.CommitmentHash)
	}
	return HashParts(parts...)
}

func ComputeLookupCacheRoot(records []LookupCacheRecord) string {
	ordered := normalizeLookupCacheRecords(records)
	parts := []string{"routing-cache-root", fmt.Sprintf("%d", len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.CacheHash)
	}
	return HashParts(parts...)
}

func ComputeRoutingTablesRoot(tables []RoutingTable) string {
	ordered := normalizeRoutingTables(tables)
	parts := []string{"routing-tables-root", fmt.Sprintf("%d", len(ordered))}
	for _, table := range ordered {
		parts = append(parts, table.TableRoot)
	}
	return HashParts(parts...)
}

func ComputeDNLRoutingStateRoot(state DNLRoutingState) string {
	return HashParts(
		"dnl-routing-state-root",
		fmt.Sprintf("%d", state.Height),
		state.NodesRoot,
		state.ZonesRoot,
		state.ServicesRoot,
		state.ReputationRoot,
		state.CacheRoot,
		state.TablesRoot,
	)
}

func (state *DNLRoutingState) populateIndexes() error {
	zones := make([]RoutingIndexRecord, 0)
	services := make([]RoutingIndexRecord, 0)
	for _, node := range state.Nodes {
		node = NormalizeNodeRecord(node)
		for _, zoneID := range node.ZonesSupported {
			key, err := RoutingZoneKey(zoneID, node.NodeID)
			if err != nil {
				return err
			}
			zones = append(zones, newRoutingIndexRecord(key, node.NodeID))
		}
		for _, serviceID := range node.ServiceIDs {
			key, err := RoutingServiceKey(serviceID, node.NodeID)
			if err != nil {
				return err
			}
			services = append(services, newRoutingIndexRecord(key, node.NodeID))
		}
	}
	state.ZoneIndex = normalizeRoutingIndexRecords(zones)
	state.ServiceIndex = normalizeRoutingIndexRecords(services)
	return nil
}

func newRoutingIndexRecord(key, value string) RoutingIndexRecord {
	return RoutingIndexRecord{Key: key, Value: value, EntryHash: ComputeRoutingIndexEntryHash(RoutingIndexRecord{Key: key, Value: value})}
}

func mergeStringSets(maxBytes int, sets ...[]string) []string {
	out := make([]string, 0)
	seen := map[string]struct{}{}
	for _, set := range sets {
		normalized, err := normalizeStringSet("merged", set, maxBytes)
		if err != nil {
			continue
		}
		for _, value := range normalized {
			if _, found := seen[value]; found {
				continue
			}
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	sortStrings(out)
	return out
}

func normalizePreferredStringSet(maxBytes int, preferred []string, fallbacks ...[]string) []string {
	normalizedPreferred, err := normalizeStringSet("preferred", preferred, maxBytes)
	if err == nil && len(normalizedPreferred) > 0 {
		return normalizedPreferred
	}
	return mergeStringSets(maxBytes, fallbacks...)
}

func normalizeNodeLatencyVector(entries []NodeLatencyVectorEntry) []NodeLatencyVectorEntry {
	out := make([]NodeLatencyVectorEntry, len(entries))
	for i, entry := range entries {
		out[i] = NormalizeNodeLatencyVectorEntry(entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ZoneID != out[j].ZoneID {
			return out[i].ZoneID < out[j].ZoneID
		}
		if out[i].PeerNodeID != out[j].PeerNodeID {
			return out[i].PeerNodeID < out[j].PeerNodeID
		}
		return out[i].VectorHash < out[j].VectorHash
	})
	return out
}

func normalizeRoutingNodeRecords(nodes []NodeRecord) []NodeRecord {
	out := make([]NodeRecord, len(nodes))
	for i, node := range nodes {
		out[i] = NormalizeNodeRecord(node)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}

func normalizeRoutingIndexRecords(records []RoutingIndexRecord) []RoutingIndexRecord {
	out := make([]RoutingIndexRecord, len(records))
	for i, record := range records {
		out[i] = RoutingIndexRecord{Key: strings.TrimSpace(record.Key), Value: normalizeHashText(record.Value), EntryHash: normalizeHashText(record.EntryHash)}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func normalizeReputationCommitments(records []ReputationCommitment) []ReputationCommitment {
	out := make([]ReputationCommitment, len(records))
	for i, record := range records {
		out[i] = NormalizeReputationCommitment(record)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}

func normalizeLookupCacheRecords(records []LookupCacheRecord) []LookupCacheRecord {
	out := make([]LookupCacheRecord, len(records))
	for i, record := range records {
		out[i] = NormalizeLookupCacheRecord(record)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].LookupKey < out[j].LookupKey })
	return out
}

func normalizeRoutingTables(tables []RoutingTable) []RoutingTable {
	out := make([]RoutingTable, len(tables))
	for i, table := range tables {
		out[i] = NormalizeRoutingTable(table)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Epoch < out[j].Epoch })
	return out
}

func validateNodeLatencyVector(entries []NodeLatencyVectorEntry, nodeID string) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if nodeID != "" && entry.NodeID != nodeID {
			return errors.New("networking latency vector must reference node")
		}
		key := entry.ZoneID + "/" + entry.PeerNodeID
		if _, found := seen[key]; found {
			return errors.New("networking duplicate latency vector entry")
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("networking latency vector must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validatePeerScore(score PeerScore) error {
	if score.ScoreBps > BasisPoints ||
		score.LatencyBps > BasisPoints ||
		score.ReliabilityBps > BasisPoints ||
		score.ThroughputBps > BasisPoints ||
		score.PenaltyBps > BasisPoints {
		return fmt.Errorf("networking peer score components must be <= %d bps", BasisPoints)
	}
	return nil
}

func validateRoutingNodes(nodes []NodeRecord) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, node := range nodes {
		node = NormalizeNodeRecord(node)
		if err := node.ValidateBasic(); err != nil {
			return err
		}
		if _, found := seen[node.NodeID]; found {
			return errors.New("networking duplicate routing node")
		}
		seen[node.NodeID] = struct{}{}
		if previous != "" && previous >= node.NodeID {
			return errors.New("networking routing nodes must be sorted canonically")
		}
		previous = node.NodeID
	}
	return nil
}

func validateRoutingIndexRecords(name string, records []RoutingIndexRecord) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, record := range records {
		if record.Key == "" {
			return fmt.Errorf("networking routing %s index key is required", name)
		}
		if err := ValidateHash("networking routing "+name+" index value", record.Value); err != nil {
			return err
		}
		if err := ValidateHash("networking routing "+name+" index hash", record.EntryHash); err != nil {
			return err
		}
		if record.EntryHash != ComputeRoutingIndexEntryHash(record) {
			return fmt.Errorf("networking routing %s index hash mismatch", name)
		}
		if _, found := seen[record.Key]; found {
			return fmt.Errorf("networking duplicate routing %s index", name)
		}
		seen[record.Key] = struct{}{}
		if previous != "" && previous >= record.Key {
			return fmt.Errorf("networking routing %s index must be sorted canonically", name)
		}
		previous = record.Key
	}
	return nil
}

func validateRoutingReputations(reputation []ReputationCommitment, nodes []NodeRecord) error {
	nodeIDs := map[string]struct{}{}
	for _, node := range nodes {
		nodeIDs[NormalizeNodeRecord(node).NodeID] = struct{}{}
	}
	seen := map[string]struct{}{}
	previous := ""
	for _, commitment := range reputation {
		commitment = NormalizeReputationCommitment(commitment)
		if err := commitment.Validate(); err != nil {
			return err
		}
		if _, found := nodeIDs[commitment.NodeID]; !found {
			return errors.New("networking routing reputation must reference registered node")
		}
		if _, found := seen[commitment.NodeID]; found {
			return errors.New("networking duplicate routing reputation")
		}
		seen[commitment.NodeID] = struct{}{}
		if previous != "" && previous >= commitment.NodeID {
			return errors.New("networking routing reputation must be sorted canonically")
		}
		previous = commitment.NodeID
	}
	for nodeID := range nodeIDs {
		if _, found := seen[nodeID]; !found {
			return errors.New("networking routing node requires reputation commitment")
		}
	}
	return nil
}

func validateLookupCacheRecords(records []LookupCacheRecord, height uint64) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, record := range records {
		record = NormalizeLookupCacheRecord(record)
		if err := record.Validate(); err != nil {
			return err
		}
		if height > 0 && height > record.ExpiryHeight {
			return errors.New("networking routing cache record is expired")
		}
		if _, found := seen[record.LookupKey]; found {
			return errors.New("networking duplicate routing cache record")
		}
		seen[record.LookupKey] = struct{}{}
		if previous != "" && previous >= record.LookupKey {
			return errors.New("networking routing cache records must be sorted canonically")
		}
		previous = record.LookupKey
	}
	return nil
}

func validateRoutingTables(tables []RoutingTable) error {
	seen := map[uint64]struct{}{}
	var previous uint64
	for i, table := range tables {
		table = NormalizeRoutingTable(table)
		if err := table.Validate(); err != nil {
			return err
		}
		if _, found := seen[table.Epoch]; found {
			return errors.New("networking duplicate routing table epoch")
		}
		seen[table.Epoch] = struct{}{}
		if i > 0 && previous >= table.Epoch {
			return errors.New("networking routing tables must be sorted canonically")
		}
		previous = table.Epoch
	}
	return nil
}
