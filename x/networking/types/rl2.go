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
	RL2PayloadLargeBlock		RL2PayloadType	= "large_block"
	RL2PayloadBlockChunk		RL2PayloadType	= "block_chunk"
	RL2PayloadStateSyncStream	RL2PayloadType	= "state_sync_stream"
	RL2PayloadZoneSnapshot		RL2PayloadType	= "zone_snapshot"
	RL2PayloadExecutionResult	RL2PayloadType	= "execution_result"
	RL2PayloadStorageObject		RL2PayloadType	= "storage_object"
	RL2PayloadProofSet		RL2PayloadType	= "proof_set"
)

type RL2FECPolicy string

const (
	RL2FECNone		RL2FECPolicy	= "none"
	RL2FECXORParity		RL2FECPolicy	= "xor_parity"
	RL2FECReedSolomon	RL2FECPolicy	= "reed_solomon"
)

type RL2TransferState string

const (
	RL2StateOffered			RL2TransferState	= "offered"
	RL2StateAccepted		RL2TransferState	= "accepted"
	RL2StateStreaming		RL2TransferState	= "streaming"
	RL2StatePaused			RL2TransferState	= "paused"
	RL2StateResumed			RL2TransferState	= "resumed"
	RL2StateVerified		RL2TransferState	= "verified"
	RL2StateCompleted		RL2TransferState	= "completed"
	RL2StateTimeout			RL2TransferState	= "timeout"
	RL2StateCancelled		RL2TransferState	= "cancelled"
	RL2StateInvalidChunk		RL2TransferState	= "invalid_chunk"
	RL2StateRootMismatch		RL2TransferState	= "root_mismatch"
	RL2StatePeerDisconnected	RL2TransferState	= "peer_disconnected"
)

type RL2Transfer struct {
	TransferID	string
	SourceNode	string
	TargetNode	string
	PayloadType	RL2PayloadType
	PayloadRoot	string
	ChunkCount	uint32
	ChunkSize	uint64
	FECPolicy	RL2FECPolicy
	Priority	uint32
	DeadlineHeight	uint64
	ResumeToken	string
}

type RL2TransferProgress struct {
	TransferID	string
	ReceivedChunks	[]uint32
	VerifiedBitmap	[]bool
	ResumeToken	string
}

type RL2ChunkDescriptor struct {
	TransferID	string
	ChunkIndex	uint32
	ChunkHash	string
	ChunkSize	uint64
	RangeStart	uint64
	RangeEnd	uint64
	ProofPath	[]string
}

type RL2ChunkRequest struct {
	TransferID	string
	MissingIndexes	[]uint32
	ResumeToken	string
}

type RL2StreamingPlan struct {
	TransferID		string
	Channel			ChannelClass
	PriorityLane		uint32
	ParallelStreams		uint32
	MaxInFlightChunks	uint32
	ChunkBudgetBytes	uint64
	BackpressureActive	bool
	FECEnabled		bool
}

type RL2TransferOffer struct {
	OfferID			string
	Transfer		RL2Transfer
	DescriptorRoot		string
	OfferedHeight		uint64
	ExpiresHeight		uint64
	SuggestedChunkSize	uint64
	MaxParallelStreams	uint32
	ResumeToken		string
}

type RL2TransferAcceptance struct {
	OfferID			string
	TransferID		string
	AcceptedHeight		uint64
	AcceptedChunkSize	uint64
	MaxParallelStreams	uint32
	ResumeToken		string
}

type RL2BackpressureSignal struct {
	TransferID		string
	QueuedBytes		uint64
	AvailableBytes		uint64
	MaxInFlightChunks	uint32
	PauseRequested		bool
	ResumeToken		string
}

