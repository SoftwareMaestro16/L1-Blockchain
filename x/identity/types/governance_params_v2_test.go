package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestIdentityGovernanceParamsV2DefaultsCoverSection16(t *testing.T) {
	params, err := DefaultIdentityGovernanceParamsV2()
	require.NoError(t, err)
	require.NoError(t, ValidateIdentityGovernanceParamsV2(params))

	require.Equal(t, uint8(MaxDomainLabels), params.NameParams.MaximumLabels)
	require.Equal(t, uint16(MaxDomainFullBytes), params.NameParams.MaximumNameBytes)
	require.Equal(t, NameNormalizationVersionV2, params.NameParams.SupportedNormalizationVersion)
	require.Contains(t, params.NameParams.ReservedLabels, "admin")
	require.Contains(t, params.NameParams.ReservedLabels, "aet")
	require.Contains(t, params.NameParams.ReservedLabels, "root")
	require.Equal(t, IdentityGovernanceInvalidSpoofPatternSetVersionV2, params.NameParams.InvalidSpoofingPatternSetVersion)
	require.Equal(t, uint8(MaxDomainLabels-1), params.NameParams.MaximumSubdomainDepth)

	require.Equal(t, IdentityGovernanceMinimumRegistrationDurationV2, params.LifecycleParams.MinimumRegistrationDuration)
	require.Equal(t, DefaultIdentityPricingMaxRegistrationDuration, params.LifecycleParams.MaximumRegistrationDuration)
	require.Equal(t, DefaultRenewalWindowBlocks, params.LifecycleParams.RenewalWindowDuration)
	require.Equal(t, DefaultRenewalWindowBlocks, params.LifecycleParams.GracePeriodDuration)
	require.Equal(t, DefaultCommitTTLBlocks, params.LifecycleParams.CommitmentRevealWindow)
	require.Equal(t, IdentityGovernanceCommitmentTombstoneRetentionV2, params.LifecycleParams.CommitmentTombstoneRetention)
	require.Equal(t, IdentityGovernanceExpiryProcessingLimitV2, params.LifecycleParams.ExpiryProcessingLimitPerBlock)

	require.Equal(t, "naet", params.PricingParams.Denom)
	require.Equal(t, DefaultDomainParams().LowStartPrice, params.PricingParams.BaseRegistrationFee)
	require.Equal(t, DefaultIdentityPricingScarcityShortNameBps, params.PricingParams.ShortNameMultiplierBps)
	require.Equal(t, sdkmath.NewInt(DefaultIdentityPricingLabelDepthFeeNaet), params.PricingParams.LabelDepthFee)
	require.Equal(t, DomainRenewalDiscountBps, params.PricingParams.RenewalFeeMultiplierBps)
	require.Equal(t, DefaultIdentityPricingGraceRecoveryMultiplier, params.PricingParams.GraceRecoveryMultiplierBps)
	require.Equal(t, sdkmath.NewInt(DefaultIdentityResolverStorageFeePerByte), params.PricingParams.ResolverByteFee)
	require.Equal(t, sdkmath.NewInt(DefaultIdentitySpamSubdomainCreationFeeNaet), params.PricingParams.SubdomainCreationFee)
	require.Equal(t, sdkmath.NewInt(DefaultIdentityPricingDetachedSubdomainFeeNaet), params.PricingParams.DetachedSubdomainFee)
	require.Equal(t, DefaultDomainParams().LowStartPrice, params.PricingParams.AuctionMinimumBid)
	require.NotEmpty(t, params.ParamsHash)
}

