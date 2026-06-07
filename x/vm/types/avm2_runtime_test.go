package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVM2BytecodeCodecRejectsMalformedOpcodesBeforeExecution(t *testing.T) {
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	module, err := NewAVM2BytecodeModule(AVM2BytecodeModule{
		Instructions: []AVM2Instruction{{Opcode: AVM2OpPush, Value: []byte("ok")}},
	}, limits, gasTable)
	require.NoError(t, err)
	require.NoError(t, module.Validate(limits, gasTable))
	require.Equal(t, ComputeAVM2BytecodeHash(module), module.BytecodeHash)
	require.True(t, len(module.CanonicalBytes) > 0)

	malformed := module
	malformed.Instructions = []AVM2Instruction{{Opcode: AVM2OpExternalNetwork}}
	malformed.CanonicalBytes = EncodeAVM2Bytecode(malformed)
	malformed.BytecodeHash = ComputeAVM2BytecodeHash(malformed)
	require.ErrorContains(t, malformed.Validate(limits, gasTable), "forbidden")

	tampered := module
	tampered.CanonicalBytes = []byte("bad-codec")
	_, err = DecodeAVM2Bytecode(tampered, limits, gasTable)
	require.ErrorContains(t, err, "codec mismatch")
}

func TestAVM2StoreV2AdapterAppliesPrefixLocalWrites(t *testing.T) {
	adapter := mustAVM2StoreAdapter(t, "contract-a", []AVM2StoreV2Entry{
		{Key: AVMContractStorageKey("contract-a", "existing"), ValueHash: ComputeAVM2BytesHash([]byte("1")), ValueBytes: 1},
	})
	updated, err := ApplyAVM2StoreV2Writes(adapter, []AVM2StorageWrite{
		{Key: AVMContractStorageKey("contract-a", "new"), ValueHash: ComputeAVM2BytesHash([]byte("2"))},
	})
	require.NoError(t, err)
	require.Len(t, updated.Entries, 2)
	require.NotEqual(t, adapter.AdapterRoot, updated.AdapterRoot)

	_, err = ApplyAVM2StoreV2Writes(adapter, []AVM2StorageWrite{
		{Key: AVMContractStorageKey("contract-b", "new"), ValueHash: ComputeAVM2BytesHash([]byte("2"))},
	})
	require.ErrorContains(t, err, "contract-local prefix")
}

func TestAVM2MessageDrivenTransitionIsDeterministic(t *testing.T) {
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	instruction := AVM2Instruction{Opcode: AVM2OpKVSet, Key: AVMContractStorageKey("contract-a", "balance"), Value: []byte("100")}
	inputA := mustAVM2RuntimeInput(t, []AVM2Instruction{instruction}, []AVM2StoreV2Entry{
		{Key: AVMContractStorageKey("contract-a", "z"), ValueHash: ComputeAVM2BytesHash([]byte("z")), ValueBytes: 1},
		{Key: AVMContractStorageKey("contract-a", "a"), ValueHash: ComputeAVM2BytesHash([]byte("a")), ValueBytes: 1},
	})
	inputB := mustAVM2RuntimeInput(t, []AVM2Instruction{instruction}, []AVM2StoreV2Entry{
		{Key: AVMContractStorageKey("contract-a", "a"), ValueHash: ComputeAVM2BytesHash([]byte("a")), ValueBytes: 1},
		{Key: AVMContractStorageKey("contract-a", "z"), ValueHash: ComputeAVM2BytesHash([]byte("z")), ValueBytes: 1},
	})
	require.Equal(t, inputA.CurrentState.AdapterRoot, inputB.CurrentState.AdapterRoot)

	transitionA, err := ExecuteAVM2MessageTransition(inputA, limits, gasTable)
	require.NoError(t, err)
	transitionB, err := ExecuteAVM2MessageTransition(inputB, limits, gasTable)
	require.NoError(t, err)
	require.Equal(t, transitionA.StateTransitionHash, transitionB.StateTransitionHash)
	require.Equal(t, transitionA.ReceiptRoot, transitionB.ReceiptRoot)
	require.Equal(t, AVMReceiptStatusExecuted, transitionA.Receipt.Status)
	require.Equal(t, transitionA.UpdatedState.AdapterRoot, transitionA.StorageRoot)
}

func TestAVM2MessageDrivenTransitionCommitsMessagesProofsPromisesABIAndReceipts(t *testing.T) {
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	msg := mustAVM2Message(t)
	proof := mustAVM2Proof(t, limits)
	promise := mustAVM2Promise(t, msg.ID)
	abi := mustAVM2ABI(t)
	event := mustAVM2Event(t)
	input := mustAVM2RuntimeInput(t, []AVM2Instruction{
		{Opcode: AVM2OpVerifyMessageProof, Proof: proof},
		{Opcode: AVM2OpPromiseNew, Promise: promise},
		{Opcode: AVM2OpABIExport, ABI: abi},
		{Opcode: AVM2OpEmitEvent, Event: event},
		{Opcode: AVM2OpMsgSend, Message: msg},
	}, nil)
	transition, err := ExecuteAVM2MessageTransition(input, limits, gasTable)
	require.NoError(t, err)
	require.NoError(t, transition.Validate())
	require.Len(t, transition.Execution.ProofsVerified, 1)
	require.Len(t, transition.Execution.Promises, 1)
	require.Len(t, transition.Execution.ABIDescriptors, 1)
	require.Len(t, transition.Execution.Events, 1)
	require.Len(t, transition.Execution.OutputMessages, 1)
	require.Equal(t, transition.Execution.MessageRoot, transition.OutboxRoot)
	require.Equal(t, ComputeAVM2ReceiptRoot([]AVMExecutionReceipt{transition.Receipt}), transition.ReceiptRoot)
}

