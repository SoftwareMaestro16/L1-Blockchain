package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

const (
	ModuleName		= "proofregistry"
	DefaultHistoryWindow	= uint64(4096)
)

type ProofRegistryQueryKind string
type ProofFailureCategory string

const (
	QueryKindZone			ProofRegistryQueryKind	= "zone"
	QueryKindShard			ProofRegistryQueryKind	= "shard"
	QueryKindStore			ProofRegistryQueryKind	= "store"
	QueryKindMessageInclusion	ProofRegistryQueryKind	= "message_inclusion"
	QueryKindReceipt		ProofRegistryQueryKind	= "receipt"
	QueryKindIdentity		ProofRegistryQueryKind	= "identity"
	QueryKindContractState		ProofRegistryQueryKind	= "contract_state"
	QueryKindPaymentSettlement	ProofRegistryQueryKind	= "payment_settlement"

	FailureCategoryTrust		ProofFailureCategory	= "trust"
	FailureCategoryScope		ProofFailureCategory	= "scope"
	FailureCategoryRoot		ProofFailureCategory	= "root"
	FailureCategoryObject		ProofFailureCategory	= "object"
	FailureCategoryNonExistence	ProofFailureCategory	= "non_existence"
)

type ProofFailureCodeDescriptor struct {
	Code		coretypes.UniversalProofFailureCode
	Meaning		string
	Category	ProofFailureCategory
}

type ProofRootMetadata struct {
	RootType	coretypes.RootType
	Source		string
	Description	string
	ExpiresAtHeight	uint64
	MetadataHash	string
}

type ProofRegistrySnapshot struct {
	Height			uint64
	AppHash			string
	AetraCoreRoot		string
	GlobalZoneRoot		string
	GlobalMessageRoot	string
	ReceiptRoot		string
	Roots			[]coretypes.ProofRoot
	Metadata		[]ProofRootMetadata
	SnapshotHash		string
}

type ProofRegistryEntry struct {
	Kind		ProofRegistryQueryKind
	Height		uint64
	ProofType	coretypes.UniversalProofType
	RootType	coretypes.RootType
	ZoneID		coretypes.ZoneID
	ShardID		coretypes.ShardID
	Key		[]byte
	MessageID	string
	Envelope	coretypes.UniversalProofEnvelope
	EntryHash	string
}

type ProofRegistryState struct {
	HistoryWindow	uint64
	Snapshots	[]ProofRegistrySnapshot
	Entries		[]ProofRegistryEntry
	TestVectors	[]ProofTestVector
}

type ProofRegistryQuery struct {
	Kind		ProofRegistryQueryKind
	Height		uint64
	ProofType	coretypes.UniversalProofType
	RootType	coretypes.RootType
	ZoneID		coretypes.ZoneID
	ShardID		coretypes.ShardID
	Key		[]byte
	MessageID	string
}

type ProofRegistryResponse struct {
	Found		bool
	Snapshot	ProofRegistrySnapshot
	Root		coretypes.ProofRoot
	Envelope	coretypes.UniversalProofEnvelope
	FailureCode	coretypes.UniversalProofFailureCode
	ResponseHash	string
}

type StoreV2ProofAdapterRequest struct {
	ProofType		coretypes.UniversalProofType
	ChainID			string
	Height			uint64
	AppHash			string
	RootType		coretypes.RootType
	ZoneID			coretypes.ZoneID
	ShardID			coretypes.ShardID
	Key			[]byte
	Value			[]byte
	NonExistenceMarker	[]byte
	ObjectExpiryHeight	uint64
	StoreRoot		string
	ProofOps		[]string
	VerificationPath	[]coretypes.UniversalRootStep
	ZoneCommitment		coretypes.ZoneCommitment
	ShardCommitment		coretypes.UniversalShardCommitment
	MessageCommitment	coretypes.UniversalMessageCommitment
}

type ProofTestVector struct {
	VectorID		string
	Name			string
	Positive		bool
	Proof			coretypes.UniversalProofEnvelope
	TrustedHeader		coretypes.UniversalTrustedHeader
	ExpectedFailureCode	coretypes.UniversalProofFailureCode
	VectorHash		string
}

