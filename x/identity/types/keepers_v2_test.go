package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentityKeeperV2LifecycleAndNFTInvariant(t *testing.T) {
	keeper := NewIdentityKeeperV2()
	state := EmptyIdentityState(DefaultIdentityParams())
	commitment, err := ComputeRegistrationCommitment("alice.aet", addr(1), "salt")
	require.NoError(t, err)

	state, err = keeper.CommitRegistration(state, "alice", addr(1), commitment, 10)
	require.NoError(t, err)
	state, domain, err := keeper.RevealRegistration(state, "alice", addr(1), "salt", 11)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", domain.Name)
	require.NoError(t, ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{State: state, Height: 12}))

	state, renewed, err := keeper.Renew(state, "alice.aet", addr(1), 12)
	require.NoError(t, err)
	require.Greater(t, renewed.ExpiryHeight, domain.ExpiryHeight)
	state, transferred, err := keeper.Transfer(state, "alice.aet", addr(1), addr(2), 13)
	require.NoError(t, err)
	require.Equal(t, addr(2), transferred.Owner)
	require.NoError(t, ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{State: state, Height: 14}))

	broken := state.Clone()
	broken.DomainNFTs[0].Owner = addr(9)
	require.ErrorContains(t, ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{State: broken, Height: 14}), "NFT")
}

func TestResolverKeeperV2AuthorizationReverseAndTTL(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	keeper := NewResolverKeeperV2()
	_, _, err := keeper.UpdateUnifiedRecord(state, "alice.aet", addr(9), ResolverPatch{Primary: addr(2)}, 12, 30)
	require.ErrorContains(t, err, "requires owner")

	state, unified, err := keeper.UpdateUnifiedRecord(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(3),
		},
	}, 12, 30)
	require.NoError(t, err)
	require.NoError(t, keeper.ValidateVersionedTTL(unified, unified.RecordVersion, 40))
	require.ErrorContains(t, keeper.ValidateVersionedTTL(unified, unified.RecordVersion+1, 40), "version mismatch")
	require.ErrorContains(t, keeper.ValidateVersionedTTL(unified, unified.RecordVersion, 43), "ttl expired")

	reverse, err := NewReverseResolutionRecordV2(addr(3), "alice.aet", true, 13, 100)
	require.NoError(t, err)
	require.NoError(t, keeper.VerifyReverseRecord(state, reverse, 14, []string{ResolverKeyWallet}))
	require.ErrorContains(t, keeper.VerifyReverseRecord(state, reverse, 14, nil), "forward primary or authorized alias")
}

func TestDelegationKeeperV2ScopesSubdomainsAndZoneControl(t *testing.T) {
	keeper := NewDelegationKeeperV2()
	subdomain, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeSubdomainCreate, []string{"create"}, 100, 2, "", 10)
	require.NoError(t, err)
	require.NoError(t, keeper.EnforceSubdomainCreate(subdomain, "api", 1, 20))
	require.ErrorContains(t, keeper.EnforceSubdomainCreate(subdomain, "api", 3, 20), "subtree limit")

	zone, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeZoneAdmin, []string{"execute"}, 100, 0, "zone.", 10)
	require.NoError(t, err)
	require.NoError(t, keeper.EnforceZoneExecution(zone, "execute", "zone.contract", 20))
	require.ErrorContains(t, keeper.EnforceZoneExecution(zone, "execute", "service.rpc", 20), "record prefix")
	require.ErrorContains(t, keeper.Authorize(zone, DelegationScopeResolverUpdate, "execute", "zone.contract", 0, 20), "scope mismatch")
}

