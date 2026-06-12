package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/wasmconfig"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	MaxVMRuntimeActionCount = 16
)

type VMRuntimeTrait struct {
	Runtime			string
	AdapterKind		VMAdapterKind
	DeterminismProfile	VMDeterminismProfile
	SupportedActions	[]string
	SupportsBytecodeDeploy	bool
	SupportsStorageAdapter	bool
	SupportsOutboundMessage	bool
	EmitsReceipts		bool
	CommitsVMRoot		bool
	TraitHash		string
}

type VMBytecodeValidation struct {
	Runtime		string
	BytecodeHash	string
	CodeBytes	uint64
	Deterministic	bool
	Validated	bool
	ValidationHash	string
}

type VMGasTable struct {
	Runtime		string
	AVMSchedule	AVMGasSchedule
	WASMConversion	AVMWASMGasConversionTable
	TableHash	string
}

type VMStorageAdapter struct {
	Runtime		string
	ZoneID		zonestypes.ZoneID
	StoreKey	string
	KeyPrefix	string
	MaxKeyBytes	uint32
	MaxValueBytes	uint64
	ReadGas		uint64
	WriteGas	uint64
	AdapterHash	string
}

type VMOutboundMessageRequest struct {
	ChainID		string
	Source		string
	Destination	string
	SourceZone	zonestypes.ZoneID
	DestinationZone	zonestypes.ZoneID
	Payload		[]byte
	PayloadType	string
	GasLimit	uint64
	ForwardingFee	uint64
	SenderNonce	uint64
	CreatedHeight	uint64
	ExpiryHeight	uint64
	RouteHint	string
}

type VMOutboundMessageSyscall struct {
	Runtime		string
	Syscall		VMSyscallMeter
	Message		AVMAsyncMessage
	SyscallHash	string
}

type VMReceiptEmission struct {
	Runtime		string
	Receipt		AVMExecutionReceipt
	ReceiptRoot	string
	EmissionHash	string
}

type VMRuntimeAdapter struct {
	Trait			VMRuntimeTrait
	BoundaryManifest	VMAdapterBoundaryManifest
	BytecodeValidation	VMBytecodeValidation
	GasTable		VMGasTable
	StorageAdapter		VMStorageAdapter
	OutboundSyscalls	[]VMOutboundMessageSyscall
	ReceiptEmission		VMReceiptEmission
	AdapterHash		string
}

type VMRuntimeRootCommitment struct {
	Height			uint64
	ZoneID			zonestypes.ZoneID
	Runtime			string
	AdapterHash		string
	BytecodeHash		string
	GasTableHash		string
	StorageAdapterHash	string
	OutboundMessageRoot	string
	ReceiptRoot		string
	VMRootHash		string
}

func NewVMRuntimeTrait(trait VMRuntimeTrait) (VMRuntimeTrait, error) {
	trait = canonicalVMRuntimeTrait(trait)
	if trait.Runtime == "" {
		trait.Runtime = RuntimeAVM
	}
	if trait.AdapterKind == "" {
		trait.AdapterKind = VMAdapterAVM
	}
	if trait.DeterminismProfile.ProfileHash == "" {
		profile, err := DefaultVMDeterminismProfile(trait.Runtime)
		if err != nil {
			return VMRuntimeTrait{}, err
		}
		trait.DeterminismProfile = profile
	}
	if len(trait.SupportedActions) == 0 {
		trait.SupportedActions = []string{ActionDeploy, ActionExternalCall, ActionInternalCall, ActionBouncedCall, ActionQuery}
	}
	trait.SupportsBytecodeDeploy = true
	trait.SupportsStorageAdapter = true
	trait.SupportsOutboundMessage = true
	trait.EmitsReceipts = true
	trait.CommitsVMRoot = true
	trait.TraitHash = ComputeVMRuntimeTraitHash(trait)
	return trait, trait.Validate()
}

