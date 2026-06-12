package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestProofFailureCodeCatalogCoversSectionSevenSix(t *testing.T) {
	codes := SupportedProofFailureCodes()
	require.Len(t, codes, 11)

	seen := make(map[coretypes.UniversalProofFailureCode]ProofFailureCategory, len(codes))
	for _, descriptor := range codes {
		require.NotEmpty(t, descriptor.Meaning)
		require.NotEmpty(t, descriptor.Category)
		seen[descriptor.Code] = descriptor.Category
	}

	for _, code := range []coretypes.UniversalProofFailureCode{
		coretypes.ProofFailureUntrustedHeader,
		coretypes.ProofFailureChainIDMismatch,
		coretypes.ProofFailureHeightUnavailable,
		coretypes.ProofFailureRootMismatch,
		coretypes.ProofFailureZoneNotFound,
		coretypes.ProofFailureShardNotFound,
		coretypes.ProofFailureStoreProofInvalid,
		coretypes.ProofFailureMessageNotIncluded,
		coretypes.ProofFailureReceiptNotFound,
		coretypes.ProofFailureObjectExpired,
		coretypes.ProofFailureNonExistenceProofInvalid,
	} {
		_, found := seen[code]
		require.True(t, found, "missing failure code %s", code)
	}
}

func TestProofRegistryStoresSnapshotsAndPrunesHistory(t *testing.T) {
	state, err := NewProofRegistryState(2)
	require.NoError(t, err)

	for height := uint64(1); height <= 3; height++ {
		snapshot := testProofRegistrySnapshot(t, height, proofRegistryTestHash("app", fmt.Sprint(height)))
		state, err = CommitProofRegistrySnapshot(state, snapshot)
		require.NoError(t, err)
	}

	require.Len(t, state.Snapshots, 2)
	require.Equal(t, uint64(2), state.Snapshots[0].Height)
	require.Equal(t, uint64(3), state.Snapshots[1].Height)
	require.NoError(t, state.Validate())
}

func TestProofRegistryQueriesStoreZoneShardMessageAndReceiptProofs(t *testing.T) {
	height := uint64(11)
	appHash := proofRegistryTestHash("app")
	accountRoot := proofRegistryTestHash("account-root")
	messageRoot := proofRegistryTestHash("message-root")
	receiptRoot := proofRegistryTestHash("receipt-root")

	state, err := NewProofRegistryState(8)
	require.NoError(t, err)
	state, err = CommitProofRegistrySnapshot(state, testProofRegistrySnapshotWithRoots(t, height, appHash, []coretypes.ProofRoot{
		testProofRoot(height, coretypes.AccountProofRootType, "", accountRoot),
		testProofRoot(height, coretypes.MessageProofRootType, "", messageRoot),
		testProofRoot(height, coretypes.ReceiptProofRootType, "", receiptRoot),
		testProofRoot(height, coretypes.ZoneStateProofRootType, coretypes.ZoneIDFinancial, proofRegistryTestHash("financial-zone-root")),
		testProofRoot(height, coretypes.ShardStateProofRootType, coretypes.ZoneIDFinancial, proofRegistryTestHash("financial-shard-root")),
	}))
	require.NoError(t, err)

	account := testProofRegistryEnvelope(t, coretypes.ProofTypeAccountState, coretypes.AccountProofRootType, "", "", height, appHash, accountRoot, []byte("account/alice"), []byte("nonce:9"), nil)
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindStore, "", account))
	require.NoError(t, err)

	storeResp, err := QueryProof(state, ProofRegistryQuery{Kind: QueryKindStore, Height: height, ProofType: coretypes.ProofTypeAccountState, RootType: coretypes.AccountProofRootType, Key: []byte("account/alice")})
	require.NoError(t, err)
	require.True(t, storeResp.Found)
	require.True(t, coretypes.VerifyUniversalProof(storeResp.Envelope, testProofRegistryTrustedHeader(storeResp.Envelope)).Verified)

	zoneProof := testProofRegistryZoneProof(t, height, appHash, coretypes.ZoneIDFinancial)
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindZone, "", zoneProof))
	require.NoError(t, err)
	zoneResp, err := QueryZoneProof(state, height, coretypes.ZoneIDFinancial)
	require.NoError(t, err)
	require.True(t, zoneResp.Found)

	shardProof := testProofRegistryShardProof(t, height, appHash, coretypes.ZoneIDFinancial, "0")
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindShard, "", shardProof))
	require.NoError(t, err)
	shardResp, err := QueryShardProof(state, height, coretypes.ZoneIDFinancial, "0")
	require.NoError(t, err)
	require.True(t, shardResp.Found)

	messageCommit := testProofRegistryMessageCommitment(t, height, messageRoot, receiptRoot, false)
	message := testProofRegistryEnvelope(t, coretypes.ProofTypeMessageInclusion, coretypes.MessageProofRootType, "", "", height, appHash, messageRoot, []byte("messages/msg"), []byte("encoded-message"), nil)
	message.MessageCommit = messageCommit
	message.HasMessageCommit = true
	message, err = coretypes.NewUniversalProofEnvelope(message)
	require.NoError(t, err)
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindMessageInclusion, messageCommit.MessageID, message))
	require.NoError(t, err)
	msgResp, err := QueryMessageInclusionProof(state, height, messageCommit.MessageID)
	require.NoError(t, err)
	require.True(t, msgResp.Found)

	receipt := testProofRegistryEnvelope(t, coretypes.ProofTypeMessageReceipt, coretypes.ReceiptProofRootType, "", "", height, appHash, receiptRoot, []byte("receipts/msg"), []byte("executed"), nil)
	receipt.MessageCommit = messageCommit
	receipt.HasMessageCommit = true
	receipt, err = coretypes.NewUniversalProofEnvelope(receipt)
	require.NoError(t, err)
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindReceipt, messageCommit.MessageID, receipt))
	require.NoError(t, err)
	receiptResp, err := QueryReceiptProof(state, height, messageCommit.MessageID)
	require.NoError(t, err)
	require.True(t, receiptResp.Found)
}

