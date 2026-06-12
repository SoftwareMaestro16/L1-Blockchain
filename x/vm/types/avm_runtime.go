package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMBytecodeMagic	= "AVM"
	AVMBytecodeCodec	= "AVM.bytecode.v1"
	AVMRuntimeExecutor	= "AVM-runtime"
	AVMRuntimeErrorAbort	= "ERR_AVM_ABORT"
	AVMRuntimeErrorGas	= "ERR_AVM_GAS"
	AVMRuntimeErrorInvalid	= "ERR_AVM_INVALID"
)

type AVMBytecodeModule struct {
	Magic			string
	VMVersion		uint64
	InstructionSetVersion	uint64
	Instructions		[]AVMInstruction
	MeteringProfile		string
	CanonicalBytes		[]byte
	BytecodeHash		string
}

type AVMStoreV2Entry struct {
	Key		string
	ValueHash	string
	ValueBytes	uint64
	EntryHash	string
}

type AVMStoreV2Adapter struct {
	ContractAddress	string
	Entries		[]AVMStoreV2Entry
	AdapterRoot	string
}

type AVMMessageDrivenInput struct {
	Message		AVMAsyncMessage
	CurrentState	AVMStoreV2Adapter
	Context		AVMExecutionContext
	Bytecode	AVMBytecodeModule
	InputHash	string
}

type AVMStateTransition struct {
	Input			AVMMessageDrivenInput
	Execution		AVMExecutionResult
	UpdatedState		AVMStoreV2Adapter
	Receipt			AVMExecutionReceipt
	StorageRoot		string
	EventRoot		string
	OutboxRoot		string
	ReceiptRoot		string
	StateTransitionHash	string
}

type AVMContractShardRouteSet struct {
	LayoutEpoch	uint64
	InstanceRoute	coretypes.ShardRoute
	StorageRoutes	[]coretypes.ShardRoute
	EventRoutes	[]coretypes.ShardRoute
	MessageRoutes	[]coretypes.ShardRoute
	RouteSetHash	string
}

func NewAVMBytecodeModule(module AVMBytecodeModule, limits AVMLimits, gasTable AVMGasTable) (AVMBytecodeModule, error) {
	module = canonicalAVMBytecodeModule(module)
	if module.Magic == "" {
		module.Magic = AVMBytecodeMagic
	}
	if module.VMVersion == 0 {
		module.VMVersion = AVMVMVersion
	}
	if module.InstructionSetVersion == 0 {
		module.InstructionSetVersion = AVMDefaultInstructionSet
	}
	if module.MeteringProfile == "" {
		module.MeteringProfile = AVMMeteringProfileDefault
	}
	module.CanonicalBytes = EncodeAVMBytecode(module)
	module.BytecodeHash = ComputeAVMBytecodeHash(module)
	return module, module.Validate(limits, gasTable)
}

func NewAVMStoreV2Adapter(adapter AVMStoreV2Adapter) (AVMStoreV2Adapter, error) {
	adapter = canonicalAVMStoreV2Adapter(adapter)
	for i := range adapter.Entries {
		adapter.Entries[i].EntryHash = ComputeAVMStoreV2EntryHash(adapter.Entries[i])
	}
	adapter = canonicalAVMStoreV2Adapter(adapter)
	adapter.AdapterRoot = ComputeAVMStoreV2AdapterRoot(adapter)
	return adapter, adapter.Validate()
}

func NewAVMMessageDrivenInput(input AVMMessageDrivenInput) (AVMMessageDrivenInput, error) {
	input = canonicalAVMMessageDrivenInput(input)
	input.InputHash = ComputeAVMMessageDrivenInputHash(input)
	return input, input.Validate()
}

