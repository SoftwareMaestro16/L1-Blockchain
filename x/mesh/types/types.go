package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MaxZoneIDLength	= 64
	MaxShardIDBytes	= 128
	HashHexLength	= 64
	MaxFinalityAge	= 1_000_000
)

type ZoneID string

type ShardID string

type MessageKind string

const (
	MessageKindNormal	MessageKind	= "NORMAL"
	MessageKindBounce	MessageKind	= "BOUNCE"
	MessageKindRefund	MessageKind	= "REFUND"
)

type ReceiptStatus string

const (
	ReceiptStatusSuccess		ReceiptStatus	= "SUCCESS"
	ReceiptStatusBounced		ReceiptStatus	= "BOUNCED"
	ReceiptStatusRefunded		ReceiptStatus	= "REFUNDED"
	ReceiptStatusTerminalFailure	ReceiptStatus	= "TERMINAL_FAILURE"
)

type FailureReason string

const (
	FailureReasonNone		FailureReason	= ""
	FailureReasonInvalidDestination	FailureReason	= "INVALID_DESTINATION"
	FailureReasonExpired		FailureReason	= "EXPIRED"
	FailureReasonExecutionFailed	FailureReason	= "EXECUTION_FAILED"
)

type MeshParams struct {
	MaxFinalityAge uint64
}

type MeshDestination struct {
	ZoneID	ZoneID
	ShardID	ShardID
	Active	bool
}

type FinalityReference struct {
	Height		uint64
	CommitmentHash	string
}

type FinalizedCommitment struct {
	ZoneID		ZoneID
	ShardID		ShardID
	Height		uint64
	CommitmentHash	string
	MessageRoot	string
	ReceiptRoot	string
}

type MeshProof struct {
	SourceCommitment	string
	MessageRoot		string
	ProofHash		string
}

type MeshMessage struct {
	SourceZone		ZoneID
	SourceShard		ShardID
	DestinationZone		ZoneID
	DestinationShard	ShardID
	MessageID		string
	Nonce			uint64
	Sender			[]byte
	Recipient		[]byte
	AssetCommitment		string
	PayloadHash		string
	TimeoutHeight		uint64
	Finality		FinalityReference
	Proof			MeshProof
	Sequence		uint64
	SourceLogicalTime	uint64
	Kind			MessageKind
	ParentMessageID		string
}

type ExecutionResult struct {
	Success		bool
	Code		uint32
	ResultHash	string
}

type MeshReceipt struct {
	MessageID		string
	SourceZone		ZoneID
	SourceShard		ShardID
	DestinationZone		ZoneID
	DestinationShard	ShardID
	Status			ReceiptStatus
	Reason			FailureReason
	Height			uint64
	Sequence		uint64
	ExecutionCode		uint32
	ResultHash		string
	ReceiptHash		string
}

type ReplayMarker struct {
	MessageID	string
	ReceiptHash	string
	Reason		FailureReason
	Height		uint64
}

type BounceReceipt struct {
	MessageID		string
	SourceMessageID		string
	BounceMessageID		string
	DestinationZone		ZoneID
	DestinationShard	ShardID
	Reason			FailureReason
	Height			uint64
	ReceiptHash		string
}

type RefundReceipt struct {
	MessageID	string
	SourceMessageID	string
	Recipient	[]byte
	AssetCommitment	string
	Reason		FailureReason
	Height		uint64
	ReceiptHash	string
}

type MeshState struct {
	CurrentHeight		uint64
	Params			MeshParams
	Destinations		[]MeshDestination
	FinalizedCommitments	[]FinalizedCommitment
	ReplayMarkers		[]ReplayMarker
	Receipts		[]MeshReceipt
	BounceReceipts		[]BounceReceipt
	RefundReceipts		[]RefundReceipt
}

func DefaultParams() MeshParams {
	return MeshParams{MaxFinalityAge: 256}
}

func NormalizeParams(params MeshParams) MeshParams {
	if params.MaxFinalityAge == 0 {
		return DefaultParams()
	}
	return params
}

func (p MeshParams) Validate() error {
	if p.MaxFinalityAge == 0 {
		return errors.New("mesh max finality age must be positive")
	}
	if p.MaxFinalityAge > MaxFinalityAge {
		return fmt.Errorf("mesh max finality age must be <= %d", MaxFinalityAge)
	}
	return nil
}

func (d MeshDestination) Validate() error {
	if err := ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	return ValidateShardID(d.ShardID)
}

func (f FinalityReference) Validate() error {
	if f.Height == 0 {
		return errors.New("mesh finality height must be positive")
	}
	return ValidateHash("mesh finality commitment", f.CommitmentHash)
}

