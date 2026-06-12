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
	AVMAsyncFailureInvalidPayload		AVMAsyncFailureClass	= "invalid_payload"
	AVMAsyncFailureInsufficientGas		AVMAsyncFailureClass	= "insufficient_gas"
	AVMAsyncFailureDestinationNotFound	AVMAsyncFailureClass	= "destination_not_found"
	AVMAsyncFailureDestinationDisabled	AVMAsyncFailureClass	= "destination_disabled"
	AVMAsyncFailureExpiredMessage		AVMAsyncFailureClass	= "expired_message"
	AVMAsyncFailureHandlerFailure		AVMAsyncFailureClass	= "handler_failure"
	AVMAsyncFailureStorageLimitExceeded	AVMAsyncFailureClass	= "storage_limit_exceeded"
	AVMAsyncFailureProofVerificationFailure	AVMAsyncFailureClass	= "proof_verification_failure"
	AVMAsyncFailureRetryExhausted		AVMAsyncFailureClass	= "retry_exhausted"

	AVMAsyncErrorInvalidPayload		AVMAsyncErrorCode	= "ERR_INVALID_PAYLOAD"
	AVMAsyncErrorInsufficientGas		AVMAsyncErrorCode	= "ERR_INSUFFICIENT_GAS"
	AVMAsyncErrorDestinationNotFound	AVMAsyncErrorCode	= "ERR_DESTINATION_NOT_FOUND"
	AVMAsyncErrorDestinationDisabled	AVMAsyncErrorCode	= "ERR_DESTINATION_DISABLED"
	AVMAsyncErrorExpiredMessage		AVMAsyncErrorCode	= "ERR_EXPIRED_MESSAGE"
	AVMAsyncErrorHandlerFailure		AVMAsyncErrorCode	= "ERR_HANDLER_FAILURE"
	AVMAsyncErrorStorageLimitExceeded	AVMAsyncErrorCode	= "ERR_STORAGE_LIMIT_EXCEEDED"
	AVMAsyncErrorProofVerificationFailure	AVMAsyncErrorCode	= "ERR_PROOF_VERIFICATION_FAILURE"
	AVMAsyncErrorRetryExhausted		AVMAsyncErrorCode	= "ERR_RETRY_EXHAUSTED"
	AVMAsyncErrorBounceFailed		AVMAsyncErrorCode	= "ERR_BOUNCE_FAILED"

	AVMDeadLetterTriggerRetryExhausted			AVMDeadLetterTrigger	= "retry_exhausted"
	AVMDeadLetterTriggerDestinationPermanentlyDisabled	AVMDeadLetterTrigger	= "destination_permanently_disabled"
	AVMDeadLetterTriggerBounceFailed			AVMDeadLetterTrigger	= "bounce_failed"
	AVMDeadLetterTriggerExpiredBeforeExecution		AVMDeadLetterTrigger	= "expired_before_execution"
	AVMDeadLetterTriggerInvalidMessageFormat		AVMDeadLetterTrigger	= "invalid_message_format"

	AVMAsyncBouncePayloadType	= "async.bounce"

	MaxAVMBouncePayloadBytes	= 1024
	MaxAVMFailureErrorCodeLength	= MaxAVMReceiptErrorCode
)

type AVMAsyncFailureClass string
type AVMAsyncErrorCode string
type AVMDeadLetterTrigger string

type AVMAsyncFailureRecord struct {
	MessageID	string
	ZoneID		zonestypes.ZoneID
	FailureClass	AVMAsyncFailureClass
	ErrorCode	string
	FailedHeight	uint64
	Attempt		uint32
	GasUsed		uint64
	RetryExhausted	bool
	FailureHash	string
}

type AVMAsyncBounceOutcome struct {
	OriginalMessage		AVMAsyncMessage
	Failure			AVMAsyncFailureRecord
	BounceMessage		AVMAsyncMessage
	BounceReceipt		AVMExecutionReceipt
	RemainingValueNAET	uint64
	BounceGasUsed		uint64
	BoundedBounceGas	uint64
	OutcomeHash		string
}

type AVMAsyncBounceDeadLetterOutcome struct {
	BounceOutcome		AVMAsyncBounceOutcome
	DeadLetter		AVMDeadLetterRecord
	DeadLetterReceipt	AVMExecutionReceipt
	OutcomeHash		string
}

type AVMAsyncFailedReceiptModel struct {
	Message		AVMAsyncMessage
	Failure		AVMAsyncFailureRecord
	Receipt		AVMExecutionReceipt
	ModelHash	string
}

type AVMZoneDeadLetterQueue struct {
	ZoneID		zonestypes.ZoneID
	Records		[]AVMDeadLetterRecord
	QueueRoot	string
}

type AVMRetryExhaustionOutcome struct {
	Queue		AVMZoneDeadLetterQueue
	Message		AVMAsyncMessage
	Failure		AVMAsyncFailureRecord
	Receipt		AVMExecutionReceipt
	DeadLetter	AVMDeadLetterRecord
	OutcomeHash	string
}

type AVMAsyncFailureValueConservationCheck struct {
	OriginalMessageID	string
	OriginalValueNAET	uint64
	BounceValueNAET		uint64
	DeadLetterRefundNAET	uint64
	ConservationCheckHash	string
}

func NewAVMAsyncFailureRecord(record AVMAsyncFailureRecord) (AVMAsyncFailureRecord, error) {
	record = canonicalAVMAsyncFailureRecord(record)
	if record.ErrorCode == "" {
		if code, ok := AVMAsyncFailureErrorCode(record.FailureClass); ok {
			record.ErrorCode = string(code)
		}
	}
	record.FailureHash = ComputeAVMAsyncFailureHash(record)
	return record, record.Validate()
}

