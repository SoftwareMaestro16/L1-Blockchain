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
	AVMContractStatePrefixCode		= "contract/code"
	AVMContractStatePrefixInstance		= "contract/instance"
	AVMContractStatePrefixStorage		= "contract/storage"
	AVMContractStatePrefixABI		= "contract/abi"
	AVMContractStatePrefixEvents		= "contract/events"
	AVMContractStatePrefixMessageNonce	= "contract/message_nonce"

	AVMContractValueCodeRecord	= "CodeRecord"
	AVMContractValueContractRecord	= "ContractRecord"
	AVMContractValueStorageValue	= "StorageValue"
	AVMContractValueABIDescriptor	= "AbiDescriptor"
	AVMContractValueContractEvent	= "ContractEvent"
	AVMContractValueMessageNonce	= "uint64"

	AVMMeteringProfileDefault	= "default-AVM"
)

type AVMContractStateEntry struct {
	Key		string
	ValueType	string
	Purpose		string
	ShardKey	string
	EntryHash	string
}

type AVMContractStateLayout struct {
	Entries		[]AVMContractStateEntry
	LayoutRoot	string
}

type AVMCodeRecord struct {
	CodeID			uint64
	CodeHash		string
	VMVersion		uint64
	InstructionSetVersion	uint64
	ABIHash			string
	Deployer		string
	CreatedAtHeight		uint64
	CodeBytesRef		string
	MeteringProfile		string
	Enabled			bool
	RecordHash		string
}

type AVMContractRecord struct {
	ContractAddr	string
	CodeID		uint64
	Creator		string
	AdminOptional	string
	StorageRoot	string
	BalanceNAET	sdkmath.Int
	CreatedAtHeight	uint64
	UpdatedAtHeight	uint64
	InstanceVersion	uint64
	ShardID		uint32
	RecordHash	string
}

type AVMContractStorageValue struct {
	ContractAddr	string
	StorageKey	string
	ValueHash	string
	ValueBytes	uint64
	ShardKey	string
	RecordHash	string
}

type AVMContractABIDescriptorRecord struct {
	CodeID		uint64
	ABI		AVMABIDescriptor
	Key		string
	ShardKey	string
	RecordHash	string
}

type AVMContractEventRecord struct {
	Event		AVMEvent
	Key		string
	ShardKey	string
	RecordHash	string
}

type AVMContractMessageNonceRecord struct {
	ContractAddr	string
	Nonce		uint64
	Key		string
	ShardKey	string
	RecordHash	string
}

type AVMContractStateIndex struct {
	Layout		AVMContractStateLayout
	Codes		[]AVMCodeRecord
	Contracts	[]AVMContractRecord
	Storage		[]AVMContractStorageValue
	ABIs		[]AVMContractABIDescriptorRecord
	Events		[]AVMContractEventRecord
	MessageNonces	[]AVMContractMessageNonceRecord
	StateRoot	string
}

func DefaultAVMContractStateLayout() (AVMContractStateLayout, error) {
	layout := AVMContractStateLayout{Entries: []AVMContractStateEntry{
		{
			Key:		AVMContractCodeStateKey(1),
			ValueType:	AVMContractValueCodeRecord,
			Purpose:	"code metadata, hashes, VM version, metering profile, and enablement",
			ShardKey:	"code_id",
		},
		{
			Key:		AVMContractInstanceStateKey("contract_addr"),
			ValueType:	AVMContractValueContractRecord,
			Purpose:	"contract instance metadata, code binding, admin, balance, storage root, and shard assignment",
			ShardKey:	"contract_addr",
		},
		{
			Key:		AVMContractStorageStateKey("contract_addr", "storage_key"),
			ValueType:	AVMContractValueStorageValue,
			Purpose:	"contract-owned persistent key-value state",
			ShardKey:	"contract_addr/storage_key_prefix",
		},
		{
			Key:		AVMContractABIStateKey(1, 1),
			ValueType:	AVMContractValueABIDescriptor,
			Purpose:	"versioned ABI and schema metadata for calls, events, and errors",
			ShardKey:	"code_id",
		},
		{
			Key:		AVMContractEventStateKey(1, "contract_addr", "event_id"),
			ValueType:	AVMContractValueContractEvent,
			Purpose:	"deterministic contract event output included in event roots",
			ShardKey:	"contract_addr",
		},
		{
			Key:		AVMContractMessageNonceStateKey("contract_addr"),
			ValueType:	AVMContractValueMessageNonce,
			Purpose:	"replay-safe nonce for contract-emitted messages",
			ShardKey:	"contract_addr",
		},
	}}
	layout = canonicalAVMContractStateLayout(layout)
	for i := range layout.Entries {
		layout.Entries[i].EntryHash = ComputeAVMContractStateEntryHash(layout.Entries[i])
	}
	layout = canonicalAVMContractStateLayout(layout)
	layout.LayoutRoot = ComputeAVMContractStateLayoutRoot(layout)
	return layout, layout.Validate()
}

