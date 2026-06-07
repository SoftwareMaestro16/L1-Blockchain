package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVM2ProgramRejectsForbiddenSurfaces(t *testing.T) {
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	for _, opcode := range []AVM2Opcode{
		AVM2OpExternalNetwork,
		AVM2OpWallClock,
		AVM2OpNonDeterministic,
		AVM2OpKVRangeUnbounded,
		AVM2OpDirectRemoteMutation,
		AVM2OpUnboundedRecursion,
	} {
		program := AVM2Program{
			VMVersion:             AVM2VMVersion,
			InstructionSetVersion: AVM2DefaultInstructionSet,
			Instructions:          []AVM2Instruction{{Opcode: opcode}},
			MaxRecursionDepth:     1,
		}
		program.ProgramHash = ComputeAVM2ProgramHash(program)
		require.Error(t, ValidateAVM2Program(program, limits, gasTable), string(opcode))
	}
}

func TestAVM2InstructionSetCoversSpecAndGasTable(t *testing.T) {
	set, err := DefaultAVM2InstructionSet()
	require.NoError(t, err)
	require.Len(t, set.Opcodes, len(AllAVM2SupportedOpcodes()))
	require.Equal(t, ComputeAVM2InstructionSetHash(set), set.SetHash)

	byCategory := map[AVM2InstructionCategory]int{}
	for _, descriptor := range set.Opcodes {
		byCategory[descriptor.Category]++
	}
	for _, category := range []AVM2InstructionCategory{
		AVM2CategoryCoreStack,
		AVM2CategoryArithmetic,
		AVM2CategoryControlFlow,
		AVM2CategoryMemory,
		AVM2CategoryStorage,
		AVM2CategoryCryptoProof,
		AVM2CategoryMessages,
		AVM2CategoryPromises,
		AVM2CategoryABI,
		AVM2CategoryContext,
	} {
		require.Positive(t, byCategory[category], string(category))
	}

	table := mustAVM2GasTable(t)
	for _, opcode := range AllAVM2SupportedOpcodes() {
		gas, found := table.OpcodeGas(opcode)
		require.True(t, found, string(opcode))
		require.Positive(t, gas)
	}

	missing := table
	missing.OpcodeCosts = missing.OpcodeCosts[1:]
	missing.TableHash = ComputeAVM2GasTableHash(missing)
	require.ErrorContains(t, missing.Validate(), "missing opcode")
}

func TestAVM2ExecutionEnforcesBoundsAndGas(t *testing.T) {
	limits := DefaultAVM2Limits()
	limits.MaxStackDepth = 1
	limits.MaxMemoryBytes = 8
	gasTable := mustAVM2GasTable(t)
	ctx := mustAVM2Context(t, false)

	stackOverflow, err := NewAVM2Program(AVM2Program{
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		Instructions: []AVM2Instruction{
			{Opcode: AVM2OpPush, Value: []byte("a")},
			{Opcode: AVM2OpPush, Value: []byte("b")},
		},
		MaxRecursionDepth: 1,
	}, limits, gasTable)
	require.NoError(t, err)
	_, err = ExecuteAVM2Program(stackOverflow, ctx, limits, gasTable)
	require.ErrorContains(t, err, "stack depth")

	memoryOverflow, err := NewAVM2Program(AVM2Program{
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		Instructions:          []AVM2Instruction{{Opcode: AVM2OpMemGrow, MemoryGrow: 9}},
		MaxRecursionDepth:     1,
	}, limits, gasTable)
	require.NoError(t, err)
	_, err = ExecuteAVM2Program(memoryOverflow, ctx, limits, gasTable)
	require.ErrorContains(t, err, "memory limit")

	lowGas := ctx
	lowGas.GasLimit = 1
	lowGas.ContextHash = ComputeAVM2ContextHash(lowGas)
	_, err = ExecuteAVM2Program(memoryOverflow, lowGas, limits, gasTable)
	require.ErrorContains(t, err, "exhausted gas")
}

func TestAVM2InstructionGasChargesResourceInputs(t *testing.T) {
	limits := DefaultAVM2Limits()
	table := mustAVM2GasTable(t)
	ctx := mustAVM2Context(t, false)

	smallWrite := AVM2Instruction{Opcode: AVM2OpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "a"), Value: []byte("x")}
	largeWrite := AVM2Instruction{Opcode: AVM2OpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "a"), Value: []byte("xxxxxxxx")}
	smallGas, err := AVM2InstructionGas(smallWrite, table, limits)
	require.NoError(t, err)
	largeGas, err := AVM2InstructionGas(largeWrite, table, limits)
	require.NoError(t, err)
	require.Greater(t, largeGas, smallGas)

	proof := mustAVM2Proof(t, limits)
	proofOp := AVM2Instruction{Opcode: AVM2OpVerifyMerkleProof, Proof: proof}
	proofGas, err := AVM2InstructionGas(proofOp, table, limits)
	require.NoError(t, err)
	proof.ProofBytes = append(proof.ProofBytes, []byte("more-proof")...)
	proof.ProofVersion++
	proof.ProofHash = ComputeAVM2ProofHash(proof)
	largerProofGas, err := AVM2InstructionGas(AVM2Instruction{Opcode: AVM2OpVerifyMerkleProof, Proof: proof}, table, limits)
	require.NoError(t, err)
	require.Greater(t, largerProofGas, proofGas)

	msg := mustAVM2Message(t)
	builderGas, err := AVM2InstructionGas(AVM2Instruction{Opcode: AVM2OpMsgNew, Message: msg}, table, limits)
	require.NoError(t, err)
	sendGas, err := AVM2InstructionGas(AVM2Instruction{Opcode: AVM2OpMsgSend, Message: msg}, table, limits)
	require.NoError(t, err)
	require.Greater(t, sendGas, builderGas)
}

