package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type RetryPaymentPolicy string

const (
	RetryPaymentOriginalOnly	RetryPaymentPolicy	= "original_call_only"
	RetryPaymentChargeAttempts	RetryPaymentPolicy	= "charge_attempts"
)

type ServiceRetryPolicy struct {
	MaxAttempts		uint32
	MaxDeadlineDelta	uint64
	PaymentPolicy		RetryPaymentPolicy
	PolicyHash		string
}

type ServiceRetryAttempt struct {
	OriginalCallID		string
	RetryCallID		string
	AttemptNumber		uint32
	IdempotencyKey		string
	CreatedHeight		uint64
	DeadlineHeight		uint64
	ChargeAttempt		bool
	PaymentChargeHash	string
	AttemptHash		string
}

type ServiceRetryReceiptLink struct {
	OriginalCallID	string
	RetryCallID	string
	ReceiptHash	string
	LinkHash	string
}

type ReceiptCommitment struct {
	ReceiptID	string
	ReceiptHash	string
	CommitmentHash	string
}

type DeterministicReceiptRequirements struct {
	OnChainCallIDs		[]string
	OffChainResultCallIDs	[]string
	MixedSettlementCallIDs	[]string
}

type DeterministicReceiptRoots struct {
	ServiceReceiptsRoot	string
	CallReceiptsRoot	string
	PaymentReceiptsRoot	string
	StorageReceiptsRoot	string
	ReceiptRootsHash	string
}

func NewServiceRetryPolicy(policy ServiceRetryPolicy) (ServiceRetryPolicy, error) {
	if policy.PolicyHash != "" {
		return ServiceRetryPolicy{}, errors.New("services retry policy hash must be empty before construction")
	}
	if err := policy.ValidateFormat(); err != nil {
		return ServiceRetryPolicy{}, err
	}
	policy.PolicyHash = ComputeServiceRetryPolicyHash(policy)
	return policy, policy.Validate()
}

func NewServiceRetryAttempt(ctx coretypes.ServiceConsensusContext, policy ServiceRetryPolicy, original UnifiedServiceCall, retry UnifiedServiceCall, previous []ServiceRetryAttempt) (ServiceRetryAttempt, error) {
	if err := ctx.Validate(); err != nil {
		return ServiceRetryAttempt{}, err
	}
	if err := policy.Validate(); err != nil {
		return ServiceRetryAttempt{}, err
	}
	if err := original.ValidateBasic(ctx); err != nil {
		return ServiceRetryAttempt{}, err
	}
	if err := retry.ValidateBasic(ctx); err != nil {
		return ServiceRetryAttempt{}, err
	}
	if err := validateServiceRetryAttempts(previous); err != nil {
		return ServiceRetryAttempt{}, err
	}
	if retry.Kind != coretypes.ServiceCallKindRetry {
		return ServiceRetryAttempt{}, errors.New("services retry attempt requires retry call kind")
	}
	if retry.RetryOf != original.CallID {
		return ServiceRetryAttempt{}, errors.New("services retry attempt must reference original call id")
	}
	if retry.IdempotencyKey == "" {
		return ServiceRetryAttempt{}, errors.New("services retry attempt requires idempotency key")
	}
	if uint32(len(previous)) >= policy.MaxAttempts {
		return ServiceRetryAttempt{}, errors.New("services retry attempt count exceeds policy")
	}
	if retry.DeadlineHeight > original.CreatedHeight+policy.MaxDeadlineDelta {
		return ServiceRetryAttempt{}, errors.New("services retry attempt deadline exceeds policy")
	}
	chargeAttempt := policy.PaymentPolicy == RetryPaymentChargeAttempts
	attempt := ServiceRetryAttempt{
		OriginalCallID:	original.CallID,
		RetryCallID:	retry.CallID,
		AttemptNumber:	uint32(len(previous)) + 1,
		IdempotencyKey:	retry.IdempotencyKey,
		CreatedHeight:	retry.CreatedHeight,
		DeadlineHeight:	retry.DeadlineHeight,
		ChargeAttempt:	chargeAttempt,
	}
	attempt.PaymentChargeHash = ComputeServiceRetryPaymentChargeHash(policy, original, retry, chargeAttempt)
	attempt.AttemptHash = ComputeServiceRetryAttemptHash(attempt)
	return attempt, attempt.Validate()
}

