package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RoutingIntegrationStateObjectV2 string
type RoutingIntegrationMessageNameV2 string
type RoutingIntegrationQueryNameV2 string
type RoutingIntegrationFailureModeV2 string
type RoutingIntegrationIntegrationPointV2 string

const (
	RoutingIntegrationModulePathV2	= "routing-integration-module"

	RoutingIntegrationStateRoutingPolicy			RoutingIntegrationStateObjectV2	= "RoutingPolicy"
	RoutingIntegrationStateIdentityTransactionMapping	RoutingIntegrationStateObjectV2	= "IdentityTransactionMapping"
	RoutingIntegrationStateContractInvocationMapping	RoutingIntegrationStateObjectV2	= "ContractInvocationMapping"
	RoutingIntegrationStateServiceMapping			RoutingIntegrationStateObjectV2	= "ServiceMapping"
	RoutingIntegrationStateInterfaceMapping			RoutingIntegrationStateObjectV2	= "InterfaceMapping"
	RoutingIntegrationStateExecutionHintPolicy		RoutingIntegrationStateObjectV2	= "ExecutionHintPolicy"

	RoutingIntegrationMsgSetRoutingPolicy			RoutingIntegrationMessageNameV2	= "MsgSetRoutingPolicy"
	RoutingIntegrationMsgUpdateExecutionHints		RoutingIntegrationMessageNameV2	= "MsgUpdateExecutionHints"
	RoutingIntegrationMsgRegisterInterfaceDescriptor	RoutingIntegrationMessageNameV2	= "MsgRegisterInterfaceDescriptor"
	RoutingIntegrationMsgRegisterServiceEndpoint		RoutingIntegrationMessageNameV2	= "MsgRegisterServiceEndpoint"
	RoutingIntegrationMsgClearRoutingMetadata		RoutingIntegrationMessageNameV2	= "MsgClearRoutingMetadata"

	RoutingIntegrationQueryTransactionMapping		RoutingIntegrationQueryNameV2	= "QueryTransactionMapping"
	RoutingIntegrationQueryContractInvocationMapping	RoutingIntegrationQueryNameV2	= "QueryContractInvocationMapping"
	RoutingIntegrationQueryServiceMapping			RoutingIntegrationQueryNameV2	= "QueryServiceMapping"
	RoutingIntegrationQueryInterfaceMapping			RoutingIntegrationQueryNameV2	= "QueryInterfaceMapping"
	RoutingIntegrationQueryExecutionHints			RoutingIntegrationQueryNameV2	= "QueryExecutionHints"
	RoutingIntegrationQueryResolvedExecutionTarget		RoutingIntegrationQueryNameV2	= "QueryResolvedExecutionTarget"

	RoutingIntegrationFailureStaleMetadata			RoutingIntegrationFailureModeV2	= "stale_routing_metadata_after_resolver_update"
	RoutingIntegrationFailureWrongInterfaceTarget		RoutingIntegrationFailureModeV2	= "interface_descriptor_points_to_wrong_target"
	RoutingIntegrationFailureServiceEndpointUnavailable	RoutingIntegrationFailureModeV2	= "service_endpoint_proof_succeeds_but_endpoint_unavailable"
	RoutingIntegrationFailureExecutionHintConflict		RoutingIntegrationFailureModeV2	= "execution_hint_conflicts_with_target_module_rules"
	RoutingIntegrationFailureAdvisoryAsAuthorization	RoutingIntegrationFailureModeV2	= "client_treats_advisory_metadata_as_authorization"

	RoutingIntegrationIntegrationResolverModule	RoutingIntegrationIntegrationPointV2	= "resolver_module"
	RoutingIntegrationIntegrationFeeModule		RoutingIntegrationIntegrationPointV2	= "fee_module"
	RoutingIntegrationIntegrationContractExecution	RoutingIntegrationIntegrationPointV2	= "contract_execution_layer"
	RoutingIntegrationIntegrationWalletSDK		RoutingIntegrationIntegrationPointV2	= "wallet_sdk"
	RoutingIntegrationIntegrationServiceClients	RoutingIntegrationIntegrationPointV2	= "service_clients"
)

type RoutingIntegrationFailureCoverageV2 struct {
	Mode		RoutingIntegrationFailureModeV2
	Guard		string
	StoreScope	string
}

type RoutingIntegrationModuleBreakdownV2 struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]RoutingIntegrationStateObjectV2
	Messages		[]RoutingIntegrationMessageNameV2
	Queries			[]RoutingIntegrationQueryNameV2
	FailureModes		[]RoutingIntegrationFailureCoverageV2
	IntegrationPoints	[]RoutingIntegrationIntegrationPointV2
	BackingPrimitives	[]string
	StoreKeys		[]string
	BreakdownHash		string
}

type RoutingIntegrationPolicyV2 struct {
	PolicyID			string
	AllowAdvisoryRoutingMetadata	bool
	RequireProofForWalletSDK	bool
	RequireInterfaceHash		bool
	RequireSimulationWhenHinted	bool
	EndpointAvailabilityIsAdvisory	bool
	PolicyHash			string
}

type IdentityTransactionMappingV2 struct {
	Name			string
	Address			sdk.AccAddress
	ProofVerified		bool
	ProofHeight		uint64
	RecordVersion		uint64
	FreshUntilHeight	uint64
	AuditMemo		string
	MappingHash		string
}

type ContractInvocationMappingV2 struct {
	Name			string
	TargetID		string
	InterfaceID		string
	InterfaceHash		string
	ContractAddress		sdk.AccAddress
	Entrypoint		string
	ProofVerified		bool
	ProofHeight		uint64
	RecordVersion		uint64
	RequiresConfirm		bool
	SimulationNeeded	bool
	MappingHash		string
}

