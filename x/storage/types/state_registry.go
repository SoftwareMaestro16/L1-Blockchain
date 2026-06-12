package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	StorageObjectStatePrefix	= "storage/objects"
	StorageContentIndexPrefix	= "storage/content"
	StorageChunkStatePrefix		= "storage/chunks"
	StorageOwnerIndexPrefix		= "storage/owner_index"
	StorageAccessPrefix		= "storage/access"
	StorageReplicationPrefix	= "storage/replication"
	StorageRootPrefix		= "storage/root"
)

type StorageIndexEntry struct {
	Key		string
	Value		string
	EntryHash	string
}

type StorageChunkDescriptorRecord struct {
	ObjectID	string
	Descriptor	StorageChunkDescriptor
	RecordHash	string
}

type ReplicationStatusCommitment struct {
	ObjectID		string
	ReplicationPolicy	string
	StorageClass		string
	ReplicaCount		uint32
	AvailabilityBps		uint32
	LastVerifiedHeight	uint64
	CommitmentHash		string
}

type StorageRoot struct {
	Height			uint64
	ObjectRoot		string
	ContentIndexRoot	string
	ChunkRoot		string
	OwnerIndexRoot		string
	AccessRoot		string
	ReplicationRoot		string
	StateRoot		string
}

type StorageStateV2 struct {
	Objects		[]StorageObject
	Chunks		[]StorageChunkDescriptorRecord
	ContentIndex	[]StorageIndexEntry
	OwnerIndex	[]StorageIndexEntry
	AccessReceipts	[]StorageAccessReceipt
	Replications	[]ReplicationStatusCommitment
	Height		uint64
	Root		StorageRoot
}

type MsgRegisterStorageObject struct {
	Authority	string
	Object		StorageObject
	Chunks		[]StorageChunkDescriptor
	Replication	ReplicationStatusCommitment
	Height		uint64
	MessageHash	string
}

type MsgUpdateStoragePolicy struct {
	Authority		string
	ObjectID		string
	ReplicationPolicy	string
	AccessPolicy		string
	StorageClass		string
	Replication		ReplicationStatusCommitment
	Height			uint64
	MessageHash		string
}

type MsgRenewStorageObject struct {
	Authority	string
	ObjectID	string
	ExpiresHeight	uint64
	Height		uint64
	MessageHash	string
}

type MsgDeleteStorageObject struct {
	Authority	string
	ObjectID	string
	Height		uint64
	MessageHash	string
}

type MsgSubmitStorageReceipt struct {
	Authority	string
	Receipt		StorageAccessReceipt
	Height		uint64
	MessageHash	string
}

type MsgVerifyStorageProof struct {
	Authority	string
	ObjectID	string
	Proof		StorageChunkInclusionProof
	Height		uint64
	MessageHash	string
}

type StorageStateProof struct {
	Key		string
	ValueHash	string
	Root		string
	Height		uint64
	ProofHashes	[]string
	ProofHash	string
}

func StorageObjectKey(objectID string) (string, error) {
	if err := validateStorageToken("storage object key object id", objectID); err != nil {
		return "", err
	}
	return StorageObjectStatePrefix + "/" + objectID, nil
}

func StorageContentIndexKey(contentHash string) (string, error) {
	if err := validateStorageHash("storage content index content hash", contentHash); err != nil {
		return "", err
	}
	return StorageContentIndexPrefix + "/" + contentHash, nil
}

func StorageChunkDescriptorKey(objectID string, chunkIndex uint32) (string, error) {
	if err := validateStorageToken("storage chunk key object id", objectID); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%020d", StorageChunkStatePrefix, objectID, chunkIndex), nil
}

func StorageOwnerIndexKey(owner, objectID string) (string, error) {
	if err := validateStorageToken("storage owner index owner", owner); err != nil {
		return "", err
	}
	if err := validateStorageToken("storage owner index object id", objectID); err != nil {
		return "", err
	}
	return StorageOwnerIndexPrefix + "/" + owner + "/" + objectID, nil
}

func StorageAccessReceiptKey(objectID, accessID string) (string, error) {
	if err := validateStorageToken("storage access key object id", objectID); err != nil {
		return "", err
	}
	if err := validateStorageToken("storage access key access id", accessID); err != nil {
		return "", err
	}
	return StorageAccessPrefix + "/" + objectID + "/" + accessID, nil
}

func StorageReplicationKey(objectID string) (string, error) {
	if err := validateStorageToken("storage replication key object id", objectID); err != nil {
		return "", err
	}
	return StorageReplicationPrefix + "/" + objectID, nil
}

