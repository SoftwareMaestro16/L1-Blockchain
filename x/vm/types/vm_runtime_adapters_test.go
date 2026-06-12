package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/wasmconfig"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestVMRuntimeTraitCommitsRequiredRuntimeCapabilities(t *testing.T) {
	trait, err := NewVMRuntimeTrait(VMRuntimeTrait{Runtime: RuntimeAVM, AdapterKind: VMAdapterAVM})
	require.NoError(t, err)
	require.NoError(t, trait.Validate())
	require.True(t, trait.SupportsBytecodeDeploy)
	require.True(t, trait.SupportsStorageAdapter)
	require.True(t, trait.SupportsOutboundMessage)
	require.True(t, trait.EmitsReceipts)
	require.True(t, trait.CommitsVMRoot)
	require.Equal(t, ComputeVMRuntimeTraitHash(trait), trait.TraitHash)

	missingRoot := trait
	missingRoot.CommitsVMRoot = false
	missingRoot.TraitHash = ComputeVMRuntimeTraitHash(missingRoot)
	require.ErrorContains(t, missingRoot.Validate(), "root capabilities")

	wrongAdapter := trait
	wrongAdapter.AdapterKind = VMAdapterCosmWasm
	wrongAdapter.TraitHash = ComputeVMRuntimeTraitHash(wrongAdapter)
	require.ErrorContains(t, wrongAdapter.Validate(), "AVM adapter")
}

func TestDeterministicVMBytecodeValidationRejectsForbiddenAVMOpcodes(t *testing.T) {
	policy := DefaultRuntimePolicy()
	bytecode := testVMRuntimeAVMBytecode(t, avm.Instruction{Op: avm.OpReturn, Arg: 0})
	validation, err := ValidateDeterministicVMBytecode(RuntimeAVM, bytecode, policy)
	require.NoError(t, err)
	require.NoError(t, validation.Validate())
	require.True(t, validation.Deterministic)
	require.True(t, validation.Validated)
	require.Equal(t, ComputeVMBytecodeValidationHash(validation), validation.ValidationHash)

	forbidden := testVMRuntimeAVMBytecode(t, avm.Instruction{Op: avm.OpRandom})
	_, err = ValidateDeterministicVMBytecode(RuntimeAVM, forbidden, policy)
	require.ErrorContains(t, err, "nondeterministic")
}

func TestAVMRuntimeAdapterEmitsOutboundMessageReceiptAndRootCommitment(t *testing.T) {
	zoneID := zonestypes.ZoneIDContract
	policy := DefaultRuntimePolicy()
	manifest, err := NewVMAdapterBoundaryManifest(VMAdapterBoundaryManifest{ZoneID: zoneID})
	require.NoError(t, err)
	trait, err := NewVMRuntimeTrait(VMRuntimeTrait{Runtime: RuntimeAVM, AdapterKind: VMAdapterAVM})
	require.NoError(t, err)
	bytecode, err := ValidateDeterministicVMBytecode(RuntimeAVM, testVMRuntimeAVMBytecode(t, avm.Instruction{Op: avm.OpReturn, Arg: 7}), policy)
	require.NoError(t, err)
	gasTable, err := NewVMGasTable(RuntimeAVM, policy)
	require.NoError(t, err)
	storageAdapter, err := NewVMStorageAdapter(RuntimeAVM, zoneID, manifest)
	require.NoError(t, err)
	outbound, err := NewVMOutboundMessageSyscall(RuntimeAVM, manifest.AVM.MessageSyscall, VMOutboundMessageRequest{
		ChainID:		"aetra-test-1",
		Source:			"contract-a",
		Destination:		"app-b",
		SourceZone:		zoneID,
		DestinationZone:	zonestypes.ZoneIDApplication,
		Payload:		[]byte("advance"),
		PayloadType:		"application.workflow",
		GasLimit:		100,
		ForwardingFee:		1,
		SenderNonce:		3,
		CreatedHeight:		10,
		ExpiryHeight:		50,
		RouteHint:		"contract/app",
	})
	require.NoError(t, err)
	receipt, err := NewVMReceiptEmission(RuntimeAVM, AVMExecutionReceipt{
		MessageID:		outbound.Message.ID,
		ZoneID:			zoneID,
		Executor:		"avm-runtime",
		Status:			AVMReceiptStatusExecuted,
		GasUsed:		77,
		StorageWritten:		1,
		EventsHash:		engineHash("events"),
		OutputMessagesRoot:	ComputeAVMZoneOutputMessageRoot(zoneID, []AVMAsyncMessage{outbound.Message}),
		CreatedHeight:		11,
	}, zoneID)
	require.NoError(t, err)

	adapter, err := NewVMRuntimeAdapter(VMRuntimeAdapter{
		Trait:			trait,
		BoundaryManifest:	manifest,
		BytecodeValidation:	bytecode,
		GasTable:		gasTable,
		StorageAdapter:		storageAdapter,
		OutboundSyscalls:	[]VMOutboundMessageSyscall{outbound},
		ReceiptEmission:	receipt,
	})
	require.NoError(t, err)
	require.NoError(t, adapter.Validate())
	require.Equal(t, ComputeVMRuntimeAdapterHash(adapter), adapter.AdapterHash)

	root, err := NewVMRuntimeRootCommitment(11, zoneID, adapter)
	require.NoError(t, err)
	require.NoError(t, root.Validate())
	require.Equal(t, adapter.AdapterHash, root.AdapterHash)
	require.Equal(t, bytecode.BytecodeHash, root.BytecodeHash)
	require.Equal(t, ComputeVMRuntimeRootCommitmentHash(root), root.VMRootHash)

	tampered := root
	tampered.ReceiptRoot = engineHash("different-receipt-root")
	require.ErrorContains(t, tampered.Validate(), "root hash mismatch")
}

