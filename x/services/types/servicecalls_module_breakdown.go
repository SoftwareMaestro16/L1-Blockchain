package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type XServiceCallsStateObject string
type XServiceCallsMessageName string
type XServiceCallsQueryName string
type XServiceCallsFailureMode string
type XServiceCallsIntegrationPoint string

const (
	XServiceCallsStateServiceCall		XServiceCallsStateObject	= "ServiceCall"
	XServiceCallsStateCallNonce		XServiceCallsStateObject	= "CallNonce"
	XServiceCallsStateIdempotencyRecord	XServiceCallsStateObject	= "IdempotencyRecord"
	XServiceCallsStateCallbackRecord	XServiceCallsStateObject	= "CallbackRecord"
	XServiceCallsStateCallReceipt		XServiceCallsStateObject	= "CallReceipt"

	XServiceCallsMsgSubmitServiceCall	XServiceCallsMessageName	= "MsgSubmitServiceCall"
	XServiceCallsMsgAnchorServiceResult	XServiceCallsMessageName	= "MsgAnchorServiceResult"
	XServiceCallsMsgRetryServiceCall	XServiceCallsMessageName	= "MsgRetryServiceCall"
	XServiceCallsMsgSubmitCallback		XServiceCallsMessageName	= "MsgSubmitCallback"
	XServiceCallsMsgExpireServiceCall	XServiceCallsMessageName	= "MsgExpireServiceCall"

	XServiceCallsQueryServiceCall	XServiceCallsQueryName	= "QueryServiceCall"
	XServiceCallsQueryCallReceipt	XServiceCallsQueryName	= "QueryCallReceipt"
	XServiceCallsQueryCallsByCaller	XServiceCallsQueryName	= "QueryCallsByCaller"
	XServiceCallsQueryCallProof	XServiceCallsQueryName	= "QueryCallProof"

	XServiceCallsFailureNonceReplay			XServiceCallsFailureMode	= "nonce_replay"
	XServiceCallsFailureDuplicateIdempotencyKey	XServiceCallsFailureMode	= "duplicate_idempotency_key_misuse"
	XServiceCallsFailureCallbackMismatch		XServiceCallsFailureMode	= "callback_mismatch"
	XServiceCallsFailureExpiredCallAnchoredLate	XServiceCallsFailureMode	= "expired_call_anchored_late"
	XServiceCallsFailureResultHashMismatch		XServiceCallsFailureMode	= "result_hash_mismatch"

	XServiceCallsIntegrationServices		XServiceCallsIntegrationPoint	= "x/services"
	XServiceCallsIntegrationServicePayments		XServiceCallsIntegrationPoint	= "x/servicepayments"
	XServiceCallsIntegrationServiceReceipts		XServiceCallsIntegrationPoint	= "x/servicereceipts"
	XServiceCallsIntegrationABCIProposalHandling	XServiceCallsIntegrationPoint	= "abci_plus_proposal_handling"
)

type XServiceCallsFailureCoverage struct {
	Mode	XServiceCallsFailureMode
	Guard	string
	Scope	string
}

type XServiceCallsModuleBreakdown struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]XServiceCallsStateObject
	Messages		[]XServiceCallsMessageName
	Queries			[]XServiceCallsQueryName
	FailureModes		[]XServiceCallsFailureCoverage
	IntegrationPoints	[]XServiceCallsIntegrationPoint
	BreakdownHash		string
}

type MsgSubmitServiceCall struct {
	Authority	string
	Call		UnifiedServiceCall
	MessageHash	string
}

type MsgAnchorServiceResult struct {
	Authority		string
	CallID			string
	ExpectedResultHash	string
	Outcome			ServiceExecutionOutcome
	MessageHash		string
}

type MsgRetryServiceCall struct {
	Authority	string
	OriginalCallID	string
	RetryCall	UnifiedServiceCall
	MessageHash	string
}

type MsgSubmitCallback struct {
	Authority	string
	Callback	UnifiedServiceCallback
	MessageHash	string
}

type MsgExpireServiceCall struct {
	Authority	string
	CallID		string
	ServiceID	string
	ExpireHeight	uint64
	MessageHash	string
}

type QueryServiceCall struct {
	CallID string
}

type QueryServiceCallResponse struct {
	Call	UnifiedServiceCall
	Found	bool
}

type QueryCallReceipt struct {
	CallID string
}

