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
	ContinuationStatusPending	ContinuationStatus	= "pending"
	ContinuationStatusScheduled	ContinuationStatus	= "scheduled"
	ContinuationStatusResumed	ContinuationStatus	= "resumed"
	ContinuationStatusExpired	ContinuationStatus	= "expired"
	ContinuationStatusFailed	ContinuationStatus	= "failed"

	ContinuationResumeByScheduler	= "scheduler"

	MaxActorRuntimeMailboxMessages		= 1024
	MaxActorRuntimeEmittedMessages		= 256
	MaxActorRuntimeStateWrites		= 256
	MaxActorRuntimeTokenLength		= 128
	MaxContinuationPartialStateBytes	= 64 * 1024
	MaxContinuationRecordsPerActorBlock	= 256
)

type ContinuationStatus string

type ActorRuntimeActor struct {
	ActorID		string
	CodeRef		string
	StateRoot	string
	Mailbox		[]ActorMailboxMessage
}

type ActorMailboxMessage struct {
	Sequence		uint64
	SourceActor		string
	TargetActor		string
	CreatedLogicalTime	uint64
	Envelope		async.MessageEnvelope
}

type ActorStateWrite struct {
	ActorID	string
	Key	string
	Hash	string
}

type ActorExecution struct {
	ActorID		string
	MessageSequence	uint64
	Handler		string
	GasLimit	uint64
	GasUsed		uint64
	StateWrites	[]ActorStateWrite
	EmittedMessages	[]async.MessageEnvelope
	ResultCode	uint32
	Error		string
}

type ContinuationRecord struct {
	ContinuationID		string
	ActorID			string
	StepIndex		uint32
	PartialStateHash	string
	PartialStateBytes	uint32
	ResumeHeight		uint64
	ExpiryHeight		uint64
	GasReserved		uint64
	Status			ContinuationStatus
	ResumeBy		string
	FailureReceipt		async.ExecutionReceipt
}

type ActorRuntimePlan struct {
	Height		uint64
	Actors		[]ActorRuntimeActor
	Executions	[]ActorExecution
	Continuations	[]ContinuationRecord
	PlanRoot	string
}

func NewActorRuntimePlan(plan ActorRuntimePlan) (ActorRuntimePlan, error) {
	plan = canonicalActorRuntimePlan(plan)
	plan.PlanRoot = ComputeActorRuntimePlanRoot(plan)
	return plan, plan.Validate()
}

func (p ActorRuntimePlan) Validate() error {
	p = canonicalActorRuntimePlan(p)
	if p.Height == 0 {
		return errors.New("actor runtime height must be positive")
	}
	actors := make(map[string]ActorRuntimeActor, len(p.Actors))
	for i, actor := range p.Actors {
		if err := actor.Validate(); err != nil {
			return err
		}
		if _, found := actors[actor.ActorID]; found {
			return fmt.Errorf("duplicate actor runtime actor %q", actor.ActorID)
		}
		actors[actor.ActorID] = actor
		if i > 0 && p.Actors[i-1].ActorID >= actor.ActorID {
			return errors.New("actor runtime actors must be sorted canonically")
		}
	}
	if err := validateActorExecutions(p.Executions, actors); err != nil {
		return err
	}
	if err := validateContinuationRecords(p.Continuations, actors, p.Height); err != nil {
		return err
	}
	if p.PlanRoot == "" {
		return errors.New("actor runtime plan root is required")
	}
	if p.PlanRoot != ComputeActorRuntimePlanRoot(p) {
		return errors.New("actor runtime plan root mismatch")
	}
	return nil
}

