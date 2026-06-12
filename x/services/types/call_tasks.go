package types

import (
	"errors"
	"fmt"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceCallAnteValidation struct {
	CallID			string
	ServiceID		string
	Caller			string
	Nonce			uint64
	Accept			bool
	ReservePayment		bool
	RequiresProof		bool
	RequiresIdempotency	bool
	Retry			bool
	OriginalCallID		string
	ReplayIndexHash		string
	RoutingHash		string
	AnteValidationHash	string
}

type ServiceReceiptAnchorResult struct {
	Msg			MsgAnchorServiceReceipt
	Receipt			ServiceReceipt
	Tombstone		ServiceReceiptTombstone
	ReplayIndex		ServiceCallReplayIndex
	AnchorHash		string
	AnchorResultHash	string
}

type QueryServiceCallProof struct {
	ServiceID	string
	CallID		string
}

type ServiceCallProof struct {
	ServiceID	string
	CallID		string
	ReceiptHash	string
	TombstoneHash	string
	ReplayIndexHash	string
	ProofHeight	uint64
	ProofHash	string
}

type QueryServiceCallProofResponse struct {
	Proof	ServiceCallProof
	Found	bool
}

type SDKServiceCallBuildRequest struct {
	Context		coretypes.ServiceConsensusContext
	Descriptor	ServiceDescriptor
	MethodID	string
	Caller		string
	Nonce		uint64
	PayloadHash	string
	MaxFeeAmount	string
	SignatureHash	string
	TimeoutDelta	uint64
	IdempotencyKey	string
	CallbackTarget	string
	Kind		coretypes.ServiceCallKind
	RetryOf		string
}

type SDKServiceCallBuildResult struct {
	Call		UnifiedServiceCall
	Envelope	ServiceCallEnvelope
	Routing		UnifiedCallRoutingPlan
	Schema		WalletCLISchema
	ResultHash	string
}

func ValidateServiceCallAnte(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, index ServiceCallReplayIndex, call UnifiedServiceCall) (ServiceCallAnteValidation, error) {
	if err := index.Validate(); err != nil {
		return ServiceCallAnteValidation{}, err
	}
	if err := ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, call); err != nil {
		return ServiceCallAnteValidation{}, err
	}
	if existing, found := index.EntryByServiceCallerNonce(call.TargetService, call.Caller, call.Nonce); found {
		return ServiceCallAnteValidation{}, fmt.Errorf("services ante nonce already used by call %s", existing.CallID)
	}
	if call.Kind == coretypes.ServiceCallKindRetry && !index.ContainsCallID(call.RetryOf) {
		return ServiceCallAnteValidation{}, errors.New("services ante retry references unknown original call")
	}
	route, err := RouteUnifiedServiceCall(ctx, descriptor, call)
	if err != nil {
		return ServiceCallAnteValidation{}, err
	}
	_, err = AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	if err != nil {
		return ServiceCallAnteValidation{}, err
	}
	ante := ServiceCallAnteValidation{
		CallID:			call.CallID,
		ServiceID:		call.TargetService,
		Caller:			call.Caller,
		Nonce:			call.Nonce,
		Accept:			true,
		ReservePayment:		route.ReserveFundsBeforeExecution,
		RequiresProof:		route.VerifyResultProofBeforeAccept,
		RequiresIdempotency:	call.IdempotencyKey != "",
		Retry:			call.Kind == coretypes.ServiceCallKindRetry,
		OriginalCallID:		call.RetryOf,
		ReplayIndexHash:	index.IndexHash,
		RoutingHash:		route.RoutingHash,
	}
	ante.AnteValidationHash = ComputeServiceCallAnteValidationHash(ante)
	return ante, ante.Validate()
}

