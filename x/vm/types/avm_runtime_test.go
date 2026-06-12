package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMBytecodeCodecRejectsMalformedOpcodesBeforeExecution(t *testing.T) {
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	module, err := NewAVMBytecodeModule(AVMBytecodeModule{
		Instructions: []AVMInstruction{{Opcode: AVMOpPush, Value: []byte("ok")}},
	}, limits, gasTable)
	require.NoError(t, err)
	require.NoError(t, module.Validate(limits, gasTable))
	require.Equal(t, ComputeAVMBytecodeHash(module), module.BytecodeHash)
	require.True(t, len(module.CanonicalBytes) > 0)

	malformed := module
	malformed.Instructions = []AVMInstruction{{Opcode: AVMOpExternalNetwork}}
	malformed.CanonicalBytes = EncodeAVMBytecode(malformed)
	malformed.BytecodeHash = ComputeAVMBytecodeHash(malformed)
	require.ErrorContains(t, malformed.Validate(limits, gasTable), "forbidden")

	tampered := module
	tampered.CanonicalBytes = []byte("bad-codec")
	_, err = DecodeAVMBytecode(tampered, limits, gasTable)
	require.ErrorContains(t, err, "codec mismatch")
}

func TestAVMStoreV2AdapterAppliesPrefixLocalWrites(t *testing.T) {
	adapter := mustAVMStoreAdapter(t, "contract-a", []AVMStoreV2Entry{
		{Key: AVMContractStorageKey("contract-a", "existing"), ValueHash: ComputeAVMBytesHash([]byte("1")), ValueBytes: 1},
	})
	updated, err := ApplyAVMStoreV2Writes(adapter, []AVMStorageWrite{
		{Key: AVMContractStorageKey("contract-a", "new"), ValueHash: ComputeAVMBytesHash([]byte("2"))},
	})
	require.NoError(t, err)
	require.Len(t, updated.Entries, 2)
	require.NotEqual(t, adapter.AdapterRoot, updated.AdapterRoot)

	_, err = ApplyAVMStoreV2Writes(adapter, []AVMStorageWrite{
		{Key: AVMContractStorageKey("contract-b", "new"), ValueHash: ComputeAVMBytesHash([]byte("2"))},
	})
	require.ErrorContains(t, err, "contract-local prefix")
}

func TestAVMMessageDrivenTransitionIsDeterministic(t *testing.T) {
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	instruction := AVMInstruction{Opcode: AVMOpKVSet, Key: AVMContractStorageKey("contract-a", "balance"), Value: []byte("100")}
	inputA := mustAVMRuntimeInput(t, []AVMInstruction{instruction}, []AVMStoreV2Entry{
		{Key: AVMContractStorageKey("contract-a", "z"), ValueHash: ComputeAVMBytesHash([]byte("z")), ValueBytes: 1},
		{Key: AVMContractStorageKey("contract-a", "a"), ValueHash: ComputeAVMBytesHash([]byte("a")), ValueBytes: 1},
	})
	inputB := mustAVMRuntimeInput(t, []AVMInstruction{instruction}, []AVMStoreV2Entry{
		{Key: AVMContractStorageKey("contract-a", "a"), ValueHash: ComputeAVMBytesHash([]byte("a")), ValueBytes: 1},
		{Key: AVMContractStorageKey("contract-a", "z"), ValueHash: ComputeAVMBytesHash([]byte("z")), ValueBytes: 1},
	})
	require.Equal(t, inputA.CurrentState.AdapterRoot, inputB.CurrentState.AdapterRoot)

	transitionA, err := ExecuteAVMMessageTransition(inputA, limits, gasTable)
	require.NoError(t, err)
	transitionB, err := ExecuteAVMMessageTransition(inputB, limits, gasTable)
	require.NoError(t, err)
	require.Equal(t, transitionA.StateTransitionHash, transitionB.StateTransitionHash)
	require.Equal(t, transitionA.ReceiptRoot, transitionB.ReceiptRoot)
	require.Equal(t, AVMReceiptStatusExecuted, transitionA.Receipt.Status)
	require.Equal(t, transitionA.UpdatedState.AdapterRoot, transitionA.StorageRoot)
}

