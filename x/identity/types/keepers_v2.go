package types

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IdentityKeeperV2 struct{}

type ResolverKeeperV2 struct{}

type DelegationKeeperV2 struct{}

type AuctionKeeperV2 struct{}

type ProofKeeperV2 struct{}

type RoutingIntegrationKeeperV2 struct{}

type IdentityKeeperInvariantInputV2 struct {
	State		IdentityState
	Height		uint64
	CacheRecords	[]ResolutionCacheRecordV2
	CacheContexts	[]ResolutionCacheUseContextV2
}

func NewIdentityKeeperV2() IdentityKeeperV2	{ return IdentityKeeperV2{} }

func NewResolverKeeperV2() ResolverKeeperV2	{ return ResolverKeeperV2{} }

func NewDelegationKeeperV2() DelegationKeeperV2	{ return DelegationKeeperV2{} }

func NewAuctionKeeperV2() AuctionKeeperV2	{ return AuctionKeeperV2{} }

func NewProofKeeperV2() ProofKeeperV2	{ return ProofKeeperV2{} }

func NewRoutingIntegrationKeeperV2() RoutingIntegrationKeeperV2	{ return RoutingIntegrationKeeperV2{} }

func (IdentityKeeperV2) CommitRegistration(state IdentityState, name string, owner sdk.AccAddress, commitmentHash string, height uint64) (IdentityState, error) {
	return CommitDomainRegistration(state, name, owner, commitmentHash, height)
}

func (IdentityKeeperV2) RevealRegistration(state IdentityState, name string, owner sdk.AccAddress, salt string, height uint64) (IdentityState, Domain, error) {
	next, domain, err := RevealRegisterDomain(state, name, owner, salt, height)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	if _, _, err := ValidateRegistryNFTAuthority(next, domain.Name, height); err != nil {
		return IdentityState{}, Domain{}, err
	}
	return next, domain, nil
}

func (IdentityKeeperV2) Renew(state IdentityState, name string, actor sdk.AccAddress, height uint64) (IdentityState, Domain, error) {
	next, domain, err := RenewIdentityDomain(state, name, actor, height)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	if _, _, err := ValidateRegistryNFTAuthority(next, domain.Name, height); err != nil {
		return IdentityState{}, Domain{}, err
	}
	return next, domain, nil
}

func (IdentityKeeperV2) Transfer(state IdentityState, name string, actor sdk.AccAddress, newOwner sdk.AccAddress, height uint64) (IdentityState, Domain, error) {
	next, domain, err := TransferDomainNFT(state, name, actor, newOwner, height)
	if err != nil {
		return IdentityState{}, Domain{}, err
	}
	if _, _, err := ValidateRegistryNFTAuthority(next, domain.Name, height); err != nil {
		return IdentityState{}, Domain{}, err
	}
	return next, domain, nil
}

func (IdentityKeeperV2) ExpiredDomains(state IdentityState, height uint64, limit int) ([]Domain, error) {
	if limit <= 0 {
		return nil, errors.New("identity keeper expiry limit must be positive")
	}
	if err := state.Validate(); err != nil {
		return nil, err
	}
	out := make([]Domain, 0, limit)
	for _, domain := range state.Domains {
		if domain.ExpiryHeight <= height {
			out = append(out, cloneDomain(domain))
			if len(out) == limit {
				break
			}
		}
	}
	return out, nil
}

func (ResolverKeeperV2) UpdateUnifiedRecord(state IdentityState, name string, actor sdk.AccAddress, patch ResolverPatch, height uint64, ttl uint64) (IdentityState, UnifiedResolutionRecordV2, error) {
	next, _, err := PatchIdentityResolver(state, name, actor, patch, height)
	if err != nil {
		return IdentityState{}, UnifiedResolutionRecordV2{}, err
	}
	record, err := BuildUnifiedResolutionRecordV2(next, name, height, ttl)
	if err != nil {
		return IdentityState{}, UnifiedResolutionRecordV2{}, err
	}
	return next, record, nil
}

func (ResolverKeeperV2) VerifyReverseRecord(state IdentityState, record ReverseResolutionRecordV2, height uint64, authorizedAliasKeys []string) error {
	return ValidateReverseResolutionRecordV2(state, record, height, authorizedAliasKeys)
}

func (ResolverKeeperV2) ValidateVersionedTTL(record UnifiedResolutionRecordV2, currentVersion uint64, height uint64) error {
	if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
		return err
	}
	if record.RecordVersion != currentVersion {
		return errors.New("identity resolver record version mismatch")
	}
	if height > record.UpdatedAtHeight+record.RecordTTL {
		return errors.New("identity resolver record ttl expired")
	}
	return nil
}

func (DelegationKeeperV2) Authorize(record DelegationRecordV2, scope DelegationScopeV2, permission string, recordKey string, subtreeDepth uint8, height uint64) error {
	return ValidateDelegationRecordV2Use(record, scope, permission, recordKey, subtreeDepth, height)
}

func (DelegationKeeperV2) EnforceSubdomainCreate(record DelegationRecordV2, childLabel string, depth uint8, height uint64) error {
	if err := validateDomainLabel(childLabel); err != nil {
		return err
	}
	return ValidateDelegationRecordV2Use(record, DelegationScopeSubdomainCreate, "create", childLabel, depth, height)
}

func (DelegationKeeperV2) EnforceZoneExecution(record DelegationRecordV2, permission string, zoneKey string, height uint64) error {
	return ValidateDelegationRecordV2Use(record, DelegationScopeZoneAdmin, permission, zoneKey, 0, height)
}

func (AuctionKeeperV2) Start(state IdentityState, name string, startHeight uint64, minBid uint64, feeSplitID string) (IdentityState, AuctionRecordV2, error) {
	next, auction, err := StartSealedAuction(state, name, startHeight)
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	record, err := BuildAuctionRecordV2(auction, minBid, feeSplitID)
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	return next, record, nil
}

