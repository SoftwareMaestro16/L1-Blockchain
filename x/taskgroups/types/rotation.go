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
	ProposerStatusReady		= "ready"
	ProposerStatusUnavailable	= "unavailable"
	ProposerStatusFallback		= "fallback"
)

type ProposerPriority struct {
	EpochID			uint64
	Slot			uint64
	TaskGroupID		string
	ValidatorAddress	string
	PriorityScore		sdkmath.Int
	FallbackOrder		uint32
	ProposerStatus		string
}

type ProposerPriorityInput struct {
	ValidatorScore			sdkmath.Int
	PriorProposerPerformanceBps	uint32
	MissedProposalCount		uint64
	TaskReliabilityBps		uint32
	StakeSaturationDampeningBps	uint32
}

type ProposerSelectionInput struct {
	Group		postypes.TaskGroup
	ValidatorScores	map[string]sdkmath.Int
	PriorityInputs	map[string]ProposerPriorityInput
	Unavailable	map[string]bool
}

type ProposerSelection struct {
	EpochID			uint64
	Slot			uint64
	TaskGroupID		string
	CanonicalProposer	string
	VerifierValidators	[]string
	Priorities		[]ProposerPriority
	CanonicalPriority	ProposerPriority
	FallbackUsed		bool
}

type SlotAssignmentInput struct {
	SelectionInput			ProposerSelectionInput
	Slot				uint64
	CurrentHeight			uint64
	MissedProposalTimeoutHeight	uint64
}

type SlotAssignmentRecord struct {
	EpochID				uint64
	Slot				uint64
	TaskGroupID			string
	CanonicalProposer		string
	VerifierValidators		[]string
	FallbackOrder			[]string
	MissedProposalTimeoutHeight	uint64
	FallbackActivated		bool
	EligibilityProof		ProposerEligibilityProof
}

type ProposerEligibilityProof struct {
	EpochID			uint64
	Slot			uint64
	TaskGroupID		string
	ValidatorAddress	string
	FallbackOrder		uint32
	PriorityScore		sdkmath.Int
	AssignmentSeed		string
	ProofHash		string
}

type MissedProposalRecord struct {
	EpochID			uint64
	TaskGroupID		string
	ValidatorAddress	string
	MissedProposalCount	uint64
	LastMissedSlot		uint64
}

func BuildProposerPriorities(input ProposerSelectionInput, slot uint64) ([]ProposerPriority, error) {
	if slot == 0 {
		return nil, errors.New("proposer slot is required")
	}
	if err := input.Group.Validate(); err != nil {
		return nil, err
	}
	if len(input.Group.ProposerOrder) == 0 {
		return nil, errors.New("task group proposer order is required")
	}
	priorities := make([]ProposerPriority, 0, len(input.Group.ProposerOrder))
	for fallbackOrder, validatorID := range input.Group.ProposerOrder {
		priorityInput := input.PriorityInputs[validatorID]
		if priorityInput.ValidatorScore.IsNil() {
			priorityInput.ValidatorScore = input.ValidatorScores[validatorID]
		}
		if priorityInput.PriorProposerPerformanceBps == 0 {
			priorityInput.PriorProposerPerformanceBps = postypes.BasisPoints
		}
		if priorityInput.TaskReliabilityBps == 0 {
			priorityInput.TaskReliabilityBps = postypes.BasisPoints
		}
		if priorityInput.StakeSaturationDampeningBps == 0 {
			priorityInput.StakeSaturationDampeningBps = postypes.BasisPoints
		}
		score, err := ComputeProposerPriorityScore(priorityInput)
		if err != nil {
			return nil, err
		}
		status := ProposerStatusReady
		if input.Unavailable[validatorID] {
			status = ProposerStatusUnavailable
		}
		priorities = append(priorities, ProposerPriority{
			EpochID:		input.Group.EpochID,
			Slot:			slot,
			TaskGroupID:		input.Group.TaskGroupID,
			ValidatorAddress:	validatorID,
			PriorityScore:		score,
			FallbackOrder:		uint32(fallbackOrder),
			ProposerStatus:		status,
		})
	}
	sortProposerPriorities(priorities)
	for i := range priorities {
		priorities[i].FallbackOrder = uint32(i)
	}
	return priorities, nil
}

