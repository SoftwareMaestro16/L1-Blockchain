package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
)

type AetherFailureKind string
type AetherRetryDecisionKind string

const (
	AetherFailureTransientQueueLimit	AetherFailureKind	= "transient_queue_limit"
	AetherFailureInvalidPayload		AetherFailureKind	= "invalid_payload"
	AetherFailureExecution			AetherFailureKind	= "execution_failure"
	AetherFailureExpired			AetherFailureKind	= "expired"

	AetherRetryNone		AetherRetryDecisionKind	= "none"
	AetherRetrySchedule	AetherRetryDecisionKind	= "schedule"
	AetherRetryExpired	AetherRetryDecisionKind	= "expired"
	AetherRetryBounce	AetherRetryDecisionKind	= "bounce"
)

type AetherRetryPolicy struct {
	MaxRetryCount		uint32
	BaseDelayHeights	uint64
	MaxDelayHeights		uint64
	RetryFee		sdkmath.Int
	BounceGasLimit		uint64
	MaxBouncePayloadLen	uint32
	PolicyHash		string
}

type AetherRetryState struct {
	MsgID			string
	RetryCount		uint32
	ForwardingFeeEscrow	sdkmath.Int
	LastAttemptHeight	uint64
	StateHash		string
}

type AetherRetryDecision struct {
	Kind			AetherRetryDecisionKind
	MsgID			string
	FailureKind		AetherFailureKind
	NextEligibleHeight	uint64
	RetryCount		uint32
	FeeCharged		sdkmath.Int
	RemainingFeeEscrow	sdkmath.Int
	DecisionHash		string
}

type AetherBounceEnvelope struct {
	OriginalMsgID		string
	BounceMsg		AetherMessage
	RemainingValueNAET	sdkmath.Int
	RemainingFee		sdkmath.Int
	BounceHash		string
}

func DecideAetherRetry(msg AetherMessage, receipt AetherMessageReceipt, state AetherRetryState, policy AetherRetryPolicy, currentHeight uint64, failure AetherFailureKind) (AetherRetryDecision, error) {
	if err := msg.Validate(); err != nil {
		return AetherRetryDecision{}, err
	}
	if err := receipt.Validate(); err != nil {
		return AetherRetryDecision{}, err
	}
	if receipt.MsgID != msg.MsgID {
		return AetherRetryDecision{}, errors.New("aether retry receipt message mismatch")
	}
	if currentHeight == 0 {
		return AetherRetryDecision{}, errors.New("aether retry current height must be positive")
	}
	if err := policy.Validate(); err != nil {
		return AetherRetryDecision{}, err
	}
	state = normalizeAetherRetryState(state)
	if state.MsgID == "" {
		state.MsgID = msg.MsgID
	}
	if err := state.Validate(); err != nil {
		return AetherRetryDecision{}, err
	}
	if state.MsgID != msg.MsgID {
		return AetherRetryDecision{}, errors.New("aether retry state message mismatch")
	}
	decision := AetherRetryDecision{
		Kind:			AetherRetryNone,
		MsgID:			msg.MsgID,
		FailureKind:		failure,
		RetryCount:		state.RetryCount,
		FeeCharged:		sdkmath.ZeroInt(),
		RemainingFeeEscrow:	state.ForwardingFeeEscrow,
	}
	if currentHeight > msg.ExpiryHeight || receipt.Status == ReceiptStatusExpired || failure == AetherFailureExpired {
		decision.Kind = AetherRetryExpired
		decision.DecisionHash = ComputeAetherRetryDecisionHash(decision)
		return decision, decision.Validate()
	}
	if failure != AetherFailureTransientQueueLimit {
		decision.DecisionHash = ComputeAetherRetryDecisionHash(decision)
		return decision, decision.Validate()
	}
	if state.RetryCount >= policy.MaxRetryCount {
		decision.DecisionHash = ComputeAetherRetryDecisionHash(decision)
		return decision, decision.Validate()
	}
	if state.ForwardingFeeEscrow.LT(policy.RetryFee) {
		decision.DecisionHash = ComputeAetherRetryDecisionHash(decision)
		return decision, decision.Validate()
	}
	nextRetry := state.RetryCount + 1
	delay := policy.BaseDelayHeights * uint64(nextRetry)
	if delay == 0 {
		delay = 1
	}
	if policy.MaxDelayHeights > 0 && delay > policy.MaxDelayHeights {
		delay = policy.MaxDelayHeights
	}
	decision.Kind = AetherRetrySchedule
	decision.RetryCount = nextRetry
	decision.NextEligibleHeight = currentHeight + delay
	decision.FeeCharged = policy.RetryFee
	decision.RemainingFeeEscrow = state.ForwardingFeeEscrow.Sub(policy.RetryFee)
	decision.DecisionHash = ComputeAetherRetryDecisionHash(decision)
	return decision, decision.Validate()
}

