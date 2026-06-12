package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const CosmosModuleBoundarySpecVersion = uint64(1)

type ExistingCosmosModuleModification struct {
	Module			string
	RequiredModification	string
	Boundary		string
	DescriptorHash		string
}

type CosmosModuleBoundaryRule struct {
	Rule		string
	Enforcement	string
	DescriptorHash	string
}

type CosmosModuleBoundarySpec struct {
	Version		uint64
	Modifications	[]ExistingCosmosModuleModification
	BoundaryRules	[]CosmosModuleBoundaryRule
	Root		string
}

func ExistingCosmosModuleModifications() []ExistingCosmosModuleModification {
	return []ExistingCosmosModuleModification{
		existingCosmosModuleModification("bank", "Add zone-aware account routing, cross-shard escrow transfer flow, and message-driven transfer settlement.", "Balance state moves to Financial Zone prefixes and cross-zone effects use messages."),
		existingCosmosModuleModification("staking", "Expose validator set commitment to Aether Core, keep validator operations in core or Financial Zone depending on migration phase, and add validator metadata for zone service support where needed.", "Validator consensus state remains globally committed and deterministic."),
		existingCosmosModuleModification("slashing", "Preserve consensus slashing in core and keep payment fraud penalties separate from validator slashing.", "Validator safety penalties cannot be conflated with payment dispute penalties."),
		existingCosmosModuleModification("mint/distribution", "Aggregate zone fees into distribution flow and preserve deterministic reward accounting.", "Rewards derive from committed fee roots, not live zone-local observations."),
		existingCosmosModuleModification("fees", "Add zone-local fee policies, forwarding fee escrow, and congestion metrics by zone and shard.", "Fee policy remains deterministic and proof-backed."),
		existingCosmosModuleModification("contract-assets", "Add zone-aware denom authority and token routing rules for cross-zone messages.", "Denom authority is Financial Zone state and cross-zone mint/burn effects require messages."),
		existingCosmosModuleModification("avm-dex-contract", "Add pool shard placement, async swap flow for cross-shard routes, and pool proof queries.", "Pool state routes by committed pool ID and cross-shard swaps settle through receipts."),
	}
}

func CosmosModuleBoundaryRules() []CosmosModuleBoundaryRule {
	return []CosmosModuleBoundaryRule{
		cosmosModuleBoundaryRule("Core modules commit roots and schedule work.", "Core modules cannot own zone-local application state or contract storage."),
		cosmosModuleBoundaryRule("Zone modules own local state transitions.", "Zone writes are restricted to zone and shard prefixes."),
		cosmosModuleBoundaryRule("Message module connects zones and shards.", "Cross-zone and cross-shard effects must be represented as committed messages and receipts."),
		cosmosModuleBoundaryRule("Proof module verifies committed state only.", "Proof verification cannot read live mempool, external APIs, or uncommitted caches."),
		cosmosModuleBoundaryRule("VM module cannot mutate state outside Contract Zone except by message.", "AVM syscalls emit messages for remote effects and verify proofs for remote reads."),
		cosmosModuleBoundaryRule("Identity module cannot transfer funds except through Financial Zone messages.", "Identity ownership and resolver changes cannot directly debit or credit balances."),
		cosmosModuleBoundaryRule("Payments module cannot resolve names except through Identity Zone proof or message.", "Payment routing must use verified .aet resolution or explicit account addresses."),
	}
}

func BuildCosmosModuleBoundarySpec(modifications []ExistingCosmosModuleModification, rules []CosmosModuleBoundaryRule) (CosmosModuleBoundarySpec, error) {
	spec := CosmosModuleBoundarySpec{
		Version:	CosmosModuleBoundarySpecVersion,
		Modifications:	normalizeExistingCosmosModuleModifications(modifications),
		BoundaryRules:	normalizeCosmosModuleBoundaryRules(rules),
	}
	if err := spec.ValidateFormat(); err != nil {
		return CosmosModuleBoundarySpec{}, err
	}
	spec.Root = ComputeCosmosModuleBoundarySpecRoot(spec.Modifications, spec.BoundaryRules)
	return spec, spec.Validate()
}

func DefaultCosmosModuleBoundarySpec() (CosmosModuleBoundarySpec, error) {
	return BuildCosmosModuleBoundarySpec(ExistingCosmosModuleModifications(), CosmosModuleBoundaryRules())
}

func (spec CosmosModuleBoundarySpec) Normalize() CosmosModuleBoundarySpec {
	if spec.Version == 0 {
		spec.Version = CosmosModuleBoundarySpecVersion
	}
	spec.Modifications = normalizeExistingCosmosModuleModifications(spec.Modifications)
	spec.BoundaryRules = normalizeCosmosModuleBoundaryRules(spec.BoundaryRules)
	spec.Root = strings.ToLower(strings.TrimSpace(spec.Root))
	return spec
}

