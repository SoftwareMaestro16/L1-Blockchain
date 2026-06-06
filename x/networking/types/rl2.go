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
	VerifiedBitmap []bool
	ResumeToken    string
}

type RL2ChunkDescriptor struct {
	TransferID string
	ChunkIndex uint32
	ChunkHash  string
	ChunkSize  uint64
	RangeStart uint64
	RangeEnd   uint64
	ProofPath  []string
}

type RL2ChunkRequest struct {
	TransferID     string
	MissingIndexes []uint32
	ResumeToken    string
}

type RL2StreamingPlan struct {
	TransferID         string
	Channel            ChannelClass
	PriorityLane       uint32
	ParallelStreams    uint32
	MaxInFlightChunks  uint32
	ChunkBudgetBytes   uint64
	BackpressureActive bool
	FECEnabled         bool
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
	ordered, err := orderedRL2Chunks(chunks)
	if err != nil {
		return RL2Transfer{}, err
	}
	payloadRoot, err := ComputeRL2ChunkRoot(ordered)
	if err != nil {
		return RL2Transfer{}, err
	}
	first := ordered[0]
	maxChunkSize := uint64(0)
	for _, chunk := range ordered {
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
		PayloadRoot:    payloadRoot,
		ChunkCount:     first.Total,
		ChunkSize:      maxChunkSize,
		FECPolicy:      fecPolicy,
		Priority:       priority,
		DeadlineHeight: deadlineHeight,
	})
}

func ComputeRL2ChunkRoot(chunks []PayloadChunk) (string, error) {
	ordered, err := orderedRL2Chunks(chunks)
	if err != nil {
		return "", err
	}
	chunkHashes := make([]string, len(ordered))
	for i, chunk := range ordered {
		chunkHashes[i] = chunk.ChunkHash
	}
	return computeRL2ChunkRootFromHashes(chunkHashes)
}

func ComputeRL2ChunkProofPath(chunks []PayloadChunk, chunkIndex uint32) ([]string, error) {
	ordered, err := orderedRL2Chunks(chunks)
	if err != nil {
		return nil, err
	}
	if chunkIndex >= uint32(len(ordered)) {
		return nil, errors.New("networking RL2 proof chunk index out of range")
	}
	chunkHashes := make([]string, len(ordered))
	for i, chunk := range ordered {
		chunkHashes[i] = chunk.ChunkHash
	}
	_, levels, err := computeRL2MerkleLevels(chunkHashes)
	if err != nil {
		return nil, err
	}
	proof := make([]string, 0, len(levels)-1)
	position := int(chunkIndex)
	for level := 0; level < len(levels)-1; level++ {
		nodes := levels[level]
		sibling := position ^ 1
		if sibling >= len(nodes) {
			sibling = position
		}
		proof = append(proof, nodes[sibling])
		position /= 2
	}
	return proof, nil
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
	bitmap := make([]bool, transfer.ChunkCount)
	for _, index := range normalized {
		bitmap[index] = true
	}
	return RL2TransferProgress{
		TransferID:     transfer.TransferID,
		ReceivedChunks: normalized,
		VerifiedBitmap: bitmap,
		ResumeToken:    token,
	}, nil
}

func NewRL2ChunkDescriptors(transfer RL2Transfer, chunks []PayloadChunk) ([]RL2ChunkDescriptor, error) {
	transfer = transfer.Normalize()
	if err := transfer.Validate(0); err != nil {
		return nil, err
	}
	ordered, err := orderedRL2Chunks(chunks)
	if err != nil {
		return nil, err
	}
	root, err := ComputeRL2ChunkRoot(ordered)
	if err != nil {
		return nil, err
	}
	if root != transfer.PayloadRoot {
		return nil, errors.New("networking RL2 chunk root mismatch")
	}
	if uint32(len(ordered)) != transfer.ChunkCount {
		return nil, errors.New("networking RL2 descriptor chunk count mismatch")
	}
	descriptors := make([]RL2ChunkDescriptor, 0, len(ordered))
	offset := uint64(0)
	for _, chunk := range ordered {
		proofPath, err := ComputeRL2ChunkProofPath(ordered, chunk.Index)
		if err != nil {
			return nil, err
		}
		size := uint64(len(chunk.Bytes))
		descriptor := RL2ChunkDescriptor{
			TransferID: transfer.TransferID,
			ChunkIndex: chunk.Index,
			ChunkHash:  chunk.ChunkHash,
			ChunkSize:  size,
			RangeStart: offset,
			RangeEnd:   offset + size,
			ProofPath:  proofPath,
		}
		if err := descriptor.Validate(transfer); err != nil {
			return nil, err
		}
		descriptors = append(descriptors, descriptor)
		offset += size
	}
	return descriptors, nil
}