func (r AVMAsyncFailureRecord) Validate() error {
	r = canonicalAVMAsyncFailureRecord(r)
	if err := zonestypes.ValidateHash("AVM async failure message id", r.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if !IsAVMAsyncFailureClass(r.FailureClass) {
		return fmt.Errorf("invalid AVM async failure class %q", r.FailureClass)
	}
	if err := validateRouterOptionalToken("AVM async failure error code", r.ErrorCode, MaxAVMFailureErrorCodeLength); err != nil {
		return err
	}
	if r.ErrorCode == "" {
		return errors.New("AVM async failure error code is required")
	}
	if !IsAVMAsyncErrorCode(AVMAsyncErrorCode(r.ErrorCode)) {
		return fmt.Errorf("invalid AVM async error code %q", r.ErrorCode)
	}
	if r.FailedHeight == 0 {
		return errors.New("AVM async failure height must be positive")
	}
	if r.Attempt == 0 {
		return errors.New("AVM async failure attempt must be positive")
	}
	if r.FailureClass == AVMAsyncFailureRetryExhausted && !r.RetryExhausted {
		return errors.New("AVM async retry exhausted failure must set retry exhausted flag")
	}
	if r.RetryExhausted && r.FailureClass != AVMAsyncFailureRetryExhausted {
		return errors.New("AVM async retry exhausted flag requires retry_exhausted failure class")
	}
	if r.FailureHash == "" {
		return errors.New("AVM async failure hash is required")
	}
	if err := zonestypes.ValidateHash("AVM async failure hash", r.FailureHash); err != nil {
		return err
	}
	if r.FailureHash != ComputeAVMAsyncFailureHash(r) {
		return errors.New("AVM async failure hash mismatch")
	}
	return nil
}

func NewAVMAsyncFailedReceiptModel(message AVMAsyncMessage, failure AVMAsyncFailureRecord, executor string, storageWritten uint32, outputMessagesRoot string, createdHeight uint64) (AVMAsyncFailedReceiptModel, error) {
	message = canonicalAVMAsyncMessage(message)
	if err := message.Validate(); err != nil {
		return AVMAsyncFailedReceiptModel{}, err
	}
	failure = canonicalAVMAsyncFailureRecord(failure)
	if err := failure.Validate(); err != nil {
		return AVMAsyncFailedReceiptModel{}, err
	}
	if failure.MessageID != message.ID {
		return AVMAsyncFailedReceiptModel{}, errors.New("AVM async failed receipt failure message id mismatch")
	}
	if failure.ZoneID != message.DestinationZone {
		return AVMAsyncFailedReceiptModel{}, errors.New("AVM async failed receipt failure zone mismatch")
	}
	if outputMessagesRoot == "" {
		outputMessagesRoot = ComputeAVMAsyncFailureOutputRoot(failure)
	}
	status := AVMReceiptStatusFailed
	if failure.FailureClass == AVMAsyncFailureExpiredMessage {
		status = AVMReceiptStatusExpired
	}
	gasUsed := failure.GasUsed
	if gasUsed == 0 {
		gasUsed = 1
	}
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		message.ID,
		ZoneID:			message.DestinationZone,
		Executor:		executor,
		Status:			status,
		GasUsed:		gasUsed,
		StorageWritten:		storageWritten,
		EventsHash:		failure.FailureHash,
		OutputMessagesRoot:	outputMessagesRoot,
		ErrorCodeOptional:	failure.ErrorCode,
		CreatedHeight:		createdHeight,
	})
	if err != nil {
		return AVMAsyncFailedReceiptModel{}, err
	}
	model := AVMAsyncFailedReceiptModel{
		Message:	message,
		Failure:	failure,
		Receipt:	receipt,
	}
	model = canonicalAVMAsyncFailedReceiptModel(model)
	model.ModelHash = ComputeAVMAsyncFailedReceiptModelHash(model)
	return model, model.Validate()
}

func (m AVMAsyncFailedReceiptModel) Validate() error {
	m = canonicalAVMAsyncFailedReceiptModel(m)
	if err := m.Message.Validate(); err != nil {
		return err
	}
	if err := m.Failure.Validate(); err != nil {
		return err
	}
	if err := m.Receipt.Validate(); err != nil {
		return err
	}
	if m.Failure.MessageID != m.Message.ID || m.Receipt.MessageID != m.Message.ID {
		return errors.New("AVM async failed receipt message id mismatch")
	}
	if m.Failure.ZoneID != m.Message.DestinationZone || m.Receipt.ZoneID != m.Message.DestinationZone {
		return errors.New("AVM async failed receipt zone mismatch")
	}
	if m.Receipt.Status != AVMReceiptStatusFailed && m.Receipt.Status != AVMReceiptStatusExpired {
		return errors.New("AVM async failed receipt must use failed or expired status")
	}
	if m.Failure.FailureClass == AVMAsyncFailureExpiredMessage && m.Receipt.Status != AVMReceiptStatusExpired {
		return errors.New("AVM async expired failure must emit expired receipt")
	}
	if m.Receipt.ErrorCodeOptional != m.Failure.ErrorCode {
		return errors.New("AVM async failed receipt error code mismatch")
	}
	if m.Receipt.EventsHash != m.Failure.FailureHash {
		return errors.New("AVM async failed receipt events hash must commit failure")
	}
	if m.ModelHash == "" {
		return errors.New("AVM async failed receipt model hash is required")
	}
	if err := zonestypes.ValidateHash("AVM async failed receipt model hash", m.ModelHash); err != nil {
		return err
	}
	if m.ModelHash != ComputeAVMAsyncFailedReceiptModelHash(m) {
		return errors.New("AVM async failed receipt model hash mismatch")
	}
	return nil
}

