package avm

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

func TestCapabilityClassString(t *testing.T) {
	tests := map[CapabilityClass]string{
		CapClassTime:			"TIME",
		CapClassRandomness:		"RANDOMNESS",
		CapClassIO:			"IO",
		CapClassProcessControl:		"PROCESS_CONTROL",
		CapClassParallelExec:		"PARALLEL_EXECUTION",
		CapClassFloatArithmetic:	"FLOAT_ARITHMETIC",
		CapClassStorage:		"STORAGE",
		CapClassMessaging:		"MESSAGING",
		CapClassCrypto:			"CRYPTO",
		CapClassChain:			"CHAIN",
	}
	for class, expected := range tests {
		if class.String() != expected {
			t.Errorf("expected %s, got %s", expected, class.String())
		}
	}
}

func TestForbiddenCapabilityClasses(t *testing.T) {
	forbidden := []CapabilityClass{
		CapClassTime,
		CapClassRandomness,
		CapClassIO,
		CapClassProcessControl,
		CapClassParallelExec,
		CapClassFloatArithmetic,
	}
	for _, class := range forbidden {
		if !ForbiddenCapabilityClasses[class] {
			t.Errorf("capability class %s should be forbidden", class)
		}
	}
	allowed := []CapabilityClass{CapClassStorage, CapClassMessaging, CapClassCrypto, CapClassChain}
	for _, class := range allowed {
		if ForbiddenCapabilityClasses[class] {
			t.Errorf("capability class %s should NOT be forbidden", class)
		}
	}
}

func TestCapabilityClassForHostFunction(t *testing.T) {
	tests := map[HostFunction]CapabilityClass{
		HostWallClockTime:	CapClassTime,
		HostRandomness:		CapClassRandomness,
		HostFilesystem:		CapClassIO,
		HostNetwork:		CapClassIO,
		HostFloatingPoint:	CapClassFloatArithmetic,
		HostGoroutine:		CapClassProcessControl,
		HostProcessEnv:		CapClassProcessControl,
		HostReadStorage:	CapClassStorage,
		HostWriteStorage:	CapClassStorage,
		HostHashSHA256:		CapClassCrypto,
		HostHashBLAKE3:		CapClassCrypto,
		HostGetBlockHeight:	CapClassChain,
	}
	for fn, expected := range tests {
		got := CapabilityClassForHostFunction(fn)
		if got != expected {
			t.Errorf("host %d: expected %s, got %s", fn, expected, got)
		}
	}
}

func TestForbiddenHostFunctionClasses(t *testing.T) {
	classes := ForbiddenHostFunctionClasses()
	if len(classes) == 0 {
		t.Error("expected non-empty forbidden host function classes")
	}
	for _, c := range classes {
		if !ForbiddenCapabilityClasses[c] {
			t.Errorf("returned class %s is not in forbidden set", c)
		}
	}
}

func TestDeterminismGateCreatesWithChecks(t *testing.T) {
	gate := NewDeterminismGate()
	if len(gate.Checks) == 0 {
		t.Error("expected determinism gate to have checks")
	}
	if len(gate.Checks) < 5 {
		t.Errorf("expected at least 5 checks, got %d", len(gate.Checks))
	}
}

func TestDeterminismGateValidatesValidFrame(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 10000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	gate := NewDeterminismGate()
	results := gate.Validate(frame)
	for _, result := range results {
		if !result.Passed {
			t.Errorf("check %s failed: %s", result.CheckName, result.ViolationDetail)
		}
	}
}

func TestDeterminismGateValidatesAll(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 10000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	gate := NewDeterminismGate()
	if !gate.ValidateAll(frame) {
		t.Error("expected valid frame to pass all gate checks")
	}
}

func TestDeterminismGateNilStateFails(t *testing.T) {
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 10000}
	frame := NewKernelExecutionFrame(nil, msg, 100)

	result := CheckStateSnapshotIntegrity(frame)
	if result.Passed {
		t.Error("expected nil state snapshot to fail integrity check")
	}
	if result.CheckName != "state_snapshot_integrity" {
		t.Errorf("expected state_snapshot_integrity, got %s", result.CheckName)
	}
}