func ValidateDeterministicVMBytecode(runtime string, bytecode []byte, policy RuntimePolicy) (VMBytecodeValidation, error) {
	runtime = strings.TrimSpace(runtime)
	if len(bytecode) == 0 {
		return VMBytecodeValidation{}, errors.New("VM bytecode must not be empty")
	}
	if err := ValidateRuntimePolicy(policy); err != nil {
		return VMBytecodeValidation{}, err
	}
	switch runtime {
	case RuntimeAVM:
		module, err := avm.DecodeModule(bytecode)
		if err != nil {
			return VMBytecodeValidation{}, err
		}
		verifier, err := avm.NewVerifier(policy.AVMParams)
		if err != nil {
			return VMBytecodeValidation{}, err
		}
		if err := verifier.Verify(module); err != nil {
			return VMBytecodeValidation{}, err
		}
	case RuntimeCosmWasm:
		cw := policy.CosmWasmPolicy
		cw.Enabled = true
		if err := wasmconfig.ValidateContractCodeSize(uint64(len(bytecode)), false, cw); err != nil {
			return VMBytecodeValidation{}, err
		}
	default:
		return VMBytecodeValidation{}, fmt.Errorf("invalid VM runtime %q", runtime)
	}
	sum := sha256.Sum256(bytecode)
	validation := VMBytecodeValidation{
		Runtime:	runtime,
		BytecodeHash:	hex.EncodeToString(sum[:]),
		CodeBytes:	uint64(len(bytecode)),
		Deterministic:	true,
		Validated:	true,
	}
	validation.ValidationHash = ComputeVMBytecodeValidationHash(validation)
	return validation, validation.Validate()
}

func NewVMGasTable(runtime string, policy RuntimePolicy) (VMGasTable, error) {
	runtime = strings.TrimSpace(runtime)
	if err := ValidateRuntimePolicy(policy); err != nil {
		return VMGasTable{}, err
	}
	table := VMGasTable{Runtime: runtime}
	switch runtime {
	case RuntimeAVM:
		schedule, err := DefaultAVMGasSchedule()
		if err != nil {
			return VMGasTable{}, err
		}
		table.AVMSchedule = schedule
	case RuntimeCosmWasm:
		conversion, err := DefaultAVMWASMGasConversionTable()
		if err != nil {
			return VMGasTable{}, err
		}
		table.WASMConversion = conversion
	default:
		return VMGasTable{}, fmt.Errorf("invalid VM runtime %q", runtime)
	}
	table.TableHash = ComputeVMGasTableHash(table)
	return table, table.Validate()
}

func NewVMStorageAdapter(runtime string, zoneID zonestypes.ZoneID, manifest VMAdapterBoundaryManifest) (VMStorageAdapter, error) {
	manifest = canonicalVMAdapterBoundaryManifest(manifest)
	if err := manifest.Validate(); err != nil {
		return VMStorageAdapter{}, err
	}
	adapter := VMStorageAdapter{Runtime: strings.TrimSpace(runtime), ZoneID: zoneID}
	switch adapter.Runtime {
	case RuntimeAVM:
		adapter.StoreKey = manifest.AVM.StoreKey
		adapter.KeyPrefix = manifest.AVM.KVPrefix
		adapter.MaxKeyBytes = DefaultMaxStorageKeyBytes
		adapter.MaxValueBytes = DefaultMaxStorageValueBytes
		adapter.ReadGas = manifest.AVM.DeterministicGasSchedule.ClassBudget(AVMGasClassStorage)
		adapter.WriteGas = adapter.ReadGas
	case RuntimeCosmWasm:
		adapter.StoreKey = manifest.CosmWasm.StoreAdapter.StoreKey
		adapter.KeyPrefix = manifest.CosmWasm.StoreAdapter.KeyPrefix
		adapter.MaxKeyBytes = manifest.CosmWasm.StoreAdapter.MaxKeyBytes
		adapter.MaxValueBytes = manifest.CosmWasm.StoreAdapter.MaxValueBytes
		adapter.ReadGas = manifest.CosmWasm.GasConversion.StorageReadGas
		adapter.WriteGas = manifest.CosmWasm.GasConversion.StorageWriteGas
	default:
		return VMStorageAdapter{}, fmt.Errorf("invalid VM runtime %q", adapter.Runtime)
	}
	adapter.AdapterHash = ComputeVMStorageAdapterHash(adapter)
	return adapter, adapter.Validate()
}

