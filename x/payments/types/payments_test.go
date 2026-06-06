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
	require.Len(t, state.Events, 1)
	require.Equal(t, event, state.Events[0])
	require.Equal(t, "channel-open", event.EventType)
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