func NewAVMCodeRecord(record AVMCodeRecord) (AVMCodeRecord, error) {
	record = canonicalAVMCodeRecord(record)
	record.RecordHash = ComputeAVMCodeRecordHash(record)
	return record, record.Validate()
}

func NewAVMContractRecord(record AVMContractRecord) (AVMContractRecord, error) {
	record = canonicalAVMContractRecord(record)
	record.RecordHash = ComputeAVMContractRecordHash(record)
	return record, record.Validate()
}

func NewAVMContractStorageValue(value AVMContractStorageValue) (AVMContractStorageValue, error) {
	value = canonicalAVMContractStorageValue(value)
	if value.ShardKey == "" {
		value.ShardKey = AVMContractStorageShardKey(value.ContractAddr, value.StorageKey)
	}
	value.RecordHash = ComputeAVMContractStorageValueHash(value)
	return value, value.Validate()
}

func NewAVMContractABIDescriptorRecord(record AVMContractABIDescriptorRecord) (AVMContractABIDescriptorRecord, error) {
	record = canonicalAVMContractABIDescriptorRecord(record)
	if record.CodeID == 0 {
		record.CodeID = record.ABI.CodeID
	}
	if record.Key == "" {
		record.Key = AVMContractABIStateKey(record.CodeID, record.ABI.ABIVersion)
	}
	if record.ShardKey == "" {
		record.ShardKey = AVMContractCodeShardKey(record.CodeID)
	}
	record.RecordHash = ComputeAVMContractABIDescriptorRecordHash(record)
	return record, record.Validate()
}

func NewAVMContractEventRecord(record AVMContractEventRecord) (AVMContractEventRecord, error) {
	record = canonicalAVMContractEventRecord(record)
	if record.Key == "" {
		record.Key = AVMContractEventStateKey(record.Event.Height, record.Event.ContractAddress, record.Event.EventID)
	}
	if record.ShardKey == "" {
		record.ShardKey = record.Event.ContractAddress
	}
	record.RecordHash = ComputeAVMContractEventRecordHash(record)
	return record, record.Validate()
}

func NewAVMContractMessageNonceRecord(record AVMContractMessageNonceRecord) (AVMContractMessageNonceRecord, error) {
	record = canonicalAVMContractMessageNonceRecord(record)
	if record.Key == "" {
		record.Key = AVMContractMessageNonceStateKey(record.ContractAddr)
	}
	if record.ShardKey == "" {
		record.ShardKey = record.ContractAddr
	}
	record.RecordHash = ComputeAVMContractMessageNonceRecordHash(record)
	return record, record.Validate()
}

func NewAVMContractStateIndex(index AVMContractStateIndex) (AVMContractStateIndex, error) {
	index = canonicalAVMContractStateIndex(index)
	if len(index.Layout.Entries) == 0 {
		layout, err := DefaultAVMContractStateLayout()
		if err != nil {
			return AVMContractStateIndex{}, err
		}
		index.Layout = layout
	}
	index.StateRoot = ComputeAVMContractStateIndexRoot(index)
	return index, index.Validate()
}

func ValidateAVMContractInstantiation(code AVMCodeRecord, contract AVMContractRecord) error {
	code = canonicalAVMCodeRecord(code)
	contract = canonicalAVMContractRecord(contract)
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
	if code.VMVersion != AVMVMVersion {
		return errors.New("AVM 2.0 code record has incompatible VM version")
	}
	return nil
}

func (l AVMContractStateLayout) Validate() error {
	l = canonicalAVMContractStateLayout(l)
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
		AVMContractValueCodeRecord,
		AVMContractValueContractRecord,
		AVMContractValueStorageValue,
		AVMContractValueABIDescriptor,
		AVMContractValueContractEvent,
		AVMContractValueMessageNonce,
	} {
		if _, found := seen[required]; !found {
			return fmt.Errorf("AVM 2.0 contract state layout missing %s", required)
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract state layout root", l.LayoutRoot); err != nil {
		return err
	}
	if l.LayoutRoot != ComputeAVMContractStateLayoutRoot(l) {
		return errors.New("AVM 2.0 contract state layout root mismatch")
	}
	return nil
}

