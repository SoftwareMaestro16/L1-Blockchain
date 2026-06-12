package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/wasmconfig"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMWASMHostStorageRead	AVMWASMHostFunctionKind	= "storage_read"
	AVMWASMHostStorageWrite	AVMWASMHostFunctionKind	= "storage_write"
	AVMWASMHostCrypto	AVMWASMHostFunctionKind	= "crypto"
	AVMWASMHostQuery	AVMWASMHostFunctionKind	= "query"
	AVMWASMHostAsyncEmit	AVMWASMHostFunctionKind	= "async_emit"

	WASMMemoryPageBytes	uint64	= 64 * 1024

	MaxAVMWASMHostFunctionName	= 128
)

type AVMWASMHostFunctionKind string

type AVMWASMHostFunction struct {
	Name		string
	Kind		AVMWASMHostFunctionKind
	Deterministic	bool
	GasCost		uint64
}

type AVMWASMGasConversionTable struct {
	WASMGasUnit		uint64
	AVMGasPerWASMGas	uint64
	StorageReadGas		uint64
	StorageWriteGas		uint64
	MessageEmitGas		uint64
	ConversionHash		string
}

type AVMWASMStoreV2KVAdapter struct {
	ZoneID		zonestypes.ZoneID
	StoreKey	string
	KeyPrefix	string
	MaxKeyBytes	uint32
	MaxValueBytes	uint64
	AdapterHash	string
}

type AVMWASMSandboxPolicy struct {
	Enabled			bool
	Optional		bool
	RuntimePolicy		wasmconfig.Policy
	MaxMemoryPages		uint32
	HostFunctions		[]AVMWASMHostFunction
	GasConversion		AVMWASMGasConversionTable
	StoreAdapter		AVMWASMStoreV2KVAdapter
	ExternalNetwork		bool
	CrossZoneAsyncOnly	bool
	SandboxPolicyHash	string
}

type AVMWASMContractRouteCall struct {
	SandboxPolicy		AVMWASMSandboxPolicy
	Call			VMCall
	ZoneID			zonestypes.ZoneID
	Backend			RouterBackend
	DispatchMode		RouterDispatchMode
	GasMeter		RouterGasMeter
	EmittedMessages		[]AVMAsyncMessage
	DirectCrossZoneCall	bool
	NetworkAccessAttempt	bool
	RouteCallHash		string
}

func DefaultAVMWASMHostFunctions() []AVMWASMHostFunction {
	return []AVMWASMHostFunction{
		{Name: "addr_validate", Kind: AVMWASMHostQuery, Deterministic: true, GasCost: 1},
		{Name: "crypto_verify", Kind: AVMWASMHostCrypto, Deterministic: true, GasCost: 5},
		{Name: "db_read", Kind: AVMWASMHostStorageRead, Deterministic: true, GasCost: 2},
		{Name: "db_write", Kind: AVMWASMHostStorageWrite, Deterministic: true, GasCost: 4},
		{Name: "emit_async", Kind: AVMWASMHostAsyncEmit, Deterministic: true, GasCost: 10},
	}
}

func DefaultAVMWASMGasConversionTable() (AVMWASMGasConversionTable, error) {
	table := AVMWASMGasConversionTable{
		WASMGasUnit:		1,
		AVMGasPerWASMGas:	wasmconfig.DefaultGasMultiplier,
		StorageReadGas:		2,
		StorageWriteGas:	4,
		MessageEmitGas:		10,
	}
	table.ConversionHash = ComputeAVMWASMGasConversionHash(table)
	return table, table.Validate()
}

func NewAVMWASMStoreV2KVAdapter(adapter AVMWASMStoreV2KVAdapter) (AVMWASMStoreV2KVAdapter, error) {
	adapter = canonicalAVMWASMStoreV2KVAdapter(adapter)
	if adapter.StoreKey == "" {
		adapter.StoreKey = DefaultAVMStoreKey
	}
	if adapter.KeyPrefix == "" {
		adapter.KeyPrefix = ContractZoneKVPrefix(adapter.ZoneID)
	}
	adapter.AdapterHash = ComputeAVMWASMStoreV2KVAdapterHash(adapter)
	return adapter, adapter.Validate()
}

