package types

import (
	"bytes"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/wasmconfig"
	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

func TestContractZoneAVMDeployExecuteQueryExecutableSpec(t *testing.T) {
	state, code := uploadTestCode(t, counterSpecModule(), govAddr(), 10)

	state, contract, deployReceipt, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("salt"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)
	require.Equal(t, async.ResultOK, deployReceipt.ResultCode)
	require.Equal(t, uint64(1), avm.DecodeU64(StorageToAVM(contract.Storage)["counter"]))

	state, receipt, err := ExecuteAVMContract(state, ContractCall{
		Actor:		govAddr(),
		Contract:	contract.Address,
		Entrypoint:	avm.EntryReceiveExternal,
		GasLimit:	state.Policy.GasModel.ExecuteGas,
	}, 12)
	require.NoError(t, err)
	require.Equal(t, async.ResultOK, receipt.ResultCode)
	contract, found := findContract(state, contract.Address)
	require.True(t, found)
	require.Equal(t, uint64(2), avm.DecodeU64(StorageToAVM(contract.Storage)["counter"]))

	query, err := QueryAVMContract(state, ContractCall{
		Actor:			govAddr(),
		Contract:		contract.Address,
		GasLimit:		state.Policy.GasModel.QueryGas,
		QueryDepth:		1,
		QueryResponseBytes:	16,
	}, 13)
	require.NoError(t, err)
	require.Equal(t, async.ResultOK, query.ResultCode)
	require.Len(t, query.QueryResponse, 16)
	unchanged, _ := findContract(state, contract.Address)
	require.Equal(t, contract.Storage, unchanged.Storage)
}

func TestContractZoneRejectsUnauthorizedUploadInstantiateAndMigrate(t *testing.T) {
	state := EmptyContractZoneState(DefaultContractZonePolicy())

	_, _, err := UploadAVMCode(state, userAddr(2), counterSpecModule(), 10)
	require.ErrorContains(t, err, "governance authority")

	state, code, err := UploadAVMCode(state, govAddr(), counterSpecModule(), 10)
	require.NoError(t, err)
	_, _, _, err = InstantiateAVMContract(state, userAddr(2), code.CodeID, govAddr(), []byte("salt"), nil, 11, state.Policy.GasModel.DeployGas)
	require.ErrorContains(t, err, "code owner")

	state, contract, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("salt"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)
	_, _, err = MigrateAVMContract(state, ContractCall{Actor: userAddr(2), Contract: contract.Address, GasLimit: state.Policy.GasModel.ExecuteGas}, code.CodeID, 12)
	require.ErrorContains(t, err, "contract admin")

	state.Policy.MigrationsEnabled = false
	_, _, err = MigrateAVMContract(state, ContractCall{Actor: govAddr(), Contract: contract.Address, GasLimit: state.Policy.GasModel.ExecuteGas}, code.CodeID, 12)
	require.ErrorContains(t, err, "disabled by governance")
}

func TestContractZoneAllowlistIsTestnetOnly(t *testing.T) {
	state := EmptyContractZoneState(DefaultContractZonePolicy())
	state.Policy.UploadMode = UploadModeAllowlistTestnet
	state.Policy.UploadAllowlist = []sdkAddr{userAddr(2)}

	_, _, err := UploadAVMCode(state, userAddr(2), counterSpecModule(), 10)
	require.ErrorContains(t, err, "testnet-only")

	state.Policy.TestnetAllowlist = true
	next, _, err := UploadAVMCode(state, userAddr(2), counterSpecModule(), 10)
	require.NoError(t, err)
	require.Len(t, next.Codes, 1)
}

func TestContractZonePublicInstantiateRequiresExplicitPolicy(t *testing.T) {
	state, code := uploadTestCode(t, counterSpecModule(), govAddr(), 10)
	state.Policy.InstantiateMode = InstantiateModeEverybody

	next, contract, _, err := InstantiateAVMContract(state, userAddr(2), code.CodeID, userAddr(2), []byte("public"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)
	require.Equal(t, userAddr(2), contract.Owner)
	require.Len(t, next.Contracts, 1)
}

func TestContractZoneRejectsZeroAdminOversizedCodeAndState(t *testing.T) {
	state := EmptyContractZoneState(DefaultContractZonePolicy())
	state.Policy.Limits.MaxCodeSizeBytes = 16
	_, _, err := UploadAVMCode(state, govAddr(), counterSpecModule(), 10)
	require.ErrorContains(t, err, "code size")

	state = EmptyContractZoneState(DefaultContractZonePolicy())
	state, code, err := UploadAVMCode(state, govAddr(), counterSpecModule(), 10)
	require.NoError(t, err)
	_, _, _, err = InstantiateAVMContract(state, govAddr(), code.CodeID, bytes.Repeat([]byte{0}, 20), []byte("salt"), nil, 11, state.Policy.GasModel.DeployGas)
	require.ErrorContains(t, err, "zero address")

	state.Policy.Limits.MaxStateSizeBytes = 32
	initial := avm.Storage{"oversized": bytes.Repeat([]byte{1}, 64)}
	_, _, _, err = InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("salt"), initial, 11, state.Policy.GasModel.DeployGas)
	require.ErrorContains(t, err, "state size")
}

func TestContractZoneQueryGasAndResponseLimitsEnforced(t *testing.T) {
	state, code := uploadTestCode(t, counterSpecModule(), govAddr(), 10)
	state, contract, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("salt"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)

	_, err = QueryAVMContract(state, ContractCall{Actor: govAddr(), Contract: contract.Address, GasLimit: state.Policy.GasModel.QueryGas + 1, QueryDepth: 1}, 12)
	require.ErrorContains(t, err, "gas")
	_, err = QueryAVMContract(state, ContractCall{Actor: govAddr(), Contract: contract.Address, GasLimit: state.Policy.GasModel.QueryGas, QueryDepth: state.Policy.Limits.MaxQueryDepth + 1}, 12)
	require.ErrorContains(t, err, "query depth")
	_, err = QueryAVMContract(state, ContractCall{Actor: govAddr(), Contract: contract.Address, GasLimit: state.Policy.GasModel.QueryGas, QueryDepth: 1, QueryResponseBytes: state.Policy.Limits.MaxQueryResponseBytes + 1}, 12)
	require.ErrorContains(t, err, "query response")
}

func TestContractZoneGasLimitAndMessageLimitsEnforced(t *testing.T) {
	state, code := uploadTestCode(t, counterSpecModule(), govAddr(), 10)
	state, contract, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("salt"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)

	next, receipt, err := ExecuteAVMContract(state, ContractCall{Actor: govAddr(), Contract: contract.Address, Entrypoint: avm.EntryReceiveExternal, GasLimit: 100}, 12)
	require.NoError(t, err)
	require.Equal(t, async.ResultLimitExceeded, receipt.ResultCode)
	unchanged, _ := findContract(next, contract.Address)
	require.Equal(t, uint64(1), avm.DecodeU64(StorageToAVM(unchanged.Storage)["counter"]))

	state, emitCode := uploadTestCode(t, doubleEmitterSpecModule(), govAddr(), 20)
	state, emitting, _, err := InstantiateAVMContract(state, govAddr(), emitCode.CodeID, govAddr(), []byte("emit"), nil, 21, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)
	state.Policy.Limits.MaxEmittedMessages = 1
	_, _, err = ExecuteAVMContract(state, ContractCall{Actor: govAddr(), Contract: emitting.Address, Entrypoint: avm.EntryReceiveExternal, GasLimit: state.Policy.GasModel.ExecuteGas, EmitDestination: contract.Address}, 22)
	require.ErrorContains(t, err, "emitted messages")
}

func TestContractZoneExportImportPreservesStateAndQueues(t *testing.T) {
	state, code := uploadTestCode(t, emitterSpecModule(), govAddr(), 10)
	state, destination, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("dest"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)
	state, source, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("source"), nil, 12, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)

	state, _, err = ExecuteAVMContract(state, ContractCall{Actor: govAddr(), Contract: source.Address, Entrypoint: avm.EntryReceiveExternal, GasLimit: state.Policy.GasModel.ExecuteGas, EmitDestination: destination.Address}, 13)
	require.NoError(t, err)
	require.Len(t, state.QueuedMessages, 1)

	exported := state.Export()
	imported, err := ImportContractZoneState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported, imported))
}

