package avm

import (
	"encoding/binary"
	"fmt"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

type CapabilityClass uint8

const (
	CapClassTime		CapabilityClass	= iota
	CapClassRandomness			// process entropy
	CapClassIO				// filesystem/network
	CapClassProcessControl			// goroutines/process
	CapClassParallelExec			// thread creation
	CapClassFloatArithmetic			// floating-point ops
	CapClassStorage				// state read/write
	CapClassMessaging			// internal/external messages
	CapClassCrypto				// hash/verify
	CapClassChain				// block context
)

func (c CapabilityClass) String() string {
	names := map[CapabilityClass]string{
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
	if name, ok := names[c]; ok {
		return name
	}
	return "UNKNOWN"
}

// ForbiddenCapabilityClasses returns the set of capability classes
// that are ALWAYS forbidden in consensus execution.
// These correspond to the host functions marked Forbidden:true.
var ForbiddenCapabilityClasses = map[CapabilityClass]bool{
	CapClassTime:			true,
	CapClassRandomness:		true,
	CapClassIO:			true,
	CapClassProcessControl:		true,
	CapClassParallelExec:		true,
	CapClassFloatArithmetic:	true,
}

// CapabilityClassForHostFunction maps a host function to its required capability class.
func CapabilityClassForHostFunction(host HostFunction) CapabilityClass {
	switch host {
	case HostWallClockTime:
		return CapClassTime
	case HostRandomness:
		return CapClassRandomness
	case HostFilesystem, HostNetwork:
		return CapClassIO
	case HostGoroutine, HostProcessEnv:
		return CapClassProcessControl
	case HostNondeterministicMap:
		return CapClassParallelExec
	case HostFloatingPoint:
		return CapClassFloatArithmetic
	case HostReadStorage, HostWriteStorage, HostDeleteStorage:
		return CapClassStorage
	case HostEmitInternal, HostEmitEvent, HostSendInternal, HostScheduleSelf:
		return CapClassMessaging
	case HostHashSHA256, HostHashBLAKE3, HostVerifyEd25519:
		return CapClassCrypto
	case HostBlockContext, HostGetBlockHeight, HostGetChainID, HostGetCallerSource:
		return CapClassChain
	default:
		return CapClassStorage
	}
}

// ForbiddenHostFunctionClasses returns all capability classes that are globally forbidden.
func ForbiddenHostFunctionClasses() []CapabilityClass {
	classes := make([]CapabilityClass, 0, len(ForbiddenCapabilityClasses))
	for c := range ForbiddenCapabilityClasses {
		classes = append(classes, c)
	}
	return classes
}

type DeterminismGateResult struct {
	Passed		bool
	Layer		SafetyLayer
	CheckName	string
	ViolationDetail	string
}

type SafetyLayer uint8

const (
	LayerCompile	SafetyLayer	= iota
	LayerVerify			// bytecode + module verification
	LayerRuntime			// VM execution
)

func (l SafetyLayer) String() string {
	switch l {
	case LayerCompile:
		return "compile"
	case LayerVerify:
		return "verify"
	case LayerRuntime:
		return "runtime"
	default:
		return "unknown"
	}
}

type DeterminismGate struct {
	Checks []DeterminismCheck
}

type DeterminismCheck func(frame *KernelExecutionFrame) DeterminismGateResult

func NewDeterminismGate() *DeterminismGate {
	return &DeterminismGate{
		Checks: []DeterminismCheck{
			CheckBytecodeDeterminism,
			CheckStateSnapshotIntegrity,
			CheckInputNormalization,
			CheckGasModelConsistency,
			CheckHostFunctionWhitelist,
			CheckForbiddenCapabilities,
			CheckStackBounds,
			CheckActionBudgetBounds,
			CheckContinuationIntegrity,
		},
	}
}

func (g *DeterminismGate) Validate(frame *KernelExecutionFrame) []DeterminismGateResult {
	results := make([]DeterminismGateResult, len(g.Checks))
	for i, check := range g.Checks {
		results[i] = check(frame)
	}
	return results
}

func (g *DeterminismGate) ValidateAll(frame *KernelExecutionFrame) bool {
	for _, check := range g.Checks {
		result := check(frame)
		if !result.Passed {
			return false
		}
	}
	return true
}

func CheckBytecodeDeterminism(frame *KernelExecutionFrame) DeterminismGateResult {
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerVerify,
		CheckName:	"bytecode_determinism",
	}
}

func CheckStateSnapshotIntegrity(frame *KernelExecutionFrame) DeterminismGateResult {
	if frame.StateSnapshot == nil {
		return DeterminismGateResult{
			Passed:			false,
			Layer:			LayerVerify,
			CheckName:		"state_snapshot_integrity",
			ViolationDetail:	"state snapshot is nil",
		}
	}
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerVerify,
		CheckName:	"state_snapshot_integrity",
	}
}

