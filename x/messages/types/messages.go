package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	CanonicalMessageEncodingVersion	= uint16(1)
	MessageIDBytes			= 32
	MaxChainIDLength		= 96
	MaxOpcodeLength			= 96
	MaxAuthScopeLength		= 96
	MaxRouteIDLength		= 128
	MaxProofPathItems		= 256

	MessagesPrefix		= "messages"
	MessagesOutboxPrefix	= MessagesPrefix + "/outbox"
	MessagesInboxPrefix	= MessagesPrefix + "/inbox"
	MessagesReceiptPrefix	= MessagesPrefix + "/receipts"
	MessagesNoncePrefix	= MessagesPrefix + "/nonces"
	MessagesReplayPrefix	= MessagesPrefix + "/replay"
	MessagesExpiryPrefix	= MessagesPrefix + "/expiry"
)

type MessageStatus string
type MessageProofKind string

const (
	MessageStatusQueued	MessageStatus	= "queued"
	MessageStatusExecuted	MessageStatus	= "executed"
	MessageStatusFailed	MessageStatus	= "failed"
	MessageStatusExpired	MessageStatus	= "expired"
	MessageStatusBounced	MessageStatus	= "bounced"
	MessageStatusRejected	MessageStatus	= "rejected"

	MessageProofInclusion	MessageProofKind	= "message_inclusion"
	MessageProofReceipt	MessageProofKind	= "execution_receipt"
)

type MessageParams struct {
	ChainID			string
	MaxPayloadSize		uint32
	MinGasLimit		uint64
	MaxGasLimit		uint64
	MinFeeLimit		sdkmath.Int
	ProofHorizon		uint64
	MaxDrainPerBlock	uint32
	BounceGasReserve	uint64
	ParamsHash		string
}

type Message struct {
	MessageID	[]byte
	ChainID		string
	SourceZone	zonestypes.ZoneID
	DestinationZone	zonestypes.ZoneID
	Sender		sdk.AccAddress
	Recipient	sdk.AccAddress
	Value		sdkmath.Int
	Opcode		string
	Payload		[]byte
	GasLimit	uint64
	Deadline	uint64
	Nonce		uint64
	SourceSequence	uint64
	RouteID		string
	Bounce		bool
	FeeLimit	sdkmath.Int
	CreatedHeight	uint64
	PayloadHash	[]byte
	AuthScope	string
}

type MessageReceipt struct {
	MessageID		[]byte
	SourceZone		zonestypes.ZoneID
	DestinationZone		zonestypes.ZoneID
	Status			MessageStatus
	GasUsed			uint64
	FeeCharged		sdkmath.Int
	ReturnPayloadHash	[]byte
	ErrorCode		uint32
	HasErrorCode		bool
	ExecutedHeight		uint64
	ReceiptHash		[]byte
}

type ReplayTombstone struct {
	MessageID	[]byte
	SourceZone	zonestypes.ZoneID
	Sender		sdk.AccAddress
	Nonce		uint64
	SourceSequence	uint64
	ConsumedHeight	uint64
	RetainUntil	uint64
	TombstoneHash	[]byte
}

type QueueItem struct {
	Message		Message
	EnqueuedHeight	uint64
}

type SenderNonce struct {
	SourceZone	zonestypes.ZoneID
	Sender		sdk.AccAddress
	Nonce		uint64
}

type ExpiryItem struct {
	Deadline	uint64
	MessageID	[]byte
}

type KeeperState struct {
	Outbox		[]QueueItem
	Inbox		[]QueueItem
	Receipts	[]MessageReceipt
	Nonces		[]SenderNonce
	Tombstones	[]ReplayTombstone
	Expiry		[]ExpiryItem
	Height		uint64
	Params		MessageParams
	StateRoot	string
}

type MessageRoots struct {
	OutboxRoot	string
	InboxRoot	string
	MessageRoot	string
	ReceiptRoot	string
	NonceRoot	string
	TombstoneRoot	string
	ExpiryRoot	string
	StateRoot	string
	ParamsHash	string
}

type MsgSubmitCrossZoneMessage struct {
	Message Message
}

type SubmitCrossZoneMessageResponse struct {
	MessageID	[]byte
	MessageRoot	string
	OutboxKey	string
}

type QueryMessageRequest struct {
	MessageID []byte
}

type QueryMessageResponse struct {
	Message	Message
	Found	bool
}

type QueryReceiptRequest struct {
	MessageID []byte
}

type QueryReceiptResponse struct {
	Receipt	MessageReceipt
	Found	bool
}

type QueryProofRequest struct {
	Kind		MessageProofKind
	MessageID	[]byte
	Root		string
	Limit		uint32
}

type QueryProofResponse struct {
	Kind		MessageProofKind
	MessageID	[]byte
	Root		string
	ValueHash	string
	Path		[]string
	ProofHash	string
}

type MsgServer interface {
	SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage) (SubmitCrossZoneMessageResponse, error)
	SubmitCrossZoneCall(MsgCrossZoneCall, CrossZoneCallAdmission) (SubmitCrossZoneMessageResponse, error)
}

type QueryServer interface {
	Message(QueryMessageRequest) (QueryMessageResponse, error)
	Receipt(QueryReceiptRequest) (QueryReceiptResponse, error)
	MessageProof(QueryProofRequest) (QueryProofResponse, error)
}

type MessageKeeper struct {
	state KeeperState
}

type KeeperMsgServer struct {
	keeper *MessageKeeper
}

func DefaultMessageParams(chainID string) MessageParams {
	return MessageParams{
		ChainID:		chainID,
		MaxPayloadSize:		1024 * 1024,
		MinGasLimit:		1,
		MaxGasLimit:		100_000_000,
		MinFeeLimit:		sdkmath.ZeroInt(),
		ProofHorizon:		10_000,
		MaxDrainPerBlock:	1024,
		BounceGasReserve:	1,
		ParamsHash:		EmptyHash(),
	}
}

