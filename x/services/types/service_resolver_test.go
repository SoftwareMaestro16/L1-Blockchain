package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestResolveServiceFromOnChainRegistryReturnsDiscoveryOutput(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	state := testResolverRegistryState(t, descriptor, nil, 20)

	out, err := ResolveService(ServiceResolverInput{
		ServiceName:		descriptor.Discovery.ServiceName,
		Registry:		state,
		ResolutionHeight:	25,
		RequireFreshProof:	true,
	})
	require.NoError(t, err)
	require.Equal(t, ServiceResolutionOnChainRegistry, out.Source)
	require.Equal(t, descriptor.ServiceID, out.ServiceID)
	require.Equal(t, descriptor.Execution.Endpoint, out.Endpoint)
	require.Equal(t, descriptor.Interface.InterfaceHash, out.InterfaceHash)
	require.Equal(t, descriptor.Interface, out.InterfaceDescriptor)
	require.Equal(t, descriptor.Verification.TrustModel, out.TrustModel)
	require.Equal(t, descriptor.Verification.Model, out.VerificationModel)
	require.Equal(t, uint64(90), out.ExpiryHeight)
	require.NotEmpty(t, out.PaymentModel)
	require.NotEmpty(t, out.ProofChain.RegistryProofHash)
	require.NotEmpty(t, out.ProofChain.ChainHash)
	require.NoError(t, out.Validate())
}

func TestResolveServiceFromIdentityBinding(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.Discovery.IdentityName = "portable.aet"
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	binding, err := coretypes.NewIdentityServiceBindingFromDescriptor(descriptor)
	require.NoError(t, err)
	state := testResolverRegistryState(t, descriptor, []coretypes.IdentityServiceBinding{binding}, 20)

	out, err := ResolveService(ServiceResolverInput{
		ServiceName:		"portable.aet",
		Registry:		state,
		ResolutionHeight:	25,
	})
	require.NoError(t, err)
	require.Equal(t, ServiceResolutionIdentityRecord, out.Source)
	require.Equal(t, descriptor.ServiceID, out.ServiceID)
	require.Equal(t, []string{binding.BindingHash}, out.ProofChain.IdentityBindingHashes)
	require.NotEmpty(t, out.ProofChain.RegistryProofHash)
	require.NoError(t, out.Validate())
}

