package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

type ConditionalPaymentMessageType string

const (
	ConditionalMsgRegisterPromise			ConditionalPaymentMessageType	= "MsgRegisterPromise"
	ConditionalMsgResolveWithPreimage		ConditionalPaymentMessageType	= "MsgResolveWithPreimage"
	ConditionalMsgExpirePromise			ConditionalPaymentMessageType	= "MsgExpirePromise"
	ConditionalMsgBatchResolvePromises		ConditionalPaymentMessageType	= "MsgBatchResolvePromises"
	ConditionalMsgDisputeCondition			ConditionalPaymentMessageType	= "MsgDisputeCondition"
	ConditionalMsgFinalizeConditionSettlement	ConditionalPaymentMessageType	= "MsgFinalizeConditionSettlement"
)

type ConditionRoot struct {
	ChannelID	string
	Nonce		uint64
	RootHash	string
	ConditionCount	uint32
	PromiseIDs	[]string
	CommitmentHash	string
}

type PreimageClaim struct {
	ClaimID		string
	ChannelID	string
	PromiseID	string
	Resolver	string
	PreimageHash	string
	ResolvedHeight	uint64
	ExpiresHeight	uint64
	EvidenceHash	string
}

type PromiseLink struct {
	LinkID		string
	RouteID		string
	PreviousID	string
	PromiseID	string
	NextID		string
	ChannelID	string
	TimeoutHeight	uint64
	HashLock	string
}

type PromiseTimeout struct {
	TimeoutID	string
	ChannelID	string
	PromiseID	string
	TimeoutHeight	uint64
	SafetyMargin	uint64
	Expired		bool
}

type ConditionSettlementRecord struct {
	SettlementID	string
	RouteID		string
	Mode		ConditionSettlementMode
	Resolver	string
	ResolvedAt	uint64
	EvidenceHash	string
	Resolutions	[]ConditionResolution
	RootUpdates	[]ConditionRootUpdate
	FeeClaims	[]RouteFeeClaim
}

type ConditionalPaymentsModuleState struct {
	Promises	[]ConditionalPromise
	ConditionRoots	[]ConditionRoot
	PreimageClaims	[]PreimageClaim
	PromiseLinks	[]PromiseLink
	PromiseTimeouts	[]PromiseTimeout
	Settlements	[]ConditionSettlementRecord
	ExpiredClaims	[]ConditionClaimRecord
}

type ConditionalPaymentMessage interface {
	ConditionalPaymentType() ConditionalPaymentMessageType
	ValidateBasic() error
}

type MsgRegisterPromise struct {
	Signer		string
	ChannelID	string
	BaseState	ChannelState
	Promises	[]ConditionalPromise
	CurrentHeight	uint64
}

type MsgResolveWithPreimage struct {
	Request PreimageRevealRequest
}

type MsgExpirePromise struct {
	Request PromiseExpiryRequest
}

type MsgBatchResolvePromises struct {
	Request BatchConditionSettlementRequest
}

type MsgDisputeCondition struct {
	Signer			string
	ChannelID		string
	Promise			ConditionalPromise
	Resolution		ConditionResolution
	Reason			string
	CurrentHeight		uint64
	VerificationFeePaid	string
}

type MsgFinalizeConditionSettlement struct {
	Signer		string
	Settlement	ConditionSettlementRecord
	CurrentHeight	uint64
}