func TestAVM2ContextAndStackOpcodesAreDeterministic(t *testing.T) {
	limits := DefaultAVM2Limits()
	table := mustAVM2GasTable(t)
	ctx := mustAVM2Context(t, false)

	program, err := NewAVM2Program(AVM2Program{
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		Instructions: []AVM2Instruction{
			{Opcode: AVM2OpCtxHeight},
			{Opcode: AVM2OpDup},
			{Opcode: AVM2OpSwap},
			{Opcode: AVM2OpPop},
			{Opcode: AVM2OpCtxChainID},
			{Opcode: AVM2OpCtxGasLeft},
		},
		MaxRecursionDepth: 1,
	}, limits, table)
	require.NoError(t, err)

	result, err := ExecuteAVM2Program(program, ctx, limits, table)
	require.NoError(t, err)
	require.Len(t, result.Stack, 3)
	require.Equal(t, "00000000000000000010", result.Stack[0])
	require.Equal(t, "aetra-local", result.Stack[1])
	require.NotEmpty(t, result.Stack[2])
}

func TestAVM2ScopeCommitsStoreMessagesProofsABIEvents(t *testing.T) {
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	ctx := mustAVM2Context(t, false)
	msg := mustAVM2Message(t)
	proof := mustAVM2Proof(t, limits)
	promise := mustAVM2Promise(t, msg.ID)
	abi := mustAVM2ABI(t)
	event := mustAVM2Event(t)

	program, err := NewAVM2Program(AVM2Program{
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		Instructions: []AVM2Instruction{
			{Opcode: AVM2OpPush, Value: []byte("deterministic")},
			{Opcode: AVM2OpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "balance"), Value: []byte("100")},
			{Opcode: AVM2OpKVGet, Key: AVMContractStorageKey(ctx.ContractAddress, "balance")},
			{Opcode: AVM2OpKVRangeBounded, Key: AVMContractStorageKey(ctx.ContractAddress, "prefix"), RangeLimit: 8},
			{Opcode: AVM2OpVerifyMerkleProof, Proof: proof},
			{Opcode: AVM2OpMsgSend, Message: msg},
			{Opcode: AVM2OpPromiseNew, Promise: promise},
			{Opcode: AVM2OpABIExport, ABI: abi},
			{Opcode: AVM2OpEmitEvent, Event: event},
		},
		MaxRecursionDepth: 1,
	}, limits, gasTable)
	require.NoError(t, err)

	result, err := ExecuteAVM2Program(program, ctx, limits, gasTable)
	require.NoError(t, err)
	require.NoError(t, result.Validate())
	require.Len(t, result.Stack, 1)
	require.Len(t, result.StorageWrites, 1)
	require.Len(t, result.StorageReads, 2)
	require.Len(t, result.OutputMessages, 1)
	require.Len(t, result.ProofsVerified, 1)
	require.Len(t, result.Promises, 1)
	require.Len(t, result.ABIDescriptors, 1)
	require.Len(t, result.Events, 1)
	require.Equal(t, ComputeAVM2StorageRoot(result.StorageReads, result.StorageWrites), result.StorageRoot)
	require.Equal(t, ComputeAVM2MessageRoot(result.OutputMessages), result.MessageRoot)
	require.Equal(t, ComputeAVM2PromiseRoot(result.Promises), result.PromiseRoot)
	require.Equal(t, ComputeAVM2ABIRoot(result.ABIDescriptors), result.ABIRoot)
	require.Equal(t, ComputeAVM2EventRoot(result.Events), result.EventRoot)

	shuffled := result
	shuffled.StorageReads = []AVM2StorageRead{result.StorageReads[1], result.StorageReads[0]}
	require.Equal(t, result.StorageRoot, ComputeAVM2StorageRoot(shuffled.StorageReads, shuffled.StorageWrites))
}

