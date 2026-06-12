package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	FinancialSettlementZoneID	= "FINANCIAL_ZONE"
	MaxPaymentRouteHintLength	= 128
)

type PaymentSettlementMode string

const (
	PaymentSettlementDirect			PaymentSettlementMode	= "DIRECT"
	PaymentSettlementStreaming		PaymentSettlementMode	= "STREAMING"
	PaymentSettlementConditional		PaymentSettlementMode	= "CONDITIONAL"
	PaymentSettlementOffchainChannel	PaymentSettlementMode	= "OFFCHAIN_CHANNEL"
	PaymentSettlementZoneToZone		PaymentSettlementMode	= "ZONE_TO_ZONE"
)

type PaymentStatus string

const (
	PaymentStatusQueued	PaymentStatus	= "queued"
	PaymentStatusSettled	PaymentStatus	= "settled"
	PaymentStatusExpired	PaymentStatus	= "expired"
	PaymentStatusDisputed	PaymentStatus	= "disputed"
	PaymentStatusRejected	PaymentStatus	= "rejected"
)

type Payment struct {
	From		string
	To		string
	Amount		string
	ConditionHash	string
	Expiry		uint64
	RouteHint	string
	PaymentID	string
	SourceZone	string
	DestinationZone	string
	Denom		string
	FeeLimit	string
	SettlementMode	PaymentSettlementMode
	Nonce		uint64
	Signature	string
}

type PaymentAbstractionReceipt struct {
	PaymentID	string
	Status		PaymentStatus
	SettlementZone	string
	GasUsed		uint64
	FeeCharged	string
	ErrorCode	string
	ExecutedHeight	uint64
	ReceiptHash	string
}

type PaymentNonceRecord struct {
	SourceZone	string
	Sender		string
	Nonce		uint64
	PaymentID	string
	ExpiresHeight	uint64
}

type PaymentAbstractionState struct {
	Height		uint64
	Payments	[]Payment
	Receipts	[]PaymentAbstractionReceipt
	Nonces		[]PaymentNonceRecord
}

type PaymentRouteQuote struct {
	RouteHint	string
	SourceZone	string
	DestinationZone	string
	FeeAmount	string
	Capacity	string
	ExpiresHeight	uint64
	RouteScoreHash	string
}

func EmptyPaymentAbstractionState(height uint64) PaymentAbstractionState {
	return PaymentAbstractionState{
		Height:		height,
		Payments:	[]Payment{},
		Receipts:	[]PaymentAbstractionReceipt{},
		Nonces:		[]PaymentNonceRecord{},
	}
}

func BuildPayment(payment Payment) (Payment, error) {
	if payment.PaymentID != "" {
		return Payment{}, errors.New("payments abstraction payment id must be empty before construction")
	}
	if payment.Signature != "" {
		return Payment{}, errors.New("payments abstraction signature must be empty before construction")
	}
	payment = payment.Normalize()
	if err := payment.ValidateFormat(); err != nil {
		return Payment{}, err
	}
	payment.PaymentID = ComputePaymentID(payment)
	payment.Signature = ComputePaymentSignatureHash(payment)
	return payment, payment.Validate()
}

func RegisterPayment(state PaymentAbstractionState, payment Payment, currentHeight uint64) (PaymentAbstractionState, Payment, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentAbstractionState{}, Payment{}, errors.New("payments abstraction register height must be positive")
	}
	if err := state.Validate(); err != nil {
		return PaymentAbstractionState{}, Payment{}, err
	}
	payment = payment.Normalize()
	if payment.PaymentID == "" {
		var err error
		payment, err = BuildPayment(payment)
		if err != nil {
			return PaymentAbstractionState{}, Payment{}, err
		}
	} else if err := payment.Validate(); err != nil {
		return PaymentAbstractionState{}, Payment{}, err
	}
	if currentHeight > payment.Expiry {
		return PaymentAbstractionState{}, Payment{}, errors.New("payments abstraction cannot register expired payment")
	}
	if _, found := state.PaymentByID(payment.PaymentID); found {
		return PaymentAbstractionState{}, Payment{}, errors.New("payments abstraction payment already exists")
	}
	if _, found := state.NonceRecord(payment.SourceZone, payment.From, payment.Nonce); found {
		return PaymentAbstractionState{}, Payment{}, errors.New("payments abstraction sender nonce already used")
	}
	next := state.Clone()
	next.Height = currentHeight
	next.Payments = append(next.Payments, payment)
	next.Nonces = append(next.Nonces, PaymentNonceRecord{
		SourceZone:	payment.SourceZone,
		Sender:		payment.From,
		Nonce:		payment.Nonce,
		PaymentID:	payment.PaymentID,
		ExpiresHeight:	payment.Expiry + DefaultReplayHorizon,
	}.Normalize())
	sortPayments(next.Payments)
	sortPaymentNonceRecords(next.Nonces)
	return next, payment, next.Validate()
}

