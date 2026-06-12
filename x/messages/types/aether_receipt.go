package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const MaxAetherErrorCodeLength = 128

type ReceiptStatus string

const (
	ReceiptStatusAccepted	ReceiptStatus	= "accepted"
	ReceiptStatusExecuted	ReceiptStatus	= "executed"
	ReceiptStatusFailed	ReceiptStatus	= "failed"
	ReceiptStatusExpired	ReceiptStatus	= "expired"
	ReceiptStatusBounced	ReceiptStatus	= "bounced"
	ReceiptStatusRejected	ReceiptStatus	= "rejected"
	ReceiptStatusDeferred	ReceiptStatus	= "deferred"
)

type AetherMessageReceipt struct {
	MsgID			string
	Height			uint64
	ReceiverZoneID		zonestypes.ZoneID
	ReceiverShardID		string
	Status			ReceiptStatus
	GasUsed			uint64
	FeeCharged		sdkmath.Int
	ReturnPayload		[]byte
	ErrorCode		string
	OutputMessagesRoot	string
	StateWriteSummaryHash	string
	ReceiptHash		string
}

func NewAetherMessageReceipt(receipt AetherMessageReceipt) (AetherMessageReceipt, error) {
	if receipt.ReceiptHash != "" {
		return AetherMessageReceipt{}, errors.New("aether message receipt hash must be empty before construction")
	}
	receipt = normalizeAetherMessageReceipt(receipt)
	if err := receipt.ValidateForHash(); err != nil {
		return AetherMessageReceipt{}, err
	}
	receipt.ReceiptHash = ComputeAetherMessageReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func AetherReceiptFromMessage(msg AetherMessage, status ReceiptStatus, height uint64, gasUsed uint64, feeCharged sdkmath.Int, returnPayload []byte, errorCode string, outputMessagesRoot string, stateWriteSummaryHash string) (AetherMessageReceipt, error) {
	if err := msg.Validate(); err != nil {
		return AetherMessageReceipt{}, err
	}
	return NewAetherMessageReceipt(AetherMessageReceipt{
		MsgID:			msg.MsgID,
		Height:			height,
		ReceiverZoneID:		msg.ReceiverZoneID,
		ReceiverShardID:	msg.ReceiverShardID,
		Status:			status,
		GasUsed:		gasUsed,
		FeeCharged:		feeCharged,
		ReturnPayload:		returnPayload,
		ErrorCode:		errorCode,
		OutputMessagesRoot:	outputMessagesRoot,
		StateWriteSummaryHash:	stateWriteSummaryHash,
	})
}

func (r AetherMessageReceipt) Validate() error {
	r = normalizeAetherMessageReceipt(r)
	if err := r.ValidateForHash(); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("aether message receipt hash", r.ReceiptHash); err != nil {
		return err
	}
	if r.ReceiptHash != ComputeAetherMessageReceiptHash(r) {
		return errors.New("aether message receipt hash mismatch")
	}
	return nil
}

func (r AetherMessageReceipt) ValidateForHash() error {
	r = normalizeAetherMessageReceipt(r)
	if err := zonestypes.ValidateHash("aether receipt message id", r.MsgID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("aether receipt height must be positive")
	}
	if err := zonestypes.ValidateZoneID(r.ReceiverZoneID); err != nil {
		return err
	}
	if err := validateToken("aether receipt receiver shard id", r.ReceiverShardID, MaxShardIDLength); err != nil {
		return err
	}
	if !IsReceiptStatus(r.Status) {
		return fmt.Errorf("unknown aether receipt status %q", r.Status)
	}
	if r.FeeCharged.IsNil() {
		return errors.New("aether receipt fee charged is required")
	}
	if r.FeeCharged.IsNegative() {
		return errors.New("aether receipt fee charged must be non-negative")
	}
	if r.ErrorCode != "" {
		if err := validateToken("aether receipt error code", r.ErrorCode, MaxAetherErrorCodeLength); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("aether receipt output messages root", r.OutputMessagesRoot); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("aether receipt state write summary hash", r.StateWriteSummaryHash); err != nil {
		return err
	}
	return nil
}

func (r AetherMessageReceipt) Clone() AetherMessageReceipt {
	r.ReturnPayload = append([]byte(nil), r.ReturnPayload...)
	return r
}

func ComputeAetherMessageReceiptHash(receipt AetherMessageReceipt) string {
	receipt = normalizeAetherMessageReceipt(receipt)
	return hashParts(
		"aetra-aether-message-receipt-v1",
		receipt.MsgID,
		fmt.Sprint(receipt.Height),
		string(receipt.ReceiverZoneID),
		receipt.ReceiverShardID,
		string(receipt.Status),
		fmt.Sprint(receipt.GasUsed),
		receipt.FeeCharged.String(),
		ComputeAetherPayloadHash(receipt.ReturnPayload),
		receipt.ErrorCode,
		receipt.OutputMessagesRoot,
		receipt.StateWriteSummaryHash,
	)
}

func ComputeAetherReceiptRoot(receipts []AetherMessageReceipt) (string, error) {
	ordered := cloneAetherMessageReceipts(receipts)
	sortAetherMessageReceipts(ordered)
	parts := []string{"aetra-aether-message-receipt-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		if err := receipt.Validate(); err != nil {
			return "", err
		}
		parts = append(parts, receipt.ReceiptHash)
	}
	return hashParts(parts...), nil
}

func CanonicalAetherMessageReceiptBinary(receipt AetherMessageReceipt) ([]byte, error) {
	if err := receipt.Validate(); err != nil {
		return nil, err
	}
	receipt = normalizeAetherMessageReceipt(receipt)
	buf := bytes.NewBuffer(nil)
	for _, part := range []string{
		receipt.MsgID,
		fmt.Sprint(receipt.Height),
		string(receipt.ReceiverZoneID),
		receipt.ReceiverShardID,
		string(receipt.Status),
		fmt.Sprint(receipt.GasUsed),
		receipt.FeeCharged.String(),
		hex.EncodeToString(receipt.ReturnPayload),
		receipt.ErrorCode,
		receipt.OutputMessagesRoot,
		receipt.StateWriteSummaryHash,
		receipt.ReceiptHash,
	} {
		writeString(buf.Write, part)
	}
	return buf.Bytes(), nil
}

func IsReceiptStatus(status ReceiptStatus) bool {
	switch status {
	case ReceiptStatusAccepted, ReceiptStatusExecuted, ReceiptStatusFailed, ReceiptStatusExpired, ReceiptStatusBounced, ReceiptStatusRejected, ReceiptStatusDeferred:
		return true
	default:
		return false
	}
}

func normalizeAetherMessageReceipt(receipt AetherMessageReceipt) AetherMessageReceipt {
	receipt.MsgID = strings.ToLower(strings.TrimSpace(receipt.MsgID))
	receipt.ReceiverShardID = strings.TrimSpace(receipt.ReceiverShardID)
	if receipt.FeeCharged.IsNil() {
		receipt.FeeCharged = sdkmath.ZeroInt()
	}
	receipt.ReturnPayload = append([]byte(nil), receipt.ReturnPayload...)
	receipt.ErrorCode = strings.TrimSpace(receipt.ErrorCode)
	receipt.OutputMessagesRoot = strings.ToLower(strings.TrimSpace(receipt.OutputMessagesRoot))
	receipt.StateWriteSummaryHash = strings.ToLower(strings.TrimSpace(receipt.StateWriteSummaryHash))
	receipt.ReceiptHash = strings.ToLower(strings.TrimSpace(receipt.ReceiptHash))
	return receipt
}

func cloneAetherMessageReceipts(receipts []AetherMessageReceipt) []AetherMessageReceipt {
	out := make([]AetherMessageReceipt, len(receipts))
	for i, receipt := range receipts {
		out[i] = receipt.Clone()
	}
	return out
}

func sortAetherMessageReceipts(receipts []AetherMessageReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		if receipts[i].Height != receipts[j].Height {
			return receipts[i].Height < receipts[j].Height
		}
		if receipts[i].ReceiverZoneID != receipts[j].ReceiverZoneID {
			return receipts[i].ReceiverZoneID < receipts[j].ReceiverZoneID
		}
		if receipts[i].ReceiverShardID != receipts[j].ReceiverShardID {
			return receipts[i].ReceiverShardID < receipts[j].ReceiverShardID
		}
		return receipts[i].MsgID < receipts[j].MsgID
	})
}
