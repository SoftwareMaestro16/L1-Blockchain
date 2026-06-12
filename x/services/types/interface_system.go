package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceInterfaceMethodSchema struct {
	MethodID		string
	Name			string
	InputSchemaHash		string
	OutputSchemaHash	string
	ExecutionType		coretypes.ServiceMethodExecutionType
	GasModel		string
	VerificationModel	coretypes.ServiceVerificationModel
	TimeoutPolicy		ServiceMethodTimeoutPolicy
	IdempotencyRequired	bool
	CallbackSupported	bool
	PaymentRequirements	string
	MethodHash		string
}

type ServiceMethodTimeoutPolicy struct {
	TimeoutHeightDelta	uint64
	FailureBehavior		coretypes.ServiceFailureBehavior
	PolicyHash		string
}

type FormalServiceInterface struct {
	InterfaceHash		string
	InterfaceName		string
	Version			uint64
	Methods			[]ServiceInterfaceMethodSchema
	Events			[]string
	Errors			[]string
	AuthModel		string
	PaymentRequirements	string
	SchemaEncoding		string
	MetadataHash		string
	CreatedHeight		uint64
	DefinitionHash		string
}

type ServiceInterfaceCallPreparation struct {
	ServiceID		string
	InterfaceHash		string
	MethodID		string
	MethodName		string
	InputSchemaHash		string
	OutputSchemaHash	string
	ExecutionType		coretypes.ServiceMethodExecutionType
	GasModel		string
	VerificationModel	coretypes.ServiceVerificationModel
	AuthModel		string
	PaymentRequirements	string
	SchemaEncoding		string
	PayloadHash		string
	Caller			string
	Nonce			uint64
	PreparedHeight		uint64
	EventStream		bool
	PreparationHash		string
}

func NewFormalServiceInterface(iface ServiceInterface) (FormalServiceInterface, error) {
	iface = coretypes.CanonicalServiceInterfaceDescriptor(iface)
	if err := iface.Validate(); err != nil {
		return FormalServiceInterface{}, err
	}
	methods := make([]ServiceInterfaceMethodSchema, 0, len(iface.Methods))
	for _, method := range iface.Methods {
		methodSchema := ServiceInterfaceMethodSchema{
			MethodID:		method.MethodID,
			Name:			method.Name,
			InputSchemaHash:	method.InputSchemaHash,
			OutputSchemaHash:	method.OutputSchemaHash,
			ExecutionType:		method.ExecutionType,
			GasModel:		method.GasModel,
			VerificationModel:	method.VerificationModel,
			TimeoutPolicy: ServiceMethodTimeoutPolicy{
				TimeoutHeightDelta:	method.TimeoutHeightDelta,
				FailureBehavior:	method.FailureBehavior,
			},
			IdempotencyRequired:	method.IdempotencyRequired,
			CallbackSupported:	method.CallbackSupported,
			PaymentRequirements:	method.RequiredPaymentModel,
		}
		methodSchema.TimeoutPolicy.PolicyHash = ComputeServiceMethodTimeoutPolicyHash(methodSchema.TimeoutPolicy)
		methodSchema.MethodHash = ComputeServiceInterfaceMethodSchemaHash(methodSchema)
		methods = append(methods, methodSchema)
	}
	sortServiceInterfaceMethodSchemas(methods)
	definition := FormalServiceInterface{
		InterfaceHash:		iface.InterfaceHash,
		InterfaceName:		iface.InterfaceName,
		Version:		iface.Version,
		Methods:		methods,
		Events:			append([]string(nil), iface.Events...),
		Errors:			append([]string(nil), iface.Errors...),
		AuthModel:		iface.AuthModel,
		PaymentRequirements:	iface.PaymentModel,
		SchemaEncoding:		iface.SchemaEncoding,
		MetadataHash:		iface.MetadataHash,
		CreatedHeight:		iface.CreatedHeight,
	}
	sort.Strings(definition.Events)
	sort.Strings(definition.Errors)
	definition.DefinitionHash = ComputeFormalServiceInterfaceHash(definition)
	return definition, definition.Validate()
}

