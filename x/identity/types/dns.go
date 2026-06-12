package types

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ResolverPatch struct {
	Domain			string
	Primary			sdk.AccAddress
	ClearPrimary		bool
	Contract		sdk.AccAddress
	ClearContract		bool
	ZoneEndpoint		string
	ClearZoneEndpoint	bool
	Records			map[string]sdk.AccAddress
	DeleteRecords		[]string
	Metadata		[]byte
	ClearMetadata		bool
	UpdatedAtUnix		int64
}

type IdentityResolution struct {
	QueryDomain	string
	ResolverDomain	string
	AuthorityDomain	Domain
	Record		ResolverRecord
	Depth		uint8
}

type IdentityRouteTarget struct {
	QueryDomain	string
	ResolverDomain	string
	AuthorityName	string
	Key		string
	Address		sdk.AccAddress
	Primary		sdk.AccAddress
	Contract	sdk.AccAddress
	ZoneEndpoint	string
}

func PatchIdentityResolver(state IdentityState, domainName string, actor sdk.AccAddress, patch ResolverPatch, height uint64) (IdentityState, ResolverRecord, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	normalized, err := NormalizeAETDomain(domainName)
	if err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	if patch.Domain != "" {
		if patchDomain, err := NormalizeAETDomain(patch.Domain); err != nil {
			return IdentityState{}, ResolverRecord{}, err
		} else if patchDomain != normalized {
			return IdentityState{}, ResolverRecord{}, errors.New("identity resolver patch domain mismatch")
		}
	}
	authority, err := requireResolverAuthorityDomain(state, normalized, height)
	if err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	if err := canControlResolver(state, authority, actor); err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	existing, found := findResolverByNormalizedDomain(state, normalized)
	if found && !bytes.Equal(existing.Owner, authority.Owner) {
		return IdentityState{}, ResolverRecord{}, errors.New("identity resolver owner must match registry owner")
	}
	patch.Domain = normalized
	patch.UpdatedAtUnix = int64(height)
	record, err := ApplyResolverPatch(existingOrNil(existing, found), authorityDomainRecord(authority), patch, int64(height))
	if err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	next := state.Clone()
	next.Resolvers = upsertResolver(next.Resolvers, record)
	next, _, err = InvalidateReverseRecordsForDomainV2(next, normalized, height, nil)
	if err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	sortIdentityState(&next)
	return next, record, next.Validate()
}

func ApplyResolverPatch(existing *ResolverRecord, domainRecord DomainRecord, patch ResolverPatch, nowUnix int64) (ResolverRecord, error) {
	normalizedDomain, err := NormalizeResolverDomain(patch.Domain)
	if err != nil {
		return ResolverRecord{}, err
	}
	if err := ValidateDomainUsableForResolver(domainRecord, normalizedDomain, nowUnix); err != nil {
		return ResolverRecord{}, err
	}
	touchedKeys, err := ResolverPatchKeys(patch)
	if err != nil {
		return ResolverRecord{}, err
	}
	if len(touchedKeys) == 0 {
		return ResolverRecord{}, errors.New("resolver patch must update at least one key")
	}
	record := ResolverRecord{
		Domain:		normalizedDomain,
		Owner:		cloneAddress(domainRecord.Owner),
		Records:	map[string]sdk.AccAddress{},
		UpdatedAtUnix:	patch.UpdatedAtUnix,
	}
	if existing != nil {
		if err := ValidateResolverRecordForDomain(*existing, domainRecord, nowUnix); err != nil {
			return ResolverRecord{}, err
		}
		if existing.Domain != normalizedDomain {
			return ResolverRecord{}, errors.New("resolver patch domain cannot change existing record domain")
		}
		record = cloneResolver(*existing)
		record.Owner = cloneAddress(domainRecord.Owner)
		record.UpdatedAtUnix = patch.UpdatedAtUnix
		if record.Records == nil {
			record.Records = map[string]sdk.AccAddress{}
		}
	}
	if patch.ClearPrimary {
		record.Primary = nil
	}
	if len(patch.Primary) > 0 {
		record.Primary = cloneAddress(patch.Primary)
	}
	if patch.ClearContract {
		record.Contract = nil
	}
	if len(patch.Contract) > 0 {
		record.Contract = cloneAddress(patch.Contract)
	}
	if patch.ClearZoneEndpoint {
		record.ZoneEndpoint = ""
	}
	if patch.ZoneEndpoint != "" {
		record.ZoneEndpoint = strings.TrimSpace(patch.ZoneEndpoint)
	}
	if patch.ClearMetadata {
		record.Metadata = nil
	}
	if patch.Metadata != nil {
		record.Metadata = append([]byte(nil), patch.Metadata...)
	}
	for _, key := range sortedUniqueResolverKeys(patch.DeleteRecords) {
		delete(record.Records, key)
	}
	for _, key := range sortedResolverKeys(patch.Records) {
		record.Records[key] = cloneAddress(patch.Records[key])
	}
	if len(record.Records) == 0 {
		record.Records = nil
	}
	if err := ValidateResolverRecordForDomain(record, domainRecord, nowUnix); err != nil {
		return ResolverRecord{}, err
	}
	return record, nil
}

