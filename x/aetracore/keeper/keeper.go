package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.AetraCoreParams
	State	types.AetraCoreState
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		params,
		State:		types.EmptyState(params),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("aetracore prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	state := gs.State
	state.Params = gs.Params
	return state.Validate()
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
	if k.storeService == nil {
		return nil
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) UpdateParams(authority string, params types.AetraCoreParams) error {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	if err := params.Validate(); err != nil {
		return err
	}
	k.genesis.Params = params
	k.genesis.State.Params = params
	return nil
}

func (k *Keeper) RegisterZone(zone types.ZoneDescriptor) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := types.RegisterZone(k.genesis.State, zone)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) RegisterZoneDescriptor(descriptor types.ZoneDescriptor) error {
	return k.RegisterZone(descriptor)
}

func (k *Keeper) RegisterServiceDescriptor(descriptor types.ServiceDescriptor) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := types.RegisterServiceDescriptor(k.genesis.State, descriptor)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) RegisterShardLayout(layout types.ShardLayout) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := types.RegisterShardLayout(k.genesis.State, layout)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) CommitRoutingTable(table types.RoutingTableCommitment) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := types.CommitRoutingTable(k.genesis.State, table)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) AppendZoneCommitment(commitment types.ZoneCommitment) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := types.AppendZoneCommitment(k.genesis.State, commitment)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) CommitBlockRoots(height uint64) (types.RootSnapshot, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.RootSnapshot{}, err
	}
	next, snapshot, err := types.CommitBlockRoots(k.genesis.State, height)
	if err != nil {
		return types.RootSnapshot{}, err
	}
	k.genesis.State = next
	return snapshot, nil
}

func (k Keeper) BuildProposalSchedule(height uint64, items []types.ProposalItem) (types.ProposalSchedule, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ProposalSchedule{}, err
	}
	schedule, err := types.BuildProposalSchedule(height, items, k.genesis.Params)
	if err != nil {
		return types.ProposalSchedule{}, err
	}
	if err := types.ValidateProposalScheduleForState(schedule, k.genesis.State); err != nil {
		return types.ProposalSchedule{}, err
	}
	return schedule, nil
}

func (k Keeper) PrepareKernelProposal(ctx types.KernelConsensusContext, items []types.ProposalItem) (types.KernelBlockPlan, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.KernelBlockPlan{}, err
	}
	return types.PrepareKernelProposal(ctx, k.genesis.State, items)
}

func (k Keeper) ProcessKernelProposal(ctx types.KernelConsensusContext, plan types.KernelBlockPlan) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return types.ProcessKernelProposal(ctx, k.genesis.State, plan)
}

func (k Keeper) PrepareKernelABCIProposal(ctx types.KernelConsensusContext, localTxs, routedMessages []types.KernelMessageEnvelope, limits types.KernelGasLimits) (types.KernelABCIProposal, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.KernelABCIProposal{}, err
	}
	return types.PrepareKernelABCIProposal(ctx, k.genesis.State, localTxs, routedMessages, limits)
}

func (k Keeper) ProcessKernelABCIProposal(ctx types.KernelConsensusContext, proposal types.KernelABCIProposal, envelopes []types.KernelMessageEnvelope, limits types.KernelGasLimits) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return types.ProcessKernelABCIProposal(ctx, k.genesis.State, proposal, envelopes, limits)
}

func (k *Keeper) FinalizeKernelBlock(ctx types.KernelConsensusContext, plan types.KernelBlockPlan, input types.KernelFinalizationInput) (types.KernelFinalization, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.KernelFinalization{}, err
	}
	next, finalization, err := types.FinalizeKernelBlock(ctx, k.genesis.State, plan, input)
	if err != nil {
		return types.KernelFinalization{}, err
	}
	k.genesis.State = next
	return finalization, nil
}

func (k *Keeper) FinalizeKernelABCIBlock(ctx types.KernelConsensusContext, proposal types.KernelABCIProposal, envelopes []types.KernelMessageEnvelope, input types.KernelFinalizationInput, cleanupQueue []types.KernelCleanupItem, cleanupLimit uint64) (types.KernelFinalization, types.KernelCleanupResult, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.KernelFinalization{}, types.KernelCleanupResult{}, err
	}
	next, finalization, cleanup, err := types.FinalizeKernelABCIBlock(ctx, k.genesis.State, proposal, envelopes, input, cleanupQueue, cleanupLimit)
	if err != nil {
		return types.KernelFinalization{}, types.KernelCleanupResult{}, err
	}
	k.genesis.State = next
	return finalization, cleanup, nil
}

func (k Keeper) CommitKernelABCIBlock(finalization types.KernelFinalization, appHash string) (types.KernelABCICommitRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.KernelABCICommitRecord{}, err
	}
	return types.CommitKernelABCIBlock(finalization, appHash)
}

func (k Keeper) CollectZoneExecutionSummary(height uint64, zoneID types.ZoneID, envelopes []types.KernelMessageEnvelope, receipts []types.ExecutionReceipt, gasUsed uint64, eventsRoot string) (types.ZoneExecutionSummary, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ZoneExecutionSummary{}, err
	}
	return types.CollectZoneExecutionSummary(height, zoneID, envelopes, receipts, gasUsed, eventsRoot)
}

