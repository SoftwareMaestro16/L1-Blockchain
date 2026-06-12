package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityResolutionProofLightClientV2ReturnsVerifiedPrimary(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolvePrimary, 14, 30, nil)

	target, err := VerifyIdentityResolutionProofLightClientV2(IdentityLightClientVerificationRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		TargetType:		IdentityResolutionTargetPrimary,
		CurrentHeight:		20,
	})
	require.NoError(t, err)
	require.Equal(t, "alice.aet", target.Name)
	require.Equal(t, proof.NameHash, target.NameHash)
	require.Equal(t, addr(2), target.Address)
	require.Equal(t, uint64(12), target.RecordVersion)
	require.Equal(t, uint64(44), target.FreshUntilHeight)
}

func TestIdentityResolutionProofLightClientV2FailureCodes(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolvePrimary, 14, 30, nil)

	base := IdentityLightClientVerificationRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		TargetType:		IdentityResolutionTargetPrimary,
		CurrentHeight:		20,
	}

	invalidName := base
	invalidName.RequestedName = "Alice.aet"
	requireLightClientCodeV2(t, invalidName, IdentityLightClientErrInvalidName)

	badNormalization := base
	badNormalization.NormalizationVersion = NameNormalizationVersionV2 + 1
	requireLightClientCodeV2(t, badNormalization, IdentityLightClientErrUnsupportedNormalizationVersion)

	untrustedHeight := base
	untrustedHeight.TrustedHeader.Height++
	requireLightClientCodeV2(t, untrustedHeight, IdentityLightClientErrProofHeightUntrusted)

	expired := base
	expired.Proof = cloneLightClientProofForTestV2(proof)
	expired.Proof.DomainRecord.ExpiryHeight = expired.Proof.Height
	expired.Proof.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(expired.Proof)
	requireLightClientCodeV2(t, expired, IdentityLightClientErrDomainExpired)

	brokenNFT := base
	brokenNFT.Proof = cloneLightClientProofForTestV2(proof)
	brokenNFT.Proof.NFTBinding.Owner = addr(9)
	brokenNFT.Proof.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(brokenNFT.Proof)
	requireLightClientCodeV2(t, brokenNFT, IdentityLightClientErrNFTBindingMismatch)

	stale := base
	stale.CurrentHeight = proof.ResolverRecord.UpdatedAtHeight + proof.ResolverRecord.RecordTTL + 1
	requireLightClientCodeV2(t, stale, IdentityLightClientErrRecordStale)

	missingTarget := base
	missingTarget.TargetType = IdentityResolutionTargetService
	missingTarget.TargetKey = "rpc"
	requireLightClientCodeV2(t, missingTarget, IdentityLightClientErrTargetNotFound)
}

func TestIdentityResolutionProofLightClientV2ReverseConsistency(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveReverse, 14, 30, addr(2))

	base := IdentityLightClientVerificationRequestV2{
		ExpectedChainID:		"aetra-local-1",
		RequestedName:			"alice.aet",
		TrustedHeader:			trustedHeaderForProofV2(proof),
		Proof:				proof,
		TargetType:			IdentityResolutionTargetPrimary,
		RequireReverseResolution:	true,
	}
	_, err = VerifyIdentityResolutionProofLightClientV2(base)
	require.NoError(t, err)

	base.Proof.ReverseRecordOptional.Verified = false
	base.Proof.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(base.Proof)
	requireLightClientCodeV2(t, base, IdentityLightClientErrReverseNotVerified)
}

func TestIdentityResolutionProofLightClientV2VerifiesRecursivePath(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolvePrimary, 14, 30, nil)
	recursive, err := BuildRecursiveResolutionProofV2(state, "aetra-local-1", "alice.aet", "alice.aet", 14, 30, nil)
	require.NoError(t, err)

	_, err = VerifyIdentityResolutionProofLightClientV2(IdentityLightClientVerificationRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		RecursiveProof:		&recursive,
		TargetType:		IdentityResolutionTargetPrimary,
	})
	require.NoError(t, err)

	recursive.PathDomainRecords = nil
	req := IdentityLightClientVerificationRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		RecursiveProof:		&recursive,
		TargetType:		IdentityResolutionTargetPrimary,
	}
	requireLightClientCodeV2(t, req, IdentityLightClientErrProofInvalid)
}

