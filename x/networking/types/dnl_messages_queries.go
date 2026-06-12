package types

import (
	"errors"
	"fmt"
	"strings"
)

type MsgRegisterNodeRecord struct {
	Authority	string
	Record		NodeRecord
	Reputation	ReputationCommitment
	NetworkSalt	[]byte
	Height		uint64
	MessageHash	string
}

type MsgUpdateNodeRecord struct {
	Authority	string
	Record		NodeRecord
	Reputation	ReputationCommitment
	NetworkSalt	[]byte
	Height		uint64
	MessageHash	string
}

type MsgExpireNodeRecord struct {
	Authority	string
	NodeID		string
	ReasonHash	string
	Height		uint64
	MessageHash	string
}

type MsgSubmitReputationCommitment struct {
	Authority	string
	Commitment	ReputationCommitment
	Height		uint64
	MessageHash	string
}

type MsgUpdateRoutingTable struct {
	Authority	string
	Table		RoutingTable
	Height		uint64
	MessageHash	string
}

type QueryNodeRecord struct {
	NodeID		string
	IncludeProof	bool
}

type QueryNodeRecordResponse struct {
	Record	NodeRecord
	Proof	RoutingStateProof
	Found	bool
}

type QueryNodesByZone struct {
	ZoneID string
}

type QueryNodesByZoneResponse struct {
	Records	[]NodeRecord
	Proofs	[]RoutingStateProof
}

type QueryNodesByService struct {
	ServiceID string
}

type QueryNodesByServiceResponse struct {
	Records	[]NodeRecord
	Proofs	[]RoutingStateProof
}

type QueryRoutingTable struct {
	Epoch		uint64
	IncludeProof	bool
}

type QueryRoutingTableResponse struct {
	Table	RoutingTable
	Proof	RoutingStateProof
	Found	bool
}

type QueryLookupProof struct {
	Key string
}

type QueryLookupProofResponse struct {
	Proof	RoutingStateProof
	Found	bool
}

type QueryReputationCommitment struct {
	NodeID		string
	IncludeProof	bool
}

type QueryReputationCommitmentResponse struct {
	Commitment	ReputationCommitment
	Proof		RoutingStateProof
	Found		bool
}

type RoutingStateProof struct {
	Key		string
	ValueHash	string
	StateRoot	string
	Height		uint64
	Path		[]string
	ProofHash	string
}

type DNLRoutingExportManifest struct {
	Height			uint64
	NodeCount		uint64
	ZoneIndexCount		uint64
	ServiceIndexCount	uint64
	ReputationCount		uint64
	CacheCount		uint64
	TableCount		uint64
	NodesRoot		string
	ZonesRoot		string
	ServicesRoot		string
	ReputationRoot		string
	CacheRoot		string
	TablesRoot		string
	StateRoot		string
	ManifestHash		string
}

type DNLRoutingExport struct {
	State		DNLRoutingState
	Manifest	DNLRoutingExportManifest
}

func RegisterNodeRecordInRoutingState(state DNLRoutingState, msg MsgRegisterNodeRecord) (DNLRoutingState, error) {
	if err := msg.Validate(); err != nil {
		return DNLRoutingState{}, err
	}
	record := NormalizeNodeRecord(msg.Record)
	if _, found := QueryRoutingNode(state, record.NodeID); found {
		return DNLRoutingState{}, fmt.Errorf("networking DNL node %s already exists", record.NodeID)
	}
	if record.OperatorAddress != msg.Authority {
		return DNLRoutingState{}, errors.New("networking DNL register authority must match operator")
	}
	nodes := append([]NodeRecord(nil), state.Nodes...)
	nodes = append(nodes, record)
	reputation := upsertDNLReputation(state.Reputation, msg.Reputation)
	return BuildDNLRoutingState(nodes, reputation, state.Cache, state.Tables, msg.Height)
}

