package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	FinancialPaymentsPrefix			= "financial/payments"
	FinancialPaymentIntentsPrefix		= FinancialPaymentsPrefix + "/intents"
	FinancialPaymentChannelsPrefix		= FinancialPaymentsPrefix + "/channels"
	FinancialPaymentConditionsPrefix	= FinancialPaymentsPrefix + "/conditions"
	FinancialPaymentRoutesPrefix		= FinancialPaymentsPrefix + "/routes"
	FinancialPaymentSettlementsPrefix	= FinancialPaymentsPrefix + "/settlements"
	FinancialPaymentDisputesPrefix		= FinancialPaymentsPrefix + "/disputes"
	FinancialPaymentReceiptsPrefix		= FinancialPaymentsPrefix + "/receipts"
	FinancialPaymentProofsPrefix		= FinancialPaymentsPrefix + "/proofs"
	FinancialPaymentFeesPrefix		= FinancialPaymentsPrefix + "/fees"
	FinancialPaymentMessagesPrefix		= FinancialPaymentsPrefix + "/messages"
	FinancialPaymentCanonicalPrefix		= FinancialPaymentsPrefix + "/canonical"
)

type PaymentImplementationPriority string

const (
	PaymentImplementationP0	PaymentImplementationPriority	= "P0"
	PaymentImplementationP1	PaymentImplementationPriority	= "P1"
)

type PaymentImplementationTaskID string

const (
	PaymentTaskFinancialZoneModule		PaymentImplementationTaskID	= "financial-zone-payments-module"
	PaymentTaskCanonicalEnvelope		PaymentImplementationTaskID	= "payment-envelope-canonical-encoding"
	PaymentTaskChannelCollateralEscrow	PaymentImplementationTaskID	= "channel-collateral-escrow"
	PaymentTaskSettlementStateProof		PaymentImplementationTaskID	= "settlement-state-and-proof"
	PaymentTaskConditionalLocks		PaymentImplementationTaskID	= "conditional-hash-time-locks"
	PaymentTaskVirtualChannelProof		PaymentImplementationTaskID	= "virtual-channel-proof-model"
	PaymentTaskRouteCommitmentFormat	PaymentImplementationTaskID	= "route-commitment-format"
	PaymentTaskSettlementProofQuery		PaymentImplementationTaskID	= "settlement-proof-query"
	PaymentTaskChannelDisputeReplay		PaymentImplementationTaskID	= "channel-dispute-replay"
	PaymentTaskRouteFeeAccounting		PaymentImplementationTaskID	= "route-fee-accounting"
	PaymentTaskCrossZoneMessages		PaymentImplementationTaskID	= "cross-zone-settlement-messages"
	PaymentTaskReceiptRoot			PaymentImplementationTaskID	= "payment-receipt-root"
	PaymentTaskAdversarialTests		PaymentImplementationTaskID	= "adversarial-payment-tests"
)

type PaymentImplementationTask struct {
	Priority		PaymentImplementationPriority
	TaskID			PaymentImplementationTaskID
	Task			string
	Target			string
	AcceptanceCriteria	[]string
}

type PaymentIntentType string

const (
	PaymentIntentInitiate		PaymentIntentType	= "initiate_payment"
	PaymentIntentReserveRoute	PaymentIntentType	= "reserve_route"
	PaymentIntentSettle		PaymentIntentType	= "settle_payment"
)

type PaymentDisputeStatus string

const (
	PaymentDisputeOpen	PaymentDisputeStatus	= "open"
	PaymentDisputeAccepted	PaymentDisputeStatus	= "accepted"
	PaymentDisputeRejected	PaymentDisputeStatus	= "rejected"
	PaymentDisputeResolved	PaymentDisputeStatus	= "resolved"
)

type PaymentEnvelopeObjectType string

const (
	PaymentEnvelopeIntent		PaymentEnvelopeObjectType	= "intent"
	PaymentEnvelopeRoute		PaymentEnvelopeObjectType	= "route"
	PaymentEnvelopeChannel		PaymentEnvelopeObjectType	= "channel_update"
	PaymentEnvelopeCondition	PaymentEnvelopeObjectType	= "condition"
	PaymentEnvelopeSettlement	PaymentEnvelopeObjectType	= "settlement"
	PaymentEnvelopeReceipt		PaymentEnvelopeObjectType	= "receipt"
	PaymentEnvelopeDispute		PaymentEnvelopeObjectType	= "dispute"
	PaymentEnvelopeFee		PaymentEnvelopeObjectType	= "fee"
	PaymentEnvelopeMessage		PaymentEnvelopeObjectType	= "message"
)

type CrossZonePaymentMessageType string

const (
	CrossZonePaymentMessageRoute	CrossZonePaymentMessageType	= "payment_route"
	CrossZonePaymentMessageReserve	CrossZonePaymentMessageType	= "payment_reserve"
	CrossZonePaymentMessageSettle	CrossZonePaymentMessageType	= "payment_settle"
	CrossZonePaymentMessageRefund	CrossZonePaymentMessageType	= "payment_refund"
	CrossZonePaymentMessageBounce	CrossZonePaymentMessageType	= "payment_bounce"
	CrossZonePaymentMessageReceipt	CrossZonePaymentMessageType	= "payment_receipt"
)

type PaymentIntent struct {
	PaymentID	string
	IntentType	PaymentIntentType
	Payer		string
	Payee		string
	TargetIdentity	string
	Amount		string
	MaxFee		string
	RouteIDOptional	string
	ExpiryHeight	uint64
	IntentHash	string
}

type PaymentSettlement struct {
	PaymentID	string
	ChannelID	string
	RouteID		string
	FinalStateHash	string
	ReceiptHash	string
	RefundHash	string
	TimeoutHash	string
	CloseStatus	NativePaymentSettlementStatus
	ProofRoot	string
	SettledHeight	uint64
	SettlementHash	string
}

type PaymentDispute struct {
	DisputeID	string
	PaymentID	string
	ChannelID	string
	FraudProofHash	string
	StaleStateHash	string
	NewerStateHash	string
	SubmittedBy	string
	OpenedHeight	uint64
	ChallengeEnd	uint64
	Status		PaymentDisputeStatus
	DisputeRoot	string
}