func NewMessageKeeper(params MessageParams) (MessageKeeper, error) {
	if err := params.Validate(); err != nil {
		return MessageKeeper{}, err
	}
	state := KeeperState{Params: params}
	state.StateRoot = ComputeKeeperStateRoot(state)
	return MessageKeeper{state: state}, nil
}

func NewMsgServer(keeper *MessageKeeper) (KeeperMsgServer, error) {
	if keeper == nil {
		return KeeperMsgServer{}, errors.New("message keeper is required")
	}
	return KeeperMsgServer{keeper: keeper}, nil
}

func NewQueryServer(keeper MessageKeeper) QueryServer {
	return keeper
}

func (s KeeperMsgServer) SubmitCrossZoneMessage(req MsgSubmitCrossZoneMessage) (SubmitCrossZoneMessageResponse, error) {
	if s.keeper == nil {
		return SubmitCrossZoneMessageResponse{}, errors.New("message keeper is required")
	}
	next, resp, err := s.keeper.SubmitCrossZoneMessage(req)
	if err != nil {
		return SubmitCrossZoneMessageResponse{}, err
	}
	*s.keeper = next
	return resp, nil
}

func (k MessageKeeper) State() KeeperState {
	return k.state.Normalize()
}

func (k MessageKeeper) SubmitCrossZoneMessage(req MsgSubmitCrossZoneMessage) (MessageKeeper, SubmitCrossZoneMessageResponse, error) {
	msg, err := NewMessage(req.Message, k.state.Params)
	if err != nil {
		return MessageKeeper{}, SubmitCrossZoneMessageResponse{}, err
	}
	next := k.state.Normalize()
	if hasTombstone(next.Tombstones, msg.MessageID) || hasQueueMessage(next.Outbox, msg.MessageID) || hasQueueMessage(next.Inbox, msg.MessageID) || hasReceipt(next.Receipts, msg.MessageID) {
		return MessageKeeper{}, SubmitCrossZoneMessageResponse{}, errors.New("message id must be unique")
	}
	if msg.Nonce <= nonceFor(next.Nonces, msg.SourceZone, msg.Sender) {
		return MessageKeeper{}, SubmitCrossZoneMessageResponse{}, errors.New("sender nonce cannot be reused")
	}
	next.Nonces = upsertNonce(next.Nonces, SenderNonce{SourceZone: msg.SourceZone, Sender: msg.Sender, Nonce: msg.Nonce})
	next.Outbox = append(next.Outbox, QueueItem{Message: msg.Clone(), EnqueuedHeight: msg.CreatedHeight})
	next.Expiry = append(next.Expiry, ExpiryItem{Deadline: msg.Deadline, MessageID: append([]byte(nil), msg.MessageID...)})
	next.StateRoot = ComputeKeeperStateRoot(next)
	outboxKey, err := OutboxKey(msg.SourceZone, msg.Sender, msg.SourceSequence)
	if err != nil {
		return MessageKeeper{}, SubmitCrossZoneMessageResponse{}, err
	}
	return MessageKeeper{state: next.Normalize()}, SubmitCrossZoneMessageResponse{
		MessageID:	append([]byte(nil), msg.MessageID...),
		MessageRoot:	ComputeMessageRoot(next.Outbox, next.Inbox),
		OutboxKey:	outboxKey,
	}, nil
}

func (k MessageKeeper) RouteOutboxToInbox(messageID []byte, height uint64) (MessageKeeper, error) {
	if height == 0 {
		return MessageKeeper{}, errors.New("route height must be positive")
	}
	next := k.state.Normalize()
	item, found := findQueueMessage(next.Outbox, messageID)
	if !found {
		return MessageKeeper{}, errors.New("outbox message not found")
	}
	next.Outbox = removeQueueMessage(next.Outbox, messageID)
	item.EnqueuedHeight = height
	next.Inbox = append(next.Inbox, item)
	next.StateRoot = ComputeKeeperStateRoot(next)
	return MessageKeeper{state: next.Normalize()}, nil
}

func (k MessageKeeper) DrainInbox(height uint64, max uint32, executor func(Message) (MessageReceipt, error)) (MessageKeeper, []MessageReceipt, error) {
	if height == 0 {
		return MessageKeeper{}, nil, errors.New("drain height must be positive")
	}
	if max == 0 || max > k.state.Params.MaxDrainPerBlock {
		max = k.state.Params.MaxDrainPerBlock
	}
	next := k.state.Normalize()
	ready := normalizeQueueItems(next.Inbox)
	receipts := make([]MessageReceipt, 0, max)
	for _, item := range ready {
		if uint32(len(receipts)) >= max {
			break
		}
		if item.Message.Deadline < height {
			receipt := ExpiredReceiptFromMessage(item.Message, height)
			var err error
			next, receipt, err = appendReceiptAndTombstone(next, receipt)
			if err != nil {
				return MessageKeeper{}, nil, err
			}
			next.Inbox = removeQueueMessage(next.Inbox, item.Message.MessageID)
			next.Expiry = removeExpiry(next.Expiry, item.Message.MessageID)
			receipts = append(receipts, receipt)
			continue
		}
		if executor == nil {
			return MessageKeeper{}, nil, errors.New("message executor is required")
		}
		receipt, err := executor(item.Message.Clone())
		if err != nil {
			return MessageKeeper{}, nil, err
		}
		receipt.MessageID = append([]byte(nil), item.Message.MessageID...)
		receipt.SourceZone = item.Message.SourceZone
		receipt.DestinationZone = item.Message.DestinationZone
		if receipt.ExecutedHeight == 0 {
			receipt.ExecutedHeight = height
		}
		next, receipt, err = appendReceiptAndTombstone(next, receipt)
		if err != nil {
			return MessageKeeper{}, nil, err
		}
		next.Inbox = removeQueueMessage(next.Inbox, item.Message.MessageID)
		next.Expiry = removeExpiry(next.Expiry, item.Message.MessageID)
		receipts = append(receipts, receipt)
	}
	next.StateRoot = ComputeKeeperStateRoot(next)
	return MessageKeeper{state: next.Normalize()}, receipts, nil
}

