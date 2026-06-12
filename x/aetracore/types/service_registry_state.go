package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	ServiceRegistryDescriptorsPrefix	= "services/descriptors"
	ServiceRegistryAnchorsPrefix		= "services/anchors"
	ServiceRegistryInterfacesPrefix		= "services/interfaces"
	ServiceRegistryOwnersPrefix		= "services/owners"
	ServiceRegistryNamesPrefix		= "services/names"
	ServiceRegistryIdentityBindingsPrefix	= "services/identity_bindings"
	ServiceRegistryProvidersPrefix		= "services/providers"
	ServiceRegistryExpiryPrefix		= "services/expiry"
	ServiceRegistryReputationPrefix		= "services/reputation"
	ServiceRegistryReceiptsPrefix		= "services/receipts"
)

type ServiceInterface = ServiceInterfaceDescriptor
type ServiceReceipt = ServiceCallReceipt

type ServiceRegistryStateEntryType string

const (
	ServiceRegistryStateDescriptor		ServiceRegistryStateEntryType	= "SERVICE_DESCRIPTOR"
	ServiceRegistryStateAnchor		ServiceRegistryStateEntryType	= "SERVICE_ANCHOR"
	ServiceRegistryStateInterface		ServiceRegistryStateEntryType	= "SERVICE_INTERFACE"
	ServiceRegistryStateOwnerIndex		ServiceRegistryStateEntryType	= "SERVICE_OWNER_INDEX"
	ServiceRegistryStateNameIndex		ServiceRegistryStateEntryType	= "SERVICE_NAME_INDEX"
	ServiceRegistryStateIdentityBinding	ServiceRegistryStateEntryType	= "IDENTITY_SERVICE_BINDING"
	ServiceRegistryStateProvider		ServiceRegistryStateEntryType	= "PROVIDER_RECORD"
	ServiceRegistryStateExpiryIndex		ServiceRegistryStateEntryType	= "SERVICE_EXPIRY_INDEX"
	ServiceRegistryStateReputation		ServiceRegistryStateEntryType	= "REPUTATION_RECORD"
	ServiceRegistryStateReceipt		ServiceRegistryStateEntryType	= "SERVICE_RECEIPT"
)

type ServiceAnchor struct {
	ServiceID		string
	DescriptorHash		string
	InterfaceHash		string
	ProviderRoot		string
	VerificationModel	ServiceVerificationModel
	ExpiryHeight		uint64
	AnchorHash		string
}

type IdentityServiceBinding struct {
	IdentityName	string
	ServiceID	string
	Owner		string
	DescriptorHash	string
	CreatedHeight	uint64
	ExpiryHeight	uint64
	BindingHash	string
}

type ProviderRecord struct {
	ServiceID	string
	Provider	FogProviderRecord
	RecordHash	string
}

type ReputationRecord struct {
	ProviderID	string
	Score		uint64
	Successes	uint64
	Failures	uint64
	UpdatedHeight	uint64
	RecordHash	string
}

type ServiceRegistryStateEntry struct {
	Key		string
	Value		string
	EntryType	ServiceRegistryStateEntryType
	EntryHash	string
}

type ServiceRegistryState struct {
	Descriptors		[]ServiceDescriptor
	Anchors			[]ServiceAnchor
	Interfaces		[]ServiceInterface
	OwnerIndex		[]ServiceRegistryStateEntry
	NameIndex		[]ServiceRegistryStateEntry
	IdentityBindings	[]IdentityServiceBinding
	Providers		[]ProviderRecord
	ExpiryIndex		[]ServiceRegistryStateEntry
	Reputations		[]ReputationRecord
	Receipts		[]ServiceReceipt
	Entries			[]ServiceRegistryStateEntry
	StateRoot		string
	UpdatedHeight		uint64
}

func ServiceDescriptorStateKey(serviceID string) (string, error) {
	if err := validatePolicyID("aetracore service registry state service id", serviceID); err != nil {
		return "", err
	}
	return ServiceRegistryDescriptorsPrefix + "/" + serviceID, nil
}

func ServiceAnchorStateKey(serviceID string) (string, error) {
	if err := validatePolicyID("aetracore service registry anchor service id", serviceID); err != nil {
		return "", err
	}
	return ServiceRegistryAnchorsPrefix + "/" + serviceID, nil
}

