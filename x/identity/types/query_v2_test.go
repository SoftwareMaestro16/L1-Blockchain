package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentityQueryServiceV2CoreQueries(t *testing.T) {
	state, service, delegation, auctionRecord := queryFixtureV2(t)
	aliceHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)

	domain := service.QueryDomain(aliceHash, true)
	require.Equal(t, IdentityQueryOK, domain.Code)
	require.Equal(t, "alice.aet", domain.Domain.Name)
	require.NotNil(t, domain.Proof)
	byName := service.QueryDomainByName("alice.aet", false)
	require.Equal(t, IdentityQueryOK, byName.Code)
	require.Equal(t, aliceHash, byName.Domain.NameHash)

	byOwner := service.QueryDomainsByOwner(addr(1), IdentityQueryPageRequestV2{Limit: 1})
	require.Equal(t, IdentityQueryOK, byOwner.Code)
	require.Len(t, byOwner.Domains, 1)
	require.GreaterOrEqual(t, byOwner.Page.Total, uint64(2))
	require.NotZero(t, byOwner.Page.NextOffset)

	binding := service.QueryDomainNFTBinding(aliceHash)
	require.Equal(t, IdentityQueryOK, binding.Code)
	require.Equal(t, aliceHash, binding.Binding.NameHash)

	resolver := service.QueryResolver("alice.aet", true)
	require.Equal(t, IdentityQueryOK, resolver.Code)
	require.Equal(t, addr(2), resolver.Resolver.PrimaryAddress)
	require.NotNil(t, resolver.Proof)
	primary := service.QueryResolvePrimary("svc.alice.aet")
	require.Equal(t, IdentityQueryOK, primary.Code)
	require.Equal(t, addr(2), primary.Address)
	target := service.QueryResolveTarget("alice.aet", ResolverKeyWallet)
	require.Equal(t, IdentityQueryOK, target.Code)
	require.Equal(t, addr(4), target.Address)
	serviceResp := service.QueryResolveService("alice.aet", "rpc")
	require.Equal(t, IdentityQueryOK, serviceResp.Code)
	require.Equal(t, "https://rpc.aet", serviceResp.Service.Endpoint)
	interfaceResp := service.QueryResolveInterface("alice.aet", "aw5")
	require.Equal(t, IdentityQueryOK, interfaceResp.Code)
	expectedDescriptorHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)
	require.Equal(t, expectedDescriptorHash, interfaceResp.Interface.Descriptor)
	route := service.QueryResolveRoute("alice.aet")
	require.Equal(t, IdentityQueryOK, route.Code)
	require.Equal(t, "CONTRACT_ZONE", route.Route.ZoneID)

	reverse := service.QueryReverse(addr(2))
	require.Equal(t, IdentityQueryOK, reverse.Code)
	require.False(t, reverse.Reverse.Verified)
	verified := service.QueryVerifiedReverse(addr(2), nil)
	require.Equal(t, IdentityQueryOK, verified.Code)
	require.True(t, verified.Reverse.Verified)

	subdomains := service.QuerySubdomains("alice.aet", IdentityQueryPageRequestV2{Limit: 1})
	require.Equal(t, IdentityQueryOK, subdomains.Code)
	require.Len(t, subdomains.Subdomains, 1)
	require.Equal(t, "api.alice.aet", subdomains.Subdomains[0].Name)
	delegations := service.QueryDelegations(delegation.NameHash, IdentityQueryPageRequestV2{Limit: 1})
	require.Equal(t, IdentityQueryOK, delegations.Code)
	require.Len(t, delegations.Delegations, 1)
	require.Equal(t, addr(7), delegations.Delegations[0].Delegate)

	auction := service.QueryAuction(auctionRecord.AuctionID, "")
	require.Equal(t, IdentityQueryOK, auction.Code)
	require.Equal(t, auctionRecord.AuctionID, auction.Auction.AuctionID)
	proof := service.QueryResolutionProof("svc.alice.aet")
	require.Equal(t, IdentityQueryOK, proof.Code)
	require.NotNil(t, proof.Proof)
	recursiveProof := service.QueryRecursiveResolutionProof("svc.alice.aet")
	require.Equal(t, IdentityQueryOK, recursiveProof.Code)
	require.Equal(t, "alice.aet", recursiveProof.ProofResult.ResolverDomain)
	lifecycle := service.QueryDomainLifecycle("alice.aet")
	require.Equal(t, IdentityQueryOK, lifecycle.Code)
	require.Equal(t, DomainLifecycleActive, lifecycle.Lifecycle)
	params := service.QueryIdentityParams()
	require.Equal(t, IdentityQueryOK, params.Code)
	require.Equal(t, DefaultIdentityParams(), *params.Params)

	require.NoError(t, state.Validate())
}

