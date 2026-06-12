package async

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ReceiptIDLength = 64
)

func newExecutionReceipt(queued QueuedMessage, height uint64, stateRootBefore string) ExecutionReceipt {
	msg := queued.Envelope
	receipt := ExecutionReceipt{
		Sequence:		queued.Sequence,
		TxHash:			ComputeQueueTxHash(queued),
		MessageID:		append([]byte(nil), queued.MessageID...),
		ExecutionKind:		executionKind(msg),
		ContractAddress:	append(sdk.AccAddress(nil), msg.Destination...),
		Caller:			append(sdk.AccAddress(nil), msg.Source...),
		Source:			append(sdk.AccAddress(nil), msg.Source...),
		Destination:		append(sdk.AccAddress(nil), msg.Destination...),
		Opcode:			msg.Opcode,
		QueryID:		msg.QueryID,
		GasLimit:		msg.GasLimit,
		GasUsed:		0,
		StorageFeeNaet:		sdkmath.ZeroInt(),
		ForwardFeeNaet:		msg.ForwardFee.Amount,
		FeeChargedNaet:		sdkmath.ZeroInt(),
		ValueInNaet:		normalizeReceiptInt(msg.Value.Amount),
		ValueOutNaet:		sdkmath.ZeroInt(),
		StateRootBefore:	stateRootBefore,
		StateRootAfter:		stateRootBefore,
		Bounced:		msg.Bounced,
		RefundAmountNaet:	sdkmath.ZeroInt(),
		RefundFeeNaet:		sdkmath.ZeroInt(),
		RetryCount:		msg.RetryCount,
		Height:			height,
		LogicalTime:		msg.CreatedLogicalTime,
	}
	if receipt.StateRootBefore == "" {
		receipt.StateRootBefore = EmptyAVMStateRoot()
		receipt.StateRootAfter = receipt.StateRootBefore
	}
	return receipt
}

func finalizeReceipt(receipt *ExecutionReceipt) {
	if receipt == nil {
		return
	}
	if receipt.ExitCode == 0 && receipt.ResultCode != 0 {
		receipt.ExitCode = receipt.ResultCode
	}
	if receipt.ResultCode == ResultOK {
		receipt.ExitCode = ResultOK
	}
	if receipt.ExitReason == "" {
		receipt.ExitReason = receipt.Error
	}
	if receipt.ExitReason == "" && receipt.ResultCode == ResultOK {
		receipt.ExitReason = "ok"
	}
	if receipt.QueueStatus == "" {
		receipt.QueueStatus = receiptQueueStatus(*receipt)
	}
	receipt.StorageFeeNaet = normalizeReceiptInt(receipt.StorageFeeNaet)
	receipt.ForwardFeeNaet = normalizeReceiptInt(receipt.ForwardFeeNaet)
	receipt.RefundAmountNaet = normalizeReceiptInt(receipt.RefundAmountNaet)
	receipt.RefundFeeNaet = normalizeReceiptInt(receipt.RefundFeeNaet)
	receipt.ValueInNaet = normalizeReceiptInt(receipt.ValueInNaet)
	receipt.ValueOutNaet = normalizeReceiptInt(receipt.ValueOutNaet)
	receipt.FeeChargedNaet = receipt.StorageFeeNaet.Add(receipt.ForwardFeeNaet)
	if receipt.StateRootBefore == "" {
		receipt.StateRootBefore = EmptyAVMStateRoot()
	}
	if receipt.StateRootAfter == "" {
		receipt.StateRootAfter = receipt.StateRootBefore
	}
	if !hasExecutionEvent(*receipt) {
		event := executionEvent(*receipt)
		receipt.Events = append([]AVMEvent{event}, receipt.Events...)
	}
	receipt.ReceiptID = ComputeExecutionReceiptID(*receipt)
}