func ValidateRL2ChunkDescriptors(transfer RL2Transfer, descriptors []RL2ChunkDescriptor) error {
	transfer = transfer.Normalize()
	if err := transfer.Validate(0); err != nil {
		return err
	}
	if uint32(len(descriptors)) != transfer.ChunkCount {
		return errors.New("networking RL2 descriptor count mismatch")
	}
	ordered := append([]RL2ChunkDescriptor(nil), descriptors...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].ChunkIndex < ordered[j].ChunkIndex
	})
	chunkHashes := make([]string, len(ordered))
	nextStart := uint64(0)
	for i, descriptor := range ordered {
		if descriptor.ChunkIndex != uint32(i) {
			return errors.New("networking RL2 descriptor sequence gap")
		}
		if descriptor.RangeStart != nextStart {
			return errors.New("networking RL2 descriptor ranges must be contiguous")
		}
		if err := descriptor.Validate(transfer); err != nil {
			return err
		}
		nextStart = descriptor.RangeEnd
		chunkHashes[i] = descriptor.ChunkHash
	}
	root, err := computeRL2ChunkRootFromHashes(chunkHashes)
	if err != nil {
		return err
	}
	if root != transfer.PayloadRoot {
		return errors.New("networking RL2 descriptor root mismatch")
	}
	return nil
}

func (d RL2ChunkDescriptor) Normalize() RL2ChunkDescriptor {
	d.TransferID = strings.ToLower(strings.TrimSpace(d.TransferID))
	d.ChunkHash = strings.ToLower(strings.TrimSpace(d.ChunkHash))
	d.ProofPath = normalizeHashPath(d.ProofPath)
	return d
}

func (d RL2ChunkDescriptor) Validate(transfer RL2Transfer) error {
	descriptor := d.Normalize()
	transfer = transfer.Normalize()
	if err := transfer.Validate(0); err != nil {
		return err
	}
	if descriptor.TransferID != transfer.TransferID {
		return errors.New("networking RL2 descriptor transfer mismatch")
	}
	if descriptor.ChunkIndex >= transfer.ChunkCount {
		return errors.New("networking RL2 descriptor chunk index out of range")
	}
	if err := ValidateHash("networking RL2 descriptor chunk hash", descriptor.ChunkHash); err != nil {
		return err
	}
	if descriptor.ChunkSize == 0 || descriptor.ChunkSize > MaxChunkBytes || descriptor.ChunkSize > transfer.ChunkSize {
		return errors.New("networking RL2 descriptor chunk size out of bounds")
	}
	if descriptor.RangeEnd <= descriptor.RangeStart {
		return errors.New("networking RL2 descriptor range must be non-empty")
	}
	if descriptor.RangeEnd-descriptor.RangeStart != descriptor.ChunkSize {
		return errors.New("networking RL2 descriptor range must match chunk size")
	}
	if descriptor.RangeEnd > uint64(transfer.ChunkCount)*transfer.ChunkSize {
		return errors.New("networking RL2 descriptor range exceeds transfer bounds")
	}
	for _, proofHash := range descriptor.ProofPath {
		if err := ValidateHash("networking RL2 descriptor proof hash", proofHash); err != nil {
			return err
		}
	}
	if len(descriptor.ProofPath) > 0 {
		if err := VerifyRL2ChunkProof(descriptor, transfer.PayloadRoot, transfer.ChunkCount); err != nil {
			return err
		}
	}
	return nil
}

