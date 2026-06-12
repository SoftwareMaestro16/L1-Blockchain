package types

import (
	"bytes"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	IdentityDemandClassStandardV2	= IdentityDemandClassV2("standard")
	IdentityDemandClassPremiumV2	= IdentityDemandClassV2("premium")
	IdentityDemandClassContestedV2	= IdentityDemandClassV2("contested")

	IdentitySubdomainModeRootV2		= IdentitySubdomainModeV2("root")
	IdentitySubdomainModeOwnerControlledV2	= IdentitySubdomainModeV2("owner_controlled")
	IdentitySubdomainModeDetachedPaidV2	= IdentitySubdomainModeV2("detached_paid")
	IdentitySubdomainModeEphemeralV2	= IdentitySubdomainModeV2("ephemeral_service")

	DefaultIdentityPricingBaseDurationBlocks	= DefaultRegistrationPeriodBlocks
	DefaultIdentityPricingMaxRegistrationDuration	= DefaultRegistrationPeriodBlocks * 5
	DefaultIdentityPricingMaxRenewalPeriods		= uint32(5)
	DefaultIdentityPricingScarcityShortNameBps	= uint32(15_000)
	DefaultIdentityPricingScarcityMediumNameBps	= uint32(12_500)
	DefaultIdentityPricingScarcityStandardBps	= uint32(10_000)
	DefaultIdentityPricingDemandPremiumBps		= uint32(2_500)
	DefaultIdentityPricingDemandContestedBps	= uint32(5_000)
	DefaultIdentityPricingAuctionSettlementBps	= uint32(500)
	DefaultIdentityPricingMultiPeriodDiscountBps	= uint32(500)
	DefaultIdentityPricingGraceRecoveryMultiplier	= uint32(2_000)
	DefaultIdentityPricingLabelDepthFeeNaet		= int64(1_000_000_000)
	DefaultIdentityPricingDetachedSubdomainFeeNaet	= int64(10_000_000_000)
	DefaultIdentityPricingEphemeralSubdomainFeeNaet	= int64(500_000_000)
)

type IdentityDemandClassV2 string

type IdentitySubdomainModeV2 string

type IdentityPricingParamsV2 struct {
	AntiSquattingParams		IdentityAntiSquattingParamsV2
	BaseDurationBlocks		uint64
	MaxRegistrationDuration		uint64
	MaxRenewalPeriods		uint32
	LabelDepthFee			sdkmath.Int
	ShortNameScarcityBps		uint32
	MediumNameScarcityBps		uint32
	StandardScarcityBps		uint32
	DemandPremiumBps		uint32
	DemandContestedBps		uint32
	DetachedSubdomainFee		sdkmath.Int
	EphemeralSubdomainFee		sdkmath.Int
	AuctionSettlementBps		uint32
	MultiPeriodDiscountEnabled	bool
	MultiPeriodDiscountBps		uint32
	GraceRecoveryMultiplierBps	uint32
	RenewalRequiresWindow		bool
	GraceRecoveryRepairsBindings	bool
}

type IdentityDomainPriceRequestV2 struct {
	Name			string
	DurationBlocks		uint64
	DemandClass		IdentityDemandClassV2
	Renewal			bool
	Auction			bool
	ResolverPayloadBytes	uint64
	SubdomainMode		IdentitySubdomainModeV2
	RenewalPeriods		uint32
	GraceRecovery		bool
}

type IdentityDomainPriceQuoteV2 struct {
	Denom			string
	Name			string
	PricingLabel		string
	NameLength		uint64
	LabelCount		uint64
	DurationBlocks		uint64
	RenewalPeriods		uint32
	DemandClass		IdentityDemandClassV2
	SubdomainMode		IdentitySubdomainModeV2
	BaseRegistrationFee	sdkmath.Int
	ScarcityFee		sdkmath.Int
	DemandFee		sdkmath.Int
	LabelDepthFee		sdkmath.Int
	StorageFootprintFee	sdkmath.Int
	CommitmentDeposit	sdkmath.Int
	RenewalFee		sdkmath.Int
	AuctionSettlementFee	sdkmath.Int
	SubdomainModeFee	sdkmath.Int
	GraceRecoveryFee	sdkmath.Int
	MultiPeriodDiscount	sdkmath.Int
	Total			sdkmath.Int
	ResolverPayloadBytes	uint64
	DeterministicFormula	string
	RenewalCheaperThanFresh	bool
}

