package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const CosmosModuleMapVersion = uint64(1)

type CosmosModuleDescriptor struct {
	Module			string
	Responsibility		string
	Zone			string
	Dependencies		[]string
	AcceptanceSignal	string
	DescriptorHash		string
}

type CosmosModuleMap struct {
	Version	uint64
	Modules	[]CosmosModuleDescriptor
	Root	string
}

func CosmosSDKNewModules() []CosmosModuleDescriptor {
	return []CosmosModuleDescriptor{
		cosmosModule("x/aetracore", "Zone registry, commitments, routing epochs, proof roots.", "Core", []string{"Store v2", "ABCI++", "CometBFT finality"}, "Replays identical zone commitments and global roots from reordered inputs."),
		cosmosModule("x/msgbus", "First-class messages, inbox/outbox, receipts, routing.", "Core + zones", []string{"x/aetracore", "x/shards", "x/proofregistry"}, "Proves message inclusion and receipt delivery across zones and shards."),
		cosmosModule("x/zones", "Zone lifecycle, zone execution adapter, zone params.", "Core", []string{"x/aetracore"}, "Every zone exposes deterministic execute, inbound, export, and import hooks."),
		cosmosModule("x/shards", "Shard layout, split/merge, shard metrics.", "Zones", []string{"x/aetracore", "Store v2", "BlockSTM"}, "Routing is stable inside an epoch and changes only at committed layout boundaries."),
		cosmosModule("x/proofregistry", "Universal proof root registry and proof queries.", "Core", []string{"Store v2", "x/aetracore"}, "Clients verify account, message, zone, shard, identity, contract, and payment proofs."),
		cosmosModule("x/avm", "AVM 2.0 code, contracts, execution, ABI.", "Contract Zone", []string{"x/msgbus", "x/proofregistry", "Store v2"}, "VM execution is deterministic, metered, and emits committed message and event roots."),
		cosmosModule("x/identity", ".aet registry, resolver, reverse lookup, proofs.", "Identity Zone", []string{"x/proofregistry", "x/msgbus"}, "Cross-zone identity lookup returns proof-backed async replies."),
		cosmosModule("x/payments", "Payment channels, conditional settlement, routes.", "Financial Zone", []string{"x/msgbus", "x/proofregistry", "x/zonefees"}, "Channels and conditions settle with value-conserving receipts and proofs."),
		cosmosModule("x/scheduler", "Async jobs, finalization queues, expiry processing.", "Application Zone", []string{"x/msgbus", "x/zones"}, "Due tasks, retries, expiries, and callbacks execute in deterministic order."),
		cosmosModule("x/zonefees", "Zone-local fee accounting and aggregation.", "Core + zones", []string{"x/aetracore", "x/payments"}, "Per-shard fees aggregate into zone and global fee roots."),
		cosmosModule("x/zonemempool", "Mempool classification and transaction grouping.", "Node-side + core checks", []string{"x/aetracore", "x/shards"}, "Proposal grouping is deterministic and validators reject malformed schedules."),
	}
}

func BuildCosmosModuleMap(modules []CosmosModuleDescriptor) (CosmosModuleMap, error) {
	normalized := normalizeCosmosModuleDescriptors(modules)
	moduleMap := CosmosModuleMap{
		Version:	CosmosModuleMapVersion,
		Modules:	normalized,
	}
	if err := moduleMap.ValidateFormat(); err != nil {
		return CosmosModuleMap{}, err
	}
	moduleMap.Root = ComputeCosmosModuleMapRoot(moduleMap.Modules)
	return moduleMap, moduleMap.Validate()
}

func DefaultCosmosModuleMap() (CosmosModuleMap, error) {
	return BuildCosmosModuleMap(CosmosSDKNewModules())
}

func (m CosmosModuleMap) Normalize() CosmosModuleMap {
	if m.Version == 0 {
		m.Version = CosmosModuleMapVersion
	}
	m.Modules = normalizeCosmosModuleDescriptors(m.Modules)
	m.Root = strings.ToLower(strings.TrimSpace(m.Root))
	return m
}

func (m CosmosModuleMap) ValidateFormat() error {
	m = m.Normalize()
	if m.Version != CosmosModuleMapVersion {
		return fmt.Errorf("aetracore cosmos module map version must be %d", CosmosModuleMapVersion)
	}
	if len(m.Modules) == 0 {
		return errors.New("aetracore cosmos module map requires modules")
	}
	seen := make(map[string]struct{}, len(m.Modules))
	previous := ""
	for _, module := range m.Modules {
		if err := module.Validate(); err != nil {
			return err
		}
		if _, found := seen[module.Module]; found {
			return fmt.Errorf("aetracore duplicate cosmos module %s", module.Module)
		}
		seen[module.Module] = struct{}{}
		if previous != "" && previous >= module.Module {
			return errors.New("aetracore cosmos module map must be sorted by module")
		}
		previous = module.Module
	}
	if m.Root != "" {
		if err := ValidateHash("aetracore cosmos module map root", m.Root); err != nil {
			return err
		}
	}
	return nil
}

func (m CosmosModuleMap) Validate() error {
	m = m.Normalize()
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.Root == "" {
		return errors.New("aetracore cosmos module map root is required")
	}
	expected := ComputeCosmosModuleMapRoot(m.Modules)
	if m.Root != expected {
		return fmt.Errorf("aetracore cosmos module map root mismatch: expected %s", expected)
	}
	return nil
}