func ServiceInterfaceStateKey(interfaceHash string) (string, error) {
	if err := ValidateHash("aetracore service registry interface hash", interfaceHash); err != nil {
		return "", err
	}
	return ServiceRegistryInterfacesPrefix + "/" + strings.ToLower(strings.TrimSpace(interfaceHash)), nil
}

func ServiceOwnerStateKey(owner, serviceID string) (string, error) {
	if err := addressing.ValidateAuthorityAddress("aetracore service registry state owner", owner); err != nil {
		return "", err
	}
	if err := validatePolicyID("aetracore service registry owner service id", serviceID); err != nil {
		return "", err
	}
	return ServiceRegistryOwnersPrefix + "/" + strings.TrimSpace(owner) + "/" + serviceID, nil
}

func ServiceNameStateKey(serviceName string) (string, error) {
	if err := validatePolicyID("aetracore service registry service name", serviceName); err != nil {
		return "", err
	}
	return ServiceRegistryNamesPrefix + "/" + serviceName, nil
}

func IdentityServiceBindingStateKey(identityName, serviceID string) (string, error) {
	if err := validatePolicyID("aetracore service registry identity name", identityName); err != nil {
		return "", err
	}
	if err := validatePolicyID("aetracore service registry identity service id", serviceID); err != nil {
		return "", err
	}
	return ServiceRegistryIdentityBindingsPrefix + "/" + identityName + "/" + serviceID, nil
}

func ServiceProviderStateKey(serviceID, providerID string) (string, error) {
	if err := validatePolicyID("aetracore service registry provider service id", serviceID); err != nil {
		return "", err
	}
	if err := validatePolicyID("aetracore service registry provider id", providerID); err != nil {
		return "", err
	}
	return ServiceRegistryProvidersPrefix + "/" + serviceID + "/" + providerID, nil
}

func ServiceExpiryStateKey(expiryHeight uint64, serviceID string) (string, error) {
	if expiryHeight == 0 {
		return "", errors.New("aetracore service registry expiry height must be positive")
	}
	if err := validatePolicyID("aetracore service registry expiry service id", serviceID); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%020d/%s", ServiceRegistryExpiryPrefix, expiryHeight, serviceID), nil
}

func ServiceReputationStateKey(providerID string) (string, error) {
	if err := validatePolicyID("aetracore service registry reputation provider id", providerID); err != nil {
		return "", err
	}
	return ServiceRegistryReputationPrefix + "/" + providerID, nil
}

func ServiceReceiptStateKey(serviceID, callID string) (string, error) {
	if err := validatePolicyID("aetracore service registry receipt service id", serviceID); err != nil {
		return "", err
	}
	if err := ValidateHash("aetracore service registry receipt call id", callID); err != nil {
		return "", err
	}
	return ServiceRegistryReceiptsPrefix + "/" + serviceID + "/" + strings.ToLower(strings.TrimSpace(callID)), nil
}

func NewServiceAnchorFromDescriptor(descriptor ServiceDescriptor) (ServiceAnchor, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceAnchor{}, err
	}
	anchor := ServiceAnchor{
		ServiceID:		descriptor.ServiceID,
		DescriptorHash:		ComputeServiceDescriptorHash(descriptor),
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		ProviderRoot:		registryProviderSet(descriptor),
		VerificationModel:	descriptor.Verification.Model,
		ExpiryHeight:		descriptor.ExpiryHeight,
	}
	anchor.AnchorHash = ComputeServiceAnchorHash(anchor)
	return anchor, anchor.Validate()
}

func NewServiceAnchorFromHybridAnchor(anchor HybridServiceRegistryAnchor) (ServiceAnchor, error) {
	anchor = CanonicalHybridServiceRegistryAnchor(anchor)
	if err := anchor.Validate(); err != nil {
		return ServiceAnchor{}, err
	}
	serviceAnchor := ServiceAnchor{
		ServiceID:		anchor.ServiceID,
		DescriptorHash:		anchor.DescriptorHash,
		InterfaceHash:		anchor.InterfaceHash,
		ProviderRoot:		anchor.ProviderRoot,
		VerificationModel:	anchor.VerificationModel,
		ExpiryHeight:		anchor.ExpiryHeight,
	}
	serviceAnchor.AnchorHash = ComputeServiceAnchorHash(serviceAnchor)
	return serviceAnchor, serviceAnchor.Validate()
}