func (k MessageKeeper) ProcessExpiry(height uint64, max uint32) (MessageKeeper, []MessageReceipt, error) {
	if height == 0 {
		return MessageKeeper{}, nil, errors.New("expiry height must be positive")
	}
	if max == 0 || max > k.state.Params.MaxDrainPerBlock {
		max = k.state.Params.MaxDrainPerBlock
	}
	next := k.state.Normalize()
	receipts := make([]MessageReceipt, 0, max)
	for _, expiry := range normalizeExpiry(next.Expiry) {
		if uint32(len(receipts)) >= max || expiry.Deadline > height {
			continue
		}
		item, found := findQueueMessage(next.Inbox, expiry.MessageID)
		if !found {
			item, found = findQueueMessage(next.Outbox, expiry.MessageID)
		}
		if !found {
			next.Expiry = removeExpiry(next.Expiry, expiry.MessageID)
			continue
		}
		receipt := ExpiredReceiptFromMessage(item.Message, height)
		var err error
		next, receipt, err = appendReceiptAndTombstone(next, receipt)
		if err != nil {
			return MessageKeeper{}, nil, err
		}
		next.Inbox = removeQueueMessage(next.Inbox, item.Message.MessageID)
		next.Outbox = removeQueueMessage(next.Outbox, item.Message.MessageID)
		next.Expiry = removeExpiry(next.Expiry, item.Message.MessageID)
		receipts = append(receipts, receipt)
	}
	next.StateRoot = ComputeKeeperStateRoot(next)
	return MessageKeeper{state: next.Normalize()}, receipts, nil
}

func (k MessageKeeper) PruneTombstones(height uint64) MessageKeeper {
	next := k.state.Normalize()
	kept := make([]ReplayTombstone, 0, len(next.Tombstones))
	for _, tombstone := range next.Tombstones {
		if tombstone.RetainUntil >= height {
			kept = append(kept, tombstone)
		}
	}
	next.Tombstones = normalizeTombstones(kept)
	next.StateRoot = ComputeKeeperStateRoot(next)
	return MessageKeeper{state: next}
}

