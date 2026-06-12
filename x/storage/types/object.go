package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxStorageChunkRoots		= 4096
	MaxStorageObjectSize		= uint64(1 << 40)
	MaxStorageTokenBytes		= 128
	MaxStorageReceiptAccesses	= 1024
	StorageObjectVersionV1		= uint64(1)

	StorageClassHot		= "hot"
	StorageClassWarm	= "warm"
	StorageClassCold	= "cold"
	StorageClassArchive	= "archive"

	ReplicationPolicySingle		= "single"
	ReplicationPolicyRegional	= "regional"
	ReplicationPolicyMultiZone	= "multi_zone"
	ReplicationPolicyErasure	= "erasure_coded"

	AccessPolicyPrivate		= "private"
	AccessPolicyPublicRead		= "public_read"
	AccessPolicyPermissioned	= "permissioned"
)

type StorageObject struct {
	ContentHash		string
	ChunkRoots		[]string
	Size			uint64
	ReplicationPolicy	string
	AccessPolicy		string
	ObjectID		string
	Owner			string
	StorageClass		string
	CreatedHeight		uint64
	ExpiresHeightOptional	uint64
	MetadataHashOptional	string
	AvailabilityCommitment	string
	Version			uint64
	ObjectHash		string
}

type StorageChunkSet struct {
	ObjectID	string
	ChunkRoots	[]string
	ChunkRoot	string
	ChunkCount	uint32
}

type StorageRetrievalProof struct {
	ObjectID	string
	ContentHash	string
	ChunkRoot	string
	ChunkIndex	uint32
	ChunkHash	string
	ProofPath	[]string
	ProofHash	string
}

type StorageAccessReceipt struct {
	ReceiptID	string
	ObjectID	string
	Accessor	string
	AccessType	string
	AccessHeight	uint64
	ContentHash	string
	ChunkRoot	string
	PolicyHash	string
	RetrievalProof	string
	ReceiptHash	string
}

type StorageObjectCommitmentState struct {
	Objects		[]StorageObject
	Receipts	[]StorageAccessReceipt
	Height		uint64
	RootHash	string
}

func NewStorageObject(object StorageObject) (StorageObject, error) {
	if object.ObjectHash != "" {
		return StorageObject{}, errors.New("storage object hash must be empty before construction")
	}
	object.ChunkRoots = normalizeStorageHashes(object.ChunkRoots)
	if object.ObjectID == "" {
		object.ObjectID = ComputeStorageObjectID(object)
	}
	if object.Version == 0 {
		object.Version = StorageObjectVersionV1
	}
	if err := object.ValidateFormat(); err != nil {
		return StorageObject{}, err
	}
	object.ObjectHash = ComputeStorageObjectHash(object)
	return object, object.Validate()
}

func NewStorageChunkSet(objectID string, chunkRoots []string) (StorageChunkSet, error) {
	if err := validateStorageToken("storage chunk set object id", objectID); err != nil {
		return StorageChunkSet{}, err
	}
	ordered := normalizeStorageHashes(chunkRoots)
	if len(ordered) == 0 {
		return StorageChunkSet{}, errors.New("storage chunk set requires roots")
	}
	if len(ordered) > MaxStorageChunkRoots {
		return StorageChunkSet{}, fmt.Errorf("storage chunk set must not exceed %d roots", MaxStorageChunkRoots)
	}
	set := StorageChunkSet{
		ObjectID:	objectID,
		ChunkRoots:	ordered,
		ChunkRoot:	ComputeStorageChunkRoot(ordered),
		ChunkCount:	uint32(len(ordered)),
	}
	return set, set.Validate()
}

func NewStorageRetrievalProof(proof StorageRetrievalProof) (StorageRetrievalProof, error) {
	if proof.ProofHash != "" {
		return StorageRetrievalProof{}, errors.New("storage retrieval proof hash must be empty before construction")
	}
	if err := proof.ValidateFormat(); err != nil {
		return StorageRetrievalProof{}, err
	}
	proof.ProofHash = ComputeStorageRetrievalProofHash(proof)
	return proof, proof.Validate()
}