func (spec CosmosModuleBoundarySpec) ValidateFormat() error {
	spec = spec.Normalize()
	if spec.Version != CosmosModuleBoundarySpecVersion {
		return fmt.Errorf("aetracore cosmos module boundary spec version must be %d", CosmosModuleBoundarySpecVersion)
	}
	if len(spec.Modifications) == 0 {
		return errors.New("aetracore cosmos module boundary spec requires modifications")
	}
	if len(spec.BoundaryRules) == 0 {
		return errors.New("aetracore cosmos module boundary spec requires rules")
	}
	seenModules := make(map[string]struct{}, len(spec.Modifications))
	previousModule := ""
	for _, modification := range spec.Modifications {
		if err := modification.Validate(); err != nil {
			return err
		}
		if _, found := seenModules[modification.Module]; found {
			return fmt.Errorf("aetracore duplicate existing cosmos module %s", modification.Module)
		}
		seenModules[modification.Module] = struct{}{}
		if previousModule != "" && previousModule >= modification.Module {
			return errors.New("aetracore existing cosmos module modifications must be sorted")
		}
		previousModule = modification.Module
	}
	seenRules := make(map[string]struct{}, len(spec.BoundaryRules))
	previousRule := ""
	for _, rule := range spec.BoundaryRules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seenRules[rule.Rule]; found {
			return fmt.Errorf("aetracore duplicate cosmos module boundary rule %s", rule.Rule)
		}
		seenRules[rule.Rule] = struct{}{}
		if previousRule != "" && previousRule >= rule.Rule {
			return errors.New("aetracore cosmos module boundary rules must be sorted")
		}
		previousRule = rule.Rule
	}
	if spec.Root != "" {
		if err := ValidateHash("aetracore cosmos module boundary spec root", spec.Root); err != nil {
			return err
		}
	}
	return nil
}

func (spec CosmosModuleBoundarySpec) Validate() error {
	spec = spec.Normalize()
	if err := spec.ValidateFormat(); err != nil {
		return err
	}
	if spec.Root == "" {
		return errors.New("aetracore cosmos module boundary spec root is required")
	}
	expected := ComputeCosmosModuleBoundarySpecRoot(spec.Modifications, spec.BoundaryRules)
	if spec.Root != expected {
		return fmt.Errorf("aetracore cosmos module boundary spec root mismatch: expected %s", expected)
	}
	return nil
}

func BuildExistingCosmosModuleModification(modification ExistingCosmosModuleModification) (ExistingCosmosModuleModification, error) {
	modification = modification.Normalize()
	if modification.DescriptorHash != "" {
		return ExistingCosmosModuleModification{}, errors.New("aetracore existing cosmos module descriptor hash must be empty before construction")
	}
	if err := modification.ValidateFormat(); err != nil {
		return ExistingCosmosModuleModification{}, err
	}
	modification.DescriptorHash = ComputeExistingCosmosModuleModificationHash(modification)
	return modification, modification.Validate()
}

func (modification ExistingCosmosModuleModification) Normalize() ExistingCosmosModuleModification {
	modification.Module = strings.TrimSpace(modification.Module)
	modification.RequiredModification = normalizeModuleMapText(modification.RequiredModification)
	modification.Boundary = normalizeModuleMapText(modification.Boundary)
	modification.DescriptorHash = strings.ToLower(strings.TrimSpace(modification.DescriptorHash))
	return modification
}