func (k MessageKeeper) BuildBounce(original Message, receipt MessageReceipt, nonce uint64, sourceSequence uint64, height uint64) (Message, error) {
	if err := original.Validate(k.state.Params); err != nil {
		return Message{}, err
	}
	if err := receipt.Validate(); err != nil {
		return Message{}, err
	}
	if !bytes.Equal(original.MessageID, receipt.MessageID) {
		return Message{}, errors.New("bounce receipt message mismatch")
	}
	payload := []byte(fmt.Sprintf(
		"message_id=%s;status=%s;error_code=%s;return_payload_hash=%s",
		hex.EncodeToString(original.MessageID),
		receipt.Status,
		receiptErrorCodeString(receipt),
		hex.EncodeToString(receipt.ReturnPayloadHash),
	))
	bounce := Message{
		ChainID:		original.ChainID,
		SourceZone:		original.DestinationZone,
		DestinationZone:	original.SourceZone,
		Sender:			append(sdk.AccAddress(nil), original.Recipient...),
		Recipient:		append(sdk.AccAddress(nil), original.Sender...),
		Value:			original.Value,
		Opcode:			"aether.bounce",
		Payload:		payload,
		GasLimit:		original.GasLimit + k.state.Params.BounceGasReserve,
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
	return NewMessage(bounce, k.state.Params)
}

func (k MessageKeeper) Message(req QueryMessageRequest) (QueryMessageResponse, error) {
	if len(req.MessageID) != MessageIDBytes {
		return QueryMessageResponse{}, fmt.Errorf("query message id must be %d bytes", MessageIDBytes)
	}
	if item, found := findQueueMessage(k.state.Inbox, req.MessageID); found {
		return QueryMessageResponse{Message: item.Message.Clone(), Found: true}, nil
	}
	if item, found := findQueueMessage(k.state.Outbox, req.MessageID); found {
		return QueryMessageResponse{Message: item.Message.Clone(), Found: true}, nil
	}
	return QueryMessageResponse{}, nil
}

func (k MessageKeeper) Receipt(req QueryReceiptRequest) (QueryReceiptResponse, error) {
	if len(req.MessageID) != MessageIDBytes {
		return QueryReceiptResponse{}, fmt.Errorf("query receipt message id must be %d bytes", MessageIDBytes)
	}
	for _, receipt := range k.state.Receipts {
		if bytes.Equal(receipt.MessageID, req.MessageID) {
			return QueryReceiptResponse{Receipt: receipt.Clone(), Found: true}, nil
		}
	}
	return QueryReceiptResponse{}, nil
}

func (k MessageKeeper) MessageProof(req QueryProofRequest) (QueryProofResponse, error) {
	if err := req.Validate(); err != nil {
		return QueryProofResponse{}, err
	}
	var path []string
	var valueHash string
	switch req.Kind {
	case MessageProofInclusion:
		items := normalizeQueueItems(append(k.state.Outbox, k.state.Inbox...))
		for _, item := range items {
			itemHash := QueueItemHash(item)
			path = append(path, itemHash)
			if bytes.Equal(item.Message.MessageID, req.MessageID) {
				valueHash = itemHash
			}
		}
	case MessageProofReceipt:
		for _, receipt := range normalizeReceipts(k.state.Receipts) {
			itemHash := hex.EncodeToString(receipt.ReceiptHash)
			path = append(path, itemHash)
			if bytes.Equal(receipt.MessageID, req.MessageID) {
				valueHash = itemHash
			}
		}
	default:
		return QueryProofResponse{}, fmt.Errorf("unknown message proof kind %q", req.Kind)
	}
	if valueHash == "" {
		return QueryProofResponse{}, errors.New("proof item not found")
	}
	if req.Limit != 0 && len(path) > int(req.Limit) {
		return QueryProofResponse{}, errors.New("proof path exceeds limit")
	}
	proof := QueryProofResponse{
		Kind:		req.Kind,
		MessageID:	append([]byte(nil), req.MessageID...),
		Root:		req.Root,
		ValueHash:	valueHash,
		Path:		path,
	}
	proof.ProofHash = ComputeProofHash(proof)
	return proof, proof.ValidateFor(req)
}

func NewMessage(msg Message, params MessageParams) (Message, error) {
	if params == (MessageParams{}) {
		params = DefaultMessageParams(msg.ChainID)
	}
	if len(msg.MessageID) != 0 {
		return Message{}, errors.New("message id must be empty before construction")
	}
	if len(msg.PayloadHash) != 0 {
		return Message{}, errors.New("payload hash must be empty before construction")
	}
	msg.ChainID = params.ChainID
	msg.Sender = append(sdk.AccAddress(nil), msg.Sender...)
	msg.Recipient = append(sdk.AccAddress(nil), msg.Recipient...)
	msg.Payload = append([]byte(nil), msg.Payload...)
	msg.PayloadHash = ComputePayloadHash(msg.Payload)
	msg.MessageID = DeriveMessageID(params.ChainID, msg.SourceZone, msg.Sender, msg.Nonce, msg.PayloadHash)
	return msg, msg.Validate(params)
}

func DeriveMessageID(chainID string, sourceZone zonestypes.ZoneID, sender sdk.AccAddress, nonce uint64, payloadHash []byte) []byte {
	buf := bytes.NewBuffer(nil)
	writeString(buf.Write, chainID)
	writeString(buf.Write, string(sourceZone))
	writeBytes(buf.Write, sender)
	writeU64(buf.Write, nonce)
	writeBytes(buf.Write, payloadHash)
	sum := sha256.Sum256(buf.Bytes())
	return append([]byte(nil), sum[:]...)
}

func ComputePayloadHash(payload []byte) []byte {
	sum := sha256.Sum256(payload)
	return append([]byte(nil), sum[:]...)
}

func CanonicalMessageBinary(msg Message) ([]byte, error) {
	if err := msg.Validate(DefaultMessageParams(msg.ChainID)); err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	writeU32(buf.Write, uint32(CanonicalMessageEncodingVersion))
	writeBytes(buf.Write, msg.MessageID)
	writeString(buf.Write, msg.ChainID)
	writeString(buf.Write, string(msg.SourceZone))
	writeString(buf.Write, string(msg.DestinationZone))
	writeBytes(buf.Write, msg.Sender)
	writeBytes(buf.Write, msg.Recipient)
	writeString(buf.Write, msg.Value.String())
	writeString(buf.Write, msg.Opcode)
	writeBytes(buf.Write, msg.Payload)
	writeU64(buf.Write, msg.GasLimit)
	writeU64(buf.Write, msg.Deadline)
	writeU64(buf.Write, msg.Nonce)
	writeU64(buf.Write, msg.SourceSequence)
	writeString(buf.Write, msg.RouteID)
	writeBool(buf.Write, msg.Bounce)
	writeString(buf.Write, msg.FeeLimit.String())
	writeU64(buf.Write, msg.CreatedHeight)
	writeBytes(buf.Write, msg.PayloadHash)
	writeString(buf.Write, msg.AuthScope)
	return buf.Bytes(), nil
}

func (m Message) Validate(params MessageParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if len(m.MessageID) != MessageIDBytes {
		return fmt.Errorf("message id must be %d bytes", MessageIDBytes)
	}
	expected := DeriveMessageID(params.ChainID, m.SourceZone, m.Sender, m.Nonce, m.PayloadHash)
	if !bytes.Equal(m.MessageID, expected) {
		return errors.New("message id mismatch")
	}
	if m.ChainID != params.ChainID {
		return errors.New("message chain id mismatch")
	}
	if err := zonestypes.ValidateZoneID(m.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.DestinationZone); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("message sender", m.Sender); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("message recipient", m.Recipient); err != nil {
		return err
	}
	if m.Value.IsNil() || m.Value.IsNegative() {
		return errors.New("message value must be non-negative")
	}
	if len(m.Payload) > int(params.MaxPayloadSize) {
		return fmt.Errorf("message payload size must be <= %d", params.MaxPayloadSize)
	}
	if !bytes.Equal(m.PayloadHash, ComputePayloadHash(m.Payload)) {
		return errors.New("message payload hash mismatch")
	}
	if err := validateToken("message opcode", m.Opcode, MaxOpcodeLength); err != nil {
		return err
	}
	if err := validateToken("message route id", m.RouteID, MaxRouteIDLength); err != nil {
		return err
	}
	if err := validateToken("message auth scope", m.AuthScope, MaxAuthScopeLength); err != nil {
		return err
	}
	if m.GasLimit < params.MinGasLimit || m.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("message gas limit must be in %d..%d", params.MinGasLimit, params.MaxGasLimit)
	}
	if m.CreatedHeight == 0 {
		return errors.New("message created height must be positive")
	}
	if m.Deadline == 0 || m.Deadline < m.CreatedHeight {
		return errors.New("message deadline must be at or after created height")
	}
	if m.Nonce == 0 || m.SourceSequence == 0 {
		return errors.New("message nonce and source sequence must be positive")
	}
	if m.FeeLimit.IsNil() || m.FeeLimit.IsNegative() || m.FeeLimit.LT(params.MinFeeLimit) {
		return errors.New("message fee limit is below minimum")
	}
	return nil
}

func (m Message) Clone() Message {
	m.MessageID = append([]byte(nil), m.MessageID...)
	m.Sender = append(sdk.AccAddress(nil), m.Sender...)
	m.Recipient = append(sdk.AccAddress(nil), m.Recipient...)
	m.Payload = append([]byte(nil), m.Payload...)
	m.PayloadHash = append([]byte(nil), m.PayloadHash...)
	return m
}