func TestProofRegistryQueriesIdentityContractAndPaymentProofs(t *testing.T) {
	height := uint64(17)
	appHash := proofRegistryTestHash("app-scoped")
	state, err := NewProofRegistryState(8)
	require.NoError(t, err)
	state, err = CommitProofRegistrySnapshot(state, testProofRegistrySnapshotWithRoots(t, height, appHash, []coretypes.ProofRoot{
		testProofRoot(height, coretypes.DomainOwnershipProofRootType, coretypes.ZoneIDIdentity, proofRegistryTestHash("identity-root")),
		testProofRoot(height, coretypes.ContractStateProofRootType, coretypes.ZoneIDContract, proofRegistryTestHash("contract-root")),
		testProofRoot(height, coretypes.PaymentSettlementProofRootType, coretypes.ZoneIDFinancial, proofRegistryTestHash("payment-root")),
	}))
	require.NoError(t, err)

	identity := testProofRegistryScopedProof(t, height, appHash, coretypes.ProofTypeDomainOwnership, coretypes.DomainOwnershipProofRootType, coretypes.ZoneIDIdentity, "0", []byte("identity/domains/alice.aet"), []byte("owner:alice"))
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindIdentity, "", identity))
	require.NoError(t, err)
	identityResp, err := QueryIdentityProof(state, height, []byte("identity/domains/alice.aet"))
	require.NoError(t, err)
	require.True(t, identityResp.Found)

	contract := testProofRegistryScopedProof(t, height, appHash, coretypes.ProofTypeContractState, coretypes.ContractStateProofRootType, coretypes.ZoneIDContract, "1", []byte("contract/storage/c1/k"), []byte("v"))
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindContractState, "", contract))
	require.NoError(t, err)
	contractResp, err := QueryContractStateProof(state, height, []byte("contract/storage/c1/k"))
	require.NoError(t, err)
	require.True(t, contractResp.Found)

	payment := testProofRegistryScopedProof(t, height, appHash, coretypes.ProofTypePaymentSettlement, coretypes.PaymentSettlementProofRootType, coretypes.ZoneIDFinancial, "2", []byte("financial/payments/settlements/p1"), []byte("settled"))
	state, err = AddProofRegistryEntry(state, testProofRegistryEntry(t, QueryKindPaymentSettlement, "", payment))
	require.NoError(t, err)
	paymentResp, err := QueryPaymentSettlementProof(state, height, []byte("financial/payments/settlements/p1"))
	require.NoError(t, err)
	require.True(t, paymentResp.Found)
}

