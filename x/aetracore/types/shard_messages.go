package types

import (
	"errors"
	"fmt"
	"sort"
)

type ShardQueueKind string

const (
	ShardQueueInbox		ShardQueueKind	= "inbox"
	ShardQueueOutbox	ShardQueueKind	= "outbox"
)

type ShardMigrationStatus string

const (
	ShardMigrationStatusExecuted ShardMigrationStatus = "executed"
)

type ShardMessageEnvelope struct {
	MsgID			string
	TraceID			string
	Sender			string
	Receiver		string
	SenderZoneID		ZoneID
	SenderShardID		ShardID
	ReceiverZoneID		ZoneID
	ReceiverShardID		ShardID
	DestinationStateKey	string
	PayloadType		string
	PayloadHash		string
	Priority		uint32
	AdmissionHeight		uint64
	CreatedAtHeight		uint64
	ExpiryHeight		uint64
	MessageIndex		uint32
	SourceLayoutEpoch	uint64
	DeliveryLayoutEpoch	uint64
	MessageHash		string
}

type ShardMessageStore struct {
	ZoneID		ZoneID
	ShardID		ShardID
	Height		uint64
	QueueKind	ShardQueueKind
	Messages	[]ShardMessageEnvelope
	StoreRoot	string
}

type ShardMessageStoreEntry struct {
	Key		string
	MsgID		string
	MessageHash	string
}

type ShardMigrationReceipt struct {
	TaskID			string
	TaskHash		string
	ZoneID			ZoneID
	SourceShardID		ShardID
	DestinationShardID	ShardID
	SourceLayoutEpoch	uint64
	TargetLayoutEpoch	uint64
	DeliveryEpoch		uint64
	Height			uint64
	Status			ShardMigrationStatus
	StateRootBefore		string
	StateRootAfter		string
	ReceiptHash		string
}

func NewShardMessageEnvelope(msg ShardMessageEnvelope) (ShardMessageEnvelope, error) {
	if msg.MessageHash != "" {
		return ShardMessageEnvelope{}, errors.New("aetracore shard message hash must be empty before construction")
	}
	if err := msg.ValidateFormat(); err != nil {
		return ShardMessageEnvelope{}, err
	}
	msg.MessageHash = ComputeShardMessageHash(msg)
	return msg, msg.ValidateHash()
}

func (m ShardMessageEnvelope) ValidateHash() error {
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.MessageHash != ComputeShardMessageHash(m) {
		return errors.New("aetracore shard message hash mismatch")
	}
	return nil
}

func (m ShardMessageEnvelope) ValidateFormat() error {
	if err := validateToken("aetracore shard message id", m.MsgID, MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore shard message trace id", m.TraceID, MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore shard message sender", m.Sender, MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore shard message receiver", m.Receiver, MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateZoneID(m.SenderZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(m.SenderShardID); err != nil {
		return err
	}
	if err := ValidateZoneID(m.ReceiverZoneID); err != nil {
		return err
	}
	if m.ReceiverShardID != "" {
		if err := ValidateShardID(m.ReceiverShardID); err != nil {
			return err
		}
	}
	if err := validateToken("aetracore shard message destination key", m.DestinationStateKey, MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore shard message payload type", m.PayloadType, MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateHash("aetracore shard message payload hash", m.PayloadHash); err != nil {
		return err
	}
	if m.AdmissionHeight == 0 || m.CreatedAtHeight == 0 || m.ExpiryHeight == 0 {
		return errors.New("aetracore shard message heights must be positive")
	}
	if m.ExpiryHeight <= m.CreatedAtHeight {
		return errors.New("aetracore shard message expiry height must be future")
	}
	if m.SourceLayoutEpoch == 0 || m.DeliveryLayoutEpoch == 0 {
		return errors.New("aetracore shard message layout epochs must be positive")
	}
	if m.MessageHash != "" {
		return ValidateHash("aetracore shard message hash", m.MessageHash)
	}
	return nil
}

func NewShardMessageStore(store ShardMessageStore) (ShardMessageStore, error) {
	if store.StoreRoot != "" {
		return ShardMessageStore{}, errors.New("aetracore shard message store root must be empty before construction")
	}
	normalized := store
	normalized.Messages = cloneShardMessages(store.Messages)
	sortShardMessages(normalized.Messages)
	if err := normalized.ValidateFormat(); err != nil {
		return ShardMessageStore{}, err
	}
	normalized.StoreRoot = ComputeShardMessageStoreRoot(normalized)
	return normalized, normalized.ValidateHash()
}

