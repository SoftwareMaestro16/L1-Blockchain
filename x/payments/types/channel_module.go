package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type PaymentChannelMessageType string

const (
	PaymentChannelMsgOpenChannel			PaymentChannelMessageType	= "MsgOpenChannel"
	PaymentChannelMsgCooperativeClose		PaymentChannelMessageType	= "MsgCooperativeClose"
	PaymentChannelMsgUnilateralClose		PaymentChannelMessageType	= "MsgUnilateralClose"
	PaymentChannelMsgDisputeClose			PaymentChannelMessageType	= "MsgDisputeClose"
	PaymentChannelMsgFinalizeClose			PaymentChannelMessageType	= "MsgFinalizeClose"
	PaymentChannelMsgSubmitCheckpoint		PaymentChannelMessageType	= "MsgSubmitCheckpoint"
	PaymentChannelMsgCancelExpiredChannel		PaymentChannelMessageType	= "MsgCancelExpiredChannel"
	PaymentChannelMsgRegisterChannelAdvertisement	PaymentChannelMessageType	= "MsgRegisterChannelAdvertisement"
)

type ChannelParticipant struct {
	ChannelID	string
	Address		string
	Signer		bool
	Balance		string
	Reserve		string
}

type ChannelConfig struct {
	ChannelID		string
	ChannelType		ChannelType
	CloseDelay		uint64
	ChallengePeriod		uint64
	FeePolicyID		string
	RoutingAdvertised	bool
	ConditionalPayments	bool
}

type ChannelFeeAccumulator struct {
	BlockHeight	uint64
	Bucket		string
	Denom		string
	FeeAmount	string
	PenaltyAmount	string
	OperationCount	uint64
	AccumulatorKey	string
}

type PaymentChannelModuleState struct {
	Channels		[]ChannelRecord
	Participants		[]ChannelParticipant
	Configs			[]ChannelConfig
	PendingCloses		[]PendingClose
	Settlements		[]SettlementRecord
	SettlementTombstones	[]ClosedChannelTombstone
	FeeAccumulators		[]ChannelFeeAccumulator
}

type PaymentChannelAnteFee struct {
	MsgType		PaymentChannelMessageType
	FeeClass	PaymentFeeClass
	ChannelID	string
	Payer		string
	Paid		string
	Required	string
	StorageBytes	uint64
	MultiplierBps	uint32
}

type PaymentChannelMessageResult struct {
	MsgType		PaymentChannelMessageType
	ChannelID	string
	Settlement	SettlementRecord
	Checkpoint	ChannelUpdateResult
	Advertisement	LiquidityAdvertisement
	AnteFee		PaymentChannelAnteFee
}

type PaymentChannelModuleMessage interface {
	PaymentChannelType() PaymentChannelMessageType
	ValidateBasic() error
	Signers() []string
}

type MsgOpenChannel struct {
	Signer	string
	Request	ChannelOpenRequest
}

type MsgCooperativeClose struct {
	Signer	string
	Request	ChannelCloseRequest
}

type MsgUnilateralClose struct {
	Signer	string
	Request	ChannelCloseRequest
}

type MsgDisputeClose struct {
	Signer	string
	Request	ChannelDisputeRequest
}

type MsgFinalizeClose struct {
	Signer	string
	Request	FinalSettlementRequest
}

type MsgSubmitCheckpoint struct {
	Signer	string
	Request	ChannelUpdateRequest
}

type MsgCancelExpiredChannel struct {
	Signer		string
	ChannelID	string
	CurrentHeight	uint64
	SettlementFee	string
}

type MsgRegisterChannelAdvertisement struct {
	Signer		string
	Advertisement	LiquidityAdvertisement
	RequiredDeposit	string
	CurrentHeight	uint64
	FeePaid		string
}

func SnapshotPaymentChannelModuleState(state PaymentsState, blockHeight uint64) (PaymentChannelModuleState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentChannelModuleState{}, err
	}
	out := PaymentChannelModuleState{
		Channels:		append([]ChannelRecord(nil), state.Channels...),
		Settlements:		append([]SettlementRecord(nil), state.Settlements...),
		SettlementTombstones:	append([]ClosedChannelTombstone(nil), state.ClosedChannels...),
	}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		out.Configs = append(out.Configs, ChannelConfigForChannel(channel))
		for _, participant := range channel.Participants {
			balance, reserve := channelParticipantAmounts(channel.LatestState, participant)
			out.Participants = append(out.Participants, ChannelParticipant{
				ChannelID:	channel.ChannelID,
				Address:	participant,
				Signer:		containsString(channel.RequiredSigners, participant),
				Balance:	balance,
				Reserve:	reserve,
			}.Normalize())
		}
		if channel.Status == ChannelStatusPendingClose {
			out.PendingCloses = append(out.PendingCloses, channel.PendingClose.Normalize())
		}
	}
	if blockHeight > 0 {
		acc := ChannelFeeAccumulatorFromBlock(PaymentBlockAccumulator{BlockHeight: blockHeight}, "settlement")
		out.FeeAccumulators = append(out.FeeAccumulators, acc)
	}
	return out.Normalize(), out.Validate()
}

