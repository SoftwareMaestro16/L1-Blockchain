package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const StatePrefixModelVersion = uint64(1)

type StatePrefixNamespace string

const (
	StatePrefixNamespaceCore	StatePrefixNamespace	= "Core"
	StatePrefixNamespaceZone	StatePrefixNamespace	= "Zone"
	StatePrefixNamespaceShard	StatePrefixNamespace	= "Shard"
)

type GlobalStatePrefixDescriptor struct {
	Namespace	StatePrefixNamespace
	Prefix		string
	Purpose		string
	ProofScope	string
	DescriptorHash	string
}

type GlobalStatePrefixModel struct {
	Version	uint64
	Entries	[]GlobalStatePrefixDescriptor
	Root	string
}

func GlobalStatePrefixDescriptors() []GlobalStatePrefixDescriptor {
	return []GlobalStatePrefixDescriptor{
		statePrefix(StatePrefixNamespaceCore, "core/*", "Aether Core global state and module metadata.", "Core root"),
		statePrefix(StatePrefixNamespaceCore, "core/zones/*", "Zone descriptors, capabilities, versions, and enabled state.", "Zone descriptor root"),
		statePrefix(StatePrefixNamespaceCore, "core/zone_roots/*", "Per-height ZoneCommitment records.", "Global zone root"),
		statePrefix(StatePrefixNamespaceCore, "core/message_roots/*", "Global inbox, outbox, message, and receipt roots.", "Global message root"),
		statePrefix(StatePrefixNamespaceCore, "core/proof_roots/*", "Universal proof root registry by height and root type.", "Proof registry root"),
		statePrefix(StatePrefixNamespaceZone, "zone/{zone_id}/params", "Zone-local params and execution limits.", "Zone state root"),
		statePrefix(StatePrefixNamespaceZone, "zone/{zone_id}/shards/*", "Shard layout, descriptors, metrics, and migration state.", "Shard layout root"),
		statePrefix(StatePrefixNamespaceZone, "zone/{zone_id}/state/*", "Zone-owned application or module state.", "Zone state root"),
		statePrefix(StatePrefixNamespaceZone, "zone/{zone_id}/inbox/*", "Zone-level inbound message queue.", "Zone inbox root"),
		statePrefix(StatePrefixNamespaceZone, "zone/{zone_id}/outbox/*", "Zone-level outbound message queue.", "Zone outbox root"),
		statePrefix(StatePrefixNamespaceZone, "zone/{zone_id}/receipts/*", "Zone-level message and execution receipts.", "Zone receipt root"),
		statePrefix(StatePrefixNamespaceZone, "zone/{zone_id}/events/*", "Deterministic zone event records.", "Zone event root"),
		statePrefix(StatePrefixNamespaceShard, "zone/{zone_id}/shard/{shard_id}/state/*", "Shard-local partition of zone state.", "Shard state root"),
		statePrefix(StatePrefixNamespaceShard, "zone/{zone_id}/shard/{shard_id}/inbox/*", "Shard-local inbound message queue.", "Shard inbox root"),
		statePrefix(StatePrefixNamespaceShard, "zone/{zone_id}/shard/{shard_id}/outbox/*", "Shard-local outbound message queue.", "Shard outbox root"),
		statePrefix(StatePrefixNamespaceShard, "zone/{zone_id}/shard/{shard_id}/receipts/*", "Shard-local message receipts.", "Shard receipt root"),
		statePrefix(StatePrefixNamespaceShard, "zone/{zone_id}/shard/{shard_id}/metrics/*", "Gas, fee, queue, state-size, conflict, and proof-latency metrics.", "Shard metrics root"),
	}
}

func BuildGlobalStatePrefixModel(entries []GlobalStatePrefixDescriptor) (GlobalStatePrefixModel, error) {
	model := GlobalStatePrefixModel{
		Version:	StatePrefixModelVersion,
		Entries:	normalizeGlobalStatePrefixDescriptors(entries),
	}
	if err := model.ValidateFormat(); err != nil {
		return GlobalStatePrefixModel{}, err
	}
	model.Root = ComputeGlobalStatePrefixModelRoot(model.Entries)
	return model, model.Validate()
}

func DefaultGlobalStatePrefixModel() (GlobalStatePrefixModel, error) {
	return BuildGlobalStatePrefixModel(GlobalStatePrefixDescriptors())
}

func (m GlobalStatePrefixModel) Normalize() GlobalStatePrefixModel {
	if m.Version == 0 {
		m.Version = StatePrefixModelVersion
	}
	m.Entries = normalizeGlobalStatePrefixDescriptors(m.Entries)
	m.Root = strings.ToLower(strings.TrimSpace(m.Root))
	return m
}

