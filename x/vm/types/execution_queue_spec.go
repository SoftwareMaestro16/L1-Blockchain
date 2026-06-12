package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMQueueStatePrefixPriority	= "queue/priority"
	AVMQueueStatePrefixDelayed	= "queue/delayed"
	AVMQueueStatePrefixRetry	= "queue/retry"
	AVMQueueStatePrefixFailed	= "queue/failed"

	AVMQueueLanePriority	AVMQueueLane	= "priority"
	AVMQueueLaneDelayed	AVMQueueLane	= "delayed"
	AVMQueueLaneRetry	AVMQueueLane	= "retry"
	AVMQueueLaneFailed	AVMQueueLane	= "failed"
)

type AVMQueueLane string

type AVMZoneQueueEntry struct {
	ZoneID		zonestypes.ZoneID
	Lane		AVMQueueLane
	MessageID	string
	SortKey		string
	Priority	uint8
	ScheduledHeight	uint64
	SenderHash	string
	Nonce		uint64
	GasLimit	uint64
}

type AVMZoneQueue struct {
	ZoneID		zonestypes.ZoneID
	PriorityQueue	[]AVMZoneQueueEntry
	DelayedQueue	[]AVMZoneQueueEntry
	RetryQueue	[]AVMZoneQueueEntry
	FailedQueue	[]AVMZoneQueueEntry
	QueueRoot	string
}

type AVMZoneQueueSelection struct {
	Ready		[]AVMAsyncMessage
	Expired		[]AVMAsyncMessage
	Remaining	AVMZoneQueue
	Budget		zonestypes.ZoneExecutionBudget
}

type AVMZoneQueueProof struct {
	ZoneID		zonestypes.ZoneID
	Lane		AVMQueueLane
	MessageID	string
	SortKey		string
	StateKey	string
	QueueRoot	string
}

func NewAVMZoneQueue(queue AVMZoneQueue) (AVMZoneQueue, error) {
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	return queue, queue.Validate()
}

func NewAVMZoneQueueEntry(lane AVMQueueLane, msg AVMAsyncMessage, scheduledHeight uint64) (AVMZoneQueueEntry, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if scheduledHeight == 0 {
		scheduledHeight = AVMMessageScheduledHeight(msg)
	}
	entry := AVMZoneQueueEntry{
		ZoneID:			msg.DestinationZone,
		Lane:			lane,
		MessageID:		msg.ID,
		Priority:		msg.Priority,
		ScheduledHeight:	scheduledHeight,
		SenderHash:		AVMQueueSenderHash(msg.SourceZone, msg.Source),
		Nonce:			msg.SenderNonce,
		GasLimit:		msg.GasLimit,
	}
	entry.SortKey = AVMQueueSortKey(entry.Priority, entry.ScheduledHeight, entry.SenderHash, entry.Nonce, entry.MessageID)
	return entry, entry.Validate()
}

func (q AVMZoneQueue) Validate() error {
	q = canonicalAVMZoneQueue(q)
	if err := zonestypes.ValidateZoneID(q.ZoneID); err != nil {
		return err
	}
	if err := validateAVMQueueLane(q.ZoneID, AVMQueueLanePriority, q.PriorityQueue); err != nil {
		return err
	}
	if err := validateAVMQueueLane(q.ZoneID, AVMQueueLaneDelayed, q.DelayedQueue); err != nil {
		return err
	}
	if err := validateAVMQueueLane(q.ZoneID, AVMQueueLaneRetry, q.RetryQueue); err != nil {
		return err
	}
	if err := validateAVMQueueLane(q.ZoneID, AVMQueueLaneFailed, q.FailedQueue); err != nil {
		return err
	}
	if err := validateAVMQueueUniqueMessages(q); err != nil {
		return err
	}
	if q.QueueRoot == "" {
		return errors.New("AVM zone queue root is required")
	}
	if err := zonestypes.ValidateHash("AVM zone queue root", q.QueueRoot); err != nil {
		return err
	}
	if q.QueueRoot != ComputeAVMZoneQueueRoot(q) {
		return errors.New("AVM zone queue root mismatch")
	}
	return nil
}

