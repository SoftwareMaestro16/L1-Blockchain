package avm

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

func TestContinuationSlotDefaults(t *testing.T) {
	slot := ContinuationSlot{}
	if slot.ReturnPtr != 0 {
		t.Errorf("expected ReturnPtr=0, got %d", slot.ReturnPtr)
	}
	if slot.AltReturnPtr != 0 {
		t.Errorf("expected AltReturnPtr=0, got %d", slot.AltReturnPtr)
	}
	if slot.ErrorHandlerPtr != 0 {
		t.Errorf("expected ErrorHandlerPtr=0, got %d", slot.ErrorHandlerPtr)
	}
	if slot.DispatcherPtr != 0 {
		t.Errorf("expected DispatcherPtr=0, got %d", slot.DispatcherPtr)
	}
}

func TestExecutionSlotString(t *testing.T) {
	slots := map[ExecutionSlot]string{
		SlotReturn:	"SLOT_RETURN",
		SlotAltReturn:	"SLOT_ALT_RETURN",
		SlotError:	"SLOT_ERROR",
		SlotDispatch:	"SLOT_DISPATCH",
		SlotState:	"SLOT_STATE",
		SlotActions:	"SLOT_ACTIONS",
		SlotEnv:	"SLOT_ENV",
	}
	for slot, expected := range slots {
		if slot.String() != expected {
			t.Errorf("expected %s, got %s", expected, slot.String())
		}
	}
	if ExecutionSlot(99).String() != "SLOT_UNKNOWN" {
		t.Errorf("expected SLOT_UNKNOWN for invalid slot, got %s", ExecutionSlot(99).String())
	}
}

func TestSetGetContinuation(t *testing.T) {
	frame := &KernelExecutionFrame{
		Stack:		make([]StackValue, 0),
		PhaseGas:	make(map[Phase]uint64),
		ActionQueue:	NewActionQueueChunk(),
	}

	frame.SetContinuation(SlotReturn, 42)
	frame.SetContinuation(SlotAltReturn, 100)
	frame.SetContinuation(SlotError, 200)
	frame.SetContinuation(SlotDispatch, 300)

	if frame.GetContinuation(SlotReturn) != 42 {
		t.Errorf("expected ReturnPtr=42, got %d", frame.GetContinuation(SlotReturn))
	}
	if frame.GetContinuation(SlotAltReturn) != 100 {
		t.Errorf("expected AltReturnPtr=100, got %d", frame.GetContinuation(SlotAltReturn))
	}
	if frame.GetContinuation(SlotError) != 200 {
		t.Errorf("expected ErrorHandlerPtr=200, got %d", frame.GetContinuation(SlotError))
	}
	if frame.GetContinuation(SlotDispatch) != 300 {
		t.Errorf("expected DispatcherPtr=300, got %d", frame.GetContinuation(SlotDispatch))
	}
	if frame.GetContinuation(SlotState) != 0 {
		t.Errorf("expected 0 for SlotState (not a continuation), got %d", frame.GetContinuation(SlotState))
	}
}

func TestStructuredExitCodeRoundTrip(t *testing.T) {
	tests := []StructuredExitCode{
		ExitSuccess,
		ExitValidationFailed,
		ExitStackOverflow,
		ExitTypeMismatch,
		ExitDivZero,
		ExitGasExhausted,
		ExitActionBudget,
	}

	for _, code := range tests {
		packed := code.ToUint32()
		unpacked := StructuredExitCodeFromUint32(packed)
		if unpacked.Category != code.Category {
			t.Errorf("category mismatch: expected %v, got %v", code.Category, unpacked.Category)
		}
		if unpacked.Subcode != code.Subcode {
			t.Errorf("subcode mismatch: expected %d, got %d", code.Subcode, unpacked.Subcode)
		}
	}
}

func TestExitCategoryStrings(t *testing.T) {
	categories := map[ExitCategory]string{
		ExitCategorySuccess:		"SUCCESS",
		ExitCategoryVMError:		"VM_ERROR",
		ExitCategoryTypeError:		"TYPE_ERROR",
		ExitCategoryExecutionError:	"EXECUTION_ERROR",
		ExitCategoryActionError:	"ACTION_ERROR",
		ExitCategoryStateError:		"STATE_ERROR",
		ExitCategoryGasError:		"GAS_ERROR",
	}
	for cat, expected := range categories {
		if cat.String() != expected {
			t.Errorf("expected %s, got %s", expected, cat.String())
		}
	}
}

