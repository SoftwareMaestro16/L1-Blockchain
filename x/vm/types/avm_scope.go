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
	AVMVMVersion			uint64	= 2
	AVMDefaultInstructionSet	uint64	= 1
	MaxAVMTokenLength			= 128
	MaxAVMPayloadTypeLength			= MaxAsyncMessagePayloadType

	AVMOpNoop		AVMOpcode	= "NOOP"
	AVMOpPush		AVMOpcode	= "PUSH"
	AVMOpPop		AVMOpcode	= "POP"
	AVMOpDup		AVMOpcode	= "DUP"
	AVMOpSwap		AVMOpcode	= "SWAP"
	AVMOpLoadLocal		AVMOpcode	= "LOAD_LOCAL"
	AVMOpStoreLocal		AVMOpcode	= "STORE_LOCAL"
	AVMOpAdd		AVMOpcode	= "ADD"
	AVMOpSub		AVMOpcode	= "SUB"
	AVMOpMul		AVMOpcode	= "MUL"
	AVMOpDiv		AVMOpcode	= "DIV"
	AVMOpMod		AVMOpcode	= "MOD"
	AVMOpNeg		AVMOpcode	= "NEG"
	AVMOpCmp		AVMOpcode	= "CMP"
	AVMOpJmp		AVMOpcode	= "JMP"
	AVMOpJmpIf		AVMOpcode	= "JMP_IF"
	AVMOpCallInternal	AVMOpcode	= "CALL_INTERNAL"
	AVMOpRet		AVMOpcode	= "RET"
	AVMOpAbort		AVMOpcode	= "ABORT"
	AVMOpMemLoad		AVMOpcode	= "MEM_LOAD"
	AVMOpMemStore		AVMOpcode	= "MEM_STORE"
	AVMOpMemCopy		AVMOpcode	= "MEM_COPY"
	AVMOpMemSize		AVMOpcode	= "MEM_SIZE"
	AVMOpMemGrow		AVMOpcode	= "MEM_GROW"
	AVMOpKVGet		AVMOpcode	= "KV_GET"
	AVMOpKVSet		AVMOpcode	= "KV_SET"
	AVMOpKVDelete		AVMOpcode	= "KV_DELETE"
	AVMOpKVExists		AVMOpcode	= "KV_EXISTS"
	AVMOpKVRangeBounded	AVMOpcode	= "KV_RANGE_BOUNDED"
	AVMOpHash		AVMOpcode	= "HASH"
	AVMOpVerifySig		AVMOpcode	= "VERIFY_SIG"
	AVMOpVerifyMerkleProof	AVMOpcode	= "VERIFY_MERKLE_PROOF"
	AVMOpVerifyMessageProof	AVMOpcode	= "VERIFY_MESSAGE_PROOF"
	AVMOpVerifyZoneRoot	AVMOpcode	= "VERIFY_ZONE_ROOT"
	AVMOpMsgNew		AVMOpcode	= "MSG_NEW"
	AVMOpMsgSetValue	AVMOpcode	= "MSG_SET_VALUE"
	AVMOpMsgSetPayload	AVMOpcode	= "MSG_SET_PAYLOAD"
	AVMOpMsgSetGas		AVMOpcode	= "MSG_SET_GAS"
	AVMOpMsgSetExpiry	AVMOpcode	= "MSG_SET_EXPIRY"
	AVMOpMsgSend		AVMOpcode	= "MSG_SEND"
	AVMOpMsgBounce		AVMOpcode	= "MSG_BOUNCE"
	AVMOpPromiseNew		AVMOpcode	= "PROMISE_NEW"
	AVMOpPromiseAwait	AVMOpcode	= "PROMISE_AWAIT"
	AVMOpPromiseResolve	AVMOpcode	= "PROMISE_RESOLVE"
	AVMOpPromiseReject	AVMOpcode	= "PROMISE_REJECT"
	AVMOpPromiseTimeout	AVMOpcode	= "PROMISE_TIMEOUT"
	AVMOpABIExport		AVMOpcode	= "ABI_EXPORT"
	AVMOpABIMethod		AVMOpcode	= "ABI_METHOD"
	AVMOpABIEvent		AVMOpcode	= "ABI_EVENT"
	AVMOpABIRequire		AVMOpcode	= "ABI_REQUIRE"
	AVMOpEmitEvent		AVMOpcode	= "EMIT_EVENT"
	AVMOpCtxHeight		AVMOpcode	= "CTX_HEIGHT"
	AVMOpCtxChainID		AVMOpcode	= "CTX_CHAIN_ID"
	AVMOpCtxZoneID		AVMOpcode	= "CTX_ZONE_ID"
	AVMOpCtxShardID		AVMOpcode	= "CTX_SHARD_ID"
	AVMOpCtxCaller		AVMOpcode	= "CTX_CALLER"
	AVMOpCtxContract	AVMOpcode	= "CTX_CONTRACT"
	AVMOpCtxValue		AVMOpcode	= "CTX_VALUE"
	AVMOpCtxGasLeft		AVMOpcode	= "CTX_GAS_LEFT"

	AVMOpExternalNetwork		AVMOpcode		= "SYSCALL_EXTERNAL_NETWORK"
	AVMOpWallClock			AVMOpcode		= "SYSCALL_WALL_CLOCK"
	AVMOpNonDeterministic		AVMOpcode		= "SYSCALL_NON_DETERMINISTIC"
	AVMOpKVRangeUnbounded		AVMOpcode		= "KV_RANGE_UNBOUNDED"
	AVMOpDirectRemoteMutation	AVMOpcode		= "REMOTE_ZONE_MUTATE"
	AVMOpUnboundedRecursion		AVMOpcode		= "CALL_RECURSIVE"
	AVMPromisePending		AVMPromiseStatus	= "pending"
	AVMPromiseResolved		AVMPromiseStatus	= "resolved"
	AVMPromiseRejected		AVMPromiseStatus	= "rejected"
	AVMPromiseTimedOut		AVMPromiseStatus	= "timed_out"
	AVMPromiseRefunded		AVMPromiseStatus	= "refunded"
	AVMProofRootZone		AVMProofRootType	= "zone"
	AVMProofRootShard		AVMProofRootType	= "shard"
	AVMProofRootMessage		AVMProofRootType	= "message"
	AVMProofRootReceipt		AVMProofRootType	= "receipt"
	AVMProofRootContractState	AVMProofRootType	= "contract_state"
	AVMProofRootResolverRecord	AVMProofRootType	= "resolver_record"
)

type AVMOpcode string
type AVMInstructionCategory string
type AVMPromiseStatus string
type AVMProofRootType string

const (
	AVMCategoryCoreStack	AVMInstructionCategory	= "core_stack"
	AVMCategoryArithmetic	AVMInstructionCategory	= "arithmetic"
	AVMCategoryControlFlow	AVMInstructionCategory	= "control_flow"
	AVMCategoryMemory	AVMInstructionCategory	= "memory"
	AVMCategoryStorage	AVMInstructionCategory	= "storage"
	AVMCategoryCryptoProof	AVMInstructionCategory	= "crypto_and_proofs"
	AVMCategoryMessages	AVMInstructionCategory	= "messages"
	AVMCategoryPromises	AVMInstructionCategory	= "promises"
	AVMCategoryABI		AVMInstructionCategory	= "abi"
	AVMCategoryContext	AVMInstructionCategory	= "context"
)

type AVMOpcodeDescriptor struct {
	Opcode		AVMOpcode
	Category	AVMInstructionCategory
	Purpose		string
	GasCost		uint64
}

type AVMInstructionSet struct {
	Version	uint64
	Opcodes	[]AVMOpcodeDescriptor
	SetHash	string
}

type AVMOpcodeGasEntry struct {
	Opcode	AVMOpcode
	Gas	uint64
}

type AVMLimits struct {
	MaxInstructions		uint32
	MaxStackDepth		uint32
	MaxMemoryBytes		uint64
	MaxStorageKeyBytes	uint32
	MaxStorageValueBytes	uint64
	MaxOutputMessages	uint32
	MaxEvents		uint32
	MaxProofBytes		uint32
	MaxRecursionDepth	uint32
	MaxBoundedRangeItems	uint32
	MaxInstructionGas	uint64
	MaxABIEntries		uint32
	MaxEventPayloadBytes	uint32
	MaxPromiseStateWrites	uint32
}

type AVMGasTable struct {
	BaseInstructionGas	uint64
	OpcodeCosts		[]AVMOpcodeGasEntry
	MemoryByteGas		uint64
	StorageReadGas		uint64
	StorageWriteGas		uint64
	StorageReadByteGas	uint64
	StorageWriteByteGas	uint64
	ProofByteGas		uint64
	ProofDepthGas		uint64
	MessageCreateGas	uint64
	MessageEnvelopeGas	uint64
	MessageByteGas		uint64
	ForwardingFeeReserveGas	uint64
	EventByteGas		uint64
	ABIExportGas		uint64
	PromiseGas		uint64
	TableHash		string
}

type AVMExecutionContext struct {
	ChainID		string
	Height		uint64
	ZoneID		zonestypes.ZoneID
	ShardID		uint32
	ContractAddress	string
	Caller		string
	ValueNAET	uint64
	GasLimit	uint64
	ReadOnly	bool
	ContextHash	string
}

type AVMInstruction struct {
	Opcode		AVMOpcode
	Key		string
	Value		[]byte
	Message		AVMAsyncMessage
	Proof		AVMProofInput
	Promise		AVMPromiseState
	ABI		AVMABIDescriptor
	Event		AVMEvent
	MemoryGrow	uint32
	RangeLimit	uint32
	GasOverride	uint64
}

type AVMProgram struct {
	VMVersion		uint64
	InstructionSetVersion	uint64
	Instructions		[]AVMInstruction
	MaxRecursionDepth	uint32
	ProgramHash		string
}

