package sim

import (
	"errors"
	"fmt"
	"sort"

	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
)

const MaxSimActiveShards = uint32(64)

type ShardActivationPolicy struct {
	WorkchainID		int32
	LoadParams		loadtypes.Params
	PartialShardCount	uint32
	MaxShardCount		uint32
	CooldownBlocks		uint64
	RoutingEpoch		uint64
}

type LoadShardTransition struct {
	WorkchainID		int32
	Height			uint64
	LoadResult		loadtypes.Result
	PreviousShardCount	uint32
	DesiredShardCount	uint32
	AppliedShardCount	uint32
	CooldownStarted		bool
	CooldownSatisfied	bool
	ValidatorReassigned	bool
}

func DefaultShardActivationPolicy(workchainID int32) ShardActivationPolicy {
	return ShardActivationPolicy{
		WorkchainID:		workchainID,
		LoadParams:		loadtypes.DefaultParams(),
		PartialShardCount:	2,
		MaxShardCount:		4,
		CooldownBlocks:		2,
	}
}

func (p ShardActivationPolicy) Validate() error {
	if p.WorkchainID == MasterchainID {
		return errors.New("load sharding policy cannot target masterchain")
	}
	if err := normalizeLoadParams(p.LoadParams).Validate(); err != nil {
		return err
	}
	if p.PartialShardCount == 0 {
		return errors.New("partial shard count must be positive")
	}
	if p.MaxShardCount == 0 {
		return errors.New("max shard count must be positive")
	}
	if p.PartialShardCount > p.MaxShardCount {
		return errors.New("partial shard count cannot exceed max shard count")
	}
	if p.MaxShardCount > MaxSimActiveShards {
		return fmt.Errorf("max shard count must be <= %d", MaxSimActiveShards)
	}
	if !isPowerOfTwo(p.PartialShardCount) || !isPowerOfTwo(p.MaxShardCount) {
		return errors.New("partial and max shard counts must be powers of two")
	}
	return nil
}

func (s *Simulator) UpdateLoadAndShards(policy ShardActivationPolicy, metrics loadtypes.Metrics, height uint64) (LoadShardTransition, error) {
	if err := policy.Validate(); err != nil {
		return LoadShardTransition{}, err
	}
	if _, ok := s.state.Workchains[policy.WorkchainID]; !ok {
		return LoadShardTransition{}, errors.New("workchain is not registered")
	}
	s.ensureLoadState(policy.WorkchainID)
	current := s.ActiveShardCount(policy.WorkchainID)
	if current == 0 {
		return LoadShardTransition{}, errors.New("workchain has no active shards")
	}

	loadState := s.state.LoadStates[policy.WorkchainID]
	params := normalizeLoadParams(policy.LoadParams)
	loadResult, err := loadtypes.ComputeLoadScore(params, loadState.EMA, metrics)
	if err != nil {
		return LoadShardTransition{}, err
	}

	desired := desiredShardCount(policy, loadResult.Band)
	applied := desired
	cooldownStarted := false
	cooldownSatisfied := false
	if desired < current {
		if loadState.BelowTargetSinceHeight == 0 || height < loadState.BelowTargetSinceHeight {
			loadState.BelowTargetSinceHeight = height
			cooldownStarted = true
		}
		if policy.CooldownBlocks > 0 && height-loadState.BelowTargetSinceHeight < policy.CooldownBlocks {
			applied = current
		} else {
			cooldownSatisfied = true
		}
	} else {
		loadState.BelowTargetSinceHeight = 0
	}

	validatorReassigned := false
	if policy.RoutingEpoch != loadState.RoutingEpoch {
		loadState.RoutingEpoch = policy.RoutingEpoch
		loadState.LastValidatorEpochHeight = height
		validatorReassigned = true
	}

	loadState.EMA = loadResult.EMA
	loadState.LastLoadScoreBps = loadResult.LoadScoreBps
	loadState.LastLoadBand = loadResult.Band
	loadState.TargetShardCount = desired
	loadState.CooldownBlocks = policy.CooldownBlocks
	loadState.LastUpdateHeight = height
	s.state.LoadStates[policy.WorkchainID] = loadState

	if applied != current {
		if err := s.SetActiveShardCount(policy.WorkchainID, applied); err != nil {
			return LoadShardTransition{}, err
		}
	}
	if validatorReassigned {
		s.ReassignValidators(height)
	}

	loadState = s.state.LoadStates[policy.WorkchainID]
	loadState.ActiveShardCount = s.ActiveShardCount(policy.WorkchainID)
	s.state.LoadStates[policy.WorkchainID] = loadState
	s.state.Height = max(s.state.Height, height)

	return LoadShardTransition{
		WorkchainID:		policy.WorkchainID,
		Height:			height,
		LoadResult:		loadResult,
		PreviousShardCount:	current,
		DesiredShardCount:	desired,
		AppliedShardCount:	loadState.ActiveShardCount,
		CooldownStarted:	cooldownStarted,
		CooldownSatisfied:	cooldownSatisfied,
		ValidatorReassigned:	validatorReassigned,
	}, nil
}

