package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/sharding-coordinator/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	ShardParams	types.ShardingCoordinatorParams
	State		types.ShardingCoordinatorState
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
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		prototype.DefaultParams(),
		ShardParams:	types.DefaultShardingCoordinatorParams(),
		State:		types.EmptyShardingCoordinatorState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("sharding coordinator prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.ShardParams)
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

func (k *Keeper) RegisterShard(msg types.MsgRegisterShard) (types.Shard, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.Shard{}, err
	}
	shard := msg.Shard.Normalize()
	if shard.RegisteredHeight == 0 {
		return types.Shard{}, errors.New("shard registration height must be positive")
	}
	shard.UpdatedHeight = shard.RegisteredHeight
	if _, _, found := shardIndex(k.genesis.State.Shards, shard.ShardID); found {
		return types.Shard{}, errors.New("shard already registered")
	}
	next := cloneGenesis(k.genesis)
	next.State.Shards = append(next.State.Shards, shard)
	if shard.Status == types.ShardStatusActive {
		next.State.CrossShardRoutes = ensureRoutesForActivatedShard(next.State, shard, next.ShardParams)
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.Shard{}, err
	}
	k.genesis = next
	return shard.Normalize(), nil
}

func (k *Keeper) UpdateShardStatus(msg types.MsgUpdateShardStatus) (types.Shard, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.Shard{}, err
	}
	if msg.Height == 0 {
		return types.Shard{}, errors.New("shard status update height must be positive")
	}
	if !types.IsShardStatus(msg.Status) {
		return types.Shard{}, errors.New("shard status is invalid")
	}
	idx, shard, found := shardIndex(k.genesis.State.Shards, msg.ShardID)
	if !found {
		return types.Shard{}, errors.New("shard not found")
	}
	shard.Status = msg.Status
	shard.UpdatedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Shards[idx] = shard.Normalize()
	if msg.Status == types.ShardStatusActive {
		next.State.CrossShardRoutes = ensureRoutesForActivatedShard(next.State, shard.Normalize(), next.ShardParams)
	} else {
		for i := range next.State.CrossShardRoutes {
			if next.State.CrossShardRoutes[i].SourceShardID == shard.ShardID || next.State.CrossShardRoutes[i].TargetShardID == shard.ShardID {
				next.State.CrossShardRoutes[i].Enabled = false
			}
		}
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.Shard{}, err
	}
	k.genesis = next
	return shard.Normalize(), nil
}

func (k *Keeper) AssignValidatorsToShard(msg types.MsgAssignValidatorsToShard) (types.ShardValidatorAssignment, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ShardValidatorAssignment{}, err
	}
	assignment := msg.Assignment.Normalize()
	if _, _, found := shardIndex(k.genesis.State.Shards, assignment.ShardID); !found {
		return types.ShardValidatorAssignment{}, errors.New("assignment references unknown shard")
	}
	next := cloneGenesis(k.genesis)
	if idx, _, found := assignmentIndex(next.State.Assignments, assignment.ShardID); found {
		next.State.Assignments[idx] = assignment
	} else {
		next.State.Assignments = append(next.State.Assignments, assignment)
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ShardValidatorAssignment{}, err
	}
	k.genesis = next
	return assignment, nil
}

func (k *Keeper) SubmitShardLoad(msg types.MsgSubmitShardLoad) (types.ShardLoadMetric, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ShardLoadMetric{}, err
	}
	if msg.Reporter == "" {
		return types.ShardLoadMetric{}, errors.New("shard load reporter is required")
	}
	load := msg.Load.Normalize()
	if _, _, found := shardIndex(k.genesis.State.Shards, load.ShardID); !found {
		return types.ShardLoadMetric{}, errors.New("load metric references unknown shard")
	}
	if _, _, found := loadIndex(k.genesis.State.LoadMetrics, load.ShardID, load.ReportedHeight); found {
		return types.ShardLoadMetric{}, errors.New("load metric already submitted for shard height")
	}
	next := cloneGenesis(k.genesis)
	next.State.LoadMetrics = append(next.State.LoadMetrics, load)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ShardLoadMetric{}, err
	}
	k.genesis = next
	return load, nil
}

