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
	AVM2VMVersion             uint64 = 2
	AVM2DefaultInstructionSet uint64 = 1
	MaxAVM2TokenLength               = 128
	MaxAVM2PayloadTypeLength         = MaxAsyncMessagePayloadType

	AVM2OpNoop              AVM2Opcode = "NOOP"
	AVM2OpPush              AVM2Opcode = "PUSH"
	AVM2OpPop               AVM2Opcode = "POP"
	AVM2OpMemGrow           AVM2Opcode = "MEM_GROW"
	AVM2OpKVGet             AVM2Opcode = "KV_GET"
	AVM2OpKVSet             AVM2Opcode = "KV_SET"
	AVM2OpKVDelete          AVM2Opcode = "KV_DELETE"
	AVM2OpKVRangeBounded    AVM2Opcode = "KV_RANGE_BOUNDED"
	AVM2OpVerifyMerkleProof AVM2Opcode = "VERIFY_MERKLE_PROOF"
	AVM2OpMsgSend           AVM2Opcode = "MSG_SEND"
	AVM2OpPromiseNew        AVM2Opcode = "PROMISE_NEW"
	AVM2OpPromiseResolve    AVM2Opcode = "PROMISE_RESOLVE"
	AVM2OpABIExport         AVM2Opcode = "ABI_EXPORT"
	AVM2OpEmitEvent         AVM2Opcode = "EMIT_EVENT"

	AVM2OpExternalNetwork       AVM2Opcode        = "SYSCALL_EXTERNAL_NETWORK"
	AVM2OpWallClock             AVM2Opcode        = "SYSCALL_WALL_CLOCK"
	AVM2OpNonDeterministic      AVM2Opcode        = "SYSCALL_NON_DETERMINISTIC"
	AVM2OpKVRangeUnbounded      AVM2Opcode        = "KV_RANGE_UNBOUNDED"
	AVM2OpDirectRemoteMutation  AVM2Opcode        = "REMOTE_ZONE_MUTATE"
	AVM2OpUnboundedRecursion    AVM2Opcode        = "CALL_RECURSIVE"
	AVM2PromisePending          AVM2PromiseStatus = "pending"
	AVM2PromiseResolved         AVM2PromiseStatus = "resolved"
	AVM2PromiseRejected         AVM2PromiseStatus = "rejected"
	AVM2PromiseTimedOut         AVM2PromiseStatus = "timed_out"
	AVM2PromiseRefunded         AVM2PromiseStatus = "refunded"
	AVM2ProofRootZone           AVM2ProofRootType = "zone"
	AVM2ProofRootShard          AVM2ProofRootType = "shard"
	AVM2ProofRootMessage        AVM2ProofRootType = "message"
	AVM2ProofRootReceipt        AVM2ProofRootType = "receipt"
	AVM2ProofRootContractState  AVM2ProofRootType = "contract_state"
	AVM2ProofRootResolverRecord AVM2ProofRootType = "resolver_record"
)

type AVM2Opcode string
type AVM2PromiseStatus string
type AVM2ProofRootType string

type AVM2Limits struct {
	MaxInstructions       uint32
	MaxStackDepth         uint32
	MaxMemoryBytes        uint64
	MaxStorageKeyBytes    uint32
	MaxStorageValueBytes  uint64
	MaxOutputMessages     uint32
	MaxEvents             uint32
	MaxProofBytes         uint32
	MaxRecursionDepth     uint32
	MaxBoundedRangeItems  uint32
	MaxInstructionGas     uint64
	MaxABIEntries         uint32
	MaxEventPayloadBytes  uint32
	MaxPromiseStateWrites uint32
}

type AVM2GasTable struct {
	BaseInstructionGas uint64
	MemoryByteGas      uint64
	StorageReadGas     uint64
	StorageWriteGas    uint64
	ProofByteGas       uint64
	MessageCreateGas   uint64
	MessageByteGas     uint64
	EventByteGas       uint64
	ABIExportGas       uint64
	PromiseGas         uint64
	TableHash          string
}

type AVM2ExecutionContext struct {
	ChainID         string
	Height          uint64
	ZoneID          zonestypes.ZoneID
	ShardID         uint32
	ContractAddress string
	Caller          string
	ValueNAET       uint64
	GasLimit        uint64
	ReadOnly        bool
	ContextHash     string
}

type AVM2Instruction struct {
	Opcode      AVM2Opcode
	Key         string
	Value       []byte
	Message     AVMAsyncMessage
	Proof       AVM2ProofInput
	Promise     AVM2PromiseState
	ABI         AVM2ABIDescriptor
	Event       AVM2Event
	MemoryGrow  uint32
	RangeLimit  uint32
	GasOverride uint64
}

type AVM2Program struct {
	VMVersion             uint64
	InstructionSetVersion uint64
	Instructions          []AVM2Instruction
	MaxRecursionDepth     uint32
	ProgramHash           string
}