func BuildCosmosModuleDescriptor(desc CosmosModuleDescriptor) (CosmosModuleDescriptor, error) {
	desc = desc.Normalize()
	if desc.DescriptorHash != "" {
		return CosmosModuleDescriptor{}, errors.New("aetracore cosmos module descriptor hash must be empty before construction")
	}
	if err := desc.ValidateFormat(); err != nil {
		return CosmosModuleDescriptor{}, err
	}
	desc.DescriptorHash = ComputeCosmosModuleDescriptorHash(desc)
	return desc, desc.Validate()
}

func (desc CosmosModuleDescriptor) Normalize() CosmosModuleDescriptor {
	desc.Module = strings.TrimSpace(desc.Module)
	desc.Responsibility = normalizeModuleMapText(desc.Responsibility)
	desc.Zone = normalizeModuleMapText(desc.Zone)
	for i := range desc.Dependencies {
		desc.Dependencies[i] = normalizeModuleMapText(desc.Dependencies[i])
	}
	desc.Dependencies = compactSortedStrings(desc.Dependencies)
	desc.AcceptanceSignal = normalizeModuleMapText(desc.AcceptanceSignal)
	desc.DescriptorHash = strings.ToLower(strings.TrimSpace(desc.DescriptorHash))
	return desc
}

func (desc CosmosModuleDescriptor) ValidateFormat() error {
	desc = desc.Normalize()
	if !strings.HasPrefix(desc.Module, "x/") {
		return errors.New("aetracore cosmos module name must start with x/")
	}
	if err := validateToken("aetracore cosmos module name", desc.Module, MaxScopeLength); err != nil {
		return err
	}
	if desc.Responsibility == "" {
		return errors.New("aetracore cosmos module responsibility is required")
	}
	if desc.Zone == "" {
		return errors.New("aetracore cosmos module zone is required")
	}
	if len(desc.Dependencies) == 0 {
		return errors.New("aetracore cosmos module dependencies are required")
	}
	if desc.AcceptanceSignal == "" {
		return errors.New("aetracore cosmos module acceptance signal is required")
	}
	seenDeps := make(map[string]struct{}, len(desc.Dependencies))
	previous := ""
	for _, dep := range desc.Dependencies {
		if dep == "" {
			return errors.New("aetracore cosmos module dependency is required")
		}
		if _, found := seenDeps[dep]; found {
			return fmt.Errorf("aetracore duplicate cosmos module dependency %s", dep)
		}
		seenDeps[dep] = struct{}{}
		if previous != "" && previous >= dep {
			return errors.New("aetracore cosmos module dependencies must be sorted")
		}
		previous = dep
	}
	if desc.DescriptorHash != "" {
		if err := ValidateHash("aetracore cosmos module descriptor hash", desc.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (desc CosmosModuleDescriptor) Validate() error {
	desc = desc.Normalize()
	if err := desc.ValidateFormat(); err != nil {
		return err
	}
	if desc.DescriptorHash == "" {
		return errors.New("aetracore cosmos module descriptor hash is required")
	}
	expected := ComputeCosmosModuleDescriptorHash(desc)
	if desc.DescriptorHash != expected {
		return fmt.Errorf("aetracore cosmos module descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeCosmosModuleDescriptorHash(desc CosmosModuleDescriptor) string {
	desc = desc.Normalize()
	parts := []string{
		"aetra-cosmos-module-descriptor-v1",
		desc.Module,
		desc.Responsibility,
		desc.Zone,
		desc.AcceptanceSignal,
	}
	parts = append(parts, desc.Dependencies...)
	return hashParts(parts...)
}

func ComputeCosmosModuleMapRoot(modules []CosmosModuleDescriptor) string {
	normalized := normalizeCosmosModuleDescriptors(modules)
	parts := []string{"aetra-cosmos-module-map-v1", fmt.Sprintf("%020d", CosmosModuleMapVersion)}
	for _, module := range normalized {
		parts = append(parts, module.Module, module.DescriptorHash)
	}
	return hashParts(parts...)
}

func ValidateCosmosSDKNewModuleMap() error {
	moduleMap, err := DefaultCosmosModuleMap()
	if err != nil {
		return err
	}
	required := []string{
		"x/aetracore",
		"x/msgbus",
		"x/zones",
		"x/shards",
		"x/proofregistry",
		"x/avm",
		"x/identity",
		"x/payments",
		"x/scheduler",
		"x/zonefees",
		"x/zonemempool",
	}
	byModule := make(map[string]CosmosModuleDescriptor, len(moduleMap.Modules))
	for _, module := range moduleMap.Modules {
		byModule[module.Module] = module
	}
	for _, moduleName := range required {
		if _, found := byModule[moduleName]; !found {
			return fmt.Errorf("aetracore cosmos module map missing %s", moduleName)
		}
	}
	return nil
}

func cosmosModule(module, responsibility, zone string, dependencies []string, acceptance string) CosmosModuleDescriptor {
	desc, err := BuildCosmosModuleDescriptor(CosmosModuleDescriptor{
		Module:			module,
		Responsibility:		responsibility,
		Zone:			zone,
		Dependencies:		dependencies,
		AcceptanceSignal:	acceptance,
	})
	if err != nil {
		panic(err)
	}
	return desc
}

func normalizeCosmosModuleDescriptors(modules []CosmosModuleDescriptor) []CosmosModuleDescriptor {
	out := make([]CosmosModuleDescriptor, len(modules))
	for i, module := range modules {
		normalized := module.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeCosmosModuleDescriptorHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Module < out[j].Module
	})
	return out
}

func normalizeModuleMapText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func compactSortedStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = normalizeModuleMapText(value)
		if value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
