package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type IdentityV2Component string

const (
	IdentityV2Root			IdentityV2Component	= "x/identity v2"
	IdentityV2Core			IdentityV2Component	= "identity core"
	IdentityV2Resolver		IdentityV2Component	= "resolver module"
	IdentityV2Subdomain		IdentityV2Component	= "subdomain module"
	IdentityV2Auction		IdentityV2Component	= "auction module"
	IdentityV2Proof			IdentityV2Component	= "proof module"
	IdentityV2RoutingIntegration	IdentityV2Component	= "routing integration"
)

type IdentityV2ModuleSpec struct {
	Component	IdentityV2Component
	Parent		IdentityV2Component
	DependsOn	[]IdentityV2Component
	OwnsPrefixes	[]string
	ProofProducer	bool
	RoutingConsumer	bool
}

type IdentityV2Architecture struct {
	Root	IdentityV2Component
	Modules	[]IdentityV2ModuleSpec
}

func DefaultIdentityV2Architecture() IdentityV2Architecture {
	return IdentityV2Architecture{
		Root:	IdentityV2Root,
		Modules: []IdentityV2ModuleSpec{
			{
				Component:	IdentityV2Core,
				Parent:		IdentityV2Root,
				OwnsPrefixes:	[]string{IdentityStoreV2DomainPrefix, IdentityStoreV2NFTPrefix, IdentityStoreV2OwnerIndexPrefix, IdentityStoreV2CommitPrefix},
				ProofProducer:	true,
			},
			{
				Component:	IdentityV2Resolver,
				Parent:		IdentityV2Root,
				DependsOn:	[]IdentityV2Component{IdentityV2Core},
				OwnsPrefixes:	[]string{IdentityStoreV2ResolverPrefix, IdentityStoreV2ReversePrefix, IdentityStoreV2PendingResolverPrefix},
				ProofProducer:	true,
			},
			{
				Component:	IdentityV2Subdomain,
				Parent:		IdentityV2Root,
				DependsOn:	[]IdentityV2Component{IdentityV2Core, IdentityV2Resolver},
				OwnsPrefixes:	[]string{IdentityStoreV2SubdomainIndexPrefix},
				ProofProducer:	true,
			},
			{
				Component:	IdentityV2Auction,
				Parent:		IdentityV2Root,
				DependsOn:	[]IdentityV2Component{IdentityV2Core},
				OwnsPrefixes:	[]string{IdentityStoreV2AuctionPrefix},
				ProofProducer:	true,
			},
			{
				Component:	IdentityV2Proof,
				Parent:		IdentityV2Root,
				DependsOn:	[]IdentityV2Component{IdentityV2Core, IdentityV2Resolver, IdentityV2Subdomain, IdentityV2Auction},
			},
			{
				Component:		IdentityV2RoutingIntegration,
				Parent:			IdentityV2Root,
				DependsOn:		[]IdentityV2Component{IdentityV2Proof, IdentityV2Resolver},
				RoutingConsumer:	true,
			},
		},
	}
}

func ValidateIdentityV2Architecture(arch IdentityV2Architecture) error {
	if arch.Root != IdentityV2Root {
		return fmt.Errorf("identity architecture root must be %q", IdentityV2Root)
	}
	if len(arch.Modules) == 0 {
		return errors.New("identity architecture modules are required")
	}
	byID := make(map[IdentityV2Component]IdentityV2ModuleSpec, len(arch.Modules))
	for _, module := range arch.Modules {
		if module.Component == "" {
			return errors.New("identity architecture module component is required")
		}
		if module.Component == arch.Root {
			return errors.New("identity architecture root cannot be a child module")
		}
		if module.Parent != arch.Root {
			return fmt.Errorf("identity architecture module %q must be parented by root", module.Component)
		}
		if _, found := byID[module.Component]; found {
			return fmt.Errorf("duplicate identity architecture module %q", module.Component)
		}
		byID[module.Component] = cloneIdentityV2ModuleSpec(module)
	}
	for _, required := range requiredIdentityV2Components() {
		if _, found := byID[required]; !found {
			return fmt.Errorf("identity architecture missing module %q", required)
		}
	}
	if err := validateIdentityV2Dependencies(byID); err != nil {
		return err
	}
	if err := validateIdentityV2PrefixOwnership(arch.Modules); err != nil {
		return err
	}
	return validateIdentityV2SemanticEdges(byID)
}

func IdentityV2ExecutionOrder(arch IdentityV2Architecture) ([]IdentityV2Component, error) {
	if err := ValidateIdentityV2Architecture(arch); err != nil {
		return nil, err
	}
	byID := make(map[IdentityV2Component]IdentityV2ModuleSpec, len(arch.Modules))
	for _, module := range arch.Modules {
		byID[module.Component] = module
	}
	visited := map[IdentityV2Component]bool{}
	order := make([]IdentityV2Component, 0, len(arch.Modules))
	for _, component := range requiredIdentityV2Components() {
		visitIdentityV2Component(component, byID, visited, &order)
	}
	return order, nil
}

func IdentityV2ComponentForStoreKey(arch IdentityV2Architecture, key string) (IdentityV2Component, bool, error) {
	if err := ValidateIdentityV2Architecture(arch); err != nil {
		return "", false, err
	}
	if key == "" {
		return "", false, errors.New("identity store key is required")
	}
	var owner IdentityV2Component
	longest := -1
	for _, module := range arch.Modules {
		for _, prefix := range module.OwnsPrefixes {
			if key == prefix || strings.HasPrefix(key, prefix+"/") {
				if len(prefix) > longest {
					owner = module.Component
					longest = len(prefix)
				}
			}
		}
	}
	return owner, longest >= 0, nil
}

