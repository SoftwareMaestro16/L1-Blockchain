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
	MaxAVMDeadLetterReasonLength	= 256
	MaxAVMDeadLetterErrorCodeLength	= MaxAVMReceiptErrorCode
)

type AVMDeadLetterRecord struct {
	MessageID		string
	ZoneID			zonestypes.ZoneID
	Reason			string
	FailedAttempts		uint32
	LastErrorCode		string
	FinalHeight		uint64
	RefundAmountOptional	uint64
	ReceiptID		string
	RecordHash		string
}

func NewAVMDeadLetterRecord(record AVMDeadLetterRecord) (AVMDeadLetterRecord, error) {
	record = canonicalAVMDeadLetterRecord(record)
	record.RecordHash = ComputeAVMDeadLetterRecordHash(record)
	return record, record.Validate()
}

func (r AVMDeadLetterRecord) Validate() error {
	r = canonicalAVMDeadLetterRecord(r)
	if err := zonestypes.ValidateHash("AVM dead letter message id", r.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if err := validateDeadLetterReason(r.Reason); err != nil {
		return err
	}
	if r.FailedAttempts == 0 {
		return errors.New("AVM dead letter failed attempts must be positive")
	}
	if err := validateRouterOptionalToken("AVM dead letter last error code", r.LastErrorCode, MaxAVMDeadLetterErrorCodeLength); err != nil {
		return err
	}
	if r.LastErrorCode == "" {
		return errors.New("AVM dead letter last error code is required")
	}
	if r.FinalHeight == 0 {
		return errors.New("AVM dead letter final height must be positive")
	}
	if err := zonestypes.ValidateHash("AVM dead letter receipt id", r.ReceiptID); err != nil {
		return err
	}
	if err := validateAVMStatePrefix("AVM dead letter proof key", r.ProofKey()); err != nil {
		return err
	}
	if r.RecordHash == "" {
		return errors.New("AVM dead letter record hash is required")
	}
	if err := zonestypes.ValidateHash("AVM dead letter record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMDeadLetterRecordHash(r) {
		return errors.New("AVM dead letter record hash mismatch")
	}
	return nil
}

func (r AVMDeadLetterRecord) ValidateWithReceipt(receipt AVMExecutionReceipt) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := receipt.Validate(); err != nil {
		return fmt.Errorf("AVM dead letter receipt: %w", err)
	}
	if receipt.Status != AVMReceiptStatusDeadLettered {
		return errors.New("AVM dead letter receipt must be terminal dead_lettered status")
	}
	if receipt.MessageID != r.MessageID {
		return errors.New("AVM dead letter receipt message id mismatch")
	}
	if receipt.ZoneID != r.ZoneID {
		return errors.New("AVM dead letter receipt zone id mismatch")
	}
	if receipt.ReceiptID != r.ReceiptID {
		return errors.New("AVM dead letter receipt id mismatch")
	}
	if receipt.CreatedHeight != r.FinalHeight {
		return errors.New("AVM dead letter final height must match receipt height")
	}
	return nil
}

func (r AVMDeadLetterRecord) ProofKey() string {
	return AVMDeadLetterProofKey(r.ZoneID, r.MessageID)
}

func AVMDeadLetterProofKey(zoneID zonestypes.ZoneID, messageID string) string {
	return AVMAsyncDeadLetterKey(zoneID, messageID)
}

func (r AVMDeadLetterRecord) CanTriggerBounce(message AVMAsyncMessage) bool {
	r = canonicalAVMDeadLetterRecord(r)
	message = canonicalAVMAsyncMessage(message)
	return message.BounceFlag &&
		message.ID == r.MessageID &&
		message.DestinationZone == r.ZoneID
}

func ComputeAVMDeadLetterRecordHash(record AVMDeadLetterRecord) string {
	record = canonicalAVMDeadLetterRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-dead-letter-record-v1")
	writeEnginePart(h, record.MessageID)
	writeEnginePart(h, string(record.ZoneID))
	writeEnginePart(h, record.Reason)
	writeEngineUint64(h, uint64(record.FailedAttempts))
	writeEnginePart(h, record.LastErrorCode)
	writeEngineUint64(h, record.FinalHeight)
	writeEngineUint64(h, record.RefundAmountOptional)
	writeEnginePart(h, record.ReceiptID)
	writeEnginePart(h, record.ProofKey())
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMDeadLetterRecord(record AVMDeadLetterRecord) AVMDeadLetterRecord {
	record.MessageID = strings.TrimSpace(record.MessageID)
	record.Reason = strings.TrimSpace(record.Reason)
	record.LastErrorCode = strings.TrimSpace(record.LastErrorCode)
	record.ReceiptID = strings.TrimSpace(record.ReceiptID)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func validateDeadLetterReason(reason string) error {
	if strings.TrimSpace(reason) != reason {
		return errors.New("AVM dead letter reason must not have surrounding whitespace")
	}
	if reason == "" {
		return errors.New("AVM dead letter reason is required")
	}
	if len(reason) > MaxAVMDeadLetterReasonLength {
		return fmt.Errorf("AVM dead letter reason must be <= %d bytes", MaxAVMDeadLetterReasonLength)
	}
	for _, r := range reason {
		if r < 0x20 || r == 0x7f {
			return errors.New("AVM dead letter reason contains control character")
		}
	}
	return nil
}