func NewIdentityServiceBindingFromDescriptor(descriptor ServiceDescriptor) (IdentityServiceBinding, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return IdentityServiceBinding{}, err
	}
	if descriptor.Discovery.IdentityName == "" {
		return IdentityServiceBinding{}, errors.New("aetracore identity service binding requires identity name")
	}
	binding := IdentityServiceBinding{
		IdentityName:	descriptor.Discovery.IdentityName,
		ServiceID:	descriptor.ServiceID,
		Owner:		descriptor.Owner,
		DescriptorHash:	ComputeServiceDescriptorHash(descriptor),
		CreatedHeight:	descriptor.CreatedHeight,
		ExpiryHeight:	descriptor.ExpiryHeight,
	}
	binding.BindingHash = ComputeIdentityServiceBindingHash(binding)
	return binding, binding.Validate()
}

func NewProviderRecord(serviceID string, provider FogProviderRecord) (ProviderRecord, error) {
	provider = CanonicalFogProviderRecord(provider)
	if err := provider.Validate(); err != nil {
		return ProviderRecord{}, err
	}
	record := ProviderRecord{
		ServiceID:	serviceID,
		Provider:	provider,
	}
	record.RecordHash = ComputeProviderRecordHash(record)
	return record, record.Validate()
}

func NewReputationRecord(providerID string, score, successes, failures, updatedHeight uint64) (ReputationRecord, error) {
	record := ReputationRecord{
		ProviderID:	strings.TrimSpace(providerID),
		Score:		score,
		Successes:	successes,
		Failures:	failures,
		UpdatedHeight:	updatedHeight,
	}
	record.RecordHash = ComputeReputationRecordHash(record)
	return record, record.Validate()
}