func SnapshotConditionalPaymentsModuleState(state PaymentsState) (ConditionalPaymentsModuleState, error) {
	state = state.Export()
	out := ConditionalPaymentsModuleState{
		ExpiredClaims: append([]ConditionClaimRecord(nil), state.ConditionClaims...),
	}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.LatestState.ConditionRoot != "" && channel.LatestState.ConditionCount > 0 {
			out.ConditionRoots = append(out.ConditionRoots, ConditionRootFromState(channel, channel.LatestState))
		}
		for _, condition := range channel.LatestState.Conditions {
			promise := PromiseFromCondition(channel, condition)
			out.Promises = append(out.Promises, promise)
			out.PromiseTimeouts = append(out.PromiseTimeouts, PromiseTimeoutFromPromise(promise, channel.CloseDelay+channel.DisputePeriod, false))
			if promise.PreviousPromiseIDOptional != "" || promise.NextPromiseIDOptional != "" || promise.RouteIDOptional != "" {
				out.PromiseLinks = append(out.PromiseLinks, PromiseLinkFromPromise(promise))
			}
		}
	}
	for _, claim := range state.ConditionClaims {
		claim = claim.Normalize()
		if claim.PreimageHash == "" {
			continue
		}
		out.PreimageClaims = append(out.PreimageClaims, PreimageClaim{
			ClaimID:	HashParts("preimage-claim", claim.ChannelID, claim.ConditionID, claim.EvidenceHash),
			ChannelID:	claim.ChannelID,
			PromiseID:	claim.ConditionID,
			PreimageHash:	claim.PreimageHash,
			ResolvedHeight:	claim.ResolvedHeight,
			ExpiresHeight:	claim.ExpiresHeight,
			EvidenceHash:	claim.EvidenceHash,
		}.Normalize())
	}
	return out.Normalize(), out.Validate()
}