func TestAVM2ReadOnlySimulationCannotMutate(t *testing.T) {
	limits := DefaultAVM2Limits()
	gasTable := mustAVM2GasTable(t)
	ctx := mustAVM2Context(t, true)
	mutating, err := NewAVM2Program(AVM2Program{
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		Instructions:          []AVM2Instruction{{Opcode: AVM2OpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "k"), Value: []byte("v")}},
		MaxRecursionDepth:     1,
	}, limits, gasTable)
	require.NoError(t, err)
	_, err = ExecuteAVM2Program(mutating, ctx, limits, gasTable)
	require.ErrorContains(t, err, "read-only simulation")

	readonly, err := NewAVM2Program(AVM2Program{
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		Instructions:          []AVM2Instruction{{Opcode: AVM2OpKVGet, Key: AVMContractStorageKey(ctx.ContractAddress, "k")}},
		MaxRecursionDepth:     1,
	}, limits, gasTable)
	require.NoError(t, err)
	result, err := ExecuteAVM2Program(readonly, ctx, limits, gasTable)
	require.NoError(t, err)
	require.True(t, result.ReadOnlySimulation)
	require.Empty(t, result.StorageWrites)
	require.Empty(t, result.OutputMessages)
	require.Empty(t, result.Events)
	require.Empty(t, result.Promises)
}

func TestAVM2StoreV2RejectsRemotePrefix(t *testing.T) {
	limits := DefaultAVM2Limits()
	ctx := mustAVM2Context(t, false)
	err := ValidateAVM2StoreV2Key(ctx, AVMContractStorageKey("other-contract", "balance"), limits)
	require.ErrorContains(t, err, "contract-local prefix")
}

func mustAVM2GasTable(t *testing.T) AVM2GasTable {
	t.Helper()
	table, err := DefaultAVM2GasTable()
	require.NoError(t, err)
	return table
}

func mustAVM2Context(t *testing.T, readOnly bool) AVM2ExecutionContext {
	t.Helper()
	ctx := AVM2ExecutionContext{
		ChainID:         "aetra-local",
		Height:          10,
		ZoneID:          zonestypes.ZoneID("CONTRACT"),
		ShardID:         1,
		ContractAddress: "contract-a",
		Caller:          "caller-a",
		GasLimit:        10_000,
		ReadOnly:        readOnly,
	}
	ctx.ContextHash = ComputeAVM2ContextHash(ctx)
	require.NoError(t, ctx.Validate())
	return ctx
}

func mustAVM2Message(t *testing.T) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:         "aetra-local",
		Source:          "contract-a",
		Destination:     "contract-b",
		Payload:         []byte("call"),
		GasLimit:        100,
		ExpiryHeight:    20,
		RetryPolicy:     DefaultAVMRetryPolicy(20),
		BounceFlag:      true,
		SourceZone:      zonestypes.ZoneID("CONTRACT"),
		DestinationZone: zonestypes.ZoneID("IDENTITY"),
		SenderNonce:     7,
		PayloadType:     "contract.call",
		ValueNAET:       1,
		ForwardingFee:   1,
		Priority:        1,
		CreatedHeight:   10,
	})
	require.NoError(t, err)
	return msg
}

func mustAVM2Proof(t *testing.T, limits AVM2Limits) AVM2ProofInput {
	t.Helper()
	proof := AVM2ProofInput{
		ProofVersion: 1,
		ChainID:      "aetra-local",
		Height:       9,
		RootType:     AVM2ProofRootContractState,
		RootHash:     ComputeAVM2BytesHash([]byte("root")),
		Key:          "avm/contracts/storage/contract-b/value",
		ValueHash:    ComputeAVM2BytesHash([]byte("value")),
		ProofBytes:   []byte("merkle-branch"),
	}
	proof.ProofHash = ComputeAVM2ProofHash(proof)
	require.NoError(t, proof.Validate(limits))
	return proof
}

func mustAVM2Promise(t *testing.T, messageID string) AVM2PromiseState {
	t.Helper()
	promise := AVM2PromiseState{
		PromiseID:     ComputeAVM2BytesHash([]byte("promise-1")),
		Contract:      "contract-a",
		MessageID:     messageID,
		Status:        AVM2PromisePending,
		CreatedHeight: 10,
		ExpiryHeight:  20,
	}
	promise.PromiseHash = ComputeAVM2PromiseHash(promise)
	require.NoError(t, promise.Validate())
	return promise
}

func mustAVM2ABI(t *testing.T) AVM2ABIDescriptor {
	t.Helper()
	abi := AVM2ABIDescriptor{
		ABIVersion:    1,
		CodeID:        1,
		Methods:       []string{"transfer", "balance"},
		Events:        []string{"sent"},
		Errors:        []string{"insufficient_funds"},
		RequiredFunds: []string{"naet"},
		GasHints:      []string{"transfer/100"},
	}
	abi.InterfaceHash = ComputeAVM2ABIInterfaceHash(abi)
	require.NoError(t, abi.Validate(DefaultAVM2Limits()))
	return abi
}

func mustAVM2Event(t *testing.T) AVM2Event {
	t.Helper()
	event := AVM2Event{
		Height:          10,
		ContractAddress: "contract-a",
		EventID:         "event-1",
		Name:            "sent",
		PayloadHash:     ComputeAVM2BytesHash([]byte("payload")),
	}
	event.EventHash = ComputeAVM2EventHash(event)
	require.NoError(t, event.Validate(DefaultAVM2Limits()))
	return event
}
