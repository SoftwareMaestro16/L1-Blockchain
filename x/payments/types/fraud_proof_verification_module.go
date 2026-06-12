package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

type FraudProofVerificationMessageType string

const (
	FraudProofMsgSubmitStaleCloseProof		FraudProofVerificationMessageType	= "MsgSubmitStaleCloseProof"
	FraudProofMsgSubmitDoubleSignProof		FraudProofVerificationMessageType	= "MsgSubmitDoubleSignProof"
	FraudProofMsgSubmitInvalidConditionProof	FraudProofVerificationMessageType	= "MsgSubmitInvalidConditionProof"
	FraudProofMsgSubmitReplayProof			FraudProofVerificationMessageType	= "MsgSubmitReplayProof"
	FraudProofMsgSubmitAsyncOverexposureProof	FraudProofVerificationMessageType	= "MsgSubmitAsyncOverexposureProof"
	FraudProofMsgClaimReporterReward		FraudProofVerificationMessageType	= "MsgClaimReporterReward"
)

type EvidenceRecord struct {
	EvidenceID		string
	ChannelID		string
	ProofID			string
	ProofType		FraudProofType
	CanonicalHash		string
	SubmittedBy		string
	OffendingSigner		string
	SubmittedHeight		uint64
	ExpiresHeight		uint64
	GasUsed			uint64
	VerificationHash	string
}

type PenaltyRecord struct {
	PenaltyID		string
	EvidenceID		string
	ChannelID		string
	ProofID			string
	Offender		string
	TotalPenalty		string
	ReporterReward		string
	CounterpartyComp	string
	Allocations		[]PenaltyAllocation
	Penalties		[]Penalty
	RecordedHeight		uint64
	RecordHash		string
}

type ReporterReward struct {
	RewardID	string
	EvidenceID	string
	ChannelID	string
	ProofID		string
	Reporter	string
	Denom		string
	Amount		string
	Claimed		bool
	ClaimedHeight	uint64
	RewardHash	string
}

type DoubleSignEvidence struct {
	EvidenceID	string
	ChannelID	string
	OffendingSigner	string
	Epoch		uint64
	Nonce		uint64
	StateHashA	string
	StateHashB	string
	CanonicalHash	string
}

type ReplayEvidence struct {
	EvidenceID	string
	ChannelID	string
	ChainID		string
	ReplayChainID	string
	ReplayChannelID	string
	Nonce		uint64
	FinalizedNonce	uint64
	StateHash	string
	CanonicalHash	string
}

type FraudProofVerificationState struct {
	EvidenceRecords		[]EvidenceRecord
	PenaltyRecords		[]PenaltyRecord
	ReporterRewards		[]ReporterReward
	DoubleSignEvidence	[]DoubleSignEvidence
	ReplayEvidence		[]ReplayEvidence
}

type FraudProofGasMeter struct {
	ProofID		string
	ProofType	FraudProofType
	GasLimit	uint64
	GasUsed		uint64
	WithinLimit	bool
	MeterHash	string
}

type FraudProofSubmission struct {
	ChannelID	string
	Proof		FraudProof
	CurrentHeight	uint64
	Policy		FraudPenaltyPolicy
	GasLimit	uint64
}

type FraudProofVerificationMessage interface {
	FraudProofVerificationType() FraudProofVerificationMessageType
	Submission() FraudProofSubmission
	ValidateBasic() error
}

type MsgSubmitStaleCloseProof struct{ Input FraudProofSubmission }
type MsgSubmitDoubleSignProof struct{ Input FraudProofSubmission }
type MsgSubmitInvalidConditionProof struct{ Input FraudProofSubmission }
type MsgSubmitReplayProof struct{ Input FraudProofSubmission }
type MsgSubmitAsyncOverexposureProof struct{ Input FraudProofSubmission }

type MsgClaimReporterReward struct {
	RewardID	string
	Reporter	string
	CurrentHeight	uint64
}

func EmptyFraudProofVerificationState() FraudProofVerificationState {
	return FraudProofVerificationState{}
}

func ApplyFraudProofVerificationMessage(chain PaymentsState, module FraudProofVerificationState, msg interface{}) (PaymentsState, FraudProofVerificationState, error) {
	chain = chain.Export()
	module = module.Export()
	switch m := msg.(type) {
	case FraudProofVerificationMessage:
		if err := m.ValidateBasic(); err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, err
		}
		nextChain, nextModule, err := SubmitFraudProofVerification(chain, module, m.Submission(), m.FraudProofVerificationType())
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, err
		}
		return nextChain.Export(), nextModule.Export(), nil
	case MsgClaimReporterReward:
		next, err := ClaimReporterReward(module, m)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, err
		}
		return chain.Export(), next.Export(), nil
	default:
		return PaymentsState{}, FraudProofVerificationState{}, errors.New("payments fraud proof verification message type is unsupported")
	}
}

