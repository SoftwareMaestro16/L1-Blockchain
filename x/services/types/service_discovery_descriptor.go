package types

import (
	"errors"
	"fmt"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	identitytypes "github.com/sovereign-l1/l1/x/identity/types"
)

type ServiceDiscoveryDescriptorV1 struct {
	ServiceID		string
	ServiceName		string
	Owner			string
	ServiceType		coretypes.ServiceType
	Endpoint		string
	InterfaceHash		string
	TrustModel		coretypes.ServiceTrustModel
	PaymentModel		string
	VerificationModel	coretypes.ServiceVerificationModel
	Status			coretypes.ServiceStatus
	ExpiryHeight		uint64
	ProofOptional		string
	SignatureOptional	string
	DescriptorHash		string
}

type AETServiceBinding struct {
	IdentityName	string
	ServiceID	string
	InterfaceHash	string
	Endpoint	string
	BindingHash	string
}

type AETServiceBindingProof struct {
	Binding			AETServiceBinding
	Descriptor		ServiceDiscoveryDescriptorV1
	RegistryDescriptor	ServiceDescriptor
	IdentityTrustedRoot	string
	IdentityProof		identitytypes.IdentityResolutionProof
	ServiceRegistryProof	ServiceRegistryProof
	Height			uint64
	ProofHash		string
}

type AETServiceBindingVerification struct {
	IdentityName		string
	ServiceID		string
	InterfaceHash		string
	Endpoint		string
	IdentityProofHash	string
	ServiceProofHash	string
	BindingHash		string
	DescriptorHash		string
	VerificationHash	string
}

func NewServiceDiscoveryDescriptorV1(descriptor ServiceDiscoveryDescriptorV1) (ServiceDiscoveryDescriptorV1, error) {
	if descriptor.DescriptorHash != "" {
		return ServiceDiscoveryDescriptorV1{}, errors.New("service discovery descriptor hash must be empty before construction")
	}
	descriptor = canonicalServiceDiscoveryDescriptorV1(descriptor)
	if err := descriptor.ValidateFormat(); err != nil {
		return ServiceDiscoveryDescriptorV1{}, err
	}
	descriptor.DescriptorHash = ComputeServiceDiscoveryDescriptorV1Hash(descriptor)
	return descriptor, descriptor.Validate()
}

func ProjectServiceDiscoveryDescriptorV1(resolution ServiceResolutionOutput) (ServiceDiscoveryDescriptorV1, error) {
	if err := resolution.Validate(); err != nil {
		return ServiceDiscoveryDescriptorV1{}, err
	}
	return NewServiceDiscoveryDescriptorV1(ServiceDiscoveryDescriptorV1{
		ServiceID:		resolution.ServiceID,
		ServiceName:		resolution.ServiceName,
		Owner:			resolution.Descriptor.Owner,
		ServiceType:		resolution.Descriptor.ServiceType,
		Endpoint:		resolution.Endpoint,
		InterfaceHash:		resolution.InterfaceHash,
		TrustModel:		resolution.TrustModel,
		PaymentModel:		resolution.PaymentModel,
		VerificationModel:	resolution.VerificationModel,
		Status:			resolution.Descriptor.Status,
		ExpiryHeight:		resolution.ExpiryHeight,
		ProofOptional:		resolution.ProofChain.RegistryProofHash,
		SignatureOptional:	firstString(resolution.ProofChain.SignatureHashes),
	})
}

func ProjectServiceDiscoveryDescriptorFromCore(descriptor ServiceDescriptor, serviceName string, proofHash, signatureHash string) (ServiceDiscoveryDescriptorV1, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceDiscoveryDescriptorV1{}, err
	}
	if serviceName == "" {
		serviceName = descriptor.Discovery.ServiceName
	}
	return NewServiceDiscoveryDescriptorV1(ServiceDiscoveryDescriptorV1{
		ServiceID:		descriptor.ServiceID,
		ServiceName:		serviceName,
		Owner:			descriptor.Owner,
		ServiceType:		descriptor.ServiceType,
		Endpoint:		resolverEndpointFromDescriptor(descriptor),
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		TrustModel:		descriptor.Verification.TrustModel,
		PaymentModel:		registryPaymentModelFromDescriptor(descriptor),
		VerificationModel:	descriptor.Verification.Model,
		Status:			descriptor.Status,
		ExpiryHeight:		resolverExpiryHeight(descriptor),
		ProofOptional:		proofHash,
		SignatureOptional:	signatureHash,
	})
}

