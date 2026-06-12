package types

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityStoreV2SpecPrimaryKeyLayout(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	auctionID := identityHash("auction-id")
	pathHash := identityHash("resolution-path")
	commitmentHash := identityHash("commitment")

	domainKey, err := IdentityStoreV2SpecDomainKey("alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecDomainsPrefix+"/"+nameHash, domainKey)
	nameKey, err := IdentityStoreV2SpecDomainNameKey("ALICE.AET")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecDomainNamesPrefix+"/alice.aet", nameKey)
	commitKey, err := IdentityStoreV2SpecCommitmentKey(commitmentHash)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecCommitmentsPrefix+"/"+commitmentHash, commitKey)
	nftKey, err := IdentityStoreV2SpecNFTBindingKey("domain", "alice")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecNFTBindingsPrefix+"/domain/alice", nftKey)
	nftByNameKey, err := IdentityStoreV2SpecNFTBindingByNameKey("alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecNFTBindingsByNamePrefix+"/"+nameHash, nftByNameKey)
	resolverKey, err := IdentityStoreV2SpecResolverKey("alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecResolversPrefix+"/"+nameHash, resolverKey)
	reverseKey, err := IdentityStoreV2SpecReverseKey(addr(1))
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecReversePrefix+"/"+hex.EncodeToString(addr(1)), reverseKey)
	delegationKey, err := IdentityStoreV2SpecDelegationKey("alice.aet", addr(2), DelegationScopeResolverUpdate)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecDelegationsPrefix+"/"+nameHash+"/"+hex.EncodeToString(addr(2))+"/"+string(DelegationScopeResolverUpdate), delegationKey)
	subdomainKey, err := IdentityStoreV2SpecSubdomainKey("alice.aet", "api")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecSubdomainsPrefix+"/"+nameHash+"/"+identityHash("identity-v2-child-label", "api"), subdomainKey)
	auctionKey, err := IdentityStoreV2SpecAuctionKey(auctionID)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecAuctionsPrefix+"/"+auctionID, auctionKey)
	auctionByNameKey, err := IdentityStoreV2SpecAuctionByNameKey("alice.aet", auctionID)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecAuctionsByNamePrefix+"/"+nameHash+"/"+auctionID, auctionByNameKey)
	cacheKey, err := IdentityStoreV2SpecResolutionCacheKey("alice.aet", pathHash)
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecResolutionCachePrefix+"/"+nameHash+"/"+pathHash, cacheKey)
	expiryKey, err := IdentityStoreV2SpecExpiryIndexKey(123, "alice.aet")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s/%020d/%s", IdentityStoreV2SpecExpiryIndexPrefix, uint64(123), nameHash), expiryKey)
	ownerKey, err := IdentityStoreV2SpecOwnerIndexKey(addr(3), "alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecOwnerIndexPrefix+"/"+hex.EncodeToString(addr(3))+"/"+nameHash, ownerKey)
	resolverIndexKey, err := IdentityStoreV2SpecResolverIndexKey(addr(4), "alice.aet")
	require.NoError(t, err)
	require.Equal(t, IdentityStoreV2SpecResolverIndexPrefix+"/"+hex.EncodeToString(addr(4))+"/"+nameHash, resolverIndexKey)
	interfaceMetadataKey, err := IdentityStoreV2SpecInterfaceMetadataKey(identityHash("interface-schema"))
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(interfaceMetadataKey, IdentityStoreV2SpecInterfaceMetadataPrefix+"/"))
}

