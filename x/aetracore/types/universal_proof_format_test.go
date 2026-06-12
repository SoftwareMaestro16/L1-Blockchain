package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniversalProofTypesCoverSectionSevenThree(t *testing.T) {
	require.Equal(t, []UniversalProofType{
		ProofTypeAccountState,
		ProofTypeBalance,
		ProofTypeZoneRoot,
		ProofTypeShardRoot,
		ProofTypeMessageInclusion,
		ProofTypeMessageReceipt,
		ProofTypeDomainOwnership,
		ProofTypeResolverRecord,
		ProofTypeContractState,
		ProofTypePaymentSettlement,
		ProofTypeNonExistence,
	}, SupportedUniversalProofTypes())

	for _, proofType := range SupportedUniversalProofTypes() {
		_, found := UniversalProofRequirementForType(proofType)
		require.True(t, found, "missing requirement for %s", proofType)
	}
}

func TestVerifyUniversalProofAcceptsAccountStateValue(t *testing.T) {
	proof := testUniversalProofEnvelope(t, ProofTypeAccountState, AccountProofRootType, "", "", testHash("account-root"), []byte("account/alice"), []byte("nonce:7"), nil)
	result := VerifyUniversalProof(proof, testTrustedHeader(proof))

	require.True(t, result.Verified)
	require.False(t, result.VerifiedAbsent)
	require.Equal(t, []byte("nonce:7"), result.Value)
	require.Empty(t, result.FailureCode)
}

func TestVerifyUniversalProofRejectsHeaderChainAndRootMismatch(t *testing.T) {
	proof := testUniversalProofEnvelope(t, ProofTypeAccountState, AccountProofRootType, "", "", testHash("account-root"), []byte("account/alice"), []byte("nonce:7"), nil)

	untrusted := testTrustedHeader(proof)
	untrusted.Trusted = false
	require.Equal(t, ProofFailureUntrustedHeader, VerifyUniversalProof(proof, untrusted).FailureCode)

	wrongChain := testTrustedHeader(proof)
	wrongChain.ChainID = "other-chain"
	require.Equal(t, ProofFailureChainIDMismatch, VerifyUniversalProof(proof, wrongChain).FailureCode)

	rootDrift := proof
	rootDrift.VerificationPath = append([]UniversalRootStep(nil), proof.VerificationPath...)
	rootDrift.VerificationPath[0].ToRoot = testHash("wrong-root")
	rootDrift.VerificationPath[0].StepHash = ComputeUniversalRootStepHash(rootDrift.VerificationPath[0])
	rootDrift.ProofHash = ComputeUniversalProofEnvelopeHash(rootDrift)
	require.Equal(t, ProofFailureRootMismatch, VerifyUniversalProof(rootDrift, testTrustedHeader(rootDrift)).FailureCode)
}