type ServiceMappingV2 struct {
	Name					string
	ServiceID				string
	Endpoint				ServiceEndpointV2
	FallbackEndpoints			[]ServiceEndpointV2
	ProofVerified				bool
	ProofHeight				uint64
	RecordVersion				uint64
	EndpointAvailabilityConsensusGuaranteed	bool
	MappingHash				string
}

type InterfaceMappingV2 struct {
	Name				string
	InterfaceID			string
	SchemaHash			string
	SchemaHashVerified		bool
	RenderPolicySupported		bool
	UserConfirmationRequired	bool
	ExecutionTargetImmutable	bool
	ProofVerified			bool
	ProofHeight			uint64
	RecordVersion			uint64
	MappingHash			string
}

type ExecutionHintPolicyV2 struct {
	MaxGasLimitHint		uint64
	AllowedFeeModes		[]string
	RequiresMemo		bool
	SimulationRequired	bool
	AllowAsyncExecution	bool
	PolicyHash		string
}

type RoutingIntegrationResolvedExecutionTargetV2 struct {
	Name			string
	TargetType		IdentityResolutionTargetTypeV2
	TargetKey		string
	Address			sdk.AccAddress
	Endpoint		string
	Descriptor		string
	Route			RoutingMetadataV2
	RecordVersion		uint64
	ProofHeight		uint64
	ProofVerified		bool
	FreshUntilHeight	uint64
	ResultHash		string
}

type RoutingIntegrationResolvedExecutionTargetRequestV2 struct {
	Name			string
	TargetType		IdentityResolutionTargetTypeV2
	TargetKey		string
	InterfaceID		string
	ExpectedHash		string
	Method			string
	PayloadHash		string
	State			IdentityState
	Height			uint64
	RecordTTL		uint64
	CurrentHeight		uint64
	FreshnessThreshold	uint64
	ExpectedChainID		string
	TrustedHeader		IdentityTrustedHeaderV2
	Proof			*IdentityResolutionProofFormatV2
}

type RoutingIntegrationWalletSDKHelperV2 struct {
	TransactionMapping	*IdentityTransactionMappingV2
	ContractMapping		*ContractInvocationMappingV2
	ServiceMapping		*ServiceMappingV2
	InterfaceMapping	*InterfaceMappingV2
	ProofRequired		bool
	UserConfirmation	bool
	AdvisoryOnly		bool
	HelperHash		string
}

func DefaultRoutingIntegrationModuleBreakdownV2() (RoutingIntegrationModuleBreakdownV2, error) {
	breakdown := RoutingIntegrationModuleBreakdownV2{
		ModulePath:	RoutingIntegrationModulePathV2,
		Purpose: []string{
			"contract_invocation_mapping",
			"execution_hint_mapping",
			"interface_mapping",
			"resolved_execution_target_query",
			"service_mapping",
			"transaction_mapping",
		},
		StateObjects:	requiredRoutingIntegrationStateObjectsV2(),
		Messages:	requiredRoutingIntegrationMessagesV2(),
		Queries:	requiredRoutingIntegrationQueriesV2(),
		FailureModes: []RoutingIntegrationFailureCoverageV2{
			{Mode: RoutingIntegrationFailureAdvisoryAsAuthorization, Guard: "ValidateRoutingIntegrationAdvisoryAuthorizationV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: RoutingIntegrationFailureExecutionHintConflict, Guard: "ValidateRoutingIntegrationExecutionHintsV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: RoutingIntegrationFailureServiceEndpointUnavailable, Guard: "ValidateRoutingIntegrationServiceEndpointAvailabilityV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: RoutingIntegrationFailureStaleMetadata, Guard: "ValidateRoutingIntegrationMetadataFreshnessV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: RoutingIntegrationFailureWrongInterfaceTarget, Guard: "ValidateRoutingIntegrationInterfaceTargetV2", StoreScope: IdentityStoreV2SpecInterfaceMetadataPrefix},
		},
		IntegrationPoints:	requiredRoutingIntegrationIntegrationPointsV2(),
		BackingPrimitives: []string{
			"BuildIdentityInterfaceSchemaMappingV2",
			"BuildIdentityInvokeByNameV2",
			"BuildIdentitySendByNameV2",
			"BuildIdentityServiceDiscoveryV2",
			"ResolveNamedExecutionTarget",
		},
		StoreKeys: []string{
			IdentityStoreV2SpecInterfaceMetadataPrefix,
			IdentityStoreV2SpecResolversPrefix,
			IdentityStoreV2SpecResolverIndexPrefix,
		},
	}
	return NewRoutingIntegrationModuleBreakdownV2(breakdown)
}

func NewRoutingIntegrationModuleBreakdownV2(breakdown RoutingIntegrationModuleBreakdownV2) (RoutingIntegrationModuleBreakdownV2, error) {
	breakdown = canonicalRoutingIntegrationModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return RoutingIntegrationModuleBreakdownV2{}, err
	}
	breakdown.BreakdownHash = ComputeRoutingIntegrationModuleBreakdownHashV2(breakdown)
	return breakdown, breakdown.Validate()
}

