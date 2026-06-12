package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverSpoofingPreventionUnauthorizedUpdatesV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)

	_, _, err := PatchIdentityResolver(state, domain.Name, addr(9), ResolverPatch{Primary: addr(2)}, 12)
	require.ErrorContains(t, err, "requires owner")

	v2 := mustDomainRecordV2FromSpec(t, domain, 12)
	require.NoError(t, ValidateResolverUpdateAuthorizationV2(v2, addr(1), nil, ResolverKeyPrimary, 12))
	require.ErrorContains(t, ValidateResolverUpdateAuthorizationV2(v2, addr(9), nil, ResolverKeyPrimary, 12), "owner or delegated")

	delegate, err := NewDelegationRecordV2(domain.Name, addr(7), DelegationScopeResolverUpdate, []string{ResolverKeyPrimary}, 100, 0, ResolverKeyPrimary, 12)
	require.NoError(t, err)
	require.NoError(t, ValidateResolverUpdateAuthorizationV2(v2, addr(7), &delegate, ResolverKeyPrimary, 13))
	require.ErrorContains(t, ValidateResolverUpdateAuthorizationV2(v2, addr(8), &delegate, ResolverKeyPrimary, 13), "delegate mismatch")
	require.ErrorContains(t, ValidateResolverUpdateAuthorizationV2(v2, addr(7), &delegate, ResolverKeyMetadata, 13), "permission")
}

func TestResolverSpoofingPreventionResolverOwnerMustMatchActiveDomainV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state.Resolvers = []ResolverRecord{{
		Domain:		domain.Name,
		Owner:		addr(9),
		Primary:	addr(2),
		UpdatedAtUnix:	12,
	}}
	require.ErrorContains(t, state.Validate(), "resolver owner must match registry owner")

	state.Resolvers[0].Owner = addr(1)
	require.NoError(t, state.Validate())
}

func TestResolverSpoofingPreventionInterfaceHashMismatchV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	interfaceKey, err := ResolverMetadataInterfaceKey("wallet")
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Metadata:	mustMetadataV2(t, []ResolverMetadataEntry{{Key: interfaceKey, Value: `{"version":"v1"}`}}),
	}, 12)
	require.NoError(t, err)
	record, err := BuildUnifiedResolutionRecordV2(state, "alice.aet", 14, 30)
	require.NoError(t, err)
	goodHash := record.InterfaceDescriptors[0].SchemaHash
	badHash, err := InterfaceDescriptorHashV2(`{"version":"v2"}`)
	require.NoError(t, err)
	require.NotEqual(t, goodHash, badHash)

	_, err = VerifyIdentityInterfaceDescriptorForInvokeV2(record, "wallet", badHash)
	require.ErrorContains(t, err, "hash mismatch")

	inline := `{"type":"wallet","version":"v1"}`
	inlineHash, err := InterfaceDescriptorHashV2(inline)
	require.NoError(t, err)
	record.InterfaceDescriptors = []InterfaceDescriptorV2{{
		InterfaceID:		"wallet",
		SchemaHash:		badHash,
		SchemaInlineOptional:	inline,
		Version:		"v1",
		RenderPolicy:		IdentityRenderPolicyConfirmV2,
	}}
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "inline schema hash mismatch")
	record.InterfaceDescriptors[0].SchemaHash = inlineHash
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
}

func TestResolverSpoofingPreventionReverseRecordsRequireForwardConsistencyV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, domain.Name, addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)

	_, _, err = SetIdentityReverse(state, addr(3), addr(3), domain.Name, 13)
	require.ErrorContains(t, err, "does not point")

	spoofed, err := NewReverseResolutionRecordV2(addr(3), domain.Name, false, 13, domain.ExpiryHeight)
	require.NoError(t, err)
	require.NoError(t, ValidateReverseResolutionRecordV2(state, spoofed, 14, nil), "claimed reverse is allowed only as unverified")
	spoofed.Verified = true
	require.ErrorContains(t, ValidateReverseResolutionRecordV2(state, spoofed, 14, nil), "forward primary or authorized alias")

	valid, err := NewReverseResolutionRecordV2(addr(2), domain.Name, true, 13, domain.ExpiryHeight)
	require.NoError(t, err)
	require.NoError(t, ValidateReverseResolutionRecordV2(state, valid, 14, nil))
}

func TestResolverSpoofingPreventionServiceEndpointDisplayPolicyV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	rpcKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Metadata:	mustMetadataV2(t, []ResolverMetadataEntry{{Key: rpcKey, Value: "https://rpc.aet"}}),
	}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveRecord, 14, 30, nil)

	result, err := BuildIdentityServiceDiscoveryV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, result.DisplayPolicy.DisplayEndpoint)
	require.True(t, result.DisplayPolicy.DisplayAsVerifiedService)
	require.False(t, result.DisplayPolicy.DisplayAsVerifiedOwnership)
	require.False(t, result.DisplayPolicy.DisplayMetadataAsOwnership)
	require.Equal(t, "registry_nft_binding", result.DisplayPolicy.OwnershipVerificationSource)
	require.NoError(t, ValidateIdentityServiceEndpointDisplayPolicyV2(result.DisplayPolicy))

	forged := result.DisplayPolicy
	forged.DisplayAsVerifiedOwnership = true
	require.ErrorContains(t, ValidateIdentityServiceEndpointDisplayPolicyV2(forged), "must not be displayed")
}

func TestResolverSpoofingPreventionNormalizationRejectsSpoofPatternsV2(t *testing.T) {
	for _, name := range []string{
		"аlice.aet",
		"alice..aet",
		"alice.aet.evil",
		"ALICE.aet",
	} {
		_, err := NormalizeAETDomain(name)
		if name == "ALICE.aet" {
			require.NoError(t, err)
			continue
		}
		require.Error(t, err, name)
	}
}

func mustDomainRecordV2FromSpec(t *testing.T, domain Domain, height uint64) DomainRecordV2 {
	t.Helper()
	nameHash, err := DomainRecordV2NameHash(domain.Name)
	require.NoError(t, err)
	parentHash, err := DomainRecordV2ParentNameHash(domain.Name)
	require.NoError(t, err)
	return DomainRecordV2{
		Name:			domain.Name,
		NameHash:		nameHash,
		NormalizedName:		domain.Name,
		ParentNameHash:		parentHash,
		TLD:			DomainTLD,
		Owner:			cloneSpecAddress(domain.Owner),
		ExpiryHeight:		domain.ExpiryHeight,
		RenewalStartHeight:	domain.ExpiryHeight - DefaultRenewalWindowBlocks,
		NFTClassID:		"identity-domain",
		NFTItemID:		domain.NFTID,
		Status:			DomainRecordV2Active,
		CreatedAtHeight:	domain.RegisteredHeight,
		UpdatedAtHeight:	height,
		Version:		1,
	}
}
