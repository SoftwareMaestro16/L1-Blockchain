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
	require.Empty(t, state.CustodyLocks)
}

func TestSettlementArbitrationBoundaryRejectsNonDeterministicInputs(t *testing.T) {
	alice := testAddress(0x12)
	bob := testAddress(0x13)
	channel := signedChannel(t, "settlement-boundary", "1000", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})

	valid := SettlementArbitrationInput{
		Operation:     SettlementArbitrationUnilateralClose,
		ChannelID:     channel.ChannelID,
		SignedState:   closeState,
		CurrentHeight: 20,
	}
	require.NoError(t, valid.ValidateForChannel(channel))

	withRoute := valid
	withRoute.RouteHints = []ChannelEdge{{ChannelID: channel.ChannelID}}
	require.ErrorContains(t, withRoute.ValidateForChannel(channel), "must not select payment routes")

	withGossip := valid
	withGossip.GossipStateHash = HashParts("gossip", channel.ChannelID)
	require.ErrorContains(t, withGossip.ValidateForChannel(channel), "must not trust gossip")

	withLiquidity := valid
	withLiquidity.ExternalLiquidity = []Balance{{Participant: alice, Amount: "1000"}}
	require.ErrorContains(t, withLiquidity.ValidateForChannel(channel), "must not depend on external liquidity")

	withUnsignedBalance := valid
	withUnsignedBalance.UnsignedBalances = []Balance{{Participant: bob, Amount: "1000"}}
	require.ErrorContains(t, withUnsignedBalance.ValidateForChannel(channel), "must not accept unsigned balance")

	withIntent := valid
	withIntent.OffchainIntent = "alice verbally approved this close"
	require.ErrorContains(t, withIntent.ValidateForChannel(channel), "must not infer participant intent")
}

func TestSettlementArbitrationBoundaryRequiresSignedState(t *testing.T) {
	alice := testAddress(0x16)
	bob := testAddress(0x17)
	channel := signedChannel(t, "settlement-signed-state", "1000", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "300"},
		{Participant: bob, Amount: "700"},
	})
	closeState.Signatures = nil

	err := (SettlementArbitrationInput{
		Operation:     SettlementArbitrationUnilateralClose,
		ChannelID:     channel.ChannelID,
		SignedState:   closeState,
		CurrentHeight: 20,
	}).ValidateForChannel(channel)
	require.ErrorContains(t, err, "quorum")
}

func TestUnilateralCloseRequestStoresReasonAndDetachedSignatures(t *testing.T) {
	alice := testAddress(0x18)
	bob := testAddress(0x19)
	channel := signedChannel(t, "unilateral-close-request", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "425"},
		{Participant: bob, Amount: "575"},
	})
	detached := closeState.Signatures
	closeState.Signatures = nil
	state, err = SubmitCloseWithRequest(state, ChannelCloseRequest{
		ChannelID:     channel.ChannelID,
		ClosingState:  closeState,
		Signatures:    detached,
		CloseReason:   CloseReasonUnilateral,
		Submitter:     bob,
		CurrentHeight: 20,
		SettlementFee: "3",
	})
	require.NoError(t, err)
	require.Equal(t, ChannelStatusPendingClose, state.Channels[0].Status)
	require.Equal(t, CloseReasonUnilateral, state.Channels[0].PendingClose.CloseReason)
	require.Equal(t, uint64(20), state.Channels[0].PendingClose.SubmittedHeight)
	require.Equal(t, uint64(28), state.Channels[0].PendingClose.SettleAfterHeight)
	require.Equal(t, "3", state.Channels[0].PendingClose.SettlementFee)
}

func TestFraudProofInvalidBalanceRoutesPenaltyRemainder(t *testing.T) {
	alice := testAddress(0x1a)
	bob := testAddress(0x1b)
	channel := signedChannel(t, "fraud-penalty-routing", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	invalid := mutateCanonicalState(closeState, func(next *ChannelState) {
		next.Nonce = 3
		next.PreviousStateHash = closeState.StateHash
		next.Balances = []Balance{
			{Participant: alice, Amount: "900"},
			{Participant: bob, Amount: "900"},
		}
		next.BalanceA = "900"
		next.BalanceB = "900"
	})
	invalid = resignState(t, channel, invalid)
	require.ErrorContains(t, invalid.ValidateForChannel(channel, false), "conserve")

	proof := FraudProof{
		ProofID:         HashParts("invalid-balance", channel.ChannelID),
		ProofType:       FraudProofTypeInvalidBalance,
		SubmittedBy:     bob,
		OffendingSigner: alice,
		StateA:          invalid,
		PenaltyAmount:   "25",
		EvidenceHash:    HashParts("evidence", invalid.StateHash),
	}
	state, err = SubmitFraudProofWithPolicy(state, channel.ChannelID, proof, 21, FraudPenaltyPolicy{
		ReporterRewardCap:       "10",
		BurnShareBps:            5000,
		SecurityReserveShareBps: 2500,
		CommunityPoolShareBps:   2500,
	})
	require.NoError(t, err)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 1)
	require.Equal(t, "10", state.Channels[0].PendingClose.Penalties[0].Amount)
	require.Equal(t, "7", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteBurn))
	require.Equal(t, "3", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteSecurityReserve))
	require.Equal(t, "5", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteCommunityPool))

	state, settlement, err := FinalizeSettlement(state, channel.ChannelID, 50)
	require.NoError(t, err)
	require.Equal(t, "375", amountFor(settlement.FinalBalances, alice))
	require.Equal(t, "610", amountFor(settlement.FinalBalances, bob))
	require.Len(t, settlement.PenaltyAllocations, 3)
	require.NoError(t, settlement.ValidateForChannel(state.Channels[0]))
}

func TestSettlementFinalityTransitionsPendingHeightAndEvents(t *testing.T) {
	alice := testAddress(0x1c)
	bob := testAddress(0x1d)
	channel := signedChannel(t, "settlement-finality", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityOpen, state.Channels[0].Finality)
	require.Equal(t, "channel-finality-transition", state.Events[len(state.Events)-1].EventType)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "410"},
		{Participant: bob, Amount: "590"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPendingClose, state.Channels[0].Finality)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	height, found, err := state.PendingFinalizationHeight(channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(28), height)

	state, err = AdvanceChannelFinality(state, channel.ChannelID, 27)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPendingClose, state.Channels[0].Finality)

	state, err = AdvanceChannelFinality(state, channel.ChannelID, 28)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityFinalizable, state.Channels[0].Finality)
	require.Equal(t, "channel-finality-transition", state.Events[len(state.Events)-1].EventType)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	state, settlement, err := FinalizeSettlement(state, channel.ChannelID, 28)
	require.NoError(t, err)
	require.NoError(t, settlement.ValidateForChannel(state.Channels[0]))
	require.Equal(t, ChannelFinalitySettled, state.Channels[0].Finality)
	require.Empty(t, state.CustodyLocks)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
}

func TestDisputeAndPenaltyFinalityTransitionsRetainCollateralUntilSettlement(t *testing.T) {
	alice := testAddress(0x1e)
	bob := testAddress(0x1f)
	channel := signedChannel(t, "dispute-penalty-finality", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "350"},
		{Participant: bob, Amount: "650"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newerState, bob, 21)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityInDispute, state.Channels[0].Finality)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	conflicting := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	proof := FraudProof{
		ProofID:         HashParts("finality-proof", channel.ChannelID),
		ProofType:       FraudProofTypeDoubleSign,
		SubmittedBy:     bob,
		OffendingSigner: alice,
		StateA:          newerState,
		StateB:          conflicting,
		PenaltyAmount:   "20",
		EvidenceHash:    HashParts("evidence", newerState.StateHash, conflicting.StateHash),
	}
	state, err = SubmitFraudProof(state, channel.ChannelID, proof, 22)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPenalized, state.Channels[0].Finality)
	require.NotEmpty(t, state.CustodyLocks)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	state, _, err = FraudClose(state, channel.ChannelID, 23)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPenalized, state.Channels[0].Finality)
	require.Empty(t, state.CustodyLocks)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
}

func TestLockedCollateralInvariantForEveryFinalityState(t *testing.T) {
	alice := testAddress(0x20)
	bob := testAddress(0x21)
	base := signedChannel(t, "finality-invariant", "1000", alice, bob)
	lock := CustodyLock{ChannelID: base.ChannelID, Denom: NativeDenom, Amount: base.Collateral}

	cases := []struct {
		name     string
		finality ChannelFinality
		status   ChannelStatus
		locks    []CustodyLock
	}{
		{"open", ChannelFinalityOpen, ChannelStatusOpen, []CustodyLock{lock}},
		{"pending-close", ChannelFinalityPendingClose, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"in-dispute", ChannelFinalityInDispute, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"pending-condition", ChannelFinalityPendingConditionResolution, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"finalizable", ChannelFinalityFinalizable, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"pending-penalized", ChannelFinalityPenalized, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"expired", ChannelFinalityExpired, ChannelStatusOpen, []CustodyLock{lock}},
		{"settled", ChannelFinalitySettled, ChannelStatusSettled, nil},
		{"settled-penalized", ChannelFinalityPenalized, ChannelStatusSettled, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			channel := base
			channel.Status = tc.status
			channel.Finality = tc.finality
			require.NoError(t, ValidateLockedCollateralForFinality(PaymentsState{
				Channels:     []ChannelRecord{channel},
				CustodyLocks: tc.locks,
			}))
		})
	}

	missing := base
	missing.Finality = ChannelFinalityFinalizable
	missing.Status = ChannelStatusPendingClose
	require.ErrorContains(t, ValidateLockedCollateralForFinality(PaymentsState{Channels: []ChannelRecord{missing}}), "retain custody")
}

func TestConditionalPromiseObjectSignatureReserveAndReplayRules(t *testing.T) {
	alice := testAddress(0x22)
	bob := testAddress(0x23)
	channel := signedChannel(t, "promise-object", "1000", alice, bob)
	channel.LatestState = signedReserveState(t, channel, 2, channel.OpeningStateHash, "40", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "60"},
	})
	require.NoError(t, channel.Validate())

	promise := signedPromise(t, channel, "promise-a", alice, bob, "25", "5", 7, 40)
	require.NoError(t, promise.ValidateForChannel(channel))
	require.Equal(t, promise.PromiseID, promise.ToConditionalPayment().ConditionID)
	require.Equal(t, promise.Amount, promise.ToConditionalPayment().Amount)
	require.NoError(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise}, nil))

	duplicate := signedPromise(t, channel, "promise-a", alice, bob, "1", "0", 8, 40)
	require.ErrorContains(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise, duplicate}, nil), "duplicate promise")

	overReserve := signedPromise(t, channel, "promise-b", alice, bob, "10", "1", 9, 40)
	require.ErrorContains(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise, overReserve}, nil), "reserve")

	settled := []ConditionClaimRecord{{
		ChainID:        channel.ChainID,
		ChannelID:      channel.ChannelID,
		ConditionID:    promise.PromiseID,
		EvidenceHash:   HashParts("promise-settled", promise.PromiseID),
		ResolvedHeight: 50,
		ExpiresHeight:  100,
	}}
	require.ErrorContains(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise}, settled), "already been settled")

	late := signedPromise(t, channel, "promise-late", alice, bob, "1", "0", 10, channel.LatestState.TimeoutHeight)
	require.ErrorContains(t, late.ValidateForChannel(channel), "dispute window")

	wrongSigner := promise
	sig, err := SignatureForPromise(channel, wrongSigner, bob)
	require.NoError(t, err)
	wrongSigner.Signature = sig
	require.ErrorContains(t, wrongSigner.ValidateForChannel(channel), "signer must be source")
}

func TestHashLockedPreimageRevealResolvesLinkedPromisesAndTracksPreimage(t *testing.T) {
	alice := testAddress(0x24)
	bob := testAddress(0x25)
	channel := signedChannel(t, "promise-reveal", "1000", alice, bob)
	reserveState := signedReserveState(t, channel, 2, channel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	require.NoError(t, channel.Validate())

	preimage := "shared-secret"
	hashLock := HashParts(preimage)
	first := signedPromiseWithHashLock(t, channel, "reveal-a", alice, bob, "20", "1", 7, 40, hashLock)
	second := signedPromiseWithHashLock(t, channel, "reveal-b", alice, bob, "10", "1", 8, 40, hashLock)

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, channel.ChannelID, reserveState, 20)
	require.NoError(t, err)
	freshState := state.Clone()

	state, resolutions, err := RevealPromisePreimage(state, PreimageRevealRequest{
		ChannelID:     channel.ChannelID,
		Promises:      []ConditionalPromise{first, second},
		Preimage:      preimage,
		Revealer:      bob,
		CurrentHeight: 30,
	})
	require.NoError(t, err)
	require.Len(t, resolutions, 2)
	require.Equal(t, bob, resolutions[0].Recipient)
	require.Equal(t, first.PromiseID, resolutions[0].ConditionID)
	require.Len(t, state.ConditionClaims, 2)
	require.Equal(t, hashLock, state.ConditionClaims[0].PreimageHash)
	require.Equal(t, hashLock, state.ConditionClaims[1].PreimageHash)

	_, _, err = RevealPromisePreimage(state, PreimageRevealRequest{
		ChannelID:     channel.ChannelID,
		Promises:      []ConditionalPromise{first},
		Preimage:      preimage,
		Revealer:      bob,
		CurrentHeight: 31,
	})
	require.ErrorContains(t, err, "already been settled")

	_, _, err = RevealPromisePreimage(freshState.Clone(), PreimageRevealRequest{
		ChannelID:     channel.ChannelID,
		Promises:      []ConditionalPromise{first},
		Preimage:      "wrong-secret",
		Revealer:      bob,
		CurrentHeight: 30,
	})
	require.ErrorContains(t, err, "does not satisfy hash lock")

	_, _, err = RevealPromisePreimage(freshState.Clone(), PreimageRevealRequest{
		ChannelID:     channel.ChannelID,
		Promises:      []ConditionalPromise{first},
		Preimage:      preimage,
		Revealer:      bob,
		CurrentHeight: 41,
	})
	require.ErrorContains(t, err, "timed out")
}