func (c FinalizedCommitment) Validate() error {
	if err := ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(c.ShardID); err != nil {
		return err
	}
	if c.Height == 0 {
		return errors.New("mesh finalized commitment height must be positive")
	}
	if err := ValidateHash("mesh finalized commitment", c.CommitmentHash); err != nil {
		return err
	}
	if err := ValidateHash("mesh finalized message root", c.MessageRoot); err != nil {
		return err
	}
	return ValidateHash("mesh finalized receipt root", c.ReceiptRoot)
}

func (p MeshProof) Validate() error {
	if err := ValidateHash("mesh proof source commitment", p.SourceCommitment); err != nil {
		return err
	}
	if err := ValidateHash("mesh proof message root", p.MessageRoot); err != nil {
		return err
	}
	return ValidateHash("mesh proof hash", p.ProofHash)
}

func (m MeshMessage) Normalize() MeshMessage {
	m.Sender = cloneBytes(m.Sender)
	m.Recipient = cloneBytes(m.Recipient)
	if m.Kind == "" {
		m.Kind = MessageKindNormal
	}
	if m.MessageID == "" {
		m.MessageID = ComputeMessageID(m)
	}
	return m
}

func (m MeshMessage) Validate() error {
	m = m.Normalize()
	if err := ValidateZoneID(m.SourceZone); err != nil {
		return err
	}
	if err := ValidateShardID(m.SourceShard); err != nil {
		return err
	}
	if err := ValidateZoneID(m.DestinationZone); err != nil {
		return err
	}
	if err := ValidateShardID(m.DestinationShard); err != nil {
		return err
	}
	if err := ValidateHash("mesh message id", m.MessageID); err != nil {
		return err
	}
	if expected := ComputeMessageID(m); m.MessageID != expected {
		return errors.New("mesh message id mismatch")
	}
	if isZeroBytes(m.Sender) {
		return errors.New("mesh sender must not be zero address")
	}
	if len(m.Sender) == 0 {
		return errors.New("mesh sender is required")
	}
	if isZeroBytes(m.Recipient) {
		return errors.New("mesh recipient must not be zero address")
	}
	if len(m.Recipient) == 0 {
		return errors.New("mesh recipient is required")
	}
	if err := ValidateHash("mesh asset commitment", m.AssetCommitment); err != nil {
		return err
	}
	if err := ValidateHash("mesh payload hash", m.PayloadHash); err != nil {
		return err
	}
	if m.TimeoutHeight == 0 {
		return errors.New("mesh timeout height must be positive")
	}
	if err := m.Finality.Validate(); err != nil {
		return err
	}
	if !IsMessageKind(m.Kind) {
		return fmt.Errorf("unknown mesh message kind %q", m.Kind)
	}
	if m.Kind == MessageKindNormal && m.ParentMessageID != "" {
		return errors.New("normal mesh message must not have parent message id")
	}
	if m.Kind != MessageKindNormal {
		if err := ValidateHash("mesh parent message id", m.ParentMessageID); err != nil {
			return err
		}
	}
	return nil
}

func (r ExecutionResult) Validate() error {
	return ValidateHash("mesh execution result hash", r.ResultHash)
}

func (r MeshReceipt) Validate() error {
	if err := ValidateHash("mesh receipt message id", r.MessageID); err != nil {
		return err
	}
	if err := ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := ValidateShardID(r.SourceShard); err != nil {
		return err
	}
	if err := ValidateZoneID(r.DestinationZone); err != nil {
		return err
	}
	if err := ValidateShardID(r.DestinationShard); err != nil {
		return err
	}
	if !IsReceiptStatus(r.Status) {
		return fmt.Errorf("unknown mesh receipt status %q", r.Status)
	}
	if r.Status == ReceiptStatusSuccess && r.Reason != FailureReasonNone {
		return errors.New("successful mesh receipt must not include failure reason")
	}
	if r.Status != ReceiptStatusSuccess && !IsFailureReason(r.Reason) {
		return fmt.Errorf("unknown mesh receipt failure reason %q", r.Reason)
	}
	if r.Height == 0 {
		return errors.New("mesh receipt height must be positive")
	}
	if err := ValidateHash("mesh receipt result hash", r.ResultHash); err != nil {
		return err
	}
	if err := ValidateHash("mesh receipt hash", r.ReceiptHash); err != nil {
		return err
	}
	if expected := ComputeReceiptHash(r); r.ReceiptHash != expected {
		return errors.New("mesh receipt hash mismatch")
	}
	return nil
}

