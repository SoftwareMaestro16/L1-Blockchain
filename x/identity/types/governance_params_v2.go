package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	IdentityGovernanceInvalidSpoofPatternSetVersionV2	uint64	= 1
	IdentityGovernanceCommitmentTombstoneRetentionV2	uint64	= DefaultCommitTTLBlocks * 30
	IdentityGovernanceExpiryProcessingLimitV2		uint32	= 100
	IdentityGovernanceMinimumRegistrationDurationV2		uint64	= DefaultRegistrationPeriodBlocks / 12
	IdentityGovernanceMaximumResolverTTLV2			uint64	= DefaultRegistrationPeriodBlocks
	IdentityGovernanceMaximumDelegationDurationV2		uint64	= DefaultRegistrationPeriodBlocks
	IdentityGovernanceMaximumScopedDelegatesPerDomain	uint32	= 64
	IdentityGovernanceMaximumZonePolicyBytesV2		uint64	= 4 * 1024
	IdentityGovernanceAuctionFinalizationDelayV2		uint64	= 1
	IdentityGovernanceCacheRecordMaximumLifetimeV2		uint64	= DefaultRegistrationPeriodBlocks
	IdentityGovernanceStorePruningHorizonV2			uint64	= DefaultRegistrationPeriodBlocks
)

type IdentityGovernanceNameParamsV2 struct {
	MaximumLabels				uint8
	MaximumNameBytes			uint16
	SupportedNormalizationVersion		uint64
	ReservedLabels				[]string
	InvalidSpoofingPatternSetVersion	uint64
	MaximumSubdomainDepth			uint8
}

type IdentityGovernanceLifecycleParamsV2 struct {
	MinimumRegistrationDuration	uint64
	MaximumRegistrationDuration	uint64
	RenewalWindowDuration		uint64
	GracePeriodDuration		uint64
	CommitmentRevealWindow		uint64
	CommitmentTombstoneRetention	uint64
	ExpiryProcessingLimitPerBlock	uint32
}

type IdentityGovernancePricingParamsV2 struct {
	Denom				string
	BaseRegistrationFee		sdkmath.Int
	ShortNameMultiplierBps		uint32
	LabelDepthFee			sdkmath.Int
	RenewalFeeMultiplierBps		uint32
	GraceRecoveryMultiplierBps	uint32
	ResolverByteFee			sdkmath.Int
	SubdomainCreationFee		sdkmath.Int
	DetachedSubdomainFee		sdkmath.Int
	AuctionMinimumBid		sdkmath.Int
}

type IdentityGovernanceResolverParamsV2 struct {
	MaximumResolverRecordBytes	uint64
	MaximumContractTargets		uint32
	MaximumServiceEndpoints		uint32
	MaximumInterfaceDescriptors	uint32
	MaximumRoutingMetadataBytes	uint64
	MaximumInlineSchemaBytes	uint64
	MinimumResolverTTL		uint64
	MaximumResolverTTL		uint64
	AllowedEndpointSchemes		[]string
}

type IdentityGovernanceDelegationParamsV2 struct {
	MaximumDelegationDuration	uint64
	MaximumScopedDelegatesPerDomain	uint32
	MaximumZonePolicySizeBytes	uint64
	DetachedSubdomainAllowed	bool
	TimeLockedDelegationAllowed	bool
}

type IdentityGovernanceAuctionParamsV2 struct {
	CommitPhaseDuration		uint64
	RevealPhaseDuration		uint64
	BidDepositMinimum		sdkmath.Int
	UnrevealedBidPenaltyBps		uint32
	TieBreakRule			IdentityAuctionTieBreakRuleV2
	AuctionFinalizationDelay	uint64
	FeeBurnBps			uint32
	FeeTreasuryBps			uint32
	FeeRewardsBps			uint32
	FeeCommunityPoolBps		uint32
}