type AVMStorageRead struct {
	Key	string
	KeyHash	string
}

type AVMStorageWrite struct {
	Key		string
	ValueHash	string
	Deleted		bool
}

type AVMProofInput struct {
	ProofVersion	uint64
	ChainID		string
	Height		uint64
	RootType	AVMProofRootType
	RootHash	string
	Key		string
	ValueHash	string
	ProofBytes	[]byte
	ProofHash	string
}

type AVMPromiseState struct {
	PromiseID	string
	Contract	string
	MessageID	string
	Status		AVMPromiseStatus
	CreatedHeight	uint64
	ExpiryHeight	uint64
	ReceiptHash	string
	ReturnHash	string
	PromiseHash	string
}

type AVMABIDescriptor struct {
	ABIVersion	uint64
	CodeID		uint64
	Methods		[]string
	Events		[]string
	Errors		[]string
	RequiredFunds	[]string
	GasHints	[]string
	InterfaceHash	string
}

type AVMEvent struct {
	Height		uint64
	ContractAddress	string
	EventID		string
	Name		string
	PayloadHash	string
	EventHash	string
}

type AVMExecutionResult struct {
	GasUsed			uint64
	MemoryBytes		uint64
	Stack			[]string
	StorageReads		[]AVMStorageRead
	StorageWrites		[]AVMStorageWrite
	OutputMessages		[]AVMAsyncMessage
	ProofsVerified		[]AVMProofInput
	Promises		[]AVMPromiseState
	ABIDescriptors		[]AVMABIDescriptor
	Events			[]AVMEvent
	StorageRoot		string
	MessageRoot		string
	PromiseRoot		string
	ABIRoot			string
	EventRoot		string
	ExecutionHash		string
	ReadOnlySimulation	bool
}

func DefaultAVMLimits() AVMLimits {
	return AVMLimits{
		MaxInstructions:	4096,
		MaxStackDepth:		1024,
		MaxMemoryBytes:		1024 * 1024,
		MaxStorageKeyBytes:	256,
		MaxStorageValueBytes:	256 * 1024,
		MaxOutputMessages:	128,
		MaxEvents:		1024,
		MaxProofBytes:		64 * 1024,
		MaxRecursionDepth:	32,
		MaxBoundedRangeItems:	1024,
		MaxInstructionGas:	1_000_000,
		MaxABIEntries:		256,
		MaxEventPayloadBytes:	64 * 1024,
		MaxPromiseStateWrites:	256,
	}
}

func DefaultAVMInstructionSet() (AVMInstructionSet, error) {
	opcodes := []AVMOpcodeDescriptor{
		{Opcode: AVMOpPush, Category: AVMCategoryCoreStack, Purpose: "push canonical value onto stack", GasCost: 1},
		{Opcode: AVMOpPop, Category: AVMCategoryCoreStack, Purpose: "pop stack value", GasCost: 1},
		{Opcode: AVMOpDup, Category: AVMCategoryCoreStack, Purpose: "duplicate stack value", GasCost: 1},
		{Opcode: AVMOpSwap, Category: AVMCategoryCoreStack, Purpose: "swap top stack values", GasCost: 1},
		{Opcode: AVMOpLoadLocal, Category: AVMCategoryCoreStack, Purpose: "load bounded local value", GasCost: 2},
		{Opcode: AVMOpStoreLocal, Category: AVMCategoryCoreStack, Purpose: "store bounded local value", GasCost: 2},
		{Opcode: AVMOpAdd, Category: AVMCategoryArithmetic, Purpose: "bounded integer addition", GasCost: 2},
		{Opcode: AVMOpSub, Category: AVMCategoryArithmetic, Purpose: "bounded integer subtraction", GasCost: 2},
		{Opcode: AVMOpMul, Category: AVMCategoryArithmetic, Purpose: "bounded integer multiplication", GasCost: 3},
		{Opcode: AVMOpDiv, Category: AVMCategoryArithmetic, Purpose: "bounded integer division", GasCost: 3},
		{Opcode: AVMOpMod, Category: AVMCategoryArithmetic, Purpose: "bounded integer modulo", GasCost: 3},
		{Opcode: AVMOpNeg, Category: AVMCategoryArithmetic, Purpose: "bounded integer negation", GasCost: 2},
		{Opcode: AVMOpCmp, Category: AVMCategoryArithmetic, Purpose: "bounded integer comparison", GasCost: 2},
		{Opcode: AVMOpJmp, Category: AVMCategoryControlFlow, Purpose: "deterministic jump target", GasCost: 2},
		{Opcode: AVMOpJmpIf, Category: AVMCategoryControlFlow, Purpose: "deterministic conditional jump target", GasCost: 2},
		{Opcode: AVMOpCallInternal, Category: AVMCategoryControlFlow, Purpose: "bounded internal call", GasCost: 5},
		{Opcode: AVMOpRet, Category: AVMCategoryControlFlow, Purpose: "return from internal call", GasCost: 2},
		{Opcode: AVMOpAbort, Category: AVMCategoryControlFlow, Purpose: "abort execution with consumed gas", GasCost: 2},
		{Opcode: AVMOpMemLoad, Category: AVMCategoryMemory, Purpose: "bounded memory load", GasCost: 2},
		{Opcode: AVMOpMemStore, Category: AVMCategoryMemory, Purpose: "bounded memory store", GasCost: 2},
		{Opcode: AVMOpMemCopy, Category: AVMCategoryMemory, Purpose: "bounded memory copy", GasCost: 3},
		{Opcode: AVMOpMemSize, Category: AVMCategoryMemory, Purpose: "read bounded memory size", GasCost: 1},
		{Opcode: AVMOpMemGrow, Category: AVMCategoryMemory, Purpose: "grow bounded memory", GasCost: 2},
		{Opcode: AVMOpKVGet, Category: AVMCategoryStorage, Purpose: "Store v2 contract-local get", GasCost: 4},
		{Opcode: AVMOpKVSet, Category: AVMCategoryStorage, Purpose: "Store v2 contract-local set", GasCost: 8},
		{Opcode: AVMOpKVDelete, Category: AVMCategoryStorage, Purpose: "Store v2 contract-local delete", GasCost: 8},
		{Opcode: AVMOpKVExists, Category: AVMCategoryStorage, Purpose: "Store v2 contract-local exists", GasCost: 4},
		{Opcode: AVMOpKVRangeBounded, Category: AVMCategoryStorage, Purpose: "Store v2 bounded range scan", GasCost: 8},
		{Opcode: AVMOpHash, Category: AVMCategoryCryptoProof, Purpose: "deterministic hashing", GasCost: 4},
		{Opcode: AVMOpVerifySig, Category: AVMCategoryCryptoProof, Purpose: "deterministic signature proof validation", GasCost: 20},
		{Opcode: AVMOpVerifyMerkleProof, Category: AVMCategoryCryptoProof, Purpose: "Merkle proof validation", GasCost: 20},
		{Opcode: AVMOpVerifyMessageProof, Category: AVMCategoryCryptoProof, Purpose: "message inclusion proof validation", GasCost: 20},
		{Opcode: AVMOpVerifyZoneRoot, Category: AVMCategoryCryptoProof, Purpose: "zone root proof validation", GasCost: 20},
		{Opcode: AVMOpMsgNew, Category: AVMCategoryMessages, Purpose: "construct message envelope", GasCost: 10},
		{Opcode: AVMOpMsgSetValue, Category: AVMCategoryMessages, Purpose: "set message value", GasCost: 3},
		{Opcode: AVMOpMsgSetPayload, Category: AVMCategoryMessages, Purpose: "set message payload", GasCost: 3},
		{Opcode: AVMOpMsgSetGas, Category: AVMCategoryMessages, Purpose: "set message gas limit", GasCost: 3},
		{Opcode: AVMOpMsgSetExpiry, Category: AVMCategoryMessages, Purpose: "set message expiry", GasCost: 3},
		{Opcode: AVMOpMsgSend, Category: AVMCategoryMessages, Purpose: "emit prepaid message", GasCost: 20},
		{Opcode: AVMOpMsgBounce, Category: AVMCategoryMessages, Purpose: "emit bounded bounce message", GasCost: 20},
		{Opcode: AVMOpPromiseNew, Category: AVMCategoryPromises, Purpose: "create pending promise", GasCost: 8},
		{Opcode: AVMOpPromiseAwait, Category: AVMCategoryPromises, Purpose: "register non-blocking promise await", GasCost: 8},
		{Opcode: AVMOpPromiseResolve, Category: AVMCategoryPromises, Purpose: "resolve promise from receipt", GasCost: 8},
		{Opcode: AVMOpPromiseReject, Category: AVMCategoryPromises, Purpose: "reject promise from failure receipt", GasCost: 8},
		{Opcode: AVMOpPromiseTimeout, Category: AVMCategoryPromises, Purpose: "time out promise deterministically", GasCost: 8},
		{Opcode: AVMOpABIExport, Category: AVMCategoryABI, Purpose: "export committed ABI descriptor", GasCost: 5},
		{Opcode: AVMOpABIMethod, Category: AVMCategoryABI, Purpose: "validate ABI method selector", GasCost: 3},
		{Opcode: AVMOpABIEvent, Category: AVMCategoryABI, Purpose: "validate ABI event descriptor", GasCost: 3},
		{Opcode: AVMOpABIRequire, Category: AVMCategoryABI, Purpose: "validate required ABI interface", GasCost: 3},
		{Opcode: AVMOpEmitEvent, Category: AVMCategoryABI, Purpose: "emit deterministic contract event", GasCost: 5},
		{Opcode: AVMOpCtxHeight, Category: AVMCategoryContext, Purpose: "read consensus height", GasCost: 1},
		{Opcode: AVMOpCtxChainID, Category: AVMCategoryContext, Purpose: "read chain id", GasCost: 1},
		{Opcode: AVMOpCtxZoneID, Category: AVMCategoryContext, Purpose: "read zone id", GasCost: 1},
		{Opcode: AVMOpCtxShardID, Category: AVMCategoryContext, Purpose: "read shard id", GasCost: 1},
		{Opcode: AVMOpCtxCaller, Category: AVMCategoryContext, Purpose: "read caller", GasCost: 1},
		{Opcode: AVMOpCtxContract, Category: AVMCategoryContext, Purpose: "read contract address", GasCost: 1},
		{Opcode: AVMOpCtxValue, Category: AVMCategoryContext, Purpose: "read call value", GasCost: 1},
		{Opcode: AVMOpCtxGasLeft, Category: AVMCategoryContext, Purpose: "read remaining gas after current charge", GasCost: 1},
		{Opcode: AVMOpNoop, Category: AVMCategoryCoreStack, Purpose: "deterministic no-op", GasCost: 1},
	}
	set := AVMInstructionSet{Version: AVMDefaultInstructionSet, Opcodes: opcodes}
	set = canonicalAVMInstructionSet(set)
	set.SetHash = ComputeAVMInstructionSetHash(set)
	return set, set.Validate()
}

