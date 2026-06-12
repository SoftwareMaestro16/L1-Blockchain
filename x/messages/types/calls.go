package types

import (
	"bytes"
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
	MaxPayloadTypeLength	= 96
	MaxReplyModeLength	= 32

	ReplyModeNone	= "none"
	ReplyModeCaller	= "caller"
	ReplyModeCallee	= "callee"
	ReplyModeError	= "error"
)

type MsgCrossZoneCall struct {
	Caller			sdk.AccAddress
	Callee			sdk.AccAddress
	SourceZoneID		zonestypes.ZoneID
	DestinationZoneID	zonestypes.ZoneID
	PayloadType		string
	Payload			[]byte
	ValueNAET		sdkmath.Int
	GasLimit		uint64
	ForwardingFee		sdkmath.Int
	ReplyMode		string
	ExpiryHeight		uint64
}

type CrossZoneCallEscrow struct {
	Caller			sdk.AccAddress
	SourceZoneID		zonestypes.ZoneID
	DestinationZoneID	zonestypes.ZoneID
	ValueNAET		sdkmath.Int
	ForwardingFee		sdkmath.Int
	ExpiryHeight		uint64
	Escrowed		bool
}

type CrossZoneCallAdmission struct {
	CreatedHeight		uint64
	Nonce			uint64
	SourceSequence		uint64
	SupportedPayloadTypes	[]DestinationPayloadTypes
	SupportedReplyModes	[]string
	Escrows			[]CrossZoneCallEscrow
}

type DestinationPayloadTypes struct {
	ZoneID		zonestypes.ZoneID
	PayloadTypes	[]string
}

func (k MessageKeeper) SubmitCrossZoneCall(req MsgCrossZoneCall, admission CrossZoneCallAdmission) (MessageKeeper, SubmitCrossZoneMessageResponse, error) {
	msg, err := NewMessageFromCrossZoneCall(req, admission, k.state.Params)
	if err != nil {
		return MessageKeeper{}, SubmitCrossZoneMessageResponse{}, err
	}
	return k.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutIDs(msg)})
}

func (s KeeperMsgServer) SubmitCrossZoneCall(req MsgCrossZoneCall, admission CrossZoneCallAdmission) (SubmitCrossZoneMessageResponse, error) {
	if s.keeper == nil {
		return SubmitCrossZoneMessageResponse{}, errors.New("message keeper is required")
	}
	next, resp, err := s.keeper.SubmitCrossZoneCall(req, admission)
	if err != nil {
		return SubmitCrossZoneMessageResponse{}, err
	}
	*s.keeper = next
	return resp, nil
}

func NewMessageFromCrossZoneCall(req MsgCrossZoneCall, admission CrossZoneCallAdmission, params MessageParams) (Message, error) {
	req = req.Normalize()
	admission = admission.Normalize()
	if err := req.Validate(admission, params); err != nil {
		return Message{}, err
	}
	return NewMessage(Message{
		SourceZone:		req.SourceZoneID,
		DestinationZone:	req.DestinationZoneID,
		Sender:			req.Caller,
		Recipient:		req.Callee,
		Value:			req.ValueNAET,
		Opcode:			req.PayloadType,
		Payload:		req.Payload,
		GasLimit:		req.GasLimit,
		Deadline:		req.ExpiryHeight,
		Nonce:			admission.Nonce,
		SourceSequence:		admission.SourceSequence,
		RouteID:		CrossZoneCallRouteID(req),
		Bounce:			req.ReplyMode != ReplyModeNone,
		FeeLimit:		req.ForwardingFee,
		CreatedHeight:		admission.CreatedHeight,
		AuthScope:		"cross-zone-call/" + req.ReplyMode,
	}, params)
}

