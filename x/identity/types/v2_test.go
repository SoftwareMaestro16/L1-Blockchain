package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentityStoreV2KeysAndBlockSTMAccessSets(t *testing.T) {
	key, err := IdentityDomainStoreKey("API.DEX.Alice.AET")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(key, IdentityStoreV2DomainPrefix+"/"))
	require.True(t, strings.HasSuffix(key, "/aet/alice/dex/api"))

	alice, err := IdentityResolverPatchAccessSet("api.alice.aet")
	require.NoError(t, err)
	bob, err := IdentityResolverPatchAccessSet("api.bob.aet")
	require.NoError(t, err)
	require.False(t, alice.Conflicts(bob))

	otherAlice, err := IdentityResolverPatchAccessSet("api.alice.aet")
	require.NoError(t, err)
	require.True(t, alice.Conflicts(otherAlice))

	readAlice, err := IdentityResolutionAccessSet("api.alice.aet")
	require.NoError(t, err)
	require.True(t, alice.Conflicts(readAlice))
}

func TestUnifiedResolverMetadataAndNamedExecution(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	interfaceKey, err := ResolverMetadataInterfaceKey("aw5")
	require.NoError(t, err)
	metadata, err := EncodeResolverMetadata([]ResolverMetadataEntry{
		{Key: ResolverMetadataRouteZone, Value: "CONTRACT_ZONE"},
		{Key: ResolverMetadataRouteShard, Value: "0:1"},
		{Key: ResolverMetadataRouteVM, Value: "AVM"},
		{Key: ResolverMetadataRouteEntrypoint, Value: "swap"},
		{Key: interfaceKey, Value: "1"},
	})
	require.NoError(t, err)

	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Contract:	addr(3),
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(4),
		},
		Metadata:	metadata,
	}, 12)
	require.NoError(t, err)

	view, err := BuildUnifiedResolverView(state, "alice.aet", 13)
	require.NoError(t, err)
	require.Equal(t, addr(2), view.Primary)
	require.Equal(t, addr(3), view.Contract)
	require.Equal(t, "CONTRACT_ZONE", view.Route.ZoneID)
	require.Equal(t, "0:1", view.Route.ShardID)
	require.Equal(t, "AVM", view.Route.VM)
	require.Equal(t, "swap", view.Route.Entrypoint)

	sendTarget, err := ResolveNamedExecutionTarget(state, NamedExecutionRequest{
		Kind:		NamedExecutionSend,
		Name:		"alice.aet",
		RecordKey:	ResolverKeyWallet,
	}, 13)
	require.NoError(t, err)
	require.Equal(t, addr(4), sendTarget.Address)

	invokeTarget, err := ResolveNamedExecutionTarget(state, NamedExecutionRequest{
		Kind:		NamedExecutionInvoke,
		Name:		"alice.aet",
		InterfaceID:	"aw5",
		Method:		"swap",
		PayloadHash:	identityHash("payload", "swap"),
	}, 13)
	require.NoError(t, err)
	require.Equal(t, addr(3), invokeTarget.Contract)
	require.Equal(t, "swap", invokeTarget.Route.Entrypoint)
}

func TestIdentityResolutionProofVerifiesRecursiveLookupAndAbsence(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)

	proof, err := BuildIdentityResolutionProof(state, "api.alice.aet", 13)
	require.NoError(t, err)
	require.Equal(t, "api.alice.aet", proof.QueryDomain)
	require.Equal(t, "alice.aet", proof.ResolverDomain)
	require.Len(t, proof.Candidates, 2)
	require.NotNil(t, proof.Candidates[0].DomainAbsence)
	require.NotNil(t, proof.Candidates[0].ResolverAbsence)
	require.NotNil(t, proof.Candidates[1].DomainProof)
	require.NotNil(t, proof.Candidates[1].ResolverProof)

	resolution, err := VerifyIdentityResolutionProof(proof, 13)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", resolution.AuthorityDomain.Name)
	require.Equal(t, addr(2), resolution.Record.Primary)
	require.Equal(t, uint8(1), resolution.Depth)

	proof.Resolver.Primary = addr(9)
	_, err = VerifyIdentityResolutionProof(proof, 13)
	require.ErrorContains(t, err, "resolver proof value mismatch")
}

func TestIdentityResolutionProofRejectsWrongCandidateKey(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof, err := BuildIdentityResolutionProof(state, "api.alice.aet", 13)
	require.NoError(t, err)

	proof.Candidates[0].ResolverAbsence.Key = proof.Candidates[0].DomainAbsence.Key
	_, err = VerifyIdentityResolutionProof(proof, 13)
	require.ErrorContains(t, err, "resolver absence key mismatch")
}