func TestIdentityStoreV2SpecPerformanceAccessPatterns(t *testing.T) {
	direct, err := IdentityStoreV2SpecDirectResolverReadAccessSet("alice.aet")
	require.NoError(t, err)
	require.Len(t, direct.Reads, 1)
	require.True(t, strings.HasPrefix(direct.Reads[0], IdentityStoreV2SpecResolversPrefix+"/"))

	compactDirect, err := IdentityStoreV2SpecDirectResolutionReadAccessSet("alice.aet", false)
	require.NoError(t, err)
	require.Len(t, compactDirect.Reads, 2)
	require.Contains(t, compactDirect.Reads, mustStoreV2SpecDomainKey(t, "alice.aet"))
	require.Contains(t, compactDirect.Reads, mustStoreV2SpecResolverKey(t, "alice.aet"))
	compactProof, err := IdentityStoreV2SpecDirectResolutionReadAccessSet("alice.aet", true)
	require.NoError(t, err)
	require.Len(t, compactProof.Reads, 3)
	require.Contains(t, compactProof.Reads, mustStoreV2SpecNFTBindingByNameKey(t, "alice.aet"))

	reverse, err := IdentityStoreV2SpecReverseResolutionReadAccessSet(addr(1), "alice.aet")
	require.NoError(t, err)
	require.Len(t, reverse.Reads, 3)
	require.Contains(t, reverse.Reads, IdentityStoreV2SpecReversePrefix+"/"+hex.EncodeToString(addr(1)))

	pathKeys, err := IdentityStoreV2SpecRecursiveResolutionPathKeys("api.dex.alice.aet")
	require.NoError(t, err)
	require.Len(t, pathKeys, 6)
	require.True(t, strings.HasPrefix(pathKeys[0], IdentityStoreV2SpecDomainsPrefix+"/"))
	require.True(t, strings.HasPrefix(pathKeys[1], IdentityStoreV2SpecResolversPrefix+"/"))
	require.True(t, strings.HasPrefix(pathKeys[4], IdentityStoreV2SpecDomainsPrefix+"/"))
	require.True(t, strings.HasPrefix(pathKeys[5], IdentityStoreV2SpecResolversPrefix+"/"))

	compactRecursive, err := IdentityStoreV2SpecRecursiveResolutionReadAccessSet("api.dex.alice.aet", false)
	require.NoError(t, err)
	require.Len(t, compactRecursive.Reads, 4)
	require.Contains(t, compactRecursive.Reads, mustStoreV2SpecDomainKey(t, "alice.aet"))
	require.Contains(t, compactRecursive.Reads, mustStoreV2SpecDomainKey(t, "dex.alice.aet"))
	require.Contains(t, compactRecursive.Reads, mustStoreV2SpecDomainKey(t, "api.dex.alice.aet"))
	require.Contains(t, compactRecursive.Reads, mustStoreV2SpecResolverKey(t, "api.dex.alice.aet"))

	delegatedRecursive, err := IdentityStoreV2SpecRecursiveResolutionReadAccessSet("api.dex.alice.aet", true)
	require.NoError(t, err)
	require.Len(t, delegatedRecursive.Reads, 6)
	delegationPrefix, err := IdentityStoreV2SpecDelegationsByNamePrefix("alice.aet")
	require.NoError(t, err)
	require.Contains(t, delegatedRecursive.Reads, delegationPrefix)

	proofReads, err := IdentityStoreV2SpecResolutionProofReadAccessSet("api.dex.alice.aet", true)
	require.NoError(t, err)
	require.Equal(t, delegatedRecursive.Reads, proofReads.Reads)

	expiryPrefix, err := IdentityStoreV2SpecBoundedExpiryScanPrefix(123)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s/%020d", IdentityStoreV2SpecExpiryIndexPrefix, uint64(123)), expiryPrefix)
}

func TestIdentityStoreV2ExportImportLargeStateRoundTrip(t *testing.T) {
	state := testIdentityStoreV2LargeState(t, 128)
	exported := state.Export()
	imported, err := ImportIdentityState(exported)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
	require.Len(t, imported.Domains, 128)
	require.Len(t, imported.Resolvers, 128)
	require.Len(t, imported.DomainNFTs, 128)

	for _, domain := range imported.Domains {
		domainKey, err := IdentityStoreV2SpecDomainKey(domain.Name)
		require.NoError(t, err)
		resolverKey, err := IdentityStoreV2SpecResolverKey(domain.Name)
		require.NoError(t, err)
		ownerKey, err := IdentityStoreV2SpecOwnerIndexKey(domain.Owner, domain.Name)
		require.NoError(t, err)
		expiryKey, err := IdentityStoreV2SpecExpiryIndexKey(domain.ExpiryHeight, domain.Name)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(domainKey, IdentityStoreV2SpecDomainsPrefix+"/"))
		require.True(t, strings.HasPrefix(resolverKey, IdentityStoreV2SpecResolversPrefix+"/"))
		require.True(t, strings.HasPrefix(ownerKey, IdentityStoreV2SpecOwnerIndexPrefix+"/"))
		require.True(t, strings.HasPrefix(expiryKey, IdentityStoreV2SpecExpiryIndexPrefix+"/"))
	}
}

func mustStoreV2SpecDomainKey(t *testing.T, name string) string {
	t.Helper()
	key, err := IdentityStoreV2SpecDomainKey(name)
	require.NoError(t, err)
	return key
}

func mustStoreV2SpecResolverKey(t *testing.T, name string) string {
	t.Helper()
	key, err := IdentityStoreV2SpecResolverKey(name)
	require.NoError(t, err)
	return key
}

func mustStoreV2SpecNFTBindingByNameKey(t *testing.T, name string) string {
	t.Helper()
	key, err := IdentityStoreV2SpecNFTBindingByNameKey(name)
	require.NoError(t, err)
	return key
}

func testIdentityStoreV2LargeState(t *testing.T, count int) IdentityState {
	t.Helper()
	state := EmptyIdentityState(DefaultIdentityParams())
	for i := 0; i < count; i++ {
		name := benchmarkIdentityName(i)
		owner := benchmarkIdentityAddress(i + 1)
		nftID, err := DomainNFTID(name)
		require.NoError(t, err)
		state.Domains = append(state.Domains, Domain{
			Name:			name,
			Owner:			owner,
			NFTID:			nftID,
			RegisteredHeight:	benchmarkIdentityRevealHeight,
			ExpiryHeight:		benchmarkIdentityRevealHeight + state.Params.RegistrationPeriodBlocks,
			UpdatedHeight:		benchmarkIdentityResolveHeight,
		})
		state.DomainNFTs = append(state.DomainNFTs, DomainNFT{
			ID:		nftID,
			Domain:		name,
			Owner:		owner,
			MintHeight:	benchmarkIdentityRevealHeight,
		})
		state.Resolvers = append(state.Resolvers, ResolverRecord{
			Domain:		name,
			Owner:		owner,
			Primary:	benchmarkIdentityAddress(i + 10_000),
			UpdatedAtUnix:	int64(benchmarkIdentityResolveHeight),
		})
	}
	state = state.Export()
	require.NoError(t, state.Validate())
	return state
}