func UpdateNodeRecordInRoutingState(state DNLRoutingState, msg MsgUpdateNodeRecord) (DNLRoutingState, error) {
	if err := msg.Validate(); err != nil {
		return DNLRoutingState{}, err
	}
	record := NormalizeNodeRecord(msg.Record)
	nodes := append([]NodeRecord(nil), state.Nodes...)
	index, existing, found := routingNodeIndex(nodes, record.NodeID)
	if !found {
		return DNLRoutingState{}, fmt.Errorf("networking DNL node %s not found", record.NodeID)
	}
	if existing.OperatorAddress != msg.Authority || record.OperatorAddress != msg.Authority {
		return DNLRoutingState{}, errors.New("networking DNL update authority must match operator")
	}
	nodes[index] = record
	reputation := upsertDNLReputation(state.Reputation, msg.Reputation)
	return BuildDNLRoutingState(nodes, reputation, state.Cache, state.Tables, msg.Height)
}

func ExpireNodeRecordInRoutingState(state DNLRoutingState, msg MsgExpireNodeRecord) (DNLRoutingState, error) {
	if err := msg.Validate(); err != nil {
		return DNLRoutingState{}, err
	}
	node, found := QueryRoutingNode(state, msg.NodeID)
	if !found {
		return DNLRoutingState{}, fmt.Errorf("networking DNL node %s not found", msg.NodeID)
	}
	if node.OperatorAddress != msg.Authority {
		return DNLRoutingState{}, errors.New("networking DNL expire authority must match operator")
	}
	nodes := make([]NodeRecord, 0, len(state.Nodes))
	for _, candidate := range state.Nodes {
		if NormalizeNodeRecord(candidate).NodeID != node.NodeID {
			nodes = append(nodes, candidate)
		}
	}
	reputation := make([]ReputationCommitment, 0, len(state.Reputation))
	for _, commitment := range state.Reputation {
		if NormalizeReputationCommitment(commitment).NodeID != node.NodeID {
			reputation = append(reputation, commitment)
		}
	}
	tables := make([]RoutingTable, 0, len(state.Tables))
	for _, table := range state.Tables {
		routes := make([]DNLRoutingTableEntry, 0, len(table.Routes))
		for _, route := range table.Routes {
			if route.NextHopNodeID != node.NodeID {
				routes = append(routes, route)
			}
		}
		if len(routes) == 0 {
			continue
		}
		next, err := NewRoutingTable(RoutingTable{Epoch: table.Epoch, Routes: routes})
		if err != nil {
			return DNLRoutingState{}, err
		}
		tables = append(tables, next)
	}
	return BuildDNLRoutingState(nodes, reputation, state.Cache, tables, msg.Height)
}

func SubmitReputationCommitmentInRoutingState(state DNLRoutingState, msg MsgSubmitReputationCommitment) (DNLRoutingState, error) {
	if err := msg.Validate(); err != nil {
		return DNLRoutingState{}, err
	}
	node, found := QueryRoutingNode(state, msg.Commitment.NodeID)
	if !found {
		return DNLRoutingState{}, fmt.Errorf("networking DNL reputation node %s not found", msg.Commitment.NodeID)
	}
	if node.OperatorAddress != msg.Authority {
		return DNLRoutingState{}, errors.New("networking DNL reputation authority must match operator")
	}
	return BuildDNLRoutingState(state.Nodes, upsertDNLReputation(state.Reputation, msg.Commitment), state.Cache, state.Tables, msg.Height)
}

func UpdateRoutingTableInRoutingState(state DNLRoutingState, msg MsgUpdateRoutingTable) (DNLRoutingState, error) {
	if err := msg.Validate(); err != nil {
		return DNLRoutingState{}, err
	}
	for _, route := range msg.Table.Routes {
		node, found := QueryRoutingNode(state, route.NextHopNodeID)
		if !found {
			return DNLRoutingState{}, errors.New("networking DNL routing table route must reference registered node")
		}
		if !containsString(node.ZonesSupported, route.ZoneID) || !containsString(node.ServiceIDs, route.ServiceID) {
			return DNLRoutingState{}, errors.New("networking DNL routing table route must match node capabilities")
		}
	}
	tables := upsertDNLRoutingTable(state.Tables, msg.Table)
	return BuildDNLRoutingState(state.Nodes, state.Reputation, state.Cache, tables, msg.Height)
}

