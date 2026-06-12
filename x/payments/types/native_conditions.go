package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type NativeConditionalPaymentStatus string

const (
	NativeConditionalPaymentPending		NativeConditionalPaymentStatus	= "pending"
	NativeConditionalPaymentResolved	NativeConditionalPaymentStatus	= "resolved"
	NativeConditionalPaymentTimedOut	NativeConditionalPaymentStatus	= "timed_out"
	NativeConditionalPaymentRefunded	NativeConditionalPaymentStatus	= "refunded"
	NativeConditionalPaymentFailed		NativeConditionalPaymentStatus	= "failed"
	NativeConditionalPaymentSettled		NativeConditionalPaymentStatus	= "settled"
)

type NativeConditionalPayment struct {
	ConditionID			string
	Payer				string
	Payee				string
	Amount				string
	HashLock			string
	TimeoutHeight			uint64
	RouteID				string
	NextConditionIDOptional		string
	PreviousConditionIDOptional	string
	Status				NativeConditionalPaymentStatus
	ConditionRoot			string
}

type NativeConditionalPaymentOutcome struct {
	ConditionID		string
	Status			NativeConditionalPaymentStatus
	Recipient		string
	Amount			string
	PreimageHash		string
	ResolvedHeight		uint64
	PaymentStateRoot	string
	OutcomeRoot		string
}

type CrossZonePaymentRoutingInput struct {
	SourceAccount		string
	TargetAccount		string
	TargetIdentity		string
	Amount			string
	MaxFee			string
	ExpiryHeight		uint64
	RoutePolicy		RoutePolicy
	LiquidityHints		[]PaymentRouteBalance
	RouteCommitment		PaymentRouteCommitment
	ReservationRoot		string
	UnifiedMessageRoot	string
	FinancialFallbackRoot	string
	RoutingRoot		string
}

type CrossZonePaymentMessage struct {
	MessageID		string
	SourceZoneID		string
	DestinationZoneID	string
	SourceShardID		uint32
	DestinationShardID	uint32
	PayloadType		string
	RouteID			string
	RouteCommitmentHash	string
	PaymentStateRoot	string
	UnifiedMessageRoot	string
	ExpiryHeight		uint64
	MessageHash		string
}

func NewNativeConditionalPaymentFromPromise(promise ConditionalPromise) (NativeConditionalPayment, error) {
	promise = promise.Normalize()
	if err := promise.ValidateBasic(); err != nil {
		return NativeConditionalPayment{}, err
	}
	return BuildNativeConditionalPayment(NativeConditionalPayment{
		ConditionID:			promise.PromiseID,
		Payer:				promise.Source,
		Payee:				promise.Destination,
		Amount:				promise.Amount,
		HashLock:			promise.HashLock,
		TimeoutHeight:			promise.TimeoutHeight,
		RouteID:			promise.RouteIDOptional,
		NextConditionIDOptional:	promise.NextPromiseIDOptional,
		PreviousConditionIDOptional:	promise.PreviousPromiseIDOptional,
		Status:				NativeConditionalPaymentPending,
	})
}

func BuildNativeConditionalPayment(condition NativeConditionalPayment) (NativeConditionalPayment, error) {
	condition = condition.Normalize()
	if condition.ConditionRoot != "" {
		return NativeConditionalPayment{}, errors.New("payments native condition root must be empty before construction")
	}
	if err := condition.ValidateFormat(); err != nil {
		return NativeConditionalPayment{}, err
	}
	condition.ConditionRoot = ComputeNativeConditionalPaymentRoot(condition)
	return condition, condition.Validate()
}

func (condition NativeConditionalPayment) Normalize() NativeConditionalPayment {
	condition.ConditionID = normalizeHash(condition.ConditionID)
	condition.Payer = strings.TrimSpace(condition.Payer)
	condition.Payee = strings.TrimSpace(condition.Payee)
	condition.Amount = strings.TrimSpace(condition.Amount)
	condition.HashLock = normalizeOptionalHash(condition.HashLock)
	condition.RouteID = normalizeOptionalHash(condition.RouteID)
	condition.NextConditionIDOptional = normalizeOptionalHash(condition.NextConditionIDOptional)
	condition.PreviousConditionIDOptional = normalizeOptionalHash(condition.PreviousConditionIDOptional)
	if condition.Status == "" {
		condition.Status = NativeConditionalPaymentPending
	}
	condition.ConditionRoot = normalizeOptionalHash(condition.ConditionRoot)
	return condition
}

