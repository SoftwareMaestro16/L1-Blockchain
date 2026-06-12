package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	ModuleName	= "native-evidence"
	StoreKey	= ModuleName

	SubmitterRoleReporter	= "reporter"
	SubmitterRoleFisherman	= "fisherman"

	VoterRoleValidator	= "validator"
	VoterRoleFisherman	= "fisherman"

	DefaultReporterDepositNaet		= int64(100_000_000)
	DefaultFishermanDepositNaet		= int64(25_000_000)
	DefaultInvalidEvidenceBurnBps		= uint32(5_000)
	DefaultInvalidEvidenceRedirectBps	= uint32(5_000)
)

type Params struct {
	ReporterDepositNaet		sdkmath.Int
	FishermanDepositNaet		sdkmath.Int
	ReporterRewardBps		uint32
	InvalidEvidenceBurnBps		uint32
	InvalidEvidenceRedirectBps	uint32
	Authority			string
	MaxEvidence			uint32
	MaxPendingEvidence		uint32
	MaxProofHashBytes		uint32
	MaxPayloadBytes			uint32
	MaxVotes			uint32
	MaxSideEffectHistory		uint32
	EvidenceTTLBlocks		uint64
	ReviewQuorumBps			uint32
	MinSlashFractionBps		uint32
	MaxSlashFractionBps		uint32
	CriticalFaultSlashFractionBps	uint32
	MaxReporterRewardNaet		uint64
	DoubleSignJailBlocks		uint64
	DowntimeFirstJailBlocks		uint64
	DowntimeRepeatJailBlocks	uint64
	FrozenStakeBlocks		uint64
	DowntimeRepeatMultiplier	uint32
	DowntimeChronicMultiplier	uint32
}

type EvidenceSubmission struct {
	Evidence		postypes.EvidenceRecord
	SubmitterID		string
	SubmitterRole		string
	DepositNaet		sdkmath.Int
	SubmittedHeight		int64
	SubmissionHash		string
	ProofPayload		EvidenceProofPayload
	VerificationGroup	postypes.EvidenceVerificationGroup
}

type EvidenceProofPayload struct {
	EvidenceID		string
	ObjectHash		string
	ProofPayloadHash	string
	PayloadSignature	string
}

type EvidenceDecisionVote struct {
	EvidenceID	string
	VoterID		string
	VoterRole	string
	Accept		bool
	VotingPowerBps	uint32
	SignatureHash	string
	Height		int64
}

type EvidenceMarketSettlement struct {
	EvidenceID		string
	SubmitterID		string
	SubmitterRole		string
	ValidEvidence		bool
	DepositReturnedNaet	sdkmath.Int
	DepositBurnedNaet	sdkmath.Int
	DepositRedirectedNaet	sdkmath.Int
	RewardNaet		sdkmath.Int
	TotalPayoutNaet		sdkmath.Int
	PenaltyAmountNaet	sdkmath.Int
	SettlementHash		string
}

func DefaultParams() Params {
	return Params{
		ReporterDepositNaet:		sdkmath.NewInt(DefaultReporterDepositNaet),
		FishermanDepositNaet:		sdkmath.NewInt(DefaultFishermanDepositNaet),
		ReporterRewardBps:		postypes.DefaultReporterRewardBps,
		InvalidEvidenceBurnBps:		DefaultInvalidEvidenceBurnBps,
		InvalidEvidenceRedirectBps:	DefaultInvalidEvidenceRedirectBps,
		Authority:			prototype.DefaultAuthority,
		MaxEvidence:			MaxEvidenceV1,
		MaxPendingEvidence:		MaxPendingEvidenceV1,
		MaxProofHashBytes:		MaxProofHashBytesV1,
		MaxPayloadBytes:		MaxPayloadBytesV1,
		MaxVotes:			MaxVotesV1,
		MaxSideEffectHistory:		MaxSideEffectHistoryV1,
		EvidenceTTLBlocks:		DefaultEvidenceTTLBlocks,
		ReviewQuorumBps:		DefaultReviewQuorumBps,
		MinSlashFractionBps:		DefaultMinSlashFractionBps,
		MaxSlashFractionBps:		DefaultMaxSlashFractionBps,
		CriticalFaultSlashFractionBps:	DefaultCriticalSlashFractionBps,
		MaxReporterRewardNaet:		DefaultReporterRewardNaet,
		DoubleSignJailBlocks:		DefaultDoubleSignJailBlocks,
		DowntimeFirstJailBlocks:	DefaultDowntimeFirstJailBlocks,
		DowntimeRepeatJailBlocks:	DefaultDowntimeRepeatJailBlocks,
		FrozenStakeBlocks:		DefaultFrozenStakeBlocks,
		DowntimeRepeatMultiplier:	DefaultDowntimeRepeatMultiplier,
		DowntimeChronicMultiplier:	DefaultDowntimeChronicMultiplier,
	}
}