func TestProofRegistryTestVectorsCoverEveryFailureCode(t *testing.T) {
	base := testProofRegistryEnvelope(t, coretypes.ProofTypeAccountState, coretypes.AccountProofRootType, "", "", 31, proofRegistryTestHash("vector-app"), proofRegistryTestHash("account-root"), []byte("account/alice"), []byte("nonce:1"), nil)
	trusted := testProofRegistryTrustedHeader(base)
	vectors := testProofFailureVectors(t, base, trusted)
	require.Len(t, vectors, len(SupportedProofFailureCodes()))

	seen := make(map[coretypes.UniversalProofFailureCode]struct{}, len(vectors))
	state, err := NewProofRegistryState(8)
	require.NoError(t, err)
	for _, vector := range vectors {
		seen[vector.ExpectedFailureCode] = struct{}{}
		state, err = AddProofTestVector(state, vector)
		require.NoError(t, err)
	}
	for _, descriptor := range SupportedProofFailureCodes() {
		_, found := seen[descriptor.Code]
		require.True(t, found, "missing vector for %s", descriptor.Code)
	}
	require.Len(t, state.TestVectors, len(SupportedProofFailureCodes()))
}

func testProofRegistrySnapshot(t *testing.T, height uint64, appHash string) ProofRegistrySnapshot {
	t.Helper()
	return testProofRegistrySnapshotWithRoots(t, height, appHash, []coretypes.ProofRoot{
		testProofRoot(height, coretypes.AccountProofRootType, "", proofRegistryTestHash("account", fmt.Sprint(height))),
	})
}

func testProofRegistrySnapshotWithRoots(t *testing.T, height uint64, appHash string, roots []coretypes.ProofRoot) ProofRegistrySnapshot {
	t.Helper()
	metadata := make([]ProofRootMetadata, 0, len(roots))
	for _, root := range roots {
		item, err := NewProofRootMetadata(ProofRootMetadata{RootType: root.RootType, Source: root.Source, Description: "test root"})
		require.NoError(t, err)
		metadata = append(metadata, item)
	}
	snapshot, err := NewProofRegistrySnapshot(ProofRegistrySnapshot{
		Height:			height,
		AppHash:		appHash,
		AetraCoreRoot:		proofRegistryTestHash("core", fmt.Sprint(height)),
		GlobalZoneRoot:		proofRegistryTestHash("zones", fmt.Sprint(height)),
		GlobalMessageRoot:	proofRegistryTestHash("messages", fmt.Sprint(height)),
		ReceiptRoot:		proofRegistryTestHash("receipts", fmt.Sprint(height)),
		Roots:			roots,
		Metadata:		metadata,
	})
	require.NoError(t, err)
	return snapshot
}

func testProofRoot(height uint64, rootType coretypes.RootType, zoneID coretypes.ZoneID, rootHash string) coretypes.ProofRoot {
	return coretypes.ProofRoot{Height: height, RootType: rootType, ZoneID: zoneID, RootHash: rootHash, Source: "proofregistry.test"}
}

func testProofRegistryEntry(t *testing.T, kind ProofRegistryQueryKind, messageID string, envelope coretypes.UniversalProofEnvelope) ProofRegistryEntry {
	t.Helper()
	entry, err := NewProofRegistryEntry(ProofRegistryEntry{Kind: kind, MessageID: messageID, Envelope: envelope})
	require.NoError(t, err)
	return entry
}