type IdentityGovernancePerformanceParamsV2 struct {
	BatchResolverUpdateMaximumSize		uint32
	BatchRenewalMaximumSize			uint32
	RecursiveProofMaximumDepth		uint8
	CacheRecordMaximumLifetime		uint64
	StorePruningHorizonForProofAvailability	uint64
	ABCIExpiryWorkLimit			uint32
}

type IdentityGovernanceParamsV2 struct {
	NameParams		IdentityGovernanceNameParamsV2
	LifecycleParams		IdentityGovernanceLifecycleParamsV2
	PricingParams		IdentityGovernancePricingParamsV2
	ResolverParams		IdentityGovernanceResolverParamsV2
	DelegationParams	IdentityGovernanceDelegationParamsV2
	AuctionParams		IdentityGovernanceAuctionParamsV2
	PerformanceParams	IdentityGovernancePerformanceParamsV2
	ParamsHash		string
}

func DefaultIdentityGovernanceNameParamsV2() IdentityGovernanceNameParamsV2 {
	return IdentityGovernanceNameParamsV2{
		MaximumLabels:				MaxDomainLabels,
		MaximumNameBytes:			MaxDomainFullBytes,
		SupportedNormalizationVersion:		NameNormalizationVersionV2,
		ReservedLabels:				[]string{"admin", "aet", "null", "root", "undefined"},
		InvalidSpoofingPatternSetVersion:	IdentityGovernanceInvalidSpoofPatternSetVersionV2,
		MaximumSubdomainDepth:			MaxDomainLabels - 1,
	}
}

func DefaultIdentityGovernanceLifecycleParamsV2() IdentityGovernanceLifecycleParamsV2 {
	identityParams := DefaultIdentityParams()
	pricing := DefaultIdentityPricingParamsV2()
	antiSquatting := pricing.AntiSquattingParams
	return IdentityGovernanceLifecycleParamsV2{
		MinimumRegistrationDuration:	IdentityGovernanceMinimumRegistrationDurationV2,
		MaximumRegistrationDuration:	pricing.MaxRegistrationDuration,
		RenewalWindowDuration:		identityParams.RenewalWindowBlocks,
		GracePeriodDuration:		antiSquatting.ExpiredDomainGracePeriodBlocks,
		CommitmentRevealWindow:		identityParams.CommitTTLBlocks,
		CommitmentTombstoneRetention:	IdentityGovernanceCommitmentTombstoneRetentionV2,
		ExpiryProcessingLimitPerBlock:	IdentityGovernanceExpiryProcessingLimitV2,
	}
}

func DefaultIdentityGovernancePricingParamsV2() IdentityGovernancePricingParamsV2 {
	pricing := DefaultIdentityPricingParamsV2()
	antiSquatting := pricing.AntiSquattingParams
	return IdentityGovernancePricingParamsV2{
		Denom:				"naet",
		BaseRegistrationFee:		antiSquatting.DomainParams.LowStartPrice,
		ShortNameMultiplierBps:		pricing.ShortNameScarcityBps,
		LabelDepthFee:			pricing.LabelDepthFee,
		RenewalFeeMultiplierBps:	antiSquatting.DomainParams.RenewalDiscountBps,
		GraceRecoveryMultiplierBps:	pricing.GraceRecoveryMultiplierBps,
		ResolverByteFee:		antiSquatting.ResolverStorageFeePerByte,
		SubdomainCreationFee:		sdkmath.NewInt(DefaultIdentitySpamSubdomainCreationFeeNaet),
		DetachedSubdomainFee:		pricing.DetachedSubdomainFee,
		AuctionMinimumBid:		antiSquatting.DomainParams.LowStartPrice,
	}
}