func NewServiceRegistryState(descriptors []ServiceDescriptor, anchors []ServiceAnchor, identityBindings []IdentityServiceBinding, providers []ProviderRecord, reputations []ReputationRecord, receipts []ServiceReceipt, height uint64) (ServiceRegistryState, error) {
	if height == 0 {
		return ServiceRegistryState{}, errors.New("aetracore service registry state height must be positive")
	}
	state := ServiceRegistryState{
		Descriptors:		make([]ServiceDescriptor, 0, len(descriptors)),
		Anchors:		make([]ServiceAnchor, 0, len(anchors)),
		Interfaces:		[]ServiceInterface{},
		OwnerIndex:		[]ServiceRegistryStateEntry{},
		NameIndex:		[]ServiceRegistryStateEntry{},
		IdentityBindings:	make([]IdentityServiceBinding, 0, len(identityBindings)),
		Providers:		make([]ProviderRecord, 0, len(providers)),
		ExpiryIndex:		[]ServiceRegistryStateEntry{},
		Reputations:		make([]ReputationRecord, 0, len(reputations)),
		Receipts:		make([]ServiceReceipt, 0, len(receipts)),
		Entries:		[]ServiceRegistryStateEntry{},
		UpdatedHeight:		height,
	}
	entries := map[string]ServiceRegistryStateEntry{}
	for _, descriptor := range descriptors {
		descriptor = CanonicalServiceDescriptor(descriptor)
		if err := descriptor.Validate(); err != nil {
			return ServiceRegistryState{}, err
		}
		state.Descriptors = append(state.Descriptors, descriptor)
		descriptorHash := ComputeServiceDescriptorHash(descriptor)
		descriptorKey, err := ServiceDescriptorStateKey(descriptor.ServiceID)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateDescriptor, descriptorKey, descriptorHash); err != nil {
			return ServiceRegistryState{}, err
		}
		interfaceKey, err := ServiceInterfaceStateKey(descriptor.Interface.InterfaceHash)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateInterface, interfaceKey, descriptor.Interface.InterfaceHash); err != nil {
			return ServiceRegistryState{}, err
		}
		state.Interfaces = appendUniqueServiceInterface(state.Interfaces, descriptor.Interface)
		ownerKey, err := ServiceOwnerStateKey(descriptor.Owner, descriptor.ServiceID)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateOwnerIndex, ownerKey, descriptor.ServiceID); err != nil {
			return ServiceRegistryState{}, err
		}
		if descriptor.Discovery.ServiceName != "" {
			nameKey, err := ServiceNameStateKey(descriptor.Discovery.ServiceName)
			if err != nil {
				return ServiceRegistryState{}, err
			}
			if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateNameIndex, nameKey, descriptor.ServiceID); err != nil {
				return ServiceRegistryState{}, err
			}
		}
		if descriptor.Discovery.IdentityName != "" {
			binding, err := NewIdentityServiceBindingFromDescriptor(descriptor)
			if err != nil {
				return ServiceRegistryState{}, err
			}
			identityBindings = append(identityBindings, binding)
		}
		if descriptor.ExpiryHeight != 0 {
			expiryKey, err := ServiceExpiryStateKey(descriptor.ExpiryHeight, descriptor.ServiceID)
			if err != nil {
				return ServiceRegistryState{}, err
			}
			if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateExpiryIndex, expiryKey, descriptor.ServiceID); err != nil {
				return ServiceRegistryState{}, err
			}
		}
	}
	for _, anchor := range anchors {
		anchor = CanonicalServiceAnchor(anchor)
		if err := anchor.Validate(); err != nil {
			return ServiceRegistryState{}, err
		}
		state.Anchors = append(state.Anchors, anchor)
		key, err := ServiceAnchorStateKey(anchor.ServiceID)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateAnchor, key, anchor.AnchorHash); err != nil {
			return ServiceRegistryState{}, err
		}
	}
	for _, binding := range identityBindings {
		binding = CanonicalIdentityServiceBinding(binding)
		if err := binding.Validate(); err != nil {
			return ServiceRegistryState{}, err
		}
		state.IdentityBindings = append(state.IdentityBindings, binding)
		key, err := IdentityServiceBindingStateKey(binding.IdentityName, binding.ServiceID)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateIdentityBinding, key, binding.BindingHash); err != nil {
			return ServiceRegistryState{}, err
		}
	}
	for _, provider := range providers {
		provider = CanonicalProviderRecord(provider)
		if err := provider.Validate(); err != nil {
			return ServiceRegistryState{}, err
		}
		state.Providers = append(state.Providers, provider)
		key, err := ServiceProviderStateKey(provider.ServiceID, provider.Provider.ProviderID)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateProvider, key, provider.RecordHash); err != nil {
			return ServiceRegistryState{}, err
		}
	}
	for _, reputation := range reputations {
		reputation = CanonicalReputationRecord(reputation)
		if err := reputation.Validate(); err != nil {
			return ServiceRegistryState{}, err
		}
		state.Reputations = append(state.Reputations, reputation)
		key, err := ServiceReputationStateKey(reputation.ProviderID)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateReputation, key, reputation.RecordHash); err != nil {
			return ServiceRegistryState{}, err
		}
	}
	for _, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return ServiceRegistryState{}, err
		}
		state.Receipts = append(state.Receipts, receipt)
		key, err := ServiceReceiptStateKey(receipt.ServiceID, receipt.CallID)
		if err != nil {
			return ServiceRegistryState{}, err
		}
		if err := addServiceRegistryStateEntry(entries, ServiceRegistryStateReceipt, key, receipt.ReceiptHash); err != nil {
			return ServiceRegistryState{}, err
		}
	}
	for _, entry := range entries {
		state.Entries = append(state.Entries, entry)
		switch entry.EntryType {
		case ServiceRegistryStateOwnerIndex:
			state.OwnerIndex = append(state.OwnerIndex, entry)
		case ServiceRegistryStateNameIndex:
			state.NameIndex = append(state.NameIndex, entry)
		case ServiceRegistryStateExpiryIndex:
			state.ExpiryIndex = append(state.ExpiryIndex, entry)
		}
	}
	sortServiceRegistryState(state)
	state.StateRoot = ComputeServiceRegistryStateRoot(state)
	return state, state.Validate()
}

func CanonicalServiceAnchor(anchor ServiceAnchor) ServiceAnchor {
	anchor.ServiceID = strings.TrimSpace(anchor.ServiceID)
	anchor.DescriptorHash = strings.ToLower(strings.TrimSpace(anchor.DescriptorHash))
	anchor.InterfaceHash = strings.ToLower(strings.TrimSpace(anchor.InterfaceHash))
	anchor.ProviderRoot = strings.ToLower(strings.TrimSpace(anchor.ProviderRoot))
	anchor.AnchorHash = strings.ToLower(strings.TrimSpace(anchor.AnchorHash))
	if anchor.AnchorHash == "" {
		anchor.AnchorHash = ComputeServiceAnchorHash(anchor)
	}
	return anchor
}

