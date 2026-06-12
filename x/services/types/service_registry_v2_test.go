package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceRegistryV2StateKeysMatchSpec(t *testing.T) {
	methodHash := testDistributedHash("method")

	descriptorKey, err := ServiceDescriptorV2Key("svc-alpha")
	require.NoError(t, err)
	require.Equal(t, "services/descriptors/svc-alpha", descriptorKey)

	interfaceKey, err := ServiceInterfaceV2Key(testDistributedHash("interface"))
	require.NoError(t, err)
	require.Equal(t, "services/interfaces/"+testDistributedHash("interface"), interfaceKey)

	ownerKey, err := ServiceOwnerIndexV2Key("owner-alpha", "svc-alpha")
	require.NoError(t, err)
	require.Equal(t, "services/owner_index/owner-alpha/svc-alpha", ownerKey)

	zoneKey, err := ServiceZoneIndexV2Key("financial", "svc-alpha")
	require.NoError(t, err)
	require.Equal(t, "services/zone_index/financial/svc-alpha", zoneKey)

	methodKey, err := ServiceMethodIndexV2Key(methodHash, "svc-alpha")
	require.NoError(t, err)
	require.Equal(t, "services/method_index/"+methodHash+"/svc-alpha", methodKey)

	receiptKey, err := ServiceReceiptV2Key("svc-alpha", "receipt-1")
	require.NoError(t, err)
	require.Equal(t, "services/receipts/svc-alpha/receipt-1", receiptKey)
}

func TestServiceRegistryV2BuildsDeterministicIndexesAndRoot(t *testing.T) {
	ifaceA := testRegistryInterface(t, "payments")
	ifaceB := testRegistryInterface(t, "identity")
	serviceA := testRegistryDescriptor(t, "svc-pay", ifaceA.InterfaceHash, "owner-alpha", "financial", []string{"settle", "quote"})
	serviceB := testRegistryDescriptor(t, "svc-id", ifaceB.InterfaceHash, "owner-alpha", "identity", []string{"resolve"})
	receipt := testRegistryReceipt(t, serviceA.ServiceID, "receipt-1")

	state, err := BuildServiceRegistryStateV2(
		[]CanonicalServiceDescriptor{serviceA, serviceB},
		[]DistributedInterfaceDescriptor{ifaceA, ifaceB},
		[]ServiceReceiptV2{receipt},
		nil,
		20,
	)
	require.NoError(t, err)
	require.Equal(t, []string{"svc-id", "svc-pay"}, []string{state.Descriptors[0].ServiceID, state.Descriptors[1].ServiceID})
	require.Len(t, state.OwnerIndex, 2)
	require.Len(t, state.ZoneIndex, 2)
	require.Len(t, state.MethodIndex, 3)
	require.NotEmpty(t, state.StateRoot)

	reordered, err := BuildServiceRegistryStateV2(
		[]CanonicalServiceDescriptor{serviceB, serviceA},
		[]DistributedInterfaceDescriptor{ifaceB, ifaceA},
		[]ServiceReceiptV2{receipt},
		nil,
		20,
	)
	require.NoError(t, err)
	require.Equal(t, state.StateRoot, reordered.StateRoot)

	ownerServices, err := QueryServicesByOwnerFromRegistryV2(state, QueryServicesByOwnerV2{Owner: "owner-alpha"})
	require.NoError(t, err)
	require.Equal(t, []string{"svc-id", "svc-pay"}, []string{ownerServices[0].ServiceID, ownerServices[1].ServiceID})

	zoneServices, err := QueryServicesByZoneFromRegistryV2(state, QueryServicesByZoneV2{ZoneID: "financial"})
	require.NoError(t, err)
	require.Equal(t, []string{"svc-pay"}, []string{zoneServices[0].ServiceID})

	methodHash := ComputeServiceRegistryMethodHashV2(ifaceA.InterfaceHash, "settle")
	methodServices, err := QueryServiceByMethodFromRegistryV2(state, QueryServiceByMethodV2{MethodHash: methodHash})
	require.NoError(t, err)
	require.Equal(t, []string{"svc-pay"}, []string{methodServices[0].ServiceID})
}

func TestServiceRegistryV2MessagesMutateState(t *testing.T) {
	iface := testRegistryInterface(t, "automation")
	service := testRegistryDescriptor(t, "svc-auto", iface.InterfaceHash, "owner-alpha", "apps", []string{"execute"})

	state, err := BuildServiceRegistryStateV2(nil, nil, nil, nil, 1)
	require.NoError(t, err)

	state, err = RegisterInterfaceInRegistryV2(state, MsgRegisterInterfaceV2{
		Authority:	"owner-alpha",
		Descriptor:	iface,
	}, 2)
	require.NoError(t, err)
	require.Len(t, state.Interfaces, 1)

	state, err = RegisterServiceInRegistryV2(state, MsgRegisterServiceV2{
		Authority:	"owner-alpha",
		Descriptor:	service,
	}, 3)
	require.NoError(t, err)
	require.Len(t, state.Descriptors, 1)
	require.Len(t, state.OwnerIndex, 1)

	updated := service
	updated.Version = 2
	updated.MetadataHash = testDistributedHash("svc-auto/metadata-v2")
	updated.DescriptorHash = ""
	updated, err = NewCanonicalServiceDescriptor(updated)
	require.NoError(t, err)
	state, err = UpdateServiceInRegistryV2(state, MsgUpdateServiceV2{
		Authority:	"owner-alpha",
		Descriptor:	updated,
	}, 4)
	require.NoError(t, err)
	require.Equal(t, uint64(2), state.Descriptors[0].Version)

	state, err = BindServiceToIdentityInRegistryV2(state, MsgBindServiceToIdentityV2{
		Authority:	"owner-alpha",
		ServiceID:	service.ServiceID,
		IdentityName:	"svc-auto.aet",
	}, 5)
	require.NoError(t, err)
	require.Len(t, state.IdentityBindings, 1)

	state, err = DisableServiceInRegistryV2(state, MsgDisableServiceV2{
		Authority:	"owner-alpha",
		ServiceID:	service.ServiceID,
		Height:		6,
	}, 6)
	require.NoError(t, err)
	require.Equal(t, CanonicalServiceStatusDisabled, state.Descriptors[0].Status)

	state, err = UnbindServiceFromIdentityInRegistryV2(state, MsgUnbindServiceFromIdentityV2{
		Authority:	"owner-alpha",
		ServiceID:	service.ServiceID,
		IdentityName:	"svc-auto.aet",
	}, 7)
	require.NoError(t, err)
	require.Empty(t, state.IdentityBindings)
}

