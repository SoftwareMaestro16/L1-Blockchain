package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceResolutionSource string

const (
	ServiceResolutionOnChainRegistry	ServiceResolutionSource	= "on_chain_registry"
	ServiceResolutionOffChainIndex		ServiceResolutionSource	= "off_chain_index"
	ServiceResolutionSignedCache		ServiceResolutionSource	= "signed_cache"
	ServiceResolutionIdentityRecord		ServiceResolutionSource	= "aet_identity_record"
	ServiceResolutionDistributedMesh	ServiceResolutionSource	= "distributed_service_mesh"
)

type ServiceResolverInput struct {
	ServiceName		string
	Registry		ServiceRegistryState
	Distributed		DistributedServiceDiscoveryState
	OffChainIndex		[]ServiceResolverSourceRecord
	CachedSigned		[]ServiceResolverSourceRecord
	IdentityBindings	[]IdentityServiceBinding
	ResolutionHeight	uint64
	RequireFreshProof	bool
}

type ServiceResolverSourceRecord struct {
	Source			ServiceResolutionSource
	ServiceName		string
	ServiceID		string
	Descriptor		ServiceDescriptor
	Endpoint		string
	InterfaceDescriptor	ServiceInterface
	TrustMetadata		string
	PaymentModel		string
	VerificationModel	coretypes.ServiceVerificationModel
	ExpiryHeight		uint64
	ProofHash		string
	SignatureHash		string
	IdentityBinding		IdentityServiceBinding
	DistributedRecord	DistributedServiceRecord
	DistributedEndpoint	DistributedServiceEndpoint
	DistributedInterface	DistributedInterfaceDescriptor
	SourceHash		string
}

type ServiceResolutionProofChain struct {
	SourceHashes			[]string
	RegistryProofHash		string
	SignatureHashes			[]string
	IdentityBindingHashes		[]string
	DistributedRecordHashes		[]string
	DistributedEndpointHashes	[]string
	DistributedInterfaceHashes	[]string
	DistributedCommitmentHashes	[]string
	ChainHash			string
}

type ServiceResolutionOutput struct {
	ServiceName		string
	ServiceID		string
	Source			ServiceResolutionSource
	Descriptor		ServiceDescriptor
	Endpoint		string
	InterfaceHash		string
	InterfaceDescriptor	ServiceInterface
	TrustModel		coretypes.ServiceTrustModel
	TrustMetadata		string
	PaymentModel		string
	VerificationModel	coretypes.ServiceVerificationModel
	ExpiryHeight		uint64
	ProofChain		ServiceResolutionProofChain
	ResolutionHash		string
}

func ResolveService(input ServiceResolverInput) (ServiceResolutionOutput, error) {
	input.ServiceName = strings.TrimSpace(input.ServiceName)
	if err := validateInterfaceToken("service resolver service name", input.ServiceName); err != nil {
		return ServiceResolutionOutput{}, err
	}
	if input.ResolutionHeight == 0 {
		return ServiceResolutionOutput{}, errors.New("service resolver resolution height must be positive")
	}
	if err := input.Registry.Validate(); err != nil {
		return ServiceResolutionOutput{}, err
	}

	candidates := make([]ServiceResolverSourceRecord, 0)
	onChain, err := resolveOnChainRegistryCandidate(input)
	if err != nil {
		return ServiceResolutionOutput{}, err
	}
	if onChain.ServiceID != "" {
		candidates = append(candidates, onChain)
	}

	identity, err := resolveIdentityCandidates(input)
	if err != nil {
		return ServiceResolutionOutput{}, err
	}
	candidates = append(candidates, identity...)

	cached, err := normalizeResolverSourceRecords(ServiceResolutionSignedCache, input.CachedSigned, input)
	if err != nil {
		return ServiceResolutionOutput{}, err
	}
	candidates = append(candidates, cached...)

	offChain, err := normalizeResolverSourceRecords(ServiceResolutionOffChainIndex, input.OffChainIndex, input)
	if err != nil {
		return ServiceResolutionOutput{}, err
	}
	candidates = append(candidates, offChain...)

	distributed, err := resolveDistributedCandidates(input)
	if err != nil {
		return ServiceResolutionOutput{}, err
	}
	candidates = append(candidates, distributed...)

	if len(candidates) == 0 {
		return ServiceResolutionOutput{}, fmt.Errorf("service %q not resolved", input.ServiceName)
	}
	sortResolverCandidates(candidates)
	out, err := buildServiceResolutionOutput(input, candidates[0])
	if err != nil {
		return ServiceResolutionOutput{}, err
	}
	return out, out.Validate()
}

