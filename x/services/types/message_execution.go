package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type UnifiedInteractionClass string

const (
	InteractionOnChainTransaction	UnifiedInteractionClass	= "on_chain_transaction"
	InteractionOffChainServiceCall	UnifiedInteractionClass	= "off_chain_service_call"
	InteractionHybridExecutionFlow	UnifiedInteractionClass	= "hybrid_execution_flow"
	InteractionAsyncCallback	UnifiedInteractionClass	= "async_callback"
	InteractionRetry		UnifiedInteractionClass	= "retry"
	InteractionEventedSubscription	UnifiedInteractionClass	= "evented_subscription"
)

type UnifiedInteractionPlan struct {
	CallID			string
	ServiceID		string
	MethodID		string
	InteractionClass	UnifiedInteractionClass
	Kind			coretypes.ServiceCallKind
	ExecutionType		coretypes.ServiceMethodExecutionType
	Routes			[]UnifiedServiceRoute
	ReplaySafe		bool
	PlanHash		string
}

type UnifiedServiceCallback struct {
	OriginalCallID		string
	CallbackTarget		string
	CallbackMethod		string
	CallbackPayloadHash	string
	CallbackDeadline	uint64
	CallbackPaymentPolicy	string
	Caller			string
	Nonce			uint64
	IdempotencyKey		string
	CallbackCallID		string
	CallbackHash		string
}

type ServiceCallbackReceiptEmission struct {
	CallbackHash	string
	Receipt		ServiceReceipt
	EmissionHash	string
}

func ClassifyUnifiedInteraction(descriptor ServiceDescriptor, call UnifiedServiceCall) (UnifiedInteractionClass, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	method, found := descriptor.Interface.MethodByID(call.MethodID)
	if !found {
		return "", fmt.Errorf("services interaction method %s is not registered", call.MethodID)
	}
	if call.Kind == coretypes.ServiceCallKindCallback {
		return InteractionAsyncCallback, nil
	}
	if call.Kind == coretypes.ServiceCallKindRetry {
		return InteractionRetry, nil
	}
	if method.ExecutionType == coretypes.ServiceMethodEvented {
		return InteractionEventedSubscription, nil
	}
	switch descriptor.ServiceType {
	case coretypes.ServiceTypeOnChain:
		return InteractionOnChainTransaction, nil
	case coretypes.ServiceTypeOffChain, coretypes.ServiceTypeFogMarket:
		return InteractionOffChainServiceCall, nil
	case coretypes.ServiceTypeMixed:
		return InteractionHybridExecutionFlow, nil
	default:
		return "", fmt.Errorf("services interaction unknown service type %q", descriptor.ServiceType)
	}
}

func BuildUnifiedInteractionPlan(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, call UnifiedServiceCall) (UnifiedInteractionPlan, error) {
	if err := ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, call); err != nil {
		return UnifiedInteractionPlan{}, err
	}
	class, err := ClassifyUnifiedInteraction(descriptor, call)
	if err != nil {
		return UnifiedInteractionPlan{}, err
	}
	routing, err := RouteUnifiedServiceCall(ctx, descriptor, call)
	if err != nil {
		return UnifiedInteractionPlan{}, err
	}
	method, _ := descriptor.Interface.MethodByID(call.MethodID)
	plan := UnifiedInteractionPlan{
		CallID:			call.CallID,
		ServiceID:		call.TargetService,
		MethodID:		call.MethodID,
		InteractionClass:	class,
		Kind:			call.Kind,
		ExecutionType:		method.ExecutionType,
		Routes:			append([]UnifiedServiceRoute(nil), routing.Routes...),
		ReplaySafe:		call.Nonce != 0 && call.IdempotencyKey != "",
	}
	sortUnifiedRoutes(plan.Routes)
	plan.PlanHash = ComputeUnifiedInteractionPlanHash(plan)
	return plan, plan.Validate()
}

func NewUnifiedServiceCallback(ctx coretypes.ServiceConsensusContext, original UnifiedServiceCall, target ServiceDescriptor, callbackMethod, payloadHash, paymentPolicy string, deadline uint64, nonce uint64, idempotencyKey string) (UnifiedServiceCallback, error) {
	target = coretypes.CanonicalServiceDescriptor(target)
	if err := ctx.Validate(); err != nil {
		return UnifiedServiceCallback{}, err
	}
	if err := original.ValidateBasic(ctx); err != nil {
		return UnifiedServiceCallback{}, err
	}
	if err := target.Validate(); err != nil {
		return UnifiedServiceCallback{}, err
	}
	callback := UnifiedServiceCallback{
		OriginalCallID:		original.CallID,
		CallbackTarget:		target.ServiceID,
		CallbackMethod:		strings.TrimSpace(callbackMethod),
		CallbackPayloadHash:	strings.ToLower(strings.TrimSpace(payloadHash)),
		CallbackDeadline:	deadline,
		CallbackPaymentPolicy:	strings.TrimSpace(paymentPolicy),
		Caller:			original.Caller,
		Nonce:			nonce,
		IdempotencyKey:		strings.TrimSpace(idempotencyKey),
	}
	envelope := coretypes.NormalizeServiceCall(ctx, callback.ToServiceCallEnvelope(target))
	callback.CallbackCallID = envelope.CallID
	callback.CallbackHash = ComputeUnifiedServiceCallbackHash(callback)
	return callback, callback.ValidateForTarget(ctx, target)
}