func ChannelConfigForChannel(channel ChannelRecord) ChannelConfig {
	channel = channel.Normalize()
	return ChannelConfig{
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		CloseDelay:		channel.CloseDelay,
		ChallengePeriod:	channel.DisputePeriod,
		FeePolicyID:		channel.LatestState.FeePolicyID,
		RoutingAdvertised:	channel.RoutingAdvertised,
		ConditionalPayments:	channel.ConditionalPayments,
	}.Normalize()
}

func ChannelFeeAccumulatorFromBlock(acc PaymentBlockAccumulator, bucket string) ChannelFeeAccumulator {
	acc = acc.Normalize()
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		bucket = "settlement"
	}
	return ChannelFeeAccumulator{
		BlockHeight:	acc.BlockHeight,
		Bucket:		bucket,
		Denom:		NativeDenom,
		FeeAmount:	acc.FeeAmount,
		PenaltyAmount:	acc.PenaltyAmount,
		OperationCount:	acc.OperationCount,
		AccumulatorKey:	StoreV2FeeAccumulatorKey(fmt.Sprintf("%020d", acc.BlockHeight), bucket),
	}.Normalize()
}

func ApplyPaymentChannelMessage(state PaymentsState, msg PaymentChannelModuleMessage) (PaymentsState, PaymentChannelMessageResult, error) {
	state = state.Export()
	if msg == nil {
		return PaymentsState{}, PaymentChannelMessageResult{}, errors.New("payments channel module message is required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return PaymentsState{}, PaymentChannelMessageResult{}, err
	}
	ante, err := ValidatePaymentChannelMessageFee(state, msg)
	if err != nil {
		return PaymentsState{}, PaymentChannelMessageResult{}, err
	}
	result := PaymentChannelMessageResult{MsgType: msg.PaymentChannelType(), AnteFee: ante}
	switch m := msg.(type) {
	case MsgOpenChannel:
		next, _, err := OpenChannelFromRequest(state, m.Normalize().Request)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		result.ChannelID = m.Normalize().Request.ChannelID
		return next, result, nil
	case MsgCooperativeClose:
		req := m.Normalize().Request
		settlement, err := validateCloseMessageAndCooperativeState(state, req, m.Signer)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		next, closed, err := CooperativeClose(state, req.ChannelID, req.ClosingStateWithSignatures(), m.Normalize().Signer, req.CurrentHeight, req.SettlementFee)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		_ = settlement
		result.ChannelID = req.ChannelID
		result.Settlement = closed
		return next, result, nil
	case MsgUnilateralClose:
		req := m.Normalize().Request
		next, err := SubmitCloseWithRequest(state, req)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		result.ChannelID = req.ChannelID
		return next, result, nil
	case MsgDisputeClose:
		req := m.Normalize().Request
		next, err := DisputeChannel(state, req)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		result.ChannelID = req.ChannelID
		return next, result, nil
	case MsgFinalizeClose:
		req := m.Normalize().Request
		next, settlement, err := FinalizeSettlementWithRequest(state, req)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		result.ChannelID = req.ChannelID
		result.Settlement = settlement
		return next, result, nil
	case MsgSubmitCheckpoint:
		req := m.Normalize().Request
		next, checkpoint, err := RegisterUpdateCheckpoint(state, req)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		result.ChannelID = req.ChannelID
		result.Checkpoint = checkpoint
		return next, result, nil
	case MsgCancelExpiredChannel:
		msg := m.Normalize()
		next, err := ForcedClose(state, msg.ChannelID, msg.Signer, msg.CurrentHeight, msg.SettlementFee)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		result.ChannelID = msg.ChannelID
		return next, result, nil
	case MsgRegisterChannelAdvertisement:
		msg := m.Normalize()
		next, ad, err := RegisterChannelAdvertisement(state, msg)
		if err != nil {
			return PaymentsState{}, PaymentChannelMessageResult{}, err
		}
		result.ChannelID = ad.ChannelID
		result.Advertisement = ad
		return next, result, nil
	default:
		return PaymentsState{}, PaymentChannelMessageResult{}, errors.New("payments channel module message type is unsupported")
	}
}

func ValidatePaymentChannelMessageFee(state PaymentsState, msg PaymentChannelModuleMessage) (PaymentChannelAnteFee, error) {
	state = state.Export()
	if msg == nil {
		return PaymentChannelAnteFee{}, errors.New("payments ante message is required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return PaymentChannelAnteFee{}, err
	}
	feeClass, channel, payer, paid, objectID, height, err := channelMessageFeeContext(state, msg)
	if err != nil {
		return PaymentChannelAnteFee{}, err
	}
	required, storageBytes, multiplier, err := RequiredPaymentFee(state, feeClass, channel)
	if err != nil {
		return PaymentChannelAnteFee{}, err
	}
	if err := requirePaidAtLeast("payments ante fee paid", paid, required); err != nil {
		return PaymentChannelAnteFee{}, err
	}
	_ = objectID
	_ = height
	ante := PaymentChannelAnteFee{
		MsgType:	msg.PaymentChannelType(),
		FeeClass:	feeClass,
		ChannelID:	channel.ChannelID,
		Payer:		payer,
		Paid:		paid,
		Required:	required,
		StorageBytes:	storageBytes,
		MultiplierBps:	multiplier,
	}.Normalize()
	if err := ante.Validate(); err != nil {
		return PaymentChannelAnteFee{}, err
	}
	return ante, nil
}

func RegisterChannelAdvertisement(state PaymentsState, msg MsgRegisterChannelAdvertisement) (PaymentsState, LiquidityAdvertisement, error) {
	state = state.Export()
	msg = msg.Normalize()
	if err := msg.ValidateBasic(); err != nil {
		return PaymentsState{}, LiquidityAdvertisement{}, err
	}
	channel, found := state.ChannelByID(msg.Advertisement.ChannelID)
	if !found {
		return PaymentsState{}, LiquidityAdvertisement{}, errors.New("payments advertisement channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, LiquidityAdvertisement{}, errors.New("payments advertisement requires open channel")
	}
	if !containsString(channel.Participants, msg.Signer) {
		return PaymentsState{}, LiquidityAdvertisement{}, errors.New("payments advertisement signer must be channel participant")
	}
	requiredDeposit := msg.RequiredDeposit
	if requiredDeposit == "" {
		requiredDeposit = state.FeeSchedule.Normalize().RoutingAdvertisementDeposit
	}
	ad, err := BuildLiquidityAdvertisement(msg.Advertisement, requiredDeposit)
	if err != nil {
		return PaymentsState{}, LiquidityAdvertisement{}, err
	}
	if ad.Advertiser != msg.Signer {
		return PaymentsState{}, LiquidityAdvertisement{}, errors.New("payments advertisement signer mismatch")
	}
	charged, _, err := ChargePaymentFee(state, PaymentFeeClassRoutingAdvertisement, channel, msg.Signer, ad.AdvertisementID, msg.FeePaid, msg.CurrentHeight)
	if err != nil {
		return PaymentsState{}, LiquidityAdvertisement{}, err
	}
	next, err := RegisterRoutingEdge(charged, ChannelEdge{
		ChannelID:	ad.ChannelID,
		From:		ad.Advertiser,
		To:		ad.Counterparty,
		Capacity:	ad.Capacity,
		FeeDenom:	ad.FeeDenom,
		FeeAmount:	ad.BaseFee,
		ExpiresHeight:	ad.ValidUntilHeight,
		Active:		true,
	})
	if err != nil {
		return PaymentsState{}, LiquidityAdvertisement{}, err
	}
	return next, ad, nil
}

func PaymentChannelMessageAccessPlan(msg PaymentChannelModuleMessage, blockHeight uint64) (BlockSTMAccessPlan, error) {
	if msg == nil {
		return BlockSTMAccessPlan{}, errors.New("payments blockstm message is required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return BlockSTMAccessPlan{}, err
	}
	op, err := settlementOperationForMessage(msg)
	if err != nil {
		return BlockSTMAccessPlan{}, err
	}
	return AccessPlanForSettlementOperation(op, blockHeight)
}

func PaymentChannelMessagesConflictProfile(messages []PaymentChannelModuleMessage, blockHeight uint64) (BlockSTMConflictProfile, error) {
	plans := make([]BlockSTMAccessPlan, 0, len(messages))
	for _, msg := range messages {
		plan, err := PaymentChannelMessageAccessPlan(msg, blockHeight)
		if err != nil {
			return BlockSTMConflictProfile{}, err
		}
		plans = append(plans, plan)
	}
	return ProfileBlockSTMConflicts(plans), nil
}

func (m MsgOpenChannel) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgOpenChannel
}
func (m MsgOpenChannel) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgOpenChannel) Normalize() MsgOpenChannel {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Request = m.Request.Normalize()
	return m
}
func (m MsgOpenChannel) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg open signer", msg.Signer); err != nil {
		return err
	}
	if !containsString(msg.Request.Participants, msg.Signer) {
		return errors.New("payments msg open signer must be participant")
	}
	return msg.Request.Validate()
}

