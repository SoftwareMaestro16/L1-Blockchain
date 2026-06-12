package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	CrossZoneMessagesPrefix	= "messages"
	CrossZoneOutboxPrefix	= CrossZoneMessagesPrefix + "/outbox"
	CrossZoneInboxPrefix	= CrossZoneMessagesPrefix + "/inbox"
	CrossZoneReceiptsPrefix	= CrossZoneMessagesPrefix + "/receipts"
	CrossZoneNoncesPrefix	= CrossZoneMessagesPrefix + "/nonces"
	CrossZoneReplayPrefix	= CrossZoneMessagesPrefix + "/replay"
	CrossZoneExpiryPrefix	= CrossZoneMessagesPrefix + "/expiry"
)

type CrossZoneQueueKind string
type CrossZoneReceiptStatus string

const (
	CrossZoneQueueOutbox	CrossZoneQueueKind	= "OUTBOX"
	CrossZoneQueueInbox	CrossZoneQueueKind	= "INBOX"

	CrossZoneReceiptQueued		CrossZoneReceiptStatus	= "queued"
	CrossZoneReceiptExecuted	CrossZoneReceiptStatus	= "executed"
	CrossZoneReceiptFailed		CrossZoneReceiptStatus	= "failed"
	CrossZoneReceiptExpired		CrossZoneReceiptStatus	= "expired"
	CrossZoneReceiptBounced		CrossZoneReceiptStatus	= "bounced"
	CrossZoneReceiptRejected	CrossZoneReceiptStatus	= "rejected"

	CrossZoneReceiptPending	CrossZoneReceiptStatus	= CrossZoneReceiptQueued
	CrossZoneReceiptSuccess	CrossZoneReceiptStatus	= CrossZoneReceiptExecuted
)

type CrossZoneQueueItem struct {
	Kind		CrossZoneQueueKind
	Message		CrossZoneMessageEnvelope
	EnqueuedHeight	uint64
}

type CrossZoneMessageReceipt struct {
	MessageID		[]byte
	SourceZone		zonestypes.ZoneID
	DestinationZone		zonestypes.ZoneID
	Sender			sdk.AccAddress
	Recipient		sdk.AccAddress
	Status			CrossZoneReceiptStatus
	GasUsed			uint64
	FeeCharged		sdkmath.Int
	ReturnPayloadHash	[]byte
	ErrorCode		uint32
	HasErrorCode		bool
	ExecutedHeight		uint64
	ResultHash		string
	Height			uint64
	SourceSequence		uint64
	Nonce			uint64
	ReceiptHash		string
}

type CrossZoneSenderNonce struct {
	SourceZone	zonestypes.ZoneID
	Sender		sdk.AccAddress
	Nonce		uint64
}

type CrossZoneReplayTombstone struct {
	MessageID	[]byte
	SourceZone	zonestypes.ZoneID
	Sender		sdk.AccAddress
	Nonce		uint64
	SourceSequence	uint64
	CreatedHeight	uint64
	TombstoneHeight	uint64
	ExpiryHeight	uint64
	TombstoneHash	string
}

type CrossZoneExpiryItem struct {
	Deadline	uint64
	MessageID	[]byte
}

type CrossZoneQueueState struct {
	Outbox		[]CrossZoneQueueItem
	Inbox		[]CrossZoneQueueItem
	Receipts	[]CrossZoneMessageReceipt
	Nonces		[]CrossZoneSenderNonce
	Replay		[]CrossZoneReplayTombstone
	Expiry		[]CrossZoneExpiryItem
	StateRoot	string
	Height		uint64
	ParamsHash	string
}

type CrossZoneQueueRoots struct {
	OutboxRoot	string
	InboxRoot	string
	ReceiptRoot	string
	NonceRoot	string
	ReplayRoot	string
	ExpiryRoot	string
	StateRoot	string
}

type CrossZoneRoutingRules struct {
	KernelMediated			bool
	MessageDrivenStateOnly		bool
	NonceScopedBySourceSender	bool
	SenderOrderedDestination	bool
	ReceiptExpiredMessages		bool
	BounceFailures			bool
}

type CrossZoneRoutingResult struct {
	State		CrossZoneQueueState
	Routed		[]CrossZoneQueueItem
	Receipts	[]CrossZoneMessageReceipt
	BounceMessages	[]CrossZoneMessageEnvelope
}

func DefaultCrossZoneRoutingRules() CrossZoneRoutingRules {
	return CrossZoneRoutingRules{
		KernelMediated:			true,
		MessageDrivenStateOnly:		true,
		NonceScopedBySourceSender:	true,
		SenderOrderedDestination:	true,
		ReceiptExpiredMessages:		true,
		BounceFailures:			true,
	}
}