func ExecuteAVMMessageTransition(input AVMMessageDrivenInput, limits AVMLimits, gasTable AVMGasTable) (AVMStateTransition, error) {
	input = canonicalAVMMessageDrivenInput(input)
	if err := input.Validate(); err != nil {
		return AVMStateTransition{}, err
	}
	program, err := NewAVMProgram(AVMProgram{
		VMVersion:		input.Bytecode.VMVersion,
		InstructionSetVersion:	input.Bytecode.InstructionSetVersion,
		Instructions:		input.Bytecode.Instructions,
		MaxRecursionDepth:	1,
	}, limits, gasTable)
	if err != nil {
		return AVMStateTransition{}, err
	}
	execution, execErr := ExecuteAVMProgram(program, input.Context, limits, gasTable)
	status := AVMReceiptStatusExecuted
	errorCode := ""
	if execErr != nil {
		status = AVMReceiptStatusFailed
		errorCode = AVMRuntimeErrorInvalid
		if strings.Contains(execErr.Error(), "exhausted gas") {
			errorCode = AVMRuntimeErrorGas
		}
		if strings.Contains(execErr.Error(), "aborted") {
			errorCode = AVMRuntimeErrorAbort
		}
		execution = AVMExecutionResult{
			GasUsed:		maxUint64(1, gasUsedBeforeFailure(program, gasTable, limits, input.Context.GasLimit)),
			StorageRoot:		ComputeAVMStorageRoot(nil, nil),
			MessageRoot:		ComputeAVMMessageRoot(nil),
			PromiseRoot:		ComputeAVMPromiseRoot(nil),
			ABIRoot:		ComputeAVMABIRoot(nil),
			EventRoot:		ComputeAVMEventRoot(nil),
			ReadOnlySimulation:	input.Context.ReadOnly,
		}
		execution.ExecutionHash = ComputeAVMExecutionHash(execution)
	}
	updated, err := ApplyAVMStoreV2Writes(input.CurrentState, execution.StorageWrites)
	if err != nil && execErr == nil {
		return AVMStateTransition{}, err
	}
	if execErr != nil {
		updated = input.CurrentState
	}
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		input.Message.ID,
		ZoneID:			zonestypes.ZoneID(coretypes.ZoneIDContract),
		Executor:		AVMRuntimeExecutor,
		Status:			status,
		GasUsed:		execution.GasUsed,
		StorageWritten:		uint32(len(execution.StorageWrites)),
		EventsHash:		execution.EventRoot,
		OutputMessagesRoot:	execution.MessageRoot,
		ErrorCodeOptional:	errorCode,
		CreatedHeight:		input.Context.Height,
	})
	if err != nil {
		return AVMStateTransition{}, err
	}
	transition := AVMStateTransition{
		Input:		input,
		Execution:	execution,
		UpdatedState:	updated,
		Receipt:	receipt,
		StorageRoot:	updated.AdapterRoot,
		EventRoot:	execution.EventRoot,
		OutboxRoot:	execution.MessageRoot,
		ReceiptRoot:	ComputeAVMReceiptRoot([]AVMExecutionReceipt{receipt}),
	}
	transition.StateTransitionHash = ComputeAVMStateTransitionHash(transition)
	if err := transition.Validate(); err != nil {
		return AVMStateTransition{}, err
	}
	return transition, execErr
}

func RouteAVMContractState(layout coretypes.ShardLayout, contract AVMContractRecord, storage []AVMContractStorageValue, events []AVMContractEventRecord, messages []AVMAsyncMessage) (AVMContractShardRouteSet, error) {
	contract = canonicalAVMContractRecord(contract)
	layout.ActiveShards = append([]coretypes.ShardDescriptor(nil), layout.ActiveShards...)
	instanceRoute, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
		ZoneID:			layout.ZoneID,
		StateKey:		AVMContractInstanceStateKey(contract.ContractAddr),
		ShardLayoutEpoch:	layout.LayoutEpoch,
	})
	if err != nil {
		return AVMContractShardRouteSet{}, err
	}
	var storageRoutes []coretypes.ShardRoute
	for _, value := range canonicalAVMStorageValuesForRoute(storage) {
		route, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
			ZoneID:			layout.ZoneID,
			StateKey:		AVMContractStorageStateKey(value.ContractAddr, value.StorageKey),
			ShardLayoutEpoch:	layout.LayoutEpoch,
		})
		if err != nil {
			return AVMContractShardRouteSet{}, err
		}
		storageRoutes = append(storageRoutes, route)
	}
	var eventRoutes []coretypes.ShardRoute
	for _, event := range canonicalAVMEventRecordsForRoute(events) {
		route, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
			ZoneID:			layout.ZoneID,
			StateKey:		event.Key,
			ShardLayoutEpoch:	layout.LayoutEpoch,
		})
		if err != nil {
			return AVMContractShardRouteSet{}, err
		}
		eventRoutes = append(eventRoutes, route)
	}
	var messageRoutes []coretypes.ShardRoute
	for _, msg := range canonicalAVMMessages(messages) {
		route, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
			ZoneID:			layout.ZoneID,
			StateKey:		AVMAsyncMessageKey(msg.ID),
			ShardLayoutEpoch:	layout.LayoutEpoch,
		})
		if err != nil {
			return AVMContractShardRouteSet{}, err
		}
		messageRoutes = append(messageRoutes, route)
	}
	set := AVMContractShardRouteSet{
		LayoutEpoch:	layout.LayoutEpoch,
		InstanceRoute:	instanceRoute,
		StorageRoutes:	storageRoutes,
		EventRoutes:	eventRoutes,
		MessageRoutes:	messageRoutes,
	}
	set = canonicalAVMContractShardRouteSet(set)
	set.RouteSetHash = ComputeAVMContractShardRouteSetHash(set)
	return set, set.Validate()
}