func SelectCanonicalProposer(input ProposerSelectionInput, slot uint64) (ProposerSelection, error) {
	priorities, err := BuildProposerPriorities(input, slot)
	if err != nil {
		return ProposerSelection{}, err
	}
	var selected ProposerPriority
	found := false
	for _, priority := range priorities {
		if priority.ProposerStatus == ProposerStatusUnavailable {
			continue
		}
		selected = priority
		found = true
		break
	}
	if !found {
		return ProposerSelection{}, errors.New("no available proposer for slot")
	}
	fallbackUsed := false
	if selected.FallbackOrder != 0 || priorities[0].ValidatorAddress != selected.ValidatorAddress {
		fallbackUsed = true
		selected.ProposerStatus = ProposerStatusFallback
	}
	for i := range priorities {
		if priorities[i].ValidatorAddress == selected.ValidatorAddress {
			priorities[i].ProposerStatus = selected.ProposerStatus
			break
		}
	}
	verifiers := make([]string, 0, len(input.Group.ValidatorMembers)-1)
	for _, validatorID := range input.Group.ValidatorMembers {
		if validatorID != selected.ValidatorAddress {
			verifiers = append(verifiers, validatorID)
		}
	}
	sort.Strings(verifiers)
	return ProposerSelection{
		EpochID:		input.Group.EpochID,
		Slot:			slot,
		TaskGroupID:		input.Group.TaskGroupID,
		CanonicalProposer:	selected.ValidatorAddress,
		VerifierValidators:	verifiers,
		Priorities:		priorities,
		CanonicalPriority:	selected,
		FallbackUsed:		fallbackUsed,
	}, nil
}

func BuildSlotAssignment(input SlotAssignmentInput) (SlotAssignmentRecord, error) {
	if input.MissedProposalTimeoutHeight == 0 {
		return SlotAssignmentRecord{}, errors.New("missed proposal timeout height is required")
	}
	priorities, err := BuildProposerPriorities(input.SelectionInput, input.Slot)
	if err != nil {
		return SlotAssignmentRecord{}, err
	}
	if len(priorities) == 0 {
		return SlotAssignmentRecord{}, errors.New("slot assignment requires proposer priorities")
	}
	top := priorities[0]
	fallbackOrder := fallbackOrderFromPriorities(priorities)
	selectionInput := input.SelectionInput
	if top.ProposerStatus == ProposerStatusUnavailable {
		if input.CurrentHeight < input.MissedProposalTimeoutHeight {
			return SlotAssignmentRecord{}, errors.New("fallback cannot activate before missed proposal timeout")
		}
	} else {
		selectionInput.Unavailable = map[string]bool{}
	}
	selection, err := SelectCanonicalProposer(selectionInput, input.Slot)
	if err != nil {
		return SlotAssignmentRecord{}, err
	}
	proof := BuildProposerEligibilityProof(selection.CanonicalPriority, selectionInput.Group)
	record := SlotAssignmentRecord{
		EpochID:			selection.EpochID,
		Slot:				selection.Slot,
		TaskGroupID:			selection.TaskGroupID,
		CanonicalProposer:		selection.CanonicalProposer,
		VerifierValidators:		cloneStrings(selection.VerifierValidators),
		FallbackOrder:			fallbackOrder,
		MissedProposalTimeoutHeight:	input.MissedProposalTimeoutHeight,
		FallbackActivated:		selection.FallbackUsed,
		EligibilityProof:		proof,
	}
	return record, record.Validate(selectionInput.Group)
}

func QueryFallbackOrder(input ProposerSelectionInput, slot uint64) ([]string, error) {
	priorities, err := BuildProposerPriorities(input, slot)
	if err != nil {
		return nil, err
	}
	return fallbackOrderFromPriorities(priorities), nil
}

func BuildProposerEligibilityProof(priority ProposerPriority, group postypes.TaskGroup) ProposerEligibilityProof {
	proof := ProposerEligibilityProof{
		EpochID:		priority.EpochID,
		Slot:			priority.Slot,
		TaskGroupID:		priority.TaskGroupID,
		ValidatorAddress:	priority.ValidatorAddress,
		FallbackOrder:		priority.FallbackOrder,
		PriorityScore:		priority.PriorityScore,
		AssignmentSeed:		group.AssignmentSeed,
	}
	proof.ProofHash = computeEligibilityProofHash(proof)
	return proof
}

