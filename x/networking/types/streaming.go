package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultStreamParallelism	= uint32(1)
	MaxStreamParallelism		= uint32(64)
	MinStreamChunkBytes		= uint64(256)
)

type StreamingPayloadType string

const (
	StreamingPayloadStateSync		StreamingPayloadType	= "state_sync"
	StreamingPayloadZoneSnapshot		StreamingPayloadType	= "zone_snapshot"
	StreamingPayloadBlockPropagation	StreamingPayloadType	= "block_propagation"
	StreamingPayloadExecutionReceipts	StreamingPayloadType	= "execution_receipts"
	StreamingPayloadStorageObject		StreamingPayloadType	= "storage_object"
	StreamingPayloadProofBundle		StreamingPayloadType	= "proof_bundle"
	StreamingPayloadHistoricalQueryRange	StreamingPayloadType	= "historical_query_range"
)

type StreamSessionState string

const (
	StreamStateOpening	StreamSessionState	= "opening"
	StreamStateActive	StreamSessionState	= "active"
	StreamStatePaused	StreamSessionState	= "paused"
	StreamStateDraining	StreamSessionState	= "draining"
	StreamStateClosed	StreamSessionState	= "closed"
	StreamStateFailed	StreamSessionState	= "failed"
)

type StreamBackpressureSignal string

const (
	StreamSignalWindowUpdate	StreamBackpressureSignal	= "window_update"
	StreamSignalPause		StreamBackpressureSignal	= "pause"
	StreamSignalResume		StreamBackpressureSignal	= "resume"
	StreamSignalSlowDown		StreamBackpressureSignal	= "slow_down"
	StreamSignalCancel		StreamBackpressureSignal	= "cancel"
)

type StreamSession struct {
	StreamID		string
	SessionID		string
	PayloadType		StreamingPayloadType
	Priority		uint32
	FlowControlWindow	uint64
	ChunkSize		uint64
	Parallelism		uint32
	BytesSent		uint64
	BytesAcknowledged	uint64
	State			StreamSessionState
}

type StreamWindowUpdate struct {
	StreamID		string
	BytesSent		uint64
	BytesAcknowledged	uint64
	AvailableWindow		uint64
	Backpressure		bool
}

type StreamChunkPlan struct {
	StreamID		string
	NextOffset		uint64
	ChunkSize		uint64
	MaxInFlightChunks	uint32
	Parallelism		uint32
	Backpressure		bool
}

type StreamBackpressureFrame struct {
	StreamID		string
	Signal			StreamBackpressureSignal
	FlowControlWindow	uint64
	CumulativeAcknowledge	uint64
	SuggestedChunkSize	uint64
	Reason			string
}

type StreamAdaptiveChunkInputs struct {
	ObservedThroughputBps	uint64
	LossRateBps		uint32
	PeerScoreBps		uint32
	PayloadPriority		uint32
	StreamClass		QoSClass
}

type StreamFetchRequest struct {
	StreamID	string
	ChunkIndex	uint32
	RangeStart	uint64
	RangeEnd	uint64
	ChunkSize	uint64
	AssignedPeer	string
}

type StreamParallelFetchPlan struct {
	StreamID	string
	ChunkSize	uint64
	PayloadBytes	uint64
	TotalChunks	uint32
	Requests	[]StreamFetchRequest
	RecoveryResume	bool
}

type StreamMetrics struct {
	StreamID		string
	PayloadType		StreamingPayloadType
	State			StreamSessionState
	BytesSent		uint64
	BytesAcknowledged	uint64
	InFlightBytes		uint64
	AvailableWindow		uint64
	ThroughputBytesBps	uint64
	StallCount		uint64
	BackpressureEvents	uint64
	BackpressureActive	bool
	CompletionBps		uint32
}

func NewStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.StreamID == "" {
		stream.StreamID = ComputeStreamSessionID(stream)
	}
	if err := stream.Validate(); err != nil {
		return StreamSession{}, err
	}
	return stream, nil
}

