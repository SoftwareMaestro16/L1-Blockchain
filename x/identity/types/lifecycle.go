package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	DomainAuctionDurationSeconds		= int64(24 * 60 * 60)
	DomainAntiSnipeWindowSeconds		= int64(10 * 60)
	DomainAntiSnipeExtendSeconds		= int64(10 * 60)
	DomainMaxExtensions			= uint32(6)
	DomainRegistrationPeriodSeconds		= int64(365 * 24 * 60 * 60)
	DomainMinBidIncrementBps		= uint32(500)
	DomainRenewalDiscountBps		= uint32(1000)
	DomainFeeBurnBps			= uint32(4000)
	DomainFeeTreasuryBps			= uint32(4000)
	DomainFeeRewardsBps			= uint32(2000)
	DomainDistributionDenominatorBps	= uint32(10_000)
)

var (
	DomainPremiumStartPrice	= aetToNaet(10_000)
	DomainHighStartPrice	= aetToNaet(1_000)
	DomainMediumStartPrice	= aetToNaet(100)
	DomainLowStartPrice	= aetToNaet(10)
)

type DomainParams struct {
	PremiumStartPrice		sdkmath.Int
	HighStartPrice			sdkmath.Int
	MediumStartPrice		sdkmath.Int
	LowStartPrice			sdkmath.Int
	MinBidIncrementBps		uint32
	AuctionDurationSeconds		int64
	AntiSnipeWindowSeconds		int64
	AntiSnipeExtendSeconds		int64
	MaxExtensions			uint32
	RegistrationPeriodSeconds	int64
	RenewalDiscountBps		uint32
	FeeBurnBps			uint32
	FeeTreasuryBps			uint32
	FeeRewardsBps			uint32
}

type AuctionState struct {
	Name		string
	StartUnix	int64
	EndUnix		int64
	Extensions	uint32
	HighestBidder	sdk.AccAddress
	HighestBid	sdkmath.Int
	Finalized	bool
}

type DomainFeeDistribution struct {
	Burn		sdkmath.Int
	Treasury	sdkmath.Int
	Rewards		sdkmath.Int
}

func DefaultDomainParams() DomainParams {
	return DomainParams{
		PremiumStartPrice:		DomainPremiumStartPrice,
		HighStartPrice:			DomainHighStartPrice,
		MediumStartPrice:		DomainMediumStartPrice,
		LowStartPrice:			DomainLowStartPrice,
		MinBidIncrementBps:		DomainMinBidIncrementBps,
		AuctionDurationSeconds:		DomainAuctionDurationSeconds,
		AntiSnipeWindowSeconds:		DomainAntiSnipeWindowSeconds,
		AntiSnipeExtendSeconds:		DomainAntiSnipeExtendSeconds,
		MaxExtensions:			DomainMaxExtensions,
		RegistrationPeriodSeconds:	DomainRegistrationPeriodSeconds,
		RenewalDiscountBps:		DomainRenewalDiscountBps,
		FeeBurnBps:			DomainFeeBurnBps,
		FeeTreasuryBps:			DomainFeeTreasuryBps,
		FeeRewardsBps:			DomainFeeRewardsBps,
	}
}

func ValidateDomainParams(params DomainParams) error {
	for _, price := range []struct {
		label	string
		amount	sdkmath.Int
	}{
		{label: "premium_start_price", amount: params.PremiumStartPrice},
		{label: "high_start_price", amount: params.HighStartPrice},
		{label: "medium_start_price", amount: params.MediumStartPrice},
		{label: "low_start_price", amount: params.LowStartPrice},
	} {
		if !price.amount.IsPositive() {
			return fmt.Errorf("%s must be positive", price.label)
		}
	}
	if params.MinBidIncrementBps == 0 || params.MinBidIncrementBps > DomainDistributionDenominatorBps {
		return errors.New("min bid increment bps must be in 1..10000")
	}
	if params.AuctionDurationSeconds <= 0 {
		return errors.New("auction duration must be positive")
	}
	if params.AntiSnipeWindowSeconds <= 0 || params.AntiSnipeExtendSeconds <= 0 {
		return errors.New("anti-snipe window and extension must be positive")
	}
	if params.RegistrationPeriodSeconds <= 0 {
		return errors.New("registration period must be positive")
	}
	if params.RenewalDiscountBps == 0 || params.RenewalDiscountBps > DomainDistributionDenominatorBps {
		return errors.New("renewal discount bps must be in 1..10000")
	}
	if params.FeeBurnBps+params.FeeTreasuryBps+params.FeeRewardsBps != DomainDistributionDenominatorBps {
		return errors.New("domain fee distribution bps must sum to 10000")
	}
	return nil
}