type AVM2StorageRead struct {
	Key     string
	KeyHash string
}

type AVM2StorageWrite struct {
	Key       string
	ValueHash string
	Deleted   bool
}

type AVM2ProofInput struct {
	ProofVersion uint64
	ChainID      string
	Height       uint64
	RootType     AVM2ProofRootType
	RootHash     string
	Key          string
	ValueHash    string
	ProofBytes   []byte
	ProofHash    string
}

type AVM2PromiseState struct {
	PromiseID     string
	Contract      string
	MessageID     string
	Status        AVM2PromiseStatus
	CreatedHeight uint64
	ExpiryHeight  uint64
	ReceiptHash   string
	ReturnHash    string
	PromiseHash   string
}

type AVM2ABIDescriptor struct {
	ABIVersion    uint64
	CodeID        uint64
	Methods       []string
	Events        []string
	Errors        []string
	RequiredFunds []string
	GasHints      []string
	InterfaceHash string
}

type AVM2Event struct {
	Height          uint64
	ContractAddress string
	EventID         string
	Name            string
	PayloadHash     string
	EventHash       string
}

type AVM2ExecutionResult struct {
	GasUsed            uint64
	MemoryBytes        uint64
	Stack              []string
	StorageReads       []AVM2StorageRead
	StorageWrites      []AVM2StorageWrite
	OutputMessages     []AVMAsyncMessage
	ProofsVerified     []AVM2ProofInput
	Promises           []AVM2PromiseState
	ABIDescriptors     []AVM2ABIDescriptor
	Events             []AVM2Event
	StorageRoot        string
	MessageRoot        string
	PromiseRoot        string
	ABIRoot            string
	EventRoot          string
	ExecutionHash      string
	ReadOnlySimulation bool
}

func DefaultAVM2Limits() AVM2Limits {
	return AVM2Limits{
		MaxInstructions:       4096,
		MaxStackDepth:         1024,
		MaxMemoryBytes:        1024 * 1024,
		MaxStorageKeyBytes:    256,
		MaxStorageValueBytes:  256 * 1024,
		MaxOutputMessages:     128,
		MaxEvents:             1024,
		MaxProofBytes:         64 * 1024,
		MaxRecursionDepth:     32,
		MaxBoundedRangeItems:  1024,
		MaxInstructionGas:     1_000_000,
		MaxABIEntries:         256,
		MaxEventPayloadBytes:  64 * 1024,
		MaxPromiseStateWrites: 256,
	}
}

func DefaultAVM2GasTable() (AVM2GasTable, error) {
	table := AVM2GasTable{
		BaseInstructionGas: 1,
		MemoryByteGas:      1,
		StorageReadGas:     2,
		StorageWriteGas:    4,
		ProofByteGas:       3,
		MessageCreateGas:   20,
		MessageByteGas:     1,
		EventByteGas:       1,
		ABIExportGas:       5,
		PromiseGas:         8,
	}
	table.TableHash = ComputeAVM2GasTableHash(table)
	return table, table.Validate()
}

func NewAVM2Program(program AVM2Program, limits AVM2Limits, gasTable AVM2GasTable) (AVM2Program, error) {
	program = canonicalAVM2Program(program)
	program.ProgramHash = ComputeAVM2ProgramHash(program)
	return program, ValidateAVM2Program(program, limits, gasTable)
}

