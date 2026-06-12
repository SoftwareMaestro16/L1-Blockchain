package types

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxResolverMetadataEntries	= 16
	MaxResolverMetadataKeyBytes	= 48
	MaxResolverMetadataValueBytes	= 128

	ResolverMetadataRouteZone	= "route.zone"
	ResolverMetadataRouteShard	= "route.shard"
	ResolverMetadataRouteVM		= "route.vm"
	ResolverMetadataRouteEntrypoint	= "route.entrypoint"

	ResolverMetadataServicePrefix	= "service."
	ResolverMetadataInterfacePrefix	= "interface."
)

type ResolverMetadataEntry struct {
	Key	string
	Value	string
}

func ResolverMetadataServiceKey(service string) (string, error) {
	return resolverMetadataScopedKey(ResolverMetadataServicePrefix, service)
}

func ResolverMetadataInterfaceKey(interfaceID string) (string, error) {
	return resolverMetadataScopedKey(ResolverMetadataInterfacePrefix, interfaceID)
}

func EncodeResolverMetadata(entries []ResolverMetadataEntry) ([]byte, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	if len(entries) > MaxResolverMetadataEntries {
		return nil, fmt.Errorf("resolver metadata entries must not exceed %d", MaxResolverMetadataEntries)
	}
	ordered := append([]ResolverMetadataEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Key < ordered[j].Key })
	var out bytes.Buffer
	seen := map[string]struct{}{}
	for _, entry := range ordered {
		if err := ValidateResolverMetadataEntry(entry); err != nil {
			return nil, err
		}
		if _, found := seen[entry.Key]; found {
			return nil, fmt.Errorf("duplicate resolver metadata key %q", entry.Key)
		}
		seen[entry.Key] = struct{}{}
		out.WriteString(entry.Key)
		out.WriteByte('=')
		out.WriteString(entry.Value)
		out.WriteByte('\n')
	}
	if out.Len() > MaxResolverMetadataBytes {
		return nil, fmt.Errorf("resolver metadata must not exceed %d bytes", MaxResolverMetadataBytes)
	}
	return out.Bytes(), nil
}

func DecodeResolverMetadata(metadata []byte) ([]ResolverMetadataEntry, error) {
	if len(metadata) == 0 {
		return nil, nil
	}
	if len(metadata) > MaxResolverMetadataBytes {
		return nil, fmt.Errorf("resolver metadata must not exceed %d bytes", MaxResolverMetadataBytes)
	}
	raw := string(metadata)
	if !strings.HasSuffix(raw, "\n") {
		return nil, errors.New("resolver metadata must be newline terminated")
	}
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	entries := make([]ResolverMetadataEntry, 0, len(lines))
	for _, line := range lines {
		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, errors.New("resolver metadata entry must use key=value")
		}
		entry := ResolverMetadataEntry{Key: key, Value: value}
		if err := ValidateResolverMetadataEntry(entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	encoded, err := EncodeResolverMetadata(entries)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(encoded, metadata) {
		return nil, errors.New("resolver metadata entries must be sorted canonically")
	}
	return entries, nil
}

func ResolverMetadataValue(metadata []byte, key string) (string, bool, error) {
	if err := ValidateResolverMetadataKey(key); err != nil {
		return "", false, err
	}
	entries, err := DecodeResolverMetadata(metadata)
	if err != nil {
		return "", false, err
	}
	for _, entry := range entries {
		if entry.Key == key {
			return entry.Value, true, nil
		}
	}
	return "", false, nil
}

func ValidateResolverMetadataEntry(entry ResolverMetadataEntry) error {
	if err := ValidateResolverMetadataKey(entry.Key); err != nil {
		return err
	}
	if entry.Value == "" {
		return fmt.Errorf("resolver metadata value for %q is required", entry.Key)
	}
	if strings.TrimSpace(entry.Value) != entry.Value {
		return fmt.Errorf("resolver metadata value for %q must not have surrounding whitespace", entry.Key)
	}
	if len(entry.Value) > MaxResolverMetadataValueBytes {
		return fmt.Errorf("resolver metadata value for %q must not exceed %d bytes", entry.Key, MaxResolverMetadataValueBytes)
	}
	for i := 0; i < len(entry.Value); i++ {
		c := entry.Value[i]
		if c < 0x21 || c > 0x7e || c == '=' {
			return fmt.Errorf("resolver metadata value for %q contains unsupported character %q", entry.Key, c)
		}
	}
	return nil
}

func ValidateResolverMetadataKey(key string) error {
	if key == "" {
		return errors.New("resolver metadata key is required")
	}
	if strings.TrimSpace(key) != key {
		return errors.New("resolver metadata key must not have surrounding whitespace")
	}
	if len(key) > MaxResolverMetadataKeyBytes {
		return fmt.Errorf("resolver metadata key must not exceed %d bytes", MaxResolverMetadataKeyBytes)
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			continue
		}
		return fmt.Errorf("resolver metadata key contains unsupported character %q", c)
	}
	return nil
}

func resolverMetadataScopedKey(prefix string, value string) (string, error) {
	if value == "" {
		return "", errors.New("resolver metadata scoped key value is required")
	}
	key := prefix + value
	if err := ValidateResolverMetadataKey(key); err != nil {
		return "", err
	}
	return key, nil
}