type QueryCallReceiptResponse struct {
	Receipt	ServiceReceiptCanonicalView
	Found	bool
}

type QueryCallsByCaller struct {
	Caller string
}

type QueryCallsByCallerResponse struct {
	Calls		[]UnifiedServiceCall
	Total		uint64
	ResponseHash	string
}

type QueryCallProof struct {
	ServiceID	string
	CallID		string
}

type ServiceCallIdempotencyRecord struct {
	ServiceID	string
	Caller		string
	IdempotencyKey	string
	PayloadHash	string
	CallID		string
	RecordHash	string
}

type ServiceCallExpiryRecord struct {
	ServiceID	string
	CallID		string
	DeadlineHeight	uint64
	AnchorHeight	uint64
	RecordHash	string
}

type ServiceCallsABCIProposalContract struct {
	ClassifyByTargetService		bool
	VerifySameSenderOrdering	bool
	RejectExpiredCalls		bool
	IncludeCallbacksAndRetries	bool
	AnchorReceiptsInFinalizeBlock	bool
	ContractHash			string
}

func DefaultXServiceCallsModuleBreakdown() (XServiceCallsModuleBreakdown, error) {
	breakdown := XServiceCallsModuleBreakdown{
		ModulePath:	ServiceModuleCalls,
		Purpose: []string{
			"call_envelopes",
			"callbacks",
			"execution_receipts",
			"idempotency",
			"retries",
		},
		StateObjects: []XServiceCallsStateObject{
			XServiceCallsStateServiceCall,
			XServiceCallsStateCallNonce,
			XServiceCallsStateIdempotencyRecord,
			XServiceCallsStateCallbackRecord,
			XServiceCallsStateCallReceipt,
		},
		Messages: []XServiceCallsMessageName{
			XServiceCallsMsgSubmitServiceCall,
			XServiceCallsMsgAnchorServiceResult,
			XServiceCallsMsgRetryServiceCall,
			XServiceCallsMsgSubmitCallback,
			XServiceCallsMsgExpireServiceCall,
		},
		Queries: []XServiceCallsQueryName{
			XServiceCallsQueryServiceCall,
			XServiceCallsQueryCallReceipt,
			XServiceCallsQueryCallsByCaller,
			XServiceCallsQueryCallProof,
		},
		FailureModes: []XServiceCallsFailureCoverage{
			newXServiceCallsFailureCoverage(XServiceCallsFailureNonceReplay, "ValidateServiceCallAnte", ServiceStoreV2CallPrefix),
			newXServiceCallsFailureCoverage(XServiceCallsFailureDuplicateIdempotencyKey, "NewServiceCallIdempotencyRecord", ServiceStoreV2CallPrefix),
			newXServiceCallsFailureCoverage(XServiceCallsFailureCallbackMismatch, "UnifiedServiceCallback.ValidateForTarget", ServiceStoreV2CallPrefix),
			newXServiceCallsFailureCoverage(XServiceCallsFailureExpiredCallAnchoredLate, "ValidateServiceCallResultAnchorWindow", ServiceStoreV2ReceiptPrefix),
			newXServiceCallsFailureCoverage(XServiceCallsFailureResultHashMismatch, "ValidateServiceCallResultHash", ServiceStoreV2ReceiptPrefix),
		},
		IntegrationPoints: []XServiceCallsIntegrationPoint{
			XServiceCallsIntegrationServices,
			XServiceCallsIntegrationServicePayments,
			XServiceCallsIntegrationServiceReceipts,
			XServiceCallsIntegrationABCIProposalHandling,
		},
	}
	return NewXServiceCallsModuleBreakdown(breakdown)
}

func NewXServiceCallsModuleBreakdown(breakdown XServiceCallsModuleBreakdown) (XServiceCallsModuleBreakdown, error) {
	breakdown = canonicalXServiceCallsModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return XServiceCallsModuleBreakdown{}, err
	}
	breakdown.BreakdownHash = ComputeXServiceCallsModuleBreakdownHash(breakdown)
	return breakdown, breakdown.Validate()
}

