package types

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultIdentityQueryLimitV2	= uint64(50)
	MaxIdentityQueryLimitV2		= uint64(100)
)

type IdentityQueryCodeV2 string

const (
	IdentityQueryOK			IdentityQueryCodeV2	= "ok"
	IdentityQueryInvalidRequest	IdentityQueryCodeV2	= "invalid_request"
	IdentityQueryNotFound		IdentityQueryCodeV2	= "not_found"
	IdentityQueryExpired		IdentityQueryCodeV2	= "expired"
	IdentityQueryVerificationFailed	IdentityQueryCodeV2	= "verification_failed"
)

type IdentityQueryPageRequestV2 struct {
	Offset	uint64
	Limit	uint64
}

type IdentityQueryPageResponseV2 struct {
	Total		uint64
	NextOffset	uint64
}

type IdentityQueryContextV2 struct {
	State		IdentityState
	Height		uint64
	DefaultTTL	uint64
	Delegations	[]DelegationRecordV2
}

type IdentityQueryServiceV2 struct {
	ctx IdentityQueryContextV2
}

type IdentityQueryResponseV2 struct {
	Code			IdentityQueryCodeV2
	FailureCode		IdentityLightClientFailureCodeV2
	Error			string
	Height			uint64
	RecordVersion		uint64
	Page			IdentityQueryPageResponseV2
	Proof			*IdentityResolutionProof
	RecursiveProof		*RecursiveResolutionProofV2
	PathCommitment		*IdentityPathCommitmentV2
	AbsenceProof		*IdentityAbsenceProof
	Domain			*DomainRecordV2
	Domains			[]DomainRecordV2
	Binding			*DomainNFTBinding
	Resolver		*UnifiedResolutionRecordV2
	Address			sdk.AccAddress
	Target			*NamedExecutionTarget
	ContractTarget		*ContractTargetV2
	Service			*ServiceEndpointV2
	Services		[]ServiceEndpointV2
	Interface		*InterfaceDescriptorV2
	Route			*RoutingMetadataV2
	Reverse			*ReverseResolutionRecordV2
	Subdomains		[]SubdomainRecord
	Delegations		[]DelegationRecordV2
	Auction			*AuctionRecordV2
	Consistency		*IdentityConsistencyAuditResultV2
	ProofResult		*IdentityResolution
	Lifecycle		DomainLifecycleStatus
	Params			*IdentityParams
	RegistrationPrice	*IdentityDomainPriceQuoteV2
	RenewalPrice		*IdentityDomainPriceQuoteV2
}

func NewIdentityQueryServiceV2(ctx IdentityQueryContextV2) IdentityQueryServiceV2 {
	if ctx.DefaultTTL == 0 {
		ctx.DefaultTTL = 1
	}
	ctx.State = normalizeIdentityStateParams(ctx.State)
	return IdentityQueryServiceV2{ctx: ctx}
}

func (q IdentityQueryServiceV2) QueryDomain(nameHash string, includeProof bool) IdentityQueryResponseV2 {
	if err := validateHexHash("identity v2 query domain name hash", nameHash); err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	domain, found, err := q.findDomainByHash(nameHash)
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	if !found {
		return q.failure(IdentityQueryNotFound, errors.New("identity v2 query domain not found"))
	}
	return q.domainResponse(domain, includeProof)
}

func (q IdentityQueryServiceV2) QueryDomainByName(name string, includeProof bool) IdentityQueryResponseV2 {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return q.failureTyped(IdentityQueryInvalidRequest, IdentityLightClientErrInvalidName, err, nil, 0)
	}
	domain, found := findDomain(q.ctx.State, normalized)
	if !found {
		var absence *IdentityAbsenceProof
		if includeProof {
			proof, err := BuildIdentityAbsenceProof(q.ctx.State, mustIdentityDomainStoreKeyV2(normalized))
			if err != nil {
				return q.failureTyped(IdentityQueryVerificationFailed, IdentityLightClientErrProofInvalid, err, nil, 0)
			}
			absence = &proof
		}
		return q.failureTyped(IdentityQueryNotFound, IdentityLightClientErrDomainNotFound, errors.New("identity v2 query domain not found"), absence, 0)
	}
	return q.domainResponse(domain, includeProof)
}

