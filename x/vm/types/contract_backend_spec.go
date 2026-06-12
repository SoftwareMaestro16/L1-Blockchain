package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMContractBackendNativeModule	AVMContractBackendKind	= "native_module"
	AVMContractBackendWASMContract	AVMContractBackendKind	= "wasm_contract"
	AVMContractBackendActorContract	AVMContractBackendKind	= "avm_actor_contract"

	MaxAVMBackendNameLength		= 128
	MaxAVMBackendServiceLength	= 128
	MaxAVMBackendMethodLength	= 96
	MaxAVMBackendRegistrySize	= 16
	MaxAVMNativeStateWriteZones	= 4
)

type AVMContractBackendKind string

type AVMContractBackendDescriptor struct {
	Kind			AVMContractBackendKind
	RouterBackend		RouterBackend
	Runtime			string
	Enabled			bool
	Optional		bool
	Deterministic		bool
	KeeperBacked		bool
	MsgServerCompatible	bool
	BackendHash		string
}

type AVMContractBackendRegistry struct {
	Backends	[]AVMContractBackendDescriptor
	RegistryHash	string
}

type AVMNativeModuleDescriptor struct {
	ModuleName		string
	ZoneID			zonestypes.ZoneID
	KeeperService		string
	MsgServerService	string
	ServiceInterfaceHash	string
	AllowedMessageTypes	[]string
	AllowedMethods		[]string
	DescriptorHash		string
}

type AVMNativeModuleRouteCall struct {
	Descriptor		AVMNativeModuleDescriptor
	RouteKey		string
	Method			string
	ZoneID			zonestypes.ZoneID
	Lane			RouterLane
	Backend			RouterBackend
	DispatchMode		RouterDispatchMode
	ReceiptPolicy		RouterReceiptPolicy
	GasMeter		RouterGasMeter
	Receipt			AVMExecutionReceipt
	CalledThroughAVM	bool
	UsesInterfaceSystem	bool
	StateWriteZones		[]zonestypes.ZoneID
	CallHash		string
}

func DefaultAVMContractBackendRegistry(runtime RuntimePolicy) (AVMContractBackendRegistry, error) {
	registry := AVMContractBackendRegistry{Backends: []AVMContractBackendDescriptor{
		{
			Kind:			AVMContractBackendNativeModule,
			RouterBackend:		RouterBackendNativeModule,
			Runtime:		"native",
			Enabled:		true,
			Optional:		false,
			Deterministic:		true,
			KeeperBacked:		true,
			MsgServerCompatible:	true,
		},
		{
			Kind:			AVMContractBackendActorContract,
			RouterBackend:		RouterBackendAVMActor,
			Runtime:		RuntimeAVM,
			Enabled:		runtime.AVMEnabled,
			Optional:		false,
			Deterministic:		true,
			KeeperBacked:		false,
			MsgServerCompatible:	false,
		},
		{
			Kind:			AVMContractBackendWASMContract,
			RouterBackend:		RouterBackendWASMAdapter,
			Runtime:		RuntimeCosmWasm,
			Enabled:		runtime.CosmWasmEnabled,
			Optional:		true,
			Deterministic:		true,
			KeeperBacked:		false,
			MsgServerCompatible:	false,
		},
	}}
	for i := range registry.Backends {
		registry.Backends[i] = canonicalAVMContractBackendDescriptor(registry.Backends[i])
		registry.Backends[i].BackendHash = ComputeAVMContractBackendHash(registry.Backends[i])
	}
	registry = canonicalAVMContractBackendRegistry(registry)
	registry.RegistryHash = ComputeAVMContractBackendRegistryHash(registry)
	return registry, registry.Validate()
}

func NewAVMNativeModuleDescriptor(descriptor AVMNativeModuleDescriptor) (AVMNativeModuleDescriptor, error) {
	descriptor = canonicalAVMNativeModuleDescriptor(descriptor)
	descriptor.DescriptorHash = ComputeAVMNativeModuleDescriptorHash(descriptor)
	return descriptor, descriptor.Validate()
}

func NewAVMNativeModuleRouteCall(call AVMNativeModuleRouteCall) (AVMNativeModuleRouteCall, error) {
	call = canonicalAVMNativeModuleRouteCall(call)
	call.CallHash = ComputeAVMNativeModuleRouteCallHash(call)
	return call, call.Validate()
}

