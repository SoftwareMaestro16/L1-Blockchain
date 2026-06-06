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
	height, found, err := k.QueryPendingFinalizationHeight(channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(28), height)
	require.NoError(t, k.AdvanceChannelFinality(channel.ChannelID, 28))
	debug, err := k.QueryStateHash(channel.ChannelID)
	require.NoError(t, err)
	require.Equal(t, paymentstypes.ChannelStatusPendingClose, debug.Status)
	require.Equal(t, closeState.Nonce, debug.PendingNonce)
	require.Equal(t, closeState.StateHash, debug.PendingStateHash)
	require.Equal(t, paymentstypes.ComputeStateHash(closeState), debug.ComputedPendingStateHash)
	require.Equal(t, closeState.Nonce, debug.DisputedNonce)

	settlement, err := k.FinalizeSettlement(channel.ChannelID, 30)
	require.NoError(t, err)
	require.Equal(t, "245", keeperAmountFor(settlement.FinalBalances, alice))
	require.Equal(t, "250", keeperAmountFor(settlement.FinalBalances, bob))

	settlements, page, err := k.Settlements(nil)
	require.NoError(t, err)
	require.Zero(t, page.NextOffset)
	require.Equal(t, []paymentstypes.SettlementRecord{settlement}, settlements)
}

func TestKeeperStoreV2ParticipantChannelPagination(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x63)
	bob := keeperAddress(0x64)
	first := keeperSignedChannel(t, "keeper-store-v2-first", "100", alice, bob)
	second := keeperSignedChannel(t, "keeper-store-v2-second", "100", alice, bob)
	require.NoError(t, k.OpenChannel(first))
	require.NoError(t, k.OpenChannel(second))

	layout, err := k.StoreV2Layout()
	require.NoError(t, err)
	require.Len(t, layout.Channels, 2)
	entries, page, err := k.ParticipantChannels(alice, &prototype.PageRequest{Limit: 1})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, uint64(1), page.NextOffset)
	next, page, err := k.ParticipantChannels(alice, &prototype.PageRequest{Offset: page.NextOffset, Limit: 1})
	require.NoError(t, err)
	require.Len(t, next, 1)
	require.Zero(t, page.NextOffset)
	require.NotEqual(t, entries[0].ChannelID, next[0].ChannelID)
}

func TestKeeperAdaptiveSyncSnapshotRecovery(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x65)
	bob := keeperAddress(0x66)
	channel := keeperSignedChannel(t, "keeper-adaptive-sync", "100", alice, bob)
	require.NoError(t, k.OpenChannel(channel))
	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, closeState, alice, 20, "0"))
	newer := keeperSignedState(t, channel, 3, closeState.StateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "35"},
		{Participant: bob, Amount: "65"},
	})
	require.NoError(t, k.DisputeClose(channel.ChannelID, newer, bob, 21))

	snapshot, err := k.AdaptiveSyncSnapshot(22)
	require.NoError(t, err)
	require.Len(t, snapshot.ActiveDisputes, 1)
	recovered, err := k.RecoverAdaptiveSyncSafety(snapshot)
	require.NoError(t, err)
	require.Contains(t, recovered.ActiveDisputeChannelIDs, channel.ChannelID)
	require.Contains(t, recovered.PendingFinalizationIDs, channel.ChannelID)
}

func TestKeeperValidatorAssistedWatchDispute(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x67)
	bob := keeperAddress(0x68)
	validator := keeperAddress(0x69)
	service := keeperAddress(0x6a)
	channel := keeperSignedChannel(t, "keeper-validator-watch", "100", alice, bob)
	require.NoError(t, k.OpenChannel(channel))
	require.NoError(t, k.RegisterValidatorPaymentService(paymentstypes.ValidatorPaymentServiceMetadata{
		ValidatorAddress: validator,
		ServiceAddress:   service,
		WatchEndpoint:    "https://validator.example/watch",
		MinDelegation:    "10",
		Active:           true,
		UpdatedHeight:    10,
	}))
	require.NoError(t, k.RegisterValidatorWatchService(paymentstypes.ValidatorWatchRegistration{
		ValidatorAddress: validator,
		Delegator:        bob,
		RegisteredHeight: 11,
	}))
	stale := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	newer := keeperSignedState(t, channel, 3, stale.StateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "45"},
		{Participant: bob, Amount: "55"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, stale, alice, 20, "0"))
	require.NoError(t, k.SubmitValidatorAssistedDispute(paymentstypes.ValidatorAssistedDisputeSubmission{
		ValidatorAddress:      validator,
		ServiceAddress:        service,
		Delegator:             bob,
		ChannelID:             channel.ChannelID,
		ClosingStateReference: stale.StateHash,
		NewerState:            newer,
		CurrentHeight:         21,
		EvidenceHash:          paymentstypes.HashParts("keeper-validator-watch", channel.ChannelID, newer.StateHash),
	}))
	debug, err := k.QueryStateHash(channel.ChannelID)
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, debug.PendingStateHash)
}

func keeperSignedChannel(t *testing.T, salt, collateral, left, right string) paymentstypes.ChannelRecord {
	t.Helper()

	channel := paymentstypes.ChannelRecord{
		ChainID:             "aetheris-test-1",
		ChannelID:           paymentstypes.HashParts(salt, left, right),
		ChannelType:         paymentstypes.ChannelTypeBidirectional,
		Participants:        []string{left, right},
		Denom:               paymentstypes.NativeDenom,
		Collateral:          collateral,
		OpenHeight:          10,
		CloseDelay:          8,
		DisputePeriod:       8,
		OpeningFeePaid:      paymentstypes.DefaultOpeningFee,
		ConditionalPayments: true,
		CustodyDenom:        paymentstypes.NativeDenom,
		CustodyAmount:       collateral,
		Status:              paymentstypes.ChannelStatusOpen,
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