func NewProofRegistryState(historyWindow uint64) (ProofRegistryState, error) {
	if historyWindow == 0 {
		historyWindow = DefaultHistoryWindow
	}
	state := ProofRegistryState{HistoryWindow: historyWindow}
	return state, state.Validate()
}

func SupportedProofFailureCodes() []ProofFailureCodeDescriptor {
	return []ProofFailureCodeDescriptor{
		{Code: coretypes.ProofFailureUntrustedHeader, Meaning: "header at proof height is missing, untrusted, or fails light-client verification", Category: FailureCategoryTrust},
		{Code: coretypes.ProofFailureChainIDMismatch, Meaning: "proof chain ID does not match verifier expectations", Category: FailureCategoryTrust},
		{Code: coretypes.ProofFailureHeightUnavailable, Meaning: "proof height is outside available root history or trusted header range", Category: FailureCategoryTrust},
		{Code: coretypes.ProofFailureRootMismatch, Meaning: "a root in the verification path does not match the committed parent root", Category: FailureCategoryRoot},
		{Code: coretypes.ProofFailureZoneNotFound, Meaning: "zone commitment is missing or zone ID is not registered at proof height", Category: FailureCategoryScope},
		{Code: coretypes.ProofFailureShardNotFound, Meaning: "shard root is missing or shard ID is not active at layout epoch", Category: FailureCategoryScope},
		{Code: coretypes.ProofFailureStoreProofInvalid, Meaning: "Store v2 proof fails key/value verification", Category: FailureCategoryObject},
		{Code: coretypes.ProofFailureMessageNotIncluded, Meaning: "message is absent from the claimed inbox, outbox, or global message root", Category: FailureCategoryObject},
		{Code: coretypes.ProofFailureReceiptNotFound, Meaning: "receipt is absent from the claimed receipt root", Category: FailureCategoryObject},
		{Code: coretypes.ProofFailureObjectExpired, Meaning: "object existed but was expired or invalid at the proof height", Category: FailureCategoryObject},
		{Code: coretypes.ProofFailureNonExistenceProofInvalid, Meaning: "non-existence boundary or range proof is invalid", Category: FailureCategoryNonExistence},
	}
}

func NewProofRootMetadata(metadata ProofRootMetadata) (ProofRootMetadata, error) {
	metadata.MetadataHash = ComputeProofRootMetadataHash(metadata)
	return metadata, metadata.Validate()
}

func NewProofRegistrySnapshot(snapshot ProofRegistrySnapshot) (ProofRegistrySnapshot, error) {
	snapshot.Roots = append([]coretypes.ProofRoot(nil), snapshot.Roots...)
	sortProofRegistryRoots(snapshot.Roots)
	snapshot.Metadata = append([]ProofRootMetadata(nil), snapshot.Metadata...)
	sortProofRootMetadata(snapshot.Metadata)
	snapshot.SnapshotHash = ComputeProofRegistrySnapshotHash(snapshot)
	return snapshot, snapshot.Validate()
}

func CommitProofRegistrySnapshot(state ProofRegistryState, snapshot ProofRegistrySnapshot) (ProofRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ProofRegistryState{}, err
	}
	if err := snapshot.Validate(); err != nil {
		return ProofRegistryState{}, err
	}
	for _, existing := range state.Snapshots {
		if existing.Height == snapshot.Height {
			return ProofRegistryState{}, fmt.Errorf("proofregistry snapshot height %d already exists", snapshot.Height)
		}
	}
	next := state.Clone()
	next.Snapshots = append(next.Snapshots, snapshot)
	sortProofRegistrySnapshots(next.Snapshots)
	next.pruneHistory()
	return next, next.Validate()
}

