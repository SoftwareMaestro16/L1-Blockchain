package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ExecutionRoutingClass string

const (
	ExecutionRoutingTx			ExecutionRoutingClass	= "tx"
	ExecutionRoutingZone			ExecutionRoutingClass	= "zone"
	ExecutionRoutingShard			ExecutionRoutingClass	= "shard"
	ExecutionRoutingExecutionOverlay	ExecutionRoutingClass	= "execution_overlay"
)

type ReceiptPolicy string

const (
	ReceiptPolicyNone		ReceiptPolicy	= "none"
	ReceiptPolicyOnDelivery		ReceiptPolicy	= "on_delivery"
	ReceiptPolicyOnExecution	ReceiptPolicy	= "on_execution"
	ReceiptPolicyOnFailure		ReceiptPolicy	= "on_failure"
	ReceiptPolicyAlways		ReceiptPolicy	= "always"
)

type CrossZoneReceiptStatus string

const (
	CrossZoneReceiptDelivered	CrossZoneReceiptStatus	= "delivered"
	CrossZoneReceiptExecuted	CrossZoneReceiptStatus	= "executed"
	CrossZoneReceiptExpired		CrossZoneReceiptStatus	= "expired"
	CrossZoneReceiptBounced		CrossZoneReceiptStatus	= "bounced"
	CrossZoneReceiptRolledBack	CrossZoneReceiptStatus	= "rolled_back"
)

type ExecutionMessageSchedule struct {
	ScheduleID		string
	ZoneID			string
	ShardID			string
	RoutingClass		ExecutionRoutingClass
	Committed		bool
	Ordered			bool
	ScheduleHash		string
	TransactionIDs		[]string
	MessageIDs		[]string
	FirstZoneSequence	uint64
	LastZoneSequence	uint64
}

type ExecutionZoneMessage struct {
	Message			AetherMeshMessage
	RoutingClass		ExecutionRoutingClass
	ZoneID			string
	ShardID			string
	ExecutionOverlayID	string
	ExecutionGroupID	string
	BlockSTMGroupID		string
	ZoneSequence		uint64
	NetworkDeliveryOrdinal	uint64
	ConsensusScheduleID	string
	ConsensusScheduleHash	string
	ConsensusScheduleOrder	uint64
	Async			bool
	ParallelZoneExecution	bool
	DeterministicOrdering	bool
	CrossZone		CrossZoneMessage
}

type CrossZoneMessage struct {
	SourceZone	string
	DestinationZone	string
	SourceSequence	uint64
	MessageHash	string
	ExpiryHeight	uint64
	ReceiptPolicy	ReceiptPolicy
	ProofRequired	bool
}

type CrossZoneReceipt struct {
	ReceiptID	string
	SourceZone	string
	DestinationZone	string
	SourceSequence	uint64
	MessageHash	string
	Status		CrossZoneReceiptStatus
	ReceiptPolicy	ReceiptPolicy
	ProofHash	string
	ReceiptHeight	uint64
	RollbackSafe	bool
	ProofQueryable	bool
	Bounced		bool
	Error		string
}

type CrossZoneReplayGuard struct {
	ExecutedKeys []string
}

func NewExecutionMessageSchedule(schedule ExecutionMessageSchedule) (ExecutionMessageSchedule, error) {
	schedule = NormalizeExecutionMessageSchedule(schedule)
	if schedule.ScheduleHash == "" {
		schedule.ScheduleHash = ComputeExecutionMessageScheduleHash(schedule)
	}
	if schedule.ScheduleID == "" {
		schedule.ScheduleID = ComputeExecutionMessageScheduleID(schedule)
	}
	if err := schedule.Validate(); err != nil {
		return ExecutionMessageSchedule{}, err
	}
	return schedule, nil
}

