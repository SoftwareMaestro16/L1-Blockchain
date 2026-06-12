package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageChunkDescriptorsCommitContentHashToOrderedRoots(t *testing.T) {
	params := StorageChunkParams{MaxChunkBytes: 1024}
	chunks, contentHash, err := BuildStorageChunkDescriptors(
		[]string{storageTestHash("chunk-0"), storageTestHash("chunk-1")},
		[]uint64{512, 256},
		[]string{"rs-0", "rs-0"},
		params,
	)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	require.Equal(t, uint32(0), chunks[0].ChunkIndex)
	require.Equal(t, ComputeStorageContentHashFromChunks(chunks), contentHash)

	reordered := []StorageChunkDescriptor{chunks[1], chunks[0]}
	require.Equal(t, contentHash, ComputeStorageContentHashFromChunks(reordered))

	swapped := append([]StorageChunkDescriptor(nil), chunks...)
	swapped[0], swapped[1] = swapped[1], swapped[0]
	swapped[0].ChunkIndex = 0
	swapped[1].ChunkIndex = 1
	swapped[0].ChunkProofRoot = ComputeStorageChunkProofRoot(swapped[0].ChunkIndex, swapped[0].ChunkHash, swapped[0].ChunkSize, swapped[0].ErasureGroupOptional)
	swapped[1].ChunkProofRoot = ComputeStorageChunkProofRoot(swapped[1].ChunkIndex, swapped[1].ChunkHash, swapped[1].ChunkSize, swapped[1].ErasureGroupOptional)
	swapped[0].DescriptorHash = ComputeStorageChunkDescriptorHash(swapped[0])
	swapped[1].DescriptorHash = ComputeStorageChunkDescriptorHash(swapped[1])
	require.NotEqual(t, contentHash, ComputeStorageContentHashFromChunks(swapped))
}

func TestStorageChunkSizeBoundedByParams(t *testing.T) {
	params := StorageChunkParams{MaxChunkBytes: 128}
	_, _, err := BuildStorageChunkDescriptors([]string{storageTestHash("chunk")}, []uint64{129}, nil, params)
	require.ErrorContains(t, err, "chunk size")

	_, _, err = BuildStorageChunkDescriptors([]string{storageTestHash("chunk")}, []uint64{128}, nil, params)
	require.NoError(t, err)
}

func TestStorageObjectBuiltFromChunksAndInclusionProofVerifiesObjectRoot(t *testing.T) {
	params := StorageChunkParams{MaxChunkBytes: 1024}
	chunks, _, err := BuildStorageChunkDescriptors(
		[]string{storageTestHash("chunk-0"), storageTestHash("chunk-1")},
		[]uint64{512, 256},
		nil,
		params,
	)
	require.NoError(t, err)

	object, err := BuildStorageObjectFromChunkDescriptors(StorageObject{
		Size:			768,
		ReplicationPolicy:	ReplicationPolicyMultiZone,
		AccessPolicy:		AccessPolicyPermissioned,
		Owner:			"alice",
		StorageClass:		StorageClassHot,
		CreatedHeight:		10,
		AvailabilityCommitment:	storageTestHash("availability"),
	}, chunks, params)
	require.NoError(t, err)
	require.NoError(t, VerifyStorageObjectContentHash(object, chunks, params))

	proof, err := NewStorageChunkInclusionProof(StorageChunkInclusionProof{
		ObjectID:	object.ObjectID,
		ContentHash:	object.ContentHash,
		ObjectRoot:	object.ObjectHash,
		ChunkIndex:	chunks[0].ChunkIndex,
		ChunkHash:	chunks[0].ChunkHash,
		ChunkProofRoot:	chunks[0].ChunkProofRoot,
		ProofPath:	[]string{chunks[1].ChunkProofRoot},
	})
	require.NoError(t, err)
	require.NoError(t, VerifyStorageChunkInclusionInObject(proof, object))

	proof.ChunkProofRoot = storageTestHash("not-in-object")
	proof.ProofHash = ComputeStorageChunkInclusionProofHash(proof)
	err = VerifyStorageChunkInclusionInObject(proof, object)
	require.ErrorContains(t, err, "not in object root")
}

func TestStorageChunkModelRejectsSequenceGapsAndContentMismatch(t *testing.T) {
	params := StorageChunkParams{MaxChunkBytes: 1024}
	chunks, _, err := BuildStorageChunkDescriptors(
		[]string{storageTestHash("chunk-0"), storageTestHash("chunk-1")},
		[]uint64{512, 256},
		nil,
		params,
	)
	require.NoError(t, err)
	chunks[1].ChunkIndex = 3
	chunks[1].ChunkProofRoot = ComputeStorageChunkProofRoot(chunks[1].ChunkIndex, chunks[1].ChunkHash, chunks[1].ChunkSize, "")
	chunks[1].DescriptorHash = ComputeStorageChunkDescriptorHash(chunks[1])
	err = validateStorageChunkDescriptors(chunks, params)
	require.ErrorContains(t, err, "sequence gap")

	chunks, _, err = BuildStorageChunkDescriptors(
		[]string{storageTestHash("chunk-0"), storageTestHash("chunk-1")},
		[]uint64{512, 256},
		nil,
		params,
	)
	require.NoError(t, err)
	object := testStorageObject(t, "alice", []string{storageTestHash("different")})
	err = VerifyStorageObjectContentHash(object, chunks, params)
	require.ErrorContains(t, err, "content_hash")
}

func TestUnrelatedStorageTransitionRequiresCommitmentsOnly(t *testing.T) {
	params := StorageChunkParams{MaxChunkBytes: 1024}
	chunks, _, err := BuildStorageChunkDescriptors([]string{storageTestHash("chunk-0")}, []uint64{512}, nil, params)
	require.NoError(t, err)
	object, err := BuildStorageObjectFromChunkDescriptors(StorageObject{
		Size:			512,
		ReplicationPolicy:	ReplicationPolicySingle,
		AccessPolicy:		AccessPolicyPrivate,
		Owner:			"alice",
		StorageClass:		StorageClassCold,
		CreatedHeight:		10,
		AvailabilityCommitment:	storageTestHash("availability"),
	}, chunks, params)
	require.NoError(t, err)
	require.NoError(t, ValidateUnrelatedStorageTransitionDoesNotRequireChunks(object))
}