func NewServiceResolverSourceRecord(record ServiceResolverSourceRecord) (ServiceResolverSourceRecord, error) {
	if record.SourceHash != "" {
		return ServiceResolverSourceRecord{}, errors.New("service resolver source hash must be empty before construction")
	}
	record = canonicalResolverSourceRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return ServiceResolverSourceRecord{}, err
	}
	record.SourceHash = ComputeServiceResolverSourceHash(record)
	return record, record.Validate()
}

func (record ServiceResolverSourceRecord) ValidateFormat() error {
	record = canonicalResolverSourceRecord(record)
	if !IsServiceResolutionSource(record.Source) {
		return fmt.Errorf("unknown service resolution source %q", record.Source)
	}
	if err := validateInterfaceToken("service resolver source service name", record.ServiceName); err != nil {
		return err
	}
	if record.ServiceID != "" {
		if err := validateInterfaceToken("service resolver source service id", record.ServiceID); err != nil {
			return err
		}
	}
	if record.Descriptor.ServiceID != "" {
		if err := record.Descriptor.Validate(); err != nil {
			return err
		}
		if record.ServiceID != "" && record.ServiceID != record.Descriptor.ServiceID {
			return errors.New("service resolver source service id mismatch")
		}
	}
	if record.Endpoint != "" && strings.TrimSpace(record.Endpoint) != record.Endpoint {
		return errors.New("service resolver source endpoint must not have surrounding whitespace")
	}
	if record.InterfaceDescriptor.InterfaceHash != "" {
		if err := record.InterfaceDescriptor.Validate(); err != nil {
			return err
		}
	}
	if record.VerificationModel != "" && !coretypes.IsServiceVerificationModel(record.VerificationModel) {
		return fmt.Errorf("unknown service resolver verification model %q", record.VerificationModel)
	}
	if record.ExpiryHeight != 0 && record.ExpiryHeight <= 1 {
		return errors.New("service resolver source expiry height must be greater than one when set")
	}
	if record.ProofHash != "" {
		if err := coretypes.ValidateHash("service resolver source proof hash", record.ProofHash); err != nil {
			return err
		}
	}
	if record.SignatureHash != "" {
		if err := coretypes.ValidateHash("service resolver source signature hash", record.SignatureHash); err != nil {
			return err
		}
	}
	if record.Source == ServiceResolutionSignedCache && record.SignatureHash == "" {
		return errors.New("signed cache resolution source requires signature hash")
	}
	if record.Source == ServiceResolutionOffChainIndex && record.ProofHash == "" && record.SignatureHash == "" {
		return errors.New("off-chain index resolution source requires proof or signature hash")
	}
	if record.IdentityBinding.BindingHash != "" {
		if err := record.IdentityBinding.Validate(); err != nil {
			return err
		}
	}
	if record.DistributedRecord.RecordHash != "" {
		if err := record.DistributedRecord.Validate(); err != nil {
			return err
		}
	}
	if record.DistributedEndpoint.CommitmentHash != "" {
		if err := record.DistributedEndpoint.Validate(); err != nil {
			return err
		}
	}
	if record.DistributedInterface.DescriptorHash != "" {
		if err := record.DistributedInterface.Validate(); err != nil {
			return err
		}
	}
	if record.SourceHash != "" {
		return coretypes.ValidateHash("service resolver source hash", record.SourceHash)
	}
	return nil
}

func (record ServiceResolverSourceRecord) Validate() error {
	record = canonicalResolverSourceRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return err
	}
	if record.SourceHash == "" {
		return errors.New("service resolver source hash is required")
	}
	if expected := ComputeServiceResolverSourceHash(record); record.SourceHash != expected {
		return fmt.Errorf("service resolver source hash mismatch: expected %s", expected)
	}
	return nil
}