func SettlePaymentInFinancialZone(state PaymentAbstractionState, paymentID string, gasUsed uint64, feeCharged string, currentHeight uint64) (PaymentAbstractionState, PaymentAbstractionReceipt, error) {
	return appendPaymentAbstractionReceipt(state, PaymentAbstractionReceipt{
		PaymentID:	paymentID,
		Status:		PaymentStatusSettled,
		SettlementZone:	FinancialSettlementZoneID,
		GasUsed:	gasUsed,
		FeeCharged:	feeCharged,
		ExecutedHeight:	currentHeight,
	})
}

func RejectPaymentInFinancialZone(state PaymentAbstractionState, paymentID, errorCode string, currentHeight uint64) (PaymentAbstractionState, PaymentAbstractionReceipt, error) {
	return appendPaymentAbstractionReceipt(state, PaymentAbstractionReceipt{
		PaymentID:	paymentID,
		Status:		PaymentStatusRejected,
		SettlementZone:	FinancialSettlementZoneID,
		ErrorCode:	errorCode,
		ExecutedHeight:	currentHeight,
	})
}

func ExpirePayments(state PaymentAbstractionState, currentHeight uint64) (PaymentAbstractionState, []PaymentAbstractionReceipt, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentAbstractionState{}, nil, errors.New("payments abstraction expiry height must be positive")
	}
	if err := state.Validate(); err != nil {
		return PaymentAbstractionState{}, nil, err
	}
	next := state.Clone()
	next.Height = currentHeight
	created := make([]PaymentAbstractionReceipt, 0)
	for _, payment := range state.Payments {
		payment = payment.Normalize()
		if payment.Expiry >= currentHeight {
			continue
		}
		if _, found := state.ReceiptByPaymentID(payment.PaymentID); found {
			continue
		}
		receipt, err := BuildPaymentAbstractionReceipt(PaymentAbstractionReceipt{
			PaymentID:	payment.PaymentID,
			Status:		PaymentStatusExpired,
			SettlementZone:	FinancialSettlementZoneID,
			ExecutedHeight:	currentHeight,
		})
		if err != nil {
			return PaymentAbstractionState{}, nil, err
		}
		next.Receipts = append(next.Receipts, receipt)
		created = append(created, receipt)
	}
	sortPaymentAbstractionReceipts(next.Receipts)
	return next, created, next.Validate()
}

func appendPaymentAbstractionReceipt(state PaymentAbstractionState, receipt PaymentAbstractionReceipt) (PaymentAbstractionState, PaymentAbstractionReceipt, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentAbstractionState{}, PaymentAbstractionReceipt{}, err
	}
	receipt = receipt.Normalize()
	if receipt.ExecutedHeight == 0 {
		return PaymentAbstractionState{}, PaymentAbstractionReceipt{}, errors.New("payments abstraction receipt height must be positive")
	}
	if _, found := state.PaymentByID(receipt.PaymentID); !found {
		return PaymentAbstractionState{}, PaymentAbstractionReceipt{}, errors.New("payments abstraction receipt references unknown payment")
	}
	if _, found := state.ReceiptByPaymentID(receipt.PaymentID); found {
		return PaymentAbstractionState{}, PaymentAbstractionReceipt{}, errors.New("payments abstraction receipt already exists")
	}
	built, err := BuildPaymentAbstractionReceipt(receipt)
	if err != nil {
		return PaymentAbstractionState{}, PaymentAbstractionReceipt{}, err
	}
	next := state.Clone()
	next.Height = receipt.ExecutedHeight
	next.Receipts = append(next.Receipts, built)
	sortPaymentAbstractionReceipts(next.Receipts)
	return next, built, next.Validate()
}

