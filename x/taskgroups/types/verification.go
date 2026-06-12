package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	VerificationDutyReexecuteStateTransition	= "reexecute_state_transition"
	VerificationDutyValidateCrossDomainProof	= "validate_cross_domain_proof"
	VerificationDutyVerifyTaskGroup			= "verify_task_group"
	VerificationDutyValidateConsensusOrdering	= "validate_consensus_ordering"
	VerificationDutyVerifyMessageReceipt		= "verify_message_inclusion_receipt"
	VerificationDutySignValidOutput			= "sign_valid_output"
	VerificationDutyRejectInvalidOutput		= "reject_invalid_output"
	VerificationDutySubmitEvidence			= "submit_evidence"

	VerificationResultValid		= "valid"
	VerificationResultInvalid	= "invalid"
	VerificationResultAbstain	= "abstain"
	VerificationResultUnavailable	= "unavailable"

	ProofKindZoneRoot		= "zone_root"
	ProofKindShardRoot		= "shard_root"
	ProofKindMessageRoot		= "message_root"
	ProofKindReceiptRoot		= "receipt_root"
	ProofKindIdentity		= "identity_proof"
	ProofKindPaymentSettlement	= "payment_settlement_proof"
	ProofKindContractExecution	= "contract_execution_proof"
)

type VerificationReceipt struct {
	EpochID			uint64
	TaskGroupID		string
	WorkloadID		string
	ValidatorAddress	string
	VerifiedObjectHash	string
	Result			string
	Signature		string
	GasOrCostOptional	sdkmath.Int
	CreatedHeight		uint64
}

type CrossDomainProof struct {
	ProofID				string
	WorkloadID			string
	ProofKind			string
	SubjectID			string
	RootHash			string
	ParentRootHash			string
	ProofHash			string
	ContractExecutionRequired	bool
	CreatedHeight			uint64
}

type VerificationReceiptSet struct {
	EpochID		uint64
	Receipts	[]VerificationReceipt
	Root		string
}

type VerifierParticipation struct {
	EpochID			uint64
	TaskGroupID		string
	WorkloadID		string
	ValidatorAddress	string
	VerifiedObjectHash	string
	Result			string
	Participated		bool
	ReceiptHash		string
}

type InvalidResultEvidence struct {
	EpochID			uint64
	TaskGroupID		string
	WorkloadID		string
	VerifiedObjectHash	string
	InvalidReceipts		[]VerificationReceipt
	EvidenceHash		string
}

type VerificationAggregation struct {
	EpochID			uint64
	TaskGroupID		string
	WorkloadID		string
	VerifiedObjectHash	string
	ReceiptRoot		string
	ValidCount		uint32
	InvalidCount		uint32
	AbstainCount		uint32
	UnavailableCount	uint32
	ParticipationBps	uint32
	QuorumBps		uint32
	QuorumReached		bool
	InvalidEvidence		*InvalidResultEvidence
}

func RequiredVerificationDuties() []string {
	return []string{
		VerificationDutyReexecuteStateTransition,
		VerificationDutyValidateCrossDomainProof,
		VerificationDutyVerifyTaskGroup,
		VerificationDutyValidateConsensusOrdering,
		VerificationDutyVerifyMessageReceipt,
		VerificationDutySignValidOutput,
		VerificationDutyRejectInvalidOutput,
		VerificationDutySubmitEvidence,
	}
}

func NewVerificationReceipt(group postypes.TaskGroup, validatorAddress string, verifiedObjectHash string, result string, signature string, gasOrCostOptional sdkmath.Int, createdHeight uint64) (VerificationReceipt, error) {
	receipt := VerificationReceipt{
		EpochID:		group.EpochID,
		TaskGroupID:		group.TaskGroupID,
		WorkloadID:		group.WorkloadID,
		ValidatorAddress:	strings.TrimSpace(validatorAddress),
		VerifiedObjectHash:	strings.TrimSpace(verifiedObjectHash),
		Result:			strings.TrimSpace(result),
		Signature:		strings.TrimSpace(signature),
		GasOrCostOptional:	gasOrCostOptional,
		CreatedHeight:		createdHeight,
	}
	if receipt.GasOrCostOptional.IsNil() {
		receipt.GasOrCostOptional = sdkmath.ZeroInt()
	}
	return receipt, receipt.Validate(group)
}

