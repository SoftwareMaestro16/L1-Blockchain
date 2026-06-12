package avm

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

// ContinuationSlot defines the 4 continuation pointers for AVM control flow.
// Replaces the TVM c0-c7 model with a deterministic, continuation-driven system.
//
// Invariants:
//   - return_ptr → normal completion continuation
//   - alt_return_ptr → alternate execution path (e.g., bounced messages)
//   - error_handler_ptr → exception recovery path
//   - dispatcher_ptr → entry point resolver for contract logic
//   - All slots are IMMUTABLE during execution
//   - Slots are replaced ONLY via continuation transitions
type ContinuationSlot struct {
	ReturnPtr	uint32
	AltReturnPtr	uint32
	ErrorHandlerPtr	uint32
	DispatcherPtr	uint32
}

// ExecutionSlot defines named VM control registers (rebranded from TVM c0-c7).
// All slots MUST be immutable during execution and replaced only via continuation transitions.
type ExecutionSlot int

const (
	SlotReturn	ExecutionSlot	= iota	// Normal continuation
	SlotAltReturn				// Alternate continuation (e.g., bounced)
	SlotError				// Exception handler
	SlotDispatch				// Contract entry resolver
	SlotState				// StateRootChunk reference
	SlotActions				// ActionQueueChunk reference
	SlotEnv					// ExecutionContextChunk reference
)

func (s ExecutionSlot) String() string {
	switch s {
	case SlotReturn:
		return "SLOT_RETURN"
	case SlotAltReturn:
		return "SLOT_ALT_RETURN"
	case SlotError:
		return "SLOT_ERROR"
	case SlotDispatch:
		return "SLOT_DISPATCH"
	case SlotState:
		return "SLOT_STATE"
	case SlotActions:
		return "SLOT_ACTIONS"
	case SlotEnv:
		return "SLOT_ENV"
	default:
		return "SLOT_UNKNOWN"
	}
}

// ExitCategory defines the category of an AVM exit code.
type ExitCategory uint8

const (
	ExitCategorySuccess		ExitCategory	= 0
	ExitCategoryVMError		ExitCategory	= 1
	ExitCategoryTypeError		ExitCategory	= 2
	ExitCategoryExecutionError	ExitCategory	= 3
	ExitCategoryActionError		ExitCategory	= 4
	ExitCategoryStateError		ExitCategory	= 5
	ExitCategoryGasError		ExitCategory	= 6
)

func (c ExitCategory) String() string {
	switch c {
	case ExitCategorySuccess:
		return "SUCCESS"
	case ExitCategoryVMError:
		return "VM_ERROR"
	case ExitCategoryTypeError:
		return "TYPE_ERROR"
	case ExitCategoryExecutionError:
		return "EXECUTION_ERROR"
	case ExitCategoryActionError:
		return "ACTION_ERROR"
	case ExitCategoryStateError:
		return "STATE_ERROR"
	case ExitCategoryGasError:
		return "GAS_ERROR"
	default:
		return "UNKNOWN"
	}
}

// StructuredExitCode replaces flat exit codes with a structured model.
//
// Invariants:
//   - VM errors MUST NOT mutate state
//   - Only SUCCESS reaches commit phase
//   - All failures produce deterministic receipt
type StructuredExitCode struct {
	Category	ExitCategory
	Subcode		uint16
	MessageHex	string
}

func (e StructuredExitCode) ToUint32() uint32 {
	return uint32(e.Category)<<16 | uint32(e.Subcode)
}

func StructuredExitCodeFromUint32(code uint32) StructuredExitCode {
	return StructuredExitCode{
		Category:	ExitCategory(code >> 16),
		Subcode:	uint16(code & 0xFFFF),
	}
}

// Predefined structured exit codes
var (
	ExitSuccess		= StructuredExitCode{ExitCategorySuccess, 0, ""}
	ExitValidationFailed	= StructuredExitCode{ExitCategoryVMError, 1, ""}
	ExitUnauthorized	= StructuredExitCode{ExitCategoryVMError, 2, ""}
	ExitStackOverflow	= StructuredExitCode{ExitCategoryVMError, 3, ""}
	ExitStackUnderflow	= StructuredExitCode{ExitCategoryVMError, 4, ""}
	ExitInvalidJump		= StructuredExitCode{ExitCategoryVMError, 5, ""}
	ExitCallStackOverflow	= StructuredExitCode{ExitCategoryVMError, 6, ""}
	ExitTypeMismatch	= StructuredExitCode{ExitCategoryTypeError, 1, ""}
	ExitInvalidDecode	= StructuredExitCode{ExitCategoryTypeError, 2, ""}
	ExitChunkError		= StructuredExitCode{ExitCategoryStateError, 1, ""}
	ExitStateCorruption	= StructuredExitCode{ExitCategoryStateError, 2, ""}
	ExitStateMutation	= StructuredExitCode{ExitCategoryStateError, 3, ""}
	ExitDivZero		= StructuredExitCode{ExitCategoryExecutionError, 1, ""}
	ExitOverflow		= StructuredExitCode{ExitCategoryExecutionError, 2, ""}
	ExitGasExhausted	= StructuredExitCode{ExitCategoryGasError, 1, ""}
	ExitGasLimit		= StructuredExitCode{ExitCategoryGasError, 2, ""}
	ExitActionBudget	= StructuredExitCode{ExitCategoryActionError, 1, ""}
	ExitContractAbort	= StructuredExitCode{ExitCategoryExecutionError, 3, ""}
	ExitForbiddenCall	= StructuredExitCode{ExitCategoryVMError, 7, ""}
)