func (breakdown RoutingIntegrationModuleBreakdownV2) ValidateFormat() error {
	if breakdown.ModulePath != RoutingIntegrationModulePathV2 {
		return errors.New("routing integration breakdown must describe routing-integration-module")
	}
	if err := validateBreakdownTokenSetV2("routing integration purpose", breakdown.Purpose, nil); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("routing integration module", "state object", breakdown.StateObjects, requiredRoutingIntegrationStateObjectsV2(), IsRoutingIntegrationStateObjectV2); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("routing integration module", "message", breakdown.Messages, requiredRoutingIntegrationMessagesV2(), IsRoutingIntegrationMessageNameV2); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("routing integration module", "query", breakdown.Queries, requiredRoutingIntegrationQueriesV2(), IsRoutingIntegrationQueryNameV2); err != nil {
		return err
	}
	if err := validateRoutingIntegrationFailuresV2(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateModuleEnumSetV2("routing integration module", "integration", breakdown.IntegrationPoints, requiredRoutingIntegrationIntegrationPointsV2(), IsRoutingIntegrationIntegrationPointV2); err != nil {
		return err
	}
	if err := validateBreakdownTokenSetV2("routing integration backing primitive", breakdown.BackingPrimitives, []string{"BuildIdentityInterfaceSchemaMappingV2", "BuildIdentityInvokeByNameV2", "BuildIdentitySendByNameV2", "BuildIdentityServiceDiscoveryV2", "ResolveNamedExecutionTarget"}); err != nil {
		return err
	}
	if err := validateBreakdownStoreKeysV2("routing integration", breakdown.StoreKeys, []string{IdentityStoreV2SpecInterfaceMetadataPrefix, IdentityStoreV2SpecResolverIndexPrefix, IdentityStoreV2SpecResolversPrefix}); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return validateHexHash("routing integration breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown RoutingIntegrationModuleBreakdownV2) Validate() error {
	breakdown = canonicalRoutingIntegrationModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("routing integration breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeRoutingIntegrationModuleBreakdownHashV2(breakdown) {
		return errors.New("routing integration breakdown hash mismatch")
	}
	return nil
}

func BuildRoutingIntegrationTransactionMappingV2(request IdentitySendByNameRequestV2) (IdentityTransactionMappingV2, error) {
	result, err := BuildIdentitySendByNameV2(request)
	if err != nil {
		return IdentityTransactionMappingV2{}, err
	}
	mapping := IdentityTransactionMappingV2{
		Name:			result.NormalizedName,
		Address:		cloneSpecAddress(result.Address),
		ProofVerified:		result.ProofVerified,
		ProofHeight:		result.ProofHeight,
		RecordVersion:		result.RecordVersion,
		FreshUntilHeight:	result.FreshUntilHeight,
		AuditMemo:		result.AuditMemo,
	}
	mapping.MappingHash = ComputeIdentityTransactionMappingHashV2(mapping)
	return mapping, ValidateIdentityTransactionMappingV2(mapping)
}

func BuildRoutingIntegrationContractInvocationMappingV2(request IdentityInvokeByNameRequestV2) (ContractInvocationMappingV2, error) {
	result, err := BuildIdentityInvokeByNameV2(request)
	if err != nil {
		return ContractInvocationMappingV2{}, err
	}
	mapping := ContractInvocationMappingV2{
		Name:			result.NormalizedName,
		TargetID:		result.TargetID,
		InterfaceID:		result.InterfaceID,
		InterfaceHash:		result.InterfaceHash,
		ContractAddress:	cloneSpecAddress(result.ContractAddress),
		Entrypoint:		result.Entrypoint,
		ProofVerified:		result.ProofVerified,
		ProofHeight:		result.ProofHeight,
		RecordVersion:		result.RecordVersion,
		RequiresConfirm:	result.RequiresInterfaceConfirmation,
		SimulationNeeded:	result.SimulationRequiredBeforeSigning,
	}
	if result.VerifiedInterfaceDescriptor != nil {
		if err := ValidateRoutingIntegrationInterfaceTargetV2(result.TargetID, *result.VerifiedInterfaceDescriptor); err != nil {
			return ContractInvocationMappingV2{}, err
		}
	}
	mapping.MappingHash = ComputeContractInvocationMappingHashV2(mapping)
	return mapping, ValidateContractInvocationMappingV2(mapping)
}

func BuildRoutingIntegrationServiceMappingV2(request IdentityServiceDiscoveryRequestV2) (ServiceMappingV2, error) {
	result, err := BuildIdentityServiceDiscoveryV2(request)
	if err != nil {
		return ServiceMappingV2{}, err
	}
	mapping := ServiceMappingV2{
		Name:						request.Name,
		ServiceID:					result.Endpoint.ServiceID,
		Endpoint:					result.Endpoint,
		FallbackEndpoints:				append([]ServiceEndpointV2(nil), result.FallbackEndpoints...),
		ProofVerified:					result.ProofVerified,
		ProofHeight:					result.ProofHeight,
		RecordVersion:					result.RecordVersion,
		EndpointAvailabilityConsensusGuaranteed:	result.EndpointAvailabilityConsensusGuaranteed,
	}
	mapping.MappingHash = ComputeServiceMappingHashV2(mapping)
	return mapping, ValidateServiceMappingV2(mapping)
}

func BuildRoutingIntegrationInterfaceMappingV2(request IdentityInterfaceSchemaRequestV2) (InterfaceMappingV2, error) {
	result, err := BuildIdentityInterfaceSchemaMappingV2(request)
	if err != nil {
		return InterfaceMappingV2{}, err
	}
	mapping := InterfaceMappingV2{
		Name:				request.Name,
		InterfaceID:			result.Descriptor.InterfaceID,
		SchemaHash:			result.SchemaHash,
		SchemaHashVerified:		result.SchemaHashVerified,
		RenderPolicySupported:		result.RenderPolicySupported,
		UserConfirmationRequired:	result.UserConfirmationRequired,
		ExecutionTargetImmutable:	result.ExecutionTargetImmutable,
		ProofVerified:			result.ProofVerified,
		ProofHeight:			result.ProofHeight,
		RecordVersion:			result.RecordVersion,
	}
	mapping.MappingHash = ComputeInterfaceMappingHashV2(mapping)
	return mapping, ValidateInterfaceMappingV2(mapping)
}

func QueryRoutingIntegrationResolvedExecutionTargetV2(request RoutingIntegrationResolvedExecutionTargetRequestV2) (RoutingIntegrationResolvedExecutionTargetV2, error) {
	normalized, err := NormalizeAETDomain(request.Name)
	if err != nil {
		return RoutingIntegrationResolvedExecutionTargetV2{}, err
	}
	switch request.TargetType {
	case IdentityResolutionTargetPrimary:
		mapping, err := BuildRoutingIntegrationTransactionMappingV2(IdentitySendByNameRequestV2{
			Name:			normalized,
			State:			request.State,
			Height:			request.Height,
			RecordTTL:		request.RecordTTL,
			CurrentHeight:		request.CurrentHeight,
			FreshnessThreshold:	request.FreshnessThreshold,
			ExpectedChainID:	request.ExpectedChainID,
			TrustedHeader:		request.TrustedHeader,
			Proof:			request.Proof,
		})
		if err != nil {
			return RoutingIntegrationResolvedExecutionTargetV2{}, err
		}
		return routingIntegrationResolvedTargetFromTransactionV2(mapping), nil
	case IdentityResolutionTargetContract:
		mapping, err := BuildRoutingIntegrationContractInvocationMappingV2(IdentityInvokeByNameRequestV2{
			Name:			normalized,
			TargetID:		request.TargetKey,
			InterfaceID:		request.InterfaceID,
			ExpectedInterfaceHash:	request.ExpectedHash,
			Method:			request.Method,
			PayloadHash:		request.PayloadHash,
			State:			request.State,
			Height:			request.Height,
			RecordTTL:		request.RecordTTL,
			CurrentHeight:		request.CurrentHeight,
			FreshnessThreshold:	request.FreshnessThreshold,
			ExpectedChainID:	request.ExpectedChainID,
			TrustedHeader:		request.TrustedHeader,
			Proof:			request.Proof,
		})
		if err != nil {
			return RoutingIntegrationResolvedExecutionTargetV2{}, err
		}
		return routingIntegrationResolvedTargetFromContractV2(mapping), nil
	case IdentityResolutionTargetService:
		mapping, err := BuildRoutingIntegrationServiceMappingV2(IdentityServiceDiscoveryRequestV2{
			Name:			normalized,
			ServiceID:		request.TargetKey,
			SupportedTransports:	[]string{"https", "grpcs", "wss", "aetra", "ipfs"},
			AllowedAuthPolicies:	[]string{"none", "token"},
			State:			request.State,
			Height:			request.Height,
			RecordTTL:		request.RecordTTL,
			CurrentHeight:		request.CurrentHeight,
			ExpectedChainID:	request.ExpectedChainID,
			TrustedHeader:		request.TrustedHeader,
			Proof:			request.Proof,
		})
		if err != nil {
			return RoutingIntegrationResolvedExecutionTargetV2{}, err
		}
		return routingIntegrationResolvedTargetFromServiceV2(mapping), nil
	case IdentityResolutionTargetInterface:
		mapping, err := BuildRoutingIntegrationInterfaceMappingV2(IdentityInterfaceSchemaRequestV2{
			Name:			normalized,
			InterfaceID:		request.TargetKey,
			ExpectedSchemaHash:	request.ExpectedHash,
			WalletPolicy:		DefaultIdentityInterfaceWalletPolicyV2(),
			State:			request.State,
			Height:			request.Height,
			RecordTTL:		request.RecordTTL,
			CurrentHeight:		request.CurrentHeight,
			ExpectedChainID:	request.ExpectedChainID,
			TrustedHeader:		request.TrustedHeader,
			Proof:			request.Proof,
		})
		if err != nil {
			return RoutingIntegrationResolvedExecutionTargetV2{}, err
		}
		return routingIntegrationResolvedTargetFromInterfaceV2(mapping), nil
	default:
		return RoutingIntegrationResolvedExecutionTargetV2{}, fmt.Errorf("routing integration unsupported target type %q", request.TargetType)
	}
}

func ValidateRoutingIntegrationMetadataFreshnessV2(mappingVersion uint64, currentResolverVersion uint64) error {
	if mappingVersion == 0 || currentResolverVersion == 0 {
		return errors.New("routing integration metadata versions are required")
	}
	if mappingVersion != currentResolverVersion {
		return fmt.Errorf("%s: mapping version %d current resolver version %d", RoutingIntegrationFailureStaleMetadata, mappingVersion, currentResolverVersion)
	}
	return nil
}

func ValidateRoutingIntegrationInterfaceTargetV2(targetID string, descriptor InterfaceDescriptorV2) error {
	if err := validateInterfaceDescriptorsV2([]InterfaceDescriptorV2{descriptor}); err != nil {
		return err
	}
	if descriptor.ContractTargetIDOptional != "" && descriptor.ContractTargetIDOptional != targetID {
		return fmt.Errorf("%s: descriptor target %s requested %s", RoutingIntegrationFailureWrongInterfaceTarget, descriptor.ContractTargetIDOptional, targetID)
	}
	return nil
}

func ValidateRoutingIntegrationServiceEndpointAvailabilityV2(mapping ServiceMappingV2, endpointAvailable bool) error {
	if err := ValidateServiceMappingV2(mapping); err != nil {
		return err
	}
	if mapping.EndpointAvailabilityConsensusGuaranteed {
		return errors.New("routing integration service endpoint availability must remain advisory")
	}
	if !endpointAvailable {
		return fmt.Errorf("%s: endpoint %s unavailable", RoutingIntegrationFailureServiceEndpointUnavailable, mapping.Endpoint.Endpoint)
	}
	return nil
}

func ValidateRoutingIntegrationExecutionHintsV2(record UnifiedResolutionRecordV2, policy ExecutionHintPolicyV2) error {
	if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
		return err
	}
	if policy.MaxGasLimitHint == 0 {
		policy.MaxGasLimitHint = MaxExecutionGasLimitHintV2
	}
	allowedFeeModes := stringSet(policy.AllowedFeeModes)
	for _, hint := range record.ExecutionHints {
		if hint.DefaultGasLimitHint > policy.MaxGasLimitHint {
			return fmt.Errorf("%s: gas hint exceeds target module limit", RoutingIntegrationFailureExecutionHintConflict)
		}
		if hint.PreferredFeeMode != "" && len(allowedFeeModes) > 0 {
			if _, found := allowedFeeModes[hint.PreferredFeeMode]; !found {
				return fmt.Errorf("%s: fee mode %s not allowed", RoutingIntegrationFailureExecutionHintConflict, hint.PreferredFeeMode)
			}
		}
		if policy.RequiresMemo && !hint.RequiresMemo {
			return fmt.Errorf("%s: target module requires memo", RoutingIntegrationFailureExecutionHintConflict)
		}
		if policy.SimulationRequired && !hint.SimulationRequired {
			return fmt.Errorf("%s: target module requires simulation", RoutingIntegrationFailureExecutionHintConflict)
		}
		if !policy.AllowAsyncExecution && hint.AsyncAllowed {
			return fmt.Errorf("%s: async execution is not allowed by target module", RoutingIntegrationFailureExecutionHintConflict)
		}
	}
	return nil
}

func ValidateRoutingIntegrationAdvisoryAuthorizationV2(helper RoutingIntegrationWalletSDKHelperV2, metadataUsedAsAuthorization bool) error {
	if metadataUsedAsAuthorization {
		return fmt.Errorf("%s: resolver routing metadata is advisory only", RoutingIntegrationFailureAdvisoryAsAuthorization)
	}
	if helper.AdvisoryOnly && !helper.UserConfirmation {
		return errors.New("routing integration advisory helper requires user confirmation")
	}
	return nil
}

func BuildRoutingIntegrationWalletSDKHelperV2(tx *IdentityTransactionMappingV2, contract *ContractInvocationMappingV2, service *ServiceMappingV2, iface *InterfaceMappingV2, requireProof bool) (RoutingIntegrationWalletSDKHelperV2, error) {
	helper := RoutingIntegrationWalletSDKHelperV2{
		TransactionMapping:	tx,
		ContractMapping:	contract,
		ServiceMapping:		service,
		InterfaceMapping:	iface,
		ProofRequired:		requireProof,
		UserConfirmation:	true,
		AdvisoryOnly:		true,
	}
	if tx != nil {
		if err := ValidateIdentityTransactionMappingV2(*tx); err != nil {
			return RoutingIntegrationWalletSDKHelperV2{}, err
		}
	}
	if contract != nil {
		if err := ValidateContractInvocationMappingV2(*contract); err != nil {
			return RoutingIntegrationWalletSDKHelperV2{}, err
		}
	}
	if service != nil {
		if err := ValidateServiceMappingV2(*service); err != nil {
			return RoutingIntegrationWalletSDKHelperV2{}, err
		}
	}
	if iface != nil {
		if err := ValidateInterfaceMappingV2(*iface); err != nil {
			return RoutingIntegrationWalletSDKHelperV2{}, err
		}
	}
	helper.HelperHash = ComputeRoutingIntegrationWalletSDKHelperHashV2(helper)
	return helper, ValidateRoutingIntegrationAdvisoryAuthorizationV2(helper, false)
}

func ValidateIdentityTransactionMappingV2(mapping IdentityTransactionMappingV2) error {
	if _, err := NormalizeAETDomain(mapping.Name); err != nil {
		return err
	}
	if err := validateSpecAddress("routing integration transaction address", mapping.Address); err != nil {
		return err
	}
	if mapping.ProofHeight == 0 || mapping.RecordVersion == 0 || mapping.FreshUntilHeight == 0 {
		return errors.New("routing integration transaction mapping heights and version are required")
	}
	if mapping.MappingHash == "" || mapping.MappingHash != ComputeIdentityTransactionMappingHashV2(mapping) {
		return errors.New("routing integration transaction mapping hash mismatch")
	}
	return nil
}

func ValidateContractInvocationMappingV2(mapping ContractInvocationMappingV2) error {
	if _, err := NormalizeAETDomain(mapping.Name); err != nil {
		return err
	}
	if err := validateUnifiedRecordKey("routing integration target_id", mapping.TargetID); err != nil {
		return err
	}
	if err := validateSpecAddress("routing integration contract address", mapping.ContractAddress); err != nil {
		return err
	}
	if err := validateContractEntrypointV2(mapping.Entrypoint); err != nil {
		return err
	}
	if mapping.InterfaceHash != "" {
		if err := ValidateInterfaceDescriptorHashFormatV2(mapping.InterfaceHash); err != nil {
			return err
		}
	}
	if mapping.ProofHeight == 0 || mapping.RecordVersion == 0 {
		return errors.New("routing integration contract mapping height and version are required")
	}
	if mapping.MappingHash == "" || mapping.MappingHash != ComputeContractInvocationMappingHashV2(mapping) {
		return errors.New("routing integration contract mapping hash mismatch")
	}
	return nil
}

func ValidateServiceMappingV2(mapping ServiceMappingV2) error {
	if _, err := NormalizeAETDomain(mapping.Name); err != nil {
		return err
	}
	if err := validateServiceEndpointsV2([]ServiceEndpointV2{mapping.Endpoint}, mapping.Endpoint.TTL); err != nil {
		return err
	}
	if mapping.ProofHeight == 0 || mapping.RecordVersion == 0 {
		return errors.New("routing integration service mapping height and version are required")
	}
	if mapping.MappingHash == "" || mapping.MappingHash != ComputeServiceMappingHashV2(mapping) {
		return errors.New("routing integration service mapping hash mismatch")
	}
	return nil
}

func ValidateInterfaceMappingV2(mapping InterfaceMappingV2) error {
	if _, err := NormalizeAETDomain(mapping.Name); err != nil {
		return err
	}
	if err := validateUnifiedRecordKey("routing integration interface_id", mapping.InterfaceID); err != nil {
		return err
	}
	if err := ValidateInterfaceDescriptorHashFormatV2(mapping.SchemaHash); err != nil {
		return err
	}
	if !mapping.SchemaHashVerified {
		return errors.New("routing integration interface mapping requires verified schema hash")
	}
	if !mapping.UserConfirmationRequired {
		return errors.New("routing integration interface mapping requires user confirmation")
	}
	if mapping.ProofHeight == 0 || mapping.RecordVersion == 0 {
		return errors.New("routing integration interface mapping height and version are required")
	}
	if mapping.MappingHash == "" || mapping.MappingHash != ComputeInterfaceMappingHashV2(mapping) {
		return errors.New("routing integration interface mapping hash mismatch")
	}
	return nil
}

func ComputeRoutingIntegrationModuleBreakdownHashV2(breakdown RoutingIntegrationModuleBreakdownV2) string {
	breakdown = canonicalRoutingIntegrationModuleBreakdownV2(breakdown)
	parts := []string{"aetra-routing-integration-module-breakdown-v2", breakdown.ModulePath}
	parts = appendBreakdownStringsV2(parts, "purpose", breakdown.Purpose)
	for _, value := range breakdown.StateObjects {
		parts = append(parts, "state", string(value))
	}
	for _, value := range breakdown.Messages {
		parts = append(parts, "message", string(value))
	}
	for _, value := range breakdown.Queries {
		parts = append(parts, "query", string(value))
	}
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, "failure", string(failure.Mode), failure.Guard, failure.StoreScope)
	}
	for _, value := range breakdown.IntegrationPoints {
		parts = append(parts, "integration", string(value))
	}
	parts = appendBreakdownStringsV2(parts, "primitive", breakdown.BackingPrimitives)
	parts = appendBreakdownStringsV2(parts, "store", breakdown.StoreKeys)
	return identityHash(parts...)
}

