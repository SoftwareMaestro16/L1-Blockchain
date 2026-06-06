package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMAsyncFailureInvalidPayload           AVMAsyncFailureClass = "invalid_payload"
	AVMAsyncFailureInsufficientGas          AVMAsyncFailureClass = "insufficient_gas"
	AVMAsyncFailureDestinationNotFound      AVMAsyncFailureClass = "destination_not_found"
	AVMAsyncFailureDestinationDisabled      AVMAsyncFailureClass = "destination_disabled"
	AVMAsyncFailureExpiredMessage           AVMAsyncFailureClass = "expired_message"
	AVMAsyncFailureHandlerFailure           AVMAsyncFailureClass = "handler_failure"
	AVMAsyncFailureStorageLimitExceeded     AVMAsyncFailureClass = "storage_limit_exceeded"
	AVMAsyncFailureProofVerificationFailure AVMAsyncFailureClass = "proof_verification_failure"
	AVMAsyncFailureRetryExhausted           AVMAsyncFailureClass = "retry_exhausted"

	AVMAsyncBouncePayloadType = "async.bounce"

	MaxAVMBouncePayloadBytes     = 1024
	MaxAVMFailureErrorCodeLength = MaxAVMReceiptErrorCode
)

type AVMAsyncFailureClass string

type AVMAsyncFailureRecord struct {
	MessageID      string
	ZoneID         zonestypes.ZoneID
	FailureClass   AVMAsyncFailureClass
	ErrorCode      string
	FailedHeight   uint64
	Attempt        uint32
	GasUsed        uint64
	RetryExhausted bool
	FailureHash    string
}

type AVMAsyncBounceOutcome struct {
	OriginalMessage    AVMAsyncMessage
	Failure            AVMAsyncFailureRecord
	BounceMessage      AVMAsyncMessage
	BounceReceipt      AVMExecutionReceipt
	RemainingValueNAET uint64
	BounceGasUsed      uint64
	BoundedBounceGas   uint64
	OutcomeHash        string
}

type AVMAsyncBounceDeadLetterOutcome struct {
	BounceOutcome     AVMAsyncBounceOutcome
	DeadLetter        AVMDeadLetterRecord
	DeadLetterReceipt AVMExecutionReceipt
	OutcomeHash       string
}