func NewAVMAsyncBounceOutcome(original AVMAsyncMessage, failure AVMAsyncFailureRecord, schedule AVMGasSchedule, remainingValueNAET uint64, bounceSenderNonce uint64, createdHeight uint64) (AVMAsyncBounceOutcome, error) {
	original = canonicalAVMAsyncMessage(original)
	if err := original.Validate(); err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	failure = canonicalAVMAsyncFailureRecord(failure)
	if err := failure.Validate(); err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	schedule = canonicalAVMGasSchedule(schedule)
	if err := schedule.Validate(); err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	if bounceSenderNonce == 0 {
		return AVMAsyncBounceOutcome{}, errors.New("AVM async bounce sender nonce must be positive")
	}
	if createdHeight == 0 {
		return AVMAsyncBounceOutcome{}, errors.New("AVM async bounce created height must be positive")
	}
	if !original.BounceFlag {
		return AVMAsyncBounceOutcome{}, errors.New("AVM async bounce requires original bounce flag")
	}
	if failure.MessageID != original.ID {
		return AVMAsyncBounceOutcome{}, errors.New("AVM async bounce failure message id mismatch")
	}
	if failure.ZoneID != original.DestinationZone {
		return AVMAsyncBounceOutcome{}, errors.New("AVM async bounce failure zone must match original destination zone")
	}
	if remainingValueNAET > original.ValueNAET {
		return AVMAsyncBounceOutcome{}, errors.New("AVM async bounce cannot create more value than original remaining value")
	}
	payload, err := AVMAsyncBouncePayload(original.ID, failure.ErrorCode)
	if err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	expiryHeight := original.ExpiryHeight
	if expiryHeight <= createdHeight {
		expiryHeight = createdHeight + 1
	}
	bounce, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:			original.ChainID,
		Source:				original.Destination,
		Destination:			original.Source,
		Payload:			payload,
		GasLimit:			schedule.BounceGas,
		ExpiryHeight:			expiryHeight,
		RetryPolicy:			AVMRetryPolicy{Mode: AVMRetryModeNone, BackoffMode: AVMBackoffModeNone},
		BounceFlag:			false,
		SourceZone:			original.DestinationZone,
		DestinationZone:		original.SourceZone,
		SourceActorOptional:		original.DestinationActorOptional,
		DestinationActorOptional:	original.SourceActorOptional,
		SenderNonce:			bounceSenderNonce,
		PayloadType:			AVMAsyncBouncePayloadType,
		ValueNAET:			remainingValueNAET,
		ForwardingFee:			original.ForwardingFee,
		Priority:			original.Priority,
		CreatedHeight:			createdHeight,
		RouteHintOptional:		AVMAsyncBounceRouteHint(original.ID),
	})
	if err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		original.ID,
		ZoneID:			original.DestinationZone,
		Executor:		"async-bounce",
		Status:			AVMReceiptStatusBounced,
		GasUsed:		schedule.BounceGas,
		EventsHash:		failure.FailureHash,
		OutputMessagesRoot:	ComputeAVMAsyncBounceMessageRoot(bounce),
		ErrorCodeOptional:	failure.ErrorCode,
		CreatedHeight:		createdHeight,
	})
	if err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	outcome := AVMAsyncBounceOutcome{
		OriginalMessage:	original,
		Failure:		failure,
		BounceMessage:		bounce,
		BounceReceipt:		receipt,
		RemainingValueNAET:	remainingValueNAET,
		BounceGasUsed:		schedule.BounceGas,
		BoundedBounceGas:	schedule.BounceGas,
	}
	outcome = canonicalAVMAsyncBounceOutcome(outcome)
	outcome.OutcomeHash = ComputeAVMAsyncBounceOutcomeHash(outcome)
	return outcome, outcome.Validate()
}

func (o AVMAsyncBounceOutcome) Validate() error {
	o = canonicalAVMAsyncBounceOutcome(o)
	if err := o.OriginalMessage.Validate(); err != nil {
		return fmt.Errorf("AVM async bounce original: %w", err)
	}
	if err := o.Failure.Validate(); err != nil {
		return fmt.Errorf("AVM async bounce failure: %w", err)
	}
	if err := o.BounceMessage.Validate(); err != nil {
		return fmt.Errorf("AVM async bounce message: %w", err)
	}
	if err := o.BounceReceipt.Validate(); err != nil {
		return fmt.Errorf("AVM async bounce receipt: %w", err)
	}
	if !o.OriginalMessage.BounceFlag {
		return errors.New("AVM async bounce requires original bounce flag")
	}
	if o.Failure.MessageID != o.OriginalMessage.ID {
		return errors.New("AVM async bounce failure message id mismatch")
	}
	if o.Failure.ZoneID != o.OriginalMessage.DestinationZone {
		return errors.New("AVM async bounce failure zone must match original destination zone")
	}
	if o.RemainingValueNAET > o.OriginalMessage.ValueNAET {
		return errors.New("AVM async bounce cannot create more value than original remaining value")
	}
	if o.BounceMessage.Source != o.OriginalMessage.Destination ||
		o.BounceMessage.Destination != o.OriginalMessage.Source ||
		o.BounceMessage.SourceZone != o.OriginalMessage.DestinationZone ||
		o.BounceMessage.DestinationZone != o.OriginalMessage.SourceZone {
		return errors.New("AVM async bounce message must reverse original source and destination")
	}
	if o.BounceMessage.ValueNAET != o.RemainingValueNAET {
		return errors.New("AVM async bounce message value must equal remaining value")
	}
	if o.BounceMessage.PayloadType != AVMAsyncBouncePayloadType {
		return errors.New("AVM async bounce message payload type mismatch")
	}
	if len(o.BounceMessage.Payload) > MaxAVMBouncePayloadBytes {
		return fmt.Errorf("AVM async bounce payload must be <= %d bytes", MaxAVMBouncePayloadBytes)
	}
	payload := string(o.BounceMessage.Payload)
	if !strings.Contains(payload, o.OriginalMessage.ID) || !strings.Contains(payload, o.Failure.ErrorCode) {
		return errors.New("AVM async bounce payload must include original message id and error code")
	}
	if !strings.Contains(o.BounceMessage.RouteHintOptional, o.OriginalMessage.ID) {
		return errors.New("AVM async bounce route hint must reference original message id")
	}
	if o.BoundedBounceGas == 0 {
		return errors.New("AVM async bounded bounce gas must be positive")
	}
	if o.BounceGasUsed == 0 {
		return errors.New("AVM async bounce gas used must be positive")
	}
	if o.BounceGasUsed > o.BoundedBounceGas {
		return errors.New("AVM async bounce gas used exceeds bounded bounce gas")
	}
	if o.BounceMessage.GasLimit != o.BoundedBounceGas {
		return errors.New("AVM async bounce message gas must equal bounded bounce gas")
	}
	if o.BounceReceipt.Status != AVMReceiptStatusBounced {
		return errors.New("AVM async bounce receipt must use bounced status")
	}
	if o.BounceReceipt.MessageID != o.OriginalMessage.ID {
		return errors.New("AVM async bounce receipt must reference original message")
	}
	if o.BounceReceipt.ZoneID != o.OriginalMessage.DestinationZone {
		return errors.New("AVM async bounce receipt zone must match original destination zone")
	}
	if o.BounceReceipt.GasUsed != o.BounceGasUsed {
		return errors.New("AVM async bounce receipt gas must match bounce gas used")
	}
	if o.BounceReceipt.OutputMessagesRoot != ComputeAVMAsyncBounceMessageRoot(o.BounceMessage) {
		return errors.New("AVM async bounce receipt output message root mismatch")
	}
	if o.OutcomeHash == "" {
		return errors.New("AVM async bounce outcome hash is required")
	}
	if err := zonestypes.ValidateHash("AVM async bounce outcome hash", o.OutcomeHash); err != nil {
		return err
	}
	if o.OutcomeHash != ComputeAVMAsyncBounceOutcomeHash(o) {
		return errors.New("AVM async bounce outcome hash mismatch")
	}
	return nil
}