func (r VerificationReceipt) Validate(group postypes.TaskGroup) error {
	if r.EpochID != group.EpochID {
		return errors.New("verification receipt epoch does not match task group")
	}
	if r.TaskGroupID != group.TaskGroupID {
		return errors.New("verification receipt task group mismatch")
	}
	if r.WorkloadID != group.WorkloadID {
		return errors.New("verification receipt workload mismatch")
	}
	if !containsString(group.ValidatorMembers, r.ValidatorAddress) {
		return errors.New("verification receipt validator is not assigned to task group")
	}
	if len(r.VerifiedObjectHash) != postypes.PosHashHexLength {
		return fmt.Errorf("verified object hash must be %d hex chars", postypes.PosHashHexLength)
	}
	if _, err := hex.DecodeString(r.VerifiedObjectHash); err != nil {
		return fmt.Errorf("verified object hash must be hex: %w", err)
	}
	if !isVerificationResult(r.Result) {
		return fmt.Errorf("unsupported verification result %q", r.Result)
	}
	if r.Signature == "" {
		return errors.New("verification receipt signature is required")
	}
	if !r.GasOrCostOptional.IsNil() && r.GasOrCostOptional.IsNegative() {
		return errors.New("verification receipt gas or cost cannot be negative")
	}
	if r.CreatedHeight < group.ActivationHeight || r.CreatedHeight > group.ExpiryHeight {
		return errors.New("verification receipt height outside task group activity window")
	}
	return nil
}

func (p CrossDomainProof) Validate(group postypes.TaskGroup) error {
	if strings.TrimSpace(p.ProofID) == "" {
		return errors.New("cross-domain proof id is required")
	}
	if p.WorkloadID != group.WorkloadID {
		return errors.New("cross-domain proof workload mismatch")
	}
	if strings.TrimSpace(p.SubjectID) == "" {
		return errors.New("cross-domain proof subject id is required")
	}
	if !isProofKind(p.ProofKind) {
		return fmt.Errorf("unsupported cross-domain proof kind %q", p.ProofKind)
	}
	if p.ProofKind == ProofKindContractExecution && !p.ContractExecutionRequired {
		return errors.New("contract execution proof is not configured for workload")
	}
	if err := validateHexHash("cross-domain proof root hash", p.RootHash); err != nil {
		return err
	}
	if err := validateHexHash("cross-domain proof parent root hash", p.ParentRootHash); err != nil {
		return err
	}
	if err := validateHexHash("cross-domain proof hash", p.ProofHash); err != nil {
		return err
	}
	if p.CreatedHeight < group.ActivationHeight || p.CreatedHeight > group.ExpiryHeight {
		return errors.New("cross-domain proof height outside task group activity window")
	}
	return nil
}

func VerifyCrossDomainProof(group postypes.TaskGroup, proof CrossDomainProof) error {
	return proof.Validate(group)
}

func NewVerificationReceiptSet(group postypes.TaskGroup, receipts []VerificationReceipt) (VerificationReceiptSet, error) {
	ordered := make([]VerificationReceipt, len(receipts))
	copy(ordered, receipts)
	sortVerificationReceipts(ordered)
	set := VerificationReceiptSet{
		EpochID:	group.EpochID,
		Receipts:	ordered,
		Root:		ComputeVerificationReceiptRoot(group, ordered),
	}
	return set, set.Validate(group)
}

func (s VerificationReceiptSet) Validate(group postypes.TaskGroup) error {
	if s.EpochID != group.EpochID {
		return errors.New("verification receipt set epoch mismatch")
	}
	seen := make(map[string]struct{}, len(s.Receipts))
	for i, receipt := range s.Receipts {
		if err := receipt.Validate(group); err != nil {
			return err
		}
		key := receipt.ValidatorAddress + "|" + receipt.VerifiedObjectHash
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate verification receipt %s", key)
		}
		seen[key] = struct{}{}
		if i > 0 && compareVerificationReceipts(s.Receipts[i-1], receipt) >= 0 {
			return errors.New("verification receipts must be sorted canonically")
		}
	}
	expectedRoot := ComputeVerificationReceiptRoot(group, s.Receipts)
	if s.Root != expectedRoot {
		return errors.New("verification receipt root mismatch")
	}
	return nil
}