func DefaultIdentityPricingParamsV2() IdentityPricingParamsV2 {
	return IdentityPricingParamsV2{
		AntiSquattingParams:		DefaultIdentityAntiSquattingParamsV2(),
		BaseDurationBlocks:		DefaultIdentityPricingBaseDurationBlocks,
		MaxRegistrationDuration:	DefaultIdentityPricingMaxRegistrationDuration,
		MaxRenewalPeriods:		DefaultIdentityPricingMaxRenewalPeriods,
		LabelDepthFee:			sdkmath.NewInt(DefaultIdentityPricingLabelDepthFeeNaet),
		ShortNameScarcityBps:		DefaultIdentityPricingScarcityShortNameBps,
		MediumNameScarcityBps:		DefaultIdentityPricingScarcityMediumNameBps,
		StandardScarcityBps:		DefaultIdentityPricingScarcityStandardBps,
		DemandPremiumBps:		DefaultIdentityPricingDemandPremiumBps,
		DemandContestedBps:		DefaultIdentityPricingDemandContestedBps,
		DetachedSubdomainFee:		sdkmath.NewInt(DefaultIdentityPricingDetachedSubdomainFeeNaet),
		EphemeralSubdomainFee:		sdkmath.NewInt(DefaultIdentityPricingEphemeralSubdomainFeeNaet),
		AuctionSettlementBps:		DefaultIdentityPricingAuctionSettlementBps,
		MultiPeriodDiscountEnabled:	true,
		MultiPeriodDiscountBps:		DefaultIdentityPricingMultiPeriodDiscountBps,
		GraceRecoveryMultiplierBps:	DefaultIdentityPricingGraceRecoveryMultiplier,
		RenewalRequiresWindow:		true,
		GraceRecoveryRepairsBindings:	true,
	}
}

func ValidateIdentityPricingParamsV2(params IdentityPricingParamsV2) error {
	if err := ValidateIdentityAntiSquattingParamsV2(params.AntiSquattingParams); err != nil {
		return err
	}
	if params.BaseDurationBlocks == 0 {
		return errors.New("identity pricing base duration is required")
	}
	if params.MaxRegistrationDuration < params.BaseDurationBlocks {
		return errors.New("identity pricing max registration duration must cover base duration")
	}
	if params.MaxRenewalPeriods == 0 {
		return errors.New("identity pricing max renewal periods is required")
	}
	for _, amount := range []struct {
		label	string
		value	sdkmath.Int
	}{
		{label: "label depth fee", value: params.LabelDepthFee},
		{label: "detached subdomain fee", value: params.DetachedSubdomainFee},
		{label: "ephemeral subdomain fee", value: params.EphemeralSubdomainFee},
	} {
		if amount.value.IsNil() || amount.value.IsNegative() {
			return fmt.Errorf("identity pricing %s must not be negative", amount.label)
		}
	}
	for _, bps := range []struct {
		label	string
		value	uint32
	}{
		{label: "demand premium bps", value: params.DemandPremiumBps},
		{label: "demand contested bps", value: params.DemandContestedBps},
		{label: "auction settlement bps", value: params.AuctionSettlementBps},
		{label: "multi period discount bps", value: params.MultiPeriodDiscountBps},
		{label: "grace recovery multiplier bps", value: params.GraceRecoveryMultiplierBps},
	} {
		if bps.value > DomainDistributionDenominatorBps {
			return fmt.Errorf("identity pricing %s must be <= 10000", bps.label)
		}
	}
	if params.ShortNameScarcityBps < DomainDistributionDenominatorBps ||
		params.MediumNameScarcityBps < DomainDistributionDenominatorBps ||
		params.StandardScarcityBps < DomainDistributionDenominatorBps {
		return errors.New("identity pricing scarcity bps must be at least 10000")
	}
	for _, bps := range []struct {
		label	string
		value	uint32
	}{
		{label: "short name scarcity bps", value: params.ShortNameScarcityBps},
		{label: "medium name scarcity bps", value: params.MediumNameScarcityBps},
		{label: "standard scarcity bps", value: params.StandardScarcityBps},
	} {
		if bps.value > DomainDistributionDenominatorBps*10 {
			return fmt.Errorf("identity pricing %s must be <= 100000", bps.label)
		}
	}
	return nil
}