func NewAVMAsyncBounceDeadLetterOutcome(outcome AVMAsyncBounceOutcome, reason string, failedAttempts uint32, finalHeight uint64) (AVMAsyncBounceDeadLetterOutcome, error) {
	outcome = canonicalAVMAsyncBounceOutcome(outcome)
	if err := outcome.Validate(); err != nil {
		return AVMAsyncBounceDeadLetterOutcome{}, err
	}
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		outcome.BounceMessage.ID,
		ZoneID:			outcome.BounceMessage.DestinationZone,
		Executor:		"async-bounce-dead-letter",
		Status:			AVMReceiptStatusDeadLettered,
		GasUsed:		outcome.BounceGasUsed,
		EventsHash:		outcome.Failure.FailureHash,
		OutputMessagesRoot:	ComputeAVMAsyncBounceDeadLetterOutputRoot(outcome.BounceMessage.ID),
		ErrorCodeOptional:	outcome.Failure.ErrorCode,
		CreatedHeight:		finalHeight,
	})
	if err != nil {
		return AVMAsyncBounceDeadLetterOutcome{}, err
	}
	dead, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:		outcome.BounceMessage.ID,
		ZoneID:			outcome.BounceMessage.DestinationZone,
		Reason:			reason,
		FailedAttempts:		failedAttempts,
		LastErrorCode:		outcome.Failure.ErrorCode,
		FinalHeight:		finalHeight,
		RefundAmountOptional:	outcome.RemainingValueNAET,
		ReceiptID:		receipt.ReceiptID,
	})
	if err != nil {
		return AVMAsyncBounceDeadLetterOutcome{}, err
	}
	next := AVMAsyncBounceDeadLetterOutcome{
		BounceOutcome:		outcome,
		DeadLetter:		dead,
		DeadLetterReceipt:	receipt,
	}
	next = canonicalAVMAsyncBounceDeadLetterOutcome(next)
	next.OutcomeHash = ComputeAVMAsyncBounceDeadLetterOutcomeHash(next)
	return next, next.Validate()
}

func NewAVMZoneDeadLetterQueue(queue AVMZoneDeadLetterQueue) (AVMZoneDeadLetterQueue, error) {
	queue = canonicalAVMZoneDeadLetterQueue(queue)
	queue.QueueRoot = ComputeAVMZoneDeadLetterQueueRoot(queue)
	return queue, queue.Validate()
}

func (q AVMZoneDeadLetterQueue) Validate() error {
	q = canonicalAVMZoneDeadLetterQueue(q)
	if err := zonestypes.ValidateZoneID(q.ZoneID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(q.Records))
	var previous string
	for i, record := range q.Records {
		if err := record.Validate(); err != nil {
			return fmt.Errorf("AVM zone dead letter record %d: %w", i, err)
		}
		if record.ZoneID != q.ZoneID {
			return errors.New("AVM zone dead letter record zone mismatch")
		}
		sortKey := AVMZoneDeadLetterSortKey(record)
		if i > 0 && previous >= sortKey {
			return errors.New("AVM zone dead letter queue records must be canonically ordered")
		}
		previous = sortKey
		if _, found := seen[record.MessageID]; found {
			return fmt.Errorf("duplicate AVM zone dead letter message id %q", record.MessageID)
		}
		seen[record.MessageID] = struct{}{}
	}
	if q.QueueRoot == "" {
		return errors.New("AVM zone dead letter queue root is required")
	}
	if err := zonestypes.ValidateHash("AVM zone dead letter queue root", q.QueueRoot); err != nil {
		return err
	}
	if q.QueueRoot != ComputeAVMZoneDeadLetterQueueRoot(q) {
		return errors.New("AVM zone dead letter queue root mismatch")
	}
	return nil
}