// ExecutionContextChunk replaces the runtime tuple with an immutable,
// content-addressed environment chunk.
//
// Invariants:
//   - No wall-clock time allowed
//   - No process-level entropy
//   - All values come from block-provided context
//   - Immutable during execution
type ExecutionContextChunk struct {
	Caller		string
	Origin		string
	AttachedValue	uint64
	BlockHeight	int64
	ChainID		string
	ContractAddress	string
	MessageHash	[]byte
	Timestamp	int64
}

// ToChunk encodes the ExecutionContextChunk into an immutable Chunk.
func (ctx ExecutionContextChunk) ToChunk() (*chunk.Chunk, error) {
	builder := chunk.NewBuilder()
	builder.SetTypeTag(chunk.TypeNormal)
	buf := make([]byte, 0, 256)
	buf = appendUint64BigEndian(buf, uint64(ctx.BlockHeight))
	buf = appendUint64BigEndian(buf, ctx.AttachedValue)
	buf = appendUint64BigEndian(buf, uint64(ctx.Timestamp))
	buf = appendString(buf, ctx.Caller)
	buf = appendString(buf, ctx.Origin)
	buf = appendString(buf, ctx.ChainID)
	buf = appendString(buf, ctx.ContractAddress)
	if len(ctx.MessageHash) > 0 {
		buf = appendUint32BigEndian(buf, uint32(len(ctx.MessageHash)))
		buf = append(buf, ctx.MessageHash...)
	} else {
		buf = appendUint32BigEndian(buf, 0)
	}
	builder.SetData(buf, uint16(len(buf)*8))
	return builder.Build()
}

// FromBlockContext creates an ExecutionContextChunk from an existing BlockContext.
func ExecutionContextFromBlockContext(blockCtx BlockContext, msg Message) ExecutionContextChunk {
	return ExecutionContextChunk{
		Caller:			msg.Sender,
		AttachedValue:		msg.Value,
		BlockHeight:		blockCtx.Height,
		ChainID:		blockCtx.ChainID,
		ContractAddress:	msg.Target,
		MessageHash:		msg.Hash,
		Timestamp:		blockCtx.Timestamp,
	}
}

// ActionQueueChunk represents the accumulated actions during execution.
// Actions are NOT executed immediately — they are flushed only during Finalization Phase.
//
// Invariants:
//   - Action ordering MUST be deterministic
//   - Actions are accumulated during compute phase
//   - Actions are NOT executed immediately
//   - Actions are flushed only during Finalization Phase
//   - No mutation of existing queue entries
type ActionQueueChunk struct {
	Actions	[]Action
	Hash	[]byte
}

// NewActionQueueChunk creates an empty action queue.
func NewActionQueueChunk() *ActionQueueChunk {
	return &ActionQueueChunk{
		Actions:	make([]Action, 0),
		Hash:		nil,
	}
}

// EmitAction adds an action to the queue (accumulated during compute phase).
func (q *ActionQueueChunk) EmitAction(actionType ActionType, target string, payload *chunk.Chunk, value uint64) {
	q.Actions = append(q.Actions, Action{
		Type:		actionType,
		Target:		target,
		Payload:	payload,
		Value:		value,
	})
}

// EmitMessage adds an internal message action.
func (q *ActionQueueChunk) EmitMessage(target string, payload *chunk.Chunk, value uint64, gasLimit uint64) {
	q.Actions = append(q.Actions, Action{
		Type:		ActionInternal,
		Target:		target,
		Payload:	payload,
		Value:		value,
	})
}

// EmitEvent adds an event notification.
func (q *ActionQueueChunk) EmitEvent(payload *chunk.Chunk) {
	q.Actions = append(q.Actions, Action{
		Type:		ActionEvent,
		Payload:	payload,
	})
}

// Finalize sorts actions deterministically and computes hash.
func (q *ActionQueueChunk) Finalize() []byte {
	sortActionsCanonical(q.Actions)
	h := sha256.New()
	for _, a := range q.Actions {
		h.Write([]byte{byte(a.Type)})
		h.Write([]byte(a.Target))
		if a.Payload != nil {
			h.Write(a.Payload.Hash())
		}
		h.Write(make([]byte, 8))
		binary.BigEndian.PutUint64(h.Sum(nil)[:8], a.Value)
	}
	q.Hash = h.Sum(nil)
	return q.Hash
}

// sortActionsCanonical sorts actions by (type, target) for deterministic ordering.
func sortActionsCanonical(actions []Action) {
	sort.SliceStable(actions, func(i, j int) bool {
		if actions[i].Type != actions[j].Type {
			return actions[i].Type < actions[j].Type
		}
		return actions[i].Target < actions[j].Target
	})
}

// StateRootChunk represents the immutable root of contract state.
// All persistent contract state MUST be stored in Chunk graphs.
// No mutable memory allowed.
// State transitions MUST produce new root Chunk.
//
// Invariants:
//   - storage → StateRootChunk
//   - persistent variables → sub-chunks inside state graph
//   - StateRootChunk is content-addressed and immutable
//   - Each state transition produces exactly one new StateRootChunk
type StateRootChunk struct {
	Root *chunk.Chunk
}

// NewStateRootChunk creates a state root from an existing Chunk.
func NewStateRootChunk(root *chunk.Chunk) *StateRootChunk {
	return &StateRootChunk{Root: root}
}