func (m MsgCooperativeClose) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgCooperativeClose
}
func (m MsgCooperativeClose) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgCooperativeClose) Normalize() MsgCooperativeClose {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Request = m.Request.Normalize()
	m.Request.CloseReason = CloseReasonCooperative
	m.Request.Submitter = m.Signer
	if strings.TrimSpace(m.Request.SettlementFee) == "" {
		m.Request.SettlementFee = "0"
	}
	return m
}
func (m MsgCooperativeClose) ValidateBasic() error {
	msg := m.Normalize()
	if err := validateCloseMessageBasic("payments msg cooperative close", msg.Signer, msg.Request); err != nil {
		return err
	}
	if msg.Request.CloseReason != CloseReasonCooperative {
		return errors.New("payments msg cooperative close reason mismatch")
	}
	return nil
}

func (m MsgUnilateralClose) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgUnilateralClose
}
func (m MsgUnilateralClose) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgUnilateralClose) Normalize() MsgUnilateralClose {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Request = m.Request.Normalize()
	m.Request.CloseReason = CloseReasonUnilateral
	m.Request.Submitter = m.Signer
	if strings.TrimSpace(m.Request.SettlementFee) == "" {
		m.Request.SettlementFee = "0"
	}
	return m
}
func (m MsgUnilateralClose) ValidateBasic() error {
	msg := m.Normalize()
	if err := validateCloseMessageBasic("payments msg unilateral close", msg.Signer, msg.Request); err != nil {
		return err
	}
	if msg.Request.CloseReason != CloseReasonUnilateral {
		return errors.New("payments msg unilateral close reason mismatch")
	}
	return nil
}

