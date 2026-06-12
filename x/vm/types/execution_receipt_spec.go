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
	AVMReceiptStatusSubmitted	AVMReceiptStatus	= "submitted"
	AVMReceiptStatusScheduled	AVMReceiptStatus	= "scheduled"
	AVMReceiptStatusExecuted	AVMReceiptStatus	= "executed"
	AVMReceiptStatusFailed		AVMReceiptStatus	= "failed"
	AVMReceiptStatusRetried		AVMReceiptStatus	= "retried"
	AVMReceiptStatusExpired		AVMReceiptStatus	= "expired"
	AVMReceiptStatusBounced		AVMReceiptStatus	= "bounced"
	AVMReceiptStatusDeadLettered	AVMReceiptStatus	= "dead_lettered"

	MaxAVMReceiptExecutorLength	= 128
	MaxAVMReceiptErrorCode		= 128
)

type AVMReceiptStatus string

type AVMExecutionReceipt struct {
	ReceiptID		string
	MessageID		string
	ZoneID			zonestypes.ZoneID
	Executor		string
	Status			AVMReceiptStatus
	GasUsed			uint64
	StorageWritten		uint32
	EventsHash		string
	OutputMessagesRoot	string
	ErrorCodeOptional	string
	CreatedHeight		uint64
	ReceiptHash		string
}

func NewAVMExecutionReceipt(receipt AVMExecutionReceipt) (AVMExecutionReceipt, error) {
	receipt = canonicalAVMExecutionReceipt(receipt)
	if receipt.ReceiptID == "" {
		receipt.ReceiptID = DeriveAVMReceiptID(receipt)
	}
	receipt.ReceiptHash = ComputeAVMExecutionReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func (r AVMExecutionReceipt) Validate() error {
	r = canonicalAVMExecutionReceipt(r)
	if err := zonestypes.ValidateHash("AVM receipt id", r.ReceiptID); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM receipt message id", r.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM receipt executor", r.Executor, MaxAVMReceiptExecutorLength); err != nil {
		return err
	}
	if r.Executor == "" {
		return errors.New("AVM receipt executor is required")
	}
	if !IsAVMReceiptStatus(r.Status) {
		return fmt.Errorf("invalid AVM receipt status %q", r.Status)
	}
	if err := zonestypes.ValidateHash("AVM receipt events hash", r.EventsHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM receipt output messages root", r.OutputMessagesRoot); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM receipt error code", r.ErrorCodeOptional, MaxAVMReceiptErrorCode); err != nil {
		return err
	}
	if r.CreatedHeight == 0 {
		return errors.New("AVM receipt created height must be positive")
	}
	if requiresReceiptGas(r.Status) && r.GasUsed == 0 {
		return errors.New("AVM receipt gas used must be positive for terminal execution status")
	}
	if r.Status == AVMReceiptStatusSubmitted || r.Status == AVMReceiptStatusScheduled {
		if r.StorageWritten != 0 {
			return errors.New("non-executed AVM receipt must not write storage")
		}
		if r.ErrorCodeOptional != "" {
			return errors.New("non-terminal AVM receipt must not carry error code")
		}
	}
	if r.ReceiptID != DeriveAVMReceiptID(r) {
		return errors.New("AVM receipt id mismatch")
	}
	if r.ReceiptHash == "" {
		return errors.New("AVM receipt hash is required")
	}
	if r.ReceiptHash != ComputeAVMExecutionReceiptHash(r) {
		return errors.New("AVM receipt hash mismatch")
	}
	return nil
}

func DeriveAVMReceiptID(receipt AVMExecutionReceipt) string {
	receipt = canonicalAVMExecutionReceipt(receipt)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-receipt-id-v1")
	writeEnginePart(h, receipt.MessageID)
	writeEnginePart(h, string(receipt.ZoneID))
	writeEnginePart(h, receipt.Executor)
	writeEngineUint64(h, receipt.CreatedHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMExecutionReceiptHash(receipt AVMExecutionReceipt) string {
	receipt = canonicalAVMExecutionReceipt(receipt)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-execution-receipt-v1")
	writeEnginePart(h, receipt.ReceiptID)
	writeEnginePart(h, receipt.MessageID)
	writeEnginePart(h, string(receipt.ZoneID))
	writeEnginePart(h, receipt.Executor)
	writeEnginePart(h, string(receipt.Status))
	writeEngineUint64(h, receipt.GasUsed)
	writeEngineUint64(h, uint64(receipt.StorageWritten))
	writeEnginePart(h, receipt.EventsHash)
	writeEnginePart(h, receipt.OutputMessagesRoot)
	writeEnginePart(h, receipt.ErrorCodeOptional)
	writeEngineUint64(h, receipt.CreatedHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func IsAVMReceiptStatus(status AVMReceiptStatus) bool {
	switch status {
	case AVMReceiptStatusSubmitted,
		AVMReceiptStatusScheduled,
		AVMReceiptStatusExecuted,
		AVMReceiptStatusFailed,
		AVMReceiptStatusRetried,
		AVMReceiptStatusExpired,
		AVMReceiptStatusBounced,
		AVMReceiptStatusDeadLettered:
		return true
	default:
		return false
	}
}

func requiresReceiptGas(status AVMReceiptStatus) bool {
	switch status {
	case AVMReceiptStatusExecuted,
		AVMReceiptStatusFailed,
		AVMReceiptStatusRetried,
		AVMReceiptStatusExpired,
		AVMReceiptStatusBounced,
		AVMReceiptStatusDeadLettered:
		return true
	default:
		return false
	}
}

func canonicalAVMExecutionReceipt(receipt AVMExecutionReceipt) AVMExecutionReceipt {
	receipt.ReceiptID = strings.TrimSpace(receipt.ReceiptID)
	receipt.MessageID = strings.TrimSpace(receipt.MessageID)
	receipt.Executor = strings.TrimSpace(receipt.Executor)
	receipt.EventsHash = strings.TrimSpace(receipt.EventsHash)
	receipt.OutputMessagesRoot = strings.TrimSpace(receipt.OutputMessagesRoot)
	receipt.ErrorCodeOptional = strings.TrimSpace(receipt.ErrorCodeOptional)
	receipt.ReceiptHash = strings.TrimSpace(receipt.ReceiptHash)
	return receipt
}
