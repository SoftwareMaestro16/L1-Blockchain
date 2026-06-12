package types

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type UnifiedServiceRoute string

const (
	UnifiedRouteDeliverTx		UnifiedServiceRoute	= "DELIVER_TX"
	UnifiedRouteFinalizeBlock	UnifiedServiceRoute	= "FINALIZE_BLOCK"
	UnifiedRouteServiceNetwork	UnifiedServiceRoute	= "SERVICE_NETWORK"
	UnifiedRouteOnChainCommitment	UnifiedServiceRoute	= "ON_CHAIN_COMMITMENT"
	UnifiedRouteOnChainPayment	UnifiedServiceRoute	= "ON_CHAIN_PAYMENT"
	UnifiedRouteOnChainDispute	UnifiedServiceRoute	= "ON_CHAIN_DISPUTE"
	UnifiedRouteOnChainSettlement	UnifiedServiceRoute	= "ON_CHAIN_SETTLEMENT"
	UnifiedRouteProofVerification	UnifiedServiceRoute	= "PROOF_VERIFICATION"
)

type UnifiedServicePayment struct {
	Model		string
	Denom		string
	MaxFee		string
	Reserve		bool
	Escrow		bool
	PaymentHash	string
}

type UnifiedServiceCall struct {
	CallID			string
	TargetService		string
	Method			string
	PayloadHash		string
	Payment			UnifiedServicePayment
	ProofRequirement	coretypes.ServiceVerificationModel
	TimeoutHeight		uint64
	SignatureHash		string

	Caller			string
	InterfaceHash		string
	MethodID		string
	ExecutionLocation	coretypes.ServiceLocation
	IdempotencyKey		string
	CallbackTarget		string
	MaxFee			string
	CreatedHeight		uint64
	DeadlineHeight		uint64
	Nonce			uint64
	Kind			coretypes.ServiceCallKind
	RetryOf			string
	StateReadSet		[]string
	StateWriteSet		[]string
	UnifiedCallHash		string
}

type UnifiedCallRoutingPlan struct {
	CallID				string
	TargetService			string
	MethodID			string
	ServiceType			coretypes.ServiceType
	ExecutionLocation		coretypes.ServiceLocation
	Kind				coretypes.ServiceCallKind
	Routes				[]UnifiedServiceRoute
	ReserveFundsBeforeExecution	bool
	VerifyResultProofBeforeAccept	bool
	CommitResultOnChain		bool
	DisputeEligible			bool
	PaymentRouteRequired		bool
	OffChainExecutionRequired	bool
	OnChainExecutionRequired	bool
	ConsensusAcceptanceRoute	UnifiedServiceRoute
	RoutingHash			string
}