func QuoteIdentityDomainPriceV2(req IdentityDomainPriceRequestV2, params IdentityPricingParamsV2) (IdentityDomainPriceQuoteV2, error) {
	if err := ValidateIdentityPricingParamsV2(params); err != nil {
		return IdentityDomainPriceQuoteV2{}, err
	}
	normalization, err := NormalizeAETDomainVersioned(req.Name, NameNormalizationVersionV2)
	if err != nil {
		return IdentityDomainPriceQuoteV2{}, err
	}
	pricingLabel, err := IdentityPricingLabelV2(normalization.NormalizedName)
	if err != nil {
		return IdentityDomainPriceQuoteV2{}, err
	}
	duration := req.DurationBlocks
	if duration == 0 {
		duration = params.BaseDurationBlocks
	}
	if duration > params.MaxRegistrationDuration {
		return IdentityDomainPriceQuoteV2{}, errors.New("identity pricing duration exceeds maximum registration duration")
	}
	periods := req.RenewalPeriods
	if periods == 0 {
		periods = 1
	}
	if periods > params.MaxRenewalPeriods {
		return IdentityDomainPriceQuoteV2{}, errors.New("identity pricing renewal periods exceed maximum")
	}
	if req.Renewal {
		duration = params.BaseDurationBlocks * uint64(periods)
		if duration > params.MaxRegistrationDuration {
			return IdentityDomainPriceQuoteV2{}, errors.New("identity pricing renewal duration exceeds maximum")
		}
	}
	baseStart, err := StartPriceForName(pricingLabel, params.AntiSquattingParams.DomainParams)
	if err != nil {
		return IdentityDomainPriceQuoteV2{}, err
	}
	scaledBase := ceilDivIntByUint64V2(baseStart.Mul(sdkmath.NewIntFromUint64(duration)), params.BaseDurationBlocks)
	storageFee := params.AntiSquattingParams.ResolverStorageFeePerByte.Mul(sdkmath.NewIntFromUint64(req.ResolverPayloadBytes))
	labelDepthFee := sdkmath.ZeroInt()
	if len(normalization.Labels) > 1 {
		labelDepthFee = params.LabelDepthFee.Mul(sdkmath.NewInt(int64(len(normalization.Labels) - 1)))
	}
	quote := IdentityDomainPriceQuoteV2{
		Denom:			appparams.BaseDenom,
		Name:			normalization.NormalizedName,
		PricingLabel:		pricingLabel,
		NameLength:		uint64(len(normalization.NormalizedName)),
		LabelCount:		uint64(len(normalization.Labels)),
		DurationBlocks:		duration,
		RenewalPeriods:		periods,
		DemandClass:		normalizeIdentityDemandClassV2(req.DemandClass),
		SubdomainMode:		normalizeIdentitySubdomainModeV2(req.SubdomainMode),
		StorageFootprintFee:	storageFee,
		LabelDepthFee:		labelDepthFee,
		ResolverPayloadBytes:	req.ResolverPayloadBytes,
		DeterministicFormula:	"duration_scaled_length_price + scarcity_fee + demand_fee + label_depth_fee + resolver_payload_bytes*storage_fee_per_byte + optional_commitment_or_renewal_or_auction_or_grace_components",
	}
	if req.Renewal {
		renewalBase, err := RenewalFee(pricingLabel, params.AntiSquattingParams.DomainParams)
		if err != nil {
			return IdentityDomainPriceQuoteV2{}, err
		}
		quote.RenewalFee = renewalBase.Mul(sdkmath.NewInt(int64(periods)))
		if params.MultiPeriodDiscountEnabled && periods > 1 && params.MultiPeriodDiscountBps > 0 {
			quote.MultiPeriodDiscount = ceilBps(quote.RenewalFee, params.MultiPeriodDiscountBps)
			quote.RenewalFee = quote.RenewalFee.Sub(quote.MultiPeriodDiscount)
		}
		quote.Total = quote.RenewalFee.Add(storageFee)
	} else {
		quote.BaseRegistrationFee = scaledBase
		quote.ScarcityFee = scarcityFeeForNameLengthV2(scaledBase, len(pricingLabel), params)
		quote.DemandFee = demandFeeV2(scaledBase, quote.DemandClass, params)
		quote.CommitmentDeposit = params.AntiSquattingParams.CommitmentDeposit
		quote.AuctionSettlementFee = auctionSettlementFeeV2(scaledBase, req.Auction, params)
		quote.SubdomainModeFee = subdomainModeFeeV2(quote.SubdomainMode, params)
		quote.Total = quote.BaseRegistrationFee.
			Add(quote.ScarcityFee).
			Add(quote.DemandFee).
			Add(quote.LabelDepthFee).
			Add(quote.StorageFootprintFee).
			Add(quote.CommitmentDeposit).
			Add(quote.AuctionSettlementFee).
			Add(quote.SubdomainModeFee)
	}
	if req.GraceRecovery {
		recovery := ceilBps(params.AntiSquattingParams.GraceRecoveryCost, params.GraceRecoveryMultiplierBps)
		quote.GraceRecoveryFee = recovery
		quote.Total = quote.Total.Add(recovery)
	}
	if req.Renewal {
		fresh, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{
			Name:			normalization.NormalizedName,
			DurationBlocks:		params.BaseDurationBlocks,
			DemandClass:		quote.DemandClass,
			Auction:		req.Auction,
			ResolverPayloadBytes:	req.ResolverPayloadBytes,
			SubdomainMode:		quote.SubdomainMode,
		}, params.withoutRecursiveGraceDiscountV2())
		if err == nil {
			quote.RenewalCheaperThanFresh = quote.Total.LT(fresh.Total)
		}
	}
	return quote, nil
}

