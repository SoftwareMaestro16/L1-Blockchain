package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	MaxAsyncMessageEndpointLength	= 128
	MaxAsyncMessagePayloadBytes	= 128 * 1024
	MaxAsyncMessagePayloadType	= 96
	MaxAsyncMessageRouteHint	= 128
	MaxAsyncMessagePriority		= 255

	AVMRetryModeNone		AVMRetryMode	= "none"
	AVMRetryModeFixed		AVMRetryMode	= "fixed"
	AVMRetryModeBoundedBackoff	AVMRetryMode	= "bounded_backoff"

	AVMBackoffModeNone		AVMBackoffMode	= "none"
	AVMBackoffModeLinear		AVMBackoffMode	= "linear"
	AVMBackoffModeExponential	AVMBackoffMode	= "exponential"
)

type AVMRetryMode string
type AVMBackoffMode string

type AVMRetryPolicy struct {
	Mode		AVMRetryMode
	MaxAttempts	uint32
	RetryDelay	uint64
	BackoffMode	AVMBackoffMode
	MaxRetryHeight	uint64
	ChargeRetryGas	bool
}

type AVMAsyncMessage struct {
	ID				string
	ChainID				string
	Source				string
	Destination			string
	Payload				[]byte
	GasLimit			uint64
	DelayHeight			uint64
	ExpiryHeight			uint64
	RetryPolicy			AVMRetryPolicy
	BounceFlag			bool
	SourceZone			zonestypes.ZoneID
	DestinationZone			zonestypes.ZoneID
	SourceActorOptional		string
	DestinationActorOptional	string
	SenderNonce			uint64
	PayloadType			string
	PayloadHash			string
	ValueNAET			uint64
	ForwardingFee			uint64
	Priority			uint8
	CreatedHeight			uint64
	RouteHintOptional		string
	AuthProofOptional		string
	StateProofOptional		string
}

type AVMAsyncReplayTombstone struct {
	MessageID	string
	ConsumedHeight	uint64
}

type AVMAsyncMessageRegistry struct {
	Messages		[]AVMAsyncMessage
	ConsumedMessageIDs	[]string
	ReplayTombstones	[]AVMAsyncReplayTombstone
}

func DefaultAVMRetryPolicy(expiryHeight uint64) AVMRetryPolicy {
	return AVMRetryPolicy{
		Mode:		AVMRetryModeFixed,
		MaxAttempts:	3,
		RetryDelay:	1,
		BackoffMode:	AVMBackoffModeNone,
		MaxRetryHeight:	expiryHeight,
		ChargeRetryGas:	true,
	}
}

func NewAVMAsyncMessage(msg AVMAsyncMessage) (AVMAsyncMessage, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if msg.PayloadHash == "" && len(msg.Payload) > 0 {
		msg.PayloadHash = ComputeAVMAsyncPayloadHash(msg.Payload)
	}
	msg.ID = DeriveAVMAsyncMessageID(msg)
	return msg, msg.Validate()
}