func StreamSessionFromSpec(session SessionChannel, spec StreamSpec, payloadType StreamingPayloadType, chunkSize uint64, parallelism uint32) (StreamSession, error) {
	if err := session.Validate(); err != nil {
		return StreamSession{}, err
	}
	spec.StreamID = strings.TrimSpace(spec.StreamID)
	spec.EncryptionContext = strings.TrimSpace(spec.EncryptionContext)
	found := false
	for _, stream := range session.Streams {
		if strings.TrimSpace(stream.StreamID) == spec.StreamID {
			found = true
			spec = stream
			break
		}
	}
	if !found {
		return StreamSession{}, errors.New("networking stream spec is not part of session")
	}
	if err := spec.Validate(); err != nil {
		return StreamSession{}, err
	}
	if chunkSize == 0 {
		chunkSize = spec.MaxMessageBytes
		if chunkSize > MaxChunkBytes {
			chunkSize = MaxChunkBytes
		}
	}
	return NewStreamSession(StreamSession{
		SessionID:		session.SessionID,
		PayloadType:		payloadType,
		Priority:		spec.Priority,
		FlowControlWindow:	spec.FlowControlWindow,
		ChunkSize:		chunkSize,
		Parallelism:		parallelism,
		State:			StreamStateOpening,
	})
}

func (s StreamSession) Normalize() StreamSession {
	s.StreamID = normalizeHashText(s.StreamID)
	s.SessionID = normalizeHashText(s.SessionID)
	s.State = normalizeStreamingState(s.State)
	if s.Parallelism == 0 {
		s.Parallelism = DefaultStreamParallelism
	}
	if s.State == "" {
		s.State = StreamStateOpening
	}
	return s
}

func (s StreamSession) Validate() error {
	stream := s.Normalize()
	if err := ValidateHash("networking stream session id", stream.StreamID); err != nil {
		return err
	}
	if ComputeStreamSessionID(stream) != stream.StreamID {
		return errors.New("networking stream session id mismatch")
	}
	if err := ValidateHash("networking parent session id", stream.SessionID); err != nil {
		return err
	}
	if !IsStreamingPayloadType(stream.PayloadType) {
		return fmt.Errorf("unknown networking streaming payload type %q", stream.PayloadType)
	}
	if stream.Priority > MaxRL2Priority {
		return fmt.Errorf("networking stream priority must be <= %d", MaxRL2Priority)
	}
	if stream.FlowControlWindow == 0 || stream.FlowControlWindow > MaxStreamMessageBytes*2 {
		return fmt.Errorf("networking stream flow control window must be between 1 and %d", uint64(MaxStreamMessageBytes*2))
	}
	if stream.ChunkSize == 0 || stream.ChunkSize > MaxChunkBytes {
		return fmt.Errorf("networking stream chunk size must be between 1 and %d", MaxChunkBytes)
	}
	if stream.ChunkSize > stream.FlowControlWindow {
		return errors.New("networking stream chunk size exceeds flow control window")
	}
	if stream.Parallelism == 0 || stream.Parallelism > MaxStreamParallelism {
		return fmt.Errorf("networking stream parallelism must be between 1 and %d", MaxStreamParallelism)
	}
	if stream.BytesAcknowledged > stream.BytesSent {
		return errors.New("networking stream acknowledged bytes exceed sent bytes")
	}
	if !IsStreamSessionState(stream.State) {
		return fmt.Errorf("unknown networking stream session state %q", stream.State)
	}
	if stream.State == StreamStateOpening && (stream.BytesSent != 0 || stream.BytesAcknowledged != 0) {
		return errors.New("networking opening stream session cannot have byte progress")
	}
	return nil
}

func ComputeStreamSessionID(stream StreamSession) string {
	stream.StreamID = ""
	stream = stream.Normalize()
	return HashParts(
		"stream-session",
		stream.SessionID,
		string(stream.PayloadType),
		fmt.Sprintf("%d", stream.Priority),
	)
}

func OpenStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateOpening {
		return StreamSession{}, errors.New("networking stream session can open only from opening state")
	}
	stream.State = StreamStateActive
	return stream, stream.Validate()
}

func PauseStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateActive {
		return StreamSession{}, errors.New("networking stream session can pause only from active state")
	}
	stream.State = StreamStatePaused
	return stream, stream.Validate()
}

func ResumeStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStatePaused {
		return StreamSession{}, errors.New("networking stream session can resume only from paused state")
	}
	stream.State = StreamStateActive
	return stream, stream.Validate()
}

func DrainStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateActive && stream.State != StreamStatePaused {
		return StreamSession{}, errors.New("networking stream session can drain only from active or paused state")
	}
	stream.State = StreamStateDraining
	return stream, stream.Validate()
}

func CloseStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateDraining {
		return StreamSession{}, errors.New("networking stream session can close only from draining state")
	}
	if stream.BytesAcknowledged != stream.BytesSent {
		return StreamSession{}, errors.New("networking stream session cannot close with unacknowledged bytes")
	}
	stream.State = StreamStateClosed
	return stream, stream.Validate()
}

func FailStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State == StreamStateClosed || stream.State == StreamStateFailed {
		return StreamSession{}, errors.New("networking terminal stream session cannot fail again")
	}
	stream.State = StreamStateFailed
	return stream, stream.Validate()
}

func RecordStreamBytesSent(stream StreamSession, bytes uint64) (StreamSession, StreamWindowUpdate, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateActive && stream.State != StreamStateDraining {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream session must be active or draining to send bytes")
	}
	if bytes == 0 {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream sent byte increment must be positive")
	}
	available := StreamAvailableWindow(stream)
	if bytes > available {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream flow control window exceeded")
	}
	stream.BytesSent += bytes
	if err := stream.Validate(); err != nil {
		return StreamSession{}, StreamWindowUpdate{}, err
	}
	return stream, StreamWindow(stream), nil
}

func AcknowledgeStreamBytes(stream StreamSession, acknowledged uint64) (StreamSession, StreamWindowUpdate, error) {
	stream = stream.Normalize()
	if stream.State == StreamStateOpening || stream.State == StreamStateClosed || stream.State == StreamStateFailed {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream session cannot acknowledge bytes in current state")
	}
	if acknowledged < stream.BytesAcknowledged {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream acknowledged bytes cannot regress")
	}
	if acknowledged > stream.BytesSent {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream acknowledged bytes exceed sent bytes")
	}
	stream.BytesAcknowledged = acknowledged
	if err := stream.Validate(); err != nil {
		return StreamSession{}, StreamWindowUpdate{}, err
	}
	return stream, StreamWindow(stream), nil
}

func ApplyStreamBackpressure(stream StreamSession, frame StreamBackpressureFrame) (StreamSession, StreamWindowUpdate, error) {
	stream = stream.Normalize()
	frame.StreamID = normalizeHashText(frame.StreamID)
	if err := stream.Validate(); err != nil {
		return StreamSession{}, StreamWindowUpdate{}, err
	}
	if frame.StreamID != stream.StreamID {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream backpressure frame id mismatch")
	}
	if frame.CumulativeAcknowledge > 0 || frame.Signal == StreamSignalWindowUpdate {
		var err error
		stream, _, err = AcknowledgeStreamBytes(stream, frame.CumulativeAcknowledge)
		if err != nil {
			return StreamSession{}, StreamWindowUpdate{}, err
		}
	}
	switch frame.Signal {
	case StreamSignalWindowUpdate:
		if frame.FlowControlWindow == 0 || frame.FlowControlWindow > MaxStreamMessageBytes*2 {
			return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream window update out of bounds")
		}
		if frame.FlowControlWindow < stream.ChunkSize {
			return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream window update cannot be smaller than chunk size")
		}
		inFlight := stream.BytesSent - stream.BytesAcknowledged
		if frame.FlowControlWindow < inFlight {
			return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream window update below in-flight bytes")
		}
		stream.FlowControlWindow = frame.FlowControlWindow
	case StreamSignalPause:
		var err error
		stream, err = PauseStreamSession(stream)
		if err != nil {
			return StreamSession{}, StreamWindowUpdate{}, err
		}
	case StreamSignalResume:
		var err error
		stream, err = ResumeStreamSession(stream)
		if err != nil {
			return StreamSession{}, StreamWindowUpdate{}, err
		}
	case StreamSignalSlowDown:
		next := frame.SuggestedChunkSize
		if next == 0 || next >= stream.ChunkSize {
			next = stream.ChunkSize / 2
		}
		if next < MinStreamChunkBytes {
			next = MinStreamChunkBytes
		}
		var err error
		stream, err = ApplyStreamChunkSize(stream, next)
		if err != nil {
			return StreamSession{}, StreamWindowUpdate{}, err
		}
	case StreamSignalCancel:
		var err error
		stream, err = FailStreamSession(stream)
		if err != nil {
			return StreamSession{}, StreamWindowUpdate{}, err
		}
	default:
		return StreamSession{}, StreamWindowUpdate{}, fmt.Errorf("unknown networking stream backpressure signal %q", frame.Signal)
	}
	if err := stream.Validate(); err != nil {
		return StreamSession{}, StreamWindowUpdate{}, err
	}
	return stream, StreamWindow(stream), nil
}

