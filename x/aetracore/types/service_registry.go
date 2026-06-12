package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type ServiceRegistryMode string

const (
	ServiceRegistryOnChain		ServiceRegistryMode	= "ON_CHAIN_REGISTRY"
	ServiceRegistryHybrid		ServiceRegistryMode	= "HYBRID_REGISTRY"
	ServiceRegistryMesh		ServiceRegistryMode	= "DISTRIBUTED_MESH_REGISTRY"
	ServiceRegistryMeshModeKey				= "mesh"
)

type ServiceRegistry struct {
	Mode		ServiceRegistryMode
	Records		[]ServiceRegistryRecord
	InterfaceIndex	[]ServiceInterfaceVersionRecord
	RegistryRoot	string
	UpdatedHeight	uint64
}

type ServiceRegistryRecord struct {
	ServiceID		string
	Owner			string
	ServiceType		ServiceType
	InterfaceHash		string
	Endpoint		string
	ExecutionMode		ExecutionMode
	PaymentModel		string
	Stake			string
	Collateral		string
	Reputation		uint64
	ExpiryHeight		uint64
	ServiceName		string
	IdentityName		string
	ZoneID			ZoneID
	ContractAddress		string
	ModuleRoute		string
	ProviderSet		string
	VerificationModel	ServiceVerificationModel
	TrustModel		ServiceTrustModel
	Status			ServiceStatus
	Version			uint64
	DescriptorHash		string
	MetadataHash		string
	CreatedHeight		uint64
	UpdatedHeight		uint64
	RecordHash		string
}

type ServiceInterfaceVersionRecord struct {
	InterfaceID	string
	Version		uint64
	InterfaceHash	string
	ServiceIDs	[]string
	MetadataHash	string
	CreatedHeight	uint64
	RecordHash	string
}

type ServiceRegistryProof struct {
	ServiceID	string
	RegistryMode	ServiceRegistryMode
	RegistryRoot	string
	RecordHash	string
	DescriptorHash	string
	InterfaceHash	string
	ProofHeight	uint64
	ProofHash	string
}

type ServicePaymentDiscovery struct {
	ServiceID	string
	PaymentModel	string
	SettlementMode	ServicePaymentSettlementMode
	Denom		string
	Amount		string
	MaxAmount	string
	PricingUnit	ServicePricingUnit
	EscrowRequired	bool
	EscrowID	string
	MeterID		string
	DiscoveryHash	string
}

type ServiceTrustDiscovery struct {
	ServiceID		string
	TrustModel		ServiceTrustModel
	VerificationModel	ServiceVerificationModel
	ProofFormat		string
	ChallengeWindow		uint64
	CollateralDenom		string
	CollateralAmount	string
	FallbackServiceID	string
	DiscoveryHash		string
}

func NewServiceRegistry(mode ServiceRegistryMode, descriptors []ServiceDescriptor, height uint64) (ServiceRegistry, error) {
	if !IsServiceRegistryMode(mode) {
		return ServiceRegistry{}, fmt.Errorf("unknown aetracore service registry mode %q", mode)
	}
	if height == 0 {
		return ServiceRegistry{}, errors.New("aetracore service registry height must be positive")
	}
	registry := ServiceRegistry{
		Mode:		mode,
		Records:	make([]ServiceRegistryRecord, 0, len(descriptors)),
		UpdatedHeight:	height,
	}
	for _, descriptor := range descriptors {
		record, err := NewServiceRegistryRecord(descriptor)
		if err != nil {
			return ServiceRegistry{}, err
		}
		registry.Records = append(registry.Records, record)
	}
	sortServiceRegistryRecords(registry.Records)
	index, err := BuildServiceInterfaceVersionIndex(registry.Records)
	if err != nil {
		return ServiceRegistry{}, err
	}
	registry.InterfaceIndex = index
	registry.RegistryRoot = ComputeServiceRegistryRoot(registry)
	return registry, registry.Validate()
}

func RegisterServiceRegistryRecord(registry ServiceRegistry, record ServiceRegistryRecord, height uint64) (ServiceRegistry, error) {
	if err := registry.Validate(); err != nil {
		return ServiceRegistry{}, err
	}
	if height < registry.UpdatedHeight {
		return ServiceRegistry{}, errors.New("aetracore service registry update height must not go backwards")
	}
	record = CanonicalServiceRegistryRecord(record)
	if err := record.Validate(); err != nil {
		return ServiceRegistry{}, err
	}
	if _, found := registry.RecordByID(record.ServiceID); found {
		return ServiceRegistry{}, fmt.Errorf("aetracore service registry record %s already exists", record.ServiceID)
	}
	next := registry.clone()
	next.Records = append(next.Records, record)
	return finalizeServiceRegistry(next, height)
}