func TestExitCodeNoStateMutation(t *testing.T) {
	errorCodes := []StructuredExitCode{
		ExitStackOverflow,
		ExitStackUnderflow,
		ExitTypeMismatch,
		ExitDivZero,
		ExitGasExhausted,
		ExitStateMutation,
	}
	for _, code := range errorCodes {
		if code.Category == ExitCategorySuccess {
			t.Errorf("error code %v should not be SUCCESS category", code)
		}
	}
}

func TestStackValueConstructors(t *testing.T) {
	intVal := StackValueInt256(-42)
	if intVal.Type != StackTypeInt256 || intVal.IntVal != -42 {
		t.Errorf("StackValueInt256 failed: type=%v, intVal=%d", intVal.Type, intVal.IntVal)
	}

	uintVal := StackValueUint256(100)
	if uintVal.Type != StackTypeInt256 || uintVal.UintVal != 100 {
		t.Errorf("StackValueUint256 failed: type=%v, uintVal=%d", uintVal.Type, uintVal.UintVal)
	}

	boolVal := StackValueBool(true)
	if boolVal.Type != StackTypeBool || !boolVal.BoolVal {
		t.Errorf("StackValueBool failed")
	}

	addrVal := StackValueAddress("addr1")
	if addrVal.Type != StackTypeAddress || addrVal.AddrVal != "addr1" {
		t.Errorf("StackValueAddress failed")
	}

	nullVal := StackValueNull()
	if nullVal.Type != StackTypeNull {
		t.Errorf("StackValueNull failed")
	}

	coinsVal := StackValueCoins(500)
	if coinsVal.Type != StackTypeCoins || coinsVal.UintVal != 500 {
		t.Errorf("StackValueCoins failed")
	}

	hashVal := StackValueHash([]byte{1, 2, 3})
	if hashVal.Type != StackTypeHash {
		t.Errorf("StackValueHash failed")
	}
}

func TestStackValueTypeStrings(t *testing.T) {
	expected := map[StackValueType]string{
		StackTypeInt256:	"int256",
		StackTypeBool:		"bool",
		StackTypeChunkRef:	"ChunkRef",
		StackTypeFrameRef:	"ExecutionFrameRef",
		StackTypeTuple:		"tuple",
		StackTypeAddress:	"address",
		StackTypeHash:		"hash",
		StackTypeCoins:		"coins",
		StackTypeString:	"string",
		StackTypeBytes:		"bytes",
		StackTypeNull:		"null",
	}
	for typ, str := range expected {
		if typ.String() != str {
			t.Errorf("expected %s, got %s", str, typ.String())
		}
	}
}

func TestPushPopStack(t *testing.T) {
	frame := &KernelExecutionFrame{
		Stack:		make([]StackValue, 0),
		PhaseGas:	make(map[Phase]uint64),
		ActionQueue:	NewActionQueueChunk(),
		GasLimit:	100000,
	}

	err := frame.PushValue(StackValueInt256(10))
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	err = frame.PushValue(StackValueBool(true))
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}

	if len(frame.Stack) != 2 {
		t.Fatalf("expected stack depth 2, got %d", len(frame.Stack))
	}

	val, err := frame.PopValue()
	if err != nil {
		t.Fatalf("pop failed: %v", err)
	}
	if val.Type != StackTypeBool {
		t.Errorf("expected bool on top, got %v", val.Type)
	}

	val, err = frame.PopValue()
	if err != nil {
		t.Fatalf("pop failed: %v", err)
	}
	if val.Type != StackTypeInt256 || val.IntVal != 10 {
		t.Errorf("expected int256(10), got %v(%d)", val.Type, val.IntVal)
	}

	if len(frame.Stack) != 0 {
		t.Errorf("expected empty stack, got %d items", len(frame.Stack))
	}
}

func TestPopValueOfTypeEnforcement(t *testing.T) {
	frame := &KernelExecutionFrame{
		Stack:		make([]StackValue, 0),
		PhaseGas:	make(map[Phase]uint64),
		ActionQueue:	NewActionQueueChunk(),
		GasLimit:	100000,
	}

	frame.PushValue(StackValueInt256(42))

	_, err := frame.PopValueOfType(StackTypeBool)
	if err == nil {
		t.Error("expected type mismatch error, got nil")
	}
	if !frame.Aborted {
		t.Error("frame should be aborted on type mismatch")
	}
	if frame.ErrorState.Category != ExitCategoryTypeError {
		t.Errorf("expected TYPE_ERROR, got %v", frame.ErrorState.Category)
	}
}