func TestTimeoutOrderingAndExpiryResolutionReleaseConditionRoot(t *testing.T) {
	alice := testAddress(0x26)
	bob := testAddress(0x27)
	channel := signedChannel(t, "promise-expiry", "1000", alice, bob)
	base := signedReserveState(t, channel, 2, channel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	promiseChannel := channel
	promiseChannel.LatestState = base
	hashLock := HashParts("race-preimage")
	downstreamID := HashParts("promise", channel.ChannelID, "downstream")
	upstreamID := HashParts("promise", channel.ChannelID, "upstream")
	downstream := signedLinkedPromise(t, promiseChannel, downstreamID, alice, bob, "20", "1", 9, 40, hashLock, "", upstreamID)
	upstream := signedLinkedPromise(t, promiseChannel, upstreamID, alice, bob, "15", "1", 10, 70, hashLock, downstreamID, "")

	require.NoError(t, ValidatePromiseTimeoutOrdering(promiseChannel, upstream, downstream, DefaultTimeoutMargin))
	require.NoError(t, ValidatePromiseTimeoutChain(promiseChannel, []ConditionalPromise{downstream, upstream}, DefaultTimeoutMargin))
	require.ErrorContains(t, ValidatePromiseTimeoutOrdering(promiseChannel, upstream, downstream, 4), "margin")
	badUpstream := signedLinkedPromise(t, promiseChannel, HashParts("promise", channel.ChannelID, "bad-upstream"), alice, bob, "15", "1", 11, 50, hashLock, downstreamID, "")
	require.ErrorContains(t, ValidatePromiseTimeoutOrdering(promiseChannel, badUpstream, downstream, DefaultTimeoutMargin), "downstream timeout")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	conditioned, rootUpdate, err := BuildConditionRootUpdateFromPromises(promiseChannel, base, []ConditionalPromise{downstream, upstream}, nil)
	require.NoError(t, err)
	require.Equal(t, uint32(2), rootUpdate.ConditionCount)
	conditioned = resignState(t, channel, conditioned)
	state, err = AcceptSignedState(state, channel.ChannelID, conditioned, 20)
	require.NoError(t, err)

	_, _, _, err = ExpireConditionalPromises(state, PromiseExpiryRequest{
		ChannelID:     channel.ChannelID,
		Promises:      []ConditionalPromise{downstream},
		Resolver:      alice,
		CurrentHeight: downstream.TimeoutHeight,
	})
	require.ErrorContains(t, err, "has not expired")

	state, resolutions, expiryUpdate, err := ExpireConditionalPromises(state, PromiseExpiryRequest{
		ChannelID:     channel.ChannelID,
		Promises:      []ConditionalPromise{downstream},
		Resolver:      alice,
		CurrentHeight: downstream.TimeoutHeight + 1,
	})
	require.NoError(t, err)
	require.Len(t, resolutions, 1)
	require.True(t, resolutions[0].Expired)
	require.Equal(t, alice, resolutions[0].Recipient)
	require.Len(t, state.ConditionClaims, 1)
	require.Equal(t, uint32(1), expiryUpdate.ConditionCount)
	require.NotEqual(t, rootUpdate.ConditionRoot, expiryUpdate.ConditionRoot)

	_, _, err = RevealPromisePreimage(state, PreimageRevealRequest{
		ChannelID:     channel.ChannelID,
		Promises:      []ConditionalPromise{downstream},
		Preimage:      "race-preimage",
		Revealer:      bob,
		CurrentHeight: downstream.TimeoutHeight + 1,
	})
	require.ErrorContains(t, err, "timed out")
}

func TestBatchConditionSettlementAtomicallyResolvesChainedPromises(t *testing.T) {
	alice := testAddress(0x28)
	router := testAddress(0x29)
	bob := testAddress(0x2a)
	routeID := HashParts("route", alice, router, bob)
	hashLock := HashParts("atomic-preimage")
	firstChannel := signedChannel(t, "chain-first", "1000", alice, router)
	secondChannel := signedChannel(t, "chain-second", "1000", router, bob)
	firstID := HashParts("promise", routeID, "first")
	secondID := HashParts("promise", routeID, "second")

	firstBase := signedReserveState(t, firstChannel, 2, firstChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: router, Amount: "20"},
	})
	secondBase := signedReserveState(t, secondChannel, 2, secondChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: router, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	firstPromiseChannel := firstChannel
	firstPromiseChannel.LatestState = firstBase
	secondPromiseChannel := secondChannel
	secondPromiseChannel.LatestState = secondBase
	first := signedRoutePromise(t, firstPromiseChannel, firstID, routeID, alice, router, "31", "0", 9, 70, hashLock, "", secondID)
	second := signedRoutePromise(t, secondPromiseChannel, secondID, routeID, router, bob, "30", "1", 10, 40, hashLock, firstID, "")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, firstChannel)
	require.NoError(t, err)
	state, err = OpenChannel(state, secondChannel)
	require.NoError(t, err)
	firstConditioned, firstRoot, err := BuildConditionRootUpdateFromPromises(firstPromiseChannel, firstBase, []ConditionalPromise{first}, nil)
	require.NoError(t, err)
	secondConditioned, secondRoot, err := BuildConditionRootUpdateFromPromises(secondPromiseChannel, secondBase, []ConditionalPromise{second}, nil)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, firstChannel.ChannelID, resignState(t, firstChannel, firstConditioned), 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, secondChannel.ChannelID, resignState(t, secondChannel, secondConditioned), 20)
	require.NoError(t, err)

	proof := ConditionLinkageProof{
		RouteID:       routeID,
		Promises:      []ConditionalPromise{first, second},
		Sender:        alice,
		Receiver:      bob,
		Amount:        "30",
		TotalFees:     "1",
		HashLock:      hashLock,
		TimeoutMargin: DefaultTimeoutMargin,
	}
	state, result, err := BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof:  proof,
		Mode:          ConditionSettlementModePreimage,
		Preimage:      "atomic-preimage",
		Resolver:      bob,
		CurrentHeight: 30,
	})
	require.NoError(t, err)
	require.Len(t, result.Resolutions, 2)
	require.Len(t, result.FeeClaims, 1)
	require.Equal(t, router, result.FeeClaims[0].Recipient)
	require.Equal(t, "1", result.FeeClaims[0].Amount)
	require.Len(t, result.ConditionRootUpdates, 2)
	require.NotContains(t, []string{firstRoot.ConditionRoot, secondRoot.ConditionRoot}, result.ConditionRootUpdates[0].ConditionRoot)
	require.Len(t, state.ConditionClaims, 2)
	require.Equal(t, hashLock, state.ConditionClaims[0].PreimageHash)
	require.Equal(t, hashLock, state.ConditionClaims[1].PreimageHash)

	_, _, err = BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof:  proof,
		Mode:          ConditionSettlementModePreimage,
		Preimage:      "atomic-preimage",
		Resolver:      bob,
		CurrentHeight: 31,
	})
	require.ErrorContains(t, err, "already been settled")
}

func TestBatchConditionSettlementRejectsBrokenRouteInvariants(t *testing.T) {
	alice := testAddress(0x2b)
	router := testAddress(0x2c)
	bob := testAddress(0x2d)
	routeID := HashParts("route-bad", alice, router, bob)
	hashLock := HashParts("bad-route-preimage")
	firstChannel := signedChannel(t, "chain-bad-first", "1000", alice, router)
	secondChannel := signedChannel(t, "chain-bad-second", "1000", router, bob)
	openFirst := firstChannel
	openSecond := secondChannel
	firstBase := signedReserveState(t, firstChannel, 2, firstChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: router, Amount: "20"},
	})
	secondBase := signedReserveState(t, secondChannel, 2, secondChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: router, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	firstChannel.LatestState = firstBase
	secondChannel.LatestState = secondBase
	firstID := HashParts("promise", routeID, "first")
	secondID := HashParts("promise", routeID, "second")
	first := signedRoutePromise(t, firstChannel, firstID, routeID, alice, router, "30", "0", 9, 70, hashLock, "", secondID)
	second := signedRoutePromise(t, secondChannel, secondID, routeID, router, bob, "30", "1", 10, 40, hashLock, firstID, "")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, openFirst)
	require.NoError(t, err)
	state, err = OpenChannel(state, openSecond)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, openFirst.ChannelID, firstBase, 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, openSecond.ChannelID, secondBase, 20)
	require.NoError(t, err)
	require.ErrorContains(t, ConditionLinkageProof{
		RouteID:       routeID,
		Promises:      []ConditionalPromise{first, second},
		Sender:        alice,
		Receiver:      bob,
		Amount:        "30",
		TotalFees:     "1",
		HashLock:      hashLock,
		TimeoutMargin: DefaultTimeoutMargin,
	}.ValidateForState(state, nil), "amount conservation")

	firstConservedID := HashParts("promise", routeID, "first-conserved")
	badTimeoutID := HashParts("promise", routeID, "bad-timeout")
	firstConserved := signedRoutePromise(t, firstChannel, firstConservedID, routeID, alice, router, "31", "0", 12, 70, hashLock, "", badTimeoutID)
	badTimeout := signedRoutePromise(t, secondChannel, badTimeoutID, routeID, router, bob, "30", "1", 11, 60, hashLock, firstConservedID, "")
	require.ErrorContains(t, ConditionLinkageProof{
		RouteID:       routeID,
		Promises:      []ConditionalPromise{firstConserved, badTimeout},
		Sender:        alice,
		Receiver:      bob,
		Amount:        "30",
		TotalFees:     "1",
		HashLock:      hashLock,
		TimeoutMargin: DefaultTimeoutMargin,
	}.ValidateForState(state, nil), "downstream timeout")

	partialState, partialResult, err := BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof: ConditionLinkageProof{
			RouteID:                    routeID,
			Promises:                   []ConditionalPromise{second},
			Sender:                     router,
			Receiver:                   bob,
			Amount:                     "30",
			TotalFees:                  "0",
			HashLock:                   hashLock,
			TimeoutMargin:              DefaultTimeoutMargin,
			PartialDispute:             true,
			OffchainResolvedPromiseIDs: []string{firstID},
		},
		Mode:          ConditionSettlementModePreimage,
		Preimage:      "bad-route-preimage",
		Resolver:      bob,
		CurrentHeight: 30,
	})
	require.NoError(t, err)
	require.Len(t, partialResult.Resolutions, 1)
	require.Empty(t, partialResult.FeeClaims)
	require.Len(t, partialState.ConditionClaims, 1)
}

