package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type PaymentSettlementRuleID string

const (
	PaymentSettlementRuleIntermediateOffchain	PaymentSettlementRuleID	= "intermediate-states-offchain-or-queued"
	PaymentSettlementRuleDeterministicReplay	PaymentSettlementRuleID	= "disputes-resolve-by-deterministic-replay"
	PaymentSettlementRuleFinancialStateWrite	PaymentSettlementRuleID	= "final-settlement-writes-financial-zone-state"
	PaymentSettlementRuleConditionalProof		PaymentSettlementRuleID	= "conditional-transfers-require-resolution"
	PaymentSettlementRuleExpiryRelease		PaymentSettlementRuleID	= "expired-conditions-release-reserved-funds"
	PaymentSettlementRuleCommittedRoutesOnly	PaymentSettlementRuleID	= "route-hints-advisory-unless-committed"
)

type PaymentMessageType string

const (
	PaymentMessageCreateIntent	PaymentMessageType	= "MsgCreatePaymentIntent"
	PaymentMessageOpenChannel	PaymentMessageType	= "MsgOpenPaymentChannel"
	PaymentMessageUpdateChannel	PaymentMessageType	= "MsgUpdatePaymentChannel"
	PaymentMessageCloseChannel	PaymentMessageType	= "MsgClosePaymentChannel"
	PaymentMessageDisputeChannel	PaymentMessageType	= "MsgDisputePaymentChannel"
	PaymentMessageCreateCondition	PaymentMessageType	= "MsgCreateConditionalPayment"
	PaymentMessageResolveCondition	PaymentMessageType	= "MsgResolveConditionalPayment"
	PaymentMessageExpireCondition	PaymentMessageType	= "MsgExpireConditionalPayment"
	PaymentMessageSettlePayment	PaymentMessageType	= "MsgSettlePayment"
)

type PaymentSettlementRule struct {
	Rule		PaymentSettlementRuleID
	ConsensusEffect	string
}

type PaymentMessageDescriptor struct {
	Message			PaymentMessageType
	Purpose			string
	RequiredValidation	[]string
}

type MsgCreatePaymentIntent struct {
	PaymentID	string
	Payer		string
	Payee		string
	TargetIdentity	string
	Amount		string
	Denom		string
	MaxFee		string
	ExpiryHeight	uint64
	IdempotencyKey	string
	MessageHash	string
}

type MsgOpenPaymentChannel struct {
	Channel			PaymentChannel
	CollateralAvailable	string
	ParticipantSignatures	[]string
	IdempotencyKey		string
	MessageHash		string
}

type MsgUpdatePaymentChannel struct {
	ChannelID	string
	Submitter	string
	PreviousNonce	uint64
	NewNonce	uint64
	SignedStateHash	string
	BalanceRoot	string
	ConditionRoot	string
	MessageHash	string
}

type MsgClosePaymentChannel struct {
	PaymentID		string
	ChannelID		string
	LatestStateHash		string
	ChallengeStart		uint64
	ChallengeEnd		uint64
	SettlementStatus	NativePaymentSettlementStatus
	CollateralRoot		string
	MessageHash		string
}

type MsgDisputePaymentChannel struct {
	Dispute		PaymentDispute
	StaleNonce	uint64
	NewerNonce	uint64
	CurrentHeight	uint64
	MessageHash	string
}

type MsgCreateConditionalPayment struct {
	Condition		NativeConditionalPayment
	Payer			string
	ReservedLiquidity	string
	LinkedConditions	[]NativeConditionalPayment
	MessageHash		string
}

type MsgResolveConditionalPayment struct {
	Conditions		[]NativeConditionalPayment
	Preimage		string
	ProofRoot		string
	PromiseResultHash	string
	PaymentStateRoot	string
	CurrentHeight		uint64
	MessageHash		string
}

type MsgExpireConditionalPayment struct {
	Condition		NativeConditionalPayment
	Resolver		string
	RefundRouteRoot		string
	PaymentStateRoot	string
	CurrentHeight		uint64
	MessageHash		string
}

type MsgSettlePayment struct {
	Settlement		PaymentSettlement
	RouteCommitment		PaymentRouteCommitment
	ReceiptRoot		string
	PaymentStateRoot	string
	FinancialStateRoot	string
	MessageHash		string
}