func UpdateServiceRegistryRecord(registry ServiceRegistry, record ServiceRegistryRecord, height uint64) (ServiceRegistry, error) {
	if err := registry.Validate(); err != nil {
		return ServiceRegistry{}, err
	}
	if height < registry.UpdatedHeight {
		return ServiceRegistry{}, errors.New("aetracore service registry update height must not go backwards")
	}
	record = CanonicalServiceRegistryRecord(record)
	if err := record.Validate(); err != nil {
		return ServiceRegistry{}, err
	}
	next := registry.clone()
	for i := range next.Records {
		if next.Records[i].ServiceID == record.ServiceID {
			if record.Version < next.Records[i].Version {
				return ServiceRegistry{}, errors.New("aetracore service registry version must not decrease")
			}
			next.Records[i] = record
			return finalizeServiceRegistry(next, height)
		}
	}
	return ServiceRegistry{}, fmt.Errorf("aetracore service registry record %s not found", record.ServiceID)
}

func RenewServiceRegistryRecord(registry ServiceRegistry, serviceID string, expiryHeight uint64, updatedHeight uint64) (ServiceRegistry, ServiceRegistryRecord, error) {
	record, found := registry.RecordByID(serviceID)
	if !found {
		return ServiceRegistry{}, ServiceRegistryRecord{}, fmt.Errorf("aetracore service registry record %s not found", serviceID)
	}
	if expiryHeight <= record.ExpiryHeight {
		return ServiceRegistry{}, ServiceRegistryRecord{}, errors.New("aetracore service registry renewal must extend expiry")
	}
	if updatedHeight < record.UpdatedHeight {
		return ServiceRegistry{}, ServiceRegistryRecord{}, errors.New("aetracore service registry renewal height must not go backwards")
	}
	record.ExpiryHeight = expiryHeight
	record.UpdatedHeight = updatedHeight
	record.Status = ServiceStatusActive
	record.RecordHash = ComputeServiceRegistryRecordHash(record)
	next, err := UpdateServiceRegistryRecord(registry, record, updatedHeight)
	if err != nil {
		return ServiceRegistry{}, ServiceRegistryRecord{}, err
	}
	return next, record, nil
}

func ExpireServiceRegistryRecords(registry ServiceRegistry, height uint64) (ServiceRegistry, []string, error) {
	if err := registry.Validate(); err != nil {
		return ServiceRegistry{}, nil, err
	}
	if height == 0 || height < registry.UpdatedHeight {
		return ServiceRegistry{}, nil, errors.New("aetracore service registry expiry height is invalid")
	}
	next := registry.clone()
	expired := make([]string, 0)
	for i := range next.Records {
		if next.Records[i].ExpiryHeight != 0 && next.Records[i].ExpiryHeight <= height && next.Records[i].Status == ServiceStatusActive {
			next.Records[i].Status = ServiceStatusDisabled
			next.Records[i].UpdatedHeight = height
			next.Records[i].RecordHash = ComputeServiceRegistryRecordHash(next.Records[i])
			expired = append(expired, next.Records[i].ServiceID)
		}
	}
	sort.Strings(expired)
	next, err := finalizeServiceRegistry(next, height)
	if err != nil {
		return ServiceRegistry{}, nil, err
	}
	return next, expired, nil
}

