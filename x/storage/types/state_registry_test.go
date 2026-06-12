package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageStateV2KeysMatchSpec(t *testing.T) {
	object := testStorageStateObject(t, "alice", "a")
	objectKey, err := StorageObjectKey(object.ObjectID)
	require.NoError(t, err)
	require.Equal(t, "storage/objects/"+object.ObjectID, objectKey)

	contentKey, err := StorageContentIndexKey(object.ContentHash)
	require.NoError(t, err)
	require.Equal(t, "storage/content/"+object.ContentHash, contentKey)

	chunkKey, err := StorageChunkDescriptorKey(object.ObjectID, 3)
	require.NoError(t, err)
	require.Equal(t, "storage/chunks/"+object.ObjectID+"/00000000000000000003", chunkKey)

	ownerKey, err := StorageOwnerIndexKey(object.Owner, object.ObjectID)
	require.NoError(t, err)
	require.Equal(t, "storage/owner_index/"+object.Owner+"/"+object.ObjectID, ownerKey)

	accessKey, err := StorageAccessReceiptKey(object.ObjectID, "receipt/alice")
	require.NoError(t, err)
	require.Equal(t, "storage/access/"+object.ObjectID+"/receipt/alice", accessKey)

	replicationKey, err := StorageReplicationKey(object.ObjectID)
	require.NoError(t, err)
	require.Equal(t, "storage/replication/"+object.ObjectID, replicationKey)

	rootKey, err := StorageRootKey(9)
	require.NoError(t, err)
	require.Equal(t, "storage/root/00000000000000000009", rootKey)
}

func TestStorageStateV2BuildsDeterministicIndexesAndRoot(t *testing.T) {
	first, firstChunks := testStorageStateObjectWithChunks(t, "alice", "a")
	second, secondChunks := testStorageStateObjectWithChunks(t, "bob", "b")
	firstReceipt := testStorageStateReceipt(t, first, 21)
	secondReceipt := testStorageStateReceipt(t, second, 22)
	firstReplication := testStorageReplication(t, first, 21)
	secondReplication := testStorageReplication(t, second, 22)

	left, err := BuildStorageStateV2(
		[]StorageObject{second, first},
		append(testStorageChunkRecords(first, firstChunks), testStorageChunkRecords(second, secondChunks)...),
		[]StorageAccessReceipt{secondReceipt, firstReceipt},
		[]ReplicationStatusCommitment{secondReplication, firstReplication},
		30,
	)
	require.NoError(t, err)
	right, err := BuildStorageStateV2(
		[]StorageObject{first, second},
		append(testStorageChunkRecords(second, secondChunks), testStorageChunkRecords(first, firstChunks)...),
		[]StorageAccessReceipt{firstReceipt, secondReceipt},
		[]ReplicationStatusCommitment{firstReplication, secondReplication},
		30,
	)
	require.NoError(t, err)

	require.Equal(t, left.Root.StateRoot, right.Root.StateRoot)
	require.Len(t, left.ContentIndex, 2)
	require.Len(t, left.OwnerIndex, 2)
	require.ElementsMatch(t, []string{first.ObjectID, second.ObjectID}, []string{left.ContentIndex[0].Value, left.ContentIndex[1].Value})
	require.NoError(t, left.Validate())
}

