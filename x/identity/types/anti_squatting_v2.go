package types

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultIdentityCommitmentDepositNaet		= int64(10_000_000_000)
	DefaultIdentityResolverStorageFeePerByte	= int64(1_000_000)
	DefaultIdentityGraceRecoveryCostNaet		= int64(25_000_000_000)
	DefaultIdentityBulkRegistrationWindow		= uint64(720)
	DefaultIdentityBulkRegistrationLimit		= uint32(5)
	IdentityAntiSquattingResultAllowedV2		= "allowed"
	IdentityAntiSquattingResultBulkLimitedV2	= "bulk_limited"
	IdentityAntiSquattingResultInvalidInputV2	= "invalid_input"
)

type IdentityAntiSquattingParamsV2 struct {
	DomainParams				DomainParams
	CommitmentDeposit			sdkmath.Int
	ResolverStorageFeePerByte		sdkmath.Int
	GraceRecoveryCost			sdkmath.Int
	BulkRegistrationWindowBlocks		uint64
	MaxBulkRegistrationsPerAccount		uint32
	BulkRegistrationLimitEnabled		bool
	ContestedNameAuctionRequired		bool
	ExpiredDomainGracePeriodBlocks		uint64
	ExpiredDomainReleaseWindowBlocks	uint64
}

type IdentityRegistrationPriceQuoteV2 struct {
	Name				string
	PricingLabel			string
	RegistrationFee			sdkmath.Int
	RenewalFee			sdkmath.Int
	CommitmentDeposit		sdkmath.Int
	ResolverStorageFee		sdkmath.Int
	GraceRecoveryCost		sdkmath.Int
	TotalRegistrationCost		sdkmath.Int
	TotalRenewalCost		sdkmath.Int
	ResolverPayloadBytes		uint64
	ContestedAuctionRequired	bool
	DeterministicFormula		string
}

type IdentityBulkRegistrationAttemptV2 struct {
	Owner			sdk.AccAddress
	Name			string
	Height			uint64
	ResolverPayloadBytes	uint64
}

type IdentityBulkRegistrationResultV2 struct {
	InputIndex	uint32
	Owner		sdk.AccAddress
	Name		string
	WindowID	uint64
	Status		string
	Error		string
	Quote		*IdentityRegistrationPriceQuoteV2
}

type IdentityBulkRegistrationSimulationV2 struct {
	ResultOrder	string
	Allowed		uint32
	Rejected	uint32
	Results		[]IdentityBulkRegistrationResultV2
}

func DefaultIdentityAntiSquattingParamsV2() IdentityAntiSquattingParamsV2 {
	return IdentityAntiSquattingParamsV2{
		DomainParams:				DefaultDomainParams(),
		CommitmentDeposit:			sdkmath.NewInt(DefaultIdentityCommitmentDepositNaet),
		ResolverStorageFeePerByte:		sdkmath.NewInt(DefaultIdentityResolverStorageFeePerByte),
		GraceRecoveryCost:			sdkmath.NewInt(DefaultIdentityGraceRecoveryCostNaet),
		BulkRegistrationWindowBlocks:		DefaultIdentityBulkRegistrationWindow,
		MaxBulkRegistrationsPerAccount:		DefaultIdentityBulkRegistrationLimit,
		BulkRegistrationLimitEnabled:		true,
		ContestedNameAuctionRequired:		true,
		ExpiredDomainGracePeriodBlocks:		DefaultRenewalWindowBlocks,
		ExpiredDomainReleaseWindowBlocks:	DefaultRenewalWindowBlocks + 1,
	}
}

func ValidateIdentityAntiSquattingParamsV2(params IdentityAntiSquattingParamsV2) error {
	if err := ValidateDomainParams(params.DomainParams); err != nil {
		return err
	}
	for _, amount := range []struct {
		label	string
		value	sdkmath.Int
	}{
		{label: "commitment deposit", value: params.CommitmentDeposit},
		{label: "resolver storage fee per byte", value: params.ResolverStorageFeePerByte},
		{label: "grace recovery cost", value: params.GraceRecoveryCost},
	} {
		if amount.value.IsNil() || amount.value.IsNegative() {
			return fmt.Errorf("identity anti-squatting %s must not be negative", amount.label)
		}
	}
	if params.BulkRegistrationLimitEnabled {
		if params.BulkRegistrationWindowBlocks == 0 {
			return errors.New("identity anti-squatting bulk registration window is required")
		}
		if params.MaxBulkRegistrationsPerAccount == 0 {
			return errors.New("identity anti-squatting bulk registration limit is required")
		}
	}
	if params.ExpiredDomainReleaseWindowBlocks != 0 && params.ExpiredDomainReleaseWindowBlocks <= params.ExpiredDomainGracePeriodBlocks {
		return errors.New("identity anti-squatting release window must exceed grace period")
	}
	return nil
}

