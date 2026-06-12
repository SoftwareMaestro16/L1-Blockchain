package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMarketCannotReplaceBaseFee(t *testing.T) {
	params := DefaultParams()
	order := marketOrder("o1", marketAddr(0x11), ResourceCompute, 10, false, false, 1)

	selection, err := Select(params, []Order{order}, 1)
	require.NoError(t, err)
	require.Empty(t, selection.Accepted)
	require.Len(t, selection.Rejected, 1)
	require.Contains(t, selection.Rejected[0].Reason, "base naet fee")
	require.False(t, CanReplaceBaseFee())
}

func TestMarketPremiumAndPriorityAreCapped(t *testing.T) {
	params := DefaultParams()
	params.MaxPremiumNaet = 100
	params.FairnessPriorityCap = 50

	order := marketOrder("o1", marketAddr(0x11), ResourcePriority, 101, true, false, 1)
	selection, err := Select(params, []Order{order}, 1)
	require.NoError(t, err)
	require.Empty(t, selection.Accepted)
	require.Contains(t, selection.Rejected[0].Reason, "premium exceeds cap")

	capped := marketOrder("o2", marketAddr(0x11), ResourcePriority, 90, true, false, 1)
	require.Equal(t, uint64(50), PriorityScore(params, capped))
}

func TestMarketFairnessReservesNormalUserCapacity(t *testing.T) {
	params := DefaultParams()
	params.MinNormalSlots = 1
	params.MaxAccountShareBps = 10_000
	wealthy := marketOrder("wealthy", marketAddr(0x11), ResourcePriority, 1000, true, false, 1)
	normal := marketOrder("normal", marketAddr(0x22), ResourcePriority, 1, true, true, 2)

	selection, err := Select(params, []Order{wealthy, normal}, 1)
	require.NoError(t, err)
	require.Len(t, selection.Accepted, 1)
	require.Equal(t, "normal", selection.Accepted[0].ID)
}

func TestMarketAccountShareCapsStarvation(t *testing.T) {
	params := DefaultParams()
	params.MinNormalSlots = 0
	params.MaxAccountShareBps = 2_500
	orders := []Order{
		marketOrder("a1", marketAddr(0x11), ResourceCompute, 100, true, false, 1),
		marketOrder("a2", marketAddr(0x11), ResourceCompute, 90, true, false, 2),
		marketOrder("b1", marketAddr(0x22), ResourceCompute, 80, true, false, 3),
		marketOrder("c1", marketAddr(0x33), ResourceCompute, 70, true, false, 4),
	}

	selection, err := Select(params, orders, 4)
	require.NoError(t, err)
	require.Len(t, selection.Accepted, 3)
	require.Equal(t, []string{"a1", "b1", "c1"}, acceptedIDs(selection.Accepted))
	require.NotEmpty(t, selection.Rejected)
	require.Contains(t, selection.Rejected[0].Reason, "account share cap")
}

func marketOrder(id string, account sdk.AccAddress, resource string, premium uint64, baseFeePaid bool, normal bool, sequence uint64) Order {
	return Order{
		ID:		id,
		Account:	account,
		Resource:	resource,
		Quantity:	1,
		PremiumNaet:	premium,
		BaseFeePaid:	baseFeePaid,
		NormalUser:	normal,
		Sequence:	sequence,
	}
}

func marketAddr(fill byte) sdk.AccAddress {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return sdk.AccAddress(out)
}

func acceptedIDs(orders []Order) []string {
	out := make([]string, len(orders))
	for i, order := range orders {
		out[i] = order.ID
	}
	return out
}