func (modification ExistingCosmosModuleModification) ValidateFormat() error {
	modification = modification.Normalize()
	if err := validateToken("aetracore existing cosmos module name", modification.Module, MaxScopeLength); err != nil {
		return err
	}
	if modification.RequiredModification == "" {
		return errors.New("aetracore existing cosmos module modification is required")
	}
	if modification.Boundary == "" {
		return errors.New("aetracore existing cosmos module boundary is required")
	}
	if modification.DescriptorHash != "" {
		if err := ValidateHash("aetracore existing cosmos module descriptor hash", modification.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (modification ExistingCosmosModuleModification) Validate() error {
	modification = modification.Normalize()
	if err := modification.ValidateFormat(); err != nil {
		return err
	}
	if modification.DescriptorHash == "" {
		return errors.New("aetracore existing cosmos module descriptor hash is required")
	}
	expected := ComputeExistingCosmosModuleModificationHash(modification)
	if modification.DescriptorHash != expected {
		return fmt.Errorf("aetracore existing cosmos module descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildCosmosModuleBoundaryRule(rule CosmosModuleBoundaryRule) (CosmosModuleBoundaryRule, error) {
	rule = rule.Normalize()
	if rule.DescriptorHash != "" {
		return CosmosModuleBoundaryRule{}, errors.New("aetracore cosmos module boundary rule hash must be empty before construction")
	}
	if err := rule.ValidateFormat(); err != nil {
		return CosmosModuleBoundaryRule{}, err
	}
	rule.DescriptorHash = ComputeCosmosModuleBoundaryRuleHash(rule)
	return rule, rule.Validate()
}

func (rule CosmosModuleBoundaryRule) Normalize() CosmosModuleBoundaryRule {
	rule.Rule = normalizeModuleMapText(rule.Rule)
	rule.Enforcement = normalizeModuleMapText(rule.Enforcement)
	rule.DescriptorHash = strings.ToLower(strings.TrimSpace(rule.DescriptorHash))
	return rule
}

func (rule CosmosModuleBoundaryRule) ValidateFormat() error {
	rule = rule.Normalize()
	if rule.Rule == "" {
		return errors.New("aetracore cosmos module boundary rule is required")
	}
	if rule.Enforcement == "" {
		return errors.New("aetracore cosmos module boundary enforcement is required")
	}
	if rule.DescriptorHash != "" {
		if err := ValidateHash("aetracore cosmos module boundary rule hash", rule.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (rule CosmosModuleBoundaryRule) Validate() error {
	rule = rule.Normalize()
	if err := rule.ValidateFormat(); err != nil {
		return err
	}
	if rule.DescriptorHash == "" {
		return errors.New("aetracore cosmos module boundary rule hash is required")
	}
	expected := ComputeCosmosModuleBoundaryRuleHash(rule)
	if rule.DescriptorHash != expected {
		return fmt.Errorf("aetracore cosmos module boundary rule hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeExistingCosmosModuleModificationHash(modification ExistingCosmosModuleModification) string {
	modification = modification.Normalize()
	return hashParts(
		"aetra-existing-cosmos-module-modification-v1",
		modification.Module,
		modification.RequiredModification,
		modification.Boundary,
	)
}

func ComputeCosmosModuleBoundaryRuleHash(rule CosmosModuleBoundaryRule) string {
	rule = rule.Normalize()
	return hashParts(
		"aetra-cosmos-module-boundary-rule-v1",
		rule.Rule,
		rule.Enforcement,
	)
}

func ComputeCosmosModuleBoundarySpecRoot(modifications []ExistingCosmosModuleModification, rules []CosmosModuleBoundaryRule) string {
	normalizedModifications := normalizeExistingCosmosModuleModifications(modifications)
	normalizedRules := normalizeCosmosModuleBoundaryRules(rules)
	parts := []string{"aetra-cosmos-module-boundary-spec-v1", fmt.Sprintf("%020d", CosmosModuleBoundarySpecVersion)}
	for _, modification := range normalizedModifications {
		parts = append(parts, "modification", modification.Module, modification.DescriptorHash)
	}
	for _, rule := range normalizedRules {
		parts = append(parts, "rule", rule.Rule, rule.DescriptorHash)
	}
	return hashParts(parts...)
}

func ValidateCosmosModuleBoundarySpec() error {
	spec, err := DefaultCosmosModuleBoundarySpec()
	if err != nil {
		return err
	}
	requiredModules := []string{"bank", "staking", "slashing", "mint/distribution", "fees", "contract-assets", "avm-dex-contract"}
	byModule := make(map[string]ExistingCosmosModuleModification, len(spec.Modifications))
	for _, modification := range spec.Modifications {
		byModule[modification.Module] = modification
	}
	for _, module := range requiredModules {
		if _, found := byModule[module]; !found {
			return fmt.Errorf("aetracore cosmos module boundary spec missing %s", module)
		}
	}
	requiredRuleFragments := []string{"Core modules", "Zone modules", "Message module", "Proof module", "VM module", "Identity module", "Payments module"}
	for _, fragment := range requiredRuleFragments {
		found := false
		for _, rule := range spec.BoundaryRules {
			if strings.Contains(rule.Rule, fragment) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("aetracore cosmos module boundary spec missing rule %s", fragment)
		}
	}
	return nil
}

func existingCosmosModuleModification(module, requiredModification, boundary string) ExistingCosmosModuleModification {
	modification, err := BuildExistingCosmosModuleModification(ExistingCosmosModuleModification{
		Module:			module,
		RequiredModification:	requiredModification,
		Boundary:		boundary,
	})
	if err != nil {
		panic(err)
	}
	return modification
}

func cosmosModuleBoundaryRule(rule, enforcement string) CosmosModuleBoundaryRule {
	descriptor, err := BuildCosmosModuleBoundaryRule(CosmosModuleBoundaryRule{
		Rule:		rule,
		Enforcement:	enforcement,
	})
	if err != nil {
		panic(err)
	}
	return descriptor
}

func normalizeExistingCosmosModuleModifications(modifications []ExistingCosmosModuleModification) []ExistingCosmosModuleModification {
	out := make([]ExistingCosmosModuleModification, len(modifications))
	for i, modification := range modifications {
		normalized := modification.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeExistingCosmosModuleModificationHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Module < out[j].Module
	})
	return out
}

func normalizeCosmosModuleBoundaryRules(rules []CosmosModuleBoundaryRule) []CosmosModuleBoundaryRule {
	out := make([]CosmosModuleBoundaryRule, len(rules))
	for i, rule := range rules {
		normalized := rule.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeCosmosModuleBoundaryRuleHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Rule < out[j].Rule
	})
	return out
}