func QueryNodeRecordFromRoutingState(state DNLRoutingState, query QueryNodeRecord) (QueryNodeRecordResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryNodeRecordResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryNodeRecordResponse{}, err
	}
	record, found := QueryRoutingNode(state, query.NodeID)
	if !found {
		return QueryNodeRecordResponse{Found: false}, nil
	}
	response := QueryNodeRecordResponse{Record: record, Found: true}
	if query.IncludeProof {
		key, _ := RoutingNodeKey(record.NodeID)
		proof, err := QueryLookupProofFromRoutingState(state, QueryLookupProof{Key: key})
		if err != nil {
			return QueryNodeRecordResponse{}, err
		}
		response.Proof = proof.Proof
	}
	return response, nil
}

func QueryNodesByZoneFromRoutingState(state DNLRoutingState, query QueryNodesByZone) (QueryNodesByZoneResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryNodesByZoneResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryNodesByZoneResponse{}, err
	}
	nodes := QueryRoutingNodesByZone(state, query.ZoneID)
	proofs := make([]RoutingStateProof, 0, len(nodes))
	for _, node := range nodes {
		key, _ := RoutingZoneKey(query.ZoneID, node.NodeID)
		proof, err := QueryLookupProofFromRoutingState(state, QueryLookupProof{Key: key})
		if err == nil && proof.Found {
			proofs = append(proofs, proof.Proof)
		}
	}
	return QueryNodesByZoneResponse{Records: nodes, Proofs: proofs}, nil
}

func QueryNodesByServiceFromRoutingState(state DNLRoutingState, query QueryNodesByService) (QueryNodesByServiceResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryNodesByServiceResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryNodesByServiceResponse{}, err
	}
	nodes := QueryRoutingNodesByService(state, query.ServiceID)
	proofs := make([]RoutingStateProof, 0, len(nodes))
	for _, node := range nodes {
		key, _ := RoutingServiceKey(query.ServiceID, node.NodeID)
		proof, err := QueryLookupProofFromRoutingState(state, QueryLookupProof{Key: key})
		if err == nil && proof.Found {
			proofs = append(proofs, proof.Proof)
		}
	}
	return QueryNodesByServiceResponse{Records: nodes, Proofs: proofs}, nil
}

func QueryRoutingTableFromRoutingState(state DNLRoutingState, query QueryRoutingTable) (QueryRoutingTableResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryRoutingTableResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryRoutingTableResponse{}, err
	}
	table, found := LookupRoutingTable(state, query.Epoch)
	if !found {
		return QueryRoutingTableResponse{Found: false}, nil
	}
	response := QueryRoutingTableResponse{Table: table, Found: true}
	if query.IncludeProof {
		key, _ := RoutingTableKey(table.Epoch)
		proof, err := QueryLookupProofFromRoutingState(state, QueryLookupProof{Key: key})
		if err != nil {
			return QueryRoutingTableResponse{}, err
		}
		response.Proof = proof.Proof
	}
	return response, nil
}

func QueryLookupProofFromRoutingState(state DNLRoutingState, query QueryLookupProof) (QueryLookupProofResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryLookupProofResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryLookupProofResponse{}, err
	}
	entries, err := routingProofEntries(state)
	if err != nil {
		return QueryLookupProofResponse{}, err
	}
	valueHash := ""
	path := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Key == query.Key {
			valueHash = entry.Value
			continue
		}
		path = append(path, entry.EntryHash)
	}
	if valueHash == "" {
		return QueryLookupProofResponse{Found: false}, nil
	}
	sortStrings(path)
	proof := RoutingStateProof{Key: query.Key, ValueHash: valueHash, StateRoot: state.StateRoot, Height: state.Height, Path: path}
	proof.ProofHash = ComputeRoutingStateProofHash(proof)
	return QueryLookupProofResponse{Proof: proof, Found: true}, proof.Validate()
}

func QueryReputationCommitmentFromRoutingState(state DNLRoutingState, query QueryReputationCommitment) (QueryReputationCommitmentResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryReputationCommitmentResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryReputationCommitmentResponse{}, err
	}
	for _, commitment := range state.Reputation {
		commitment = NormalizeReputationCommitment(commitment)
		if commitment.NodeID != normalizeHashText(query.NodeID) {
			continue
		}
		response := QueryReputationCommitmentResponse{Commitment: commitment, Found: true}
		if query.IncludeProof {
			key, _ := RoutingReputationKey(commitment.NodeID)
			proof, err := QueryLookupProofFromRoutingState(state, QueryLookupProof{Key: key})
			if err != nil {
				return QueryReputationCommitmentResponse{}, err
			}
			response.Proof = proof.Proof
		}
		return response, nil
	}
	return QueryReputationCommitmentResponse{Found: false}, nil
}

