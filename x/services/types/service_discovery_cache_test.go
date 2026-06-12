package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServiceDiscoveryCacheRejectsStaleAndOutlivingBounds(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	state := testResolverRegistryState(t, descriptor, nil, 20)
	proof, err := state.QueryServiceProof(QueryServiceProof{ServiceID: descriptor.ServiceID})
	require.NoError(t, err)
	discovery, err := ProjectServiceDiscoveryDescriptorFromCore(descriptor, descriptor.Discovery.ServiceName, proof.Proof.ProofHash, "")
	require.NoError(t, err)

	constraints := ServiceDiscoveryCacheConstraints{
		RegistryExpiryHeight:		descriptor.ExpiryHeight,
		InterfaceCompatibleUntil:	descriptor.Discovery.CacheExpiryHeight,
		CurrentHeight:			25,
	}
	record, err := NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		discovery.ServiceID,
		DescriptorHash:		discovery.DescriptorHash,
		InterfaceHash:		discovery.InterfaceHash,
		Source:			ServiceResolutionOnChainRegistry,
		ProofHeightOptional:	proof.Proof.ProofHeight,
		ExpiresHeight:		80,
		FetchedAtHeight:	24,
		Trust:			ServiceDiscoveryCacheVerified,
	}, constraints)
	require.NoError(t, err)
	require.NoError(t, record.Validate(constraints))

	_, err = NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		discovery.ServiceID,
		DescriptorHash:		discovery.DescriptorHash,
		InterfaceHash:		discovery.InterfaceHash,
		Source:			ServiceResolutionOnChainRegistry,
		ExpiresHeight:		descriptor.ExpiryHeight + 1,
		FetchedAtHeight:	24,
		Trust:			ServiceDiscoveryCacheAdvisory,
	}, constraints)
	require.ErrorContains(t, err, "registry expiry")

	_, err = NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		discovery.ServiceID,
		DescriptorHash:		discovery.DescriptorHash,
		InterfaceHash:		discovery.InterfaceHash,
		Source:			ServiceResolutionOnChainRegistry,
		ExpiresHeight:		descriptor.Discovery.CacheExpiryHeight + 1,
		FetchedAtHeight:	24,
		Trust:			ServiceDiscoveryCacheAdvisory,
	}, constraints)
	require.ErrorContains(t, err, "interface compatibility")

	stale := constraints
	stale.CurrentHeight = record.ExpiresHeight
	require.ErrorContains(t, record.Validate(stale), "stale")

	_, err = NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		discovery.ServiceID,
		DescriptorHash:		discovery.DescriptorHash,
		InterfaceHash:		discovery.InterfaceHash,
		Source:			ServiceResolutionDistributedMesh,
		ExpiresHeight:		70,
		FetchedAtHeight:	24,
		Trust:			ServiceDiscoveryCacheVerified,
	}, constraints)
	require.ErrorContains(t, err, "advisory")
}

func TestDiscoveryCacheInvalidatesOnServiceUpdate(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	discovery, err := ProjectServiceDiscoveryDescriptorFromCore(descriptor, descriptor.Discovery.ServiceName, "", "")
	require.NoError(t, err)
	constraints := ServiceDiscoveryCacheConstraints{RegistryExpiryHeight: 100, InterfaceCompatibleUntil: 90, CurrentHeight: 25}
	updated, err := NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		discovery.ServiceID,
		DescriptorHash:		discovery.DescriptorHash,
		InterfaceHash:		discovery.InterfaceHash,
		Source:			ServiceResolutionDistributedMesh,
		ExpiresHeight:		80,
		FetchedAtHeight:	24,
		Trust:			ServiceDiscoveryCacheAdvisory,
	}, constraints)
	require.NoError(t, err)
	other, err := NewServiceDiscoveryCacheRecord(ServiceDiscoveryCacheRecord{
		ServiceID:		"other-service",
		DescriptorHash:		testInterfaceHash("other/descriptor"),
		InterfaceHash:		testInterfaceHash("other/interface"),
		Source:			ServiceResolutionDistributedMesh,
		ExpiresHeight:		80,
		FetchedAtHeight:	24,
		Trust:			ServiceDiscoveryCacheAdvisory,
	}, constraints)
	require.NoError(t, err)
	event, err := NewServiceDiscoveryUpdateEvent(discovery.ServiceID, testInterfaceHash("new/descriptor"), discovery.InterfaceHash, 30)
	require.NoError(t, err)

	remaining, err := InvalidateDiscoveryCacheOnServiceUpdate([]ServiceDiscoveryCacheRecord{updated, other}, event)
	require.NoError(t, err)
	require.Len(t, remaining, 1)
	require.Equal(t, other.ServiceID, remaining[0].ServiceID)
}