func StreamWindow(stream StreamSession) StreamWindowUpdate {
	stream = stream.Normalize()
	available := StreamAvailableWindow(stream)
	return StreamWindowUpdate{
		StreamID:		stream.StreamID,
		BytesSent:		stream.BytesSent,
		BytesAcknowledged:	stream.BytesAcknowledged,
		AvailableWindow:	available,
		Backpressure:		available == 0 || stream.State == StreamStatePaused,
	}
}

func StreamAvailableWindow(stream StreamSession) uint64 {
	stream = stream.Normalize()
	inFlight := stream.BytesSent - stream.BytesAcknowledged
	if inFlight >= stream.FlowControlWindow {
		return 0
	}
	return stream.FlowControlWindow - inFlight
}

func PlanStreamChunks(stream StreamSession, remainingBytes uint64) (StreamChunkPlan, error) {
	stream = stream.Normalize()
	if err := stream.Validate(); err != nil {
		return StreamChunkPlan{}, err
	}
	if stream.State != StreamStateActive && stream.State != StreamStateDraining {
		return StreamChunkPlan{}, errors.New("networking stream session must be active or draining to plan chunks")
	}
	if remainingBytes == 0 {
		return StreamChunkPlan{
			StreamID:	stream.StreamID,
			NextOffset:	stream.BytesSent,
			Parallelism:	stream.Parallelism,
			Backpressure:	false,
		}, nil
	}
	available := StreamAvailableWindow(stream)
	if available == 0 {
		return StreamChunkPlan{
			StreamID:	stream.StreamID,
			NextOffset:	stream.BytesSent,
			Parallelism:	stream.Parallelism,
			Backpressure:	true,
		}, nil
	}
	chunkSize := minStreamingUint64(stream.ChunkSize, remainingBytes)
	chunkSize = minStreamingUint64(chunkSize, available)
	maxChunks := available / chunkSize
	if maxChunks == 0 {
		maxChunks = 1
	}
	if maxChunks > uint64(stream.Parallelism) {
		maxChunks = uint64(stream.Parallelism)
	}
	return StreamChunkPlan{
		StreamID:		stream.StreamID,
		NextOffset:		stream.BytesSent,
		ChunkSize:		chunkSize,
		MaxInFlightChunks:	uint32(maxChunks),
		Parallelism:		stream.Parallelism,
		Backpressure:		false,
	}, nil
}

func RecommendStreamChunkSize(stream StreamSession, inputs StreamAdaptiveChunkInputs) (uint64, error) {
	stream = stream.Normalize()
	if err := stream.Validate(); err != nil {
		return 0, err
	}
	if inputs.PeerScoreBps > BasisPoints || inputs.LossRateBps > BasisPoints {
		return 0, errors.New("networking stream adaptive bps inputs out of bounds")
	}
	if inputs.PayloadPriority > MaxRL2Priority {
		return 0, fmt.Errorf("networking stream adaptive payload priority must be <= %d", MaxRL2Priority)
	}
	if inputs.StreamClass != "" {
		if _, found := findQoSClassPolicy(DefaultQoSClassPolicies(), inputs.StreamClass); !found {
			return 0, fmt.Errorf("unknown networking stream adaptive class %q", inputs.StreamClass)
		}
	}
	next := stream.ChunkSize
	switch {
	case inputs.LossRateBps >= 500 || inputs.PeerScoreBps < 5_000:
		next = stream.ChunkSize / 2
	case inputs.ObservedThroughputBps >= 16<<20 && inputs.LossRateBps <= 100 && inputs.PeerScoreBps >= 8_000:
		next = stream.ChunkSize * 2
	case inputs.PayloadPriority <= PriorityForChannel(ChannelExecution) && inputs.LossRateBps <= 250 && inputs.PeerScoreBps >= 7_000:
		next = stream.ChunkSize + stream.ChunkSize/2
	}
	if inputs.StreamClass == QoSClassBulkData && inputs.LossRateBps > 0 {
		next = minStreamingUint64(next, stream.ChunkSize)
	}
	if next < MinStreamChunkBytes {
		next = MinStreamChunkBytes
	}
	limit := minStreamingUint64(MaxChunkBytes, stream.FlowControlWindow)
	if next > limit {
		next = limit
	}
	return next, nil
}