func (p MessageParams) Validate() error {
	if err := validateToken("message chain id", p.ChainID, MaxChainIDLength); err != nil {
		return err
	}
	if p.MaxPayloadSize == 0 {
		return errors.New("message max payload size must be positive")
	}
	if p.MinGasLimit == 0 || p.MaxGasLimit < p.MinGasLimit {
		return errors.New("message gas limits are invalid")
	}
	if p.MinFeeLimit.IsNil() || p.MinFeeLimit.IsNegative() {
		return errors.New("message min fee limit must be non-negative")
	}
	if p.ProofHorizon == 0 {
		return errors.New("message proof horizon must be positive")
	}
	if p.MaxDrainPerBlock == 0 {
		return errors.New("message max drain per block must be positive")
	}
	if p.ParamsHash != "" {
		return zonestypes.ValidateHash("message params hash", p.ParamsHash)
	}
	return nil
}

func NewMessageReceipt(receipt MessageReceipt) (MessageReceipt, error) {
	if len(receipt.ReceiptHash) != 0 {
		return MessageReceipt{}, errors.New("receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return MessageReceipt{}, err
	}
	receipt.ReceiptHash = ComputeReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func ReceiptFromMessage(msg Message, status MessageStatus, gasUsed uint64, feeCharged sdkmath.Int, returnPayloadHash []byte, errorCode *uint32, executedHeight uint64) MessageReceipt {
	receipt := MessageReceipt{
		MessageID:		append([]byte(nil), msg.MessageID...),
		SourceZone:		msg.SourceZone,
		DestinationZone:	msg.DestinationZone,
		Status:			status,
		GasUsed:		gasUsed,
		FeeCharged:		feeCharged,
		ReturnPayloadHash:	append([]byte(nil), returnPayloadHash...),
		ExecutedHeight:		executedHeight,
	}
	if errorCode != nil {
		receipt.ErrorCode = *errorCode
		receipt.HasErrorCode = true
	}
	return receipt
}

func ExpiredReceiptFromMessage(msg Message, executedHeight uint64) MessageReceipt {
	code := uint32(1)
	return ReceiptFromMessage(msg, MessageStatusExpired, 0, sdkmath.ZeroInt(), nil, &code, executedHeight)
}

func (r MessageReceipt) ValidateFormat() error {
	if len(r.MessageID) != MessageIDBytes {
		return fmt.Errorf("receipt message id must be %d bytes", MessageIDBytes)
	}
	if err := zonestypes.ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.DestinationZone); err != nil {
		return err
	}
	if !IsMessageStatus(r.Status) {
		return fmt.Errorf("unknown message status %q", r.Status)
	}
	if r.FeeCharged.IsNil() {
		r.FeeCharged = sdkmath.ZeroInt()
	}
	if r.FeeCharged.IsNegative() {
		return errors.New("receipt fee charged must be non-negative")
	}
	if len(r.ReturnPayloadHash) != 0 && len(r.ReturnPayloadHash) != sha256.Size {
		return fmt.Errorf("receipt return payload hash must be %d bytes", sha256.Size)
	}
	if r.ExecutedHeight == 0 {
		return errors.New("receipt executed height must be positive")
	}
	if len(r.ReceiptHash) != 0 && len(r.ReceiptHash) != MessageIDBytes {
		return fmt.Errorf("receipt hash must be %d bytes", MessageIDBytes)
	}
	return nil
}

func (r MessageReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if len(r.ReceiptHash) != MessageIDBytes {
		return fmt.Errorf("receipt hash must be %d bytes", MessageIDBytes)
	}
	if !bytes.Equal(r.ReceiptHash, ComputeReceiptHash(r)) {
		return errors.New("receipt hash mismatch")
	}
	return nil
}

func (r MessageReceipt) Clone() MessageReceipt {
	r.MessageID = append([]byte(nil), r.MessageID...)
	r.ReturnPayloadHash = append([]byte(nil), r.ReturnPayloadHash...)
	r.ReceiptHash = append([]byte(nil), r.ReceiptHash...)
	return r
}

func NewReplayTombstone(msg Message, consumedHeight uint64, proofHorizon uint64) (ReplayTombstone, error) {
	if consumedHeight == 0 || proofHorizon == 0 {
		return ReplayTombstone{}, errors.New("tombstone height and proof horizon must be positive")
	}
	tombstone := ReplayTombstone{
		MessageID:	append([]byte(nil), msg.MessageID...),
		SourceZone:	msg.SourceZone,
		Sender:		append(sdk.AccAddress(nil), msg.Sender...),
		Nonce:		msg.Nonce,
		SourceSequence:	msg.SourceSequence,
		ConsumedHeight:	consumedHeight,
		RetainUntil:	consumedHeight + proofHorizon,
	}
	tombstone.TombstoneHash = ComputeTombstoneHash(tombstone)
	return tombstone, tombstone.Validate()
}

func (t ReplayTombstone) Validate() error {
	if len(t.MessageID) != MessageIDBytes {
		return fmt.Errorf("tombstone message id must be %d bytes", MessageIDBytes)
	}
	if err := zonestypes.ValidateZoneID(t.SourceZone); err != nil {
		return err
	}
	if len(t.Sender) == 0 || t.Nonce == 0 || t.SourceSequence == 0 || t.ConsumedHeight == 0 || t.RetainUntil < t.ConsumedHeight {
		return errors.New("invalid replay tombstone")
	}
	if len(t.TombstoneHash) != MessageIDBytes {
		return fmt.Errorf("tombstone hash must be %d bytes", MessageIDBytes)
	}
	if !bytes.Equal(t.TombstoneHash, ComputeTombstoneHash(t)) {
		return errors.New("tombstone hash mismatch")
	}
	return nil
}

func (s KeeperState) Normalize() KeeperState {
	s.Outbox = normalizeQueueItems(s.Outbox)
	s.Inbox = normalizeQueueItems(s.Inbox)
	s.Receipts = normalizeReceipts(s.Receipts)
	s.Nonces = normalizeNonces(s.Nonces)
	s.Tombstones = normalizeTombstones(s.Tombstones)
	s.Expiry = normalizeExpiry(s.Expiry)
	return s
}

func ComputeMessageIDRoot(messages []Message) string {
	ordered := cloneMessages(messages)
	parts := []string{"aetra-messages-id-root-v1", fmt.Sprint(len(ordered))}
	for _, msg := range ordered {
		parts = append(parts, hex.EncodeToString(msg.MessageID), fmt.Sprint(msg.SourceSequence), fmt.Sprint(msg.Nonce))
	}
	return hashParts(parts...)
}

func ComputeMessageRoot(outbox []QueueItem, inbox []QueueItem) string {
	return hashParts(
		"aetra-messages-root-v1",
		ComputeQueueRoot("outbox", outbox),
		ComputeQueueRoot("inbox", inbox),
	)
}

func ComputeReceiptRoot(receipts []MessageReceipt) string {
	ordered := normalizeReceipts(receipts)
	parts := []string{"aetra-message-receipts-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, hex.EncodeToString(receipt.ReceiptHash))
	}
	return hashParts(parts...)
}

func ComputeKeeperRoots(state KeeperState) MessageRoots {
	state = state.Normalize()
	roots := MessageRoots{
		OutboxRoot:	ComputeQueueRoot("outbox", state.Outbox),
		InboxRoot:	ComputeQueueRoot("inbox", state.Inbox),
		ReceiptRoot:	ComputeReceiptRoot(state.Receipts),
		NonceRoot:	ComputeNonceRoot(state.Nonces),
		TombstoneRoot:	ComputeTombstoneRoot(state.Tombstones),
		ExpiryRoot:	ComputeExpiryRoot(state.Expiry),
		ParamsHash:	state.Params.ParamsHash,
	}
	roots.MessageRoot = hashParts("aetra-message-queue-pair-root-v1", roots.OutboxRoot, roots.InboxRoot)
	roots.StateRoot = hashParts("aetra-message-keeper-state-root-v1", roots.MessageRoot, roots.ReceiptRoot, roots.NonceRoot, roots.TombstoneRoot, roots.ExpiryRoot, roots.ParamsHash)
	return roots
}

func ComputeKeeperStateRoot(state KeeperState) string {
	return ComputeKeeperRoots(state).StateRoot
}

func ComputeQueueRoot(kind string, items []QueueItem) string {
	ordered := normalizeQueueItems(items)
	parts := []string{"aetra-message-queue-root-v1", kind, fmt.Sprint(len(ordered))}
	for _, item := range ordered {
		parts = append(parts, QueueItemHash(item))
	}
	return hashParts(parts...)
}

func QueueItemHash(item QueueItem) string {
	return hashParts("aetra-message-queue-item-v1", hex.EncodeToString(item.Message.MessageID), fmt.Sprint(item.EnqueuedHeight), fmt.Sprint(item.Message.SourceSequence), fmt.Sprint(item.Message.Nonce))
}

func ComputeNonceRoot(nonces []SenderNonce) string {
	ordered := normalizeNonces(nonces)
	parts := []string{"aetra-message-nonce-root-v1", fmt.Sprint(len(ordered))}
	for _, nonce := range ordered {
		parts = append(parts, string(nonce.SourceZone), hex.EncodeToString(nonce.Sender), fmt.Sprint(nonce.Nonce))
	}
	return hashParts(parts...)
}

func ComputeTombstoneRoot(tombstones []ReplayTombstone) string {
	ordered := normalizeTombstones(tombstones)
	parts := []string{"aetra-message-tombstone-root-v1", fmt.Sprint(len(ordered))}
	for _, tombstone := range ordered {
		parts = append(parts, hex.EncodeToString(tombstone.TombstoneHash))
	}
	return hashParts(parts...)
}

func ComputeExpiryRoot(expiry []ExpiryItem) string {
	ordered := normalizeExpiry(expiry)
	parts := []string{"aetra-message-expiry-root-v1", fmt.Sprint(len(ordered))}
	for _, item := range ordered {
		parts = append(parts, fmt.Sprint(item.Deadline), hex.EncodeToString(item.MessageID))
	}
	return hashParts(parts...)
}

func ComputeReceiptHash(receipt MessageReceipt) []byte {
	buf := bytes.NewBuffer(nil)
	writeString(buf.Write, "aetra-message-receipt-v1")
	writeBytes(buf.Write, receipt.MessageID)
	writeString(buf.Write, string(receipt.SourceZone))
	writeString(buf.Write, string(receipt.DestinationZone))
	writeString(buf.Write, string(receipt.Status))
	writeU64(buf.Write, receipt.GasUsed)
	writeString(buf.Write, intString(receipt.FeeCharged))
	writeBytes(buf.Write, receipt.ReturnPayloadHash)
	writeBool(buf.Write, receipt.HasErrorCode)
	writeU64(buf.Write, uint64(receipt.ErrorCode))
	writeU64(buf.Write, receipt.ExecutedHeight)
	sum := sha256.Sum256(buf.Bytes())
	return append([]byte(nil), sum[:]...)
}

func ComputeTombstoneHash(tombstone ReplayTombstone) []byte {
	buf := bytes.NewBuffer(nil)
	writeString(buf.Write, "aetra-message-tombstone-v1")
	writeBytes(buf.Write, tombstone.MessageID)
	writeString(buf.Write, string(tombstone.SourceZone))
	writeBytes(buf.Write, tombstone.Sender)
	writeU64(buf.Write, tombstone.Nonce)
	writeU64(buf.Write, tombstone.SourceSequence)
	writeU64(buf.Write, tombstone.ConsumedHeight)
	writeU64(buf.Write, tombstone.RetainUntil)
	sum := sha256.Sum256(buf.Bytes())
	return append([]byte(nil), sum[:]...)
}

func ComputeProofHash(proof QueryProofResponse) string {
	parts := []string{"aetra-message-proof-v1", string(proof.Kind), hex.EncodeToString(proof.MessageID), proof.Root, proof.ValueHash}
	parts = append(parts, proof.Path...)
	return hashParts(parts...)
}

func IsMessageStatus(status MessageStatus) bool {
	switch status {
	case MessageStatusQueued, MessageStatusExecuted, MessageStatusFailed, MessageStatusExpired, MessageStatusBounced, MessageStatusRejected:
		return true
	default:
		return false
	}
}

func OutboxKey(sourceZone zonestypes.ZoneID, sender sdk.AccAddress, sequence uint64) (string, error) {
	if err := zonestypes.ValidateZoneID(sourceZone); err != nil {
		return "", err
	}
	if len(sender) == 0 || sequence == 0 {
		return "", errors.New("outbox sender and sequence are required")
	}
	return MessagesOutboxPrefix + "/" + string(sourceZone) + "/" + hex.EncodeToString(sender) + "/" + fmt.Sprint(sequence), nil
}

func InboxKey(destinationZone zonestypes.ZoneID, sender sdk.AccAddress, sequence uint64) (string, error) {
	if err := zonestypes.ValidateZoneID(destinationZone); err != nil {
		return "", err
	}
	if len(sender) == 0 || sequence == 0 {
		return "", errors.New("inbox sender and sequence are required")
	}
	return MessagesInboxPrefix + "/" + string(destinationZone) + "/" + hex.EncodeToString(sender) + "/" + fmt.Sprint(sequence), nil
}

func ReceiptKey(messageID []byte) (string, error) {
	if len(messageID) != MessageIDBytes {
		return "", fmt.Errorf("receipt message id must be %d bytes", MessageIDBytes)
	}
	return MessagesReceiptPrefix + "/" + hex.EncodeToString(messageID), nil
}

func NonceKey(sourceZone zonestypes.ZoneID, sender sdk.AccAddress) (string, error) {
	if err := zonestypes.ValidateZoneID(sourceZone); err != nil {
		return "", err
	}
	if len(sender) == 0 {
		return "", errors.New("nonce sender is required")
	}
	return MessagesNoncePrefix + "/" + string(sourceZone) + "/" + hex.EncodeToString(sender), nil
}

func ReplayKey(messageID []byte) (string, error) {
	if len(messageID) != MessageIDBytes {
		return "", fmt.Errorf("replay message id must be %d bytes", MessageIDBytes)
	}
	return MessagesReplayPrefix + "/" + hex.EncodeToString(messageID), nil
}

func ExpiryKey(deadline uint64, messageID []byte) (string, error) {
	if deadline == 0 || len(messageID) != MessageIDBytes {
		return "", errors.New("expiry deadline and message id are required")
	}
	return MessagesExpiryPrefix + "/" + fmt.Sprint(deadline) + "/" + hex.EncodeToString(messageID), nil
}

func (r QueryProofRequest) Validate() error {
	if r.Kind != MessageProofInclusion && r.Kind != MessageProofReceipt {
		return fmt.Errorf("unknown message proof kind %q", r.Kind)
	}
	if len(r.MessageID) != MessageIDBytes {
		return fmt.Errorf("proof message id must be %d bytes", MessageIDBytes)
	}
	if err := zonestypes.ValidateHash("proof root", r.Root); err != nil {
		return err
	}
	if r.Limit > MaxProofPathItems {
		return fmt.Errorf("proof limit must be <= %d", MaxProofPathItems)
	}
	return nil
}

func (p QueryProofResponse) ValidateFor(req QueryProofRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if p.Kind != req.Kind || !bytes.Equal(p.MessageID, req.MessageID) || p.Root != req.Root {
		return errors.New("proof request mismatch")
	}
	if err := zonestypes.ValidateHash("proof value hash", p.ValueHash); err != nil {
		return err
	}
	if len(p.Path) > MaxProofPathItems || req.Limit != 0 && len(p.Path) > int(req.Limit) {
		return errors.New("proof path exceeds limit")
	}
	for _, item := range p.Path {
		if err := zonestypes.ValidateHash("proof path item", item); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("proof hash", p.ProofHash); err != nil {
		return err
	}
	if p.ProofHash != ComputeProofHash(p) {
		return errors.New("proof hash mismatch")
	}
	return nil
}

func appendReceiptAndTombstone(state KeeperState, receipt MessageReceipt) (KeeperState, MessageReceipt, error) {
	if hasReceipt(state.Receipts, receipt.MessageID) {
		return KeeperState{}, MessageReceipt{}, errors.New("receipt already exists")
	}
	msg, found := findQueueMessage(state.Inbox, receipt.MessageID)
	if !found {
		msg, found = findQueueMessage(state.Outbox, receipt.MessageID)
	}
	if !found {
		return KeeperState{}, MessageReceipt{}, errors.New("message not found for receipt")
	}
	receipt, err := NewMessageReceipt(receipt)
	if err != nil {
		return KeeperState{}, MessageReceipt{}, err
	}
	tombstone, err := NewReplayTombstone(msg.Message, receipt.ExecutedHeight, state.Params.ProofHorizon)
	if err != nil {
		return KeeperState{}, MessageReceipt{}, err
	}
	state.Receipts = append(state.Receipts, receipt)
	state.Tombstones = append(state.Tombstones, tombstone)
	return state.Normalize(), receipt, nil
}

func normalizeQueueItems(items []QueueItem) []QueueItem {
	out := make([]QueueItem, len(items))
	for i, item := range items {
		out[i] = QueueItem{Message: item.Message.Clone(), EnqueuedHeight: item.EnqueuedHeight}
	}
	sort.SliceStable(out, func(i, j int) bool { return compareMessages(out[i].Message, out[j].Message) < 0 })
	return out
}

func normalizeReceipts(receipts []MessageReceipt) []MessageReceipt {
	out := make([]MessageReceipt, len(receipts))
	for i, receipt := range receipts {
		out[i] = receipt.Clone()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ExecutedHeight != out[j].ExecutedHeight {
			return out[i].ExecutedHeight < out[j].ExecutedHeight
		}
		return hex.EncodeToString(out[i].MessageID) < hex.EncodeToString(out[j].MessageID)
	})
	return out
}

func normalizeNonces(nonces []SenderNonce) []SenderNonce {
	out := append([]SenderNonce(nil), nonces...)
	sort.SliceStable(out, func(i, j int) bool {
		left := string(out[i].SourceZone) + "/" + hex.EncodeToString(out[i].Sender)
		right := string(out[j].SourceZone) + "/" + hex.EncodeToString(out[j].Sender)
		return left < right
	})
	return out
}

func normalizeTombstones(tombstones []ReplayTombstone) []ReplayTombstone {
	out := append([]ReplayTombstone(nil), tombstones...)
	sort.SliceStable(out, func(i, j int) bool {
		return hex.EncodeToString(out[i].MessageID) < hex.EncodeToString(out[j].MessageID)
	})
	return out
}

func normalizeExpiry(expiry []ExpiryItem) []ExpiryItem {
	out := append([]ExpiryItem(nil), expiry...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Deadline != out[j].Deadline {
			return out[i].Deadline < out[j].Deadline
		}
		return hex.EncodeToString(out[i].MessageID) < hex.EncodeToString(out[j].MessageID)
	})
	return out
}

