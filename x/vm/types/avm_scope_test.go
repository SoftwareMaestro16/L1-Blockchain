package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMProgramRejectsForbiddenSurfaces(t *testing.T) {
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	for _, opcode := range []AVMOpcode{
		AVMOpExternalNetwork,
		AVMOpWallClock,
		AVMOpNonDeterministic,
		AVMOpKVRangeUnbounded,
		AVMOpDirectRemoteMutation,
		AVMOpUnboundedRecursion,
	} {
		program := AVMProgram{
			VMVersion:		AVMVMVersion,
			InstructionSetVersion:	AVMDefaultInstructionSet,
			Instructions:		[]AVMInstruction{{Opcode: opcode}},
			MaxRecursionDepth:	1,
		}
		program.ProgramHash = ComputeAVMProgramHash(program)
		require.Error(t, ValidateAVMProgram(program, limits, gasTable), string(opcode))
	}
}

func TestAVMInstructionSetCoversSpecAndGasTable(t *testing.T) {
	set, err := DefaultAVMInstructionSet()
	require.NoError(t, err)
	require.Len(t, set.Opcodes, len(AllAVMSupportedOpcodes()))
	require.Equal(t, ComputeAVMInstructionSetHash(set), set.SetHash)

	byCategory := map[AVMInstructionCategory]int{}
	for _, descriptor := range set.Opcodes {
		byCategory[descriptor.Category]++
	}
	for _, category := range []AVMInstructionCategory{
		AVMCategoryCoreStack,
		AVMCategoryArithmetic,
		AVMCategoryControlFlow,
		AVMCategoryMemory,
		AVMCategoryStorage,
		AVMCategoryCryptoProof,
		AVMCategoryMessages,
		AVMCategoryPromises,
		AVMCategoryABI,
		AVMCategoryContext,
	} {
		require.Positive(t, byCategory[category], string(category))
	}

	table := mustAVMGasTable(t)
	for _, opcode := range AllAVMSupportedOpcodes() {
		gas, found := table.OpcodeGas(opcode)
		require.True(t, found, string(opcode))
		require.Positive(t, gas)
	}

	missing := table
	missing.OpcodeCosts = missing.OpcodeCosts[1:]
	missing.TableHash = ComputeAVMGasTableHash(missing)
	require.ErrorContains(t, missing.Validate(), "missing opcode")
}

func TestAVMExecutionEnforcesBoundsAndGas(t *testing.T) {
	limits := DefaultAVMLimits()
	limits.MaxStackDepth = 1
	limits.MaxMemoryBytes = 8
	gasTable := mustAVMGasTable(t)
	ctx := mustAVMContext(t, false)

	stackOverflow, err := NewAVMProgram(AVMProgram{
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		Instructions: []AVMInstruction{
			{Opcode: AVMOpPush, Value: []byte("a")},
			{Opcode: AVMOpPush, Value: []byte("b")},
		},
		MaxRecursionDepth:	1,
	}, limits, gasTable)
	require.NoError(t, err)
	_, err = ExecuteAVMProgram(stackOverflow, ctx, limits, gasTable)
	require.ErrorContains(t, err, "stack depth")

	memoryOverflow, err := NewAVMProgram(AVMProgram{
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		Instructions:		[]AVMInstruction{{Opcode: AVMOpMemGrow, MemoryGrow: 9}},
		MaxRecursionDepth:	1,
	}, limits, gasTable)
	require.NoError(t, err)
	_, err = ExecuteAVMProgram(memoryOverflow, ctx, limits, gasTable)
	require.ErrorContains(t, err, "memory limit")

	lowGas := ctx
	lowGas.GasLimit = 1
	lowGas.ContextHash = ComputeAVMContextHash(lowGas)
	_, err = ExecuteAVMProgram(memoryOverflow, lowGas, limits, gasTable)
	require.ErrorContains(t, err, "exhausted gas")
}

func TestAVMInstructionGasChargesResourceInputs(t *testing.T) {
	limits := DefaultAVMLimits()
	table := mustAVMGasTable(t)
	ctx := mustAVMContext(t, false)

	smallWrite := AVMInstruction{Opcode: AVMOpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "a"), Value: []byte("x")}
	largeWrite := AVMInstruction{Opcode: AVMOpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "a"), Value: []byte("xxxxxxxx")}
	smallGas, err := AVMInstructionGas(smallWrite, table, limits)
	require.NoError(t, err)
	largeGas, err := AVMInstructionGas(largeWrite, table, limits)
	require.NoError(t, err)
	require.Greater(t, largeGas, smallGas)

	proof := mustAVMProof(t, limits)
	proofOp := AVMInstruction{Opcode: AVMOpVerifyMerkleProof, Proof: proof}
	proofGas, err := AVMInstructionGas(proofOp, table, limits)
	require.NoError(t, err)
	proof.ProofBytes = append(proof.ProofBytes, []byte("more-proof")...)
	proof.ProofVersion++
	proof.ProofHash = ComputeAVMProofHash(proof)
	largerProofGas, err := AVMInstructionGas(AVMInstruction{Opcode: AVMOpVerifyMerkleProof, Proof: proof}, table, limits)
	require.NoError(t, err)
	require.Greater(t, largerProofGas, proofGas)

	msg := mustAVMMessage(t)
	builderGas, err := AVMInstructionGas(AVMInstruction{Opcode: AVMOpMsgNew, Message: msg}, table, limits)
	require.NoError(t, err)
	sendGas, err := AVMInstructionGas(AVMInstruction{Opcode: AVMOpMsgSend, Message: msg}, table, limits)
	require.NoError(t, err)
	require.Greater(t, sendGas, builderGas)
}