func TestBatchConditionSettlementExpiryIsAtomicWithoutFees(t *testing.T) {
	alice := testAddress(0x2e)
	router := testAddress(0x2f)
	bob := testAddress(0x30)
	routeID := HashParts("route-expiry", alice, router, bob)
	hashLock := HashParts("expiry-preimage")
	firstChannel := signedChannel(t, "chain-expiry-first", "1000", alice, router)
	secondChannel := signedChannel(t, "chain-expiry-second", "1000", router, bob)
	firstID := HashParts("promise", routeID, "first")
	secondID := HashParts("promise", routeID, "second")
	firstBase := signedReserveState(t, firstChannel, 2, firstChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: router, Amount: "20"},
	})
	secondBase := signedReserveState(t, secondChannel, 2, secondChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: router, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	firstPromiseChannel := firstChannel
	firstPromiseChannel.LatestState = firstBase
	secondPromiseChannel := secondChannel
	secondPromiseChannel.LatestState = secondBase
	first := signedRoutePromise(t, firstPromiseChannel, firstID, routeID, alice, router, "31", "0", 9, 70, hashLock, "", secondID)
	second := signedRoutePromise(t, secondPromiseChannel, secondID, routeID, router, bob, "30", "1", 10, 40, hashLock, firstID, "")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, firstChannel)
	require.NoError(t, err)
	state, err = OpenChannel(state, secondChannel)
	require.NoError(t, err)
	firstConditioned, _, err := BuildConditionRootUpdateFromPromises(firstPromiseChannel, firstBase, []ConditionalPromise{first}, nil)
	require.NoError(t, err)
	secondConditioned, _, err := BuildConditionRootUpdateFromPromises(secondPromiseChannel, secondBase, []ConditionalPromise{second}, nil)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, firstChannel.ChannelID, resignState(t, firstChannel, firstConditioned), 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, secondChannel.ChannelID, resignState(t, secondChannel, secondConditioned), 20)
	require.NoError(t, err)

	_, _, err = BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof: ConditionLinkageProof{
			RouteID:       routeID,
			Promises:      []ConditionalPromise{first, second},
			Sender:        alice,
			Receiver:      bob,
			Amount:        "30",
			TotalFees:     "1",
			HashLock:      hashLock,
			TimeoutMargin: DefaultTimeoutMargin,
		},
		Mode:          ConditionSettlementModeExpiry,
		Resolver:      alice,
		CurrentHeight: 70,
	})
	require.ErrorContains(t, err, "has not expired")

	state, result, err := BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof: ConditionLinkageProof{
			RouteID:       routeID,
			Promises:      []ConditionalPromise{first, second},
			Sender:        alice,
			Receiver:      bob,
			Amount:        "30",
			TotalFees:     "1",
			HashLock:      hashLock,
			TimeoutMargin: DefaultTimeoutMargin,
		},
		Mode:          ConditionSettlementModeExpiry,
		Resolver:      alice,
		CurrentHeight: 71,
	})
	require.NoError(t, err)
	require.Len(t, result.Resolutions, 2)
	require.Empty(t, result.FeeClaims)
	require.True(t, result.Resolutions[0].Expired)
	require.True(t, result.Resolutions[1].Expired)
	require.Len(t, result.ConditionRootUpdates, 2)
	require.Len(t, state.ConditionClaims, 2)
}

func TestDisputeRequestEmitsEventAndAppliesOptionalFraudProof(t *testing.T) {
	alice := testAddress(0x14)
	bob := testAddress(0x15)
	channel := signedChannel(t, "dispute-request", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "350"},
		{Participant: bob, Amount: "650"},
	})
	conflicting := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	state, err = DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:             channel.ChannelID,
		ClosingStateReference: closeState.StateHash,
		NewerState:            newerState,
		FraudProof: FraudProof{
			ProofID:         HashParts("dispute-proof", channel.ChannelID),
			ProofType:       FraudProofTypeDoubleSign,
			SubmittedBy:     bob,
			OffendingSigner: alice,
			StateA:          newerState,
			StateB:          conflicting,
			PenaltyAmount:   "25",
			EvidenceHash:    HashParts("evidence", newerState.StateHash, conflicting.StateHash),
		},
		Submitter:     bob,
		CurrentHeight: 25,
	})
	require.NoError(t, err)
	require.Equal(t, newerState.StateHash, state.Channels[0].PendingClose.State.StateHash)
	require.Len(t, state.Channels[0].PendingClose.FraudProofs, 1)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 1)
	require.Equal(t, "channel-dispute", state.Events[len(state.Events)-1].EventType)

	_, err = DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:             channel.ChannelID,
		ClosingStateReference: closeState.StateHash,
		NewerState:            newerState,
		Submitter:             bob,
		CurrentHeight:         26,
	})
	require.ErrorContains(t, err, "reference")
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
		Version:           CurrentStateVersion,
		Epoch:             1,
		Nonce:             2,
		PreviousStateHash: channel.OpeningStateHash,
		TimeoutHeight:     64,
		CloseDelay:        channel.DisputePeriod,
		FeePolicyID:       NativeDenom,
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

func TestChannelOpenLifecycleLocksFeeAndEmitsEvent(t *testing.T) {
	alice := testAddress(0x33)
	bob := testAddress(0x34)
	req := ChannelOpenRequest{
		ChainID:                      "aetheris-test-1",
		ChannelID:                    HashParts("open-lifecycle", alice, bob),
		Participants:                 []string{alice, bob},
		InitialBalances:              []Balance{{Participant: alice, Amount: "700"}, {Participant: bob, Amount: "300"}},
		ChannelType:                  ChannelTypeBidirectional,
		Collateral:                   "1000",
		CloseDelay:                   8,
		ChallengePeriod:              12,
		FeePolicyID:                  NativeDenom,
		OpeningFeeDenom:              NativeDenom,
		OpeningFeePaid:               DefaultOpeningFee,
		RoutingAdvertised:            true,
		ConditionalPaymentsSupported: true,
		OpenHeight:                   11,
	}

	state, event, err := OpenChannelFromRequest(EmptyState(), req)
	require.NoError(t, err)
	require.Len(t, state.Channels, 1)
	require.Len(t, state.CustodyLocks, 1)
	require.Len(t, state.Events, 2)
	require.Equal(t, event, state.Events[0])
	require.Equal(t, "channel-open", event.EventType)
	require.Equal(t, "channel-finality-transition", state.Events[1].EventType)
	require.Equal(t, ChannelFinalityOpen, state.Channels[0].Finality)
	require.Equal(t, req.ChannelID, state.CustodyLocks[0].ChannelID)
	require.Equal(t, "1000", state.CustodyLocks[0].Amount)
	require.Equal(t, DefaultOpeningFee, state.Channels[0].OpeningFeePaid)
	require.True(t, state.Channels[0].RoutingAdvertised)
	require.True(t, state.Channels[0].ConditionalPayments)
	require.Equal(t, uint64(8), state.Channels[0].CloseDelay)
	require.Equal(t, uint64(12), state.Channels[0].DisputePeriod)

	_, _, err = OpenChannelFromRequest(state, req)
	require.ErrorContains(t, err, "already exists")

	badFee := req
	badFee.ChannelID = HashParts("open-bad-fee", alice, bob)
	badFee.OpeningFeePaid = "0"
	_, _, err = OpenChannelFromRequest(EmptyState(), badFee)
	require.ErrorContains(t, err, "opening fee")

	badDelay := req
	badDelay.ChannelID = HashParts("open-bad-delay", alice, bob)
	badDelay.CloseDelay = 0
	_, _, err = OpenChannelFromRequest(EmptyState(), badDelay)
	require.ErrorContains(t, err, "close delay")

	badChallenge := req
	badChallenge.ChannelID = HashParts("open-bad-challenge", alice, bob)
	badChallenge.ChallengePeriod = MaxChallengePeriod + 1
	_, _, err = OpenChannelFromRequest(EmptyState(), badChallenge)
	require.ErrorContains(t, err, "challenge period")

	badBalances := req
	badBalances.ChannelID = HashParts("open-bad-balances", alice, bob)
	badBalances.InitialBalances = []Balance{{Participant: alice, Amount: "999"}, {Participant: bob, Amount: "0"}}
	_, _, err = OpenChannelFromRequest(EmptyState(), badBalances)
	require.ErrorContains(t, err, "sum to collateral")
}

func TestBidirectionalStateCommitmentIncludesDomainFields(t *testing.T) {
	alice := testAddress(0x37)
	bob := testAddress(0x38)
	channel := signedChannel(t, "bidirectional-domain", "1000", alice, bob)
	channel = channel.Normalize()

	condition := ConditionalPayment{
		ConditionID:   HashParts("condition", channel.ChannelID),
		ConditionType: ConditionTypeHashLock,
		Payer:         channel.Participants[0],
		Payee:         channel.Participants[1],
		Amount:        "40",
		HashLock:      HashParts("preimage"),
		TimeoutHeight: 88,
		NonceStart:    2,
		NonceEnd:      5,
	}
	state, err := BuildState(ChannelState{
		ChainID:           channel.ChainID,
		ChannelID:         channel.ChannelID,
		ChannelType:       channel.ChannelType,
		Denom:             channel.Denom,
		Version:           CurrentStateVersion,
		Epoch:             1,
		Nonce:             2,
		PreviousStateHash: channel.OpeningStateHash,
		Balances:          []Balance{{Participant: channel.Participants[0], Amount: "460"}, {Participant: channel.Participants[1], Amount: "500"}},
		ReserveA:          "25",
		ReserveB:          "15",
		Conditions:        []ConditionalPayment{condition},
		TimeoutHeight:     96,
		CloseDelay:        channel.DisputePeriod,
		FeePolicyID:       NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	require.Equal(t, ComputeConditionsRoot(state.Conditions), state.PendingConditionsRoot)

	changedTimeout := state
	changedTimeout.TimeoutHeight++
	changedTimeout.StateHash = ""
	changedTimeout.Signatures = nil
	changedTimeout, err = BuildState(changedTimeout)
	require.NoError(t, err)
	require.NotEqual(t, state.StateHash, changedTimeout.StateHash)

	changedReserve := state
	changedReserve.ReserveB = "16"
	changedReserve.StateHash = ""
	changedReserve.Signatures = nil
	changedReserve, err = BuildState(changedReserve)
	require.NoError(t, err)
	require.NotEqual(t, state.StateHash, changedReserve.StateHash)

	badRoot := state
	badRoot.PendingConditionsRoot = HashParts("wrong-root")
	badRoot.StateHash = ComputeStateHash(badRoot)
	for i := range badRoot.Signatures {
		sig, err := SignatureForState(badRoot, badRoot.Signatures[i].Signer)
		require.NoError(t, err)
		badRoot.Signatures[i] = sig
	}
	require.ErrorContains(t, badRoot.ValidateForChannel(channel, true), "conditions root")
}

func TestCanonicalChannelStateIncludesAllStateDomains(t *testing.T) {
	alice := testAddress(0x3b)
	bob := testAddress(0x3c)
	channel := signedChannel(t, "canonical-state", "1000", alice, bob)
	channel = channel.Normalize()

	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "475"},
		{Participant: bob, Amount: "525"},
	})

	require.Equal(t, CurrentAppVersion, state.AppVersion)
	require.Equal(t, ModuleName, state.ModuleName)
	require.Equal(t, ComputeParticipantSetHash(channel.Participants), state.ParticipantSetHash)
	require.Equal(t, "0", state.AccruedFees)
	require.Equal(t, channel.DisputePeriod, state.ChallengePeriod)
	require.Equal(t, ComputeConditionsRoot(state.Conditions), state.ConditionRoot)
	require.Equal(t, state.PendingConditionsRoot, state.ConditionRoot)
	require.Equal(t, uint32(len(state.Conditions)), state.ConditionCount)
	require.Equal(t, ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners), state.RequiredSignerBitmap)
	require.Equal(t, SignatureSchemeEd25519, state.SignatureScheme)
	require.Equal(t, ComputeStateSignaturePreimageHash(state), state.SignaturePreimageHash)

	changedFees := state
	changedFees.AccruedFees = "1"
	changedFees.StateHash = ""
	changedFees.Signatures = nil
	changedFees, err := BuildState(changedFees)
	require.NoError(t, err)
	require.NotEqual(t, state.StateHash, changedFees.StateHash)

	badParticipantSet := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.ParticipantSetHash = HashParts("wrong-participant-set")
	}))
	require.ErrorContains(t, badParticipantSet.ValidateForChannel(channel, true), "participant set")

	badChallenge := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.ChallengePeriod++
	}))
	require.ErrorContains(t, badChallenge.ValidateForChannel(channel, true), "challenge period")

	badScheme := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.SignatureScheme = "secp256k1"
	}))
	require.ErrorContains(t, badScheme.ValidateForChannel(channel, true), "signature scheme")
}

func TestStateHashEncodingVersionAndDomainSeparation(t *testing.T) {
	alice := testAddress(0x3f)
	bob := testAddress(0x40)
	channel := signedChannel(t, "state-hash-version", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	v1Hash, err := ComputeStateHashForEncodingVersion(state, CanonicalEncodingVersion)
	require.NoError(t, err)
	require.Equal(t, state.StateHash, v1Hash)
	_, err = ComputeStateHashForEncodingVersion(state, CanonicalEncodingVersion+1)
	require.ErrorContains(t, err, "unsupported")

	condition := ConditionalPayment{
		ConditionID:   HashParts("promise", channel.ChannelID),
		ConditionType: ConditionTypeHashLock,
		Payer:         alice,
		Payee:         bob,
		Amount:        "10",
		HashLock:      HashParts("promise-preimage"),
		TimeoutHeight: 64,
		NonceStart:    2,
		NonceEnd:      3,
	}
	async := signedAsyncChannel(t, "domain-async", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	delta := signedAsyncDelta(t, async, "domain-delta", alice, bob, "5", 2, 2, 70)
	conflicting := signedState(t, channel, state.Nonce, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "510"},
		{Participant: bob, Amount: "490"},
	})
	proof := FraudProof{
		ProofID:         HashParts("domain-proof", channel.ChannelID),
		ProofType:       FraudProofTypeDoubleSign,
		SubmittedBy:     bob,
		OffendingSigner: alice,
		StateA:          state,
		StateB:          conflicting,
		PenaltyAmount:   "10",
		EvidenceHash:    HashParts("evidence", state.StateHash, conflicting.StateHash),
	}
	vc := VirtualChannel{
		VirtualChannelID: HashParts("domain-vc", alice, bob),
		ChainID:          channel.ChainID,
		Nonce:            1,
		ParentChannelIDs: []string{channel.ChannelID},
		Endpoints:        []string{alice, bob},
		Capacity:         "100",
		ExpiresHeight:    90,
	}
	vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	vc.StateHash = ComputeVirtualChannelStateHash(vc)

	hashes := []string{
		state.StateHash,
		delta.DeltaHash,
		ComputeConditionalPromiseHash(condition),
		ComputeCooperativeCloseHash(channel.ChainID, channel.ChannelID, state.StateHash, state.Nonce),
		ComputeDisputeProofHash(proof),
		vc.StateHash,
	}
	seen := map[string]struct{}{}
	for _, hash := range hashes {
		require.NotContains(t, seen, hash)
		seen[hash] = struct{}{}
	}
}