func NewVMOutboundMessageSyscall(runtime string, meter VMSyscallMeter, req VMOutboundMessageRequest) (VMOutboundMessageSyscall, error) {
	runtime = strings.TrimSpace(runtime)
	meter = canonicalVMSyscallMeter(meter)
	if err := meter.Validate(); err != nil {
		return VMOutboundMessageSyscall{}, err
	}
	if meter.GasClass != AVMGasClassCrossZoneRouting {
		return VMOutboundMessageSyscall{}, errors.New("VM outbound message syscall must use cross-zone routing gas")
	}
	msg, err := NewAVMAsyncMessage(AVMAsyncMessage{
		ChainID:		req.ChainID,
		Source:			req.Source,
		Destination:		req.Destination,
		Payload:		append([]byte(nil), req.Payload...),
		GasLimit:		req.GasLimit,
		ExpiryHeight:		req.ExpiryHeight,
		RetryPolicy:		DefaultAVMRetryPolicy(req.ExpiryHeight),
		BounceFlag:		true,
		SourceZone:		req.SourceZone,
		DestinationZone:	req.DestinationZone,
		SenderNonce:		req.SenderNonce,
		PayloadType:		req.PayloadType,
		ForwardingFee:		req.ForwardingFee,
		CreatedHeight:		req.CreatedHeight,
		RouteHintOptional:	req.RouteHint,
	})
	if err != nil {
		return VMOutboundMessageSyscall{}, err
	}
	syscall := VMOutboundMessageSyscall{Runtime: runtime, Syscall: meter, Message: msg}
	syscall.SyscallHash = ComputeVMOutboundMessageSyscallHash(syscall)
	return syscall, syscall.Validate()
}

func NewVMReceiptEmission(runtime string, receipt AVMExecutionReceipt, zoneID zonestypes.ZoneID) (VMReceiptEmission, error) {
	receipt, err := NewAVMExecutionReceipt(receipt)
	if err != nil {
		return VMReceiptEmission{}, err
	}
	if receipt.ZoneID != zoneID {
		return VMReceiptEmission{}, errors.New("VM receipt emission zone mismatch")
	}
	emission := VMReceiptEmission{
		Runtime:	strings.TrimSpace(runtime),
		Receipt:	receipt,
		ReceiptRoot:	ComputeAVMContractReceiptRoot([]AVMExecutionReceipt{receipt}),
	}
	emission.EmissionHash = ComputeVMReceiptEmissionHash(emission)
	return emission, emission.Validate()
}

func NewVMRuntimeAdapter(adapter VMRuntimeAdapter) (VMRuntimeAdapter, error) {
	adapter = canonicalVMRuntimeAdapter(adapter)
	if err := adapter.Trait.Validate(); err != nil {
		return VMRuntimeAdapter{}, err
	}
	if err := adapter.BoundaryManifest.Validate(); err != nil {
		return VMRuntimeAdapter{}, err
	}
	if adapter.Trait.Runtime != adapter.BytecodeValidation.Runtime ||
		adapter.Trait.Runtime != adapter.GasTable.Runtime ||
		adapter.Trait.Runtime != adapter.StorageAdapter.Runtime ||
		adapter.Trait.Runtime != adapter.ReceiptEmission.Runtime {
		return VMRuntimeAdapter{}, errors.New("VM runtime adapter component runtime mismatch")
	}
	if err := adapter.BytecodeValidation.Validate(); err != nil {
		return VMRuntimeAdapter{}, err
	}
	if err := adapter.GasTable.Validate(); err != nil {
		return VMRuntimeAdapter{}, err
	}
	if err := adapter.StorageAdapter.Validate(); err != nil {
		return VMRuntimeAdapter{}, err
	}
	for _, syscall := range adapter.OutboundSyscalls {
		if err := syscall.Validate(); err != nil {
			return VMRuntimeAdapter{}, err
		}
		if syscall.Runtime != adapter.Trait.Runtime {
			return VMRuntimeAdapter{}, errors.New("VM runtime adapter outbound syscall runtime mismatch")
		}
	}
	if err := adapter.ReceiptEmission.Validate(); err != nil {
		return VMRuntimeAdapter{}, err
	}
	adapter.AdapterHash = ComputeVMRuntimeAdapterHash(adapter)
	return adapter, adapter.Validate()
}