func TestAVMContextAndStackOpcodesAreDeterministic(t *testing.T) {
	limits := DefaultAVMLimits()
	table := mustAVMGasTable(t)
	ctx := mustAVMContext(t, false)

	program, err := NewAVMProgram(AVMProgram{
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		Instructions: []AVMInstruction{
			{Opcode: AVMOpCtxHeight},
			{Opcode: AVMOpDup},
			{Opcode: AVMOpSwap},
			{Opcode: AVMOpPop},
			{Opcode: AVMOpCtxChainID},
			{Opcode: AVMOpCtxGasLeft},
		},
		MaxRecursionDepth:	1,
	}, limits, table)
	require.NoError(t, err)

	result, err := ExecuteAVMProgram(program, ctx, limits, table)
	require.NoError(t, err)
	require.Len(t, result.Stack, 3)
	require.Equal(t, "00000000000000000010", result.Stack[0])
	require.Equal(t, "aetra-local", result.Stack[1])
	require.NotEmpty(t, result.Stack[2])
}

func TestAVMScopeCommitsStoreMessagesProofsABIEvents(t *testing.T) {
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	ctx := mustAVMContext(t, false)
	msg := mustAVMMessage(t)
	proof := mustAVMProof(t, limits)
	promise := mustAVMPromise(t, msg.ID)
	abi := mustAVMABI(t)
	event := mustAVMEvent(t)

	program, err := NewAVMProgram(AVMProgram{
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		Instructions: []AVMInstruction{
			{Opcode: AVMOpPush, Value: []byte("deterministic")},
			{Opcode: AVMOpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "balance"), Value: []byte("100")},
			{Opcode: AVMOpKVGet, Key: AVMContractStorageKey(ctx.ContractAddress, "balance")},
			{Opcode: AVMOpKVRangeBounded, Key: AVMContractStorageKey(ctx.ContractAddress, "prefix"), RangeLimit: 8},
			{Opcode: AVMOpVerifyMerkleProof, Proof: proof},
			{Opcode: AVMOpMsgSend, Message: msg},
			{Opcode: AVMOpPromiseNew, Promise: promise},
			{Opcode: AVMOpABIExport, ABI: abi},
			{Opcode: AVMOpEmitEvent, Event: event},
		},
		MaxRecursionDepth:	1,
	}, limits, gasTable)
	require.NoError(t, err)

	result, err := ExecuteAVMProgram(program, ctx, limits, gasTable)
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
	require.Equal(t, ComputeAVMStorageRoot(result.StorageReads, result.StorageWrites), result.StorageRoot)
	require.Equal(t, ComputeAVMMessageRoot(result.OutputMessages), result.MessageRoot)
	require.Equal(t, ComputeAVMPromiseRoot(result.Promises), result.PromiseRoot)
	require.Equal(t, ComputeAVMABIRoot(result.ABIDescriptors), result.ABIRoot)
	require.Equal(t, ComputeAVMEventRoot(result.Events), result.EventRoot)

	shuffled := result
	shuffled.StorageReads = []AVMStorageRead{result.StorageReads[1], result.StorageReads[0]}
	require.Equal(t, result.StorageRoot, ComputeAVMStorageRoot(shuffled.StorageReads, shuffled.StorageWrites))
}

func TestAVMReadOnlySimulationCannotMutate(t *testing.T) {
	limits := DefaultAVMLimits()
	gasTable := mustAVMGasTable(t)
	ctx := mustAVMContext(t, true)
	mutating, err := NewAVMProgram(AVMProgram{
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		Instructions:		[]AVMInstruction{{Opcode: AVMOpKVSet, Key: AVMContractStorageKey(ctx.ContractAddress, "k"), Value: []byte("v")}},
		MaxRecursionDepth:	1,
	}, limits, gasTable)
	require.NoError(t, err)
	_, err = ExecuteAVMProgram(mutating, ctx, limits, gasTable)
	require.ErrorContains(t, err, "read-only simulation")

	readonly, err := NewAVMProgram(AVMProgram{
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		Instructions:		[]AVMInstruction{{Opcode: AVMOpKVGet, Key: AVMContractStorageKey(ctx.ContractAddress, "k")}},
		MaxRecursionDepth:	1,
	}, limits, gasTable)
	require.NoError(t, err)
	result, err := ExecuteAVMProgram(readonly, ctx, limits, gasTable)
	require.NoError(t, err)
	require.True(t, result.ReadOnlySimulation)
	require.Empty(t, result.StorageWrites)
	require.Empty(t, result.OutputMessages)
	require.Empty(t, result.Events)
	require.Empty(t, result.Promises)
}