func CrossZoneOutboxKey(sourceZone zonestypes.ZoneID, sender sdk.AccAddress, sequence uint64) (string, error) {
	if err := zonestypes.ValidateZoneID(sourceZone); err != nil {
		return "", err
	}
	if len(sender) == 0 {
		return "", errors.New("cross-zone outbox sender is required")
	}
	if sequence == 0 {
		return "", errors.New("cross-zone outbox sequence must be positive")
	}
	return CrossZoneOutboxPrefix + "/" + string(sourceZone) + "/" + hex.EncodeToString(sender) + "/" + fmt.Sprint(sequence), nil
}

func CrossZoneInboxKey(destinationZone zonestypes.ZoneID, sender sdk.AccAddress, sequence uint64) (string, error) {
	if err := zonestypes.ValidateZoneID(destinationZone); err != nil {
		return "", err
	}
	if len(sender) == 0 {
		return "", errors.New("cross-zone inbox sender is required")
	}
	if sequence == 0 {
		return "", errors.New("cross-zone inbox sequence must be positive")
	}
	return CrossZoneInboxPrefix + "/" + string(destinationZone) + "/" + hex.EncodeToString(sender) + "/" + fmt.Sprint(sequence), nil
}

func CrossZoneReceiptKey(messageID []byte) (string, error) {
	if len(messageID) != MessageIDBytes {
		return "", fmt.Errorf("cross-zone receipt message id must be %d bytes", MessageIDBytes)
	}
	return CrossZoneReceiptsPrefix + "/" + hex.EncodeToString(messageID), nil
}

func CrossZoneNonceKey(sourceZone zonestypes.ZoneID, sender sdk.AccAddress) (string, error) {
	if err := zonestypes.ValidateZoneID(sourceZone); err != nil {
		return "", err
	}
	if len(sender) == 0 {
		return "", errors.New("cross-zone nonce sender is required")
	}
	return CrossZoneNoncesPrefix + "/" + string(sourceZone) + "/" + hex.EncodeToString(sender), nil
}

func CrossZoneReplayKey(messageID []byte) (string, error) {
	if len(messageID) != MessageIDBytes {
		return "", fmt.Errorf("cross-zone replay message id must be %d bytes", MessageIDBytes)
	}
	return CrossZoneReplayPrefix + "/" + hex.EncodeToString(messageID), nil
}

func CrossZoneExpiryKey(deadline uint64, messageID []byte) (string, error) {
	if deadline == 0 {
		return "", errors.New("cross-zone expiry deadline must be positive")
	}
	if len(messageID) != MessageIDBytes {
		return "", fmt.Errorf("cross-zone expiry message id must be %d bytes", MessageIDBytes)
	}
	return CrossZoneExpiryPrefix + "/" + fmt.Sprint(deadline) + "/" + hex.EncodeToString(messageID), nil
}

func EnqueueCrossZoneOutbox(state CrossZoneQueueState, msg CrossZoneMessageEnvelope, params CrossZoneMessageParams) (CrossZoneQueueState, error) {
	if err := msg.Validate(params); err != nil {
		return CrossZoneQueueState{}, err
	}
	next := state.Normalize()
	if hasCrossZoneReplayTombstone(next.Replay, msg.MessageID) {
		return CrossZoneQueueState{}, errors.New("cross-zone message has replay tombstone")
	}
	if hasCrossZoneQueueMessage(next.Outbox, msg.MessageID) || hasCrossZoneQueueMessage(next.Inbox, msg.MessageID) {
		return CrossZoneQueueState{}, errors.New("cross-zone message already queued")
	}
	lastNonce := crossZoneNonceFor(next.Nonces, msg.SourceZone, msg.Sender)
	if msg.Nonce <= lastNonce {
		return CrossZoneQueueState{}, errors.New("cross-zone message nonce must increase per sender")
	}
	next.Nonces = upsertCrossZoneNonce(next.Nonces, CrossZoneSenderNonce{SourceZone: msg.SourceZone, Sender: msg.Sender, Nonce: msg.Nonce})
	next.Outbox = append(next.Outbox, CrossZoneQueueItem{Kind: CrossZoneQueueOutbox, Message: msg.Clone(), EnqueuedHeight: msg.CreatedHeight})
	next.Expiry = append(next.Expiry, CrossZoneExpiryItem{Deadline: msg.Deadline, MessageID: append([]byte(nil), msg.MessageID...)})
	return next.WithRoot(params)
}

func RouteCrossZoneOutboxToInbox(state CrossZoneQueueState, messageID []byte, height uint64, params CrossZoneMessageParams) (CrossZoneQueueState, CrossZoneQueueItem, error) {
	if height == 0 {
		return CrossZoneQueueState{}, CrossZoneQueueItem{}, errors.New("cross-zone route height must be positive")
	}
	next := state.Normalize()
	var routed CrossZoneQueueItem
	outbox := make([]CrossZoneQueueItem, 0, len(next.Outbox))
	for _, item := range next.Outbox {
		if bytesEqual(item.Message.MessageID, messageID) {
			routed = item.Clone()
			continue
		}
		outbox = append(outbox, item)
	}
	if len(routed.Message.MessageID) == 0 {
		return CrossZoneQueueState{}, CrossZoneQueueItem{}, errors.New("cross-zone outbox message not found")
	}
	routed.Kind = CrossZoneQueueInbox
	routed.EnqueuedHeight = height
	next.Outbox = outbox
	next.Inbox = append(next.Inbox, routed)
	next, err := next.WithRoot(params)
	return next, routed, err
}