func (definition FormalServiceInterface) Validate() error {
	if err := coretypes.ValidateHash("services interface hash", definition.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface name", definition.InterfaceName); err != nil {
		return err
	}
	if definition.Version == 0 {
		return errors.New("services interface version must be positive")
	}
	if len(definition.Methods) == 0 {
		return errors.New("services interface requires methods")
	}
	if err := validateInterfaceMethods(definition.Methods); err != nil {
		return err
	}
	if err := validateSortedTokens("services interface event", definition.Events); err != nil {
		return err
	}
	if err := validateSortedTokens("services interface error", definition.Errors); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface auth model", definition.AuthModel); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface payment requirements", definition.PaymentRequirements); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface schema encoding", definition.SchemaEncoding); err != nil {
		return err
	}
	if !IsSupportedServiceSchemaEncoding(definition.SchemaEncoding) {
		return fmt.Errorf("services interface schema encoding %q is not supported", definition.SchemaEncoding)
	}
	if definition.MetadataHash != "" {
		if err := coretypes.ValidateHash("services interface metadata hash", definition.MetadataHash); err != nil {
			return err
		}
	}
	if definition.CreatedHeight == 0 {
		return errors.New("services interface created height must be positive")
	}
	if err := coretypes.ValidateHash("services interface definition hash", definition.DefinitionHash); err != nil {
		return err
	}
	if expected := ComputeFormalServiceInterfaceHash(definition); definition.DefinitionHash != expected {
		return fmt.Errorf("services interface definition hash mismatch: expected %s", expected)
	}
	return nil
}

func (method ServiceInterfaceMethodSchema) Validate() error {
	if err := validateInterfaceToken("services interface method id", method.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface method name", method.Name); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface input schema", method.InputSchemaHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface output schema", method.OutputSchemaHash); err != nil {
		return err
	}
	if !coretypes.IsServiceMethodExecutionType(method.ExecutionType) {
		return fmt.Errorf("services interface unknown execution type %q", method.ExecutionType)
	}
	if method.GasModel != "" {
		if err := validateInterfaceToken("services interface gas model", method.GasModel); err != nil {
			return err
		}
	}
	if !coretypes.IsServiceVerificationModel(method.VerificationModel) {
		return fmt.Errorf("services interface unknown verification model %q", method.VerificationModel)
	}
	if err := method.TimeoutPolicy.Validate(); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface method payment requirements", method.PaymentRequirements); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface method hash", method.MethodHash); err != nil {
		return err
	}
	if expected := ComputeServiceInterfaceMethodSchemaHash(method); method.MethodHash != expected {
		return fmt.Errorf("services interface method hash mismatch: expected %s", expected)
	}
	return nil
}

func (policy ServiceMethodTimeoutPolicy) Validate() error {
	if policy.TimeoutHeightDelta == 0 {
		return errors.New("services interface timeout policy requires positive timeout")
	}
	if !coretypes.IsServiceFailureBehavior(policy.FailureBehavior) {
		return fmt.Errorf("services interface timeout policy unknown failure behavior %q", policy.FailureBehavior)
	}
	if err := coretypes.ValidateHash("services interface timeout policy hash", policy.PolicyHash); err != nil {
		return err
	}
	if expected := ComputeServiceMethodTimeoutPolicyHash(policy); policy.PolicyHash != expected {
		return fmt.Errorf("services interface timeout policy hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateServiceInterfaceForDescriptor(descriptor ServiceDescriptor) error {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return err
	}
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	if err != nil {
		return err
	}
	if descriptor.Interface.InterfaceHash != coretypes.ComputeServiceInterfaceHash(descriptor.Interface) {
		return errors.New("services interface hash must commit to descriptor interface fields")
	}
	if definition.InterfaceHash != descriptor.Interface.InterfaceHash {
		return errors.New("services formal interface hash mismatch")
	}
	for _, method := range definition.Methods {
		switch descriptor.ServiceType {
		case coretypes.ServiceTypeOnChain:
			if method.GasModel == "" {
				return fmt.Errorf("services on-chain method %s requires gas model", method.MethodID)
			}
		case coretypes.ServiceTypeOffChain:
			if !methodHasResponseVerification(method.VerificationModel) {
				return fmt.Errorf("services off-chain method %s requires response verification model", method.MethodID)
			}
		case coretypes.ServiceTypeMixed:
			if !methodHasDisputeOrFallback(method, descriptor) {
				return fmt.Errorf("services mixed method %s requires dispute or fallback model", method.MethodID)
			}
		}
	}
	return nil
}

func ValidateServiceInterfaceVersionChange(previous ServiceInterface, next ServiceInterface) error {
	previous = coretypes.CanonicalServiceInterfaceDescriptor(previous)
	next = coretypes.CanonicalServiceInterfaceDescriptor(next)
	if err := previous.Validate(); err != nil {
		return err
	}
	if next.Version <= previous.Version {
		return errors.New("services interface version must increase")
	}
	if next.InterfaceHash == previous.InterfaceHash {
		return errors.New("services interface version change must create a new interface hash")
	}
	if err := next.Validate(); err != nil {
		return err
	}
	return nil
}

func PrepareServiceInterfaceCall(descriptor ServiceDescriptor, methodName, caller string, nonce uint64, payloadHash string, preparedHeight uint64) (ServiceInterfaceCallPreparation, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceInterfaceCallPreparation{}, err
	}
	definition, err := NewFormalServiceInterface(descriptor.Interface)
	if err != nil {
		return ServiceInterfaceCallPreparation{}, err
	}
	method, found := definition.MethodByName(methodName)
	if !found {
		return ServiceInterfaceCallPreparation{}, fmt.Errorf("services interface method %s not found", methodName)
	}
	if err := validateInterfaceToken("services interface caller", caller); err != nil {
		return ServiceInterfaceCallPreparation{}, err
	}
	if nonce == 0 {
		return ServiceInterfaceCallPreparation{}, errors.New("services interface call nonce must be positive")
	}
	if err := coretypes.ValidateHash("services interface payload hash", payloadHash); err != nil {
		return ServiceInterfaceCallPreparation{}, err
	}
	if preparedHeight == 0 {
		return ServiceInterfaceCallPreparation{}, errors.New("services interface call prepared height must be positive")
	}
	preparation := ServiceInterfaceCallPreparation{
		ServiceID:		descriptor.ServiceID,
		InterfaceHash:		definition.InterfaceHash,
		MethodID:		method.MethodID,
		MethodName:		method.Name,
		InputSchemaHash:	method.InputSchemaHash,
		OutputSchemaHash:	method.OutputSchemaHash,
		ExecutionType:		method.ExecutionType,
		GasModel:		method.GasModel,
		VerificationModel:	method.VerificationModel,
		AuthModel:		definition.AuthModel,
		PaymentRequirements:	method.PaymentRequirements,
		SchemaEncoding:		definition.SchemaEncoding,
		PayloadHash:		strings.ToLower(strings.TrimSpace(payloadHash)),
		Caller:			strings.TrimSpace(caller),
		Nonce:			nonce,
		PreparedHeight:		preparedHeight,
		EventStream:		method.ExecutionType == coretypes.ServiceMethodEvented,
	}
	preparation.PreparationHash = ComputeServiceInterfaceCallPreparationHash(preparation)
	return preparation, preparation.Validate()
}