// EmptyStateRootChunk creates a state root from an empty ChunkMap.
func EmptyStateRootChunk() *StateRootChunk {
	emptyMap := chunk.NewEmptyMap()
	emptyKey := []byte("__empty__")
	emptyVal, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeSystem).SetData([]byte{}, 0).Build()
	rootMap, _ := emptyMap.Put(emptyKey, emptyVal)
	return &StateRootChunk{Root: rootMap.Root()}
}

// RootHash returns the hash of the state root chunk.
func (s *StateRootChunk) RootHash() []byte {
	if s.Root == nil {
		return nil
	}
	return s.Root.Hash()
}

// KernelExecutionFrame is the expanded VM context layer for the kernel architecture.
//
//	ExecutionFrame {
//	  ip (instruction pointer)
//	  stack_snapshot
//	  local_context_chunk
//	  continuation_slot
//	  env_chunk
//	  pending_actions
//	  error_state
//	}
type KernelExecutionFrame struct {
	// Instruction pointer
	IP	uint32

	// Typed evaluation stack: int256, bool, ChunkRef, ExecutionFrameRef, tuple, address, hash
	Stack	[]StackValue

	// Continuation slot (replaces call stack)
	Continuation	ContinuationSlot

	// Current execution phase
	Phase	Phase

	// Immutable state snapshot (read-only)
	StateSnapshot	*chunk.Chunk

	// Working state root (write target)
	WorkingState	*chunk.Chunk

	// Local context chunk (replaces temporary VM state storage)
	LocalContext	*chunk.Chunk

	// Environment context (immutable during execution)
	EnvChunk	ExecutionContextChunk

	// Action queue (accumulated, not executed immediately)
	ActionQueue	*ActionQueueChunk

	// Error state (structured)
	ErrorState	StructuredExitCode

	// Gas tracking
	GasLimit	uint64
	GasUsed		uint64
	PhaseGas	map[Phase]uint64

	// Capabilities
	Capabilities	CapabilityMask

	// Block context (immutable)
	BlockCtx	BlockContext

	// Message being processed
	Message	Message

	// Execution trace for determinism verification
	Trace	KernelExecutionTrace

	// Abort flag
	Aborted	bool

	// Action budget
	ActionBudget	uint32
	ActionsUsed	uint32

	// Host call trace for auditing
	HostCallTrace	[]HostCallRecord
}

// NewKernelExecutionFrame creates a new kernel execution frame.
func NewKernelExecutionFrame(state *chunk.Chunk, msg Message, maxActions uint32) *KernelExecutionFrame {
	return &KernelExecutionFrame{
		IP:		0,
		Stack:		make([]StackValue, 0),
		Continuation:	ContinuationSlot{},
		Phase:		PhaseStorage,
		StateSnapshot:	state,
		WorkingState:	state,
		ActionQueue:	NewActionQueueChunk(),
		EnvChunk:	ExecutionContextFromBlockContext(BlockContext{}, msg),
		GasLimit:	msg.GasLimit,
		PhaseGas:	make(map[Phase]uint64),
		Capabilities:	AllowAllCapabilities,
		Message:	msg,
		ActionBudget:	maxActions,
		Trace:		KernelExecutionTrace{Steps: make([]KernelTraceStep, 0)},
	}
}

// StackValueType defines the type of a value on the AVM evaluation stack.
type StackValueType uint8

const (
	StackTypeInt256	StackValueType	= iota
	StackTypeBool
	StackTypeChunkRef
	StackTypeFrameRef
	StackTypeTuple
	StackTypeAddress
	StackTypeHash
	StackTypeCoins
	StackTypeString
	StackTypeBytes
	StackTypeNull
)

func (t StackValueType) String() string {
	switch t {
	case StackTypeInt256:
		return "int256"
	case StackTypeBool:
		return "bool"
	case StackTypeChunkRef:
		return "ChunkRef"
	case StackTypeFrameRef:
		return "ExecutionFrameRef"
	case StackTypeTuple:
		return "tuple"
	case StackTypeAddress:
		return "address"
	case StackTypeHash:
		return "hash"
	case StackTypeCoins:
		return "coins"
	case StackTypeString:
		return "string"
	case StackTypeBytes:
		return "bytes"
	case StackTypeNull:
		return "null"
	default:
		return "unknown"
	}
}

// StackValue represents a typed value on the AVM evaluation stack.
// The stack is strictly typed at the opcode level — no implicit conversions.
// Invalid type usage → immediate trap (EXIT_TYPE_ERROR).
type StackValue struct {
	Type		StackValueType
	IntVal		int64
	UintVal		uint64
	BoolVal		bool
	ChunkVal	*chunk.Chunk
	BytesVal	[]byte
	StrVal		string
	AddrVal		string
	TupleVal	[]StackValue
}

// StackValueInt256 creates a signed integer stack value.
func StackValueInt256(val int64) StackValue {
	return StackValue{Type: StackTypeInt256, IntVal: val}
}

// StackValueUint256 creates an unsigned integer stack value.
func StackValueUint256(val uint64) StackValue {
	return StackValue{Type: StackTypeInt256, UintVal: val}
}

// StackValueBool creates a boolean stack value.
func StackValueBool(val bool) StackValue {
	return StackValue{Type: StackTypeBool, BoolVal: val}
}

// StackValueChunkRef creates a Chunk reference stack value.
func StackValueChunkRef(c *chunk.Chunk) StackValue {
	return StackValue{Type: StackTypeChunkRef, ChunkVal: c}
}

// StackValueAddress creates an address stack value.
func StackValueAddress(addr string) StackValue {
	return StackValue{Type: StackTypeAddress, AddrVal: addr}
}

