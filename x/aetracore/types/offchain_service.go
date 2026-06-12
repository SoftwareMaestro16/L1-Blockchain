package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type OffChainEndpointType string
type OffChainRequestSigningPolicy string
type OffChainResponseSigningPolicy string
type OffChainProofAnchorPolicy string
type OffChainAvailabilityPolicy string

const (
	OffChainEndpointRPC		OffChainEndpointType	= "RPC_ENDPOINT"
	OffChainEndpointOverlay		OffChainEndpointType	= "OVERLAY"
	OffChainEndpointProviderMesh	OffChainEndpointType	= "PROVIDER_MESH"
	OffChainEndpointServiceNetwork	OffChainEndpointType	= "SERVICE_NETWORK"

	OffChainRequestCallerSigned	OffChainRequestSigningPolicy	= "CALLER_SIGNED"

	OffChainResponseProviderSigned	OffChainResponseSigningPolicy	= "PROVIDER_SIGNED"

	OffChainProofAnchorNone		OffChainProofAnchorPolicy	= "NONE"
	OffChainProofAnchorOptional	OffChainProofAnchorPolicy	= "OPTIONAL"
	OffChainProofAnchorRequired	OffChainProofAnchorPolicy	= "REQUIRED"

	OffChainAvailabilitySignedAdvertisement	OffChainAvailabilityPolicy	= "SIGNED_ADVERTISEMENT"
	OffChainAvailabilityRenewableEndpoint	OffChainAvailabilityPolicy	= "RENEWABLE_ENDPOINT"
)

type OffChainServiceDefinition struct {
	ServiceID		string
	Owner			string
	ZoneID			ZoneID
	Endpoint		string
	EndpointType		OffChainEndpointType
	ProviderKey		string
	RequestSigningPolicy	OffChainRequestSigningPolicy
	ResponseSigningPolicy	OffChainResponseSigningPolicy
	ProofAnchorPolicy	OffChainProofAnchorPolicy
	AvailabilityPolicy	OffChainAvailabilityPolicy
	ResultExpiry		uint64
	InterfaceID		string
	InterfaceName		string
	EndpointKey		string
	Version			uint64
	AvailabilityHash	string
	ProviderRoot		string
	PaymentDenom		string
	PaymentAmount		string
	Methods			[]OffChainServiceMethod
	MetadataHash		string
	CreatedHeight		uint64
	UpdatedHeight		uint64
	ExpiryHeight		uint64
}

type OffChainServiceMethod struct {
	MethodID		string
	Name			string
	InputSchemaHash		string
	OutputSchemaHash	string
	RequiredPaymentModel	string
	VerificationModel	ServiceVerificationModel
	TimeoutHeightDelta	uint64
	IdempotencyRequired	bool
	CallbackSupported	bool
	FailurePolicy		ServiceFailureBehavior
}

type OffChainEndpointMetadata struct {
	Endpoint	string
	EndpointType	OffChainEndpointType
	ProviderKey	string
	ExpiryHeight	uint64
	MetadataHash	string
}

type OffChainServiceDescriptorSchema struct {
	Descriptor		ServiceDescriptor
	Endpoint		OffChainEndpointMetadata
	RequestSigningPolicy	OffChainRequestSigningPolicy
	ResponseSigningPolicy	OffChainResponseSigningPolicy
	ProofAnchorPolicy	OffChainProofAnchorPolicy
	AvailabilityPolicy	OffChainAvailabilityPolicy
	ResultExpiry		uint64
	SchemaHash		string
}

type OffChainSignedAdvertisement struct {
	ServiceID		string
	Owner			string
	InterfaceHash		string
	Endpoint		string
	EndpointType		OffChainEndpointType
	ProviderKey		string
	ExpiryHeight		uint64
	AdvertisementHash	string
	Signer			string
	SignatureHash		string
}

type OffChainServiceRequest struct {
	CallID		string
	ServiceID	string
	MethodID	string
	Caller		string
	Nonce		uint64
	IdempotencyKey	string
	PayloadHash	string
	ProviderKey	string
	DeadlineHeight	uint64
	RequestHash	string
}

type OffChainSignedRequest struct {
	Request		OffChainServiceRequest
	Signer		string
	SignatureHash	string
}

type OffChainServiceResponse struct {
	CallID			string
	ServiceID		string
	MethodID		string
	RequestHash		string
	ResponseHash		string
	ProviderKey		string
	Height			uint64
	ResultExpiryHeight	uint64
	SettlementUse		bool
	ChallengeUse		bool
}

type OffChainSignedResponse struct {
	Response	OffChainServiceResponse
	Signer		string
	SignatureHash	string
}

type OffChainReceiptAnchorMessage struct {
	ServiceID	string
	CallID		string
	RequestHash	string
	ResponseHash	string
	ProviderKey	string
	Height		uint64
	ProofHash	string
	AnchorHash	string
}