func DefaultIdentityGovernanceResolverParamsV2() IdentityGovernanceResolverParamsV2 {
	return IdentityGovernanceResolverParamsV2{
		MaximumResolverRecordBytes:	MaxUnifiedPayloadBytesV2,
		MaximumContractTargets:		MaxUnifiedContractTargets,
		MaximumServiceEndpoints:	MaxUnifiedServiceEndpoints,
		MaximumInterfaceDescriptors:	MaxUnifiedInterfaceDescriptors,
		MaximumRoutingMetadataBytes:	MaxUnifiedRoutingMetadataBytes,
		MaximumInlineSchemaBytes:	MaxInterfaceInlineSchemaBytesV2,
		MinimumResolverTTL:		1,
		MaximumResolverTTL:		IdentityGovernanceMaximumResolverTTLV2,
		AllowedEndpointSchemes:		[]string{"aetra", "grpcs", "https", "ipfs", "wss"},
	}
}

func DefaultIdentityGovernanceDelegationParamsV2() IdentityGovernanceDelegationParamsV2 {
	return IdentityGovernanceDelegationParamsV2{
		MaximumDelegationDuration:		IdentityGovernanceMaximumDelegationDurationV2,
		MaximumScopedDelegatesPerDomain:	IdentityGovernanceMaximumScopedDelegatesPerDomain,
		MaximumZonePolicySizeBytes:		IdentityGovernanceMaximumZonePolicyBytesV2,
		DetachedSubdomainAllowed:		true,
		TimeLockedDelegationAllowed:		true,
	}
}

func DefaultIdentityGovernanceAuctionParamsV2() IdentityGovernanceAuctionParamsV2 {
	fairness := DefaultIdentityAuctionFairnessParamsV2("aetra-local-1")
	return IdentityGovernanceAuctionParamsV2{
		CommitPhaseDuration:		DefaultAuctionCommitBlocks,
		RevealPhaseDuration:		DefaultAuctionRevealBlocks,
		BidDepositMinimum:		fairness.BidDeposit,
		UnrevealedBidPenaltyBps:	fairness.UnrevealedForfeitBps,
		TieBreakRule:			fairness.TieBreakRule,
		AuctionFinalizationDelay:	IdentityGovernanceAuctionFinalizationDelayV2,
		FeeBurnBps:			fairness.FeeBurnBps,
		FeeTreasuryBps:			fairness.FeeTreasuryBps,
		FeeRewardsBps:			fairness.FeeRewardsBps,
		FeeCommunityPoolBps:		fairness.FeeCommunityPoolBps,
	}
}

func DefaultIdentityGovernancePerformanceParamsV2() IdentityGovernancePerformanceParamsV2 {
	return IdentityGovernancePerformanceParamsV2{
		BatchResolverUpdateMaximumSize:			MaxIdentityTxBatchResolverUpdatesV2,
		BatchRenewalMaximumSize:			MaxIdentityTxBatchRenewDomainsV2,
		RecursiveProofMaximumDepth:			MaxDomainLabels,
		CacheRecordMaximumLifetime:			IdentityGovernanceCacheRecordMaximumLifetimeV2,
		StorePruningHorizonForProofAvailability:	IdentityGovernanceStorePruningHorizonV2,
		ABCIExpiryWorkLimit:				IdentityGovernanceExpiryProcessingLimitV2,
	}
}

