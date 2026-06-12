package types

import (
	"strings"
	"testing"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/stretchr/testify/require"
)

func TestIdentityZonePrefixShardDescriptorAndNFTBinding(t *testing.T) {
	descriptor := DefaultIdentityZoneStateMachineDescriptor()
	require.NoError(t, descriptor.Validate())
	require.Equal(t, IdentityStoreV2Prefix, descriptor.StorePrefix)
	require.Contains(t, descriptor.MessageHandlers, IdentityMessageRegisterIdentity)
	require.Contains(t, descriptor.MessageHandlers, IdentityMessageFinalizeIdentityAuction)
	require.Contains(t, descriptor.ProofQueries, IdentityProofIdentityRoot)

	shard, err := IdentityNameShard("API.Alice.AET")
	require.NoError(t, err)
	require.Len(t, shard, 4)

	domainKey, err := IdentityDomainStoreKey("api.alice.aet")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(domainKey, IdentityStoreV2DomainPrefix+"/"+shard+"/"))

	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	require.NoError(t, CheckIdentityNFTBinding(state, "alice.aet", 11))
}

func TestIdentityZoneV2ShardRoutingProofRootsAndLookupMessages(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)

	domainRoute, err := RouteIdentityDomainShard("alice.aet", 8, 3)
	require.NoError(t, err)
	resolverRoute, err := RouteIdentityResolverShard("alice.aet", 8, 3)
	require.NoError(t, err)
	require.Equal(t, domainRoute.ShardID, resolverRoute.ShardID)
	require.Equal(t, IdentityRouteNameHash, domainRoute.RoutingMode)
	require.True(t, strings.HasPrefix(domainRoute.StateKey, IdentityStoreV2SpecDomainsPrefix+"/"))

	reverseRoute, err := RouteIdentityReverseShard(addr(2), 8, 3)
	require.NoError(t, err)
	require.Equal(t, IdentityRouteAddress, reverseRoute.RoutingMode)

	auctionID := identityHash("auction-route")
	auctionRoute, err := RouteIdentityAuctionShard(auctionID, 8, 3)
	require.NoError(t, err)
	require.Equal(t, IdentityRouteAuction, auctionRoute.RoutingMode)

	msg, err := NewMsgResolveIdentity(MsgResolveIdentity{
		RequestID:	"req-1",
		Requester:	"financial/alice",
		SourceZoneID:	"FINANCIAL_ZONE",
		TargetName:	"alice.aet",
		TargetType:	IdentityLookupTargetResolver,
		ProofRequired:	true,
		ReplyTo:	"zone/financial/inbox",
		ExpiryHeight:	100,
	})
	require.NoError(t, err)
	require.NoError(t, msg.Validate())

	reverseResponse, err := BuildProofBackedIdentityReverseLookup(state, addr(2), 13)
	require.NoError(t, err)
	require.NoError(t, reverseResponse.Validate())

	result, err := NewMsgIdentityResolutionResult(MsgIdentityResolutionResult{
		RequestID:		msg.RequestID,
		Name:			msg.TargetName,
		TargetType:		msg.TargetType,
		ResolvedValue:		"addr:2",
		ResolverRecordVersion:	1,
		ProofHashOptional:	reverseResponse.ProofIndex.ProofHash,
		Status:			IdentityResolutionStatusResolved,
		ExpiryHeight:		100,
	})
	require.NoError(t, err)
	require.NoError(t, result.Validate())

	proofIndex, err := QueryIdentityZoneLightClientProof(state, IdentityProofResolver, "alice.aet", 13)
	require.NoError(t, err)
	roots, err := BuildIdentityZoneRoots(13, state, nil, nil, nil, []IdentityZoneProofIndexEntry{proofIndex, reverseResponse.ProofIndex})
	require.NoError(t, err)
	proofRoots, err := BuildIdentityZoneProofRoots(13, roots)
	require.NoError(t, err)
	require.Len(t, proofRoots, 5)
	require.True(t, hasIdentityProofRoot(proofRoots, "identity", roots.StateRoot))
	require.True(t, hasIdentityProofRoot(proofRoots, "resolver", roots.ResolverRoot))
	require.True(t, hasIdentityProofRoot(proofRoots, "reverse", roots.ReverseRoot))

	layout, err := DefaultIdentityStoreV2NameHashLayout()
	require.NoError(t, err)
	require.NoError(t, layout.Validate())
	require.Equal(t, IdentityZoneShardKey, layout.PrimaryShardKey)
}

func hasIdentityProofRoot(roots []coretypes.ProofRoot, rootType string, rootHash string) bool {
	for _, root := range roots {
		if root.ZoneID == coretypes.ZoneIDIdentity && string(root.RootType) == rootType && root.RootHash == rootHash {
			return true
		}
	}
	return false
}