func (s *Simulator) ActiveShardCount(workchainID int32) uint32 {
	return uint32(countWorkchainShards(s.state.Shards, workchainID))
}

func (s *Simulator) WorkchainShardIDs(workchainID int32) []ShardID {
	shards := make([]ShardID, 0)
	for _, shard := range s.state.Shards {
		if shard.ID.WorkchainID == workchainID {
			shards = append(shards, shard.ID)
		}
	}
	sortShardIDs(shards)
	return shards
}

func (s *Simulator) SetActiveShardCount(workchainID int32, target uint32) error {
	if target == 0 {
		return errors.New("target active shard count must be positive")
	}
	if target > MaxSimActiveShards {
		return fmt.Errorf("target active shard count must be <= %d", MaxSimActiveShards)
	}
	if !isPowerOfTwo(target) {
		return errors.New("target active shard count must be a power of two")
	}
	if _, ok := s.state.Workchains[workchainID]; !ok {
		return errors.New("workchain is not registered")
	}
	for s.ActiveShardCount(workchainID) < target {
		candidate, err := s.nextShardToSplit(workchainID)
		if err != nil {
			return err
		}
		if err := s.SplitShard(candidate); err != nil {
			return err
		}
	}
	for s.ActiveShardCount(workchainID) > target {
		left, right, err := s.nextShardPairToMerge(workchainID)
		if err != nil {
			return err
		}
		if err := s.MergeShards(left, right); err != nil {
			return err
		}
	}
	s.ensureLoadState(workchainID)
	loadState := s.state.LoadStates[workchainID]
	loadState.ActiveShardCount = target
	s.state.LoadStates[workchainID] = loadState
	return nil
}

func (s *Simulator) RouteWork(workchainID int32, routingKey []byte) (ShardID, error) {
	if len(routingKey) == 0 {
		return ShardID{}, errors.New("routing key must not be empty")
	}
	s.ensureLoadState(workchainID)
	shards := s.WorkchainShardIDs(workchainID)
	if len(shards) == 0 {
		return ShardID{}, errors.New("workchain has no active shards")
	}
	epoch := s.state.LoadStates[workchainID].RoutingEpoch
	index := routingtypes.AssignShard(routingtypes.ZoneID(workchainZoneID(workchainID)), routingKey, epoch, uint32(len(shards)))
	selected := shards[int(index)]
	shard := s.state.Shards[selected.Key()]
	if !shard.Available {
		return ShardID{}, errors.New("selected shard data unavailable")
	}
	if _, ok := s.state.Headers[selected.Key()]; !ok {
		return ShardID{}, errors.New("selected shard header not found")
	}
	return selected, nil
}

func (s *Simulator) ensureLoadState(workchainID int32) {
	if s.state.LoadStates == nil {
		s.state.LoadStates = make(map[int32]WorkchainLoadState)
	}
	if _, ok := s.state.LoadStates[workchainID]; ok {
		return
	}
	count := s.ActiveShardCount(workchainID)
	if count == 0 {
		count = 1
	}
	s.state.LoadStates[workchainID] = WorkchainLoadState{
		WorkchainID:		workchainID,
		ActiveShardCount:	count,
		TargetShardCount:	count,
	}
}

