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
	VMAdapterAVM		VMAdapterKind	= "AVM"
	VMAdapterCosmWasm	VMAdapterKind	= "COSMWASM"

	AVMBytecodeNative		AVMBytecodeKind	= "native_bytecode"
	AVMBytecodeIntermediateIR	AVMBytecodeKind	= "intermediate_ir"

	MaxVMBoundaryTokenLength	= 128
	MaxVMSyscalls			= 16
)

type VMAdapterKind string
type AVMBytecodeKind string

type VMDeterminismProfile struct {
	Runtime				string
	NoExternalAPICalls		bool
	NoTimeBasedRandomness		bool
	SortedMessageApplication	bool
	BoundedIteration		bool
	MaxIterationCount		uint64
	BoundedMemory			bool
	MaxMemoryBytes			uint64
	ReproducibleStateTransition	bool
	MeteredStorageAccess		bool
	MeteredProofVerification	bool
	ProfileHash			string
}

type VMSyscallMeter struct {
	Name		string
	GasClass	AVMGasClass
	GasCost		uint64
	Metered		bool
}

type AVMAdapterBoundary struct {
	BytecodeKind			AVMBytecodeKind
	Runtime				string
	DeterministicGasSchedule	AVMGasSchedule
	StoreKey			string
	StoreV2Backed			bool
	KVPrefix			string
	MessageSyscall			VMSyscallMeter
	ProofVerificationSyscall	VMSyscallMeter
	BoundaryHash			string
}

type CosmWasmAdapterBoundary struct {
	Runtime				string
	IsolatedAdapterModule		bool
	GasConversion			AVMWASMGasConversionTable
	StoreAdapter			AVMWASMStoreV2KVAdapter
	ExplicitStorageKeyPrefix	bool
	DirectNonContractState		bool
	CrossZoneMessagesOrProofs	bool
	ExternalNetwork			bool
	HostFunctions			[]AVMWASMHostFunction
	BoundaryHash			string
}

type VMAdapterBoundaryManifest struct {
	ZoneID			zonestypes.ZoneID
	DeterminismProfile	VMDeterminismProfile
	AVM			AVMAdapterBoundary
	CosmWasm		CosmWasmAdapterBoundary
	ManifestHash		string
}

func DefaultVMDeterminismProfile(runtime string) (VMDeterminismProfile, error) {
	profile := VMDeterminismProfile{
		Runtime:			strings.TrimSpace(runtime),
		NoExternalAPICalls:		true,
		NoTimeBasedRandomness:		true,
		SortedMessageApplication:	true,
		BoundedIteration:		true,
		MaxIterationCount:		10_000,
		BoundedMemory:			true,
		MaxMemoryBytes:			64 * 1024 * 1024,
		ReproducibleStateTransition:	true,
		MeteredStorageAccess:		true,
		MeteredProofVerification:	true,
	}
	profile.ProfileHash = ComputeVMDeterminismProfileHash(profile)
	return profile, profile.Validate()
}

func NewAVMAdapterBoundary(boundary AVMAdapterBoundary, zoneID zonestypes.ZoneID) (AVMAdapterBoundary, error) {
	boundary = canonicalAVMAdapterBoundary(boundary)
	if boundary.Runtime == "" {
		boundary.Runtime = RuntimeAVM
	}
	if boundary.BytecodeKind == "" {
		boundary.BytecodeKind = AVMBytecodeNative
	}
	if boundary.StoreKey == "" {
		boundary.StoreKey = DefaultAVMStoreKey
	}
	if boundary.KVPrefix == "" {
		boundary.KVPrefix = ContractZoneKVPrefix(zoneID)
	}
	if boundary.DeterministicGasSchedule.ScheduleHash == "" {
		schedule, err := DefaultAVMGasSchedule()
		if err != nil {
			return AVMAdapterBoundary{}, err
		}
		boundary.DeterministicGasSchedule = schedule
	}
	if boundary.MessageSyscall.Name == "" {
		boundary.MessageSyscall = VMSyscallMeter{Name: "message_emit", GasClass: AVMGasClassCrossZoneRouting, GasCost: boundary.DeterministicGasSchedule.CrossZoneRoutingGas, Metered: true}
	}
	if boundary.ProofVerificationSyscall.Name == "" {
		boundary.ProofVerificationSyscall = VMSyscallMeter{Name: "proof_verify", GasClass: AVMGasClassProofVerification, GasCost: boundary.DeterministicGasSchedule.ProofVerificationGas, Metered: true}
	}
	boundary.StoreV2Backed = true
	boundary.BoundaryHash = ComputeAVMAdapterBoundaryHash(boundary)
	return boundary, boundary.ValidateForZone(zoneID)
}