func NewUnifiedServiceCall(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, methodID, caller string, nonce uint64, payloadHash, maxFeeAmount, signatureHash string, timeoutDelta uint64, idempotencyKey, callbackTarget string) (UnifiedServiceCall, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := ctx.Validate(); err != nil {
		return UnifiedServiceCall{}, err
	}
	if err := descriptor.Validate(); err != nil {
		return UnifiedServiceCall{}, err
	}
	if timeoutDelta == 0 {
		return UnifiedServiceCall{}, errors.New("services unified call timeout must be positive")
	}
	method, found := descriptor.Interface.MethodByID(methodID)
	if !found {
		return UnifiedServiceCall{}, fmt.Errorf("services unified call method %s is not registered", methodID)
	}
	kind, err := defaultUnifiedCallKind(descriptor)
	if err != nil {
		return UnifiedServiceCall{}, err
	}
	deadline := ctx.Height + timeoutDelta
	payment := UnifiedServicePayment{
		Model:		registryPaymentModelFromDescriptor(descriptor),
		Denom:		descriptor.Payment.Denom,
		MaxFee:		strings.TrimSpace(maxFeeAmount),
		Reserve:	servicePaymentRequiresReserve(descriptor.Payment.SettlementMode),
		Escrow:		descriptor.Payment.EscrowRequired || descriptor.Payment.SettlementMode == coretypes.ServicePaymentEscrow,
	}
	payment.PaymentHash = ComputeUnifiedServicePaymentHash(payment)
	call := UnifiedServiceCall{
		TargetService:		descriptor.ServiceID,
		Method:			method.Name,
		PayloadHash:		strings.ToLower(strings.TrimSpace(payloadHash)),
		Payment:		payment,
		ProofRequirement:	method.VerificationModel,
		TimeoutHeight:		timeoutDelta,
		SignatureHash:		strings.ToLower(strings.TrimSpace(signatureHash)),
		Caller:			strings.TrimSpace(caller),
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		MethodID:		method.MethodID,
		ExecutionLocation:	descriptor.Execution.Location,
		IdempotencyKey:		strings.TrimSpace(idempotencyKey),
		CallbackTarget:		strings.TrimSpace(callbackTarget),
		MaxFee:			strings.TrimSpace(maxFeeAmount),
		CreatedHeight:		ctx.Height,
		DeadlineHeight:		deadline,
		Nonce:			nonce,
		Kind:			kind,
		RetryOf:		"",
		StateReadSet:		[]string{descriptor.ServiceID + "/" + method.MethodID + "/read"},
		StateWriteSet:		[]string{descriptor.ServiceID + "/" + method.MethodID + "/write"},
	}
	envelope := coretypes.NormalizeServiceCall(ctx, call.ToServiceCallEnvelope())
	call.CallID = envelope.CallID
	call.UnifiedCallHash = ComputeUnifiedServiceCallHash(call)
	return call, ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, call)
}

func (call UnifiedServiceCall) ToServiceCallEnvelope() coretypes.ServiceCallEnvelope {
	return coretypes.ServiceCallEnvelope{
		CallID:			strings.ToLower(strings.TrimSpace(call.CallID)),
		ServiceID:		strings.TrimSpace(call.TargetService),
		Caller:			strings.TrimSpace(call.Caller),
		Nonce:			call.Nonce,
		IdempotencyKey:		strings.TrimSpace(call.IdempotencyKey),
		MethodID:		strings.TrimSpace(call.MethodID),
		InterfaceHash:		strings.ToLower(strings.TrimSpace(call.InterfaceHash)),
		PayloadHash:		strings.ToLower(strings.TrimSpace(call.PayloadHash)),
		PaymentDenom:		strings.TrimSpace(call.Payment.Denom),
		MaxFeeAmount:		strings.TrimSpace(call.MaxFee),
		ProofRequirement:	call.ProofRequirement,
		Kind:			call.Kind,
		CreatedHeight:		call.CreatedHeight,
		DeadlineHeight:		call.DeadlineHeight,
		Callback:		call.CallbackTarget != "",
		RetryOf:		strings.ToLower(strings.TrimSpace(call.RetryOf)),
		StateReadSet:		append([]string(nil), call.StateReadSet...),
		StateWriteSet:		append([]string(nil), call.StateWriteSet...),
	}
}

