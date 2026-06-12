package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	SyncStageTx		SyncExecutionStage	= "tx"
	SyncStageAnte		SyncExecutionStage	= "ante_handler"
	SyncStageMsgServer	SyncExecutionStage	= "msg_server"
	SyncStageKeeper		SyncExecutionStage	= "keeper"
	SyncStageStore		SyncExecutionStage	= "store_v2_kvstore"
	SyncStageEvents		SyncExecutionStage	= "events"
	SyncStageReceipt	SyncExecutionStage	= "receipt"

	SyncModuleBank			SyncModule	= "bank"
	SyncModuleStaking		SyncModule	= "staking"
	SyncModuleGovernance		SyncModule	= "governance"
	SyncModuleContractAssets	SyncModule	= "contract_assets"
	SyncModuleDEX			SyncModule	= "dex"
	SyncModuleIdentity		SyncModule	= "identity"
	SyncModulePayments		SyncModule	= "payments"

	SyncReceiptCommitted	SyncReceiptPolicy	= "committed"
	SyncReceiptDisabled	SyncReceiptPolicy	= "disabled"

	AsyncEngineLaneScheduled	AsyncEngineLane	= "scheduled"
	AsyncEngineLaneCrossZone	AsyncEngineLane	= "cross_zone"
	AsyncEngineLaneContinuation	AsyncEngineLane	= "continuation"

	MaxSyncEngineEvents		= 256
	MaxSyncEngineWrites		= 256
	MaxSyncEngineDetailLength	= 128
	MaxAsyncContinuations		= 256
	MaxContinuationTokenLength	= 128
)

type SyncExecutionStage string
type SyncModule string
type SyncReceiptPolicy string
type AsyncEngineLane string

type SyncExecutionStep struct {
	Stage	SyncExecutionStage
	Detail	string
}

type SyncStateWrite struct {
	StoreKey	string
	Key		string
}

type SyncExecutionPlan struct {
	Height		uint64
	Module		SyncModule
	Route		string
	GasLimit	uint64
	GasUsed		uint64
	Atomic		bool
	ReceiptPolicy	SyncReceiptPolicy
	Steps		[]SyncExecutionStep
	StateWrites	[]SyncStateWrite
	Events		[]string
	ResultCode	uint32
	Error		string
	ReceiptRoot	string
}

type AsyncContinuation struct {
	Token		string
	ZoneID		zonestypes.ZoneID
	Contract	string
	DeliverAtBlock	uint64
	DeadlineBlock	uint64
	StateRoot	string
}

type AsyncZoneQueue struct {
	ZoneID		zonestypes.ZoneID
	QueueID		string
	Lane		AsyncEngineLane
	Messages	[]async.MessageEnvelope
	MaxMessages	uint32
	MaxGas		uint64
}

type AsyncRetryPolicy struct {
	MaxRetries		uint32
	RetryDelayBlocks	uint64
	DeadlineBlock		uint64
	Bounce			bool
	DeadLetter		bool
}

type AsyncExecutionPlan struct {
	Height		uint64
	Queues		[]AsyncZoneQueue
	RetryPolicy	AsyncRetryPolicy
	Receipts	[]async.ExecutionReceipt
	DeadLetters	[]async.DeadLetter
	Continuations	[]AsyncContinuation
	PlanRoot	string
}

func NewSyncExecutionPlan(plan SyncExecutionPlan) (SyncExecutionPlan, error) {
	plan = canonicalSyncExecutionPlan(plan)
	if plan.ReceiptPolicy == "" {
		plan.ReceiptPolicy = SyncReceiptCommitted
	}
	plan.ReceiptRoot = ComputeSyncExecutionReceiptRoot(plan)
	return plan, plan.Validate()
}

