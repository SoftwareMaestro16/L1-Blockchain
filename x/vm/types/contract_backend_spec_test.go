package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMContractBackendRegistrySupportsNativeWASMAndActorBackends(t *testing.T) {
	registry, err := DefaultAVMContractBackendRegistry(DefaultRuntimePolicy())
	require.NoError(t, err)
	require.NoError(t, registry.Validate())
	require.Equal(t, ComputeAVMContractBackendRegistryHash(registry), registry.RegistryHash)

	byKind := make(map[AVMContractBackendKind]AVMContractBackendDescriptor)
	for _, backend := range registry.Backends {
		byKind[backend.Kind] = backend
	}
	require.Equal(t, RouterBackendNativeModule, byKind[AVMContractBackendNativeModule].RouterBackend)
	require.True(t, byKind[AVMContractBackendNativeModule].KeeperBacked)
	require.True(t, byKind[AVMContractBackendNativeModule].MsgServerCompatible)
	require.True(t, byKind[AVMContractBackendNativeModule].Deterministic)
	require.Equal(t, RuntimeAVM, byKind[AVMContractBackendActorContract].Runtime)
	require.Equal(t, RuntimeCosmWasm, byKind[AVMContractBackendWASMContract].Runtime)
	require.True(t, byKind[AVMContractBackendWASMContract].Optional)
	require.False(t, byKind[AVMContractBackendWASMContract].Enabled)

	mutated := registry
	mutated.Backends = append([]AVMContractBackendDescriptor(nil), registry.Backends...)
	mutated.Backends[0].Deterministic = false
	mutated.Backends[0].BackendHash = ComputeAVMContractBackendHash(mutated.Backends[0])
	mutated.RegistryHash = ComputeAVMContractBackendRegistryHash(mutated)
	require.ErrorContains(t, mutated.Validate(), "deterministic")
}

func TestAVMNativeModuleRouteRequiresDescriptorReceiptAndZoneIsolation(t *testing.T) {
	descriptor := testAVMNativeModuleDescriptor(t)
	receipt := testAVMNativeModuleReceipt(t, descriptor.ZoneID, AVMReceiptStatusExecuted)
	call, err := NewAVMNativeModuleRouteCall(AVMNativeModuleRouteCall{
		Descriptor:		descriptor,
		RouteKey:		"financial.bank.send",
		Method:			"MsgSend",
		ZoneID:			descriptor.ZoneID,
		Lane:			RouterLaneSync,
		Backend:		RouterBackendNativeModule,
		DispatchMode:		RouterDispatchModeDirect,
		ReceiptPolicy:		RouterReceiptCommit,
		GasMeter:		RouterGasMeter{Class: RouterGasClassStandard, Limit: 100, Reserved: 80},
		Receipt:		receipt,
		CalledThroughAVM:	true,
		UsesInterfaceSystem:	true,
		StateWriteZones:	[]zonestypes.ZoneID{descriptor.ZoneID},
	})
	require.NoError(t, err)
	require.NoError(t, call.Validate())
	require.Equal(t, ComputeAVMNativeModuleRouteCallHash(call), call.CallHash)

	noReceipt := call
	noReceipt.CalledThroughAVM = true
	noReceipt.Receipt = AVMExecutionReceipt{}
	noReceipt.CallHash = ComputeAVMNativeModuleRouteCallHash(noReceipt)
	require.ErrorContains(t, noReceipt.Validate(), "receipt")

	crossZoneWrite := call
	crossZoneWrite.StateWriteZones = []zonestypes.ZoneID{zonestypes.ZoneIDFinancial, zonestypes.ZoneIDContract}
	crossZoneWrite.CallHash = ComputeAVMNativeModuleRouteCallHash(crossZoneWrite)
	require.ErrorContains(t, crossZoneWrite.Validate(), "other zones")
}

func TestAVMNativeModuleRejectsNonNativeBackendMissingInterfaceAndUnknownMethod(t *testing.T) {
	descriptor := testAVMNativeModuleDescriptor(t)
	receipt := testAVMNativeModuleReceipt(t, descriptor.ZoneID, AVMReceiptStatusExecuted)
	valid := AVMNativeModuleRouteCall{
		Descriptor:		descriptor,
		RouteKey:		"financial.bank.send",
		Method:			"MsgSend",
		ZoneID:			descriptor.ZoneID,
		Lane:			RouterLaneSync,
		Backend:		RouterBackendNativeModule,
		DispatchMode:		RouterDispatchModeDirect,
		ReceiptPolicy:		RouterReceiptCommit,
		GasMeter:		RouterGasMeter{Class: RouterGasClassStandard, Limit: 100, Reserved: 80},
		Receipt:		receipt,
		CalledThroughAVM:	true,
		UsesInterfaceSystem:	true,
		StateWriteZones:	[]zonestypes.ZoneID{descriptor.ZoneID},
	}

	wrongBackend := valid
	wrongBackend.Backend = RouterBackendAVMActor
	wrongBackend.CallHash = ComputeAVMNativeModuleRouteCallHash(wrongBackend)
	require.ErrorContains(t, wrongBackend.Validate(), "native backend")

	unknownMethod := valid
	unknownMethod.Method = "MsgBurn"
	unknownMethod.CallHash = ComputeAVMNativeModuleRouteCallHash(unknownMethod)
	require.ErrorContains(t, unknownMethod.Validate(), "not exposed")

	missingInterface := descriptor
	missingInterface.ServiceInterfaceHash = ""
	missingInterface.DescriptorHash = ComputeAVMNativeModuleDescriptorHash(missingInterface)
	require.ErrorContains(t, missingInterface.Validate(), "interface hash")
}

func testAVMNativeModuleDescriptor(t *testing.T) AVMNativeModuleDescriptor {
	t.Helper()
	descriptor, err := NewAVMNativeModuleDescriptor(AVMNativeModuleDescriptor{
		ModuleName:		"bank",
		ZoneID:			zonestypes.ZoneIDFinancial,
		KeeperService:		"x.bank.Keeper",
		MsgServerService:	"cosmos.bank.v1.MsgServer",
		ServiceInterfaceHash:	engineHash("bank-interface"),
		AllowedMessageTypes:	[]string{"financial.transfer"},
		AllowedMethods:		[]string{"MsgSend", "MsgMultiSend"},
	})
	require.NoError(t, err)
	return descriptor
}

func testAVMNativeModuleReceipt(t *testing.T, zoneID zonestypes.ZoneID, status AVMReceiptStatus) AVMExecutionReceipt {
	t.Helper()
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		engineHash("native-module-message"),
		ZoneID:			zoneID,
		Executor:		"native-bank-msgserver",
		Status:			status,
		GasUsed:		12,
		StorageWritten:		1,
		EventsHash:		engineHash("native-events"),
		OutputMessagesRoot:	engineHash("native-output"),
		CreatedHeight:		21,
	})
	require.NoError(t, err)
	return receipt
}
