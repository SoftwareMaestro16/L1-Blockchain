package sim

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func (s *Simulator) Export() MasterchainState {
	out := s.state
	out.Validators = cloneValidators(s.state.Validators)
	out.StakingSnapshot = cloneIntMap(s.state.StakingSnapshot)
	out.Workchains = cloneWorkchains(s.state.Workchains)
	out.Shards = cloneShards(s.state.Shards)
	out.Headers = cloneHeaders(s.state.Headers)
	out.CrossShardReceipts = cloneReceipts(s.state.CrossShardReceipts)
	out.LoadStates = cloneLoadStates(s.state.LoadStates)
	out.ConfigUpdates = append([]string(nil), s.state.ConfigUpdates...)
	out.Evidence = append([]EquivocationEvidence(nil), s.state.Evidence...)
	return out
}

func Import(state MasterchainState) (*Simulator, error) {
	sim, err := New(state.Validators, state.RandomnessSeed)
	if err != nil {
		return nil, err
	}
	if err := ValidateState(state); err != nil {
		return nil, err
	}
	state = normalizeStateMaps(state)
	sim.state = state
	sim.processed = make(map[string]struct{}, len(state.CrossShardReceipts))
	for id := range state.CrossShardReceipts {
		sim.processed[id] = struct{}{}
	}
	sim.pendingReceipt = make(map[string]CrossShardMessage)
	return sim, nil
}

func ValidateState(state MasterchainState) error {
	state = normalizeStateMaps(state)
	if len(state.Validators) == 0 {
		return errors.New("validator set must not be empty")
	}
	for id, wc := range state.Workchains {
		if id != wc.ID {
			return errors.New("workchain registry key mismatch")
		}
		if wc.FeeDenom != FeeDenomNaet {
			return errors.New("workchain fee policy must use naet")
		}
	}
	for key, shard := range state.Shards {
		if key != shard.ID.Key() {
			return errors.New("shard registry key mismatch")
		}
		header, ok := state.Headers[key]
		if !ok {
			return errors.New("missing shard header")
		}
		if header.Commitment != headerCommitment(shard) {
			return errors.New("invalid shard header commitment")
		}
	}
	for id, receipt := range state.CrossShardReceipts {
		if id != receipt.MessageID {
			return errors.New("receipt registry key mismatch")
		}
		if receipt.Proof == "" {
			return errors.New("receipt proof must be set")
		}
	}
	for workchainID, loadState := range state.LoadStates {
		if workchainID != loadState.WorkchainID {
			return errors.New("load state workchain key mismatch")
		}
		if _, ok := state.Workchains[workchainID]; !ok {
			return errors.New("load state workchain is not registered")
		}
		if err := loadState.EMA.Validate(); err != nil {
			return err
		}
		if loadState.LastLoadScoreBps > 10_000 {
			return errors.New("load score must be <= 10000 bps")
		}
		if loadState.ActiveShardCount != uint32(countWorkchainShards(state.Shards, workchainID)) {
			return errors.New("load active shard count does not match shard registry")
		}
	}
	return nil
}

func normalizeStateMaps(state MasterchainState) MasterchainState {
	if state.StakingSnapshot == nil {
		state.StakingSnapshot = make(map[string]int64)
	}
	if state.Workchains == nil {
		state.Workchains = make(map[int32]WorkchainConfig)
	}
	if state.Shards == nil {
		state.Shards = make(map[string]ShardState)
	}
	if state.Headers == nil {
		state.Headers = make(map[string]ShardHeader)
	}
	if state.CrossShardReceipts == nil {
		state.CrossShardReceipts = make(map[string]Receipt)
	}
	if state.LoadStates == nil {
		state.LoadStates = make(map[int32]WorkchainLoadState)
	}
	return state
}

func MessageID(source ShardID, destination ShardID, nonce uint64, payload []byte) string {
	return HashParts("message", source.Key(), destination.Key(), fmt.Sprint(nonce), string(payload))
}

func HashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte{0})
		h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (id ShardID) Key() string {
	return fmt.Sprintf("%d:%s", id.WorkchainID, id.Prefix)
}

