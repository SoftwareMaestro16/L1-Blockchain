package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const DefaultOverlayQueueLimit = uint32(1024)

type OverlayMessageQueue struct {
	OverlayID	string
	MaxMessages	uint32
	Messages	[]AetherMeshMessage
	SeenMessageIDs	[]string
}

type CrossZoneSequenceState struct {
	SourceZone	string
	DestinationZone	string
	NextSequence	uint64
	LastAccepted	uint64
}

type CrossZoneSequenceTracker struct {
	States []CrossZoneSequenceState
}

type ReceiptDeliveryState string

const (
	ReceiptDeliveryPending		ReceiptDeliveryState	= "pending"
	ReceiptDeliveryAcknowledged	ReceiptDeliveryState	= "acknowledged"
)

type ReceiptDelivery struct {
	DeliveryID	string
	Receipt		CrossZoneReceipt
	DestinationNode	string
	DeliveredHeight	uint64
	AckHash		string
	State		ReceiptDeliveryState
}

type QueryResponseProof struct {
	ResponseID	string
	RequestID	string
	Responder	string
	PayloadHash	string
	Proof		AetherMeshProof
	ResponseHeight	uint64
	Signature	[]byte
}

type L3OverlayMetrics struct {
	OverlayID	string
	QueuedCount	uint64
	DeliveredCount	uint64
	ReplayDropCount	uint64
	ExpiredCount	uint64
	ReceiptCount	uint64
	QueryProofCount	uint64
}

func NewOverlayMessageQueue(overlayID string, maxMessages uint32) (OverlayMessageQueue, error) {
	overlayID = normalizeHashText(overlayID)
	if err := ValidateHash("networking overlay message queue overlay id", overlayID); err != nil {
		return OverlayMessageQueue{}, err
	}
	if maxMessages == 0 {
		maxMessages = DefaultOverlayQueueLimit
	}
	return OverlayMessageQueue{OverlayID: overlayID, MaxMessages: maxMessages}, nil
}

func EnqueueOverlayMessage(queue OverlayMessageQueue, msg AetherMeshMessage, currentHeight uint64) (OverlayMessageQueue, error) {
	queue = NormalizeOverlayMessageQueue(queue)
	if err := queue.Validate(currentHeight); err != nil {
		return OverlayMessageQueue{}, err
	}
	msg = NormalizeAetherMeshMessage(msg)
	if err := msg.ValidateBasic(currentHeight); err != nil {
		return OverlayMessageQueue{}, err
	}
	if msg.OverlayID != queue.OverlayID {
		return OverlayMessageQueue{}, errors.New("networking overlay message queue overlay mismatch")
	}
	if containsString(queue.SeenMessageIDs, msg.MessageID) {
		return OverlayMessageQueue{}, errors.New("networking overlay message replay detected")
	}
	if uint32(len(queue.Messages)) >= queue.MaxMessages {
		return OverlayMessageQueue{}, errors.New("networking overlay message queue is full")
	}
	next := queue
	next.Messages = append(next.Messages, msg)
	next.SeenMessageIDs = append(next.SeenMessageIDs, msg.MessageID)
	sortAetherMeshMessages(next.Messages)
	sortStrings(next.SeenMessageIDs)
	return next, next.Validate(currentHeight)
}

func DequeueOverlayMessages(queue OverlayMessageQueue, limit uint32) (OverlayMessageQueue, []AetherMeshMessage, error) {
	queue = NormalizeOverlayMessageQueue(queue)
	if err := queue.Validate(0); err != nil {
		return OverlayMessageQueue{}, nil, err
	}
	if limit == 0 || limit > uint32(len(queue.Messages)) {
		limit = uint32(len(queue.Messages))
	}
	out := append([]AetherMeshMessage(nil), queue.Messages[:limit]...)
	next := queue
	next.Messages = append([]AetherMeshMessage(nil), queue.Messages[limit:]...)
	return next, out, nil
}

func NormalizeOverlayMessageQueue(queue OverlayMessageQueue) OverlayMessageQueue {
	queue.OverlayID = normalizeHashText(queue.OverlayID)
	if queue.MaxMessages == 0 {
		queue.MaxMessages = DefaultOverlayQueueLimit
	}
	for i := range queue.Messages {
		queue.Messages[i] = NormalizeAetherMeshMessage(queue.Messages[i])
	}
	queue.SeenMessageIDs = normalizeHashSet(queue.SeenMessageIDs)
	sortAetherMeshMessages(queue.Messages)
	return queue
}