func (d AVMContractBackendDescriptor) Validate() error {
	d = canonicalAVMContractBackendDescriptor(d)
	if !IsAVMContractBackendKind(d.Kind) {
		return fmt.Errorf("invalid AVM contract backend kind %q", d.Kind)
	}
	if !IsRouterBackend(d.RouterBackend) {
		return fmt.Errorf("invalid AVM contract router backend %q", d.RouterBackend)
	}
	if err := validateRouterOptionalToken("AVM contract backend runtime", d.Runtime, MaxAVMBackendNameLength); err != nil {
		return err
	}
	if d.Runtime == "" {
		return errors.New("AVM contract backend runtime is required")
	}
	if !d.Deterministic {
		return errors.New("AVM contract backend must be deterministic")
	}
	switch d.Kind {
	case AVMContractBackendNativeModule:
		if d.RouterBackend != RouterBackendNativeModule {
			return errors.New("native module backend must use native router backend")
		}
		if !d.Enabled || d.Optional {
			return errors.New("native module backend must be enabled and non-optional")
		}
		if !d.KeeperBacked || !d.MsgServerCompatible {
			return errors.New("native module backend must be keeper-backed and MsgServer compatible")
		}
	case AVMContractBackendActorContract:
		if d.RouterBackend != RouterBackendAVMActor || d.Runtime != RuntimeAVM {
			return errors.New("AVM actor contract backend must use AVM actor runtime")
		}
	case AVMContractBackendWASMContract:
		if d.RouterBackend != RouterBackendWASMAdapter || d.Runtime != RuntimeCosmWasm {
			return errors.New("WASM contract backend must use CosmWasm adapter runtime")
		}
		if !d.Optional {
			return errors.New("WASM contract backend must be optional")
		}
	}
	if d.BackendHash == "" {
		return errors.New("AVM contract backend hash is required")
	}
	if err := zonestypes.ValidateHash("AVM contract backend hash", d.BackendHash); err != nil {
		return err
	}
	if d.BackendHash != ComputeAVMContractBackendHash(d) {
		return errors.New("AVM contract backend hash mismatch")
	}
	return nil
}

func (r AVMContractBackendRegistry) Validate() error {
	r = canonicalAVMContractBackendRegistry(r)
	if len(r.Backends) == 0 {
		return errors.New("AVM contract backend registry must contain backends")
	}
	if len(r.Backends) > MaxAVMBackendRegistrySize {
		return fmt.Errorf("AVM contract backend registry must contain <= %d backends", MaxAVMBackendRegistrySize)
	}
	seen := make(map[AVMContractBackendKind]struct{}, len(r.Backends))
	required := map[AVMContractBackendKind]struct{}{
		AVMContractBackendNativeModule:		{},
		AVMContractBackendActorContract:	{},
		AVMContractBackendWASMContract:		{},
	}
	for i, backend := range r.Backends {
		if err := backend.Validate(); err != nil {
			return err
		}
		if _, found := seen[backend.Kind]; found {
			return fmt.Errorf("duplicate AVM contract backend kind %q", backend.Kind)
		}
		seen[backend.Kind] = struct{}{}
		if i > 0 && r.Backends[i-1].Kind >= backend.Kind {
			return errors.New("AVM contract backends must be sorted canonically")
		}
	}
	for kind := range required {
		if _, found := seen[kind]; !found {
			return fmt.Errorf("AVM contract backend registry missing %q", kind)
		}
	}
	if r.RegistryHash == "" {
		return errors.New("AVM contract backend registry hash is required")
	}
	if err := zonestypes.ValidateHash("AVM contract backend registry hash", r.RegistryHash); err != nil {
		return err
	}
	if r.RegistryHash != ComputeAVMContractBackendRegistryHash(r) {
		return errors.New("AVM contract backend registry hash mismatch")
	}
	return nil
}

func (d AVMNativeModuleDescriptor) Validate() error {
	d = canonicalAVMNativeModuleDescriptor(d)
	if err := validateRouterOptionalToken("AVM native module name", d.ModuleName, MaxAVMBackendNameLength); err != nil {
		return err
	}
	if d.ModuleName == "" {
		return errors.New("AVM native module name is required")
	}
	if err := zonestypes.ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM native module keeper service", value: d.KeeperService},
		{name: "AVM native module MsgServer service", value: d.MsgServerService},
	} {
		if err := validateRouterOptionalToken(item.name, item.value, MaxAVMBackendServiceLength); err != nil {
			return err
		}
		if item.value == "" {
			return fmt.Errorf("%s is required", item.name)
		}
	}
	if err := zonestypes.ValidateHash("AVM native module service interface hash", d.ServiceInterfaceHash); err != nil {
		return err
	}
	if err := validateNativeModuleTokens("AVM native module message type", d.AllowedMessageTypes); err != nil {
		return err
	}
	if err := validateNativeModuleTokens("AVM native module method", d.AllowedMethods); err != nil {
		return err
	}
	if d.DescriptorHash == "" {
		return errors.New("AVM native module descriptor hash is required")
	}
	if err := zonestypes.ValidateHash("AVM native module descriptor hash", d.DescriptorHash); err != nil {
		return err
	}
	if d.DescriptorHash != ComputeAVMNativeModuleDescriptorHash(d) {
		return errors.New("AVM native module descriptor hash mismatch")
	}
	return nil
}

