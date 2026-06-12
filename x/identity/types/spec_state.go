package types

import (
	"bytes"
	"errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultIdentityParams() IdentityParams {
	return IdentityParams{
		RegistrationPeriodBlocks:	DefaultRegistrationPeriodBlocks,
		RenewalWindowBlocks:		DefaultRenewalWindowBlocks,
		CommitTTLBlocks:		DefaultCommitTTLBlocks,
		AuctionCommitBlocks:		DefaultAuctionCommitBlocks,
		AuctionRevealBlocks:		DefaultAuctionRevealBlocks,
	}
}

func EmptyIdentityState(params IdentityParams) IdentityState {
	return IdentityState{Params: normalizeIdentityParams(params)}
}

func CommitDomainRegistration(state IdentityState, name string, owner sdk.AccAddress, commitmentHash string, height uint64) (IdentityState, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, err
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityState{}, err
	}
	if err := validateSpecAddress("identity commit owner", owner); err != nil {
		return IdentityState{}, err
	}
	if err := validateHexHash("identity registration commitment", commitmentHash); err != nil {
		return IdentityState{}, err
	}
	if height == 0 {
		return IdentityState{}, errors.New("identity commit height must be positive")
	}
	if !IsDomainAvailable(state, normalized, height) {
		return IdentityState{}, errors.New("identity domain is not available")
	}
	if _, found := findActiveCommit(state, normalized, height); found {
		return IdentityState{}, errors.New("identity domain already has active commit")
	}
	if isRegistrationCommitmentUsed(state, commitmentHash) {
		return IdentityState{}, errors.New("identity registration commitment already used")
	}
	next := state.Clone()
	next.Commits = append(next.Commits, DomainCommit{
		Name:		normalized,
		Owner:		cloneSpecAddress(owner),
		CommitmentHash:	commitmentHash,
		CommitHeight:	height,
		ExpiresHeight:	height + state.Params.CommitTTLBlocks,
	})
	sortDomainCommits(next.Commits)
	return next, next.Validate()
}

func RevealRegisterDomain(state IdentityState, name string, owner sdk.AccAddress, salt string, height uint64) (IdentityState, Domain, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Domain{}, err
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	expected, err := ComputeRegistrationCommitment(normalized, owner, salt)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	commit, found := findActiveCommitByHash(state, normalized, owner, expected, height)
	if !found {
		return IdentityState{}, Domain{}, errors.New("identity registration commit not found")
	}
	if height < commit.CommitHeight {
		return IdentityState{}, Domain{}, errors.New("identity reveal height is before commit")
	}
	if !isDomainAvailableIgnoringCommits(state, normalized, height) {
		return IdentityState{}, Domain{}, errors.New("identity domain is not available")
	}
	nftID, err := DomainNFTID(normalized)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	domain := Domain{
		Name:			normalized,
		Owner:			cloneSpecAddress(owner),
		NFTID:			nftID,
		RegisteredHeight:	height,
		ExpiryHeight:		height + state.Params.RegistrationPeriodBlocks,
		UpdatedHeight:		height,
	}
	nft := DomainNFT{
		ID:		nftID,
		Domain:		normalized,
		Owner:		cloneSpecAddress(owner),
		MintHeight:	height,
	}
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	next.DomainNFTs = upsertDomainNFT(next.DomainNFTs, nft)
	next.Commits = removeDomainCommit(next.Commits, commit)
	next.UsedCommitments = append(next.UsedCommitments, NewUsedDomainCommitment(commit, height))
	sortIdentityState(&next)
	if err := next.Validate(); err != nil {
		return IdentityState{}, Domain{}, err
	}
	return next, domain, nil
}

func DomainLifecycle(state IdentityState, name string, height uint64) (DomainLifecycleStatus, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	if _, found := findActiveCommit(state, normalized, height); found {
		return DomainLifecycleCommitted, nil
	}
	domain, found := findDomain(state, normalized)
	if !found {
		return DomainLifecycleAvailable, nil
	}
	if domain.ExpiryHeight <= height {
		return DomainLifecycleExpired, nil
	}
	if domain.ExpiryHeight-height <= normalizeIdentityParams(state.Params).RenewalWindowBlocks {
		return DomainLifecycleRenewalWindow, nil
	}
	return DomainLifecycleActive, nil
}

func IsDomainAvailable(state IdentityState, name string, height uint64) bool {
	status, err := DomainLifecycle(state, name, height)
	if err != nil {
		return false
	}
	return status == DomainLifecycleAvailable || status == DomainLifecycleExpired
}

func isDomainAvailableIgnoringCommits(state IdentityState, name string, height uint64) bool {
	domain, found := findDomain(state, name)
	if !found {
		return true
	}
	return domain.ExpiryHeight <= height
}

