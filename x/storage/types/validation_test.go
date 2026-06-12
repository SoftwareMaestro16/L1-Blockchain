package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageValidationRulesAcceptValidStateFeeAndProof(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "validation")
	proof := testStorageChunkProof(t, object, chunks, 0)
	receipt := testStorageReceiptWithProof(t, object, proof, 31)
	state, err := BuildStorageStateV2(
		[]StorageObject{object},
		testStorageChunkRecords(object, chunks),
		[]StorageAccessReceipt{receipt},
		[]ReplicationStatusCommitment{testStorageReplication(t, object, 30)},
		40,
	)
	require.NoError(t, err)

	require.NoError(t, ValidateStorageStateV2Rules(state, DefaultStorageValidationParams()))
	require.NoError(t, ValidateStorageRetrievalWithProof(state, object.ObjectID, proof, DefaultStorageValidationParams()))

	quote, err := NewStorageFeeQuote("alice", object, chunks, 1, DefaultStorageFeeParams())
	require.NoError(t, err)
	require.Equal(t, DefaultStorageFeeDenom, quote.Denom)
	require.Equal(t, object.Size, quote.ObjectBytes)
	require.Equal(t, object.Size, quote.ChunkBytes)
	require.Equal(t, ComputeStorageFeeQuoteHash(quote), quote.QuoteHash)
}

func TestStorageValidationRejectsContentSizeAndPolicyMismatches(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "size")
	object.Size = object.Size + 1
	object.ObjectHash = ComputeStorageObjectHash(object)
	_, err := BuildStorageStateV2(
		[]StorageObject{object},
		testStorageChunkRecords(object, chunks),
		nil,
		[]ReplicationStatusCommitment{testStorageReplication(t, object, 20)},
		30,
	)
	require.ErrorContains(t, err, "size")

	object, chunks = testStorageStateObjectWithChunks(t, "alice", "policy")
	params := DefaultStorageValidationParams()
	params.AllowedAccessPolicies = []string{AccessPolicyPrivate}
	err = ValidateStorageObjectAgainstChunks(object, chunks, params)
	require.ErrorContains(t, err, "access policy")
}

func TestStorageValidationRejectsInvalidReplicationParameters(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "replication")
	object.ReplicationPolicy = ReplicationPolicySingle
	object.ObjectHash = ComputeStorageObjectHash(object)
	replication, err := NewReplicationStatusCommitment(ReplicationStatusCommitment{
		ObjectID:		object.ObjectID,
		ReplicationPolicy:	ReplicationPolicySingle,
		StorageClass:		object.StorageClass,
		ReplicaCount:		2,
		AvailabilityBps:	9000,
		LastVerifiedHeight:	20,
	})
	require.NoError(t, err)
	state, err := BuildStorageStateV2(
		[]StorageObject{object},
		testStorageChunkRecords(object, chunks),
		nil,
		[]ReplicationStatusCommitment{replication},
		30,
	)
	require.NoError(t, err)

	err = ValidateStorageStateV2Rules(state, DefaultStorageValidationParams())
	require.ErrorContains(t, err, "replica count")
}

func TestStorageValidationRejectsUnregisteredReceiptAndUnbackedRetrieval(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "receipt")
	receipt := testStorageStateReceipt(t, object, 31)
	state, err := BuildStorageStateV2(
		[]StorageObject{object},
		testStorageChunkRecords(object, chunks),
		[]StorageAccessReceipt{receipt},
		[]ReplicationStatusCommitment{testStorageReplication(t, object, 30)},
		40,
	)
	require.NoError(t, err)
	err = ValidateStorageStateV2Rules(state, DefaultStorageValidationParams())
	require.ErrorContains(t, err, "proof-backed")

	unknown := receipt
	unknown.ObjectID = "object/missing"
	unknown.ReceiptID = ComputeStorageAccessReceiptID(unknown)
	unknown.ReceiptHash = ComputeStorageAccessReceiptHash(unknown)
	err = ValidateStorageReceiptReferencesRegisteredObject(unknown, state)
	require.ErrorContains(t, err, "unregistered")
}

func TestLazyFetchBoundaryStaysOutsideConsensus(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "lazy")
	proof := testStorageChunkProof(t, object, chunks, 1)
	request, err := NewLazyFetchRequest(LazyFetchRequest{
		ObjectID:		object.ObjectID,
		ContentHash:		object.ContentHash,
		ChunkIndexOptional:	proof.ChunkIndex,
		ChunkProofRoot:		proof.ChunkProofRoot,
		Requester:		"service/alice",
		MaxBytes:		1024,
		RequestHeight:		50,
	})
	require.NoError(t, err)
	result, err := NewLazyFetchResultBoundary(LazyFetchResultBoundary{
		RequestHash:	request.RequestHash,
		Provider:	"provider/one",
		PayloadHash:	proof.ChunkHash,
		Proof:		proof,
		ResultHeight:	51,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateLazyFetchBoundaryOutsideConsensus(request, result, object))

	broken := result
	broken.Provider = "provider/two"
	require.ErrorContains(t, ValidateLazyFetchBoundaryOutsideConsensus(request, broken, object), "result hash")
}

func testStorageChunkProof(t *testing.T, object StorageObject, chunks []StorageChunkDescriptor, index uint32) StorageChunkInclusionProof {
	t.Helper()
	proofPath := make([]string, 0, len(chunks)-1)
	for _, chunk := range chunks {
		if chunk.ChunkIndex != index {
			proofPath = append(proofPath, chunk.ChunkProofRoot)
		}
	}
	chunk := chunks[index]
	proof, err := NewStorageChunkInclusionProof(StorageChunkInclusionProof{
		ObjectID:	object.ObjectID,
		ContentHash:	object.ContentHash,
		ObjectRoot:	object.ObjectHash,
		ChunkIndex:	chunk.ChunkIndex,
		ChunkHash:	chunk.ChunkHash,
		ChunkProofRoot:	chunk.ChunkProofRoot,
		ProofPath:	proofPath,
	})
	require.NoError(t, err)
	return proof
}

func testStorageReceiptWithProof(t *testing.T, object StorageObject, proof StorageChunkInclusionProof, height uint64) StorageAccessReceipt {
	t.Helper()
	receipt, err := NewStorageAccessReceipt(StorageAccessReceipt{
		ObjectID:	object.ObjectID,
		Accessor:	object.Owner,
		AccessType:	"read",
		AccessHeight:	height,
		ContentHash:	object.ContentHash,
		ChunkRoot:	ComputeStorageChunkRoot(object.ChunkRoots),
		PolicyHash:	ComputeStoragePolicyHash(object.ReplicationPolicy, object.AccessPolicy, object.StorageClass),
		RetrievalProof:	proof.ProofHash,
	})
	require.NoError(t, err)
	return receipt
}
