package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

const (
	ServiceStateDependencyCommitted		= "committed"
	ServiceStateDependencyZoneRoot		= "zone_root"
	ServiceStateDependencyModuleRoot	= "module_root"
	ServiceStateDependencyGlobalRoot	= "global_root"
	ServiceStateDependencyServiceRoot	= "service_root"
	ServiceStateDependencyIdentityRoot	= "identity_root"
	ServiceStateDependencyStorageRoot	= "storage_root"
	ServiceStateDependencyMessageRoot	= "message_root"
	ServiceStateDependencyReceiptsRoot	= "receipts_root"
)

type ServiceRegistryParamsV2 struct {
	MaxTTLDelta			uint64
	AllowedEndpointTypes		[]CanonicalServiceEndpointType
	AllowedStateDependencyRoots	[]string
	ProofHorizon			uint64
	ParamsHash			string
}

type ServiceVersionedInterfaceEntryV2 struct {
	InterfaceHash	string
	Version		uint64
	Key		string
	DescriptorHash	string
	EntryHash	string
}

type ServiceRegistryExportV2 struct {
	State		ServiceRegistryStateV2
	Params		ServiceRegistryParamsV2
	Roots		ServiceRegistryRootsV2
	ExportHash	string
}

func DefaultServiceRegistryParamsV2() ServiceRegistryParamsV2 {
	params, _ := NewServiceRegistryParamsV2(ServiceRegistryParamsV2{
		MaxTTLDelta:	100000,
		AllowedEndpointTypes: []CanonicalServiceEndpointType{
			CanonicalEndpointAPI,
			CanonicalEndpointApplication,
			CanonicalEndpointHybrid,
			CanonicalEndpointOffChain,
			CanonicalEndpointZoneAware,
		},
		AllowedStateDependencyRoots: []string{
			ServiceStateDependencyCommitted,
			ServiceStateDependencyGlobalRoot,
			ServiceStateDependencyIdentityRoot,
			ServiceStateDependencyMessageRoot,
			ServiceStateDependencyModuleRoot,
			ServiceStateDependencyReceiptsRoot,
			ServiceStateDependencyServiceRoot,
			ServiceStateDependencyStorageRoot,
			ServiceStateDependencyZoneRoot,
		},
		ProofHorizon:	10000,
	})
	return params
}

func NewServiceRegistryParamsV2(params ServiceRegistryParamsV2) (ServiceRegistryParamsV2, error) {
	if params.ParamsHash != "" {
		return ServiceRegistryParamsV2{}, errors.New("service registry params hash must be empty before construction")
	}
	params = CanonicalizeServiceRegistryParamsV2(params)
	if err := params.ValidateFormat(); err != nil {
		return ServiceRegistryParamsV2{}, err
	}
	params.ParamsHash = ComputeServiceRegistryParamsV2Hash(params)
	return params, params.Validate()
}

func CanonicalizeServiceRegistryParamsV2(params ServiceRegistryParamsV2) ServiceRegistryParamsV2 {
	params.AllowedEndpointTypes = normalizeEndpointTypesV2(params.AllowedEndpointTypes)
	params.AllowedStateDependencyRoots = normalizeDescriptorTokens(params.AllowedStateDependencyRoots)
	return params
}

func (params ServiceRegistryParamsV2) ValidateFormat() error {
	if params.MaxTTLDelta == 0 {
		return errors.New("service registry params max ttl delta must be positive")
	}
	if len(params.AllowedEndpointTypes) == 0 {
		return errors.New("service registry params require allowed endpoint types")
	}
	previousEndpoint := ""
	for _, endpointType := range params.AllowedEndpointTypes {
		if !IsCanonicalEndpointType(endpointType) {
			return fmt.Errorf("service registry params endpoint type %q is not supported", endpointType)
		}
		if previousEndpoint != "" && previousEndpoint >= string(endpointType) {
			return errors.New("service registry params endpoint types must be sorted canonically")
		}
		previousEndpoint = string(endpointType)
	}
	if len(params.AllowedStateDependencyRoots) == 0 {
		return errors.New("service registry params require state dependency roots")
	}
	previousRoot := ""
	for _, rootType := range params.AllowedStateDependencyRoots {
		if err := validateServiceStateDependencyRootTypeV2(rootType); err != nil {
			return err
		}
		if previousRoot != "" && previousRoot >= rootType {
			return errors.New("service registry params state dependency roots must be sorted canonically")
		}
		previousRoot = rootType
	}
	if params.ProofHorizon == 0 {
		return errors.New("service registry params proof horizon must be positive")
	}
	if params.ParamsHash != "" {
		return coretypes.ValidateHash("service registry params hash", params.ParamsHash)
	}
	return nil
}

