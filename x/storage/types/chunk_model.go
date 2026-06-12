package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	DefaultMaxStorageChunkBytes	= uint64(4 * 1024 * 1024)
	MaxStorageErasureGroupBytes	= 64
)

type StorageChunkParams struct {
	MaxChunkBytes uint64
}

type StorageChunkDescriptor struct {
	ChunkIndex		uint32
	ChunkHash		string
	ChunkSize		uint64
	ChunkProofRoot		string
	ErasureGroupOptional	string
	DescriptorHash		string
}

type StorageChunkInclusionProof struct {
	ObjectID	string
	ContentHash	string
	ObjectRoot	string
	ChunkIndex	uint32
	ChunkHash	string
	ChunkProofRoot	string
	ProofPath	[]string
	ProofHash	string
}

func DefaultStorageChunkParams() StorageChunkParams {
	return StorageChunkParams{MaxChunkBytes: DefaultMaxStorageChunkBytes}
}

func (params StorageChunkParams) Validate() error {
	if params.MaxChunkBytes == 0 {
		return errors.New("storage chunk max bytes must be positive")
	}
	return nil
}

func NewStorageChunkDescriptor(descriptor StorageChunkDescriptor, params StorageChunkParams) (StorageChunkDescriptor, error) {
	if descriptor.DescriptorHash != "" {
		return StorageChunkDescriptor{}, errors.New("storage chunk descriptor hash must be empty before construction")
	}
	if err := descriptor.ValidateFormat(params); err != nil {
		return StorageChunkDescriptor{}, err
	}
	descriptor.DescriptorHash = ComputeStorageChunkDescriptorHash(descriptor)
	return descriptor, descriptor.Validate(params)
}

func NewStorageChunkInclusionProof(proof StorageChunkInclusionProof) (StorageChunkInclusionProof, error) {
	if proof.ProofHash != "" {
		return StorageChunkInclusionProof{}, errors.New("storage chunk inclusion proof hash must be empty before construction")
	}
	if err := proof.ValidateFormat(); err != nil {
		return StorageChunkInclusionProof{}, err
	}
	proof.ProofHash = ComputeStorageChunkInclusionProofHash(proof)
	return proof, proof.Validate()
}

func BuildStorageChunkDescriptors(chunkHashes []string, chunkSizes []uint64, erasureGroups []string, params StorageChunkParams) ([]StorageChunkDescriptor, string, error) {
	if err := params.Validate(); err != nil {
		return nil, "", err
	}
	if len(chunkHashes) == 0 {
		return nil, "", errors.New("storage chunk descriptors require chunks")
	}
	if len(chunkHashes) > MaxStorageChunkRoots {
		return nil, "", fmt.Errorf("storage chunk descriptors must not exceed %d", MaxStorageChunkRoots)
	}
	if len(chunkHashes) != len(chunkSizes) {
		return nil, "", errors.New("storage chunk descriptors hash and size count mismatch")
	}
	if len(erasureGroups) != 0 && len(erasureGroups) != len(chunkHashes) {
		return nil, "", errors.New("storage chunk descriptors erasure group count mismatch")
	}
	out := make([]StorageChunkDescriptor, 0, len(chunkHashes))
	for i, chunkHash := range chunkHashes {
		erasureGroup := ""
		if len(erasureGroups) > 0 {
			erasureGroup = erasureGroups[i]
		}
		proofRoot := ComputeStorageChunkProofRoot(uint32(i), chunkHash, chunkSizes[i], erasureGroup)
		descriptor, err := NewStorageChunkDescriptor(StorageChunkDescriptor{
			ChunkIndex:		uint32(i),
			ChunkHash:		chunkHash,
			ChunkSize:		chunkSizes[i],
			ChunkProofRoot:		proofRoot,
			ErasureGroupOptional:	erasureGroup,
		}, params)
		if err != nil {
			return nil, "", err
		}
		out = append(out, descriptor)
	}
	contentHash := ComputeStorageContentHashFromChunks(out)
	return out, contentHash, nil
}

