package avm

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

const (
	Magic		= "AVM1"
	Version	uint16	= 1

	MetadataHashLength	= 32
	MaxKeySize		= 128

	EntryDeploy		Entrypoint	= 1
	EntryReceiveExternal	Entrypoint	= 2
	EntryReceiveInternal	Entrypoint	= 3
	EntryReceiveBounced	Entrypoint	= 4
	EntryQuery		Entrypoint	= 5
	EntryMigrate		Entrypoint	= 6

	HostReadStorage		HostFunction	= 1
	HostWriteStorage	HostFunction	= 2
	HostEmitInternal	HostFunction	= 3
	HostInspectMsg		HostFunction	= 4
	HostBlockContext	HostFunction	= 5
	HostChargeGas		HostFunction	= 6
	HostReturn		HostFunction	= 7
	HostScheduleSelf	HostFunction	= 8

	OpNop			Opcode	= 0x00
	OpPushU64		Opcode	= 0x01
	OpReadStorage		Opcode	= 0x02
	OpWriteStorage		Opcode	= 0x03
	OpAdd			Opcode	= 0x04
	OpEmitInternal		Opcode	= 0x05
	OpReturn		Opcode	= 0x06
	OpReadMsgOpcode		Opcode	= 0x07
	OpReadMsgQueryID	Opcode	= 0x08
	OpReadBlock		Opcode	= 0x09
	OpChargeGas		Opcode	= 0x0a
	OpScheduleSelf		Opcode	= 0x0b

	OpWallClock	Opcode	= 0xf0
	OpRandom	Opcode	= 0xf1
	OpFileRead	Opcode	= 0xf2
	OpFloatAdd	Opcode	= 0xf3
	OpIterMap	Opcode	= 0xf4
)

type Entrypoint uint8
type HostFunction uint16
type Opcode uint8

type Params struct {
	MaxCodeBytes	uint32
	MaxInstructions	uint32
	MaxImports	uint16
	MaxStackDepth	uint32
	MaxMemoryBytes	uint32
	GasSchedule	map[Opcode]uint64
}

type Module struct {
	Version		uint16
	Imports		[]HostFunction
	Exports		map[Entrypoint]uint32
	MetadataHash	[MetadataHashLength]byte
	Code		[]Instruction
}

type Instruction struct {
	Op	Opcode
	Arg	uint64
	Data	[]byte
}

type RuntimeContext struct {
	Entry		Entrypoint
	ContractAddress	sdk.AccAddress
	Message		async.MessageEnvelope
	BlockHeight	uint64
	GasLimit	uint64
	EmitDestination	sdk.AccAddress
}

type Execution struct {
	State		Storage
	Outgoing	[]async.MessageEnvelope
	GasUsed		uint64
	ResultCode	uint32
	StorageWrites	uint32
	ReturnValue	uint64
	ExecutedOpcode	[]Opcode
}

type Storage map[string][]byte

type SnapshotEntry struct {
	Key	string
	Value	[]byte
}

type ExecutionProof struct {
	ModuleHash	[32]byte
	BeforeRoot	[32]byte
	AfterRoot	[32]byte
	ContextHash	[32]byte
	OutgoingRoot	[32]byte
	TraceHash	[32]byte
	GasUsed		uint64
	ResultCode	uint32
	StorageWrites	uint32
	ReturnValue	uint64
}

type Verifier struct {
	params Params
}

type Runner struct {
	params Params
}

func DefaultParams() Params {
	return Params{
		MaxCodeBytes:		64 * 1024,
		MaxInstructions:	4096,
		MaxImports:		32,
		MaxStackDepth:		1024,
		MaxMemoryBytes:		1024 * 1024,
		GasSchedule: map[Opcode]uint64{
			OpNop:			1,
			OpPushU64:		2,
			OpReadStorage:		20,
			OpWriteStorage:		50,
			OpAdd:			3,
			OpEmitInternal:		100,
			OpReturn:		1,
			OpReadMsgOpcode:	5,
			OpReadMsgQueryID:	5,
			OpReadBlock:		5,
			OpChargeGas:		1,
			OpScheduleSelf:		100,
		},
	}
}