func (m MsgCrossZoneCall) Normalize() MsgCrossZoneCall {
	m.Caller = append(sdk.AccAddress(nil), m.Caller...)
	m.Callee = append(sdk.AccAddress(nil), m.Callee...)
	m.PayloadType = strings.TrimSpace(m.PayloadType)
	m.Payload = append([]byte(nil), m.Payload...)
	if m.ValueNAET.IsNil() {
		m.ValueNAET = sdkmath.ZeroInt()
	}
	if m.ForwardingFee.IsNil() {
		m.ForwardingFee = sdkmath.ZeroInt()
	}
	m.ReplyMode = strings.ToLower(strings.TrimSpace(m.ReplyMode))
	if m.ReplyMode == "" {
		m.ReplyMode = ReplyModeNone
	}
	return m
}

func (m MsgCrossZoneCall) Validate(admission CrossZoneCallAdmission, params MessageParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	m = m.Normalize()
	admission = admission.Normalize()
	if err := admission.Validate(); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("cross-zone call caller", m.Caller); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("cross-zone call callee", m.Callee); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.SourceZoneID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.DestinationZoneID); err != nil {
		return err
	}
	if err := validateToken("cross-zone call payload type", m.PayloadType, MaxPayloadTypeLength); err != nil {
		return err
	}
	if len(m.Payload) > int(params.MaxPayloadSize) {
		return fmt.Errorf("cross-zone call payload size must be <= %d", params.MaxPayloadSize)
	}
	if m.ValueNAET.IsNegative() {
		return errors.New("cross-zone call value must be non-negative")
	}
	if m.ForwardingFee.IsNegative() || m.ForwardingFee.LT(params.MinFeeLimit) {
		return errors.New("cross-zone call forwarding fee is below minimum")
	}
	if m.GasLimit < params.MinGasLimit || m.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("cross-zone call gas limit must be in %d..%d", params.MinGasLimit, params.MaxGasLimit)
	}
	if m.ExpiryHeight == 0 || m.ExpiryHeight < admission.CreatedHeight {
		return errors.New("cross-zone call expiry height must be at or after created height")
	}
	if !admission.SupportsPayloadType(m.DestinationZoneID, m.PayloadType) {
		return errors.New("cross-zone call destination does not support payload type")
	}
	if !admission.SupportsReplyMode(m.ReplyMode) {
		return errors.New("cross-zone call reply mode is not supported")
	}
	if !admission.HasEscrow(m) {
		return errors.New("cross-zone call value escrow must succeed before enqueue")
	}
	return nil
}

func (a CrossZoneCallAdmission) Normalize() CrossZoneCallAdmission {
	a.SupportedPayloadTypes = normalizeDestinationPayloadTypes(a.SupportedPayloadTypes)
	a.SupportedReplyModes = normalizeStringList(a.SupportedReplyModes)
	if len(a.SupportedReplyModes) == 0 {
		a.SupportedReplyModes = []string{ReplyModeNone, ReplyModeCaller, ReplyModeCallee, ReplyModeError}
	}
	a.Escrows = normalizeCrossZoneCallEscrows(a.Escrows)
	return a
}

