package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	IdentityGovernanceInvalidSpoofPatternSetVersionV2 uint64 = 1
	IdentityGovernanceCommitmentTombstoneRetentionV2  uint64 = DefaultCommitTTLBlocks * 30
	IdentityGovernanceExpiryProcessingLimitV2         uint32 = 100
	IdentityGovernanceMinimumRegistrationDurationV2   uint64 = DefaultRegistrationPeriodBlocks / 12
)

type IdentityGovernanceNameParamsV2 struct {
	MaximumLabels                    uint8
	MaximumNameBytes                 uint16
	SupportedNormalizationVersion    uint64
	ReservedLabels                   []string
	InvalidSpoofingPatternSetVersion uint64
	MaximumSubdomainDepth            uint8
}

type IdentityGovernanceLifecycleParamsV2 struct {
	MinimumRegistrationDuration   uint64
	MaximumRegistrationDuration   uint64
	RenewalWindowDuration         uint64
	GracePeriodDuration           uint64
	CommitmentRevealWindow        uint64
	CommitmentTombstoneRetention  uint64
	ExpiryProcessingLimitPerBlock uint32
}

type IdentityGovernancePricingParamsV2 struct {
	Denom                      string
	BaseRegistrationFee        sdkmath.Int
	ShortNameMultiplierBps     uint32
	LabelDepthFee              sdkmath.Int
	RenewalFeeMultiplierBps    uint32
	GraceRecoveryMultiplierBps uint32
	ResolverByteFee            sdkmath.Int
	SubdomainCreationFee       sdkmath.Int
	DetachedSubdomainFee       sdkmath.Int
	AuctionMinimumBid          sdkmath.Int
}

type IdentityGovernanceParamsV2 struct {
	NameParams      IdentityGovernanceNameParamsV2
	LifecycleParams IdentityGovernanceLifecycleParamsV2
	PricingParams   IdentityGovernancePricingParamsV2
	ParamsHash      string
}

func DefaultIdentityGovernanceNameParamsV2() IdentityGovernanceNameParamsV2 {
	return IdentityGovernanceNameParamsV2{
		MaximumLabels:                    MaxDomainLabels,
		MaximumNameBytes:                 MaxDomainFullBytes,
		SupportedNormalizationVersion:    NameNormalizationVersionV2,
		ReservedLabels:                   []string{"admin", "aet", "null", "root", "undefined"},
		InvalidSpoofingPatternSetVersion: IdentityGovernanceInvalidSpoofPatternSetVersionV2,
		MaximumSubdomainDepth:            MaxDomainLabels - 1,
	}
}

func DefaultIdentityGovernanceLifecycleParamsV2() IdentityGovernanceLifecycleParamsV2 {
	identityParams := DefaultIdentityParams()
	pricing := DefaultIdentityPricingParamsV2()
	antiSquatting := pricing.AntiSquattingParams
	return IdentityGovernanceLifecycleParamsV2{
		MinimumRegistrationDuration:   IdentityGovernanceMinimumRegistrationDurationV2,
		MaximumRegistrationDuration:   pricing.MaxRegistrationDuration,
		RenewalWindowDuration:         identityParams.RenewalWindowBlocks,
		GracePeriodDuration:           antiSquatting.ExpiredDomainGracePeriodBlocks,
		CommitmentRevealWindow:        identityParams.CommitTTLBlocks,
		CommitmentTombstoneRetention:  IdentityGovernanceCommitmentTombstoneRetentionV2,
		ExpiryProcessingLimitPerBlock: IdentityGovernanceExpiryProcessingLimitV2,
	}
}

func DefaultIdentityGovernancePricingParamsV2() IdentityGovernancePricingParamsV2 {
	pricing := DefaultIdentityPricingParamsV2()
	antiSquatting := pricing.AntiSquattingParams
	return IdentityGovernancePricingParamsV2{
		Denom:                      "naet",
		BaseRegistrationFee:        antiSquatting.DomainParams.LowStartPrice,
		ShortNameMultiplierBps:     pricing.ShortNameScarcityBps,
		LabelDepthFee:              pricing.LabelDepthFee,
		RenewalFeeMultiplierBps:    antiSquatting.DomainParams.RenewalDiscountBps,
		GraceRecoveryMultiplierBps: pricing.GraceRecoveryMultiplierBps,
		ResolverByteFee:            antiSquatting.ResolverStorageFeePerByte,
		SubdomainCreationFee:       sdkmath.NewInt(DefaultIdentitySpamSubdomainCreationFeeNaet),
		DetachedSubdomainFee:       pricing.DetachedSubdomainFee,
		AuctionMinimumBid:          antiSquatting.DomainParams.LowStartPrice,
	}
}

func DefaultIdentityGovernanceParamsV2() (IdentityGovernanceParamsV2, error) {
	params := IdentityGovernanceParamsV2{
		NameParams:      DefaultIdentityGovernanceNameParamsV2(),
		LifecycleParams: DefaultIdentityGovernanceLifecycleParamsV2(),
		PricingParams:   DefaultIdentityGovernancePricingParamsV2(),
	}
	params.ParamsHash = ComputeIdentityGovernanceParamsHashV2(params)
	return params, ValidateIdentityGovernanceParamsV2(params)
}