type OffChainEndpointRenewal struct {
	ServiceID		string
	Endpoint		string
	EndpointType		OffChainEndpointType
	ProviderKey		string
	RenewedAtHeight		uint64
	ExpiryHeight		uint64
	AdvertisementHash	string
	Signer			string
	SignatureHash		string
}

func BuildOffChainServiceDescriptor(definition OffChainServiceDefinition) (OffChainServiceDescriptorSchema, error) {
	definition = CanonicalOffChainServiceDefinition(definition)
	if err := definition.Validate(); err != nil {
		return OffChainServiceDescriptorSchema{}, err
	}
	methods := make([]ServiceMethodDescriptor, len(definition.Methods))
	for i, method := range definition.Methods {
		methods[i] = method.ServiceMethodDescriptor()
	}
	interfaceDescriptor := ServiceInterfaceDescriptor{
		InterfaceID:	definition.InterfaceID,
		InterfaceName:	definition.InterfaceName,
		Version:	definition.Version,
		SchemaEncoding:	DefaultOnChainSchemaEncoding,
		Methods:	methods,
		Events:		[]string{"offchain.receipt_anchored"},
		Errors:		[]string{"offchain.response_rejected"},
		AuthModel:	DefaultOnChainAuthModel,
		PaymentModel:	definition.Methods[0].RequiredPaymentModel,
		MetadataHash:	definition.MetadataHash,
		CreatedHeight:	definition.CreatedHeight,
	}
	interfaceDescriptor = CanonicalServiceInterfaceDescriptor(interfaceDescriptor)
	interfaceDescriptor.InterfaceHash = ComputeServiceInterfaceHash(interfaceDescriptor)

	location := ServiceLocationExternal
	providerPoolID := ""
	if definition.EndpointType == OffChainEndpointProviderMesh {
		location = ServiceLocationProviderPool
		providerPoolID = definition.ProviderKey
	}
	descriptor := ServiceDescriptor{
		ServiceID:		definition.ServiceID,
		Owner:			definition.Owner,
		ServiceType:		ServiceTypeOffChain,
		ZoneID:			definition.ZoneID,
		InterfaceID:		definition.InterfaceID,
		EndpointKey:		definition.EndpointKey,
		Version:		definition.Version,
		AvailabilityHash:	definition.AvailabilityHash,
		Enabled:		true,
		Status:			ServiceStatusActive,
		ExpiryHeight:		definition.ExpiryHeight,
		CreatedHeight:		definition.CreatedHeight,
		UpdatedHeight:		definition.UpdatedHeight,
		Interface:		interfaceDescriptor,
		Execution: ServiceExecutionDescriptor{
			Location:		location,
			Target:			definition.EndpointKey,
			Endpoint:		definition.Endpoint,
			ProviderPoolID:		providerPoolID,
			Mode:			ExecutionModeAsync,
			Deterministic:		false,
			ReceiptPolicy:		ServiceReceiptCommitted,
			FailureBehavior:	aggregateOffChainFailurePolicy(definition.Methods),
			ResultExpiry:		definition.ResultExpiry,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:		definition.ServiceID,
			ProviderRoot:		definition.ProviderRoot,
			MetadataHash:		definition.MetadataHash,
			SignaturePolicy:	string(definition.AvailabilityPolicy),
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode:	ServicePaymentPrepaid,
			Denom:		definition.PaymentDenom,
			Amount:		definition.PaymentAmount,
			PricingUnit:	ServicePricingPerCall,
			ExpiryHeight:	definition.ExpiryHeight,
		},
		Storage: ServiceStorageDescriptor{
			Model:		ServiceStorageDistributedOffChain,
			ProofRequired:	definition.ProofAnchorPolicy == OffChainProofAnchorRequired,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:			ServiceTrustFullyTrusted,
			Model:				offChainVerificationModel(definition.ProofAnchorPolicy),
			RequestSigningRequired:		true,
			ResponseSigningRequired:	true,
			ProofFormat:			offChainProofFormat(definition.ProofAnchorPolicy),
		},
	}
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return OffChainServiceDescriptorSchema{}, err
	}
	schema := OffChainServiceDescriptorSchema{
		Descriptor:	descriptor,
		Endpoint: OffChainEndpointMetadata{
			Endpoint:	definition.Endpoint,
			EndpointType:	definition.EndpointType,
			ProviderKey:	definition.ProviderKey,
			ExpiryHeight:	definition.ExpiryHeight,
			MetadataHash:	definition.MetadataHash,
		},
		RequestSigningPolicy:	definition.RequestSigningPolicy,
		ResponseSigningPolicy:	definition.ResponseSigningPolicy,
		ProofAnchorPolicy:	definition.ProofAnchorPolicy,
		AvailabilityPolicy:	definition.AvailabilityPolicy,
		ResultExpiry:		definition.ResultExpiry,
	}
	schema.SchemaHash = ComputeOffChainServiceDescriptorSchemaHash(schema)
	return schema, schema.Validate()
}

