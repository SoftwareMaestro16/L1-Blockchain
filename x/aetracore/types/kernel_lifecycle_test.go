package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKernelLifecycleFinalizesCommittedRootsAndReceipts(t *testing.T) {
	state := EmptyState(TestnetParams())
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDContract, ZoneTypeContract, "contract"))
	require.NoError(t, err)
	state, err = RegisterShardLayout(state, testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"0"}))
	require.NoError(t, err)
	state, err = RegisterShardLayout(state, testShardLayout(t, ZoneIDContract, 1, []ShardID{"0"}))
	require.NoError(t, err)

	ctx := KernelConsensusContext{ChainID: "aetra-testnet", Height: 12, BlockTimeUnix: 1_700_000_000}
	left := testProposalItem(ZoneIDFinancial, "0", "left", 2, 11, 1)
	right := testProposalItem(ZoneIDContract, "0", "right", 1, 11, 0)
	planA, err := PrepareKernelProposal(ctx, state, []ProposalItem{left, right})
	require.NoError(t, err)
	planB, err := PrepareKernelProposal(ctx, state, []ProposalItem{right, left})
	require.NoError(t, err)
	require.Equal(t, planA, planB)
	require.NoError(t, ProcessKernelProposal(ctx, state, planA))

	classified, err := ClassifyTransaction(state, ClassificationInput{
		Height:			ctx.Height,
		TxHash:			testHash("kernel-sync-tx"),
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDFinancial,
		DestinationShard:	"0",
		AdmissionHeight:	11,
	})
	require.NoError(t, err)
	receipt, err := ExecuteSync(classified, ExecutionResult{Success: true, ResultHash: testHash("kernel-sync-result")}, ctx.Height, 0)
	require.NoError(t, err)
	receiptsRoot, err := ComputeExecutionReceiptsRoot([]ExecutionReceipt{receipt})
	require.NoError(t, err)
	contributions := testContributions(ctx.Height)
	contributions.ReceiptsRoot = receiptsRoot

	next, finalization, err := FinalizeKernelBlock(ctx, state, planA, KernelFinalizationInput{
		ZoneCommitments: []ZoneCommitment{
			testCommitment(t, ctx.Height, ZoneIDContract),
			testCommitment(t, ctx.Height, ZoneIDFinancial),
		},
		Receipts:	[]ExecutionReceipt{receipt},
		Contributions:	contributions,
	})
	require.NoError(t, err)
	require.NoError(t, finalization.Validate())
	require.Equal(t, ctx.BlockTimeUnix, finalization.Header.TimeUnix)
	require.Equal(t, planA.PreviousGlobalRoot, finalization.Header.PreviousAppHash)
	require.Equal(t, finalization.GlobalRoot.ZonesRoot, finalization.Header.ZonesRoot)
	require.Equal(t, finalization.GlobalRoot.GlobalRoot, finalization.RootSnapshot.Finality.GlobalStateRoot)
	require.Equal(t, receiptsRoot, finalization.RootSnapshot.Finality.ExecutionReceiptRoot)
	require.Len(t, next.GlobalRoots, 1)
	require.Len(t, next.RootSnapshots, 1)

	finality, err := CommitKernelBlock(finalization)
	require.NoError(t, err)
	require.Equal(t, finalization.GlobalRoot.GlobalRoot, finality.GlobalStateRoot)

	manifest, err := BuildKernelExportManifest(next, ctx.Height, testHash("apphash"))
	require.NoError(t, err)
	require.NoError(t, ValidateKernelImport(next, manifest))
}