type PaymentFeeAccountingRecord struct {
	FeeID			string
	RouteID			string
	ForwardingFee		string
	RouteFee		string
	ReserveFee		string
	SettlementGasFee	string
	RecordedHeight		uint64
	FeeRoot			string
}

type PaymentEnvelopeCanonicalRecord struct {
	ObjectType	PaymentEnvelopeObjectType
	ObjectID	string
	StateKey	string
	ObjectHash	string
	EncodingVersion	byte
	EnvelopeHash	string
}

type PaymentSettlementProofQuery struct {
	PaymentID	string
	ProofType	SettlementProofType
}

type PaymentSettlementProofQueryResponse struct {
	Proof	SettlementProof
	Found	bool
}

type FinancialZonePaymentState struct {
	Height			uint64
	Intents			[]PaymentIntent
	Channels		[]PaymentChannel
	Conditions		[]NativeConditionalPayment
	Routes			[]PaymentRouteCommitment
	Settlements		[]PaymentSettlement
	Disputes		[]PaymentDispute
	Receipts		[]PaymentReceipt
	Proofs			[]SettlementProof
	Fees			[]PaymentFeeAccountingRecord
	Messages		[]CrossZonePaymentMessage
	CanonicalEnvelopes	[]PaymentEnvelopeCanonicalRecord
	IntentRoot		string
	ChannelRoot		string
	ConditionRoot		string
	RouteRoot		string
	SettlementRoot		string
	DisputeRoot		string
	ReceiptRoot		string
	ProofRoot		string
	FeeRoot			string
	MessageRoot		string
	CanonicalRoot		string
	PaymentRoot		string
}

func PaymentImplementationTasks() []PaymentImplementationTask {
	return []PaymentImplementationTask{
		paymentTask(PaymentImplementationP0, PaymentTaskFinancialZoneModule, "Implement x/payments under Financial Zone", "payments module and Financial Zone adapter", []string{"commit-channels", "commit-conditions", "commit-routes", "commit-reservations", "commit-settlements", "commit-receipts"}),
		paymentTask(PaymentImplementationP0, PaymentTaskCanonicalEnvelope, "Add payment envelope canonical encoding", "payment codec", []string{"hash-payment-intents", "hash-routes", "hash-channel-updates", "hash-conditions", "hash-settlements", "hash-receipts"}),
		paymentTask(PaymentImplementationP0, PaymentTaskChannelCollateralEscrow, "Add channel collateral escrow in Financial Zone", "Financial Zone payment state", []string{"reject-overallocated-balances", "preserve-value-conservation", "lock-collateral-by-channel"}),
		paymentTask(PaymentImplementationP0, PaymentTaskSettlementStateProof, "Add settlement state and proof", "payment settlement state", []string{"include-final-state", "include-receipt", "include-route", "include-close-status", "include-proof-path"}),
		paymentTask(PaymentImplementationP0, PaymentTaskConditionalLocks, "Add conditional hash and time locks", "condition state machine", []string{"resolve-hash-locks", "resolve-timeouts", "resolve-chained-conditions", "resolve-promise-conditions"}),
		paymentTask(PaymentImplementationP0, PaymentTaskVirtualChannelProof, "Add virtual channel proof model", "virtual channel state", []string{"prove-underlying-channel-capacity", "prove-route-commitments"}),
		paymentTask(PaymentImplementationP0, PaymentTaskRouteCommitmentFormat, "Add route commitment format", "payment routing codec", []string{"bind-hops", "bind-capacity", "bind-fees", "bind-expiry", "bind-participants", "bind-signatures-or-reservations"}),
		paymentTask(PaymentImplementationP1, PaymentTaskSettlementProofQuery, "Add settlement proof query", "payment proof API", []string{"prove-latest-state", "prove-close-status", "prove-settlement", "prove-refund", "prove-timeout"}),
		paymentTask(PaymentImplementationP1, PaymentTaskChannelDisputeReplay, "Add channel dispute replay", "dispute state machine", []string{"newer-state-supersedes-stale-close", "emit-dispute-receipt"}),
		paymentTask(PaymentImplementationP1, PaymentTaskRouteFeeAccounting, "Add route fee accounting", "fee accounting", []string{"bound-route-fees", "bound-forwarding-fees", "bound-reserve-fees", "bound-settlement-gas"}),
		paymentTask(PaymentImplementationP1, PaymentTaskCrossZoneMessages, "Add cross-zone settlement messages", "unified message layer", []string{"route-message", "reserve-message", "settle-message", "refund-message", "bounce-message", "receipt-message"}),
		paymentTask(PaymentImplementationP1, PaymentTaskReceiptRoot, "Add payment receipt root", "payment receipt state", []string{"receipt-per-mutation", "include-payment-receipt-root", "include-zone-receipt-root"}),
		paymentTask(PaymentImplementationP1, PaymentTaskAdversarialTests, "Add stale close, timeout, hash-lock, and route failure tests", "payment tests", []string{"preserve-value", "preserve-roots", "preserve-receipts", "preserve-deterministic-replay"}),
	}
}

func ValidatePaymentImplementationTasks() error {
	seen := map[PaymentImplementationTaskID]struct{}{}
	for _, task := range PaymentImplementationTasks() {
		if task.Priority == "" || task.TaskID == "" || strings.TrimSpace(task.Task) == "" || strings.TrimSpace(task.Target) == "" {
			return errors.New("payments implementation task is incomplete")
		}
		if _, found := seen[task.TaskID]; found {
			return fmt.Errorf("payments duplicate implementation task %s", task.TaskID)
		}
		if len(task.AcceptanceCriteria) == 0 {
			return fmt.Errorf("payments implementation task %s requires acceptance criteria", task.TaskID)
		}
		seen[task.TaskID] = struct{}{}
	}
	return nil
}

func FinancialPaymentIntentStateKey(paymentID string) (string, error) {
	if err := ValidateHash("payments financial intent id", normalizeHash(paymentID)); err != nil {
		return "", err
	}
	return FinancialPaymentIntentsPrefix + "/" + normalizeHash(paymentID), nil
}