func NormalizeExecutionMessageSchedule(schedule ExecutionMessageSchedule) ExecutionMessageSchedule {
	schedule.ScheduleID = normalizeHashText(schedule.ScheduleID)
	schedule.ZoneID = strings.TrimSpace(schedule.ZoneID)
	schedule.ShardID = strings.TrimSpace(schedule.ShardID)
	schedule.RoutingClass = ExecutionRoutingClass(strings.ToLower(strings.TrimSpace(string(schedule.RoutingClass))))
	schedule.ScheduleHash = normalizeHashText(schedule.ScheduleHash)
	schedule.TransactionIDs = normalizeHashSet(schedule.TransactionIDs)
	schedule.MessageIDs = normalizeHashSet(schedule.MessageIDs)
	return schedule
}

func ComputeExecutionMessageScheduleHash(schedule ExecutionMessageSchedule) string {
	schedule = NormalizeExecutionMessageSchedule(schedule)
	parts := []string{
		"execution-message-schedule",
		schedule.ZoneID,
		schedule.ShardID,
		string(schedule.RoutingClass),
		fmt.Sprintf("%t", schedule.Committed),
		fmt.Sprintf("%t", schedule.Ordered),
		fmt.Sprintf("%d", schedule.FirstZoneSequence),
		fmt.Sprintf("%d", schedule.LastZoneSequence),
	}
	parts = append(parts, schedule.TransactionIDs...)
	parts = append(parts, schedule.MessageIDs...)
	return HashParts(parts...)
}

func ComputeExecutionMessageScheduleID(schedule ExecutionMessageSchedule) string {
	schedule = NormalizeExecutionMessageSchedule(schedule)
	return HashParts("execution-message-schedule-id", schedule.ScheduleHash, schedule.ZoneID, schedule.ShardID)
}

func (s ExecutionMessageSchedule) Validate() error {
	schedule := NormalizeExecutionMessageSchedule(s)
	if err := ValidateHash("networking execution schedule id", schedule.ScheduleID); err != nil {
		return err
	}
	if err := ValidateHash("networking execution schedule hash", schedule.ScheduleHash); err != nil {
		return err
	}
	if schedule.ScheduleHash != ComputeExecutionMessageScheduleHash(schedule) {
		return errors.New("networking execution schedule hash does not match payload")
	}
	if schedule.ScheduleID != ComputeExecutionMessageScheduleID(schedule) {
		return errors.New("networking execution schedule id does not match payload")
	}
	if !IsExecutionRoutingClass(schedule.RoutingClass) {
		return fmt.Errorf("unknown networking execution routing class %q", schedule.RoutingClass)
	}
	if schedule.ZoneID == "" || schedule.ShardID == "" {
		return errors.New("networking execution schedule requires zone and shard")
	}
	if len(schedule.TransactionIDs) == 0 && len(schedule.MessageIDs) == 0 {
		return errors.New("networking execution schedule requires transactions or messages")
	}
	for _, id := range append(append([]string(nil), schedule.TransactionIDs...), schedule.MessageIDs...) {
		if err := ValidateHash("networking execution schedule item id", id); err != nil {
			return err
		}
	}
	if schedule.Ordered && (schedule.FirstZoneSequence == 0 || schedule.LastZoneSequence < schedule.FirstZoneSequence) {
		return errors.New("networking ordered execution schedule requires sequence range")
	}
	return nil
}

func NewExecutionZoneMessage(msg ExecutionZoneMessage, schedule ExecutionMessageSchedule) (ExecutionZoneMessage, error) {
	msg = NormalizeExecutionZoneMessage(msg)
	if schedule.ScheduleID != "" {
		schedule = NormalizeExecutionMessageSchedule(schedule)
		msg.ConsensusScheduleID = schedule.ScheduleID
		msg.ConsensusScheduleHash = schedule.ScheduleHash
		if schedule.Ordered {
			msg.DeterministicOrdering = true
		}
	}
	if err := msg.Validate(schedule); err != nil {
		return ExecutionZoneMessage{}, err
	}
	return msg, nil
}