func TestStackOverflowProtection(t *testing.T) {
	frame := &KernelExecutionFrame{
		Stack:		make([]StackValue, 1023),
		PhaseGas:	make(map[Phase]uint64),
		ActionQueue:	NewActionQueueChunk(),
		GasLimit:	100000,
	}

	err := frame.PushValue(StackValueInt256(1))
	if err != nil {
		t.Error("should accept 1024th element")
	}

	err = frame.PushValue(StackValueInt256(2))
	if err == nil {
		t.Error("should reject 1025th element (overflow)")
	}
	if !frame.Aborted {
		t.Error("frame should be aborted on stack overflow")
	}
	if frame.ErrorState.Category != ExitCategoryVMError {
		t.Errorf("expected VM_ERROR, got %v", frame.ErrorState.Category)
	}
}

func TestStackUnderflowProtection(t *testing.T) {
	frame := &KernelExecutionFrame{
		Stack:		make([]StackValue, 0),
		PhaseGas:	make(map[Phase]uint64),
		ActionQueue:	NewActionQueueChunk(),
		GasLimit:	100000,
	}

	_, err := frame.PopValue()
	if err == nil {
		t.Error("expected stack underflow error")
	}
	if !frame.Aborted {
		t.Error("frame should be aborted on stack underflow")
	}
}

func TestChargeGas(t *testing.T) {
	frame := &KernelExecutionFrame{
		PhaseGas:	make(map[Phase]uint64),
		GasLimit:	1000,
		ActionQueue:	NewActionQueueChunk(),
	}

	frame.Phase = PhaseCompute

	if !frame.ChargeGas(100) {
		t.Error("charge 100 should succeed with limit 1000")
	}
	if frame.GasUsed != 100 {
		t.Errorf("expected GasUsed=100, got %d", frame.GasUsed)
	}
	if frame.PhaseGas[PhaseCompute] != 100 {
		t.Errorf("expected PhaseGas[Compute]=100, got %d", frame.PhaseGas[PhaseCompute])
	}

	if !frame.ChargeGas(900) {
		t.Error("charge 900 should succeed (100+900=1000)")
	}
	if frame.GasUsed != 1000 {
		t.Errorf("expected GasUsed=1000, got %d", frame.GasUsed)
	}

	if frame.ChargeGas(1) {
		t.Error("charge 1 should fail (already at limit)")
	}
	if !frame.Aborted {
		t.Error("frame should be aborted on gas exhaustion")
	}
}

func TestISAOpcodeTableCompleteness(t *testing.T) {
	required := []ISAOpcode{
		OpISANop, OpISAPush, OpISADup, OpISASwap, OpISADrop, OpISAOver,
		OpISAAdd, OpISASub, OpISAMul, OpISADiv, OpISAMod,
		OpISAEq, OpISANeq, OpISALt, OpISALte, OpISAGt, OpISAGte,
		OpISAAnd, OpISAOr, OpISANot,
		OpISACallFrame, OpISAReturnFrame, OpISAJumpCond, OpISAJumpUncond,
		OpISARaiseError, OpISATryBegin, OpISATryEnd,
		OpISALoadState, OpISAStoreState, OpISACloneState, OpISAMergeState,
		OpISAChunkMapGet, OpISAChunkMapPut, OpISAChunkMapDelete, OpISAChunkMapProof,
		OpISAEmitAction, OpISAQueueMessage, OpISAFlushActions,
		OpISAGetCaller, OpISAGetOrigin, OpISAGetValue, OpISAGetBlockHeight,
		OpISAHashChunk, OpISAHashData, OpISAVerifySig,
		OpISAEncode, OpISADecode,
	}
	for _, op := range required {
		if _, ok := ISAOpcodeTable[op]; !ok {
			t.Errorf("opcode 0x%04X missing from ISAOpcodeTable", op)
		}
	}
}

func TestISAStackEffect(t *testing.T) {
	tests := map[ISAOpcode]int{
		OpISANop:	0,
		OpISAPush:	1,
		OpISADup:	1,
		OpISADrop:	-1,
		OpISAAdd:	-1,
		OpISASwap:	0,
		OpISAEq:	-1,
	}
	for op, expected := range tests {
		effect, ok := ISAStackEffect(op)
		if !ok {
			t.Errorf("opcode 0x%04X not found", op)
			continue
		}
		if effect != expected {
			t.Errorf("opcode 0x%04X: expected effect %d, got %d", op, expected, effect)
		}
	}
}