func VerifyRL2Chunk(transfer RL2Transfer, descriptor RL2ChunkDescriptor, chunk PayloadChunk) error {
	transfer = transfer.Normalize()
	descriptor = descriptor.Normalize()
	if err := descriptor.Validate(transfer); err != nil {
		return err
	}
	if err := chunk.Validate(); err != nil {
		return err
	}
	if chunk.Index != descriptor.ChunkIndex {
		return errors.New("networking RL2 chunk index mismatch")
	}
	if uint64(len(chunk.Bytes)) != descriptor.ChunkSize {
		return errors.New("networking RL2 chunk size mismatch")
	}
	if chunk.ChunkHash != descriptor.ChunkHash {
		return errors.New("networking RL2 chunk hash mismatch")
	}
	return nil
}

func VerifyRL2ChunkProof(descriptor RL2ChunkDescriptor, payloadRoot string, chunkCount uint32) error {
	descriptor = descriptor.Normalize()
	payloadRoot = strings.ToLower(strings.TrimSpace(payloadRoot))
	if err := ValidateHash("networking RL2 payload root", payloadRoot); err != nil {
		return err
	}
	if chunkCount == 0 || chunkCount > MaxPayloadChunks {
		return fmt.Errorf("networking RL2 chunk count must be between 1 and %d", MaxPayloadChunks)
	}
	if descriptor.ChunkIndex >= chunkCount {
		return errors.New("networking RL2 proof chunk index out of range")
	}
	if err := ValidateHash("networking RL2 proof chunk hash", descriptor.ChunkHash); err != nil {
		return err
	}
	if chunkCount > 1 && len(descriptor.ProofPath) == 0 {
		return errors.New("networking RL2 proof path is required")
	}
	current := HashParts("rl2-chunk-leaf", descriptor.ChunkHash)
	position := descriptor.ChunkIndex
	width := chunkCount
	for _, sibling := range descriptor.ProofPath {
		if err := ValidateHash("networking RL2 proof sibling", sibling); err != nil {
			return err
		}
		if position%2 == 0 {
			current = HashParts("rl2-chunk-node", current, sibling)
		} else {
			current = HashParts("rl2-chunk-node", sibling, current)
		}
		position /= 2
		width = (width + 1) / 2
	}
	if width != 1 {
		return errors.New("networking RL2 proof path is incomplete")
	}
	if current != payloadRoot {
		return errors.New("networking RL2 proof root mismatch")
	}
	return nil
}

func MissingRL2ChunkIndexes(transfer RL2Transfer, verifiedChunks []uint32) ([]uint32, error) {
	progress, err := NewRL2TransferProgress(transfer, verifiedChunks)
	if err != nil {
		return nil, err
	}
	missing := make([]uint32, 0)
	for index, verified := range progress.VerifiedBitmap {
		if !verified {
			missing = append(missing, uint32(index))
		}
	}
	return missing, nil
}

func NewRL2ChunkRequest(transfer RL2Transfer, verifiedChunks []uint32) (RL2ChunkRequest, error) {
	progress, err := NewRL2TransferProgress(transfer, verifiedChunks)
	if err != nil {
		return RL2ChunkRequest{}, err
	}
	missing, err := MissingRL2ChunkIndexes(transfer, verifiedChunks)
	if err != nil {
		return RL2ChunkRequest{}, err
	}
	return RL2ChunkRequest{
		TransferID:     progress.TransferID,
		MissingIndexes: missing,
		ResumeToken:    progress.ResumeToken,
	}, nil
}