func NewProofRegistryEntry(entry ProofRegistryEntry) (ProofRegistryEntry, error) {
	entry.Key = append([]byte(nil), entry.Key...)
	if entry.Height == 0 {
		entry.Height = entry.Envelope.Height
	}
	if entry.ProofType == "" {
		entry.ProofType = entry.Envelope.ProofType
	}
	if entry.RootType == "" {
		entry.RootType = entry.Envelope.RootType
	}
	if entry.ZoneID == "" {
		entry.ZoneID = entry.Envelope.ZoneID
	}
	if entry.ShardID == "" {
		entry.ShardID = entry.Envelope.ShardID
	}
	if len(entry.Key) == 0 {
		entry.Key = append([]byte(nil), entry.Envelope.Key...)
	}
	if entry.MessageID == "" && entry.Envelope.HasMessageCommit {
		entry.MessageID = entry.Envelope.MessageCommit.MessageID
	}
	entry.EntryHash = ComputeProofRegistryEntryHash(entry)
	return entry, entry.Validate()
}

func AddProofRegistryEntry(state ProofRegistryState, entry ProofRegistryEntry) (ProofRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ProofRegistryState{}, err
	}
	if err := entry.Validate(); err != nil {
		return ProofRegistryState{}, err
	}
	if _, found := state.SnapshotAtHeight(entry.Height); !found {
		return ProofRegistryState{}, fmt.Errorf("proofregistry missing root snapshot at height %d", entry.Height)
	}
	for _, existing := range state.Entries {
		if existing.EntryHash == entry.EntryHash {
			return ProofRegistryState{}, errors.New("proofregistry duplicate proof entry")
		}
	}
	next := state.Clone()
	next.Entries = append(next.Entries, entry)
	sortProofRegistryEntries(next.Entries)
	return next, next.Validate()
}

func BuildStoreV2ProofEnvelope(req StoreV2ProofAdapterRequest) (coretypes.UniversalProofEnvelope, error) {
	storeProof, err := coretypes.NewUniversalStoreProof(coretypes.UniversalStoreProof{
		ProofVersion:		coretypes.UniversalProofVersionV1,
		Key:			req.Key,
		Value:			req.Value,
		NonExistenceMarker:	req.NonExistenceMarker,
		StoreRoot:		req.StoreRoot,
		ProofOps:		req.ProofOps,
	})
	if err != nil {
		return coretypes.UniversalProofEnvelope{}, err
	}
	proof := coretypes.UniversalProofEnvelope{
		ProofType:		req.ProofType,
		ProofVersion:		coretypes.UniversalProofVersionV1,
		ChainID:		req.ChainID,
		Height:			req.Height,
		AppHash:		req.AppHash,
		RootType:		req.RootType,
		ZoneID:			req.ZoneID,
		ShardID:		req.ShardID,
		Key:			req.Key,
		Value:			req.Value,
		AbsenceMarker:		req.NonExistenceMarker,
		ObjectExpiryHeight:	req.ObjectExpiryHeight,
		StoreProof:		storeProof,
		VerificationPath:	append([]coretypes.UniversalRootStep(nil), req.VerificationPath...),
	}
	if req.ZoneCommitment.CommitmentHash != "" {
		proof.ZoneCommitment = req.ZoneCommitment
		proof.HasZoneCommit = true
	}
	if req.ShardCommitment.CommitmentHash != "" {
		proof.ShardCommitment = req.ShardCommitment
		proof.HasShardCommit = true
	}
	if req.MessageCommitment.DeliveryCommitmentHash != "" {
		proof.MessageCommit = req.MessageCommitment
		proof.HasMessageCommit = true
	}
	return coretypes.NewUniversalProofEnvelope(proof)
}

func QueryProof(state ProofRegistryState, query ProofRegistryQuery) (ProofRegistryResponse, error) {
	if err := state.Validate(); err != nil {
		return ProofRegistryResponse{}, err
	}
	if err := query.Validate(); err != nil {
		return ProofRegistryResponse{}, err
	}
	snapshot, found := state.SnapshotAtHeight(query.Height)
	if !found {
		return ProofRegistryResponse{Found: false, FailureCode: coretypes.ProofFailureHeightUnavailable}, nil
	}
	for _, entry := range state.Entries {
		if !entry.matches(query) {
			continue
		}
		root, _ := snapshot.RootByType(entry.RootType, entry.ZoneID)
		resp := ProofRegistryResponse{Found: true, Snapshot: snapshot, Root: root, Envelope: entry.Envelope}
		resp.ResponseHash = ComputeProofRegistryResponseHash(resp)
		return resp, resp.Validate()
	}
	code := missingFailureForQueryKind(query.Kind)
	return ProofRegistryResponse{Found: false, Snapshot: snapshot, FailureCode: code}, nil
}