func TestISAVerifyStackContract(t *testing.T) {
	if !ISAVerifyStackContract(OpISANop, 0) {
		t.Error("nop should require 0 stack items")
	}
	if !ISAVerifyStackContract(OpISAAdd, 2) {
		t.Error("add should pass with depth 2")
	}
	if ISAVerifyStackContract(OpISAAdd, 1) {
		t.Error("add should fail with depth 1 (needs 2)")
	}
	if !ISAVerifyStackContract(OpISADup, 1) {
		t.Error("dup should pass with depth 1")
	}
}

func TestISAGasCosts(t *testing.T) {
	tests := map[ISAOpcode]uint64{
		OpISANop:		1,
		OpISAPush:		2,
		OpISAAdd:		3,
		OpISADiv:		5,
		OpISACallFrame:		100,
		OpISAHashChunk:		90,
		OpISAChunkMapGet:	30,
	}
	for op, expectedMin := range tests {
		cost := ISAGasCost(op)
		if cost < expectedMin {
			t.Errorf("opcode 0x%04X: gas cost %d < min expected %d", op, cost, expectedMin)
		}
	}

	unknownCost := ISAGasCost(ISAOpcode(0xFFFF))
	if unknownCost != 1000 {
		t.Errorf("unknown opcode should cost 1000, got %d", unknownCost)
	}
}

func TestISAInstructionSpecFields(t *testing.T) {
	for op, spec := range ISAOpcodeTable {
		if spec.Mnemonic == "" {
			t.Errorf("opcode 0x%04X missing mnemonic", op)
		}
		if spec.Opcode != op {
			t.Errorf("opcode 0x%04X spec has wrong Opcode field 0x%04X", op, spec.Opcode)
		}
		if spec.BaseGasCost == 0 && op != OpISANop {
			t.Errorf("opcode 0x%04X (%s) has zero BaseGasCost", op, spec.Mnemonic)
		}
		if spec.DeterminismRule == "" {
			t.Errorf("opcode 0x%04X (%s) missing determinism rule", op, spec.Mnemonic)
		}
		if spec.StackInputs < 0 {
			t.Errorf("opcode 0x%04X has negative StackInputs", op)
		}
	}
}

func TestActionQueueEmitAndFinalize(t *testing.T) {
	q := NewActionQueueChunk()

	b := chunk.NewBuilder()
	b.SetTypeTag(chunk.TypeNormal)
	b.SetData([]byte("hello"), 40)
	payload, _ := b.Build()

	q.EmitMessage("target1", payload, 100, 0)
	q.EmitMessage("target2", nil, 200, 0)
	q.EmitEvent(payload)

	if len(q.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(q.Actions))
	}
	if q.Actions[0].Target != "target1" {
		t.Errorf("expected target1, got %s", q.Actions[0].Target)
	}
	if q.Actions[1].Target != "target2" {
		t.Errorf("expected target2, got %s", q.Actions[1].Target)
	}
	if q.Actions[2].Type != ActionEvent {
		t.Errorf("expected ActionEvent, got %v", q.Actions[2].Type)
	}

	hash := q.Finalize()
	if hash == nil {
		t.Error("Finalize should produce non-nil hash")
	}
	if q.Hash == nil {
		t.Error("Finalize should set q.Hash")
	}
}

func TestActionQueueDeterministicOrdering(t *testing.T) {
	q1 := NewActionQueueChunk()
	q1.EmitMessage("zzz", nil, 1, 0)
	q1.EmitMessage("aaa", nil, 2, 0)
	q1.EmitMessage("mmm", nil, 3, 0)
	q1.Finalize()

	q2 := NewActionQueueChunk()
	q2.EmitMessage("zzz", nil, 1, 0)
	q2.EmitMessage("aaa", nil, 2, 0)
	q2.EmitMessage("mmm", nil, 3, 0)
	q2.Finalize()

	if string(q1.Hash) != string(q2.Hash) {
		t.Error("same actions in same order should produce same hash (deterministic)")
	}
}