func TestIdentityGovernanceParamsV2ApplyToRuntime(t *testing.T) {
	params, err := DefaultIdentityGovernanceParamsV2()
	require.NoError(t, err)
	params.LifecycleParams.MinimumRegistrationDuration = 1_000
	params.LifecycleParams.MaximumRegistrationDuration = 5_000
	params.LifecycleParams.RenewalWindowDuration = 100
	params.LifecycleParams.GracePeriodDuration = 50
	params.LifecycleParams.CommitmentRevealWindow = 25
	params.PricingParams.BaseRegistrationFee = sdkmath.NewInt(123)
	params.PricingParams.ShortNameMultiplierBps = 12_000
	params.PricingParams.LabelDepthFee = sdkmath.NewInt(7)
	params.PricingParams.RenewalFeeMultiplierBps = 800
	params.PricingParams.GraceRecoveryMultiplierBps = 1_500
	params.PricingParams.ResolverByteFee = sdkmath.NewInt(3)
	params.PricingParams.SubdomainCreationFee = sdkmath.NewInt(9)
	params.PricingParams.DetachedSubdomainFee = sdkmath.NewInt(11)
	params.PricingParams.AuctionMinimumBid = sdkmath.NewInt(456)
	params.ParamsHash = ComputeIdentityGovernanceParamsHashV2(params)

	identityParams, pricing, spam, err := ApplyIdentityGovernanceParamsToRuntimeV2(params)
	require.NoError(t, err)
	require.Equal(t, uint64(1_000), identityParams.RegistrationPeriodBlocks)
	require.Equal(t, uint64(100), identityParams.RenewalWindowBlocks)
	require.Equal(t, uint64(25), identityParams.CommitTTLBlocks)
	require.Equal(t, uint64(1_000), pricing.BaseDurationBlocks)
	require.Equal(t, uint64(5_000), pricing.MaxRegistrationDuration)
	require.Equal(t, sdkmath.NewInt(123), pricing.AntiSquattingParams.DomainParams.LowStartPrice)
	require.Equal(t, uint32(12_000), pricing.ShortNameScarcityBps)
	require.Equal(t, sdkmath.NewInt(7), pricing.LabelDepthFee)
	require.Equal(t, uint32(800), pricing.AntiSquattingParams.DomainParams.RenewalDiscountBps)
	require.Equal(t, uint32(1_500), pricing.GraceRecoveryMultiplierBps)
	require.Equal(t, sdkmath.NewInt(3), pricing.AntiSquattingParams.ResolverStorageFeePerByte)
	require.Equal(t, uint64(50), pricing.AntiSquattingParams.ExpiredDomainGracePeriodBlocks)
	require.Equal(t, uint64(51), pricing.AntiSquattingParams.ExpiredDomainReleaseWindowBlocks)
	require.Equal(t, sdkmath.NewInt(456), pricing.AntiSquattingParams.DomainParams.HighStartPrice)
	require.Equal(t, pricing, spam.PricingParams)
	require.Equal(t, sdkmath.NewInt(9), spam.SubdomainCreationFee)
}

func TestIdentityGovernanceParamsV2RejectsInvalidNameLifecyclePricing(t *testing.T) {
	params, err := DefaultIdentityGovernanceParamsV2()
	require.NoError(t, err)

	badName := params
	badName.NameParams.MaximumLabels = MaxDomainLabels + 1
	badName.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badName)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badName), "maximum labels")

	badReserved := params
	badReserved.NameParams.ReservedLabels = []string{"aet", "admin", "root"}
	badReserved.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badReserved)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badReserved), "reserved labels must be sorted")

	badNorm := params
	badNorm.NameParams.SupportedNormalizationVersion = NameNormalizationVersionV2 + 1
	badNorm.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badNorm)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badNorm), "unsupported")

	badLifecycle := params
	badLifecycle.LifecycleParams.MaximumRegistrationDuration = badLifecycle.LifecycleParams.MinimumRegistrationDuration - 1
	badLifecycle.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badLifecycle)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badLifecycle), "maximum registration duration")

	badTombstone := params
	badTombstone.LifecycleParams.CommitmentTombstoneRetention = badTombstone.LifecycleParams.CommitmentRevealWindow - 1
	badTombstone.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badTombstone)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badTombstone), "tombstone retention")

	badDenom := params
	badDenom.PricingParams.Denom = "uatom"
	badDenom.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badDenom)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badDenom), "denom must be naet")

	badFee := params
	badFee.PricingParams.ResolverByteFee = sdkmath.NewInt(-1)
	badFee.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badFee)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badFee), "resolver byte fee")

	badMultiplier := params
	badMultiplier.PricingParams.ShortNameMultiplierBps = DomainDistributionDenominatorBps - 1
	badMultiplier.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badMultiplier)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badMultiplier), "short-name multiplier")

	tampered := params
	tampered.PricingParams.SubdomainCreationFee = sdkmath.NewInt(99)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(tampered), "hash mismatch")
}
