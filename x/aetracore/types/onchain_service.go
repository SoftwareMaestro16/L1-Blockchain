package types

import (
	"errors"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	DefaultOnChainSchemaEncoding	= "json-schema-v1"
	DefaultOnChainAuthModel		= "aetra-account"
	DefaultOnChainPaymentModel	= "naet-on-chain"
)

type OnChainServiceWrapper struct {
	ServiceID		string
	Owner			string
	ZoneID			ZoneID
	ModuleName		string
	ContractAddress		string
	InterfaceID		string
	InterfaceName		string
	EndpointKey		string
	Version			uint64
	AvailabilityHash	string
	StateRootType		RootType
	ReceiptPolicy		ServiceReceiptPolicy
	PaymentDenom		string
	PaymentAmount		string
	Methods			[]OnChainServiceMethod
	MetadataHash		string
	CreatedHeight		uint64
	UpdatedHeight		uint64
	ExpiryHeight		uint64
}

type OnChainServiceMethod struct {
	MethodID		string
	Name			string
	InputSchemaHash		string
	OutputSchemaHash	string
	GasModel		string
	RequiredPaymentModel	string
	FailurePolicy		ServiceFailureBehavior
	TimeoutHeightDelta	uint64
	IdempotencyRequired	bool
	CallbackSupported	bool
}

type ServiceStateProofQuery struct {
	ServiceID	string
	Height		uint64
	StateRootType	RootType
	KeyHash		string
	ProofRoot	string
	QueryHash	string
}

func BuildOnChainServiceDescriptor(wrapper OnChainServiceWrapper) (ServiceDescriptor, error) {
	wrapper = CanonicalOnChainServiceWrapper(wrapper)
	if err := wrapper.Validate(); err != nil {
		return ServiceDescriptor{}, err
	}
	methods := make([]ServiceMethodDescriptor, len(wrapper.Methods))
	for i, method := range wrapper.Methods {
		methods[i] = method.ServiceMethodDescriptor()
	}
	interfaceDescriptor := ServiceInterfaceDescriptor{
		InterfaceID:	wrapper.InterfaceID,
		InterfaceName:	wrapper.InterfaceName,
		Version:	wrapper.Version,
		SchemaEncoding:	DefaultOnChainSchemaEncoding,
		Methods:	methods,
		Events:		[]string{"service.receipt"},
		Errors:		[]string{"service.error"},
		AuthModel:	DefaultOnChainAuthModel,
		PaymentModel:	DefaultOnChainPaymentModel,
		MetadataHash:	wrapper.MetadataHash,
		CreatedHeight:	wrapper.CreatedHeight,
	}
	interfaceDescriptor = CanonicalServiceInterfaceDescriptor(interfaceDescriptor)
	interfaceDescriptor.InterfaceHash = ComputeServiceInterfaceHash(interfaceDescriptor)

	location := ServiceLocationModule
	target := wrapper.ModuleName
	if wrapper.ContractAddress != "" {
		location = ServiceLocationContract
		target = wrapper.ContractAddress
	}
	descriptor := ServiceDescriptor{
		ServiceID:		wrapper.ServiceID,
		Owner:			wrapper.Owner,
		ServiceType:		ServiceTypeOnChain,
		ZoneID:			wrapper.ZoneID,
		InterfaceID:		wrapper.InterfaceID,
		EndpointKey:		wrapper.EndpointKey,
		Version:		wrapper.Version,
		AvailabilityHash:	wrapper.AvailabilityHash,
		Enabled:		true,
		Status:			ServiceStatusActive,
		ExpiryHeight:		wrapper.ExpiryHeight,
		CreatedHeight:		wrapper.CreatedHeight,
		UpdatedHeight:		wrapper.UpdatedHeight,
		Interface:		interfaceDescriptor,
		Execution: ServiceExecutionDescriptor{
			Location:		location,
			Target:			target,
			ModuleRoute:		wrapper.ModuleName,
			ContractAddress:	wrapper.ContractAddress,
			Mode:			ExecutionModeSync,
			Deterministic:		true,
			ReceiptPolicy:		wrapper.ReceiptPolicy,
			FailureBehavior:	aggregateOnChainFailurePolicy(wrapper.Methods),
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:	wrapper.ServiceID,
			MetadataHash:	wrapper.MetadataHash,
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode:	ServicePaymentOnChain,
			Denom:		wrapper.PaymentDenom,
			Amount:		wrapper.PaymentAmount,
			PricingUnit:	ServicePricingPerCall,
		},
		Storage: ServiceStorageDescriptor{
			Model:		ServiceStorageOnChain,
			StateRootType:	wrapper.StateRootType,
			ProofRequired:	wrapper.ReceiptPolicy == ServiceReceiptCommittedAndProof,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:	ServiceTrustConsensusExecuted,
			Model:		ServiceVerificationConsensusReceipt,
		},
	}
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceDescriptor{}, err
	}
	return descriptor, nil
}