func VerifyProposerEligibilityProof(proof ProposerEligibilityProof, priority ProposerPriority, group postypes.TaskGroup) error {
	expected := BuildProposerEligibilityProof(priority, group)
	if proof != expected {
		return errors.New("proposer eligibility proof mismatch")
	}
	return nil
}

func RecordMissedProposal(records []MissedProposalRecord, epochID uint64, slot uint64, taskGroupID string, validatorAddress string) ([]MissedProposalRecord, error) {
	if epochID == 0 || slot == 0 {
		return nil, errors.New("missed proposal epoch and slot are required")
	}
	taskGroupID = strings.TrimSpace(taskGroupID)
	validatorAddress = strings.TrimSpace(validatorAddress)
	if taskGroupID == "" || validatorAddress == "" {
		return nil, errors.New("missed proposal task group and validator are required")
	}
	out := make([]MissedProposalRecord, len(records))
	copy(out, records)
	for i, record := range out {
		if record.EpochID == epochID && record.TaskGroupID == taskGroupID && record.ValidatorAddress == validatorAddress {
			out[i].MissedProposalCount++
			out[i].LastMissedSlot = slot
			sortMissedProposalRecords(out)
			return out, nil
		}
	}
	out = append(out, MissedProposalRecord{
		EpochID:		epochID,
		TaskGroupID:		taskGroupID,
		ValidatorAddress:	validatorAddress,
		MissedProposalCount:	1,
		LastMissedSlot:		slot,
	})
	sortMissedProposalRecords(out)
	return out, nil
}

func ApplyMissedProposalTracking(inputs map[string]ProposerPriorityInput, records []MissedProposalRecord, epochID uint64, taskGroupID string) map[string]ProposerPriorityInput {
	out := make(map[string]ProposerPriorityInput, len(inputs))
	for validatorID, input := range inputs {
		out[validatorID] = input
	}
	for _, record := range records {
		if record.EpochID != epochID || record.TaskGroupID != taskGroupID {
			continue
		}
		input := out[record.ValidatorAddress]
		input.MissedProposalCount += record.MissedProposalCount
		out[record.ValidatorAddress] = input
	}
	return out
}

func ComputeProposerPriorityScore(input ProposerPriorityInput) (sdkmath.Int, error) {
	if input.ValidatorScore.IsNil() || input.ValidatorScore.IsNegative() {
		return sdkmath.Int{}, errors.New("validator score must be non-negative")
	}
	if input.PriorProposerPerformanceBps > postypes.BasisPoints {
		return sdkmath.Int{}, fmt.Errorf("prior proposer performance must be <= %d bps", postypes.BasisPoints)
	}
	if input.TaskReliabilityBps > postypes.BasisPoints {
		return sdkmath.Int{}, fmt.Errorf("task reliability must be <= %d bps", postypes.BasisPoints)
	}
	if input.StakeSaturationDampeningBps > postypes.BasisPoints {
		return sdkmath.Int{}, fmt.Errorf("stake saturation dampening must be <= %d bps", postypes.BasisPoints)
	}
	score := mulIntBps(input.ValidatorScore, input.PriorProposerPerformanceBps)
	score = mulIntBps(score, input.TaskReliabilityBps)
	score = mulIntBps(score, input.StakeSaturationDampeningBps)
	if input.MissedProposalCount == 0 {
		return score, nil
	}
	penaltyBps := input.MissedProposalCount * 1_000
	if penaltyBps >= uint64(postypes.BasisPoints) {
		return sdkmath.ZeroInt(), nil
	}
	return mulIntBps(score, uint32(uint64(postypes.BasisPoints)-penaltyBps)), nil
}

func (p ProposerPriority) Validate() error {
	if p.EpochID == 0 {
		return errors.New("proposer priority epoch id is required")
	}
	if p.Slot == 0 {
		return errors.New("proposer priority slot is required")
	}
	if strings.TrimSpace(p.TaskGroupID) == "" {
		return errors.New("proposer priority task group id is required")
	}
	if strings.TrimSpace(p.ValidatorAddress) == "" {
		return errors.New("proposer priority validator address is required")
	}
	if p.PriorityScore.IsNil() || p.PriorityScore.IsNegative() {
		return errors.New("proposer priority score must be non-negative")
	}
	switch p.ProposerStatus {
	case ProposerStatusReady, ProposerStatusUnavailable, ProposerStatusFallback:
		return nil
	default:
		return fmt.Errorf("unsupported proposer status %q", p.ProposerStatus)
	}
}