func (callback UnifiedServiceCallback) ToServiceCallEnvelope(target ServiceDescriptor) coretypes.ServiceCallEnvelope {
	target = coretypes.CanonicalServiceDescriptor(target)
	method, _ := target.Interface.MethodByID(callback.CallbackMethod)
	methodID := callback.CallbackMethod
	interfaceHash := target.Interface.InterfaceHash
	paymentDenom := target.Payment.Denom
	proofRequirement := method.VerificationModel
	if method.MethodID != "" {
		methodID = method.MethodID
	}
	return coretypes.ServiceCallEnvelope{
		CallID:			strings.ToLower(strings.TrimSpace(callback.CallbackCallID)),
		ServiceID:		strings.TrimSpace(callback.CallbackTarget),
		Caller:			strings.TrimSpace(callback.Caller),
		Nonce:			callback.Nonce,
		IdempotencyKey:		strings.TrimSpace(callback.IdempotencyKey),
		MethodID:		strings.TrimSpace(methodID),
		InterfaceHash:		strings.ToLower(strings.TrimSpace(interfaceHash)),
		PayloadHash:		strings.ToLower(strings.TrimSpace(callback.CallbackPayloadHash)),
		PaymentDenom:		strings.TrimSpace(paymentDenom),
		MaxFeeAmount:		"0",
		ProofRequirement:	proofRequirement,
		Kind:			coretypes.ServiceCallKindCallback,
		CreatedHeight:		1,
		DeadlineHeight:		callback.CallbackDeadline,
		Callback:		true,
		RetryOf:		strings.ToLower(strings.TrimSpace(callback.OriginalCallID)),
		StateReadSet:		[]string{callback.CallbackTarget + "/" + methodID + "/callback/read"},
		StateWriteSet:		[]string{callback.CallbackTarget + "/" + methodID + "/callback/write"},
	}
}

func (callback UnifiedServiceCallback) ValidateForTarget(ctx coretypes.ServiceConsensusContext, target ServiceDescriptor) error {
	target = coretypes.CanonicalServiceDescriptor(target)
	if err := ctx.Validate(); err != nil {
		return err
	}
	if err := callback.ValidateBasic(); err != nil {
		return err
	}
	if callback.CallbackTarget != target.ServiceID {
		return errors.New("services callback target service mismatch")
	}
	method, found := target.Interface.MethodByID(callback.CallbackMethod)
	if !found {
		return fmt.Errorf("services callback method %s is not registered", callback.CallbackMethod)
	}
	if !method.CallbackSupported {
		return fmt.Errorf("services callback method %s does not support callbacks", callback.CallbackMethod)
	}
	if callback.CallbackDeadline < ctx.Height {
		return errors.New("services callback deadline is expired")
	}
	if method.IdempotencyRequired && callback.IdempotencyKey == "" {
		return errors.New("services callback requires idempotency key")
	}
	envelope := coretypes.NormalizeServiceCall(ctx, callback.ToServiceCallEnvelope(target))
	if callback.CallbackCallID != envelope.CallID {
		return fmt.Errorf("services callback call id mismatch: expected %s", envelope.CallID)
	}
	if err := envelope.ValidateBasic(ctx); err != nil {
		return err
	}
	if expected := ComputeUnifiedServiceCallbackHash(callback); callback.CallbackHash != expected {
		return fmt.Errorf("services callback hash mismatch: expected %s", expected)
	}
	return nil
}

func (callback UnifiedServiceCallback) ValidateBasic() error {
	if err := coretypes.ValidateHash("services callback original call id", callback.OriginalCallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services callback target", callback.CallbackTarget); err != nil {
		return err
	}
	if err := validateInterfaceToken("services callback method", callback.CallbackMethod); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services callback payload hash", callback.CallbackPayloadHash); err != nil {
		return err
	}
	if callback.CallbackDeadline == 0 {
		return errors.New("services callback deadline must be positive")
	}
	if err := validateInterfaceToken("services callback payment policy", callback.CallbackPaymentPolicy); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("services callback caller", callback.Caller); err != nil {
		return err
	}
	if callback.Nonce == 0 {
		return errors.New("services callback nonce must be positive")
	}
	if callback.IdempotencyKey == "" {
		return errors.New("services callback must be replay-safe with idempotency key")
	}
	if err := validateInterfaceToken("services callback idempotency key", callback.IdempotencyKey); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services callback call id", callback.CallbackCallID); err != nil {
		return err
	}
	return coretypes.ValidateHash("services callback hash", callback.CallbackHash)
}

