package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceReceiptStatusText string

const (
	ServiceReceiptStatusAccepted	ServiceReceiptStatusText	= "accepted"
	ServiceReceiptStatusExecuted	ServiceReceiptStatusText	= "executed"
	ServiceReceiptStatusFailed	ServiceReceiptStatusText	= "failed"
	ServiceReceiptStatusExpired	ServiceReceiptStatusText	= "expired"
	ServiceReceiptStatusChallenged	ServiceReceiptStatusText	= "challenged"
	ServiceReceiptStatusSettled	ServiceReceiptStatusText	= "settled"
	ServiceReceiptStatusReverted	ServiceReceiptStatusText	= "reverted"
)

type ServiceReceiptCanonicalView struct {
	CallID		string
	ServiceID	string
	MethodID	string
	Caller		string
	Status		ServiceReceiptStatusText
	RequestHash	string
	ResponseHash	string
	ProofHash	string
	PaymentStatus	string
	GasUsed		uint64
	ProviderID	string
	ExecutedHeight	uint64
	AnchoredHeight	uint64
	ErrorCode	string
	ReceiptHash	string
	ViewHash	string
}

type ServiceCallReplayEntry struct {
	ServiceID	string
	Caller		string
	Nonce		uint64
	CallID		string
	IdempotencyKey	string
	PayloadHash	string
	CreatedHeight	uint64
	DeadlineHeight	uint64
	RetryOf		string
	EntryHash	string
}

type ServiceReceiptTombstone struct {
	CallID			string
	ServiceID		string
	Caller			string
	Nonce			uint64
	IdempotencyKey		string
	ReceiptHash		string
	CreatedHeight		uint64
	ReceiptHeight		uint64
	RetainUntilHeight	uint64
	TombstoneHash		string
}

type ServiceCallReplayIndex struct {
	ProofHorizon	uint64
	Entries		[]ServiceCallReplayEntry
	Tombstones	[]ServiceReceiptTombstone
	IndexHash	string
}

func ComputeUnifiedServiceCallID(ctx coretypes.ServiceConsensusContext, serviceID, caller string, nonce uint64, idempotencyKey, payloadHash string) string {
	return coretypes.ComputeServiceCallID(ctx, coretypes.ServiceCallEnvelope{
		ServiceID:	strings.TrimSpace(serviceID),
		Caller:		strings.TrimSpace(caller),
		Nonce:		nonce,
		IdempotencyKey:	strings.TrimSpace(idempotencyKey),
		PayloadHash:	strings.ToLower(strings.TrimSpace(payloadHash)),
	})
}

func NewServiceReceiptCanonicalView(receipt ServiceReceipt) (ServiceReceiptCanonicalView, error) {
	if err := receipt.Validate(); err != nil {
		return ServiceReceiptCanonicalView{}, err
	}
	status, err := CanonicalServiceReceiptStatus(receipt.Status)
	if err != nil {
		return ServiceReceiptCanonicalView{}, err
	}
	view := ServiceReceiptCanonicalView{
		CallID:		receipt.CallID,
		ServiceID:	receipt.ServiceID,
		MethodID:	receipt.MethodID,
		Caller:		receipt.Caller,
		Status:		status,
		RequestHash:	receipt.RequestHash,
		ResponseHash:	receipt.ResponseHash,
		ProofHash:	receipt.ProofHash,
		PaymentStatus:	strings.ToLower(string(receipt.PaymentStatus)),
		GasUsed:	receipt.GasUsed,
		ProviderID:	receipt.ProviderID,
		ExecutedHeight:	receipt.ExecutedHeight,
		AnchoredHeight:	receipt.AnchoredHeight,
		ErrorCode:	receipt.ErrorCode,
		ReceiptHash:	receipt.ReceiptHash,
	}
	view.ViewHash = ComputeServiceReceiptCanonicalViewHash(view)
	return view, view.Validate()
}

