package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ShardStatusPaused	= "paused"
	ShardStatusActive	= "active"
	ShardStatusDraining	= "draining"
	ShardStatusDisabled	= "disabled"

	ShardSecurityStandard	= "standard"
	ShardSecurityHigh	= "high"
	ShardSecurityCritical	= "critical"

	RebalancePending	= "pending"
	RebalanceExecuted	= "executed"
	RebalanceRejected	= "rejected"
)

type ShardingCoordinatorParams struct {
	MaxShards			uint32
	MaxValidatorsPerShard		uint32
	MinValidatorCoverage		uint32
	MaxShardAssignmentsPerValidator	uint32
	MaxLoadMetrics			uint32
	MaxRebalanceProposals		uint32
	MaxCrossShardRoutes		uint32
	MaxStateRootReferences		uint32
	MaxRouteTimeoutBlocks		uint64
	MaxRouteInFlight		uint32
	MaxLoadTransactionsPerBlock	uint64
	MaxLoadGasPerBlock		uint64
	MaxLoadPendingMessages		uint64
}

type ShardingCoordinatorState struct {
	Shards			[]Shard
	Assignments		[]ShardValidatorAssignment
	LoadMetrics		[]ShardLoadMetric
	RebalanceProposals	[]RebalanceProposal
	CrossShardRoutes	[]CrossShardRoute
	StateRootReferences	[]ShardStateRootReference
}

type Shard struct {
	ShardID			string
	Status			string
	SecurityLevel		string
	RequiredValidatorCount	uint32
	CrossShardRoutingParams	CrossShardRoutingParams
	RegisteredHeight	uint64
	UpdatedHeight		uint64
}

type CrossShardRoutingParams struct {
	AllowInbound		bool
	AllowOutbound		bool
	MaxMessageBytes		uint32
	MaxTimeoutBlocks	uint64
	DefaultRouteLimit	uint32
}

type ShardValidatorAssignment struct {
	ShardID		string
	Validators	[]string
	AssignedHeight	uint64
	AssignmentEpoch	uint64
}

type ShardLoadMetric struct {
	ShardID			string
	TransactionsPerBlock	uint64
	GasPerBlock		uint64
	StateBytes		uint64
	PendingMessages		uint64
	ReportedHeight		uint64
}

type RebalanceProposal struct {
	ProposalID	string
	SourceShardID	string
	TargetShardID	string
	ValidatorMoves	[]ValidatorMove
	Reason		string
	Status		string
	ProposedHeight	uint64
	ExecutedHeight	uint64
}

type ValidatorMove struct {
	ValidatorID	string
	FromShardID	string
	ToShardID	string
	Weight		uint32
	Sequence	uint64
}

type CrossShardRoute struct {
	RouteID		string
	SourceShardID	string
	TargetShardID	string
	Enabled		bool
	MaxInFlight	uint32
	TimeoutBlocks	uint64
}

type ShardStateRootReference struct {
	ShardID	string
	Height	uint64
	RootHex	string
}

type MsgRegisterShard struct {
	Authority	string
	Shard		Shard
}

type MsgUpdateShardStatus struct {
	Authority	string
	ShardID		string
	Status		string
	Height		uint64
}

type MsgAssignValidatorsToShard struct {
	Authority	string
	Assignment	ShardValidatorAssignment
}

type MsgSubmitShardLoad struct {
	Reporter	string
	Load		ShardLoadMetric
}

type MsgProposeShardRebalance struct {
	Authority	string
	Proposal	RebalanceProposal
}

type MsgExecuteShardRebalance struct {
	Authority	string
	ProposalID	string
	Height		uint64
}

func DefaultShardingCoordinatorParams() ShardingCoordinatorParams {
	return ShardingCoordinatorParams{
		MaxShards:				4_096,
		MaxValidatorsPerShard:			512,
		MinValidatorCoverage:			2,
		MaxShardAssignmentsPerValidator:	4,
		MaxLoadMetrics:				100_000,
		MaxRebalanceProposals:			100_000,
		MaxCrossShardRoutes:			100_000,
		MaxStateRootReferences:			100_000,
		MaxRouteTimeoutBlocks:			1_000_000,
		MaxRouteInFlight:			100_000,
		MaxLoadTransactionsPerBlock:		10_000_000,
		MaxLoadGasPerBlock:			1_000_000_000,
		MaxLoadPendingMessages:			10_000_000,
	}
}