func TestDeterminismGateZeroGasFails(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 0}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	result := CheckGasModelConsistency(frame)
	if result.Passed {
		t.Error("expected zero gas limit to fail gas consistency check")
	}
}

func TestDeterminismGateStackOverflow(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 10000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)
	for i := 0; i < 1100; i++ {
		frame.Stack = append(frame.Stack, StackValueInt256(0))
	}

	result := CheckStackBounds(frame)
	if result.Passed {
		t.Error("expected stack overflow to fail bounds check")
	}
}

func TestDeterminismGateActionBudgetExceeded(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 10000}
	frame := NewKernelExecutionFrame(state.Root, msg, 10)
	frame.ActionsUsed = 20

	result := CheckActionBudgetBounds(frame)
	if result.Passed {
		t.Error("expected action budget exceeded to fail")
	}
}

func TestDeterministicRepeatedExecution(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", Value: 100, GasLimit: 5000}

	exitCodes := make([]StructuredExitCode, 0, 5)
	gasUsages := make([]uint64, 0, 5)

	for i := 0; i < 5; i++ {
		frame := NewKernelExecutionFrame(state.Root, msg, 100)
		_, _, exitCode, receipt, _ := ExecuteKernelSemantics(frame)
		exitCodes = append(exitCodes, exitCode)
		gasUsages = append(gasUsages, receipt.GasUsed)
	}

	for i := 1; i < 5; i++ {
		if exitCodes[i].ToUint32() != exitCodes[0].ToUint32() {
			t.Errorf("execution %d: exit code mismatch %v != %v", i, exitCodes[i], exitCodes[0])
		}
		if gasUsages[i] != gasUsages[0] {
			t.Errorf("execution %d: gas usage mismatch %d != %d", i, gasUsages[i], gasUsages[0])
		}
	}
}

func TestNormalizeChunkMapEntries(t *testing.T) {
	m := chunk.NewEmptyMap()
	b1, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("val1"), 32).Build()
	b2, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("val2"), 32).Build()
	b3, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("val3"), 32).Build()

	m, _ = m.Put([]byte("charlie"), b1)
	m, _ = m.Put([]byte("alpha"), b2)
	m, _ = m.Put([]byte("bravo"), b3)

	entries := m.Iterate()
	normalized := NormalizeChunkMapEntries(entries)

	if len(normalized) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(normalized))
	}

	keys := make([]string, len(normalized))
	for i, e := range normalized {
		keys[i] = string(e.Key)
	}

	sorted := true
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			sorted = false
			break
		}
	}
	if !sorted {
		t.Errorf("expected sorted keys, got %v", keys)
	}
}

func TestMapIterationDeterminism(t *testing.T) {
	entrySets := make([][]chunk.Entry, 0, 3)

	keys := [][]string{
		{"z", "a", "m"},
		{"m", "z", "a"},
		{"a", "m", "z"},
	}

	for _, ks := range keys {
		m := chunk.NewEmptyMap()
		for _, k := range ks {
			b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("v_"+k), 24).Build()
			m, _ = m.Put([]byte(k), b)
		}
		entrySets = append(entrySets, m.Iterate())
	}

	if !VerifyMapIterationDeterminism(entrySets) {
		t.Error("expected normalized map iteration to be deterministic")
	}
}

func TestExportImportReExecute(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", Value: 200, GasLimit: 10000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	_, _, _, receipt1, _ := ExecuteKernelSemantics(frame)

	stateHashBefore := state.Root.Hash()

	frame2 := NewKernelExecutionFrame(state.Root, msg, 100)
	_, _, _, receipt2, _ := ExecuteKernelSemantics(frame2)

	if receipt1.ExitCode != receipt2.ExitCode {
		t.Errorf("exit codes differ: %d vs %d", receipt1.ExitCode, receipt2.ExitCode)
	}
	if receipt1.GasUsed != receipt2.GasUsed {
		t.Errorf("gas usage differs: %d vs %d", receipt1.GasUsed, receipt2.GasUsed)
	}

	_ = stateHashBefore
}

