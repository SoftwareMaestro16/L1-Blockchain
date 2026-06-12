package sim

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

func New(validators []Validator, seed string) (*Simulator, error) {
	if len(validators) == 0 {
		return nil, errors.New("validator set must not be empty")
	}
	normalized := cloneValidators(validators)
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Address < normalized[j].Address })
	staking := make(map[string]int64, len(normalized))
	for _, validator := range normalized {
		if strings.TrimSpace(validator.Address) == "" {
			return nil, errors.New("validator address must not be empty")
		}
		if validator.Power <= 0 {
			return nil, errors.New("validator power must be positive")
		}
		staking[validator.Address] = validator.Power
	}
	return &Simulator{
		state: MasterchainState{
			Validators:		normalized,
			StakingSnapshot:	staking,
			Workchains:		make(map[int32]WorkchainConfig),
			Shards:			make(map[string]ShardState),
			Headers:		make(map[string]ShardHeader),
			CrossShardReceipts:	make(map[string]Receipt),
			LoadStates:		make(map[int32]WorkchainLoadState),
			FinalityLag:		2,
			RandomnessSeed:		seed,
		},
		processed:	make(map[string]struct{}),
		pendingReceipt:	make(map[string]CrossShardMessage),
	}, nil
}

func (s *Simulator) AddWorkchain(config WorkchainConfig) error {
	if config.ID == MasterchainID {
		return errors.New("masterchain id is reserved")
	}
	if len(config.AllowedVMs) == 0 {
		return errors.New("workchain must allow at least one VM")
	}
	if config.FeeDenom != FeeDenomNaet {
		return errors.New("workchain fee policy must use naet")
	}
	if strings.TrimSpace(config.AddressFormat) == "" {
		return errors.New("workchain address format must be set")
	}
	if strings.TrimSpace(config.GenesisStateHash) == "" {
		return errors.New("workchain genesis state hash must be set")
	}
	if _, exists := s.state.Workchains[config.ID]; exists {
		return errors.New("workchain already registered")
	}
	config.AllowedVMs = append([]string(nil), config.AllowedVMs...)
	sort.Strings(config.AllowedVMs)
	s.state.Workchains[config.ID] = config
	s.state.ConfigUpdates = append(s.state.ConfigUpdates, fmt.Sprintf("add-workchain:%d", config.ID))
	if err := s.AddShard(ShardID{WorkchainID: config.ID, Prefix: BaseShardID}); err != nil {
		return err
	}
	s.state.LoadStates[config.ID] = WorkchainLoadState{
		WorkchainID:		config.ID,
		ActiveShardCount:	1,
		TargetShardCount:	1,
	}
	return nil
}

func (s *Simulator) AddShard(id ShardID) error {
	if err := s.validateShardID(id); err != nil {
		return err
	}
	key := id.Key()
	if _, exists := s.state.Shards[key]; exists {
		return errors.New("shard already registered")
	}
	shard := ShardState{
		ID:			id,
		StateRoot:		HashParts("state", key, "0"),
		ValidatorSubset:	s.AssignValidators(id, 0),
		Receipts:		make(map[string]Receipt),
		Available:		true,
	}
	shard.MessageQueueRoot = hashQueue(shard.Queue)
	shard.ReceiptRoot = hashReceipts(shard.Receipts)
	s.state.Shards[key] = shard
	s.commitHeader(shard)
	s.syncLoadShardCount(id.WorkchainID)
	return nil
}

func (s *Simulator) AssignValidators(id ShardID, height uint64) []string {
	validators := cloneValidators(s.state.Validators)
	sort.Slice(validators, func(i, j int) bool {
		left := HashParts(s.state.RandomnessSeed, id.Key(), fmt.Sprint(height), validators[i].Address)
		right := HashParts(s.state.RandomnessSeed, id.Key(), fmt.Sprint(height), validators[j].Address)
		if left == right {
			return validators[i].Address < validators[j].Address
		}
		return left < right
	})
	limit := 3
	if len(validators) < limit {
		limit = len(validators)
	}
	out := make([]string, limit)
	for i := 0; i < limit; i++ {
		out[i] = validators[i].Address
	}
	sort.Strings(out)
	return out
}

func (s *Simulator) ReassignValidators(height uint64) {
	keys := sortedShardKeys(s.state.Shards)
	for _, key := range keys {
		shard := s.state.Shards[key]
		shard.ValidatorSubset = s.AssignValidators(shard.ID, height)
		s.state.Shards[key] = shard
		s.commitHeader(shard)
	}
	s.state.Height = max(s.state.Height, height)
}