func TestActionQueueEmitActionWithType(t *testing.T) {
	q := NewActionQueueChunk()
	q.EmitAction(ActionInternal, "addr1", nil, 50)
	q.EmitAction(ActionExternal, "addr2", nil, 100)
	q.EmitAction(ActionSystem, "addr3", nil, 0)
	q.EmitAction(ActionEvent, "addr4", nil, 0)

	if len(q.Actions) != 4 {
		t.Fatalf("expected 4 actions, got %d", len(q.Actions))
	}
	if q.Actions[0].Type != ActionInternal {
		t.Error("expected ActionInternal")
	}
	if q.Actions[1].Type != ActionExternal {
		t.Error("expected ActionExternal")
	}
	if q.Actions[2].Type != ActionSystem {
		t.Error("expected ActionSystem")
	}
	if q.Actions[3].Type != ActionEvent {
		t.Error("expected ActionEvent")
	}
}

func TestEmptyStateRootChunk(t *testing.T) {
	root := EmptyStateRootChunk()
	if root == nil {
		t.Fatal("EmptyStateRootChunk should not return nil")
	}
	if root.Root == nil {
		t.Error("EmptyStateRootChunk should have non-nil Root")
	}
	hash := root.RootHash()
	if hash == nil {
		t.Error("RootHash should not be nil")
	}
}

func TestStateRootChunkFromMap(t *testing.T) {
	m := chunk.NewEmptyMap()
	b := chunk.NewBuilder()
	b.SetTypeTag(chunk.TypeNormal)
	b.SetData([]byte("value"), 40)
	val, _ := b.Build()
	m, _ = m.Put([]byte("key"), val)

	root := NewStateRootChunk(m.Root())
	if root.Root == nil {
		t.Error("StateRootChunk should have non-nil Root")
	}
	hash1 := root.RootHash()

	m2, _ := m.Put([]byte("key2"), val)
	root2 := NewStateRootChunk(m2.Root())
	hash2 := root2.RootHash()

	if string(hash1) == string(hash2) {
		t.Error("different state roots should have different hashes")
	}
}

func TestStateRootImmutability(t *testing.T) {
	m := chunk.NewEmptyMap()
	b := chunk.NewBuilder()
	b.SetTypeTag(chunk.TypeNormal)
	b.SetData([]byte("initial"), 56)
	val, _ := b.Build()
	m, _ = m.Put([]byte("key1"), val)

	root1 := NewStateRootChunk(m.Root())
	hash1 := root1.RootHash()

	m2, _ := m.Put([]byte("key2"), val)
	root2 := NewStateRootChunk(m2.Root())
	hash2 := root2.RootHash()

	if string(hash1) == string(hash2) {
		t.Error("state mutation should produce different root hash")
	}
}

func TestExecutionContextChunkToChunk(t *testing.T) {
	ctx := ExecutionContextChunk{
		Caller:			"caller1",
		Origin:			"origin1",
		AttachedValue:		1000,
		BlockHeight:		42,
		ChainID:		"testnet-1",
		ContractAddress:	"contract1",
		MessageHash:		[]byte{1, 2, 3, 4},
		Timestamp:		1234567890,
	}

	chk, err := ctx.ToChunk()
	if err != nil {
		t.Fatalf("ToChunk failed: %v", err)
	}
	if chk == nil {
		t.Fatal("chunk should not be nil")
	}
	if chk.Hash() == nil {
		t.Error("chunk hash should not be nil")
	}
}

func TestExecutionContextFromBlockContext(t *testing.T) {
	blockCtx := BlockContext{
		Height:		100,
		Timestamp:	1234567890,
		ChainID:	"mainnet",
	}
	msg := Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		500,
		GasLimit:	10000,
	}

	ctx := ExecutionContextFromBlockContext(blockCtx, msg)
	if ctx.Caller != "sender1" {
		t.Errorf("expected Caller=sender1, got %s", ctx.Caller)
	}
	if ctx.ContractAddress != "target1" {
		t.Errorf("expected ContractAddress=target1, got %s", ctx.ContractAddress)
	}
	if ctx.BlockHeight != 100 {
		t.Errorf("expected BlockHeight=100, got %d", ctx.BlockHeight)
	}
	if ctx.ChainID != "mainnet" {
		t.Errorf("expected ChainID=mainnet, got %s", ctx.ChainID)
	}
}

