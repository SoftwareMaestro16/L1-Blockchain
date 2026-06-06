package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVM2ContractStatePrefixCode         = "contract/code"
	AVM2ContractStatePrefixInstance     = "contract/instance"
	AVM2ContractStatePrefixStorage      = "contract/storage"
	AVM2ContractStatePrefixABI          = "contract/abi"
	AVM2ContractStatePrefixEvents       = "contract/events"
	AVM2ContractStatePrefixMessageNonce = "contract/message_nonce"

	AVM2ContractValueCodeRecord     = "CodeRecord"
	AVM2ContractValueContractRecord = "ContractRecord"
	AVM2ContractValueStorageValue   = "StorageValue"
	AVM2ContractValueABIDescriptor  = "AbiDescriptor"
	AVM2ContractValueContractEvent  = "ContractEvent"
	AVM2ContractValueMessageNonce   = "uint64"

	AVM2MeteringProfileDefault = "default-avm2"
)

type AVM2ContractStateEntry struct {
	Key       string
	ValueType string
	Purpose   string
	ShardKey  string
	EntryHash string
}

type AVM2ContractStateLayout struct {
	Entries    []AVM2ContractStateEntry
	LayoutRoot string
}

type AVM2CodeRecord struct {
	CodeID                uint64
	CodeHash              string
	VMVersion             uint64
	InstructionSetVersion uint64
	ABIHash               string
	Deployer              string
	CreatedAtHeight       uint64
	CodeBytesRef          string
	MeteringProfile       string
	Enabled               bool
	RecordHash            string
}

type AVM2ContractRecord struct {
	ContractAddr    string
	CodeID          uint64
	Creator         string
	AdminOptional   string
	StorageRoot     string
	BalanceNAET     sdkmath.Int
	CreatedAtHeight uint64
	UpdatedAtHeight uint64
	InstanceVersion uint64
	ShardID         uint32
	RecordHash      string
}

type AVM2ContractStorageValue struct {
	ContractAddr string
	StorageKey   string
	ValueHash    string
	ValueBytes   uint64
	ShardKey     string
	RecordHash   string
}

type AVM2ContractABIDescriptorRecord struct {
	CodeID     uint64
	ABI        AVM2ABIDescriptor
	Key        string
	ShardKey   string
	RecordHash string
}

type AVM2ContractEventRecord struct {
	Event      AVM2Event
	Key        string
	ShardKey   string
	RecordHash string
}

type AVM2ContractMessageNonceRecord struct {
	ContractAddr string
	Nonce        uint64
	Key          string
	ShardKey     string
	RecordHash   string
}

type AVM2ContractStateIndex struct {
	Layout        AVM2ContractStateLayout
	Codes         []AVM2CodeRecord
	Contracts     []AVM2ContractRecord
	Storage       []AVM2ContractStorageValue
	ABIs          []AVM2ContractABIDescriptorRecord
	Events        []AVM2ContractEventRecord
	MessageNonces []AVM2ContractMessageNonceRecord
	StateRoot     string
}

func DefaultAVM2ContractStateLayout() (AVM2ContractStateLayout, error) {
	layout := AVM2ContractStateLayout{Entries: []AVM2ContractStateEntry{
		{
			Key:       AVM2ContractCodeStateKey(1),
			ValueType: AVM2ContractValueCodeRecord,
			Purpose:   "code metadata, hashes, VM version, metering profile, and enablement",
			ShardKey:  "code_id",
		},
		{
			Key:       AVM2ContractInstanceStateKey("contract_addr"),
			ValueType: AVM2ContractValueContractRecord,
			Purpose:   "contract instance metadata, code binding, admin, balance, storage root, and shard assignment",
			ShardKey:  "contract_addr",
		},
		{
			Key:       AVM2ContractStorageStateKey("contract_addr", "storage_key"),
			ValueType: AVM2ContractValueStorageValue,
			Purpose:   "contract-owned persistent key-value state",
			ShardKey:  "contract_addr/storage_key_prefix",
		},
		{
			Key:       AVM2ContractABIStateKey(1, 1),
			ValueType: AVM2ContractValueABIDescriptor,
			Purpose:   "versioned ABI and schema metadata for calls, events, and errors",
			ShardKey:  "code_id",
		},
		{
			Key:       AVM2ContractEventStateKey(1, "contract_addr", "event_id"),
			ValueType: AVM2ContractValueContractEvent,
			Purpose:   "deterministic contract event output included in event roots",
			ShardKey:  "contract_addr",
		},
		{
			Key:       AVM2ContractMessageNonceStateKey("contract_addr"),
			ValueType: AVM2ContractValueMessageNonce,
			Purpose:   "replay-safe nonce for contract-emitted messages",
			ShardKey:  "contract_addr",
		},
	}}
	layout = canonicalAVM2ContractStateLayout(layout)
	for i := range layout.Entries {
		layout.Entries[i].EntryHash = ComputeAVM2ContractStateEntryHash(layout.Entries[i])
	}
	layout = canonicalAVM2ContractStateLayout(layout)
	layout.LayoutRoot = ComputeAVM2ContractStateLayoutRoot(layout)
	return layout, layout.Validate()
}