func (q IdentityQueryServiceV2) QueryDomainsByOwner(owner sdk.AccAddress, page IdentityQueryPageRequestV2) IdentityQueryResponseV2 {
	if err := validateSpecAddress("identity v2 query owner", owner); err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	records := make([]DomainRecordV2, 0)
	for _, domain := range q.ctx.State.Domains {
		if !bytes.Equal(domain.Owner, owner) {
			continue
		}
		record, err := q.domainRecordV2(domain)
		if err != nil {
			return q.failure(IdentityQueryInvalidRequest, err)
		}
		records = append(records, record)
	}
	window, pageResp := paginateDomainRecordsV2(records, page)
	resp := q.ok()
	resp.Domains = window
	resp.Page = pageResp
	return resp
}

func (q IdentityQueryServiceV2) QueryDomainNFTBinding(nameHash string) IdentityQueryResponseV2 {
	if err := validateHexHash("identity v2 query nft binding name hash", nameHash); err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	domain, found, err := q.findDomainByHash(nameHash)
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	if !found {
		return q.failure(IdentityQueryNotFound, errors.New("identity v2 query domain not found"))
	}
	nft, found := findDomainNFTByID(q.ctx.State, domain.NFTID)
	if !found {
		return q.failure(IdentityQueryNotFound, errors.New("identity v2 query domain nft binding not found"))
	}
	binding, err := NewDomainNFTBinding(domain.Name, nft.ID, nft.Owner, q.ctx.Height)
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	resp := q.ok()
	resp.Binding = &binding
	resp.RecordVersion = binding.BindingVersion
	return resp
}

func (q IdentityQueryServiceV2) QueryResolver(name string, includeProof bool) IdentityQueryResponseV2 {
	record, err := BuildUnifiedResolutionRecordV2(q.ctx.State, name, q.ctx.Height, q.ctx.DefaultTTL)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	record.RecordVersion = q.resolverRecordVersionForName(name, record.RecordVersion)
	resp := q.ok()
	resp.Resolver = &record
	resp.RecordVersion = record.RecordVersion
	if includeProof {
		proof, err := BuildIdentityResolutionProof(q.ctx.State, name, q.ctx.Height)
		if err != nil {
			return q.failure(IdentityQueryVerificationFailed, err)
		}
		resp.Proof = &proof
	}
	return resp
}

func (q IdentityQueryServiceV2) QueryResolvePrimary(name string) IdentityQueryResponseV2 {
	resolution, err := ResolveIdentityRecordRecursive(q.ctx.State, name, q.ctx.Height)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	if len(resolution.Record.Primary) == 0 {
		return q.failureTyped(IdentityQueryNotFound, IdentityLightClientErrTargetNotFound, errors.New("identity v2 query primary target not found"), nil, ResolverRecordVersionV2(resolution.Record))
	}
	resp := q.ok()
	resp.Address = cloneSpecAddress(resolution.Record.Primary)
	resp.RecordVersion = uint64(resolution.Record.UpdatedAtUnix)
	return resp
}

func (q IdentityQueryServiceV2) QueryResolveTarget(name string, recordKey string) IdentityQueryResponseV2 {
	target, err := ResolveNamedExecutionTarget(q.ctx.State, NamedExecutionRequest{Kind: NamedExecutionSend, Name: name, RecordKey: recordKey}, q.ctx.Height)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	resp := q.ok()
	resp.Target = &target
	resp.Address = cloneSpecAddress(target.Address)
	return resp
}

func (q IdentityQueryServiceV2) QueryResolveContractTarget(name string, targetID string) IdentityQueryResponseV2 {
	record, target, err := ResolveIdentityContractTargetByNameV2(q.ctx.State, name, targetID, q.ctx.Height, q.ctx.DefaultTTL)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	record.RecordVersion = q.resolverRecordVersionForName(name, record.RecordVersion)
	resp := q.ok()
	resp.ContractTarget = &target
	resp.Address = cloneSpecAddress(contractTargetAddressV2(target))
	resp.RecordVersion = record.RecordVersion
	return resp
}

func (q IdentityQueryServiceV2) QueryResolveService(name string, service string) IdentityQueryResponseV2 {
	return q.QueryResolveServiceRecord(name, service, false)
}