func TestStateRejectsUnknownRequiredFields(t *testing.T) {
	alice := testAddress(0x41)
	bob := testAddress(0x42)
	channel := signedChannel(t, "unknown-required-field", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	bad := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.RequiredFields = append(s.RequiredFields, "future_required_field")
	}))
	require.ErrorContains(t, bad.ValidateForChannel(channel, true), "unknown required field")
}

func TestCommitmentModelBindsChannelDomainAndPayloads(t *testing.T) {
	alice := testAddress(0x59)
	bob := testAddress(0x5a)
	first := signedChannel(t, "commitment-first", "1000", alice, bob)
	second := signedChannel(t, "commitment-second", "1000", alice, bob)

	firstState := signedConditionalState(t, first, 2, first.OpeningStateHash, "25", []Balance{
		{Participant: alice, Amount: "975"},
		{Participant: bob, Amount: "0"},
	})
	secondState := firstState
	secondState.ChannelID = second.ChannelID
	secondState.PreviousStateHash = second.OpeningStateHash
	secondState.ParticipantSetHash = ComputeParticipantSetHash(second.Participants)
	secondState.StateHash = ""
	secondState.Signatures = nil
	var err error
	secondState, err = BuildState(secondState)
	require.NoError(t, err)

	require.NotEqual(t, ComputeOpeningCommitment(first), ComputeOpeningCommitment(second))
	require.NotEqual(t, ComputeBalanceStateCommitment(first, firstState), ComputeBalanceStateCommitment(second, secondState))
	require.NotEqual(t, ComputeConditionRootCommitment(first, firstState), ComputeConditionRootCommitment(second, secondState))
	require.Equal(t, ComputeConditionsRoot(firstState.Conditions), firstState.ConditionRoot)

	asyncFirst := signedAsyncChannel(t, "commitment-async-first", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	asyncSecond := signedAsyncChannel(t, "commitment-async-second", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	delta := signedAsyncDelta(t, asyncFirst, "commitment-delta", alice, bob, "10", 2, 2, 70)
	require.NotEqual(t, ComputeAsyncDeltaRootForChannel(asyncFirst, []AsyncPaymentDelta{delta}), ComputeAsyncDeltaRootForChannel(asyncSecond, []AsyncPaymentDelta{delta}))

	vc := VirtualChannel{
		VirtualChannelID: HashParts("commitment-vc", alice, bob),
		ChainID:          first.ChainID,
		Nonce:            1,
		ParentChannelIDs: []string{first.ChannelID},
		Endpoints:        []string{alice, bob},
		Capacity:         "100",
		ExpiresHeight:    90,
	}
	vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	changedVC := vc
	changedVC.Capacity = "101"
	require.NotEqual(t, vc.AnchorCommitment, ComputeVirtualChannelAnchor(changedVC))
	changedVC = vc
	changedVC.ExpiresHeight++
	require.NotEqual(t, vc.AnchorCommitment, ComputeVirtualChannelAnchor(changedVC))

	settlement := SettlementRecord{
		ChainID:            first.ChainID,
		ChannelID:          first.ChannelID,
		StateHash:          firstState.StateHash,
		Nonce:              firstState.Nonce,
		FinalBalances:      firstState.Balances,
		SettlementFeeDenom: NativeDenom,
		SettlementFee:      "0",
		Penalties:          []Penalty{{Offender: alice, Recipient: bob, Denom: NativeDenom, Amount: "1"}},
		SettledHeight:      100,
	}
	penaltyRoute := settlement
	penaltyRoute.Penalties = []Penalty{{Offender: bob, Recipient: alice, Denom: NativeDenom, Amount: "1"}}
	otherDomain := settlement
	otherDomain.ChannelID = second.ChannelID
	otherDomain.ChainID = "aetheris-test-2"
	require.NotEqual(t, ComputeSettlementResultCommitment(first, settlement), ComputeSettlementResultCommitment(first, penaltyRoute))
	require.NotEqual(t, ComputeSettlementHash(settlement), ComputeSettlementHash(otherDomain))
}

func TestSignatureEnvelopeRejectsReplayAndWrongCommitment(t *testing.T) {
	alice := testAddress(0x5b)
	bob := testAddress(0x5c)
	channel := signedChannel(t, "signature-envelope", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	wrongChain := state
	wrongChain.Signatures = append([]StateSignature(nil), state.Signatures...)
	wrongChain.Signatures[0].ChainID = "other-chain"
	wrongChain.Signatures[0].SignatureHash = ComputeSignatureEnvelopeHash(
		wrongChain.Signatures[0].Signer,
		wrongChain.Signatures[0].ChainID,
		wrongChain.Signatures[0].ChannelID,
		wrongChain.Signatures[0].ObjectType,
		wrongChain.Signatures[0].Version,
		wrongChain.Signatures[0].Nonce,
		wrongChain.Signatures[0].ObjectID,
		wrongChain.Signatures[0].ExpirationHeight,
		wrongChain.Signatures[0].CommitmentHash,
	)
	require.ErrorContains(t, wrongChain.ValidateForChannel(channel, true), "chain id")

	wrongCommitment := state
	wrongCommitment.Signatures = append([]StateSignature(nil), state.Signatures...)
	wrongCommitment.Signatures[0].CommitmentHash = HashParts("wrong-commitment")
	wrongCommitment.Signatures[0].SignatureHash = ComputeSignatureEnvelopeHash(
		wrongCommitment.Signatures[0].Signer,
		wrongCommitment.Signatures[0].ChainID,
		wrongCommitment.Signatures[0].ChannelID,
		wrongCommitment.Signatures[0].ObjectType,
		wrongCommitment.Signatures[0].Version,
		wrongCommitment.Signatures[0].Nonce,
		wrongCommitment.Signatures[0].ObjectID,
		wrongCommitment.Signatures[0].ExpirationHeight,
		wrongCommitment.Signatures[0].CommitmentHash,
	)
	require.ErrorContains(t, wrongCommitment.ValidateForChannel(channel, true), "commitment")

	duplicate := state
	duplicate.Signatures = append([]StateSignature(nil), state.Signatures...)
	duplicate.Signatures = append(duplicate.Signatures, duplicate.Signatures[0])
	duplicate.Signatures = normalizeSignatures(duplicate.Signatures)
	require.ErrorContains(t, duplicate.ValidateForChannel(channel, true), "duplicate")
}

func TestClaimAndDeltaSignatureEnvelopeValidation(t *testing.T) {
	payer := testAddress(0x5d)
	receiver := testAddress(0x5e)
	channel := signedUnidirectionalChannel(t, "claim-envelope", "1000", payer, receiver, false)
	claim := signedUnidirectionalClaim(t, channel, "100", 2, 80, false)

	wrongChannel := claim
	wrongChannel.PayerSignature.ChannelID = HashParts("wrong-channel")
	wrongChannel.PayerSignature.SignatureHash = ComputeSignatureEnvelopeHash(
		wrongChannel.PayerSignature.Signer,
		wrongChannel.PayerSignature.ChainID,
		wrongChannel.PayerSignature.ChannelID,
		wrongChannel.PayerSignature.ObjectType,
		wrongChannel.PayerSignature.Version,
		wrongChannel.PayerSignature.Nonce,
		wrongChannel.PayerSignature.ObjectID,
		wrongChannel.PayerSignature.ExpirationHeight,
		wrongChannel.PayerSignature.CommitmentHash,
	)
	require.ErrorContains(t, wrongChannel.ValidateForChannel(channel), "channel id")

	alice := testAddress(0x5f)
	bob := testAddress(0x60)
	async := signedAsyncChannel(t, "delta-envelope", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	delta := signedAsyncDelta(t, async, "delta-envelope", alice, bob, "10", 2, 2, 70)

	wrongType := delta
	wrongType.Signature.ObjectType = SignatureObjectClaim
	wrongType.Signature.SignatureHash = ComputeSignatureEnvelopeHash(
		wrongType.Signature.Signer,
		wrongType.Signature.ChainID,
		wrongType.Signature.ChannelID,
		wrongType.Signature.ObjectType,
		wrongType.Signature.Version,
		wrongType.Signature.Nonce,
		wrongType.Signature.ObjectID,
		wrongType.Signature.ExpirationHeight,
		wrongType.Signature.CommitmentHash,
	)
	require.ErrorContains(t, wrongType.ValidateForChannel(async, 30), "object type")
	require.ErrorContains(t, delta.ValidateForChannel(async, 71), "expired")
}

func TestLocalSignerWriteAheadPreventsDoubleSign(t *testing.T) {
	alice := testAddress(0x61)
	bob := testAddress(0x62)
	channel := signedChannel(t, "signer-wal", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	records, sig, err := SignStateWithWriteAhead(nil, state, alice, SignerIsolationHardware)
	require.NoError(t, err)
	require.Equal(t, alice, sig.Signer)
	require.Len(t, records, 1)
	require.True(t, records[0].Released)
	require.Equal(t, SignerIsolationHardware, records[0].IsolationMode)
	require.Equal(t, ComputeSignedNonceWALHash(records[0]), records[0].WALHash)

	records, _, err = SignStateWithWriteAhead(records, state, alice, SignerIsolationHardware)
	require.NoError(t, err)
	require.Len(t, records, 1)

	next := signedState(t, channel, 3, state.StateHash, []Balance{
		{Participant: alice, Amount: "480"},
		{Participant: bob, Amount: "520"},
	})
	persistence := SignerPersistence{Records: records, IsolationMode: SignerIsolationHardware}
	persistence, _, err = persistence.SignState(next, alice)
	require.NoError(t, err)
	require.Equal(t, uint64(3), persistence.HighestSignedNonce(alice, channel.ChainID, channel.ChannelID, 1))
	_, _, err = SignStateWithWriteAhead(persistence.Records, state, alice, SignerIsolationHardware)
	require.ErrorContains(t, err, "below highest signed nonce")

	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "510"},
		{Participant: bob, Amount: "490"},
	})
	_, _, err = SignStateWithWriteAhead(records, conflicting, alice, SignerIsolationHardware)
	require.ErrorContains(t, err, "same nonce replacement")
}

func TestRollbackVectorsRejectNonceAndPreviousHashRollback(t *testing.T) {
	alice := testAddress(0x67)
	bob := testAddress(0x68)
	channel := signedChannel(t, "rollback-vectors", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	update := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})
	state, err = AcceptSignedState(state, channel.ChannelID, update, 18)
	require.NoError(t, err)

	lowerNonce := signedState(t, channel, 1, "", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	})
	_, err = AcceptSignedState(state, channel.ChannelID, lowerNonce, 19)
	require.ErrorContains(t, err, "strictly increase")

	wrongPrevious := signedState(t, channel, 3, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "480"},
		{Participant: bob, Amount: "520"},
	})
	require.ErrorContains(t, ValidatePreviousHashContinuity(state.Channels[0], wrongPrevious), "previous hash")
	_, err = AcceptSignedState(state, channel.ChannelID, wrongPrevious, 20)
	require.ErrorContains(t, err, "previous hash")
}

