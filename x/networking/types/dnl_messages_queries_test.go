package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDNLMessagesRegisterUpdateReputationTableAndExpire(t *testing.T) {
	salt := []byte("aetra-test-network")
	state, err := BuildDNLRoutingState(nil, nil, nil, nil, 1)
	require.NoError(t, err)
	node := testDNLNodeRecord(t, 0x91, salt, []string{"financial"}, []string{"svc-pay"}, 100)
	reputation := testDNLReputation(t, node.NodeID, 10)

	state, err = RegisterNodeRecordInRoutingState(state, MsgRegisterNodeRecord{
		Authority:	node.OperatorAddress,
		Record:		node,
		Reputation:	reputation,
		NetworkSalt:	salt,
		Height:		10,
	})
	require.NoError(t, err)
	require.Len(t, state.Nodes, 1)
	require.Len(t, state.ZoneIndex, 1)
	require.Len(t, state.ServiceIndex, 1)

	updated := testDNLNodeRecord(t, 0x91, salt, []string{"financial", "identity"}, []string{"svc-pay", "svc-name"}, 120)
	updatedReputation := testDNLReputation(t, updated.NodeID, 11)
	state, err = UpdateNodeRecordInRoutingState(state, MsgUpdateNodeRecord{
		Authority:	updated.OperatorAddress,
		Record:		updated,
		Reputation:	updatedReputation,
		NetworkSalt:	salt,
		Height:		11,
	})
	require.NoError(t, err)
	require.Len(t, state.ZoneIndex, 2)
	require.Len(t, state.ServiceIndex, 2)

	betterReputation := testDNLReputation(t, updated.NodeID, 12)
	betterReputation.Reputation.ScoreBps = 9_500
	betterReputation.CommitmentHash = ComputeReputationCommitmentHash(betterReputation)
	state, err = SubmitReputationCommitmentInRoutingState(state, MsgSubmitReputationCommitment{
		Authority:	updated.OperatorAddress,
		Commitment:	betterReputation,
		Height:		12,
	})
	require.NoError(t, err)

	route := testDNLRouteForNode(t, "financial", "svc-pay", updated.NodeID, 1, 8_500, 100)
	table := testRoutingTableForRoutes(t, 7, route)
	state, err = UpdateRoutingTableInRoutingState(state, MsgUpdateRoutingTable{
		Authority:	updated.OperatorAddress,
		Table:		table,
		Height:		13,
	})
	require.NoError(t, err)
	require.Len(t, state.Tables, 1)

	state, err = ExpireNodeRecordInRoutingState(state, MsgExpireNodeRecord{
		Authority:	updated.OperatorAddress,
		NodeID:		updated.NodeID,
		ReasonHash:	HashParts("operator-expired"),
		Height:		14,
	})
	require.NoError(t, err)
	require.Empty(t, state.Nodes)
	require.Empty(t, state.Reputation)
	require.Empty(t, state.Tables)
}