func RouteCrossZoneOutboxViaKernel(state CrossZoneQueueState, messageID []byte, height uint64, params CrossZoneMessageParams, rules CrossZoneRoutingRules) (CrossZoneRoutingResult, error) {
	if err := rules.Validate(); err != nil {
		return CrossZoneRoutingResult{}, err
	}
	if height == 0 {
		return CrossZoneRoutingResult{}, errors.New("cross-zone kernel route height must be positive")
	}
	next := state.Normalize()
	item, found := findCrossZoneQueueMessage(next.Outbox, messageID)
	if !found {
		return CrossZoneRoutingResult{}, errors.New("cross-zone kernel route message not found")
	}
	if height > item.Message.Deadline {
		receipt := ExpiredCrossZoneReceiptFromMessage(item.Message, height)
		next, err := RecordCrossZoneReceipt(next, receipt, params)
		if err != nil {
			return CrossZoneRoutingResult{}, err
		}
		next.Outbox = removeCrossZoneQueueMessage(next.Outbox, messageID)
		next.Expiry = removeCrossZoneExpiry(next.Expiry, messageID)
		next, err = next.WithRoot(params)
		if err != nil {
			return CrossZoneRoutingResult{}, err
		}
		return CrossZoneRoutingResult{State: next, Receipts: []CrossZoneMessageReceipt{next.Receipts[len(next.Receipts)-1]}}, nil
	}
	next, routed, err := RouteCrossZoneOutboxToInbox(next, messageID, height, params)
	if err != nil {
		return CrossZoneRoutingResult{}, err
	}
	return CrossZoneRoutingResult{State: next, Routed: []CrossZoneQueueItem{routed}}, nil
}

func ApplyCrossZoneExpiryQueue(state CrossZoneQueueState, height uint64, max uint32, params CrossZoneMessageParams) (CrossZoneRoutingResult, error) {
	if height == 0 {
		return CrossZoneRoutingResult{}, errors.New("cross-zone expiry height must be positive")
	}
	if max == 0 {
		return CrossZoneRoutingResult{}, errors.New("cross-zone expiry max must be positive")
	}
	next := state.Normalize()
	receipts := make([]CrossZoneMessageReceipt, 0)
	for _, expiry := range next.Expiry {
		if uint32(len(receipts)) >= max || expiry.Deadline > height {
			continue
		}
		item, found := findCrossZoneQueueMessage(next.Inbox, expiry.MessageID)
		if !found {
			item, found = findCrossZoneQueueMessage(next.Outbox, expiry.MessageID)
		}
		if !found {
			next.Expiry = removeCrossZoneExpiry(next.Expiry, expiry.MessageID)
			continue
		}
		receipt := ExpiredCrossZoneReceiptFromMessage(item.Message, height)
		var err error
		next, err = RecordCrossZoneReceipt(next, receipt, params)
		if err != nil {
			return CrossZoneRoutingResult{}, err
		}
		next.Outbox = removeCrossZoneQueueMessage(next.Outbox, expiry.MessageID)
		next.Inbox = removeCrossZoneQueueMessage(next.Inbox, expiry.MessageID)
		next.Expiry = removeCrossZoneExpiry(next.Expiry, expiry.MessageID)
		receipts = append(receipts, next.Receipts[len(next.Receipts)-1])
	}
	next, err := next.WithRoot(params)
	if err != nil {
		return CrossZoneRoutingResult{}, err
	}
	return CrossZoneRoutingResult{State: next, Receipts: receipts}, nil
}

func BuildCrossZoneBounceMessage(original CrossZoneMessageEnvelope, receipt CrossZoneMessageReceipt, nonce uint64, sourceSequence uint64, height uint64, params CrossZoneMessageParams) (CrossZoneMessageEnvelope, error) {
	if err := original.Validate(params); err != nil {
		return CrossZoneMessageEnvelope{}, err
	}
	if err := receipt.Validate(); err != nil {
		return CrossZoneMessageEnvelope{}, err
	}
	if !bytesEqual(original.MessageID, receipt.MessageID) {
		return CrossZoneMessageEnvelope{}, errors.New("cross-zone bounce receipt message mismatch")
	}
	if height == 0 || nonce == 0 || sourceSequence == 0 {
		return CrossZoneMessageEnvelope{}, errors.New("cross-zone bounce height, nonce, and sequence must be positive")
	}
	payload := []byte(fmt.Sprintf(
		"message_id=%s;status=%s;error_code=%s;return_payload_hash=%s",
		hex.EncodeToString(original.MessageID),
		receipt.Status,
		receiptErrorCodeString(receipt),
		hex.EncodeToString(receipt.ReturnPayloadHash),
	))
	bounce := CrossZoneMessageEnvelope{
		SourceZone:		original.DestinationZone,
		DestinationZone:	original.SourceZone,
		Sender:			append(sdk.AccAddress(nil), original.Recipient...),
		Recipient:		append(sdk.AccAddress(nil), original.Sender...),
		Value:			original.Value,
		Opcode:			"aether.bounce",
		Payload:		payload,
		GasLimit:		original.GasLimit,
		Deadline:		original.Deadline,
		Nonce:			nonce,
		SourceSequence:		sourceSequence,
		RouteID:		"bounce/" + hex.EncodeToString(original.MessageID),
		Bounce:			false,
		FeeLimit:		original.FeeLimit,
		CreatedHeight:		height,
		AuthScope:		"bounce",
	}
	if bounce.Deadline < height {
		bounce.Deadline = height
	}
	return NewCrossZoneMessageEnvelope(bounce, params)
}

