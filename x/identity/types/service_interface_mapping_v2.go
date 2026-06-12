package types

import (
	"errors"
	"sort"
)

const (
	IdentityServiceTypeRPCV1	= "rpc.v1"
	IdentityServiceTypeIndexerV1	= "indexer.v1"
	IdentityServiceTypeGatewayV1	= "gateway.v1"
	IdentityServiceTypeGenericV1	= "service.v1"
	IdentityRenderPolicyConfirmV2	= "wallet_confirm"
	IdentityRenderPolicyReadOnlyV2	= "read_only"
	IdentityRenderPolicyFormV1	= "form.v1"
)

type IdentityServiceEndpointTypeV2 struct {
	ServiceType		string
	SupportedTransports	[]string
	AllowedAuthPolicies	[]string
	SchemaHashOptional	string
}

type IdentityServiceDiscoveryRequestV2 struct {
	Name			string
	ServiceID		string
	SupportedTransports	[]string
	AllowedAuthPolicies	[]string
	SupportedServiceTypes	[]string
	ExternalMetadata	string
	State			IdentityState
	Height			uint64
	RecordTTL		uint64
	CurrentHeight		uint64
	FreshnessThreshold	uint64

	ExpectedChainID	string
	TrustedHeader	IdentityTrustedHeaderV2
	Proof		*IdentityResolutionProofFormatV2
}

type IdentityServiceDiscoveryResultV2 struct {
	Endpoint				ServiceEndpointV2
	FallbackEndpoints			[]ServiceEndpointV2
	DisplayPolicy				IdentityServiceEndpointDisplayPolicyV2
	TypeRegistryMatched			bool
	MetadataHashVerified			bool
	ProofVerified				bool
	ProofHeight				uint64
	RecordVersion				uint64
	FreshUntilHeight			uint64
	FreshnessWarning			bool
	EndpointTTLRespected			bool
	EndpointAvailabilityConsensusGuaranteed	bool
}

type IdentityServiceEndpointDisplayPolicyV2 struct {
	DisplayEndpoint			bool
	DisplayAsVerifiedService	bool
	DisplayAsVerifiedOwnership	bool
	DisplayMetadataAsOwnership	bool
	MetadataHashVerified		bool
	ProofVerified			bool
	EndpointAvailabilityAdvisory	bool
	OwnershipVerificationSource	string
	UserFacingVerificationWarning	string
}

type IdentityInterfaceWalletPolicyV2 struct {
	SupportedRenderPolicies	[]string
	RequireUserConfirmation	bool
	AllowExternalSchemas	bool
	AllowInlineSchemas	bool
}

type IdentityInterfaceSchemaRequestV2 struct {
	Name			string
	InterfaceID		string
	ExpectedSchemaHash	string
	ExternalSchema		string
	WalletPolicy		IdentityInterfaceWalletPolicyV2
	State			IdentityState
	Height			uint64
	RecordTTL		uint64
	CurrentHeight		uint64
	FreshnessThreshold	uint64

	ExpectedChainID	string
	TrustedHeader	IdentityTrustedHeaderV2
	Proof		*IdentityResolutionProofFormatV2
}

type IdentityInterfaceSchemaResultV2 struct {
	Descriptor			InterfaceDescriptorV2
	SchemaHash			string
	SchemaSource			string
	SchemaHashVerified		bool
	ExternalSchemaHashVerified	bool
	RenderPolicySupported		bool
	UserConfirmationRequired	bool
	ExecutionTargetImmutable	bool
	ProofVerified			bool
	ProofHeight			uint64
	RecordVersion			uint64
	FreshUntilHeight		uint64
	FreshnessWarning		bool
}

func DefaultIdentityServiceEndpointTypeRegistryV2() []IdentityServiceEndpointTypeV2 {
	return []IdentityServiceEndpointTypeV2{
		{ServiceType: IdentityServiceTypeRPCV1, SupportedTransports: []string{"https", "grpcs"}, AllowedAuthPolicies: []string{"none", "token"}},
		{ServiceType: IdentityServiceTypeIndexerV1, SupportedTransports: []string{"https"}, AllowedAuthPolicies: []string{"none", "token"}},
		{ServiceType: IdentityServiceTypeGatewayV1, SupportedTransports: []string{"https", "wss"}, AllowedAuthPolicies: []string{"none", "token"}},
		{ServiceType: IdentityServiceTypeGenericV1, SupportedTransports: []string{"https", "grpcs", "wss", "aetra", "ipfs"}, AllowedAuthPolicies: []string{"none", "token"}},
	}
}

