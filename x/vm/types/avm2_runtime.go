package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aethercore/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVM2BytecodeMagic       = "AVM2"
	AVM2BytecodeCodec       = "avm2.bytecode.v1"
	AVM2RuntimeExecutor     = "avm2-runtime"
	AVM2RuntimeErrorAbort   = "ERR_AVM2_ABORT"
	AVM2RuntimeErrorGas     = "ERR_AVM2_GAS"
	AVM2RuntimeErrorInvalid = "ERR_AVM2_INVALID"
)

type AVM2BytecodeModule struct {
	Magic                 string
	VMVersion             uint64
	InstructionSetVersion uint64
	Instructions          []AVM2Instruction
	MeteringProfile       string
	CanonicalBytes        []byte
	BytecodeHash          string
}

type AVM2StoreV2Entry struct {
	Key        string
	ValueHash  string
	ValueBytes uint64
	EntryHash  string
}

type AVM2StoreV2Adapter struct {
	ContractAddress string
	Entries         []AVM2StoreV2Entry
	AdapterRoot     string
}

type AVM2MessageDrivenInput struct {
	Message      AVMAsyncMessage
	CurrentState AVM2StoreV2Adapter
	Context      AVM2ExecutionContext
	Bytecode     AVM2BytecodeModule
	InputHash    string
}

type AVM2StateTransition struct {
	Input               AVM2MessageDrivenInput
	Execution           AVM2ExecutionResult
	UpdatedState        AVM2StoreV2Adapter
	Receipt             AVMExecutionReceipt
	StorageRoot         string
	EventRoot           string
	OutboxRoot          string
	ReceiptRoot         string
	StateTransitionHash string
}

type AVM2ContractShardRouteSet struct {
	LayoutEpoch   uint64
	InstanceRoute coretypes.ShardRoute
	StorageRoutes []coretypes.ShardRoute
	EventRoutes   []coretypes.ShardRoute
	MessageRoutes []coretypes.ShardRoute
	RouteSetHash  string
}

func NewAVM2BytecodeModule(module AVM2BytecodeModule, limits AVM2Limits, gasTable AVM2GasTable) (AVM2BytecodeModule, error) {
	module = canonicalAVM2BytecodeModule(module)
	if module.Magic == "" {
		module.Magic = AVM2BytecodeMagic
	}
	if module.VMVersion == 0 {
		module.VMVersion = AVM2VMVersion
	}
	if module.InstructionSetVersion == 0 {
		module.InstructionSetVersion = AVM2DefaultInstructionSet
	}
	if module.MeteringProfile == "" {
		module.MeteringProfile = AVM2MeteringProfileDefault
	}
	module.CanonicalBytes = EncodeAVM2Bytecode(module)
	module.BytecodeHash = ComputeAVM2BytecodeHash(module)
	return module, module.Validate(limits, gasTable)
}

func NewAVM2StoreV2Adapter(adapter AVM2StoreV2Adapter) (AVM2StoreV2Adapter, error) {
	adapter = canonicalAVM2StoreV2Adapter(adapter)
	for i := range adapter.Entries {
		adapter.Entries[i].EntryHash = ComputeAVM2StoreV2EntryHash(adapter.Entries[i])
	}
	adapter = canonicalAVM2StoreV2Adapter(adapter)
	adapter.AdapterRoot = ComputeAVM2StoreV2AdapterRoot(adapter)
	return adapter, adapter.Validate()
}

func NewAVM2MessageDrivenInput(input AVM2MessageDrivenInput) (AVM2MessageDrivenInput, error) {
	input = canonicalAVM2MessageDrivenInput(input)
	input.InputHash = ComputeAVM2MessageDrivenInputHash(input)
	return input, input.Validate()
}