func (p Params) Validate() error {
	if !p.ReporterDepositNaet.IsPositive() {
		return errors.New("reporter evidence deposit must be positive")
	}
	if !p.FishermanDepositNaet.IsPositive() {
		return errors.New("fisherman evidence deposit must be positive")
	}
	if p.ReporterRewardBps > postypes.MaxReporterRewardBps {
		return fmt.Errorf("reporter reward must be <= %d bps", postypes.MaxReporterRewardBps)
	}
	totalPenaltyBps := uint64(p.InvalidEvidenceBurnBps) + uint64(p.InvalidEvidenceRedirectBps)
	if totalPenaltyBps != uint64(postypes.BasisPoints) {
		return fmt.Errorf("invalid evidence deposit routing must sum to %d bps", postypes.BasisPoints)
	}
	if err := addressing.ValidateAuthorityAddress("native evidence authority", p.Authority); err != nil {
		return err
	}
	if p.MaxEvidence == 0 || p.MaxEvidence > MaxEvidenceV1 {
		return fmt.Errorf("native evidence max evidence must be between 1 and %d", MaxEvidenceV1)
	}
	if p.MaxPendingEvidence == 0 || p.MaxPendingEvidence > p.MaxEvidence || p.MaxPendingEvidence > MaxPendingEvidenceV1 {
		return fmt.Errorf("native evidence max pending evidence must be between 1 and %d and <= max evidence", MaxPendingEvidenceV1)
	}
	if p.MaxProofHashBytes == 0 || p.MaxProofHashBytes > MaxProofHashBytesV1 {
		return fmt.Errorf("native evidence max proof hash bytes must be between 1 and %d", MaxProofHashBytesV1)
	}
	if p.MaxPayloadBytes == 0 || p.MaxPayloadBytes > MaxPayloadBytesV1 {
		return fmt.Errorf("native evidence max payload bytes must be between 1 and %d", MaxPayloadBytesV1)
	}
	if p.MaxVotes == 0 || p.MaxVotes > MaxVotesV1 {
		return fmt.Errorf("native evidence max votes must be between 1 and %d", MaxVotesV1)
	}
	if p.MaxSideEffectHistory == 0 || p.MaxSideEffectHistory > MaxSideEffectHistoryV1 {
		return fmt.Errorf("native evidence max side effect history must be between 1 and %d", MaxSideEffectHistoryV1)
	}
	if p.EvidenceTTLBlocks == 0 {
		return errors.New("native evidence ttl must be positive")
	}
	if p.ReviewQuorumBps == 0 || p.ReviewQuorumBps > MaxBasisPoints {
		return fmt.Errorf("native evidence review quorum must be within 1..%d bps", MaxBasisPoints)
	}
	if p.MinSlashFractionBps == 0 || p.MinSlashFractionBps > p.MaxSlashFractionBps || p.MaxSlashFractionBps > MaxBasisPoints {
		return fmt.Errorf("native evidence slash bounds must be within 1..%d bps", MaxBasisPoints)
	}
	if p.CriticalFaultSlashFractionBps < p.MinSlashFractionBps || p.CriticalFaultSlashFractionBps > p.MaxSlashFractionBps {
		return errors.New("native evidence critical slash fraction must be inside slash bounds")
	}
	if p.DowntimeFirstJailBlocks == 0 || p.DowntimeRepeatJailBlocks <= p.DowntimeFirstJailBlocks {
		return errors.New("native evidence downtime jail durations are invalid")
	}
	if p.FrozenStakeBlocks == 0 {
		return errors.New("native evidence frozen stake duration must be positive")
	}
	if p.DowntimeRepeatMultiplier == 0 || p.DowntimeChronicMultiplier < p.DowntimeRepeatMultiplier {
		return errors.New("native evidence downtime repeat multipliers are invalid")
	}
	return nil
}