func ComputeQueueTxHash(queued QueuedMessage) string {
	var buf bytes.Buffer
	buf.WriteString("aetra-avm-queue-tx-v1")
	writeU64(&buf, queued.TxHeight)
	writeU64(&buf, queued.TxIndex)
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func ComputeExecutionReceiptID(receipt ExecutionReceipt) string {
	var buf bytes.Buffer
	buf.WriteString("aetra-avm-execution-receipt-v1")
	writeU64(&buf, receipt.Sequence)
	writeString(&buf, receipt.TxHash)
	writeBytes(&buf, receipt.MessageID)
	writeString(&buf, receipt.ExecutionKind)
	writeAddress(&buf, receipt.ContractAddress)
	writeAddress(&buf, receipt.Caller)
	writeAddress(&buf, receipt.Source)
	writeAddress(&buf, receipt.Destination)
	writeU32(&buf, receipt.Opcode)
	writeU64(&buf, receipt.QueryID)
	writeU64(&buf, receipt.GasLimit)
	writeU64(&buf, receipt.GasUsed)
	writeString(&buf, normalizeReceiptInt(receipt.FeeChargedNaet).String())
	writeString(&buf, normalizeReceiptInt(receipt.ValueInNaet).String())
	writeString(&buf, normalizeReceiptInt(receipt.ValueOutNaet).String())
	writeU32(&buf, receipt.ExitCode)
	writeString(&buf, receipt.ExitReason)
	writeString(&buf, receipt.FailedPhase)
	writeString(&buf, receipt.StateRootBefore)
	writeString(&buf, receipt.StateRootAfter)
	writeU32(&buf, uint32(len(receipt.EmittedMessageIDs)))
	for _, id := range receipt.EmittedMessageIDs {
		writeBytes(&buf, id)
	}
	writeU32(&buf, uint32(len(receipt.Events)))
	for _, event := range receipt.Events {
		writeString(&buf, event.Type)
		writeU32(&buf, uint32(len(event.Attributes)))
		for _, attr := range event.Attributes {
			writeString(&buf, attr.Key)
			writeString(&buf, attr.Value)
		}
	}
	writeBool(&buf, receipt.BounceCreated)
	writeString(&buf, normalizeReceiptInt(receipt.RefundAmountNaet).String())
	writeU64(&buf, receipt.Height)
	writeU64(&buf, receipt.LogicalTime)
	writeBool(&buf, receipt.StateCommitted)
	writeString(&buf, receipt.QueueStatus)
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func ValidateExecutionReceipt(receipt ExecutionReceipt) error {
	if strings.TrimSpace(receipt.ReceiptID) == "" || len(receipt.ReceiptID) != ReceiptIDLength {
		return errors.New("AVM receipt id must be 32-byte lowercase hex")
	}
	if receipt.ReceiptID != strings.ToLower(receipt.ReceiptID) {
		return errors.New("AVM receipt id must be lowercase")
	}
	if _, err := hex.DecodeString(receipt.ReceiptID); err != nil {
		return err
	}
	if _, err := hex.DecodeString(receipt.TxHash); len(receipt.TxHash) != 64 || err != nil {
		return errors.New("AVM receipt tx hash must be 32-byte lowercase hex")
	}
	if len(receipt.MessageID) != MessageIDLength {
		return fmt.Errorf("AVM receipt message id must be %d bytes", MessageIDLength)
	}
	if !validExecutionKind(receipt.ExecutionKind) {
		return fmt.Errorf("AVM receipt invalid execution kind %q", receipt.ExecutionKind)
	}
	if len(receipt.ContractAddress) == 0 || len(receipt.Caller) == 0 || len(receipt.Destination) == 0 {
		return errors.New("AVM receipt addresses are required")
	}
	if receipt.GasLimit == 0 {
		return errors.New("AVM receipt gas limit must be positive")
	}
	if receipt.GasUsed > receipt.GasLimit {
		return errors.New("AVM receipt gas used exceeds gas limit")
	}
	for _, item := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "fee charged", value: receipt.FeeChargedNaet},
		{name: "value in", value: receipt.ValueInNaet},
		{name: "value out", value: receipt.ValueOutNaet},
		{name: "refund amount", value: receipt.RefundAmountNaet},
	} {
		if item.value.IsNil() || item.value.IsNegative() {
			return fmt.Errorf("AVM receipt %s must be non-negative", item.name)
		}
	}
	if receipt.StateRootBefore == "" || receipt.StateRootAfter == "" {
		return errors.New("AVM receipt state roots are required")
	}
	if !validQueueStatus(receipt.QueueStatus) {
		return fmt.Errorf("AVM receipt invalid queue status %q", receipt.QueueStatus)
	}
	if receipt.Height == 0 {
		return errors.New("AVM receipt height must be positive")
	}
	for _, id := range receipt.EmittedMessageIDs {
		if len(id) != MessageIDLength {
			return fmt.Errorf("AVM receipt emitted message id must be %d bytes", MessageIDLength)
		}
	}
	for _, event := range receipt.Events {
		if err := event.Validate(); err != nil {
			return err
		}
	}
	expected := ComputeExecutionReceiptID(receipt)
	if receipt.ReceiptID != expected {
		return errors.New("AVM receipt id mismatch")
	}
	return nil
}

