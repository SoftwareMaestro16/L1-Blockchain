package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultIdentitySpamActiveRecordStorageFeeNaet	= int64(2_000_000_000)
	DefaultIdentitySpamSubdomainCreationFeeNaet	= int64(1_000_000_000)
	DefaultIdentitySpamAuctionBidDepositNaet	= int64(5_000_000_000)
	DefaultIdentitySpamEndpointRecordFeeNaet	= int64(250_000_000)
	DefaultIdentitySpamProofQueryCostUnit		= uint64(1)
	DefaultIdentitySpamQueryRateLimitPerMinute	= uint64(600)
	DefaultIdentitySpamResolverUpdateGas		= uint64(25_000)
	DefaultIdentityAuctionUnrevealedForfeitBps	= uint32(5_000)
	DefaultIdentityAuctionModuleVersionV2		= uint64(2)

	IdentityAuctionTieBreakEarliestRevealThenCommitmentV2	IdentityAuctionTieBreakRuleV2	= "earliest_reveal_then_commitment_hash"
	IdentityAuctionTieBreakCommitmentHashV2			IdentityAuctionTieBreakRuleV2	= "commitment_hash"
)

type IdentitySpamCostParamsV2 struct {
	PricingParams			IdentityPricingParamsV2
	ActiveRecordStorageFee		sdkmath.Int
	SubdomainCreationFee		sdkmath.Int
	AuctionBidDeposit		sdkmath.Int
	EndpointRecordFee		sdkmath.Int
	ProofQueryCostUnit		uint64
	QueryRateLimitPerMinute		uint64
	ResolverUpdateGasPerWrite	uint64
}

type IdentitySpamCostRequestV2 struct {
	Registrations			[]IdentityDomainPriceRequestV2
	CommitCount			uint64
	ResolverUpdateCount		uint64
	ResolverPayloadBytes		uint64
	SubdomainCreates		uint64
	AuctionCommitments		uint64
	ServiceEndpointCount		uint64
	ProofQueries			uint64
	BatchResolverGasPerUpdate	uint64
}

type IdentitySpamCostQuoteV2 struct {
	Denom				string
	RegistrationCost		sdkmath.Int
	CommitmentDepositCost		sdkmath.Int
	ActiveRecordStorageCost		sdkmath.Int
	ResolverPayloadCost		sdkmath.Int
	ResolverUpdateGas		uint64
	SubdomainCreationCost		sdkmath.Int
	AuctionBidDepositCost		sdkmath.Int
	ServiceEndpointCost		sdkmath.Int
	ProofQueryNodeCostUnits		uint64
	QueryRateLimitPerMinute		uint64
	TotalConsensusCost		sdkmath.Int
	DeterministicFormula		string
	QueryRateLimitConsensus		bool
	EndpointCountLimit		uint64
	EndpointCountLimitPassed	bool
}

type IdentitySpamRegistrationSimulationV2 struct {
	Attempts	uint64
	Allowed		uint64
	Rejected	uint64
	TotalCost	sdkmath.Int
	HighestQuote	sdkmath.Int
	Results		[]IdentityBulkRegistrationResultV2
	ResultOrder	string
	BulkWindowSize	uint64
}

type IdentityAuctionTieBreakRuleV2 string

type IdentityAuctionFairnessParamsV2 struct {
	ChainID				string
	ModuleVersion			uint64
	TieBreakRule			IdentityAuctionTieBreakRuleV2
	BidDeposit			sdkmath.Int
	UnrevealedForfeitBps		uint32
	FeeBurnBps			uint32
	FeeTreasuryBps			uint32
	FeeRewardsBps			uint32
	FeeCommunityPoolBps		uint32
	AllowPermissionlessFinal	bool
}

type IdentityAuctionStateMachineV2 struct {
	CommitStartHeight	uint64
	CommitEndHeight		uint64
	RevealStartHeight	uint64
	RevealEndHeight		uint64
	TieBreakRule		IdentityAuctionTieBreakRuleV2
	FinalizerPolicy		string
}

type IdentityAuctionFeeSplitV2 struct {
	Burn		sdkmath.Int
	Treasury	sdkmath.Int
	Rewards		sdkmath.Int
	CommunityPool	sdkmath.Int
}

