package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IdentityStoreV2Prefix	= "zone/identity/v2"

	IdentityStoreV2DomainPrefix		= IdentityStoreV2Prefix + "/domain"
	IdentityStoreV2ResolverPrefix		= IdentityStoreV2Prefix + "/resolver"
	IdentityStoreV2NFTPrefix		= IdentityStoreV2Prefix + "/nft"
	IdentityStoreV2OwnerIndexPrefix		= IdentityStoreV2Prefix + "/owner"
	IdentityStoreV2ReversePrefix		= IdentityStoreV2Prefix + "/reverse"
	IdentityStoreV2CommitPrefix		= IdentityStoreV2Prefix + "/commit"
	IdentityStoreV2SubdomainIndexPrefix	= IdentityStoreV2Prefix + "/subdomain"
	IdentityStoreV2AuctionPrefix		= IdentityStoreV2Prefix + "/auction"
	IdentityStoreV2PendingResolverPrefix	= IdentityStoreV2Prefix + "/pending-resolver"

	IdentityStoreV2SpecDomainsPrefix		= IdentityStoreV2Prefix + "/domains"
	IdentityStoreV2SpecDomainNamesPrefix		= IdentityStoreV2Prefix + "/domain_names"
	IdentityStoreV2SpecCommitmentsPrefix		= IdentityStoreV2Prefix + "/commitments"
	IdentityStoreV2SpecNFTBindingsPrefix		= IdentityStoreV2Prefix + "/nft_bindings"
	IdentityStoreV2SpecNFTBindingsByNamePrefix	= IdentityStoreV2Prefix + "/nft_bindings_by_name"
	IdentityStoreV2SpecResolversPrefix		= IdentityStoreV2Prefix + "/resolvers"
	IdentityStoreV2SpecReversePrefix		= IdentityStoreV2Prefix + "/reverse"
	IdentityStoreV2SpecDelegationsPrefix		= IdentityStoreV2Prefix + "/delegations"
	IdentityStoreV2SpecSubdomainsPrefix		= IdentityStoreV2Prefix + "/subdomains"
	IdentityStoreV2SpecAuctionsPrefix		= IdentityStoreV2Prefix + "/auctions"
	IdentityStoreV2SpecAuctionsByNamePrefix		= IdentityStoreV2Prefix + "/auctions_by_name"
	IdentityStoreV2SpecResolutionCachePrefix	= IdentityStoreV2Prefix + "/resolution_cache"
	IdentityStoreV2SpecExpiryIndexPrefix		= IdentityStoreV2Prefix + "/expiry_index"
	IdentityStoreV2SpecOwnerIndexPrefix		= IdentityStoreV2Prefix + "/owner_index"
	IdentityStoreV2SpecResolverIndexPrefix		= IdentityStoreV2Prefix + "/resolver_index"
	IdentityStoreV2SpecInterfaceMetadataPrefix	= IdentityStoreV2Prefix + "/interface_metadata"
)

type IdentityAccessSet struct {
	Reads	[]string
	Writes	[]string
}

func IdentityDomainStoreKey(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2DomainPrefix, identityStoreShard(normalized), reversedDomainStorePath(normalized)), nil
}

func IdentityResolverStoreKey(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2ResolverPrefix, identityStoreShard(normalized), reversedDomainStorePath(normalized)), nil
}

func IdentityNFTStoreKey(id string) (string, error) {
	if strings.TrimSpace(id) == "" || strings.Contains(id, "/") {
		return "", errors.New("identity nft id is invalid for store key")
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2NFTPrefix, identityHash("store-nft", id)), nil
}

func IdentityOwnerDomainIndexKey(owner sdk.AccAddress, name string) (string, error) {
	if err := validateSpecAddress("identity owner index address", owner); err != nil {
		return "", err
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2OwnerIndexPrefix, hex.EncodeToString(owner), reversedDomainStorePath(normalized)), nil
}

func IdentityReverseStoreKey(address sdk.AccAddress) (string, error) {
	if err := validateSpecAddress("identity reverse address", address); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2ReversePrefix, hex.EncodeToString(address)), nil
}