func cloneMessages(messages []Message) []Message {
	out := make([]Message, len(messages))
	for i, msg := range messages {
		out[i] = msg.Clone()
	}
	sort.SliceStable(out, func(i, j int) bool { return compareMessages(out[i], out[j]) < 0 })
	return out
}

func compareMessages(left, right Message) int {
	for _, pair := range [][2]string{
		{string(left.SourceZone), string(right.SourceZone)},
		{hex.EncodeToString(left.Sender), hex.EncodeToString(right.Sender)},
		{string(left.DestinationZone), string(right.DestinationZone)},
		{left.RouteID, right.RouteID},
	} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	if left.SourceSequence != right.SourceSequence {
		if left.SourceSequence < right.SourceSequence {
			return -1
		}
		return 1
	}
	if left.Nonce != right.Nonce {
		if left.Nonce < right.Nonce {
			return -1
		}
		return 1
	}
	return strings.Compare(hex.EncodeToString(left.MessageID), hex.EncodeToString(right.MessageID))
}

func upsertNonce(nonces []SenderNonce, update SenderNonce) []SenderNonce {
	out := append([]SenderNonce(nil), nonces...)
	for i := range out {
		if out[i].SourceZone == update.SourceZone && bytes.Equal(out[i].Sender, update.Sender) {
			out[i] = update
			return normalizeNonces(out)
		}
	}
	out = append(out, update)
	return normalizeNonces(out)
}