func (condition NativeConditionalPayment) ValidateFormat() error {
	condition = condition.Normalize()
	if err := ValidateHash("payments native condition id", condition.ConditionID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments native condition payer", condition.Payer); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments native condition payee", condition.Payee); err != nil {
		return err
	}
	if condition.Payer == condition.Payee {
		return errors.New("payments native condition parties must differ")
	}
	if err := validatePositiveInt("payments native condition amount", condition.Amount); err != nil {
		return err
	}
	if condition.HashLock != "" {
		if err := ValidateHash("payments native condition hash lock", condition.HashLock); err != nil {
			return err
		}
	}
	if condition.TimeoutHeight == 0 {
		return errors.New("payments native condition timeout height must be positive")
	}
	if condition.RouteID != "" {
		if err := ValidateHash("payments native condition route id", condition.RouteID); err != nil {
			return err
		}
	}
	if condition.NextConditionIDOptional != "" {
		if err := ValidateHash("payments native condition next id", condition.NextConditionIDOptional); err != nil {
			return err
		}
	}
	if condition.PreviousConditionIDOptional != "" {
		if err := ValidateHash("payments native condition previous id", condition.PreviousConditionIDOptional); err != nil {
			return err
		}
	}
	if condition.NextConditionIDOptional == condition.ConditionID || condition.PreviousConditionIDOptional == condition.ConditionID {
		return errors.New("payments native condition cannot link to itself")
	}
	if !IsNativeConditionalPaymentStatus(condition.Status) {
		return fmt.Errorf("unknown payments native condition status %q", condition.Status)
	}
	if condition.ConditionRoot != "" {
		return ValidateHash("payments native condition root", condition.ConditionRoot)
	}
	return nil
}

func (condition NativeConditionalPayment) Validate() error {
	condition = condition.Normalize()
	if err := condition.ValidateFormat(); err != nil {
		return err
	}
	if condition.ConditionRoot == "" {
		return errors.New("payments native condition root is required")
	}
	if expected := ComputeNativeConditionalPaymentRoot(condition); condition.ConditionRoot != expected {
		return fmt.Errorf("payments native condition root mismatch: expected %s", expected)
	}
	return nil
}

func ResolveNativeConditionalPaymentChain(conditions []NativeConditionalPayment, preimage string, currentHeight uint64, paymentStateRoot string) ([]NativeConditionalPayment, []NativeConditionalPaymentOutcome, error) {
	if currentHeight == 0 {
		return nil, nil, errors.New("payments native condition resolve height must be positive")
	}
	paymentStateRoot = normalizeHash(paymentStateRoot)
	if err := ValidateHash("payments native condition payment state root", paymentStateRoot); err != nil {
		return nil, nil, err
	}
	ordered, err := orderNativeConditionChain(conditions)
	if err != nil {
		return nil, nil, err
	}
	if len(ordered) == 0 {
		return nil, nil, errors.New("payments native condition chain is required")
	}
	preimageHash := HashParts(preimage)
	out := make([]NativeConditionalPayment, len(ordered))
	outcomes := make([]NativeConditionalPaymentOutcome, len(ordered))
	for i, condition := range ordered {
		if condition.Status != NativeConditionalPaymentPending {
			return nil, nil, errors.New("payments native condition must be pending to resolve")
		}
		if currentHeight > condition.TimeoutHeight {
			return nil, nil, errors.New("payments native condition has timed out")
		}
		if condition.HashLock != "" && condition.HashLock != preimageHash {
			return nil, nil, errors.New("payments native condition preimage mismatch")
		}
		condition.Status = NativeConditionalPaymentResolved
		condition.ConditionRoot = ""
		condition, err = BuildNativeConditionalPayment(condition)
		if err != nil {
			return nil, nil, err
		}
		outcome, err := BuildNativeConditionalPaymentOutcome(NativeConditionalPaymentOutcome{
			ConditionID:		condition.ConditionID,
			Status:			NativeConditionalPaymentResolved,
			Recipient:		condition.Payee,
			Amount:			condition.Amount,
			PreimageHash:		preimageHash,
			ResolvedHeight:		currentHeight,
			PaymentStateRoot:	paymentStateRoot,
		})
		if err != nil {
			return nil, nil, err
		}
		out[i] = condition
		outcomes[i] = outcome
	}
	return out, outcomes, nil
}