func (q IdentityQueryServiceV2) QueryResolveServiceRecord(name string, service string, includeFallbacks bool) IdentityQueryResponseV2 {
	record, err := BuildUnifiedResolutionRecordV2(q.ctx.State, name, q.ctx.Height, q.ctx.DefaultTTL)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	record.RecordVersion = q.resolverRecordVersionForName(name, record.RecordVersion)
	for _, endpoint := range record.ServiceEndpoints {
		if serviceEndpointIDV2(endpoint) == service {
			resp := q.ok()
			resp.Service = &endpoint
			if includeFallbacks {
				resp.Services = compatibleServiceEndpointsV2(record.ServiceEndpoints, IdentityServiceDiscoveryRequestV2{SupportedServiceTypes: []string{endpoint.ServiceType}})
			}
			resp.RecordVersion = record.RecordVersion
			return resp
		}
	}
	return q.failureTyped(IdentityQueryNotFound, IdentityLightClientErrTargetNotFound, errors.New("identity v2 query service endpoint not found"), nil, record.RecordVersion)
}

func (q IdentityQueryServiceV2) QueryResolveInterface(name string, interfaceID string) IdentityQueryResponseV2 {
	record, err := BuildUnifiedResolutionRecordV2(q.ctx.State, name, q.ctx.Height, q.ctx.DefaultTTL)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	record.RecordVersion = q.resolverRecordVersionForName(name, record.RecordVersion)
	for _, descriptor := range record.InterfaceDescriptors {
		if descriptor.InterfaceID == interfaceID {
			resp := q.ok()
			resp.Interface = &descriptor
			resp.RecordVersion = record.RecordVersion
			return resp
		}
	}
	return q.failureTyped(IdentityQueryNotFound, IdentityLightClientErrTargetNotFound, errors.New("identity v2 query interface descriptor not found"), nil, record.RecordVersion)
}

func (q IdentityQueryServiceV2) QueryResolveRoute(name string) IdentityQueryResponseV2 {
	record, err := BuildUnifiedResolutionRecordV2(q.ctx.State, name, q.ctx.Height, q.ctx.DefaultTTL)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	record.RecordVersion = q.resolverRecordVersionForName(name, record.RecordVersion)
	resp := q.ok()
	resp.Route = &record.RoutingMetadata
	resp.RecordVersion = record.RecordVersion
	return resp
}

func (q IdentityQueryServiceV2) QueryReverse(address sdk.AccAddress) IdentityQueryResponseV2 {
	record, err := q.reverseRecord(address, false, nil)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	resp := q.ok()
	resp.Reverse = &record
	resp.RecordVersion = uint64(record.UpdatedAtHeight)
	return resp
}

func (q IdentityQueryServiceV2) QueryVerifiedReverse(address sdk.AccAddress, authorizedAliasKeys []string) IdentityQueryResponseV2 {
	record, err := q.reverseRecord(address, true, authorizedAliasKeys)
	if err != nil {
		return q.failure(queryCodeForResolutionErrorV2(err), err)
	}
	resp := q.ok()
	resp.Reverse = &record
	resp.RecordVersion = uint64(record.UpdatedAtHeight)
	return resp
}

func (q IdentityQueryServiceV2) QuerySubdomains(parentName string, page IdentityQueryPageRequestV2) IdentityQueryResponseV2 {
	parent, err := NormalizeAETDomain(parentName)
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	records := make([]SubdomainRecord, 0)
	for _, record := range q.ctx.State.Subdomains {
		if record.ParentName == parent {
			records = append(records, cloneSubdomainRecord(record))
		}
	}
	window, pageResp := paginateSubdomainRecordsV2(records, page)
	resp := q.ok()
	resp.Subdomains = window
	resp.Page = pageResp
	return resp
}

func (q IdentityQueryServiceV2) QueryDelegations(nameHash string, page IdentityQueryPageRequestV2) IdentityQueryResponseV2 {
	if err := validateHexHash("identity v2 query delegation name hash", nameHash); err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	records := make([]DelegationRecordV2, 0)
	for _, record := range q.ctx.Delegations {
		if record.NameHash == nameHash {
			records = append(records, cloneDelegationRecordV2(record))
		}
	}
	window, pageResp := paginateDelegationRecordsV2(records, page)
	resp := q.ok()
	resp.Delegations = window
	resp.Page = pageResp
	return resp
}