func EmptyShardingCoordinatorState() ShardingCoordinatorState {
	return ShardingCoordinatorState{
		Shards:			[]Shard{},
		Assignments:		[]ShardValidatorAssignment{},
		LoadMetrics:		[]ShardLoadMetric{},
		RebalanceProposals:	[]RebalanceProposal{},
		CrossShardRoutes:	[]CrossShardRoute{},
		StateRootReferences:	[]ShardStateRootReference{},
	}
}

func (p ShardingCoordinatorParams) Validate() error {
	if p.MaxShards == 0 || p.MaxValidatorsPerShard == 0 || p.MinValidatorCoverage == 0 || p.MaxShardAssignmentsPerValidator == 0 {
		return errors.New("sharding coordinator shard and validator limits must be positive")
	}
	if p.MaxLoadMetrics == 0 || p.MaxRebalanceProposals == 0 || p.MaxCrossShardRoutes == 0 || p.MaxStateRootReferences == 0 {
		return errors.New("sharding coordinator state limits must be positive")
	}
	if p.MinValidatorCoverage > p.MaxValidatorsPerShard {
		return errors.New("minimum validator coverage cannot exceed validators per shard")
	}
	if p.MaxRouteTimeoutBlocks == 0 || p.MaxRouteInFlight == 0 {
		return errors.New("cross-shard route bounds must be positive")
	}
	if p.MaxLoadTransactionsPerBlock == 0 || p.MaxLoadGasPerBlock == 0 || p.MaxLoadPendingMessages == 0 {
		return errors.New("load metric bounds must be positive")
	}
	return nil
}

func (s ShardingCoordinatorState) Export() ShardingCoordinatorState {
	out := ShardingCoordinatorState{
		Shards:			cloneShards(s.Shards),
		Assignments:		cloneAssignments(s.Assignments),
		LoadMetrics:		cloneLoadMetrics(s.LoadMetrics),
		RebalanceProposals:	cloneProposals(s.RebalanceProposals),
		CrossShardRoutes:	cloneRoutes(s.CrossShardRoutes),
		StateRootReferences:	cloneStateRoots(s.StateRootReferences),
	}
	SortShards(out.Shards)
	SortAssignments(out.Assignments)
	SortLoadMetrics(out.LoadMetrics)
	SortRebalanceProposals(out.RebalanceProposals)
	SortCrossShardRoutes(out.CrossShardRoutes)
	SortStateRoots(out.StateRootReferences)
	return out
}

