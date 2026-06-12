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
	AVMActorContractActive	AVMActorContractStatus	= "active"
	AVMActorContractPaused	AVMActorContractStatus	= "paused"
	AVMActorContractStopped	AVMActorContractStatus	= "stopped"

	AVMContractProofCode		AVMContractProofKind	= "code"
	AVMContractProofInstance	AVMContractProofKind	= "instance"
	AVMContractProofStorage		AVMContractProofKind	= "storage"
	AVMContractProofReceipt		AVMContractProofKind	= "receipt"

	MaxAVMContractRuntimeToken	= 128
	MaxAVMContractProofRecords	= 4096
)

type AVMActorContractStatus string
type AVMContractProofKind string

type AVMContractBackendInterface struct {
	Name			string
	BackendKind		AVMContractBackendKind
	RouterBackend		RouterBackend
	SupportsSync		bool
	SupportsAsync		bool
	EmitsReceipt		bool
	SupportsProofQuery	bool
	InterfaceHash		string
	BackendInterfaceHash	string
}

type AVMNativeModuleAdapter struct {
	Interface	AVMContractBackendInterface
	Descriptor	AVMNativeModuleDescriptor
	AdapterHash	string
}

type AVMWASMAdapterBoundary struct {
	Interface	AVMContractBackendInterface
	SandboxPolicy	AVMWASMSandboxPolicy
	AdapterHash	string
}

type AVMActorContractState struct {
	ActorID			string
	CodeID			uint64
	Owner			string
	MailboxRoot		string
	StateRoot		string
	ContinuationRoot	string
	InterfaceHash		string
	Status			AVMActorContractStatus
	StateHash		string
}

type AVMActorStateRead struct {
	ActorID	string
	Key	string
	Hash	string
}

type AVMActorContractExecution struct {
	Actor			AVMActorContractState
	Message			AVMAsyncMessage
	ActiveMessageCount	uint32
	StateReads		[]AVMActorStateRead
	StateWrites		[]ActorStateWrite
	EmittedMessages		[]AVMAsyncMessage
	StoredContinuations	[]ContinuationRecord
	Receipt			AVMExecutionReceipt
	DirectMutableReads	bool
	ExecutionHash		string
}

type AVMContractCodeRegistryRecord struct {
	CodeID		uint64
	BackendKind	AVMContractBackendKind
	CodeHash	string
	InterfaceHash	string
	Owner		string
	Enabled		bool
	CodeKey		string
	RecordHash	string
}

type AVMContractCodeRegistry struct {
	Codes		[]AVMContractCodeRegistryRecord
	RegistryRoot	string
}

type AVMContractInstanceRegistryRecord struct {
	ContractAddress	string
	CodeID		uint64
	BackendKind	AVMContractBackendKind
	ActorID		string
	Owner		string
	StoragePrefix	string
	StateRoot	string
	InterfaceHash	string
	Status		AVMActorContractStatus
	InstanceKey	string
	RecordHash	string
}

type AVMContractInstanceRegistry struct {
	Instances	[]AVMContractInstanceRegistryRecord
	RegistryRoot	string
}

type AVMContractStoragePrefixDescriptor struct {
	ContractAddress	string
	Prefix		string
	StorageRoot	string
	PrefixHash	string
}

type AVMContractProof struct {
	Kind		AVMContractProofKind
	QueryKey	string
	Root		string
	RecordHash	string
	ProofHash	string
}

type AVMContractProofIndex struct {
	CodeRegistry		AVMContractCodeRegistry
	InstanceRegistry	AVMContractInstanceRegistry
	StoragePrefixes		[]AVMContractStoragePrefixDescriptor
	Receipts		[]AVMExecutionReceipt
}

func NewAVMContractBackendInterface(iface AVMContractBackendInterface) (AVMContractBackendInterface, error) {
	iface = canonicalAVMContractBackendInterface(iface)
	iface.BackendInterfaceHash = ComputeAVMContractBackendInterfaceHash(iface)
	return iface, iface.Validate()
}

func NewAVMNativeModuleAdapter(adapter AVMNativeModuleAdapter) (AVMNativeModuleAdapter, error) {
	adapter = canonicalAVMNativeModuleAdapter(adapter)
	adapter.AdapterHash = ComputeAVMNativeModuleAdapterHash(adapter)
	return adapter, adapter.Validate()
}

func NewAVMWASMAdapterBoundary(boundary AVMWASMAdapterBoundary) (AVMWASMAdapterBoundary, error) {
	boundary = canonicalAVMWASMAdapterBoundary(boundary)
	boundary.AdapterHash = ComputeAVMWASMAdapterBoundaryHash(boundary)
	return boundary, boundary.Validate()
}

func NewAVMActorContractState(state AVMActorContractState) (AVMActorContractState, error) {
	state = canonicalAVMActorContractState(state)
	state.StateHash = ComputeAVMActorContractStateHash(state)
	return state, state.Validate()
}