func DefaultAVMGasTable() (AVMGasTable, error) {
	set, err := DefaultAVMInstructionSet()
	if err != nil {
		return AVMGasTable{}, err
	}
	table := AVMGasTable{
		BaseInstructionGas:		1,
		OpcodeCosts:			AVMOpcodeGasEntriesFromInstructionSet(set),
		MemoryByteGas:			1,
		StorageReadGas:			2,
		StorageWriteGas:		4,
		StorageReadByteGas:		1,
		StorageWriteByteGas:		2,
		ProofByteGas:			3,
		ProofDepthGas:			5,
		MessageCreateGas:		20,
		MessageEnvelopeGas:		8,
		MessageByteGas:			1,
		ForwardingFeeReserveGas:	1,
		EventByteGas:			1,
		ABIExportGas:			5,
		PromiseGas:			8,
	}
	table.TableHash = ComputeAVMGasTableHash(table)
	return table, table.Validate()
}

func NewAVMProgram(program AVMProgram, limits AVMLimits, gasTable AVMGasTable) (AVMProgram, error) {
	program = canonicalAVMProgram(program)
	program.ProgramHash = ComputeAVMProgramHash(program)
	return program, ValidateAVMProgram(program, limits, gasTable)
}

func ExecuteAVMProgram(program AVMProgram, ctx AVMExecutionContext, limits AVMLimits, gasTable AVMGasTable) (AVMExecutionResult, error) {
	ctx = canonicalAVMExecutionContext(ctx)
	if err := ctx.Validate(); err != nil {
		return AVMExecutionResult{}, err
	}
	if err := ValidateAVMProgram(program, limits, gasTable); err != nil {
		return AVMExecutionResult{}, err
	}
	var result AVMExecutionResult
	result.ReadOnlySimulation = ctx.ReadOnly
	for _, instruction := range canonicalAVMProgram(program).Instructions {
		gas, err := AVMInstructionGas(instruction, gasTable, limits)
		if err != nil {
			return AVMExecutionResult{}, err
		}
		result.GasUsed, err = checkedAVMGasAdd(result.GasUsed, gas)
		if err != nil {
			return AVMExecutionResult{}, err
		}
		if result.GasUsed > ctx.GasLimit {
			return AVMExecutionResult{}, errors.New("AVM 2.0 execution exhausted gas")
		}
		switch instruction.Opcode {
		case AVMOpNoop:
		case AVMOpPush:
			if len(result.Stack) >= int(limits.MaxStackDepth) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack depth exceeded")
			}
			result.Stack = append(result.Stack, hex.EncodeToString(instruction.Value))
		case AVMOpPop, AVMOpStoreLocal:
			if len(result.Stack) == 0 {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack underflow")
			}
			result.Stack = result.Stack[:len(result.Stack)-1]
		case AVMOpDup:
			if len(result.Stack) == 0 {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack underflow")
			}
			if len(result.Stack) >= int(limits.MaxStackDepth) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack depth exceeded")
			}
			result.Stack = append(result.Stack, result.Stack[len(result.Stack)-1])
		case AVMOpSwap:
			if len(result.Stack) < 2 {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack underflow")
			}
			result.Stack[len(result.Stack)-1], result.Stack[len(result.Stack)-2] = result.Stack[len(result.Stack)-2], result.Stack[len(result.Stack)-1]
		case AVMOpLoadLocal:
			if len(result.Stack) >= int(limits.MaxStackDepth) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack depth exceeded")
			}
			result.Stack = append(result.Stack, ComputeAVMBytesHash([]byte(instruction.Key)))
		case AVMOpAdd, AVMOpSub, AVMOpMul, AVMOpDiv, AVMOpMod, AVMOpCmp:
			if len(result.Stack) < 2 {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack underflow")
			}
			right := result.Stack[len(result.Stack)-1]
			left := result.Stack[len(result.Stack)-2]
			result.Stack = result.Stack[:len(result.Stack)-2]
			result.Stack = append(result.Stack, ComputeAVMBytesHash([]byte(string(instruction.Opcode)+":"+left+":"+right)))
		case AVMOpNeg:
			if len(result.Stack) == 0 {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack underflow")
			}
			result.Stack[len(result.Stack)-1] = ComputeAVMBytesHash([]byte("NEG:" + result.Stack[len(result.Stack)-1]))
		case AVMOpJmp, AVMOpJmpIf, AVMOpCallInternal, AVMOpRet:
			if instruction.Opcode == AVMOpJmpIf && len(result.Stack) == 0 {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack underflow")
			}
		case AVMOpAbort:
			return AVMExecutionResult{}, errors.New("AVM 2.0 execution aborted")
		case AVMOpMemGrow:
			result.MemoryBytes += uint64(instruction.MemoryGrow)
			if result.MemoryBytes > limits.MaxMemoryBytes {
				return AVMExecutionResult{}, errors.New("AVM 2.0 memory limit exceeded")
			}
		case AVMOpMemLoad, AVMOpMemStore, AVMOpMemCopy:
			if instruction.MemoryGrow > 0 {
				result.MemoryBytes += uint64(instruction.MemoryGrow)
				if result.MemoryBytes > limits.MaxMemoryBytes {
					return AVMExecutionResult{}, errors.New("AVM 2.0 memory limit exceeded")
				}
			}
		case AVMOpMemSize:
			if len(result.Stack) >= int(limits.MaxStackDepth) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack depth exceeded")
			}
			result.Stack = append(result.Stack, fmt.Sprintf("%020d", result.MemoryBytes))
		case AVMOpKVGet, AVMOpKVExists, AVMOpKVRangeBounded:
			if err := ValidateAVMStoreV2Key(ctx, instruction.Key, limits); err != nil {
				return AVMExecutionResult{}, err
			}
			result.StorageReads = append(result.StorageReads, AVMStorageRead{Key: instruction.Key, KeyHash: ComputeAVMBytesHash([]byte(instruction.Key))})
		case AVMOpKVSet:
			if ctx.ReadOnly {
				return AVMExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot write storage")
			}
			if err := ValidateAVMStoreV2Key(ctx, instruction.Key, limits); err != nil {
				return AVMExecutionResult{}, err
			}
			if uint64(len(instruction.Value)) > limits.MaxStorageValueBytes {
				return AVMExecutionResult{}, fmt.Errorf("AVM 2.0 storage value bytes must be <= %d", limits.MaxStorageValueBytes)
			}
			result.StorageWrites = append(result.StorageWrites, AVMStorageWrite{Key: instruction.Key, ValueHash: ComputeAVMBytesHash(instruction.Value)})
		case AVMOpKVDelete:
			if ctx.ReadOnly {
				return AVMExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot delete storage")
			}
			if err := ValidateAVMStoreV2Key(ctx, instruction.Key, limits); err != nil {
				return AVMExecutionResult{}, err
			}
			result.StorageWrites = append(result.StorageWrites, AVMStorageWrite{Key: instruction.Key, ValueHash: ComputeAVMBytesHash(nil), Deleted: true})
		case AVMOpHash:
			if len(result.Stack) >= int(limits.MaxStackDepth) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack depth exceeded")
			}
			result.Stack = append(result.Stack, ComputeAVMBytesHash(instruction.Value))
		case AVMOpVerifySig, AVMOpVerifyMerkleProof, AVMOpVerifyMessageProof, AVMOpVerifyZoneRoot:
			proof := canonicalAVMProofInput(instruction.Proof)
			if err := proof.Validate(limits); err != nil {
				return AVMExecutionResult{}, err
			}
			result.ProofsVerified = append(result.ProofsVerified, proof)
		case AVMOpMsgNew, AVMOpMsgSetValue, AVMOpMsgSetPayload, AVMOpMsgSetGas, AVMOpMsgSetExpiry:
			if err := canonicalAVMAsyncMessage(instruction.Message).Validate(); err != nil {
				return AVMExecutionResult{}, fmt.Errorf("AVM 2.0 message builder: %w", err)
			}
		case AVMOpMsgSend, AVMOpMsgBounce:
			if ctx.ReadOnly {
				return AVMExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot emit messages")
			}
			msg := canonicalAVMAsyncMessage(instruction.Message)
			if err := msg.Validate(); err != nil {
				return AVMExecutionResult{}, fmt.Errorf("AVM 2.0 output message: %w", err)
			}
			if len(result.OutputMessages) >= int(limits.MaxOutputMessages) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 output message limit exceeded")
			}
			result.OutputMessages = append(result.OutputMessages, msg)
		case AVMOpPromiseAwait:
			promise := canonicalAVMPromiseState(instruction.Promise)
			if err := promise.Validate(); err != nil {
				return AVMExecutionResult{}, err
			}
		case AVMOpPromiseNew, AVMOpPromiseResolve, AVMOpPromiseReject, AVMOpPromiseTimeout:
			if ctx.ReadOnly {
				return AVMExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot mutate promise state")
			}
			promise := canonicalAVMPromiseState(instruction.Promise)
			if err := promise.Validate(); err != nil {
				return AVMExecutionResult{}, err
			}
			if len(result.Promises) >= int(limits.MaxPromiseStateWrites) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 promise write limit exceeded")
			}
			result.Promises = append(result.Promises, promise)
		case AVMOpABIExport, AVMOpABIMethod, AVMOpABIEvent, AVMOpABIRequire:
			abi := canonicalAVMABIDescriptor(instruction.ABI)
			if err := abi.Validate(limits); err != nil {
				return AVMExecutionResult{}, err
			}
			if instruction.Opcode == AVMOpABIExport {
				result.ABIDescriptors = append(result.ABIDescriptors, abi)
			}
		case AVMOpEmitEvent:
			if ctx.ReadOnly {
				return AVMExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot emit events")
			}
			event := canonicalAVMEvent(instruction.Event)
			if err := event.Validate(limits); err != nil {
				return AVMExecutionResult{}, err
			}
			if len(result.Events) >= int(limits.MaxEvents) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 event limit exceeded")
			}
			result.Events = append(result.Events, event)
		case AVMOpCtxHeight, AVMOpCtxChainID, AVMOpCtxZoneID, AVMOpCtxShardID, AVMOpCtxCaller, AVMOpCtxContract, AVMOpCtxValue, AVMOpCtxGasLeft:
			if len(result.Stack) >= int(limits.MaxStackDepth) {
				return AVMExecutionResult{}, errors.New("AVM 2.0 stack depth exceeded")
			}
			result.Stack = append(result.Stack, AVMContextValue(instruction.Opcode, ctx, result.GasUsed))
		default:
			return AVMExecutionResult{}, fmt.Errorf("unsupported AVM 2.0 opcode %q", instruction.Opcode)
		}
	}
	result = canonicalAVMExecutionResult(result)
	result.StorageRoot = ComputeAVMStorageRoot(result.StorageReads, result.StorageWrites)
	result.MessageRoot = ComputeAVMMessageRoot(result.OutputMessages)
	result.PromiseRoot = ComputeAVMPromiseRoot(result.Promises)
	result.ABIRoot = ComputeAVMABIRoot(result.ABIDescriptors)
	result.EventRoot = ComputeAVMEventRoot(result.Events)
	result.ExecutionHash = ComputeAVMExecutionHash(result)
	return result, result.Validate()
}