func IdentityCommitStoreKey(name string, owner sdk.AccAddress, commitmentHash string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity commit owner", owner); err != nil {
		return "", err
	}
	if err := validateHexHash("identity commit hash", commitmentHash); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s", IdentityStoreV2CommitPrefix, identityStoreShard(normalized), reversedDomainStorePath(normalized), hex.EncodeToString(owner), commitmentHash), nil
}

func IdentitySubdomainIndexKey(parentName string, childName string) (string, error) {
	parent, err := NormalizeAETDomain(parentName)
	if err != nil {
		return "", err
	}
	child, err := NormalizeAETDomain(childName)
	if err != nil {
		return "", err
	}
	if child != parent && !stringsHasSuffixLabel(child, parent) {
		return "", errors.New("identity subdomain index child must be below parent")
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2SubdomainIndexPrefix, reversedDomainStorePath(parent), reversedDomainStorePath(child)), nil
}

func IdentityAuctionStoreKey(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2AuctionPrefix, identityStoreShard(normalized), reversedDomainStorePath(normalized)), nil
}

func IdentityPendingResolverStoreKey(domain string, actor sdk.AccAddress, nonce uint64) (string, error) {
	normalized, err := NormalizeAETDomain(domain)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity pending resolver actor", actor); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s/%s/%020d", IdentityStoreV2PendingResolverPrefix, identityStoreShard(normalized), reversedDomainStorePath(normalized), hex.EncodeToString(actor), nonce), nil
}

func IdentityStoreV2SpecDomainKey(name string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return IdentityStoreV2SpecDomainKeyByHash(nameHash)
}

func IdentityStoreV2SpecDomainKeyByHash(nameHash string) (string, error) {
	if err := validateHexHash("identity v2 domain name hash", nameHash); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecDomainsPrefix, nameHash), nil
}

func IdentityStoreV2SpecDomainNameKey(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecDomainNamesPrefix, normalized), nil
}

func IdentityStoreV2SpecCommitmentKey(commitmentHash string) (string, error) {
	if err := validateHexHash("identity v2 commitment hash", commitmentHash); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecCommitmentsPrefix, commitmentHash), nil
}

func IdentityStoreV2SpecNFTBindingKey(nftClassID string, nftItemID string) (string, error) {
	if err := validateStorePathSegmentV2("identity v2 nft class id", nftClassID); err != nil {
		return "", err
	}
	if err := validateStorePathSegmentV2("identity v2 nft item id", nftItemID); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2SpecNFTBindingsPrefix, nftClassID, nftItemID), nil
}

func IdentityStoreV2SpecNFTBindingByNameKey(name string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecNFTBindingsByNamePrefix, nameHash), nil
}

func IdentityStoreV2SpecResolverKey(name string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecResolversPrefix, nameHash), nil
}

func IdentityStoreV2SpecReverseKey(address sdk.AccAddress) (string, error) {
	if err := validateSpecAddress("identity v2 reverse address", address); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecReversePrefix, hex.EncodeToString(address)), nil
}

func IdentityStoreV2SpecDelegationKey(name string, delegate sdk.AccAddress, scope DelegationScopeV2) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 delegation delegate", delegate); err != nil {
		return "", err
	}
	if err := validateDelegationScopeV2(scope); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s/%s", IdentityStoreV2SpecDelegationsPrefix, nameHash, hex.EncodeToString(delegate), scope), nil
}

func IdentityStoreV2SpecDelegationsByNamePrefix(name string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecDelegationsPrefix, nameHash), nil
}

func IdentityStoreV2SpecSubdomainKey(parentName string, childLabel string) (string, error) {
	parentHash, err := DomainRecordV2NameHash(parentName)
	if err != nil {
		return "", err
	}
	if err := validateDomainLabel(childLabel); err != nil {
		return "", err
	}
	childLabelHash := identityHash("identity-v2-child-label", childLabel)
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2SpecSubdomainsPrefix, parentHash, childLabelHash), nil
}

func IdentityStoreV2SpecAuctionKey(auctionID string) (string, error) {
	if err := validateHexHash("identity v2 auction id", auctionID); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecAuctionsPrefix, auctionID), nil
}