func (s *Simulator) EnqueueMessage(msg CrossShardMessage) error {
	if err := s.validateMessage(msg); err != nil {
		return err
	}
	msg.MessageID = MessageID(msg.Source, msg.Destination, msg.Nonce, msg.Payload)
	source, ok := s.state.Shards[msg.Source.Key()]
	if !ok {
		return errors.New("source shard not registered")
	}
	header := s.state.Headers[msg.Source.Key()]
	msg.Proof = header.Commitment
	source.Queue = append(source.Queue, cloneMessage(msg))
	source.MessageQueueRoot = hashQueue(source.Queue)
	s.state.Shards[msg.Source.Key()] = source
	s.commitHeader(source)
	return nil
}

func (s *Simulator) ProcessNext(sourceID ShardID, height uint64) (Receipt, error) {
	source, ok := s.state.Shards[sourceID.Key()]
	if !ok {
		return Receipt{}, errors.New("source shard not registered")
	}
	if !source.Available {
		return Receipt{}, errors.New("source shard data unavailable")
	}
	if len(source.Queue) == 0 {
		return Receipt{}, errors.New("source shard queue is empty")
	}
	msg := source.Queue[0]
	source.Queue = source.Queue[1:]
	source.MessageQueueRoot = hashQueue(source.Queue)
	s.state.Shards[sourceID.Key()] = source
	s.commitHeader(source)
	return s.Deliver(msg, height)
}

func (s *Simulator) Deliver(msg CrossShardMessage, height uint64) (Receipt, error) {
	if _, exists := s.processed[msg.MessageID]; exists {
		return Receipt{}, errors.New("replayed cross-shard message")
	}
	if msg.Proof != s.state.Headers[msg.Source.Key()].Commitment {
		return Receipt{}, errors.New("invalid shard proof")
	}
	if height > msg.Timeout {
		if msg.Bounce && !msg.Bounced {
			bounced := msg
			bounced.Source, bounced.Destination = msg.Destination, msg.Source
			bounced.Bounced = true
			bounced.Nonce++
			bounced.MessageID = MessageID(bounced.Source, bounced.Destination, bounced.Nonce, bounced.Payload)
			_ = s.EnqueueMessage(bounced)
		}
		return Receipt{}, errors.New("cross-shard message timeout")
	}
	dest, ok := s.state.Shards[msg.Destination.Key()]
	if !ok {
		return Receipt{}, errors.New("destination shard not registered")
	}
	if !dest.Available {
		s.pendingReceipt[msg.MessageID] = cloneMessage(msg)
		return Receipt{}, errors.New("destination shard data unavailable")
	}
	receipt := Receipt{
		MessageID:	msg.MessageID,
		Source:		msg.Source,
		Destination:	msg.Destination,
		Success:	true,
		Height:		height,
		Proof:		HashParts("receipt", msg.MessageID, fmt.Sprint(height), dest.StateRoot),
	}
	if err := s.CommitReceipt(receipt); err != nil {
		return Receipt{}, err
	}
	s.processed[msg.MessageID] = struct{}{}
	return receipt, nil
}

func (s *Simulator) CommitReceipt(receipt Receipt) error {
	if receipt.MessageID == "" {
		return errors.New("receipt message id must be set")
	}
	if _, exists := s.state.CrossShardReceipts[receipt.MessageID]; exists {
		return errors.New("duplicate cross-shard receipt")
	}
	if _, ok := s.state.Shards[receipt.Destination.Key()]; !ok {
		return errors.New("receipt destination shard not registered")
	}
	if receipt.Proof == "" {
		return errors.New("receipt proof must be set")
	}
	s.state.CrossShardReceipts[receipt.MessageID] = receipt
	dest := s.state.Shards[receipt.Destination.Key()]
	dest.Receipts[receipt.MessageID] = receipt
	dest.ReceiptRoot = hashReceipts(dest.Receipts)
	s.state.Shards[dest.ID.Key()] = dest
	s.commitHeader(dest)
	return nil
}

func (s *Simulator) RequireReceipt(messageID string) error {
	if _, ok := s.state.CrossShardReceipts[messageID]; !ok {
		return errors.New("missing cross-shard receipt")
	}
	return nil
}

func (s *Simulator) VerifyHeaderFresh(id ShardID, height uint64) error {
	header, ok := s.state.Headers[id.Key()]
	if !ok {
		return errors.New("shard header not found")
	}
	if height > header.Height+s.state.FinalityLag {
		return errors.New("stale shard header")
	}
	return nil
}