func PaymentSettlementRules() []PaymentSettlementRule {
	return []PaymentSettlementRule{
		{Rule: PaymentSettlementRuleIntermediateOffchain, ConsensusEffect: "fast updates are allowed without committing every intermediate state"},
		{Rule: PaymentSettlementRuleDeterministicReplay, ConsensusEffect: "fraud proofs, newer states, and challenge outcomes reproduce on every validator"},
		{Rule: PaymentSettlementRuleFinancialStateWrite, ConsensusEffect: "balances, collateral, receipts, and payment roots reflect the final outcome"},
		{Rule: PaymentSettlementRuleConditionalProof, ConsensusEffect: "conditions cannot settle without preimage, timeout, promise, or linked proof satisfaction"},
		{Rule: PaymentSettlementRuleExpiryRelease, ConsensusEffect: "liquidity cannot remain locked after deterministic expiry handling"},
		{Rule: PaymentSettlementRuleCommittedRoutesOnly, ConsensusEffect: "only signed or reserved route commitments affect consensus settlement"},
	}
}

func PaymentMessageDescriptors() []PaymentMessageDescriptor {
	return []PaymentMessageDescriptor{
		paymentMessageDescriptor(PaymentMessageCreateIntent, "Create a payment intent before route reservation or settlement", "payer-authorization", "amount", "denom", "expiry", "idempotency-key"),
		paymentMessageDescriptor(PaymentMessageOpenChannel, "Lock collateral and create a payment channel", "participant-signatures", "collateral-availability", "challenge-period", "channel-id-uniqueness"),
		paymentMessageDescriptor(PaymentMessageUpdateChannel, "Submit or commit a newer signed channel state", "participant-signatures", "nonce-monotonicity", "balance-conservation", "condition-root-validity"),
		paymentMessageDescriptor(PaymentMessageCloseChannel, "Start or finalize channel close", "latest-state-proof", "challenge-window", "settlement-status", "collateral-conservation"),
		paymentMessageDescriptor(PaymentMessageDisputeChannel, "Challenge stale close with newer state or fraud proof", "newer-nonce", "valid-signatures", "evidence-hash", "active-challenge-period"),
		paymentMessageDescriptor(PaymentMessageCreateCondition, "Reserve value behind hash, time, promise, or route condition", "payer-authorization", "reserved-liquidity", "timeout-ordering", "route-id-validity"),
		paymentMessageDescriptor(PaymentMessageResolveCondition, "Resolve a condition with preimage, proof, or promise result", "preimage-or-proof-validity", "active-status", "linked-condition-rules", "amount-conservation"),
		paymentMessageDescriptor(PaymentMessageExpireCondition, "Expire unresolved condition and release reserved liquidity", "timeout-height-reached", "active-status", "refund-route-validity"),
		paymentMessageDescriptor(PaymentMessageSettlePayment, "Commit final payment settlement, refund, or failure", "settlement-proof", "route-commitment", "receipt-root", "payment-state-root-consistency"),
	}
}

func ValidatePaymentSettlementRulesAndMessages() error {
	seenRules := map[PaymentSettlementRuleID]struct{}{}
	for _, rule := range PaymentSettlementRules() {
		if rule.Rule == "" || strings.TrimSpace(rule.ConsensusEffect) == "" {
			return errors.New("payments settlement rule descriptor is incomplete")
		}
		if _, found := seenRules[rule.Rule]; found {
			return fmt.Errorf("payments duplicate settlement rule %s", rule.Rule)
		}
		seenRules[rule.Rule] = struct{}{}
	}
	seenMessages := map[PaymentMessageType]struct{}{}
	for _, descriptor := range PaymentMessageDescriptors() {
		if descriptor.Message == "" || strings.TrimSpace(descriptor.Purpose) == "" || len(descriptor.RequiredValidation) == 0 {
			return errors.New("payments message descriptor is incomplete")
		}
		if _, found := seenMessages[descriptor.Message]; found {
			return fmt.Errorf("payments duplicate message descriptor %s", descriptor.Message)
		}
		seenMessages[descriptor.Message] = struct{}{}
	}
	return nil
}