func NewMsgSubmitServiceCall(authority string, call UnifiedServiceCall) (MsgSubmitServiceCall, error) {
	msg := MsgSubmitServiceCall{Authority: strings.TrimSpace(authority), Call: call}
	msg.MessageHash = ComputeMsgSubmitServiceCallHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgAnchorServiceResult(authority string, call UnifiedServiceCall, expectedResultHash string, outcome ServiceExecutionOutcome) (MsgAnchorServiceResult, error) {
	msg := MsgAnchorServiceResult{
		Authority:		strings.TrimSpace(authority),
		CallID:			strings.ToLower(strings.TrimSpace(call.CallID)),
		ExpectedResultHash:	strings.ToLower(strings.TrimSpace(expectedResultHash)),
		Outcome:		outcome,
	}
	msg.MessageHash = ComputeMsgAnchorServiceResultHash(msg)
	return msg, msg.ValidateForCall(call)
}

func NewMsgRetryServiceCall(authority string, original UnifiedServiceCall, retry UnifiedServiceCall) (MsgRetryServiceCall, error) {
	msg := MsgRetryServiceCall{Authority: strings.TrimSpace(authority), OriginalCallID: strings.ToLower(strings.TrimSpace(original.CallID)), RetryCall: retry}
	msg.MessageHash = ComputeMsgRetryServiceCallHash(msg)
	return msg, msg.ValidateForOriginal(original)
}

func NewMsgSubmitCallback(authority string, callback UnifiedServiceCallback) (MsgSubmitCallback, error) {
	msg := MsgSubmitCallback{Authority: strings.TrimSpace(authority), Callback: callback}
	msg.MessageHash = ComputeMsgSubmitCallbackHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgExpireServiceCall(authority string, call UnifiedServiceCall, expireHeight uint64) (MsgExpireServiceCall, error) {
	msg := MsgExpireServiceCall{
		Authority:	strings.TrimSpace(authority),
		CallID:		strings.ToLower(strings.TrimSpace(call.CallID)),
		ServiceID:	strings.TrimSpace(call.TargetService),
		ExpireHeight:	expireHeight,
	}
	msg.MessageHash = ComputeMsgExpireServiceCallHash(msg)
	return msg, msg.ValidateForCall(call)
}

func NewServiceCallIdempotencyRecord(call UnifiedServiceCall, existing []ServiceCallIdempotencyRecord) (ServiceCallIdempotencyRecord, error) {
	if err := validateServiceCallIdempotencyRecords(existing); err != nil {
		return ServiceCallIdempotencyRecord{}, err
	}
	if call.IdempotencyKey == "" {
		return ServiceCallIdempotencyRecord{}, errors.New("x/servicecalls idempotency key is required")
	}
	for _, record := range existing {
		if record.ServiceID == call.TargetService && record.Caller == call.Caller && record.IdempotencyKey == call.IdempotencyKey {
			if record.PayloadHash != call.PayloadHash || record.CallID != call.CallID {
				return ServiceCallIdempotencyRecord{}, errors.New("x/servicecalls duplicate idempotency key misuse")
			}
			return record, nil
		}
	}
	record := ServiceCallIdempotencyRecord{
		ServiceID:	call.TargetService,
		Caller:		call.Caller,
		IdempotencyKey:	call.IdempotencyKey,
		PayloadHash:	call.PayloadHash,
		CallID:		call.CallID,
	}
	record.RecordHash = ComputeServiceCallIdempotencyRecordHash(record)
	return record, record.Validate()
}

func ValidateServiceCallResultAnchorWindow(call UnifiedServiceCall, anchorHeight uint64) (ServiceCallExpiryRecord, error) {
	record := ServiceCallExpiryRecord{
		ServiceID:	call.TargetService,
		CallID:		call.CallID,
		DeadlineHeight:	call.DeadlineHeight,
		AnchorHeight:	anchorHeight,
	}
	record.RecordHash = ComputeServiceCallExpiryRecordHash(record)
	if err := record.Validate(); err != nil {
		return record, err
	}
	if anchorHeight > call.DeadlineHeight {
		return record, fmt.Errorf("x/servicecalls expired call %s anchored late", call.CallID)
	}
	return record, nil
}

func ValidateServiceCallResultHash(call UnifiedServiceCall, expectedResultHash string, outcome ServiceExecutionOutcome) error {
	if err := coretypes.ValidateHash("x/servicecalls expected result hash", expectedResultHash); err != nil {
		return err
	}
	outcome.ResponseHash = strings.ToLower(strings.TrimSpace(outcome.ResponseHash))
	if outcome.CallID != call.CallID {
		return errors.New("x/servicecalls result call mismatch")
	}
	if outcome.ResponseHash != strings.ToLower(strings.TrimSpace(expectedResultHash)) {
		return errors.New("x/servicecalls result hash mismatch")
	}
	return nil
}

func QueryServiceCallFromCalls(calls []UnifiedServiceCall, query QueryServiceCall) (QueryServiceCallResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryServiceCallResponse{}, err
	}
	for _, call := range calls {
		if call.CallID == query.CallID {
			return QueryServiceCallResponse{Call: call, Found: true}, nil
		}
	}
	return QueryServiceCallResponse{Found: false}, nil
}

func QueryCallReceiptFromReceipts(receipts []ServiceReceipt, query QueryCallReceipt) (QueryCallReceiptResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryCallReceiptResponse{}, err
	}
	for _, receipt := range receipts {
		if receipt.CallID == query.CallID {
			view, err := NewServiceReceiptCanonicalView(receipt)
			if err != nil {
				return QueryCallReceiptResponse{}, err
			}
			return QueryCallReceiptResponse{Receipt: view, Found: true}, nil
		}
	}
	return QueryCallReceiptResponse{Found: false}, nil
}