func ComputeIdentityTransactionMappingHashV2(mapping IdentityTransactionMappingV2) string {
	return identityHash("routing-integration-transaction-mapping-v2", mapping.Name, string(mapping.Address), fmt.Sprint(mapping.ProofVerified), fmt.Sprint(mapping.ProofHeight), fmt.Sprint(mapping.RecordVersion), fmt.Sprint(mapping.FreshUntilHeight), mapping.AuditMemo)
}

func ComputeContractInvocationMappingHashV2(mapping ContractInvocationMappingV2) string {
	return identityHash("routing-integration-contract-mapping-v2", mapping.Name, mapping.TargetID, mapping.InterfaceID, mapping.InterfaceHash, string(mapping.ContractAddress), mapping.Entrypoint, fmt.Sprint(mapping.ProofVerified), fmt.Sprint(mapping.ProofHeight), fmt.Sprint(mapping.RecordVersion), fmt.Sprint(mapping.RequiresConfirm), fmt.Sprint(mapping.SimulationNeeded))
}

func ComputeServiceMappingHashV2(mapping ServiceMappingV2) string {
	parts := []string{"routing-integration-service-mapping-v2", mapping.Name, mapping.ServiceID, mapping.Endpoint.Endpoint, fmt.Sprint(mapping.ProofVerified), fmt.Sprint(mapping.ProofHeight), fmt.Sprint(mapping.RecordVersion), fmt.Sprint(mapping.EndpointAvailabilityConsensusGuaranteed)}
	for _, endpoint := range mapping.FallbackEndpoints {
		parts = append(parts, endpoint.ServiceID, endpoint.Endpoint)
	}
	return identityHash(parts...)
}