func NewServiceRegistryRecord(descriptor ServiceDescriptor) (ServiceRegistryRecord, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceRegistryRecord{}, err
	}
	record := ServiceRegistryRecord{
		ServiceID:		descriptor.ServiceID,
		Owner:			descriptor.Owner,
		ServiceType:		descriptor.ServiceType,
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		Endpoint:		registryEndpoint(descriptor),
		ExecutionMode:		descriptor.Execution.Mode,
		PaymentModel:		registryPaymentModel(descriptor),
		Stake:			registryStake(descriptor),
		Collateral:		registryCollateral(descriptor),
		Reputation:		0,
		ExpiryHeight:		descriptor.ExpiryHeight,
		ServiceName:		descriptor.Discovery.ServiceName,
		IdentityName:		descriptor.Discovery.IdentityName,
		ZoneID:			descriptor.ZoneID,
		ContractAddress:	descriptor.Execution.ContractAddress,
		ModuleRoute:		descriptor.Execution.ModuleRoute,
		ProviderSet:		registryProviderSet(descriptor),
		VerificationModel:	descriptor.Verification.Model,
		TrustModel:		descriptor.Verification.TrustModel,
		Status:			descriptor.Status,
		Version:		descriptor.Version,
		DescriptorHash:		ComputeServiceDescriptorHash(descriptor),
		MetadataHash:		descriptor.Discovery.MetadataHash,
		CreatedHeight:		descriptor.CreatedHeight,
		UpdatedHeight:		descriptor.UpdatedHeight,
	}
	record.RecordHash = ComputeServiceRegistryRecordHash(record)
	return record, record.Validate()
}

func CanonicalServiceRegistryRecord(record ServiceRegistryRecord) ServiceRegistryRecord {
	record.ServiceID = strings.TrimSpace(record.ServiceID)
	record.Owner = strings.TrimSpace(record.Owner)
	record.InterfaceHash = strings.ToLower(strings.TrimSpace(record.InterfaceHash))
	record.Endpoint = strings.TrimSpace(record.Endpoint)
	record.PaymentModel = strings.TrimSpace(record.PaymentModel)
	record.Stake = strings.TrimSpace(record.Stake)
	record.Collateral = strings.TrimSpace(record.Collateral)
	record.ServiceName = strings.TrimSpace(record.ServiceName)
	record.IdentityName = strings.TrimSpace(record.IdentityName)
	record.ContractAddress = strings.TrimSpace(record.ContractAddress)
	record.ModuleRoute = strings.TrimSpace(record.ModuleRoute)
	record.ProviderSet = strings.TrimSpace(record.ProviderSet)
	record.DescriptorHash = strings.ToLower(strings.TrimSpace(record.DescriptorHash))
	record.MetadataHash = strings.ToLower(strings.TrimSpace(record.MetadataHash))
	record.RecordHash = strings.ToLower(strings.TrimSpace(record.RecordHash))
	if record.Stake == "" {
		record.Stake = "0"
	}
	if record.Collateral == "" {
		record.Collateral = "0"
	}
	if record.RecordHash == "" {
		record.RecordHash = ComputeServiceRegistryRecordHash(record)
	}
	return record
}

func (record ServiceRegistryRecord) Validate() error {
	record = CanonicalServiceRegistryRecord(record)
	if err := validatePolicyID("aetracore service registry service id", record.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore service registry owner", record.Owner); err != nil {
		return err
	}
	if !IsServiceType(record.ServiceType) {
		return fmt.Errorf("unknown aetracore service registry type %q", record.ServiceType)
	}
	if err := ValidateHash("aetracore service registry interface hash", record.InterfaceHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service registry endpoint", record.Endpoint); err != nil {
		return err
	}
	if record.ExecutionMode != ExecutionModeSync && record.ExecutionMode != ExecutionModeAsync {
		return fmt.Errorf("unknown aetracore service registry execution mode %q", record.ExecutionMode)
	}
	if err := validatePolicyID("aetracore service registry payment model", record.PaymentModel); err != nil {
		return err
	}
	if err := validateAmountString("aetracore service registry stake", record.Stake); err != nil {
		return err
	}
	if err := validateAmountString("aetracore service registry collateral", record.Collateral); err != nil {
		return err
	}
	if record.ServiceName != "" {
		if err := validatePolicyID("aetracore service registry service name", record.ServiceName); err != nil {
			return err
		}
	}
	if record.IdentityName != "" {
		if err := validatePolicyID("aetracore service registry identity name", record.IdentityName); err != nil {
			return err
		}
	}
	if record.ZoneID != "" {
		if err := ValidateZoneID(record.ZoneID); err != nil {
			return err
		}
	}
	if record.ContractAddress != "" {
		if err := validatePolicyID("aetracore service registry contract address", record.ContractAddress); err != nil {
			return err
		}
	}
	if record.ModuleRoute != "" {
		if err := validateModuleName(record.ModuleRoute); err != nil {
			return err
		}
	}
	if record.ProviderSet != "" {
		if err := validatePolicyID("aetracore service registry provider set", record.ProviderSet); err != nil {
			return err
		}
	}
	if !IsServiceVerificationModel(record.VerificationModel) {
		return fmt.Errorf("unknown aetracore service registry verification model %q", record.VerificationModel)
	}
	if !IsServiceTrustModel(record.TrustModel) {
		return fmt.Errorf("unknown aetracore service registry trust model %q", record.TrustModel)
	}
	if !IsServiceStatus(record.Status) {
		return fmt.Errorf("unknown aetracore service registry status %q", record.Status)
	}
	if record.Version == 0 {
		return errors.New("aetracore service registry version must be positive")
	}
	if err := ValidateHash("aetracore service registry descriptor hash", record.DescriptorHash); err != nil {
		return err
	}
	if err := validateOptionalHash("aetracore service registry metadata hash", record.MetadataHash); err != nil {
		return err
	}
	if record.CreatedHeight == 0 || record.UpdatedHeight < record.CreatedHeight {
		return errors.New("aetracore service registry heights are invalid")
	}
	if record.ExpiryHeight != 0 && record.ExpiryHeight <= record.UpdatedHeight && record.Status == ServiceStatusActive {
		return errors.New("aetracore active service registry record must not be expired")
	}
	if err := ValidateHash("aetracore service registry record hash", record.RecordHash); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryRecordHash(record); record.RecordHash != expected {
		return fmt.Errorf("aetracore service registry record hash mismatch: expected %s", expected)
	}
	return nil
}