func QuoteIdentityRenewalPriceV2(name string, periods uint32, resolverPayloadBytes uint64, params IdentityPricingParamsV2) (IdentityDomainPriceQuoteV2, error) {
	return QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{
		Name:			name,
		Renewal:		true,
		RenewalPeriods:		periods,
		ResolverPayloadBytes:	resolverPayloadBytes,
	}, params)
}

func RenewIdentityDomainWithEconomicsV2(state IdentityState, name string, actor sdk.AccAddress, height uint64, periods uint32, payment sdkmath.Int, params IdentityPricingParamsV2) (IdentityState, Domain, IdentityDomainPriceQuoteV2, error) {
	if err := ValidateIdentityPricingParamsV2(params); err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	if periods == 0 {
		periods = 1
	}
	if periods > params.MaxRenewalPeriods {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal periods exceed maximum")
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	domain, found := findDomain(state, normalized)
	if !found {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal domain not found")
	}
	lifecycle, err := DomainLifecycle(state, domain.Name, height)
	if err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	if params.RenewalRequiresWindow && lifecycle != DomainLifecycleRenewalWindow {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal requires renewal window")
	}
	if lifecycle == DomainLifecycleExpired {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal requires grace recovery after expiry")
	}
	if !bytes.Equal(actor, domain.Owner) {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal requires owner")
	}
	if _, _, err := ValidateRegistryNFTAuthority(state, domain.Name, height); err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	duration := params.BaseDurationBlocks * uint64(periods)
	if duration > params.MaxRegistrationDuration {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal duration exceeds maximum")
	}
	base := domain.ExpiryHeight
	if height > base {
		base = height
	}
	if base+duration < base {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal final expiry overflows")
	}
	quote, err := QuoteIdentityRenewalPriceV2(domain.Name, periods, 0, params)
	if err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	if payment.IsNil() || payment.LT(quote.Total) {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity renewal payment below quoted price")
	}
	domain.ExpiryHeight = base + duration
	domain.UpdatedHeight = height
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	sortIdentityState(&next)
	return next, domain, quote, next.Validate()
}

func RecoverExpiredIdentityDomainWithEconomicsV2(state IdentityState, name string, actor sdk.AccAddress, height uint64, payment sdkmath.Int, params IdentityPricingParamsV2) (IdentityState, Domain, IdentityDomainPriceQuoteV2, error) {
	if err := ValidateIdentityPricingParamsV2(params); err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	quote, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{Name: name, Renewal: true, GraceRecovery: true}, params)
	if err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	if payment.IsNil() || payment.LT(quote.Total) {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity grace recovery payment below quoted price")
	}
	next, domain, err := RecoverExpiredIdentityDomainV2(state, name, actor, height, payment, params.AntiSquattingParams)
	if err != nil {
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
	}
	if _, _, err := ValidateRegistryNFTAuthority(next, domain.Name, height); err != nil {
		if !params.GraceRecoveryRepairsBindings {
			return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, err
		}
		return IdentityState{}, Domain{}, IdentityDomainPriceQuoteV2{}, errors.New("identity grace recovery cannot repair externally inconsistent nft binding in legacy state")
	}
	return next, domain, quote, nil
}

