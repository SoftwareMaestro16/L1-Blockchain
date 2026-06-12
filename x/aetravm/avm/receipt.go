package avm

import (
	"encoding/binary"
	"fmt"

	"lukechampine.com/blake3"
)

// AVMLedgerReceipt is the enhanced receipt with full accounting,
// message state flags, bounce tracking, and value conservation proof.
type AVMLedgerReceipt struct {
	ExitCode		StructuredExitCode
	GasUsed			uint64
	GasRefunded		uint64
	GasBreakdown		GasBreakdown
	StorageFee		uint64
	ValueIn			uint64
	ValueOut		uint64
	StateRootBefore		[]byte
	StateRootAfter		[]byte
	EmittedActionsHash	[]byte
	EventsHash		[]byte
	ProofHash		[]byte
	MessageFlags		MessageFlags
	BounceMessage		*BounceMessage
	Actions			[]Action
	Events			[]EventRecord
}

// GasBreakdown splits gas into compute, storage, message, and bounce components.
type GasBreakdown struct {
	ComputeGas	uint64
	StorageGas	uint64
	MessageGas	uint64
	BounceGas	uint64
}

func (g GasBreakdown) Total() uint64 {
	return g.ComputeGas + g.StorageGas + g.MessageGas + g.BounceGas
}

// MessageFlags tracks the lifecycle state of a message.
type MessageFlags struct {
	Consumed	bool
	Bounced		bool
	BounceRequested	bool
	RefundIssued	bool
	RefundLocked	bool
}

func (f MessageFlags) String() string {
	return fmt.Sprintf("consumed=%v bounced=%v bounce_requested=%v refund_issued=%v refund_locked=%v",
		f.Consumed, f.Bounced, f.BounceRequested, f.RefundIssued, f.RefundLocked)
}

type MessageState uint8

const (
	MessagePending	MessageState	= iota
	MessageExecuted
	MessageFailed
	MessageBounced
	MessageFinalized
)

func (s MessageState) String() string {
	switch s {
	case MessagePending:
		return "PENDING"
	case MessageExecuted:
		return "EXECUTED"
	case MessageFailed:
		return "FAILED"
	case MessageBounced:
		return "BOUNCED"
	case MessageFinalized:
		return "FINALIZED"
	default:
		return "UNKNOWN"
	}
}

type BounceMessage struct {
	OriginalMessageHash	[]byte
	FailureExitCode		StructuredExitCode
	BounceFlag		bool
	PartialSnapshot		[]byte
}

func NewBounceMessage(originalHash []byte, exitCode StructuredExitCode) *BounceMessage {
	return &BounceMessage{
		OriginalMessageHash:	originalHash,
		FailureExitCode:	exitCode,
		BounceFlag:		true,
	}
}

type EventRecord struct {
	Index		uint32
	Topic		string
	Payload		[]byte
	Sender		string
	Contract	string
}

// ComputeEventsHash computes BLAKE3(canonical_event_list) deterministically.
func ComputeEventsHash(events []EventRecord) []byte {
	h := blake3.New(32, nil)
	for _, e := range events {
		h.Write([]byte{byte(e.Index >> 24), byte(e.Index >> 16), byte(e.Index >> 8), byte(e.Index)})
		h.Write([]byte(e.Topic))
		h.Write(e.Payload)
		h.Write([]byte(e.Sender))
		h.Write([]byte(e.Contract))
	}
	return h.Sum(nil)
}

type ValueConservationProof struct {
	ValueIn			uint64
	ValueOut		uint64
	StorageFeePaid		uint64
	RefundIssued		uint64
	RemainingBalanceDelta	int64
	Balanced		bool
}

// VerifyValueConservation checks that value_in - value_out - refunds - storage_fee == remaining_delta.
func VerifyValueConservation(receipt *AVMLedgerReceipt) ValueConservationProof {
	proof := ValueConservationProof{
		ValueIn:	receipt.ValueIn,
		ValueOut:	receipt.ValueOut,
		StorageFeePaid:	receipt.StorageFee,
		RefundIssued:	receipt.GasRefunded,
	}

	expected := int64(receipt.ValueOut) + int64(receipt.StorageFee) + int64(receipt.GasRefunded)
	actual := int64(receipt.ValueIn)
	proof.RemainingBalanceDelta = actual - expected
	proof.Balanced = proof.RemainingBalanceDelta == 0

	return proof
}

