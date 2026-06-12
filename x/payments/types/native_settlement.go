package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

type NativePaymentSettlementStatus string

const (
	NativePaymentSettlementOpen		NativePaymentSettlementStatus	= "open"
	NativePaymentSettlementClosing		NativePaymentSettlementStatus	= "closing"
	NativePaymentSettlementChallenged	NativePaymentSettlementStatus	= "challenged"
	NativePaymentSettlementSettled		NativePaymentSettlementStatus	= "settled"
	NativePaymentSettlementExpired		NativePaymentSettlementStatus	= "expired"
	NativePaymentSettlementDisputed		NativePaymentSettlementStatus	= "disputed"
)

type NativePaymentChannelSettlementState struct {
	ChannelID		string
	Participants		[]string
	ZoneID			string
	ShardID			uint32
	Balances		map[string]string
	Nonce			uint64
	ConditionRoot		string
	ExpiryHeight		uint64
	ChallengePeriod		uint64
	LatestStateHash		string
	SettlementStatus	NativePaymentSettlementStatus
	StateRoot		string
}

type NativePaymentChannelFinalityCommitment struct {
	ChannelID		string
	SettlementHash		string
	FinalStateHash		string
	FinalNonce		uint64
	FinancialZoneRoot	string
	AetraCoreProofRoot	string
	PaymentReceiptRoot	string
	SettledHeight		uint64
	FinalityCommitment	string
}

func NewNativePaymentChannelSettlementStateFromRecord(channel ChannelRecord, zoneID string, shardID uint32, expiryHeight uint64) (NativePaymentChannelSettlementState, error) {
	channel = channel.Normalize()
	if err := channel.Validate(); err != nil {
		return NativePaymentChannelSettlementState{}, err
	}
	if expiryHeight == 0 {
		expiryHeight = channel.LatestState.TimeoutHeight
	}
	return BuildNativePaymentChannelSettlementState(NativePaymentChannelSettlementState{
		ChannelID:		channel.ChannelID,
		Participants:		channel.Participants,
		ZoneID:			zoneID,
		ShardID:		shardID,
		Balances:		balancesToMap(channel.LatestState.Balances),
		Nonce:			channel.LatestState.Nonce,
		ConditionRoot:		channel.LatestState.ConditionRoot,
		ExpiryHeight:		expiryHeight,
		ChallengePeriod:	channel.DisputePeriod,
		LatestStateHash:	channel.LatestState.StateHash,
		SettlementStatus:	NativePaymentSettlementOpen,
	})
}

func BuildNativePaymentChannelSettlementState(state NativePaymentChannelSettlementState) (NativePaymentChannelSettlementState, error) {
	state = state.Normalize()
	if state.StateRoot != "" {
		return NativePaymentChannelSettlementState{}, errors.New("payments native settlement state root must be empty before construction")
	}
	if err := state.ValidateFormat(); err != nil {
		return NativePaymentChannelSettlementState{}, err
	}
	state.StateRoot = ComputeNativePaymentChannelSettlementStateRoot(state)
	return state, state.Validate()
}

func (state NativePaymentChannelSettlementState) Normalize() NativePaymentChannelSettlementState {
	state.ChannelID = normalizeHash(state.ChannelID)
	state.Participants = normalizeAddressSet(state.Participants)
	state.ZoneID = strings.TrimSpace(state.ZoneID)
	state.Balances = normalizeSettlementBalanceMap(state.Balances)
	state.ConditionRoot = normalizeOptionalHash(state.ConditionRoot)
	if state.ConditionRoot == "" {
		state.ConditionRoot = ComputeConditionsRoot(nil)
	}
	state.LatestStateHash = normalizeHash(state.LatestStateHash)
	if state.SettlementStatus == "" {
		state.SettlementStatus = NativePaymentSettlementOpen
	}
	state.StateRoot = normalizeOptionalHash(state.StateRoot)
	return state
}