func CanonicalOffChainServiceDefinition(definition OffChainServiceDefinition) OffChainServiceDefinition {
	if definition.PaymentDenom == "" {
		definition.PaymentDenom = NativeFeePolicyID
	}
	if definition.PaymentAmount == "" {
		definition.PaymentAmount = "0"
	}
	if definition.RequestSigningPolicy == "" {
		definition.RequestSigningPolicy = OffChainRequestCallerSigned
	}
	if definition.ResponseSigningPolicy == "" {
		definition.ResponseSigningPolicy = OffChainResponseProviderSigned
	}
	if definition.ProofAnchorPolicy == "" {
		definition.ProofAnchorPolicy = OffChainProofAnchorOptional
	}
	if definition.AvailabilityPolicy == "" {
		definition.AvailabilityPolicy = OffChainAvailabilitySignedAdvertisement
	}
	if definition.InterfaceName == "" {
		definition.InterfaceName = definition.InterfaceID
	}
	definition.Methods = cloneOffChainMethods(definition.Methods)
	sort.SliceStable(definition.Methods, func(i, j int) bool {
		return definition.Methods[i].MethodID < definition.Methods[j].MethodID
	})
	return definition
}

func (definition OffChainServiceDefinition) Validate() error {
	if err := validatePolicyID("aetracore off-chain service id", definition.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore off-chain service owner", definition.Owner); err != nil {
		return err
	}
	if err := ValidateZoneID(definition.ZoneID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain service endpoint", definition.Endpoint); err != nil {
		return err
	}
	if !IsOffChainEndpointType(definition.EndpointType) {
		return fmt.Errorf("unknown aetracore off-chain endpoint type %q", definition.EndpointType)
	}
	if err := validatePolicyID("aetracore off-chain provider key", definition.ProviderKey); err != nil {
		return err
	}
	if !IsOffChainRequestSigningPolicy(definition.RequestSigningPolicy) {
		return fmt.Errorf("unknown aetracore off-chain request signing policy %q", definition.RequestSigningPolicy)
	}
	if !IsOffChainResponseSigningPolicy(definition.ResponseSigningPolicy) {
		return fmt.Errorf("unknown aetracore off-chain response signing policy %q", definition.ResponseSigningPolicy)
	}
	if !IsOffChainProofAnchorPolicy(definition.ProofAnchorPolicy) {
		return fmt.Errorf("unknown aetracore off-chain proof anchor policy %q", definition.ProofAnchorPolicy)
	}
	if !IsOffChainAvailabilityPolicy(definition.AvailabilityPolicy) {
		return fmt.Errorf("unknown aetracore off-chain availability policy %q", definition.AvailabilityPolicy)
	}
	if definition.ResultExpiry == 0 {
		return errors.New("aetracore off-chain service result expiry must be positive")
	}
	if err := validatePolicyID("aetracore off-chain interface id", definition.InterfaceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain endpoint key", definition.EndpointKey); err != nil {
		return err
	}
	if definition.Version == 0 {
		return errors.New("aetracore off-chain service version must be positive")
	}
	if err := ValidateHash("aetracore off-chain availability hash", definition.AvailabilityHash); err != nil {
		return err
	}
	if err := validateOptionalHash("aetracore off-chain provider root", definition.ProviderRoot); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain payment denom", definition.PaymentDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore off-chain payment amount", definition.PaymentAmount); err != nil {
		return err
	}
	if definition.CreatedHeight == 0 {
		return errors.New("aetracore off-chain service created height must be positive")
	}
	if definition.UpdatedHeight < definition.CreatedHeight {
		return errors.New("aetracore off-chain service updated height must not precede created height")
	}
	if definition.ExpiryHeight != 0 && definition.ExpiryHeight <= definition.UpdatedHeight {
		return errors.New("aetracore off-chain service expiry height must exceed updated height")
	}
	if err := validateOptionalHash("aetracore off-chain metadata hash", definition.MetadataHash); err != nil {
		return err
	}
	return validateOffChainMethods(definition.Methods)
}

func (method OffChainServiceMethod) Validate() error {
	if err := validatePolicyID("aetracore off-chain method id", method.MethodID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain method name", method.Name); err != nil {
		return err
	}
	if err := ValidateHash("aetracore off-chain method input schema hash", method.InputSchemaHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore off-chain method output schema hash", method.OutputSchemaHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain method payment model", method.RequiredPaymentModel); err != nil {
		return err
	}
	if !IsServiceVerificationModel(method.VerificationModel) {
		return fmt.Errorf("unknown aetracore off-chain method verification model %q", method.VerificationModel)
	}
	if method.VerificationModel == ServiceVerificationConsensusReceipt {
		return errors.New("aetracore off-chain method cannot require consensus receipt verification")
	}
	if method.TimeoutHeightDelta == 0 {
		return errors.New("aetracore off-chain method timeout must be positive")
	}
	if !IsServiceFailureBehavior(method.FailurePolicy) {
		return fmt.Errorf("unknown aetracore off-chain method failure policy %q", method.FailurePolicy)
	}
	return nil
}

