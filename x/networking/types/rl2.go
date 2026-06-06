package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxRL2Priority = uint32(6)
)

type RL2PayloadType string

const (
	RL2PayloadLargeBlock      RL2PayloadType = "large_block"
	RL2PayloadBlockChunk      RL2PayloadType = "block_chunk"
	RL2PayloadStateSyncStream RL2PayloadType = "state_sync_stream"
	RL2PayloadZoneSnapshot    RL2PayloadType = "zone_snapshot"
	RL2PayloadExecutionResult RL2PayloadType = "execution_result"
	RL2PayloadStorageObject   RL2PayloadType = "storage_object"
	RL2PayloadProofSet        RL2PayloadType = "proof_set"
)

type RL2FECPolicy string

const (
	RL2FECNone        RL2FECPolicy = "none"
	RL2FECXORParity   RL2FECPolicy = "xor_parity"
	RL2FECReedSolomon RL2FECPolicy = "reed_solomon"
)

type RL2Transfer struct {
	TransferID     string
	SourceNode     string
	TargetNode     string
	PayloadType    RL2PayloadType
	PayloadRoot    string
	ChunkCount     uint32
	ChunkSize      uint64
	FECPolicy      RL2FECPolicy
	Priority       uint32
	DeadlineHeight uint64
	ResumeToken    string
}

type RL2TransferProgress struct {
	TransferID     string
	ReceivedChunks []uint32
	ResumeToken    string
}

func NewRL2Transfer(transfer RL2Transfer) (RL2Transfer, error) {
	transfer = transfer.Normalize()
	if transfer.TransferID == "" {
		transfer.TransferID = ComputeRL2TransferID(transfer)
	}
	if err := transfer.Validate(0); err != nil {
		return RL2Transfer{}, err
	}
	return transfer, nil
}

func NewRL2TransferFromChunks(sourceNode, targetNode string, payloadType RL2PayloadType, chunks []PayloadChunk, priority uint32, deadlineHeight uint64, fecPolicy RL2FECPolicy) (RL2Transfer, error) {
	if len(chunks) == 0 {
		return RL2Transfer{}, errors.New("networking RL2 transfer requires chunks")
	}
	if _, err := ReassemblePayload(chunks); err != nil {
		return RL2Transfer{}, err
	}
	first := chunks[0]
	maxChunkSize := uint64(0)
	for _, chunk := range chunks {
		if uint64(len(chunk.Bytes)) > maxChunkSize {
			maxChunkSize = uint64(len(chunk.Bytes))
		}
		if chunk.PayloadHash != first.PayloadHash || chunk.Total != first.Total {
			return RL2Transfer{}, errors.New("networking RL2 chunk set mismatch")
		}
	}
	return NewRL2Transfer(RL2Transfer{
		SourceNode:     sourceNode,
		TargetNode:     targetNode,
		PayloadType:    payloadType,
		PayloadRoot:    first.PayloadHash,
		ChunkCount:     first.Total,
		ChunkSize:      maxChunkSize,
		FECPolicy:      fecPolicy,
		Priority:       priority,
		DeadlineHeight: deadlineHeight,
	})
}

func (t RL2Transfer) Normalize() RL2Transfer {
	t.TransferID = strings.ToLower(strings.TrimSpace(t.TransferID))
	t.SourceNode = strings.ToLower(strings.TrimSpace(t.SourceNode))
	t.TargetNode = strings.ToLower(strings.TrimSpace(t.TargetNode))
	t.PayloadType = RL2PayloadType(strings.ToLower(strings.TrimSpace(string(t.PayloadType))))
	t.PayloadRoot = strings.ToLower(strings.TrimSpace(t.PayloadRoot))
	t.FECPolicy = RL2FECPolicy(strings.ToLower(strings.TrimSpace(string(t.FECPolicy))))
	if t.FECPolicy == "" {
		t.FECPolicy = RL2FECNone
	}
	t.ResumeToken = strings.ToLower(strings.TrimSpace(t.ResumeToken))
	return t
}

func (t RL2Transfer) ValidateBasic() error {
	return t.Validate(0)
}

func (t RL2Transfer) Validate(currentHeight uint64) error {
	transfer := t.Normalize()
	if err := ValidateHash("networking RL2 transfer id", transfer.TransferID); err != nil {
		return err
	}
	if transfer.TransferID != ComputeRL2TransferID(transfer) {
		return errors.New("networking RL2 transfer id mismatch")
	}
	if err := ValidateHash("networking RL2 source node", transfer.SourceNode); err != nil {
		return err
	}
	if err := ValidateHash("networking RL2 target node", transfer.TargetNode); err != nil {
		return err
	}
	if transfer.SourceNode == transfer.TargetNode {
		return errors.New("networking RL2 transfer source and target must differ")
	}
	if !IsRL2PayloadType(transfer.PayloadType) {
		return fmt.Errorf("unknown networking RL2 payload type %q", transfer.PayloadType)
	}
	if err := ValidateHash("networking RL2 payload root", transfer.PayloadRoot); err != nil {
		return err
	}
	if transfer.ChunkCount == 0 || transfer.ChunkCount > MaxPayloadChunks {
		return fmt.Errorf("networking RL2 chunk count must be between 1 and %d", MaxPayloadChunks)
	}
	if transfer.ChunkSize == 0 || transfer.ChunkSize > MaxChunkBytes {
		return fmt.Errorf("networking RL2 chunk size must be between 1 and %d", MaxChunkBytes)
	}
	if !IsRL2FECPolicy(transfer.FECPolicy) {
		return fmt.Errorf("unknown networking RL2 FEC policy %q", transfer.FECPolicy)
	}
	if transfer.Priority > MaxRL2Priority {
		return fmt.Errorf("networking RL2 priority must be <= %d", MaxRL2Priority)
	}
	if transfer.DeadlineHeight > 0 && currentHeight > transfer.DeadlineHeight {
		return errors.New("networking RL2 transfer is expired")
	}
	if transfer.ResumeToken != "" {
		if err := ValidateHash("networking RL2 resume token", transfer.ResumeToken); err != nil {
			return err
		}
	}
	return nil
}