func (state NativePaymentChannelSettlementState) ValidateFormat() error {
	state = state.Normalize()
	if err := ValidateHash("payments native settlement channel id", state.ChannelID); err != nil {
		return err
	}
	if err := validateAddressSet("payments native settlement participant", state.Participants, 2, MaxParticipants); err != nil {
		return err
	}
	if err := validatePaymentRoutingToken("payments native settlement zone id", state.ZoneID); err != nil {
		return err
	}
	if len(state.Balances) != len(state.Participants) {
		return errors.New("payments native settlement balances must cover every participant")
	}
	for _, participant := range state.Participants {
		amount, found := state.Balances[participant]
		if !found {
			return errors.New("payments native settlement participant balance is missing")
		}
		if err := validateNonNegativeInt("payments native settlement balance", amount); err != nil {
			return err
		}
	}
	if total, err := sumSettlementBalanceMap(state.Balances); err != nil {
		return err
	} else if !total.IsPositive() {
		return errors.New("payments native settlement balances must be positive")
	}
	if state.Nonce == 0 {
		return errors.New("payments native settlement nonce must be positive")
	}
	if err := ValidateHash("payments native settlement condition root", state.ConditionRoot); err != nil {
		return err
	}
	if state.ExpiryHeight == 0 {
		return errors.New("payments native settlement expiry height must be positive")
	}
	if state.ChallengePeriod == 0 {
		return errors.New("payments native settlement challenge period must be positive")
	}
	if err := ValidateHash("payments native settlement latest state hash", state.LatestStateHash); err != nil {
		return err
	}
	if !IsNativePaymentSettlementStatus(state.SettlementStatus) {
		return fmt.Errorf("unknown payments native settlement status %q", state.SettlementStatus)
	}
	if state.StateRoot != "" {
		return ValidateHash("payments native settlement state root", state.StateRoot)
	}
	return nil
}

func (state NativePaymentChannelSettlementState) Validate() error {
	state = state.Normalize()
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.StateRoot == "" {
		return errors.New("payments native settlement state root is required")
	}
	if expected := ComputeNativePaymentChannelSettlementStateRoot(state); state.StateRoot != expected {
		return fmt.Errorf("payments native settlement state root mismatch: expected %s", expected)
	}
	return nil
}

func ValidateNativePaymentChannelCollateralLock(channel ChannelRecord, settlement NativePaymentChannelSettlementState, lock CustodyLock) error {
	channel = channel.Normalize()
	settlement = settlement.Normalize()
	lock = lock.Normalize()
	if err := channel.Validate(); err != nil {
		return err
	}
	if err := settlement.Validate(); err != nil {
		return err
	}
	if lock.ChannelID != channel.ChannelID || lock.ChannelID != settlement.ChannelID {
		return errors.New("payments native settlement custody lock channel mismatch")
	}
	if err := lock.ValidateForChannel(channel); err != nil {
		return err
	}
	total, err := sumSettlementBalanceMap(settlement.Balances)
	if err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments native settlement channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	if !total.Equal(collateral) {
		return errors.New("payments native settlement balances cannot exceed committed escrow")
	}
	return nil
}

