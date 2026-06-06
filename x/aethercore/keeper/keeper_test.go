package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aethercore/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestDefaultGenesisValidates(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.ZoneDescriptors)
}

func TestFeatureDisabledRejectsMutatingMessages(t *testing.T) {
	keeper := NewKeeper()

	err := keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDFinancial, types.ZoneTypeFinancial, "financial"))
	require.ErrorContains(t, err, "disabled")
}

func TestRegisterCommitAndQueryRoots(t *testing.T) {
	keeper := NewKeeper()
	params := types.TestnetParams()
	params.DefaultQueryLimit = 1
	params.MaxQueryLimit = 2
	require.NoError(t, keeper.UpdateParams(types.DefaultAuthority, params))

	require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDFinancial, types.ZoneTypeFinancial, "financial")))
	require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDContract, types.ZoneTypeContract, "contract")))
	require.NoError(t, keeper.AppendZoneCommitment(keeperCommitment(t, 3, types.ZoneIDFinancial)))
	require.NoError(t, keeper.AppendZoneCommitment(keeperCommitment(t, 3, types.ZoneIDContract)))

	root, err := keeper.CommitGlobalRoot(3, keeperContributions(3))
	require.NoError(t, err)
	require.NoError(t, root.ValidateHash())

	roots, page, err := keeper.GlobalRoots(nil)
	require.NoError(t, err)
	require.Len(t, roots, 1)
	require.Zero(t, page.NextOffset)

	zones, page, err := keeper.ZoneDescriptors(nil)
	require.NoError(t, err)
	require.Len(t, zones, 1)
	require.NotZero(t, page.NextOffset)

	zones, page, err = keeper.ZoneDescriptors(&prototype.PageRequest{Offset: page.NextOffset, Limit: 2})
	require.NoError(t, err)
	require.Len(t, zones, 1)
	require.Zero(t, page.NextOffset)
}

func TestKeeperReplayIdenticalRootsAcrossNodes(t *testing.T) {
	nodeA := keeperWithState(t, []types.ZoneID{types.ZoneIDFinancial, types.ZoneIDContract})
	nodeB := keeperWithState(t, []types.ZoneID{types.ZoneIDContract, types.ZoneIDFinancial})

	rootA, err := nodeA.CommitGlobalRoot(5, keeperContributions(5))
	require.NoError(t, err)
	rootB, err := nodeB.CommitGlobalRoot(5, keeperContributions(5))
	require.NoError(t, err)

	require.Equal(t, rootA, rootB)
	require.Equal(t, nodeA.ExportGenesis(), nodeB.ExportGenesis())
}

func TestBuildProposalScheduleChecksRegisteredZones(t *testing.T) {
	keeper := NewKeeper()
	require.NoError(t, keeper.UpdateParams(types.DefaultAuthority, types.TestnetParams()))
	require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDFinancial, types.ZoneTypeFinancial, "financial")))
	require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDContract, types.ZoneTypeContract, "contract")))

	_, err := keeper.BuildProposalSchedule(8, []types.ProposalItem{
		keeperProposalItem(types.ZoneIDFinancial, "1", "financial-b", 2, 8, 1),
	})
	require.ErrorContains(t, err, "missing shard layout")

	require.NoError(t, keeper.RegisterShardLayout(keeperLayout(t, types.ZoneIDFinancial, 1, []types.ShardID{"1"})))
	require.NoError(t, keeper.RegisterShardLayout(keeperLayout(t, types.ZoneIDContract, 1, []types.ShardID{"0"})))
	schedule, err := keeper.BuildProposalSchedule(8, []types.ProposalItem{
		keeperProposalItem(types.ZoneIDFinancial, "1", "financial-b", 2, 8, 1),
		keeperProposalItem(types.ZoneIDContract, "0", "contract-a", 1, 8, 0),
		keeperProposalItem(types.ZoneIDFinancial, "1", "financial-a", 1, 7, 2),
	})
	require.NoError(t, err)
	require.Len(t, schedule.Groups, 2)
	require.Equal(t, types.ZoneID(types.ZoneIDContract), schedule.Groups[0].ZoneID)
	require.Equal(t, types.ZoneID(types.ZoneIDFinancial), schedule.Groups[1].ZoneID)
	require.Equal(t, keeperHash("financial-a"), schedule.Groups[1].Items[0].TxHash)

	_, err = keeper.BuildProposalSchedule(8, []types.ProposalItem{
		keeperProposalItem(types.ZoneIDIdentity, "0", "identity-a", 1, 8, 0),
	})
	require.ErrorContains(t, err, "not registered")
}