func RecordCrossZoneReceipt(state CrossZoneQueueState, receipt CrossZoneMessageReceipt, params CrossZoneMessageParams) (CrossZoneQueueState, error) {
	next := state.Normalize()
	receipt, err := NewCrossZoneMessageReceipt(receipt)
	if err != nil {
		return CrossZoneQueueState{}, err
	}
	if hasCrossZoneReceipt(next.Receipts, receipt.MessageID) {
		return CrossZoneQueueState{}, errors.New("cross-zone receipt already exists")
	}
	next.Receipts = append(next.Receipts, receipt)
	tombstone, err := NewCrossZoneReplayTombstone(CrossZoneReplayTombstone{
		MessageID:		receipt.MessageID,
		SourceZone:		receipt.SourceZone,
		Sender:			receipt.Sender,
		Nonce:			receipt.Nonce,
		SourceSequence:		receipt.SourceSequence,
		TombstoneHeight:	receipt.EffectiveHeight(),
		CreatedHeight:		receipt.EffectiveHeight(),
		ExpiryHeight:		receipt.EffectiveHeight(),
	})
	if err != nil {
		return CrossZoneQueueState{}, err
	}
	next.Replay = append(next.Replay, tombstone)
	next.Inbox = removeCrossZoneQueueMessage(next.Inbox, receipt.MessageID)
	next.Expiry = removeCrossZoneExpiry(next.Expiry, receipt.MessageID)
	return next.WithRoot(params)
}