func ApplyStreamChunkSize(stream StreamSession, chunkSize uint64) (StreamSession, error) {
	stream = stream.Normalize()
	if err := stream.Validate(); err != nil {
		return StreamSession{}, err
	}
	if stream.BytesSent != stream.BytesAcknowledged {
		return StreamSession{}, errors.New("networking stream chunk size can change only at chunk boundary")
	}
	if chunkSize < MinStreamChunkBytes || chunkSize > MaxChunkBytes {
		return StreamSession{}, fmt.Errorf("networking stream chunk size must be between %d and %d", MinStreamChunkBytes, uint64(MaxChunkBytes))
	}
	if chunkSize > stream.FlowControlWindow {
		return StreamSession{}, errors.New("networking stream chunk size exceeds flow control window")
	}
	stream.ChunkSize = chunkSize
	return stream, stream.Validate()
}

func PlanParallelStreamFetch(stream StreamSession, payloadBytes uint64, verifiedBitmap []bool, peers []string) (StreamParallelFetchPlan, error) {
	stream = stream.Normalize()
	if err := stream.Validate(); err != nil {
		return StreamParallelFetchPlan{}, err
	}
	if stream.State != StreamStateActive && stream.State != StreamStateDraining {
		return StreamParallelFetchPlan{}, errors.New("networking stream session must be active or draining to fetch chunks")
	}
	if payloadBytes == 0 {
		return StreamParallelFetchPlan{}, errors.New("networking stream fetch payload bytes are required")
	}
	if len(peers) == 0 {
		return StreamParallelFetchPlan{}, errors.New("networking stream fetch peers are required")
	}
	totalChunks64 := (payloadBytes + stream.ChunkSize - 1) / stream.ChunkSize
	if totalChunks64 == 0 || totalChunks64 > MaxPayloadChunks {
		return StreamParallelFetchPlan{}, fmt.Errorf("networking stream fetch chunks must be between 1 and %d", MaxPayloadChunks)
	}
	totalChunks := uint32(totalChunks64)
	if len(verifiedBitmap) > int(totalChunks) {
		return StreamParallelFetchPlan{}, errors.New("networking stream verified bitmap exceeds chunk count")
	}
	requests := make([]StreamFetchRequest, 0, stream.Parallelism)
	limit := int(stream.Parallelism)
	for index := uint32(0); index < totalChunks && len(requests) < limit; index++ {
		if int(index) < len(verifiedBitmap) && verifiedBitmap[index] {
			continue
		}
		start := uint64(index) * stream.ChunkSize
		end := start + stream.ChunkSize
		if end > payloadBytes {
			end = payloadBytes
		}
		requests = append(requests, StreamFetchRequest{
			StreamID:	stream.StreamID,
			ChunkIndex:	index,
			RangeStart:	start,
			RangeEnd:	end,
			ChunkSize:	end - start,
			AssignedPeer:	peers[len(requests)%len(peers)],
		})
	}
	return StreamParallelFetchPlan{
		StreamID:	stream.StreamID,
		ChunkSize:	stream.ChunkSize,
		PayloadBytes:	payloadBytes,
		TotalChunks:	totalChunks,
		Requests:	requests,
		RecoveryResume:	len(verifiedBitmap) > 0,
	}, nil
}

func SortStreamPriorityLanes(streams []StreamSession) ([]StreamSession, error) {
	out := make([]StreamSession, len(streams))
	for i, stream := range streams {
		stream = stream.Normalize()
		if err := stream.Validate(); err != nil {
			return nil, err
		}
		out[i] = stream
	}
	sortStreamingSessions(out)
	return out, nil
}

func ComputeStreamMetrics(stream StreamSession, elapsedMillis, stallCount, backpressureEvents uint64) (StreamMetrics, error) {
	stream = stream.Normalize()
	if err := stream.Validate(); err != nil {
		return StreamMetrics{}, err
	}
	inFlight := stream.BytesSent - stream.BytesAcknowledged
	throughput := uint64(0)
	if elapsedMillis > 0 {
		throughput = stream.BytesAcknowledged * 1_000 / elapsedMillis
	}
	completion := uint32(0)
	if stream.BytesSent > 0 {
		completion = uint32(stream.BytesAcknowledged * uint64(BasisPoints) / stream.BytesSent)
	}
	window := StreamWindow(stream)
	return StreamMetrics{
		StreamID:		stream.StreamID,
		PayloadType:		stream.PayloadType,
		State:			stream.State,
		BytesSent:		stream.BytesSent,
		BytesAcknowledged:	stream.BytesAcknowledged,
		InFlightBytes:		inFlight,
		AvailableWindow:	window.AvailableWindow,
		ThroughputBytesBps:	throughput,
		StallCount:		stallCount,
		BackpressureEvents:	backpressureEvents,
		BackpressureActive:	window.Backpressure || stream.State == StreamStatePaused,
		CompletionBps:		completion,
	}, nil
}

