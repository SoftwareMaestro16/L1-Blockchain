package avm

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

func TestMessageFlagsString(t *testing.T) {
	flags := MessageFlags{
		Consumed:		true,
		Bounced:		false,
		BounceRequested:	true,
		RefundIssued:		false,
		RefundLocked:		false,
	}
	s := flags.String()
	if s == "" {
		t.Error("flags string should not be empty")
	}
}

func TestMessageStateTransitionPendingSuccess(t *testing.T) {
	state, flags := TransitionMessageState(MessagePending, true, false)
	if state != MessageExecuted {
		t.Errorf("expected EXECUTED, got %v", state)
	}
	if !flags.Consumed {
		t.Error("message should be consumed")
	}
}

func TestMessageStateTransitionPendingFailure(t *testing.T) {
	state, flags := TransitionMessageState(MessagePending, false, false)
	if state != MessageFailed {
		t.Errorf("expected FAILED, got %v", state)
	}
	if !flags.Consumed {
		t.Error("message should be consumed")
	}
}

func TestMessageStateTransitionFailedBounce(t *testing.T) {
	state, flags := TransitionMessageState(MessageFailed, false, true)
	if state != MessageBounced {
		t.Errorf("expected BOUNCED, got %v", state)
	}
	if !flags.BounceRequested {
		t.Error("bounce should be requested")
	}
}

func TestMessageStateTransitionFailedNoBounce(t *testing.T) {
	state, _ := TransitionMessageState(MessageFailed, false, false)
	if state != MessageFinalized {
		t.Errorf("expected FINALIZED, got %v", state)
	}
}

func TestMessageStateTransitionBouncedFinalized(t *testing.T) {
	state, flags := TransitionMessageState(MessageBounced, false, true)
	if state != MessageFinalized {
		t.Errorf("expected FINALIZED, got %v", state)
	}
	if !flags.Bounced {
		t.Error("should be bounced")
	}
}

func TestMessageStateStrings(t *testing.T) {
	tests := map[MessageState]string{
		MessagePending:		"PENDING",
		MessageExecuted:	"EXECUTED",
		MessageFailed:		"FAILED",
		MessageBounced:		"BOUNCED",
		MessageFinalized:	"FINALIZED",
	}
	for state, expected := range tests {
		if state.String() != expected {
			t.Errorf("expected %s, got %s", expected, state.String())
		}
	}
}

func TestBounceEligibleFailedBounceRequested(t *testing.T) {
	flags := MessageFlags{
		Consumed:		true,
		Bounced:		false,
		BounceRequested:	true,
	}
	exitCode := StructuredExitCode{Category: ExitCategoryVMError, Subcode: 1}

	if !IsBounceEligible(flags, exitCode) {
		t.Error("failed bounceable message should be eligible for bounce")
	}
}

func TestBounceNotEligibleSuccess(t *testing.T) {
	flags := MessageFlags{
		Consumed:		true,
		Bounced:		false,
		BounceRequested:	true,
	}
	exitCode := ExitSuccess

	if IsBounceEligible(flags, exitCode) {
		t.Error("successful message should not be eligible for bounce")
	}
}

func TestBounceNotEligibleAlreadyBounced(t *testing.T) {
	flags := MessageFlags{
		Consumed:		true,
		Bounced:		true,
		BounceRequested:	true,
	}
	exitCode := StructuredExitCode{Category: ExitCategoryVMError, Subcode: 1}

	if IsBounceEligible(flags, exitCode) {
		t.Error("already bounced message should not be eligible for bounce again")
	}
}

func TestBounceNotEligibleNotRequested(t *testing.T) {
	flags := MessageFlags{
		Consumed:		true,
		Bounced:		false,
		BounceRequested:	false,
	}
	exitCode := StructuredExitCode{Category: ExitCategoryVMError, Subcode: 1}

	if IsBounceEligible(flags, exitCode) {
		t.Error("non-bounceable message should not be eligible for bounce")
	}
}

func TestProcessBounceCreatesBounceMessage(t *testing.T) {
	msg := &Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		100,
		GasLimit:	50000,
		Hash:		[]byte{1, 2, 3, 4},
	}
	flags := MessageFlags{
		Consumed:		true,
		BounceRequested:	true,
		Bounced:		false,
	}
	exitCode := StructuredExitCode{Category: ExitCategoryVMError, Subcode: 1}

	bounce := ProcessBounce(msg, flags, exitCode)
	if bounce == nil {
		t.Fatal("bounce should be created for eligible message")
	}
	if !bounce.BounceFlag {
		t.Error("bounce flag should be true")
	}
	if len(bounce.OriginalMessageHash) == 0 {
		t.Error("original message hash should not be empty")
	}
}