func ExpireNativeConditionalPayment(condition NativeConditionalPayment, resolver string, currentHeight uint64, paymentStateRoot string) (NativeConditionalPayment, NativeConditionalPaymentOutcome, error) {
	condition = condition.Normalize()
	resolver = strings.TrimSpace(resolver)
	paymentStateRoot = normalizeHash(paymentStateRoot)
	if err := condition.Validate(); err != nil {
		return NativeConditionalPayment{}, NativeConditionalPaymentOutcome{}, err
	}
	if err := addressing.ValidateUserAddress("payments native condition timeout resolver", resolver); err != nil {
		return NativeConditionalPayment{}, NativeConditionalPaymentOutcome{}, err
	}
	if currentHeight <= condition.TimeoutHeight {
		return NativeConditionalPayment{}, NativeConditionalPaymentOutcome{}, errors.New("payments native condition timeout height has not passed")
	}
	if condition.Status != NativeConditionalPaymentPending {
		return NativeConditionalPayment{}, NativeConditionalPaymentOutcome{}, errors.New("payments native condition must be pending to expire")
	}
	if err := ValidateHash("payments native condition payment state root", paymentStateRoot); err != nil {
		return NativeConditionalPayment{}, NativeConditionalPaymentOutcome{}, err
	}
	condition.Status = NativeConditionalPaymentRefunded
	condition.ConditionRoot = ""
	condition, err := BuildNativeConditionalPayment(condition)
	if err != nil {
		return NativeConditionalPayment{}, NativeConditionalPaymentOutcome{}, err
	}
	outcome, err := BuildNativeConditionalPaymentOutcome(NativeConditionalPaymentOutcome{
		ConditionID:		condition.ConditionID,
		Status:			NativeConditionalPaymentRefunded,
		Recipient:		condition.Payer,
		Amount:			condition.Amount,
		ResolvedHeight:		currentHeight,
		PaymentStateRoot:	paymentStateRoot,
	})
	if err != nil {
		return NativeConditionalPayment{}, NativeConditionalPaymentOutcome{}, err
	}
	return condition, outcome, nil
}

func ValidateNativeConditionTimeoutOrdering(conditions []NativeConditionalPayment, margin uint64) error {
	ordered, err := orderNativeConditionChain(conditions)
	if err != nil {
		return err
	}
	if len(ordered) < 2 {
		return nil
	}
	for i := 0; i < len(ordered)-1; i++ {
		upstream := ordered[i]
		downstream := ordered[i+1]
		if upstream.TimeoutHeight <= downstream.TimeoutHeight+margin {
			return errors.New("payments native condition timeout ordering must protect intermediaries")
		}
	}
	return nil
}

func BuildNativeConditionalPaymentOutcome(outcome NativeConditionalPaymentOutcome) (NativeConditionalPaymentOutcome, error) {
	outcome = outcome.Normalize()
	if outcome.OutcomeRoot != "" {
		return NativeConditionalPaymentOutcome{}, errors.New("payments native condition outcome root must be empty before construction")
	}
	if err := outcome.ValidateFormat(); err != nil {
		return NativeConditionalPaymentOutcome{}, err
	}
	outcome.OutcomeRoot = ComputeNativeConditionalPaymentOutcomeRoot(outcome)
	return outcome, outcome.Validate()
}

func (outcome NativeConditionalPaymentOutcome) Normalize() NativeConditionalPaymentOutcome {
	outcome.ConditionID = normalizeHash(outcome.ConditionID)
	outcome.Recipient = strings.TrimSpace(outcome.Recipient)
	outcome.Amount = strings.TrimSpace(outcome.Amount)
	outcome.PreimageHash = normalizeOptionalHash(outcome.PreimageHash)
	outcome.PaymentStateRoot = normalizeHash(outcome.PaymentStateRoot)
	outcome.OutcomeRoot = normalizeOptionalHash(outcome.OutcomeRoot)
	return outcome
}