func (q OverlayMessageQueue) Validate(currentHeight uint64) error {
	queue := NormalizeOverlayMessageQueue(q)
	if err := ValidateHash("networking overlay message queue overlay id", queue.OverlayID); err != nil {
		return err
	}
	if queue.MaxMessages == 0 {
		return errors.New("networking overlay message queue max messages must be positive")
	}
	if uint32(len(queue.Messages)) > queue.MaxMessages {
		return errors.New("networking overlay message queue exceeds max messages")
	}
	seen := make(map[string]struct{}, len(queue.Messages))
	for _, msg := range queue.Messages {
		if err := msg.ValidateBasic(currentHeight); err != nil {
			return err
		}
		if msg.OverlayID != queue.OverlayID {
			return errors.New("networking overlay message queue contains wrong overlay")
		}
		if _, found := seen[msg.MessageID]; found {
			return errors.New("networking duplicate overlay queued message")
		}
		seen[msg.MessageID] = struct{}{}
	}
	for _, id := range queue.SeenMessageIDs {
		if err := ValidateHash("networking overlay seen message id", id); err != nil {
			return err
		}
	}
	return nil
}

func NextCrossZoneSequence(tracker CrossZoneSequenceTracker, sourceZone, destinationZone string) (CrossZoneSequenceTracker, uint64, error) {
	sourceZone = strings.TrimSpace(sourceZone)
	destinationZone = strings.TrimSpace(destinationZone)
	if sourceZone == "" || destinationZone == "" || sourceZone == destinationZone {
		return CrossZoneSequenceTracker{}, 0, errors.New("networking cross-zone sequence requires distinct zones")
	}
	next := tracker.Clone()
	for i, state := range next.States {
		if state.SourceZone == sourceZone && state.DestinationZone == destinationZone {
			seq := state.NextSequence
			if seq == 0 {
				seq = 1
			}
			next.States[i].NextSequence = seq + 1
			sortCrossZoneSequenceStates(next.States)
			return next, seq, nil
		}
	}
	next.States = append(next.States, CrossZoneSequenceState{
		SourceZone:		sourceZone,
		DestinationZone:	destinationZone,
		NextSequence:		2,
	})
	sortCrossZoneSequenceStates(next.States)
	return next, 1, nil
}

func AcceptCrossZoneSequence(tracker CrossZoneSequenceTracker, msg CrossZoneMessage, ordered bool, currentHeight uint64) (CrossZoneSequenceTracker, error) {
	msg = NormalizeCrossZoneMessage(msg)
	if err := msg.Validate(currentHeight); err != nil {
		return CrossZoneSequenceTracker{}, err
	}
	next := tracker.Clone()
	for i, state := range next.States {
		if state.SourceZone != msg.SourceZone || state.DestinationZone != msg.DestinationZone {
			continue
		}
		if msg.SourceSequence <= state.LastAccepted {
			return CrossZoneSequenceTracker{}, errors.New("networking cross-zone sequence replay detected")
		}
		if ordered && msg.SourceSequence != state.LastAccepted+1 {
			return CrossZoneSequenceTracker{}, errors.New("networking cross-zone sequence gap")
		}
		next.States[i].LastAccepted = msg.SourceSequence
		if next.States[i].NextSequence <= msg.SourceSequence {
			next.States[i].NextSequence = msg.SourceSequence + 1
		}
		return next, nil
	}
	if ordered && msg.SourceSequence != 1 {
		return CrossZoneSequenceTracker{}, errors.New("networking cross-zone sequence gap")
	}
	next.States = append(next.States, CrossZoneSequenceState{
		SourceZone:		msg.SourceZone,
		DestinationZone:	msg.DestinationZone,
		NextSequence:		msg.SourceSequence + 1,
		LastAccepted:		msg.SourceSequence,
	})
	sortCrossZoneSequenceStates(next.States)
	return next, nil
}

func (t CrossZoneSequenceTracker) Clone() CrossZoneSequenceTracker {
	out := CrossZoneSequenceTracker{States: append([]CrossZoneSequenceState(nil), t.States...)}
	for i := range out.States {
		out.States[i].SourceZone = strings.TrimSpace(out.States[i].SourceZone)
		out.States[i].DestinationZone = strings.TrimSpace(out.States[i].DestinationZone)
	}
	sortCrossZoneSequenceStates(out.States)
	return out
}

