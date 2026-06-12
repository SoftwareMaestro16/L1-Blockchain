package types

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxNamespaceBytes		= 64
	MaxKeyBytes			= 128
	DefaultMaxStateBytes		= uint64(64 * 1024)
	DefaultStorageRentPerByte	= uint64(1)
)

type Params struct {
	MaxStateBytes		uint64
	StorageRentPerByte	uint64
}

type Key struct {
	Namespace	string
	Path		string
}

type Entry struct {
	Key	Key
	Value	[]byte
	Version	uint64
}

type Store struct {
	params	Params
	version	uint64
	entries	map[string]Entry
}

type Snapshot struct {
	Version		uint64
	Entries		[]Entry
	StateRoot	[]byte
}

func DefaultParams() Params {
	return Params{MaxStateBytes: DefaultMaxStateBytes, StorageRentPerByte: DefaultStorageRentPerByte}
}

func NewStore(params Params) (*Store, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &Store{params: params, entries: make(map[string]Entry)}, nil
}

func (p Params) Validate() error {
	if p.MaxStateBytes == 0 {
		return errors.New("max state bytes must be positive")
	}
	return nil
}

func ContractNamespace(address string) string {
	return "contract/" + address
}

func FormatKey(key Key) (string, error) {
	if err := ValidateKey(key); err != nil {
		return "", err
	}
	return key.Namespace + "/" + key.Path, nil
}

func ValidateKey(key Key) error {
	if strings.TrimSpace(key.Namespace) == "" {
		return errors.New("storage namespace is required")
	}
	if len(key.Namespace) > MaxNamespaceBytes {
		return fmt.Errorf("storage namespace must be <= %d bytes", MaxNamespaceBytes)
	}
	if strings.TrimSpace(key.Path) == "" {
		return errors.New("storage key path is required")
	}
	if len(key.Path) > MaxKeyBytes {
		return fmt.Errorf("storage key path must be <= %d bytes", MaxKeyBytes)
	}
	return nil
}

func (s *Store) Set(key Key, value []byte) error {
	formatted, err := FormatKey(key)
	if err != nil {
		return err
	}
	next := Entry{Key: key, Value: append([]byte(nil), value...), Version: s.version + 1}
	old, hadOld := s.entries[formatted]
	s.entries[formatted] = next
	if s.StateBytes() > s.params.MaxStateBytes {
		if hadOld {
			s.entries[formatted] = old
		} else {
			delete(s.entries, formatted)
		}
		return fmt.Errorf("storage state size must be <= %d bytes", s.params.MaxStateBytes)
	}
	s.version++
	return nil
}

func (s *Store) Get(key Key) (Entry, bool, error) {
	formatted, err := FormatKey(key)
	if err != nil {
		return Entry{}, false, err
	}
	entry, ok := s.entries[formatted]
	entry.Value = append([]byte(nil), entry.Value...)
	return entry, ok, nil
}

func (s *Store) Iterate(namespace string, limit uint32) ([]Entry, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, errors.New("storage iterate namespace is required")
	}
	if limit == 0 {
		return nil, errors.New("storage iterate limit must be positive")
	}
	keys := make([]string, 0)
	prefix := namespace + "/"
	for formatted := range s.entries {
		if strings.HasPrefix(formatted, prefix) {
			keys = append(keys, formatted)
		}
	}
	sort.Strings(keys)
	if len(keys) > int(limit) {
		keys = keys[:limit]
	}
	out := make([]Entry, len(keys))
	for i, formatted := range keys {
		entry := s.entries[formatted]
		entry.Value = append([]byte(nil), entry.Value...)
		out[i] = entry
	}
	return out, nil
}

func (s *Store) Snapshot() Snapshot {
	entries := make([]Entry, 0, len(s.entries))
	keys := make([]string, 0, len(s.entries))
	for key := range s.entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		entry := s.entries[key]
		entry.Value = append([]byte(nil), entry.Value...)
		entries = append(entries, entry)
	}
	return Snapshot{Version: s.version, Entries: entries, StateRoot: StateRoot(entries)}
}

func ImportSnapshot(params Params, snapshot Snapshot) (*Store, error) {
	store, err := NewStore(params)
	if err != nil {
		return nil, err
	}
	store.version = snapshot.Version
	for _, entry := range snapshot.Entries {
		formatted, err := FormatKey(entry.Key)
		if err != nil {
			return nil, err
		}
		store.entries[formatted] = Entry{Key: entry.Key, Value: append([]byte(nil), entry.Value...), Version: entry.Version}
	}
	if !equalBytes(StateRoot(snapshot.Entries), snapshot.StateRoot) {
		return nil, errors.New("storage snapshot state root mismatch")
	}
	if store.StateBytes() > params.MaxStateBytes {
		return nil, fmt.Errorf("storage state size must be <= %d bytes", params.MaxStateBytes)
	}
	return store, nil
}

func (s *Store) StateBytes() uint64 {
	var total uint64
	for _, entry := range s.entries {
		total += uint64(len(entry.Key.Namespace) + len(entry.Key.Path) + len(entry.Value))
	}
	return total
}

func (s *Store) StorageRent() uint64 {
	return s.StateBytes() * s.params.StorageRentPerByte
}

func StateRoot(entries []Entry) []byte {
	ordered := append([]Entry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, _ := FormatKey(ordered[i].Key)
		right, _ := FormatKey(ordered[j].Key)
		return left < right
	})
	h := sha256.New()
	for _, entry := range ordered {
		formatted, _ := FormatKey(entry.Key)
		h.Write([]byte(formatted))
		h.Write([]byte{0})
		h.Write(entry.Value)
	}
	return h.Sum(nil)
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