func (s ShardMessageStore) ValidateHash() error {
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.StoreRoot != ComputeShardMessageStoreRoot(s) {
		return errors.New("aetracore shard message store root mismatch")
	}
	return nil
}

func (s ShardMessageStore) ValidateFormat() error {
	if err := ValidateZoneID(s.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(s.ShardID); err != nil {
		return err
	}
	if s.Height == 0 {
		return errors.New("aetracore shard message store height must be positive")
	}
	if !IsShardQueueKind(s.QueueKind) {
		return fmt.Errorf("unknown aetracore shard queue kind %q", s.QueueKind)
	}
	for i, msg := range s.Messages {
		if err := msg.ValidateHash(); err != nil {
			return err
		}
		if err := validateShardMessageStoreRoute(s, msg); err != nil {
			return err
		}
		if i > 0 && compareShardMessages(s.Messages[i-1], msg) >= 0 {
			return errors.New("aetracore shard messages must be sorted canonically")
		}
	}
	if s.StoreRoot != "" {
		return ValidateHash("aetracore shard message store root", s.StoreRoot)
	}
	return nil
}

func BuildShardMessageStoreEntries(store ShardMessageStore) ([]ShardMessageStoreEntry, error) {
	if err := store.ValidateHash(); err != nil {
		return nil, err
	}
	entries := make([]ShardMessageStoreEntry, len(store.Messages))
	for i, msg := range store.Messages {
		var key string
		var err error
		switch store.QueueKind {
		case ShardQueueInbox:
			key, err = ShardInboxKey(store.ZoneID, store.ShardID, msg.MsgID)
		case ShardQueueOutbox:
			key, err = ShardOutboxKey(store.ZoneID, store.ShardID, msg.MsgID)
		default:
			err = fmt.Errorf("unknown aetracore shard queue kind %q", store.QueueKind)
		}
		if err != nil {
			return nil, err
		}
		entries[i] = ShardMessageStoreEntry{Key: key, MsgID: msg.MsgID, MessageHash: msg.MessageHash}
	}
	return entries, nil
}

func ResolveShardMessageDeliveryRoute(layouts []ShardLayout, msg ShardMessageEnvelope) (ShardRoute, error) {
	if err := msg.ValidateHash(); err != nil {
		return ShardRoute{}, err
	}
	layout, found := SelectShardLayoutForEpoch(layouts, msg.ReceiverZoneID, msg.DeliveryLayoutEpoch)
	if !found {
		return ShardRoute{}, errors.New("aetracore shard message delivery missing committed layout")
	}
	return RouteKeyToShard(layout, ShardRoutingInput{
		ZoneID:			msg.ReceiverZoneID,
		StateKey:		msg.DestinationStateKey,
		ShardLayoutEpoch:	layout.LayoutEpoch,
		PlacementOverride:	msg.ReceiverShardID,
	})
}

func SelectShardLayoutForEpoch(layouts []ShardLayout, zoneID ZoneID, layoutEpoch uint64) (ShardLayout, bool) {
	ordered := cloneShardLayouts(layouts)
	sortShardLayouts(ordered)
	var selected ShardLayout
	found := false
	for _, layout := range ordered {
		if layout.ZoneID != zoneID || layout.LayoutEpoch > layoutEpoch {
			continue
		}
		if !found || layout.LayoutEpoch > selected.LayoutEpoch {
			selected = layout
			found = true
		}
	}
	return selected, found
}

func ExecuteShardMigrationTasks(tasks []ShardMigrationTask, height uint64, stateRootBefore string) ([]ShardMigrationReceipt, string, error) {
	if height == 0 {
		return nil, "", errors.New("aetracore shard migration execution height must be positive")
	}
	if err := ValidateHash("aetracore shard migration state root before", stateRootBefore); err != nil {
		return nil, "", err
	}
	ordered := cloneShardMigrationTasks(tasks)
	sortShardMigrationTasks(ordered)
	receipts := make([]ShardMigrationReceipt, len(ordered))
	currentRoot := stateRootBefore
	for i, task := range ordered {
		if err := task.ValidateHash(); err != nil {
			return nil, "", err
		}
		receipt := ShardMigrationReceipt{
			TaskID:			task.TaskID,
			TaskHash:		task.TaskHash,
			ZoneID:			task.ZoneID,
			SourceShardID:		task.SourceShardID,
			DestinationShardID:	task.DestinationShardID,
			SourceLayoutEpoch:	task.SourceLayoutEpoch,
			TargetLayoutEpoch:	task.TargetLayoutEpoch,
			DeliveryEpoch:		task.DeliveryEpoch,
			Height:			height,
			Status:			ShardMigrationStatusExecuted,
			StateRootBefore:	currentRoot,
		}
		receipt.StateRootAfter = hashParts("aetra-aek-shard-migration-state-root-v1", currentRoot, task.TaskHash, fmt.Sprint(height))
		receipt.ReceiptHash = ComputeShardMigrationReceiptHash(receipt)
		if err := receipt.ValidateHash(); err != nil {
			return nil, "", err
		}
		receipts[i] = receipt
		currentRoot = receipt.StateRootAfter
	}
	root, err := ComputeShardMigrationReceiptRoot(receipts)
	if err != nil {
		return nil, "", err
	}
	return receipts, root, nil
}

func (r ShardMigrationReceipt) ValidateHash() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash != ComputeShardMigrationReceiptHash(r) {
		return errors.New("aetracore shard migration receipt hash mismatch")
	}
	return nil
}