func TestIdentityZoneSpecStateKeys(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)

	domainKey, err := IdentityZoneDomainKeyByHash(nameHash)
	require.NoError(t, err)
	require.Equal(t, "identity/domains/"+nameHash, domainKey)

	resolverKey, err := IdentityZoneResolverKey("alice.aet")
	require.NoError(t, err)
	require.Equal(t, "identity/resolvers/"+nameHash, resolverKey)

	reverseKey, err := IdentityZoneReverseKey(addr(2))
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(reverseKey, "identity/reverse/"))

	nftKey, err := IdentityZoneNFTBindingKeyByHash(nameHash)
	require.NoError(t, err)
	require.Equal(t, "identity/nft_bindings/"+nameHash, nftKey)

	grantKey, err := IdentityZoneGrantKey(nameHash, addr(3), DelegationScopeResolverUpdate)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(grantKey, "identity/grants/"+nameHash+"/"))

	auctionID := identityHash("auction-id")
	auctionKey, err := IdentityZoneAuctionKey(auctionID)
	require.NoError(t, err)
	require.Equal(t, "identity/auctions/"+auctionID, auctionKey)

	proofKey, err := IdentityZoneProofIndexKey(42, nameHash)
	require.NoError(t, err)
	require.Equal(t, "identity/proofs/index/00000000000000000042/"+nameHash, proofKey)
}

func TestIdentityZoneMessagesReceiptsAndProofQueries(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)

	msg, err := NewIdentityLookupMessage(IdentityLookupMessage{
		RequestID:	"req-1",
		QueryDomain:	"alice.aet",
		RecordKey:	ResolverKeyPrimary,
		SourceZone:	"FINANCIAL_ZONE",
		SourceShard:	"0",
		ReplyTo:	"financial/replies",
		Height:		13,
		PayloadHash:	identityHash("lookup-payload"),
	})
	require.NoError(t, err)
	require.NoError(t, msg.Validate())

	response, err := NewIdentityResponseReceipt(IdentityResponseReceipt{
		RequestID:	msg.RequestID,
		QueryDomain:	msg.QueryDomain,
		ResolverDomain:	"alice.aet",
		ResponseHash:	identityHash("lookup-response"),
		Height:		13,
		Success:	true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, response.ReceiptHash)

	auctionReceipt, err := NewIdentityAuctionFinalizationReceipt(IdentityAuctionFinalizationReceipt{
		Domain:		"alice.aet",
		WinnerHash:	identityHash("winner"),
		Height:		20,
		AuctionHash:	identityHash("auction"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, auctionReceipt.ReceiptHash)

	resolverProof, err := BuildIdentityResolverProof(state, "alice.aet")
	require.NoError(t, err)
	require.NoError(t, VerifyIdentityProof(resolverProof))

	reverseProof, err := BuildIdentityReverseLookupProof(state, addr(2))
	require.NoError(t, err)
	require.NoError(t, VerifyIdentityProof(reverseProof))
}

func TestIdentityZoneResolverHooksGraphBindingsAndRoots(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 12)
	require.NoError(t, err)
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	stateRoot, err := IdentityStateRoot(state)
	require.NoError(t, err)

	hook, err := NewIdentityResolverVMHook(IdentityResolverVMHook{
		HookID:		"primary-hook",
		NameHash:	nameHash,
		RecordKey:	ResolverKeyPrimary,
		InputHash:	identityHash("resolver-input"),
		OutputHash:	identityHash("resolver-output"),
		GasLimit:	10_000,
		Version:	1,
	})
	require.NoError(t, err)
	require.Equal(t, ComputeIdentityResolverVMHookHash(hook), hook.HookHash)

	graph, err := NewIdentityResolutionGraph(IdentityResolutionGraph{
		Height:	12,
		Nodes: []IdentityResolutionGraphNode{
			{NodeID: "resolver", NameHash: nameHash, RecordKey: ResolverKeyPrimary, TargetHash: identityHash("target")},
			{NodeID: "domain", NameHash: nameHash, TargetHash: stateRoot},
		},
		Edges: []IdentityResolutionGraphEdge{
			{FromNodeID: "domain", ToNodeID: "resolver", EdgeKind: "resolves"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, ComputeIdentityResolutionGraphHash(graph), graph.GraphHash)

	binding, err := NewIdentityCrossZoneBinding(IdentityCrossZoneBinding{
		NameHash:	nameHash,
		ZoneID:		"APPLICATION_ZONE",
		BindingKey:	"application/contracts/app-1",
		BindingRoot:	identityHash("binding-root"),
		Height:		12,
	})
	require.NoError(t, err)

	proofIndex, err := QueryIdentityZoneLightClientProof(state, IdentityProofResolver, "alice.aet", 12)
	require.NoError(t, err)
	reverseIndex, _, err := QueryIdentityZoneReverseLookupProof(state, addr(2), 12)
	require.NoError(t, err)

	roots, err := BuildIdentityZoneRoots(12, state, []IdentityResolverVMHook{hook}, []IdentityResolutionGraph{graph}, []IdentityCrossZoneBinding{binding}, []IdentityZoneProofIndexEntry{proofIndex, reverseIndex})
	require.NoError(t, err)
	require.NoError(t, roots.Validate())
	require.Len(t, roots.StateRoot, 64)
	require.Equal(t, ComputeIdentityResolverVMHookRoot([]IdentityResolverVMHook{hook}), roots.ResolverHookRoot)
	require.Equal(t, ComputeIdentityResolutionGraphRoot([]IdentityResolutionGraph{graph}), roots.GraphRoot)
	require.Equal(t, ComputeIdentityCrossZoneBindingRoot([]IdentityCrossZoneBinding{binding}), roots.CrossZoneRoot)
}