func (e AVMZoneQueueEntry) Validate() error {
	e = canonicalAVMZoneQueueEntry(e)
	if err := zonestypes.ValidateZoneID(e.ZoneID); err != nil {
		return err
	}
	if !IsAVMQueueLane(e.Lane) {
		return fmt.Errorf("invalid AVM queue lane %q", e.Lane)
	}
	if err := zonestypes.ValidateHash("AVM queue message id", e.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM queue sender hash", e.SenderHash); err != nil {
		return err
	}
	if e.ScheduledHeight == 0 {
		return errors.New("AVM queue scheduled height must be positive")
	}
	if e.GasLimit == 0 {
		return errors.New("AVM queue gas limit must be positive")
	}
	if e.SortKey != AVMQueueSortKey(e.Priority, e.ScheduledHeight, e.SenderHash, e.Nonce, e.MessageID) {
		return errors.New("AVM queue sort key mismatch")
	}
	return validateAVMQueueStateKey("AVM queue state key", e.StateKey())
}

func IsAVMQueueLane(lane AVMQueueLane) bool {
	switch lane {
	case AVMQueueLanePriority, AVMQueueLaneDelayed, AVMQueueLaneRetry, AVMQueueLaneFailed:
		return true
	default:
		return false
	}
}

func (e AVMZoneQueueEntry) StateKey() string {
	switch e.Lane {
	case AVMQueueLanePriority:
		return AVMQueuePriorityKey(e.ZoneID, e.SortKey)
	case AVMQueueLaneDelayed:
		return AVMQueueDelayedKey(e.ZoneID, e.ScheduledHeight, e.SortKey)
	case AVMQueueLaneRetry:
		return AVMQueueRetryKey(e.ZoneID, e.ScheduledHeight, e.SortKey)
	case AVMQueueLaneFailed:
		return AVMQueueFailedKey(e.ZoneID, e.SortKey)
	default:
		return ""
	}
}

func AVMQueuePriorityKey(zoneID zonestypes.ZoneID, sortKey string) string {
	return AVMQueueStatePrefixPriority + "/" + string(zoneID) + "/" + strings.TrimSpace(sortKey)
}

func AVMQueueDelayedKey(zoneID zonestypes.ZoneID, resumeHeight uint64, sortKey string) string {
	return fmt.Sprintf("%s/%s/%020d/%s", AVMQueueStatePrefixDelayed, zoneID, resumeHeight, strings.TrimSpace(sortKey))
}

func AVMQueueRetryKey(zoneID zonestypes.ZoneID, retryHeight uint64, sortKey string) string {
	return fmt.Sprintf("%s/%s/%020d/%s", AVMQueueStatePrefixRetry, zoneID, retryHeight, strings.TrimSpace(sortKey))
}

func AVMQueueFailedKey(zoneID zonestypes.ZoneID, sortKey string) string {
	return AVMQueueStatePrefixFailed + "/" + string(zoneID) + "/" + strings.TrimSpace(sortKey)
}

func AVMQueueSortKey(priority uint8, scheduledHeight uint64, senderHash string, nonce uint64, messageID string) string {
	return fmt.Sprintf("%03d/%020d/%s/%020d/%s", MaxAsyncMessagePriority-priority, scheduledHeight, strings.TrimSpace(senderHash), nonce, strings.TrimSpace(messageID))
}

func AVMQueueSenderHash(sourceZone zonestypes.ZoneID, source string) string {
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-queue-sender-v1")
	writeEnginePart(h, string(sourceZone))
	writeEnginePart(h, strings.TrimSpace(source))
	return hex.EncodeToString(h.Sum(nil))
}

