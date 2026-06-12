package types

import (
	"errors"
	"fmt"
	"sort"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

const (
	ServiceDescriptorRegistryPrefix	= ServiceStorePrefix + "descriptors"
	ServiceInterfaceRegistryPrefix	= ServiceStorePrefix + "interfaces"
	ServiceOwnerIndexPrefix		= ServiceStorePrefix + "owner_index"
	ServiceZoneIndexPrefix		= ServiceStorePrefix + "zone_index"
	ServiceMethodIndexPrefix	= ServiceStorePrefix + "method_index"
	ServiceReceiptRegistryPrefix	= ServiceStorePrefix + "receipts"
)

type ServiceRegistryIndexEntryV2 struct {
	Key		string
	Value		string
	EntryHash	string
}

type ServiceReceiptV2 struct {
	ServiceID	string
	ReceiptID	string
	CallID		string
	Status		string
	ResultHash	string
	Height		uint64
	ReceiptHash	string
}

type ServiceIdentityBindingV2 struct {
	ServiceID	string
	IdentityName	string
	Owner		string
	BoundHeight	uint64
	BindingHash	string
}

type ServiceRegistryStateV2 struct {
	Descriptors		[]CanonicalServiceDescriptor
	Interfaces		[]DistributedInterfaceDescriptor
	OwnerIndex		[]ServiceRegistryIndexEntryV2
	ZoneIndex		[]ServiceRegistryIndexEntryV2
	MethodIndex		[]ServiceRegistryIndexEntryV2
	Receipts		[]ServiceReceiptV2
	IdentityBindings	[]ServiceIdentityBindingV2
	Height			uint64
	StateRoot		string
}

type ServiceRegistryRootsV2 struct {
	DescriptorRoot	string
	InterfaceRoot	string
	OwnerRoot	string
	ZoneRoot	string
	MethodRoot	string
	ReceiptRoot	string
	BindingRoot	string
	StateRoot	string
}

type ServiceRegistryProofV2 struct {
	Query		string
	Key		string
	ValueHash	string
	Root		string
	Height		uint64
	ProofHashes	[]string
}

type MsgRegisterServiceV2 struct {
	Authority	string
	Descriptor	CanonicalServiceDescriptor
}

type MsgUpdateServiceV2 struct {
	Authority	string
	Descriptor	CanonicalServiceDescriptor
}

type MsgDisableServiceV2 struct {
	Authority	string
	ServiceID	string
	Height		uint64
}

type MsgRegisterInterfaceV2 struct {
	Authority	string
	Descriptor	DistributedInterfaceDescriptor
}

type MsgUpdateInterfaceV2 struct {
	Authority	string
	Descriptor	DistributedInterfaceDescriptor
}

type MsgBindServiceToIdentityV2 struct {
	Authority	string
	ServiceID	string
	IdentityName	string
	Height		uint64
}

type MsgUnbindServiceFromIdentityV2 struct {
	Authority	string
	ServiceID	string
	IdentityName	string
}

type QueryServiceV2 struct {
	ServiceID string
}

type QueryServicesByOwnerV2 struct {
	Owner string
}

type QueryServicesByZoneV2 struct {
	ZoneID string
}

type QueryInterfaceV2 struct {
	InterfaceHash string
}

type QueryServiceByMethodV2 struct {
	MethodHash string
}

type QueryServiceRootV2 struct{}

type QueryServiceProofV2 struct {
	Key string
}

func ServiceDescriptorV2Key(serviceID string) (string, error) {
	if err := validateInterfaceToken("service registry service id", serviceID); err != nil {
		return "", err
	}
	return ServiceDescriptorRegistryPrefix + "/" + serviceID, nil
}

func ServiceInterfaceV2Key(interfaceHash string) (string, error) {
	if err := coretypes.ValidateHash("service registry interface hash", interfaceHash); err != nil {
		return "", err
	}
	return ServiceInterfaceRegistryPrefix + "/" + interfaceHash, nil
}

func ServiceOwnerIndexV2Key(owner, serviceID string) (string, error) {
	if err := validateInterfaceToken("service registry owner", owner); err != nil {
		return "", err
	}
	if err := validateInterfaceToken("service registry owner service id", serviceID); err != nil {
		return "", err
	}
	return ServiceOwnerIndexPrefix + "/" + owner + "/" + serviceID, nil
}

func ServiceZoneIndexV2Key(zoneID, serviceID string) (string, error) {
	if err := validateInterfaceToken("service registry zone id", zoneID); err != nil {
		return "", err
	}
	if err := validateInterfaceToken("service registry zone service id", serviceID); err != nil {
		return "", err
	}
	return ServiceZoneIndexPrefix + "/" + zoneID + "/" + serviceID, nil
}

func ServiceMethodIndexV2Key(methodHash, serviceID string) (string, error) {
	if err := coretypes.ValidateHash("service registry method hash", methodHash); err != nil {
		return "", err
	}
	if err := validateInterfaceToken("service registry method service id", serviceID); err != nil {
		return "", err
	}
	return ServiceMethodIndexPrefix + "/" + methodHash + "/" + serviceID, nil
}

func ServiceReceiptV2Key(serviceID, receiptID string) (string, error) {
	if err := validateInterfaceToken("service registry receipt service id", serviceID); err != nil {
		return "", err
	}
	if err := validateInterfaceToken("service registry receipt id", receiptID); err != nil {
		return "", err
	}
	return ServiceReceiptRegistryPrefix + "/" + serviceID + "/" + receiptID, nil
}

func NewServiceReceiptV2(receipt ServiceReceiptV2) (ServiceReceiptV2, error) {
	if receipt.ReceiptHash != "" {
		return ServiceReceiptV2{}, errors.New("service registry receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return ServiceReceiptV2{}, err
	}
	receipt.ReceiptHash = ComputeServiceReceiptV2Hash(receipt)
	return receipt, receipt.Validate()
}

func NewServiceIdentityBindingV2(binding ServiceIdentityBindingV2) (ServiceIdentityBindingV2, error) {
	if binding.BindingHash != "" {
		return ServiceIdentityBindingV2{}, errors.New("service registry binding hash must be empty before construction")
	}
	if err := binding.ValidateFormat(); err != nil {
		return ServiceIdentityBindingV2{}, err
	}
	binding.BindingHash = ComputeServiceIdentityBindingV2Hash(binding)
	return binding, binding.Validate()
}

func BuildServiceRegistryStateV2(descriptors []CanonicalServiceDescriptor, interfaces []DistributedInterfaceDescriptor, receipts []ServiceReceiptV2, bindings []ServiceIdentityBindingV2, height uint64) (ServiceRegistryStateV2, error) {
	state := ServiceRegistryStateV2{
		Descriptors:		normalizeServiceRegistryDescriptorsV2(descriptors),
		Interfaces:		normalizeDistributedInterfaces(interfaces),
		Receipts:		normalizeServiceReceiptsV2(receipts),
		IdentityBindings:	normalizeServiceIdentityBindingsV2(bindings),
		Height:			height,
	}
	if err := state.rebuildIndexes(); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	if err := state.ValidateFormat(); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	state.StateRoot = ComputeServiceRegistryStateRootV2(state)
	return state, state.Validate()
}

func RegisterServiceInRegistryV2(state ServiceRegistryStateV2, msg MsgRegisterServiceV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := validateInterfaceToken("service registry register authority", msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	if msg.Authority != msg.Descriptor.Owner {
		return ServiceRegistryStateV2{}, errors.New("service registry register authority must own service")
	}
	if _, found := findServiceDescriptorV2(state.Descriptors, msg.Descriptor.ServiceID); found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry descriptor %s already exists", msg.Descriptor.ServiceID)
	}
	state.Descriptors = append(state.Descriptors, msg.Descriptor)
	return BuildServiceRegistryStateV2(state.Descriptors, state.Interfaces, state.Receipts, state.IdentityBindings, height)
}

func UpdateServiceInRegistryV2(state ServiceRegistryStateV2, msg MsgUpdateServiceV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := validateInterfaceToken("service registry update authority", msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	idx, found := findServiceDescriptorV2(state.Descriptors, msg.Descriptor.ServiceID)
	if !found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry descriptor %s not found", msg.Descriptor.ServiceID)
	}
	if state.Descriptors[idx].Owner != msg.Authority || msg.Descriptor.Owner != msg.Authority {
		return ServiceRegistryStateV2{}, errors.New("service registry update authority must own service")
	}
	state.Descriptors[idx] = msg.Descriptor
	return BuildServiceRegistryStateV2(state.Descriptors, state.Interfaces, state.Receipts, state.IdentityBindings, height)
}

func DisableServiceInRegistryV2(state ServiceRegistryStateV2, msg MsgDisableServiceV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := validateInterfaceToken("service registry disable authority", msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	idx, found := findServiceDescriptorV2(state.Descriptors, msg.ServiceID)
	if !found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry descriptor %s not found", msg.ServiceID)
	}
	descriptor := state.Descriptors[idx]
	if descriptor.Owner != msg.Authority {
		return ServiceRegistryStateV2{}, errors.New("service registry disable authority must own service")
	}
	descriptor.Status = CanonicalServiceStatusDisabled
	if msg.Height > 0 {
		descriptor.TTLHeight = msg.Height
	}
	descriptor.DescriptorHash = ""
	descriptor, err := NewCanonicalServiceDescriptor(descriptor)
	if err != nil {
		return ServiceRegistryStateV2{}, err
	}
	state.Descriptors[idx] = descriptor
	return BuildServiceRegistryStateV2(state.Descriptors, state.Interfaces, state.Receipts, state.IdentityBindings, height)
}

func RegisterInterfaceInRegistryV2(state ServiceRegistryStateV2, msg MsgRegisterInterfaceV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := validateInterfaceToken("service registry register interface authority", msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	if _, found := findDistributedInterfaceV2(state.Interfaces, msg.Descriptor.InterfaceHash); found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry interface %s already exists", msg.Descriptor.InterfaceHash)
	}
	state.Interfaces = append(state.Interfaces, msg.Descriptor)
	return BuildServiceRegistryStateV2(state.Descriptors, state.Interfaces, state.Receipts, state.IdentityBindings, height)
}

func UpdateInterfaceInRegistryV2(state ServiceRegistryStateV2, msg MsgUpdateInterfaceV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := validateInterfaceToken("service registry update interface authority", msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	idx, found := findDistributedInterfaceV2(state.Interfaces, msg.Descriptor.InterfaceHash)
	if !found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry interface %s not found", msg.Descriptor.InterfaceHash)
	}
	state.Interfaces[idx] = msg.Descriptor
	return BuildServiceRegistryStateV2(state.Descriptors, state.Interfaces, state.Receipts, state.IdentityBindings, height)
}

func BindServiceToIdentityInRegistryV2(state ServiceRegistryStateV2, msg MsgBindServiceToIdentityV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := validateInterfaceToken("service registry bind authority", msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	idx, found := findServiceDescriptorV2(state.Descriptors, msg.ServiceID)
	if !found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry descriptor %s not found", msg.ServiceID)
	}
	if state.Descriptors[idx].Owner != msg.Authority {
		return ServiceRegistryStateV2{}, errors.New("service registry bind authority must own service")
	}
	binding, err := NewServiceIdentityBindingV2(ServiceIdentityBindingV2{
		ServiceID:	msg.ServiceID,
		IdentityName:	msg.IdentityName,
		Owner:		msg.Authority,
		BoundHeight:	firstPositiveHeight(msg.Height, height),
	})
	if err != nil {
		return ServiceRegistryStateV2{}, err
	}
	if _, found := findServiceIdentityBindingV2(state.IdentityBindings, binding.ServiceID, binding.IdentityName); found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry identity binding %s/%s already exists", binding.ServiceID, binding.IdentityName)
	}
	state.IdentityBindings = append(state.IdentityBindings, binding)
	return BuildServiceRegistryStateV2(state.Descriptors, state.Interfaces, state.Receipts, state.IdentityBindings, height)
}

func UnbindServiceFromIdentityInRegistryV2(state ServiceRegistryStateV2, msg MsgUnbindServiceFromIdentityV2, height uint64) (ServiceRegistryStateV2, error) {
	if err := validateInterfaceToken("service registry unbind authority", msg.Authority); err != nil {
		return ServiceRegistryStateV2{}, err
	}
	idx, found := findServiceIdentityBindingV2(state.IdentityBindings, msg.ServiceID, msg.IdentityName)
	if !found {
		return ServiceRegistryStateV2{}, fmt.Errorf("service registry identity binding %s/%s not found", msg.ServiceID, msg.IdentityName)
	}
	if state.IdentityBindings[idx].Owner != msg.Authority {
		return ServiceRegistryStateV2{}, errors.New("service registry unbind authority must own binding")
	}
	state.IdentityBindings = append(state.IdentityBindings[:idx], state.IdentityBindings[idx+1:]...)
	return BuildServiceRegistryStateV2(state.Descriptors, state.Interfaces, state.Receipts, state.IdentityBindings, height)
}

func QueryServiceFromRegistryV2(state ServiceRegistryStateV2, query QueryServiceV2) (CanonicalServiceDescriptor, error) {
	if err := state.Validate(); err != nil {
		return CanonicalServiceDescriptor{}, err
	}
	idx, found := findServiceDescriptorV2(state.Descriptors, query.ServiceID)
	if !found {
		return CanonicalServiceDescriptor{}, fmt.Errorf("service registry descriptor %s not found", query.ServiceID)
	}
	return state.Descriptors[idx], nil
}

func QueryServicesByOwnerFromRegistryV2(state ServiceRegistryStateV2, query QueryServicesByOwnerV2) ([]CanonicalServiceDescriptor, error) {
	if err := state.Validate(); err != nil {
		return nil, err
	}
	if err := validateInterfaceToken("service registry query owner", query.Owner); err != nil {
		return nil, err
	}
	return queryServiceRegistryByIndexV2(state, ServiceOwnerIndexPrefix+"/"+query.Owner+"/")
}

func QueryServicesByZoneFromRegistryV2(state ServiceRegistryStateV2, query QueryServicesByZoneV2) ([]CanonicalServiceDescriptor, error) {
	if err := state.Validate(); err != nil {
		return nil, err
	}
	if err := validateInterfaceToken("service registry query zone id", query.ZoneID); err != nil {
		return nil, err
	}
	return queryServiceRegistryByIndexV2(state, ServiceZoneIndexPrefix+"/"+query.ZoneID+"/")
}

func QueryInterfaceFromRegistryV2(state ServiceRegistryStateV2, query QueryInterfaceV2) (DistributedInterfaceDescriptor, error) {
	if err := state.Validate(); err != nil {
		return DistributedInterfaceDescriptor{}, err
	}
	idx, found := findDistributedInterfaceV2(state.Interfaces, query.InterfaceHash)
	if !found {
		return DistributedInterfaceDescriptor{}, fmt.Errorf("service registry interface %s not found", query.InterfaceHash)
	}
	return state.Interfaces[idx], nil
}

func QueryServiceByMethodFromRegistryV2(state ServiceRegistryStateV2, query QueryServiceByMethodV2) ([]CanonicalServiceDescriptor, error) {
	if err := state.Validate(); err != nil {
		return nil, err
	}
	if err := coretypes.ValidateHash("service registry query method hash", query.MethodHash); err != nil {
		return nil, err
	}
	return queryServiceRegistryByIndexV2(state, ServiceMethodIndexPrefix+"/"+query.MethodHash+"/")
}

func QueryServiceRootFromRegistryV2(state ServiceRegistryStateV2, _ QueryServiceRootV2) (ServiceRegistryRootsV2, error) {
	if err := state.Validate(); err != nil {
		return ServiceRegistryRootsV2{}, err
	}
	return ComputeServiceRegistryRootsV2(state), nil
}

func QueryServiceProofFromRegistryV2(state ServiceRegistryStateV2, query QueryServiceProofV2) (ServiceRegistryProofV2, error) {
	if err := state.Validate(); err != nil {
		return ServiceRegistryProofV2{}, err
	}
	if !IsServiceStoreKey(query.Key) {
		return ServiceRegistryProofV2{}, errors.New("service registry proof key must be a services store key")
	}
	entryHash, proofHashes, found := serviceRegistryProofEntryV2(state, query.Key)
	if !found {
		return ServiceRegistryProofV2{}, fmt.Errorf("service registry proof key %s not found", query.Key)
	}
	return ServiceRegistryProofV2{
		Query:		"QueryServiceProof",
		Key:		query.Key,
		ValueHash:	entryHash,
		Root:		state.StateRoot,
		Height:		state.Height,
		ProofHashes:	proofHashes,
	}, nil
}

func (state ServiceRegistryStateV2) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("service registry height must be positive")
	}
	if err := validateServiceRegistryDescriptorsV2(state.Descriptors, state.Interfaces); err != nil {
		return err
	}
	if err := validateServiceRegistryInterfacesV2(state.Interfaces); err != nil {
		return err
	}
	if err := validateServiceRegistryIndexEntriesV2("owner", state.OwnerIndex); err != nil {
		return err
	}
	if err := validateServiceRegistryIndexEntriesV2("zone", state.ZoneIndex); err != nil {
		return err
	}
	if err := validateServiceRegistryIndexEntriesV2("method", state.MethodIndex); err != nil {
		return err
	}
	if err := validateServiceReceiptsV2(state.Receipts); err != nil {
		return err
	}
	if err := validateServiceIdentityBindingsV2(state.IdentityBindings, state.Descriptors); err != nil {
		return err
	}
	if err := validateServiceRegistryGeneratedIndexesV2(state); err != nil {
		return err
	}
	if state.StateRoot != "" {
		return coretypes.ValidateHash("service registry state root", state.StateRoot)
	}
	return nil
}