func TestVerifyUniversalProofRequiresZoneAndShardCommitments(t *testing.T) {
	height := uint64(22)
	shardRoot := testHash("financial-shard-0-root")
	shardAggregate := testHash("financial-shards-root")
	zoneCommitment := testCommitment(t, height, ZoneIDFinancial)
	zoneCommitment.ShardRootsRoot = shardAggregate
	zoneCommitment.CommitmentHash = ComputeZoneCommitmentHash(zoneCommitment)
	shardCommitment, err := NewUniversalShardCommitment(UniversalShardCommitment{
		Height:		height,
		ZoneID:		ZoneIDFinancial,
		ShardID:	"0",
		ShardRoot:	shardRoot,
		ShardRootsRoot:	shardAggregate,
	})
	require.NoError(t, err)

	proof := testUniversalProofEnvelope(t, ProofTypeShardRoot, ShardStateProofRootType, ZoneIDFinancial, "0", shardRoot, []byte("financial/balances/alice/naet"), []byte("100"), nil)
	proof.Height = height
	proof.StoreProof, err = NewUniversalStoreProof(UniversalStoreProof{ProofVersion: UniversalProofVersionV1, Key: proof.Key, Value: proof.Value, StoreRoot: shardRoot, ProofOps: []string{testHash("op")}})
	require.NoError(t, err)
	proof.VerificationPath = testUniversalProofPath(t, proof.AppHash, ShardStateProofRootType, shardRoot)
	proof.ZoneCommitment = zoneCommitment
	proof.HasZoneCommit = true
	proof.ShardCommitment = shardCommitment
	proof.HasShardCommit = true
	proof.ProofHash = ComputeUniversalProofEnvelopeHash(proof)
	proof, err = NewUniversalProofEnvelope(proof)
	require.NoError(t, err)

	result := VerifyUniversalProof(proof, testTrustedHeader(proof))
	require.True(t, result.Verified)

	noShard := proof
	noShard.HasShardCommit = false
	noShard.ProofHash = ComputeUniversalProofEnvelopeHash(noShard)
	require.Equal(t, ProofFailureRootMismatch, VerifyUniversalProof(noShard, testTrustedHeader(noShard)).FailureCode)

	wrongShard := proof
	wrongShard.ShardCommitment.ShardID = "1"
	wrongShard.ShardCommitment.CommitmentHash = ComputeUniversalShardCommitmentHash(wrongShard.ShardCommitment)
	wrongShard.ProofHash = ComputeUniversalProofEnvelopeHash(wrongShard)
	require.Equal(t, ProofFailureShardNotFound, VerifyUniversalProof(wrongShard, testTrustedHeader(wrongShard)).FailureCode)
}

func TestVerifyUniversalProofMessageInclusionAndReceipt(t *testing.T) {
	height := uint64(28)
	messageRoot := testHash("global-message-root")
	receiptRoot := testHash("receipt-root")
	messageCommit, err := NewUniversalMessageCommitment(UniversalMessageCommitment{
		Height:			height,
		MessageID:		testHash("msg-id"),
		MessageRoot:		messageRoot,
		SourceOutboxRoot:	testHash("outbox"),
		DestinationInboxRoot:	testHash("inbox"),
		ReceiptRoot:		receiptRoot,
		ReceiptHash:		testHash("receipt"),
	})
	require.NoError(t, err)

	inclusion := testUniversalProofEnvelope(t, ProofTypeMessageInclusion, MessageProofRootType, "", "", messageRoot, []byte("message/msg-id"), []byte("message-bytes"), nil)
	inclusion.Height = height
	inclusion.MessageCommit = messageCommit
	inclusion.HasMessageCommit = true
	inclusion.StoreProof, err = NewUniversalStoreProof(UniversalStoreProof{ProofVersion: UniversalProofVersionV1, Key: inclusion.Key, Value: inclusion.Value, StoreRoot: messageRoot, ProofOps: []string{testHash("message-op")}})
	require.NoError(t, err)
	inclusion.VerificationPath = testUniversalProofPath(t, inclusion.AppHash, MessageProofRootType, messageRoot)
	inclusion.ProofHash = ComputeUniversalProofEnvelopeHash(inclusion)
	inclusion, err = NewUniversalProofEnvelope(inclusion)
	require.NoError(t, err)
	require.True(t, VerifyUniversalProof(inclusion, testTrustedHeader(inclusion)).Verified)

	receipt := testUniversalProofEnvelope(t, ProofTypeMessageReceipt, ReceiptProofRootType, "", "", receiptRoot, []byte("receipt/msg-id"), []byte("executed"), nil)
	receipt.Height = height
	receipt.MessageCommit = messageCommit
	receipt.HasMessageCommit = true
	receipt.StoreProof, err = NewUniversalStoreProof(UniversalStoreProof{ProofVersion: UniversalProofVersionV1, Key: receipt.Key, Value: receipt.Value, StoreRoot: receiptRoot, ProofOps: []string{testHash("receipt-op")}})
	require.NoError(t, err)
	receipt.VerificationPath = testUniversalProofPath(t, receipt.AppHash, ReceiptProofRootType, receiptRoot)
	receipt.ProofHash = ComputeUniversalProofEnvelopeHash(receipt)
	receipt, err = NewUniversalProofEnvelope(receipt)
	require.NoError(t, err)
	require.True(t, VerifyUniversalProof(receipt, testTrustedHeader(receipt)).Verified)

	missingReceipt := receipt
	missingReceipt.MessageCommit.ReceiptHash = EmptyRootHash
	missingReceipt.MessageCommit, err = NewUniversalMessageCommitment(missingReceipt.MessageCommit)
	require.NoError(t, err)
	missingReceipt.ProofHash = ComputeUniversalProofEnvelopeHash(missingReceipt)
	require.Equal(t, ProofFailureReceiptNotFound, VerifyUniversalProof(missingReceipt, testTrustedHeader(missingReceipt)).FailureCode)
}