func (ante ServiceCallAnteValidation) Validate() error {
	if err := coretypes.ValidateHash("services ante call id", ante.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services ante service id", ante.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services ante caller", ante.Caller); err != nil {
		return err
	}
	if ante.Nonce == 0 {
		return errors.New("services ante nonce must be positive")
	}
	if !ante.Accept {
		return errors.New("services ante validation result must accept")
	}
	if ante.Retry {
		if err := coretypes.ValidateHash("services ante original call id", ante.OriginalCallID); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("services ante replay index hash", ante.ReplayIndexHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services ante routing hash", ante.RoutingHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services ante validation hash", ante.AnteValidationHash); err != nil {
		return err
	}
	if expected := ComputeServiceCallAnteValidationHash(ante); ante.AnteValidationHash != expected {
		return fmt.Errorf("services ante validation hash mismatch: expected %s", expected)
	}
	return nil
}

func AnchorUnifiedServiceReceipt(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, index ServiceCallReplayIndex, call UnifiedServiceCall, outcome ServiceExecutionOutcome, authority string) (ServiceReceiptAnchorResult, error) {
	if _, err := ValidateServiceCallAnte(ctx, descriptor, index, call); err != nil {
		return ServiceReceiptAnchorResult{}, err
	}
	outcome = coretypes.NormalizeServiceExecutionOutcome(ctx, outcome)
	if outcome.CallID != call.CallID {
		return ServiceReceiptAnchorResult{}, errors.New("services receipt anchor outcome call mismatch")
	}
	if err := coretypes.ValidateServiceExecutionOutcome(ctx, call.ToServiceCallEnvelope(), outcome); err != nil {
		return ServiceReceiptAnchorResult{}, err
	}
	receipt, err := coretypes.NewServiceCallReceipt(call.ToServiceCallEnvelope(), outcome)
	if err != nil {
		return ServiceReceiptAnchorResult{}, err
	}
	nextIndex, err := AcceptUnifiedServiceCall(ctx, descriptor, index, call)
	if err != nil {
		return ServiceReceiptAnchorResult{}, err
	}
	nextIndex, tombstone, err := TombstoneServiceReceipt(ctx, nextIndex, call, receipt)
	if err != nil {
		return ServiceReceiptAnchorResult{}, err
	}
	anchorHash := ComputeServiceReceiptAnchorHash(call, receipt, tombstone)
	msg, err := coretypes.NewMsgAnchorServiceReceipt(authority, receipt, anchorHash)
	if err != nil {
		return ServiceReceiptAnchorResult{}, err
	}
	result := ServiceReceiptAnchorResult{
		Msg:		msg,
		Receipt:	receipt,
		Tombstone:	tombstone,
		ReplayIndex:	nextIndex,
		AnchorHash:	anchorHash,
	}
	result.AnchorResultHash = ComputeServiceReceiptAnchorResultHash(result)
	return result, result.Validate()
}

func (result ServiceReceiptAnchorResult) Validate() error {
	if err := result.Msg.ValidateBasic(); err != nil {
		return err
	}
	if err := result.Receipt.Validate(); err != nil {
		return err
	}
	if err := result.Tombstone.Validate(); err != nil {
		return err
	}
	if err := result.ReplayIndex.Validate(); err != nil {
		return err
	}
	if result.Receipt.CallID != result.Tombstone.CallID {
		return errors.New("services receipt anchor tombstone mismatch")
	}
	if err := coretypes.ValidateHash("services receipt anchor hash", result.AnchorHash); err != nil {
		return err
	}
	if expected := ComputeServiceReceiptAnchorHashForParts(result.Receipt.CallID, result.Receipt.ReceiptHash, result.Tombstone.TombstoneHash); result.AnchorHash != expected {
		return fmt.Errorf("services receipt anchor hash mismatch: expected %s", expected)
	}
	if err := coretypes.ValidateHash("services receipt anchor result hash", result.AnchorResultHash); err != nil {
		return err
	}
	if expected := ComputeServiceReceiptAnchorResultHash(result); result.AnchorResultHash != expected {
		return fmt.Errorf("services receipt anchor result hash mismatch: expected %s", expected)
	}
	return nil
}

func (q QueryServiceCallProof) Validate() error {
	if err := validateInterfaceToken("services call proof service id", q.ServiceID); err != nil {
		return err
	}
	return coretypes.ValidateHash("services call proof call id", q.CallID)
}

func QueryServiceCallProofFromReplayIndex(index ServiceCallReplayIndex, receipts []ServiceReceipt, q QueryServiceCallProof, proofHeight uint64) (QueryServiceCallProofResponse, error) {
	if proofHeight == 0 {
		return QueryServiceCallProofResponse{}, errors.New("services call proof height must be positive")
	}
	if err := index.Validate(); err != nil {
		return QueryServiceCallProofResponse{}, err
	}
	if err := q.Validate(); err != nil {
		return QueryServiceCallProofResponse{}, err
	}
	tombstone, found := index.TombstoneByCallID(q.CallID)
	if !found || tombstone.ServiceID != q.ServiceID {
		return QueryServiceCallProofResponse{Found: false}, nil
	}
	receiptHash := tombstone.ReceiptHash
	for _, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return QueryServiceCallProofResponse{}, err
		}
		if receipt.CallID == q.CallID && receipt.ServiceID == q.ServiceID {
			receiptHash = receipt.ReceiptHash
			break
		}
	}
	proof := ServiceCallProof{
		ServiceID:		q.ServiceID,
		CallID:			q.CallID,
		ReceiptHash:		receiptHash,
		TombstoneHash:		tombstone.TombstoneHash,
		ReplayIndexHash:	index.IndexHash,
		ProofHeight:		proofHeight,
	}
	proof.ProofHash = ComputeServiceCallProofHash(proof)
	return QueryServiceCallProofResponse{Proof: proof, Found: true}, proof.Validate()
}