func ExecuteAVM2MessageTransition(input AVM2MessageDrivenInput, limits AVM2Limits, gasTable AVM2GasTable) (AVM2StateTransition, error) {
	input = canonicalAVM2MessageDrivenInput(input)
	if err := input.Validate(); err != nil {
		return AVM2StateTransition{}, err
	}
	program, err := NewAVM2Program(AVM2Program{
		VMVersion:             input.Bytecode.VMVersion,
		InstructionSetVersion: input.Bytecode.InstructionSetVersion,
		Instructions:          input.Bytecode.Instructions,
		MaxRecursionDepth:     1,
	}, limits, gasTable)
	if err != nil {
		return AVM2StateTransition{}, err
	}
	execution, execErr := ExecuteAVM2Program(program, input.Context, limits, gasTable)
	status := AVMReceiptStatusExecuted
	errorCode := ""
	if execErr != nil {
		status = AVMReceiptStatusFailed
		errorCode = AVM2RuntimeErrorInvalid
		if strings.Contains(execErr.Error(), "exhausted gas") {
			errorCode = AVM2RuntimeErrorGas
		}
		if strings.Contains(execErr.Error(), "aborted") {
			errorCode = AVM2RuntimeErrorAbort
		}
		execution = AVM2ExecutionResult{
			GasUsed:            maxUint64(1, gasUsedBeforeFailure(program, gasTable, limits, input.Context.GasLimit)),
			StorageRoot:        ComputeAVM2StorageRoot(nil, nil),
			MessageRoot:        ComputeAVM2MessageRoot(nil),
			PromiseRoot:        ComputeAVM2PromiseRoot(nil),
			ABIRoot:            ComputeAVM2ABIRoot(nil),
			EventRoot:          ComputeAVM2EventRoot(nil),
			ReadOnlySimulation: input.Context.ReadOnly,
		}
		execution.ExecutionHash = ComputeAVM2ExecutionHash(execution)
	}
	updated, err := ApplyAVM2StoreV2Writes(input.CurrentState, execution.StorageWrites)
	if err != nil && execErr == nil {
		return AVM2StateTransition{}, err
	}
	if execErr != nil {
		updated = input.CurrentState
	}
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:          input.Message.ID,
		ZoneID:             zonestypes.ZoneID(coretypes.ZoneIDContract),
		Executor:           AVM2RuntimeExecutor,
		Status:             status,
		GasUsed:            execution.GasUsed,
		StorageWritten:     uint32(len(execution.StorageWrites)),
		EventsHash:         execution.EventRoot,
		OutputMessagesRoot: execution.MessageRoot,
		ErrorCodeOptional:  errorCode,
		CreatedHeight:      input.Context.Height,
	})
	if err != nil {
		return AVM2StateTransition{}, err
	}
	transition := AVM2StateTransition{
		Input:        input,
		Execution:    execution,
		UpdatedState: updated,
		Receipt:      receipt,
		StorageRoot:  updated.AdapterRoot,
		EventRoot:    execution.EventRoot,
		OutboxRoot:   execution.MessageRoot,
		ReceiptRoot:  ComputeAVM2ReceiptRoot([]AVMExecutionReceipt{receipt}),
	}
	transition.StateTransitionHash = ComputeAVM2StateTransitionHash(transition)
	if err := transition.Validate(); err != nil {
		return AVM2StateTransition{}, err
	}
	return transition, execErr
}