// StackValueHash creates a hash stack value.
func StackValueHash(hash []byte) StackValue {
	return StackValue{Type: StackTypeHash, BytesVal: hash}
}

// StackValueNull creates a null stack value.
func StackValueNull() StackValue {
	return StackValue{Type: StackTypeNull}
}

// StackValueCoins creates a coins value.
func StackValueCoins(val uint64) StackValue {
	return StackValue{Type: StackTypeCoins, UintVal: val}
}

// KernelTraceStep records a single step of kernel execution.
type KernelTraceStep struct {
	Opcode		ISAOpcode
	StackDepth	int
	GasUsed		uint64
	Phase		Phase
	IP		uint32
}

// KernelExecutionTrace holds the complete deterministic trace.
type KernelExecutionTrace struct {
	Steps []KernelTraceStep
}

// ISAOpcode defines the AVM instruction set architecture opcodes.
// Every instruction MUST define:
//   - mnemonic
//   - opcode
//   - stack input signature
//   - stack output signature
//   - gas cost model
//   - overflow behavior
//   - failure exit code
//   - determinism rule
type ISAOpcode uint16

const (
	// Stack operations
	OpISANop	ISAOpcode	= 0x00
	OpISAPush	ISAOpcode	= 0x01
	OpISADup	ISAOpcode	= 0x10
	OpISASwap	ISAOpcode	= 0x11
	OpISADrop	ISAOpcode	= 0x12
	OpISAOver	ISAOpcode	= 0x13

	// Arithmetic (checked, on int256)
	OpISAAdd	ISAOpcode	= 0x20
	OpISASub	ISAOpcode	= 0x21
	OpISAMul	ISAOpcode	= 0x22
	OpISADiv	ISAOpcode	= 0x23
	OpISAMod	ISAOpcode	= 0x24

	// Comparison
	OpISAEq		ISAOpcode	= 0x30
	OpISANeq	ISAOpcode	= 0x31
	OpISALt		ISAOpcode	= 0x32
	OpISALte	ISAOpcode	= 0x33
	OpISAGt		ISAOpcode	= 0x34
	OpISAGte	ISAOpcode	= 0x35

	// Boolean
	OpISAAnd	ISAOpcode	= 0x40
	OpISAOr		ISAOpcode	= 0x41
	OpISANot	ISAOpcode	= 0x42

	// Control flow (continuation-aware)
	OpISACallFrame		ISAOpcode	= 0x50
	OpISAReturnFrame	ISAOpcode	= 0x51
	OpISAJumpCond		ISAOpcode	= 0x52
	OpISAJumpUncond		ISAOpcode	= 0x53
	OpISARaiseError		ISAOpcode	= 0x54
	OpISATryBegin		ISAOpcode	= 0x55
	OpISATryEnd		ISAOpcode	= 0x56

	// State operations (Chunk DAG)
	OpISALoadState	ISAOpcode	= 0x60
	OpISAStoreState	ISAOpcode	= 0x61
	OpISACloneState	ISAOpcode	= 0x62
	OpISAMergeState	ISAOpcode	= 0x63

	// ChunkMap operations
	OpISAChunkMapGet	ISAOpcode	= 0x70
	OpISAChunkMapPut	ISAOpcode	= 0x71
	OpISAChunkMapDelete	ISAOpcode	= 0x72
	OpISAChunkMapProof	ISAOpcode	= 0x73

	// Action system
	OpISAEmitAction		ISAOpcode	= 0x80
	OpISAQueueMessage	ISAOpcode	= 0x81
	OpISAFlushActions	ISAOpcode	= 0x82

	// Context operations
	OpISAGetCaller		ISAOpcode	= 0x90
	OpISAGetOrigin		ISAOpcode	= 0x91
	OpISAGetValue		ISAOpcode	= 0x92
	OpISAGetBlockHeight	ISAOpcode	= 0x93
	OpISAGetChainID		ISAOpcode	= 0x94
	OpISAGetAddress		ISAOpcode	= 0x95
	OpISAGetBody		ISAOpcode	= 0x96
	OpISAGetQueryID		ISAOpcode	= 0x97

	// Cryptographic operations
	OpISAHashChunk	ISAOpcode	= 0xA0
	OpISAHashData	ISAOpcode	= 0xA1
	OpISAVerifySig	ISAOpcode	= 0xA2

	// Type codec operations
	OpISAEncode	ISAOpcode	= 0xB0
	OpISADecode	ISAOpcode	= 0xB1

	// Original AVM opcodes (backward compatible)
	OpISAReadStorage	ISAOpcode	= 0x02
	OpISAWriteStorage	ISAOpcode	= 0x03
	OpISAAddLegacy		ISAOpcode	= 0x04
	OpISAReturn		ISAOpcode	= 0x06
)

// ISAInstructionSpec defines the complete specification for each ISA opcode.
type ISAInstructionSpec struct {
	Mnemonic		string
	Opcode			ISAOpcode
	StackInputs		int
	StackOutputs		int
	InputTypes		[]StackValueType
	OutputTypes		[]StackValueType
	BaseGasCost		uint64
	StackGasCost		uint64
	ChunkGasCost		uint64
	OverflowBehavior	string
	FailureExitCode		StructuredExitCode
	DeterminismRule		string
}