func (descriptor ServiceDiscoveryDescriptorV1) ValidateFormat() error {
	descriptor = canonicalServiceDiscoveryDescriptorV1(descriptor)
	if err := validateInterfaceToken("service discovery descriptor service id", descriptor.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("service discovery descriptor service name", descriptor.ServiceName); err != nil {
		return err
	}
	if err := validateInterfaceToken("service discovery descriptor owner", descriptor.Owner); err != nil {
		return err
	}
	if !coretypes.IsServiceType(descriptor.ServiceType) {
		return fmt.Errorf("unknown service discovery descriptor service type %q", descriptor.ServiceType)
	}
	if descriptor.Endpoint == "" || strings.TrimSpace(descriptor.Endpoint) != descriptor.Endpoint {
		return errors.New("service discovery descriptor endpoint is required and must not have surrounding whitespace")
	}
	if err := coretypes.ValidateHash("service discovery descriptor interface hash", descriptor.InterfaceHash); err != nil {
		return err
	}
	if !coretypes.IsServiceTrustModel(descriptor.TrustModel) {
		return fmt.Errorf("unknown service discovery descriptor trust model %q", descriptor.TrustModel)
	}
	if descriptor.PaymentModel == "" || strings.TrimSpace(descriptor.PaymentModel) != descriptor.PaymentModel {
		return errors.New("service discovery descriptor payment model is required and must not have surrounding whitespace")
	}
	if !coretypes.IsServiceVerificationModel(descriptor.VerificationModel) {
		return fmt.Errorf("unknown service discovery descriptor verification model %q", descriptor.VerificationModel)
	}
	if !coretypes.IsServiceStatus(descriptor.Status) {
		return fmt.Errorf("unknown service discovery descriptor status %q", descriptor.Status)
	}
	if descriptor.ExpiryHeight == 0 {
		return errors.New("service discovery descriptor expiry height is required")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "service discovery descriptor proof", value: descriptor.ProofOptional},
		{name: "service discovery descriptor signature", value: descriptor.SignatureOptional},
		{name: "service discovery descriptor hash", value: descriptor.DescriptorHash},
	} {
		if item.value == "" {
			continue
		}
		if err := coretypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (descriptor ServiceDiscoveryDescriptorV1) Validate() error {
	descriptor = canonicalServiceDiscoveryDescriptorV1(descriptor)
	if err := descriptor.ValidateFormat(); err != nil {
		return err
	}
	if descriptor.DescriptorHash == "" {
		return errors.New("service discovery descriptor hash is required")
	}
	if expected := ComputeServiceDiscoveryDescriptorV1Hash(descriptor); descriptor.DescriptorHash != expected {
		return fmt.Errorf("service discovery descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func NewAETServiceBinding(binding AETServiceBinding) (AETServiceBinding, error) {
	if binding.BindingHash != "" {
		return AETServiceBinding{}, errors.New("aet service binding hash must be empty before construction")
	}
	binding = canonicalAETServiceBinding(binding)
	if err := binding.ValidateFormat(); err != nil {
		return AETServiceBinding{}, err
	}
	binding.BindingHash = ComputeAETServiceBindingHash(binding)
	return binding, binding.Validate()
}

func NewAETServiceBindingFromDescriptor(identityName string, descriptor ServiceDiscoveryDescriptorV1) (AETServiceBinding, error) {
	if err := descriptor.Validate(); err != nil {
		return AETServiceBinding{}, err
	}
	return NewAETServiceBinding(AETServiceBinding{
		IdentityName:	identityName,
		ServiceID:	descriptor.ServiceID,
		InterfaceHash:	descriptor.InterfaceHash,
		Endpoint:	descriptor.Endpoint,
	})
}

func (binding AETServiceBinding) ValidateFormat() error {
	binding = canonicalAETServiceBinding(binding)
	if err := identitytypes.ValidateResolverDomain(binding.IdentityName); err != nil {
		return err
	}
	if err := validateInterfaceToken("aet service binding service id", binding.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("aet service binding interface hash", binding.InterfaceHash); err != nil {
		return err
	}
	if binding.Endpoint == "" || strings.TrimSpace(binding.Endpoint) != binding.Endpoint {
		return errors.New("aet service binding endpoint is required and must not have surrounding whitespace")
	}
	if binding.BindingHash != "" {
		return coretypes.ValidateHash("aet service binding hash", binding.BindingHash)
	}
	return nil
}

func (binding AETServiceBinding) Validate() error {
	binding = canonicalAETServiceBinding(binding)
	if err := binding.ValidateFormat(); err != nil {
		return err
	}
	if binding.BindingHash == "" {
		return errors.New("aet service binding hash is required")
	}
	if expected := ComputeAETServiceBindingHash(binding); binding.BindingHash != expected {
		return fmt.Errorf("aet service binding hash mismatch: expected %s", expected)
	}
	return nil
}

func NewAETServiceBindingProof(proof AETServiceBindingProof) (AETServiceBindingProof, error) {
	if proof.ProofHash != "" {
		return AETServiceBindingProof{}, errors.New("aet service binding proof hash must be empty before construction")
	}
	proof = canonicalAETServiceBindingProof(proof)
	if err := proof.ValidateFormat(); err != nil {
		return AETServiceBindingProof{}, err
	}
	proof.ProofHash = ComputeAETServiceBindingProofHash(proof)
	return proof, proof.Validate()
}

func (proof AETServiceBindingProof) ValidateFormat() error {
	proof = canonicalAETServiceBindingProof(proof)
	if err := proof.Binding.Validate(); err != nil {
		return err
	}
	if err := proof.Descriptor.Validate(); err != nil {
		return err
	}
	proof.RegistryDescriptor = coretypes.CanonicalServiceDescriptor(proof.RegistryDescriptor)
	if err := proof.RegistryDescriptor.Validate(); err != nil {
		return err
	}
	if proof.Binding.ServiceID != proof.Descriptor.ServiceID ||
		proof.Binding.InterfaceHash != proof.Descriptor.InterfaceHash ||
		proof.Binding.Endpoint != proof.Descriptor.Endpoint {
		return errors.New("aet service binding descriptor mismatch")
	}
	if proof.RegistryDescriptor.ServiceID != proof.Descriptor.ServiceID ||
		proof.RegistryDescriptor.Interface.InterfaceHash != proof.Descriptor.InterfaceHash {
		return errors.New("aet service binding registry descriptor mismatch")
	}
	if err := coretypes.ValidateHash("aet identity trusted root", proof.IdentityTrustedRoot); err != nil {
		return err
	}
	if err := proof.ServiceRegistryProof.Validate(); err != nil {
		return err
	}
	if proof.Height == 0 {
		return errors.New("aet service binding proof height is required")
	}
	if proof.ProofHash != "" {
		return coretypes.ValidateHash("aet service binding proof hash", proof.ProofHash)
	}
	return nil
}

func (proof AETServiceBindingProof) Validate() error {
	proof = canonicalAETServiceBindingProof(proof)
	if err := proof.ValidateFormat(); err != nil {
		return err
	}
	if proof.ProofHash == "" {
		return errors.New("aet service binding proof hash is required")
	}
	if expected := ComputeAETServiceBindingProofHash(proof); proof.ProofHash != expected {
		return fmt.Errorf("aet service binding proof hash mismatch: expected %s", expected)
	}
	return nil
}

func VerifyAETServiceBinding(proof AETServiceBindingProof) (AETServiceBindingVerification, error) {
	if err := proof.Validate(); err != nil {
		return AETServiceBindingVerification{}, err
	}
	resolution, err := identitytypes.ValidateResolutionProofBoundary(proof.IdentityProof, proof.IdentityTrustedRoot, proof.Height)
	if err != nil {
		return AETServiceBindingVerification{}, fmt.Errorf("aet identity proof failed: %w", err)
	}
	if resolution.QueryDomain != proof.Binding.IdentityName {
		return AETServiceBindingVerification{}, errors.New("aet identity proof query mismatch")
	}
	if resolution.Record.ZoneEndpoint != proof.Binding.Endpoint {
		return AETServiceBindingVerification{}, errors.New("aet resolver endpoint mismatch")
	}
	if proof.ServiceRegistryProof.ServiceID != proof.Binding.ServiceID {
		return AETServiceBindingVerification{}, errors.New("aet service proof service id mismatch")
	}
	if proof.ServiceRegistryProof.InterfaceHash != proof.Binding.InterfaceHash {
		return AETServiceBindingVerification{}, errors.New("aet service proof interface hash mismatch")
	}
	if proof.ServiceRegistryProof.DescriptorHash != coretypes.ComputeServiceDescriptorHash(proof.RegistryDescriptor) {
		return AETServiceBindingVerification{}, errors.New("aet service proof descriptor hash mismatch")
	}
	verification := AETServiceBindingVerification{
		IdentityName:		proof.Binding.IdentityName,
		ServiceID:		proof.Binding.ServiceID,
		InterfaceHash:		proof.Binding.InterfaceHash,
		Endpoint:		proof.Binding.Endpoint,
		IdentityProofHash:	proof.IdentityTrustedRoot,
		ServiceProofHash:	proof.ServiceRegistryProof.ProofHash,
		BindingHash:		proof.Binding.BindingHash,
		DescriptorHash:		proof.Descriptor.DescriptorHash,
	}
	verification.VerificationHash = ComputeAETServiceBindingVerificationHash(verification)
	return verification, verification.Validate()
}

func VerifyAETServiceBindings(proofs []AETServiceBindingProof) ([]AETServiceBindingVerification, error) {
	if len(proofs) == 0 {
		return nil, errors.New("aet service bindings are required")
	}
	out := make([]AETServiceBindingVerification, 0, len(proofs))
	for _, proof := range proofs {
		verification, err := VerifyAETServiceBinding(proof)
		if err != nil {
			return nil, err
		}
		out = append(out, verification)
	}
	return out, nil
}

func (verification AETServiceBindingVerification) Validate() error {
	if err := identitytypes.ValidateResolverDomain(verification.IdentityName); err != nil {
		return err
	}
	if err := validateInterfaceToken("aet service binding verification service id", verification.ServiceID); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "aet service binding verification interface hash", value: verification.InterfaceHash},
		{name: "aet service binding verification identity proof hash", value: verification.IdentityProofHash},
		{name: "aet service binding verification service proof hash", value: verification.ServiceProofHash},
		{name: "aet service binding verification binding hash", value: verification.BindingHash},
		{name: "aet service binding verification descriptor hash", value: verification.DescriptorHash},
		{name: "aet service binding verification hash", value: verification.VerificationHash},
	} {
		if err := coretypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if verification.Endpoint == "" || strings.TrimSpace(verification.Endpoint) != verification.Endpoint {
		return errors.New("aet service binding verification endpoint is required and must not have surrounding whitespace")
	}
	if expected := ComputeAETServiceBindingVerificationHash(verification); verification.VerificationHash != expected {
		return fmt.Errorf("aet service binding verification hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceDiscoveryDescriptorV1Hash(descriptor ServiceDiscoveryDescriptorV1) string {
	descriptor = canonicalServiceDiscoveryDescriptorV1(descriptor)
	return servicesHashParts(
		"aetra-services-discovery-descriptor-v1",
		descriptor.ServiceID,
		descriptor.ServiceName,
		descriptor.Owner,
		string(descriptor.ServiceType),
		descriptor.Endpoint,
		descriptor.InterfaceHash,
		string(descriptor.TrustModel),
		descriptor.PaymentModel,
		string(descriptor.VerificationModel),
		string(descriptor.Status),
		fmt.Sprint(descriptor.ExpiryHeight),
		descriptor.ProofOptional,
		descriptor.SignatureOptional,
	)
}

func ComputeAETServiceBindingHash(binding AETServiceBinding) string {
	binding = canonicalAETServiceBinding(binding)
	return servicesHashParts(
		"aetra-services-aet-binding-v1",
		binding.IdentityName,
		binding.ServiceID,
		binding.InterfaceHash,
		binding.Endpoint,
	)
}

func ComputeAETServiceBindingProofHash(proof AETServiceBindingProof) string {
	proof = canonicalAETServiceBindingProof(proof)
	return servicesHashParts(
		"aetra-services-aet-binding-proof-v1",
		proof.Binding.BindingHash,
		proof.Descriptor.DescriptorHash,
		coretypes.ComputeServiceDescriptorHash(proof.RegistryDescriptor),
		proof.IdentityTrustedRoot,
		proof.IdentityProof.StateRoot,
		proof.IdentityProof.QueryDomain,
		proof.IdentityProof.ResolverDomain,
		proof.ServiceRegistryProof.ProofHash,
		fmt.Sprint(proof.Height),
	)
}

func ComputeAETServiceBindingVerificationHash(verification AETServiceBindingVerification) string {
	return servicesHashParts(
		"aetra-services-aet-binding-verification-v1",
		verification.IdentityName,
		verification.ServiceID,
		verification.InterfaceHash,
		verification.Endpoint,
		verification.IdentityProofHash,
		verification.ServiceProofHash,
		verification.BindingHash,
		verification.DescriptorHash,
	)
}

func canonicalServiceDiscoveryDescriptorV1(descriptor ServiceDiscoveryDescriptorV1) ServiceDiscoveryDescriptorV1 {
	descriptor.ServiceID = strings.TrimSpace(descriptor.ServiceID)
	descriptor.ServiceName = strings.TrimSpace(descriptor.ServiceName)
	descriptor.Owner = strings.TrimSpace(descriptor.Owner)
	descriptor.Endpoint = strings.TrimSpace(descriptor.Endpoint)
	descriptor.InterfaceHash = strings.ToLower(strings.TrimSpace(descriptor.InterfaceHash))
	descriptor.PaymentModel = strings.TrimSpace(descriptor.PaymentModel)
	descriptor.ProofOptional = strings.ToLower(strings.TrimSpace(descriptor.ProofOptional))
	descriptor.SignatureOptional = strings.ToLower(strings.TrimSpace(descriptor.SignatureOptional))
	descriptor.DescriptorHash = strings.ToLower(strings.TrimSpace(descriptor.DescriptorHash))
	return descriptor
}

func canonicalAETServiceBinding(binding AETServiceBinding) AETServiceBinding {
	binding.IdentityName = strings.ToLower(strings.TrimSpace(binding.IdentityName))
	binding.ServiceID = strings.TrimSpace(binding.ServiceID)
	binding.InterfaceHash = strings.ToLower(strings.TrimSpace(binding.InterfaceHash))
	binding.Endpoint = strings.TrimSpace(binding.Endpoint)
	binding.BindingHash = strings.ToLower(strings.TrimSpace(binding.BindingHash))
	return binding
}

func canonicalAETServiceBindingProof(proof AETServiceBindingProof) AETServiceBindingProof {
	proof.Binding = canonicalAETServiceBinding(proof.Binding)
	proof.Descriptor = canonicalServiceDiscoveryDescriptorV1(proof.Descriptor)
	proof.IdentityTrustedRoot = strings.ToLower(strings.TrimSpace(proof.IdentityTrustedRoot))
	proof.ProofHash = strings.ToLower(strings.TrimSpace(proof.ProofHash))
	return proof
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
