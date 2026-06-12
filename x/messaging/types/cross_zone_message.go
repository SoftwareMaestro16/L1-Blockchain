package types

import (
	"bytes"
	"crypto/sha256"
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
	CrossZoneMessageEncodingVersion	= uint16(1)
	MaxCrossZoneOpcodeLength	= 96
	MaxCrossZoneRouteIDLength	= 128
	MaxCrossZoneAuthScopeLength	= 96
)

type CrossZoneMessageEnvelope struct {
	MessageID	[]byte
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

type CrossZoneMessageParams struct {
	MaxPayloadSize	uint32
	MinGasLimit	uint64
	MaxGasLimit	uint64
	MinFeeLimit	sdkmath.Int
}

type CrossZoneReplayState struct {
	ConsumedMessageIDs	map[string]struct{}
	LastNonceByScope	map[string]uint64
	LastSequenceByScope	map[string]uint64
}

func DefaultCrossZoneMessageParams() CrossZoneMessageParams {
	return CrossZoneMessageParams{
		MaxPayloadSize:	1024 * 1024,
		MinGasLimit:	1,
		MaxGasLimit:	100_000_000,
		MinFeeLimit:	sdkmath.ZeroInt(),
	}
}

func NewCrossZoneMessageEnvelope(msg CrossZoneMessageEnvelope, params CrossZoneMessageParams) (CrossZoneMessageEnvelope, error) {
	if params == (CrossZoneMessageParams{}) {
		params = DefaultCrossZoneMessageParams()
	}
	if len(msg.MessageID) != 0 {
		return CrossZoneMessageEnvelope{}, errors.New("cross-zone message id must be empty before construction")
	}
	if len(msg.PayloadHash) != 0 {
		return CrossZoneMessageEnvelope{}, errors.New("cross-zone payload hash must be empty before construction")
	}
	msg.Sender = append(sdk.AccAddress(nil), msg.Sender...)
	msg.Recipient = append(sdk.AccAddress(nil), msg.Recipient...)
	msg.Payload = append([]byte(nil), msg.Payload...)
	payloadHash := ComputeCrossZonePayloadHash(msg.Payload)
	msg.PayloadHash = payloadHash
	id, err := DeriveCrossZoneMessageID(msg, params)
	if err != nil {
		return CrossZoneMessageEnvelope{}, err
	}
	msg.MessageID = id
	return msg, msg.Validate(params)
}

func DeriveCrossZoneMessageID(msg CrossZoneMessageEnvelope, params CrossZoneMessageParams) ([]byte, error) {
	if params == (CrossZoneMessageParams{}) {
		params = DefaultCrossZoneMessageParams()
	}
	if err := msg.ValidateForID(params); err != nil {
		return nil, err
	}
	encoded, err := EncodeCrossZoneMessageForID(msg)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(encoded)
	return append([]byte(nil), sum[:]...), nil
}

func ComputeCrossZonePayloadHash(payload []byte) []byte {
	sum := sha256.Sum256(payload)
	return append([]byte(nil), sum[:]...)
}

func CanonicalCrossZoneMessageBinary(msg CrossZoneMessageEnvelope) ([]byte, error) {
	if len(msg.MessageID) != MessageIDBytes {
		return nil, fmt.Errorf("cross-zone message id must be %d bytes", MessageIDBytes)
	}
	body, err := EncodeCrossZoneMessageForID(msg)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	writeU32(buf.Write, uint32(CrossZoneMessageEncodingVersion))
	writeBytes(buf.Write, msg.MessageID)
	writeBytes(buf.Write, body)
	return buf.Bytes(), nil
}

func EncodeCrossZoneMessageForID(msg CrossZoneMessageEnvelope) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	writeU32(buf.Write, uint32(CrossZoneMessageEncodingVersion))
	writeCrossZoneString(buf.Write, string(msg.SourceZone))
	writeCrossZoneString(buf.Write, string(msg.DestinationZone))
	writeBytes(buf.Write, msg.Sender)
	writeBytes(buf.Write, msg.Recipient)
	writeCrossZoneString(buf.Write, msg.Value.String())
	writeCrossZoneString(buf.Write, msg.Opcode)
	writeBytes(buf.Write, msg.Payload)
	writeU64(buf.Write, msg.GasLimit)
	writeU64(buf.Write, msg.Deadline)
	writeU64(buf.Write, msg.Nonce)
	writeU64(buf.Write, msg.SourceSequence)
	writeCrossZoneString(buf.Write, msg.RouteID)
	writeBool(buf.Write, msg.Bounce)
	writeCrossZoneString(buf.Write, msg.FeeLimit.String())
	writeU64(buf.Write, msg.CreatedHeight)
	writeBytes(buf.Write, msg.PayloadHash)
	writeCrossZoneString(buf.Write, msg.AuthScope)
	return buf.Bytes(), nil
}