func NewVerifier(params Params) (*Verifier, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &Verifier{params: params}, nil
}

func NewRunner(params Params) (*Runner, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &Runner{params: params}, nil
}

func (p Params) Validate() error {
	if p.MaxCodeBytes == 0 {
		return errors.New("max code bytes must be positive")
	}
	if p.MaxInstructions == 0 {
		return errors.New("max instructions must be positive")
	}
	if p.MaxImports == 0 {
		return errors.New("max imports must be positive")
	}
	if p.MaxStackDepth == 0 {
		return errors.New("max stack depth must be positive")
	}
	if p.MaxMemoryBytes == 0 {
		return errors.New("max memory bytes must be positive")
	}
	for _, op := range []Opcode{
		OpNop,
		OpPushU64,
		OpReadStorage,
		OpWriteStorage,
		OpAdd,
		OpEmitInternal,
		OpReturn,
		OpReadMsgOpcode,
		OpReadMsgQueryID,
		OpReadBlock,
		OpChargeGas,
		OpScheduleSelf,
	} {
		if p.GasSchedule[op] == 0 {
			return fmt.Errorf("gas schedule missing opcode 0x%02x", byte(op))
		}
	}
	return nil
}

func (v *Verifier) Verify(module Module) error {
	if module.Version != Version {
		return fmt.Errorf("unsupported AVM version %d", module.Version)
	}
	if len(module.Code) == 0 {
		return errors.New("AVM module code must not be empty")
	}
	if len(module.Code) > int(v.params.MaxInstructions) {
		return fmt.Errorf("AVM instruction count must be <= %d", v.params.MaxInstructions)
	}
	encoded, err := EncodeModule(module)
	if err != nil {
		return err
	}
	if len(encoded) > int(v.params.MaxCodeBytes) {
		return fmt.Errorf("AVM code bytes must be <= %d", v.params.MaxCodeBytes)
	}
	if len(module.Imports) > int(v.params.MaxImports) {
		return fmt.Errorf("AVM import count must be <= %d", v.params.MaxImports)
	}
	if len(module.Exports) == 0 {
		return errors.New("AVM module must export at least one entrypoint")
	}
	for _, host := range module.Imports {
		if !IsAllowedHostFunction(host) {
			return fmt.Errorf("AVM host function %d is not allowed", host)
		}
	}
	imports := hostImportSet(module.Imports)
	for entry, offset := range module.Exports {
		if !IsValidEntrypoint(entry) {
			return fmt.Errorf("AVM entrypoint %d is invalid", entry)
		}
		if offset >= uint32(len(module.Code)) {
			return fmt.Errorf("AVM entrypoint %d offset out of range", entry)
		}
	}
	for _, ins := range module.Code {
		if IsForbiddenOpcode(ins.Op) {
			return fmt.Errorf("AVM opcode 0x%02x is nondeterministic or forbidden", byte(ins.Op))
		}
		if !IsAllowedOpcode(ins.Op) {
			return fmt.Errorf("AVM opcode 0x%02x is unknown", byte(ins.Op))
		}
		if required, ok := RequiredHostFunction(ins.Op); ok {
			if _, imported := imports[required]; !imported {
				return fmt.Errorf("AVM opcode 0x%02x requires host function %d", byte(ins.Op), required)
			}
		}
		if len(ins.Data) > MaxKeySize {
			return fmt.Errorf("AVM instruction data must be <= %d bytes", MaxKeySize)
		}
	}
	return nil
}