func TestContractZoneCosmWasmStaysGovernanceGated(t *testing.T) {
	policy := DefaultRuntimePolicy()
	require.False(t, policy.CosmWasmEnabled)
	require.ErrorContains(t, ValidateVMCall(VMCall{Runtime: RuntimeCosmWasm, Action: ActionDeploy, CodeBytes: 1, GasLimit: 1}, policy), "disabled")

	wasm := wasmconfig.DefaultPolicy()
	wasm.Enabled = true
	wasm.GovernanceAuthority = governance
	wasm.MigrationsEnabled = false
	require.ErrorContains(t, wasmconfig.CanMigrate(wasmOwnerAddr, wasmOwnerAddr, wasm), "disabled by governance")
}

func FuzzContractZoneExecuteRejectsMalformedMessagesWithoutPanic(f *testing.F) {
	f.Add([]byte("seed"), uint64(1))
	f.Fuzz(func(t *testing.T, body []byte, gas uint64) {
		state, code := uploadTestCode(t, counterSpecModule(), govAddr(), 10)
		state, contract, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("salt"), nil, 11, state.Policy.GasModel.DeployGas)
		if err != nil {
			t.Fatal(err)
		}
		if gas == 0 {
			gas = 1
		}
		if gas > state.Policy.GasModel.ExecuteGas {
			gas = state.Policy.GasModel.ExecuteGas
		}
		_, _, _ = ExecuteAVMContract(state, ContractCall{Actor: govAddr(), Contract: contract.Address, Entrypoint: avm.EntryReceiveExternal, GasLimit: gas, Body: body}, 12)
	})
}