func TestExportImportRoundTripDeterministic(t *testing.T) {
	source := keeperWithState(t, []types.ZoneID{types.ZoneIDFinancial, types.ZoneIDContract})
	_, err := source.CommitGlobalRoot(5, keeperContributions(5))
	require.NoError(t, err)

	exported := source.ExportGenesis()
	imported := NewKeeper()
	require.NoError(t, imported.InitGenesis(exported))
	require.Equal(t, exported, imported.ExportGenesis())
}

func keeperProposalItem(zoneID types.ZoneID, shardID types.ShardID, seed string, priority uint32, height uint64, txIndex uint32) types.ProposalItem {
	return types.ProposalItem{
		ZoneID:          zoneID,
		ShardID:         shardID,
		TxHash:          keeperHash(seed),
		PriorityClass:   priority,
		AdmissionHeight: height,
		TxIndex:         txIndex,
	}
}

func keeperWithState(t *testing.T, order []types.ZoneID) Keeper {
	t.Helper()
	keeper := NewKeeper()
	require.NoError(t, keeper.UpdateParams(types.DefaultAuthority, types.TestnetParams()))
	for _, zoneID := range order {
		switch zoneID {
		case types.ZoneIDFinancial:
			require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDFinancial, types.ZoneTypeFinancial, "financial")))
		case types.ZoneIDContract:
			require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDContract, types.ZoneTypeContract, "contract")))
		default:
			t.Fatalf("unsupported zone %s", zoneID)
		}
		require.NoError(t, keeper.AppendZoneCommitment(keeperCommitment(t, 5, zoneID)))
	}
	return keeper
}

func keeperZone(zoneID types.ZoneID, zoneType types.ZoneType, moduleName string) types.ZoneDescriptor {
	return types.ZoneDescriptor{
		ZoneID:              zoneID,
		ZoneType:            zoneType,
		ModuleName:          moduleName,
		Enabled:             true,
		StateMachineVersion: 1,
		MempoolPolicyID:     types.DefaultMempoolPolicy,
		FeePolicyID:         types.NativeFeePolicyID,
		ShardLayoutEpoch:    1,
		MaxShards:           4,
		MessageCapabilities: []string{"async-inbox", "async-outbox"},
		ProofCapabilities:   []string{"account", "message", "receipt"},
	}
}

func keeperCommitment(t *testing.T, height uint64, zoneID types.ZoneID) types.ZoneCommitment {
	t.Helper()
	commitment, err := types.NewZoneCommitment(
		height,
		zoneID,
		keeperHash(fmt.Sprintf("%d/%s/state", height, zoneID)),
		keeperHash(fmt.Sprintf("%d/%s/inbox", height, zoneID)),
		keeperHash(fmt.Sprintf("%d/%s/outbox", height, zoneID)),
		keeperHash(fmt.Sprintf("%d/%s/receipts", height, zoneID)),
		keeperHash(fmt.Sprintf("%d/%s/events", height, zoneID)),
		keeperHash(fmt.Sprintf("%d/%s/params", height, zoneID)),
		keeperHash(fmt.Sprintf("%d/%s/execution", height, zoneID)),
	)
	require.NoError(t, err)
	return commitment
}

func keeperLayout(t *testing.T, zoneID types.ZoneID, epoch uint64, shardIDs []types.ShardID) types.ShardLayout {
	t.Helper()
	shards := make([]types.ShardDescriptor, len(shardIDs))
	for i, shardID := range shardIDs {
		shards[i] = types.ShardDescriptor{
			ShardID:          shardID,
			StatePrefix:      fmt.Sprintf("zone/%s/shard/%s", zoneID, shardID),
			ActivationHeight: 1,
			ValidatorSetHash: keeperHash(fmt.Sprintf("%s/%s/validators", zoneID, shardID)),
			Available:        true,
		}
	}
	layout, err := types.NewShardLayout(zoneID, epoch, 1, keeperHash(fmt.Sprintf("%s/%d/routing-seed", zoneID, epoch)), shards)
	require.NoError(t, err)
	return layout
}

func keeperContributions(height uint64) types.RootContributions {
	return types.RootContributions{
		IdentityRoot: keeperHash(fmt.Sprintf("%d/identity", height)),
		StorageRoot:  keeperHash(fmt.Sprintf("%d/storage", height)),
		MessageRoot:  keeperHash(fmt.Sprintf("%d/messages", height)),
		ReceiptsRoot: keeperHash(fmt.Sprintf("%d/receipts", height)),
		PaymentsRoot: keeperHash(fmt.Sprintf("%d/payments", height)),
		VMRoot:       keeperHash(fmt.Sprintf("%d/vm", height)),
		ParamsHash:   keeperHash(fmt.Sprintf("%d/params", height)),
	}
}

func keeperHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
