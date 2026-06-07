package async

import (
	"errors"
	"fmt"
	"sort"
)

func (e *Executor) EnqueueTxMessages(messages []MessageEnvelope) error {
	if len(messages) == 0 {
		return errors.New("tx message count must be positive")
	}
	if len(messages) > int(e.params.MaxMessagesPerTx) {
		return fmt.Errorf("messages per tx must be <= %d", e.params.MaxMessagesPerTx)
	}
	txIndex := e.nextTxIndex
	e.nextTxIndex++
	for i, msg := range messages {
		if err := e.enqueueMessageWithOrder(msg, txIndex, uint32(i)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) EnqueueMessage(msg MessageEnvelope) error {
	txIndex := e.nextTxIndex
	e.nextTxIndex++
	return e.enqueueMessageWithOrder(msg, txIndex, 0)
}

func (e *Executor) enqueueMessageWithOrder(msg MessageEnvelope, txIndex uint64, messageIndex uint32) error {
	msg.ExecutionBlockHeight = 0
	if err := msg.Validate(e.params); err != nil {
		return err
	}
	queued := QueuedMessage{
		TxIndex:           txIndex,
		MessageIndex:      messageIndex,
		SourceLogicalTime: msg.CreatedLogicalTime,
		DestinationKey:    string(msg.Destination),
		Sequence:          e.nextSequence,
		EnqueuedBlock:     e.blockHeight,
		Envelope:          cloneMessage(msg),
	}
	e.nextSequence++
	e.queue = append(e.queue, queued)
	sort.SliceStable(e.queue, func(i, j int) bool {
		return queuedMessageLess(e.queue[i], e.queue[j])
	})
	destinationKey := string(msg.Destination)
	sourceKey := string(msg.Source)
	e.inbox[destinationKey] = append(e.inbox[destinationKey], queued)
	e.outbox[sourceKey] = append(e.outbox[sourceKey], queued)
	sort.SliceStable(e.inbox[destinationKey], func(i, j int) bool {
		return queuedMessageLess(e.inbox[destinationKey][i], e.inbox[destinationKey][j])
	})
	sort.SliceStable(e.outbox[sourceKey], func(i, j int) bool {
		return queuedMessageLess(e.outbox[sourceKey][i], e.outbox[sourceKey][j])
	})
	e.metrics.QueuedMessages++
	return nil
}

func (e *Executor) ProcessBlock(height uint64) ([]ExecutionReceipt, error) {
	e.blockHeight = height
	e.deploysInBlock = 0
	if e.params.MaxMessagesPerBlock == 0 {
		return nil, errors.New("max messages per block must be positive")
	}
	count := uint32(0)
	receipts := make([]ExecutionReceipt, 0)
	for len(e.queue) > 0 && count < e.params.MaxMessagesPerBlock {
		if readyBlock(e.queue[0]) > height {
			break
		}
		receipt, err := e.processNext()
		if err != nil {
			return receipts, err
		}
		receipts = append(receipts, receipt)
		count++
	}
	e.updateQueueLag()
	return receipts, nil
}

func (e *Executor) updateQueueLag() {
	if len(e.queue) == 0 {
		e.metrics.QueueLag = 0
		return
	}
	oldest := e.queue[0].EnqueuedBlock
	if readyBlock(e.queue[0]) > e.blockHeight {
		e.metrics.QueueLag = 0
		return
	}
	if e.blockHeight > oldest {
		e.metrics.QueueLag = e.blockHeight - oldest
		return
	}
	e.metrics.QueueLag = 0
}

func queuedMessageLess(a, b QueuedMessage) bool {
	if readyBlock(a) != readyBlock(b) {
		return readyBlock(a) < readyBlock(b)
	}
	if a.TxIndex != b.TxIndex {
		return a.TxIndex < b.TxIndex
	}
	if a.MessageIndex != b.MessageIndex {
		return a.MessageIndex < b.MessageIndex
	}
	if a.SourceLogicalTime != b.SourceLogicalTime {
		return a.SourceLogicalTime < b.SourceLogicalTime
	}
	if a.DestinationKey != b.DestinationKey {
		return a.DestinationKey < b.DestinationKey
	}
	return a.Sequence < b.Sequence
}

func readyBlock(msg QueuedMessage) uint64 {
	return msg.Envelope.DeliverAtBlock
}