func (m MsgDisputeClose) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgDisputeClose
}
func (m MsgDisputeClose) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgDisputeClose) Normalize() MsgDisputeClose {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Request = m.Request.Normalize()
	m.Request.Submitter = m.Signer
	return m
}
func (m MsgDisputeClose) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg dispute signer", msg.Signer); err != nil {
		return err
	}
	if err := ValidateHash("payments msg dispute channel id", msg.Request.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments msg dispute close reference", msg.Request.ClosingStateReference); err != nil {
		return err
	}
	if msg.Request.NewerState.StateHash == "" {
		return errors.New("payments msg dispute newer state is required")
	}
	if msg.Request.CurrentHeight == 0 {
		return errors.New("payments msg dispute height must be positive")
	}
	return validateNonNegativeInt("payments msg dispute fee", msg.Request.DisputeFeePaid)
}

func (m MsgFinalizeClose) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgFinalizeClose
}
func (m MsgFinalizeClose) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgFinalizeClose) Normalize() MsgFinalizeClose {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Request = m.Request.Normalize()
	return m
}
func (m MsgFinalizeClose) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg finalize signer", msg.Signer); err != nil {
		return err
	}
	if err := ValidateHash("payments msg finalize channel id", msg.Request.ChannelID); err != nil {
		return err
	}
	if msg.Request.CurrentHeight == 0 {
		return errors.New("payments msg finalize height must be positive")
	}
	return nil
}

func (m MsgSubmitCheckpoint) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgSubmitCheckpoint
}
func (m MsgSubmitCheckpoint) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgSubmitCheckpoint) Normalize() MsgSubmitCheckpoint {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Request = m.Request.Normalize()
	m.Request.Submitter = m.Signer
	m.Request.RegisterCheckpoint = true
	return m
}
func (m MsgSubmitCheckpoint) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg checkpoint signer", msg.Signer); err != nil {
		return err
	}
	if err := ValidateHash("payments msg checkpoint channel id", msg.Request.ChannelID); err != nil {
		return err
	}
	if msg.Request.State.StateHash == "" {
		return errors.New("payments msg checkpoint state is required")
	}
	if msg.Request.CurrentHeight == 0 {
		return errors.New("payments msg checkpoint height must be positive")
	}
	return validateNonNegativeInt("payments msg checkpoint fee", msg.Request.CheckpointFeePaid)
}

func (m MsgCancelExpiredChannel) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgCancelExpiredChannel
}
func (m MsgCancelExpiredChannel) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgCancelExpiredChannel) Normalize() MsgCancelExpiredChannel {
	m.Signer = strings.TrimSpace(m.Signer)
	m.ChannelID = normalizeHash(m.ChannelID)
	m.SettlementFee = strings.TrimSpace(m.SettlementFee)
	if m.SettlementFee == "" {
		m.SettlementFee = "0"
	}
	return m
}
func (m MsgCancelExpiredChannel) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg cancel expired signer", msg.Signer); err != nil {
		return err
	}
	if err := ValidateHash("payments msg cancel expired channel id", msg.ChannelID); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg cancel expired height must be positive")
	}
	return validateNonNegativeInt("payments msg cancel expired fee", msg.SettlementFee)
}