func AggregateVerificationReceipts(group postypes.TaskGroup, set VerificationReceiptSet, verifiedObjectHash string, quorumBps uint32) (VerificationAggregation, error) {
	if quorumBps == 0 || quorumBps > postypes.BasisPoints {
		return VerificationAggregation{}, fmt.Errorf("verification quorum must be between 1 and %d bps", postypes.BasisPoints)
	}
	if err := set.Validate(group); err != nil {
		return VerificationAggregation{}, err
	}
	verifiedObjectHash = strings.TrimSpace(verifiedObjectHash)
	if err := validateHexHash("verified object hash", verifiedObjectHash); err != nil {
		return VerificationAggregation{}, err
	}
	aggregation := VerificationAggregation{
		EpochID:		group.EpochID,
		TaskGroupID:		group.TaskGroupID,
		WorkloadID:		group.WorkloadID,
		VerifiedObjectHash:	verifiedObjectHash,
		ReceiptRoot:		set.Root,
		QuorumBps:		quorumBps,
	}
	invalidReceipts := make([]VerificationReceipt, 0)
	for _, receipt := range set.Receipts {
		if receipt.VerifiedObjectHash != verifiedObjectHash {
			continue
		}
		switch receipt.Result {
		case VerificationResultValid:
			aggregation.ValidCount++
		case VerificationResultInvalid:
			aggregation.InvalidCount++
			invalidReceipts = append(invalidReceipts, receipt)
		case VerificationResultAbstain:
			aggregation.AbstainCount++
		case VerificationResultUnavailable:
			aggregation.UnavailableCount++
		}
	}
	participantCount := aggregation.ValidCount + aggregation.InvalidCount + aggregation.AbstainCount
	if len(group.ValidatorMembers) > 0 {
		aggregation.ParticipationBps = uint32((uint64(participantCount) * uint64(postypes.BasisPoints)) / uint64(len(group.ValidatorMembers)))
	}
	aggregation.QuorumReached = aggregation.ParticipationBps >= quorumBps
	if len(invalidReceipts) > 0 {
		evidence, err := BuildInvalidResultEvidence(group, verifiedObjectHash, invalidReceipts)
		if err != nil {
			return VerificationAggregation{}, err
		}
		aggregation.InvalidEvidence = &evidence
	}
	return aggregation, nil
}

func TrackVerifierParticipation(group postypes.TaskGroup, set VerificationReceiptSet, verifiedObjectHash string) ([]VerifierParticipation, error) {
	if err := set.Validate(group); err != nil {
		return nil, err
	}
	verifiedObjectHash = strings.TrimSpace(verifiedObjectHash)
	if err := validateHexHash("verified object hash", verifiedObjectHash); err != nil {
		return nil, err
	}
	byValidator := make(map[string]VerificationReceipt, len(set.Receipts))
	for _, receipt := range set.Receipts {
		if receipt.VerifiedObjectHash == verifiedObjectHash {
			byValidator[receipt.ValidatorAddress] = receipt
		}
	}
	members := cloneStrings(group.ValidatorMembers)
	sort.Strings(members)
	out := make([]VerifierParticipation, 0, len(members))
	for _, validatorID := range members {
		receipt, found := byValidator[validatorID]
		result := VerificationResultUnavailable
		receiptHash := ""
		if found {
			result = receipt.Result
			receiptHash = ComputeVerificationReceiptHash(receipt)
		}
		out = append(out, VerifierParticipation{
			EpochID:		group.EpochID,
			TaskGroupID:		group.TaskGroupID,
			WorkloadID:		group.WorkloadID,
			ValidatorAddress:	validatorID,
			VerifiedObjectHash:	verifiedObjectHash,
			Result:			result,
			Participated:		found && receipt.Result != VerificationResultUnavailable,
			ReceiptHash:		receiptHash,
		})
	}
	return out, nil
}