func (r ShardMigrationReceipt) ValidateFormat() error {
	if err := validateToken("aetracore shard migration receipt task id", r.TaskID, MaxIDLength); err != nil {
		return err
	}
	if err := ValidateHash("aetracore shard migration receipt task hash", r.TaskHash); err != nil {
		return err
	}
	if err := ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(r.SourceShardID); err != nil {
		return err
	}
	if err := ValidateShardID(r.DestinationShardID); err != nil {
		return err
	}
	if r.SourceLayoutEpoch == 0 || r.TargetLayoutEpoch == 0 || r.DeliveryEpoch == 0 {
		return errors.New("aetracore shard migration receipt epochs must be positive")
	}
	if r.TargetLayoutEpoch <= r.SourceLayoutEpoch {
		return errors.New("aetracore shard migration receipt target epoch must be future")
	}
	if r.DeliveryEpoch < r.TargetLayoutEpoch {
		return errors.New("aetracore shard migration receipt delivery epoch must not precede target epoch")
	}
	if r.Height == 0 {
		return errors.New("aetracore shard migration receipt height must be positive")
	}
	if !IsShardMigrationStatus(r.Status) {
		return fmt.Errorf("unknown aetracore shard migration status %q", r.Status)
	}
	if err := ValidateHash("aetracore shard migration receipt state root before", r.StateRootBefore); err != nil {
		return err
	}
	if err := ValidateHash("aetracore shard migration receipt state root after", r.StateRootAfter); err != nil {
		return err
	}
	if r.ReceiptHash != "" {
		return ValidateHash("aetracore shard migration receipt hash", r.ReceiptHash)
	}
	return nil
}

func BuildZoneCommitmentFromShardRoots(
	height uint64,
	zoneID ZoneID,
	shardRoots []ShardRoot,
	stateRoot string,
	inboxRoot string,
	outboxRoot string,
	receiptsRoot string,
	eventsRoot string,
	paramsHash string,
	executionSummaryHash string,
) (ZoneCommitment, error) {
	shardRootsRoot, err := ComputeShardRootsRoot(shardRoots)
	if err != nil {
		return ZoneCommitment{}, err
	}
	return NewZoneCommitment(height, zoneID, stateRoot, inboxRoot, outboxRoot, receiptsRoot, eventsRoot, shardRootsRoot, paramsHash, executionSummaryHash)
}

func ComputeShardMessageHash(msg ShardMessageEnvelope) string {
	return hashParts(
		"aetra-aek-shard-message-v1",
		msg.MsgID,
		msg.TraceID,
		msg.Sender,
		msg.Receiver,
		string(msg.SenderZoneID),
		string(msg.SenderShardID),
		string(msg.ReceiverZoneID),
		string(msg.ReceiverShardID),
		msg.DestinationStateKey,
		msg.PayloadType,
		msg.PayloadHash,
		fmt.Sprint(msg.Priority),
		fmt.Sprint(msg.AdmissionHeight),
		fmt.Sprint(msg.CreatedAtHeight),
		fmt.Sprint(msg.ExpiryHeight),
		fmt.Sprint(msg.MessageIndex),
		fmt.Sprint(msg.SourceLayoutEpoch),
		fmt.Sprint(msg.DeliveryLayoutEpoch),
	)
}

