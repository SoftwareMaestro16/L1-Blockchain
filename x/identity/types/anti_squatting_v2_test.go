package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestIdentityAntiSquattingPriceQuoteV2(t *testing.T) {
	params := DefaultIdentityAntiSquattingParamsV2()
	quote, err := QuoteIdentityRegistrationPriceV2("Alice.AET", 256, true, params)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", quote.Name)
	require.Equal(t, "alice", quote.PricingLabel)
	require.Equal(t, DomainHighStartPrice, quote.RegistrationFee)
	expectedRenewal, err := RenewalFee("alice", params.DomainParams)
	require.NoError(t, err)
	require.Equal(t, expectedRenewal, quote.RenewalFee)
	require.Equal(t, params.ResolverStorageFeePerByte.Mul(sdkmath.NewInt(256)), quote.ResolverStorageFee)
	require.Equal(t, quote.RegistrationFee.Add(params.CommitmentDeposit).Add(quote.ResolverStorageFee), quote.TotalRegistrationCost)
	require.True(t, quote.ContestedAuctionRequired)
	require.Contains(t, quote.DeterministicFormula, "length_price")
}

func TestIdentityAntiSquattingBulkRegistrationSimulationV2(t *testing.T) {
	params := DefaultIdentityAntiSquattingParamsV2()
	params.BulkRegistrationWindowBlocks = 10
	params.MaxBulkRegistrationsPerAccount = 2
	owner := addr(1)
	attempts := []IdentityBulkRegistrationAttemptV2{
		{Owner: owner, Name: "alpha.aet", Height: 10},
		{Owner: owner, Name: "bravo.aet", Height: 11},
		{Owner: owner, Name: "charlie.aet", Height: 12},
		{Owner: owner, Name: "delta.aet", Height: 20},
		{Owner: addr(2), Name: "echo.aet", Height: 12},
	}

	sim := SimulateIdentityBulkRegistrationWindowV2(attempts, params)
	require.Equal(t, "input_index", sim.ResultOrder)
	require.Equal(t, uint32(4), sim.Allowed)
	require.Equal(t, uint32(1), sim.Rejected)
	require.Equal(t, IdentityAntiSquattingResultAllowedV2, sim.Results[0].Status)
	require.Equal(t, IdentityAntiSquattingResultAllowedV2, sim.Results[1].Status)
	require.Equal(t, IdentityAntiSquattingResultBulkLimitedV2, sim.Results[2].Status)
	require.Equal(t, IdentityAntiSquattingResultAllowedV2, sim.Results[3].Status)
	require.Equal(t, IdentityAntiSquattingResultAllowedV2, sim.Results[4].Status)
	require.Equal(t, uint64(1), sim.Results[0].WindowID)
	require.Equal(t, uint64(2), sim.Results[3].WindowID)
	require.NotNil(t, sim.Results[0].Quote)
}

func TestIdentityAntiSquattingRenewalIncentiveV2(t *testing.T) {
	params := DefaultIdentityAntiSquattingParamsV2()
	quote, err := QuoteIdentityRegistrationPriceV2("market.aet", 0, false, params)
	require.NoError(t, err)
	require.True(t, quote.RenewalFee.LT(quote.RegistrationFee), "renewal must be cheaper than fresh registration")

	state, domain := registerSpecDomain(t, "market", addr(1), "salt", 10)
	renewed, recovered, err := RecoverExpiredIdentityDomainV2(state, domain.Name, addr(1), domain.ExpiryHeight, params.GraceRecoveryCost, params)
	require.NoError(t, err)
	require.Greater(t, recovered.ExpiryHeight, domain.ExpiryHeight)
	require.Equal(t, DomainLifecycleActive, mustLifecycle(t, renewed, domain.Name, domain.ExpiryHeight+1))
}

func TestIdentityAntiSquattingExpiredDomainRecoveryAndReleaseV2(t *testing.T) {
	params := DefaultIdentityAntiSquattingParamsV2()
	params.ExpiredDomainGracePeriodBlocks = 5
	params.ExpiredDomainReleaseWindowBlocks = 6
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := SetIdentityResolver(state, domain.Name, addr(1), ResolverUpdate{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), domain.Name, 13)
	require.NoError(t, err)
	for i := range state.Domains {
		if state.Domains[i].Name == domain.Name {
			state.Domains[i].ExpiryHeight = 20
			state.Domains[i].UpdatedHeight = 14
		}
	}
	require.NoError(t, state.Validate())

	_, _, err = RecoverExpiredIdentityDomainV2(state, domain.Name, addr(1), 21, params.GraceRecoveryCost.SubRaw(1), params)
	require.ErrorContains(t, err, "recovery cost")
	recoveredState, recovered, err := RecoverExpiredIdentityDomainV2(state, domain.Name, addr(1), 21, params.GraceRecoveryCost, params)
	require.NoError(t, err)
	require.Greater(t, recovered.ExpiryHeight, uint64(21))
	require.Len(t, recoveredState.Domains, 1)

	_, _, err = RecoverExpiredIdentityDomainV2(state, domain.Name, addr(1), 26, params.GraceRecoveryCost, params)
	require.ErrorContains(t, err, "grace period ended")
	_, _, err = ReleaseExpiredIdentityDomainV2(state, domain.Name, 25, params)
	require.ErrorContains(t, err, "release window")

	released, previous, err := ReleaseExpiredIdentityDomainV2(state, domain.Name, 26, params)
	require.NoError(t, err)
	require.Equal(t, domain.Name, previous.Name)
	require.Empty(t, released.Domains)
	require.Empty(t, released.DomainNFTs)
	require.Empty(t, released.Resolvers)
	require.Empty(t, released.ReverseRecords)
	require.True(t, IsDomainAvailable(released, domain.Name, 26))
}