func DefaultIdentityGovernanceParamsV2() (IdentityGovernanceParamsV2, error) {
	params := IdentityGovernanceParamsV2{
		NameParams:		DefaultIdentityGovernanceNameParamsV2(),
		LifecycleParams:	DefaultIdentityGovernanceLifecycleParamsV2(),
		PricingParams:		DefaultIdentityGovernancePricingParamsV2(),
		ResolverParams:		DefaultIdentityGovernanceResolverParamsV2(),
		DelegationParams:	DefaultIdentityGovernanceDelegationParamsV2(),
		AuctionParams:		DefaultIdentityGovernanceAuctionParamsV2(),
		PerformanceParams:	DefaultIdentityGovernancePerformanceParamsV2(),
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
	if err := ValidateIdentityGovernanceResolverParamsV2(params.ResolverParams); err != nil {
		return err
	}
	if err := ValidateIdentityGovernanceDelegationParamsV2(params.DelegationParams); err != nil {
		return err
	}
	if err := ValidateIdentityGovernanceAuctionParamsV2(params.AuctionParams); err != nil {
		return err
	}
	if err := ValidateIdentityGovernancePerformanceParamsV2(params.PerformanceParams); err != nil {
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
		label	string
		value	sdkmath.Int
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
		label	string
		value	uint32
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

func ValidateIdentityGovernanceResolverParamsV2(params IdentityGovernanceResolverParamsV2) error {
	if params.MaximumResolverRecordBytes == 0 || params.MaximumResolverRecordBytes > MaxUnifiedPayloadBytesV2 {
		return fmt.Errorf("identity governance resolver maximum record bytes must be in 1..%d", MaxUnifiedPayloadBytesV2)
	}
	if params.MaximumContractTargets == 0 || params.MaximumContractTargets > MaxUnifiedContractTargets {
		return fmt.Errorf("identity governance resolver maximum contract targets must be in 1..%d", MaxUnifiedContractTargets)
	}
	if params.MaximumServiceEndpoints == 0 || params.MaximumServiceEndpoints > MaxUnifiedServiceEndpoints {
		return fmt.Errorf("identity governance resolver maximum service endpoints must be in 1..%d", MaxUnifiedServiceEndpoints)
	}
	if params.MaximumInterfaceDescriptors == 0 || params.MaximumInterfaceDescriptors > MaxUnifiedInterfaceDescriptors {
		return fmt.Errorf("identity governance resolver maximum interface descriptors must be in 1..%d", MaxUnifiedInterfaceDescriptors)
	}
	if params.MaximumRoutingMetadataBytes == 0 || params.MaximumRoutingMetadataBytes > MaxUnifiedRoutingMetadataBytes {
		return fmt.Errorf("identity governance resolver maximum routing metadata bytes must be in 1..%d", MaxUnifiedRoutingMetadataBytes)
	}
	if params.MaximumInlineSchemaBytes == 0 || params.MaximumInlineSchemaBytes > MaxInterfaceInlineSchemaBytesV2 {
		return fmt.Errorf("identity governance resolver maximum inline schema bytes must be in 1..%d", MaxInterfaceInlineSchemaBytesV2)
	}
	if params.MinimumResolverTTL == 0 {
		return errors.New("identity governance resolver minimum ttl is required")
	}
	if params.MaximumResolverTTL < params.MinimumResolverTTL {
		return errors.New("identity governance resolver maximum ttl must cover minimum ttl")
	}
	return validateGovernanceEndpointSchemesV2(params.AllowedEndpointSchemes)
}

func ValidateIdentityGovernanceDelegationParamsV2(params IdentityGovernanceDelegationParamsV2) error {
	if params.MaximumDelegationDuration == 0 {
		return errors.New("identity governance delegation maximum duration is required")
	}
	if params.MaximumScopedDelegatesPerDomain == 0 {
		return errors.New("identity governance maximum scoped delegates per domain is required")
	}
	if params.MaximumZonePolicySizeBytes == 0 {
		return errors.New("identity governance maximum zone policy size is required")
	}
	return nil
}

func ValidateIdentityGovernanceAuctionParamsV2(params IdentityGovernanceAuctionParamsV2) error {
	if params.CommitPhaseDuration == 0 {
		return errors.New("identity governance auction commit phase duration is required")
	}
	if params.RevealPhaseDuration == 0 {
		return errors.New("identity governance auction reveal phase duration is required")
	}
	if params.BidDepositMinimum.IsNil() || params.BidDepositMinimum.IsNegative() {
		return errors.New("identity governance auction bid deposit minimum must not be negative")
	}
	if params.UnrevealedBidPenaltyBps > DomainDistributionDenominatorBps {
		return errors.New("identity governance auction unrevealed bid penalty must be <= 10000")
	}
	switch params.TieBreakRule {
	case IdentityAuctionTieBreakEarliestRevealThenCommitmentV2, IdentityAuctionTieBreakCommitmentHashV2:
	default:
		return fmt.Errorf("unsupported identity governance auction tie-break rule %q", params.TieBreakRule)
	}
	if params.AuctionFinalizationDelay == 0 {
		return errors.New("identity governance auction finalization delay is required")
	}
	if params.FeeBurnBps+params.FeeTreasuryBps+params.FeeRewardsBps+params.FeeCommunityPoolBps != DomainDistributionDenominatorBps {
		return errors.New("identity governance auction fee split weights must sum to 10000")
	}
	return nil
}

func ValidateIdentityGovernancePerformanceParamsV2(params IdentityGovernancePerformanceParamsV2) error {
	if params.BatchResolverUpdateMaximumSize == 0 || params.BatchResolverUpdateMaximumSize > MaxIdentityTxBatchResolverUpdatesV2 {
		return fmt.Errorf("identity governance performance batch resolver update maximum size must be in 1..%d", MaxIdentityTxBatchResolverUpdatesV2)
	}
	if params.BatchRenewalMaximumSize == 0 || params.BatchRenewalMaximumSize > MaxIdentityTxBatchRenewDomainsV2 {
		return fmt.Errorf("identity governance performance batch renewal maximum size must be in 1..%d", MaxIdentityTxBatchRenewDomainsV2)
	}
	if params.RecursiveProofMaximumDepth == 0 || params.RecursiveProofMaximumDepth > MaxDomainLabels {
		return fmt.Errorf("identity governance performance recursive proof maximum depth must be in 1..%d", MaxDomainLabels)
	}
	if params.CacheRecordMaximumLifetime == 0 {
		return errors.New("identity governance performance cache record maximum lifetime is required")
	}
	if params.StorePruningHorizonForProofAvailability == 0 {
		return errors.New("identity governance performance store pruning horizon for proof availability is required")
	}
	if params.StorePruningHorizonForProofAvailability < params.CacheRecordMaximumLifetime {
		return errors.New("identity governance performance store pruning horizon must cover cache record maximum lifetime")
	}
	if params.ABCIExpiryWorkLimit == 0 {
		return errors.New("identity governance performance ABCI++ expiry work limit is required")
	}
	return nil
}

func IdentityGovernanceValidateBatchResolverUpdateCountV2(params IdentityGovernanceParamsV2, count int) error {
	if err := ValidateIdentityGovernanceParamsV2(params); err != nil {
		return err
	}
	if count < 0 || count > int(params.PerformanceParams.BatchResolverUpdateMaximumSize) {
		return fmt.Errorf("identity governance resolver batch update count must be in 0..%d", params.PerformanceParams.BatchResolverUpdateMaximumSize)
	}
	return nil
}

func IdentityGovernanceValidateBatchRenewalCountV2(params IdentityGovernanceParamsV2, count int) error {
	if err := ValidateIdentityGovernanceParamsV2(params); err != nil {
		return err
	}
	if count < 0 || count > int(params.PerformanceParams.BatchRenewalMaximumSize) {
		return fmt.Errorf("identity governance batch renewal count must be in 0..%d", params.PerformanceParams.BatchRenewalMaximumSize)
	}
	return nil
}

func ApplyIdentityGovernanceParamsToRuntimeV2(params IdentityGovernanceParamsV2) (IdentityParams, IdentityPricingParamsV2, IdentitySpamCostParamsV2, IdentityAuctionFairnessParamsV2, error) {
	if err := ValidateIdentityGovernanceParamsV2(params); err != nil {
		return IdentityParams{}, IdentityPricingParamsV2{}, IdentitySpamCostParamsV2{}, IdentityAuctionFairnessParamsV2{}, err
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

	auction := DefaultIdentityAuctionFairnessParamsV2("aetra-local-1")
	auction.BidDeposit = params.AuctionParams.BidDepositMinimum
	auction.UnrevealedForfeitBps = params.AuctionParams.UnrevealedBidPenaltyBps
	auction.TieBreakRule = params.AuctionParams.TieBreakRule
	auction.FeeBurnBps = params.AuctionParams.FeeBurnBps
	auction.FeeTreasuryBps = params.AuctionParams.FeeTreasuryBps
	auction.FeeRewardsBps = params.AuctionParams.FeeRewardsBps
	auction.FeeCommunityPoolBps = params.AuctionParams.FeeCommunityPoolBps
	return identityParams, pricing, spam, auction, nil
}

func ComputeIdentityGovernanceParamsHashV2(params IdentityGovernanceParamsV2) string {
	name := params.NameParams
	lifecycle := params.LifecycleParams
	pricing := params.PricingParams
	resolver := params.ResolverParams
	delegation := params.DelegationParams
	auction := params.AuctionParams
	performance := params.PerformanceParams
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
		fmt.Sprint(resolver.MaximumResolverRecordBytes),
		fmt.Sprint(resolver.MaximumContractTargets),
		fmt.Sprint(resolver.MaximumServiceEndpoints),
		fmt.Sprint(resolver.MaximumInterfaceDescriptors),
		fmt.Sprint(resolver.MaximumRoutingMetadataBytes),
		fmt.Sprint(resolver.MaximumInlineSchemaBytes),
		fmt.Sprint(resolver.MinimumResolverTTL),
		fmt.Sprint(resolver.MaximumResolverTTL),
		fmt.Sprint(delegation.MaximumDelegationDuration),
		fmt.Sprint(delegation.MaximumScopedDelegatesPerDomain),
		fmt.Sprint(delegation.MaximumZonePolicySizeBytes),
		fmt.Sprint(delegation.DetachedSubdomainAllowed),
		fmt.Sprint(delegation.TimeLockedDelegationAllowed),
		fmt.Sprint(auction.CommitPhaseDuration),
		fmt.Sprint(auction.RevealPhaseDuration),
		auction.BidDepositMinimum.String(),
		fmt.Sprint(auction.UnrevealedBidPenaltyBps),
		string(auction.TieBreakRule),
		fmt.Sprint(auction.AuctionFinalizationDelay),
		fmt.Sprint(auction.FeeBurnBps),
		fmt.Sprint(auction.FeeTreasuryBps),
		fmt.Sprint(auction.FeeRewardsBps),
		fmt.Sprint(auction.FeeCommunityPoolBps),
		fmt.Sprint(performance.BatchResolverUpdateMaximumSize),
		fmt.Sprint(performance.BatchRenewalMaximumSize),
		fmt.Sprint(performance.RecursiveProofMaximumDepth),
		fmt.Sprint(performance.CacheRecordMaximumLifetime),
		fmt.Sprint(performance.StorePruningHorizonForProofAvailability),
		fmt.Sprint(performance.ABCIExpiryWorkLimit),
	}
	parts = append(parts, sortedBreakdownStringsV2(name.ReservedLabels)...)
	parts = append(parts, sortedBreakdownStringsV2(resolver.AllowedEndpointSchemes)...)
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

func validateGovernanceEndpointSchemesV2(schemes []string) error {
	if len(schemes) == 0 {
		return errors.New("identity governance resolver allowed endpoint schemes are required")
	}
	ordered := append([]string(nil), schemes...)
	sort.Strings(ordered)
	for i, scheme := range ordered {
		if scheme == "" {
			return errors.New("identity governance resolver allowed endpoint scheme is required")
		}
		if scheme != schemes[i] {
			return errors.New("identity governance resolver allowed endpoint schemes must be sorted")
		}
		if i > 0 && ordered[i-1] == scheme {
			return fmt.Errorf("duplicate identity governance endpoint scheme %q", scheme)
		}
		switch scheme {
		case "aetra", "grpcs", "https", "ipfs", "wss":
		default:
			return fmt.Errorf("unsupported identity governance endpoint scheme %q", scheme)
		}
	}
	return nil
}
