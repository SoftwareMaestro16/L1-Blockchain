package types

import (
	"fmt"
)

type IdentityConsistencySeverityV2 string

const (
	IdentityConsistencySeverityErrorV2	IdentityConsistencySeverityV2	= "error"
	IdentityConsistencySeverityWarningV2	IdentityConsistencySeverityV2	= "warning"
)

type IdentityConsistencyIssueV2 struct {
	Code		string
	Severity	IdentityConsistencySeverityV2
	Name		string
	NameHash	string
	Message		string
}

type IdentityConsistencyAuditRequestV2 struct {
	State				IdentityState
	Height				uint64
	Delegations			[]DelegationRecordV2
	AllowReservedResolver		bool
	AllowRenewalAuctionForName	bool
}

type IdentityConsistencyAuditResultV2 struct {
	Height	uint64
	Valid	bool
	Issues	[]IdentityConsistencyIssueV2
}

func QueryIdentityPeriodicConsistencyAuditV2(request IdentityConsistencyAuditRequestV2) IdentityConsistencyAuditResultV2 {
	result := IdentityConsistencyAuditResultV2{Height: request.Height, Valid: true}
	add := func(code string, name string, message string) {
		nameHash := ""
		if name != "" {
			if hash, err := DomainRecordV2NameHash(name); err == nil {
				nameHash = hash
			}
		}
		result.Valid = false
		result.Issues = append(result.Issues, IdentityConsistencyIssueV2{
			Code:		code,
			Severity:	IdentityConsistencySeverityErrorV2,
			Name:		name,
			NameHash:	nameHash,
			Message:	message,
		})
	}
	if request.Height == 0 {
		add("audit_height_missing", "", "identity consistency audit height is required")
	}
	state := normalizeIdentityStateParams(request.State)
	if err := state.Validate(); err != nil {
		add("state_validate", "", err.Error())
	}
	namesByHash := map[string]string{}
	for _, domain := range state.Domains {
		hash, err := DomainRecordV2NameHash(domain.Name)
		if err != nil {
			add("invalid_domain_name_hash", domain.Name, err.Error())
			continue
		}
		if existing, found := namesByHash[hash]; found && existing != domain.Name {
			add("duplicate_name_hash", domain.Name, fmt.Sprintf("identity name_hash %s maps to both %s and %s", hash, existing, domain.Name))
		}
		namesByHash[hash] = domain.Name
		if domain.ExpiryHeight > request.Height {
			nft, found := findDomainNFTByID(state, domain.NFTID)
			if !found || !addressesEqual(domain.Owner, nft.Owner) {
				add("active_domain_without_nft_binding", domain.Name, "active domain requires active nft binding with matching owner")
			}
		}
		if domain.ParentName != "" {
			parent, found := findDomain(state, domain.ParentName)
			if !found {
				add("subdomain_parent_missing", domain.Name, "subdomain requires existing parent unless detached authorization exists")
			} else if domain.ExpiryHeight > parent.ExpiryHeight {
				add("subdomain_outlives_parent", domain.Name, "subdomain cannot outlive parent unless detached and paid")
			}
		}
	}
	for _, resolver := range state.Resolvers {
		authority, found := findResolverAuthorityDomain(state, resolver.Domain)
		if !found {
			if !request.AllowReservedResolver || !isReservedIdentityResolverRecordV2(resolver.Domain) {
				add("resolver_without_domain", resolver.Domain, "resolver record requires existing domain")
			}
			continue
		}
		if !addressesEqual(resolver.Owner, authority.Owner) {
			add("resolver_owner_mismatch", resolver.Domain, "resolver owner must match active registry owner")
		}
	}
	for _, reverse := range state.ReverseRecords {
		record, err := reverseRecordV2FromLegacy(state, reverse, true)
		if err != nil {
			add("verified_reverse_invalid", reverse.Domain, err.Error())
			continue
		}
		if err := ValidateReverseResolutionRecordV2(state, record, request.Height, nil); err != nil {
			add("verified_reverse_without_forward_consistency", reverse.Domain, err.Error())
		}
	}
	for _, subdomain := range state.Subdomains {
		if subdomain.Detached && subdomain.ParentAuthorized {
			continue
		}
		if _, found := findDomain(state, subdomain.ParentName); !found {
			add("subdomain_without_parent", subdomain.Name, "subdomain requires parent or detached authorization")
		}
	}
	for _, auction := range state.Auctions {
		if auction.Phase == AuctionPhaseFinalized {
			continue
		}
		if domain, found := findDomain(state, auction.Name); found && domain.ExpiryHeight > request.Height && !request.AllowRenewalAuctionForName {
			add("active_auction_for_active_name", auction.Name, "active auction cannot exist for active name unless renewal auction mode is enabled")
		}
	}
	for _, delegation := range request.Delegations {
		if err := ValidateDelegationRecordV2(delegation); err != nil {
			add("delegation_invalid", "", err.Error())
			continue
		}
		domain, found := findDomainByNameHashV2(state, delegation.NameHash)
		if !found {
			add("delegation_domain_missing", "", "delegation requires existing domain")
			continue
		}
		if delegation.ExpiresAtHeight > domain.ExpiryHeight && !delegationDetachedPaidV2(delegation) {
			add("delegation_outlives_domain", domain.Name, "delegation cannot extend beyond domain expiry unless detached and paid")
		}
	}
	return result
}

func ValidateIdentityStateExportV2(state IdentityState, height uint64, delegations []DelegationRecordV2) error {
	exported, err := ImportIdentityState(state)
	if err != nil {
		return err
	}
	audit := QueryIdentityPeriodicConsistencyAuditV2(IdentityConsistencyAuditRequestV2{State: exported, Height: height, Delegations: delegations})
	if !audit.Valid {
		return fmt.Errorf("identity state export consistency audit failed: %s", audit.Issues[0].Message)
	}
	return nil
}

func ValidateIdentityStateMigrationV2(before IdentityState, after IdentityState, height uint64, delegations []DelegationRecordV2) error {
	if err := ValidateIdentityStateExportV2(after, height, delegations); err != nil {
		return err
	}
	beforeExport := before.Export()
	afterExport := after.Export()
	if len(afterExport.Domains) < len(beforeExport.Domains) {
		return fmt.Errorf("identity state migration removed domains: before=%d after=%d", len(beforeExport.Domains), len(afterExport.Domains))
	}
	return nil
}

func RunIdentityEndBlockConsistencyAuditV2(state IdentityState, height uint64, delegations []DelegationRecordV2) IdentityConsistencyAuditResultV2 {
	return QueryIdentityPeriodicConsistencyAuditV2(IdentityConsistencyAuditRequestV2{State: state, Height: height, Delegations: delegations})
}

func findDomainByNameHashV2(state IdentityState, nameHash string) (Domain, bool) {
	for _, domain := range state.Domains {
		hash, err := DomainRecordV2NameHash(domain.Name)
		if err == nil && hash == nameHash {
			return cloneDomain(domain), true
		}
	}
	return Domain{}, false
}

func isReservedIdentityResolverRecordV2(domain string) bool {
	return domain == "_system.aet" || domain == "_reserved.aet"
}

func delegationDetachedPaidV2(delegation DelegationRecordV2) bool {
	for _, permission := range delegation.Permissions {
		if permission == "detached_paid" {
			return true
		}
	}
	return false
}