func TestDoubleSignFraudAppliesIndependentPenalties(t *testing.T) {
	alice := testAddress(0x63)
	bob := testAddress(0x64)
	channel := signedChannel(t, "both-double-sign", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "600"},
		{Participant: bob, Amount: "400"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	aliceProof := FraudProof{
		ProofID:         HashParts("double-sign-alice", channel.ChannelID),
		ProofType:       FraudProofTypeDoubleSign,
		SubmittedBy:     bob,
		OffendingSigner: alice,
		StateA:          closeState,
		StateB:          conflicting,
		PenaltyAmount:   "25",
		EvidenceHash:    HashParts("evidence", "alice", closeState.StateHash, conflicting.StateHash),
	}
	bobProof := aliceProof
	bobProof.ProofID = HashParts("double-sign-bob", channel.ChannelID)
	bobProof.SubmittedBy = alice
	bobProof.OffendingSigner = bob
	bobProof.EvidenceHash = HashParts("evidence", "bob", closeState.StateHash, conflicting.StateHash)

	state, err = SubmitFraudProof(state, channel.ChannelID, aliceProof, 21)
	require.NoError(t, err)
	state, err = SubmitFraudProof(state, channel.ChannelID, bobProof, 22)
	require.NoError(t, err)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 2)
	require.Equal(t, alice, state.Channels[0].PendingClose.Penalties[0].Offender)
	require.Equal(t, bob, state.Channels[0].PendingClose.Penalties[1].Offender)
}

func TestBidirectionalCloseAndUpdateRules(t *testing.T) {
	alice := testAddress(0x39)
	bob := testAddress(0x3a)
	channel := signedChannel(t, "bidirectional-lifecycle", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	update := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "475"},
		{Participant: bob, Amount: "525"},
	})
	oneSignature := update
	oneSignature.Signatures = oneSignature.Signatures[:1]
	require.ErrorContains(t, oneSignature.ValidateForChannel(channel, true), "quorum")

	state, err = AcceptSignedState(state, channel.ChannelID, update, 18)
	require.NoError(t, err)
	_, err = AcceptSignedState(state, channel.ChannelID, update, 19)
	require.ErrorContains(t, err, "strictly increase")

	staleClose := signedState(t, channel, 1, "", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	})
	_, err = SubmitClose(state, channel.ChannelID, staleClose, alice, 20, "0")
	require.ErrorContains(t, err, "latest accepted nonce")

	cooperative := signedState(t, channel, 3, update.StateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	state, settlement, err := CooperativeClose(state, channel.ChannelID, cooperative, bob, 21, "5")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Empty(t, state.Channels[0].PendingClose.State.StateHash)
	require.Equal(t, cooperative.Nonce, settlement.Nonce)
	require.Equal(t, "545", amountFor(settlement.FinalBalances, bob))
}

func TestChannelUpdateLifecycleValidatesOffchainAndRegistersCheckpoint(t *testing.T) {
	alice := testAddress(0x46)
	bob := testAddress(0x47)
	channel := signedChannel(t, "update-lifecycle", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	update := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "25", []Balance{
		{Participant: channel.Participants[0], Amount: "475"},
		{Participant: channel.Participants[1], Amount: "500"},
	})
	req := ChannelUpdateRequest{
		ChannelID:            channel.ChannelID,
		State:                update,
		ConditionCommitments: update.Conditions,
		Submitter:            alice,
		CurrentHeight:        18,
	}
	result, err := ValidateOffchainUpdate(channel, req)
	require.NoError(t, err)
	require.True(t, result.ValidatedOffChain)
	require.False(t, result.CheckpointRegistered)

	unchanged, result, err := RegisterUpdateCheckpoint(state, req)
	require.NoError(t, err)
	require.False(t, result.CheckpointRegistered)
	require.Equal(t, channel.LatestState.StateHash, unchanged.Channels[0].LatestState.StateHash)

	req.RegisterCheckpoint = true
	state, result, err = RegisterUpdateCheckpoint(state, req)
	require.NoError(t, err)
	require.True(t, result.CheckpointRegistered)
	require.Equal(t, update.StateHash, state.Channels[0].LatestState.StateHash)

	overReserve := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "30", []Balance{
		{Participant: channel.Participants[0], Amount: "475"},
		{Participant: channel.Participants[1], Amount: "500"},
	})
	_, err = ValidateOffchainUpdate(channel, ChannelUpdateRequest{
		ChannelID:            channel.ChannelID,
		State:                overReserve,
		ConditionCommitments: overReserve.Conditions,
		Submitter:            alice,
		CurrentHeight:        18,
	})
	require.ErrorContains(t, err, "reserve")
}

func TestAsyncUpdateBatchCanRegisterCheckpoint(t *testing.T) {
	alice := testAddress(0x48)
	bob := testAddress(0x49)
	channel := signedAsyncChannel(t, "async-update-lifecycle", "1000", []Balance{
		{Participant: alice, Amount: "700"},
		{Participant: bob, Amount: "300"},
	}, 4, 8, "100", 90, alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	delta := signedAsyncDelta(t, channel, "update-delta", alice, bob, "40", 2, 2, 80)
	checkpoint, err := BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{delta}, 3, 30)
	require.NoError(t, err)
	checkpoint = signAsyncCheckpoint(t, channel, checkpoint)
	state, result, err := RegisterUpdateCheckpoint(state, ChannelUpdateRequest{
		ChannelID:          channel.ChannelID,
		State:              checkpoint,
		AsyncDeltas:        []AsyncPaymentDelta{delta},
		RegisterCheckpoint: true,
		Submitter:          bob,
		CurrentHeight:      30,
	})
	require.NoError(t, err)
	require.True(t, result.CheckpointRegistered)
	require.Equal(t, checkpoint.StateHash, state.Channels[0].LatestState.StateHash)
}

func TestUnidirectionalReceiverCloseUsesSinglePayerSignature(t *testing.T) {
	payer := testAddress(0x3b)
	receiver := testAddress(0x3c)
	channel := signedUnidirectionalChannel(t, "uni-close", "1000", payer, receiver, false)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	claim := signedUnidirectionalClaim(t, channel, "320", 2, 80, false)
	require.Empty(t, claim.ReceiverAckOptional.SignatureHash)
	require.NoError(t, claim.ValidateForChannel(channel))

	badClaim := claim
	badClaim.PayerSignature, err = SignatureForClaim(badClaim, receiver)
	require.NoError(t, err)
	require.ErrorContains(t, badClaim.ValidateForChannel(channel), "payer signature")

	state, settlement, err := ReceiverClose(state, channel.ChannelID, claim, receiver, 30, "5")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, claim.Nonce, state.Channels[0].FinalizedNonce)
	require.Equal(t, claim.StateHash, settlement.StateHash)
	require.Equal(t, "680", amountFor(settlement.FinalBalances, payer))
	require.Equal(t, "315", amountFor(settlement.FinalBalances, receiver))
}

func TestUnidirectionalAcknowledgementModeAndPayerReclaim(t *testing.T) {
	payer := testAddress(0x3d)
	receiver := testAddress(0x3e)
	channel := signedUnidirectionalChannel(t, "uni-ack", "1000", payer, receiver, true)

	claim, err := BuildUnidirectionalClaim(UnidirectionalClaim{
		ChainID:             channel.ChainID,
		ChannelID:           channel.ChannelID,
		Payer:               payer,
		Receiver:            receiver,
		LockedAmount:        channel.Collateral,
		ClaimedAmount:       "125",
		Nonce:               2,
		ExpirationHeight:    80,
		ExpirationTimestamp: 0,
	})
	require.NoError(t, err)
	claim.PayerSignature, err = SignatureForClaim(claim, payer)
	require.NoError(t, err)
	require.ErrorContains(t, claim.ValidateForChannel(channel), "acknowledgement")
	claim.ReceiverAckOptional, err = SignatureForClaim(claim, receiver)
	require.NoError(t, err)
	require.NoError(t, claim.ValidateForChannel(channel))

	state := EmptyState()
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	_, _, err = PayerReclaim(state, channel.ChannelID, payer, channel.ExpirationHeight+channel.DisputePeriod, "3")
	require.ErrorContains(t, err, "dispute window")
	state, settlement, err := PayerReclaim(state, channel.ChannelID, payer, channel.ExpirationHeight+channel.DisputePeriod+1, "3")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, "997", amountFor(settlement.FinalBalances, payer))
	require.Equal(t, "0", amountFor(settlement.FinalBalances, receiver))
}

func TestUnidirectionalStreamingPaymentHelperFormat(t *testing.T) {
	payer := testAddress(0x3f)
	receiver := testAddress(0x40)
	channel := signedUnidirectionalChannel(t, "uni-stream", "1000", payer, receiver, false)

	claim, err := StreamingClaimForChannel(channel, StreamingPaymentFrame{
		ChannelID:           channel.ChannelID,
		StreamID:            HashParts("stream", channel.ChannelID),
		Payer:               payer,
		Receiver:            receiver,
		PreviousClaimed:     "10",
		RatePerBlock:        "5",
		StartHeight:         20,
		CurrentHeight:       32,
		Nonce:               2,
		ExpirationHeight:    90,
		ExpirationTimestamp: 0,
	})
	require.NoError(t, err)
	require.Equal(t, "70", claim.ClaimedAmount)
	claim.PayerSignature, err = SignatureForClaim(claim, payer)
	require.NoError(t, err)
	require.NoError(t, claim.ValidateForChannel(channel))
}

func TestAsyncCheckpointAggregationExposureExpiryAndProof(t *testing.T) {
	alice := testAddress(0x44)
	bob := testAddress(0x45)
	channel := signedAsyncChannel(t, "async-main", "1000", []Balance{
		{Participant: alice, Amount: "700"},
		{Participant: bob, Amount: "300"},
	}, 4, 8, "100", 90, alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	delta := signedAsyncDelta(t, channel, "delta-1", alice, bob, "40", 2, 2, 80)
	checkpoint, err := BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{delta}, 3, 30)
	require.NoError(t, err)
	checkpoint = signAsyncCheckpoint(t, channel, checkpoint)
	proof := AsyncDeltaDisputeProof{
		ProofID:         HashParts("proof", checkpoint.StateHash),
		ChannelID:       channel.ChannelID,
		CheckpointState: checkpoint,
		Deltas:          []AsyncPaymentDelta{delta},
		EvidenceHash:    HashParts("async-dispute", checkpoint.StateHash, ComputeAsyncDeltaRootForChannel(channel, []AsyncPaymentDelta{delta})),
	}
	require.NoError(t, proof.ValidateForChannel(channel, 30))

	state, err = AcceptAsyncCheckpoint(state, channel.ChannelID, checkpoint, []AsyncPaymentDelta{delta}, bob, 30)
	require.NoError(t, err)
	require.Equal(t, "660", amountFor(state.Channels[0].LatestState.Balances, alice))
	require.Equal(t, "340", amountFor(state.Channels[0].LatestState.Balances, bob))
	require.Equal(t, checkpoint.StateHash, state.Channels[0].LatestState.StateHash)

	badSigner := delta
	badSigner.Signature, err = SignatureForAsyncDelta(badSigner, bob)
	require.NoError(t, err)
	require.ErrorContains(t, badSigner.ValidateForChannel(channel, 30), "sender")

	tooMuch := []AsyncPaymentDelta{
		signedAsyncDelta(t, channel, "delta-2", alice, bob, "60", 3, 3, 80),
		signedAsyncDelta(t, channel, "delta-3", alice, bob, "50", 4, 4, 80),
	}
	_, err = BuildAsyncCheckpointState(channel, tooMuch, 5, 30)
	require.ErrorContains(t, err, "exposure")

	expired := signedAsyncDelta(t, channel, "delta-expired", bob, alice, "10", 3, 3, 31)
	_, err = BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{expired}, 4, 32)
	require.ErrorContains(t, err, "expired")

	badProof := proof
	badProof.Deltas = nil
	badProof.EvidenceHash = HashParts("async-dispute", checkpoint.StateHash, ComputeAsyncDeltaRootForChannel(channel, nil))
	require.ErrorContains(t, badProof.ValidateForChannel(channel, 30), "signed deltas")
}

func TestAsyncCheckpointRejectsDuplicateDeltaNonce(t *testing.T) {
	alice := testAddress(0x57)
	bob := testAddress(0x58)
	channel := signedAsyncChannel(t, "async-duplicate-nonce", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)

	first := signedAsyncDelta(t, channel, "first", alice, bob, "10", 2, 2, 70)
	second := signedAsyncDelta(t, channel, "second", alice, bob, "15", 2, 2, 70)
	_, err := BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{first, second}, 3, 30)
	require.ErrorContains(t, err, "duplicate async delta nonce")
}