func AVMMessageScheduledHeight(msg AVMAsyncMessage) uint64 {
	msg = canonicalAVMAsyncMessage(msg)
	if msg.DelayHeight > ^uint64(0)-msg.CreatedHeight {
		return ^uint64(0)
	}
	return msg.CreatedHeight + msg.DelayHeight
}

func SelectAVMZoneQueueWork(queue AVMZoneQueue, messages []AVMAsyncMessage, height uint64, budget zonestypes.ZoneExecutionBudget) (AVMZoneQueueSelection, error) {
	if height == 0 {
		return AVMZoneQueueSelection{}, errors.New("AVM queue execution height must be positive")
	}
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	if err := queue.Validate(); err != nil {
		return AVMZoneQueueSelection{}, err
	}
	if err := budget.Validate(); err != nil {
		return AVMZoneQueueSelection{}, err
	}
	messageByID, err := indexAVMQueueMessages(queue.ZoneID, messages)
	if err != nil {
		return AVMZoneQueueSelection{}, err
	}

	candidates := queueExecutableEntries(queue)
	readyIDs := make(map[string]struct{})
	expiredIDs := make(map[string]struct{})
	nextBudget := budget
	for _, entry := range candidates {
		msg := messageByID[entry.MessageID]
		if height > msg.ExpiryHeight {
			consumed, err := nextBudget.Consume(0, 1)
			if err != nil {
				break
			}
			nextBudget = consumed
			expiredIDs[msg.ID] = struct{}{}
			continue
		}
		if height < entry.ScheduledHeight || height < AVMMessageScheduledHeight(msg) {
			continue
		}
		consumed, err := nextBudget.Consume(msg.GasLimit, 1)
		if err != nil {
			break
		}
		nextBudget = consumed
		readyIDs[msg.ID] = struct{}{}
	}

	selection := AVMZoneQueueSelection{Budget: nextBudget}
	for _, entry := range candidates {
		msg := messageByID[entry.MessageID]
		if _, found := expiredIDs[msg.ID]; found {
			selection.Expired = append(selection.Expired, msg)
			continue
		}
		if _, found := readyIDs[msg.ID]; found {
			selection.Ready = append(selection.Ready, msg)
		}
	}
	selection.Remaining = removeAVMZoneQueueMessages(queue, readyIDs, expiredIDs)
	return selection, nil
}

func AdmitAVMZoneQueueMessage(queue AVMZoneQueue, msg AVMAsyncMessage, height uint64, maxDepth uint32) (AVMZoneQueue, AVMZoneQueueEntry, error) {
	if height == 0 {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM queue admission height must be positive")
	}
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	if err := queue.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	if msg.DestinationZone != queue.ZoneID {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM queue admission zone mismatch")
	}
	if height > msg.ExpiryHeight {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM queue admission message is expired")
	}
	if maxDepth > 0 && uint32(len(allAVMQueueEntries(queue))) >= maxDepth {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM queue admission exceeds max queue depth")
	}
	if _, found := findAVMQueueEntry(queue, msg.ID); found {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, fmt.Errorf("duplicate AVM queued message %q", msg.ID)
	}
	lane := AVMQueueLanePriority
	if height < AVMMessageScheduledHeight(msg) {
		lane = AVMQueueLaneDelayed
	}
	entry, err := NewAVMZoneQueueEntry(lane, msg, 0)
	if err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	next := appendAVMQueueEntry(queue, entry)
	next, err = NewAVMZoneQueue(next)
	return next, entry, err
}