func CanonicalIdentityServiceBinding(binding IdentityServiceBinding) IdentityServiceBinding {
	binding.IdentityName = strings.TrimSpace(binding.IdentityName)
	binding.ServiceID = strings.TrimSpace(binding.ServiceID)
	binding.Owner = strings.TrimSpace(binding.Owner)
	binding.DescriptorHash = strings.ToLower(strings.TrimSpace(binding.DescriptorHash))
	binding.BindingHash = strings.ToLower(strings.TrimSpace(binding.BindingHash))
	if binding.BindingHash == "" {
		binding.BindingHash = ComputeIdentityServiceBindingHash(binding)
	}
	return binding
}

func CanonicalProviderRecord(record ProviderRecord) ProviderRecord {
	record.ServiceID = strings.TrimSpace(record.ServiceID)
	record.Provider = CanonicalFogProviderRecord(record.Provider)
	record.RecordHash = strings.ToLower(strings.TrimSpace(record.RecordHash))
	if record.RecordHash == "" {
		record.RecordHash = ComputeProviderRecordHash(record)
	}
	return record
}

func CanonicalReputationRecord(record ReputationRecord) ReputationRecord {
	record.ProviderID = strings.TrimSpace(record.ProviderID)
	record.RecordHash = strings.ToLower(strings.TrimSpace(record.RecordHash))
	if record.RecordHash == "" {
		record.RecordHash = ComputeReputationRecordHash(record)
	}
	return record
}

func (anchor ServiceAnchor) Validate() error {
	anchor = CanonicalServiceAnchor(anchor)
	if err := validatePolicyID("aetracore service anchor service id", anchor.ServiceID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service anchor descriptor hash", anchor.DescriptorHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service anchor interface hash", anchor.InterfaceHash); err != nil {
		return err
	}
	if anchor.ProviderRoot != "" {
		if err := ValidateHash("aetracore service anchor provider root", anchor.ProviderRoot); err != nil {
			return err
		}
	}
	if !IsServiceVerificationModel(anchor.VerificationModel) {
		return fmt.Errorf("unknown aetracore service anchor verification model %q", anchor.VerificationModel)
	}
	if err := ValidateHash("aetracore service anchor hash", anchor.AnchorHash); err != nil {
		return err
	}
	if expected := ComputeServiceAnchorHash(anchor); anchor.AnchorHash != expected {
		return fmt.Errorf("aetracore service anchor hash mismatch: expected %s", expected)
	}
	return nil
}

func (binding IdentityServiceBinding) Validate() error {
	binding = CanonicalIdentityServiceBinding(binding)
	if err := validatePolicyID("aetracore identity service binding identity name", binding.IdentityName); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore identity service binding service id", binding.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore identity service binding owner", binding.Owner); err != nil {
		return err
	}
	if err := ValidateHash("aetracore identity service binding descriptor hash", binding.DescriptorHash); err != nil {
		return err
	}
	if binding.CreatedHeight == 0 {
		return errors.New("aetracore identity service binding created height must be positive")
	}
	if binding.ExpiryHeight != 0 && binding.ExpiryHeight <= binding.CreatedHeight {
		return errors.New("aetracore identity service binding expiry must exceed created height")
	}
	if err := ValidateHash("aetracore identity service binding hash", binding.BindingHash); err != nil {
		return err
	}
	if expected := ComputeIdentityServiceBindingHash(binding); binding.BindingHash != expected {
		return fmt.Errorf("aetracore identity service binding hash mismatch: expected %s", expected)
	}
	return nil
}

func (record ProviderRecord) Validate() error {
	record = CanonicalProviderRecord(record)
	if err := validatePolicyID("aetracore service provider record service id", record.ServiceID); err != nil {
		return err
	}
	if err := record.Provider.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service provider record hash", record.RecordHash); err != nil {
		return err
	}
	if expected := ComputeProviderRecordHash(record); record.RecordHash != expected {
		return fmt.Errorf("aetracore service provider record hash mismatch: expected %s", expected)
	}
	return nil
}

