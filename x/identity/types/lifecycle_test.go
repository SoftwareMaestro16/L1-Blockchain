package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestStartPriceByLength(t *testing.T) {
	params := DefaultDomainParams()

	_, err := StartPriceForName("abc", params)
	require.ErrorContains(t, err, "reserved")

	price, err := StartPriceForName("abcd", params)
	require.NoError(t, err)
	require.Equal(t, DomainPremiumStartPrice, price)

	price, err = StartPriceForName("abcde", params)
	require.NoError(t, err)
	require.Equal(t, DomainHighStartPrice, price)

	price, err = StartPriceForName("abcdefg", params)
	require.NoError(t, err)
	require.Equal(t, DomainMediumStartPrice, price)

	price, err = StartPriceForName("abcdefghijk", params)
	require.NoError(t, err)
	require.Equal(t, DomainLowStartPrice, price)
}

func TestCanStartAuctionLifecycle(t *testing.T) {
	now := int64(1_000)
	require.True(t, CanStartAuction(nil, now))
	require.True(t, CanStartAuction(&DomainRecord{Status: DomainStatusExpired, ExpiryUnix: now - 1}, now))
	require.True(t, CanStartAuction(&DomainRecord{Status: DomainStatusActive, ExpiryUnix: now}, now))
	require.False(t, CanStartAuction(&DomainRecord{Status: DomainStatusActive, ExpiryUnix: now + 1}, now))
	require.False(t, CanStartAuction(&DomainRecord{Status: DomainStatusAuction, ExpiryUnix: now - 1}, now))
}

func TestPlaceBidRequiresIncrementAndExtendsAntiSnipeBounded(t *testing.T) {
	params := DefaultDomainParams()
	auction, err := StartAuction("alice", 100, params)
	require.NoError(t, err)

	start, err := StartPriceForName("alice", params)
	require.NoError(t, err)

	_, err = PlaceBid(auction, []byte{1}, start.SubRaw(1), 101, params)
	require.ErrorContains(t, err, "bid below minimum")

	auction, err = PlaceBid(auction, []byte{1}, start, auction.EndUnix-1, params)
	require.NoError(t, err)
	require.Equal(t, uint32(1), auction.Extensions)

	minNext, err := MinimumNextBid(auction, params)
	require.NoError(t, err)
	require.Equal(t, start.Add(start.MulRaw(5).QuoRaw(100)), minNext)

	_, err = PlaceBid(auction, []byte{2}, minNext.SubRaw(1), auction.EndUnix-1, params)
	require.ErrorContains(t, err, "bid below minimum")

	for auction.Extensions < params.MaxExtensions {
		minNext, err = MinimumNextBid(auction, params)
		require.NoError(t, err)
		auction, err = PlaceBid(auction, []byte{3}, minNext, auction.EndUnix-1, params)
		require.NoError(t, err)
	}
	endBeforeCapHit := auction.EndUnix
	minNext, err = MinimumNextBid(auction, params)
	require.NoError(t, err)
	auction, err = PlaceBid(auction, []byte{4}, minNext, auction.EndUnix-1, params)
	require.NoError(t, err)
	require.Equal(t, params.MaxExtensions, auction.Extensions)
	require.Equal(t, endBeforeCapHit, auction.EndUnix)
}

func TestFinalizeAuctionAssignsOwnerNFTExpiryAndDistribution(t *testing.T) {
	params := DefaultDomainParams()
	auction, err := StartAuction("alice", 10, params)
	require.NoError(t, err)
	price, err := StartPriceForName("alice", params)
	require.NoError(t, err)
	winner := []byte{9}
	auction, err = PlaceBid(auction, winner, price, 11, params)
	require.NoError(t, err)

	_, _, _, err = FinalizeAuction(auction, auction.EndUnix-1, params)
	require.ErrorContains(t, err, "still active")

	record, split, finalized, err := FinalizeAuction(auction, auction.EndUnix, params)
	require.NoError(t, err)
	require.Equal(t, DomainStatusActive, record.Status)
	require.Equal(t, winner, []byte(record.Owner))
	require.Equal(t, DomainNFTItemID("alice"), record.NFTItemID)
	require.Equal(t, auction.EndUnix+params.RegistrationPeriodSeconds, record.ExpiryUnix)
	require.True(t, finalized.Finalized)
	require.Equal(t, price.MulRaw(40).QuoRaw(100), split.Burn)
	require.Equal(t, price.MulRaw(40).QuoRaw(100), split.Treasury)
	require.Equal(t, price.MulRaw(20).QuoRaw(100), split.Rewards)

	_, _, _, err = FinalizeAuction(finalized, auction.EndUnix, params)
	require.ErrorContains(t, err, "already finalized")
}

func TestRenewDomainExtendsExpiry(t *testing.T) {
	params := DefaultDomainParams()
	record := DomainRecord{
		Name:		"alice",
		TLD:		DomainTLD,
		Owner:		[]byte{1},
		ExpiryUnix:	1_000,
		NFTItemID:	DomainNFTItemID("alice"),
		Status:		DomainStatusActive,
		CreatedAtUnix:	1,
		UpdatedAtUnix:	2,
	}

	renewed, fee, err := RenewDomain(record, 500, params)
	require.NoError(t, err)
	require.Equal(t, int64(1_000)+params.RegistrationPeriodSeconds, renewed.ExpiryUnix)
	expectedFee, err := RenewalFee("alice", params)
	require.NoError(t, err)
	require.Equal(t, expectedFee, fee)

	renewed, _, err = RenewDomain(record, 2_000, params)
	require.NoError(t, err)
	require.Equal(t, int64(2_000)+params.RegistrationPeriodSeconds, renewed.ExpiryUnix)
}

func TestValidateDomainParams(t *testing.T) {
	params := DefaultDomainParams()
	require.NoError(t, ValidateDomainParams(params))

	params.FeeRewardsBps = 1
	require.ErrorContains(t, ValidateDomainParams(params), "sum")

	params = DefaultDomainParams()
	params.RenewalDiscountBps = 0
	require.ErrorContains(t, ValidateDomainParams(params), "renewal")

	params = DefaultDomainParams()
	params.LowStartPrice = sdkmath.ZeroInt()
	require.ErrorContains(t, ValidateDomainParams(params), "low_start_price")
}
