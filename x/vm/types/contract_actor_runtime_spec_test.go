package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMContractBackendInterfacesAndAdapters(t *testing.T) {
	nativeDescriptor := testAVMNativeModuleDescriptor(t)
	nativeInterface, err := NewAVMContractBackendInterface(AVMContractBackendInterface{
		Name:			"native-bank",
		BackendKind:		AVMContractBackendNativeModule,
		RouterBackend:		RouterBackendNativeModule,
		SupportsSync:		true,
		EmitsReceipt:		true,
		SupportsProofQuery:	true,
		InterfaceHash:		nativeDescriptor.ServiceInterfaceHash,
	})
	require.NoError(t, err)
	nativeAdapter, err := NewAVMNativeModuleAdapter(AVMNativeModuleAdapter{Interface: nativeInterface, Descriptor: nativeDescriptor})
	require.NoError(t, err)
	require.NoError(t, nativeAdapter.Validate())

	wasmInterface, err := NewAVMContractBackendInterface(AVMContractBackendInterface{
		Name:			"wasm-adapter",
		BackendKind:		AVMContractBackendWASMContract,
		RouterBackend:		RouterBackendWASMAdapter,
		SupportsAsync:		true,
		EmitsReceipt:		true,
		SupportsProofQuery:	true,
		InterfaceHash:		engineHash("wasm-interface"),
	})
	require.NoError(t, err)
	wasmBoundary, err := NewAVMWASMAdapterBoundary(AVMWASMAdapterBoundary{
		Interface:	wasmInterface,
		SandboxPolicy:	testAVMWASMSandboxPolicy(t),
	})
	require.NoError(t, err)
	require.NoError(t, wasmBoundary.Validate())

	noReceipt := nativeInterface
	noReceipt.EmitsReceipt = false
	noReceipt.BackendInterfaceHash = ComputeAVMContractBackendInterfaceHash(noReceipt)
	require.ErrorContains(t, noReceipt.Validate(), "emit receipts")
}

func TestAVMNativeActorContractExecutesOneMailboxMessageEmitsReceiptAndContinuation(t *testing.T) {
	actor := testAVMActorContractState(t)
	msg := testAVMActorContractMessage(t, actor.ActorID)
	receipt := testAVMActorContractReceipt(t, msg, AVMReceiptStatusExecuted)
	emitted := testAVMActorEmittedMessage(t, actor.ActorID)
	continuation := testAVMActorContinuation(actor.ActorID, msg.CreatedHeight)

	execution, err := NewAVMActorContractExecution(AVMActorContractExecution{
		Actor:			actor,
		Message:		msg,
		ActiveMessageCount:	1,
		StateReads: []AVMActorStateRead{{
			ActorID:	actor.ActorID,
			Key:		ActorStateKeyPrefix(actor.ActorID) + "balance",
			Hash:		engineHash("actor-read"),
		}},
		StateWrites: []ActorStateWrite{{
			ActorID:	actor.ActorID,
			Key:		ActorStateKeyPrefix(actor.ActorID) + "balance",
			Hash:		engineHash("actor-write"),
		}},
		EmittedMessages:	[]AVMAsyncMessage{emitted},
		StoredContinuations:	[]ContinuationRecord{continuation},
		Receipt:		receipt,
	})
	require.NoError(t, err)
	require.NoError(t, execution.Validate())
	require.Equal(t, ComputeAVMActorContractExecutionHash(execution), execution.ExecutionHash)
}

func TestAVMActorContractRejectsConcurrentMessagesForeignReadsAndMissingReceipt(t *testing.T) {
	actor := testAVMActorContractState(t)
	msg := testAVMActorContractMessage(t, actor.ActorID)
	receipt := testAVMActorContractReceipt(t, msg, AVMReceiptStatusExecuted)
	valid := AVMActorContractExecution{
		Actor:			actor,
		Message:		msg,
		ActiveMessageCount:	1,
		StateReads: []AVMActorStateRead{{
			ActorID:	actor.ActorID,
			Key:		ActorStateKeyPrefix(actor.ActorID) + "balance",
			Hash:		engineHash("actor-read"),
		}},
		Receipt:	receipt,
	}

	concurrent := valid
	concurrent.ActiveMessageCount = 2
	concurrent.ExecutionHash = ComputeAVMActorContractExecutionHash(concurrent)
	require.ErrorContains(t, concurrent.Validate(), "one message")

	foreignRead := valid
	foreignRead.StateReads = []AVMActorStateRead{{
		ActorID:	"other-actor",
		Key:		ActorStateKeyPrefix("other-actor") + "balance",
		Hash:		engineHash("foreign-read"),
	}}
	foreignRead.ExecutionHash = ComputeAVMActorContractExecutionHash(foreignRead)
	require.ErrorContains(t, foreignRead.Validate(), "another actor")

	noReceipt := valid
	noReceipt.Receipt = AVMExecutionReceipt{}
	noReceipt.ExecutionHash = ComputeAVMActorContractExecutionHash(noReceipt)
	require.ErrorContains(t, noReceipt.Validate(), "receipt")
}