func QuoteIdentityRegistrationPriceV2(name string, resolverPayloadBytes uint64, contested bool, params IdentityAntiSquattingParamsV2) (IdentityRegistrationPriceQuoteV2, error) {
	if err := ValidateIdentityAntiSquattingParamsV2(params); err != nil {
		return IdentityRegistrationPriceQuoteV2{}, err
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityRegistrationPriceQuoteV2{}, err
	}
	pricingLabel, err := IdentityPricingLabelV2(normalized)
	if err != nil {
		return IdentityRegistrationPriceQuoteV2{}, err
	}
	registrationFee, err := StartPriceForName(pricingLabel, params.DomainParams)
	if err != nil {
		return IdentityRegistrationPriceQuoteV2{}, err
	}
	renewalFee, err := RenewalFee(pricingLabel, params.DomainParams)
	if err != nil {
		return IdentityRegistrationPriceQuoteV2{}, err
	}
	storageFee := params.ResolverStorageFeePerByte.Mul(sdkmath.NewIntFromUint64(resolverPayloadBytes))
	quote := IdentityRegistrationPriceQuoteV2{
		Name:				normalized,
		PricingLabel:			pricingLabel,
		RegistrationFee:		registrationFee,
		RenewalFee:			renewalFee,
		CommitmentDeposit:		params.CommitmentDeposit,
		ResolverStorageFee:		storageFee,
		GraceRecoveryCost:		params.GraceRecoveryCost,
		ResolverPayloadBytes:		resolverPayloadBytes,
		ContestedAuctionRequired:	contested && params.ContestedNameAuctionRequired,
		DeterministicFormula:		"length_price + commitment_deposit + resolver_payload_bytes*storage_fee_per_byte",
	}
	quote.TotalRegistrationCost = registrationFee.Add(params.CommitmentDeposit).Add(storageFee)
	quote.TotalRenewalCost = renewalFee
	return quote, nil
}

func IdentityPricingLabelV2(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	labelsPart := strings.TrimSuffix(normalized, DomainTLD)
	labels := strings.Split(labelsPart, ".")
	if len(labels) == 0 || labels[0] == "" {
		return "", errors.New("identity pricing label is required")
	}
	if err := ValidateDomainName(labels[0]); err != nil {
		return "", err
	}
	return labels[0], nil
}

func SimulateIdentityBulkRegistrationWindowV2(attempts []IdentityBulkRegistrationAttemptV2, params IdentityAntiSquattingParamsV2) IdentityBulkRegistrationSimulationV2 {
	sim := IdentityBulkRegistrationSimulationV2{
		ResultOrder:	"input_index",
		Results:	make([]IdentityBulkRegistrationResultV2, 0, len(attempts)),
	}
	if err := ValidateIdentityAntiSquattingParamsV2(params); err != nil {
		for i, attempt := range attempts {
			sim.Rejected++
			sim.Results = append(sim.Results, IdentityBulkRegistrationResultV2{InputIndex: uint32(i), Owner: cloneSpecAddress(attempt.Owner), Name: attempt.Name, Status: IdentityAntiSquattingResultInvalidInputV2, Error: err.Error()})
		}
		return sim
	}
	windowCounts := map[string]uint32{}
	for i, attempt := range attempts {
		result := IdentityBulkRegistrationResultV2{
			InputIndex:	uint32(i),
			Owner:		cloneSpecAddress(attempt.Owner),
			Name:		attempt.Name,
			Status:		IdentityAntiSquattingResultAllowedV2,
		}
		normalized, err := NormalizeAETDomain(attempt.Name)
		if err != nil {
			result.Status = IdentityAntiSquattingResultInvalidInputV2
			result.Error = err.Error()
			sim.Rejected++
			sim.Results = append(sim.Results, result)
			continue
		}
		result.Name = normalized
		if err := validateSpecAddress("identity bulk registration owner", attempt.Owner); err != nil {
			result.Status = IdentityAntiSquattingResultInvalidInputV2
			result.Error = err.Error()
			sim.Rejected++
			sim.Results = append(sim.Results, result)
			continue
		}
		if params.BulkRegistrationWindowBlocks != 0 {
			result.WindowID = attempt.Height / params.BulkRegistrationWindowBlocks
		}
		quote, err := QuoteIdentityRegistrationPriceV2(normalized, attempt.ResolverPayloadBytes, false, params)
		if err != nil {
			result.Status = IdentityAntiSquattingResultInvalidInputV2
			result.Error = err.Error()
			sim.Rejected++
			sim.Results = append(sim.Results, result)
			continue
		}
		result.Quote = &quote
		if params.BulkRegistrationLimitEnabled {
			key := string(attempt.Owner) + "/" + fmt.Sprintf("%020d", result.WindowID)
			if windowCounts[key] >= params.MaxBulkRegistrationsPerAccount {
				result.Status = IdentityAntiSquattingResultBulkLimitedV2
				result.Error = "identity anti-squatting bulk registration limit exceeded"
				sim.Rejected++
				sim.Results = append(sim.Results, result)
				continue
			}
			windowCounts[key]++
		}
		sim.Allowed++
		sim.Results = append(sim.Results, result)
	}
	return sim
}