func NewAVMWASMSandboxPolicy(policy AVMWASMSandboxPolicy) (AVMWASMSandboxPolicy, error) {
	policy = canonicalAVMWASMSandboxPolicy(policy)
	if len(policy.HostFunctions) == 0 {
		policy.HostFunctions = DefaultAVMWASMHostFunctions()
	}
	if policy.GasConversion.ConversionHash == "" {
		table, err := DefaultAVMWASMGasConversionTable()
		if err != nil {
			return AVMWASMSandboxPolicy{}, err
		}
		policy.GasConversion = table
	}
	if policy.StoreAdapter.AdapterHash == "" {
		adapter, err := NewAVMWASMStoreV2KVAdapter(policy.StoreAdapter)
		if err != nil {
			return AVMWASMSandboxPolicy{}, err
		}
		policy.StoreAdapter = adapter
	}
	if policy.MaxMemoryPages == 0 {
		policy.MaxMemoryPages = uint32((uint64(policy.RuntimePolicy.MemoryCacheSizeMiB) * 1024 * 1024) / WASMMemoryPageBytes)
	}
	policy.SandboxPolicyHash = ComputeAVMWASMSandboxPolicyHash(policy)
	return policy, policy.Validate()
}

func NewAVMWASMContractRouteCall(call AVMWASMContractRouteCall) (AVMWASMContractRouteCall, error) {
	call = canonicalAVMWASMContractRouteCall(call)
	call.RouteCallHash = ComputeAVMWASMContractRouteCallHash(call)
	return call, call.Validate()
}

func (f AVMWASMHostFunction) Validate() error {
	f = canonicalAVMWASMHostFunction(f)
	if err := validateRouterOptionalToken("AVM WASM host function name", f.Name, MaxAVMWASMHostFunctionName); err != nil {
		return err
	}
	if f.Name == "" {
		return errors.New("AVM WASM host function name is required")
	}
	if !IsAVMWASMHostFunctionKind(f.Kind) {
		return fmt.Errorf("invalid AVM WASM host function kind %q", f.Kind)
	}
	if !f.Deterministic {
		return errors.New("AVM WASM host functions must be deterministic")
	}
	if f.GasCost == 0 {
		return errors.New("AVM WASM host function gas cost must be positive")
	}
	return nil
}

func (t AVMWASMGasConversionTable) Validate() error {
	if t.WASMGasUnit == 0 || t.AVMGasPerWASMGas == 0 || t.StorageReadGas == 0 || t.StorageWriteGas == 0 || t.MessageEmitGas == 0 {
		return errors.New("AVM WASM gas conversion table values must be positive")
	}
	if t.AVMGasPerWASMGas != wasmconfig.DefaultGasMultiplier {
		return errors.New("AVM WASM gas conversion multiplier must match benchmarked wasm policy")
	}
	if t.ConversionHash == "" {
		return errors.New("AVM WASM gas conversion hash is required")
	}
	if err := zonestypes.ValidateHash("AVM WASM gas conversion hash", t.ConversionHash); err != nil {
		return err
	}
	if t.ConversionHash != ComputeAVMWASMGasConversionHash(t) {
		return errors.New("AVM WASM gas conversion hash mismatch")
	}
	return nil
}