func (m GlobalStatePrefixModel) ValidateFormat() error {
	m = m.Normalize()
	if m.Version != StatePrefixModelVersion {
		return fmt.Errorf("aetracore state prefix model version must be %d", StatePrefixModelVersion)
	}
	if len(m.Entries) == 0 {
		return errors.New("aetracore state prefix model requires entries")
	}
	seenPrefixes := make(map[string]struct{}, len(m.Entries))
	previousKey := ""
	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seenPrefixes[entry.Prefix]; found {
			return fmt.Errorf("aetracore duplicate state prefix %s", entry.Prefix)
		}
		seenPrefixes[entry.Prefix] = struct{}{}
		key := string(entry.Namespace) + "/" + entry.Prefix
		if previousKey != "" && previousKey >= key {
			return errors.New("aetracore state prefixes must be sorted canonically")
		}
		previousKey = key
	}
	if m.Root != "" {
		if err := ValidateHash("aetracore state prefix model root", m.Root); err != nil {
			return err
		}
	}
	return nil
}

func (m GlobalStatePrefixModel) Validate() error {
	m = m.Normalize()
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.Root == "" {
		return errors.New("aetracore state prefix model root is required")
	}
	expected := ComputeGlobalStatePrefixModelRoot(m.Entries)
	if m.Root != expected {
		return fmt.Errorf("aetracore state prefix model root mismatch: expected %s", expected)
	}
	return nil
}

func BuildGlobalStatePrefixDescriptor(entry GlobalStatePrefixDescriptor) (GlobalStatePrefixDescriptor, error) {
	entry = entry.Normalize()
	if entry.DescriptorHash != "" {
		return GlobalStatePrefixDescriptor{}, errors.New("aetracore state prefix descriptor hash must be empty before construction")
	}
	if err := entry.ValidateFormat(); err != nil {
		return GlobalStatePrefixDescriptor{}, err
	}
	entry.DescriptorHash = ComputeGlobalStatePrefixDescriptorHash(entry)
	return entry, entry.Validate()
}

func (entry GlobalStatePrefixDescriptor) Normalize() GlobalStatePrefixDescriptor {
	entry.Prefix = strings.TrimSpace(entry.Prefix)
	entry.Purpose = normalizeModuleMapText(entry.Purpose)
	entry.ProofScope = normalizeModuleMapText(entry.ProofScope)
	entry.DescriptorHash = strings.ToLower(strings.TrimSpace(entry.DescriptorHash))
	return entry
}