func SubmitEvidence(params Params, existing []EvidenceSubmission, evidence postypes.EvidenceRecord, proof EvidenceProofPayload, submitterID string, submitterRole string, deposit sdkmath.Int, group postypes.EvidenceVerificationGroup) (EvidenceSubmission, error) {
	if err := params.Validate(); err != nil {
		return EvidenceSubmission{}, err
	}
	if err := evidence.Validate(); err != nil {
		return EvidenceSubmission{}, err
	}
	if err := proof.Validate(evidence); err != nil {
		return EvidenceSubmission{}, err
	}
	submitterID = strings.TrimSpace(submitterID)
	if submitterID == "" {
		return EvidenceSubmission{}, errors.New("evidence submitter id is required")
	}
	submitterRole = strings.TrimSpace(submitterRole)
	if !isSubmitterRole(submitterRole) {
		return EvidenceSubmission{}, fmt.Errorf("unsupported evidence submitter role %q", submitterRole)
	}
	if submitterRole == SubmitterRoleReporter && submitterID != evidence.Reporter {
		return EvidenceSubmission{}, errors.New("reporter submitter must match evidence reporter")
	}
	requiredDeposit := RequiredDeposit(params, submitterRole)
	if deposit.IsNil() || deposit.LT(requiredDeposit) {
		return EvidenceSubmission{}, fmt.Errorf("evidence deposit below required amount %s", requiredDeposit.String())
	}
	if err := rejectDuplicateEvidence(existing, evidence); err != nil {
		return EvidenceSubmission{}, err
	}
	if err := group.Validate(); err != nil {
		return EvidenceSubmission{}, err
	}
	if group.EvidenceID != evidence.EvidenceID {
		return EvidenceSubmission{}, errors.New("verification group evidence id mismatch")
	}
	submission := EvidenceSubmission{
		Evidence:		evidence,
		SubmitterID:		submitterID,
		SubmitterRole:		submitterRole,
		DepositNaet:		deposit,
		SubmittedHeight:	evidence.SubmittedHeight,
		ProofPayload:		proof,
		VerificationGroup:	group,
	}
	submission.SubmissionHash = ComputeSubmissionHash(submission)
	return submission, submission.Validate()
}

func (s EvidenceSubmission) Validate() error {
	if err := s.Evidence.Validate(); err != nil {
		return err
	}
	if err := s.ProofPayload.Validate(s.Evidence); err != nil {
		return err
	}
	if strings.TrimSpace(s.SubmitterID) == "" {
		return errors.New("evidence submission submitter id is required")
	}
	if !isSubmitterRole(s.SubmitterRole) {
		return fmt.Errorf("unsupported evidence submitter role %q", s.SubmitterRole)
	}
	if s.DepositNaet.IsNil() || !s.DepositNaet.IsPositive() {
		return errors.New("evidence submission deposit must be positive")
	}
	if s.SubmittedHeight < 0 {
		return errors.New("evidence submission height cannot be negative")
	}
	if err := s.VerificationGroup.Validate(); err != nil {
		return err
	}
	if s.VerificationGroup.EvidenceID != s.Evidence.EvidenceID {
		return errors.New("evidence submission verification group mismatch")
	}
	expected := ComputeSubmissionHash(s)
	if s.SubmissionHash != expected {
		return errors.New("evidence submission hash mismatch")
	}
	return nil
}