func ExecuteAVM2Program(program AVM2Program, ctx AVM2ExecutionContext, limits AVM2Limits, gasTable AVM2GasTable) (AVM2ExecutionResult, error) {
	ctx = canonicalAVM2ExecutionContext(ctx)
	if err := ctx.Validate(); err != nil {
		return AVM2ExecutionResult{}, err
	}
	if err := ValidateAVM2Program(program, limits, gasTable); err != nil {
		return AVM2ExecutionResult{}, err
	}
	var result AVM2ExecutionResult
	result.ReadOnlySimulation = ctx.ReadOnly
	for _, instruction := range canonicalAVM2Program(program).Instructions {
		gas, err := AVM2InstructionGas(instruction, gasTable, limits)
		if err != nil {
			return AVM2ExecutionResult{}, err
		}
		result.GasUsed, err = checkedAVMGasAdd(result.GasUsed, gas)
		if err != nil {
			return AVM2ExecutionResult{}, err
		}
		if result.GasUsed > ctx.GasLimit {
			return AVM2ExecutionResult{}, errors.New("AVM 2.0 execution exhausted gas")
		}
		switch instruction.Opcode {
		case AVM2OpNoop:
		case AVM2OpPush:
			if len(result.Stack) >= int(limits.MaxStackDepth) {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 stack depth exceeded")
			}
			result.Stack = append(result.Stack, hex.EncodeToString(instruction.Value))
		case AVM2OpPop:
			if len(result.Stack) == 0 {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 stack underflow")
			}
			result.Stack = result.Stack[:len(result.Stack)-1]
		case AVM2OpMemGrow:
			result.MemoryBytes += uint64(instruction.MemoryGrow)
			if result.MemoryBytes > limits.MaxMemoryBytes {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 memory limit exceeded")
			}
		case AVM2OpKVGet, AVM2OpKVRangeBounded:
			if err := ValidateAVM2StoreV2Key(ctx, instruction.Key, limits); err != nil {
				return AVM2ExecutionResult{}, err
			}
			result.StorageReads = append(result.StorageReads, AVM2StorageRead{Key: instruction.Key, KeyHash: ComputeAVM2BytesHash([]byte(instruction.Key))})
		case AVM2OpKVSet:
			if ctx.ReadOnly {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot write storage")
			}
			if err := ValidateAVM2StoreV2Key(ctx, instruction.Key, limits); err != nil {
				return AVM2ExecutionResult{}, err
			}
			if uint64(len(instruction.Value)) > limits.MaxStorageValueBytes {
				return AVM2ExecutionResult{}, fmt.Errorf("AVM 2.0 storage value bytes must be <= %d", limits.MaxStorageValueBytes)
			}
			result.StorageWrites = append(result.StorageWrites, AVM2StorageWrite{Key: instruction.Key, ValueHash: ComputeAVM2BytesHash(instruction.Value)})
		case AVM2OpKVDelete:
			if ctx.ReadOnly {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot delete storage")
			}
			if err := ValidateAVM2StoreV2Key(ctx, instruction.Key, limits); err != nil {
				return AVM2ExecutionResult{}, err
			}
			result.StorageWrites = append(result.StorageWrites, AVM2StorageWrite{Key: instruction.Key, ValueHash: ComputeAVM2BytesHash(nil), Deleted: true})
		case AVM2OpVerifyMerkleProof:
			proof := canonicalAVM2ProofInput(instruction.Proof)
			if err := proof.Validate(limits); err != nil {
				return AVM2ExecutionResult{}, err
			}
			result.ProofsVerified = append(result.ProofsVerified, proof)
		case AVM2OpMsgSend:
			if ctx.ReadOnly {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot emit messages")
			}
			msg := canonicalAVMAsyncMessage(instruction.Message)
			if err := msg.Validate(); err != nil {
				return AVM2ExecutionResult{}, fmt.Errorf("AVM 2.0 output message: %w", err)
			}
			if len(result.OutputMessages) >= int(limits.MaxOutputMessages) {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 output message limit exceeded")
			}
			result.OutputMessages = append(result.OutputMessages, msg)
		case AVM2OpPromiseNew, AVM2OpPromiseResolve:
			if ctx.ReadOnly {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot mutate promise state")
			}
			promise := canonicalAVM2PromiseState(instruction.Promise)
			if err := promise.Validate(); err != nil {
				return AVM2ExecutionResult{}, err
			}
			if len(result.Promises) >= int(limits.MaxPromiseStateWrites) {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 promise write limit exceeded")
			}
			result.Promises = append(result.Promises, promise)
		case AVM2OpABIExport:
			abi := canonicalAVM2ABIDescriptor(instruction.ABI)
			if err := abi.Validate(limits); err != nil {
				return AVM2ExecutionResult{}, err
			}
			result.ABIDescriptors = append(result.ABIDescriptors, abi)
		case AVM2OpEmitEvent:
			if ctx.ReadOnly {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 read-only simulation cannot emit events")
			}
			event := canonicalAVM2Event(instruction.Event)
			if err := event.Validate(limits); err != nil {
				return AVM2ExecutionResult{}, err
			}
			if len(result.Events) >= int(limits.MaxEvents) {
				return AVM2ExecutionResult{}, errors.New("AVM 2.0 event limit exceeded")
			}
			result.Events = append(result.Events, event)
		default:
			return AVM2ExecutionResult{}, fmt.Errorf("unsupported AVM 2.0 opcode %q", instruction.Opcode)
		}
	}
	result = canonicalAVM2ExecutionResult(result)
	result.StorageRoot = ComputeAVM2StorageRoot(result.StorageReads, result.StorageWrites)
	result.MessageRoot = ComputeAVM2MessageRoot(result.OutputMessages)
	result.PromiseRoot = ComputeAVM2PromiseRoot(result.Promises)
	result.ABIRoot = ComputeAVM2ABIRoot(result.ABIDescriptors)
	result.EventRoot = ComputeAVM2EventRoot(result.Events)
	result.ExecutionHash = ComputeAVM2ExecutionHash(result)
	return result, result.Validate()
}