func (m MsgRegisterChannelAdvertisement) PaymentChannelType() PaymentChannelMessageType {
	return PaymentChannelMsgRegisterChannelAdvertisement
}
func (m MsgRegisterChannelAdvertisement) Signers() []string	{ return []string{m.Normalize().Signer} }
func (m MsgRegisterChannelAdvertisement) Normalize() MsgRegisterChannelAdvertisement {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Advertisement = m.Advertisement.Normalize()
	m.RequiredDeposit = strings.TrimSpace(m.RequiredDeposit)
	if m.RequiredDeposit == "" {
		m.RequiredDeposit = "0"
	}
	m.FeePaid = strings.TrimSpace(m.FeePaid)
	if m.FeePaid == "" {
		m.FeePaid = "0"
	}
	return m
}
func (m MsgRegisterChannelAdvertisement) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg advertisement signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg advertisement height must be positive")
	}
	if err := validateNonNegativeInt("payments msg advertisement required deposit", msg.RequiredDeposit); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments msg advertisement fee", msg.FeePaid); err != nil {
		return err
	}
	if msg.Advertisement.Advertiser != msg.Signer {
		return errors.New("payments msg advertisement signer mismatch")
	}
	return msg.Advertisement.Validate(msg.RequiredDeposit)
}

func (s PaymentChannelModuleState) Normalize() PaymentChannelModuleState {
	s.Channels = normalizeChannelModuleChannels(s.Channels)
	s.Participants = normalizeChannelParticipants(s.Participants)
	s.Configs = normalizeChannelConfigs(s.Configs)
	for i := range s.PendingCloses {
		s.PendingCloses[i] = s.PendingCloses[i].Normalize()
	}
	sort.SliceStable(s.PendingCloses, func(i, j int) bool {
		return s.PendingCloses[i].State.ChannelID < s.PendingCloses[j].State.ChannelID
	})
	s.Settlements = normalizeChannelModuleSettlements(s.Settlements)
	s.SettlementTombstones = normalizeChannelModuleTombstones(s.SettlementTombstones)
	s.FeeAccumulators = normalizeChannelFeeAccumulators(s.FeeAccumulators)
	return s
}

func (s PaymentChannelModuleState) Validate() error {
	state := s.Normalize()
	for _, channel := range state.Channels {
		if err := channel.Validate(); err != nil {
			return err
		}
	}
	for _, participant := range state.Participants {
		if err := participant.Validate(); err != nil {
			return err
		}
	}
	for _, config := range state.Configs {
		if err := config.Validate(); err != nil {
			return err
		}
	}
	for _, settlement := range state.Settlements {
		channel, found := channelByIDInSlice(state.Channels, settlement.ChannelID)
		if !found {
			return errors.New("payments channel module settlement references unknown channel")
		}
		if err := settlement.ValidateForChannel(channel); err != nil {
			return err
		}
	}
	for _, tombstone := range state.SettlementTombstones {
		if err := tombstone.Validate(); err != nil {
			return err
		}
	}
	for _, acc := range state.FeeAccumulators {
		if err := acc.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p ChannelParticipant) Normalize() ChannelParticipant {
	p.ChannelID = normalizeHash(p.ChannelID)
	p.Address = strings.TrimSpace(p.Address)
	p.Balance = strings.TrimSpace(p.Balance)
	if p.Balance == "" {
		p.Balance = "0"
	}
	p.Reserve = strings.TrimSpace(p.Reserve)
	if p.Reserve == "" {
		p.Reserve = "0"
	}
	return p
}

func (p ChannelParticipant) Validate() error {
	participant := p.Normalize()
	if err := ValidateHash("payments channel participant channel id", participant.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments channel participant address", participant.Address); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments channel participant balance", participant.Balance); err != nil {
		return err
	}
	return validateNonNegativeInt("payments channel participant reserve", participant.Reserve)
}

func (c ChannelConfig) Normalize() ChannelConfig {
	c.ChannelID = normalizeHash(c.ChannelID)
	c.FeePolicyID = strings.TrimSpace(c.FeePolicyID)
	if c.FeePolicyID == "" {
		c.FeePolicyID = NativeDenom
	}
	return c
}

func (c ChannelConfig) Validate() error {
	config := c.Normalize()
	if err := ValidateHash("payments channel config channel id", config.ChannelID); err != nil {
		return err
	}
	if !IsChannelType(config.ChannelType) {
		return fmt.Errorf("unknown payments channel config type %q", config.ChannelType)
	}
	if err := validateCloseDelay(config.CloseDelay); err != nil {
		return err
	}
	if err := validateChallengePeriod(config.ChallengePeriod); err != nil {
		return err
	}
	if config.FeePolicyID != NativeDenom {
		return fmt.Errorf("payments channel config fee policy must be %s", NativeDenom)
	}
	return nil
}

func (a ChannelFeeAccumulator) Normalize() ChannelFeeAccumulator {
	a.Bucket = strings.TrimSpace(a.Bucket)
	if a.Bucket == "" {
		a.Bucket = "settlement"
	}
	a.Denom = normalizeAssetDenom(a.Denom)
	a.FeeAmount = strings.TrimSpace(a.FeeAmount)
	if a.FeeAmount == "" {
		a.FeeAmount = "0"
	}
	a.PenaltyAmount = strings.TrimSpace(a.PenaltyAmount)
	if a.PenaltyAmount == "" {
		a.PenaltyAmount = "0"
	}
	if a.AccumulatorKey == "" && a.BlockHeight > 0 {
		a.AccumulatorKey = StoreV2FeeAccumulatorKey(fmt.Sprintf("%020d", a.BlockHeight), a.Bucket)
	}
	a.AccumulatorKey = strings.TrimSpace(a.AccumulatorKey)
	return a
}

func (a ChannelFeeAccumulator) Validate() error {
	acc := a.Normalize()
	if acc.BlockHeight == 0 {
		return errors.New("payments channel fee accumulator height must be positive")
	}
	if acc.Bucket == "" {
		return errors.New("payments channel fee accumulator bucket is required")
	}
	if acc.Denom != NativeDenom {
		return fmt.Errorf("payments channel fee accumulator denom must be %s", NativeDenom)
	}
	if acc.AccumulatorKey != StoreV2FeeAccumulatorKey(fmt.Sprintf("%020d", acc.BlockHeight), acc.Bucket) {
		return errors.New("payments channel fee accumulator key mismatch")
	}
	if err := validateNonNegativeInt("payments channel fee accumulator fees", acc.FeeAmount); err != nil {
		return err
	}
	return validateNonNegativeInt("payments channel fee accumulator penalties", acc.PenaltyAmount)
}

func (f PaymentChannelAnteFee) Normalize() PaymentChannelAnteFee {
	f.ChannelID = normalizeOptionalHash(f.ChannelID)
	f.Payer = strings.TrimSpace(f.Payer)
	f.Paid = strings.TrimSpace(f.Paid)
	if f.Paid == "" {
		f.Paid = "0"
	}
	f.Required = strings.TrimSpace(f.Required)
	if f.Required == "" {
		f.Required = "0"
	}
	return f
}

func (f PaymentChannelAnteFee) Validate() error {
	fee := f.Normalize()
	if !IsPaymentChannelMessageType(fee.MsgType) {
		return fmt.Errorf("unknown payments channel ante msg type %q", fee.MsgType)
	}
	if !IsPaymentFeeClass(fee.FeeClass) {
		return fmt.Errorf("unknown payments channel ante fee class %q", fee.FeeClass)
	}
	if fee.ChannelID != "" {
		if err := ValidateHash("payments channel ante channel id", fee.ChannelID); err != nil {
			return err
		}
	}
	if err := addressing.ValidateUserAddress("payments channel ante payer", fee.Payer); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments channel ante paid", fee.Paid); err != nil {
		return err
	}
	return validateNonNegativeInt("payments channel ante required", fee.Required)
}