func QueryCallsByCallerFromCalls(calls []UnifiedServiceCall, query QueryCallsByCaller) (QueryCallsByCallerResponse, error) {
	if err := query.Validate(); err != nil {
		return QueryCallsByCallerResponse{}, err
	}
	filtered := []UnifiedServiceCall{}
	for _, call := range calls {
		if call.Caller == query.Caller {
			filtered = append(filtered, call)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CallID < filtered[j].CallID })
	response := QueryCallsByCallerResponse{Calls: filtered, Total: uint64(len(filtered))}
	response.ResponseHash = ComputeQueryCallsByCallerResponseHash(response)
	return response, response.Validate()
}

func DefaultServiceCallsABCIProposalContract() (ServiceCallsABCIProposalContract, error) {
	contract := ServiceCallsABCIProposalContract{
		ClassifyByTargetService:	true,
		VerifySameSenderOrdering:	true,
		RejectExpiredCalls:		true,
		IncludeCallbacksAndRetries:	true,
		AnchorReceiptsInFinalizeBlock:	true,
	}
	contract.ContractHash = ComputeServiceCallsABCIProposalContractHash(contract)
	return contract, contract.Validate()
}

func (breakdown XServiceCallsModuleBreakdown) ValidateFormat() error {
	if breakdown.ModulePath != ServiceModuleCalls {
		return errors.New("x/servicecalls breakdown must describe x/servicecalls")
	}
	if err := validateSortedTokens("x/servicecalls purpose", breakdown.Purpose); err != nil {
		return err
	}
	if err := validateXServiceCallsStateObjects(breakdown.StateObjects); err != nil {
		return err
	}
	if err := validateXServiceCallsMessages(breakdown.Messages); err != nil {
		return err
	}
	if err := validateXServiceCallsQueries(breakdown.Queries); err != nil {
		return err
	}
	if err := validateXServiceCallsFailureCoverage(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateXServiceCallsIntegrationPoints(breakdown.IntegrationPoints); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return coretypes.ValidateHash("x/servicecalls breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown XServiceCallsModuleBreakdown) Validate() error {
	breakdown = canonicalXServiceCallsModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("x/servicecalls breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeXServiceCallsModuleBreakdownHash(breakdown) {
		return errors.New("x/servicecalls breakdown hash mismatch")
	}
	return nil
}

func (coverage XServiceCallsFailureCoverage) Validate() error {
	if !IsXServiceCallsFailureMode(coverage.Mode) {
		return fmt.Errorf("x/servicecalls unknown failure mode %q", coverage.Mode)
	}
	if err := validateInterfaceToken("x/servicecalls failure guard", coverage.Guard); err != nil {
		return err
	}
	if !IsServiceStoreKey(coverage.Scope + "/_") {
		return fmt.Errorf("x/servicecalls failure scope %s must be services store key", coverage.Scope)
	}
	return nil
}

func (msg MsgSubmitServiceCall) ValidateBasic() error {
	if err := validateInterfaceToken("x/servicecalls submit authority", msg.Authority); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls submit call hash", msg.Call.UnifiedCallHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls submit message hash", msg.MessageHash); err != nil {
		return err
	}
	if msg.MessageHash != ComputeMsgSubmitServiceCallHash(msg) {
		return errors.New("x/servicecalls submit message hash mismatch")
	}
	return nil
}

func (msg MsgAnchorServiceResult) ValidateForCall(call UnifiedServiceCall) error {
	if err := validateInterfaceToken("x/servicecalls result authority", msg.Authority); err != nil {
		return err
	}
	if msg.CallID != call.CallID {
		return errors.New("x/servicecalls result call mismatch")
	}
	if err := ValidateServiceCallResultHash(call, msg.ExpectedResultHash, msg.Outcome); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls result message hash", msg.MessageHash); err != nil {
		return err
	}
	if msg.MessageHash != ComputeMsgAnchorServiceResultHash(msg) {
		return errors.New("x/servicecalls result message hash mismatch")
	}
	return nil
}