func ExportDNLRoutingState(state DNLRoutingState) (DNLRoutingExport, error) {
	if err := state.Validate(); err != nil {
		return DNLRoutingExport{}, err
	}
	manifest := DNLRoutingExportManifest{
		Height:			state.Height,
		NodeCount:		uint64(len(state.Nodes)),
		ZoneIndexCount:		uint64(len(state.ZoneIndex)),
		ServiceIndexCount:	uint64(len(state.ServiceIndex)),
		ReputationCount:	uint64(len(state.Reputation)),
		CacheCount:		uint64(len(state.Cache)),
		TableCount:		uint64(len(state.Tables)),
		NodesRoot:		state.NodesRoot,
		ZonesRoot:		state.ZonesRoot,
		ServicesRoot:		state.ServicesRoot,
		ReputationRoot:		state.ReputationRoot,
		CacheRoot:		state.CacheRoot,
		TablesRoot:		state.TablesRoot,
		StateRoot:		state.StateRoot,
	}
	manifest.ManifestHash = ComputeDNLRoutingExportManifestHash(manifest)
	if err := manifest.Validate(); err != nil {
		return DNLRoutingExport{}, err
	}
	return DNLRoutingExport{State: state, Manifest: manifest}, nil
}

func ImportDNLRoutingState(export DNLRoutingExport) (DNLRoutingState, error) {
	if err := export.Manifest.Validate(); err != nil {
		return DNLRoutingState{}, err
	}
	state, err := BuildDNLRoutingState(export.State.Nodes, export.State.Reputation, export.State.Cache, export.State.Tables, export.Manifest.Height)
	if err != nil {
		return DNLRoutingState{}, err
	}
	if state.StateRoot != export.Manifest.StateRoot ||
		state.NodesRoot != export.Manifest.NodesRoot ||
		state.ZonesRoot != export.Manifest.ZonesRoot ||
		state.ServicesRoot != export.Manifest.ServicesRoot ||
		state.ReputationRoot != export.Manifest.ReputationRoot ||
		state.CacheRoot != export.Manifest.CacheRoot ||
		state.TablesRoot != export.Manifest.TablesRoot {
		return DNLRoutingState{}, errors.New("networking DNL routing import manifest root mismatch")
	}
	return state, nil
}

func (msg MsgRegisterNodeRecord) Validate() error {
	if err := validateRoutingAuthority(msg.Authority); err != nil {
		return err
	}
	record := NormalizeNodeRecord(msg.Record)
	if err := validateNodeRecordForMessage(record, msg.NetworkSalt, msg.Height); err != nil {
		return err
	}
	if err := msg.Reputation.Validate(); err != nil {
		return err
	}
	if msg.Reputation.NodeID != record.NodeID {
		return errors.New("networking DNL register reputation must reference node")
	}
	if msg.Height == 0 {
		return errors.New("networking DNL register height must be positive")
	}
	return validateOptionalRoutingHash("networking DNL register message hash", msg.MessageHash)
}

func (msg MsgUpdateNodeRecord) Validate() error {
	if err := validateRoutingAuthority(msg.Authority); err != nil {
		return err
	}
	record := NormalizeNodeRecord(msg.Record)
	if err := validateNodeRecordForMessage(record, msg.NetworkSalt, msg.Height); err != nil {
		return err
	}
	if err := msg.Reputation.Validate(); err != nil {
		return err
	}
	if msg.Reputation.NodeID != record.NodeID {
		return errors.New("networking DNL update reputation must reference node")
	}
	if msg.Height == 0 {
		return errors.New("networking DNL update height must be positive")
	}
	return validateOptionalRoutingHash("networking DNL update message hash", msg.MessageHash)
}