func IsPaymentChannelMessageType(value PaymentChannelMessageType) bool {
	switch value {
	case PaymentChannelMsgOpenChannel,
		PaymentChannelMsgCooperativeClose,
		PaymentChannelMsgUnilateralClose,
		PaymentChannelMsgDisputeClose,
		PaymentChannelMsgFinalizeClose,
		PaymentChannelMsgSubmitCheckpoint,
		PaymentChannelMsgCancelExpiredChannel,
		PaymentChannelMsgRegisterChannelAdvertisement:
		return true
	default:
		return false
	}
}

func channelMessageFeeContext(state PaymentsState, msg PaymentChannelModuleMessage) (PaymentFeeClass, ChannelRecord, string, string, string, uint64, error) {
	switch m := msg.(type) {
	case MsgOpenChannel:
		msg := m.Normalize()
		channel, err := BuildChannelFromOpenRequest(msg.Request)
		if err != nil {
			return "", ChannelRecord{}, "", "", "", 0, err
		}
		return PaymentFeeClassChannelOpen, channel, msg.Signer, msg.Request.OpeningFeePaid, channel.ChannelID, msg.Request.OpenHeight, nil
	case MsgCooperativeClose:
		msg := m.Normalize()
		channel, found := state.ChannelByID(msg.Request.ChannelID)
		if !found {
			return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante channel not found")
		}
		return PaymentFeeClassCooperativeClose, channel, msg.Signer, msg.Request.SettlementFee, msg.Request.ClosingStateWithSignatures().StateHash, msg.Request.CurrentHeight, nil
	case MsgUnilateralClose:
		msg := m.Normalize()
		channel, found := state.ChannelByID(msg.Request.ChannelID)
		if !found {
			return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante channel not found")
		}
		return PaymentFeeClassUnilateralClose, channel, msg.Signer, msg.Request.SettlementFee, msg.Request.ClosingStateWithSignatures().StateHash, msg.Request.CurrentHeight, nil
	case MsgDisputeClose:
		msg := m.Normalize()
		channel, found := state.ChannelByID(msg.Request.ChannelID)
		if !found {
			return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante channel not found")
		}
		return PaymentFeeClassDispute, channel, msg.Signer, msg.Request.DisputeFeePaid, msg.Request.NewerState.StateHash, msg.Request.CurrentHeight, nil
	case MsgFinalizeClose:
		msg := m.Normalize()
		channel, found := state.ChannelByID(msg.Request.ChannelID)
		if !found {
			return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante channel not found")
		}
		return PaymentFeeClassUnilateralClose, channel, msg.Signer, "0", msg.Request.ChannelID, msg.Request.CurrentHeight, nil
	case MsgSubmitCheckpoint:
		msg := m.Normalize()
		channel, found := state.ChannelByID(msg.Request.ChannelID)
		if !found {
			return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante channel not found")
		}
		return PaymentFeeClassChannelCheckpoint, channel, msg.Signer, msg.Request.CheckpointFeePaid, msg.Request.State.StateHash, msg.Request.CurrentHeight, nil
	case MsgCancelExpiredChannel:
		msg := m.Normalize()
		channel, found := state.ChannelByID(msg.ChannelID)
		if !found {
			return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante channel not found")
		}
		return PaymentFeeClassUnilateralClose, channel, msg.Signer, msg.SettlementFee, channel.LatestState.StateHash, msg.CurrentHeight, nil
	case MsgRegisterChannelAdvertisement:
		msg := m.Normalize()
		channel, found := state.ChannelByID(msg.Advertisement.ChannelID)
		if !found {
			return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante channel not found")
		}
		return PaymentFeeClassRoutingAdvertisement, channel, msg.Signer, msg.FeePaid, msg.Advertisement.AdvertisementID, msg.CurrentHeight, nil
	default:
		return "", ChannelRecord{}, "", "", "", 0, errors.New("payments ante unsupported message type")
	}
}