func BuildInvalidResultEvidence(group postypes.TaskGroup, verifiedObjectHash string, receipts []VerificationReceipt) (InvalidResultEvidence, error) {
	verifiedObjectHash = strings.TrimSpace(verifiedObjectHash)
	if err := validateHexHash("verified object hash", verifiedObjectHash); err != nil {
		return InvalidResultEvidence{}, err
	}
	ordered := make([]VerificationReceipt, 0, len(receipts))
	for _, receipt := range receipts {
		if err := receipt.Validate(group); err != nil {
			return InvalidResultEvidence{}, err
		}
		if receipt.VerifiedObjectHash != verifiedObjectHash {
			return InvalidResultEvidence{}, errors.New("invalid evidence receipt object hash mismatch")
		}
		if receipt.Result != VerificationResultInvalid {
			return InvalidResultEvidence{}, errors.New("invalid result evidence only accepts invalid receipts")
		}
		ordered = append(ordered, receipt)
	}
	sortVerificationReceipts(ordered)
	evidence := InvalidResultEvidence{
		EpochID:		group.EpochID,
		TaskGroupID:		group.TaskGroupID,
		WorkloadID:		group.WorkloadID,
		VerifiedObjectHash:	verifiedObjectHash,
		InvalidReceipts:	ordered,
	}
	evidence.EvidenceHash = ComputeInvalidResultEvidenceHash(evidence)
	return evidence, nil
}

func ComputeVerificationReceiptHash(receipt VerificationReceipt) string {
	h := sha256.New()
	writeHashPart(h, fmt.Sprintf("%d", receipt.EpochID))
	writeHashPart(h, receipt.TaskGroupID)
	writeHashPart(h, receipt.WorkloadID)
	writeHashPart(h, receipt.ValidatorAddress)
	writeHashPart(h, receipt.VerifiedObjectHash)
	writeHashPart(h, receipt.Result)
	writeHashPart(h, receipt.Signature)
	writeHashPart(h, receipt.GasOrCostOptional.String())
	writeHashPart(h, fmt.Sprintf("%d", receipt.CreatedHeight))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeInvalidResultEvidenceHash(evidence InvalidResultEvidence) string {
	h := sha256.New()
	writeHashPart(h, fmt.Sprintf("%d", evidence.EpochID))
	writeHashPart(h, evidence.TaskGroupID)
	writeHashPart(h, evidence.WorkloadID)
	writeHashPart(h, evidence.VerifiedObjectHash)
	writeHashPart(h, fmt.Sprintf("%d", len(evidence.InvalidReceipts)))
	for _, receipt := range evidence.InvalidReceipts {
		writeHashPart(h, ComputeVerificationReceiptHash(receipt))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVerificationReceiptRoot(group postypes.TaskGroup, receipts []VerificationReceipt) string {
	h := sha256.New()
	writeHashPart(h, fmt.Sprintf("%d", group.EpochID))
	writeHashPart(h, group.TaskGroupID)
	writeHashPart(h, group.WorkloadID)
	writeHashPart(h, fmt.Sprintf("%d", len(receipts)))
	for _, receipt := range receipts {
		writeHashPart(h, ComputeVerificationReceiptHash(receipt))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func isVerificationResult(result string) bool {
	switch result {
	case VerificationResultValid, VerificationResultInvalid, VerificationResultAbstain, VerificationResultUnavailable:
		return true
	default:
		return false
	}
}

func isProofKind(proofKind string) bool {
	switch proofKind {
	case ProofKindZoneRoot,
		ProofKindShardRoot,
		ProofKindMessageRoot,
		ProofKindReceiptRoot,
		ProofKindIdentity,
		ProofKindPaymentSettlement,
		ProofKindContractExecution:
		return true
	default:
		return false
	}
}

func validateHexHash(fieldName string, value string) error {
	value = strings.TrimSpace(value)
	if len(value) != postypes.PosHashHexLength {
		return fmt.Errorf("%s must be %d hex chars", fieldName, postypes.PosHashHexLength)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("%s must be hex: %w", fieldName, err)
	}
	return nil
}

func sortVerificationReceipts(receipts []VerificationReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return compareVerificationReceipts(receipts[i], receipts[j]) < 0
	})
}

func compareVerificationReceipts(left VerificationReceipt, right VerificationReceipt) int {
	if left.ValidatorAddress < right.ValidatorAddress {
		return -1
	}
	if left.ValidatorAddress > right.ValidatorAddress {
		return 1
	}
	if left.VerifiedObjectHash < right.VerifiedObjectHash {
		return -1
	}
	if left.VerifiedObjectHash > right.VerifiedObjectHash {
		return 1
	}
	return 0
}