func TestResolveServiceFromSignedCacheRequiresSignatureChain(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.Discovery.ServiceName = "cached-service"
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	state, err := coretypes.NewServiceRegistryState(nil, nil, nil, nil, nil, nil, 20)
	require.NoError(t, err)

	_, err = ResolveService(ServiceResolverInput{
		ServiceName:		"cached-service",
		Registry:		state,
		ResolutionHeight:	25,
		CachedSigned: []ServiceResolverSourceRecord{{
			ServiceName:	"cached-service",
			Descriptor:	descriptor,
		}},
	})
	require.ErrorContains(t, err, "signature hash")

	out, err := ResolveService(ServiceResolverInput{
		ServiceName:		"cached-service",
		Registry:		state,
		ResolutionHeight:	25,
		RequireFreshProof:	true,
		CachedSigned: []ServiceResolverSourceRecord{{
			ServiceName:		"cached-service",
			Descriptor:		descriptor,
			SignatureHash:		testInterfaceHash("cached/owner-signature"),
			TrustMetadata:		"owner-signed-cache-record",
			ExpiryHeight:		80,
			PaymentModel:		registryPaymentModelFromDescriptor(descriptor),
			VerificationModel:	descriptor.Verification.Model,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ServiceResolutionSignedCache, out.Source)
	require.Equal(t, []string{testInterfaceHash("cached/owner-signature")}, out.ProofChain.SignatureHashes)
	require.Equal(t, uint64(80), out.ExpiryHeight)
	require.NoError(t, out.Validate())
}

func TestResolveServiceFromDistributedMeshIncludesCommitmentChain(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	state := testResolverRegistryState(t, descriptor, nil, 20)

	iface, err := NewDistributedInterfaceDescriptor(DistributedInterfaceDescriptor{
		InterfaceHash:	descriptor.Interface.InterfaceHash,
		InterfaceName:	descriptor.Interface.InterfaceName,
		Version:	descriptor.Interface.Version,
		SchemaHash:	testInterfaceHash("mesh/schema"),
		MethodRoot:	testInterfaceHash("mesh/methods"),
		EventRoot:	testInterfaceHash("mesh/events"),
		ErrorRoot:	testInterfaceHash("mesh/errors"),
	})
	require.NoError(t, err)
	endpoint, err := NewDistributedServiceEndpoint(DistributedServiceEndpoint{
		ServiceID:	descriptor.ServiceID,
		EndpointID:	"mesh-primary",
		Kind:		DistributedEndpointHybrid,
		ZoneID:		string(descriptor.ZoneID),
		Target:		"https://mesh.aetra.local/portable",
		InterfaceHash:	descriptor.Interface.InterfaceHash,
		Priority:	1,
		Weight:		10,
		MetadataHash:	testInterfaceHash("mesh/endpoint"),
	})
	require.NoError(t, err)
	record, err := NewDistributedServiceRecord(DistributedServiceRecord{
		ServiceID:	descriptor.ServiceID,
		ServiceName:	"portable-mesh",
		Kind:		DistributedServiceHybrid,
		Owner:		coretypes.DefaultAuthority,
		ZoneID:		string(descriptor.ZoneID),
		InterfaceHash:	descriptor.Interface.InterfaceHash,
		EndpointRoot:	ComputeDistributedEndpointRoot([]DistributedServiceEndpoint{endpoint}),
		DescriptorHash:	coretypes.ComputeServiceDescriptorHash(descriptor),
		MetadataHash:	testInterfaceHash("mesh/metadata"),
		CreatedHeight:	20,
		UpdatedHeight:	20,
		ExpiryHeight:	90,
		Discoverable:	true,
	})
	require.NoError(t, err)
	commitment, err := NewDistributedExecutionCommitment(DistributedExecutionCommitment{
		ServiceID:		descriptor.ServiceID,
		EndpointID:		endpoint.EndpointID,
		Kind:			DistributedCommitmentProof,
		ProofHash:		testInterfaceHash("mesh/proof"),
		ResultHash:		testInterfaceHash("mesh/result"),
		CommittedHeight:	30,
	})
	require.NoError(t, err)
	distributed, err := BuildDistributedDiscoveryState(
		[]DistributedServiceRecord{record},
		[]DistributedServiceEndpoint{endpoint},
		[]DistributedInterfaceDescriptor{iface},
		[]DistributedExecutionCommitment{commitment},
		30,
	)
	require.NoError(t, err)

	out, err := ResolveService(ServiceResolverInput{
		ServiceName:		"portable-mesh",
		Registry:		state,
		Distributed:		distributed,
		ResolutionHeight:	31,
	})
	require.NoError(t, err)
	require.Equal(t, ServiceResolutionDistributedMesh, out.Source)
	require.Equal(t, endpoint.Target, out.Endpoint)
	require.Equal(t, []string{record.RecordHash}, out.ProofChain.DistributedRecordHashes)
	require.Equal(t, []string{endpoint.CommitmentHash}, out.ProofChain.DistributedEndpointHashes)
	require.Equal(t, []string{iface.DescriptorHash}, out.ProofChain.DistributedInterfaceHashes)
	require.Equal(t, []string{commitment.CommitmentHash}, out.ProofChain.DistributedCommitmentHashes)
	require.NoError(t, out.Validate())
}

func testResolverRegistryState(t *testing.T, descriptor coretypes.ServiceDescriptor, bindings []coretypes.IdentityServiceBinding, height uint64) coretypes.ServiceRegistryState {
	t.Helper()
	anchor, err := coretypes.NewServiceAnchorFromDescriptor(descriptor)
	require.NoError(t, err)
	state, err := coretypes.NewServiceRegistryState(
		[]coretypes.ServiceDescriptor{descriptor},
		[]coretypes.ServiceAnchor{anchor},
		bindings,
		nil,
		nil,
		nil,
		height,
	)
	require.NoError(t, err)
	return state
}