func (record ReputationRecord) Validate() error {
	record = CanonicalReputationRecord(record)
	if err := validatePolicyID("aetracore reputation provider id", record.ProviderID); err != nil {
		return err
	}
	if record.UpdatedHeight == 0 {
		return errors.New("aetracore reputation record updated height must be positive")
	}
	if err := ValidateHash("aetracore reputation record hash", record.RecordHash); err != nil {
		return err
	}
	if expected := ComputeReputationRecordHash(record); record.RecordHash != expected {
		return fmt.Errorf("aetracore reputation record hash mismatch: expected %s", expected)
	}
	return nil
}

func (entry ServiceRegistryStateEntry) Validate() error {
	if !IsServiceRegistryStateEntryType(entry.EntryType) {
		return fmt.Errorf("unknown aetracore service registry state entry type %q", entry.EntryType)
	}
	if err := validateServiceRegistryStateKey(entry.Key); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service registry state entry value", entry.Value); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service registry state entry hash", entry.EntryHash); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryStateEntryHash(entry); entry.EntryHash != expected {
		return fmt.Errorf("aetracore service registry state entry hash mismatch: expected %s", expected)
	}
	return nil
}

func (state ServiceRegistryState) Validate() error {
	if state.UpdatedHeight == 0 {
		return errors.New("aetracore service registry state height must be positive")
	}
	if err := validateServiceRegistryStateEntries(state.Entries); err != nil {
		return err
	}
	for _, descriptor := range state.Descriptors {
		descriptor = CanonicalServiceDescriptor(descriptor)
		if err := descriptor.Validate(); err != nil {
			return err
		}
	}
	for _, anchor := range state.Anchors {
		if err := anchor.Validate(); err != nil {
			return err
		}
	}
	for _, iface := range state.Interfaces {
		if err := iface.Validate(); err != nil {
			return err
		}
	}
	for _, binding := range state.IdentityBindings {
		if err := binding.Validate(); err != nil {
			return err
		}
	}
	for _, provider := range state.Providers {
		if err := provider.Validate(); err != nil {
			return err
		}
	}
	for _, reputation := range state.Reputations {
		if err := reputation.Validate(); err != nil {
			return err
		}
	}
	for _, receipt := range state.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore service registry state root", state.StateRoot); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryStateRoot(state); state.StateRoot != expected {
		return fmt.Errorf("aetracore service registry state root mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceAnchorHash(anchor ServiceAnchor) string {
	anchor.AnchorHash = ""
	return hashParts(
		"aetra-aek-service-anchor-v1",
		anchor.ServiceID,
		anchor.DescriptorHash,
		anchor.InterfaceHash,
		anchor.ProviderRoot,
		string(anchor.VerificationModel),
		fmt.Sprint(anchor.ExpiryHeight),
	)
}

func ComputeIdentityServiceBindingHash(binding IdentityServiceBinding) string {
	binding.BindingHash = ""
	return hashParts(
		"aetra-aek-identity-service-binding-v1",
		binding.IdentityName,
		binding.ServiceID,
		binding.Owner,
		binding.DescriptorHash,
		fmt.Sprint(binding.CreatedHeight),
		fmt.Sprint(binding.ExpiryHeight),
	)
}

func ComputeProviderRecordHash(record ProviderRecord) string {
	record.RecordHash = ""
	return hashParts(
		"aetra-aek-service-provider-record-v1",
		record.ServiceID,
		record.Provider.ProviderID,
		record.Provider.ProviderHash,
	)
}

func ComputeReputationRecordHash(record ReputationRecord) string {
	record.RecordHash = ""
	return hashParts(
		"aetra-aek-reputation-record-v1",
		record.ProviderID,
		fmt.Sprint(record.Score),
		fmt.Sprint(record.Successes),
		fmt.Sprint(record.Failures),
		fmt.Sprint(record.UpdatedHeight),
	)
}

func ComputeServiceRegistryStateEntryHash(entry ServiceRegistryStateEntry) string {
	entry.EntryHash = ""
	return hashParts(
		"aetra-aek-service-registry-state-entry-v1",
		string(entry.EntryType),
		entry.Key,
		entry.Value,
	)
}

func ComputeServiceRegistryStateRoot(state ServiceRegistryState) string {
	entries := append([]ServiceRegistryStateEntry(nil), state.Entries...)
	sortServiceRegistryStateEntries(entries)
	parts := []string{
		"aetra-aek-service-registry-state-root-v1",
		fmt.Sprint(state.UpdatedHeight),
		fmt.Sprint(len(entries)),
	}
	for _, entry := range entries {
		parts = append(parts, entry.EntryHash)
	}
	return hashParts(parts...)
}

func IsServiceRegistryStateEntryType(entryType ServiceRegistryStateEntryType) bool {
	switch entryType {
	case ServiceRegistryStateDescriptor, ServiceRegistryStateAnchor, ServiceRegistryStateInterface,
		ServiceRegistryStateOwnerIndex, ServiceRegistryStateNameIndex, ServiceRegistryStateIdentityBinding,
		ServiceRegistryStateProvider, ServiceRegistryStateExpiryIndex, ServiceRegistryStateReputation,
		ServiceRegistryStateReceipt:
		return true
	default:
		return false
	}
}

func addServiceRegistryStateEntry(entries map[string]ServiceRegistryStateEntry, entryType ServiceRegistryStateEntryType, key, value string) error {
	entry := ServiceRegistryStateEntry{
		Key:		key,
		Value:		value,
		EntryType:	entryType,
	}
	entry.EntryHash = ComputeServiceRegistryStateEntryHash(entry)
	if err := entry.Validate(); err != nil {
		return err
	}
	existing, found := entries[key]
	if found {
		if existing.Value != entry.Value || existing.EntryType != entry.EntryType {
			return fmt.Errorf("aetracore service registry state key collision %s", key)
		}
		return nil
	}
	entries[key] = entry
	return nil
}

func validateServiceRegistryStateKey(key string) error {
	if err := validatePolicyID("aetracore service registry state key", key); err != nil {
		return err
	}
	if !strings.HasPrefix(key, "services/") {
		return errors.New("aetracore service registry state key must use services prefix")
	}
	if strings.Contains(key, "//") {
		return errors.New("aetracore service registry state key contains empty segment")
	}
	return nil
}

func validateServiceRegistryStateEntries(entries []ServiceRegistryStateEntry) error {
	var previous string
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seen[entry.Key]; found {
			return fmt.Errorf("duplicate aetracore service registry state key %s", entry.Key)
		}
		seen[entry.Key] = struct{}{}
		if previous != "" && previous >= entry.Key {
			return errors.New("aetracore service registry state entries must be sorted canonically")
		}
		previous = entry.Key
	}
	return nil
}