func ValidateAVMProgram(program AVMProgram, limits AVMLimits, gasTable AVMGasTable) error {
	program = canonicalAVMProgram(program)
	if err := limits.Validate(); err != nil {
		return err
	}
	if err := gasTable.Validate(); err != nil {
		return err
	}
	if program.VMVersion != AVMVMVersion {
		return errors.New("AVM 2.0 program must use VM version 2")
	}
	if program.InstructionSetVersion == 0 {
		return errors.New("AVM 2.0 instruction set version must be positive")
	}
	if len(program.Instructions) == 0 {
		return errors.New("AVM 2.0 program must include instructions")
	}
	if len(program.Instructions) > int(limits.MaxInstructions) {
		return fmt.Errorf("AVM 2.0 instructions must be <= %d", limits.MaxInstructions)
	}
	if program.MaxRecursionDepth > limits.MaxRecursionDepth {
		return errors.New("AVM 2.0 recursion depth exceeds bounded limit")
	}
	for _, instruction := range program.Instructions {
		if err := instruction.Validate(limits, gasTable); err != nil {
			return err
		}
	}
	if program.ProgramHash == "" {
		return errors.New("AVM 2.0 program hash is required")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 program hash", program.ProgramHash); err != nil {
		return err
	}
	if program.ProgramHash != ComputeAVMProgramHash(program) {
		return errors.New("AVM 2.0 program hash mismatch")
	}
	return nil
}

func (l AVMLimits) Validate() error {
	if l.MaxInstructions == 0 || l.MaxStackDepth == 0 || l.MaxMemoryBytes == 0 {
		return errors.New("AVM 2.0 instruction, stack, and memory limits must be positive")
	}
	if l.MaxStorageKeyBytes == 0 || l.MaxStorageValueBytes == 0 {
		return errors.New("AVM 2.0 storage limits must be positive")
	}
	if l.MaxOutputMessages == 0 || l.MaxEvents == 0 || l.MaxProofBytes == 0 {
		return errors.New("AVM 2.0 message, event, and proof limits must be positive")
	}
	if l.MaxRecursionDepth == 0 || l.MaxBoundedRangeItems == 0 {
		return errors.New("AVM 2.0 recursion and range limits must be positive")
	}
	if l.MaxInstructionGas == 0 || l.MaxABIEntries == 0 || l.MaxPromiseStateWrites == 0 {
		return errors.New("AVM 2.0 gas, ABI, and promise limits must be positive")
	}
	return nil
}

func (s AVMInstructionSet) Validate() error {
	s = canonicalAVMInstructionSet(s)
	if s.Version == 0 {
		return errors.New("AVM 2.0 instruction set version must be positive")
	}
	if len(s.Opcodes) == 0 {
		return errors.New("AVM 2.0 instruction set must declare opcodes")
	}
	seen := make(map[AVMOpcode]struct{}, len(s.Opcodes))
	for i, opcode := range s.Opcodes {
		if !IsAVMSupportedOpcode(opcode.Opcode) {
			return fmt.Errorf("AVM 2.0 instruction set contains unsupported opcode %q", opcode.Opcode)
		}
		if IsAVMForbiddenOpcode(opcode.Opcode) {
			return fmt.Errorf("AVM 2.0 instruction set contains forbidden opcode %q", opcode.Opcode)
		}
		if !IsAVMInstructionCategory(opcode.Category) {
			return fmt.Errorf("invalid AVM 2.0 instruction category %q", opcode.Category)
		}
		if strings.TrimSpace(opcode.Purpose) == "" {
			return errors.New("AVM 2.0 opcode purpose is required")
		}
		if len(opcode.Purpose) > 256 {
			return errors.New("AVM 2.0 opcode purpose must be <= 256 bytes")
		}
		if opcode.GasCost == 0 {
			return errors.New("AVM 2.0 opcode gas cost must be positive")
		}
		if _, found := seen[opcode.Opcode]; found {
			return fmt.Errorf("duplicate AVM 2.0 opcode %q", opcode.Opcode)
		}
		seen[opcode.Opcode] = struct{}{}
		if i > 0 && s.Opcodes[i-1].Opcode >= opcode.Opcode {
			return errors.New("AVM 2.0 opcodes must be sorted canonically")
		}
	}
	for _, opcode := range AllAVMSupportedOpcodes() {
		if _, found := seen[opcode]; !found {
			return fmt.Errorf("AVM 2.0 instruction set missing opcode %q", opcode)
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 instruction set hash", s.SetHash); err != nil {
		return err
	}
	if s.SetHash != ComputeAVMInstructionSetHash(s) {
		return errors.New("AVM 2.0 instruction set hash mismatch")
	}
	return nil
}

func (t AVMGasTable) Validate() error {
	t = canonicalAVMGasTable(t)
	if t.BaseInstructionGas == 0 || t.MemoryByteGas == 0 || t.StorageReadGas == 0 || t.StorageWriteGas == 0 {
		return errors.New("AVM 2.0 gas table execution and storage gas must be positive")
	}
	if t.StorageReadByteGas == 0 || t.StorageWriteByteGas == 0 || t.ProofByteGas == 0 || t.ProofDepthGas == 0 {
		return errors.New("AVM 2.0 gas table byte and proof gas must be positive")
	}
	if t.MessageCreateGas == 0 || t.MessageEnvelopeGas == 0 || t.MessageByteGas == 0 || t.ForwardingFeeReserveGas == 0 {
		return errors.New("AVM 2.0 gas table message gas must be positive")
	}
	if t.EventByteGas == 0 || t.ABIExportGas == 0 || t.PromiseGas == 0 {
		return errors.New("AVM 2.0 gas table proof, message, event, ABI, and promise gas must be positive")
	}
	if err := validateAVMOpcodeGasEntries(t.OpcodeCosts); err != nil {
		return err
	}
	if t.TableHash == "" {
		return errors.New("AVM 2.0 gas table hash is required")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 gas table hash", t.TableHash); err != nil {
		return err
	}
	if t.TableHash != ComputeAVMGasTableHash(t) {
		return errors.New("AVM 2.0 gas table hash mismatch")
	}
	return nil
}

func validateAVMOpcodeGasEntries(entries []AVMOpcodeGasEntry) error {
	if len(entries) == 0 {
		return errors.New("AVM 2.0 gas table must declare opcode costs")
	}
	entries = append([]AVMOpcodeGasEntry(nil), entries...)
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Opcode < entries[j].Opcode })
	seen := make(map[AVMOpcode]struct{}, len(entries))
	for i, entry := range entries {
		if !IsAVMSupportedOpcode(entry.Opcode) {
			return fmt.Errorf("AVM 2.0 gas table contains unsupported opcode %q", entry.Opcode)
		}
		if entry.Gas == 0 {
			return fmt.Errorf("AVM 2.0 gas table opcode %q cost must be positive", entry.Opcode)
		}
		if _, found := seen[entry.Opcode]; found {
			return fmt.Errorf("duplicate AVM 2.0 gas table opcode %q", entry.Opcode)
		}
		seen[entry.Opcode] = struct{}{}
		if i > 0 && entries[i-1].Opcode >= entry.Opcode {
			return errors.New("AVM 2.0 gas table opcode entries must be sorted canonically")
		}
	}
	for _, opcode := range AllAVMSupportedOpcodes() {
		if _, found := seen[opcode]; !found {
			return fmt.Errorf("AVM 2.0 gas table missing opcode %q", opcode)
		}
	}
	return nil
}