func (m CrossZoneMessageEnvelope) Validate(params CrossZoneMessageParams) error {
	if params == (CrossZoneMessageParams{}) {
		params = DefaultCrossZoneMessageParams()
	}
	if len(m.MessageID) != MessageIDBytes {
		return fmt.Errorf("cross-zone message id must be %d bytes", MessageIDBytes)
	}
	expected, err := DeriveCrossZoneMessageID(m, params)
	if err != nil {
		return err
	}
	if !bytes.Equal(m.MessageID, expected) {
		return errors.New("cross-zone message id mismatch")
	}
	return m.ValidateForID(params)
}

func (m CrossZoneMessageEnvelope) ValidateForID(params CrossZoneMessageParams) error {
	if params == (CrossZoneMessageParams{}) {
		params = DefaultCrossZoneMessageParams()
	}
	if err := params.Validate(); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.DestinationZone); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("cross-zone sender", m.Sender); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("cross-zone recipient", m.Recipient); err != nil {
		return err
	}
	if m.Value.IsNil() || m.Value.IsNegative() {
		return errors.New("cross-zone message value must be non-negative")
	}
	if err := validateCrossZoneToken("cross-zone opcode", m.Opcode, MaxCrossZoneOpcodeLength); err != nil {
		return err
	}
	if len(m.Payload) > int(params.MaxPayloadSize) {
		return fmt.Errorf("cross-zone payload size must be <= %d", params.MaxPayloadSize)
	}
	if m.GasLimit < params.MinGasLimit || m.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("cross-zone gas limit must be in %d..%d", params.MinGasLimit, params.MaxGasLimit)
	}
	if m.CreatedHeight == 0 {
		return errors.New("cross-zone created height must be positive")
	}
	if m.Deadline == 0 || m.Deadline < m.CreatedHeight {
		return errors.New("cross-zone deadline must be at or after created height")
	}
	if m.Nonce == 0 {
		return errors.New("cross-zone nonce must be positive")
	}
	if m.SourceSequence == 0 {
		return errors.New("cross-zone source sequence must be positive")
	}
	if err := validateCrossZoneToken("cross-zone route id", m.RouteID, MaxCrossZoneRouteIDLength); err != nil {
		return err
	}
	if m.FeeLimit.IsNil() || m.FeeLimit.IsNegative() {
		return errors.New("cross-zone fee limit must be non-negative")
	}
	if m.FeeLimit.LT(params.MinFeeLimit) {
		return errors.New("cross-zone fee limit is below minimum")
	}
	if len(m.PayloadHash) != sha256.Size {
		return fmt.Errorf("cross-zone payload hash must be %d bytes", sha256.Size)
	}
	if !bytes.Equal(m.PayloadHash, ComputeCrossZonePayloadHash(m.Payload)) {
		return errors.New("cross-zone payload hash mismatch")
	}
	if err := validateCrossZoneToken("cross-zone auth scope", m.AuthScope, MaxCrossZoneAuthScopeLength); err != nil {
		return err
	}
	return nil
}

func (p CrossZoneMessageParams) Validate() error {
	if p.MaxPayloadSize == 0 {
		return errors.New("cross-zone max payload size must be positive")
	}
	if p.MinGasLimit == 0 {
		return errors.New("cross-zone min gas limit must be positive")
	}
	if p.MaxGasLimit < p.MinGasLimit {
		return errors.New("cross-zone max gas limit must be >= min gas limit")
	}
	if p.MinFeeLimit.IsNil() || p.MinFeeLimit.IsNegative() {
		return errors.New("cross-zone min fee limit must be non-negative")
	}
	return nil
}

func (m CrossZoneMessageEnvelope) ZoneMessage() (zonestypes.ZoneMessage, error) {
	if err := m.Validate(DefaultCrossZoneMessageParams()); err != nil {
		return zonestypes.ZoneMessage{}, err
	}
	return zonestypes.ZoneMessage{
		ZoneID:		m.DestinationZone,
		MessageType:	m.Opcode,
		Source:		CrossZoneAddressScope(m.SourceZone, m.Sender),
		Destination:	CrossZoneAddressScope(m.DestinationZone, m.Recipient),
		GasLimit:	m.GasLimit,
		PayloadHash:	hex.EncodeToString(m.PayloadHash),
		Sequence:	m.SourceSequence,
	}, nil
}

func NewCrossZoneReplayState() CrossZoneReplayState {
	return CrossZoneReplayState{
		ConsumedMessageIDs:	make(map[string]struct{}),
		LastNonceByScope:	make(map[string]uint64),
		LastSequenceByScope:	make(map[string]uint64),
	}
}