func TestKernelLifecycleRejectsReceiptRootMismatchAndDuplicateReceipts(t *testing.T) {
	state := EmptyState(TestnetParams())
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	state, err = RegisterShardLayout(state, testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"0"}))
	require.NoError(t, err)

	ctx := KernelConsensusContext{ChainID: "aetra-testnet", Height: 13, BlockTimeUnix: 1_700_000_001}
	plan, err := PrepareKernelProposal(ctx, state, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "0", "single", 1, 12, 0),
	})
	require.NoError(t, err)
	classified, err := ClassifyTransaction(state, ClassificationInput{
		Height:			ctx.Height,
		TxHash:			testHash("duplicate-receipt-tx"),
		SourceZone:		ZoneIDFinancial,
		SourceShard:		"0",
		DestinationZone:	ZoneIDFinancial,
		DestinationShard:	"0",
		AdmissionHeight:	12,
	})
	require.NoError(t, err)
	receipt, err := ExecuteSync(classified, ExecutionResult{Success: true, ResultHash: testHash("duplicate-receipt-result")}, ctx.Height, 0)
	require.NoError(t, err)
	_, err = ComputeExecutionReceiptsRoot([]ExecutionReceipt{receipt, receipt})
	require.ErrorContains(t, err, "duplicate")

	contributions := testContributions(ctx.Height)
	contributions.ReceiptsRoot = testHash("wrong-receipts-root")
	_, _, err = FinalizeKernelBlock(ctx, state, plan, KernelFinalizationInput{
		ZoneCommitments:	[]ZoneCommitment{testCommitment(t, ctx.Height, ZoneIDFinancial)},
		Receipts:		[]ExecutionReceipt{receipt},
		Contributions:		contributions,
	})
	require.ErrorContains(t, err, "receipts root contribution mismatch")
}

func TestAetherKernelStateKeysMatchSpecification(t *testing.T) {
	require.Equal(t, AetherKernelParamsKey, "aek/params")
	require.Equal(t, CoreParamsKey, "core/params")

	zoneKey, err := AetherKernelZoneKey(ZoneIDFinancial)
	require.NoError(t, err)
	require.Equal(t, "aek/zones/FINANCIAL_ZONE", zoneKey)
	coreZoneKey, err := CoreZoneKey(ZoneIDFinancial)
	require.NoError(t, err)
	require.Equal(t, "core/zones/FINANCIAL_ZONE", coreZoneKey)

	commitmentKey, err := AetherKernelZoneCommitmentKey(7, ZoneIDFinancial)
	require.NoError(t, err)
	require.Equal(t, "aek/zone_commitments/00000000000000000007/FINANCIAL_ZONE", commitmentKey)
	coreZoneRootKey, err := CoreZoneRootKey(7, ZoneIDFinancial)
	require.NoError(t, err)
	require.Equal(t, "core/zone_roots/00000000000000000007/FINANCIAL_ZONE", coreZoneRootKey)

	for _, tc := range []struct {
		name	string
		fn	func(uint64) (string, error)
		want	string
	}{
		{name: "messages", fn: AetherKernelMessageRootKey, want: "aek/messages/root/00000000000000000007"},
		{name: "receipts", fn: AetherKernelReceiptsRootKey, want: "aek/receipts/root/00000000000000000007"},
		{name: "services", fn: AetherKernelServicesRootKey, want: "aek/services/root/00000000000000000007"},
		{name: "identity", fn: AetherKernelIdentityRootKey, want: "aek/identity/root/00000000000000000007"},
		{name: "storage", fn: AetherKernelStorageRootKey, want: "aek/storage/root/00000000000000000007"},
		{name: "export", fn: AetherKernelExportKey, want: "aek/export/00000000000000000007"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.fn(7)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}

	routingKey, err := AetherKernelRoutingTableKey(3)
	require.NoError(t, err)
	require.Equal(t, "aek/routing/table/00000000000000000003", routingKey)

	coreMessageKey, err := CoreMessageRootKey(7)
	require.NoError(t, err)
	require.Equal(t, "core/message_roots/00000000000000000007", coreMessageKey)
	coreShardLayoutKey, err := CoreShardLayoutKey(ZoneIDFinancial, 2)
	require.NoError(t, err)
	require.Equal(t, "core/shard_layouts/FINANCIAL_ZONE/00000000000000000002", coreShardLayoutKey)
	coreRoutingKey, err := CoreRoutingTableKey(3)
	require.NoError(t, err)
	require.Equal(t, "core/routing_table/00000000000000000003", coreRoutingKey)
	coreProofKey, err := CoreProofRootKey(7, MessageProofRootType)
	require.NoError(t, err)
	require.Equal(t, "core/proof_roots/00000000000000000007/messages", coreProofKey)
	coreFinalityKey, err := CoreFinalityKey(7)
	require.NoError(t, err)
	require.Equal(t, "core/finality/00000000000000000007", coreFinalityKey)
}