func TestProcessBounceNotEligible(t *testing.T) {
	msg := &Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		100,
		GasLimit:	50000,
	}
	flags := MessageFlags{
		Consumed:		true,
		BounceRequested:	false,
		Bounced:		false,
	}
	exitCode := StructuredExitCode{Category: ExitCategoryVMError, Subcode: 1}

	bounce := ProcessBounce(msg, flags, exitCode)
	if bounce != nil {
		t.Error("non-bounceable message should NOT create a bounce message")
	}
}

func TestProcessBounceAlreadyBouncedNoLoop(t *testing.T) {
	msg := &Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		100,
		GasLimit:	50000,
		Hash:		[]byte{1, 2, 3, 4},
	}
	flags := MessageFlags{
		Consumed:		true,
		BounceRequested:	true,
		Bounced:		true,
	}
	exitCode := StructuredExitCode{Category: ExitCategoryVMError, Subcode: 1}

	bounce := ProcessBounce(msg, flags, exitCode)
	if bounce != nil {
		t.Error("already bounced message should NOT create another bounce (prevents infinite loop)")
	}
}

func TestRefundIssuedOnce(t *testing.T) {
	acc := NewRefundAccounting()
	err := acc.IssueRefund(500)
	if err != nil {
		t.Fatalf("first refund should succeed: %v", err)
	}
	if acc.GasRefunded != 500 {
		t.Errorf("expected 500 refunded, got %d", acc.GasRefunded)
	}
	if !acc.RefundIssued {
		t.Error("refund should be marked as issued")
	}
}

func TestDoubleRefundRejected(t *testing.T) {
	acc := NewRefundAccounting()
	acc.IssueRefund(500)

	err := acc.IssueRefund(300)
	if err == nil {
		t.Error("second refund should be rejected (double refund prevention)")
	}
}

func TestRefundLockedRejected(t *testing.T) {
	acc := NewRefundAccounting()
	acc.LockRefund()

	err := acc.IssueRefund(500)
	if err == nil {
		t.Error("locked refund should be rejected")
	}
}

func TestNoDoubleRefundValidation(t *testing.T) {
	flags := MessageFlags{RefundIssued: true}
	err := ValidateNoDoubleRefund(flags)
	if err == nil {
		t.Error("should reject double refund")
	}

	flags2 := MessageFlags{RefundIssued: false}
	err = ValidateNoDoubleRefund(flags2)
	if err != nil {
		t.Errorf("should accept first refund: %v", err)
	}
}

func TestComputeEventsHash(t *testing.T) {
	events := []EventRecord{
		{Index: 0, Topic: "transfer", Payload: []byte{1, 2, 3}, Sender: "addr1", Contract: "contract1"},
		{Index: 1, Topic: "approve", Payload: []byte{4, 5, 6}, Sender: "addr2", Contract: "contract1"},
	}

	hash1 := ComputeEventsHash(events)
	if len(hash1) == 0 {
		t.Error("events hash should not be empty")
	}

	hash2 := ComputeEventsHash(events)
	if string(hash1) != string(hash2) {
		t.Error("same events should produce same hash (deterministic)")
	}
}

func TestEventsHashChangesOnReorder(t *testing.T) {
	events1 := []EventRecord{
		{Index: 0, Topic: "transfer", Payload: []byte{1}, Sender: "a", Contract: "c"},
		{Index: 1, Topic: "approve", Payload: []byte{2}, Sender: "b", Contract: "c"},
	}
	events2 := []EventRecord{
		{Index: 1, Topic: "approve", Payload: []byte{2}, Sender: "b", Contract: "c"},
		{Index: 0, Topic: "transfer", Payload: []byte{1}, Sender: "a", Contract: "c"},
	}

	hash1 := ComputeEventsHash(events1)
	hash2 := ComputeEventsHash(events2)
	if string(hash1) == string(hash2) {
		t.Error("different event orders should produce different hashes")
	}
}

func TestValueConservationBalanced(t *testing.T) {
	receipt := &AVMLedgerReceipt{
		ValueIn:	1000,
		ValueOut:	600,
		StorageFee:	200,
		GasRefunded:	200,
	}

	proof := VerifyValueConservation(receipt)
	if !proof.Balanced {
		t.Error("receipt should be balanced: 1000 = 600 + 200 + 200")
	}
}

func TestValueConservationImbalanced(t *testing.T) {
	receipt := &AVMLedgerReceipt{
		ValueIn:	1000,
		ValueOut:	900,
		StorageFee:	200,
		GasRefunded:	0,
	}

	proof := VerifyValueConservation(receipt)
	if proof.Balanced {
		t.Error("receipt should NOT be balanced: 1000 != 900 + 200 + 0")
	}
}