func (e AVMEvent) Validate() error {
	if !validAVMEventType(e.Type) {
		return fmt.Errorf("invalid AVM event type %q", e.Type)
	}
	seen := map[string]struct{}{}
	for _, attr := range e.Attributes {
		if strings.TrimSpace(attr.Key) == "" {
			return errors.New("AVM event attribute key is required")
		}
		if _, found := seen[attr.Key]; found {
			return fmt.Errorf("duplicate AVM event attribute %q", attr.Key)
		}
		seen[attr.Key] = struct{}{}
	}
	return nil
}

func NewAVMEvent(eventType string, attrs ...AVMEventAttribute) AVMEvent {
	return AVMEvent{Type: eventType, Attributes: append([]AVMEventAttribute(nil), attrs...)}
}

func EventAttr(key, value string) AVMEventAttribute {
	return AVMEventAttribute{Key: key, Value: value}
}

func executionEvent(receipt ExecutionReceipt) AVMEvent {
	eventType := EventInternalExecuted
	if receipt.ExecutionKind == ExecutionKindExternal {
		eventType = EventExternalExecuted
	}
	return NewAVMEvent(eventType,
		EventAttr("message_id", hex.EncodeToString(receipt.MessageID)),
		EventAttr("contract", queueAddressKey(receipt.ContractAddress)),
		EventAttr("caller", queueAddressKey(receipt.Caller)),
		EventAttr("destination", queueAddressKey(receipt.Destination)),
		EventAttr("opcode", fmt.Sprintf("%d", receipt.Opcode)),
		EventAttr("query_id", fmt.Sprintf("%d", receipt.QueryID)),
		EventAttr("exit_code", fmt.Sprintf("%d", receipt.ExitCode)),
		EventAttr("gas_used", fmt.Sprintf("%d", receipt.GasUsed)),
		EventAttr("height", fmt.Sprintf("%d", receipt.Height)),
		EventAttr("state_committed", fmt.Sprintf("%t", receipt.StateCommitted)),
	)
}

func hasExecutionEvent(receipt ExecutionReceipt) bool {
	for _, event := range receipt.Events {
		if event.Type == EventInternalExecuted || event.Type == EventExternalExecuted {
			return true
		}
	}
	return false
}

func messageQueuedEvent(queued QueuedMessage) AVMEvent {
	return NewAVMEvent(EventMessageQueued,
		EventAttr("message_id", hex.EncodeToString(queued.MessageID)),
		EventAttr("source", queueAddressKey(queued.Envelope.Source)),
		EventAttr("destination", queueAddressKey(queued.Envelope.Destination)),
		EventAttr("opcode", fmt.Sprintf("%d", queued.Envelope.Opcode)),
		EventAttr("query_id", fmt.Sprintf("%d", queued.Envelope.QueryID)),
		EventAttr("scheduled_height", fmt.Sprintf("%d", queued.ScheduledHeight)),
		EventAttr("sequence", fmt.Sprintf("%d", queued.Sequence)),
	)
}