func (s *Simulator) validateShardID(id ShardID) error {
	if _, ok := s.state.Workchains[id.WorkchainID]; !ok {
		return errors.New("workchain is not registered")
	}
	if len(id.Prefix) > 60 {
		return errors.New("shard prefix length must be <= 60")
	}
	for _, r := range id.Prefix {
		if r != '0' && r != '1' {
			return errors.New("shard prefix must be binary")
		}
	}
	return nil
}

func (s *Simulator) validateMessage(msg CrossShardMessage) error {
	if msg.Source.Key() == msg.Destination.Key() {
		return errors.New("cross-shard message requires different shards")
	}
	if _, ok := s.state.Shards[msg.Destination.Key()]; !ok {
		return errors.New("destination shard not registered")
	}
	if msg.Timeout == 0 {
		return errors.New("cross-shard message timeout must be set")
	}
	return nil
}

func (s *Simulator) commitHeader(shard ShardState) {
	header := ShardHeader{
		ShardID:		shard.ID,
		Height:			shard.Height,
		StateRoot:		shard.StateRoot,
		MessageQueueRoot:	shard.MessageQueueRoot,
		ReceiptRoot:		shard.ReceiptRoot,
		ValidatorSubset:	append([]string(nil), shard.ValidatorSubset...),
		Available:		shard.Available,
	}
	header.Commitment = headerCommitment(shard)
	s.state.Headers[shard.ID.Key()] = header
}

func headerCommitment(shard ShardState) string {
	return HashParts("header", shard.ID.Key(), fmt.Sprint(shard.Height), shard.StateRoot, shard.MessageQueueRoot, shard.ReceiptRoot, strings.Join(shard.ValidatorSubset, ","), fmt.Sprint(shard.Available))
}

func hashQueue(queue []CrossShardMessage) string {
	ids := make([]string, len(queue))
	for i, msg := range queue {
		ids[i] = msg.MessageID
	}
	sort.Strings(ids)
	return HashParts(append([]string{"queue"}, ids...)...)
}

func hashReceipts(receipts map[string]Receipt) string {
	ids := make([]string, 0, len(receipts))
	for id := range receipts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return HashParts(append([]string{"receipts"}, ids...)...)
}

func sortedShardKeys(shards map[string]ShardState) []string {
	keys := make([]string, 0, len(shards))
	for key := range shards {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneValidators(in []Validator) []Validator {
	out := make([]Validator, len(in))
	copy(out, in)
	return out
}

func cloneMessage(msg CrossShardMessage) CrossShardMessage {
	msg.Payload = append([]byte(nil), msg.Payload...)
	msg.RoutingKey = append([]byte(nil), msg.RoutingKey...)
	return msg
}

func cloneQueue(in []CrossShardMessage) []CrossShardMessage {
	out := make([]CrossShardMessage, len(in))
	for i, msg := range in {
		out[i] = cloneMessage(msg)
	}
	return out
}

func cloneIntMap(in map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneWorkchains(in map[int32]WorkchainConfig) map[int32]WorkchainConfig {
	out := make(map[int32]WorkchainConfig, len(in))
	for key, value := range in {
		value.AllowedVMs = append([]string(nil), value.AllowedVMs...)
		out[key] = value
	}
	return out
}

func cloneShards(in map[string]ShardState) map[string]ShardState {
	out := make(map[string]ShardState, len(in))
	for key, value := range in {
		value.ValidatorSubset = append([]string(nil), value.ValidatorSubset...)
		value.Queue = cloneQueue(value.Queue)
		value.Receipts = cloneReceipts(value.Receipts)
		out[key] = value
	}
	return out
}

func cloneHeaders(in map[string]ShardHeader) map[string]ShardHeader {
	out := make(map[string]ShardHeader, len(in))
	for key, value := range in {
		value.ValidatorSubset = append([]string(nil), value.ValidatorSubset...)
		out[key] = value
	}
	return out
}

func cloneReceipts(in map[string]Receipt) map[string]Receipt {
	out := make(map[string]Receipt, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneLoadStates(in map[int32]WorkchainLoadState) map[int32]WorkchainLoadState {
	out := make(map[int32]WorkchainLoadState, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func mergeReceipts(left, right map[string]Receipt) map[string]Receipt {
	out := cloneReceipts(left)
	for key, value := range right {
		out[key] = value
	}
	return out
}

func max(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}