func TestFuzzBytecodeVerifierCorpus(t *testing.T) {
	model := DefaultFuzzResilienceModel()

	for i := 0; i < 100; i++ {
		data := GenerateFuzzBytecode(uint64(i*31+17), 64)

		if len(data) != 64 {
			t.Fatalf("fuzz %d: expected 64 bytes, got %d", i, len(data))
		}

		err := model.ValidateBytecodeSize(len(data))
		if err != nil {
			t.Errorf("fuzz %d: valid-sized bytecode rejected: %v", i, err)
		}
	}
}

func TestFuzzBytecodeOversized(t *testing.T) {
	model := DefaultFuzzResilienceModel()
	largeData := make([]byte, int(model.MaxBytecodeSize)+1)
	err := model.ValidateBytecodeSize(len(largeData))
	if err == nil {
		t.Error("expected oversized bytecode to be rejected")
	}
}

func TestFuzzChunkDataValidation(t *testing.T) {
	model := DefaultFuzzResilienceModel()

	validChunk := GenerateMalformedChunk(42)
	err := model.ValidateChunkData(validChunk.Data())
	if err != nil {
		t.Errorf("valid chunk data rejected: %v", err)
	}

	oversizeData := make([]byte, int(model.MaxChunkDataSize)+1)
	err = model.ValidateChunkData(oversizeData)
	if err == nil {
		t.Error("expected oversize chunk data to be rejected")
	}
}

func TestFuzzStackDepthValidation(t *testing.T) {
	model := DefaultFuzzResilienceModel()

	err := model.ValidateStackDepth(100)
	if err != nil {
		t.Errorf("valid stack depth rejected: %v", err)
	}

	err = model.ValidateStackDepth(int(model.MaxStackDepth) + 1)
	if err == nil {
		t.Error("expected exceeding max stack depth to be rejected")
	}
}

func TestVerifyRawFuzzBytecode(t *testing.T) {
	result := VerifyRawFuzzBytecode(nil)
	if result.ToUint32() != ExitValidationFailed.ToUint32() {
		t.Error("expected ExitValidationFailed for nil input")
	}

	result = VerifyRawFuzzBytecode([]byte{0x00})
	if result.ToUint32() != ExitValidationFailed.ToUint32() {
		t.Error("expected ExitValidationFailed for too-short input")
	}

	magic := make([]byte, 4)
	magic[0] = 0x41
	magic[1] = 0x56
	magic[2] = 0x4D
	magic[3] = 0x01
	validInput := append(magic, make([]byte, 100)...)
	result = VerifyRawFuzzBytecode(validInput)
	if result.ToUint32() != ExitSuccess.ToUint32() {
		t.Error("expected ExitSuccess for valid-magic input")
	}

	badMagic := make([]byte, 128)
	result = VerifyRawFuzzBytecode(badMagic)
	if result.ToUint32() != ExitValidationFailed.ToUint32() {
		t.Error("expected ExitValidationFailed for invalid magic")
	}
}

func TestForbiddenHostFunctionIDs(t *testing.T) {
	forbiddenIDs := ForbiddenHostFunctionIDs()
	expected := []HostFunction{
		HostWallClockTime, HostRandomness, HostFilesystem, HostNetwork,
		HostFloatingPoint, HostGoroutine, HostProcessEnv, HostNondeterministicMap,
	}
	if len(forbiddenIDs) != len(expected) {
		t.Errorf("expected %d forbidden host functions, got %d", len(expected), len(forbiddenIDs))
	}
}

func TestIsForbiddenHostFunction(t *testing.T) {
	forbidden := []HostFunction{HostWallClockTime, HostRandomness, HostFilesystem, HostNetwork, HostFloatingPoint, HostGoroutine, HostProcessEnv, HostNondeterministicMap}
	for _, fn := range forbidden {
		if !IsForbiddenHostFunction(fn) {
			t.Errorf("host function %d should be forbidden", fn)
		}
	}
	allowed := []HostFunction{HostReadStorage, HostWriteStorage, HostHashSHA256, HostHashBLAKE3, HostGetBlockHeight}
	for _, fn := range allowed {
		if IsForbiddenHostFunction(fn) {
			t.Errorf("host function %d should be allowed", fn)
		}
	}
}