func CanStartAuction(record *DomainRecord, nowUnix int64) bool {
	if record == nil {
		return true
	}
	if record.Status == DomainStatusAuction {
		return false
	}
	if record.Status == DomainStatusExpired {
		return true
	}
	return record.ExpiryUnix <= nowUnix
}

func StartAuction(name string, nowUnix int64, params DomainParams) (AuctionState, error) {
	if err := ValidateDomainParams(params); err != nil {
		return AuctionState{}, err
	}
	normalized, err := NormalizeDomainName(name)
	if err != nil {
		return AuctionState{}, err
	}
	if err := ValidateDomainName(normalized); err != nil {
		return AuctionState{}, err
	}
	if _, err := StartPriceForName(normalized, params); err != nil {
		return AuctionState{}, err
	}
	return AuctionState{
		Name:		normalized,
		StartUnix:	nowUnix,
		EndUnix:	nowUnix + params.AuctionDurationSeconds,
		HighestBid:	sdkmath.ZeroInt(),
	}, nil
}

func PlaceBid(auction AuctionState, bidder sdk.AccAddress, amount sdkmath.Int, nowUnix int64, params DomainParams) (AuctionState, error) {
	if err := ValidateAuctionState(auction); err != nil {
		return AuctionState{}, err
	}
	if err := ValidateDomainParams(params); err != nil {
		return AuctionState{}, err
	}
	if len(bidder) == 0 {
		return AuctionState{}, errors.New("bidder is required")
	}
	if err := addressing.RejectZeroAddress("bidder", bidder); err != nil {
		return AuctionState{}, err
	}
	if auction.Finalized {
		return AuctionState{}, errors.New("auction already finalized")
	}
	if nowUnix < auction.StartUnix {
		return AuctionState{}, errors.New("auction has not started")
	}
	if nowUnix >= auction.EndUnix {
		return AuctionState{}, errors.New("auction has ended")
	}
	minBid, err := MinimumNextBid(auction, params)
	if err != nil {
		return AuctionState{}, err
	}
	if amount.LT(minBid) {
		return AuctionState{}, fmt.Errorf("bid below minimum %s%s", minBid.String(), appparams.BaseDenom)
	}
	auction.HighestBidder = bidder
	auction.HighestBid = amount
	if auction.EndUnix-nowUnix <= params.AntiSnipeWindowSeconds && auction.Extensions < params.MaxExtensions {
		auction.EndUnix += params.AntiSnipeExtendSeconds
		auction.Extensions++
	}
	return auction, nil
}

func MinimumNextBid(auction AuctionState, params DomainParams) (sdkmath.Int, error) {
	if !auction.HighestBid.IsPositive() {
		return StartPriceForName(auction.Name, params)
	}
	increment := ceilBps(auction.HighestBid, params.MinBidIncrementBps)
	return auction.HighestBid.Add(increment), nil
}

func FinalizeAuction(auction AuctionState, nowUnix int64, params DomainParams) (DomainRecord, DomainFeeDistribution, AuctionState, error) {
	if err := ValidateAuctionState(auction); err != nil {
		return DomainRecord{}, DomainFeeDistribution{}, AuctionState{}, err
	}
	if err := ValidateDomainParams(params); err != nil {
		return DomainRecord{}, DomainFeeDistribution{}, AuctionState{}, err
	}
	if auction.Finalized {
		return DomainRecord{}, DomainFeeDistribution{}, AuctionState{}, errors.New("auction already finalized")
	}
	if nowUnix < auction.EndUnix {
		return DomainRecord{}, DomainFeeDistribution{}, AuctionState{}, errors.New("auction still active")
	}
	if len(auction.HighestBidder) == 0 || !auction.HighestBid.IsPositive() {
		return DomainRecord{}, DomainFeeDistribution{}, AuctionState{}, errors.New("auction has no valid winning bid")
	}
	if err := addressing.RejectZeroAddress("winner", auction.HighestBidder); err != nil {
		return DomainRecord{}, DomainFeeDistribution{}, AuctionState{}, err
	}
	record := DomainRecord{
		Name:		auction.Name,
		TLD:		DomainTLD,
		Owner:		auction.HighestBidder,
		ExpiryUnix:	nowUnix + params.RegistrationPeriodSeconds,
		NFTItemID:	DomainNFTItemID(auction.Name),
		Status:		DomainStatusActive,
		CreatedAtUnix:	nowUnix,
		UpdatedAtUnix:	nowUnix,
	}
	if err := ValidateDomainRecord(record); err != nil {
		return DomainRecord{}, DomainFeeDistribution{}, AuctionState{}, err
	}
	auction.Finalized = true
	return record, SplitDomainFee(auction.HighestBid, params), auction, nil
}

