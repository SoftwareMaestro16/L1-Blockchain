package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	MaxBlockSTMItems		= 100_000
	MaxBlockSTMAccessKeys		= 256
	MaxBlockSTMRemoteMessages	= 256
)

type BlockSTMExecutionItem struct {
	TxID			string
	TxIndex			uint32
	MessageIndex		uint32
	ZoneID			string
	ShardID			string
	ObjectID		string
	ObjectVersion		uint64
	FeeAmount		string
	ReadKeys		[]string
	WriteKeys		[]string
	RemoteWrites		[]BlockSTMRemoteWrite
	LocalMultiStepOp	bool
}

type BlockSTMRemoteWrite struct {
	DestinationZoneID	string
	DestinationShardID	string
	ObjectID		string
	PayloadHash		string
}

type BlockSTMExecutionGroup struct {
	GroupID		string
	ZoneID		string
	ShardID		string
	ParallelBatch	uint32
	Items		[]BlockSTMExecutionItem
	ReadSetRoot	string
	WriteSetRoot	string
	ConflictKeySet	[]string
	GroupHash	string
}

type ShardFeeAccumulator struct {
	ZoneID		string
	ShardID		string
	Amount		string
	Count		uint64
	RootHash	string
}

type ShardMessageQueue struct {
	ZoneID		string
	ShardID		string
	Messages	[]ShardOutputMessage
	QueueHash	string
}

type ShardOutputMessage struct {
	MessageID		string
	SourceTxID		string
	SourceZoneID		string
	SourceShardID		string
	DestinationZoneID	string
	DestinationShardID	string
	ObjectID		string
	PayloadHash		string
	Sequence		uint64
	MessageHash		string
}

type VersionedObjectUpdate struct {
	ZoneID		string
	ShardID		string
	ObjectID	string
	ExpectedVersion	uint64
	NextVersion	uint64
	StateKey	string
	UpdateHash	string
}

type BlockSTMStrategyPlan struct {
	Height		uint64
	Groups		[]BlockSTMExecutionGroup
	FeeAccumulators	[]ShardFeeAccumulator
	MessageQueues	[]ShardMessageQueue
	ObjectUpdates	[]VersionedObjectUpdate
	PlanHash	string
}

func BuildBlockSTMStrategyPlan(height uint64, items []BlockSTMExecutionItem) (BlockSTMStrategyPlan, error) {
	if height == 0 {
		return BlockSTMStrategyPlan{}, errors.New("performance BlockSTM height must be positive")
	}
	if len(items) == 0 {
		return BlockSTMStrategyPlan{}, errors.New("performance BlockSTM plan requires items")
	}
	if len(items) > MaxBlockSTMItems {
		return BlockSTMStrategyPlan{}, errors.New("performance BlockSTM plan exceeds max items")
	}
	ordered := normalizeBlockSTMItems(items)
	seenTx := make(map[string]struct{}, len(ordered))
	for _, item := range ordered {
		if err := item.Validate(); err != nil {
			return BlockSTMStrategyPlan{}, err
		}
		key := itemOrderingKey(item)
		if _, found := seenTx[key]; found {
			return BlockSTMStrategyPlan{}, errors.New("performance BlockSTM duplicate item position")
		}
		seenTx[key] = struct{}{}
	}
	groups, err := buildBlockSTMGroups(ordered)
	if err != nil {
		return BlockSTMStrategyPlan{}, err
	}
	fees, err := buildShardFeeAccumulators(ordered)
	if err != nil {
		return BlockSTMStrategyPlan{}, err
	}
	queues, err := buildShardMessageQueues(ordered)
	if err != nil {
		return BlockSTMStrategyPlan{}, err
	}
	updates := buildVersionedObjectUpdates(ordered)
	plan := BlockSTMStrategyPlan{
		Height:			height,
		Groups:			groups,
		FeeAccumulators:	fees,
		MessageQueues:		queues,
		ObjectUpdates:		updates,
	}
	plan.PlanHash = ComputeBlockSTMStrategyPlanHash(plan)
	return plan, plan.Validate()
}

