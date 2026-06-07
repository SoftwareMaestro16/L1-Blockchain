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
	require.Equal(t, uint64(MaxUnifiedPayloadBytesV2), params.ResolverParams.MaximumResolverRecordBytes)
	require.Equal(t, uint32(MaxUnifiedContractTargets), params.ResolverParams.MaximumContractTargets)
	require.Equal(t, uint32(MaxUnifiedServiceEndpoints), params.ResolverParams.MaximumServiceEndpoints)
	require.Equal(t, uint32(MaxUnifiedInterfaceDescriptors), params.ResolverParams.MaximumInterfaceDescriptors)
	require.Equal(t, uint64(MaxUnifiedRoutingMetadataBytes), params.ResolverParams.MaximumRoutingMetadataBytes)
	require.Equal(t, uint64(MaxInterfaceInlineSchemaBytesV2), params.ResolverParams.MaximumInlineSchemaBytes)
	require.Equal(t, uint64(1), params.ResolverParams.MinimumResolverTTL)
	require.Equal(t, IdentityGovernanceMaximumResolverTTLV2, params.ResolverParams.MaximumResolverTTL)
	require.ElementsMatch(t, []string{"aetra", "grpcs", "https", "ipfs", "wss"}, params.ResolverParams.AllowedEndpointSchemes)

	require.Equal(t, IdentityGovernanceMaximumDelegationDurationV2, params.DelegationParams.MaximumDelegationDuration)
	require.Equal(t, IdentityGovernanceMaximumScopedDelegatesPerDomain, params.DelegationParams.MaximumScopedDelegatesPerDomain)
	require.Equal(t, IdentityGovernanceMaximumZonePolicyBytesV2, params.DelegationParams.MaximumZonePolicySizeBytes)
	require.True(t, params.DelegationParams.DetachedSubdomainAllowed)
	require.True(t, params.DelegationParams.TimeLockedDelegationAllowed)

	require.Equal(t, DefaultAuctionCommitBlocks, params.AuctionParams.CommitPhaseDuration)
	require.Equal(t, DefaultAuctionRevealBlocks, params.AuctionParams.RevealPhaseDuration)
	require.Equal(t, sdkmath.NewInt(DefaultIdentitySpamAuctionBidDepositNaet), params.AuctionParams.BidDepositMinimum)
	require.Equal(t, DefaultIdentityAuctionUnrevealedForfeitBps, params.AuctionParams.UnrevealedBidPenaltyBps)
	require.Equal(t, IdentityAuctionTieBreakEarliestRevealThenCommitmentV2, params.AuctionParams.TieBreakRule)
	require.Equal(t, IdentityGovernanceAuctionFinalizationDelayV2, params.AuctionParams.AuctionFinalizationDelay)
	require.Equal(t, uint32(3_000), params.AuctionParams.FeeBurnBps)
	require.Equal(t, uint32(3_000), params.AuctionParams.FeeTreasuryBps)
	require.Equal(t, uint32(2_000), params.AuctionParams.FeeRewardsBps)
	require.Equal(t, uint32(2_000), params.AuctionParams.FeeCommunityPoolBps)

	require.Equal(t, uint32(MaxIdentityTxBatchResolverUpdatesV2), params.PerformanceParams.BatchResolverUpdateMaximumSize)
	require.Equal(t, uint32(MaxIdentityTxBatchRenewDomainsV2), params.PerformanceParams.BatchRenewalMaximumSize)
	require.Equal(t, uint8(MaxDomainLabels), params.PerformanceParams.RecursiveProofMaximumDepth)
	require.Equal(t, IdentityGovernanceCacheRecordMaximumLifetimeV2, params.PerformanceParams.CacheRecordMaximumLifetime)
	require.Equal(t, IdentityGovernanceStorePruningHorizonV2, params.PerformanceParams.StorePruningHorizonForProofAvailability)
	require.Equal(t, IdentityGovernanceExpiryProcessingLimitV2, params.PerformanceParams.ABCIExpiryWorkLimit)
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
	params.AuctionParams.BidDepositMinimum = sdkmath.NewInt(55)
	params.AuctionParams.UnrevealedBidPenaltyBps = 1_000
	params.AuctionParams.TieBreakRule = IdentityAuctionTieBreakCommitmentHashV2
	params.AuctionParams.FeeBurnBps = 2_500
	params.AuctionParams.FeeTreasuryBps = 2_500
	params.AuctionParams.FeeRewardsBps = 2_500
	params.AuctionParams.FeeCommunityPoolBps = 2_500
	params.ParamsHash = ComputeIdentityGovernanceParamsHashV2(params)

	identityParams, pricing, spam, auction, err := ApplyIdentityGovernanceParamsToRuntimeV2(params)
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
	require.Equal(t, sdkmath.NewInt(55), auction.BidDeposit)
	require.Equal(t, uint32(1_000), auction.UnrevealedForfeitBps)
	require.Equal(t, IdentityAuctionTieBreakCommitmentHashV2, auction.TieBreakRule)
	require.Equal(t, uint32(2_500), auction.FeeBurnBps)
	require.Equal(t, uint32(2_500), auction.FeeTreasuryBps)
	require.Equal(t, uint32(2_500), auction.FeeRewardsBps)
	require.Equal(t, uint32(2_500), auction.FeeCommunityPoolBps)
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

