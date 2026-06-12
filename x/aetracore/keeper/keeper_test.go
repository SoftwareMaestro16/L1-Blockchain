package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aetracore/types"
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

func TestKeeperABCILifecycleHooksAndInvariants(t *testing.T) {
	keeper := NewKeeper()
	require.NoError(t, keeper.UpdateParams(types.DefaultAuthority, types.TestnetParams()))
	require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDFinancial, types.ZoneTypeFinancial, "financial")))
	require.NoError(t, keeper.RegisterZoneDescriptor(keeperZone(types.ZoneIDContract, types.ZoneTypeContract, "contract")))
	require.NoError(t, keeper.RegisterShardLayout(keeperLayout(t, types.ZoneIDFinancial, 1, []types.ShardID{"0"})))
	require.NoError(t, keeper.RegisterShardLayout(keeperLayout(t, types.ZoneIDContract, 1, []types.ShardID{"0"})))
	require.NoError(t, keeper.AppendZoneCommitment(keeperCommitment(t, 10, types.ZoneIDFinancial)))
	_, err := keeper.CommitBlockRoots(10)
	require.NoError(t, err)

	ctx := types.KernelConsensusContext{ChainID: "aetra-testnet", Height: 11, BlockTimeUnix: 1_700_000_011}
	envelope := types.KernelMessageEnvelope{
		Kind:			types.KernelMessageLocalTx,
		TxHash:			keeperHash("keeper-abci-local"),
		SourceZone:		types.ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	types.ZoneIDFinancial,
		DestinationShard:	"0",
		Sender:			"keeper.sender",
		Nonce:			1,
		GasLimit:		100,
		PriorityClass:		1,
		AdmissionHeight:	11,
	}
	proposal, err := keeper.PrepareKernelABCIProposal(ctx, []types.KernelMessageEnvelope{envelope}, nil, types.KernelGasLimits{MaxBlockGas: 1_000, MaxZoneGas: 500})
	require.NoError(t, err)
	require.NoError(t, keeper.ProcessKernelABCIProposal(ctx, proposal, []types.KernelMessageEnvelope{envelope}, types.KernelGasLimits{MaxBlockGas: 1_000, MaxZoneGas: 500}))

	classified, err := types.ClassifyTransaction(keeper.ExportGenesis().State, types.ClassificationInput{
		Height:			11,
		TxHash:			envelope.TxHash,
		SourceZone:		types.ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	types.ZoneIDFinancial,
		DestinationShard:	"0",
		AdmissionHeight:	11,
	})
	require.NoError(t, err)
	receipt, err := types.ExecuteSync(classified, types.ExecutionResult{Success: true, ResultHash: keeperHash("keeper-abci-result")}, 11, 1)
	require.NoError(t, err)
	receiptsRoot, err := types.ComputeExecutionReceiptsRoot([]types.ExecutionReceipt{receipt})
	require.NoError(t, err)
	contributions := keeperContributions(11)
	contributions.ReceiptsRoot = receiptsRoot
	finalization, cleanup, err := keeper.FinalizeKernelABCIBlock(ctx, proposal, []types.KernelMessageEnvelope{envelope}, types.KernelFinalizationInput{
		ZoneCommitments: []types.ZoneCommitment{
			keeperCommitment(t, 11, types.ZoneIDFinancial),
			keeperCommitment(t, 11, types.ZoneIDContract),
		},
		Receipts:	[]types.ExecutionReceipt{receipt},
		Contributions:	contributions,
	}, []types.KernelCleanupItem{{QueueID: "receipts", ItemID: "old", HeightDue: 11, DeleteRoot: keeperHash("cleanup")}}, 1)
	require.NoError(t, err)
	require.Len(t, cleanup.Processed, 1)
	record, err := keeper.CommitKernelABCIBlock(finalization, keeperHash("keeper-abci-apphash"))
	require.NoError(t, err)
	require.Equal(t, finalization.GlobalRoot.GlobalRoot, record.GlobalRoot)
	require.NoError(t, keeper.ValidateRootAggregationInvariants())

	summary, err := keeper.CollectZoneExecutionSummary(11, types.ZoneIDFinancial, []types.KernelMessageEnvelope{envelope}, []types.ExecutionReceipt{receipt}, 100, keeperHash("events"))
	require.NoError(t, err)
	require.Equal(t, uint64(1), summary.LocalTxCount)
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
		ZoneID:			zoneID,
		ShardID:		shardID,
		TxHash:			keeperHash(seed),
		PriorityClass:		priority,
		AdmissionHeight:	height,
		TxIndex:		txIndex,
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
		ZoneID:			zoneID,
		ZoneType:		zoneType,
		ModuleName:		moduleName,
		Enabled:		true,
		StateMachineVersion:	1,
		MempoolPolicyID:	types.DefaultMempoolPolicy,
		FeePolicyID:		types.NativeFeePolicyID,
		ShardLayoutEpoch:	1,
		MaxShards:		4,
		MessageCapabilities:	[]string{"async-inbox", "async-outbox"},
		ProofCapabilities:	[]string{"account", "message", "receipt"},
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
		keeperHash(fmt.Sprintf("%d/%s/shards", height, zoneID)),
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
			ShardID:		shardID,
			StatePrefix:		fmt.Sprintf("zone/%s/shard/%s", zoneID, shardID),
			ActivationHeight:	1,
			ValidatorSetHash:	keeperHash(fmt.Sprintf("%s/%s/validators", zoneID, shardID)),
			Available:		true,
		}
	}
	layout, err := types.NewShardLayout(zoneID, epoch, 1, keeperHash(fmt.Sprintf("%s/%d/routing-seed", zoneID, epoch)), shards)
	require.NoError(t, err)
	return layout
}

func keeperContributions(height uint64) types.RootContributions {
	return types.RootContributions{
		IdentityRoot:	keeperHash(fmt.Sprintf("%d/identity", height)),
		StorageRoot:	keeperHash(fmt.Sprintf("%d/storage", height)),
		MessageRoot:	keeperHash(fmt.Sprintf("%d/messages", height)),
		ReceiptsRoot:	keeperHash(fmt.Sprintf("%d/receipts", height)),
		RoutingRoot:	keeperHash(fmt.Sprintf("%d/routing", height)),
		PaymentsRoot:	keeperHash(fmt.Sprintf("%d/payments", height)),
		ContractsRoot:	keeperHash(fmt.Sprintf("%d/contracts", height)),
		VMRoot:		keeperHash(fmt.Sprintf("%d/vm", height)),
		ParamsHash:	keeperHash(fmt.Sprintf("%d/params", height)),
	}
}

func keeperHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