func BuildModuleMsgServiceDescriptor(wrapper OnChainServiceWrapper) (ServiceDescriptor, error) {
	if wrapper.ModuleName == "" {
		return ServiceDescriptor{}, errors.New("aetracore module MsgServer service requires module name")
	}
	if wrapper.ContractAddress != "" {
		return ServiceDescriptor{}, errors.New("aetracore module MsgServer service must not define contract address")
	}
	return BuildOnChainServiceDescriptor(wrapper)
}

func BuildContractServiceDescriptor(wrapper OnChainServiceWrapper) (ServiceDescriptor, error) {
	if wrapper.ContractAddress == "" {
		return ServiceDescriptor{}, errors.New("aetracore contract service requires contract address")
	}
	if wrapper.ModuleName != "" {
		return ServiceDescriptor{}, errors.New("aetracore contract service must not define module name")
	}
	return BuildOnChainServiceDescriptor(wrapper)
}

func NewOnChainServiceReceipt(ctx ServiceConsensusContext, state CoreState, call ServiceCallEnvelope, result ExecutionResult, gasUsed uint64) (ServiceCallReceipt, error) {
	call = NormalizeServiceCall(ctx, call)
	if err := ValidateServiceCallForState(ctx, state, call); err != nil {
		return ServiceCallReceipt{}, err
	}
	if call.Kind != ServiceCallKindOnChain {
		return ServiceCallReceipt{}, errors.New("aetracore on-chain service receipt requires on-chain call kind")
	}
	descriptor, found := state.ServiceByID(call.ServiceID)
	if !found {
		return ServiceCallReceipt{}, fmt.Errorf("aetracore service %s is not registered", call.ServiceID)
	}
	if descriptor.ServiceType != ServiceTypeOnChain {
		return ServiceCallReceipt{}, errors.New("aetracore on-chain service receipt requires on-chain service")
	}
	if err := result.Validate(); err != nil {
		return ServiceCallReceipt{}, err
	}
	status := ServiceCallStatusExecuted
	paymentStatus := ServicePaymentStatusSettled
	errorCode := ""
	if !result.Success {
		status = ServiceCallStatusFailed
		paymentStatus = ServicePaymentStatusRefunded
		errorCode = string(FailureReasonExecutionFailed)
	}
	return NewServiceCallReceipt(call, ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		status,
		ResponseHash:	result.ResultHash,
		PaymentStatus:	paymentStatus,
		GasUsed:	gasUsed,
		ExecutedHeight:	ctx.Height,
		AnchoredHeight:	ctx.Height,
		ErrorCode:	errorCode,
	})
}

func NewServiceStateProofQuery(descriptor ServiceDescriptor, height uint64, keyHash string, proofRoot string) (ServiceStateProofQuery, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceStateProofQuery{}, err
	}
	if descriptor.ServiceType != ServiceTypeOnChain {
		return ServiceStateProofQuery{}, errors.New("aetracore service state proof query requires on-chain service")
	}
	query := ServiceStateProofQuery{
		ServiceID:	descriptor.ServiceID,
		Height:		height,
		StateRootType:	descriptor.Storage.StateRootType,
		KeyHash:	keyHash,
		ProofRoot:	proofRoot,
	}
	if err := query.Validate(); err != nil {
		return ServiceStateProofQuery{}, err
	}
	query.QueryHash = ComputeServiceStateProofQueryHash(query)
	return query, query.Validate()
}