func (s ShardingCoordinatorState) Validate(params ShardingCoordinatorParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.Shards)) > params.MaxShards {
		return errors.New("sharding coordinator shard count exceeds limit")
	}
	if uint32(len(s.LoadMetrics)) > params.MaxLoadMetrics {
		return errors.New("sharding coordinator load metric count exceeds limit")
	}
	if uint32(len(s.RebalanceProposals)) > params.MaxRebalanceProposals {
		return errors.New("sharding coordinator rebalance proposal count exceeds limit")
	}
	if uint32(len(s.CrossShardRoutes)) > params.MaxCrossShardRoutes {
		return errors.New("sharding coordinator route count exceeds limit")
	}
	if uint32(len(s.StateRootReferences)) > params.MaxStateRootReferences {
		return errors.New("sharding coordinator state root reference count exceeds limit")
	}

	shards := map[string]Shard{}
	activeShards := []Shard{}
	for _, shard := range s.Shards {
		if err := shard.Validate(params); err != nil {
			return err
		}
		if _, found := shards[shard.ShardID]; found {
			return fmt.Errorf("duplicate shard id %q", shard.ShardID)
		}
		shards[shard.ShardID] = shard
		if shard.Status == ShardStatusActive {
			activeShards = append(activeShards, shard)
		}
	}

	assignmentByShard := map[string]ShardValidatorAssignment{}
	validatorAssignmentCounts := map[string]uint32{}
	for _, assignment := range s.Assignments {
		if err := assignment.Validate(params); err != nil {
			return err
		}
		if _, found := shards[assignment.ShardID]; !found {
			return fmt.Errorf("assignment references unknown shard %q", assignment.ShardID)
		}
		if _, found := assignmentByShard[assignment.ShardID]; found {
			return fmt.Errorf("duplicate validator assignment for shard %q", assignment.ShardID)
		}
		assignmentByShard[assignment.ShardID] = assignment
		for _, validator := range assignment.Validators {
			validatorAssignmentCounts[validator]++
			if validatorAssignmentCounts[validator] > params.MaxShardAssignmentsPerValidator {
				return fmt.Errorf("validator %q exceeds shard assignment limit", validator)
			}
		}
	}

	for _, shard := range activeShards {
		required := max32(params.MinValidatorCoverage, shard.RequiredValidatorCount)
		assignment, found := assignmentByShard[shard.ShardID]
		if !found || uint32(len(assignment.Validators)) < required {
			return fmt.Errorf("active shard %q has insufficient validator coverage", shard.ShardID)
		}
	}

	routePairs := map[string]struct{}{}
	seenRoutes := map[string]struct{}{}
	for _, route := range s.CrossShardRoutes {
		if err := route.Validate(params); err != nil {
			return err
		}
		source, sourceFound := shards[route.SourceShardID]
		target, targetFound := shards[route.TargetShardID]
		if !sourceFound || !targetFound {
			return fmt.Errorf("cross-shard route %q references unknown shard", route.RouteID)
		}
		if _, found := seenRoutes[route.RouteID]; found {
			return fmt.Errorf("duplicate cross-shard route id %q", route.RouteID)
		}
		seenRoutes[route.RouteID] = struct{}{}
		if route.Enabled && (source.Status != ShardStatusActive || target.Status != ShardStatusActive || !source.CrossShardRoutingParams.AllowOutbound || !target.CrossShardRoutingParams.AllowInbound) {
			return fmt.Errorf("enabled cross-shard route %q requires active routable shards", route.RouteID)
		}
		if route.Enabled {
			routePairs[shardPairKey(route.SourceShardID, route.TargetShardID)] = struct{}{}
		}
	}
	for i := range activeShards {
		for j := i + 1; j < len(activeShards); j++ {
			if _, found := routePairs[shardPairKey(activeShards[i].ShardID, activeShards[j].ShardID)]; !found {
				return fmt.Errorf("cross-shard route required for active shard pair %q/%q", activeShards[i].ShardID, activeShards[j].ShardID)
			}
		}
	}

	seenLoads := map[string]struct{}{}
	for _, load := range s.LoadMetrics {
		if err := load.Validate(params); err != nil {
			return err
		}
		if _, found := shards[load.ShardID]; !found {
			return fmt.Errorf("load metric references unknown shard %q", load.ShardID)
		}
		key := load.ShardID + "\x00" + fmt.Sprint(load.ReportedHeight)
		if _, found := seenLoads[key]; found {
			return fmt.Errorf("duplicate load metric for shard %q at height %d", load.ShardID, load.ReportedHeight)
		}
		seenLoads[key] = struct{}{}
	}

	seenProposals := map[string]struct{}{}
	for _, proposal := range s.RebalanceProposals {
		if err := proposal.Validate(); err != nil {
			return err
		}
		if _, found := seenProposals[proposal.ProposalID]; found {
			return fmt.Errorf("duplicate rebalance proposal id %q", proposal.ProposalID)
		}
		seenProposals[proposal.ProposalID] = struct{}{}
		if _, found := shards[proposal.SourceShardID]; !found {
			return fmt.Errorf("rebalance proposal %q references unknown source shard", proposal.ProposalID)
		}
		if _, found := shards[proposal.TargetShardID]; !found {
			return fmt.Errorf("rebalance proposal %q references unknown target shard", proposal.ProposalID)
		}
	}

	seenRoots := map[string]struct{}{}
	for _, root := range s.StateRootReferences {
		if err := root.Validate(); err != nil {
			return err
		}
		if _, found := shards[root.ShardID]; !found {
			return fmt.Errorf("state root references unknown shard %q", root.ShardID)
		}
		key := root.ShardID + "\x00" + fmt.Sprint(root.Height)
		if _, found := seenRoots[key]; found {
			return fmt.Errorf("duplicate state root reference for shard %q at height %d", root.ShardID, root.Height)
		}
		seenRoots[key] = struct{}{}
	}
	return nil
}

