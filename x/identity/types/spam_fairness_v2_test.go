package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestIdentitySpamCostModelV2HighVolumeRegistrationSimulation(t *testing.T) {
	params := DefaultIdentitySpamCostParamsV2()
	params.PricingParams.AntiSquattingParams.BulkRegistrationWindowBlocks = 10
	params.PricingParams.AntiSquattingParams.MaxBulkRegistrationsPerAccount = 5
	owner := addr(1)
	attempts := make([]IdentityBulkRegistrationAttemptV2, 12)
	for i := range attempts {
		attempts[i] = IdentityBulkRegistrationAttemptV2{
			Owner:			owner,
			Name:			fmt.Sprintf("bulk%04d.aet", i),
			Height:			10,
			ResolverPayloadBytes:	uint64(i * 16),
		}
	}

	sim, err := SimulateHighVolumeIdentityRegistrationSpamV2(attempts, params)
	require.NoError(t, err)
	require.Equal(t, uint64(12), sim.Attempts)
	require.Equal(t, uint64(5), sim.Allowed)
	require.Equal(t, uint64(7), sim.Rejected)
	require.Equal(t, "input_index", sim.ResultOrder)
	require.True(t, sim.TotalCost.IsPositive())
	require.True(t, sim.HighestQuote.IsPositive())
	require.Equal(t, IdentityAntiSquattingResultBulkLimitedV2, sim.Results[5].Status)
}

func TestIdentitySpamCostModelV2SurfacesAndEndpointLimit(t *testing.T) {
	params := DefaultIdentitySpamCostParamsV2()
	quote, err := EstimateIdentitySpamCostV2(IdentitySpamCostRequestV2{
		Registrations: []IdentityDomainPriceRequestV2{
			{Name: "market.aet", ResolverPayloadBytes: 256},
			{Name: "vault.aet", DemandClass: IdentityDemandClassContestedV2, Auction: true},
		},
		CommitCount:			2,
		ResolverUpdateCount:		8,
		ResolverPayloadBytes:		2048,
		SubdomainCreates:		4,
		AuctionCommitments:		3,
		ServiceEndpointCount:		MaxUnifiedServiceEndpoints + 1,
		ProofQueries:			1000,
		BatchResolverGasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
	}, params)
	require.NoError(t, err)
	require.True(t, quote.RegistrationCost.IsPositive())
	require.Equal(t, params.PricingParams.AntiSquattingParams.CommitmentDeposit.Mul(sdkmath.NewInt(2)), quote.CommitmentDepositCost)
	require.Equal(t, params.SubdomainCreationFee.Mul(sdkmath.NewInt(4)), quote.SubdomainCreationCost)
	require.Equal(t, params.AuctionBidDeposit.Mul(sdkmath.NewInt(3)), quote.AuctionBidDepositCost)
	require.Equal(t, MinIdentityBatchResolverUpdateGasV2*8, quote.ResolverUpdateGas)
	require.False(t, quote.EndpointCountLimitPassed)
	require.False(t, quote.QueryRateLimitConsensus)
	require.Equal(t, params.ProofQueryCostUnit*1000, quote.ProofQueryNodeCostUnits)
	require.True(t, quote.TotalConsensusCost.GT(quote.RegistrationCost))
}

func TestIdentitySpamSubdomainCreationStressV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	const children = 128
	for i := 0; i < children; i++ {
		var err error
		state, _, err = IssueSubdomain(state, domain.Name, fmt.Sprintf("svc%04d", i), addr(1), addr(byte(i%200)+2), false, uint64(11+i))
		require.NoError(t, err)
	}
	require.Len(t, state.Subdomains, children)
	require.Len(t, state.Domains, children+1)
	require.NoError(t, state.Validate())
}