func SubmitNativePaymentChannelLatestSignedState(settlement NativePaymentChannelSettlementState, channel ChannelRecord, signedState ChannelState, submitter string, currentHeight uint64) (NativePaymentChannelSettlementState, error) {
	settlement = settlement.Normalize()
	channel = channel.Normalize()
	signedState = signedState.Normalize()
	submitter = strings.TrimSpace(submitter)
	if err := settlement.Validate(); err != nil {
		return NativePaymentChannelSettlementState{}, err
	}
	if err := channel.Validate(); err != nil {
		return NativePaymentChannelSettlementState{}, err
	}
	if err := addressing.ValidateUserAddress("payments native settlement submitter", submitter); err != nil {
		return NativePaymentChannelSettlementState{}, err
	}
	if !containsString(channel.Participants, submitter) {
		return NativePaymentChannelSettlementState{}, errors.New("payments native settlement submitter must be participant")
	}
	if currentHeight == 0 {
		return NativePaymentChannelSettlementState{}, errors.New("payments native settlement submit height must be positive")
	}
	if settlement.ExpiryHeight != 0 && currentHeight > settlement.ExpiryHeight {
		return NativePaymentChannelSettlementState{}, errors.New("payments native settlement update has expired")
	}
	if err := signedState.ValidateForChannel(channel, false); err != nil {
		return NativePaymentChannelSettlementState{}, err
	}
	if signedState.Nonce <= settlement.Nonce {
		return NativePaymentChannelSettlementState{}, errors.New("payments native settlement signed state must increase nonce")
	}
	next := settlement
	next.Balances = balancesToMap(signedState.Balances)
	next.Nonce = signedState.Nonce
	next.ConditionRoot = signedState.ConditionRoot
	next.LatestStateHash = signedState.StateHash
	next.SettlementStatus = NativePaymentSettlementClosing
	next.StateRoot = ""
	return BuildNativePaymentChannelSettlementState(next)
}

func SupersedeNativePaymentStaleCloseWithFraudProof(settlement NativePaymentChannelSettlementState, channel ChannelRecord, staleClose ChannelState, newerState ChannelState, proof FraudProof, submitter string, currentHeight uint64) (NativePaymentChannelSettlementState, SettlementProof, error) {
	settlement = settlement.Normalize()
	channel = channel.Normalize()
	staleClose = staleClose.Normalize()
	newerState = newerState.Normalize()
	proof = proof.Normalize()
	submitter = strings.TrimSpace(submitter)
	if err := settlement.Validate(); err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	if err := channel.Validate(); err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	if err := addressing.ValidateUserAddress("payments native settlement fraud submitter", submitter); err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	if !containsString(channel.Participants, submitter) {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, errors.New("payments native settlement fraud submitter must be participant")
	}
	if currentHeight == 0 {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, errors.New("payments native settlement fraud height must be positive")
	}
	if currentHeight > settlement.ExpiryHeight+settlement.ChallengePeriod {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, errors.New("payments native settlement challenge period has closed")
	}
	if staleClose.StateHash != settlement.LatestStateHash {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, errors.New("payments native settlement stale close hash mismatch")
	}
	if err := staleClose.ValidateForChannel(channel, false); err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	if err := newerState.ValidateForChannel(channel, false); err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	if newerState.Nonce <= staleClose.Nonce || newerState.Nonce <= settlement.Nonce {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, errors.New("payments native settlement fraud state must be newer")
	}
	if proof.ProofType != FraudProofTypeStaleClose {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, errors.New("payments native settlement requires stale-close fraud proof")
	}
	if proof.StateA.StateHash != staleClose.StateHash || proof.StateB.StateHash != newerState.StateHash {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, errors.New("payments native settlement fraud proof state mismatch")
	}
	if err := proof.ValidateForChannel(channel); err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	next := settlement
	next.Balances = balancesToMap(newerState.Balances)
	next.Nonce = newerState.Nonce
	next.ConditionRoot = newerState.ConditionRoot
	next.LatestStateHash = newerState.StateHash
	next.SettlementStatus = NativePaymentSettlementChallenged
	next.StateRoot = ""
	next, err := BuildNativePaymentChannelSettlementState(next)
	if err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	settlementProof, err := BuildSettlementProof(SettlementProof{
		ProofID:		HashParts("native-channel-stale-close-proof", settlement.ChannelID, staleClose.StateHash, newerState.StateHash, fmt.Sprintf("%020d", currentHeight)),
		ProofType:		SettlementProofFraud,
		ChannelID:		settlement.ChannelID,
		LatestStateHash:	newerState.StateHash,
		FraudProofHashOptional:	proof.EvidenceHash,
		SubmittedBy:		submitter,
		Height:			currentHeight,
	})
	if err != nil {
		return NativePaymentChannelSettlementState{}, SettlementProof{}, err
	}
	return next, settlementProof, nil
}