func NewAVMActorContractExecution(execution AVMActorContractExecution) (AVMActorContractExecution, error) {
	execution = canonicalAVMActorContractExecution(execution)
	execution.ExecutionHash = ComputeAVMActorContractExecutionHash(execution)
	return execution, execution.Validate()
}

func NewAVMContractCodeRegistry(registry AVMContractCodeRegistry) (AVMContractCodeRegistry, error) {
	registry = canonicalAVMContractCodeRegistry(registry)
	for i := range registry.Codes {
		if registry.Codes[i].CodeKey == "" {
			registry.Codes[i].CodeKey = AVMContractCodeKey(registry.Codes[i].CodeID)
		}
		registry.Codes[i].RecordHash = ComputeAVMContractCodeRecordHash(registry.Codes[i])
	}
	registry = canonicalAVMContractCodeRegistry(registry)
	registry.RegistryRoot = ComputeAVMContractCodeRegistryRoot(registry)
	return registry, registry.Validate()
}

func NewAVMContractInstanceRegistry(registry AVMContractInstanceRegistry) (AVMContractInstanceRegistry, error) {
	registry = canonicalAVMContractInstanceRegistry(registry)
	for i := range registry.Instances {
		if registry.Instances[i].InstanceKey == "" {
			registry.Instances[i].InstanceKey = AVMContractInstanceKey(registry.Instances[i].ContractAddress)
		}
		registry.Instances[i].RecordHash = ComputeAVMContractInstanceRecordHash(registry.Instances[i])
	}
	registry = canonicalAVMContractInstanceRegistry(registry)
	registry.RegistryRoot = ComputeAVMContractInstanceRegistryRoot(registry)
	return registry, registry.Validate()
}

func NewAVMContractStoragePrefixDescriptor(prefix AVMContractStoragePrefixDescriptor) (AVMContractStoragePrefixDescriptor, error) {
	prefix = canonicalAVMContractStoragePrefixDescriptor(prefix)
	if prefix.Prefix == "" {
		prefix.Prefix = AVMStatePrefixContractStorage + "/" + prefix.ContractAddress + "/"
	}
	prefix.PrefixHash = ComputeAVMContractStoragePrefixHash(prefix)
	return prefix, prefix.Validate()
}

func QueryAVMContractProof(index AVMContractProofIndex, kind AVMContractProofKind, queryKey string) (AVMContractProof, error) {
	index = canonicalAVMContractProofIndex(index)
	queryKey = strings.TrimSpace(queryKey)
	if !IsAVMContractProofKind(kind) {
		return AVMContractProof{}, fmt.Errorf("invalid AVM contract proof kind %q", kind)
	}
	var root, recordHash string
	switch kind {
	case AVMContractProofCode:
		if err := index.CodeRegistry.Validate(); err != nil {
			return AVMContractProof{}, err
		}
		root = index.CodeRegistry.RegistryRoot
		record, found := findAVMContractCodeRecord(index.CodeRegistry.Codes, queryKey)
		if !found {
			return AVMContractProof{}, errors.New("AVM contract code proof record not found")
		}
		recordHash = record.RecordHash
	case AVMContractProofInstance:
		if err := index.InstanceRegistry.Validate(); err != nil {
			return AVMContractProof{}, err
		}
		root = index.InstanceRegistry.RegistryRoot
		record, found := findAVMContractInstanceRecord(index.InstanceRegistry.Instances, queryKey)
		if !found {
			return AVMContractProof{}, errors.New("AVM contract instance proof record not found")
		}
		recordHash = record.RecordHash
	case AVMContractProofStorage:
		prefix, found := findAVMContractStoragePrefix(index.StoragePrefixes, queryKey)
		if !found {
			return AVMContractProof{}, errors.New("AVM contract storage proof record not found")
		}
		if err := prefix.Validate(); err != nil {
			return AVMContractProof{}, err
		}
		root = prefix.StorageRoot
		recordHash = prefix.PrefixHash
	case AVMContractProofReceipt:
		receipt, found := findAVMContractReceipt(index.Receipts, queryKey)
		if !found {
			return AVMContractProof{}, errors.New("AVM contract receipt proof record not found")
		}
		if err := receipt.Validate(); err != nil {
			return AVMContractProof{}, err
		}
		root = ComputeAVMContractReceiptRoot(index.Receipts)
		recordHash = receipt.ReceiptHash
	}
	proof := AVMContractProof{Kind: kind, QueryKey: queryKey, Root: root, RecordHash: recordHash}
	proof.ProofHash = ComputeAVMContractProofHash(proof)
	return proof, proof.Validate()
}

