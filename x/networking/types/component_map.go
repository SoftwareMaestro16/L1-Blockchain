package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NodeSideComponent string

const (
	ComponentANA		NodeSideComponent	= "ana"
	ComponentSessionMgr	NodeSideComponent	= "sessionmgr"
	ComponentOverlayMgr	NodeSideComponent	= "overlaymgr"
	ComponentDRT		NodeSideComponent	= "drt"
	ComponentRL2		NodeSideComponent	= "rl2"
	ComponentMesh		NodeSideComponent	= "mesh"
	ComponentBroadcast	NodeSideComponent	= "broadcast"
)

type OnChainSupportModule string

const (
	SupportModuleNetwork	OnChainSupportModule	= "x/network"
	SupportModuleRouting	OnChainSupportModule	= "x/routing"
	SupportModuleServices	OnChainSupportModule	= "x/services"
	SupportModuleStorage	OnChainSupportModule	= "x/storage"
	SupportModuleMessages	OnChainSupportModule	= "x/messages"
)

type NodeSideComponentSpec struct {
	Component		NodeSideComponent
	Layer			NetworkLayer
	Channels		[]ChannelClass
	Responsibilities	[]string
	DependsOn		[]NodeSideComponent
	RuntimeOnly		bool
	AdvisoryUntilCommitted	bool
	WritesCommittedState	bool
	ExtendsCometBFT		bool
}

type OnChainSupportModuleSpec struct {
	Module				OnChainSupportModule
	OwnsCommittedState		bool
	StateObjects			[]string
	ConsumesNetworkProofs		bool
	AllowsExternalNetworkCall	bool
	Optional			bool
}

type NetworkingComponentMap struct {
	NodeComponents	[]NodeSideComponentSpec
	SupportModules	[]OnChainSupportModuleSpec
	MapRoot		string
}