func ExpireAetherMessage(msg AetherMessage, height uint64, outputMessagesRoot string, stateWriteSummaryHash string) (AetherMessageReceipt, error) {
	if err := msg.Validate(); err != nil {
		return AetherMessageReceipt{}, err
	}
	if height <= msg.ExpiryHeight {
		return AetherMessageReceipt{}, errors.New("aether message is not expired")
	}
	return AetherReceiptFromMessage(msg, ReceiptStatusExpired, height, 0, sdkmath.ZeroInt(), nil, "ERR_MESSAGE_EXPIRED", outputMessagesRoot, stateWriteSummaryHash)
}

func BuildAetherBounce(original AetherMessage, receipt AetherMessageReceipt, remainingValue sdkmath.Int, remainingFee sdkmath.Int, nonce uint64, height uint64, policy AetherRetryPolicy) (AetherBounceEnvelope, error) {
	if err := original.Validate(); err != nil {
		return AetherBounceEnvelope{}, err
	}
	if err := receipt.Validate(); err != nil {
		return AetherBounceEnvelope{}, err
	}
	if err := policy.Validate(); err != nil {
		return AetherBounceEnvelope{}, err
	}
	if !original.Bounce {
		return AetherBounceEnvelope{}, errors.New("aether message bounce is disabled")
	}
	if receipt.MsgID != original.MsgID {
		return AetherBounceEnvelope{}, errors.New("aether bounce receipt message mismatch")
	}
	if remainingValue.IsNil() || remainingFee.IsNil() || remainingValue.IsNegative() || remainingFee.IsNegative() {
		return AetherBounceEnvelope{}, errors.New("aether bounce remaining value and fee must be non-negative")
	}
	if remainingValue.GT(original.ValueNAET) || remainingFee.GT(original.ForwardingFee) {
		return AetherBounceEnvelope{}, errors.New("aether bounce cannot create more value or fee than original")
	}
	payload := []byte(fmt.Sprintf("parent=%s;status=%s;error=%s", original.MsgID, receipt.Status, receipt.ErrorCode))
	if policy.MaxBouncePayloadLen > 0 && len(payload) > int(policy.MaxBouncePayloadLen) {
		payload = payload[:policy.MaxBouncePayloadLen]
	}
	bounceExpiry := original.ExpiryHeight
	if bounceExpiry < height {
		bounceExpiry = height
	}
	bounce, err := NewAetherMessage(AetherMessage{
		ParentMsgID:		original.MsgID,
		TraceID:		original.TraceID,
		Sender:			original.Receiver,
		SenderZoneID:		original.ReceiverZoneID,
		SenderShardID:		original.ReceiverShardID,
		Receiver:		original.Sender,
		ReceiverZoneID:		original.SenderZoneID,
		ReceiverShardID:	original.SenderShardID,
		ValueNAET:		remainingValue,
		Payload:		payload,
		PayloadType:		"aether.bounce",
		GasLimit:		policy.BounceGasLimit,
		GasPrice:		sdkmath.ZeroInt(),
		ForwardingFee:		remainingFee,
		ExpiryHeight:		bounceExpiry,
		Bounce:			false,
		ExecutionMode:		ExecutionModeAsync,
		OrderingClass:		OrderingClassStrictTraceOrder,
		RouteCommitment:	original.RouteCommitment,
		CreatedAtHeight:	height,
		Nonce:			nonce,
	})
	if err != nil {
		return AetherBounceEnvelope{}, err
	}
	envelope := AetherBounceEnvelope{
		OriginalMsgID:		original.MsgID,
		BounceMsg:		bounce,
		RemainingValueNAET:	remainingValue,
		RemainingFee:		remainingFee,
	}
	envelope.BounceHash = ComputeAetherBounceHash(envelope)
	return envelope, envelope.Validate()
}

func (p AetherRetryPolicy) Validate() error {
	if p.MaxRetryCount == 0 {
		return errors.New("aether retry max count must be positive")
	}
	if p.BaseDelayHeights == 0 {
		return errors.New("aether retry base delay must be positive")
	}
	if p.MaxDelayHeights > 0 && p.MaxDelayHeights < p.BaseDelayHeights {
		return errors.New("aether retry max delay must be >= base delay")
	}
	if p.RetryFee.IsNil() || p.RetryFee.IsNegative() {
		return errors.New("aether retry fee must be non-negative")
	}
	if p.BounceGasLimit == 0 {
		return errors.New("aether bounce gas limit must be positive")
	}
	if p.PolicyHash != "" {
		return validateOptionalHash("aether retry policy hash", p.PolicyHash)
	}
	return nil
}