func (a AVMWASMStoreV2KVAdapter) Validate() error {
	a = canonicalAVMWASMStoreV2KVAdapter(a)
	if err := zonestypes.ValidateZoneID(a.ZoneID); err != nil {
		return err
	}
	if err := validateSDKToken("AVM WASM Store v2 adapter store key", a.StoreKey); err != nil {
		return err
	}
	expectedPrefix := ContractZoneKVPrefix(a.ZoneID)
	if a.KeyPrefix != expectedPrefix {
		return fmt.Errorf("AVM WASM Store v2 adapter prefix must be %q", expectedPrefix)
	}
	if a.MaxKeyBytes == 0 {
		return errors.New("AVM WASM Store v2 adapter max key bytes must be positive")
	}
	if a.MaxValueBytes == 0 {
		return errors.New("AVM WASM Store v2 adapter max value bytes must be positive")
	}
	if a.AdapterHash == "" {
		return errors.New("AVM WASM Store v2 adapter hash is required")
	}
	if err := zonestypes.ValidateHash("AVM WASM Store v2 adapter hash", a.AdapterHash); err != nil {
		return err
	}
	if a.AdapterHash != ComputeAVMWASMStoreV2KVAdapterHash(a) {
		return errors.New("AVM WASM Store v2 adapter hash mismatch")
	}
	return nil
}

func (p AVMWASMSandboxPolicy) Validate() error {
	p = canonicalAVMWASMSandboxPolicy(p)
	if !p.Optional {
		return errors.New("AVM WASM backend must remain optional")
	}
	if !p.Enabled {
		return errors.New("AVM WASM sandbox policy must be enabled for execution")
	}
	wasmPolicy := p.RuntimePolicy
	wasmPolicy.Enabled = true
	if err := wasmPolicy.Validate(); err != nil {
		return err
	}
	if p.ExternalNetwork {
		return errors.New("AVM WASM sandbox forbids external network access")
	}
	if !p.CrossZoneAsyncOnly {
		return errors.New("AVM WASM cross-zone calls must be async messages only")
	}
	if p.MaxMemoryPages == 0 {
		return errors.New("AVM WASM memory pages must be bounded")
	}
	maxBytes := uint64(wasmPolicy.MemoryCacheSizeMiB) * 1024 * 1024
	if uint64(p.MaxMemoryPages)*WASMMemoryPageBytes > maxBytes {
		return errors.New("AVM WASM memory bound exceeds runtime policy")
	}
	if len(p.HostFunctions) == 0 {
		return errors.New("AVM WASM sandbox must declare host functions")
	}
	seen := make(map[string]struct{}, len(p.HostFunctions))
	for i, host := range p.HostFunctions {
		if err := host.Validate(); err != nil {
			return err
		}
		if _, found := seen[host.Name]; found {
			return fmt.Errorf("duplicate AVM WASM host function %q", host.Name)
		}
		seen[host.Name] = struct{}{}
		if i > 0 && p.HostFunctions[i-1].Name >= host.Name {
			return errors.New("AVM WASM host functions must be sorted canonically")
		}
	}
	if err := p.GasConversion.Validate(); err != nil {
		return err
	}
	if err := p.StoreAdapter.Validate(); err != nil {
		return err
	}
	if p.SandboxPolicyHash == "" {
		return errors.New("AVM WASM sandbox policy hash is required")
	}
	if err := zonestypes.ValidateHash("AVM WASM sandbox policy hash", p.SandboxPolicyHash); err != nil {
		return err
	}
	if p.SandboxPolicyHash != ComputeAVMWASMSandboxPolicyHash(p) {
		return errors.New("AVM WASM sandbox policy hash mismatch")
	}
	return nil
}