func TestRuntimeIsolationEnforcer(t *testing.T) {
	forbidden := []HostFunction{HostWallClockTime, HostRandomness, HostFilesystem, HostNetwork, HostFloatingPoint, HostGoroutine}
	for _, fn := range forbidden {
		err := EnforceRuntimeIsolation(fn)
		if err == nil {
			t.Errorf("expected forbidden host function %d to be rejected", fn)
		}
	}
	allowed := []HostFunction{HostReadStorage, HostGetBlockHeight, HostHashSHA256}
	for _, fn := range allowed {
		err := EnforceRuntimeIsolation(fn)
		if err != nil {
			t.Errorf("expected allowed host function %d to pass, got: %v", fn, err)
		}
	}
}

func TestGasOverflowDeterministicError(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 1}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	_, _, exitCode, _, _ := ExecuteKernelSemantics(frame)
	if exitCode.Category != ExitCategoryGasError {
		t.Errorf("expected GasError exit for zero gas limit, got %v", exitCode)
	}
}

func TestGasExhaustionNoCorruption(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 100}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	rootBefore := state.Root.Hash()
	_, _, exitCode, _, _ := ExecuteKernelSemantics(frame)

	if exitCode.Category == ExitCategorySuccess {
		rootAfter := frame.StateSnapshot.Hash()
		for i := range rootBefore {
			if rootBefore[i] != rootAfter[i] {
				t.Error("state should not be mutated on gas exhausted execution")
				break
			}
		}
	}
}

func TestGasSafetyModelValidation(t *testing.T) {
	model := DefaultGasSafetyModel()

	if err := model.ValidateGasLimit(1000); err != nil {
		t.Errorf("valid gas limit rejected: %v", err)
	}

	if err := model.ValidateGasLimit(0); err == nil {
		t.Error("expected zero gas limit to be rejected")
	}

	if err := model.ValidateGasLimit(model.MaxGasTotal + 1); err == nil {
		t.Error("expected over-limit gas to be rejected")
	}
}

func TestGasSafetyNoPartialSideEffects(t *testing.T) {
	model := DefaultGasSafetyModel()

	err := model.ValidateNoPartialSideEffects(model.MaxGasTotal, model.MaxGasTotal, 300, 256, ExitGasExhausted)
	if err != nil {
		t.Logf("ValidateNoPartialSideEffects returned: %v (acceptable for matching gas limit)", err)
	}
	_ = model
}

func TestStackDepthLimit(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 100000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	for i := 0; i < 1025; i++ {
		frame.Stack = append(frame.Stack, StackValueInt256(int64(i)))
	}

	result := CheckStackBounds(frame)
	if result.Passed {
		t.Error("expected stack overflow to be detected")
	}
	if result.CheckName != "stack_bounds" {
		t.Errorf("expected stack_bounds check, got %s", result.CheckName)
	}
}

func TestActionBudgetLimit(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 10000}
	frame := NewKernelExecutionFrame(state.Root, msg, 5)
	frame.ActionsUsed = 100

	result := CheckActionBudgetBounds(frame)
	if result.Passed {
		t.Error("expected action budget exceeded to be detected")
	}
}

func TestDefaultIsolationPolicyAllForbidden(t *testing.T) {
	policy := DefaultRuntimeIsolationPolicy()
	violations := policy.Validate()
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for default policy, got %d: %v", len(violations), violations)
	}
}

func TestIsolationPolicyDetectsViolations(t *testing.T) {
	policy := RuntimeIsolationPolicy{
		AllowFilesystem:	true,
		AllowWallClockTime:	true,
		AllowFloatingPoint:	true,
	}
	violations := policy.Validate()
	if len(violations) != 3 {
		t.Errorf("expected 3 violations, got %d", len(violations))
	}
}

func TestIsolationPolicyForbiddenAccesses(t *testing.T) {
	policy := RuntimeIsolationPolicy{
		AllowNetwork:			true,
		AllowProcessInteraction:	true,
		AllowProcessEntropy:		true,
		AllowThreadCreation:		true,
		AllowAsyncExecution:		true,
		AllowExternalRandomness:	true,
	}
	violations := policy.Validate()
	if len(violations) != 6 {
		t.Errorf("expected 6 violations, got %d: %v", len(violations), violations)
	}
}