func BuildMsgCreatePaymentIntent(msg MsgCreatePaymentIntent) (MsgCreatePaymentIntent, PaymentIntent, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgCreatePaymentIntent{}, PaymentIntent{}, errors.New("payments create intent message hash must be empty before construction")
	}
	if err := msg.ValidateFormat(); err != nil {
		return MsgCreatePaymentIntent{}, PaymentIntent{}, err
	}
	msg.MessageHash = ComputeMsgCreatePaymentIntentHash(msg)
	intent, err := BuildPaymentIntent(PaymentIntent{
		PaymentID:	msg.PaymentID,
		IntentType:	PaymentIntentInitiate,
		Payer:		msg.Payer,
		Payee:		msg.Payee,
		TargetIdentity:	msg.TargetIdentity,
		Amount:		msg.Amount,
		MaxFee:		msg.MaxFee,
		ExpiryHeight:	msg.ExpiryHeight,
	})
	if err != nil {
		return MsgCreatePaymentIntent{}, PaymentIntent{}, err
	}
	return msg, intent, msg.Validate()
}

func (msg MsgCreatePaymentIntent) Normalize() MsgCreatePaymentIntent {
	msg.PaymentID = normalizeHash(msg.PaymentID)
	msg.Payer = strings.TrimSpace(msg.Payer)
	msg.Payee = strings.TrimSpace(msg.Payee)
	msg.TargetIdentity = strings.TrimSpace(msg.TargetIdentity)
	msg.Amount = strings.TrimSpace(msg.Amount)
	msg.Denom = normalizeAssetDenom(msg.Denom)
	msg.MaxFee = strings.TrimSpace(msg.MaxFee)
	msg.IdempotencyKey = normalizeHash(msg.IdempotencyKey)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgCreatePaymentIntent) ValidateFormat() error {
	msg = msg.Normalize()
	if _, err := FinancialPaymentIntentStateKey(msg.PaymentID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments create intent payer", msg.Payer); err != nil {
		return err
	}
	if msg.Payee == "" && msg.TargetIdentity == "" {
		return errors.New("payments create intent payee or identity is required")
	}
	if msg.Payee != "" {
		if err := addressing.ValidateUserAddress("payments create intent payee", msg.Payee); err != nil {
			return err
		}
	}
	if msg.TargetIdentity != "" && !strings.HasSuffix(msg.TargetIdentity, ".aet") {
		return errors.New("payments create intent identity must be .aet")
	}
	if err := validatePositiveInt("payments create intent amount", msg.Amount); err != nil {
		return err
	}
	if msg.Denom != NativeDenom {
		return fmt.Errorf("payments create intent denom must be %s", NativeDenom)
	}
	if err := validateNonNegativeInt("payments create intent max fee", msg.MaxFee); err != nil {
		return err
	}
	if msg.ExpiryHeight == 0 {
		return errors.New("payments create intent expiry height must be positive")
	}
	if err := ValidateHash("payments create intent idempotency key", msg.IdempotencyKey); err != nil {
		return err
	}
	if msg.MessageHash != "" {
		return ValidateHash("payments create intent message hash", msg.MessageHash)
	}
	return nil
}