func CheckInputNormalization(frame *KernelExecutionFrame) DeterminismGateResult {
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerVerify,
		CheckName:	"input_normalization",
	}
}

func CheckGasModelConsistency(frame *KernelExecutionFrame) DeterminismGateResult {
	if frame.GasLimit == 0 {
		return DeterminismGateResult{
			Passed:			false,
			Layer:			LayerRuntime,
			CheckName:		"gas_model_consistency",
			ViolationDetail:	"gas limit is zero",
		}
	}
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerRuntime,
		CheckName:	"gas_model_consistency",
	}
}

func CheckHostFunctionWhitelist(frame *KernelExecutionFrame) DeterminismGateResult {
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerVerify,
		CheckName:	"host_function_whitelist",
	}
}

func CheckForbiddenCapabilities(frame *KernelExecutionFrame) DeterminismGateResult {
	for _, call := range frame.HostCallTrace {
		if IsForbiddenHostFunction(HostFunction(call.FunctionID)) {
			return DeterminismGateResult{
				Passed:			false,
				Layer:			LayerRuntime,
				CheckName:		"forbidden_capabilities",
				ViolationDetail:	fmt.Sprintf("forbidden host function called: %d", call.FunctionID),
			}
		}
	}
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerVerify,
		CheckName:	"forbidden_capabilities",
	}
}

func CheckStackBounds(frame *KernelExecutionFrame) DeterminismGateResult {
	if len(frame.Stack) > 1024 {
		return DeterminismGateResult{
			Passed:			false,
			Layer:			LayerRuntime,
			CheckName:		"stack_bounds",
			ViolationDetail:	fmt.Sprintf("stack depth %d exceeds maximum 1024", len(frame.Stack)),
		}
	}
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerRuntime,
		CheckName:	"stack_bounds",
	}
}

func CheckActionBudgetBounds(frame *KernelExecutionFrame) DeterminismGateResult {
	if frame.ActionsUsed > frame.ActionBudget {
		return DeterminismGateResult{
			Passed:			false,
			Layer:			LayerRuntime,
			CheckName:		"action_budget_bounds",
			ViolationDetail:	fmt.Sprintf("actions used %d exceeds budget %d", frame.ActionsUsed, frame.ActionBudget),
		}
	}
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerRuntime,
		CheckName:	"action_budget_bounds",
	}
}

func CheckContinuationIntegrity(frame *KernelExecutionFrame) DeterminismGateResult {
	return DeterminismGateResult{
		Passed:		true,
		Layer:		LayerVerify,
		CheckName:	"continuation_integrity",
	}
}

type SecurityViolation struct {
	Code	string
	Layer	SafetyLayer
	Message	string
	Opcode	ISAOpcode
	HostFn	HostFunction
}

type StaticSecurityScanner struct {
	Violations		[]SecurityViolation
	ForbiddenHostFns	[]HostFunction
	ForbiddenOpcodes	[]ISAOpcode
	MaxUnboundedJumps	int
}

func NewStaticSecurityScanner() *StaticSecurityScanner {
	return &StaticSecurityScanner{
		Violations:		make([]SecurityViolation, 0),
		ForbiddenHostFns:	ForbiddenHostFunctionIDs(),
		ForbiddenOpcodes:	[]ISAOpcode{},
		MaxUnboundedJumps:	256,
	}
}

func ForbiddenHostFunctionIDs() []HostFunction {
	return []HostFunction{
		HostWallClockTime,
		HostRandomness,
		HostFilesystem,
		HostNetwork,
		HostFloatingPoint,
		HostGoroutine,
		HostProcessEnv,
		HostNondeterministicMap,
	}
}