func (outcome NativeConditionalPaymentOutcome) ValidateFormat() error {
	outcome = outcome.Normalize()
	if err := ValidateHash("payments native condition outcome id", outcome.ConditionID); err != nil {
		return err
	}
	if !IsNativeConditionalPaymentStatus(outcome.Status) || outcome.Status == NativeConditionalPaymentPending {
		return fmt.Errorf("unknown payments native condition outcome status %q", outcome.Status)
	}
	if err := addressing.ValidateUserAddress("payments native condition outcome recipient", outcome.Recipient); err != nil {
		return err
	}
	if err := validatePositiveInt("payments native condition outcome amount", outcome.Amount); err != nil {
		return err
	}
	if outcome.PreimageHash != "" {
		if err := ValidateHash("payments native condition outcome preimage", outcome.PreimageHash); err != nil {
			return err
		}
	}
	if outcome.ResolvedHeight == 0 {
		return errors.New("payments native condition outcome height must be positive")
	}
	if err := ValidateHash("payments native condition outcome state root", outcome.PaymentStateRoot); err != nil {
		return err
	}
	if outcome.OutcomeRoot != "" {
		return ValidateHash("payments native condition outcome root", outcome.OutcomeRoot)
	}
	return nil
}

func (outcome NativeConditionalPaymentOutcome) Validate() error {
	outcome = outcome.Normalize()
	if err := outcome.ValidateFormat(); err != nil {
		return err
	}
	if outcome.OutcomeRoot == "" {
		return errors.New("payments native condition outcome root is required")
	}
	if expected := ComputeNativeConditionalPaymentOutcomeRoot(outcome); outcome.OutcomeRoot != expected {
		return fmt.Errorf("payments native condition outcome root mismatch: expected %s", expected)
	}
	return nil
}

func BuildCrossZonePaymentRoutingInput(input CrossZonePaymentRoutingInput) (CrossZonePaymentRoutingInput, error) {
	input = input.Normalize()
	if input.RoutingRoot != "" {
		return CrossZonePaymentRoutingInput{}, errors.New("payments cross-zone route root must be empty before construction")
	}
	if err := input.ValidateFormat(); err != nil {
		return CrossZonePaymentRoutingInput{}, err
	}
	input.RoutingRoot = ComputeCrossZonePaymentRoutingInputRoot(input)
	return input, input.Validate()
}

func (input CrossZonePaymentRoutingInput) Normalize() CrossZonePaymentRoutingInput {
	input.SourceAccount = strings.TrimSpace(input.SourceAccount)
	input.TargetAccount = strings.TrimSpace(input.TargetAccount)
	input.TargetIdentity = strings.TrimSpace(input.TargetIdentity)
	input.Amount = strings.TrimSpace(input.Amount)
	input.MaxFee = strings.TrimSpace(input.MaxFee)
	input.RoutePolicy = input.RoutePolicy.Normalize()
	for i := range input.LiquidityHints {
		input.LiquidityHints[i] = input.LiquidityHints[i].Normalize()
	}
	sort.SliceStable(input.LiquidityHints, func(i, j int) bool {
		return input.LiquidityHints[i].Participant < input.LiquidityHints[j].Participant
	})
	input.RouteCommitment = input.RouteCommitment.Normalize()
	input.ReservationRoot = normalizeOptionalHash(input.ReservationRoot)
	input.UnifiedMessageRoot = normalizeHash(input.UnifiedMessageRoot)
	input.FinancialFallbackRoot = normalizeHash(input.FinancialFallbackRoot)
	input.RoutingRoot = normalizeOptionalHash(input.RoutingRoot)
	return input
}