func (i BlockSTMExecutionItem) Normalize() BlockSTMExecutionItem {
	i.TxID = strings.TrimSpace(i.TxID)
	i.ZoneID = strings.TrimSpace(i.ZoneID)
	i.ShardID = strings.TrimSpace(i.ShardID)
	i.ObjectID = strings.TrimSpace(i.ObjectID)
	i.FeeAmount = strings.TrimSpace(i.FeeAmount)
	i.ReadKeys = normalizeStringSet(i.ReadKeys)
	i.WriteKeys = normalizeStringSet(i.WriteKeys)
	for j := range i.RemoteWrites {
		i.RemoteWrites[j] = i.RemoteWrites[j].Normalize()
	}
	sort.SliceStable(i.RemoteWrites, func(left, right int) bool {
		return remoteWriteKey(i.RemoteWrites[left]) < remoteWriteKey(i.RemoteWrites[right])
	})
	return i
}

func (i BlockSTMExecutionItem) Validate() error {
	item := i.Normalize()
	if item.TxID == "" {
		return errors.New("performance BlockSTM tx id is required")
	}
	if err := validateExecutionToken("performance BlockSTM zone id", item.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance BlockSTM shard id", item.ShardID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance BlockSTM object id", item.ObjectID); err != nil {
		return err
	}
	if item.ObjectVersion == 0 {
		return errors.New("performance BlockSTM object version must be positive")
	}
	if item.FeeAmount == "" {
		item.FeeAmount = "0"
	}
	if _, err := parsePerformanceNonNegativeInt("performance BlockSTM fee amount", item.FeeAmount); err != nil {
		return err
	}
	if len(item.ReadKeys)+len(item.WriteKeys) == 0 {
		return errors.New("performance BlockSTM item requires read or write keys")
	}
	if len(item.ReadKeys)+len(item.WriteKeys) > MaxBlockSTMAccessKeys {
		return errors.New("performance BlockSTM access set exceeds max keys")
	}
	for _, key := range append(append([]string{}, item.ReadKeys...), item.WriteKeys...) {
		if err := validateStateAccessKey(item.ZoneID, item.ShardID, key); err != nil {
			return err
		}
	}
	if item.LocalMultiStepOp && len(item.WriteKeys) == 0 {
		return errors.New("performance BlockSTM local multi-step operation requires object-local write lock")
	}
	for _, key := range item.WriteKeys {
		if isGlobalHotKey(key) {
			return errors.New("performance BlockSTM hot path must not write global counters or locks")
		}
	}
	if len(item.RemoteWrites) > MaxBlockSTMRemoteMessages {
		return errors.New("performance BlockSTM remote writes exceed max messages")
	}
	for _, remote := range item.RemoteWrites {
		if err := remote.Validate(); err != nil {
			return err
		}
		if remote.DestinationZoneID == item.ZoneID && remote.DestinationShardID == item.ShardID {
			return errors.New("performance BlockSTM local writes must stay in write set, not remote queue")
		}
	}
	return nil
}

func (w BlockSTMRemoteWrite) Normalize() BlockSTMRemoteWrite {
	w.DestinationZoneID = strings.TrimSpace(w.DestinationZoneID)
	w.DestinationShardID = strings.TrimSpace(w.DestinationShardID)
	w.ObjectID = strings.TrimSpace(w.ObjectID)
	w.PayloadHash = normalizeLowerHex(w.PayloadHash)
	return w
}

func (w BlockSTMRemoteWrite) Validate() error {
	remote := w.Normalize()
	if err := validateExecutionToken("performance BlockSTM remote zone id", remote.DestinationZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance BlockSTM remote shard id", remote.DestinationShardID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance BlockSTM remote object id", remote.ObjectID); err != nil {
		return err
	}
	return validateHexHash("performance BlockSTM remote payload hash", remote.PayloadHash)
}

func (p BlockSTMStrategyPlan) Validate() error {
	if p.Height == 0 {
		return errors.New("performance BlockSTM plan height must be positive")
	}
	if len(p.Groups) == 0 {
		return errors.New("performance BlockSTM plan requires groups")
	}
	previousGroup := ""
	for _, group := range p.Groups {
		group = group.Normalize()
		if err := group.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%020d/%s/%s/%s", group.ParallelBatch, group.ZoneID, group.ShardID, group.GroupID)
		if previousGroup != "" && previousGroup >= key {
			return errors.New("performance BlockSTM groups must be sorted canonically")
		}
		previousGroup = key
	}
	for _, accumulator := range p.FeeAccumulators {
		if err := accumulator.Validate(); err != nil {
			return err
		}
	}
	for _, queue := range p.MessageQueues {
		if err := queue.Validate(); err != nil {
			return err
		}
	}
	for _, update := range p.ObjectUpdates {
		if err := update.Validate(); err != nil {
			return err
		}
	}
	if p.PlanHash != ComputeBlockSTMStrategyPlanHash(p) {
		return errors.New("performance BlockSTM plan hash mismatch")
	}
	return nil
}

