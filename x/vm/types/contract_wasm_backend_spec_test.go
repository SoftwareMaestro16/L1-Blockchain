package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/wasmconfig"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMWASMSandboxPolicyModelsOptionalSandboxGasAndStoreV2Adapter(t *testing.T) {
	policy := testAVMWASMSandboxPolicy(t)

	require.NoError(t, policy.Validate())
	require.True(t, policy.Enabled)
	require.True(t, policy.Optional)
	require.False(t, policy.ExternalNetwork)
	require.True(t, policy.CrossZoneAsyncOnly)
	require.Equal(t, wasmconfig.DefaultGasMultiplier, policy.GasConversion.AVMGasPerWASMGas)
	require.Equal(t, DefaultAVMStoreKey, policy.StoreAdapter.StoreKey)
	require.Equal(t, ContractZoneKVPrefix(zonestypes.ZoneIDContract), policy.StoreAdapter.KeyPrefix)
	require.Equal(t, ComputeAVMWASMSandboxPolicyHash(policy), policy.SandboxPolicyHash)
}

func TestAVMWASMSandboxRejectsNondeterminismNetworkMemoryAndGasDrift(t *testing.T) {
	policy := testAVMWASMSandboxPolicy(t)

	nondeterministic := policy
	nondeterministic.HostFunctions = append([]AVMWASMHostFunction(nil), policy.HostFunctions...)
	nondeterministic.HostFunctions[0].Deterministic = false
	nondeterministic.SandboxPolicyHash = ComputeAVMWASMSandboxPolicyHash(nondeterministic)
	require.ErrorContains(t, nondeterministic.Validate(), "deterministic")

	network := policy
	network.ExternalNetwork = true
	network.SandboxPolicyHash = ComputeAVMWASMSandboxPolicyHash(network)
	require.ErrorContains(t, network.Validate(), "external network")

	unboundedMemory := policy
	unboundedMemory.MaxMemoryPages = uint32((uint64(policy.RuntimePolicy.MemoryCacheSizeMiB)*1024*1024)/WASMMemoryPageBytes) + 1
	unboundedMemory.SandboxPolicyHash = ComputeAVMWASMSandboxPolicyHash(unboundedMemory)
	require.ErrorContains(t, unboundedMemory.Validate(), "memory bound")

	badGas := policy
	badGas.GasConversion.AVMGasPerWASMGas++
	badGas.GasConversion.ConversionHash = ComputeAVMWASMGasConversionHash(badGas.GasConversion)
	badGas.SandboxPolicyHash = ComputeAVMWASMSandboxPolicyHash(badGas)
	require.ErrorContains(t, badGas.Validate(), "gas conversion multiplier")
}

func TestAVMWASMContractRouteAllowsCrossZoneOnlyViaAsyncMessages(t *testing.T) {
	policy := testAVMWASMSandboxPolicy(t)
	msg := testAVMWASMCrossZoneMessage(t)
	call, err := NewAVMWASMContractRouteCall(AVMWASMContractRouteCall{
		SandboxPolicy:	policy,
		Call: VMCall{
			Runtime:	RuntimeCosmWasm,
			Action:		ActionExternalCall,
			GasLimit:	wasmconfig.DefaultSimulationGasLimit,
		},
		ZoneID:			zonestypes.ZoneIDContract,
		Backend:		RouterBackendWASMAdapter,
		DispatchMode:		RouterDispatchModeCrossZone,
		GasMeter:		RouterGasMeter{Class: RouterGasClassStandard, Limit: wasmconfig.DefaultSimulationGasLimit, Reserved: 1_000},
		EmittedMessages:	[]AVMAsyncMessage{msg},
	})
	require.NoError(t, err)
	require.NoError(t, call.Validate())
	require.Equal(t, ComputeAVMWASMContractRouteCallHash(call), call.RouteCallHash)

	directCrossZone := call
	directCrossZone.DirectCrossZoneCall = true
	directCrossZone.RouteCallHash = ComputeAVMWASMContractRouteCallHash(directCrossZone)
	require.ErrorContains(t, directCrossZone.Validate(), "async messages")

	network := call
	network.NetworkAccessAttempt = true
	network.RouteCallHash = ComputeAVMWASMContractRouteCallHash(network)
	require.ErrorContains(t, network.Validate(), "external network")

	queuedOnly := call
	queuedOnly.DispatchMode = RouterDispatchModeQueued
	queuedOnly.RouteCallHash = ComputeAVMWASMContractRouteCallHash(queuedOnly)
	require.ErrorContains(t, queuedOnly.Validate(), "cross-zone async dispatch")
}