func TestIdentityQueryServiceV2ExplicitFailureCodes(t *testing.T) {
	_, service, _, _ := queryFixtureV2(t)

	missing := service.QueryDomainByName("missing.aet", true)
	require.Equal(t, IdentityQueryNotFound, missing.Code)
	require.Equal(t, IdentityLightClientErrDomainNotFound, missing.FailureCode)
	require.NotEmpty(t, missing.Error)
	require.NotNil(t, missing.AbsenceProof)
	require.Nil(t, missing.Domain)
	require.Nil(t, missing.Target)
	invalid := service.QueryDomain("bad-hash", false)
	require.Equal(t, IdentityQueryInvalidRequest, invalid.Code)
	require.Equal(t, IdentityLightClientErrInvalidName, invalid.FailureCode)
	require.NotEmpty(t, invalid.Error)
	serviceMissing := service.QueryResolveService("alice.aet", "missing")
	require.Equal(t, IdentityQueryNotFound, serviceMissing.Code)
	require.Equal(t, IdentityLightClientErrTargetNotFound, serviceMissing.FailureCode)
	require.Equal(t, uint64(14), serviceMissing.RecordVersion)
	require.Nil(t, serviceMissing.Target)
	require.Nil(t, serviceMissing.Service)
	invalidReverse := service.QueryVerifiedReverse(addr(4), nil)
	require.Equal(t, IdentityQueryNotFound, invalidReverse.Code)
}

func queryFixtureV2(t *testing.T) (IdentityState, IdentityQueryServiceV2, DelegationRecordV2, AuctionRecordV2) {
	t.Helper()
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	commitment, err := ComputeRegistrationCommitment("bob.aet", addr(1), "bob-salt")
	require.NoError(t, err)
	state, err = CommitDomainRegistration(state, "bob.aet", addr(1), commitment, 11)
	require.NoError(t, err)
	state, _, err = RevealRegisterDomain(state, "bob.aet", addr(1), "bob-salt", 12)
	require.NoError(t, err)
	state, _, err = IssueSubdomain(state, "alice.aet", "api", addr(1), addr(5), true, 13)
	require.NoError(t, err)
	interfaceKey, err := ResolverMetadataInterfaceKey("aw5")
	require.NoError(t, err)
	serviceKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	metadata, err := EncodeResolverMetadata([]ResolverMetadataEntry{
		{Key: ResolverMetadataRouteZone, Value: "CONTRACT_ZONE"},
		{Key: ResolverMetadataRouteShard, Value: "0:1"},
		{Key: ResolverMetadataRouteVM, Value: "AVM"},
		{Key: ResolverMetadataRouteEntrypoint, Value: "swap"},
		{Key: interfaceKey, Value: "wallet-v1"},
		{Key: serviceKey, Value: "https://rpc.aet"},
	})
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Contract:	addr(3),
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(4),
		},
		Metadata:	metadata,
	}, 14)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 15)
	require.NoError(t, err)
	state, auction, err := StartSealedAuction(state, "market.aet", 16)
	require.NoError(t, err)
	auctionRecord, err := BuildAuctionRecordV2(auction, 1, "domain.fees")
	require.NoError(t, err)
	delegation, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeSubdomainCreate, []string{"create"}, 100, 1, "", 16)
	require.NoError(t, err)
	service := NewIdentityQueryServiceV2(IdentityQueryContextV2{
		State:		state,
		Height:		20,
		DefaultTTL:	30,
		Delegations:	[]DelegationRecordV2{delegation},
	})
	return state, service, delegation, auctionRecord
}
