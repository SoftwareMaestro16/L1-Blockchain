package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMCrossZoneBounceDisabled AVMCrossZoneBounceBehavior = "disabled"
	AVMCrossZoneBounceAllowed  AVMCrossZoneBounceBehavior = "allowed"
	AVMCrossZoneBounceRequired AVMCrossZoneBounceBehavior = "required"

	AVMCrossZoneProofNone         AVMCrossZoneProofRequirement = "none"
	AVMCrossZoneProofAuth         AVMCrossZoneProofRequirement = "auth"
	AVMCrossZoneProofState        AVMCrossZoneProofRequirement = "state"
	AVMCrossZoneProofAuthAndState AVMCrossZoneProofRequirement = "auth_and_state"

	AVMCrossZoneValueNone       AVMCrossZoneValueAccounting = "none"
	AVMCrossZoneValueEscrow     AVMCrossZoneValueAccounting = "escrow"
	AVMCrossZoneValueMessage    AVMCrossZoneValueAccounting = "message_value"
	AVMCrossZoneValueEscrowPlus AVMCrossZoneValueAccounting = "escrow_plus_message_value"

	AVMCrossZoneFailureNone         AVMCrossZoneFailureResolution = "none"
	AVMCrossZoneFailureBounced      AVMCrossZoneFailureResolution = "bounced"
	AVMCrossZoneFailureDeadLettered AVMCrossZoneFailureResolution = "dead_lettered"

	MaxAVMCrossZoneOpcodeLength = 96
)

type AVMCrossZoneBounceBehavior string
type AVMCrossZoneProofRequirement string
type AVMCrossZoneValueAccounting string
type AVMCrossZoneFailureResolution string

type AVMCrossZoneRoutePolicy struct {
	SourceZone       zonestypes.ZoneID
	DestinationZone  zonestypes.ZoneID
	GasPolicy        zonestypes.ZoneGasPolicy
	ExecutionBudget  zonestypes.ZoneExecutionBudget
	MessageFilter    zonestypes.ZoneMessageFilter
	AllowedOpcodes   []string
	BounceBehavior   AVMCrossZoneBounceBehavior
	ProofRequirement AVMCrossZoneProofRequirement
	ValueAccounting  AVMCrossZoneValueAccounting
	PolicyHash       string
}

type AVMCrossZoneExecution struct {
	Message                   AVMAsyncMessage
	RoutePolicy               AVMCrossZoneRoutePolicy
	DestinationQueueEntry     AVMZoneQueueEntry
	DestinationReceipt        AVMExecutionReceipt
	SourceOutputMessagesRoot  string
	DestinationReceiptRoot    string
	DirectStateWriteAttempted bool
	ValueEscrowedNAET         uint64
	ValueAccountedNAET        uint64
	FailureResolution         AVMCrossZoneFailureResolution
	BounceMessageOptional     AVMAsyncMessage
	DeadLetterOptional        AVMDeadLetterRecord
	ExecutionHash             string
}

func NewAVMCrossZoneRoutePolicy(policy AVMCrossZoneRoutePolicy) (AVMCrossZoneRoutePolicy, error) {
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	if policy.GasPolicy.Denom == "" {
		policy.GasPolicy = zonestypes.DefaultZoneGasPolicy()
	}
	if policy.ExecutionBudget.MaxGas == 0 && policy.ExecutionBudget.MaxMessages == 0 {
		policy.ExecutionBudget = zonestypes.DefaultZoneExecutionBudget()
	}
	if len(policy.MessageFilter.AllowedMessageTypes) == 0 {
		policy.MessageFilter = zonestypes.DefaultZoneMessageFilter()
	}
	if policy.BounceBehavior == "" {
		policy.BounceBehavior = AVMCrossZoneBounceAllowed
	}
	if policy.ProofRequirement == "" {
		policy.ProofRequirement = AVMCrossZoneProofNone
	}
	if policy.ValueAccounting == "" {
		policy.ValueAccounting = AVMCrossZoneValueNone
	}
	policy.PolicyHash = ComputeAVMCrossZoneRoutePolicyHash(policy)
	return policy, policy.Validate()
}