func (k *Keeper) ProposeShardRebalance(msg types.MsgProposeShardRebalance) (types.RebalanceProposal, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.RebalanceProposal{}, err
	}
	proposal := msg.Proposal.Normalize()
	if proposal.Status == "" {
		proposal.Status = types.RebalancePending
	}
	if _, _, found := proposalIndex(k.genesis.State.RebalanceProposals, proposal.ProposalID); found {
		return types.RebalanceProposal{}, errors.New("rebalance proposal already exists")
	}
	if _, _, found := shardIndex(k.genesis.State.Shards, proposal.SourceShardID); !found {
		return types.RebalanceProposal{}, errors.New("rebalance proposal references unknown source shard")
	}
	if _, _, found := shardIndex(k.genesis.State.Shards, proposal.TargetShardID); !found {
		return types.RebalanceProposal{}, errors.New("rebalance proposal references unknown target shard")
	}
	next := cloneGenesis(k.genesis)
	next.State.RebalanceProposals = append(next.State.RebalanceProposals, proposal)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.RebalanceProposal{}, err
	}
	k.genesis = next
	return proposal, nil
}

func (k *Keeper) ExecuteShardRebalance(msg types.MsgExecuteShardRebalance) (types.RebalanceProposal, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.RebalanceProposal{}, err
	}
	if msg.Height == 0 {
		return types.RebalanceProposal{}, errors.New("rebalance execution height must be positive")
	}
	idx, proposal, found := proposalIndex(k.genesis.State.RebalanceProposals, msg.ProposalID)
	if !found {
		return types.RebalanceProposal{}, errors.New("rebalance proposal not found")
	}
	if proposal.Status != types.RebalancePending || proposal.ExecutedHeight != 0 {
		return types.RebalanceProposal{}, errors.New("rebalance proposal cannot execute twice")
	}
	next := cloneGenesis(k.genesis)
	if err := applyValidatorMoves(&next.State, proposal); err != nil {
		return types.RebalanceProposal{}, err
	}
	proposal.Status = types.RebalanceExecuted
	proposal.ExecutedHeight = msg.Height
	next.State.RebalanceProposals[idx] = proposal.Normalize()
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.RebalanceProposal{}, err
	}
	k.genesis = next
	return proposal.Normalize(), nil
}

func (k Keeper) Shard(shardID string) (types.Shard, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.Shard{}, false, err
	}
	_, shard, found := shardIndex(k.genesis.State.Shards, shardID)
	return shard, found, nil
}

func (k Keeper) Shards() ([]types.Shard, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	return k.genesis.State.Export().Shards, nil
}

func (k Keeper) ShardValidators(shardID string) ([]string, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, false, err
	}
	_, assignment, found := assignmentIndex(k.genesis.State.Assignments, shardID)
	return append([]string(nil), assignment.Validators...), found, nil
}

func (k Keeper) ShardLoad(shardID string) ([]types.ShardLoadMetric, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	out := make([]types.ShardLoadMetric, 0)
	for _, load := range k.genesis.State.Export().LoadMetrics {
		if shardID == "" || load.ShardID == shardID {
			out = append(out, load)
		}
	}
	types.SortLoadMetrics(out)
	return out, nil
}

func (k Keeper) RebalanceProposal(proposalID string) (types.RebalanceProposal, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.RebalanceProposal{}, false, err
	}
	_, proposal, found := proposalIndex(k.genesis.State.RebalanceProposals, proposalID)
	return proposal, found, nil
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

func (k Keeper) requireAuthority(authority string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return k.genesis.Params.Authorize(authority)
}

func ensureRoutesForActivatedShard(state types.ShardingCoordinatorState, activated types.Shard, params types.ShardingCoordinatorParams) []types.CrossShardRoute {
	routes := cloneRoutes(state.CrossShardRoutes)
	if !activated.CrossShardRoutingParams.AllowInbound || !activated.CrossShardRoutingParams.AllowOutbound {
		return routes
	}
	for _, shard := range state.Shards {
		shard = shard.Normalize()
		if shard.ShardID == activated.ShardID || shard.Status != types.ShardStatusActive {
			continue
		}
		if !shard.CrossShardRoutingParams.AllowInbound || !shard.CrossShardRoutingParams.AllowOutbound {
			continue
		}
		if routePairExists(routes, activated.ShardID, shard.ShardID) {
			continue
		}
		routes = append(routes, types.CrossShardRoute{
			RouteID:	deterministicRouteID(activated.ShardID, shard.ShardID),
			SourceShardID:	activated.ShardID,
			TargetShardID:	shard.ShardID,
			Enabled:	true,
			MaxInFlight:	min32(activated.CrossShardRoutingParams.DefaultRouteLimit, params.MaxRouteInFlight),
			TimeoutBlocks:	min64(activated.CrossShardRoutingParams.MaxTimeoutBlocks, params.MaxRouteTimeoutBlocks),
		})
	}
	return routes
}