func SelectOptimalPaymentRoute(payment Payment, quotes []PaymentRouteQuote, currentHeight uint64) (PaymentRouteQuote, error) {
	payment = payment.Normalize()
	if err := payment.Validate(); err != nil {
		return PaymentRouteQuote{}, err
	}
	if currentHeight == 0 {
		return PaymentRouteQuote{}, errors.New("payments abstraction route selection height must be positive")
	}
	amount, err := parsePositiveInt("payments abstraction route amount", payment.Amount)
	if err != nil {
		return PaymentRouteQuote{}, err
	}
	feeLimit, err := parseNonNegativeInt("payments abstraction route fee limit", payment.FeeLimit)
	if err != nil {
		return PaymentRouteQuote{}, err
	}
	candidates := make([]PaymentRouteQuote, 0, len(quotes))
	for _, quote := range quotes {
		quote = quote.Normalize()
		if err := quote.ValidateFormat(); err != nil {
			continue
		}
		if quote.SourceZone != payment.SourceZone || quote.DestinationZone != payment.DestinationZone {
			continue
		}
		if quote.ExpiresHeight < currentHeight {
			continue
		}
		fee, _ := parseNonNegativeInt("payments abstraction route quote fee", quote.FeeAmount)
		if feeLimit.IsPositive() && fee.GT(feeLimit) {
			continue
		}
		capacity, _ := parsePositiveInt("payments abstraction route quote capacity", quote.Capacity)
		if capacity.LT(amount) {
			continue
		}
		quote.RouteScoreHash = ComputePaymentRouteScoreHash(payment, quote)
		candidates = append(candidates, quote)
	}
	if len(candidates) == 0 {
		return PaymentRouteQuote{}, errors.New("payments abstraction no eligible route")
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		leftFee, _ := parseNonNegativeInt("payments abstraction route fee", candidates[i].FeeAmount)
		rightFee, _ := parseNonNegativeInt("payments abstraction route fee", candidates[j].FeeAmount)
		if !leftFee.Equal(rightFee) {
			return leftFee.LT(rightFee)
		}
		if candidates[i].ExpiresHeight != candidates[j].ExpiresHeight {
			return candidates[i].ExpiresHeight < candidates[j].ExpiresHeight
		}
		return candidates[i].RouteHint < candidates[j].RouteHint
	})
	return candidates[0], candidates[0].Validate()
}