func AdmitAVMZoneRetryMessage(queue AVMZoneQueue, msg AVMAsyncMessage, retryHeight uint64, maxDepth uint32) (AVMZoneQueue, AVMZoneQueueEntry, error) {
	if retryHeight == 0 {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM retry queue height must be positive")
	}
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	if err := queue.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	if msg.DestinationZone != queue.ZoneID {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM retry queue zone mismatch")
	}
	if retryHeight > msg.ExpiryHeight {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM retry queue cannot exceed message expiry")
	}
	if maxDepth > 0 && uint32(len(allAVMQueueEntries(queue))) >= maxDepth {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM retry queue exceeds max queue depth")
	}
	if _, found := findAVMQueueEntry(queue, msg.ID); found {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, fmt.Errorf("duplicate AVM queued message %q", msg.ID)
	}
	entry, err := NewAVMZoneQueueEntry(AVMQueueLaneRetry, msg, retryHeight)
	if err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	next := appendAVMQueueEntry(queue, entry)
	next, err = NewAVMZoneQueue(next)
	return next, entry, err
}

func PromoteAVMZoneQueue(queue AVMZoneQueue, height uint64, maxMessages uint32) (AVMZoneQueue, []AVMZoneQueueEntry, error) {
	if height == 0 {
		return AVMZoneQueue{}, nil, errors.New("AVM queue promotion height must be positive")
	}
	if maxMessages == 0 {
		return AVMZoneQueue{}, nil, errors.New("AVM queue promotion max messages must be positive")
	}
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	if err := queue.Validate(); err != nil {
		return AVMZoneQueue{}, nil, err
	}

	due := append([]AVMZoneQueueEntry(nil), queue.DelayedQueue...)
	due = append(due, queue.RetryQueue...)
	sort.SliceStable(due, func(i, j int) bool {
		return compareAVMQueueEntries(due[i], due[j]) < 0
	})
	promoteIDs := make(map[string]struct{})
	promoted := make([]AVMZoneQueueEntry, 0, maxMessages)
	for _, entry := range due {
		if entry.ScheduledHeight > height || uint32(len(promoted)) >= maxMessages {
			continue
		}
		promotedEntry := entry
		promotedEntry.Lane = AVMQueueLanePriority
		promoted = append(promoted, promotedEntry)
		promoteIDs[entry.MessageID] = struct{}{}
	}

	keep := func(entries []AVMZoneQueueEntry) []AVMZoneQueueEntry {
		out := make([]AVMZoneQueueEntry, 0, len(entries))
		for _, entry := range entries {
			if _, found := promoteIDs[entry.MessageID]; found {
				continue
			}
			out = append(out, entry)
		}
		return out
	}
	queue.DelayedQueue = keep(queue.DelayedQueue)
	queue.RetryQueue = keep(queue.RetryQueue)
	queue.PriorityQueue = append(queue.PriorityQueue, promoted...)
	next, err := NewAVMZoneQueue(queue)
	return next, promoted, err
}

func DeadLetterAVMZoneQueueMessage(queue AVMZoneQueue, msg AVMAsyncMessage, receipt AVMExecutionReceipt, reason string, failedAttempts uint32, refundAmountOptional uint64) (AVMZoneQueue, AVMDeadLetterRecord, error) {
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	if err := queue.Validate(); err != nil {
		return AVMZoneQueue{}, AVMDeadLetterRecord{}, err
	}
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMZoneQueue{}, AVMDeadLetterRecord{}, err
	}
	if msg.DestinationZone != queue.ZoneID {
		return AVMZoneQueue{}, AVMDeadLetterRecord{}, errors.New("AVM dead letter queue zone mismatch")
	}
	record, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:		msg.ID,
		ZoneID:			queue.ZoneID,
		Reason:			reason,
		FailedAttempts:		failedAttempts,
		LastErrorCode:		receipt.ErrorCodeOptional,
		FinalHeight:		receipt.CreatedHeight,
		RefundAmountOptional:	refundAmountOptional,
		ReceiptID:		receipt.ReceiptID,
	})
	if err != nil {
		return AVMZoneQueue{}, AVMDeadLetterRecord{}, err
	}
	if err := record.ValidateWithReceipt(receipt); err != nil {
		return AVMZoneQueue{}, AVMDeadLetterRecord{}, err
	}
	failedEntry, err := NewAVMZoneQueueEntry(AVMQueueLaneFailed, msg, receipt.CreatedHeight)
	if err != nil {
		return AVMZoneQueue{}, AVMDeadLetterRecord{}, err
	}
	removeIDs := map[string]struct{}{msg.ID: {}}
	queue = removeAVMZoneQueueMessages(queue, removeIDs, nil)
	queue.FailedQueue = append(queue.FailedQueue, failedEntry)
	next, err := NewAVMZoneQueue(queue)
	return next, record, err
}