func AdmitAVMCrossZoneMessage(queue AVMZoneQueue, msg AVMAsyncMessage, height uint64, maxDepth uint32, policy AVMCrossZoneRoutePolicy) (AVMZoneQueue, AVMZoneQueueEntry, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	if err := policy.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	if err := ValidateAVMCrossZoneRoute(msg, policy); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	if queue.ZoneID != msg.DestinationZone {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM cross-zone destination queue mismatch")
	}
	return AdmitAVMZoneQueueMessage(queue, msg, height, maxDepth)
}

func NewAVMCrossZoneExecution(execution AVMCrossZoneExecution) (AVMCrossZoneExecution, error) {
	execution = canonicalAVMCrossZoneExecution(execution)
	execution.ExecutionHash = ComputeAVMCrossZoneExecutionHash(execution)
	return execution, execution.Validate()
}

func ValidateAVMCrossZoneRoute(msg AVMAsyncMessage, policy AVMCrossZoneRoutePolicy) error {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return err
	}
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	if err := policy.Validate(); err != nil {
		return err
	}
	if msg.SourceZone == msg.DestinationZone {
		return errors.New("AVM cross-zone route requires distinct source and destination zones")
	}
	if msg.SourceZone != policy.SourceZone {
		return errors.New("AVM cross-zone route source zone mismatch")
	}
	if msg.DestinationZone != policy.DestinationZone {
		return errors.New("AVM cross-zone route destination zone mismatch")
	}
	if !policy.MessageFilter.Allows(msg.PayloadType) {
		return fmt.Errorf("AVM cross-zone payload type %q is not allowed by destination filter", msg.PayloadType)
	}
	if !crossZoneOpcodeAllowed(policy.AllowedOpcodes, msg.PayloadType) {
		return fmt.Errorf("AVM cross-zone opcode %q is not allowed", msg.PayloadType)
	}
	if err := validateCrossZoneProofs(msg, policy.ProofRequirement); err != nil {
		return err
	}
	if msg.ValueNAET > 0 && policy.ValueAccounting == AVMCrossZoneValueNone {
		return errors.New("AVM cross-zone value transfer requires escrow or message value accounting")
	}
	return nil
}

