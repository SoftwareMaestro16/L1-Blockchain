package types

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestDistributedServiceRecordsEndpointsAndKeys(t *testing.T) {
	iface := testDistributedInterface(t, "l1.api.v1")
	endpoint := testDistributedEndpoint(t, "svc-api", "api-main", DistributedEndpointAPI, iface.InterfaceHash, 1)
	record := testDistributedRecord(t, "svc-api", "ledger-api", DistributedServiceAPI, coretypes.ZoneIDApplication, iface.InterfaceHash, []DistributedServiceEndpoint{endpoint})

	recordKey, err := DistributedServiceRecordKey(record.ServiceID)
	require.NoError(t, err)
	require.Equal(t, "services/distributed/records/svc-api", recordKey)

	endpointKey, err := DistributedServiceEndpointKey(endpoint.ServiceID, endpoint.EndpointID)
	require.NoError(t, err)
	require.Equal(t, "services/distributed/endpoints/svc-api/api-main", endpointKey)

	interfaceKey, err := DistributedServiceInterfaceKey(iface.InterfaceHash)
	require.NoError(t, err)
	require.Equal(t, "services/distributed/interfaces/"+iface.InterfaceHash, interfaceKey)

	proofKey, err := DistributedServiceProofKey(record.ServiceID, endpoint.CommitmentHash)
	require.NoError(t, err)
	require.Equal(t, "services/distributed/proofs/svc-api/"+endpoint.CommitmentHash, proofKey)

	require.NoError(t, record.Validate())
	require.NoError(t, endpoint.Validate())
	require.NoError(t, iface.Validate())
}

func TestDistributedDiscoveryIsDeterministicAndZoneAware(t *testing.T) {
	apiIface := testDistributedInterface(t, "l1.api.v1")
	computeIface := testDistributedInterface(t, "l1.compute.v1")
	apiEndpoint := testDistributedEndpoint(t, "svc-api", "api-main", DistributedEndpointAPI, apiIface.InterfaceHash, 10)
	computeEndpoint := testDistributedEndpoint(t, "svc-compute", "compute-main", DistributedEndpointCompute, computeIface.InterfaceHash, 20)
	apiRecord := testDistributedRecord(t, "svc-api", "ledger-api", DistributedServiceAPI, coretypes.ZoneIDApplication, apiIface.InterfaceHash, []DistributedServiceEndpoint{apiEndpoint})
	computeRecord := testDistributedRecord(t, "svc-compute", "fog-compute", DistributedServiceOffChain, coretypes.ZoneIDContract, computeIface.InterfaceHash, []DistributedServiceEndpoint{computeEndpoint})

	stateA, err := BuildDistributedDiscoveryState(
		[]DistributedServiceRecord{computeRecord, apiRecord},
		[]DistributedServiceEndpoint{computeEndpoint, apiEndpoint},
		[]DistributedInterfaceDescriptor{computeIface, apiIface},
		nil,
		30,
	)
	require.NoError(t, err)
	stateB, err := BuildDistributedDiscoveryState(
		[]DistributedServiceRecord{apiRecord, computeRecord},
		[]DistributedServiceEndpoint{apiEndpoint, computeEndpoint},
		[]DistributedInterfaceDescriptor{apiIface, computeIface},
		nil,
		30,
	)
	require.NoError(t, err)
	require.Equal(t, stateA.StateRoot, stateB.StateRoot)

	apis, err := DiscoverDistributedServices(stateA, DistributedServiceAPI, coretypes.ZoneIDApplication, apiIface.InterfaceHash, 31)
	require.NoError(t, err)
	require.Len(t, apis, 1)
	require.Equal(t, "svc-api", apis[0].ServiceID)

	missing, err := DiscoverDistributedServices(stateA, DistributedServiceApplication, coretypes.ZoneIDApplication, "", 31)
	require.NoError(t, err)
	require.Empty(t, missing)

	roots := ComputeDistributedServiceDiscoveryRoots(stateA)
	require.Equal(t, stateA.StateRoot, roots.StateRoot)
	require.NoError(t, coretypes.ValidateHash("distributed state root", roots.StateRoot))
}

