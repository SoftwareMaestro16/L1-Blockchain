package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalResolutionPathV2RootToTarget(t *testing.T) {
	path, err := CanonicalResolutionPathV2("service.api.alice.aet")
	require.NoError(t, err)
	require.Equal(t, "service.api.alice.aet", path.TargetName)
	require.Equal(t, []string{"alice", "api", "service"}, path.Labels)
	require.Equal(t, []string{"alice.aet", "api.alice.aet", "service.api.alice.aet"}, path.Path)
	require.Len(t, path.PathHashes, 3)
	for i, name := range path.Path {
		hash, err := DomainRecordV2NameHash(name)
		require.NoError(t, err)
		require.Equal(t, hash, path.PathHashes[i])
	}

	_, err = CanonicalResolutionPathV2("Service.api.alice.aet")
	require.ErrorContains(t, err, "lowercase")
}

func TestDeterministicResolutionPathV2VerifiesPrimaryFreshnessAndExpiry(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)

	result, err := VerifyDeterministicResolutionPathV2(state, "alice.aet", DeterministicResolutionPathValidationV2{
		TargetType:	IdentityResolutionTargetPrimary,
		Height:		14,
		RecordTTL:	30,
		MaxRecordAge:	3,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"alice.aet"}, result.Path.Path)
	require.Equal(t, "alice.aet", result.Resolution.ResolverDomain)
	require.Equal(t, addr(2), result.Resolution.Record.Primary)

	_, err = VerifyDeterministicResolutionPathV2(state, "alice.aet", DeterministicResolutionPathValidationV2{
		TargetType:	IdentityResolutionTargetPrimary,
		Height:		14,
		RecordTTL:	30,
		MaxRecordAge:	1,
	})
	require.ErrorContains(t, err, "stale")

	_, err = VerifyDeterministicResolutionPathV2(state, "alice.aet", DeterministicResolutionPathValidationV2{
		TargetType:	IdentityResolutionTargetPrimary,
		Height:		state.Domains[0].ExpiryHeight,
		RecordTTL:	30,
	})
	require.ErrorContains(t, err, "expired")
}

func TestDeterministicResolutionPathV2RequiresParentOrDelegation(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)

	_, err = VerifyDeterministicResolutionPathV2(state, "service.api.alice.aet", DeterministicResolutionPathValidationV2{
		TargetType:	IdentityResolutionTargetPrimary,
		Height:		14,
		RecordTTL:	30,
	})
	require.ErrorContains(t, err, "requires parent domain or delegation")

	delegation, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeSubdomainCreate, []string{"create"}, 100, 2, "", 10)
	require.NoError(t, err)
	result, err := VerifyDeterministicResolutionPathV2(state, "service.api.alice.aet", DeterministicResolutionPathValidationV2{
		TargetType:	IdentityResolutionTargetPrimary,
		Height:		14,
		RecordTTL:	30,
		Delegations:	[]DelegationRecordV2{delegation},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"alice.aet", "api.alice.aet", "service.api.alice.aet"}, result.Path.Path)
	require.False(t, result.Steps[0].Delegated)
	require.True(t, result.Steps[1].Delegated)
	require.True(t, result.Steps[2].Delegated)
}

func TestDeterministicResolutionPathV2VerifiesServiceTarget(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	serviceKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	metadata, err := EncodeResolverMetadata([]ResolverMetadataEntry{{Key: serviceKey, Value: "https://rpc.alice.aet"}})
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2), Metadata: metadata}, 12)
	require.NoError(t, err)

	_, err = VerifyDeterministicResolutionPathV2(state, "alice.aet", DeterministicResolutionPathValidationV2{
		TargetType:	IdentityResolutionTargetService,
		TargetKey:	"rpc",
		Height:		14,
		RecordTTL:	30,
	})
	require.NoError(t, err)

	_, err = VerifyDeterministicResolutionPathV2(state, "alice.aet", DeterministicResolutionPathValidationV2{
		TargetType:	IdentityResolutionTargetService,
		TargetKey:	"grpc",
		Height:		14,
		RecordTTL:	30,
	})
	require.ErrorContains(t, err, "service target")
}

func TestBuildRecursiveResolutionProofV2UsesCanonicalRootToTargetPath(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)

	proof, err := BuildRecursiveResolutionProofV2(state, "aetra-local-1", "alice.aet", "api.alice.aet", 14, 30, nil)
	require.NoError(t, err)
	path, err := CanonicalResolutionPathV2("api.alice.aet")
	require.NoError(t, err)
	require.Equal(t, []string{"alice", "api"}, proof.PathLabels)
	require.Equal(t, path.PathHashes, proof.PathHashes)

	_, err = BuildRecursiveResolutionProofV2(state, "aetra-local-1", "bob.aet", "api.alice.aet", 14, 30, nil)
	require.ErrorContains(t, err, "root_name")
}
