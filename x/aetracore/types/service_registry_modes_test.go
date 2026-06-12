package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnChainRegistryModeStoresFullDescriptorAndProofFields(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	authHash := ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10)

	state, err := NewOnChainServiceRegistryState(service, authHash)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, state.Descriptor.ServiceID)
	require.Equal(t, ComputeServiceDescriptorHash(service), state.DescriptorHash)
	require.Equal(t, service.Interface.InterfaceHash, state.InterfaceDescriptorHash)
	require.Equal(t, registryPaymentModel(service), state.PaymentModel)
	require.Equal(t, service.Verification.Model, state.VerificationModel)
	require.Equal(t, authHash, state.OwnerAuthorizationHash)
	require.Equal(t, service.ExpiryHeight, state.ExpiryHeight)
	require.NoError(t, state.Validate())

	modeState, err := BuildServiceRegistryModeState(
		ServiceRegistryOnChain,
		[]ServiceDescriptor{service},
		map[string]string{service.ServiceID: authHash},
		10,
	)
	require.NoError(t, err)
	require.Len(t, modeState.OnChainStates, 1)
	require.Empty(t, modeState.HybridAnchors)
	require.NoError(t, modeState.Validate())

	descriptor, proof, found := modeState.OnChainDescriptorByID(service.ServiceID)
	require.True(t, found)
	require.Equal(t, service.ServiceID, descriptor.ServiceID)
	require.Equal(t, ServiceRegistryOnChain, proof.RegistryMode)
	require.Equal(t, modeState.StateRoot, proof.RegistryRoot)
	require.Equal(t, state.StateHash, proof.RecordHash)
	require.Equal(t, state.DescriptorHash, proof.DescriptorHash)
	require.NoError(t, proof.Validate())
}

func TestHybridRegistryModeStoresMinimalAnchor(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	service.Discovery.ProviderRoot = testHash(service.ServiceID + "/providers")

	anchor, err := NewHybridServiceRegistryAnchor(service)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, anchor.ServiceID)
	require.Equal(t, service.Owner, anchor.Owner)
	require.Equal(t, ComputeServiceDescriptorHash(service), anchor.DescriptorHash)
	require.Equal(t, service.Interface.InterfaceHash, anchor.InterfaceHash)
	require.Equal(t, service.Discovery.ProviderRoot, anchor.ProviderRoot)
	require.Equal(t, service.ExpiryHeight, anchor.ExpiryHeight)
	require.Equal(t, service.Verification.Model, anchor.VerificationModel)
	require.NoError(t, anchor.Validate())

	modeState, err := BuildServiceRegistryModeState(ServiceRegistryHybrid, []ServiceDescriptor{service}, nil, 12)
	require.NoError(t, err)
	require.Empty(t, modeState.OnChainStates)
	require.Len(t, modeState.HybridAnchors, 1)
	require.NoError(t, modeState.Validate())

	lookup, proof, found := modeState.HybridAnchorByID(service.ServiceID)
	require.True(t, found)
	require.Equal(t, anchor.ServiceID, lookup.ServiceID)
	require.Equal(t, ServiceRegistryHybrid, proof.RegistryMode)
	require.Equal(t, modeState.StateRoot, proof.RegistryRoot)
	require.Equal(t, anchor.AnchorHash, proof.RecordHash)
	require.Equal(t, anchor.DescriptorHash, proof.DescriptorHash)
	require.NoError(t, proof.Validate())
}