func NewAVMAsyncFailureRecord(record AVMAsyncFailureRecord) (AVMAsyncFailureRecord, error) {
	record = canonicalAVMAsyncFailureRecord(record)
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
		ChainID:                  original.ChainID,
		Source:                   original.Destination,
		Destination:              original.Source,
		Payload:                  payload,
		GasLimit:                 schedule.BounceGas,
		ExpiryHeight:             expiryHeight,
		RetryPolicy:              AVMRetryPolicy{Mode: AVMRetryModeNone, BackoffMode: AVMBackoffModeNone},
		BounceFlag:               false,
		SourceZone:               original.DestinationZone,
		DestinationZone:          original.SourceZone,
		SourceActorOptional:      original.DestinationActorOptional,
		DestinationActorOptional: original.SourceActorOptional,
		SenderNonce:              bounceSenderNonce,
		PayloadType:              AVMAsyncBouncePayloadType,
		ValueNAET:                remainingValueNAET,
		ForwardingFee:            original.ForwardingFee,
		Priority:                 original.Priority,
		CreatedHeight:            createdHeight,
		RouteHintOptional:        AVMAsyncBounceRouteHint(original.ID),
	})
	if err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:          original.ID,
		ZoneID:             original.DestinationZone,
		Executor:           "async-bounce",
		Status:             AVMReceiptStatusBounced,
		GasUsed:            schedule.BounceGas,
		EventsHash:         failure.FailureHash,
		OutputMessagesRoot: ComputeAVMAsyncBounceMessageRoot(bounce),
		ErrorCodeOptional:  failure.ErrorCode,
		CreatedHeight:      createdHeight,
	})
	if err != nil {
		return AVMAsyncBounceOutcome{}, err
	}
	outcome := AVMAsyncBounceOutcome{
		OriginalMessage:    original,
		Failure:            failure,
		BounceMessage:      bounce,
		BounceReceipt:      receipt,
		RemainingValueNAET: remainingValueNAET,
		BounceGasUsed:      schedule.BounceGas,
		BoundedBounceGas:   schedule.BounceGas,
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
		MessageID:          outcome.BounceMessage.ID,
		ZoneID:             outcome.BounceMessage.DestinationZone,
		Executor:           "async-bounce-dead-letter",
		Status:             AVMReceiptStatusDeadLettered,
		GasUsed:            outcome.BounceGasUsed,
		EventsHash:         outcome.Failure.FailureHash,
		OutputMessagesRoot: ComputeAVMAsyncBounceDeadLetterOutputRoot(outcome.BounceMessage.ID),
		ErrorCodeOptional:  outcome.Failure.ErrorCode,
		CreatedHeight:      finalHeight,
	})
	if err != nil {
		return AVMAsyncBounceDeadLetterOutcome{}, err
	}
	dead, err := NewAVMDeadLetterRecord(AVMDeadLetterRecord{
		MessageID:            outcome.BounceMessage.ID,
		ZoneID:               outcome.BounceMessage.DestinationZone,
		Reason:               reason,
		FailedAttempts:       failedAttempts,
		LastErrorCode:        outcome.Failure.ErrorCode,
		FinalHeight:          finalHeight,
		RefundAmountOptional: outcome.RemainingValueNAET,
		ReceiptID:            receipt.ReceiptID,
	})
	if err != nil {
		return AVMAsyncBounceDeadLetterOutcome{}, err
	}
	next := AVMAsyncBounceDeadLetterOutcome{
		BounceOutcome:     outcome,
		DeadLetter:        dead,
		DeadLetterReceipt: receipt,
	}
	next = canonicalAVMAsyncBounceDeadLetterOutcome(next)
	next.OutcomeHash = ComputeAVMAsyncBounceDeadLetterOutcomeHash(next)
	return next, next.Validate()
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
	writeEnginePart(h, "aetheris-avm-async-failure-v1")
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
	writeEnginePart(h, "aetheris-avm-async-bounce-message-root-v1")
	writeAVMAsyncMessageParts(h, message)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncBounceOutcomeHash(outcome AVMAsyncBounceOutcome) string {
	outcome = canonicalAVMAsyncBounceOutcome(outcome)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-async-bounce-outcome-v1")
	writeAVMAsyncMessageParts(h, outcome.OriginalMessage)
	writeAVMAsyncFailureParts(h, outcome.Failure)
	writeAVMAsyncMessageParts(h, outcome.BounceMessage)
	writeEnginePart(h, outcome.BounceReceipt.ReceiptHash)
	writeEngineUint64(h, outcome.RemainingValueNAET)
	writeEngineUint64(h, outcome.BounceGasUsed)
	writeEngineUint64(h, outcome.BoundedBounceGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncBounceDeadLetterOutcomeHash(outcome AVMAsyncBounceDeadLetterOutcome) string {
	outcome = canonicalAVMAsyncBounceDeadLetterOutcome(outcome)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-async-bounce-dead-letter-v1")
	writeEnginePart(h, outcome.BounceOutcome.OutcomeHash)
	writeEnginePart(h, outcome.DeadLetter.RecordHash)
	writeEnginePart(h, outcome.DeadLetterReceipt.ReceiptHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncBounceDeadLetterOutputRoot(bounceMessageID string) string {
	bounceMessageID = strings.TrimSpace(bounceMessageID)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-async-bounce-dead-letter-output-root-v1")
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

func canonicalAVMAsyncFailureRecord(record AVMAsyncFailureRecord) AVMAsyncFailureRecord {
	record.MessageID = strings.TrimSpace(record.MessageID)
	record.ErrorCode = strings.TrimSpace(record.ErrorCode)
	record.FailureHash = strings.TrimSpace(record.FailureHash)
	return record
}

func canonicalAVMAsyncBounceOutcome(outcome AVMAsyncBounceOutcome) AVMAsyncBounceOutcome {
	outcome.OriginalMessage = canonicalAVMAsyncMessage(outcome.OriginalMessage)
	outcome.Failure = canonicalAVMAsyncFailureRecord(outcome.Failure)
	outcome.BounceMessage = canonicalAVMAsyncMessage(outcome.BounceMessage)
	outcome.BounceReceipt = canonicalAVMExecutionReceipt(outcome.BounceReceipt)
	outcome.OutcomeHash = strings.TrimSpace(outcome.OutcomeHash)
	return outcome
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