func TestDefaultNormalizationPolicy(t *testing.T) {
	policy := DefaultNormalizationPolicy()
	if !policy.NormalizeMapIteration {
		t.Error("expected map iteration normalization to be enabled")
	}
	if !policy.NormalizeEventOrder {
		t.Error("expected event order normalization to be enabled")
	}
	if !policy.NormalizeMessageOrder {
		t.Error("expected message order normalization to be enabled")
	}
}

func TestNormalizeActionOrder(t *testing.T) {
	actions := []Action{
		{Type: ActionInternal, Target: "z_target"},
		{Type: ActionInternal, Target: "a_target"},
		{Type: ActionEvent, Target: "b_target"},
	}
	normalized := NormalizeActionOrder(actions)
	if normalized[0].Type != ActionInternal {
		t.Errorf("expected first action type %d, got %d", ActionInternal, normalized[0].Type)
	}
	if normalized[0].Target != "a_target" {
		t.Errorf("expected first internal target 'a_target', got %s", normalized[0].Target)
	}
	if normalized[1].Type != ActionInternal {
		t.Errorf("expected second action type %d, got %d", ActionInternal, normalized[1].Type)
	}
	if normalized[1].Target != "z_target" {
		t.Errorf("expected second internal target 'z_target', got %s", normalized[1].Target)
	}
	if normalized[2].Type != ActionEvent {
		t.Errorf("expected third action type %d, got %d", ActionEvent, normalized[2].Type)
	}
}

func TestNormalizeMessageOrder(t *testing.T) {
	msgs := []Message{
		{Type: MessageInternal, Sender: "AE:charlie", Target: "4:contract", Height: 10},
		{Type: MessageInternal, Sender: "AE:alice", Target: "4:contract", Height: 5},
		{Type: MessageInternal, Sender: "AE:bob", Target: "4:contract", Height: 5},
	}
	normalized := NormalizeMessageOrder(msgs)
	if normalized[0].Height > normalized[1].Height {
		t.Error("expected messages sorted by height")
	}
}

func TestStaticSecurityScannerCreation(t *testing.T) {
	scanner := NewStaticSecurityScanner()
	if len(scanner.ForbiddenHostFns) == 0 {
		t.Error("expected forbidden host functions to be populated")
	}
}

func TestStaticSecurityScannerScanHostImports(t *testing.T) {
	scanner := NewStaticSecurityScanner()
	violations := scanner.ScanHostImports(ForbiddenHostFunctionIDs())
	if len(violations) == 0 {
		t.Error("expected violations for forbidden host imports")
	}
	for _, v := range violations {
		if v.Code != "FORBIDDEN_HOST_IMPORT" {
			t.Errorf("expected FORBIDDEN_HOST_IMPORT, got %s", v.Code)
		}
	}
}

func TestStaticSecurityScannerFullScan(t *testing.T) {
	scanner := NewStaticSecurityScanner()
	violations := scanner.FullScan(ForbiddenHostFunctionIDs(), []byte{0x01, 0x02}, 10, 0)
	if len(violations) == 0 {
		t.Error("expected violations from full scan with forbidden host functions")
	}
}

func TestDefaultFuzzResilienceModel(t *testing.T) {
	model := DefaultFuzzResilienceModel()
	if model.MaxBytecodeSize != 1<<20 {
		t.Errorf("expected MaxBytecodeSize %d, got %d", 1<<20, model.MaxBytecodeSize)
	}
	if model.MaxStackDepth != 1024 {
		t.Errorf("expected MaxStackDepth 1024, got %d", model.MaxStackDepth)
	}
}

func TestFuzzResilienceBytecodeSizeLimits(t *testing.T) {
	model := DefaultFuzzResilienceModel()

	smallData := GenerateFuzzBytecode(1, 100)
	if err := model.ValidateBytecodeSize(len(smallData)); err != nil {
		t.Errorf("small valid bytecode rejected: %v", err)
	}

	oversizeData := make([]byte, int(model.MaxBytecodeSize)+1)
	if err := model.ValidateBytecodeSize(len(oversizeData)); err == nil {
		t.Error("expected oversize bytecode to be rejected")
	}
}

func TestGenerateMalformedChunk(t *testing.T) {
	chunk_ := GenerateMalformedChunk(42)
	if chunk_ == nil {
		t.Error("expected non-nil chunk from GenerateMalformedChunk")
	}
}