func BuildStorageObjectFromChunkDescriptors(object StorageObject, descriptors []StorageChunkDescriptor, params StorageChunkParams) (StorageObject, error) {
	if err := validateStorageChunkDescriptors(descriptors, params); err != nil {
		return StorageObject{}, err
	}
	object.ContentHash = ComputeStorageContentHashFromChunks(descriptors)
	object.ChunkRoots = storageChunkDescriptorRoots(descriptors)
	return NewStorageObject(object)
}

func VerifyStorageObjectContentHash(object StorageObject, descriptors []StorageChunkDescriptor, params StorageChunkParams) error {
	if err := object.Validate(); err != nil {
		return err
	}
	if err := validateStorageChunkDescriptors(descriptors, params); err != nil {
		return err
	}
	contentHash := ComputeStorageContentHashFromChunks(descriptors)
	if object.ContentHash != contentHash {
		return errors.New("storage object content_hash must commit to ordered chunk roots")
	}
	descriptorRoots := normalizeStorageHashes(storageChunkDescriptorRoots(descriptors))
	if !equalStringSlices(object.ChunkRoots, descriptorRoots) {
		return errors.New("storage object chunk roots mismatch descriptors")
	}
	return nil
}

func VerifyStorageChunkInclusionInObject(proof StorageChunkInclusionProof, object StorageObject) error {
	if err := proof.Validate(); err != nil {
		return err
	}
	if err := object.Validate(); err != nil {
		return err
	}
	if proof.ObjectID != object.ObjectID {
		return errors.New("storage chunk inclusion proof object mismatch")
	}
	if proof.ContentHash != object.ContentHash {
		return errors.New("storage chunk inclusion proof content hash mismatch")
	}
	if proof.ObjectRoot != object.ObjectHash {
		return errors.New("storage chunk inclusion proof object root mismatch")
	}
	found := false
	for _, chunkRoot := range object.ChunkRoots {
		if chunkRoot == proof.ChunkProofRoot {
			found = true
			break
		}
	}
	if !found {
		return errors.New("storage chunk inclusion proof chunk root not in object root")
	}
	return nil
}

func ValidateUnrelatedStorageTransitionDoesNotRequireChunks(object StorageObject) error {
	if err := object.Validate(); err != nil {
		return err
	}
	if len(object.ChunkRoots) == 0 || object.ContentHash == "" || object.ObjectHash == "" {
		return errors.New("storage unrelated transition requires commitments only")
	}
	return nil
}

func (descriptor StorageChunkDescriptor) ValidateFormat(params StorageChunkParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validateStorageHash("storage chunk hash", descriptor.ChunkHash); err != nil {
		return err
	}
	if descriptor.ChunkSize == 0 || descriptor.ChunkSize > params.MaxChunkBytes {
		return fmt.Errorf("storage chunk size must be between 1 and %d", params.MaxChunkBytes)
	}
	if err := validateStorageHash("storage chunk proof root", descriptor.ChunkProofRoot); err != nil {
		return err
	}
	if descriptor.ErasureGroupOptional != "" {
		if len(descriptor.ErasureGroupOptional) > MaxStorageErasureGroupBytes {
			return fmt.Errorf("storage erasure group must be <= %d bytes", MaxStorageErasureGroupBytes)
		}
		if err := validateStorageToken("storage erasure group", descriptor.ErasureGroupOptional); err != nil {
			return err
		}
	}
	if descriptor.DescriptorHash != "" {
		return validateStorageHash("storage chunk descriptor hash", descriptor.DescriptorHash)
	}
	return nil
}

func (descriptor StorageChunkDescriptor) Validate(params StorageChunkParams) error {
	if err := descriptor.ValidateFormat(params); err != nil {
		return err
	}
	if descriptor.ChunkProofRoot != ComputeStorageChunkProofRoot(descriptor.ChunkIndex, descriptor.ChunkHash, descriptor.ChunkSize, descriptor.ErasureGroupOptional) {
		return errors.New("storage chunk proof root mismatch")
	}
	if descriptor.DescriptorHash == "" {
		return errors.New("storage chunk descriptor hash is required")
	}
	if descriptor.DescriptorHash != ComputeStorageChunkDescriptorHash(descriptor) {
		return errors.New("storage chunk descriptor hash mismatch")
	}
	return nil
}