func TestPaymentAssetScopeRejectsNonNaetFeesAndPenalties(t *testing.T) {
	alice := testAddress(0x35)
	bob := testAddress(0x36)
	channel := signedChannel(t, "asset-scope", "1000", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})

	edge := ChannelEdge{
		ChannelID: channel.ChannelID,
		From:      alice,
		To:        bob,
		Capacity:  "100",
		FeeDenom:  "uatom",
		FeeAmount: "1",
		Active:    true,
	}
	require.ErrorContains(t, edge.Validate(), "naet")

	pending := PendingClose{
		Submitter:          alice,
		SubmittedHeight:    20,
		SettleAfterHeight:  28,
		SettlementFeeDenom: "uatom",
		SettlementFee:      "1",
		State:              closeState,
	}
	require.ErrorContains(t, pending.ValidateForChannel(channel), "naet")

	proof := FraudProof{
		ProofID:         HashParts("bad-penalty-denom"),
		ProofType:       FraudProofTypeStaleClose,
		SubmittedBy:     bob,
		OffendingSigner: alice,
		StateA:          closeState,
		StateB: signedState(t, channel, 3, closeState.StateHash, []Balance{
			{Participant: alice, Amount: "450"},
			{Participant: bob, Amount: "550"},
		}),
		PenaltyDenom:  "uatom",
		PenaltyAmount: "10",
		EvidenceHash:  HashParts("evidence", "bad-penalty-denom"),
	}
	require.ErrorContains(t, proof.ValidateForChannel(channel), "naet")

	penalty := Penalty{Offender: alice, Recipient: bob, Denom: "uatom", Amount: "10"}
	require.ErrorContains(t, penalty.ValidateForChannel(channel), "naet")

	settlement := SettlementRecord{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		StateHash:          closeState.StateHash,
		Nonce:              closeState.Nonce,
		FinalBalances:      closeState.Balances,
		SettlementFeeDenom: "uatom",
		SettlementFee:      "0",
		SettledHeight:      40,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	require.ErrorContains(t, settlement.ValidateForChannel(channel), "naet")
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

func TestScoredRouteSelectionPenalizesFeeStaleLiquidityAndFailures(t *testing.T) {
	alice := testAddress(0x37)
	router := testAddress(0x38)
	bob := testAddress(0x39)
	direct := signedChannel(t, "score-direct", "1000", alice, bob)
	first := signedChannel(t, "score-first", "1000", alice, router)
	second := signedChannel(t, "score-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{direct, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: direct.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "20", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy.HopPenalty = "1"
	policy.StaleLiquidityAfter = 10
	policy.EdgeStats = []EdgeRoutingStats{
		{ChannelID: direct.ChannelID, From: alice, To: bob, SuccessRateBps: 3_000, LiquidityUpdatedHeight: 1, FailureCount: 4, CongestionBps: 7_500, NodeAvailabilityBps: 5_000, TimeoutMargin: 4},
		{ChannelID: first.ChannelID, From: alice, To: router, SuccessRateBps: 10_000, LiquidityUpdatedHeight: 95, NodeAvailabilityBps: 10_000, TimeoutMargin: DefaultTimeoutMargin},
		{ChannelID: second.ChannelID, From: router, To: bob, SuccessRateBps: 10_000, LiquidityUpdatedHeight: 95, NodeAvailabilityBps: 10_000, TimeoutMargin: DefaultTimeoutMargin},
	}
	route, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{
		From:          alice,
		To:            bob,
		Amount:        "100",
		CurrentHeight: 100,
		Policy:        policy,
	})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{route.Edges[0].ChannelID, route.Edges[1].ChannelID})
	require.Equal(t, "2", route.TotalFee)
	require.NotEmpty(t, route.ScoreHash)

	again, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 100, Policy: policy})
	require.NoError(t, err)
	require.Equal(t, route.ScoreHash, again.ScoreHash)
	sim, err := SimulateRoute(route, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 100, Policy: policy})
	require.NoError(t, err)
	require.True(t, sim.Attemptable)
}

func TestScoredRouteSelectionExcludesInsufficientCapacityAndMaxFee(t *testing.T) {
	alice := testAddress(0x3a)
	router := testAddress(0x3b)
	bob := testAddress(0x3c)
	low := signedChannel(t, "score-low-capacity", "1000", alice, bob)
	first := signedChannel(t, "score-cap-first", "1000", alice, router)
	second := signedChannel(t, "score-cap-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{low, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: low.ChannelID, From: alice, To: bob, Capacity: "50", FeeAmount: "0", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "150", FeeAmount: "4", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "150", FeeAmount: "4", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy.MaxFeeAmount = "10"
	route, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, Policy: policy})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, "8", route.TotalFee)

	policy.MaxFeeAmount = "3"
	_, err = SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, Policy: policy})
	require.ErrorContains(t, err, "not found")
}

func TestMultiPathSplittingUsesIndependentCapacityAwareRoutes(t *testing.T) {
	alice := testAddress(0x3d)
	r1 := testAddress(0x3e)
	r2 := testAddress(0x3f)
	bob := testAddress(0x40)
	channels := []ChannelRecord{
		signedChannel(t, "split-a", "1000", alice, r1),
		signedChannel(t, "split-b", "1000", r1, bob),
		signedChannel(t, "split-c", "1000", alice, r2),
		signedChannel(t, "split-d", "1000", r2, bob),
	}
	state := EmptyState()
	var err error
	for _, channel := range channels {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	edges := []ChannelEdge{
		{ChannelID: channels[0].ChannelID, From: alice, To: r1, Capacity: "60", FeeAmount: "1", Active: true},
		{ChannelID: channels[1].ChannelID, From: r1, To: bob, Capacity: "60", FeeAmount: "1", Active: true},
		{ChannelID: channels[2].ChannelID, From: alice, To: r2, Capacity: "60", FeeAmount: "2", Active: true},
		{ChannelID: channels[3].ChannelID, From: r2, To: bob, Capacity: "60", FeeAmount: "2", Active: true},
	}
	for _, edge := range edges {
		state, err = RegisterRoutingEdge(state, edge)
		require.NoError(t, err)
	}
	policy := DefaultRoutePolicy()
	policy.EnableMultiPath = true
	policy.MaxSplits = 2
	result, err := SplitPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 30, Policy: policy})
	require.NoError(t, err)
	require.Len(t, result.Parts, 2)
	require.Equal(t, "100", result.TotalAmount)
	require.Equal(t, "6", result.TotalFee)
	require.NotEqual(t, result.Parts[0].Edges[0].ChannelID, result.Parts[1].Edges[0].ChannelID)
	require.NotEmpty(t, result.ScoreHash)
}

func TestCongestionSignalsIncreaseWeightAndReduceMaxPaymentSize(t *testing.T) {
	alice := testAddress(0x57)
	router := testAddress(0x58)
	bob := testAddress(0x59)
	direct := signedChannel(t, "congestion-direct", "1000", alice, bob)
	first := signedChannel(t, "congestion-first", "1000", alice, router)
	second := signedChannel(t, "congestion-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{direct, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: direct.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy, err = ApplyCongestionSnapshot(policy, CongestionSnapshot{
		ChannelID:                   direct.ChannelID,
		From:                        alice,
		To:                          bob,
		ChannelUpdateFailureRateBps: 9_000,
		PendingConditionCount:       8,
		AvgResolutionLatency:        500,
		RouteRetryCount:             3,
		ReservePressureBps:          8_000,
		NodeQueueDelay:              400,
		LiquidityUpdatedHeight:      10,
		ObservedHeight:              100,
	})
	require.NoError(t, err)

	route, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 100, Policy: policy})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{route.Edges[0].ChannelID, route.Edges[1].ChannelID})

	cappedPolicy := policy
	cappedPolicy.ExcludedChannels = []string{first.ChannelID, second.ChannelID}
	_, err = SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "300", CurrentHeight: 100, Policy: cappedPolicy})
	require.ErrorContains(t, err, "eligible")
}

func TestCongestionPenaltyDecayRestoresRoutePreference(t *testing.T) {
	alice := testAddress(0x5a)
	bob := testAddress(0x5b)
	cheap := signedChannel(t, "decay-cheap", "1000", alice, bob)
	expensive := signedChannel(t, "decay-expensive", "1000", alice, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, cheap)
	require.NoError(t, err)
	state, err = OpenChannel(state, expensive)
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: cheap.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: expensive.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "9", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy.DecayHalfLife = 10
	policy, err = ApplyCongestionSnapshot(policy, CongestionSnapshot{
		ChannelID:                   cheap.ChannelID,
		From:                        alice,
		To:                          bob,
		ChannelUpdateFailureRateBps: 9_000,
		PendingConditionCount:       10,
		ReservePressureBps:          1_000,
		ObservedHeight:              10,
	})
	require.NoError(t, err)
	congested, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 10, Policy: policy})
	require.NoError(t, err)
	require.Equal(t, expensive.ChannelID, congested.Edges[0].ChannelID)

	decayed := DecayRoutePolicyPenalties(policy, 120)
	recovered, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 120, Policy: decayed})
	require.NoError(t, err)
	require.Equal(t, cheap.ChannelID, recovered.Edges[0].ChannelID)
}

func TestCongestionAwareRetryPolicySelectsAlternateRoute(t *testing.T) {
	alice := testAddress(0x5c)
	router := testAddress(0x5d)
	bob := testAddress(0x5e)
	direct := signedChannel(t, "retry-direct", "1000", alice, bob)
	first := signedChannel(t, "retry-first", "1000", alice, router)
	second := signedChannel(t, "retry-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{direct, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: direct.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)

	result, err := RetryPaymentRoute(state, TopologyStore{}, RouteRetryRequest{
		Selection: RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 50, Policy: DefaultRoutePolicy()},
		Failures: []RouteFailureReport{{
			ChannelID:      direct.ChannelID,
			From:           alice,
			To:             bob,
			FailureClass:   ClassifyRouteFailure("node queue congestion"),
			Retryable:      true,
			ObservedHeight: 50,
		}},
		Policy: RouteRetryPolicy{MaxAttempts: 3, AlternateRouteLimit: 2, ExcludeFailedEdges: true},
	})
	require.NoError(t, err)
	require.True(t, result.Retryable)
	require.Equal(t, uint32(2), result.Attempts)
	require.Len(t, result.Route.Edges, 2)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{result.Route.Edges[0].ChannelID, result.Route.Edges[1].ChannelID})
	require.NotEmpty(t, result.PolicyHash)

	exhausted, err := RetryPaymentRoute(state, TopologyStore{}, RouteRetryRequest{
		Selection: RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 50, Policy: DefaultRoutePolicy()},
		Failures: []RouteFailureReport{
			{ChannelID: direct.ChannelID, From: alice, To: bob, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 50},
			{ChannelID: first.ChannelID, From: alice, To: router, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 51},
			{ChannelID: second.ChannelID, From: router, To: bob, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 52},
		},
		Policy: RouteRetryPolicy{MaxAttempts: 3, AlternateRouteLimit: 2, ExcludeFailedEdges: true},
	})
	require.NoError(t, err)
	require.False(t, exhausted.Retryable)
	require.Contains(t, exhausted.Reason, "attempts exhausted")
}

func TestForwardingPacketsExposeOnlyPerHopMetadata(t *testing.T) {
	alice := testAddress(0x67)
	router1 := testAddress(0x68)
	router2 := testAddress(0x69)
	bob := testAddress(0x6a)
	route := ScoredRoute{
		Edges: []ChannelEdge{
			{ChannelID: HashParts("privacy-channel-1"), From: alice, To: router1, Capacity: "500", FeeAmount: "1", Active: true},
			{ChannelID: HashParts("privacy-channel-2"), From: router1, To: router2, Capacity: "500", FeeAmount: "2", Active: true},
			{ChannelID: HashParts("privacy-channel-3"), From: router2, To: bob, Capacity: "500", FeeAmount: "3", Active: true},
		},
		Amount:      "100",
		TotalFee:    "6",
		TotalCost:   "9",
		MinCapacity: "500",
		ScoreHash:   HashParts("privacy-score"),
	}
	packets, err := BuildForwardingPackets(route, "payment-seed", 7, 100)
	require.NoError(t, err)
	require.Len(t, packets, 3)
	require.Equal(t, alice, packets[0].ForwardingNode)
	require.Equal(t, router1, packets[0].NextNode)
	require.Equal(t, router1, packets[1].ForwardingNode)
	require.Equal(t, router2, packets[1].NextNode)
	require.NotEqual(t, packets[0].RouteID, packets[1].RouteID)
	require.NotEqual(t, packets[1].RouteID, packets[2].RouteID)
	require.NotEqual(t, packets[0].HopPaymentID, packets[1].HopPaymentID)
	require.Equal(t, packets[1].PacketHash, packets[0].NextPacketHash)
	require.Equal(t, packets[2].PacketHash, packets[1].NextPacketHash)
	require.Empty(t, packets[2].NextPacketHash)

	logRecord, err := PrivacySafeForwardingLog(packets[1], 50)
	require.NoError(t, err)
	require.Equal(t, packets[1].PacketID, logRecord.PacketID)
	require.NotEqual(t, packets[1].NextNode, logRecord.NextNodeHash)
	require.NotEqual(t, packets[1].Amount, logRecord.AmountHash)
	require.Equal(t, router1, logRecord.ForwardingNode)
}