func QueryZoneProof(state ProofRegistryState, height uint64, zoneID coretypes.ZoneID) (ProofRegistryResponse, error) {
	return QueryProof(state, ProofRegistryQuery{Kind: QueryKindZone, Height: height, ProofType: coretypes.ProofTypeZoneRoot, RootType: coretypes.ZoneStateProofRootType, ZoneID: zoneID})
}

func QueryShardProof(state ProofRegistryState, height uint64, zoneID coretypes.ZoneID, shardID coretypes.ShardID) (ProofRegistryResponse, error) {
	return QueryProof(state, ProofRegistryQuery{Kind: QueryKindShard, Height: height, ProofType: coretypes.ProofTypeShardRoot, RootType: coretypes.ShardStateProofRootType, ZoneID: zoneID, ShardID: shardID})
}

func QueryMessageInclusionProof(state ProofRegistryState, height uint64, messageID string) (ProofRegistryResponse, error) {
	return QueryProof(state, ProofRegistryQuery{Kind: QueryKindMessageInclusion, Height: height, ProofType: coretypes.ProofTypeMessageInclusion, RootType: coretypes.MessageProofRootType, MessageID: messageID})
}

func QueryReceiptProof(state ProofRegistryState, height uint64, messageID string) (ProofRegistryResponse, error) {
	return QueryProof(state, ProofRegistryQuery{Kind: QueryKindReceipt, Height: height, ProofType: coretypes.ProofTypeMessageReceipt, RootType: coretypes.ReceiptProofRootType, MessageID: messageID})
}

func QueryIdentityProof(state ProofRegistryState, height uint64, key []byte) (ProofRegistryResponse, error) {
	return QueryProof(state, ProofRegistryQuery{Kind: QueryKindIdentity, Height: height, ProofType: coretypes.ProofTypeDomainOwnership, RootType: coretypes.DomainOwnershipProofRootType, ZoneID: coretypes.ZoneIDIdentity, Key: key})
}

func QueryContractStateProof(state ProofRegistryState, height uint64, key []byte) (ProofRegistryResponse, error) {
	return QueryProof(state, ProofRegistryQuery{Kind: QueryKindContractState, Height: height, ProofType: coretypes.ProofTypeContractState, RootType: coretypes.ContractStateProofRootType, ZoneID: coretypes.ZoneIDContract, Key: key})
}

func QueryPaymentSettlementProof(state ProofRegistryState, height uint64, key []byte) (ProofRegistryResponse, error) {
	return QueryProof(state, ProofRegistryQuery{Kind: QueryKindPaymentSettlement, Height: height, ProofType: coretypes.ProofTypePaymentSettlement, RootType: coretypes.PaymentSettlementProofRootType, ZoneID: coretypes.ZoneIDFinancial, Key: key})
}

func NewProofTestVector(vector ProofTestVector) (ProofTestVector, error) {
	if vector.VectorID == "" {
		vector.VectorID = hashParts("proofregistry-vector", vector.Name, string(vector.ExpectedFailureCode), vector.Proof.ProofHash)
	}
	vector.VectorHash = ComputeProofTestVectorHash(vector)
	return vector, vector.Validate()
}

func AddProofTestVector(state ProofRegistryState, vector ProofTestVector) (ProofRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ProofRegistryState{}, err
	}
	if err := vector.Validate(); err != nil {
		return ProofRegistryState{}, err
	}
	next := state.Clone()
	next.TestVectors = append(next.TestVectors, vector)
	sortProofTestVectors(next.TestVectors)
	return next, next.Validate()
}