func (msg MsgExpireNodeRecord) Validate() error {
	if err := validateRoutingAuthority(msg.Authority); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL expire node id", normalizeHashText(msg.NodeID)); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL expire reason hash", normalizeHashText(msg.ReasonHash)); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("networking DNL expire height must be positive")
	}
	return validateOptionalRoutingHash("networking DNL expire message hash", msg.MessageHash)
}

func (msg MsgSubmitReputationCommitment) Validate() error {
	if err := validateRoutingAuthority(msg.Authority); err != nil {
		return err
	}
	if err := msg.Commitment.Validate(); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("networking DNL reputation height must be positive")
	}
	return validateOptionalRoutingHash("networking DNL reputation message hash", msg.MessageHash)
}

func (msg MsgUpdateRoutingTable) Validate() error {
	if err := validateRoutingAuthority(msg.Authority); err != nil {
		return err
	}
	if err := msg.Table.Validate(); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("networking DNL routing table height must be positive")
	}
	return validateOptionalRoutingHash("networking DNL routing table message hash", msg.MessageHash)
}

func (query QueryNodeRecord) Validate() error {
	return ValidateHash("networking DNL query node id", normalizeHashText(query.NodeID))
}

func (query QueryNodesByZone) Validate() error {
	return validateIdentifierSet("DNL query zone id", []string{query.ZoneID}, MaxZoneIDBytes)
}

func (query QueryNodesByService) Validate() error {
	return validateIdentifierSet("DNL query service id", []string{query.ServiceID}, MaxServiceIDBytes)
}

func (query QueryRoutingTable) Validate() error {
	if query.Epoch == 0 {
		return errors.New("networking DNL query routing table epoch must be positive")
	}
	return nil
}

func (query QueryLookupProof) Validate() error {
	if strings.TrimSpace(query.Key) == "" {
		return errors.New("networking DNL lookup proof key is required")
	}
	return nil
}

func (query QueryReputationCommitment) Validate() error {
	return ValidateHash("networking DNL query reputation node id", normalizeHashText(query.NodeID))
}

func (proof RoutingStateProof) Validate() error {
	if strings.TrimSpace(proof.Key) == "" {
		return errors.New("networking DNL routing proof key is required")
	}
	if err := ValidateHash("networking DNL routing proof value hash", proof.ValueHash); err != nil {
		return err
	}
	if err := ValidateHash("networking DNL routing proof state root", proof.StateRoot); err != nil {
		return err
	}
	if proof.Height == 0 {
		return errors.New("networking DNL routing proof height must be positive")
	}
	for _, item := range proof.Path {
		if err := ValidateHash("networking DNL routing proof path item", item); err != nil {
			return err
		}
	}
	if err := ValidateHash("networking DNL routing proof hash", proof.ProofHash); err != nil {
		return err
	}
	if proof.ProofHash != ComputeRoutingStateProofHash(proof) {
		return errors.New("networking DNL routing proof hash mismatch")
	}
	return nil
}