func (view ServiceReceiptCanonicalView) Validate() error {
	if err := coretypes.ValidateHash("services receipt view call id", view.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services receipt view service id", view.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services receipt view method id", view.MethodID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("services receipt view caller", view.Caller); err != nil {
		return err
	}
	if !IsServiceReceiptStatusText(view.Status) {
		return fmt.Errorf("services receipt view unknown status %q", view.Status)
	}
	if err := coretypes.ValidateHash("services receipt view request hash", view.RequestHash); err != nil {
		return err
	}
	if view.ResponseHash != "" {
		if err := coretypes.ValidateHash("services receipt view response hash", view.ResponseHash); err != nil {
			return err
		}
	}
	if view.ProofHash != "" {
		if err := coretypes.ValidateHash("services receipt view proof hash", view.ProofHash); err != nil {
			return err
		}
	}
	if err := validateInterfaceToken("services receipt view payment status", view.PaymentStatus); err != nil {
		return err
	}
	if view.ProviderID != "" {
		if err := validateInterfaceToken("services receipt view provider id", view.ProviderID); err != nil {
			return err
		}
	}
	if view.ExecutedHeight == 0 || view.AnchoredHeight == 0 {
		return errors.New("services receipt view heights must be positive")
	}
	if view.ErrorCode != "" {
		if err := validateInterfaceToken("services receipt view error code", view.ErrorCode); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("services receipt view receipt hash", view.ReceiptHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services receipt view hash", view.ViewHash); err != nil {
		return err
	}
	if expected := ComputeServiceReceiptCanonicalViewHash(view); view.ViewHash != expected {
		return fmt.Errorf("services receipt view hash mismatch: expected %s", expected)
	}
	return nil
}

func CanonicalServiceReceiptStatus(status coretypes.ServiceCallStatus) (ServiceReceiptStatusText, error) {
	switch status {
	case coretypes.ServiceCallStatusAccepted:
		return ServiceReceiptStatusAccepted, nil
	case coretypes.ServiceCallStatusExecuted:
		return ServiceReceiptStatusExecuted, nil
	case coretypes.ServiceCallStatusFailed:
		return ServiceReceiptStatusFailed, nil
	case coretypes.ServiceCallStatusExpired:
		return ServiceReceiptStatusExpired, nil
	case coretypes.ServiceCallStatusChallenged:
		return ServiceReceiptStatusChallenged, nil
	case coretypes.ServiceCallStatusSettled:
		return ServiceReceiptStatusSettled, nil
	case coretypes.ServiceCallStatusReverted:
		return ServiceReceiptStatusReverted, nil
	default:
		return "", fmt.Errorf("services receipt unknown status %q", status)
	}
}

func IsServiceReceiptStatusText(status ServiceReceiptStatusText) bool {
	switch status {
	case ServiceReceiptStatusAccepted, ServiceReceiptStatusExecuted, ServiceReceiptStatusFailed,
		ServiceReceiptStatusExpired, ServiceReceiptStatusChallenged, ServiceReceiptStatusSettled,
		ServiceReceiptStatusReverted:
		return true
	default:
		return false
	}
}

func NewServiceCallReplayIndex(proofHorizon uint64) (ServiceCallReplayIndex, error) {
	if proofHorizon == 0 {
		return ServiceCallReplayIndex{}, errors.New("services replay proof horizon must be positive")
	}
	index := ServiceCallReplayIndex{ProofHorizon: proofHorizon}
	index.IndexHash = ComputeServiceCallReplayIndexHash(index)
	return index, index.Validate()
}

func AcceptUnifiedServiceCall(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, index ServiceCallReplayIndex, call UnifiedServiceCall) (ServiceCallReplayIndex, error) {
	if err := index.Validate(); err != nil {
		return ServiceCallReplayIndex{}, err
	}
	if err := ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, call); err != nil {
		return ServiceCallReplayIndex{}, err
	}
	if existing, found := index.EntryByServiceCallerNonce(call.TargetService, call.Caller, call.Nonce); found {
		return ServiceCallReplayIndex{}, fmt.Errorf("services replay nonce already used by call %s", existing.CallID)
	}
	if call.Kind == coretypes.ServiceCallKindRetry {
		if call.IdempotencyKey == "" {
			return ServiceCallReplayIndex{}, errors.New("services replay retry requires idempotency key")
		}
		if call.RetryOf == "" {
			return ServiceCallReplayIndex{}, errors.New("services replay retry must reference original call id")
		}
		if !index.ContainsCallID(call.RetryOf) {
			return ServiceCallReplayIndex{}, errors.New("services replay retry references unknown original call")
		}
	}
	entry := ServiceCallReplayEntry{
		ServiceID:	call.TargetService,
		Caller:		call.Caller,
		Nonce:		call.Nonce,
		CallID:		call.CallID,
		IdempotencyKey:	call.IdempotencyKey,
		PayloadHash:	call.PayloadHash,
		CreatedHeight:	call.CreatedHeight,
		DeadlineHeight:	call.DeadlineHeight,
		RetryOf:	call.RetryOf,
	}
	entry.EntryHash = ComputeServiceCallReplayEntryHash(entry)
	next := cloneServiceCallReplayIndex(index)
	next.Entries = append(next.Entries, entry)
	sortServiceCallReplayEntries(next.Entries)
	next.IndexHash = ComputeServiceCallReplayIndexHash(next)
	return next, next.Validate()
}

func TombstoneServiceReceipt(ctx coretypes.ServiceConsensusContext, index ServiceCallReplayIndex, call UnifiedServiceCall, receipt ServiceReceipt) (ServiceCallReplayIndex, ServiceReceiptTombstone, error) {
	if err := index.Validate(); err != nil {
		return ServiceCallReplayIndex{}, ServiceReceiptTombstone{}, err
	}
	if err := call.ValidateBasic(ctx); err != nil {
		return ServiceCallReplayIndex{}, ServiceReceiptTombstone{}, err
	}
	if err := receipt.Validate(); err != nil {
		return ServiceCallReplayIndex{}, ServiceReceiptTombstone{}, err
	}
	if receipt.CallID != call.CallID || receipt.ServiceID != call.TargetService || receipt.Caller != call.Caller {
		return ServiceCallReplayIndex{}, ServiceReceiptTombstone{}, errors.New("services receipt tombstone call mismatch")
	}
	if _, found := index.TombstoneByCallID(call.CallID); found {
		return ServiceCallReplayIndex{}, ServiceReceiptTombstone{}, fmt.Errorf("services receipt tombstone %s already exists", call.CallID)
	}
	receiptHeight := receipt.AnchoredHeight
	if receiptHeight < receipt.ExecutedHeight {
		receiptHeight = receipt.ExecutedHeight
	}
	if receiptHeight < call.CreatedHeight {
		receiptHeight = call.CreatedHeight
	}
	tombstone := ServiceReceiptTombstone{
		CallID:			call.CallID,
		ServiceID:		call.TargetService,
		Caller:			call.Caller,
		Nonce:			call.Nonce,
		IdempotencyKey:		call.IdempotencyKey,
		ReceiptHash:		receipt.ReceiptHash,
		CreatedHeight:		call.CreatedHeight,
		ReceiptHeight:		receiptHeight,
		RetainUntilHeight:	receiptHeight + index.ProofHorizon,
	}
	tombstone.TombstoneHash = ComputeServiceReceiptTombstoneHash(tombstone)
	next := cloneServiceCallReplayIndex(index)
	next.Tombstones = append(next.Tombstones, tombstone)
	sortServiceReceiptTombstones(next.Tombstones)
	next.IndexHash = ComputeServiceCallReplayIndexHash(next)
	return next, tombstone, next.Validate()
}

func PruneExpiredReceiptTombstones(index ServiceCallReplayIndex, height uint64) (ServiceCallReplayIndex, error) {
	if height == 0 {
		return ServiceCallReplayIndex{}, errors.New("services receipt tombstone prune height must be positive")
	}
	if err := index.Validate(); err != nil {
		return ServiceCallReplayIndex{}, err
	}
	next := cloneServiceCallReplayIndex(index)
	kept := make([]ServiceReceiptTombstone, 0, len(index.Tombstones))
	for _, tombstone := range index.Tombstones {
		if tombstone.RetainUntilHeight >= height {
			kept = append(kept, tombstone)
		}
	}
	next.Tombstones = kept
	sortServiceReceiptTombstones(next.Tombstones)
	next.IndexHash = ComputeServiceCallReplayIndexHash(next)
	return next, next.Validate()
}

func (index ServiceCallReplayIndex) Validate() error {
	if index.ProofHorizon == 0 {
		return errors.New("services replay proof horizon must be positive")
	}
	if err := validateServiceCallReplayEntries(index.Entries); err != nil {
		return err
	}
	if err := validateServiceReceiptTombstones(index.Tombstones); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services replay index hash", index.IndexHash); err != nil {
		return err
	}
	if expected := ComputeServiceCallReplayIndexHash(index); index.IndexHash != expected {
		return fmt.Errorf("services replay index hash mismatch: expected %s", expected)
	}
	return nil
}