func ComputeInterfaceMappingHashV2(mapping InterfaceMappingV2) string {
	return identityHash("routing-integration-interface-mapping-v2", mapping.Name, mapping.InterfaceID, mapping.SchemaHash, fmt.Sprint(mapping.SchemaHashVerified), fmt.Sprint(mapping.RenderPolicySupported), fmt.Sprint(mapping.UserConfirmationRequired), fmt.Sprint(mapping.ExecutionTargetImmutable), fmt.Sprint(mapping.ProofVerified), fmt.Sprint(mapping.ProofHeight), fmt.Sprint(mapping.RecordVersion))
}

func ComputeRoutingIntegrationResolvedTargetHashV2(target RoutingIntegrationResolvedExecutionTargetV2) string {
	return identityHash("routing-integration-resolved-target-v2", target.Name, string(target.TargetType), target.TargetKey, string(target.Address), target.Endpoint, target.Descriptor, target.Route.RouteID, target.Route.TargetType, target.Route.PreferredTarget, fmt.Sprint(target.RecordVersion), fmt.Sprint(target.ProofHeight), fmt.Sprint(target.ProofVerified), fmt.Sprint(target.FreshUntilHeight))
}

func ComputeRoutingIntegrationWalletSDKHelperHashV2(helper RoutingIntegrationWalletSDKHelperV2) string {
	parts := []string{"routing-integration-wallet-sdk-helper-v2", fmt.Sprint(helper.ProofRequired), fmt.Sprint(helper.UserConfirmation), fmt.Sprint(helper.AdvisoryOnly)}
	if helper.TransactionMapping != nil {
		parts = append(parts, helper.TransactionMapping.MappingHash)
	}
	if helper.ContractMapping != nil {
		parts = append(parts, helper.ContractMapping.MappingHash)
	}
	if helper.ServiceMapping != nil {
		parts = append(parts, helper.ServiceMapping.MappingHash)
	}
	if helper.InterfaceMapping != nil {
		parts = append(parts, helper.InterfaceMapping.MappingHash)
	}
	return identityHash(parts...)
}

