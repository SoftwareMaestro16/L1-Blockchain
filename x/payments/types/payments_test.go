package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestPaymentChannelCloseDisputeFraudAndSettlement(t *testing.T) {
	alice := testAddress(0x11)
	bob := testAddress(0x22)
	channel := signedChannel(t, "channel-main", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "10")
	require.NoError(t, err)

	newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "350"},
		{Participant: bob, Amount: "650"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newerState, bob, 25)
	require.NoError(t, err)

	conflicting := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	proof := FraudProof{
		ProofID:         HashParts("fraud", channel.ChannelID),
		ProofType:       FraudProofTypeDoubleSign,
		SubmittedBy:     bob,
		OffendingSigner: alice,
		StateA:          newerState,
		StateB:          conflicting,
		PenaltyAmount:   "25",
		EvidenceHash:    HashParts("evidence", newerState.StateHash, conflicting.StateHash),
	}
	state, err = SubmitFraudProof(state, channel.ChannelID, proof, 26)
	require.NoError(t, err)

	state, settlement, err := FinalizeSettlement(state, channel.ChannelID, 50)
	require.NoError(t, err)
	require.NoError(t, settlement.ValidateForChannel(state.Channels[0]))
	require.Equal(t, "315", amountFor(settlement.FinalBalances, alice))
	require.Equal(t, "675", amountFor(settlement.FinalBalances, bob))
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, newerState.Nonce, state.Channels[0].FinalizedNonce)
}

func TestPaymentStateRejectsNonNaetAndCollateralMismatch(t *testing.T) {
	alice := testAddress(0x31)
	bob := testAddress(0x32)
	channel := signedChannel(t, "bad-denom", "1000", alice, bob)
	channel.Denom = "uatom"
	err := channel.Validate()
	require.ErrorContains(t, err, "naet")

	channel = signedChannel(t, "bad-collateral", "1000", alice, bob)
	badState, err := BuildState(ChannelState{
		ChainID:           channel.ChainID,
		ChannelID:         channel.ChannelID,
		ChannelType:       channel.ChannelType,
		Denom:             channel.Denom,
		Epoch:             1,
		Nonce:             2,
		PreviousStateHash: channel.OpeningStateHash,
		Balances: []Balance{
			{Participant: alice, Amount: "999"},
			{Participant: bob, Amount: "0"},
		},
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(badState, signer)
		require.NoError(t, err)
		badState.Signatures = append(badState.Signatures, sig)
	}
	channel.LatestState = badState
	err = channel.LatestState.ValidateForChannel(channel, true)
	require.ErrorContains(t, err, "conserve")
}

func TestRoutePaymentAndVirtualChannelUseExistingLiquidity(t *testing.T) {
	alice := testAddress(0x41)
	router := testAddress(0x42)
	bob := testAddress(0x43)
	first := signedChannel(t, "route-1", "700", alice, router)
	second := signedChannel(t, "route-2", "700", bob, router)

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)

	path, err := RoutePayment(state, alice, bob, "250", 10, 4)
	require.NoError(t, err)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{path[0].ChannelID, path[1].ChannelID})

	vc := VirtualChannel{
		VirtualChannelID: HashParts("virtual", alice, bob),
		ParentChannelIDs: []string{first.ChannelID, second.ChannelID},
		Endpoints:        []string{alice, bob},
		Capacity:         "250",
		ExpiresHeight:    100,
	}
	vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	state, err = OpenVirtualChannel(state, vc)
	require.NoError(t, err)
	require.Len(t, state.VirtualChannels, 1)
}