func (q IdentityQueryServiceV2) QueryAuction(auctionID string, nameHash string) IdentityQueryResponseV2 {
	if err := validateIdentityTxAuctionIDOrNameHashV2(auctionID, nameHash); err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	for _, auction := range q.ctx.State.Auctions {
		record, err := BuildAuctionRecordV2(auction, 1, "domain.fees")
		if err != nil {
			return q.failure(IdentityQueryInvalidRequest, err)
		}
		if (auctionID != "" && record.AuctionID == auctionID) || (nameHash != "" && record.NameHash == nameHash) {
			resp := q.ok()
			resp.Auction = &record
			resp.RecordVersion = uint64(len(auction.Commitments) + len(auction.Reveals))
			return resp
		}
	}
	return q.failure(IdentityQueryNotFound, errors.New("identity v2 query auction not found"))
}

func (q IdentityQueryServiceV2) QueryResolutionProof(name string) IdentityQueryResponseV2 {
	proof, err := BuildIdentityResolutionProof(q.ctx.State, name, q.ctx.Height)
	if err != nil {
		return q.failure(IdentityQueryVerificationFailed, err)
	}
	resp := q.ok()
	resp.Proof = &proof
	resp.RecordVersion = uint64(len(proof.Candidates))
	return resp
}

func (q IdentityQueryServiceV2) QueryRecursiveResolutionProof(name string) IdentityQueryResponseV2 {
	proofResp := q.QueryResolutionProof(name)
	if proofResp.Code != IdentityQueryOK {
		return proofResp
	}
	resolution, err := VerifyIdentityResolutionProof(*proofResp.Proof, q.ctx.Height)
	if err != nil {
		return q.failure(IdentityQueryVerificationFailed, err)
	}
	proofResp.ProofResult = &resolution
	return proofResp
}

func (q IdentityQueryServiceV2) QueryOptimizedRecursiveResolutionProof(rootName string, targetName string, cache *ResolutionCacheRecordV2, sourceVersion uint64, parentEpoch uint64, childEpoch uint64, lightClient bool, proofVerified bool) IdentityQueryResponseV2 {
	proof, commitment, err := BuildOptimizedRecursiveResolutionProofV2(OptimizedRecursiveResolutionProofRequestV2{
		State:		q.ctx.State,
		ChainID:	"identity-query-local",
		RootName:	rootName,
		TargetName:	targetName,
		Height:		q.ctx.Height,
		TTL:		q.ctx.DefaultTTL,
		Cache:		cache,
		SourceVersion:	sourceVersion,
		ParentEpoch:	parentEpoch,
		ChildEpoch:	childEpoch,
		LightClient:	lightClient,
		ProofVerified:	proofVerified,
	})
	if err != nil {
		return q.failure(IdentityQueryVerificationFailed, err)
	}
	resp := q.ok()
	resp.RecursiveProof = &proof
	resp.PathCommitment = &commitment
	resp.RecordVersion = sourceVersion
	return resp
}

func (q IdentityQueryServiceV2) QueryDomainLifecycle(name string) IdentityQueryResponseV2 {
	lifecycle, err := DomainLifecycle(q.ctx.State, name, q.ctx.Height)
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	resp := q.ok()
	resp.Lifecycle = lifecycle
	return resp
}

func (q IdentityQueryServiceV2) QueryIdentityParams() IdentityQueryResponseV2 {
	params := normalizeIdentityParams(q.ctx.State.Params)
	resp := q.ok()
	resp.Params = &params
	resp.RecordVersion = 1
	return resp
}

func (q IdentityQueryServiceV2) QueryRegistrationPrice(name string, durationBlocks uint64, demandClass IdentityDemandClassV2, auction bool, resolverPayloadBytes uint64, subdomainMode IdentitySubdomainModeV2) IdentityQueryResponseV2 {
	quote, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{
		Name:			name,
		DurationBlocks:		durationBlocks,
		DemandClass:		demandClass,
		Auction:		auction,
		ResolverPayloadBytes:	resolverPayloadBytes,
		SubdomainMode:		subdomainMode,
	}, DefaultIdentityPricingParamsV2())
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	resp := q.ok()
	resp.RegistrationPrice = &quote
	resp.RecordVersion = quote.DurationBlocks
	return resp
}