func TestServiceRegistryModeStateBuildsDeterministicRoots(t *testing.T) {
	first := testService("identity-resolver", ZoneIDIdentity)
	second := testService("payments-settlement", ZoneIDPayment)
	auth := map[string]string{
		first.ServiceID:	ComputeServiceOwnerAuthorizationHash(first.ServiceID, first.Owner, 20),
		second.ServiceID:	ComputeServiceOwnerAuthorizationHash(second.ServiceID, second.Owner, 20),
	}

	onChain, err := BuildServiceRegistryModeState(ServiceRegistryOnChain, []ServiceDescriptor{second, first}, auth, 20)
	require.NoError(t, err)
	require.Equal(t, []string{"identity-resolver", "payments-settlement"}, []string{
		onChain.OnChainStates[0].Descriptor.ServiceID,
		onChain.OnChainStates[1].Descriptor.ServiceID,
	})
	require.Equal(t, ComputeServiceRegistryModeStateRoot(onChain), onChain.StateRoot)

	hybridFirst := testOffChainService("indexer-feed", ZoneIDApplication)
	hybridFirst.Discovery.ProviderRoot = testHash(hybridFirst.ServiceID + "/providers")
	hybridSecond := testFogMarketService("fog-compute", ZoneIDApplication)
	hybrid, err := BuildServiceRegistryModeState(ServiceRegistryHybrid, []ServiceDescriptor{hybridFirst, hybridSecond}, nil, 21)
	require.NoError(t, err)
	require.Equal(t, []string{"fog-compute", "indexer-feed"}, []string{
		hybrid.HybridAnchors[0].ServiceID,
		hybrid.HybridAnchors[1].ServiceID,
	})
	require.Equal(t, ComputeServiceRegistryModeStateRoot(hybrid), hybrid.StateRoot)
}

func TestHybridRegistryRejectsMissingProviderRoot(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	service.Discovery.ProviderRoot = ""
	service.Execution.ProviderPoolID = ""

	_, err := NewHybridServiceRegistryAnchor(service)
	require.ErrorContains(t, err, "provider root")
}

func TestOnChainRegistryRejectsMissingOwnerAuthorization(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)

	_, err := NewOnChainServiceRegistryState(service, "")
	require.ErrorContains(t, err, "owner authorization")

	_, err = BuildServiceRegistryModeState(ServiceRegistryOnChain, []ServiceDescriptor{service}, nil, 10)
	require.ErrorContains(t, err, "requires owner authorization")
}

func TestDistributedRegistryMeshStoresSignedAdvisoryCache(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	service.Discovery.ProviderRoot = testHash(service.ServiceID + "/providers")
	providerID := "provider-alpha"
	signatureHash := ComputeMeshServiceRegistryAdvertisementSignatureHash(
		service.ServiceID,
		service.Owner,
		providerID,
		ComputeServiceDescriptorHash(service),
		30,
	)

	advertisement, err := NewMeshServiceRegistryAdvertisement(
		service,
		providerID,
		service.Execution.Endpoint,
		"aetra.services.v1",
		signatureHash,
		30,
	)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, advertisement.ServiceID)
	require.Equal(t, service.Owner, advertisement.Owner)
	require.Equal(t, providerID, advertisement.ProviderID)
	require.Equal(t, service.Execution.Endpoint, advertisement.Endpoint)
	require.Equal(t, ComputeServiceDescriptorHash(service), advertisement.DescriptorHash)
	require.Equal(t, service.Interface.InterfaceHash, advertisement.InterfaceHash)
	require.Equal(t, service.Discovery.ProviderRoot, advertisement.ProviderRoot)
	require.Equal(t, signatureHash, advertisement.SignatureHash)
	require.False(t, advertisement.HasAnchorProof())
	require.NoError(t, VerifyMeshServiceRegistryAdvertisementDescriptor(advertisement, service))

	cacheRecord, err := NewMeshServiceRegistryCacheRecord(advertisement, 88, 31)
	require.NoError(t, err)
	require.True(t, cacheRecord.AdvisoryOnly)
	require.Equal(t, uint64(88), cacheRecord.LocalReputation)
	require.Equal(t, service.ExpiryHeight, cacheRecord.ExpiryHeight)

	meshState, err := BuildDistributedRegistryMeshState([]MeshServiceRegistryCacheRecord{cacheRecord}, 31)
	require.NoError(t, err)
	require.Equal(t, ServiceRegistryMesh, meshState.Mode)
	require.Len(t, meshState.MeshRecords, 1)
	require.Empty(t, meshState.OnChainStates)
	require.Empty(t, meshState.HybridAnchors)
	require.NoError(t, meshState.Validate())

	lookup, proof, found := meshState.MeshLookup(service.ServiceID)
	require.True(t, found)
	require.True(t, lookup.AdvisoryOnly)
	require.Equal(t, cacheRecord.CacheHash, proof.RecordHash)
	require.Equal(t, ServiceRegistryMesh, proof.RegistryMode)
	require.Equal(t, meshState.StateRoot, proof.RegistryRoot)
	require.NoError(t, proof.Validate())
}