func (msg MsgCreatePaymentIntent) Validate() error {
	msg = msg.Normalize()
	if err := msg.ValidateFormat(); err != nil {
		return err
	}
	if msg.MessageHash == "" {
		return errors.New("payments create intent message hash is required")
	}
	if expected := ComputeMsgCreatePaymentIntentHash(msg); msg.MessageHash != expected {
		return fmt.Errorf("payments create intent message hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildMsgOpenPaymentChannel(msg MsgOpenPaymentChannel, existing []PaymentChannel) (MsgOpenPaymentChannel, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgOpenPaymentChannel{}, errors.New("payments open channel message hash must be empty before construction")
	}
	if err := msg.ValidateWithExisting(existing); err != nil {
		return MsgOpenPaymentChannel{}, err
	}
	msg.MessageHash = ComputeMsgOpenPaymentChannelHash(msg)
	return msg, msg.ValidateWithExisting(existing)
}

func (msg MsgOpenPaymentChannel) Normalize() MsgOpenPaymentChannel {
	msg.Channel = msg.Channel.Normalize()
	msg.CollateralAvailable = strings.TrimSpace(msg.CollateralAvailable)
	msg.ParticipantSignatures = normalizeAddressSet(msg.ParticipantSignatures)
	msg.IdempotencyKey = normalizeHash(msg.IdempotencyKey)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgOpenPaymentChannel) ValidateWithExisting(existing []PaymentChannel) error {
	msg = msg.Normalize()
	if err := msg.Channel.Validate(); err != nil {
		return err
	}
	available, err := parsePositiveInt("payments open channel available collateral", msg.CollateralAvailable)
	if err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments open channel collateral", msg.Channel.Collateral)
	if err != nil {
		return err
	}
	if available.LT(collateral) {
		return errors.New("payments open channel collateral availability is insufficient")
	}
	if err := validateAddressSet("payments open channel participant signature", msg.ParticipantSignatures, len(msg.Channel.Participants), len(msg.Channel.Participants)); err != nil {
		return err
	}
	for _, participant := range msg.Channel.Participants {
		if !containsString(msg.ParticipantSignatures, participant) {
			return errors.New("payments open channel missing participant signature")
		}
	}
	for _, channel := range normalizePaymentChannels(existing) {
		if channel.ChannelID == msg.Channel.ChannelID {
			return errors.New("payments open channel id must be unique")
		}
	}
	if err := ValidateHash("payments open channel idempotency key", msg.IdempotencyKey); err != nil {
		return err
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments open channel message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgOpenPaymentChannelHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments open channel message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func BuildMsgUpdatePaymentChannel(msg MsgUpdatePaymentChannel, channel PaymentChannel) (MsgUpdatePaymentChannel, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgUpdatePaymentChannel{}, errors.New("payments update channel message hash must be empty before construction")
	}
	if err := msg.ValidateForChannel(channel); err != nil {
		return MsgUpdatePaymentChannel{}, err
	}
	msg.MessageHash = ComputeMsgUpdatePaymentChannelHash(msg)
	return msg, msg.ValidateForChannel(channel)
}

func (msg MsgUpdatePaymentChannel) Normalize() MsgUpdatePaymentChannel {
	msg.ChannelID = normalizeHash(msg.ChannelID)
	msg.Submitter = strings.TrimSpace(msg.Submitter)
	msg.SignedStateHash = normalizeHash(msg.SignedStateHash)
	msg.BalanceRoot = normalizeHash(msg.BalanceRoot)
	msg.ConditionRoot = normalizeHash(msg.ConditionRoot)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgUpdatePaymentChannel) ValidateForChannel(channel PaymentChannel) error {
	msg = msg.Normalize()
	channel = channel.Normalize()
	if err := channel.Validate(); err != nil {
		return err
	}
	if msg.ChannelID != channel.ChannelID {
		return errors.New("payments update channel id mismatch")
	}
	if err := addressing.ValidateUserAddress("payments update channel submitter", msg.Submitter); err != nil {
		return err
	}
	if !containsString(channel.Participants, msg.Submitter) {
		return errors.New("payments update channel submitter must be participant")
	}
	if msg.PreviousNonce != channel.LatestNonce || msg.NewNonce <= msg.PreviousNonce {
		return errors.New("payments update channel nonce monotonicity failed")
	}
	if err := ValidateHash("payments update channel signed state", msg.SignedStateHash); err != nil {
		return err
	}
	if err := ValidateHash("payments update channel balance root", msg.BalanceRoot); err != nil {
		return err
	}
	if err := ValidateHash("payments update channel condition root", msg.ConditionRoot); err != nil {
		return err
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments update channel message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgUpdatePaymentChannelHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments update channel message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func BuildMsgClosePaymentChannel(msg MsgClosePaymentChannel, channel PaymentChannel) (MsgClosePaymentChannel, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgClosePaymentChannel{}, errors.New("payments close channel message hash must be empty before construction")
	}
	if err := msg.ValidateForChannel(channel); err != nil {
		return MsgClosePaymentChannel{}, err
	}
	msg.MessageHash = ComputeMsgClosePaymentChannelHash(msg)
	return msg, msg.ValidateForChannel(channel)
}

func (msg MsgClosePaymentChannel) Normalize() MsgClosePaymentChannel {
	msg.PaymentID = normalizeHash(msg.PaymentID)
	msg.ChannelID = normalizeHash(msg.ChannelID)
	msg.LatestStateHash = normalizeHash(msg.LatestStateHash)
	msg.CollateralRoot = normalizeHash(msg.CollateralRoot)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgClosePaymentChannel) ValidateForChannel(channel PaymentChannel) error {
	msg = msg.Normalize()
	channel = channel.Normalize()
	if err := ValidateHash("payments close channel payment id", msg.PaymentID); err != nil {
		return err
	}
	if err := channel.Validate(); err != nil {
		return err
	}
	if msg.ChannelID != channel.ChannelID {
		return errors.New("payments close channel id mismatch")
	}
	if msg.LatestStateHash != channel.LatestStateHash {
		return errors.New("payments close channel latest state proof mismatch")
	}
	if msg.ChallengeStart == 0 || msg.ChallengeEnd <= msg.ChallengeStart {
		return errors.New("payments close channel challenge window is invalid")
	}
	if !IsNativePaymentSettlementStatus(msg.SettlementStatus) {
		return fmt.Errorf("unknown payments close channel status %q", msg.SettlementStatus)
	}
	if err := ValidateHash("payments close channel collateral root", msg.CollateralRoot); err != nil {
		return err
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments close channel message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgClosePaymentChannelHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments close channel message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func BuildMsgDisputePaymentChannel(msg MsgDisputePaymentChannel) (MsgDisputePaymentChannel, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgDisputePaymentChannel{}, errors.New("payments dispute channel message hash must be empty before construction")
	}
	if err := msg.Validate(); err != nil {
		return MsgDisputePaymentChannel{}, err
	}
	msg.MessageHash = ComputeMsgDisputePaymentChannelHash(msg)
	return msg, msg.Validate()
}

func (msg MsgDisputePaymentChannel) Normalize() MsgDisputePaymentChannel {
	msg.Dispute = msg.Dispute.Normalize()
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgDisputePaymentChannel) Validate() error {
	msg = msg.Normalize()
	if err := msg.Dispute.Validate(); err != nil {
		return err
	}
	if msg.StaleNonce == 0 || msg.NewerNonce <= msg.StaleNonce {
		return errors.New("payments dispute channel newer nonce is required")
	}
	if msg.CurrentHeight == 0 || msg.CurrentHeight > msg.Dispute.ChallengeEnd {
		return errors.New("payments dispute channel challenge period is inactive")
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments dispute channel message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgDisputePaymentChannelHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments dispute channel message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func BuildMsgCreateConditionalPayment(msg MsgCreateConditionalPayment) (MsgCreateConditionalPayment, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgCreateConditionalPayment{}, errors.New("payments create condition message hash must be empty before construction")
	}
	if err := msg.Validate(); err != nil {
		return MsgCreateConditionalPayment{}, err
	}
	msg.MessageHash = ComputeMsgCreateConditionalPaymentHash(msg)
	return msg, msg.Validate()
}

func (msg MsgCreateConditionalPayment) Normalize() MsgCreateConditionalPayment {
	msg.Condition = msg.Condition.Normalize()
	msg.Payer = strings.TrimSpace(msg.Payer)
	msg.ReservedLiquidity = strings.TrimSpace(msg.ReservedLiquidity)
	msg.LinkedConditions = normalizeNativeConditionalPayments(msg.LinkedConditions)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgCreateConditionalPayment) Validate() error {
	msg = msg.Normalize()
	if err := msg.Condition.Validate(); err != nil {
		return err
	}
	if msg.Condition.Payer != msg.Payer {
		return errors.New("payments create condition payer authorization mismatch")
	}
	reserved, err := parsePositiveInt("payments create condition reserved liquidity", msg.ReservedLiquidity)
	if err != nil {
		return err
	}
	amount, err := parsePositiveInt("payments create condition amount", msg.Condition.Amount)
	if err != nil {
		return err
	}
	if reserved.LT(amount) {
		return errors.New("payments create condition reserved liquidity is insufficient")
	}
	if msg.Condition.RouteID != "" {
		if err := ValidateHash("payments create condition route id", msg.Condition.RouteID); err != nil {
			return err
		}
	}
	if len(msg.LinkedConditions) > 0 {
		chain := append([]NativeConditionalPayment{msg.Condition}, msg.LinkedConditions...)
		if err := ValidateNativeConditionTimeoutOrdering(chain, DefaultTimeoutMargin); err != nil {
			return err
		}
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments create condition message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgCreateConditionalPaymentHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments create condition message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func BuildMsgResolveConditionalPayment(msg MsgResolveConditionalPayment) (MsgResolveConditionalPayment, []NativeConditionalPaymentOutcome, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgResolveConditionalPayment{}, nil, errors.New("payments resolve condition message hash must be empty before construction")
	}
	if err := msg.ValidateFormat(); err != nil {
		return MsgResolveConditionalPayment{}, nil, err
	}
	_, outcomes, err := ResolveNativeConditionalPaymentChain(msg.Conditions, msg.Preimage, msg.CurrentHeight, msg.PaymentStateRoot)
	if err != nil && msg.ProofRoot == "" && msg.PromiseResultHash == "" {
		return MsgResolveConditionalPayment{}, nil, err
	}
	msg.MessageHash = ComputeMsgResolveConditionalPaymentHash(msg)
	return msg, outcomes, msg.ValidateFormat()
}

func (msg MsgResolveConditionalPayment) Normalize() MsgResolveConditionalPayment {
	msg.Conditions = normalizeNativeConditionalPayments(msg.Conditions)
	msg.ProofRoot = normalizeOptionalHash(msg.ProofRoot)
	msg.PromiseResultHash = normalizeOptionalHash(msg.PromiseResultHash)
	msg.PaymentStateRoot = normalizeHash(msg.PaymentStateRoot)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgResolveConditionalPayment) ValidateFormat() error {
	msg = msg.Normalize()
	if len(msg.Conditions) == 0 {
		return errors.New("payments resolve condition requires conditions")
	}
	for _, condition := range msg.Conditions {
		if err := condition.Validate(); err != nil {
			return err
		}
		if condition.Status != NativeConditionalPaymentPending {
			return errors.New("payments resolve condition requires active status")
		}
	}
	if msg.Preimage == "" && msg.ProofRoot == "" && msg.PromiseResultHash == "" {
		return errors.New("payments resolve condition requires preimage, proof, or promise result")
	}
	if msg.ProofRoot != "" {
		if err := ValidateHash("payments resolve condition proof root", msg.ProofRoot); err != nil {
			return err
		}
	}
	if msg.PromiseResultHash != "" {
		if err := ValidateHash("payments resolve condition promise result", msg.PromiseResultHash); err != nil {
			return err
		}
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments resolve condition height must be positive")
	}
	if err := ValidateHash("payments resolve condition state root", msg.PaymentStateRoot); err != nil {
		return err
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments resolve condition message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgResolveConditionalPaymentHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments resolve condition message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func BuildMsgExpireConditionalPayment(msg MsgExpireConditionalPayment) (MsgExpireConditionalPayment, NativeConditionalPaymentOutcome, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgExpireConditionalPayment{}, NativeConditionalPaymentOutcome{}, errors.New("payments expire condition message hash must be empty before construction")
	}
	if err := msg.ValidateFormat(); err != nil {
		return MsgExpireConditionalPayment{}, NativeConditionalPaymentOutcome{}, err
	}
	_, outcome, err := ExpireNativeConditionalPayment(msg.Condition, msg.Resolver, msg.CurrentHeight, msg.PaymentStateRoot)
	if err != nil {
		return MsgExpireConditionalPayment{}, NativeConditionalPaymentOutcome{}, err
	}
	msg.MessageHash = ComputeMsgExpireConditionalPaymentHash(msg)
	return msg, outcome, msg.ValidateFormat()
}

func (msg MsgExpireConditionalPayment) Normalize() MsgExpireConditionalPayment {
	msg.Condition = msg.Condition.Normalize()
	msg.Resolver = strings.TrimSpace(msg.Resolver)
	msg.RefundRouteRoot = normalizeHash(msg.RefundRouteRoot)
	msg.PaymentStateRoot = normalizeHash(msg.PaymentStateRoot)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgExpireConditionalPayment) ValidateFormat() error {
	msg = msg.Normalize()
	if err := msg.Condition.Validate(); err != nil {
		return err
	}
	if msg.Condition.Status != NativeConditionalPaymentPending {
		return errors.New("payments expire condition requires active status")
	}
	if err := addressing.ValidateUserAddress("payments expire condition resolver", msg.Resolver); err != nil {
		return err
	}
	if msg.CurrentHeight <= msg.Condition.TimeoutHeight {
		return errors.New("payments expire condition timeout height has not been reached")
	}
	if err := ValidateHash("payments expire condition refund route", msg.RefundRouteRoot); err != nil {
		return err
	}
	if err := ValidateHash("payments expire condition state root", msg.PaymentStateRoot); err != nil {
		return err
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments expire condition message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgExpireConditionalPaymentHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments expire condition message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func BuildMsgSettlePayment(msg MsgSettlePayment, state FinancialZonePaymentState) (MsgSettlePayment, error) {
	msg = msg.Normalize()
	if msg.MessageHash != "" {
		return MsgSettlePayment{}, errors.New("payments settle message hash must be empty before construction")
	}
	if err := msg.ValidateAgainstState(state); err != nil {
		return MsgSettlePayment{}, err
	}
	msg.MessageHash = ComputeMsgSettlePaymentHash(msg)
	return msg, msg.ValidateAgainstState(state)
}

func (msg MsgSettlePayment) Normalize() MsgSettlePayment {
	msg.Settlement = msg.Settlement.Normalize()
	msg.RouteCommitment = msg.RouteCommitment.Normalize()
	msg.ReceiptRoot = normalizeHash(msg.ReceiptRoot)
	msg.PaymentStateRoot = normalizeHash(msg.PaymentStateRoot)
	msg.FinancialStateRoot = normalizeHash(msg.FinancialStateRoot)
	msg.MessageHash = normalizeOptionalHash(msg.MessageHash)
	return msg
}

func (msg MsgSettlePayment) ValidateAgainstState(state FinancialZonePaymentState) error {
	msg = msg.Normalize()
	state = state.Normalize()
	if err := msg.Settlement.Validate(); err != nil {
		return err
	}
	if err := validatePaymentRouteCommitments([]PaymentRouteCommitment{msg.RouteCommitment}); err != nil {
		return err
	}
	if msg.Settlement.RouteID != "" && msg.RouteCommitment.RouteID != msg.Settlement.RouteID {
		return errors.New("payments settle route commitment mismatch")
	}
	if err := ValidateHash("payments settle receipt root", msg.ReceiptRoot); err != nil {
		return err
	}
	if msg.ReceiptRoot != state.ReceiptRoot {
		return errors.New("payments settle receipt root mismatch")
	}
	if msg.PaymentStateRoot != state.PaymentRoot {
		return errors.New("payments settle payment state root mismatch")
	}
	if msg.FinancialStateRoot != state.PaymentRoot {
		return errors.New("payments settle financial state root mismatch")
	}
	if msg.MessageHash != "" {
		if err := ValidateHash("payments settle message hash", msg.MessageHash); err != nil {
			return err
		}
		if expected := ComputeMsgSettlePaymentHash(msg); msg.MessageHash != expected {
			return fmt.Errorf("payments settle message hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func ValidateFinalSettlementWritesFinancialZoneState(before, after FinancialZonePaymentState, settlement PaymentSettlement) error {
	before = before.Normalize()
	after = after.Normalize()
	settlement = settlement.Normalize()
	if before.PaymentRoot == "" || after.PaymentRoot == "" {
		return errors.New("payments final settlement requires committed payment roots")
	}
	if before.PaymentRoot == after.PaymentRoot {
		return errors.New("payments final settlement must change Financial Zone payment root")
	}
	found := false
	for _, existing := range after.Settlements {
		if existing.PaymentID == settlement.PaymentID && existing.SettlementHash == settlement.SettlementHash {
			found = true
			break
		}
	}
	if !found {
		return errors.New("payments final settlement was not written to Financial Zone state")
	}
	return after.Validate()
}

func ComputeMsgCreatePaymentIntentHash(msg MsgCreatePaymentIntent) string {
	msg = msg.Normalize()
	return HashParts("aetra-msg-create-payment-intent-v1", msg.PaymentID, msg.Payer, msg.Payee, msg.TargetIdentity, msg.Amount, msg.Denom, msg.MaxFee, fmt.Sprintf("%020d", msg.ExpiryHeight), msg.IdempotencyKey)
}

func ComputeMsgOpenPaymentChannelHash(msg MsgOpenPaymentChannel) string {
	msg = msg.Normalize()
	parts := []string{"aetra-msg-open-payment-channel-v1", msg.Channel.ChannelRoot, msg.CollateralAvailable, msg.IdempotencyKey}
	parts = append(parts, msg.ParticipantSignatures...)
	return HashParts(parts...)
}

func ComputeMsgUpdatePaymentChannelHash(msg MsgUpdatePaymentChannel) string {
	msg = msg.Normalize()
	return HashParts("aetra-msg-update-payment-channel-v1", msg.ChannelID, msg.Submitter, fmt.Sprintf("%020d", msg.PreviousNonce), fmt.Sprintf("%020d", msg.NewNonce), msg.SignedStateHash, msg.BalanceRoot, msg.ConditionRoot)
}

func ComputeMsgClosePaymentChannelHash(msg MsgClosePaymentChannel) string {
	msg = msg.Normalize()
	return HashParts("aetra-msg-close-payment-channel-v1", msg.PaymentID, msg.ChannelID, msg.LatestStateHash, fmt.Sprintf("%020d", msg.ChallengeStart), fmt.Sprintf("%020d", msg.ChallengeEnd), string(msg.SettlementStatus), msg.CollateralRoot)
}

func ComputeMsgDisputePaymentChannelHash(msg MsgDisputePaymentChannel) string {
	msg = msg.Normalize()
	return HashParts("aetra-msg-dispute-payment-channel-v1", msg.Dispute.DisputeRoot, fmt.Sprintf("%020d", msg.StaleNonce), fmt.Sprintf("%020d", msg.NewerNonce), fmt.Sprintf("%020d", msg.CurrentHeight))
}

func ComputeMsgCreateConditionalPaymentHash(msg MsgCreateConditionalPayment) string {
	msg = msg.Normalize()
	parts := []string{"aetra-msg-create-conditional-payment-v1", msg.Condition.ConditionRoot, msg.Payer, msg.ReservedLiquidity}
	for _, condition := range msg.LinkedConditions {
		parts = append(parts, condition.ConditionRoot)
	}
	return HashParts(parts...)
}

func ComputeMsgResolveConditionalPaymentHash(msg MsgResolveConditionalPayment) string {
	msg = msg.Normalize()
	parts := []string{"aetra-msg-resolve-conditional-payment-v1", msg.Preimage, msg.ProofRoot, msg.PromiseResultHash, msg.PaymentStateRoot, fmt.Sprintf("%020d", msg.CurrentHeight)}
	for _, condition := range msg.Conditions {
		parts = append(parts, condition.ConditionRoot)
	}
	return HashParts(parts...)
}

func ComputeMsgExpireConditionalPaymentHash(msg MsgExpireConditionalPayment) string {
	msg = msg.Normalize()
	return HashParts("aetra-msg-expire-conditional-payment-v1", msg.Condition.ConditionRoot, msg.Resolver, msg.RefundRouteRoot, msg.PaymentStateRoot, fmt.Sprintf("%020d", msg.CurrentHeight))
}

func ComputeMsgSettlePaymentHash(msg MsgSettlePayment) string {
	msg = msg.Normalize()
	return HashParts("aetra-msg-settle-payment-v1", msg.Settlement.SettlementHash, msg.RouteCommitment.RouteID, msg.RouteCommitment.CommitmentHash, fmt.Sprintf("%t", msg.RouteCommitment.Signed), fmt.Sprintf("%t", msg.RouteCommitment.Reserved), msg.ReceiptRoot, msg.PaymentStateRoot, msg.FinancialStateRoot)
}

func paymentMessageDescriptor(message PaymentMessageType, purpose string, validations ...string) PaymentMessageDescriptor {
	sort.Strings(validations)
	return PaymentMessageDescriptor{Message: message, Purpose: purpose, RequiredValidation: append([]string{}, validations...)}
}