func DecodeAVMBytecode(module AVMBytecodeModule, limits AVMLimits, gasTable AVMGasTable) (AVMBytecodeModule, error) {
	module = canonicalAVMBytecodeModule(module)
	if len(module.CanonicalBytes) == 0 {
		return AVMBytecodeModule{}, errors.New("AVM 2.0 bytecode bytes are required")
	}
	if module.BytecodeHash != ComputeAVMBytecodeHash(module) {
		return AVMBytecodeModule{}, errors.New("AVM 2.0 bytecode hash mismatch")
	}
	return module, module.Validate(limits, gasTable)
}

func (m AVMBytecodeModule) Validate(limits AVMLimits, gasTable AVMGasTable) error {
	m = canonicalAVMBytecodeModule(m)
	if m.Magic != AVMBytecodeMagic {
		return errors.New("AVM 2.0 bytecode magic mismatch")
	}
	if m.VMVersion != AVMVMVersion {
		return errors.New("AVM 2.0 bytecode VM version must be 2")
	}
	if m.InstructionSetVersion == 0 {
		return errors.New("AVM 2.0 bytecode instruction set version must be positive")
	}
	if err := validateEngineToken("AVM 2.0 bytecode metering profile", m.MeteringProfile, MaxAVMTokenLength); err != nil {
		return err
	}
	if len(m.CanonicalBytes) == 0 {
		return errors.New("AVM 2.0 canonical bytecode is required")
	}
	if !strings.HasPrefix(string(m.CanonicalBytes), AVMBytecodeCodec+"|") {
		return errors.New("AVM 2.0 canonical bytecode codec mismatch")
	}
	program := AVMProgram{
		VMVersion:		m.VMVersion,
		InstructionSetVersion:	m.InstructionSetVersion,
		Instructions:		m.Instructions,
		MaxRecursionDepth:	1,
	}
	program.ProgramHash = ComputeAVMProgramHash(program)
	if err := ValidateAVMProgram(program, limits, gasTable); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 bytecode hash", m.BytecodeHash); err != nil {
		return err
	}
	if m.BytecodeHash != ComputeAVMBytecodeHash(m) {
		return errors.New("AVM 2.0 bytecode hash mismatch")
	}
	return nil
}

func (e AVMStoreV2Entry) Validate(contractAddress string) error {
	e = canonicalAVMStoreV2Entry(e)
	if err := validateAVMStatePrefix("AVM 2.0 Store v2 adapter key", e.Key); err != nil {
		return err
	}
	expected := AVMContractStorageKey(contractAddress, "")
	if !strings.HasPrefix(e.Key, expected) {
		return fmt.Errorf("AVM 2.0 Store v2 adapter key must use prefix %q", expected)
	}
	if err := zonestypes.ValidateHash("AVM 2.0 Store v2 value hash", e.ValueHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 Store v2 entry hash", e.EntryHash); err != nil {
		return err
	}
	if e.EntryHash != ComputeAVMStoreV2EntryHash(e) {
		return errors.New("AVM 2.0 Store v2 entry hash mismatch")
	}
	return nil
}