func ValidateIdentityGovernanceParamsV2(params IdentityGovernanceParamsV2) error {
	if err := ValidateIdentityGovernanceNameParamsV2(params.NameParams); err != nil {
		return err
	}
	if err := ValidateIdentityGovernanceLifecycleParamsV2(params.LifecycleParams); err != nil {
		return err
	}
	if err := ValidateIdentityGovernancePricingParamsV2(params.PricingParams); err != nil {
		return err
	}
	if params.LifecycleParams.RenewalWindowDuration >= params.LifecycleParams.MaximumRegistrationDuration {
		return errors.New("identity governance renewal window must be below maximum registration duration")
	}
	if params.LifecycleParams.GracePeriodDuration > params.LifecycleParams.MaximumRegistrationDuration {
		return errors.New("identity governance grace period must not exceed maximum registration duration")
	}
	if params.ParamsHash == "" || params.ParamsHash != ComputeIdentityGovernanceParamsHashV2(params) {
		return errors.New("identity governance params hash mismatch")
	}
	return nil
}

func ValidateIdentityGovernanceNameParamsV2(params IdentityGovernanceNameParamsV2) error {
	if params.MaximumLabels == 0 || params.MaximumLabels > MaxDomainLabels {
		return fmt.Errorf("identity governance maximum labels must be in 1..%d", MaxDomainLabels)
	}
	if params.MaximumNameBytes == 0 || params.MaximumNameBytes > MaxDomainFullBytes {
		return fmt.Errorf("identity governance maximum name bytes must be in 1..%d", MaxDomainFullBytes)
	}
	if err := ValidateNameNormalizationVersionV2(params.SupportedNormalizationVersion); err != nil {
		return err
	}
	if params.InvalidSpoofingPatternSetVersion == 0 {
		return errors.New("identity governance invalid spoofing pattern set version is required")
	}
	if params.MaximumSubdomainDepth == 0 || params.MaximumSubdomainDepth > params.MaximumLabels-1 {
		return errors.New("identity governance maximum subdomain depth must be in 1..maximum_labels-1")
	}
	if err := validateGovernanceReservedLabelsV2(params.ReservedLabels); err != nil {
		return err
	}
	return nil
}

func ValidateIdentityGovernanceLifecycleParamsV2(params IdentityGovernanceLifecycleParamsV2) error {
	if params.MinimumRegistrationDuration == 0 {
		return errors.New("identity governance minimum registration duration is required")
	}
	if params.MaximumRegistrationDuration < params.MinimumRegistrationDuration {
		return errors.New("identity governance maximum registration duration must cover minimum")
	}
	if params.RenewalWindowDuration == 0 {
		return errors.New("identity governance renewal window duration is required")
	}
	if params.CommitmentRevealWindow == 0 {
		return errors.New("identity governance commitment reveal window is required")
	}
	if params.CommitmentTombstoneRetention < params.CommitmentRevealWindow {
		return errors.New("identity governance commitment tombstone retention must cover reveal window")
	}
	if params.ExpiryProcessingLimitPerBlock == 0 {
		return errors.New("identity governance expiry processing limit per block is required")
	}
	return nil
}

func ValidateIdentityGovernancePricingParamsV2(params IdentityGovernancePricingParamsV2) error {
	if params.Denom != "naet" {
		return errors.New("identity governance pricing denom must be naet")
	}
	for _, amount := range []struct {
		label string
		value sdkmath.Int
	}{
		{label: "base registration fee", value: params.BaseRegistrationFee},
		{label: "label depth fee", value: params.LabelDepthFee},
		{label: "resolver byte fee", value: params.ResolverByteFee},
		{label: "subdomain creation fee", value: params.SubdomainCreationFee},
		{label: "detached subdomain fee", value: params.DetachedSubdomainFee},
		{label: "auction minimum bid", value: params.AuctionMinimumBid},
	} {
		if amount.value.IsNil() || amount.value.IsNegative() {
			return fmt.Errorf("identity governance pricing %s must not be negative", amount.label)
		}
	}
	for _, bps := range []struct {
		label string
		value uint32
	}{
		{label: "short name multiplier bps", value: params.ShortNameMultiplierBps},
		{label: "renewal fee multiplier bps", value: params.RenewalFeeMultiplierBps},
		{label: "grace recovery multiplier bps", value: params.GraceRecoveryMultiplierBps},
	} {
		if bps.value == 0 || bps.value > DomainDistributionDenominatorBps*10 {
			return fmt.Errorf("identity governance pricing %s must be in 1..100000", bps.label)
		}
	}
	if params.ShortNameMultiplierBps < DomainDistributionDenominatorBps {
		return errors.New("identity governance short-name multiplier must be at least 10000 bps")
	}
	return nil
}