type IdentityAuctionFairFinalizationV2 struct {
	Auction			Auction
	Winner			sdk.AccAddress
	WinningBid		uint64
	TieBreakRule		IdentityAuctionTieBreakRuleV2
	UnrevealedForfeits	[]AuctionRefundReceipt
	LosingBidRefunds	[]AuctionRefundReceipt
	FeeSplit		IdentityAuctionFeeSplitV2
	FinalizedBy		sdk.AccAddress
	PermissionlessFinal	bool
}

func DefaultIdentitySpamCostParamsV2() IdentitySpamCostParamsV2 {
	return IdentitySpamCostParamsV2{
		PricingParams:			DefaultIdentityPricingParamsV2(),
		ActiveRecordStorageFee:		sdkmath.NewInt(DefaultIdentitySpamActiveRecordStorageFeeNaet),
		SubdomainCreationFee:		sdkmath.NewInt(DefaultIdentitySpamSubdomainCreationFeeNaet),
		AuctionBidDeposit:		sdkmath.NewInt(DefaultIdentitySpamAuctionBidDepositNaet),
		EndpointRecordFee:		sdkmath.NewInt(DefaultIdentitySpamEndpointRecordFeeNaet),
		ProofQueryCostUnit:		DefaultIdentitySpamProofQueryCostUnit,
		QueryRateLimitPerMinute:	DefaultIdentitySpamQueryRateLimitPerMinute,
		ResolverUpdateGasPerWrite:	DefaultIdentitySpamResolverUpdateGas,
	}
}

func ValidateIdentitySpamCostParamsV2(params IdentitySpamCostParamsV2) error {
	if err := ValidateIdentityPricingParamsV2(params.PricingParams); err != nil {
		return err
	}
	for _, amount := range []struct {
		label	string
		value	sdkmath.Int
	}{
		{label: "active record storage fee", value: params.ActiveRecordStorageFee},
		{label: "subdomain creation fee", value: params.SubdomainCreationFee},
		{label: "auction bid deposit", value: params.AuctionBidDeposit},
		{label: "endpoint record fee", value: params.EndpointRecordFee},
	} {
		if amount.value.IsNil() || amount.value.IsNegative() {
			return fmt.Errorf("identity spam %s must not be negative", amount.label)
		}
	}
	if params.ProofQueryCostUnit == 0 {
		return errors.New("identity spam proof query cost unit is required")
	}
	if params.QueryRateLimitPerMinute == 0 {
		return errors.New("identity spam query rate limit is required")
	}
	if params.ResolverUpdateGasPerWrite == 0 {
		return errors.New("identity spam resolver update gas is required")
	}
	return nil
}

func EstimateIdentitySpamCostV2(req IdentitySpamCostRequestV2, params IdentitySpamCostParamsV2) (IdentitySpamCostQuoteV2, error) {
	if err := ValidateIdentitySpamCostParamsV2(params); err != nil {
		return IdentitySpamCostQuoteV2{}, err
	}
	quote := IdentitySpamCostQuoteV2{
		Denom:				"naet",
		RegistrationCost:		sdkmath.ZeroInt(),
		CommitmentDepositCost:		sdkmath.ZeroInt(),
		ActiveRecordStorageCost:	sdkmath.ZeroInt(),
		ResolverPayloadCost:		sdkmath.ZeroInt(),
		SubdomainCreationCost:		sdkmath.ZeroInt(),
		AuctionBidDepositCost:		sdkmath.ZeroInt(),
		ServiceEndpointCost:		sdkmath.ZeroInt(),
		TotalConsensusCost:		sdkmath.ZeroInt(),
		QueryRateLimitPerMinute:	params.QueryRateLimitPerMinute,
		DeterministicFormula:		"registration_quotes + commit_deposits + active_record_storage + resolver_bytes*fee_per_byte + resolver_update_gas + subdomain_fee + auction_bid_deposit + endpoint_fee; proof_query_rate_limit_is_node_local",
		QueryRateLimitConsensus:	false,
		EndpointCountLimit:		MaxUnifiedServiceEndpoints,
		EndpointCountLimitPassed:	req.ServiceEndpointCount <= MaxUnifiedServiceEndpoints,
	}
	for _, reg := range req.Registrations {
		regQuote, err := QuoteIdentityDomainPriceV2(reg, params.PricingParams)
		if err != nil {
			return IdentitySpamCostQuoteV2{}, err
		}
		quote.RegistrationCost = quote.RegistrationCost.Add(regQuote.Total)
	}
	quote.CommitmentDepositCost = params.PricingParams.AntiSquattingParams.CommitmentDeposit.Mul(sdkmath.NewIntFromUint64(req.CommitCount))
	quote.ActiveRecordStorageCost = params.ActiveRecordStorageFee.Mul(sdkmath.NewIntFromUint64(uint64(len(req.Registrations)) + req.SubdomainCreates))
	quote.ResolverPayloadCost = params.PricingParams.AntiSquattingParams.ResolverStorageFeePerByte.Mul(sdkmath.NewIntFromUint64(req.ResolverPayloadBytes))
	perUpdateGas := req.BatchResolverGasPerUpdate
	if perUpdateGas == 0 {
		perUpdateGas = params.ResolverUpdateGasPerWrite
	}
	quote.ResolverUpdateGas = perUpdateGas * req.ResolverUpdateCount
	quote.SubdomainCreationCost = params.SubdomainCreationFee.Mul(sdkmath.NewIntFromUint64(req.SubdomainCreates))
	quote.AuctionBidDepositCost = params.AuctionBidDeposit.Mul(sdkmath.NewIntFromUint64(req.AuctionCommitments))
	quote.ServiceEndpointCost = params.EndpointRecordFee.Mul(sdkmath.NewIntFromUint64(req.ServiceEndpointCount))
	quote.ProofQueryNodeCostUnits = params.ProofQueryCostUnit * req.ProofQueries
	quote.TotalConsensusCost = quote.RegistrationCost.
		Add(quote.CommitmentDepositCost).
		Add(quote.ActiveRecordStorageCost).
		Add(quote.ResolverPayloadCost).
		Add(quote.SubdomainCreationCost).
		Add(quote.AuctionBidDepositCost).
		Add(quote.ServiceEndpointCost)
	return quote, nil
}