func RouteAVM2ContractState(layout coretypes.ShardLayout, contract AVM2ContractRecord, storage []AVM2ContractStorageValue, events []AVM2ContractEventRecord, messages []AVMAsyncMessage) (AVM2ContractShardRouteSet, error) {
	contract = canonicalAVM2ContractRecord(contract)
	layout.ActiveShards = append([]coretypes.ShardDescriptor(nil), layout.ActiveShards...)
	instanceRoute, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
		ZoneID:           layout.ZoneID,
		StateKey:         AVM2ContractInstanceStateKey(contract.ContractAddr),
		ShardLayoutEpoch: layout.LayoutEpoch,
	})
	if err != nil {
		return AVM2ContractShardRouteSet{}, err
	}
	var storageRoutes []coretypes.ShardRoute
	for _, value := range canonicalAVM2StorageValuesForRoute(storage) {
		route, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
			ZoneID:           layout.ZoneID,
			StateKey:         AVM2ContractStorageStateKey(value.ContractAddr, value.StorageKey),
			ShardLayoutEpoch: layout.LayoutEpoch,
		})
		if err != nil {
			return AVM2ContractShardRouteSet{}, err
		}
		storageRoutes = append(storageRoutes, route)
	}
	var eventRoutes []coretypes.ShardRoute
	for _, event := range canonicalAVM2EventRecordsForRoute(events) {
		route, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
			ZoneID:           layout.ZoneID,
			StateKey:         event.Key,
			ShardLayoutEpoch: layout.LayoutEpoch,
		})
		if err != nil {
			return AVM2ContractShardRouteSet{}, err
		}
		eventRoutes = append(eventRoutes, route)
	}
	var messageRoutes []coretypes.ShardRoute
	for _, msg := range canonicalAVM2Messages(messages) {
		route, err := coretypes.RouteKeyToShard(layout, coretypes.ShardRoutingInput{
			ZoneID:           layout.ZoneID,
			StateKey:         AVMAsyncMessageKey(msg.ID),
			ShardLayoutEpoch: layout.LayoutEpoch,
		})
		if err != nil {
			return AVM2ContractShardRouteSet{}, err
		}
		messageRoutes = append(messageRoutes, route)
	}
	set := AVM2ContractShardRouteSet{
		LayoutEpoch:   layout.LayoutEpoch,
		InstanceRoute: instanceRoute,
		StorageRoutes: storageRoutes,
		EventRoutes:   eventRoutes,
		MessageRoutes: messageRoutes,
	}
	set = canonicalAVM2ContractShardRouteSet(set)
	set.RouteSetHash = ComputeAVM2ContractShardRouteSetHash(set)
	return set, set.Validate()
}

func DecodeAVM2Bytecode(module AVM2BytecodeModule, limits AVM2Limits, gasTable AVM2GasTable) (AVM2BytecodeModule, error) {
	module = canonicalAVM2BytecodeModule(module)
	if len(module.CanonicalBytes) == 0 {
		return AVM2BytecodeModule{}, errors.New("AVM 2.0 bytecode bytes are required")
	}
	if module.BytecodeHash != ComputeAVM2BytecodeHash(module) {
		return AVM2BytecodeModule{}, errors.New("AVM 2.0 bytecode hash mismatch")
	}
	return module, module.Validate(limits, gasTable)
}

func (m AVM2BytecodeModule) Validate(limits AVM2Limits, gasTable AVM2GasTable) error {
	m = canonicalAVM2BytecodeModule(m)
	if m.Magic != AVM2BytecodeMagic {
		return errors.New("AVM 2.0 bytecode magic mismatch")
	}
	if m.VMVersion != AVM2VMVersion {
		return errors.New("AVM 2.0 bytecode VM version must be 2")
	}
	if m.InstructionSetVersion == 0 {
		return errors.New("AVM 2.0 bytecode instruction set version must be positive")
	}
	if err := validateEngineToken("AVM 2.0 bytecode metering profile", m.MeteringProfile, MaxAVM2TokenLength); err != nil {
		return err
	}
	if len(m.CanonicalBytes) == 0 {
		return errors.New("AVM 2.0 canonical bytecode is required")
	}
	if !strings.HasPrefix(string(m.CanonicalBytes), AVM2BytecodeCodec+"|") {
		return errors.New("AVM 2.0 canonical bytecode codec mismatch")
	}
	program := AVM2Program{
		VMVersion:             m.VMVersion,
		InstructionSetVersion: m.InstructionSetVersion,
		Instructions:          m.Instructions,
		MaxRecursionDepth:     1,
	}
	program.ProgramHash = ComputeAVM2ProgramHash(program)
	if err := ValidateAVM2Program(program, limits, gasTable); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 bytecode hash", m.BytecodeHash); err != nil {
		return err
	}
	if m.BytecodeHash != ComputeAVM2BytecodeHash(m) {
		return errors.New("AVM 2.0 bytecode hash mismatch")
	}
	return nil
}

