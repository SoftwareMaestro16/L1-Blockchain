package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentityDNSPatchPreservesAndDeletesRecords(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)

	state, record, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Contract:	addr(3),
		ZoneEndpoint:	"contract-zone/0:1",
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(4),
		},
		Metadata:	[]byte("profile-v1"),
	}, 12)
	require.NoError(t, err)
	require.Equal(t, addr(2), record.Primary)
	require.Equal(t, addr(3), record.Contract)
	require.Equal(t, addr(4), record.Records[ResolverKeyWallet])

	state, record, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Records: map[string]sdk.AccAddress{
			ResolverKeyDEX: addr(5),
		},
		DeleteRecords:		[]string{ResolverKeyWallet},
		ClearMetadata:		true,
		ClearZoneEndpoint:	true,
	}, 13)
	require.NoError(t, err)
	require.Equal(t, addr(2), record.Primary)
	require.Equal(t, addr(3), record.Contract)
	require.Equal(t, addr(5), record.Records[ResolverKeyDEX])
	require.NotContains(t, record.Records, ResolverKeyWallet)
	require.Empty(t, record.Metadata)
	require.Empty(t, record.ZoneEndpoint)
	require.NoError(t, state.Validate())
}

func TestIdentityDNSRecursiveResolutionForUnregisteredSubdomain(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)

	state, record, err := PatchIdentityResolver(state, "api.dex.alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Contract:	addr(3),
		ZoneEndpoint:	"contract-zone/0:1",
	}, 12)
	require.NoError(t, err)
	require.Equal(t, "api.dex.alice.aet", record.Domain)

	resolved, err := ResolveIdentityAddressRecursive(state, "API.DEX.ALICE.AET", 13)
	require.NoError(t, err)
	require.Equal(t, addr(2), resolved)

	route, err := ResolveIdentityRoute(state, "api.dex.alice.aet", ResolverKeyContract, 13)
	require.NoError(t, err)
	require.Equal(t, "api.dex.alice.aet", route.ResolverDomain)
	require.Equal(t, "alice.aet", route.AuthorityName)
	require.Equal(t, addr(3), route.Address)
	require.Equal(t, "contract-zone/0:1", route.ZoneEndpoint)
}

func TestIdentityDNSRecursiveResolutionStopsAtRegisteredSubdomainBoundary(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := IssueSubdomain(state, "alice.aet", "dex", addr(1), addr(3), false, 11)
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)

	_, err = ResolveIdentityAddressRecursive(state, "api.dex.alice.aet", 13)
	require.ErrorContains(t, err, "not resolved")

	_, _, err = PatchIdentityResolver(state, "api.dex.alice.aet", addr(1), ResolverPatch{Primary: addr(4)}, 13)
	require.ErrorContains(t, err, "requires owner")

	state, _, err = PatchIdentityResolver(state, "api.dex.alice.aet", addr(3), ResolverPatch{Primary: addr(4)}, 13)
	require.NoError(t, err)
	resolved, err := ResolveIdentityAddressRecursive(state, "api.dex.alice.aet", 14)
	require.NoError(t, err)
	require.Equal(t, addr(4), resolved)
}

func TestIdentityDNSTransferUpdatesSubtreeResolverOwnership(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "api.alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state.PendingResolverUpdates = []ResolverUpdateIntent{{Domain: "api.alice.aet", Actor: addr(1), Nonce: 1}}
	require.NoError(t, state.Validate())

	next, _, err := TransferDomainNFT(state, "alice.aet", addr(1), addr(9), 20)
	require.NoError(t, err)
	require.Empty(t, next.PendingResolverUpdates)
	resolver, found := findResolverByNormalizedDomain(next, "api.alice.aet")
	require.True(t, found)
	require.Equal(t, addr(9), resolver.Owner)

	_, _, err = PatchIdentityResolver(next, "api.alice.aet", addr(1), ResolverPatch{Primary: addr(3)}, 21)
	require.ErrorContains(t, err, "requires owner")
	_, _, err = PatchIdentityResolver(next, "api.alice.aet", addr(9), ResolverPatch{Primary: addr(3)}, 21)
	require.NoError(t, err)
}

func TestIdentityDNSValidateRejectsResolverOwnerMismatch(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state.Resolvers = append(state.Resolvers, ResolverRecord{
		Domain:		"alice.aet",
		Owner:		addr(9),
		Primary:	addr(2),
		UpdatedAtUnix:	12,
	})

	require.ErrorContains(t, state.Validate(), "resolver owner must match registry owner")
}