func TestCosmWasmRuntimeAdapterUsesExplicitGasAndStorageBoundary(t *testing.T) {
	zoneID := zonestypes.ZoneIDContract
	policy := DefaultRuntimePolicy()
	policy.CosmWasmEnabled = true
	policy.CosmWasmPolicy = wasmconfig.DefaultPolicy()
	policy.CosmWasmPolicy.Enabled = true
	policy.CosmWasmPolicy.GovernanceAuthority = governance

	profile, err := DefaultVMDeterminismProfile(RuntimeCosmWasm)
	require.NoError(t, err)
	manifest, err := NewVMAdapterBoundaryManifest(VMAdapterBoundaryManifest{
		ZoneID:			zoneID,
		DeterminismProfile:	profile,
	})
	require.NoError(t, err)
	trait, err := NewVMRuntimeTrait(VMRuntimeTrait{Runtime: RuntimeCosmWasm, AdapterKind: VMAdapterCosmWasm, DeterminismProfile: profile})
	require.NoError(t, err)
	bytecode, err := ValidateDeterministicVMBytecode(RuntimeCosmWasm, []byte{0x00, 0x61, 0x73, 0x6d}, policy)
	require.NoError(t, err)
	gasTable, err := NewVMGasTable(RuntimeCosmWasm, policy)
	require.NoError(t, err)
	require.Equal(t, wasmconfig.DefaultGasMultiplier, gasTable.WASMConversion.AVMGasPerWASMGas)
	storageAdapter, err := NewVMStorageAdapter(RuntimeCosmWasm, zoneID, manifest)
	require.NoError(t, err)
	require.Equal(t, ContractZoneKVPrefix(zoneID), storageAdapter.KeyPrefix)
	require.Equal(t, manifest.CosmWasm.GasConversion.StorageReadGas, storageAdapter.ReadGas)
	require.Equal(t, manifest.CosmWasm.GasConversion.StorageWriteGas, storageAdapter.WriteGas)

	receipt, err := NewVMReceiptEmission(RuntimeCosmWasm, AVMExecutionReceipt{
		MessageID:		engineHash("wasm-message"),
		ZoneID:			zoneID,
		Executor:		"cosmwasm-adapter",
		Status:			AVMReceiptStatusExecuted,
		GasUsed:		100,
		EventsHash:		engineHash("wasm-events"),
		OutputMessagesRoot:	ComputeAVMZoneOutputMessageRoot(zoneID, nil),
		CreatedHeight:		12,
	}, zoneID)
	require.NoError(t, err)
	adapter, err := NewVMRuntimeAdapter(VMRuntimeAdapter{
		Trait:			trait,
		BoundaryManifest:	manifest,
		BytecodeValidation:	bytecode,
		GasTable:		gasTable,
		StorageAdapter:		storageAdapter,
		ReceiptEmission:	receipt,
	})
	require.NoError(t, err)
	require.NoError(t, adapter.Validate())
	require.Equal(t, ComputeVMRuntimeAdapterHash(adapter), adapter.AdapterHash)
}

func testVMRuntimeAVMBytecode(t *testing.T, code ...avm.Instruction) []byte {
	t.Helper()
	module := avm.Module{
		Version:	avm.Version,
		Imports: []avm.HostFunction{
			avm.HostReturn,
		},
		Exports: map[avm.Entrypoint]uint32{
			avm.EntryReceiveExternal: 0,
		},
		Code:	code,
	}
	encoded, err := avm.EncodeModule(module)
	require.NoError(t, err)
	return encoded
}