func (e AVM2StoreV2Entry) Validate(contractAddress string) error {
	e = canonicalAVM2StoreV2Entry(e)
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
	if e.EntryHash != ComputeAVM2StoreV2EntryHash(e) {
		return errors.New("AVM 2.0 Store v2 entry hash mismatch")
	}
	return nil
}

func (a AVM2StoreV2Adapter) Validate() error {
	a = canonicalAVM2StoreV2Adapter(a)
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
	if a.AdapterRoot != ComputeAVM2StoreV2AdapterRoot(a) {
		return errors.New("AVM 2.0 Store v2 adapter root mismatch")
	}
	return nil
}

func (i AVM2MessageDrivenInput) Validate() error {
	i = canonicalAVM2MessageDrivenInput(i)
	if err := i.Message.Validate(); err != nil {
		return err
	}
	if err := i.CurrentState.Validate(); err != nil {
		return err
	}
	if err := i.Context.Validate(); err != nil {
		return err
	}
	if err := i.Bytecode.Validate(DefaultAVM2Limits(), mustDefaultAVM2GasTable()); err != nil {
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
	if i.InputHash != ComputeAVM2MessageDrivenInputHash(i) {
		return errors.New("AVM 2.0 message-driven input hash mismatch")
	}
	return nil
}

func (t AVM2StateTransition) Validate() error {
	t = canonicalAVM2StateTransition(t)
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
	if t.ReceiptRoot != ComputeAVM2ReceiptRoot([]AVMExecutionReceipt{t.Receipt}) {
		return errors.New("AVM 2.0 transition receipt root mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 state transition hash", t.StateTransitionHash); err != nil {
		return err
	}
	if t.StateTransitionHash != ComputeAVM2StateTransitionHash(t) {
		return errors.New("AVM 2.0 state transition hash mismatch")
	}
	return nil
}

func (s AVM2ContractShardRouteSet) Validate() error {
	s = canonicalAVM2ContractShardRouteSet(s)
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
	if s.RouteSetHash != ComputeAVM2ContractShardRouteSetHash(s) {
		return errors.New("AVM 2.0 route set hash mismatch")
	}
	return nil
}

func ApplyAVM2StoreV2Writes(adapter AVM2StoreV2Adapter, writes []AVM2StorageWrite) (AVM2StoreV2Adapter, error) {
	adapter = canonicalAVM2StoreV2Adapter(adapter)
	if err := adapter.Validate(); err != nil {
		return AVM2StoreV2Adapter{}, err
	}
	byKey := make(map[string]AVM2StoreV2Entry, len(adapter.Entries))
	for _, entry := range adapter.Entries {
		byKey[entry.Key] = entry
	}
	for _, write := range canonicalAVM2StorageWrites(writes) {
		if err := ValidateAVM2StoreV2Key(AVM2ExecutionContext{ContractAddress: adapter.ContractAddress}, write.Key, DefaultAVM2Limits()); err != nil {
			return AVM2StoreV2Adapter{}, err
		}
		if write.Deleted {
			delete(byKey, write.Key)
			continue
		}
		byKey[write.Key] = AVM2StoreV2Entry{Key: write.Key, ValueHash: write.ValueHash, ValueBytes: uint64(len(write.ValueHash))}
	}
	out := AVM2StoreV2Adapter{ContractAddress: adapter.ContractAddress}
	for _, entry := range byKey {
		entry.EntryHash = ComputeAVM2StoreV2EntryHash(entry)
		out.Entries = append(out.Entries, entry)
	}
	return NewAVM2StoreV2Adapter(out)
}

func EncodeAVM2Bytecode(module AVM2BytecodeModule) []byte {
	module = canonicalAVM2BytecodeModule(module)
	parts := []string{
		AVM2BytecodeCodec,
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

func ComputeAVM2BytecodeHash(module AVM2BytecodeModule) string {
	module = canonicalAVM2BytecodeModule(module)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-bytecode-v1")
	writeEnginePart(h, string(EncodeAVM2Bytecode(module)))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2StoreV2EntryHash(entry AVM2StoreV2Entry) string {
	entry = canonicalAVM2StoreV2Entry(entry)
	entry.EntryHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-storev2-entry-v1")
	writeEnginePart(h, entry.Key)
	writeEnginePart(h, entry.ValueHash)
	writeEngineUint64(h, entry.ValueBytes)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2StoreV2AdapterRoot(adapter AVM2StoreV2Adapter) string {
	adapter = canonicalAVM2StoreV2Adapter(adapter)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-storev2-root-v1")
	writeEnginePart(h, adapter.ContractAddress)
	writeEngineUint64(h, uint64(len(adapter.Entries)))
	for _, entry := range adapter.Entries {
		writeEnginePart(h, entry.EntryHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2MessageDrivenInputHash(input AVM2MessageDrivenInput) string {
	input = canonicalAVM2MessageDrivenInput(input)
	input.InputHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-message-driven-input-v1")
	writeEnginePart(h, input.Message.ID)
	writeEnginePart(h, input.CurrentState.AdapterRoot)
	writeEnginePart(h, input.Context.ContextHash)
	writeEnginePart(h, input.Bytecode.BytecodeHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2StateTransitionHash(transition AVM2StateTransition) string {
	transition = canonicalAVM2StateTransition(transition)
	transition.StateTransitionHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-state-transition-v1")
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

func ComputeAVM2ReceiptRoot(receipts []AVMExecutionReceipt) string {
	out := append([]AVMExecutionReceipt(nil), receipts...)
	for i := range out {
		out[i] = canonicalAVMExecutionReceipt(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReceiptID < out[j].ReceiptID })
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-receipt-root-v1")
	writeEngineUint64(h, uint64(len(out)))
	for _, receipt := range out {
		writeEnginePart(h, receipt.ReceiptHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractShardRouteSetHash(set AVM2ContractShardRouteSet) string {
	set = canonicalAVM2ContractShardRouteSet(set)
	set.RouteSetHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-shard-routes-v1")
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

func canonicalAVM2BytecodeModule(module AVM2BytecodeModule) AVM2BytecodeModule {
	module.Magic = strings.TrimSpace(module.Magic)
	module.MeteringProfile = strings.TrimSpace(module.MeteringProfile)
	module.Instructions = append([]AVM2Instruction(nil), module.Instructions...)
	for i := range module.Instructions {
		module.Instructions[i] = canonicalAVM2Instruction(module.Instructions[i])
	}
	module.CanonicalBytes = append([]byte(nil), module.CanonicalBytes...)
	module.BytecodeHash = strings.TrimSpace(module.BytecodeHash)
	return module
}

func canonicalAVM2StoreV2Entry(entry AVM2StoreV2Entry) AVM2StoreV2Entry {
	entry.Key = strings.TrimSpace(entry.Key)
	entry.ValueHash = strings.TrimSpace(entry.ValueHash)
	entry.EntryHash = strings.TrimSpace(entry.EntryHash)
	return entry
}

func canonicalAVM2StoreV2Adapter(adapter AVM2StoreV2Adapter) AVM2StoreV2Adapter {
	adapter.ContractAddress = strings.TrimSpace(adapter.ContractAddress)
	adapter.Entries = append([]AVM2StoreV2Entry(nil), adapter.Entries...)
	for i := range adapter.Entries {
		adapter.Entries[i] = canonicalAVM2StoreV2Entry(adapter.Entries[i])
	}
	sort.SliceStable(adapter.Entries, func(i, j int) bool { return adapter.Entries[i].Key < adapter.Entries[j].Key })
	adapter.AdapterRoot = strings.TrimSpace(adapter.AdapterRoot)
	return adapter
}

func canonicalAVM2MessageDrivenInput(input AVM2MessageDrivenInput) AVM2MessageDrivenInput {
	input.Message = canonicalAVMAsyncMessage(input.Message)
	input.CurrentState = canonicalAVM2StoreV2Adapter(input.CurrentState)
	input.Context = canonicalAVM2ExecutionContext(input.Context)
	input.Bytecode = canonicalAVM2BytecodeModule(input.Bytecode)
	input.InputHash = strings.TrimSpace(input.InputHash)
	return input
}

func canonicalAVM2StateTransition(transition AVM2StateTransition) AVM2StateTransition {
	transition.Input = canonicalAVM2MessageDrivenInput(transition.Input)
	transition.Execution = canonicalAVM2ExecutionResult(transition.Execution)
	transition.UpdatedState = canonicalAVM2StoreV2Adapter(transition.UpdatedState)
	transition.Receipt = canonicalAVMExecutionReceipt(transition.Receipt)
	transition.StorageRoot = strings.TrimSpace(transition.StorageRoot)
	transition.EventRoot = strings.TrimSpace(transition.EventRoot)
	transition.OutboxRoot = strings.TrimSpace(transition.OutboxRoot)
	transition.ReceiptRoot = strings.TrimSpace(transition.ReceiptRoot)
	transition.StateTransitionHash = strings.TrimSpace(transition.StateTransitionHash)
	return transition
}

func canonicalAVM2ContractShardRouteSet(set AVM2ContractShardRouteSet) AVM2ContractShardRouteSet {
	set.StorageRoutes = append([]coretypes.ShardRoute(nil), set.StorageRoutes...)
	sort.SliceStable(set.StorageRoutes, func(i, j int) bool { return set.StorageRoutes[i].StateKey < set.StorageRoutes[j].StateKey })
	set.EventRoutes = append([]coretypes.ShardRoute(nil), set.EventRoutes...)
	sort.SliceStable(set.EventRoutes, func(i, j int) bool { return set.EventRoutes[i].StateKey < set.EventRoutes[j].StateKey })
	set.MessageRoutes = append([]coretypes.ShardRoute(nil), set.MessageRoutes...)
	sort.SliceStable(set.MessageRoutes, func(i, j int) bool { return set.MessageRoutes[i].StateKey < set.MessageRoutes[j].StateKey })
	set.RouteSetHash = strings.TrimSpace(set.RouteSetHash)
	return set
}

func canonicalAVM2StorageValuesForRoute(values []AVM2ContractStorageValue) []AVM2ContractStorageValue {
	out := append([]AVM2ContractStorageValue(nil), values...)
	for i := range out {
		out[i] = canonicalAVM2ContractStorageValue(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		return AVM2ContractStorageStateKey(out[i].ContractAddr, out[i].StorageKey) < AVM2ContractStorageStateKey(out[j].ContractAddr, out[j].StorageKey)
	})
	return out
}

func canonicalAVM2EventRecordsForRoute(events []AVM2ContractEventRecord) []AVM2ContractEventRecord {
	out := append([]AVM2ContractEventRecord(nil), events...)
	for i := range out {
		out[i] = canonicalAVM2ContractEventRecord(out[i])
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

func gasUsedBeforeFailure(program AVM2Program, gasTable AVM2GasTable, limits AVM2Limits, gasLimit uint64) uint64 {
	var gas uint64
	for _, instruction := range canonicalAVM2Program(program).Instructions {
		cost, err := AVM2InstructionGas(instruction, gasTable, limits)
		if err != nil {
			return maxUint64(1, gas)
		}
		next, err := checkedAVMGasAdd(gas, cost)
		if err != nil {
			return maxUint64(1, gas)
		}
		gas = next
		if gas >= gasLimit || instruction.Opcode == AVM2OpAbort {
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

func mustDefaultAVM2GasTable() AVM2GasTable {
	table, err := DefaultAVM2GasTable()
	if err != nil {
		panic(err)
	}
	return table
}