func IdentityStoreV2SpecAuctionByNameKey(name string, auctionID string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	if err := validateHexHash("identity v2 auction id", auctionID); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2SpecAuctionsByNamePrefix, nameHash, auctionID), nil
}

func IdentityStoreV2SpecResolutionCacheKey(name string, pathHash string) (string, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	if err := validateHexHash("identity v2 resolution path hash", pathHash); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2SpecResolutionCachePrefix, nameHash, pathHash), nil
}

func IdentityStoreV2SpecExpiryIndexKey(expiryHeight uint64, name string) (string, error) {
	if expiryHeight == 0 {
		return "", errors.New("identity v2 expiry index height is required")
	}
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%020d/%s", IdentityStoreV2SpecExpiryIndexPrefix, expiryHeight, nameHash), nil
}

func IdentityStoreV2SpecOwnerIndexKey(owner sdk.AccAddress, name string) (string, error) {
	if err := validateSpecAddress("identity v2 owner index address", owner); err != nil {
		return "", err
	}
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2SpecOwnerIndexPrefix, hex.EncodeToString(owner), nameHash), nil
}

func IdentityStoreV2SpecResolverIndexKey(resolver sdk.AccAddress, name string) (string, error) {
	if err := validateSpecAddress("identity v2 resolver index address", resolver); err != nil {
		return "", err
	}
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", IdentityStoreV2SpecResolverIndexPrefix, hex.EncodeToString(resolver), nameHash), nil
}

func IdentityStoreV2SpecInterfaceMetadataKey(schemaHash string) (string, error) {
	if err := validateHexHash("identity v2 interface metadata hash", schemaHash); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", IdentityStoreV2SpecInterfaceMetadataPrefix, schemaHash), nil
}

func IdentityStoreV2SpecDirectResolverReadAccessSet(name string) (IdentityAccessSet, error) {
	key, err := IdentityStoreV2SpecResolverKey(name)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	return newIdentityAccessSet([]string{key}, nil), nil
}

func IdentityStoreV2SpecDirectResolutionReadAccessSet(name string, includeNFTBinding bool) (IdentityAccessSet, error) {
	domainKey, err := IdentityStoreV2SpecDomainKey(name)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	resolverKey, err := IdentityStoreV2SpecResolverKey(name)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	reads := []string{domainKey, resolverKey}
	if includeNFTBinding {
		nftKey, err := IdentityStoreV2SpecNFTBindingByNameKey(name)
		if err != nil {
			return IdentityAccessSet{}, err
		}
		reads = append(reads, nftKey)
	}
	return newIdentityAccessSet(reads, nil), nil
}

func IdentityStoreV2SpecReverseResolutionReadAccessSet(address sdk.AccAddress, name string) (IdentityAccessSet, error) {
	reverseKey, err := IdentityStoreV2SpecReverseKey(address)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	domainKey, err := IdentityStoreV2SpecDomainKey(name)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	resolverKey, err := IdentityStoreV2SpecResolverKey(name)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	return newIdentityAccessSet([]string{reverseKey, domainKey, resolverKey}, nil), nil
}

func IdentityStoreV2SpecRecursiveResolutionPathKeys(name string) ([]string, error) {
	candidates, err := resolverDomainCandidates(name)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(candidates)*2)
	for _, candidate := range candidates {
		domainKey, err := IdentityStoreV2SpecDomainKey(candidate)
		if err != nil {
			return nil, err
		}
		resolverKey, err := IdentityStoreV2SpecResolverKey(candidate)
		if err != nil {
			return nil, err
		}
		keys = append(keys, domainKey, resolverKey)
	}
	return keys, nil
}