func (p SyncExecutionPlan) Validate() error {
	p = canonicalSyncExecutionPlan(p)
	if p.Height == 0 {
		return errors.New("sync execution height must be positive")
	}
	if !IsSyncModule(p.Module) {
		return fmt.Errorf("unsupported sync execution module %q", p.Module)
	}
	if strings.TrimSpace(p.Route) != p.Route || p.Route == "" {
		return errors.New("sync execution route is required")
	}
	if p.GasLimit == 0 {
		return errors.New("sync execution gas limit must be positive")
	}
	if p.GasUsed > p.GasLimit {
		return errors.New("sync execution gas used exceeds limit")
	}
	if !p.Atomic {
		return errors.New("sync execution must be atomic")
	}
	if !IsSyncReceiptPolicy(p.ReceiptPolicy) {
		return fmt.Errorf("invalid sync execution receipt policy %q", p.ReceiptPolicy)
	}
	if err := validateSyncSteps(p.Steps); err != nil {
		return err
	}
	if len(p.Events) > MaxSyncEngineEvents {
		return fmt.Errorf("sync execution events must be <= %d", MaxSyncEngineEvents)
	}
	if len(p.StateWrites) > MaxSyncEngineWrites {
		return fmt.Errorf("sync execution writes must be <= %d", MaxSyncEngineWrites)
	}
	if err := validateSyncWrites(p.StateWrites); err != nil {
		return err
	}
	if err := validateEngineTokens("sync execution event", p.Events, MaxSyncEngineDetailLength); err != nil {
		return err
	}
	if p.Error != "" && (len(p.StateWrites) > 0 || len(p.Events) > 0) {
		return errors.New("failed sync execution must not commit state writes or events")
	}
	if p.ReceiptPolicy == SyncReceiptCommitted && p.ReceiptRoot == "" {
		return errors.New("sync execution committed receipt root is required")
	}
	if p.ReceiptPolicy == SyncReceiptCommitted && p.ReceiptRoot != ComputeSyncExecutionReceiptRoot(p) {
		return errors.New("sync execution receipt root mismatch")
	}
	return nil
}

func NewAsyncExecutionPlan(plan AsyncExecutionPlan) (AsyncExecutionPlan, error) {
	plan = canonicalAsyncExecutionPlan(plan)
	plan.PlanRoot = ComputeAsyncExecutionPlanRoot(plan)
	return plan, plan.Validate()
}

func (p AsyncExecutionPlan) Validate() error {
	p = canonicalAsyncExecutionPlan(p)
	if p.Height == 0 {
		return errors.New("async execution height must be positive")
	}
	if err := p.RetryPolicy.Validate(); err != nil {
		return err
	}
	seenQueues := make(map[string]struct{}, len(p.Queues))
	for i, queue := range p.Queues {
		if err := queue.Validate(p.Height); err != nil {
			return err
		}
		key := string(queue.ZoneID) + "/" + queue.QueueID
		if _, found := seenQueues[key]; found {
			return fmt.Errorf("duplicate async zone queue %q", key)
		}
		seenQueues[key] = struct{}{}
		if i > 0 && compareAsyncZoneQueues(p.Queues[i-1], queue) >= 0 {
			return errors.New("async zone queues must be sorted canonically")
		}
	}
	if err := validateAsyncReceipts(p.Receipts); err != nil {
		return err
	}
	if err := validateAsyncDeadLetters(p.DeadLetters, p.Height); err != nil {
		return err
	}
	if len(p.DeadLetters) > 0 && !p.RetryPolicy.DeadLetter {
		return errors.New("async dead letters require dead letter policy")
	}
	if len(p.Continuations) > MaxAsyncContinuations {
		return fmt.Errorf("async continuations must be <= %d", MaxAsyncContinuations)
	}
	for i, continuation := range p.Continuations {
		if err := continuation.Validate(p.Height); err != nil {
			return err
		}
		if i > 0 && compareAsyncContinuations(p.Continuations[i-1], continuation) >= 0 {
			return errors.New("async continuations must be sorted canonically")
		}
	}
	if p.PlanRoot == "" {
		return errors.New("async execution plan root is required")
	}
	if p.PlanRoot != ComputeAsyncExecutionPlanRoot(p) {
		return errors.New("async execution plan root mismatch")
	}
	return nil
}

