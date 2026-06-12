package async

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const MessageIDLength = 32

func buildQueuedMessage(msg MessageEnvelope, txHeight, txIndex uint64, messageIndex uint32, sequence uint64) QueuedMessage {
	msg = cloneMessage(msg)
	msg.ExecutionBlockHeight = 0
	queued := QueuedMessage{
		TxHeight:		txHeight,
		TxIndex:		txIndex,
		MessageIndex:		messageIndex,
		SourceLogicalTime:	msg.CreatedLogicalTime,
		DestinationKey:		queueAddressKey(msg.Destination),
		Sequence:		sequence,
		EnqueuedBlock:		txHeight,
		CreatedHeight:		txHeight,
		ScheduledHeight:	scheduledHeight(msg),
		Attempts:		processingAttempts(msg),
		Status:			QueueStatusPending,
		Envelope:		msg,
	}
	queued.MessageID = QueueMessageID(queued)
	return queued
}

func QueueMessageID(queued QueuedMessage) []byte {
	msg := queued.Envelope
	var buf bytes.Buffer
	buf.WriteString("aetra-avm-queue-message-v1")
	writeU64(&buf, queued.ScheduledHeight)
	writeU64(&buf, queued.TxHeight)
	writeU64(&buf, queued.TxIndex)
	writeU32(&buf, queued.MessageIndex)
	writeU64(&buf, queued.SourceLogicalTime)
	writeU64(&buf, queued.Sequence)
	writeString(&buf, queueAddressKey(msg.Destination))
	writeAddress(&buf, msg.Source)
	writeAddress(&buf, msg.Destination)
	writeString(&buf, msg.Value.Denom)
	writeString(&buf, msg.Value.Amount.String())
	writeU32(&buf, msg.Opcode)
	writeU64(&buf, msg.QueryID)
	writeBytes(&buf, msg.Body)
	writeBool(&buf, msg.Bounce)
	writeBool(&buf, msg.Bounced)
	writeU64(&buf, msg.DeadlineBlock)
	writeU64(&buf, msg.GasLimit)
	writeU32(&buf, msg.RetryCount)
	writeU32(&buf, msg.MaxRetries)
	writeU64(&buf, msg.RetryDelayBlocks)
	writeU32(&buf, msg.Depth)
	writeU64(&buf, msg.RefundOfSequence)
	sum := sha256.Sum256(buf.Bytes())
	return sum[:]
}

func validateQueuedMessage(queued QueuedMessage, params Params) error {
	if err := queued.Envelope.Validate(params); err != nil {
		return err
	}
	if queued.TxHeight != queued.EnqueuedBlock {
		return fmt.Errorf("queued message %d tx height drift", queued.Sequence)
	}
	if queued.CreatedHeight != queued.EnqueuedBlock {
		return fmt.Errorf("queued message %d created height drift", queued.Sequence)
	}
	if queued.ScheduledHeight != scheduledHeight(queued.Envelope) {
		return fmt.Errorf("queued message %d scheduled height drift", queued.Sequence)
	}
	if queued.Attempts != processingAttempts(queued.Envelope) {
		return fmt.Errorf("queued message %d attempts drift", queued.Sequence)
	}
	if queued.Attempts == 0 || queued.Attempts > params.MaxProcessingAttempts {
		return fmt.Errorf("queued message %d attempts must be between 1 and %d", queued.Sequence, params.MaxProcessingAttempts)
	}
	if queued.Status != QueueStatusPending {
		return fmt.Errorf("queued message %d status must be pending", queued.Sequence)
	}
	if !validQueueStatus(queued.Status) {
		return fmt.Errorf("queued message %d invalid status", queued.Sequence)
	}
	if queued.SourceLogicalTime != queued.Envelope.CreatedLogicalTime {
		return fmt.Errorf("queued message %d source logical time drift", queued.Sequence)
	}
	if queued.DestinationKey != queueAddressKey(queued.Envelope.Destination) {
		return fmt.Errorf("queued message %d destination key drift", queued.Sequence)
	}
	if queued.Envelope.ExecutionBlockHeight != 0 {
		return fmt.Errorf("queued message %d execution block height must be zero", queued.Sequence)
	}
	if len(queued.MessageID) != MessageIDLength {
		return fmt.Errorf("queued message id must be %d bytes", MessageIDLength)
	}
	if !bytes.Equal(queued.MessageID, QueueMessageID(queued)) {
		return fmt.Errorf("queued message %d message id mismatch", queued.Sequence)
	}
	return nil
}

func validQueueStatus(status string) bool {
	switch status {
	case QueueStatusPending, QueueStatusExecuted, QueueStatusFailed, QueueStatusExpired, QueueStatusBounced:
		return true
	default:
		return false
	}
}

func receiptQueueStatus(receipt ExecutionReceipt) string {
	switch receipt.ResultCode {
	case ResultOK:
		return QueueStatusExecuted
	case ResultExpired:
		return QueueStatusExpired
	default:
		if receipt.BounceCreated || receipt.Bounced {
			return QueueStatusBounced
		}
		return QueueStatusFailed
	}
}

func processingAttempts(msg MessageEnvelope) uint32 {
	return msg.RetryCount + 1
}

func scheduledHeight(msg MessageEnvelope) uint64 {
	return msg.DeliverAtBlock
}

func readyBlock(msg QueuedMessage) uint64 {
	return msg.ScheduledHeight
}

func queueAddressKey(addr sdk.AccAddress) string {
	return addressing.Format(addr)
}

func inboxKey(addr sdk.AccAddress) string {
	return queueAddressKey(addr)
}

func outboxKey(addr sdk.AccAddress) string {
	return queueAddressKey(addr)
}

func queuedMessageLess(a, b QueuedMessage) bool {
	if a.ScheduledHeight != b.ScheduledHeight {
		return a.ScheduledHeight < b.ScheduledHeight
	}
	if a.TxHeight != b.TxHeight {
		return a.TxHeight < b.TxHeight
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
	if a.Sequence != b.Sequence {
		return a.Sequence < b.Sequence
	}
	return a.DestinationKey < b.DestinationKey
}

func writeAddress(buf *bytes.Buffer, addr sdk.AccAddress) {
	writeString(buf, queueAddressKey(addr))
}

func writeBool(buf *bytes.Buffer, value bool) {
	if value {
		buf.WriteByte(1)
		return
	}
	buf.WriteByte(0)
}

func writeString(buf *bytes.Buffer, value string) {
	writeBytes(buf, []byte(value))
}

func writeBytes(buf *bytes.Buffer, value []byte) {
	writeU32(buf, uint32(len(value)))
	buf.Write(value)
}

func writeU32(buf *bytes.Buffer, value uint32) {
	var out [4]byte
	binary.BigEndian.PutUint32(out[:], value)
	buf.Write(out[:])
}

func writeU64(buf *bytes.Buffer, value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	buf.Write(out[:])
}

func validateMessageBatch(params Params, messages []MessageEnvelope) error {
	if len(messages) == 0 {
		return errors.New("tx message count must be positive")
	}
	if len(messages) > int(params.MaxMessagesPerTx) {
		return fmt.Errorf("messages per tx must be <= %d", params.MaxMessagesPerTx)
	}
	for _, msg := range messages {
		if err := msg.Validate(params); err != nil {
			return err
		}
		if processingAttempts(msg) > params.MaxProcessingAttempts {
			return fmt.Errorf("message processing attempts must be <= %d", params.MaxProcessingAttempts)
		}
	}
	return nil
}