func TestStorageStateV2MessagesMutateState(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "a")
	replication := testStorageReplication(t, object, 10)
	state, err := BuildStorageStateV2(nil, nil, nil, nil, 1)
	require.NoError(t, err)

	state, err = RegisterStorageObjectInStateV2(state, MsgRegisterStorageObject{
		Authority:	"alice",
		Object:		object,
		Chunks:		chunks,
		Replication:	replication,
		Height:		10,
	}, DefaultStorageChunkParams())
	require.NoError(t, err)
	require.Len(t, state.Objects, 1)

	updatedReplication, err := NewReplicationStatusCommitment(ReplicationStatusCommitment{
		ObjectID:		object.ObjectID,
		ReplicationPolicy:	ReplicationPolicyRegional,
		StorageClass:		StorageClassHot,
		ReplicaCount:		2,
		AvailabilityBps:	9900,
		LastVerifiedHeight:	11,
	})
	require.NoError(t, err)
	state, err = UpdateStoragePolicyInStateV2(state, MsgUpdateStoragePolicy{
		Authority:		"alice",
		ObjectID:		object.ObjectID,
		ReplicationPolicy:	ReplicationPolicyRegional,
		AccessPolicy:		AccessPolicyPublicRead,
		StorageClass:		StorageClassHot,
		Replication:		updatedReplication,
		Height:			11,
	})
	require.NoError(t, err)
	updated, found := QueryStorageObject(state, object.ObjectID)
	require.True(t, found)
	require.Equal(t, AccessPolicyPublicRead, updated.AccessPolicy)

	state, err = RenewStorageObjectInStateV2(state, MsgRenewStorageObject{
		Authority:	"alice",
		ObjectID:	object.ObjectID,
		ExpiresHeight:	500,
		Height:		12,
	})
	require.NoError(t, err)
	updated, found = QueryStorageObject(state, object.ObjectID)
	require.True(t, found)
	require.Equal(t, uint64(500), updated.ExpiresHeightOptional)

	receipt := testStorageStateReceipt(t, updated, 13)
	state, err = SubmitStorageReceiptInStateV2(state, MsgSubmitStorageReceipt{
		Authority:	"alice",
		Receipt:	receipt,
		Height:		13,
	})
	require.NoError(t, err)

	proof, err := NewStorageChunkInclusionProof(StorageChunkInclusionProof{
		ObjectID:	updated.ObjectID,
		ContentHash:	updated.ContentHash,
		ObjectRoot:	updated.ObjectHash,
		ChunkIndex:	chunks[0].ChunkIndex,
		ChunkHash:	chunks[0].ChunkHash,
		ChunkProofRoot:	chunks[0].ChunkProofRoot,
		ProofPath:	[]string{chunks[1].ChunkProofRoot},
	})
	require.NoError(t, err)
	stateProof, err := VerifyStorageProofInStateV2(state, MsgVerifyStorageProof{
		Authority:	"alice",
		ObjectID:	updated.ObjectID,
		Proof:		proof,
		Height:		14,
	})
	require.NoError(t, err)
	require.Equal(t, state.Root.StateRoot, stateProof.Root)

	state, err = DeleteStorageObjectInStateV2(state, MsgDeleteStorageObject{
		Authority:	"alice",
		ObjectID:	object.ObjectID,
		Height:		15,
	})
	require.NoError(t, err)
	require.Empty(t, state.Objects)
	require.Empty(t, state.Chunks)
	require.Empty(t, state.AccessReceipts)
}

func TestStorageStateV2QueriesAndProof(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "a")
	receipt := testStorageStateReceipt(t, object, 21)
	state, err := BuildStorageStateV2(
		[]StorageObject{object},
		testStorageChunkRecords(object, chunks),
		[]StorageAccessReceipt{receipt},
		[]ReplicationStatusCommitment{testStorageReplication(t, object, 20)},
		30,
	)
	require.NoError(t, err)

	got, found := QueryStorageObject(state, object.ObjectID)
	require.True(t, found)
	require.Equal(t, object.ObjectHash, got.ObjectHash)

	got, found = QueryObjectByContentHash(state, object.ContentHash)
	require.True(t, found)
	require.Equal(t, object.ObjectID, got.ObjectID)

	chunk, found := QueryChunkDescriptor(state, object.ObjectID, 1)
	require.True(t, found)
	require.Equal(t, chunks[1].DescriptorHash, chunk.DescriptorHash)

	owned := QueryStorageObjectsByOwner(state, "alice")
	require.Len(t, owned, 1)
	require.Equal(t, object.ObjectID, owned[0].ObjectID)

	storedReceipt, found := QueryStorageAccessReceipt(state, object.ObjectID, receipt.ReceiptID)
	require.True(t, found)
	require.Equal(t, receipt.ReceiptHash, storedReceipt.ReceiptHash)

	root, found := QueryStorageRoot(state, 30)
	require.True(t, found)
	require.Equal(t, state.Root.StateRoot, root.StateRoot)

	key, err := StorageAccessReceiptKey(object.ObjectID, receipt.ReceiptID)
	require.NoError(t, err)
	proof, err := QueryStorageProof(state, key)
	require.NoError(t, err)
	require.Equal(t, receipt.ReceiptHash, proof.ValueHash)
	require.Equal(t, state.Root.StateRoot, proof.Root)
}

