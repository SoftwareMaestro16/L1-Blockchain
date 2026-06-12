package types

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

const (
	ContractZonePrefix	= "contract"
	ContractCodePrefix	= ContractZonePrefix + "/code"
	ContractInstancePrefix	= ContractZonePrefix + "/instance"
	ContractStoragePrefix	= ContractZonePrefix + "/storage"
	ContractABIPrefix	= ContractZonePrefix + "/abi"
	ContractInboxPrefix	= ContractZonePrefix + "/inbox"
	ContractOutboxPrefix	= ContractZonePrefix + "/outbox"
	ContractReceiptPrefix	= ContractZonePrefix + "/receipts"
	ContractGasTablePrefix	= ContractZonePrefix + "/gas"
	ContractEventPrefix	= ContractZonePrefix + "/events"
)

type ContractRuntimeKind string
type ContractMessageKind string
type ContractProofKind string
type ContractShardRoutingMode string
type ContractProofRootType string

const (
	ContractRuntimeAVM	ContractRuntimeKind	= "AVM"
	ContractRuntimeCosmWasm	ContractRuntimeKind	= "COSMWASM"

	ContractMessageStoreCode	ContractMessageKind	= "MsgStoreCode"
	ContractMessageInstantiate	ContractMessageKind	= "MsgInstantiateContract"
	ContractMessageExecute		ContractMessageKind	= "MsgExecuteContract"
	ContractMessageMigrate		ContractMessageKind	= "MsgMigrateContract"
	ContractMessageCallback		ContractMessageKind	= "MsgContractCallback"
	ContractMessageProofVerify	ContractMessageKind	= "MsgContractProofVerify"
	ContractMessageAsyncCall	ContractMessageKind	= "MsgAsyncContractCall"

	ContractProofCode	ContractProofKind	= "QueryCode"
	ContractProofContract	ContractProofKind	= "QueryContract"
	ContractProofState	ContractProofKind	= "QueryContractState"
	ContractProofABI	ContractProofKind	= "QueryContractABI"
	ContractProofReceipt	ContractProofKind	= "QueryContractReceipt"
	ContractProofGasTable	ContractProofKind	= "QueryContractGasTable"
	ContractProofEvent	ContractProofKind	= "QueryContractEvent"
	ContractProofInbox	ContractProofKind	= "QueryContractInbox"
	ContractProofOutbox	ContractProofKind	= "QueryContractOutbox"

	ContractRouteAddress		ContractShardRoutingMode	= "contract_address"
	ContractRouteStorageKey		ContractShardRoutingMode	= "storage_key_prefix"
	ContractRouteCodeRegistry	ContractShardRoutingMode	= "code_id"

	ContractProofRootCode		ContractProofRootType	= "contract_code"
	ContractProofRootInstance	ContractProofRootType	= "contract_instance"
	ContractProofRootStorage	ContractProofRootType	= "contract_storage"
	ContractProofRootABI		ContractProofRootType	= "contract_abi"
	ContractProofRootGasTable	ContractProofRootType	= "contract_gas_table"
	ContractProofRootEvent		ContractProofRootType	= "contract_event"
	ContractProofRootInbox		ContractProofRootType	= "contract_inbox"
	ContractProofRootOutbox		ContractProofRootType	= "contract_outbox"
)

type ContractBytecodeRuntime interface {
	RuntimeID() string
	ValidateBytecode(context.Context, ContractCode) error
	ExecuteContractMessage(context.Context, ContractInstance, ContractInboxMessage) (ContractExecutionReceipt, []ContractStorageEntry, []ContractInboxMessage, error)
	ComputeBytecodeRoot(context.Context, []ContractCode) (string, error)
	ComputeContractStateRoot(context.Context, ContractInstance, []ContractStorageEntry) (string, error)
}

type AVMBytecodeRuntime = ContractBytecodeRuntime

type CosmWasmAdapterBoundary interface {
	AdapterID() string
	ValidateCosmWasmCode(context.Context, ContractCode) error
	TranslateCosmWasmMessage(context.Context, ContractInboxMessage) (ContractInboxMessage, error)
	AdapterStateRoot(context.Context) (string, error)
}

type ContractZoneBoundary struct {
	ZoneID		ZoneID
	OwnsPrefixes	[]string
	Messages	[]ContractMessageKind
	ProofKinds	[]ContractProofKind
}

type AVMInstructionGasCost struct {
	Instruction	string
	BaseGas		uint64
}

type AVMGasTable struct {
	Version		uint64
	Costs		[]AVMInstructionGasCost
	TableHash	string
}

type AVMInstruction struct {
	Opcode		string
	OperandHash	string
}

type AVMExecutionKeeper struct {
	RuntimeID	string
	GasTable	AVMGasTable
	ZoneGasLimit	uint64
	MaxInstructions	uint32
	KeeperHash	string
}

type ContractShardRoute struct {
	ZoneID		ZoneID
	LayoutEpoch	uint64
	ShardCount	uint32
	ShardID		uint32
	RoutingMode	ContractShardRoutingMode
	RouteKey	string
	StateKey	string
	RouteHash	string
}

type ContractAsyncCall struct {
	CallID		string
	Source		string
	SourceZone	ZoneID
	TargetContract	string
	PayloadHash	string
	GasLimit	uint64
	RetryNonce	uint64
	ExpiryHeight	uint64
	CreatedHeight	uint64
	CallHash	string
}

type ContractEvent struct {
	ContractAddr	string
	EventID		string
	Topic		string
	PayloadHash	string
	Height		uint64
	Sequence	uint64
	EventHash	string
}

type ContractProofRootExport struct {
	ZoneID		ZoneID
	Height		uint64
	RootType	ContractProofRootType
	RootHash	string
	Source		string
}

type ContractBytecodeInterface struct {
	Runtime		ContractRuntimeKind
	InstructionSet	string
	BytecodeHash	string
	ABIHash		string
	DeterminismHash	string
	MaxCodeBytes	uint64
	InterfaceHash	string
}

type ContractCosmWasmAdapterDescriptor struct {
	AdapterID	string
	Version		string
	PolicyHash	string
	CapabilityRoot	string
	DescriptorHash	string
}

type ContractCode struct {
	CodeID		string
	Runtime		ContractRuntimeKind
	BytecodeHash	string
	BytecodeSize	uint64
	ABIHash		string
	InterfaceHash	string
	Uploader	string
	UploadedHeight	uint64
}

type ContractInstance struct {
	ContractAddr	string
	CodeID		string
	Runtime		ContractRuntimeKind
	Admin		string
	StorageRoot	string
	CreatedHeight	uint64
	UpdatedHeight	uint64
}

type ContractStorageEntry struct {
	ContractAddr	string
	Key		string
	ValueHash	string
}

type ContractABIDescriptor struct {
	CodeID			string
	ABIHash			string
	InterfaceHash		string
	ExportedMethods		[]string
	RegisteredHeight	uint64
}

type ContractInboxMessage struct {
	ContractAddr	string
	MsgID		string
	MessageKind	ContractMessageKind
	Source		string
	PayloadHash	string
	GasLimit	uint64
	Sequence	uint64
	ReceivedHeight	uint64
}

type ContractExecutionReceipt struct {
	ContractAddr	string
	ReceiptID	string
	MsgID		string
	MessageKind	ContractMessageKind
	Status		ZoneReceiptStatus
	GasUsed		uint64
	OutputHash	string
	StorageRoot	string
	EmittedMessages	uint32
	Height		uint64
	Sequence	uint64
	ReceiptHash	string
}