func TestServiceRegistryV2QueriesAndProof(t *testing.T) {
	iface := testRegistryInterface(t, "contracts")
	service := testRegistryDescriptor(t, "svc-contract", iface.InterfaceHash, "owner-beta", "contract", []string{"call"})
	receipt := testRegistryReceipt(t, service.ServiceID, "receipt-1")

	state, err := BuildServiceRegistryStateV2(
		[]CanonicalServiceDescriptor{service},
		[]DistributedInterfaceDescriptor{iface},
		[]ServiceReceiptV2{receipt},
		nil,
		30,
	)
	require.NoError(t, err)

	gotService, err := QueryServiceFromRegistryV2(state, QueryServiceV2{ServiceID: service.ServiceID})
	require.NoError(t, err)
	require.Equal(t, service.DescriptorHash, gotService.DescriptorHash)

	gotInterface, err := QueryInterfaceFromRegistryV2(state, QueryInterfaceV2{InterfaceHash: iface.InterfaceHash})
	require.NoError(t, err)
	require.Equal(t, iface.DescriptorHash, gotInterface.DescriptorHash)

	roots, err := QueryServiceRootFromRegistryV2(state, QueryServiceRootV2{})
	require.NoError(t, err)
	require.Equal(t, state.StateRoot, roots.StateRoot)
	require.NotEmpty(t, roots.DescriptorRoot)
	require.NotEmpty(t, roots.ReceiptRoot)

	key, err := ServiceDescriptorV2Key(service.ServiceID)
	require.NoError(t, err)
	proof, err := QueryServiceProofFromRegistryV2(state, QueryServiceProofV2{Key: key})
	require.NoError(t, err)
	require.Equal(t, key, proof.Key)
	require.Equal(t, service.DescriptorHash, proof.ValueHash)
	require.Equal(t, state.StateRoot, proof.Root)
	require.Equal(t, uint64(30), proof.Height)
	require.NotEmpty(t, proof.ProofHashes)
}

func TestServiceRegistryV2RejectsTamperedIndexes(t *testing.T) {
	iface := testRegistryInterface(t, "storage")
	service := testRegistryDescriptor(t, "svc-storage", iface.InterfaceHash, "owner-gamma", "storage", []string{"pin"})
	state, err := BuildServiceRegistryStateV2([]CanonicalServiceDescriptor{service}, []DistributedInterfaceDescriptor{iface}, nil, nil, 40)
	require.NoError(t, err)

	state.OwnerIndex[0].Value = "svc-other"
	state.OwnerIndex[0].EntryHash = ComputeServiceRegistryIndexEntryV2Hash(state.OwnerIndex[0])
	err = state.Validate()
	require.ErrorContains(t, err, "owner index does not match descriptors")
}

func testRegistryInterface(t *testing.T, name string) DistributedInterfaceDescriptor {
	t.Helper()
	return testDistributedInterface(t, name)
}

func testRegistryDescriptor(t *testing.T, serviceID, interfaceHash, owner, zoneID string, methods []string) CanonicalServiceDescriptor {
	t.Helper()
	descriptor, err := NewCanonicalServiceDescriptor(CanonicalServiceDescriptor{
		ServiceID:		serviceID,
		EndpointType:		CanonicalEndpointZoneAware,
		InterfaceHash:		interfaceHash,
		SupportedMethods:	methods,
		AuthModel:		"account",
		StateDependency:	"committed",
		Owner:			owner,
		ZoneID:			zoneID,
		Version:		1,
		EndpointURIHash:	testDistributedHash(serviceID + "/endpoint"),
		MetadataHash:		testDistributedHash(serviceID + "/metadata"),
		TTLHeight:		100,
		Status:			CanonicalServiceStatusActive,
		Capabilities:		[]string{"proofs"},
	})
	require.NoError(t, err)
	return descriptor
}

func testRegistryReceipt(t *testing.T, serviceID, receiptID string) ServiceReceiptV2 {
	t.Helper()
	receipt, err := NewServiceReceiptV2(ServiceReceiptV2{
		ServiceID:	serviceID,
		ReceiptID:	receiptID,
		CallID:		"call-1",
		Status:		"committed",
		ResultHash:	testDistributedHash(serviceID + "/" + receiptID + "/result"),
		Height:		12,
	})
	require.NoError(t, err)
	return receipt
}