func NewReceiptDelivery(receipt CrossZoneReceipt, destinationNode string, deliveredHeight uint64) (ReceiptDelivery, error) {
	receipt = NormalizeCrossZoneReceipt(receipt)
	if err := receipt.Validate(); err != nil {
		return ReceiptDelivery{}, err
	}
	destinationNode = normalizeHashText(destinationNode)
	if err := ValidateHash("networking receipt delivery destination node", destinationNode); err != nil {
		return ReceiptDelivery{}, err
	}
	delivery := ReceiptDelivery{
		Receipt:		receipt,
		DestinationNode:	destinationNode,
		DeliveredHeight:	deliveredHeight,
		State:			ReceiptDeliveryPending,
	}
	delivery.DeliveryID = ComputeReceiptDeliveryID(delivery)
	if err := delivery.Validate(); err != nil {
		return ReceiptDelivery{}, err
	}
	return delivery, nil
}

func AckReceiptDelivery(delivery ReceiptDelivery, ackHash string) (ReceiptDelivery, error) {
	delivery = NormalizeReceiptDelivery(delivery)
	if err := delivery.Validate(); err != nil {
		return ReceiptDelivery{}, err
	}
	ackHash = normalizeHashText(ackHash)
	if err := ValidateHash("networking receipt delivery ack hash", ackHash); err != nil {
		return ReceiptDelivery{}, err
	}
	delivery.AckHash = ackHash
	delivery.State = ReceiptDeliveryAcknowledged
	delivery.DeliveryID = ComputeReceiptDeliveryID(delivery)
	return delivery, delivery.Validate()
}

func NormalizeReceiptDelivery(delivery ReceiptDelivery) ReceiptDelivery {
	delivery.DeliveryID = normalizeHashText(delivery.DeliveryID)
	delivery.Receipt = NormalizeCrossZoneReceipt(delivery.Receipt)
	delivery.DestinationNode = normalizeHashText(delivery.DestinationNode)
	delivery.AckHash = normalizeHashText(delivery.AckHash)
	delivery.State = ReceiptDeliveryState(strings.ToLower(strings.TrimSpace(string(delivery.State))))
	return delivery
}

func ComputeReceiptDeliveryID(delivery ReceiptDelivery) string {
	delivery = NormalizeReceiptDelivery(delivery)
	return HashParts(
		"receipt-delivery",
		delivery.Receipt.ReceiptID,
		delivery.DestinationNode,
		fmt.Sprintf("%d", delivery.DeliveredHeight),
		delivery.AckHash,
		string(delivery.State),
	)
}

func (d ReceiptDelivery) Validate() error {
	delivery := NormalizeReceiptDelivery(d)
	if err := ValidateHash("networking receipt delivery id", delivery.DeliveryID); err != nil {
		return err
	}
	if delivery.DeliveryID != ComputeReceiptDeliveryID(delivery) {
		return errors.New("networking receipt delivery id does not match payload")
	}
	if err := delivery.Receipt.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("networking receipt delivery destination node", delivery.DestinationNode); err != nil {
		return err
	}
	if delivery.DeliveredHeight == 0 {
		return errors.New("networking receipt delivery height must be positive")
	}
	if delivery.State != ReceiptDeliveryPending && delivery.State != ReceiptDeliveryAcknowledged {
		return fmt.Errorf("unknown networking receipt delivery state %q", delivery.State)
	}
	if delivery.State == ReceiptDeliveryAcknowledged {
		if err := ValidateHash("networking receipt delivery ack hash", delivery.AckHash); err != nil {
			return err
		}
	}
	return nil
}

func NewQueryResponseProof(response QueryResponseProof) (QueryResponseProof, error) {
	response = NormalizeQueryResponseProof(response)
	if response.ResponseID == "" {
		response.ResponseID = ComputeQueryResponseID(response)
	}
	if err := response.Validate(); err != nil {
		return QueryResponseProof{}, err
	}
	return response, nil
}

func NormalizeQueryResponseProof(response QueryResponseProof) QueryResponseProof {
	response.ResponseID = normalizeHashText(response.ResponseID)
	response.RequestID = normalizeHashText(response.RequestID)
	response.Responder = normalizeHashText(response.Responder)
	response.PayloadHash = normalizeHashText(response.PayloadHash)
	response.Proof.ProofType = strings.TrimSpace(response.Proof.ProofType)
	response.Proof.ProofHash = normalizeHashText(response.Proof.ProofHash)
	response.Signature = cloneBytes(response.Signature)
	return response
}

