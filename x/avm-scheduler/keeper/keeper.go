package keeper

import (
	"context"
	"errors"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/avm-scheduler/types"
	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	SchedulerParams	types.AVMSchedulerParams
	State		types.AVMSchedulerState
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
	runtimeCtx	context.Context
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version:		prototype.CurrentGenesisVersion,
		Params:			prototype.DefaultParams(),
		SchedulerParams:	types.DefaultAVMSchedulerParams(),
		State:			types.EmptyAVMSchedulerState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("AVM scheduler prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.SchedulerParams)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	k.runtimeCtx = ctx
	if k.storeService == nil {
		return nil
	}
	return prefixgenesis.Save(ctx, k.storeService, genesisKey, k.genesis)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	gs, _, err := prefixgenesis.Load(ctx, k.storeService, genesisKey, DefaultGenesis())
	if err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) UpdateAVMSchedulerParams(msg types.MsgUpdateAVMSchedulerParams) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return err
	}
	if err := k.genesis.State.Validate(msg.SchedulerParams); err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	next.SchedulerParams = msg.SchedulerParams
	return k.saveGenesis(next)
}

func (k *Keeper) SubmitAVMExecutionBatch(msg types.MsgSubmitAVMExecutionBatch) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return err
	}
	if err := msg.Batch.Validate(k.genesis.SchedulerParams); err != nil {
		return err
	}
	batch := msg.Batch.Normalize()
	if _, _, found := k.findQueuedBatch(batch.BatchID); found {
		return errors.New("AVM execution batch already queued")
	}
	if _, found := k.findGraph(batch.BatchID); found {
		return errors.New("AVM execution batch graph already exists")
	}
	graph, err := types.BuildDependencyGraph(batch, k.genesis.SchedulerParams, false)
	if err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	next.State.ExecutionQueue = append(next.State.ExecutionQueue, batch)
	next.State.DependencyGraphs = append(next.State.DependencyGraphs, graph)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k *Keeper) FinalizeAVMExecutionBatch(msg types.MsgFinalizeAVMExecutionBatch) ([]types.AVMExecutionReceipt, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return nil, err
	}
	if msg.Height == 0 {
		return nil, errors.New("AVM scheduler finalize height must be positive")
	}
	index, batch, found := k.findQueuedBatch(msg.BatchID)
	if !found {
		return nil, errors.New("AVM execution batch not found")
	}
	if uint32(len(batch.Tasks)) > k.genesis.SchedulerParams.MaxExecutionsPerBlock {
		return nil, errors.New("AVM execution batch exceeds per-block execution limit")
	}
	graph, found := k.findGraph(batch.BatchID)
	if !found || msg.ForceSerial {
		var err error
		graph, err = types.BuildDependencyGraph(batch, k.genesis.SchedulerParams, msg.ForceSerial)
		if err != nil {
			return nil, err
		}
	}
	receipts, err := finalizeBatch(batch, graph, msg.Height, msg.FailedTaskIDs)
	if err != nil {
		return nil, err
	}
	next := cloneGenesis(k.genesis)
	next.State.ExecutionQueue = append(next.State.ExecutionQueue[:index], next.State.ExecutionQueue[index+1:]...)
	next.State.DependencyGraphs = upsertGraph(next.State.DependencyGraphs, graph)
	next.State.ExecutionReceipts = append(next.State.ExecutionReceipts, receipts...)
	if uint32(len(next.State.ExecutionReceipts)) > next.SchedulerParams.MaxReceipts {
		next.State.ExecutionReceipts = next.State.ExecutionReceipts[len(next.State.ExecutionReceipts)-int(next.SchedulerParams.MaxReceipts):]
	}
	next.State.ConflictCounters = applyConflictCounters(next.State.ConflictCounters, graph, msg.Height)
	next.State.LastFinalizedHeight = msg.Height
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return nil, err
	}
	if err := k.saveGenesis(next); err != nil {
		return nil, err
	}
	return receipts, nil
}

func (k *Keeper) saveGenesis(next GenesisState) error {
	next = cloneGenesis(next)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return prefixgenesis.Save(k.runtimeCtx, k.storeService, genesisKey, next)
}