func SimulateHighVolumeIdentityRegistrationSpamV2(attempts []IdentityBulkRegistrationAttemptV2, params IdentitySpamCostParamsV2) (IdentitySpamRegistrationSimulationV2, error) {
	if err := ValidateIdentitySpamCostParamsV2(params); err != nil {
		return IdentitySpamRegistrationSimulationV2{}, err
	}
	bulk := SimulateIdentityBulkRegistrationWindowV2(attempts, params.PricingParams.AntiSquattingParams)
	sim := IdentitySpamRegistrationSimulationV2{
		Attempts:	uint64(len(attempts)),
		Allowed:	uint64(bulk.Allowed),
		Rejected:	uint64(bulk.Rejected),
		TotalCost:	sdkmath.ZeroInt(),
		HighestQuote:	sdkmath.ZeroInt(),
		Results:	append([]IdentityBulkRegistrationResultV2(nil), bulk.Results...),
		ResultOrder:	bulk.ResultOrder,
		BulkWindowSize:	params.PricingParams.AntiSquattingParams.BulkRegistrationWindowBlocks,
	}
	for _, result := range bulk.Results {
		if result.Quote == nil {
			continue
		}
		cost := result.Quote.TotalRegistrationCost.Add(params.ActiveRecordStorageFee)
		sim.TotalCost = sim.TotalCost.Add(cost)
		if cost.GT(sim.HighestQuote) {
			sim.HighestQuote = cost
		}
	}
	return sim, nil
}

func DefaultIdentityAuctionFairnessParamsV2(chainID string) IdentityAuctionFairnessParamsV2 {
	return IdentityAuctionFairnessParamsV2{
		ChainID:			chainID,
		ModuleVersion:			DefaultIdentityAuctionModuleVersionV2,
		TieBreakRule:			IdentityAuctionTieBreakEarliestRevealThenCommitmentV2,
		BidDeposit:			sdkmath.NewInt(DefaultIdentitySpamAuctionBidDepositNaet),
		UnrevealedForfeitBps:		DefaultIdentityAuctionUnrevealedForfeitBps,
		FeeBurnBps:			3_000,
		FeeTreasuryBps:			3_000,
		FeeRewardsBps:			2_000,
		FeeCommunityPoolBps:		2_000,
		AllowPermissionlessFinal:	true,
	}
}