func (msg MsgRetryServiceCall) ValidateForOriginal(original UnifiedServiceCall) error {
	if err := validateInterfaceToken("x/servicecalls retry authority", msg.Authority); err != nil {
		return err
	}
	if msg.OriginalCallID != original.CallID || msg.RetryCall.RetryOf != original.CallID {
		return errors.New("x/servicecalls retry original call mismatch")
	}
	if msg.RetryCall.Kind != coretypes.ServiceCallKindRetry {
		return errors.New("x/servicecalls retry message requires retry call kind")
	}
	if err := coretypes.ValidateHash("x/servicecalls retry message hash", msg.MessageHash); err != nil {
		return err
	}
	if msg.MessageHash != ComputeMsgRetryServiceCallHash(msg) {
		return errors.New("x/servicecalls retry message hash mismatch")
	}
	return nil
}

func (msg MsgSubmitCallback) ValidateBasic() error {
	if err := validateInterfaceToken("x/servicecalls callback authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Callback.ValidateBasic(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls callback message hash", msg.MessageHash); err != nil {
		return err
	}
	if msg.MessageHash != ComputeMsgSubmitCallbackHash(msg) {
		return errors.New("x/servicecalls callback message hash mismatch")
	}
	return nil
}

func (msg MsgExpireServiceCall) ValidateForCall(call UnifiedServiceCall) error {
	if err := validateInterfaceToken("x/servicecalls expire authority", msg.Authority); err != nil {
		return err
	}
	if msg.CallID != call.CallID || msg.ServiceID != call.TargetService {
		return errors.New("x/servicecalls expire call mismatch")
	}
	if msg.ExpireHeight <= call.DeadlineHeight {
		return errors.New("x/servicecalls expire height must be after call deadline")
	}
	if err := coretypes.ValidateHash("x/servicecalls expire message hash", msg.MessageHash); err != nil {
		return err
	}
	if msg.MessageHash != ComputeMsgExpireServiceCallHash(msg) {
		return errors.New("x/servicecalls expire message hash mismatch")
	}
	return nil
}

func (query QueryServiceCall) Validate() error {
	return coretypes.ValidateHash("x/servicecalls query call id", query.CallID)
}

func (query QueryCallReceipt) Validate() error {
	return coretypes.ValidateHash("x/servicecalls query receipt call id", query.CallID)
}

func (query QueryCallsByCaller) Validate() error {
	return validateInterfaceToken("x/servicecalls query caller", query.Caller)
}

func (query QueryCallProof) ToServiceCallProofQuery() QueryServiceCallProof {
	return QueryServiceCallProof{ServiceID: query.ServiceID, CallID: query.CallID}
}

func (query QueryCallProof) Validate() error {
	return query.ToServiceCallProofQuery().Validate()
}

func (response QueryCallsByCallerResponse) Validate() error {
	if response.Total != uint64(len(response.Calls)) {
		return errors.New("x/servicecalls caller query total mismatch")
	}
	previous := ""
	for _, call := range response.Calls {
		if previous != "" && previous >= call.CallID {
			return errors.New("x/servicecalls caller query calls must be sorted")
		}
		previous = call.CallID
	}
	if err := coretypes.ValidateHash("x/servicecalls caller query response hash", response.ResponseHash); err != nil {
		return err
	}
	if response.ResponseHash != ComputeQueryCallsByCallerResponseHash(response) {
		return errors.New("x/servicecalls caller query response hash mismatch")
	}
	return nil
}

func (record ServiceCallIdempotencyRecord) Validate() error {
	if err := validateInterfaceToken("x/servicecalls idempotency service id", record.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/servicecalls idempotency caller", record.Caller); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/servicecalls idempotency key", record.IdempotencyKey); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls idempotency payload hash", record.PayloadHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls idempotency call id", record.CallID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls idempotency record hash", record.RecordHash); err != nil {
		return err
	}
	if record.RecordHash != ComputeServiceCallIdempotencyRecordHash(record) {
		return errors.New("x/servicecalls idempotency record hash mismatch")
	}
	return nil
}