func TestAVMMessageDrivenTransitionCommitsMessagesProofsPromisesABIAndReceipts(t *testing.T) {
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	msg := mustAVMMessage(t)
	proof := mustAVMProof(t, limits)
	promise := mustAVMPromise(t, msg.ID)
	abi := mustAVMABI(t)
	event := mustAVMEvent(t)
	input := mustAVMRuntimeInput(t, []AVMInstruction{
		{Opcode: AVMOpVerifyMessageProof, Proof: proof},
		{Opcode: AVMOpPromiseNew, Promise: promise},
		{Opcode: AVMOpABIExport, ABI: abi},
		{Opcode: AVMOpEmitEvent, Event: event},
		{Opcode: AVMOpMsgSend, Message: msg},
	}, nil)
	transition, err := ExecuteAVMMessageTransition(input, limits, gasTable)
	require.NoError(t, err)
	require.NoError(t, transition.Validate())
	require.Len(t, transition.Execution.ProofsVerified, 1)
	require.Len(t, transition.Execution.Promises, 1)
	require.Len(t, transition.Execution.ABIDescriptors, 1)
	require.Len(t, transition.Execution.Events, 1)
	require.Len(t, transition.Execution.OutputMessages, 1)
	require.Equal(t, transition.Execution.MessageRoot, transition.OutboxRoot)
	require.Equal(t, ComputeAVMReceiptRoot([]AVMExecutionReceipt{transition.Receipt}), transition.ReceiptRoot)
}

func TestAVMFailedExecutionConsumesGasAndReceipt(t *testing.T) {
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	input := mustAVMRuntimeInput(t, []AVMInstruction{{Opcode: AVMOpAbort}}, nil)
	transition, err := ExecuteAVMMessageTransition(input, limits, gasTable)
	require.ErrorContains(t, err, "aborted")
	require.Equal(t, AVMReceiptStatusFailed, transition.Receipt.Status)
	require.Equal(t, AVMRuntimeErrorAbort, transition.Receipt.ErrorCodeOptional)
	require.Positive(t, transition.Receipt.GasUsed)
	require.Equal(t, input.CurrentState.AdapterRoot, transition.UpdatedState.AdapterRoot)
}

func TestAVMContractShardRoutingUsesCommittedLayout(t *testing.T) {
	layout := mustAVMContractLayout(t, 3)
	contract := mustAVMContractRecord(t, 1, "contract-a")
	storage := mustAVMStorageValue(t, "contract-a", "balance/main")
	event := mustAVMEventRecord(t, "contract-a")
	msg := mustAVMMessage(t)

	routesA, err := RouteAVMContractState(layout, contract, []AVMContractStorageValue{storage}, []AVMContractEventRecord{event}, []AVMAsyncMessage{msg})
	require.NoError(t, err)
	reordered := layout
	reordered.ActiveShards = []coretypes.ShardDescriptor{layout.ActiveShards[2], layout.ActiveShards[0], layout.ActiveShards[1]}
	reordered.LayoutHash = coretypes.ComputeShardLayoutHash(reordered)
	routesB, err := RouteAVMContractState(reordered, contract, []AVMContractStorageValue{storage}, []AVMContractEventRecord{event}, []AVMAsyncMessage{msg})
	require.NoError(t, err)
	require.Equal(t, routesA.RouteSetHash, routesB.RouteSetHash)
	require.Equal(t, layout.LayoutEpoch, routesA.LayoutEpoch)
}

func TestAVMGasMeteringFuzzStyleBounds(t *testing.T) {
	limits := DefaultAVMLimits()
	limits.MaxMemoryBytes = 64
	gasTable := mustAVMGasTable(t)
	ctx := mustAVMContext(t, false)
	for i, opcode := range AllAVMSupportedOpcodes() {
		instruction := fuzzableAVMInstruction(t, opcode, ctx, uint64(i+20))
		gas, err := AVMInstructionGas(instruction, gasTable, limits)
		require.NoError(t, err, string(opcode))
		require.Positive(t, gas, string(opcode))
		require.LessOrEqual(t, gas, limits.MaxInstructionGas, string(opcode))
	}
	tooMuchMemory := AVMInstruction{Opcode: AVMOpMemGrow, MemoryGrow: 65}
	program := mustAVMRuntimeInput(t, []AVMInstruction{tooMuchMemory}, nil)
	_, err := ExecuteAVMMessageTransition(program, limits, gasTable)
	require.ErrorContains(t, err, "memory limit")
}

func mustAVMRuntimeInput(t *testing.T, instructions []AVMInstruction, entries []AVMStoreV2Entry) AVMMessageDrivenInput {
	t.Helper()
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	bytecode, err := NewAVMBytecodeModule(AVMBytecodeModule{Instructions: instructions}, limits, gasTable)
	require.NoError(t, err)
	state := mustAVMStoreAdapter(t, "contract-a", entries)
	msg := mustAVMRuntimeMessage(t, "contract-a")
	ctx := AVMExecutionContext{
		ChainID:		"aetra-local",
		Height:			12,
		ZoneID:			zonestypes.ZoneID("CONTRACT_ZONE"),
		ShardID:		1,
		ContractAddress:	"contract-a",
		Caller:			"caller-a",
		GasLimit:		20_000,
	}
	ctx.ContextHash = ComputeAVMContextHash(ctx)
	input, err := NewAVMMessageDrivenInput(AVMMessageDrivenInput{
		Message:	msg,
		CurrentState:	state,
		Context:	ctx,
		Bytecode:	bytecode,
	})
	require.NoError(t, err)
	return input
}