func EmitServiceCallbackReceipt(ctx coretypes.ServiceConsensusContext, target ServiceDescriptor, callback UnifiedServiceCallback, outcome ServiceExecutionOutcome) (ServiceCallbackReceiptEmission, error) {
	if err := callback.ValidateForTarget(ctx, target); err != nil {
		return ServiceCallbackReceiptEmission{}, err
	}
	outcome = coretypes.NormalizeServiceExecutionOutcome(ctx, outcome)
	if outcome.CallID != callback.CallbackCallID {
		return ServiceCallbackReceiptEmission{}, errors.New("services callback outcome call mismatch")
	}
	envelope := callback.ToServiceCallEnvelope(target)
	envelope = coretypes.NormalizeServiceCall(ctx, envelope)
	receipt, err := coretypes.NewServiceCallReceipt(envelope, outcome)
	if err != nil {
		return ServiceCallbackReceiptEmission{}, err
	}
	emission := ServiceCallbackReceiptEmission{
		CallbackHash:	callback.CallbackHash,
		Receipt:	receipt,
	}
	emission.EmissionHash = ComputeServiceCallbackReceiptEmissionHash(emission)
	return emission, emission.Validate()
}

func (plan UnifiedInteractionPlan) Validate() error {
	if err := coretypes.ValidateHash("services interaction plan call id", plan.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interaction plan service id", plan.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interaction plan method id", plan.MethodID); err != nil {
		return err
	}
	if !IsUnifiedInteractionClass(plan.InteractionClass) {
		return fmt.Errorf("services interaction unknown class %q", plan.InteractionClass)
	}
	if !coretypes.IsServiceCallKind(plan.Kind) {
		return fmt.Errorf("services interaction unknown kind %q", plan.Kind)
	}
	if !coretypes.IsServiceMethodExecutionType(plan.ExecutionType) {
		return fmt.Errorf("services interaction unknown execution type %q", plan.ExecutionType)
	}
	if len(plan.Routes) == 0 {
		return errors.New("services interaction plan requires routes")
	}
	if err := validateUnifiedRoutes(plan.Routes); err != nil {
		return err
	}
	if !plan.ReplaySafe {
		return errors.New("services interaction plan must be replay-safe")
	}
	if err := coretypes.ValidateHash("services interaction plan hash", plan.PlanHash); err != nil {
		return err
	}
	if expected := ComputeUnifiedInteractionPlanHash(plan); plan.PlanHash != expected {
		return fmt.Errorf("services interaction plan hash mismatch: expected %s", expected)
	}
	return nil
}

func (emission ServiceCallbackReceiptEmission) Validate() error {
	if err := coretypes.ValidateHash("services callback emission callback hash", emission.CallbackHash); err != nil {
		return err
	}
	if err := emission.Receipt.Validate(); err != nil {
		return err
	}
	if emission.Receipt.Status == coretypes.ServiceCallStatusAccepted {
		return errors.New("services callback execution must emit terminal receipt")
	}
	if err := coretypes.ValidateHash("services callback emission hash", emission.EmissionHash); err != nil {
		return err
	}
	if expected := ComputeServiceCallbackReceiptEmissionHash(emission); emission.EmissionHash != expected {
		return fmt.Errorf("services callback emission hash mismatch: expected %s", expected)
	}
	return nil
}

func IsUnifiedInteractionClass(class UnifiedInteractionClass) bool {
	switch class {
	case InteractionOnChainTransaction, InteractionOffChainServiceCall, InteractionHybridExecutionFlow,
		InteractionAsyncCallback, InteractionRetry, InteractionEventedSubscription:
		return true
	default:
		return false
	}
}

func ComputeUnifiedInteractionPlanHash(plan UnifiedInteractionPlan) string {
	routes := append([]UnifiedServiceRoute(nil), plan.Routes...)
	sortUnifiedRoutes(routes)
	parts := []string{
		"aetra-services-interaction-plan-v1",
		plan.CallID,
		plan.ServiceID,
		plan.MethodID,
		string(plan.InteractionClass),
		string(plan.Kind),
		string(plan.ExecutionType),
		fmt.Sprint(plan.ReplaySafe),
	}
	for _, route := range routes {
		parts = append(parts, string(route))
	}
	return servicesHashParts(parts...)
}

func ComputeUnifiedServiceCallbackHash(callback UnifiedServiceCallback) string {
	return servicesHashParts(
		"aetra-services-callback-v1",
		callback.OriginalCallID,
		callback.CallbackTarget,
		callback.CallbackMethod,
		callback.CallbackPayloadHash,
		fmt.Sprint(callback.CallbackDeadline),
		callback.CallbackPaymentPolicy,
		callback.Caller,
		fmt.Sprint(callback.Nonce),
		callback.IdempotencyKey,
		callback.CallbackCallID,
	)
}

func ComputeServiceCallbackReceiptEmissionHash(emission ServiceCallbackReceiptEmission) string {
	return servicesHashParts("aetra-services-callback-receipt-emission-v1", emission.CallbackHash, emission.Receipt.ReceiptHash)
}