func (c AVMExecutionContext) Validate() error {
	c = canonicalAVMExecutionContext(c)
	if err := validateEngineToken("AVM 2.0 chain id", c.ChainID, MaxAVMTokenLength); err != nil {
		return err
	}
	if c.Height == 0 {
		return errors.New("AVM 2.0 context height must be positive")
	}
	if err := zonestypes.ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 contract address", c.ContractAddress, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 caller", c.Caller, MaxAVMTokenLength); err != nil {
		return err
	}
	if c.GasLimit == 0 {
		return errors.New("AVM 2.0 context gas limit must be positive")
	}
	if c.ContextHash == "" {
		return errors.New("AVM 2.0 context hash is required")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 context hash", c.ContextHash); err != nil {
		return err
	}
	if c.ContextHash != ComputeAVMContextHash(c) {
		return errors.New("AVM 2.0 context hash mismatch")
	}
	return nil
}

func (i AVMInstruction) Validate(limits AVMLimits, gasTable AVMGasTable) error {
	if IsAVMForbiddenOpcode(i.Opcode) {
		return fmt.Errorf("AVM 2.0 forbidden opcode %q", i.Opcode)
	}
	if !IsAVMSupportedOpcode(i.Opcode) {
		return fmt.Errorf("unsupported AVM 2.0 opcode %q", i.Opcode)
	}
	gas, err := AVMInstructionGas(i, gasTable, limits)
	if err != nil {
		return err
	}
	if gas == 0 || gas > limits.MaxInstructionGas {
		return fmt.Errorf("AVM 2.0 instruction gas must be in 1..%d", limits.MaxInstructionGas)
	}
	switch i.Opcode {
	case AVMOpLoadLocal, AVMOpStoreLocal:
		if err := validateEngineToken("AVM 2.0 local key", i.Key, MaxAVMTokenLength); err != nil {
			return err
		}
	case AVMOpJmp, AVMOpJmpIf, AVMOpCallInternal:
		if i.RangeLimit == 0 {
			return errors.New("AVM 2.0 control flow target must be explicit")
		}
	case AVMOpMemGrow:
		if i.MemoryGrow == 0 {
			return errors.New("AVM 2.0 memory growth must be positive")
		}
	case AVMOpMemLoad, AVMOpMemStore, AVMOpMemCopy:
		if uint64(i.MemoryGrow) > limits.MaxMemoryBytes {
			return errors.New("AVM 2.0 memory operation exceeds memory limit")
		}
	case AVMOpKVGet, AVMOpKVSet, AVMOpKVDelete, AVMOpKVExists, AVMOpKVRangeBounded:
		if strings.TrimSpace(i.Key) == "" {
			return errors.New("AVM 2.0 storage key is required")
		}
		if uint32(len(i.Key)) > limits.MaxStorageKeyBytes {
			return fmt.Errorf("AVM 2.0 storage key bytes must be <= %d", limits.MaxStorageKeyBytes)
		}
		if i.Opcode == AVMOpKVRangeBounded && (i.RangeLimit == 0 || i.RangeLimit > limits.MaxBoundedRangeItems) {
			return fmt.Errorf("AVM 2.0 bounded range limit must be in 1..%d", limits.MaxBoundedRangeItems)
		}
	case AVMOpVerifySig, AVMOpVerifyMerkleProof, AVMOpVerifyMessageProof, AVMOpVerifyZoneRoot:
		return canonicalAVMProofInput(i.Proof).Validate(limits)
	case AVMOpMsgNew, AVMOpMsgSetValue, AVMOpMsgSetPayload, AVMOpMsgSetGas, AVMOpMsgSetExpiry, AVMOpMsgSend, AVMOpMsgBounce:
		return canonicalAVMAsyncMessage(i.Message).Validate()
	case AVMOpPromiseNew, AVMOpPromiseAwait, AVMOpPromiseResolve, AVMOpPromiseReject, AVMOpPromiseTimeout:
		return canonicalAVMPromiseState(i.Promise).Validate()
	case AVMOpABIExport, AVMOpABIMethod, AVMOpABIEvent, AVMOpABIRequire:
		return canonicalAVMABIDescriptor(i.ABI).Validate(limits)
	case AVMOpEmitEvent:
		return canonicalAVMEvent(i.Event).Validate(limits)
	}
	return nil
}

func (p AVMProofInput) Validate(limits AVMLimits) error {
	p = canonicalAVMProofInput(p)
	if p.ProofVersion == 0 {
		return errors.New("AVM 2.0 proof version must be positive")
	}
	if err := validateEngineToken("AVM 2.0 proof chain id", p.ChainID, MaxAVMTokenLength); err != nil {
		return err
	}
	if p.Height == 0 {
		return errors.New("AVM 2.0 proof height must be positive")
	}
	if !IsAVMProofRootType(p.RootType) {
		return fmt.Errorf("invalid AVM 2.0 proof root type %q", p.RootType)
	}
	if err := zonestypes.ValidateHash("AVM 2.0 proof root hash", p.RootHash); err != nil {
		return err
	}
	if strings.TrimSpace(p.Key) == "" {
		return errors.New("AVM 2.0 proof key is required")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 proof value hash", p.ValueHash); err != nil {
		return err
	}
	if len(p.ProofBytes) == 0 || uint32(len(p.ProofBytes)) > limits.MaxProofBytes {
		return fmt.Errorf("AVM 2.0 proof bytes must be in 1..%d", limits.MaxProofBytes)
	}
	if err := zonestypes.ValidateHash("AVM 2.0 proof hash", p.ProofHash); err != nil {
		return err
	}
	if p.ProofHash != ComputeAVMProofHash(p) {
		return errors.New("AVM 2.0 proof hash mismatch")
	}
	return nil
}

func (p AVMPromiseState) Validate() error {
	p = canonicalAVMPromiseState(p)
	if err := zonestypes.ValidateHash("AVM 2.0 promise id", p.PromiseID); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 promise contract", p.Contract, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise message id", p.MessageID); err != nil {
		return err
	}
	if !IsAVMPromiseStatus(p.Status) {
		return fmt.Errorf("invalid AVM 2.0 promise status %q", p.Status)
	}
	if p.CreatedHeight == 0 || p.ExpiryHeight <= p.CreatedHeight {
		return errors.New("AVM 2.0 promise heights are invalid")
	}
	if p.Status != AVMPromisePending {
		if err := zonestypes.ValidateHash("AVM 2.0 promise receipt hash", p.ReceiptHash); err != nil {
			return err
		}
	}
	if p.ReturnHash != "" {
		if err := zonestypes.ValidateHash("AVM 2.0 promise return hash", p.ReturnHash); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise hash", p.PromiseHash); err != nil {
		return err
	}
	if p.PromiseHash != ComputeAVMPromiseHash(p) {
		return errors.New("AVM 2.0 promise hash mismatch")
	}
	return nil
}

func (a AVMABIDescriptor) Validate(limits AVMLimits) error {
	a = canonicalAVMABIDescriptor(a)
	if a.ABIVersion == 0 || a.CodeID == 0 {
		return errors.New("AVM 2.0 ABI version and code id must be positive")
	}
	for _, set := range []struct {
		name	string
		values	[]string
	}{
		{name: "AVM 2.0 ABI method", values: a.Methods},
		{name: "AVM 2.0 ABI event", values: a.Events},
		{name: "AVM 2.0 ABI error", values: a.Errors},
		{name: "AVM 2.0 ABI required fund", values: a.RequiredFunds},
		{name: "AVM 2.0 ABI gas hint", values: a.GasHints},
	} {
		if len(set.values) > int(limits.MaxABIEntries) {
			return fmt.Errorf("%s entries must be <= %d", set.name, limits.MaxABIEntries)
		}
		if err := validateEngineTokens(set.name, set.values, MaxAVMTokenLength); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI interface hash", a.InterfaceHash); err != nil {
		return err
	}
	if a.InterfaceHash != ComputeAVMABIInterfaceHash(a) {
		return errors.New("AVM 2.0 ABI interface hash mismatch")
	}
	return nil
}

func (e AVMEvent) Validate(limits AVMLimits) error {
	e = canonicalAVMEvent(e)
	if e.Height == 0 {
		return errors.New("AVM 2.0 event height must be positive")
	}
	if err := validateEngineToken("AVM 2.0 event contract", e.ContractAddress, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 event id", e.EventID, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 event name", e.Name, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 event payload hash", e.PayloadHash); err != nil {
		return err
	}
	if uint32(len(e.PayloadHash)) > limits.MaxEventPayloadBytes {
		return fmt.Errorf("AVM 2.0 event payload bytes must be <= %d", limits.MaxEventPayloadBytes)
	}
	if err := zonestypes.ValidateHash("AVM 2.0 event hash", e.EventHash); err != nil {
		return err
	}
	if e.EventHash != ComputeAVMEventHash(e) {
		return errors.New("AVM 2.0 event hash mismatch")
	}
	return nil
}