func NewAVM2CodeRecord(record AVM2CodeRecord) (AVM2CodeRecord, error) {
	record = canonicalAVM2CodeRecord(record)
	record.RecordHash = ComputeAVM2CodeRecordHash(record)
	return record, record.Validate()
}

func NewAVM2ContractRecord(record AVM2ContractRecord) (AVM2ContractRecord, error) {
	record = canonicalAVM2ContractRecord(record)
	record.RecordHash = ComputeAVM2ContractRecordHash(record)
	return record, record.Validate()
}

func NewAVM2ContractStorageValue(value AVM2ContractStorageValue) (AVM2ContractStorageValue, error) {
	value = canonicalAVM2ContractStorageValue(value)
	if value.ShardKey == "" {
		value.ShardKey = AVM2ContractStorageShardKey(value.ContractAddr, value.StorageKey)
	}
	value.RecordHash = ComputeAVM2ContractStorageValueHash(value)
	return value, value.Validate()
}

func NewAVM2ContractABIDescriptorRecord(record AVM2ContractABIDescriptorRecord) (AVM2ContractABIDescriptorRecord, error) {
	record = canonicalAVM2ContractABIDescriptorRecord(record)
	if record.CodeID == 0 {
		record.CodeID = record.ABI.CodeID
	}
	if record.Key == "" {
		record.Key = AVM2ContractABIStateKey(record.CodeID, record.ABI.ABIVersion)
	}
	if record.ShardKey == "" {
		record.ShardKey = AVM2ContractCodeShardKey(record.CodeID)
	}
	record.RecordHash = ComputeAVM2ContractABIDescriptorRecordHash(record)
	return record, record.Validate()
}

func NewAVM2ContractEventRecord(record AVM2ContractEventRecord) (AVM2ContractEventRecord, error) {
	record = canonicalAVM2ContractEventRecord(record)
	if record.Key == "" {
		record.Key = AVM2ContractEventStateKey(record.Event.Height, record.Event.ContractAddress, record.Event.EventID)
	}
	if record.ShardKey == "" {
		record.ShardKey = record.Event.ContractAddress
	}
	record.RecordHash = ComputeAVM2ContractEventRecordHash(record)
	return record, record.Validate()
}

func NewAVM2ContractMessageNonceRecord(record AVM2ContractMessageNonceRecord) (AVM2ContractMessageNonceRecord, error) {
	record = canonicalAVM2ContractMessageNonceRecord(record)
	if record.Key == "" {
		record.Key = AVM2ContractMessageNonceStateKey(record.ContractAddr)
	}
	if record.ShardKey == "" {
		record.ShardKey = record.ContractAddr
	}
	record.RecordHash = ComputeAVM2ContractMessageNonceRecordHash(record)
	return record, record.Validate()
}

func NewAVM2ContractStateIndex(index AVM2ContractStateIndex) (AVM2ContractStateIndex, error) {
	index = canonicalAVM2ContractStateIndex(index)
	if len(index.Layout.Entries) == 0 {
		layout, err := DefaultAVM2ContractStateLayout()
		if err != nil {
			return AVM2ContractStateIndex{}, err
		}
		index.Layout = layout
	}
	index.StateRoot = ComputeAVM2ContractStateIndexRoot(index)
	return index, index.Validate()
}