func (input CrossZonePaymentRoutingInput) ValidateFormat() error {
	input = input.Normalize()
	if err := addressing.ValidateUserAddress("payments cross-zone source account", input.SourceAccount); err != nil {
		return err
	}
	if input.TargetAccount == "" && input.TargetIdentity == "" {
		return errors.New("payments cross-zone target account or identity is required")
	}
	if input.TargetAccount != "" {
		if err := addressing.ValidateUserAddress("payments cross-zone target account", input.TargetAccount); err != nil {
			return err
		}
	}
	if input.TargetIdentity != "" && !strings.HasSuffix(input.TargetIdentity, ".aet") {
		return errors.New("payments cross-zone target identity must be .aet")
	}
	if err := validatePositiveInt("payments cross-zone amount", input.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments cross-zone max fee", input.MaxFee); err != nil {
		return err
	}
	if input.ExpiryHeight == 0 {
		return errors.New("payments cross-zone expiry height must be positive")
	}
	if err := input.RoutePolicy.Validate(); err != nil {
		return err
	}
	for _, hint := range input.LiquidityHints {
		if err := addressing.ValidateUserAddress("payments cross-zone liquidity participant", hint.Participant); err != nil {
			return err
		}
		if err := validateNonNegativeInt("payments cross-zone liquidity hint", hint.Available); err != nil {
			return err
		}
	}
	if !input.RouteCommitment.Signed && !input.RouteCommitment.Reserved {
		return errors.New("payments cross-zone route commitment must be signed or reserved")
	}
	if input.RouteCommitment.Signed {
		if err := addressing.ValidateUserAddress("payments cross-zone route committer", input.RouteCommitment.Committer); err != nil {
			return err
		}
	}
	if err := ValidateHash("payments cross-zone route id", input.RouteCommitment.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("payments cross-zone commitment hash", input.RouteCommitment.CommitmentHash); err != nil {
		return err
	}
	if input.RouteCommitment.ExpiresHeight == 0 || input.RouteCommitment.ExpiresHeight > input.ExpiryHeight {
		return errors.New("payments cross-zone route commitment expiry exceeds route expiry")
	}
	if input.ReservationRoot != "" {
		if err := ValidateHash("payments cross-zone reservation root", input.ReservationRoot); err != nil {
			return err
		}
	}
	if err := ValidateHash("payments cross-zone unified message root", input.UnifiedMessageRoot); err != nil {
		return err
	}
	if err := ValidateHash("payments cross-zone financial fallback root", input.FinancialFallbackRoot); err != nil {
		return err
	}
	if input.RoutingRoot != "" {
		return ValidateHash("payments cross-zone route root", input.RoutingRoot)
	}
	return nil
}

func ValidateCrossZonePaymentRouteSettlement(input CrossZonePaymentRoutingInput, route MsgPaymentRoute, currentHeight uint64) error {
	input = input.Normalize()
	route = route.Normalize()
	if currentHeight == 0 {
		return errors.New("payments cross-zone route settlement height must be positive")
	}
	if err := input.Validate(); err != nil {
		return err
	}
	if err := route.ValidateBasic(); err != nil {
		return err
	}
	if input.SourceAccount != route.Payer {
		return errors.New("payments cross-zone route source mismatch")
	}
	if input.TargetAccount != "" && input.TargetAccount != route.Payee {
		return errors.New("payments cross-zone route target mismatch")
	}
	if input.Amount != route.Amount || input.MaxFee != route.MaxFee || input.ExpiryHeight != route.ExpiryHeight {
		return errors.New("payments cross-zone route value or expiry mismatch")
	}
	if err := input.RouteCommitment.ValidateForRoute(route, currentHeight); err != nil {
		return err
	}
	if input.RouteCommitment.Reserved && input.ReservationRoot == "" {
		return errors.New("payments cross-zone reserved route requires reservation root")
	}
	if currentHeight > input.ExpiryHeight {
		return errors.New("payments cross-zone route has expired")
	}
	return nil
}

func (input CrossZonePaymentRoutingInput) Validate() error {
	input = input.Normalize()
	if err := input.ValidateFormat(); err != nil {
		return err
	}
	if input.RoutingRoot == "" {
		return errors.New("payments cross-zone route root is required")
	}
	if expected := ComputeCrossZonePaymentRoutingInputRoot(input); input.RoutingRoot != expected {
		return fmt.Errorf("payments cross-zone route root mismatch: expected %s", expected)
	}
	return nil
}

func BuildCrossZonePaymentMessage(message CrossZonePaymentMessage) (CrossZonePaymentMessage, error) {
	message = message.Normalize()
	if message.MessageHash != "" {
		return CrossZonePaymentMessage{}, errors.New("payments cross-zone message hash must be empty before construction")
	}
	if err := message.ValidateFormat(); err != nil {
		return CrossZonePaymentMessage{}, err
	}
	message.MessageHash = ComputeCrossZonePaymentMessageHash(message)
	if message.MessageID == "" {
		message.MessageID = HashParts("cross-zone-payment-message-id", message.MessageHash)
	}
	return message, message.Validate()
}

func (message CrossZonePaymentMessage) Normalize() CrossZonePaymentMessage {
	message.MessageID = normalizeOptionalHash(message.MessageID)
	message.SourceZoneID = strings.TrimSpace(message.SourceZoneID)
	message.DestinationZoneID = strings.TrimSpace(message.DestinationZoneID)
	message.PayloadType = strings.TrimSpace(message.PayloadType)
	message.RouteID = normalizeHash(message.RouteID)
	message.RouteCommitmentHash = normalizeHash(message.RouteCommitmentHash)
	message.PaymentStateRoot = normalizeHash(message.PaymentStateRoot)
	message.UnifiedMessageRoot = normalizeHash(message.UnifiedMessageRoot)
	message.MessageHash = normalizeOptionalHash(message.MessageHash)
	return message
}

func (message CrossZonePaymentMessage) ValidateFormat() error {
	message = message.Normalize()
	if message.MessageID != "" {
		if err := ValidateHash("payments cross-zone message id", message.MessageID); err != nil {
			return err
		}
	}
	if err := validatePaymentRoutingToken("payments cross-zone source zone", message.SourceZoneID); err != nil {
		return err
	}
	if err := validatePaymentRoutingToken("payments cross-zone destination zone", message.DestinationZoneID); err != nil {
		return err
	}
	if message.SourceZoneID == message.DestinationZoneID {
		return errors.New("payments cross-zone message requires distinct zones")
	}
	if err := validatePaymentRoutingToken("payments cross-zone payload type", message.PayloadType); err != nil {
		return err
	}
	if err := ValidateHash("payments cross-zone message route id", message.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("payments cross-zone route commitment", message.RouteCommitmentHash); err != nil {
		return err
	}
	if err := ValidateHash("payments cross-zone payment state root", message.PaymentStateRoot); err != nil {
		return err
	}
	if err := ValidateHash("payments cross-zone unified message root", message.UnifiedMessageRoot); err != nil {
		return err
	}
	if message.ExpiryHeight == 0 {
		return errors.New("payments cross-zone message expiry height must be positive")
	}
	if message.MessageHash != "" {
		return ValidateHash("payments cross-zone message hash", message.MessageHash)
	}
	return nil
}

func (message CrossZonePaymentMessage) Validate() error {
	message = message.Normalize()
	if err := message.ValidateFormat(); err != nil {
		return err
	}
	if message.MessageHash == "" {
		return errors.New("payments cross-zone message hash is required")
	}
	if expected := ComputeCrossZonePaymentMessageHash(message); message.MessageHash != expected {
		return fmt.Errorf("payments cross-zone message hash mismatch: expected %s", expected)
	}
	if message.MessageID != HashParts("cross-zone-payment-message-id", message.MessageHash) {
		return errors.New("payments cross-zone message id mismatch")
	}
	return nil
}

func ComputeNativeConditionalPaymentRoot(condition NativeConditionalPayment) string {
	condition = condition.Normalize()
	return HashParts(
		"aetra-native-conditional-payment-v1",
		condition.ConditionID,
		condition.Payer,
		condition.Payee,
		condition.Amount,
		condition.HashLock,
		fmt.Sprintf("%020d", condition.TimeoutHeight),
		condition.RouteID,
		condition.NextConditionIDOptional,
		condition.PreviousConditionIDOptional,
		string(condition.Status),
	)
}

func ComputeNativeConditionalPaymentOutcomeRoot(outcome NativeConditionalPaymentOutcome) string {
	outcome = outcome.Normalize()
	return HashParts(
		"aetra-native-conditional-payment-outcome-v1",
		outcome.ConditionID,
		string(outcome.Status),
		outcome.Recipient,
		outcome.Amount,
		outcome.PreimageHash,
		fmt.Sprintf("%020d", outcome.ResolvedHeight),
		outcome.PaymentStateRoot,
	)
}

func ComputeCrossZonePaymentRoutingInputRoot(input CrossZonePaymentRoutingInput) string {
	input = input.Normalize()
	parts := []string{
		"aetra-cross-zone-payment-routing-input-v1",
		input.SourceAccount,
		input.TargetAccount,
		input.TargetIdentity,
		input.Amount,
		input.MaxFee,
		fmt.Sprintf("%020d", input.ExpiryHeight),
		routePolicyHash(input.RoutePolicy),
		input.RouteCommitment.RouteID,
		input.RouteCommitment.Committer,
		input.RouteCommitment.CommitmentHash,
		fmt.Sprintf("%t", input.RouteCommitment.Signed),
		fmt.Sprintf("%t", input.RouteCommitment.Reserved),
		fmt.Sprintf("%020d", input.RouteCommitment.ExpiresHeight),
		input.ReservationRoot,
		input.UnifiedMessageRoot,
		input.FinancialFallbackRoot,
	}
	for _, hint := range input.LiquidityHints {
		parts = append(parts, hint.Participant, hint.Available)
	}
	return HashParts(parts...)
}

func ComputeCrossZonePaymentMessageHash(message CrossZonePaymentMessage) string {
	message = message.Normalize()
	return HashParts(
		"aetra-cross-zone-payment-message-v1",
		message.SourceZoneID,
		message.DestinationZoneID,
		fmt.Sprintf("%020d", uint64(message.SourceShardID)),
		fmt.Sprintf("%020d", uint64(message.DestinationShardID)),
		message.PayloadType,
		message.RouteID,
		message.RouteCommitmentHash,
		message.PaymentStateRoot,
		message.UnifiedMessageRoot,
		fmt.Sprintf("%020d", message.ExpiryHeight),
	)
}

func IsNativeConditionalPaymentStatus(status NativeConditionalPaymentStatus) bool {
	switch status {
	case NativeConditionalPaymentPending,
		NativeConditionalPaymentResolved,
		NativeConditionalPaymentTimedOut,
		NativeConditionalPaymentRefunded,
		NativeConditionalPaymentFailed,
		NativeConditionalPaymentSettled:
		return true
	default:
		return false
	}
}

func orderNativeConditionChain(conditions []NativeConditionalPayment) ([]NativeConditionalPayment, error) {
	if len(conditions) == 0 {
		return nil, nil
	}
	byID := make(map[string]NativeConditionalPayment, len(conditions))
	var head string
	for _, condition := range conditions {
		condition = condition.Normalize()
		if err := condition.Validate(); err != nil {
			return nil, err
		}
		if _, found := byID[condition.ConditionID]; found {
			return nil, errors.New("payments native condition chain has duplicate condition")
		}
		byID[condition.ConditionID] = condition
		if condition.PreviousConditionIDOptional == "" {
			if head != "" {
				return nil, errors.New("payments native condition chain has multiple heads")
			}
			head = condition.ConditionID
		}
	}
	if head == "" {
		return nil, errors.New("payments native condition chain head is missing")
	}
	ordered := make([]NativeConditionalPayment, 0, len(conditions))
	seen := map[string]struct{}{}
	for currentID := head; currentID != ""; {
		if _, found := seen[currentID]; found {
			return nil, errors.New("payments native condition chain contains cycle")
		}
		current, found := byID[currentID]
		if !found {
			return nil, errors.New("payments native condition chain link missing")
		}
		seen[currentID] = struct{}{}
		ordered = append(ordered, current)
		nextID := current.NextConditionIDOptional
		if nextID != "" {
			next, found := byID[nextID]
			if !found {
				return nil, errors.New("payments native condition chain next link missing")
			}
			if next.PreviousConditionIDOptional != current.ConditionID {
				return nil, errors.New("payments native condition chain previous link mismatch")
			}
		}
		currentID = nextID
	}
	if len(ordered) != len(conditions) {
		return nil, errors.New("payments native condition chain is disconnected")
	}
	return ordered, nil
}