func testProofRegistryEnvelope(t *testing.T, proofType coretypes.UniversalProofType, rootType coretypes.RootType, zoneID coretypes.ZoneID, shardID coretypes.ShardID, height uint64, appHash string, storeRoot string, key []byte, value []byte, absence []byte) coretypes.UniversalProofEnvelope {
	t.Helper()
	path := testProofRegistryPath(t, appHash, rootType, storeRoot)
	proof, err := BuildStoreV2ProofEnvelope(StoreV2ProofAdapterRequest{
		ProofType:		proofType,
		ChainID:		"aetra-test",
		Height:			height,
		AppHash:		appHash,
		RootType:		rootType,
		ZoneID:			zoneID,
		ShardID:		shardID,
		Key:			key,
		Value:			value,
		NonExistenceMarker:	absence,
		StoreRoot:		storeRoot,
		ProofOps:		[]string{proofRegistryTestHash("op", string(key))},
		VerificationPath:	path,
	})
	require.NoError(t, err)
	return proof
}

func testProofRegistryZoneProof(t *testing.T, height uint64, appHash string, zoneID coretypes.ZoneID) coretypes.UniversalProofEnvelope {
	t.Helper()
	stateRoot := proofRegistryTestHash("zone-state", string(zoneID))
	zone, err := coretypes.NewZoneCommitment(height, zoneID, stateRoot, proofRegistryTestHash("inbox"), proofRegistryTestHash("outbox"), proofRegistryTestHash("receipts"), proofRegistryTestHash("events"), proofRegistryTestHash("shards"), proofRegistryTestHash("params"), proofRegistryTestHash("summary"))
	require.NoError(t, err)
	proof, err := BuildStoreV2ProofEnvelope(StoreV2ProofAdapterRequest{
		ProofType:		coretypes.ProofTypeZoneRoot,
		ChainID:		"aetra-test",
		Height:			height,
		AppHash:		appHash,
		RootType:		coretypes.ZoneStateProofRootType,
		ZoneID:			zoneID,
		Key:			[]byte("zone/" + string(zoneID)),
		Value:			[]byte("zone-root"),
		StoreRoot:		stateRoot,
		ProofOps:		[]string{proofRegistryTestHash("zone-op")},
		VerificationPath:	testProofRegistryPath(t, appHash, coretypes.ZoneStateProofRootType, stateRoot),
		ZoneCommitment:		zone,
	})
	require.NoError(t, err)
	return proof
}

func testProofRegistryShardProof(t *testing.T, height uint64, appHash string, zoneID coretypes.ZoneID, shardID coretypes.ShardID) coretypes.UniversalProofEnvelope {
	t.Helper()
	shardRoot := proofRegistryTestHash("shard-root", string(shardID))
	shardAggregate := proofRegistryTestHash("shard-aggregate")
	zone, err := coretypes.NewZoneCommitment(height, zoneID, proofRegistryTestHash("zone-state"), proofRegistryTestHash("inbox"), proofRegistryTestHash("outbox"), proofRegistryTestHash("receipts"), proofRegistryTestHash("events"), shardAggregate, proofRegistryTestHash("params"), proofRegistryTestHash("summary"))
	require.NoError(t, err)
	shard, err := coretypes.NewUniversalShardCommitment(coretypes.UniversalShardCommitment{Height: height, ZoneID: zoneID, ShardID: shardID, ShardRoot: shardRoot, ShardRootsRoot: shardAggregate})
	require.NoError(t, err)
	proof, err := BuildStoreV2ProofEnvelope(StoreV2ProofAdapterRequest{
		ProofType:		coretypes.ProofTypeShardRoot,
		ChainID:		"aetra-test",
		Height:			height,
		AppHash:		appHash,
		RootType:		coretypes.ShardStateProofRootType,
		ZoneID:			zoneID,
		ShardID:		shardID,
		Key:			[]byte("shard/" + string(shardID)),
		Value:			[]byte("shard-root"),
		StoreRoot:		shardRoot,
		ProofOps:		[]string{proofRegistryTestHash("shard-op")},
		VerificationPath:	testProofRegistryPath(t, appHash, coretypes.ShardStateProofRootType, shardRoot),
		ZoneCommitment:		zone,
		ShardCommitment:	shard,
	})
	require.NoError(t, err)
	return proof
}