func (a ActorRuntimeActor) Validate() error {
	if err := validateEngineToken("actor runtime actor id", a.ActorID, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("actor runtime code ref", a.CodeRef, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("actor runtime state root", a.StateRoot); err != nil {
		return err
	}
	if len(a.Mailbox) > MaxActorRuntimeMailboxMessages {
		return fmt.Errorf("actor runtime mailbox messages must be <= %d", MaxActorRuntimeMailboxMessages)
	}
	seenSequences := make(map[uint64]struct{}, len(a.Mailbox))
	for i, msg := range a.Mailbox {
		if err := msg.Validate(a.ActorID); err != nil {
			return err
		}
		if _, found := seenSequences[msg.Sequence]; found {
			return fmt.Errorf("duplicate actor runtime mailbox sequence %d", msg.Sequence)
		}
		seenSequences[msg.Sequence] = struct{}{}
		if i > 0 && compareActorMailboxMessages(a.Mailbox[i-1], msg) >= 0 {
			return errors.New("actor runtime mailbox must be sorted canonically")
		}
	}
	return nil
}

func (m ActorMailboxMessage) Validate(actorID string) error {
	if m.Sequence == 0 {
		return errors.New("actor runtime mailbox sequence must be positive")
	}
	if err := validateEngineToken("actor runtime mailbox source actor", m.SourceActor, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("actor runtime mailbox target actor", m.TargetActor, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	if m.TargetActor != actorID {
		return errors.New("actor runtime mailbox target must match actor")
	}
	if m.CreatedLogicalTime == 0 {
		return errors.New("actor runtime mailbox logical time must be positive")
	}
	if m.Envelope.GasLimit == 0 {
		return errors.New("actor runtime mailbox message gas limit must be positive")
	}
	return nil
}

func (e ActorExecution) Validate(actors map[string]ActorRuntimeActor) error {
	if _, found := actors[e.ActorID]; !found {
		return fmt.Errorf("actor runtime execution actor %q is not declared", e.ActorID)
	}
	if e.MessageSequence == 0 {
		return errors.New("actor runtime execution message sequence must be positive")
	}
	if err := validateEngineToken("actor runtime handler", e.Handler, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	if e.GasLimit == 0 {
		return errors.New("actor runtime execution gas limit must be positive")
	}
	if e.GasUsed > e.GasLimit {
		return errors.New("actor runtime execution gas used exceeds limit")
	}
	if len(e.StateWrites) > MaxActorRuntimeStateWrites {
		return fmt.Errorf("actor runtime state writes must be <= %d", MaxActorRuntimeStateWrites)
	}
	for i, write := range e.StateWrites {
		if err := write.Validate(e.ActorID); err != nil {
			return err
		}
		if i > 0 && compareActorStateWrites(e.StateWrites[i-1], write) >= 0 {
			return errors.New("actor runtime state writes must be sorted canonically")
		}
	}
	if len(e.EmittedMessages) > MaxActorRuntimeEmittedMessages {
		return fmt.Errorf("actor runtime emitted messages must be <= %d", MaxActorRuntimeEmittedMessages)
	}
	for i, msg := range e.EmittedMessages {
		if msg.GasLimit == 0 {
			return errors.New("actor runtime emitted message gas limit must be positive")
		}
		if i > 0 && compareAsyncMessages(e.EmittedMessages[i-1], msg) >= 0 {
			return errors.New("actor runtime emitted messages must be sorted canonically")
		}
	}
	if e.Error != "" && len(e.StateWrites) > 0 {
		return errors.New("failed actor execution must not commit state writes")
	}
	return nil
}

func (w ActorStateWrite) Validate(actorID string) error {
	if w.ActorID != actorID {
		return errors.New("actor runtime state write cannot target another actor")
	}
	if err := validateEngineToken("actor runtime state write actor id", w.ActorID, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	expectedPrefix := ActorStateKeyPrefix(actorID)
	if !strings.HasPrefix(w.Key, expectedPrefix) {
		return fmt.Errorf("actor runtime state write key must use prefix %q", expectedPrefix)
	}
	if err := validateRouterOptionalToken("actor runtime state write key", w.Key, MaxRouterTargetLength*2); err != nil {
		return err
	}
	return zonestypes.ValidateHash("actor runtime state write hash", w.Hash)
}

func (c ContinuationRecord) Validate(actors map[string]ActorRuntimeActor, height uint64) error {
	if err := validateEngineToken("continuation id", c.ContinuationID, MaxContinuationTokenLength); err != nil {
		return err
	}
	if _, found := actors[c.ActorID]; !found {
		return fmt.Errorf("continuation actor %q is not declared", c.ActorID)
	}
	if c.StepIndex == 0 {
		return errors.New("continuation step index must be positive")
	}
	if err := zonestypes.ValidateHash("continuation partial state hash", c.PartialStateHash); err != nil {
		return err
	}
	if c.PartialStateBytes == 0 || c.PartialStateBytes > MaxContinuationPartialStateBytes {
		return fmt.Errorf("continuation partial state must be bounded to 1..%d bytes", MaxContinuationPartialStateBytes)
	}
	if c.ResumeHeight == 0 {
		return errors.New("continuation resume height must be explicit")
	}
	if c.ExpiryHeight == 0 {
		return errors.New("continuation expiry height must be explicit")
	}
	if c.ExpiryHeight < c.ResumeHeight {
		return errors.New("continuation expiry must not precede resume height")
	}
	if c.GasReserved == 0 {
		return errors.New("continuation gas reserved must be positive")
	}
	if !IsContinuationStatus(c.Status) {
		return fmt.Errorf("invalid continuation status %q", c.Status)
	}
	if c.ResumeBy != "" && c.ResumeBy != ContinuationResumeByScheduler {
		return errors.New("continuation can resume only through scheduler")
	}
	if c.Status == ContinuationStatusResumed && c.ResumeBy != ContinuationResumeByScheduler {
		return errors.New("resumed continuation must be resumed by scheduler")
	}
	expiredAtHeight := height >= c.ExpiryHeight
	if expiredAtHeight && c.Status != ContinuationStatusExpired && c.Status != ContinuationStatusResumed {
		return errors.New("expired continuation must emit failure receipt")
	}
	if c.Status == ContinuationStatusExpired {
		return validateExpiredContinuationReceipt(c.FailureReceipt)
	}
	if c.FailureReceipt.Sequence != 0 {
		return errors.New("non-expired continuation must not carry failure receipt")
	}
	return nil
}

func ActorStateKeyPrefix(actorID string) string {
	return fmt.Sprintf("actor/%s/", strings.TrimSpace(actorID))
}

func IsContinuationStatus(status ContinuationStatus) bool {
	switch status {
	case ContinuationStatusPending,
		ContinuationStatusScheduled,
		ContinuationStatusResumed,
		ContinuationStatusExpired,
		ContinuationStatusFailed:
		return true
	default:
		return false
	}
}

func ComputeActorRuntimePlanRoot(plan ActorRuntimePlan) string {
	plan = canonicalActorRuntimePlan(plan)
	h := sha256.New()
	writeEnginePart(h, "aetra-actor-runtime-plan-v1")
	writeEngineUint64(h, plan.Height)
	writeEngineUint64(h, uint64(len(plan.Actors)))
	for _, actor := range plan.Actors {
		writeEnginePart(h, actor.ActorID)
		writeEnginePart(h, actor.CodeRef)
		writeEnginePart(h, actor.StateRoot)
		writeEngineUint64(h, uint64(len(actor.Mailbox)))
		for _, msg := range actor.Mailbox {
			writeActorMailboxMessage(h, msg)
		}
	}
	writeEngineUint64(h, uint64(len(plan.Executions)))
	for _, execution := range plan.Executions {
		writeEnginePart(h, execution.ActorID)
		writeEngineUint64(h, execution.MessageSequence)
		writeEnginePart(h, execution.Handler)
		writeEngineUint64(h, execution.GasLimit)
		writeEngineUint64(h, execution.GasUsed)
		writeEngineUint64(h, uint64(len(execution.StateWrites)))
		for _, write := range execution.StateWrites {
			writeEnginePart(h, write.ActorID)
			writeEnginePart(h, write.Key)
			writeEnginePart(h, write.Hash)
		}
		writeEngineUint64(h, uint64(len(execution.EmittedMessages)))
		for _, msg := range execution.EmittedMessages {
			writeAsyncMessage(h, msg)
		}
		writeEngineUint64(h, uint64(execution.ResultCode))
		writeEnginePart(h, execution.Error)
	}
	writeEngineUint64(h, uint64(len(plan.Continuations)))
	for _, continuation := range plan.Continuations {
		writeContinuationRecord(h, continuation)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalActorRuntimePlan(plan ActorRuntimePlan) ActorRuntimePlan {
	out := ActorRuntimePlan{
		Height:		plan.Height,
		Actors:		cloneActorRuntimeActors(plan.Actors),
		Executions:	cloneActorExecutions(plan.Executions),
		Continuations:	append([]ContinuationRecord(nil), plan.Continuations...),
		PlanRoot:	strings.TrimSpace(plan.PlanRoot),
	}
	sort.SliceStable(out.Actors, func(i, j int) bool {
		return out.Actors[i].ActorID < out.Actors[j].ActorID
	})
	sort.SliceStable(out.Executions, func(i, j int) bool {
		return compareActorExecutions(out.Executions[i], out.Executions[j]) < 0
	})
	sort.SliceStable(out.Continuations, func(i, j int) bool {
		return compareContinuationRecords(out.Continuations[i], out.Continuations[j]) < 0
	})
	return out
}

func validateActorExecutions(executions []ActorExecution, actors map[string]ActorRuntimeActor) error {
	seen := make(map[string]struct{}, len(executions))
	mailboxSequences := make(map[string]map[uint64]struct{}, len(actors))
	for actorID, actor := range actors {
		mailboxSequences[actorID] = make(map[uint64]struct{}, len(actor.Mailbox))
		for _, msg := range actor.Mailbox {
			mailboxSequences[actorID][msg.Sequence] = struct{}{}
		}
	}
	for i, execution := range executions {
		if err := execution.Validate(actors); err != nil {
			return err
		}
		if _, found := mailboxSequences[execution.ActorID][execution.MessageSequence]; !found {
			return errors.New("actor runtime execution message must be declared in mailbox")
		}
		key := fmt.Sprintf("%s/%020d", execution.ActorID, execution.MessageSequence)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate actor runtime execution %s", key)
		}
		seen[key] = struct{}{}
		if i > 0 && compareActorExecutions(executions[i-1], execution) >= 0 {
			return errors.New("actor runtime executions must be sorted canonically")
		}
	}
	return nil
}

func validateContinuationRecords(records []ContinuationRecord, actors map[string]ActorRuntimeActor, height uint64) error {
	seen := make(map[string]struct{}, len(records))
	perActor := make(map[string]uint32, len(records))
	for i, record := range records {
		if err := record.Validate(actors, height); err != nil {
			return err
		}
		if _, found := seen[record.ContinuationID]; found {
			return fmt.Errorf("duplicate continuation %q", record.ContinuationID)
		}
		seen[record.ContinuationID] = struct{}{}
		perActor[record.ActorID]++
		if perActor[record.ActorID] > MaxContinuationRecordsPerActorBlock {
			return fmt.Errorf("continuations per actor per block must be <= %d", MaxContinuationRecordsPerActorBlock)
		}
		if i > 0 && compareContinuationRecords(records[i-1], record) >= 0 {
			return errors.New("continuations must be sorted canonically")
		}
	}
	return nil
}

func validateExpiredContinuationReceipt(receipt async.ExecutionReceipt) error {
	if receipt.Sequence == 0 {
		return errors.New("expired continuation failure receipt sequence must be positive")
	}
	if receipt.ResultCode == async.ResultOK {
		return errors.New("expired continuation failure receipt must not be successful")
	}
	if receipt.GasUsed == 0 {
		return errors.New("expired continuation failure receipt gas used must be positive")
	}
	if strings.TrimSpace(receipt.Error) == "" {
		return errors.New("expired continuation failure receipt error is required")
	}
	return nil
}

func cloneActorRuntimeActors(actors []ActorRuntimeActor) []ActorRuntimeActor {
	out := make([]ActorRuntimeActor, len(actors))
	for i, actor := range actors {
		out[i] = actor
		out[i].ActorID = strings.TrimSpace(actor.ActorID)
		out[i].CodeRef = strings.TrimSpace(actor.CodeRef)
		out[i].StateRoot = strings.TrimSpace(actor.StateRoot)
		out[i].Mailbox = append([]ActorMailboxMessage(nil), actor.Mailbox...)
		sort.SliceStable(out[i].Mailbox, func(left, right int) bool {
			return compareActorMailboxMessages(out[i].Mailbox[left], out[i].Mailbox[right]) < 0
		})
	}
	return out
}

func cloneActorExecutions(executions []ActorExecution) []ActorExecution {
	out := make([]ActorExecution, len(executions))
	for i, execution := range executions {
		out[i] = execution
		out[i].ActorID = strings.TrimSpace(execution.ActorID)
		out[i].Handler = strings.TrimSpace(execution.Handler)
		out[i].Error = strings.TrimSpace(execution.Error)
		out[i].StateWrites = append([]ActorStateWrite(nil), execution.StateWrites...)
		out[i].EmittedMessages = append([]async.MessageEnvelope(nil), execution.EmittedMessages...)
		sort.SliceStable(out[i].StateWrites, func(left, right int) bool {
			return compareActorStateWrites(out[i].StateWrites[left], out[i].StateWrites[right]) < 0
		})
		sort.SliceStable(out[i].EmittedMessages, func(left, right int) bool {
			return compareAsyncMessages(out[i].EmittedMessages[left], out[i].EmittedMessages[right]) < 0
		})
	}
	return out
}

func compareActorMailboxMessages(left, right ActorMailboxMessage) int {
	if left.Sequence < right.Sequence {
		return -1
	}
	if left.Sequence > right.Sequence {
		return 1
	}
	if left.CreatedLogicalTime < right.CreatedLogicalTime {
		return -1
	}
	if left.CreatedLogicalTime > right.CreatedLogicalTime {
		return 1
	}
	if left.SourceActor < right.SourceActor {
		return -1
	}
	if left.SourceActor > right.SourceActor {
		return 1
	}
	return 0
}

func compareActorExecutions(left, right ActorExecution) int {
	if left.ActorID < right.ActorID {
		return -1
	}
	if left.ActorID > right.ActorID {
		return 1
	}
	if left.MessageSequence < right.MessageSequence {
		return -1
	}
	if left.MessageSequence > right.MessageSequence {
		return 1
	}
	return 0
}

func compareActorStateWrites(left, right ActorStateWrite) int {
	if left.ActorID < right.ActorID {
		return -1
	}
	if left.ActorID > right.ActorID {
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

func compareContinuationRecords(left, right ContinuationRecord) int {
	if left.ActorID < right.ActorID {
		return -1
	}
	if left.ActorID > right.ActorID {
		return 1
	}
	if left.ResumeHeight < right.ResumeHeight {
		return -1
	}
	if left.ResumeHeight > right.ResumeHeight {
		return 1
	}
	if left.ContinuationID < right.ContinuationID {
		return -1
	}
	if left.ContinuationID > right.ContinuationID {
		return 1
	}
	return 0
}

func writeActorMailboxMessage(w engineByteWriter, msg ActorMailboxMessage) {
	writeEngineUint64(w, msg.Sequence)
	writeEnginePart(w, msg.SourceActor)
	writeEnginePart(w, msg.TargetActor)
	writeEngineUint64(w, msg.CreatedLogicalTime)
	writeAsyncMessage(w, msg.Envelope)
}

func writeContinuationRecord(w engineByteWriter, continuation ContinuationRecord) {
	writeEnginePart(w, continuation.ContinuationID)
	writeEnginePart(w, continuation.ActorID)
	writeEngineUint64(w, uint64(continuation.StepIndex))
	writeEnginePart(w, continuation.PartialStateHash)
	writeEngineUint64(w, uint64(continuation.PartialStateBytes))
	writeEngineUint64(w, continuation.ResumeHeight)
	writeEngineUint64(w, continuation.ExpiryHeight)
	writeEngineUint64(w, continuation.GasReserved)
	writeEnginePart(w, string(continuation.Status))
	writeEnginePart(w, continuation.ResumeBy)
	writeAsyncReceipt(w, continuation.FailureReceipt)
}