func TestAVMStoreV2RejectsRemotePrefix(t *testing.T) {
	limits := DefaultAVMLimits()
	ctx := mustAVMContext(t, false)
	err := ValidateAVMStoreV2Key(ctx, AVMContractStorageKey("other-contract", "balance"), limits)
	require.ErrorContains(t, err, "contract-local prefix")
}

func mustAVMGasTable(t *testing.T) AVMGasTable {
	t.Helper()
	table, err := DefaultAVMGasTable()
	require.NoError(t, err)
	return table
}

func mustAVMContext(t *testing.T, readOnly bool) AVMExecutionContext {
	t.Helper()
	ctx := AVMExecutionContext{
		ChainID:		"aetra-local",
		Height:			10,
		ZoneID:			zonestypes.ZoneID("CONTRACT"),
		ShardID:		1,
		ContractAddress:	"contract-a",
		Caller:			"caller-a",
		GasLimit:		10_000,
		ReadOnly:		readOnly,
	}
	ctx.ContextHash = ComputeAVMContextHash(ctx)
	require.NoError(t, ctx.Validate())
	return ctx
}

func mustAVMMessage(t *testing.T) AVMAsyncMessage {
	t.Helper()
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:		"aetra-local",
		Source:			"contract-a",
		Destination:		"contract-b",
		Payload:		[]byte("call"),
		GasLimit:		100,
		ExpiryHeight:		20,
		RetryPolicy:		DefaultAVMRetryPolicy(20),
		BounceFlag:		true,
		SourceZone:		zonestypes.ZoneID("CONTRACT"),
		DestinationZone:	zonestypes.ZoneID("IDENTITY"),
		SenderNonce:		7,
		PayloadType:		"contract.call",
		ValueNAET:		1,
		ForwardingFee:		1,
		Priority:		1,
		CreatedHeight:		10,
	})
	require.NoError(t, err)
	return msg
}

func mustAVMProof(t *testing.T, limits AVMLimits) AVMProofInput {
	t.Helper()
	proof := AVMProofInput{
		ProofVersion:	1,
		ChainID:	"aetra-local",
		Height:		9,
		RootType:	AVMProofRootContractState,
		RootHash:	ComputeAVMBytesHash([]byte("root")),
		Key:		"avm/contracts/storage/contract-b/value",
		ValueHash:	ComputeAVMBytesHash([]byte("value")),
		ProofBytes:	[]byte("merkle-branch"),
	}
	proof.ProofHash = ComputeAVMProofHash(proof)
	require.NoError(t, proof.Validate(limits))
	return proof
}

func mustAVMPromise(t *testing.T, messageID string) AVMPromiseState {
	t.Helper()
	promise := AVMPromiseState{
		PromiseID:	ComputeAVMBytesHash([]byte("promise-1")),
		Contract:	"contract-a",
		MessageID:	messageID,
		Status:		AVMPromisePending,
		CreatedHeight:	10,
		ExpiryHeight:	20,
	}
	promise.PromiseHash = ComputeAVMPromiseHash(promise)
	require.NoError(t, promise.Validate())
	return promise
}

func mustAVMABI(t *testing.T) AVMABIDescriptor {
	t.Helper()
	abi := AVMABIDescriptor{
		ABIVersion:	1,
		CodeID:		1,
		Methods:	[]string{"transfer", "balance"},
		Events:		[]string{"sent"},
		Errors:		[]string{"insufficient_funds"},
		RequiredFunds:	[]string{"naet"},
		GasHints:	[]string{"transfer/100"},
	}
	abi.InterfaceHash = ComputeAVMABIInterfaceHash(abi)
	require.NoError(t, abi.Validate(DefaultAVMLimits()))
	return abi
}

func mustAVMEvent(t *testing.T) AVMEvent {
	t.Helper()
	event := AVMEvent{
		Height:			10,
		ContractAddress:	"contract-a",
		EventID:		"event-1",
		Name:			"sent",
		PayloadHash:		ComputeAVMBytesHash([]byte("payload")),
	}
	event.EventHash = ComputeAVMEventHash(event)
	require.NoError(t, event.Validate(DefaultAVMLimits()))
	return event
}