func NormalizeExecutionZoneMessage(msg ExecutionZoneMessage) ExecutionZoneMessage {
	msg.Message = NormalizeAetherMeshMessage(msg.Message)
	msg.RoutingClass = ExecutionRoutingClass(strings.ToLower(strings.TrimSpace(string(msg.RoutingClass))))
	msg.ZoneID = strings.TrimSpace(msg.ZoneID)
	msg.ShardID = strings.TrimSpace(msg.ShardID)
	msg.ExecutionOverlayID = normalizeHashText(msg.ExecutionOverlayID)
	msg.ExecutionGroupID = normalizeHashText(msg.ExecutionGroupID)
	msg.BlockSTMGroupID = normalizeHashText(msg.BlockSTMGroupID)
	msg.ConsensusScheduleID = normalizeHashText(msg.ConsensusScheduleID)
	msg.ConsensusScheduleHash = normalizeHashText(msg.ConsensusScheduleHash)
	msg.CrossZone = NormalizeCrossZoneMessage(msg.CrossZone)
	return msg
}

func (m ExecutionZoneMessage) Validate(schedule ExecutionMessageSchedule) error {
	msg := NormalizeExecutionZoneMessage(m)
	if err := msg.Message.ValidateBasic(0); err != nil {
		return err
	}
	if !IsExecutionRoutingClass(msg.RoutingClass) {
		return fmt.Errorf("unknown networking execution routing class %q", msg.RoutingClass)
	}
	if msg.ZoneID == "" || msg.ShardID == "" {
		return errors.New("networking execution zone message requires zone and shard")
	}
	if err := ValidateHash("networking execution overlay id", msg.ExecutionOverlayID); err != nil {
		return err
	}
	if msg.ZoneSequence == 0 {
		return errors.New("networking execution zone message requires zone sequence")
	}
	if msg.Message.DestinationZone != "" && msg.Message.DestinationZone != msg.ZoneID {
		return errors.New("networking execution zone message destination zone mismatch")
	}
	if msg.Message.Type == MeshMessageCrossZone {
		if err := msg.CrossZone.Validate(0); err != nil {
			return err
		}
		if msg.CrossZone.SourceSequence != msg.Message.Sequence {
			return errors.New("networking cross-zone source sequence must match mesh sequence")
		}
		if msg.CrossZone.SourceZone != msg.Message.SourceZone || msg.CrossZone.DestinationZone != msg.Message.DestinationZone {
			return errors.New("networking cross-zone envelope zones must match mesh message")
		}
	}
	if msg.ParallelZoneExecution && msg.BlockSTMGroupID == "" {
		return errors.New("networking parallel zone execution requires BlockSTM group id")
	}
	if msg.Async && msg.ExecutionGroupID == "" {
		return errors.New("networking async execution requires execution group id")
	}
	if msg.DeterministicOrdering || msg.Message.ConsensusEffect {
		if err := validateExecutionScheduleBinding(msg, schedule); err != nil {
			return err
		}
	}
	return nil
}

func NewCrossZoneMessage(msg CrossZoneMessage) (CrossZoneMessage, error) {
	msg = NormalizeCrossZoneMessage(msg)
	if err := msg.Validate(0); err != nil {
		return CrossZoneMessage{}, err
	}
	return msg, nil
}

func NormalizeCrossZoneMessage(msg CrossZoneMessage) CrossZoneMessage {
	msg.SourceZone = strings.TrimSpace(msg.SourceZone)
	msg.DestinationZone = strings.TrimSpace(msg.DestinationZone)
	msg.MessageHash = normalizeHashText(msg.MessageHash)
	msg.ReceiptPolicy = ReceiptPolicy(strings.ToLower(strings.TrimSpace(string(msg.ReceiptPolicy))))
	if msg.ReceiptPolicy == "" {
		msg.ReceiptPolicy = ReceiptPolicyOnExecution
	}
	return msg
}

func ComputeCrossZoneExecutionKey(msg CrossZoneMessage) string {
	msg = NormalizeCrossZoneMessage(msg)
	return HashParts("cross-zone-execution-key", msg.SourceZone, msg.DestinationZone, fmt.Sprintf("%d", msg.SourceSequence), msg.MessageHash)
}