// ISAOpcodeTable defines the complete ISA specification.
// Every instruction has formal stack contracts, gas rules, overflow behavior, and exit codes.
var ISAOpcodeTable = map[ISAOpcode]ISAInstructionSpec{
	OpISANop: {
		Mnemonic:	"nop", Opcode: OpISANop,
		StackInputs:	0, StackOutputs: 0,
		InputTypes:	nil, OutputTypes: nil,
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "always succeeds",
	},
	OpISAPush: {
		Mnemonic:	"push", Opcode: OpISAPush,
		StackInputs:	0, StackOutputs: 1,
		InputTypes:	nil, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "constant value, always succeeds",
	},
	OpISADup: {
		Mnemonic:	"dup", Opcode: OpISADup,
		StackInputs:	1, StackOutputs: 2,
		InputTypes:	[]StackValueType{StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256, StackTypeInt256},
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitStackOverflow, DeterminismRule: "pure, always deterministic",
	},
	OpISASwap: {
		Mnemonic:	"swap", Opcode: OpISASwap,
		StackInputs:	2, StackOutputs: 2,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256, StackTypeInt256},
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitStackUnderflow, DeterminismRule: "pure, always deterministic",
	},
	OpISADrop: {
		Mnemonic:	"drop", Opcode: OpISADrop,
		StackInputs:	1, StackOutputs: 0,
		InputTypes:	[]StackValueType{StackTypeInt256}, OutputTypes: nil,
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitStackUnderflow, DeterminismRule: "pure, always deterministic",
	},
	OpISAAdd: {
		Mnemonic:	"add", Opcode: OpISAAdd,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	3, OverflowBehavior: "trap on overflow",
		FailureExitCode:	ExitOverflow, DeterminismRule: "checked arithmetic, deterministic",
	},
	OpISASub: {
		Mnemonic:	"sub", Opcode: OpISASub,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	3, OverflowBehavior: "trap on underflow",
		FailureExitCode:	ExitOverflow, DeterminismRule: "checked arithmetic, deterministic",
	},
	OpISAMul: {
		Mnemonic:	"mul", Opcode: OpISAMul,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	5, OverflowBehavior: "trap on overflow",
		FailureExitCode:	ExitOverflow, DeterminismRule: "checked arithmetic, deterministic",
	},
	OpISADiv: {
		Mnemonic:	"div", Opcode: OpISADiv,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	5, OverflowBehavior: "trap on division by zero",
		FailureExitCode:	ExitDivZero, DeterminismRule: "trap on div by zero, else deterministic",
	},
	OpISAMod: {
		Mnemonic:	"mod", Opcode: OpISAMod,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	5, OverflowBehavior: "trap on division by zero",
		FailureExitCode:	ExitDivZero, DeterminismRule: "trap on div by zero, else deterministic",
	},
	OpISAEq: {
		Mnemonic:	"eq", Opcode: OpISAEq,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pure, always deterministic",
	},
	OpISAAnd: {
		Mnemonic:	"and", Opcode: OpISAAnd,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeBool, StackTypeBool}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitTypeMismatch, DeterminismRule: "pure, always deterministic",
	},
	OpISAOr: {
		Mnemonic:	"or", Opcode: OpISAOr,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeBool, StackTypeBool}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitTypeMismatch, DeterminismRule: "pure, always deterministic",
	},
	OpISANot: {
		Mnemonic:	"not", Opcode: OpISANot,
		StackInputs:	1, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeBool}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitTypeMismatch, DeterminismRule: "pure, always deterministic",
	},
	OpISACallFrame: {
		Mnemonic:	"call_frame", Opcode: OpISACallFrame,
		StackInputs:	1, StackOutputs: 0,
		InputTypes:	[]StackValueType{StackTypeInt256}, OutputTypes: nil,
		BaseGasCost:	100, OverflowBehavior: "none",
		FailureExitCode:	ExitInvalidJump, DeterminismRule: "pushes continuation frame, deterministic",
	},
	OpISAReturnFrame: {
		Mnemonic:	"return_frame", Opcode: OpISAReturnFrame,
		StackInputs:	0, StackOutputs: 0,
		InputTypes:	nil, OutputTypes: nil,
		BaseGasCost:	50, OverflowBehavior: "none",
		FailureExitCode:	ExitStackUnderflow, DeterminismRule: "restores continuation slot, deterministic",
	},
	OpISAJumpCond: {
		Mnemonic:	"jump_cond", Opcode: OpISAJumpCond,
		StackInputs:	2, StackOutputs: 0,
		InputTypes:	[]StackValueType{StackTypeBool, StackTypeInt256}, OutputTypes: nil,
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitInvalidJump, DeterminismRule: "target MUST be CFG block entry",
	},
	OpISARaiseError: {
		Mnemonic:	"raise_error", Opcode: OpISARaiseError,
		StackInputs:	1, StackOutputs: 0,
		InputTypes:	[]StackValueType{StackTypeInt256}, OutputTypes: nil,
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitContractAbort, DeterminismRule: "terminates execution frame, deterministic",
	},
	OpISALoadState: {
		Mnemonic:	"load_state_chunk", Opcode: OpISALoadState,
		StackInputs:	1, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef}, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	20, StackGasCost: 0, ChunkGasCost: 10,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitChunkError, DeterminismRule: "reads immutable snapshot, deterministic",
	},
	OpISAStoreState: {
		Mnemonic:	"store_state_chunk", Opcode: OpISAStoreState,
		StackInputs:	1, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef}, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	50, StackGasCost: 0, ChunkGasCost: 25,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitStateMutation, DeterminismRule: "creates new root, does not mutate in-place",
	},
	OpISAHashChunk: {
		Mnemonic:	"hash_chunk", Opcode: OpISAHashChunk,
		StackInputs:	1, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef}, OutputTypes: []StackValueType{StackTypeHash},
		BaseGasCost:	80, ChunkGasCost: 10,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitChunkError, DeterminismRule: "BLAKE3 hash, deterministic",
	},
	OpISAHashData: {
		Mnemonic:	"hash_data", Opcode: OpISAHashData,
		StackInputs:	1, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeBytes}, OutputTypes: []StackValueType{StackTypeHash},
		BaseGasCost:	80, StackGasCost: 1,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "SHA256 hash, deterministic",
	},
	OpISAEmitAction: {
		Mnemonic:	"emit_action", Opcode: OpISAEmitAction,
		StackInputs:	2, StackOutputs: 0,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeChunkRef}, OutputTypes: nil,
		BaseGasCost:	100, OverflowBehavior: "none",
		FailureExitCode:	ExitActionBudget, DeterminismRule: "accumulates in ActionQueue, not immediate",
	},
	OpISAQueueMessage: {
		Mnemonic:	"queue_message", Opcode: OpISAQueueMessage,
		StackInputs:	3, StackOutputs: 0,
		InputTypes:	[]StackValueType{StackTypeAddress, StackTypeChunkRef, StackTypeCoins}, OutputTypes: nil,
		BaseGasCost:	250, OverflowBehavior: "none",
		FailureExitCode:	ExitActionBudget, DeterminismRule: "accumulates in ActionQueue, not immediate",
	},
	OpISAGetCaller: {
		Mnemonic:	"get_caller", Opcode: OpISAGetCaller,
		StackInputs:	0, StackOutputs: 1,
		InputTypes:	nil, OutputTypes: []StackValueType{StackTypeAddress},
		BaseGasCost:	10, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "reads from immutable ExecutionContextChunk",
	},
	OpISAChunkMapGet: {
		Mnemonic:	"chunkmap_get", Opcode: OpISAChunkMapGet,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef, StackTypeBytes}, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	20, ChunkGasCost: 10,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitChunkError, DeterminismRule: "reads from immutable snapshot, deterministic",
	},
	OpISAChunkMapPut: {
		Mnemonic:	"chunkmap_put", Opcode: OpISAChunkMapPut,
		StackInputs:	3, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef, StackTypeBytes, StackTypeChunkRef}, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	60, ChunkGasCost: 25,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitStateMutation, DeterminismRule: "returns new root, does not mutate in-place",
	},
	OpISAChunkMapDelete: {
		Mnemonic:	"chunkmap_delete", Opcode: OpISAChunkMapDelete,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef, StackTypeBytes}, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	150, ChunkGasCost: 15,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitStateMutation, DeterminismRule: "returns new root, does not mutate in-place",
	},
	OpISAChunkMapProof: {
		Mnemonic:	"chunkmap_proof", Opcode: OpISAChunkMapProof,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef, StackTypeBytes}, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	200, ChunkGasCost: 30,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitChunkError, DeterminismRule: "generates Merkle proof, deterministic",
	},
	OpISAOver: {
		Mnemonic:	"over", Opcode: OpISAOver,
		StackInputs:	2, StackOutputs: 3,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeInt256, StackTypeInt256, StackTypeInt256},
		BaseGasCost:	1, OverflowBehavior: "none",
		FailureExitCode:	ExitStackUnderflow, DeterminismRule: "pure, always deterministic",
	},
	OpISANeq: {
		Mnemonic:	"neq", Opcode: OpISANeq,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pure, always deterministic",
	},
	OpISALt: {
		Mnemonic:	"lt", Opcode: OpISALt,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pure, always deterministic",
	},
	OpISALte: {
		Mnemonic:	"lte", Opcode: OpISALte,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pure, always deterministic",
	},
	OpISAGt: {
		Mnemonic:	"gt", Opcode: OpISAGt,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pure, always deterministic",
	},
	OpISAGte: {
		Mnemonic:	"gte", Opcode: OpISAGte,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256, StackTypeInt256}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pure, always deterministic",
	},
	OpISAJumpUncond: {
		Mnemonic:	"jump", Opcode: OpISAJumpUncond,
		StackInputs:	1, StackOutputs: 0,
		InputTypes:	[]StackValueType{StackTypeInt256}, OutputTypes: nil,
		BaseGasCost:	2, OverflowBehavior: "none",
		FailureExitCode:	ExitInvalidJump, DeterminismRule: "target MUST be CFG block entry",
	},
	OpISATryBegin: {
		Mnemonic:	"try_begin", Opcode: OpISATryBegin,
		StackInputs:	0, StackOutputs: 0,
		InputTypes:	nil, OutputTypes: nil,
		BaseGasCost:	10, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pushes error handler, deterministic",
	},
	OpISATryEnd: {
		Mnemonic:	"try_end", Opcode: OpISATryEnd,
		StackInputs:	0, StackOutputs: 0,
		InputTypes:	nil, OutputTypes: nil,
		BaseGasCost:	5, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "pops error handler, deterministic",
	},
	OpISACloneState: {
		Mnemonic:	"clone_state", Opcode: OpISACloneState,
		StackInputs:	0, StackOutputs: 1,
		InputTypes:	nil, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	30, ChunkGasCost: 15,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitChunkError, DeterminismRule: "creates copy of state root, deterministic",
	},
	OpISAMergeState: {
		Mnemonic:	"merge_state", Opcode: OpISAMergeState,
		StackInputs:	2, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeChunkRef, StackTypeChunkRef}, OutputTypes: []StackValueType{StackTypeChunkRef},
		BaseGasCost:	50, ChunkGasCost: 25,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitChunkError, DeterminismRule: "merges two state branches, deterministic",
	},
	OpISAFlushActions: {
		Mnemonic:	"flush_actions", Opcode: OpISAFlushActions,
		StackInputs:	0, StackOutputs: 0,
		InputTypes:	nil, OutputTypes: nil,
		BaseGasCost:	200, OverflowBehavior: "none",
		FailureExitCode:	ExitActionBudget, DeterminismRule: "finalizes and sorts actions, deterministic",
	},
	OpISAGetOrigin: {
		Mnemonic:	"get_origin", Opcode: OpISAGetOrigin,
		StackInputs:	0, StackOutputs: 1,
		InputTypes:	nil, OutputTypes: []StackValueType{StackTypeAddress},
		BaseGasCost:	10, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "reads from immutable ExecutionContextChunk",
	},
	OpISAGetValue: {
		Mnemonic:	"get_value", Opcode: OpISAGetValue,
		StackInputs:	0, StackOutputs: 1,
		InputTypes:	nil, OutputTypes: []StackValueType{StackTypeCoins},
		BaseGasCost:	10, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "reads from immutable ExecutionContextChunk",
	},
	OpISAGetBlockHeight: {
		Mnemonic:	"get_block_height", Opcode: OpISAGetBlockHeight,
		StackInputs:	0, StackOutputs: 1,
		InputTypes:	nil, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	10, OverflowBehavior: "none",
		FailureExitCode:	ExitSuccess, DeterminismRule: "reads from immutable ExecutionContextChunk",
	},
	OpISAVerifySig: {
		Mnemonic:	"verify_sig", Opcode: OpISAVerifySig,
		StackInputs:	3, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeBytes, StackTypeBytes, StackTypeBytes}, OutputTypes: []StackValueType{StackTypeBool},
		BaseGasCost:	5000, OverflowBehavior: "none",
		FailureExitCode:	ExitChunkError, DeterminismRule: "Ed25519 signature verify, deterministic",
	},
	OpISAEncode: {
		Mnemonic:	"encode", Opcode: OpISAEncode,
		StackInputs:	1, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeInt256}, OutputTypes: []StackValueType{StackTypeBytes},
		BaseGasCost:	5, StackGasCost: 1,
		OverflowBehavior:	"none",
		FailureExitCode:	ExitTypeMismatch, DeterminismRule: "Codec<T>-verified, deterministic",
	},
	OpISADecode: {
		Mnemonic:	"decode", Opcode: OpISADecode,
		StackInputs:	1, StackOutputs: 1,
		InputTypes:	[]StackValueType{StackTypeBytes}, OutputTypes: []StackValueType{StackTypeInt256},
		BaseGasCost:	5, StackGasCost: 1,
		OverflowBehavior:	"trap on invalid decode",
		FailureExitCode:	ExitInvalidDecode, DeterminismRule: "Codec<T>-verified, invalid decode → trap",
	},
}