func (e AVMContractStateEntry) Validate() error {
	e = canonicalAVMContractStateEntry(e)
	if err := validateAVMContractStateKey("AVM 2.0 contract state key", e.Key); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 contract state value type", e.ValueType, MaxAVMTokenLength); err != nil {
		return err
	}
	if strings.TrimSpace(e.Purpose) == "" {
		return errors.New("AVM 2.0 contract state purpose is required")
	}
	if len(e.Purpose) > 256 {
		return errors.New("AVM 2.0 contract state purpose must be <= 256 bytes")
	}
	if err := validateEngineToken("AVM 2.0 contract state shard key", strings.ReplaceAll(e.ShardKey, "_prefix", ""), MaxAVMTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract state entry hash", e.EntryHash); err != nil {
		return err
	}
	if e.EntryHash != ComputeAVMContractStateEntryHash(e) {
		return errors.New("AVM 2.0 contract state entry hash mismatch")
	}
	return nil
}

func (r AVMCodeRecord) Validate() error {
	r = canonicalAVMCodeRecord(r)
	if r.CodeID == 0 {
		return errors.New("AVM 2.0 code id must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 code hash", r.CodeHash); err != nil {
		return err
	}
	if r.VMVersion != AVMVMVersion {
		return errors.New("AVM 2.0 code record VM version must be 2")
	}
	if r.InstructionSetVersion == 0 {
		return errors.New("AVM 2.0 code instruction set version must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 code ABI hash", r.ABIHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 code deployer", r.Deployer, MaxAVMTokenLength); err != nil {
		return err
	}
	if r.CreatedAtHeight == 0 {
		return errors.New("AVM 2.0 code created height must be positive")
	}
	if err := validateAVMContentRef("AVM 2.0 code bytes ref", r.CodeBytesRef); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 metering profile", r.MeteringProfile, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 code record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMCodeRecordHash(r) {
		return errors.New("AVM 2.0 code record hash mismatch")
	}
	return nil
}

func (r AVMContractRecord) Validate() error {
	r = canonicalAVMContractRecord(r)
	if err := validateEngineToken("AVM 2.0 contract address", r.ContractAddr, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if r.CodeID == 0 {
		return errors.New("AVM 2.0 contract code id must be positive")
	}
	if err := validateEngineToken("AVM 2.0 contract creator", r.Creator, MaxAVMTokenLength); err != nil {
		return err
	}
	if r.AdminOptional != "" {
		if err := validateEngineToken("AVM 2.0 contract admin", r.AdminOptional, MaxAVMTokenLength); err != nil {
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
	if r.RecordHash != ComputeAVMContractRecordHash(r) {
		return errors.New("AVM 2.0 contract record hash mismatch")
	}
	return nil
}

func (v AVMContractStorageValue) Validate() error {
	v = canonicalAVMContractStorageValue(v)
	if err := validateEngineToken("AVM 2.0 storage contract address", v.ContractAddr, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 storage key", v.StorageKey, MaxAVMStorageKeyLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 storage value hash", v.ValueHash); err != nil {
		return err
	}
	if v.ShardKey != AVMContractStorageShardKey(v.ContractAddr, v.StorageKey) {
		return errors.New("AVM 2.0 storage shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 storage record hash", v.RecordHash); err != nil {
		return err
	}
	if v.RecordHash != ComputeAVMContractStorageValueHash(v) {
		return errors.New("AVM 2.0 storage record hash mismatch")
	}
	return nil
}

func (r AVMContractABIDescriptorRecord) Validate() error {
	r = canonicalAVMContractABIDescriptorRecord(r)
	if r.CodeID == 0 || r.CodeID != r.ABI.CodeID {
		return errors.New("AVM 2.0 ABI record code id mismatch")
	}
	if err := r.ABI.Validate(DefaultAVMLimits()); err != nil {
		return err
	}
	if r.Key != AVMContractABIStateKey(r.CodeID, r.ABI.ABIVersion) {
		return errors.New("AVM 2.0 ABI state key mismatch")
	}
	if r.ShardKey != AVMContractCodeShardKey(r.CodeID) {
		return errors.New("AVM 2.0 ABI shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMContractABIDescriptorRecordHash(r) {
		return errors.New("AVM 2.0 ABI record hash mismatch")
	}
	return nil
}

func (r AVMContractEventRecord) Validate() error {
	r = canonicalAVMContractEventRecord(r)
	if err := r.Event.Validate(DefaultAVMLimits()); err != nil {
		return err
	}
	if r.Key != AVMContractEventStateKey(r.Event.Height, r.Event.ContractAddress, r.Event.EventID) {
		return errors.New("AVM 2.0 event state key mismatch")
	}
	if r.ShardKey != r.Event.ContractAddress {
		return errors.New("AVM 2.0 event shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 event record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMContractEventRecordHash(r) {
		return errors.New("AVM 2.0 event record hash mismatch")
	}
	return nil
}

func (r AVMContractMessageNonceRecord) Validate() error {
	r = canonicalAVMContractMessageNonceRecord(r)
	if err := validateEngineToken("AVM 2.0 nonce contract address", r.ContractAddr, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if r.Key != AVMContractMessageNonceStateKey(r.ContractAddr) {
		return errors.New("AVM 2.0 message nonce key mismatch")
	}
	if r.ShardKey != r.ContractAddr {
		return errors.New("AVM 2.0 message nonce shard key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 message nonce record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMContractMessageNonceRecordHash(r) {
		return errors.New("AVM 2.0 message nonce record hash mismatch")
	}
	return nil
}

func (i AVMContractStateIndex) Validate() error {
	i = canonicalAVMContractStateIndex(i)
	if err := i.Layout.Validate(); err != nil {
		return err
	}
	if err := validateAVMCodeRecords(i.Codes); err != nil {
		return err
	}
	if err := validateAVMContractRecords(i.Contracts, i.Codes); err != nil {
		return err
	}
	if err := validateAVMStorageValues(i.Storage); err != nil {
		return err
	}
	if err := validateAVMABIRecords(i.ABIs); err != nil {
		return err
	}
	if err := validateAVMEventRecords(i.Events); err != nil {
		return err
	}
	if err := validateAVMNonceRecords(i.MessageNonces); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 contract state index root", i.StateRoot); err != nil {
		return err
	}
	if i.StateRoot != ComputeAVMContractStateIndexRoot(i) {
		return errors.New("AVM 2.0 contract state index root mismatch")
	}
	return nil
}

func AVMContractCodeStateKey(codeID uint64) string {
	return fmt.Sprintf("%s/%020d", AVMContractStatePrefixCode, codeID)
}

func AVMContractInstanceStateKey(contractAddr string) string {
	return AVMContractStatePrefixInstance + "/" + strings.TrimSpace(contractAddr)
}

func AVMContractStorageStateKey(contractAddr, storageKey string) string {
	return AVMContractStatePrefixStorage + "/" + strings.TrimSpace(contractAddr) + "/" + strings.TrimSpace(storageKey)
}

func AVMContractABIStateKey(codeID, abiVersion uint64) string {
	return fmt.Sprintf("%s/%020d/%020d", AVMContractStatePrefixABI, codeID, abiVersion)
}

func AVMContractEventStateKey(height uint64, contractAddr, eventID string) string {
	return fmt.Sprintf("%s/%020d/%s/%s", AVMContractStatePrefixEvents, height, strings.TrimSpace(contractAddr), strings.TrimSpace(eventID))
}

func AVMContractMessageNonceStateKey(contractAddr string) string {
	return AVMContractStatePrefixMessageNonce + "/" + strings.TrimSpace(contractAddr)
}

func AVMContractCodeShardKey(codeID uint64) string {
	return fmt.Sprintf("code_id/%020d", codeID)
}

func AVMContractStorageShardKey(contractAddr, storageKey string) string {
	prefix := strings.TrimSpace(storageKey)
	if idx := strings.Index(prefix, "/"); idx >= 0 {
		prefix = prefix[:idx]
	}
	return strings.TrimSpace(contractAddr) + "/" + prefix
}

func ComputeAVMContractStateEntryHash(entry AVMContractStateEntry) string {
	entry = canonicalAVMContractStateEntry(entry)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-state-entry-v1")
	writeEnginePart(h, entry.Key)
	writeEnginePart(h, entry.ValueType)
	writeEnginePart(h, entry.Purpose)
	writeEnginePart(h, entry.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractStateLayoutRoot(layout AVMContractStateLayout) string {
	layout = canonicalAVMContractStateLayout(layout)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-state-layout-v1")
	writeEngineUint64(h, uint64(len(layout.Entries)))
	for _, entry := range layout.Entries {
		writeEnginePart(h, entry.EntryHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCodeRecordHash(record AVMCodeRecord) string {
	record = canonicalAVMCodeRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-code-record-v1")
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

func ComputeAVMContractRecordHash(record AVMContractRecord) string {
	record = canonicalAVMContractRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-record-v1")
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

func ComputeAVMContractStorageValueHash(value AVMContractStorageValue) string {
	value = canonicalAVMContractStorageValue(value)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-storage-value-v1")
	writeEnginePart(h, value.ContractAddr)
	writeEnginePart(h, value.StorageKey)
	writeEnginePart(h, value.ValueHash)
	writeEngineUint64(h, value.ValueBytes)
	writeEnginePart(h, value.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractABIDescriptorRecordHash(record AVMContractABIDescriptorRecord) string {
	record = canonicalAVMContractABIDescriptorRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-abi-record-v1")
	writeEngineUint64(h, record.CodeID)
	writeEnginePart(h, record.ABI.InterfaceHash)
	writeEnginePart(h, record.Key)
	writeEnginePart(h, record.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractEventRecordHash(record AVMContractEventRecord) string {
	record = canonicalAVMContractEventRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-event-record-v1")
	writeEnginePart(h, record.Event.EventHash)
	writeEnginePart(h, record.Key)
	writeEnginePart(h, record.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractMessageNonceRecordHash(record AVMContractMessageNonceRecord) string {
	record = canonicalAVMContractMessageNonceRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-message-nonce-v1")
	writeEnginePart(h, record.ContractAddr)
	writeEngineUint64(h, record.Nonce)
	writeEnginePart(h, record.Key)
	writeEnginePart(h, record.ShardKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractStateIndexRoot(index AVMContractStateIndex) string {
	index = canonicalAVMContractStateIndex(index)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-contract-state-index-v1")
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

func canonicalAVMContractStateEntry(entry AVMContractStateEntry) AVMContractStateEntry {
	entry.Key = strings.TrimSpace(entry.Key)
	entry.ValueType = strings.TrimSpace(entry.ValueType)
	entry.Purpose = strings.TrimSpace(entry.Purpose)
	entry.ShardKey = strings.TrimSpace(entry.ShardKey)
	entry.EntryHash = strings.TrimSpace(entry.EntryHash)
	return entry
}

func canonicalAVMContractStateLayout(layout AVMContractStateLayout) AVMContractStateLayout {
	layout.Entries = append([]AVMContractStateEntry(nil), layout.Entries...)
	for i := range layout.Entries {
		layout.Entries[i] = canonicalAVMContractStateEntry(layout.Entries[i])
	}
	sort.SliceStable(layout.Entries, func(i, j int) bool {
		return layout.Entries[i].Key < layout.Entries[j].Key
	})
	layout.LayoutRoot = strings.TrimSpace(layout.LayoutRoot)
	return layout
}

func canonicalAVMCodeRecord(record AVMCodeRecord) AVMCodeRecord {
	record.CodeHash = strings.TrimSpace(record.CodeHash)
	record.ABIHash = strings.TrimSpace(record.ABIHash)
	record.Deployer = strings.TrimSpace(record.Deployer)
	record.CodeBytesRef = strings.TrimSpace(record.CodeBytesRef)
	record.MeteringProfile = strings.TrimSpace(record.MeteringProfile)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMContractRecord(record AVMContractRecord) AVMContractRecord {
	record.ContractAddr = strings.TrimSpace(record.ContractAddr)
	record.Creator = strings.TrimSpace(record.Creator)
	record.AdminOptional = strings.TrimSpace(record.AdminOptional)
	record.StorageRoot = strings.TrimSpace(record.StorageRoot)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMContractStorageValue(value AVMContractStorageValue) AVMContractStorageValue {
	value.ContractAddr = strings.TrimSpace(value.ContractAddr)
	value.StorageKey = strings.TrimSpace(value.StorageKey)
	value.ValueHash = strings.TrimSpace(value.ValueHash)
	value.ShardKey = strings.TrimSpace(value.ShardKey)
	value.RecordHash = strings.TrimSpace(value.RecordHash)
	return value
}

func canonicalAVMContractABIDescriptorRecord(record AVMContractABIDescriptorRecord) AVMContractABIDescriptorRecord {
	record.ABI = canonicalAVMABIDescriptor(record.ABI)
	record.Key = strings.TrimSpace(record.Key)
	record.ShardKey = strings.TrimSpace(record.ShardKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMContractEventRecord(record AVMContractEventRecord) AVMContractEventRecord {
	record.Event = canonicalAVMEvent(record.Event)
	record.Key = strings.TrimSpace(record.Key)
	record.ShardKey = strings.TrimSpace(record.ShardKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMContractMessageNonceRecord(record AVMContractMessageNonceRecord) AVMContractMessageNonceRecord {
	record.ContractAddr = strings.TrimSpace(record.ContractAddr)
	record.Key = strings.TrimSpace(record.Key)
	record.ShardKey = strings.TrimSpace(record.ShardKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMContractStateIndex(index AVMContractStateIndex) AVMContractStateIndex {
	index.Layout = canonicalAVMContractStateLayout(index.Layout)
	index.Codes = append([]AVMCodeRecord(nil), index.Codes...)
	for i := range index.Codes {
		index.Codes[i] = canonicalAVMCodeRecord(index.Codes[i])
	}
	sort.SliceStable(index.Codes, func(i, j int) bool { return index.Codes[i].CodeID < index.Codes[j].CodeID })
	index.Contracts = append([]AVMContractRecord(nil), index.Contracts...)
	for i := range index.Contracts {
		index.Contracts[i] = canonicalAVMContractRecord(index.Contracts[i])
	}
	sort.SliceStable(index.Contracts, func(i, j int) bool { return index.Contracts[i].ContractAddr < index.Contracts[j].ContractAddr })
	index.Storage = append([]AVMContractStorageValue(nil), index.Storage...)
	for i := range index.Storage {
		index.Storage[i] = canonicalAVMContractStorageValue(index.Storage[i])
	}
	sort.SliceStable(index.Storage, func(i, j int) bool {
		if index.Storage[i].ContractAddr == index.Storage[j].ContractAddr {
			return index.Storage[i].StorageKey < index.Storage[j].StorageKey
		}
		return index.Storage[i].ContractAddr < index.Storage[j].ContractAddr
	})
	index.ABIs = append([]AVMContractABIDescriptorRecord(nil), index.ABIs...)
	for i := range index.ABIs {
		index.ABIs[i] = canonicalAVMContractABIDescriptorRecord(index.ABIs[i])
	}
	sort.SliceStable(index.ABIs, func(i, j int) bool {
		if index.ABIs[i].CodeID == index.ABIs[j].CodeID {
			return index.ABIs[i].ABI.ABIVersion < index.ABIs[j].ABI.ABIVersion
		}
		return index.ABIs[i].CodeID < index.ABIs[j].CodeID
	})
	index.Events = append([]AVMContractEventRecord(nil), index.Events...)
	for i := range index.Events {
		index.Events[i] = canonicalAVMContractEventRecord(index.Events[i])
	}
	sort.SliceStable(index.Events, func(i, j int) bool { return index.Events[i].Key < index.Events[j].Key })
	index.MessageNonces = append([]AVMContractMessageNonceRecord(nil), index.MessageNonces...)
	for i := range index.MessageNonces {
		index.MessageNonces[i] = canonicalAVMContractMessageNonceRecord(index.MessageNonces[i])
	}
	sort.SliceStable(index.MessageNonces, func(i, j int) bool { return index.MessageNonces[i].ContractAddr < index.MessageNonces[j].ContractAddr })
	index.StateRoot = strings.TrimSpace(index.StateRoot)
	return index
}

func validateAVMCodeRecords(records []AVMCodeRecord) error {
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

func validateAVMContractRecords(records []AVMContractRecord, codes []AVMCodeRecord) error {
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

func validateAVMStorageValues(values []AVMContractStorageValue) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
		key := AVMContractStorageStateKey(value.ContractAddr, value.StorageKey)
		if _, found := seen[key]; found {
			return errors.New("duplicate AVM 2.0 storage value")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateAVMABIRecords(records []AVMContractABIDescriptorRecord) error {
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

func validateAVMEventRecords(records []AVMContractEventRecord) error {
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

func validateAVMNonceRecords(records []AVMContractMessageNonceRecord) error {
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

func validateAVMContractStateKey(fieldName, value string) error {
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

func validateAVMContentRef(fieldName, value string) error {
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