func (m CrossZoneMessage) Validate(currentHeight uint64) error {
	msg := NormalizeCrossZoneMessage(m)
	if msg.SourceZone == "" || msg.DestinationZone == "" {
		return errors.New("networking cross-zone message requires source and destination zones")
	}
	if msg.SourceZone == msg.DestinationZone {
		return errors.New("networking cross-zone message requires distinct zones")
	}
	if msg.SourceSequence == 0 {
		return errors.New("networking cross-zone message requires source sequence")
	}
	if err := ValidateHash("networking cross-zone message hash", msg.MessageHash); err != nil {
		return err
	}
	if msg.ExpiryHeight == 0 {
		return errors.New("networking cross-zone message requires expiry height")
	}
	if currentHeight > 0 && currentHeight > msg.ExpiryHeight {
		return errors.New("networking cross-zone message is expired")
	}
	if !IsReceiptPolicy(msg.ReceiptPolicy) {
		return fmt.Errorf("unknown networking receipt policy %q", msg.ReceiptPolicy)
	}
	return nil
}

func NewCrossZoneReceipt(receipt CrossZoneReceipt) (CrossZoneReceipt, error) {
	receipt = NormalizeCrossZoneReceipt(receipt)
	if receipt.ReceiptID == "" {
		receipt.ReceiptID = ComputeCrossZoneReceiptID(receipt)
	}
	if err := receipt.Validate(); err != nil {
		return CrossZoneReceipt{}, err
	}
	return receipt, nil
}

func NormalizeCrossZoneReceipt(receipt CrossZoneReceipt) CrossZoneReceipt {
	receipt.ReceiptID = normalizeHashText(receipt.ReceiptID)
	receipt.SourceZone = strings.TrimSpace(receipt.SourceZone)
	receipt.DestinationZone = strings.TrimSpace(receipt.DestinationZone)
	receipt.MessageHash = normalizeHashText(receipt.MessageHash)
	receipt.Status = CrossZoneReceiptStatus(strings.ToLower(strings.TrimSpace(string(receipt.Status))))
	receipt.ReceiptPolicy = ReceiptPolicy(strings.ToLower(strings.TrimSpace(string(receipt.ReceiptPolicy))))
	receipt.ProofHash = normalizeHashText(receipt.ProofHash)
	receipt.Error = strings.TrimSpace(receipt.Error)
	return receipt
}

func ComputeCrossZoneReceiptID(receipt CrossZoneReceipt) string {
	receipt = NormalizeCrossZoneReceipt(receipt)
	return HashParts(
		"cross-zone-receipt",
		receipt.SourceZone,
		receipt.DestinationZone,
		fmt.Sprintf("%d", receipt.SourceSequence),
		receipt.MessageHash,
		string(receipt.Status),
		string(receipt.ReceiptPolicy),
		receipt.ProofHash,
		fmt.Sprintf("%d", receipt.ReceiptHeight),
		fmt.Sprintf("%t", receipt.RollbackSafe),
		fmt.Sprintf("%t", receipt.ProofQueryable),
		fmt.Sprintf("%t", receipt.Bounced),
	)
}