// ISALookup returns the specification for a given opcode.
func ISALookup(op ISAOpcode) (ISAInstructionSpec, bool) {
	spec, ok := ISAOpcodeTable[op]
	return spec, ok
}

// ISAStackEffect returns the net stack effect (output - input) for an opcode.
func ISAStackEffect(op ISAOpcode) (int, bool) {
	spec, ok := ISAOpcodeTable[op]
	if !ok {
		return 0, false
	}
	return spec.StackOutputs - spec.StackInputs, true
}

// ISAVerifyStackContract checks that the stack has enough values for an opcode.
func ISAVerifyStackContract(op ISAOpcode, stackDepth int) bool {
	spec, ok := ISAOpcodeTable[op]
	if !ok {
		return false
	}
	return stackDepth >= spec.StackInputs
}

// ISAGasCost computes the total gas cost for an opcode execution.
func ISAGasCost(op ISAOpcode) uint64 {
	spec, ok := ISAOpcodeTable[op]
	if !ok {
		return 1000
	}
	return spec.BaseGasCost + spec.StackGasCost + spec.ChunkGasCost
}

// ExecuteKernelSemantics enforces the formal execution rule:
//
//	(StateRootChunk, Message) → (NewStateRootChunk, ActionQueueChunk, ExitCode, Receipt)
//
// Invariants:
//   - Deterministic: same input → same output
//   - No state mutation on error
//   - All failures produce deterministic receipt
//   - Only SUCCESS reaches commit phase
func ExecuteKernelSemantics(frame *KernelExecutionFrame) (*StateRootChunk, *ActionQueueChunk, StructuredExitCode, AVMReceipt, error) {

	frame.Phase = PhaseStorage
	if !frame.ChargeGas(500) {
		return nil, nil, ExitGasExhausted, AVMReceipt{}, nil
	}

	frame.Phase = PhaseCredit
	if !frame.ChargeGas(100) {
		return nil, nil, ExitGasExhausted, AVMReceipt{}, nil
	}
	if frame.Message.Value > 0 && frame.WorkingState != nil {
		credited, err := applyAttachedValueToWorkingState(frame.WorkingState, frame.Message.Value)
		if err != nil {
			frame.Aborted = true
			frame.ErrorState = ExitStateCorruption
			return nil, nil, ExitStateCorruption, AVMReceipt{}, nil
		}
		frame.WorkingState = credited
	}

	frame.Phase = PhaseCompute
	if !frame.ChargeGas(1000) {
		return nil, nil, ExitGasExhausted, AVMReceipt{}, nil
	}

	frame.Phase = PhaseAction
	if !frame.ChargeGas(200) {
		return nil, nil, ExitGasExhausted, AVMReceipt{}, nil
	}

	if uint32(len(frame.ActionQueue.Actions)) > frame.ActionBudget {
		frame.Aborted = true
		return nil, nil, ExitActionBudget, AVMReceipt{}, nil
	}

	frame.Phase = PhaseFinalization
	if !frame.ChargeGas(300) {
		return nil, nil, ExitGasExhausted, AVMReceipt{}, nil
	}

	workingState := frame.WorkingState
	if workingState == nil {
		workingState = frame.StateSnapshot
	}
	if workingState == nil {
		workingState = chunk.NewEmptyMap().Root()
	}

	newRoot := NewStateRootChunk(workingState)
	frame.ActionQueue.Finalize()

	exitCode := ExitSuccess
	if frame.Aborted {
		exitCode = frame.ErrorState
	}

	var stateRootBefore, stateRootAfter string
	if frame.StateSnapshot != nil {
		stateRootBefore = string(frame.StateSnapshot.Hash())
	}
	if workingState != nil {
		stateRootAfter = string(workingState.Hash())
	}

	receipt := AVMReceipt{
		ExitCode:		exitCode.ToUint32(),
		GasUsed:		frame.GasUsed,
		GasLimit:		frame.GasLimit,
		PhaseGas:		frame.PhaseGas,
		StateRootBefore:	stateRootBefore,
		StateRootAfter:		stateRootAfter,
		EmittedActionsHash:	string(frame.ActionQueue.Hash),
		ExecutionTraceHash:	string(frame.finalizeTrace()),
	}

	return newRoot, frame.ActionQueue, exitCode, receipt, nil
}