func (a CrossZoneCallAdmission) Validate() error {
	if a.CreatedHeight == 0 {
		return errors.New("cross-zone call created height must be positive")
	}
	if a.Nonce == 0 || a.SourceSequence == 0 {
		return errors.New("cross-zone call nonce and source sequence must be positive")
	}
	for _, destination := range a.SupportedPayloadTypes {
		if err := zonestypes.ValidateZoneID(destination.ZoneID); err != nil {
			return err
		}
		if len(destination.PayloadTypes) == 0 {
			return errors.New("cross-zone call destination payload types are required")
		}
		for _, payloadType := range destination.PayloadTypes {
			if err := validateToken("cross-zone call supported payload type", payloadType, MaxPayloadTypeLength); err != nil {
				return err
			}
		}
	}
	for _, mode := range a.SupportedReplyModes {
		if !IsReplyMode(mode) {
			return fmt.Errorf("unknown cross-zone call reply mode %q", mode)
		}
	}
	for _, escrow := range a.Escrows {
		if err := escrow.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a CrossZoneCallAdmission) SupportsPayloadType(zoneID zonestypes.ZoneID, payloadType string) bool {
	payloadType = strings.TrimSpace(payloadType)
	for _, destination := range a.SupportedPayloadTypes {
		if destination.ZoneID != zoneID {
			continue
		}
		for _, supported := range destination.PayloadTypes {
			if supported == payloadType {
				return true
			}
		}
	}
	return false
}

func (a CrossZoneCallAdmission) SupportsReplyMode(mode string) bool {
	mode = strings.ToLower(strings.TrimSpace(mode))
	for _, supported := range a.SupportedReplyModes {
		if supported == mode {
			return true
		}
	}
	return false
}

func (a CrossZoneCallAdmission) HasEscrow(msg MsgCrossZoneCall) bool {
	msg = msg.Normalize()
	for _, escrow := range a.Escrows {
		escrow = escrow.Normalize()
		if !escrow.Escrowed {
			continue
		}
		if !escrow.Matches(msg) {
			continue
		}
		if escrow.ValueNAET.LT(msg.ValueNAET) || escrow.ForwardingFee.LT(msg.ForwardingFee) {
			continue
		}
		if escrow.ExpiryHeight < msg.ExpiryHeight {
			continue
		}
		return true
	}
	return false
}

func (e CrossZoneCallEscrow) Normalize() CrossZoneCallEscrow {
	e.Caller = append(sdk.AccAddress(nil), e.Caller...)
	if e.ValueNAET.IsNil() {
		e.ValueNAET = sdkmath.ZeroInt()
	}
	if e.ForwardingFee.IsNil() {
		e.ForwardingFee = sdkmath.ZeroInt()
	}
	return e
}

func (e CrossZoneCallEscrow) Validate() error {
	e = e.Normalize()
	if err := addressing.RejectZeroAddress("cross-zone call escrow caller", e.Caller); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(e.SourceZoneID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(e.DestinationZoneID); err != nil {
		return err
	}
	if e.ValueNAET.IsNegative() || e.ForwardingFee.IsNegative() {
		return errors.New("cross-zone call escrow amounts must be non-negative")
	}
	if e.ExpiryHeight == 0 {
		return errors.New("cross-zone call escrow expiry height must be positive")
	}
	return nil
}

func (e CrossZoneCallEscrow) Matches(msg MsgCrossZoneCall) bool {
	return bytes.Equal(e.Caller, msg.Caller) &&
		e.SourceZoneID == msg.SourceZoneID &&
		e.DestinationZoneID == msg.DestinationZoneID
}

func IsReplyMode(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ReplyModeNone, ReplyModeCaller, ReplyModeCallee, ReplyModeError:
		return true
	default:
		return false
	}
}

func CrossZoneCallRouteID(msg MsgCrossZoneCall) string {
	msg = msg.Normalize()
	return strings.Join([]string{
		"call",
		string(msg.SourceZoneID),
		string(msg.DestinationZoneID),
		msg.PayloadType,
		hex.EncodeToString(msg.Caller),
	}, "/")
}

func normalizeDestinationPayloadTypes(values []DestinationPayloadTypes) []DestinationPayloadTypes {
	out := make([]DestinationPayloadTypes, len(values))
	for i, value := range values {
		out[i] = DestinationPayloadTypes{
			ZoneID:		value.ZoneID,
			PayloadTypes:	normalizeStringList(value.PayloadTypes),
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ZoneID < out[j].ZoneID
	})
	return out
}

func normalizeCrossZoneCallEscrows(values []CrossZoneCallEscrow) []CrossZoneCallEscrow {
	out := make([]CrossZoneCallEscrow, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if cmp := bytes.Compare(out[i].Caller, out[j].Caller); cmp != 0 {
			return cmp < 0
		}
		if out[i].SourceZoneID != out[j].SourceZoneID {
			return out[i].SourceZoneID < out[j].SourceZoneID
		}
		return out[i].DestinationZoneID < out[j].DestinationZoneID
	})
	return out
}

func normalizeStringList(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func msgWithoutIDs(msg Message) Message {
	msg.MessageID = nil
	msg.PayloadHash = nil
	return msg
}