func TestAuctionKeeperV2DeterministicWinnerAndFeeSplit(t *testing.T) {
	keeper := NewAuctionKeeperV2()
	state := EmptyIdentityState(DefaultIdentityParams())
	state, started, err := keeper.Start(state, "market", 100, 100, "domain.fees")
	require.NoError(t, err)
	require.Equal(t, AuctionRecordV2Commit, started.Status)

	leftCommit, err := ComputeAuctionCommitment("market.aet", addr(1), 100, "left")
	require.NoError(t, err)
	rightCommit, err := ComputeAuctionCommitment("market.aet", addr(2), 150, "right")
	require.NoError(t, err)
	state, _, err = keeper.CommitBid(state, "market.aet", addr(1), leftCommit, 101)
	require.NoError(t, err)
	state, _, err = keeper.CommitBid(state, "market.aet", addr(2), rightCommit, 102)
	require.NoError(t, err)
	state, _, err = keeper.RevealBid(state, "market.aet", addr(1), 100, "left", started.RevealStartHeight)
	require.NoError(t, err)
	state, _, err = keeper.RevealBid(state, "market.aet", addr(2), 150, "right", started.RevealStartHeight+1)
	require.NoError(t, err)
	state, finalized, err := keeper.Finalize(state, "market.aet", started.RevealEndHeight, 100, "domain.fees")
	require.NoError(t, err)
	require.Equal(t, AuctionRecordV2Finalized, finalized.Status)
	require.Equal(t, addr(2), finalized.Winner)
	require.Equal(t, uint64(150), finalized.WinningBid)
	require.Equal(t, "domain.fees", finalized.FeeSplitID)
	require.NoError(t, ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{State: state, Height: started.RevealEndHeight}))

	broken := state.Clone()
	broken.Auctions[0].Winner = addr(9)
	require.ErrorContains(t, ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{State: broken, Height: started.RevealEndHeight}), "winner proof")
}

func TestProofAndRoutingIntegrationKeepersV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	interfaceKey, err := ResolverMetadataInterfaceKey("aw5")
	require.NoError(t, err)
	serviceKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	metadata, err := EncodeResolverMetadata([]ResolverMetadataEntry{
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
	}, 12)
	require.NoError(t, err)

	proofKeeper := NewProofKeeperV2()
	proof, err := proofKeeper.BuildResolutionProof(state, "alice.aet", 13)
	require.NoError(t, err)
	root, err := IdentityStateRoot(state)
	require.NoError(t, err)
	resolution, err := proofKeeper.VerifyResolutionProof(proof, root, 13)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", resolution.ResolverDomain)

	routing := NewRoutingIntegrationKeeperV2()
	send, err := routing.ResolveTransactionTarget(state, "alice.aet", ResolverKeyWallet, 13)
	require.NoError(t, err)
	require.Equal(t, addr(4), send.Address)
	invoke, err := routing.ResolveContractInvocation(state, "alice.aet", "aw5", "swap", identityHash("payload"), 13)
	require.NoError(t, err)
	require.Equal(t, addr(3), invoke.Contract)
	service, err := routing.ResolveServiceMetadata(state, "alice.aet", 13, 30, "rpc")
	require.NoError(t, err)
	require.Equal(t, "https://rpc.aet", service.Endpoint)
}

func TestKeeperInvariantRejectsStaleResolutionCache(t *testing.T) {
	record, err := NewResolutionCacheRecordV2("alice.aet", identityHash("path"), identityHash("record"), 100, 7, 2, 3)
	require.NoError(t, err)
	state := EmptyIdentityState(DefaultIdentityParams())
	require.NoError(t, ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{
		State:		state,
		Height:		10,
		CacheRecords:	[]ResolutionCacheRecordV2{record},
		CacheContexts: []ResolutionCacheUseContextV2{{
			Height:		10,
			SourceVersion:	7,
			ParentEpoch:	2,
			ChildEpoch:	3,
		}},
	}))
	require.ErrorContains(t, ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{
		State:		state,
		Height:		10,
		CacheRecords:	[]ResolutionCacheRecordV2{record},
		CacheContexts: []ResolutionCacheUseContextV2{{
			Height:		10,
			SourceVersion:	8,
			ParentEpoch:	2,
			ChildEpoch:	3,
		}},
	}), "cached resolution invalid")
}