func NewServiceRetryReceiptLink(original UnifiedServiceCall, retry UnifiedServiceCall, receipt ServiceReceipt) (ServiceRetryReceiptLink, error) {
	if err := receipt.Validate(); err != nil {
		return ServiceRetryReceiptLink{}, err
	}
	if retry.Kind != coretypes.ServiceCallKindRetry || retry.RetryOf != original.CallID {
		return ServiceRetryReceiptLink{}, errors.New("services retry receipt link requires retry referencing original call")
	}
	if receipt.CallID != retry.CallID {
		return ServiceRetryReceiptLink{}, errors.New("services retry receipt link call mismatch")
	}
	link := ServiceRetryReceiptLink{
		OriginalCallID:	original.CallID,
		RetryCallID:	retry.CallID,
		ReceiptHash:	receipt.ReceiptHash,
	}
	link.LinkHash = ComputeServiceRetryReceiptLinkHash(link)
	return link, link.Validate()
}

func NewReceiptCommitment(receiptID, receiptHash string) (ReceiptCommitment, error) {
	commitment := ReceiptCommitment{
		ReceiptID:	strings.TrimSpace(receiptID),
		ReceiptHash:	strings.ToLower(strings.TrimSpace(receiptHash)),
	}
	if err := commitment.ValidateFormat(); err != nil {
		return ReceiptCommitment{}, err
	}
	commitment.CommitmentHash = ComputeReceiptCommitmentHash(commitment)
	return commitment, commitment.Validate()
}

func BuildDeterministicReceiptRoots(serviceReceipts []ServiceReceipt, paymentReceipts []ReceiptCommitment, storageReceipts []ReceiptCommitment, requirements DeterministicReceiptRequirements) (DeterministicReceiptRoots, error) {
	orderedService := cloneServiceReceipts(serviceReceipts)
	sortServiceReceipts(orderedService)
	if err := validateServiceReceipts(orderedService); err != nil {
		return DeterministicReceiptRoots{}, err
	}
	if err := requirements.ValidateForReceipts(orderedService); err != nil {
		return DeterministicReceiptRoots{}, err
	}
	serviceRoot, err := coretypes.ComputeServiceReceiptsRoot(orderedService)
	if err != nil {
		return DeterministicReceiptRoots{}, err
	}
	callRoot, err := ComputeCanonicalCallReceiptsRoot(orderedService)
	if err != nil {
		return DeterministicReceiptRoots{}, err
	}
	roots := DeterministicReceiptRoots{
		ServiceReceiptsRoot:	serviceRoot,
		CallReceiptsRoot:	callRoot,
		PaymentReceiptsRoot:	ComputeReceiptCommitmentRoot("payment_receipts_root", paymentReceipts),
		StorageReceiptsRoot:	ComputeReceiptCommitmentRoot("storage_receipts_root", storageReceipts),
	}
	roots.ReceiptRootsHash = ComputeDeterministicReceiptRootsHash(roots)
	return roots, roots.Validate()
}

func (policy ServiceRetryPolicy) ValidateFormat() error {
	if policy.MaxAttempts == 0 {
		return errors.New("services retry policy max attempts must be positive")
	}
	if policy.MaxDeadlineDelta == 0 {
		return errors.New("services retry policy max deadline delta must be positive")
	}
	if !IsRetryPaymentPolicy(policy.PaymentPolicy) {
		return fmt.Errorf("services retry policy unknown payment policy %q", policy.PaymentPolicy)
	}
	if policy.PolicyHash != "" {
		return coretypes.ValidateHash("services retry policy hash", policy.PolicyHash)
	}
	return nil
}

func (policy ServiceRetryPolicy) Validate() error {
	if err := policy.ValidateFormat(); err != nil {
		return err
	}
	if policy.PolicyHash == "" {
		return errors.New("services retry policy hash is required")
	}
	if expected := ComputeServiceRetryPolicyHash(policy); policy.PolicyHash != expected {
		return fmt.Errorf("services retry policy hash mismatch: expected %s", expected)
	}
	return nil
}