func RenewIdentityDomain(state IdentityState, name string, actor sdk.AccAddress, height uint64) (IdentityState, Domain, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Domain{}, err
	}
	domain, err := requireActiveDomain(state, name, height)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	if !bytes.Equal(actor, domain.Owner) {
		return IdentityState{}, Domain{}, errors.New("identity renewal requires owner")
	}
	base := domain.ExpiryHeight
	if height > base {
		base = height
	}
	domain.ExpiryHeight = base + state.Params.RegistrationPeriodBlocks
	domain.UpdatedHeight = height
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	sortIdentityState(&next)
	return next, domain, next.Validate()
}

func SetIdentityResolver(state IdentityState, domainName string, actor sdk.AccAddress, update ResolverUpdate, height uint64) (IdentityState, ResolverRecord, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	domain, err := requireActiveDomain(state, domainName, height)
	if err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	if err := canControlResolver(state, domain, actor); err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	update.Domain = domain.Name
	update.UpdatedAtUnix = int64(height)
	record := ResolverRecord{
		Domain:		domain.Name,
		Owner:		cloneSpecAddress(domain.Owner),
		Primary:	cloneSpecAddress(update.Primary),
		Contract:	cloneSpecAddress(update.Contract),
		ZoneEndpoint:	strings.TrimSpace(update.ZoneEndpoint),
		Records:	cloneResolverRecords(update.Records),
		Metadata:	append([]byte(nil), update.Metadata...),
		UpdatedAtUnix:	int64(height),
	}
	if err := ValidateResolverRecord(record); err != nil {
		return IdentityState{}, ResolverRecord{}, err
	}
	next := state.Clone()
	next.Resolvers = upsertResolver(next.Resolvers, record)
	sortIdentityState(&next)
	return next, record, next.Validate()
}

func ResolveIdentityAddress(state IdentityState, name string, height uint64) (sdk.AccAddress, error) {
	domain, err := requireActiveDomain(state, name, height)
	if err != nil {
		return nil, err
	}
	resolver, found := findResolver(state, domain.Name)
	if !found || len(resolver.Primary) == 0 {
		return nil, errors.New("identity domain not resolved")
	}
	return cloneSpecAddress(resolver.Primary), nil
}

func SetIdentityReverse(state IdentityState, addressOwner sdk.AccAddress, address sdk.AccAddress, domainName string, height uint64) (IdentityState, ReverseRecord, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, ReverseRecord{}, err
	}
	if err := validateSpecAddress("identity reverse actor", addressOwner); err != nil {
		return IdentityState{}, ReverseRecord{}, err
	}
	if !bytes.Equal(addressOwner, address) {
		return IdentityState{}, ReverseRecord{}, errors.New("identity reverse update requires address owner")
	}
	domain, err := requireActiveDomain(state, domainName, height)
	if err != nil {
		return IdentityState{}, ReverseRecord{}, err
	}
	resolver, found := findResolver(state, domain.Name)
	if !found || !ResolverRecordContainsAddress(resolver, address) {
		return IdentityState{}, ReverseRecord{}, errors.New("identity resolver does not point to reverse address")
	}
	reverse := ReverseRecord{Address: cloneSpecAddress(address), Domain: domain.Name, UpdatedAtUnix: int64(height)}
	next := state.Clone()
	next.ReverseRecords = upsertReverse(next.ReverseRecords, reverse)
	sortIdentityState(&next)
	return next, reverse, next.Validate()
}

func IssueSubdomain(state IdentityState, parentName string, label string, actor sdk.AccAddress, childOwner sdk.AccAddress, parentControlsResolver bool, height uint64) (IdentityState, SubdomainRecord, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	parent, err := requireActiveDomain(state, parentName, height)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	if !bytes.Equal(actor, parent.Owner) {
		return IdentityState{}, SubdomainRecord{}, errors.New("identity subdomain issuance requires parent owner")
	}
	if err := validateSpecAddress("identity subdomain owner", childOwner); err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	if err := validateDomainLabel(label); err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	childName, err := NormalizeAETDomain(label + "." + parent.Name)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	if !IsDomainAvailable(state, childName, height) {
		return IdentityState{}, SubdomainRecord{}, errors.New("identity subdomain already exists")
	}
	nftID, err := DomainNFTID(childName)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	domain := Domain{Name: childName, Owner: cloneSpecAddress(childOwner), NFTID: nftID, RegisteredHeight: height, ExpiryHeight: parent.ExpiryHeight, UpdatedHeight: height, ParentName: parent.Name, ParentControlsRecord: parentControlsResolver}
	nft := DomainNFT{ID: nftID, Domain: childName, Owner: cloneSpecAddress(childOwner), MintHeight: height}
	record := SubdomainRecord{ParentName: parent.Name, Name: childName, Owner: cloneSpecAddress(childOwner), ParentControlsRecord: parentControlsResolver, CreatedHeight: height}
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	next.DomainNFTs = upsertDomainNFT(next.DomainNFTs, nft)
	next.Subdomains = append(next.Subdomains, record)
	sortIdentityState(&next)
	return next, record, next.Validate()
}