func DeadLetterAVMAsyncFailure(queue AVMZoneDeadLetterQueue, failure AVMAsyncFailureRecord, trigger AVMDeadLetterTrigger, reason string, refundAmountNAET uint64, gasUsed uint64, finalHeight uint64) (AVMZoneDeadLetterQueue, AVMDeadLetterRecord, AVMExecutionReceipt, error) {
	queue = canonicalAVMZoneDeadLetterQueue(queue)
	if err := queue.Validate(); err != nil {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, err
	}
	failure = canonicalAVMAsyncFailureRecord(failure)
	if err := failure.Validate(); err != nil {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, err
	}
	if failure.ZoneID != queue.ZoneID {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, errors.New("AVM dead letter failure zone mismatch")
	}
	if !IsAVMDeadLetterTrigger(trigger) {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, fmt.Errorf("invalid AVM dead letter trigger %q", trigger)
	}
	if err := ValidateAVMDeadLetterTriggerForFailure(trigger, failure); err != nil {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, err
	}
	if gasUsed == 0 {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, errors.New("AVM dead letter receipt gas used must be positive")
	}
	if finalHeight == 0 {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, errors.New("AVM dead letter final height must be positive")
	}
	if reason == "" {
		reason = string(trigger)
	}
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		failure.MessageID,
		ZoneID:			failure.ZoneID,
		Executor:		"async-dead-letter",
		Status:			AVMReceiptStatusDeadLettered,
		GasUsed:		gasUsed,
		EventsHash:		failure.FailureHash,
		OutputMessagesRoot:	ComputeAVMAsyncFailureOutputRoot(failure),
		ErrorCodeOptional:	failure.ErrorCode,
		CreatedHeight:		finalHeight,
	})
	if err != nil {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, err
	}
	record, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:		failure.MessageID,
		ZoneID:			failure.ZoneID,
		Reason:			reason,
		FailedAttempts:		failure.Attempt,
		LastErrorCode:		failure.ErrorCode,
		FinalHeight:		finalHeight,
		RefundAmountOptional:	refundAmountNAET,
		ReceiptID:		receipt.ReceiptID,
	})
	if err != nil {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, err
	}
	if err := record.ValidateWithReceipt(receipt); err != nil {
		return AVMZoneDeadLetterQueue{}, AVMDeadLetterRecord{}, AVMExecutionReceipt{}, err
	}
	queue.Records = append(queue.Records, record)
	queue = canonicalAVMZoneDeadLetterQueue(queue)
	queue.QueueRoot = ComputeAVMZoneDeadLetterQueueRoot(queue)
	return queue, record, receipt, queue.Validate()
}

func NewAVMRetryExhaustionOutcome(queue AVMZoneDeadLetterQueue, message AVMAsyncMessage, attempt uint32, gasUsed uint64, finalHeight uint64, refundAmountNAET uint64) (AVMRetryExhaustionOutcome, error) {
	message = canonicalAVMAsyncMessage(message)
	if err := message.Validate(); err != nil {
		return AVMRetryExhaustionOutcome{}, err
	}
	failure, err := NewAVMAsyncFailureRecord(AVMAsyncFailureRecord{
		MessageID:	message.ID,
		ZoneID:		message.DestinationZone,
		FailureClass:	AVMAsyncFailureRetryExhausted,
		FailedHeight:	finalHeight,
		Attempt:	attempt,
		GasUsed:	gasUsed,
		RetryExhausted:	true,
	})
	if err != nil {
		return AVMRetryExhaustionOutcome{}, err
	}
	nextQueue, dead, receipt, err := DeadLetterAVMAsyncFailure(queue, failure, AVMDeadLetterTriggerRetryExhausted, "retry exhausted", refundAmountNAET, maxAVMFailureUint64(gasUsed, 1), finalHeight)
	if err != nil {
		return AVMRetryExhaustionOutcome{}, err
	}
	outcome := AVMRetryExhaustionOutcome{
		Queue:		nextQueue,
		Message:	message,
		Failure:	failure,
		Receipt:	receipt,
		DeadLetter:	dead,
	}
	outcome = canonicalAVMRetryExhaustionOutcome(outcome)
	outcome.OutcomeHash = ComputeAVMRetryExhaustionOutcomeHash(outcome)
	return outcome, outcome.Validate()
}

func (o AVMRetryExhaustionOutcome) Validate() error {
	o = canonicalAVMRetryExhaustionOutcome(o)
	if err := o.Queue.Validate(); err != nil {
		return err
	}
	if err := o.Message.Validate(); err != nil {
		return err
	}
	if err := o.Failure.Validate(); err != nil {
		return err
	}
	if err := o.DeadLetter.ValidateWithReceipt(o.Receipt); err != nil {
		return err
	}
	if o.Failure.FailureClass != AVMAsyncFailureRetryExhausted || !o.Failure.RetryExhausted {
		return errors.New("AVM retry exhaustion outcome requires retry_exhausted failure")
	}
	if o.Message.ID != o.Failure.MessageID || o.Message.ID != o.DeadLetter.MessageID {
		return errors.New("AVM retry exhaustion outcome message id mismatch")
	}
	if o.OutcomeHash == "" {
		return errors.New("AVM retry exhaustion outcome hash is required")
	}
	if err := zonestypes.ValidateHash("AVM retry exhaustion outcome hash", o.OutcomeHash); err != nil {
		return err
	}
	if o.OutcomeHash != ComputeAVMRetryExhaustionOutcomeHash(o) {
		return errors.New("AVM retry exhaustion outcome hash mismatch")
	}
	return nil
}

func NewAVMAsyncFailureValueConservationCheck(original AVMAsyncMessage, bounceValueNAET uint64, deadLetterRefundNAET uint64) (AVMAsyncFailureValueConservationCheck, error) {
	original = canonicalAVMAsyncMessage(original)
	if err := original.Validate(); err != nil {
		return AVMAsyncFailureValueConservationCheck{}, err
	}
	check := AVMAsyncFailureValueConservationCheck{
		OriginalMessageID:	original.ID,
		OriginalValueNAET:	original.ValueNAET,
		BounceValueNAET:	bounceValueNAET,
		DeadLetterRefundNAET:	deadLetterRefundNAET,
	}
	check = canonicalAVMAsyncFailureValueConservationCheck(check)
	check.ConservationCheckHash = ComputeAVMAsyncFailureValueConservationHash(check)
	return check, check.Validate()
}

func (c AVMAsyncFailureValueConservationCheck) Validate() error {
	c = canonicalAVMAsyncFailureValueConservationCheck(c)
	if err := zonestypes.ValidateHash("AVM async value conservation message id", c.OriginalMessageID); err != nil {
		return err
	}
	if c.BounceValueNAET > c.OriginalValueNAET {
		return errors.New("AVM async failure bounce value exceeds original value")
	}
	if c.DeadLetterRefundNAET > c.OriginalValueNAET {
		return errors.New("AVM async failure dead letter refund exceeds original value")
	}
	if c.BounceValueNAET > 0 && c.DeadLetterRefundNAET > 0 {
		return errors.New("AVM async failure value cannot be both bounced and dead-letter refunded")
	}
	if c.ConservationCheckHash == "" {
		return errors.New("AVM async failure value conservation hash is required")
	}
	if err := zonestypes.ValidateHash("AVM async failure value conservation hash", c.ConservationCheckHash); err != nil {
		return err
	}
	if c.ConservationCheckHash != ComputeAVMAsyncFailureValueConservationHash(c) {
		return errors.New("AVM async failure value conservation hash mismatch")
	}
	return nil
}