func (chain ServiceResolutionProofChain) Validate() error {
	for _, item := range []struct {
		name	string
		hashes	[]string
	}{
		{name: "service resolver source hash", hashes: chain.SourceHashes},
		{name: "service resolver signature hash", hashes: chain.SignatureHashes},
		{name: "service resolver identity binding hash", hashes: chain.IdentityBindingHashes},
		{name: "service resolver distributed record hash", hashes: chain.DistributedRecordHashes},
		{name: "service resolver distributed endpoint hash", hashes: chain.DistributedEndpointHashes},
		{name: "service resolver distributed interface hash", hashes: chain.DistributedInterfaceHashes},
		{name: "service resolver distributed commitment hash", hashes: chain.DistributedCommitmentHashes},
	} {
		if err := validateHashList(item.name, item.hashes); err != nil {
			return err
		}
	}
	if chain.RegistryProofHash != "" {
		if err := coretypes.ValidateHash("service resolver registry proof hash", chain.RegistryProofHash); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("service resolver proof chain hash", chain.ChainHash); err != nil {
		return err
	}
	if expected := ComputeServiceResolutionProofChainHash(chain); chain.ChainHash != expected {
		return fmt.Errorf("service resolver proof chain hash mismatch: expected %s", expected)
	}
	return nil
}

func (out ServiceResolutionOutput) Validate() error {
	if err := validateInterfaceToken("service resolution service name", out.ServiceName); err != nil {
		return err
	}
	if err := validateInterfaceToken("service resolution service id", out.ServiceID); err != nil {
		return err
	}
	if !IsServiceResolutionSource(out.Source) {
		return fmt.Errorf("unknown service resolution source %q", out.Source)
	}
	if err := out.Descriptor.Validate(); err != nil {
		return err
	}
	if out.Descriptor.ServiceID != out.ServiceID {
		return errors.New("service resolution descriptor service id mismatch")
	}
	if out.InterfaceHash == "" || out.InterfaceHash != out.Descriptor.Interface.InterfaceHash {
		return errors.New("service resolution interface hash mismatch")
	}
	if err := out.InterfaceDescriptor.Validate(); err != nil {
		return err
	}
	if out.InterfaceDescriptor.InterfaceHash != out.InterfaceHash {
		return errors.New("service resolution interface descriptor hash mismatch")
	}
	if out.Endpoint == "" || strings.TrimSpace(out.Endpoint) != out.Endpoint {
		return errors.New("service resolution endpoint is required and must not have surrounding whitespace")
	}
	if !coretypes.IsServiceTrustModel(out.TrustModel) {
		return fmt.Errorf("unknown service resolution trust model %q", out.TrustModel)
	}
	if out.PaymentModel == "" {
		return errors.New("service resolution payment model is required")
	}
	if !coretypes.IsServiceVerificationModel(out.VerificationModel) {
		return fmt.Errorf("unknown service resolution verification model %q", out.VerificationModel)
	}
	if out.ExpiryHeight == 0 {
		return errors.New("service resolution expiry height is required")
	}
	if err := out.ProofChain.Validate(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("service resolution hash", out.ResolutionHash); err != nil {
		return err
	}
	if expected := ComputeServiceResolutionOutputHash(out); out.ResolutionHash != expected {
		return fmt.Errorf("service resolution hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceResolverSourceHash(record ServiceResolverSourceRecord) string {
	record = canonicalResolverSourceRecord(record)
	return servicesHashParts(
		"aetra-services-resolver-source-v1",
		string(record.Source),
		record.ServiceName,
		record.ServiceID,
		coretypes.ComputeServiceDescriptorHash(record.Descriptor),
		record.Endpoint,
		record.InterfaceDescriptor.InterfaceHash,
		record.TrustMetadata,
		record.PaymentModel,
		string(record.VerificationModel),
		fmt.Sprint(record.ExpiryHeight),
		record.ProofHash,
		record.SignatureHash,
		record.IdentityBinding.BindingHash,
		record.DistributedRecord.RecordHash,
		record.DistributedEndpoint.CommitmentHash,
		record.DistributedInterface.DescriptorHash,
	)
}

func ComputeServiceResolutionProofChainHash(chain ServiceResolutionProofChain) string {
	return servicesHashParts(
		"aetra-services-resolution-proof-chain-v1",
		strings.Join(sortedStrings(chain.SourceHashes), ","),
		chain.RegistryProofHash,
		strings.Join(sortedStrings(chain.SignatureHashes), ","),
		strings.Join(sortedStrings(chain.IdentityBindingHashes), ","),
		strings.Join(sortedStrings(chain.DistributedRecordHashes), ","),
		strings.Join(sortedStrings(chain.DistributedEndpointHashes), ","),
		strings.Join(sortedStrings(chain.DistributedInterfaceHashes), ","),
		strings.Join(sortedStrings(chain.DistributedCommitmentHashes), ","),
	)
}

func ComputeServiceResolutionOutputHash(out ServiceResolutionOutput) string {
	return servicesHashParts(
		"aetra-services-resolution-output-v1",
		out.ServiceName,
		out.ServiceID,
		string(out.Source),
		coretypes.ComputeServiceDescriptorHash(out.Descriptor),
		out.Endpoint,
		out.InterfaceHash,
		out.InterfaceDescriptor.InterfaceHash,
		string(out.TrustModel),
		out.TrustMetadata,
		out.PaymentModel,
		string(out.VerificationModel),
		fmt.Sprint(out.ExpiryHeight),
		out.ProofChain.ChainHash,
	)
}

func IsServiceResolutionSource(source ServiceResolutionSource) bool {
	switch source {
	case ServiceResolutionOnChainRegistry, ServiceResolutionOffChainIndex, ServiceResolutionSignedCache, ServiceResolutionIdentityRecord, ServiceResolutionDistributedMesh:
		return true
	default:
		return false
	}
}

func resolveOnChainRegistryCandidate(input ServiceResolverInput) (ServiceResolverSourceRecord, error) {
	response, err := input.Registry.QueryServiceByName(QueryServiceByName{ServiceName: input.ServiceName, IncludeProof: true})
	if err != nil {
		return ServiceResolverSourceRecord{}, err
	}
	if !response.Found || isDescriptorExpired(response.Descriptor, input.ResolutionHeight) {
		return ServiceResolverSourceRecord{}, nil
	}
	record := ServiceResolverSourceRecord{
		Source:			ServiceResolutionOnChainRegistry,
		ServiceName:		input.ServiceName,
		ServiceID:		response.Descriptor.ServiceID,
		Descriptor:		response.Descriptor,
		Endpoint:		resolverEndpointFromDescriptor(response.Descriptor),
		InterfaceDescriptor:	response.Descriptor.Interface,
		PaymentModel:		registryPaymentModelFromDescriptor(response.Descriptor),
		VerificationModel:	response.Descriptor.Verification.Model,
		ExpiryHeight:		resolverExpiryHeight(response.Descriptor),
		ProofHash:		response.Proof.ProofHash,
	}
	record.TrustMetadata = resolverTrustMetadata(record.Descriptor)
	return NewServiceResolverSourceRecord(record)
}

func resolveIdentityCandidates(input ServiceResolverInput) ([]ServiceResolverSourceRecord, error) {
	bindings := append([]IdentityServiceBinding(nil), input.Registry.IdentityBindings...)
	bindings = append(bindings, input.IdentityBindings...)
	out := make([]ServiceResolverSourceRecord, 0, len(bindings))
	seen := map[string]struct{}{}
	for _, binding := range bindings {
		binding = coretypes.CanonicalIdentityServiceBinding(binding)
		if binding.IdentityName != input.ServiceName {
			continue
		}
		if err := binding.Validate(); err != nil {
			return nil, err
		}
		if binding.ExpiryHeight != 0 && binding.ExpiryHeight < input.ResolutionHeight {
			continue
		}
		if _, ok := seen[binding.BindingHash]; ok {
			continue
		}
		seen[binding.BindingHash] = struct{}{}
		descriptor, found := input.Registry.ServiceDescriptorByID(binding.ServiceID)
		if !found || isDescriptorExpired(descriptor, input.ResolutionHeight) {
			continue
		}
		if binding.DescriptorHash != coretypes.ComputeServiceDescriptorHash(descriptor) {
			return nil, fmt.Errorf("identity binding descriptor hash mismatch for %s", binding.ServiceID)
		}
		record := ServiceResolverSourceRecord{
			Source:			ServiceResolutionIdentityRecord,
			ServiceName:		input.ServiceName,
			ServiceID:		descriptor.ServiceID,
			Descriptor:		descriptor,
			Endpoint:		resolverEndpointFromDescriptor(descriptor),
			InterfaceDescriptor:	descriptor.Interface,
			PaymentModel:		registryPaymentModelFromDescriptor(descriptor),
			VerificationModel:	descriptor.Verification.Model,
			ExpiryHeight:		minNonZero(resolverExpiryHeight(descriptor), binding.ExpiryHeight),
			IdentityBinding:	binding,
			ProofHash:		resolverRegistryProofHash(input.Registry, descriptor.ServiceID),
		}
		record.TrustMetadata = resolverTrustMetadata(record.Descriptor)
		built, err := NewServiceResolverSourceRecord(record)
		if err != nil {
			return nil, err
		}
		out = append(out, built)
	}
	return out, nil
}

func normalizeResolverSourceRecords(source ServiceResolutionSource, records []ServiceResolverSourceRecord, input ServiceResolverInput) ([]ServiceResolverSourceRecord, error) {
	out := make([]ServiceResolverSourceRecord, 0, len(records))
	for _, record := range records {
		record = canonicalResolverSourceRecord(record)
		record.Source = source
		if record.ServiceName != input.ServiceName {
			continue
		}
		if record.ExpiryHeight != 0 && record.ExpiryHeight < input.ResolutionHeight {
			continue
		}
		if record.Descriptor.ServiceID != "" {
			if isDescriptorExpired(record.Descriptor, input.ResolutionHeight) {
				continue
			}
			record.ServiceID = record.Descriptor.ServiceID
			if record.Endpoint == "" {
				record.Endpoint = resolverEndpointFromDescriptor(record.Descriptor)
			}
			if record.InterfaceDescriptor.InterfaceHash == "" {
				record.InterfaceDescriptor = record.Descriptor.Interface
			}
			if record.PaymentModel == "" {
				record.PaymentModel = registryPaymentModelFromDescriptor(record.Descriptor)
			}
			if record.VerificationModel == "" {
				record.VerificationModel = record.Descriptor.Verification.Model
			}
			if record.ExpiryHeight == 0 {
				record.ExpiryHeight = resolverExpiryHeight(record.Descriptor)
			}
			if record.TrustMetadata == "" {
				record.TrustMetadata = resolverTrustMetadata(record.Descriptor)
			}
		}
		record.SourceHash = ""
		built, err := NewServiceResolverSourceRecord(record)
		if err != nil {
			return nil, err
		}
		out = append(out, built)
	}
	return out, nil
}

func resolveDistributedCandidates(input ServiceResolverInput) ([]ServiceResolverSourceRecord, error) {
	if input.Distributed.StateRoot == "" {
		return nil, nil
	}
	if err := input.Distributed.Validate(); err != nil {
		return nil, err
	}
	records, err := DiscoverDistributedServices(input.Distributed, "", "", "", input.ResolutionHeight)
	if err != nil {
		return nil, err
	}
	out := make([]ServiceResolverSourceRecord, 0)
	for _, record := range records {
		if record.ServiceName != input.ServiceName {
			continue
		}
		descriptor, found := input.Registry.ServiceDescriptorByID(record.ServiceID)
		if !found || isDescriptorExpired(descriptor, input.ResolutionHeight) {
			continue
		}
		if record.DescriptorHash != coretypes.ComputeServiceDescriptorHash(descriptor) {
			return nil, fmt.Errorf("distributed descriptor hash mismatch for %s", record.ServiceID)
		}
		endpoint, found := bestDistributedEndpoint(input.Distributed.Endpoints, record)
		if !found {
			continue
		}
		iface, found := distributedInterfaceByHash(input.Distributed.Interfaces, record.InterfaceHash)
		if !found {
			continue
		}
		if record.InterfaceHash != descriptor.Interface.InterfaceHash {
			return nil, fmt.Errorf("distributed interface hash mismatch for %s", record.ServiceID)
		}
		source := ServiceResolverSourceRecord{
			Source:			ServiceResolutionDistributedMesh,
			ServiceName:		input.ServiceName,
			ServiceID:		descriptor.ServiceID,
			Descriptor:		descriptor,
			Endpoint:		endpoint.Target,
			InterfaceDescriptor:	descriptor.Interface,
			TrustMetadata:		resolverTrustMetadata(descriptor),
			PaymentModel:		registryPaymentModelFromDescriptor(descriptor),
			VerificationModel:	descriptor.Verification.Model,
			ExpiryHeight:		minNonZero(resolverExpiryHeight(descriptor), record.ExpiryHeight),
			ProofHash:		resolverRegistryProofHash(input.Registry, descriptor.ServiceID),
			DistributedRecord:	record,
			DistributedEndpoint:	endpoint,
			DistributedInterface:	iface,
		}
		built, err := NewServiceResolverSourceRecord(source)
		if err != nil {
			return nil, err
		}
		out = append(out, built)
	}
	return out, nil
}

func buildServiceResolutionOutput(input ServiceResolverInput, record ServiceResolverSourceRecord) (ServiceResolutionOutput, error) {
	if err := record.Validate(); err != nil {
		return ServiceResolutionOutput{}, err
	}
	chain := buildResolutionProofChain(input, record)
	chain.ChainHash = ComputeServiceResolutionProofChainHash(chain)
	out := ServiceResolutionOutput{
		ServiceName:		input.ServiceName,
		ServiceID:		record.Descriptor.ServiceID,
		Source:			record.Source,
		Descriptor:		record.Descriptor,
		Endpoint:		record.Endpoint,
		InterfaceHash:		record.Descriptor.Interface.InterfaceHash,
		InterfaceDescriptor:	record.InterfaceDescriptor,
		TrustModel:		record.Descriptor.Verification.TrustModel,
		TrustMetadata:		record.TrustMetadata,
		PaymentModel:		record.PaymentModel,
		VerificationModel:	record.VerificationModel,
		ExpiryHeight:		record.ExpiryHeight,
		ProofChain:		chain,
	}
	if out.TrustMetadata == "" {
		out.TrustMetadata = resolverTrustMetadata(out.Descriptor)
	}
	out.ResolutionHash = ComputeServiceResolutionOutputHash(out)
	if input.RequireFreshProof && out.ProofChain.RegistryProofHash == "" && len(out.ProofChain.SignatureHashes) == 0 {
		return ServiceResolutionOutput{}, errors.New("service resolver requires registry proof or signature chain")
	}
	return out, nil
}

func buildResolutionProofChain(input ServiceResolverInput, record ServiceResolverSourceRecord) ServiceResolutionProofChain {
	chain := ServiceResolutionProofChain{
		SourceHashes: []string{record.SourceHash},
	}
	if record.ProofHash != "" {
		chain.RegistryProofHash = record.ProofHash
	}
	if record.SignatureHash != "" {
		chain.SignatureHashes = append(chain.SignatureHashes, record.SignatureHash)
	}
	if record.IdentityBinding.BindingHash != "" {
		chain.IdentityBindingHashes = append(chain.IdentityBindingHashes, record.IdentityBinding.BindingHash)
	}
	if record.DistributedRecord.RecordHash != "" {
		chain.DistributedRecordHashes = append(chain.DistributedRecordHashes, record.DistributedRecord.RecordHash)
	}
	if record.DistributedEndpoint.CommitmentHash != "" {
		chain.DistributedEndpointHashes = append(chain.DistributedEndpointHashes, record.DistributedEndpoint.CommitmentHash)
	}
	if record.DistributedInterface.DescriptorHash != "" {
		chain.DistributedInterfaceHashes = append(chain.DistributedInterfaceHashes, record.DistributedInterface.DescriptorHash)
	}
	for _, commitment := range input.Distributed.Commitments {
		if commitment.ServiceID == record.ServiceID {
			chain.DistributedCommitmentHashes = append(chain.DistributedCommitmentHashes, commitment.CommitmentHash)
		}
	}
	chain.SourceHashes = sortedStrings(chain.SourceHashes)
	chain.SignatureHashes = sortedStrings(chain.SignatureHashes)
	chain.IdentityBindingHashes = sortedStrings(chain.IdentityBindingHashes)
	chain.DistributedRecordHashes = sortedStrings(chain.DistributedRecordHashes)
	chain.DistributedEndpointHashes = sortedStrings(chain.DistributedEndpointHashes)
	chain.DistributedInterfaceHashes = sortedStrings(chain.DistributedInterfaceHashes)
	chain.DistributedCommitmentHashes = sortedStrings(chain.DistributedCommitmentHashes)
	return chain
}

func canonicalResolverSourceRecord(record ServiceResolverSourceRecord) ServiceResolverSourceRecord {
	record.ServiceName = strings.TrimSpace(record.ServiceName)
	record.ServiceID = strings.TrimSpace(record.ServiceID)
	record.Endpoint = strings.TrimSpace(record.Endpoint)
	record.TrustMetadata = strings.TrimSpace(record.TrustMetadata)
	record.PaymentModel = strings.TrimSpace(record.PaymentModel)
	record.ProofHash = strings.ToLower(strings.TrimSpace(record.ProofHash))
	record.SignatureHash = strings.ToLower(strings.TrimSpace(record.SignatureHash))
	record.SourceHash = strings.ToLower(strings.TrimSpace(record.SourceHash))
	if record.Descriptor.ServiceID != "" {
		record.Descriptor = coretypes.CanonicalServiceDescriptor(record.Descriptor)
	}
	return record
}

func sortResolverCandidates(candidates []ServiceResolverSourceRecord) {
	policy, _ := NewServiceResolverFallbackPolicy(nil)
	priority := serviceResolverFallbackRank(policy)
	sort.SliceStable(candidates, func(i, j int) bool {
		if priority[candidates[i].Source] != priority[candidates[j].Source] {
			return priority[candidates[i].Source] < priority[candidates[j].Source]
		}
		if candidates[i].ExpiryHeight != candidates[j].ExpiryHeight {
			return candidates[i].ExpiryHeight > candidates[j].ExpiryHeight
		}
		return candidates[i].SourceHash < candidates[j].SourceHash
	})
}

func resolverEndpointFromDescriptor(descriptor ServiceDescriptor) string {
	switch {
	case descriptor.Execution.Endpoint != "":
		return descriptor.Execution.Endpoint
	case descriptor.Execution.Target != "":
		return descriptor.Execution.Target
	case descriptor.Execution.ContractAddress != "":
		return descriptor.Execution.ContractAddress
	case descriptor.Execution.ModuleRoute != "":
		return descriptor.Execution.ModuleRoute
	case descriptor.Execution.ProviderPoolID != "":
		return descriptor.Execution.ProviderPoolID
	default:
		return descriptor.EndpointKey
	}
}

func resolverTrustMetadata(descriptor ServiceDescriptor) string {
	return strings.Join([]string{
		string(descriptor.Verification.TrustModel),
		string(descriptor.Verification.Model),
		descriptor.Verification.ProofFormat,
		fmt.Sprint(descriptor.Verification.ChallengeWindow),
		descriptor.Verification.FallbackServiceID,
		descriptor.Verification.ProviderCollateralDenom,
		descriptor.Verification.ProviderCollateralAmount,
	}, ":")
}

func resolverExpiryHeight(descriptor ServiceDescriptor) uint64 {
	return minNonZero(descriptor.ExpiryHeight, descriptor.Discovery.CacheExpiryHeight, descriptor.Payment.ExpiryHeight)
}

func resolverRegistryProofHash(state ServiceRegistryState, serviceID string) string {
	proof, err := state.QueryServiceProof(QueryServiceProof{ServiceID: serviceID})
	if err != nil || !proof.Found {
		return ""
	}
	return proof.Proof.ProofHash
}

func isDescriptorExpired(descriptor ServiceDescriptor, height uint64) bool {
	return descriptor.ExpiryHeight != 0 && descriptor.ExpiryHeight < height
}

func bestDistributedEndpoint(endpoints []DistributedServiceEndpoint, record DistributedServiceRecord) (DistributedServiceEndpoint, bool) {
	var matches []DistributedServiceEndpoint
	for _, endpoint := range endpoints {
		if endpoint.ServiceID == record.ServiceID && endpoint.InterfaceHash == record.InterfaceHash {
			matches = append(matches, endpoint)
		}
	}
	if len(matches) == 0 {
		return DistributedServiceEndpoint{}, false
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority < matches[j].Priority
		}
		if matches[i].Weight != matches[j].Weight {
			return matches[i].Weight > matches[j].Weight
		}
		return matches[i].CommitmentHash < matches[j].CommitmentHash
	})
	return matches[0], true
}

func distributedInterfaceByHash(interfaces []DistributedInterfaceDescriptor, interfaceHash string) (DistributedInterfaceDescriptor, bool) {
	for _, iface := range interfaces {
		if iface.InterfaceHash == interfaceHash {
			return iface, true
		}
	}
	return DistributedInterfaceDescriptor{}, false
}

func validateHashList(name string, hashes []string) error {
	for _, hash := range hashes {
		if err := coretypes.ValidateHash(name, hash); err != nil {
			return err
		}
	}
	return nil
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func minNonZero(values ...uint64) uint64 {
	var min uint64
	for _, value := range values {
		if value == 0 {
			continue
		}
		if min == 0 || value < min {
			min = value
		}
	}
	return min
}