func (registry ServiceRegistry) Validate() error {
	if !IsServiceRegistryMode(registry.Mode) {
		return fmt.Errorf("unknown aetracore service registry mode %q", registry.Mode)
	}
	if registry.UpdatedHeight == 0 {
		return errors.New("aetracore service registry updated height must be positive")
	}
	if err := validateServiceRegistryRecords(registry.Records); err != nil {
		return err
	}
	if err := validateServiceInterfaceVersionRecords(registry.InterfaceIndex); err != nil {
		return err
	}
	index, err := BuildServiceInterfaceVersionIndex(registry.Records)
	if err != nil {
		return err
	}
	if ComputeServiceInterfaceVersionIndexHash(index) != ComputeServiceInterfaceVersionIndexHash(registry.InterfaceIndex) {
		return errors.New("aetracore service registry interface index mismatch")
	}
	if err := ValidateHash("aetracore service registry root", registry.RegistryRoot); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryRoot(registry); registry.RegistryRoot != expected {
		return fmt.Errorf("aetracore service registry root mismatch: expected %s", expected)
	}
	return nil
}

func (registry ServiceRegistry) RecordByID(serviceID string) (ServiceRegistryRecord, bool) {
	for _, record := range registry.Records {
		if record.ServiceID == serviceID {
			return record, true
		}
	}
	return ServiceRegistryRecord{}, false
}

func (registry ServiceRegistry) Lookup(serviceID string) (ServiceRegistryRecord, ServiceRegistryProof, bool) {
	record, found := registry.RecordByID(serviceID)
	if !found {
		return ServiceRegistryRecord{}, ServiceRegistryProof{}, false
	}
	proof := ServiceRegistryProof{
		ServiceID:	record.ServiceID,
		RegistryMode:	registry.Mode,
		RegistryRoot:	registry.RegistryRoot,
		RecordHash:	record.RecordHash,
		DescriptorHash:	record.DescriptorHash,
		InterfaceHash:	record.InterfaceHash,
		ProofHeight:	registry.UpdatedHeight,
	}
	proof.ProofHash = ComputeServiceRegistryProofHash(proof)
	return record, proof, true
}

func (proof ServiceRegistryProof) ValidateForRegistry(registry ServiceRegistry) error {
	if err := registry.Validate(); err != nil {
		return err
	}
	record, found := registry.RecordByID(proof.ServiceID)
	if !found {
		return fmt.Errorf("aetracore service registry proof references unknown service %s", proof.ServiceID)
	}
	if proof.RegistryMode != registry.Mode || proof.RegistryRoot != registry.RegistryRoot ||
		proof.RecordHash != record.RecordHash || proof.DescriptorHash != record.DescriptorHash ||
		proof.InterfaceHash != record.InterfaceHash || proof.ProofHeight != registry.UpdatedHeight {
		return errors.New("aetracore service registry proof does not match registry")
	}
	return proof.Validate()
}