func DefaultNetworkingComponentMap() NetworkingComponentMap {
	componentMap := NetworkingComponentMap{
		NodeComponents: []NodeSideComponentSpec{
			{
				Component:		ComponentANA,
				Layer:			LayerL0Physical,
				Channels:		[]ChannelClass{ChannelConsensus, ChannelMempool, ChannelBlock, ChannelStateSync, ChannelData, ChannelExecution, ChannelService, ChannelRouting, ChannelDiscovery},
				Responsibilities:	[]string{"Aether Networking Adapter", "peer scoring", "channel prioritization", "adaptive fanout"},
				RuntimeOnly:		true,
				AdvisoryUntilCommitted:	true,
				ExtendsCometBFT:	true,
			},
			{
				Component:		ComponentSessionMgr,
				Layer:			LayerL1Session,
				Channels:		[]ChannelClass{ChannelConsensus, ChannelBlock, ChannelStateSync, ChannelExecution, ChannelService, ChannelRouting, ChannelDiscovery, ChannelData},
				Responsibilities:	[]string{"node identity", "handshake", "session keys", "stream multiplexing"},
				DependsOn:		[]NodeSideComponent{ComponentANA},
				RuntimeOnly:		true,
				AdvisoryUntilCommitted:	true,
				ExtendsCometBFT:	true,
			},
			{
				Component:		ComponentOverlayMgr,
				Layer:			LayerL2Overlay,
				Channels:		[]ChannelClass{ChannelRouting, ChannelExecution, ChannelService, ChannelData},
				Responsibilities:	[]string{"overlay membership", "peer sets", "route graph building"},
				DependsOn:		[]NodeSideComponent{ComponentSessionMgr, ComponentDRT},
				RuntimeOnly:		true,
				AdvisoryUntilCommitted:	true,
				ExtendsCometBFT:	true,
			},
			{
				Component:		ComponentDRT,
				Layer:			LayerL2Overlay,
				Channels:		[]ChannelClass{ChannelDiscovery, ChannelRouting, ChannelService},
				Responsibilities:	[]string{"distributed routing table", "lease advertisements", "proof-attached responses"},
				DependsOn:		[]NodeSideComponent{ComponentSessionMgr},
				RuntimeOnly:		true,
				AdvisoryUntilCommitted:	true,
				ExtendsCometBFT:	true,
			},
			{
				Component:		ComponentRL2,
				Layer:			LayerL2Overlay,
				Channels:		[]ChannelClass{ChannelBlock, ChannelStateSync, ChannelData, ChannelExecution},
				Responsibilities:	[]string{"reliable chunked transport", "resumable transfers", "Merkle verification"},
				DependsOn:		[]NodeSideComponent{ComponentSessionMgr},
				RuntimeOnly:		true,
				AdvisoryUntilCommitted:	true,
				ExtendsCometBFT:	true,
			},
			{
				Component:		ComponentMesh,
				Layer:			LayerL3Application,
				Channels:		[]ChannelClass{ChannelExecution, ChannelService, ChannelData, ChannelDiscovery},
				Responsibilities:	[]string{"application networking", "service flow", "cross-zone messages", "receipts"},
				DependsOn:		[]NodeSideComponent{ComponentOverlayMgr, ComponentRL2},
				RuntimeOnly:		true,
				AdvisoryUntilCommitted:	true,
				ExtendsCometBFT:	true,
			},
			{
				Component:		ComponentBroadcast,
				Layer:			LayerL0Physical,
				Channels:		[]ChannelClass{ChannelBlock, ChannelData, ChannelService, ChannelRouting},
				Responsibilities:	[]string{"hybrid gossip tree", "deduplication", "header-first block propagation"},
				DependsOn:		[]NodeSideComponent{ComponentANA, ComponentOverlayMgr, ComponentRL2},
				RuntimeOnly:		true,
				AdvisoryUntilCommitted:	true,
				ExtendsCometBFT:	true,
			},
		},
		SupportModules: []OnChainSupportModuleSpec{
			{
				Module:			SupportModuleNetwork,
				OwnsCommittedState:	true,
				StateObjects:		[]string{"committed network parameters", "node records"},
				Optional:		true,
			},
			{
				Module:			SupportModuleRouting,
				OwnsCommittedState:	true,
				StateObjects:		[]string{"routing table commitments", "overlay descriptors"},
				ConsumesNetworkProofs:	true,
				Optional:		true,
			},
			{
				Module:			SupportModuleServices,
				OwnsCommittedState:	true,
				StateObjects:		[]string{"service endpoint records", "provider records"},
				ConsumesNetworkProofs:	true,
				Optional:		true,
			},
			{
				Module:			SupportModuleStorage,
				OwnsCommittedState:	true,
				StateObjects:		[]string{"storage provider commitments"},
				ConsumesNetworkProofs:	true,
				Optional:		true,
			},
			{
				Module:			SupportModuleMessages,
				OwnsCommittedState:	true,
				StateObjects:		[]string{"cross-zone message receipts", "replay protection"},
				ConsumesNetworkProofs:	true,
				Optional:		true,
			},
		},
	}
	componentMap.MapRoot = ComputeNetworkingComponentMapRoot(componentMap)
	return componentMap
}

func ValidateNetworkingComponentMap(componentMap NetworkingComponentMap) error {
	componentMap = NormalizeNetworkingComponentMap(componentMap)
	if len(componentMap.NodeComponents) == 0 {
		return errors.New("networking component map requires node-side components")
	}
	if len(componentMap.SupportModules) == 0 {
		return errors.New("networking component map requires support modules")
	}
	if componentMap.MapRoot != ComputeNetworkingComponentMapRoot(componentMap) {
		return errors.New("networking component map root mismatch")
	}
	components := make(map[NodeSideComponent]NodeSideComponentSpec, len(componentMap.NodeComponents))
	for _, component := range componentMap.NodeComponents {
		if err := component.Validate(); err != nil {
			return err
		}
		if _, found := components[component.Component]; found {
			return errors.New("networking duplicate node-side component")
		}
		components[component.Component] = component
	}
	for _, required := range requiredNodeSideComponents() {
		if _, found := components[required]; !found {
			return fmt.Errorf("networking component map missing node-side component %s", required)
		}
	}
	for _, component := range componentMap.NodeComponents {
		for _, dependency := range component.DependsOn {
			if _, found := components[dependency]; !found {
				return fmt.Errorf("networking component %s missing dependency %s", component.Component, dependency)
			}
		}
	}
	modules := make(map[OnChainSupportModule]struct{}, len(componentMap.SupportModules))
	for _, module := range componentMap.SupportModules {
		if err := module.Validate(); err != nil {
			return err
		}
		if _, found := modules[module.Module]; found {
			return errors.New("networking duplicate support module")
		}
		modules[module.Module] = struct{}{}
	}
	for _, required := range requiredSupportModules() {
		if _, found := modules[required]; !found {
			return fmt.Errorf("networking component map missing support module %s", required)
		}
	}
	return nil
}