func SubmitFraudProofVerification(chain PaymentsState, module FraudProofVerificationState, input FraudProofSubmission, messageType FraudProofVerificationMessageType) (PaymentsState, FraudProofVerificationState, error) {
	chain = chain.Export()
	module = module.Export()
	input = input.Normalize()
	if err := input.ValidateBasic(); err != nil {
		return PaymentsState{}, FraudProofVerificationState{}, err
	}
	if err := validateFraudProofMessageType(input.Proof.ProofType, messageType); err != nil {
		return PaymentsState{}, FraudProofVerificationState{}, err
	}
	index, channel, found := chain.ChannelIndex(input.ChannelID)
	if !found {
		return PaymentsState{}, FraudProofVerificationState{}, errors.New("payments fraud proof verification channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, FraudProofVerificationState{}, errors.New("payments fraud proof verification requires pending close")
	}
	if input.CurrentHeight > channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, FraudProofVerificationState{}, errors.New("payments fraud proof evidence accepted after finalization horizon")
	}
	proof := input.Proof.Normalize()
	if err := proof.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, FraudProofVerificationState{}, err
	}
	canonicalHash := ComputeCanonicalFraudEvidenceHash(channel, proof)
	if module.HasEvidence(canonicalHash) {
		return PaymentsState{}, FraudProofVerificationState{}, errors.New("payments duplicate fraud evidence")
	}
	meter, err := MeterFraudProofVerification(channel, proof, input.GasLimit)
	if err != nil {
		return PaymentsState{}, FraudProofVerificationState{}, err
	}
	accounting, err := BuildPenaltyRouteAccounting(channel, proof, DefaultPenaltyMatrix(), input.Policy)
	if err != nil {
		return PaymentsState{}, FraudProofVerificationState{}, err
	}
	if err := ValidatePenaltyWithinAvailableBalance(channel, accounting); err != nil {
		return PaymentsState{}, FraudProofVerificationState{}, err
	}
	nextChain, err := SubmitFraudProofWithPolicy(chain, input.ChannelID, proof, input.CurrentHeight, input.Policy)
	if err != nil {
		return PaymentsState{}, FraudProofVerificationState{}, err
	}
	if _, nextChannel, found := nextChain.ChannelIndex(input.ChannelID); found {
		channel = nextChannel
	} else {
		channel = chain.Channels[index]
	}
	evidence := EvidenceRecord{
		EvidenceID:		HashParts("fraud-evidence-record", canonicalHash),
		ChannelID:		input.ChannelID,
		ProofID:		proof.ProofID,
		ProofType:		proof.ProofType,
		CanonicalHash:		canonicalHash,
		SubmittedBy:		proof.SubmittedBy,
		OffendingSigner:	proof.OffendingSigner,
		SubmittedHeight:	input.CurrentHeight,
		ExpiresHeight:		input.CurrentHeight + DefaultReplayHorizon,
		GasUsed:		meter.GasUsed,
		VerificationHash:	meter.MeterHash,
	}.WithHash()
	penalty := PenaltyRecordFromAccounting(evidence, accounting, input.CurrentHeight)
	reward := ReporterRewardFromPenaltyRecord(evidence, penalty)
	nextModule := module.Clone()
	nextModule.EvidenceRecords = append(nextModule.EvidenceRecords, evidence)
	nextModule.PenaltyRecords = append(nextModule.PenaltyRecords, penalty)
	if reward.Amount != "0" {
		nextModule.ReporterRewards = append(nextModule.ReporterRewards, reward)
	}
	if proof.ProofType == FraudProofTypeDoubleSign {
		nextModule.DoubleSignEvidence = append(nextModule.DoubleSignEvidence, DoubleSignEvidenceFromProof(evidence, proof))
	}
	if proof.ProofType == FraudProofTypeReplayAttempt {
		nextModule.ReplayEvidence = append(nextModule.ReplayEvidence, ReplayEvidenceFromProof(channel, evidence, proof))
	}
	return nextChain.Export(), nextModule.Export(), nil
}