func IsRoutingIntegrationStateObjectV2(value RoutingIntegrationStateObjectV2) bool {
	switch value {
	case RoutingIntegrationStateRoutingPolicy, RoutingIntegrationStateIdentityTransactionMapping, RoutingIntegrationStateContractInvocationMapping, RoutingIntegrationStateServiceMapping, RoutingIntegrationStateInterfaceMapping, RoutingIntegrationStateExecutionHintPolicy:
		return true
	default:
		return false
	}
}

func IsRoutingIntegrationMessageNameV2(value RoutingIntegrationMessageNameV2) bool {
	switch value {
	case RoutingIntegrationMsgSetRoutingPolicy, RoutingIntegrationMsgUpdateExecutionHints, RoutingIntegrationMsgRegisterInterfaceDescriptor, RoutingIntegrationMsgRegisterServiceEndpoint, RoutingIntegrationMsgClearRoutingMetadata:
		return true
	default:
		return false
	}
}

func IsRoutingIntegrationQueryNameV2(value RoutingIntegrationQueryNameV2) bool {
	switch value {
	case RoutingIntegrationQueryTransactionMapping, RoutingIntegrationQueryContractInvocationMapping, RoutingIntegrationQueryServiceMapping, RoutingIntegrationQueryInterfaceMapping, RoutingIntegrationQueryExecutionHints, RoutingIntegrationQueryResolvedExecutionTarget:
		return true
	default:
		return false
	}
}