func (entry GlobalStatePrefixDescriptor) ValidateFormat() error {
	entry = entry.Normalize()
	if !IsStatePrefixNamespace(entry.Namespace) {
		return fmt.Errorf("unknown aetracore state prefix namespace %q", entry.Namespace)
	}
	if err := validateStatePrefixPattern(entry.Namespace, entry.Prefix); err != nil {
		return err
	}
	if entry.Purpose == "" {
		return errors.New("aetracore state prefix purpose is required")
	}
	if entry.ProofScope == "" {
		return errors.New("aetracore state prefix proof scope is required")
	}
	if entry.DescriptorHash != "" {
		if err := ValidateHash("aetracore state prefix descriptor hash", entry.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (entry GlobalStatePrefixDescriptor) Validate() error {
	entry = entry.Normalize()
	if err := entry.ValidateFormat(); err != nil {
		return err
	}
	if entry.DescriptorHash == "" {
		return errors.New("aetracore state prefix descriptor hash is required")
	}
	expected := ComputeGlobalStatePrefixDescriptorHash(entry)
	if entry.DescriptorHash != expected {
		return fmt.Errorf("aetracore state prefix descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func IsStatePrefixNamespace(namespace StatePrefixNamespace) bool {
	switch namespace {
	case StatePrefixNamespaceCore, StatePrefixNamespaceZone, StatePrefixNamespaceShard:
		return true
	default:
		return false
	}
}

func ComputeGlobalStatePrefixDescriptorHash(entry GlobalStatePrefixDescriptor) string {
	entry = entry.Normalize()
	return hashParts(
		"aetra-global-state-prefix-descriptor-v1",
		string(entry.Namespace),
		entry.Prefix,
		entry.Purpose,
		entry.ProofScope,
	)
}

func ComputeGlobalStatePrefixModelRoot(entries []GlobalStatePrefixDescriptor) string {
	normalized := normalizeGlobalStatePrefixDescriptors(entries)
	parts := []string{"aetra-global-state-prefix-model-v1", fmt.Sprintf("%020d", StatePrefixModelVersion)}
	for _, entry := range normalized {
		parts = append(parts, string(entry.Namespace), entry.Prefix, entry.DescriptorHash)
	}
	return hashParts(parts...)
}

func ValidateGlobalStatePrefixModel() error {
	model, err := DefaultGlobalStatePrefixModel()
	if err != nil {
		return err
	}
	required := []string{
		"Core|core/*",
		"Core|core/zones/*",
		"Core|core/zone_roots/*",
		"Core|core/message_roots/*",
		"Core|core/proof_roots/*",
		"Zone|zone/{zone_id}/params",
		"Zone|zone/{zone_id}/shards/*",
		"Zone|zone/{zone_id}/state/*",
		"Zone|zone/{zone_id}/inbox/*",
		"Zone|zone/{zone_id}/outbox/*",
		"Zone|zone/{zone_id}/receipts/*",
		"Zone|zone/{zone_id}/events/*",
		"Shard|zone/{zone_id}/shard/{shard_id}/state/*",
		"Shard|zone/{zone_id}/shard/{shard_id}/inbox/*",
		"Shard|zone/{zone_id}/shard/{shard_id}/outbox/*",
		"Shard|zone/{zone_id}/shard/{shard_id}/receipts/*",
		"Shard|zone/{zone_id}/shard/{shard_id}/metrics/*",
	}
	seen := make(map[string]struct{}, len(model.Entries))
	for _, entry := range model.Entries {
		seen[string(entry.Namespace)+"|"+entry.Prefix] = struct{}{}
	}
	for _, key := range required {
		if _, found := seen[key]; !found {
			return fmt.Errorf("aetracore state prefix model missing %s", key)
		}
	}
	return nil
}

func MaterializeStatePrefix(pattern string, zoneID ZoneID, shardID ShardID) (string, error) {
	pattern = strings.TrimSpace(pattern)
	if strings.Contains(pattern, "{zone_id}") {
		if err := ValidateZoneID(zoneID); err != nil {
			return "", err
		}
		pattern = strings.ReplaceAll(pattern, "{zone_id}", string(zoneID))
	}
	if strings.Contains(pattern, "{shard_id}") {
		if err := ValidateShardID(shardID); err != nil {
			return "", err
		}
		pattern = strings.ReplaceAll(pattern, "{shard_id}", string(shardID))
	}
	return strings.TrimSuffix(pattern, "*"), nil
}

func statePrefix(namespace StatePrefixNamespace, prefix, purpose, proofScope string) GlobalStatePrefixDescriptor {
	entry, err := BuildGlobalStatePrefixDescriptor(GlobalStatePrefixDescriptor{
		Namespace:	namespace,
		Prefix:		prefix,
		Purpose:	purpose,
		ProofScope:	proofScope,
	})
	if err != nil {
		panic(err)
	}
	return entry
}

func normalizeGlobalStatePrefixDescriptors(entries []GlobalStatePrefixDescriptor) []GlobalStatePrefixDescriptor {
	out := make([]GlobalStatePrefixDescriptor, len(entries))
	for i, entry := range entries {
		normalized := entry.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeGlobalStatePrefixDescriptorHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := string(out[i].Namespace) + "/" + out[i].Prefix
		right := string(out[j].Namespace) + "/" + out[j].Prefix
		return left < right
	})
	return out
}

func validateStatePrefixPattern(namespace StatePrefixNamespace, prefix string) error {
	if strings.TrimSpace(prefix) != prefix || prefix == "" {
		return errors.New("aetracore state prefix is required and must not have surrounding whitespace")
	}
	if strings.Contains(prefix, "//") {
		return errors.New("aetracore state prefix must not contain empty segments")
	}
	for _, r := range prefix {
		if r <= ' ' || r == 0x7f {
			return errors.New("aetracore state prefix must not contain whitespace or control characters")
		}
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' ||
			r == '_' || r == '-' || r == '/' || r == '*' || r == '{' || r == '}' {
			continue
		}
		return fmt.Errorf("aetracore state prefix contains invalid character %q", r)
	}
	switch namespace {
	case StatePrefixNamespaceCore:
		if !strings.HasPrefix(prefix, "core/") {
			return errors.New("aetracore core state prefix must start with core/")
		}
		if strings.Contains(prefix, "{zone_id}") || strings.Contains(prefix, "{shard_id}") {
			return errors.New("aetracore core state prefix cannot contain zone or shard placeholders")
		}
	case StatePrefixNamespaceZone:
		if !strings.HasPrefix(prefix, "zone/{zone_id}/") {
			return errors.New("aetracore zone state prefix must start with zone/{zone_id}/")
		}
		if strings.Contains(prefix, "{shard_id}") {
			return errors.New("aetracore zone state prefix cannot contain shard placeholder")
		}
	case StatePrefixNamespaceShard:
		if !strings.HasPrefix(prefix, "zone/{zone_id}/shard/{shard_id}/") {
			return errors.New("aetracore shard state prefix must start with zone/{zone_id}/shard/{shard_id}/")
		}
	default:
		return fmt.Errorf("unknown aetracore state prefix namespace %q", namespace)
	}
	if strings.Count(prefix, "*") > 1 {
		return errors.New("aetracore state prefix supports at most one wildcard")
	}
	if strings.Contains(prefix, "*") && !strings.HasSuffix(prefix, "*") {
		return errors.New("aetracore state prefix wildcard must be terminal")
	}
	if strings.Contains(prefix, "{zone_id}") && strings.Count(prefix, "{zone_id}") != 1 {
		return errors.New("aetracore state prefix must contain at most one zone placeholder")
	}
	if strings.Contains(prefix, "{shard_id}") && strings.Count(prefix, "{shard_id}") != 1 {
		return errors.New("aetracore state prefix must contain at most one shard placeholder")
	}
	return nil
}
