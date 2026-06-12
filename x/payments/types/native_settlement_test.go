package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativePaymentChannelSettlementStateMapsSectionNineThreeFields(t *testing.T) {
	alice := testAddress(0x91)
	bob := testAddress(0x92)
	channel := signedChannel(t, "native-settlement-fields", "1000", alice, bob)

	settlement, err := NewNativePaymentChannelSettlementStateFromRecord(channel, "financial", 7, 100)
	require.NoError(t, err)
	require.NoError(t, settlement.Validate())
	require.Equal(t, channel.ChannelID, settlement.ChannelID)
	require.Equal(t, []string{alice, bob}, settlement.Participants)
	require.Equal(t, "financial", settlement.ZoneID)
	require.Equal(t, uint32(7), settlement.ShardID)
	require.Equal(t, "1000", settlement.Balances[alice])
	require.Equal(t, "0", settlement.Balances[bob])
	require.Equal(t, channel.LatestState.Nonce, settlement.Nonce)
	require.Equal(t, channel.LatestState.StateHash, settlement.LatestStateHash)
	require.Equal(t, NativePaymentSettlementOpen, settlement.SettlementStatus)

	lock := CustodyLock{ChannelID: channel.ChannelID, Denom: NativeDenom, Amount: channel.Collateral}
	require.NoError(t, ValidateNativePaymentChannelCollateralLock(channel, settlement, lock))

	overAllocated := settlement
	overAllocated.Balances = map[string]string{alice: "1000", bob: "1"}
	overAllocated.StateRoot = ""
	overAllocated, err = BuildNativePaymentChannelSettlementState(overAllocated)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateNativePaymentChannelCollateralLock(channel, overAllocated, lock), "escrow")
}

func TestNativePaymentChannelAcceptsLatestSignedStateFromAnyParticipant(t *testing.T) {
	alice := testAddress(0x93)
	bob := testAddress(0x94)
	channel := signedChannel(t, "native-settlement-latest", "1000", alice, bob)
	settlement, err := NewNativePaymentChannelSettlementStateFromRecord(channel, "financial", 2, 100)
	require.NoError(t, err)

	newer := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	updated, err := SubmitNativePaymentChannelLatestSignedState(settlement, channel, newer, bob, 20)
	require.NoError(t, err)
	require.Equal(t, newer.Nonce, updated.Nonce)
	require.Equal(t, newer.StateHash, updated.LatestStateHash)
	require.Equal(t, "400", updated.Balances[alice])
	require.Equal(t, "600", updated.Balances[bob])
	require.Equal(t, NativePaymentSettlementClosing, updated.SettlementStatus)

	_, err = SubmitNativePaymentChannelLatestSignedState(updated, channel, newer, alice, 21)
	require.ErrorContains(t, err, "increase nonce")

	outsider := testAddress(0x95)
	later := signedState(t, channel, 3, newer.StateHash, []Balance{
		{Participant: alice, Amount: "350"},
		{Participant: bob, Amount: "650"},
	})
	_, err = SubmitNativePaymentChannelLatestSignedState(updated, channel, later, outsider, 22)
	require.ErrorContains(t, err, "participant")
}

func TestNativePaymentStaleCloseFraudProofSupersedesLowerNonce(t *testing.T) {
	alice := testAddress(0x96)
	bob := testAddress(0x97)
	channel := signedChannel(t, "native-settlement-fraud", "1000", alice, bob)
	settlement, err := NewNativePaymentChannelSettlementStateFromRecord(channel, "financial", 2, 100)
	require.NoError(t, err)

	stale := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	closing, err := SubmitNativePaymentChannelLatestSignedState(settlement, channel, stale, alice, 20)
	require.NoError(t, err)

	newer := signedState(t, channel, 3, stale.StateHash, []Balance{
		{Participant: alice, Amount: "300"},
		{Participant: bob, Amount: "700"},
	})
	proof := FraudProof{
		ProofID:		HashParts("native-stale-close-proof", channel.ChannelID),
		ProofType:		FraudProofTypeStaleClose,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			stale,
		StateB:			newer,
		PenaltyDenom:		NativeDenom,
		PenaltyAmount:		"10",
		EvidenceHash:		HashParts("native-stale-close-evidence", stale.StateHash, newer.StateHash),
	}
	challenged, settlementProof, err := SupersedeNativePaymentStaleCloseWithFraudProof(closing, channel, stale, newer, proof, bob, 25)
	require.NoError(t, err)
	require.Equal(t, NativePaymentSettlementChallenged, challenged.SettlementStatus)
	require.Equal(t, newer.Nonce, challenged.Nonce)
	require.Equal(t, newer.StateHash, challenged.LatestStateHash)
	require.Equal(t, "300", challenged.Balances[alice])
	require.Equal(t, "700", challenged.Balances[bob])
	require.Equal(t, SettlementProofFraud, settlementProof.ProofType)
	require.Equal(t, proof.EvidenceHash, settlementProof.FraudProofHashOptional)

	_, _, err = SupersedeNativePaymentStaleCloseWithFraudProof(closing, channel, stale, stale, proof, bob, 25)
	require.ErrorContains(t, err, "newer")
}

func TestNativePaymentFinalityCommitmentBindsFinancialAndAetraCoreRoots(t *testing.T) {
	alice := testAddress(0x98)
	bob := testAddress(0x99)
	channel := signedChannel(t, "native-settlement-finality", "1000", alice, bob)
	settlement, err := NewNativePaymentChannelSettlementStateFromRecord(channel, "financial", 1, 100)
	require.NoError(t, err)

	finalState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "375"},
		{Participant: bob, Amount: "625"},
	})
	closing, err := SubmitNativePaymentChannelLatestSignedState(settlement, channel, finalState, alice, 20)
	require.NoError(t, err)
	closing.SettlementStatus = NativePaymentSettlementSettled
	closing.StateRoot = ""
	closing, err = BuildNativePaymentChannelSettlementState(closing)
	require.NoError(t, err)

	record := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		finalState.StateHash,
		Nonce:			finalState.Nonce,
		FinalBalances:		settlementBalancesFromMap(closing.Balances),
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		"0",
		SettledHeight:		50,
	}
	record.SettlementHash = ComputeSettlementHash(record)
	require.NoError(t, record.ValidateForChannel(channel))

	commitment, err := BuildNativePaymentChannelFinalityCommitment(
		closing,
		record,
		HashParts("financial-zone-root", record.SettlementHash),
		HashParts("aether-core-proof-root", record.SettlementHash),
		HashParts("payment-receipt-root", record.SettlementHash),
	)
	require.NoError(t, err)
	require.NoError(t, commitment.Validate())
	require.Equal(t, record.SettlementHash, commitment.SettlementHash)
	require.Equal(t, finalState.StateHash, commitment.FinalStateHash)
}