func StreamReassemblyRoot(payload []byte) (string, error) {
	if len(payload) == 0 {
		return "", errors.New("networking stream reassembly payload is required")
	}
	return hashBytes("aetra-networking-payload-v1", payload), nil
}

func VerifyStreamChunkHash(chunk PayloadChunk) error {
	if chunk.ChunkHash != ComputeChunkHash(chunk) {
		return errors.New("networking stream chunk hash mismatch")
	}
	return nil
}

func IsStreamingPayloadType(payloadType StreamingPayloadType) bool {
	switch payloadType {
	case StreamingPayloadStateSync,
		StreamingPayloadZoneSnapshot,
		StreamingPayloadBlockPropagation,
		StreamingPayloadExecutionReceipts,
		StreamingPayloadStorageObject,
		StreamingPayloadProofBundle,
		StreamingPayloadHistoricalQueryRange:
		return true
	default:
		return false
	}
}

func IsStreamSessionState(state StreamSessionState) bool {
	switch state {
	case StreamStateOpening,
		StreamStateActive,
		StreamStatePaused,
		StreamStateDraining,
		StreamStateClosed,
		StreamStateFailed:
		return true
	default:
		return false
	}
}

func StreamingPayloadRL2Type(payloadType StreamingPayloadType) (RL2PayloadType, error) {
	switch payloadType {
	case StreamingPayloadStateSync:
		return RL2PayloadStateSyncStream, nil
	case StreamingPayloadZoneSnapshot:
		return RL2PayloadZoneSnapshot, nil
	case StreamingPayloadBlockPropagation:
		return RL2PayloadLargeBlock, nil
	case StreamingPayloadExecutionReceipts:
		return RL2PayloadExecutionResult, nil
	case StreamingPayloadStorageObject, StreamingPayloadHistoricalQueryRange:
		return RL2PayloadStorageObject, nil
	case StreamingPayloadProofBundle:
		return RL2PayloadProofSet, nil
	default:
		return "", fmt.Errorf("unknown networking streaming payload type %q", payloadType)
	}
}

func StreamingPayloadChannel(payloadType StreamingPayloadType) (ChannelClass, error) {
	switch payloadType {
	case StreamingPayloadStateSync:
		return ChannelStateSync, nil
	case StreamingPayloadBlockPropagation:
		return ChannelBlock, nil
	case StreamingPayloadExecutionReceipts:
		return ChannelExecution, nil
	case StreamingPayloadZoneSnapshot,
		StreamingPayloadStorageObject,
		StreamingPayloadProofBundle,
		StreamingPayloadHistoricalQueryRange:
		return ChannelData, nil
	default:
		return "", fmt.Errorf("unknown networking streaming payload type %q", payloadType)
	}
}

func normalizeStreamingState(state StreamSessionState) StreamSessionState {
	return StreamSessionState(strings.ToLower(strings.TrimSpace(string(state))))
}

func streamPayloadOrder(payloadType StreamingPayloadType) uint32 {
	channel, err := StreamingPayloadChannel(payloadType)
	if err != nil {
		return MaxRL2Priority + 1
	}
	return PriorityForChannel(channel)
}

func sortStreamingSessions(streams []StreamSession) {
	for i := 1; i < len(streams); i++ {
		current := streams[i]
		j := i - 1
		for j >= 0 && streamLess(current, streams[j]) {
			streams[j+1] = streams[j]
			j--
		}
		streams[j+1] = current
	}
}

func streamLess(left, right StreamSession) bool {
	leftOrder := streamPayloadOrder(left.PayloadType)
	rightOrder := streamPayloadOrder(right.PayloadType)
	if leftOrder != rightOrder {
		return leftOrder < rightOrder
	}
	if left.Priority != right.Priority {
		return left.Priority < right.Priority
	}
	if left.BytesSent-left.BytesAcknowledged != right.BytesSent-right.BytesAcknowledged {
		return left.BytesSent-left.BytesAcknowledged < right.BytesSent-right.BytesAcknowledged
	}
	return left.StreamID < right.StreamID
}

func minStreamingUint64(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}
