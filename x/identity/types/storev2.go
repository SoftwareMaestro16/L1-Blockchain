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
	IdentityStoreV2Prefix = "zone/identity/v2"

	IdentityStoreV2DomainPrefix          = IdentityStoreV2Prefix + "/domain"
	IdentityStoreV2ResolverPrefix        = IdentityStoreV2Prefix + "/resolver"
	IdentityStoreV2NFTPrefix             = IdentityStoreV2Prefix + "/nft"
	IdentityStoreV2OwnerIndexPrefix      = IdentityStoreV2Prefix + "/owner"
	IdentityStoreV2ReversePrefix         = IdentityStoreV2Prefix + "/reverse"
	IdentityStoreV2CommitPrefix          = IdentityStoreV2Prefix + "/commit"
	IdentityStoreV2SubdomainIndexPrefix  = IdentityStoreV2Prefix + "/subdomain"
	IdentityStoreV2AuctionPrefix         = IdentityStoreV2Prefix + "/auction"
	IdentityStoreV2PendingResolverPrefix = IdentityStoreV2Prefix + "/pending-resolver"
)

type IdentityAccessSet struct {
	Reads  []string
	Writes []string
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
		Reads:  sortedUniqueStrings(reads),
		Writes: sortedUniqueStrings(writes),
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