func (method OffChainServiceMethod) ServiceMethodDescriptor() ServiceMethodDescriptor {
	return ServiceMethodDescriptor{
		MethodID:		method.MethodID,
		Name:			method.Name,
		InputSchemaHash:	method.InputSchemaHash,
		OutputSchemaHash:	method.OutputSchemaHash,
		ExecutionType:		ServiceMethodAsync,
		RequiredPaymentModel:	method.RequiredPaymentModel,
		VerificationModel:	method.VerificationModel,
		TimeoutHeightDelta:	method.TimeoutHeightDelta,
		IdempotencyRequired:	method.IdempotencyRequired,
		CallbackSupported:	method.CallbackSupported,
		FailureBehavior:	method.FailurePolicy,
	}
}

func (schema OffChainServiceDescriptorSchema) Validate() error {
	if err := schema.Descriptor.Validate(); err != nil {
		return err
	}
	if schema.Descriptor.ServiceType != ServiceTypeOffChain {
		return errors.New("aetracore off-chain descriptor schema requires off-chain service descriptor")
	}
	if schema.Descriptor.Interface.InterfaceHash == "" {
		return errors.New("aetracore off-chain descriptor schema requires interface hash")
	}
	if err := schema.Endpoint.Validate(); err != nil {
		return err
	}
	if schema.Endpoint.Endpoint != schema.Descriptor.Execution.Endpoint {
		return errors.New("aetracore off-chain endpoint metadata mismatch")
	}
	if !IsOffChainRequestSigningPolicy(schema.RequestSigningPolicy) {
		return fmt.Errorf("unknown aetracore off-chain request signing policy %q", schema.RequestSigningPolicy)
	}
	if !IsOffChainResponseSigningPolicy(schema.ResponseSigningPolicy) {
		return fmt.Errorf("unknown aetracore off-chain response signing policy %q", schema.ResponseSigningPolicy)
	}
	if !IsOffChainProofAnchorPolicy(schema.ProofAnchorPolicy) {
		return fmt.Errorf("unknown aetracore off-chain proof anchor policy %q", schema.ProofAnchorPolicy)
	}
	if !IsOffChainAvailabilityPolicy(schema.AvailabilityPolicy) {
		return fmt.Errorf("unknown aetracore off-chain availability policy %q", schema.AvailabilityPolicy)
	}
	if schema.ResultExpiry == 0 {
		return errors.New("aetracore off-chain descriptor schema requires result expiry")
	}
	if err := ValidateHash("aetracore off-chain descriptor schema hash", schema.SchemaHash); err != nil {
		return err
	}
	if expected := ComputeOffChainServiceDescriptorSchemaHash(schema); schema.SchemaHash != expected {
		return fmt.Errorf("aetracore off-chain descriptor schema hash mismatch: expected %s", expected)
	}
	return nil
}

func (metadata OffChainEndpointMetadata) Validate() error {
	if err := validatePolicyID("aetracore off-chain endpoint", metadata.Endpoint); err != nil {
		return err
	}
	if !IsOffChainEndpointType(metadata.EndpointType) {
		return fmt.Errorf("unknown aetracore off-chain endpoint type %q", metadata.EndpointType)
	}
	if err := validatePolicyID("aetracore off-chain provider key", metadata.ProviderKey); err != nil {
		return err
	}
	if metadata.ExpiryHeight == 0 {
		return errors.New("aetracore off-chain endpoint expiry must be positive")
	}
	return validateOptionalHash("aetracore off-chain endpoint metadata hash", metadata.MetadataHash)
}

func NewOffChainSignedAdvertisement(schema OffChainServiceDescriptorSchema, signer string) (OffChainSignedAdvertisement, error) {
	if err := schema.Validate(); err != nil {
		return OffChainSignedAdvertisement{}, err
	}
	ad := OffChainSignedAdvertisement{
		ServiceID:	schema.Descriptor.ServiceID,
		Owner:		schema.Descriptor.Owner,
		InterfaceHash:	schema.Descriptor.Interface.InterfaceHash,
		Endpoint:	schema.Endpoint.Endpoint,
		EndpointType:	schema.Endpoint.EndpointType,
		ProviderKey:	schema.Endpoint.ProviderKey,
		ExpiryHeight:	schema.Endpoint.ExpiryHeight,
		Signer:		strings.TrimSpace(signer),
	}
	ad.AdvertisementHash = ComputeOffChainAdvertisementHash(ad)
	ad.SignatureHash = ComputeOffChainAdvertisementSignatureHash(ad, ad.Signer)
	return ad, ad.ValidateForSchema(schema)
}