func (a AVMStoreV2Adapter) Validate() error {
	a = canonicalAVMStoreV2Adapter(a)
	if err := validateEngineToken("AVM 2.0 Store v2 contract address", a.ContractAddress, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(a.Entries))
	for i, entry := range a.Entries {
		if err := entry.Validate(a.ContractAddress); err != nil {
			return err
		}
		if _, found := seen[entry.Key]; found {
			return errors.New("duplicate AVM 2.0 Store v2 key")
		}
		seen[entry.Key] = struct{}{}
		if i > 0 && a.Entries[i-1].Key >= entry.Key {
			return errors.New("AVM 2.0 Store v2 entries must be sorted canonically")
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 Store v2 adapter root", a.AdapterRoot); err != nil {
		return err
	}
	if a.AdapterRoot != ComputeAVMStoreV2AdapterRoot(a) {
		return errors.New("AVM 2.0 Store v2 adapter root mismatch")
	}
	return nil
}

func (i AVMMessageDrivenInput) Validate() error {
	i = canonicalAVMMessageDrivenInput(i)
	if err := i.Message.Validate(); err != nil {
		return err
	}
	if err := i.CurrentState.Validate(); err != nil {
		return err
	}
	if err := i.Context.Validate(); err != nil {
		return err
	}
	if err := i.Bytecode.Validate(DefaultAVMLimits(), mustDefaultAVMGasTable()); err != nil {
		return err
	}
	if i.Context.ContractAddress != i.CurrentState.ContractAddress {
		return errors.New("AVM 2.0 message-driven context and state contract mismatch")
	}
	if i.Context.ContractAddress != i.Message.Destination {
		return errors.New("AVM 2.0 message-driven execution destination must match contract")
	}
	if i.Context.Height < i.Message.CreatedHeight {
		return errors.New("AVM 2.0 message-driven execution cannot precede message creation")
	}
	if i.InputHash != ComputeAVMMessageDrivenInputHash(i) {
		return errors.New("AVM 2.0 message-driven input hash mismatch")
	}
	return nil
}

func (t AVMStateTransition) Validate() error {
	t = canonicalAVMStateTransition(t)
	if err := t.Input.Validate(); err != nil {
		return err
	}
	if err := t.Execution.Validate(); err != nil {
		return err
	}
	if err := t.UpdatedState.Validate(); err != nil {
		return err
	}
	if err := t.Receipt.Validate(); err != nil {
		return err
	}
	if t.StorageRoot != t.UpdatedState.AdapterRoot {
		return errors.New("AVM 2.0 transition storage root mismatch")
	}
	if t.EventRoot != t.Execution.EventRoot {
		return errors.New("AVM 2.0 transition event root mismatch")
	}
	if t.OutboxRoot != t.Execution.MessageRoot {
		return errors.New("AVM 2.0 transition outbox root mismatch")
	}
	if t.ReceiptRoot != ComputeAVMReceiptRoot([]AVMExecutionReceipt{t.Receipt}) {
		return errors.New("AVM 2.0 transition receipt root mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 state transition hash", t.StateTransitionHash); err != nil {
		return err
	}
	if t.StateTransitionHash != ComputeAVMStateTransitionHash(t) {
		return errors.New("AVM 2.0 state transition hash mismatch")
	}
	return nil
}

func (s AVMContractShardRouteSet) Validate() error {
	s = canonicalAVMContractShardRouteSet(s)
	if s.LayoutEpoch == 0 {
		return errors.New("AVM 2.0 route set layout epoch must be positive")
	}
	if err := validateCoreRoute("AVM 2.0 instance route", s.InstanceRoute, s.LayoutEpoch); err != nil {
		return err
	}
	for _, route := range s.StorageRoutes {
		if err := validateCoreRoute("AVM 2.0 storage route", route, s.LayoutEpoch); err != nil {
			return err
		}
	}
	for _, route := range s.EventRoutes {
		if err := validateCoreRoute("AVM 2.0 event route", route, s.LayoutEpoch); err != nil {
			return err
		}
	}
	for _, route := range s.MessageRoutes {
		if err := validateCoreRoute("AVM 2.0 message route", route, s.LayoutEpoch); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 route set hash", s.RouteSetHash); err != nil {
		return err
	}
	if s.RouteSetHash != ComputeAVMContractShardRouteSetHash(s) {
		return errors.New("AVM 2.0 route set hash mismatch")
	}
	return nil
}

func ApplyAVMStoreV2Writes(adapter AVMStoreV2Adapter, writes []AVMStorageWrite) (AVMStoreV2Adapter, error) {
	adapter = canonicalAVMStoreV2Adapter(adapter)
	if err := adapter.Validate(); err != nil {
		return AVMStoreV2Adapter{}, err
	}
	byKey := make(map[string]AVMStoreV2Entry, len(adapter.Entries))
	for _, entry := range adapter.Entries {
		byKey[entry.Key] = entry
	}
	for _, write := range canonicalAVMStorageWrites(writes) {
		if err := ValidateAVMStoreV2Key(AVMExecutionContext{ContractAddress: adapter.ContractAddress}, write.Key, DefaultAVMLimits()); err != nil {
			return AVMStoreV2Adapter{}, err
		}
		if write.Deleted {
			delete(byKey, write.Key)
			continue
		}
		byKey[write.Key] = AVMStoreV2Entry{Key: write.Key, ValueHash: write.ValueHash, ValueBytes: uint64(len(write.ValueHash))}
	}
	out := AVMStoreV2Adapter{ContractAddress: adapter.ContractAddress}
	for _, entry := range byKey {
		entry.EntryHash = ComputeAVMStoreV2EntryHash(entry)
		out.Entries = append(out.Entries, entry)
	}
	return NewAVMStoreV2Adapter(out)
}

func EncodeAVMBytecode(module AVMBytecodeModule) []byte {
	module = canonicalAVMBytecodeModule(module)
	parts := []string{
		AVMBytecodeCodec,
		module.Magic,
		fmt.Sprint(module.VMVersion),
		fmt.Sprint(module.InstructionSetVersion),
		module.MeteringProfile,
		fmt.Sprint(len(module.Instructions)),
	}
	for _, instruction := range module.Instructions {
		parts = append(parts,
			string(instruction.Opcode),
			instruction.Key,
			hex.EncodeToString(instruction.Value),
			fmt.Sprint(instruction.MemoryGrow),
			fmt.Sprint(instruction.RangeLimit),
			fmt.Sprint(instruction.GasOverride),
			instruction.Message.ID,
			instruction.Proof.ProofHash,
			instruction.Promise.PromiseHash,
			instruction.ABI.InterfaceHash,
			instruction.Event.EventHash,
		)
	}
	return []byte(strings.Join(parts, "|"))
}

func ComputeAVMBytecodeHash(module AVMBytecodeModule) string {
	module = canonicalAVMBytecodeModule(module)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-bytecode-v1")
	writeEnginePart(h, string(EncodeAVMBytecode(module)))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2EntryHash(entry AVMStoreV2Entry) string {
	entry = canonicalAVMStoreV2Entry(entry)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-storev2-entry-v1")
	writeEnginePart(h, entry.Key)
	writeEnginePart(h, entry.ValueHash)
	writeEngineUint64(h, entry.ValueBytes)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2AdapterRoot(adapter AVMStoreV2Adapter) string {
	adapter = canonicalAVMStoreV2Adapter(adapter)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-storev2-root-v1")
	writeEnginePart(h, adapter.ContractAddress)
	writeEngineUint64(h, uint64(len(adapter.Entries)))
	for _, entry := range adapter.Entries {
		writeEnginePart(h, entry.EntryHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMMessageDrivenInputHash(input AVMMessageDrivenInput) string {
	input = canonicalAVMMessageDrivenInput(input)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-message-driven-input-v1")
	writeEnginePart(h, input.Message.ID)
	writeEnginePart(h, input.CurrentState.AdapterRoot)
	writeEnginePart(h, input.Context.ContextHash)
	writeEnginePart(h, input.Bytecode.BytecodeHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStateTransitionHash(transition AVMStateTransition) string {
	transition = canonicalAVMStateTransition(transition)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-state-transition-v1")
	writeEnginePart(h, transition.Input.InputHash)
	writeEnginePart(h, transition.Execution.ExecutionHash)
	writeEnginePart(h, transition.UpdatedState.AdapterRoot)
	writeEnginePart(h, transition.Receipt.ReceiptHash)
	writeEnginePart(h, transition.StorageRoot)
	writeEnginePart(h, transition.EventRoot)
	writeEnginePart(h, transition.OutboxRoot)
	writeEnginePart(h, transition.ReceiptRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMReceiptRoot(receipts []AVMExecutionReceipt) string {
	out := append([]AVMExecutionReceipt(nil), receipts...)
	for i := range out {
		out[i] = canonicalAVMExecutionReceipt(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReceiptID < out[j].ReceiptID })
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-receipt-root-v1")
	writeEngineUint64(h, uint64(len(out)))
	for _, receipt := range out {
		writeEnginePart(h, receipt.ReceiptHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractShardRouteSetHash(set AVMContractShardRouteSet) string {
	set = canonicalAVMContractShardRouteSet(set)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-shard-routes-v1")
	writeEngineUint64(h, set.LayoutEpoch)
	writeCoreRouteParts(h, set.InstanceRoute)
	writeEngineUint64(h, uint64(len(set.StorageRoutes)))
	for _, route := range set.StorageRoutes {
		writeCoreRouteParts(h, route)
	}
	writeEngineUint64(h, uint64(len(set.EventRoutes)))
	for _, route := range set.EventRoutes {
		writeCoreRouteParts(h, route)
	}
	writeEngineUint64(h, uint64(len(set.MessageRoutes)))
	for _, route := range set.MessageRoutes {
		writeCoreRouteParts(h, route)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMBytecodeModule(module AVMBytecodeModule) AVMBytecodeModule {
	module.Magic = strings.TrimSpace(module.Magic)
	module.MeteringProfile = strings.TrimSpace(module.MeteringProfile)
	module.Instructions = append([]AVMInstruction(nil), module.Instructions...)
	for i := range module.Instructions {
		module.Instructions[i] = canonicalAVMInstruction(module.Instructions[i])
	}
	module.CanonicalBytes = append([]byte(nil), module.CanonicalBytes...)
	module.BytecodeHash = strings.TrimSpace(module.BytecodeHash)
	return module
}

func canonicalAVMStoreV2Entry(entry AVMStoreV2Entry) AVMStoreV2Entry {
	entry.Key = strings.TrimSpace(entry.Key)
	entry.ValueHash = strings.TrimSpace(entry.ValueHash)
	entry.EntryHash = strings.TrimSpace(entry.EntryHash)
	return entry
}

func canonicalAVMStoreV2Adapter(adapter AVMStoreV2Adapter) AVMStoreV2Adapter {
	adapter.ContractAddress = strings.TrimSpace(adapter.ContractAddress)
	adapter.Entries = append([]AVMStoreV2Entry(nil), adapter.Entries...)
	for i := range adapter.Entries {
		adapter.Entries[i] = canonicalAVMStoreV2Entry(adapter.Entries[i])
	}
	sort.SliceStable(adapter.Entries, func(i, j int) bool { return adapter.Entries[i].Key < adapter.Entries[j].Key })
	adapter.AdapterRoot = strings.TrimSpace(adapter.AdapterRoot)
	return adapter
}

func canonicalAVMMessageDrivenInput(input AVMMessageDrivenInput) AVMMessageDrivenInput {
	input.Message = canonicalAVMAsyncMessage(input.Message)
	input.CurrentState = canonicalAVMStoreV2Adapter(input.CurrentState)
	input.Context = canonicalAVMExecutionContext(input.Context)
	input.Bytecode = canonicalAVMBytecodeModule(input.Bytecode)
	input.InputHash = strings.TrimSpace(input.InputHash)
	return input
}

func canonicalAVMStateTransition(transition AVMStateTransition) AVMStateTransition {
	transition.Input = canonicalAVMMessageDrivenInput(transition.Input)
	transition.Execution = canonicalAVMExecutionResult(transition.Execution)
	transition.UpdatedState = canonicalAVMStoreV2Adapter(transition.UpdatedState)
	transition.Receipt = canonicalAVMExecutionReceipt(transition.Receipt)
	transition.StorageRoot = strings.TrimSpace(transition.StorageRoot)
	transition.EventRoot = strings.TrimSpace(transition.EventRoot)
	transition.OutboxRoot = strings.TrimSpace(transition.OutboxRoot)
	transition.ReceiptRoot = strings.TrimSpace(transition.ReceiptRoot)
	transition.StateTransitionHash = strings.TrimSpace(transition.StateTransitionHash)
	return transition
}

func canonicalAVMContractShardRouteSet(set AVMContractShardRouteSet) AVMContractShardRouteSet {
	set.StorageRoutes = append([]coretypes.ShardRoute(nil), set.StorageRoutes...)
	sort.SliceStable(set.StorageRoutes, func(i, j int) bool { return set.StorageRoutes[i].StateKey < set.StorageRoutes[j].StateKey })
	set.EventRoutes = append([]coretypes.ShardRoute(nil), set.EventRoutes...)
	sort.SliceStable(set.EventRoutes, func(i, j int) bool { return set.EventRoutes[i].StateKey < set.EventRoutes[j].StateKey })
	set.MessageRoutes = append([]coretypes.ShardRoute(nil), set.MessageRoutes...)
	sort.SliceStable(set.MessageRoutes, func(i, j int) bool { return set.MessageRoutes[i].StateKey < set.MessageRoutes[j].StateKey })
	set.RouteSetHash = strings.TrimSpace(set.RouteSetHash)
	return set
}

func canonicalAVMStorageValuesForRoute(values []AVMContractStorageValue) []AVMContractStorageValue {
	out := append([]AVMContractStorageValue(nil), values...)
	for i := range out {
		out[i] = canonicalAVMContractStorageValue(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		return AVMContractStorageStateKey(out[i].ContractAddr, out[i].StorageKey) < AVMContractStorageStateKey(out[j].ContractAddr, out[j].StorageKey)
	})
	return out
}

func canonicalAVMEventRecordsForRoute(events []AVMContractEventRecord) []AVMContractEventRecord {
	out := append([]AVMContractEventRecord(nil), events...)
	for i := range out {
		out[i] = canonicalAVMContractEventRecord(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func validateCoreRoute(fieldName string, route coretypes.ShardRoute, epoch uint64) error {
	if route.LayoutEpoch != epoch {
		return fmt.Errorf("%s layout epoch mismatch", fieldName)
	}
	if err := coretypes.ValidateShardID(route.ShardID); err != nil {
		return err
	}
	if strings.TrimSpace(route.StateKey) == "" {
		return fmt.Errorf("%s state key is required", fieldName)
	}
	return nil
}

func writeCoreRouteParts(h engineByteWriter, route coretypes.ShardRoute) {
	writeEnginePart(h, string(route.ZoneID))
	writeEnginePart(h, route.StateKey)
	writeEngineUint64(h, route.LayoutEpoch)
	writeEnginePart(h, string(route.AssignmentMode))
	writeEnginePart(h, string(route.ShardID))
	writeEngineUint64(h, uint64(route.ShardCount))
	writeEngineBool(h, route.ReadOnlyReplicated)
}

func gasUsedBeforeFailure(program AVMProgram, gasTable AVMGasTable, limits AVMLimits, gasLimit uint64) uint64 {
	var gas uint64
	for _, instruction := range canonicalAVMProgram(program).Instructions {
		cost, err := AVMInstructionGas(instruction, gasTable, limits)
		if err != nil {
			return maxUint64(1, gas)
		}
		next, err := checkedAVMGasAdd(gas, cost)
		if err != nil {
			return maxUint64(1, gas)
		}
		gas = next
		if gas >= gasLimit || instruction.Opcode == AVMOpAbort {
			return gas
		}
	}
	return gas
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func mustDefaultAVMGasTable() AVMGasTable {
	table, err := DefaultAVMGasTable()
	if err != nil {
		panic(err)
	}
	return table
}