func StorageRootKey(height uint64) (string, error) {
	if height == 0 {
		return "", errors.New("storage root key height must be positive")
	}
	return fmt.Sprintf("%s/%020d", StorageRootPrefix, height), nil
}

func NewReplicationStatusCommitment(commitment ReplicationStatusCommitment) (ReplicationStatusCommitment, error) {
	if commitment.CommitmentHash != "" {
		return ReplicationStatusCommitment{}, errors.New("storage replication commitment hash must be empty before construction")
	}
	if err := commitment.ValidateFormat(); err != nil {
		return ReplicationStatusCommitment{}, err
	}
	commitment.CommitmentHash = ComputeStorageReplicationStatusCommitmentHash(commitment)
	return commitment, commitment.Validate()
}

func BuildStorageStateV2(objects []StorageObject, chunks []StorageChunkDescriptorRecord, receipts []StorageAccessReceipt, replications []ReplicationStatusCommitment, height uint64) (StorageStateV2, error) {
	state := StorageStateV2{
		Objects:	normalizeStorageObjects(objects),
		Chunks:		normalizeStorageChunkDescriptorRecords(chunks),
		AccessReceipts:	normalizeStorageReceipts(receipts),
		Replications:	normalizeStorageReplications(replications),
		Height:		height,
	}
	if err := state.populateIndexes(); err != nil {
		return StorageStateV2{}, err
	}
	if err := state.ValidateFormat(); err != nil {
		return StorageStateV2{}, err
	}
	state.Root = ComputeStorageRootV2(state)
	return state, state.Validate()
}

func RegisterStorageObjectInStateV2(state StorageStateV2, msg MsgRegisterStorageObject, params StorageChunkParams) (StorageStateV2, error) {
	if err := msg.Validate(params); err != nil {
		return StorageStateV2{}, err
	}
	if msg.Authority != msg.Object.Owner {
		return StorageStateV2{}, errors.New("storage register authority must own object")
	}
	records := make([]StorageChunkDescriptorRecord, 0, len(msg.Chunks))
	for _, descriptor := range msg.Chunks {
		records = append(records, StorageChunkDescriptorRecord{ObjectID: msg.Object.ObjectID, Descriptor: descriptor})
	}
	objects := append([]StorageObject(nil), state.Objects...)
	if _, found := findStorageObject(objects, msg.Object.ObjectID); found {
		return StorageStateV2{}, fmt.Errorf("storage object %s already exists", msg.Object.ObjectID)
	}
	objects = append(objects, msg.Object)
	chunks := append([]StorageChunkDescriptorRecord(nil), state.Chunks...)
	chunks = append(chunks, records...)
	replications := append([]ReplicationStatusCommitment(nil), state.Replications...)
	replications = upsertStorageReplication(replications, msg.Replication)
	return BuildStorageStateV2(objects, chunks, state.AccessReceipts, replications, msg.Height)
}

func UpdateStoragePolicyInStateV2(state StorageStateV2, msg MsgUpdateStoragePolicy) (StorageStateV2, error) {
	if err := msg.Validate(); err != nil {
		return StorageStateV2{}, err
	}
	objects := append([]StorageObject(nil), state.Objects...)
	index, found := findStorageObject(objects, msg.ObjectID)
	if !found {
		return StorageStateV2{}, fmt.Errorf("storage object %s not found", msg.ObjectID)
	}
	if objects[index].Owner != msg.Authority {
		return StorageStateV2{}, errors.New("storage policy update authority must own object")
	}
	objects[index].ReplicationPolicy = msg.ReplicationPolicy
	objects[index].AccessPolicy = msg.AccessPolicy
	objects[index].StorageClass = msg.StorageClass
	objects[index].ObjectHash = ComputeStorageObjectHash(objects[index])
	replications := upsertStorageReplication(append([]ReplicationStatusCommitment(nil), state.Replications...), msg.Replication)
	return BuildStorageStateV2(objects, state.Chunks, state.AccessReceipts, replications, msg.Height)
}

