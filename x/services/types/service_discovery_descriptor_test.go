package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	identitytypes "github.com/sovereign-l1/l1/x/identity/types"
)

func TestServiceDiscoveryDescriptorV1ProjectsResolutionOutput(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	state := testResolverRegistryState(t, descriptor, nil, 20)
	resolution, err := ResolveService(ServiceResolverInput{
		ServiceName:		descriptor.Discovery.ServiceName,
		Registry:		state,
		ResolutionHeight:	25,
	})
	require.NoError(t, err)

	discovery, err := ProjectServiceDiscoveryDescriptorV1(resolution)
	require.NoError(t, err)
	require.Equal(t, descriptor.ServiceID, discovery.ServiceID)
	require.Equal(t, descriptor.Discovery.ServiceName, discovery.ServiceName)
	require.Equal(t, descriptor.Owner, discovery.Owner)
	require.Equal(t, descriptor.ServiceType, discovery.ServiceType)
	require.Equal(t, descriptor.Execution.Endpoint, discovery.Endpoint)
	require.Equal(t, descriptor.Interface.InterfaceHash, discovery.InterfaceHash)
	require.Equal(t, descriptor.Verification.TrustModel, discovery.TrustModel)
	require.Equal(t, descriptor.Verification.Model, discovery.VerificationModel)
	require.Equal(t, descriptor.Status, discovery.Status)
	require.NotEmpty(t, discovery.PaymentModel)
	require.NotEmpty(t, discovery.ProofOptional)
	require.NotEmpty(t, discovery.DescriptorHash)
	require.NoError(t, discovery.Validate())
}

func TestAETServiceBindingRequiresIdentityAndServiceProofs(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.Discovery.IdentityName = "portable.aet"
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	registry := testResolverRegistryState(t, descriptor, nil, 20)
	serviceProof, err := registry.QueryServiceProof(QueryServiceProof{ServiceID: descriptor.ServiceID})
	require.NoError(t, err)
	require.True(t, serviceProof.Found)

	discovery, err := ProjectServiceDiscoveryDescriptorFromCore(descriptor, "portable.aet", serviceProof.Proof.ProofHash, "")
	require.NoError(t, err)
	binding, err := NewAETServiceBindingFromDescriptor("portable.aet", discovery)
	require.NoError(t, err)

	identityState := testAETIdentityState(t, "portable", descriptor.Execution.Endpoint)
	root, err := identitytypes.IdentityStateRoot(identityState)
	require.NoError(t, err)
	identityProof, err := identitytypes.BuildIdentityResolutionProof(identityState, "portable.aet", 13)
	require.NoError(t, err)

	proof, err := NewAETServiceBindingProof(AETServiceBindingProof{
		Binding:		binding,
		Descriptor:		discovery,
		RegistryDescriptor:	descriptor,
		IdentityTrustedRoot:	root,
		IdentityProof:		identityProof,
		ServiceRegistryProof:	serviceProof.Proof,
		Height:			13,
	})
	require.NoError(t, err)

	verification, err := VerifyAETServiceBinding(proof)
	require.NoError(t, err)
	require.Equal(t, "portable.aet", verification.IdentityName)
	require.Equal(t, descriptor.ServiceID, verification.ServiceID)
	require.Equal(t, descriptor.Interface.InterfaceHash, verification.InterfaceHash)
	require.Equal(t, descriptor.Execution.Endpoint, verification.Endpoint)
	require.Equal(t, binding.BindingHash, verification.BindingHash)
	require.Equal(t, discovery.DescriptorHash, verification.DescriptorHash)
	require.NoError(t, verification.Validate())

	tamperedEndpoint := proof
	tamperedEndpoint.Binding.Endpoint = "https://wrong.aetra.local/v1"
	tamperedEndpoint.Binding.BindingHash = ""
	tamperedEndpoint.Binding, err = NewAETServiceBinding(tamperedEndpoint.Binding)
	require.NoError(t, err)
	tamperedEndpoint.ProofHash = ""
	_, err = NewAETServiceBindingProof(tamperedEndpoint)
	require.ErrorContains(t, err, "descriptor mismatch")

	tamperedProof := proof
	tamperedProof.IdentityTrustedRoot = testInterfaceHash("wrong/identity-root")
	tamperedProof.ProofHash = ""
	tamperedProof, err = NewAETServiceBindingProof(tamperedProof)
	require.NoError(t, err)
	_, err = VerifyAETServiceBinding(tamperedProof)
	require.ErrorContains(t, err, "identity proof failed")
}

func TestAETServiceBindingsSupportMultipleServices(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	registry := testResolverRegistryState(t, descriptor, nil, 20)
	serviceProof, err := registry.QueryServiceProof(QueryServiceProof{ServiceID: descriptor.ServiceID})
	require.NoError(t, err)
	discovery, err := ProjectServiceDiscoveryDescriptorFromCore(descriptor, "portable.aet", serviceProof.Proof.ProofHash, "")
	require.NoError(t, err)
	binding, err := NewAETServiceBindingFromDescriptor("portable.aet", discovery)
	require.NoError(t, err)

	identityState := testAETIdentityState(t, "portable", descriptor.Execution.Endpoint)
	root, err := identitytypes.IdentityStateRoot(identityState)
	require.NoError(t, err)
	identityProof, err := identitytypes.BuildIdentityResolutionProof(identityState, "portable.aet", 13)
	require.NoError(t, err)
	proof, err := NewAETServiceBindingProof(AETServiceBindingProof{
		Binding:		binding,
		Descriptor:		discovery,
		RegistryDescriptor:	descriptor,
		IdentityTrustedRoot:	root,
		IdentityProof:		identityProof,
		ServiceRegistryProof:	serviceProof.Proof,
		Height:			13,
	})
	require.NoError(t, err)

	verified, err := VerifyAETServiceBindings([]AETServiceBindingProof{proof, proof})
	require.NoError(t, err)
	require.Len(t, verified, 2)

	_, err = VerifyAETServiceBindings(nil)
	require.ErrorContains(t, err, "required")
}

func testAETIdentityState(t *testing.T, name string, endpoint string) identitytypes.IdentityState {
	t.Helper()
	owner := testAETAddress(1)
	state := identitytypes.EmptyIdentityState(identitytypes.DefaultIdentityParams())
	commitment, err := identitytypes.ComputeRegistrationCommitment(name, owner, "salt")
	require.NoError(t, err)
	state, err = identitytypes.CommitDomainRegistration(state, name, owner, commitment, 10)
	require.NoError(t, err)
	state, _, err = identitytypes.RevealRegisterDomain(state, name, owner, "salt", 11)
	require.NoError(t, err)
	state, _, err = identitytypes.PatchIdentityResolver(state, name+".aet", owner, identitytypes.ResolverPatch{
		ZoneEndpoint:	endpoint,
		UpdatedAtUnix:	12,
	}, 12)
	require.NoError(t, err)
	return state
}

func testAETAddress(seed byte) sdk.AccAddress {
	out := make([]byte, 20)
	out[19] = seed
	return sdk.AccAddress(out)
}