func BuildPaymentAbstractionReceipt(receipt PaymentAbstractionReceipt) (PaymentAbstractionReceipt, error) {
	if receipt.ReceiptHash != "" {
		return PaymentAbstractionReceipt{}, errors.New("payments abstraction receipt hash must be empty before construction")
	}
	receipt = receipt.Normalize()
	if err := receipt.ValidateFormat(); err != nil {
		return PaymentAbstractionReceipt{}, err
	}
	receipt.ReceiptHash = ComputePaymentAbstractionReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func ComputePaymentID(payment Payment) string {
	payment = payment.Normalize()
	return HashParts(
		"aetra-payment-envelope-id-v1",
		payment.SourceZone,
		payment.DestinationZone,
		payment.From,
		payment.To,
		payment.Denom,
		payment.Amount,
		payment.ConditionHash,
		fmt.Sprintf("%020d", payment.Expiry),
		payment.RouteHint,
		payment.FeeLimit,
		string(payment.SettlementMode),
		fmt.Sprintf("%020d", payment.Nonce),
	)
}

func ComputePaymentSignatureHash(payment Payment) string {
	payment = payment.Normalize()
	return HashParts(
		"aetra-payment-envelope-signature-v1",
		payment.From,
		payment.SourceZone,
		fmt.Sprintf("%020d", payment.Nonce),
		payment.PaymentID,
	)
}

func ComputePaymentAbstractionReceiptHash(receipt PaymentAbstractionReceipt) string {
	receipt = receipt.Normalize()
	return HashParts(
		"aetra-payment-receipt-v1",
		receipt.PaymentID,
		string(receipt.Status),
		receipt.SettlementZone,
		fmt.Sprintf("%020d", receipt.GasUsed),
		receipt.FeeCharged,
		receipt.ErrorCode,
		fmt.Sprintf("%020d", receipt.ExecutedHeight),
	)
}

func ComputePaymentRouteScoreHash(payment Payment, quote PaymentRouteQuote) string {
	payment = payment.Normalize()
	quote = quote.Normalize()
	return HashParts(
		"aetra-payment-route-score-v1",
		payment.PaymentID,
		quote.RouteHint,
		quote.SourceZone,
		quote.DestinationZone,
		quote.FeeAmount,
		quote.Capacity,
		fmt.Sprintf("%020d", quote.ExpiresHeight),
	)
}

func ComputePaymentAbstractionStateRoot(state PaymentAbstractionState) string {
	state = state.Export()
	parts := []string{"aetra-payment-abstraction-state-root-v1", fmt.Sprintf("%020d", state.Height)}
	for _, payment := range state.Payments {
		parts = append(parts, payment.PaymentID)
	}
	for _, receipt := range state.Receipts {
		parts = append(parts, receipt.ReceiptHash)
	}
	for _, nonce := range state.Nonces {
		parts = append(parts, nonce.NonceKey(), nonce.PaymentID, fmt.Sprintf("%020d", nonce.ExpiresHeight))
	}
	return HashParts(parts...)
}

func (payment Payment) Normalize() Payment {
	payment.From = strings.TrimSpace(payment.From)
	payment.To = strings.TrimSpace(payment.To)
	payment.Amount = strings.TrimSpace(payment.Amount)
	payment.ConditionHash = normalizeOptionalHash(payment.ConditionHash)
	payment.RouteHint = strings.TrimSpace(payment.RouteHint)
	payment.PaymentID = normalizeOptionalHash(payment.PaymentID)
	payment.SourceZone = strings.TrimSpace(payment.SourceZone)
	payment.DestinationZone = strings.TrimSpace(payment.DestinationZone)
	payment.Denom = normalizeAssetDenom(payment.Denom)
	payment.FeeLimit = strings.TrimSpace(payment.FeeLimit)
	if payment.FeeLimit == "" {
		payment.FeeLimit = "0"
	}
	payment.Signature = normalizeOptionalHash(payment.Signature)
	if payment.SettlementMode == "" {
		payment.SettlementMode = PaymentSettlementDirect
	}
	return payment
}

func (payment Payment) ValidateFormat() error {
	payment = payment.Normalize()
	if err := addressing.ValidateUserAddress("payments abstraction from", payment.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments abstraction to", payment.To); err != nil {
		return err
	}
	if payment.From == payment.To {
		return errors.New("payments abstraction parties must differ")
	}
	if err := validatePositiveInt("payments abstraction amount", payment.Amount); err != nil {
		return err
	}
	if err := validatePaymentZoneID("payments abstraction source zone", payment.SourceZone); err != nil {
		return err
	}
	if err := validatePaymentZoneID("payments abstraction destination zone", payment.DestinationZone); err != nil {
		return err
	}
	if err := validatePaymentDenom(payment.Denom); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments abstraction fee limit", payment.FeeLimit); err != nil {
		return err
	}
	if payment.Expiry == 0 {
		return errors.New("payments abstraction expiry must be positive")
	}
	if payment.Nonce == 0 {
		return errors.New("payments abstraction nonce must be positive")
	}
	if !IsPaymentSettlementMode(payment.SettlementMode) {
		return fmt.Errorf("unknown payments abstraction settlement mode %q", payment.SettlementMode)
	}
	if len(payment.RouteHint) > MaxPaymentRouteHintLength {
		return fmt.Errorf("payments abstraction route hint must be <= %d bytes", MaxPaymentRouteHintLength)
	}
	if payment.ConditionHash != "" {
		if err := ValidateHash("payments abstraction condition hash", payment.ConditionHash); err != nil {
			return err
		}
	}
	if payment.SettlementMode == PaymentSettlementConditional && payment.ConditionHash == "" {
		return errors.New("payments abstraction conditional payment requires condition hash")
	}
	if payment.PaymentID != "" {
		if err := ValidateHash("payments abstraction payment id", payment.PaymentID); err != nil {
			return err
		}
	}
	if payment.Signature != "" {
		if err := ValidateHash("payments abstraction signature", payment.Signature); err != nil {
			return err
		}
	}
	return nil
}