func IsRoutingIntegrationFailureModeV2(value RoutingIntegrationFailureModeV2) bool {
	switch value {
	case RoutingIntegrationFailureStaleMetadata, RoutingIntegrationFailureWrongInterfaceTarget, RoutingIntegrationFailureServiceEndpointUnavailable, RoutingIntegrationFailureExecutionHintConflict, RoutingIntegrationFailureAdvisoryAsAuthorization:
		return true
	default:
		return false
	}
}

func IsRoutingIntegrationIntegrationPointV2(value RoutingIntegrationIntegrationPointV2) bool {
	switch value {
	case RoutingIntegrationIntegrationResolverModule, RoutingIntegrationIntegrationFeeModule, RoutingIntegrationIntegrationContractExecution, RoutingIntegrationIntegrationWalletSDK, RoutingIntegrationIntegrationServiceClients:
		return true
	default:
		return false
	}
}

func requiredRoutingIntegrationStateObjectsV2() []RoutingIntegrationStateObjectV2 {
	return []RoutingIntegrationStateObjectV2{RoutingIntegrationStateContractInvocationMapping, RoutingIntegrationStateExecutionHintPolicy, RoutingIntegrationStateIdentityTransactionMapping, RoutingIntegrationStateInterfaceMapping, RoutingIntegrationStateRoutingPolicy, RoutingIntegrationStateServiceMapping}
}

func requiredRoutingIntegrationMessagesV2() []RoutingIntegrationMessageNameV2 {
	return []RoutingIntegrationMessageNameV2{RoutingIntegrationMsgClearRoutingMetadata, RoutingIntegrationMsgRegisterInterfaceDescriptor, RoutingIntegrationMsgRegisterServiceEndpoint, RoutingIntegrationMsgSetRoutingPolicy, RoutingIntegrationMsgUpdateExecutionHints}
}

func requiredRoutingIntegrationQueriesV2() []RoutingIntegrationQueryNameV2 {
	return []RoutingIntegrationQueryNameV2{RoutingIntegrationQueryContractInvocationMapping, RoutingIntegrationQueryExecutionHints, RoutingIntegrationQueryInterfaceMapping, RoutingIntegrationQueryResolvedExecutionTarget, RoutingIntegrationQueryServiceMapping, RoutingIntegrationQueryTransactionMapping}
}

func requiredRoutingIntegrationFailuresV2() []RoutingIntegrationFailureModeV2 {
	return []RoutingIntegrationFailureModeV2{RoutingIntegrationFailureAdvisoryAsAuthorization, RoutingIntegrationFailureExecutionHintConflict, RoutingIntegrationFailureServiceEndpointUnavailable, RoutingIntegrationFailureStaleMetadata, RoutingIntegrationFailureWrongInterfaceTarget}
}