func (proof StorageChunkInclusionProof) ValidateFormat() error {
	if err := validateStorageToken("storage chunk inclusion proof object id", proof.ObjectID); err != nil {
		return err
	}
	if err := validateStorageHash("storage chunk inclusion proof content hash", proof.ContentHash); err != nil {
		return err
	}
	if err := validateStorageHash("storage chunk inclusion proof object root", proof.ObjectRoot); err != nil {
		return err
	}
	if err := validateStorageHash("storage chunk inclusion proof chunk hash", proof.ChunkHash); err != nil {
		return err
	}
	if err := validateStorageHash("storage chunk inclusion proof root", proof.ChunkProofRoot); err != nil {
		return err
	}
	for _, item := range proof.ProofPath {
		if err := validateStorageHash("storage chunk inclusion proof path", item); err != nil {
			return err
		}
	}
	if proof.ProofHash != "" {
		return validateStorageHash("storage chunk inclusion proof hash", proof.ProofHash)
	}
	return nil
}

func (proof StorageChunkInclusionProof) Validate() error {
	if err := proof.ValidateFormat(); err != nil {
		return err
	}
	if proof.ProofHash == "" {
		return errors.New("storage chunk inclusion proof hash is required")
	}
	if proof.ProofHash != ComputeStorageChunkInclusionProofHash(proof) {
		return errors.New("storage chunk inclusion proof hash mismatch")
	}
	return nil
}

func ComputeStorageChunkProofRoot(chunkIndex uint32, chunkHash string, chunkSize uint64, erasureGroup string) string {
	return storageHashParts("storage-chunk-proof-root-v1", fmt.Sprintf("%020d", chunkIndex), chunkHash, fmt.Sprintf("%020d", chunkSize), erasureGroup)
}

func ComputeStorageChunkDescriptorHash(descriptor StorageChunkDescriptor) string {
	return storageHashParts(
		"storage-chunk-descriptor-v1",
		fmt.Sprintf("%020d", descriptor.ChunkIndex),
		descriptor.ChunkHash,
		fmt.Sprintf("%020d", descriptor.ChunkSize),
		descriptor.ChunkProofRoot,
		descriptor.ErasureGroupOptional,
	)
}

func ComputeStorageContentHashFromChunks(descriptors []StorageChunkDescriptor) string {
	ordered := append([]StorageChunkDescriptor(nil), descriptors...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ChunkIndex < ordered[j].ChunkIndex })
	parts := []string{"storage-content-hash-ordered-chunks-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, descriptor := range ordered {
		parts = append(parts, descriptor.ChunkProofRoot)
	}
	return storageHashParts(parts...)
}

func ComputeStorageChunkInclusionProofHash(proof StorageChunkInclusionProof) string {
	parts := []string{
		"storage-chunk-inclusion-proof-v1",
		proof.ObjectID,
		proof.ContentHash,
		proof.ObjectRoot,
		fmt.Sprintf("%020d", proof.ChunkIndex),
		proof.ChunkHash,
		proof.ChunkProofRoot,
		fmt.Sprintf("%020d", len(proof.ProofPath)),
	}
	parts = append(parts, proof.ProofPath...)
	return storageHashParts(parts...)
}

func validateStorageChunkDescriptors(descriptors []StorageChunkDescriptor, params StorageChunkParams) error {
	if len(descriptors) == 0 {
		return errors.New("storage chunk descriptors are required")
	}
	if len(descriptors) > MaxStorageChunkRoots {
		return fmt.Errorf("storage chunk descriptors must not exceed %d", MaxStorageChunkRoots)
	}
	ordered := append([]StorageChunkDescriptor(nil), descriptors...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ChunkIndex < ordered[j].ChunkIndex })
	for i, descriptor := range ordered {
		if err := descriptor.Validate(params); err != nil {
			return err
		}
		if descriptor.ChunkIndex != uint32(i) {
			return errors.New("storage chunk descriptor sequence gap")
		}
	}
	return nil
}

func storageChunkDescriptorRoots(descriptors []StorageChunkDescriptor) []string {
	ordered := append([]StorageChunkDescriptor(nil), descriptors...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ChunkIndex < ordered[j].ChunkIndex })
	out := make([]string, 0, len(ordered))
	for _, descriptor := range ordered {
		out = append(out, descriptor.ChunkProofRoot)
	}
	return out
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