type ContractZoneState struct {
	Height			uint64
	Codes			[]ContractCode
	Instances		[]ContractInstance
	Storage			[]ContractStorageEntry
	ABIs			[]ContractABIDescriptor
	Inbox			[]ContractInboxMessage
	Outbox			[]ContractInboxMessage
	Receipts		[]ContractExecutionReceipt
	BytecodeInterfaces	[]ContractBytecodeInterface
	CosmWasmAdapters	[]ContractCosmWasmAdapterDescriptor
	GasTable		AVMGasTable
	Events			[]ContractEvent
}

type ContractZoneRoots struct {
	Height			uint64
	CodeRoot		string
	InstanceRoot		string
	StorageRoot		string
	ABIRoot			string
	InboxRoot		string
	OutboxRoot		string
	ReceiptRoot		string
	GasTableRoot		string
	EventRoot		string
	BytecodeInterfaceRoot	string
	CosmWasmAdapterRoot	string
	ExecutionRoot		string
	ProofRoot		string
	StateRoot		string
}

func DefaultContractZoneBoundary() ContractZoneBoundary {
	return ContractZoneBoundary{
		ZoneID:	ZoneIDContract,
		OwnsPrefixes: []string{
			ContractABIPrefix,
			ContractCodePrefix,
			ContractEventPrefix,
			ContractGasTablePrefix,
			ContractInboxPrefix,
			ContractInstancePrefix,
			ContractOutboxPrefix,
			ContractReceiptPrefix,
			ContractStoragePrefix,
		},
		Messages: []ContractMessageKind{
			ContractMessageStoreCode,
			ContractMessageInstantiate,
			ContractMessageExecute,
			ContractMessageMigrate,
			ContractMessageCallback,
			ContractMessageProofVerify,
			ContractMessageAsyncCall,
		},
		ProofKinds: []ContractProofKind{
			ContractProofCode,
			ContractProofContract,
			ContractProofState,
			ContractProofABI,
			ContractProofReceipt,
			ContractProofGasTable,
			ContractProofEvent,
			ContractProofInbox,
			ContractProofOutbox,
		},
	}
}

func (b ContractZoneBoundary) Validate() error {
	if b.ZoneID != ZoneIDContract {
		return errors.New("contract zone boundary must use CONTRACT_ZONE")
	}
	if len(b.OwnsPrefixes) == 0 || len(b.Messages) == 0 || len(b.ProofKinds) == 0 {
		return errors.New("contract zone boundary requires prefixes, messages, and proof kinds")
	}
	for i, prefix := range b.OwnsPrefixes {
		if err := validateRuntimeToken("contract zone prefix", prefix, MaxZoneNamespaceLength); err != nil {
			return err
		}
		if i > 0 && b.OwnsPrefixes[i-1] >= prefix {
			return errors.New("contract zone prefixes must be sorted canonically")
		}
	}
	for _, msg := range b.Messages {
		if !IsContractMessageKind(msg) {
			return fmt.Errorf("unknown contract message kind %q", msg)
		}
	}
	for _, proof := range b.ProofKinds {
		if !IsContractProofKind(proof) {
			return fmt.Errorf("unknown contract proof kind %q", proof)
		}
	}
	return nil
}