func RenewStorageObjectInStateV2(state StorageStateV2, msg MsgRenewStorageObject) (StorageStateV2, error) {
	if err := msg.Validate(); err != nil {
		return StorageStateV2{}, err
	}
	objects := append([]StorageObject(nil), state.Objects...)
	index, found := findStorageObject(objects, msg.ObjectID)
	if !found {
		return StorageStateV2{}, fmt.Errorf("storage object %s not found", msg.ObjectID)
	}
	if objects[index].Owner != msg.Authority {
		return StorageStateV2{}, errors.New("storage renew authority must own object")
	}
	if msg.ExpiresHeight <= objects[index].CreatedHeight {
		return StorageStateV2{}, errors.New("storage renew expiry must be after creation")
	}
	objects[index].ExpiresHeightOptional = msg.ExpiresHeight
	objects[index].ObjectHash = ComputeStorageObjectHash(objects[index])
	return BuildStorageStateV2(objects, state.Chunks, state.AccessReceipts, state.Replications, msg.Height)
}

func DeleteStorageObjectInStateV2(state StorageStateV2, msg MsgDeleteStorageObject) (StorageStateV2, error) {
	if err := msg.Validate(); err != nil {
		return StorageStateV2{}, err
	}
	object, found := QueryStorageObject(state, msg.ObjectID)
	if !found {
		return StorageStateV2{}, fmt.Errorf("storage object %s not found", msg.ObjectID)
	}
	if object.Owner != msg.Authority {
		return StorageStateV2{}, errors.New("storage delete authority must own object")
	}
	objects := make([]StorageObject, 0, len(state.Objects))
	for _, candidate := range state.Objects {
		if candidate.ObjectID != msg.ObjectID {
			objects = append(objects, candidate)
		}
	}
	chunks := make([]StorageChunkDescriptorRecord, 0, len(state.Chunks))
	for _, record := range state.Chunks {
		if record.ObjectID != msg.ObjectID {
			chunks = append(chunks, record)
		}
	}
	receipts := make([]StorageAccessReceipt, 0, len(state.AccessReceipts))
	for _, receipt := range state.AccessReceipts {
		if receipt.ObjectID != msg.ObjectID {
			receipts = append(receipts, receipt)
		}
	}
	replications := make([]ReplicationStatusCommitment, 0, len(state.Replications))
	for _, replication := range state.Replications {
		if replication.ObjectID != msg.ObjectID {
			replications = append(replications, replication)
		}
	}
	return BuildStorageStateV2(objects, chunks, receipts, replications, msg.Height)
}

func SubmitStorageReceiptInStateV2(state StorageStateV2, msg MsgSubmitStorageReceipt) (StorageStateV2, error) {
	if err := msg.Validate(); err != nil {
		return StorageStateV2{}, err
	}
	object, found := QueryStorageObject(state, msg.Receipt.ObjectID)
	if !found {
		return StorageStateV2{}, fmt.Errorf("storage object %s not found", msg.Receipt.ObjectID)
	}
	if object.Owner != msg.Authority && msg.Receipt.Accessor != msg.Authority {
		return StorageStateV2{}, errors.New("storage receipt authority must be owner or accessor")
	}
	receipts := append([]StorageAccessReceipt(nil), state.AccessReceipts...)
	if _, found := findStorageReceipt(receipts, msg.Receipt.ReceiptID); found {
		return StorageStateV2{}, fmt.Errorf("storage receipt %s already exists", msg.Receipt.ReceiptID)
	}
	receipts = append(receipts, msg.Receipt)
	return BuildStorageStateV2(state.Objects, state.Chunks, receipts, state.Replications, msg.Height)
}

func VerifyStorageProofInStateV2(state StorageStateV2, msg MsgVerifyStorageProof) (StorageStateProof, error) {
	if err := msg.Validate(); err != nil {
		return StorageStateProof{}, err
	}
	object, found := QueryStorageObject(state, msg.ObjectID)
	if !found {
		return StorageStateProof{}, fmt.Errorf("storage object %s not found", msg.ObjectID)
	}
	if err := VerifyStorageChunkInclusionInObject(msg.Proof, object); err != nil {
		return StorageStateProof{}, err
	}
	key, err := StorageChunkDescriptorKey(msg.ObjectID, msg.Proof.ChunkIndex)
	if err != nil {
		return StorageStateProof{}, err
	}
	return QueryStorageProof(state, key)
}

func QueryStorageObject(state StorageStateV2, objectID string) (StorageObject, bool) {
	for _, object := range state.Objects {
		if object.ObjectID == objectID {
			return object, true
		}
	}
	return StorageObject{}, false
}

func QueryObjectByContentHash(state StorageStateV2, contentHash string) (StorageObject, bool) {
	for _, object := range state.Objects {
		if object.ContentHash == contentHash {
			return object, true
		}
	}
	return StorageObject{}, false
}