func (r *Runner) Run(module Module, storage Storage, ctx RuntimeContext) (Execution, error) {
	verifier, err := NewVerifier(r.params)
	if err != nil {
		return Execution{}, err
	}
	if err := verifier.Verify(module); err != nil {
		return Execution{}, err
	}
	pc, ok := module.Exports[ctx.Entry]
	if !ok {
		return Execution{}, fmt.Errorf("AVM entrypoint %d is not exported", ctx.Entry)
	}
	state := CloneStorage(storage)
	stack := make([]uint64, 0)
	exec := Execution{State: state}
	gasLimit := ctx.GasLimit
	if gasLimit == 0 {
		gasLimit = ctx.Message.GasLimit
	}
	if gasLimit == 0 {
		return Execution{}, errors.New("AVM gas limit must be positive")
	}

	for ; int(pc) < len(module.Code); pc++ {
		ins := module.Code[pc]
		gas := r.params.GasSchedule[ins.Op]
		nextGas, overflow := safeAddU64(exec.GasUsed, gas)
		if overflow {
			exec.ResultCode = async.ResultLimitExceeded
			return exec, nil
		}
		exec.GasUsed = nextGas
		if exec.GasUsed > gasLimit {
			exec.ResultCode = async.ResultLimitExceeded
			return exec, nil
		}
		exec.ExecutedOpcode = append(exec.ExecutedOpcode, ins.Op)
		switch ins.Op {
		case OpNop:
		case OpPushU64:
			if len(stack) >= int(r.params.MaxStackDepth) {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
			stack = append(stack, ins.Arg)
		case OpReadStorage:
			value := DecodeU64(state[string(ins.Data)])
			if len(stack) >= int(r.params.MaxStackDepth) {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
			stack = append(stack, value)
		case OpWriteStorage:
			value, ok := pop(&stack)
			if !ok {
				return exec, errors.New("AVM stack underflow on write storage")
			}
			state[string(ins.Data)] = EncodeU64(value)
			exec.StorageWrites++
			if StorageMemoryBytes(state) > uint64(r.params.MaxMemoryBytes) {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
		case OpAdd:
			right, ok := pop(&stack)
			if !ok {
				return exec, errors.New("AVM stack underflow on add")
			}
			left, ok := pop(&stack)
			if !ok {
				return exec, errors.New("AVM stack underflow on add")
			}
			stack = append(stack, left+right)
		case OpEmitInternal:
			if len(ctx.EmitDestination) == 0 {
				return exec, errors.New("AVM emit internal requires destination")
			}
			exec.Outgoing = append(exec.Outgoing, async.MessageEnvelope{
				Destination:	ctx.EmitDestination,
				Value:		sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
				Opcode:		uint32(ins.Arg),
				QueryID:	ctx.Message.QueryID,
				Body:		append([]byte(nil), ins.Data...),
				Bounce:		true,
				GasLimit:	ctx.Message.GasLimit,
				ForwardFee:	sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
			})
		case OpReadMsgOpcode:
			if len(stack) >= int(r.params.MaxStackDepth) {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
			stack = append(stack, uint64(ctx.Message.Opcode))
		case OpReadMsgQueryID:
			if len(stack) >= int(r.params.MaxStackDepth) {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
			stack = append(stack, ctx.Message.QueryID)
		case OpReadBlock:
			if len(stack) >= int(r.params.MaxStackDepth) {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
			stack = append(stack, ctx.BlockHeight)
		case OpChargeGas:
			nextGas, overflow := safeAddU64(exec.GasUsed, ins.Arg)
			if overflow || nextGas > gasLimit {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
			exec.GasUsed = nextGas
		case OpScheduleSelf:
			if len(ctx.ContractAddress) == 0 {
				return exec, errors.New("AVM schedule self requires contract address")
			}
			if ctx.BlockHeight == 0 {
				return exec, errors.New("AVM schedule self requires block height")
			}
			if ins.Arg == 0 {
				return exec, errors.New("AVM schedule self delay must be positive")
			}
			deliverAt, overflow := safeAddU64(ctx.BlockHeight, ins.Arg)
			if overflow {
				exec.ResultCode = async.ResultLimitExceeded
				return exec, nil
			}
			exec.Outgoing = append(exec.Outgoing, async.MessageEnvelope{
				Destination:	append(sdk.AccAddress(nil), ctx.ContractAddress...),
				Value:		sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
				Opcode:		ctx.Message.Opcode,
				QueryID:	ctx.Message.QueryID,
				Body:		append([]byte(nil), ins.Data...),
				Bounce:		false,
				DeliverAtBlock:	deliverAt,
				DeadlineBlock:	ctx.Message.DeadlineBlock,
				GasLimit:	ctx.Message.GasLimit,
				ForwardFee:	sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
			})
		case OpReturn:
			exec.ResultCode = uint32(ins.Arg)
			if len(stack) > 0 {
				exec.ReturnValue = stack[len(stack)-1]
			}
			return exec, nil
		default:
			return exec, fmt.Errorf("AVM opcode 0x%02x is not executable", byte(ins.Op))
		}
	}
	exec.ResultCode = async.ResultOK
	if len(stack) > 0 {
		exec.ReturnValue = stack[len(stack)-1]
	}
	return exec, nil
}

func (r *Runner) AsyncHandler(module Module, storage Storage, ctx RuntimeContext) async.Handler {
	return func(contract async.ContractAccount, msg async.MessageEnvelope) async.ExecutionResult {
		baseStorage := CloneStorage(storage)
		if storage == nil && len(contract.State) > 0 {
			decoded, err := DecodeSnapshot(contract.State)
			if err != nil {
				return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
			}
			baseStorage = decoded
		}
		ctx.ContractAddress = contract.Address
		ctx.Message = msg
		if msg.ExecutionBlockHeight != 0 {
			ctx.BlockHeight = msg.ExecutionBlockHeight
		}
		if msg.Bounced {
			ctx.Entry = EntryReceiveBounced
		} else {
			ctx.Entry = EntryReceiveInternal
		}
		ctx.GasLimit = msg.GasLimit
		exec, err := r.Run(module, baseStorage, ctx)
		if err != nil {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
		}
		snapshot := EncodeSnapshot(exec.State)
		return async.ExecutionResult{
			NewState:	snapshot,
			Outgoing:	exec.Outgoing,
			GasUsed:	exec.GasUsed,
			StorageWrites:	exec.StorageWrites,
			ResultCode:	exec.ResultCode,
		}
	}
}

func EncodeModule(module Module) ([]byte, error) {
	if len(module.Code) == 0 {
		return nil, errors.New("AVM module code must not be empty")
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteString(Magic)
	writeU16(buf, module.Version)
	buf.Write(module.MetadataHash[:])
	writeU16(buf, uint16(len(module.Imports)))
	for _, host := range module.Imports {
		writeU16(buf, uint16(host))
	}
	writeU16(buf, uint16(len(module.Exports)))
	entries := make([]int, 0, len(module.Exports))
	for entry := range module.Exports {
		entries = append(entries, int(entry))
	}
	sort.Ints(entries)
	for _, raw := range entries {
		entry := Entrypoint(raw)
		buf.WriteByte(byte(entry))
		writeU32(buf, module.Exports[entry])
	}
	writeU32(buf, uint32(len(module.Code)))
	for _, ins := range module.Code {
		buf.WriteByte(byte(ins.Op))
		writeU64(buf, ins.Arg)
		if len(ins.Data) > MaxKeySize {
			return nil, fmt.Errorf("AVM instruction data must be <= %d bytes", MaxKeySize)
		}
		writeU16(buf, uint16(len(ins.Data)))
		buf.Write(ins.Data)
	}
	return buf.Bytes(), nil
}

func DecodeModule(bz []byte) (Module, error) {
	if len(bz) < 4+2+MetadataHashLength {
		return Module{}, errors.New("AVM bytecode is malformed")
	}
	reader := bytes.NewReader(bz)
	magic := make([]byte, 4)
	if _, err := reader.Read(magic); err != nil {
		return Module{}, err
	}
	if string(magic) != Magic {
		return Module{}, errors.New("AVM bytecode has invalid module header")
	}
	version, err := readU16(reader)
	if err != nil {
		return Module{}, err
	}
	var metadata [MetadataHashLength]byte
	if _, err := reader.Read(metadata[:]); err != nil {
		return Module{}, err
	}
	importCount, err := readU16(reader)
	if err != nil {
		return Module{}, err
	}
	imports := make([]HostFunction, importCount)
	for i := range imports {
		value, err := readU16(reader)
		if err != nil {
			return Module{}, err
		}
		imports[i] = HostFunction(value)
	}
	exportCount, err := readU16(reader)
	if err != nil {
		return Module{}, err
	}
	exports := make(map[Entrypoint]uint32, exportCount)
	for i := uint16(0); i < exportCount; i++ {
		entry, err := reader.ReadByte()
		if err != nil {
			return Module{}, err
		}
		offset, err := readU32(reader)
		if err != nil {
			return Module{}, err
		}
		exports[Entrypoint(entry)] = offset
	}
	codeCount, err := readU32(reader)
	if err != nil {
		return Module{}, err
	}
	code := make([]Instruction, codeCount)
	for i := range code {
		op, err := reader.ReadByte()
		if err != nil {
			return Module{}, err
		}
		arg, err := readU64(reader)
		if err != nil {
			return Module{}, err
		}
		dataLen, err := readU16(reader)
		if err != nil {
			return Module{}, err
		}
		var data []byte
		if dataLen > 0 {
			data = make([]byte, dataLen)
			if _, err := io.ReadFull(reader, data); err != nil {
				return Module{}, err
			}
		}
		code[i] = Instruction{Op: Opcode(op), Arg: arg, Data: data}
	}
	if reader.Len() != 0 {
		return Module{}, errors.New("AVM bytecode has trailing data")
	}
	return Module{Version: version, Imports: imports, Exports: exports, MetadataHash: metadata, Code: code}, nil
}

func CodeHash(module Module) ([32]byte, error) {
	encoded, err := EncodeModule(module)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(encoded), nil
}

func IsAllowedHostFunction(host HostFunction) bool {
	switch host {
	case HostReadStorage, HostWriteStorage, HostEmitInternal, HostInspectMsg, HostBlockContext, HostChargeGas, HostReturn, HostScheduleSelf:
		return true
	default:
		return false
	}
}

func IsValidEntrypoint(entry Entrypoint) bool {
	switch entry {
	case EntryDeploy, EntryReceiveExternal, EntryReceiveInternal, EntryReceiveBounced, EntryQuery, EntryMigrate:
		return true
	default:
		return false
	}
}

func IsForbiddenOpcode(op Opcode) bool {
	switch op {
	case OpWallClock, OpRandom, OpFileRead, OpFloatAdd, OpIterMap:
		return true
	default:
		return false
	}
}

func IsAllowedOpcode(op Opcode) bool {
	switch op {
	case OpNop,
		OpPushU64,
		OpReadStorage,
		OpWriteStorage,
		OpAdd,
		OpEmitInternal,
		OpReturn,
		OpReadMsgOpcode,
		OpReadMsgQueryID,
		OpReadBlock,
		OpChargeGas,
		OpScheduleSelf:
		return true
	default:
		return false
	}
}

func RequiredHostFunction(op Opcode) (HostFunction, bool) {
	switch op {
	case OpReadStorage:
		return HostReadStorage, true
	case OpWriteStorage:
		return HostWriteStorage, true
	case OpEmitInternal:
		return HostEmitInternal, true
	case OpReturn:
		return HostReturn, true
	case OpReadMsgOpcode, OpReadMsgQueryID:
		return HostInspectMsg, true
	case OpReadBlock:
		return HostBlockContext, true
	case OpChargeGas:
		return HostChargeGas, true
	case OpScheduleSelf:
		return HostScheduleSelf, true
	default:
		return 0, false
	}
}

func hostImportSet(imports []HostFunction) map[HostFunction]struct{} {
	out := make(map[HostFunction]struct{}, len(imports))
	for _, host := range imports {
		out[host] = struct{}{}
	}
	return out
}

func EncodeU64(value uint64) []byte {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	return out[:]
}

func DecodeU64(bz []byte) uint64 {
	if len(bz) != 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func CloneStorage(storage Storage) Storage {
	out := make(Storage, len(storage))
	for key, value := range storage {
		out[key] = append([]byte(nil), value...)
	}
	return out
}

func StorageMemoryBytes(storage Storage) uint64 {
	var total uint64
	for key, value := range storage {
		total += uint64(len(key) + len(value))
	}
	return total
}

func Snapshot(storage Storage) []SnapshotEntry {
	keys := make([]string, 0, len(storage))
	for key := range storage {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	entries := make([]SnapshotEntry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, SnapshotEntry{Key: key, Value: append([]byte(nil), storage[key]...)})
	}
	return entries
}

func EncodeSnapshot(storage Storage) []byte {
	entries := Snapshot(storage)
	buf := bytes.NewBuffer(nil)
	writeU32(buf, uint32(len(entries)))
	for _, entry := range entries {
		writeU16(buf, uint16(len(entry.Key)))
		buf.WriteString(entry.Key)
		writeU32(buf, uint32(len(entry.Value)))
		buf.Write(entry.Value)
	}
	return buf.Bytes()
}

func DecodeSnapshot(bz []byte) (Storage, error) {
	if len(bz) == 0 {
		return nil, errors.New("AVM snapshot is empty")
	}
	reader := bytes.NewReader(bz)
	count, err := readU32(reader)
	if err != nil {
		return nil, err
	}
	storage := make(Storage, count)
	var previous string
	for i := uint32(0); i < count; i++ {
		keyLen, err := readU16(reader)
		if err != nil {
			return nil, err
		}
		if keyLen == 0 {
			return nil, errors.New("AVM snapshot key must not be empty")
		}
		if keyLen > MaxKeySize {
			return nil, fmt.Errorf("AVM snapshot key must be <= %d bytes", MaxKeySize)
		}
		keyBytes := make([]byte, keyLen)
		if _, err := io.ReadFull(reader, keyBytes); err != nil {
			return nil, err
		}
		key := string(keyBytes)
		if i > 0 && previous >= key {
			return nil, errors.New("AVM snapshot keys must be sorted and unique")
		}
		valueLen, err := readU32(reader)
		if err != nil {
			return nil, err
		}
		value := make([]byte, valueLen)
		if valueLen > 0 {
			if _, err := io.ReadFull(reader, value); err != nil {
				return nil, err
			}
		}
		storage[key] = value
		previous = key
	}
	if reader.Len() != 0 {
		return nil, errors.New("AVM snapshot has trailing data")
	}
	return storage, nil
}

func StorageRoot(storage Storage) [32]byte {
	return sha256.Sum256(EncodeSnapshot(storage))
}

func BuildExecutionProof(module Module, before Storage, ctx RuntimeContext, exec Execution) (ExecutionProof, error) {
	moduleHash, err := CodeHash(module)
	if err != nil {
		return ExecutionProof{}, err
	}
	return ExecutionProof{
		ModuleHash:	moduleHash,
		BeforeRoot:	StorageRoot(before),
		AfterRoot:	StorageRoot(exec.State),
		ContextHash:	RuntimeContextHash(ctx),
		OutgoingRoot:	OutgoingMessagesRoot(exec.Outgoing),
		TraceHash:	OpcodeTraceHash(exec.ExecutedOpcode),
		GasUsed:	exec.GasUsed,
		ResultCode:	exec.ResultCode,
		StorageWrites:	exec.StorageWrites,
		ReturnValue:	exec.ReturnValue,
	}, nil
}

func ExecutionProofHash(proof ExecutionProof) [32]byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(proof.ModuleHash[:])
	buf.Write(proof.BeforeRoot[:])
	buf.Write(proof.AfterRoot[:])
	buf.Write(proof.ContextHash[:])
	buf.Write(proof.OutgoingRoot[:])
	buf.Write(proof.TraceHash[:])
	writeU64(buf, proof.GasUsed)
	writeU32(buf, proof.ResultCode)
	writeU32(buf, proof.StorageWrites)
	writeU64(buf, proof.ReturnValue)
	return sha256.Sum256(buf.Bytes())
}

func RuntimeContextHash(ctx RuntimeContext) [32]byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(ctx.Entry))
	writeBytes(buf, ctx.ContractAddress)
	writeMessageEnvelope(buf, ctx.Message)
	writeU64(buf, ctx.BlockHeight)
	writeU64(buf, ctx.GasLimit)
	writeBytes(buf, ctx.EmitDestination)
	return sha256.Sum256(buf.Bytes())
}

func OutgoingMessagesRoot(messages []async.MessageEnvelope) [32]byte {
	buf := bytes.NewBuffer(nil)
	writeU32(buf, uint32(len(messages)))
	for _, msg := range messages {
		writeMessageEnvelope(buf, msg)
	}
	return sha256.Sum256(buf.Bytes())
}

func OpcodeTraceHash(trace []Opcode) [32]byte {
	buf := bytes.NewBuffer(nil)
	writeU32(buf, uint32(len(trace)))
	for _, op := range trace {
		buf.WriteByte(byte(op))
	}
	return sha256.Sum256(buf.Bytes())
}

func pop(stack *[]uint64) (uint64, bool) {
	if len(*stack) == 0 {
		return 0, false
	}
	last := len(*stack) - 1
	value := (*stack)[last]
	*stack = (*stack)[:last]
	return value, true
}

func writeU16(buf *bytes.Buffer, value uint16) {
	var out [2]byte
	binary.BigEndian.PutUint16(out[:], value)
	buf.Write(out[:])
}

func writeU32(buf *bytes.Buffer, value uint32) {
	var out [4]byte
	binary.BigEndian.PutUint32(out[:], value)
	buf.Write(out[:])
}

func writeU64(buf *bytes.Buffer, value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	buf.Write(out[:])
}

func writeBytes(buf *bytes.Buffer, value []byte) {
	writeU32(buf, uint32(len(value)))
	buf.Write(value)
}

func writeString(buf *bytes.Buffer, value string) {
	writeBytes(buf, []byte(value))
}

func writeMessageEnvelope(buf *bytes.Buffer, msg async.MessageEnvelope) {
	writeBytes(buf, msg.Source)
	writeBytes(buf, msg.Destination)
	writeString(buf, msg.Value.Denom)
	writeString(buf, msg.Value.Amount.String())
	writeU32(buf, msg.Opcode)
	writeU64(buf, msg.QueryID)
	writeBytes(buf, msg.Body)
	if msg.Bounce {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	if msg.Bounced {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	writeU64(buf, msg.CreatedLogicalTime)
	writeU64(buf, msg.DeliverAtBlock)
	writeU32(buf, msg.RetryCount)
	writeU32(buf, msg.MaxRetries)
	writeU64(buf, msg.RetryDelayBlocks)
	writeU64(buf, msg.DeadlineBlock)
	writeU64(buf, msg.GasLimit)
	writeString(buf, msg.ForwardFee.Denom)
	writeString(buf, msg.ForwardFee.Amount.String())
	writeU32(buf, msg.Depth)
}

func safeAddU64(left, right uint64) (uint64, bool) {
	if right > ^uint64(0)-left {
		return 0, true
	}
	return left + right, false
}

func readU16(reader *bytes.Reader) (uint16, error) {
	var in [2]byte
	if _, err := io.ReadFull(reader, in[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(in[:]), nil
}

func readU32(reader *bytes.Reader) (uint32, error) {
	var in [4]byte
	if _, err := io.ReadFull(reader, in[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(in[:]), nil
}

func readU64(reader *bytes.Reader) (uint64, error) {
	var in [8]byte
	if _, err := io.ReadFull(reader, in[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(in[:]), nil
}
