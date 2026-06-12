package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceDiscoveryCacheTrust string

const (
	ServiceDiscoveryCacheVerified	ServiceDiscoveryCacheTrust	= "verified"
	ServiceDiscoveryCacheAdvisory	ServiceDiscoveryCacheTrust	= "advisory"
)

type ServiceDiscoveryCacheRecord struct {
	ServiceID		string
	DescriptorHash		string
	InterfaceHash		string
	Source			ServiceResolutionSource
	ProofHeightOptional	uint64
	SignatureOptional	string
	ExpiresHeight		uint64
	FetchedAtHeight		uint64
	Trust			ServiceDiscoveryCacheTrust
	CacheHash		string
}

type ServiceDiscoveryCacheConstraints struct {
	RegistryExpiryHeight		uint64
	InterfaceCompatibleUntil	uint64
	CurrentHeight			uint64
}

type ServiceDiscoveryUpdateEvent struct {
	ServiceID	string
	DescriptorHash	string
	InterfaceHash	string
	Height		uint64
	EventHash	string
}

type SignedServiceAdvertisement struct {
	ServiceName		string
	Descriptor		ServiceDiscoveryDescriptorV1
	Endpoint		string
	InterfaceHash		string
	Signer			string
	ProviderOptional	string
	ExpiresHeight		uint64
	IssuedAtHeight		uint64
	Nonce			uint64
	AdvertisementHash	string
	SignatureHash		string
}

type QueryServiceDiscovery struct {
	ServiceName	string
	IncludeProof	bool
}

type QueryServiceDiscoveriesByIdentity struct {
	IdentityName	string
	IncludeProof	bool
	CurrentHeight	uint64
}

type QueryServiceDiscoveryResponse struct {
	Descriptor		ServiceDiscoveryDescriptorV1
	RegistryDescriptor	ServiceDescriptor
	Proof			ServiceRegistryProof
	Found			bool
}

type QueryServiceDiscoveriesResponse struct {
	Services	[]QueryServiceDiscoveryResponse
	Total		uint64
}

type ServiceResolverFallbackPolicy struct {
	Order		[]ServiceResolutionSource
	PolicyHash	string
}

func NewServiceDiscoveryCacheRecord(record ServiceDiscoveryCacheRecord, constraints ServiceDiscoveryCacheConstraints) (ServiceDiscoveryCacheRecord, error) {
	if record.CacheHash != "" {
		return ServiceDiscoveryCacheRecord{}, errors.New("service discovery cache hash must be empty before construction")
	}
	record = canonicalServiceDiscoveryCacheRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return ServiceDiscoveryCacheRecord{}, err
	}
	if err := ValidateServiceDiscoveryCacheConstraints(record, constraints); err != nil {
		return ServiceDiscoveryCacheRecord{}, err
	}
	record.CacheHash = ComputeServiceDiscoveryCacheRecordHash(record)
	return record, record.Validate(constraints)
}