func (proof ServiceRegistryProof) Validate() error {
	if err := validatePolicyID("aetracore service registry proof service id", proof.ServiceID); err != nil {
		return err
	}
	if !IsServiceRegistryMode(proof.RegistryMode) {
		return fmt.Errorf("unknown aetracore service registry proof mode %q", proof.RegistryMode)
	}
	if err := ValidateHash("aetracore service registry proof root", proof.RegistryRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service registry proof record hash", proof.RecordHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service registry proof descriptor hash", proof.DescriptorHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service registry proof interface hash", proof.InterfaceHash); err != nil {
		return err
	}
	if proof.ProofHeight == 0 {
		return errors.New("aetracore service registry proof height must be positive")
	}
	if err := ValidateHash("aetracore service registry proof hash", proof.ProofHash); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryProofHash(proof); proof.ProofHash != expected {
		return fmt.Errorf("aetracore service registry proof hash mismatch: expected %s", expected)
	}
	return nil
}

func (registry ServiceRegistry) PaymentDiscovery(serviceID string) (ServicePaymentDiscovery, bool) {
	record, found := registry.RecordByID(serviceID)
	if !found {
		return ServicePaymentDiscovery{}, false
	}
	discovery := ServicePaymentDiscovery{
		ServiceID:	record.ServiceID,
		PaymentModel:	record.PaymentModel,
	}
	parts := strings.Split(record.PaymentModel, ":")
	if len(parts) >= 4 {
		discovery.SettlementMode = ServicePaymentSettlementMode(parts[0])
		discovery.Denom = parts[1]
		discovery.Amount = parts[2]
		discovery.PricingUnit = ServicePricingUnit(parts[3])
	}
	if len(parts) >= 5 {
		discovery.MaxAmount = parts[4]
	}
	if len(parts) >= 6 {
		discovery.EscrowRequired = parts[5] == "escrow"
	}
	if len(parts) >= 7 {
		discovery.EscrowID = parts[6]
	}
	if len(parts) >= 8 {
		discovery.MeterID = parts[7]
	}
	discovery.DiscoveryHash = ComputeServicePaymentDiscoveryHash(discovery)
	return discovery, true
}

func (registry ServiceRegistry) TrustDiscovery(serviceID string) (ServiceTrustDiscovery, bool) {
	record, found := registry.RecordByID(serviceID)
	if !found {
		return ServiceTrustDiscovery{}, false
	}
	discovery := ServiceTrustDiscovery{
		ServiceID:		record.ServiceID,
		TrustModel:		record.TrustModel,
		VerificationModel:	record.VerificationModel,
		CollateralAmount:	record.Collateral,
	}
	discovery.DiscoveryHash = ComputeServiceTrustDiscoveryHash(discovery)
	return discovery, true
}

func BuildServiceInterfaceVersionIndex(records []ServiceRegistryRecord) ([]ServiceInterfaceVersionRecord, error) {
	byKey := make(map[string]*ServiceInterfaceVersionRecord)
	for _, record := range records {
		record = CanonicalServiceRegistryRecord(record)
		if err := record.Validate(); err != nil {
			return nil, err
		}
		key := record.InterfaceHash + "/" + fmt.Sprint(record.Version)
		entry, found := byKey[key]
		if !found {
			entry = &ServiceInterfaceVersionRecord{
				InterfaceID:	record.InterfaceHash,
				Version:	record.Version,
				InterfaceHash:	record.InterfaceHash,
				ServiceIDs:	[]string{},
				MetadataHash:	record.MetadataHash,
				CreatedHeight:	record.CreatedHeight,
			}
			byKey[key] = entry
		}
		entry.ServiceIDs = append(entry.ServiceIDs, record.ServiceID)
		if entry.CreatedHeight > record.CreatedHeight {
			entry.CreatedHeight = record.CreatedHeight
		}
	}
	out := make([]ServiceInterfaceVersionRecord, 0, len(byKey))
	for _, entry := range byKey {
		sort.Strings(entry.ServiceIDs)
		entry.RecordHash = ComputeServiceInterfaceVersionRecordHash(*entry)
		out = append(out, *entry)
	}
	sortServiceInterfaceVersionRecords(out)
	return out, nil
}