func NewStorageAccessReceipt(receipt StorageAccessReceipt) (StorageAccessReceipt, error) {
	if receipt.ReceiptHash != "" {
		return StorageAccessReceipt{}, errors.New("storage access receipt hash must be empty before construction")
	}
	if receipt.ReceiptID == "" {
		receipt.ReceiptID = ComputeStorageAccessReceiptID(receipt)
	}
	if err := receipt.ValidateFormat(); err != nil {
		return StorageAccessReceipt{}, err
	}
	receipt.ReceiptHash = ComputeStorageAccessReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func BuildStorageObjectCommitmentState(objects []StorageObject, receipts []StorageAccessReceipt, height uint64) (StorageObjectCommitmentState, error) {
	state := StorageObjectCommitmentState{
		Objects:	normalizeStorageObjects(objects),
		Receipts:	normalizeStorageReceipts(receipts),
		Height:		height,
	}
	if err := state.ValidateFormat(); err != nil {
		return StorageObjectCommitmentState{}, err
	}
	state.RootHash = ComputeStorageObjectCommitmentStateRoot(state)
	return state, state.Validate()
}

func (object StorageObject) ValidateFormat() error {
	if err := validateStorageHash("storage object content hash", object.ContentHash); err != nil {
		return err
	}
	if len(object.ChunkRoots) == 0 {
		return errors.New("storage object requires chunk roots")
	}
	if len(object.ChunkRoots) > MaxStorageChunkRoots {
		return fmt.Errorf("storage object chunk roots must not exceed %d", MaxStorageChunkRoots)
	}
	for i, root := range object.ChunkRoots {
		if err := validateStorageHash("storage object chunk root", root); err != nil {
			return err
		}
		if i > 0 && object.ChunkRoots[i-1] >= root {
			return errors.New("storage object chunk roots must be sorted canonically")
		}
	}
	if object.Size == 0 || object.Size > MaxStorageObjectSize {
		return fmt.Errorf("storage object size must be between 1 and %d", MaxStorageObjectSize)
	}
	if !IsStorageReplicationPolicy(object.ReplicationPolicy) {
		return fmt.Errorf("unknown storage replication policy %q", object.ReplicationPolicy)
	}
	if !IsStorageAccessPolicy(object.AccessPolicy) {
		return fmt.Errorf("unknown storage access policy %q", object.AccessPolicy)
	}
	if err := validateStorageToken("storage object id", object.ObjectID); err != nil {
		return err
	}
	if err := validateStorageToken("storage object owner", object.Owner); err != nil {
		return err
	}
	if !IsStorageClass(object.StorageClass) {
		return fmt.Errorf("unknown storage class %q", object.StorageClass)
	}
	if object.CreatedHeight == 0 {
		return errors.New("storage object created height must be positive")
	}
	if object.ExpiresHeightOptional != 0 && object.ExpiresHeightOptional <= object.CreatedHeight {
		return errors.New("storage object expiry must be after creation")
	}
	if object.MetadataHashOptional != "" {
		if err := validateStorageHash("storage object metadata hash", object.MetadataHashOptional); err != nil {
			return err
		}
	}
	if err := validateStorageHash("storage object availability commitment", object.AvailabilityCommitment); err != nil {
		return err
	}
	if object.Version == 0 {
		return errors.New("storage object version must be positive")
	}
	if object.ObjectHash != "" {
		return validateStorageHash("storage object hash", object.ObjectHash)
	}
	return nil
}

func (object StorageObject) Validate() error {
	if err := object.ValidateFormat(); err != nil {
		return err
	}
	if object.ObjectHash == "" {
		return errors.New("storage object hash is required")
	}
	if object.ObjectHash != ComputeStorageObjectHash(object) {
		return errors.New("storage object hash mismatch")
	}
	return nil
}

func (set StorageChunkSet) Validate() error {
	if err := validateStorageToken("storage chunk set object id", set.ObjectID); err != nil {
		return err
	}
	if len(set.ChunkRoots) == 0 {
		return errors.New("storage chunk set requires roots")
	}
	for i, root := range set.ChunkRoots {
		if err := validateStorageHash("storage chunk set chunk root", root); err != nil {
			return err
		}
		if i > 0 && set.ChunkRoots[i-1] >= root {
			return errors.New("storage chunk set roots must be sorted canonically")
		}
	}
	if set.ChunkCount != uint32(len(set.ChunkRoots)) {
		return errors.New("storage chunk set count mismatch")
	}
	if set.ChunkRoot != ComputeStorageChunkRoot(set.ChunkRoots) {
		return errors.New("storage chunk set root mismatch")
	}
	return validateStorageHash("storage chunk set root", set.ChunkRoot)
}

func (proof StorageRetrievalProof) ValidateFormat() error {
	if err := validateStorageToken("storage retrieval proof object id", proof.ObjectID); err != nil {
		return err
	}
	if err := validateStorageHash("storage retrieval proof content hash", proof.ContentHash); err != nil {
		return err
	}
	if err := validateStorageHash("storage retrieval proof chunk root", proof.ChunkRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage retrieval proof chunk hash", proof.ChunkHash); err != nil {
		return err
	}
	for _, item := range proof.ProofPath {
		if err := validateStorageHash("storage retrieval proof path", item); err != nil {
			return err
		}
	}
	if proof.ProofHash != "" {
		return validateStorageHash("storage retrieval proof hash", proof.ProofHash)
	}
	return nil
}

func (proof StorageRetrievalProof) Validate() error {
	if err := proof.ValidateFormat(); err != nil {
		return err
	}
	if proof.ProofHash == "" {
		return errors.New("storage retrieval proof hash is required")
	}
	if proof.ProofHash != ComputeStorageRetrievalProofHash(proof) {
		return errors.New("storage retrieval proof hash mismatch")
	}
	return nil
}

func (receipt StorageAccessReceipt) ValidateFormat() error {
	if err := validateStorageToken("storage access receipt id", receipt.ReceiptID); err != nil {
		return err
	}
	if err := validateStorageToken("storage access receipt object id", receipt.ObjectID); err != nil {
		return err
	}
	if err := validateStorageToken("storage access receipt accessor", receipt.Accessor); err != nil {
		return err
	}
	if err := validateStorageToken("storage access receipt access type", receipt.AccessType); err != nil {
		return err
	}
	if receipt.AccessHeight == 0 {
		return errors.New("storage access receipt height must be positive")
	}
	if err := validateStorageHash("storage access receipt content hash", receipt.ContentHash); err != nil {
		return err
	}
	if err := validateStorageHash("storage access receipt chunk root", receipt.ChunkRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage access receipt policy hash", receipt.PolicyHash); err != nil {
		return err
	}
	if receipt.RetrievalProof != "" {
		if err := validateStorageHash("storage access receipt retrieval proof", receipt.RetrievalProof); err != nil {
			return err
		}
	}
	if receipt.ReceiptHash != "" {
		return validateStorageHash("storage access receipt hash", receipt.ReceiptHash)
	}
	return nil
}

func (receipt StorageAccessReceipt) Validate() error {
	if err := receipt.ValidateFormat(); err != nil {
		return err
	}
	if receipt.ReceiptHash == "" {
		return errors.New("storage access receipt hash is required")
	}
	if receipt.ReceiptHash != ComputeStorageAccessReceiptHash(receipt) {
		return errors.New("storage access receipt hash mismatch")
	}
	return nil
}

func (state StorageObjectCommitmentState) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("storage object commitment state height must be positive")
	}
	if err := validateStorageObjects(state.Objects); err != nil {
		return err
	}
	if err := validateStorageReceipts(state.Receipts); err != nil {
		return err
	}
	if state.RootHash != "" {
		return validateStorageHash("storage object commitment state root", state.RootHash)
	}
	return nil
}