func FinancialPaymentChannelStateKey(channelID string) (string, error) {
	if err := ValidateHash("payments financial channel id", normalizeHash(channelID)); err != nil {
		return "", err
	}
	return FinancialPaymentChannelsPrefix + "/" + normalizeHash(channelID), nil
}

func FinancialPaymentConditionStateKey(conditionID string) (string, error) {
	if err := ValidateHash("payments financial condition id", normalizeHash(conditionID)); err != nil {
		return "", err
	}
	return FinancialPaymentConditionsPrefix + "/" + normalizeHash(conditionID), nil
}

func FinancialPaymentRouteStateKey(routeID string) (string, error) {
	if err := ValidateHash("payments financial route id", normalizeHash(routeID)); err != nil {
		return "", err
	}
	return FinancialPaymentRoutesPrefix + "/" + normalizeHash(routeID), nil
}

func FinancialPaymentSettlementStateKey(paymentID string) (string, error) {
	if err := ValidateHash("payments financial settlement payment id", normalizeHash(paymentID)); err != nil {
		return "", err
	}
	return FinancialPaymentSettlementsPrefix + "/" + normalizeHash(paymentID), nil
}

func FinancialPaymentDisputeStateKey(disputeID string) (string, error) {
	if err := ValidateHash("payments financial dispute id", normalizeHash(disputeID)); err != nil {
		return "", err
	}
	return FinancialPaymentDisputesPrefix + "/" + normalizeHash(disputeID), nil
}

func BuildPaymentIntent(intent PaymentIntent) (PaymentIntent, error) {
	intent = intent.Normalize()
	if intent.IntentHash != "" {
		return PaymentIntent{}, errors.New("payments intent hash must be empty before construction")
	}
	if err := intent.ValidateFormat(); err != nil {
		return PaymentIntent{}, err
	}
	intent.IntentHash = ComputePaymentIntentHash(intent)
	return intent, intent.Validate()
}

func (intent PaymentIntent) Normalize() PaymentIntent {
	intent.PaymentID = normalizeHash(intent.PaymentID)
	intent.Payer = strings.TrimSpace(intent.Payer)
	intent.Payee = strings.TrimSpace(intent.Payee)
	intent.TargetIdentity = strings.TrimSpace(intent.TargetIdentity)
	intent.Amount = strings.TrimSpace(intent.Amount)
	intent.MaxFee = strings.TrimSpace(intent.MaxFee)
	intent.RouteIDOptional = normalizeOptionalHash(intent.RouteIDOptional)
	if intent.IntentType == "" {
		intent.IntentType = PaymentIntentInitiate
	}
	intent.IntentHash = normalizeOptionalHash(intent.IntentHash)
	return intent
}

func (intent PaymentIntent) ValidateFormat() error {
	intent = intent.Normalize()
	if _, err := FinancialPaymentIntentStateKey(intent.PaymentID); err != nil {
		return err
	}
	if !IsPaymentIntentType(intent.IntentType) {
		return fmt.Errorf("unknown payments intent type %q", intent.IntentType)
	}
	if err := addressing.ValidateUserAddress("payments intent payer", intent.Payer); err != nil {
		return err
	}
	if intent.Payee == "" && intent.TargetIdentity == "" {
		return errors.New("payments intent payee or target identity is required")
	}
	if intent.Payee != "" {
		if err := addressing.ValidateUserAddress("payments intent payee", intent.Payee); err != nil {
			return err
		}
	}
	if intent.TargetIdentity != "" && !strings.HasSuffix(intent.TargetIdentity, ".aet") {
		return errors.New("payments intent target identity must be .aet")
	}
	if err := validatePositiveInt("payments intent amount", intent.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments intent max fee", intent.MaxFee); err != nil {
		return err
	}
	if intent.RouteIDOptional != "" {
		if err := ValidateHash("payments intent route id", intent.RouteIDOptional); err != nil {
			return err
		}
	}
	if intent.ExpiryHeight == 0 {
		return errors.New("payments intent expiry height must be positive")
	}
	if intent.IntentHash != "" {
		return ValidateHash("payments intent hash", intent.IntentHash)
	}
	return nil
}