func (record ServiceInterfaceVersionRecord) Validate() error {
	if err := ValidateHash("aetracore service interface version id", record.InterfaceID); err != nil {
		return err
	}
	if record.Version == 0 {
		return errors.New("aetracore service interface version must be positive")
	}
	if err := ValidateHash("aetracore service interface version hash", record.InterfaceHash); err != nil {
		return err
	}
	if err := validateSortedStringSet("aetracore service interface version service", record.ServiceIDs); err != nil {
		return err
	}
	if len(record.ServiceIDs) == 0 {
		return errors.New("aetracore service interface version requires services")
	}
	if err := validateOptionalHash("aetracore service interface version metadata", record.MetadataHash); err != nil {
		return err
	}
	if record.CreatedHeight == 0 {
		return errors.New("aetracore service interface version created height must be positive")
	}
	if err := ValidateHash("aetracore service interface version record hash", record.RecordHash); err != nil {
		return err
	}
	if expected := ComputeServiceInterfaceVersionRecordHash(record); record.RecordHash != expected {
		return fmt.Errorf("aetracore service interface version hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceRegistryRecordHash(record ServiceRegistryRecord) string {
	record.RecordHash = ""
	return hashParts(
		"aetra-aek-service-registry-record-v1",
		record.ServiceID,
		record.Owner,
		string(record.ServiceType),
		record.InterfaceHash,
		record.Endpoint,
		string(record.ExecutionMode),
		record.PaymentModel,
		record.Stake,
		record.Collateral,
		fmt.Sprint(record.Reputation),
		fmt.Sprint(record.ExpiryHeight),
		record.ServiceName,
		record.IdentityName,
		string(record.ZoneID),
		record.ContractAddress,
		record.ModuleRoute,
		record.ProviderSet,
		string(record.VerificationModel),
		string(record.TrustModel),
		string(record.Status),
		fmt.Sprint(record.Version),
		record.DescriptorHash,
		record.MetadataHash,
		fmt.Sprint(record.CreatedHeight),
		fmt.Sprint(record.UpdatedHeight),
	)
}

func ComputeServiceRegistryRoot(registry ServiceRegistry) string {
	records := append([]ServiceRegistryRecord(nil), registry.Records...)
	sortServiceRegistryRecords(records)
	interfaces := append([]ServiceInterfaceVersionRecord(nil), registry.InterfaceIndex...)
	sortServiceInterfaceVersionRecords(interfaces)
	parts := []string{
		"aetra-aek-service-registry-root-v1",
		string(registry.Mode),
		fmt.Sprint(registry.UpdatedHeight),
		fmt.Sprint(len(records)),
	}
	for _, record := range records {
		parts = append(parts, record.RecordHash)
	}
	parts = append(parts, ComputeServiceInterfaceVersionIndexHash(interfaces))
	return hashParts(parts...)
}

func ComputeServiceRegistryProofHash(proof ServiceRegistryProof) string {
	return hashParts(
		"aetra-aek-service-registry-proof-v1",
		proof.ServiceID,
		string(proof.RegistryMode),
		proof.RegistryRoot,
		proof.RecordHash,
		proof.DescriptorHash,
		proof.InterfaceHash,
		fmt.Sprint(proof.ProofHeight),
	)
}

func ComputeServiceInterfaceVersionRecordHash(record ServiceInterfaceVersionRecord) string {
	parts := []string{
		"aetra-aek-service-interface-version-v1",
		record.InterfaceID,
		fmt.Sprint(record.Version),
		record.InterfaceHash,
		record.MetadataHash,
		fmt.Sprint(record.CreatedHeight),
	}
	parts = appendStringSliceParts(parts, "services", record.ServiceIDs)
	return hashParts(parts...)
}

func ComputeServiceInterfaceVersionIndexHash(records []ServiceInterfaceVersionRecord) string {
	ordered := append([]ServiceInterfaceVersionRecord(nil), records...)
	sortServiceInterfaceVersionRecords(ordered)
	parts := []string{"aetra-aek-service-interface-index-v1", fmt.Sprint(len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.RecordHash)
	}
	return hashParts(parts...)
}

func ComputeServicePaymentDiscoveryHash(discovery ServicePaymentDiscovery) string {
	return hashParts(
		"aetra-aek-service-payment-discovery-v1",
		discovery.ServiceID,
		discovery.PaymentModel,
		string(discovery.SettlementMode),
		discovery.Denom,
		discovery.Amount,
		discovery.MaxAmount,
		string(discovery.PricingUnit),
		fmt.Sprint(discovery.EscrowRequired),
		discovery.EscrowID,
		discovery.MeterID,
	)
}

func ComputeServiceTrustDiscoveryHash(discovery ServiceTrustDiscovery) string {
	return hashParts(
		"aetra-aek-service-trust-discovery-v1",
		discovery.ServiceID,
		string(discovery.TrustModel),
		string(discovery.VerificationModel),
		discovery.ProofFormat,
		fmt.Sprint(discovery.ChallengeWindow),
		discovery.CollateralDenom,
		discovery.CollateralAmount,
		discovery.FallbackServiceID,
	)
}

func IsServiceRegistryMode(mode ServiceRegistryMode) bool {
	switch mode {
	case ServiceRegistryOnChain, ServiceRegistryHybrid, ServiceRegistryMesh:
		return true
	default:
		return false
	}
}

func validateServiceRegistryRecords(records []ServiceRegistryRecord) error {
	var previous string
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		record = CanonicalServiceRegistryRecord(record)
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.ServiceID]; found {
			return fmt.Errorf("duplicate aetracore service registry record %s", record.ServiceID)
		}
		seen[record.ServiceID] = struct{}{}
		if previous != "" && previous >= record.ServiceID {
			return errors.New("aetracore service registry records must be sorted canonically")
		}
		previous = record.ServiceID
	}
	return nil
}