func (attempt ServiceRetryAttempt) Validate() error {
	if err := coretypes.ValidateHash("services retry attempt original call id", attempt.OriginalCallID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services retry attempt call id", attempt.RetryCallID); err != nil {
		return err
	}
	if attempt.AttemptNumber == 0 {
		return errors.New("services retry attempt number must be positive")
	}
	if err := validateInterfaceToken("services retry attempt idempotency key", attempt.IdempotencyKey); err != nil {
		return err
	}
	if attempt.CreatedHeight == 0 || attempt.DeadlineHeight < attempt.CreatedHeight {
		return errors.New("services retry attempt heights are invalid")
	}
	if err := coretypes.ValidateHash("services retry payment charge hash", attempt.PaymentChargeHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services retry attempt hash", attempt.AttemptHash); err != nil {
		return err
	}
	if expected := ComputeServiceRetryAttemptHash(attempt); attempt.AttemptHash != expected {
		return fmt.Errorf("services retry attempt hash mismatch: expected %s", expected)
	}
	return nil
}

func (link ServiceRetryReceiptLink) Validate() error {
	if err := coretypes.ValidateHash("services retry receipt original call id", link.OriginalCallID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services retry receipt retry call id", link.RetryCallID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services retry receipt hash", link.ReceiptHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services retry receipt link hash", link.LinkHash); err != nil {
		return err
	}
	if expected := ComputeServiceRetryReceiptLinkHash(link); link.LinkHash != expected {
		return fmt.Errorf("services retry receipt link hash mismatch: expected %s", expected)
	}
	return nil
}

func (commitment ReceiptCommitment) ValidateFormat() error {
	if err := validateInterfaceToken("services receipt commitment id", commitment.ReceiptID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services receipt commitment hash", commitment.ReceiptHash); err != nil {
		return err
	}
	if commitment.CommitmentHash != "" {
		return coretypes.ValidateHash("services receipt commitment hash", commitment.CommitmentHash)
	}
	return nil
}

func (commitment ReceiptCommitment) Validate() error {
	if err := commitment.ValidateFormat(); err != nil {
		return err
	}
	if commitment.CommitmentHash == "" {
		return errors.New("services receipt commitment hash is required")
	}
	if expected := ComputeReceiptCommitmentHash(commitment); commitment.CommitmentHash != expected {
		return fmt.Errorf("services receipt commitment hash mismatch: expected %s", expected)
	}
	return nil
}

func (requirements DeterministicReceiptRequirements) ValidateForReceipts(receipts []ServiceReceipt) error {
	byCall := map[string]struct{}{}
	for _, receipt := range receipts {
		byCall[receipt.CallID] = struct{}{}
	}
	for label, calls := range map[string][]string{
		"on-chain service call":	requirements.OnChainCallIDs,
		"anchored off-chain result":	requirements.OffChainResultCallIDs,
		"mixed-service settlement":	requirements.MixedSettlementCallIDs,
	} {
		for _, callID := range normalizeReceiptCallIDs(calls) {
			if err := coretypes.ValidateHash("services deterministic receipt required "+label, callID); err != nil {
				return err
			}
			if _, found := byCall[callID]; !found {
				return fmt.Errorf("services deterministic receipts missing %s receipt %s", label, callID)
			}
		}
	}
	return nil
}