func (intent PaymentIntent) Validate() error {
	intent = intent.Normalize()
	if err := intent.ValidateFormat(); err != nil {
		return err
	}
	if intent.IntentHash == "" {
		return errors.New("payments intent hash is required")
	}
	if expected := ComputePaymentIntentHash(intent); intent.IntentHash != expected {
		return fmt.Errorf("payments intent hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildPaymentSettlement(settlement PaymentSettlement) (PaymentSettlement, error) {
	settlement = settlement.Normalize()
	if settlement.SettlementHash != "" {
		return PaymentSettlement{}, errors.New("payments settlement hash must be empty before construction")
	}
	if err := settlement.ValidateFormat(); err != nil {
		return PaymentSettlement{}, err
	}
	settlement.SettlementHash = ComputePaymentSettlementStateHash(settlement)
	return settlement, settlement.Validate()
}

func (settlement PaymentSettlement) Normalize() PaymentSettlement {
	settlement.PaymentID = normalizeHash(settlement.PaymentID)
	settlement.ChannelID = normalizeOptionalHash(settlement.ChannelID)
	settlement.RouteID = normalizeOptionalHash(settlement.RouteID)
	settlement.FinalStateHash = normalizeHash(settlement.FinalStateHash)
	settlement.ReceiptHash = normalizeHash(settlement.ReceiptHash)
	settlement.RefundHash = normalizeOptionalHash(settlement.RefundHash)
	settlement.TimeoutHash = normalizeOptionalHash(settlement.TimeoutHash)
	settlement.ProofRoot = normalizeHash(settlement.ProofRoot)
	if settlement.CloseStatus == "" {
		settlement.CloseStatus = NativePaymentSettlementSettled
	}
	settlement.SettlementHash = normalizeOptionalHash(settlement.SettlementHash)
	return settlement
}

func (settlement PaymentSettlement) ValidateFormat() error {
	settlement = settlement.Normalize()
	if _, err := FinancialPaymentSettlementStateKey(settlement.PaymentID); err != nil {
		return err
	}
	if settlement.ChannelID == "" && settlement.RouteID == "" {
		return errors.New("payments settlement requires channel or route id")
	}
	if settlement.ChannelID != "" {
		if err := ValidateHash("payments settlement channel id", settlement.ChannelID); err != nil {
			return err
		}
	}
	if settlement.RouteID != "" {
		if err := ValidateHash("payments settlement route id", settlement.RouteID); err != nil {
			return err
		}
	}
	if err := ValidateHash("payments settlement final state", settlement.FinalStateHash); err != nil {
		return err
	}
	if err := ValidateHash("payments settlement receipt hash", settlement.ReceiptHash); err != nil {
		return err
	}
	if settlement.RefundHash != "" {
		if err := ValidateHash("payments settlement refund hash", settlement.RefundHash); err != nil {
			return err
		}
	}
	if settlement.TimeoutHash != "" {
		if err := ValidateHash("payments settlement timeout hash", settlement.TimeoutHash); err != nil {
			return err
		}
	}
	if !IsNativePaymentSettlementStatus(settlement.CloseStatus) {
		return fmt.Errorf("unknown payments settlement close status %q", settlement.CloseStatus)
	}
	if err := ValidateHash("payments settlement proof root", settlement.ProofRoot); err != nil {
		return err
	}
	if settlement.SettledHeight == 0 {
		return errors.New("payments settlement height must be positive")
	}
	if settlement.SettlementHash != "" {
		return ValidateHash("payments settlement hash", settlement.SettlementHash)
	}
	return nil
}

func (settlement PaymentSettlement) Validate() error {
	settlement = settlement.Normalize()
	if err := settlement.ValidateFormat(); err != nil {
		return err
	}
	if settlement.SettlementHash == "" {
		return errors.New("payments settlement hash is required")
	}
	if expected := ComputePaymentSettlementStateHash(settlement); settlement.SettlementHash != expected {
		return fmt.Errorf("payments settlement hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildPaymentDispute(dispute PaymentDispute) (PaymentDispute, error) {
	dispute = dispute.Normalize()
	if dispute.DisputeRoot != "" {
		return PaymentDispute{}, errors.New("payments dispute root must be empty before construction")
	}
	if err := dispute.ValidateFormat(); err != nil {
		return PaymentDispute{}, err
	}
	dispute.DisputeRoot = ComputePaymentDisputeRoot(dispute)
	return dispute, dispute.Validate()
}

func (dispute PaymentDispute) Normalize() PaymentDispute {
	dispute.DisputeID = normalizeHash(dispute.DisputeID)
	dispute.PaymentID = normalizeHash(dispute.PaymentID)
	dispute.ChannelID = normalizeHash(dispute.ChannelID)
	dispute.FraudProofHash = normalizeHash(dispute.FraudProofHash)
	dispute.StaleStateHash = normalizeHash(dispute.StaleStateHash)
	dispute.NewerStateHash = normalizeHash(dispute.NewerStateHash)
	dispute.SubmittedBy = strings.TrimSpace(dispute.SubmittedBy)
	if dispute.Status == "" {
		dispute.Status = PaymentDisputeOpen
	}
	dispute.DisputeRoot = normalizeOptionalHash(dispute.DisputeRoot)
	return dispute
}

func (dispute PaymentDispute) ValidateFormat() error {
	dispute = dispute.Normalize()
	if _, err := FinancialPaymentDisputeStateKey(dispute.DisputeID); err != nil {
		return err
	}
	if err := ValidateHash("payments dispute payment id", dispute.PaymentID); err != nil {
		return err
	}
	if err := ValidateHash("payments dispute channel id", dispute.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments dispute fraud proof", dispute.FraudProofHash); err != nil {
		return err
	}
	if err := ValidateHash("payments dispute stale state", dispute.StaleStateHash); err != nil {
		return err
	}
	if err := ValidateHash("payments dispute newer state", dispute.NewerStateHash); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments dispute submitter", dispute.SubmittedBy); err != nil {
		return err
	}
	if dispute.OpenedHeight == 0 || dispute.ChallengeEnd <= dispute.OpenedHeight {
		return errors.New("payments dispute height range is invalid")
	}
	if !IsPaymentDisputeStatus(dispute.Status) {
		return fmt.Errorf("unknown payments dispute status %q", dispute.Status)
	}
	if dispute.DisputeRoot != "" {
		return ValidateHash("payments dispute root", dispute.DisputeRoot)
	}
	return nil
}

func (dispute PaymentDispute) Validate() error {
	dispute = dispute.Normalize()
	if err := dispute.ValidateFormat(); err != nil {
		return err
	}
	if dispute.DisputeRoot == "" {
		return errors.New("payments dispute root is required")
	}
	if expected := ComputePaymentDisputeRoot(dispute); dispute.DisputeRoot != expected {
		return fmt.Errorf("payments dispute root mismatch: expected %s", expected)
	}
	return nil
}

func BuildPaymentFeeAccountingRecord(record PaymentFeeAccountingRecord) (PaymentFeeAccountingRecord, error) {
	record = record.Normalize()
	if record.FeeRoot != "" {
		return PaymentFeeAccountingRecord{}, errors.New("payments fee root must be empty before construction")
	}
	if err := record.ValidateFormat(); err != nil {
		return PaymentFeeAccountingRecord{}, err
	}
	record.FeeRoot = ComputePaymentFeeAccountingRoot(record)
	return record, record.Validate()
}

func (record PaymentFeeAccountingRecord) Normalize() PaymentFeeAccountingRecord {
	record.FeeID = normalizeHash(record.FeeID)
	record.RouteID = normalizeHash(record.RouteID)
	record.ForwardingFee = strings.TrimSpace(record.ForwardingFee)
	record.RouteFee = strings.TrimSpace(record.RouteFee)
	record.ReserveFee = strings.TrimSpace(record.ReserveFee)
	record.SettlementGasFee = strings.TrimSpace(record.SettlementGasFee)
	record.FeeRoot = normalizeOptionalHash(record.FeeRoot)
	return record
}

func (record PaymentFeeAccountingRecord) ValidateFormat() error {
	record = record.Normalize()
	if err := ValidateHash("payments fee id", record.FeeID); err != nil {
		return err
	}
	if err := ValidateHash("payments fee route id", record.RouteID); err != nil {
		return err
	}
	for _, value := range []struct{ field, amount string }{
		{"payments forwarding fee", record.ForwardingFee},
		{"payments route fee", record.RouteFee},
		{"payments reserve fee", record.ReserveFee},
		{"payments settlement gas fee", record.SettlementGasFee},
	} {
		if err := validateNonNegativeInt(value.field, value.amount); err != nil {
			return err
		}
	}
	if record.RecordedHeight == 0 {
		return errors.New("payments fee recorded height must be positive")
	}
	if record.FeeRoot != "" {
		return ValidateHash("payments fee root", record.FeeRoot)
	}
	return nil
}

func (record PaymentFeeAccountingRecord) Validate() error {
	record = record.Normalize()
	if err := record.ValidateFormat(); err != nil {
		return err
	}
	if record.FeeRoot == "" {
		return errors.New("payments fee root is required")
	}
	if expected := ComputePaymentFeeAccountingRoot(record); record.FeeRoot != expected {
		return fmt.Errorf("payments fee root mismatch: expected %s", expected)
	}
	return nil
}

func BuildPaymentEnvelopeCanonicalRecord(record PaymentEnvelopeCanonicalRecord) (PaymentEnvelopeCanonicalRecord, error) {
	record = record.Normalize()
	if record.EnvelopeHash != "" {
		return PaymentEnvelopeCanonicalRecord{}, errors.New("payments canonical envelope hash must be empty before construction")
	}
	if err := record.ValidateFormat(); err != nil {
		return PaymentEnvelopeCanonicalRecord{}, err
	}
	record.EnvelopeHash = ComputePaymentEnvelopeCanonicalHash(record)
	return record, record.Validate()
}

func (record PaymentEnvelopeCanonicalRecord) Normalize() PaymentEnvelopeCanonicalRecord {
	record.ObjectID = normalizeHash(record.ObjectID)
	record.StateKey = strings.TrimSpace(record.StateKey)
	record.ObjectHash = normalizeHash(record.ObjectHash)
	if record.EncodingVersion == 0 {
		record.EncodingVersion = CanonicalEncodingVersion
	}
	record.EnvelopeHash = normalizeOptionalHash(record.EnvelopeHash)
	return record
}

func (record PaymentEnvelopeCanonicalRecord) ValidateFormat() error {
	record = record.Normalize()
	if !IsPaymentEnvelopeObjectType(record.ObjectType) {
		return fmt.Errorf("unknown payments canonical envelope object type %q", record.ObjectType)
	}
	if err := ValidateHash("payments canonical envelope object id", record.ObjectID); err != nil {
		return err
	}
	if !strings.HasPrefix(record.StateKey, FinancialPaymentsPrefix+"/") {
		return errors.New("payments canonical envelope state key must be under financial payments prefix")
	}
	if err := ValidateHash("payments canonical envelope object hash", record.ObjectHash); err != nil {
		return err
	}
	if record.EncodingVersion != CanonicalEncodingVersion {
		return fmt.Errorf("payments canonical envelope unsupported encoding version %d", record.EncodingVersion)
	}
	if record.EnvelopeHash != "" {
		return ValidateHash("payments canonical envelope hash", record.EnvelopeHash)
	}
	return nil
}

func (record PaymentEnvelopeCanonicalRecord) Validate() error {
	record = record.Normalize()
	if err := record.ValidateFormat(); err != nil {
		return err
	}
	if record.EnvelopeHash == "" {
		return errors.New("payments canonical envelope hash is required")
	}
	if expected := ComputePaymentEnvelopeCanonicalHash(record); record.EnvelopeHash != expected {
		return fmt.Errorf("payments canonical envelope hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildFinancialZonePaymentState(state FinancialZonePaymentState) (FinancialZonePaymentState, error) {
	state = state.Normalize()
	if state.PaymentRoot != "" {
		return FinancialZonePaymentState{}, errors.New("payments financial state root must be empty before construction")
	}
	if state.Height == 0 {
		return FinancialZonePaymentState{}, errors.New("payments financial state height must be positive")
	}
	if err := state.ValidateObjects(); err != nil {
		return FinancialZonePaymentState{}, err
	}
	state.IntentRoot = ComputePaymentIntentSetRoot(state.Intents)
	state.ChannelRoot = ComputePaymentChannelSetRoot(state.Channels)
	state.ConditionRoot = ComputeNativeConditionalPaymentSetRoot(state.Conditions)
	state.RouteRoot = ComputePaymentRouteCommitmentSetRoot(state.Routes)
	state.SettlementRoot = ComputePaymentSettlementSetRoot(state.Settlements)
	state.DisputeRoot = ComputePaymentDisputeSetRoot(state.Disputes)
	state.ReceiptRoot = ComputeNativePaymentReceiptSetRoot(state.Receipts)
	state.ProofRoot = ComputeSettlementProofSetRoot(state.Proofs)
	state.FeeRoot = ComputePaymentFeeAccountingSetRoot(state.Fees)
	state.MessageRoot = ComputeCrossZonePaymentMessageSetRoot(state.Messages)
	state.CanonicalRoot = ComputePaymentEnvelopeCanonicalSetRoot(state.CanonicalEnvelopes)
	state.PaymentRoot = ComputeFinancialZonePaymentStateRoot(state)
	return state, state.Validate()
}

func (state FinancialZonePaymentState) Normalize() FinancialZonePaymentState {
	state.Intents = normalizePaymentIntents(state.Intents)
	state.Channels = normalizePaymentChannels(state.Channels)
	state.Conditions = normalizeNativeConditionalPayments(state.Conditions)
	state.Routes = normalizePaymentRouteCommitments(state.Routes)
	state.Settlements = normalizePaymentSettlements(state.Settlements)
	state.Disputes = normalizePaymentDisputes(state.Disputes)
	state.Receipts = normalizePaymentReceipts(state.Receipts)
	state.Proofs = normalizeSettlementProofs(state.Proofs)
	state.Fees = normalizePaymentFeeRecords(state.Fees)
	state.Messages = normalizeCrossZonePaymentMessages(state.Messages)
	state.CanonicalEnvelopes = normalizePaymentCanonicalEnvelopes(state.CanonicalEnvelopes)
	state.IntentRoot = normalizeOptionalHash(state.IntentRoot)
	state.ChannelRoot = normalizeOptionalHash(state.ChannelRoot)
	state.ConditionRoot = normalizeOptionalHash(state.ConditionRoot)
	state.RouteRoot = normalizeOptionalHash(state.RouteRoot)
	state.SettlementRoot = normalizeOptionalHash(state.SettlementRoot)
	state.DisputeRoot = normalizeOptionalHash(state.DisputeRoot)
	state.ReceiptRoot = normalizeOptionalHash(state.ReceiptRoot)
	state.ProofRoot = normalizeOptionalHash(state.ProofRoot)
	state.FeeRoot = normalizeOptionalHash(state.FeeRoot)
	state.MessageRoot = normalizeOptionalHash(state.MessageRoot)
	state.CanonicalRoot = normalizeOptionalHash(state.CanonicalRoot)
	state.PaymentRoot = normalizeOptionalHash(state.PaymentRoot)
	return state
}

func (state FinancialZonePaymentState) ValidateObjects() error {
	if err := validateUniquePaymentIntents(state.Intents); err != nil {
		return err
	}
	if err := validatePaymentChannels(state.Channels); err != nil {
		return err
	}
	if err := validateNativeConditionalPayments(state.Conditions); err != nil {
		return err
	}
	if err := validatePaymentRouteCommitments(state.Routes); err != nil {
		return err
	}
	if err := validatePaymentSettlements(state.Settlements); err != nil {
		return err
	}
	if err := validatePaymentDisputes(state.Disputes); err != nil {
		return err
	}
	if err := validatePaymentReceipts(state.Receipts); err != nil {
		return err
	}
	if err := validateSettlementProofs(state.Proofs); err != nil {
		return err
	}
	if err := validatePaymentFeeRecords(state.Fees); err != nil {
		return err
	}
	if err := validateCrossZonePaymentMessages(state.Messages); err != nil {
		return err
	}
	return validatePaymentCanonicalEnvelopes(state.CanonicalEnvelopes)
}

func (state FinancialZonePaymentState) Validate() error {
	state = state.Normalize()
	if state.Height == 0 {
		return errors.New("payments financial state height must be positive")
	}
	if err := state.ValidateObjects(); err != nil {
		return err
	}
	expectedRoots := map[string][2]string{
		"payments financial intent root":	{state.IntentRoot, ComputePaymentIntentSetRoot(state.Intents)},
		"payments financial channel root":	{state.ChannelRoot, ComputePaymentChannelSetRoot(state.Channels)},
		"payments financial condition root":	{state.ConditionRoot, ComputeNativeConditionalPaymentSetRoot(state.Conditions)},
		"payments financial route root":	{state.RouteRoot, ComputePaymentRouteCommitmentSetRoot(state.Routes)},
		"payments financial settlement root":	{state.SettlementRoot, ComputePaymentSettlementSetRoot(state.Settlements)},
		"payments financial dispute root":	{state.DisputeRoot, ComputePaymentDisputeSetRoot(state.Disputes)},
		"payments financial receipt root":	{state.ReceiptRoot, ComputeNativePaymentReceiptSetRoot(state.Receipts)},
		"payments financial proof root":	{state.ProofRoot, ComputeSettlementProofSetRoot(state.Proofs)},
		"payments financial fee root":		{state.FeeRoot, ComputePaymentFeeAccountingSetRoot(state.Fees)},
		"payments financial message root":	{state.MessageRoot, ComputeCrossZonePaymentMessageSetRoot(state.Messages)},
		"payments financial canonical root":	{state.CanonicalRoot, ComputePaymentEnvelopeCanonicalSetRoot(state.CanonicalEnvelopes)},
	}
	for field, roots := range expectedRoots {
		if roots[0] == "" {
			return fmt.Errorf("%s is required", field)
		}
		if roots[0] != roots[1] {
			return fmt.Errorf("%s mismatch: expected %s", field, roots[1])
		}
	}
	if state.PaymentRoot == "" {
		return errors.New("payments financial root is required")
	}
	if expected := ComputeFinancialZonePaymentStateRoot(state); state.PaymentRoot != expected {
		return fmt.Errorf("payments financial root mismatch: expected %s", expected)
	}
	return nil
}

func QueryPaymentSettlementProofFromState(state FinancialZonePaymentState, query PaymentSettlementProofQuery) (PaymentSettlementProofQueryResponse, error) {
	state = state.Normalize()
	query.PaymentID = normalizeHash(query.PaymentID)
	if err := state.Validate(); err != nil {
		return PaymentSettlementProofQueryResponse{}, err
	}
	if err := ValidateHash("payments settlement proof query payment id", query.PaymentID); err != nil {
		return PaymentSettlementProofQueryResponse{}, err
	}
	if !IsSettlementProofType(query.ProofType) {
		return PaymentSettlementProofQueryResponse{}, fmt.Errorf("unknown payments settlement proof query type %q", query.ProofType)
	}
	for _, settlement := range state.Settlements {
		if settlement.PaymentID != query.PaymentID {
			continue
		}
		for _, proof := range state.Proofs {
			if proof.ChannelID == settlement.ChannelID && proof.ProofType == query.ProofType {
				return PaymentSettlementProofQueryResponse{Proof: proof, Found: true}, nil
			}
		}
	}
	return PaymentSettlementProofQueryResponse{Found: false}, nil
}

func ComputePaymentIntentHash(intent PaymentIntent) string {
	intent = intent.Normalize()
	return HashParts("aetra-financial-payment-intent-v1", intent.PaymentID, string(intent.IntentType), intent.Payer, intent.Payee, intent.TargetIdentity, intent.Amount, intent.MaxFee, intent.RouteIDOptional, fmt.Sprintf("%020d", intent.ExpiryHeight))
}

func ComputePaymentSettlementStateHash(settlement PaymentSettlement) string {
	settlement = settlement.Normalize()
	return HashParts("aetra-financial-payment-settlement-v1", settlement.PaymentID, settlement.ChannelID, settlement.RouteID, settlement.FinalStateHash, settlement.ReceiptHash, settlement.RefundHash, settlement.TimeoutHash, string(settlement.CloseStatus), settlement.ProofRoot, fmt.Sprintf("%020d", settlement.SettledHeight))
}

func ComputePaymentDisputeRoot(dispute PaymentDispute) string {
	dispute = dispute.Normalize()
	return HashParts("aetra-financial-payment-dispute-v1", dispute.DisputeID, dispute.PaymentID, dispute.ChannelID, dispute.FraudProofHash, dispute.StaleStateHash, dispute.NewerStateHash, dispute.SubmittedBy, fmt.Sprintf("%020d", dispute.OpenedHeight), fmt.Sprintf("%020d", dispute.ChallengeEnd), string(dispute.Status))
}

func ComputePaymentFeeAccountingRoot(record PaymentFeeAccountingRecord) string {
	record = record.Normalize()
	return HashParts("aetra-financial-payment-fee-v1", record.FeeID, record.RouteID, record.ForwardingFee, record.RouteFee, record.ReserveFee, record.SettlementGasFee, fmt.Sprintf("%020d", record.RecordedHeight))
}

func ComputePaymentEnvelopeCanonicalHash(record PaymentEnvelopeCanonicalRecord) string {
	record = record.Normalize()
	return HashParts("aetra-financial-payment-canonical-envelope-v1", string(record.ObjectType), record.ObjectID, record.StateKey, record.ObjectHash, fmt.Sprintf("%03d", record.EncodingVersion))
}

func ComputePaymentIntentSetRoot(intents []PaymentIntent) string {
	parts := []string{"aetra-financial-payment-intent-root-v1"}
	for _, intent := range normalizePaymentIntents(intents) {
		parts = append(parts, intent.IntentHash)
	}
	return HashParts(parts...)
}

func ComputeNativeConditionalPaymentSetRoot(conditions []NativeConditionalPayment) string {
	parts := []string{"aetra-financial-native-condition-root-v1"}
	for _, condition := range normalizeNativeConditionalPayments(conditions) {
		parts = append(parts, condition.ConditionRoot)
	}
	return HashParts(parts...)
}

func ComputePaymentRouteCommitmentSetRoot(routes []PaymentRouteCommitment) string {
	parts := []string{"aetra-financial-payment-route-commitment-root-v1"}
	for _, route := range normalizePaymentRouteCommitments(routes) {
		parts = append(parts, route.RouteID, route.Committer, route.CommitmentHash, fmt.Sprintf("%t", route.Signed), fmt.Sprintf("%t", route.Reserved), fmt.Sprintf("%020d", route.ExpiresHeight))
	}
	return HashParts(parts...)
}

func ComputePaymentSettlementSetRoot(settlements []PaymentSettlement) string {
	parts := []string{"aetra-financial-payment-settlement-root-v1"}
	for _, settlement := range normalizePaymentSettlements(settlements) {
		parts = append(parts, settlement.SettlementHash)
	}
	return HashParts(parts...)
}

func ComputePaymentDisputeSetRoot(disputes []PaymentDispute) string {
	parts := []string{"aetra-financial-payment-dispute-root-v1"}
	for _, dispute := range normalizePaymentDisputes(disputes) {
		parts = append(parts, dispute.DisputeRoot)
	}
	return HashParts(parts...)
}

func ComputePaymentFeeAccountingSetRoot(records []PaymentFeeAccountingRecord) string {
	parts := []string{"aetra-financial-payment-fee-root-v1"}
	for _, record := range normalizePaymentFeeRecords(records) {
		parts = append(parts, record.FeeRoot)
	}
	return HashParts(parts...)
}

func ComputeCrossZonePaymentMessageSetRoot(messages []CrossZonePaymentMessage) string {
	parts := []string{"aetra-financial-payment-message-root-v1"}
	for _, message := range normalizeCrossZonePaymentMessages(messages) {
		parts = append(parts, message.MessageHash)
	}
	return HashParts(parts...)
}

func ComputePaymentEnvelopeCanonicalSetRoot(records []PaymentEnvelopeCanonicalRecord) string {
	parts := []string{"aetra-financial-payment-canonical-root-v1"}
	for _, record := range normalizePaymentCanonicalEnvelopes(records) {
		parts = append(parts, record.EnvelopeHash)
	}
	return HashParts(parts...)
}

func ComputeFinancialZonePaymentStateRoot(state FinancialZonePaymentState) string {
	state = state.Normalize()
	return HashParts("aetra-financial-payment-state-root-v1", fmt.Sprintf("%020d", state.Height), state.IntentRoot, state.ChannelRoot, state.ConditionRoot, state.RouteRoot, state.SettlementRoot, state.DisputeRoot, state.ReceiptRoot, state.ProofRoot, state.FeeRoot, state.MessageRoot, state.CanonicalRoot)
}

func IsPaymentIntentType(value PaymentIntentType) bool {
	switch value {
	case PaymentIntentInitiate, PaymentIntentReserveRoute, PaymentIntentSettle:
		return true
	default:
		return false
	}
}

func IsPaymentDisputeStatus(value PaymentDisputeStatus) bool {
	switch value {
	case PaymentDisputeOpen, PaymentDisputeAccepted, PaymentDisputeRejected, PaymentDisputeResolved:
		return true
	default:
		return false
	}
}

func IsPaymentEnvelopeObjectType(value PaymentEnvelopeObjectType) bool {
	switch value {
	case PaymentEnvelopeIntent, PaymentEnvelopeRoute, PaymentEnvelopeChannel, PaymentEnvelopeCondition, PaymentEnvelopeSettlement, PaymentEnvelopeReceipt, PaymentEnvelopeDispute, PaymentEnvelopeFee, PaymentEnvelopeMessage:
		return true
	default:
		return false
	}
}

func paymentTask(priority PaymentImplementationPriority, id PaymentImplementationTaskID, task, target string, criteria []string) PaymentImplementationTask {
	return PaymentImplementationTask{Priority: priority, TaskID: id, Task: task, Target: target, AcceptanceCriteria: append([]string{}, criteria...)}
}

func normalizePaymentIntents(intents []PaymentIntent) []PaymentIntent {
	out := make([]PaymentIntent, len(intents))
	for i, intent := range intents {
		out[i] = intent.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PaymentID < out[j].PaymentID })
	return out
}

func normalizeNativeConditionalPayments(conditions []NativeConditionalPayment) []NativeConditionalPayment {
	out := make([]NativeConditionalPayment, len(conditions))
	for i, condition := range conditions {
		out[i] = condition.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ConditionID < out[j].ConditionID })
	return out
}

func normalizePaymentRouteCommitments(routes []PaymentRouteCommitment) []PaymentRouteCommitment {
	out := make([]PaymentRouteCommitment, len(routes))
	for i, route := range routes {
		out[i] = route.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RouteID < out[j].RouteID })
	return out
}

func normalizePaymentSettlements(settlements []PaymentSettlement) []PaymentSettlement {
	out := make([]PaymentSettlement, len(settlements))
	for i, settlement := range settlements {
		out[i] = settlement.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PaymentID < out[j].PaymentID })
	return out
}

func normalizePaymentDisputes(disputes []PaymentDispute) []PaymentDispute {
	out := make([]PaymentDispute, len(disputes))
	for i, dispute := range disputes {
		out[i] = dispute.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].DisputeID < out[j].DisputeID })
	return out
}