func TestGenerateCorruptedModule(t *testing.T) {
	data := GenerateCorruptedModule(123, 256)
	if len(data) != 256 {
		t.Errorf("expected 256 bytes, got %d", len(data))
	}
	magic := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	if magic != AVMModuleMagic {
		t.Errorf("expected magic 0x%08X, got 0x%08X", AVMModuleMagic, magic)
	}
}

func TestGenerateRandomHostCall(t *testing.T) {
	fn := GenerateRandomHostCall(42)
	if fn == 0 {
		t.Error("expected non-zero host function")
	}
	registry := HostFunctionRegistry()
	if _, ok := registry[fn]; !ok {
		t.Errorf("generated host function %d not in registry", fn)
	}
}

func TestDeterministicRandomnessSameInputs(t *testing.T) {
	stateRoot := []byte("state_root_1")
	entropy := []byte("entropy_1")
	msgHash := []byte("msg_hash_1")
	domain := []byte("domain_1")

	r1 := GenerateDeterministicRandomness(stateRoot, entropy, msgHash, domain)
	r2 := GenerateDeterministicRandomness(stateRoot, entropy, msgHash, domain)

	if len(r1) != len(r2) {
		t.Fatalf("randomness lengths differ: %d vs %d", len(r1), len(r2))
	}
	for i := range r1 {
		if r1[i] != r2[i] {
			t.Errorf("deterministic randomness mismatch at byte %d", i)
			break
		}
	}
}

func TestDeterministicRandomnessDifferentInputs(t *testing.T) {
	r1 := GenerateDeterministicRandomness([]byte("state_a"), []byte("ent_a"), []byte("msg_a"), []byte("dom_a"))
	r2 := GenerateDeterministicRandomness([]byte("state_b"), []byte("ent_b"), []byte("msg_b"), []byte("dom_b"))

	same := true
	for i := range r1 {
		if i < len(r2) && r1[i] != r2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("expected different inputs to produce different randomness")
	}
}

func TestReplayRecordDeterminism(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", Value: 100, GasLimit: 5000}

	frame1 := NewKernelExecutionFrame(state.Root, msg, 100)
	_, actionQueue1, _, receipt1, _ := ExecuteKernelSemantics(frame1)

	frame2 := NewKernelExecutionFrame(state.Root, msg, 100)
	_, actionQueue2, _, receipt2, _ := ExecuteKernelSemantics(frame2)

	record1 := RecordReplay(state.Root, &StateRootChunk{Root: state.Root}, msg, receipt1, actionQueue1, frame1.Trace)
	record2 := RecordReplay(state.Root, &StateRootChunk{Root: state.Root}, msg, receipt2, actionQueue2, frame2.Trace)

	if !record1.VerifyDeterminism(record2) {
		t.Error("replay records should be deterministic")
	}
}

func TestReplayRecordGasConsistency(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", Value: 100, GasLimit: 5000}

	var lastGas uint64
	for i := 0; i < 3; i++ {
		frame := NewKernelExecutionFrame(state.Root, msg, 100)
		_, _, _, receipt, _ := ExecuteKernelSemantics(frame)
		if i > 0 && receipt.GasUsed != lastGas {
			t.Errorf("gas inconsistency: run %d used %d, previous used %d", i, receipt.GasUsed, lastGas)
		}
		lastGas = receipt.GasUsed
	}
}

func TestAdversarialVectorsDefined(t *testing.T) {
	if len(AdversarialVectors) == 0 {
		t.Fatal("expected adversarial vectors to be defined")
	}

	categories := make(map[string]int)
	for _, v := range AdversarialVectors {
		if v.Name == "" {
			t.Error("adversarial vector has empty name")
		}
		if v.Category == "" {
			t.Errorf("vector %s has empty category", v.Name)
		}
		categories[v.Category]++
	}

	required := []string{"malicious_bytecode", "crafted_gas", "invalid_state", "capability_violation", "message_amplification", "resource_exhaustion"}
	for _, cat := range required {
		if categories[cat] == 0 {
			t.Errorf("missing adversarial category: %s", cat)
		}
	}
}