func ValidateAVM2Program(program AVM2Program, limits AVM2Limits, gasTable AVM2GasTable) error {
	program = canonicalAVM2Program(program)
	if err := limits.Validate(); err != nil {
		return err
	}
	if err := gasTable.Validate(); err != nil {
		return err
	}
	if program.VMVersion != AVM2VMVersion {
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
	if program.ProgramHash != ComputeAVM2ProgramHash(program) {
		return errors.New("AVM 2.0 program hash mismatch")
	}
	return nil
}

func (l AVM2Limits) Validate() error {
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

func (t AVM2GasTable) Validate() error {
	if t.BaseInstructionGas == 0 || t.MemoryByteGas == 0 || t.StorageReadGas == 0 || t.StorageWriteGas == 0 {
		return errors.New("AVM 2.0 gas table execution and storage gas must be positive")
	}
	if t.ProofByteGas == 0 || t.MessageCreateGas == 0 || t.MessageByteGas == 0 || t.EventByteGas == 0 || t.ABIExportGas == 0 || t.PromiseGas == 0 {
		return errors.New("AVM 2.0 gas table proof, message, event, ABI, and promise gas must be positive")
	}
	if t.TableHash == "" {
		return errors.New("AVM 2.0 gas table hash is required")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 gas table hash", t.TableHash); err != nil {
		return err
	}
	if t.TableHash != ComputeAVM2GasTableHash(t) {
		return errors.New("AVM 2.0 gas table hash mismatch")
	}
	return nil
}

func (c AVM2ExecutionContext) Validate() error {
	c = canonicalAVM2ExecutionContext(c)
	if err := validateEngineToken("AVM 2.0 chain id", c.ChainID, MaxAVM2TokenLength); err != nil {
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
	if err := validateEngineToken("AVM 2.0 caller", c.Caller, MaxAVM2TokenLength); err != nil {
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
	if c.ContextHash != ComputeAVM2ContextHash(c) {
		return errors.New("AVM 2.0 context hash mismatch")
	}
	return nil
}

func (i AVM2Instruction) Validate(limits AVM2Limits, gasTable AVM2GasTable) error {
	if IsAVM2ForbiddenOpcode(i.Opcode) {
		return fmt.Errorf("AVM 2.0 forbidden opcode %q", i.Opcode)
	}
	if !IsAVM2SupportedOpcode(i.Opcode) {
		return fmt.Errorf("unsupported AVM 2.0 opcode %q", i.Opcode)
	}
	gas, err := AVM2InstructionGas(i, gasTable, limits)
	if err != nil {
		return err
	}
	if gas == 0 || gas > limits.MaxInstructionGas {
		return fmt.Errorf("AVM 2.0 instruction gas must be in 1..%d", limits.MaxInstructionGas)
	}
	switch i.Opcode {
	case AVM2OpMemGrow:
		if i.MemoryGrow == 0 {
			return errors.New("AVM 2.0 memory growth must be positive")
		}
	case AVM2OpKVGet, AVM2OpKVSet, AVM2OpKVDelete, AVM2OpKVRangeBounded:
		if strings.TrimSpace(i.Key) == "" {
			return errors.New("AVM 2.0 storage key is required")
		}
		if uint32(len(i.Key)) > limits.MaxStorageKeyBytes {
			return fmt.Errorf("AVM 2.0 storage key bytes must be <= %d", limits.MaxStorageKeyBytes)
		}
		if i.Opcode == AVM2OpKVRangeBounded && (i.RangeLimit == 0 || i.RangeLimit > limits.MaxBoundedRangeItems) {
			return fmt.Errorf("AVM 2.0 bounded range limit must be in 1..%d", limits.MaxBoundedRangeItems)
		}
	case AVM2OpVerifyMerkleProof:
		return canonicalAVM2ProofInput(i.Proof).Validate(limits)
	case AVM2OpMsgSend:
		return canonicalAVMAsyncMessage(i.Message).Validate()
	case AVM2OpPromiseNew, AVM2OpPromiseResolve:
		return canonicalAVM2PromiseState(i.Promise).Validate()
	case AVM2OpABIExport:
		return canonicalAVM2ABIDescriptor(i.ABI).Validate(limits)
	case AVM2OpEmitEvent:
		return canonicalAVM2Event(i.Event).Validate(limits)
	}
	return nil
}

func (p AVM2ProofInput) Validate(limits AVM2Limits) error {
	p = canonicalAVM2ProofInput(p)
	if p.ProofVersion == 0 {
		return errors.New("AVM 2.0 proof version must be positive")
	}
	if err := validateEngineToken("AVM 2.0 proof chain id", p.ChainID, MaxAVM2TokenLength); err != nil {
		return err
	}
	if p.Height == 0 {
		return errors.New("AVM 2.0 proof height must be positive")
	}
	if !IsAVM2ProofRootType(p.RootType) {
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
	if p.ProofHash != ComputeAVM2ProofHash(p) {
		return errors.New("AVM 2.0 proof hash mismatch")
	}
	return nil
}

func (p AVM2PromiseState) Validate() error {
	p = canonicalAVM2PromiseState(p)
	if err := zonestypes.ValidateHash("AVM 2.0 promise id", p.PromiseID); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 promise contract", p.Contract, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise message id", p.MessageID); err != nil {
		return err
	}
	if !IsAVM2PromiseStatus(p.Status) {
		return fmt.Errorf("invalid AVM 2.0 promise status %q", p.Status)
	}
	if p.CreatedHeight == 0 || p.ExpiryHeight <= p.CreatedHeight {
		return errors.New("AVM 2.0 promise heights are invalid")
	}
	if p.Status != AVM2PromisePending {
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
	if p.PromiseHash != ComputeAVM2PromiseHash(p) {
		return errors.New("AVM 2.0 promise hash mismatch")
	}
	return nil
}

func (a AVM2ABIDescriptor) Validate(limits AVM2Limits) error {
	a = canonicalAVM2ABIDescriptor(a)
	if a.ABIVersion == 0 || a.CodeID == 0 {
		return errors.New("AVM 2.0 ABI version and code id must be positive")
	}
	for _, set := range []struct {
		name   string
		values []string
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
		if err := validateEngineTokens(set.name, set.values, MaxAVM2TokenLength); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI interface hash", a.InterfaceHash); err != nil {
		return err
	}
	if a.InterfaceHash != ComputeAVM2ABIInterfaceHash(a) {
		return errors.New("AVM 2.0 ABI interface hash mismatch")
	}
	return nil
}

func (e AVM2Event) Validate(limits AVM2Limits) error {
	e = canonicalAVM2Event(e)
	if e.Height == 0 {
		return errors.New("AVM 2.0 event height must be positive")
	}
	if err := validateEngineToken("AVM 2.0 event contract", e.ContractAddress, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 event id", e.EventID, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 event name", e.Name, MaxAVM2TokenLength); err != nil {
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
	if e.EventHash != ComputeAVM2EventHash(e) {
		return errors.New("AVM 2.0 event hash mismatch")
	}
	return nil
}

func (r AVM2ExecutionResult) Validate() error {
	r = canonicalAVM2ExecutionResult(r)
	for i, read := range r.StorageReads {
		if strings.TrimSpace(read.Key) == "" {
			return errors.New("AVM 2.0 storage read key is required")
		}
		if err := zonestypes.ValidateHash("AVM 2.0 storage read key hash", read.KeyHash); err != nil {
			return err
		}
		if read.KeyHash != ComputeAVM2BytesHash([]byte(read.Key)) {
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
		if err := proof.Validate(DefaultAVM2Limits()); err != nil {
			return err
		}
	}
	for _, promise := range r.Promises {
		if err := promise.Validate(); err != nil {
			return err
		}
	}
	for _, abi := range r.ABIDescriptors {
		if err := abi.Validate(DefaultAVM2Limits()); err != nil {
			return err
		}
	}
	for _, event := range r.Events {
		if err := event.Validate(DefaultAVM2Limits()); err != nil {
			return err
		}
	}
	for _, item := range []struct {
		name  string
		value string
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
	if r.StorageRoot != ComputeAVM2StorageRoot(r.StorageReads, r.StorageWrites) {
		return errors.New("AVM 2.0 storage root mismatch")
	}
	if r.MessageRoot != ComputeAVM2MessageRoot(r.OutputMessages) {
		return errors.New("AVM 2.0 message root mismatch")
	}
	if r.PromiseRoot != ComputeAVM2PromiseRoot(r.Promises) {
		return errors.New("AVM 2.0 promise root mismatch")
	}
	if r.ABIRoot != ComputeAVM2ABIRoot(r.ABIDescriptors) {
		return errors.New("AVM 2.0 ABI root mismatch")
	}
	if r.EventRoot != ComputeAVM2EventRoot(r.Events) {
		return errors.New("AVM 2.0 event root mismatch")
	}
	if r.ExecutionHash != ComputeAVM2ExecutionHash(r) {
		return errors.New("AVM 2.0 execution hash mismatch")
	}
	if r.ReadOnlySimulation && (len(r.StorageWrites) > 0 || len(r.OutputMessages) > 0 || len(r.Events) > 0 || len(r.Promises) > 0) {
		return errors.New("AVM 2.0 read-only simulation produced mutable outputs")
	}
	return nil
}

func ValidateAVM2StoreV2Key(ctx AVM2ExecutionContext, key string, limits AVM2Limits) error {
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

func AVM2InstructionGas(i AVM2Instruction, table AVM2GasTable, limits AVM2Limits) (uint64, error) {
	if i.GasOverride > 0 {
		return i.GasOverride, nil
	}
	gas := table.BaseInstructionGas
	var err error
	switch i.Opcode {
	case AVM2OpMemGrow:
		gas, err = checkedAVMGasAdd(gas, uint64(i.MemoryGrow)*table.MemoryByteGas)
	case AVM2OpKVGet, AVM2OpKVRangeBounded:
		gas, err = checkedAVMGasAdd(gas, table.StorageReadGas)
	case AVM2OpKVSet, AVM2OpKVDelete:
		gas, err = checkedAVMGasAdd(gas, table.StorageWriteGas)
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(len(i.Value))*table.StorageWriteGas)
		}
	case AVM2OpVerifyMerkleProof:
		gas, err = checkedAVMGasAdd(gas, uint64(len(i.Proof.ProofBytes))*table.ProofByteGas)
	case AVM2OpMsgSend:
		gas, err = checkedAVMGasAdd(gas, table.MessageCreateGas)
		if err == nil {
			gas, err = checkedAVMGasAdd(gas, uint64(len(i.Message.Payload))*table.MessageByteGas)
		}
	case AVM2OpPromiseNew, AVM2OpPromiseResolve:
		gas, err = checkedAVMGasAdd(gas, table.PromiseGas)
	case AVM2OpABIExport:
		gas, err = checkedAVMGasAdd(gas, table.ABIExportGas)
	case AVM2OpEmitEvent:
		gas, err = checkedAVMGasAdd(gas, table.EventByteGas)
	}
	if err != nil {
		return 0, err
	}
	if gas > limits.MaxInstructionGas {
		return 0, fmt.Errorf("AVM 2.0 instruction gas must be <= %d", limits.MaxInstructionGas)
	}
	return gas, nil
}

func IsAVM2SupportedOpcode(op AVM2Opcode) bool {
	switch op {
	case AVM2OpNoop, AVM2OpPush, AVM2OpPop, AVM2OpMemGrow, AVM2OpKVGet, AVM2OpKVSet, AVM2OpKVDelete, AVM2OpKVRangeBounded,
		AVM2OpVerifyMerkleProof, AVM2OpMsgSend, AVM2OpPromiseNew, AVM2OpPromiseResolve, AVM2OpABIExport, AVM2OpEmitEvent:
		return true
	default:
		return false
	}
}

func IsAVM2ForbiddenOpcode(op AVM2Opcode) bool {
	switch op {
	case AVM2OpExternalNetwork, AVM2OpWallClock, AVM2OpNonDeterministic, AVM2OpKVRangeUnbounded, AVM2OpDirectRemoteMutation, AVM2OpUnboundedRecursion:
		return true
	default:
		return false
	}
}

func IsAVM2PromiseStatus(status AVM2PromiseStatus) bool {
	switch status {
	case AVM2PromisePending, AVM2PromiseResolved, AVM2PromiseRejected, AVM2PromiseTimedOut, AVM2PromiseRefunded:
		return true
	default:
		return false
	}
}

func IsAVM2ProofRootType(rootType AVM2ProofRootType) bool {
	switch rootType {
	case AVM2ProofRootZone, AVM2ProofRootShard, AVM2ProofRootMessage, AVM2ProofRootReceipt, AVM2ProofRootContractState, AVM2ProofRootResolverRecord:
		return true
	default:
		return false
	}
}

func ComputeAVM2BytesHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func ComputeAVM2GasTableHash(table AVM2GasTable) string {
	table.TableHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-gas-table-v1")
	writeEngineUint64(h, table.BaseInstructionGas)
	writeEngineUint64(h, table.MemoryByteGas)
	writeEngineUint64(h, table.StorageReadGas)
	writeEngineUint64(h, table.StorageWriteGas)
	writeEngineUint64(h, table.ProofByteGas)
	writeEngineUint64(h, table.MessageCreateGas)
	writeEngineUint64(h, table.MessageByteGas)
	writeEngineUint64(h, table.EventByteGas)
	writeEngineUint64(h, table.ABIExportGas)
	writeEngineUint64(h, table.PromiseGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContextHash(ctx AVM2ExecutionContext) string {
	ctx.ContextHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-context-v1")
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

func ComputeAVM2ProgramHash(program AVM2Program) string {
	program = canonicalAVM2Program(program)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-program-v1")
	writeEngineUint64(h, program.VMVersion)
	writeEngineUint64(h, program.InstructionSetVersion)
	writeEngineUint64(h, uint64(program.MaxRecursionDepth))
	writeEngineUint64(h, uint64(len(program.Instructions)))
	for _, instruction := range program.Instructions {
		writeAVM2InstructionParts(h, instruction)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ProofHash(proof AVM2ProofInput) string {
	proof = canonicalAVM2ProofInput(proof)
	proof.ProofHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-proof-v1")
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

func ComputeAVM2PromiseHash(promise AVM2PromiseState) string {
	promise = canonicalAVM2PromiseState(promise)
	promise.PromiseHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-promise-v1")
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

func ComputeAVM2ABIInterfaceHash(abi AVM2ABIDescriptor) string {
	abi = canonicalAVM2ABIDescriptor(abi)
	abi.InterfaceHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-abi-v1")
	writeEngineUint64(h, abi.ABIVersion)
	writeEngineUint64(h, abi.CodeID)
	writeStringSet(h, abi.Methods)
	writeStringSet(h, abi.Events)
	writeStringSet(h, abi.Errors)
	writeStringSet(h, abi.RequiredFunds)
	writeStringSet(h, abi.GasHints)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2EventHash(event AVM2Event) string {
	event = canonicalAVM2Event(event)
	event.EventHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-event-v1")
	writeEngineUint64(h, event.Height)
	writeEnginePart(h, event.ContractAddress)
	writeEnginePart(h, event.EventID)
	writeEnginePart(h, event.Name)
	writeEnginePart(h, event.PayloadHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2StorageRoot(reads []AVM2StorageRead, writes []AVM2StorageWrite) string {
	reads = canonicalAVM2StorageReads(reads)
	writes = canonicalAVM2StorageWrites(writes)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-storage-root-v1")
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

func ComputeAVM2MessageRoot(messages []AVMAsyncMessage) string {
	messages = canonicalAVM2Messages(messages)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-message-root-v1")
	writeEngineUint64(h, uint64(len(messages)))
	for _, msg := range messages {
		writeAVMAsyncMessageParts(h, msg)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2PromiseRoot(promises []AVM2PromiseState) string {
	promises = canonicalAVM2Promises(promises)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-promise-root-v1")
	writeEngineUint64(h, uint64(len(promises)))
	for _, promise := range promises {
		writeEnginePart(h, promise.PromiseHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ABIRoot(abis []AVM2ABIDescriptor) string {
	abis = canonicalAVM2ABIs(abis)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-abi-root-v1")
	writeEngineUint64(h, uint64(len(abis)))
	for _, abi := range abis {
		writeEnginePart(h, abi.InterfaceHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2EventRoot(events []AVM2Event) string {
	events = canonicalAVM2Events(events)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-event-root-v1")
	writeEngineUint64(h, uint64(len(events)))
	for _, event := range events {
		writeEnginePart(h, event.EventHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ExecutionHash(result AVM2ExecutionResult) string {
	result = canonicalAVM2ExecutionResult(result)
	result.ExecutionHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-execution-v1")
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

func canonicalAVM2ExecutionContext(ctx AVM2ExecutionContext) AVM2ExecutionContext {
	ctx.ChainID = strings.TrimSpace(ctx.ChainID)
	ctx.ContractAddress = strings.TrimSpace(ctx.ContractAddress)
	ctx.Caller = strings.TrimSpace(ctx.Caller)
	ctx.ContextHash = strings.TrimSpace(ctx.ContextHash)
	return ctx
}

func canonicalAVM2Program(program AVM2Program) AVM2Program {
	program.Instructions = append([]AVM2Instruction(nil), program.Instructions...)
	for i := range program.Instructions {
		program.Instructions[i] = canonicalAVM2Instruction(program.Instructions[i])
	}
	program.ProgramHash = strings.TrimSpace(program.ProgramHash)
	return program
}

func canonicalAVM2Instruction(instruction AVM2Instruction) AVM2Instruction {
	instruction.Key = strings.TrimSpace(instruction.Key)
	instruction.Message = canonicalAVMAsyncMessage(instruction.Message)
	instruction.Proof = canonicalAVM2ProofInput(instruction.Proof)
	instruction.Promise = canonicalAVM2PromiseState(instruction.Promise)
	instruction.ABI = canonicalAVM2ABIDescriptor(instruction.ABI)
	instruction.Event = canonicalAVM2Event(instruction.Event)
	instruction.Value = append([]byte(nil), instruction.Value...)
	return instruction
}

func canonicalAVM2ProofInput(proof AVM2ProofInput) AVM2ProofInput {
	proof.ChainID = strings.TrimSpace(proof.ChainID)
	proof.Key = strings.TrimSpace(proof.Key)
	proof.RootHash = strings.TrimSpace(proof.RootHash)
	proof.ValueHash = strings.TrimSpace(proof.ValueHash)
	proof.ProofHash = strings.TrimSpace(proof.ProofHash)
	proof.ProofBytes = append([]byte(nil), proof.ProofBytes...)
	return proof
}

func canonicalAVM2PromiseState(promise AVM2PromiseState) AVM2PromiseState {
	promise.PromiseID = strings.TrimSpace(promise.PromiseID)
	promise.Contract = strings.TrimSpace(promise.Contract)
	promise.MessageID = strings.TrimSpace(promise.MessageID)
	promise.ReceiptHash = strings.TrimSpace(promise.ReceiptHash)
	promise.ReturnHash = strings.TrimSpace(promise.ReturnHash)
	promise.PromiseHash = strings.TrimSpace(promise.PromiseHash)
	return promise
}

func canonicalAVM2ABIDescriptor(abi AVM2ABIDescriptor) AVM2ABIDescriptor {
	abi.Methods = cloneSortedStrings(abi.Methods)
	abi.Events = cloneSortedStrings(abi.Events)
	abi.Errors = cloneSortedStrings(abi.Errors)
	abi.RequiredFunds = cloneSortedStrings(abi.RequiredFunds)
	abi.GasHints = cloneSortedStrings(abi.GasHints)
	abi.InterfaceHash = strings.TrimSpace(abi.InterfaceHash)
	return abi
}

func canonicalAVM2Event(event AVM2Event) AVM2Event {
	event.ContractAddress = strings.TrimSpace(event.ContractAddress)
	event.EventID = strings.TrimSpace(event.EventID)
	event.Name = strings.TrimSpace(event.Name)
	event.PayloadHash = strings.TrimSpace(event.PayloadHash)
	event.EventHash = strings.TrimSpace(event.EventHash)
	return event
}

func canonicalAVM2ExecutionResult(result AVM2ExecutionResult) AVM2ExecutionResult {
	result.Stack = append([]string(nil), result.Stack...)
	for i := range result.Stack {
		result.Stack[i] = strings.TrimSpace(result.Stack[i])
	}
	result.StorageReads = canonicalAVM2StorageReads(result.StorageReads)
	result.StorageWrites = canonicalAVM2StorageWrites(result.StorageWrites)
	result.OutputMessages = canonicalAVM2Messages(result.OutputMessages)
	result.ProofsVerified = canonicalAVM2Proofs(result.ProofsVerified)
	result.Promises = canonicalAVM2Promises(result.Promises)
	result.ABIDescriptors = canonicalAVM2ABIs(result.ABIDescriptors)
	result.Events = canonicalAVM2Events(result.Events)
	result.StorageRoot = strings.TrimSpace(result.StorageRoot)
	result.MessageRoot = strings.TrimSpace(result.MessageRoot)
	result.PromiseRoot = strings.TrimSpace(result.PromiseRoot)
	result.ABIRoot = strings.TrimSpace(result.ABIRoot)
	result.EventRoot = strings.TrimSpace(result.EventRoot)
	result.ExecutionHash = strings.TrimSpace(result.ExecutionHash)
	return result
}

func canonicalAVM2StorageReads(reads []AVM2StorageRead) []AVM2StorageRead {
	out := append([]AVM2StorageRead(nil), reads...)
	for i := range out {
		out[i].Key = strings.TrimSpace(out[i].Key)
		out[i].KeyHash = strings.TrimSpace(out[i].KeyHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func canonicalAVM2StorageWrites(writes []AVM2StorageWrite) []AVM2StorageWrite {
	out := append([]AVM2StorageWrite(nil), writes...)
	for i := range out {
		out[i].Key = strings.TrimSpace(out[i].Key)
		out[i].ValueHash = strings.TrimSpace(out[i].ValueHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func canonicalAVM2Messages(messages []AVMAsyncMessage) []AVMAsyncMessage {
	out := append([]AVMAsyncMessage(nil), messages...)
	for i := range out {
		out[i] = canonicalAVMAsyncMessage(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func canonicalAVM2Proofs(proofs []AVM2ProofInput) []AVM2ProofInput {
	out := append([]AVM2ProofInput(nil), proofs...)
	for i := range out {
		out[i] = canonicalAVM2ProofInput(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ProofHash < out[j].ProofHash })
	return out
}

func canonicalAVM2Promises(promises []AVM2PromiseState) []AVM2PromiseState {
	out := append([]AVM2PromiseState(nil), promises...)
	for i := range out {
		out[i] = canonicalAVM2PromiseState(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PromiseID < out[j].PromiseID })
	return out
}

func canonicalAVM2ABIs(abis []AVM2ABIDescriptor) []AVM2ABIDescriptor {
	out := append([]AVM2ABIDescriptor(nil), abis...)
	for i := range out {
		out[i] = canonicalAVM2ABIDescriptor(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CodeID == out[j].CodeID {
			return out[i].ABIVersion < out[j].ABIVersion
		}
		return out[i].CodeID < out[j].CodeID
	})
	return out
}

func canonicalAVM2Events(events []AVM2Event) []AVM2Event {
	out := append([]AVM2Event(nil), events...)
	for i := range out {
		out[i] = canonicalAVM2Event(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height == out[j].Height {
			return out[i].EventID < out[j].EventID
		}
		return out[i].Height < out[j].Height
	})
	return out
}

func writeAVM2InstructionParts(h engineByteWriter, instruction AVM2Instruction) {
	instruction = canonicalAVM2Instruction(instruction)
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