func ValidateUnifiedServiceCallForDescriptor(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, call UnifiedServiceCall) error {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := ctx.Validate(); err != nil {
		return err
	}
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if err := call.ValidateBasic(ctx); err != nil {
		return err
	}
	if call.TargetService != descriptor.ServiceID {
		return errors.New("services unified call target service mismatch")
	}
	if !descriptor.Enabled || descriptor.Status != coretypes.ServiceStatusActive {
		return fmt.Errorf("services unified call target %s is not active", descriptor.ServiceID)
	}
	if descriptor.ExpiryHeight != 0 && call.DeadlineHeight > descriptor.ExpiryHeight {
		return errors.New("services unified call deadline must not exceed service expiry")
	}
	if call.InterfaceHash != descriptor.Interface.InterfaceHash {
		return errors.New("services unified call interface hash mismatch")
	}
	if call.ExecutionLocation != descriptor.Execution.Location {
		return errors.New("services unified call execution location mismatch")
	}
	method, found := descriptor.Interface.MethodByID(call.MethodID)
	if !found {
		return fmt.Errorf("services unified call method %s is not registered", call.MethodID)
	}
	if call.Method != method.Name {
		return errors.New("services unified call method name mismatch")
	}
	if method.IdempotencyRequired && call.IdempotencyKey == "" {
		return errors.New("services unified call requires idempotency key")
	}
	if call.ProofRequirement != method.VerificationModel {
		return errors.New("services unified call proof requirement mismatch")
	}
	if call.Payment.Denom != descriptor.Payment.Denom {
		return errors.New("services unified call payment denom mismatch")
	}
	if call.Payment.Model != registryPaymentModelFromDescriptor(descriptor) {
		return errors.New("services unified call payment model mismatch")
	}
	if call.Payment.Reserve != servicePaymentRequiresReserve(descriptor.Payment.SettlementMode) {
		return errors.New("services unified call payment reserve policy mismatch")
	}
	if call.Payment.Escrow != (descriptor.Payment.EscrowRequired || descriptor.Payment.SettlementMode == coretypes.ServicePaymentEscrow) {
		return errors.New("services unified call escrow policy mismatch")
	}
	if !unifiedCallKindAllowed(descriptor.ServiceType, call.Kind) {
		return fmt.Errorf("services unified call kind %s is incompatible with %s service", call.Kind, descriptor.ServiceType)
	}
	if call.Kind == coretypes.ServiceCallKindRetry {
		if call.RetryOf == "" {
			return errors.New("services unified retry call must reference original call id")
		}
		if call.IdempotencyKey == "" {
			return errors.New("services unified retry call requires idempotency key")
		}
	}
	if err := coretypes.NormalizeServiceCall(ctx, call.ToServiceCallEnvelope()).ValidateBasic(ctx); err != nil {
		return err
	}
	return nil
}

func (call UnifiedServiceCall) ValidateBasic(ctx coretypes.ServiceConsensusContext) error {
	if err := validateInterfaceToken("services unified call target service", call.TargetService); err != nil {
		return err
	}
	if err := validateInterfaceToken("services unified call method", call.Method); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services unified call payload hash", call.PayloadHash); err != nil {
		return err
	}
	if err := call.Payment.Validate(); err != nil {
		return err
	}
	if !coretypes.IsServiceVerificationModel(call.ProofRequirement) {
		return fmt.Errorf("services unified call unknown proof requirement %q", call.ProofRequirement)
	}
	if call.TimeoutHeight == 0 {
		return errors.New("services unified call timeout must be positive")
	}
	if err := coretypes.ValidateHash("services unified call signature hash", call.SignatureHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services unified call id", call.CallID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("services unified call caller", call.Caller); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services unified call interface hash", call.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("services unified call method id", call.MethodID); err != nil {
		return err
	}
	if !coretypes.IsServiceLocation(call.ExecutionLocation) {
		return fmt.Errorf("services unified call unknown execution location %q", call.ExecutionLocation)
	}
	if call.IdempotencyKey != "" {
		if err := validateInterfaceToken("services unified call idempotency key", call.IdempotencyKey); err != nil {
			return err
		}
	}
	if call.CallbackTarget != "" {
		if err := validateInterfaceToken("services unified call callback target", call.CallbackTarget); err != nil {
			return err
		}
	}
	if call.MaxFee != call.Payment.MaxFee {
		return errors.New("services unified call max fee field mismatch")
	}
	if call.CreatedHeight == 0 || call.CreatedHeight > ctx.Height {
		return errors.New("services unified call created height is invalid")
	}
	if call.DeadlineHeight == 0 || call.DeadlineHeight < ctx.Height {
		return errors.New("services unified call is expired")
	}
	if call.TimeoutHeight != call.DeadlineHeight-call.CreatedHeight {
		return errors.New("services unified call timeout and deadline mismatch")
	}
	if call.Nonce == 0 {
		return errors.New("services unified call nonce must be positive")
	}
	if !coretypes.IsServiceCallKind(call.Kind) {
		return fmt.Errorf("services unified call unknown kind %q", call.Kind)
	}
	if call.RetryOf != "" {
		if err := coretypes.ValidateHash("services unified call retry original id", call.RetryOf); err != nil {
			return err
		}
	}
	if err := validateUnifiedStateKeys("services unified call read set", call.StateReadSet); err != nil {
		return err
	}
	if err := validateUnifiedStateKeys("services unified call write set", call.StateWriteSet); err != nil {
		return err
	}
	expected := coretypes.NormalizeServiceCall(ctx, call.ToServiceCallEnvelope()).CallID
	if call.CallID != expected {
		return fmt.Errorf("services unified call id mismatch: expected %s", expected)
	}
	if err := coretypes.ValidateHash("services unified call hash", call.UnifiedCallHash); err != nil {
		return err
	}
	if expectedHash := ComputeUnifiedServiceCallHash(call); call.UnifiedCallHash != expectedHash {
		return fmt.Errorf("services unified call hash mismatch: expected %s", expectedHash)
	}
	return nil
}