func normalizePaymentFeeRecords(records []PaymentFeeAccountingRecord) []PaymentFeeAccountingRecord {
	out := make([]PaymentFeeAccountingRecord, len(records))
	for i, record := range records {
		out[i] = record.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].FeeID < out[j].FeeID })
	return out
}

func normalizeCrossZonePaymentMessages(messages []CrossZonePaymentMessage) []CrossZonePaymentMessage {
	out := make([]CrossZonePaymentMessage, len(messages))
	for i, message := range messages {
		out[i] = message.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].MessageID < out[j].MessageID })
	return out
}

func normalizePaymentCanonicalEnvelopes(records []PaymentEnvelopeCanonicalRecord) []PaymentEnvelopeCanonicalRecord {
	out := make([]PaymentEnvelopeCanonicalRecord, len(records))
	for i, record := range records {
		out[i] = record.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].StateKey < out[j].StateKey })
	return out
}

func validateUniquePaymentIntents(intents []PaymentIntent) error {
	seen := map[string]struct{}{}
	for _, intent := range intents {
		if err := intent.Validate(); err != nil {
			return err
		}
		if _, found := seen[intent.Normalize().PaymentID]; found {
			return errors.New("payments duplicate intent")
		}
		seen[intent.Normalize().PaymentID] = struct{}{}
	}
	return nil
}