func QueryChunkDescriptor(state StorageStateV2, objectID string, chunkIndex uint32) (StorageChunkDescriptor, bool) {
	for _, record := range state.Chunks {
		if record.ObjectID == objectID && record.Descriptor.ChunkIndex == chunkIndex {
			return record.Descriptor, true
		}
	}
	return StorageChunkDescriptor{}, false
}

func QueryStorageObjectsByOwner(state StorageStateV2, owner string) []StorageObject {
	out := make([]StorageObject, 0)
	for _, object := range state.Objects {
		if object.Owner == owner {
			out = append(out, object)
		}
	}
	return normalizeStorageObjects(out)
}

func QueryStorageAccessReceipt(state StorageStateV2, objectID, accessID string) (StorageAccessReceipt, bool) {
	for _, receipt := range state.AccessReceipts {
		if receipt.ObjectID == objectID && receipt.ReceiptID == accessID {
			return receipt, true
		}
	}
	return StorageAccessReceipt{}, false
}

func QueryStorageRoot(state StorageStateV2, height uint64) (StorageRoot, bool) {
	if state.Height != height || state.Root.Height != height {
		return StorageRoot{}, false
	}
	return state.Root, true
}

func QueryStorageProof(state StorageStateV2, key string) (StorageStateProof, error) {
	entries, err := storageStateProofEntries(state)
	if err != nil {
		return StorageStateProof{}, err
	}
	valueHash := ""
	proofHashes := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Key == key {
			valueHash = entry.Value
			continue
		}
		proofHashes = append(proofHashes, entry.EntryHash)
	}
	if valueHash == "" {
		return StorageStateProof{}, fmt.Errorf("storage proof key %s not found", key)
	}
	proof := StorageStateProof{
		Key:		key,
		ValueHash:	valueHash,
		Root:		state.Root.StateRoot,
		Height:		state.Height,
		ProofHashes:	normalizeStorageHashes(proofHashes),
	}
	proof.ProofHash = ComputeStorageStateProofHash(proof)
	return proof, proof.Validate()
}

func (commitment ReplicationStatusCommitment) ValidateFormat() error {
	if err := validateStorageToken("storage replication object id", commitment.ObjectID); err != nil {
		return err
	}
	if !IsStorageReplicationPolicy(commitment.ReplicationPolicy) {
		return fmt.Errorf("unknown storage replication policy %q", commitment.ReplicationPolicy)
	}
	if !IsStorageClass(commitment.StorageClass) {
		return fmt.Errorf("unknown storage class %q", commitment.StorageClass)
	}
	if commitment.ReplicaCount == 0 {
		return errors.New("storage replication replica count must be positive")
	}
	if commitment.AvailabilityBps > 10000 {
		return errors.New("storage replication availability bps must be <= 10000")
	}
	if commitment.LastVerifiedHeight == 0 {
		return errors.New("storage replication last verified height must be positive")
	}
	if commitment.CommitmentHash != "" {
		return validateStorageHash("storage replication commitment hash", commitment.CommitmentHash)
	}
	return nil
}

func (commitment ReplicationStatusCommitment) Validate() error {
	if err := commitment.ValidateFormat(); err != nil {
		return err
	}
	if commitment.CommitmentHash == "" {
		return errors.New("storage replication commitment hash is required")
	}
	if commitment.CommitmentHash != ComputeStorageReplicationStatusCommitmentHash(commitment) {
		return errors.New("storage replication commitment hash mismatch")
	}
	return nil
}

func (record StorageChunkDescriptorRecord) Validate(params StorageChunkParams) error {
	if err := validateStorageToken("storage chunk record object id", record.ObjectID); err != nil {
		return err
	}
	if err := record.Descriptor.Validate(params); err != nil {
		return err
	}
	if record.RecordHash == "" {
		return errors.New("storage chunk record hash is required")
	}
	if record.RecordHash != ComputeStorageChunkDescriptorRecordHash(record) {
		return errors.New("storage chunk record hash mismatch")
	}
	return nil
}

func (entry StorageIndexEntry) Validate() error {
	if entry.Key == "" || entry.Value == "" {
		return errors.New("storage index entry key and value are required")
	}
	if entry.EntryHash == "" {
		return errors.New("storage index entry hash is required")
	}
	if err := validateStorageHash("storage index entry hash", entry.EntryHash); err != nil {
		return err
	}
	if entry.EntryHash != ComputeStorageIndexEntryHash(entry.Key, entry.Value) {
		return errors.New("storage index entry hash mismatch")
	}
	return nil
}