func (o AVMAsyncBounceDeadLetterOutcome) Validate() error {
	o = canonicalAVMAsyncBounceDeadLetterOutcome(o)
	if err := o.BounceOutcome.Validate(); err != nil {
		return err
	}
	if err := o.DeadLetter.ValidateWithReceipt(o.DeadLetterReceipt); err != nil {
		return err
	}
	if o.DeadLetter.MessageID != o.BounceOutcome.BounceMessage.ID {
		return errors.New("AVM async bounce dead letter must target bounce message")
	}
	if o.DeadLetter.ZoneID != o.BounceOutcome.BounceMessage.DestinationZone {
		return errors.New("AVM async bounce dead letter zone must match bounce destination")
	}
	if o.OutcomeHash == "" {
		return errors.New("AVM async bounce dead letter outcome hash is required")
	}
	if err := zonestypes.ValidateHash("AVM async bounce dead letter outcome hash", o.OutcomeHash); err != nil {
		return err
	}
	if o.OutcomeHash != ComputeAVMAsyncBounceDeadLetterOutcomeHash(o) {
		return errors.New("AVM async bounce dead letter outcome hash mismatch")
	}
	return nil
}

func AVMAsyncBouncePayload(originalMessageID, errorCode string) ([]byte, error) {
	if err := zonestypes.ValidateHash("AVM async bounce original message id", strings.TrimSpace(originalMessageID)); err != nil {
		return nil, err
	}
	if err := validateRouterOptionalToken("AVM async bounce error code", strings.TrimSpace(errorCode), MaxAVMFailureErrorCodeLength); err != nil {
		return nil, err
	}
	if strings.TrimSpace(errorCode) == "" {
		return nil, errors.New("AVM async bounce error code is required")
	}
	if !IsAVMAsyncErrorCode(AVMAsyncErrorCode(strings.TrimSpace(errorCode))) {
		return nil, fmt.Errorf("invalid AVM async error code %q", errorCode)
	}
	payload := []byte(fmt.Sprintf("bounce:%s:%s", originalMessageID, errorCode))
	if len(payload) > MaxAVMBouncePayloadBytes {
		return nil, fmt.Errorf("AVM async bounce payload must be <= %d bytes", MaxAVMBouncePayloadBytes)
	}
	return payload, nil
}

func AVMAsyncBounceRouteHint(originalMessageID string) string {
	return "bounce:" + strings.TrimSpace(originalMessageID)
}