func TestAVM2FailedExecutionConsumesGasAndReceipt(t *testing.T) {
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	input := mustAVM2RuntimeInput(t, []AVM2Instruction{{Opcode: AVM2OpAbort}}, nil)
	transition, err := ExecuteAVM2MessageTransition(input, limits, gasTable)
	require.ErrorContains(t, err, "aborted")
	require.Equal(t, AVMReceiptStatusFailed, transition.Receipt.Status)
	require.Equal(t, AVM2RuntimeErrorAbort, transition.Receipt.ErrorCodeOptional)
	require.Positive(t, transition.Receipt.GasUsed)
	require.Equal(t, input.CurrentState.AdapterRoot, transition.UpdatedState.AdapterRoot)
}

func TestAVM2ContractShardRoutingUsesCommittedLayout(t *testing.T) {
	layout := mustAVM2ContractLayout(t, 3)
	contract := mustAVM2ContractRecord(t, 1, "contract-a")
	storage := mustAVM2StorageValue(t, "contract-a", "balance/main")
	event := mustAVM2EventRecord(t, "contract-a")
	msg := mustAVM2Message(t)

	routesA, err := RouteAVM2ContractState(layout, contract, []AVM2ContractStorageValue{storage}, []AVM2ContractEventRecord{event}, []AVMAsyncMessage{msg})
	require.NoError(t, err)
	reordered := layout
	reordered.ActiveShards = []coretypes.ShardDescriptor{layout.ActiveShards[2], layout.ActiveShards[0], layout.ActiveShards[1]}
	reordered.LayoutHash = coretypes.ComputeShardLayoutHash(reordered)
	routesB, err := RouteAVM2ContractState(reordered, contract, []AVM2ContractStorageValue{storage}, []AVM2ContractEventRecord{event}, []AVMAsyncMessage{msg})
	require.NoError(t, err)
	require.Equal(t, routesA.RouteSetHash, routesB.RouteSetHash)
	require.Equal(t, layout.LayoutEpoch, routesA.LayoutEpoch)
}

func TestAVM2GasMeteringFuzzStyleBounds(t *testing.T) {
	limits := DefaultAVM2Limits()
	limits.MaxMemoryBytes = 64
	gasTable := mustAVM2GasTable(t)
	ctx := mustAVM2Context(t, false)
	for i, opcode := range AllAVM2SupportedOpcodes() {
		instruction := fuzzableAVM2Instruction(t, opcode, ctx, uint64(i+20))
		gas, err := AVM2InstructionGas(instruction, gasTable, limits)
		require.NoError(t, err, string(opcode))
		require.Positive(t, gas, string(opcode))
		require.LessOrEqual(t, gas, limits.MaxInstructionGas, string(opcode))
	}
	tooMuchMemory := AVM2Instruction{Opcode: AVM2OpMemGrow, MemoryGrow: 65}
	program := mustAVM2RuntimeInput(t, []AVM2Instruction{tooMuchMemory}, nil)
	_, err := ExecuteAVM2MessageTransition(program, limits, gasTable)
	require.ErrorContains(t, err, "memory limit")
}

func mustAVM2RuntimeInput(t *testing.T, instructions []AVM2Instruction, entries []AVM2StoreV2Entry) AVM2MessageDrivenInput {
	t.Helper()
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	bytecode, err := NewAVM2BytecodeModule(AVM2BytecodeModule{Instructions: instructions}, limits, gasTable)
	require.NoError(t, err)
	state := mustAVM2StoreAdapter(t, "contract-a", entries)
	msg := mustAVM2RuntimeMessage(t, "contract-a")
	ctx := AVM2ExecutionContext{
		ChainID:         "aetra-local",
		Height:          12,
		ZoneID:          zonestypes.ZoneID("CONTRACT_ZONE"),
		ShardID:         1,
		ContractAddress: "contract-a",
		Caller:          "caller-a",
		GasLimit:        20_000,
	}
	ctx.ContextHash = ComputeAVM2ContextHash(ctx)
	input, err := NewAVM2MessageDrivenInput(AVM2MessageDrivenInput{
		Message:      msg,
		CurrentState: state,
		Context:      ctx,
		Bytecode:     bytecode,
	})
	require.NoError(t, err)
	return input
}