func DefaultIdentityInterfaceWalletPolicyV2() IdentityInterfaceWalletPolicyV2 {
	return IdentityInterfaceWalletPolicyV2{
		SupportedRenderPolicies:	[]string{IdentityRenderPolicyConfirmV2, IdentityRenderPolicyReadOnlyV2, IdentityRenderPolicyFormV1},
		RequireUserConfirmation:	true,
		AllowExternalSchemas:		true,
		AllowInlineSchemas:		true,
	}
}

func BuildIdentityServiceDiscoveryV2(request IdentityServiceDiscoveryRequestV2) (IdentityServiceDiscoveryResultV2, error) {
	record, proofVerified, proofHeight, recordVersion, freshUntil, err := unifiedRecordForServiceInterfaceRequestV2(
		request.Name, request.State, request.Height, request.RecordTTL, request.CurrentHeight,
		request.ExpectedChainID, request.TrustedHeader, request.Proof,
		IdentityResolutionTargetService, request.ServiceID,
	)
	if err != nil {
		return IdentityServiceDiscoveryResultV2{}, err
	}
	endpoints := compatibleServiceEndpointsV2(record.ServiceEndpoints, request)
	if len(endpoints) == 0 {
		return IdentityServiceDiscoveryResultV2{}, errors.New("identity v2 service discovery failed closed: no compatible verified endpoint")
	}
	selected := endpoints[0]
	metadataVerified, err := verifyServiceExternalMetadataV2(selected, request.ExternalMetadata)
	if err != nil {
		return IdentityServiceDiscoveryResultV2{}, err
	}
	fallbacks := append([]ServiceEndpointV2(nil), endpoints[1:]...)
	displayPolicy := BuildIdentityServiceEndpointDisplayPolicyV2(proofVerified, metadataVerified)
	return IdentityServiceDiscoveryResultV2{
		Endpoint:					selected,
		FallbackEndpoints:				fallbacks,
		DisplayPolicy:					displayPolicy,
		TypeRegistryMatched:				serviceEndpointMatchesRegistryV2(selected, DefaultIdentityServiceEndpointTypeRegistryV2()),
		MetadataHashVerified:				metadataVerified,
		ProofVerified:					proofVerified,
		ProofHeight:					proofHeight,
		RecordVersion:					recordVersion,
		FreshUntilHeight:				freshUntil,
		FreshnessWarning:				EvaluateIdentityStaleProofWarningV2(IdentityStaleProofPolicyV2{CurrentHeight: request.CurrentHeight, ProofHeight: proofHeight, FreshUntilHeight: freshUntil, FreshnessThreshold: request.FreshnessThreshold}),
		EndpointTTLRespected:				endpointTTLRespectedV2(selected, request.CurrentHeight, proofHeight),
		EndpointAvailabilityConsensusGuaranteed:	false,
	}, nil
}

func BuildIdentityServiceEndpointDisplayPolicyV2(proofVerified bool, metadataHashVerified bool) IdentityServiceEndpointDisplayPolicyV2 {
	return IdentityServiceEndpointDisplayPolicyV2{
		DisplayEndpoint:		true,
		DisplayAsVerifiedService:	proofVerified,
		DisplayAsVerifiedOwnership:	false,
		DisplayMetadataAsOwnership:	false,
		MetadataHashVerified:		metadataHashVerified,
		ProofVerified:			proofVerified,
		EndpointAvailabilityAdvisory:	true,
		OwnershipVerificationSource:	"registry_nft_binding",
		UserFacingVerificationWarning:	"service endpoint metadata is not verified ownership",
	}
}

func ValidateIdentityServiceEndpointDisplayPolicyV2(policy IdentityServiceEndpointDisplayPolicyV2) error {
	if policy.DisplayAsVerifiedOwnership {
		return errors.New("identity v2 service endpoint metadata must not be displayed as verified ownership")
	}
	if policy.DisplayMetadataAsOwnership {
		return errors.New("identity v2 service endpoint metadata cannot be used as ownership evidence")
	}
	if policy.OwnershipVerificationSource != "" && policy.OwnershipVerificationSource != "registry_nft_binding" {
		return errors.New("identity v2 verified ownership display requires registry nft binding")
	}
	return nil
}