func BuildNativePaymentChannelFinalityCommitment(settlement NativePaymentChannelSettlementState, record SettlementRecord, financialZoneRoot, aetherCoreProofRoot, receiptRoot string) (NativePaymentChannelFinalityCommitment, error) {
	settlement = settlement.Normalize()
	record = record.Normalize()
	commitment := NativePaymentChannelFinalityCommitment{
		ChannelID:		settlement.ChannelID,
		SettlementHash:		record.SettlementHash,
		FinalStateHash:		record.StateHash,
		FinalNonce:		record.Nonce,
		FinancialZoneRoot:	normalizeHash(financialZoneRoot),
		AetraCoreProofRoot:	normalizeHash(aetherCoreProofRoot),
		PaymentReceiptRoot:	normalizeHash(receiptRoot),
		SettledHeight:		record.SettledHeight,
	}
	if err := settlement.Validate(); err != nil {
		return NativePaymentChannelFinalityCommitment{}, err
	}
	if settlement.SettlementStatus != NativePaymentSettlementSettled {
		return NativePaymentChannelFinalityCommitment{}, errors.New("payments native finality commitment requires settled channel state")
	}
	if record.ChannelID != settlement.ChannelID {
		return NativePaymentChannelFinalityCommitment{}, errors.New("payments native finality channel mismatch")
	}
	if record.StateHash != settlement.LatestStateHash || record.Nonce != settlement.Nonce {
		return NativePaymentChannelFinalityCommitment{}, errors.New("payments native finality latest state mismatch")
	}
	if err := ValidateHash("payments native finality settlement hash", commitment.SettlementHash); err != nil {
		return NativePaymentChannelFinalityCommitment{}, err
	}
	if err := ValidateHash("payments native finality financial zone root", commitment.FinancialZoneRoot); err != nil {
		return NativePaymentChannelFinalityCommitment{}, err
	}
	if err := ValidateHash("payments native finality aether core proof root", commitment.AetraCoreProofRoot); err != nil {
		return NativePaymentChannelFinalityCommitment{}, err
	}
	if err := ValidateHash("payments native finality receipt root", commitment.PaymentReceiptRoot); err != nil {
		return NativePaymentChannelFinalityCommitment{}, err
	}
	if commitment.SettledHeight == 0 {
		return NativePaymentChannelFinalityCommitment{}, errors.New("payments native finality settled height must be positive")
	}
	commitment.FinalityCommitment = ComputeNativePaymentChannelFinalityCommitmentHash(commitment)
	return commitment, commitment.Validate()
}

func (commitment NativePaymentChannelFinalityCommitment) Validate() error {
	commitment.ChannelID = normalizeHash(commitment.ChannelID)
	commitment.SettlementHash = normalizeHash(commitment.SettlementHash)
	commitment.FinalStateHash = normalizeHash(commitment.FinalStateHash)
	commitment.FinancialZoneRoot = normalizeHash(commitment.FinancialZoneRoot)
	commitment.AetraCoreProofRoot = normalizeHash(commitment.AetraCoreProofRoot)
	commitment.PaymentReceiptRoot = normalizeHash(commitment.PaymentReceiptRoot)
	commitment.FinalityCommitment = normalizeHash(commitment.FinalityCommitment)
	if err := ValidateHash("payments native finality channel id", commitment.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments native finality settlement hash", commitment.SettlementHash); err != nil {
		return err
	}
	if err := ValidateHash("payments native finality state hash", commitment.FinalStateHash); err != nil {
		return err
	}
	if commitment.FinalNonce == 0 {
		return errors.New("payments native finality nonce must be positive")
	}
	if err := ValidateHash("payments native finality financial zone root", commitment.FinancialZoneRoot); err != nil {
		return err
	}
	if err := ValidateHash("payments native finality aether core proof root", commitment.AetraCoreProofRoot); err != nil {
		return err
	}
	if err := ValidateHash("payments native finality receipt root", commitment.PaymentReceiptRoot); err != nil {
		return err
	}
	if commitment.SettledHeight == 0 {
		return errors.New("payments native finality settled height must be positive")
	}
	if err := ValidateHash("payments native finality commitment", commitment.FinalityCommitment); err != nil {
		return err
	}
	if expected := ComputeNativePaymentChannelFinalityCommitmentHash(commitment); commitment.FinalityCommitment != expected {
		return fmt.Errorf("payments native finality commitment mismatch: expected %s", expected)
	}
	return nil
}