func TestIdentityGovernanceParamsV2RejectsInvalidResolverDelegationAuction(t *testing.T) {
	params, err := DefaultIdentityGovernanceParamsV2()
	require.NoError(t, err)

	badResolverBytes := params
	badResolverBytes.ResolverParams.MaximumResolverRecordBytes = MaxUnifiedPayloadBytesV2 + 1
	badResolverBytes.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badResolverBytes)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badResolverBytes), "maximum record bytes")

	badTTL := params
	badTTL.ResolverParams.MaximumResolverTTL = badTTL.ResolverParams.MinimumResolverTTL - 1
	badTTL.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badTTL)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badTTL), "maximum ttl")

	badSchemeOrder := params
	badSchemeOrder.ResolverParams.AllowedEndpointSchemes = []string{"https", "aetra"}
	badSchemeOrder.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badSchemeOrder)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badSchemeOrder), "schemes must be sorted")

	badScheme := params
	badScheme.ResolverParams.AllowedEndpointSchemes = []string{"ftp"}
	badScheme.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badScheme)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badScheme), "unsupported")

	badDelegationDuration := params
	badDelegationDuration.DelegationParams.MaximumDelegationDuration = 0
	badDelegationDuration.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badDelegationDuration)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badDelegationDuration), "maximum duration")

	badDelegates := params
	badDelegates.DelegationParams.MaximumScopedDelegatesPerDomain = 0
	badDelegates.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badDelegates)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badDelegates), "maximum scoped delegates")

	badAuctionDuration := params
	badAuctionDuration.AuctionParams.CommitPhaseDuration = 0
	badAuctionDuration.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badAuctionDuration)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badAuctionDuration), "commit phase duration")

	badPenalty := params
	badPenalty.AuctionParams.UnrevealedBidPenaltyBps = DomainDistributionDenominatorBps + 1
	badPenalty.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badPenalty)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badPenalty), "unrevealed bid penalty")

	badTieBreak := params
	badTieBreak.AuctionParams.TieBreakRule = IdentityAuctionTieBreakRuleV2("random")
	badTieBreak.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badTieBreak)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badTieBreak), "tie-break")

	badSplit := params
	badSplit.AuctionParams.FeeRewardsBps++
	badSplit.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badSplit)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badSplit), "fee split")
}

func TestIdentityGovernanceParamsV2RejectsInvalidPerformance(t *testing.T) {
	params, err := DefaultIdentityGovernanceParamsV2()
	require.NoError(t, err)

	badResolverBatch := params
	badResolverBatch.PerformanceParams.BatchResolverUpdateMaximumSize = MaxIdentityTxBatchResolverUpdatesV2 + 1
	badResolverBatch.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badResolverBatch)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badResolverBatch), "batch resolver update maximum size")

	badRenewalBatch := params
	badRenewalBatch.PerformanceParams.BatchRenewalMaximumSize = 0
	badRenewalBatch.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badRenewalBatch)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badRenewalBatch), "batch renewal maximum size")

	badRecursiveDepth := params
	badRecursiveDepth.PerformanceParams.RecursiveProofMaximumDepth = MaxDomainLabels + 1
	badRecursiveDepth.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badRecursiveDepth)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badRecursiveDepth), "recursive proof maximum depth")

	badCacheLifetime := params
	badCacheLifetime.PerformanceParams.CacheRecordMaximumLifetime = 0
	badCacheLifetime.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badCacheLifetime)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badCacheLifetime), "cache record maximum lifetime")

	badPruningHorizon := params
	badPruningHorizon.PerformanceParams.StorePruningHorizonForProofAvailability = badPruningHorizon.PerformanceParams.CacheRecordMaximumLifetime - 1
	badPruningHorizon.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badPruningHorizon)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badPruningHorizon), "store pruning horizon")

	badExpiryWork := params
	badExpiryWork.PerformanceParams.ABCIExpiryWorkLimit = 0
	badExpiryWork.ParamsHash = ComputeIdentityGovernanceParamsHashV2(badExpiryWork)
	require.ErrorContains(t, ValidateIdentityGovernanceParamsV2(badExpiryWork), "ABCI++ expiry work limit")

	require.NoError(t, IdentityGovernanceValidateBatchResolverUpdateCountV2(params, int(params.PerformanceParams.BatchResolverUpdateMaximumSize)))
	require.ErrorContains(t, IdentityGovernanceValidateBatchResolverUpdateCountV2(params, int(params.PerformanceParams.BatchResolverUpdateMaximumSize)+1), "resolver batch update count")
	require.NoError(t, IdentityGovernanceValidateBatchRenewalCountV2(params, int(params.PerformanceParams.BatchRenewalMaximumSize)))
	require.ErrorContains(t, IdentityGovernanceValidateBatchRenewalCountV2(params, int(params.PerformanceParams.BatchRenewalMaximumSize)+1), "batch renewal count")
}
