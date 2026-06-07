package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)
	appHash, err := IdentityStateRoot(state)
	require.NoError(t, err)

	proof, err := BuildIdentityResolutionProofFormatV2(state, "aetra-local-1", appHash, "alice.aet", IdentityProofQueryResolvePrimary, 14, 30, addr(2))
	require.NoError(t, err)
	require.Equal(t, IdentityProofSchemaVersionV2, proof.ProofVersion)
	require.Equal(t, IdentityResolutionProofFormatV2FieldOrder[len(IdentityResolutionProofFormatV2FieldOrder)-1], "proof_commitment_hash")
	require.Equal(t, "aetra-local-1", proof.ChainID)
	require.Equal(t, uint64(14), proof.Height)
	require.Equal(t, appHash, proof.AppHash)
	require.Equal(t, "alice.aet", proof.Name)
	require.NotEmpty(t, proof.NormalizedNameProof)
	require.NotNil(t, proof.DomainRecord)
	require.NotNil(t, proof.DomainRecordProof)
	require.NotNil(t, proof.NFTBinding)
	require.NotNil(t, proof.NFTBindingProof)
	require.NotNil(t, proof.ResolverRecord)
	require.NotNil(t, proof.ResolverRecordProof)
	require.NotNil(t, proof.ReverseRecordOptional)
	require.NotNil(t, proof.ReverseRecordProofOptional)
	require.Equal(t, uint64(12), proof.RecordVersion)
	require.NoError(t, ValidateIdentityResolutionProofFormatV2(proof))

	encoded1, err := EncodeIdentityResolutionProofFormatV2(proof)
	require.NoError(t, err)
	encoded2, err := EncodeIdentityResolutionProofFormatV2(proof)
	require.NoError(t, err)
	require.True(t, bytes.Equal(encoded1, encoded2))
	require.Equal(t, ComputeIdentityResolutionProofCommitmentHashV2(proof), proof.ProofCommitmentHash)

	tampered := proof
	tampered.RecordVersion++
	require.ErrorContains(t, ValidateIdentityResolutionProofFormatV2(tampered), "commitment hash mismatch")
	tampered.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(tampered)
	require.NoError(t, ValidateIdentityResolutionProofFormatV2(tampered))
	require.NotEqual(t, proof.ProofCommitmentHash, tampered.ProofCommitmentHash)
}

func TestIdentityResolutionProofFormatV2NonExistenceProof(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	appHash, err := IdentityStateRoot(state)
	require.NoError(t, err)

	proof, err := BuildIdentityResolutionProofFormatV2(state, "aetra-local-1", appHash, "missing.aet", IdentityProofQueryDomainAbsent, 14, 30, nil)
	require.NoError(t, err)
	require.Nil(t, proof.DomainRecord)
	require.NotNil(t, proof.NonExistenceProofOptional)
	require.Nil(t, proof.ResolverRecord)
	require.Equal(t, uint64(1), proof.RecordVersion)
	require.NoError(t, ValidateIdentityResolutionProofFormatV2(proof))
}

func TestRecursiveResolutionProofV2EncodingCommitmentAndCache(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	path, err := CanonicalResolutionPathV2("api.alice.aet")
	require.NoError(t, err)
	pathHash, err := ComputeResolutionPathHashV2(path.Path)
	require.NoError(t, err)
	finalRecord, err := BuildUnifiedResolutionRecordV2(state, "api.alice.aet", 14, 30)
	require.NoError(t, err)
	resolvedHash, err := ComputeResolvedRecordHashV2(finalRecord)
	require.NoError(t, err)
	cache, err := NewResolutionCacheRecordV2("api.alice.aet", pathHash, resolvedHash, 24, ResolverRecordVersionV2(finalRecordToResolverRecord(t, state, "alice.aet")), 1, 1)
	require.NoError(t, err)

	proof, err := BuildRecursiveResolutionProofV2(state, "aetra-local-1", "alice.aet", "api.alice.aet", 14, 30, &cache)
	require.NoError(t, err)
	require.Equal(t, RecursiveResolutionProofV2FieldOrder[0], "proof_version")
	require.Equal(t, "alice.aet", proof.RootName)
	require.Equal(t, "api.alice.aet", proof.TargetName)
	require.Equal(t, []string{"alice", "api"}, proof.PathLabels)
	require.Equal(t, path.PathHashes, proof.PathHashes)
	require.Len(t, proof.PathHashes, 2)
	require.NotEmpty(t, proof.PathDomainRecords)
	require.NotEmpty(t, proof.PathResolverRecords)
	require.NotNil(t, proof.CacheRecordOptional)
	require.NoError(t, ValidateRecursiveResolutionProofV2(proof))

	encoded1, err := EncodeRecursiveResolutionProofV2(proof)
	require.NoError(t, err)
	encoded2, err := EncodeRecursiveResolutionProofV2(proof)
	require.NoError(t, err)
	require.True(t, bytes.Equal(encoded1, encoded2))
	require.Equal(t, ComputeRecursiveResolutionProofCommitmentHashV2(proof), proof.ProofCommitmentHash)

	resolutionProof, err := BuildIdentityResolutionProofFormatV2(state, "aetra-local-1", proof.FinalRecordProof.RootHash, "api.alice.aet", IdentityProofQueryResolvePrimary, 14, 30, nil)
	require.NoError(t, err)
	require.NotEqual(t, resolutionProof.ProofCommitmentHash, proof.ProofCommitmentHash)
}

func finalRecordToResolverRecord(t *testing.T, state IdentityState, name string) ResolverRecord {
	t.Helper()
	record, found := findResolver(state, name)
	require.True(t, found)
	return record
}