func NewVMRuntimeRootCommitment(height uint64, zoneID zonestypes.ZoneID, adapter VMRuntimeAdapter) (VMRuntimeRootCommitment, error) {
	adapter = canonicalVMRuntimeAdapter(adapter)
	if err := adapter.Validate(); err != nil {
		return VMRuntimeRootCommitment{}, err
	}
	outbound := make([]AVMAsyncMessage, 0, len(adapter.OutboundSyscalls))
	for _, syscall := range adapter.OutboundSyscalls {
		outbound = append(outbound, syscall.Message)
	}
	root := VMRuntimeRootCommitment{
		Height:			height,
		ZoneID:			zoneID,
		Runtime:		adapter.Trait.Runtime,
		AdapterHash:		adapter.AdapterHash,
		BytecodeHash:		adapter.BytecodeValidation.BytecodeHash,
		GasTableHash:		adapter.GasTable.TableHash,
		StorageAdapterHash:	adapter.StorageAdapter.AdapterHash,
		OutboundMessageRoot:	ComputeAVMZoneOutputMessageRoot(zoneID, outbound),
		ReceiptRoot:		adapter.ReceiptEmission.ReceiptRoot,
	}
	root.VMRootHash = ComputeVMRuntimeRootCommitmentHash(root)
	return root, root.Validate()
}

func (t VMRuntimeTrait) Validate() error {
	t = canonicalVMRuntimeTrait(t)
	if t.Runtime != RuntimeAVM && t.Runtime != RuntimeCosmWasm {
		return fmt.Errorf("VM runtime trait runtime %q is not supported", t.Runtime)
	}
	switch t.Runtime {
	case RuntimeAVM:
		if t.AdapterKind != VMAdapterAVM {
			return errors.New("AVM runtime trait must use AVM adapter")
		}
	case RuntimeCosmWasm:
		if t.AdapterKind != VMAdapterCosmWasm {
			return errors.New("CosmWasm runtime trait must use CosmWasm adapter")
		}
	}
	if err := t.DeterminismProfile.Validate(); err != nil {
		return err
	}
	if t.DeterminismProfile.Runtime != t.Runtime {
		return errors.New("VM runtime trait determinism profile runtime mismatch")
	}
	if len(t.SupportedActions) == 0 || len(t.SupportedActions) > MaxVMRuntimeActionCount {
		return fmt.Errorf("VM runtime trait actions must be between 1 and %d", MaxVMRuntimeActionCount)
	}
	for i, action := range t.SupportedActions {
		if !IsVMAction(action) {
			return fmt.Errorf("VM runtime trait action %q is invalid", action)
		}
		if i > 0 && t.SupportedActions[i-1] >= action {
			return errors.New("VM runtime trait actions must be sorted and unique")
		}
	}
	if !t.SupportsBytecodeDeploy || !t.SupportsStorageAdapter || !t.SupportsOutboundMessage || !t.EmitsReceipts || !t.CommitsVMRoot {
		return errors.New("VM runtime trait must expose bytecode, storage, message, receipt, and root capabilities")
	}
	if t.TraitHash == "" {
		return errors.New("VM runtime trait hash is required")
	}
	if err := zonestypes.ValidateHash("VM runtime trait hash", t.TraitHash); err != nil {
		return err
	}
	if t.TraitHash != ComputeVMRuntimeTraitHash(t) {
		return errors.New("VM runtime trait hash mismatch")
	}
	return nil
}