func ApplyConditionalPaymentMessage(state PaymentsState, msg ConditionalPaymentMessage) (PaymentsState, ConditionalPaymentsModuleState, error) {
	state = state.Export()
	if msg == nil {
		return PaymentsState{}, ConditionalPaymentsModuleState{}, errors.New("payments conditional message is required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return PaymentsState{}, ConditionalPaymentsModuleState{}, err
	}
	var next PaymentsState
	var err error
	switch m := msg.(type) {
	case MsgRegisterPromise:
		next, _, err = RegisterConditionalPromises(state, m)
	case MsgResolveWithPreimage:
		req := m.Normalize().Request
		next, _, err = RevealPromisePreimage(state, req)
		if err == nil {
			next, err = removePromisesFromLatestState(next, req.ChannelID, req.Promises)
		}
	case MsgExpirePromise:
		req := m.Normalize().Request
		var update ConditionRootUpdate
		next, _, update, err = ExpireConditionalPromises(state, req)
		if err == nil {
			next, err = applyConditionRootUpdateToLatestState(next, update)
		}
	case MsgBatchResolvePromises:
		var result BatchConditionSettlementResult
		next, result, err = BatchSettleLinkedPromises(state, m.Normalize().Request)
		if err == nil {
			for _, update := range result.ConditionRootUpdates {
				next, err = applyConditionRootUpdateToLatestState(next, update)
				if err != nil {
					break
				}
			}
		}
		if err == nil {
			next = appendConditionSettlementRecord(next, ConditionSettlementRecordFromBatch(m.Normalize().Request, result))
		}
	case MsgDisputeCondition:
		next, err = DisputeConditionResolution(state, m)
	case MsgFinalizeConditionSettlement:
		next, err = FinalizeConditionSettlementRecord(state, m)
	default:
		return PaymentsState{}, ConditionalPaymentsModuleState{}, errors.New("payments conditional message type is unsupported")
	}
	if err != nil {
		return PaymentsState{}, ConditionalPaymentsModuleState{}, err
	}
	snapshot, err := SnapshotConditionalPaymentsModuleState(next)
	if err != nil {
		return PaymentsState{}, ConditionalPaymentsModuleState{}, err
	}
	return next, snapshot, nil
}

func RegisterConditionalPromises(state PaymentsState, msg MsgRegisterPromise) (PaymentsState, ConditionRootUpdate, error) {
	state = state.Export()
	msg = msg.Normalize()
	if err := msg.ValidateBasic(); err != nil {
		return PaymentsState{}, ConditionRootUpdate{}, err
	}
	index, channel, found := state.ChannelIndex(msg.ChannelID)
	if !found {
		return PaymentsState{}, ConditionRootUpdate{}, errors.New("payments conditional register channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, ConditionRootUpdate{}, errors.New("payments conditional register requires open channel")
	}
	if !containsString(channel.Participants, msg.Signer) {
		return PaymentsState{}, ConditionRootUpdate{}, errors.New("payments conditional register signer must be participant")
	}
	base := msg.BaseState
	if base.StateHash == "" {
		base = channel.LatestState
	}
	nextState, update, err := BuildConditionRootUpdateFromPromises(channel, base, msg.Promises, state.ConditionClaims)
	if err != nil {
		return PaymentsState{}, ConditionRootUpdate{}, err
	}
	if err := ValidateReservedBalancesForConditions(channel, nextState); err != nil {
		return PaymentsState{}, ConditionRootUpdate{}, err
	}
	nextState, err = signConditionModuleState(channel, nextState)
	if err != nil {
		return PaymentsState{}, ConditionRootUpdate{}, err
	}
	next := state.Clone()
	next.Channels[index].LatestState = nextState.Normalize()
	sortChannels(next.Channels)
	return next, update, next.Validate()
}

func DisputeConditionResolution(state PaymentsState, msg MsgDisputeCondition) (PaymentsState, error) {
	state = state.Export()
	msg = msg.Normalize()
	if err := msg.ValidateBasic(); err != nil {
		return PaymentsState{}, err
	}
	channel, found := state.ChannelByID(msg.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments condition dispute channel not found")
	}
	if !containsString(channel.Participants, msg.Signer) {
		return PaymentsState{}, errors.New("payments condition dispute signer must be participant")
	}
	if err := msg.Promise.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	condition := msg.Promise.ToConditionalPayment()
	if err := msg.Resolution.ValidateForCondition(condition, channel); err == nil {
		return PaymentsState{}, errors.New("payments condition dispute requires invalid resolution evidence")
	}
	next := state.Clone()
	next.Events = append(next.Events, PaymentEvent{
		EventID:	HashParts("condition-dispute", msg.ChannelID, msg.Promise.PromiseID, msg.Resolution.EvidenceHash),
		EventType:	"condition_dispute",
		ChannelID:	msg.ChannelID,
		Height:		msg.CurrentHeight,
		Attributes: []PaymentEventAttribute{
			{Key: "promise_id", Value: msg.Promise.PromiseID},
			{Key: "reason", Value: msg.Reason},
			{Key: "evidence_hash", Value: HashParts("condition-dispute-evidence", msg.Reason, msg.Promise.PromiseHash, msg.Resolution.EvidenceHash)},
		},
	}.Normalize())
	return next, next.Validate()
}

func FinalizeConditionSettlementRecord(state PaymentsState, msg MsgFinalizeConditionSettlement) (PaymentsState, error) {
	state = state.Export()
	msg = msg.Normalize()
	if err := msg.ValidateBasic(); err != nil {
		return PaymentsState{}, err
	}
	next := appendConditionSettlementRecord(state, msg.Settlement)
	return next, next.Validate()
}

func ValidateReservedBalancesForConditions(channel ChannelRecord, state ChannelState) error {
	channel = channel.Normalize()
	state = state.Normalize()
	reserved := make(map[string]sdkmath.Int, len(channel.Participants))
	for _, condition := range state.Conditions {
		condition = condition.Normalize()
		if !containsString(channel.Participants, condition.Payer) {
			return errors.New("payments condition payer must be channel participant")
		}
		amount, err := parsePositiveInt("payments condition reserved amount", condition.Amount)
		if err != nil {
			return err
		}
		current, found := reserved[condition.Payer]
		if !found {
			current = sdkmath.ZeroInt()
		}
		reserved[condition.Payer] = current.Add(amount)
	}
	balances := make(map[string]sdkmath.Int, len(state.Balances))
	for _, balance := range state.Balances {
		amount, err := parseNonNegativeInt("payments condition balance", balance.Amount)
		if err != nil {
			return err
		}
		balances[balance.Participant] = amount
	}
	for participant, amount := range reserved {
		balance, found := balances[participant]
		if !found {
			balance = sdkmath.ZeroInt()
		}
		if balance.LT(amount) {
			return errors.New("payments reserved condition amount exceeds participant balance")
		}
	}
	return nil
}

func removePromisesFromLatestState(state PaymentsState, channelID string, promises []ConditionalPromise) (PaymentsState, error) {
	state = state.Export()
	_, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments condition root update channel not found")
	}
	_, update, err := BuildConditionRootAfterExpiry(channel.LatestState, promises)
	if err != nil {
		return PaymentsState{}, err
	}
	return applyConditionRootUpdateToLatestState(state, update)
}

func applyConditionRootUpdateToLatestState(state PaymentsState, update ConditionRootUpdate) (PaymentsState, error) {
	state = state.Export()
	update = normalizeConditionRootUpdates([]ConditionRootUpdate{update})[0]
	index, channel, found := state.ChannelIndex(update.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments condition root update channel not found")
	}
	latest := channel.LatestState.Normalize()
	if latest.Nonce != update.Nonce {
		return PaymentsState{}, errors.New("payments condition root update nonce mismatch")
	}
	latest.Conditions = normalizeConditions(update.Conditions)
	latest.ConditionRoot = update.ConditionRoot
	latest.PendingConditionsRoot = update.ConditionRoot
	latest.ConditionCount = update.ConditionCount
	signed, err := signConditionModuleState(channel, latest)
	if err != nil {
		return PaymentsState{}, err
	}
	if err := ValidateReservedBalancesForConditions(channel, signed); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	next.Channels[index].LatestState = signed.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func signConditionModuleState(channel ChannelRecord, state ChannelState) (ChannelState, error) {
	channel = channel.Normalize()
	state = state.Normalize()
	state.Signatures = nil
	state.SignaturePreimageHash = ComputeStateSignaturePreimageHash(state)
	state.StateHash = ComputeStateHash(state)
	for _, signer := range channel.RequiredSigners {
		signature, err := SignatureForState(state, signer)
		if err != nil {
			return ChannelState{}, err
		}
		state.Signatures = append(state.Signatures, signature)
	}
	state = state.Normalize()
	if err := state.ValidateForChannel(channel, false); err != nil {
		return ChannelState{}, err
	}
	return state, nil
}

func ConditionRootFromState(channel ChannelRecord, state ChannelState) ConditionRoot {
	channel = channel.Normalize()
	state = state.Normalize()
	ids := make([]string, 0, len(state.Conditions))
	for _, condition := range state.Conditions {
		ids = append(ids, condition.Normalize().ConditionID)
	}
	sort.Strings(ids)
	root := ConditionRoot{
		ChannelID:	channel.ChannelID,
		Nonce:		state.Nonce,
		RootHash:	state.ConditionRoot,
		ConditionCount:	state.ConditionCount,
		PromiseIDs:	ids,
		CommitmentHash:	ComputeConditionRootCommitment(channel, state),
	}
	return root.Normalize()
}

func PromiseFromCondition(channel ChannelRecord, condition ConditionalPayment) ConditionalPromise {
	channel = channel.Normalize()
	condition = condition.Normalize()
	promise := ConditionalPromise{
		PromiseID:	condition.ConditionID,
		ChannelID:	channel.ChannelID,
		Source:		condition.Payer,
		Destination:	condition.Payee,
		Amount:		condition.Amount,
		Fee:		"0",
		HashLock:	condition.HashLock,
		TimeoutHeight:	condition.TimeoutHeight,
		ConditionType:	condition.ConditionType,
		Nonce:		condition.NonceStart,
	}
	promise.PromiseHash = ComputeConditionalTransferPromiseHash(promise)
	return promise.Normalize()
}

func PromiseLinkFromPromise(promise ConditionalPromise) PromiseLink {
	promise = promise.Normalize()
	return PromiseLink{
		LinkID:		HashParts("promise-link", promise.RouteIDOptional, promise.PreviousPromiseIDOptional, promise.PromiseID, promise.NextPromiseIDOptional),
		RouteID:	promise.RouteIDOptional,
		PreviousID:	promise.PreviousPromiseIDOptional,
		PromiseID:	promise.PromiseID,
		NextID:		promise.NextPromiseIDOptional,
		ChannelID:	promise.ChannelID,
		TimeoutHeight:	promise.TimeoutHeight,
		HashLock:	promise.HashLock,
	}.Normalize()
}

func PromiseTimeoutFromPromise(promise ConditionalPromise, margin uint64, expired bool) PromiseTimeout {
	promise = promise.Normalize()
	return PromiseTimeout{
		TimeoutID:	HashParts("promise-timeout", promise.ChannelID, promise.PromiseID, fmt.Sprintf("%020d", promise.TimeoutHeight)),
		ChannelID:	promise.ChannelID,
		PromiseID:	promise.PromiseID,
		TimeoutHeight:	promise.TimeoutHeight,
		SafetyMargin:	margin,
		Expired:	expired,
	}.Normalize()
}

func ConditionSettlementRecordFromBatch(req BatchConditionSettlementRequest, result BatchConditionSettlementResult) ConditionSettlementRecord {
	req = req.Normalize()
	result = result.Normalize()
	record := ConditionSettlementRecord{
		SettlementID:	HashParts("condition-settlement", result.RouteID, result.EvidenceHash, fmt.Sprintf("%020d", req.CurrentHeight)),
		RouteID:	result.RouteID,
		Mode:		req.Mode,
		Resolver:	req.Resolver,
		ResolvedAt:	req.CurrentHeight,
		EvidenceHash:	result.EvidenceHash,
		Resolutions:	result.Resolutions,
		RootUpdates:	result.ConditionRootUpdates,
		FeeClaims:	result.FeeClaims,
	}
	return record.Normalize()
}

func (m MsgRegisterPromise) ConditionalPaymentType() ConditionalPaymentMessageType {
	return ConditionalMsgRegisterPromise
}
func (m MsgRegisterPromise) Normalize() MsgRegisterPromise {
	m.Signer = strings.TrimSpace(m.Signer)
	m.ChannelID = normalizeHash(m.ChannelID)
	m.BaseState = m.BaseState.Normalize()
	m.Promises = normalizeConditionalPromises(m.Promises)
	return m
}
func (m MsgRegisterPromise) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg register promise signer", msg.Signer); err != nil {
		return err
	}
	if err := ValidateHash("payments msg register promise channel id", msg.ChannelID); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg register promise height must be positive")
	}
	if len(msg.Promises) == 0 {
		return errors.New("payments msg register promise requires promises")
	}
	for _, promise := range msg.Promises {
		if promise.ChannelID != msg.ChannelID {
			return errors.New("payments msg register promise channel mismatch")
		}
		if err := promise.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

func (m MsgResolveWithPreimage) ConditionalPaymentType() ConditionalPaymentMessageType {
	return ConditionalMsgResolveWithPreimage
}
func (m MsgResolveWithPreimage) Normalize() MsgResolveWithPreimage {
	m.Request = m.Request.Normalize()
	return m
}
func (m MsgResolveWithPreimage) ValidateBasic() error {
	req := m.Normalize().Request
	if req.CurrentHeight == 0 || req.Preimage == "" || len(req.Promises) == 0 {
		return errors.New("payments msg preimage resolve requires height preimage and promises")
	}
	return addressing.ValidateUserAddress("payments msg preimage resolver", req.Revealer)
}

func (m MsgExpirePromise) ConditionalPaymentType() ConditionalPaymentMessageType {
	return ConditionalMsgExpirePromise
}
func (m MsgExpirePromise) Normalize() MsgExpirePromise	{ m.Request = m.Request.Normalize(); return m }
func (m MsgExpirePromise) ValidateBasic() error {
	req := m.Normalize().Request
	if req.CurrentHeight == 0 || len(req.Promises) == 0 {
		return errors.New("payments msg expire promise requires height and promises")
	}
	return addressing.ValidateUserAddress("payments msg expire resolver", req.Resolver)
}

func (m MsgBatchResolvePromises) ConditionalPaymentType() ConditionalPaymentMessageType {
	return ConditionalMsgBatchResolvePromises
}
func (m MsgBatchResolvePromises) Normalize() MsgBatchResolvePromises {
	m.Request = m.Request.Normalize()
	return m
}
func (m MsgBatchResolvePromises) ValidateBasic() error {
	req := m.Normalize().Request
	if req.CurrentHeight == 0 {
		return errors.New("payments msg batch resolve height must be positive")
	}
	return addressing.ValidateUserAddress("payments msg batch resolver", req.Resolver)
}

func (m MsgDisputeCondition) ConditionalPaymentType() ConditionalPaymentMessageType {
	return ConditionalMsgDisputeCondition
}
func (m MsgDisputeCondition) Normalize() MsgDisputeCondition {
	m.Signer = strings.TrimSpace(m.Signer)
	m.ChannelID = normalizeHash(m.ChannelID)
	m.Promise = m.Promise.Normalize()
	m.Resolution = m.Resolution.Normalize()
	m.Reason = strings.TrimSpace(m.Reason)
	m.VerificationFeePaid = strings.TrimSpace(m.VerificationFeePaid)
	if m.VerificationFeePaid == "" {
		m.VerificationFeePaid = "0"
	}
	return m
}
func (m MsgDisputeCondition) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg condition dispute signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 || msg.Reason == "" {
		return errors.New("payments msg condition dispute requires height and reason")
	}
	if err := msg.Promise.ValidateBasic(); err != nil {
		return err
	}
	if msg.Resolution.ConditionID == "" {
		return errors.New("payments msg condition dispute resolution is required")
	}
	return validateNonNegativeInt("payments msg condition dispute fee", msg.VerificationFeePaid)
}