func (record ServiceCallExpiryRecord) Validate() error {
	if err := validateInterfaceToken("x/servicecalls expiry service id", record.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/servicecalls expiry call id", record.CallID); err != nil {
		return err
	}
	if record.DeadlineHeight == 0 || record.AnchorHeight == 0 {
		return errors.New("x/servicecalls expiry heights must be positive")
	}
	if err := coretypes.ValidateHash("x/servicecalls expiry record hash", record.RecordHash); err != nil {
		return err
	}
	if record.RecordHash != ComputeServiceCallExpiryRecordHash(record) {
		return errors.New("x/servicecalls expiry record hash mismatch")
	}
	return nil
}

func (contract ServiceCallsABCIProposalContract) Validate() error {
	if !contract.ClassifyByTargetService || !contract.VerifySameSenderOrdering || !contract.RejectExpiredCalls ||
		!contract.IncludeCallbacksAndRetries || !contract.AnchorReceiptsInFinalizeBlock {
		return errors.New("x/servicecalls ABCI++ proposal contract must enable all service call handling rules")
	}
	if err := coretypes.ValidateHash("x/servicecalls ABCI++ contract hash", contract.ContractHash); err != nil {
		return err
	}
	if contract.ContractHash != ComputeServiceCallsABCIProposalContractHash(contract) {
		return errors.New("x/servicecalls ABCI++ contract hash mismatch")
	}
	return nil
}

func IsXServiceCallsStateObject(object XServiceCallsStateObject) bool {
	switch object {
	case XServiceCallsStateServiceCall, XServiceCallsStateCallNonce, XServiceCallsStateIdempotencyRecord, XServiceCallsStateCallbackRecord, XServiceCallsStateCallReceipt:
		return true
	default:
		return false
	}
}

func IsXServiceCallsMessageName(message XServiceCallsMessageName) bool {
	switch message {
	case XServiceCallsMsgSubmitServiceCall, XServiceCallsMsgAnchorServiceResult, XServiceCallsMsgRetryServiceCall, XServiceCallsMsgSubmitCallback, XServiceCallsMsgExpireServiceCall:
		return true
	default:
		return false
	}
}

func IsXServiceCallsQueryName(query XServiceCallsQueryName) bool {
	switch query {
	case XServiceCallsQueryServiceCall, XServiceCallsQueryCallReceipt, XServiceCallsQueryCallsByCaller, XServiceCallsQueryCallProof:
		return true
	default:
		return false
	}
}

func IsXServiceCallsFailureMode(mode XServiceCallsFailureMode) bool {
	switch mode {
	case XServiceCallsFailureNonceReplay, XServiceCallsFailureDuplicateIdempotencyKey, XServiceCallsFailureCallbackMismatch, XServiceCallsFailureExpiredCallAnchoredLate, XServiceCallsFailureResultHashMismatch:
		return true
	default:
		return false
	}
}

func IsXServiceCallsIntegrationPoint(point XServiceCallsIntegrationPoint) bool {
	switch point {
	case XServiceCallsIntegrationServices, XServiceCallsIntegrationServicePayments, XServiceCallsIntegrationServiceReceipts, XServiceCallsIntegrationABCIProposalHandling:
		return true
	default:
		return false
	}
}

func ComputeXServiceCallsModuleBreakdownHash(breakdown XServiceCallsModuleBreakdown) string {
	breakdown = canonicalXServiceCallsModuleBreakdown(breakdown)
	parts := []string{"aetra-x-servicecalls-module-breakdown-v1", breakdown.ModulePath, "purpose", fmt.Sprint(len(breakdown.Purpose))}
	parts = append(parts, breakdown.Purpose...)
	parts = append(parts, "state", fmt.Sprint(len(breakdown.StateObjects)))
	for _, object := range breakdown.StateObjects {
		parts = append(parts, string(object))
	}
	parts = append(parts, "messages", fmt.Sprint(len(breakdown.Messages)))
	for _, message := range breakdown.Messages {
		parts = append(parts, string(message))
	}
	parts = append(parts, "queries", fmt.Sprint(len(breakdown.Queries)))
	for _, query := range breakdown.Queries {
		parts = append(parts, string(query))
	}
	parts = append(parts, "failures", fmt.Sprint(len(breakdown.FailureModes)))
	for _, coverage := range breakdown.FailureModes {
		parts = append(parts, string(coverage.Mode), coverage.Guard, coverage.Scope)
	}
	parts = append(parts, "integrations", fmt.Sprint(len(breakdown.IntegrationPoints)))
	for _, point := range breakdown.IntegrationPoints {
		parts = append(parts, string(point))
	}
	return servicesHashParts(parts...)
}