func (m ProofRootMetadata) Validate() error {
	if err := coretypes.ValidateHash("proofregistry metadata hash", m.MetadataHash); err != nil {
		return err
	}
	if m.MetadataHash != ComputeProofRootMetadataHash(m) {
		return errors.New("proofregistry metadata hash mismatch")
	}
	if string(m.RootType) == "" {
		return errors.New("proofregistry metadata root type is required")
	}
	if m.Source == "" {
		return errors.New("proofregistry metadata source is required")
	}
	return nil
}

func (s ProofRegistrySnapshot) Validate() error {
	if s.Height == 0 {
		return errors.New("proofregistry snapshot height must be positive")
	}
	for _, field := range []struct {
		name	string
		value	string
	}{
		{"proofregistry app hash", s.AppHash},
		{"proofregistry aether core root", s.AetraCoreRoot},
		{"proofregistry global zone root", s.GlobalZoneRoot},
		{"proofregistry global message root", s.GlobalMessageRoot},
		{"proofregistry receipt root", s.ReceiptRoot},
		{"proofregistry snapshot hash", s.SnapshotHash},
	} {
		if err := coretypes.ValidateHash(field.name, field.value); err != nil {
			return err
		}
	}
	if len(s.Roots) == 0 {
		return errors.New("proofregistry snapshot requires proof roots")
	}
	if err := validateProofRegistryRoots(s.Height, s.Roots); err != nil {
		return err
	}
	for _, metadata := range s.Metadata {
		if err := metadata.Validate(); err != nil {
			return err
		}
	}
	if s.SnapshotHash != ComputeProofRegistrySnapshotHash(s) {
		return errors.New("proofregistry snapshot hash mismatch")
	}
	return nil
}