func (state ServiceRegistryStateV2) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.StateRoot == "" {
		return errors.New("service registry state root is required")
	}
	if state.StateRoot != ComputeServiceRegistryStateRootV2(state) {
		return errors.New("service registry state root mismatch")
	}
	return nil
}

func (state *ServiceRegistryStateV2) rebuildIndexes() error {
	owner, zone, method, err := buildServiceRegistryIndexesV2(state.Descriptors)
	if err != nil {
		return err
	}
	state.OwnerIndex = owner
	state.ZoneIndex = zone
	state.MethodIndex = method
	return nil
}

func (receipt ServiceReceiptV2) ValidateFormat() error {
	if _, err := ServiceReceiptV2Key(receipt.ServiceID, receipt.ReceiptID); err != nil {
		return err
	}
	if err := validateInterfaceToken("service registry receipt call id", receipt.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("service registry receipt status", receipt.Status); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("service registry receipt result hash", receipt.ResultHash); err != nil {
		return err
	}
	if receipt.Height == 0 {
		return errors.New("service registry receipt height must be positive")
	}
	if receipt.ReceiptHash != "" {
		return coretypes.ValidateHash("service registry receipt hash", receipt.ReceiptHash)
	}
	return nil
}

func (receipt ServiceReceiptV2) Validate() error {
	if err := receipt.ValidateFormat(); err != nil {
		return err
	}
	if receipt.ReceiptHash == "" {
		return errors.New("service registry receipt hash is required")
	}
	if receipt.ReceiptHash != ComputeServiceReceiptV2Hash(receipt) {
		return errors.New("service registry receipt hash mismatch")
	}
	return nil
}

func (binding ServiceIdentityBindingV2) ValidateFormat() error {
	if _, err := ServiceDescriptorV2Key(binding.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("service registry identity name", binding.IdentityName); err != nil {
		return err
	}
	if err := validateInterfaceToken("service registry identity owner", binding.Owner); err != nil {
		return err
	}
	if binding.BoundHeight == 0 {
		return errors.New("service registry identity binding height must be positive")
	}
	if binding.BindingHash != "" {
		return coretypes.ValidateHash("service registry identity binding hash", binding.BindingHash)
	}
	return nil
}

func (binding ServiceIdentityBindingV2) Validate() error {
	if err := binding.ValidateFormat(); err != nil {
		return err
	}
	if binding.BindingHash == "" {
		return errors.New("service registry identity binding hash is required")
	}
	if binding.BindingHash != ComputeServiceIdentityBindingV2Hash(binding) {
		return errors.New("service registry identity binding hash mismatch")
	}
	return nil
}

func ComputeServiceRegistryMethodHashV2(interfaceHash, method string) string {
	return servicesHashParts("aetra-services-method-index-v1", interfaceHash, method)
}

func ComputeServiceReceiptV2Hash(receipt ServiceReceiptV2) string {
	return servicesHashParts("aetra-services-registry-receipt-v1", receipt.ServiceID, receipt.ReceiptID, receipt.CallID, receipt.Status, receipt.ResultHash, fmt.Sprint(receipt.Height))
}

func ComputeServiceIdentityBindingV2Hash(binding ServiceIdentityBindingV2) string {
	return servicesHashParts("aetra-services-identity-binding-v1", binding.ServiceID, binding.IdentityName, binding.Owner, fmt.Sprint(binding.BoundHeight))
}

func ComputeServiceRegistryIndexEntryV2Hash(entry ServiceRegistryIndexEntryV2) string {
	return servicesHashParts("aetra-services-registry-index-entry-v1", entry.Key, entry.Value)
}

func ComputeServiceRegistryDescriptorRootV2(descriptors []CanonicalServiceDescriptor) string {
	ordered := normalizeServiceRegistryDescriptorsV2(descriptors)
	parts := []string{"aetra-services-registry-descriptor-root-v1", fmt.Sprint(len(ordered))}
	for _, descriptor := range ordered {
		parts = append(parts, descriptor.DescriptorHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRegistryInterfaceRootV2(interfaces []DistributedInterfaceDescriptor) string {
	return ComputeDistributedInterfaceRoot(interfaces)
}

func ComputeServiceRegistryIndexRootV2(kind string, entries []ServiceRegistryIndexEntryV2) string {
	ordered := normalizeServiceRegistryIndexEntriesV2(entries)
	parts := []string{"aetra-services-registry-index-root-v1", kind, fmt.Sprint(len(ordered))}
	for _, entry := range ordered {
		parts = append(parts, entry.EntryHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceReceiptRootV2(receipts []ServiceReceiptV2) string {
	ordered := normalizeServiceReceiptsV2(receipts)
	parts := []string{"aetra-services-registry-receipt-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ReceiptHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceIdentityBindingRootV2(bindings []ServiceIdentityBindingV2) string {
	ordered := normalizeServiceIdentityBindingsV2(bindings)
	parts := []string{"aetra-services-identity-binding-root-v1", fmt.Sprint(len(ordered))}
	for _, binding := range ordered {
		parts = append(parts, binding.BindingHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRegistryRootsV2(state ServiceRegistryStateV2) ServiceRegistryRootsV2 {
	roots := ServiceRegistryRootsV2{
		DescriptorRoot:	ComputeServiceRegistryDescriptorRootV2(state.Descriptors),
		InterfaceRoot:	ComputeServiceRegistryInterfaceRootV2(state.Interfaces),
		OwnerRoot:	ComputeServiceRegistryIndexRootV2("owner", state.OwnerIndex),
		ZoneRoot:	ComputeServiceRegistryIndexRootV2("zone", state.ZoneIndex),
		MethodRoot:	ComputeServiceRegistryIndexRootV2("method", state.MethodIndex),
		ReceiptRoot:	ComputeServiceReceiptRootV2(state.Receipts),
		BindingRoot:	ComputeServiceIdentityBindingRootV2(state.IdentityBindings),
	}
	roots.StateRoot = servicesHashParts(
		"aetra-services-registry-state-root-v1",
		roots.DescriptorRoot,
		roots.InterfaceRoot,
		roots.OwnerRoot,
		roots.ZoneRoot,
		roots.MethodRoot,
		roots.ReceiptRoot,
		roots.BindingRoot,
		fmt.Sprint(state.Height),
	)
	return roots
}

func ComputeServiceRegistryStateRootV2(state ServiceRegistryStateV2) string {
	return ComputeServiceRegistryRootsV2(state).StateRoot
}

func normalizeServiceRegistryDescriptorsV2(descriptors []CanonicalServiceDescriptor) []CanonicalServiceDescriptor {
	out := append([]CanonicalServiceDescriptor(nil), descriptors...)
	for i := range out {
		out[i] = CanonicalizeServiceDescriptor(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ServiceID < out[j].ServiceID })
	return out
}

func normalizeServiceRegistryIndexEntriesV2(entries []ServiceRegistryIndexEntryV2) []ServiceRegistryIndexEntryV2 {
	out := append([]ServiceRegistryIndexEntryV2(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func normalizeServiceReceiptsV2(receipts []ServiceReceiptV2) []ServiceReceiptV2 {
	out := append([]ServiceReceiptV2(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ServiceID != out[j].ServiceID {
			return out[i].ServiceID < out[j].ServiceID
		}
		return out[i].ReceiptID < out[j].ReceiptID
	})
	return out
}

func normalizeServiceIdentityBindingsV2(bindings []ServiceIdentityBindingV2) []ServiceIdentityBindingV2 {
	out := append([]ServiceIdentityBindingV2(nil), bindings...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ServiceID != out[j].ServiceID {
			return out[i].ServiceID < out[j].ServiceID
		}
		return out[i].IdentityName < out[j].IdentityName
	})
	return out
}

func buildServiceRegistryIndexesV2(descriptors []CanonicalServiceDescriptor) ([]ServiceRegistryIndexEntryV2, []ServiceRegistryIndexEntryV2, []ServiceRegistryIndexEntryV2, error) {
	owner := make([]ServiceRegistryIndexEntryV2, 0, len(descriptors))
	zone := make([]ServiceRegistryIndexEntryV2, 0, len(descriptors))
	method := []ServiceRegistryIndexEntryV2{}
	for _, descriptor := range normalizeServiceRegistryDescriptorsV2(descriptors) {
		ownerKey, err := ServiceOwnerIndexV2Key(descriptor.Owner, descriptor.ServiceID)
		if err != nil {
			return nil, nil, nil, err
		}
		zoneKey, err := ServiceZoneIndexV2Key(descriptor.ZoneID, descriptor.ServiceID)
		if err != nil {
			return nil, nil, nil, err
		}
		owner = append(owner, newServiceRegistryIndexEntryV2(ownerKey, descriptor.ServiceID))
		zone = append(zone, newServiceRegistryIndexEntryV2(zoneKey, descriptor.ServiceID))
		for _, methodName := range descriptor.SupportedMethods {
			methodHash := ComputeServiceRegistryMethodHashV2(descriptor.InterfaceHash, methodName)
			methodKey, err := ServiceMethodIndexV2Key(methodHash, descriptor.ServiceID)
			if err != nil {
				return nil, nil, nil, err
			}
			method = append(method, newServiceRegistryIndexEntryV2(methodKey, descriptor.ServiceID))
		}
	}
	return normalizeServiceRegistryIndexEntriesV2(owner), normalizeServiceRegistryIndexEntriesV2(zone), normalizeServiceRegistryIndexEntriesV2(method), nil
}

func newServiceRegistryIndexEntryV2(key, value string) ServiceRegistryIndexEntryV2 {
	entry := ServiceRegistryIndexEntryV2{Key: key, Value: value}
	entry.EntryHash = ComputeServiceRegistryIndexEntryV2Hash(entry)
	return entry
}

func validateServiceRegistryDescriptorsV2(descriptors []CanonicalServiceDescriptor, interfaces []DistributedInterfaceDescriptor) error {
	interfaceIndex := map[string]struct{}{}
	for _, iface := range interfaces {
		interfaceIndex[iface.InterfaceHash] = struct{}{}
	}
	seen := map[string]struct{}{}
	previous := ""
	for _, descriptor := range descriptors {
		if err := descriptor.Validate(); err != nil {
			return err
		}
		if _, found := interfaceIndex[descriptor.InterfaceHash]; !found {
			return fmt.Errorf("service registry descriptor %s missing interface %s", descriptor.ServiceID, descriptor.InterfaceHash)
		}
		if _, found := seen[descriptor.ServiceID]; found {
			return fmt.Errorf("duplicate service registry descriptor %s", descriptor.ServiceID)
		}
		seen[descriptor.ServiceID] = struct{}{}
		if previous != "" && previous >= descriptor.ServiceID {
			return errors.New("service registry descriptors must be sorted canonically")
		}
		previous = descriptor.ServiceID
	}
	return nil
}

func validateServiceRegistryInterfacesV2(interfaces []DistributedInterfaceDescriptor) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, iface := range interfaces {
		if err := iface.Validate(); err != nil {
			return err
		}
		if _, found := seen[iface.InterfaceHash]; found {
			return fmt.Errorf("duplicate service registry interface %s", iface.InterfaceHash)
		}
		seen[iface.InterfaceHash] = struct{}{}
		if previous != "" && previous >= iface.InterfaceHash {
			return errors.New("service registry interfaces must be sorted canonically")
		}
		previous = iface.InterfaceHash
	}
	return nil
}

func validateServiceRegistryIndexEntriesV2(kind string, entries []ServiceRegistryIndexEntryV2) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, entry := range entries {
		if !IsServiceStoreKey(entry.Key) {
			return fmt.Errorf("service registry %s index key %s is outside services store", kind, entry.Key)
		}
		if err := validateInterfaceToken("service registry "+kind+" index value", entry.Value); err != nil {
			return err
		}
		if entry.EntryHash != ComputeServiceRegistryIndexEntryV2Hash(entry) {
			return fmt.Errorf("service registry %s index entry hash mismatch %s", kind, entry.Key)
		}
		if _, found := seen[entry.Key]; found {
			return fmt.Errorf("duplicate service registry %s index key %s", kind, entry.Key)
		}
		seen[entry.Key] = struct{}{}
		if previous != "" && previous >= entry.Key {
			return fmt.Errorf("service registry %s index must be sorted canonically", kind)
		}
		previous = entry.Key
	}
	return nil
}

func validateServiceReceiptsV2(receipts []ServiceReceiptV2) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		key, _ := ServiceReceiptV2Key(receipt.ServiceID, receipt.ReceiptID)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate service registry receipt %s", key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("service registry receipts must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validateServiceIdentityBindingsV2(bindings []ServiceIdentityBindingV2, descriptors []CanonicalServiceDescriptor) error {
	services := map[string]CanonicalServiceDescriptor{}
	for _, descriptor := range descriptors {
		services[descriptor.ServiceID] = descriptor
	}
	seen := map[string]struct{}{}
	previous := ""
	for _, binding := range bindings {
		if err := binding.Validate(); err != nil {
			return err
		}
		descriptor, found := services[binding.ServiceID]
		if !found {
			return fmt.Errorf("service registry identity binding %s missing service", binding.ServiceID)
		}
		if descriptor.Owner != binding.Owner {
			return fmt.Errorf("service registry identity binding %s owner mismatch", binding.ServiceID)
		}
		key := binding.ServiceID + "/" + binding.IdentityName
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate service registry identity binding %s", key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("service registry identity bindings must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validateServiceRegistryGeneratedIndexesV2(state ServiceRegistryStateV2) error {
	owner, zone, method, err := buildServiceRegistryIndexesV2(state.Descriptors)
	if err != nil {
		return err
	}
	if !equalServiceRegistryIndexEntriesV2(owner, state.OwnerIndex) {
		return errors.New("service registry owner index does not match descriptors")
	}
	if !equalServiceRegistryIndexEntriesV2(zone, state.ZoneIndex) {
		return errors.New("service registry zone index does not match descriptors")
	}
	if !equalServiceRegistryIndexEntriesV2(method, state.MethodIndex) {
		return errors.New("service registry method index does not match descriptors")
	}
	return nil
}

func equalServiceRegistryIndexEntriesV2(a, b []ServiceRegistryIndexEntryV2) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func findServiceDescriptorV2(descriptors []CanonicalServiceDescriptor, serviceID string) (int, bool) {
	for i, descriptor := range descriptors {
		if descriptor.ServiceID == serviceID {
			return i, true
		}
	}
	return -1, false
}

func findDistributedInterfaceV2(interfaces []DistributedInterfaceDescriptor, interfaceHash string) (int, bool) {
	for i, iface := range interfaces {
		if iface.InterfaceHash == interfaceHash {
			return i, true
		}
	}
	return -1, false
}

func findServiceIdentityBindingV2(bindings []ServiceIdentityBindingV2, serviceID, identityName string) (int, bool) {
	for i, binding := range bindings {
		if binding.ServiceID == serviceID && binding.IdentityName == identityName {
			return i, true
		}
	}
	return -1, false
}

func queryServiceRegistryByIndexV2(state ServiceRegistryStateV2, keyPrefix string) ([]CanonicalServiceDescriptor, error) {
	descriptors := map[string]CanonicalServiceDescriptor{}
	for _, descriptor := range state.Descriptors {
		descriptors[descriptor.ServiceID] = descriptor
	}
	indexes := append([]ServiceRegistryIndexEntryV2(nil), state.OwnerIndex...)
	indexes = append(indexes, state.ZoneIndex...)
	indexes = append(indexes, state.MethodIndex...)
	results := []CanonicalServiceDescriptor{}
	for _, entry := range indexes {
		if len(entry.Key) < len(keyPrefix) || entry.Key[:len(keyPrefix)] != keyPrefix {
			continue
		}
		descriptor, found := descriptors[entry.Value]
		if !found {
			return nil, fmt.Errorf("service registry index points to missing descriptor %s", entry.Value)
		}
		results = append(results, descriptor)
	}
	return normalizeServiceRegistryDescriptorsV2(results), nil
}

func serviceRegistryProofEntryV2(state ServiceRegistryStateV2, key string) (string, []string, bool) {
	entries := serviceRegistryProofEntriesV2(state)
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	for i, entry := range entries {
		if entry.Key != key {
			continue
		}
		proof := make([]string, 0, len(entries)-1)
		for j, other := range entries {
			if i == j {
				continue
			}
			proof = append(proof, other.ValueHash)
		}
		return entry.ValueHash, proof, true
	}
	return "", nil, false
}

func serviceRegistryProofEntriesV2(state ServiceRegistryStateV2) []struct {
	Key		string
	ValueHash	string
} {
	entries := []struct {
		Key		string
		ValueHash	string
	}{}
	for _, descriptor := range state.Descriptors {
		key, _ := ServiceDescriptorV2Key(descriptor.ServiceID)
		entries = append(entries, struct {
			Key		string
			ValueHash	string
		}{Key: key, ValueHash: descriptor.DescriptorHash})
	}
	for _, iface := range state.Interfaces {
		key, _ := ServiceInterfaceV2Key(iface.InterfaceHash)
		entries = append(entries, struct {
			Key		string
			ValueHash	string
		}{Key: key, ValueHash: iface.DescriptorHash})
	}
	for _, entry := range state.OwnerIndex {
		entries = append(entries, struct {
			Key		string
			ValueHash	string
		}{Key: entry.Key, ValueHash: entry.EntryHash})
	}
	for _, entry := range state.ZoneIndex {
		entries = append(entries, struct {
			Key		string
			ValueHash	string
		}{Key: entry.Key, ValueHash: entry.EntryHash})
	}
	for _, entry := range state.MethodIndex {
		entries = append(entries, struct {
			Key		string
			ValueHash	string
		}{Key: entry.Key, ValueHash: entry.EntryHash})
	}
	for _, receipt := range state.Receipts {
		key, _ := ServiceReceiptV2Key(receipt.ServiceID, receipt.ReceiptID)
		entries = append(entries, struct {
			Key		string
			ValueHash	string
		}{Key: key, ValueHash: receipt.ReceiptHash})
	}
	for _, binding := range state.IdentityBindings {
		key := ServiceStorePrefix + "identity_bindings/" + binding.ServiceID + "/" + binding.IdentityName
		entries = append(entries, struct {
			Key		string
			ValueHash	string
		}{Key: key, ValueHash: binding.BindingHash})
	}
	return entries
}

func firstPositiveHeight(primary, fallback uint64) uint64 {
	if primary > 0 {
		return primary
	}
	return fallback
}