func TestAVMWASMContractRouteRejectsWrongBackendAndInvalidStoreAdapter(t *testing.T) {
	policy := testAVMWASMSandboxPolicy(t)
	wrongBackend := AVMWASMContractRouteCall{
		SandboxPolicy:	policy,
		Call: VMCall{
			Runtime:	RuntimeCosmWasm,
			Action:		ActionExternalCall,
			GasLimit:	1,
		},
		ZoneID:		zonestypes.ZoneIDContract,
		Backend:	RouterBackendAVMActor,
		DispatchMode:	RouterDispatchModeQueued,
		GasMeter:	RouterGasMeter{Class: RouterGasClassStandard, Limit: 10, Reserved: 1},
	}
	wrongBackend.RouteCallHash = ComputeAVMWASMContractRouteCallHash(wrongBackend)
	require.ErrorContains(t, wrongBackend.Validate(), "WASM adapter backend")

	badAdapter := policy
	badAdapter.StoreAdapter.KeyPrefix = ContractZoneKVPrefix(zonestypes.ZoneIDApplication)
	badAdapter.StoreAdapter.AdapterHash = ComputeAVMWASMStoreV2KVAdapterHash(badAdapter.StoreAdapter)
	badAdapter.SandboxPolicyHash = ComputeAVMWASMSandboxPolicyHash(badAdapter)
	require.ErrorContains(t, badAdapter.Validate(), "Store v2 adapter prefix")
}

func testAVMWASMSandboxPolicy(t *testing.T) AVMWASMSandboxPolicy {
	t.Helper()
	runtime := wasmconfig.DefaultPolicy()
	runtime.Enabled = true
	runtime.GovernanceAuthority = governance
	table, err := DefaultAVMWASMGasConversionTable()
	require.NoError(t, err)
	adapter, err := NewAVMWASMStoreV2KVAdapter(AVMWASMStoreV2KVAdapter{
		ZoneID:		zonestypes.ZoneIDContract,
		StoreKey:	DefaultAVMStoreKey,
		KeyPrefix:	ContractZoneKVPrefix(zonestypes.ZoneIDContract),
		MaxKeyBytes:	DefaultMaxStorageKeyBytes,
		MaxValueBytes:	DefaultMaxStorageValueBytes,
	})
	require.NoError(t, err)
	policy, err := NewAVMWASMSandboxPolicy(AVMWASMSandboxPolicy{
		Enabled:		true,
		Optional:		true,
		RuntimePolicy:		runtime,
		MaxMemoryPages:		32,
		HostFunctions:		DefaultAVMWASMHostFunctions(),
		GasConversion:		table,
		StoreAdapter:		adapter,
		ExternalNetwork:	false,
		CrossZoneAsyncOnly:	true,
	})
	require.NoError(t, err)
	return policy
}

func testAVMWASMCrossZoneMessage(t *testing.T) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage("wasm-contract", zonestypes.ZoneIDContract, "app-workflow", zonestypes.ZoneIDApplication, 29, 12)
	msg.PayloadType = "application.workflow"
	msg.RouteHintOptional = "wasm.cross-zone"
	msg.ValueNAET = 0
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}
