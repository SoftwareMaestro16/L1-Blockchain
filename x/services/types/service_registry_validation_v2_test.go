package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceRegistryV2ValidationRules(t *testing.T) {
	params := testServiceRegistryParamsV2(t, 25, []CanonicalServiceEndpointType{CanonicalEndpointZoneAware}, []string{ServiceStateDependencyZoneRoot})
	iface := testVersionedRegistryInterface(t, "payments", 1)
	descriptor := testStrictRegistryDescriptor(t, "svc-pay", iface.InterfaceHash, "owner-alpha", "financial", CanonicalEndpointZoneAware, "zone_root:financial", 20)

	require.NoError(t, ValidateServiceDescriptorAgainstParamsV2(descriptor, []DistributedInterfaceDescriptor{iface}, params, 10))

	badHash := iface
	badHash.InterfaceHash = testDistributedHash("tampered-interface")
	badHash.DescriptorHash = ComputeDistributedInterfaceDescriptorHash(badHash)
	err := ValidateServiceDescriptorAgainstParamsV2(descriptor, []DistributedInterfaceDescriptor{badHash}, params, 10)
	require.ErrorContains(t, err, "missing interface")

	badTTL := descriptor
	badTTL.TTLHeight = 36
	badTTL.DescriptorHash = ""
	badTTL, err = NewCanonicalServiceDescriptor(badTTL)
	require.NoError(t, err)
	err = ValidateServiceDescriptorAgainstParamsV2(badTTL, []DistributedInterfaceDescriptor{iface}, params, 10)
	require.ErrorContains(t, err, "ttl exceeds configured maximum")

	badEndpoint := descriptor
	badEndpoint.EndpointType = CanonicalEndpointAPI
	badEndpoint.DescriptorHash = ""
	badEndpoint, err = NewCanonicalServiceDescriptor(badEndpoint)
	require.NoError(t, err)
	err = ValidateServiceDescriptorAgainstParamsV2(badEndpoint, []DistributedInterfaceDescriptor{iface}, params, 10)
	require.ErrorContains(t, err, "not allowed by params")

	badDependency := descriptor
	badDependency.StateDependency = "module_root:bank"
	badDependency.DescriptorHash = ""
	badDependency, err = NewCanonicalServiceDescriptor(badDependency)
	require.NoError(t, err)
	err = ValidateServiceDescriptorAgainstParamsV2(badDependency, []DistributedInterfaceDescriptor{iface}, params, 10)
	require.ErrorContains(t, err, "not allowed by params")
}

func TestServiceRegistryV2StrictMessagesAndOwnerAuthorization(t *testing.T) {
	params := testServiceRegistryParamsV2(t, 100, []CanonicalServiceEndpointType{CanonicalEndpointZoneAware}, []string{ServiceStateDependencyZoneRoot})
	iface := testVersionedRegistryInterface(t, "automation", 1)
	descriptor := testStrictRegistryDescriptor(t, "svc-auto", iface.InterfaceHash, "owner-alpha", "apps", CanonicalEndpointZoneAware, "zone_root:apps", 40)

	state, err := BuildServiceRegistryStateV2(nil, []DistributedInterfaceDescriptor{iface}, nil, nil, 10)
	require.NoError(t, err)

	state, err = RegisterServiceInRegistryWithParamsV2(state, MsgRegisterServiceV2{
		Authority:	"owner-alpha",
		Descriptor:	descriptor,
	}, params, 11)
	require.NoError(t, err)
	require.NoError(t, ValidateServiceOwnerAuthorizationV2(state, descriptor.ServiceID, "owner-alpha"))
	require.ErrorContains(t, ValidateServiceOwnerAuthorizationV2(state, descriptor.ServiceID, "owner-beta"), "owner must authorize")

	updated := descriptor
	updated.Version = 2
	updated.MetadataHash = testDistributedHash("svc-auto/strict-metadata-v2")
	updated.DescriptorHash = ""
	updated, err = NewCanonicalServiceDescriptor(updated)
	require.NoError(t, err)
	state, err = UpdateServiceInRegistryWithParamsV2(state, MsgUpdateServiceV2{
		Authority:	"owner-alpha",
		Descriptor:	updated,
	}, params, 12)
	require.NoError(t, err)
	require.Equal(t, uint64(2), state.Descriptors[0].Version)

	updated.Owner = "owner-beta"
	updated.DescriptorHash = ""
	updated, err = NewCanonicalServiceDescriptor(updated)
	require.NoError(t, err)
	_, err = UpdateServiceInRegistryWithParamsV2(state, MsgUpdateServiceV2{
		Authority:	"owner-beta",
		Descriptor:	updated,
	}, params, 13)
	require.ErrorContains(t, err, "owner must authorize")
}