func IdentityV2ArchitectureHash(arch IdentityV2Architecture) (string, error) {
	if err := ValidateIdentityV2Architecture(arch); err != nil {
		return "", err
	}
	order, err := IdentityV2ExecutionOrder(arch)
	if err != nil {
		return "", err
	}
	byID := make(map[IdentityV2Component]IdentityV2ModuleSpec, len(arch.Modules))
	for _, module := range arch.Modules {
		byID[module.Component] = module
	}
	parts := []string{"identity-v2-architecture", string(arch.Root)}
	for _, component := range order {
		module := byID[component]
		parts = append(parts, string(module.Component), string(module.Parent))
		for _, dependency := range sortIdentityV2Components(module.DependsOn) {
			parts = append(parts, "dep", string(dependency))
		}
		for _, prefix := range sortedUniqueStrings(module.OwnsPrefixes) {
			parts = append(parts, "prefix", prefix)
		}
		parts = append(parts, fmt.Sprintf("proof=%t", module.ProofProducer), fmt.Sprintf("routing=%t", module.RoutingConsumer))
	}
	return identityHash(parts...), nil
}

func validateIdentityV2Dependencies(modules map[IdentityV2Component]IdentityV2ModuleSpec) error {
	visiting := map[IdentityV2Component]bool{}
	visited := map[IdentityV2Component]bool{}
	var walk func(IdentityV2Component) error
	walk = func(component IdentityV2Component) error {
		if visited[component] {
			return nil
		}
		if visiting[component] {
			return fmt.Errorf("identity architecture dependency cycle at %q", component)
		}
		module, found := modules[component]
		if !found {
			return fmt.Errorf("identity architecture dependency %q is not declared", component)
		}
		visiting[component] = true
		for _, dependency := range module.DependsOn {
			if dependency == component {
				return fmt.Errorf("identity architecture module %q depends on itself", component)
			}
			if err := walk(dependency); err != nil {
				return err
			}
		}
		visiting[component] = false
		visited[component] = true
		return nil
	}
	for component := range modules {
		if err := walk(component); err != nil {
			return err
		}
	}
	return nil
}

func validateIdentityV2PrefixOwnership(modules []IdentityV2ModuleSpec) error {
	owners := map[string]IdentityV2Component{}
	for _, module := range modules {
		for _, prefix := range module.OwnsPrefixes {
			if strings.TrimSpace(prefix) == "" {
				return fmt.Errorf("identity architecture module %q owns an empty prefix", module.Component)
			}
			if !strings.HasPrefix(prefix, IdentityStoreV2Prefix) {
				return fmt.Errorf("identity architecture prefix %q must be below %q", prefix, IdentityStoreV2Prefix)
			}
			if owner, found := owners[prefix]; found {
				return fmt.Errorf("identity architecture prefix %q owned by both %q and %q", prefix, owner, module.Component)
			}
			owners[prefix] = module.Component
		}
	}
	return nil
}

func validateIdentityV2SemanticEdges(modules map[IdentityV2Component]IdentityV2ModuleSpec) error {
	proof := modules[IdentityV2Proof]
	for _, dependency := range []IdentityV2Component{IdentityV2Core, IdentityV2Resolver, IdentityV2Subdomain, IdentityV2Auction} {
		if !identityV2HasDependency(proof, dependency) {
			return fmt.Errorf("identity proof module must depend on %q", dependency)
		}
	}
	routing := modules[IdentityV2RoutingIntegration]
	for _, dependency := range []IdentityV2Component{IdentityV2Proof, IdentityV2Resolver} {
		if !identityV2HasDependency(routing, dependency) {
			return fmt.Errorf("identity routing integration must depend on %q", dependency)
		}
	}
	for _, producer := range []IdentityV2Component{IdentityV2Core, IdentityV2Resolver, IdentityV2Subdomain, IdentityV2Auction} {
		if !modules[producer].ProofProducer {
			return fmt.Errorf("identity module %q must produce proof leaves", producer)
		}
	}
	if !routing.RoutingConsumer {
		return errors.New("identity routing integration must consume resolver output")
	}
	return nil
}

func visitIdentityV2Component(component IdentityV2Component, modules map[IdentityV2Component]IdentityV2ModuleSpec, visited map[IdentityV2Component]bool, order *[]IdentityV2Component) {
	if visited[component] {
		return
	}
	module := modules[component]
	for _, dependency := range sortIdentityV2Components(module.DependsOn) {
		visitIdentityV2Component(dependency, modules, visited, order)
	}
	visited[component] = true
	*order = append(*order, component)
}

func requiredIdentityV2Components() []IdentityV2Component {
	return []IdentityV2Component{
		IdentityV2Core,
		IdentityV2Resolver,
		IdentityV2Subdomain,
		IdentityV2Auction,
		IdentityV2Proof,
		IdentityV2RoutingIntegration,
	}
}

func sortIdentityV2Components(values []IdentityV2Component) []IdentityV2Component {
	out := append([]IdentityV2Component(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func identityV2HasDependency(module IdentityV2ModuleSpec, dependency IdentityV2Component) bool {
	for _, candidate := range module.DependsOn {
		if candidate == dependency {
			return true
		}
	}
	return false
}

func cloneIdentityV2ModuleSpec(module IdentityV2ModuleSpec) IdentityV2ModuleSpec {
	module.DependsOn = append([]IdentityV2Component(nil), module.DependsOn...)
	module.OwnsPrefixes = append([]string(nil), module.OwnsPrefixes...)
	return module
}