func (r CrossZoneReceipt) Validate() error {
	receipt := NormalizeCrossZoneReceipt(r)
	if err := ValidateHash("networking cross-zone receipt id", receipt.ReceiptID); err != nil {
		return err
	}
	if receipt.ReceiptID != ComputeCrossZoneReceiptID(receipt) {
		return errors.New("networking cross-zone receipt id does not match payload")
	}
	if receipt.SourceZone == "" || receipt.DestinationZone == "" || receipt.SourceZone == receipt.DestinationZone {
		return errors.New("networking cross-zone receipt requires distinct zones")
	}
	if receipt.SourceSequence == 0 {
		return errors.New("networking cross-zone receipt requires source sequence")
	}
	if err := ValidateHash("networking cross-zone receipt message hash", receipt.MessageHash); err != nil {
		return err
	}
	if !IsCrossZoneReceiptStatus(receipt.Status) {
		return fmt.Errorf("unknown networking cross-zone receipt status %q", receipt.Status)
	}
	if !IsReceiptPolicy(receipt.ReceiptPolicy) {
		return fmt.Errorf("unknown networking receipt policy %q", receipt.ReceiptPolicy)
	}
	if receipt.ReceiptHeight == 0 {
		return errors.New("networking cross-zone receipt height must be positive")
	}
	if !receipt.RollbackSafe || !receipt.ProofQueryable {
		return errors.New("networking cross-zone receipt must be rollback-safe and proof-queryable")
	}
	if receipt.ReceiptPolicy != ReceiptPolicyNone {
		if err := ValidateHash("networking cross-zone receipt proof hash", receipt.ProofHash); err != nil {
			return err
		}
	}
	if receipt.Status == CrossZoneReceiptBounced && !receipt.Bounced {
		return errors.New("networking bounced receipt must set bounced flag")
	}
	return nil
}

func (g CrossZoneReplayGuard) Accept(msg CrossZoneMessage, currentHeight uint64) (CrossZoneReplayGuard, error) {
	msg = NormalizeCrossZoneMessage(msg)
	if err := msg.Validate(currentHeight); err != nil {
		return CrossZoneReplayGuard{}, err
	}
	key := ComputeCrossZoneExecutionKey(msg)
	for _, existing := range g.ExecutedKeys {
		if normalizeHashText(existing) == key {
			return CrossZoneReplayGuard{}, errors.New("networking cross-zone message already executed")
		}
	}
	next := CrossZoneReplayGuard{ExecutedKeys: append([]string(nil), g.ExecutedKeys...)}
	next.ExecutedKeys = append(next.ExecutedKeys, key)
	sortStrings(next.ExecutedKeys)
	return next, nil
}

func IsExecutionRoutingClass(class ExecutionRoutingClass) bool {
	switch class {
	case ExecutionRoutingTx, ExecutionRoutingZone, ExecutionRoutingShard, ExecutionRoutingExecutionOverlay:
		return true
	default:
		return false
	}
}

func IsReceiptPolicy(policy ReceiptPolicy) bool {
	switch policy {
	case ReceiptPolicyNone, ReceiptPolicyOnDelivery, ReceiptPolicyOnExecution, ReceiptPolicyOnFailure, ReceiptPolicyAlways:
		return true
	default:
		return false
	}
}

func IsCrossZoneReceiptStatus(status CrossZoneReceiptStatus) bool {
	switch status {
	case CrossZoneReceiptDelivered, CrossZoneReceiptExecuted, CrossZoneReceiptExpired, CrossZoneReceiptBounced, CrossZoneReceiptRolledBack:
		return true
	default:
		return false
	}
}

func validateExecutionScheduleBinding(msg ExecutionZoneMessage, schedule ExecutionMessageSchedule) error {
	schedule = NormalizeExecutionMessageSchedule(schedule)
	if err := schedule.Validate(); err != nil {
		return err
	}
	if !schedule.Committed {
		return errors.New("networking consensus execution order requires committed schedule")
	}
	if msg.ConsensusScheduleID != schedule.ScheduleID || msg.ConsensusScheduleHash != schedule.ScheduleHash {
		return errors.New("networking execution message schedule binding mismatch")
	}
	if msg.ConsensusScheduleOrder == 0 {
		return errors.New("networking execution message requires consensus schedule order")
	}
	if msg.ZoneSequence < schedule.FirstZoneSequence || msg.ZoneSequence > schedule.LastZoneSequence {
		return errors.New("networking execution message zone sequence outside committed schedule")
	}
	if !containsString(schedule.MessageIDs, msg.Message.MessageID) && !containsString(schedule.TransactionIDs, msg.Message.MessageID) {
		return errors.New("networking execution message not present in committed schedule")
	}
	return nil
}

func normalizeHashSet(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = normalizeHashText(value)
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