func ClaimReporterReward(module FraudProofVerificationState, msg MsgClaimReporterReward) (FraudProofVerificationState, error) {
	module = module.Export()
	msg = msg.Normalize()
	if err := msg.ValidateBasic(); err != nil {
		return FraudProofVerificationState{}, err
	}
	index, reward, found := module.ReporterRewardIndex(msg.RewardID)
	if !found {
		return FraudProofVerificationState{}, errors.New("payments reporter reward not found")
	}
	if reward.Reporter != msg.Reporter {
		return FraudProofVerificationState{}, errors.New("payments reporter reward claimant mismatch")
	}
	if reward.Claimed {
		return FraudProofVerificationState{}, errors.New("payments reporter reward already claimed")
	}
	reward.Claimed = true
	reward.ClaimedHeight = msg.CurrentHeight
	reward = reward.WithHash()
	next := module.Clone()
	next.ReporterRewards[index] = reward
	return next.Export(), nil
}

func ComputeCanonicalFraudEvidenceHash(channel ChannelRecord, proof FraudProof) string {
	channel = channel.Normalize()
	proof = proof.Normalize()
	parts := []string{
		"fraud-evidence-v1",
		channel.ChainID,
		channel.ChannelID,
		string(proof.ProofType),
		proof.OffendingSigner,
	}
	switch proof.ProofType {
	case FraudProofTypeDoubleSign:
		left, right := orderedPair(proof.StateA.StateHash, proof.StateB.StateHash)
		parts = append(parts, fmt.Sprintf("%020d", proof.StateA.Epoch), fmt.Sprintf("%020d", proof.StateA.Nonce), left, right)
	case FraudProofTypeStaleClose:
		parts = append(parts, proof.StateA.StateHash, fmt.Sprintf("%020d", proof.StateA.Nonce), proof.StateB.StateHash, fmt.Sprintf("%020d", proof.StateB.Nonce))
	case FraudProofTypeInvalidClose, FraudProofTypeInvalidBalance, FraudProofTypeInvalidCondition:
		parts = append(parts, proof.StateA.StateHash, fmt.Sprintf("%020d", proof.StateA.Nonce), proof.StateA.ConditionRoot, fmt.Sprintf("%010d", proof.StateA.ConditionCount))
	case FraudProofTypeReplayAttempt:
		parts = append(parts, proof.StateA.ChainID, proof.StateA.ChannelID, fmt.Sprintf("%020d", proof.StateA.Epoch), fmt.Sprintf("%020d", proof.StateA.Nonce), proof.StateA.StateHash, fmt.Sprintf("%020d", channel.FinalizedNonce))
	case FraudProofTypeAsyncOverexposure:
		parts = append(parts, proof.AsyncProof.CheckpointState.StateHash, ComputeAsyncDeltaRootForChannel(channel, proof.AsyncProof.Deltas), proof.AsyncProof.EvidenceHash)
	default:
		parts = append(parts, proof.EvidenceHash)
	}
	return HashParts(parts...)
}

func MeterFraudProofVerification(channel ChannelRecord, proof FraudProof, gasLimit uint64) (FraudProofGasMeter, error) {
	channel = channel.Normalize()
	proof = proof.Normalize()
	estimate, err := EstimateSettlementMessageGas(SettlementArbitrationInput{
		Operation:	SettlementArbitrationFraudProof,
		ChannelID:	channel.ChannelID,
		FraudProof:	proof,
		CurrentHeight:	channel.OpenHeight,
	}, SettlementGasCostSchedule{})
	if err != nil {
		return FraudProofGasMeter{}, err
	}
	gasUsed := estimate.TotalGas
	if proof.ProofType == FraudProofTypeAsyncOverexposure {
		gasUsed += uint64(len(proof.AsyncProof.Deltas)) * estimate.BaseGas / 10
	}
	if gasLimit > 0 && gasUsed > gasLimit {
		return FraudProofGasMeter{}, errors.New("payments fraud proof verification gas limit exceeded")
	}
	meter := FraudProofGasMeter{
		ProofID:	proof.ProofID,
		ProofType:	proof.ProofType,
		GasLimit:	gasLimit,
		GasUsed:	gasUsed,
		WithinLimit:	true,
	}
	meter.MeterHash = HashParts("fraud-proof-gas-meter", meter.ProofID, string(meter.ProofType), fmt.Sprintf("%020d", meter.GasLimit), fmt.Sprintf("%020d", meter.GasUsed))
	return meter, nil
}