func (s CrossZoneReplayState) CheckAndRecord(msg CrossZoneMessageEnvelope, params CrossZoneMessageParams) (CrossZoneReplayState, error) {
	if err := msg.Validate(params); err != nil {
		return CrossZoneReplayState{}, err
	}
	next := s.Clone()
	idKey := hex.EncodeToString(msg.MessageID)
	if _, exists := next.ConsumedMessageIDs[idKey]; exists {
		return CrossZoneReplayState{}, errors.New("cross-zone replay: message_id already consumed")
	}
	nonceScope := CrossZoneNonceScopeKey(msg)
	if last := next.LastNonceByScope[nonceScope]; msg.Nonce <= last {
		return CrossZoneReplayState{}, errors.New("cross-zone replay: nonce must increase within sender scope")
	}
	sequenceScope := CrossZoneSequenceScopeKey(msg)
	if last := next.LastSequenceByScope[sequenceScope]; msg.SourceSequence <= last {
		return CrossZoneReplayState{}, errors.New("cross-zone replay: source sequence must increase within route scope")
	}
	next.ConsumedMessageIDs[idKey] = struct{}{}
	next.LastNonceByScope[nonceScope] = msg.Nonce
	next.LastSequenceByScope[sequenceScope] = msg.SourceSequence
	return next, nil
}

func (s CrossZoneReplayState) Clone() CrossZoneReplayState {
	out := CrossZoneReplayState{
		ConsumedMessageIDs:	make(map[string]struct{}, len(s.ConsumedMessageIDs)),
		LastNonceByScope:	make(map[string]uint64, len(s.LastNonceByScope)),
		LastSequenceByScope:	make(map[string]uint64, len(s.LastSequenceByScope)),
	}
	for key := range s.ConsumedMessageIDs {
		out.ConsumedMessageIDs[key] = struct{}{}
	}
	for key, value := range s.LastNonceByScope {
		out.LastNonceByScope[key] = value
	}
	for key, value := range s.LastSequenceByScope {
		out.LastSequenceByScope[key] = value
	}
	return out
}

func CrossZoneNonceScopeKey(msg CrossZoneMessageEnvelope) string {
	return strings.Join([]string{
		string(msg.SourceZone),
		hex.EncodeToString(msg.Sender),
		msg.AuthScope,
	}, "/")
}

func CrossZoneSequenceScopeKey(msg CrossZoneMessageEnvelope) string {
	return strings.Join([]string{
		string(msg.SourceZone),
		string(msg.DestinationZone),
		msg.RouteID,
		hex.EncodeToString(msg.Sender),
	}, "/")
}

func CrossZoneAddressScope(zoneID zonestypes.ZoneID, address sdk.AccAddress) string {
	return string(zoneID) + "/" + hex.EncodeToString(address)
}

func ComputeCrossZoneMessageRoot(messages []CrossZoneMessageEnvelope, params CrossZoneMessageParams) (string, error) {
	ordered := cloneCrossZoneMessages(messages)
	h := sha256.New()
	writeCrossZoneString(h.Write, "aether-cross-zone-message-root-v1")
	writeU64(h.Write, uint64(len(ordered)))
	for _, msg := range ordered {
		if err := msg.Validate(params); err != nil {
			return "", err
		}
		writeBytes(h.Write, msg.MessageID)
		writeU64(h.Write, msg.SourceSequence)
		writeU64(h.Write, msg.Nonce)
		writeCrossZoneString(h.Write, msg.RouteID)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeCrossZoneString(write func([]byte) (int, error), value string) {
	writeBytes(write, []byte(value))
}

func validateCrossZoneToken(fieldName, value string, maxLen int) error {
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

func cloneCrossZoneMessages(messages []CrossZoneMessageEnvelope) []CrossZoneMessageEnvelope {
	out := make([]CrossZoneMessageEnvelope, len(messages))
	for i, msg := range messages {
		out[i] = msg.Clone()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareCrossZoneMessages(out[i], out[j]) < 0
	})
	return out
}

func (m CrossZoneMessageEnvelope) Clone() CrossZoneMessageEnvelope {
	m.MessageID = append([]byte(nil), m.MessageID...)
	m.Sender = append(sdk.AccAddress(nil), m.Sender...)
	m.Recipient = append(sdk.AccAddress(nil), m.Recipient...)
	m.Payload = append([]byte(nil), m.Payload...)
	m.PayloadHash = append([]byte(nil), m.PayloadHash...)
	return m
}

func compareCrossZoneMessages(left, right CrossZoneMessageEnvelope) int {
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
	if left.SourceSequence < right.SourceSequence {
		return -1
	}
	if left.SourceSequence > right.SourceSequence {
		return 1
	}
	if left.Nonce < right.Nonce {
		return -1
	}
	if left.Nonce > right.Nonce {
		return 1
	}
	leftID := hex.EncodeToString(left.MessageID)
	rightID := hex.EncodeToString(right.MessageID)
	if leftID < rightID {
		return -1
	}
	if leftID > rightID {
		return 1
	}
	return 0
}