func (root StorageRoot) Validate() error {
	if root.Height == 0 {
		return errors.New("storage root height must be positive")
	}
	if err := validateStorageHash("storage object root", root.ObjectRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage content index root", root.ContentIndexRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage chunk root", root.ChunkRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage owner index root", root.OwnerIndexRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage access root", root.AccessRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage replication root", root.ReplicationRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage state root", root.StateRoot); err != nil {
		return err
	}
	return nil
}

func (state StorageStateV2) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("storage state height must be positive")
	}
	if err := validateStorageObjects(state.Objects); err != nil {
		return err
	}
	if err := validateStorageChunkRecords(state.Chunks, state.Objects, DefaultStorageChunkParams()); err != nil {
		return err
	}
	if err := validateStorageIndexEntries("storage content index", state.ContentIndex); err != nil {
		return err
	}
	if err := validateStorageIndexEntries("storage owner index", state.OwnerIndex); err != nil {
		return err
	}
	if err := validateStorageReceiptsForObjects(state.AccessReceipts, state.Objects); err != nil {
		return err
	}
	if err := validateStorageReplications(state.Replications, state.Objects); err != nil {
		return err
	}
	if state.Root.StateRoot != "" {
		if err := state.Root.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (state StorageStateV2) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.Root.StateRoot == "" {
		return errors.New("storage state root is required")
	}
	expected := ComputeStorageRootV2(state)
	if state.Root != expected {
		return errors.New("storage state root mismatch")
	}
	return nil
}

func (msg MsgRegisterStorageObject) Validate(params StorageChunkParams) error {
	if err := validateStorageToken("storage register authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Object.Validate(); err != nil {
		return err
	}
	if err := validateStorageChunkDescriptors(msg.Chunks, params); err != nil {
		return err
	}
	validationParams := DefaultStorageValidationParams()
	validationParams.ChunkParams = params
	if err := ValidateStorageObjectAgainstChunks(msg.Object, msg.Chunks, validationParams); err != nil {
		return err
	}
	if err := msg.Replication.Validate(); err != nil {
		return err
	}
	if msg.Replication.ObjectID != msg.Object.ObjectID {
		return errors.New("storage register replication object mismatch")
	}
	if msg.Height == 0 {
		return errors.New("storage register height must be positive")
	}
	return validateOptionalStorageMessageHash("storage register message hash", msg.MessageHash)
}

func (msg MsgUpdateStoragePolicy) Validate() error {
	if err := validateStorageToken("storage policy authority", msg.Authority); err != nil {
		return err
	}
	if err := validateStorageToken("storage policy object id", msg.ObjectID); err != nil {
		return err
	}
	if !IsStorageReplicationPolicy(msg.ReplicationPolicy) {
		return fmt.Errorf("unknown storage replication policy %q", msg.ReplicationPolicy)
	}
	if !IsStorageAccessPolicy(msg.AccessPolicy) {
		return fmt.Errorf("unknown storage access policy %q", msg.AccessPolicy)
	}
	if !IsStorageClass(msg.StorageClass) {
		return fmt.Errorf("unknown storage class %q", msg.StorageClass)
	}
	if err := msg.Replication.Validate(); err != nil {
		return err
	}
	if msg.Replication.ObjectID != msg.ObjectID {
		return errors.New("storage policy replication object mismatch")
	}
	if msg.Replication.ReplicationPolicy != msg.ReplicationPolicy || msg.Replication.StorageClass != msg.StorageClass {
		return errors.New("storage policy replication settings mismatch")
	}
	if msg.Height == 0 {
		return errors.New("storage policy height must be positive")
	}
	return validateOptionalStorageMessageHash("storage policy message hash", msg.MessageHash)
}

func (msg MsgRenewStorageObject) Validate() error {
	if err := validateStorageToken("storage renew authority", msg.Authority); err != nil {
		return err
	}
	if err := validateStorageToken("storage renew object id", msg.ObjectID); err != nil {
		return err
	}
	if msg.ExpiresHeight == 0 {
		return errors.New("storage renew expiry must be positive")
	}
	if msg.Height == 0 {
		return errors.New("storage renew height must be positive")
	}
	return validateOptionalStorageMessageHash("storage renew message hash", msg.MessageHash)
}

func (msg MsgDeleteStorageObject) Validate() error {
	if err := validateStorageToken("storage delete authority", msg.Authority); err != nil {
		return err
	}
	if err := validateStorageToken("storage delete object id", msg.ObjectID); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("storage delete height must be positive")
	}
	return validateOptionalStorageMessageHash("storage delete message hash", msg.MessageHash)
}