func validateNativeConditionalPayments(conditions []NativeConditionalPayment) error {
	seen := map[string]struct{}{}
	for _, condition := range conditions {
		if err := condition.Validate(); err != nil {
			return err
		}
		if _, found := seen[condition.Normalize().ConditionID]; found {
			return errors.New("payments duplicate condition")
		}
		seen[condition.Normalize().ConditionID] = struct{}{}
	}
	return nil
}

func validatePaymentRouteCommitments(routes []PaymentRouteCommitment) error {
	seen := map[string]struct{}{}
	for _, route := range routes {
		route = route.Normalize()
		if err := ValidateHash("payments state route id", route.RouteID); err != nil {
			return err
		}
		if err := ValidateHash("payments state route commitment", route.CommitmentHash); err != nil {
			return err
		}
		if route.Signed {
			if err := addressing.ValidateUserAddress("payments state route committer", route.Committer); err != nil {
				return err
			}
		}
		if !route.Signed && !route.Reserved {
			return errors.New("payments state route must be signed or reserved")
		}
		if route.ExpiresHeight == 0 {
			return errors.New("payments state route expiry is required")
		}
		if _, found := seen[route.RouteID]; found {
			return errors.New("payments duplicate route")
		}
		seen[route.RouteID] = struct{}{}
	}
	return nil
}