func ApplyIdentityGovernanceParamsToRuntimeV2(params IdentityGovernanceParamsV2) (IdentityParams, IdentityPricingParamsV2, IdentitySpamCostParamsV2, error) {
	if err := ValidateIdentityGovernanceParamsV2(params); err != nil {
		return IdentityParams{}, IdentityPricingParamsV2{}, IdentitySpamCostParamsV2{}, err
	}
	identityParams := DefaultIdentityParams()
	identityParams.RegistrationPeriodBlocks = params.LifecycleParams.MinimumRegistrationDuration
	identityParams.RenewalWindowBlocks = params.LifecycleParams.RenewalWindowDuration
	identityParams.CommitTTLBlocks = params.LifecycleParams.CommitmentRevealWindow

	pricing := DefaultIdentityPricingParamsV2()
	pricing.BaseDurationBlocks = params.LifecycleParams.MinimumRegistrationDuration
	pricing.MaxRegistrationDuration = params.LifecycleParams.MaximumRegistrationDuration
	pricing.LabelDepthFee = params.PricingParams.LabelDepthFee
	pricing.ShortNameScarcityBps = params.PricingParams.ShortNameMultiplierBps
	pricing.GraceRecoveryMultiplierBps = params.PricingParams.GraceRecoveryMultiplierBps
	pricing.DetachedSubdomainFee = params.PricingParams.DetachedSubdomainFee
	pricing.AntiSquattingParams.DomainParams.LowStartPrice = params.PricingParams.BaseRegistrationFee
	pricing.AntiSquattingParams.DomainParams.RenewalDiscountBps = params.PricingParams.RenewalFeeMultiplierBps
	pricing.AntiSquattingParams.ResolverStorageFeePerByte = params.PricingParams.ResolverByteFee
	pricing.AntiSquattingParams.ExpiredDomainGracePeriodBlocks = params.LifecycleParams.GracePeriodDuration
	pricing.AntiSquattingParams.ExpiredDomainReleaseWindowBlocks = params.LifecycleParams.GracePeriodDuration + 1
	pricing.AntiSquattingParams.DomainParams.HighStartPrice = params.PricingParams.AuctionMinimumBid

	spam := DefaultIdentitySpamCostParamsV2()
	spam.PricingParams = pricing
	spam.SubdomainCreationFee = params.PricingParams.SubdomainCreationFee
	return identityParams, pricing, spam, nil
}

func ComputeIdentityGovernanceParamsHashV2(params IdentityGovernanceParamsV2) string {
	name := params.NameParams
	lifecycle := params.LifecycleParams
	pricing := params.PricingParams
	parts := []string{
		"identity-governance-params-v2",
		fmt.Sprint(name.MaximumLabels),
		fmt.Sprint(name.MaximumNameBytes),
		fmt.Sprint(name.SupportedNormalizationVersion),
		fmt.Sprint(name.InvalidSpoofingPatternSetVersion),
		fmt.Sprint(name.MaximumSubdomainDepth),
		fmt.Sprint(lifecycle.MinimumRegistrationDuration),
		fmt.Sprint(lifecycle.MaximumRegistrationDuration),
		fmt.Sprint(lifecycle.RenewalWindowDuration),
		fmt.Sprint(lifecycle.GracePeriodDuration),
		fmt.Sprint(lifecycle.CommitmentRevealWindow),
		fmt.Sprint(lifecycle.CommitmentTombstoneRetention),
		fmt.Sprint(lifecycle.ExpiryProcessingLimitPerBlock),
		pricing.Denom,
		pricing.BaseRegistrationFee.String(),
		fmt.Sprint(pricing.ShortNameMultiplierBps),
		pricing.LabelDepthFee.String(),
		fmt.Sprint(pricing.RenewalFeeMultiplierBps),
		fmt.Sprint(pricing.GraceRecoveryMultiplierBps),
		pricing.ResolverByteFee.String(),
		pricing.SubdomainCreationFee.String(),
		pricing.DetachedSubdomainFee.String(),
		pricing.AuctionMinimumBid.String(),
	}
	parts = append(parts, sortedBreakdownStringsV2(name.ReservedLabels)...)
	return identityHash(parts...)
}

func validateGovernanceReservedLabelsV2(labels []string) error {
	if len(labels) == 0 {
		return errors.New("identity governance reserved labels are required")
	}
	ordered := append([]string(nil), labels...)
	sort.Strings(ordered)
	for i, label := range ordered {
		if label == "" {
			return errors.New("identity governance reserved label is required")
		}
		if label != labels[i] {
			return errors.New("identity governance reserved labels must be sorted")
		}
		if err := validateDomainLabel(label); err != nil {
			return err
		}
		if i > 0 && ordered[i-1] == label {
			return fmt.Errorf("duplicate identity governance reserved label %q", label)
		}
	}
	for _, required := range []string{"admin", "aet", "root"} {
		found := false
		for _, label := range labels {
			if label == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("identity governance reserved labels missing %q", required)
		}
	}
	return nil
}
