package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultXServicesModuleBreakdownCoversSection151(t *testing.T) {
	breakdown, err := DefaultXServicesModuleBreakdown()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.Equal(t, ServiceModuleServices, breakdown.ModulePath)
	require.NotEmpty(t, breakdown.BreakdownHash)

	require.Contains(t, breakdown.StateObjects, XServicesStateDescriptor)
	require.Contains(t, breakdown.StateObjects, XServicesStateAnchor)
	require.Contains(t, breakdown.StateObjects, XServicesStateIdentityBinding)
	require.Contains(t, breakdown.StateObjects, XServicesStateStatus)
	require.Contains(t, breakdown.StateObjects, XServicesStateParams)

	require.Contains(t, breakdown.Messages, XServicesMsgRegisterService)
	require.Contains(t, breakdown.Messages, XServicesMsgUpdateService)
	require.Contains(t, breakdown.Messages, XServicesMsgRenewService)
	require.Contains(t, breakdown.Messages, XServicesMsgDisableService)
	require.Contains(t, breakdown.Messages, XServicesMsgTransferService)
	require.Contains(t, breakdown.Messages, XServicesMsgBindServiceIdentity)
	require.Contains(t, breakdown.Messages, XServicesMsgUnbindServiceIdentity)

	require.Contains(t, breakdown.Queries, XServicesQueryService)
	require.Contains(t, breakdown.Queries, XServicesQueryServiceByName)
	require.Contains(t, breakdown.Queries, XServicesQueryServicesByOwner)
	require.Contains(t, breakdown.Queries, XServicesQueryServicesByIdentity)
	require.Contains(t, breakdown.Queries, XServicesQueryServiceProof)

	require.Contains(t, breakdown.IntegrationPoints, XServicesIntegrationIdentity)
	require.Contains(t, breakdown.IntegrationPoints, XServicesIntegrationServiceInterface)
	require.Contains(t, breakdown.IntegrationPoints, XServicesIntegrationServiceCalls)
	require.Contains(t, breakdown.IntegrationPoints, XServicesIntegrationStoreV2ProofQuery)
}

func TestXServicesModuleBreakdownRejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultXServicesModuleBreakdown()
	require.NoError(t, err)
	breakdown.Messages = removeXServicesMessageForTest(breakdown.Messages, XServicesMsgTransferService)
	breakdown.BreakdownHash = ComputeXServicesModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "message")

	breakdown, err = DefaultXServicesModuleBreakdown()
	require.NoError(t, err)
	breakdown.FailureModes = breakdown.FailureModes[1:]
	breakdown.BreakdownHash = ComputeXServicesModuleBreakdownHash(breakdown)
	require.ErrorContains(t, breakdown.Validate(), "failure")
}

func TestXServicesFailureModesUseRegistryGuards(t *testing.T) {
	state, descriptor, _ := testXServicesRegistryState(t)
	_, err := RegisterServiceInRegistryV2(state, MsgRegisterServiceV2{
		Authority:	descriptor.Owner,
		Descriptor:	descriptor,
	}, 20)
	require.ErrorContains(t, err, "already exists")

	updated := descriptor
	updated.Version = 2
	updated.DescriptorHash = ""
	updated, err = NewCanonicalServiceDescriptor(updated)
	require.NoError(t, err)
	_, err = UpdateServiceInRegistryV2(state, MsgUpdateServiceV2{
		Authority:	"attacker",
		Descriptor:	updated,
	}, 21)
	require.ErrorContains(t, err, "must own service")
}

func TestXServicesDescriptorUsableForCallRejectsExpiredAndInterfaceMismatch(t *testing.T) {
	_, descriptor, iface := testXServicesRegistryState(t)
	check, err := ValidateXServicesDescriptorUsableForCall(descriptor, iface.InterfaceHash, 25)
	require.NoError(t, err)
	require.NoError(t, check.Validate())
	require.Equal(t, descriptor.ServiceID, check.ServiceID)

	_, err = ValidateXServicesDescriptorUsableForCall(descriptor, iface.InterfaceHash, descriptor.TTLHeight+1)
	require.ErrorContains(t, err, "expired descriptor")

	_, err = ValidateXServicesDescriptorUsableForCall(descriptor, testInterfaceHash("x-services/wrong-interface"), 25)
	require.ErrorContains(t, err, "interface hash mismatch")

	disabled := descriptor
	disabled.Status = CanonicalServiceStatusDisabled
	disabled.DescriptorHash = ""
	disabled, err = NewCanonicalServiceDescriptor(disabled)
	require.NoError(t, err)
	_, err = ValidateXServicesDescriptorUsableForCall(disabled, iface.InterfaceHash, 25)
	require.ErrorContains(t, err, "not active")
}