func (i AVMContractBackendInterface) Validate() error {
	i = canonicalAVMContractBackendInterface(i)
	if err := validateEngineToken("AVM contract backend interface name", i.Name, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	if !IsAVMContractBackendKind(i.BackendKind) {
		return fmt.Errorf("invalid AVM contract backend interface kind %q", i.BackendKind)
	}
	if !IsRouterBackend(i.RouterBackend) {
		return fmt.Errorf("invalid AVM contract backend interface router backend %q", i.RouterBackend)
	}
	if !i.EmitsReceipt {
		return errors.New("AVM contract backend interface must emit receipts")
	}
	if !i.SupportsProofQuery {
		return errors.New("AVM contract backend interface must support proof query")
	}
	if err := zonestypes.ValidateHash("AVM contract backend interface hash", i.InterfaceHash); err != nil {
		return err
	}
	if i.BackendKind == AVMContractBackendActorContract && !i.SupportsAsync {
		return errors.New("AVM actor contract backend must support async execution")
	}
	if i.BackendInterfaceHash == "" {
		return errors.New("AVM contract backend interface commitment hash is required")
	}
	if err := zonestypes.ValidateHash("AVM contract backend interface commitment hash", i.BackendInterfaceHash); err != nil {
		return err
	}
	if i.BackendInterfaceHash != ComputeAVMContractBackendInterfaceHash(i) {
		return errors.New("AVM contract backend interface hash mismatch")
	}
	return nil
}

func (a AVMNativeModuleAdapter) Validate() error {
	a = canonicalAVMNativeModuleAdapter(a)
	if err := a.Interface.Validate(); err != nil {
		return err
	}
	if err := a.Descriptor.Validate(); err != nil {
		return err
	}
	if a.Interface.BackendKind != AVMContractBackendNativeModule || a.Interface.RouterBackend != RouterBackendNativeModule {
		return errors.New("AVM native module adapter must use native backend interface")
	}
	if a.Interface.InterfaceHash != a.Descriptor.ServiceInterfaceHash {
		return errors.New("AVM native module adapter interface hash mismatch")
	}
	return validateAdapterHash("AVM native module adapter", a.AdapterHash, ComputeAVMNativeModuleAdapterHash(a))
}

func (b AVMWASMAdapterBoundary) Validate() error {
	b = canonicalAVMWASMAdapterBoundary(b)
	if err := b.Interface.Validate(); err != nil {
		return err
	}
	if err := b.SandboxPolicy.Validate(); err != nil {
		return err
	}
	if b.Interface.BackendKind != AVMContractBackendWASMContract || b.Interface.RouterBackend != RouterBackendWASMAdapter {
		return errors.New("AVM WASM adapter boundary must use WASM backend interface")
	}
	return validateAdapterHash("AVM WASM adapter boundary", b.AdapterHash, ComputeAVMWASMAdapterBoundaryHash(b))
}

func (s AVMActorContractState) Validate() error {
	s = canonicalAVMActorContractState(s)
	if err := validateEngineToken("AVM actor contract actor id", s.ActorID, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	if s.CodeID == 0 {
		return errors.New("AVM actor contract code id must be positive")
	}
	if err := validateEngineToken("AVM actor contract owner", s.Owner, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM actor contract mailbox root", value: s.MailboxRoot},
		{name: "AVM actor contract state root", value: s.StateRoot},
		{name: "AVM actor contract continuation root", value: s.ContinuationRoot},
		{name: "AVM actor contract interface hash", value: s.InterfaceHash},
		{name: "AVM actor contract state hash", value: s.StateHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !IsAVMActorContractStatus(s.Status) {
		return fmt.Errorf("invalid AVM actor contract status %q", s.Status)
	}
	if s.StateHash != ComputeAVMActorContractStateHash(s) {
		return errors.New("AVM actor contract state hash mismatch")
	}
	return nil
}

func (r AVMActorStateRead) Validate(actorID string) error {
	if r.ActorID != actorID {
		return errors.New("AVM actor contract cannot read another actor mutable state directly")
	}
	if err := validateEngineToken("AVM actor contract state read actor id", r.ActorID, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	if !strings.HasPrefix(r.Key, ActorStateKeyPrefix(actorID)) {
		return fmt.Errorf("AVM actor contract state read key must use prefix %q", ActorStateKeyPrefix(actorID))
	}
	return zonestypes.ValidateHash("AVM actor contract state read hash", r.Hash)
}

func (e AVMActorContractExecution) Validate() error {
	e = canonicalAVMActorContractExecution(e)
	if err := e.Actor.Validate(); err != nil {
		return err
	}
	if e.Actor.Status != AVMActorContractActive {
		return errors.New("AVM actor contract execution requires active actor")
	}
	if err := e.Message.Validate(); err != nil {
		return err
	}
	if e.Message.DestinationActorOptional != e.Actor.ActorID {
		return errors.New("AVM actor contract message must target actor")
	}
	if e.ActiveMessageCount != 1 {
		return errors.New("AVM actor contract handles exactly one message at a time")
	}
	if e.DirectMutableReads {
		return errors.New("AVM actor contract cannot read another actor mutable state directly")
	}
	for i, read := range e.StateReads {
		if err := read.Validate(e.Actor.ActorID); err != nil {
			return err
		}
		if i > 0 && compareAVMActorStateReads(e.StateReads[i-1], read) >= 0 {
			return errors.New("AVM actor contract state reads must be sorted canonically")
		}
	}
	for i, write := range e.StateWrites {
		if err := write.Validate(e.Actor.ActorID); err != nil {
			return err
		}
		if i > 0 && compareActorStateWrites(e.StateWrites[i-1], write) >= 0 {
			return errors.New("AVM actor contract state writes must be sorted canonically")
		}
	}
	for _, msg := range e.EmittedMessages {
		if err := msg.Validate(); err != nil {
			return err
		}
		if msg.SourceActorOptional != e.Actor.ActorID {
			return errors.New("AVM actor contract emitted messages must use actor as source")
		}
	}
	actorMap := map[string]ActorRuntimeActor{
		e.Actor.ActorID: {ActorID: e.Actor.ActorID, CodeRef: fmt.Sprintf("code-%020d", e.Actor.CodeID), StateRoot: e.Actor.StateRoot},
	}
	for _, continuation := range e.StoredContinuations {
		if err := continuation.Validate(actorMap, e.Message.CreatedHeight); err != nil {
			return err
		}
	}
	if err := e.Receipt.Validate(); err != nil {
		return err
	}
	if e.Receipt.MessageID != e.Message.ID || e.Receipt.ZoneID != e.Message.DestinationZone {
		return errors.New("AVM actor contract receipt mismatch")
	}
	if e.ExecutionHash == "" {
		return errors.New("AVM actor contract execution hash is required")
	}
	if err := zonestypes.ValidateHash("AVM actor contract execution hash", e.ExecutionHash); err != nil {
		return err
	}
	if e.ExecutionHash != ComputeAVMActorContractExecutionHash(e) {
		return errors.New("AVM actor contract execution hash mismatch")
	}
	return nil
}

func (r AVMContractCodeRegistryRecord) Validate() error {
	r = canonicalAVMContractCodeRegistryRecord(r)
	if r.CodeID == 0 {
		return errors.New("AVM contract code registry code id must be positive")
	}
	if !IsAVMContractBackendKind(r.BackendKind) {
		return fmt.Errorf("invalid AVM contract code backend kind %q", r.BackendKind)
	}
	if err := zonestypes.ValidateHash("AVM contract code hash", r.CodeHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM contract code interface hash", r.InterfaceHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM contract code owner", r.Owner, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	if r.CodeKey != AVMContractCodeKey(r.CodeID) {
		return errors.New("AVM contract code registry key mismatch")
	}
	return validateAdapterHash("AVM contract code registry record", r.RecordHash, ComputeAVMContractCodeRecordHash(r))
}

func (r AVMContractCodeRegistry) Validate() error {
	r = canonicalAVMContractCodeRegistry(r)
	if len(r.Codes) == 0 {
		return errors.New("AVM contract code registry must contain code")
	}
	if len(r.Codes) > MaxAVMContractProofRecords {
		return fmt.Errorf("AVM contract code registry must contain <= %d records", MaxAVMContractProofRecords)
	}
	seen := make(map[uint64]struct{}, len(r.Codes))
	for i, code := range r.Codes {
		if err := code.Validate(); err != nil {
			return err
		}
		if _, found := seen[code.CodeID]; found {
			return fmt.Errorf("duplicate AVM contract code id %d", code.CodeID)
		}
		seen[code.CodeID] = struct{}{}
		if i > 0 && r.Codes[i-1].CodeID >= code.CodeID {
			return errors.New("AVM contract code registry records must be sorted canonically")
		}
	}
	return validateAdapterHash("AVM contract code registry", r.RegistryRoot, ComputeAVMContractCodeRegistryRoot(r))
}

func (r AVMContractInstanceRegistryRecord) Validate() error {
	r = canonicalAVMContractInstanceRegistryRecord(r)
	if err := validateRouterOptionalToken("AVM contract instance address", r.ContractAddress, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	if r.ContractAddress == "" {
		return errors.New("AVM contract instance address is required")
	}
	if r.CodeID == 0 {
		return errors.New("AVM contract instance code id must be positive")
	}
	if !IsAVMContractBackendKind(r.BackendKind) {
		return fmt.Errorf("invalid AVM contract instance backend kind %q", r.BackendKind)
	}
	if r.BackendKind == AVMContractBackendActorContract {
		if err := validateEngineToken("AVM contract instance actor id", r.ActorID, MaxAVMContractRuntimeToken); err != nil {
			return err
		}
	}
	if err := validateEngineToken("AVM contract instance owner", r.Owner, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	expectedPrefix := AVMStatePrefixContractStorage + "/" + r.ContractAddress + "/"
	if r.StoragePrefix != expectedPrefix {
		return fmt.Errorf("AVM contract instance storage prefix must be %q", expectedPrefix)
	}
	if err := zonestypes.ValidateHash("AVM contract instance state root", r.StateRoot); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM contract instance interface hash", r.InterfaceHash); err != nil {
		return err
	}
	if !IsAVMActorContractStatus(r.Status) {
		return fmt.Errorf("invalid AVM contract instance status %q", r.Status)
	}
	if r.InstanceKey != AVMContractInstanceKey(r.ContractAddress) {
		return errors.New("AVM contract instance registry key mismatch")
	}
	return validateAdapterHash("AVM contract instance registry record", r.RecordHash, ComputeAVMContractInstanceRecordHash(r))
}

func (r AVMContractInstanceRegistry) Validate() error {
	r = canonicalAVMContractInstanceRegistry(r)
	if len(r.Instances) == 0 {
		return errors.New("AVM contract instance registry must contain instances")
	}
	seen := make(map[string]struct{}, len(r.Instances))
	for i, instance := range r.Instances {
		if err := instance.Validate(); err != nil {
			return err
		}
		if _, found := seen[instance.ContractAddress]; found {
			return fmt.Errorf("duplicate AVM contract instance %q", instance.ContractAddress)
		}
		seen[instance.ContractAddress] = struct{}{}
		if i > 0 && r.Instances[i-1].ContractAddress >= instance.ContractAddress {
			return errors.New("AVM contract instance registry records must be sorted canonically")
		}
	}
	return validateAdapterHash("AVM contract instance registry", r.RegistryRoot, ComputeAVMContractInstanceRegistryRoot(r))
}

func (p AVMContractStoragePrefixDescriptor) Validate() error {
	p = canonicalAVMContractStoragePrefixDescriptor(p)
	if err := validateRouterOptionalToken("AVM contract storage address", p.ContractAddress, MaxAVMContractRuntimeToken); err != nil {
		return err
	}
	expected := AVMStatePrefixContractStorage + "/" + p.ContractAddress + "/"
	if p.Prefix != expected {
		return fmt.Errorf("AVM contract storage prefix must be %q", expected)
	}
	if err := zonestypes.ValidateHash("AVM contract storage root", p.StorageRoot); err != nil {
		return err
	}
	return validateAdapterHash("AVM contract storage prefix", p.PrefixHash, ComputeAVMContractStoragePrefixHash(p))
}

func (p AVMContractProof) Validate() error {
	p = canonicalAVMContractProof(p)
	if !IsAVMContractProofKind(p.Kind) {
		return fmt.Errorf("invalid AVM contract proof kind %q", p.Kind)
	}
	if p.QueryKey == "" {
		return errors.New("AVM contract proof query key is required")
	}
	if err := zonestypes.ValidateHash("AVM contract proof root", p.Root); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM contract proof record hash", p.RecordHash); err != nil {
		return err
	}
	return validateAdapterHash("AVM contract proof", p.ProofHash, ComputeAVMContractProofHash(p))
}

func IsAVMActorContractStatus(status AVMActorContractStatus) bool {
	switch status {
	case AVMActorContractActive, AVMActorContractPaused, AVMActorContractStopped:
		return true
	default:
		return false
	}
}

func IsAVMContractProofKind(kind AVMContractProofKind) bool {
	switch kind {
	case AVMContractProofCode, AVMContractProofInstance, AVMContractProofStorage, AVMContractProofReceipt:
		return true
	default:
		return false
	}
}

func ComputeAVMContractBackendInterfaceHash(i AVMContractBackendInterface) string {
	i = canonicalAVMContractBackendInterface(i)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-backend-interface-v1")
	writeEnginePart(h, i.Name)
	writeEnginePart(h, string(i.BackendKind))
	writeEnginePart(h, string(i.RouterBackend))
	writeEngineBool(h, i.SupportsSync)
	writeEngineBool(h, i.SupportsAsync)
	writeEngineBool(h, i.EmitsReceipt)
	writeEngineBool(h, i.SupportsProofQuery)
	writeEnginePart(h, i.InterfaceHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMNativeModuleAdapterHash(a AVMNativeModuleAdapter) string {
	a = canonicalAVMNativeModuleAdapter(a)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-native-module-adapter-v1")
	writeEnginePart(h, a.Interface.BackendInterfaceHash)
	writeEnginePart(h, a.Descriptor.DescriptorHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMWASMAdapterBoundaryHash(b AVMWASMAdapterBoundary) string {
	b = canonicalAVMWASMAdapterBoundary(b)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-wasm-adapter-boundary-v1")
	writeEnginePart(h, b.Interface.BackendInterfaceHash)
	writeEnginePart(h, b.SandboxPolicy.SandboxPolicyHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMActorContractStateHash(s AVMActorContractState) string {
	s = canonicalAVMActorContractState(s)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-actor-contract-state-v1")
	writeEnginePart(h, s.ActorID)
	writeEngineUint64(h, s.CodeID)
	writeEnginePart(h, s.Owner)
	writeEnginePart(h, s.MailboxRoot)
	writeEnginePart(h, s.StateRoot)
	writeEnginePart(h, s.ContinuationRoot)
	writeEnginePart(h, s.InterfaceHash)
	writeEnginePart(h, string(s.Status))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMActorContractExecutionHash(e AVMActorContractExecution) string {
	e = canonicalAVMActorContractExecution(e)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-actor-contract-execution-v1")
	writeEnginePart(h, e.Actor.StateHash)
	writeEnginePart(h, e.Message.ID)
	writeEngineUint64(h, uint64(e.ActiveMessageCount))
	writeEngineUint64(h, uint64(len(e.StateReads)))
	for _, read := range e.StateReads {
		writeEnginePart(h, read.Key)
		writeEnginePart(h, read.Hash)
	}
	writeEngineUint64(h, uint64(len(e.StateWrites)))
	for _, write := range e.StateWrites {
		writeEnginePart(h, write.Key)
		writeEnginePart(h, write.Hash)
	}
	writeEngineUint64(h, uint64(len(e.EmittedMessages)))
	for _, msg := range e.EmittedMessages {
		writeEnginePart(h, msg.ID)
	}
	writeEngineUint64(h, uint64(len(e.StoredContinuations)))
	for _, continuation := range e.StoredContinuations {
		writeEnginePart(h, continuation.ContinuationID)
	}
	writeEnginePart(h, e.Receipt.ReceiptHash)
	writeEngineBool(h, e.DirectMutableReads)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractCodeRecordHash(r AVMContractCodeRegistryRecord) string {
	r = canonicalAVMContractCodeRegistryRecord(r)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-code-record-v1")
	writeEngineUint64(h, r.CodeID)
	writeEnginePart(h, string(r.BackendKind))
	writeEnginePart(h, r.CodeHash)
	writeEnginePart(h, r.InterfaceHash)
	writeEnginePart(h, r.Owner)
	writeEngineBool(h, r.Enabled)
	writeEnginePart(h, r.CodeKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractCodeRegistryRoot(r AVMContractCodeRegistry) string {
	r = canonicalAVMContractCodeRegistry(r)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-code-registry-v1")
	writeEngineUint64(h, uint64(len(r.Codes)))
	for _, code := range r.Codes {
		writeEnginePart(h, code.RecordHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractInstanceRecordHash(r AVMContractInstanceRegistryRecord) string {
	r = canonicalAVMContractInstanceRegistryRecord(r)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-instance-record-v1")
	writeEnginePart(h, r.ContractAddress)
	writeEngineUint64(h, r.CodeID)
	writeEnginePart(h, string(r.BackendKind))
	writeEnginePart(h, r.ActorID)
	writeEnginePart(h, r.Owner)
	writeEnginePart(h, r.StoragePrefix)
	writeEnginePart(h, r.StateRoot)
	writeEnginePart(h, r.InterfaceHash)
	writeEnginePart(h, string(r.Status))
	writeEnginePart(h, r.InstanceKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractInstanceRegistryRoot(r AVMContractInstanceRegistry) string {
	r = canonicalAVMContractInstanceRegistry(r)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-instance-registry-v1")
	writeEngineUint64(h, uint64(len(r.Instances)))
	for _, instance := range r.Instances {
		writeEnginePart(h, instance.RecordHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractStoragePrefixHash(p AVMContractStoragePrefixDescriptor) string {
	p = canonicalAVMContractStoragePrefixDescriptor(p)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-storage-prefix-v1")
	writeEnginePart(h, p.ContractAddress)
	writeEnginePart(h, p.Prefix)
	writeEnginePart(h, p.StorageRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractReceiptRoot(receipts []AVMExecutionReceipt) string {
	out := append([]AVMExecutionReceipt(nil), receipts...)
	for i := range out {
		out[i] = canonicalAVMExecutionReceipt(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReceiptID < out[j].ReceiptID })
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-receipts-v1")
	writeEngineUint64(h, uint64(len(out)))
	for _, receipt := range out {
		writeEnginePart(h, receipt.ReceiptHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractProofHash(p AVMContractProof) string {
	p = canonicalAVMContractProof(p)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-proof-v1")
	writeEnginePart(h, string(p.Kind))
	writeEnginePart(h, p.QueryKey)
	writeEnginePart(h, p.Root)
	writeEnginePart(h, p.RecordHash)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMContractBackendInterface(i AVMContractBackendInterface) AVMContractBackendInterface {
	i.Name = strings.TrimSpace(i.Name)
	i.InterfaceHash = strings.TrimSpace(i.InterfaceHash)
	i.BackendInterfaceHash = strings.TrimSpace(i.BackendInterfaceHash)
	return i
}

func canonicalAVMNativeModuleAdapter(a AVMNativeModuleAdapter) AVMNativeModuleAdapter {
	a.Interface = canonicalAVMContractBackendInterface(a.Interface)
	a.Descriptor = canonicalAVMNativeModuleDescriptor(a.Descriptor)
	a.AdapterHash = strings.TrimSpace(a.AdapterHash)
	return a
}

func canonicalAVMWASMAdapterBoundary(b AVMWASMAdapterBoundary) AVMWASMAdapterBoundary {
	b.Interface = canonicalAVMContractBackendInterface(b.Interface)
	b.SandboxPolicy = canonicalAVMWASMSandboxPolicy(b.SandboxPolicy)
	b.AdapterHash = strings.TrimSpace(b.AdapterHash)
	return b
}

func canonicalAVMActorContractState(s AVMActorContractState) AVMActorContractState {
	s.ActorID = strings.TrimSpace(s.ActorID)
	s.Owner = strings.TrimSpace(s.Owner)
	s.MailboxRoot = strings.TrimSpace(s.MailboxRoot)
	s.StateRoot = strings.TrimSpace(s.StateRoot)
	s.ContinuationRoot = strings.TrimSpace(s.ContinuationRoot)
	s.InterfaceHash = strings.TrimSpace(s.InterfaceHash)
	s.StateHash = strings.TrimSpace(s.StateHash)
	return s
}

func canonicalAVMActorContractExecution(e AVMActorContractExecution) AVMActorContractExecution {
	e.Actor = canonicalAVMActorContractState(e.Actor)
	e.Message = canonicalAVMAsyncMessage(e.Message)
	e.Receipt = canonicalAVMExecutionReceipt(e.Receipt)
	e.StateReads = append([]AVMActorStateRead(nil), e.StateReads...)
	for i := range e.StateReads {
		e.StateReads[i].ActorID = strings.TrimSpace(e.StateReads[i].ActorID)
		e.StateReads[i].Key = strings.TrimSpace(e.StateReads[i].Key)
		e.StateReads[i].Hash = strings.TrimSpace(e.StateReads[i].Hash)
	}
	sort.SliceStable(e.StateReads, func(i, j int) bool { return compareAVMActorStateReads(e.StateReads[i], e.StateReads[j]) < 0 })
	e.StateWrites = append([]ActorStateWrite(nil), e.StateWrites...)
	sort.SliceStable(e.StateWrites, func(i, j int) bool { return compareActorStateWrites(e.StateWrites[i], e.StateWrites[j]) < 0 })
	e.EmittedMessages = append([]AVMAsyncMessage(nil), e.EmittedMessages...)
	for i := range e.EmittedMessages {
		e.EmittedMessages[i] = canonicalAVMAsyncMessage(e.EmittedMessages[i])
	}
	sort.SliceStable(e.EmittedMessages, func(i, j int) bool { return e.EmittedMessages[i].ID < e.EmittedMessages[j].ID })
	e.StoredContinuations = append([]ContinuationRecord(nil), e.StoredContinuations...)
	sort.SliceStable(e.StoredContinuations, func(i, j int) bool {
		return e.StoredContinuations[i].ContinuationID < e.StoredContinuations[j].ContinuationID
	})
	e.ExecutionHash = strings.TrimSpace(e.ExecutionHash)
	return e
}

func canonicalAVMContractCodeRegistryRecord(r AVMContractCodeRegistryRecord) AVMContractCodeRegistryRecord {
	r.CodeHash = strings.TrimSpace(r.CodeHash)
	r.InterfaceHash = strings.TrimSpace(r.InterfaceHash)
	r.Owner = strings.TrimSpace(r.Owner)
	r.CodeKey = strings.TrimSpace(r.CodeKey)
	r.RecordHash = strings.TrimSpace(r.RecordHash)
	return r
}

func canonicalAVMContractCodeRegistry(r AVMContractCodeRegistry) AVMContractCodeRegistry {
	r.RegistryRoot = strings.TrimSpace(r.RegistryRoot)
	r.Codes = append([]AVMContractCodeRegistryRecord(nil), r.Codes...)
	for i := range r.Codes {
		r.Codes[i] = canonicalAVMContractCodeRegistryRecord(r.Codes[i])
	}
	sort.SliceStable(r.Codes, func(i, j int) bool { return r.Codes[i].CodeID < r.Codes[j].CodeID })
	return r
}

func canonicalAVMContractInstanceRegistryRecord(r AVMContractInstanceRegistryRecord) AVMContractInstanceRegistryRecord {
	r.ContractAddress = strings.TrimSpace(r.ContractAddress)
	r.ActorID = strings.TrimSpace(r.ActorID)
	r.Owner = strings.TrimSpace(r.Owner)
	r.StoragePrefix = strings.TrimSpace(r.StoragePrefix)
	r.StateRoot = strings.TrimSpace(r.StateRoot)
	r.InterfaceHash = strings.TrimSpace(r.InterfaceHash)
	r.InstanceKey = strings.TrimSpace(r.InstanceKey)
	r.RecordHash = strings.TrimSpace(r.RecordHash)
	return r
}

func canonicalAVMContractInstanceRegistry(r AVMContractInstanceRegistry) AVMContractInstanceRegistry {
	r.RegistryRoot = strings.TrimSpace(r.RegistryRoot)
	r.Instances = append([]AVMContractInstanceRegistryRecord(nil), r.Instances...)
	for i := range r.Instances {
		r.Instances[i] = canonicalAVMContractInstanceRegistryRecord(r.Instances[i])
	}
	sort.SliceStable(r.Instances, func(i, j int) bool { return r.Instances[i].ContractAddress < r.Instances[j].ContractAddress })
	return r
}

func canonicalAVMContractStoragePrefixDescriptor(p AVMContractStoragePrefixDescriptor) AVMContractStoragePrefixDescriptor {
	p.ContractAddress = strings.TrimSpace(p.ContractAddress)
	p.Prefix = strings.TrimSpace(p.Prefix)
	p.StorageRoot = strings.TrimSpace(p.StorageRoot)
	p.PrefixHash = strings.TrimSpace(p.PrefixHash)
	return p
}

func canonicalAVMContractProof(p AVMContractProof) AVMContractProof {
	p.QueryKey = strings.TrimSpace(p.QueryKey)
	p.Root = strings.TrimSpace(p.Root)
	p.RecordHash = strings.TrimSpace(p.RecordHash)
	p.ProofHash = strings.TrimSpace(p.ProofHash)
	return p
}

func canonicalAVMContractProofIndex(index AVMContractProofIndex) AVMContractProofIndex {
	index.CodeRegistry = canonicalAVMContractCodeRegistry(index.CodeRegistry)
	index.InstanceRegistry = canonicalAVMContractInstanceRegistry(index.InstanceRegistry)
	index.StoragePrefixes = append([]AVMContractStoragePrefixDescriptor(nil), index.StoragePrefixes...)
	for i := range index.StoragePrefixes {
		index.StoragePrefixes[i] = canonicalAVMContractStoragePrefixDescriptor(index.StoragePrefixes[i])
	}
	sort.SliceStable(index.StoragePrefixes, func(i, j int) bool {
		return index.StoragePrefixes[i].ContractAddress < index.StoragePrefixes[j].ContractAddress
	})
	index.Receipts = append([]AVMExecutionReceipt(nil), index.Receipts...)
	for i := range index.Receipts {
		index.Receipts[i] = canonicalAVMExecutionReceipt(index.Receipts[i])
	}
	sort.SliceStable(index.Receipts, func(i, j int) bool { return index.Receipts[i].ReceiptID < index.Receipts[j].ReceiptID })
	return index
}

func compareAVMActorStateReads(left, right AVMActorStateRead) int {
	if left.ActorID < right.ActorID {
		return -1
	}
	if left.ActorID > right.ActorID {
		return 1
	}
	if left.Key < right.Key {
		return -1
	}
	if left.Key > right.Key {
		return 1
	}
	return 0
}

func findAVMContractCodeRecord(records []AVMContractCodeRegistryRecord, queryKey string) (AVMContractCodeRegistryRecord, bool) {
	for _, record := range records {
		if record.CodeKey == queryKey || fmt.Sprintf("%d", record.CodeID) == queryKey {
			return record, true
		}
	}
	return AVMContractCodeRegistryRecord{}, false
}

func findAVMContractInstanceRecord(records []AVMContractInstanceRegistryRecord, queryKey string) (AVMContractInstanceRegistryRecord, bool) {
	for _, record := range records {
		if record.InstanceKey == queryKey || record.ContractAddress == queryKey {
			return record, true
		}
	}
	return AVMContractInstanceRegistryRecord{}, false
}

func findAVMContractStoragePrefix(records []AVMContractStoragePrefixDescriptor, queryKey string) (AVMContractStoragePrefixDescriptor, bool) {
	for _, record := range records {
		if record.Prefix == queryKey || record.ContractAddress == queryKey {
			return record, true
		}
	}
	return AVMContractStoragePrefixDescriptor{}, false
}

func findAVMContractReceipt(receipts []AVMExecutionReceipt, queryKey string) (AVMExecutionReceipt, bool) {
	for _, receipt := range receipts {
		if receipt.ReceiptID == queryKey || receipt.MessageID == queryKey {
			return receipt, true
		}
	}
	return AVMExecutionReceipt{}, false
}

func validateAdapterHash(name, actual, expected string) error {
	if actual == "" {
		return fmt.Errorf("%s hash is required", name)
	}
	if err := zonestypes.ValidateHash(name+" hash", actual); err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("%s hash mismatch", name)
	}
	return nil
}