func (c AVMNativeModuleRouteCall) Validate() error {
	c = canonicalAVMNativeModuleRouteCall(c)
	if err := c.Descriptor.Validate(); err != nil {
		return err
	}
	if c.ZoneID != c.Descriptor.ZoneID {
		return errors.New("AVM native module route zone mismatch")
	}
	if err := validateRouterOptionalToken("AVM native module route key", c.RouteKey, MaxRouterRouteKeyLength); err != nil {
		return err
	}
	if c.RouteKey == "" {
		return errors.New("AVM native module route key is required")
	}
	if !containsString(c.Descriptor.AllowedMethods, c.Method) {
		return fmt.Errorf("AVM native module method %q is not exposed by descriptor", c.Method)
	}
	if c.Backend != RouterBackendNativeModule {
		return errors.New("AVM native module route must use native backend")
	}
	if c.Lane != RouterLaneSync {
		return errors.New("AVM native module route must use sync engine lane")
	}
	if c.DispatchMode != RouterDispatchModeDirect {
		return errors.New("AVM native module route must use direct dispatch")
	}
	if c.ReceiptPolicy != RouterReceiptCommit {
		return errors.New("AVM native module route must commit execution receipt")
	}
	if err := c.GasMeter.Validate(); err != nil {
		return err
	}
	if c.CalledThroughAVM {
		if err := c.Receipt.Validate(); err != nil {
			return err
		}
		if c.Receipt.ZoneID != c.ZoneID {
			return errors.New("AVM native module receipt zone mismatch")
		}
	}
	if c.UsesInterfaceSystem && c.Descriptor.ServiceInterfaceHash == "" {
		return errors.New("AVM native module interface descriptor is required")
	}
	if err := validateNativeModuleStateWrites(c.ZoneID, c.StateWriteZones); err != nil {
		return err
	}
	if c.CallHash == "" {
		return errors.New("AVM native module route call hash is required")
	}
	if err := zonestypes.ValidateHash("AVM native module route call hash", c.CallHash); err != nil {
		return err
	}
	if c.CallHash != ComputeAVMNativeModuleRouteCallHash(c) {
		return errors.New("AVM native module route call hash mismatch")
	}
	return nil
}

func IsAVMContractBackendKind(kind AVMContractBackendKind) bool {
	switch kind {
	case AVMContractBackendNativeModule, AVMContractBackendWASMContract, AVMContractBackendActorContract:
		return true
	default:
		return false
	}
}