func testProofRegistryScopedProof(t *testing.T, height uint64, appHash string, proofType coretypes.UniversalProofType, rootType coretypes.RootType, zoneID coretypes.ZoneID, shardID coretypes.ShardID, key []byte, value []byte) coretypes.UniversalProofEnvelope {
	t.Helper()
	storeRoot := proofRegistryTestHash("scoped-root", string(proofType), string(shardID))
	shardAggregate := proofRegistryTestHash("scoped-shards", string(proofType))
	zone, err := coretypes.NewZoneCommitment(height, zoneID, proofRegistryTestHash("zone-state", string(zoneID)), proofRegistryTestHash("inbox"), proofRegistryTestHash("outbox"), proofRegistryTestHash("receipts"), proofRegistryTestHash("events"), shardAggregate, proofRegistryTestHash("params"), proofRegistryTestHash("summary"))
	require.NoError(t, err)
	shard, err := coretypes.NewUniversalShardCommitment(coretypes.UniversalShardCommitment{Height: height, ZoneID: zoneID, ShardID: shardID, ShardRoot: storeRoot, ShardRootsRoot: shardAggregate})
	require.NoError(t, err)
	proof, err := BuildStoreV2ProofEnvelope(StoreV2ProofAdapterRequest{
		ProofType:		proofType,
		ChainID:		"aetra-test",
		Height:			height,
		AppHash:		appHash,
		RootType:		rootType,
		ZoneID:			zoneID,
		ShardID:		shardID,
		Key:			key,
		Value:			value,
		StoreRoot:		storeRoot,
		ProofOps:		[]string{proofRegistryTestHash("scoped-op", string(key))},
		VerificationPath:	testProofRegistryPath(t, appHash, rootType, storeRoot),
		ZoneCommitment:		zone,
		ShardCommitment:	shard,
	})
	require.NoError(t, err)
	return proof
}

func testProofRegistryMessageCommitment(t *testing.T, height uint64, messageRoot string, receiptRoot string, emptyReceipt bool) coretypes.UniversalMessageCommitment {
	t.Helper()
	receiptHash := proofRegistryTestHash("receipt-hash")
	if emptyReceipt {
		receiptHash = coretypes.EmptyRootHash
	}
	commitment, err := coretypes.NewUniversalMessageCommitment(coretypes.UniversalMessageCommitment{
		Height:			height,
		MessageID:		proofRegistryTestHash("message-id"),
		MessageRoot:		messageRoot,
		SourceOutboxRoot:	proofRegistryTestHash("outbox"),
		DestinationInboxRoot:	proofRegistryTestHash("inbox"),
		ReceiptRoot:		receiptRoot,
		ReceiptHash:		receiptHash,
	})
	require.NoError(t, err)
	return commitment
}

func testProofRegistryPath(t *testing.T, appHash string, rootType coretypes.RootType, storeRoot string) []coretypes.UniversalRootStep {
	t.Helper()
	step, err := coretypes.NewUniversalRootStep(coretypes.UniversalRootStep{
		Index:		0,
		FromRootType:	coretypes.DefaultProofRootType,
		FromRoot:	appHash,
		ToRootType:	rootType,
		ToRoot:		storeRoot,
		Scope:		"proofregistry",
	})
	require.NoError(t, err)
	return []coretypes.UniversalRootStep{step}
}

func testProofRegistryTrustedHeader(proof coretypes.UniversalProofEnvelope) coretypes.UniversalTrustedHeader {
	return coretypes.UniversalTrustedHeader{ChainID: proof.ChainID, Height: proof.Height, AppHash: proof.AppHash, HeaderHash: proofRegistryTestHash("header", proof.AppHash), Trusted: true}
}

