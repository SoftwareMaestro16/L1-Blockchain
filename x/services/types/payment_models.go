package types

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type PaymentSettlementTiming string

const (
	PaymentSettlementBeforeExecution	PaymentSettlementTiming	= "before_execution"
	PaymentSettlementAfterExecution		PaymentSettlementTiming	= "after_execution"
	PaymentSettlementAfterChallenge		PaymentSettlementTiming	= "after_challenge_window"
)

type PaymentMeteringRecord struct {
	ServiceID	string
	CallID		string
	RequestBytes	uint64
	ResponseBytes	uint64
	StorageBytes	uint64
	MeterID		string
	MeterHeight	uint64
	RecordHash	string
}

type PaymentUsageReceipt struct {
	ServiceID	string
	CallID		string
	ProviderID	string
	ComputeUnits	uint64
	ReceiptHeight	uint64
	SignedBy	string
	SignatureHash	string
	ProofHash	string
	ReceiptHash	string
}

type PaymentSubscriptionEntitlement struct {
	SubscriptionID	string
	Payer		string
	ServiceID	string
	StartHeight	uint64
	EndHeight	uint64
	StartUnix	int64
	EndUnix		int64
	StateBacked	bool
	ProofHash	string
	EntitlementHash	string
}

type PaymentEscrowSettlement struct {
	EscrowID		string
	ServiceID		string
	ReceiptHeight		uint64
	ProofHeight		uint64
	ChallengeWindow		uint64
	SettleAfterHeight	uint64
	SettlementHash		string
}

type PaymentModelQuote struct {
	Envelope				PaymentEnvelope
	Units					uint64
	UnitAmount				string
	AmountDue				string
	SettlementTiming			PaymentSettlementTiming
	RequiresDeterministicMeterRecord	bool
	RequiresUsageReceipt			bool
	RequiresSubscriptionEntitlement		bool
	RequiresEscrowLock			bool
	SettleAfterHeight			uint64
	QuoteHash				string
}

func NewPaymentMeteringRecord(record PaymentMeteringRecord) (PaymentMeteringRecord, error) {
	if record.RecordHash != "" {
		return PaymentMeteringRecord{}, errors.New("services payment metering record hash must be empty before construction")
	}
	record = canonicalPaymentMeteringRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return PaymentMeteringRecord{}, err
	}
	record.RecordHash = ComputePaymentMeteringRecordHash(record)
	return record, record.Validate()
}

func (record PaymentMeteringRecord) TotalBytes() uint64 {
	return record.RequestBytes + record.ResponseBytes + record.StorageBytes
}

func (record PaymentMeteringRecord) ValidateFormat() error {
	record = canonicalPaymentMeteringRecord(record)
	if err := validateInterfaceToken("services payment meter service id", record.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services payment meter call id", record.CallID); err != nil {
		return err
	}
	if record.TotalBytes() == 0 {
		return errors.New("services payment meter must record at least one byte")
	}
	if err := validateInterfaceToken("services payment meter id", record.MeterID); err != nil {
		return err
	}
	if record.MeterHeight == 0 {
		return errors.New("services payment meter height is required")
	}
	if record.RecordHash != "" {
		return coretypes.ValidateHash("services payment meter record hash", record.RecordHash)
	}
	return nil
}

func (record PaymentMeteringRecord) Validate() error {
	record = canonicalPaymentMeteringRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return err
	}
	if record.RecordHash == "" {
		return errors.New("services payment meter record hash is required")
	}
	if expected := ComputePaymentMeteringRecordHash(record); record.RecordHash != expected {
		return fmt.Errorf("services payment meter record hash mismatch: expected %s", expected)
	}
	return nil
}