const wasmOwnerAddr = "4:0000000000000000000000000000000000000000000000000000000000000003"

type sdkAddr = sdk.AccAddress

func uploadTestCode(t *testing.T, module avm.Module, actor sdkAddr, height uint64) (ContractZoneState, ContractCode) {
	t.Helper()
	state := EmptyContractZoneState(DefaultContractZonePolicy())
	next, code, err := UploadAVMCode(state, actor, module, height)
	require.NoError(t, err)
	return next, code
}

func counterSpecModule() avm.Module {
	return avm.Module{
		Version:	avm.Version,
		Imports: []avm.HostFunction{
			avm.HostReadStorage,
			avm.HostWriteStorage,
			avm.HostChargeGas,
			avm.HostReturn,
		},
		Exports:	allSpecExports(),
		Code: []avm.Instruction{
			{Op: avm.OpReadStorage, Data: []byte("counter")},
			{Op: avm.OpPushU64, Arg: 1},
			{Op: avm.OpAdd},
			{Op: avm.OpWriteStorage, Data: []byte("counter")},
			{Op: avm.OpReturn, Arg: uint64(async.ResultOK)},
		},
	}
}

func emitterSpecModule() avm.Module {
	return avm.Module{
		Version:	avm.Version,
		Imports: []avm.HostFunction{
			avm.HostEmitInternal,
			avm.HostReturn,
		},
		Exports:	emitSpecExports(),
		Code: []avm.Instruction{
			{Op: avm.OpReturn, Arg: uint64(async.ResultOK)},
			{Op: avm.OpEmitInternal, Arg: 7, Data: []byte("out")},
			{Op: avm.OpReturn, Arg: uint64(async.ResultOK)},
		},
	}
}

func doubleEmitterSpecModule() avm.Module {
	module := emitterSpecModule()
	module.Code = []avm.Instruction{
		{Op: avm.OpReturn, Arg: uint64(async.ResultOK)},
		{Op: avm.OpEmitInternal, Arg: 7, Data: []byte("out-1")},
		{Op: avm.OpEmitInternal, Arg: 8, Data: []byte("out-2")},
		{Op: avm.OpReturn, Arg: uint64(async.ResultOK)},
	}
	return module
}

func allSpecExports() map[avm.Entrypoint]uint32 {
	return map[avm.Entrypoint]uint32{
		avm.EntryDeploy:		0,
		avm.EntryReceiveExternal:	0,
		avm.EntryReceiveInternal:	0,
		avm.EntryReceiveBounced:	0,
		avm.EntryQuery:			0,
		avm.EntryMigrate:		0,
	}
}

func emitSpecExports() map[avm.Entrypoint]uint32 {
	return map[avm.Entrypoint]uint32{
		avm.EntryDeploy:		0,
		avm.EntryReceiveExternal:	1,
		avm.EntryReceiveInternal:	1,
		avm.EntryReceiveBounced:	1,
		avm.EntryQuery:			0,
		avm.EntryMigrate:		0,
	}
}

func govAddr() sdkAddr {
	return []byte(DefaultContractZonePolicy().GovernanceAuthority)
}

func userAddr(fill byte) sdkAddr {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