func QueryAVMZoneQueueProof(queue AVMZoneQueue, lane AVMQueueLane, messageID string) (AVMZoneQueueProof, error) {
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	if err := queue.Validate(); err != nil {
		return AVMZoneQueueProof{}, err
	}
	messageID = strings.TrimSpace(messageID)
	entries, err := queueEntriesForLane(queue, lane)
	if err != nil {
		return AVMZoneQueueProof{}, err
	}
	for _, entry := range entries {
		if entry.MessageID != messageID {
			continue
		}
		proof := AVMZoneQueueProof{
			ZoneID:		queue.ZoneID,
			Lane:		lane,
			MessageID:	entry.MessageID,
			SortKey:	entry.SortKey,
			StateKey:	entry.StateKey(),
			QueueRoot:	queue.QueueRoot,
		}
		return proof, proof.Validate()
	}
	return AVMZoneQueueProof{}, fmt.Errorf("AVM queue message %q not found in %s lane", messageID, lane)
}

func (p AVMZoneQueueProof) Validate() error {
	p.MessageID = strings.TrimSpace(p.MessageID)
	p.SortKey = strings.TrimSpace(p.SortKey)
	p.StateKey = strings.TrimSpace(p.StateKey)
	p.QueueRoot = strings.TrimSpace(p.QueueRoot)
	if err := zonestypes.ValidateZoneID(p.ZoneID); err != nil {
		return err
	}
	if !IsAVMQueueLane(p.Lane) {
		return fmt.Errorf("invalid AVM queue proof lane %q", p.Lane)
	}
	if err := zonestypes.ValidateHash("AVM queue proof message id", p.MessageID); err != nil {
		return err
	}
	if p.SortKey == "" || !strings.Contains(p.SortKey, p.MessageID) {
		return errors.New("AVM queue proof sort key must reference message id")
	}
	if err := validateAVMQueueStateKey("AVM queue proof state key", p.StateKey); err != nil {
		return err
	}
	if !strings.HasSuffix(p.StateKey, p.SortKey) {
		return errors.New("AVM queue proof state key must end with sort key")
	}
	return zonestypes.ValidateHash("AVM queue proof root", p.QueueRoot)
}

func ComputeAVMZoneQueueRoot(queue AVMZoneQueue) string {
	queue = canonicalAVMZoneQueue(queue)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-queue-v1")
	writeEnginePart(h, string(queue.ZoneID))
	writeAVMQueueEntries(h, queue.PriorityQueue)
	writeAVMQueueEntries(h, queue.DelayedQueue)
	writeAVMQueueEntries(h, queue.RetryQueue)
	writeAVMQueueEntries(h, queue.FailedQueue)
	return hex.EncodeToString(h.Sum(nil))
}

func validateAVMQueueLane(zoneID zonestypes.ZoneID, lane AVMQueueLane, entries []AVMZoneQueueEntry) error {
	for i, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if entry.ZoneID != zoneID {
			return errors.New("AVM queue entry zone mismatch")
		}
		if entry.Lane != lane {
			return errors.New("AVM queue entry lane mismatch")
		}
		if i > 0 && compareAVMQueueEntries(entries[i-1], entry) >= 0 {
			return errors.New("AVM queue entries must be sorted canonically")
		}
	}
	return nil
}

