package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	MaxAVMInvariantIDLength		= 128
	MaxAVMWorkflowIDLength		= 128
	MaxAVMStorageKeyLength		= 256
	MaxAVMRootRecordsPerPlan	= 4096
)

type AVMRoot struct {
	Height			uint64
	RouterRoot		string
	AsyncMessageRoot	string
	ActorRoot		string
	ContractRoot		string
	ContinuationRoot	string
	InterfaceRoot		string
	ReceiptRoot		string
	RootHash		string
}

type AVMZoneStateRoot struct {
	ZoneID			zonestypes.ZoneID
	Height			uint64
	StateRoot		string
	MessageRoot		string
	ExecutionRoot		string
	ContinuationRoot	string
	RootHash		string
}

type AVMStoredMessageRecord struct {
	MessageID		string
	ZoneID			zonestypes.ZoneID
	Expired			bool
	Bounced			bool
	OriginalMessageID	string
}

type AVMQueuedMessageRef struct {
	MessageID	string
	ZoneID		zonestypes.ZoneID
	QueueID		string
	SortKey		string
}

type AVMExecutedMessageRef struct {
	MessageID string
}

type AVMReceiptRecord struct {
	ReceiptID	string
	MessageID	string
	ResultCode	uint32
}

type AVMContinuationRef struct {
	ContinuationID	string
	ActorID		string
	WorkflowID	string
}

type AVMActorRef struct {
	ActorID string
}

type AVMWorkflowRef struct {
	WorkflowID string
}

type AVMActorMailboxEntry struct {
	ActorID		string
	SortKey		string
	MessageID	string
}

type AVMContractStorageRef struct {
	ContractAddress	string
	Key		string
}

type AVMStateInvariantSet struct {
	StoredMessages		[]AVMStoredMessageRecord
	QueuedMessages		[]AVMQueuedMessageRef
	ExecutedMessages	[]AVMExecutedMessageRef
	Receipts		[]AVMReceiptRecord
	Continuations		[]AVMContinuationRef
	Actors			[]AVMActorRef
	Workflows		[]AVMWorkflowRef
	MailboxEntries		[]AVMActorMailboxEntry
	ContractStorage		[]AVMContractStorageRef
	ZoneRoots		[]AVMZoneStateRoot
}

func NewAVMRoot(root AVMRoot) (AVMRoot, error) {
	root = canonicalAVMRoot(root)
	root.RootHash = ComputeAVMRootHash(root)
	return root, root.Validate()
}

func NewAVMZoneStateRoot(root AVMZoneStateRoot) (AVMZoneStateRoot, error) {
	root = canonicalAVMZoneStateRoot(root)
	root.RootHash = ComputeAVMZoneStateRootHash(root)
	return root, root.Validate()
}

