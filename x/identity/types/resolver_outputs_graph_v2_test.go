package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityGraphStateKeysMatchSpec(t *testing.T) {
	nodeKey, err := IdentityGraphNodeKey("domain/alice.aet")
	require.NoError(t, err)
	require.Equal(t, "identity/graph/node/domain/alice.aet", nodeKey)

	edgeKey, err := IdentityGraphEdgeKey("domain/alice.aet", IdentityGraphNodeService, "svc-pay")
	require.NoError(t, err)
	require.Equal(t, "identity/graph/edge/domain/alice.aet/service/svc-pay", edgeKey)

	rootKey, err := IdentityGraphRootKey(42)
	require.NoError(t, err)
	require.Equal(t, "identity/graph/root/00000000000000000042", rootKey)
}

func TestIdentityResolverOutputTypesValidateAndHash(t *testing.T) {
	nameHash := identityHash("alice.aet")
	service := testIdentityGraphServiceEndpoint()
	contract := NewContractTargetV2(ResolverKeyContract, addr(3), 12)
	composite, err := NewCompositeIdentityObjectV2(CompositeIdentityObjectV2{
		ObjectID:	"composite/alice",
		NameHash:	nameHash,
		ComponentIDs:	[]string{"account/" + identityHash("account"), "service/svc-pay"},
		MetadataHash:	identityHash("composite-metadata"),
	})
	require.NoError(t, err)

	outputs := []IdentityResolverOutputV2{
		{OutputID: "account/" + identityHash("account"), NameHash: nameHash, OutputType: IdentityResolverOutputAccountAddress, Account: addr(2), Height: 12},
		{OutputID: "zone/" + identityHash("zone"), NameHash: nameHash, OutputType: IdentityResolverOutputZoneEndpoint, ZoneEndpoint: "IDENTITY_ZONE/0001/resolve", Height: 12},
		{OutputID: "service/svc-pay", NameHash: nameHash, OutputType: IdentityResolverOutputServiceEndpoint, Service: &service, Height: 12},
		{OutputID: "contract/contract", NameHash: nameHash, OutputType: IdentityResolverOutputContractEndpoint, Contract: &contract, Height: 12},
		{OutputID: "composite/alice", NameHash: nameHash, OutputType: IdentityResolverOutputCompositeIdentityObject, Composite: &composite, Height: 12},
	}
	for _, output := range outputs {
		constructed, err := NewIdentityResolverOutputV2(output)
		require.NoError(t, err)
		require.NoError(t, constructed.Validate())
		require.Equal(t, ComputeIdentityResolverOutputV2Hash(constructed), constructed.OutputHash)
	}
	require.NotEmpty(t, ComputeIdentityResolverOutputRootV2(outputsWithHashes(t, outputs)))
}

func TestIdentityGraphBuildsFromResolverOutputsDeterministically(t *testing.T) {
	record := testIdentityGraphUnifiedRecord(t)
	outputs, err := BuildIdentityResolverOutputsFromUnifiedRecordV2(record, 12)
	require.NoError(t, err)
	require.Len(t, outputs, 4)

	graph, err := BuildIdentityGraphFromResolverOutputsV2(record.NameHash, record.Owner, outputs, 12)
	require.NoError(t, err)
	require.NoError(t, graph.Validate())
	require.Len(t, graph.Nodes, 5)
	require.Len(t, graph.Edges, 4)
	require.Equal(t, graph.Root.RootHash, ComputeIdentityGraphRootHash(graph.Root))

	reordered := append([]IdentityResolverOutputV2{outputs[3], outputs[0]}, outputs[1:3]...)
	repeated, err := BuildIdentityGraphFromResolverOutputsV2(record.NameHash, record.Owner, reordered, 12)
	require.NoError(t, err)
	require.Equal(t, graph.Root.RootHash, repeated.Root.RootHash)
	require.Equal(t, graph.Root.NodeRoot, repeated.Root.NodeRoot)
	require.Equal(t, graph.Root.EdgeRoot, repeated.Root.EdgeRoot)
}

func TestIdentityGraphRejectsTamperedEdgeTargetHash(t *testing.T) {
	record := testIdentityGraphUnifiedRecord(t)
	outputs, err := BuildIdentityResolverOutputsFromUnifiedRecordV2(record, 12)
	require.NoError(t, err)
	graph, err := BuildIdentityGraphFromResolverOutputsV2(record.NameHash, record.Owner, outputs, 12)
	require.NoError(t, err)

	graph.Edges[0].TargetHash = identityHash("wrong-target")
	graph.Edges[0].EdgeHash = ComputeIdentityEdgeHash(graph.Edges[0])
	root, err := BuildIdentityGraphRoot(graph.Height, graph.Nodes, graph.Edges, graph.Outputs)
	require.NoError(t, err)
	graph.Root = root
	err = graph.Validate()
	require.ErrorContains(t, err, "target hash mismatch")
}

func TestIdentityGraphRejectsUnknownOutputType(t *testing.T) {
	_, err := NewIdentityResolverOutputV2(IdentityResolverOutputV2{
		OutputID:	"bad/output",
		NameHash:	identityHash("alice.aet"),
		OutputType:	"bad",
		Height:		12,
	})
	require.ErrorContains(t, err, "unknown identity resolver output type")
}

func testIdentityGraphUnifiedRecord(t *testing.T) UnifiedResolutionRecordV2 {
	t.Helper()
	nameHash := identityHash("alice.aet")
	record := UnifiedResolutionRecordV2{
		NameHash:	nameHash,
		Owner:		addr(1),
		PrimaryAddress:	addr(2),
		ContractTargets: []ContractTargetV2{
			NewContractTargetV2(ResolverKeyContract, addr(3), 12),
		},
		ServiceEndpoints: []ServiceEndpointV2{
			testIdentityGraphServiceEndpoint(),
		},
		RoutingMetadata: RoutingMetadataV2{
			ZoneID:			"IDENTITY_ZONE",
			ShardID:		"0001",
			RouteID:		"resolve",
			TargetType:		string(IdentityResolutionTargetRoute),
			PreferredTarget:	"identity/alice",
		},
		RecordVersion:		1,
		RecordTTL:		30,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
	return record
}

func testIdentityGraphServiceEndpoint() ServiceEndpointV2 {
	return ServiceEndpointV2{
		Key:		"svc-pay",
		Endpoint:	"https://pay.alice.aet",
		ServiceID:	"svc-pay",
		ServiceType:	"payment.v1",
		Transport:	"https",
		AuthPolicy:	"none",
		Priority:	10,
		Weight:		1,
		TTL:		30,
	}
}

func outputsWithHashes(t *testing.T, outputs []IdentityResolverOutputV2) []IdentityResolverOutputV2 {
	t.Helper()
	constructed := make([]IdentityResolverOutputV2, 0, len(outputs))
	for _, output := range outputs {
		next, err := NewIdentityResolverOutputV2(output)
		require.NoError(t, err)
		constructed = append(constructed, next)
	}
	return constructed
}