func TestVerifyUniversalProofNonExistence(t *testing.T) {
	proof := testUniversalProofEnvelope(t, ProofTypeNonExistence, ResolverProofRootType, "", "", testHash("resolver-root"), []byte("identity/resolvers/missing.aet"), nil, []byte("between:alice.aet:bob.aet"))
	result := VerifyUniversalProof(proof, testTrustedHeader(proof))

	require.True(t, result.Verified)
	require.True(t, result.VerifiedAbsent)

	badMarker := proof
	badMarker.StoreProof.NonExistenceMarker = []byte("different-boundary")
	badMarker.StoreProof.ProofHash = ComputeUniversalStoreProofHash(badMarker.StoreProof)
	badMarker.ProofHash = ComputeUniversalProofEnvelopeHash(badMarker)
	require.Equal(t, ProofFailureNonExistenceProofInvalid, VerifyUniversalProof(badMarker, testTrustedHeader(badMarker)).FailureCode)
}

func testUniversalProofEnvelope(
	t *testing.T,
	proofType UniversalProofType,
	rootType RootType,
	zoneID ZoneID,
	shardID ShardID,
	storeRoot string,
	key []byte,
	value []byte,
	absence []byte,
) UniversalProofEnvelope {
	t.Helper()
	height := uint64(9)
	appHash := testHash("app-hash-" + string(proofType))
	storeProof, err := NewUniversalStoreProof(UniversalStoreProof{
		ProofVersion:		UniversalProofVersionV1,
		Key:			key,
		Value:			value,
		NonExistenceMarker:	absence,
		StoreRoot:		storeRoot,
		ProofOps:		[]string{testHash("store-proof-op-" + string(proofType))},
	})
	require.NoError(t, err)
	proof, err := NewUniversalProofEnvelope(UniversalProofEnvelope{
		ProofType:		proofType,
		ProofVersion:		UniversalProofVersionV1,
		ChainID:		"aetra-test",
		Height:			height,
		AppHash:		appHash,
		RootType:		rootType,
		ZoneID:			zoneID,
		ShardID:		shardID,
		Key:			key,
		Value:			value,
		AbsenceMarker:		absence,
		StoreProof:		storeProof,
		VerificationPath:	testUniversalProofPath(t, appHash, rootType, storeRoot),
	})
	require.NoError(t, err)
	return proof
}

func testUniversalProofPath(t *testing.T, appHash string, rootType RootType, storeRoot string) []UniversalRootStep {
	t.Helper()
	step, err := NewUniversalRootStep(UniversalRootStep{
		Index:		0,
		FromRootType:	DefaultProofRootType,
		FromRoot:	appHash,
		ToRootType:	rootType,
		ToRoot:		storeRoot,
		Scope:		"store",
	})
	require.NoError(t, err)
	return []UniversalRootStep{step}
}

func testTrustedHeader(proof UniversalProofEnvelope) UniversalTrustedHeader {
	return UniversalTrustedHeader{
		ChainID:	proof.ChainID,
		Height:		proof.Height,
		AppHash:	proof.AppHash,
		HeaderHash:	testHash("header-" + proof.AppHash),
		Trusted:	true,
	}
}