func (t RL2Transfer) TransportEnvelope(enqueuedHeight, sequence uint64) (TransportEnvelope, error) {
	transfer := t.Normalize()
	if err := transfer.Validate(enqueuedHeight); err != nil {
		return TransportEnvelope{}, err
	}
	envelope := TransportEnvelope{
		Channel:        RL2ChannelForPayloadType(transfer.PayloadType),
		SizeBytes:      transfer.ChunkSize,
		EnqueuedHeight: enqueuedHeight,
		Sequence:       sequence,
		PayloadHash:    transfer.PayloadRoot,
	}
	envelope = envelope.Normalize()
	if err := envelope.Validate(DefaultChannelPolicies()); err != nil {
		return TransportEnvelope{}, err
	}
	return envelope, nil
}

func ComputeRL2TransferID(transfer RL2Transfer) string {
	transfer = transfer.Normalize()
	return HashParts(
		"rl2-transfer",
		transfer.SourceNode,
		transfer.TargetNode,
		string(transfer.PayloadType),
		transfer.PayloadRoot,
		fmt.Sprintf("%d", transfer.ChunkCount),
		fmt.Sprintf("%d", transfer.ChunkSize),
		string(transfer.FECPolicy),
		fmt.Sprintf("%d", transfer.Priority),
		fmt.Sprintf("%d", transfer.DeadlineHeight),
	)
}

func ComputeRL2ResumeToken(transfer RL2Transfer, receivedChunks []uint32) (string, error) {
	transfer = transfer.Normalize()
	if err := transfer.Validate(0); err != nil {
		return "", err
	}
	normalized := append([]uint32(nil), receivedChunks...)
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i] < normalized[j]
	})
	parts := []string{"rl2-resume-token", transfer.TransferID}
	var previous uint32
	for i, index := range normalized {
		if index >= transfer.ChunkCount {
			return "", errors.New("networking RL2 resume chunk index out of range")
		}
		if i > 0 && index == previous {
			return "", errors.New("networking RL2 resume chunks must be unique")
		}
		previous = index
		parts = append(parts, fmt.Sprintf("%d", index))
	}
	return HashParts(parts...), nil
}

func NewRL2TransferProgress(transfer RL2Transfer, receivedChunks []uint32) (RL2TransferProgress, error) {
	transfer = transfer.Normalize()
	token, err := ComputeRL2ResumeToken(transfer, receivedChunks)
	if err != nil {
		return RL2TransferProgress{}, err
	}
	normalized := append([]uint32(nil), receivedChunks...)
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i] < normalized[j]
	})
	return RL2TransferProgress{
		TransferID:     transfer.TransferID,
		ReceivedChunks: normalized,
		ResumeToken:    token,
	}, nil
}

func PlanRL2Transfer(adapter AetherNetworkingAdapter, transfer RL2Transfer, enqueuedHeight, sequence uint64, peerCount uint32, score PeerScore) (PropagationPlan, error) {
	envelope, err := transfer.TransportEnvelope(enqueuedHeight, sequence)
	if err != nil {
		return PropagationPlan{}, err
	}
	return PlanPropagation(adapter, envelope, peerCount, score)
}

func RL2ChannelForPayloadType(payloadType RL2PayloadType) ChannelClass {
	switch payloadType {
	case RL2PayloadLargeBlock, RL2PayloadBlockChunk:
		return ChannelBlock
	case RL2PayloadStateSyncStream, RL2PayloadZoneSnapshot:
		return ChannelStateSync
	case RL2PayloadExecutionResult:
		return ChannelExecution
	case RL2PayloadStorageObject, RL2PayloadProofSet:
		return ChannelData
	default:
		return ""
	}
}

func DefaultRL2Priority(payloadType RL2PayloadType) uint32 {
	channel := RL2ChannelForPayloadType(payloadType)
	if channel == "" {
		return MaxRL2Priority
	}
	return PriorityForChannel(channel)
}

func IsRL2PayloadType(payloadType RL2PayloadType) bool {
	switch payloadType {
	case RL2PayloadLargeBlock,
		RL2PayloadBlockChunk,
		RL2PayloadStateSyncStream,
		RL2PayloadZoneSnapshot,
		RL2PayloadExecutionResult,
		RL2PayloadStorageObject,
		RL2PayloadProofSet:
		return true
	default:
		return false
	}
}

func IsRL2FECPolicy(policy RL2FECPolicy) bool {
	switch policy {
	case RL2FECNone, RL2FECXORParity, RL2FECReedSolomon:
		return true
	default:
		return false
	}
}
