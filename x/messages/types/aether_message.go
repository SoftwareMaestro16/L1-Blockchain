package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	MaxAetherAddressLength		= 128
	MaxAetherPayloadType		= 128
	MaxAetherProofTypeLength	= 96
	MaxAetherSignatureLength	= 512
)

type ExecutionMode = UnifiedExecutionMode
type OrderingClass = UnifiedOrderingClass

const (
	ExecutionModeSyncLocal	ExecutionMode	= UnifiedExecutionSyncLocal
	ExecutionModeAsync	ExecutionMode	= UnifiedExecutionAsync
	ExecutionModeDeferred	ExecutionMode	= UnifiedExecutionDeferred
	ExecutionModePromise	ExecutionMode	= UnifiedExecutionPromise

	OrderingClassUnordered		OrderingClass	= UnifiedOrderingUnordered
	OrderingClassSenderOrdered	OrderingClass	= UnifiedOrderingSenderOrdered
	OrderingClassReceiverOrdered	OrderingClass	= UnifiedOrderingReceiverOrdered
	OrderingClassObjectOrdered	OrderingClass	= UnifiedOrderingObjectOrdered
	OrderingClassStrictTraceOrder	OrderingClass	= UnifiedOrderingStrictTraceOrder
)

type AetherProof struct {
	ProofType	string
	RootHash	string
	ProofHash	string
}

type AetherSignature struct {
	Signer		string
	SignatureHex	string
}

type AetherMessage struct {
	MsgID		string
	ParentMsgID	string
	TraceID		string
	Sender		string
	SenderZoneID	zonestypes.ZoneID
	SenderShardID	string
	Receiver	string
	ReceiverZoneID	zonestypes.ZoneID
	ReceiverShardID	string
	ValueNAET	sdkmath.Int
	Payload		[]byte
	PayloadType	string
	GasLimit	uint64
	GasPrice	sdkmath.Int
	ForwardingFee	sdkmath.Int
	ExpiryHeight	uint64
	Bounce		bool
	ExecutionMode	ExecutionMode
	OrderingClass	OrderingClass
	RouteCommitment	string
	AuthProof	AetherProof
	StateProof	AetherProof
	CreatedAtHeight	uint64
	Nonce		uint64
	Signature	AetherSignature
}

func NewAetherMessage(msg AetherMessage) (AetherMessage, error) {
	if msg.MsgID != "" {
		return AetherMessage{}, errors.New("aether message id must be empty before construction")
	}
	msg = normalizeAetherMessage(msg)
	if msg.TraceID == "" {
		msg.TraceID = ComputeAetherTraceID(msg)
	}
	if err := msg.ValidateForID(); err != nil {
		return AetherMessage{}, err
	}
	msg.MsgID = ComputeAetherMessageID(msg)
	return msg, msg.Validate()
}

func (m AetherMessage) Validate() error {
	m = normalizeAetherMessage(m)
	if err := m.ValidateForID(); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("aether message id", m.MsgID); err != nil {
		return err
	}
	if m.MsgID != ComputeAetherMessageID(m) {
		return errors.New("aether message id mismatch")
	}
	return nil
}