func NewPaymentUsageReceipt(receipt PaymentUsageReceipt) (PaymentUsageReceipt, error) {
	if receipt.ReceiptHash != "" {
		return PaymentUsageReceipt{}, errors.New("services payment usage receipt hash must be empty before construction")
	}
	receipt = canonicalPaymentUsageReceipt(receipt)
	if err := receipt.ValidateFormat(); err != nil {
		return PaymentUsageReceipt{}, err
	}
	receipt.ReceiptHash = ComputePaymentUsageReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func (receipt PaymentUsageReceipt) ValidateFormat() error {
	receipt = canonicalPaymentUsageReceipt(receipt)
	if err := validateInterfaceToken("services payment usage service id", receipt.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services payment usage call id", receipt.CallID); err != nil {
		return err
	}
	if receipt.ProviderID != "" {
		if err := validateInterfaceToken("services payment usage provider id", receipt.ProviderID); err != nil {
			return err
		}
	}
	if receipt.ComputeUnits == 0 {
		return errors.New("services payment usage compute units are required")
	}
	if receipt.ReceiptHeight == 0 {
		return errors.New("services payment usage receipt height is required")
	}
	if receipt.SignatureHash == "" && receipt.ProofHash == "" {
		return errors.New("services payment usage receipt requires signature or proof")
	}
	if receipt.SignatureHash != "" {
		if err := validateInterfaceToken("services payment usage signer", receipt.SignedBy); err != nil {
			return err
		}
		if err := coretypes.ValidateHash("services payment usage signature hash", receipt.SignatureHash); err != nil {
			return err
		}
	}
	if receipt.ProofHash != "" {
		if err := coretypes.ValidateHash("services payment usage proof hash", receipt.ProofHash); err != nil {
			return err
		}
	}
	if receipt.ReceiptHash != "" {
		return coretypes.ValidateHash("services payment usage receipt hash", receipt.ReceiptHash)
	}
	return nil
}

func (receipt PaymentUsageReceipt) Validate() error {
	receipt = canonicalPaymentUsageReceipt(receipt)
	if err := receipt.ValidateFormat(); err != nil {
		return err
	}
	if receipt.ReceiptHash == "" {
		return errors.New("services payment usage receipt hash is required")
	}
	if expected := ComputePaymentUsageReceiptHash(receipt); receipt.ReceiptHash != expected {
		return fmt.Errorf("services payment usage receipt hash mismatch: expected %s", expected)
	}
	return nil
}

func NewPaymentSubscriptionEntitlement(entitlement PaymentSubscriptionEntitlement) (PaymentSubscriptionEntitlement, error) {
	if entitlement.EntitlementHash != "" {
		return PaymentSubscriptionEntitlement{}, errors.New("services payment subscription entitlement hash must be empty before construction")
	}
	entitlement = canonicalPaymentSubscriptionEntitlement(entitlement)
	if err := entitlement.ValidateFormat(); err != nil {
		return PaymentSubscriptionEntitlement{}, err
	}
	entitlement.EntitlementHash = ComputePaymentSubscriptionEntitlementHash(entitlement)
	return entitlement, entitlement.Validate()
}

func (entitlement PaymentSubscriptionEntitlement) ValidateFormat() error {
	entitlement = canonicalPaymentSubscriptionEntitlement(entitlement)
	if err := validateInterfaceToken("services payment subscription id", entitlement.SubscriptionID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("services payment subscription payer", entitlement.Payer); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment subscription service id", entitlement.ServiceID); err != nil {
		return err
	}
	if entitlement.StartHeight == 0 || entitlement.EndHeight == 0 || entitlement.EndHeight <= entitlement.StartHeight {
		return errors.New("services payment subscription height range is invalid")
	}
	if entitlement.EndUnix != 0 && entitlement.EndUnix <= entitlement.StartUnix {
		return errors.New("services payment subscription time range is invalid")
	}
	if !entitlement.StateBacked && entitlement.ProofHash == "" {
		return errors.New("services payment subscription requires state or proof-backed entitlement")
	}
	if entitlement.ProofHash != "" {
		if err := coretypes.ValidateHash("services payment subscription proof hash", entitlement.ProofHash); err != nil {
			return err
		}
	}
	if entitlement.EntitlementHash != "" {
		return coretypes.ValidateHash("services payment subscription entitlement hash", entitlement.EntitlementHash)
	}
	return nil
}

func (entitlement PaymentSubscriptionEntitlement) Validate() error {
	entitlement = canonicalPaymentSubscriptionEntitlement(entitlement)
	if err := entitlement.ValidateFormat(); err != nil {
		return err
	}
	if entitlement.EntitlementHash == "" {
		return errors.New("services payment subscription entitlement hash is required")
	}
	if expected := ComputePaymentSubscriptionEntitlementHash(entitlement); entitlement.EntitlementHash != expected {
		return fmt.Errorf("services payment subscription entitlement hash mismatch: expected %s", expected)
	}
	return nil
}

func (entitlement PaymentSubscriptionEntitlement) ActiveAt(height uint64, unixTime int64) bool {
	if height < entitlement.StartHeight || height > entitlement.EndHeight {
		return false
	}
	if entitlement.EndUnix != 0 && (unixTime < entitlement.StartUnix || unixTime > entitlement.EndUnix) {
		return false
	}
	return true
}

func NewPaymentEscrowSettlement(settlement PaymentEscrowSettlement) (PaymentEscrowSettlement, error) {
	if settlement.SettlementHash != "" {
		return PaymentEscrowSettlement{}, errors.New("services payment escrow settlement hash must be empty before construction")
	}
	settlement = canonicalPaymentEscrowSettlement(settlement)
	if err := settlement.ValidateFormat(); err != nil {
		return PaymentEscrowSettlement{}, err
	}
	settlement.SettleAfterHeight = maxPaymentHeight(settlement.ReceiptHeight, settlement.ProofHeight) + settlement.ChallengeWindow
	settlement.SettlementHash = ComputePaymentEscrowSettlementHash(settlement)
	return settlement, settlement.Validate()
}

func (settlement PaymentEscrowSettlement) ValidateFormat() error {
	settlement = canonicalPaymentEscrowSettlement(settlement)
	if err := validateInterfaceToken("services payment escrow id", settlement.EscrowID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment escrow service id", settlement.ServiceID); err != nil {
		return err
	}
	if settlement.ReceiptHeight == 0 {
		return errors.New("services payment escrow requires receipt height")
	}
	if settlement.SettleAfterHeight != 0 && settlement.SettleAfterHeight < maxPaymentHeight(settlement.ReceiptHeight, settlement.ProofHeight)+settlement.ChallengeWindow {
		return errors.New("services payment escrow settlement is before challenge window")
	}
	if settlement.SettlementHash != "" {
		return coretypes.ValidateHash("services payment escrow settlement hash", settlement.SettlementHash)
	}
	return nil
}

func (settlement PaymentEscrowSettlement) Validate() error {
	settlement = canonicalPaymentEscrowSettlement(settlement)
	if err := settlement.ValidateFormat(); err != nil {
		return err
	}
	if settlement.SettleAfterHeight == 0 {
		return errors.New("services payment escrow settle height is required")
	}
	if settlement.SettlementHash == "" {
		return errors.New("services payment escrow settlement hash is required")
	}
	if expected := ComputePaymentEscrowSettlementHash(settlement); settlement.SettlementHash != expected {
		return fmt.Errorf("services payment escrow settlement hash mismatch: expected %s", expected)
	}
	return nil
}

func QuotePerCallPayment(envelope PaymentEnvelope, trust coretypes.ServiceTrustModel) (PaymentModelQuote, error) {
	if err := envelope.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if envelope.PricingUnit != coretypes.ServicePricingPerCall {
		return PaymentModelQuote{}, errors.New("services per-call payment requires CALL pricing unit")
	}
	quote := PaymentModelQuote{
		Envelope:		envelope,
		Units:			1,
		UnitAmount:		envelope.Amount,
		AmountDue:		envelope.Amount,
		SettlementTiming:	paymentTimingForTrust(envelope.SettlementMode, trust),
	}
	quote.QuoteHash = ComputePaymentModelQuoteHash(quote)
	return quote, quote.Validate()
}

func QuotePerBytePayment(envelope PaymentEnvelope, meter PaymentMeteringRecord) (PaymentModelQuote, error) {
	if err := envelope.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if envelope.PricingUnit != coretypes.ServicePricingPerByte {
		return PaymentModelQuote{}, errors.New("services per-byte payment requires BYTE pricing unit")
	}
	if err := meter.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if meter.ServiceID != envelope.PayeeService {
		return PaymentModelQuote{}, errors.New("services per-byte payment meter service mismatch")
	}
	due := multiplyPaymentAmount(envelope.Amount, meter.TotalBytes())
	if err := validatePaymentMax(envelope, due); err != nil {
		return PaymentModelQuote{}, err
	}
	quote := PaymentModelQuote{
		Envelope:				envelope,
		Units:					meter.TotalBytes(),
		UnitAmount:				envelope.Amount,
		AmountDue:				due,
		SettlementTiming:			PaymentSettlementAfterExecution,
		RequiresDeterministicMeterRecord:	true,
	}
	quote.QuoteHash = ComputePaymentModelQuoteHash(quote)
	return quote, quote.Validate()
}

func QuotePerComputeUnitPayment(envelope PaymentEnvelope, receipt PaymentUsageReceipt) (PaymentModelQuote, error) {
	if err := envelope.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if envelope.PricingUnit != coretypes.ServicePricingPerComputeUnit {
		return PaymentModelQuote{}, errors.New("services compute payment requires COMPUTE_UNIT pricing unit")
	}
	if err := receipt.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if receipt.ServiceID != envelope.PayeeService {
		return PaymentModelQuote{}, errors.New("services compute payment receipt service mismatch")
	}
	due := multiplyPaymentAmount(envelope.Amount, receipt.ComputeUnits)
	if err := validatePaymentMax(envelope, due); err != nil {
		return PaymentModelQuote{}, err
	}
	quote := PaymentModelQuote{
		Envelope:		envelope,
		Units:			receipt.ComputeUnits,
		UnitAmount:		envelope.Amount,
		AmountDue:		due,
		SettlementTiming:	PaymentSettlementAfterExecution,
		RequiresUsageReceipt:	true,
	}
	quote.QuoteHash = ComputePaymentModelQuoteHash(quote)
	return quote, quote.Validate()
}

func QuoteSubscriptionPayment(envelope PaymentEnvelope, entitlement PaymentSubscriptionEntitlement, height uint64, unixTime int64) (PaymentModelQuote, error) {
	if err := envelope.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if envelope.PricingUnit != coretypes.ServicePricingSubscription {
		return PaymentModelQuote{}, errors.New("services subscription payment requires SUBSCRIPTION pricing unit")
	}
	if err := entitlement.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if entitlement.ServiceID != envelope.PayeeService || entitlement.Payer != envelope.Payer {
		return PaymentModelQuote{}, errors.New("services subscription entitlement mismatch")
	}
	if !entitlement.ActiveAt(height, unixTime) {
		return PaymentModelQuote{}, errors.New("services subscription entitlement is not active")
	}
	quote := PaymentModelQuote{
		Envelope:				envelope,
		Units:					entitlement.EndHeight - entitlement.StartHeight + 1,
		UnitAmount:				envelope.Amount,
		AmountDue:				envelope.Amount,
		SettlementTiming:			PaymentSettlementBeforeExecution,
		RequiresSubscriptionEntitlement:	true,
	}
	quote.QuoteHash = ComputePaymentModelQuoteHash(quote)
	return quote, quote.Validate()
}

func PlanEscrowPaymentSettlement(envelope PaymentEnvelope, settlement PaymentEscrowSettlement) (PaymentModelQuote, error) {
	if err := envelope.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	if envelope.SettlementMode != coretypes.ServicePaymentEscrow {
		return PaymentModelQuote{}, errors.New("services escrow payment requires ESCROW settlement mode")
	}
	if envelope.EscrowIDOptional != settlement.EscrowID || envelope.PayeeService != settlement.ServiceID {
		return PaymentModelQuote{}, errors.New("services escrow settlement envelope mismatch")
	}
	if err := settlement.Validate(); err != nil {
		return PaymentModelQuote{}, err
	}
	quote := PaymentModelQuote{
		Envelope:		envelope,
		Units:			1,
		UnitAmount:		envelope.Amount,
		AmountDue:		envelope.Amount,
		SettlementTiming:	PaymentSettlementAfterChallenge,
		RequiresEscrowLock:	true,
		SettleAfterHeight:	settlement.SettleAfterHeight,
	}
	quote.QuoteHash = ComputePaymentModelQuoteHash(quote)
	return quote, quote.Validate()
}

func (quote PaymentModelQuote) Validate() error {
	if err := quote.Envelope.Validate(); err != nil {
		return err
	}
	if quote.Units == 0 {
		return errors.New("services payment quote units are required")
	}
	if err := validatePositivePaymentAmount("services payment quote unit amount", quote.UnitAmount); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("services payment quote amount due", quote.AmountDue); err != nil {
		return err
	}
	if !IsPaymentSettlementTiming(quote.SettlementTiming) {
		return fmt.Errorf("unknown services payment settlement timing %q", quote.SettlementTiming)
	}
	if quote.SettlementTiming == PaymentSettlementAfterChallenge && quote.SettleAfterHeight == 0 {
		return errors.New("services payment quote challenge settlement height is required")
	}
	if err := coretypes.ValidateHash("services payment quote hash", quote.QuoteHash); err != nil {
		return err
	}
	if expected := ComputePaymentModelQuoteHash(quote); quote.QuoteHash != expected {
		return fmt.Errorf("services payment quote hash mismatch: expected %s", expected)
	}
	return validatePaymentMax(quote.Envelope, quote.AmountDue)
}

func ComputePaymentMeteringRecordHash(record PaymentMeteringRecord) string {
	record = canonicalPaymentMeteringRecord(record)
	return servicesHashParts("aetra-services-payment-meter-v1", record.ServiceID, record.CallID, fmt.Sprint(record.RequestBytes), fmt.Sprint(record.ResponseBytes), fmt.Sprint(record.StorageBytes), record.MeterID, fmt.Sprint(record.MeterHeight))
}

func ComputePaymentUsageReceiptHash(receipt PaymentUsageReceipt) string {
	receipt = canonicalPaymentUsageReceipt(receipt)
	return servicesHashParts("aetra-services-payment-usage-receipt-v1", receipt.ServiceID, receipt.CallID, receipt.ProviderID, fmt.Sprint(receipt.ComputeUnits), fmt.Sprint(receipt.ReceiptHeight), receipt.SignedBy, receipt.SignatureHash, receipt.ProofHash)
}

func ComputePaymentSubscriptionEntitlementHash(entitlement PaymentSubscriptionEntitlement) string {
	entitlement = canonicalPaymentSubscriptionEntitlement(entitlement)
	return servicesHashParts("aetra-services-payment-subscription-v1", entitlement.SubscriptionID, entitlement.Payer, entitlement.ServiceID, fmt.Sprint(entitlement.StartHeight), fmt.Sprint(entitlement.EndHeight), fmt.Sprint(entitlement.StartUnix), fmt.Sprint(entitlement.EndUnix), fmt.Sprint(entitlement.StateBacked), entitlement.ProofHash)
}

func ComputePaymentEscrowSettlementHash(settlement PaymentEscrowSettlement) string {
	settlement = canonicalPaymentEscrowSettlement(settlement)
	return servicesHashParts("aetra-services-payment-escrow-settlement-v1", settlement.EscrowID, settlement.ServiceID, fmt.Sprint(settlement.ReceiptHeight), fmt.Sprint(settlement.ProofHeight), fmt.Sprint(settlement.ChallengeWindow), fmt.Sprint(settlement.SettleAfterHeight))
}

func ComputePaymentModelQuoteHash(quote PaymentModelQuote) string {
	return servicesHashParts("aetra-services-payment-model-quote-v1", quote.Envelope.EnvelopeHash, fmt.Sprint(quote.Units), quote.UnitAmount, quote.AmountDue, string(quote.SettlementTiming), fmt.Sprint(quote.RequiresDeterministicMeterRecord), fmt.Sprint(quote.RequiresUsageReceipt), fmt.Sprint(quote.RequiresSubscriptionEntitlement), fmt.Sprint(quote.RequiresEscrowLock), fmt.Sprint(quote.SettleAfterHeight))
}

func IsPaymentSettlementTiming(timing PaymentSettlementTiming) bool {
	switch timing {
	case PaymentSettlementBeforeExecution, PaymentSettlementAfterExecution, PaymentSettlementAfterChallenge:
		return true
	default:
		return false
	}
}

func paymentTimingForTrust(mode coretypes.ServicePaymentSettlementMode, trust coretypes.ServiceTrustModel) PaymentSettlementTiming {
	if mode == coretypes.ServicePaymentPrepaid || mode == coretypes.ServicePaymentStreaming || trust == coretypes.ServiceTrustEconomicallySecured || trust == coretypes.ServiceTrustFullyTrusted {
		return PaymentSettlementBeforeExecution
	}
	return PaymentSettlementAfterExecution
}

func canonicalPaymentMeteringRecord(record PaymentMeteringRecord) PaymentMeteringRecord {
	record.ServiceID = strings.TrimSpace(record.ServiceID)
	record.CallID = strings.ToLower(strings.TrimSpace(record.CallID))
	record.MeterID = strings.TrimSpace(record.MeterID)
	record.RecordHash = strings.ToLower(strings.TrimSpace(record.RecordHash))
	return record
}

func canonicalPaymentUsageReceipt(receipt PaymentUsageReceipt) PaymentUsageReceipt {
	receipt.ServiceID = strings.TrimSpace(receipt.ServiceID)
	receipt.CallID = strings.ToLower(strings.TrimSpace(receipt.CallID))
	receipt.ProviderID = strings.TrimSpace(receipt.ProviderID)
	receipt.SignedBy = strings.TrimSpace(receipt.SignedBy)
	receipt.SignatureHash = strings.ToLower(strings.TrimSpace(receipt.SignatureHash))
	receipt.ProofHash = strings.ToLower(strings.TrimSpace(receipt.ProofHash))
	receipt.ReceiptHash = strings.ToLower(strings.TrimSpace(receipt.ReceiptHash))
	return receipt
}

func canonicalPaymentSubscriptionEntitlement(entitlement PaymentSubscriptionEntitlement) PaymentSubscriptionEntitlement {
	entitlement.SubscriptionID = strings.TrimSpace(entitlement.SubscriptionID)
	entitlement.Payer = strings.TrimSpace(entitlement.Payer)
	entitlement.ServiceID = strings.TrimSpace(entitlement.ServiceID)
	entitlement.ProofHash = strings.ToLower(strings.TrimSpace(entitlement.ProofHash))
	entitlement.EntitlementHash = strings.ToLower(strings.TrimSpace(entitlement.EntitlementHash))
	return entitlement
}

func canonicalPaymentEscrowSettlement(settlement PaymentEscrowSettlement) PaymentEscrowSettlement {
	settlement.EscrowID = strings.TrimSpace(settlement.EscrowID)
	settlement.ServiceID = strings.TrimSpace(settlement.ServiceID)
	settlement.SettlementHash = strings.ToLower(strings.TrimSpace(settlement.SettlementHash))
	return settlement
}

func multiplyPaymentAmount(amount string, units uint64) string {
	n, _ := new(big.Int).SetString(amount, 10)
	n.Mul(n, new(big.Int).SetUint64(units))
	return n.String()
}

func validatePaymentMax(envelope PaymentEnvelope, amountDue string) error {
	if envelope.MaxAmountOptional == "" {
		return nil
	}
	if comparePaymentAmounts(envelope.MaxAmountOptional, amountDue) < 0 {
		return errors.New("services payment amount due exceeds max amount")
	}
	return nil
}

func maxPaymentHeight(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}
