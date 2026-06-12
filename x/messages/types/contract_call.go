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
	ContractCallPayloadType	= "contract.call"
	ContractCallAuthScope	= "contract-zone-call"
	MaxContractMethodLength	= 96
)

type MsgContractCall struct {
	Caller		sdk.AccAddress
	ContractAddr	sdk.AccAddress
	Method		string
	Args		[]byte
	Funds		sdkmath.Int
	GasLimit	uint64
	ReplyToOptional	sdk.AccAddress
	ExpiryHeight	uint64
}

type ContractCallAdmission struct {
	CreatedHeight		uint64
	Nonce			uint64
	SourceSequence		uint64
	ContractExists		bool
	ContractEnabled		bool
	EnabledMethods		[]ContractCallMethod
	FundEscrows		[]ContractCallFundsEscrow
	MaxArgsBytes		uint64
	MinGasLimit		uint64
	MaxGasLimit		uint64
	ContractShardID		string
	ReplyShardID		string
	CommittedRouteHash	string
}

type ContractCallMethod struct {
	ContractAddr	sdk.AccAddress
	Method		string
	SelectorHash	string
	Enabled		bool
}

type ContractCallFundsEscrow struct {
	Caller		sdk.AccAddress
	ContractAddr	sdk.AccAddress
	Amount		sdkmath.Int
	ExpiryHeight	uint64
	Escrowed	bool
}

func (k MessageKeeper) SubmitContractCall(req MsgContractCall, admission ContractCallAdmission) (MessageKeeper, SubmitCrossZoneMessageResponse, error) {
	msg, err := NewMessageFromContractCall(req, admission, k.state.Params)
	if err != nil {
		return MessageKeeper{}, SubmitCrossZoneMessageResponse{}, err
	}
	return k.SubmitCrossZoneMessage(MsgSubmitCrossZoneMessage{Message: msgWithoutIDs(msg)})
}

func (s KeeperMsgServer) SubmitContractCall(req MsgContractCall, admission ContractCallAdmission) (SubmitCrossZoneMessageResponse, error) {
	if s.keeper == nil {
		return SubmitCrossZoneMessageResponse{}, errors.New("message keeper is required")
	}
	next, resp, err := s.keeper.SubmitContractCall(req, admission)
	if err != nil {
		return SubmitCrossZoneMessageResponse{}, err
	}
	*s.keeper = next
	return resp, nil
}

func NewMessageFromContractCall(req MsgContractCall, admission ContractCallAdmission, params MessageParams) (Message, error) {
	req = req.Normalize()
	admission = admission.Normalize()
	if err := req.Validate(admission, params); err != nil {
		return Message{}, err
	}
	return NewMessage(Message{
		SourceZone:		zonestypes.ZoneIDContract,
		DestinationZone:	zonestypes.ZoneIDContract,
		Sender:			req.Caller,
		Recipient:		req.ContractAddr,
		Value:			req.Funds,
		Opcode:			ContractCallPayloadType,
		Payload:		req.CanonicalPayload(),
		GasLimit:		req.GasLimit,
		Deadline:		req.ExpiryHeight,
		Nonce:			admission.Nonce,
		SourceSequence:		admission.SourceSequence,
		RouteID:		ContractCallRouteID(req),
		Bounce:			len(req.ReplyToOptional) > 0,
		FeeLimit:		params.MinFeeLimit,
		CreatedHeight:		admission.CreatedHeight,
		AuthScope:		ContractCallAuthScope,
	}, params)
}

func NewAetherMessageFromContractCall(req MsgContractCall, admission ContractCallAdmission, params MessageParams) (AetherMessage, error) {
	msg, err := NewMessageFromContractCall(req, admission, params)
	if err != nil {
		return AetherMessage{}, err
	}
	routeCommitment := admission.CommittedRouteHash
	if routeCommitment == "" {
		routeCommitment = ComputeContractCallRouteCommitment(req, admission)
	}
	return NewAetherMessage(AetherMessage{
		Sender:			hex.EncodeToString(msg.Sender),
		SenderZoneID:		msg.SourceZone,
		SenderShardID:		admission.ContractShardID,
		Receiver:		hex.EncodeToString(msg.Recipient),
		ReceiverZoneID:		msg.DestinationZone,
		ReceiverShardID:	admission.ContractShardID,
		ValueNAET:		msg.Value,
		Payload:		msg.Payload,
		PayloadType:		ContractCallPayloadType,
		GasLimit:		msg.GasLimit,
		GasPrice:		msg.FeeLimit,
		ForwardingFee:		msg.FeeLimit,
		ExpiryHeight:		msg.Deadline,
		Bounce:			msg.Bounce,
		ExecutionMode:		ExecutionModeAsync,
		OrderingClass:		OrderingClassObjectOrdered,
		RouteCommitment:	routeCommitment,
		CreatedAtHeight:	msg.CreatedHeight,
		Nonce:			msg.Nonce,
	})
}