func (q AsyncZoneQueue) Validate(height uint64) error {
	if err := zonestypes.ValidateZoneID(q.ZoneID); err != nil {
		return err
	}
	if err := validateEngineToken("async queue id", q.QueueID, MaxSyncEngineDetailLength); err != nil {
		return err
	}
	if !IsAsyncEngineLane(q.Lane) {
		return fmt.Errorf("invalid async engine lane %q", q.Lane)
	}
	if q.MaxMessages == 0 {
		return errors.New("async zone queue max messages must be positive")
	}
	if q.MaxGas == 0 {
		return errors.New("async zone queue max gas must be positive")
	}
	if len(q.Messages) > int(q.MaxMessages) {
		return fmt.Errorf("async zone queue messages must be <= %d", q.MaxMessages)
	}
	var gas uint64
	var previous async.MessageEnvelope
	for i, msg := range q.Messages {
		if msg.GasLimit == 0 {
			return errors.New("async message gas limit must be positive")
		}
		if msg.DeadlineBlock != 0 && msg.DeadlineBlock < height {
			return errors.New("async message deadline is expired")
		}
		if msg.GasLimit > ^uint64(0)-gas {
			return errors.New("async zone queue gas overflow")
		}
		gas += msg.GasLimit
		if i > 0 && compareAsyncMessages(previous, msg) >= 0 {
			return errors.New("async zone queue messages must be sorted canonically")
		}
		previous = msg
	}
	if gas > q.MaxGas {
		return errors.New("async zone queue gas exceeds max gas")
	}
	return nil
}

func (p AsyncRetryPolicy) Validate() error {
	if p.MaxRetries > 0 && p.RetryDelayBlocks == 0 {
		return errors.New("async retry delay must be positive when retries are enabled")
	}
	if p.DeadLetter && p.MaxRetries == 0 {
		return errors.New("async dead letters require bounded retry policy")
	}
	return nil
}