func TestOffConsensusExecutionRequiresCommittedMessageOrProof(t *testing.T) {
	err := ValidateOffConsensusAuthority(DistributedAuthorityMessageCommit, nil)
	require.ErrorContains(t, err, "requires message or proof commitment")

	messageCommitment, err := NewDistributedExecutionCommitment(DistributedExecutionCommitment{
		ServiceID:		"svc-api",
		EndpointID:		"api-main",
		Kind:			DistributedCommitmentMessage,
		MessageID:		testDistributedHash("message"),
		ResultHash:		testDistributedHash("result"),
		CommittedHeight:	40,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateOffConsensusAuthority(DistributedAuthorityMessageCommit, []DistributedExecutionCommitment{messageCommitment}))

	proofCommitment, err := NewDistributedExecutionCommitment(DistributedExecutionCommitment{
		ServiceID:		"svc-api",
		EndpointID:		"api-main",
		Kind:			DistributedCommitmentProof,
		ProofHash:		testDistributedHash("proof"),
		ResultHash:		testDistributedHash("result"),
		CommittedHeight:	40,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateOffConsensusAuthority(DistributedAuthorityProofCommit, []DistributedExecutionCommitment{proofCommitment}))
	require.ErrorContains(t, ValidateOffConsensusAuthority(DistributedAuthorityProofCommit, []DistributedExecutionCommitment{messageCommitment}), "proof commitment")
}

func TestDistributedDiscoveryRejectsInconsistentIndexes(t *testing.T) {
	iface := testDistributedInterface(t, "l1.api.v1")
	endpoint := testDistributedEndpoint(t, "svc-api", "api-main", DistributedEndpointAPI, iface.InterfaceHash, 1)
	record := testDistributedRecord(t, "svc-api", "ledger-api", DistributedServiceAPI, coretypes.ZoneIDApplication, iface.InterfaceHash, []DistributedServiceEndpoint{endpoint})

	badEndpoint := endpoint
	badEndpoint.InterfaceHash = testDistributedHash("other-interface")
	badEndpoint.CommitmentHash = ""
	badEndpoint, err := NewDistributedServiceEndpoint(badEndpoint)
	require.NoError(t, err)
	_, err = BuildDistributedDiscoveryState([]DistributedServiceRecord{record}, []DistributedServiceEndpoint{badEndpoint}, []DistributedInterfaceDescriptor{iface}, nil, 30)
	require.ErrorContains(t, err, "interface mismatch")

	_, err = BuildDistributedDiscoveryState([]DistributedServiceRecord{record}, []DistributedServiceEndpoint{endpoint}, nil, nil, 30)
	require.ErrorContains(t, err, "missing interface descriptor")
}

func testDistributedRecord(t *testing.T, serviceID, name string, kind DistributedServiceKind, zoneID string, interfaceHash string, endpoints []DistributedServiceEndpoint) DistributedServiceRecord {
	t.Helper()
	record, err := NewDistributedServiceRecord(DistributedServiceRecord{
		ServiceID:	serviceID,
		ServiceName:	name,
		Kind:		kind,
		Owner:		coretypes.DefaultAuthority,
		ZoneID:		zoneID,
		InterfaceHash:	interfaceHash,
		EndpointRoot:	ComputeDistributedEndpointRoot(endpoints),
		DescriptorHash:	testDistributedHash(serviceID + "/descriptor"),
		MetadataHash:	testDistributedHash(serviceID + "/metadata"),
		CreatedHeight:	10,
		UpdatedHeight:	10,
		ExpiryHeight:	100,
		Discoverable:	true,
	})
	require.NoError(t, err)
	return record
}

func testDistributedEndpoint(t *testing.T, serviceID, endpointID string, kind DistributedEndpointKind, interfaceHash string, priority uint32) DistributedServiceEndpoint {
	t.Helper()
	target := "app/" + endpointID
	if kind == DistributedEndpointAPI {
		target = "https://api.aetra.local/" + endpointID
	}
	endpoint, err := NewDistributedServiceEndpoint(DistributedServiceEndpoint{
		ServiceID:	serviceID,
		EndpointID:	endpointID,
		Kind:		kind,
		ZoneID:		coretypes.ZoneIDApplication,
		Target:		target,
		InterfaceHash:	interfaceHash,
		Priority:	priority,
		Weight:		1,
		MetadataHash:	testDistributedHash(endpointID + "/metadata"),
	})
	require.NoError(t, err)
	return endpoint
}

func testDistributedInterface(t *testing.T, name string) DistributedInterfaceDescriptor {
	t.Helper()
	iface, err := NewDistributedInterfaceDescriptor(DistributedInterfaceDescriptor{
		InterfaceHash:	testDistributedHash(name + "/interface"),
		InterfaceName:	name,
		Version:	1,
		SchemaHash:	testDistributedHash(name + "/schema"),
		MethodRoot:	testDistributedHash(name + "/methods"),
		EventRoot:	testDistributedHash(name + "/events"),
		ErrorRoot:	testDistributedHash(name + "/errors"),
	})
	require.NoError(t, err)
	return iface
}

func testDistributedHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