func validateAVMQueueUniqueMessages(queue AVMZoneQueue) error {
	seen := map[string]struct{}{}
	for _, entry := range allAVMQueueEntries(queue) {
		if _, found := seen[entry.MessageID]; found {
			return fmt.Errorf("duplicate AVM queued message %q", entry.MessageID)
		}
		seen[entry.MessageID] = struct{}{}
	}
	return nil
}

func indexAVMQueueMessages(zoneID zonestypes.ZoneID, messages []AVMAsyncMessage) (map[string]AVMAsyncMessage, error) {
	out := make(map[string]AVMAsyncMessage, len(messages))
	for _, msg := range messages {
		msg = canonicalAVMAsyncMessage(msg)
		if err := msg.Validate(); err != nil {
			return nil, err
		}
		if msg.DestinationZone != zoneID {
			return nil, errors.New("AVM queue message destination zone mismatch")
		}
		if _, found := out[msg.ID]; found {
			return nil, fmt.Errorf("duplicate AVM queue message %q", msg.ID)
		}
		out[msg.ID] = msg
	}
	return out, nil
}

func queueExecutableEntries(queue AVMZoneQueue) []AVMZoneQueueEntry {
	entries := append([]AVMZoneQueueEntry(nil), queue.PriorityQueue...)
	entries = append(entries, queue.DelayedQueue...)
	entries = append(entries, queue.RetryQueue...)
	sort.SliceStable(entries, func(i, j int) bool {
		return compareAVMQueueEntries(entries[i], entries[j]) < 0
	})
	return entries
}