func (entry ServiceCallReplayEntry) Validate() error {
	if err := validateInterfaceToken("services replay entry service id", entry.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("services replay entry caller", entry.Caller); err != nil {
		return err
	}
	if entry.Nonce == 0 {
		return errors.New("services replay entry nonce must be positive")
	}
	if err := coretypes.ValidateHash("services replay entry call id", entry.CallID); err != nil {
		return err
	}
	if entry.IdempotencyKey != "" {
		if err := validateInterfaceToken("services replay entry idempotency key", entry.IdempotencyKey); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("services replay entry payload hash", entry.PayloadHash); err != nil {
		return err
	}
	if entry.CreatedHeight == 0 || entry.DeadlineHeight < entry.CreatedHeight {
		return errors.New("services replay entry heights are invalid")
	}
	if entry.RetryOf != "" {
		if err := coretypes.ValidateHash("services replay entry retry original", entry.RetryOf); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("services replay entry hash", entry.EntryHash); err != nil {
		return err
	}
	if expected := ComputeServiceCallReplayEntryHash(entry); entry.EntryHash != expected {
		return fmt.Errorf("services replay entry hash mismatch: expected %s", expected)
	}
	return nil
}

func (tombstone ServiceReceiptTombstone) Validate() error {
	if err := coretypes.ValidateHash("services receipt tombstone call id", tombstone.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services receipt tombstone service id", tombstone.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("services receipt tombstone caller", tombstone.Caller); err != nil {
		return err
	}
	if tombstone.Nonce == 0 {
		return errors.New("services receipt tombstone nonce must be positive")
	}
	if tombstone.IdempotencyKey != "" {
		if err := validateInterfaceToken("services receipt tombstone idempotency key", tombstone.IdempotencyKey); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("services receipt tombstone receipt hash", tombstone.ReceiptHash); err != nil {
		return err
	}
	if tombstone.CreatedHeight == 0 || tombstone.ReceiptHeight < tombstone.CreatedHeight || tombstone.RetainUntilHeight < tombstone.ReceiptHeight {
		return errors.New("services receipt tombstone heights are invalid")
	}
	if err := coretypes.ValidateHash("services receipt tombstone hash", tombstone.TombstoneHash); err != nil {
		return err
	}
	if expected := ComputeServiceReceiptTombstoneHash(tombstone); tombstone.TombstoneHash != expected {
		return fmt.Errorf("services receipt tombstone hash mismatch: expected %s", expected)
	}
	return nil
}

func (index ServiceCallReplayIndex) EntryByServiceCallerNonce(serviceID, caller string, nonce uint64) (ServiceCallReplayEntry, bool) {
	for _, entry := range index.Entries {
		if entry.ServiceID == serviceID && entry.Caller == caller && entry.Nonce == nonce {
			return entry, true
		}
	}
	return ServiceCallReplayEntry{}, false
}

func (index ServiceCallReplayIndex) TombstoneByCallID(callID string) (ServiceReceiptTombstone, bool) {
	callID = strings.ToLower(strings.TrimSpace(callID))
	for _, tombstone := range index.Tombstones {
		if tombstone.CallID == callID {
			return tombstone, true
		}
	}
	return ServiceReceiptTombstone{}, false
}

func (index ServiceCallReplayIndex) ContainsCallID(callID string) bool {
	callID = strings.ToLower(strings.TrimSpace(callID))
	for _, entry := range index.Entries {
		if entry.CallID == callID {
			return true
		}
	}
	_, found := index.TombstoneByCallID(callID)
	return found
}

func ComputeServiceReceiptCanonicalViewHash(view ServiceReceiptCanonicalView) string {
	return servicesHashParts(
		"aetra-services-receipt-view-v1",
		view.CallID,
		view.ServiceID,
		view.MethodID,
		view.Caller,
		string(view.Status),
		view.RequestHash,
		view.ResponseHash,
		view.ProofHash,
		view.PaymentStatus,
		fmt.Sprint(view.GasUsed),
		view.ProviderID,
		fmt.Sprint(view.ExecutedHeight),
		fmt.Sprint(view.AnchoredHeight),
		view.ErrorCode,
		view.ReceiptHash,
	)
}

func ComputeServiceCallReplayEntryHash(entry ServiceCallReplayEntry) string {
	return servicesHashParts(
		"aetra-services-replay-entry-v1",
		entry.ServiceID,
		entry.Caller,
		fmt.Sprint(entry.Nonce),
		entry.CallID,
		entry.IdempotencyKey,
		entry.PayloadHash,
		fmt.Sprint(entry.CreatedHeight),
		fmt.Sprint(entry.DeadlineHeight),
		entry.RetryOf,
	)
}

func ComputeServiceReceiptTombstoneHash(tombstone ServiceReceiptTombstone) string {
	return servicesHashParts(
		"aetra-services-receipt-tombstone-v1",
		tombstone.CallID,
		tombstone.ServiceID,
		tombstone.Caller,
		fmt.Sprint(tombstone.Nonce),
		tombstone.IdempotencyKey,
		tombstone.ReceiptHash,
		fmt.Sprint(tombstone.CreatedHeight),
		fmt.Sprint(tombstone.ReceiptHeight),
		fmt.Sprint(tombstone.RetainUntilHeight),
	)
}

func ComputeServiceCallReplayIndexHash(index ServiceCallReplayIndex) string {
	entries := append([]ServiceCallReplayEntry(nil), index.Entries...)
	tombstones := append([]ServiceReceiptTombstone(nil), index.Tombstones...)
	sortServiceCallReplayEntries(entries)
	sortServiceReceiptTombstones(tombstones)
	parts := []string{
		"aetra-services-replay-index-v1",
		fmt.Sprint(index.ProofHorizon),
		fmt.Sprint(len(entries)),
		fmt.Sprint(len(tombstones)),
	}
	for _, entry := range entries {
		parts = append(parts, entry.EntryHash)
	}
	for _, tombstone := range tombstones {
		parts = append(parts, tombstone.TombstoneHash)
	}
	return servicesHashParts(parts...)
}

func validateServiceCallReplayEntries(entries []ServiceCallReplayEntry) error {
	previousKey := ""
	seenNonce := map[string]struct{}{}
	seenCall := map[string]struct{}{}
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		nonceKey := fmt.Sprintf("%s/%s/%020d", entry.ServiceID, entry.Caller, entry.Nonce)
		if _, found := seenNonce[nonceKey]; found {
			return fmt.Errorf("services replay duplicate service caller nonce %s", nonceKey)
		}
		seenNonce[nonceKey] = struct{}{}
		if _, found := seenCall[entry.CallID]; found {
			return fmt.Errorf("services replay duplicate call id %s", entry.CallID)
		}
		seenCall[entry.CallID] = struct{}{}
		sortKey := nonceKey + "/" + entry.CallID
		if previousKey != "" && previousKey >= sortKey {
			return errors.New("services replay entries must be sorted canonically")
		}
		previousKey = sortKey
	}
	return nil
}

func validateServiceReceiptTombstones(tombstones []ServiceReceiptTombstone) error {
	previous := ""
	seen := map[string]struct{}{}
	for _, tombstone := range tombstones {
		if err := tombstone.Validate(); err != nil {
			return err
		}
		if _, found := seen[tombstone.CallID]; found {
			return fmt.Errorf("services receipt duplicate tombstone %s", tombstone.CallID)
		}
		seen[tombstone.CallID] = struct{}{}
		if previous != "" && previous >= tombstone.CallID {
			return errors.New("services receipt tombstones must be sorted canonically")
		}
		previous = tombstone.CallID
	}
	return nil
}

func cloneServiceCallReplayIndex(index ServiceCallReplayIndex) ServiceCallReplayIndex {
	index.Entries = append([]ServiceCallReplayEntry(nil), index.Entries...)
	index.Tombstones = append([]ServiceReceiptTombstone(nil), index.Tombstones...)
	return index
}

func sortServiceCallReplayEntries(entries []ServiceCallReplayEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		left := fmt.Sprintf("%s/%s/%020d/%s", entries[i].ServiceID, entries[i].Caller, entries[i].Nonce, entries[i].CallID)
		right := fmt.Sprintf("%s/%s/%020d/%s", entries[j].ServiceID, entries[j].Caller, entries[j].Nonce, entries[j].CallID)
		return left < right
	})
}

func sortServiceReceiptTombstones(tombstones []ServiceReceiptTombstone) {
	sort.SliceStable(tombstones, func(i, j int) bool { return tombstones[i].CallID < tombstones[j].CallID })
}
