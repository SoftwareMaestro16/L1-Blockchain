package avm

import (
	"bytes"
	"reflect"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

func TestDeployValidModuleAndRejectMalformedBytecode(t *testing.T) {
	module := counterModule()
	encoded, err := EncodeModule(module)
	require.NoError(t, err)
	decoded, err := DecodeModule(encoded)
	require.NoError(t, err)
	require.Equal(t, module, decoded)
	hash, err := CodeHash(module)
	require.NoError(t, err)
	require.NotEqual(t, [32]byte{}, hash)

	verifier := newTestVerifier(t)
	require.NoError(t, verifier.Verify(decoded))
	_, err = DecodeModule([]byte("bad"))
	require.ErrorContains(t, err, "malformed")
	_, err = DecodeModule(append([]byte("BAD!"), encoded[4:]...))
	require.ErrorContains(t, err, "module header")
}

func TestBytecodeEncodeDecodeDifferentialRoundTrip(t *testing.T) {
	modules := []Module{counterModule(), emitterModule()}
	for _, module := range modules {
		encoded, err := EncodeModule(module)
		require.NoError(t, err)
		decoded, err := DecodeModule(encoded)
		require.NoError(t, err)
		reencoded, err := EncodeModule(decoded)
		require.NoError(t, err)
		require.Equal(t, encoded, reencoded)
	}
}

func TestVerifierRejectsOversizedCodeAndNondeterministicOpcode(t *testing.T) {
	params := DefaultParams()
	params.MaxInstructions = 2
	verifier, err := NewVerifier(params)
	require.NoError(t, err)
	oversized := counterModule()
	require.ErrorContains(t, verifier.Verify(oversized), "instruction count")

	params = DefaultParams()
	verifier, err = NewVerifier(params)
	require.NoError(t, err)
	forbidden := counterModule()
	forbidden.Code = []Instruction{{Op: OpWallClock}}
	require.ErrorContains(t, verifier.Verify(forbidden), "forbidden")

	badImport := counterModule()
	badImport.Imports = append(badImport.Imports, HostFunction(999))
	require.ErrorContains(t, verifier.Verify(badImport), "not allowed")

	missingImport := counterModule()
	missingImport.Imports = []HostFunction{HostReadStorage, HostReturn}
	require.ErrorContains(t, verifier.Verify(missingImport), "requires host function")
}

func TestRunSimpleCounterDeterministicallyAndBoundsGas(t *testing.T) {
	runner := newTestRunner(t)
	module := counterModule()
	ctx := runtimeCtx(EntryReceiveInternal)
	storage := Storage{"counter": EncodeU64(41)}

	first, err := runner.Run(module, storage, ctx)
	require.NoError(t, err)
	require.Equal(t, async.ResultOK, first.ResultCode)
	require.Equal(t, uint64(42), DecodeU64(first.State["counter"]))
	require.Equal(t, uint32(1), first.StorageWrites)

	second, err := runner.Run(module, storage, ctx)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(first.State, second.State))
	require.Equal(t, first.GasUsed, second.GasUsed)

	ctx.GasLimit = 1
	limited, err := runner.Run(module, storage, ctx)
	require.NoError(t, err)
	require.Equal(t, async.ResultLimitExceeded, limited.ResultCode)
}

func TestStorageSnapshotIsDeterministicAndBounded(t *testing.T) {
	storage := Storage{
		"z":	EncodeU64(26),
		"a":	EncodeU64(1),
		"m":	EncodeU64(13),
	}
	snapshot := Snapshot(storage)
	require.Equal(t, []string{"a", "m", "z"}, []string{snapshot[0].Key, snapshot[1].Key, snapshot[2].Key})
	require.Equal(t, EncodeSnapshot(storage), EncodeSnapshot(storage))

	params := DefaultParams()
	params.MaxMemoryBytes = 4
	runner, err := NewRunner(params)
	require.NoError(t, err)
	module := Module{
		Version:	Version,
		Imports: []HostFunction{
			HostWriteStorage,
			HostReturn,
		},
		Exports:	map[Entrypoint]uint32{EntryReceiveInternal: 0},
		Code: []Instruction{
			{Op: OpPushU64, Arg: 1},
			{Op: OpWriteStorage, Data: []byte("large-key")},
			{Op: OpReturn, Arg: uint64(async.ResultOK)},
		},
	}
	result, err := runner.Run(module, nil, runtimeCtx(EntryReceiveInternal))
	require.NoError(t, err)
	require.Equal(t, async.ResultLimitExceeded, result.ResultCode)
}