func applyValidatorMoves(state *types.ShardingCoordinatorState, proposal types.RebalanceProposal) error {
	proposal = proposal.Normalize()
	sourceIdx, source, found := assignmentIndex(state.Assignments, proposal.SourceShardID)
	if !found {
		return errors.New("source shard assignment not found")
	}
	targetIdx, target, found := assignmentIndex(state.Assignments, proposal.TargetShardID)
	if !found {
		target = types.ShardValidatorAssignment{
			ShardID:		proposal.TargetShardID,
			AssignedHeight:		proposal.ProposedHeight,
			AssignmentEpoch:	source.AssignmentEpoch + 1,
		}
	}
	sourceSet := map[string]struct{}{}
	for _, validator := range source.Validators {
		sourceSet[validator] = struct{}{}
	}
	targetSet := map[string]struct{}{}
	for _, validator := range target.Validators {
		targetSet[validator] = struct{}{}
	}
	for _, move := range proposal.ValidatorMoves {
		if _, found := sourceSet[move.ValidatorID]; !found {
			return errors.New("rebalance validator is not assigned to source shard")
		}
		delete(sourceSet, move.ValidatorID)
		targetSet[move.ValidatorID] = struct{}{}
	}
	source.Validators = setToSortedSlice(sourceSet)
	target.Validators = setToSortedSlice(targetSet)
	source.AssignmentEpoch++
	target.AssignmentEpoch = max64(target.AssignmentEpoch, source.AssignmentEpoch)
	source.AssignedHeight = proposal.ProposedHeight
	target.AssignedHeight = proposal.ProposedHeight
	state.Assignments[sourceIdx] = source.Normalize()
	if found {
		state.Assignments[targetIdx] = target.Normalize()
	} else {
		state.Assignments = append(state.Assignments, target.Normalize())
	}
	return nil
}

func shardIndex(shards []types.Shard, shardID string) (int, types.Shard, bool) {
	for i, shard := range shards {
		if shard.ShardID == shardID {
			return i, shard, true
		}
	}
	return -1, types.Shard{}, false
}

func assignmentIndex(assignments []types.ShardValidatorAssignment, shardID string) (int, types.ShardValidatorAssignment, bool) {
	for i, assignment := range assignments {
		if assignment.ShardID == shardID {
			return i, assignment, true
		}
	}
	return -1, types.ShardValidatorAssignment{}, false
}

func loadIndex(loads []types.ShardLoadMetric, shardID string, height uint64) (int, types.ShardLoadMetric, bool) {
	for i, load := range loads {
		if load.ShardID == shardID && load.ReportedHeight == height {
			return i, load, true
		}
	}
	return -1, types.ShardLoadMetric{}, false
}

func proposalIndex(proposals []types.RebalanceProposal, proposalID string) (int, types.RebalanceProposal, bool) {
	for i, proposal := range proposals {
		if proposal.ProposalID == proposalID {
			return i, proposal, true
		}
	}
	return -1, types.RebalanceProposal{}, false
}

func routePairExists(routes []types.CrossShardRoute, a, b string) bool {
	for _, route := range routes {
		if !route.Enabled {
			continue
		}
		if (route.SourceShardID == a && route.TargetShardID == b) || (route.SourceShardID == b && route.TargetShardID == a) {
			return true
		}
	}
	return false
}

func deterministicRouteID(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "-to-" + b
}

func cloneRoutes(routes []types.CrossShardRoute) []types.CrossShardRoute {
	out := append([]types.CrossShardRoute(nil), routes...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func setToSortedSlice(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sortStrings(out)
	return out
}

func sortStrings(values []string) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

func min32(a, b uint32) uint32 {
	if a == 0 || a > b {
		return b
	}
	return a
}

func min64(a, b uint64) uint64 {
	if a == 0 || a > b {
		return b
	}
	return a
}

func max64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