func ComputeMsgSubmitServiceCallHash(msg MsgSubmitServiceCall) string {
	return servicesHashParts("aetra-x-servicecalls-msg-submit-v1", msg.Authority, msg.Call.CallID, msg.Call.UnifiedCallHash)
}

func ComputeMsgAnchorServiceResultHash(msg MsgAnchorServiceResult) string {
	return servicesHashParts("aetra-x-servicecalls-msg-anchor-result-v1", msg.Authority, msg.CallID, msg.ExpectedResultHash, msg.Outcome.ResponseHash, msg.Outcome.ProofHash, string(msg.Outcome.Status), fmt.Sprint(msg.Outcome.ExecutedHeight))
}

func ComputeMsgRetryServiceCallHash(msg MsgRetryServiceCall) string {
	return servicesHashParts("aetra-x-servicecalls-msg-retry-v1", msg.Authority, msg.OriginalCallID, msg.RetryCall.CallID, msg.RetryCall.UnifiedCallHash)
}

func ComputeMsgSubmitCallbackHash(msg MsgSubmitCallback) string {
	return servicesHashParts("aetra-x-servicecalls-msg-callback-v1", msg.Authority, msg.Callback.CallbackCallID, msg.Callback.CallbackHash)
}

func ComputeMsgExpireServiceCallHash(msg MsgExpireServiceCall) string {
	return servicesHashParts("aetra-x-servicecalls-msg-expire-v1", msg.Authority, msg.ServiceID, msg.CallID, fmt.Sprint(msg.ExpireHeight))
}

func ComputeServiceCallIdempotencyRecordHash(record ServiceCallIdempotencyRecord) string {
	return servicesHashParts("aetra-x-servicecalls-idempotency-v1", record.ServiceID, record.Caller, record.IdempotencyKey, record.PayloadHash, record.CallID)
}

func ComputeServiceCallExpiryRecordHash(record ServiceCallExpiryRecord) string {
	return servicesHashParts("aetra-x-servicecalls-expiry-v1", record.ServiceID, record.CallID, fmt.Sprint(record.DeadlineHeight), fmt.Sprint(record.AnchorHeight))
}