func ValidateAVM2ContractInstantiation(code AVM2CodeRecord, contract AVM2ContractRecord) error {
	code = canonicalAVM2CodeRecord(code)
	contract = canonicalAVM2ContractRecord(contract)
	if err := code.Validate(); err != nil {
		return err
	}
	if err := contract.Validate(); err != nil {
		return err
	}
	if !code.Enabled {
		return errors.New("AVM 2.0 disabled code cannot instantiate contract")
	}
	if contract.CodeID != code.CodeID {
		return errors.New("AVM 2.0 contract code id does not match code record")
	}
	if code.VMVersion != AVM2VMVersion {
		return errors.New("AVM 2.0 code record has incompatible VM version")
	}
	return nil
}

func (l AVM2ContractStateLayout) Validate() error {
	l = canonicalAVM2ContractStateLayout(l)
	if len(l.Entries) != 6 {
		return errors.New("AVM 2.0 contract state layout must declare six canonical entries")
	}
	seen := make(map[string]struct{}, len(l.Entries))
	for i, entry := range l.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seen[entry.ValueType]; found {
			return fmt.Errorf("duplicate AVM 2.0 contract state value type %q", entry.ValueType)
		}
		seen[entry.ValueType] = struct{}{}
		if i > 0 && l.Entries[i-1].Key >= entry.Key {
			return errors.New("AVM 2.0 contract state layout entries must be sorted canonically")
		}
	}
	for _, required := range []string{
		AVM2ContractValueCodeRecord,
		AVM2ContractValueContractRecord,
		AVM2ContractValueStorageValue,
		AVM2ContractValueABIDescriptor,
		AVM2ContractValueContractEvent,
		AVM2ContractValueMessageNonce,
	} {
		if _, found := seen[required]; !found {
			return fmt.Errorf("AVM 2.0 contract state layout missing %s", required)
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract state layout root", l.LayoutRoot); err != nil {
		return err
	}
	if l.LayoutRoot != ComputeAVM2ContractStateLayoutRoot(l) {
		return errors.New("AVM 2.0 contract state layout root mismatch")
	}
	return nil
}

func (e AVM2ContractStateEntry) Validate() error {
	e = canonicalAVM2ContractStateEntry(e)
	if err := validateAVM2ContractStateKey("AVM 2.0 contract state key", e.Key); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 contract state value type", e.ValueType, MaxAVM2TokenLength); err != nil {
		return err
	}
	if strings.TrimSpace(e.Purpose) == "" {
		return errors.New("AVM 2.0 contract state purpose is required")
	}
	if len(e.Purpose) > 256 {
		return errors.New("AVM 2.0 contract state purpose must be <= 256 bytes")
	}
	if err := validateEngineToken("AVM 2.0 contract state shard key", strings.ReplaceAll(e.ShardKey, "_prefix", ""), MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract state entry hash", e.EntryHash); err != nil {
		return err
	}
	if e.EntryHash != ComputeAVM2ContractStateEntryHash(e) {
		return errors.New("AVM 2.0 contract state entry hash mismatch")
	}
	return nil
}

