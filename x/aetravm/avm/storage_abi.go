package avm

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultMaxStorageKeyBytes		= uint32(MaxKeySize)
	DefaultMaxStorageValueBytes		= uint32(64 * 1024)
	DefaultMaxContractStorageBytes		= uint64(1024 * 1024)
	DefaultMaxStorageReadsPerExecution	= uint32(128)
	DefaultMaxStorageWritesPerExecution	= uint32(64)
	DefaultMaxStorageDeletesPerExecution	= uint32(64)
	DefaultMaxStorageIterationLimit		= uint32(1024)

	avmStorageKeyPrefix		= "avm/storage/"
	avmContractStorageRootDomain	= "aetra-avm-contract-storage-root-v1"
	avmGlobalStorageRootDomain	= "aetra-avm-global-storage-root-v1"
)

type StorageABIParams struct {
	MaxKeyBytes			uint32
	MaxValueBytes			uint32
	MaxContractStorageBytes		uint64
	MaxReadsPerExecution		uint32
	MaxWritesPerExecution		uint32
	MaxDeletesPerExecution		uint32
	MaxStorageIterationLimit	uint32
}

type KVPair struct {
	Key	[]byte
	Value	[]byte
}

type KVBackend interface {
	Get(key []byte) ([]byte, bool, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
	Iterate(prefix []byte, limit uint32) ([]KVPair, error)
}

type MapKVBackend struct {
	data map[string][]byte
}

type StorageABI struct {
	params	StorageABIParams
	kv	KVBackend
}

type StorageExecution struct {
	abi	*StorageABI
	reads	uint32
	writes	uint32
	deletes	uint32
}

type AVMStorageEntry struct {
	Key	[]byte
	Value	[]byte
}

type ContractStorageExport struct {
	Contract	string
	Entries		[]AVMStorageEntry
	Root		string
}

type AVMStorageState struct {
	Contracts	[]ContractStorageExport
	Root		string
}

func DefaultStorageABIParams() StorageABIParams {
	return StorageABIParams{
		MaxKeyBytes:			DefaultMaxStorageKeyBytes,
		MaxValueBytes:			DefaultMaxStorageValueBytes,
		MaxContractStorageBytes:	DefaultMaxContractStorageBytes,
		MaxReadsPerExecution:		DefaultMaxStorageReadsPerExecution,
		MaxWritesPerExecution:		DefaultMaxStorageWritesPerExecution,
		MaxDeletesPerExecution:		DefaultMaxStorageDeletesPerExecution,
		MaxStorageIterationLimit:	DefaultMaxStorageIterationLimit,
	}
}

func (p StorageABIParams) Validate() error {
	if p.MaxKeyBytes == 0 {
		return errors.New("AVM storage max key bytes must be positive")
	}
	if p.MaxValueBytes == 0 {
		return errors.New("AVM storage max value bytes must be positive")
	}
	if p.MaxContractStorageBytes == 0 {
		return errors.New("AVM storage max contract bytes must be positive")
	}
	if p.MaxReadsPerExecution == 0 {
		return errors.New("AVM storage max reads per execution must be positive")
	}
	if p.MaxWritesPerExecution == 0 {
		return errors.New("AVM storage max writes per execution must be positive")
	}
	if p.MaxDeletesPerExecution == 0 {
		return errors.New("AVM storage max deletes per execution must be positive")
	}
	if p.MaxStorageIterationLimit == 0 {
		return errors.New("AVM storage max iteration limit must be positive")
	}
	if uint64(p.MaxKeyBytes) > p.MaxContractStorageBytes {
		return errors.New("AVM storage max key bytes exceeds max contract storage bytes")
	}
	if uint64(p.MaxValueBytes) > p.MaxContractStorageBytes {
		return errors.New("AVM storage max value bytes exceeds max contract storage bytes")
	}
	return nil
}

func NewMapKVBackend() *MapKVBackend {
	return &MapKVBackend{data: make(map[string][]byte)}
}

func (m *MapKVBackend) Get(key []byte) ([]byte, bool, error) {
	if m == nil {
		return nil, false, errors.New("AVM storage KV backend is nil")
	}
	value, ok := m.data[string(key)]
	if !ok {
		return nil, false, nil
	}
	return append([]byte(nil), value...), true, nil
}

func (m *MapKVBackend) Set(key []byte, value []byte) error {
	if m == nil {
		return errors.New("AVM storage KV backend is nil")
	}
	m.data[string(key)] = cloneStorageValue(value)
	return nil
}

func (m *MapKVBackend) Delete(key []byte) error {
	if m == nil {
		return errors.New("AVM storage KV backend is nil")
	}
	delete(m.data, string(key))
	return nil
}

func (m *MapKVBackend) Iterate(prefix []byte, limit uint32) ([]KVPair, error) {
	if m == nil {
		return nil, errors.New("AVM storage KV backend is nil")
	}
	if limit == 0 {
		return nil, errors.New("AVM storage iteration limit is required")
	}
	keys := make([]string, 0)
	prefixText := string(prefix)
	for key := range m.data {
		if strings.HasPrefix(key, prefixText) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if uint32(len(keys)) > limit {
		keys = keys[:limit]
	}
	out := make([]KVPair, 0, len(keys))
	for _, key := range keys {
		out = append(out, KVPair{
			Key:	[]byte(key),
			Value:	append([]byte(nil), m.data[key]...),
		})
	}
	return out, nil
}

func NewStorageABI(params StorageABIParams, kv KVBackend) (*StorageABI, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, errors.New("AVM storage KV backend is required")
	}
	return &StorageABI{params: params, kv: kv}, nil
}

func MustNewStorageABI(params StorageABIParams, kv KVBackend) *StorageABI {
	abi, err := NewStorageABI(params, kv)
	if err != nil {
		panic(err)
	}
	return abi
}

func (a *StorageABI) Params() StorageABIParams {
	if a == nil {
		return StorageABIParams{}
	}
	return a.params
}

func (a *StorageABI) BeginExecution() *StorageExecution {
	return &StorageExecution{abi: a}
}

func (a *StorageABI) GetStorage(contract string, key []byte) ([]byte, bool, error) {
	return a.BeginExecution().GetStorage(contract, key)
}

func (a *StorageABI) SetStorage(contract string, key []byte, value []byte) error {
	return a.BeginExecution().SetStorage(contract, key, value)
}

func (a *StorageABI) DeleteStorage(contract string, key []byte) error {
	return a.BeginExecution().DeleteStorage(contract, key)
}

func (a *StorageABI) IterateStorage(contract string, prefix []byte, limit uint32) ([]AVMStorageEntry, error) {
	return a.BeginExecution().IterateStorage(contract, prefix, limit)
}

func (e *StorageExecution) GetStorage(contract string, key []byte) ([]byte, bool, error) {
	if err := e.validateReady(); err != nil {
		return nil, false, err
	}
	if err := e.abi.validateContractAndKey(contract, key); err != nil {
		return nil, false, err
	}
	if e.reads+1 > e.abi.params.MaxReadsPerExecution {
		return nil, false, errors.New("AVM storage read limit exceeded")
	}
	value, ok, err := e.abi.kv.Get(contractStorageKVKey(contract, key))
	if err != nil {
		return nil, false, err
	}
	e.reads++
	return value, ok, nil
}

func (e *StorageExecution) SetStorage(contract string, key []byte, value []byte) error {
	if err := e.validateReady(); err != nil {
		return err
	}
	if err := e.abi.validateContractAndKey(contract, key); err != nil {
		return err
	}
	if uint32(len(value)) > e.abi.params.MaxValueBytes {
		return fmt.Errorf("AVM storage value must be <= %d bytes", e.abi.params.MaxValueBytes)
	}
	if e.writes+1 > e.abi.params.MaxWritesPerExecution {
		return errors.New("AVM storage write limit exceeded")
	}
	nextSize, err := e.abi.contractStorageBytesAfterSet(contract, key, value)
	if err != nil {
		return err
	}
	if nextSize > e.abi.params.MaxContractStorageBytes {
		return fmt.Errorf("AVM contract storage exceeds %d bytes", e.abi.params.MaxContractStorageBytes)
	}
	if err := e.abi.kv.Set(contractStorageKVKey(contract, key), cloneStorageValue(value)); err != nil {
		return err
	}
	e.writes++
	return nil
}

func (e *StorageExecution) DeleteStorage(contract string, key []byte) error {
	if err := e.validateReady(); err != nil {
		return err
	}
	if err := e.abi.validateContractAndKey(contract, key); err != nil {
		return err
	}
	if e.deletes+1 > e.abi.params.MaxDeletesPerExecution {
		return errors.New("AVM storage delete limit exceeded")
	}
	if err := e.abi.kv.Delete(contractStorageKVKey(contract, key)); err != nil {
		return err
	}
	e.deletes++
	return nil
}

func (e *StorageExecution) IterateStorage(contract string, prefix []byte, limit uint32) ([]AVMStorageEntry, error) {
	if err := e.validateReady(); err != nil {
		return nil, err
	}
	if err := validateRawContractAddress(contract); err != nil {
		return nil, err
	}
	if uint32(len(prefix)) > e.abi.params.MaxKeyBytes {
		return nil, fmt.Errorf("AVM storage prefix must be <= %d bytes", e.abi.params.MaxKeyBytes)
	}
	if limit == 0 {
		return nil, errors.New("AVM storage iteration limit is required")
	}
	if limit > e.abi.params.MaxStorageIterationLimit {
		return nil, fmt.Errorf("AVM storage iteration limit must be <= %d", e.abi.params.MaxStorageIterationLimit)
	}
	pairs, err := e.abi.kv.Iterate(contractStorageKVPrefixForPrefix(contract, prefix), limit)
	if err != nil {
		return nil, err
	}
	entries, err := storageEntriesFromKVPairs(contract, pairs)
	if err != nil {
		return nil, err
	}
	if uint32(len(entries)) > e.abi.params.MaxReadsPerExecution-e.reads {
		return nil, errors.New("AVM storage read limit exceeded")
	}
	e.reads += uint32(len(entries))
	return entries, nil
}

func (a *StorageABI) ContractStateRoot(contract string) (string, error) {
	if a == nil {
		return "", errors.New("AVM storage ABI is nil")
	}
	if err := validateRawContractAddress(contract); err != nil {
		return "", err
	}
	entries, err := a.contractEntries(contract)
	if err != nil {
		return "", err
	}
	return ComputeContractStateRoot(contract, entries), nil
}

func (a *StorageABI) GlobalStateRoot() (string, error) {
	state, err := a.ExportState()
	if err != nil {
		return "", err
	}
	return state.Root, nil
}

func (a *StorageABI) ExportState() (AVMStorageState, error) {
	if a == nil {
		return AVMStorageState{}, errors.New("AVM storage ABI is nil")
	}
	pairs, err := a.kv.Iterate([]byte(avmStorageKeyPrefix), ^uint32(0))
	if err != nil {
		return AVMStorageState{}, err
	}
	byContract := make(map[string][]AVMStorageEntry)
	for _, pair := range pairs {
		contract, entry, err := decodeContractStorageKVPair(pair)
		if err != nil {
			return AVMStorageState{}, err
		}
		byContract[contract] = append(byContract[contract], entry)
	}
	contracts := make([]string, 0, len(byContract))
	for contract := range byContract {
		contracts = append(contracts, contract)
	}
	sort.Strings(contracts)
	out := AVMStorageState{Contracts: make([]ContractStorageExport, 0, len(contracts))}
	for _, contract := range contracts {
		entries := cloneAndSortStorageEntries(byContract[contract])
		out.Contracts = append(out.Contracts, ContractStorageExport{
			Contract:	contract,
			Entries:	entries,
			Root:		ComputeContractStateRoot(contract, entries),
		})
	}
	out.Root = ComputeGlobalAVMStateRoot(out)
	return out, nil
}

func ImportAVMStorageState(params StorageABIParams, state AVMStorageState) (*StorageABI, error) {
	abi, err := NewStorageABI(params, NewMapKVBackend())
	if err != nil {
		return nil, err
	}
	seenContracts := make(map[string]struct{}, len(state.Contracts))
	for _, contractState := range state.Contracts {
		if err := validateRawContractAddress(contractState.Contract); err != nil {
			return nil, err
		}
		if _, exists := seenContracts[contractState.Contract]; exists {
			return nil, fmt.Errorf("duplicate AVM storage contract namespace %s", contractState.Contract)
		}
		seenContracts[contractState.Contract] = struct{}{}
		entries := cloneAndSortStorageEntries(contractState.Entries)
		if contractState.Root != "" && contractState.Root != ComputeContractStateRoot(contractState.Contract, entries) {
			return nil, fmt.Errorf("AVM storage contract root mismatch for %s", contractState.Contract)
		}
		if err := validateSortedStorageEntries(entries); err != nil {
			return nil, err
		}
		var size uint64
		for _, entry := range entries {
			if err := abi.validateStorageEntry(entry); err != nil {
				return nil, err
			}
			size += uint64(len(entry.Key) + len(entry.Value))
			if size > params.MaxContractStorageBytes {
				return nil, fmt.Errorf("AVM contract storage exceeds %d bytes", params.MaxContractStorageBytes)
			}
			if err := abi.kv.Set(contractStorageKVKey(contractState.Contract, entry.Key), cloneStorageValue(entry.Value)); err != nil {
				return nil, err
			}
		}
	}
	exported, err := abi.ExportState()
	if err != nil {
		return nil, err
	}
	if state.Root != "" && state.Root != exported.Root {
		return nil, errors.New("AVM storage global root mismatch")
	}
	return abi, nil
}

func ComputeContractStateRoot(contract string, entries []AVMStorageEntry) string {
	sorted := cloneAndSortStorageEntries(entries)
	buf := bytes.NewBuffer(nil)
	writeString(buf, avmContractStorageRootDomain)
	writeString(buf, strings.TrimSpace(contract))
	writeU32(buf, uint32(len(sorted)))
	for _, entry := range sorted {
		writeU32(buf, uint32(len(entry.Key)))
		buf.Write(entry.Key)
		writeU32(buf, uint32(len(entry.Value)))
		buf.Write(entry.Value)
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func ComputeGlobalAVMStateRoot(state AVMStorageState) string {
	contracts := append([]ContractStorageExport(nil), state.Contracts...)
	sort.SliceStable(contracts, func(i, j int) bool {
		return contracts[i].Contract < contracts[j].Contract
	})
	buf := bytes.NewBuffer(nil)
	writeString(buf, avmGlobalStorageRootDomain)
	writeU32(buf, uint32(len(contracts)))
	for _, contractState := range contracts {
		entries := cloneAndSortStorageEntries(contractState.Entries)
		root := contractState.Root
		if root == "" {
			root = ComputeContractStateRoot(contractState.Contract, entries)
		}
		writeString(buf, strings.TrimSpace(contractState.Contract))
		writeString(buf, root)
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func (e *StorageExecution) validateReady() error {
	if e == nil || e.abi == nil {
		return errors.New("AVM storage execution is nil")
	}
	if e.abi.kv == nil {
		return errors.New("AVM storage KV backend is required")
	}
	return nil
}

func (a *StorageABI) validateContractAndKey(contract string, key []byte) error {
	if a == nil {
		return errors.New("AVM storage ABI is nil")
	}
	if err := validateRawContractAddress(contract); err != nil {
		return err
	}
	if len(key) == 0 {
		return errors.New("AVM storage key is required")
	}
	if uint32(len(key)) > a.params.MaxKeyBytes {
		return fmt.Errorf("AVM storage key must be <= %d bytes", a.params.MaxKeyBytes)
	}
	return nil
}

func (a *StorageABI) validateStorageEntry(entry AVMStorageEntry) error {
	if len(entry.Key) == 0 {
		return errors.New("AVM storage key is required")
	}
	if uint32(len(entry.Key)) > a.params.MaxKeyBytes {
		return fmt.Errorf("AVM storage key must be <= %d bytes", a.params.MaxKeyBytes)
	}
	if uint32(len(entry.Value)) > a.params.MaxValueBytes {
		return fmt.Errorf("AVM storage value must be <= %d bytes", a.params.MaxValueBytes)
	}
	return nil
}

func validateRawContractAddress(contract string) error {
	trimmed := strings.TrimSpace(contract)
	if trimmed == "" {
		return errors.New("AVM storage contract raw address is required")
	}
	if trimmed != contract {
		return errors.New("AVM storage contract raw address must not contain surrounding whitespace")
	}
	if !strings.HasPrefix(contract, "4:") {
		return errors.New("AVM storage contract namespace must be raw 4: address")
	}
	return nil
}

func (a *StorageABI) contractEntries(contract string) ([]AVMStorageEntry, error) {
	pairs, err := a.kv.Iterate(contractStorageKVPrefix(contract), ^uint32(0))
	if err != nil {
		return nil, err
	}
	return storageEntriesFromKVPairs(contract, pairs)
}

func (a *StorageABI) contractStorageBytesAfterSet(contract string, key []byte, value []byte) (uint64, error) {
	entries, err := a.contractEntries(contract)
	if err != nil {
		return 0, err
	}
	var total uint64
	updated := false
	for _, entry := range entries {
		if bytes.Equal(entry.Key, key) {
			total += uint64(len(key) + len(value))
			updated = true
			continue
		}
		total += uint64(len(entry.Key) + len(entry.Value))
	}
	if !updated {
		total += uint64(len(key) + len(value))
	}
	return total, nil
}

func contractStorageKVPrefix(contract string) []byte {
	return []byte(avmStorageKeyPrefix + contract + "/")
}

func contractStorageKVPrefixForPrefix(contract string, prefix []byte) []byte {
	return []byte(avmStorageKeyPrefix + contract + "/" + hex.EncodeToString(prefix))
}

func contractStorageKVKey(contract string, key []byte) []byte {
	return []byte(avmStorageKeyPrefix + contract + "/" + hex.EncodeToString(key))
}

func storageEntriesFromKVPairs(contract string, pairs []KVPair) ([]AVMStorageEntry, error) {
	out := make([]AVMStorageEntry, 0, len(pairs))
	prefix := string(contractStorageKVPrefix(contract))
	for _, pair := range pairs {
		if !strings.HasPrefix(string(pair.Key), prefix) {
			return nil, errors.New("AVM storage KV pair outside contract namespace")
		}
		keyHex := string(pair.Key[len(prefix):])
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			return nil, fmt.Errorf("AVM storage key decode failed: %w", err)
		}
		out = append(out, AVMStorageEntry{
			Key:	key,
			Value:	append([]byte(nil), pair.Value...),
		})
	}
	return cloneAndSortStorageEntries(out), nil
}

func decodeContractStorageKVPair(pair KVPair) (string, AVMStorageEntry, error) {
	key := string(pair.Key)
	if !strings.HasPrefix(key, avmStorageKeyPrefix) {
		return "", AVMStorageEntry{}, errors.New("AVM storage KV pair outside AVM namespace")
	}
	rest := strings.TrimPrefix(key, avmStorageKeyPrefix)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 {
		return "", AVMStorageEntry{}, errors.New("AVM storage KV key missing contract namespace")
	}
	contract := parts[0]
	if err := validateRawContractAddress(contract); err != nil {
		return "", AVMStorageEntry{}, err
	}
	decodedKey, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", AVMStorageEntry{}, fmt.Errorf("AVM storage key decode failed: %w", err)
	}
	return contract, AVMStorageEntry{Key: decodedKey, Value: append([]byte(nil), pair.Value...)}, nil
}

func cloneAndSortStorageEntries(entries []AVMStorageEntry) []AVMStorageEntry {
	out := make([]AVMStorageEntry, len(entries))
	for i, entry := range entries {
		out[i] = AVMStorageEntry{
			Key:	append([]byte(nil), entry.Key...),
			Value:	cloneStorageValue(entry.Value),
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return bytes.Compare(out[i].Key, out[j].Key) < 0
	})
	return out
}

func cloneStorageValue(value []byte) []byte {
	out := make([]byte, len(value))
	copy(out, value)
	return out
}

func validateSortedStorageEntries(entries []AVMStorageEntry) error {
	for i := 1; i < len(entries); i++ {
		if bytes.Compare(entries[i-1].Key, entries[i].Key) >= 0 {
			return errors.New("AVM storage entries must be sorted and unique")
		}
	}
	return nil
}