func TestServiceRegistryV2ExportImportVersionedInterfacesAndLookup(t *testing.T) {
	params := DefaultServiceRegistryParamsV2()
	ifaceV1 := testVersionedRegistryInterface(t, "contracts", 1)
	ifaceV2 := testVersionedRegistryInterface(t, "contracts", 2)
	serviceA := testStrictRegistryDescriptor(t, "svc-contract-a", ifaceV1.InterfaceHash, "owner-beta", "contract", CanonicalEndpointZoneAware, "zone_root:contract", 100)
	serviceB := testStrictRegistryDescriptor(t, "svc-contract-b", ifaceV2.InterfaceHash, "owner-beta", "contract", CanonicalEndpointHybrid, "module_root:wasm", 100)
	receipt := testRegistryReceipt(t, serviceA.ServiceID, "receipt-1")

	state, err := BuildServiceRegistryStateV2(
		[]CanonicalServiceDescriptor{serviceB, serviceA},
		[]DistributedInterfaceDescriptor{ifaceV2, ifaceV1},
		[]ServiceReceiptV2{receipt},
		nil,
		20,
	)
	require.NoError(t, err)
	require.NoError(t, ValidateServiceRegistryStateWithParamsV2(state, params, 20))

	versioned, err := BuildServiceVersionedInterfaceEntriesV2(state.Interfaces)
	require.NoError(t, err)
	require.Len(t, versioned, 2)
	versionedByVersion := map[uint64]ServiceVersionedInterfaceEntryV2{}
	for _, entry := range versioned {
		versionedByVersion[entry.Version] = entry
	}
	require.Equal(t, ifaceV1.InterfaceHash, versionedByVersion[1].InterfaceHash)
	require.Equal(t, ifaceV2.InterfaceHash, versionedByVersion[2].InterfaceHash)
	require.Contains(t, versionedByVersion[1].Key, "/versions/00000000000000000001")

	versionedRoot, err := ComputeServiceVersionedInterfaceRootV2(state.Interfaces)
	require.NoError(t, err)
	require.NotEmpty(t, versionedRoot)

	descriptorKey, err := ServiceDescriptorV2Key(serviceA.ServiceID)
	require.NoError(t, err)
	lookup, found, err := LookupServiceByRegistryKeyV2(state, descriptorKey)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, serviceA.ServiceID, lookup.ServiceID)

	ownerKey, err := ServiceOwnerIndexV2Key("owner-beta", serviceB.ServiceID)
	require.NoError(t, err)
	lookup, found, err = LookupServiceByRegistryKeyV2(state, ownerKey)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, serviceB.ServiceID, lookup.ServiceID)

	exported, err := ExportServiceRegistryV2(state, params)
	require.NoError(t, err)
	imported, err := ImportServiceRegistryV2(exported)
	require.NoError(t, err)
	require.Equal(t, state.StateRoot, imported.StateRoot)
	require.Equal(t, exported.Roots.StateRoot, imported.StateRoot)

	exported.Roots.StateRoot = testDistributedHash("wrong-root")
	err = exported.Validate()
	require.ErrorContains(t, err, "roots mismatch")
}

func TestServiceRegistryV2VersionedInterfaceRejectsHashNotCommittingToBytes(t *testing.T) {
	iface := testVersionedRegistryInterface(t, "storage", 1)
	iface.InterfaceHash = testDistributedHash("storage/manual-interface")
	iface.DescriptorHash = ComputeDistributedInterfaceDescriptorHash(iface)

	err := ValidateVersionedServiceInterfaceDescriptorV2(iface)
	require.ErrorContains(t, err, "interface hash must commit to descriptor bytes")
}

func testServiceRegistryParamsV2(t *testing.T, maxTTL uint64, endpoints []CanonicalServiceEndpointType, dependencies []string) ServiceRegistryParamsV2 {
	t.Helper()
	params, err := NewServiceRegistryParamsV2(ServiceRegistryParamsV2{
		MaxTTLDelta:			maxTTL,
		AllowedEndpointTypes:		endpoints,
		AllowedStateDependencyRoots:	dependencies,
		ProofHorizon:			128,
	})
	require.NoError(t, err)
	return params
}

func testVersionedRegistryInterface(t *testing.T, name string, version uint64) DistributedInterfaceDescriptor {
	t.Helper()
	iface, err := NewVersionedServiceInterfaceDescriptorV2(DistributedInterfaceDescriptor{
		InterfaceName:	name,
		Version:	version,
		SchemaHash:	testDistributedHash(name + "/" + string(rune('0'+version)) + "/schema"),
		MethodRoot:	testDistributedHash(name + "/" + string(rune('0'+version)) + "/methods"),
		EventRoot:	testDistributedHash(name + "/" + string(rune('0'+version)) + "/events"),
		ErrorRoot:	testDistributedHash(name + "/" + string(rune('0'+version)) + "/errors"),
	})
	require.NoError(t, err)
	return iface
}

func testStrictRegistryDescriptor(t *testing.T, serviceID, interfaceHash, owner, zoneID string, endpointType CanonicalServiceEndpointType, stateDependency string, ttl uint64) CanonicalServiceDescriptor {
	t.Helper()
	descriptor, err := NewCanonicalServiceDescriptor(CanonicalServiceDescriptor{
		ServiceID:		serviceID,
		EndpointType:		endpointType,
		InterfaceHash:		interfaceHash,
		SupportedMethods:	[]string{"call"},
		AuthModel:		"account",
		StateDependency:	stateDependency,
		Owner:			owner,
		ZoneID:			zoneID,
		Version:		1,
		EndpointURIHash:	testDistributedHash(serviceID + "/strict-endpoint"),
		MetadataHash:		testDistributedHash(serviceID + "/strict-metadata"),
		TTLHeight:		ttl,
		Status:			CanonicalServiceStatusActive,
		Capabilities:		[]string{"proofs"},
	})
	require.NoError(t, err)
	return descriptor
}