func (p EvidenceProofPayload) Validate(evidence postypes.EvidenceRecord) error {
	if p.EvidenceID != evidence.EvidenceID {
		return errors.New("proof payload evidence id mismatch")
	}
	if p.ObjectHash != evidence.ObjectHash {
		return errors.New("proof payload object hash mismatch")
	}
	if p.ProofPayloadHash != evidence.ProofPayloadHash {
		return errors.New("proof payload hash mismatch")
	}
	if len(p.PayloadSignature) != postypes.PosHashHexLength {
		return fmt.Errorf("proof payload signature must be %d lowercase hex chars", postypes.PosHashHexLength)
	}
	if _, err := hex.DecodeString(p.PayloadSignature); err != nil {
		return fmt.Errorf("proof payload signature must be hex: %w", err)
	}
	return nil
}

func SelectVerificationGroup(params postypes.Params, epoch postypes.EpochRecord, validators []postypes.ScoredValidator, evidence postypes.EvidenceRecord, minimumGroupSize uint32, decisionThresholdBps uint32) (postypes.EvidenceVerificationGroup, error) {
	return postypes.SelectEvidenceVerificationGroup(postypes.EvidenceVerificationGroupInput{
		Params:			params,
		Epoch:			epoch,
		ActiveValidators:	validators,
		Evidence:		evidence,
		MinimumGroupSize:	minimumGroupSize,
		DecisionThresholdBps:	decisionThresholdBps,
	})
}

func NewDecisionVote(evidenceID string, voterID string, voterRole string, accept bool, votingPowerBps uint32, signatureHash string, height int64, activeValidators []string) (EvidenceDecisionVote, error) {
	vote := EvidenceDecisionVote{
		EvidenceID:	strings.TrimSpace(evidenceID),
		VoterID:	strings.TrimSpace(voterID),
		VoterRole:	strings.TrimSpace(voterRole),
		Accept:		accept,
		VotingPowerBps:	votingPowerBps,
		SignatureHash:	strings.TrimSpace(signatureHash),
		Height:		height,
	}
	return vote, vote.Validate(activeValidators)
}

func (v EvidenceDecisionVote) Validate(activeValidators []string) error {
	if strings.TrimSpace(v.EvidenceID) == "" {
		return errors.New("evidence decision vote id is required")
	}
	if strings.TrimSpace(v.VoterID) == "" {
		return errors.New("evidence decision voter id is required")
	}
	if v.VoterRole == VoterRoleFisherman {
		return errors.New("fishermen cannot decide evidence outcome")
	}
	if v.VoterRole != VoterRoleValidator {
		return fmt.Errorf("unsupported evidence decision voter role %q", v.VoterRole)
	}
	if !contains(activeValidators, v.VoterID) {
		return errors.New("evidence decision voter must be an active validator")
	}
	if v.VotingPowerBps == 0 || v.VotingPowerBps > postypes.BasisPoints {
		return fmt.Errorf("evidence decision voting power must be within 1..%d bps", postypes.BasisPoints)
	}
	if len(v.SignatureHash) != postypes.PosHashHexLength {
		return fmt.Errorf("evidence decision signature must be %d lowercase hex chars", postypes.PosHashHexLength)
	}
	if _, err := hex.DecodeString(v.SignatureHash); err != nil {
		return fmt.Errorf("evidence decision signature must be hex: %w", err)
	}
	if v.Height < 0 {
		return errors.New("evidence decision vote height cannot be negative")
	}
	return nil
}

