package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniversalProofObjectivesBindToExpectedRoots(t *testing.T) {
	objectives := SupportedUniversalProofObjectives()
	require.Len(t, objectives, 10)

	seen := make(map[UniversalProofObjective]struct{}, len(objectives))
	for _, objective := range objectives {
		seen[objective.Objective] = struct{}{}
		root := ProofRoot{
			Height:		9,
			ZoneID:		objective.ZoneID,
			RootType:	objective.RootType,
			RootHash:	testHash(string(objective.Objective)),
			Source:		"aetracore.universal_proofs",
		}
		require.NoError(t, ValidateProofRootForObjective(objective.Objective, root))

		wrong := root
		wrong.RootType = StateProofRootType
		if objective.RootType == StateProofRootType {
			wrong.RootType = MessageProofRootType
		}
		require.ErrorContains(t, ValidateProofRootForObjective(objective.Objective, wrong), "requires root type")
	}

	for _, objective := range []UniversalProofObjective{
		ProofObjectiveAccountState,
		ProofObjectiveBalanceState,
		ProofObjectiveMessageInclusion,
		ProofObjectiveMessageReceipt,
		ProofObjectiveZoneStateRoot,
		ProofObjectiveShardStateRoot,
		ProofObjectiveDomainOwnership,
		ProofObjectiveResolverRecords,
		ProofObjectiveContractState,
		ProofObjectivePaymentSettlement,
	} {
		_, found := seen[objective]
		require.True(t, found, "missing objective %s", objective)
	}
}

func TestUniversalProofObjectiveRejectsWrongZone(t *testing.T) {
	root := ProofRoot{
		Height:		9,
		ZoneID:		ZoneIDApplication,
		RootType:	BalanceProofRootType,
		RootHash:	testHash("balance"),
		Source:		"aetracore.universal_proofs",
	}

	require.ErrorContains(t, ValidateProofRootForObjective(ProofObjectiveBalanceState, root), "requires zone FINANCIAL_ZONE")
}

func TestUniversalRootHierarchyCanonicalizesZonesAndShards(t *testing.T) {
	height := uint64(12)
	financial := testUniversalZoneBranch(t, height, ZoneIDFinancial, "financial", []ShardID{"2", "0", "1"})
	identity := testUniversalZoneBranch(t, height, ZoneIDIdentity, "identity", []ShardID{"1", "0"})
	application := testUniversalZoneBranch(t, height, ZoneIDApplication, "application", []ShardID{"0"})
	contract := testUniversalZoneBranch(t, height, ZoneIDContract, "contract", []ShardID{"3", "1"})
	messages := testUniversalMessageBranch(t, height)

	a, err := NewUniversalRootHierarchy(
		height,
		testHash("aether-core"),
		[]UniversalZoneRootBranch{contract, financial, application, identity},
		messages,
	)
	require.NoError(t, err)
	require.NoError(t, a.Validate())
	require.NoError(t, a.ValidateRequiredZones(RequiredAetraNextProofZones()))

	b, err := NewUniversalRootHierarchy(
		height,
		testHash("aether-core"),
		[]UniversalZoneRootBranch{identity, application, financial, contract},
		messages,
	)
	require.NoError(t, err)

	require.Equal(t, a.GlobalZoneRoot, b.GlobalZoneRoot)
	require.Equal(t, a.GlobalMessageRoot, b.GlobalMessageRoot)
	require.Equal(t, a.AppHash, b.AppHash)
}

func TestUniversalRootHierarchyRejectsRootDrift(t *testing.T) {
	height := uint64(12)
	hierarchy, err := NewUniversalRootHierarchy(
		height,
		testHash("aether-core"),
		[]UniversalZoneRootBranch{
			testUniversalZoneBranch(t, height, ZoneIDFinancial, "financial", []ShardID{"0"}),
			testUniversalZoneBranch(t, height, ZoneIDIdentity, "identity", []ShardID{"0"}),
			testUniversalZoneBranch(t, height, ZoneIDApplication, "application", []ShardID{"0"}),
			testUniversalZoneBranch(t, height, ZoneIDContract, "contract", []ShardID{"0"}),
		},
		testUniversalMessageBranch(t, height),
	)
	require.NoError(t, err)

	drifted := hierarchy
	drifted.Zones = append([]UniversalZoneRootBranch(nil), hierarchy.Zones...)
	drifted.Zones[0].ShardRootsRoot = testHash("drifted-shard-root")
	require.ErrorContains(t, drifted.Validate(), "shard roots root mismatch")

	missingZone := hierarchy
	missingZone.Zones = missingZone.Zones[:3]
	missingZone.GlobalZoneRoot, err = ComputeUniversalGlobalZoneRoot(height, missingZone.Zones)
	require.NoError(t, err)
	missingZone.AppHash = ComputeUniversalAppHash(missingZone)
	require.ErrorContains(t, missingZone.ValidateRequiredZones(RequiredAetraNextProofZones()), "missing required zone")
}

func TestUniversalMessageRootBindsOutboxInboxAndReceipts(t *testing.T) {
	height := uint64(15)
	base := testUniversalMessageBranch(t, height)
	changedReceipt, err := NewUniversalMessageRootBranch(height, base.ZoneOutboxRoot, base.ZoneInboxRoot, testHash("changed-receipt"))
	require.NoError(t, err)

	require.NotEqual(t, base.MessageRoot, changedReceipt.MessageRoot)
	require.NoError(t, changedReceipt.Validate(height))

	tampered := base
	tampered.ReceiptRoot = testHash("changed-receipt")
	require.ErrorContains(t, tampered.Validate(height), "message root mismatch")
}

func testUniversalZoneBranch(t *testing.T, height uint64, zoneID ZoneID, label string, shardIDs []ShardID) UniversalZoneRootBranch {
	t.Helper()
	shards := make([]UniversalShardRootBranch, 0, len(shardIDs))
	for _, shardID := range shardIDs {
		shards = append(shards, UniversalShardRootBranch{
			ZoneID:		zoneID,
			ShardID:	shardID,
			ShardRoot:	testHash(label + "-shard-" + string(shardID)),
		})
	}
	branch, err := NewUniversalZoneRootBranch(height, zoneID, testHash(label+"-zone"), shards)
	require.NoError(t, err)
	return branch
}

func testUniversalMessageBranch(t *testing.T, height uint64) UniversalMessageRootBranch {
	t.Helper()
	branch, err := NewUniversalMessageRootBranch(
		height,
		testHash("zone-outbox-root"),
		testHash("zone-inbox-root"),
		testHash("receipt-root"),
	)
	require.NoError(t, err)
	return branch
}