func PlanRL2Streaming(transfer RL2Transfer, score PeerScore, availableBytesPerBlock, queuedBytes uint64, maxParallelStreams uint32) (RL2StreamingPlan, error) {
	transfer = transfer.Normalize()
	if err := transfer.Validate(0); err != nil {
		return RL2StreamingPlan{}, err
	}
	if availableBytesPerBlock == 0 {
		return RL2StreamingPlan{}, errors.New("networking RL2 available bandwidth must be positive")
	}
	if score.ScoreBps > BasisPoints {
		return RL2StreamingPlan{}, fmt.Errorf("networking RL2 score must be <= %d bps", BasisPoints)
	}
	if maxParallelStreams == 0 {
		maxParallelStreams = 1
	}
	if maxParallelStreams > transfer.ChunkCount {
		maxParallelStreams = transfer.ChunkCount
	}
	effectiveBudget := availableBytesPerBlock
	if score.ScoreBps > 0 {
		effectiveBudget = (availableBytesPerBlock * uint64(score.ScoreBps)) / uint64(BasisPoints)
	}
	if effectiveBudget < transfer.ChunkSize {
		effectiveBudget = transfer.ChunkSize
	}
	maxInFlight := uint32(effectiveBudget / transfer.ChunkSize)
	if maxInFlight == 0 {
		maxInFlight = 1
	}
	if maxInFlight > transfer.ChunkCount {
		maxInFlight = transfer.ChunkCount
	}
	if maxInFlight > maxParallelStreams*2 {
		maxInFlight = maxParallelStreams * 2
	}
	backpressure := queuedBytes >= effectiveBudget
	parallelStreams := maxParallelStreams
	if backpressure {
		parallelStreams = 1
		maxInFlight = 1
	}
	return RL2StreamingPlan{
		TransferID:         transfer.TransferID,
		Channel:            RL2ChannelForPayloadType(transfer.PayloadType),
		PriorityLane:       transfer.Priority,
		ParallelStreams:    parallelStreams,
		MaxInFlightChunks:  maxInFlight,
		ChunkBudgetBytes:   effectiveBudget,
		BackpressureActive: backpressure,
		FECEnabled:         transfer.FECPolicy != RL2FECNone,
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

func orderedRL2Chunks(chunks []PayloadChunk) ([]PayloadChunk, error) {
	if len(chunks) == 0 {
		return nil, errors.New("networking RL2 transfer requires chunks")
	}
	if _, err := ReassemblePayload(chunks); err != nil {
		return nil, err
	}
	ordered := append([]PayloadChunk(nil), chunks...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Index < ordered[j].Index
	})
	for i, chunk := range ordered {
		if chunk.Index != uint32(i) {
			return nil, errors.New("networking RL2 chunk sequence gap")
		}
	}
	return ordered, nil
}

func computeRL2ChunkRootFromHashes(chunkHashes []string) (string, error) {
	root, _, err := computeRL2MerkleLevels(chunkHashes)
	return root, err
}

func computeRL2MerkleLevels(chunkHashes []string) (string, [][]string, error) {
	if len(chunkHashes) == 0 || len(chunkHashes) > MaxPayloadChunks {
		return "", nil, fmt.Errorf("networking RL2 chunk hashes must be between 1 and %d", MaxPayloadChunks)
	}
	current := make([]string, len(chunkHashes))
	for i, chunkHash := range chunkHashes {
		chunkHash = strings.ToLower(strings.TrimSpace(chunkHash))
		if err := ValidateHash("networking RL2 chunk hash", chunkHash); err != nil {
			return "", nil, err
		}
		current[i] = HashParts("rl2-chunk-leaf", chunkHash)
	}
	levels := [][]string{append([]string(nil), current...)}
	for len(current) > 1 {
		next := make([]string, 0, (len(current)+1)/2)
		for i := 0; i < len(current); i += 2 {
			left := current[i]
			right := left
			if i+1 < len(current) {
				right = current[i+1]
			}
			next = append(next, HashParts("rl2-chunk-node", left, right))
		}
		current = next
		levels = append(levels, append([]string(nil), current...))
	}
	return current[0], levels, nil
}

func normalizeHashPath(path []string) []string {
	out := make([]string, 0, len(path))
	for _, hash := range path {
		hash = strings.ToLower(strings.TrimSpace(hash))
		if hash == "" {
			continue
		}
		out = append(out, hash)
	}
	return out
}