func (record ServiceDiscoveryCacheRecord) ValidateFormat() error {
	record = canonicalServiceDiscoveryCacheRecord(record)
	if err := validateInterfaceToken("service discovery cache service id", record.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("service discovery cache descriptor hash", record.DescriptorHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("service discovery cache interface hash", record.InterfaceHash); err != nil {
		return err
	}
	if !IsServiceResolutionSource(record.Source) {
		return fmt.Errorf("unknown service discovery cache source %q", record.Source)
	}
	if record.SignatureOptional != "" {
		if err := coretypes.ValidateHash("service discovery cache signature", record.SignatureOptional); err != nil {
			return err
		}
	}
	if record.ExpiresHeight == 0 || record.FetchedAtHeight == 0 {
		return errors.New("service discovery cache fetched and expiry heights are required")
	}
	if record.ExpiresHeight <= record.FetchedAtHeight {
		return errors.New("service discovery cache expiry must exceed fetched height")
	}
	if !IsServiceDiscoveryCacheTrust(record.Trust) {
		return fmt.Errorf("unknown service discovery cache trust %q", record.Trust)
	}
	if record.ProofHeightOptional == 0 && record.SignatureOptional == "" && record.Trust != ServiceDiscoveryCacheAdvisory {
		return errors.New("unverified service discovery cache entries must be advisory")
	}
	if record.CacheHash != "" {
		return coretypes.ValidateHash("service discovery cache hash", record.CacheHash)
	}
	return nil
}

func (record ServiceDiscoveryCacheRecord) Validate(constraints ServiceDiscoveryCacheConstraints) error {
	record = canonicalServiceDiscoveryCacheRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return err
	}
	if err := ValidateServiceDiscoveryCacheConstraints(record, constraints); err != nil {
		return err
	}
	if record.CacheHash == "" {
		return errors.New("service discovery cache hash is required")
	}
	if expected := ComputeServiceDiscoveryCacheRecordHash(record); record.CacheHash != expected {
		return fmt.Errorf("service discovery cache hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateServiceDiscoveryCacheConstraints(record ServiceDiscoveryCacheRecord, constraints ServiceDiscoveryCacheConstraints) error {
	if constraints.RegistryExpiryHeight != 0 && record.ExpiresHeight > constraints.RegistryExpiryHeight {
		return errors.New("service discovery cache cannot outlive registry expiry")
	}
	if constraints.InterfaceCompatibleUntil != 0 && record.ExpiresHeight > constraints.InterfaceCompatibleUntil {
		return errors.New("service discovery cache cannot outlive interface compatibility")
	}
	if constraints.CurrentHeight != 0 && constraints.CurrentHeight >= record.ExpiresHeight {
		return errors.New("service discovery cache record is stale")
	}
	return nil
}

func NewServiceDiscoveryUpdateEvent(serviceID, descriptorHash, interfaceHash string, height uint64) (ServiceDiscoveryUpdateEvent, error) {
	event := ServiceDiscoveryUpdateEvent{
		ServiceID:	strings.TrimSpace(serviceID),
		DescriptorHash:	strings.ToLower(strings.TrimSpace(descriptorHash)),
		InterfaceHash:	strings.ToLower(strings.TrimSpace(interfaceHash)),
		Height:		height,
	}
	if err := event.ValidateFormat(); err != nil {
		return ServiceDiscoveryUpdateEvent{}, err
	}
	event.EventHash = ComputeServiceDiscoveryUpdateEventHash(event)
	return event, event.Validate()
}

func (event ServiceDiscoveryUpdateEvent) ValidateFormat() error {
	if err := validateInterfaceToken("service discovery update service id", event.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("service discovery update descriptor hash", event.DescriptorHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("service discovery update interface hash", event.InterfaceHash); err != nil {
		return err
	}
	if event.Height == 0 {
		return errors.New("service discovery update height is required")
	}
	if event.EventHash != "" {
		return coretypes.ValidateHash("service discovery update event hash", event.EventHash)
	}
	return nil
}

func (event ServiceDiscoveryUpdateEvent) Validate() error {
	if err := event.ValidateFormat(); err != nil {
		return err
	}
	if event.EventHash == "" {
		return errors.New("service discovery update event hash is required")
	}
	if expected := ComputeServiceDiscoveryUpdateEventHash(event); event.EventHash != expected {
		return fmt.Errorf("service discovery update event hash mismatch: expected %s", expected)
	}
	return nil
}

func InvalidateDiscoveryCacheOnServiceUpdate(records []ServiceDiscoveryCacheRecord, event ServiceDiscoveryUpdateEvent) ([]ServiceDiscoveryCacheRecord, error) {
	if err := event.Validate(); err != nil {
		return nil, err
	}
	out := make([]ServiceDiscoveryCacheRecord, 0, len(records))
	for _, record := range records {
		record = canonicalServiceDiscoveryCacheRecord(record)
		if record.ServiceID == event.ServiceID {
			continue
		}
		out = append(out, record)
	}
	sortServiceDiscoveryCacheRecords(out)
	return out, nil
}

func NewSignedServiceAdvertisement(ad SignedServiceAdvertisement) (SignedServiceAdvertisement, error) {
	if ad.AdvertisementHash != "" || ad.SignatureHash != "" {
		return SignedServiceAdvertisement{}, errors.New("signed service advertisement hashes must be empty before construction")
	}
	ad = canonicalSignedServiceAdvertisement(ad)
	if err := ad.ValidateFormat(); err != nil {
		return SignedServiceAdvertisement{}, err
	}
	ad.AdvertisementHash = ComputeSignedServiceAdvertisementHash(ad)
	ad.SignatureHash = ComputeSignedServiceAdvertisementSignatureHash(ad)
	return ad, ad.Validate()
}

func (ad SignedServiceAdvertisement) ValidateFormat() error {
	ad = canonicalSignedServiceAdvertisement(ad)
	if err := validateInterfaceToken("signed service advertisement service name", ad.ServiceName); err != nil {
		return err
	}
	if err := ad.Descriptor.Validate(); err != nil {
		return err
	}
	if ad.Descriptor.ServiceName != ad.ServiceName {
		return errors.New("signed service advertisement service name mismatch")
	}
	if ad.Endpoint == "" || ad.Endpoint != ad.Descriptor.Endpoint {
		return errors.New("signed service advertisement endpoint mismatch")
	}
	if ad.InterfaceHash != ad.Descriptor.InterfaceHash {
		return errors.New("signed service advertisement interface hash mismatch")
	}
	if err := validateInterfaceToken("signed service advertisement signer", ad.Signer); err != nil {
		return err
	}
	if ad.Signer != ad.Descriptor.Owner && ad.Signer != ad.ProviderOptional {
		return errors.New("signed service advertisement signer must be owner or provider")
	}
	if ad.ProviderOptional != "" {
		if err := validateInterfaceToken("signed service advertisement provider", ad.ProviderOptional); err != nil {
			return err
		}
	}
	if ad.IssuedAtHeight == 0 || ad.ExpiresHeight == 0 {
		return errors.New("signed service advertisement heights are required")
	}
	if ad.ExpiresHeight <= ad.IssuedAtHeight {
		return errors.New("signed service advertisement expiry must exceed issue height")
	}
	if ad.ExpiresHeight > ad.Descriptor.ExpiryHeight {
		return errors.New("signed service advertisement cannot outlive descriptor")
	}
	if ad.Nonce == 0 {
		return errors.New("signed service advertisement nonce is required")
	}
	if ad.AdvertisementHash != "" {
		if err := coretypes.ValidateHash("signed service advertisement hash", ad.AdvertisementHash); err != nil {
			return err
		}
	}
	if ad.SignatureHash != "" {
		if err := coretypes.ValidateHash("signed service advertisement signature hash", ad.SignatureHash); err != nil {
			return err
		}
	}
	return nil
}

func (ad SignedServiceAdvertisement) Validate() error {
	ad = canonicalSignedServiceAdvertisement(ad)
	if err := ad.ValidateFormat(); err != nil {
		return err
	}
	if ad.AdvertisementHash == "" || ad.SignatureHash == "" {
		return errors.New("signed service advertisement hashes are required")
	}
	if expected := ComputeSignedServiceAdvertisementHash(ad); ad.AdvertisementHash != expected {
		return fmt.Errorf("signed service advertisement hash mismatch: expected %s", expected)
	}
	if expected := ComputeSignedServiceAdvertisementSignatureHash(ad); ad.SignatureHash != expected {
		return fmt.Errorf("signed service advertisement signature mismatch: expected %s", expected)
	}
	return nil
}

func ServiceResolverSourceRecordFromAdvertisement(ad SignedServiceAdvertisement, descriptor ServiceDescriptor) (ServiceResolverSourceRecord, error) {
	if err := ad.Validate(); err != nil {
		return ServiceResolverSourceRecord{}, err
	}
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceResolverSourceRecord{}, err
	}
	if descriptor.ServiceID != ad.Descriptor.ServiceID ||
		descriptor.Interface.InterfaceHash != ad.Descriptor.InterfaceHash ||
		resolverEndpointFromDescriptor(descriptor) != ad.Endpoint {
		return ServiceResolverSourceRecord{}, errors.New("signed service advertisement registry descriptor mismatch")
	}
	record := ServiceResolverSourceRecord{
		Source:			ServiceResolutionSignedCache,
		ServiceName:		ad.ServiceName,
		ServiceID:		descriptor.ServiceID,
		Descriptor:		descriptor,
		Endpoint:		ad.Endpoint,
		TrustMetadata:		string(ad.Descriptor.TrustModel),
		PaymentModel:		ad.Descriptor.PaymentModel,
		VerificationModel:	ad.Descriptor.VerificationModel,
		ExpiryHeight:		ad.ExpiresHeight,
		SignatureHash:		ad.SignatureHash,
	}
	return NewServiceResolverSourceRecord(record)
}

func QueryServiceDiscoveryFromState(state ServiceRegistryState, q QueryServiceDiscovery, height uint64) (QueryServiceDiscoveryResponse, error) {
	if height == 0 {
		return QueryServiceDiscoveryResponse{}, errors.New("service discovery query height is required")
	}
	q.ServiceName = strings.TrimSpace(q.ServiceName)
	if err := validateInterfaceToken("service discovery query service name", q.ServiceName); err != nil {
		return QueryServiceDiscoveryResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return QueryServiceDiscoveryResponse{}, err
	}
	response, err := state.QueryServiceByName(QueryServiceByName{ServiceName: q.ServiceName, IncludeProof: q.IncludeProof})
	if err != nil {
		return QueryServiceDiscoveryResponse{}, err
	}
	if !response.Found || isDescriptorExpired(response.Descriptor, height) {
		return QueryServiceDiscoveryResponse{Found: false}, nil
	}
	proofHash := ""
	if q.IncludeProof {
		proofHash = response.Proof.ProofHash
	}
	descriptor, err := ProjectServiceDiscoveryDescriptorFromCore(response.Descriptor, q.ServiceName, proofHash, "")
	if err != nil {
		return QueryServiceDiscoveryResponse{}, err
	}
	return QueryServiceDiscoveryResponse{
		Descriptor:		descriptor,
		RegistryDescriptor:	response.Descriptor,
		Proof:			response.Proof,
		Found:			true,
	}, nil
}

func QueryServiceDiscoveriesByIdentityFromState(state ServiceRegistryState, q QueryServiceDiscoveriesByIdentity) (QueryServiceDiscoveriesResponse, error) {
	q.IdentityName = strings.TrimSpace(q.IdentityName)
	if err := validateInterfaceToken("service discovery identity query identity name", q.IdentityName); err != nil {
		return QueryServiceDiscoveriesResponse{}, err
	}
	if q.CurrentHeight == 0 {
		return QueryServiceDiscoveriesResponse{}, errors.New("service discovery identity query height is required")
	}
	if err := state.Validate(); err != nil {
		return QueryServiceDiscoveriesResponse{}, err
	}
	ids := state.ServiceIDsByIdentity(q.IdentityName)
	seen := map[string]struct{}{}
	out := make([]QueryServiceDiscoveryResponse, 0, len(ids))
	for _, serviceID := range ids {
		if _, ok := seen[serviceID]; ok {
			continue
		}
		seen[serviceID] = struct{}{}
		descriptor, found := state.ServiceDescriptorByID(serviceID)
		if !found || isDescriptorExpired(descriptor, q.CurrentHeight) {
			continue
		}
		proof := ServiceRegistryProof{}
		proofHash := ""
		if q.IncludeProof {
			proofResponse, err := state.QueryServiceProof(QueryServiceProof{ServiceID: serviceID})
			if err != nil {
				return QueryServiceDiscoveriesResponse{}, err
			}
			if proofResponse.Found {
				proof = proofResponse.Proof
				proofHash = proof.ProofHash
			}
		}
		projected, err := ProjectServiceDiscoveryDescriptorFromCore(descriptor, descriptor.Discovery.ServiceName, proofHash, "")
		if err != nil {
			return QueryServiceDiscoveriesResponse{}, err
		}
		out = append(out, QueryServiceDiscoveryResponse{
			Descriptor:		projected,
			RegistryDescriptor:	descriptor,
			Proof:			proof,
			Found:			true,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Descriptor.ServiceID < out[j].Descriptor.ServiceID
	})
	return QueryServiceDiscoveriesResponse{Services: out, Total: uint64(len(out))}, nil
}

func NewServiceResolverFallbackPolicy(order []ServiceResolutionSource) (ServiceResolverFallbackPolicy, error) {
	policy := ServiceResolverFallbackPolicy{Order: append([]ServiceResolutionSource(nil), order...)}
	if len(policy.Order) == 0 {
		policy.Order = DefaultServiceResolverFallbackOrder()
	}
	if err := policy.ValidateFormat(); err != nil {
		return ServiceResolverFallbackPolicy{}, err
	}
	policy.PolicyHash = ComputeServiceResolverFallbackPolicyHash(policy)
	return policy, policy.Validate()
}

func DefaultServiceResolverFallbackOrder() []ServiceResolutionSource {
	return []ServiceResolutionSource{
		ServiceResolutionOnChainRegistry,
		ServiceResolutionIdentityRecord,
		ServiceResolutionSignedCache,
		ServiceResolutionOffChainIndex,
		ServiceResolutionDistributedMesh,
	}
}

func (policy ServiceResolverFallbackPolicy) ValidateFormat() error {
	if len(policy.Order) == 0 {
		return errors.New("service resolver fallback order is required")
	}
	seen := map[ServiceResolutionSource]struct{}{}
	for _, source := range policy.Order {
		if !IsServiceResolutionSource(source) {
			return fmt.Errorf("unknown service resolver fallback source %q", source)
		}
		if _, ok := seen[source]; ok {
			return fmt.Errorf("duplicate service resolver fallback source %q", source)
		}
		seen[source] = struct{}{}
	}
	if policy.PolicyHash != "" {
		return coretypes.ValidateHash("service resolver fallback policy hash", policy.PolicyHash)
	}
	return nil
}

func (policy ServiceResolverFallbackPolicy) Validate() error {
	if err := policy.ValidateFormat(); err != nil {
		return err
	}
	if policy.PolicyHash == "" {
		return errors.New("service resolver fallback policy hash is required")
	}
	if expected := ComputeServiceResolverFallbackPolicyHash(policy); policy.PolicyHash != expected {
		return fmt.Errorf("service resolver fallback policy hash mismatch: expected %s", expected)
	}
	return nil
}

func serviceResolverFallbackRank(policy ServiceResolverFallbackPolicy) map[ServiceResolutionSource]int {
	ranks := map[ServiceResolutionSource]int{}
	for i, source := range policy.Order {
		ranks[source] = i
	}
	return ranks
}

func ComputeServiceDiscoveryCacheRecordHash(record ServiceDiscoveryCacheRecord) string {
	record = canonicalServiceDiscoveryCacheRecord(record)
	return servicesHashParts(
		"aetra-services-discovery-cache-v1",
		record.ServiceID,
		record.DescriptorHash,
		record.InterfaceHash,
		string(record.Source),
		fmt.Sprint(record.ProofHeightOptional),
		record.SignatureOptional,
		fmt.Sprint(record.ExpiresHeight),
		fmt.Sprint(record.FetchedAtHeight),
		string(record.Trust),
	)
}

func ComputeServiceDiscoveryUpdateEventHash(event ServiceDiscoveryUpdateEvent) string {
	return servicesHashParts(
		"aetra-services-discovery-update-event-v1",
		event.ServiceID,
		event.DescriptorHash,
		event.InterfaceHash,
		fmt.Sprint(event.Height),
	)
}

func ComputeSignedServiceAdvertisementHash(ad SignedServiceAdvertisement) string {
	ad = canonicalSignedServiceAdvertisement(ad)
	return servicesHashParts(
		"aetra-services-signed-advertisement-v1",
		ad.ServiceName,
		ad.Descriptor.DescriptorHash,
		ad.Endpoint,
		ad.InterfaceHash,
		ad.Signer,
		ad.ProviderOptional,
		fmt.Sprint(ad.ExpiresHeight),
		fmt.Sprint(ad.IssuedAtHeight),
		fmt.Sprint(ad.Nonce),
	)
}

func ComputeSignedServiceAdvertisementSignatureHash(ad SignedServiceAdvertisement) string {
	ad = canonicalSignedServiceAdvertisement(ad)
	return servicesHashParts(
		"aetra-services-signed-advertisement-signature-v1",
		ad.Signer,
		ComputeSignedServiceAdvertisementHash(ad),
	)
}

func ComputeServiceResolverFallbackPolicyHash(policy ServiceResolverFallbackPolicy) string {
	parts := []string{"aetra-services-resolver-fallback-policy-v1", fmt.Sprint(len(policy.Order))}
	for _, source := range policy.Order {
		parts = append(parts, string(source))
	}
	return servicesHashParts(parts...)
}

func IsServiceDiscoveryCacheTrust(trust ServiceDiscoveryCacheTrust) bool {
	switch trust {
	case ServiceDiscoveryCacheVerified, ServiceDiscoveryCacheAdvisory:
		return true
	default:
		return false
	}
}

func canonicalServiceDiscoveryCacheRecord(record ServiceDiscoveryCacheRecord) ServiceDiscoveryCacheRecord {
	record.ServiceID = strings.TrimSpace(record.ServiceID)
	record.DescriptorHash = strings.ToLower(strings.TrimSpace(record.DescriptorHash))
	record.InterfaceHash = strings.ToLower(strings.TrimSpace(record.InterfaceHash))
	record.SignatureOptional = strings.ToLower(strings.TrimSpace(record.SignatureOptional))
	record.CacheHash = strings.ToLower(strings.TrimSpace(record.CacheHash))
	return record
}

func canonicalSignedServiceAdvertisement(ad SignedServiceAdvertisement) SignedServiceAdvertisement {
	ad.ServiceName = strings.TrimSpace(ad.ServiceName)
	ad.Descriptor = canonicalServiceDiscoveryDescriptorV1(ad.Descriptor)
	ad.Endpoint = strings.TrimSpace(ad.Endpoint)
	ad.InterfaceHash = strings.ToLower(strings.TrimSpace(ad.InterfaceHash))
	ad.Signer = strings.TrimSpace(ad.Signer)
	ad.ProviderOptional = strings.TrimSpace(ad.ProviderOptional)
	ad.AdvertisementHash = strings.ToLower(strings.TrimSpace(ad.AdvertisementHash))
	ad.SignatureHash = strings.ToLower(strings.TrimSpace(ad.SignatureHash))
	return ad
}

func sortServiceDiscoveryCacheRecords(records []ServiceDiscoveryCacheRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].ServiceID != records[j].ServiceID {
			return records[i].ServiceID < records[j].ServiceID
		}
		return records[i].CacheHash < records[j].CacheHash
	})
}