func (payment Payment) Validate() error {
	payment = payment.Normalize()
	if err := payment.ValidateFormat(); err != nil {
		return err
	}
	if payment.PaymentID == "" {
		return errors.New("payments abstraction payment id is required")
	}
	if expected := ComputePaymentID(payment); payment.PaymentID != expected {
		return fmt.Errorf("payments abstraction payment id mismatch: expected %s", expected)
	}
	if payment.Signature == "" {
		return errors.New("payments abstraction signature is required")
	}
	if expected := ComputePaymentSignatureHash(payment); payment.Signature != expected {
		return fmt.Errorf("payments abstraction signature mismatch: expected %s", expected)
	}
	return nil
}

func (receipt PaymentAbstractionReceipt) Normalize() PaymentAbstractionReceipt {
	receipt.PaymentID = normalizeHash(receipt.PaymentID)
	receipt.SettlementZone = strings.TrimSpace(receipt.SettlementZone)
	receipt.FeeCharged = strings.TrimSpace(receipt.FeeCharged)
	if receipt.FeeCharged == "" {
		receipt.FeeCharged = "0"
	}
	receipt.ErrorCode = strings.TrimSpace(receipt.ErrorCode)
	receipt.ReceiptHash = normalizeOptionalHash(receipt.ReceiptHash)
	return receipt
}

func (receipt PaymentAbstractionReceipt) ValidateFormat() error {
	receipt = receipt.Normalize()
	if err := ValidateHash("payments abstraction receipt payment id", receipt.PaymentID); err != nil {
		return err
	}
	if !IsPaymentStatus(receipt.Status) {
		return fmt.Errorf("unknown payments abstraction receipt status %q", receipt.Status)
	}
	if receipt.SettlementZone != FinancialSettlementZoneID {
		return errors.New("payments abstraction final settlement must use financial zone")
	}
	if err := validateNonNegativeInt("payments abstraction receipt fee", receipt.FeeCharged); err != nil {
		return err
	}
	if receipt.ExecutedHeight == 0 {
		return errors.New("payments abstraction receipt height must be positive")
	}
	if receipt.Status == PaymentStatusRejected && receipt.ErrorCode == "" {
		return errors.New("payments abstraction rejected receipt requires error code")
	}
	if receipt.ReceiptHash != "" {
		if err := ValidateHash("payments abstraction receipt hash", receipt.ReceiptHash); err != nil {
			return err
		}
	}
	return nil
}

func (receipt PaymentAbstractionReceipt) Validate() error {
	receipt = receipt.Normalize()
	if err := receipt.ValidateFormat(); err != nil {
		return err
	}
	if receipt.ReceiptHash == "" {
		return errors.New("payments abstraction receipt hash is required")
	}
	if expected := ComputePaymentAbstractionReceiptHash(receipt); receipt.ReceiptHash != expected {
		return fmt.Errorf("payments abstraction receipt hash mismatch: expected %s", expected)
	}
	return nil
}

func (record PaymentNonceRecord) Normalize() PaymentNonceRecord {
	record.SourceZone = strings.TrimSpace(record.SourceZone)
	record.Sender = strings.TrimSpace(record.Sender)
	record.PaymentID = normalizeHash(record.PaymentID)
	return record
}