func NewCrossZoneMessageReceipt(receipt CrossZoneMessageReceipt) (CrossZoneMessageReceipt, error) {
	if receipt.ReceiptHash != "" {
		return CrossZoneMessageReceipt{}, errors.New("cross-zone receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return CrossZoneMessageReceipt{}, err
	}
	receipt.ReceiptHash = ComputeCrossZoneReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func ReceiptFromCrossZoneMessage(msg CrossZoneMessageEnvelope, status CrossZoneReceiptStatus, gasUsed uint64, feeCharged sdkmath.Int, returnPayloadHash []byte, errorCode *uint32, executedHeight uint64) CrossZoneMessageReceipt {
	receipt := CrossZoneMessageReceipt{
		MessageID:		append([]byte(nil), msg.MessageID...),
		SourceZone:		msg.SourceZone,
		DestinationZone:	msg.DestinationZone,
		Sender:			append(sdk.AccAddress(nil), msg.Sender...),
		Recipient:		append(sdk.AccAddress(nil), msg.Recipient...),
		Status:			status,
		GasUsed:		gasUsed,
		FeeCharged:		feeCharged,
		ReturnPayloadHash:	append([]byte(nil), returnPayloadHash...),
		ExecutedHeight:		executedHeight,
		Height:			executedHeight,
		SourceSequence:		msg.SourceSequence,
		Nonce:			msg.Nonce,
	}
	if errorCode != nil {
		receipt.ErrorCode = *errorCode
		receipt.HasErrorCode = true
	}
	if len(receipt.ReturnPayloadHash) != 0 {
		receipt.ResultHash = hex.EncodeToString(receipt.ReturnPayloadHash)
	} else {
		receipt.ResultHash = hashCrossZoneQueueParts("empty-return-payload")
	}
	return receipt
}

func ExpiredCrossZoneReceiptFromMessage(msg CrossZoneMessageEnvelope, executedHeight uint64) CrossZoneMessageReceipt {
	code := uint32(1)
	return ReceiptFromCrossZoneMessage(msg, CrossZoneReceiptExpired, 0, sdkmath.ZeroInt(), nil, &code, executedHeight)
}

func NewCrossZoneReplayTombstone(tombstone CrossZoneReplayTombstone) (CrossZoneReplayTombstone, error) {
	if tombstone.TombstoneHash != "" {
		return CrossZoneReplayTombstone{}, errors.New("cross-zone replay tombstone hash must be empty before construction")
	}
	if err := tombstone.ValidateFormat(); err != nil {
		return CrossZoneReplayTombstone{}, err
	}
	tombstone.TombstoneHash = ComputeCrossZoneReplayTombstoneHash(tombstone)
	return tombstone, tombstone.Validate()
}

func (s CrossZoneQueueState) Normalize() CrossZoneQueueState {
	s.Outbox = normalizeCrossZoneQueueItems(s.Outbox, CrossZoneQueueOutbox)
	s.Inbox = normalizeCrossZoneQueueItems(s.Inbox, CrossZoneQueueInbox)
	s.Receipts = normalizeCrossZoneReceipts(s.Receipts)
	s.Nonces = normalizeCrossZoneNonces(s.Nonces)
	s.Replay = normalizeCrossZoneReplay(s.Replay)
	s.Expiry = normalizeCrossZoneExpiry(s.Expiry)
	return s
}

func (s CrossZoneQueueState) WithRoot(params CrossZoneMessageParams) (CrossZoneQueueState, error) {
	next := s.Normalize()
	next.StateRoot = ""
	if err := next.Validate(params); err != nil {
		return CrossZoneQueueState{}, err
	}
	next.StateRoot = ComputeCrossZoneQueueStateRoot(next, params)
	return next, next.Validate(params)
}

func (s CrossZoneQueueState) Validate(params CrossZoneMessageParams) error {
	normalized := s.Normalize()
	for _, item := range normalized.Outbox {
		if err := item.Validate(CrossZoneQueueOutbox, params); err != nil {
			return err
		}
	}
	for _, item := range normalized.Inbox {
		if err := item.Validate(CrossZoneQueueInbox, params); err != nil {
			return err
		}
	}
	for _, receipt := range normalized.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
	}
	for _, nonce := range normalized.Nonces {
		if err := nonce.Validate(); err != nil {
			return err
		}
	}
	for _, tombstone := range normalized.Replay {
		if err := tombstone.Validate(); err != nil {
			return err
		}
	}
	for _, expiry := range normalized.Expiry {
		if err := expiry.Validate(); err != nil {
			return err
		}
	}
	if normalized.StateRoot != "" && normalized.StateRoot != ComputeCrossZoneQueueStateRoot(normalized, params) {
		return errors.New("cross-zone queue state root mismatch")
	}
	if normalized.ParamsHash != "" {
		return zonestypes.ValidateHash("cross-zone queue params hash", normalized.ParamsHash)
	}
	if err := validateCrossZoneDestinationOrdering(normalized.Inbox); err != nil {
		return err
	}
	return nil
}

func (i CrossZoneQueueItem) Validate(expected CrossZoneQueueKind, params CrossZoneMessageParams) error {
	if i.Kind != expected {
		return errors.New("cross-zone queue item kind mismatch")
	}
	if err := i.Message.Validate(params); err != nil {
		return err
	}
	if i.EnqueuedHeight == 0 {
		return errors.New("cross-zone queue item enqueued height must be positive")
	}
	return nil
}

func (i CrossZoneQueueItem) Clone() CrossZoneQueueItem {
	i.Message = i.Message.Clone()
	return i
}

func (r CrossZoneMessageReceipt) ValidateFormat() error {
	if len(r.MessageID) != MessageIDBytes {
		return fmt.Errorf("cross-zone receipt message id must be %d bytes", MessageIDBytes)
	}
	if err := zonestypes.ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.DestinationZone); err != nil {
		return err
	}
	if len(r.Sender) == 0 || len(r.Recipient) == 0 {
		return errors.New("cross-zone receipt sender and recipient are required")
	}
	if !IsCrossZoneReceiptStatus(r.Status) {
		return fmt.Errorf("unknown cross-zone receipt status %q", r.Status)
	}
	if r.FeeCharged.IsNil() {
		r.FeeCharged = sdkmath.ZeroInt()
	}
	if r.FeeCharged.IsNegative() {
		return errors.New("cross-zone receipt fee charged must be non-negative")
	}
	if len(r.ReturnPayloadHash) != 0 && len(r.ReturnPayloadHash) != sha256.Size {
		return fmt.Errorf("cross-zone return payload hash must be %d bytes", sha256.Size)
	}
	if err := zonestypes.ValidateHash("cross-zone receipt result hash", r.ResultHash); err != nil {
		return err
	}
	if r.EffectiveHeight() == 0 || r.SourceSequence == 0 || r.Nonce == 0 {
		return errors.New("cross-zone receipt height, source sequence, and nonce must be positive")
	}
	if r.Height != 0 && r.ExecutedHeight != 0 && r.Height != r.ExecutedHeight {
		return errors.New("cross-zone receipt height and executed height mismatch")
	}
	if r.ReceiptHash != "" {
		return zonestypes.ValidateHash("cross-zone receipt hash", r.ReceiptHash)
	}
	return nil
}

func (r CrossZoneMessageReceipt) EffectiveHeight() uint64 {
	if r.ExecutedHeight != 0 {
		return r.ExecutedHeight
	}
	return r.Height
}

func (r CrossZoneMessageReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash == "" {
		return errors.New("cross-zone receipt hash is required")
	}
	if r.ReceiptHash != ComputeCrossZoneReceiptHash(r) {
		return errors.New("cross-zone receipt hash mismatch")
	}
	return nil
}

