package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestCanonicalServiceDescriptorFieldsAndHash(t *testing.T) {
	descriptor, err := NewCanonicalServiceDescriptor(CanonicalServiceDescriptor{
		ServiceID:		"svc-api",
		EndpointType:		CanonicalEndpointAPI,
		InterfaceHash:		testDistributedHash("iface"),
		SupportedMethods:	[]string{"submit", "query", "query"},
		AuthModel:		"owner-or-provider",
		StateDependency:	"message-or-proof",
		Owner:			coretypes.DefaultAuthority,
		ZoneID:			coretypes.ZoneIDApplication,
		Version:		2,
		EndpointURIHash:	testDistributedHash("endpoint"),
		MetadataHash:		testDistributedHash("metadata"),
		TTLHeight:		100,
		Status:			CanonicalServiceStatusActive,
		Capabilities:		[]string{"proofs", "callbacks", "proofs"},
	})
	require.NoError(t, err)
	require.Equal(t, "svc-api", descriptor.ServiceID)
	require.Equal(t, CanonicalEndpointAPI, descriptor.EndpointType)
	require.Equal(t, []string{"query", "submit"}, descriptor.SupportedMethods)
	require.Equal(t, []string{"callbacks", "proofs"}, descriptor.Capabilities)
	require.Equal(t, ComputeCanonicalServiceDescriptorHash(descriptor), descriptor.DescriptorHash)
	require.NoError(t, descriptor.Validate())
}

func TestProjectDistributedServiceDescriptor(t *testing.T) {
	iface := testDistributedInterface(t, "l1.api.v1")
	endpoint := testDistributedEndpoint(t, "svc-api", "api-main", DistributedEndpointAPI, iface.InterfaceHash, 1)
	record := testDistributedRecord(t, "svc-api", "ledger-api", DistributedServiceAPI, coretypes.ZoneIDApplication, iface.InterfaceHash, []DistributedServiceEndpoint{endpoint})

	descriptor, err := ProjectDistributedServiceDescriptor(record, endpoint, iface, []string{"submit", "query"}, "owner-or-provider", "messages/proofs", []string{"callbacks", "proofs"})
	require.NoError(t, err)
	require.Equal(t, record.ServiceID, descriptor.ServiceID)
	require.Equal(t, CanonicalEndpointAPI, descriptor.EndpointType)
	require.Equal(t, record.InterfaceHash, descriptor.InterfaceHash)
	require.Equal(t, endpoint.CommitmentHash, descriptor.EndpointURIHash)
	require.Equal(t, record.MetadataHash, descriptor.MetadataHash)
	require.Equal(t, record.ExpiryHeight, descriptor.TTLHeight)
	require.Equal(t, CanonicalServiceStatusActive, descriptor.Status)
}

func TestCanonicalServiceDescriptorValidationRejectsMalformedFields(t *testing.T) {
	descriptor, err := NewCanonicalServiceDescriptor(CanonicalServiceDescriptor{
		ServiceID:		"svc-api",
		EndpointType:		CanonicalEndpointAPI,
		InterfaceHash:		testDistributedHash("iface"),
		SupportedMethods:	[]string{"query"},
		AuthModel:		"owner-or-provider",
		StateDependency:	"messages/proofs",
		Owner:			coretypes.DefaultAuthority,
		ZoneID:			coretypes.ZoneIDApplication,
		Version:		1,
		EndpointURIHash:	testDistributedHash("endpoint"),
		MetadataHash:		testDistributedHash("metadata"),
		TTLHeight:		100,
		Status:			CanonicalServiceStatusActive,
		Capabilities:		[]string{"proofs"},
	})
	require.NoError(t, err)

	badStatus := descriptor
	badStatus.Status = CanonicalServiceStatus("unknown")
	require.ErrorContains(t, badStatus.Validate(), "status")

	badMethods := descriptor
	badMethods.SupportedMethods = []string{}
	badMethods.DescriptorHash = ""
	_, err = NewCanonicalServiceDescriptor(badMethods)
	require.ErrorContains(t, err, "supported method")

	badHash := descriptor
	badHash.MetadataHash = testDistributedHash("changed")
	require.ErrorContains(t, badHash.Validate(), "hash mismatch")
}

func TestProjectDistributedServiceDescriptorRejectsMismatchedEndpoint(t *testing.T) {
	iface := testDistributedInterface(t, "l1.api.v1")
	endpoint := testDistributedEndpoint(t, "svc-api", "api-main", DistributedEndpointAPI, iface.InterfaceHash, 1)
	record := testDistributedRecord(t, "svc-api", "ledger-api", DistributedServiceAPI, coretypes.ZoneIDApplication, iface.InterfaceHash, []DistributedServiceEndpoint{endpoint})
	endpoint.ServiceID = "other"

	_, err := ProjectDistributedServiceDescriptor(record, endpoint, iface, []string{"query"}, "owner", "messages/proofs", []string{"proofs"})
	require.ErrorContains(t, err, "service mismatch")
}