func (msg MsgSubmitStorageReceipt) Validate() error {
	if err := validateStorageToken("storage receipt authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Receipt.Validate(); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("storage receipt height must be positive")
	}
	return validateOptionalStorageMessageHash("storage receipt message hash", msg.MessageHash)
}

func (msg MsgVerifyStorageProof) Validate() error {
	if err := validateStorageToken("storage proof authority", msg.Authority); err != nil {
		return err
	}
	if err := validateStorageToken("storage proof object id", msg.ObjectID); err != nil {
		return err
	}
	if err := msg.Proof.Validate(); err != nil {
		return err
	}
	if msg.Proof.ObjectID != msg.ObjectID {
		return errors.New("storage proof object mismatch")
	}
	if msg.Height == 0 {
		return errors.New("storage proof height must be positive")
	}
	return validateOptionalStorageMessageHash("storage proof message hash", msg.MessageHash)
}

func (proof StorageStateProof) Validate() error {
	if proof.Key == "" {
		return errors.New("storage proof key is required")
	}
	if err := validateStorageHash("storage proof value hash", proof.ValueHash); err != nil {
		return err
	}
	if err := validateStorageHash("storage proof root", proof.Root); err != nil {
		return err
	}
	if proof.Height == 0 {
		return errors.New("storage proof height must be positive")
	}
	for _, hash := range proof.ProofHashes {
		if err := validateStorageHash("storage proof path hash", hash); err != nil {
			return err
		}
	}
	if err := validateStorageHash("storage proof hash", proof.ProofHash); err != nil {
		return err
	}
	if proof.ProofHash != ComputeStorageStateProofHash(proof) {
		return errors.New("storage proof hash mismatch")
	}
	return nil
}

func ComputeStorageReplicationStatusCommitmentHash(commitment ReplicationStatusCommitment) string {
	return storageHashParts(
		"storage-replication-status-v1",
		commitment.ObjectID,
		commitment.ReplicationPolicy,
		commitment.StorageClass,
		fmt.Sprintf("%020d", commitment.ReplicaCount),
		fmt.Sprintf("%020d", commitment.AvailabilityBps),
		fmt.Sprintf("%020d", commitment.LastVerifiedHeight),
	)
}

func ComputeStorageChunkDescriptorRecordHash(record StorageChunkDescriptorRecord) string {
	return storageHashParts("storage-chunk-record-v1", record.ObjectID, fmt.Sprintf("%020d", record.Descriptor.ChunkIndex), record.Descriptor.DescriptorHash)
}

func ComputeStorageIndexEntryHash(key, value string) string {
	return storageHashParts("storage-index-entry-v1", key, value)
}

func ComputeStorageIndexRoot(entries []StorageIndexEntry) string {
	ordered := normalizeStorageIndexEntries(entries)
	parts := []string{"storage-index-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, entry := range ordered {
		parts = append(parts, entry.EntryHash)
	}
	return storageHashParts(parts...)
}

func ComputeStorageChunkDescriptorRecordRoot(records []StorageChunkDescriptorRecord) string {
	ordered := normalizeStorageChunkDescriptorRecords(records)
	parts := []string{"storage-chunk-record-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.RecordHash)
	}
	return storageHashParts(parts...)
}

func ComputeStorageReplicationRoot(commitments []ReplicationStatusCommitment) string {
	ordered := normalizeStorageReplications(commitments)
	parts := []string{"storage-replication-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, commitment := range ordered {
		parts = append(parts, commitment.CommitmentHash)
	}
	return storageHashParts(parts...)
}

func ComputeStorageRootV2(state StorageStateV2) StorageRoot {
	root := StorageRoot{
		Height:			state.Height,
		ObjectRoot:		ComputeStorageObjectRoot(state.Objects),
		ContentIndexRoot:	ComputeStorageIndexRoot(state.ContentIndex),
		ChunkRoot:		ComputeStorageChunkDescriptorRecordRoot(state.Chunks),
		OwnerIndexRoot:		ComputeStorageIndexRoot(state.OwnerIndex),
		AccessRoot:		ComputeStorageAccessReceiptRoot(state.AccessReceipts),
		ReplicationRoot:	ComputeStorageReplicationRoot(state.Replications),
	}
	root.StateRoot = storageHashParts(
		"storage-state-root-v2",
		fmt.Sprintf("%020d", root.Height),
		root.ObjectRoot,
		root.ContentIndexRoot,
		root.ChunkRoot,
		root.OwnerIndexRoot,
		root.AccessRoot,
		root.ReplicationRoot,
	)
	return root
}