func (m MsgFinalizeConditionSettlement) ConditionalPaymentType() ConditionalPaymentMessageType {
	return ConditionalMsgFinalizeConditionSettlement
}
func (m MsgFinalizeConditionSettlement) Normalize() MsgFinalizeConditionSettlement {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Settlement = m.Settlement.Normalize()
	return m
}
func (m MsgFinalizeConditionSettlement) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg finalize condition signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg finalize condition height must be positive")
	}
	return msg.Settlement.Validate()
}

func (s ConditionalPaymentsModuleState) Normalize() ConditionalPaymentsModuleState {
	s.Promises = normalizeConditionalPromises(s.Promises)
	s.ConditionRoots = normalizeConditionRoots(s.ConditionRoots)
	s.PreimageClaims = normalizePreimageClaims(s.PreimageClaims)
	s.PromiseLinks = normalizePromiseLinks(s.PromiseLinks)
	s.PromiseTimeouts = normalizePromiseTimeouts(s.PromiseTimeouts)
	s.Settlements = normalizeConditionSettlementRecords(s.Settlements)
	for i := range s.ExpiredClaims {
		s.ExpiredClaims[i] = s.ExpiredClaims[i].Normalize()
	}
	sortConditionClaimRecords(s.ExpiredClaims)
	return s
}