func SettleMarketplace(params Params, submission EvidenceSubmission, validEvidence bool, penaltyAmount sdkmath.Int) (EvidenceMarketSettlement, error) {
	if err := params.Validate(); err != nil {
		return EvidenceMarketSettlement{}, err
	}
	if err := submission.Validate(); err != nil {
		return EvidenceMarketSettlement{}, err
	}
	if penaltyAmount.IsNil() || penaltyAmount.IsNegative() {
		return EvidenceMarketSettlement{}, errors.New("evidence penalty amount cannot be negative")
	}
	settlement := EvidenceMarketSettlement{
		EvidenceID:		submission.Evidence.EvidenceID,
		SubmitterID:		submission.SubmitterID,
		SubmitterRole:		submission.SubmitterRole,
		ValidEvidence:		validEvidence,
		DepositReturnedNaet:	sdkmath.ZeroInt(),
		DepositBurnedNaet:	sdkmath.ZeroInt(),
		DepositRedirectedNaet:	sdkmath.ZeroInt(),
		RewardNaet:		sdkmath.ZeroInt(),
		TotalPayoutNaet:	sdkmath.ZeroInt(),
		PenaltyAmountNaet:	penaltyAmount,
	}
	if validEvidence {
		reward := mulIntBps(penaltyAmount, params.ReporterRewardBps)
		if reward.GT(penaltyAmount) {
			reward = penaltyAmount
		}
		settlement.DepositReturnedNaet = submission.DepositNaet
		settlement.RewardNaet = reward
		settlement.TotalPayoutNaet = settlement.DepositReturnedNaet.Add(settlement.RewardNaet)
	} else {
		settlement.DepositBurnedNaet = mulIntBps(submission.DepositNaet, params.InvalidEvidenceBurnBps)
		settlement.DepositRedirectedNaet = submission.DepositNaet.Sub(settlement.DepositBurnedNaet)
	}
	settlement.SettlementHash = ComputeMarketSettlementHash(settlement)
	return settlement, nil
}

func RequiredDeposit(params Params, submitterRole string) sdkmath.Int {
	switch submitterRole {
	case SubmitterRoleFisherman:
		return params.FishermanDepositNaet
	default:
		return params.ReporterDepositNaet
	}
}

func ComputeSubmissionHash(submission EvidenceSubmission) string {
	return hashParts("aetra-evidence-submission-v1",
		submission.Evidence.EvidenceID,
		submission.Evidence.ObjectHash,
		submission.Evidence.ProofPayloadHash,
		submission.SubmitterID,
		submission.SubmitterRole,
		submission.DepositNaet.String(),
		fmt.Sprintf("%d", submission.SubmittedHeight),
		submission.VerificationGroup.VerificationGroupID,
		submission.ProofPayload.PayloadSignature,
	)
}

func ComputeMarketSettlementHash(settlement EvidenceMarketSettlement) string {
	return hashParts("aetra-evidence-market-settlement-v1",
		settlement.EvidenceID,
		settlement.SubmitterID,
		settlement.SubmitterRole,
		fmt.Sprintf("%t", settlement.ValidEvidence),
		settlement.DepositReturnedNaet.String(),
		settlement.DepositBurnedNaet.String(),
		settlement.DepositRedirectedNaet.String(),
		settlement.RewardNaet.String(),
		settlement.TotalPayoutNaet.String(),
		settlement.PenaltyAmountNaet.String(),
	)
}

func rejectDuplicateEvidence(existing []EvidenceSubmission, evidence postypes.EvidenceRecord) error {
	for _, submission := range existing {
		if submission.Evidence.EvidenceID == evidence.EvidenceID {
			return fmt.Errorf("duplicate evidence id %q", evidence.EvidenceID)
		}
		if submission.Evidence.ObjectHash == evidence.ObjectHash && submission.Evidence.ProofPayloadHash == evidence.ProofPayloadHash {
			return errors.New("duplicate evidence object and proof payload")
		}
	}
	return nil
}

func isSubmitterRole(role string) bool {
	return role == SubmitterRoleReporter || role == SubmitterRoleFisherman
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func mulIntBps(value sdkmath.Int, bps uint32) sdkmath.Int {
	return value.MulRaw(int64(bps)).QuoRaw(int64(postypes.BasisPoints))
}

func hashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func SortedEvidenceSubmissions(submissions []EvidenceSubmission) []EvidenceSubmission {
	out := make([]EvidenceSubmission, len(submissions))
	copy(out, submissions)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Evidence.EvidenceID != out[j].Evidence.EvidenceID {
			return out[i].Evidence.EvidenceID < out[j].Evidence.EvidenceID
		}
		return out[i].SubmitterID < out[j].SubmitterID
	})
	return out
}