func (k Keeper) AVMExecutionQueue(req *prototype.PageRequest) ([]types.AVMExecutionBatch, prototype.PageResponse, error) {
	queue := k.genesis.State.Export().ExecutionQueue
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(queue))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return append([]types.AVMExecutionBatch(nil), queue[start:end]...), res, nil
}

func (k Keeper) AVMExecutionReceipt(batchID, taskID string) (types.AVMExecutionReceipt, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.AVMExecutionReceipt{}, false, err
	}
	for _, receipt := range k.genesis.State.Export().ExecutionReceipts {
		if receipt.BatchID == batchID && receipt.TaskID == taskID {
			return receipt, true, nil
		}
	}
	return types.AVMExecutionReceipt{}, false, nil
}

func (k Keeper) AVMDependencyGraph(batchID string) (types.AVMDependencyGraph, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.AVMDependencyGraph{}, false, err
	}
	graph, found := k.findGraph(batchID)
	return graph, found, nil
}

func (k Keeper) AVMSchedulerParams() types.AVMSchedulerParams {
	return k.genesis.SchedulerParams
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	return m.keeper.ExportGenesis().Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func finalizeBatch(batch types.AVMExecutionBatch, graph types.AVMDependencyGraph, height uint64, failedTaskIDs []string) ([]types.AVMExecutionReceipt, error) {
	batch = batch.Normalize()
	byID := map[string]types.AVMExecutionTask{}
	for _, task := range batch.Tasks {
		byID[task.TaskID] = task
	}
	failed := failedSet(failedTaskIDs)
	root := batch.InitialStateRoot
	receipts := make([]types.AVMExecutionReceipt, 0, len(batch.Tasks))
	for i, taskID := range types.ExecutionOrder(graph) {
		task := byID[taskID]
		before := root
		status := types.ReceiptStatusSuccess
		errText := ""
		gasUsed := task.GasLimit
		if _, isFailed := failed[taskID]; isFailed {
			status = types.ReceiptStatusFailure
			errText = "AVM contract execution failed"
			gasUsed = 0
		} else {
			root = types.ApplyTaskStateRoot(root, task)
		}
		receipt := types.AVMExecutionReceipt{
			BatchID:		batch.BatchID,
			TaskID:			task.TaskID,
			ContractAddress:	task.ContractAddress,
			Height:			height,
			Order:			uint32(i),
			Status:			status,
			GasUsed:		gasUsed,
			StateRootBefore:	before,
			StateRootAfter:		root,
			Error:			errText,
			FallbackSerial:		graph.FallbackSerial,
		}
		receipt.ReceiptID = types.ComputeReceiptID(receipt)
		if err := receipt.Validate(); err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

func applyConflictCounters(counters []types.AVMConflictCounter, graph types.AVMDependencyGraph, height uint64) []types.AVMConflictCounter {
	for i, left := range graph.Nodes {
		for _, right := range graph.Nodes[i+1:] {
			if !types.NodesConflict(left, right) {
				continue
			}
			counters = types.UpsertConflictCounter(counters, left.ContractAddress, height, 1)
			if right.ContractAddress != left.ContractAddress {
				counters = types.UpsertConflictCounter(counters, right.ContractAddress, height, 1)
			}
		}
	}
	return counters
}

func upsertGraph(graphs []types.AVMDependencyGraph, graph types.AVMDependencyGraph) []types.AVMDependencyGraph {
	out := make([]types.AVMDependencyGraph, 0, len(graphs)+1)
	replaced := false
	for _, existing := range graphs {
		if existing.BatchID == graph.BatchID {
			out = append(out, graph)
			replaced = true
			continue
		}
		out = append(out, existing)
	}
	if !replaced {
		out = append(out, graph)
	}
	types.SortGraphs(out)
	return out
}

func (k Keeper) findQueuedBatch(batchID string) (int, types.AVMExecutionBatch, bool) {
	for i, batch := range k.genesis.State.Export().ExecutionQueue {
		if batch.BatchID == batchID {
			return i, batch, true
		}
	}
	return -1, types.AVMExecutionBatch{}, false
}

func (k Keeper) findGraph(batchID string) (types.AVMDependencyGraph, bool) {
	for _, graph := range k.genesis.State.Export().DependencyGraphs {
		if graph.BatchID == batchID {
			return graph, true
		}
	}
	return types.AVMDependencyGraph{}, false
}

func failedSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