func ValidatePenaltyWithinAvailableBalance(channel ChannelRecord, accounting PenaltyRouteAccounting) error {
	channel = channel.Normalize()
	accounting = accounting.Normalize()
	byOffender := map[string]sdkmath.Int{}
	for _, penalty := range accounting.Penalties {
		amount, err := parsePositiveInt("payments fraud penalty record amount", penalty.Amount)
		if err != nil {
			return err
		}
		current := byOffender[penalty.Offender]
		if current.IsNil() {
			current = sdkmath.ZeroInt()
		}
		byOffender[penalty.Offender] = current.Add(amount)
	}
	for _, allocation := range accounting.Allocations {
		amount, err := parsePositiveInt("payments fraud penalty allocation amount", allocation.Amount)
		if err != nil {
			return err
		}
		current := byOffender[allocation.Offender]
		if current.IsNil() {
			current = sdkmath.ZeroInt()
		}
		byOffender[allocation.Offender] = current.Add(amount)
	}
	for offender, total := range byOffender {
		available, err := balanceForParticipant(channel.PendingClose.State.Balances, offender)
		if err != nil {
			return err
		}
		if total.GT(available) {
			return errors.New("payments fraud penalty exceeds available balance")
		}
	}
	return nil
}

func PenaltyRecordFromAccounting(evidence EvidenceRecord, accounting PenaltyRouteAccounting, height uint64) PenaltyRecord {
	accounting = accounting.Normalize()
	record := PenaltyRecord{
		PenaltyID:		HashParts("fraud-penalty-record", evidence.EvidenceID),
		EvidenceID:		evidence.EvidenceID,
		ChannelID:		evidence.ChannelID,
		ProofID:		evidence.ProofID,
		Offender:		evidence.OffendingSigner,
		TotalPenalty:		accounting.TotalPenalty,
		ReporterReward:		accounting.ReporterReward,
		CounterpartyComp:	accounting.CounterpartyComp,
		Allocations:		accounting.Allocations,
		Penalties:		accounting.Penalties,
		RecordedHeight:		height,
	}
	return record.WithHash()
}

func ReporterRewardFromPenaltyRecord(evidence EvidenceRecord, penalty PenaltyRecord) ReporterReward {
	return ReporterReward{
		RewardID:	HashParts("fraud-reporter-reward", evidence.EvidenceID, evidence.SubmittedBy),
		EvidenceID:	evidence.EvidenceID,
		ChannelID:	evidence.ChannelID,
		ProofID:	evidence.ProofID,
		Reporter:	evidence.SubmittedBy,
		Denom:		NativeDenom,
		Amount:		penalty.ReporterReward,
	}.WithHash()
}

func DoubleSignEvidenceFromProof(evidence EvidenceRecord, proof FraudProof) DoubleSignEvidence {
	left, right := orderedPair(proof.StateA.StateHash, proof.StateB.StateHash)
	return DoubleSignEvidence{
		EvidenceID:		evidence.EvidenceID,
		ChannelID:		evidence.ChannelID,
		OffendingSigner:	proof.OffendingSigner,
		Epoch:			proof.StateA.Epoch,
		Nonce:			proof.StateA.Nonce,
		StateHashA:		left,
		StateHashB:		right,
		CanonicalHash:		evidence.CanonicalHash,
	}.Normalize()
}

func ReplayEvidenceFromProof(channel ChannelRecord, evidence EvidenceRecord, proof FraudProof) ReplayEvidence {
	channel = channel.Normalize()
	proof = proof.Normalize()
	return ReplayEvidence{
		EvidenceID:		evidence.EvidenceID,
		ChannelID:		channel.ChannelID,
		ChainID:		channel.ChainID,
		ReplayChainID:		proof.StateA.ChainID,
		ReplayChannelID:	proof.StateA.ChannelID,
		Nonce:			proof.StateA.Nonce,
		FinalizedNonce:		channel.FinalizedNonce,
		StateHash:		proof.StateA.StateHash,
		CanonicalHash:		evidence.CanonicalHash,
	}.Normalize()
}

func (s FraudProofVerificationState) Export() FraudProofVerificationState {
	return s.Clone().Normalize()
}