func ResolverPatchKeys(patch ResolverPatch) ([]string, error) {
	keys := make([]string, 0, len(patch.Records)+len(patch.DeleteRecords)+4)
	seen := map[string]struct{}{}
	add := func(key string) error {
		if err := ValidateResolverGrantKey(key); err != nil {
			return err
		}
		if _, found := seen[key]; !found {
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
		return nil
	}
	if patch.ClearPrimary || len(patch.Primary) > 0 {
		if err := add(ResolverKeyPrimary); err != nil {
			return nil, err
		}
	}
	if patch.ClearContract || len(patch.Contract) > 0 {
		if err := add(ResolverKeyContract); err != nil {
			return nil, err
		}
	}
	if patch.ClearZoneEndpoint || patch.ZoneEndpoint != "" {
		if err := add(ResolverKeyZoneEndpoint); err != nil {
			return nil, err
		}
	}
	if patch.ClearMetadata || patch.Metadata != nil {
		if err := add(ResolverKeyMetadata); err != nil {
			return nil, err
		}
	}
	for _, key := range sortedResolverKeys(patch.Records) {
		if containsString(patch.DeleteRecords, key) {
			return nil, fmt.Errorf("resolver patch cannot set and delete key %q", key)
		}
		if err := add(key); err != nil {
			return nil, err
		}
	}
	for _, key := range sortedUniqueResolverKeys(patch.DeleteRecords) {
		if err := add(key); err != nil {
			return nil, err
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func ResolveIdentityAddressRecursive(state IdentityState, name string, height uint64) (sdk.AccAddress, error) {
	resolution, err := ResolveIdentityRecordRecursive(state, name, height)
	if err != nil {
		return nil, err
	}
	if len(resolution.Record.Primary) == 0 {
		return nil, errors.New("identity domain not resolved")
	}
	return cloneSpecAddress(resolution.Record.Primary), nil
}

func ResolveIdentityRoute(state IdentityState, name string, key string, height uint64) (IdentityRouteTarget, error) {
	if key == "" {
		key = ResolverKeyPrimary
	}
	if err := ValidateResolverGrantKey(key); err != nil {
		return IdentityRouteTarget{}, err
	}
	resolution, err := ResolveIdentityRecordRecursive(state, name, height)
	if err != nil {
		return IdentityRouteTarget{}, err
	}
	record := resolution.Record
	target := IdentityRouteTarget{
		QueryDomain:	resolution.QueryDomain,
		ResolverDomain:	resolution.ResolverDomain,
		AuthorityName:	resolution.AuthorityDomain.Name,
		Key:		key,
		Primary:	cloneSpecAddress(record.Primary),
		Contract:	cloneSpecAddress(record.Contract),
		ZoneEndpoint:	record.ZoneEndpoint,
	}
	switch key {
	case ResolverKeyPrimary:
		target.Address = cloneSpecAddress(record.Primary)
	case ResolverKeyContract:
		target.Address = cloneSpecAddress(record.Contract)
	default:
		target.Address = cloneSpecAddress(record.Records[key])
	}
	if len(target.Address) == 0 {
		return IdentityRouteTarget{}, fmt.Errorf("identity resolver key %q is not resolved", key)
	}
	return target, nil
}

func ResolveIdentityRecordRecursive(state IdentityState, name string, height uint64) (IdentityResolution, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityResolution{}, err
	}
	query, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityResolution{}, err
	}
	candidates, err := resolverDomainCandidates(query)
	if err != nil {
		return IdentityResolution{}, err
	}
	var nearestAuthority *Domain
	for depth, candidate := range candidates {
		if domain, found := findDomainByNormalizedName(state, candidate); found {
			if domain.ExpiryHeight <= height {
				return IdentityResolution{}, errors.New("identity domain is expired")
			}
			nearestAuthority = &domain
		}
		if resolver, found := findResolverByNormalizedDomain(state, candidate); found {
			authority, err := requireResolverAuthorityDomain(state, resolver.Domain, height)
			if err != nil {
				return IdentityResolution{}, err
			}
			if !bytes.Equal(resolver.Owner, authority.Owner) {
				return IdentityResolution{}, errors.New("identity resolver owner must match registry owner")
			}
			return IdentityResolution{
				QueryDomain:		query,
				ResolverDomain:		resolver.Domain,
				AuthorityDomain:	authority,
				Record:			resolver,
				Depth:			uint8(depth),
			}, nil
		}
		if nearestAuthority != nil && nearestAuthority.Name == candidate {
			return IdentityResolution{}, errors.New("identity domain not resolved")
		}
	}
	return IdentityResolution{}, errors.New("identity domain not found")
}

func requireResolverAuthorityDomain(state IdentityState, resolverDomain string, height uint64) (Domain, error) {
	authority, found := findResolverAuthorityDomain(state, resolverDomain)
	if !found {
		return Domain{}, errors.New("identity resolver authority domain not found")
	}
	if authority.ExpiryHeight <= height {
		return Domain{}, errors.New("identity domain is expired")
	}
	return authority, nil
}

func findResolverAuthorityDomain(state IdentityState, resolverDomain string) (Domain, bool) {
	candidates, err := resolverDomainCandidates(resolverDomain)
	if err != nil {
		return Domain{}, false
	}
	for _, candidate := range candidates {
		if domain, found := findDomainByNormalizedName(state, candidate); found {
			return domain, true
		}
	}
	return Domain{}, false
}

func resolverDomainCandidates(name string) ([]string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return nil, err
	}
	labelsPart := strings.TrimSuffix(normalized, DomainTLD)
	labels := strings.Split(labelsPart, ".")
	candidates := make([]string, 0, len(labels))
	for i := 0; i < len(labels); i++ {
		candidates = append(candidates, strings.Join(labels[i:], ".")+DomainTLD)
	}
	return candidates, nil
}

func authorityDomainRecord(domain Domain) DomainRecord {
	baseName, err := BaseDomainFromResolverDomain(domain.Name)
	if err != nil {
		baseName = strings.TrimSuffix(domain.Name, DomainTLD)
	}
	return DomainRecord{
		Name:		baseName,
		TLD:		DomainTLD,
		Owner:		cloneSpecAddress(domain.Owner),
		ExpiryUnix:	int64(domain.ExpiryHeight),
		NFTItemID:	domain.NFTID,
		Status:		DomainStatusActive,
		CreatedAtUnix:	int64(domain.RegisteredHeight),
		UpdatedAtUnix:	int64(domain.UpdatedHeight),
	}
}

func findDomainByNormalizedName(state IdentityState, normalized string) (Domain, bool) {
	for _, domain := range state.Domains {
		if domain.Name == normalized {
			return cloneDomain(domain), true
		}
	}
	return Domain{}, false
}

func findResolverByNormalizedDomain(state IdentityState, normalized string) (ResolverRecord, bool) {
	for _, resolver := range state.Resolvers {
		if resolver.Domain == normalized {
			return cloneResolver(resolver), true
		}
	}
	return ResolverRecord{}, false
}

func existingOrNil(record ResolverRecord, found bool) *ResolverRecord {
	if !found {
		return nil
	}
	return &record
}

func sortedUniqueResolverKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func domainMatchesOrBelow(child, parent string) bool {
	return child == parent || stringsHasSuffixLabel(child, parent)
}