func TestIdentityAuctionFairnessV2StateMachineTieForfeitAndFinalization(t *testing.T) {
	params := DefaultIdentityAuctionFairnessParamsV2("aetra-local-1")
	state := EmptyIdentityState(DefaultIdentityParams())
	state, auction, err := StartSealedAuction(state, "market.aet", 100)
	require.NoError(t, err)

	machine, err := DescribeIdentityAuctionStateMachineV2(auction, params)
	require.NoError(t, err)
	require.Equal(t, auction.RevealStartHeight, machine.CommitEndHeight)
	require.Equal(t, "any_account_after_reveal_window", machine.FinalizerPolicy)

	leftCommit, err := ComputeAuctionCommitment(auction.Name, addr(1), 100, "left")
	require.NoError(t, err)
	rightCommit, err := ComputeAuctionCommitment(auction.Name, addr(2), 100, "right")
	require.NoError(t, err)
	hiddenCommit, err := ComputeAuctionCommitment(auction.Name, addr(3), 200, "hidden")
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(1), leftCommit, auction.CommitStartHeight)
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(2), rightCommit, auction.CommitStartHeight+1)
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(3), hiddenCommit, auction.CommitStartHeight+2)
	require.NoError(t, err)

	_, _, err = RevealAuctionBid(state, auction.Name, addr(1), 100, "left", auction.RevealStartHeight-1)
	require.ErrorContains(t, err, "not in reveal phase")
	state, _, err = RevealAuctionBid(state, auction.Name, addr(2), 100, "right", auction.RevealStartHeight)
	require.NoError(t, err)
	state, _, err = RevealAuctionBid(state, auction.Name, addr(1), 100, "left", auction.RevealStartHeight+1)
	require.NoError(t, err)
	_, _, err = RevealAuctionBid(state, auction.Name, addr(3), 200, "hidden", auction.RevealEndHeight)
	require.ErrorContains(t, err, "not in reveal phase")

	_, _, err = FinalizeSealedAuctionFairV2(state, auction.Name, addr(9), auction.RevealEndHeight-1, params)
	require.ErrorContains(t, err, "reveal phase is not over")
	finalState, result, err := FinalizeSealedAuctionFairV2(state, auction.Name, addr(9), auction.RevealEndHeight, params)
	require.NoError(t, err)
	require.Equal(t, addr(2), result.Winner, "earlier reveal wins equal bids")
	require.Equal(t, uint64(100), result.WinningBid)
	require.True(t, result.PermissionlessFinal)
	require.Equal(t, addr(9), result.FinalizedBy)
	require.Len(t, result.LosingBidRefunds, 1)
	require.Len(t, result.UnrevealedForfeits, 1)
	require.Equal(t, "unrevealed_bid_forfeit", result.UnrevealedForfeits[0].Reason)
	require.Equal(t, ceilBps(params.BidDeposit, params.UnrevealedForfeitBps).Uint64(), result.UnrevealedForfeits[0].Amount)
	require.Equal(t, AuctionPhaseFinalized, finalState.Auctions[0].Phase)
	require.Len(t, finalState.Auctions[0].Refunds, 2)
	totalSplit := result.FeeSplit.Burn.Add(result.FeeSplit.Treasury).Add(result.FeeSplit.Rewards).Add(result.FeeSplit.CommunityPool)
	require.Equal(t, sdkmath.NewIntFromUint64(result.WinningBid), totalSplit)
}

func TestIdentityAuctionCommitmentV2BindsChainDomain(t *testing.T) {
	params := DefaultIdentityAuctionFairnessParamsV2("aetra-local-1")
	commitment, err := ComputeAuctionCommitmentV2("market.aet", addr(1), 500, "salt", params.ChainID, params.ModuleVersion)
	require.NoError(t, err)
	require.True(t, AuctionCommitmentMatchesChainDomainV2("market.aet", addr(1), 500, "salt", params.ChainID, params.ModuleVersion, commitment))
	require.False(t, AuctionCommitmentMatchesChainDomainV2("market.aet", addr(1), 500, "salt", "other-chain", params.ModuleVersion, commitment))
	require.False(t, AuctionCommitmentMatchesChainDomainV2("market.aet", addr(1), 501, "salt", params.ChainID, params.ModuleVersion, commitment))
}