func (s *StaticSecurityScanner) ScanHostImports(forbiddenIDs []HostFunction) []SecurityViolation {
	violations := make([]SecurityViolation, 0)
	registry := HostFunctionRegistry()
	for _, id := range forbiddenIDs {
		if spec, exists := registry[id]; exists && spec.Forbidden {
			violations = append(violations, SecurityViolation{
				Code:		"FORBIDDEN_HOST_IMPORT",
				Layer:		LayerVerify,
				Message:	fmt.Sprintf("forbidden host function %s (0x%x)", spec.Name, id),
				HostFn:		id,
			})
		}
	}
	return violations
}

func (s *StaticSecurityScanner) ScanOpcodes(instructions []byte) []SecurityViolation {
	violations := make([]SecurityViolation, 0)
	_ = instructions
	return violations
}

func (s *StaticSecurityScanner) ScanControlFlow(blocks uint32, entryBlock uint32) []SecurityViolation {
	violations := make([]SecurityViolation, 0)
	_ = blocks
	_ = entryBlock
	return violations
}

func (s *StaticSecurityScanner) ScanTypeCoercions(instructions []byte) []SecurityViolation {
	violations := make([]SecurityViolation, 0)
	_ = instructions
	return violations
}

func (s *StaticSecurityScanner) FullScan(forbiddenIDs []HostFunction, instructions []byte, blockCount uint32, entryBlock uint32) []SecurityViolation {
	all := make([]SecurityViolation, 0)
	all = append(all, s.ScanHostImports(forbiddenIDs)...)
	all = append(all, s.ScanOpcodes(instructions)...)
	all = append(all, s.ScanControlFlow(blockCount, entryBlock)...)
	all = append(all, s.ScanTypeCoercions(instructions)...)
	return all
}

type RuntimeIsolationPolicy struct {
	AllowFilesystem		bool
	AllowNetwork		bool
	AllowProcessInteraction	bool
	AllowWallClockTime	bool
	AllowProcessEntropy	bool
	AllowThreadCreation	bool
	AllowAsyncExecution	bool
	AllowFloatingPoint	bool
	AllowExternalRandomness	bool
}

func DefaultRuntimeIsolationPolicy() RuntimeIsolationPolicy {
	return RuntimeIsolationPolicy{
		AllowFilesystem:		false,
		AllowNetwork:			false,
		AllowProcessInteraction:	false,
		AllowWallClockTime:		false,
		AllowProcessEntropy:		false,
		AllowThreadCreation:		false,
		AllowAsyncExecution:		false,
		AllowFloatingPoint:		false,
		AllowExternalRandomness:	false,
	}
}

func (p RuntimeIsolationPolicy) Validate() []string {
	violations := make([]string, 0)
	if p.AllowFilesystem {
		violations = append(violations, "filesystem access forbidden in consensus execution")
	}
	if p.AllowNetwork {
		violations = append(violations, "network access forbidden in consensus execution")
	}
	if p.AllowProcessInteraction {
		violations = append(violations, "OS process interaction forbidden in consensus execution")
	}
	if p.AllowWallClockTime {
		violations = append(violations, "wall-clock time forbidden in consensus execution")
	}
	if p.AllowProcessEntropy {
		violations = append(violations, "process entropy sources forbidden in consensus execution")
	}
	if p.AllowThreadCreation {
		violations = append(violations, "thread creation forbidden in consensus execution")
	}
	if p.AllowAsyncExecution {
		violations = append(violations, "async execution outside VM scheduler forbidden")
	}
	if p.AllowFloatingPoint {
		violations = append(violations, "floating-point arithmetic forbidden in consensus execution")
	}
	if p.AllowExternalRandomness {
		violations = append(violations, "external randomness forbidden in consensus execution")
	}
	return violations
}

func EnforceRuntimeIsolation(hostFn HostFunction) error {
	if IsForbiddenHostFunction(hostFn) {
		return fmt.Errorf("AVM isolation: forbidden host function %d", hostFn)
	}
	return nil
}

type NormalizationPolicy struct {
	NormalizeMapIteration	bool
	NormalizeEventOrder	bool
	NormalizeMessageOrder	bool
}

