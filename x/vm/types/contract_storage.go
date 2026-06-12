package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

func ContractNamespace(address []byte) string {
	return "contract/" + hex.EncodeToString(address)
}

func StorageFromAVM(namespace string, storage avm.Storage) []StorageEntry {
	entries := make([]StorageEntry, 0, len(storage))
	for _, entry := range avm.Snapshot(storage) {
		entries = append(entries, StorageEntry{
			Namespace:	namespace,
			Key:		entry.Key,
			Value:		append([]byte(nil), entry.Value...),
		})
	}
	sortStorage(entries)
	return entries
}

func StorageToAVM(entries []StorageEntry) avm.Storage {
	out := make(avm.Storage, len(entries))
	for _, entry := range entries {
		out[entry.Key] = append([]byte(nil), entry.Value...)
	}
	return out
}

func ValidateStorageEntries(entries []StorageEntry, limits ContractLimits) error {
	if err := limits.Validate(); err != nil {
		return err
	}
	var total uint64
	for i, entry := range entries {
		if entry.Namespace == "" {
			return errors.New("contract storage namespace is required")
		}
		if entry.Key == "" {
			return errors.New("contract storage key is required")
		}
		if uint32(len(entry.Key)) > limits.MaxStorageKeyBytes {
			return fmt.Errorf("contract storage key bytes must be <= %d", limits.MaxStorageKeyBytes)
		}
		if uint64(len(entry.Value)) > limits.MaxStorageValueBytes {
			return fmt.Errorf("contract storage value bytes must be <= %d", limits.MaxStorageValueBytes)
		}
		if i > 0 && compareStorageEntries(entries[i-1], entry) >= 0 {
			return errors.New("contract storage entries must be sorted canonically")
		}
		total += uint64(len(entry.Namespace) + len(entry.Key) + len(entry.Value))
		if total > limits.MaxStateSizeBytes {
			return fmt.Errorf("contract state size must be <= %d bytes", limits.MaxStateSizeBytes)
		}
	}
	return nil
}

func cloneStorageEntries(entries []StorageEntry) []StorageEntry {
	out := make([]StorageEntry, len(entries))
	for i, entry := range entries {
		entry.Value = append([]byte(nil), entry.Value...)
		out[i] = entry
	}
	return out
}

func sortStorage(entries []StorageEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return compareStorageEntries(entries[i], entries[j]) < 0
	})
}

func compareStorageEntries(left, right StorageEntry) int {
	if left.Namespace < right.Namespace {
		return -1
	}
	if left.Namespace > right.Namespace {
		return 1
	}
	if left.Key < right.Key {
		return -1
	}
	if left.Key > right.Key {
		return 1
	}
	return 0
}