func (AuctionKeeperV2) CommitBid(state IdentityState, name string, bidder sdk.AccAddress, commitmentHash string, height uint64) (IdentityState, AuctionRecordV2, error) {
	next, auction, err := CommitAuctionBid(state, name, bidder, commitmentHash, height)
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	record, err := BuildAuctionRecordV2(auction, 1, "domain.fees")
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	return next, record, nil
}

func (AuctionKeeperV2) RevealBid(state IdentityState, name string, bidder sdk.AccAddress, bid uint64, salt string, height uint64) (IdentityState, AuctionRecordV2, error) {
	next, auction, err := RevealAuctionBid(state, name, bidder, bid, salt, height)
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	record, err := BuildAuctionRecordV2(auction, 1, "domain.fees")
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	return next, record, nil
}

func (AuctionKeeperV2) Finalize(state IdentityState, name string, height uint64, minBid uint64, feeSplitID string) (IdentityState, AuctionRecordV2, error) {
	next, auction, err := FinalizeSealedAuction(state, name, height)
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	if err := validateDeterministicAuctionWinnerV2(auction); err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	record, err := BuildAuctionRecordV2(auction, minBid, feeSplitID)
	if err != nil {
		return IdentityState{}, AuctionRecordV2{}, err
	}
	return next, record, nil
}

func (ProofKeeperV2) BuildResolutionProof(state IdentityState, name string, height uint64) (IdentityResolutionProof, error) {
	return BuildIdentityResolutionProof(state, name, height)
}

func (ProofKeeperV2) VerifyResolutionProof(proof IdentityResolutionProof, trustedRoot string, height uint64) (IdentityResolution, error) {
	return ValidateResolutionProofBoundary(proof, trustedRoot, height)
}

func (RoutingIntegrationKeeperV2) ResolveTransactionTarget(state IdentityState, name string, recordKey string, height uint64) (NamedExecutionTarget, error) {
	return ResolveNamedExecutionTarget(state, NamedExecutionRequest{
		Kind:		NamedExecutionSend,
		Name:		name,
		RecordKey:	recordKey,
	}, height)
}

func (RoutingIntegrationKeeperV2) ResolveContractInvocation(state IdentityState, name string, interfaceID string, method string, payloadHash string, height uint64) (NamedExecutionTarget, error) {
	return ResolveNamedExecutionTarget(state, NamedExecutionRequest{
		Kind:		NamedExecutionInvoke,
		Name:		name,
		InterfaceID:	interfaceID,
		Method:		method,
		PayloadHash:	payloadHash,
	}, height)
}

func (RoutingIntegrationKeeperV2) ResolveServiceMetadata(state IdentityState, name string, height uint64, ttl uint64, serviceKey string) (ServiceEndpointV2, error) {
	record, err := BuildUnifiedResolutionRecordV2(state, name, height, ttl)
	if err != nil {
		return ServiceEndpointV2{}, err
	}
	for _, endpoint := range record.ServiceEndpoints {
		if serviceEndpointIDV2(endpoint) == serviceKey {
			return endpoint, nil
		}
	}
	return ServiceEndpointV2{}, errors.New("identity routing service endpoint not found")
}

func ValidateIdentityKeeperInvariantsV2(input IdentityKeeperInvariantInputV2) error {
	state := normalizeIdentityStateParams(input.State)
	if err := state.Validate(); err != nil {
		return err
	}
	for _, domain := range state.Domains {
		if domain.ExpiryHeight > input.Height {
			if _, _, err := ValidateRegistryNFTAuthority(state, domain.Name, input.Height); err != nil {
				return errors.New("identity keeper invariant: active domain without valid NFT binding")
			}
		}
		if domain.ParentName != "" {
			parent, found := findDomain(state, domain.ParentName)
			if !found {
				return errors.New("identity keeper invariant: subdomain parent missing")
			}
			if domain.ExpiryHeight > parent.ExpiryHeight {
				return errors.New("identity keeper invariant: subdomain outlives parent")
			}
		}
	}
	for _, resolver := range state.Resolvers {
		authority, found := findResolverAuthorityDomain(state, resolver.Domain)
		if !found || !bytes.Equal(resolver.Owner, authority.Owner) {
			return errors.New("identity keeper invariant: resolver owner mismatch")
		}
	}
	for _, auction := range state.Auctions {
		if auction.Phase == AuctionPhaseFinalized {
			if err := validateDeterministicAuctionWinnerV2(auction); err != nil {
				return err
			}
		}
	}
	if len(input.CacheRecords) != len(input.CacheContexts) {
		return errors.New("identity keeper invariant: cache records and contexts length mismatch")
	}
	for i, cacheRecord := range input.CacheRecords {
		if err := ValidateResolutionCacheRecordV2Use(cacheRecord, input.CacheContexts[i]); err != nil {
			return errors.New("identity keeper invariant: cached resolution invalid after source change")
		}
	}
	return nil
}

func validateDeterministicAuctionWinnerV2(auction Auction) error {
	if err := validateAuction(auction); err != nil {
		return err
	}
	if auction.Phase != AuctionPhaseFinalized {
		return errors.New("identity auction is not finalized")
	}
	if len(auction.Reveals) == 0 {
		return errors.New("identity auction winner proof requires reveals")
	}
	winner := chooseAuctionWinner(auction.Reveals)
	if !bytes.Equal(winner.Bidder, auction.Winner) || winner.Bid != auction.WinningBid || winner.CommitmentHash != auction.WinningCommitment {
		return errors.New("identity auction deterministic winner proof mismatch")
	}
	return nil
}