func (q IdentityQueryServiceV2) QueryRenewalPrice(name string, periods uint32, resolverPayloadBytes uint64) IdentityQueryResponseV2 {
	quote, err := QuoteIdentityRenewalPriceV2(name, periods, resolverPayloadBytes, DefaultIdentityPricingParamsV2())
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	resp := q.ok()
	resp.RenewalPrice = &quote
	resp.RecordVersion = uint64(quote.RenewalPeriods)
	return resp
}

func (q IdentityQueryServiceV2) QueryPeriodicConsistencyAudit() IdentityQueryResponseV2 {
	audit := QueryIdentityPeriodicConsistencyAuditV2(IdentityConsistencyAuditRequestV2{
		State:		q.ctx.State,
		Height:		q.ctx.Height,
		Delegations:	q.ctx.Delegations,
	})
	resp := q.ok()
	resp.Consistency = &audit
	if !audit.Valid {
		resp.Code = IdentityQueryVerificationFailed
		if len(audit.Issues) > 0 {
			resp.Error = audit.Issues[0].Message
		}
	}
	return resp
}

func (q IdentityQueryServiceV2) ok() IdentityQueryResponseV2 {
	return IdentityQueryResponseV2{Code: IdentityQueryOK, Height: q.ctx.Height}
}

func (q IdentityQueryServiceV2) failure(code IdentityQueryCodeV2, err error) IdentityQueryResponseV2 {
	return q.failureTyped(code, queryFailureCodeV2(code), err, nil, 0)
}

func (q IdentityQueryServiceV2) failureTyped(code IdentityQueryCodeV2, failureCode IdentityLightClientFailureCodeV2, err error, absence *IdentityAbsenceProof, recordVersion uint64) IdentityQueryResponseV2 {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return IdentityQueryResponseV2{Code: code, FailureCode: failureCode, Error: message, Height: q.ctx.Height, AbsenceProof: absence, RecordVersion: recordVersion}
}

func (q IdentityQueryServiceV2) domainResponse(domain Domain, includeProof bool) IdentityQueryResponseV2 {
	record, err := q.domainRecordV2(domain)
	if err != nil {
		return q.failure(IdentityQueryInvalidRequest, err)
	}
	resp := q.ok()
	resp.Domain = &record
	resp.RecordVersion = record.Version
	if includeProof {
		proof, err := BuildIdentityResolutionProof(q.ctx.State, domain.Name, q.ctx.Height)
		if err != nil {
			return q.failure(IdentityQueryVerificationFailed, err)
		}
		resp.Proof = &proof
	}
	return resp
}

func (q IdentityQueryServiceV2) domainRecordV2(domain Domain) (DomainRecordV2, error) {
	status, err := domainRecordStatusForQueryV2(q.ctx.State, domain.Name, q.ctx.Height)
	if err != nil {
		return DomainRecordV2{}, err
	}
	return NewDomainRecordV2FromDomain(domain, status, 0, q.ctx.Height)
}

func (q IdentityQueryServiceV2) resolverRecordVersionForName(name string, fallback uint64) uint64 {
	resolution, err := ResolveIdentityRecordRecursive(q.ctx.State, name, q.ctx.Height)
	if err != nil {
		return fallback
	}
	version := ResolverRecordVersionV2(resolution.Record)
	if version == 0 {
		return fallback
	}
	return version
}

func (q IdentityQueryServiceV2) findDomainByHash(nameHash string) (Domain, bool, error) {
	for _, domain := range q.ctx.State.Domains {
		hash, err := DomainRecordV2NameHash(domain.Name)
		if err != nil {
			return Domain{}, false, err
		}
		if hash == nameHash {
			return cloneDomain(domain), true, nil
		}
	}
	return Domain{}, false, nil
}

func (q IdentityQueryServiceV2) reverseRecord(address sdk.AccAddress, verified bool, authorizedAliasKeys []string) (ReverseResolutionRecordV2, error) {
	if err := validateSpecAddress("identity v2 query reverse address", address); err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	for _, reverse := range q.ctx.State.ReverseRecords {
		if !bytes.Equal(reverse.Address, address) {
			continue
		}
		domain, found := findDomain(q.ctx.State, reverse.Domain)
		if !found {
			return ReverseResolutionRecordV2{}, errors.New("identity v2 query reverse domain not found")
		}
		record, err := NewReverseResolutionRecordV2(address, domain.Name, verified, uint64(reverse.UpdatedAtUnix), domain.ExpiryHeight)
		if err != nil {
			return ReverseResolutionRecordV2{}, err
		}
		if verified {
			if err := ValidateReverseResolutionRecordV2(q.ctx.State, record, q.ctx.Height, authorizedAliasKeys); err != nil {
				return ReverseResolutionRecordV2{}, err
			}
		}
		return record, nil
	}
	return ReverseResolutionRecordV2{}, errors.New("identity v2 query reverse record not found")
}

