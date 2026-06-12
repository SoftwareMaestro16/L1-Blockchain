package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityLightweightResolutionFlowV2VerifiesCachesAndChecks(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolvePrimary, 14, 30, nil)

	result, err := ResolveIdentityLightweightV2(IdentityLightweightResolutionRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		TargetType:		IdentityResolutionTargetPrimary,
		CurrentHeight:		20,
		FreshnessThreshold:	10,
	})
	require.NoError(t, err)
	require.True(t, result.Verified)
	require.Equal(t, IdentityLightClientProofRequestResolutionV2, result.ProofRequest)
	require.Equal(t, "alice.aet", result.NormalizedName)
	require.Equal(t, proof.NameHash, result.NameHash)
	require.Equal(t, addr(2), result.Target.Address)
	require.NotNil(t, result.CacheMetadata)
	require.Equal(t, proof.Height, result.CacheMetadata.ProofHeight)
	require.Equal(t, proof.Height, result.CacheMetadata.TrustedHeaderHeight)
	require.Equal(t, uint64(24), result.CacheMetadata.FreshUntilHeight)
	require.Equal(t, uint64(12), result.CacheMetadata.Key.RecordVersion)
	require.Equal(t, IdentityCacheLayerWalletVerifiedV2, result.CacheMetadata.Key.Layer)
	require.Empty(t, result.FailureCode)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckHeaderTrustV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckNameNormalizationV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckDomainProofV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckNFTBindingV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckOwnershipV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckResolverProofV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckTargetExistsV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckTTLExpiryV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckDelegationProofV2, IdentityLightClientCheckSkippedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckReverseConsistencyV2, IdentityLightClientCheckSkippedV2)
}

func TestIdentityLightweightResolutionFlowV2RecursiveAndReverseChecks(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = IssueSubdomain(state, "alice.aet", "api", addr(1), addr(1), true, 13)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 14)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveReverse, 15, 30, addr(2))
	recursive, err := BuildRecursiveResolutionProofV2(state, "aetra-local-1", "alice.aet", "alice.aet", 15, 30, nil)
	require.NoError(t, err)

	result, err := ResolveIdentityLightweightV2(IdentityLightweightResolutionRequestV2{
		ExpectedChainID:		"aetra-local-1",
		RequestedName:			"alice.aet",
		TrustedHeader:			trustedHeaderForProofV2(proof),
		Proof:				proof,
		RecursiveProof:			&recursive,
		TargetType:			IdentityResolutionTargetPrimary,
		RequireReverseResolution:	true,
	})
	require.NoError(t, err)
	require.True(t, result.Verified)
	require.Equal(t, IdentityLightClientProofRequestRecursiveResolutionV2, result.ProofRequest)
	require.NotNil(t, result.CacheMetadata)
	require.Equal(t, recursive.ProofCommitmentHash, result.CacheMetadata.Key.PathHash)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckDelegationProofV2, IdentityLightClientCheckPassedV2)
	requireLightClientCheckStatusV2(t, result, IdentityLightClientCheckReverseConsistencyV2, IdentityLightClientCheckPassedV2)
}

func TestIdentityLightweightResolutionFlowV2FailureChecklist(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolvePrimary, 14, 30, nil)

	stale, err := ResolveIdentityLightweightV2(IdentityLightweightResolutionRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		TargetType:		IdentityResolutionTargetPrimary,
		CurrentHeight:		100,
	})
	require.Error(t, err)
	require.False(t, stale.Verified)
	require.Equal(t, IdentityLightClientErrRecordStale, stale.FailureCode)
	requireLightClientCheckStatusV2(t, stale, IdentityLightClientCheckTTLExpiryV2, IdentityLightClientCheckFailedV2)
	require.Nil(t, stale.CacheMetadata)

	badChain, err := ResolveIdentityLightweightV2(IdentityLightweightResolutionRequestV2{
		ExpectedChainID:	"other-chain",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		TargetType:		IdentityResolutionTargetPrimary,
	})
	require.Error(t, err)
	require.Equal(t, IdentityLightClientErrProofInvalid, badChain.FailureCode)
	requireLightClientCheckStatusV2(t, badChain, IdentityLightClientCheckChainIDV2, IdentityLightClientCheckFailedV2)

	missingTarget, err := ResolveIdentityLightweightV2(IdentityLightweightResolutionRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			proof,
		TargetType:		IdentityResolutionTargetService,
		TargetKey:		"rpc",
	})
	require.Error(t, err)
	require.Equal(t, IdentityLightClientErrTargetNotFound, missingTarget.FailureCode)
	requireLightClientCheckStatusV2(t, missingTarget, IdentityLightClientCheckTargetExistsV2, IdentityLightClientCheckFailedV2)
}

func requireLightClientCheckStatusV2(t *testing.T, result IdentityLightweightResolutionResultV2, name IdentityLightClientCheckNameV2, status IdentityLightClientCheckStatusV2) {
	t.Helper()
	for _, check := range result.Checks {
		if check.Name == name {
			require.Equal(t, status, check.Status, "check %s", name)
			return
		}
	}
	t.Fatalf("missing light-client check %s", name)
}