func (n CrossZoneSenderNonce) Validate() error {
	if _, err := CrossZoneNonceKey(n.SourceZone, n.Sender); err != nil {
		return err
	}
	if n.Nonce == 0 {
		return errors.New("cross-zone sender nonce must be positive")
	}
	return nil
}

func (t CrossZoneReplayTombstone) ValidateFormat() error {
	if _, err := CrossZoneReplayKey(t.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(t.SourceZone); err != nil {
		return err
	}
	if len(t.Sender) == 0 {
		return errors.New("cross-zone replay tombstone sender is required")
	}
	if t.Nonce == 0 || t.SourceSequence == 0 || t.CreatedHeight == 0 || t.TombstoneHeight == 0 || t.ExpiryHeight == 0 {
		return errors.New("cross-zone replay tombstone nonce, sequence, and heights must be positive")
	}
	if t.TombstoneHash != "" {
		return zonestypes.ValidateHash("cross-zone replay tombstone hash", t.TombstoneHash)
	}
	return nil
}

func (t CrossZoneReplayTombstone) Validate() error {
	if err := t.ValidateFormat(); err != nil {
		return err
	}
	if t.TombstoneHash == "" {
		return errors.New("cross-zone replay tombstone hash is required")
	}
	if t.TombstoneHash != ComputeCrossZoneReplayTombstoneHash(t) {
		return errors.New("cross-zone replay tombstone hash mismatch")
	}
	return nil
}

func (e CrossZoneExpiryItem) Validate() error {
	_, err := CrossZoneExpiryKey(e.Deadline, e.MessageID)
	return err
}

func IsCrossZoneReceiptStatus(status CrossZoneReceiptStatus) bool {
	switch status {
	case CrossZoneReceiptQueued, CrossZoneReceiptExecuted, CrossZoneReceiptFailed, CrossZoneReceiptExpired, CrossZoneReceiptBounced, CrossZoneReceiptRejected:
		return true
	default:
		return false
	}
}

func (r CrossZoneRoutingRules) Validate() error {
	if !r.KernelMediated {
		return errors.New("cross-zone routing must be mediated by Aether Core kernel")
	}
	if !r.MessageDrivenStateOnly {
		return errors.New("cross-zone routing must not perform direct state writes across zones")
	}
	if !r.NonceScopedBySourceSender {
		return errors.New("cross-zone nonces must be scoped by source zone and sender")
	}
	if !r.SenderOrderedDestination {
		return errors.New("cross-zone destination queues must apply sender order")
	}
	if !r.ReceiptExpiredMessages {
		return errors.New("cross-zone expired messages must be receipted")
	}
	if !r.BounceFailures {
		return errors.New("cross-zone failures must be bounce-capable")
	}
	return nil
}

func ComputeCrossZoneReceiptHash(receipt CrossZoneMessageReceipt) string {
	return hashCrossZoneQueueParts(
		"aether-cross-zone-receipt-v1",
		hex.EncodeToString(receipt.MessageID),
		string(receipt.SourceZone),
		string(receipt.DestinationZone),
		hex.EncodeToString(receipt.Sender),
		hex.EncodeToString(receipt.Recipient),
		string(receipt.Status),
		fmt.Sprint(receipt.GasUsed),
		receipt.FeeChargedString(),
		hex.EncodeToString(receipt.ReturnPayloadHash),
		receiptErrorCodeString(receipt),
		receipt.ResultHash,
		fmt.Sprint(receipt.EffectiveHeight()),
		fmt.Sprint(receipt.SourceSequence),
		fmt.Sprint(receipt.Nonce),
	)
}

func (r CrossZoneMessageReceipt) FeeChargedString() string {
	if r.FeeCharged.IsNil() {
		return sdkmath.ZeroInt().String()
	}
	return r.FeeCharged.String()
}

func ComputeCrossZoneReplayTombstoneHash(tombstone CrossZoneReplayTombstone) string {
	return hashCrossZoneQueueParts(
		"aether-cross-zone-replay-tombstone-v1",
		hex.EncodeToString(tombstone.MessageID),
		string(tombstone.SourceZone),
		hex.EncodeToString(tombstone.Sender),
		fmt.Sprint(tombstone.Nonce),
		fmt.Sprint(tombstone.SourceSequence),
		fmt.Sprint(tombstone.CreatedHeight),
		fmt.Sprint(tombstone.TombstoneHeight),
		fmt.Sprint(tombstone.ExpiryHeight),
	)
}

func ComputeCrossZoneQueueRoots(state CrossZoneQueueState, params CrossZoneMessageParams) (CrossZoneQueueRoots, error) {
	normalized := state.Normalize()
	if err := normalized.Validate(params); err != nil {
		return CrossZoneQueueRoots{}, err
	}
	roots := CrossZoneQueueRoots{
		OutboxRoot:	ComputeCrossZoneQueueRoot(CrossZoneQueueOutbox, normalized.Outbox, params),
		InboxRoot:	ComputeCrossZoneQueueRoot(CrossZoneQueueInbox, normalized.Inbox, params),
		ReceiptRoot:	ComputeCrossZoneReceiptQueueRoot(normalized.Receipts),
		NonceRoot:	ComputeCrossZoneNonceRoot(normalized.Nonces),
		ReplayRoot:	ComputeCrossZoneReplayRoot(normalized.Replay),
		ExpiryRoot:	ComputeCrossZoneExpiryRoot(normalized.Expiry),
	}
	roots.StateRoot = hashCrossZoneQueueParts("aether-cross-zone-queue-state-v1", roots.OutboxRoot, roots.InboxRoot, roots.ReceiptRoot, roots.NonceRoot, roots.ReplayRoot, roots.ExpiryRoot)
	return roots, nil
}

func ComputeCrossZoneQueueStateRoot(state CrossZoneQueueState, params CrossZoneMessageParams) string {
	roots := CrossZoneQueueRoots{
		OutboxRoot:	ComputeCrossZoneQueueRoot(CrossZoneQueueOutbox, state.Outbox, params),
		InboxRoot:	ComputeCrossZoneQueueRoot(CrossZoneQueueInbox, state.Inbox, params),
		ReceiptRoot:	ComputeCrossZoneReceiptQueueRoot(state.Receipts),
		NonceRoot:	ComputeCrossZoneNonceRoot(state.Nonces),
		ReplayRoot:	ComputeCrossZoneReplayRoot(state.Replay),
		ExpiryRoot:	ComputeCrossZoneExpiryRoot(state.Expiry),
	}
	return hashCrossZoneQueueParts("aether-cross-zone-queue-state-v1", roots.OutboxRoot, roots.InboxRoot, roots.ReceiptRoot, roots.NonceRoot, roots.ReplayRoot, roots.ExpiryRoot)
}

func ComputeCrossZoneQueueRoot(kind CrossZoneQueueKind, items []CrossZoneQueueItem, params CrossZoneMessageParams) string {
	ordered := normalizeCrossZoneQueueItems(items, kind)
	parts := []string{"aether-cross-zone-queue-root-v1", string(kind), fmt.Sprint(len(ordered))}
	for _, item := range ordered {
		parts = append(parts, hex.EncodeToString(item.Message.MessageID), fmt.Sprint(item.EnqueuedHeight), fmt.Sprint(item.Message.SourceSequence), fmt.Sprint(item.Message.Nonce))
	}
	return hashCrossZoneQueueParts(parts...)
}

func ComputeCrossZoneReceiptQueueRoot(receipts []CrossZoneMessageReceipt) string {
	ordered := normalizeCrossZoneReceipts(receipts)
	parts := []string{"aether-cross-zone-receipt-queue-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ReceiptHash)
	}
	return hashCrossZoneQueueParts(parts...)
}

func ComputeCrossZoneNonceRoot(nonces []CrossZoneSenderNonce) string {
	ordered := normalizeCrossZoneNonces(nonces)
	parts := []string{"aether-cross-zone-nonce-root-v1", fmt.Sprint(len(ordered))}
	for _, nonce := range ordered {
		parts = append(parts, string(nonce.SourceZone), hex.EncodeToString(nonce.Sender), fmt.Sprint(nonce.Nonce))
	}
	return hashCrossZoneQueueParts(parts...)
}

func ComputeCrossZoneReplayRoot(tombstones []CrossZoneReplayTombstone) string {
	ordered := normalizeCrossZoneReplay(tombstones)
	parts := []string{"aether-cross-zone-replay-root-v1", fmt.Sprint(len(ordered))}
	for _, tombstone := range ordered {
		parts = append(parts, tombstone.TombstoneHash)
	}
	return hashCrossZoneQueueParts(parts...)
}

func ComputeCrossZoneExpiryRoot(expiry []CrossZoneExpiryItem) string {
	ordered := normalizeCrossZoneExpiry(expiry)
	parts := []string{"aether-cross-zone-expiry-root-v1", fmt.Sprint(len(ordered))}
	for _, item := range ordered {
		parts = append(parts, fmt.Sprint(item.Deadline), hex.EncodeToString(item.MessageID))
	}
	return hashCrossZoneQueueParts(parts...)
}

func normalizeCrossZoneQueueItems(items []CrossZoneQueueItem, kind CrossZoneQueueKind) []CrossZoneQueueItem {
	out := make([]CrossZoneQueueItem, 0, len(items))
	for _, item := range items {
		item = item.Clone()
		item.Kind = kind
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool { return compareCrossZoneQueueItems(out[i], out[j]) < 0 })
	return out
}

func normalizeCrossZoneReceipts(receipts []CrossZoneMessageReceipt) []CrossZoneMessageReceipt {
	out := append([]CrossZoneMessageReceipt(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		if out[i].SourceSequence != out[j].SourceSequence {
			return out[i].SourceSequence < out[j].SourceSequence
		}
		return hex.EncodeToString(out[i].MessageID) < hex.EncodeToString(out[j].MessageID)
	})
	return out
}

func normalizeCrossZoneNonces(nonces []CrossZoneSenderNonce) []CrossZoneSenderNonce {
	out := append([]CrossZoneSenderNonce(nil), nonces...)
	sort.SliceStable(out, func(i, j int) bool {
		left, _ := CrossZoneNonceKey(out[i].SourceZone, out[i].Sender)
		right, _ := CrossZoneNonceKey(out[j].SourceZone, out[j].Sender)
		return left < right
	})
	return out
}

func normalizeCrossZoneReplay(tombstones []CrossZoneReplayTombstone) []CrossZoneReplayTombstone {
	out := append([]CrossZoneReplayTombstone(nil), tombstones...)
	sort.SliceStable(out, func(i, j int) bool {
		return hex.EncodeToString(out[i].MessageID) < hex.EncodeToString(out[j].MessageID)
	})
	return out
}

func normalizeCrossZoneExpiry(expiry []CrossZoneExpiryItem) []CrossZoneExpiryItem {
	out := append([]CrossZoneExpiryItem(nil), expiry...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Deadline != out[j].Deadline {
			return out[i].Deadline < out[j].Deadline
		}
		return hex.EncodeToString(out[i].MessageID) < hex.EncodeToString(out[j].MessageID)
	})
	return out
}