func (roots DeterministicReceiptRoots) Validate() error {
	if err := coretypes.ValidateHash("service_receipts_root", roots.ServiceReceiptsRoot); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("call_receipts_root", roots.CallReceiptsRoot); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("payment_receipts_root", roots.PaymentReceiptsRoot); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("storage_receipts_root", roots.StorageReceiptsRoot); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("deterministic receipt roots hash", roots.ReceiptRootsHash); err != nil {
		return err
	}
	if expected := ComputeDeterministicReceiptRootsHash(roots); roots.ReceiptRootsHash != expected {
		return fmt.Errorf("deterministic receipt roots hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeCanonicalCallReceiptsRoot(receipts []ServiceReceipt) (string, error) {
	views := make([]ServiceReceiptCanonicalView, 0, len(receipts))
	for _, receipt := range receipts {
		view, err := NewServiceReceiptCanonicalView(receipt)
		if err != nil {
			return "", err
		}
		views = append(views, view)
	}
	sort.SliceStable(views, func(i, j int) bool { return views[i].CallID < views[j].CallID })
	parts := []string{"call_receipts_root", fmt.Sprint(len(views))}
	for _, view := range views {
		parts = append(parts, view.ViewHash)
	}
	return servicesHashParts(parts...), nil
}

func ComputeReceiptCommitmentRoot(label string, commitments []ReceiptCommitment) string {
	ordered := normalizeReceiptCommitments(commitments)
	parts := []string{label, fmt.Sprint(len(ordered))}
	for _, commitment := range ordered {
		parts = append(parts, commitment.CommitmentHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRetryPolicyHash(policy ServiceRetryPolicy) string {
	return servicesHashParts("services-retry-policy-v1", fmt.Sprint(policy.MaxAttempts), fmt.Sprint(policy.MaxDeadlineDelta), string(policy.PaymentPolicy))
}

func ComputeServiceRetryPaymentChargeHash(policy ServiceRetryPolicy, original UnifiedServiceCall, retry UnifiedServiceCall, chargeAttempt bool) string {
	paymentHash := original.Payment.PaymentHash
	if chargeAttempt {
		paymentHash = retry.Payment.PaymentHash
	}
	return servicesHashParts("services-retry-payment-charge-v1", policy.PolicyHash, original.CallID, retry.CallID, fmt.Sprint(chargeAttempt), paymentHash)
}

func ComputeServiceRetryAttemptHash(attempt ServiceRetryAttempt) string {
	return servicesHashParts(
		"services-retry-attempt-v1",
		attempt.OriginalCallID,
		attempt.RetryCallID,
		fmt.Sprint(attempt.AttemptNumber),
		attempt.IdempotencyKey,
		fmt.Sprint(attempt.CreatedHeight),
		fmt.Sprint(attempt.DeadlineHeight),
		fmt.Sprint(attempt.ChargeAttempt),
		attempt.PaymentChargeHash,
	)
}

func ComputeServiceRetryReceiptLinkHash(link ServiceRetryReceiptLink) string {
	return servicesHashParts("services-retry-receipt-link-v1", link.OriginalCallID, link.RetryCallID, link.ReceiptHash)
}

func ComputeReceiptCommitmentHash(commitment ReceiptCommitment) string {
	return servicesHashParts("services-receipt-commitment-v1", commitment.ReceiptID, commitment.ReceiptHash)
}

func ComputeDeterministicReceiptRootsHash(roots DeterministicReceiptRoots) string {
	return servicesHashParts("deterministic_receipt_roots", roots.ServiceReceiptsRoot, roots.CallReceiptsRoot, roots.PaymentReceiptsRoot, roots.StorageReceiptsRoot)
}

func IsRetryPaymentPolicy(policy RetryPaymentPolicy) bool {
	switch policy {
	case RetryPaymentOriginalOnly, RetryPaymentChargeAttempts:
		return true
	default:
		return false
	}
}

func validateServiceRetryAttempts(attempts []ServiceRetryAttempt) error {
	previous := uint32(0)
	seen := map[string]struct{}{}
	for _, attempt := range attempts {
		if err := attempt.Validate(); err != nil {
			return err
		}
		if attempt.AttemptNumber <= previous {
			return errors.New("services retry attempts must be ordered by attempt number")
		}
		previous = attempt.AttemptNumber
		if _, found := seen[attempt.RetryCallID]; found {
			return fmt.Errorf("services retry duplicate attempt call %s", attempt.RetryCallID)
		}
		seen[attempt.RetryCallID] = struct{}{}
	}
	return nil
}

func validateServiceReceipts(receipts []ServiceReceipt) error {
	previous := ""
	for _, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= receipt.CallID {
			return errors.New("services deterministic receipts must be sorted by call id")
		}
		previous = receipt.CallID
	}
	return nil
}

func cloneServiceReceipts(receipts []ServiceReceipt) []ServiceReceipt {
	out := make([]ServiceReceipt, len(receipts))
	copy(out, receipts)
	return out
}

func sortServiceReceipts(receipts []ServiceReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool { return receipts[i].CallID < receipts[j].CallID })
}

func normalizeReceiptCommitments(commitments []ReceiptCommitment) []ReceiptCommitment {
	out := append([]ReceiptCommitment(nil), commitments...)
	for i := range out {
		out[i].ReceiptID = strings.TrimSpace(out[i].ReceiptID)
		out[i].ReceiptHash = strings.ToLower(strings.TrimSpace(out[i].ReceiptHash))
		out[i].CommitmentHash = strings.ToLower(strings.TrimSpace(out[i].CommitmentHash))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReceiptID < out[j].ReceiptID })
	return out
}

func normalizeReceiptCallIDs(callIDs []string) []string {
	out := append([]string(nil), callIDs...)
	for i := range out {
		out[i] = strings.ToLower(strings.TrimSpace(out[i]))
	}
	sort.Strings(out)
	return out
}