func nonceFor(nonces []SenderNonce, sourceZone zonestypes.ZoneID, sender sdk.AccAddress) uint64 {
	for _, nonce := range nonces {
		if nonce.SourceZone == sourceZone && bytes.Equal(nonce.Sender, sender) {
			return nonce.Nonce
		}
	}
	return 0
}

func findQueueMessage(items []QueueItem, messageID []byte) (QueueItem, bool) {
	for _, item := range items {
		if bytes.Equal(item.Message.MessageID, messageID) {
			return QueueItem{Message: item.Message.Clone(), EnqueuedHeight: item.EnqueuedHeight}, true
		}
	}
	return QueueItem{}, false
}

func hasQueueMessage(items []QueueItem, messageID []byte) bool {
	_, found := findQueueMessage(items, messageID)
	return found
}

func removeQueueMessage(items []QueueItem, messageID []byte) []QueueItem {
	out := make([]QueueItem, 0, len(items))
	for _, item := range items {
		if !bytes.Equal(item.Message.MessageID, messageID) {
			out = append(out, item)
		}
	}
	return normalizeQueueItems(out)
}

func hasReceipt(receipts []MessageReceipt, messageID []byte) bool {
	for _, receipt := range receipts {
		if bytes.Equal(receipt.MessageID, messageID) {
			return true
		}
	}
	return false
}