func (s ConditionalPaymentsModuleState) Validate() error {
	state := s.Normalize()
	for _, promise := range state.Promises {
		if err := promise.ValidateBasic(); err != nil {
			return err
		}
	}
	for _, root := range state.ConditionRoots {
		if err := root.Validate(); err != nil {
			return err
		}
	}
	for _, claim := range state.PreimageClaims {
		if err := claim.Validate(); err != nil {
			return err
		}
	}
	for _, link := range state.PromiseLinks {
		if err := link.Validate(); err != nil {
			return err
		}
	}
	for _, timeout := range state.PromiseTimeouts {
		if err := timeout.Validate(); err != nil {
			return err
		}
	}
	for _, settlement := range state.Settlements {
		if err := settlement.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r ConditionRoot) Normalize() ConditionRoot {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.RootHash = normalizeHash(r.RootHash)
	r.PromiseIDs = normalizeHashSlice(r.PromiseIDs)
	r.CommitmentHash = normalizeOptionalHash(r.CommitmentHash)
	return r
}
func (r ConditionRoot) Validate() error {
	root := r.Normalize()
	if err := ValidateHash("payments condition root channel", root.ChannelID); err != nil {
		return err
	}
	if root.Nonce == 0 {
		return errors.New("payments condition root nonce must be positive")
	}
	if root.ConditionCount != uint32(len(root.PromiseIDs)) {
		return errors.New("payments condition root count mismatch")
	}
	if err := ValidateHash("payments condition root hash", root.RootHash); err != nil {
		return err
	}
	return ValidateHash("payments condition root commitment", root.CommitmentHash)
}

func (c PreimageClaim) Normalize() PreimageClaim {
	c.ClaimID = normalizeOptionalHash(c.ClaimID)
	c.ChannelID = normalizeHash(c.ChannelID)
	c.PromiseID = normalizeHash(c.PromiseID)
	c.Resolver = strings.TrimSpace(c.Resolver)
	c.PreimageHash = normalizeHash(c.PreimageHash)
	c.EvidenceHash = normalizeHash(c.EvidenceHash)
	return c
}
func (c PreimageClaim) Validate() error {
	claim := c.Normalize()
	if err := ValidateHash("payments preimage claim id", claim.ClaimID); err != nil {
		return err
	}
	if err := ValidateHash("payments preimage claim channel", claim.ChannelID); err != nil {
		return err
	}
	if claim.ResolvedHeight == 0 || claim.ExpiresHeight <= claim.ResolvedHeight {
		return errors.New("payments preimage claim replay horizon must advance")
	}
	if err := ValidateHash("payments preimage claim promise", claim.PromiseID); err != nil {
		return err
	}
	if err := ValidateHash("payments preimage claim hash", claim.PreimageHash); err != nil {
		return err
	}
	return ValidateHash("payments preimage claim evidence", claim.EvidenceHash)
}

func (l PromiseLink) Normalize() PromiseLink {
	l.LinkID = normalizeOptionalHash(l.LinkID)
	l.RouteID = normalizeOptionalHash(l.RouteID)
	l.PreviousID = normalizeOptionalHash(l.PreviousID)
	l.PromiseID = normalizeHash(l.PromiseID)
	l.NextID = normalizeOptionalHash(l.NextID)
	l.ChannelID = normalizeHash(l.ChannelID)
	l.HashLock = normalizeHash(l.HashLock)
	return l
}
func (l PromiseLink) Validate() error {
	link := l.Normalize()
	if err := ValidateHash("payments promise link id", link.LinkID); err != nil {
		return err
	}
	if err := ValidateHash("payments promise link promise", link.PromiseID); err != nil {
		return err
	}
	if link.PreviousID == "" && link.NextID == "" && link.RouteID == "" {
		return errors.New("payments promise link requires route or adjacent promise")
	}
	if link.TimeoutHeight == 0 {
		return errors.New("payments promise link timeout must be positive")
	}
	return ValidateHash("payments promise link hash lock", link.HashLock)
}

func (t PromiseTimeout) Normalize() PromiseTimeout {
	t.TimeoutID = normalizeOptionalHash(t.TimeoutID)
	t.ChannelID = normalizeHash(t.ChannelID)
	t.PromiseID = normalizeHash(t.PromiseID)
	return t
}
func (t PromiseTimeout) Validate() error {
	timeout := t.Normalize()
	if err := ValidateHash("payments promise timeout id", timeout.TimeoutID); err != nil {
		return err
	}
	if timeout.TimeoutHeight == 0 {
		return errors.New("payments promise timeout height must be positive")
	}
	return ValidateHash("payments promise timeout promise", timeout.PromiseID)
}

func (r ConditionSettlementRecord) Normalize() ConditionSettlementRecord {
	r.SettlementID = normalizeOptionalHash(r.SettlementID)
	r.RouteID = normalizeHash(r.RouteID)
	r.Resolver = strings.TrimSpace(r.Resolver)
	r.EvidenceHash = normalizeHash(r.EvidenceHash)
	r.Resolutions = normalizeConditionResolutions(r.Resolutions)
	r.RootUpdates = normalizeConditionRootUpdates(r.RootUpdates)
	for i := range r.FeeClaims {
		r.FeeClaims[i] = r.FeeClaims[i].Normalize()
	}
	sort.SliceStable(r.FeeClaims, func(i, j int) bool { return r.FeeClaims[i].PromiseID < r.FeeClaims[j].PromiseID })
	return r
}
func (r ConditionSettlementRecord) Validate() error {
	record := r.Normalize()
	if err := ValidateHash("payments condition settlement id", record.SettlementID); err != nil {
		return err
	}
	if record.Mode != ConditionSettlementModePreimage && record.Mode != ConditionSettlementModeExpiry {
		return errors.New("payments condition settlement mode is invalid")
	}
	if err := addressing.ValidateUserAddress("payments condition settlement resolver", record.Resolver); err != nil {
		return err
	}
	if record.ResolvedAt == 0 || len(record.Resolutions) == 0 || len(record.RootUpdates) == 0 {
		return errors.New("payments condition settlement requires height resolutions and root updates")
	}
	return ValidateHash("payments condition settlement evidence", record.EvidenceHash)
}

func appendConditionSettlementRecord(state PaymentsState, record ConditionSettlementRecord) PaymentsState {
	record = record.Normalize()
	event := PaymentEvent{
		EventID:	HashParts("condition-settlement-event", record.SettlementID),
		EventType:	"condition_settlement",
		ChannelID:	record.RootUpdates[0].ChannelID,
		Height:		record.ResolvedAt,
		Attributes: []PaymentEventAttribute{
			{Key: "route_id", Value: record.RouteID},
			{Key: "mode", Value: string(record.Mode)},
		},
	}.Normalize()
	next := state.Clone()
	next.Events = append(next.Events, event)
	return next
}

func normalizeConditionRoots(values []ConditionRoot) []ConditionRoot {
	out := make([]ConditionRoot, len(values))
	for i, v := range values {
		out[i] = v.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}
func normalizePreimageClaims(values []PreimageClaim) []PreimageClaim {
	out := make([]PreimageClaim, len(values))
	for i, v := range values {
		out[i] = v.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ClaimID < out[j].ClaimID })
	return out
}
func normalizePromiseLinks(values []PromiseLink) []PromiseLink {
	out := make([]PromiseLink, len(values))
	for i, v := range values {
		out[i] = v.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].LinkID < out[j].LinkID })
	return out
}
func normalizePromiseTimeouts(values []PromiseTimeout) []PromiseTimeout {
	out := make([]PromiseTimeout, len(values))
	for i, v := range values {
		out[i] = v.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].TimeoutID < out[j].TimeoutID })
	return out
}
func normalizeConditionSettlementRecords(values []ConditionSettlementRecord) []ConditionSettlementRecord {
	out := make([]ConditionSettlementRecord, len(values))
	for i, v := range values {
		out[i] = v.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].SettlementID < out[j].SettlementID })
	return out
}