func (c AVMWASMContractRouteCall) Validate() error {
	c = canonicalAVMWASMContractRouteCall(c)
	if err := c.SandboxPolicy.Validate(); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if c.Backend != RouterBackendWASMAdapter {
		return errors.New("AVM WASM contract route must use WASM adapter backend")
	}
	if c.DispatchMode != RouterDispatchModeQueued && c.DispatchMode != RouterDispatchModeCrossZone {
		return errors.New("AVM WASM contract route must be queued or cross-zone async")
	}
	if err := c.GasMeter.Validate(); err != nil {
		return err
	}
	runtime := DefaultRuntimePolicy()
	runtime.CosmWasmEnabled = true
	runtime.CosmWasmPolicy = c.SandboxPolicy.RuntimePolicy
	runtime.CosmWasmPolicy.Enabled = true
	if err := ValidateVMCall(c.Call, runtime); err != nil {
		return err
	}
	if c.Call.Runtime != RuntimeCosmWasm {
		return errors.New("AVM WASM contract route must execute CosmWasm runtime")
	}
	if c.NetworkAccessAttempt {
		return errors.New("AVM WASM contract attempted external network access")
	}
	if c.DirectCrossZoneCall {
		return errors.New("AVM WASM cross-zone calls must use async messages")
	}
	for _, msg := range c.EmittedMessages {
		msg = canonicalAVMAsyncMessage(msg)
		if err := msg.Validate(); err != nil {
			return err
		}
		if msg.SourceZone != msg.DestinationZone && c.DispatchMode != RouterDispatchModeCrossZone {
			return errors.New("AVM WASM cross-zone emitted message requires cross-zone async dispatch")
		}
	}
	if c.RouteCallHash == "" {
		return errors.New("AVM WASM route call hash is required")
	}
	if err := zonestypes.ValidateHash("AVM WASM route call hash", c.RouteCallHash); err != nil {
		return err
	}
	if c.RouteCallHash != ComputeAVMWASMContractRouteCallHash(c) {
		return errors.New("AVM WASM route call hash mismatch")
	}
	return nil
}

func IsAVMWASMHostFunctionKind(kind AVMWASMHostFunctionKind) bool {
	switch kind {
	case AVMWASMHostStorageRead, AVMWASMHostStorageWrite, AVMWASMHostCrypto, AVMWASMHostQuery, AVMWASMHostAsyncEmit:
		return true
	default:
		return false
	}
}