func RenewalWindowTransitionStatusV2(state IdentityState, name string, height uint64) (DomainLifecycleStatus, error) {
	return DomainLifecycle(state, name, height)
}

func ValidateExpiredDomainResolverSoftFreezeV2(state IdentityState, name string, actor sdk.AccAddress, recordKey string, height uint64) error {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return err
	}
	domain, found := findDomain(state, normalized)
	if !found {
		return errors.New("identity resolver soft-freeze domain not found")
	}
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Expired, 0, height)
	if err != nil {
		return err
	}
	return ValidateResolverUpdateAuthorizationV2(record, actor, nil, recordKey, height)
}

func (params IdentityPricingParamsV2) withoutRecursiveGraceDiscountV2() IdentityPricingParamsV2 {
	params.MultiPeriodDiscountEnabled = false
	return params
}

func normalizeIdentityDemandClassV2(class IdentityDemandClassV2) IdentityDemandClassV2 {
	switch class {
	case "", IdentityDemandClassStandardV2:
		return IdentityDemandClassStandardV2
	case IdentityDemandClassPremiumV2, IdentityDemandClassContestedV2:
		return class
	default:
		return IdentityDemandClassStandardV2
	}
}

func normalizeIdentitySubdomainModeV2(mode IdentitySubdomainModeV2) IdentitySubdomainModeV2 {
	switch mode {
	case "", IdentitySubdomainModeRootV2:
		return IdentitySubdomainModeRootV2
	case IdentitySubdomainModeOwnerControlledV2, IdentitySubdomainModeDetachedPaidV2, IdentitySubdomainModeEphemeralV2:
		return mode
	default:
		return IdentitySubdomainModeRootV2
	}
}

func scarcityFeeForNameLengthV2(base sdkmath.Int, labelLen int, params IdentityPricingParamsV2) sdkmath.Int {
	bps := params.StandardScarcityBps
	switch {
	case labelLen <= 4:
		bps = params.ShortNameScarcityBps
	case labelLen <= 6:
		bps = params.MediumNameScarcityBps
	}
	if bps <= DomainDistributionDenominatorBps {
		return sdkmath.ZeroInt()
	}
	return ceilBps(base, bps-DomainDistributionDenominatorBps)
}

func demandFeeV2(base sdkmath.Int, class IdentityDemandClassV2, params IdentityPricingParamsV2) sdkmath.Int {
	switch class {
	case IdentityDemandClassPremiumV2:
		return ceilBps(base, params.DemandPremiumBps)
	case IdentityDemandClassContestedV2:
		return ceilBps(base, params.DemandContestedBps)
	default:
		return sdkmath.ZeroInt()
	}
}

func auctionSettlementFeeV2(base sdkmath.Int, auction bool, params IdentityPricingParamsV2) sdkmath.Int {
	if !auction || params.AuctionSettlementBps == 0 {
		return sdkmath.ZeroInt()
	}
	return ceilBps(base, params.AuctionSettlementBps)
}

func subdomainModeFeeV2(mode IdentitySubdomainModeV2, params IdentityPricingParamsV2) sdkmath.Int {
	switch mode {
	case IdentitySubdomainModeDetachedPaidV2:
		return params.DetachedSubdomainFee
	case IdentitySubdomainModeEphemeralV2:
		return params.EphemeralSubdomainFee
	default:
		return sdkmath.ZeroInt()
	}
}

func ceilDivIntByUint64V2(amount sdkmath.Int, divisor uint64) sdkmath.Int {
	if divisor == 0 {
		return amount
	}
	d := sdkmath.NewIntFromUint64(divisor)
	return amount.Add(d).SubRaw(1).Quo(d)
}