func ComputeQueryCallsByCallerResponseHash(response QueryCallsByCallerResponse) string {
	calls := append([]UnifiedServiceCall(nil), response.Calls...)
	sort.SliceStable(calls, func(i, j int) bool { return calls[i].CallID < calls[j].CallID })
	parts := []string{"aetra-x-servicecalls-by-caller-v1", fmt.Sprint(response.Total)}
	for _, call := range calls {
		parts = append(parts, call.CallID, call.UnifiedCallHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceCallsABCIProposalContractHash(contract ServiceCallsABCIProposalContract) string {
	return servicesHashParts(
		"aetra-x-servicecalls-abci-contract-v1",
		fmt.Sprint(contract.ClassifyByTargetService),
		fmt.Sprint(contract.VerifySameSenderOrdering),
		fmt.Sprint(contract.RejectExpiredCalls),
		fmt.Sprint(contract.IncludeCallbacksAndRetries),
		fmt.Sprint(contract.AnchorReceiptsInFinalizeBlock),
	)
}

func newXServiceCallsFailureCoverage(mode XServiceCallsFailureMode, guard, scope string) XServiceCallsFailureCoverage {
	return XServiceCallsFailureCoverage{Mode: mode, Guard: guard, Scope: scope}
}

func canonicalXServiceCallsModuleBreakdown(breakdown XServiceCallsModuleBreakdown) XServiceCallsModuleBreakdown {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedStrings(breakdown.Purpose)
	breakdown.StateObjects = sortedXServiceCallsStateObjects(breakdown.StateObjects)
	breakdown.Messages = sortedXServiceCallsMessages(breakdown.Messages)
	breakdown.Queries = sortedXServiceCallsQueries(breakdown.Queries)
	breakdown.FailureModes = sortedXServiceCallsFailureCoverage(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedXServiceCallsIntegrationPoints(breakdown.IntegrationPoints)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func validateServiceCallIdempotencyRecords(records []ServiceCallIdempotencyRecord) error {
	seen := map[string]struct{}{}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		key := record.ServiceID + "/" + record.Caller + "/" + record.IdempotencyKey
		if _, found := seen[key]; found {
			return fmt.Errorf("x/servicecalls duplicate idempotency record %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateXServiceCallsStateObjects(objects []XServiceCallsStateObject) error {
	required := []XServiceCallsStateObject{XServiceCallsStateServiceCall, XServiceCallsStateCallNonce, XServiceCallsStateIdempotencyRecord, XServiceCallsStateCallbackRecord, XServiceCallsStateCallReceipt}
	return validateXServiceCallsEnumSet("state object", objects, required, IsXServiceCallsStateObject)
}

func validateXServiceCallsMessages(messages []XServiceCallsMessageName) error {
	required := []XServiceCallsMessageName{XServiceCallsMsgSubmitServiceCall, XServiceCallsMsgAnchorServiceResult, XServiceCallsMsgRetryServiceCall, XServiceCallsMsgSubmitCallback, XServiceCallsMsgExpireServiceCall}
	return validateXServiceCallsEnumSet("message", messages, required, IsXServiceCallsMessageName)
}

func validateXServiceCallsQueries(queries []XServiceCallsQueryName) error {
	required := []XServiceCallsQueryName{XServiceCallsQueryServiceCall, XServiceCallsQueryCallReceipt, XServiceCallsQueryCallsByCaller, XServiceCallsQueryCallProof}
	return validateXServiceCallsEnumSet("query", queries, required, IsXServiceCallsQueryName)
}

func validateXServiceCallsFailureCoverage(coverage []XServiceCallsFailureCoverage) error {
	required := []XServiceCallsFailureMode{XServiceCallsFailureNonceReplay, XServiceCallsFailureDuplicateIdempotencyKey, XServiceCallsFailureCallbackMismatch, XServiceCallsFailureExpiredCallAnchoredLate, XServiceCallsFailureResultHashMismatch}
	if len(coverage) != len(required) {
		return fmt.Errorf("x/servicecalls expected %d failure modes", len(required))
	}
	seen := map[XServiceCallsFailureMode]struct{}{}
	for _, item := range coverage {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, found := seen[item.Mode]; found {
			return fmt.Errorf("x/servicecalls duplicate failure mode %s", item.Mode)
		}
		seen[item.Mode] = struct{}{}
	}
	for _, mode := range required {
		if _, found := seen[mode]; !found {
			return fmt.Errorf("x/servicecalls missing failure mode %s", mode)
		}
	}
	return nil
}

func validateXServiceCallsIntegrationPoints(points []XServiceCallsIntegrationPoint) error {
	required := []XServiceCallsIntegrationPoint{XServiceCallsIntegrationServices, XServiceCallsIntegrationServicePayments, XServiceCallsIntegrationServiceReceipts, XServiceCallsIntegrationABCIProposalHandling}
	return validateXServiceCallsEnumSet("integration", points, required, IsXServiceCallsIntegrationPoint)
}

func validateXServiceCallsEnumSet[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("x/servicecalls expected %d %s entries", len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("x/servicecalls unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("x/servicecalls %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("x/servicecalls duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/servicecalls missing %s %s", label, value)
		}
	}
	return nil
}

func sortedXServiceCallsStateObjects(values []XServiceCallsStateObject) []XServiceCallsStateObject {
	out := append([]XServiceCallsStateObject(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServiceCallsMessages(values []XServiceCallsMessageName) []XServiceCallsMessageName {
	out := append([]XServiceCallsMessageName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServiceCallsQueries(values []XServiceCallsQueryName) []XServiceCallsQueryName {
	out := append([]XServiceCallsQueryName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServiceCallsFailureCoverage(values []XServiceCallsFailureCoverage) []XServiceCallsFailureCoverage {
	out := append([]XServiceCallsFailureCoverage(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].Scope = strings.Trim(strings.TrimSpace(out[i].Scope), "/")
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func sortedXServiceCallsIntegrationPoints(values []XServiceCallsIntegrationPoint) []XServiceCallsIntegrationPoint {
	out := append([]XServiceCallsIntegrationPoint(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