func compareCrossZoneQueueItems(left, right CrossZoneQueueItem) int {
	return compareCrossZoneMessages(left.Message, right.Message)
}

func validateCrossZoneDestinationOrdering(inbox []CrossZoneQueueItem) error {
	ordered := normalizeCrossZoneQueueItems(inbox, CrossZoneQueueInbox)
	for i, item := range ordered {
		if i == 0 {
			continue
		}
		prev := ordered[i-1].Message
		cur := item.Message
		if prev.DestinationZone == cur.DestinationZone && bytesEqual(prev.Sender, cur.Sender) && prev.SourceSequence >= cur.SourceSequence {
			return errors.New("cross-zone destination inbox must be sender ordered")
		}
	}
	return nil
}

func upsertCrossZoneNonce(nonces []CrossZoneSenderNonce, update CrossZoneSenderNonce) []CrossZoneSenderNonce {
	out := append([]CrossZoneSenderNonce(nil), nonces...)
	for i := range out {
		if out[i].SourceZone == update.SourceZone && bytesEqual(out[i].Sender, update.Sender) {
			out[i] = update
			return normalizeCrossZoneNonces(out)
		}
	}
	out = append(out, update)
	return normalizeCrossZoneNonces(out)
}

func crossZoneNonceFor(nonces []CrossZoneSenderNonce, sourceZone zonestypes.ZoneID, sender sdk.AccAddress) uint64 {
	for _, nonce := range nonces {
		if nonce.SourceZone == sourceZone && bytesEqual(nonce.Sender, sender) {
			return nonce.Nonce
		}
	}
	return 0
}