func messageBouncedEvent(receipt ExecutionReceipt, queued QueuedMessage) AVMEvent {
	return NewAVMEvent(EventMessageBounced,
		EventAttr("original_message_id", hex.EncodeToString(receipt.MessageID)),
		EventAttr("bounce_message_id", hex.EncodeToString(queued.MessageID)),
		EventAttr("source", queueAddressKey(queued.Envelope.Source)),
		EventAttr("destination", queueAddressKey(queued.Envelope.Destination)),
		EventAttr("refund_amount", normalizeReceiptInt(receipt.RefundAmountNaet).String()),
		EventAttr("height", fmt.Sprintf("%d", receipt.Height)),
	)
}

func contractFrozenEvent(contract ContractAccount, height uint64) AVMEvent {
	return NewAVMEvent(EventContractFrozen,
		EventAttr("contract", queueAddressKey(contract.Address)),
		EventAttr("storage_debt", normalizeReceiptInt(contract.StorageRentDebtNaet).String()),
		EventAttr("height", fmt.Sprintf("%d", height)),
	)
}

func ContractStateRoot(contract ContractAccount) string {
	var buf bytes.Buffer
	buf.WriteString("aetra-avm-contract-state-v1")
	writeAddress(&buf, contract.Address)
	writeBytes(&buf, contract.CodeHash)
	writeBytes(&buf, contract.State)
	writeString(&buf, normalizeReceiptInt(contract.BalanceNaet).String())
	writeU64(&buf, contract.LogicalTime)
	writeString(&buf, contract.NormalizedStatus())
	writeString(&buf, normalizeReceiptInt(contract.StorageRentDebtNaet).String())
	writeU64(&buf, contract.LastStorageChargeHeight)
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func EmptyAVMStateRoot() string {
	sum := sha256.Sum256([]byte("aetra-avm-empty-state-v1"))
	return hex.EncodeToString(sum[:])
}

func (e *Executor) QueryReceipts(contract sdk.AccAddress, limit uint32) ([]ExecutionReceipt, error) {
	if limit == 0 {
		return nil, errors.New("receipt query limit must be positive")
	}
	out := make([]ExecutionReceipt, 0, limit)
	for _, receipt := range e.receipts {
		if len(contract) != 0 && !receipt.ContractAddress.Equals(contract) {
			continue
		}
		out = append(out, cloneReceipt(receipt))
		if uint32(len(out)) == limit {
			break
		}
	}
	return out, nil
}

func validateReceipts(receipts []ExecutionReceipt) error {
	seen := map[string]struct{}{}
	last := ""
	for _, receipt := range receipts {
		if err := ValidateExecutionReceipt(receipt); err != nil {
			return err
		}
		if _, found := seen[receipt.ReceiptID]; found {
			return errors.New("duplicate AVM receipt")
		}
		seen[receipt.ReceiptID] = struct{}{}
		key := receiptSortKey(receipt)
		if last != "" && key < last {
			return errors.New("AVM receipts must be sorted canonically")
		}
		last = key
	}
	return nil
}

func receiptSortKey(receipt ExecutionReceipt) string {
	return fmt.Sprintf("%020d/%020d/%s", receipt.Height, receipt.Sequence, receipt.ReceiptID)
}

func normalizeReceiptInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}

func validExecutionKind(kind string) bool {
	switch kind {
	case ExecutionKindExternal, ExecutionKindInternal, ExecutionKindBounced, ExecutionKindSystem:
		return true
	default:
		return false
	}
}

func executionKind(msg MessageEnvelope) string {
	if msg.Bounced {
		return ExecutionKindBounced
	}
	if msg.Opcode == RefundOpcode {
		return ExecutionKindSystem
	}
	return ExecutionKindInternal
}

func validAVMEventType(eventType string) bool {
	switch eventType {
	case EventCodeStored, EventContractDeployed, EventExternalExecuted, EventInternalExecuted,
		EventMessageQueued, EventMessageBounced, EventContractFrozen, EventContractUnfrozen, EventRentPaid:
		return true
	default:
		return false
	}
}
