package types

import (
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
	ZoneTransferPayloadType	= "zone.transfer"
	ZoneTransferAuthScope	= "financial-zone-transfer"
	MaxDenomLength		= 128
	MaxDeliveryWindow	= uint64(1_000_000)
)

type MsgZoneTransfer struct {
	FromAddress		sdk.AccAddress
	ToAddress		sdk.AccAddress
	SourceZoneID		zonestypes.ZoneID
	DestinationZoneID	zonestypes.ZoneID
	Amount			sdkmath.Int
	Denom			string
	GasLimit		uint64
	ForwardingFee		sdkmath.Int
	ExpiryHeight		uint64
	MemoHashOptional	string
}

type ZoneTransferAdmission struct {
	CreatedHeight		uint64
	Nonce			uint64
	SourceSequence		uint64
	SourceSpendable		sdkmath.Int
	EnabledZones		[]zonestypes.ZoneID
	RoutableDenoms		[]RoutableDenom
	MinimumRouteFees	[]ZoneTransferRouteFee
	SourceShardID		string
	DestinationShardID	string
	MaxDeliveryWindow	uint64
	CommittedRouteHash	string
	SourceEscrowed		bool
	DestinationCredited	bool
}

type RoutableDenom struct {
	Denom			string
	SourceZoneID		zonestypes.ZoneID
	DestinationZoneID	zonestypes.ZoneID
	AuthorityPath		string
}

type ZoneTransferRouteFee struct {
	SourceZoneID		zonestypes.ZoneID
	DestinationZoneID	zonestypes.ZoneID
	Denom			string
	MinimumFee		sdkmath.Int
}

type MsgZoneTransferResult struct {
	Message		Message
	AetherMessage	AetherMessage
	Escrow		AetherValueEscrow
	Receipt		AetherMessageReceipt
}

func (k MessageKeeper) SubmitZoneTransfer(req MsgZoneTransfer, admission ZoneTransferAdmission) (MessageKeeper, SubmitCrossZoneMessageResponse, error) {
	msg, err := NewMessageFromZoneTransfer(req, admission, k.state.Params)
	if err != nil {
		return MessageKeeper{}, SubmitCrossZoneMessageResponse{}, err
	}
	return k.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutIDs(msg)})
}

func (s KeeperMsgServer) SubmitZoneTransfer(req MsgZoneTransfer, admission ZoneTransferAdmission) (SubmitCrossZoneMessageResponse, error) {
	if s.keeper == nil {
		return SubmitCrossZoneMessageResponse{}, errors.New("message keeper is required")
	}
	next, resp, err := s.keeper.SubmitZoneTransfer(req, admission)
	if err != nil {
		return SubmitCrossZoneMessageResponse{}, err
	}
	*s.keeper = next
	return resp, nil
}

func NewMessageFromZoneTransfer(req MsgZoneTransfer, admission ZoneTransferAdmission, params MessageParams) (Message, error) {
	req = req.Normalize()
	admission = admission.Normalize()
	if err := req.Validate(admission, params); err != nil {
		return Message{}, err
	}
	return NewMessage(Message{
		SourceZone:		req.SourceZoneID,
		DestinationZone:	req.DestinationZoneID,
		Sender:			req.FromAddress,
		Recipient:		req.ToAddress,
		Value:			req.Amount,
		Opcode:			ZoneTransferPayloadType,
		Payload:		req.CanonicalPayload(),
		GasLimit:		req.GasLimit,
		Deadline:		req.ExpiryHeight,
		Nonce:			admission.Nonce,
		SourceSequence:		admission.SourceSequence,
		RouteID:		ZoneTransferRouteID(req),
		Bounce:			true,
		FeeLimit:		req.ForwardingFee,
		CreatedHeight:		admission.CreatedHeight,
		AuthScope:		ZoneTransferAuthScope,
	}, params)
}

func NewAetherMessageFromZoneTransfer(req MsgZoneTransfer, admission ZoneTransferAdmission, params MessageParams) (AetherMessage, error) {
	msg, err := NewMessageFromZoneTransfer(req, admission, params)
	if err != nil {
		return AetherMessage{}, err
	}
	routeCommitment := admission.CommittedRouteHash
	if routeCommitment == "" {
		routeCommitment = ComputeZoneTransferRouteCommitment(req, admission)
	}
	return NewAetherMessage(AetherMessage{
		Sender:			hex.EncodeToString(msg.Sender),
		SenderZoneID:		msg.SourceZone,
		SenderShardID:		admission.SourceShardID,
		Receiver:		hex.EncodeToString(msg.Recipient),
		ReceiverZoneID:		msg.DestinationZone,
		ReceiverShardID:	admission.DestinationShardID,
		ValueNAET:		msg.Value,
		Payload:		msg.Payload,
		PayloadType:		ZoneTransferPayloadType,
		GasLimit:		msg.GasLimit,
		GasPrice:		msg.FeeLimit,
		ForwardingFee:		msg.FeeLimit,
		ExpiryHeight:		msg.Deadline,
		Bounce:			true,
		ExecutionMode:		ExecutionModeAsync,
		OrderingClass:		OrderingClassSenderOrdered,
		RouteCommitment:	routeCommitment,
		CreatedAtHeight:	msg.CreatedHeight,
		Nonce:			msg.Nonce,
	})
}