func ValidateIdentityAuctionFairnessParamsV2(params IdentityAuctionFairnessParamsV2) error {
	if params.ChainID == "" {
		return errors.New("identity auction fairness chain_id is required")
	}
	if params.ModuleVersion == 0 {
		return errors.New("identity auction fairness module_version is required")
	}
	switch params.TieBreakRule {
	case "", IdentityAuctionTieBreakEarliestRevealThenCommitmentV2, IdentityAuctionTieBreakCommitmentHashV2:
	default:
		return fmt.Errorf("unsupported identity auction tie break rule %q", params.TieBreakRule)
	}
	if params.BidDeposit.IsNil() || params.BidDeposit.IsNegative() {
		return errors.New("identity auction fairness bid deposit must not be negative")
	}
	if params.UnrevealedForfeitBps > DomainDistributionDenominatorBps {
		return errors.New("identity auction fairness unrevealed forfeit bps must be <= 10000")
	}
	if params.FeeBurnBps+params.FeeTreasuryBps+params.FeeRewardsBps+params.FeeCommunityPoolBps != DomainDistributionDenominatorBps {
		return errors.New("identity auction fairness fee split bps must sum to 10000")
	}
	return nil
}

func ComputeAuctionCommitmentV2(name string, bidder sdk.AccAddress, bid uint64, salt string, chainID string, moduleVersion uint64) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 auction bidder", bidder); err != nil {
		return "", err
	}
	if bid == 0 {
		return "", errors.New("identity v2 auction bid must be positive")
	}
	if salt == "" {
		return "", errors.New("identity v2 auction salt is required")
	}
	if chainID == "" {
		return "", errors.New("identity v2 auction chain_id is required")
	}
	if moduleVersion == 0 {
		return "", errors.New("identity v2 auction module_version is required")
	}
	return identityHash("identity-v2-auction-commitment", chainID, fmt.Sprintf("%020d", moduleVersion), normalized, string(bidder), fmt.Sprintf("%020d", bid), salt), nil
}

func DescribeIdentityAuctionStateMachineV2(auction Auction, params IdentityAuctionFairnessParamsV2) (IdentityAuctionStateMachineV2, error) {
	if err := validateAuction(auction); err != nil {
		return IdentityAuctionStateMachineV2{}, err
	}
	if err := ValidateIdentityAuctionFairnessParamsV2(params); err != nil {
		return IdentityAuctionStateMachineV2{}, err
	}
	finalizerPolicy := "auction_admin"
	if params.AllowPermissionlessFinal {
		finalizerPolicy = "any_account_after_reveal_window"
	}
	rule := params.TieBreakRule
	if rule == "" {
		rule = IdentityAuctionTieBreakEarliestRevealThenCommitmentV2
	}
	return IdentityAuctionStateMachineV2{
		CommitStartHeight:	auction.CommitStartHeight,
		CommitEndHeight:	auction.RevealStartHeight,
		RevealStartHeight:	auction.RevealStartHeight,
		RevealEndHeight:	auction.RevealEndHeight,
		TieBreakRule:		rule,
		FinalizerPolicy:	finalizerPolicy,
	}, nil
}

func FinalizeSealedAuctionFairV2(state IdentityState, name string, finalizer sdk.AccAddress, height uint64, params IdentityAuctionFairnessParamsV2) (IdentityState, IdentityAuctionFairFinalizationV2, error) {
	if err := ValidateIdentityAuctionFairnessParamsV2(params); err != nil {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, err
	}
	if err := validateSpecAddress("identity v2 auction finalizer", finalizer); err != nil {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, err
	}
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, err
	}
	auction, found := findAuction(state, name)
	if !found {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, errors.New("identity auction not found")
	}
	if height < auction.RevealEndHeight {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, errors.New("identity auction reveal phase is not over")
	}
	if auction.Phase == AuctionPhaseFinalized {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, errors.New("identity auction already finalized")
	}
	if len(auction.Reveals) == 0 {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, errors.New("identity auction has no revealed bids")
	}
	winner := chooseAuctionWinnerByRuleV2(auction.Reveals, params.TieBreakRule)
	losingRefunds := buildAuctionRefunds(auction, winner)
	forfeits := buildAuctionUnrevealedForfeitsV2(auction, params)
	auction.Phase = AuctionPhaseFinalized
	auction.Winner = cloneSpecAddress(winner.Bidder)
	auction.WinningBid = winner.Bid
	auction.WinningCommitment = winner.CommitmentHash
	auction.Refunds = append(cloneAuctionRefunds(losingRefunds), forfeits...)
	sortAuctionRefunds(auction.Refunds)
	next := state.Clone()
	next.Auctions = upsertAuction(next.Auctions, auction)
	sortAuctions(next.Auctions)
	if err := next.Validate(); err != nil {
		return IdentityState{}, IdentityAuctionFairFinalizationV2{}, err
	}
	result := IdentityAuctionFairFinalizationV2{
		Auction:		auction,
		Winner:			cloneSpecAddress(winner.Bidder),
		WinningBid:		winner.Bid,
		TieBreakRule:		effectiveAuctionTieBreakRuleV2(params.TieBreakRule),
		UnrevealedForfeits:	forfeits,
		LosingBidRefunds:	losingRefunds,
		FeeSplit:		SplitIdentityAuctionFeeV2(sdkmath.NewIntFromUint64(winner.Bid), params),
		FinalizedBy:		cloneSpecAddress(finalizer),
		PermissionlessFinal:	params.AllowPermissionlessFinal,
	}
	return next, result, nil
}