func hasTombstone(tombstones []ReplayTombstone, messageID []byte) bool {
	for _, tombstone := range tombstones {
		if bytes.Equal(tombstone.MessageID, messageID) {
			return true
		}
	}
	return false
}

func removeExpiry(expiry []ExpiryItem, messageID []byte) []ExpiryItem {
	out := make([]ExpiryItem, 0, len(expiry))
	for _, item := range expiry {
		if !bytes.Equal(item.MessageID, messageID) {
			out = append(out, item)
		}
	}
	return normalizeExpiry(out)
}

func EmptyHash() string {
	return hashParts("empty")
}

func hashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		writeString(h.Write, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func receiptErrorCodeString(receipt MessageReceipt) string {
	if !receipt.HasErrorCode {
		return ""
	}
	return fmt.Sprint(receipt.ErrorCode)
}

func intString(value sdkmath.Int) string {
	if value.IsNil() {
		return sdkmath.ZeroInt().String()
	}
	return value.String()
}

func validateToken(fieldName, value string, maxLen int) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxLen)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func writeString(write func([]byte) (int, error), value string) {
	writeBytes(write, []byte(value))
}

func writeBytes(write func([]byte) (int, error), bz []byte) {
	writeU32(write, uint32(len(bz)))
	_, _ = write(bz)
}

func writeBool(write func([]byte) (int, error), value bool) {
	if value {
		_, _ = write([]byte{1})
		return
	}
	_, _ = write([]byte{0})
}

func writeU32(write func([]byte) (int, error), value uint32) {
	var out [4]byte
	binary.BigEndian.PutUint32(out[:], value)
	_, _ = write(out[:])
}

func writeU64(write func([]byte) (int, error), value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	_, _ = write(out[:])
}