func (g BlockSTMExecutionGroup) Normalize() BlockSTMExecutionGroup {
	g.GroupID = normalizeLowerHex(g.GroupID)
	g.ZoneID = strings.TrimSpace(g.ZoneID)
	g.ShardID = strings.TrimSpace(g.ShardID)
	g.Items = normalizeBlockSTMItems(g.Items)
	g.ConflictKeySet = normalizeStringSet(g.ConflictKeySet)
	g.ReadSetRoot = normalizeLowerHex(g.ReadSetRoot)
	g.WriteSetRoot = normalizeLowerHex(g.WriteSetRoot)
	g.GroupHash = normalizeLowerHex(g.GroupHash)
	return g
}

func (g BlockSTMExecutionGroup) Validate() error {
	group := g.Normalize()
	if err := validateHexHash("performance BlockSTM group id", group.GroupID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance BlockSTM group zone id", group.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance BlockSTM group shard id", group.ShardID); err != nil {
		return err
	}
	if len(group.Items) == 0 {
		return errors.New("performance BlockSTM group requires items")
	}
	for _, item := range group.Items {
		if item.ZoneID != group.ZoneID || item.ShardID != group.ShardID {
			return errors.New("performance BlockSTM group item zone or shard mismatch")
		}
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if hasAccessConflicts(group.Items) {
		return errors.New("performance BlockSTM group contains conflicting access sets")
	}
	if group.ReadSetRoot != computeAccessRoot(group.Items, true) {
		return errors.New("performance BlockSTM read set root mismatch")
	}
	if group.WriteSetRoot != computeAccessRoot(group.Items, false) {
		return errors.New("performance BlockSTM write set root mismatch")
	}
	if group.GroupHash != ComputeBlockSTMGroupHash(group) {
		return errors.New("performance BlockSTM group hash mismatch")
	}
	return nil
}

func (a ShardFeeAccumulator) Normalize() ShardFeeAccumulator {
	a.ZoneID = strings.TrimSpace(a.ZoneID)
	a.ShardID = strings.TrimSpace(a.ShardID)
	a.Amount = strings.TrimSpace(a.Amount)
	a.RootHash = normalizeLowerHex(a.RootHash)
	return a
}

func (a ShardFeeAccumulator) Validate() error {
	accumulator := a.Normalize()
	if err := validateExecutionToken("performance shard fee zone id", accumulator.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance shard fee shard id", accumulator.ShardID); err != nil {
		return err
	}
	if _, err := parsePerformanceNonNegativeInt("performance shard fee amount", accumulator.Amount); err != nil {
		return err
	}
	if accumulator.Count == 0 {
		return errors.New("performance shard fee accumulator count must be positive")
	}
	if accumulator.RootHash != ComputeShardFeeAccumulatorHash(accumulator) {
		return errors.New("performance shard fee accumulator hash mismatch")
	}
	return nil
}

func (q ShardMessageQueue) Normalize() ShardMessageQueue {
	q.ZoneID = strings.TrimSpace(q.ZoneID)
	q.ShardID = strings.TrimSpace(q.ShardID)
	for i := range q.Messages {
		q.Messages[i] = q.Messages[i].Normalize()
	}
	sort.SliceStable(q.Messages, func(i, j int) bool {
		return q.Messages[i].Sequence < q.Messages[j].Sequence
	})
	q.QueueHash = normalizeLowerHex(q.QueueHash)
	return q
}

func (q ShardMessageQueue) Validate() error {
	queue := q.Normalize()
	if err := validateExecutionToken("performance shard queue zone id", queue.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance shard queue shard id", queue.ShardID); err != nil {
		return err
	}
	for i, message := range queue.Messages {
		if err := message.Validate(); err != nil {
			return err
		}
		if message.DestinationZoneID != queue.ZoneID || message.DestinationShardID != queue.ShardID {
			return errors.New("performance shard queue destination mismatch")
		}
		if uint64(i+1) != message.Sequence {
			return errors.New("performance shard queue sequence must be contiguous")
		}
	}
	if queue.QueueHash != ComputeShardMessageQueueHash(queue) {
		return errors.New("performance shard queue hash mismatch")
	}
	return nil
}

func (m ShardOutputMessage) Normalize() ShardOutputMessage {
	m.MessageID = normalizeLowerHex(m.MessageID)
	m.SourceTxID = strings.TrimSpace(m.SourceTxID)
	m.SourceZoneID = strings.TrimSpace(m.SourceZoneID)
	m.SourceShardID = strings.TrimSpace(m.SourceShardID)
	m.DestinationZoneID = strings.TrimSpace(m.DestinationZoneID)
	m.DestinationShardID = strings.TrimSpace(m.DestinationShardID)
	m.ObjectID = strings.TrimSpace(m.ObjectID)
	m.PayloadHash = normalizeLowerHex(m.PayloadHash)
	m.MessageHash = normalizeLowerHex(m.MessageHash)
	return m
}

func (m ShardOutputMessage) Validate() error {
	message := m.Normalize()
	if err := validateHexHash("performance shard message id", message.MessageID); err != nil {
		return err
	}
	if message.SourceTxID == "" {
		return errors.New("performance shard message source tx id is required")
	}
	if err := validateExecutionToken("performance shard message source zone id", message.SourceZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance shard message source shard id", message.SourceShardID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance shard message destination zone id", message.DestinationZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance shard message destination shard id", message.DestinationShardID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance shard message object id", message.ObjectID); err != nil {
		return err
	}
	if err := validateHexHash("performance shard message payload hash", message.PayloadHash); err != nil {
		return err
	}
	if message.Sequence == 0 {
		return errors.New("performance shard message sequence must be positive")
	}
	if message.MessageHash != ComputeShardOutputMessageHash(message) {
		return errors.New("performance shard message hash mismatch")
	}
	if message.MessageID != hashStrings("performance-shard-message-id", message.MessageHash) {
		return errors.New("performance shard message id mismatch")
	}
	return nil
}

func (u VersionedObjectUpdate) Normalize() VersionedObjectUpdate {
	u.ZoneID = strings.TrimSpace(u.ZoneID)
	u.ShardID = strings.TrimSpace(u.ShardID)
	u.ObjectID = strings.TrimSpace(u.ObjectID)
	u.StateKey = strings.TrimSpace(u.StateKey)
	u.UpdateHash = normalizeLowerHex(u.UpdateHash)
	return u
}

func (u VersionedObjectUpdate) Validate() error {
	update := u.Normalize()
	if err := validateExecutionToken("performance versioned update zone id", update.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance versioned update shard id", update.ShardID); err != nil {
		return err
	}
	if err := validateExecutionToken("performance versioned update object id", update.ObjectID); err != nil {
		return err
	}
	if update.ExpectedVersion == 0 {
		return errors.New("performance versioned update expected version must be positive")
	}
	if update.NextVersion != update.ExpectedVersion+1 {
		return errors.New("performance versioned update next version must increment by one")
	}
	expectedKey := VersionedObjectStateKey(update.ZoneID, update.ShardID, update.ObjectID, update.ExpectedVersion)
	if update.StateKey != expectedKey {
		return errors.New("performance versioned update state key mismatch")
	}
	if update.UpdateHash != ComputeVersionedObjectUpdateHash(update) {
		return errors.New("performance versioned update hash mismatch")
	}
	return nil
}

func VersionedObjectStateKey(zoneID, shardID, objectID string, version uint64) string {
	return fmt.Sprintf("zone/%s/shard/%s/object/%s/version/%020d", strings.TrimSpace(zoneID), strings.TrimSpace(shardID), strings.TrimSpace(objectID), version)
}

func ComputeBlockSTMStrategyPlanHash(plan BlockSTMStrategyPlan) string {
	parts := []string{"performance-blockstm-plan", fmt.Sprintf("%020d", plan.Height)}
	for _, group := range plan.Groups {
		parts = append(parts, group.GroupHash)
	}
	for _, accumulator := range plan.FeeAccumulators {
		parts = append(parts, accumulator.RootHash)
	}
	for _, queue := range plan.MessageQueues {
		parts = append(parts, queue.QueueHash)
	}
	for _, update := range plan.ObjectUpdates {
		parts = append(parts, update.UpdateHash)
	}
	return hashStrings(parts...)
}

func ComputeBlockSTMGroupHash(group BlockSTMExecutionGroup) string {
	group = group.Normalize()
	parts := []string{
		"performance-blockstm-group",
		group.ZoneID,
		group.ShardID,
		fmt.Sprintf("%020d", uint64(group.ParallelBatch)),
		group.ReadSetRoot,
		group.WriteSetRoot,
	}
	for _, key := range group.ConflictKeySet {
		parts = append(parts, key)
	}
	for _, item := range group.Items {
		parts = append(parts, itemOrderingKey(item), item.ObjectID, fmt.Sprintf("%020d", item.ObjectVersion))
	}
	return hashStrings(parts...)
}

func ComputeShardFeeAccumulatorHash(accumulator ShardFeeAccumulator) string {
	accumulator = accumulator.Normalize()
	return hashStrings(
		"performance-shard-fee-accumulator",
		accumulator.ZoneID,
		accumulator.ShardID,
		accumulator.Amount,
		fmt.Sprintf("%020d", accumulator.Count),
	)
}

func ComputeShardMessageQueueHash(queue ShardMessageQueue) string {
	queue = queue.Normalize()
	parts := []string{"performance-shard-message-queue", queue.ZoneID, queue.ShardID}
	for _, message := range queue.Messages {
		parts = append(parts, message.MessageHash)
	}
	return hashStrings(parts...)
}

func ComputeShardOutputMessageHash(message ShardOutputMessage) string {
	message = message.Normalize()
	return hashStrings(
		"performance-shard-output-message",
		message.SourceTxID,
		message.SourceZoneID,
		message.SourceShardID,
		message.DestinationZoneID,
		message.DestinationShardID,
		message.ObjectID,
		message.PayloadHash,
		fmt.Sprintf("%020d", message.Sequence),
	)
}

func ComputeVersionedObjectUpdateHash(update VersionedObjectUpdate) string {
	update = update.Normalize()
	return hashStrings(
		"performance-versioned-object-update",
		update.ZoneID,
		update.ShardID,
		update.ObjectID,
		fmt.Sprintf("%020d", update.ExpectedVersion),
		fmt.Sprintf("%020d", update.NextVersion),
		update.StateKey,
	)
}

func buildBlockSTMGroups(items []BlockSTMExecutionItem) ([]BlockSTMExecutionGroup, error) {
	batches := make([][]BlockSTMExecutionItem, 0)
	for _, item := range items {
		placed := false
		for i := range batches {
			if !hasAccessConflicts(append(append([]BlockSTMExecutionItem{}, batches[i]...), item)) {
				batches[i] = append(batches[i], item)
				placed = true
				break
			}
		}
		if !placed {
			batches = append(batches, []BlockSTMExecutionItem{item})
		}
	}
	groups := make([]BlockSTMExecutionGroup, 0, len(items))
	for batchIndex, batch := range batches {
		byShard := make(map[string][]BlockSTMExecutionItem)
		keys := make([]string, 0)
		for _, item := range batch {
			key := item.ZoneID + "/" + item.ShardID
			if _, found := byShard[key]; !found {
				keys = append(keys, key)
			}
			byShard[key] = append(byShard[key], item)
		}
		sort.Strings(keys)
		for _, key := range keys {
			shardItems := normalizeBlockSTMItems(byShard[key])
			group := BlockSTMExecutionGroup{
				ZoneID:		shardItems[0].ZoneID,
				ShardID:	shardItems[0].ShardID,
				ParallelBatch:	uint32(batchIndex + 1),
				Items:		shardItems,
				ReadSetRoot:	computeAccessRoot(shardItems, true),
				WriteSetRoot:	computeAccessRoot(shardItems, false),
				ConflictKeySet:	conflictKeys(shardItems),
			}
			group.GroupHash = ComputeBlockSTMGroupHash(group)
			group.GroupID = hashStrings("performance-blockstm-group-id", group.GroupHash)
			groups = append(groups, group.Normalize())
		}
	}
	sort.SliceStable(groups, func(i, j int) bool {
		left := fmt.Sprintf("%020d/%s/%s/%s", groups[i].ParallelBatch, groups[i].ZoneID, groups[i].ShardID, groups[i].GroupID)
		right := fmt.Sprintf("%020d/%s/%s/%s", groups[j].ParallelBatch, groups[j].ZoneID, groups[j].ShardID, groups[j].GroupID)
		return left < right
	})
	return groups, nil
}

func buildShardFeeAccumulators(items []BlockSTMExecutionItem) ([]ShardFeeAccumulator, error) {
	type bucket struct {
		total	sdkmath.Int
		count	uint64
	}
	buckets := make(map[string]bucket)
	keys := make([]string, 0)
	for _, item := range items {
		fee, err := parsePerformanceNonNegativeInt("performance BlockSTM fee amount", item.FeeAmount)
		if err != nil {
			return nil, err
		}
		key := item.ZoneID + "/" + item.ShardID
		value := buckets[key]
		if value.count == 0 {
			keys = append(keys, key)
			value.total = sdkmath.ZeroInt()
		}
		value.total = value.total.Add(fee)
		value.count++
		buckets[key] = value
	}
	sort.Strings(keys)
	out := make([]ShardFeeAccumulator, 0, len(keys))
	for _, key := range keys {
		parts := strings.SplitN(key, "/", 2)
		value := buckets[key]
		accumulator := ShardFeeAccumulator{ZoneID: parts[0], ShardID: parts[1], Amount: value.total.String(), Count: value.count}
		accumulator.RootHash = ComputeShardFeeAccumulatorHash(accumulator)
		out = append(out, accumulator.Normalize())
	}
	return out, nil
}

func buildShardMessageQueues(items []BlockSTMExecutionItem) ([]ShardMessageQueue, error) {
	queues := make(map[string][]ShardOutputMessage)
	keys := make([]string, 0)
	for _, item := range items {
		for _, remote := range item.RemoteWrites {
			key := remote.DestinationZoneID + "/" + remote.DestinationShardID
			if _, found := queues[key]; !found {
				keys = append(keys, key)
			}
			sequence := uint64(len(queues[key]) + 1)
			message := ShardOutputMessage{
				SourceTxID:		item.TxID,
				SourceZoneID:		item.ZoneID,
				SourceShardID:		item.ShardID,
				DestinationZoneID:	remote.DestinationZoneID,
				DestinationShardID:	remote.DestinationShardID,
				ObjectID:		remote.ObjectID,
				PayloadHash:		remote.PayloadHash,
				Sequence:		sequence,
			}
			message.MessageHash = ComputeShardOutputMessageHash(message)
			message.MessageID = hashStrings("performance-shard-message-id", message.MessageHash)
			queues[key] = append(queues[key], message.Normalize())
		}
	}
	sort.Strings(keys)
	out := make([]ShardMessageQueue, 0, len(keys))
	for _, key := range keys {
		parts := strings.SplitN(key, "/", 2)
		queue := ShardMessageQueue{ZoneID: parts[0], ShardID: parts[1], Messages: queues[key]}
		queue.QueueHash = ComputeShardMessageQueueHash(queue)
		out = append(out, queue.Normalize())
	}
	return out, nil
}

func buildVersionedObjectUpdates(items []BlockSTMExecutionItem) []VersionedObjectUpdate {
	updates := make([]VersionedObjectUpdate, 0, len(items))
	seen := make(map[string]struct{})
	for _, item := range items {
		if len(item.WriteKeys) == 0 {
			continue
		}
		key := item.ZoneID + "/" + item.ShardID + "/" + item.ObjectID
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		update := VersionedObjectUpdate{
			ZoneID:			item.ZoneID,
			ShardID:		item.ShardID,
			ObjectID:		item.ObjectID,
			ExpectedVersion:	item.ObjectVersion,
			NextVersion:		item.ObjectVersion + 1,
			StateKey:		VersionedObjectStateKey(item.ZoneID, item.ShardID, item.ObjectID, item.ObjectVersion),
		}
		update.UpdateHash = ComputeVersionedObjectUpdateHash(update)
		updates = append(updates, update.Normalize())
	}
	sort.SliceStable(updates, func(i, j int) bool {
		return updates[i].StateKey < updates[j].StateKey
	})
	return updates
}

func normalizeBlockSTMItems(items []BlockSTMExecutionItem) []BlockSTMExecutionItem {
	out := make([]BlockSTMExecutionItem, len(items))
	for i, item := range items {
		out[i] = item.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return itemOrderingKey(out[i]) < itemOrderingKey(out[j])
	})
	return out
}