func mustAVM2RuntimeMessage(t *testing.T, destination string) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:         "aetra-local",
		Source:          "caller-a",
		Destination:     destination,
		Payload:         []byte("execute"),
		GasLimit:        100,
		ExpiryHeight:    20,
		RetryPolicy:     DefaultAVMRetryPolicy(20),
		BounceFlag:      true,
		SourceZone:      zonestypes.ZoneID("APPLICATION_ZONE"),
		DestinationZone: zonestypes.ZoneID("CONTRACT_ZONE"),
		SenderNonce:     1,
		PayloadType:     "contract.execute",
		ValueNAET:       1,
		ForwardingFee:   1,
		Priority:        1,
		CreatedHeight:   10,
	})
	require.NoError(t, err)
	return msg
}

func mustAVM2StoreAdapter(t *testing.T, contract string, entries []AVM2StoreV2Entry) AVM2StoreV2Adapter {
	t.Helper()
	for i := range entries {
		entries[i].EntryHash = ComputeAVM2StoreV2EntryHash(entries[i])
	}
	adapter, err := NewAVM2StoreV2Adapter(AVM2StoreV2Adapter{
		ContractAddress: contract,
		Entries:         entries,
	})
	require.NoError(t, err)
	return adapter
}

func mustAVM2ContractLayout(t *testing.T, epoch uint64) coretypes.ShardLayout {
	t.Helper()
	shards := []coretypes.ShardDescriptor{
		{ShardID: "0", StatePrefix: "contract/0", ActivationHeight: 1, ValidatorSetHash: ComputeAVM2BytesHash([]byte("v0")), HashRangeStart: 0, HashRangeEnd: 100, Available: true},
		{ShardID: "1", StatePrefix: "contract/1", ActivationHeight: 1, ValidatorSetHash: ComputeAVM2BytesHash([]byte("v1")), HashRangeStart: 101, HashRangeEnd: 200, Available: true},
		{ShardID: "2", StatePrefix: "contract/2", ActivationHeight: 1, ValidatorSetHash: ComputeAVM2BytesHash([]byte("v2")), HashRangeStart: 201, HashRangeEnd: 300, Available: true},
	}
	layout, err := coretypes.NewShardLayout(coretypes.ZoneIDContract, epoch, 1, ComputeAVM2BytesHash([]byte("routing-seed")), shards)
	require.NoError(t, err)
	return layout
}

func fuzzableAVM2Instruction(t *testing.T, opcode AVM2Opcode, ctx AVM2ExecutionContext, nonce uint64) AVM2Instruction {
	t.Helper()
	switch opcode {
	case AVM2OpLoadLocal, AVM2OpStoreLocal:
		return AVM2Instruction{Opcode: opcode, Key: "local"}
	case AVM2OpJmp, AVM2OpJmpIf, AVM2OpCallInternal:
		return AVM2Instruction{Opcode: opcode, RangeLimit: 1}
	case AVM2OpMemGrow, AVM2OpMemLoad, AVM2OpMemStore, AVM2OpMemCopy:
		return AVM2Instruction{Opcode: opcode, MemoryGrow: 1}
	case AVM2OpKVGet, AVM2OpKVSet, AVM2OpKVDelete, AVM2OpKVExists, AVM2OpKVRangeBounded:
		return AVM2Instruction{Opcode: opcode, Key: AVMContractStorageKey(ctx.ContractAddress, fmt.Sprintf("k%d", nonce)), Value: []byte("v"), RangeLimit: 1}
	case AVM2OpVerifySig, AVM2OpVerifyMerkleProof, AVM2OpVerifyMessageProof, AVM2OpVerifyZoneRoot:
		return AVM2Instruction{Opcode: opcode, Proof: mustAVM2Proof(t, DefaultAVM2Limits())}
	case AVM2OpMsgNew, AVM2OpMsgSetValue, AVM2OpMsgSetPayload, AVM2OpMsgSetGas, AVM2OpMsgSetExpiry, AVM2OpMsgSend, AVM2OpMsgBounce:
		msg := mustAVM2Message(t)
		msg.SenderNonce = nonce
		msg.ID = DeriveAVMAsyncMessageID(msg)
		return AVM2Instruction{Opcode: opcode, Message: msg}
	case AVM2OpPromiseNew, AVM2OpPromiseAwait, AVM2OpPromiseResolve, AVM2OpPromiseReject, AVM2OpPromiseTimeout:
		msg := mustAVM2Message(t)
		msg.SenderNonce = nonce
		msg.ID = DeriveAVMAsyncMessageID(msg)
		return AVM2Instruction{Opcode: opcode, Promise: mustAVM2Promise(t, msg.ID)}
	case AVM2OpABIExport, AVM2OpABIMethod, AVM2OpABIEvent, AVM2OpABIRequire:
		return AVM2Instruction{Opcode: opcode, ABI: mustAVM2ABI(t)}
	case AVM2OpEmitEvent:
		return AVM2Instruction{Opcode: opcode, Event: mustAVM2Event(t)}
	default:
		return AVM2Instruction{Opcode: opcode, Value: []byte("v")}
	}
}