func (record PaymentNonceRecord) NonceKey() string {
	record = record.Normalize()
	return fmt.Sprintf("%s/%s/%020d", record.SourceZone, record.Sender, record.Nonce)
}

func (record PaymentNonceRecord) Validate() error {
	record = record.Normalize()
	if err := validatePaymentZoneID("payments abstraction nonce source zone", record.SourceZone); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments abstraction nonce sender", record.Sender); err != nil {
		return err
	}
	if record.Nonce == 0 {
		return errors.New("payments abstraction nonce record nonce must be positive")
	}
	if err := ValidateHash("payments abstraction nonce payment id", record.PaymentID); err != nil {
		return err
	}
	if record.ExpiresHeight == 0 {
		return errors.New("payments abstraction nonce record expiry must be positive")
	}
	return nil
}

func (quote PaymentRouteQuote) Normalize() PaymentRouteQuote {
	quote.RouteHint = strings.TrimSpace(quote.RouteHint)
	quote.SourceZone = strings.TrimSpace(quote.SourceZone)
	quote.DestinationZone = strings.TrimSpace(quote.DestinationZone)
	quote.FeeAmount = strings.TrimSpace(quote.FeeAmount)
	if quote.FeeAmount == "" {
		quote.FeeAmount = "0"
	}
	quote.Capacity = strings.TrimSpace(quote.Capacity)
	quote.RouteScoreHash = normalizeOptionalHash(quote.RouteScoreHash)
	return quote
}

func (quote PaymentRouteQuote) ValidateFormat() error {
	quote = quote.Normalize()
	if quote.RouteHint == "" {
		return errors.New("payments abstraction route hint is required")
	}
	if len(quote.RouteHint) > MaxPaymentRouteHintLength {
		return fmt.Errorf("payments abstraction route hint must be <= %d bytes", MaxPaymentRouteHintLength)
	}
	if err := validatePaymentZoneID("payments abstraction route source zone", quote.SourceZone); err != nil {
		return err
	}
	if err := validatePaymentZoneID("payments abstraction route destination zone", quote.DestinationZone); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments abstraction route fee", quote.FeeAmount); err != nil {
		return err
	}
	if err := validatePositiveInt("payments abstraction route capacity", quote.Capacity); err != nil {
		return err
	}
	if quote.ExpiresHeight == 0 {
		return errors.New("payments abstraction route expiry must be positive")
	}
	if quote.RouteScoreHash != "" {
		if err := ValidateHash("payments abstraction route score hash", quote.RouteScoreHash); err != nil {
			return err
		}
	}
	return nil
}

func (quote PaymentRouteQuote) Validate() error {
	quote = quote.Normalize()
	if err := quote.ValidateFormat(); err != nil {
		return err
	}
	if quote.RouteScoreHash == "" {
		return errors.New("payments abstraction route score hash is required")
	}
	return nil
}

func (state PaymentAbstractionState) Export() PaymentAbstractionState {
	out := state.Clone()
	sortPayments(out.Payments)
	sortPaymentAbstractionReceipts(out.Receipts)
	sortPaymentNonceRecords(out.Nonces)
	return out
}

func (state PaymentAbstractionState) Clone() PaymentAbstractionState {
	out := PaymentAbstractionState{
		Height:		state.Height,
		Payments:	make([]Payment, len(state.Payments)),
		Receipts:	make([]PaymentAbstractionReceipt, len(state.Receipts)),
		Nonces:		make([]PaymentNonceRecord, len(state.Nonces)),
	}
	for i, payment := range state.Payments {
		out.Payments[i] = payment.Normalize()
	}
	for i, receipt := range state.Receipts {
		out.Receipts[i] = receipt.Normalize()
	}
	for i, nonce := range state.Nonces {
		out.Nonces[i] = nonce.Normalize()
	}
	return out
}