func domainRecordStatusForQueryV2(state IdentityState, name string, height uint64) (DomainRecordV2Status, error) {
	status, err := DomainLifecycle(state, name, height)
	if err != nil {
		return "", err
	}
	switch status {
	case DomainLifecycleCommitted:
		return DomainRecordV2Committed, nil
	case DomainLifecycleAvailable:
		return DomainRecordV2Available, nil
	case DomainLifecycleActive:
		return DomainRecordV2Active, nil
	case DomainLifecycleRenewalWindow:
		return DomainRecordV2RenewalWindow, nil
	case DomainLifecycleExpired:
		return DomainRecordV2Expired, nil
	default:
		return "", fmt.Errorf("unsupported identity lifecycle status %q", status)
	}
}

func queryCodeForResolutionErrorV2(err error) IdentityQueryCodeV2 {
	if err == nil {
		return IdentityQueryOK
	}
	message := err.Error()
	switch {
	case bytes.Contains([]byte(message), []byte("not found")):
		return IdentityQueryNotFound
	case bytes.Contains([]byte(message), []byte("expired")):
		return IdentityQueryExpired
	case bytes.Contains([]byte(message), []byte("proof")), bytes.Contains([]byte(message), []byte("verified")), bytes.Contains([]byte(message), []byte("forward")):
		return IdentityQueryVerificationFailed
	default:
		return IdentityQueryInvalidRequest
	}
}

func normalizeIdentityQueryPageV2(page IdentityQueryPageRequestV2, total int) (start int, end int, resp IdentityQueryPageResponseV2) {
	limit := page.Limit
	if limit == 0 {
		limit = DefaultIdentityQueryLimitV2
	}
	if limit > MaxIdentityQueryLimitV2 {
		limit = MaxIdentityQueryLimitV2
	}
	if page.Offset >= uint64(total) {
		return total, total, IdentityQueryPageResponseV2{Total: uint64(total)}
	}
	start = int(page.Offset)
	end = start + int(limit)
	if end > total {
		end = total
	}
	next := uint64(0)
	if end < total {
		next = uint64(end)
	}
	return start, end, IdentityQueryPageResponseV2{Total: uint64(total), NextOffset: next}
}

func paginateDomainRecordsV2(records []DomainRecordV2, page IdentityQueryPageRequestV2) ([]DomainRecordV2, IdentityQueryPageResponseV2) {
	start, end, resp := normalizeIdentityQueryPageV2(page, len(records))
	return append([]DomainRecordV2(nil), records[start:end]...), resp
}

func paginateSubdomainRecordsV2(records []SubdomainRecord, page IdentityQueryPageRequestV2) ([]SubdomainRecord, IdentityQueryPageResponseV2) {
	start, end, resp := normalizeIdentityQueryPageV2(page, len(records))
	out := make([]SubdomainRecord, 0, end-start)
	for _, record := range records[start:end] {
		out = append(out, cloneSubdomainRecord(record))
	}
	return out, resp
}

func paginateDelegationRecordsV2(records []DelegationRecordV2, page IdentityQueryPageRequestV2) ([]DelegationRecordV2, IdentityQueryPageResponseV2) {
	start, end, resp := normalizeIdentityQueryPageV2(page, len(records))
	out := make([]DelegationRecordV2, 0, end-start)
	for _, record := range records[start:end] {
		out = append(out, cloneDelegationRecordV2(record))
	}
	return out, resp
}

func cloneSubdomainRecord(record SubdomainRecord) SubdomainRecord {
	record.Owner = cloneSpecAddress(record.Owner)
	return record
}

func cloneDelegationRecordV2(record DelegationRecordV2) DelegationRecordV2 {
	record.Delegate = cloneSpecAddress(record.Delegate)
	record.Permissions = append([]string(nil), record.Permissions...)
	return record
}