func testProofFailureVectors(t *testing.T, base coretypes.UniversalProofEnvelope, trusted coretypes.UniversalTrustedHeader) []ProofTestVector {
	t.Helper()
	vector := func(name string, proof coretypes.UniversalProofEnvelope, header coretypes.UniversalTrustedHeader, code coretypes.UniversalProofFailureCode) ProofTestVector {
		out, err := NewProofTestVector(ProofTestVector{Name: name, Positive: false, Proof: proof, TrustedHeader: header, ExpectedFailureCode: code})
		require.NoError(t, err)
		return out
	}
	rehash := func(proof coretypes.UniversalProofEnvelope) coretypes.UniversalProofEnvelope {
		proof.ProofHash = coretypes.ComputeUniversalProofEnvelopeHash(proof)
		return proof
	}

	untrusted := trusted
	untrusted.Trusted = false
	wrongChain := trusted
	wrongChain.ChainID = "other-chain"
	wrongHeight := trusted
	wrongHeight.Height++
	rootDrift := base
	rootDrift.VerificationPath = append([]coretypes.UniversalRootStep(nil), rootDrift.VerificationPath...)
	rootDrift.VerificationPath[0].ToRoot = proofRegistryTestHash("wrong-root")
	rootDrift.VerificationPath[0].StepHash = coretypes.ComputeUniversalRootStepHash(rootDrift.VerificationPath[0])
	rootDrift = rehash(rootDrift)
	zoneMissing := base
	zoneMissing.ZoneID = coretypes.ZoneIDFinancial
	zoneMissing = rehash(zoneMissing)
	shardMissing := base
	shardMissing.ShardID = "0"
	shardMissing = rehash(shardMissing)
	storeInvalid := base
	storeInvalid.Value = []byte("different")
	storeInvalid = rehash(storeInvalid)
	msgCommit := testProofRegistryMessageCommitment(t, base.Height, proofRegistryTestHash("different-message-root"), proofRegistryTestHash("receipt-root"), false)
	messageMissing := testProofRegistryEnvelope(t, coretypes.ProofTypeMessageInclusion, coretypes.MessageProofRootType, "", "", base.Height, base.AppHash, base.StoreProof.StoreRoot, []byte("messages/missing"), []byte("message"), nil)
	messageMissing.MessageCommit = msgCommit
	messageMissing.HasMessageCommit = true
	messageMissing = rehash(messageMissing)
	receiptCommit := testProofRegistryMessageCommitment(t, base.Height, proofRegistryTestHash("message-root"), base.StoreProof.StoreRoot, true)
	receiptMissing := testProofRegistryEnvelope(t, coretypes.ProofTypeMessageReceipt, coretypes.ReceiptProofRootType, "", "", base.Height, base.AppHash, base.StoreProof.StoreRoot, []byte("receipts/missing"), []byte("receipt"), nil)
	receiptMissing.MessageCommit = receiptCommit
	receiptMissing.HasMessageCommit = true
	receiptMissing = rehash(receiptMissing)
	expired := base
	expired.ObjectExpiryHeight = base.Height - 1
	expired = rehash(expired)
	nonExistence := testProofRegistryEnvelope(t, coretypes.ProofTypeNonExistence, coretypes.ResolverProofRootType, "", "", base.Height, base.AppHash, base.StoreProof.StoreRoot, []byte("missing"), nil, []byte("left:right"))
	nonExistence.StoreProof.NonExistenceMarker = []byte("wrong")
	nonExistence.StoreProof.ProofHash = coretypes.ComputeUniversalStoreProofHash(nonExistence.StoreProof)
	nonExistence = rehash(nonExistence)

	return []ProofTestVector{
		vector("untrusted-header", base, untrusted, coretypes.ProofFailureUntrustedHeader),
		vector("chain-id-mismatch", base, wrongChain, coretypes.ProofFailureChainIDMismatch),
		vector("height-unavailable", base, wrongHeight, coretypes.ProofFailureHeightUnavailable),
		vector("root-mismatch", rootDrift, trusted, coretypes.ProofFailureRootMismatch),
		vector("zone-not-found", zoneMissing, trusted, coretypes.ProofFailureZoneNotFound),
		vector("shard-not-found", shardMissing, trusted, coretypes.ProofFailureShardNotFound),
		vector("store-proof-invalid", storeInvalid, trusted, coretypes.ProofFailureStoreProofInvalid),
		vector("message-not-included", messageMissing, trusted, coretypes.ProofFailureMessageNotIncluded),
		vector("receipt-not-found", receiptMissing, trusted, coretypes.ProofFailureReceiptNotFound),
		vector("object-expired", expired, trusted, coretypes.ProofFailureObjectExpired),
		vector("non-existence-invalid", nonExistence, trusted, coretypes.ProofFailureNonExistenceProofInvalid),
	}
}

func proofRegistryTestHash(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