func TestSignedServiceAdvertisementRejectsForgedDiscoveryRecords(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	discovery, err := ProjectServiceDiscoveryDescriptorFromCore(descriptor, descriptor.Discovery.ServiceName, "", "")
	require.NoError(t, err)
	ad, err := NewSignedServiceAdvertisement(SignedServiceAdvertisement{
		ServiceName:	discovery.ServiceName,
		Descriptor:	discovery,
		Endpoint:	discovery.Endpoint,
		InterfaceHash:	discovery.InterfaceHash,
		Signer:		discovery.Owner,
		ExpiresHeight:	80,
		IssuedAtHeight:	25,
		Nonce:		1,
	})
	require.NoError(t, err)
	require.NoError(t, ad.Validate())

	source, err := ServiceResolverSourceRecordFromAdvertisement(ad, descriptor)
	require.NoError(t, err)
	require.Equal(t, ServiceResolutionSignedCache, source.Source)
	require.Equal(t, ad.SignatureHash, source.SignatureHash)

	forged := ad
	forged.Endpoint = "https://forged.aetra.local/v1"
	require.ErrorContains(t, forged.Validate(), "endpoint mismatch")

	tampered := ad
	tampered.Nonce++
	require.ErrorContains(t, tampered.Validate(), "hash mismatch")

	forgedSigner := ad
	forgedSigner.Signer = "unknown-provider"
	forgedSigner.AdvertisementHash = ""
	forgedSigner.SignatureHash = ""
	_, err = NewSignedServiceAdvertisement(forgedSigner)
	require.ErrorContains(t, err, "signer")
}

func TestDiscoveryQueryIndexesAndFallbackPolicy(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.Discovery.IdentityName = "portable.aet"
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	binding, err := coretypes.NewIdentityServiceBindingFromDescriptor(descriptor)
	require.NoError(t, err)
	state := testResolverRegistryState(t, descriptor, []coretypes.IdentityServiceBinding{binding}, 20)

	byName, err := QueryServiceDiscoveryFromState(state, QueryServiceDiscovery{ServiceName: descriptor.Discovery.ServiceName, IncludeProof: true}, 25)
	require.NoError(t, err)
	require.True(t, byName.Found)
	require.Equal(t, descriptor.ServiceID, byName.Descriptor.ServiceID)
	require.Equal(t, byName.Proof.ProofHash, byName.Descriptor.ProofOptional)

	byIdentity, err := QueryServiceDiscoveriesByIdentityFromState(state, QueryServiceDiscoveriesByIdentity{IdentityName: "portable.aet", IncludeProof: true, CurrentHeight: 25})
	require.NoError(t, err)
	require.Equal(t, uint64(1), byIdentity.Total)
	require.Equal(t, descriptor.ServiceID, byIdentity.Services[0].Descriptor.ServiceID)
	require.NotEmpty(t, byIdentity.Services[0].Proof.ProofHash)

	policy, err := NewServiceResolverFallbackPolicy(nil)
	require.NoError(t, err)
	require.Equal(t, DefaultServiceResolverFallbackOrder(), policy.Order)
	require.NoError(t, policy.Validate())
	require.ErrorContains(t, ServiceResolverFallbackPolicy{Order: []ServiceResolutionSource{ServiceResolutionSignedCache, ServiceResolutionSignedCache}}.ValidateFormat(), "duplicate")
}