func ContractCodeKey(codeID string) (string, error) {
	if err := validateRuntimeToken("contract code id", codeID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractCodePrefix + "/" + codeID, nil
}

func ContractInstanceKey(contractAddr string) (string, error) {
	if err := validateRuntimeToken("contract address", contractAddr, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractInstancePrefix + "/" + contractAddr, nil
}

func ContractStorageKey(contractAddr, key string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract storage key", key, MaxZoneProofKeyLength); err != nil {
		return "", err
	}
	return ContractStoragePrefix + "/" + contractAddr + "/" + key, nil
}

func ContractABIKey(codeID string) (string, error) {
	if _, err := ContractCodeKey(codeID); err != nil {
		return "", err
	}
	return ContractABIPrefix + "/" + codeID, nil
}

func ContractInboxKey(contractAddr, msgID string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract inbox message id", msgID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractInboxPrefix + "/" + contractAddr + "/" + msgID, nil
}

func ContractReceiptKey(contractAddr, receiptID string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract receipt id", receiptID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractReceiptPrefix + "/" + contractAddr + "/" + receiptID, nil
}

func ContractOutboxKey(contractAddr, msgID string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract outbox message id", msgID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractOutboxPrefix + "/" + contractAddr + "/" + msgID, nil
}

func ContractGasTableKey(version uint64) (string, error) {
	if version == 0 {
		return "", errors.New("contract gas table version must be positive")
	}
	return fmt.Sprintf("%s/%020d", ContractGasTablePrefix, version), nil
}

func ContractEventKey(contractAddr, eventID string) (string, error) {
	if _, err := ContractInstanceKey(contractAddr); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("contract event id", eventID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ContractEventPrefix + "/" + contractAddr + "/" + eventID, nil
}

func RouteContractInstanceShard(contractAddr string, shardCount uint32, layoutEpoch uint64) (ContractShardRoute, error) {
	key, err := ContractInstanceKey(contractAddr)
	if err != nil {
		return ContractShardRoute{}, err
	}
	return routeContractStateKey(ContractRouteAddress, contractAddr, key, shardCount, layoutEpoch)
}

func RouteContractStorageShard(contractAddr, storageKeyPrefix string, shardCount uint32, layoutEpoch uint64) (ContractShardRoute, error) {
	key, err := ContractStorageKey(contractAddr, storageKeyPrefix)
	if err != nil {
		return ContractShardRoute{}, err
	}
	return routeContractStateKey(ContractRouteStorageKey, contractAddr+"/"+storageKeyPrefix, key, shardCount, layoutEpoch)
}

func RouteContractCodeShard(codeID string, shardCount uint32, layoutEpoch uint64) (ContractShardRoute, error) {
	key, err := ContractCodeKey(codeID)
	if err != nil {
		return ContractShardRoute{}, err
	}
	return routeContractStateKey(ContractRouteCodeRegistry, codeID, key, shardCount, layoutEpoch)
}

func NewContractBytecodeInterface(iface ContractBytecodeInterface) (ContractBytecodeInterface, error) {
	if iface.InterfaceHash != "" {
		return ContractBytecodeInterface{}, errors.New("contract bytecode interface hash must be empty before construction")
	}
	if err := iface.ValidateFormat(); err != nil {
		return ContractBytecodeInterface{}, err
	}
	iface.InterfaceHash = ComputeContractBytecodeInterfaceHash(iface)
	return iface, iface.Validate()
}

func DefaultAVMGasTable(version uint64) (AVMGasTable, error) {
	table := AVMGasTable{
		Version:	version,
		Costs: []AVMInstructionGasCost{
			{Instruction: "ABI_EXPORT", BaseGas: 3},
			{Instruction: "ADD", BaseGas: 1},
			{Instruction: "CTX_GAS_LEFT", BaseGas: 1},
			{Instruction: "KV_GET", BaseGas: 25},
			{Instruction: "KV_SET", BaseGas: 50},
			{Instruction: "MSG_SEND", BaseGas: 100},
			{Instruction: "PUSH", BaseGas: 1},
			{Instruction: "RET", BaseGas: 1},
			{Instruction: "VERIFY_MERKLE_PROOF", BaseGas: 250},
		},
	}
	table.TableHash = ComputeAVMGasTableHash(table)
	return table, table.ValidateHash()
}

func NewAVMExecutionKeeper(runtimeID string, gasTable AVMGasTable, zoneGasLimit uint64, maxInstructions uint32) (AVMExecutionKeeper, error) {
	keeper := AVMExecutionKeeper{
		RuntimeID:		runtimeID,
		GasTable:		gasTable,
		ZoneGasLimit:		zoneGasLimit,
		MaxInstructions:	maxInstructions,
	}
	keeper.KeeperHash = ComputeAVMExecutionKeeperHash(keeper)
	return keeper, keeper.ValidateHash()
}

func (k AVMExecutionKeeper) Execute(instance ContractInstance, msg ContractInboxMessage, instructions []AVMInstruction, height uint64, sequence uint64) (ContractExecutionReceipt, error) {
	if err := k.ValidateHash(); err != nil {
		return ContractExecutionReceipt{}, err
	}
	if err := instance.Validate(); err != nil {
		return ContractExecutionReceipt{}, err
	}
	if err := msg.Validate(); err != nil {
		return ContractExecutionReceipt{}, err
	}
	if instance.ContractAddr != msg.ContractAddr {
		return ContractExecutionReceipt{}, errors.New("AVM execution message contract mismatch")
	}
	if height == 0 || sequence == 0 {
		return ContractExecutionReceipt{}, errors.New("AVM execution height and sequence must be positive")
	}
	if len(instructions) == 0 || uint32(len(instructions)) > k.MaxInstructions {
		return ContractExecutionReceipt{}, errors.New("AVM execution instruction count outside bounds")
	}
	gasUsed, instructionRoot, err := k.MeterInstructions(instructions)
	if err != nil {
		return ContractExecutionReceipt{}, err
	}
	if gasUsed > msg.GasLimit || gasUsed > k.ZoneGasLimit {
		return ContractExecutionReceipt{}, errors.New("AVM execution gas exceeds limit")
	}
	return NewContractExecutionReceipt(ContractExecutionReceipt{
		ContractAddr:		instance.ContractAddr,
		ReceiptID:		hashRuntimeParts("contract-avm-receipt-id-v1", msg.MsgID, fmt.Sprint(height), fmt.Sprint(sequence)),
		MsgID:			msg.MsgID,
		MessageKind:		msg.MessageKind,
		Status:			ZoneReceiptStatusSuccess,
		GasUsed:		gasUsed,
		OutputHash:		instructionRoot,
		StorageRoot:		instance.StorageRoot,
		EmittedMessages:	0,
		Height:			height,
		Sequence:		sequence,
	})
}

func (k AVMExecutionKeeper) MeterInstructions(instructions []AVMInstruction) (uint64, string, error) {
	var gasUsed uint64
	parts := []string{"aetra-avm-instruction-trace-v1", fmt.Sprint(len(instructions))}
	for _, instruction := range instructions {
		if err := instruction.Validate(); err != nil {
			return 0, "", err
		}
		cost, found := k.GasTable.CostFor(instruction.Opcode)
		if !found {
			return 0, "", fmt.Errorf("AVM instruction %q missing gas cost", instruction.Opcode)
		}
		next, err := addZoneGas(gasUsed, cost)
		if err != nil {
			return 0, "", err
		}
		gasUsed = next
		parts = append(parts, instruction.Opcode, instruction.OperandHash, fmt.Sprint(cost))
	}
	return gasUsed, hashRuntimeParts(parts...), nil
}

func NewContractCosmWasmAdapterDescriptor(adapter ContractCosmWasmAdapterDescriptor) (ContractCosmWasmAdapterDescriptor, error) {
	if adapter.DescriptorHash != "" {
		return ContractCosmWasmAdapterDescriptor{}, errors.New("contract CosmWasm adapter descriptor hash must be empty before construction")
	}
	if err := adapter.ValidateFormat(); err != nil {
		return ContractCosmWasmAdapterDescriptor{}, err
	}
	adapter.DescriptorHash = ComputeContractCosmWasmAdapterHash(adapter)
	return adapter, adapter.Validate()
}

func NewContractExecutionReceipt(receipt ContractExecutionReceipt) (ContractExecutionReceipt, error) {
	if receipt.ReceiptHash != "" {
		return ContractExecutionReceipt{}, errors.New("contract receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return ContractExecutionReceipt{}, err
	}
	receipt.ReceiptHash = ComputeContractExecutionReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func UpsertContractStorage(entries []ContractStorageEntry, update ContractStorageEntry) ([]ContractStorageEntry, error) {
	if err := update.Validate(); err != nil {
		return nil, err
	}
	next := append([]ContractStorageEntry(nil), entries...)
	for i, entry := range next {
		if entry.ContractAddr == update.ContractAddr && entry.Key == update.Key {
			next[i] = update
			return normalizeContractStorage(next), nil
		}
	}
	next = append(next, update)
	return normalizeContractStorage(next), nil
}

func EnqueueContractInbox(inbox []ContractInboxMessage, msg ContractInboxMessage) ([]ContractInboxMessage, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	next := append([]ContractInboxMessage(nil), inbox...)
	for _, existing := range next {
		if existing.ContractAddr != msg.ContractAddr {
			continue
		}
		if existing.MsgID == msg.MsgID {
			return nil, errors.New("contract inbox message id already exists")
		}
		if existing.Sequence == msg.Sequence {
			return nil, errors.New("contract inbox sequence already exists")
		}
	}
	next = append(next, msg)
	return normalizeContractInbox(next), nil
}

func EnqueueContractOutbox(outbox []ContractInboxMessage, msg ContractInboxMessage) ([]ContractInboxMessage, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	next := append([]ContractInboxMessage(nil), outbox...)
	for _, existing := range next {
		if existing.ContractAddr == msg.ContractAddr && existing.MsgID == msg.MsgID {
			return nil, errors.New("contract outbox message id already exists")
		}
	}
	next = append(next, msg)
	return normalizeContractInbox(next), nil
}

func NewContractAsyncCall(call ContractAsyncCall) (ContractAsyncCall, ContractInboxMessage, error) {
	if call.CallID != "" || call.CallHash != "" {
		return ContractAsyncCall{}, ContractInboxMessage{}, errors.New("contract async call hash fields must be empty before construction")
	}
	if err := call.ValidateFormat(); err != nil {
		return ContractAsyncCall{}, ContractInboxMessage{}, err
	}
	call.CallID = ComputeContractAsyncCallID(call)
	call.CallHash = ComputeContractAsyncCallHash(call)
	if err := call.ValidateHash(); err != nil {
		return ContractAsyncCall{}, ContractInboxMessage{}, err
	}
	msg := ContractInboxMessage{
		ContractAddr:	call.TargetContract,
		MsgID:		call.CallID,
		MessageKind:	ContractMessageAsyncCall,
		Source:		string(call.SourceZone) + ":" + call.Source,
		PayloadHash:	call.CallHash,
		GasLimit:	call.GasLimit,
		Sequence:	call.RetryNonce,
		ReceivedHeight:	call.CreatedHeight,
	}
	return call, msg, msg.Validate()
}

func ContractProofRequest(kind ContractProofKind, key string, height uint64, root string, limit uint32) (ZoneProofRequest, error) {
	if !IsContractProofKind(kind) {
		return ZoneProofRequest{}, fmt.Errorf("unknown contract proof kind %q", kind)
	}
	if err := validateRuntimeToken("contract proof key", key, MaxZoneProofKeyLength); err != nil {
		return ZoneProofRequest{}, err
	}
	req := ZoneProofRequest{
		ZoneID:	ZoneIDContract,
		Height:	height,
		Kind:	ZoneProofKindState,
		Key:	string(kind) + "/" + key,
		Root:	root,
		Limit:	limit,
	}
	return req, req.Validate()
}

func BuildContractZoneRoot(roots ContractZoneRoots) (ZoneRoot, error) {
	if err := roots.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	stateRoot := roots.StateRoot
	if stateRoot == "" {
		stateRoot = hashRuntimeParts(
			"aetra-contract-zone-state-v1",
			roots.CodeRoot,
			roots.InstanceRoot,
			roots.StorageRoot,
			roots.ABIRoot,
			roots.InboxRoot,
			roots.OutboxRoot,
			roots.ReceiptRoot,
			roots.GasTableRoot,
			roots.EventRoot,
			roots.BytecodeInterfaceRoot,
			roots.CosmWasmAdapterRoot,
		)
	}
	root := ZoneRoot{
		ZoneID:			ZoneIDContract,
		Height:			roots.Height,
		ZoneStateRoot:		stateRoot,
		InboxRoot:		roots.InboxRoot,
		OutboxRoot:		roots.OutboxRoot,
		ReceiptRoot:		roots.ReceiptRoot,
		EventRoot:		roots.EventRoot,
		ExecutionResultRoot:	roots.ExecutionRoot,
		ProofRoot:		roots.ProofRoot,
	}
	root.RootHash = ComputeZoneRootHash(root)
	return root, root.Validate()
}

func BuildContractZoneRootFromState(height uint64, state ContractZoneState, proofRoot string) (ZoneRoot, error) {
	normalized := state.Normalize()
	roots := ContractZoneRoots{
		Height:			height,
		CodeRoot:		ComputeContractCodeRoot(normalized.Codes),
		InstanceRoot:		ComputeContractInstanceRoot(normalized.Instances),
		StorageRoot:		ComputeContractStorageRoot(normalized.Storage),
		ABIRoot:		ComputeContractABIRoot(normalized.ABIs),
		InboxRoot:		ComputeContractInboxRoot(normalized.Inbox),
		OutboxRoot:		ComputeContractOutboxRoot(normalized.Outbox),
		ReceiptRoot:		ComputeContractReceiptRoot(normalized.Receipts),
		GasTableRoot:		ComputeAVMGasTableRoot(normalized.GasTable),
		EventRoot:		ComputeContractEventRoot(normalized.Events),
		BytecodeInterfaceRoot:	ComputeContractBytecodeInterfaceRoot(normalized.BytecodeInterfaces),
		CosmWasmAdapterRoot:	ComputeContractCosmWasmAdapterRoot(normalized.CosmWasmAdapters),
		ExecutionRoot:		ComputeContractExecutionRoot(normalized.Receipts),
		ProofRoot:		proofRoot,
		StateRoot:		ComputeContractZoneStateRoot(normalized),
	}
	return BuildContractZoneRoot(roots)
}

func (s ContractZoneState) Normalize() ContractZoneState {
	s.Codes = normalizeContractCodes(s.Codes)
	s.Instances = normalizeContractInstances(s.Instances)
	s.Storage = normalizeContractStorage(s.Storage)
	s.ABIs = normalizeContractABIs(s.ABIs)
	s.Inbox = normalizeContractInbox(s.Inbox)
	s.Outbox = normalizeContractInbox(s.Outbox)
	s.Receipts = normalizeContractReceipts(s.Receipts)
	s.BytecodeInterfaces = normalizeContractBytecodeInterfaces(s.BytecodeInterfaces)
	s.CosmWasmAdapters = normalizeContractCosmWasmAdapters(s.CosmWasmAdapters)
	s.GasTable.Costs = normalizeAVMInstructionGasCosts(s.GasTable.Costs)
	s.Events = normalizeContractEvents(s.Events)
	return s
}

func (s ContractZoneState) Validate() error {
	normalized := s.Normalize()
	for _, code := range normalized.Codes {
		if err := code.Validate(); err != nil {
			return err
		}
	}
	for _, instance := range normalized.Instances {
		if err := instance.Validate(); err != nil {
			return err
		}
	}
	for _, entry := range normalized.Storage {
		if err := entry.Validate(); err != nil {
			return err
		}
	}
	for _, abi := range normalized.ABIs {
		if err := abi.Validate(); err != nil {
			return err
		}
	}
	for _, msg := range normalized.Inbox {
		if err := msg.Validate(); err != nil {
			return err
		}
	}
	for _, msg := range normalized.Outbox {
		if err := msg.Validate(); err != nil {
			return err
		}
	}
	for _, receipt := range normalized.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
	}
	for _, iface := range normalized.BytecodeInterfaces {
		if err := iface.Validate(); err != nil {
			return err
		}
	}
	for _, adapter := range normalized.CosmWasmAdapters {
		if err := adapter.Validate(); err != nil {
			return err
		}
	}
	if normalized.GasTable.Version != 0 {
		if err := normalized.GasTable.ValidateHash(); err != nil {
			return err
		}
	}
	for _, event := range normalized.Events {
		if err := event.ValidateHash(); err != nil {
			return err
		}
	}
	return nil
}

func (i ContractBytecodeInterface) ValidateFormat() error {
	if !IsContractRuntimeKind(i.Runtime) {
		return fmt.Errorf("unknown contract runtime kind %q", i.Runtime)
	}
	if err := validateRuntimeToken("contract instruction set", i.InstructionSet, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract bytecode interface bytecode hash", i.BytecodeHash); err != nil {
		return err
	}
	if err := ValidateHash("contract bytecode interface ABI hash", i.ABIHash); err != nil {
		return err
	}
	if err := ValidateHash("contract bytecode interface determinism hash", i.DeterminismHash); err != nil {
		return err
	}
	if i.MaxCodeBytes == 0 {
		return errors.New("contract bytecode interface max code bytes must be positive")
	}
	if i.InterfaceHash != "" {
		return ValidateHash("contract bytecode interface hash", i.InterfaceHash)
	}
	return nil
}

func (i ContractBytecodeInterface) Validate() error {
	if err := i.ValidateFormat(); err != nil {
		return err
	}
	if i.InterfaceHash == "" {
		return errors.New("contract bytecode interface hash is required")
	}
	if i.InterfaceHash != ComputeContractBytecodeInterfaceHash(i) {
		return errors.New("contract bytecode interface hash mismatch")
	}
	return nil
}

func (a ContractCosmWasmAdapterDescriptor) ValidateFormat() error {
	if err := validateRuntimeToken("contract CosmWasm adapter id", a.AdapterID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract CosmWasm adapter version", a.Version, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract CosmWasm adapter policy hash", a.PolicyHash); err != nil {
		return err
	}
	if err := ValidateHash("contract CosmWasm adapter capability root", a.CapabilityRoot); err != nil {
		return err
	}
	if a.DescriptorHash != "" {
		return ValidateHash("contract CosmWasm adapter descriptor hash", a.DescriptorHash)
	}
	return nil
}

func (a ContractCosmWasmAdapterDescriptor) Validate() error {
	if err := a.ValidateFormat(); err != nil {
		return err
	}
	if a.DescriptorHash == "" {
		return errors.New("contract CosmWasm adapter descriptor hash is required")
	}
	if a.DescriptorHash != ComputeContractCosmWasmAdapterHash(a) {
		return errors.New("contract CosmWasm adapter descriptor hash mismatch")
	}
	return nil
}

func (c AVMInstructionGasCost) Validate() error {
	if err := validateRuntimeToken("AVM instruction", c.Instruction, MaxZoneEndpointLength); err != nil {
		return err
	}
	if c.BaseGas == 0 {
		return errors.New("AVM instruction gas cost must be positive")
	}
	return nil
}

func (t AVMGasTable) ValidateHash() error {
	if t.Version == 0 {
		return errors.New("AVM gas table version must be positive")
	}
	if _, err := ContractGasTableKey(t.Version); err != nil {
		return err
	}
	ordered := normalizeAVMInstructionGasCosts(t.Costs)
	if len(ordered) == 0 {
		return errors.New("AVM gas table requires costs")
	}
	for i, cost := range ordered {
		if err := cost.Validate(); err != nil {
			return err
		}
		if i > 0 && ordered[i-1].Instruction >= cost.Instruction {
			return errors.New("AVM gas table costs must be sorted canonically")
		}
	}
	if err := ValidateHash("AVM gas table hash", t.TableHash); err != nil {
		return err
	}
	if t.TableHash != ComputeAVMGasTableHash(t) {
		return errors.New("AVM gas table hash mismatch")
	}
	return nil
}

func (t AVMGasTable) CostFor(opcode string) (uint64, bool) {
	for _, cost := range normalizeAVMInstructionGasCosts(t.Costs) {
		if cost.Instruction == opcode {
			return cost.BaseGas, true
		}
	}
	return 0, false
}

func (i AVMInstruction) Validate() error {
	if err := validateRuntimeToken("AVM opcode", i.Opcode, MaxZoneEndpointLength); err != nil {
		return err
	}
	return ValidateHash("AVM instruction operand hash", i.OperandHash)
}

func (k AVMExecutionKeeper) ValidateHash() error {
	if err := validateRuntimeToken("AVM execution keeper runtime id", k.RuntimeID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := k.GasTable.ValidateHash(); err != nil {
		return err
	}
	if k.ZoneGasLimit == 0 || k.MaxInstructions == 0 {
		return errors.New("AVM execution keeper requires gas and instruction bounds")
	}
	if err := ValidateHash("AVM execution keeper hash", k.KeeperHash); err != nil {
		return err
	}
	if k.KeeperHash != ComputeAVMExecutionKeeperHash(k) {
		return errors.New("AVM execution keeper hash mismatch")
	}
	return nil
}

func (c ContractCode) Validate() error {
	if _, err := ContractCodeKey(c.CodeID); err != nil {
		return err
	}
	if !IsContractRuntimeKind(c.Runtime) {
		return fmt.Errorf("unknown contract runtime kind %q", c.Runtime)
	}
	if err := ValidateHash("contract bytecode hash", c.BytecodeHash); err != nil {
		return err
	}
	if c.BytecodeSize == 0 {
		return errors.New("contract bytecode size must be positive")
	}
	if err := ValidateHash("contract ABI hash", c.ABIHash); err != nil {
		return err
	}
	if err := ValidateHash("contract interface hash", c.InterfaceHash); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract code uploader", c.Uploader, MaxZoneEndpointLength); err != nil {
		return err
	}
	if c.UploadedHeight == 0 {
		return errors.New("contract code uploaded height must be positive")
	}
	return nil
}

func (i ContractInstance) Validate() error {
	if _, err := ContractInstanceKey(i.ContractAddr); err != nil {
		return err
	}
	if _, err := ContractCodeKey(i.CodeID); err != nil {
		return err
	}
	if !IsContractRuntimeKind(i.Runtime) {
		return fmt.Errorf("unknown contract runtime kind %q", i.Runtime)
	}
	if err := validateRuntimeToken("contract admin", i.Admin, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract instance storage root", i.StorageRoot); err != nil {
		return err
	}
	if i.CreatedHeight == 0 || i.UpdatedHeight == 0 {
		return errors.New("contract instance heights must be positive")
	}
	if i.UpdatedHeight < i.CreatedHeight {
		return errors.New("contract instance updated height must not precede created height")
	}
	return nil
}

func (e ContractStorageEntry) Validate() error {
	if _, err := ContractStorageKey(e.ContractAddr, e.Key); err != nil {
		return err
	}
	return ValidateHash("contract storage value hash", e.ValueHash)
}

func (a ContractABIDescriptor) Validate() error {
	if _, err := ContractABIKey(a.CodeID); err != nil {
		return err
	}
	if err := ValidateHash("contract ABI hash", a.ABIHash); err != nil {
		return err
	}
	if err := ValidateHash("contract interface hash", a.InterfaceHash); err != nil {
		return err
	}
	if a.RegisteredHeight == 0 {
		return errors.New("contract ABI registered height must be positive")
	}
	methods := append([]string(nil), a.ExportedMethods...)
	sort.Strings(methods)
	for i, method := range methods {
		if err := validateRuntimeToken("contract ABI method", method, MaxZoneEndpointLength); err != nil {
			return err
		}
		if i > 0 && methods[i-1] == method {
			return errors.New("contract ABI methods must be unique")
		}
	}
	return nil
}

func (m ContractInboxMessage) Validate() error {
	if _, err := ContractInboxKey(m.ContractAddr, m.MsgID); err != nil {
		return err
	}
	if !IsContractMessageKind(m.MessageKind) {
		return fmt.Errorf("unknown contract inbox message kind %q", m.MessageKind)
	}
	if err := validateRuntimeToken("contract inbox source", m.Source, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract inbox payload hash", m.PayloadHash); err != nil {
		return err
	}
	if m.GasLimit == 0 {
		return errors.New("contract inbox gas limit must be positive")
	}
	if m.Sequence == 0 || m.ReceivedHeight == 0 {
		return errors.New("contract inbox sequence and height must be positive")
	}
	return nil
}

func (r ContractExecutionReceipt) ValidateFormat() error {
	if _, err := ContractReceiptKey(r.ContractAddr, r.ReceiptID); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract receipt message id", r.MsgID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if !IsContractMessageKind(r.MessageKind) {
		return fmt.Errorf("unknown contract receipt message kind %q", r.MessageKind)
	}
	if !IsZoneReceiptStatus(r.Status) {
		return fmt.Errorf("unknown contract receipt status %q", r.Status)
	}
	if err := ValidateHash("contract receipt output hash", r.OutputHash); err != nil {
		return err
	}
	if err := ValidateHash("contract receipt storage root", r.StorageRoot); err != nil {
		return err
	}
	if r.Height == 0 || r.Sequence == 0 {
		return errors.New("contract receipt height and sequence must be positive")
	}
	if r.ReceiptHash != "" {
		return ValidateHash("contract receipt hash", r.ReceiptHash)
	}
	return nil
}

func (r ContractExecutionReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash == "" {
		return errors.New("contract receipt hash is required")
	}
	if r.ReceiptHash != ComputeContractExecutionReceiptHash(r) {
		return errors.New("contract receipt hash mismatch")
	}
	return nil
}

func (r ContractExecutionReceipt) ZoneReceipt() (ZoneReceipt, error) {
	return NewZoneReceipt(ZoneReceipt{
		ZoneID:		ZoneIDContract,
		Height:		r.Height,
		ItemHash:	hashRuntimeParts("contract-zone-receipt-item-v1", r.ContractAddr, r.ReceiptID, r.MsgID),
		Status:		r.Status,
		GasUsed:	r.GasUsed,
		ResultHash:	r.OutputHash,
		Sequence:	r.Sequence,
	})
}

func (c ContractAsyncCall) ValidateFormat() error {
	if c.CallID != "" {
		if _, err := ContractInboxKey(c.TargetContract, c.CallID); err != nil {
			return err
		}
	}
	if err := validateRuntimeToken("contract async call source", c.Source, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateZoneID(c.SourceZone); err != nil {
		return err
	}
	if _, err := ContractInstanceKey(c.TargetContract); err != nil {
		return err
	}
	if err := ValidateHash("contract async call payload hash", c.PayloadHash); err != nil {
		return err
	}
	if c.GasLimit == 0 || c.RetryNonce == 0 || c.ExpiryHeight == 0 || c.CreatedHeight == 0 {
		return errors.New("contract async call gas, nonce, expiry, and height must be positive")
	}
	if c.ExpiryHeight <= c.CreatedHeight {
		return errors.New("contract async call expiry must be after creation")
	}
	if c.CallHash != "" {
		return ValidateHash("contract async call hash", c.CallHash)
	}
	return nil
}

func (c ContractAsyncCall) ValidateHash() error {
	if err := c.ValidateFormat(); err != nil {
		return err
	}
	if c.CallID != ComputeContractAsyncCallID(c) {
		return errors.New("contract async call id mismatch")
	}
	if c.CallHash != ComputeContractAsyncCallHash(c) {
		return errors.New("contract async call hash mismatch")
	}
	return nil
}

func (e ContractEvent) ValidateFormat() error {
	if _, err := ContractEventKey(e.ContractAddr, e.EventID); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract event topic", e.Topic, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("contract event payload hash", e.PayloadHash); err != nil {
		return err
	}
	if e.Height == 0 || e.Sequence == 0 {
		return errors.New("contract event height and sequence must be positive")
	}
	if e.EventHash != "" {
		return ValidateHash("contract event hash", e.EventHash)
	}
	return nil
}

func (e ContractEvent) ValidateHash() error {
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EventHash != ComputeContractEventHash(e) {
		return errors.New("contract event hash mismatch")
	}
	return nil
}

func (r ContractShardRoute) ValidateHash() error {
	if r.ZoneID != ZoneIDContract {
		return errors.New("contract shard route must use CONTRACT_ZONE")
	}
	if r.LayoutEpoch == 0 || r.ShardCount == 0 {
		return errors.New("contract shard route requires layout epoch and shard count")
	}
	if r.ShardID >= r.ShardCount {
		return errors.New("contract shard route shard id out of range")
	}
	if !IsContractShardRoutingMode(r.RoutingMode) {
		return fmt.Errorf("unknown contract shard routing mode %q", r.RoutingMode)
	}
	if err := validateRuntimeToken("contract shard route key", r.RouteKey, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("contract shard state key", r.StateKey, MaxZoneNamespaceLength); err != nil {
		return err
	}
	if err := ValidateHash("contract shard route hash", r.RouteHash); err != nil {
		return err
	}
	if r.RouteHash != ComputeContractShardRouteHash(r) {
		return errors.New("contract shard route hash mismatch")
	}
	return nil
}

func (p ContractProofRootExport) Validate() error {
	if p.ZoneID != ZoneIDContract {
		return errors.New("contract proof root export must use CONTRACT_ZONE")
	}
	if p.Height == 0 {
		return errors.New("contract proof root export height must be positive")
	}
	if !IsContractProofRootType(p.RootType) {
		return fmt.Errorf("unknown contract proof root type %q", p.RootType)
	}
	if err := ValidateHash("contract proof root export hash", p.RootHash); err != nil {
		return err
	}
	return validateRuntimeToken("contract proof root export source", p.Source, MaxZoneNamespaceLength)
}

func (r ContractZoneRoots) Validate() error {
	if r.Height == 0 {
		return errors.New("contract zone root height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "contract code root", value: r.CodeRoot},
		{name: "contract instance root", value: r.InstanceRoot},
		{name: "contract storage root", value: r.StorageRoot},
		{name: "contract ABI root", value: r.ABIRoot},
		{name: "contract inbox root", value: r.InboxRoot},
		{name: "contract outbox root", value: r.OutboxRoot},
		{name: "contract receipt root", value: r.ReceiptRoot},
		{name: "contract gas table root", value: r.GasTableRoot},
		{name: "contract event root", value: r.EventRoot},
		{name: "contract bytecode interface root", value: r.BytecodeInterfaceRoot},
		{name: "contract CosmWasm adapter root", value: r.CosmWasmAdapterRoot},
		{name: "contract execution root", value: r.ExecutionRoot},
		{name: "contract proof root", value: r.ProofRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.StateRoot != "" {
		return ValidateHash("contract state root", r.StateRoot)
	}
	return nil
}

func ComputeContractBytecodeInterfaceHash(iface ContractBytecodeInterface) string {
	return hashRuntimeParts(
		"aetra-contract-bytecode-interface-v1",
		string(iface.Runtime),
		iface.InstructionSet,
		iface.BytecodeHash,
		iface.ABIHash,
		iface.DeterminismHash,
		fmt.Sprint(iface.MaxCodeBytes),
	)
}

func ComputeAVMGasTableHash(table AVMGasTable) string {
	ordered := normalizeAVMInstructionGasCosts(table.Costs)
	parts := []string{"aetra-avm-gas-table-v1", fmt.Sprint(table.Version), fmt.Sprint(len(ordered))}
	for _, cost := range ordered {
		parts = append(parts, cost.Instruction, fmt.Sprint(cost.BaseGas))
	}
	return hashRuntimeParts(parts...)
}

func ComputeAVMGasTableRoot(table AVMGasTable) string {
	if table.Version == 0 {
		return EmptyRootHash()
	}
	return hashRuntimeParts("aetra-contract-gas-table-root-v1", fmt.Sprint(table.Version), table.TableHash)
}

func ComputeAVMExecutionKeeperHash(keeper AVMExecutionKeeper) string {
	return hashRuntimeParts("aetra-avm-execution-keeper-v1", keeper.RuntimeID, keeper.GasTable.TableHash, fmt.Sprint(keeper.ZoneGasLimit), fmt.Sprint(keeper.MaxInstructions))
}

func ComputeContractShardRouteHash(route ContractShardRoute) string {
	return hashRuntimeParts(
		"aetra-contract-shard-route-v1",
		string(route.ZoneID),
		fmt.Sprint(route.LayoutEpoch),
		fmt.Sprint(route.ShardCount),
		fmt.Sprint(route.ShardID),
		string(route.RoutingMode),
		route.RouteKey,
		route.StateKey,
	)
}

func ComputeContractAsyncCallID(call ContractAsyncCall) string {
	return hashRuntimeParts("aetra-contract-async-call-id-v1", call.Source, string(call.SourceZone), call.TargetContract, fmt.Sprint(call.CreatedHeight), fmt.Sprint(call.RetryNonce))
}

func ComputeContractAsyncCallHash(call ContractAsyncCall) string {
	callID := call.CallID
	if callID == "" {
		callID = ComputeContractAsyncCallID(call)
	}
	return hashRuntimeParts("aetra-contract-async-call-v1", callID, call.Source, string(call.SourceZone), call.TargetContract, call.PayloadHash, fmt.Sprint(call.GasLimit), fmt.Sprint(call.RetryNonce), fmt.Sprint(call.ExpiryHeight), fmt.Sprint(call.CreatedHeight))
}

func ComputeContractEventHash(event ContractEvent) string {
	return hashRuntimeParts("aetra-contract-event-v1", event.ContractAddr, event.EventID, event.Topic, event.PayloadHash, fmt.Sprint(event.Height), fmt.Sprint(event.Sequence))
}

func ComputeContractCosmWasmAdapterHash(adapter ContractCosmWasmAdapterDescriptor) string {
	return hashRuntimeParts(
		"aetra-contract-cosmwasm-adapter-v1",
		adapter.AdapterID,
		adapter.Version,
		adapter.PolicyHash,
		adapter.CapabilityRoot,
	)
}

func ComputeContractExecutionReceiptHash(receipt ContractExecutionReceipt) string {
	return hashRuntimeParts(
		"aetra-contract-zone-receipt-v1",
		receipt.ContractAddr,
		receipt.ReceiptID,
		receipt.MsgID,
		string(receipt.MessageKind),
		string(receipt.Status),
		fmt.Sprint(receipt.GasUsed),
		receipt.OutputHash,
		receipt.StorageRoot,
		fmt.Sprint(receipt.EmittedMessages),
		fmt.Sprint(receipt.Height),
		fmt.Sprint(receipt.Sequence),
	)
}

func ComputeContractZoneStateRoot(state ContractZoneState) string {
	normalized := state.Normalize()
	return hashRuntimeParts(
		"aetra-contract-zone-state-v1",
		ComputeContractCodeRoot(normalized.Codes),
		ComputeContractInstanceRoot(normalized.Instances),
		ComputeContractStorageRoot(normalized.Storage),
		ComputeContractABIRoot(normalized.ABIs),
		ComputeContractInboxRoot(normalized.Inbox),
		ComputeContractOutboxRoot(normalized.Outbox),
		ComputeContractReceiptRoot(normalized.Receipts),
		ComputeAVMGasTableRoot(normalized.GasTable),
		ComputeContractEventRoot(normalized.Events),
		ComputeContractBytecodeInterfaceRoot(normalized.BytecodeInterfaces),
		ComputeContractCosmWasmAdapterRoot(normalized.CosmWasmAdapters),
	)
}

func ComputeContractCodeRoot(codes []ContractCode) string {
	ordered := normalizeContractCodes(codes)
	parts := []string{"aetra-contract-code-root-v1", fmt.Sprint(len(ordered))}
	for _, code := range ordered {
		parts = append(parts, code.CodeID, string(code.Runtime), code.BytecodeHash, fmt.Sprint(code.BytecodeSize), code.ABIHash, code.InterfaceHash, code.Uploader, fmt.Sprint(code.UploadedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractInstanceRoot(instances []ContractInstance) string {
	ordered := normalizeContractInstances(instances)
	parts := []string{"aetra-contract-instance-root-v1", fmt.Sprint(len(ordered))}
	for _, instance := range ordered {
		parts = append(parts, instance.ContractAddr, instance.CodeID, string(instance.Runtime), instance.Admin, instance.StorageRoot, fmt.Sprint(instance.CreatedHeight), fmt.Sprint(instance.UpdatedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractStorageRoot(entries []ContractStorageEntry) string {
	ordered := normalizeContractStorage(entries)
	parts := []string{"aetra-contract-storage-root-v1", fmt.Sprint(len(ordered))}
	for _, entry := range ordered {
		key, _ := ContractStorageKey(entry.ContractAddr, entry.Key)
		parts = append(parts, key, entry.ValueHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractABIRoot(abis []ContractABIDescriptor) string {
	ordered := normalizeContractABIs(abis)
	parts := []string{"aetra-contract-abi-root-v1", fmt.Sprint(len(ordered))}
	for _, abi := range ordered {
		methods := append([]string(nil), abi.ExportedMethods...)
		sort.Strings(methods)
		parts = append(parts, abi.CodeID, abi.ABIHash, abi.InterfaceHash, fmt.Sprint(abi.RegisteredHeight))
		parts = append(parts, methods...)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractInboxRoot(inbox []ContractInboxMessage) string {
	ordered := normalizeContractInbox(inbox)
	parts := []string{"aetra-contract-inbox-root-v1", fmt.Sprint(len(ordered))}
	for _, msg := range ordered {
		key, _ := ContractInboxKey(msg.ContractAddr, msg.MsgID)
		parts = append(parts, key, string(msg.MessageKind), msg.Source, msg.PayloadHash, fmt.Sprint(msg.GasLimit), fmt.Sprint(msg.Sequence), fmt.Sprint(msg.ReceivedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractOutboxRoot(outbox []ContractInboxMessage) string {
	ordered := normalizeContractInbox(outbox)
	parts := []string{"aetra-contract-outbox-root-v1", fmt.Sprint(len(ordered))}
	for _, msg := range ordered {
		key, _ := ContractOutboxKey(msg.ContractAddr, msg.MsgID)
		parts = append(parts, key, string(msg.MessageKind), msg.Source, msg.PayloadHash, fmt.Sprint(msg.GasLimit), fmt.Sprint(msg.Sequence), fmt.Sprint(msg.ReceivedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractReceiptRoot(receipts []ContractExecutionReceipt) string {
	ordered := normalizeContractReceipts(receipts)
	parts := []string{"aetra-contract-receipt-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ReceiptHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractExecutionRoot(receipts []ContractExecutionReceipt) string {
	ordered := normalizeContractReceipts(receipts)
	parts := []string{"aetra-contract-execution-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ContractAddr, receipt.MsgID, string(receipt.Status), receipt.OutputHash, receipt.StorageRoot, fmt.Sprint(receipt.GasUsed))
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractEventRoot(events []ContractEvent) string {
	ordered := normalizeContractEvents(events)
	parts := []string{"aetra-contract-event-root-v1", fmt.Sprint(len(ordered))}
	for _, event := range ordered {
		parts = append(parts, event.EventHash)
	}
	return hashRuntimeParts(parts...)
}

func BuildContractProofRootExports(height uint64, roots ContractZoneRoots) ([]ContractProofRootExport, error) {
	if roots.Height != height {
		return nil, errors.New("contract proof root export height mismatch")
	}
	if err := roots.Validate(); err != nil {
		return nil, err
	}
	exports := []ContractProofRootExport{
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootCode, RootHash: roots.CodeRoot, Source: "contract.zone.code"},
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootInstance, RootHash: roots.InstanceRoot, Source: "contract.zone.instance"},
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootStorage, RootHash: roots.StorageRoot, Source: "contract.zone.storage"},
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootABI, RootHash: roots.ABIRoot, Source: "contract.zone.abi"},
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootGasTable, RootHash: roots.GasTableRoot, Source: "contract.zone.gas"},
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootEvent, RootHash: roots.EventRoot, Source: "contract.zone.event"},
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootInbox, RootHash: roots.InboxRoot, Source: "contract.zone.inbox"},
		{ZoneID: ZoneIDContract, Height: height, RootType: ContractProofRootOutbox, RootHash: roots.OutboxRoot, Source: "contract.zone.outbox"},
	}
	sort.SliceStable(exports, func(i, j int) bool { return exports[i].RootType < exports[j].RootType })
	for _, export := range exports {
		if err := export.Validate(); err != nil {
			return nil, err
		}
	}
	return exports, nil
}

func ComputeContractBytecodeInterfaceRoot(interfaces []ContractBytecodeInterface) string {
	ordered := normalizeContractBytecodeInterfaces(interfaces)
	parts := []string{"aetra-contract-bytecode-interface-root-v1", fmt.Sprint(len(ordered))}
	for _, iface := range ordered {
		parts = append(parts, string(iface.Runtime), iface.InstructionSet, iface.BytecodeHash, iface.ABIHash, iface.DeterminismHash, fmt.Sprint(iface.MaxCodeBytes), iface.InterfaceHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeContractCosmWasmAdapterRoot(adapters []ContractCosmWasmAdapterDescriptor) string {
	ordered := normalizeContractCosmWasmAdapters(adapters)
	parts := []string{"aetra-contract-cosmwasm-adapter-root-v1", fmt.Sprint(len(ordered))}
	for _, adapter := range ordered {
		parts = append(parts, adapter.AdapterID, adapter.Version, adapter.PolicyHash, adapter.CapabilityRoot, adapter.DescriptorHash)
	}
	return hashRuntimeParts(parts...)
}

func IsContractRuntimeKind(kind ContractRuntimeKind) bool {
	switch kind {
	case ContractRuntimeAVM, ContractRuntimeCosmWasm:
		return true
	default:
		return false
	}
}

func IsContractMessageKind(kind ContractMessageKind) bool {
	switch kind {
	case ContractMessageStoreCode,
		ContractMessageInstantiate,
		ContractMessageExecute,
		ContractMessageMigrate,
		ContractMessageCallback,
		ContractMessageProofVerify,
		ContractMessageAsyncCall:
		return true
	default:
		return false
	}
}

func IsContractProofKind(kind ContractProofKind) bool {
	switch kind {
	case ContractProofCode,
		ContractProofContract,
		ContractProofState,
		ContractProofABI,
		ContractProofReceipt,
		ContractProofGasTable,
		ContractProofEvent,
		ContractProofInbox,
		ContractProofOutbox:
		return true
	default:
		return false
	}
}

func IsContractShardRoutingMode(mode ContractShardRoutingMode) bool {
	switch mode {
	case ContractRouteAddress, ContractRouteStorageKey, ContractRouteCodeRegistry:
		return true
	default:
		return false
	}
}

func IsContractProofRootType(rootType ContractProofRootType) bool {
	switch rootType {
	case ContractProofRootCode, ContractProofRootInstance, ContractProofRootStorage, ContractProofRootABI, ContractProofRootGasTable, ContractProofRootEvent, ContractProofRootInbox, ContractProofRootOutbox:
		return true
	default:
		return false
	}
}

func normalizeAVMInstructionGasCosts(costs []AVMInstructionGasCost) []AVMInstructionGasCost {
	out := append([]AVMInstructionGasCost(nil), costs...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Instruction < out[j].Instruction })
	return out
}

func normalizeContractCodes(codes []ContractCode) []ContractCode {
	out := append([]ContractCode(nil), codes...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CodeID < out[j].CodeID })
	return out
}

func normalizeContractInstances(instances []ContractInstance) []ContractInstance {
	out := append([]ContractInstance(nil), instances...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ContractAddr < out[j].ContractAddr })
	return out
}

func normalizeContractStorage(entries []ContractStorageEntry) []ContractStorageEntry {
	out := append([]ContractStorageEntry(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ContractAddr != out[j].ContractAddr {
			return out[i].ContractAddr < out[j].ContractAddr
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func normalizeContractABIs(abis []ContractABIDescriptor) []ContractABIDescriptor {
	out := append([]ContractABIDescriptor(nil), abis...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CodeID < out[j].CodeID })
	return out
}

func normalizeContractInbox(inbox []ContractInboxMessage) []ContractInboxMessage {
	out := append([]ContractInboxMessage(nil), inbox...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ContractAddr != out[j].ContractAddr {
			return out[i].ContractAddr < out[j].ContractAddr
		}
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].MsgID < out[j].MsgID
	})
	return out
}

func normalizeContractReceipts(receipts []ContractExecutionReceipt) []ContractExecutionReceipt {
	out := append([]ContractExecutionReceipt(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ContractAddr != out[j].ContractAddr {
			return out[i].ContractAddr < out[j].ContractAddr
		}
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].ReceiptID < out[j].ReceiptID
	})
	return out
}

func normalizeContractBytecodeInterfaces(interfaces []ContractBytecodeInterface) []ContractBytecodeInterface {
	out := append([]ContractBytecodeInterface(nil), interfaces...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Runtime != out[j].Runtime {
			return out[i].Runtime < out[j].Runtime
		}
		return out[i].InstructionSet < out[j].InstructionSet
	})
	return out
}

func normalizeContractCosmWasmAdapters(adapters []ContractCosmWasmAdapterDescriptor) []ContractCosmWasmAdapterDescriptor {
	out := append([]ContractCosmWasmAdapterDescriptor(nil), adapters...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].AdapterID != out[j].AdapterID {
			return out[i].AdapterID < out[j].AdapterID
		}
		return out[i].Version < out[j].Version
	})
	return out
}

func normalizeContractEvents(events []ContractEvent) []ContractEvent {
	out := append([]ContractEvent(nil), events...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ContractAddr != out[j].ContractAddr {
			return out[i].ContractAddr < out[j].ContractAddr
		}
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].EventID < out[j].EventID
	})
	return out
}

func routeContractStateKey(mode ContractShardRoutingMode, routeKey string, stateKey string, shardCount uint32, layoutEpoch uint64) (ContractShardRoute, error) {
	if shardCount == 0 {
		return ContractShardRoute{}, errors.New("contract shard count must be positive")
	}
	if layoutEpoch == 0 {
		return ContractShardRoute{}, errors.New("contract shard layout epoch must be positive")
	}
	if err := validateRuntimeToken("contract shard route key", routeKey, MaxZoneEndpointLength); err != nil {
		return ContractShardRoute{}, err
	}
	if err := validateRuntimeToken("contract shard state key", stateKey, MaxZoneNamespaceLength); err != nil {
		return ContractShardRoute{}, err
	}
	hash := hashRuntimeParts("aetra-contract-route-key-v1", string(mode), routeKey, fmt.Sprint(layoutEpoch))
	bytes, err := hex.DecodeString(hash[:16])
	if err != nil {
		return ContractShardRoute{}, err
	}
	route := ContractShardRoute{
		ZoneID:		ZoneIDContract,
		LayoutEpoch:	layoutEpoch,
		ShardCount:	shardCount,
		ShardID:	uint32(binary.BigEndian.Uint64(bytes) % uint64(shardCount)),
		RoutingMode:	mode,
		RouteKey:	routeKey,
		StateKey:	stateKey,
	}
	route.RouteHash = ComputeContractShardRouteHash(route)
	return route, route.ValidateHash()
}