func ComputeStorageStateProofHash(proof StorageStateProof) string {
	parts := []string{"storage-state-proof-v1", proof.Key, proof.ValueHash, proof.Root, fmt.Sprintf("%020d", proof.Height), fmt.Sprintf("%020d", len(proof.ProofHashes))}
	parts = append(parts, normalizeStorageHashes(proof.ProofHashes)...)
	return storageHashParts(parts...)
}

func (state *StorageStateV2) populateIndexes() error {
	content := make([]StorageIndexEntry, 0, len(state.Objects))
	owner := make([]StorageIndexEntry, 0, len(state.Objects))
	for _, object := range state.Objects {
		contentKey, err := StorageContentIndexKey(object.ContentHash)
		if err != nil {
			return err
		}
		ownerKey, err := StorageOwnerIndexKey(object.Owner, object.ObjectID)
		if err != nil {
			return err
		}
		content = append(content, newStorageIndexEntry(contentKey, object.ObjectID))
		owner = append(owner, newStorageIndexEntry(ownerKey, object.ObjectID))
	}
	state.ContentIndex = normalizeStorageIndexEntries(content)
	state.OwnerIndex = normalizeStorageIndexEntries(owner)
	for i := range state.Chunks {
		state.Chunks[i].RecordHash = ComputeStorageChunkDescriptorRecordHash(state.Chunks[i])
	}
	return nil
}

func newStorageIndexEntry(key, value string) StorageIndexEntry {
	return StorageIndexEntry{Key: key, Value: value, EntryHash: ComputeStorageIndexEntryHash(key, value)}
}

func storageStateProofEntries(state StorageStateV2) ([]StorageIndexEntry, error) {
	entries := make([]StorageIndexEntry, 0, len(state.Objects)+len(state.ContentIndex)+len(state.Chunks)+len(state.OwnerIndex)+len(state.AccessReceipts)+len(state.Replications)+1)
	for _, object := range state.Objects {
		key, err := StorageObjectKey(object.ObjectID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newStorageIndexEntry(key, object.ObjectHash))
	}
	entries = append(entries, state.ContentIndex...)
	for _, record := range state.Chunks {
		key, err := StorageChunkDescriptorKey(record.ObjectID, record.Descriptor.ChunkIndex)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newStorageIndexEntry(key, record.RecordHash))
	}
	entries = append(entries, state.OwnerIndex...)
	for _, receipt := range state.AccessReceipts {
		key, err := StorageAccessReceiptKey(receipt.ObjectID, receipt.ReceiptID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newStorageIndexEntry(key, receipt.ReceiptHash))
	}
	for _, replication := range state.Replications {
		key, err := StorageReplicationKey(replication.ObjectID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, newStorageIndexEntry(key, replication.CommitmentHash))
	}
	key, err := StorageRootKey(state.Height)
	if err != nil {
		return nil, err
	}
	entries = append(entries, newStorageIndexEntry(key, state.Root.StateRoot))
	return normalizeStorageIndexEntries(entries), nil
}

func normalizeStorageIndexEntries(entries []StorageIndexEntry) []StorageIndexEntry {
	out := append([]StorageIndexEntry(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func normalizeStorageChunkDescriptorRecords(records []StorageChunkDescriptorRecord) []StorageChunkDescriptorRecord {
	out := append([]StorageChunkDescriptorRecord(nil), records...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ObjectID == out[j].ObjectID {
			return out[i].Descriptor.ChunkIndex < out[j].Descriptor.ChunkIndex
		}
		return out[i].ObjectID < out[j].ObjectID
	})
	return out
}

func normalizeStorageReplications(commitments []ReplicationStatusCommitment) []ReplicationStatusCommitment {
	out := append([]ReplicationStatusCommitment(nil), commitments...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ObjectID < out[j].ObjectID })
	return out
}

func validateStorageIndexEntries(name string, entries []StorageIndexEntry) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seen[entry.Key]; found {
			return fmt.Errorf("duplicate %s key %s", name, entry.Key)
		}
		seen[entry.Key] = struct{}{}
		if previous != "" && previous >= entry.Key {
			return fmt.Errorf("%s entries must be sorted canonically", name)
		}
		previous = entry.Key
	}
	return nil
}