func (s Shard) Normalize() Shard {
	s.ShardID = strings.TrimSpace(s.ShardID)
	s.Status = strings.TrimSpace(s.Status)
	if s.Status == "" {
		s.Status = ShardStatusPaused
	}
	s.SecurityLevel = strings.TrimSpace(s.SecurityLevel)
	if s.SecurityLevel == "" {
		s.SecurityLevel = ShardSecurityStandard
	}
	s.CrossShardRoutingParams = s.CrossShardRoutingParams.Normalize()
	return s
}

func (s Shard) Validate(params ShardingCoordinatorParams) error {
	s = s.Normalize()
	if s.ShardID == "" {
		return errors.New("shard id is required")
	}
	if !IsShardStatus(s.Status) {
		return errors.New("shard status is invalid")
	}
	if !IsShardSecurityLevel(s.SecurityLevel) {
		return errors.New("shard security level is invalid")
	}
	if s.RequiredValidatorCount == 0 || s.RequiredValidatorCount > params.MaxValidatorsPerShard {
		return errors.New("shard required validator count is invalid")
	}
	if s.RegisteredHeight == 0 || s.UpdatedHeight == 0 {
		return errors.New("shard heights must be positive")
	}
	return s.CrossShardRoutingParams.Validate(params)
}

func (r CrossShardRoutingParams) Normalize() CrossShardRoutingParams {
	if r.MaxMessageBytes == 0 {
		r.MaxMessageBytes = 1 << 20
	}
	if r.DefaultRouteLimit == 0 {
		r.DefaultRouteLimit = 1
	}
	return r
}

func (r CrossShardRoutingParams) Validate(params ShardingCoordinatorParams) error {
	r = r.Normalize()
	if r.MaxTimeoutBlocks == 0 || r.MaxTimeoutBlocks > params.MaxRouteTimeoutBlocks {
		return errors.New("cross-shard routing timeout is invalid")
	}
	if r.DefaultRouteLimit == 0 || r.DefaultRouteLimit > params.MaxRouteInFlight {
		return errors.New("cross-shard route limit is invalid")
	}
	return nil
}

func (a ShardValidatorAssignment) Normalize() ShardValidatorAssignment {
	a.ShardID = strings.TrimSpace(a.ShardID)
	a.Validators = normalizeStrings(a.Validators)
	return a
}

func (a ShardValidatorAssignment) Validate(params ShardingCoordinatorParams) error {
	a = a.Normalize()
	if a.ShardID == "" {
		return errors.New("assignment shard id is required")
	}
	if len(a.Validators) == 0 || uint32(len(a.Validators)) > params.MaxValidatorsPerShard {
		return errors.New("assignment validator count is invalid")
	}
	if a.AssignedHeight == 0 || a.AssignmentEpoch == 0 {
		return errors.New("assignment height and epoch must be positive")
	}
	return nil
}

func (l ShardLoadMetric) Normalize() ShardLoadMetric {
	l.ShardID = strings.TrimSpace(l.ShardID)
	return l
}

func (l ShardLoadMetric) Validate(params ShardingCoordinatorParams) error {
	l = l.Normalize()
	if l.ShardID == "" || l.ReportedHeight == 0 {
		return errors.New("load metric shard id and height are required")
	}
	if l.TransactionsPerBlock > params.MaxLoadTransactionsPerBlock || l.GasPerBlock > params.MaxLoadGasPerBlock || l.PendingMessages > params.MaxLoadPendingMessages {
		return errors.New("load metric exceeds configured bounds")
	}
	return nil
}