func (r AVMExecutionResult) Validate() error {
	r = canonicalAVMExecutionResult(r)
	for i, read := range r.StorageReads {
		if strings.TrimSpace(read.Key) == "" {
			return errors.New("AVM 2.0 storage read key is required")
		}
		if err := zonestypes.ValidateHash("AVM 2.0 storage read key hash", read.KeyHash); err != nil {
			return err
		}
		if read.KeyHash != ComputeAVMBytesHash([]byte(read.Key)) {
			return errors.New("AVM 2.0 storage read key hash mismatch")
		}
		if i > 0 && r.StorageReads[i-1].Key >= read.Key {
			return errors.New("AVM 2.0 storage reads must be sorted canonically")
		}
	}
	for i, write := range r.StorageWrites {
		if strings.TrimSpace(write.Key) == "" {
			return errors.New("AVM 2.0 storage write key is required")
		}
		if err := zonestypes.ValidateHash("AVM 2.0 storage write value hash", write.ValueHash); err != nil {
			return err
		}
		if i > 0 && r.StorageWrites[i-1].Key >= write.Key {
			return errors.New("AVM 2.0 storage writes must be sorted canonically")
		}
	}
	for _, msg := range r.OutputMessages {
		if err := msg.Validate(); err != nil {
			return err
		}
	}
	for _, proof := range r.ProofsVerified {
		if err := proof.Validate(DefaultAVMLimits()); err != nil {
			return err
		}
	}
	for _, promise := range r.Promises {
		if err := promise.Validate(); err != nil {
			return err
		}
	}
	for _, abi := range r.ABIDescriptors {
		if err := abi.Validate(DefaultAVMLimits()); err != nil {
			return err
		}
	}
	for _, event := range r.Events {
		if err := event.Validate(DefaultAVMLimits()); err != nil {
			return err
		}
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM 2.0 storage root", value: r.StorageRoot},
		{name: "AVM 2.0 message root", value: r.MessageRoot},
		{name: "AVM 2.0 promise root", value: r.PromiseRoot},
		{name: "AVM 2.0 ABI root", value: r.ABIRoot},
		{name: "AVM 2.0 event root", value: r.EventRoot},
		{name: "AVM 2.0 execution hash", value: r.ExecutionHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.StorageRoot != ComputeAVMStorageRoot(r.StorageReads, r.StorageWrites) {
		return errors.New("AVM 2.0 storage root mismatch")
	}
	if r.MessageRoot != ComputeAVMMessageRoot(r.OutputMessages) {
		return errors.New("AVM 2.0 message root mismatch")
	}
	if r.PromiseRoot != ComputeAVMPromiseRoot(r.Promises) {
		return errors.New("AVM 2.0 promise root mismatch")
	}
	if r.ABIRoot != ComputeAVMABIRoot(r.ABIDescriptors) {
		return errors.New("AVM 2.0 ABI root mismatch")
	}
	if r.EventRoot != ComputeAVMEventRoot(r.Events) {
		return errors.New("AVM 2.0 event root mismatch")
	}
	if r.ExecutionHash != ComputeAVMExecutionHash(r) {
		return errors.New("AVM 2.0 execution hash mismatch")
	}
	if r.ReadOnlySimulation && (len(r.StorageWrites) > 0 || len(r.OutputMessages) > 0 || len(r.Events) > 0 || len(r.Promises) > 0) {
		return errors.New("AVM 2.0 read-only simulation produced mutable outputs")
	}
	return nil
}

func ValidateAVMStoreV2Key(ctx AVMExecutionContext, key string, limits AVMLimits) error {
	if err := validateAVMStatePrefix("AVM 2.0 Store v2 key", key); err != nil {
		return err
	}
	if uint32(len(key)) > limits.MaxStorageKeyBytes {
		return fmt.Errorf("AVM 2.0 Store v2 key bytes must be <= %d", limits.MaxStorageKeyBytes)
	}
	expected := AVMContractStorageKey(ctx.ContractAddress, "")
	if !strings.HasPrefix(key, expected) {
		return fmt.Errorf("AVM 2.0 Store v2 key must use contract-local prefix %q", expected)
	}
	return nil
}

func AVMInstructionGas(i AVMInstruction, table AVMGasTable, limits AVMLimits) (uint64, error) {
	table = canonicalAVMGasTable(table)
	if i.GasOverride > 0 {
		return i.GasOverride, nil
	}
	opGas, found := table.OpcodeGas(i.Opcode)
	if !found {
		return 0, fmt.Errorf("AVM 2.0 gas table missing opcode %q", i.Opcode)
	}
	gas, err := checkedAVMGasAdd(table.BaseInstructionGas, opGas)
	if err != nil {
		return 0, err
	}
	switch i.Opcode {
	case AVMOpMemGrow, AVMOpMemLoad, AVMOpMemStore, AVMOpMemCopy:
		gas, err = checkedAVMGasAdd(gas, uint64(i.MemoryGrow)*table.MemoryByteGas)
	case AVMOpKVGet, AVMOpKVExists, AVMOpKVRangeBounded:
		gas, err = checkedAVMGasAdd(gas, table.StorageReadGas)
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(len(i.Key))*table.StorageReadByteGas)
		}
	case AVMOpKVSet, AVMOpKVDelete:
		gas, err = checkedAVMGasAdd(gas, table.StorageWriteGas)
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(len(i.Key))*table.StorageWriteByteGas)
		}
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(len(i.Value))*table.StorageWriteByteGas)
		}
	case AVMOpVerifySig, AVMOpVerifyMerkleProof, AVMOpVerifyMessageProof, AVMOpVerifyZoneRoot:
		gas, err = checkedAVMGasAdd(gas, uint64(len(i.Proof.ProofBytes))*table.ProofByteGas)
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(i.Proof.ProofVersion)*table.ProofDepthGas)
		}
	case AVMOpMsgNew, AVMOpMsgSetValue, AVMOpMsgSetPayload, AVMOpMsgSetGas, AVMOpMsgSetExpiry, AVMOpMsgSend, AVMOpMsgBounce:
		gas, err = checkedAVMGasAdd(gas, table.MessageCreateGas)
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, table.MessageEnvelopeGas)
		}
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(len(i.Message.Payload))*table.MessageByteGas)
		}
		if err == nil && (i.Opcode == AVMOpMsgSend || i.Opcode == AVMOpMsgBounce) {
			gas, err = checkedAVMGasAdd(gas, table.ForwardingFeeReserveGas)
		}
	case AVMOpPromiseNew, AVMOpPromiseAwait, AVMOpPromiseResolve, AVMOpPromiseReject, AVMOpPromiseTimeout:
		gas, err = checkedAVMGasAdd(gas, table.PromiseGas)
	case AVMOpABIExport, AVMOpABIMethod, AVMOpABIEvent, AVMOpABIRequire:
		gas, err = checkedAVMGasAdd(gas, table.ABIExportGas)
	case AVMOpEmitEvent:
		gas, err = checkedAVMGasAdd(gas, table.EventByteGas)
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(len(i.Event.PayloadHash))*table.EventByteGas)
		}
	}
	if err != nil {
		return 0, err
	}
	if gas > limits.MaxInstructionGas {
		return 0, fmt.Errorf("AVM 2.0 instruction gas must be <= %d", limits.MaxInstructionGas)
	}
	return gas, nil
}

func IsAVMSupportedOpcode(op AVMOpcode) bool {
	switch op {
	case AVMOpNoop, AVMOpPush, AVMOpPop, AVMOpDup, AVMOpSwap, AVMOpLoadLocal, AVMOpStoreLocal,
		AVMOpAdd, AVMOpSub, AVMOpMul, AVMOpDiv, AVMOpMod, AVMOpNeg, AVMOpCmp,
		AVMOpJmp, AVMOpJmpIf, AVMOpCallInternal, AVMOpRet, AVMOpAbort,
		AVMOpMemLoad, AVMOpMemStore, AVMOpMemCopy, AVMOpMemSize, AVMOpMemGrow,
		AVMOpKVGet, AVMOpKVSet, AVMOpKVDelete, AVMOpKVExists, AVMOpKVRangeBounded,
		AVMOpHash, AVMOpVerifySig, AVMOpVerifyMerkleProof, AVMOpVerifyMessageProof, AVMOpVerifyZoneRoot,
		AVMOpMsgNew, AVMOpMsgSetValue, AVMOpMsgSetPayload, AVMOpMsgSetGas, AVMOpMsgSetExpiry, AVMOpMsgSend, AVMOpMsgBounce,
		AVMOpPromiseNew, AVMOpPromiseAwait, AVMOpPromiseResolve, AVMOpPromiseReject, AVMOpPromiseTimeout,
		AVMOpABIExport, AVMOpABIMethod, AVMOpABIEvent, AVMOpABIRequire, AVMOpEmitEvent,
		AVMOpCtxHeight, AVMOpCtxChainID, AVMOpCtxZoneID, AVMOpCtxShardID, AVMOpCtxCaller, AVMOpCtxContract, AVMOpCtxValue, AVMOpCtxGasLeft:
		return true
	default:
		return false
	}
}

func AllAVMSupportedOpcodes() []AVMOpcode {
	return []AVMOpcode{
		AVMOpAbort, AVMOpABIEvent, AVMOpABIExport, AVMOpABIMethod, AVMOpABIRequire,
		AVMOpAdd, AVMOpCallInternal, AVMOpCmp, AVMOpCtxCaller, AVMOpCtxChainID,
		AVMOpCtxContract, AVMOpCtxGasLeft, AVMOpCtxHeight, AVMOpCtxShardID, AVMOpCtxValue,
		AVMOpCtxZoneID, AVMOpDiv, AVMOpDup, AVMOpEmitEvent, AVMOpHash, AVMOpJmp,
		AVMOpJmpIf, AVMOpKVDelete, AVMOpKVExists, AVMOpKVGet, AVMOpKVRangeBounded,
		AVMOpKVSet, AVMOpLoadLocal, AVMOpMemCopy, AVMOpMemGrow, AVMOpMemLoad,
		AVMOpMemSize, AVMOpMemStore, AVMOpMod, AVMOpMsgBounce, AVMOpMsgNew,
		AVMOpMsgSend, AVMOpMsgSetExpiry, AVMOpMsgSetGas, AVMOpMsgSetPayload,
		AVMOpMsgSetValue, AVMOpMul, AVMOpNeg, AVMOpNoop, AVMOpPop, AVMOpPromiseAwait,
		AVMOpPromiseNew, AVMOpPromiseReject, AVMOpPromiseResolve, AVMOpPromiseTimeout,
		AVMOpPush, AVMOpRet, AVMOpStoreLocal, AVMOpSub, AVMOpSwap,
		AVMOpVerifyMerkleProof, AVMOpVerifyMessageProof, AVMOpVerifySig, AVMOpVerifyZoneRoot,
	}
}