func NewCosmWasmAdapterBoundary(boundary CosmWasmAdapterBoundary, zoneID zonestypes.ZoneID) (CosmWasmAdapterBoundary, error) {
	boundary = canonicalCosmWasmAdapterBoundary(boundary)
	if boundary.Runtime == "" {
		boundary.Runtime = RuntimeCosmWasm
	}
	if boundary.GasConversion.ConversionHash == "" {
		table, err := DefaultAVMWASMGasConversionTable()
		if err != nil {
			return CosmWasmAdapterBoundary{}, err
		}
		boundary.GasConversion = table
	}
	if boundary.StoreAdapter.AdapterHash == "" {
		adapter, err := NewAVMWASMStoreV2KVAdapter(AVMWASMStoreV2KVAdapter{
			ZoneID:		zoneID,
			StoreKey:	DefaultAVMStoreKey,
			KeyPrefix:	ContractZoneKVPrefix(zoneID),
			MaxKeyBytes:	DefaultMaxStorageKeyBytes,
			MaxValueBytes:	DefaultMaxStorageValueBytes,
		})
		if err != nil {
			return CosmWasmAdapterBoundary{}, err
		}
		boundary.StoreAdapter = adapter
	}
	if len(boundary.HostFunctions) == 0 {
		boundary.HostFunctions = DefaultAVMWASMHostFunctions()
	}
	boundary.IsolatedAdapterModule = true
	boundary.ExplicitStorageKeyPrefix = true
	boundary.CrossZoneMessagesOrProofs = true
	boundary.BoundaryHash = ComputeCosmWasmAdapterBoundaryHash(boundary)
	return boundary, boundary.Validate()
}