func (p AVMCrossZoneRoutePolicy) Validate() error {
	p = canonicalAVMCrossZoneRoutePolicy(p)
	if err := zonestypes.ValidateZoneID(p.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(p.DestinationZone); err != nil {
		return err
	}
	if p.SourceZone == p.DestinationZone {
		return errors.New("AVM cross-zone policy requires distinct source and destination zones")
	}
	if err := p.GasPolicy.Validate(); err != nil {
		return err
	}
	if err := p.ExecutionBudget.Validate(); err != nil {
		return err
	}
	if err := p.MessageFilter.Validate(); err != nil {
		return err
	}
	if !IsAVMCrossZoneBounceBehavior(p.BounceBehavior) {
		return fmt.Errorf("invalid AVM cross-zone bounce behavior %q", p.BounceBehavior)
	}
	if !IsAVMCrossZoneProofRequirement(p.ProofRequirement) {
		return fmt.Errorf("invalid AVM cross-zone proof requirement %q", p.ProofRequirement)
	}
	if !IsAVMCrossZoneValueAccounting(p.ValueAccounting) {
		return fmt.Errorf("invalid AVM cross-zone value accounting %q", p.ValueAccounting)
	}
	if err := validateCrossZoneOpcodes(p.AllowedOpcodes); err != nil {
		return err
	}
	if p.PolicyHash == "" {
		return errors.New("AVM cross-zone route policy hash is required")
	}
	if err := zonestypes.ValidateHash("AVM cross-zone route policy hash", p.PolicyHash); err != nil {
		return err
	}
	if p.PolicyHash != ComputeAVMCrossZoneRoutePolicyHash(p) {
		return errors.New("AVM cross-zone route policy hash mismatch")
	}
	return nil
}

func (e AVMCrossZoneExecution) Validate() error {
	e = canonicalAVMCrossZoneExecution(e)
	if err := ValidateAVMCrossZoneRoute(e.Message, e.RoutePolicy); err != nil {
		return err
	}
	if err := e.DestinationQueueEntry.Validate(); err != nil {
		return err
	}
	if e.DestinationQueueEntry.ZoneID != e.Message.DestinationZone ||
		e.DestinationQueueEntry.MessageID != e.Message.ID {
		return errors.New("AVM cross-zone destination queue entry mismatch")
	}
	if err := e.DestinationReceipt.Validate(); err != nil {
		return err
	}
	if e.DestinationReceipt.MessageID != e.Message.ID ||
		e.DestinationReceipt.ZoneID != e.Message.DestinationZone {
		return errors.New("AVM cross-zone destination receipt mismatch")
	}
	if e.DirectStateWriteAttempted {
		return errors.New("AVM cross-zone execution forbids direct state writes")
	}
	for _, item := range []struct {
		name  string
		value string
	}{
		{name: "AVM cross-zone source output messages root", value: e.SourceOutputMessagesRoot},
		{name: "AVM cross-zone destination receipt root", value: e.DestinationReceiptRoot},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if err := e.validateValueAccounting(); err != nil {
		return err
	}
	if err := e.validateFailureResolution(); err != nil {
		return err
	}
	if e.ExecutionHash == "" {
		return errors.New("AVM cross-zone execution hash is required")
	}
	if err := zonestypes.ValidateHash("AVM cross-zone execution hash", e.ExecutionHash); err != nil {
		return err
	}
	if e.ExecutionHash != ComputeAVMCrossZoneExecutionHash(e) {
		return errors.New("AVM cross-zone execution hash mismatch")
	}
	return nil
}

func (e AVMCrossZoneExecution) validateValueAccounting() error {
	if e.Message.ValueNAET == 0 {
		return nil
	}
	switch e.RoutePolicy.ValueAccounting {
	case AVMCrossZoneValueEscrow:
		if e.ValueEscrowedNAET != e.Message.ValueNAET {
			return errors.New("AVM cross-zone escrow value accounting mismatch")
		}
	case AVMCrossZoneValueMessage:
		if e.ValueAccountedNAET != e.Message.ValueNAET {
			return errors.New("AVM cross-zone message value accounting mismatch")
		}
	case AVMCrossZoneValueEscrowPlus:
		if e.ValueEscrowedNAET+e.ValueAccountedNAET < e.Message.ValueNAET || e.ValueEscrowedNAET+e.ValueAccountedNAET < e.ValueEscrowedNAET {
			return errors.New("AVM cross-zone combined value accounting mismatch")
		}
	default:
		return errors.New("AVM cross-zone value transfer requires escrow or message value accounting")
	}
	return nil
}

func (e AVMCrossZoneExecution) validateFailureResolution() error {
	status := e.DestinationReceipt.Status
	if !isCrossZoneFailureStatus(status) {
		if e.FailureResolution != AVMCrossZoneFailureNone {
			return errors.New("AVM cross-zone successful execution must not carry failure resolution")
		}
		return nil
	}
	switch e.FailureResolution {
	case AVMCrossZoneFailureBounced:
		if e.RoutePolicy.BounceBehavior == AVMCrossZoneBounceDisabled {
			return errors.New("AVM cross-zone bounce is disabled")
		}
		return validateCrossZoneBounceMessage(e.Message, e.BounceMessageOptional)
	case AVMCrossZoneFailureDeadLettered:
		if e.RoutePolicy.BounceBehavior == AVMCrossZoneBounceRequired && e.Message.BounceFlag {
			return errors.New("AVM cross-zone failure must bounce when bounce is required")
		}
		if e.DeadLetterOptional.MessageID != e.Message.ID {
			return errors.New("AVM cross-zone dead letter message mismatch")
		}
		if e.DeadLetterOptional.ZoneID != e.Message.DestinationZone {
			return errors.New("AVM cross-zone dead letter zone mismatch")
		}
		if e.DestinationReceipt.Status == AVMReceiptStatusDeadLettered {
			return e.DeadLetterOptional.ValidateWithReceipt(e.DestinationReceipt)
		}
		return e.DeadLetterOptional.Validate()
	default:
		return errors.New("failed AVM cross-zone execution must bounce or dead-letter")
	}
}

func IsAVMCrossZoneBounceBehavior(value AVMCrossZoneBounceBehavior) bool {
	switch value {
	case AVMCrossZoneBounceDisabled, AVMCrossZoneBounceAllowed, AVMCrossZoneBounceRequired:
		return true
	default:
		return false
	}
}

func IsAVMCrossZoneProofRequirement(value AVMCrossZoneProofRequirement) bool {
	switch value {
	case AVMCrossZoneProofNone, AVMCrossZoneProofAuth, AVMCrossZoneProofState, AVMCrossZoneProofAuthAndState:
		return true
	default:
		return false
	}
}

func IsAVMCrossZoneValueAccounting(value AVMCrossZoneValueAccounting) bool {
	switch value {
	case AVMCrossZoneValueNone, AVMCrossZoneValueEscrow, AVMCrossZoneValueMessage, AVMCrossZoneValueEscrowPlus:
		return true
	default:
		return false
	}
}

func ComputeAVMCrossZoneRoutePolicyHash(policy AVMCrossZoneRoutePolicy) string {
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-cross-zone-route-policy-v1")
	writeEnginePart(h, string(policy.SourceZone))
	writeEnginePart(h, string(policy.DestinationZone))
	writeEnginePart(h, policy.GasPolicy.Denom)
	writeEngineUint64(h, policy.GasPolicy.MaxGasPerBlock)
	writeEngineUint64(h, policy.GasPolicy.MaxGasPerMessage)
	writeEngineUint64(h, policy.ExecutionBudget.MaxGas)
	writeEngineUint64(h, policy.ExecutionBudget.GasUsed)
	writeEngineUint64(h, uint64(policy.ExecutionBudget.MaxMessages))
	writeEngineUint64(h, uint64(policy.ExecutionBudget.MessagesUsed))
	writeEngineUint64(h, uint64(len(policy.MessageFilter.AllowedMessageTypes)))
	for _, item := range policy.MessageFilter.AllowedMessageTypes {
		writeEnginePart(h, item)
	}
	writeEngineUint64(h, uint64(len(policy.AllowedOpcodes)))
	for _, opcode := range policy.AllowedOpcodes {
		writeEnginePart(h, opcode)
	}
	writeEnginePart(h, string(policy.BounceBehavior))
	writeEnginePart(h, string(policy.ProofRequirement))
	writeEnginePart(h, string(policy.ValueAccounting))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneExecutionHash(execution AVMCrossZoneExecution) string {
	execution = canonicalAVMCrossZoneExecution(execution)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-cross-zone-execution-v1")
	writeEnginePart(h, execution.Message.ID)
	writeEnginePart(h, execution.RoutePolicy.PolicyHash)
	writeEnginePart(h, execution.DestinationQueueEntry.SortKey)
	writeEnginePart(h, execution.DestinationReceipt.ReceiptHash)
	writeEnginePart(h, execution.SourceOutputMessagesRoot)
	writeEnginePart(h, execution.DestinationReceiptRoot)
	writeEngineBool(h, execution.DirectStateWriteAttempted)
	writeEngineUint64(h, execution.ValueEscrowedNAET)
	writeEngineUint64(h, execution.ValueAccountedNAET)
	writeEnginePart(h, string(execution.FailureResolution))
	writeEnginePart(h, execution.BounceMessageOptional.ID)
	writeEnginePart(h, execution.DeadLetterOptional.ReceiptID)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMCrossZoneRoutePolicy(policy AVMCrossZoneRoutePolicy) AVMCrossZoneRoutePolicy {
	policy.PolicyHash = strings.TrimSpace(policy.PolicyHash)
	policy.AllowedOpcodes = append([]string(nil), policy.AllowedOpcodes...)
	for i := range policy.AllowedOpcodes {
		policy.AllowedOpcodes[i] = strings.TrimSpace(policy.AllowedOpcodes[i])
	}
	sort.Strings(policy.AllowedOpcodes)
	policy.MessageFilter.AllowedMessageTypes = append([]string(nil), policy.MessageFilter.AllowedMessageTypes...)
	for i := range policy.MessageFilter.AllowedMessageTypes {
		policy.MessageFilter.AllowedMessageTypes[i] = strings.TrimSpace(policy.MessageFilter.AllowedMessageTypes[i])
	}
	sort.Strings(policy.MessageFilter.AllowedMessageTypes)
	if len(policy.MessageFilter.AllowedMessageTypes) == 1 && policy.MessageFilter.AllowedMessageTypes[0] == "*" {
		return policy
	}
	return policy
}

func canonicalAVMCrossZoneExecution(execution AVMCrossZoneExecution) AVMCrossZoneExecution {
	execution.Message = canonicalAVMAsyncMessage(execution.Message)
	execution.RoutePolicy = canonicalAVMCrossZoneRoutePolicy(execution.RoutePolicy)
	execution.DestinationQueueEntry = canonicalAVMZoneQueueEntry(execution.DestinationQueueEntry)
	execution.DestinationReceipt = canonicalAVMExecutionReceipt(execution.DestinationReceipt)
	execution.SourceOutputMessagesRoot = strings.TrimSpace(execution.SourceOutputMessagesRoot)
	execution.DestinationReceiptRoot = strings.TrimSpace(execution.DestinationReceiptRoot)
	execution.BounceMessageOptional = canonicalAVMAsyncMessage(execution.BounceMessageOptional)
	execution.DeadLetterOptional = canonicalAVMDeadLetterRecord(execution.DeadLetterOptional)
	execution.ExecutionHash = strings.TrimSpace(execution.ExecutionHash)
	if execution.FailureResolution == "" {
		execution.FailureResolution = AVMCrossZoneFailureNone
	}
	return execution
}

func validateCrossZoneOpcodes(opcodes []string) error {
	if len(opcodes) == 0 {
		return errors.New("AVM cross-zone policy must define allowed opcodes")
	}
	seen := make(map[string]struct{}, len(opcodes))
	for i, opcode := range opcodes {
		if err := validateRouterOptionalToken("AVM cross-zone opcode", opcode, MaxAVMCrossZoneOpcodeLength); err != nil {
			return err
		}
		if opcode == "" {
			return errors.New("AVM cross-zone opcode is required")
		}
		if _, found := seen[opcode]; found {
			return fmt.Errorf("duplicate AVM cross-zone opcode %q", opcode)
		}
		seen[opcode] = struct{}{}
		if i > 0 && opcodes[i-1] >= opcode {
			return errors.New("AVM cross-zone opcodes must be sorted canonically")
		}
	}
	return nil
}

func validateCrossZoneProofs(msg AVMAsyncMessage, requirement AVMCrossZoneProofRequirement) error {
	switch requirement {
	case AVMCrossZoneProofNone:
		return nil
	case AVMCrossZoneProofAuth:
		if msg.AuthProofOptional == "" {
			return errors.New("AVM cross-zone route requires auth proof")
		}
	case AVMCrossZoneProofState:
		if msg.StateProofOptional == "" {
			return errors.New("AVM cross-zone route requires state proof")
		}
	case AVMCrossZoneProofAuthAndState:
		if msg.AuthProofOptional == "" || msg.StateProofOptional == "" {
			return errors.New("AVM cross-zone route requires auth and state proofs")
		}
	default:
		return fmt.Errorf("invalid AVM cross-zone proof requirement %q", requirement)
	}
	return nil
}

func crossZoneOpcodeAllowed(opcodes []string, opcode string) bool {
	for _, allowed := range opcodes {
		if allowed == opcode {
			return true
		}
	}
	return false
}

func validateCrossZoneBounceMessage(original AVMAsyncMessage, bounce AVMAsyncMessage) error {
	bounce = canonicalAVMAsyncMessage(bounce)
	if err := bounce.Validate(); err != nil {
		return err
	}
	if bounce.SourceZone != original.DestinationZone || bounce.DestinationZone != original.SourceZone {
		return errors.New("AVM cross-zone bounce message must reverse source and destination zones")
	}
	if !strings.Contains(bounce.RouteHintOptional, original.ID) {
		return errors.New("AVM cross-zone bounce message must reference original message")
	}
	return nil
}

func isCrossZoneFailureStatus(status AVMReceiptStatus) bool {
	switch status {
	case AVMReceiptStatusFailed, AVMReceiptStatusBounced, AVMReceiptStatusDeadLettered:
		return true
	default:
		return false
	}
}