func (state StorageObjectCommitmentState) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.RootHash == "" {
		return errors.New("storage object commitment state root is required")
	}
	if state.RootHash != ComputeStorageObjectCommitmentStateRoot(state) {
		return errors.New("storage object commitment state root mismatch")
	}
	return nil
}

func ComputeStorageObjectID(object StorageObject) string {
	return "object/" + storageHashParts("storage-object-id-v1", object.ContentHash, fmt.Sprintf("%020d", object.Size), object.Owner)
}

func ComputeStorageObjectHash(object StorageObject) string {
	chunkRoot := ComputeStorageChunkRoot(object.ChunkRoots)
	return storageHashParts(
		"storage-object-v1",
		object.ContentHash,
		chunkRoot,
		fmt.Sprintf("%020d", object.Size),
		object.ReplicationPolicy,
		object.AccessPolicy,
		object.ObjectID,
		object.Owner,
		object.StorageClass,
		fmt.Sprintf("%020d", object.CreatedHeight),
		fmt.Sprintf("%020d", object.ExpiresHeightOptional),
		object.MetadataHashOptional,
		object.AvailabilityCommitment,
		fmt.Sprintf("%020d", object.Version),
	)
}

func ComputeStorageChunkRoot(chunkRoots []string) string {
	ordered := normalizeStorageHashes(chunkRoots)
	parts := []string{"storage-chunk-root-v1", fmt.Sprintf("%020d", len(ordered))}
	parts = append(parts, ordered...)
	return storageHashParts(parts...)
}

func ComputeStoragePolicyHash(replicationPolicy, accessPolicy, storageClass string) string {
	return storageHashParts("storage-policy-v1", replicationPolicy, accessPolicy, storageClass)
}

func ComputeStorageRetrievalProofHash(proof StorageRetrievalProof) string {
	parts := []string{
		"storage-retrieval-proof-v1",
		proof.ObjectID,
		proof.ContentHash,
		proof.ChunkRoot,
		fmt.Sprintf("%020d", proof.ChunkIndex),
		proof.ChunkHash,
		fmt.Sprintf("%020d", len(proof.ProofPath)),
	}
	parts = append(parts, proof.ProofPath...)
	return storageHashParts(parts...)
}

