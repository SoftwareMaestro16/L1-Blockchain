package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageObjectCanonicalSchemaAndHash(t *testing.T) {
	object := testStorageObject(t, "alice", []string{storageTestHash("chunk-b"), storageTestHash("chunk-a")})

	require.NoError(t, object.Validate())
	require.Equal(t, []string{storageTestHash("chunk-a"), storageTestHash("chunk-b")}, object.ChunkRoots)
	require.NotEmpty(t, object.ObjectID)
	require.Equal(t, StorageObjectVersionV1, object.Version)
	require.Equal(t, ComputeStorageObjectHash(object), object.ObjectHash)
	require.Equal(t, ComputeStorageChunkRoot(object.ChunkRoots), ComputeStorageChunkRoot([]string{storageTestHash("chunk-b"), storageTestHash("chunk-a")}))
}

func TestStorageObjectRejectsLargePayloadFieldsAndBadPolicies(t *testing.T) {
	object := testStorageObject(t, "alice", []string{storageTestHash("chunk-a")})
	object.ChunkRoots = []string{storageTestHash("chunk-a"), storageTestHash("chunk-a")}
	object.ObjectHash = ""
	_, err := NewStorageObject(object)
	require.NoError(t, err)

	object = testStorageObject(t, "alice", []string{storageTestHash("chunk-a")})
	object.ReplicationPolicy = "external_call"
	object.ObjectHash = ""
	_, err = NewStorageObject(object)
	require.ErrorContains(t, err, "replication policy")

	object = testStorageObject(t, "alice", []string{storageTestHash("chunk-a")})
	object.ExpiresHeightOptional = object.CreatedHeight
	object.ObjectHash = ""
	_, err = NewStorageObject(object)
	require.ErrorContains(t, err, "expiry")
}

func TestStorageChunkSetRetrievalProofAndReceipt(t *testing.T) {
	object := testStorageObject(t, "alice", []string{storageTestHash("chunk-a"), storageTestHash("chunk-b")})
	chunkSet, err := NewStorageChunkSet(object.ObjectID, object.ChunkRoots)
	require.NoError(t, err)
	require.Equal(t, uint32(2), chunkSet.ChunkCount)
	require.Equal(t, object.ChunkRoots, chunkSet.ChunkRoots)

	proof, err := NewStorageRetrievalProof(StorageRetrievalProof{
		ObjectID:	object.ObjectID,
		ContentHash:	object.ContentHash,
		ChunkRoot:	chunkSet.ChunkRoot,
		ChunkIndex:	0,
		ChunkHash:	object.ChunkRoots[0],
		ProofPath:	[]string{object.ChunkRoots[1]},
	})
	require.NoError(t, err)
	require.Equal(t, ComputeStorageRetrievalProofHash(proof), proof.ProofHash)

	receipt, err := NewStorageAccessReceipt(StorageAccessReceipt{
		ObjectID:	object.ObjectID,
		Accessor:	"service/alice",
		AccessType:	"read",
		AccessHeight:	12,
		ContentHash:	object.ContentHash,
		ChunkRoot:	chunkSet.ChunkRoot,
		PolicyHash:	ComputeStoragePolicyHash(object.ReplicationPolicy, object.AccessPolicy, object.StorageClass),
		RetrievalProof:	proof.ProofHash,
	})
	require.NoError(t, err)
	require.NotEmpty(t, receipt.ReceiptID)
	require.Equal(t, ComputeStorageAccessReceiptHash(receipt), receipt.ReceiptHash)
}

func TestStorageObjectCommitmentStateDeterministicRoot(t *testing.T) {
	first := testStorageObject(t, "alice", []string{storageTestHash("chunk-a")})
	second := testStorageObject(t, "bob", []string{storageTestHash("chunk-b")})
	firstReceipt := testStorageReceipt(t, first, 12)
	secondReceipt := testStorageReceipt(t, second, 13)

	left, err := BuildStorageObjectCommitmentState([]StorageObject{second, first}, []StorageAccessReceipt{secondReceipt, firstReceipt}, 20)
	require.NoError(t, err)
	right, err := BuildStorageObjectCommitmentState([]StorageObject{first, second}, []StorageAccessReceipt{firstReceipt, secondReceipt}, 20)
	require.NoError(t, err)

	require.Equal(t, left.RootHash, right.RootHash)
	require.Equal(t, ComputeStorageObjectRoot(left.Objects), ComputeStorageObjectRoot(right.Objects))
	require.Equal(t, ComputeStorageAccessReceiptRoot(left.Receipts), ComputeStorageAccessReceiptRoot(right.Receipts))
}

func TestStorageObjectCommitmentStateRejectsTamperedReceipt(t *testing.T) {
	object := testStorageObject(t, "alice", []string{storageTestHash("chunk-a")})
	receipt := testStorageReceipt(t, object, 12)
	state, err := BuildStorageObjectCommitmentState([]StorageObject{object}, []StorageAccessReceipt{receipt}, 20)
	require.NoError(t, err)

	state.Receipts[0].AccessType = "write"
	state.Receipts[0].ReceiptHash = ComputeStorageAccessReceiptHash(state.Receipts[0])
	state.RootHash = ComputeStorageObjectCommitmentStateRoot(state)
	require.NoError(t, state.Validate())

	state.Receipts[0].ReceiptHash = storageTestHash("wrong")
	err = state.Validate()
	require.ErrorContains(t, err, "receipt hash mismatch")
}

func testStorageObject(t *testing.T, owner string, chunks []string) StorageObject {
	t.Helper()
	object, err := NewStorageObject(StorageObject{
		ContentHash:		storageTestHash(owner + "/content"),
		ChunkRoots:		chunks,
		Size:			1024,
		ReplicationPolicy:	ReplicationPolicyMultiZone,
		AccessPolicy:		AccessPolicyPermissioned,
		Owner:			owner,
		StorageClass:		StorageClassWarm,
		CreatedHeight:		10,
		ExpiresHeightOptional:	100,
		MetadataHashOptional:	storageTestHash(owner + "/metadata"),
		AvailabilityCommitment:	storageTestHash(owner + "/availability"),
	})
	require.NoError(t, err)
	return object
}

func testStorageReceipt(t *testing.T, object StorageObject, height uint64) StorageAccessReceipt {
	t.Helper()
	receipt, err := NewStorageAccessReceipt(StorageAccessReceipt{
		ObjectID:	object.ObjectID,
		Accessor:	object.Owner,
		AccessType:	"read",
		AccessHeight:	height,
		ContentHash:	object.ContentHash,
		ChunkRoot:	ComputeStorageChunkRoot(object.ChunkRoots),
		PolicyHash:	ComputeStoragePolicyHash(object.ReplicationPolicy, object.AccessPolicy, object.StorageClass),
	})
	require.NoError(t, err)
	return receipt
}

func storageTestHash(value string) string {
	return storageHashParts("test", value)
}