func (r AVMRoot) Validate() error {
	r = canonicalAVMRoot(r)
	if r.Height == 0 {
		return errors.New("AVM root height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM router root", value: r.RouterRoot},
		{name: "AVM async message root", value: r.AsyncMessageRoot},
		{name: "AVM actor root", value: r.ActorRoot},
		{name: "AVM contract root", value: r.ContractRoot},
		{name: "AVM continuation root", value: r.ContinuationRoot},
		{name: "AVM interface root", value: r.InterfaceRoot},
		{name: "AVM receipt root", value: r.ReceiptRoot},
		{name: "AVM root hash", value: r.RootHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.RootHash != ComputeAVMRootHash(r) {
		return errors.New("AVM root hash mismatch")
	}
	return nil
}

func (r AVMZoneStateRoot) Validate() error {
	r = canonicalAVMZoneStateRoot(r)
	if err := zonestypes.ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("AVM zone root height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM zone state root", value: r.StateRoot},
		{name: "AVM zone message root", value: r.MessageRoot},
		{name: "AVM zone execution root", value: r.ExecutionRoot},
		{name: "AVM zone continuation root", value: r.ContinuationRoot},
		{name: "AVM zone root hash", value: r.RootHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.RootHash != ComputeAVMZoneStateRootHash(r) {
		return errors.New("AVM zone root hash mismatch")
	}
	return nil
}

func (s AVMStateInvariantSet) Validate() error {
	s = canonicalAVMStateInvariantSet(s)
	if err := validateAVMStoredMessages(s.StoredMessages); err != nil {
		return err
	}
	if err := validateAVMQueuedMessages(s.QueuedMessages, s.StoredMessages); err != nil {
		return err
	}
	if err := validateAVMExecutedReceipts(s.ExecutedMessages, s.Receipts); err != nil {
		return err
	}
	if err := validateAVMExpiredReceipts(s.StoredMessages, s.Receipts); err != nil {
		return err
	}
	if err := validateAVMBouncedMessages(s.StoredMessages); err != nil {
		return err
	}
	if err := validateAVMContinuations(s.Continuations, s.Actors, s.Workflows); err != nil {
		return err
	}
	if err := validateAVMActorMailboxes(s.MailboxEntries, s.StoredMessages); err != nil {
		return err
	}
	if err := validateAVMContractStorage(s.ContractStorage); err != nil {
		return err
	}
	return validateAVMZoneRoots(s.ZoneRoots)
}

func ComputeAVMRootHash(root AVMRoot) string {
	root = canonicalAVMRoot(root)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-root-v1")
	writeEngineUint64(h, root.Height)
	writeEnginePart(h, root.RouterRoot)
	writeEnginePart(h, root.AsyncMessageRoot)
	writeEnginePart(h, root.ActorRoot)
	writeEnginePart(h, root.ContractRoot)
	writeEnginePart(h, root.ContinuationRoot)
	writeEnginePart(h, root.InterfaceRoot)
	writeEnginePart(h, root.ReceiptRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneStateRootHash(root AVMZoneStateRoot) string {
	root = canonicalAVMZoneStateRoot(root)
	h := sha256.New()
	writeEnginePart(h, "aetra-zone-root-v2")
	writeEnginePart(h, string(root.ZoneID))
	writeEngineUint64(h, root.Height)
	writeEnginePart(h, root.StateRoot)
	writeEnginePart(h, root.MessageRoot)
	writeEnginePart(h, root.ExecutionRoot)
	writeEnginePart(h, root.ContinuationRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStateInvariantRoot(set AVMStateInvariantSet) string {
	set = canonicalAVMStateInvariantSet(set)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-state-invariants-v1")
	writeEngineUint64(h, uint64(len(set.StoredMessages)))
	for _, msg := range set.StoredMessages {
		writeEnginePart(h, msg.MessageID)
		writeEnginePart(h, string(msg.ZoneID))
		writeEngineBool(h, msg.Expired)
		writeEngineBool(h, msg.Bounced)
		writeEnginePart(h, msg.OriginalMessageID)
	}
	writeEngineUint64(h, uint64(len(set.QueuedMessages)))
	for _, msg := range set.QueuedMessages {
		writeEnginePart(h, msg.MessageID)
		writeEnginePart(h, string(msg.ZoneID))
		writeEnginePart(h, msg.QueueID)
		writeEnginePart(h, msg.SortKey)
	}
	writeEngineUint64(h, uint64(len(set.ExecutedMessages)))
	for _, msg := range set.ExecutedMessages {
		writeEnginePart(h, msg.MessageID)
	}
	writeEngineUint64(h, uint64(len(set.Receipts)))
	for _, receipt := range set.Receipts {
		writeEnginePart(h, receipt.ReceiptID)
		writeEnginePart(h, receipt.MessageID)
		writeEngineUint64(h, uint64(receipt.ResultCode))
	}
	writeEngineUint64(h, uint64(len(set.Continuations)))
	for _, continuation := range set.Continuations {
		writeEnginePart(h, continuation.ContinuationID)
		writeEnginePart(h, continuation.ActorID)
		writeEnginePart(h, continuation.WorkflowID)
	}
	writeEngineUint64(h, uint64(len(set.MailboxEntries)))
	for _, entry := range set.MailboxEntries {
		writeEnginePart(h, entry.ActorID)
		writeEnginePart(h, entry.SortKey)
		writeEnginePart(h, entry.MessageID)
	}
	writeEngineUint64(h, uint64(len(set.ContractStorage)))
	for _, storage := range set.ContractStorage {
		writeEnginePart(h, storage.ContractAddress)
		writeEnginePart(h, storage.Key)
	}
	writeEngineUint64(h, uint64(len(set.ZoneRoots)))
	for _, root := range set.ZoneRoots {
		writeEnginePart(h, root.RootHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMRoot(root AVMRoot) AVMRoot {
	root.RouterRoot = strings.TrimSpace(root.RouterRoot)
	root.AsyncMessageRoot = strings.TrimSpace(root.AsyncMessageRoot)
	root.ActorRoot = strings.TrimSpace(root.ActorRoot)
	root.ContractRoot = strings.TrimSpace(root.ContractRoot)
	root.ContinuationRoot = strings.TrimSpace(root.ContinuationRoot)
	root.InterfaceRoot = strings.TrimSpace(root.InterfaceRoot)
	root.ReceiptRoot = strings.TrimSpace(root.ReceiptRoot)
	root.RootHash = strings.TrimSpace(root.RootHash)
	return root
}

func canonicalAVMZoneStateRoot(root AVMZoneStateRoot) AVMZoneStateRoot {
	root.StateRoot = strings.TrimSpace(root.StateRoot)
	root.MessageRoot = strings.TrimSpace(root.MessageRoot)
	root.ExecutionRoot = strings.TrimSpace(root.ExecutionRoot)
	root.ContinuationRoot = strings.TrimSpace(root.ContinuationRoot)
	root.RootHash = strings.TrimSpace(root.RootHash)
	return root
}

func canonicalAVMStateInvariantSet(set AVMStateInvariantSet) AVMStateInvariantSet {
	set.StoredMessages = append([]AVMStoredMessageRecord(nil), set.StoredMessages...)
	set.QueuedMessages = append([]AVMQueuedMessageRef(nil), set.QueuedMessages...)
	set.ExecutedMessages = append([]AVMExecutedMessageRef(nil), set.ExecutedMessages...)
	set.Receipts = append([]AVMReceiptRecord(nil), set.Receipts...)
	set.Continuations = append([]AVMContinuationRef(nil), set.Continuations...)
	set.Actors = append([]AVMActorRef(nil), set.Actors...)
	set.Workflows = append([]AVMWorkflowRef(nil), set.Workflows...)
	set.MailboxEntries = append([]AVMActorMailboxEntry(nil), set.MailboxEntries...)
	set.ContractStorage = append([]AVMContractStorageRef(nil), set.ContractStorage...)
	set.ZoneRoots = append([]AVMZoneStateRoot(nil), set.ZoneRoots...)
	for i := range set.ZoneRoots {
		set.ZoneRoots[i] = canonicalAVMZoneStateRoot(set.ZoneRoots[i])
	}
	sort.SliceStable(set.StoredMessages, func(i, j int) bool { return set.StoredMessages[i].MessageID < set.StoredMessages[j].MessageID })
	sort.SliceStable(set.QueuedMessages, func(i, j int) bool { return compareQueuedMessageRefs(set.QueuedMessages[i], set.QueuedMessages[j]) < 0 })
	sort.SliceStable(set.ExecutedMessages, func(i, j int) bool { return set.ExecutedMessages[i].MessageID < set.ExecutedMessages[j].MessageID })
	sort.SliceStable(set.Receipts, func(i, j int) bool { return set.Receipts[i].ReceiptID < set.Receipts[j].ReceiptID })
	sort.SliceStable(set.Continuations, func(i, j int) bool { return set.Continuations[i].ContinuationID < set.Continuations[j].ContinuationID })
	sort.SliceStable(set.Actors, func(i, j int) bool { return set.Actors[i].ActorID < set.Actors[j].ActorID })
	sort.SliceStable(set.Workflows, func(i, j int) bool { return set.Workflows[i].WorkflowID < set.Workflows[j].WorkflowID })
	sort.SliceStable(set.MailboxEntries, func(i, j int) bool { return compareMailboxEntries(set.MailboxEntries[i], set.MailboxEntries[j]) < 0 })
	sort.SliceStable(set.ContractStorage, func(i, j int) bool {
		return compareContractStorageRefs(set.ContractStorage[i], set.ContractStorage[j]) < 0
	})
	sort.SliceStable(set.ZoneRoots, func(i, j int) bool {
		if set.ZoneRoots[i].ZoneID != set.ZoneRoots[j].ZoneID {
			return set.ZoneRoots[i].ZoneID < set.ZoneRoots[j].ZoneID
		}
		return set.ZoneRoots[i].Height < set.ZoneRoots[j].Height
	})
	return set
}

func validateAVMStoredMessages(messages []AVMStoredMessageRecord) error {
	if len(messages) > MaxAVMRootRecordsPerPlan {
		return fmt.Errorf("AVM stored messages must be <= %d", MaxAVMRootRecordsPerPlan)
	}
	seen := make(map[string]struct{}, len(messages))
	for i, msg := range messages {
		if err := validateInvariantID("AVM message id", msg.MessageID); err != nil {
			return err
		}
		if err := zonestypes.ValidateZoneID(msg.ZoneID); err != nil {
			return err
		}
		if _, found := seen[msg.MessageID]; found {
			return fmt.Errorf("duplicate AVM message id %q", msg.MessageID)
		}
		seen[msg.MessageID] = struct{}{}
		if i > 0 && messages[i-1].MessageID >= msg.MessageID {
			return errors.New("AVM stored messages must be sorted canonically")
		}
	}
	return nil
}

func validateAVMQueuedMessages(queued []AVMQueuedMessageRef, messages []AVMStoredMessageRecord) error {
	messageByID := indexStoredMessages(messages)
	for i, msg := range queued {
		if err := validateInvariantID("AVM queued message id", msg.MessageID); err != nil {
			return err
		}
		stored, found := messageByID[msg.MessageID]
		if !found {
			return fmt.Errorf("queued message %q has no stored message record", msg.MessageID)
		}
		if stored.ZoneID != msg.ZoneID {
			return errors.New("queued message zone mismatch")
		}
		if err := zonestypes.ValidateZoneID(msg.ZoneID); err != nil {
			return err
		}
		if err := validateInvariantID("AVM queue id", msg.QueueID); err != nil {
			return err
		}
		if err := validateInvariantID("AVM queue sort key", msg.SortKey); err != nil {
			return err
		}
		if i > 0 && compareQueuedMessageRefs(queued[i-1], msg) >= 0 {
			return errors.New("AVM queued messages must be sorted canonically")
		}
	}
	return nil
}

func validateAVMExecutedReceipts(executed []AVMExecutedMessageRef, receipts []AVMReceiptRecord) error {
	receiptsByMessage := make(map[string]int, len(receipts))
	seenReceipts := make(map[string]struct{}, len(receipts))
	for i, receipt := range receipts {
		if err := validateInvariantID("AVM receipt id", receipt.ReceiptID); err != nil {
			return err
		}
		if err := validateInvariantID("AVM receipt message id", receipt.MessageID); err != nil {
			return err
		}
		if _, found := seenReceipts[receipt.ReceiptID]; found {
			return fmt.Errorf("duplicate AVM receipt id %q", receipt.ReceiptID)
		}
		seenReceipts[receipt.ReceiptID] = struct{}{}
		receiptsByMessage[receipt.MessageID]++
		if i > 0 && receipts[i-1].ReceiptID >= receipt.ReceiptID {
			return errors.New("AVM receipts must be sorted canonically")
		}
	}
	seenExecuted := make(map[string]struct{}, len(executed))
	for i, msg := range executed {
		if err := validateInvariantID("AVM executed message id", msg.MessageID); err != nil {
			return err
		}
		if _, found := seenExecuted[msg.MessageID]; found {
			return fmt.Errorf("duplicate AVM executed message %q", msg.MessageID)
		}
		seenExecuted[msg.MessageID] = struct{}{}
		if receiptsByMessage[msg.MessageID] != 1 {
			return fmt.Errorf("executed message %q must have exactly one receipt", msg.MessageID)
		}
		if i > 0 && executed[i-1].MessageID >= msg.MessageID {
			return errors.New("AVM executed messages must be sorted canonically")
		}
	}
	return nil
}

func validateAVMExpiredReceipts(messages []AVMStoredMessageRecord, receipts []AVMReceiptRecord) error {
	receiptByMessage := make(map[string]AVMReceiptRecord, len(receipts))
	for _, receipt := range receipts {
		receiptByMessage[receipt.MessageID] = receipt
	}
	for _, msg := range messages {
		if !msg.Expired {
			continue
		}
		receipt, found := receiptByMessage[msg.MessageID]
		if !found {
			return fmt.Errorf("expired message %q has no expired receipt", msg.MessageID)
		}
		if receipt.ResultCode != async.ResultExpired {
			return fmt.Errorf("expired message %q receipt must be expired", msg.MessageID)
		}
	}
	return nil
}

func validateAVMBouncedMessages(messages []AVMStoredMessageRecord) error {
	messageByID := indexStoredMessages(messages)
	for _, msg := range messages {
		if !msg.Bounced {
			continue
		}
		if msg.OriginalMessageID == "" {
			return fmt.Errorf("bounced message %q must reference original message id", msg.MessageID)
		}
		if _, found := messageByID[msg.OriginalMessageID]; !found {
			return fmt.Errorf("bounced message %q references missing original message", msg.MessageID)
		}
	}
	return nil
}

func validateAVMContinuations(continuations []AVMContinuationRef, actors []AVMActorRef, workflows []AVMWorkflowRef) error {
	actorIDs := make(map[string]struct{}, len(actors))
	for _, actor := range actors {
		if err := validateInvariantID("AVM actor id", actor.ActorID); err != nil {
			return err
		}
		actorIDs[actor.ActorID] = struct{}{}
	}
	workflowIDs := make(map[string]struct{}, len(workflows))
	for _, workflow := range workflows {
		if err := validateInvariantID("AVM workflow id", workflow.WorkflowID); err != nil {
			return err
		}
		workflowIDs[workflow.WorkflowID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(continuations))
	for i, continuation := range continuations {
		if err := validateInvariantID("AVM continuation id", continuation.ContinuationID); err != nil {
			return err
		}
		if continuation.ActorID == "" && continuation.WorkflowID == "" {
			return errors.New("continuation must reference an actor or workflow")
		}
		if continuation.ActorID != "" {
			if _, found := actorIDs[continuation.ActorID]; !found {
				return fmt.Errorf("continuation %q references missing actor", continuation.ContinuationID)
			}
		}
		if continuation.WorkflowID != "" {
			if _, found := workflowIDs[continuation.WorkflowID]; !found {
				return fmt.Errorf("continuation %q references missing workflow", continuation.ContinuationID)
			}
		}
		if _, found := seen[continuation.ContinuationID]; found {
			return fmt.Errorf("duplicate continuation %q", continuation.ContinuationID)
		}
		seen[continuation.ContinuationID] = struct{}{}
		if i > 0 && continuations[i-1].ContinuationID >= continuation.ContinuationID {
			return errors.New("AVM continuations must be sorted canonically")
		}
	}
	return nil
}

func validateAVMActorMailboxes(entries []AVMActorMailboxEntry, messages []AVMStoredMessageRecord) error {
	messageByID := indexStoredMessages(messages)
	seen := make(map[string]struct{}, len(entries))
	for i, entry := range entries {
		if err := validateInvariantID("AVM mailbox actor id", entry.ActorID); err != nil {
			return err
		}
		if err := validateInvariantID("AVM mailbox sort key", entry.SortKey); err != nil {
			return err
		}
		if err := validateInvariantID("AVM mailbox message id", entry.MessageID); err != nil {
			return err
		}
		if _, found := messageByID[entry.MessageID]; !found {
			return fmt.Errorf("mailbox entry message %q has no stored message record", entry.MessageID)
		}
		key := entry.ActorID + "/" + entry.SortKey
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate mailbox sort key %q", key)
		}
		seen[key] = struct{}{}
		if i > 0 && compareMailboxEntries(entries[i-1], entry) >= 0 {
			return errors.New("actor mailbox entries must be ordered by deterministic sort key")
		}
	}
	return nil
}

func validateAVMContractStorage(storage []AVMContractStorageRef) error {
	for i, entry := range storage {
		if err := validateInvariantID("AVM contract address", entry.ContractAddress); err != nil {
			return err
		}
		if strings.TrimSpace(entry.Key) != entry.Key || entry.Key == "" {
			return errors.New("contract storage key is required")
		}
		if len(entry.Key) > MaxAVMStorageKeyLength {
			return fmt.Errorf("contract storage key must be <= %d bytes", MaxAVMStorageKeyLength)
		}
		expectedPrefix := AVMContractStorageKey(entry.ContractAddress, "")
		if !strings.HasPrefix(entry.Key, expectedPrefix) {
			return fmt.Errorf("contract storage key must be scoped by contract address prefix %q", expectedPrefix)
		}
		if i > 0 && compareContractStorageRefs(storage[i-1], entry) >= 0 {
			return errors.New("contract storage refs must be sorted canonically")
		}
	}
	return nil
}

func validateAVMZoneRoots(roots []AVMZoneStateRoot) error {
	seen := make(map[string]struct{}, len(roots))
	for i, root := range roots {
		if err := root.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%020d", root.ZoneID, root.Height)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate AVM zone root %s", key)
		}
		seen[key] = struct{}{}
		if root.StateRoot == "" || root.MessageRoot == "" || root.ExecutionRoot == "" || root.ContinuationRoot == "" {
			return errors.New("zone root must include state, message, execution, and continuation roots")
		}
		if i > 0 {
			prev := roots[i-1]
			if prev.ZoneID > root.ZoneID || prev.ZoneID == root.ZoneID && prev.Height >= root.Height {
				return errors.New("AVM zone roots must be sorted canonically")
			}
		}
	}
	return nil
}

func indexStoredMessages(messages []AVMStoredMessageRecord) map[string]AVMStoredMessageRecord {
	out := make(map[string]AVMStoredMessageRecord, len(messages))
	for _, msg := range messages {
		out[msg.MessageID] = msg
	}
	return out
}

func validateInvariantID(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > MaxAVMInvariantIDLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxAVMInvariantIDLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func compareQueuedMessageRefs(left, right AVMQueuedMessageRef) int {
	if left.ZoneID != right.ZoneID {
		if left.ZoneID < right.ZoneID {
			return -1
		}
		return 1
	}
	if left.QueueID < right.QueueID {
		return -1
	}
	if left.QueueID > right.QueueID {
		return 1
	}
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

func compareMailboxEntries(left, right AVMActorMailboxEntry) int {
	if left.ActorID < right.ActorID {
		return -1
	}
	if left.ActorID > right.ActorID {
		return 1
	}
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

func compareContractStorageRefs(left, right AVMContractStorageRef) int {
	if left.ContractAddress < right.ContractAddress {
		return -1
	}
	if left.ContractAddress > right.ContractAddress {
		return 1
	}
	if left.Key < right.Key {
		return -1
	}
	if left.Key > right.Key {
		return 1
	}
	return 0
}