func ComputeStorageAccessReceiptID(receipt StorageAccessReceipt) string {
	return "receipt/" + storageHashParts("storage-access-receipt-id-v1", receipt.ObjectID, receipt.Accessor, receipt.AccessType, fmt.Sprintf("%020d", receipt.AccessHeight))
}

func ComputeStorageAccessReceiptHash(receipt StorageAccessReceipt) string {
	return storageHashParts(
		"storage-access-receipt-v1",
		receipt.ReceiptID,
		receipt.ObjectID,
		receipt.Accessor,
		receipt.AccessType,
		fmt.Sprintf("%020d", receipt.AccessHeight),
		receipt.ContentHash,
		receipt.ChunkRoot,
		receipt.PolicyHash,
		receipt.RetrievalProof,
	)
}

func ComputeStorageObjectRoot(objects []StorageObject) string {
	ordered := normalizeStorageObjects(objects)
	parts := []string{"storage-object-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, object := range ordered {
		parts = append(parts, object.ObjectHash)
	}
	return storageHashParts(parts...)
}

func ComputeStorageAccessReceiptRoot(receipts []StorageAccessReceipt) string {
	ordered := normalizeStorageReceipts(receipts)
	parts := []string{"storage-access-receipt-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ReceiptHash)
	}
	return storageHashParts(parts...)
}

func ComputeStorageObjectCommitmentStateRoot(state StorageObjectCommitmentState) string {
	return storageHashParts(
		"storage-object-commitment-state-root-v1",
		fmt.Sprintf("%020d", state.Height),
		ComputeStorageObjectRoot(state.Objects),
		ComputeStorageAccessReceiptRoot(state.Receipts),
	)
}

func IsStorageClass(class string) bool {
	switch class {
	case StorageClassHot, StorageClassWarm, StorageClassCold, StorageClassArchive:
		return true
	default:
		return false
	}
}

func IsStorageReplicationPolicy(policy string) bool {
	switch policy {
	case ReplicationPolicySingle, ReplicationPolicyRegional, ReplicationPolicyMultiZone, ReplicationPolicyErasure:
		return true
	default:
		return false
	}
}

func IsStorageAccessPolicy(policy string) bool {
	switch policy {
	case AccessPolicyPrivate, AccessPolicyPublicRead, AccessPolicyPermissioned:
		return true
	default:
		return false
	}
}

func normalizeStorageHashes(values []string) []string {
	out := append([]string(nil), values...)
	for i := range out {
		out[i] = strings.ToLower(strings.TrimSpace(out[i]))
	}
	sort.Strings(out)
	unique := make([]string, 0, len(out))
	for _, value := range out {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}

func normalizeStorageObjects(objects []StorageObject) []StorageObject {
	out := append([]StorageObject(nil), objects...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ObjectID < out[j].ObjectID })
	return out
}

func normalizeStorageReceipts(receipts []StorageAccessReceipt) []StorageAccessReceipt {
	out := append([]StorageAccessReceipt(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReceiptID < out[j].ReceiptID })
	return out
}

func validateStorageObjects(objects []StorageObject) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, object := range objects {
		if err := object.Validate(); err != nil {
			return err
		}
		if _, found := seen[object.ObjectID]; found {
			return fmt.Errorf("duplicate storage object %s", object.ObjectID)
		}
		seen[object.ObjectID] = struct{}{}
		if previous != "" && previous >= object.ObjectID {
			return errors.New("storage objects must be sorted canonically")
		}
		previous = object.ObjectID
	}
	return nil
}

func validateStorageReceipts(receipts []StorageAccessReceipt) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if _, found := seen[receipt.ReceiptID]; found {
			return fmt.Errorf("duplicate storage access receipt %s", receipt.ReceiptID)
		}
		seen[receipt.ReceiptID] = struct{}{}
		if previous != "" && previous >= receipt.ReceiptID {
			return errors.New("storage access receipts must be sorted canonically")
		}
		previous = receipt.ReceiptID
	}
	return nil
}

func validateStorageToken(field, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", field)
	}
	if len(value) > MaxStorageTokenBytes {
		return fmt.Errorf("%s must be <= %d bytes", field, MaxStorageTokenBytes)
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", field)
	}
	return nil
}

func validateStorageHash(field, value string) error {
	if len(value) != 64 {
		return fmt.Errorf("%s must be a 32-byte hex hash", field)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be lowercase hex", field)
	}
	return nil
}

func storageHashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))
		_, _ = h.Write(length[:])
		_, _ = h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}