func appendUniqueServiceInterface(interfaces []ServiceInterface, iface ServiceInterface) []ServiceInterface {
	for _, existing := range interfaces {
		if existing.InterfaceHash == iface.InterfaceHash {
			return interfaces
		}
	}
	return append(interfaces, iface)
}

func sortServiceRegistryState(state ServiceRegistryState) {
	sortServiceDescriptors(state.Descriptors)
	sortServiceAnchors(state.Anchors)
	sortServiceInterfaces(state.Interfaces)
	sortIdentityServiceBindings(state.IdentityBindings)
	sortProviderRecords(state.Providers)
	sortReputationRecords(state.Reputations)
	sortServiceCallReceipts(state.Receipts)
	sortServiceRegistryStateEntries(state.OwnerIndex)
	sortServiceRegistryStateEntries(state.NameIndex)
	sortServiceRegistryStateEntries(state.ExpiryIndex)
	sortServiceRegistryStateEntries(state.Entries)
}

func sortServiceAnchors(anchors []ServiceAnchor) {
	sort.SliceStable(anchors, func(i, j int) bool { return anchors[i].ServiceID < anchors[j].ServiceID })
}

func sortServiceInterfaces(interfaces []ServiceInterface) {
	sort.SliceStable(interfaces, func(i, j int) bool { return interfaces[i].InterfaceHash < interfaces[j].InterfaceHash })
}

func sortIdentityServiceBindings(bindings []IdentityServiceBinding) {
	sort.SliceStable(bindings, func(i, j int) bool {
		left := bindings[i].IdentityName + "/" + bindings[i].ServiceID
		right := bindings[j].IdentityName + "/" + bindings[j].ServiceID
		return left < right
	})
}

func sortProviderRecords(records []ProviderRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := records[i].ServiceID + "/" + records[i].Provider.ProviderID
		right := records[j].ServiceID + "/" + records[j].Provider.ProviderID
		return left < right
	})
}

func sortReputationRecords(records []ReputationRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].ProviderID < records[j].ProviderID })
}

func sortServiceRegistryStateEntries(entries []ServiceRegistryStateEntry) {
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
}