func ComputeNativePaymentChannelSettlementStateRoot(state NativePaymentChannelSettlementState) string {
	state = state.Normalize()
	parts := []string{
		"aetra-native-payment-channel-settlement-state-v1",
		state.ChannelID,
		state.ZoneID,
		fmt.Sprintf("%020d", uint64(state.ShardID)),
		fmt.Sprintf("%020d", state.Nonce),
		state.ConditionRoot,
		fmt.Sprintf("%020d", state.ExpiryHeight),
		fmt.Sprintf("%020d", state.ChallengePeriod),
		state.LatestStateHash,
		string(state.SettlementStatus),
	}
	parts = append(parts, state.Participants...)
	for _, balance := range settlementBalancesFromMap(state.Balances) {
		parts = append(parts, balance.Participant, balance.Amount)
	}
	return HashParts(parts...)
}

func ComputeNativePaymentChannelFinalityCommitmentHash(commitment NativePaymentChannelFinalityCommitment) string {
	commitment.ChannelID = normalizeHash(commitment.ChannelID)
	commitment.SettlementHash = normalizeHash(commitment.SettlementHash)
	commitment.FinalStateHash = normalizeHash(commitment.FinalStateHash)
	commitment.FinancialZoneRoot = normalizeHash(commitment.FinancialZoneRoot)
	commitment.AetraCoreProofRoot = normalizeHash(commitment.AetraCoreProofRoot)
	commitment.PaymentReceiptRoot = normalizeHash(commitment.PaymentReceiptRoot)
	return HashParts(
		"aetra-native-payment-channel-finality-v1",
		commitment.ChannelID,
		commitment.SettlementHash,
		commitment.FinalStateHash,
		fmt.Sprintf("%020d", commitment.FinalNonce),
		commitment.FinancialZoneRoot,
		commitment.AetraCoreProofRoot,
		commitment.PaymentReceiptRoot,
		fmt.Sprintf("%020d", commitment.SettledHeight),
	)
}

func IsNativePaymentSettlementStatus(status NativePaymentSettlementStatus) bool {
	switch status {
	case NativePaymentSettlementOpen,
		NativePaymentSettlementClosing,
		NativePaymentSettlementChallenged,
		NativePaymentSettlementSettled,
		NativePaymentSettlementExpired,
		NativePaymentSettlementDisputed:
		return true
	default:
		return false
	}
}

func balancesToMap(balances []Balance) map[string]string {
	out := make(map[string]string, len(balances))
	for _, balance := range normalizeBalances(balances) {
		out[balance.Participant] = balance.Amount
	}
	return out
}

func normalizeSettlementBalanceMap(balances map[string]string) map[string]string {
	out := make(map[string]string, len(balances))
	for participant, amount := range balances {
		participant = strings.TrimSpace(participant)
		amount = strings.TrimSpace(amount)
		if participant == "" {
			continue
		}
		out[participant] = amount
	}
	return out
}

func settlementBalancesFromMap(balances map[string]string) []Balance {
	out := make([]Balance, 0, len(balances))
	for participant, amount := range normalizeSettlementBalanceMap(balances) {
		out = append(out, Balance{Participant: participant, Amount: amount})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Participant < out[j].Participant })
	return out
}

func sumSettlementBalanceMap(balances map[string]string) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, balance := range settlementBalancesFromMap(balances) {
		amount, err := parseNonNegativeInt("payments native settlement balance", balance.Amount)
		if err != nil {
			return sdkmath.Int{}, err
		}
		total = total.Add(amount)
	}
	return total, nil
}