func hasAccessConflicts(items []BlockSTMExecutionItem) bool {
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if accessConflicts(items[i], items[j]) {
				return true
			}
		}
	}
	return false
}

func accessConflicts(left, right BlockSTMExecutionItem) bool {
	if left.ZoneID == right.ZoneID && left.ShardID == right.ShardID && left.ObjectID == right.ObjectID && (len(left.WriteKeys) > 0 || len(right.WriteKeys) > 0) {
		return true
	}
	leftWrites := stringSet(left.WriteKeys)
	rightWrites := stringSet(right.WriteKeys)
	for key := range leftWrites {
		if _, found := rightWrites[key]; found {
			return true
		}
	}
	for key := range leftWrites {
		if _, found := stringSet(right.ReadKeys)[key]; found {
			return true
		}
	}
	for key := range rightWrites {
		if _, found := stringSet(left.ReadKeys)[key]; found {
			return true
		}
	}
	return false
}

func computeAccessRoot(items []BlockSTMExecutionItem, reads bool) string {
	keys := make([]string, 0)
	for _, item := range items {
		if reads {
			keys = append(keys, item.ReadKeys...)
			continue
		}
		keys = append(keys, item.WriteKeys...)
	}
	keys = normalizeStringSet(keys)
	return hashStrings(append([]string{"performance-blockstm-access-root"}, keys...)...)
}