func IsBounceEligible(flags MessageFlags, exitCode StructuredExitCode) bool {
	if flags.Bounced {
		return false
	}
	if !flags.BounceRequested {
		return false
	}
	if exitCode.Category == ExitCategorySuccess {
		return false
	}
	return true
}

// ProcessBounce creates a bounce message if eligible.
// Returns nil if not eligible (no bounce loop possible).
func ProcessBounce(originalMessage *Message, flags MessageFlags, exitCode StructuredExitCode) *BounceMessage {
	if !IsBounceEligible(flags, exitCode) {
		return nil
	}

	if originalMessage == nil || len(originalMessage.Hash) == 0 {
		return NewBounceMessage([]byte{}, exitCode)
	}

	return NewBounceMessage(originalMessage.Hash, exitCode)
}

type RefundAccounting struct {
	GasRefunded	uint64
	RefundIssued	bool
	RefundLocked	bool
}

func NewRefundAccounting() *RefundAccounting {
	return &RefundAccounting{}
}

// IssueRefund attempts to issue a gas refund.
// Returns error if refund was already issued or if refund is locked.
func (r *RefundAccounting) IssueRefund(amount uint64) error {
	if r.RefundIssued {
		return fmt.Errorf("AVM refund: already issued, cannot issue second refund")
	}
	if r.RefundLocked {
		return fmt.Errorf("AVM refund: refund is locked, cannot issue")
	}
	r.GasRefunded = amount
	r.RefundIssued = true
	return nil
}

// LockRefund prevents any future refund issuance.
func (r *RefundAccounting) LockRefund() {
	r.RefundLocked = true
}

// BuildLedgerReceipt constructs a full ledger receipt from execution results.
func BuildLedgerReceipt(
	exitCode StructuredExitCode,
	frame *KernelExecutionFrame,
	refund *RefundAccounting,
	bounce *BounceMessage,
	events []EventRecord,
) *AVMLedgerReceipt {
	var stateRootBefore, stateRootAfter []byte
	if frame.StateSnapshot != nil {
		stateRootBefore = frame.StateSnapshot.Hash()
	}
	workingState := frame.WorkingState
	if workingState == nil {
		workingState = frame.StateSnapshot
	}
	if workingState != nil {
		stateRootAfter = workingState.Hash()
	}

	actionsHash := []byte{}
	if frame.ActionQueue != nil && frame.ActionQueue.Hash != nil {
		actionsHash = frame.ActionQueue.Hash
	}

	eventsHash := ComputeEventsHash(events)

	gasBreakdown := GasBreakdown{
		ComputeGas:	frame.PhaseGas[PhaseCompute] + frame.PhaseGas[PhaseStorage] + frame.PhaseGas[PhaseCredit],
		StorageGas:	frame.PhaseGas[PhaseFinalization],
		MessageGas:	frame.PhaseGas[PhaseAction],
		BounceGas:	0,
	}
	if bounce != nil {
		gasBreakdown.BounceGas = 100
	}

	storageFee := uint64(0)
	if frame.PhaseGas[PhaseFinalization] > 0 {
		storageFee = frame.PhaseGas[PhaseFinalization]
	}

	gasRefunded := uint64(0)
	if refund != nil {
		gasRefunded = refund.GasRefunded
	}

	return &AVMLedgerReceipt{
		ExitCode:		exitCode,
		GasUsed:		frame.GasUsed,
		GasRefunded:		gasRefunded,
		GasBreakdown:		gasBreakdown,
		StorageFee:		storageFee,
		ValueIn:		frame.Message.Value,
		ValueOut:		sumActionValues(frame),
		StateRootBefore:	stateRootBefore,
		StateRootAfter:		stateRootAfter,
		EmittedActionsHash:	actionsHash,
		EventsHash:		eventsHash,
		ProofHash:		computeReceiptProofHash(frame, events),
		MessageFlags:		buildMessageFlags(frame, bounce),
		BounceMessage:		bounce,
		Actions:		getFrameActions(frame),
		Events:			events,
	}
}

func sumActionValues(frame *KernelExecutionFrame) uint64 {
	if frame.ActionQueue == nil {
		return 0
	}
	total := uint64(0)
	for _, a := range frame.ActionQueue.Actions {
		total += a.Value
	}
	return total
}

func getFrameActions(frame *KernelExecutionFrame) []Action {
	if frame.ActionQueue == nil {
		return nil
	}
	return frame.ActionQueue.Actions
}