func settlementOperationForMessage(msg PaymentChannelModuleMessage) (SettlementOperation, error) {
	if err := msg.ValidateBasic(); err != nil {
		return SettlementOperation{}, err
	}
	channelID := ""
	nonce := uint64(1)
	stateHash := ""
	opType := BatchOperationSettle
	switch m := msg.(type) {
	case MsgOpenChannel:
		req := m.Normalize().Request
		channelID = req.ChannelID
		stateHash = HashParts("msg-open", channelID)
		opType = BatchOperationOpen
	case MsgCooperativeClose:
		req := m.Normalize().Request
		channelID = req.ChannelID
		nonce = req.ClosingStateWithSignatures().Nonce
		stateHash = req.ClosingStateWithSignatures().StateHash
		opType = BatchOperationSettle
	case MsgUnilateralClose:
		req := m.Normalize().Request
		channelID = req.ChannelID
		nonce = req.ClosingStateWithSignatures().Nonce
		stateHash = req.ClosingStateWithSignatures().StateHash
		opType = BatchOperationClose
	case MsgDisputeClose:
		req := m.Normalize().Request
		channelID = req.ChannelID
		nonce = req.NewerState.Nonce
		stateHash = req.NewerState.StateHash
		opType = BatchOperationDispute
	case MsgFinalizeClose:
		req := m.Normalize().Request
		channelID = req.ChannelID
		stateHash = HashParts("msg-finalize", channelID, fmt.Sprintf("%020d", req.CurrentHeight))
		opType = BatchOperationSettle
	case MsgSubmitCheckpoint:
		req := m.Normalize().Request
		channelID = req.ChannelID
		nonce = req.State.Nonce
		stateHash = req.State.StateHash
		opType = BatchOperationClose
	case MsgCancelExpiredChannel:
		msg := m.Normalize()
		channelID = msg.ChannelID
		stateHash = HashParts("msg-cancel-expired", channelID, fmt.Sprintf("%020d", msg.CurrentHeight))
		opType = BatchOperationClose
	case MsgRegisterChannelAdvertisement:
		msg := m.Normalize()
		channelID = msg.Advertisement.ChannelID
		stateHash = msg.Advertisement.AdvertisementHash
		opType = BatchOperationOpen
	default:
		return SettlementOperation{}, errors.New("payments blockstm unsupported message type")
	}
	if stateHash == "" {
		stateHash = HashParts("msg-operation", channelID, string(msg.PaymentChannelType()))
	}
	return SettlementOperation{
		OperationID:	HashParts("payment-channel-msg", string(msg.PaymentChannelType()), channelID, stateHash),
		OperationType:	opType,
		ChannelID:	channelID,
		Nonce:		nonce,
		StateHash:	stateHash,
	}.Normalize(), nil
}

func validateCloseMessageBasic(prefix, signer string, req ChannelCloseRequest) error {
	if err := addressing.ValidateUserAddress(prefix+" signer", signer); err != nil {
		return err
	}
	if err := ValidateHash(prefix+" channel id", req.ChannelID); err != nil {
		return err
	}
	if req.ClosingState.StateHash == "" {
		return errors.New(prefix + " closing state is required")
	}
	if req.CurrentHeight == 0 {
		return errors.New(prefix + " height must be positive")
	}
	if signer != req.Submitter {
		return errors.New(prefix + " signer must match submitter")
	}
	return validateNonNegativeInt(prefix+" settlement fee", req.SettlementFee)
}