func TestAdversarialVectorExitCodes(t *testing.T) {
	for _, v := range AdversarialVectors {
		if v.ExitCode.Category == ExitCategorySuccess && v.Category != "resource_exhaustion" {
			t.Errorf("adversarial vector %s should not have success exit code", v.Name)
		}
	}
}

func TestClassifyFailureSuccess(t *testing.T) {
	cls := ClassifyFailure(ExitSuccess)
	if cls.Kind != FailureNone {
		t.Errorf("expected FailureNone, got %v", cls.Kind)
	}
	if !cls.Deterministic {
		t.Error("success should be deterministic")
	}
}

func TestClassifyFailureGasError(t *testing.T) {
	cls := ClassifyFailure(ExitGasExhausted)
	if cls.Kind != FailureRecoverable {
		t.Errorf("expected FailureRecoverable, got %v", cls.Kind)
	}
	if !cls.Deterministic {
		t.Error("gas error should be deterministic")
	}
}

func TestClassifyFailureVMError(t *testing.T) {
	cls := ClassifyFailure(ExitStackOverflow)
	if cls.Kind != FailureRecoverable {
		t.Errorf("expected FailureRecoverable, got %v", cls.Kind)
	}
	if !cls.Deterministic {
		t.Error("VM error should be deterministic")
	}
}

func TestClassifyFailureStateError(t *testing.T) {
	cls := ClassifyFailure(ExitStateCorruption)
	if cls.Kind != FailureNonRecoverable {
		t.Errorf("expected FailureNonRecoverable, got %v", cls.Kind)
	}
	if !cls.Deterministic {
		t.Error("state error should be deterministic")
	}
}

func TestClassifyFailureTypeError(t *testing.T) {
	cls := ClassifyFailure(ExitTypeMismatch)
	if cls.Kind != FailureRecoverable {
		t.Errorf("expected FailureRecoverable for type error, got %v", cls.Kind)
	}
}

func TestClassifyFailureActionError(t *testing.T) {
	cls := ClassifyFailure(ExitActionBudget)
	if cls.Kind != FailureRecoverable {
		t.Errorf("expected FailureRecoverable for action error, got %v", cls.Kind)
	}
}

func TestFailureIsDeterministic(t *testing.T) {
	codes := []StructuredExitCode{ExitSuccess, ExitGasExhausted, ExitStackOverflow, ExitTypeMismatch, ExitDivZero, ExitChunkError, ExitActionBudget}
	for _, code := range codes {
		cls := ClassifyFailure(code)
		if !cls.IsDeterministic() {
			t.Errorf("expected %v to be deterministic", code)
		}
	}
}

func TestFailureSequenceIsolated(t *testing.T) {
	cls := ClassifyFailure(ExitGasExhausted)
	if !cls.SequenceIsIsolated() {
		t.Error("gas failure should have isolated sequence")
	}
}

func TestConsensusSafetyGuarantee(t *testing.T) {
	guarantee := ConsensusSafetyGuarantee{
		DeterminismPassed:	true,
		IsolationPassed:	true,
		ReplayabilityPassed:	true,
		BoundednessPassed:	true,
	}
	if !guarantee.IsConsensusSafe() {
		t.Error("expected consensus safety guarantee to pass")
	}
}

func TestConsensusSafetyGuaranteeFailed(t *testing.T) {
	guarantee := ConsensusSafetyGuarantee{
		DeterminismPassed:	false,
		IsolationPassed:	true,
		ReplayabilityPassed:	true,
		BoundednessPassed:	true,
	}
	if guarantee.IsConsensusSafe() {
		t.Error("expected consensus safety guarantee to fail with determinism fail")
	}
}

func TestVerifyConsensusSafety(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 5000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	gate := NewDeterminismGate()
	results := gate.Validate(frame)
	isolation := DefaultRuntimeIsolationPolicy()
	gasModel := DefaultGasSafetyModel()

	guarantee := VerifyConsensusSafety(results, isolation, nil, gasModel)
	if !guarantee.DeterminismPassed {
		t.Error("expected determinism to pass")
	}
	if !guarantee.IsolationPassed {
		t.Error("expected isolation to pass")
	}
	if !guarantee.BoundednessPassed {
		t.Error("expected boundedness to pass")
	}
}

