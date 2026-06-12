package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
)

func ComputeMessageID(msg MeshMessage) string {
	h := sha256.New()
	writeString(h, "aetra-mesh-message-v1")
	writeString(h, string(msg.SourceZone))
	writeString(h, string(msg.SourceShard))
	writeString(h, string(msg.DestinationZone))
	writeString(h, string(msg.DestinationShard))
	writeUint64(h, msg.Nonce)
	writeString(h, msg.PayloadHash)
	writeUint64(h, msg.SourceLogicalTime)
	return hex.EncodeToString(h.Sum(nil))
}

func BuildProof(msg MeshMessage, commitment FinalizedCommitment) MeshProof {
	msg = msg.Normalize()
	return MeshProof{
		SourceCommitment:	commitment.CommitmentHash,
		MessageRoot:		commitment.MessageRoot,
		ProofHash:		ComputeProofHash(msg, commitment),
	}
}

func ComputeProofHash(msg MeshMessage, commitment FinalizedCommitment) string {
	msg = msg.Normalize()
	h := sha256.New()
	writeString(h, "aetra-mesh-proof-v1")
	writeString(h, msg.MessageID)
	writeString(h, string(msg.SourceZone))
	writeString(h, string(msg.SourceShard))
	writeUint64(h, msg.Finality.Height)
	writeString(h, commitment.CommitmentHash)
	writeString(h, commitment.MessageRoot)
	writeUint64(h, msg.Sequence)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeReceiptHash(receipt MeshReceipt) string {
	h := sha256.New()
	writeString(h, "aetra-mesh-receipt-v1")
	writeString(h, receipt.MessageID)
	writeString(h, string(receipt.SourceZone))
	writeString(h, string(receipt.SourceShard))
	writeString(h, string(receipt.DestinationZone))
	writeString(h, string(receipt.DestinationShard))
	writeString(h, string(receipt.Status))
	writeString(h, string(receipt.Reason))
	writeUint64(h, receipt.Height)
	writeUint64(h, receipt.Sequence)
	writeUint64(h, uint64(receipt.ExecutionCode))
	writeString(h, receipt.ResultHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeBounceReceiptHash(receipt BounceReceipt) string {
	h := sha256.New()
	writeString(h, "aetra-mesh-bounce-receipt-v1")
	writeString(h, receipt.MessageID)
	writeString(h, receipt.SourceMessageID)
	writeString(h, receipt.BounceMessageID)
	writeString(h, string(receipt.DestinationZone))
	writeString(h, string(receipt.DestinationShard))
	writeString(h, string(receipt.Reason))
	writeUint64(h, receipt.Height)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeRefundReceiptHash(receipt RefundReceipt) string {
	h := sha256.New()
	writeString(h, "aetra-mesh-refund-receipt-v1")
	writeString(h, receipt.MessageID)
	writeString(h, receipt.SourceMessageID)
	writeBytes(h, receipt.Recipient)
	writeString(h, receipt.AssetCommitment)
	writeString(h, string(receipt.Reason))
	writeUint64(h, receipt.Height)
	return hex.EncodeToString(h.Sum(nil))
}

func HashParts(parts ...string) string {
	h := sha256.New()
	writeString(h, "aetra-mesh-hash-parts-v1")
	for _, part := range parts {
		writeString(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func CompareMessages(left, right MeshMessage) int {
	left = left.Normalize()
	right = right.Normalize()
	if left.Finality.Height != right.Finality.Height {
		return compareUint64(left.Finality.Height, right.Finality.Height)
	}
	if left.SourceZone != right.SourceZone {
		return compareString(string(left.SourceZone), string(right.SourceZone))
	}
	if left.SourceShard != right.SourceShard {
		return compareString(string(left.SourceShard), string(right.SourceShard))
	}
	if left.MessageID != right.MessageID {
		return compareString(left.MessageID, right.MessageID)
	}
	if left.DestinationZone != right.DestinationZone {
		return compareString(string(left.DestinationZone), string(right.DestinationZone))
	}
	if left.DestinationShard != right.DestinationShard {
		return compareString(string(left.DestinationShard), string(right.DestinationShard))
	}
	return compareUint64(left.Sequence, right.Sequence)
}

func SortMessages(messages []MeshMessage) []MeshMessage {
	out := cloneMessages(messages)
	sort.SliceStable(out, func(i, j int) bool {
		return CompareMessages(out[i], out[j]) < 0
	})
	return out
}

func writeString(w interface{ Write([]byte) (int, error) }, value string) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write([]byte(value))
}

func writeBytes(w interface{ Write([]byte) (int, error) }, value []byte) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write(value)
}

func writeUint64(w interface{ Write([]byte) (int, error) }, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}

func cloneMessage(msg MeshMessage) MeshMessage {
	msg = msg.Normalize()
	msg.Sender = cloneBytes(msg.Sender)
	msg.Recipient = cloneBytes(msg.Recipient)
	return msg
}

func cloneMessages(in []MeshMessage) []MeshMessage {
	out := make([]MeshMessage, len(in))
	for i, msg := range in {
		out[i] = cloneMessage(msg)
	}
	return out
}

func cloneBytes(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}
	return append([]byte(nil), in...)
}

func isZeroBytes(bz []byte) bool {
	if len(bz) == 0 {
		return false
	}
	for _, b := range bz {
		if b != 0 {
			return false
		}
	}
	return true
}

func compareString(left, right string) int {
	return bytes.Compare([]byte(left), []byte(right))
}

func compareUint64(left, right uint64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