func TestXServicesIdentityBindingFreshnessDetectsDomainTransfer(t *testing.T) {
	_, descriptor, _ := testXServicesRegistryState(t)
	binding, err := NewServiceIdentityBindingV2(ServiceIdentityBindingV2{
		ServiceID:	descriptor.ServiceID,
		IdentityName:	"dex.aet",
		Owner:		descriptor.Owner,
		BoundHeight:	30,
	})
	require.NoError(t, err)

	check, err := ValidateXServicesIdentityBindingFreshness(binding, descriptor.Owner, 40)
	require.NoError(t, err)
	require.NoError(t, check.Validate())

	_, err = ValidateXServicesIdentityBindingFreshness(binding, "new-owner", 40)
	require.ErrorContains(t, err, "stale after domain transfer")

	_, err = ValidateXServicesIdentityBindingFreshness(binding, descriptor.Owner, 20)
	require.ErrorContains(t, err, "cannot precede")
}

func TestXServicesStoreV2ProofQueryIntegration(t *testing.T) {
	state, descriptor, _ := testXServicesRegistryState(t)
	key, err := ServiceDescriptorV2Key(descriptor.ServiceID)
	require.NoError(t, err)
	require.True(t, IsServiceStoreKey(key))

	proof, err := QueryServiceProofFromRegistryV2(state, QueryServiceProofV2{Key: key})
	require.NoError(t, err)
	require.Equal(t, "QueryServiceProof", proof.Query)
	require.Equal(t, key, proof.Key)
	require.Equal(t, state.StateRoot, proof.Root)
	require.Equal(t, state.Height, proof.Height)
	require.NotEmpty(t, proof.ValueHash)
	require.NotEmpty(t, proof.ProofHashes)
}

func testXServicesRegistryState(t *testing.T) (ServiceRegistryStateV2, CanonicalServiceDescriptor, DistributedInterfaceDescriptor) {
	t.Helper()
	iface, err := NewDistributedInterfaceDescriptor(DistributedInterfaceDescriptor{
		InterfaceHash:	testInterfaceHash("x-services/iface"),
		InterfaceName:	"dex.v1",
		Version:	1,
		SchemaHash:	testInterfaceHash("x-services/schema"),
		MethodRoot:	testInterfaceHash("x-services/methods"),
		EventRoot:	testInterfaceHash("x-services/events"),
		ErrorRoot:	testInterfaceHash("x-services/errors"),
	})
	require.NoError(t, err)
	descriptor, err := NewCanonicalServiceDescriptor(CanonicalServiceDescriptor{
		ServiceID:		"svc.dex",
		EndpointType:		CanonicalEndpointApplication,
		InterfaceHash:		iface.InterfaceHash,
		SupportedMethods:	[]string{"swap"},
		AuthModel:		"owner",
		StateDependency:	"service_root",
		Owner:			"owner.dex",
		ZoneID:			"aetra",
		Version:		1,
		EndpointURIHash:	testInterfaceHash("x-services/endpoint"),
		MetadataHash:		testInterfaceHash("x-services/metadata"),
		TTLHeight:		100,
		Status:			CanonicalServiceStatusActive,
		Capabilities:		[]string{"dex"},
	})
	require.NoError(t, err)
	state, err := BuildServiceRegistryStateV2([]CanonicalServiceDescriptor{descriptor}, []DistributedInterfaceDescriptor{iface}, nil, nil, 10)
	require.NoError(t, err)
	return state, descriptor, iface
}

func removeXServicesMessageForTest(messages []XServicesMessageName, target XServicesMessageName) []XServicesMessageName {
	out := make([]XServicesMessageName, 0, len(messages))
	for _, message := range messages {
		if message != target {
			out = append(out, message)
		}
	}
	return out
}