func TestGasBreakdownTotal(t *testing.T) {
	gb := GasBreakdown{
		ComputeGas:	100,
		StorageGas:	50,
		MessageGas:	30,
		BounceGas:	20,
	}
	if gb.Total() != 200 {
		t.Errorf("expected total 200, got %d", gb.Total())
	}
}

func testChunk() *chunk.Chunk {
	m := chunk.NewEmptyMap()
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeSystem).SetData([]byte{}, 0).Build()
	m, _ = m.Put([]byte("__init__"), b)
	return m.Root()
}

func TestBuildLedgerReceipt(t *testing.T) {
	state := testChunk()
	msg := Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		1000,
		GasLimit:	100000,
	}
	frame := NewKernelExecutionFrame(state, msg, 100)
	refund := NewRefundAccounting()
	events := []EventRecord{
		{Index: 0, Topic: "transfer", Payload: []byte{1}, Sender: "sender1", Contract: "target1"},
	}

	receipt := BuildLedgerReceipt(ExitSuccess, frame, refund, nil, events)
	if receipt.ExitCode.Category != ExitCategorySuccess {
		t.Errorf("expected success, got %v", receipt.ExitCode.Category)
	}
	if receipt.MessageFlags.Consumed != true {
		t.Error("message should be consumed")
	}
	if receipt.BounceMessage != nil {
		t.Error("no bounce expected for successful execution")
	}
}

func TestBuildLedgerReceiptWithBounce(t *testing.T) {
	state := testChunk()
	msg := Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		1000,
		GasLimit:	100000,
		Hash:		[]byte{1, 2, 3, 4},
	}
	frame := NewKernelExecutionFrame(state, msg, 100)
	refund := NewRefundAccounting()
	bounce := NewBounceMessage([]byte{1, 2, 3, 4}, ExitValidationFailed)

	receipt := BuildLedgerReceipt(ExitValidationFailed, frame, refund, bounce, nil)
	if receipt.BounceMessage == nil {
		t.Error("bounce message should be present")
	}
	if !receipt.MessageFlags.Bounced {
		t.Error("message flags should indicate bounced")
	}
	if receipt.GasBreakdown.BounceGas != 100 {
		t.Errorf("expected bounce gas 100, got %d", receipt.GasBreakdown.BounceGas)
	}
}

func TestNewBounceMessage(t *testing.T) {
	bm := NewBounceMessage([]byte{0xAA, 0xBB}, ExitStackOverflow)
	if !bm.BounceFlag {
		t.Error("bounce flag should be true")
	}
	if len(bm.OriginalMessageHash) != 2 {
		t.Errorf("expected 2 byte hash, got %d", len(bm.OriginalMessageHash))
	}
	if bm.FailureExitCode.Category != ExitCategoryVMError {
		t.Errorf("expected VM_ERROR, got %v", bm.FailureExitCode.Category)
	}
}

func TestReceiptCanonicalEncode(t *testing.T) {
	receipt := &AVMLedgerReceipt{
		ExitCode:		ExitSuccess,
		GasUsed:		5000,
		GasRefunded:		200,
		GasBreakdown:		GasBreakdown{ComputeGas: 4000, StorageGas: 500, MessageGas: 300, BounceGas: 200},
		StorageFee:		500,
		ValueIn:		1000,
		ValueOut:		600,
		StateRootBefore:	make([]byte, 32),
		StateRootAfter:		make([]byte, 32),
		EmittedActionsHash:	make([]byte, 32),
		EventsHash:		make([]byte, 32),
		MessageFlags:		MessageFlags{Consumed: true},
	}

	encoded, err := receipt.CanonicalEncode()
	if err != nil {
		t.Fatalf("canonical encode: %v", err)
	}
	if len(encoded) == 0 {
		t.Error("encoded receipt should not be empty")
	}
}

func TestReceiptHashDeterministic(t *testing.T) {
	receipt := &AVMLedgerReceipt{
		ExitCode:	ExitSuccess,
		GasUsed:	5000,
		GasRefunded:	200,
		GasBreakdown:	GasBreakdown{ComputeGas: 4000, StorageGas: 500},
		ValueIn:	1000,
		ValueOut:	600,
		MessageFlags:	MessageFlags{Consumed: true},
	}

	h1, err := ReceiptHash(receipt)
	if err != nil {
		t.Fatalf("receipt hash: %v", err)
	}
	h2, err := ReceiptHash(receipt)
	if err != nil {
		t.Fatalf("receipt hash: %v", err)
	}
	if h1 != h2 {
		t.Error("same receipt should produce same hash")
	}
}