func TestExecuteKernelSemanticsFivePhases(t *testing.T) {
	emptyMap := chunk.NewEmptyMap()
	state := emptyMap.Root()

	msg := Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		100,
		GasLimit:	100000,
	}

	frame := NewKernelExecutionFrame(state, msg, 100)
	newRoot, actionQueue, exitCode, receipt, err := ExecuteKernelSemantics(frame)
	if err != nil {
		t.Fatalf("ExecuteKernelSemantics failed: %v", err)
	}
	if newRoot == nil {
		t.Error("newRoot should not be nil")
	}
	if actionQueue == nil {
		t.Error("actionQueue should not be nil")
	}
	if exitCode.Category != ExitCategorySuccess {
		t.Errorf("expected SUCCESS, got %v", exitCode.Category)
	}
	if receipt.GasUsed == 0 {
		t.Error("gas should be charged")
	}

	trackPhases := []Phase{PhaseStorage, PhaseCredit, PhaseCompute, PhaseAction, PhaseFinalization}
	for _, p := range trackPhases {
		if _, ok := frame.PhaseGas[p]; !ok {
			t.Errorf("phase %v should have gas charges", p)
		}
	}
}

func TestExecuteKernelSemanticsGasExhaustion(t *testing.T) {
	emptyMap := chunk.NewEmptyMap()
	state := emptyMap.Root()

	msg := Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		100,
		GasLimit:	100,
	}

	frame := NewKernelExecutionFrame(state, msg, 10)
	_, _, exitCode, _, err := ExecuteKernelSemantics(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode.Category != ExitCategoryGasError {
		t.Errorf("expected GAS_ERROR exit, got %v", exitCode.Category)
	}
}

func TestExecuteKernelSemanticsActionBudgetExceeded(t *testing.T) {
	emptyMap := chunk.NewEmptyMap()
	state := emptyMap.Root()

	msg := Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		100,
		GasLimit:	100000,
	}

	frame := NewKernelExecutionFrame(state, msg, 2)
	for i := 0; i < 5; i++ {
		frame.ActionQueue.EmitMessage("target", nil, 1, 0)
	}

	_, _, exitCode, _, err := ExecuteKernelSemantics(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode.Category != ExitCategoryActionError {
		t.Errorf("expected ACTION_ERROR for budget exceeded, got %v", exitCode.Category)
	}
}

func TestKernelExecutionTraceDeterminism(t *testing.T) {
	emptyMap := chunk.NewEmptyMap()
	state := emptyMap.Root()

	msg := Message{
		Sender:		"sender1",
		Target:		"target1",
		Value:		100,
		GasLimit:	100000,
	}

	frame1 := NewKernelExecutionFrame(state, msg, 100)
	frame1.PushValue(StackValueInt256(1))
	frame1.PushValue(StackValueInt256(2))
	frame1.PushValue(StackValueBool(true))

	frame2 := NewKernelExecutionFrame(state, msg, 100)
	frame2.PushValue(StackValueInt256(1))
	frame2.PushValue(StackValueInt256(2))
	frame2.PushValue(StackValueBool(true))

	trace1 := frame1.Trace.Steps
	trace2 := frame2.Trace.Steps
	if len(trace1) != len(trace2) {
		t.Fatalf("trace length mismatch: %d vs %d", len(trace1), len(trace2))
	}
	for i := range trace1 {
		if trace1[i].Opcode != trace2[i].Opcode {
			t.Errorf("step %d: opcode mismatch %v vs %v", i, trace1[i].Opcode, trace2[i].Opcode)
		}
		if trace1[i].StackDepth != trace2[i].StackDepth {
			t.Errorf("step %d: depth mismatch %d vs %d", i, trace1[i].StackDepth, trace2[i].StackDepth)
		}
	}
}

func TestNewKernelExecutionFrameDefaults(t *testing.T) {
	emptyMap := chunk.NewEmptyMap()
	state := emptyMap.Root()
	msg := Message{GasLimit: 50000}

	frame := NewKernelExecutionFrame(state, msg, 50)

	if frame.IP != 0 {
		t.Error("IP should start at 0")
	}
	if len(frame.Stack) != 0 {
		t.Error("stack should start empty")
	}
	if frame.Phase != PhaseStorage {
		t.Error("phase should start at Storage")
	}
	if frame.GasLimit != 50000 {
		t.Error("gas limit should match message")
	}
	if frame.GasUsed != 0 {
		t.Error("gas used should start at 0")
	}
	if frame.ActionBudget != 50 {
		t.Error("action budget should match parameter")
	}
	if frame.Aborted {
		t.Error("frame should not start aborted")
	}
	if frame.ActionQueue == nil {
		t.Error("action queue should be initialized")
	}
	if frame.Trace.Steps == nil {
		t.Error("trace steps should be initialized")
	}
}