func (definition FormalServiceInterface) MethodByName(methodName string) (ServiceInterfaceMethodSchema, bool) {
	for _, method := range definition.Methods {
		if method.Name == methodName || method.MethodID == methodName {
			return method, true
		}
	}
	return ServiceInterfaceMethodSchema{}, false
}

func (definition FormalServiceInterface) SupportsExecutionType(executionType coretypes.ServiceMethodExecutionType) bool {
	for _, method := range definition.Methods {
		if method.ExecutionType == executionType {
			return true
		}
	}
	return false
}

func (preparation ServiceInterfaceCallPreparation) Validate() error {
	if err := validateInterfaceToken("services interface call service id", preparation.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface call interface hash", preparation.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface call method id", preparation.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services interface call method name", preparation.MethodName); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface call input schema", preparation.InputSchemaHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services interface call output schema", preparation.OutputSchemaHash); err != nil {
		return err
	}
	if !coretypes.IsServiceMethodExecutionType(preparation.ExecutionType) {
		return fmt.Errorf("services interface call unknown execution type %q", preparation.ExecutionType)
	}
	if preparation.GasModel != "" {
		if err := validateInterfaceToken("services interface call gas model", preparation.GasModel); err != nil {
			return err
		}
	}
	if !coretypes.IsServiceVerificationModel(preparation.VerificationModel) {
		return fmt.Errorf("services interface call unknown verification model %q", preparation.VerificationModel)
	}
	if preparation.EventStream != (preparation.ExecutionType == coretypes.ServiceMethodEvented) {
		return errors.New("services interface call event stream flag mismatch")
	}
	if err := coretypes.ValidateHash("services interface call payload hash", preparation.PayloadHash); err != nil {
		return err
	}
	if preparation.Nonce == 0 || preparation.PreparedHeight == 0 {
		return errors.New("services interface call nonce and prepared height must be positive")
	}
	if err := coretypes.ValidateHash("services interface call preparation hash", preparation.PreparationHash); err != nil {
		return err
	}
	if expected := ComputeServiceInterfaceCallPreparationHash(preparation); preparation.PreparationHash != expected {
		return fmt.Errorf("services interface call preparation hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceInterfaceMethodSchemaHash(method ServiceInterfaceMethodSchema) string {
	return servicesHashParts(
		"aetra-services-interface-method-v1",
		method.MethodID,
		method.Name,
		method.InputSchemaHash,
		method.OutputSchemaHash,
		string(method.ExecutionType),
		method.GasModel,
		string(method.VerificationModel),
		method.TimeoutPolicy.PolicyHash,
		fmt.Sprint(method.IdempotencyRequired),
		fmt.Sprint(method.CallbackSupported),
		method.PaymentRequirements,
	)
}

func ComputeServiceMethodTimeoutPolicyHash(policy ServiceMethodTimeoutPolicy) string {
	return servicesHashParts(
		"aetra-services-interface-timeout-policy-v1",
		fmt.Sprint(policy.TimeoutHeightDelta),
		string(policy.FailureBehavior),
	)
}

func ComputeFormalServiceInterfaceHash(definition FormalServiceInterface) string {
	methods := append([]ServiceInterfaceMethodSchema(nil), definition.Methods...)
	sortServiceInterfaceMethodSchemas(methods)
	parts := []string{
		"aetra-services-interface-definition-v1",
		definition.InterfaceHash,
		definition.InterfaceName,
		fmt.Sprint(definition.Version),
		definition.AuthModel,
		definition.PaymentRequirements,
		definition.SchemaEncoding,
		definition.MetadataHash,
		fmt.Sprint(definition.CreatedHeight),
	}
	for _, method := range methods {
		parts = append(parts, method.MethodHash)
	}
	parts = appendStringParts(parts, "events", definition.Events)
	parts = appendStringParts(parts, "errors", definition.Errors)
	return servicesHashParts(parts...)
}

func ComputeServiceInterfaceCallPreparationHash(preparation ServiceInterfaceCallPreparation) string {
	return servicesHashParts(
		"aetra-services-interface-call-preparation-v1",
		preparation.ServiceID,
		preparation.InterfaceHash,
		preparation.MethodID,
		preparation.MethodName,
		preparation.InputSchemaHash,
		preparation.OutputSchemaHash,
		string(preparation.ExecutionType),
		preparation.GasModel,
		string(preparation.VerificationModel),
		preparation.AuthModel,
		preparation.PaymentRequirements,
		preparation.SchemaEncoding,
		preparation.PayloadHash,
		preparation.Caller,
		fmt.Sprint(preparation.Nonce),
		fmt.Sprint(preparation.PreparedHeight),
		fmt.Sprint(preparation.EventStream),
	)
}

func validateInterfaceMethods(methods []ServiceInterfaceMethodSchema) error {
	previous := ""
	names := map[string]struct{}{}
	for _, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := names[method.Name]; found {
			return fmt.Errorf("services interface method name %s is duplicated", method.Name)
		}
		names[method.Name] = struct{}{}
		if previous != "" && previous >= method.MethodID {
			return errors.New("services interface methods must be sorted canonically")
		}
		previous = method.MethodID
	}
	return nil
}

func IsSupportedServiceSchemaEncoding(encoding string) bool {
	switch encoding {
	case "json-schema-v1", "protobuf-v3", "openapi-v3":
		return true
	default:
		return false
	}
}

func methodHasResponseVerification(model coretypes.ServiceVerificationModel) bool {
	switch model {
	case coretypes.ServiceVerificationSignedResult, coretypes.ServiceVerificationProofAnchored,
		coretypes.ServiceVerificationChallengeWindow, coretypes.ServiceVerificationEconomicCollateral:
		return true
	default:
		return false
	}
}

func methodHasDisputeOrFallback(method ServiceInterfaceMethodSchema, descriptor ServiceDescriptor) bool {
	return method.VerificationModel == coretypes.ServiceVerificationChallengeWindow ||
		method.TimeoutPolicy.FailureBehavior == coretypes.ServiceFailureChallenge ||
		method.TimeoutPolicy.FailureBehavior == coretypes.ServiceFailureFallbackOnChain ||
		descriptor.Execution.ChallengeWindow != 0 ||
		descriptor.Verification.ChallengeWindow != 0 ||
		descriptor.Verification.FallbackServiceID != ""
}

func validateSortedTokens(fieldName string, values []string) error {
	previous := ""
	for _, value := range values {
		if err := validateInterfaceToken(fieldName, value); err != nil {
			return err
		}
		if previous != "" && previous >= value {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func validateInterfaceToken(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > 128 {
		return fmt.Errorf("%s must be <= 128 bytes", fieldName)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func sortServiceInterfaceMethodSchemas(methods []ServiceInterfaceMethodSchema) {
	sort.SliceStable(methods, func(i, j int) bool { return methods[i].MethodID < methods[j].MethodID })
}

func appendStringParts(parts []string, label string, values []string) []string {
	ordered := append([]string(nil), values...)
	sort.Strings(ordered)
	parts = append(parts, label, fmt.Sprint(len(ordered)))
	return append(parts, ordered...)
}