func (p RebalanceProposal) Normalize() RebalanceProposal {
	p.ProposalID = strings.TrimSpace(p.ProposalID)
	p.SourceShardID = strings.TrimSpace(p.SourceShardID)
	p.TargetShardID = strings.TrimSpace(p.TargetShardID)
	for i := range p.ValidatorMoves {
		p.ValidatorMoves[i] = p.ValidatorMoves[i].Normalize()
	}
	SortValidatorMoves(p.ValidatorMoves)
	p.Reason = strings.TrimSpace(p.Reason)
	p.Status = strings.TrimSpace(p.Status)
	if p.Status == "" {
		p.Status = RebalancePending
	}
	return p
}

func (p RebalanceProposal) Validate() error {
	p = p.Normalize()
	if p.ProposalID == "" || p.SourceShardID == "" || p.TargetShardID == "" {
		return errors.New("rebalance proposal identifiers are required")
	}
	if p.SourceShardID == p.TargetShardID {
		return errors.New("rebalance source and target shards must differ")
	}
	if len(p.ValidatorMoves) == 0 {
		return errors.New("rebalance proposal requires validator moves")
	}
	if !IsRebalanceStatus(p.Status) {
		return errors.New("rebalance proposal status is invalid")
	}
	if p.ProposedHeight == 0 {
		return errors.New("rebalance proposal height must be positive")
	}
	if p.Status == RebalanceExecuted && p.ExecutedHeight == 0 {
		return errors.New("executed rebalance proposal requires executed height")
	}
	if p.Status != RebalanceExecuted && p.ExecutedHeight != 0 {
		return errors.New("non-executed rebalance proposal cannot have executed height")
	}
	seen := map[string]struct{}{}
	for _, move := range p.ValidatorMoves {
		if err := move.Validate(p.SourceShardID, p.TargetShardID); err != nil {
			return err
		}
		if _, found := seen[move.ValidatorID]; found {
			return fmt.Errorf("duplicate validator move for %q", move.ValidatorID)
		}
		seen[move.ValidatorID] = struct{}{}
	}
	return nil
}

func (m ValidatorMove) Normalize() ValidatorMove {
	m.ValidatorID = strings.TrimSpace(m.ValidatorID)
	m.FromShardID = strings.TrimSpace(m.FromShardID)
	m.ToShardID = strings.TrimSpace(m.ToShardID)
	if m.Weight == 0 {
		m.Weight = 1
	}
	return m
}

func (m ValidatorMove) Validate(sourceShardID, targetShardID string) error {
	m = m.Normalize()
	if m.ValidatorID == "" || m.FromShardID == "" || m.ToShardID == "" {
		return errors.New("validator move identifiers are required")
	}
	if m.FromShardID != sourceShardID || m.ToShardID != targetShardID {
		return errors.New("validator move shard ids must match proposal")
	}
	if m.Sequence == 0 {
		return errors.New("validator move sequence must be positive")
	}
	return nil
}

func (r CrossShardRoute) Normalize() CrossShardRoute {
	r.RouteID = strings.TrimSpace(r.RouteID)
	r.SourceShardID = strings.TrimSpace(r.SourceShardID)
	r.TargetShardID = strings.TrimSpace(r.TargetShardID)
	if r.MaxInFlight == 0 {
		r.MaxInFlight = 1
	}
	return r
}

func (r CrossShardRoute) Validate(params ShardingCoordinatorParams) error {
	r = r.Normalize()
	if r.RouteID == "" || r.SourceShardID == "" || r.TargetShardID == "" {
		return errors.New("cross-shard route identifiers are required")
	}
	if r.SourceShardID == r.TargetShardID {
		return errors.New("cross-shard route source and target shards must differ")
	}
	if r.MaxInFlight == 0 || r.MaxInFlight > params.MaxRouteInFlight {
		return errors.New("cross-shard route max in-flight is invalid")
	}
	if r.TimeoutBlocks == 0 || r.TimeoutBlocks > params.MaxRouteTimeoutBlocks {
		return errors.New("cross-shard route timeout is invalid")
	}
	return nil
}