func IdentityStoreV2SpecRecursiveResolutionReadAccessSet(name string, includeDelegations bool) (IdentityAccessSet, error) {
	path, err := CanonicalResolutionPathV2(name)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	reads := make([]string, 0, len(path.Path)+1)
	for _, candidate := range path.Path {
		key, err := IdentityStoreV2SpecDomainKey(candidate)
		if err != nil {
			return IdentityAccessSet{}, err
		}
		reads = append(reads, key)
	}
	finalResolver, err := IdentityStoreV2SpecResolverKey(path.TargetName)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	reads = append(reads, finalResolver)
	if includeDelegations {
		for _, candidate := range path.Path[:len(path.Path)-1] {
			prefix, err := IdentityStoreV2SpecDelegationsByNamePrefix(candidate)
			if err != nil {
				return IdentityAccessSet{}, err
			}
			reads = append(reads, prefix)
		}
	}
	return newIdentityAccessSet(reads, nil), nil
}

func IdentityStoreV2SpecResolutionProofReadAccessSet(name string, recursive bool) (IdentityAccessSet, error) {
	var base IdentityAccessSet
	var err error
	if recursive {
		base, err = IdentityStoreV2SpecRecursiveResolutionReadAccessSet(name, true)
	} else {
		base, err = IdentityStoreV2SpecDirectResolutionReadAccessSet(name, true)
	}
	if err != nil {
		return IdentityAccessSet{}, err
	}
	return newIdentityAccessSet(base.Reads, nil), nil
}

func IdentityStoreV2SpecBoundedExpiryScanPrefix(expiryHeight uint64) (string, error) {
	if expiryHeight == 0 {
		return "", errors.New("identity v2 expiry scan height is required")
	}
	return fmt.Sprintf("%s/%020d", IdentityStoreV2SpecExpiryIndexPrefix, expiryHeight), nil
}

func IdentityNameShard(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return identityStoreShard(normalized), nil
}

func IdentityResolutionAccessSet(name string) (IdentityAccessSet, error) {
	candidates, err := resolverDomainCandidates(name)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	reads := make([]string, 0, len(candidates)*2)
	for _, candidate := range candidates {
		domainKey, err := IdentityDomainStoreKey(candidate)
		if err != nil {
			return IdentityAccessSet{}, err
		}
		resolverKey, err := IdentityResolverStoreKey(candidate)
		if err != nil {
			return IdentityAccessSet{}, err
		}
		reads = append(reads, domainKey, resolverKey)
	}
	return newIdentityAccessSet(reads, nil), nil
}

func IdentityResolverPatchAccessSet(domain string) (IdentityAccessSet, error) {
	resolutionSet, err := IdentityResolutionAccessSet(domain)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	resolverKey, err := IdentityResolverStoreKey(domain)
	if err != nil {
		return IdentityAccessSet{}, err
	}
	return newIdentityAccessSet(resolutionSet.Reads, []string{resolverKey}), nil
}

func (set IdentityAccessSet) Conflicts(other IdentityAccessSet) bool {
	leftWrites := stringSet(set.Writes)
	rightWrites := stringSet(other.Writes)
	for _, key := range other.Writes {
		if _, found := leftWrites[key]; found {
			return true
		}
	}
	for _, key := range other.Reads {
		if _, found := leftWrites[key]; found {
			return true
		}
	}
	for _, key := range set.Reads {
		if _, found := rightWrites[key]; found {
			return true
		}
	}
	return false
}

func newIdentityAccessSet(reads []string, writes []string) IdentityAccessSet {
	return IdentityAccessSet{
		Reads:	sortedUniqueStrings(reads),
		Writes:	sortedUniqueStrings(writes),
	}
}

func identityStoreShard(normalizedDomain string) string {
	return identityHash("store-v2-shard", normalizedDomain)[:4]
}

func reversedDomainStorePath(normalizedDomain string) string {
	labelsPart := strings.TrimSuffix(normalizedDomain, DomainTLD)
	labels := strings.Split(labelsPart, ".")
	for i, j := 0, len(labels)-1; i < j; i, j = i+1, j-1 {
		labels[i], labels[j] = labels[j], labels[i]
	}
	return strings.Join(append([]string{"aet"}, labels...), "/")
}

func sortedUniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func validateStorePathSegmentV2(field string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", field)
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("%s must not contain path separators", field)
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c < 0x21 || c > 0x7e {
			return fmt.Errorf("%s contains unsupported character %q", field, c)
		}
	}
	return nil
}
