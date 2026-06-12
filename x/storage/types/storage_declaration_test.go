package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageDeclarationModelsMatchServiceSpec(t *testing.T) {
	cases := []StorageDeclaration{
		{
			StorageModel:		StorageModelEphemeral,
			ContentHashOptional:	storageDeclarationTestHash("ephemeral/content"),
			RetrievalMethod:	StorageRetrievalInline,
			VerificationMethod:	StorageVerificationContentHash,
			RetentionPolicy:	StorageRetentionNone,
			AccessPolicy:		AccessPolicyPrivate,
			MaxPayloadBytes:	1024,
		},
		{
			StorageModel:		StorageModelPersistentOnChain,
			StateRootOptional:	storageDeclarationTestHash("on-chain/root"),
			RetrievalMethod:	StorageRetrievalOnChainState,
			VerificationMethod:	StorageVerificationStateRoot,
			RetentionPolicy:	StorageRetentionPermanent,
			AccessPolicy:		AccessPolicyPermissioned,
			MaxPayloadBytes:	DefaultMaxStateBytes,
		},
		{
			StorageModel:		StorageModelDistributedOffChain,
			ContentHashOptional:	storageDeclarationTestHash("distributed/content"),
			RetrievalMethod:	StorageRetrievalContentAddressed,
			VerificationMethod:	StorageVerificationChunkProof,
			RetentionPolicy:	StorageRetentionExpiry,
			AccessPolicy:		AccessPolicyPublicRead,
			MaxPayloadBytes:	1 << 20,
		},
		{
			StorageModel:			StorageModelHybrid,
			ContentHashOptional:		storageDeclarationTestHash("hybrid/content"),
			StateRootOptional:		storageDeclarationTestHash("hybrid/root"),
			ContentLocationOptional:	"ipfs/bafy-storage-object",
			RetrievalMethod:		StorageRetrievalHybridEndpoint,
			VerificationMethod:		StorageVerificationHybridCommitment,
			RetentionPolicy:		StorageRetentionHeight,
			AccessPolicy:			AccessPolicyPermissioned,
			AccessReceiptOptional:		storageDeclarationTestHash("hybrid/receipt"),
			MaxPayloadBytes:		1 << 22,
		},
	}

	for _, tc := range cases {
		declaration, err := NewStorageDeclaration(tc)
		require.NoError(t, err, tc.StorageModel)
		require.Equal(t, tc.StorageModel, declaration.StorageModel)
		require.NotEmpty(t, declaration.DeclarationHash)
		require.NoError(t, declaration.Validate())
	}
}

func TestHybridStorageDeclarationDoesNotRequireOffchainPayloadForConsensus(t *testing.T) {
	object := storageDeclarationTestObject(t)
	declaration, err := NewHybridStorageDeclaration(
		object,
		storageDeclarationTestHash("hybrid/state-root"),
		"arweave/storage-object",
		StorageRetrievalHybridEndpoint,
		storageDeclarationTestHash("hybrid/access-receipt"),
		object.Size,
	)
	require.NoError(t, err)
	require.Equal(t, object.ContentHash, declaration.ContentHashOptional)
	require.Equal(t, StorageVerificationHybridCommitment, declaration.VerificationMethod)

	plan, err := BuildStorageConsensusValidationPlan(declaration)
	require.NoError(t, err)
	require.False(t, plan.RequiresOffchainPayload)
	require.Contains(t, plan.RequiredCommitments, object.ContentHash)
	require.Contains(t, plan.RequiredCommitments, declaration.StateRootOptional)
	require.NoError(t, plan.Validate())
}

func TestStorageDeclarationRejectsInvalidModelRules(t *testing.T) {
	_, err := NewStorageDeclaration(StorageDeclaration{
		StorageModel:		StorageModelPersistentOnChain,
		RetrievalMethod:	StorageRetrievalOnChainState,
		VerificationMethod:	StorageVerificationStateRoot,
		RetentionPolicy:	StorageRetentionPermanent,
		AccessPolicy:		AccessPolicyPrivate,
		MaxPayloadBytes:	1,
	})
	require.ErrorContains(t, err, "state root")

	_, err = NewStorageDeclaration(StorageDeclaration{
		StorageModel:		StorageModelDistributedOffChain,
		RetrievalMethod:	StorageRetrievalContentAddressed,
		VerificationMethod:	StorageVerificationContentHash,
		RetentionPolicy:	StorageRetentionExpiry,
		AccessPolicy:		AccessPolicyPublicRead,
		MaxPayloadBytes:	1,
	})
	require.ErrorContains(t, err, "content hash")

	_, err = NewStorageDeclaration(StorageDeclaration{
		StorageModel:		StorageModelHybrid,
		ContentHashOptional:	storageDeclarationTestHash("hybrid/content"),
		StateRootOptional:	storageDeclarationTestHash("hybrid/root"),
		RetrievalMethod:	StorageRetrievalHybridEndpoint,
		VerificationMethod:	StorageVerificationHybridCommitment,
		RetentionPolicy:	StorageRetentionHeight,
		AccessPolicy:		AccessPolicyPermissioned,
		MaxPayloadBytes:	1,
	})
	require.ErrorContains(t, err, "content location")

	_, err = NewStorageDeclaration(StorageDeclaration{
		StorageModel:		StorageModelEphemeral,
		RetrievalMethod:	StorageRetrievalProviderRPC,
		VerificationMethod:	StorageVerificationNone,
		RetentionPolicy:	StorageRetentionNone,
		AccessPolicy:		AccessPolicyPrivate,
		MaxPayloadBytes:	1,
	})
	require.ErrorContains(t, err, "ephemeral")
}

func TestStorageConsensusPlanRejectsOffchainPayloadRequirement(t *testing.T) {
	declaration, err := NewStorageDeclaration(StorageDeclaration{
		StorageModel:			StorageModelHybrid,
		ContentHashOptional:		storageDeclarationTestHash("hybrid/content"),
		StateRootOptional:		storageDeclarationTestHash("hybrid/root"),
		ContentLocationOptional:	"provider/object",
		RetrievalMethod:		StorageRetrievalHybridEndpoint,
		VerificationMethod:		StorageVerificationHybridCommitment,
		RetentionPolicy:		StorageRetentionHeight,
		AccessPolicy:			AccessPolicyPermissioned,
		MaxPayloadBytes:		512,
	})
	require.NoError(t, err)
	plan, err := BuildStorageConsensusValidationPlan(declaration)
	require.NoError(t, err)
	plan.RequiresOffchainPayload = true
	plan.PlanHash = ComputeStorageConsensusValidationPlanHash(plan)
	require.ErrorContains(t, plan.Validate(), "off-chain payload")
}

func storageDeclarationTestObject(t *testing.T) StorageObject {
	t.Helper()
	object, err := NewStorageObject(StorageObject{
		ContentHash:		storageDeclarationTestHash("object/content"),
		ChunkRoots:		[]string{storageDeclarationTestHash("object/chunk-1"), storageDeclarationTestHash("object/chunk-2")},
		Size:			2048,
		ReplicationPolicy:	ReplicationPolicyErasure,
		AccessPolicy:		AccessPolicyPermissioned,
		Owner:			"alice",
		StorageClass:		StorageClassWarm,
		CreatedHeight:		10,
		ExpiresHeightOptional:	100,
		MetadataHashOptional:	storageDeclarationTestHash("object/metadata"),
		AvailabilityCommitment:	storageDeclarationTestHash("object/availability"),
	})
	require.NoError(t, err)
	return object
}

func storageDeclarationTestHash(seed string) string {
	return storageHashParts("storage-declaration-test", seed)
}