func TestSnapshotDecodeRoundTripRejectsNonCanonicalInput(t *testing.T) {
	storage := Storage{
		"a":	EncodeU64(1),
		"b":	EncodeU64(2),
	}
	decoded, err := DecodeSnapshot(EncodeSnapshot(storage))
	require.NoError(t, err)
	require.Equal(t, storage, decoded)

	buf := bytes.NewBuffer(nil)
	writeU32(buf, 2)
	writeU16(buf, 1)
	buf.WriteString("b")
	writeU32(buf, 1)
	buf.WriteByte(2)
	writeU16(buf, 1)
	buf.WriteString("a")
	writeU32(buf, 1)
	buf.WriteByte(1)
	_, err = DecodeSnapshot(buf.Bytes())
	require.ErrorContains(t, err, "sorted")
}

func TestExecutionProofIsDeterministicAndBindsStateContextAndTrace(t *testing.T) {
	runner := newTestRunner(t)
	module := counterModule()
	storage := Storage{"counter": EncodeU64(7)}
	ctx := runtimeCtx(EntryReceiveInternal)
	ctx.BlockHeight = 42

	exec, err := runner.Run(module, storage, ctx)
	require.NoError(t, err)
	first, err := BuildExecutionProof(module, storage, ctx, exec)
	require.NoError(t, err)
	second, err := BuildExecutionProof(module, storage, ctx, exec)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.NotEqual(t, [32]byte{}, ExecutionProofHash(first))

	exec.GasUsed++
	changed, err := BuildExecutionProof(module, storage, ctx, exec)
	require.NoError(t, err)
	require.NotEqual(t, ExecutionProofHash(first), ExecutionProofHash(changed))
}

func TestInterfaceManifestHashCommitsMetadataAndExports(t *testing.T) {
	manifest := InterfaceManifest{
		Name:		"counter",
		Version:	1,
		Methods: []InterfaceMethod{
			{Name: "increment", Entrypoint: EntryReceiveInternal, Opcode: 1, Async: true},
			{Name: "query", Entrypoint: EntryQuery, Opcode: 2},
		},
		Events:	[]InterfaceEvent{{Name: "incremented", Opcode: 10}},
	}
	hash, err := InterfaceHash(manifest)
	require.NoError(t, err)
	module := counterModule()
	module.MetadataHash = hash
	require.NoError(t, VerifyInterface(module, manifest))

	module.MetadataHash[0] ^= 1
	require.ErrorContains(t, VerifyInterface(module, manifest), "metadata hash")

	bad := manifest
	bad.Methods = append(bad.Methods, InterfaceMethod{Name: "again", Entrypoint: EntryReceiveInternal, Opcode: 1})
	_, err = InterfaceHash(bad)
	require.ErrorContains(t, err, "duplicate")
}