func (m AVMAsyncMessage) Validate() error {
	m = canonicalAVMAsyncMessage(m)
	if err := validateRouterOptionalToken("async message chain id", m.ChainID, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if m.ChainID == "" {
		return errors.New("async message chain id is required")
	}
	if err := validateRouterOptionalToken("async message source", m.Source, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if m.Source == "" {
		return errors.New("async message source is required")
	}
	if err := validateRouterOptionalToken("async message destination", m.Destination, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if m.Destination == "" {
		return errors.New("async message destination is required")
	}
	if len(m.Payload) > MaxAsyncMessagePayloadBytes {
		return fmt.Errorf("async message payload must be <= %d bytes", MaxAsyncMessagePayloadBytes)
	}
	if m.PayloadHash == "" {
		return errors.New("async message payload hash is required")
	}
	if err := zonestypes.ValidateHash("async message id", m.ID); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("async message payload hash", m.PayloadHash); err != nil {
		return err
	}
	if len(m.Payload) > 0 && m.PayloadHash != ComputeAVMAsyncPayloadHash(m.Payload) {
		return errors.New("async message payload hash mismatch")
	}
	if m.ID != DeriveAVMAsyncMessageID(m) {
		return errors.New("async message id mismatch")
	}
	if m.GasLimit == 0 {
		return errors.New("async message gas limit must be positive")
	}
	if m.CreatedHeight == 0 {
		return errors.New("async message created height must be positive")
	}
	if m.ExpiryHeight == 0 {
		return errors.New("async message expiry height must be explicit")
	}
	if m.ExpiryHeight <= m.CreatedHeight {
		return errors.New("async message expiry height must be after creation")
	}
	if m.DelayHeight > 0 {
		if m.DelayHeight > ^uint64(0)-m.CreatedHeight {
			return errors.New("async message delay height overflows")
		}
		if m.CreatedHeight+m.DelayHeight > m.ExpiryHeight {
			return errors.New("async message delay height must not exceed expiry")
		}
	}
	if err := m.RetryPolicy.ValidateForMessage(m.ExpiryHeight); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.DestinationZone); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("async message source actor", m.SourceActorOptional, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("async message destination actor", m.DestinationActorOptional, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("async message payload type", m.PayloadType, MaxAsyncMessagePayloadType); err != nil {
		return err
	}
	if m.PayloadType == "" {
		return errors.New("async message payload type is required")
	}
	if m.ForwardingFee == 0 {
		return errors.New("async message forwarding fee must be positive")
	}
	if m.Priority > MaxAsyncMessagePriority {
		return errors.New("async message priority exceeds maximum")
	}
	if err := validateRouterOptionalToken("async message route hint", m.RouteHintOptional, MaxAsyncMessageRouteHint); err != nil {
		return err
	}
	if m.AuthProofOptional != "" {
		if err := zonestypes.ValidateHash("async message auth proof", m.AuthProofOptional); err != nil {
			return err
		}
	}
	if m.StateProofOptional != "" {
		if err := zonestypes.ValidateHash("async message state proof", m.StateProofOptional); err != nil {
			return err
		}
	}
	return nil
}

func (p AVMRetryPolicy) ValidateForMessage(expiryHeight uint64) error {
	if !IsAVMRetryMode(p.Mode) {
		return fmt.Errorf("invalid async retry mode %q", p.Mode)
	}
	if !IsAVMBackoffMode(p.BackoffMode) {
		return fmt.Errorf("invalid async retry backoff mode %q", p.BackoffMode)
	}
	if p.Mode == AVMRetryModeNone {
		if p.MaxAttempts != 0 || p.RetryDelay != 0 || p.MaxRetryHeight != 0 || p.ChargeRetryGas {
			return errors.New("async retry mode none must not configure retry attempts, delay, height, or gas charge")
		}
		if p.BackoffMode != "" && p.BackoffMode != AVMBackoffModeNone {
			return errors.New("async retry mode none must use no backoff")
		}
		return nil
	}
	if p.MaxAttempts == 0 {
		return errors.New("async retry count must be bounded")
	}
	if p.RetryDelay == 0 {
		return errors.New("async retry delay must be deterministic and positive")
	}
	if p.MaxRetryHeight == 0 {
		return errors.New("async max retry height must be explicit")
	}
	if expiryHeight > 0 && p.MaxRetryHeight > expiryHeight {
		return errors.New("async retry cannot exceed message expiry")
	}
	if !p.ChargeRetryGas {
		return errors.New("async retry gas must be reserved or explicitly charged")
	}
	if p.Mode == AVMRetryModeFixed && p.BackoffMode != AVMBackoffModeNone {
		return errors.New("fixed async retry mode must use no backoff")
	}
	if p.Mode == AVMRetryModeBoundedBackoff && p.BackoffMode == AVMBackoffModeNone {
		return errors.New("bounded backoff retry mode requires deterministic backoff")
	}
	return nil
}

func IsAVMRetryMode(mode AVMRetryMode) bool {
	switch mode {
	case AVMRetryModeNone, AVMRetryModeFixed, AVMRetryModeBoundedBackoff:
		return true
	default:
		return false
	}
}

func IsAVMBackoffMode(mode AVMBackoffMode) bool {
	switch mode {
	case "", AVMBackoffModeNone, AVMBackoffModeLinear, AVMBackoffModeExponential:
		return true
	default:
		return false
	}
}

func (t AVMAsyncReplayTombstone) Validate() error {
	if err := zonestypes.ValidateHash("async replay tombstone message id", t.MessageID); err != nil {
		return err
	}
	if t.ConsumedHeight == 0 {
		return errors.New("async replay tombstone consumed height must be positive")
	}
	return nil
}

func (r AVMAsyncMessageRegistry) Validate() error {
	r = canonicalAVMAsyncMessageRegistry(r)
	messageIDs := make(map[string]AVMAsyncMessage, len(r.Messages))
	nonceScopes := make(map[string]struct{}, len(r.Messages))
	for i, msg := range r.Messages {
		if err := msg.Validate(); err != nil {
			return err
		}
		if _, found := messageIDs[msg.ID]; found {
			return fmt.Errorf("duplicate async message id %q", msg.ID)
		}
		messageIDs[msg.ID] = msg
		scope := AsyncMessageNonceScope(msg.SourceZone, msg.Source, msg.SenderNonce)
		if _, found := nonceScopes[scope]; found {
			return fmt.Errorf("duplicate async sender nonce scope %q", scope)
		}
		nonceScopes[scope] = struct{}{}
		if i > 0 && r.Messages[i-1].ID >= msg.ID {
			return errors.New("async messages must be sorted canonically")
		}
	}
	tombstones := make(map[string]AVMAsyncReplayTombstone, len(r.ReplayTombstones))
	for i, tombstone := range r.ReplayTombstones {
		if err := tombstone.Validate(); err != nil {
			return err
		}
		if _, found := tombstones[tombstone.MessageID]; found {
			return fmt.Errorf("duplicate async replay tombstone %q", tombstone.MessageID)
		}
		tombstones[tombstone.MessageID] = tombstone
		if _, replay := messageIDs[tombstone.MessageID]; replay {
			return fmt.Errorf("async message id %q is replay tombstoned", tombstone.MessageID)
		}
		if i > 0 && r.ReplayTombstones[i-1].MessageID >= tombstone.MessageID {
			return errors.New("async replay tombstones must be sorted canonically")
		}
	}
	seenConsumed := make(map[string]struct{}, len(r.ConsumedMessageIDs))
	for i, messageID := range r.ConsumedMessageIDs {
		if err := zonestypes.ValidateHash("async consumed message id", messageID); err != nil {
			return err
		}
		if _, found := seenConsumed[messageID]; found {
			return fmt.Errorf("duplicate async consumed message id %q", messageID)
		}
		seenConsumed[messageID] = struct{}{}
		if _, found := tombstones[messageID]; !found {
			return fmt.Errorf("consumed async message %q must create replay tombstone", messageID)
		}
		if i > 0 && r.ConsumedMessageIDs[i-1] >= messageID {
			return errors.New("async consumed message ids must be sorted canonically")
		}
	}
	return nil
}

func DeriveAVMAsyncMessageID(msg AVMAsyncMessage) string {
	msg = canonicalAVMAsyncMessage(msg)
	h := sha256.New()
	writeEnginePart(h, "aetra-async-message-id-v1")
	writeEnginePart(h, msg.ChainID)
	writeEnginePart(h, string(msg.SourceZone))
	writeEnginePart(h, msg.Source)
	writeEngineUint64(h, msg.SenderNonce)
	writeEnginePart(h, string(msg.DestinationZone))
	writeEnginePart(h, msg.Destination)
	writeEnginePart(h, msg.PayloadHash)
	writeEngineUint64(h, msg.CreatedHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func NextAVMRetryHeight(createdHeight uint64, attempt uint32, policy AVMRetryPolicy) (uint64, error) {
	if attempt == 0 {
		return 0, errors.New("async retry attempt must be positive")
	}
	if err := policy.ValidateForMessage(policy.MaxRetryHeight); err != nil {
		return 0, err
	}
	if attempt > policy.MaxAttempts {
		return 0, errors.New("async retry attempt exceeds max attempts")
	}
	delay := policy.RetryDelay
	switch policy.BackoffMode {
	case AVMBackoffModeLinear:
		delay *= uint64(attempt)
	case AVMBackoffModeExponential:
		if attempt > 63 {
			return 0, errors.New("async retry exponential backoff attempt overflows")
		}
		delay *= uint64(1) << (attempt - 1)
	}
	if delay > ^uint64(0)-createdHeight {
		return 0, errors.New("async retry height overflows")
	}
	next := createdHeight + delay
	if next > policy.MaxRetryHeight {
		return 0, errors.New("async retry cannot exceed max retry height")
	}
	return next, nil
}

func ComputeAVMAsyncPayloadHash(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func AsyncMessageNonceScope(sourceZone zonestypes.ZoneID, source string, senderNonce uint64) string {
	return fmt.Sprintf("%s/%s/%020d", sourceZone, strings.TrimSpace(source), senderNonce)
}

func canonicalAVMAsyncMessage(msg AVMAsyncMessage) AVMAsyncMessage {
	msg.ID = strings.TrimSpace(msg.ID)
	msg.ChainID = strings.TrimSpace(msg.ChainID)
	msg.Source = strings.TrimSpace(msg.Source)
	msg.Destination = strings.TrimSpace(msg.Destination)
	msg.Payload = append([]byte(nil), msg.Payload...)
	msg.SourceActorOptional = strings.TrimSpace(msg.SourceActorOptional)
	msg.DestinationActorOptional = strings.TrimSpace(msg.DestinationActorOptional)
	msg.PayloadType = strings.TrimSpace(msg.PayloadType)
	msg.PayloadHash = strings.TrimSpace(msg.PayloadHash)
	msg.RouteHintOptional = strings.TrimSpace(msg.RouteHintOptional)
	msg.AuthProofOptional = strings.TrimSpace(msg.AuthProofOptional)
	msg.StateProofOptional = strings.TrimSpace(msg.StateProofOptional)
	return msg
}

func canonicalAVMAsyncMessageRegistry(registry AVMAsyncMessageRegistry) AVMAsyncMessageRegistry {
	registry.Messages = append([]AVMAsyncMessage(nil), registry.Messages...)
	for i := range registry.Messages {
		registry.Messages[i] = canonicalAVMAsyncMessage(registry.Messages[i])
	}
	registry.ConsumedMessageIDs = append([]string(nil), registry.ConsumedMessageIDs...)
	for i, messageID := range registry.ConsumedMessageIDs {
		registry.ConsumedMessageIDs[i] = strings.TrimSpace(messageID)
	}
	registry.ReplayTombstones = append([]AVMAsyncReplayTombstone(nil), registry.ReplayTombstones...)
	for i := range registry.ReplayTombstones {
		registry.ReplayTombstones[i].MessageID = strings.TrimSpace(registry.ReplayTombstones[i].MessageID)
	}
	sort.SliceStable(registry.Messages, func(i, j int) bool {
		return registry.Messages[i].ID < registry.Messages[j].ID
	})
	sort.Strings(registry.ConsumedMessageIDs)
	sort.SliceStable(registry.ReplayTombstones, func(i, j int) bool {
		return registry.ReplayTombstones[i].MessageID < registry.ReplayTombstones[j].MessageID
	})
	return registry
}