func validatePaymentSettlements(settlements []PaymentSettlement) error {
	seen := map[string]struct{}{}
	for _, settlement := range settlements {
		if err := settlement.Validate(); err != nil {
			return err
		}
		if _, found := seen[settlement.Normalize().PaymentID]; found {
			return errors.New("payments duplicate settlement")
		}
		seen[settlement.Normalize().PaymentID] = struct{}{}
	}
	return nil
}

func validatePaymentDisputes(disputes []PaymentDispute) error {
	seen := map[string]struct{}{}
	for _, dispute := range disputes {
		if err := dispute.Validate(); err != nil {
			return err
		}
		if _, found := seen[dispute.Normalize().DisputeID]; found {
			return errors.New("payments duplicate dispute")
		}
		seen[dispute.Normalize().DisputeID] = struct{}{}
	}
	return nil
}

func validatePaymentFeeRecords(records []PaymentFeeAccountingRecord) error {
	seen := map[string]struct{}{}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.Normalize().FeeID]; found {
			return errors.New("payments duplicate fee record")
		}
		seen[record.Normalize().FeeID] = struct{}{}
	}
	return nil
}

func validateCrossZonePaymentMessages(messages []CrossZonePaymentMessage) error {
	seen := map[string]struct{}{}
	for _, message := range messages {
		if err := message.Validate(); err != nil {
			return err
		}
		if _, found := seen[message.Normalize().MessageID]; found {
			return errors.New("payments duplicate cross-zone message")
		}
		seen[message.Normalize().MessageID] = struct{}{}
	}
	return nil
}

func validatePaymentCanonicalEnvelopes(records []PaymentEnvelopeCanonicalRecord) error {
	seen := map[string]struct{}{}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.Normalize().StateKey]; found {
			return errors.New("payments duplicate canonical envelope")
		}
		seen[record.Normalize().StateKey] = struct{}{}
	}
	return nil
}