func removeAVMZoneQueueMessages(queue AVMZoneQueue, readyIDs, expiredIDs map[string]struct{}) AVMZoneQueue {
	remove := func(entries []AVMZoneQueueEntry) []AVMZoneQueueEntry {
		out := make([]AVMZoneQueueEntry, 0, len(entries))
		for _, entry := range entries {
			if readyIDs != nil {
				if _, found := readyIDs[entry.MessageID]; found {
					continue
				}
			}
			if expiredIDs != nil {
				if _, found := expiredIDs[entry.MessageID]; found {
					continue
				}
			}
			out = append(out, entry)
		}
		return out
	}
	queue.PriorityQueue = remove(queue.PriorityQueue)
	queue.DelayedQueue = remove(queue.DelayedQueue)
	queue.RetryQueue = remove(queue.RetryQueue)
	queue.FailedQueue = remove(queue.FailedQueue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	return queue
}

func appendAVMQueueEntry(queue AVMZoneQueue, entry AVMZoneQueueEntry) AVMZoneQueue {
	switch entry.Lane {
	case AVMQueueLanePriority:
		queue.PriorityQueue = append(queue.PriorityQueue, entry)
	case AVMQueueLaneDelayed:
		queue.DelayedQueue = append(queue.DelayedQueue, entry)
	case AVMQueueLaneRetry:
		queue.RetryQueue = append(queue.RetryQueue, entry)
	case AVMQueueLaneFailed:
		queue.FailedQueue = append(queue.FailedQueue, entry)
	}
	return queue
}

func findAVMQueueEntry(queue AVMZoneQueue, messageID string) (AVMZoneQueueEntry, bool) {
	messageID = strings.TrimSpace(messageID)
	for _, entry := range allAVMQueueEntries(queue) {
		if entry.MessageID == messageID {
			return entry, true
		}
	}
	return AVMZoneQueueEntry{}, false
}

func queueEntriesForLane(queue AVMZoneQueue, lane AVMQueueLane) ([]AVMZoneQueueEntry, error) {
	switch lane {
	case AVMQueueLanePriority:
		return queue.PriorityQueue, nil
	case AVMQueueLaneDelayed:
		return queue.DelayedQueue, nil
	case AVMQueueLaneRetry:
		return queue.RetryQueue, nil
	case AVMQueueLaneFailed:
		return queue.FailedQueue, nil
	default:
		return nil, fmt.Errorf("invalid AVM queue lane %q", lane)
	}
}

func canonicalAVMZoneQueue(queue AVMZoneQueue) AVMZoneQueue {
	queue.PriorityQueue = cloneAVMQueueEntries(queue.PriorityQueue)
	queue.DelayedQueue = cloneAVMQueueEntries(queue.DelayedQueue)
	queue.RetryQueue = cloneAVMQueueEntries(queue.RetryQueue)
	queue.FailedQueue = cloneAVMQueueEntries(queue.FailedQueue)
	queue.QueueRoot = strings.TrimSpace(queue.QueueRoot)
	sort.SliceStable(queue.PriorityQueue, func(i, j int) bool { return compareAVMQueueEntries(queue.PriorityQueue[i], queue.PriorityQueue[j]) < 0 })
	sort.SliceStable(queue.DelayedQueue, func(i, j int) bool { return compareAVMQueueEntries(queue.DelayedQueue[i], queue.DelayedQueue[j]) < 0 })
	sort.SliceStable(queue.RetryQueue, func(i, j int) bool { return compareAVMQueueEntries(queue.RetryQueue[i], queue.RetryQueue[j]) < 0 })
	sort.SliceStable(queue.FailedQueue, func(i, j int) bool { return compareAVMQueueEntries(queue.FailedQueue[i], queue.FailedQueue[j]) < 0 })
	return queue
}

func cloneAVMQueueEntries(entries []AVMZoneQueueEntry) []AVMZoneQueueEntry {
	out := append([]AVMZoneQueueEntry(nil), entries...)
	for i := range out {
		out[i] = canonicalAVMZoneQueueEntry(out[i])
	}
	return out
}

func canonicalAVMZoneQueueEntry(entry AVMZoneQueueEntry) AVMZoneQueueEntry {
	entry.MessageID = strings.TrimSpace(entry.MessageID)
	entry.SortKey = strings.TrimSpace(entry.SortKey)
	entry.SenderHash = strings.TrimSpace(entry.SenderHash)
	return entry
}

func compareAVMQueueEntries(left, right AVMZoneQueueEntry) int {
	if left.SortKey < right.SortKey {
		return -1
	}
	if left.SortKey > right.SortKey {
		return 1
	}
	if left.MessageID < right.MessageID {
		return -1
	}
	if left.MessageID > right.MessageID {
		return 1
	}
	return 0
}

func allAVMQueueEntries(queue AVMZoneQueue) []AVMZoneQueueEntry {
	entries := append([]AVMZoneQueueEntry(nil), queue.PriorityQueue...)
	entries = append(entries, queue.DelayedQueue...)
	entries = append(entries, queue.RetryQueue...)
	entries = append(entries, queue.FailedQueue...)
	return entries
}

func writeAVMQueueEntries(h engineByteWriter, entries []AVMZoneQueueEntry) {
	writeEngineUint64(h, uint64(len(entries)))
	for _, entry := range entries {
		writeEnginePart(h, string(entry.ZoneID))
		writeEnginePart(h, string(entry.Lane))
		writeEnginePart(h, entry.MessageID)
		writeEnginePart(h, entry.SortKey)
		writeEngineUint64(h, uint64(entry.Priority))
		writeEngineUint64(h, entry.ScheduledHeight)
		writeEnginePart(h, entry.SenderHash)
		writeEngineUint64(h, entry.Nonce)
		writeEngineUint64(h, entry.GasLimit)
		writeEnginePart(h, entry.StateKey())
	}
}

func validateAVMQueueStateKey(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if !strings.HasPrefix(value, "queue/") {
		return fmt.Errorf("%s must start with queue/", fieldName)
	}
	if strings.Contains(value, "//") {
		return fmt.Errorf("%s must not contain empty path segments", fieldName)
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" {
			return fmt.Errorf("%s must not contain empty path segments", fieldName)
		}
		if err := validateEngineToken(fieldName+" segment", part, MaxAVMStateKeySegmentLength); err != nil {
			return err
		}
	}
	return nil
}
