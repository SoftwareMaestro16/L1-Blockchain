package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestIdentityEconomicsV2PricingBoundariesAndQueries(t *testing.T) {
	params := DefaultIdentityPricingParamsV2()

	short, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{Name: "abcd.aet"}, params)
	require.NoError(t, err)
	long, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{Name: "longdomain.aet"}, params)
	require.NoError(t, err)
	require.Equal(t, appparams.BaseDenom, short.Denom)
	require.True(t, short.Total.GT(long.Total))
	require.True(t, short.ScarcityFee.IsPositive())

	light, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{Name: "market.aet", ResolverPayloadBytes: 1}, params)
	require.NoError(t, err)
	heavy, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{Name: "market.aet", ResolverPayloadBytes: 1024}, params)
	require.NoError(t, err)
	require.True(t, heavy.Total.GT(light.Total))
	require.Equal(t, params.AntiSquattingParams.ResolverStorageFeePerByte.Mul(sdkmath.NewInt(1024)), heavy.StorageFootprintFee)

	root, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{Name: "alice.aet"}, params)
	require.NoError(t, err)
	child, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{
		Name:		"node.alice.aet",
		SubdomainMode:	IdentitySubdomainModeDetachedPaidV2,
	}, params)
	require.NoError(t, err)
	require.Equal(t, uint64(2), child.LabelCount)
	require.True(t, child.LabelDepthFee.IsPositive())
	require.Equal(t, params.DetachedSubdomainFee, child.SubdomainModeFee)
	require.True(t, child.Total.GT(root.Total))

	auction, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{
		Name:		"vault.aet",
		DemandClass:	IdentityDemandClassContestedV2,
		Auction:	true,
	}, params)
	require.NoError(t, err)
	require.True(t, auction.DemandFee.IsPositive())
	require.True(t, auction.AuctionSettlementFee.IsPositive())

	service := NewIdentityQueryServiceV2(IdentityQueryContextV2{State: EmptyIdentityState(DefaultIdentityParams()), Height: 7})
	regResp := service.QueryRegistrationPrice("market.aet", params.BaseDurationBlocks, IdentityDemandClassPremiumV2, true, 32, IdentitySubdomainModeRootV2)
	require.Equal(t, IdentityQueryOK, regResp.Code)
	require.NotNil(t, regResp.RegistrationPrice)
	require.Equal(t, appparams.BaseDenom, regResp.RegistrationPrice.Denom)
	require.True(t, regResp.RegistrationPrice.Total.IsPositive())

	renewResp := service.QueryRenewalPrice("market.aet", 2, 0)
	require.Equal(t, IdentityQueryOK, renewResp.Code)
	require.NotNil(t, renewResp.RenewalPrice)
	require.True(t, renewResp.RenewalPrice.RenewalCheaperThanFresh)
	require.True(t, renewResp.RenewalPrice.MultiPeriodDiscount.IsPositive())

	tooLong := service.QueryRegistrationPrice("market.aet", params.MaxRegistrationDuration+1, IdentityDemandClassStandardV2, false, 0, IdentitySubdomainModeRootV2)
	require.Equal(t, IdentityQueryInvalidRequest, tooLong.Code)
	require.Contains(t, tooLong.Error, "duration exceeds")
}

func TestIdentityEconomicsV2RenewalWindowGraceAndSoftFreeze(t *testing.T) {
	params := DefaultIdentityPricingParamsV2()
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)

	_, _, _, err := RenewIdentityDomainWithEconomicsV2(state, domain.Name, addr(1), 20, 1, sdkmath.NewInt(1_000_000_000_000_000), params)
	require.ErrorContains(t, err, "renewal window")

	windowAt := domain.ExpiryHeight - state.Params.RenewalWindowBlocks
	status, err := RenewalWindowTransitionStatusV2(state, domain.Name, windowAt)
	require.NoError(t, err)
	require.Equal(t, DomainLifecycleRenewalWindow, status)

	quote, err := QuoteIdentityRenewalPriceV2(domain.Name, 2, 0, params)
	require.NoError(t, err)
	renewedState, renewed, paidQuote, err := RenewIdentityDomainWithEconomicsV2(state, domain.Name, addr(1), windowAt, 2, quote.Total, params)
	require.NoError(t, err)
	require.Equal(t, quote.Total, paidQuote.Total)
	require.Equal(t, domain.ExpiryHeight+params.BaseDurationBlocks*2, renewed.ExpiryHeight)
	require.Equal(t, DomainLifecycleActive, mustLifecycle(t, renewedState, domain.Name, domain.ExpiryHeight+1))

	_, _, _, err = RenewIdentityDomainWithEconomicsV2(state, domain.Name, addr(1), windowAt, params.MaxRenewalPeriods+1, quote.Total, params)
	require.ErrorContains(t, err, "periods exceed")
	limited := params
	limited.MaxRegistrationDuration = params.BaseDurationBlocks
	_, _, _, err = RenewIdentityDomainWithEconomicsV2(state, domain.Name, addr(1), windowAt, 2, quote.Total, limited)
	require.ErrorContains(t, err, "duration exceeds")

	expired := state.Clone()
	for i := range expired.Domains {
		if expired.Domains[i].Name == domain.Name {
			expired.Domains[i].ExpiryHeight = 25
			expired.Domains[i].UpdatedHeight = 24
		}
	}
	require.NoError(t, expired.Validate())
	_, _, err = IssueSubdomain(expired, domain.Name, "api", addr(1), addr(2), false, 25)
	require.ErrorContains(t, err, "expired")
	require.ErrorContains(t, ValidateExpiredDomainResolverSoftFreezeV2(expired, domain.Name, addr(1), ResolverKeyPrimary, 25), "expired domain owner")
	require.NoError(t, ValidateExpiredDomainResolverSoftFreezeV2(expired, domain.Name, addr(1), ResolverRecoveryMetadataKeyV2, 25))

	recoveryQuote, err := QuoteIdentityDomainPriceV2(IdentityDomainPriceRequestV2{Name: domain.Name, Renewal: true, GraceRecovery: true}, params)
	require.NoError(t, err)
	_, _, _, err = RecoverExpiredIdentityDomainWithEconomicsV2(expired, domain.Name, addr(1), 26, recoveryQuote.Total.SubRaw(1), params)
	require.ErrorContains(t, err, "payment below")
	recoveredState, recovered, recoveryPaid, err := RecoverExpiredIdentityDomainWithEconomicsV2(expired, domain.Name, addr(1), 26, recoveryQuote.Total, params)
	require.NoError(t, err)
	require.Equal(t, recoveryQuote.Total, recoveryPaid.Total)
	require.Greater(t, recovered.ExpiryHeight, uint64(26))
	_, nft, err := ValidateRegistryNFTAuthority(recoveredState, recovered.Name, 26)
	require.NoError(t, err)
	require.Equal(t, recovered.Owner, nft.Owner)
}