func ComputeAVMWASMGasConversionHash(table AVMWASMGasConversionTable) string {
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-wasm-gas-conversion-v1")
	writeEngineUint64(h, table.WASMGasUnit)
	writeEngineUint64(h, table.AVMGasPerWASMGas)
	writeEngineUint64(h, table.StorageReadGas)
	writeEngineUint64(h, table.StorageWriteGas)
	writeEngineUint64(h, table.MessageEmitGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMWASMStoreV2KVAdapterHash(adapter AVMWASMStoreV2KVAdapter) string {
	adapter = canonicalAVMWASMStoreV2KVAdapter(adapter)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-wasm-store-v2-adapter-v1")
	writeEnginePart(h, string(adapter.ZoneID))
	writeEnginePart(h, adapter.StoreKey)
	writeEnginePart(h, adapter.KeyPrefix)
	writeEngineUint64(h, uint64(adapter.MaxKeyBytes))
	writeEngineUint64(h, adapter.MaxValueBytes)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMWASMSandboxPolicyHash(policy AVMWASMSandboxPolicy) string {
	policy = canonicalAVMWASMSandboxPolicy(policy)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-wasm-sandbox-policy-v1")
	writeEngineBool(h, policy.Enabled)
	writeEngineBool(h, policy.Optional)
	writeEngineUint64(h, policy.RuntimePolicy.MaxContractSizeBytes)
	writeEngineUint64(h, policy.RuntimePolicy.MaxProposalContractSizeBytes)
	writeEngineUint64(h, policy.RuntimePolicy.SmartQueryGasLimit)
	writeEngineUint64(h, policy.RuntimePolicy.SimulationGasLimit)
	writeEngineUint64(h, policy.RuntimePolicy.GasMultiplier)
	writeEngineUint64(h, uint64(policy.RuntimePolicy.MemoryCacheSizeMiB))
	writeEngineUint64(h, policy.RuntimePolicy.MaxQueryResponseBytes)
	writeEngineUint64(h, uint64(policy.RuntimePolicy.MaxQueryDepth))
	writeEngineUint64(h, uint64(policy.MaxMemoryPages))
	writeEngineUint64(h, uint64(len(policy.HostFunctions)))
	for _, host := range policy.HostFunctions {
		writeEnginePart(h, host.Name)
		writeEnginePart(h, string(host.Kind))
		writeEngineBool(h, host.Deterministic)
		writeEngineUint64(h, host.GasCost)
	}
	writeEnginePart(h, policy.GasConversion.ConversionHash)
	writeEnginePart(h, policy.StoreAdapter.AdapterHash)
	writeEngineBool(h, policy.ExternalNetwork)
	writeEngineBool(h, policy.CrossZoneAsyncOnly)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMWASMContractRouteCallHash(call AVMWASMContractRouteCall) string {
	call = canonicalAVMWASMContractRouteCall(call)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-wasm-contract-route-call-v1")
	writeEnginePart(h, call.SandboxPolicy.SandboxPolicyHash)
	writeEnginePart(h, call.Call.Runtime)
	writeEnginePart(h, call.Call.Action)
	writeEngineUint64(h, call.Call.CodeBytes)
	writeEngineUint64(h, call.Call.GasLimit)
	writeEngineUint64(h, call.Call.QueryBytes)
	writeEngineUint64(h, uint64(call.Call.QueryDepth))
	writeEnginePart(h, string(call.ZoneID))
	writeEnginePart(h, string(call.Backend))
	writeEnginePart(h, string(call.DispatchMode))
	writeEnginePart(h, string(call.GasMeter.Class))
	writeEngineUint64(h, call.GasMeter.Limit)
	writeEngineUint64(h, call.GasMeter.Reserved)
	writeEngineUint64(h, uint64(len(call.EmittedMessages)))
	for _, msg := range call.EmittedMessages {
		writeEnginePart(h, msg.ID)
	}
	writeEngineBool(h, call.DirectCrossZoneCall)
	writeEngineBool(h, call.NetworkAccessAttempt)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMWASMHostFunction(host AVMWASMHostFunction) AVMWASMHostFunction {
	host.Name = strings.TrimSpace(host.Name)
	return host
}

func canonicalAVMWASMStoreV2KVAdapter(adapter AVMWASMStoreV2KVAdapter) AVMWASMStoreV2KVAdapter {
	adapter.StoreKey = strings.TrimSpace(adapter.StoreKey)
	adapter.KeyPrefix = strings.TrimSpace(adapter.KeyPrefix)
	adapter.AdapterHash = strings.TrimSpace(adapter.AdapterHash)
	return adapter
}

func canonicalAVMWASMSandboxPolicy(policy AVMWASMSandboxPolicy) AVMWASMSandboxPolicy {
	policy.SandboxPolicyHash = strings.TrimSpace(policy.SandboxPolicyHash)
	policy.HostFunctions = append([]AVMWASMHostFunction(nil), policy.HostFunctions...)
	for i := range policy.HostFunctions {
		policy.HostFunctions[i] = canonicalAVMWASMHostFunction(policy.HostFunctions[i])
	}
	sort.SliceStable(policy.HostFunctions, func(i, j int) bool {
		return policy.HostFunctions[i].Name < policy.HostFunctions[j].Name
	})
	policy.StoreAdapter = canonicalAVMWASMStoreV2KVAdapter(policy.StoreAdapter)
	return policy
}

func canonicalAVMWASMContractRouteCall(call AVMWASMContractRouteCall) AVMWASMContractRouteCall {
	call.SandboxPolicy = canonicalAVMWASMSandboxPolicy(call.SandboxPolicy)
	call.EmittedMessages = append([]AVMAsyncMessage(nil), call.EmittedMessages...)
	for i := range call.EmittedMessages {
		call.EmittedMessages[i] = canonicalAVMAsyncMessage(call.EmittedMessages[i])
	}
	sort.SliceStable(call.EmittedMessages, func(i, j int) bool {
		return call.EmittedMessages[i].ID < call.EmittedMessages[j].ID
	})
	call.RouteCallHash = strings.TrimSpace(call.RouteCallHash)
	return call
}