func BuildIdentityInterfaceSchemaMappingV2(request IdentityInterfaceSchemaRequestV2) (IdentityInterfaceSchemaResultV2, error) {
	policy := request.WalletPolicy
	if len(policy.SupportedRenderPolicies) == 0 {
		policy = DefaultIdentityInterfaceWalletPolicyV2()
	}
	if !policy.RequireUserConfirmation {
		return IdentityInterfaceSchemaResultV2{}, errors.New("identity v2 interface schema mapping requires explicit user confirmation")
	}
	record, proofVerified, proofHeight, recordVersion, freshUntil, err := unifiedRecordForServiceInterfaceRequestV2(
		request.Name, request.State, request.Height, request.RecordTTL, request.CurrentHeight,
		request.ExpectedChainID, request.TrustedHeader, request.Proof,
		IdentityResolutionTargetInterface, request.InterfaceID,
	)
	if err != nil {
		return IdentityInterfaceSchemaResultV2{}, err
	}
	descriptor, err := VerifyIdentityInterfaceDescriptorForInvokeV2(record, request.InterfaceID, request.ExpectedSchemaHash)
	if err != nil {
		return IdentityInterfaceSchemaResultV2{}, err
	}
	if descriptor == nil {
		return IdentityInterfaceSchemaResultV2{}, errors.New("identity v2 interface descriptor is required")
	}
	schemaHash := interfaceDescriptorSchemaHashV2(*descriptor)
	source, externalVerified, err := resolveInterfaceSchemaSourceV2(*descriptor, request.ExternalSchema, policy)
	if err != nil {
		return IdentityInterfaceSchemaResultV2{}, err
	}
	return IdentityInterfaceSchemaResultV2{
		Descriptor:			*descriptor,
		SchemaHash:			schemaHash,
		SchemaSource:			source,
		SchemaHashVerified:		true,
		ExternalSchemaHashVerified:	externalVerified,
		RenderPolicySupported:		IdentityInterfaceRenderPolicySupportedV2(*descriptor, policy),
		UserConfirmationRequired:	policy.RequireUserConfirmation,
		ExecutionTargetImmutable:	true,
		ProofVerified:			proofVerified,
		ProofHeight:			proofHeight,
		RecordVersion:			recordVersion,
		FreshUntilHeight:		freshUntil,
		FreshnessWarning:		EvaluateIdentityStaleProofWarningV2(IdentityStaleProofPolicyV2{CurrentHeight: request.CurrentHeight, ProofHeight: proofHeight, FreshUntilHeight: freshUntil, FreshnessThreshold: request.FreshnessThreshold}),
	}, nil
}

func IdentityInterfaceRenderPolicySupportedV2(descriptor InterfaceDescriptorV2, policy IdentityInterfaceWalletPolicyV2) bool {
	if len(policy.SupportedRenderPolicies) == 0 {
		policy = DefaultIdentityInterfaceWalletPolicyV2()
	}
	return stringInSetV2(descriptor.RenderPolicy, policy.SupportedRenderPolicies)
}

func unifiedRecordForServiceInterfaceRequestV2(name string, state IdentityState, height uint64, ttl uint64, currentHeight uint64, expectedChainID string, trustedHeader IdentityTrustedHeaderV2, proof *IdentityResolutionProofFormatV2, targetType IdentityResolutionTargetTypeV2, targetKey string) (UnifiedResolutionRecordV2, bool, uint64, uint64, uint64, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return UnifiedResolutionRecordV2{}, false, 0, 0, 0, err
	}
	if ttl == 0 {
		ttl = 1
	}
	if proof != nil {
		target, err := VerifyIdentityResolutionProofLightClientV2(IdentityLightClientVerificationRequestV2{
			ExpectedChainID:	expectedChainID,
			RequestedName:		normalized,
			TrustedHeader:		trustedHeader,
			Proof:			*proof,
			TargetType:		targetType,
			TargetKey:		targetKey,
			CurrentHeight:		currentHeight,
			AllowRenewalWindow:	true,
			NormalizationVersion:	NameNormalizationVersionV2,
		})
		if err != nil {
			return UnifiedResolutionRecordV2{}, false, 0, 0, 0, err
		}
		if proof.ResolverRecord == nil {
			return UnifiedResolutionRecordV2{}, false, 0, 0, 0, errors.New("identity v2 proof resolver record is required")
		}
		return *proof.ResolverRecord, true, target.ProofHeight, target.RecordVersion, target.FreshUntilHeight, nil
	}
	record, err := BuildUnifiedResolutionRecordV2(state, normalized, height, ttl)
	if err != nil {
		return UnifiedResolutionRecordV2{}, false, 0, 0, 0, err
	}
	return record, false, height, record.RecordVersion, height + ttl, nil
}