func mustAVMRuntimeMessage(t *testing.T, destination string) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:		"aetra-local",
		Source:			"caller-a",
		Destination:		destination,
		Payload:		[]byte("execute"),
		GasLimit:		100,
		ExpiryHeight:		20,
		RetryPolicy:		DefaultAVMRetryPolicy(20),
		BounceFlag:		true,
		SourceZone:		zonestypes.ZoneID("APPLICATION_ZONE"),
		DestinationZone:	zonestypes.ZoneID("CONTRACT_ZONE"),
		SenderNonce:		1,
		PayloadType:		"contract.execute",
		ValueNAET:		1,
		ForwardingFee:		1,
		Priority:		1,
		CreatedHeight:		10,
	})
	require.NoError(t, err)
	return msg
}

func mustAVMStoreAdapter(t *testing.T, contract string, entries []AVMStoreV2Entry) AVMStoreV2Adapter {
	t.Helper()
	for i := range entries {
		entries[i].EntryHash = ComputeAVMStoreV2EntryHash(entries[i])
	}
	adapter, err := NewAVMStoreV2Adapter(AVMStoreV2Adapter{
		ContractAddress:	contract,
		Entries:		entries,
	})
	require.NoError(t, err)
	return adapter
}

func mustAVMContractLayout(t *testing.T, epoch uint64) coretypes.ShardLayout {
	t.Helper()
	shards := []coretypes.ShardDescriptor{
		{ShardID: "0", StatePrefix: "contract/0", ActivationHeight: 1, ValidatorSetHash: ComputeAVMBytesHash([]byte("v0")), HashRangeStart: 0, HashRangeEnd: 100, Available: true},
		{ShardID: "1", StatePrefix: "contract/1", ActivationHeight: 1, ValidatorSetHash: ComputeAVMBytesHash([]byte("v1")), HashRangeStart: 101, HashRangeEnd: 200, Available: true},
		{ShardID: "2", StatePrefix: "contract/2", ActivationHeight: 1, ValidatorSetHash: ComputeAVMBytesHash([]byte("v2")), HashRangeStart: 201, HashRangeEnd: 300, Available: true},
	}
	layout, err := coretypes.NewShardLayout(coretypes.ZoneIDContract, epoch, 1, ComputeAVMBytesHash([]byte("routing-seed")), shards)
	require.NoError(t, err)
	return layout
}

func fuzzableAVMInstruction(t *testing.T, opcode AVMOpcode, ctx AVMExecutionContext, nonce uint64) AVMInstruction {
	t.Helper()
	switch opcode {
	case AVMOpLoadLocal, AVMOpStoreLocal:
		return AVMInstruction{Opcode: opcode, Key: "local"}
	case AVMOpJmp, AVMOpJmpIf, AVMOpCallInternal:
		return AVMInstruction{Opcode: opcode, RangeLimit: 1}
	case AVMOpMemGrow, AVMOpMemLoad, AVMOpMemStore, AVMOpMemCopy:
		return AVMInstruction{Opcode: opcode, MemoryGrow: 1}
	case AVMOpKVGet, AVMOpKVSet, AVMOpKVDelete, AVMOpKVExists, AVMOpKVRangeBounded:
		return AVMInstruction{Opcode: opcode, Key: AVMContractStorageKey(ctx.ContractAddress, fmt.Sprintf("k%d", nonce)), Value: []byte("v"), RangeLimit: 1}
	case AVMOpVerifySig, AVMOpVerifyMerkleProof, AVMOpVerifyMessageProof, AVMOpVerifyZoneRoot:
		return AVMInstruction{Opcode: opcode, Proof: mustAVMProof(t, DefaultAVMLimits())}
	case AVMOpMsgNew, AVMOpMsgSetValue, AVMOpMsgSetPayload, AVMOpMsgSetGas, AVMOpMsgSetExpiry, AVMOpMsgSend, AVMOpMsgBounce:
		msg := mustAVMMessage(t)
		msg.SenderNonce = nonce
		msg.ID = DeriveAVMAsyncMessageID(msg)
		return AVMInstruction{Opcode: opcode, Message: msg}
	case AVMOpPromiseNew, AVMOpPromiseAwait, AVMOpPromiseResolve, AVMOpPromiseReject, AVMOpPromiseTimeout:
		msg := mustAVMMessage(t)
		msg.SenderNonce = nonce
		msg.ID = DeriveAVMAsyncMessageID(msg)
		return AVMInstruction{Opcode: opcode, Promise: mustAVMPromise(t, msg.ID)}
	case AVMOpABIExport, AVMOpABIMethod, AVMOpABIEvent, AVMOpABIRequire:
		return AVMInstruction{Opcode: opcode, ABI: mustAVMABI(t)}
	case AVMOpEmitEvent:
		return AVMInstruction{Opcode: opcode, Event: mustAVMEvent(t)}
	default:
		return AVMInstruction{Opcode: opcode, Value: []byte("v")}
	}
}
