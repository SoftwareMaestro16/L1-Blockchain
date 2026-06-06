package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
)

func TestDefaultGenesisIsDisabledAndValid(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Channels)
	require.Empty(t, gs.State.Settlements)
}

func TestKeeperFeatureGateRejectsPaymentMutationWhenDisabled(t *testing.T) {
	k := NewKeeper()

	err := k.OpenChannel(paymentstypes.ChannelRecord{})
	require.ErrorContains(t, err, "disabled")
}

func TestKeeperPaymentLifecycleWhenEnabled(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x61)
	bob := keeperAddress(0x62)
	channel := keeperSignedChannel(t, "keeper-channel", "500", alice, bob)
	require.NoError(t, k.OpenChannel(channel))

	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "250"},
		{Participant: bob, Amount: "250"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, closeState, alice, 20, "5"))

	settlement, err := k.FinalizeSettlement(channel.ChannelID, 30)
	require.NoError(t, err)
	require.Equal(t, "245", keeperAmountFor(settlement.FinalBalances, alice))
	require.Equal(t, "250", keeperAmountFor(settlement.FinalBalances, bob))

	settlements, page, err := k.Settlements(nil)
	require.NoError(t, err)
	require.Zero(t, page.NextOffset)
	require.Equal(t, []paymentstypes.SettlementRecord{settlement}, settlements)
}

func keeperSignedChannel(t *testing.T, salt, collateral, left, right string) paymentstypes.ChannelRecord {
	t.Helper()

	channel := paymentstypes.ChannelRecord{
		ChainID:       "aetheris-test-1",
		ChannelID:     paymentstypes.HashParts(salt, left, right),
		ChannelType:   paymentstypes.ChannelTypeBidirectional,
		Participants:  []string{left, right},
		Denom:         paymentstypes.NativeDenom,
		Collateral:    collateral,
		OpenHeight:    10,
		DisputePeriod: 8,
		Status:        paymentstypes.ChannelStatusOpen,
	}
	openState := keeperSignedState(t, channel, 1, "", []paymentstypes.Balance{
		{Participant: left, Amount: collateral},
		{Participant: right, Amount: "0"},
	})
	channel.LatestState = openState
	channel.OpeningStateHash = openState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func keeperSignedState(t *testing.T, channel paymentstypes.ChannelRecord, nonce uint64, previous string, balances []paymentstypes.Balance) paymentstypes.ChannelState {
	t.Helper()

	state, err := paymentstypes.BuildState(paymentstypes.ChannelState{
		ChainID:           channel.ChainID,
		ChannelID:         channel.ChannelID,
		ChannelType:       channel.ChannelType,
		Denom:             channel.Denom,
		Version:           paymentstypes.CurrentStateVersion,
		Epoch:             1,
		Nonce:             nonce,
		Balances:          balances,
		PreviousStateHash: previous,
		TimeoutHeight:     channel.OpenHeight + channel.DisputePeriod + nonce,
		CloseDelay:        channel.DisputePeriod,
		FeePolicyID:       paymentstypes.NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := paymentstypes.SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func keeperAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(keeperBytes20(fill)))
}

func keeperBytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}

func keeperAmountFor(balances []paymentstypes.Balance, participant string) string {
	for _, balance := range balances {
		if balance.Participant == participant {
			return balance.Amount
		}
	}
	return ""
}