func TestStorageStateV2RejectsUnauthorizedPolicyUpdate(t *testing.T) {
	object, chunks := testStorageStateObjectWithChunks(t, "alice", "a")
	state, err := BuildStorageStateV2(
		[]StorageObject{object},
		testStorageChunkRecords(object, chunks),
		nil,
		[]ReplicationStatusCommitment{testStorageReplication(t, object, 10)},
		20,
	)
	require.NoError(t, err)
	replication, err := NewReplicationStatusCommitment(ReplicationStatusCommitment{
		ObjectID:		object.ObjectID,
		ReplicationPolicy:	ReplicationPolicySingle,
		StorageClass:		StorageClassCold,
		ReplicaCount:		1,
		AvailabilityBps:	9000,
		LastVerifiedHeight:	21,
	})
	require.NoError(t, err)

	_, err = UpdateStoragePolicyInStateV2(state, MsgUpdateStoragePolicy{
		Authority:		"mallory",
		ObjectID:		object.ObjectID,
		ReplicationPolicy:	ReplicationPolicySingle,
		AccessPolicy:		AccessPolicyPrivate,
		StorageClass:		StorageClassCold,
		Replication:		replication,
		Height:			21,
	})
	require.ErrorContains(t, err, "authority")
}

func testStorageStateObject(t *testing.T, owner, suffix string) StorageObject {
	t.Helper()
	object, _ := testStorageStateObjectWithChunks(t, owner, suffix)
	return object
}

func testStorageStateObjectWithChunks(t *testing.T, owner, suffix string) (StorageObject, []StorageChunkDescriptor) {
	t.Helper()
	params := DefaultStorageChunkParams()
	chunks, _, err := BuildStorageChunkDescriptors(
		[]string{storageTestHash(owner + "/" + suffix + "/chunk-0"), storageTestHash(owner + "/" + suffix + "/chunk-1")},
		[]uint64{512, 256},
		nil,
		params,
	)
	require.NoError(t, err)
	object, err := BuildStorageObjectFromChunkDescriptors(StorageObject{
		Size:			768,
		ReplicationPolicy:	ReplicationPolicyMultiZone,
		AccessPolicy:		AccessPolicyPermissioned,
		Owner:			owner,
		StorageClass:		StorageClassWarm,
		CreatedHeight:		10,
		ExpiresHeightOptional:	100,
		MetadataHashOptional:	storageTestHash(owner + "/" + suffix + "/metadata"),
		AvailabilityCommitment:	storageTestHash(owner + "/" + suffix + "/availability"),
	}, chunks, params)
	require.NoError(t, err)
	return object, chunks
}

func testStorageChunkRecords(object StorageObject, chunks []StorageChunkDescriptor) []StorageChunkDescriptorRecord {
	records := make([]StorageChunkDescriptorRecord, 0, len(chunks))
	for _, chunk := range chunks {
		record := StorageChunkDescriptorRecord{ObjectID: object.ObjectID, Descriptor: chunk}
		record.RecordHash = ComputeStorageChunkDescriptorRecordHash(record)
		records = append(records, record)
	}
	return records
}

func testStorageReplication(t *testing.T, object StorageObject, height uint64) ReplicationStatusCommitment {
	t.Helper()
	replication, err := NewReplicationStatusCommitment(ReplicationStatusCommitment{
		ObjectID:		object.ObjectID,
		ReplicationPolicy:	object.ReplicationPolicy,
		StorageClass:		object.StorageClass,
		ReplicaCount:		3,
		AvailabilityBps:	9950,
		LastVerifiedHeight:	height,
	})
	require.NoError(t, err)
	return replication
}

func testStorageStateReceipt(t *testing.T, object StorageObject, height uint64) StorageAccessReceipt {
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