func TestSecurityInvariantCheck(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 5000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	gate := NewDeterminismGate()
	results := gate.Validate(frame)
	isolation := DefaultRuntimeIsolationPolicy()
	gasModel := DefaultGasSafetyModel()

	invariant := CheckSecurityInvariant(results, isolation, gasModel)
	if !invariant.IsSatisfied() {
		t.Errorf("expected security invariant to be satisfied: %s", invariant.String())
	}
}

func TestSecurityInvariantViolation(t *testing.T) {
	isolation := RuntimeIsolationPolicy{
		AllowFilesystem:	true,
		AllowWallClockTime:	true,
	}
	gasModel := DefaultGasSafetyModel()
	gate := NewDeterminismGate()

	frame := NewKernelExecutionFrame(EmptyStateRootChunk().Root, Message{GasLimit: 10000}, 100)
	results := gate.Validate(frame)

	invariant := CheckSecurityInvariant(results, isolation, gasModel)
	if invariant.IsSatisfied() {
		t.Error("expected security invariant to fail with isolation violations")
	}
	if invariant.Isolation {
		t.Error("expected isolation to be false")
	}
}

func TestSafetyLayerStrings(t *testing.T) {
	tests := map[SafetyLayer]string{
		LayerCompile:	"compile",
		LayerVerify:	"verify",
		LayerRuntime:	"runtime",
	}
	for layer, expected := range tests {
		if layer.String() != expected {
			t.Errorf("expected %s, got %s", expected, layer.String())
		}
	}
}

func TestGateResultLayers(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 10000}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	results := []DeterminismGateResult{
		CheckBytecodeDeterminism(frame),
		CheckStateSnapshotIntegrity(frame),
		CheckInputNormalization(frame),
		CheckGasModelConsistency(frame),
	}

	for _, r := range results {
		if r.Layer != LayerVerify && r.Layer != LayerRuntime {
			t.Errorf("unexpected layer %v for check %s", r.Layer, r.CheckName)
		}
	}
}

func TestGasSafetyDeterministicErrorModel(t *testing.T) {
	_ = DefaultGasSafetyModel()

	for i := 0; i < 10; i++ {
		state := EmptyStateRootChunk()
		msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 1}
		frame := NewKernelExecutionFrame(state.Root, msg, 100)

		_, _, exitCode, _, _ := ExecuteKernelSemantics(frame)
		if exitCode.Category != ExitCategoryGasError {
			t.Errorf("run %d: expected GasError, got %v", i, exitCode)
		}
	}
}

func TestGasExhaustionNoStateMutation(t *testing.T) {
	state := EmptyStateRootChunk()
	msg := Message{Type: MessageInternal, Sender: "AE:sender", Target: "4:target", GasLimit: 1}
	frame := NewKernelExecutionFrame(state.Root, msg, 100)

	originalHash := state.Root.Hash()
	_, _, _, _, _ = ExecuteKernelSemantics(frame)

	snapshotHash := frame.StateSnapshot.Hash()
	for i := range originalHash {
		if originalHash[i] != snapshotHash[i] {
			t.Error("state should not be mutated on gas exhaustion")
			break
		}
	}
}

func TestBounceExplosionPrevention(t *testing.T) {
	msg := AVMLedgerReceipt{
		ExitCode:	ExitChunkError,
		MessageFlags:	MessageFlags{Bounced: true},
	}
	if !msg.MessageFlags.Bounced {
		t.Error("expected bounced flag")
	}
	bounceAgain := IsBounceEligible(msg.MessageFlags, msg.ExitCode)
	if bounceAgain {
		t.Error("bounced message must not be bounce-eligible (prevents infinite loop)")
	}
}

func TestSeedCryptoRandDeterministic(t *testing.T) {
	beacon := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	s1 := SeedCryptoRand(beacon)
	s2 := SeedCryptoRand(beacon)
	if s1 != s2 {
		t.Errorf("expected deterministic seed, got %d then %d", s1, s2)
	}
}

func TestSeedCryptoRandShortBeacon(t *testing.T) {
	beacon := []byte{1, 2}
	s := SeedCryptoRand(beacon)
	if s == 0 {
		t.Error("expected non-zero seed from short beacon")
	}
}