func ComputeAVMContractBackendHash(backend AVMContractBackendDescriptor) string {
	backend = canonicalAVMContractBackendDescriptor(backend)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-backend-v1")
	writeEnginePart(h, string(backend.Kind))
	writeEnginePart(h, string(backend.RouterBackend))
	writeEnginePart(h, backend.Runtime)
	writeEngineBool(h, backend.Enabled)
	writeEngineBool(h, backend.Optional)
	writeEngineBool(h, backend.Deterministic)
	writeEngineBool(h, backend.KeeperBacked)
	writeEngineBool(h, backend.MsgServerCompatible)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractBackendRegistryHash(registry AVMContractBackendRegistry) string {
	registry = canonicalAVMContractBackendRegistry(registry)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-backend-registry-v1")
	writeEngineUint64(h, uint64(len(registry.Backends)))
	for _, backend := range registry.Backends {
		writeEnginePart(h, backend.BackendHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMNativeModuleDescriptorHash(descriptor AVMNativeModuleDescriptor) string {
	descriptor = canonicalAVMNativeModuleDescriptor(descriptor)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-native-module-descriptor-v1")
	writeEnginePart(h, descriptor.ModuleName)
	writeEnginePart(h, string(descriptor.ZoneID))
	writeEnginePart(h, descriptor.KeeperService)
	writeEnginePart(h, descriptor.MsgServerService)
	writeEnginePart(h, descriptor.ServiceInterfaceHash)
	writeEngineUint64(h, uint64(len(descriptor.AllowedMessageTypes)))
	for _, msgType := range descriptor.AllowedMessageTypes {
		writeEnginePart(h, msgType)
	}
	writeEngineUint64(h, uint64(len(descriptor.AllowedMethods)))
	for _, method := range descriptor.AllowedMethods {
		writeEnginePart(h, method)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMNativeModuleRouteCallHash(call AVMNativeModuleRouteCall) string {
	call = canonicalAVMNativeModuleRouteCall(call)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-native-module-route-call-v1")
	writeEnginePart(h, call.Descriptor.DescriptorHash)
	writeEnginePart(h, call.RouteKey)
	writeEnginePart(h, call.Method)
	writeEnginePart(h, string(call.ZoneID))
	writeEnginePart(h, string(call.Lane))
	writeEnginePart(h, string(call.Backend))
	writeEnginePart(h, string(call.DispatchMode))
	writeEnginePart(h, string(call.ReceiptPolicy))
	writeEnginePart(h, string(call.GasMeter.Class))
	writeEngineUint64(h, call.GasMeter.Limit)
	writeEngineUint64(h, call.GasMeter.Reserved)
	writeEnginePart(h, call.Receipt.ReceiptHash)
	writeEngineBool(h, call.CalledThroughAVM)
	writeEngineBool(h, call.UsesInterfaceSystem)
	writeEngineUint64(h, uint64(len(call.StateWriteZones)))
	for _, zoneID := range call.StateWriteZones {
		writeEnginePart(h, string(zoneID))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMContractBackendDescriptor(backend AVMContractBackendDescriptor) AVMContractBackendDescriptor {
	backend.Runtime = strings.TrimSpace(backend.Runtime)
	backend.BackendHash = strings.TrimSpace(backend.BackendHash)
	return backend
}

func canonicalAVMContractBackendRegistry(registry AVMContractBackendRegistry) AVMContractBackendRegistry {
	registry.RegistryHash = strings.TrimSpace(registry.RegistryHash)
	registry.Backends = append([]AVMContractBackendDescriptor(nil), registry.Backends...)
	for i := range registry.Backends {
		registry.Backends[i] = canonicalAVMContractBackendDescriptor(registry.Backends[i])
	}
	sort.SliceStable(registry.Backends, func(i, j int) bool {
		return registry.Backends[i].Kind < registry.Backends[j].Kind
	})
	return registry
}

func canonicalAVMNativeModuleDescriptor(descriptor AVMNativeModuleDescriptor) AVMNativeModuleDescriptor {
	descriptor.ModuleName = strings.TrimSpace(descriptor.ModuleName)
	descriptor.KeeperService = strings.TrimSpace(descriptor.KeeperService)
	descriptor.MsgServerService = strings.TrimSpace(descriptor.MsgServerService)
	descriptor.ServiceInterfaceHash = strings.TrimSpace(descriptor.ServiceInterfaceHash)
	descriptor.DescriptorHash = strings.TrimSpace(descriptor.DescriptorHash)
	descriptor.AllowedMessageTypes = cloneSortedNativeModuleStrings(descriptor.AllowedMessageTypes)
	descriptor.AllowedMethods = cloneSortedNativeModuleStrings(descriptor.AllowedMethods)
	return descriptor
}

func canonicalAVMNativeModuleRouteCall(call AVMNativeModuleRouteCall) AVMNativeModuleRouteCall {
	call.Descriptor = canonicalAVMNativeModuleDescriptor(call.Descriptor)
	call.RouteKey = strings.TrimSpace(call.RouteKey)
	call.Method = strings.TrimSpace(call.Method)
	call.Receipt = canonicalAVMExecutionReceipt(call.Receipt)
	call.StateWriteZones = append([]zonestypes.ZoneID(nil), call.StateWriteZones...)
	sort.SliceStable(call.StateWriteZones, func(i, j int) bool {
		return call.StateWriteZones[i] < call.StateWriteZones[j]
	})
	call.CallHash = strings.TrimSpace(call.CallHash)
	return call
}

func validateNativeModuleTokens(fieldName string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s list must not be empty", fieldName)
	}
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		if err := validateRouterOptionalToken(fieldName, value, MaxAVMBackendMethodLength); err != nil {
			return err
		}
		if value == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s %q", fieldName, value)
		}
		seen[value] = struct{}{}
		if i > 0 && values[i-1] >= value {
			return fmt.Errorf("%s list must be sorted canonically", fieldName)
		}
	}
	return nil
}

func validateNativeModuleStateWrites(zoneID zonestypes.ZoneID, zones []zonestypes.ZoneID) error {
	if len(zones) == 0 {
		return errors.New("AVM native module route must declare state write zones")
	}
	if len(zones) > MaxAVMNativeStateWriteZones {
		return fmt.Errorf("AVM native module state write zones must be <= %d", MaxAVMNativeStateWriteZones)
	}
	for i, writeZone := range zones {
		if err := zonestypes.ValidateZoneID(writeZone); err != nil {
			return err
		}
		if writeZone != zoneID {
			return errors.New("AVM native module must not directly mutate other zones")
		}
		if i > 0 && zones[i-1] >= writeZone {
			return errors.New("AVM native module state write zones must be sorted canonically")
		}
	}
	return nil
}

func cloneSortedNativeModuleStrings(values []string) []string {
	out := append([]string(nil), values...)
	for i := range out {
		out[i] = strings.TrimSpace(out[i])
	}
	sort.Strings(out)
	return out
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