func TestDNLQueriesReturnProofAttachedResponses(t *testing.T) {
	salt := []byte("aetra-test-network")
	node := testDNLNodeRecord(t, 0x92, salt, []string{"contract"}, []string{"svc-contract"}, 100)
	reputation := testDNLReputation(t, node.NodeID, 20)
	route := testDNLRouteForNode(t, "contract", "svc-contract", node.NodeID, 1, 8_000, 100)
	table := testRoutingTableForRoutes(t, 3, route)
	state, err := BuildDNLRoutingState([]NodeRecord{node}, []ReputationCommitment{reputation}, nil, []RoutingTable{table}, 30)
	require.NoError(t, err)

	nodeRes, err := QueryNodeRecordFromRoutingState(state, QueryNodeRecord{NodeID: node.NodeID, IncludeProof: true})
	require.NoError(t, err)
	require.True(t, nodeRes.Found)
	require.Equal(t, node.NodeID, nodeRes.Record.NodeID)
	require.NoError(t, nodeRes.Proof.Validate())

	zoneRes, err := QueryNodesByZoneFromRoutingState(state, QueryNodesByZone{ZoneID: "contract"})
	require.NoError(t, err)
	require.Len(t, zoneRes.Records, 1)
	require.Len(t, zoneRes.Proofs, 1)

	serviceRes, err := QueryNodesByServiceFromRoutingState(state, QueryNodesByService{ServiceID: "svc-contract"})
	require.NoError(t, err)
	require.Len(t, serviceRes.Records, 1)
	require.Len(t, serviceRes.Proofs, 1)

	tableRes, err := QueryRoutingTableFromRoutingState(state, QueryRoutingTable{Epoch: 3, IncludeProof: true})
	require.NoError(t, err)
	require.True(t, tableRes.Found)
	require.Equal(t, table.TableRoot, tableRes.Table.TableRoot)
	require.NoError(t, tableRes.Proof.Validate())

	repRes, err := QueryReputationCommitmentFromRoutingState(state, QueryReputationCommitment{NodeID: node.NodeID, IncludeProof: true})
	require.NoError(t, err)
	require.True(t, repRes.Found)
	require.Equal(t, reputation.CommitmentHash, repRes.Commitment.CommitmentHash)
	require.NoError(t, repRes.Proof.Validate())

	key, err := RoutingNodeKey(node.NodeID)
	require.NoError(t, err)
	proofRes, err := QueryLookupProofFromRoutingState(state, QueryLookupProof{Key: key})
	require.NoError(t, err)
	require.True(t, proofRes.Found)
	require.Equal(t, state.StateRoot, proofRes.Proof.StateRoot)
}

func TestDNLRoutingExportImportRoundTrip(t *testing.T) {
	salt := []byte("aetra-test-network")
	node := testDNLNodeRecord(t, 0x93, salt, []string{"storage"}, []string{"svc-store"}, 100)
	reputation := testDNLReputation(t, node.NodeID, 20)
	table := testRoutingTableForRoutes(t, 5, testDNLRouteForNode(t, "storage", "svc-store", node.NodeID, 1, 8_000, 100))
	state, err := BuildDNLRoutingState([]NodeRecord{node}, []ReputationCommitment{reputation}, nil, []RoutingTable{table}, 30)
	require.NoError(t, err)

	exported, err := ExportDNLRoutingState(state)
	require.NoError(t, err)
	require.Equal(t, state.StateRoot, exported.Manifest.StateRoot)
	require.NotEmpty(t, exported.Manifest.ManifestHash)

	imported, err := ImportDNLRoutingState(exported)
	require.NoError(t, err)
	require.Equal(t, state.StateRoot, imported.StateRoot)

	exported.Manifest.StateRoot = HashParts("wrong-state-root")
	exported.Manifest.ManifestHash = ComputeDNLRoutingExportManifestHash(exported.Manifest)
	_, err = ImportDNLRoutingState(exported)
	require.ErrorContains(t, err, "root mismatch")
}

func TestDNLMessagesRejectUnauthorizedMutations(t *testing.T) {
	salt := []byte("aetra-test-network")
	node := testDNLNodeRecord(t, 0x94, salt, []string{"apps"}, []string{"svc-app"}, 100)
	state, err := BuildDNLRoutingState([]NodeRecord{node}, []ReputationCommitment{testDNLReputation(t, node.NodeID, 20)}, nil, nil, 20)
	require.NoError(t, err)

	_, err = ExpireNodeRecordInRoutingState(state, MsgExpireNodeRecord{
		Authority:	"operator/wrong",
		NodeID:		node.NodeID,
		ReasonHash:	HashParts("wrong"),
		Height:		21,
	})
	require.ErrorContains(t, err, "authority")

	other := testDNLNodeRecord(t, 0x95, salt, []string{"apps"}, []string{"svc-other"}, 100)
	route := testDNLRouteForNode(t, "apps", "svc-other", other.NodeID, 1, 8_000, 100)
	table := testRoutingTableForRoutes(t, 9, route)
	_, err = UpdateRoutingTableInRoutingState(state, MsgUpdateRoutingTable{
		Authority:	node.OperatorAddress,
		Table:		table,
		Height:		21,
	})
	require.ErrorContains(t, err, "registered node")
}