func validateServiceInterfaceVersionRecords(records []ServiceInterfaceVersionRecord) error {
	var previous string
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		key := record.InterfaceHash + "/" + fmt.Sprint(record.Version)
		if previous != "" && previous >= key {
			return errors.New("aetracore service interface version records must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func sortServiceRegistryRecords(records []ServiceRegistryRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].ServiceID < records[j].ServiceID })
}

func sortServiceInterfaceVersionRecords(records []ServiceInterfaceVersionRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := records[i].InterfaceHash + "/" + fmt.Sprint(records[i].Version)
		right := records[j].InterfaceHash + "/" + fmt.Sprint(records[j].Version)
		return left < right
	})
}

func finalizeServiceRegistry(registry ServiceRegistry, height uint64) (ServiceRegistry, error) {
	sortServiceRegistryRecords(registry.Records)
	index, err := BuildServiceInterfaceVersionIndex(registry.Records)
	if err != nil {
		return ServiceRegistry{}, err
	}
	registry.InterfaceIndex = index
	registry.UpdatedHeight = height
	registry.RegistryRoot = ComputeServiceRegistryRoot(registry)
	return registry, registry.Validate()
}

func (registry ServiceRegistry) clone() ServiceRegistry {
	registry.Records = append([]ServiceRegistryRecord(nil), registry.Records...)
	registry.InterfaceIndex = append([]ServiceInterfaceVersionRecord(nil), registry.InterfaceIndex...)
	return registry
}

func registryEndpoint(descriptor ServiceDescriptor) string {
	switch {
	case descriptor.Execution.Endpoint != "":
		return descriptor.Execution.Endpoint
	case descriptor.Execution.ContractAddress != "":
		return descriptor.Execution.ContractAddress
	case descriptor.Execution.ModuleRoute != "":
		return descriptor.Execution.ModuleRoute
	case descriptor.Execution.ProviderPoolID != "":
		return descriptor.Execution.ProviderPoolID
	case descriptor.Execution.Target != "":
		return descriptor.Execution.Target
	default:
		return descriptor.EndpointKey
	}
}

func registryPaymentModel(descriptor ServiceDescriptor) string {
	escrow := "no-escrow"
	if descriptor.Payment.EscrowRequired {
		escrow = "escrow"
	}
	return strings.Join([]string{
		string(descriptor.Payment.SettlementMode),
		descriptor.Payment.Denom,
		descriptor.Payment.Amount,
		string(descriptor.Payment.PricingUnit),
		descriptor.Payment.MaxAmount,
		escrow,
		descriptor.Payment.EscrowID,
		descriptor.Payment.MeterID,
	}, ":")
}

func registryStake(descriptor ServiceDescriptor) string {
	if descriptor.Verification.ProviderCollateralAmount != "" && descriptor.ServiceType == ServiceTypeFogMarket {
		return descriptor.Verification.ProviderCollateralAmount
	}
	return "0"
}

func registryCollateral(descriptor ServiceDescriptor) string {
	if descriptor.Verification.ProviderCollateralAmount != "" {
		return descriptor.Verification.ProviderCollateralAmount
	}
	return "0"
}

func registryProviderSet(descriptor ServiceDescriptor) string {
	if descriptor.Discovery.ProviderRoot != "" {
		return descriptor.Discovery.ProviderRoot
	}
	return descriptor.Execution.ProviderPoolID
}