func (m ReplayMarker) Validate() error {
	if err := ValidateHash("mesh replay marker message id", m.MessageID); err != nil {
		return err
	}
	if err := ValidateHash("mesh replay marker receipt hash", m.ReceiptHash); err != nil {
		return err
	}
	if m.Height == 0 {
		return errors.New("mesh replay marker height must be positive")
	}
	if m.Reason != FailureReasonNone && !IsFailureReason(m.Reason) {
		return fmt.Errorf("unknown mesh replay marker reason %q", m.Reason)
	}
	return nil
}

func (b BounceReceipt) Validate() error {
	if err := ValidateHash("mesh bounce receipt id", b.MessageID); err != nil {
		return err
	}
	if err := ValidateHash("mesh bounce source message id", b.SourceMessageID); err != nil {
		return err
	}
	if err := ValidateHash("mesh bounce message id", b.BounceMessageID); err != nil {
		return err
	}
	if err := ValidateZoneID(b.DestinationZone); err != nil {
		return err
	}
	if err := ValidateShardID(b.DestinationShard); err != nil {
		return err
	}
	if !IsFailureReason(b.Reason) || b.Reason == FailureReasonNone {
		return fmt.Errorf("invalid mesh bounce reason %q", b.Reason)
	}
	if b.Height == 0 {
		return errors.New("mesh bounce height must be positive")
	}
	if err := ValidateHash("mesh bounce receipt hash", b.ReceiptHash); err != nil {
		return err
	}
	if expected := ComputeBounceReceiptHash(b); b.ReceiptHash != expected {
		return errors.New("mesh bounce receipt hash mismatch")
	}
	return nil
}

func (r RefundReceipt) Validate() error {
	if err := ValidateHash("mesh refund receipt id", r.MessageID); err != nil {
		return err
	}
	if err := ValidateHash("mesh refund source message id", r.SourceMessageID); err != nil {
		return err
	}
	if len(r.Recipient) == 0 || isZeroBytes(r.Recipient) {
		return errors.New("mesh refund recipient is required")
	}
	if err := ValidateHash("mesh refund asset commitment", r.AssetCommitment); err != nil {
		return err
	}
	if !IsFailureReason(r.Reason) || r.Reason == FailureReasonNone {
		return fmt.Errorf("invalid mesh refund reason %q", r.Reason)
	}
	if r.Height == 0 {
		return errors.New("mesh refund height must be positive")
	}
	if err := ValidateHash("mesh refund receipt hash", r.ReceiptHash); err != nil {
		return err
	}
	if expected := ComputeRefundReceiptHash(r); r.ReceiptHash != expected {
		return errors.New("mesh refund receipt hash mismatch")
	}
	return nil
}

func ValidateZoneID(id ZoneID) error {
	text := string(id)
	if strings.TrimSpace(text) != text || text == "" {
		return errors.New("mesh zone id is required and must not have surrounding whitespace")
	}
	if len(text) > MaxZoneIDLength {
		return fmt.Errorf("mesh zone id must be <= %d bytes", MaxZoneIDLength)
	}
	for i, r := range text {
		if i == 0 && (r < 'A' || r > 'Z') {
			return errors.New("mesh zone id must start with A-Z")
		}
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			continue
		}
		return errors.New("mesh zone id must contain only A-Z, 0-9, or underscore")
	}
	return nil
}

func ValidateShardID(id ShardID) error {
	text := string(id)
	if strings.TrimSpace(text) != text || text == "" {
		return errors.New("mesh shard id is required and must not have surrounding whitespace")
	}
	if len(text) > MaxShardIDBytes {
		return fmt.Errorf("mesh shard id must be <= %d bytes", MaxShardIDBytes)
	}
	for _, r := range text {
		if r <= ' ' || r == 0x7f {
			return errors.New("mesh shard id must not contain whitespace or control characters")
		}
	}
	return nil
}

func ValidateHash(fieldName, value string) error {
	if len(value) != HashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	return nil
}

func IsMessageKind(kind MessageKind) bool {
	switch kind {
	case MessageKindNormal, MessageKindBounce, MessageKindRefund:
		return true
	default:
		return false
	}
}

func IsReceiptStatus(status ReceiptStatus) bool {
	switch status {
	case ReceiptStatusSuccess, ReceiptStatusBounced, ReceiptStatusRefunded, ReceiptStatusTerminalFailure:
		return true
	default:
		return false
	}
}

func IsFailureReason(reason FailureReason) bool {
	switch reason {
	case FailureReasonNone, FailureReasonInvalidDestination, FailureReasonExpired, FailureReasonExecutionFailed:
		return true
	default:
		return false
	}
}