func requiredRoutingIntegrationIntegrationPointsV2() []RoutingIntegrationIntegrationPointV2 {
	return []RoutingIntegrationIntegrationPointV2{RoutingIntegrationIntegrationContractExecution, RoutingIntegrationIntegrationFeeModule, RoutingIntegrationIntegrationResolverModule, RoutingIntegrationIntegrationServiceClients, RoutingIntegrationIntegrationWalletSDK}
}

func canonicalRoutingIntegrationModuleBreakdownV2(breakdown RoutingIntegrationModuleBreakdownV2) RoutingIntegrationModuleBreakdownV2 {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedBreakdownStringsV2(breakdown.Purpose)
	breakdown.StateObjects = sortedBreakdownTypedStringsV2(breakdown.StateObjects)
	breakdown.Messages = sortedBreakdownTypedStringsV2(breakdown.Messages)
	breakdown.Queries = sortedBreakdownTypedStringsV2(breakdown.Queries)
	breakdown.FailureModes = sortedRoutingIntegrationFailuresV2(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedBreakdownTypedStringsV2(breakdown.IntegrationPoints)
	breakdown.BackingPrimitives = sortedBreakdownStringsV2(breakdown.BackingPrimitives)
	breakdown.StoreKeys = sortedBreakdownStringsV2(breakdown.StoreKeys)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func validateRoutingIntegrationFailuresV2(values []RoutingIntegrationFailureCoverageV2) error {
	if len(values) != len(requiredRoutingIntegrationFailuresV2()) {
		return fmt.Errorf("routing integration expected %d failure modes", len(requiredRoutingIntegrationFailuresV2()))
	}
	seen := map[RoutingIntegrationFailureModeV2]struct{}{}
	for _, value := range values {
		if !IsRoutingIntegrationFailureModeV2(value.Mode) {
			return fmt.Errorf("routing integration unknown failure mode %q", value.Mode)
		}
		if strings.TrimSpace(value.Guard) == "" || !strings.HasPrefix(value.StoreScope, IdentityStoreV2Prefix+"/") {
			return fmt.Errorf("routing integration failure %s has invalid guard or store scope", value.Mode)
		}
		if _, found := seen[value.Mode]; found {
			return fmt.Errorf("routing integration duplicate failure mode %s", value.Mode)
		}
		seen[value.Mode] = struct{}{}
	}
	for _, required := range requiredRoutingIntegrationFailuresV2() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("routing integration missing failure mode %s", required)
		}
	}
	return nil
}

func sortedRoutingIntegrationFailuresV2(values []RoutingIntegrationFailureCoverageV2) []RoutingIntegrationFailureCoverageV2 {
	out := append([]RoutingIntegrationFailureCoverageV2(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].StoreScope = strings.TrimSpace(out[i].StoreScope)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func routingIntegrationResolvedTargetFromTransactionV2(mapping IdentityTransactionMappingV2) RoutingIntegrationResolvedExecutionTargetV2 {
	target := RoutingIntegrationResolvedExecutionTargetV2{Name: mapping.Name, TargetType: IdentityResolutionTargetPrimary, TargetKey: ResolverKeyPrimary, Address: cloneSpecAddress(mapping.Address), RecordVersion: mapping.RecordVersion, ProofHeight: mapping.ProofHeight, ProofVerified: mapping.ProofVerified, FreshUntilHeight: mapping.FreshUntilHeight}
	target.ResultHash = ComputeRoutingIntegrationResolvedTargetHashV2(target)
	return target
}

func routingIntegrationResolvedTargetFromContractV2(mapping ContractInvocationMappingV2) RoutingIntegrationResolvedExecutionTargetV2 {
	target := RoutingIntegrationResolvedExecutionTargetV2{Name: mapping.Name, TargetType: IdentityResolutionTargetContract, TargetKey: mapping.TargetID, Address: cloneSpecAddress(mapping.ContractAddress), Descriptor: mapping.InterfaceHash, RecordVersion: mapping.RecordVersion, ProofHeight: mapping.ProofHeight, ProofVerified: mapping.ProofVerified}
	target.Route = RoutingMetadataV2{Entrypoint: mapping.Entrypoint, TargetType: string(IdentityResolutionTargetContract), PreferredTarget: mapping.TargetID}
	target.ResultHash = ComputeRoutingIntegrationResolvedTargetHashV2(target)
	return target
}

func routingIntegrationResolvedTargetFromServiceV2(mapping ServiceMappingV2) RoutingIntegrationResolvedExecutionTargetV2 {
	target := RoutingIntegrationResolvedExecutionTargetV2{Name: mapping.Name, TargetType: IdentityResolutionTargetService, TargetKey: mapping.ServiceID, Endpoint: mapping.Endpoint.Endpoint, RecordVersion: mapping.RecordVersion, ProofHeight: mapping.ProofHeight, ProofVerified: mapping.ProofVerified}
	target.ResultHash = ComputeRoutingIntegrationResolvedTargetHashV2(target)
	return target
}

func routingIntegrationResolvedTargetFromInterfaceV2(mapping InterfaceMappingV2) RoutingIntegrationResolvedExecutionTargetV2 {
	target := RoutingIntegrationResolvedExecutionTargetV2{Name: mapping.Name, TargetType: IdentityResolutionTargetInterface, TargetKey: mapping.InterfaceID, Descriptor: mapping.SchemaHash, RecordVersion: mapping.RecordVersion, ProofHeight: mapping.ProofHeight, ProofVerified: mapping.ProofVerified}
	target.ResultHash = ComputeRoutingIntegrationResolvedTargetHashV2(target)
	return target
}