func NormalizeNetworkingComponentMap(componentMap NetworkingComponentMap) NetworkingComponentMap {
	for i := range componentMap.NodeComponents {
		componentMap.NodeComponents[i] = NormalizeNodeSideComponentSpec(componentMap.NodeComponents[i])
	}
	sortNodeSideComponentSpecs(componentMap.NodeComponents)
	for i := range componentMap.SupportModules {
		componentMap.SupportModules[i] = NormalizeOnChainSupportModuleSpec(componentMap.SupportModules[i])
	}
	sortOnChainSupportModuleSpecs(componentMap.SupportModules)
	componentMap.MapRoot = normalizeHashText(componentMap.MapRoot)
	return componentMap
}

func NormalizeNodeSideComponentSpec(component NodeSideComponentSpec) NodeSideComponentSpec {
	component.Component = NodeSideComponent(strings.ToLower(strings.TrimSpace(string(component.Component))))
	component.Responsibilities = normalizeFreeformSet(component.Responsibilities)
	component.DependsOn = normalizeNodeSideComponentSet(component.DependsOn)
	component.Channels = normalizeChannels(component.Channels)
	return component
}

func NormalizeOnChainSupportModuleSpec(module OnChainSupportModuleSpec) OnChainSupportModuleSpec {
	module.Module = OnChainSupportModule(strings.ToLower(strings.TrimSpace(string(module.Module))))
	module.StateObjects = normalizeFreeformSet(module.StateObjects)
	return module
}

func (c NodeSideComponentSpec) Validate() error {
	component := NormalizeNodeSideComponentSpec(c)
	if !IsNodeSideComponent(component.Component) {
		return fmt.Errorf("unknown networking node-side component %q", component.Component)
	}
	if !IsNetworkLayer(component.Layer) {
		return fmt.Errorf("unknown networking component layer %q", component.Layer)
	}
	if len(component.Channels) == 0 {
		return errors.New("networking node-side component requires channels")
	}
	if err := validateChannels(component.Channels); err != nil {
		return err
	}
	if len(component.Responsibilities) == 0 {
		return errors.New("networking node-side component requires responsibilities")
	}
	if !component.RuntimeOnly {
		return errors.New("networking node-side component must be runtime-only")
	}
	if component.WritesCommittedState {
		return errors.New("networking node-side component must not write committed state directly")
	}
	if !component.AdvisoryUntilCommitted {
		return errors.New("networking node-side component outputs are advisory until committed")
	}
	if !component.ExtendsCometBFT {
		return errors.New("networking node-side component must extend CometBFT baseline")
	}
	return validateComponentLayerContract(component)
}

func (m OnChainSupportModuleSpec) Validate() error {
	module := NormalizeOnChainSupportModuleSpec(m)
	if !IsOnChainSupportModule(module.Module) {
		return fmt.Errorf("unknown networking support module %q", module.Module)
	}
	if !module.OwnsCommittedState {
		return errors.New("networking support module must own committed state")
	}
	if len(module.StateObjects) == 0 {
		return errors.New("networking support module requires state objects")
	}
	if module.AllowsExternalNetworkCall {
		return errors.New("networking support module must not allow external network calls")
	}
	return nil
}

func ComponentByName(componentMap NetworkingComponentMap, name NodeSideComponent) (NodeSideComponentSpec, bool, error) {
	componentMap = NormalizeNetworkingComponentMap(componentMap)
	if err := ValidateNetworkingComponentMap(componentMap); err != nil {
		return NodeSideComponentSpec{}, false, err
	}
	name = NodeSideComponent(strings.ToLower(strings.TrimSpace(string(name))))
	for _, component := range componentMap.NodeComponents {
		if component.Component == name {
			return component, true, nil
		}
	}
	return NodeSideComponentSpec{}, false, nil
}

func SupportModuleByName(componentMap NetworkingComponentMap, name OnChainSupportModule) (OnChainSupportModuleSpec, bool, error) {
	componentMap = NormalizeNetworkingComponentMap(componentMap)
	if err := ValidateNetworkingComponentMap(componentMap); err != nil {
		return OnChainSupportModuleSpec{}, false, err
	}
	name = OnChainSupportModule(strings.ToLower(strings.TrimSpace(string(name))))
	for _, module := range componentMap.SupportModules {
		if module.Module == name {
			return module, true, nil
		}
	}
	return OnChainSupportModuleSpec{}, false, nil
}

