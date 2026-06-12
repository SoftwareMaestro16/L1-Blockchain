package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServiceStorageDescriptorProjectsDeclarationFields(t *testing.T) {
	descriptor := coretypes.ServiceStorageDescriptor{
		Model:			coretypes.ServiceStorageHybridCommitment,
		CommitmentHash:		storageTestHash("descriptor/commitment"),
		ContentHash:		storageTestHash("descriptor/content"),
		StateRoot:		storageTestHash("descriptor/state-root"),
		RetrievalMethod:	StorageRetrievalHybridEndpoint,
		VerificationMethod:	StorageVerificationHybridCommitment,
		RetentionPolicy:	StorageRetentionHeight,
		AccessPolicy:		AccessPolicyPermissioned,
		MaxPayloadBytes:	4096,
		ProofRequired:		true,
	}
	require.NoError(t, descriptor.Validate())

	declaration, err := NewStorageDeclarationFromServiceDescriptor(descriptor)
	require.NoError(t, err)
	require.Equal(t, StorageModelHybrid, declaration.StorageModel)
	require.Equal(t, descriptor.ContentHash, declaration.ContentHashOptional)
	require.Equal(t, descriptor.StateRoot, declaration.StateRootOptional)
	require.Equal(t, descriptor.RetrievalMethod, declaration.RetrievalMethod)
	require.Equal(t, descriptor.VerificationMethod, declaration.VerificationMethod)
	require.Equal(t, descriptor.MaxPayloadBytes, declaration.MaxPayloadBytes)
	require.NoError(t, declaration.Validate())
}

func TestServiceStorageReceiptAnchoringAndProofHook(t *testing.T) {
	object := testStorageObject(t, "alice", []string{storageTestHash("chunk-a"), storageTestHash("chunk-b")})
	chunkSet, err := NewStorageChunkSet(object.ObjectID, object.ChunkRoots)
	require.NoError(t, err)
	proof, err := NewStorageRetrievalProof(StorageRetrievalProof{
		ObjectID:	object.ObjectID,
		ContentHash:	object.ContentHash,
		ChunkRoot:	chunkSet.ChunkRoot,
		ChunkIndex:	0,
		ChunkHash:	object.ChunkRoots[0],
		ProofPath:	[]string{object.ChunkRoots[1]},
	})
	require.NoError(t, err)

	receipt, err := NewServiceStorageReceipt(ServiceStorageReceipt{
		ServiceID:	"service/storage",
		ObjectID:	object.ObjectID,
		RequestHash:	storageTestHash("request/read"),
		ContentHash:	object.ContentHash,
		ProviderID:	"provider-1",
		AccessHeight:	42,
		Signature:	storageTestHash("provider/signature"),
		ProofOptional:	proof.ProofHash,
	})
	require.NoError(t, err)
	require.NotEmpty(t, receipt.ReceiptID)
	require.NoError(t, VerifyServiceStorageReceiptProof(receipt, object, proof))

	anchor, err := AnchorServiceStorageReceipts("service/storage", []ServiceStorageReceipt{receipt}, 45)
	require.NoError(t, err)
	require.Equal(t, ComputeServiceStorageReceiptRoot(anchor.Receipts), anchor.RootHash)
	require.Equal(t, ComputeServiceStorageReceiptAnchorHash(anchor), anchor.AnchorHash)

	broken := receipt
	broken.ContentHash = storageTestHash("wrong-content")
	broken.ReceiptHash = ComputeServiceStorageReceiptHash(broken)
	require.ErrorContains(t, VerifyServiceStorageReceiptProof(broken, object, proof), "content hash")
}

func TestServiceStorageReceiptRequiresSignature(t *testing.T) {
	object := testStorageObject(t, "alice", []string{storageTestHash("chunk-a")})
	_, err := NewServiceStorageReceipt(ServiceStorageReceipt{
		ServiceID:	"service/storage",
		ObjectID:	object.ObjectID,
		RequestHash:	storageTestHash("request/read"),
		ContentHash:	object.ContentHash,
		ProviderID:	"provider-1",
		AccessHeight:	42,
	})
	require.ErrorContains(t, err, "signature")
}

func TestOnChainStorageFeeModelQuotesPayloads(t *testing.T) {
	declaration, err := NewStorageDeclaration(StorageDeclaration{
		StorageModel:		StorageModelPersistentOnChain,
		StateRootOptional:	storageTestHash("state/root"),
		RetrievalMethod:	StorageRetrievalOnChainState,
		VerificationMethod:	StorageVerificationStateRoot,
		RetentionPolicy:	StorageRetentionPermanent,
		AccessPolicy:		AccessPolicyPermissioned,
		MaxPayloadBytes:	1024,
	})
	require.NoError(t, err)
	model, err := NewOnChainStorageFeeModel(OnChainStorageFeeModel{
		Denom:			"naet",
		PricePerByte:		2,
		MinimumFee:		500,
		MaxPayloadBytes:	2048,
	})
	require.NoError(t, err)

	minQuote, err := QuoteOnChainStorageFee(declaration, model, 100)
	require.NoError(t, err)
	require.Equal(t, uint64(500), minQuote.FeeAmount)

	sizeQuote, err := QuoteOnChainStorageFee(declaration, model, 300)
	require.NoError(t, err)
	require.Equal(t, uint64(600), sizeQuote.FeeAmount)

	hybrid := declaration
	hybrid.StorageModel = StorageModelHybrid
	hybrid.ContentHashOptional = storageTestHash("hybrid/content")
	hybrid.ContentLocationOptional = "provider/object"
	hybrid.RetrievalMethod = StorageRetrievalHybridEndpoint
	hybrid.VerificationMethod = StorageVerificationHybridCommitment
	hybrid.DeclarationHash = ComputeStorageDeclarationHash(hybrid)
	_, err = QuoteOnChainStorageFee(hybrid, model, 10)
	require.ErrorContains(t, err, "persistent on-chain")
}