func TestHostContextOpcodesReadEnvelopeBlockAndChargeGas(t *testing.T) {
	runner := newTestRunner(t)
	module := Module{
		Version:	Version,
		Imports: []HostFunction{
			HostInspectMsg,
			HostBlockContext,
			HostChargeGas,
			HostWriteStorage,
			HostReturn,
		},
		Exports:	map[Entrypoint]uint32{EntryReceiveInternal: 0},
		Code: []Instruction{
			{Op: OpReadMsgOpcode},
			{Op: OpReadMsgQueryID},
			{Op: OpAdd},
			{Op: OpReadBlock},
			{Op: OpAdd},
			{Op: OpChargeGas, Arg: 7},
			{Op: OpWriteStorage, Data: []byte("sum")},
			{Op: OpReturn, Arg: uint64(async.ResultOK)},
		},
	}
	ctx := runtimeCtx(EntryReceiveInternal)
	ctx.Message.Opcode = 10
	ctx.Message.QueryID = 20
	ctx.BlockHeight = 5
	result, err := runner.Run(module, nil, ctx)
	require.NoError(t, err)
	require.Equal(t, async.ResultOK, result.ResultCode)
	require.Equal(t, uint64(35), DecodeU64(result.State["sum"]))
	require.GreaterOrEqual(t, result.GasUsed, uint64(7))

	ctx.GasLimit = 30
	limited, err := runner.Run(module, nil, ctx)
	require.NoError(t, err)
	require.Equal(t, async.ResultLimitExceeded, limited.ResultCode)
}

func TestAVMEmitsInternalMessageIntoAsyncQueue(t *testing.T) {
	runner := newTestRunner(t)
	module := emitterModule()
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	deployer := testAddr(1)
	source := deployAsyncContract(t, executor, deployer, []byte("source"), EncodeSnapshot(nil))
	destination := deployAsyncContract(t, executor, deployer, []byte("dest"))
	require.NoError(t, executor.RegisterHandler(source, runner.AsyncHandler(module, nil, RuntimeContext{
		EmitDestination: destination,
	})))
	require.NoError(t, executor.RegisterHandler(destination, func(contract async.ContractAccount, msg async.MessageEnvelope) async.ExecutionResult {
		return async.ExecutionResult{NewState: append([]byte("dest:"), msg.Body...), ResultCode: async.ResultOK}
	}))

	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{testAsyncMessage(testAddr(9), source, 1)}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	require.Equal(t, uint32(77), receipts[1].Opcode)
	require.Equal(t, async.ResultOK, receipts[1].ResultCode)
	contract, ok := executor.Contract(destination)
	require.True(t, ok)
	require.Equal(t, []byte("dest:avm-out"), contract.State)
}

func TestAVMAsyncHandlerRestoresSnapshotStateAcrossMessages(t *testing.T) {
	runner := newTestRunner(t)
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	deployer := testAddr(1)
	initial := EncodeSnapshot(Storage{"counter": EncodeU64(0)})
	contract := deployAsyncContract(t, executor, deployer, []byte("stateful"), initial)
	require.NoError(t, executor.RegisterHandler(contract, runner.AsyncHandler(counterModule(), nil, RuntimeContext{})))

	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{
		testAsyncMessage(testAddr(9), contract, 1),
		testAsyncMessage(testAddr(9), contract, 2),
	}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	account, ok := executor.Contract(contract)
	require.True(t, ok)
	storage, err := DecodeSnapshot(account.State)
	require.NoError(t, err)
	require.Equal(t, uint64(2), DecodeU64(storage["counter"]))
}

func TestAVMSchedulesSelfContinuationAtFutureBlock(t *testing.T) {
	runner := newTestRunner(t)
	module := continuationModule()
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	deployer := testAddr(1)
	contract := deployAsyncContract(t, executor, deployer, []byte("cont"), EncodeSnapshot(nil))
	require.NoError(t, executor.RegisterHandler(contract, runner.AsyncHandler(module, nil, RuntimeContext{})))

	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{testAsyncMessage(testAddr(9), contract, 1)}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Len(t, executor.Queue(), 1)
	require.Equal(t, uint64(3), executor.Queue()[0].Envelope.DeliverAtBlock)
	require.True(t, executor.Queue()[0].Envelope.Destination.Equals(contract))

	receipts, err = executor.ProcessBlock(2)
	require.NoError(t, err)
	require.Empty(t, receipts)
	require.Len(t, executor.Queue(), 1)

	receipts, err = executor.ProcessBlock(3)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
}