func TransferDomainNFT(state IdentityState, name string, actor sdk.AccAddress, newOwner sdk.AccAddress, height uint64) (IdentityState, Domain, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Domain{}, err
	}
	domain, err := requireActiveDomain(state, name, height)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	if !bytes.Equal(actor, domain.Owner) {
		return IdentityState{}, Domain{}, errors.New("identity NFT transfer requires owner")
	}
	if err := validateSpecAddress("identity NFT new owner", newOwner); err != nil {
		return IdentityState{}, Domain{}, err
	}
	domain.Owner = cloneSpecAddress(newOwner)
	domain.UpdatedHeight = height
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	next.DomainNFTs = transferNFT(next.DomainNFTs, domain.NFTID, newOwner, height)
	next.Resolvers = transferResolverOwnership(next.Resolvers, state.Domains, domain.Name, newOwner, height)
	next.PendingResolverUpdates = removePendingResolverUpdates(next.PendingResolverUpdates, state.Domains, domain.Name)
	next, _, err = InvalidateReverseRecordsForDomainV2(next, domain.Name, height, nil)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	sortIdentityState(&next)
	return next, domain, next.Validate()
}

func ImportIdentityState(state IdentityState) (IdentityState, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, err
	}
	return state.Export(), nil
}

func (s IdentityState) Export() IdentityState {
	out := s.Clone()
	sortIdentityState(&out)
	return out
}

func (s IdentityState) Clone() IdentityState {
	out := IdentityState{
		Params:			normalizeIdentityParams(s.Params),
		Domains:		cloneDomains(s.Domains),
		DomainNFTs:		cloneDomainNFTs(s.DomainNFTs),
		Commits:		cloneDomainCommits(s.Commits),
		UsedCommitments:	cloneUsedDomainCommitments(s.UsedCommitments),
		Resolvers:		cloneResolvers(s.Resolvers),
		ReverseRecords:		cloneReverseRecords(s.ReverseRecords),
		Subdomains:		cloneSubdomains(s.Subdomains),
		Auctions:		cloneAuctions(s.Auctions),
		PendingResolverUpdates:	cloneResolverIntents(s.PendingResolverUpdates),
	}
	return out
}

func normalizeIdentityParams(params IdentityParams) IdentityParams {
	defaults := DefaultIdentityParams()
	if params.RegistrationPeriodBlocks == 0 {
		params.RegistrationPeriodBlocks = defaults.RegistrationPeriodBlocks
	}
	if params.RenewalWindowBlocks == 0 {
		params.RenewalWindowBlocks = defaults.RenewalWindowBlocks
	}
	if params.CommitTTLBlocks == 0 {
		params.CommitTTLBlocks = defaults.CommitTTLBlocks
	}
	if params.AuctionCommitBlocks == 0 {
		params.AuctionCommitBlocks = defaults.AuctionCommitBlocks
	}
	if params.AuctionRevealBlocks == 0 {
		params.AuctionRevealBlocks = defaults.AuctionRevealBlocks
	}
	return params
}

func normalizeIdentityStateParams(state IdentityState) IdentityState {
	state.Params = normalizeIdentityParams(state.Params)
	return state
}

func requireActiveDomain(state IdentityState, name string, height uint64) (Domain, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return Domain{}, err
	}
	domain, found := findDomain(state, normalized)
	if !found {
		return Domain{}, errors.New("identity domain not found")
	}
	if domain.ExpiryHeight <= height {
		return Domain{}, errors.New("identity domain is expired")
	}
	return domain, nil
}

func canControlResolver(state IdentityState, domain Domain, actor sdk.AccAddress) error {
	if err := validateSpecAddress("identity resolver actor", actor); err != nil {
		return err
	}
	if domain.ParentControlsRecord {
		parent, found := findDomain(state, domain.ParentName)
		if !found {
			return errors.New("identity parent domain not found")
		}
		if !bytes.Equal(actor, parent.Owner) {
			return errors.New("identity resolver update requires parent owner")
		}
		return nil
	}
	if !bytes.Equal(actor, domain.Owner) {
		return errors.New("identity resolver update requires owner")
	}
	return nil
}

func sortIdentityState(state *IdentityState) {
	sortDomains(state.Domains)
	sortDomainNFTs(state.DomainNFTs)
	sortDomainCommits(state.Commits)
	sortUsedDomainCommitments(state.UsedCommitments)
	sortResolvers(state.Resolvers)
	sortReverseRecords(state.ReverseRecords)
	sortSubdomains(state.Subdomains)
	sortAuctions(state.Auctions)
	sortResolverIntents(state.PendingResolverUpdates)
}