func NewVMAdapterBoundaryManifest(manifest VMAdapterBoundaryManifest) (VMAdapterBoundaryManifest, error) {
	manifest = canonicalVMAdapterBoundaryManifest(manifest)
	if err := zonestypes.ValidateZoneID(manifest.ZoneID); err != nil {
		return VMAdapterBoundaryManifest{}, err
	}
	if manifest.DeterminismProfile.ProfileHash == "" {
		profile, err := DefaultVMDeterminismProfile(RuntimeAVM)
		if err != nil {
			return VMAdapterBoundaryManifest{}, err
		}
		manifest.DeterminismProfile = profile
	}
	if manifest.AVM.BoundaryHash == "" {
		avmBoundary, err := NewAVMAdapterBoundary(manifest.AVM, manifest.ZoneID)
		if err != nil {
			return VMAdapterBoundaryManifest{}, err
		}
		manifest.AVM = avmBoundary
	}
	if manifest.CosmWasm.BoundaryHash == "" {
		wasmBoundary, err := NewCosmWasmAdapterBoundary(manifest.CosmWasm, manifest.ZoneID)
		if err != nil {
			return VMAdapterBoundaryManifest{}, err
		}
		manifest.CosmWasm = wasmBoundary
	}
	manifest.ManifestHash = ComputeVMAdapterBoundaryManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (p VMDeterminismProfile) Validate() error {
	p = canonicalVMDeterminismProfile(p)
	if p.Runtime != RuntimeAVM && p.Runtime != RuntimeCosmWasm {
		return fmt.Errorf("VM determinism profile runtime %q is not supported", p.Runtime)
	}
	if !p.NoExternalAPICalls {
		return errors.New("VM determinism forbids external API calls")
	}
	if !p.NoTimeBasedRandomness {
		return errors.New("VM determinism forbids time-based randomness")
	}
	if !p.SortedMessageApplication {
		return errors.New("VM determinism requires sorted message application")
	}
	if !p.BoundedIteration || p.MaxIterationCount == 0 {
		return errors.New("VM determinism requires bounded iteration")
	}
	if !p.BoundedMemory || p.MaxMemoryBytes == 0 {
		return errors.New("VM determinism requires bounded memory")
	}
	if !p.ReproducibleStateTransition {
		return errors.New("VM determinism requires reproducible state transitions")
	}
	if !p.MeteredStorageAccess {
		return errors.New("VM determinism requires metered storage access")
	}
	if !p.MeteredProofVerification {
		return errors.New("VM determinism requires metered proof verification")
	}
	if p.ProfileHash == "" {
		return errors.New("VM determinism profile hash is required")
	}
	if err := zonestypes.ValidateHash("VM determinism profile hash", p.ProfileHash); err != nil {
		return err
	}
	if p.ProfileHash != ComputeVMDeterminismProfileHash(p) {
		return errors.New("VM determinism profile hash mismatch")
	}
	return nil
}

func (s VMSyscallMeter) Validate() error {
	s = canonicalVMSyscallMeter(s)
	if err := validateRouterOptionalToken("VM syscall name", s.Name, MaxVMBoundaryTokenLength); err != nil {
		return err
	}
	if s.Name == "" {
		return errors.New("VM syscall name is required")
	}
	if !IsAVMGasClass(s.GasClass) {
		return fmt.Errorf("invalid VM syscall gas class %q", s.GasClass)
	}
	if !s.Metered || s.GasCost == 0 {
		return errors.New("VM syscall must be metered with positive gas")
	}
	return nil
}

func (b AVMAdapterBoundary) ValidateForZone(zoneID zonestypes.ZoneID) error {
	b = canonicalAVMAdapterBoundary(b)
	if !IsAVMBytecodeKind(b.BytecodeKind) {
		return fmt.Errorf("invalid AVM bytecode kind %q", b.BytecodeKind)
	}
	if b.Runtime != RuntimeAVM {
		return errors.New("AVM adapter boundary must use AVM runtime")
	}
	if err := b.DeterministicGasSchedule.Validate(); err != nil {
		return err
	}
	if b.StoreKey != DefaultAVMStoreKey {
		return fmt.Errorf("AVM adapter Store v2 key must be %q", DefaultAVMStoreKey)
	}
	if !b.StoreV2Backed {
		return errors.New("AVM adapter must be Store v2-backed")
	}
	expectedPrefix := ContractZoneKVPrefix(zoneID)
	if b.KVPrefix != expectedPrefix {
		return fmt.Errorf("AVM adapter KV prefix must be %q", expectedPrefix)
	}
	if err := b.MessageSyscall.Validate(); err != nil {
		return err
	}
	if b.MessageSyscall.GasClass != AVMGasClassCrossZoneRouting {
		return errors.New("AVM message syscall must use cross-zone routing gas")
	}
	if err := b.ProofVerificationSyscall.Validate(); err != nil {
		return err
	}
	if b.ProofVerificationSyscall.GasClass != AVMGasClassProofVerification {
		return errors.New("AVM proof verification syscall must use proof verification gas")
	}
	if b.BoundaryHash == "" {
		return errors.New("AVM adapter boundary hash is required")
	}
	if err := zonestypes.ValidateHash("AVM adapter boundary hash", b.BoundaryHash); err != nil {
		return err
	}
	if b.BoundaryHash != ComputeAVMAdapterBoundaryHash(b) {
		return errors.New("AVM adapter boundary hash mismatch")
	}
	return nil
}

func (b CosmWasmAdapterBoundary) Validate() error {
	b = canonicalCosmWasmAdapterBoundary(b)
	if b.Runtime != RuntimeCosmWasm {
		return errors.New("CosmWasm adapter boundary must use CosmWasm runtime")
	}
	if !b.IsolatedAdapterModule {
		return errors.New("CosmWasm adapter must be isolated")
	}
	if err := b.GasConversion.Validate(); err != nil {
		return err
	}
	if err := b.StoreAdapter.Validate(); err != nil {
		return err
	}
	if !b.ExplicitStorageKeyPrefix {
		return errors.New("CosmWasm adapter requires explicit storage key prefixing")
	}
	if b.DirectNonContractState {
		return errors.New("CosmWasm adapter cannot access non-contract zone state directly")
	}
	if !b.CrossZoneMessagesOrProofs {
		return errors.New("CosmWasm adapter cross-zone access must use messages or proofs")
	}
	if b.ExternalNetwork {
		return errors.New("CosmWasm adapter forbids external network access")
	}
	if len(b.HostFunctions) == 0 || len(b.HostFunctions) > MaxVMSyscalls {
		return fmt.Errorf("CosmWasm adapter host functions must be between 1 and %d", MaxVMSyscalls)
	}
	for i, host := range b.HostFunctions {
		if err := host.Validate(); err != nil {
			return err
		}
		if i > 0 && b.HostFunctions[i-1].Name >= host.Name {
			return errors.New("CosmWasm adapter host functions must be sorted canonically")
		}
	}
	if b.BoundaryHash == "" {
		return errors.New("CosmWasm adapter boundary hash is required")
	}
	if err := zonestypes.ValidateHash("CosmWasm adapter boundary hash", b.BoundaryHash); err != nil {
		return err
	}
	if b.BoundaryHash != ComputeCosmWasmAdapterBoundaryHash(b) {
		return errors.New("CosmWasm adapter boundary hash mismatch")
	}
	return nil
}

func (m VMAdapterBoundaryManifest) Validate() error {
	m = canonicalVMAdapterBoundaryManifest(m)
	if err := zonestypes.ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	if err := m.DeterminismProfile.Validate(); err != nil {
		return err
	}
	if err := m.AVM.ValidateForZone(m.ZoneID); err != nil {
		return err
	}
	if err := m.CosmWasm.Validate(); err != nil {
		return err
	}
	if m.CosmWasm.StoreAdapter.ZoneID != m.ZoneID {
		return errors.New("VM adapter manifest CosmWasm store adapter zone mismatch")
	}
	if m.ManifestHash == "" {
		return errors.New("VM adapter boundary manifest hash is required")
	}
	if err := zonestypes.ValidateHash("VM adapter boundary manifest hash", m.ManifestHash); err != nil {
		return err
	}
	if m.ManifestHash != ComputeVMAdapterBoundaryManifestHash(m) {
		return errors.New("VM adapter boundary manifest hash mismatch")
	}
	return nil
}

func IsAVMBytecodeKind(kind AVMBytecodeKind) bool {
	switch kind {
	case AVMBytecodeNative, AVMBytecodeIntermediateIR:
		return true
	default:
		return false
	}
}

func ComputeVMDeterminismProfileHash(profile VMDeterminismProfile) string {
	profile = canonicalVMDeterminismProfile(profile)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-determinism-profile-v1")
	writeEnginePart(h, profile.Runtime)
	writeEngineBool(h, profile.NoExternalAPICalls)
	writeEngineBool(h, profile.NoTimeBasedRandomness)
	writeEngineBool(h, profile.SortedMessageApplication)
	writeEngineBool(h, profile.BoundedIteration)
	writeEngineUint64(h, profile.MaxIterationCount)
	writeEngineBool(h, profile.BoundedMemory)
	writeEngineUint64(h, profile.MaxMemoryBytes)
	writeEngineBool(h, profile.ReproducibleStateTransition)
	writeEngineBool(h, profile.MeteredStorageAccess)
	writeEngineBool(h, profile.MeteredProofVerification)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAdapterBoundaryHash(boundary AVMAdapterBoundary) string {
	boundary = canonicalAVMAdapterBoundary(boundary)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-adapter-boundary-v1")
	writeEnginePart(h, string(boundary.BytecodeKind))
	writeEnginePart(h, boundary.Runtime)
	writeEnginePart(h, boundary.DeterministicGasSchedule.ScheduleHash)
	writeEnginePart(h, boundary.StoreKey)
	writeEngineBool(h, boundary.StoreV2Backed)
	writeEnginePart(h, boundary.KVPrefix)
	writeSyscallMeter(h, boundary.MessageSyscall)
	writeSyscallMeter(h, boundary.ProofVerificationSyscall)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeCosmWasmAdapterBoundaryHash(boundary CosmWasmAdapterBoundary) string {
	boundary = canonicalCosmWasmAdapterBoundary(boundary)
	h := sha256.New()
	writeEnginePart(h, "aetra-cosmwasm-adapter-boundary-v1")
	writeEnginePart(h, boundary.Runtime)
	writeEngineBool(h, boundary.IsolatedAdapterModule)
	writeEnginePart(h, boundary.GasConversion.ConversionHash)
	writeEnginePart(h, boundary.StoreAdapter.AdapterHash)
	writeEngineBool(h, boundary.ExplicitStorageKeyPrefix)
	writeEngineBool(h, boundary.DirectNonContractState)
	writeEngineBool(h, boundary.CrossZoneMessagesOrProofs)
	writeEngineBool(h, boundary.ExternalNetwork)
	writeEngineUint64(h, uint64(len(boundary.HostFunctions)))
	for _, host := range boundary.HostFunctions {
		writeEnginePart(h, host.Name)
		writeEnginePart(h, string(host.Kind))
		writeEngineBool(h, host.Deterministic)
		writeEngineUint64(h, host.GasCost)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMAdapterBoundaryManifestHash(manifest VMAdapterBoundaryManifest) string {
	manifest = canonicalVMAdapterBoundaryManifest(manifest)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-adapter-boundary-manifest-v1")
	writeEnginePart(h, string(manifest.ZoneID))
	writeEnginePart(h, manifest.DeterminismProfile.ProfileHash)
	writeEnginePart(h, manifest.AVM.BoundaryHash)
	writeEnginePart(h, manifest.CosmWasm.BoundaryHash)
	return hex.EncodeToString(h.Sum(nil))
}

func RuntimePolicyForCosmWasmBoundary(boundary CosmWasmAdapterBoundary) RuntimePolicy {
	policy := DefaultRuntimePolicy()
	policy.CosmWasmEnabled = true
	policy.CosmWasmPolicy = wasmconfig.DefaultPolicy()
	policy.CosmWasmPolicy.Enabled = true
	if boundary.GasConversion.AVMGasPerWASMGas != 0 {
		policy.CosmWasmPolicy.GasMultiplier = boundary.GasConversion.AVMGasPerWASMGas
	}
	return policy
}

func canonicalVMDeterminismProfile(profile VMDeterminismProfile) VMDeterminismProfile {
	profile.Runtime = strings.TrimSpace(profile.Runtime)
	profile.ProfileHash = strings.TrimSpace(profile.ProfileHash)
	return profile
}

func canonicalVMSyscallMeter(syscall VMSyscallMeter) VMSyscallMeter {
	syscall.Name = strings.TrimSpace(syscall.Name)
	return syscall
}

func canonicalAVMAdapterBoundary(boundary AVMAdapterBoundary) AVMAdapterBoundary {
	boundary.Runtime = strings.TrimSpace(boundary.Runtime)
	boundary.StoreKey = strings.TrimSpace(boundary.StoreKey)
	boundary.KVPrefix = strings.TrimSpace(boundary.KVPrefix)
	boundary.MessageSyscall = canonicalVMSyscallMeter(boundary.MessageSyscall)
	boundary.ProofVerificationSyscall = canonicalVMSyscallMeter(boundary.ProofVerificationSyscall)
	boundary.BoundaryHash = strings.TrimSpace(boundary.BoundaryHash)
	return boundary
}

func canonicalCosmWasmAdapterBoundary(boundary CosmWasmAdapterBoundary) CosmWasmAdapterBoundary {
	boundary.Runtime = strings.TrimSpace(boundary.Runtime)
	boundary.HostFunctions = append([]AVMWASMHostFunction(nil), boundary.HostFunctions...)
	for i := range boundary.HostFunctions {
		boundary.HostFunctions[i] = canonicalAVMWASMHostFunction(boundary.HostFunctions[i])
	}
	sort.SliceStable(boundary.HostFunctions, func(i, j int) bool {
		return boundary.HostFunctions[i].Name < boundary.HostFunctions[j].Name
	})
	boundary.StoreAdapter = canonicalAVMWASMStoreV2KVAdapter(boundary.StoreAdapter)
	boundary.BoundaryHash = strings.TrimSpace(boundary.BoundaryHash)
	return boundary
}

func canonicalVMAdapterBoundaryManifest(manifest VMAdapterBoundaryManifest) VMAdapterBoundaryManifest {
	manifest.DeterminismProfile = canonicalVMDeterminismProfile(manifest.DeterminismProfile)
	manifest.AVM = canonicalAVMAdapterBoundary(manifest.AVM)
	manifest.CosmWasm = canonicalCosmWasmAdapterBoundary(manifest.CosmWasm)
	manifest.ManifestHash = strings.TrimSpace(manifest.ManifestHash)
	return manifest
}

func writeSyscallMeter(h interface{ Write([]byte) (int, error) }, syscall VMSyscallMeter) {
	syscall = canonicalVMSyscallMeter(syscall)
	writeEnginePart(h, syscall.Name)
	writeEnginePart(h, string(syscall.GasClass))
	writeEngineUint64(h, syscall.GasCost)
	writeEngineBool(h, syscall.Metered)
}