func BuildZoneTransferAcceptedResult(req MsgZoneTransfer, admission ZoneTransferAdmission, params MessageParams, receiptHeight uint64) (MsgZoneTransferResult, error) {
	msg, err := NewMessageFromZoneTransfer(req, admission, params)
	if err != nil {
		return MsgZoneTransferResult{}, err
	}
	aether, err := NewAetherMessageFromZoneTransfer(req, admission, params)
	if err != nil {
		return MsgZoneTransferResult{}, err
	}
	escrow, err := NewAetherValueEscrow(AetherValueEscrow{
		MsgID:		aether.MsgID,
		ValueLocked:	aether.ValueNAET,
		FeeLocked:	aether.ForwardingFee,
		Status:		AetherEscrowLocked,
	})
	if err != nil {
		return MsgZoneTransferResult{}, err
	}
	receipt, err := AetherReceiptFromMessage(aether, ReceiptStatusAccepted, receiptHeight, 0, sdkmath.ZeroInt(), nil, "", EmptyHash(), ComputeZoneTransferStateWriteSummary(req, admission))
	if err != nil {
		return MsgZoneTransferResult{}, err
	}
	return MsgZoneTransferResult{Message: msg, AetherMessage: aether, Escrow: escrow, Receipt: receipt}, nil
}

func (m MsgZoneTransfer) Normalize() MsgZoneTransfer {
	m.FromAddress = append(sdk.AccAddress(nil), m.FromAddress...)
	m.ToAddress = append(sdk.AccAddress(nil), m.ToAddress...)
	if m.Amount.IsNil() {
		m.Amount = sdkmath.ZeroInt()
	}
	m.Denom = strings.TrimSpace(m.Denom)
	if m.ForwardingFee.IsNil() {
		m.ForwardingFee = sdkmath.ZeroInt()
	}
	m.MemoHashOptional = strings.ToLower(strings.TrimSpace(m.MemoHashOptional))
	return m
}

func (m MsgZoneTransfer) Validate(admission ZoneTransferAdmission, params MessageParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	m = m.Normalize()
	admission = admission.Normalize()
	if err := admission.Validate(); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("zone transfer from address", m.FromAddress); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("zone transfer to address", m.ToAddress); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.SourceZoneID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(m.DestinationZoneID); err != nil {
		return err
	}
	if m.SourceZoneID == m.DestinationZoneID {
		return errors.New("zone transfer requires distinct source and destination zones")
	}
	if !m.Amount.IsPositive() {
		return errors.New("zone transfer amount must be positive")
	}
	if err := validateDenom("zone transfer denom", m.Denom); err != nil {
		return err
	}
	if m.GasLimit < params.MinGasLimit || m.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("zone transfer gas limit must be in %d..%d", params.MinGasLimit, params.MaxGasLimit)
	}
	if m.ForwardingFee.IsNegative() || m.ForwardingFee.LT(params.MinFeeLimit) {
		return errors.New("zone transfer forwarding fee is below minimum")
	}
	if m.ExpiryHeight == 0 || m.ExpiryHeight < admission.CreatedHeight {
		return errors.New("zone transfer expiry height must be at or after created height")
	}
	maxWindow := admission.MaxDeliveryWindow
	if maxWindow == 0 {
		maxWindow = MaxDeliveryWindow
	}
	if m.ExpiryHeight-admission.CreatedHeight > maxWindow {
		return errors.New("zone transfer expiry exceeds max delivery window")
	}
	if m.MemoHashOptional != "" {
		if err := zonestypes.ValidateHash("zone transfer memo hash", m.MemoHashOptional); err != nil {
			return err
		}
	}
	if admission.SourceSpendable.LT(m.Amount) {
		return errors.New("zone transfer source balance is insufficient")
	}
	if !admission.DestinationZoneEnabled(m.DestinationZoneID) {
		return errors.New("zone transfer destination zone is not enabled")
	}
	if !admission.DenomRoutable(m.SourceZoneID, m.DestinationZoneID, m.Denom) {
		return errors.New("zone transfer denom is not routable")
	}
	if m.ForwardingFee.LT(admission.MinimumFee(m.SourceZoneID, m.DestinationZoneID, m.Denom)) {
		return errors.New("zone transfer forwarding fee does not cover route")
	}
	if !admission.SourceEscrowed {
		return errors.New("zone transfer source value must be escrowed before enqueue")
	}
	return nil
}