func (params ServiceRegistryParamsV2) Validate() error {
	params = CanonicalizeServiceRegistryParamsV2(params)
	if err := params.ValidateFormat(); err != nil {
		return err
	}
	if params.ParamsHash == "" {
		return errors.New("service registry params hash is required")
	}
	if params.ParamsHash != ComputeServiceRegistryParamsV2Hash(params) {
		return errors.New("service registry params hash mismatch")
	}
	return nil
}

func NewVersionedServiceInterfaceDescriptorV2(descriptor DistributedInterfaceDescriptor) (DistributedInterfaceDescriptor, error) {
	if descriptor.InterfaceHash != "" || descriptor.DescriptorHash != "" {
		return DistributedInterfaceDescriptor{}, errors.New("versioned service interface hashes must be empty before construction")
	}
	if err := validateInterfaceToken("versioned service interface name", descriptor.InterfaceName); err != nil {
		return DistributedInterfaceDescriptor{}, err
	}
	if descriptor.Version == 0 {
		return DistributedInterfaceDescriptor{}, errors.New("versioned service interface version must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "versioned service interface schema hash", value: descriptor.SchemaHash},
		{name: "versioned service interface method root", value: descriptor.MethodRoot},
		{name: "versioned service interface event root", value: descriptor.EventRoot},
		{name: "versioned service interface error root", value: descriptor.ErrorRoot},
	} {
		if err := coretypes.ValidateHash(item.name, item.value); err != nil {
			return DistributedInterfaceDescriptor{}, err
		}
	}
	descriptor.InterfaceHash = ComputeServiceInterfaceDescriptorBytesHashV2(descriptor)
	descriptor.DescriptorHash = ComputeDistributedInterfaceDescriptorHash(descriptor)
	return descriptor, descriptor.Validate()
}

func ServiceVersionedInterfaceV2Key(interfaceHash string, version uint64) (string, error) {
	if err := coretypes.ValidateHash("service registry versioned interface hash", interfaceHash); err != nil {
		return "", err
	}
	if version == 0 {
		return "", errors.New("service registry versioned interface version must be positive")
	}
	return ServiceInterfaceRegistryPrefix + "/" + interfaceHash + "/versions/" + fmt.Sprintf("%020d", version), nil
}

func BuildServiceVersionedInterfaceEntriesV2(interfaces []DistributedInterfaceDescriptor) ([]ServiceVersionedInterfaceEntryV2, error) {
	ordered := normalizeDistributedInterfaces(interfaces)
	entries := make([]ServiceVersionedInterfaceEntryV2, 0, len(ordered))
	for _, iface := range ordered {
		if err := ValidateVersionedServiceInterfaceDescriptorV2(iface); err != nil {
			return nil, err
		}
		key, err := ServiceVersionedInterfaceV2Key(iface.InterfaceHash, iface.Version)
		if err != nil {
			return nil, err
		}
		entry := ServiceVersionedInterfaceEntryV2{
			InterfaceHash:	iface.InterfaceHash,
			Version:	iface.Version,
			Key:		key,
			DescriptorHash:	iface.DescriptorHash,
		}
		entry.EntryHash = ComputeServiceVersionedInterfaceEntryV2Hash(entry)
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	return entries, validateServiceVersionedInterfaceEntriesV2(entries)
}

func ValidateVersionedServiceInterfaceDescriptorV2(descriptor DistributedInterfaceDescriptor) error {
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if descriptor.InterfaceHash != ComputeServiceInterfaceDescriptorBytesHashV2(descriptor) {
		return errors.New("service registry interface hash must commit to descriptor bytes")
	}
	return nil
}

func ValidateServiceDescriptorAgainstParamsV2(descriptor CanonicalServiceDescriptor, interfaces []DistributedInterfaceDescriptor, params ServiceRegistryParamsV2, currentHeight uint64) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if currentHeight == 0 {
		return errors.New("service registry validation height must be positive")
	}
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if descriptor.TTLHeight > currentHeight+params.MaxTTLDelta {
		return errors.New("service registry descriptor ttl exceeds configured maximum")
	}
	if !serviceEndpointTypeAllowedV2(descriptor.EndpointType, params.AllowedEndpointTypes) {
		return fmt.Errorf("service registry endpoint type %q is not allowed by params", descriptor.EndpointType)
	}
	if err := ValidateServiceStateDependencyV2(descriptor.StateDependency, params); err != nil {
		return err
	}
	idx, found := findDistributedInterfaceV2(interfaces, descriptor.InterfaceHash)
	if !found {
		return fmt.Errorf("service registry descriptor %s missing interface %s", descriptor.ServiceID, descriptor.InterfaceHash)
	}
	return ValidateVersionedServiceInterfaceDescriptorV2(interfaces[idx])
}

func ValidateServiceRegistryStateWithParamsV2(state ServiceRegistryStateV2, params ServiceRegistryParamsV2, currentHeight uint64) error {
	if err := state.Validate(); err != nil {
		return err
	}
	versionedEntries, err := BuildServiceVersionedInterfaceEntriesV2(state.Interfaces)
	if err != nil {
		return err
	}
	if len(versionedEntries) != len(state.Interfaces) {
		return errors.New("service registry versioned interface index mismatch")
	}
	for _, descriptor := range state.Descriptors {
		if err := ValidateServiceDescriptorAgainstParamsV2(descriptor, state.Interfaces, params, currentHeight); err != nil {
			return err
		}
	}
	return nil
}

func RegisterServiceInRegistryWithParamsV2(state ServiceRegistryStateV2, msg MsgRegisterServiceV2, params ServiceRegistryParamsV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := ValidateServiceDescriptorAgainstParamsV2(msg.Descriptor, state.Interfaces, params, height); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	return RegisterServiceInRegistryV2(state, msg, height)
}

func UpdateServiceInRegistryWithParamsV2(state ServiceRegistryStateV2, msg MsgUpdateServiceV2, params ServiceRegistryParamsV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := ValidateServiceOwnerAuthorizationV2(state, msg.Descriptor.ServiceID, msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	if err := ValidateServiceDescriptorAgainstParamsV2(msg.Descriptor, state.Interfaces, params, height); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	return UpdateServiceInRegistryV2(state, msg, height)
}

func ValidateServiceOwnerAuthorizationV2(state ServiceRegistryStateV2, serviceID, authority string) error {
	if err := validateInterfaceToken("service registry authorization authority", authority); err != nil {
		return err
	}
	idx, found := findServiceDescriptorV2(state.Descriptors, serviceID)
	if !found {
		return fmt.Errorf("service registry descriptor %s not found", serviceID)
	}
	if state.Descriptors[idx].Owner != authority {
		return errors.New("service registry owner must authorize changes")
	}
	return nil
}

func ValidateServiceStateDependencyV2(stateDependency string, params ServiceRegistryParamsV2) error {
	if err := validateInterfaceToken("service registry state dependency", stateDependency); err != nil {
		return err
	}
	rootType, target := splitServiceStateDependencyV2(stateDependency)
	if !serviceStateDependencyAllowedV2(rootType, params.AllowedStateDependencyRoots) {
		return fmt.Errorf("service registry state dependency root %q is not allowed by params", rootType)
	}
	switch rootType {
	case ServiceStateDependencyZoneRoot, ServiceStateDependencyModuleRoot:
		if target == "" {
			return fmt.Errorf("service registry %s dependency requires target", rootType)
		}
		return validateInterfaceToken("service registry state dependency target", target)
	default:
		if target != "" {
			return fmt.Errorf("service registry %s dependency must not include target", rootType)
		}
		return nil
	}
}

func LookupServiceByRegistryKeyV2(state ServiceRegistryStateV2, key string) (CanonicalServiceDescriptor, bool, error) {
	if err := state.Validate(); err != nil {
		return CanonicalServiceDescriptor{}, false, err
	}
	if !IsServiceStoreKey(key) {
		return CanonicalServiceDescriptor{}, false, errors.New("service registry lookup key must be a services store key")
	}
	for _, descriptor := range state.Descriptors {
		descriptorKey, _ := ServiceDescriptorV2Key(descriptor.ServiceID)
		if key == descriptorKey {
			return descriptor, true, nil
		}
	}
	indexes := append([]ServiceRegistryIndexEntryV2(nil), state.OwnerIndex...)
	indexes = append(indexes, state.ZoneIndex...)
	indexes = append(indexes, state.MethodIndex...)
	sort.SliceStable(indexes, func(i, j int) bool { return indexes[i].Key < indexes[j].Key })
	for _, entry := range indexes {
		if entry.Key != key {
			continue
		}
		idx, found := findServiceDescriptorV2(state.Descriptors, entry.Value)
		if !found {
			return CanonicalServiceDescriptor{}, false, fmt.Errorf("service registry index points to missing descriptor %s", entry.Value)
		}
		return state.Descriptors[idx], true, nil
	}
	return CanonicalServiceDescriptor{}, false, nil
}

func ExportServiceRegistryV2(state ServiceRegistryStateV2, params ServiceRegistryParamsV2) (ServiceRegistryExportV2, error) {
	if err := ValidateServiceRegistryStateWithParamsV2(state, params, state.Height); err != nil {
		return ServiceRegistryExportV2{}, err
	}
	export := ServiceRegistryExportV2{
		State:	state,
		Params:	params,
		Roots:	ComputeServiceRegistryRootsV2(state),
	}
	export.ExportHash = ComputeServiceRegistryExportV2Hash(export)
	return export, export.Validate()
}

func ImportServiceRegistryV2(export ServiceRegistryExportV2) (ServiceRegistryStateV2, error) {
	if err := export.Validate(); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	return export.State, nil
}

func (export ServiceRegistryExportV2) Validate() error {
	if err := ValidateServiceRegistryStateWithParamsV2(export.State, export.Params, export.State.Height); err != nil {
		return err
	}
	if export.Roots != ComputeServiceRegistryRootsV2(export.State) {
		return errors.New("service registry export roots mismatch")
	}
	if err := coretypes.ValidateHash("service registry export hash", export.ExportHash); err != nil {
		return err
	}
	if export.ExportHash != ComputeServiceRegistryExportV2Hash(export) {
		return errors.New("service registry export hash mismatch")
	}
	return nil
}

func ComputeServiceRegistryParamsV2Hash(params ServiceRegistryParamsV2) string {
	params = CanonicalizeServiceRegistryParamsV2(params)
	parts := []string{
		"aetra-services-registry-params-v1",
		fmt.Sprint(params.MaxTTLDelta),
		fmt.Sprint(params.ProofHorizon),
		"endpoint_types",
		fmt.Sprint(len(params.AllowedEndpointTypes)),
	}
	for _, endpointType := range params.AllowedEndpointTypes {
		parts = append(parts, string(endpointType))
	}
	parts = append(parts, "state_dependencies", fmt.Sprint(len(params.AllowedStateDependencyRoots)))
	parts = append(parts, params.AllowedStateDependencyRoots...)
	return servicesHashParts(parts...)
}

func ComputeServiceInterfaceDescriptorBytesHashV2(descriptor DistributedInterfaceDescriptor) string {
	return servicesHashParts(
		"aetra-services-interface-descriptor-bytes-v1",
		descriptor.InterfaceName,
		fmt.Sprint(descriptor.Version),
		descriptor.SchemaHash,
		descriptor.MethodRoot,
		descriptor.EventRoot,
		descriptor.ErrorRoot,
	)
}

func ComputeServiceVersionedInterfaceEntryV2Hash(entry ServiceVersionedInterfaceEntryV2) string {
	return servicesHashParts("aetra-services-versioned-interface-entry-v1", entry.InterfaceHash, fmt.Sprint(entry.Version), entry.Key, entry.DescriptorHash)
}

func ComputeServiceVersionedInterfaceRootV2(interfaces []DistributedInterfaceDescriptor) (string, error) {
	entries, err := BuildServiceVersionedInterfaceEntriesV2(interfaces)
	if err != nil {
		return "", err
	}
	parts := []string{"aetra-services-versioned-interface-root-v1", fmt.Sprint(len(entries))}
	for _, entry := range entries {
		parts = append(parts, entry.EntryHash)
	}
	return servicesHashParts(parts...), nil
}

func ComputeServiceRegistryExportV2Hash(export ServiceRegistryExportV2) string {
	return servicesHashParts(
		"aetra-services-registry-export-v1",
		export.State.StateRoot,
		export.Params.ParamsHash,
		export.Roots.StateRoot,
		fmt.Sprint(export.State.Height),
	)
}

func normalizeEndpointTypesV2(endpointTypes []CanonicalServiceEndpointType) []CanonicalServiceEndpointType {
	raw := make([]string, 0, len(endpointTypes))
	for _, endpointType := range endpointTypes {
		raw = append(raw, string(endpointType))
	}
	sort.Strings(raw)
	unique := make([]CanonicalServiceEndpointType, 0, len(raw))
	for _, endpointType := range raw {
		if len(unique) == 0 || string(unique[len(unique)-1]) != endpointType {
			unique = append(unique, CanonicalServiceEndpointType(endpointType))
		}
	}
	return unique
}

func validateServiceStateDependencyRootTypeV2(rootType string) error {
	if err := validateInterfaceToken("service registry state dependency root type", rootType); err != nil {
		return err
	}
	switch rootType {
	case ServiceStateDependencyCommitted, ServiceStateDependencyZoneRoot, ServiceStateDependencyModuleRoot,
		ServiceStateDependencyGlobalRoot, ServiceStateDependencyServiceRoot, ServiceStateDependencyIdentityRoot,
		ServiceStateDependencyStorageRoot, ServiceStateDependencyMessageRoot, ServiceStateDependencyReceiptsRoot:
		return nil
	default:
		return fmt.Errorf("service registry state dependency root type %q is not supported", rootType)
	}
}

func splitServiceStateDependencyV2(stateDependency string) (string, string) {
	rootType, target, found := strings.Cut(stateDependency, ":")
	if !found {
		return stateDependency, ""
	}
	return rootType, target
}

func serviceEndpointTypeAllowedV2(endpointType CanonicalServiceEndpointType, allowed []CanonicalServiceEndpointType) bool {
	for _, item := range allowed {
		if item == endpointType {
			return true
		}
	}
	return false
}

func serviceStateDependencyAllowedV2(rootType string, allowed []string) bool {
	for _, item := range allowed {
		if item == rootType {
			return true
		}
	}
	return false
}

func validateServiceVersionedInterfaceEntriesV2(entries []ServiceVersionedInterfaceEntryV2) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, entry := range entries {
		if err := coretypes.ValidateHash("service registry versioned interface hash", entry.InterfaceHash); err != nil {
			return err
		}
		if entry.Version == 0 {
			return errors.New("service registry versioned interface version must be positive")
		}
		expectedKey, err := ServiceVersionedInterfaceV2Key(entry.InterfaceHash, entry.Version)
		if err != nil {
			return err
		}
		if entry.Key != expectedKey {
			return errors.New("service registry versioned interface key mismatch")
		}
		if err := coretypes.ValidateHash("service registry versioned interface descriptor hash", entry.DescriptorHash); err != nil {
			return err
		}
		if entry.EntryHash != ComputeServiceVersionedInterfaceEntryV2Hash(entry) {
			return errors.New("service registry versioned interface entry hash mismatch")
		}
		if _, found := seen[entry.Key]; found {
			return fmt.Errorf("duplicate service registry versioned interface key %s", entry.Key)
		}
		seen[entry.Key] = struct{}{}
		if previous != "" && previous >= entry.Key {
			return errors.New("service registry versioned interface entries must be sorted canonically")
		}
		previous = entry.Key
	}
	return nil
}