func (s ProofRegistryState) Validate() error {
	if s.HistoryWindow == 0 {
		return errors.New("proofregistry history window must be positive")
	}
	for _, snapshot := range s.Snapshots {
		if err := snapshot.Validate(); err != nil {
			return err
		}
	}
	for _, entry := range s.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}
	for _, vector := range s.TestVectors {
		if err := vector.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (e ProofRegistryEntry) Validate() error {
	if e.Kind == "" {
		return errors.New("proofregistry query kind is required")
	}
	if e.Height == 0 {
		return errors.New("proofregistry entry height must be positive")
	}
	if err := e.Envelope.ValidateFormat(); err != nil {
		return err
	}
	if e.Envelope.Height != e.Height || e.Envelope.ProofType != e.ProofType || e.Envelope.RootType != e.RootType {
		return errors.New("proofregistry entry envelope scope mismatch")
	}
	if string(e.RootType) == "" {
		return errors.New("proofregistry entry root type is required")
	}
	if len(e.Key) == 0 {
		return errors.New("proofregistry entry key is required")
	}
	if err := coretypes.ValidateHash("proofregistry entry hash", e.EntryHash); err != nil {
		return err
	}
	if e.EntryHash != ComputeProofRegistryEntryHash(e) {
		return errors.New("proofregistry entry hash mismatch")
	}
	return nil
}

func (q ProofRegistryQuery) Validate() error {
	if q.Kind == "" {
		return errors.New("proofregistry query kind is required")
	}
	if q.Height == 0 {
		return errors.New("proofregistry query height must be positive")
	}
	if q.ProofType == "" {
		return errors.New("proofregistry query proof type is required")
	}
	if q.RootType == "" {
		return errors.New("proofregistry query root type is required")
	}
	if q.Kind == QueryKindMessageInclusion || q.Kind == QueryKindReceipt {
		return coretypes.ValidateHash("proofregistry query message id", q.MessageID)
	}
	if q.Kind != QueryKindZone && q.Kind != QueryKindShard && q.Kind != QueryKindMessageInclusion && q.Kind != QueryKindReceipt && len(q.Key) == 0 {
		return errors.New("proofregistry query key is required")
	}
	return nil
}

func (r ProofRegistryResponse) Validate() error {
	if !r.Found {
		if r.FailureCode == "" {
			return errors.New("proofregistry missing response requires failure code")
		}
		return nil
	}
	if err := r.Snapshot.Validate(); err != nil {
		return err
	}
	if err := r.Envelope.ValidateFormat(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("proofregistry response hash", r.ResponseHash); err != nil {
		return err
	}
	if r.ResponseHash != ComputeProofRegistryResponseHash(r) {
		return errors.New("proofregistry response hash mismatch")
	}
	return nil
}

func (v ProofTestVector) Validate() error {
	if v.VectorID == "" || v.Name == "" {
		return errors.New("proofregistry test vector id and name are required")
	}
	if err := v.Proof.ValidateFormat(); err != nil {
		return err
	}
	if err := v.TrustedHeader.Validate(); err != nil {
		return err
	}
	result := coretypes.VerifyUniversalProof(v.Proof, v.TrustedHeader)
	if v.Positive && !result.Verified {
		return fmt.Errorf("proofregistry positive vector failed with %s", result.FailureCode)
	}
	if !v.Positive && result.FailureCode != v.ExpectedFailureCode {
		return fmt.Errorf("proofregistry negative vector expected %s got %s", v.ExpectedFailureCode, result.FailureCode)
	}
	if err := coretypes.ValidateHash("proofregistry vector hash", v.VectorHash); err != nil {
		return err
	}
	if v.VectorHash != ComputeProofTestVectorHash(v) {
		return errors.New("proofregistry vector hash mismatch")
	}
	return nil
}

func (s ProofRegistryState) Clone() ProofRegistryState {
	out := ProofRegistryState{HistoryWindow: s.HistoryWindow}
	out.Snapshots = append([]ProofRegistrySnapshot(nil), s.Snapshots...)
	out.Entries = append([]ProofRegistryEntry(nil), s.Entries...)
	out.TestVectors = append([]ProofTestVector(nil), s.TestVectors...)
	return out
}

func (s ProofRegistryState) SnapshotAtHeight(height uint64) (ProofRegistrySnapshot, bool) {
	for _, snapshot := range s.Snapshots {
		if snapshot.Height == height {
			return snapshot, true
		}
	}
	return ProofRegistrySnapshot{}, false
}

func (s ProofRegistrySnapshot) RootByType(rootType coretypes.RootType, zoneID coretypes.ZoneID) (coretypes.ProofRoot, bool) {
	for _, root := range s.Roots {
		if root.RootType == rootType && root.ZoneID == zoneID {
			return root, true
		}
	}
	return coretypes.ProofRoot{}, false
}

func (s *ProofRegistryState) pruneHistory() {
	if s.HistoryWindow == 0 || uint64(len(s.Snapshots)) <= s.HistoryWindow {
		return
	}
	dropBefore := len(s.Snapshots) - int(s.HistoryWindow)
	minHeight := s.Snapshots[dropBefore].Height
	s.Snapshots = append([]ProofRegistrySnapshot(nil), s.Snapshots[dropBefore:]...)
	entries := make([]ProofRegistryEntry, 0, len(s.Entries))
	for _, entry := range s.Entries {
		if entry.Height >= minHeight {
			entries = append(entries, entry)
		}
	}
	s.Entries = entries
}

func (e ProofRegistryEntry) matches(q ProofRegistryQuery) bool {
	if e.Kind != q.Kind || e.Height != q.Height || e.ProofType != q.ProofType || e.RootType != q.RootType {
		return false
	}
	if q.ZoneID != "" && e.ZoneID != q.ZoneID {
		return false
	}
	if q.ShardID != "" && e.ShardID != q.ShardID {
		return false
	}
	if q.MessageID != "" && e.MessageID != q.MessageID {
		return false
	}
	if len(q.Key) != 0 && string(e.Key) != string(q.Key) {
		return false
	}
	return true
}

func missingFailureForQueryKind(kind ProofRegistryQueryKind) coretypes.UniversalProofFailureCode {
	switch kind {
	case QueryKindZone:
		return coretypes.ProofFailureZoneNotFound
	case QueryKindShard:
		return coretypes.ProofFailureShardNotFound
	case QueryKindMessageInclusion:
		return coretypes.ProofFailureMessageNotIncluded
	case QueryKindReceipt:
		return coretypes.ProofFailureReceiptNotFound
	default:
		return coretypes.ProofFailureStoreProofInvalid
	}
}

func ComputeProofRootMetadataHash(metadata ProofRootMetadata) string {
	return hashParts("aetra-next-proofregistry-metadata-v1", string(metadata.RootType), metadata.Source, metadata.Description, fmt.Sprint(metadata.ExpiresAtHeight))
}

func ComputeProofRegistrySnapshotHash(snapshot ProofRegistrySnapshot) string {
	parts := []string{"aetra-next-proofregistry-snapshot-v1", fmt.Sprint(snapshot.Height), snapshot.AppHash, snapshot.AetraCoreRoot, snapshot.GlobalZoneRoot, snapshot.GlobalMessageRoot, snapshot.ReceiptRoot}
	for _, root := range snapshot.Roots {
		parts = append(parts, string(root.RootType), string(root.ZoneID), root.RootHash, root.Source)
	}
	for _, metadata := range snapshot.Metadata {
		parts = append(parts, metadata.MetadataHash)
	}
	return hashParts(parts...)
}

func ComputeProofRegistryEntryHash(entry ProofRegistryEntry) string {
	return hashParts(
		"aetra-next-proofregistry-entry-v1",
		string(entry.Kind),
		fmt.Sprint(entry.Height),
		string(entry.ProofType),
		string(entry.RootType),
		string(entry.ZoneID),
		string(entry.ShardID),
		string(entry.Key),
		entry.MessageID,
		entry.Envelope.ProofHash,
	)
}

func ComputeProofRegistryResponseHash(resp ProofRegistryResponse) string {
	return hashParts(
		"aetra-next-proofregistry-response-v1",
		fmt.Sprint(resp.Found),
		resp.Snapshot.SnapshotHash,
		resp.Root.RootHash,
		resp.Envelope.ProofHash,
		string(resp.FailureCode),
	)
}

func ComputeProofTestVectorHash(vector ProofTestVector) string {
	return hashParts(
		"aetra-next-proofregistry-vector-v1",
		vector.VectorID,
		vector.Name,
		fmt.Sprint(vector.Positive),
		vector.Proof.ProofHash,
		vector.TrustedHeader.AppHash,
		string(vector.ExpectedFailureCode),
	)
}

func validateProofRegistryRoots(height uint64, roots []coretypes.ProofRoot) error {
	seen := make(map[string]struct{}, len(roots))
	for _, root := range roots {
		if root.Height != height {
			return errors.New("proofregistry root height mismatch")
		}
		if err := root.Validate(); err != nil {
			return err
		}
		key := string(root.RootType) + "/" + string(root.ZoneID) + "/" + root.Source
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate proofregistry root %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func sortProofRegistrySnapshots(snapshots []ProofRegistrySnapshot) {
	sort.SliceStable(snapshots, func(i, j int) bool {
		return snapshots[i].Height < snapshots[j].Height
	})
}

func sortProofRegistryRoots(roots []coretypes.ProofRoot) {
	sort.SliceStable(roots, func(i, j int) bool {
		if roots[i].RootType == roots[j].RootType {
			if roots[i].ZoneID == roots[j].ZoneID {
				return roots[i].Source < roots[j].Source
			}
			return roots[i].ZoneID < roots[j].ZoneID
		}
		return roots[i].RootType < roots[j].RootType
	})
}

func sortProofRootMetadata(metadata []ProofRootMetadata) {
	sort.SliceStable(metadata, func(i, j int) bool {
		if metadata[i].RootType == metadata[j].RootType {
			return metadata[i].Source < metadata[j].Source
		}
		return metadata[i].RootType < metadata[j].RootType
	})
}

func sortProofRegistryEntries(entries []ProofRegistryEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].EntryHash < entries[j].EntryHash
	})
}

func sortProofTestVectors(vectors []ProofTestVector) {
	sort.SliceStable(vectors, func(i, j int) bool {
		return vectors[i].VectorID < vectors[j].VectorID
	})
}

func hashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		writePart(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func writePart(w interface{ Write([]byte) (int, error) }, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}