func SplitIdentityAuctionFeeV2(amount sdkmath.Int, params IdentityAuctionFairnessParamsV2) IdentityAuctionFeeSplitV2 {
	burn := amount.MulRaw(int64(params.FeeBurnBps)).QuoRaw(int64(DomainDistributionDenominatorBps))
	treasury := amount.MulRaw(int64(params.FeeTreasuryBps)).QuoRaw(int64(DomainDistributionDenominatorBps))
	rewards := amount.MulRaw(int64(params.FeeRewardsBps)).QuoRaw(int64(DomainDistributionDenominatorBps))
	community := amount.Sub(burn).Sub(treasury).Sub(rewards)
	return IdentityAuctionFeeSplitV2{Burn: burn, Treasury: treasury, Rewards: rewards, CommunityPool: community}
}

func chooseAuctionWinnerByRuleV2(reveals []AuctionReveal, rule IdentityAuctionTieBreakRuleV2) AuctionReveal {
	ordered := cloneAuctionReveals(reveals)
	sort.SliceStable(ordered, func(i, j int) bool {
		return compareAuctionWinnerByRuleV2(ordered[i], ordered[j], rule) < 0
	})
	return ordered[0]
}

func compareAuctionWinnerByRuleV2(left, right AuctionReveal, rule IdentityAuctionTieBreakRuleV2) int {
	if left.Bid != right.Bid {
		if left.Bid > right.Bid {
			return -1
		}
		return 1
	}
	switch effectiveAuctionTieBreakRuleV2(rule) {
	case IdentityAuctionTieBreakCommitmentHashV2:
		return stringsCompare(left.CommitmentHash, right.CommitmentHash)
	default:
		if left.RevealHeight != right.RevealHeight {
			if left.RevealHeight < right.RevealHeight {
				return -1
			}
			return 1
		}
		return stringsCompare(left.CommitmentHash, right.CommitmentHash)
	}
}

func buildAuctionUnrevealedForfeitsV2(auction Auction, params IdentityAuctionFairnessParamsV2) []AuctionRefundReceipt {
	revealed := map[string]struct{}{}
	for _, reveal := range auction.Reveals {
		revealed[reveal.CommitmentHash] = struct{}{}
	}
	forfeit := ceilBps(params.BidDeposit, params.UnrevealedForfeitBps)
	if forfeit.IsZero() {
		return nil
	}
	receipts := make([]AuctionRefundReceipt, 0)
	for _, commitment := range auction.Commitments {
		if _, found := revealed[commitment.CommitmentHash]; found {
			continue
		}
		receipts = append(receipts, AuctionRefundReceipt{
			ReceiptID:	identityHash("auction-unrevealed-forfeit", auction.Name, commitment.CommitmentHash),
			Name:		auction.Name,
			Bidder:		cloneSpecAddress(commitment.Bidder),
			Amount:		forfeit.Uint64(),
			CommitmentHash:	commitment.CommitmentHash,
			Reason:		"unrevealed_bid_forfeit",
		})
	}
	sortAuctionRefunds(receipts)
	return receipts
}

func effectiveAuctionTieBreakRuleV2(rule IdentityAuctionTieBreakRuleV2) IdentityAuctionTieBreakRuleV2 {
	if rule == "" {
		return IdentityAuctionTieBreakEarliestRevealThenCommitmentV2
	}
	return rule
}

func AuctionCommitmentMatchesChainDomainV2(name string, bidder sdk.AccAddress, bid uint64, salt string, chainID string, moduleVersion uint64, commitment string) bool {
	expected, err := ComputeAuctionCommitmentV2(name, bidder, bid, salt, chainID, moduleVersion)
	return err == nil && expected == commitment
}
