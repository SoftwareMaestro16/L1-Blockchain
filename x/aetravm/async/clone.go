package async

import sdk "github.com/cosmos/cosmos-sdk/types"

func cloneContract(contract ContractAccount) ContractAccount {
	return ContractAccount{
		Address:			append(sdk.AccAddress(nil), contract.Address...),
		CodeHash:			append([]byte(nil), contract.CodeHash...),
		State:				append([]byte(nil), contract.State...),
		BalanceNaet:			contract.BalanceNaet,
		LogicalTime:			contract.LogicalTime,
		Status:				contract.Status,
		StorageRentDebtNaet:		contract.StorageRentDebtNaet,
		LastStorageChargeHeight:	contract.LastStorageChargeHeight,
	}
}

func cloneMessage(msg MessageEnvelope) MessageEnvelope {
	msg.Source = append(sdk.AccAddress(nil), msg.Source...)
	msg.Destination = append(sdk.AccAddress(nil), msg.Destination...)
	msg.Body = append([]byte(nil), msg.Body...)
	return msg
}

func cloneQueuedMessages(messages []QueuedMessage) []QueuedMessage {
	if len(messages) == 0 {
		return nil
	}
	out := make([]QueuedMessage, len(messages))
	for i, msg := range messages {
		out[i] = QueuedMessage{
			MessageID:		append([]byte(nil), msg.MessageID...),
			TxHeight:		msg.TxHeight,
			TxIndex:		msg.TxIndex,
			MessageIndex:		msg.MessageIndex,
			SourceLogicalTime:	msg.SourceLogicalTime,
			DestinationKey:		msg.DestinationKey,
			Sequence:		msg.Sequence,
			EnqueuedBlock:		msg.EnqueuedBlock,
			CreatedHeight:		msg.CreatedHeight,
			ScheduledHeight:	msg.ScheduledHeight,
			Attempts:		msg.Attempts,
			Status:			msg.Status,
			Envelope:		cloneMessage(msg.Envelope),
		}
	}
	return out
}

func cloneQueuedMap(in map[string][]QueuedMessage) map[string][]QueuedMessage {
	out := make(map[string][]QueuedMessage, len(in))
	for key, value := range in {
		out[key] = cloneQueuedMessages(value)
	}
	return out
}

func cloneDeadLetter(dead DeadLetter) DeadLetter {
	dead.Envelope = cloneMessage(dead.Envelope)
	dead.Receipt = cloneReceipt(dead.Receipt)
	return dead
}

func cloneDeadLetters(deadLetters []DeadLetter) []DeadLetter {
	if len(deadLetters) == 0 {
		return nil
	}
	out := make([]DeadLetter, len(deadLetters))
	for i, dead := range deadLetters {
		out[i] = cloneDeadLetter(dead)
	}
	return out
}

func cloneReceipt(receipt ExecutionReceipt) ExecutionReceipt {
	receipt.MessageID = append([]byte(nil), receipt.MessageID...)
	receipt.ContractAddress = append(sdk.AccAddress(nil), receipt.ContractAddress...)
	receipt.Caller = append(sdk.AccAddress(nil), receipt.Caller...)
	receipt.Source = append(sdk.AccAddress(nil), receipt.Source...)
	receipt.Destination = append(sdk.AccAddress(nil), receipt.Destination...)
	receipt.EmittedMessageIDs = cloneByteSlices(receipt.EmittedMessageIDs)
	receipt.Events = cloneEvents(receipt.Events)
	return receipt
}

func cloneReceipts(receipts []ExecutionReceipt) []ExecutionReceipt {
	if len(receipts) == 0 {
		return nil
	}
	out := make([]ExecutionReceipt, len(receipts))
	for i, receipt := range receipts {
		out[i] = cloneReceipt(receipt)
	}
	return out
}

func cloneByteSlices(values [][]byte) [][]byte {
	if len(values) == 0 {
		return nil
	}
	out := make([][]byte, len(values))
	for i, value := range values {
		out[i] = append([]byte(nil), value...)
	}
	return out
}

func cloneEvents(values []AVMEvent) []AVMEvent {
	if len(values) == 0 {
		return nil
	}
	out := make([]AVMEvent, len(values))
	for i, event := range values {
		out[i] = AVMEvent{
			Type:		event.Type,
			Attributes:	append([]AVMEventAttribute(nil), event.Attributes...),
		}
	}
	return out
}