func TestUntrustedTopologyIsRejectedBeforeRouteUse(t *testing.T) {
	alice := testAddress(0x49)
	router := testAddress(0x4a)
	bob := testAddress(0x4b)
	channel := signedChannel(t, "verified-route", "500", alice, router)

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	_, err = RegisterRoutingEdge(state, ChannelEdge{
		ChannelID: HashParts("unknown-channel"),
		From:      alice,
		To:        bob,
		Capacity:  "100",
		FeeAmount: "1",
		Active:    true,
	})
	require.ErrorContains(t, err, "open channel")

	_, err = RegisterRoutingEdge(state, ChannelEdge{
		ChannelID: channel.ChannelID,
		From:      alice,
		To:        bob,
		Capacity:  "100",
		FeeAmount: "1",
		Active:    true,
	})
	require.ErrorContains(t, err, "participants")

	untrusted := state
	untrusted.Edges = append(untrusted.Edges, ChannelEdge{
		ChannelID: HashParts("gossip-only"),
		From:      alice,
		To:        bob,
		Capacity:  "100",
		FeeAmount: "1",
		Active:    true,
	})
	_, err = RoutePayment(untrusted, alice, bob, "50", 10, 4)
	require.ErrorContains(t, err, "unknown channel")
}

func TestSettlementPrunesRoutingEdgesForAuthoritativeClosedChannel(t *testing.T) {
	alice := testAddress(0x4c)
	bob := testAddress(0x4d)
	channel := signedChannel(t, "settlement-prunes-edge", "500", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: channel.ChannelID, From: alice, To: bob, Capacity: "100", FeeAmount: "1", Active: true})
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "250"},
		{Participant: bob, Amount: "250"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	state, _, err = FinalizeSettlement(state, channel.ChannelID, 40)
	require.NoError(t, err)
	require.Empty(t, state.Edges)

	_, err = RoutePayment(state, alice, bob, "50", 41, 4)
	require.ErrorContains(t, err, "route not found")
}

func TestSettlementBatchRequiresIndependentChannels(t *testing.T) {
	alice := testAddress(0x51)
	bob := testAddress(0x52)
	first := signedChannel(t, "batch-1", "100", alice, bob)
	second := signedChannel(t, "batch-2", "100", alice, bob)
	ops := []SettlementOperation{
		{OperationID: HashParts("op", "second"), OperationType: BatchOperationSettle, ChannelID: second.ChannelID, Nonce: 1, StateHash: second.LatestState.StateHash},
		{OperationID: HashParts("op", "first"), OperationType: BatchOperationClose, ChannelID: first.ChannelID, Nonce: 1, StateHash: first.LatestState.StateHash},
	}

	batch, err := NewSettlementBatch(HashParts("batch"), ops)
	require.NoError(t, err)
	require.Less(t, batch.Operations[0].ChannelID, batch.Operations[1].ChannelID)

	batch.Operations = append(batch.Operations, SettlementOperation{
		OperationID:   HashParts("op", "duplicate"),
		OperationType: BatchOperationDispute,
		ChannelID:     first.ChannelID,
		Nonce:         2,
		StateHash:     first.LatestState.StateHash,
	})
	batch.RootHash = ComputeBatchRoot(batch.Operations)
	require.ErrorContains(t, batch.Validate(), "independent")
}

func signedChannel(t *testing.T, salt, collateral, left, right string) ChannelRecord {
	t.Helper()

	channelID := HashParts(salt, left, right)
	channel := ChannelRecord{
		ChainID:       "aetheris-test-1",
		ChannelID:     channelID,
		ChannelType:   ChannelTypeBidirectional,
		Participants:  []string{left, right},
		Denom:         NativeDenom,
		Collateral:    collateral,
		OpenHeight:    10,
		DisputePeriod: 8,
		Status:        ChannelStatusOpen,
	}
	openState := signedState(t, channel, 1, "", []Balance{
		{Participant: left, Amount: collateral},
		{Participant: right, Amount: "0"},
	})
	channel.LatestState = openState
	channel.OpeningStateHash = openState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func signedState(t *testing.T, channel ChannelRecord, nonce uint64, previous string, balances []Balance) ChannelState {
	t.Helper()

	state, err := BuildState(ChannelState{
		ChainID:           channel.ChainID,
		ChannelID:         channel.ChannelID,
		ChannelType:       channel.ChannelType,
		Denom:             channel.Denom,
		Epoch:             1,
		Nonce:             nonce,
		Balances:          balances,
		PreviousStateHash: previous,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func testAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(bytes20(fill)))
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}

func amountFor(balances []Balance, participant string) string {
	for _, balance := range balances {
		if balance.Participant == participant {
			return balance.Amount
		}
	}
	return ""
}