func ComputeAVMAsyncFailureHash(record AVMAsyncFailureRecord) string {
	record = canonicalAVMAsyncFailureRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-failure-v1")
	writeEnginePart(h, record.MessageID)
	writeEnginePart(h, string(record.ZoneID))
	writeEnginePart(h, string(record.FailureClass))
	writeEnginePart(h, record.ErrorCode)
	writeEngineUint64(h, record.FailedHeight)
	writeEngineUint64(h, uint64(record.Attempt))
	writeEngineUint64(h, record.GasUsed)
	writeEngineBool(h, record.RetryExhausted)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncBounceMessageRoot(message AVMAsyncMessage) string {
	message = canonicalAVMAsyncMessage(message)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-bounce-message-root-v1")
	writeAVMAsyncMessageParts(h, message)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncFailureOutputRoot(failure AVMAsyncFailureRecord) string {
	failure = canonicalAVMAsyncFailureRecord(failure)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-failure-output-root-v1")
	writeAVMAsyncFailureParts(h, failure)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncFailedReceiptModelHash(model AVMAsyncFailedReceiptModel) string {
	model = canonicalAVMAsyncFailedReceiptModel(model)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-failed-receipt-model-v1")
	writeAVMAsyncMessageParts(h, model.Message)
	writeAVMAsyncFailureParts(h, model.Failure)
	writeEnginePart(h, model.Receipt.ReceiptHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncBounceOutcomeHash(outcome AVMAsyncBounceOutcome) string {
	outcome = canonicalAVMAsyncBounceOutcome(outcome)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-bounce-outcome-v1")
	writeAVMAsyncMessageParts(h, outcome.OriginalMessage)
	writeAVMAsyncFailureParts(h, outcome.Failure)
	writeAVMAsyncMessageParts(h, outcome.BounceMessage)
	writeEnginePart(h, outcome.BounceReceipt.ReceiptHash)
	writeEngineUint64(h, outcome.RemainingValueNAET)
	writeEngineUint64(h, outcome.BounceGasUsed)
	writeEngineUint64(h, outcome.BoundedBounceGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneDeadLetterQueueRoot(queue AVMZoneDeadLetterQueue) string {
	queue = canonicalAVMZoneDeadLetterQueue(queue)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-dead-letter-queue-v1")
	writeEnginePart(h, string(queue.ZoneID))
	writeEngineUint64(h, uint64(len(queue.Records)))
	for _, record := range queue.Records {
		writeEnginePart(h, record.RecordHash)
		writeEnginePart(h, AVMZoneDeadLetterSortKey(record))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMRetryExhaustionOutcomeHash(outcome AVMRetryExhaustionOutcome) string {
	outcome = canonicalAVMRetryExhaustionOutcome(outcome)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-retry-exhaustion-outcome-v1")
	writeEnginePart(h, outcome.Queue.QueueRoot)
	writeAVMAsyncMessageParts(h, outcome.Message)
	writeAVMAsyncFailureParts(h, outcome.Failure)
	writeEnginePart(h, outcome.Receipt.ReceiptHash)
	writeEnginePart(h, outcome.DeadLetter.RecordHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncFailureValueConservationHash(check AVMAsyncFailureValueConservationCheck) string {
	check = canonicalAVMAsyncFailureValueConservationCheck(check)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-failure-value-conservation-v1")
	writeEnginePart(h, check.OriginalMessageID)
	writeEngineUint64(h, check.OriginalValueNAET)
	writeEngineUint64(h, check.BounceValueNAET)
	writeEngineUint64(h, check.DeadLetterRefundNAET)
	return hex.EncodeToString(h.Sum(nil))
}

func AVMZoneDeadLetterSortKey(record AVMDeadLetterRecord) string {
	record = canonicalAVMDeadLetterRecord(record)
	return fmt.Sprintf("%020d/%s", record.FinalHeight, record.MessageID)
}

func ComputeAVMAsyncBounceDeadLetterOutcomeHash(outcome AVMAsyncBounceDeadLetterOutcome) string {
	outcome = canonicalAVMAsyncBounceDeadLetterOutcome(outcome)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-bounce-dead-letter-v1")
	writeEnginePart(h, outcome.BounceOutcome.OutcomeHash)
	writeEnginePart(h, outcome.DeadLetter.RecordHash)
	writeEnginePart(h, outcome.DeadLetterReceipt.ReceiptHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncBounceDeadLetterOutputRoot(bounceMessageID string) string {
	bounceMessageID = strings.TrimSpace(bounceMessageID)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-bounce-dead-letter-output-root-v1")
	writeEnginePart(h, bounceMessageID)
	return hex.EncodeToString(h.Sum(nil))
}

func IsAVMAsyncFailureClass(class AVMAsyncFailureClass) bool {
	switch class {
	case AVMAsyncFailureInvalidPayload,
		AVMAsyncFailureInsufficientGas,
		AVMAsyncFailureDestinationNotFound,
		AVMAsyncFailureDestinationDisabled,
		AVMAsyncFailureExpiredMessage,
		AVMAsyncFailureHandlerFailure,
		AVMAsyncFailureStorageLimitExceeded,
		AVMAsyncFailureProofVerificationFailure,
		AVMAsyncFailureRetryExhausted:
		return true
	default:
		return false
	}
}

func IsAVMAsyncErrorCode(code AVMAsyncErrorCode) bool {
	switch code {
	case AVMAsyncErrorInvalidPayload,
		AVMAsyncErrorInsufficientGas,
		AVMAsyncErrorDestinationNotFound,
		AVMAsyncErrorDestinationDisabled,
		AVMAsyncErrorExpiredMessage,
		AVMAsyncErrorHandlerFailure,
		AVMAsyncErrorStorageLimitExceeded,
		AVMAsyncErrorProofVerificationFailure,
		AVMAsyncErrorRetryExhausted,
		AVMAsyncErrorBounceFailed:
		return true
	default:
		return false
	}
}

func IsAVMDeadLetterTrigger(trigger AVMDeadLetterTrigger) bool {
	switch trigger {
	case AVMDeadLetterTriggerRetryExhausted,
		AVMDeadLetterTriggerDestinationPermanentlyDisabled,
		AVMDeadLetterTriggerBounceFailed,
		AVMDeadLetterTriggerExpiredBeforeExecution,
		AVMDeadLetterTriggerInvalidMessageFormat:
		return true
	default:
		return false
	}
}

func AVMAsyncFailureErrorCode(class AVMAsyncFailureClass) (AVMAsyncErrorCode, bool) {
	switch class {
	case AVMAsyncFailureInvalidPayload:
		return AVMAsyncErrorInvalidPayload, true
	case AVMAsyncFailureInsufficientGas:
		return AVMAsyncErrorInsufficientGas, true
	case AVMAsyncFailureDestinationNotFound:
		return AVMAsyncErrorDestinationNotFound, true
	case AVMAsyncFailureDestinationDisabled:
		return AVMAsyncErrorDestinationDisabled, true
	case AVMAsyncFailureExpiredMessage:
		return AVMAsyncErrorExpiredMessage, true
	case AVMAsyncFailureHandlerFailure:
		return AVMAsyncErrorHandlerFailure, true
	case AVMAsyncFailureStorageLimitExceeded:
		return AVMAsyncErrorStorageLimitExceeded, true
	case AVMAsyncFailureProofVerificationFailure:
		return AVMAsyncErrorProofVerificationFailure, true
	case AVMAsyncFailureRetryExhausted:
		return AVMAsyncErrorRetryExhausted, true
	default:
		return "", false
	}
}

func AVMDeadLetterTriggerForFailure(failure AVMAsyncFailureRecord) (AVMDeadLetterTrigger, bool) {
	switch failure.FailureClass {
	case AVMAsyncFailureRetryExhausted:
		return AVMDeadLetterTriggerRetryExhausted, true
	case AVMAsyncFailureDestinationDisabled:
		return AVMDeadLetterTriggerDestinationPermanentlyDisabled, true
	case AVMAsyncFailureExpiredMessage:
		return AVMDeadLetterTriggerExpiredBeforeExecution, true
	case AVMAsyncFailureInvalidPayload:
		return AVMDeadLetterTriggerInvalidMessageFormat, true
	default:
		if failure.ErrorCode == string(AVMAsyncErrorBounceFailed) {
			return AVMDeadLetterTriggerBounceFailed, true
		}
		return "", false
	}
}

func ValidateAVMDeadLetterTriggerForFailure(trigger AVMDeadLetterTrigger, failure AVMAsyncFailureRecord) error {
	if trigger == AVMDeadLetterTriggerBounceFailed {
		if failure.ErrorCode == string(AVMAsyncErrorBounceFailed) {
			return nil
		}
		return errors.New("AVM bounce failed trigger requires bounce failed error code")
	}
	expected, ok := AVMDeadLetterTriggerForFailure(failure)
	if !ok {
		return errors.New("AVM async failure is not a dead letter trigger")
	}
	if expected != trigger {
		return errors.New("AVM dead letter trigger does not match failure class")
	}
	return nil
}

func canonicalAVMAsyncFailureRecord(record AVMAsyncFailureRecord) AVMAsyncFailureRecord {
	record.MessageID = strings.TrimSpace(record.MessageID)
	record.ErrorCode = strings.TrimSpace(record.ErrorCode)
	record.FailureHash = strings.TrimSpace(record.FailureHash)
	return record
}

func canonicalAVMAsyncFailedReceiptModel(model AVMAsyncFailedReceiptModel) AVMAsyncFailedReceiptModel {
	model.Message = canonicalAVMAsyncMessage(model.Message)
	model.Failure = canonicalAVMAsyncFailureRecord(model.Failure)
	model.Receipt = canonicalAVMExecutionReceipt(model.Receipt)
	model.ModelHash = strings.TrimSpace(model.ModelHash)
	return model
}

func canonicalAVMAsyncBounceOutcome(outcome AVMAsyncBounceOutcome) AVMAsyncBounceOutcome {
	outcome.OriginalMessage = canonicalAVMAsyncMessage(outcome.OriginalMessage)
	outcome.Failure = canonicalAVMAsyncFailureRecord(outcome.Failure)
	outcome.BounceMessage = canonicalAVMAsyncMessage(outcome.BounceMessage)
	outcome.BounceReceipt = canonicalAVMExecutionReceipt(outcome.BounceReceipt)
	outcome.OutcomeHash = strings.TrimSpace(outcome.OutcomeHash)
	return outcome
}

func canonicalAVMZoneDeadLetterQueue(queue AVMZoneDeadLetterQueue) AVMZoneDeadLetterQueue {
	records := append([]AVMDeadLetterRecord(nil), queue.Records...)
	for i := range records {
		records[i] = canonicalAVMDeadLetterRecord(records[i])
	}
	sort.Slice(records, func(i, j int) bool {
		return AVMZoneDeadLetterSortKey(records[i]) < AVMZoneDeadLetterSortKey(records[j])
	})
	queue.Records = records
	queue.QueueRoot = strings.TrimSpace(queue.QueueRoot)
	return queue
}

func canonicalAVMRetryExhaustionOutcome(outcome AVMRetryExhaustionOutcome) AVMRetryExhaustionOutcome {
	outcome.Queue = canonicalAVMZoneDeadLetterQueue(outcome.Queue)
	outcome.Message = canonicalAVMAsyncMessage(outcome.Message)
	outcome.Failure = canonicalAVMAsyncFailureRecord(outcome.Failure)
	outcome.Receipt = canonicalAVMExecutionReceipt(outcome.Receipt)
	outcome.DeadLetter = canonicalAVMDeadLetterRecord(outcome.DeadLetter)
	outcome.OutcomeHash = strings.TrimSpace(outcome.OutcomeHash)
	return outcome
}

func canonicalAVMAsyncFailureValueConservationCheck(check AVMAsyncFailureValueConservationCheck) AVMAsyncFailureValueConservationCheck {
	check.OriginalMessageID = strings.TrimSpace(check.OriginalMessageID)
	check.ConservationCheckHash = strings.TrimSpace(check.ConservationCheckHash)
	return check
}

func canonicalAVMAsyncBounceDeadLetterOutcome(outcome AVMAsyncBounceDeadLetterOutcome) AVMAsyncBounceDeadLetterOutcome {
	outcome.BounceOutcome = canonicalAVMAsyncBounceOutcome(outcome.BounceOutcome)
	outcome.DeadLetter = canonicalAVMDeadLetterRecord(outcome.DeadLetter)
	outcome.DeadLetterReceipt = canonicalAVMExecutionReceipt(outcome.DeadLetterReceipt)
	outcome.OutcomeHash = strings.TrimSpace(outcome.OutcomeHash)
	return outcome
}

func writeAVMAsyncFailureParts(h engineByteWriter, record AVMAsyncFailureRecord) {
	record = canonicalAVMAsyncFailureRecord(record)
	writeEnginePart(h, record.MessageID)
	writeEnginePart(h, string(record.ZoneID))
	writeEnginePart(h, string(record.FailureClass))
	writeEnginePart(h, record.ErrorCode)
	writeEngineUint64(h, record.FailedHeight)
	writeEngineUint64(h, uint64(record.Attempt))
	writeEngineUint64(h, record.GasUsed)
	writeEngineBool(h, record.RetryExhausted)
	writeEnginePart(h, record.FailureHash)
}

func writeAVMAsyncMessageParts(h engineByteWriter, msg AVMAsyncMessage) {
	msg = canonicalAVMAsyncMessage(msg)
	writeEnginePart(h, msg.ID)
	writeEnginePart(h, msg.ChainID)
	writeEnginePart(h, msg.Source)
	writeEnginePart(h, msg.Destination)
	writeEnginePart(h, string(msg.Payload))
	writeEngineUint64(h, msg.GasLimit)
	writeEngineUint64(h, msg.DelayHeight)
	writeEngineUint64(h, msg.ExpiryHeight)
	writeEnginePart(h, string(msg.RetryPolicy.Mode))
	writeEngineUint64(h, uint64(msg.RetryPolicy.MaxAttempts))
	writeEngineUint64(h, msg.RetryPolicy.RetryDelay)
	writeEnginePart(h, string(msg.RetryPolicy.BackoffMode))
	writeEngineUint64(h, msg.RetryPolicy.MaxRetryHeight)
	writeEngineBool(h, msg.RetryPolicy.ChargeRetryGas)
	writeEngineBool(h, msg.BounceFlag)
	writeEnginePart(h, string(msg.SourceZone))
	writeEnginePart(h, string(msg.DestinationZone))
	writeEnginePart(h, msg.SourceActorOptional)
	writeEnginePart(h, msg.DestinationActorOptional)
	writeEngineUint64(h, msg.SenderNonce)
	writeEnginePart(h, msg.PayloadType)
	writeEnginePart(h, msg.PayloadHash)
	writeEngineUint64(h, msg.ValueNAET)
	writeEngineUint64(h, msg.ForwardingFee)
	writeEngineUint64(h, uint64(msg.Priority))
	writeEngineUint64(h, msg.CreatedHeight)
	writeEnginePart(h, msg.RouteHintOptional)
	writeEnginePart(h, msg.AuthProofOptional)
	writeEnginePart(h, msg.StateProofOptional)
}

func maxAVMFailureUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