func ComputeShardMessageStoreRoot(store ShardMessageStore) string {
	ordered := cloneShardMessages(store.Messages)
	sortShardMessages(ordered)
	parts := []string{
		"aetra-aek-shard-message-store-v1",
		string(store.ZoneID),
		string(store.ShardID),
		fmt.Sprint(store.Height),
		string(store.QueueKind),
		fmt.Sprint(len(ordered)),
	}
	for _, msg := range ordered {
		parts = append(parts, msg.MsgID, msg.MessageHash)
	}
	return hashParts(parts...)
}

func ComputeShardMigrationReceiptHash(receipt ShardMigrationReceipt) string {
	return hashParts(
		"aetra-aek-shard-migration-receipt-v1",
		receipt.TaskID,
		receipt.TaskHash,
		string(receipt.ZoneID),
		string(receipt.SourceShardID),
		string(receipt.DestinationShardID),
		fmt.Sprint(receipt.SourceLayoutEpoch),
		fmt.Sprint(receipt.TargetLayoutEpoch),
		fmt.Sprint(receipt.DeliveryEpoch),
		fmt.Sprint(receipt.Height),
		string(receipt.Status),
		receipt.StateRootBefore,
		receipt.StateRootAfter,
	)
}

func ComputeShardMigrationReceiptRoot(receipts []ShardMigrationReceipt) (string, error) {
	ordered := cloneShardMigrationReceipts(receipts)
	sortShardMigrationReceipts(ordered)
	parts := []string{"aetra-aek-shard-migration-receipt-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		if err := receipt.ValidateHash(); err != nil {
			return "", err
		}
		parts = append(parts, receipt.ReceiptHash)
	}
	return hashParts(parts...), nil
}

func IsShardQueueKind(kind ShardQueueKind) bool {
	switch kind {
	case ShardQueueInbox, ShardQueueOutbox:
		return true
	default:
		return false
	}
}

func IsShardMigrationStatus(status ShardMigrationStatus) bool {
	switch status {
	case ShardMigrationStatusExecuted:
		return true
	default:
		return false
	}
}

func validateShardMessageStoreRoute(store ShardMessageStore, msg ShardMessageEnvelope) error {
	switch store.QueueKind {
	case ShardQueueInbox:
		if msg.ReceiverZoneID != store.ZoneID {
			return errors.New("aetracore shard inbox message zone mismatch")
		}
		if msg.ReceiverShardID != "" && msg.ReceiverShardID != store.ShardID {
			return errors.New("aetracore shard inbox message shard mismatch")
		}
	case ShardQueueOutbox:
		if msg.SenderZoneID != store.ZoneID || msg.SenderShardID != store.ShardID {
			return errors.New("aetracore shard outbox message route mismatch")
		}
	default:
		return fmt.Errorf("unknown aetracore shard queue kind %q", store.QueueKind)
	}
	return nil
}

func sortShardMessages(messages []ShardMessageEnvelope) {
	sort.SliceStable(messages, func(i, j int) bool { return compareShardMessages(messages[i], messages[j]) < 0 })
}

func compareShardMessages(left, right ShardMessageEnvelope) int {
	if left.Priority < right.Priority {
		return -1
	}
	if left.Priority > right.Priority {
		return 1
	}
	if left.AdmissionHeight < right.AdmissionHeight {
		return -1
	}
	if left.AdmissionHeight > right.AdmissionHeight {
		return 1
	}
	if left.MessageHash < right.MessageHash {
		return -1
	}
	if left.MessageHash > right.MessageHash {
		return 1
	}
	if left.MessageIndex < right.MessageIndex {
		return -1
	}
	if left.MessageIndex > right.MessageIndex {
		return 1
	}
	return 0
}

func sortShardMigrationReceipts(receipts []ShardMigrationReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool { return receipts[i].ReceiptHash < receipts[j].ReceiptHash })
}

func cloneShardMessages(messages []ShardMessageEnvelope) []ShardMessageEnvelope {
	out := make([]ShardMessageEnvelope, len(messages))
	copy(out, messages)
	return out
}

func cloneShardMigrationReceipts(receipts []ShardMigrationReceipt) []ShardMigrationReceipt {
	out := make([]ShardMigrationReceipt, len(receipts))
	copy(out, receipts)
	return out
}