func (ad OffChainSignedAdvertisement) ValidateForSchema(schema OffChainServiceDescriptorSchema) error {
	if err := schema.Validate(); err != nil {
		return err
	}
	if ad.ServiceID != schema.Descriptor.ServiceID || ad.Owner != schema.Descriptor.Owner ||
		ad.InterfaceHash != schema.Descriptor.Interface.InterfaceHash || ad.Endpoint != schema.Endpoint.Endpoint ||
		ad.EndpointType != schema.Endpoint.EndpointType || ad.ProviderKey != schema.Endpoint.ProviderKey ||
		ad.ExpiryHeight != schema.Endpoint.ExpiryHeight {
		return errors.New("aetracore off-chain signed advertisement does not match descriptor schema")
	}
	return ad.Validate()
}

func (ad OffChainSignedAdvertisement) Validate() error {
	if err := validatePolicyID("aetracore off-chain advertisement service id", ad.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore off-chain advertisement owner", ad.Owner); err != nil {
		return err
	}
	if err := ValidateHash("aetracore off-chain advertisement interface hash", ad.InterfaceHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain advertisement endpoint", ad.Endpoint); err != nil {
		return err
	}
	if !IsOffChainEndpointType(ad.EndpointType) {
		return fmt.Errorf("unknown aetracore off-chain endpoint type %q", ad.EndpointType)
	}
	if err := validatePolicyID("aetracore off-chain advertisement provider key", ad.ProviderKey); err != nil {
		return err
	}
	if ad.ExpiryHeight == 0 {
		return errors.New("aetracore off-chain advertisement expiry must be positive")
	}
	if ad.Signer != ad.Owner && ad.Signer != ad.ProviderKey {
		return errors.New("aetracore off-chain advertisement must be signed by owner or provider")
	}
	if err := ValidateHash("aetracore off-chain advertisement hash", ad.AdvertisementHash); err != nil {
		return err
	}
	if expected := ComputeOffChainAdvertisementHash(ad); ad.AdvertisementHash != expected {
		return fmt.Errorf("aetracore off-chain advertisement hash mismatch: expected %s", expected)
	}
	if err := ValidateHash("aetracore off-chain advertisement signature", ad.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeOffChainAdvertisementSignatureHash(ad, ad.Signer); ad.SignatureHash != expected {
		return fmt.Errorf("aetracore off-chain advertisement signature mismatch: expected %s", expected)
	}
	return nil
}

func NewOffChainSignedRequest(request OffChainServiceRequest, signer string) (OffChainSignedRequest, error) {
	request = CanonicalOffChainServiceRequest(request)
	signed := OffChainSignedRequest{
		Request:	request,
		Signer:		strings.TrimSpace(signer),
	}
	signed.SignatureHash = ComputeOffChainRequestSignatureHash(request, signed.Signer)
	return signed, signed.Validate()
}

func CanonicalOffChainServiceRequest(request OffChainServiceRequest) OffChainServiceRequest {
	request.ServiceID = strings.TrimSpace(request.ServiceID)
	request.MethodID = strings.TrimSpace(request.MethodID)
	request.Caller = strings.TrimSpace(request.Caller)
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	request.PayloadHash = strings.ToLower(strings.TrimSpace(request.PayloadHash))
	request.ProviderKey = strings.TrimSpace(request.ProviderKey)
	if request.RequestHash == "" {
		request.RequestHash = ComputeOffChainServiceRequestHash(request)
	}
	return request
}

func (signed OffChainSignedRequest) Validate() error {
	request := CanonicalOffChainServiceRequest(signed.Request)
	if err := request.Validate(); err != nil {
		return err
	}
	if signed.Signer != request.Caller {
		return errors.New("aetracore off-chain request must be signed by caller")
	}
	if err := ValidateHash("aetracore off-chain request signature", signed.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeOffChainRequestSignatureHash(request, signed.Signer); signed.SignatureHash != expected {
		return fmt.Errorf("aetracore off-chain request signature mismatch: expected %s", expected)
	}
	return nil
}

func (request OffChainServiceRequest) Validate() error {
	if err := validateOptionalHash("aetracore off-chain request call id", request.CallID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain request service id", request.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain request method id", request.MethodID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain request caller", request.Caller); err != nil {
		return err
	}
	if request.Nonce == 0 && request.IdempotencyKey == "" {
		return errors.New("aetracore off-chain request requires replay-safe nonce or idempotency key")
	}
	if request.IdempotencyKey != "" {
		if err := validatePolicyID("aetracore off-chain request idempotency key", request.IdempotencyKey); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore off-chain request payload hash", request.PayloadHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain request provider key", request.ProviderKey); err != nil {
		return err
	}
	if request.DeadlineHeight == 0 {
		return errors.New("aetracore off-chain request deadline must be positive")
	}
	if err := ValidateHash("aetracore off-chain request hash", request.RequestHash); err != nil {
		return err
	}
	if expected := ComputeOffChainServiceRequestHash(request); request.RequestHash != expected {
		return fmt.Errorf("aetracore off-chain request hash mismatch: expected %s", expected)
	}
	return nil
}

func NewOffChainSignedResponse(response OffChainServiceResponse, signer string) (OffChainSignedResponse, error) {
	response = CanonicalOffChainServiceResponse(response)
	signed := OffChainSignedResponse{
		Response:	response,
		Signer:		strings.TrimSpace(signer),
	}
	signed.SignatureHash = ComputeOffChainResponseSignatureHash(response, signed.Signer)
	return signed, signed.Validate()
}

func CanonicalOffChainServiceResponse(response OffChainServiceResponse) OffChainServiceResponse {
	response.ServiceID = strings.TrimSpace(response.ServiceID)
	response.MethodID = strings.TrimSpace(response.MethodID)
	response.RequestHash = strings.ToLower(strings.TrimSpace(response.RequestHash))
	response.ResponseHash = strings.ToLower(strings.TrimSpace(response.ResponseHash))
	response.ProviderKey = strings.TrimSpace(response.ProviderKey)
	return response
}

func (signed OffChainSignedResponse) Validate() error {
	response := CanonicalOffChainServiceResponse(signed.Response)
	if err := response.Validate(); err != nil {
		return err
	}
	if signed.Signer != response.ProviderKey {
		return errors.New("aetracore off-chain response must be signed by provider")
	}
	if (response.SettlementUse || response.ChallengeUse) && signed.SignatureHash == "" {
		return errors.New("aetracore off-chain settlement or challenge response requires signature")
	}
	if err := ValidateHash("aetracore off-chain response signature", signed.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeOffChainResponseSignatureHash(response, signed.Signer); signed.SignatureHash != expected {
		return fmt.Errorf("aetracore off-chain response signature mismatch: expected %s", expected)
	}
	return nil
}

func (response OffChainServiceResponse) Validate() error {
	if err := validateOptionalHash("aetracore off-chain response call id", response.CallID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain response service id", response.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain response method id", response.MethodID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore off-chain response request hash", response.RequestHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore off-chain response hash", response.ResponseHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain response provider key", response.ProviderKey); err != nil {
		return err
	}
	if response.Height == 0 {
		return errors.New("aetracore off-chain response height must be positive")
	}
	if response.ResultExpiryHeight <= response.Height {
		return errors.New("aetracore off-chain response expiry must exceed response height")
	}
	return nil
}

func NewOffChainReceiptAnchorMessage(response OffChainServiceResponse, proofHash string) (OffChainReceiptAnchorMessage, error) {
	response = CanonicalOffChainServiceResponse(response)
	if err := response.Validate(); err != nil {
		return OffChainReceiptAnchorMessage{}, err
	}
	anchor := OffChainReceiptAnchorMessage{
		ServiceID:	response.ServiceID,
		CallID:		response.CallID,
		RequestHash:	response.RequestHash,
		ResponseHash:	response.ResponseHash,
		ProviderKey:	response.ProviderKey,
		Height:		response.Height,
		ProofHash:	strings.ToLower(strings.TrimSpace(proofHash)),
	}
	if anchor.ProofHash == "" {
		anchor.ProofHash = ComputeOffChainReceiptAnchorProofHash(anchor)
	}
	anchor.AnchorHash = ComputeOffChainReceiptAnchorHash(anchor)
	return anchor, anchor.Validate()
}

func (anchor OffChainReceiptAnchorMessage) Validate() error {
	if err := validatePolicyID("aetracore off-chain anchor service id", anchor.ServiceID); err != nil {
		return err
	}
	if err := validateOptionalHash("aetracore off-chain anchor call id", anchor.CallID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore off-chain anchor request hash", anchor.RequestHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore off-chain anchor response hash", anchor.ResponseHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain anchor provider key", anchor.ProviderKey); err != nil {
		return err
	}
	if anchor.Height == 0 {
		return errors.New("aetracore off-chain anchor height must be positive")
	}
	if err := ValidateHash("aetracore off-chain anchor proof hash", anchor.ProofHash); err != nil {
		return err
	}
	if expected := ComputeOffChainReceiptAnchorProofHash(anchor); anchor.ProofHash != expected {
		return fmt.Errorf("aetracore off-chain anchor proof hash mismatch: expected %s", expected)
	}
	if err := ValidateHash("aetracore off-chain anchor hash", anchor.AnchorHash); err != nil {
		return err
	}
	if expected := ComputeOffChainReceiptAnchorHash(anchor); anchor.AnchorHash != expected {
		return fmt.Errorf("aetracore off-chain anchor hash mismatch: expected %s", expected)
	}
	return nil
}

func NewOffChainEndpointRenewal(schema OffChainServiceDescriptorSchema, endpoint string, expiryHeight uint64, signer string) (OffChainEndpointRenewal, error) {
	if err := schema.Validate(); err != nil {
		return OffChainEndpointRenewal{}, err
	}
	renewal := OffChainEndpointRenewal{
		ServiceID:		schema.Descriptor.ServiceID,
		Endpoint:		strings.TrimSpace(endpoint),
		EndpointType:		schema.Endpoint.EndpointType,
		ProviderKey:		schema.Endpoint.ProviderKey,
		RenewedAtHeight:	schema.Descriptor.UpdatedHeight,
		ExpiryHeight:		expiryHeight,
		Signer:			strings.TrimSpace(signer),
	}
	renewal.AdvertisementHash = ComputeOffChainEndpointRenewalHash(renewal)
	renewal.SignatureHash = ComputeOffChainEndpointRenewalSignatureHash(renewal, renewal.Signer)
	return renewal, renewal.ValidateForSchema(schema)
}

func (renewal OffChainEndpointRenewal) ValidateForSchema(schema OffChainServiceDescriptorSchema) error {
	if err := schema.Validate(); err != nil {
		return err
	}
	if renewal.ServiceID != schema.Descriptor.ServiceID || renewal.EndpointType != schema.Endpoint.EndpointType ||
		renewal.ProviderKey != schema.Endpoint.ProviderKey {
		return errors.New("aetracore off-chain endpoint renewal does not match descriptor schema")
	}
	if renewal.ExpiryHeight <= schema.Endpoint.ExpiryHeight {
		return errors.New("aetracore off-chain endpoint renewal must extend endpoint expiry")
	}
	return renewal.Validate()
}

func (renewal OffChainEndpointRenewal) Validate() error {
	if err := validatePolicyID("aetracore off-chain renewal service id", renewal.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore off-chain renewal endpoint", renewal.Endpoint); err != nil {
		return err
	}
	if !IsOffChainEndpointType(renewal.EndpointType) {
		return fmt.Errorf("unknown aetracore off-chain endpoint type %q", renewal.EndpointType)
	}
	if err := validatePolicyID("aetracore off-chain renewal provider key", renewal.ProviderKey); err != nil {
		return err
	}
	if renewal.RenewedAtHeight == 0 {
		return errors.New("aetracore off-chain renewal height must be positive")
	}
	if renewal.ExpiryHeight <= renewal.RenewedAtHeight {
		return errors.New("aetracore off-chain renewal expiry must exceed renewal height")
	}
	if renewal.Signer != renewal.ProviderKey {
		return errors.New("aetracore off-chain endpoint renewal must be signed by provider")
	}
	if err := ValidateHash("aetracore off-chain renewal advertisement hash", renewal.AdvertisementHash); err != nil {
		return err
	}
	if expected := ComputeOffChainEndpointRenewalHash(renewal); renewal.AdvertisementHash != expected {
		return fmt.Errorf("aetracore off-chain renewal hash mismatch: expected %s", expected)
	}
	if err := ValidateHash("aetracore off-chain renewal signature", renewal.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeOffChainEndpointRenewalSignatureHash(renewal, renewal.Signer); renewal.SignatureHash != expected {
		return fmt.Errorf("aetracore off-chain renewal signature mismatch: expected %s", expected)
	}
	return nil
}

func ComputeOffChainServiceDescriptorSchemaHash(schema OffChainServiceDescriptorSchema) string {
	return hashParts(
		"aetra-aek-offchain-service-schema-v1",
		ComputeServiceDescriptorHash(schema.Descriptor),
		schema.Endpoint.Endpoint,
		string(schema.Endpoint.EndpointType),
		schema.Endpoint.ProviderKey,
		fmt.Sprint(schema.Endpoint.ExpiryHeight),
		schema.Endpoint.MetadataHash,
		string(schema.RequestSigningPolicy),
		string(schema.ResponseSigningPolicy),
		string(schema.ProofAnchorPolicy),
		string(schema.AvailabilityPolicy),
		fmt.Sprint(schema.ResultExpiry),
	)
}

func ComputeOffChainAdvertisementHash(ad OffChainSignedAdvertisement) string {
	return hashParts(
		"aetra-aek-offchain-advertisement-v1",
		ad.ServiceID,
		ad.Owner,
		ad.InterfaceHash,
		ad.Endpoint,
		string(ad.EndpointType),
		ad.ProviderKey,
		fmt.Sprint(ad.ExpiryHeight),
	)
}

func ComputeOffChainAdvertisementSignatureHash(ad OffChainSignedAdvertisement, signer string) string {
	return hashParts("aetra-aek-offchain-advertisement-signature-v1", ad.AdvertisementHash, signer)
}

func ComputeOffChainServiceRequestHash(request OffChainServiceRequest) string {
	return hashParts(
		"aetra-aek-offchain-request-v1",
		request.CallID,
		request.ServiceID,
		request.MethodID,
		request.Caller,
		fmt.Sprint(request.Nonce),
		request.IdempotencyKey,
		request.PayloadHash,
		request.ProviderKey,
		fmt.Sprint(request.DeadlineHeight),
	)
}

func ComputeOffChainRequestSignatureHash(request OffChainServiceRequest, signer string) string {
	return hashParts("aetra-aek-offchain-request-signature-v1", request.RequestHash, signer)
}

func ComputeOffChainResponseSignatureHash(response OffChainServiceResponse, signer string) string {
	return hashParts(
		"aetra-aek-offchain-response-signature-v1",
		response.CallID,
		response.ServiceID,
		response.MethodID,
		response.RequestHash,
		response.ResponseHash,
		response.ProviderKey,
		fmt.Sprint(response.Height),
		fmt.Sprint(response.ResultExpiryHeight),
		fmt.Sprint(response.SettlementUse),
		fmt.Sprint(response.ChallengeUse),
		signer,
	)
}

func ComputeOffChainReceiptAnchorProofHash(anchor OffChainReceiptAnchorMessage) string {
	return hashParts(
		"aetra-aek-offchain-receipt-proof-anchor-v1",
		anchor.RequestHash,
		anchor.ResponseHash,
		anchor.ProviderKey,
		fmt.Sprint(anchor.Height),
	)
}

func ComputeOffChainReceiptAnchorHash(anchor OffChainReceiptAnchorMessage) string {
	return hashParts(
		"aetra-aek-offchain-receipt-anchor-v1",
		anchor.ServiceID,
		anchor.CallID,
		anchor.RequestHash,
		anchor.ResponseHash,
		anchor.ProviderKey,
		fmt.Sprint(anchor.Height),
		anchor.ProofHash,
	)
}

func ComputeOffChainEndpointRenewalHash(renewal OffChainEndpointRenewal) string {
	return hashParts(
		"aetra-aek-offchain-endpoint-renewal-v1",
		renewal.ServiceID,
		renewal.Endpoint,
		string(renewal.EndpointType),
		renewal.ProviderKey,
		fmt.Sprint(renewal.RenewedAtHeight),
		fmt.Sprint(renewal.ExpiryHeight),
	)
}

func ComputeOffChainEndpointRenewalSignatureHash(renewal OffChainEndpointRenewal, signer string) string {
	return hashParts("aetra-aek-offchain-endpoint-renewal-signature-v1", renewal.AdvertisementHash, signer)
}

func IsOffChainEndpointType(endpointType OffChainEndpointType) bool {
	switch endpointType {
	case OffChainEndpointRPC, OffChainEndpointOverlay, OffChainEndpointProviderMesh, OffChainEndpointServiceNetwork:
		return true
	default:
		return false
	}
}

func IsOffChainRequestSigningPolicy(policy OffChainRequestSigningPolicy) bool {
	return policy == OffChainRequestCallerSigned
}

func IsOffChainResponseSigningPolicy(policy OffChainResponseSigningPolicy) bool {
	return policy == OffChainResponseProviderSigned
}

func IsOffChainProofAnchorPolicy(policy OffChainProofAnchorPolicy) bool {
	switch policy {
	case OffChainProofAnchorNone, OffChainProofAnchorOptional, OffChainProofAnchorRequired:
		return true
	default:
		return false
	}
}

func IsOffChainAvailabilityPolicy(policy OffChainAvailabilityPolicy) bool {
	switch policy {
	case OffChainAvailabilitySignedAdvertisement, OffChainAvailabilityRenewableEndpoint:
		return true
	default:
		return false
	}
}

func validateOffChainMethods(methods []OffChainServiceMethod) error {
	if len(methods) == 0 {
		return errors.New("aetracore off-chain service requires method set")
	}
	var previous string
	seen := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seen[method.MethodID]; found {
			return fmt.Errorf("duplicate aetracore off-chain method %s", method.MethodID)
		}
		seen[method.MethodID] = struct{}{}
		if i > 0 && previous >= method.MethodID {
			return errors.New("aetracore off-chain methods must be sorted canonically")
		}
		previous = method.MethodID
	}
	return nil
}

func cloneOffChainMethods(methods []OffChainServiceMethod) []OffChainServiceMethod {
	out := make([]OffChainServiceMethod, len(methods))
	copy(out, methods)
	return out
}

func aggregateOffChainFailurePolicy(methods []OffChainServiceMethod) ServiceFailureBehavior {
	for _, method := range methods {
		if method.FailurePolicy == ServiceFailureSlashProvider {
			return ServiceFailureSlashProvider
		}
		if method.FailurePolicy == ServiceFailureRetry {
			return ServiceFailureRetry
		}
	}
	if len(methods) == 0 {
		return ServiceFailureRetry
	}
	return methods[0].FailurePolicy
}

func offChainVerificationModel(policy OffChainProofAnchorPolicy) ServiceVerificationModel {
	if policy == OffChainProofAnchorRequired || policy == OffChainProofAnchorOptional {
		return ServiceVerificationProofAnchored
	}
	return ServiceVerificationSignedResult
}

func offChainProofFormat(policy OffChainProofAnchorPolicy) string {
	if policy == OffChainProofAnchorRequired || policy == OffChainProofAnchorOptional {
		return "aek-proof-anchor-v1"
	}
	return ""
}