func (m MsgContractCall) Normalize() MsgContractCall {
	m.Caller = append(sdk.AccAddress(nil), m.Caller...)
	m.ContractAddr = append(sdk.AccAddress(nil), m.ContractAddr...)
	m.Method = strings.TrimSpace(m.Method)
	m.Args = append([]byte(nil), m.Args...)
	if m.Funds.IsNil() {
		m.Funds = sdkmath.ZeroInt()
	}
	m.ReplyToOptional = append(sdk.AccAddress(nil), m.ReplyToOptional...)
	return m
}

func (m MsgContractCall) Validate(admission ContractCallAdmission, params MessageParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	m = m.Normalize()
	admission = admission.Normalize()
	if err := admission.Validate(); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("contract call caller", m.Caller); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("contract call address", m.ContractAddr); err != nil {
		return err
	}
	if !admission.ContractExists {
		return errors.New("contract call target does not exist")
	}
	if !admission.ContractEnabled {
		return errors.New("contract call target is disabled")
	}
	method, found := admission.MethodFor(m.ContractAddr, m.Method)
	if !found || !method.Enabled {
		return errors.New("contract call method is not enabled")
	}
	if method.SelectorHash != ComputeContractMethodSelectorHash(m.Method) {
		return errors.New("contract call method selector is invalid")
	}
	maxArgs := admission.MaxArgsBytes
	if maxArgs == 0 {
		maxArgs = uint64(params.MaxPayloadSize)
	}
	if uint64(len(m.Args)) > maxArgs {
		return fmt.Errorf("contract call args size must be <= %d", maxArgs)
	}
	if m.Funds.IsNegative() {
		return errors.New("contract call funds must be non-negative")
	}
	if !admission.HasEscrow(m) {
		return errors.New("contract call funds must be escrowed")
	}
	minGas := admission.MinGasLimit
	if minGas == 0 {
		minGas = params.MinGasLimit
	}
	maxGas := admission.MaxGasLimit
	if maxGas == 0 {
		maxGas = params.MaxGasLimit
	}
	if m.GasLimit < minGas || m.GasLimit > maxGas {
		return fmt.Errorf("contract call gas limit must be in %d..%d", minGas, maxGas)
	}
	if m.ExpiryHeight == 0 || m.ExpiryHeight < admission.CreatedHeight {
		return errors.New("contract call expiry height must be at or after created height")
	}
	if len(m.ReplyToOptional) != 0 {
		if err := addressing.RejectZeroAddress("contract call reply target", m.ReplyToOptional); err != nil {
			return err
		}
	}
	return nil
}

func (m MsgContractCall) CanonicalPayload() []byte {
	m = m.Normalize()
	parts := []string{
		"aetra-msg-contract-call-payload-v1",
		hex.EncodeToString(m.Caller),
		hex.EncodeToString(m.ContractAddr),
		m.Method,
		ComputeAetherPayloadHash(m.Args),
		m.Funds.String(),
		fmt.Sprint(m.GasLimit),
		hex.EncodeToString(m.ReplyToOptional),
		fmt.Sprint(m.ExpiryHeight),
	}
	return []byte(strings.Join(parts, "|"))
}

func (a ContractCallAdmission) Normalize() ContractCallAdmission {
	a.EnabledMethods = normalizeContractCallMethods(a.EnabledMethods)
	a.FundEscrows = normalizeContractCallFundsEscrows(a.FundEscrows)
	a.ContractShardID = strings.TrimSpace(a.ContractShardID)
	a.ReplyShardID = strings.TrimSpace(a.ReplyShardID)
	a.CommittedRouteHash = strings.ToLower(strings.TrimSpace(a.CommittedRouteHash))
	return a
}

func (a ContractCallAdmission) Validate() error {
	a = a.Normalize()
	if a.CreatedHeight == 0 {
		return errors.New("contract call created height must be positive")
	}
	if a.Nonce == 0 || a.SourceSequence == 0 {
		return errors.New("contract call nonce and source sequence must be positive")
	}
	for _, method := range a.EnabledMethods {
		if err := method.Validate(); err != nil {
			return err
		}
	}
	for _, escrow := range a.FundEscrows {
		if err := escrow.Validate(); err != nil {
			return err
		}
	}
	if err := validateToken("contract call shard id", a.ContractShardID, MaxShardIDLength); err != nil {
		return err
	}
	if a.ReplyShardID != "" {
		if err := validateToken("contract call reply shard id", a.ReplyShardID, MaxShardIDLength); err != nil {
			return err
		}
	}
	if a.MaxArgsBytes > uint64(MaxPayloadBytesForParamsSafety()) {
		return errors.New("contract call max args bytes exceeds safety bound")
	}
	if a.MinGasLimit != 0 && a.MaxGasLimit != 0 && a.MinGasLimit > a.MaxGasLimit {
		return errors.New("contract call gas bounds are invalid")
	}
	if a.CommittedRouteHash != "" {
		if err := zonestypes.ValidateHash("contract call committed route hash", a.CommittedRouteHash); err != nil {
			return err
		}
	}
	return nil
}