func conflictKeys(items []BlockSTMExecutionItem) []string {
	keys := make([]string, 0)
	for _, item := range items {
		keys = append(keys, item.ReadKeys...)
		keys = append(keys, item.WriteKeys...)
		if item.LocalMultiStepOp {
			keys = append(keys, "local-lock/"+item.ZoneID+"/"+item.ShardID+"/"+item.ObjectID)
		}
	}
	return normalizeStringSet(keys)
}

func itemOrderingKey(item BlockSTMExecutionItem) string {
	return fmt.Sprintf("%s/%s/%020d/%020d/%s", item.ZoneID, item.ShardID, item.TxIndex, item.MessageIndex, item.TxID)
}

func remoteWriteKey(remote BlockSTMRemoteWrite) string {
	return remote.DestinationZoneID + "/" + remote.DestinationShardID + "/" + remote.ObjectID + "/" + remote.PayloadHash
}

func validateExecutionToken(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > 96 {
		return fmt.Errorf("%s is too long", field)
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			continue
		}
		return fmt.Errorf("%s contains unsupported character", field)
	}
	return nil
}

func validateStateAccessKey(zoneID, shardID, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("performance BlockSTM state access key is required")
	}
	if strings.HasPrefix(key, "global/") || strings.HasPrefix(key, "core/global/") || strings.Contains(key, "/global-counter/") {
		return errors.New("performance BlockSTM state access key must not use global hot path")
	}
	prefix := "zone/" + zoneID + "/shard/" + shardID + "/"
	if !strings.HasPrefix(key, prefix) {
		return errors.New("performance BlockSTM state access key must stay within zone and shard")
	}
	return nil
}

func isGlobalHotKey(key string) bool {
	return strings.HasPrefix(key, "global/") ||
		strings.HasPrefix(key, "core/global/") ||
		strings.Contains(key, "/counter/") ||
		strings.Contains(key, "/global-counter/") ||
		strings.Contains(key, "/global-lock/")
}

func parsePerformanceNonNegativeInt(field, value string) (sdkmath.Int, error) {
	out, ok := sdkmath.NewIntFromString(strings.TrimSpace(value))
	if !ok || out.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf("%s must be a non-negative integer", field)
	}
	return out, nil
}

func normalizeStringSet(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
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

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func validateHexHash(field, value string) error {
	if len(value) != 64 {
		return fmt.Errorf("%s must be 64 lowercase hex chars", field)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be 64 lowercase hex chars", field)
	}
	return nil
}

func normalizeLowerHex(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func hashStrings(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(fmt.Sprintf("%020d", len(part))))
		_, _ = h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}