func CanonicalOnChainServiceWrapper(wrapper OnChainServiceWrapper) OnChainServiceWrapper {
	if wrapper.PaymentDenom == "" {
		wrapper.PaymentDenom = NativeFeePolicyID
	}
	if wrapper.PaymentAmount == "" {
		wrapper.PaymentAmount = "0"
	}
	if wrapper.ReceiptPolicy == "" {
		wrapper.ReceiptPolicy = ServiceReceiptCommitted
	}
	if wrapper.InterfaceName == "" {
		wrapper.InterfaceName = wrapper.InterfaceID
	}
	wrapper.Methods = cloneOnChainMethods(wrapper.Methods)
	sort.SliceStable(wrapper.Methods, func(i, j int) bool {
		return wrapper.Methods[i].MethodID < wrapper.Methods[j].MethodID
	})
	return wrapper
}

func (wrapper OnChainServiceWrapper) Validate() error {
	if err := validatePolicyID("aetracore on-chain service id", wrapper.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore on-chain service owner", wrapper.Owner); err != nil {
		return err
	}
	if err := ValidateZoneID(wrapper.ZoneID); err != nil {
		return err
	}
	if wrapper.ModuleName == "" && wrapper.ContractAddress == "" {
		return errors.New("aetracore on-chain service requires module name or contract address")
	}
	if wrapper.ModuleName != "" && wrapper.ContractAddress != "" {
		return errors.New("aetracore on-chain service must not define both module name and contract address")
	}
	if wrapper.ModuleName != "" {
		if err := validateModuleName(wrapper.ModuleName); err != nil {
			return err
		}
	}
	if wrapper.ContractAddress != "" {
		if err := validatePolicyID("aetracore on-chain service contract address", wrapper.ContractAddress); err != nil {
			return err
		}
	}
	if err := validatePolicyID("aetracore on-chain service interface id", wrapper.InterfaceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore on-chain service endpoint key", wrapper.EndpointKey); err != nil {
		return err
	}
	if wrapper.Version == 0 {
		return errors.New("aetracore on-chain service version must be positive")
	}
	if err := ValidateHash("aetracore on-chain service availability hash", wrapper.AvailabilityHash); err != nil {
		return err
	}
	if wrapper.StateRootType == "" {
		return errors.New("aetracore on-chain service requires state root type")
	}
	if err := validateToken("aetracore on-chain service state root type", string(wrapper.StateRootType), MaxScopeLength); err != nil {
		return err
	}
	if !IsServiceReceiptPolicy(wrapper.ReceiptPolicy) {
		return fmt.Errorf("unknown aetracore on-chain service receipt policy %q", wrapper.ReceiptPolicy)
	}
	if err := validatePolicyID("aetracore on-chain service payment denom", wrapper.PaymentDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore on-chain service payment amount", wrapper.PaymentAmount); err != nil {
		return err
	}
	if wrapper.CreatedHeight == 0 {
		return errors.New("aetracore on-chain service created height must be positive")
	}
	if wrapper.UpdatedHeight < wrapper.CreatedHeight {
		return errors.New("aetracore on-chain service updated height must not precede created height")
	}
	if wrapper.ExpiryHeight != 0 && wrapper.ExpiryHeight <= wrapper.UpdatedHeight {
		return errors.New("aetracore on-chain service expiry height must exceed updated height")
	}
	if err := validateOptionalHash("aetracore on-chain service metadata hash", wrapper.MetadataHash); err != nil {
		return err
	}
	return validateOnChainMethods(wrapper.Methods)
}