func TestIdentityResolutionProofLightClientV2SkipsDisabledContractTargets(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:	nameHash,
		Owner:		addr(1),
		ContractTargets: []ContractTargetV2{{
			TargetID:		"swap",
			ContractAddress:	addr(3),
			Enabled:		false,
			UpdatedAtHeight:	12,
		}},
		RecordVersion:		1,
		RecordTTL:		30,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))

	_, err = lightClientTargetFromRecordV2(record, IdentityLightClientVerificationRequestV2{
		TargetType:	IdentityResolutionTargetContract,
		TargetKey:	"swap",
	})
	code, ok := IdentityLightClientFailureCodeFromErrorV2(err)
	require.True(t, ok)
	require.Equal(t, IdentityLightClientErrTargetNotFound, code)
}

func buildLightClientFormatProofV2(t *testing.T, state IdentityState, name string, queryType IdentityProofQueryTypeV2, height uint64, ttl uint64, reverseAddress []byte) IdentityResolutionProofFormatV2 {
	t.Helper()
	appHash, err := IdentityStateRoot(state)
	require.NoError(t, err)
	proof, err := BuildIdentityResolutionProofFormatV2(state, "aetra-local-1", appHash, name, queryType, height, ttl, reverseAddress)
	require.NoError(t, err)
	return proof
}

func trustedHeaderForProofV2(proof IdentityResolutionProofFormatV2) IdentityTrustedHeaderV2 {
	return IdentityTrustedHeaderV2{
		ChainID:	proof.ChainID,
		Height:		proof.Height,
		AppHash:	proof.AppHash,
		Trusted:	true,
	}
}

func cloneLightClientProofForTestV2(proof IdentityResolutionProofFormatV2) IdentityResolutionProofFormatV2 {
	out := proof
	if proof.DomainRecord != nil {
		copied := *proof.DomainRecord
		out.DomainRecord = &copied
	}
	if proof.DomainRecordProof != nil {
		copied := *proof.DomainRecordProof
		out.DomainRecordProof = &copied
	}
	if proof.NFTBinding != nil {
		copied := *proof.NFTBinding
		out.NFTBinding = &copied
	}
	if proof.NFTBindingProof != nil {
		copied := *proof.NFTBindingProof
		out.NFTBindingProof = &copied
	}
	if proof.ResolverRecord != nil {
		copied := *proof.ResolverRecord
		out.ResolverRecord = &copied
	}
	if proof.ResolverRecordProof != nil {
		copied := *proof.ResolverRecordProof
		out.ResolverRecordProof = &copied
	}
	if proof.ReverseRecordOptional != nil {
		copied := *proof.ReverseRecordOptional
		out.ReverseRecordOptional = &copied
	}
	if proof.ReverseRecordProofOptional != nil {
		copied := *proof.ReverseRecordProofOptional
		out.ReverseRecordProofOptional = &copied
	}
	if proof.NonExistenceProofOptional != nil {
		copied := *proof.NonExistenceProofOptional
		out.NonExistenceProofOptional = &copied
	}
	return out
}

func requireLightClientCodeV2(t *testing.T, request IdentityLightClientVerificationRequestV2, code IdentityLightClientFailureCodeV2) {
	t.Helper()
	_, err := VerifyIdentityResolutionProofLightClientV2(request)
	require.Error(t, err)
	actual, ok := IdentityLightClientFailureCodeFromErrorV2(err)
	require.True(t, ok, "expected coded error, got %T: %v", err, err)
	require.Equal(t, code, actual)
}