func TestAVMContractRegistriesStoragePrefixesReceiptEmissionAndProofQuery(t *testing.T) {
	actor := testAVMActorContractState(t)
	msg := testAVMActorContractMessage(t, actor.ActorID)
	receipt := testAVMActorContractReceipt(t, msg, AVMReceiptStatusExecuted)
	codeRegistry, err := NewAVMContractCodeRegistry(AVMContractCodeRegistry{Codes: []AVMContractCodeRegistryRecord{{
		CodeID:		actor.CodeID,
		BackendKind:	AVMContractBackendActorContract,
		CodeHash:	engineHash("actor-code"),
		InterfaceHash:	actor.InterfaceHash,
		Owner:		actor.Owner,
		Enabled:	true,
	}}})
	require.NoError(t, err)
	instanceRegistry, err := NewAVMContractInstanceRegistry(AVMContractInstanceRegistry{Instances: []AVMContractInstanceRegistryRecord{{
		ContractAddress:	"contract-actor-1",
		CodeID:			actor.CodeID,
		BackendKind:		AVMContractBackendActorContract,
		ActorID:		actor.ActorID,
		Owner:			actor.Owner,
		StoragePrefix:		AVMStatePrefixContractStorage + "/contract-actor-1/",
		StateRoot:		actor.StateRoot,
		InterfaceHash:		actor.InterfaceHash,
		Status:			AVMActorContractActive,
	}}})
	require.NoError(t, err)
	storagePrefix, err := NewAVMContractStoragePrefixDescriptor(AVMContractStoragePrefixDescriptor{
		ContractAddress:	"contract-actor-1",
		StorageRoot:		engineHash("contract-storage"),
	})
	require.NoError(t, err)

	index := AVMContractProofIndex{
		CodeRegistry:		codeRegistry,
		InstanceRegistry:	instanceRegistry,
		StoragePrefixes:	[]AVMContractStoragePrefixDescriptor{storagePrefix},
		Receipts:		[]AVMExecutionReceipt{receipt},
	}
	codeProof, err := QueryAVMContractProof(index, AVMContractProofCode, AVMContractCodeKey(actor.CodeID))
	require.NoError(t, err)
	require.Equal(t, codeRegistry.RegistryRoot, codeProof.Root)
	instanceProof, err := QueryAVMContractProof(index, AVMContractProofInstance, "contract-actor-1")
	require.NoError(t, err)
	require.Equal(t, instanceRegistry.RegistryRoot, instanceProof.Root)
	storageProof, err := QueryAVMContractProof(index, AVMContractProofStorage, "contract-actor-1")
	require.NoError(t, err)
	require.Equal(t, storagePrefix.StorageRoot, storageProof.Root)
	receiptProof, err := QueryAVMContractProof(index, AVMContractProofReceipt, receipt.ReceiptID)
	require.NoError(t, err)
	require.Equal(t, receipt.ReceiptHash, receiptProof.RecordHash)
}

func testAVMActorContractState(t *testing.T) AVMActorContractState {
	t.Helper()
	state, err := NewAVMActorContractState(AVMActorContractState{
		ActorID:		"actor-contract-1",
		CodeID:			7,
		Owner:			"owner-1",
		MailboxRoot:		engineHash("actor-mailbox"),
		StateRoot:		engineHash("actor-state"),
		ContinuationRoot:	engineHash("actor-continuation"),
		InterfaceHash:		engineHash("actor-interface"),
		Status:			AVMActorContractActive,
	})
	require.NoError(t, err)
	return state
}

func testAVMActorContractMessage(t *testing.T, actorID string) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage("actor-source", zonestypes.ZoneIDApplication, "actor-contract", zonestypes.ZoneIDContract, 41, 18)
	msg.DestinationActorOptional = actorID
	msg.PayloadType = "actor.handle"
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}

func testAVMActorEmittedMessage(t *testing.T, actorID string) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage("actor-contract", zonestypes.ZoneIDContract, "workflow", zonestypes.ZoneIDApplication, 42, 19)
	msg.SourceActorOptional = actorID
	msg.PayloadType = "workflow.resume"
	msg.ValueNAET = 0
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}

func testAVMActorContinuation(actorID string, height uint64) ContinuationRecord {
	return ContinuationRecord{
		ContinuationID:		"continuation-actor-1",
		ActorID:		actorID,
		StepIndex:		1,
		PartialStateHash:	engineHash("continuation-partial"),
		PartialStateBytes:	32,
		ResumeHeight:		height + 1,
		ExpiryHeight:		height + 5,
		GasReserved:		100,
		Status:			ContinuationStatusScheduled,
		ResumeBy:		ContinuationResumeByScheduler,
	}
}

func testAVMActorContractReceipt(t *testing.T, msg AVMAsyncMessage, status AVMReceiptStatus) AVMExecutionReceipt {
	t.Helper()
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		msg.ID,
		ZoneID:			msg.DestinationZone,
		Executor:		"actor-contract-runtime",
		Status:			status,
		GasUsed:		30,
		StorageWritten:		1,
		EventsHash:		engineHash("actor-events"),
		OutputMessagesRoot:	engineHash("actor-output"),
		CreatedHeight:		msg.CreatedHeight + 1,
	})
	require.NoError(t, err)
	return receipt
}