func (method OnChainServiceMethod) Validate() error {
	if err := validatePolicyID("aetracore on-chain method id", method.MethodID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore on-chain method name", method.Name); err != nil {
		return err
	}
	if err := ValidateHash("aetracore on-chain method input schema hash", method.InputSchemaHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore on-chain method output schema hash", method.OutputSchemaHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore on-chain method gas model", method.GasModel); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore on-chain method payment model", method.RequiredPaymentModel); err != nil {
		return err
	}
	if !IsServiceFailureBehavior(method.FailurePolicy) {
		return fmt.Errorf("unknown aetracore on-chain method failure policy %q", method.FailurePolicy)
	}
	if method.TimeoutHeightDelta == 0 {
		return errors.New("aetracore on-chain method timeout must be positive")
	}
	return nil
}

func (method OnChainServiceMethod) ServiceMethodDescriptor() ServiceMethodDescriptor {
	return ServiceMethodDescriptor{
		MethodID:		method.MethodID,
		Name:			method.Name,
		InputSchemaHash:	method.InputSchemaHash,
		OutputSchemaHash:	method.OutputSchemaHash,
		ExecutionType:		ServiceMethodSync,
		RequiredPaymentModel:	method.RequiredPaymentModel,
		GasModel:		method.GasModel,
		VerificationModel:	ServiceVerificationConsensusReceipt,
		TimeoutHeightDelta:	method.TimeoutHeightDelta,
		IdempotencyRequired:	method.IdempotencyRequired,
		CallbackSupported:	method.CallbackSupported,
		FailureBehavior:	method.FailurePolicy,
	}
}

func (query ServiceStateProofQuery) Validate() error {
	if err := validatePolicyID("aetracore service state proof service id", query.ServiceID); err != nil {
		return err
	}
	if query.Height == 0 {
		return errors.New("aetracore service state proof height must be positive")
	}
	if query.StateRootType == "" {
		return errors.New("aetracore service state proof requires state root type")
	}
	if err := validateToken("aetracore service state proof root type", string(query.StateRootType), MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service state proof key hash", query.KeyHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service state proof root", query.ProofRoot); err != nil {
		return err
	}
	if query.QueryHash != "" {
		if err := ValidateHash("aetracore service state proof query hash", query.QueryHash); err != nil {
			return err
		}
		if expected := ComputeServiceStateProofQueryHash(query); query.QueryHash != expected {
			return fmt.Errorf("aetracore service state proof query hash mismatch: expected %s", expected)
		}
	}
	return nil
}

func ComputeServiceStateProofQueryHash(query ServiceStateProofQuery) string {
	return hashParts(
		"aetra-aek-service-state-proof-query-v1",
		query.ServiceID,
		fmt.Sprint(query.Height),
		string(query.StateRootType),
		query.KeyHash,
		query.ProofRoot,
	)
}

func validateOnChainMethods(methods []OnChainServiceMethod) error {
	if len(methods) == 0 {
		return errors.New("aetracore on-chain service requires method set")
	}
	var previous string
	seen := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seen[method.MethodID]; found {
			return fmt.Errorf("duplicate aetracore on-chain method %s", method.MethodID)
		}
		seen[method.MethodID] = struct{}{}
		if i > 0 && previous >= method.MethodID {
			return errors.New("aetracore on-chain methods must be sorted canonically")
		}
		previous = method.MethodID
	}
	return nil
}

func cloneOnChainMethods(methods []OnChainServiceMethod) []OnChainServiceMethod {
	out := make([]OnChainServiceMethod, len(methods))
	copy(out, methods)
	return out
}

func aggregateOnChainFailurePolicy(methods []OnChainServiceMethod) ServiceFailureBehavior {
	for _, method := range methods {
		if method.FailurePolicy == ServiceFailureFallbackOnChain {
			return ServiceFailureFallbackOnChain
		}
		if method.FailurePolicy == ServiceFailureRevert {
			return ServiceFailureRevert
		}
	}
	if len(methods) == 0 {
		return ServiceFailureRevert
	}
	return methods[0].FailurePolicy
}