func (r SlotAssignmentRecord) Validate(group postypes.TaskGroup) error {
	if r.EpochID != group.EpochID {
		return errors.New("slot assignment epoch does not match task group")
	}
	if r.Slot == 0 {
		return errors.New("slot assignment slot is required")
	}
	if r.TaskGroupID != group.TaskGroupID {
		return errors.New("slot assignment task group mismatch")
	}
	if strings.TrimSpace(r.CanonicalProposer) == "" {
		return errors.New("slot assignment canonical proposer is required")
	}
	if !containsString(group.ValidatorMembers, r.CanonicalProposer) {
		return errors.New("slot assignment proposer is not a group member")
	}
	if len(r.FallbackOrder) != len(group.ValidatorMembers) {
		return errors.New("slot assignment fallback order must include every group member")
	}
	if err := validateStringSet("fallback order", r.FallbackOrder, group.ValidatorMembers); err != nil {
		return err
	}
	if err := validateStringSubset("verifier", r.VerifierValidators, group.ValidatorMembers); err != nil {
		return err
	}
	if containsString(r.VerifierValidators, r.CanonicalProposer) {
		return errors.New("slot assignment proposer cannot also be verifier")
	}
	if r.MissedProposalTimeoutHeight == 0 {
		return errors.New("slot assignment missed proposal timeout height is required")
	}
	if r.EligibilityProof.ProofHash == "" {
		return errors.New("slot assignment eligibility proof is required")
	}
	return nil
}

func sortProposerPriorities(priorities []ProposerPriority) {
	sort.SliceStable(priorities, func(i, j int) bool {
		left := priorities[i]
		right := priorities[j]
		if !left.PriorityScore.Equal(right.PriorityScore) {
			return left.PriorityScore.GT(right.PriorityScore)
		}
		if left.FallbackOrder != right.FallbackOrder {
			return left.FallbackOrder < right.FallbackOrder
		}
		return left.ValidatorAddress < right.ValidatorAddress
	})
}

func fallbackOrderFromPriorities(priorities []ProposerPriority) []string {
	out := make([]string, len(priorities))
	for i, priority := range priorities {
		out[i] = priority.ValidatorAddress
	}
	return out
}

func computeEligibilityProofHash(proof ProposerEligibilityProof) string {
	h := sha256.New()
	writeHashPart(h, fmt.Sprintf("%d", proof.EpochID))
	writeHashPart(h, fmt.Sprintf("%d", proof.Slot))
	writeHashPart(h, proof.TaskGroupID)
	writeHashPart(h, proof.ValidatorAddress)
	writeHashPart(h, fmt.Sprintf("%d", proof.FallbackOrder))
	writeHashPart(h, proof.PriorityScore.String())
	writeHashPart(h, proof.AssignmentSeed)
	return hex.EncodeToString(h.Sum(nil))
}

func writeHashPart(h interface{ Write([]byte) (int, error) }, value string) {
	_, _ = h.Write([]byte(value))
	_, _ = h.Write([]byte{0})
}

func sortMissedProposalRecords(records []MissedProposalRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].EpochID != records[j].EpochID {
			return records[i].EpochID < records[j].EpochID
		}
		if records[i].TaskGroupID != records[j].TaskGroupID {
			return records[i].TaskGroupID < records[j].TaskGroupID
		}
		return records[i].ValidatorAddress < records[j].ValidatorAddress
	})
}

func validateStringSet(fieldName string, values []string, expected []string) error {
	if len(values) != len(expected) {
		return fmt.Errorf("%s must include every group member", fieldName)
	}
	if err := validateStringSubset(fieldName, values, expected); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s value %q", fieldName, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateStringSubset(fieldName string, values []string, expected []string) error {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s contains empty value", fieldName)
		}
		if !containsString(expected, value) {
			return fmt.Errorf("%s value %q is not a group member", fieldName, value)
		}
	}
	return nil
}

func containsString(values []string, needle string) bool {
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