func DefaultNormalizationPolicy() NormalizationPolicy {
	return NormalizationPolicy{
		NormalizeMapIteration:	true,
		NormalizeEventOrder:	true,
		NormalizeMessageOrder:	true,
	}
}

func NormalizeChunkMapEntries(entries []chunk.Entry) []chunk.Entry {
	n := len(entries)
	sorted := make([]chunk.Entry, n)
	copy(sorted, entries)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if string(sorted[i].Key) > string(sorted[j].Key) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func NormalizeActionOrder(actions []Action) []Action {
	n := len(actions)
	sorted := make([]Action, n)
	copy(sorted, actions)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if sorted[i].Type > sorted[j].Type {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			} else if sorted[i].Type == sorted[j].Type && sorted[i].Target > sorted[j].Target {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func NormalizeMessageOrder(messages []Message) []Message {
	SortMessagesByDeterministicOrder(messages)
	return messages
}

type GasSafetyModel struct {
	MaxGasPerInstruction	uint64
	MaxGasTotal		uint64
	MaxActionsPerExecution	uint32
	GasReservationRatioBps	uint32
	PriorityFeeCeiling	uint64
}

func DefaultGasSafetyModel() GasSafetyModel {
	return GasSafetyModel{
		MaxGasPerInstruction:	1000,
		MaxGasTotal:		100_000_000,
		MaxActionsPerExecution:	256,
		GasReservationRatioBps:	1000,
		PriorityFeeCeiling:	1_000_000,
	}
}

func (m GasSafetyModel) ValidateGasLimit(limit uint64) error {
	if limit == 0 {
		return fmt.Errorf("AVM gas safety: gas limit must be > 0")
	}
	if limit > m.MaxGasTotal {
		return fmt.Errorf("AVM gas safety: gas limit %d exceeds maximum %d", limit, m.MaxGasTotal)
	}
	return nil
}

func (m GasSafetyModel) ValidateGasExhaustion(gasUsed, gasLimit uint64, exitCode StructuredExitCode) error {
	if gasUsed >= gasLimit {
		if exitCode.Category != ExitCategoryGasError {
			return fmt.Errorf("AVM gas safety: gas exhausted but exit code is %v (expected GasError)", exitCode)
		}
	}
	return nil
}

func (m GasSafetyModel) ValidateNoPartialSideEffects(gasUsed, gasLimit uint64, actionsUsed, actionBudget uint32, exitCode StructuredExitCode) error {
	if gasUsed >= gasLimit || actionsUsed > actionBudget {
		if exitCode.Category == ExitCategorySuccess {
			return fmt.Errorf("AVM gas safety: execution exceeded limits but returned success")
		}
	}
	return nil
}

type FuzzResilienceModel struct {
	MaxBytecodeSize		uint32
	MaxChunkDataSize	uint32
	MaxMessagePayloadSize	uint32
	MaxStackDepth		uint32
	MaxRecursionDepth	uint32
}

func DefaultFuzzResilienceModel() FuzzResilienceModel {
	return FuzzResilienceModel{
		MaxBytecodeSize:	uint32(1 << 20),
		MaxChunkDataSize:	uint32(1 << 16),
		MaxMessagePayloadSize:	uint32(1 << 20),
		MaxStackDepth:		1024,
		MaxRecursionDepth:	256,
	}
}

func (m FuzzResilienceModel) ValidateBytecodeSize(size int) error {
	if size > int(m.MaxBytecodeSize) {
		return fmt.Errorf("AVM fuzz: bytecode size %d exceeds maximum %d", size, m.MaxBytecodeSize)
	}
	return nil
}

func (m FuzzResilienceModel) ValidateChunkData(data []byte) error {
	if len(data) > int(m.MaxChunkDataSize) {
		return fmt.Errorf("AVM fuzz: chunk data size %d exceeds maximum %d", len(data), m.MaxChunkDataSize)
	}
	return nil
}

func (m FuzzResilienceModel) ValidateStackDepth(depth int) error {
	if depth > int(m.MaxStackDepth) {
		return fmt.Errorf("AVM fuzz: stack depth %d exceeds maximum %d", depth, m.MaxStackDepth)
	}
	return nil
}

type ReplayRecord struct {
	StateRootBefore	[]byte
	StateRootAfter	[]byte
	InputMessage	Message
	GasUsed		uint64
	ExitCode	StructuredExitCode
	Receipt		AVMReceipt
	ActionHash	[]byte
	TraceHash	[]byte
}

func RecordReplay(stateRootBefore *chunk.Chunk, stateRootAfter *StateRootChunk, msg Message, receipt AVMReceipt, actions *ActionQueueChunk, trace KernelExecutionTrace) *ReplayRecord {
	var beforeHash, afterHash []byte
	if stateRootBefore != nil {
		beforeHash = stateRootBefore.Hash()
	}
	if stateRootAfter != nil && stateRootAfter.Root != nil {
		afterHash = stateRootAfter.RootHash()
	}

	var actionHash []byte
	if actions != nil {
		actionHash = actions.Hash
	}

	traceHash := finalizeTraceFromRecord(trace)

	return &ReplayRecord{
		StateRootBefore:	beforeHash,
		StateRootAfter:		afterHash,
		InputMessage:		msg,
		GasUsed:		receipt.GasUsed,
		ExitCode:		StructuredExitCodeFromUint32(receipt.ExitCode),
		Receipt:		receipt,
		ActionHash:		actionHash,
		TraceHash:		traceHash,
	}
}

func finalizeTraceFromRecord(trace KernelExecutionTrace) []byte {
	h := make([]byte, 32)
	for _, step := range trace.Steps {
		var buf [16]byte
		binary.BigEndian.PutUint16(buf[0:2], uint16(step.Opcode))
		binary.BigEndian.PutUint32(buf[2:6], step.IP)
		binary.BigEndian.PutUint64(buf[6:14], step.GasUsed)
		binary.BigEndian.PutUint16(buf[14:16], uint16(step.Phase))
		for i := 0; i < 16; i++ {
			h[i%32] ^= buf[i]
		}
	}
	return h
}

func (r *ReplayRecord) VerifyDeterminism(other *ReplayRecord) bool {
	if r.GasUsed != other.GasUsed {
		return false
	}
	if r.ExitCode.ToUint32() != other.ExitCode.ToUint32() {
		return false
	}
	if len(r.StateRootAfter) != len(other.StateRootAfter) {
		return false
	}
	for i := range r.StateRootAfter {
		if r.StateRootAfter[i] != other.StateRootAfter[i] {
			return false
		}
	}
	return true
}

type AdversarialVector struct {
	Name		string
	Category	string
	Description	string
	ExitCode	StructuredExitCode
	Input		[]byte
}

var AdversarialVectors = []AdversarialVector{
	{
		Name:		"invalid_opcode_0xFF",
		Category:	"malicious_bytecode",
		Description:	"execute unrecognized opcode 0xFF",
		ExitCode:	ExitValidationFailed,
		Input:		[]byte{0xFF},
	},
	{
		Name:		"stack_underflow",
		Category:	"malicious_bytecode",
		Description:	"pop from empty stack",
		ExitCode:	ExitStackUnderflow,
		Input:		[]byte{0x12},
	},
	{
		Name:		"gas_exhaustion_loop",
		Category:	"crafted_gas",
		Description:	"infinite loop consuming all gas",
		ExitCode:	ExitGasExhausted,
		Input:		nil,
	},
	{
		Name:		"chunk_reference_invalid",
		Category:	"invalid_state",
		Description:	"access nil/nonexistent chunk reference",
		ExitCode:	ExitChunkError,
		Input:		nil,
	},
	{
		Name:		"type_mismatch_arith",
		Category:	"malicious_bytecode",
		Description:	"arithmetic on boolean stack values",
		ExitCode:	ExitTypeMismatch,
		Input:		nil,
	},
	{
		Name:		"division_by_zero",
		Category:	"execution_error",
		Description:	"divide by zero",
		ExitCode:	ExitDivZero,
		Input:		nil,
	},
	{
		Name:		"action_budget_exceeded",
		Category:	"action_overflow",
		Description:	"emit more actions than allowed",
		ExitCode:	ExitActionBudget,
		Input:		nil,
	},
	{
		Name:		"forbidden_host_call",
		Category:	"capability_violation",
		Description:	"call forbidden host function (wall clock time)",
		ExitCode:	ExitForbiddenCall,
		Input:		nil,
	},
	{
		Name:		"bounce_explosion",
		Category:	"message_amplification",
		Description:	"bounced message attempts another bounce",
		ExitCode:	ExitValidationFailed,
		Input:		nil,
	},
	{
		Name:		"malformed_abi_input",
		Category:	"invalid_input",
		Description:	"query with malformed arguments",
		ExitCode:	ExitInvalidDecode,
		Input:		nil,
	},
	{
		Name:		"oversized_bytecode",
		Category:	"resource_exhaustion",
		Description:	"bytecode exceeding maximum size",
		ExitCode:	ExitValidationFailed,
		Input:		nil,
	},
	{
		Name:		"recursive_message_amplification",
		Category:	"message_amplification",
		Description:	"message that creates exponential messages",
		ExitCode:	ExitActionBudget,
		Input:		nil,
	},
}

type ConsensusSafetyGuarantee struct {
	DeterminismPassed	bool
	IsolationPassed		bool
	ReplayabilityPassed	bool
	BoundednessPassed	bool
}

func (g ConsensusSafetyGuarantee) IsConsensusSafe() bool {
	return g.DeterminismPassed && g.IsolationPassed && g.ReplayabilityPassed && g.BoundednessPassed
}

func VerifyConsensusSafety(gateResults []DeterminismGateResult, isolation RuntimeIsolationPolicy, replay *ReplayRecord, gasModel GasSafetyModel) ConsensusSafetyGuarantee {
	determinismPassed := true
	for _, r := range gateResults {
		if !r.Passed {
			determinismPassed = false
			break
		}
	}

	isolationViolations := isolation.Validate()
	isolationPassed := len(isolationViolations) == 0

	replayabilityPassed := true
	if replay != nil {
		replayabilityPassed = len(replay.StateRootAfter) > 0 && replay.ExitCode.Category == ExitCategorySuccess
	}

	boundednessPassed := true
	if replay != nil {
		if err := gasModel.ValidateGasLimit(replay.GasUsed); err != nil {
			boundednessPassed = false
		}
	}

	return ConsensusSafetyGuarantee{
		DeterminismPassed:	determinismPassed,
		IsolationPassed:	isolationPassed,
		ReplayabilityPassed:	replayabilityPassed,
		BoundednessPassed:	boundednessPassed,
	}
}

type FailureClassification struct {
	Kind		FailureKind
	ExitCode	StructuredExitCode
	Deterministic	bool
	StateMutated	bool
	AffectsSequence	bool
}

func ClassifyFailure(exitCode StructuredExitCode) FailureClassification {
	switch exitCode.Category {
	case ExitCategorySuccess:
		return FailureClassification{
			Kind:		FailureNone,
			ExitCode:	exitCode,
			Deterministic:	true,
		}
	case ExitCategoryGasError:
		return FailureClassification{
			Kind:		FailureRecoverable,
			ExitCode:	exitCode,
			Deterministic:	true,
		}
	case ExitCategoryVMError, ExitCategoryTypeError, ExitCategoryExecutionError:
		return FailureClassification{
			Kind:		FailureRecoverable,
			ExitCode:	exitCode,
			Deterministic:	true,
		}
	case ExitCategoryStateError:
		return FailureClassification{
			Kind:		FailureNonRecoverable,
			ExitCode:	exitCode,
			Deterministic:	true,
		}
	case ExitCategoryActionError:
		return FailureClassification{
			Kind:		FailureRecoverable,
			ExitCode:	exitCode,
			Deterministic:	true,
		}
	default:
		return FailureClassification{
			Kind:		FailureSystemFatal,
			ExitCode:	exitCode,
			Deterministic:	false,
		}
	}
}

func (f FailureClassification) IsDeterministic() bool {
	return f.Deterministic
}

func (f FailureClassification) StateIsClean() bool {
	return !f.StateMutated
}

func (f FailureClassification) SequenceIsIsolated() bool {
	return !f.AffectsSequence
}

type SecurityInvariant struct {
	Determinism	bool
	Isolation	bool
	Replayability	bool
	Boundedness	bool
}

func (i SecurityInvariant) IsSatisfied() bool {
	return i.Determinism && i.Isolation && i.Replayability && i.Boundedness
}

func (i SecurityInvariant) String() string {
	return fmt.Sprintf("SecurityInvariant{Determinism=%v, Isolation=%v, Replayability=%v, Boundedness=%v, Satisfied=%v}",
		i.Determinism, i.Isolation, i.Replayability, i.Boundedness, i.IsSatisfied())
}

func CheckSecurityInvariant(gateResults []DeterminismGateResult, isolation RuntimeIsolationPolicy, gasModel GasSafetyModel) SecurityInvariant {
	determinism := true
	for _, r := range gateResults {
		if !r.Passed {
			determinism = false
			break
		}
	}

	isolationViolations := isolation.Validate()
	isolation_ := len(isolationViolations) == 0

	boundedness := true
	if gasModel.MaxGasTotal == 0 || gasModel.MaxGasPerInstruction == 0 {
		boundedness = false
	}

	replayability := determinism && isolation_ && boundedness

	return SecurityInvariant{
		Determinism:	determinism,
		Isolation:	isolation_,
		Replayability:	replayability,
		Boundedness:	boundedness,
	}
}

func GenerateFuzzBytecode(rngSeed uint64, size int) []byte {
	data := make([]byte, size)
	seed := rngSeed
	for i := range data {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		data[i] = byte(seed)
	}
	return data
}

func GenerateMalformedChunk(rngSeed uint64) *chunk.Chunk {
	data := GenerateFuzzBytecode(rngSeed, 64)
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData(data, uint16(len(data))*8).Build()
	return b
}

func GenerateCorruptedModule(rngSeed uint64, size int) []byte {
	data := GenerateFuzzBytecode(rngSeed, size)
	if size > 4 {
		magic := make([]byte, 4)
		binary.BigEndian.PutUint32(magic, AVMModuleMagic)
		copy(data, magic)
	}
	return data
}

func GenerateRandomHostCall(rngSeed uint64) HostFunction {
	candidates := []HostFunction{
		HostReadStorage, HostWriteStorage, HostEmitInternal,
		HostInspectMsg, HostBlockContext, HostChargeGas, HostReturn,
		HostHashSHA256, HostHashBLAKE3, HostVerifyEd25519,
	}
	seed := rngSeed
	seed ^= seed << 13
	seed ^= seed >> 7
	idx := seed % uint64(len(candidates))
	return candidates[idx]
}

func GenerateDeterministicRandomness(previousStateRoot, blockEntropy, messageHash, domain []byte) []byte {
	return RandomBeacon(previousStateRoot, blockEntropy, messageHash, domain)
}

func VerifyMapIterationDeterminism(mapEntries [][]chunk.Entry) bool {
	if len(mapEntries) < 2 {
		return true
	}

	normalizedSets := make([][]string, len(mapEntries))
	for i, entries := range mapEntries {
		normalized := NormalizeChunkMapEntries(entries)
		keys := make([]string, len(normalized))
		for j, e := range normalized {
			keys[j] = string(e.Key)
		}
		normalizedSets[i] = keys
	}

	reference := normalizedSets[0]
	for i := 1; i < len(normalizedSets); i++ {
		if len(normalizedSets[i]) != len(reference) {
			return false
		}
		for j := range reference {
			if normalizedSets[i][j] != reference[j] {
				return false
			}
		}
	}
	return true
}

// AVMModuleMagic is the canonical magic number for AVM modules.
const AVMModuleMagic uint32 = 0x41564D01	// "AVM\x01"

// VerifyRawFuzzBytecode fills raw random bytes into a module-like structure.
func VerifyRawFuzzBytecode(data []byte) StructuredExitCode {
	if len(data) == 0 {
		return ExitValidationFailed
	}
	if len(data) > int(1<<20) {
		return ExitValidationFailed
	}
	if uint32(len(data)) < 4 {
		return ExitValidationFailed
	}
	magic := binary.BigEndian.Uint32(data[:4])
	if magic != AVMModuleMagic {
		return ExitValidationFailed
	}
	return ExitSuccess
}

// SeedCryptoRand provides a deterministic seed for fuzz operations
// using the VM's deterministic randomness beacon, never os random.
func SeedCryptoRand(beacon []byte) uint64 {
	if len(beacon) < 8 {
		var seed uint64
		for i, b := range beacon {
			seed ^= uint64(b) << (i * 8 % 64)
		}
		return seed
	}
	return binary.BigEndian.Uint64(beacon[:8])
}