func ComputeQueryResponseID(response QueryResponseProof) string {
	response = NormalizeQueryResponseProof(response)
	return HashParts(
		"query-response-proof",
		response.RequestID,
		response.Responder,
		response.PayloadHash,
		response.Proof.ProofType,
		response.Proof.ProofHash,
		fmt.Sprintf("%d", response.Proof.ProofHeight),
		fmt.Sprintf("%d", response.ResponseHeight),
	)
}

func (r QueryResponseProof) Validate() error {
	response := NormalizeQueryResponseProof(r)
	if err := ValidateHash("networking query response id", response.ResponseID); err != nil {
		return err
	}
	if response.ResponseID != ComputeQueryResponseID(response) {
		return errors.New("networking query response id does not match payload")
	}
	if err := ValidateHash("networking query request id", response.RequestID); err != nil {
		return err
	}
	if err := ValidateHash("networking query responder", response.Responder); err != nil {
		return err
	}
	if err := ValidateHash("networking query payload hash", response.PayloadHash); err != nil {
		return err
	}
	if response.ResponseHeight == 0 {
		return errors.New("networking query response height must be positive")
	}
	return response.Proof.Validate(true)
}

func EvaluateL3Metrics(queues []OverlayMessageQueue, deliveries []ReceiptDelivery, queryResponses []QueryResponseProof, replayDrops map[string]uint64, expired map[string]uint64) ([]L3OverlayMetrics, error) {
	byOverlay := make(map[string]L3OverlayMetrics)
	for _, queue := range queues {
		queue = NormalizeOverlayMessageQueue(queue)
		if err := queue.Validate(0); err != nil {
			return nil, err
		}
		metric := byOverlay[queue.OverlayID]
		metric.OverlayID = queue.OverlayID
		metric.QueuedCount += uint64(len(queue.Messages))
		metric.ReplayDropCount += replayDrops[queue.OverlayID]
		metric.ExpiredCount += expired[queue.OverlayID]
		byOverlay[queue.OverlayID] = metric
	}
	for _, delivery := range deliveries {
		delivery = NormalizeReceiptDelivery(delivery)
		if err := delivery.Validate(); err != nil {
			return nil, err
		}
		overlayID := HashParts("receipt-overlay", delivery.Receipt.SourceZone, delivery.Receipt.DestinationZone)
		metric := byOverlay[overlayID]
		metric.OverlayID = overlayID
		metric.ReceiptCount++
		if delivery.State == ReceiptDeliveryAcknowledged {
			metric.DeliveredCount++
		}
		byOverlay[overlayID] = metric
	}
	for _, response := range queryResponses {
		response = NormalizeQueryResponseProof(response)
		if err := response.Validate(); err != nil {
			return nil, err
		}
		overlayID := HashParts("query-overlay", response.Responder)
		metric := byOverlay[overlayID]
		metric.OverlayID = overlayID
		metric.QueryProofCount++
		byOverlay[overlayID] = metric
	}
	out := make([]L3OverlayMetrics, 0, len(byOverlay))
	for _, metric := range byOverlay {
		out = append(out, metric)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].OverlayID < out[j].OverlayID
	})
	return out, nil
}

func sortAetherMeshMessages(messages []AetherMeshMessage) {
	sort.SliceStable(messages, func(i, j int) bool {
		left := NormalizeAetherMeshMessage(messages[i])
		right := NormalizeAetherMeshMessage(messages[j])
		if left.Priority != right.Priority {
			return left.Priority < right.Priority
		}
		leftDeadline := left.DeadlineHeight
		rightDeadline := right.DeadlineHeight
		if leftDeadline == 0 {
			leftDeadline = ^uint64(0)
		}
		if rightDeadline == 0 {
			rightDeadline = ^uint64(0)
		}
		if leftDeadline != rightDeadline {
			return leftDeadline < rightDeadline
		}
		if left.Sequence != right.Sequence {
			return left.Sequence < right.Sequence
		}
		return left.MessageID < right.MessageID
	})
}

func sortCrossZoneSequenceStates(states []CrossZoneSequenceState) {
	sort.SliceStable(states, func(i, j int) bool {
		if states[i].SourceZone != states[j].SourceZone {
			return states[i].SourceZone < states[j].SourceZone
		}
		return states[i].DestinationZone < states[j].DestinationZone
	})
}