func (m MsgZoneTransfer) CanonicalPayload() []byte {
	m = m.Normalize()
	parts := []string{
		"aetra-msg-zone-transfer-payload-v1",
		hex.EncodeToString(m.FromAddress),
		hex.EncodeToString(m.ToAddress),
		string(m.SourceZoneID),
		string(m.DestinationZoneID),
		m.Amount.String(),
		m.Denom,
		fmt.Sprint(m.GasLimit),
		m.ForwardingFee.String(),
		fmt.Sprint(m.ExpiryHeight),
		m.MemoHashOptional,
	}
	return []byte(strings.Join(parts, "|"))
}

func (a ZoneTransferAdmission) Normalize() ZoneTransferAdmission {
	if a.SourceSpendable.IsNil() {
		a.SourceSpendable = sdkmath.ZeroInt()
	}
	a.EnabledZones = normalizeZoneIDs(a.EnabledZones)
	a.RoutableDenoms = normalizeRoutableDenoms(a.RoutableDenoms)
	a.MinimumRouteFees = normalizeZoneTransferRouteFees(a.MinimumRouteFees)
	a.SourceShardID = strings.TrimSpace(a.SourceShardID)
	a.DestinationShardID = strings.TrimSpace(a.DestinationShardID)
	a.CommittedRouteHash = strings.ToLower(strings.TrimSpace(a.CommittedRouteHash))
	return a
}

func (a ZoneTransferAdmission) Validate() error {
	a = a.Normalize()
	if a.CreatedHeight == 0 {
		return errors.New("zone transfer created height must be positive")
	}
	if a.Nonce == 0 || a.SourceSequence == 0 {
		return errors.New("zone transfer nonce and source sequence must be positive")
	}
	if a.SourceSpendable.IsNegative() {
		return errors.New("zone transfer source spendable must be non-negative")
	}
	if len(a.EnabledZones) == 0 {
		return errors.New("zone transfer enabled zones are required")
	}
	for _, zoneID := range a.EnabledZones {
		if err := zonestypes.ValidateZoneID(zoneID); err != nil {
			return err
		}
	}
	if len(a.RoutableDenoms) == 0 {
		return errors.New("zone transfer routable denoms are required")
	}
	for _, denom := range a.RoutableDenoms {
		if err := denom.Validate(); err != nil {
			return err
		}
	}
	for _, fee := range a.MinimumRouteFees {
		if err := fee.Validate(); err != nil {
			return err
		}
	}
	if err := validateToken("zone transfer source shard id", a.SourceShardID, MaxShardIDLength); err != nil {
		return err
	}
	if err := validateToken("zone transfer destination shard id", a.DestinationShardID, MaxShardIDLength); err != nil {
		return err
	}
	if a.MaxDeliveryWindow == 0 {
		return errors.New("zone transfer max delivery window must be positive")
	}
	if a.CommittedRouteHash != "" {
		if err := zonestypes.ValidateHash("zone transfer committed route hash", a.CommittedRouteHash); err != nil {
			return err
		}
	}
	return nil
}

func (a ZoneTransferAdmission) DestinationZoneEnabled(zoneID zonestypes.ZoneID) bool {
	for _, enabled := range a.EnabledZones {
		if enabled == zoneID {
			return true
		}
	}
	return false
}

func (a ZoneTransferAdmission) DenomRoutable(sourceZoneID zonestypes.ZoneID, destinationZoneID zonestypes.ZoneID, denom string) bool {
	denom = strings.TrimSpace(denom)
	for _, routable := range a.RoutableDenoms {
		routable = routable.Normalize()
		if routable.SourceZoneID == sourceZoneID && routable.DestinationZoneID == destinationZoneID && routable.Denom == denom {
			return true
		}
	}
	return false
}

func (a ZoneTransferAdmission) MinimumFee(sourceZoneID zonestypes.ZoneID, destinationZoneID zonestypes.ZoneID, denom string) sdkmath.Int {
	denom = strings.TrimSpace(denom)
	for _, fee := range a.MinimumRouteFees {
		fee = fee.Normalize()
		if fee.SourceZoneID == sourceZoneID && fee.DestinationZoneID == destinationZoneID && fee.Denom == denom {
			return fee.MinimumFee
		}
	}
	return sdkmath.ZeroInt()
}

func (d RoutableDenom) Normalize() RoutableDenom {
	d.Denom = strings.TrimSpace(d.Denom)
	d.AuthorityPath = strings.TrimSpace(d.AuthorityPath)
	return d
}