func (s FraudProofVerificationState) Clone() FraudProofVerificationState {
	out := FraudProofVerificationState{
		EvidenceRecords:	make([]EvidenceRecord, len(s.EvidenceRecords)),
		PenaltyRecords:		make([]PenaltyRecord, len(s.PenaltyRecords)),
		ReporterRewards:	make([]ReporterReward, len(s.ReporterRewards)),
		DoubleSignEvidence:	make([]DoubleSignEvidence, len(s.DoubleSignEvidence)),
		ReplayEvidence:		make([]ReplayEvidence, len(s.ReplayEvidence)),
	}
	copy(out.EvidenceRecords, s.EvidenceRecords)
	copy(out.PenaltyRecords, s.PenaltyRecords)
	copy(out.ReporterRewards, s.ReporterRewards)
	copy(out.DoubleSignEvidence, s.DoubleSignEvidence)
	copy(out.ReplayEvidence, s.ReplayEvidence)
	return out
}

func (s FraudProofVerificationState) Normalize() FraudProofVerificationState {
	s.EvidenceRecords = normalizeEvidenceRecords(s.EvidenceRecords)
	s.PenaltyRecords = normalizePenaltyRecords(s.PenaltyRecords)
	s.ReporterRewards = normalizeReporterRewards(s.ReporterRewards)
	s.DoubleSignEvidence = normalizeDoubleSignEvidenceRecords(s.DoubleSignEvidence)
	s.ReplayEvidence = normalizeReplayEvidenceRecords(s.ReplayEvidence)
	return s
}

func (s FraudProofVerificationState) Validate() error {
	state := s.Normalize()
	seenEvidence := map[string]struct{}{}
	for _, record := range state.EvidenceRecords {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seenEvidence[record.CanonicalHash]; found {
			return errors.New("payments duplicate canonical fraud evidence")
		}
		seenEvidence[record.CanonicalHash] = struct{}{}
	}
	for _, record := range state.PenaltyRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, reward := range state.ReporterRewards {
		if err := reward.Validate(); err != nil {
			return err
		}
	}
	for _, evidence := range state.DoubleSignEvidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}
	for _, evidence := range state.ReplayEvidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s FraudProofVerificationState) HasEvidence(canonicalHash string) bool {
	canonicalHash = normalizeHash(canonicalHash)
	for _, record := range s.Normalize().EvidenceRecords {
		if record.CanonicalHash == canonicalHash {
			return true
		}
	}
	return false
}

func (s FraudProofVerificationState) ReporterRewardIndex(rewardID string) (int, ReporterReward, bool) {
	rewardID = normalizeHash(rewardID)
	for i, reward := range s.Normalize().ReporterRewards {
		if reward.RewardID == rewardID {
			return i, reward, true
		}
	}
	return -1, ReporterReward{}, false
}

func (r EvidenceRecord) Normalize() EvidenceRecord {
	r.EvidenceID = normalizeOptionalHash(r.EvidenceID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ProofID = normalizeHash(r.ProofID)
	r.CanonicalHash = normalizeHash(r.CanonicalHash)
	r.SubmittedBy = strings.TrimSpace(r.SubmittedBy)
	r.OffendingSigner = strings.TrimSpace(r.OffendingSigner)
	r.VerificationHash = normalizeOptionalHash(r.VerificationHash)
	return r
}

func (r EvidenceRecord) WithHash() EvidenceRecord {
	r = r.Normalize()
	r.VerificationHash = HashParts("fraud-evidence-record", r.EvidenceID, r.ChannelID, r.ProofID, string(r.ProofType), r.CanonicalHash, r.SubmittedBy, r.OffendingSigner, fmt.Sprintf("%020d", r.SubmittedHeight), fmt.Sprintf("%020d", r.ExpiresHeight), fmt.Sprintf("%020d", r.GasUsed), r.VerificationHash)
	return r.Normalize()
}

func (r EvidenceRecord) Validate() error {
	record := r.Normalize()
	if err := ValidateHash("payments evidence record id", record.EvidenceID); err != nil {
		return err
	}
	if err := ValidateHash("payments evidence record channel", record.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments evidence record proof", record.ProofID); err != nil {
		return err
	}
	if !IsFraudProofType(record.ProofType) {
		return fmt.Errorf("unknown payments evidence proof type %q", record.ProofType)
	}
	if err := ValidateHash("payments canonical evidence hash", record.CanonicalHash); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments evidence submitter", record.SubmittedBy); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments evidence offender", record.OffendingSigner); err != nil {
		return err
	}
	if record.SubmittedHeight == 0 || record.ExpiresHeight <= record.SubmittedHeight || record.GasUsed == 0 {
		return errors.New("payments evidence record heights and gas must be positive")
	}
	return ValidateHash("payments evidence verification hash", record.VerificationHash)
}