func (r ShardStateRootReference) Normalize() ShardStateRootReference {
	r.ShardID = strings.TrimSpace(r.ShardID)
	r.RootHex = strings.TrimSpace(r.RootHex)
	return r
}

func (r ShardStateRootReference) Validate() error {
	r = r.Normalize()
	if r.ShardID == "" || r.Height == 0 {
		return errors.New("state root shard id and height are required")
	}
	if len(r.RootHex) != 64 {
		return errors.New("state root reference must be 32-byte hex")
	}
	if _, err := hex.DecodeString(r.RootHex); err != nil {
		return fmt.Errorf("state root reference must be hex: %w", err)
	}
	return nil
}

func IsShardStatus(status string) bool {
	switch status {
	case ShardStatusPaused, ShardStatusActive, ShardStatusDraining, ShardStatusDisabled:
		return true
	default:
		return false
	}
}

func IsShardSecurityLevel(level string) bool {
	switch level {
	case ShardSecurityStandard, ShardSecurityHigh, ShardSecurityCritical:
		return true
	default:
		return false
	}
}

func IsRebalanceStatus(status string) bool {
	switch status {
	case RebalancePending, RebalanceExecuted, RebalanceRejected:
		return true
	default:
		return false
	}
}

func SortShards(shards []Shard) {
	sort.SliceStable(shards, func(i, j int) bool { return shards[i].ShardID < shards[j].ShardID })
}

func SortAssignments(assignments []ShardValidatorAssignment) {
	sort.SliceStable(assignments, func(i, j int) bool { return assignments[i].ShardID < assignments[j].ShardID })
}

func SortLoadMetrics(metrics []ShardLoadMetric) {
	sort.SliceStable(metrics, func(i, j int) bool {
		if metrics[i].ShardID != metrics[j].ShardID {
			return metrics[i].ShardID < metrics[j].ShardID
		}
		return metrics[i].ReportedHeight < metrics[j].ReportedHeight
	})
}

func SortRebalanceProposals(proposals []RebalanceProposal) {
	sort.SliceStable(proposals, func(i, j int) bool {
		if proposals[i].ProposedHeight != proposals[j].ProposedHeight {
			return proposals[i].ProposedHeight < proposals[j].ProposedHeight
		}
		return proposals[i].ProposalID < proposals[j].ProposalID
	})
}

func SortValidatorMoves(moves []ValidatorMove) {
	sort.SliceStable(moves, func(i, j int) bool {
		if moves[i].Sequence != moves[j].Sequence {
			return moves[i].Sequence < moves[j].Sequence
		}
		return moves[i].ValidatorID < moves[j].ValidatorID
	})
}

func SortCrossShardRoutes(routes []CrossShardRoute) {
	sort.SliceStable(routes, func(i, j int) bool { return routes[i].RouteID < routes[j].RouteID })
}

func SortStateRoots(roots []ShardStateRootReference) {
	sort.SliceStable(roots, func(i, j int) bool {
		if roots[i].ShardID != roots[j].ShardID {
			return roots[i].ShardID < roots[j].ShardID
		}
		return roots[i].Height < roots[j].Height
	})
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func shardPairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "\x00" + b
}

func max32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func cloneShards(shards []Shard) []Shard {
	out := append([]Shard(nil), shards...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneAssignments(assignments []ShardValidatorAssignment) []ShardValidatorAssignment {
	out := append([]ShardValidatorAssignment(nil), assignments...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneLoadMetrics(metrics []ShardLoadMetric) []ShardLoadMetric {
	out := append([]ShardLoadMetric(nil), metrics...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneProposals(proposals []RebalanceProposal) []RebalanceProposal {
	out := append([]RebalanceProposal(nil), proposals...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneRoutes(routes []CrossShardRoute) []CrossShardRoute {
	out := append([]CrossShardRoute(nil), routes...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneStateRoots(roots []ShardStateRootReference) []ShardStateRootReference {
	out := append([]ShardStateRootReference(nil), roots...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}