func hasCrossZoneQueueMessage(items []CrossZoneQueueItem, messageID []byte) bool {
	for _, item := range items {
		if bytesEqual(item.Message.MessageID, messageID) {
			return true
		}
	}
	return false
}

func findCrossZoneQueueMessage(items []CrossZoneQueueItem, messageID []byte) (CrossZoneQueueItem, bool) {
	for _, item := range items {
		if bytesEqual(item.Message.MessageID, messageID) {
			return item.Clone(), true
		}
	}
	return CrossZoneQueueItem{}, false
}

func removeCrossZoneQueueMessage(items []CrossZoneQueueItem, messageID []byte) []CrossZoneQueueItem {
	out := make([]CrossZoneQueueItem, 0, len(items))
	for _, item := range items {
		if !bytesEqual(item.Message.MessageID, messageID) {
			out = append(out, item)
		}
	}
	return out
}

func hasCrossZoneReceipt(receipts []CrossZoneMessageReceipt, messageID []byte) bool {
	for _, receipt := range receipts {
		if bytesEqual(receipt.MessageID, messageID) {
			return true
		}
	}
	return false
}

func hasCrossZoneReplayTombstone(tombstones []CrossZoneReplayTombstone, messageID []byte) bool {
	for _, tombstone := range tombstones {
		if bytesEqual(tombstone.MessageID, messageID) {
			return true
		}
	}
	return false
}

func removeCrossZoneExpiry(expiry []CrossZoneExpiryItem, messageID []byte) []CrossZoneExpiryItem {
	out := make([]CrossZoneExpiryItem, 0, len(expiry))
	for _, item := range expiry {
		if !bytesEqual(item.MessageID, messageID) {
			out = append(out, item)
		}
	}
	return out
}

func bytesEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func hashCrossZoneQueueParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		writeCrossZoneString(h.Write, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func receiptErrorCodeString(receipt CrossZoneMessageReceipt) string {
	if !receipt.HasErrorCode {
		return ""
	}
	return fmt.Sprint(receipt.ErrorCode)
}