func TestForwardingPacketReplayProtectionRejectsReusedIdentifiers(t *testing.T) {
	alice := testAddress(0x6b)
	bob := testAddress(0x6c)
	route := ScoredRoute{
		Edges:       []ChannelEdge{{ChannelID: HashParts("privacy-replay-channel"), From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true}},
		Amount:      "50",
		TotalFee:    "1",
		TotalCost:   "2",
		MinCapacity: "500",
		ScoreHash:   HashParts("privacy-replay-score"),
	}
	_, err := DeriveRouteID("seed", 0)
	require.ErrorContains(t, err, "nonce")
	packets, err := BuildForwardingPackets(route, "seed", 1, 80)
	require.NoError(t, err)
	var records []ForwardingPacketReplayRecord
	records, err = RecordForwardingPacket(records, packets[0], 40)
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.ErrorContains(t, ValidateForwardingPacket(packets[0], alice, records, 41), "replay")

	reusedRoute := packets[0]
	reusedRoute.HopPaymentID = HashParts("new-hop-payment")
	reusedRoute.NextPacketHash = ""
	reusedRoute.PacketHash = ComputeForwardingPacketHash(reusedRoute)
	reusedRoute.PacketID = HashParts("forwarding-packet-id", reusedRoute.PacketHash)
	require.ErrorContains(t, ValidateForwardingPacket(reusedRoute, alice, records, 41), "route id replay")

	pruned := PruneForwardingReplayRecords(records, 40+DefaultReplayHorizon+1)
	require.Empty(t, pruned)
	require.NoError(t, ValidateForwardingPacket(packets[0], alice, pruned, 41))
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

func TestSignedGossipEnvelopeBuildsLocalTopologyStore(t *testing.T) {
	alice := testAddress(0x31)
	bob := testAddress(0x32)
	channel := signedChannel(t, "gossip-announcement", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	envelope := signedGossipEnvelope(t, GossipMessage{
		MessageType:      GossipChannelAnnouncement,
		ChainID:          channel.ChainID,
		ChannelID:        channel.ChannelID,
		NodeID:           alice,
		From:             alice,
		To:               bob,
		Capacity:         "500",
		FeeAmount:        "2",
		ValidAfterHeight: 20,
		ValidUntilHeight: 50,
		ReputationDelta:  3,
		Sequence:         1,
	}, alice, 20)

	store, err := ApplyGossipEnvelope(TopologyStore{}, state, envelope, 20)
	require.NoError(t, err)
	require.Len(t, store.Messages, 1)
	require.Len(t, store.Edges, 1)
	require.Equal(t, channel.ChannelID, store.Edges[0].ChannelID)
	require.Equal(t, int64(3), RoutingScoreForEdge(store, store.Edges[0]))

	commitmentOnly := signedGossipEnvelope(t, GossipMessage{
		MessageType:       GossipChannelAnnouncement,
		ChainID:           channel.ChainID,
		NodeID:            bob,
		From:              bob,
		To:                alice,
		Capacity:          "100",
		FeeAmount:         "1",
		ValidAfterHeight:  20,
		ValidUntilHeight:  50,
		ChannelCommitment: HashParts("verifiable-channel-commitment", bob, alice),
		Sequence:          2,
	}, bob, 20)
	store, err = ApplyGossipEnvelope(store, state, commitmentOnly, 20)
	require.NoError(t, err)
	require.Len(t, store.Messages, 2)
	require.Len(t, store.Edges, 1)
}

func TestGossipExpiryPruningAndInvalidPenaltyAffectLocalScoreOnly(t *testing.T) {
	alice := testAddress(0x33)
	bob := testAddress(0x34)
	channel := signedChannel(t, "gossip-prune", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)
	envelope := signedGossipEnvelope(t, GossipMessage{
		MessageType:      GossipChannelUpdate,
		ChainID:          channel.ChainID,
		ChannelID:        channel.ChannelID,
		NodeID:           alice,
		From:             alice,
		To:               bob,
		Capacity:         "400",
		FeeAmount:        "1",
		ValidAfterHeight: 20,
		ValidUntilHeight: 25,
		Sequence:         1,
	}, alice, 20)

	store, err := ApplyGossipEnvelope(TopologyStore{}, state, envelope, 20)
	require.NoError(t, err)
	require.Len(t, store.Edges, 1)
	pruned, err := PruneTopologyStore(store, 26)
	require.NoError(t, err)
	require.Empty(t, pruned.Messages)
	require.Empty(t, pruned.Edges)

	_, err = ApplyGossipEnvelope(TopologyStore{}, state, envelope, 26)
	require.ErrorContains(t, err, "expired")

	invalid := signedGossipEnvelope(t, GossipMessage{
		MessageType:      GossipLiquidityHint,
		ChainID:          channel.ChainID,
		ChannelID:        channel.ChannelID,
		NodeID:           alice,
		From:             alice,
		To:               bob,
		Liquidity:        "250",
		FeeAmount:        "1",
		ValidAfterHeight: 30,
		ValidUntilHeight: 60,
		Sequence:         2,
		Advisory:         true,
	}, bob, 30)
	penalized, err := ApplyGossipEnvelope(store, state, invalid, 30)
	require.ErrorContains(t, err, "signer must match")
	require.Len(t, penalized.Reputation, 1)
	require.Equal(t, uint64(1), penalized.Reputation[0].InvalidGossip)
	require.Equal(t, -InvalidGossipPenalty, RoutingScoreForEdge(penalized, store.Edges[0]))
	require.Len(t, state.Channels, 1)
	require.Empty(t, state.Edges)
}

func TestFeePolicyGossipRequiresValidityAndMaxFee(t *testing.T) {
	alice := testAddress(0x35)
	bob := testAddress(0x36)
	channel := signedChannel(t, "gossip-fee-policy", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	invalidPolicy := GossipMessage{
		MessageType:      GossipFeePolicyUpdate,
		ChainID:          channel.ChainID,
		ChannelID:        channel.ChannelID,
		NodeID:           alice,
		From:             alice,
		To:               bob,
		FeeAmount:        "1",
		ValidAfterHeight: 20,
		ValidUntilHeight: 50,
	}
	_, err := BuildGossipMessage(invalidPolicy)
	require.ErrorContains(t, err, "max fee")

	policy := signedGossipEnvelope(t, GossipMessage{
		MessageType:      GossipFeePolicyUpdate,
		ChainID:          channel.ChainID,
		ChannelID:        channel.ChannelID,
		NodeID:           alice,
		From:             alice,
		To:               bob,
		Capacity:         "300",
		FeeAmount:        "2",
		MaxFee:           "5",
		ValidAfterHeight: 20,
		ValidUntilHeight: 50,
		Sequence:         1,
	}, alice, 20)
	store, err := ApplyGossipEnvelope(TopologyStore{}, state, policy, 20)
	require.NoError(t, err)
	require.Len(t, store.Edges, 1)
	require.Equal(t, "2", store.Edges[0].FeeAmount)
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

func TestFinalSettlementRequiresResolvedConditionsAndUnlocksCustody(t *testing.T) {
	alice := testAddress(0x55)
	bob := testAddress(0x56)
	channel := signedChannel(t, "condition-settlement", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "25", []Balance{
		{Participant: channel.Participants[0], Amount: "475"},
		{Participant: channel.Participants[1], Amount: "500"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	_, _, err = FinalizeSettlement(state, channel.ChannelID, 40)
	require.ErrorContains(t, err, "conditions")

	resolution := ConditionResolution{
		ConditionID:  closeState.Conditions[0].ConditionID,
		Resolver:     bob,
		Recipient:    closeState.Conditions[0].Payee,
		Amount:       closeState.Conditions[0].Amount,
		EvidenceHash: HashParts("condition-resolution", closeState.Conditions[0].ConditionID),
	}
	state, settlement, err := FinalizeSettlementWithRequest(state, FinalSettlementRequest{
		ChannelID:          channel.ChannelID,
		ResolvedConditions: []ConditionResolution{resolution},
		CurrentHeight:      40,
	})
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Empty(t, state.CustodyLocks)
	require.Equal(t, "475", amountFor(settlement.FinalBalances, channel.Participants[0]))
	require.Equal(t, "525", amountFor(settlement.FinalBalances, channel.Participants[1]))
	require.Len(t, state.ClosedChannels, 1)
	require.Equal(t, channel.ChannelID, state.ClosedChannels[0].ChannelID)
	require.Len(t, state.ConditionClaims, 1)
	require.Equal(t, resolution.ConditionID, state.ConditionClaims[0].ConditionID)
}

func TestSettlementRejectsReusedConditionAndPreimageClaims(t *testing.T) {
	alice := testAddress(0x65)
	bob := testAddress(0x66)
	channel := signedChannel(t, "condition-replay", "1000", alice, bob)
	closeState := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "25", []Balance{
		{Participant: alice, Amount: "975"},
		{Participant: bob, Amount: "0"},
	})
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	resolution := ConditionResolution{
		ConditionID:  closeState.Conditions[0].ConditionID,
		Resolver:     alice,
		Recipient:    bob,
		Amount:       "25",
		EvidenceHash: HashParts("condition-preimage", "shared"),
	}
	reusedCondition := state
	reusedCondition.ConditionClaims = append(reusedCondition.ConditionClaims, ConditionClaimRecord{
		ChainID:        channel.ChainID,
		ChannelID:      channel.ChannelID,
		ConditionID:    resolution.ConditionID,
		EvidenceHash:   HashParts("condition-preimage", "old"),
		ResolvedHeight: 19,
		ExpiresHeight:  19 + DefaultReplayHorizon,
	})
	_, _, err = FinalizeSettlementWithRequest(reusedCondition, FinalSettlementRequest{
		ChannelID:          channel.ChannelID,
		ResolvedConditions: []ConditionResolution{resolution},
		CurrentHeight:      40,
	})
	require.ErrorContains(t, err, "condition claim")

	reusedEvidence := state
	reusedEvidence.ConditionClaims = append(reusedEvidence.ConditionClaims, ConditionClaimRecord{
		ChainID:        channel.ChainID,
		ChannelID:      channel.ChannelID,
		ConditionID:    HashParts("other-condition"),
		EvidenceHash:   resolution.EvidenceHash,
		ResolvedHeight: 19,
		ExpiresHeight:  19 + DefaultReplayHorizon,
	})
	_, _, err = FinalizeSettlementWithRequest(reusedEvidence, FinalSettlementRequest{
		ChannelID:          channel.ChannelID,
		ResolvedConditions: []ConditionResolution{resolution},
		CurrentHeight:      40,
	})
	require.ErrorContains(t, err, "evidence claim")
}

func TestForcedClosePreservesDisputeWindowAfterTimeout(t *testing.T) {
	alice := testAddress(0x4e)
	bob := testAddress(0x4f)
	channel := signedChannel(t, "forced-close", "500", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	_, err = ForcedClose(state, channel.ChannelID, alice, channel.LatestState.TimeoutHeight, "0")
	require.ErrorContains(t, err, "timeout")

	state, err = ForcedClose(state, channel.ChannelID, alice, channel.LatestState.TimeoutHeight+1, "0")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusPendingClose, state.Channels[0].Status)
	require.Equal(t, channel.LatestState.TimeoutHeight+1+channel.DisputePeriod, state.Channels[0].PendingClose.SettleAfterHeight)

	newer := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "225"},
		{Participant: bob, Amount: "275"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newer, bob, channel.LatestState.TimeoutHeight+2)
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, state.Channels[0].PendingClose.State.StateHash)
}

func TestWatchServiceSubmitsStaleCloseDispute(t *testing.T) {
	alice := testAddress(0x69)
	bob := testAddress(0x6a)
	watch := testAddress(0x6b)
	channel := signedChannel(t, "watch-stale-close", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	stale := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	newer := signedState(t, channel, 3, stale.StateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	state, err = SubmitClose(state, channel.ChannelID, stale, alice, 20, "0")
	require.NoError(t, err)

	state, err = SubmitWatchDispute(state, WatchDisputeSubmission{
		WatchService:          watch,
		Delegator:             bob,
		ChannelID:             channel.ChannelID,
		ClosingStateReference: stale.StateHash,
		NewerState:            newer,
		CurrentHeight:         21,
		EvidenceHash:          HashParts("watch-dispute", channel.ChannelID, newer.StateHash),
	})
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, state.Channels[0].PendingClose.State.StateHash)

	_, err = SubmitWatchDispute(state, WatchDisputeSubmission{
		WatchService:          watch,
		Delegator:             watch,
		ChannelID:             channel.ChannelID,
		ClosingStateReference: newer.StateHash,
		NewerState:            newer,
		CurrentHeight:         22,
	})
	require.ErrorContains(t, err, "delegator")
}

func TestFraudCloseSettlesAfterAcceptedProof(t *testing.T) {
	alice := testAddress(0x53)
	bob := testAddress(0x54)
	channel := signedChannel(t, "fraud-close", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	proof := FraudProof{
		ProofID:         HashParts("fraud-close-proof", channel.ChannelID),
		ProofType:       FraudProofTypeDoubleSign,
		SubmittedBy:     bob,
		OffendingSigner: alice,
		StateA:          closeState,
		StateB:          conflicting,
		PenaltyAmount:   "25",
		EvidenceHash:    HashParts("evidence", closeState.StateHash, conflicting.StateHash),
	}
	state, err = SubmitFraudProof(state, channel.ChannelID, proof, 21)
	require.NoError(t, err)
	state, settlement, err := FraudClose(state, channel.ChannelID, 22)
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, "375", amountFor(settlement.FinalBalances, alice))
	require.Equal(t, "625", amountFor(settlement.FinalBalances, bob))
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
		ChainID:             "aetheris-test-1",
		ChannelID:           channelID,
		ChannelType:         ChannelTypeBidirectional,
		Participants:        []string{left, right},
		Denom:               NativeDenom,
		Collateral:          collateral,
		OpenHeight:          10,
		CloseDelay:          8,
		DisputePeriod:       8,
		OpeningFeePaid:      DefaultOpeningFee,
		ConditionalPayments: true,
		CustodyDenom:        NativeDenom,
		CustodyAmount:       collateral,
		Status:              ChannelStatusOpen,
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

func EmptyStateWithChannel(t *testing.T, channel ChannelRecord) PaymentsState {
	t.Helper()

	state, err := OpenChannel(EmptyState(), channel)
	require.NoError(t, err)
	return state
}

func signedState(t *testing.T, channel ChannelRecord, nonce uint64, previous string, balances []Balance) ChannelState {
	t.Helper()

	state, err := BuildState(ChannelState{
		ChainID:           channel.ChainID,
		ChannelID:         channel.ChannelID,
		ChannelType:       channel.ChannelType,
		Denom:             channel.Denom,
		Version:           CurrentStateVersion,
		Epoch:             1,
		Nonce:             nonce,
		Balances:          balances,
		PreviousStateHash: previous,
		TimeoutHeight:     channel.OpenHeight + channel.DisputePeriod + nonce,
		CloseDelay:        channel.DisputePeriod,
		FeePolicyID:       NativeDenom,
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

func signedReserveState(t *testing.T, channel ChannelRecord, nonce uint64, previous, reserveA, reserveB string, balances []Balance) ChannelState {
	t.Helper()

	state, err := BuildState(ChannelState{
		ChainID:           channel.ChainID,
		ChannelID:         channel.ChannelID,
		ChannelType:       channel.ChannelType,
		Denom:             channel.Denom,
		Version:           CurrentStateVersion,
		Epoch:             1,
		Nonce:             nonce,
		Balances:          balances,
		ReserveA:          reserveA,
		ReserveB:          reserveB,
		PreviousStateHash: previous,
		TimeoutHeight:     channel.OpenHeight + channel.DisputePeriod + 70,
		CloseDelay:        channel.DisputePeriod,
		FeePolicyID:       NativeDenom,
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

func signedPromise(t *testing.T, channel ChannelRecord, salt, source, destination, amount, fee string, nonce, timeoutHeight uint64) ConditionalPromise {
	t.Helper()

	return signedLinkedPromise(t, channel, HashParts("promise", channel.ChannelID, salt), source, destination, amount, fee, nonce, timeoutHeight, HashParts("promise-preimage", salt), "", "")
}

func signedPromiseWithHashLock(t *testing.T, channel ChannelRecord, salt, source, destination, amount, fee string, nonce, timeoutHeight uint64, hashLock string) ConditionalPromise {
	t.Helper()

	return signedLinkedPromise(t, channel, HashParts("promise", channel.ChannelID, salt), source, destination, amount, fee, nonce, timeoutHeight, hashLock, "", "")
}

func signedLinkedPromise(t *testing.T, channel ChannelRecord, promiseID, source, destination, amount, fee string, nonce, timeoutHeight uint64, hashLock, previousID, nextID string) ConditionalPromise {
	t.Helper()

	return signedRoutePromise(t, channel, promiseID, "", source, destination, amount, fee, nonce, timeoutHeight, hashLock, previousID, nextID)
}

func signedRoutePromise(t *testing.T, channel ChannelRecord, promiseID, routeID, source, destination, amount, fee string, nonce, timeoutHeight uint64, hashLock, previousID, nextID string) ConditionalPromise {
	t.Helper()

	promise, err := BuildConditionalPromise(ConditionalPromise{
		PromiseID:                 promiseID,
		ChannelID:                 channel.ChannelID,
		Source:                    source,
		Destination:               destination,
		Amount:                    amount,
		Fee:                       fee,
		HashLock:                  hashLock,
		TimeoutHeight:             timeoutHeight,
		TimeoutTimestamp:          int64(timeoutHeight * 10),
		ConditionType:             ConditionTypeHashLock,
		RouteIDOptional:           routeID,
		PreviousPromiseIDOptional: previousID,
		NextPromiseIDOptional:     nextID,
		Nonce:                     nonce,
	})
	require.NoError(t, err)
	promise.Signature, err = SignatureForPromise(channel, promise, source)
	require.NoError(t, err)
	promise = promise.Normalize()
	return promise
}

func signedGossipEnvelope(t *testing.T, message GossipMessage, signer string, receivedAt uint64) SignedGossipEnvelope {
	t.Helper()

	built, err := BuildGossipMessage(message)
	require.NoError(t, err)
	sig, err := SignatureForGossip(built, signer)
	require.NoError(t, err)
	return SignedGossipEnvelope{
		Message:      built,
		MessageHash:  built.MessageID,
		Signature:    sig,
		ReceivedFrom: signer,
		ReceivedAt:   receivedAt,
	}.Normalize()
}

func mutateCanonicalState(state ChannelState, mutate func(*ChannelState)) ChannelState {
	state = state.Normalize()
	state.Signatures = nil
	mutate(&state)
	state.SignaturePreimageHash = ComputeStateSignaturePreimageHash(state)
	state.StateHash = ComputeStateHash(state)
	return state.Normalize()
}

func resignState(t *testing.T, channel ChannelRecord, state ChannelState) ChannelState {
	t.Helper()

	state.Signatures = nil
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	return state.Normalize()
}

func signedConditionalState(t *testing.T, channel ChannelRecord, nonce uint64, previous, conditionAmount string, balances []Balance) ChannelState {
	t.Helper()

	channel = channel.Normalize()
	condition := ConditionalPayment{
		ConditionID:   HashParts("condition", channel.ChannelID, conditionAmount),
		ConditionType: ConditionTypeHashLock,
		Payer:         channel.Participants[0],
		Payee:         channel.Participants[1],
		Amount:        conditionAmount,
		HashLock:      HashParts("condition-preimage", conditionAmount),
		TimeoutHeight: channel.OpenHeight + channel.DisputePeriod + nonce + 2,
		NonceStart:    nonce,
		NonceEnd:      nonce + 2,
	}
	state, err := BuildState(ChannelState{
		ChainID:           channel.ChainID,
		ChannelID:         channel.ChannelID,
		ChannelType:       channel.ChannelType,
		Denom:             channel.Denom,
		Version:           CurrentStateVersion,
		Epoch:             1,
		Nonce:             nonce,
		Balances:          balances,
		ReserveA:          "25",
		ReserveB:          "0",
		Conditions:        []ConditionalPayment{condition},
		PreviousStateHash: previous,
		TimeoutHeight:     channel.OpenHeight + channel.DisputePeriod + nonce,
		CloseDelay:        channel.DisputePeriod,
		FeePolicyID:       NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func signedUnidirectionalChannel(t *testing.T, salt, collateral, payer, receiver string, ackRequired bool) ChannelRecord {
	t.Helper()

	channel := ChannelRecord{
		ChainID:             "aetheris-test-1",
		ChannelID:           HashParts(salt, payer, receiver),
		ChannelType:         ChannelTypeUnidirectional,
		Participants:        []string{payer, receiver},
		Payer:               payer,
		Receiver:            receiver,
		ReceiverAckRequired: ackRequired,
		Denom:               NativeDenom,
		Collateral:          collateral,
		OpenHeight:          10,
		CloseDelay:          8,
		DisputePeriod:       8,
		ExpirationHeight:    72,
		ExpirationTimestamp: 0,
		OpeningFeePaid:      DefaultOpeningFee,
		CustodyDenom:        NativeDenom,
		CustodyAmount:       collateral,
		Status:              ChannelStatusOpen,
	}
	openState, err := BuildState(ChannelState{
		ChainID:          channel.ChainID,
		ChannelID:        channel.ChannelID,
		ChannelType:      channel.ChannelType,
		Denom:            channel.Denom,
		Version:          CurrentStateVersion,
		Epoch:            1,
		Nonce:            1,
		Balances:         []Balance{{Participant: payer, Amount: collateral}, {Participant: receiver, Amount: "0"}},
		TimeoutHeight:    channel.ExpirationHeight,
		TimeoutTimestamp: channel.ExpirationTimestamp,
		CloseDelay:       channel.DisputePeriod,
		FeePolicyID:      NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(openState, signer)
		require.NoError(t, err)
		openState.Signatures = append(openState.Signatures, sig)
	}
	channel.LatestState = openState.Normalize()
	channel.OpeningStateHash = channel.LatestState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func signedUnidirectionalClaim(t *testing.T, channel ChannelRecord, claimed string, nonce, expirationHeight uint64, ack bool) UnidirectionalClaim {
	t.Helper()

	claim, err := BuildUnidirectionalClaim(UnidirectionalClaim{
		ChainID:             channel.ChainID,
		ChannelID:           channel.ChannelID,
		Payer:               channel.Payer,
		Receiver:            channel.Receiver,
		LockedAmount:        channel.Collateral,
		ClaimedAmount:       claimed,
		Nonce:               nonce,
		ExpirationHeight:    expirationHeight,
		ExpirationTimestamp: channel.ExpirationTimestamp,
	})
	require.NoError(t, err)
	claim.PayerSignature, err = SignatureForClaim(claim, channel.Payer)
	require.NoError(t, err)
	if ack {
		claim.ReceiverAckOptional, err = SignatureForClaim(claim, channel.Receiver)
		require.NoError(t, err)
	}
	claim = claim.Normalize()
	if !channel.ReceiverAckRequired || ack {
		require.NoError(t, claim.ValidateForChannel(channel))
	}
	return claim
}

func signedAsyncChannel(t *testing.T, salt, collateral string, balances []Balance, sendWindow, receiveWindow uint64, maxUnacked string, expiryHeight uint64, participants ...string) ChannelRecord {
	t.Helper()

	channel := ChannelRecord{
		ChainID:        "aetheris-test-1",
		ChannelID:      HashParts(append([]string{salt}, participants...)...),
		ChannelType:    ChannelTypeAsync,
		Participants:   participants,
		Denom:          NativeDenom,
		Collateral:     collateral,
		OpenHeight:     10,
		CloseDelay:     8,
		DisputePeriod:  8,
		OpeningFeePaid: DefaultOpeningFee,
		CustodyDenom:   NativeDenom,
		CustodyAmount:  collateral,
		Status:         ChannelStatusOpen,
	}
	openState, err := BuildState(ChannelState{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		ChannelType:        channel.ChannelType,
		Denom:              channel.Denom,
		Version:            CurrentStateVersion,
		Epoch:              1,
		Nonce:              1,
		Balances:           balances,
		CheckpointNonce:    1,
		CheckpointBalances: balances,
		AsyncUpdateRoot:    ComputeAsyncDeltaRootForChannel(channel, nil),
		AcceptedUpdateRoot: ComputeAsyncDeltaRootForChannel(channel, nil),
		SendWindow:         sendWindow,
		ReceiveWindow:      receiveWindow,
		MaxUnackedAmount:   maxUnacked,
		ExpiryHeight:       expiryHeight,
		TimeoutHeight:      expiryHeight,
		CloseDelay:         channel.DisputePeriod,
		FeePolicyID:        NativeDenom,
	})
	require.NoError(t, err)
	channel.LatestState = signAsyncCheckpoint(t, channel, openState)
	channel.OpeningStateHash = channel.LatestState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func signedAsyncDelta(t *testing.T, channel ChannelRecord, salt, from, to, amount string, nonceStart, nonceEnd, expiryHeight uint64) AsyncPaymentDelta {
	t.Helper()

	delta, err := BuildAsyncDelta(AsyncPaymentDelta{
		UpdateID:     HashParts("async-delta", channel.ChannelID, salt),
		ChainID:      channel.ChainID,
		ChannelID:    channel.ChannelID,
		From:         from,
		To:           to,
		Direction:    AsyncDeltaDirection(from, to),
		Amount:       amount,
		NonceStart:   nonceStart,
		NonceEnd:     nonceEnd,
		ExpiryHeight: expiryHeight,
	})
	require.NoError(t, err)
	delta.Signature, err = SignatureForAsyncDelta(delta, from)
	require.NoError(t, err)
	require.NoError(t, delta.ValidateForChannel(channel, channel.OpenHeight))
	return delta.Normalize()
}

func signAsyncCheckpoint(t *testing.T, channel ChannelRecord, state ChannelState) ChannelState {
	t.Helper()

	state.Signatures = nil
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

func allocationAmountFor(allocations []PenaltyAllocation, route PenaltyRoute) string {
	for _, allocation := range allocations {
		if allocation.Route == route {
			return allocation.Amount
		}
	}
	return ""
}