func (s AetherRetryState) Validate() error {
	s = normalizeAetherRetryState(s)
	if err := validateOptionalHash("aether retry state message id", s.MsgID); err != nil {
		return err
	}
	if s.ForwardingFeeEscrow.IsNil() || s.ForwardingFeeEscrow.IsNegative() {
		return errors.New("aether retry forwarding fee escrow must be non-negative")
	}
	if s.StateHash != "" {
		if err := validateOptionalHash("aether retry state hash", s.StateHash); err != nil {
			return err
		}
		if s.StateHash != ComputeAetherRetryStateHash(s) {
			return errors.New("aether retry state hash mismatch")
		}
	}
	return nil
}

func (d AetherRetryDecision) Validate() error {
	if !IsAetherRetryDecisionKind(d.Kind) {
		return fmt.Errorf("unknown aether retry decision %q", d.Kind)
	}
	if err := validateOptionalHash("aether retry decision message id", d.MsgID); err != nil {
		return err
	}
	if !IsAetherFailureKind(d.FailureKind) {
		return fmt.Errorf("unknown aether failure kind %q", d.FailureKind)
	}
	if d.Kind == AetherRetrySchedule && d.NextEligibleHeight == 0 {
		return errors.New("aether retry schedule requires next eligible height")
	}
	if d.FeeCharged.IsNil() || d.RemainingFeeEscrow.IsNil() || d.FeeCharged.IsNegative() || d.RemainingFeeEscrow.IsNegative() {
		return errors.New("aether retry fees must be non-negative")
	}
	if err := validateOptionalHash("aether retry decision hash", d.DecisionHash); err != nil {
		return err
	}
	if d.DecisionHash != "" && d.DecisionHash != ComputeAetherRetryDecisionHash(d) {
		return errors.New("aether retry decision hash mismatch")
	}
	return nil
}

func (b AetherBounceEnvelope) Validate() error {
	if err := validateOptionalHash("aether bounce original message id", b.OriginalMsgID); err != nil {
		return err
	}
	if err := b.BounceMsg.Validate(); err != nil {
		return err
	}
	if b.BounceMsg.ParentMsgID != b.OriginalMsgID {
		return errors.New("aether bounce parent mismatch")
	}
	if b.RemainingValueNAET.IsNil() || b.RemainingFee.IsNil() || b.RemainingValueNAET.IsNegative() || b.RemainingFee.IsNegative() {
		return errors.New("aether bounce remaining value and fee must be non-negative")
	}
	if err := validateOptionalHash("aether bounce hash", b.BounceHash); err != nil {
		return err
	}
	if b.BounceHash != "" && b.BounceHash != ComputeAetherBounceHash(b) {
		return errors.New("aether bounce hash mismatch")
	}
	return nil
}

func ComputeAetherRetryStateHash(state AetherRetryState) string {
	state = normalizeAetherRetryState(state)
	return hashParts("aetra-aether-retry-state-v1", state.MsgID, fmt.Sprint(state.RetryCount), state.ForwardingFeeEscrow.String(), fmt.Sprint(state.LastAttemptHeight))
}

func ComputeAetherRetryDecisionHash(decision AetherRetryDecision) string {
	return hashParts("aetra-aether-retry-decision-v1", string(decision.Kind), decision.MsgID, string(decision.FailureKind), fmt.Sprint(decision.NextEligibleHeight), fmt.Sprint(decision.RetryCount), decision.FeeCharged.String(), decision.RemainingFeeEscrow.String())
}

func ComputeAetherBounceHash(bounce AetherBounceEnvelope) string {
	return hashParts("aetra-aether-bounce-v1", bounce.OriginalMsgID, bounce.BounceMsg.MsgID, bounce.RemainingValueNAET.String(), bounce.RemainingFee.String())
}

func IsAetherFailureKind(kind AetherFailureKind) bool {
	switch kind {
	case AetherFailureTransientQueueLimit, AetherFailureInvalidPayload, AetherFailureExecution, AetherFailureExpired:
		return true
	default:
		return false
	}
}

func IsAetherRetryDecisionKind(kind AetherRetryDecisionKind) bool {
	switch kind {
	case AetherRetryNone, AetherRetrySchedule, AetherRetryExpired, AetherRetryBounce:
		return true
	default:
		return false
	}
}

func normalizeAetherRetryState(state AetherRetryState) AetherRetryState {
	if state.ForwardingFeeEscrow.IsNil() {
		state.ForwardingFeeEscrow = sdkmath.ZeroInt()
	}
	return state
}
