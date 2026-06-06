package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityZonePrefixShardDescriptorAndNFTBinding(t *testing.T) {
	descriptor := DefaultIdentityZoneStateMachineDescriptor()
	require.NoError(t, descriptor.Validate())
	require.Equal(t, "identity", descriptor.StorePrefix)
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
		RequestID:   "req-1",
		QueryDomain: "alice.aet",
		RecordKey:   ResolverKeyPrimary,
		SourceZone:  "FINANCIAL_ZONE",
		SourceShard: "0",
		ReplyTo:     "financial/replies",
		Height:      13,
		PayloadHash: identityHash("lookup-payload"),
	})
	require.NoError(t, err)
	require.NoError(t, msg.Validate())

	response, err := NewIdentityResponseReceipt(IdentityResponseReceipt{
		RequestID:      msg.RequestID,
		QueryDomain:    msg.QueryDomain,
		ResolverDomain: "alice.aet",
		ResponseHash:   identityHash("lookup-response"),
		Height:         13,
		Success:        true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, response.ReceiptHash)

	auctionReceipt, err := NewIdentityAuctionFinalizationReceipt(IdentityAuctionFinalizationReceipt{
		Domain:      "alice.aet",
		WinnerHash:  identityHash("winner"),
		Height:      20,
		AuctionHash: identityHash("auction"),
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
		HookID:     "primary-hook",
		NameHash:   nameHash,
		RecordKey:  ResolverKeyPrimary,
		InputHash:  identityHash("resolver-input"),
		OutputHash: identityHash("resolver-output"),
		GasLimit:   10_000,
		Version:    1,
	})
	require.NoError(t, err)
	require.Equal(t, ComputeIdentityResolverVMHookHash(hook), hook.HookHash)

	graph, err := NewIdentityResolutionGraph(IdentityResolutionGraph{
		Height: 12,
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
		NameHash:    nameHash,
		ZoneID:      "APPLICATION_ZONE",
		BindingKey:  "application/contracts/app-1",
		BindingRoot: identityHash("binding-root"),
		Height:      12,
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