func buildMessageFlags(frame *KernelExecutionFrame, bounce *BounceMessage) MessageFlags {
	flags := MessageFlags{
		Consumed: true,
	}
	if bounce != nil {
		flags.Bounced = true
		flags.BounceRequested = true
	}
	return flags
}

func computeReceiptProofHash(frame *KernelExecutionFrame, events []EventRecord) []byte {
	h := blake3.New(32, nil)
	var gasUsed [8]byte
	binary.BigEndian.PutUint64(gasUsed[:], frame.GasUsed)
	h.Write(gasUsed[:])
	for _, step := range frame.Trace.Steps {
		h.Write([]byte{byte(step.Opcode >> 8), byte(step.Opcode)})
	}
	return h.Sum(nil)
}

func (r *AVMLedgerReceipt) CanonicalEncode() ([]byte, error) {
	buf := make([]byte, 0, 512)

	buf = append(buf, byte(r.ExitCode.Category))
	buf = binary.BigEndian.AppendUint16(buf, r.ExitCode.Subcode)
	buf = binary.BigEndian.AppendUint64(buf, r.GasUsed)
	buf = binary.BigEndian.AppendUint64(buf, r.GasRefunded)
	buf = binary.BigEndian.AppendUint64(buf, r.GasBreakdown.ComputeGas)
	buf = binary.BigEndian.AppendUint64(buf, r.GasBreakdown.StorageGas)
	buf = binary.BigEndian.AppendUint64(buf, r.GasBreakdown.MessageGas)
	buf = binary.BigEndian.AppendUint64(buf, r.GasBreakdown.BounceGas)
	buf = binary.BigEndian.AppendUint64(buf, r.StorageFee)
	buf = binary.BigEndian.AppendUint64(buf, r.ValueIn)
	buf = binary.BigEndian.AppendUint64(buf, r.ValueOut)

	buf = append(buf, byte(len(r.StateRootBefore)))
	if len(r.StateRootBefore) > 0 {
		buf = append(buf, r.StateRootBefore...)
	}
	buf = append(buf, byte(len(r.StateRootAfter)))
	if len(r.StateRootAfter) > 0 {
		buf = append(buf, r.StateRootAfter...)
	}

	buf = append(buf, byte(len(r.EmittedActionsHash)))
	if len(r.EmittedActionsHash) > 0 {
		buf = append(buf, r.EmittedActionsHash...)
	}
	buf = append(buf, byte(len(r.EventsHash)))
	if len(r.EventsHash) > 0 {
		buf = append(buf, r.EventsHash...)
	}

	buf = appendMsgFlags(buf, r.MessageFlags)

	if r.BounceMessage != nil {
		buf = append(buf, 0x01)
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.BounceMessage.OriginalMessageHash)))
		buf = append(buf, r.BounceMessage.OriginalMessageHash...)
		buf = append(buf, byte(r.BounceMessage.FailureExitCode.Category))
		buf = binary.BigEndian.AppendUint16(buf, r.BounceMessage.FailureExitCode.Subcode)
	} else {
		buf = append(buf, 0x00)
	}

	return buf, nil
}

func appendMsgFlags(buf []byte, f MessageFlags) []byte {
	var flags byte
	if f.Consumed {
		flags |= 0x01
	}
	if f.Bounced {
		flags |= 0x02
	}
	if f.BounceRequested {
		flags |= 0x04
	}
	if f.RefundIssued {
		flags |= 0x08
	}
	if f.RefundLocked {
		flags |= 0x10
	}
	return append(buf, flags)
}

func TransitionMessageState(current MessageState, success bool, bounceEligible bool) (MessageState, MessageFlags) {
	flags := MessageFlags{Consumed: true}

	switch current {
	case MessagePending:
		if success {
			return MessageExecuted, flags
		}
		return MessageFailed, flags
	case MessageFailed:
		if bounceEligible {
			flags.BounceRequested = true
			return MessageBounced, flags
		}
		return MessageFinalized, flags
	case MessageBounced:
		flags.Bounced = true
		return MessageFinalized, flags
	default:
		return MessageFinalized, flags
	}
}

func ValidateNoDoubleRefund(flags MessageFlags) error {
	if flags.RefundIssued {
		return fmt.Errorf("AVM refund: refund already issued, cannot issue second")
	}
	return nil
}

func ReceiptHash(receipt *AVMLedgerReceipt) ([32]byte, error) {
	encoded, err := receipt.CanonicalEncode()
	if err != nil {
		return [32]byte{}, err
	}
	return blake3.Sum256(encoded), nil
}
