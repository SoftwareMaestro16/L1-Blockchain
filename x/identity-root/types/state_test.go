package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestNormalizeNameDeterministic(t *testing.T) {
	left, err := NormalizeName(" Alice.AET. ", "aet")
	require.NoError(t, err)
	right, err := NormalizeName("alice", ".AET")
	require.NoError(t, err)
	require.Equal(t, "alice.aet", left)
	require.Equal(t, left, right)
}

func TestExpiredNameCannotResolveAsActive(t *testing.T) {
	record := NameRecord{Name: "alice.aet", ExpiryHeight: 10}
	require.True(t, IsActive(record, 9))
	require.False(t, IsActive(record, 10))
	require.False(t, IsActive(record, 11))
}

func TestOwnershipBindingInvariant(t *testing.T) {
	params := DefaultIdentityRootParams()
	params.NFTBindingEnabled = true
	state := EmptyIdentityRootState()
	state.Records = append(state.Records, NameRecord{
		Name:				"alice.aet",
		Owner:				mustAEAddress("11"),
		ResolverRoot:			DefaultResolverRoot,
		ExpiryHeight:			100,
		RenewalHeight:			1,
		LastStorageChargeHeight:	1,
		RentPayerPolicy:		DomainRentPayerOwner,
		CreatedHeight:			1,
		UpdatedHeight:			1,
		NFTBinding: IdentityNFTBindingReference{
			Name:		"alice.aet",
			Enabled:	true,
			ClassID:	"identity",
			NFTID:		"alice",
			Owner:		mustAEAddress("22"),
		},
	})

	require.ErrorContains(t, state.Validate(params), "NFT binding owner")
}

func TestDomainRentDeltaIsDeterministic(t *testing.T) {
	params := DefaultIdentityRootParams()
	record := NameRecord{
		Name:				"alice.aet",
		Owner:				mustAEAddress("11"),
		ResolverRoot:			DefaultResolverRoot,
		ExpiryHeight:			100,
		RenewalHeight:			1,
		LastStorageChargeHeight:	1,
		RentPayerPolicy:		DomainRentPayerOwner,
		CreatedHeight:			1,
		UpdatedHeight:			1,
	}.Normalize(params)
	first, err := DomainStorageRentDelta(record, params, 11)
	require.NoError(t, err)
	second, err := DomainStorageRentDelta(record, params, 11)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.NotZero(t, first)
}

func mustAEAddress(hexByte string) string {
	bz, err := addressing.Parse("4:000000000000000000000000" + strings.Repeat(hexByte, 20))
	if err != nil {
		panic(err)
	}
	text, err := addressing.FormatUserFriendly(bz)
	if err != nil {
		panic(err)
	}
	return text
}