func (d RoutableDenom) Validate() error {
	d = d.Normalize()
	if err := validateDenom("zone transfer routable denom", d.Denom); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(d.SourceZoneID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(d.DestinationZoneID); err != nil {
		return err
	}
	if d.SourceZoneID == d.DestinationZoneID {
		return errors.New("zone transfer routable denom requires distinct zones")
	}
	if err := validateToken("zone transfer authority path", d.AuthorityPath, MaxRouteIDLength); err != nil {
		return err
	}
	return nil
}

func (f ZoneTransferRouteFee) Normalize() ZoneTransferRouteFee {
	f.Denom = strings.TrimSpace(f.Denom)
	if f.MinimumFee.IsNil() {
		f.MinimumFee = sdkmath.ZeroInt()
	}
	return f
}

func (f ZoneTransferRouteFee) Validate() error {
	f = f.Normalize()
	if err := zonestypes.ValidateZoneID(f.SourceZoneID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(f.DestinationZoneID); err != nil {
		return err
	}
	if f.SourceZoneID == f.DestinationZoneID {
		return errors.New("zone transfer route fee requires distinct zones")
	}
	if err := validateDenom("zone transfer route fee denom", f.Denom); err != nil {
		return err
	}
	if f.MinimumFee.IsNegative() {
		return errors.New("zone transfer route fee must be non-negative")
	}
	return nil
}

func ZoneTransferRouteID(req MsgZoneTransfer) string {
	req = req.Normalize()
	routeHash := hashParts(
		"aetra-msg-zone-transfer-route-id-v1",
		string(req.SourceZoneID),
		string(req.DestinationZoneID),
		req.Denom,
		hex.EncodeToString(req.FromAddress),
		hex.EncodeToString(req.ToAddress),
	)
	return "zone-transfer/" + routeHash
}

func ComputeZoneTransferRouteCommitment(req MsgZoneTransfer, admission ZoneTransferAdmission) string {
	req = req.Normalize()
	admission = admission.Normalize()
	return hashParts(
		"aetra-msg-zone-transfer-route-v1",
		ZoneTransferRouteID(req),
		string(req.SourceZoneID),
		admission.SourceShardID,
		string(req.DestinationZoneID),
		admission.DestinationShardID,
		req.Denom,
		req.ForwardingFee.String(),
		fmt.Sprint(admission.CreatedHeight),
		fmt.Sprint(admission.MaxDeliveryWindow),
	)
}

func ComputeZoneTransferStateWriteSummary(req MsgZoneTransfer, admission ZoneTransferAdmission) string {
	req = req.Normalize()
	admission = admission.Normalize()
	return hashParts(
		"aetra-msg-zone-transfer-state-write-summary-v1",
		hex.EncodeToString(req.FromAddress),
		hex.EncodeToString(req.ToAddress),
		string(req.SourceZoneID),
		string(req.DestinationZoneID),
		req.Amount.String(),
		req.Denom,
		req.ForwardingFee.String(),
		fmt.Sprint(admission.SourceEscrowed),
		fmt.Sprint(admission.DestinationCredited),
	)
}

func validateDenom(fieldName string, denom string) error {
	denom = strings.TrimSpace(denom)
	if denom == "" || len(denom) > MaxDenomLength {
		return fmt.Errorf("%s must be 1..%d bytes", fieldName, MaxDenomLength)
	}
	for _, r := range denom {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' ||
			r == '_' || r == '-' || r == '.' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func normalizeZoneIDs(values []zonestypes.ZoneID) []zonestypes.ZoneID {
	out := append([]zonestypes.ZoneID(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

func normalizeRoutableDenoms(values []RoutableDenom) []RoutableDenom {
	out := make([]RoutableDenom, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SourceZoneID != out[j].SourceZoneID {
			return out[i].SourceZoneID < out[j].SourceZoneID
		}
		if out[i].DestinationZoneID != out[j].DestinationZoneID {
			return out[i].DestinationZoneID < out[j].DestinationZoneID
		}
		return out[i].Denom < out[j].Denom
	})
	return out
}

func normalizeZoneTransferRouteFees(values []ZoneTransferRouteFee) []ZoneTransferRouteFee {
	out := make([]ZoneTransferRouteFee, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SourceZoneID != out[j].SourceZoneID {
			return out[i].SourceZoneID < out[j].SourceZoneID
		}
		if out[i].DestinationZoneID != out[j].DestinationZoneID {
			return out[i].DestinationZoneID < out[j].DestinationZoneID
		}
		if out[i].Denom != out[j].Denom {
			return out[i].Denom < out[j].Denom
		}
		return out[i].MinimumFee.String() < out[j].MinimumFee.String()
	})
	return out
}