func (k *Keeper) CommitGlobalRoot(height uint64, contributions types.RootContributions) (types.GlobalStateRoot, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.GlobalStateRoot{}, err
	}
	root, err := types.BuildGlobalStateRoot(height, k.genesis.State, contributions)
	if err != nil {
		return types.GlobalStateRoot{}, err
	}
	next, err := types.AppendGlobalRoot(k.genesis.State, root)
	if err != nil {
		return types.GlobalStateRoot{}, err
	}
	k.genesis.State = next
	return root, nil
}

func (k *Keeper) AddExportManifest(appHash string, root types.GlobalStateRoot) (types.ExportManifest, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ExportManifest{}, err
	}
	manifest, err := types.NewExportManifest(root, appHash, k.genesis.State)
	if err != nil {
		return types.ExportManifest{}, err
	}
	next, err := types.AddExportManifest(k.genesis.State, manifest)
	if err != nil {
		return types.ExportManifest{}, err
	}
	k.genesis.State = next
	return manifest, nil
}

func (k Keeper) BuildKernelExportManifest(height uint64, appHash string) (types.ExportManifest, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ExportManifest{}, err
	}
	return types.BuildKernelExportManifest(k.genesis.State, height, appHash)
}

func (k Keeper) ValidateRootAggregationInvariants() error {
	return types.ValidateRootAggregationInvariants(k.genesis.State)
}

func (k Keeper) LatestRootSnapshot() (types.RootSnapshot, bool) {
	return k.genesis.State.LatestRootSnapshot()
}

func (k Keeper) Zones(req *prototype.PageRequest) ([]types.ZoneDescriptor, prototype.PageResponse, error) {
	zones := k.genesis.State.Export().Zones
	start, end, res, err := normalizePage(req, k.genesis.Params, len(zones))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]types.ZoneDescriptor, end-start)
	copy(out, zones[start:end])
	return out, res, nil
}

func (k Keeper) ZoneDescriptors(req *prototype.PageRequest) ([]types.ZoneDescriptor, prototype.PageResponse, error) {
	return k.Zones(req)
}

func (k Keeper) ServiceDescriptors(req *prototype.PageRequest) ([]types.ServiceDescriptor, prototype.PageResponse, error) {
	descriptors := k.genesis.State.Export().ServiceDescriptors
	start, end, res, err := normalizePage(req, k.genesis.Params, len(descriptors))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]types.ServiceDescriptor, end-start)
	copy(out, descriptors[start:end])
	return out, res, nil
}

func (k Keeper) ShardLayouts(req *prototype.PageRequest) ([]types.ShardLayout, prototype.PageResponse, error) {
	layouts := k.genesis.State.Export().ShardLayouts
	start, end, res, err := normalizePage(req, k.genesis.Params, len(layouts))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return types.ExportShardLayouts(layouts[start:end]), res, nil
}

func (k Keeper) RoutingTables(req *prototype.PageRequest) ([]types.RoutingTableCommitment, prototype.PageResponse, error) {
	tables := k.genesis.State.Export().RoutingTables
	start, end, res, err := normalizePage(req, k.genesis.Params, len(tables))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return types.ExportRoutingTables(tables[start:end]), res, nil
}

func (k Keeper) GlobalRoots(req *prototype.PageRequest) ([]types.GlobalStateRoot, prototype.PageResponse, error) {
	roots := k.genesis.State.Export().GlobalRoots
	start, end, res, err := normalizePage(req, k.genesis.Params, len(roots))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]types.GlobalStateRoot, end-start)
	copy(out, roots[start:end])
	return out, res, nil
}

func (k Keeper) ProofRoots(req *prototype.PageRequest) ([]types.ProofRoot, prototype.PageResponse, error) {
	snapshots := k.genesis.State.Export().RootSnapshots
	roots := make([]types.ProofRoot, 0)
	for _, snapshot := range snapshots {
		roots = append(roots, snapshot.ProofRoots...)
	}
	start, end, res, err := normalizePage(req, k.genesis.Params, len(roots))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]types.ProofRoot, end-start)
	copy(out, roots[start:end])
	return out, res, nil
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	gs := m.keeper.ExportGenesis()
	if err := gs.Validate(); err != nil {
		return err
	}
	return nil
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State.Params = gs.Params
	gs.State = gs.State.Export()
	return gs
}

func normalizePage(req *prototype.PageRequest, params types.AetraCoreParams, total int) (start int, end int, res prototype.PageResponse, err error) {
	if err := params.Validate(); err != nil {
		return 0, 0, prototype.PageResponse{}, err
	}
	if req == nil {
		req = &prototype.PageRequest{}
	}
	limit := req.Limit
	if limit == 0 {
		limit = params.DefaultQueryLimit
	}
	if limit == 0 || limit > params.MaxQueryLimit {
		return 0, 0, prototype.PageResponse{}, errors.New("aetracore query limit out of bounds")
	}
	if req.Offset > uint64(total) {
		return 0, 0, prototype.PageResponse{}, errors.New("aetracore query offset out of bounds")
	}
	start = int(req.Offset)
	end64 := req.Offset + limit
	if end64 < req.Offset {
		return 0, 0, prototype.PageResponse{}, errors.New("aetracore query offset overflow")
	}
	if end64 > uint64(total) {
		end64 = uint64(total)
	}
	end = int(end64)
	if end < total {
		res.NextOffset = uint64(end)
	}
	return start, end, res, nil
}