// ChargeGas adds gas to the frame's total used.
func (f *KernelExecutionFrame) ChargeGas(amount uint64) bool {
	if f.GasUsed+amount > f.GasLimit {
		f.GasUsed = f.GasLimit
		f.PhaseGas[f.Phase] += f.GasLimit - f.GasUsed
		f.Aborted = true
		f.ErrorState = ExitGasExhausted
		return false
	}
	f.GasUsed += amount
	f.PhaseGas[f.Phase] += amount
	return true
}

// PushValue pushes a typed value onto the evaluation stack.
func (f *KernelExecutionFrame) PushValue(val StackValue) error {
	if len(f.Stack) >= 1024 {
		f.ErrorState = ExitStackOverflow
		f.Aborted = true
		return fmt.Errorf("AVM kernel stack overflow: depth %d", len(f.Stack))
	}
	f.Stack = append(f.Stack, val)
	f.Trace.Steps = append(f.Trace.Steps, KernelTraceStep{
		Opcode:		OpISAPush,
		StackDepth:	len(f.Stack),
		GasUsed:	2,
		Phase:		f.Phase,
		IP:		f.IP,
	})
	return nil
}

// PopValue pops a typed value from the evaluation stack.
func (f *KernelExecutionFrame) PopValue() (StackValue, error) {
	if len(f.Stack) == 0 {
		f.ErrorState = ExitStackUnderflow
		f.Aborted = true
		return StackValue{}, fmt.Errorf("AVM kernel stack underflow")
	}
	val := f.Stack[len(f.Stack)-1]
	f.Stack = f.Stack[:len(f.Stack)-1]
	return val, nil
}