func (r AVM2CodeRecord) Validate() error {
	r = canonicalAVM2CodeRecord(r)
	if r.CodeID == 0 {
		return errors.New("AVM 2.0 code id must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 code hash", r.CodeHash); err != nil {
		return err
	}
	if r.VMVersion != AVM2VMVersion {
		return errors.New("AVM 2.0 code record VM version must be 2")
	}
	if r.InstructionSetVersion == 0 {
		return errors.New("AVM 2.0 code instruction set version must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 code ABI hash", r.ABIHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 code deployer", r.Deployer, MaxAVM2TokenLength); err != nil {
		return err
	}
	if r.CreatedAtHeight == 0 {
		return errors.New("AVM 2.0 code created height must be positive")
	}
	if err := validateAVM2ContentRef("AVM 2.0 code bytes ref", r.CodeBytesRef); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 metering profile", r.MeteringProfile, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 code record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVM2CodeRecordHash(r) {
		return errors.New("AVM 2.0 code record hash mismatch")
	}
	return nil
}

func (r AVM2ContractRecord) Validate() error {
	r = canonicalAVM2ContractRecord(r)
	if err := validateEngineToken("AVM 2.0 contract address", r.ContractAddr, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if r.CodeID == 0 {
		return errors.New("AVM 2.0 contract code id must be positive")
	}
	if err := validateEngineToken("AVM 2.0 contract creator", r.Creator, MaxAVM2TokenLength); err != nil {
		return err
	}
	if r.AdminOptional != "" {
		if err := validateEngineToken("AVM 2.0 contract admin", r.AdminOptional, MaxAVM2TokenLength); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract storage root", r.StorageRoot); err != nil {
		return err
	}
	if r.BalanceNAET.IsNil() {
		return errors.New("AVM 2.0 contract balance must be explicit")
	}
	if r.BalanceNAET.IsNegative() {
		return errors.New("AVM 2.0 contract balance cannot be negative")
	}
	if r.CreatedAtHeight == 0 || r.UpdatedAtHeight < r.CreatedAtHeight {
		return errors.New("AVM 2.0 contract heights are invalid")
	}
	if r.InstanceVersion == 0 {
		return errors.New("AVM 2.0 contract instance version must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVM2ContractRecordHash(r) {
		return errors.New("AVM 2.0 contract record hash mismatch")
	}
	return nil
}

func (v AVM2ContractStorageValue) Validate() error {
	v = canonicalAVM2ContractStorageValue(v)
	if err := validateEngineToken("AVM 2.0 storage contract address", v.ContractAddr, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 storage key", v.StorageKey, MaxAVMStorageKeyLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 storage value hash", v.ValueHash); err != nil {
		return err
	}
	if v.ShardKey != AVM2ContractStorageShardKey(v.ContractAddr, v.StorageKey) {
		return errors.New("AVM 2.0 storage shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 storage record hash", v.RecordHash); err != nil {
		return err
	}
	if v.RecordHash != ComputeAVM2ContractStorageValueHash(v) {
		return errors.New("AVM 2.0 storage record hash mismatch")
	}
	return nil
}

func (r AVM2ContractABIDescriptorRecord) Validate() error {
	r = canonicalAVM2ContractABIDescriptorRecord(r)
	if r.CodeID == 0 || r.CodeID != r.ABI.CodeID {
		return errors.New("AVM 2.0 ABI record code id mismatch")
	}
	if err := r.ABI.Validate(DefaultAVM2Limits()); err != nil {
		return err
	}
	if r.Key != AVM2ContractABIStateKey(r.CodeID, r.ABI.ABIVersion) {
		return errors.New("AVM 2.0 ABI state key mismatch")
	}
	if r.ShardKey != AVM2ContractCodeShardKey(r.CodeID) {
		return errors.New("AVM 2.0 ABI shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVM2ContractABIDescriptorRecordHash(r) {
		return errors.New("AVM 2.0 ABI record hash mismatch")
	}
	return nil
}

func (r AVM2ContractEventRecord) Validate() error {
	r = canonicalAVM2ContractEventRecord(r)
	if err := r.Event.Validate(DefaultAVM2Limits()); err != nil {
		return err
	}
	if r.Key != AVM2ContractEventStateKey(r.Event.Height, r.Event.ContractAddress, r.Event.EventID) {
		return errors.New("AVM 2.0 event state key mismatch")
	}
	if r.ShardKey != r.Event.ContractAddress {
		return errors.New("AVM 2.0 event shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 event record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVM2ContractEventRecordHash(r) {
		return errors.New("AVM 2.0 event record hash mismatch")
	}
	return nil
}

func (r AVM2ContractMessageNonceRecord) Validate() error {
	r = canonicalAVM2ContractMessageNonceRecord(r)
	if err := validateEngineToken("AVM 2.0 nonce contract address", r.ContractAddr, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if r.Key != AVM2ContractMessageNonceStateKey(r.ContractAddr) {
		return errors.New("AVM 2.0 message nonce key mismatch")
	}
	if r.ShardKey != r.ContractAddr {
		return errors.New("AVM 2.0 message nonce shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 message nonce record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVM2ContractMessageNonceRecordHash(r) {
		return errors.New("AVM 2.0 message nonce record hash mismatch")
	}
	return nil
}

func (i AVM2ContractStateIndex) Validate() error {
	i = canonicalAVM2ContractStateIndex(i)
	if err := i.Layout.Validate(); err != nil {
		return err
	}
	if err := validateAVM2CodeRecords(i.Codes); err != nil {
		return err
	}
	if err := validateAVM2ContractRecords(i.Contracts, i.Codes); err != nil {
		return err
	}
	if err := validateAVM2StorageValues(i.Storage); err != nil {
		return err
	}
	if err := validateAVM2ABIRecords(i.ABIs); err != nil {
		return err
	}
	if err := validateAVM2EventRecords(i.Events); err != nil {
		return err
	}
	if err := validateAVM2NonceRecords(i.MessageNonces); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract state index root", i.StateRoot); err != nil {
		return err
	}
	if i.StateRoot != ComputeAVM2ContractStateIndexRoot(i) {
		return errors.New("AVM 2.0 contract state index root mismatch")
	}
	return nil
}

func AVM2ContractCodeStateKey(codeID uint64) string {
	return fmt.Sprintf("%s/%020d", AVM2ContractStatePrefixCode, codeID)
}

func AVM2ContractInstanceStateKey(contractAddr string) string {
	return AVM2ContractStatePrefixInstance + "/" + strings.TrimSpace(contractAddr)
}

func AVM2ContractStorageStateKey(contractAddr, storageKey string) string {
	return AVM2ContractStatePrefixStorage + "/" + strings.TrimSpace(contractAddr) + "/" + strings.TrimSpace(storageKey)
}

func AVM2ContractABIStateKey(codeID, abiVersion uint64) string {
	return fmt.Sprintf("%s/%020d/%020d", AVM2ContractStatePrefixABI, codeID, abiVersion)
}

func AVM2ContractEventStateKey(height uint64, contractAddr, eventID string) string {
	return fmt.Sprintf("%s/%020d/%s/%s", AVM2ContractStatePrefixEvents, height, strings.TrimSpace(contractAddr), strings.TrimSpace(eventID))
}

func AVM2ContractMessageNonceStateKey(contractAddr string) string {
	return AVM2ContractStatePrefixMessageNonce + "/" + strings.TrimSpace(contractAddr)
}

func AVM2ContractCodeShardKey(codeID uint64) string {
	return fmt.Sprintf("code_id/%020d", codeID)
}

func AVM2ContractStorageShardKey(contractAddr, storageKey string) string {
	prefix := strings.TrimSpace(storageKey)
	if idx := strings.Index(prefix, "/"); idx >= 0 {
		prefix = prefix[:idx]
	}
	return strings.TrimSpace(contractAddr) + "/" + prefix
}

func ComputeAVM2ContractStateEntryHash(entry AVM2ContractStateEntry) string {
	entry = canonicalAVM2ContractStateEntry(entry)
	entry.EntryHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-state-entry-v1")
	writeEnginePart(h, entry.Key)
	writeEnginePart(h, entry.ValueType)
	writeEnginePart(h, entry.Purpose)
	writeEnginePart(h, entry.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractStateLayoutRoot(layout AVM2ContractStateLayout) string {
	layout = canonicalAVM2ContractStateLayout(layout)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-state-layout-v1")
	writeEngineUint64(h, uint64(len(layout.Entries)))
	for _, entry := range layout.Entries {
		writeEnginePart(h, entry.EntryHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2CodeRecordHash(record AVM2CodeRecord) string {
	record = canonicalAVM2CodeRecord(record)
	record.RecordHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-code-record-v1")
	writeEngineUint64(h, record.CodeID)
	writeEnginePart(h, record.CodeHash)
	writeEngineUint64(h, record.VMVersion)
	writeEngineUint64(h, record.InstructionSetVersion)
	writeEnginePart(h, record.ABIHash)
	writeEnginePart(h, record.Deployer)
	writeEngineUint64(h, record.CreatedAtHeight)
	writeEnginePart(h, record.CodeBytesRef)
	writeEnginePart(h, record.MeteringProfile)
	writeEngineBool(h, record.Enabled)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractRecordHash(record AVM2ContractRecord) string {
	record = canonicalAVM2ContractRecord(record)
	record.RecordHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-record-v1")
	writeEnginePart(h, record.ContractAddr)
	writeEngineUint64(h, record.CodeID)
	writeEnginePart(h, record.Creator)
	writeEnginePart(h, record.AdminOptional)
	writeEnginePart(h, record.StorageRoot)
	writeEnginePart(h, record.BalanceNAET.String())
	writeEngineUint64(h, record.CreatedAtHeight)
	writeEngineUint64(h, record.UpdatedAtHeight)
	writeEngineUint64(h, record.InstanceVersion)
	writeEngineUint64(h, uint64(record.ShardID))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractStorageValueHash(value AVM2ContractStorageValue) string {
	value = canonicalAVM2ContractStorageValue(value)
	value.RecordHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-storage-value-v1")
	writeEnginePart(h, value.ContractAddr)
	writeEnginePart(h, value.StorageKey)
	writeEnginePart(h, value.ValueHash)
	writeEngineUint64(h, value.ValueBytes)
	writeEnginePart(h, value.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractABIDescriptorRecordHash(record AVM2ContractABIDescriptorRecord) string {
	record = canonicalAVM2ContractABIDescriptorRecord(record)
	record.RecordHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-abi-record-v1")
	writeEngineUint64(h, record.CodeID)
	writeEnginePart(h, record.ABI.InterfaceHash)
	writeEnginePart(h, record.Key)
	writeEnginePart(h, record.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractEventRecordHash(record AVM2ContractEventRecord) string {
	record = canonicalAVM2ContractEventRecord(record)
	record.RecordHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-event-record-v1")
	writeEnginePart(h, record.Event.EventHash)
	writeEnginePart(h, record.Key)
	writeEnginePart(h, record.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractMessageNonceRecordHash(record AVM2ContractMessageNonceRecord) string {
	record = canonicalAVM2ContractMessageNonceRecord(record)
	record.RecordHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-message-nonce-v1")
	writeEnginePart(h, record.ContractAddr)
	writeEngineUint64(h, record.Nonce)
	writeEnginePart(h, record.Key)
	writeEnginePart(h, record.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ContractStateIndexRoot(index AVM2ContractStateIndex) string {
	index = canonicalAVM2ContractStateIndex(index)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm2-contract-state-index-v1")
	writeEnginePart(h, index.Layout.LayoutRoot)
	writeEngineUint64(h, uint64(len(index.Codes)))
	for _, code := range index.Codes {
		writeEnginePart(h, code.RecordHash)
	}
	writeEngineUint64(h, uint64(len(index.Contracts)))
	for _, contract := range index.Contracts {
		writeEnginePart(h, contract.RecordHash)
	}
	writeEngineUint64(h, uint64(len(index.Storage)))
	for _, storage := range index.Storage {
		writeEnginePart(h, storage.RecordHash)
	}
	writeEngineUint64(h, uint64(len(index.ABIs)))
	for _, abi := range index.ABIs {
		writeEnginePart(h, abi.RecordHash)
	}
	writeEngineUint64(h, uint64(len(index.Events)))
	for _, event := range index.Events {
		writeEnginePart(h, event.RecordHash)
	}
	writeEngineUint64(h, uint64(len(index.MessageNonces)))
	for _, nonce := range index.MessageNonces {
		writeEnginePart(h, nonce.RecordHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVM2ContractStateEntry(entry AVM2ContractStateEntry) AVM2ContractStateEntry {
	entry.Key = strings.TrimSpace(entry.Key)
	entry.ValueType = strings.TrimSpace(entry.ValueType)
	entry.Purpose = strings.TrimSpace(entry.Purpose)
	entry.ShardKey = strings.TrimSpace(entry.ShardKey)
	entry.EntryHash = strings.TrimSpace(entry.EntryHash)
	return entry
}

func canonicalAVM2ContractStateLayout(layout AVM2ContractStateLayout) AVM2ContractStateLayout {
	layout.Entries = append([]AVM2ContractStateEntry(nil), layout.Entries...)
	for i := range layout.Entries {
		layout.Entries[i] = canonicalAVM2ContractStateEntry(layout.Entries[i])
	}
	sort.SliceStable(layout.Entries, func(i, j int) bool {
		return layout.Entries[i].Key < layout.Entries[j].Key
	})
	layout.LayoutRoot = strings.TrimSpace(layout.LayoutRoot)
	return layout
}

func canonicalAVM2CodeRecord(record AVM2CodeRecord) AVM2CodeRecord {
	record.CodeHash = strings.TrimSpace(record.CodeHash)
	record.ABIHash = strings.TrimSpace(record.ABIHash)
	record.Deployer = strings.TrimSpace(record.Deployer)
	record.CodeBytesRef = strings.TrimSpace(record.CodeBytesRef)
	record.MeteringProfile = strings.TrimSpace(record.MeteringProfile)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVM2ContractRecord(record AVM2ContractRecord) AVM2ContractRecord {
	record.ContractAddr = strings.TrimSpace(record.ContractAddr)
	record.Creator = strings.TrimSpace(record.Creator)
	record.AdminOptional = strings.TrimSpace(record.AdminOptional)
	record.StorageRoot = strings.TrimSpace(record.StorageRoot)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVM2ContractStorageValue(value AVM2ContractStorageValue) AVM2ContractStorageValue {
	value.ContractAddr = strings.TrimSpace(value.ContractAddr)
	value.StorageKey = strings.TrimSpace(value.StorageKey)
	value.ValueHash = strings.TrimSpace(value.ValueHash)
	value.ShardKey = strings.TrimSpace(value.ShardKey)
	value.RecordHash = strings.TrimSpace(value.RecordHash)
	return value
}

func canonicalAVM2ContractABIDescriptorRecord(record AVM2ContractABIDescriptorRecord) AVM2ContractABIDescriptorRecord {
	record.ABI = canonicalAVM2ABIDescriptor(record.ABI)
	record.Key = strings.TrimSpace(record.Key)
	record.ShardKey = strings.TrimSpace(record.ShardKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVM2ContractEventRecord(record AVM2ContractEventRecord) AVM2ContractEventRecord {
	record.Event = canonicalAVM2Event(record.Event)
	record.Key = strings.TrimSpace(record.Key)
	record.ShardKey = strings.TrimSpace(record.ShardKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVM2ContractMessageNonceRecord(record AVM2ContractMessageNonceRecord) AVM2ContractMessageNonceRecord {
	record.ContractAddr = strings.TrimSpace(record.ContractAddr)
	record.Key = strings.TrimSpace(record.Key)
	record.ShardKey = strings.TrimSpace(record.ShardKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVM2ContractStateIndex(index AVM2ContractStateIndex) AVM2ContractStateIndex {
	index.Layout = canonicalAVM2ContractStateLayout(index.Layout)
	index.Codes = append([]AVM2CodeRecord(nil), index.Codes...)
	for i := range index.Codes {
		index.Codes[i] = canonicalAVM2CodeRecord(index.Codes[i])
	}
	sort.SliceStable(index.Codes, func(i, j int) bool { return index.Codes[i].CodeID < index.Codes[j].CodeID })
	index.Contracts = append([]AVM2ContractRecord(nil), index.Contracts...)
	for i := range index.Contracts {
		index.Contracts[i] = canonicalAVM2ContractRecord(index.Contracts[i])
	}
	sort.SliceStable(index.Contracts, func(i, j int) bool { return index.Contracts[i].ContractAddr < index.Contracts[j].ContractAddr })
	index.Storage = append([]AVM2ContractStorageValue(nil), index.Storage...)
	for i := range index.Storage {
		index.Storage[i] = canonicalAVM2ContractStorageValue(index.Storage[i])
	}
	sort.SliceStable(index.Storage, func(i, j int) bool {
		if index.Storage[i].ContractAddr == index.Storage[j].ContractAddr {
			return index.Storage[i].StorageKey < index.Storage[j].StorageKey
		}
		return index.Storage[i].ContractAddr < index.Storage[j].ContractAddr
	})
	index.ABIs = append([]AVM2ContractABIDescriptorRecord(nil), index.ABIs...)
	for i := range index.ABIs {
		index.ABIs[i] = canonicalAVM2ContractABIDescriptorRecord(index.ABIs[i])
	}
	sort.SliceStable(index.ABIs, func(i, j int) bool {
		if index.ABIs[i].CodeID == index.ABIs[j].CodeID {
			return index.ABIs[i].ABI.ABIVersion < index.ABIs[j].ABI.ABIVersion
		}
		return index.ABIs[i].CodeID < index.ABIs[j].CodeID
	})
	index.Events = append([]AVM2ContractEventRecord(nil), index.Events...)
	for i := range index.Events {
		index.Events[i] = canonicalAVM2ContractEventRecord(index.Events[i])
	}
	sort.SliceStable(index.Events, func(i, j int) bool { return index.Events[i].Key < index.Events[j].Key })
	index.MessageNonces = append([]AVM2ContractMessageNonceRecord(nil), index.MessageNonces...)
	for i := range index.MessageNonces {
		index.MessageNonces[i] = canonicalAVM2ContractMessageNonceRecord(index.MessageNonces[i])
	}
	sort.SliceStable(index.MessageNonces, func(i, j int) bool { return index.MessageNonces[i].ContractAddr < index.MessageNonces[j].ContractAddr })
	index.StateRoot = strings.TrimSpace(index.StateRoot)
	return index
}

func validateAVM2CodeRecords(records []AVM2CodeRecord) error {
	seen := make(map[uint64]struct{}, len(records))
	for i, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.CodeID]; found {
			return errors.New("duplicate AVM 2.0 code id")
		}
		seen[record.CodeID] = struct{}{}
		if i > 0 && records[i-1].CodeID >= record.CodeID {
			return errors.New("AVM 2.0 code records must be sorted canonically")
		}
	}
	return nil
}

func validateAVM2ContractRecords(records []AVM2ContractRecord, codes []AVM2CodeRecord) error {
	seen := make(map[string]struct{}, len(records))
	codeIDs := make(map[uint64]struct{}, len(codes))
	for _, code := range codes {
		codeIDs[code.CodeID] = struct{}{}
	}
	for i, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := codeIDs[record.CodeID]; !found {
			return errors.New("AVM 2.0 contract references missing code")
		}
		if _, found := seen[record.ContractAddr]; found {
			return errors.New("duplicate AVM 2.0 contract address")
		}
		seen[record.ContractAddr] = struct{}{}
		if i > 0 && records[i-1].ContractAddr >= record.ContractAddr {
			return errors.New("AVM 2.0 contract records must be sorted canonically")
		}
	}
	return nil
}

func validateAVM2StorageValues(values []AVM2ContractStorageValue) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
		key := AVM2ContractStorageStateKey(value.ContractAddr, value.StorageKey)
		if _, found := seen[key]; found {
			return errors.New("duplicate AVM 2.0 storage value")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateAVM2ABIRecords(records []AVM2ContractABIDescriptorRecord) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.Key]; found {
			return errors.New("duplicate AVM 2.0 ABI descriptor")
		}
		seen[record.Key] = struct{}{}
	}
	return nil
}

func validateAVM2EventRecords(records []AVM2ContractEventRecord) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.Key]; found {
			return errors.New("duplicate AVM 2.0 contract event")
		}
		seen[record.Key] = struct{}{}
	}
	return nil
}

func validateAVM2NonceRecords(records []AVM2ContractMessageNonceRecord) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.ContractAddr]; found {
			return errors.New("duplicate AVM 2.0 message nonce record")
		}
		seen[record.ContractAddr] = struct{}{}
	}
	return nil
}

func validateAVM2ContractStateKey(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if !strings.HasPrefix(value, "contract/") {
		return fmt.Errorf("%s must start with contract/", fieldName)
	}
	if strings.Contains(value, "//") {
		return fmt.Errorf("%s must not contain empty path segments", fieldName)
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" {
			return fmt.Errorf("%s must not contain empty path segments", fieldName)
		}
		if err := validateEngineToken(fieldName+" segment", part, MaxAVMStateKeySegmentLength); err != nil {
			return err
		}
	}
	return nil
}

func validateAVM2ContentRef(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > 256 {
		return fmt.Errorf("%s must be <= 256 bytes", fieldName)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == '/' || r == ':' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}