func RecoverExpiredIdentityDomainV2(state IdentityState, name string, actor sdk.AccAddress, height uint64, payment sdkmath.Int, params IdentityAntiSquattingParamsV2) (IdentityState, Domain, error) {
	if err := ValidateIdentityAntiSquattingParamsV2(params); err != nil {
		return IdentityState{}, Domain{}, err
	}
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Domain{}, err
	}
	domain, found := findDomain(state, name)
	if !found {
		return IdentityState{}, Domain{}, errors.New("identity recovery domain not found")
	}
	if height < domain.ExpiryHeight {
		return IdentityState{}, Domain{}, errors.New("identity recovery requires expired domain")
	}
	if params.ExpiredDomainGracePeriodBlocks != 0 && height > domain.ExpiryHeight+params.ExpiredDomainGracePeriodBlocks {
		return IdentityState{}, Domain{}, errors.New("identity recovery grace period ended")
	}
	if !bytes.Equal(actor, domain.Owner) {
		return IdentityState{}, Domain{}, errors.New("identity recovery requires prior owner")
	}
	if payment.IsNil() || payment.LT(params.GraceRecoveryCost) {
		return IdentityState{}, Domain{}, errors.New("identity recovery requires grace recovery cost")
	}
	base := height
	domain.ExpiryHeight = base + normalizeIdentityParams(state.Params).RegistrationPeriodBlocks
	domain.UpdatedHeight = height
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	sortIdentityState(&next)
	return next, domain, next.Validate()
}

func ReleaseExpiredIdentityDomainV2(state IdentityState, name string, height uint64, params IdentityAntiSquattingParamsV2) (IdentityState, Domain, error) {
	if err := ValidateIdentityAntiSquattingParamsV2(params); err != nil {
		return IdentityState{}, Domain{}, err
	}
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Domain{}, err
	}
	domain, found := findDomain(state, name)
	if !found {
		return IdentityState{}, Domain{}, errors.New("identity release domain not found")
	}
	if height < domain.ExpiryHeight+params.ExpiredDomainReleaseWindowBlocks {
		return IdentityState{}, Domain{}, errors.New("identity release window has not elapsed")
	}
	next := state.Clone()
	next.Domains = removeIdentityDomainByNameV2(next.Domains, domain.Name)
	next.DomainNFTs = removeIdentityNFTByIDV2(next.DomainNFTs, domain.NFTID)
	next.Resolvers = removeIdentityResolverByDomainV2(next.Resolvers, domain.Name)
	next.ReverseRecords = removeIdentityReverseByDomainV2(next.ReverseRecords, domain.Name)
	next.Subdomains = removeIdentitySubdomainByNameV2(next.Subdomains, domain.Name)
	sortIdentityState(&next)
	return next, domain, next.Validate()
}

func removeIdentityDomainByNameV2(domains []Domain, name string) []Domain {
	out := make([]Domain, 0, len(domains))
	for _, domain := range domains {
		if domain.Name == name {
			continue
		}
		out = append(out, cloneDomain(domain))
	}
	return out
}

func removeIdentityNFTByIDV2(nfts []DomainNFT, id string) []DomainNFT {
	out := make([]DomainNFT, 0, len(nfts))
	for _, nft := range nfts {
		if nft.ID == id {
			continue
		}
		out = append(out, cloneDomainNFT(nft))
	}
	return out
}

func removeIdentityResolverByDomainV2(records []ResolverRecord, domain string) []ResolverRecord {
	out := make([]ResolverRecord, 0, len(records))
	for _, record := range records {
		if record.Domain == domain {
			continue
		}
		out = append(out, cloneResolver(record))
	}
	return out
}

func removeIdentityReverseByDomainV2(records []ReverseRecord, domain string) []ReverseRecord {
	out := make([]ReverseRecord, 0, len(records))
	for _, record := range records {
		if record.Domain == domain {
			continue
		}
		out = append(out, cloneReverseRecord(record))
	}
	return out
}

func removeIdentitySubdomainByNameV2(records []SubdomainRecord, name string) []SubdomainRecord {
	out := make([]SubdomainRecord, 0, len(records))
	for _, record := range records {
		if record.Name == name || record.ParentName == name {
			continue
		}
		out = append(out, record)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