func TestAVMAsyncFailedSendBouncesAndQueueSurvivesExportImport(t *testing.T) {
	runner := newTestRunner(t)
	module := emitterModule()
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	deployer := testAddr(1)
	source := deployAsyncContract(t, executor, deployer, []byte("source"), EncodeSnapshot(nil))
	missingDestination, err := async.DeriveContractAddress(deployer, bytes.Repeat([]byte{9}, async.CodeHashLength), []byte("missing"))
	require.NoError(t, err)
	require.NoError(t, executor.RegisterHandler(source, runner.AsyncHandler(module, nil, RuntimeContext{
		EmitDestination: missingDestination,
	})))

	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{testAsyncMessage(testAddr(9), source, 1)}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 3)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	require.Equal(t, async.ResultNoDestination, receipts[1].ResultCode)
	require.Equal(t, async.BounceOpcode, receipts[2].Opcode)

	exported := executor.ExportState()
	imported, err := async.ImportState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported, imported.ExportState()))
}

func counterModule() Module {
	return Module{
		Version:	Version,
		Imports: []HostFunction{
			HostReadStorage,
			HostWriteStorage,
			HostChargeGas,
			HostReturn,
		},
		Exports: map[Entrypoint]uint32{
			EntryDeploy:		0,
			EntryReceiveExternal:	0,
			EntryReceiveInternal:	0,
			EntryReceiveBounced:	0,
			EntryQuery:		0,
			EntryMigrate:		0,
		},
		Code: []Instruction{
			{Op: OpReadStorage, Data: []byte("counter")},
			{Op: OpPushU64, Arg: 1},
			{Op: OpAdd},
			{Op: OpWriteStorage, Data: []byte("counter")},
			{Op: OpReturn, Arg: uint64(async.ResultOK)},
		},
	}
}

func emitterModule() Module {
	return Module{
		Version:	Version,
		Imports: []HostFunction{
			HostEmitInternal,
			HostReturn,
		},
		Exports:	map[Entrypoint]uint32{EntryReceiveInternal: 0},
		Code: []Instruction{
			{Op: OpEmitInternal, Arg: 77, Data: []byte("avm-out")},
			{Op: OpReturn, Arg: uint64(async.ResultOK)},
		},
	}
}

func continuationModule() Module {
	return Module{
		Version:	Version,
		Imports: []HostFunction{
			HostScheduleSelf,
			HostReturn,
		},
		Exports:	map[Entrypoint]uint32{EntryReceiveInternal: 0},
		Code: []Instruction{
			{Op: OpScheduleSelf, Arg: 2, Data: []byte("resume")},
			{Op: OpReturn, Arg: uint64(async.ResultOK)},
		},
	}
}

func newTestVerifier(t *testing.T) *Verifier {
	t.Helper()
	verifier, err := NewVerifier(DefaultParams())
	require.NoError(t, err)
	return verifier
}

func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	runner, err := NewRunner(DefaultParams())
	require.NoError(t, err)
	return runner
}

func runtimeCtx(entry Entrypoint) RuntimeContext {
	return RuntimeContext{
		Entry:		entry,
		Message:	testAsyncMessage(testAddr(9), testAddr(8), 1),
		GasLimit:	100_000,
	}
}

func deployAsyncContract(t *testing.T, executor *async.Executor, deployer sdk.AccAddress, salt []byte, state ...[]byte) sdk.AccAddress {
	t.Helper()
	initial := []byte(nil)
	if len(state) > 0 {
		initial = state[0]
	}
	address, err := executor.DeployContract(deployer, bytes.Repeat([]byte{salt[0]}, async.CodeHashLength), salt, initial, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	return address
}

func testAsyncMessage(source, destination sdk.AccAddress, queryID uint64) async.MessageEnvelope {
	return async.MessageEnvelope{
		Source:			source,
		Destination:		destination,
		Value:			sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
		Opcode:			1,
		QueryID:		queryID,
		Body:			[]byte("in"),
		Bounce:			true,
		CreatedLogicalTime:	queryID,
		GasLimit:		100_000,
		ForwardFee:		sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
	}
}

func testAddr(fill byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{fill}, 20))
}