// PopValueOfType pops a value and checks its type.
func (f *KernelExecutionFrame) PopValueOfType(expectedType StackValueType) (StackValue, error) {
	val, err := f.PopValue()
	if err != nil {
		return StackValue{}, err
	}
	if val.Type != expectedType && val.Type != StackTypeNull {
		f.ErrorState = ExitTypeMismatch
		f.Aborted = true
		return StackValue{}, fmt.Errorf("AVM kernel type mismatch: expected %s, got %s", expectedType, val.Type)
	}
	return val, nil
}

// SetContinuation sets a continuation slot pointer.
func (f *KernelExecutionFrame) SetContinuation(slot ExecutionSlot, ptr uint32) {
	switch slot {
	case SlotReturn:
		f.Continuation.ReturnPtr = ptr
	case SlotAltReturn:
		f.Continuation.AltReturnPtr = ptr
	case SlotError:
		f.Continuation.ErrorHandlerPtr = ptr
	case SlotDispatch:
		f.Continuation.DispatcherPtr = ptr
	}
}

// GetContinuation reads a continuation slot pointer.
func (f *KernelExecutionFrame) GetContinuation(slot ExecutionSlot) uint32 {
	switch slot {
	case SlotReturn:
		return f.Continuation.ReturnPtr
	case SlotAltReturn:
		return f.Continuation.AltReturnPtr
	case SlotError:
		return f.Continuation.ErrorHandlerPtr
	case SlotDispatch:
		return f.Continuation.DispatcherPtr
	default:
		return 0
	}
}

func (f *KernelExecutionFrame) finalizeTrace() []byte {
	h := sha256.New()
	for _, step := range f.Trace.Steps {
		h.Write([]byte{byte(step.Opcode >> 8), byte(step.Opcode)})
		var gas [8]byte
		binary.BigEndian.PutUint64(gas[:], step.GasUsed)
		h.Write(gas[:])
		h.Write([]byte{byte(step.Phase)})
		var ip [4]byte
		binary.BigEndian.PutUint32(ip[:], step.IP)
		h.Write(ip[:])
	}
	return h.Sum(nil)
}

func appendUint64BigEndian(buf []byte, v uint64) []byte {
	return append(buf,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v),
	)
}

func appendUint32BigEndian(buf []byte, v uint32) []byte {
	return append(buf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func appendString(buf []byte, s string) []byte {
	buf = appendUint32BigEndian(buf, uint32(len(s)))
	buf = append(buf, s...)
	return buf
}