func (s *Simulator) syncLoadShardCount(workchainID int32) {
	if s.state.LoadStates == nil {
		return
	}
	loadState, ok := s.state.LoadStates[workchainID]
	if !ok {
		return
	}
	count := s.ActiveShardCount(workchainID)
	loadState.ActiveShardCount = count
	if loadState.TargetShardCount == 0 {
		loadState.TargetShardCount = count
	}
	s.state.LoadStates[workchainID] = loadState
}

func (s *Simulator) nextShardToSplit(workchainID int32) (ShardID, error) {
	shards := s.WorkchainShardIDs(workchainID)
	sort.SliceStable(shards, func(i, j int) bool {
		if len(shards[i].Prefix) != len(shards[j].Prefix) {
			return len(shards[i].Prefix) < len(shards[j].Prefix)
		}
		return shards[i].Key() < shards[j].Key()
	})
	for _, id := range shards {
		if len(id.Prefix) < 60 {
			return id, nil
		}
	}
	return ShardID{}, errors.New("no shard can be split further")
}

func (s *Simulator) nextShardPairToMerge(workchainID int32) (ShardID, ShardID, error) {
	shards := s.WorkchainShardIDs(workchainID)
	byKey := make(map[string]ShardID, len(shards))
	for _, id := range shards {
		byKey[id.Key()] = id
	}
	sort.SliceStable(shards, func(i, j int) bool {
		if len(shards[i].Prefix) != len(shards[j].Prefix) {
			return len(shards[i].Prefix) > len(shards[j].Prefix)
		}
		return shards[i].Key() < shards[j].Key()
	})
	for _, id := range shards {
		if id.Prefix == "" {
			continue
		}
		sibling := ShardID{WorkchainID: workchainID, Prefix: siblingPrefix(id.Prefix)}
		if _, ok := byKey[sibling.Key()]; !ok {
			continue
		}
		if id.Prefix < sibling.Prefix {
			return id, sibling, nil
		}
		return sibling, id, nil
	}
	return ShardID{}, ShardID{}, errors.New("no sibling shard pair can be merged")
}

func (s *Simulator) splitMessageToRight(parent ShardID, msg CrossShardMessage) bool {
	key := messageRoutingKey(msg)
	epoch := s.state.LoadStates[parent.WorkchainID].RoutingEpoch
	shard := routingtypes.AssignShard(routingtypes.ZoneID(workchainZoneID(parent.WorkchainID)), key, epoch, 2)
	return shard == 1
}

func desiredShardCount(policy ShardActivationPolicy, band loadtypes.LoadBand) uint32 {
	switch band {
	case loadtypes.LoadBandLow:
		return 1
	case loadtypes.LoadBandMedium:
		return policy.PartialShardCount
	default:
		return policy.MaxShardCount
	}
}

func normalizeLoadParams(params loadtypes.Params) loadtypes.Params {
	if params == (loadtypes.Params{}) {
		return loadtypes.DefaultParams()
	}
	return params
}

func messageRoutingKey(msg CrossShardMessage) []byte {
	if len(msg.RoutingKey) > 0 {
		return append([]byte(nil), msg.RoutingKey...)
	}
	if msg.MessageID != "" {
		return []byte(msg.MessageID)
	}
	return []byte(MessageID(msg.Source, msg.Destination, msg.Nonce, msg.Payload))
}

func countWorkchainShards(shards map[string]ShardState, workchainID int32) uint64 {
	var count uint64
	for _, shard := range shards {
		if shard.ID.WorkchainID == workchainID {
			count++
		}
	}
	return count
}

func sortShardIDs(shards []ShardID) {
	sort.SliceStable(shards, func(i, j int) bool {
		return shards[i].Key() < shards[j].Key()
	})
}

func siblingPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	last := prefix[len(prefix)-1]
	if last == '0' {
		return prefix[:len(prefix)-1] + "1"
	}
	return prefix[:len(prefix)-1] + "0"
}

func workchainZoneID(workchainID int32) string {
	return fmt.Sprintf("WORKCHAIN_%d", workchainID)
}

func isPowerOfTwo(value uint32) bool {
	return value > 0 && value&(value-1) == 0
}