func TestDistributedRegistryMeshLookupCarriesAnchorProofWhenPresent(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	service.Discovery.ProviderRoot = testHash(service.ServiceID + "/providers")
	hybridState, err := BuildServiceRegistryModeState(ServiceRegistryHybrid, []ServiceDescriptor{service}, nil, 20)
	require.NoError(t, err)
	_, anchorProof, found := hybridState.HybridAnchorByID(service.ServiceID)
	require.True(t, found)

	providerID := "provider-alpha"
	signatureHash := ComputeMeshServiceRegistryAdvertisementSignatureHash(
		service.ServiceID,
		service.Owner,
		providerID,
		ComputeServiceDescriptorHash(service),
		30,
	)
	advertisement, err := NewMeshServiceRegistryAdvertisement(service, providerID, service.Execution.Endpoint, "aetra.services.v1", signatureHash, 30)
	require.NoError(t, err)
	advertisement, err = AttachMeshServiceRegistryAnchorProof(advertisement, anchorProof)
	require.NoError(t, err)
	require.True(t, advertisement.HasAnchorProof())

	cacheRecord, err := NewMeshServiceRegistryCacheRecord(advertisement, 91, 31)
	require.NoError(t, err)
	meshState, err := BuildDistributedRegistryMeshState([]MeshServiceRegistryCacheRecord{cacheRecord}, 31)
	require.NoError(t, err)

	lookup, _, found := meshState.MeshLookup(service.ServiceID)
	require.True(t, found)
	require.True(t, lookup.Advertisement.HasAnchorProof())
	require.Equal(t, anchorProof.ProofHash, lookup.Advertisement.AnchorProof.ProofHash)
	require.NoError(t, lookup.Validate())
}

func TestDistributedRegistryMeshRejectsUnsignedExpiredOrConsensusCache(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	service.Discovery.ProviderRoot = testHash(service.ServiceID + "/providers")
	providerID := "provider-alpha"

	_, err := NewMeshServiceRegistryAdvertisement(service, providerID, service.Execution.Endpoint, "aetra.services.v1", "", 30)
	require.ErrorContains(t, err, "signature")

	signatureHash := ComputeMeshServiceRegistryAdvertisementSignatureHash(
		service.ServiceID,
		service.Owner,
		providerID,
		ComputeServiceDescriptorHash(service),
		30,
	)
	advertisement, err := NewMeshServiceRegistryAdvertisement(service, providerID, service.Execution.Endpoint, "aetra.services.v1", signatureHash, 30)
	require.NoError(t, err)

	_, err = NewMeshServiceRegistryCacheRecord(advertisement, 10, service.ExpiryHeight)
	require.ErrorContains(t, err, "expired")

	cacheRecord, err := NewMeshServiceRegistryCacheRecord(advertisement, 10, 31)
	require.NoError(t, err)
	cacheRecord.AdvisoryOnly = false
	cacheRecord.CacheHash = ComputeMeshServiceRegistryCacheRecordHash(cacheRecord)
	require.ErrorContains(t, cacheRecord.Validate(), "advisory")

	tampered := service
	tampered.Version = 2
	tampered.Interface.Version = 2
	tampered.Interface.InterfaceHash = ComputeServiceInterfaceHash(tampered.Interface)
	require.ErrorContains(t, VerifyMeshServiceRegistryAdvertisementDescriptor(advertisement, tampered), "descriptor hash")
}