type RL2TransferSession struct {
	Offer		RL2TransferOffer
	Acceptance	RL2TransferAcceptance
	State		RL2TransferState
	Progress	RL2TransferProgress
	StreamingPlan	RL2StreamingPlan
	FailureReason	string
	LastHeight	uint64
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
		SourceNode:	sourceNode,
		TargetNode:	targetNode,
		PayloadType:	payloadType,
		PayloadRoot:	payloadRoot,
		ChunkCount:	first.Total,
		ChunkSize:	maxChunkSize,
		FECPolicy:	fecPolicy,
		Priority:	priority,
		DeadlineHeight:	deadlineHeight,
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

func NewRL2TransferOffer(offer RL2TransferOffer) (RL2TransferOffer, error) {
	offer = offer.Normalize()
	if offer.OfferID == "" {
		offer.OfferID = ComputeRL2OfferID(offer)
	}
	if err := offer.Validate(0); err != nil {
		return RL2TransferOffer{}, err
	}
	return offer, nil
}

func NewRL2TransferOfferFromChunks(sourceNode, targetNode string, payloadType RL2PayloadType, chunks []PayloadChunk, priority, offeredHeight, expiresHeight uint32, fecPolicy RL2FECPolicy, score PeerScore, availableBytesPerBlock uint64, maxParallelStreams uint32) (RL2TransferOffer, []RL2ChunkDescriptor, error) {
	transfer, err := NewRL2TransferFromChunks(sourceNode, targetNode, payloadType, chunks, priority, uint64(expiresHeight), fecPolicy)
	if err != nil {
		return RL2TransferOffer{}, nil, err
	}
	descriptors, err := NewRL2ChunkDescriptors(transfer, chunks)
	if err != nil {
		return RL2TransferOffer{}, nil, err
	}
	descriptorRoot, err := ComputeRL2ChunkDescriptorRoot(descriptors)
	if err != nil {
		return RL2TransferOffer{}, nil, err
	}
	chunkSize, err := RecommendRL2ChunkSize(uint64(len(mustReassembleRL2Payload(chunks))), score, availableBytesPerBlock, 1, transfer.ChunkSize)
	if err != nil {
		return RL2TransferOffer{}, nil, err
	}
	offer, err := NewRL2TransferOffer(RL2TransferOffer{
		Transfer:		transfer,
		DescriptorRoot:		descriptorRoot,
		OfferedHeight:		uint64(offeredHeight),
		ExpiresHeight:		uint64(expiresHeight),
		SuggestedChunkSize:	chunkSize,
		MaxParallelStreams:	maxParallelStreams,
	})
	if err != nil {
		return RL2TransferOffer{}, nil, err
	}
	return offer, descriptors, nil
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

func (o RL2TransferOffer) Normalize() RL2TransferOffer {
	o.OfferID = strings.ToLower(strings.TrimSpace(o.OfferID))
	o.Transfer = o.Transfer.Normalize()
	o.DescriptorRoot = strings.ToLower(strings.TrimSpace(o.DescriptorRoot))
	o.ResumeToken = strings.ToLower(strings.TrimSpace(o.ResumeToken))
	if o.MaxParallelStreams == 0 {
		o.MaxParallelStreams = 1
	}
	return o
}

func (o RL2TransferOffer) Validate(currentHeight uint64) error {
	offer := o.Normalize()
	if err := ValidateHash("networking RL2 offer id", offer.OfferID); err != nil {
		return err
	}
	if offer.OfferID != ComputeRL2OfferID(offer) {
		return errors.New("networking RL2 offer id mismatch")
	}
	if err := offer.Transfer.Validate(currentHeight); err != nil {
		return err
	}
	if err := ValidateHash("networking RL2 descriptor root", offer.DescriptorRoot); err != nil {
		return err
	}
	if offer.OfferedHeight == 0 {
		return errors.New("networking RL2 offer height must be positive")
	}
	if offer.ExpiresHeight == 0 || offer.ExpiresHeight < offer.OfferedHeight {
		return errors.New("networking RL2 offer expiry must be >= offer height")
	}
	if currentHeight > 0 && currentHeight > offer.ExpiresHeight {
		return errors.New("networking RL2 offer is expired")
	}
	if offer.SuggestedChunkSize == 0 || offer.SuggestedChunkSize > offer.Transfer.ChunkSize || offer.SuggestedChunkSize > MaxChunkBytes {
		return errors.New("networking RL2 offer suggested chunk size out of bounds")
	}
	if offer.MaxParallelStreams == 0 || offer.MaxParallelStreams > offer.Transfer.ChunkCount {
		return errors.New("networking RL2 offer parallel streams out of bounds")
	}
	if offer.ResumeToken != "" {
		if err := ValidateHash("networking RL2 offer resume token", offer.ResumeToken); err != nil {
			return err
		}
	}
	return nil
}

func (a RL2TransferAcceptance) Normalize() RL2TransferAcceptance {
	a.OfferID = strings.ToLower(strings.TrimSpace(a.OfferID))
	a.TransferID = strings.ToLower(strings.TrimSpace(a.TransferID))
	a.ResumeToken = strings.ToLower(strings.TrimSpace(a.ResumeToken))
	if a.MaxParallelStreams == 0 {
		a.MaxParallelStreams = 1
	}
	return a
}

func (a RL2TransferAcceptance) Validate(offer RL2TransferOffer, currentHeight uint64) error {
	acceptance := a.Normalize()
	offer = offer.Normalize()
	if err := offer.Validate(currentHeight); err != nil {
		return err
	}
	if acceptance.OfferID != offer.OfferID {
		return errors.New("networking RL2 acceptance offer mismatch")
	}
	if acceptance.TransferID != offer.Transfer.TransferID {
		return errors.New("networking RL2 acceptance transfer mismatch")
	}
	if acceptance.AcceptedHeight == 0 || acceptance.AcceptedHeight < offer.OfferedHeight {
		return errors.New("networking RL2 acceptance height must be >= offer height")
	}
	if acceptance.AcceptedHeight > offer.ExpiresHeight {
		return errors.New("networking RL2 acceptance is expired")
	}
	if acceptance.AcceptedChunkSize == 0 || acceptance.AcceptedChunkSize > offer.Transfer.ChunkSize {
		return errors.New("networking RL2 acceptance chunk size out of bounds")
	}
	if acceptance.MaxParallelStreams == 0 || acceptance.MaxParallelStreams > offer.MaxParallelStreams {
		return errors.New("networking RL2 acceptance parallel streams out of bounds")
	}
	if acceptance.ResumeToken != "" {
		if err := ValidateHash("networking RL2 acceptance resume token", acceptance.ResumeToken); err != nil {
			return err
		}
	}
	return nil
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
		Channel:	RL2ChannelForPayloadType(transfer.PayloadType),
		SizeBytes:	transfer.ChunkSize,
		EnqueuedHeight:	enqueuedHeight,
		Sequence:	sequence,
		PayloadHash:	transfer.PayloadRoot,
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

func ComputeRL2OfferID(offer RL2TransferOffer) string {
	offer = offer.Normalize()
	return HashParts(
		"rl2-offer",
		offer.Transfer.TransferID,
		offer.DescriptorRoot,
		fmt.Sprintf("%d", offer.OfferedHeight),
		fmt.Sprintf("%d", offer.ExpiresHeight),
		fmt.Sprintf("%d", offer.SuggestedChunkSize),
		fmt.Sprintf("%d", offer.MaxParallelStreams),
		offer.ResumeToken,
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

func ComputeRL2ChunkDescriptorHash(descriptor RL2ChunkDescriptor) (string, error) {
	descriptor = descriptor.Normalize()
	if err := ValidateHash("networking RL2 descriptor transfer id", descriptor.TransferID); err != nil {
		return "", err
	}
	if err := ValidateHash("networking RL2 descriptor chunk hash", descriptor.ChunkHash); err != nil {
		return "", err
	}
	parts := []string{
		"rl2-chunk-descriptor",
		descriptor.TransferID,
		fmt.Sprintf("%d", descriptor.ChunkIndex),
		descriptor.ChunkHash,
		fmt.Sprintf("%d", descriptor.ChunkSize),
		fmt.Sprintf("%d", descriptor.RangeStart),
		fmt.Sprintf("%d", descriptor.RangeEnd),
	}
	for _, proofHash := range descriptor.ProofPath {
		if err := ValidateHash("networking RL2 descriptor proof hash", proofHash); err != nil {
			return "", err
		}
		parts = append(parts, proofHash)
	}
	return HashParts(parts...), nil
}

func ComputeRL2ChunkDescriptorRoot(descriptors []RL2ChunkDescriptor) (string, error) {
	if len(descriptors) == 0 || len(descriptors) > MaxPayloadChunks {
		return "", fmt.Errorf("networking RL2 descriptors must be between 1 and %d", MaxPayloadChunks)
	}
	ordered := append([]RL2ChunkDescriptor(nil), descriptors...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].ChunkIndex < ordered[j].ChunkIndex
	})
	descriptorHashes := make([]string, len(ordered))
	for i, descriptor := range ordered {
		if descriptor.ChunkIndex != uint32(i) {
			return "", errors.New("networking RL2 descriptor sequence gap")
		}
		descriptorHash, err := ComputeRL2ChunkDescriptorHash(descriptor)
		if err != nil {
			return "", err
		}
		descriptorHashes[i] = descriptorHash
	}
	return HashParts(append([]string{"rl2-descriptor-root"}, descriptorHashes...)...), nil
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
		TransferID:	transfer.TransferID,
		ReceivedChunks:	normalized,
		VerifiedBitmap:	bitmap,
		ResumeToken:	token,
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
			TransferID:	transfer.TransferID,
			ChunkIndex:	chunk.Index,
			ChunkHash:	chunk.ChunkHash,
			ChunkSize:	size,
			RangeStart:	offset,
			RangeEnd:	offset + size,
			ProofPath:	proofPath,
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

func ReassembleRL2Payload(transfer RL2Transfer, descriptors []RL2ChunkDescriptor, chunks []PayloadChunk) ([]byte, error) {
	transfer = transfer.Normalize()
	if err := ValidateRL2ChunkDescriptors(transfer, descriptors); err != nil {
		return nil, err
	}
	if uint32(len(chunks)) != transfer.ChunkCount {
		return nil, errors.New("networking RL2 reassembly chunk count mismatch")
	}
	orderedChunks, err := orderedRL2Chunks(chunks)
	if err != nil {
		return nil, err
	}
	orderedDescriptors := append([]RL2ChunkDescriptor(nil), descriptors...)
	sort.SliceStable(orderedDescriptors, func(i, j int) bool {
		return orderedDescriptors[i].ChunkIndex < orderedDescriptors[j].ChunkIndex
	})
	for i, chunk := range orderedChunks {
		if err := VerifyRL2Chunk(transfer, orderedDescriptors[i], chunk); err != nil {
			return nil, err
		}
	}
	root, err := ComputeRL2ChunkRoot(orderedChunks)
	if err != nil {
		return nil, err
	}
	if root != transfer.PayloadRoot {
		return nil, errors.New("networking RL2 reassembly root mismatch")
	}
	return ReassemblePayload(orderedChunks)
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
		TransferID:	progress.TransferID,
		MissingIndexes:	missing,
		ResumeToken:	progress.ResumeToken,
	}, nil
}

func NewRL2BackpressureSignal(transfer RL2Transfer, queuedBytes, availableBytes uint64, verifiedChunks []uint32) (RL2BackpressureSignal, error) {
	transfer = transfer.Normalize()
	if err := transfer.Validate(0); err != nil {
		return RL2BackpressureSignal{}, err
	}
	if availableBytes == 0 {
		return RL2BackpressureSignal{}, errors.New("networking RL2 backpressure available bytes must be positive")
	}
	progress, err := NewRL2TransferProgress(transfer, verifiedChunks)
	if err != nil {
		return RL2BackpressureSignal{}, err
	}
	maxInFlight := uint32(availableBytes / transfer.ChunkSize)
	if maxInFlight == 0 {
		maxInFlight = 1
	}
	if maxInFlight > transfer.ChunkCount {
		maxInFlight = transfer.ChunkCount
	}
	return RL2BackpressureSignal{
		TransferID:		transfer.TransferID,
		QueuedBytes:		queuedBytes,
		AvailableBytes:		availableBytes,
		MaxInFlightChunks:	maxInFlight,
		PauseRequested:		queuedBytes >= availableBytes,
		ResumeToken:		progress.ResumeToken,
	}, nil
}

func RecommendRL2ChunkSize(payloadBytes uint64, score PeerScore, availableBytesPerBlock, minChunkBytes, maxChunkBytes uint64) (uint64, error) {
	if payloadBytes == 0 {
		return 0, errors.New("networking RL2 payload bytes must be positive")
	}
	if score.ScoreBps > BasisPoints {
		return 0, fmt.Errorf("networking RL2 score must be <= %d bps", BasisPoints)
	}
	if availableBytesPerBlock == 0 {
		return 0, errors.New("networking RL2 available bandwidth must be positive")
	}
	if minChunkBytes == 0 {
		minChunkBytes = 1
	}
	if maxChunkBytes == 0 || maxChunkBytes > MaxChunkBytes {
		maxChunkBytes = MaxChunkBytes
	}
	if minChunkBytes > maxChunkBytes {
		return 0, errors.New("networking RL2 min chunk size must be <= max chunk size")
	}
	target := availableBytesPerBlock / 4
	if score.ScoreBps > 0 {
		target = (target * uint64(score.ScoreBps)) / uint64(BasisPoints)
	}
	if target == 0 {
		target = minChunkBytes
	}
	if target > payloadBytes {
		target = payloadBytes
	}
	if target < minChunkBytes {
		return minChunkBytes, nil
	}
	if target > maxChunkBytes {
		return maxChunkBytes, nil
	}
	return target, nil
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
		TransferID:		transfer.TransferID,
		Channel:		RL2ChannelForPayloadType(transfer.PayloadType),
		PriorityLane:		transfer.Priority,
		ParallelStreams:	parallelStreams,
		MaxInFlightChunks:	maxInFlight,
		ChunkBudgetBytes:	effectiveBudget,
		BackpressureActive:	backpressure,
		FECEnabled:		transfer.FECPolicy != RL2FECNone,
	}, nil
}

func PlanRL2Transfer(adapter AetherNetworkingAdapter, transfer RL2Transfer, enqueuedHeight, sequence uint64, peerCount uint32, score PeerScore) (PropagationPlan, error) {
	envelope, err := transfer.TransportEnvelope(enqueuedHeight, sequence)
	if err != nil {
		return PropagationPlan{}, err
	}
	return PlanPropagation(adapter, envelope, peerCount, score)
}

func AcceptRL2TransferOffer(offer RL2TransferOffer, verifiedChunks []uint32, acceptedHeight uint64) (RL2TransferSession, error) {
	offer = offer.Normalize()
	if err := offer.Validate(acceptedHeight); err != nil {
		return RL2TransferSession{}, err
	}
	progress, err := NewRL2TransferProgress(offer.Transfer, verifiedChunks)
	if err != nil {
		return RL2TransferSession{}, err
	}
	acceptance := RL2TransferAcceptance{
		OfferID:		offer.OfferID,
		TransferID:		offer.Transfer.TransferID,
		AcceptedHeight:		acceptedHeight,
		AcceptedChunkSize:	offer.SuggestedChunkSize,
		MaxParallelStreams:	offer.MaxParallelStreams,
		ResumeToken:		progress.ResumeToken,
	}.Normalize()
	if err := acceptance.Validate(offer, acceptedHeight); err != nil {
		return RL2TransferSession{}, err
	}
	return RL2TransferSession{
		Offer:		offer,
		Acceptance:	acceptance,
		State:		RL2StateAccepted,
		Progress:	progress,
		LastHeight:	acceptedHeight,
	}, nil
}

func StartRL2Transfer(session RL2TransferSession, score PeerScore, availableBytesPerBlock, queuedBytes uint64) (RL2TransferSession, error) {
	if err := session.Validate(); err != nil {
		return RL2TransferSession{}, err
	}
	if session.State != RL2StateAccepted && session.State != RL2StateResumed {
		return RL2TransferSession{}, errors.New("networking RL2 transfer can start only from accepted or resumed")
	}
	plan, err := PlanRL2Streaming(session.Offer.Transfer, score, availableBytesPerBlock, queuedBytes, session.Acceptance.MaxParallelStreams)
	if err != nil {
		return RL2TransferSession{}, err
	}
	session.StreamingPlan = plan
	session.State = RL2StateStreaming
	return session, nil
}

func PauseRL2Transfer(session RL2TransferSession, signal RL2BackpressureSignal, height uint64) (RL2TransferSession, error) {
	if err := session.Validate(); err != nil {
		return RL2TransferSession{}, err
	}
	if session.State != RL2StateStreaming {
		return RL2TransferSession{}, errors.New("networking RL2 transfer can pause only while streaming")
	}
	if signal.TransferID != session.Offer.Transfer.TransferID || !signal.PauseRequested {
		return RL2TransferSession{}, errors.New("networking RL2 pause requires matching backpressure signal")
	}
	session.State = RL2StatePaused
	session.LastHeight = height
	return session, nil
}

func ResumeRL2Transfer(session RL2TransferSession, verifiedChunks []uint32, height uint64) (RL2TransferSession, error) {
	if err := session.Validate(); err != nil {
		return RL2TransferSession{}, err
	}
	if session.State != RL2StatePaused {
		return RL2TransferSession{}, errors.New("networking RL2 transfer can resume only from paused")
	}
	progress, err := NewRL2TransferProgress(session.Offer.Transfer, verifiedChunks)
	if err != nil {
		return RL2TransferSession{}, err
	}
	session.Progress = progress
	session.Acceptance.ResumeToken = progress.ResumeToken
	session.State = RL2StateResumed
	session.LastHeight = height
	return session, nil
}

func AcceptRL2Chunk(session RL2TransferSession, descriptor RL2ChunkDescriptor, chunk PayloadChunk, height uint64) (RL2TransferSession, error) {
	if err := session.Validate(); err != nil {
		return RL2TransferSession{}, err
	}
	if session.State != RL2StateStreaming && session.State != RL2StateResumed {
		return RL2TransferSession{}, errors.New("networking RL2 chunks require streaming or resumed state")
	}
	if err := VerifyRL2Chunk(session.Offer.Transfer, descriptor, chunk); err != nil {
		session.LastHeight = height
		session.FailureReason = err.Error()
		if strings.Contains(err.Error(), "root mismatch") {
			session.State = RL2StateRootMismatch
		} else {
			session.State = RL2StateInvalidChunk
		}
		return session, nil
	}
	received := append([]uint32(nil), session.Progress.ReceivedChunks...)
	if !session.Progress.VerifiedBitmap[chunk.Index] {
		received = append(received, chunk.Index)
	}
	progress, err := NewRL2TransferProgress(session.Offer.Transfer, received)
	if err != nil {
		return RL2TransferSession{}, err
	}
	session.Progress = progress
	session.Acceptance.ResumeToken = progress.ResumeToken
	session.LastHeight = height
	if uint32(len(progress.ReceivedChunks)) == session.Offer.Transfer.ChunkCount {
		session.State = RL2StateVerified
	}
	return session, nil
}

func VerifyRL2TransferCompletion(session RL2TransferSession, descriptors []RL2ChunkDescriptor, chunks []PayloadChunk, height uint64) (RL2TransferSession, []byte, error) {
	if err := session.Validate(); err != nil {
		return RL2TransferSession{}, nil, err
	}
	if session.State != RL2StateVerified {
		return RL2TransferSession{}, nil, errors.New("networking RL2 completion requires verified state")
	}
	payload, err := ReassembleRL2Payload(session.Offer.Transfer, descriptors, chunks)
	if err != nil {
		session.State = RL2StateRootMismatch
		session.FailureReason = err.Error()
		session.LastHeight = height
		return session, nil, nil
	}
	session.State = RL2StateCompleted
	session.LastHeight = height
	return session, payload, nil
}

func FailRL2Transfer(session RL2TransferSession, state RL2TransferState, reason string, height uint64) (RL2TransferSession, error) {
	if !IsRL2FailureState(state) {
		return RL2TransferSession{}, errors.New("networking RL2 failure state required")
	}
	if IsRL2TerminalState(session.State) {
		return RL2TransferSession{}, errors.New("networking RL2 terminal transfer cannot fail again")
	}
	session.State = state
	session.FailureReason = strings.TrimSpace(reason)
	session.LastHeight = height
	return session, nil
}

func (s RL2TransferSession) Validate() error {
	session := s
	if !IsRL2TransferState(session.State) {
		return fmt.Errorf("unknown networking RL2 state %q", session.State)
	}
	if err := session.Offer.Validate(session.LastHeight); err != nil {
		return err
	}
	if session.State == RL2StateOffered {
		return nil
	}
	if err := session.Acceptance.Validate(session.Offer, session.LastHeight); err != nil {
		return err
	}
	if session.Progress.TransferID != session.Offer.Transfer.TransferID {
		return errors.New("networking RL2 session progress transfer mismatch")
	}
	if len(session.Progress.VerifiedBitmap) != int(session.Offer.Transfer.ChunkCount) {
		return errors.New("networking RL2 session progress bitmap mismatch")
	}
	if session.State == RL2StateCompleted && len(session.Progress.ReceivedChunks) != int(session.Offer.Transfer.ChunkCount) {
		return errors.New("networking RL2 completed transfer requires all chunks")
	}
	return nil
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

func IsRL2TransferState(state RL2TransferState) bool {
	switch state {
	case RL2StateOffered,
		RL2StateAccepted,
		RL2StateStreaming,
		RL2StatePaused,
		RL2StateResumed,
		RL2StateVerified,
		RL2StateCompleted,
		RL2StateTimeout,
		RL2StateCancelled,
		RL2StateInvalidChunk,
		RL2StateRootMismatch,
		RL2StatePeerDisconnected:
		return true
	default:
		return false
	}
}

func IsRL2FailureState(state RL2TransferState) bool {
	switch state {
	case RL2StateTimeout, RL2StateCancelled, RL2StateInvalidChunk, RL2StateRootMismatch, RL2StatePeerDisconnected:
		return true
	default:
		return false
	}
}

func IsRL2TerminalState(state RL2TransferState) bool {
	return state == RL2StateCompleted || IsRL2FailureState(state)
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

func mustReassembleRL2Payload(chunks []PayloadChunk) []byte {
	payload, err := ReassemblePayload(chunks)
	if err != nil {
		return nil
	}
	return payload
}