func (s *Simulator) SplitShard(id ShardID) error {
	parent, ok := s.state.Shards[id.Key()]
	if !ok {
		return errors.New("parent shard not registered")
	}
	if len(parent.ID.Prefix) >= 60 {
		return errors.New("shard prefix is already at max depth")
	}
	delete(s.state.Shards, id.Key())
	leftID := ShardID{WorkchainID: id.WorkchainID, Prefix: id.Prefix + "0"}
	rightID := ShardID{WorkchainID: id.WorkchainID, Prefix: id.Prefix + "1"}
	left := ShardState{ID: leftID, Height: parent.Height + 1, StateRoot: HashParts(parent.StateRoot, "split-left"), ValidatorSubset: s.AssignValidators(leftID, parent.Height+1), Receipts: make(map[string]Receipt), Available: parent.Available}
	right := ShardState{ID: rightID, Height: parent.Height + 1, StateRoot: HashParts(parent.StateRoot, "split-right"), ValidatorSubset: s.AssignValidators(rightID, parent.Height+1), Receipts: make(map[string]Receipt), Available: parent.Available}
	for _, msg := range parent.Queue {
		if s.splitMessageToRight(parent.ID, msg) {
			right.Queue = append(right.Queue, msg)
		} else {
			left.Queue = append(left.Queue, msg)
		}
	}
	left.MessageQueueRoot = hashQueue(left.Queue)
	right.MessageQueueRoot = hashQueue(right.Queue)
	left.ReceiptRoot = hashReceipts(left.Receipts)
	right.ReceiptRoot = hashReceipts(right.Receipts)
	s.state.Shards[leftID.Key()] = left
	s.state.Shards[rightID.Key()] = right
	s.commitHeader(left)
	s.commitHeader(right)
	s.syncLoadShardCount(id.WorkchainID)
	return nil
}

func (s *Simulator) MergeShards(leftID ShardID, rightID ShardID) error {
	if leftID.WorkchainID != rightID.WorkchainID {
		return errors.New("cannot merge shards from different workchains")
	}
	if len(leftID.Prefix) == 0 || len(rightID.Prefix) == 0 {
		return errors.New("cannot merge root shard")
	}
	if leftID.Prefix[:len(leftID.Prefix)-1] != rightID.Prefix[:len(rightID.Prefix)-1] {
		return errors.New("can only merge sibling shards")
	}
	left, ok := s.state.Shards[leftID.Key()]
	if !ok {
		return errors.New("left shard not registered")
	}
	right, ok := s.state.Shards[rightID.Key()]
	if !ok {
		return errors.New("right shard not registered")
	}
	parentID := ShardID{WorkchainID: leftID.WorkchainID, Prefix: leftID.Prefix[:len(leftID.Prefix)-1]}
	parent := ShardState{
		ID:			parentID,
		Height:			max(left.Height, right.Height) + 1,
		StateRoot:		HashParts(left.StateRoot, right.StateRoot, "merge"),
		ValidatorSubset:	s.AssignValidators(parentID, max(left.Height, right.Height)+1),
		Queue:			append(cloneQueue(left.Queue), right.Queue...),
		Receipts:		mergeReceipts(left.Receipts, right.Receipts),
		Available:		left.Available && right.Available,
	}
	sort.Slice(parent.Queue, func(i, j int) bool { return parent.Queue[i].MessageID < parent.Queue[j].MessageID })
	parent.MessageQueueRoot = hashQueue(parent.Queue)
	parent.ReceiptRoot = hashReceipts(parent.Receipts)
	delete(s.state.Shards, leftID.Key())
	delete(s.state.Shards, rightID.Key())
	s.state.Shards[parentID.Key()] = parent
	s.commitHeader(parent)
	s.syncLoadShardCount(parentID.WorkchainID)
	return nil
}

func (s *Simulator) MarkShardAvailability(id ShardID, available bool) error {
	shard, ok := s.state.Shards[id.Key()]
	if !ok {
		return errors.New("shard not registered")
	}
	shard.Available = available
	s.state.Shards[id.Key()] = shard
	s.commitHeader(shard)
	return nil
}

func (s *Simulator) SubmitEquivocation(e EquivocationEvidence) error {
	if strings.TrimSpace(e.Validator) == "" {
		return errors.New("equivocation validator must be set")
	}
	if e.LeftRoot == "" || e.RightRoot == "" || e.LeftRoot == e.RightRoot {
		return errors.New("equivocation must include conflicting roots")
	}
	if _, ok := s.state.Shards[e.ShardID.Key()]; !ok {
		return errors.New("equivocation shard not registered")
	}
	s.state.Evidence = append(s.state.Evidence, e)
	return nil
}