func (payment UnifiedServicePayment) Validate() error {
	if err := validateInterfaceToken("services unified call payment model", payment.Model); err != nil {
		return err
	}
	if err := validateInterfaceToken("services unified call payment denom", payment.Denom); err != nil {
		return err
	}
	if err := validateUnifiedAmount("services unified call max fee", payment.MaxFee); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services unified call payment hash", payment.PaymentHash); err != nil {
		return err
	}
	if expected := ComputeUnifiedServicePaymentHash(payment); payment.PaymentHash != expected {
		return fmt.Errorf("services unified call payment hash mismatch: expected %s", expected)
	}
	return nil
}

func RouteUnifiedServiceCall(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, call UnifiedServiceCall) (UnifiedCallRoutingPlan, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, call); err != nil {
		return UnifiedCallRoutingPlan{}, err
	}
	routes, err := unifiedRoutesForDescriptor(descriptor, call)
	if err != nil {
		return UnifiedCallRoutingPlan{}, err
	}
	sortUnifiedRoutes(routes)
	plan := UnifiedCallRoutingPlan{
		CallID:				call.CallID,
		TargetService:			call.TargetService,
		MethodID:			call.MethodID,
		ServiceType:			descriptor.ServiceType,
		ExecutionLocation:		call.ExecutionLocation,
		Kind:				call.Kind,
		Routes:				routes,
		ReserveFundsBeforeExecution:	call.Payment.Reserve,
		VerifyResultProofBeforeAccept:	serviceProofRequiresResultVerification(call.ProofRequirement),
		CommitResultOnChain:		serviceResultRequiresOnChainCommitment(descriptor, call),
		DisputeEligible:		descriptor.ServiceType == coretypes.ServiceTypeMixed && descriptor.Verification.ChallengeWindow != 0,
		PaymentRouteRequired:		call.Payment.Reserve || call.Payment.Escrow,
		OffChainExecutionRequired:	descriptor.ServiceType == coretypes.ServiceTypeOffChain || descriptor.ServiceType == coretypes.ServiceTypeMixed,
		OnChainExecutionRequired:	descriptor.ServiceType == coretypes.ServiceTypeOnChain,
		ConsensusAcceptanceRoute:	consensusAcceptanceRouteForDescriptor(descriptor),
	}
	plan.RoutingHash = ComputeUnifiedCallRoutingPlanHash(plan)
	return plan, plan.Validate()
}