func validateCloseMessageAndCooperativeState(state PaymentsState, req ChannelCloseRequest, signer string) (SettlementRecord, error) {
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return SettlementRecord{}, errors.New("payments cooperative close channel not found")
	}
	if !containsString(channel.Participants, signer) {
		return SettlementRecord{}, errors.New("payments cooperative close signer must be participant")
	}
	if err := req.ValidateForChannel(channel); err != nil {
		return SettlementRecord{}, err
	}
	return SettlementRecord{}, nil
}

func requirePaidAtLeast(field, paidText, requiredText string) error {
	paid, err := parseNonNegativeInt(field, paidText)
	if err != nil {
		return err
	}
	required, err := parseNonNegativeInt(field+" required", requiredText)
	if err != nil {
		return err
	}
	if paid.LT(required) {
		return fmt.Errorf("%s below required amount %s", field, requiredText)
	}
	return nil
}

func channelParticipantAmounts(state ChannelState, participant string) (string, string) {
	state = state.Normalize()
	participant = strings.TrimSpace(participant)
	for _, balance := range state.Balances {
		if balance.Participant == participant {
			return balance.Amount, "0"
		}
	}
	return "0", "0"
}

func channelByIDInSlice(channels []ChannelRecord, channelID string) (ChannelRecord, bool) {
	channelID = normalizeHash(channelID)
	for _, channel := range channels {
		channel = channel.Normalize()
		if channel.ChannelID == channelID {
			return channel, true
		}
	}
	return ChannelRecord{}, false
}

func normalizeChannelParticipants(participants []ChannelParticipant) []ChannelParticipant {
	out := make([]ChannelParticipant, len(participants))
	for i, participant := range participants {
		out[i] = participant.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ChannelID == out[j].ChannelID {
			return out[i].Address < out[j].Address
		}
		return out[i].ChannelID < out[j].ChannelID
	})
	return out
}

func normalizeChannelModuleChannels(channels []ChannelRecord) []ChannelRecord {
	out := make([]ChannelRecord, len(channels))
	for i, channel := range channels {
		out[i] = channel.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

func normalizeChannelModuleSettlements(settlements []SettlementRecord) []SettlementRecord {
	out := make([]SettlementRecord, len(settlements))
	for i, settlement := range settlements {
		out[i] = settlement.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

func normalizeChannelModuleTombstones(tombstones []ClosedChannelTombstone) []ClosedChannelTombstone {
	out := make([]ClosedChannelTombstone, len(tombstones))
	for i, tombstone := range tombstones {
		out[i] = tombstone.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

func normalizeChannelConfigs(configs []ChannelConfig) []ChannelConfig {
	out := make([]ChannelConfig, len(configs))
	for i, config := range configs {
		out[i] = config.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

func normalizeChannelFeeAccumulators(accumulators []ChannelFeeAccumulator) []ChannelFeeAccumulator {
	out := make([]ChannelFeeAccumulator, len(accumulators))
	for i, acc := range accumulators {
		out[i] = acc.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].BlockHeight == out[j].BlockHeight {
			return out[i].Bucket < out[j].Bucket
		}
		return out[i].BlockHeight < out[j].BlockHeight
	})
	return out
}

func accumulateChannelFeeAccumulator(acc ChannelFeeAccumulator, settlement SettlementRecord) (ChannelFeeAccumulator, error) {
	acc = acc.Normalize()
	settlement = settlement.Normalize()
	fees, err := parseNonNegativeInt("payments channel accumulator fee", acc.FeeAmount)
	if err != nil {
		return ChannelFeeAccumulator{}, err
	}
	penalties, err := parseNonNegativeInt("payments channel accumulator penalty", acc.PenaltyAmount)
	if err != nil {
		return ChannelFeeAccumulator{}, err
	}
	settlementFee, err := parseNonNegativeInt("payments settlement fee", settlement.SettlementFee)
	if err != nil {
		return ChannelFeeAccumulator{}, err
	}
	penaltyTotal, err := sumPenaltyAllocations(settlement.PenaltyAllocations)
	if err != nil {
		return ChannelFeeAccumulator{}, err
	}
	acc.FeeAmount = fees.Add(settlementFee).String()
	acc.PenaltyAmount = penalties.Add(penaltyTotal).String()
	acc.OperationCount++
	return acc.Normalize(), nil
}

func ApplyChannelFeeAccumulator(acc ChannelFeeAccumulator, settlement SettlementRecord) (ChannelFeeAccumulator, error) {
	next, err := accumulateChannelFeeAccumulator(acc, settlement)
	if err != nil {
		return ChannelFeeAccumulator{}, err
	}
	if err := next.Validate(); err != nil {
		return ChannelFeeAccumulator{}, err
	}
	return next, nil
}