func (manifest DNLRoutingExportManifest) Validate() error {
	if manifest.Height == 0 {
		return errors.New("networking DNL routing export height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"nodes root", manifest.NodesRoot},
		{"zones root", manifest.ZonesRoot},
		{"services root", manifest.ServicesRoot},
		{"reputation root", manifest.ReputationRoot},
		{"cache root", manifest.CacheRoot},
		{"tables root", manifest.TablesRoot},
		{"state root", manifest.StateRoot},
		{"manifest hash", manifest.ManifestHash},
	} {
		if err := ValidateHash("networking DNL routing export "+item.name, item.value); err != nil {
			return err
		}
	}
	if manifest.ManifestHash != ComputeDNLRoutingExportManifestHash(manifest) {
		return errors.New("networking DNL routing export manifest hash mismatch")
	}
	return nil
}

func ComputeRoutingStateProofHash(proof RoutingStateProof) string {
	path := append([]string(nil), proof.Path...)
	sortStrings(path)
	parts := []string{"routing-state-proof", proof.Key, proof.ValueHash, proof.StateRoot, fmt.Sprintf("%d", proof.Height), fmt.Sprintf("%d", len(path))}
	parts = append(parts, path...)
	return HashParts(parts...)
}

func ComputeDNLRoutingExportManifestHash(manifest DNLRoutingExportManifest) string {
	return HashParts(
		"dnl-routing-export-manifest",
		fmt.Sprintf("%d", manifest.Height),
		fmt.Sprintf("%d", manifest.NodeCount),
		fmt.Sprintf("%d", manifest.ZoneIndexCount),
		fmt.Sprintf("%d", manifest.ServiceIndexCount),
		fmt.Sprintf("%d", manifest.ReputationCount),
		fmt.Sprintf("%d", manifest.CacheCount),
		fmt.Sprintf("%d", manifest.TableCount),
		manifest.NodesRoot,
		manifest.ZonesRoot,
		manifest.ServicesRoot,
		manifest.ReputationRoot,
		manifest.CacheRoot,
		manifest.TablesRoot,
		manifest.StateRoot,
	)
}

func routingProofEntries(state DNLRoutingState) ([]RoutingIndexRecord, error) {
	entries := make([]RoutingIndexRecord, 0, len(state.Nodes)+len(state.ZoneIndex)+len(state.ServiceIndex)+len(state.Reputation)+len(state.Cache)+len(state.Tables))
	for _, node := range state.Nodes {
		node = NormalizeNodeRecord(node)
		key, err := RoutingNodeKey(node.NodeID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newRoutingIndexRecord(key, ComputeNodeRecordCommitmentHash(node)))
	}
	entries = append(entries, state.ZoneIndex...)
	entries = append(entries, state.ServiceIndex...)
	for _, commitment := range state.Reputation {
		commitment = NormalizeReputationCommitment(commitment)
		key, err := RoutingReputationKey(commitment.NodeID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newRoutingIndexRecord(key, commitment.CommitmentHash))
	}
	for _, cache := range state.Cache {
		cache = NormalizeLookupCacheRecord(cache)
		key, err := RoutingCacheKey(cache.LookupKey)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newRoutingIndexRecord(key, cache.CacheHash))
	}
	for _, table := range state.Tables {
		key, err := RoutingTableKey(table.Epoch)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newRoutingIndexRecord(key, table.TableRoot))
	}
	entries = normalizeRoutingIndexRecords(entries)
	return entries, nil
}

func validateNodeRecordForMessage(record NodeRecord, networkSalt []byte, height uint64) error {
	if len(networkSalt) > 0 {
		return record.Validate(networkSalt, height)
	}
	return record.ValidateBasic()
}

func validateRoutingAuthority(authority string) error {
	if strings.TrimSpace(authority) != authority || authority == "" {
		return errors.New("networking DNL authority is required and must not have surrounding whitespace")
	}
	if len(authority) > MaxServiceIDBytes {
		return fmt.Errorf("networking DNL authority must be <= %d bytes", MaxServiceIDBytes)
	}
	return nil
}

func validateOptionalRoutingHash(field, value string) error {
	if value == "" {
		return nil
	}
	return ValidateHash(field, normalizeHashText(value))
}

func routingNodeIndex(nodes []NodeRecord, nodeID string) (int, NodeRecord, bool) {
	nodeID = normalizeHashText(nodeID)
	for i, node := range nodes {
		normalized := NormalizeNodeRecord(node)
		if normalized.NodeID == nodeID {
			return i, normalized, true
		}
	}
	return 0, NodeRecord{}, false
}

func upsertDNLReputation(records []ReputationCommitment, commitment ReputationCommitment) []ReputationCommitment {
	out := append([]ReputationCommitment(nil), records...)
	commitment = NormalizeReputationCommitment(commitment)
	for i, candidate := range out {
		if NormalizeReputationCommitment(candidate).NodeID == commitment.NodeID {
			out[i] = commitment
			return normalizeReputationCommitments(out)
		}
	}
	out = append(out, commitment)
	return normalizeReputationCommitments(out)
}

func upsertDNLRoutingTable(tables []RoutingTable, table RoutingTable) []RoutingTable {
	out := append([]RoutingTable(nil), tables...)
	table = NormalizeRoutingTable(table)
	for i, candidate := range out {
		if candidate.Epoch == table.Epoch {
			out[i] = table
			return normalizeRoutingTables(out)
		}
	}
	out = append(out, table)
	return normalizeRoutingTables(out)
}