func (state PaymentAbstractionState) Validate() error {
	state = state.Clone()
	seenPayments := make(map[string]struct{}, len(state.Payments))
	for _, payment := range state.Payments {
		payment = payment.Normalize()
		if err := payment.Validate(); err != nil {
			return err
		}
		if _, found := seenPayments[payment.PaymentID]; found {
			return errors.New("payments abstraction duplicate payment id")
		}
		seenPayments[payment.PaymentID] = struct{}{}
	}
	seenReceipts := make(map[string]struct{}, len(state.Receipts))
	for _, receipt := range state.Receipts {
		receipt = receipt.Normalize()
		if err := receipt.Validate(); err != nil {
			return err
		}
		if _, found := seenPayments[receipt.PaymentID]; !found {
			return errors.New("payments abstraction receipt references unknown payment")
		}
		if _, found := seenReceipts[receipt.PaymentID]; found {
			return errors.New("payments abstraction duplicate payment receipt")
		}
		seenReceipts[receipt.PaymentID] = struct{}{}
	}
	seenNonces := make(map[string]struct{}, len(state.Nonces))
	for _, nonce := range state.Nonces {
		nonce = nonce.Normalize()
		if err := nonce.Validate(); err != nil {
			return err
		}
		if _, found := seenPayments[nonce.PaymentID]; !found {
			return errors.New("payments abstraction nonce references unknown payment")
		}
		key := nonce.NonceKey()
		if _, found := seenNonces[key]; found {
			return errors.New("payments abstraction duplicate sender nonce")
		}
		seenNonces[key] = struct{}{}
	}
	return nil
}

func (state PaymentAbstractionState) PaymentByID(paymentID string) (Payment, bool) {
	needle := normalizeHash(paymentID)
	for _, payment := range state.Payments {
		payment = payment.Normalize()
		if payment.PaymentID == needle {
			return payment, true
		}
	}
	return Payment{}, false
}

func (state PaymentAbstractionState) ReceiptByPaymentID(paymentID string) (PaymentAbstractionReceipt, bool) {
	needle := normalizeHash(paymentID)
	for _, receipt := range state.Receipts {
		receipt = receipt.Normalize()
		if receipt.PaymentID == needle {
			return receipt, true
		}
	}
	return PaymentAbstractionReceipt{}, false
}

func (state PaymentAbstractionState) NonceRecord(sourceZone, sender string, nonce uint64) (PaymentNonceRecord, bool) {
	needle := PaymentNonceRecord{SourceZone: sourceZone, Sender: sender, Nonce: nonce}.NonceKey()
	for _, record := range state.Nonces {
		record = record.Normalize()
		if record.NonceKey() == needle {
			return record, true
		}
	}
	return PaymentNonceRecord{}, false
}

func IsPaymentSettlementMode(mode PaymentSettlementMode) bool {
	switch mode {
	case PaymentSettlementDirect, PaymentSettlementStreaming, PaymentSettlementConditional, PaymentSettlementOffchainChannel, PaymentSettlementZoneToZone:
		return true
	default:
		return false
	}
}

func IsPaymentStatus(status PaymentStatus) bool {
	switch status {
	case PaymentStatusQueued, PaymentStatusSettled, PaymentStatusExpired, PaymentStatusDisputed, PaymentStatusRejected:
		return true
	default:
		return false
	}
}

func validatePaymentZoneID(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > MaxTokenLength {
		return fmt.Errorf("%s must be <= %d bytes", field, MaxTokenLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			continue
		}
		return fmt.Errorf("%s contains invalid character %q", field, r)
	}
	return nil
}

func validatePaymentDenom(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("payments abstraction denom is required")
	}
	if len(value) > MaxTokenLength {
		return fmt.Errorf("payments abstraction denom must be <= %d bytes", MaxTokenLength)
	}
	return nil
}

func sortPayments(payments []Payment) {
	sort.SliceStable(payments, func(i, j int) bool {
		return payments[i].Normalize().PaymentID < payments[j].Normalize().PaymentID
	})
}

func sortPaymentAbstractionReceipts(receipts []PaymentAbstractionReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		return receipts[i].Normalize().PaymentID < receipts[j].Normalize().PaymentID
	})
}

func sortPaymentNonceRecords(records []PaymentNonceRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].NonceKey() < records[j].NonceKey()
	})
}