func RenewalFee(name string, params DomainParams) (sdkmath.Int, error) {
	startPrice, err := StartPriceForName(name, params)
	if err != nil {
		return sdkmath.Int{}, err
	}
	return ceilBps(startPrice, params.RenewalDiscountBps), nil
}

func RenewDomain(record DomainRecord, nowUnix int64, params DomainParams) (DomainRecord, sdkmath.Int, error) {
	if err := ValidateDomainRecord(record); err != nil {
		return DomainRecord{}, sdkmath.Int{}, err
	}
	if record.Status != DomainStatusActive {
		return DomainRecord{}, sdkmath.Int{}, errors.New("only active domain can be renewed")
	}
	fee, err := RenewalFee(record.Name, params)
	if err != nil {
		return DomainRecord{}, sdkmath.Int{}, err
	}
	base := record.ExpiryUnix
	if nowUnix > base {
		base = nowUnix
	}
	record.ExpiryUnix = base + params.RegistrationPeriodSeconds
	record.UpdatedAtUnix = nowUnix
	return record, fee, nil
}

func StartPriceForName(name string, params DomainParams) (sdkmath.Int, error) {
	if err := ValidateDomainParams(params); err != nil {
		return sdkmath.Int{}, err
	}
	normalized, err := NormalizeDomainName(name)
	if err != nil {
		return sdkmath.Int{}, err
	}
	if err := ValidateDomainName(normalized); err != nil {
		return sdkmath.Int{}, err
	}
	switch l := len(normalized); {
	case l <= 3:
		return sdkmath.Int{}, errors.New("domain length 1-3 is reserved or governance-only")
	case l == 4:
		return params.PremiumStartPrice, nil
	case l <= 6:
		return params.HighStartPrice, nil
	case l <= 10:
		return params.MediumStartPrice, nil
	default:
		return params.LowStartPrice, nil
	}
}

func SplitDomainFee(amount sdkmath.Int, params DomainParams) DomainFeeDistribution {
	burn := amount.MulRaw(int64(params.FeeBurnBps)).QuoRaw(int64(DomainDistributionDenominatorBps))
	treasury := amount.MulRaw(int64(params.FeeTreasuryBps)).QuoRaw(int64(DomainDistributionDenominatorBps))
	rewards := amount.Sub(burn).Sub(treasury)
	return DomainFeeDistribution{Burn: burn, Treasury: treasury, Rewards: rewards}
}

func ValidateAuctionState(auction AuctionState) error {
	if err := ValidateDomainName(auction.Name); err != nil {
		return err
	}
	if auction.EndUnix <= auction.StartUnix {
		return errors.New("auction end must be after start")
	}
	if auction.Extensions > DomainMaxExtensions {
		return errors.New("auction extensions exceed hard cap")
	}
	if auction.HighestBid.IsNegative() {
		return errors.New("auction highest bid must not be negative")
	}
	if auction.HighestBid.IsPositive() {
		if len(auction.HighestBidder) == 0 {
			return errors.New("auction highest bidder is required")
		}
		return addressing.RejectZeroAddress("auction highest bidder", auction.HighestBidder)
	}
	return nil
}

func DomainNFTItemID(name string) string {
	return "anft66:domain:" + name + DomainTLD
}

func aetToNaet(amount int64) sdkmath.Int {
	multiplier := int64(1_000_000_000)
	return sdkmath.NewInt(amount).MulRaw(multiplier)
}

func ceilBps(amount sdkmath.Int, bps uint32) sdkmath.Int {
	numerator := amount.MulRaw(int64(bps))
	divisor := int64(DomainDistributionDenominatorBps)
	return numerator.AddRaw(divisor - 1).QuoRaw(divisor)
}