func (v VMBytecodeValidation) Validate() error {
	v = canonicalVMBytecodeValidation(v)
	if v.Runtime != RuntimeAVM && v.Runtime != RuntimeCosmWasm {
		return fmt.Errorf("VM bytecode validation runtime %q is not supported", v.Runtime)
	}
	if err := zonestypes.ValidateHash("VM bytecode hash", v.BytecodeHash); err != nil {
		return err
	}
	if v.CodeBytes == 0 {
		return errors.New("VM bytecode validation code bytes must be positive")
	}
	if !v.Deterministic || !v.Validated {
		return errors.New("VM bytecode must be deterministic and validated")
	}
	if v.ValidationHash == "" {
		return errors.New("VM bytecode validation hash is required")
	}
	if err := zonestypes.ValidateHash("VM bytecode validation hash", v.ValidationHash); err != nil {
		return err
	}
	if v.ValidationHash != ComputeVMBytecodeValidationHash(v) {
		return errors.New("VM bytecode validation hash mismatch")
	}
	return nil
}

func (t VMGasTable) Validate() error {
	t = canonicalVMGasTable(t)
	switch t.Runtime {
	case RuntimeAVM:
		if err := t.AVMSchedule.Validate(); err != nil {
			return err
		}
	case RuntimeCosmWasm:
		if err := t.WASMConversion.Validate(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("VM gas table runtime %q is not supported", t.Runtime)
	}
	if t.TableHash == "" {
		return errors.New("VM gas table hash is required")
	}
	if err := zonestypes.ValidateHash("VM gas table hash", t.TableHash); err != nil {
		return err
	}
	if t.TableHash != ComputeVMGasTableHash(t) {
		return errors.New("VM gas table hash mismatch")
	}
	return nil
}

func (a VMStorageAdapter) Validate() error {
	a = canonicalVMStorageAdapter(a)
	if a.Runtime != RuntimeAVM && a.Runtime != RuntimeCosmWasm {
		return fmt.Errorf("VM storage adapter runtime %q is not supported", a.Runtime)
	}
	if err := zonestypes.ValidateZoneID(a.ZoneID); err != nil {
		return err
	}
	if a.StoreKey != DefaultAVMStoreKey {
		return fmt.Errorf("VM storage adapter store key must be %q", DefaultAVMStoreKey)
	}
	if a.KeyPrefix != ContractZoneKVPrefix(a.ZoneID) {
		return fmt.Errorf("VM storage adapter prefix must be %q", ContractZoneKVPrefix(a.ZoneID))
	}
	if a.MaxKeyBytes == 0 || a.MaxValueBytes == 0 {
		return errors.New("VM storage adapter limits must be positive")
	}
	if a.ReadGas == 0 || a.WriteGas == 0 {
		return errors.New("VM storage adapter access must be metered")
	}
	if a.AdapterHash == "" {
		return errors.New("VM storage adapter hash is required")
	}
	if err := zonestypes.ValidateHash("VM storage adapter hash", a.AdapterHash); err != nil {
		return err
	}
	if a.AdapterHash != ComputeVMStorageAdapterHash(a) {
		return errors.New("VM storage adapter hash mismatch")
	}
	return nil
}

func (s VMOutboundMessageSyscall) Validate() error {
	s = canonicalVMOutboundMessageSyscall(s)
	if s.Runtime != RuntimeAVM && s.Runtime != RuntimeCosmWasm {
		return fmt.Errorf("VM outbound syscall runtime %q is not supported", s.Runtime)
	}
	if err := s.Syscall.Validate(); err != nil {
		return err
	}
	if s.Syscall.GasClass != AVMGasClassCrossZoneRouting {
		return errors.New("VM outbound syscall must use cross-zone routing gas")
	}
	if err := s.Message.Validate(); err != nil {
		return err
	}
	if s.SyscallHash == "" {
		return errors.New("VM outbound syscall hash is required")
	}
	if err := zonestypes.ValidateHash("VM outbound syscall hash", s.SyscallHash); err != nil {
		return err
	}
	if s.SyscallHash != ComputeVMOutboundMessageSyscallHash(s) {
		return errors.New("VM outbound syscall hash mismatch")
	}
	return nil
}

func (e VMReceiptEmission) Validate() error {
	e = canonicalVMReceiptEmission(e)
	if e.Runtime != RuntimeAVM && e.Runtime != RuntimeCosmWasm {
		return fmt.Errorf("VM receipt emission runtime %q is not supported", e.Runtime)
	}
	if err := e.Receipt.Validate(); err != nil {
		return err
	}
	if e.ReceiptRoot != ComputeAVMContractReceiptRoot([]AVMExecutionReceipt{e.Receipt}) {
		return errors.New("VM receipt emission root mismatch")
	}
	if e.EmissionHash == "" {
		return errors.New("VM receipt emission hash is required")
	}
	if err := zonestypes.ValidateHash("VM receipt emission hash", e.EmissionHash); err != nil {
		return err
	}
	if e.EmissionHash != ComputeVMReceiptEmissionHash(e) {
		return errors.New("VM receipt emission hash mismatch")
	}
	return nil
}

func (a VMRuntimeAdapter) Validate() error {
	a = canonicalVMRuntimeAdapter(a)
	if err := a.Trait.Validate(); err != nil {
		return err
	}
	if err := a.BoundaryManifest.Validate(); err != nil {
		return err
	}
	if err := a.BytecodeValidation.Validate(); err != nil {
		return err
	}
	if err := a.GasTable.Validate(); err != nil {
		return err
	}
	if err := a.StorageAdapter.Validate(); err != nil {
		return err
	}
	for _, syscall := range a.OutboundSyscalls {
		if err := syscall.Validate(); err != nil {
			return err
		}
	}
	if err := a.ReceiptEmission.Validate(); err != nil {
		return err
	}
	if a.AdapterHash == "" {
		return errors.New("VM runtime adapter hash is required")
	}
	if err := zonestypes.ValidateHash("VM runtime adapter hash", a.AdapterHash); err != nil {
		return err
	}
	if a.AdapterHash != ComputeVMRuntimeAdapterHash(a) {
		return errors.New("VM runtime adapter hash mismatch")
	}
	return nil
}

func (r VMRuntimeRootCommitment) Validate() error {
	r = canonicalVMRuntimeRootCommitment(r)
	if r.Height == 0 {
		return errors.New("VM runtime root height must be positive")
	}
	if err := zonestypes.ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.Runtime != RuntimeAVM && r.Runtime != RuntimeCosmWasm {
		return fmt.Errorf("VM runtime root runtime %q is not supported", r.Runtime)
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"VM runtime root adapter hash", r.AdapterHash},
		{"VM runtime root bytecode hash", r.BytecodeHash},
		{"VM runtime root gas table hash", r.GasTableHash},
		{"VM runtime root storage adapter hash", r.StorageAdapterHash},
		{"VM runtime outbound message root", r.OutboundMessageRoot},
		{"VM runtime receipt root", r.ReceiptRoot},
		{"VM runtime root hash", r.VMRootHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.VMRootHash != ComputeVMRuntimeRootCommitmentHash(r) {
		return errors.New("VM runtime root hash mismatch")
	}
	return nil
}

func (s AVMGasSchedule) ClassBudget(class AVMGasClass) uint64 {
	s = canonicalAVMGasSchedule(s)
	for _, budget := range s.ClassBudgets {
		if budget.Class == class {
			return budget.Limit
		}
	}
	return 0
}

func ComputeVMRuntimeTraitHash(trait VMRuntimeTrait) string {
	trait = canonicalVMRuntimeTrait(trait)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-runtime-trait-v1")
	writeEnginePart(h, trait.Runtime)
	writeEnginePart(h, string(trait.AdapterKind))
	writeEnginePart(h, trait.DeterminismProfile.ProfileHash)
	for _, action := range trait.SupportedActions {
		writeEnginePart(h, action)
	}
	writeEngineBool(h, trait.SupportsBytecodeDeploy)
	writeEngineBool(h, trait.SupportsStorageAdapter)
	writeEngineBool(h, trait.SupportsOutboundMessage)
	writeEngineBool(h, trait.EmitsReceipts)
	writeEngineBool(h, trait.CommitsVMRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMBytecodeValidationHash(validation VMBytecodeValidation) string {
	validation = canonicalVMBytecodeValidation(validation)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-bytecode-validation-v1")
	writeEnginePart(h, validation.Runtime)
	writeEnginePart(h, validation.BytecodeHash)
	writeEngineUint64(h, validation.CodeBytes)
	writeEngineBool(h, validation.Deterministic)
	writeEngineBool(h, validation.Validated)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMGasTableHash(table VMGasTable) string {
	table = canonicalVMGasTable(table)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-gas-table-v1")
	writeEnginePart(h, table.Runtime)
	writeEnginePart(h, table.AVMSchedule.ScheduleHash)
	writeEnginePart(h, table.WASMConversion.ConversionHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMStorageAdapterHash(adapter VMStorageAdapter) string {
	adapter = canonicalVMStorageAdapter(adapter)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-storage-adapter-v1")
	writeEnginePart(h, adapter.Runtime)
	writeEnginePart(h, string(adapter.ZoneID))
	writeEnginePart(h, adapter.StoreKey)
	writeEnginePart(h, adapter.KeyPrefix)
	writeEngineUint64(h, uint64(adapter.MaxKeyBytes))
	writeEngineUint64(h, adapter.MaxValueBytes)
	writeEngineUint64(h, adapter.ReadGas)
	writeEngineUint64(h, adapter.WriteGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMOutboundMessageSyscallHash(syscall VMOutboundMessageSyscall) string {
	syscall = canonicalVMOutboundMessageSyscall(syscall)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-outbound-message-syscall-v1")
	writeEnginePart(h, syscall.Runtime)
	writeSyscallMeter(h, syscall.Syscall)
	writeEnginePart(h, syscall.Message.ID)
	writeEnginePart(h, string(syscall.Message.SourceZone))
	writeEnginePart(h, string(syscall.Message.DestinationZone))
	writeEnginePart(h, syscall.Message.PayloadHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMReceiptEmissionHash(emission VMReceiptEmission) string {
	emission = canonicalVMReceiptEmission(emission)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-receipt-emission-v1")
	writeEnginePart(h, emission.Runtime)
	writeEnginePart(h, emission.Receipt.ReceiptHash)
	writeEnginePart(h, emission.ReceiptRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMRuntimeAdapterHash(adapter VMRuntimeAdapter) string {
	adapter = canonicalVMRuntimeAdapter(adapter)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-runtime-adapter-v1")
	writeEnginePart(h, adapter.Trait.TraitHash)
	writeEnginePart(h, adapter.BoundaryManifest.ManifestHash)
	writeEnginePart(h, adapter.BytecodeValidation.ValidationHash)
	writeEnginePart(h, adapter.GasTable.TableHash)
	writeEnginePart(h, adapter.StorageAdapter.AdapterHash)
	for _, syscall := range adapter.OutboundSyscalls {
		writeEnginePart(h, syscall.SyscallHash)
	}
	writeEnginePart(h, adapter.ReceiptEmission.EmissionHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVMRuntimeRootCommitmentHash(root VMRuntimeRootCommitment) string {
	root = canonicalVMRuntimeRootCommitment(root)
	h := sha256.New()
	writeEnginePart(h, "aetra-vm-runtime-root-commitment-v1")
	writeEngineUint64(h, root.Height)
	writeEnginePart(h, string(root.ZoneID))
	writeEnginePart(h, root.Runtime)
	writeEnginePart(h, root.AdapterHash)
	writeEnginePart(h, root.BytecodeHash)
	writeEnginePart(h, root.GasTableHash)
	writeEnginePart(h, root.StorageAdapterHash)
	writeEnginePart(h, root.OutboundMessageRoot)
	writeEnginePart(h, root.ReceiptRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalVMRuntimeTrait(trait VMRuntimeTrait) VMRuntimeTrait {
	trait.Runtime = strings.TrimSpace(trait.Runtime)
	trait.DeterminismProfile = canonicalVMDeterminismProfile(trait.DeterminismProfile)
	trait.SupportedActions = append([]string(nil), trait.SupportedActions...)
	for i := range trait.SupportedActions {
		trait.SupportedActions[i] = strings.TrimSpace(trait.SupportedActions[i])
	}
	sort.Strings(trait.SupportedActions)
	trait.TraitHash = strings.TrimSpace(trait.TraitHash)
	return trait
}

func canonicalVMBytecodeValidation(validation VMBytecodeValidation) VMBytecodeValidation {
	validation.Runtime = strings.TrimSpace(validation.Runtime)
	validation.BytecodeHash = strings.TrimSpace(validation.BytecodeHash)
	validation.ValidationHash = strings.TrimSpace(validation.ValidationHash)
	return validation
}

func canonicalVMGasTable(table VMGasTable) VMGasTable {
	table.Runtime = strings.TrimSpace(table.Runtime)
	table.AVMSchedule = canonicalAVMGasSchedule(table.AVMSchedule)
	table.TableHash = strings.TrimSpace(table.TableHash)
	return table
}

func canonicalVMStorageAdapter(adapter VMStorageAdapter) VMStorageAdapter {
	adapter.Runtime = strings.TrimSpace(adapter.Runtime)
	adapter.StoreKey = strings.TrimSpace(adapter.StoreKey)
	adapter.KeyPrefix = strings.TrimSpace(adapter.KeyPrefix)
	adapter.AdapterHash = strings.TrimSpace(adapter.AdapterHash)
	return adapter
}

func canonicalVMOutboundMessageSyscall(syscall VMOutboundMessageSyscall) VMOutboundMessageSyscall {
	syscall.Runtime = strings.TrimSpace(syscall.Runtime)
	syscall.Syscall = canonicalVMSyscallMeter(syscall.Syscall)
	syscall.Message = canonicalAVMAsyncMessage(syscall.Message)
	syscall.SyscallHash = strings.TrimSpace(syscall.SyscallHash)
	return syscall
}

func canonicalVMReceiptEmission(emission VMReceiptEmission) VMReceiptEmission {
	emission.Runtime = strings.TrimSpace(emission.Runtime)
	emission.Receipt = canonicalAVMExecutionReceipt(emission.Receipt)
	emission.ReceiptRoot = strings.TrimSpace(emission.ReceiptRoot)
	emission.EmissionHash = strings.TrimSpace(emission.EmissionHash)
	return emission
}

func canonicalVMRuntimeAdapter(adapter VMRuntimeAdapter) VMRuntimeAdapter {
	adapter.Trait = canonicalVMRuntimeTrait(adapter.Trait)
	adapter.BoundaryManifest = canonicalVMAdapterBoundaryManifest(adapter.BoundaryManifest)
	adapter.BytecodeValidation = canonicalVMBytecodeValidation(adapter.BytecodeValidation)
	adapter.GasTable = canonicalVMGasTable(adapter.GasTable)
	adapter.StorageAdapter = canonicalVMStorageAdapter(adapter.StorageAdapter)
	adapter.OutboundSyscalls = append([]VMOutboundMessageSyscall(nil), adapter.OutboundSyscalls...)
	for i := range adapter.OutboundSyscalls {
		adapter.OutboundSyscalls[i] = canonicalVMOutboundMessageSyscall(adapter.OutboundSyscalls[i])
	}
	sort.SliceStable(adapter.OutboundSyscalls, func(i, j int) bool {
		return adapter.OutboundSyscalls[i].SyscallHash < adapter.OutboundSyscalls[j].SyscallHash
	})
	adapter.ReceiptEmission = canonicalVMReceiptEmission(adapter.ReceiptEmission)
	adapter.AdapterHash = strings.TrimSpace(adapter.AdapterHash)
	return adapter
}

func canonicalVMRuntimeRootCommitment(root VMRuntimeRootCommitment) VMRuntimeRootCommitment {
	root.Runtime = strings.TrimSpace(root.Runtime)
	root.AdapterHash = strings.TrimSpace(root.AdapterHash)
	root.BytecodeHash = strings.TrimSpace(root.BytecodeHash)
	root.GasTableHash = strings.TrimSpace(root.GasTableHash)
	root.StorageAdapterHash = strings.TrimSpace(root.StorageAdapterHash)
	root.OutboundMessageRoot = strings.TrimSpace(root.OutboundMessageRoot)
	root.ReceiptRoot = strings.TrimSpace(root.ReceiptRoot)
	root.VMRootHash = strings.TrimSpace(root.VMRootHash)
	return root
}