func (proof ServiceCallProof) Validate() error {
	if err := validateInterfaceToken("services call proof service id", proof.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services call proof call id", proof.CallID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services call proof receipt hash", proof.ReceiptHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services call proof tombstone hash", proof.TombstoneHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services call proof replay index hash", proof.ReplayIndexHash); err != nil {
		return err
	}
	if proof.ProofHeight == 0 {
		return errors.New("services call proof height must be positive")
	}
	if err := coretypes.ValidateHash("services call proof hash", proof.ProofHash); err != nil {
		return err
	}
	if expected := ComputeServiceCallProofHash(proof); proof.ProofHash != expected {
		return fmt.Errorf("services call proof hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildSDKServiceCall(req SDKServiceCallBuildRequest) (SDKServiceCallBuildResult, error) {
	call, err := NewUnifiedServiceCall(
		req.Context,
		req.Descriptor,
		req.MethodID,
		req.Caller,
		req.Nonce,
		req.PayloadHash,
		req.MaxFeeAmount,
		req.SignatureHash,
		req.TimeoutDelta,
		req.IdempotencyKey,
		req.CallbackTarget,
	)
	if err != nil {
		return SDKServiceCallBuildResult{}, err
	}
	if req.Kind != "" {
		call.Kind = req.Kind
	}
	if req.RetryOf != "" {
		call.RetryOf = strings.ToLower(strings.TrimSpace(req.RetryOf))
	}
	call.CallID = coretypes.NormalizeServiceCall(req.Context, call.ToServiceCallEnvelope()).CallID
	call.UnifiedCallHash = ComputeUnifiedServiceCallHash(call)
	if err := ValidateUnifiedServiceCallForDescriptor(req.Context, req.Descriptor, call); err != nil {
		return SDKServiceCallBuildResult{}, err
	}
	routing, err := RouteUnifiedServiceCall(req.Context, req.Descriptor, call)
	if err != nil {
		return SDKServiceCallBuildResult{}, err
	}
	schema, err := BuildWalletCLISchema(req.Descriptor, req.MethodID)
	if err != nil {
		return SDKServiceCallBuildResult{}, err
	}
	result := SDKServiceCallBuildResult{
		Call:		call,
		Envelope:	call.ToServiceCallEnvelope(),
		Routing:	routing,
		Schema:		schema,
	}
	result.ResultHash = ComputeSDKServiceCallBuildResultHash(result)
	return result, result.Validate()
}

func (result SDKServiceCallBuildResult) Validate() error {
	if err := coretypes.ValidateHash("services sdk call unified hash", result.Call.UnifiedCallHash); err != nil {
		return err
	}
	if result.Envelope.CallID != result.Call.CallID {
		return errors.New("services sdk call envelope mismatch")
	}
	if err := result.Routing.Validate(); err != nil {
		return err
	}
	if err := result.Schema.Validate(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services sdk call result hash", result.ResultHash); err != nil {
		return err
	}
	if expected := ComputeSDKServiceCallBuildResultHash(result); result.ResultHash != expected {
		return fmt.Errorf("services sdk call result hash mismatch: expected %s", expected)
	}
	return nil
}

func (result SDKServiceCallBuildResult) ValidateForContext(ctx coretypes.ServiceConsensusContext) error {
	if err := result.Call.ValidateBasic(ctx); err != nil {
		return err
	}
	return result.Validate()
}

func ComputeServiceCallAnteValidationHash(ante ServiceCallAnteValidation) string {
	return servicesHashParts(
		"aetra-services-call-ante-v1",
		ante.CallID,
		ante.ServiceID,
		ante.Caller,
		fmt.Sprint(ante.Nonce),
		fmt.Sprint(ante.Accept),
		fmt.Sprint(ante.ReservePayment),
		fmt.Sprint(ante.RequiresProof),
		fmt.Sprint(ante.RequiresIdempotency),
		fmt.Sprint(ante.Retry),
		ante.OriginalCallID,
		ante.ReplayIndexHash,
		ante.RoutingHash,
	)
}

func ComputeServiceReceiptAnchorHash(call UnifiedServiceCall, receipt ServiceReceipt, tombstone ServiceReceiptTombstone) string {
	return ComputeServiceReceiptAnchorHashForParts(call.CallID, receipt.ReceiptHash, tombstone.TombstoneHash)
}

func ComputeServiceReceiptAnchorHashForParts(callID, receiptHash, tombstoneHash string) string {
	return servicesHashParts("aetra-services-receipt-anchor-v1", callID, receiptHash, tombstoneHash)
}

func ComputeServiceReceiptAnchorResultHash(result ServiceReceiptAnchorResult) string {
	return servicesHashParts(
		"aetra-services-receipt-anchor-result-v1",
		result.Msg.MessageHash,
		result.Receipt.ReceiptHash,
		result.Tombstone.TombstoneHash,
		result.ReplayIndex.IndexHash,
		result.AnchorHash,
	)
}

func ComputeServiceCallProofHash(proof ServiceCallProof) string {
	return servicesHashParts(
		"aetra-services-call-proof-v1",
		proof.ServiceID,
		proof.CallID,
		proof.ReceiptHash,
		proof.TombstoneHash,
		proof.ReplayIndexHash,
		fmt.Sprint(proof.ProofHeight),
	)
}

func ComputeSDKServiceCallBuildResultHash(result SDKServiceCallBuildResult) string {
	return servicesHashParts(
		"aetra-services-sdk-call-build-v1",
		result.Call.UnifiedCallHash,
		coretypes.ComputeServiceCallEnvelopeHash(result.Envelope),
		result.Routing.RoutingHash,
		result.Schema.SchemaHash,
	)
}