func (r PenaltyRecord) Normalize() PenaltyRecord {
	r.PenaltyID = normalizeOptionalHash(r.PenaltyID)
	r.EvidenceID = normalizeHash(r.EvidenceID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ProofID = normalizeHash(r.ProofID)
	r.Offender = strings.TrimSpace(r.Offender)
	r.TotalPenalty = strings.TrimSpace(r.TotalPenalty)
	r.ReporterReward = strings.TrimSpace(r.ReporterReward)
	r.CounterpartyComp = strings.TrimSpace(r.CounterpartyComp)
	if r.ReporterReward == "" {
		r.ReporterReward = "0"
	}
	if r.CounterpartyComp == "" {
		r.CounterpartyComp = "0"
	}
	r.Allocations = normalizePenaltyAllocations(r.Allocations)
	r.Penalties = normalizePenalties(r.Penalties)
	r.RecordHash = normalizeOptionalHash(r.RecordHash)
	return r
}

func (r PenaltyRecord) WithHash() PenaltyRecord {
	r = r.Normalize()
	parts := []string{"fraud-penalty-record", r.PenaltyID, r.EvidenceID, r.ChannelID, r.ProofID, r.Offender, r.TotalPenalty, r.ReporterReward, r.CounterpartyComp, fmt.Sprintf("%020d", r.RecordedHeight)}
	for _, penalty := range r.Penalties {
		parts = append(parts, penalty.Offender, penalty.Recipient, penalty.Denom, penalty.Amount)
	}
	for _, allocation := range r.Allocations {
		parts = append(parts, allocation.Offender, string(allocation.Route), allocation.Denom, allocation.Amount)
	}
	r.RecordHash = HashParts(parts...)
	return r.Normalize()
}

func (r PenaltyRecord) Validate() error {
	record := r.Normalize()
	if err := ValidateHash("payments penalty record id", record.PenaltyID); err != nil {
		return err
	}
	if err := ValidateHash("payments penalty record evidence", record.EvidenceID); err != nil {
		return err
	}
	if err := ValidateHash("payments penalty record proof", record.ProofID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments penalty record offender", record.Offender); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		amount	string
	}{
		{"payments penalty record total", record.TotalPenalty},
		{"payments penalty record reporter", record.ReporterReward},
		{"payments penalty record counterparty", record.CounterpartyComp},
	} {
		if err := validateNonNegativeInt(item.name, item.amount); err != nil {
			return err
		}
	}
	if record.RecordedHeight == 0 {
		return errors.New("payments penalty record height must be positive")
	}
	return ValidateHash("payments penalty record hash", record.RecordHash)
}

func (r ReporterReward) Normalize() ReporterReward {
	r.RewardID = normalizeOptionalHash(r.RewardID)
	r.EvidenceID = normalizeHash(r.EvidenceID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ProofID = normalizeHash(r.ProofID)
	r.Reporter = strings.TrimSpace(r.Reporter)
	r.Denom = normalizeAssetDenom(r.Denom)
	r.Amount = strings.TrimSpace(r.Amount)
	if r.Amount == "" {
		r.Amount = "0"
	}
	r.RewardHash = normalizeOptionalHash(r.RewardHash)
	return r
}

func (r ReporterReward) WithHash() ReporterReward {
	r = r.Normalize()
	r.RewardHash = HashParts("fraud-reporter-reward", r.RewardID, r.EvidenceID, r.ChannelID, r.ProofID, r.Reporter, r.Denom, r.Amount, fmt.Sprintf("%t", r.Claimed), fmt.Sprintf("%020d", r.ClaimedHeight))
	return r.Normalize()
}

func (r ReporterReward) Validate() error {
	reward := r.Normalize()
	if err := ValidateHash("payments reporter reward id", reward.RewardID); err != nil {
		return err
	}
	if err := ValidateHash("payments reporter reward evidence", reward.EvidenceID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments reporter reward reporter", reward.Reporter); err != nil {
		return err
	}
	if reward.Denom != NativeDenom {
		return fmt.Errorf("payments reporter reward denom must be %s", NativeDenom)
	}
	if err := validateNonNegativeInt("payments reporter reward amount", reward.Amount); err != nil {
		return err
	}
	if reward.Claimed && reward.ClaimedHeight == 0 {
		return errors.New("payments reporter reward claimed height must be positive")
	}
	return ValidateHash("payments reporter reward hash", reward.RewardHash)
}

func (d DoubleSignEvidence) Normalize() DoubleSignEvidence {
	d.EvidenceID = normalizeHash(d.EvidenceID)
	d.ChannelID = normalizeHash(d.ChannelID)
	d.OffendingSigner = strings.TrimSpace(d.OffendingSigner)
	d.StateHashA = normalizeHash(d.StateHashA)
	d.StateHashB = normalizeHash(d.StateHashB)
	d.CanonicalHash = normalizeHash(d.CanonicalHash)
	return d
}

func (d DoubleSignEvidence) Validate() error {
	evidence := d.Normalize()
	if err := ValidateHash("payments double sign evidence id", evidence.EvidenceID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments double sign offender", evidence.OffendingSigner); err != nil {
		return err
	}
	if evidence.Epoch == 0 || evidence.Nonce == 0 || evidence.StateHashA == evidence.StateHashB {
		return errors.New("payments double sign evidence requires conflicting same nonce states")
	}
	return ValidateHash("payments double sign canonical evidence", evidence.CanonicalHash)
}

func (r ReplayEvidence) Normalize() ReplayEvidence {
	r.EvidenceID = normalizeHash(r.EvidenceID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ChainID = strings.TrimSpace(r.ChainID)
	r.ReplayChainID = strings.TrimSpace(r.ReplayChainID)
	r.ReplayChannelID = normalizeHash(r.ReplayChannelID)
	r.StateHash = normalizeHash(r.StateHash)
	r.CanonicalHash = normalizeHash(r.CanonicalHash)
	return r
}

func (r ReplayEvidence) Validate() error {
	evidence := r.Normalize()
	if err := ValidateHash("payments replay evidence id", evidence.EvidenceID); err != nil {
		return err
	}
	if evidence.ChainID == "" || evidence.ReplayChainID == "" {
		return errors.New("payments replay evidence chain ids are required")
	}
	if evidence.Nonce > evidence.FinalizedNonce && evidence.ChainID == evidence.ReplayChainID && evidence.ChannelID == evidence.ReplayChannelID {
		return errors.New("payments replay evidence requires finalized nonce or foreign domain")
	}
	return ValidateHash("payments replay canonical evidence", evidence.CanonicalHash)
}

func (a PenaltyRouteAccounting) Normalize() PenaltyRouteAccounting {
	a.TotalPenalty = strings.TrimSpace(a.TotalPenalty)
	a.ReporterReward = strings.TrimSpace(a.ReporterReward)
	a.CounterpartyComp = strings.TrimSpace(a.CounterpartyComp)
	if a.ReporterReward == "" {
		a.ReporterReward = "0"
	}
	if a.CounterpartyComp == "" {
		a.CounterpartyComp = "0"
	}
	a.Allocations = normalizePenaltyAllocations(a.Allocations)
	a.Penalties = normalizePenalties(a.Penalties)
	return a
}

func (s FraudProofSubmission) Normalize() FraudProofSubmission {
	s.ChannelID = normalizeHash(s.ChannelID)
	s.Proof = s.Proof.Normalize()
	s.Policy = s.Policy.Normalize()
	return s
}

func (s FraudProofSubmission) ValidateBasic() error {
	submission := s.Normalize()
	if err := ValidateHash("payments fraud submission channel", submission.ChannelID); err != nil {
		return err
	}
	if submission.CurrentHeight == 0 {
		return errors.New("payments fraud submission height must be positive")
	}
	return submission.Policy.Validate()
}

func (m MsgSubmitStaleCloseProof) FraudProofVerificationType() FraudProofVerificationMessageType {
	return FraudProofMsgSubmitStaleCloseProof
}
func (m MsgSubmitStaleCloseProof) Submission() FraudProofSubmission	{ return m.Input.Normalize() }
func (m MsgSubmitStaleCloseProof) ValidateBasic() error {
	return validateProofMessage(m.Input, FraudProofTypeStaleClose)
}

func (m MsgSubmitDoubleSignProof) FraudProofVerificationType() FraudProofVerificationMessageType {
	return FraudProofMsgSubmitDoubleSignProof
}
func (m MsgSubmitDoubleSignProof) Submission() FraudProofSubmission	{ return m.Input.Normalize() }
func (m MsgSubmitDoubleSignProof) ValidateBasic() error {
	return validateProofMessage(m.Input, FraudProofTypeDoubleSign)
}

func (m MsgSubmitInvalidConditionProof) FraudProofVerificationType() FraudProofVerificationMessageType {
	return FraudProofMsgSubmitInvalidConditionProof
}
func (m MsgSubmitInvalidConditionProof) Submission() FraudProofSubmission	{ return m.Input.Normalize() }
func (m MsgSubmitInvalidConditionProof) ValidateBasic() error {
	return validateProofMessage(m.Input, FraudProofTypeInvalidCondition)
}

func (m MsgSubmitReplayProof) FraudProofVerificationType() FraudProofVerificationMessageType {
	return FraudProofMsgSubmitReplayProof
}
func (m MsgSubmitReplayProof) Submission() FraudProofSubmission	{ return m.Input.Normalize() }
func (m MsgSubmitReplayProof) ValidateBasic() error {
	return validateProofMessage(m.Input, FraudProofTypeReplayAttempt)
}

func (m MsgSubmitAsyncOverexposureProof) FraudProofVerificationType() FraudProofVerificationMessageType {
	return FraudProofMsgSubmitAsyncOverexposureProof
}
func (m MsgSubmitAsyncOverexposureProof) Submission() FraudProofSubmission {
	return m.Input.Normalize()
}
func (m MsgSubmitAsyncOverexposureProof) ValidateBasic() error {
	return validateProofMessage(m.Input, FraudProofTypeAsyncOverexposure)
}

func (m MsgClaimReporterReward) Normalize() MsgClaimReporterReward {
	m.RewardID = normalizeHash(m.RewardID)
	m.Reporter = strings.TrimSpace(m.Reporter)
	return m
}

func (m MsgClaimReporterReward) ValidateBasic() error {
	msg := m.Normalize()
	if err := ValidateHash("payments claim reporter reward id", msg.RewardID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments claim reporter", msg.Reporter); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments claim reporter reward height must be positive")
	}
	return nil
}

func validateProofMessage(input FraudProofSubmission, expected FraudProofType) error {
	input = input.Normalize()
	if err := input.ValidateBasic(); err != nil {
		return err
	}
	if input.Proof.ProofType != expected {
		return errors.New("payments fraud proof message type mismatch")
	}
	return nil
}

func validateFraudProofMessageType(proofType FraudProofType, msgType FraudProofVerificationMessageType) error {
	switch msgType {
	case FraudProofMsgSubmitStaleCloseProof:
		if proofType == FraudProofTypeStaleClose {
			return nil
		}
	case FraudProofMsgSubmitDoubleSignProof:
		if proofType == FraudProofTypeDoubleSign {
			return nil
		}
	case FraudProofMsgSubmitInvalidConditionProof:
		if proofType == FraudProofTypeInvalidCondition {
			return nil
		}
	case FraudProofMsgSubmitReplayProof:
		if proofType == FraudProofTypeReplayAttempt {
			return nil
		}
	case FraudProofMsgSubmitAsyncOverexposureProof:
		if proofType == FraudProofTypeAsyncOverexposure {
			return nil
		}
	}
	return errors.New("payments fraud proof message type mismatch")
}

func balanceForParticipant(balances []Balance, participant string) (sdkmath.Int, error) {
	for _, balance := range normalizeBalances(balances) {
		if balance.Participant != strings.TrimSpace(participant) {
			continue
		}
		return parseNonNegativeInt("payments fraud offender balance", balance.Amount)
	}
	return sdkmath.ZeroInt(), nil
}

func orderedPair(left, right string) (string, string) {
	left = normalizeHash(left)
	right = normalizeHash(right)
	if right < left {
		return right, left
	}
	return left, right
}

func normalizeEvidenceRecords(values []EvidenceRecord) []EvidenceRecord {
	out := make([]EvidenceRecord, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].EvidenceID < out[j].EvidenceID })
	return out
}

func normalizePenaltyRecords(values []PenaltyRecord) []PenaltyRecord {
	out := make([]PenaltyRecord, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PenaltyID < out[j].PenaltyID })
	return out
}

func normalizeReporterRewards(values []ReporterReward) []ReporterReward {
	out := make([]ReporterReward, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RewardID < out[j].RewardID })
	return out
}

func normalizeDoubleSignEvidenceRecords(values []DoubleSignEvidence) []DoubleSignEvidence {
	out := make([]DoubleSignEvidence, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].EvidenceID < out[j].EvidenceID })
	return out
}

func normalizeReplayEvidenceRecords(values []ReplayEvidence) []ReplayEvidence {
	out := make([]ReplayEvidence, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].EvidenceID < out[j].EvidenceID })
	return out
}