func (m AetherMessage) ValidateForID() error {
	m = normalizeAetherMessage(m)
	if m.ParentMsgID != "" {
		if err := zonestypes.ValidateHash("aether parent message id", m.ParentMsgID); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("aether trace id", m.TraceID); err != nil {
		return err
	}
	if err := validateToken("aether sender", m.Sender, MaxAetherAddressLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.SenderZoneID); err != nil {
		return err
	}
	if err := validateToken("aether sender shard id", m.SenderShardID, MaxShardIDLength); err != nil {
		return err
	}
	if err := validateToken("aether receiver", m.Receiver, MaxAetherAddressLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.ReceiverZoneID); err != nil {
		return err
	}
	if err := validateToken("aether receiver shard id", m.ReceiverShardID, MaxShardIDLength); err != nil {
		return err
	}
	if m.ValueNAET.IsNil() {
		return errors.New("aether message value is required")
	}
	if m.ValueNAET.IsNegative() {
		return errors.New("aether message value must be non-negative")
	}
	if len(m.Payload) == 0 {
		return errors.New("aether message payload is required")
	}
	if err := validateToken("aether payload type", m.PayloadType, MaxAetherPayloadType); err != nil {
		return err
	}
	if m.GasLimit == 0 {
		return errors.New("aether message gas limit must be positive")
	}
	if m.GasPrice.IsNil() || m.ForwardingFee.IsNil() {
		return errors.New("aether message fee metadata is required")
	}
	if m.GasPrice.IsNegative() || m.ForwardingFee.IsNegative() {
		return errors.New("aether message fees must be non-negative")
	}
	if m.CreatedAtHeight == 0 || m.ExpiryHeight == 0 {
		return errors.New("aether message heights must be positive")
	}
	if m.ExpiryHeight < m.CreatedAtHeight {
		return errors.New("aether message expiry must not precede creation")
	}
	if m.Nonce == 0 {
		return errors.New("aether message nonce must be positive")
	}
	if !IsUnifiedExecutionMode(m.ExecutionMode) {
		return fmt.Errorf("unknown aether execution mode %q", m.ExecutionMode)
	}
	if !IsUnifiedOrderingClass(m.OrderingClass) {
		return fmt.Errorf("unknown aether ordering class %q", m.OrderingClass)
	}
	if err := zonestypes.ValidateHash("aether route commitment", m.RouteCommitment); err != nil {
		return err
	}
	if err := m.AuthProof.ValidateOptional("aether auth proof"); err != nil {
		return err
	}
	if err := m.StateProof.ValidateOptional("aether state proof"); err != nil {
		return err
	}
	if err := m.Signature.ValidateOptional(); err != nil {
		return err
	}
	return nil
}

func (p AetherProof) ValidateOptional(fieldName string) error {
	if p == (AetherProof{}) {
		return nil
	}
	if err := validateToken(fieldName+" type", p.ProofType, MaxAetherProofTypeLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash(fieldName+" root hash", p.RootHash); err != nil {
		return err
	}
	return zonestypes.ValidateHash(fieldName+" proof hash", p.ProofHash)
}

func (s AetherSignature) ValidateOptional() error {
	if s == (AetherSignature{}) {
		return nil
	}
	if err := validateToken("aether signature signer", s.Signer, MaxAetherAddressLength); err != nil {
		return err
	}
	if strings.TrimSpace(s.SignatureHex) != s.SignatureHex || s.SignatureHex == "" {
		return errors.New("aether signature hex is required and must not have surrounding whitespace")
	}
	if len(s.SignatureHex) > MaxAetherSignatureLength {
		return fmt.Errorf("aether signature hex must be <= %d bytes", MaxAetherSignatureLength)
	}
	_, err := hex.DecodeString(s.SignatureHex)
	if err != nil {
		return errors.New("aether signature must be hex encoded")
	}
	return nil
}

func (m AetherMessage) Clone() AetherMessage {
	m.Payload = append([]byte(nil), m.Payload...)
	return m
}

func AetherMessageFromMessage(msg Message, route UnifiedMessageRoute, traceID string, parentMsgID string) (AetherMessage, error) {
	if err := msg.Validate(DefaultMessageParams(msg.ChainID)); err != nil {
		return AetherMessage{}, err
	}
	if traceID == "" {
		traceID = hashParts("aetra-aether-message-trace-v1", hex.EncodeToString(msg.MessageID))
	}
	return NewAetherMessage(AetherMessage{
		ParentMsgID:		parentMsgID,
		TraceID:		traceID,
		Sender:			hex.EncodeToString(msg.Sender),
		SenderZoneID:		msg.SourceZone,
		SenderShardID:		route.SourceShardID,
		Receiver:		hex.EncodeToString(msg.Recipient),
		ReceiverZoneID:		msg.DestinationZone,
		ReceiverShardID:	route.DestinationShardID,
		ValueNAET:		msg.Value,
		Payload:		msg.Payload,
		PayloadType:		msg.Opcode,
		GasLimit:		msg.GasLimit,
		GasPrice:		msg.FeeLimit,
		ForwardingFee:		msg.FeeLimit,
		ExpiryHeight:		msg.Deadline,
		Bounce:			msg.Bounce,
		ExecutionMode:		route.ExecutionMode,
		OrderingClass:		route.OrderingClass,
		RouteCommitment:	route.RouteCommitment,
		CreatedAtHeight:	msg.CreatedHeight,
		Nonce:			msg.Nonce,
	})
}

func ComputeAetherMessageID(msg AetherMessage) string {
	msg = normalizeAetherMessage(msg)
	parts := aetherMessageCanonicalParts(msg)
	parts = append([]string{"aetra-aether-message-id-v1"}, parts...)
	return hashParts(parts...)
}

func ComputeAetherTraceID(msg AetherMessage) string {
	msg = normalizeAetherMessage(msg)
	return hashParts(
		"aetra-aether-message-trace-v1",
		string(msg.SenderZoneID),
		msg.SenderShardID,
		msg.Sender,
		fmt.Sprint(msg.CreatedAtHeight),
		fmt.Sprint(msg.Nonce),
	)
}

func ComputeAetherPayloadHash(payload []byte) string {
	return hashParts("aetra-aether-message-payload-v1", hex.EncodeToString(payload))
}

func CanonicalAetherMessageBinary(msg AetherMessage) ([]byte, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	for _, part := range append([]string{msg.MsgID}, aetherMessageCanonicalParts(msg)...) {
		writeString(buf.Write, part)
	}
	return buf.Bytes(), nil
}

func normalizeAetherMessage(msg AetherMessage) AetherMessage {
	msg.ParentMsgID = strings.ToLower(strings.TrimSpace(msg.ParentMsgID))
	msg.TraceID = strings.ToLower(strings.TrimSpace(msg.TraceID))
	msg.Sender = strings.TrimSpace(msg.Sender)
	msg.SenderShardID = strings.TrimSpace(msg.SenderShardID)
	msg.Receiver = strings.TrimSpace(msg.Receiver)
	msg.ReceiverShardID = strings.TrimSpace(msg.ReceiverShardID)
	msg.Payload = append([]byte(nil), msg.Payload...)
	msg.PayloadType = strings.TrimSpace(msg.PayloadType)
	if msg.ValueNAET.IsNil() {
		msg.ValueNAET = sdkmath.ZeroInt()
	}
	if msg.GasPrice.IsNil() {
		msg.GasPrice = sdkmath.ZeroInt()
	}
	if msg.ForwardingFee.IsNil() {
		msg.ForwardingFee = sdkmath.ZeroInt()
	}
	if msg.ExecutionMode == "" {
		msg.ExecutionMode = ExecutionModeAsync
	}
	if msg.OrderingClass == "" {
		msg.OrderingClass = OrderingClassSenderOrdered
	}
	msg.RouteCommitment = strings.ToLower(strings.TrimSpace(msg.RouteCommitment))
	msg.AuthProof = normalizeAetherProof(msg.AuthProof)
	msg.StateProof = normalizeAetherProof(msg.StateProof)
	msg.Signature = normalizeAetherSignature(msg.Signature)
	return msg
}

func normalizeAetherProof(proof AetherProof) AetherProof {
	proof.ProofType = strings.TrimSpace(proof.ProofType)
	proof.RootHash = strings.ToLower(strings.TrimSpace(proof.RootHash))
	proof.ProofHash = strings.ToLower(strings.TrimSpace(proof.ProofHash))
	return proof
}

func normalizeAetherSignature(signature AetherSignature) AetherSignature {
	signature.Signer = strings.TrimSpace(signature.Signer)
	signature.SignatureHex = strings.ToLower(strings.TrimSpace(signature.SignatureHex))
	return signature
}

func aetherMessageCanonicalParts(msg AetherMessage) []string {
	return []string{
		msg.ParentMsgID,
		msg.TraceID,
		msg.Sender,
		string(msg.SenderZoneID),
		msg.SenderShardID,
		msg.Receiver,
		string(msg.ReceiverZoneID),
		msg.ReceiverShardID,
		msg.ValueNAET.String(),
		ComputeAetherPayloadHash(msg.Payload),
		msg.PayloadType,
		fmt.Sprint(msg.GasLimit),
		msg.GasPrice.String(),
		msg.ForwardingFee.String(),
		fmt.Sprint(msg.ExpiryHeight),
		fmt.Sprint(msg.Bounce),
		string(msg.ExecutionMode),
		string(msg.OrderingClass),
		msg.RouteCommitment,
		msg.AuthProof.ProofType,
		msg.AuthProof.RootHash,
		msg.AuthProof.ProofHash,
		msg.StateProof.ProofType,
		msg.StateProof.RootHash,
		msg.StateProof.ProofHash,
		fmt.Sprint(msg.CreatedAtHeight),
		fmt.Sprint(msg.Nonce),
		msg.Signature.Signer,
		msg.Signature.SignatureHex,
	}
}