func validateStorageChunkRecords(records []StorageChunkDescriptorRecord, objects []StorageObject, params StorageChunkParams) error {
	byObject := make(map[string][]StorageChunkDescriptor)
	objectIDs := map[string]StorageObject{}
	for _, object := range objects {
		objectIDs[object.ObjectID] = object
	}
	seen := map[string]struct{}{}
	previousObject := ""
	var previousIndex uint32
	for _, record := range records {
		if err := record.Validate(params); err != nil {
			return err
		}
		object, found := objectIDs[record.ObjectID]
		if !found {
			return fmt.Errorf("storage chunk record references unknown object %s", record.ObjectID)
		}
		key := record.ObjectID + "/" + fmt.Sprintf("%020d", record.Descriptor.ChunkIndex)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate storage chunk record %s", key)
		}
		seen[key] = struct{}{}
		if previousObject != "" {
			if previousObject > record.ObjectID || previousObject == record.ObjectID && previousIndex >= record.Descriptor.ChunkIndex {
				return errors.New("storage chunk records must be sorted canonically")
			}
		}
		previousObject = record.ObjectID
		previousIndex = record.Descriptor.ChunkIndex
		byObject[object.ObjectID] = append(byObject[object.ObjectID], record.Descriptor)
	}
	for _, object := range objects {
		descriptors := byObject[object.ObjectID]
		if len(descriptors) == 0 {
			return fmt.Errorf("storage object %s requires chunk descriptors", object.ObjectID)
		}
		validationParams := DefaultStorageValidationParams()
		validationParams.ChunkParams = params
		if err := ValidateStorageObjectAgainstChunks(object, descriptors, validationParams); err != nil {
			return err
		}
	}
	return nil
}

func validateStorageReceiptsForObjects(receipts []StorageAccessReceipt, objects []StorageObject) error {
	if err := validateStorageReceipts(receipts); err != nil {
		return err
	}
	objectIDs := map[string]StorageObject{}
	for _, object := range objects {
		objectIDs[object.ObjectID] = object
	}
	for _, receipt := range receipts {
		object, found := objectIDs[receipt.ObjectID]
		if !found {
			return fmt.Errorf("storage receipt references unknown object %s", receipt.ObjectID)
		}
		if receipt.ContentHash != object.ContentHash {
			return errors.New("storage receipt content hash mismatch")
		}
		if receipt.PolicyHash != ComputeStoragePolicyHash(object.ReplicationPolicy, object.AccessPolicy, object.StorageClass) {
			return errors.New("storage receipt policy hash mismatch")
		}
	}
	return nil
}

func validateStorageReplications(commitments []ReplicationStatusCommitment, objects []StorageObject) error {
	seen := map[string]struct{}{}
	objectIDs := map[string]StorageObject{}
	for _, object := range objects {
		objectIDs[object.ObjectID] = object
	}
	previous := ""
	for _, commitment := range commitments {
		if err := commitment.Validate(); err != nil {
			return err
		}
		object, found := objectIDs[commitment.ObjectID]
		if !found {
			return fmt.Errorf("storage replication references unknown object %s", commitment.ObjectID)
		}
		if commitment.ReplicationPolicy != object.ReplicationPolicy || commitment.StorageClass != object.StorageClass {
			return errors.New("storage replication policy mismatch")
		}
		if _, found := seen[commitment.ObjectID]; found {
			return fmt.Errorf("duplicate storage replication %s", commitment.ObjectID)
		}
		seen[commitment.ObjectID] = struct{}{}
		if previous != "" && previous >= commitment.ObjectID {
			return errors.New("storage replications must be sorted canonically")
		}
		previous = commitment.ObjectID
	}
	return nil
}

func findStorageObject(objects []StorageObject, objectID string) (int, bool) {
	for i, object := range objects {
		if object.ObjectID == objectID {
			return i, true
		}
	}
	return 0, false
}

func findStorageReceipt(receipts []StorageAccessReceipt, receiptID string) (int, bool) {
	for i, receipt := range receipts {
		if receipt.ReceiptID == receiptID {
			return i, true
		}
	}
	return 0, false
}

func upsertStorageReplication(replications []ReplicationStatusCommitment, replication ReplicationStatusCommitment) []ReplicationStatusCommitment {
	for i, candidate := range replications {
		if candidate.ObjectID == replication.ObjectID {
			replications[i] = replication
			return replications
		}
	}
	return append(replications, replication)
}

func validateOptionalStorageMessageHash(field, value string) error {
	if value == "" {
		return nil
	}
	return validateStorageHash(field, value)
}