func (c AsyncContinuation) Validate(height uint64) error {
	if err := validateEngineToken("async continuation token", c.Token, MaxContinuationTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if err := validateEngineToken("async continuation contract", c.Contract, MaxSyncEngineDetailLength); err != nil {
		return err
	}
	if c.DeliverAtBlock <= height {
		return errors.New("async continuation must resume in a future block")
	}
	if c.DeadlineBlock != 0 && c.DeadlineBlock < c.DeliverAtBlock {
		return errors.New("async continuation deadline must not precede delivery block")
	}
	return zonestypes.ValidateHash("async continuation state root", c.StateRoot)
}

func IsSyncModule(module SyncModule) bool {
	switch module {
	case SyncModuleBank,
		SyncModuleStaking,
		SyncModuleGovernance,
		SyncModuleContractAssets,
		SyncModuleDEX,
		SyncModuleIdentity,
		SyncModulePayments:
		return true
	default:
		return false
	}
}

func IsSyncReceiptPolicy(policy SyncReceiptPolicy) bool {
	switch policy {
	case SyncReceiptCommitted, SyncReceiptDisabled:
		return true
	default:
		return false
	}
}

func IsAsyncEngineLane(lane AsyncEngineLane) bool {
	switch lane {
	case AsyncEngineLaneScheduled, AsyncEngineLaneCrossZone, AsyncEngineLaneContinuation:
		return true
	default:
		return false
	}
}

func ComputeSyncExecutionReceiptRoot(plan SyncExecutionPlan) string {
	plan = canonicalSyncExecutionPlan(plan)
	h := sha256.New()
	writeEnginePart(h, "aetra-sync-engine-receipt-v1")
	writeEngineUint64(h, plan.Height)
	writeEnginePart(h, string(plan.Module))
	writeEnginePart(h, plan.Route)
	writeEngineUint64(h, plan.GasLimit)
	writeEngineUint64(h, plan.GasUsed)
	writeEnginePart(h, string(plan.ReceiptPolicy))
	writeEngineUint64(h, uint64(plan.ResultCode))
	writeEnginePart(h, plan.Error)
	writeEngineUint64(h, uint64(len(plan.Steps)))
	for _, step := range plan.Steps {
		writeEnginePart(h, string(step.Stage))
		writeEnginePart(h, step.Detail)
	}
	writeEngineUint64(h, uint64(len(plan.StateWrites)))
	for _, write := range plan.StateWrites {
		writeEnginePart(h, write.StoreKey)
		writeEnginePart(h, write.Key)
	}
	writeEngineUint64(h, uint64(len(plan.Events)))
	for _, event := range plan.Events {
		writeEnginePart(h, event)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAsyncExecutionPlanRoot(plan AsyncExecutionPlan) string {
	plan = canonicalAsyncExecutionPlan(plan)
	h := sha256.New()
	writeEnginePart(h, "aetra-async-engine-plan-v1")
	writeEngineUint64(h, plan.Height)
	writeEngineUint64(h, uint64(plan.RetryPolicy.MaxRetries))
	writeEngineUint64(h, plan.RetryPolicy.RetryDelayBlocks)
	writeEngineUint64(h, plan.RetryPolicy.DeadlineBlock)
	writeEngineBool(h, plan.RetryPolicy.Bounce)
	writeEngineBool(h, plan.RetryPolicy.DeadLetter)
	writeEngineUint64(h, uint64(len(plan.Queues)))
	for _, queue := range plan.Queues {
		writeEnginePart(h, string(queue.ZoneID))
		writeEnginePart(h, queue.QueueID)
		writeEnginePart(h, string(queue.Lane))
		writeEngineUint64(h, uint64(queue.MaxMessages))
		writeEngineUint64(h, queue.MaxGas)
		writeEngineUint64(h, uint64(len(queue.Messages)))
		for _, msg := range queue.Messages {
			writeAsyncMessage(h, msg)
		}
	}
	writeEngineUint64(h, uint64(len(plan.Receipts)))
	for _, receipt := range plan.Receipts {
		writeAsyncReceipt(h, receipt)
	}
	writeEngineUint64(h, uint64(len(plan.DeadLetters)))
	for _, dead := range plan.DeadLetters {
		writeEngineUint64(h, dead.Sequence)
		writeEngineUint64(h, dead.FailedSequence)
		writeEngineUint64(h, dead.RecordedBlock)
		writeAsyncMessage(h, dead.Envelope)
		writeAsyncReceipt(h, dead.Receipt)
		writeEnginePart(h, dead.Reason)
	}
	writeEngineUint64(h, uint64(len(plan.Continuations)))
	for _, continuation := range plan.Continuations {
		writeEnginePart(h, continuation.Token)
		writeEnginePart(h, string(continuation.ZoneID))
		writeEnginePart(h, continuation.Contract)
		writeEngineUint64(h, continuation.DeliverAtBlock)
		writeEngineUint64(h, continuation.DeadlineBlock)
		writeEnginePart(h, continuation.StateRoot)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalSyncExecutionPlan(plan SyncExecutionPlan) SyncExecutionPlan {
	plan.Route = strings.TrimSpace(plan.Route)
	plan.Error = strings.TrimSpace(plan.Error)
	plan.ReceiptRoot = strings.TrimSpace(plan.ReceiptRoot)
	plan.Steps = cloneSyncSteps(plan.Steps)
	plan.StateWrites = cloneSyncWrites(plan.StateWrites)
	plan.Events = cloneSortedStrings(plan.Events)
	sort.SliceStable(plan.StateWrites, func(i, j int) bool {
		if plan.StateWrites[i].StoreKey != plan.StateWrites[j].StoreKey {
			return plan.StateWrites[i].StoreKey < plan.StateWrites[j].StoreKey
		}
		return plan.StateWrites[i].Key < plan.StateWrites[j].Key
	})
	return plan
}

func canonicalAsyncExecutionPlan(plan AsyncExecutionPlan) AsyncExecutionPlan {
	plan.Queues = cloneAsyncZoneQueues(plan.Queues)
	plan.Receipts = append([]async.ExecutionReceipt(nil), plan.Receipts...)
	plan.DeadLetters = append([]async.DeadLetter(nil), plan.DeadLetters...)
	plan.Continuations = append([]AsyncContinuation(nil), plan.Continuations...)
	plan.PlanRoot = strings.TrimSpace(plan.PlanRoot)
	sort.SliceStable(plan.Queues, func(i, j int) bool {
		return compareAsyncZoneQueues(plan.Queues[i], plan.Queues[j]) < 0
	})
	sort.SliceStable(plan.Receipts, func(i, j int) bool {
		return plan.Receipts[i].Sequence < plan.Receipts[j].Sequence
	})
	sort.SliceStable(plan.DeadLetters, func(i, j int) bool {
		return plan.DeadLetters[i].Sequence < plan.DeadLetters[j].Sequence
	})
	sort.SliceStable(plan.Continuations, func(i, j int) bool {
		return compareAsyncContinuations(plan.Continuations[i], plan.Continuations[j]) < 0
	})
	return plan
}

func validateSyncSteps(steps []SyncExecutionStep) error {
	expected := []SyncExecutionStage{
		SyncStageTx,
		SyncStageAnte,
		SyncStageMsgServer,
		SyncStageKeeper,
		SyncStageStore,
		SyncStageEvents,
		SyncStageReceipt,
	}
	if len(steps) != len(expected) {
		return errors.New("sync execution must include tx -> ante -> msgserver -> keeper -> store -> events -> receipt")
	}
	for i, step := range steps {
		if step.Stage != expected[i] {
			return fmt.Errorf("sync execution stage %d must be %q", i, expected[i])
		}
		if err := validateEngineToken("sync execution step detail", step.Detail, MaxSyncEngineDetailLength); err != nil {
			return err
		}
	}
	return nil
}

func validateSyncWrites(writes []SyncStateWrite) error {
	var previous SyncStateWrite
	for i, write := range writes {
		if err := validateEngineToken("sync execution store key", write.StoreKey, MaxSyncEngineDetailLength); err != nil {
			return err
		}
		if err := validateEngineToken("sync execution write key", write.Key, MaxSyncEngineDetailLength); err != nil {
			return err
		}
		if i > 0 {
			if previous.StoreKey > write.StoreKey || previous.StoreKey == write.StoreKey && previous.Key >= write.Key {
				return errors.New("sync execution writes must be sorted canonically")
			}
		}
		previous = write
	}
	return nil
}

func validateAsyncReceipts(receipts []async.ExecutionReceipt) error {
	var previous uint64
	for i, receipt := range receipts {
		if i > 0 && previous >= receipt.Sequence {
			return errors.New("async execution receipts must be sorted canonically")
		}
		if receipt.GasUsed == 0 {
			return errors.New("async execution receipt gas used must be positive")
		}
		if receipt.RetryScheduled && receipt.ResultCode == async.ResultOK {
			return errors.New("async execution successful receipt must not schedule retry")
		}
		previous = receipt.Sequence
	}
	return nil
}

func validateAsyncDeadLetters(deadLetters []async.DeadLetter, height uint64) error {
	var previous uint64
	for i, dead := range deadLetters {
		if i > 0 && previous >= dead.Sequence {
			return errors.New("async dead letters must be sorted canonically")
		}
		if dead.RecordedBlock > height {
			return errors.New("async dead letter recorded block exceeds plan height")
		}
		if strings.TrimSpace(dead.Reason) == "" {
			return errors.New("async dead letter reason is required")
		}
		previous = dead.Sequence
	}
	return nil
}

func validateEngineTokens(fieldName string, values []string, maxLen int) error {
	var previous string
	for i, value := range values {
		if err := validateEngineToken(fieldName, value, maxLen); err != nil {
			return err
		}
		if i > 0 && previous >= value {
			return fmt.Errorf("%s list must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func validateEngineToken(fieldName, value string, maxLen int) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxLen)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func cloneSyncSteps(steps []SyncExecutionStep) []SyncExecutionStep {
	out := append([]SyncExecutionStep(nil), steps...)
	for i := range out {
		out[i].Detail = strings.TrimSpace(out[i].Detail)
	}
	return out
}

func cloneSyncWrites(writes []SyncStateWrite) []SyncStateWrite {
	out := append([]SyncStateWrite(nil), writes...)
	for i := range out {
		out[i].StoreKey = strings.TrimSpace(out[i].StoreKey)
		out[i].Key = strings.TrimSpace(out[i].Key)
	}
	return out
}

func cloneAsyncZoneQueues(queues []AsyncZoneQueue) []AsyncZoneQueue {
	out := make([]AsyncZoneQueue, len(queues))
	for i, queue := range queues {
		out[i] = queue
		out[i].QueueID = strings.TrimSpace(queue.QueueID)
		out[i].Messages = append([]async.MessageEnvelope(nil), queue.Messages...)
		sort.SliceStable(out[i].Messages, func(left, right int) bool {
			return compareAsyncMessages(out[i].Messages[left], out[i].Messages[right]) < 0
		})
	}
	return out
}

func cloneSortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	for i, value := range out {
		out[i] = strings.TrimSpace(value)
	}
	sort.Strings(out)
	return out
}

func compareAsyncZoneQueues(left, right AsyncZoneQueue) int {
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.QueueID < right.QueueID {
		return -1
	}
	if left.QueueID > right.QueueID {
		return 1
	}
	return 0
}

func compareAsyncMessages(left, right async.MessageEnvelope) int {
	if left.DeliverAtBlock < right.DeliverAtBlock {
		return -1
	}
	if left.DeliverAtBlock > right.DeliverAtBlock {
		return 1
	}
	if left.CreatedLogicalTime < right.CreatedLogicalTime {
		return -1
	}
	if left.CreatedLogicalTime > right.CreatedLogicalTime {
		return 1
	}
	if string(left.Destination) < string(right.Destination) {
		return -1
	}
	if string(left.Destination) > string(right.Destination) {
		return 1
	}
	if left.QueryID < right.QueryID {
		return -1
	}
	if left.QueryID > right.QueryID {
		return 1
	}
	return 0
}

func compareAsyncContinuations(left, right AsyncContinuation) int {
	if left.DeliverAtBlock < right.DeliverAtBlock {
		return -1
	}
	if left.DeliverAtBlock > right.DeliverAtBlock {
		return 1
	}
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.Token < right.Token {
		return -1
	}
	if left.Token > right.Token {
		return 1
	}
	return 0
}

func writeEnginePart(w engineByteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func writeEngineUint64(w engineByteWriter, value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	_, _ = w.Write(out[:])
}

func writeEngineBool(w engineByteWriter, value bool) {
	if value {
		_, _ = w.Write([]byte{1})
		return
	}
	_, _ = w.Write([]byte{0})
}

func writeAsyncMessage(w engineByteWriter, msg async.MessageEnvelope) {
	writeEnginePart(w, string(msg.Source))
	writeEnginePart(w, string(msg.Destination))
	writeEnginePart(w, msg.Value.String())
	writeEngineUint64(w, uint64(msg.Opcode))
	writeEngineUint64(w, msg.QueryID)
	writeEnginePart(w, string(msg.Body))
	writeEngineBool(w, msg.Bounce)
	writeEngineBool(w, msg.Bounced)
	writeEngineUint64(w, msg.CreatedLogicalTime)
	writeEngineUint64(w, msg.DeliverAtBlock)
	writeEngineUint64(w, uint64(msg.RetryCount))
	writeEngineUint64(w, uint64(msg.MaxRetries))
	writeEngineUint64(w, msg.RetryDelayBlocks)
	writeEngineUint64(w, msg.ExecutionBlockHeight)
	writeEngineUint64(w, msg.DeadlineBlock)
	writeEngineUint64(w, msg.GasLimit)
	writeEnginePart(w, msg.ForwardFee.String())
	writeEngineUint64(w, uint64(msg.Depth))
}

func writeAsyncReceipt(w engineByteWriter, receipt async.ExecutionReceipt) {
	writeEngineUint64(w, receipt.Sequence)
	writeEnginePart(w, string(receipt.Source))
	writeEnginePart(w, string(receipt.Destination))
	writeEngineUint64(w, uint64(receipt.Opcode))
	writeEngineUint64(w, receipt.QueryID)
	writeEngineUint64(w, uint64(receipt.ResultCode))
	writeEngineUint64(w, receipt.GasUsed)
	writeEnginePart(w, receipt.StorageFeeNaet.String())
	writeEnginePart(w, receipt.ForwardFeeNaet.String())
	writeEngineBool(w, receipt.Bounced)
	writeEngineUint64(w, uint64(receipt.RetryCount))
	writeEngineBool(w, receipt.RetryScheduled)
	writeEnginePart(w, receipt.Error)
}

type engineByteWriter interface {
	Write([]byte) (int, error)
}