func IsAVMInstructionCategory(category AVMInstructionCategory) bool {
	switch category {
	case AVMCategoryCoreStack, AVMCategoryArithmetic, AVMCategoryControlFlow, AVMCategoryMemory, AVMCategoryStorage,
		AVMCategoryCryptoProof, AVMCategoryMessages, AVMCategoryPromises, AVMCategoryABI, AVMCategoryContext:
		return true
	default:
		return false
	}
}

func IsAVMForbiddenOpcode(op AVMOpcode) bool {
	switch op {
	case AVMOpExternalNetwork, AVMOpWallClock, AVMOpNonDeterministic, AVMOpKVRangeUnbounded, AVMOpDirectRemoteMutation, AVMOpUnboundedRecursion:
		return true
	default:
		return false
	}
}

func IsAVMPromiseStatus(status AVMPromiseStatus) bool {
	switch status {
	case AVMPromisePending, AVMPromiseResolved, AVMPromiseRejected, AVMPromiseTimedOut, AVMPromiseRefunded:
		return true
	default:
		return false
	}
}

func IsAVMProofRootType(rootType AVMProofRootType) bool {
	switch rootType {
	case AVMProofRootZone, AVMProofRootShard, AVMProofRootMessage, AVMProofRootReceipt, AVMProofRootContractState, AVMProofRootResolverRecord:
		return true
	default:
		return false
	}
}

func AVMOpcodeGasEntriesFromInstructionSet(set AVMInstructionSet) []AVMOpcodeGasEntry {
	set = canonicalAVMInstructionSet(set)
	entries := make([]AVMOpcodeGasEntry, 0, len(set.Opcodes))
	for _, opcode := range set.Opcodes {
		entries = append(entries, AVMOpcodeGasEntry{Opcode: opcode.Opcode, Gas: opcode.GasCost})
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Opcode < entries[j].Opcode })
	return entries
}

func (t AVMGasTable) OpcodeGas(opcode AVMOpcode) (uint64, bool) {
	t = canonicalAVMGasTable(t)
	for _, entry := range t.OpcodeCosts {
		if entry.Opcode == opcode {
			return entry.Gas, true
		}
	}
	return 0, false
}

func AVMContextValue(opcode AVMOpcode, ctx AVMExecutionContext, gasUsed uint64) string {
	ctx = canonicalAVMExecutionContext(ctx)
	switch opcode {
	case AVMOpCtxHeight:
		return fmt.Sprintf("%020d", ctx.Height)
	case AVMOpCtxChainID:
		return ctx.ChainID
	case AVMOpCtxZoneID:
		return string(ctx.ZoneID)
	case AVMOpCtxShardID:
		return fmt.Sprintf("%020d", ctx.ShardID)
	case AVMOpCtxCaller:
		return ctx.Caller
	case AVMOpCtxContract:
		return ctx.ContractAddress
	case AVMOpCtxValue:
		return fmt.Sprintf("%020d", ctx.ValueNAET)
	case AVMOpCtxGasLeft:
		if gasUsed >= ctx.GasLimit {
			return "00000000000000000000"
		}
		return fmt.Sprintf("%020d", ctx.GasLimit-gasUsed)
	default:
		return ""
	}
}

func ComputeAVMBytesHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func ComputeAVMInstructionSetHash(set AVMInstructionSet) string {
	set = canonicalAVMInstructionSet(set)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-instruction-set-v1")
	writeEngineUint64(h, set.Version)
	writeEngineUint64(h, uint64(len(set.Opcodes)))
	for _, opcode := range set.Opcodes {
		writeEnginePart(h, string(opcode.Opcode))
		writeEnginePart(h, string(opcode.Category))
		writeEnginePart(h, opcode.Purpose)
		writeEngineUint64(h, opcode.GasCost)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMGasTableHash(table AVMGasTable) string {
	table = canonicalAVMGasTable(table)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-gas-table-v1")
	writeEngineUint64(h, table.BaseInstructionGas)
	writeEngineUint64(h, uint64(len(table.OpcodeCosts)))
	for _, entry := range table.OpcodeCosts {
		writeEnginePart(h, string(entry.Opcode))
		writeEngineUint64(h, entry.Gas)
	}
	writeEngineUint64(h, table.MemoryByteGas)
	writeEngineUint64(h, table.StorageReadGas)
	writeEngineUint64(h, table.StorageWriteGas)
	writeEngineUint64(h, table.StorageReadByteGas)
	writeEngineUint64(h, table.StorageWriteByteGas)
	writeEngineUint64(h, table.ProofByteGas)
	writeEngineUint64(h, table.ProofDepthGas)
	writeEngineUint64(h, table.MessageCreateGas)
	writeEngineUint64(h, table.MessageEnvelopeGas)
	writeEngineUint64(h, table.MessageByteGas)
	writeEngineUint64(h, table.ForwardingFeeReserveGas)
	writeEngineUint64(h, table.EventByteGas)
	writeEngineUint64(h, table.ABIExportGas)
	writeEngineUint64(h, table.PromiseGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContextHash(ctx AVMExecutionContext) string {
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-context-v1")
	writeEnginePart(h, ctx.ChainID)
	writeEngineUint64(h, ctx.Height)
	writeEnginePart(h, string(ctx.ZoneID))
	writeEngineUint64(h, uint64(ctx.ShardID))
	writeEnginePart(h, ctx.ContractAddress)
	writeEnginePart(h, ctx.Caller)
	writeEngineUint64(h, ctx.ValueNAET)
	writeEngineUint64(h, ctx.GasLimit)
	writeEngineBool(h, ctx.ReadOnly)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMProgramHash(program AVMProgram) string {
	program = canonicalAVMProgram(program)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-program-v1")
	writeEngineUint64(h, program.VMVersion)
	writeEngineUint64(h, program.InstructionSetVersion)
	writeEngineUint64(h, uint64(program.MaxRecursionDepth))
	writeEngineUint64(h, uint64(len(program.Instructions)))
	for _, instruction := range program.Instructions {
		writeAVMInstructionParts(h, instruction)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMProofHash(proof AVMProofInput) string {
	proof = canonicalAVMProofInput(proof)
	proof.ProofHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-proof-v1")
	writeEngineUint64(h, proof.ProofVersion)
	writeEnginePart(h, proof.ChainID)
	writeEngineUint64(h, proof.Height)
	writeEnginePart(h, string(proof.RootType))
	writeEnginePart(h, proof.RootHash)
	writeEnginePart(h, proof.Key)
	writeEnginePart(h, proof.ValueHash)
	writeEnginePart(h, hex.EncodeToString(proof.ProofBytes))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMPromiseHash(promise AVMPromiseState) string {
	promise = canonicalAVMPromiseState(promise)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-promise-v1")
	writeEnginePart(h, promise.PromiseID)
	writeEnginePart(h, promise.Contract)
	writeEnginePart(h, promise.MessageID)
	writeEnginePart(h, string(promise.Status))
	writeEngineUint64(h, promise.CreatedHeight)
	writeEngineUint64(h, promise.ExpiryHeight)
	writeEnginePart(h, promise.ReceiptHash)
	writeEnginePart(h, promise.ReturnHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMABIInterfaceHash(abi AVMABIDescriptor) string {
	abi = canonicalAVMABIDescriptor(abi)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-abi-v1")
	writeEngineUint64(h, abi.ABIVersion)
	writeEngineUint64(h, abi.CodeID)
	writeStringSet(h, abi.Methods)
	writeStringSet(h, abi.Events)
	writeStringSet(h, abi.Errors)
	writeStringSet(h, abi.RequiredFunds)
	writeStringSet(h, abi.GasHints)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMEventHash(event AVMEvent) string {
	event = canonicalAVMEvent(event)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-event-v1")
	writeEngineUint64(h, event.Height)
	writeEnginePart(h, event.ContractAddress)
	writeEnginePart(h, event.EventID)
	writeEnginePart(h, event.Name)
	writeEnginePart(h, event.PayloadHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStorageRoot(reads []AVMStorageRead, writes []AVMStorageWrite) string {
	reads = canonicalAVMStorageReads(reads)
	writes = canonicalAVMStorageWrites(writes)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-storage-root-v1")
	writeEngineUint64(h, uint64(len(reads)))
	for _, read := range reads {
		writeEnginePart(h, read.Key)
		writeEnginePart(h, read.KeyHash)
	}
	writeEngineUint64(h, uint64(len(writes)))
	for _, write := range writes {
		writeEnginePart(h, write.Key)
		writeEnginePart(h, write.ValueHash)
		writeEngineBool(h, write.Deleted)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMMessageRoot(messages []AVMAsyncMessage) string {
	messages = canonicalAVMMessages(messages)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-message-root-v1")
	writeEngineUint64(h, uint64(len(messages)))
	for _, msg := range messages {
		writeAVMAsyncMessageParts(h, msg)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMPromiseRoot(promises []AVMPromiseState) string {
	promises = canonicalAVMPromises(promises)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-promise-root-v1")
	writeEngineUint64(h, uint64(len(promises)))
	for _, promise := range promises {
		writeEnginePart(h, promise.PromiseHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMABIRoot(abis []AVMABIDescriptor) string {
	abis = canonicalAVMABIs(abis)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-abi-root-v1")
	writeEngineUint64(h, uint64(len(abis)))
	for _, abi := range abis {
		writeEnginePart(h, abi.InterfaceHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMEventRoot(events []AVMEvent) string {
	events = canonicalAVMEvents(events)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-event-root-v1")
	writeEngineUint64(h, uint64(len(events)))
	for _, event := range events {
		writeEnginePart(h, event.EventHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMExecutionHash(result AVMExecutionResult) string {
	result = canonicalAVMExecutionResult(result)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-execution-v1")
	writeEngineUint64(h, result.GasUsed)
	writeEngineUint64(h, result.MemoryBytes)
	writeStringSet(h, result.Stack)
	writeEnginePart(h, result.StorageRoot)
	writeEnginePart(h, result.MessageRoot)
	writeEnginePart(h, result.PromiseRoot)
	writeEnginePart(h, result.ABIRoot)
	writeEnginePart(h, result.EventRoot)
	writeEngineBool(h, result.ReadOnlySimulation)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMExecutionContext(ctx AVMExecutionContext) AVMExecutionContext {
	ctx.ChainID = strings.TrimSpace(ctx.ChainID)
	ctx.ContractAddress = strings.TrimSpace(ctx.ContractAddress)
	ctx.Caller = strings.TrimSpace(ctx.Caller)
	ctx.ContextHash = strings.TrimSpace(ctx.ContextHash)
	return ctx
}

func canonicalAVMInstructionSet(set AVMInstructionSet) AVMInstructionSet {
	set.Opcodes = append([]AVMOpcodeDescriptor(nil), set.Opcodes...)
	for i := range set.Opcodes {
		set.Opcodes[i].Purpose = strings.TrimSpace(set.Opcodes[i].Purpose)
	}
	sort.SliceStable(set.Opcodes, func(i, j int) bool {
		return set.Opcodes[i].Opcode < set.Opcodes[j].Opcode
	})
	set.SetHash = strings.TrimSpace(set.SetHash)
	return set
}

func canonicalAVMGasTable(table AVMGasTable) AVMGasTable {
	table.OpcodeCosts = append([]AVMOpcodeGasEntry(nil), table.OpcodeCosts...)
	sort.SliceStable(table.OpcodeCosts, func(i, j int) bool {
		return table.OpcodeCosts[i].Opcode < table.OpcodeCosts[j].Opcode
	})
	table.TableHash = strings.TrimSpace(table.TableHash)
	return table
}

func canonicalAVMProgram(program AVMProgram) AVMProgram {
	program.Instructions = append([]AVMInstruction(nil), program.Instructions...)
	for i := range program.Instructions {
		program.Instructions[i] = canonicalAVMInstruction(program.Instructions[i])
	}
	program.ProgramHash = strings.TrimSpace(program.ProgramHash)
	return program
}

func canonicalAVMInstruction(instruction AVMInstruction) AVMInstruction {
	instruction.Key = strings.TrimSpace(instruction.Key)
	instruction.Message = canonicalAVMAsyncMessage(instruction.Message)
	instruction.Proof = canonicalAVMProofInput(instruction.Proof)
	instruction.Promise = canonicalAVMPromiseState(instruction.Promise)
	instruction.ABI = canonicalAVMABIDescriptor(instruction.ABI)
	instruction.Event = canonicalAVMEvent(instruction.Event)
	instruction.Value = append([]byte(nil), instruction.Value...)
	return instruction
}

func canonicalAVMProofInput(proof AVMProofInput) AVMProofInput {
	proof.ChainID = strings.TrimSpace(proof.ChainID)
	proof.Key = strings.TrimSpace(proof.Key)
	proof.RootHash = strings.TrimSpace(proof.RootHash)
	proof.ValueHash = strings.TrimSpace(proof.ValueHash)
	proof.ProofHash = strings.TrimSpace(proof.ProofHash)
	proof.ProofBytes = append([]byte(nil), proof.ProofBytes...)
	return proof
}

func canonicalAVMPromiseState(promise AVMPromiseState) AVMPromiseState {
	promise.PromiseID = strings.TrimSpace(promise.PromiseID)
	promise.Contract = strings.TrimSpace(promise.Contract)
	promise.MessageID = strings.TrimSpace(promise.MessageID)
	promise.ReceiptHash = strings.TrimSpace(promise.ReceiptHash)
	promise.ReturnHash = strings.TrimSpace(promise.ReturnHash)
	promise.PromiseHash = strings.TrimSpace(promise.PromiseHash)
	return promise
}

func canonicalAVMABIDescriptor(abi AVMABIDescriptor) AVMABIDescriptor {
	abi.Methods = cloneSortedStrings(abi.Methods)
	abi.Events = cloneSortedStrings(abi.Events)
	abi.Errors = cloneSortedStrings(abi.Errors)
	abi.RequiredFunds = cloneSortedStrings(abi.RequiredFunds)
	abi.GasHints = cloneSortedStrings(abi.GasHints)
	abi.InterfaceHash = strings.TrimSpace(abi.InterfaceHash)
	return abi
}

func canonicalAVMEvent(event AVMEvent) AVMEvent {
	event.ContractAddress = strings.TrimSpace(event.ContractAddress)
	event.EventID = strings.TrimSpace(event.EventID)
	event.Name = strings.TrimSpace(event.Name)
	event.PayloadHash = strings.TrimSpace(event.PayloadHash)
	event.EventHash = strings.TrimSpace(event.EventHash)
	return event
}

func canonicalAVMExecutionResult(result AVMExecutionResult) AVMExecutionResult {
	result.Stack = append([]string(nil), result.Stack...)
	for i := range result.Stack {
		result.Stack[i] = strings.TrimSpace(result.Stack[i])
	}
	result.StorageReads = canonicalAVMStorageReads(result.StorageReads)
	result.StorageWrites = canonicalAVMStorageWrites(result.StorageWrites)
	result.OutputMessages = canonicalAVMMessages(result.OutputMessages)
	result.ProofsVerified = canonicalAVMProofs(result.ProofsVerified)
	result.Promises = canonicalAVMPromises(result.Promises)
	result.ABIDescriptors = canonicalAVMABIs(result.ABIDescriptors)
	result.Events = canonicalAVMEvents(result.Events)
	result.StorageRoot = strings.TrimSpace(result.StorageRoot)
	result.MessageRoot = strings.TrimSpace(result.MessageRoot)
	result.PromiseRoot = strings.TrimSpace(result.PromiseRoot)
	result.ABIRoot = strings.TrimSpace(result.ABIRoot)
	result.EventRoot = strings.TrimSpace(result.EventRoot)
	result.ExecutionHash = strings.TrimSpace(result.ExecutionHash)
	return result
}

func canonicalAVMStorageReads(reads []AVMStorageRead) []AVMStorageRead {
	out := append([]AVMStorageRead(nil), reads...)
	for i := range out {
		out[i].Key = strings.TrimSpace(out[i].Key)
		out[i].KeyHash = strings.TrimSpace(out[i].KeyHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func canonicalAVMStorageWrites(writes []AVMStorageWrite) []AVMStorageWrite {
	out := append([]AVMStorageWrite(nil), writes...)
	for i := range out {
		out[i].Key = strings.TrimSpace(out[i].Key)
		out[i].ValueHash = strings.TrimSpace(out[i].ValueHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func canonicalAVMMessages(messages []AVMAsyncMessage) []AVMAsyncMessage {
	out := append([]AVMAsyncMessage(nil), messages...)
	for i := range out {
		out[i] = canonicalAVMAsyncMessage(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func canonicalAVMProofs(proofs []AVMProofInput) []AVMProofInput {
	out := append([]AVMProofInput(nil), proofs...)
	for i := range out {
		out[i] = canonicalAVMProofInput(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ProofHash < out[j].ProofHash })
	return out
}

func canonicalAVMPromises(promises []AVMPromiseState) []AVMPromiseState {
	out := append([]AVMPromiseState(nil), promises...)
	for i := range out {
		out[i] = canonicalAVMPromiseState(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PromiseID < out[j].PromiseID })
	return out
}

func canonicalAVMABIs(abis []AVMABIDescriptor) []AVMABIDescriptor {
	out := append([]AVMABIDescriptor(nil), abis...)
	for i := range out {
		out[i] = canonicalAVMABIDescriptor(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CodeID == out[j].CodeID {
			return out[i].ABIVersion < out[j].ABIVersion
		}
		return out[i].CodeID < out[j].CodeID
	})
	return out
}

func canonicalAVMEvents(events []AVMEvent) []AVMEvent {
	out := append([]AVMEvent(nil), events...)
	for i := range out {
		out[i] = canonicalAVMEvent(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height == out[j].Height {
			return out[i].EventID < out[j].EventID
		}
		return out[i].Height < out[j].Height
	})
	return out
}

func writeAVMInstructionParts(h engineByteWriter, instruction AVMInstruction) {
	instruction = canonicalAVMInstruction(instruction)
	writeEnginePart(h, string(instruction.Opcode))
	writeEnginePart(h, instruction.Key)
	writeEnginePart(h, hex.EncodeToString(instruction.Value))
	writeAVMAsyncMessageParts(h, instruction.Message)
	writeEnginePart(h, instruction.Proof.ProofHash)
	writeEnginePart(h, instruction.Promise.PromiseHash)
	writeEnginePart(h, instruction.ABI.InterfaceHash)
	writeEnginePart(h, instruction.Event.EventHash)
	writeEngineUint64(h, uint64(instruction.MemoryGrow))
	writeEngineUint64(h, uint64(instruction.RangeLimit))
	writeEngineUint64(h, instruction.GasOverride)
}

func writeStringSet(h engineByteWriter, values []string) {
	values = cloneSortedStrings(values)
	writeEngineUint64(h, uint64(len(values)))
	for _, value := range values {
		writeEnginePart(h, value)
	}
}