func (a ContractCallAdmission) MethodFor(contractAddr sdk.AccAddress, method string) (ContractCallMethod, bool) {
	method = strings.TrimSpace(method)
	for _, item := range a.EnabledMethods {
		item = item.Normalize()
		if bytes.Equal(item.ContractAddr, contractAddr) && item.Method == method {
			return item, true
		}
	}
	return ContractCallMethod{}, false
}

func (a ContractCallAdmission) HasEscrow(msg MsgContractCall) bool {
	msg = msg.Normalize()
	for _, escrow := range a.FundEscrows {
		escrow = escrow.Normalize()
		if !escrow.Escrowed {
			continue
		}
		if !bytes.Equal(escrow.Caller, msg.Caller) || !bytes.Equal(escrow.ContractAddr, msg.ContractAddr) {
			continue
		}
		if escrow.Amount.LT(msg.Funds) {
			continue
		}
		if escrow.ExpiryHeight < msg.ExpiryHeight {
			continue
		}
		return true
	}
	return false
}

func (m ContractCallMethod) Normalize() ContractCallMethod {
	m.ContractAddr = append(sdk.AccAddress(nil), m.ContractAddr...)
	m.Method = strings.TrimSpace(m.Method)
	m.SelectorHash = strings.ToLower(strings.TrimSpace(m.SelectorHash))
	return m
}

func (m ContractCallMethod) Validate() error {
	m = m.Normalize()
	if err := addressing.RejectZeroAddress("contract call method contract", m.ContractAddr); err != nil {
		return err
	}
	if err := validateToken("contract call method", m.Method, MaxContractMethodLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("contract call method selector hash", m.SelectorHash); err != nil {
		return err
	}
	if m.SelectorHash != ComputeContractMethodSelectorHash(m.Method) {
		return errors.New("contract call method selector hash mismatch")
	}
	return nil
}

func (e ContractCallFundsEscrow) Normalize() ContractCallFundsEscrow {
	e.Caller = append(sdk.AccAddress(nil), e.Caller...)
	e.ContractAddr = append(sdk.AccAddress(nil), e.ContractAddr...)
	if e.Amount.IsNil() {
		e.Amount = sdkmath.ZeroInt()
	}
	return e
}

func (e ContractCallFundsEscrow) Validate() error {
	e = e.Normalize()
	if err := addressing.RejectZeroAddress("contract call escrow caller", e.Caller); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("contract call escrow contract", e.ContractAddr); err != nil {
		return err
	}
	if e.Amount.IsNegative() {
		return errors.New("contract call escrow amount must be non-negative")
	}
	if e.ExpiryHeight == 0 {
		return errors.New("contract call escrow expiry height must be positive")
	}
	return nil
}

func ContractCallRouteID(req MsgContractCall) string {
	req = req.Normalize()
	return "contract-call/" + hashParts(
		"aetra-msg-contract-call-route-id-v1",
		hex.EncodeToString(req.Caller),
		hex.EncodeToString(req.ContractAddr),
		req.Method,
	)
}

func ComputeContractCallRouteCommitment(req MsgContractCall, admission ContractCallAdmission) string {
	req = req.Normalize()
	admission = admission.Normalize()
	return hashParts(
		"aetra-msg-contract-call-route-v1",
		ContractCallRouteID(req),
		admission.ContractShardID,
		admission.ReplyShardID,
		req.Method,
		ComputeContractMethodSelectorHash(req.Method),
		fmt.Sprint(admission.CreatedHeight),
	)
}

func ComputeContractMethodSelectorHash(method string) string {
	return hashParts("aetra-msg-contract-method-selector-v1", strings.TrimSpace(method))
}

func MaxPayloadBytesForParamsSafety() uint32 {
	return 16 * 1024 * 1024
}

func normalizeContractCallMethods(values []ContractCallMethod) []ContractCallMethod {
	out := make([]ContractCallMethod, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if cmp := bytes.Compare(out[i].ContractAddr, out[j].ContractAddr); cmp != 0 {
			return cmp < 0
		}
		return out[i].Method < out[j].Method
	})
	return out
}

func normalizeContractCallFundsEscrows(values []ContractCallFundsEscrow) []ContractCallFundsEscrow {
	out := make([]ContractCallFundsEscrow, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if cmp := bytes.Compare(out[i].Caller, out[j].Caller); cmp != 0 {
			return cmp < 0
		}
		return bytes.Compare(out[i].ContractAddr, out[j].ContractAddr) < 0
	})
	return out
}