func compatibleServiceEndpointsV2(endpoints []ServiceEndpointV2, request IdentityServiceDiscoveryRequestV2) []ServiceEndpointV2 {
	out := make([]ServiceEndpointV2, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if request.ServiceID != "" && serviceEndpointIDV2(endpoint) != request.ServiceID {
			continue
		}
		if len(request.SupportedTransports) > 0 && !stringInSetV2(endpoint.Transport, request.SupportedTransports) {
			continue
		}
		if len(request.AllowedAuthPolicies) > 0 && !stringInSetV2(endpoint.AuthPolicy, request.AllowedAuthPolicies) {
			continue
		}
		if len(request.SupportedServiceTypes) > 0 && !stringInSetV2(endpoint.ServiceType, request.SupportedServiceTypes) {
			continue
		}
		out = append(out, endpoint)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority < out[j].Priority
		}
		if out[i].Weight != out[j].Weight {
			return out[i].Weight > out[j].Weight
		}
		return serviceEndpointIDV2(out[i]) < serviceEndpointIDV2(out[j])
	})
	return out
}

func verifyServiceExternalMetadataV2(endpoint ServiceEndpointV2, metadata string) (bool, error) {
	if endpoint.SchemaHashOptional == "" {
		return false, nil
	}
	if metadata == "" {
		return false, errors.New("identity v2 service external metadata is required for schema hash verification")
	}
	hash, err := InterfaceDescriptorHashV2(metadata)
	if err != nil {
		return false, err
	}
	if hash != endpoint.SchemaHashOptional {
		return false, errors.New("identity v2 service external metadata hash mismatch")
	}
	return true, nil
}

func serviceEndpointMatchesRegistryV2(endpoint ServiceEndpointV2, registry []IdentityServiceEndpointTypeV2) bool {
	for _, entry := range registry {
		if entry.ServiceType != endpoint.ServiceType {
			continue
		}
		if !stringInSetV2(endpoint.Transport, entry.SupportedTransports) {
			return false
		}
		if endpoint.AuthPolicy != "" && !stringInSetV2(endpoint.AuthPolicy, entry.AllowedAuthPolicies) {
			return false
		}
		if entry.SchemaHashOptional != "" && endpoint.SchemaHashOptional != entry.SchemaHashOptional {
			return false
		}
		return true
	}
	return false
}

func endpointTTLRespectedV2(endpoint ServiceEndpointV2, currentHeight uint64, proofHeight uint64) bool {
	if endpoint.TTL == 0 || currentHeight == 0 || proofHeight == 0 {
		return false
	}
	return currentHeight <= proofHeight+endpoint.TTL
}

func resolveInterfaceSchemaSourceV2(descriptor InterfaceDescriptorV2, externalSchema string, policy IdentityInterfaceWalletPolicyV2) (string, bool, error) {
	schemaHash := interfaceDescriptorSchemaHashV2(descriptor)
	if descriptor.SchemaInlineOptional != "" {
		if !policy.AllowInlineSchemas {
			return "", false, errors.New("identity v2 wallet policy rejects inline interface schemas")
		}
		hash, err := InterfaceDescriptorHashV2(descriptor.SchemaInlineOptional)
		if err != nil {
			return "", false, err
		}
		if hash != schemaHash {
			return "", false, errors.New("identity v2 inline interface schema hash mismatch")
		}
		return descriptor.SchemaInlineOptional, false, nil
	}
	if descriptor.SchemaURIOptional != "" {
		if !policy.AllowExternalSchemas {
			return "", false, errors.New("identity v2 wallet policy rejects external interface schemas")
		}
		if externalSchema == "" {
			return "", false, errors.New("identity v2 external interface schema content is required")
		}
		hash, err := InterfaceDescriptorHashV2(externalSchema)
		if err != nil {
			return "", false, err
		}
		if hash != schemaHash {
			return "", false, errors.New("identity v2 external interface schema hash mismatch")
		}
		return externalSchema, true, nil
	}
	return "", false, errors.New("identity v2 interface descriptor has no schema source")
}

func stringInSetV2(value string, allowed []string) bool {
	for _, item := range allowed {
		if item == value {
			return true
		}
	}
	return false
}