func (plan UnifiedCallRoutingPlan) Validate() error {
	if err := coretypes.ValidateHash("services unified call routing call id", plan.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services unified call routing target", plan.TargetService); err != nil {
		return err
	}
	if err := validateInterfaceToken("services unified call routing method id", plan.MethodID); err != nil {
		return err
	}
	if !coretypes.IsServiceType(plan.ServiceType) {
		return fmt.Errorf("services unified call routing unknown service type %q", plan.ServiceType)
	}
	if !coretypes.IsServiceLocation(plan.ExecutionLocation) {
		return fmt.Errorf("services unified call routing unknown location %q", plan.ExecutionLocation)
	}
	if !coretypes.IsServiceCallKind(plan.Kind) {
		return fmt.Errorf("services unified call routing unknown kind %q", plan.Kind)
	}
	if len(plan.Routes) == 0 {
		return errors.New("services unified call routing requires routes")
	}
	if err := validateUnifiedRoutes(plan.Routes); err != nil {
		return err
	}
	if plan.PaymentRouteRequired && !containsUnifiedRoute(plan.Routes, UnifiedRouteOnChainPayment) {
		return errors.New("services unified call routing payment reserve requires on-chain payment route")
	}
	if plan.VerifyResultProofBeforeAccept && !containsUnifiedRoute(plan.Routes, UnifiedRouteProofVerification) {
		return errors.New("services unified call routing proof verification route is required")
	}
	if plan.CommitResultOnChain && !containsUnifiedRoute(plan.Routes, UnifiedRouteOnChainCommitment) {
		return errors.New("services unified call routing commitment route is required")
	}
	if plan.ServiceType == coretypes.ServiceTypeOnChain && (!containsUnifiedRoute(plan.Routes, UnifiedRouteDeliverTx) || !containsUnifiedRoute(plan.Routes, UnifiedRouteFinalizeBlock)) {
		return errors.New("services unified on-chain call must route through DeliverTx and FinalizeBlock")
	}
	if plan.ServiceType == coretypes.ServiceTypeOffChain && !containsUnifiedRoute(plan.Routes, UnifiedRouteServiceNetwork) {
		return errors.New("services unified off-chain call must route to service network")
	}
	if plan.ServiceType == coretypes.ServiceTypeMixed && (!containsUnifiedRoute(plan.Routes, UnifiedRouteServiceNetwork) || !containsUnifiedRoute(plan.Routes, UnifiedRouteOnChainCommitment)) {
		return errors.New("services unified mixed call must split network execution and on-chain commitment")
	}
	if !containsUnifiedRoute(plan.Routes, plan.ConsensusAcceptanceRoute) {
		return errors.New("services unified routing missing consensus acceptance route")
	}
	if err := coretypes.ValidateHash("services unified call routing hash", plan.RoutingHash); err != nil {
		return err
	}
	if expected := ComputeUnifiedCallRoutingPlanHash(plan); plan.RoutingHash != expected {
		return fmt.Errorf("services unified call routing hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeUnifiedServicePaymentHash(payment UnifiedServicePayment) string {
	return servicesHashParts(
		"aetra-services-unified-payment-v1",
		payment.Model,
		payment.Denom,
		payment.MaxFee,
		fmt.Sprint(payment.Reserve),
		fmt.Sprint(payment.Escrow),
	)
}

func ComputeUnifiedServiceCallHash(call UnifiedServiceCall) string {
	parts := []string{
		"aetra-services-unified-call-v1",
		call.CallID,
		call.TargetService,
		call.Method,
		call.PayloadHash,
		call.Payment.PaymentHash,
		string(call.ProofRequirement),
		fmt.Sprint(call.TimeoutHeight),
		call.SignatureHash,
		call.Caller,
		call.InterfaceHash,
		call.MethodID,
		string(call.ExecutionLocation),
		call.IdempotencyKey,
		call.CallbackTarget,
		call.MaxFee,
		fmt.Sprint(call.CreatedHeight),
		fmt.Sprint(call.DeadlineHeight),
		fmt.Sprint(call.Nonce),
		string(call.Kind),
		call.RetryOf,
	}
	parts = appendStringParts(parts, "reads", call.StateReadSet)
	parts = appendStringParts(parts, "writes", call.StateWriteSet)
	return servicesHashParts(parts...)
}

func ComputeUnifiedCallRoutingPlanHash(plan UnifiedCallRoutingPlan) string {
	routes := append([]UnifiedServiceRoute(nil), plan.Routes...)
	sortUnifiedRoutes(routes)
	parts := []string{
		"aetra-services-unified-routing-v1",
		plan.CallID,
		plan.TargetService,
		plan.MethodID,
		string(plan.ServiceType),
		string(plan.ExecutionLocation),
		string(plan.Kind),
		fmt.Sprint(plan.ReserveFundsBeforeExecution),
		fmt.Sprint(plan.VerifyResultProofBeforeAccept),
		fmt.Sprint(plan.CommitResultOnChain),
		fmt.Sprint(plan.DisputeEligible),
		fmt.Sprint(plan.PaymentRouteRequired),
		fmt.Sprint(plan.OffChainExecutionRequired),
		fmt.Sprint(plan.OnChainExecutionRequired),
		string(plan.ConsensusAcceptanceRoute),
	}
	for _, route := range routes {
		parts = append(parts, string(route))
	}
	return servicesHashParts(parts...)
}

func defaultUnifiedCallKind(descriptor ServiceDescriptor) (coretypes.ServiceCallKind, error) {
	switch descriptor.ServiceType {
	case coretypes.ServiceTypeOnChain:
		return coretypes.ServiceCallKindOnChain, nil
	case coretypes.ServiceTypeOffChain, coretypes.ServiceTypeFogMarket:
		return coretypes.ServiceCallKindOffChainReceipt, nil
	case coretypes.ServiceTypeMixed:
		return coretypes.ServiceCallKindOffChainReceipt, nil
	default:
		return "", fmt.Errorf("services unified call unknown service type %q", descriptor.ServiceType)
	}
}

func unifiedRoutesForDescriptor(descriptor ServiceDescriptor, call UnifiedServiceCall) ([]UnifiedServiceRoute, error) {
	routes := []UnifiedServiceRoute{}
	if call.Payment.Reserve || call.Payment.Escrow {
		routes = append(routes, UnifiedRouteOnChainPayment)
	}
	if serviceProofRequiresResultVerification(call.ProofRequirement) {
		routes = append(routes, UnifiedRouteProofVerification)
	}
	switch descriptor.ServiceType {
	case coretypes.ServiceTypeOnChain:
		routes = append(routes, UnifiedRouteDeliverTx, UnifiedRouteFinalizeBlock)
	case coretypes.ServiceTypeOffChain, coretypes.ServiceTypeFogMarket:
		routes = append(routes, UnifiedRouteServiceNetwork)
	case coretypes.ServiceTypeMixed:
		routes = append(routes, UnifiedRouteServiceNetwork, UnifiedRouteOnChainCommitment)
		switch call.Kind {
		case coretypes.ServiceCallKindMixedDispute:
			routes = append(routes, UnifiedRouteOnChainDispute)
		case coretypes.ServiceCallKindMixedSettlement:
			routes = append(routes, UnifiedRouteOnChainSettlement)
		}
	default:
		return nil, fmt.Errorf("services unified call unknown service type %q", descriptor.ServiceType)
	}
	return routes, nil
}

func servicePaymentRequiresReserve(mode coretypes.ServicePaymentSettlementMode) bool {
	switch mode {
	case coretypes.ServicePaymentOnChain, coretypes.ServicePaymentPrepaid, coretypes.ServicePaymentMetered, coretypes.ServicePaymentEscrow:
		return true
	default:
		return false
	}
}

func serviceProofRequiresResultVerification(model coretypes.ServiceVerificationModel) bool {
	return model != coretypes.ServiceVerificationAdvisory
}

func serviceResultRequiresOnChainCommitment(descriptor ServiceDescriptor, call UnifiedServiceCall) bool {
	if descriptor.ServiceType == coretypes.ServiceTypeMixed {
		return true
	}
	return descriptor.ServiceType == coretypes.ServiceTypeOffChain && call.ProofRequirement == coretypes.ServiceVerificationProofAnchored
}

func consensusAcceptanceRouteForDescriptor(descriptor ServiceDescriptor) UnifiedServiceRoute {
	switch descriptor.ServiceType {
	case coretypes.ServiceTypeOnChain:
		return UnifiedRouteFinalizeBlock
	case coretypes.ServiceTypeMixed:
		return UnifiedRouteOnChainCommitment
	default:
		return UnifiedRouteServiceNetwork
	}
}

func unifiedCallKindAllowed(serviceType coretypes.ServiceType, kind coretypes.ServiceCallKind) bool {
	switch kind {
	case coretypes.ServiceCallKindCallback, coretypes.ServiceCallKindRetry:
		return true
	}
	switch serviceType {
	case coretypes.ServiceTypeOnChain:
		return kind == coretypes.ServiceCallKindOnChain
	case coretypes.ServiceTypeOffChain, coretypes.ServiceTypeFogMarket:
		return kind == coretypes.ServiceCallKindOffChainReceipt
	case coretypes.ServiceTypeMixed:
		return kind == coretypes.ServiceCallKindOffChainReceipt ||
			kind == coretypes.ServiceCallKindMixedDispute ||
			kind == coretypes.ServiceCallKindMixedSettlement
	default:
		return false
	}
}

func validateUnifiedAmount(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	n, ok := new(big.Int).SetString(value, 10)
	if !ok || n.Sign() < 0 {
		return fmt.Errorf("%s must be a non-negative integer", fieldName)
	}
	return nil
}

func validateUnifiedStateKeys(fieldName string, keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("%s requires at least one key", fieldName)
	}
	previous := ""
	seen := map[string]struct{}{}
	for _, key := range keys {
		if err := validateInterfaceToken(fieldName, key); err != nil {
			return err
		}
		if _, found := seen[key]; found {
			return fmt.Errorf("%s contains duplicate key %s", fieldName, key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
		previous = key
	}
	return nil
}

func validateUnifiedRoutes(routes []UnifiedServiceRoute) error {
	previous := ""
	seen := map[UnifiedServiceRoute]struct{}{}
	for _, route := range routes {
		if !isUnifiedServiceRoute(route) {
			return fmt.Errorf("services unified call unknown route %q", route)
		}
		if _, found := seen[route]; found {
			return fmt.Errorf("services unified call duplicate route %q", route)
		}
		seen[route] = struct{}{}
		current := string(route)
		if previous != "" && previous >= current {
			return errors.New("services unified call routes must be sorted canonically")
		}
		previous = current
	}
	return nil
}

func isUnifiedServiceRoute(route UnifiedServiceRoute) bool {
	switch route {
	case UnifiedRouteDeliverTx, UnifiedRouteFinalizeBlock, UnifiedRouteServiceNetwork,
		UnifiedRouteOnChainCommitment, UnifiedRouteOnChainPayment, UnifiedRouteOnChainDispute,
		UnifiedRouteOnChainSettlement, UnifiedRouteProofVerification:
		return true
	default:
		return false
	}
}

func containsUnifiedRoute(routes []UnifiedServiceRoute, route UnifiedServiceRoute) bool {
	for _, existing := range routes {
		if existing == route {
			return true
		}
	}
	return false
}

func sortUnifiedRoutes(routes []UnifiedServiceRoute) {
	sort.SliceStable(routes, func(i, j int) bool { return routes[i] < routes[j] })
}