func ComputeNetworkingComponentMapRoot(componentMap NetworkingComponentMap) string {
	componentMap = NormalizeNetworkingComponentMap(componentMap)
	parts := []string{"networking-component-map"}
	for _, component := range componentMap.NodeComponents {
		parts = append(parts,
			string(component.Component),
			string(component.Layer),
			fmt.Sprintf("%t", component.RuntimeOnly),
			fmt.Sprintf("%t", component.AdvisoryUntilCommitted),
			fmt.Sprintf("%t", component.WritesCommittedState),
			fmt.Sprintf("%t", component.ExtendsCometBFT),
		)
		for _, channel := range component.Channels {
			parts = append(parts, string(channel))
		}
		parts = append(parts, component.Responsibilities...)
		for _, dependency := range component.DependsOn {
			parts = append(parts, string(dependency))
		}
	}
	for _, module := range componentMap.SupportModules {
		parts = append(parts,
			string(module.Module),
			fmt.Sprintf("%t", module.OwnsCommittedState),
			fmt.Sprintf("%t", module.ConsumesNetworkProofs),
			fmt.Sprintf("%t", module.AllowsExternalNetworkCall),
			fmt.Sprintf("%t", module.Optional),
		)
		parts = append(parts, module.StateObjects...)
	}
	return HashParts(parts...)
}

func IsNodeSideComponent(component NodeSideComponent) bool {
	switch component {
	case ComponentANA,
		ComponentSessionMgr,
		ComponentOverlayMgr,
		ComponentDRT,
		ComponentRL2,
		ComponentMesh,
		ComponentBroadcast:
		return true
	default:
		return false
	}
}

func IsOnChainSupportModule(module OnChainSupportModule) bool {
	switch module {
	case SupportModuleNetwork,
		SupportModuleRouting,
		SupportModuleServices,
		SupportModuleStorage,
		SupportModuleMessages:
		return true
	default:
		return false
	}
}

func requiredNodeSideComponents() []NodeSideComponent {
	return []NodeSideComponent{
		ComponentANA,
		ComponentSessionMgr,
		ComponentOverlayMgr,
		ComponentDRT,
		ComponentRL2,
		ComponentMesh,
		ComponentBroadcast,
	}
}

func requiredSupportModules() []OnChainSupportModule {
	return []OnChainSupportModule{
		SupportModuleNetwork,
		SupportModuleRouting,
		SupportModuleServices,
		SupportModuleStorage,
		SupportModuleMessages,
	}
}

func validateComponentLayerContract(component NodeSideComponentSpec) error {
	switch component.Component {
	case ComponentANA, ComponentBroadcast:
		if component.Layer != LayerL0Physical {
			return fmt.Errorf("networking component %s must be mapped to L0", component.Component)
		}
	case ComponentSessionMgr:
		if component.Layer != LayerL1Session {
			return errors.New("networking sessionmgr must be mapped to L1")
		}
	case ComponentOverlayMgr, ComponentDRT, ComponentRL2:
		if component.Layer != LayerL2Overlay {
			return fmt.Errorf("networking component %s must be mapped to L2", component.Component)
		}
	case ComponentMesh:
		if component.Layer != LayerL3Application {
			return errors.New("networking mesh must be mapped to L3")
		}
	}
	return nil
}

func normalizeNodeSideComponentSet(values []NodeSideComponent) []NodeSideComponent {
	seen := make(map[NodeSideComponent]struct{}, len(values))
	out := make([]NodeSideComponent, 0, len(values))
	for _, value := range values {
		value = NodeSideComponent(strings.ToLower(strings.TrimSpace(string(value))))
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func normalizeFreeformSet(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func sortNodeSideComponentSpecs(components []NodeSideComponentSpec) {
	sort.SliceStable(components, func(i, j int) bool {
		return components[i].Component < components[j].Component
	})
}

func sortOnChainSupportModuleSpecs(modules []OnChainSupportModuleSpec) {
	sort.SliceStable(modules, func(i, j int) bool {
		return modules[i].Module < modules[j].Module
	})
}
