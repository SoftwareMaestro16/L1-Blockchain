package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/avm-scheduler/types"
	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestDefaultGenesisIsDisabledAndValid(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.ExecutionQueue)
	require.NotZero(t, gs.SchedulerParams.MaxParallelism)
}

func TestNonConflictingContractsExecuteInParallel(t *testing.T) {
	k := enabledKeeper(t)
	batch := testBatch("parallel", task("a", "contract-a", "mailbox-a", nil, []string{"state/a"}), task("b", "contract-b", "mailbox-b", nil, []string{"state/b"}))

	require.NoError(t, k.SubmitAVMExecutionBatch(submit(batch)))
	graph, found, err := k.AVMDependencyGraph("parallel")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, graph.ParallelGroups, 1)
	require.Equal(t, []string{"a", "b"}, graph.ParallelGroups[0].TaskIDs)
	require.False(t, graph.FallbackSerial)

	receipts, err := k.FinalizeAVMExecutionBatch(finalize("parallel", 10))
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	require.Equal(t, []uint32{0, 1}, []uint32{receipts[0].Order, receipts[1].Order})
}

func TestConflictingContractsSerialize(t *testing.T) {
	k := enabledKeeper(t)
	batch := testBatch("conflict",
		task("a", "contract-a", "mailbox-a", nil, []string{"shared"}),
		task("b", "contract-b", "mailbox-b", []string{"shared"}, []string{"other"}),
	)

	require.NoError(t, k.SubmitAVMExecutionBatch(submit(batch)))
	graph, found, err := k.AVMDependencyGraph("conflict")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, graph.ParallelGroups, 2)
	require.Equal(t, []string{"a"}, graph.ParallelGroups[0].TaskIDs)
	require.Equal(t, []string{"b"}, graph.ParallelGroups[1].TaskIDs)
	require.Contains(t, graph.Nodes[1].Dependencies, "a")
}

func TestDeterministicReceiptOrderAndParallelSerialRootsMatch(t *testing.T) {
	k := enabledKeeper(t)
	batch := testBatch("roots", task("b", "contract-b", "mailbox-b", nil, []string{"state/b"}), task("a", "contract-a", "mailbox-a", nil, []string{"state/a"}))
	require.NoError(t, k.SubmitAVMExecutionBatch(submit(batch)))

	receipts, err := k.FinalizeAVMExecutionBatch(finalize("roots", 11))
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, []string{receipts[0].TaskID, receipts[1].TaskID})
	serialRoot, err := types.ComputeSerialStateRoot(batch, nil)
	require.NoError(t, err)
	require.Equal(t, serialRoot, receipts[len(receipts)-1].StateRootAfter)
}

func TestMalformedReadWriteSetRejected(t *testing.T) {
	k := enabledKeeper(t)
	bad := testBatch("bad", task("a", "contract-a", "mailbox-a", []string{"state/a", "state/a"}, []string{"state/b"}))

	err := k.SubmitAVMExecutionBatch(submit(bad))
	require.ErrorContains(t, err, "duplicate")

	bad = testBatch("bad-empty", task("a", "contract-a", "mailbox-a", []string{" state/a"}, []string{"state/b"}))
	err = k.SubmitAVMExecutionBatch(submit(bad))
	require.ErrorContains(t, err, "non-canonical")
}

func TestFailedContractExecutionDoesNotCorruptUnrelatedState(t *testing.T) {
	k := enabledKeeper(t)
	batch := testBatch("failure", task("a", "contract-a", "mailbox-a", nil, []string{"state/a"}), task("b", "contract-b", "mailbox-b", nil, []string{"state/b"}))
	require.NoError(t, k.SubmitAVMExecutionBatch(submit(batch)))

	msg := finalize("failure", 12)
	msg.FailedTaskIDs = []string{"a"}
	receipts, err := k.FinalizeAVMExecutionBatch(msg)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusFailure, receipts[0].Status)
	require.Equal(t, receipts[0].StateRootBefore, receipts[0].StateRootAfter)
	require.Equal(t, types.ReceiptStatusSuccess, receipts[1].Status)
	require.NotEqual(t, receipts[1].StateRootBefore, receipts[1].StateRootAfter)
	expectedRoot, err := types.ComputeSerialStateRoot(batch, []string{"a"})
	require.NoError(t, err)
	require.Equal(t, expectedRoot, receipts[1].StateRootAfter)
}

func TestExportImportPreservesQueueAndReceipts(t *testing.T) {
	source := enabledKeeper(t)
	queued := testBatch("queued", task("q", "contract-q", "mailbox-q", nil, []string{"state/q"}))
	done := testBatch("done", task("d", "contract-d", "mailbox-d", nil, []string{"state/d"}))
	require.NoError(t, source.SubmitAVMExecutionBatch(submit(queued)))
	require.NoError(t, source.SubmitAVMExecutionBatch(submit(done)))
	_, err := source.FinalizeAVMExecutionBatch(finalize("done", 20))
	require.NoError(t, err)

	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, source.InitGenesisState(ctx, gs))

	batch := testBatch("persistent-done", task("d", "contract-d", "mailbox-d", nil, []string{"state/d"}))
	require.NoError(t, source.SubmitAVMExecutionBatch(submit(batch)))
	receipts, err := source.FinalizeAVMExecutionBatch(finalize("persistent-done", 20))
	require.NoError(t, err)
	require.Len(t, receipts, 1)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Empty(t, exported.State.ExecutionQueue)
	require.Len(t, exported.State.ExecutionReceipts, 1)
	require.Equal(t, receipts[0].ReceiptID, exported.State.ExecutionReceipts[0].ReceiptID)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	restored, found, err := imported.AVMExecutionReceipt("persistent-done", "d")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, receipts[0], restored)
}

func enabledKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func submit(batch types.AVMExecutionBatch) types.MsgSubmitAVMExecutionBatch {
	return types.MsgSubmitAVMExecutionBatch{Authority: prototype.DefaultAuthority, Batch: batch}
}

func finalize(batchID string, height uint64) types.MsgFinalizeAVMExecutionBatch {
	return types.MsgFinalizeAVMExecutionBatch{Authority: prototype.DefaultAuthority, BatchID: batchID, Height: height}
}

func testBatch(id string, tasks ...types.AVMExecutionTask) types.AVMExecutionBatch {
	return types.AVMExecutionBatch{
		BatchID:		id,
		SubmittedHeight:	1,
		Tasks:			tasks,
	}
}

func task(id, contract, mailbox string, reads []string, writes []string) types.AVMExecutionTask {
	stateWrites := make([]types.AVMStateWrite, 0, len(writes))
	for _, key := range writes {
		stateWrites = append(stateWrites, types.AVMStateWrite{Key: key, Value: id + "-value"})
	}
	return types.AVMExecutionTask{
		TaskID:			id,
		ContractAddress:	contract,
		Mailbox:		mailbox,
		ReadSet:		reads,
		WriteSet:		writes,
		StateWrites:		stateWrites,
		GasLimit:		10,
	}
}
