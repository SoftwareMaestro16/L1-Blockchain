package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ComponentLayerCore	ComponentLayer	= "aetra_core"
	ComponentLayerRouter	ComponentLayer	= "avm_execution_router"
	ComponentLayerEngine	ComponentLayer	= "execution_engines"
	ComponentLayerBackend	ComponentLayer	= "contract_backends"
	ComponentLayerZoneState	ComponentLayer	= "zone_state"

	ComponentAetraCore	ComponentID	= "aetra_core"
	ComponentAVMRouter	ComponentID	= "avm_execution_router"
	ComponentSyncEngine	ComponentID	= "sync_engine"
	ComponentAsyncEngine	ComponentID	= "async_engine"
	ComponentActorRuntime	ComponentID	= "actor_runtime"
	ComponentNativeModules	ComponentID	= "native_modules"
	ComponentAVMActors	ComponentID	= "avm_actors"
	ComponentWASMAdapter	ComponentID	= "wasm_adapter"
	ComponentZoneStateRoots	ComponentID	= "zone_state_store_v2_roots"

	ComponentCapabilityConsensus		= "consensus"
	ComponentCapabilityBlockLifecycle	= "block_lifecycle"
	ComponentCapabilityGlobalRoots		= "global_roots"
	ComponentCapabilityZoneRegistry		= "zone_registry"
	ComponentCapabilityRouteMessages	= "route_tx_msg"
	ComponentCapabilityClassifyZone		= "classify_zone"
	ComponentCapabilityValidateBudget	= "validate_budget"
	ComponentCapabilityDispatch		= "dispatch"
	ComponentCapabilityMsgServer		= "msg_server"
	ComponentCapabilityKeeperCalls		= "keeper_calls"
	ComponentCapabilityQueues		= "queues"
	ComponentCapabilityScheduling		= "scheduling"
	ComponentCapabilityMailboxes		= "mailboxes"
	ComponentCapabilityContinuations	= "continuations"
	ComponentCapabilityNativeModules	= "native_modules"
	ComponentCapabilityAVMActors		= "avm_actors"
	ComponentCapabilityWASMAdapter		= "optional_wasm_adapter"
	ComponentCapabilityStoreV2		= "store_v2"
	ComponentCapabilityZoneRoots		= "zone_roots"

	MaxComponentIDLength		= 96
	MaxComponentCapabilityLength	= 96
)

type ComponentLayer string
type ComponentID string

type AVMComponent struct {
	ID		ComponentID
	Layer		ComponentLayer
	Capabilities	[]string
}

type AVMComponentEdge struct {
	From	ComponentID
	To	ComponentID
}

type AVMComponentMap struct {
	Components	[]AVMComponent
	Edges		[]AVMComponentEdge
	Root		string
}

func DefaultAVMComponentMap() AVMComponentMap {
	components := []AVMComponent{
		{
			ID:	ComponentAetraCore,
			Layer:	ComponentLayerCore,
			Capabilities: []string{
				ComponentCapabilityBlockLifecycle,
				ComponentCapabilityConsensus,
				ComponentCapabilityGlobalRoots,
				ComponentCapabilityZoneRegistry,
			},
		},
		{
			ID:	ComponentAVMRouter,
			Layer:	ComponentLayerRouter,
			Capabilities: []string{
				ComponentCapabilityClassifyZone,
				ComponentCapabilityDispatch,
				ComponentCapabilityRouteMessages,
				ComponentCapabilityValidateBudget,
			},
		},
		{
			ID:	ComponentSyncEngine,
			Layer:	ComponentLayerEngine,
			Capabilities: []string{
				ComponentCapabilityKeeperCalls,
				ComponentCapabilityMsgServer,
			},
		},
		{
			ID:	ComponentAsyncEngine,
			Layer:	ComponentLayerEngine,
			Capabilities: []string{
				ComponentCapabilityQueues,
				ComponentCapabilityScheduling,
			},
		},
		{
			ID:	ComponentActorRuntime,
			Layer:	ComponentLayerEngine,
			Capabilities: []string{
				ComponentCapabilityContinuations,
				ComponentCapabilityMailboxes,
			},
		},
		{
			ID:	ComponentNativeModules,
			Layer:	ComponentLayerBackend,
			Capabilities: []string{
				ComponentCapabilityNativeModules,
			},
		},
		{
			ID:	ComponentAVMActors,
			Layer:	ComponentLayerBackend,
			Capabilities: []string{
				ComponentCapabilityAVMActors,
			},
		},
		{
			ID:	ComponentWASMAdapter,
			Layer:	ComponentLayerBackend,
			Capabilities: []string{
				ComponentCapabilityWASMAdapter,
			},
		},
		{
			ID:	ComponentZoneStateRoots,
			Layer:	ComponentLayerZoneState,
			Capabilities: []string{
				ComponentCapabilityStoreV2,
				ComponentCapabilityZoneRoots,
			},
		},
	}
	edges := []AVMComponentEdge{
		{From: ComponentAetraCore, To: ComponentAVMRouter},
		{From: ComponentAVMRouter, To: ComponentSyncEngine},
		{From: ComponentAVMRouter, To: ComponentAsyncEngine},
		{From: ComponentAVMRouter, To: ComponentActorRuntime},
		{From: ComponentSyncEngine, To: ComponentNativeModules},
		{From: ComponentSyncEngine, To: ComponentAVMActors},
		{From: ComponentAsyncEngine, To: ComponentAVMActors},
		{From: ComponentActorRuntime, To: ComponentAVMActors},
		{From: ComponentActorRuntime, To: ComponentWASMAdapter},
		{From: ComponentNativeModules, To: ComponentZoneStateRoots},
		{From: ComponentAVMActors, To: ComponentZoneStateRoots},
		{From: ComponentWASMAdapter, To: ComponentZoneStateRoots},
	}
	componentMap := AVMComponentMap{Components: components, Edges: edges}
	componentMap = CanonicalAVMComponentMap(componentMap)
	componentMap.Root = ComputeAVMComponentMapRoot(componentMap)
	return componentMap
}

func CanonicalAVMComponentMap(componentMap AVMComponentMap) AVMComponentMap {
	out := AVMComponentMap{
		Components:	cloneAVMComponents(componentMap.Components),
		Edges:		append([]AVMComponentEdge(nil), componentMap.Edges...),
		Root:		strings.TrimSpace(componentMap.Root),
	}
	sort.SliceStable(out.Components, func(i, j int) bool {
		return out.Components[i].ID < out.Components[j].ID
	})
	sort.SliceStable(out.Edges, func(i, j int) bool {
		if out.Edges[i].From != out.Edges[j].From {
			return out.Edges[i].From < out.Edges[j].From
		}
		return out.Edges[i].To < out.Edges[j].To
	})
	return out
}

func (m AVMComponentMap) Validate() error {
	m = CanonicalAVMComponentMap(m)
	if len(m.Components) == 0 {
		return errors.New("AVM component map must declare components")
	}
	if len(m.Edges) == 0 {
		return errors.New("AVM component map must declare edges")
	}
	components := make(map[ComponentID]AVMComponent, len(m.Components))
	for i, component := range m.Components {
		if err := component.Validate(); err != nil {
			return err
		}
		if _, found := components[component.ID]; found {
			return fmt.Errorf("duplicate AVM component %q", component.ID)
		}
		components[component.ID] = component
		if i > 0 && m.Components[i-1].ID >= component.ID {
			return errors.New("AVM component map components must be sorted canonically")
		}
	}
	if err := validateRequiredComponents(components); err != nil {
		return err
	}
	seenEdges := make(map[string]struct{}, len(m.Edges))
	for i, edge := range m.Edges {
		if err := edge.Validate(components); err != nil {
			return err
		}
		key := string(edge.From) + "->" + string(edge.To)
		if _, found := seenEdges[key]; found {
			return fmt.Errorf("duplicate AVM component edge %s", key)
		}
		seenEdges[key] = struct{}{}
		if i > 0 && compareComponentEdges(m.Edges[i-1], edge) >= 0 {
			return errors.New("AVM component map edges must be sorted canonically")
		}
	}
	if err := validateRequiredEdges(seenEdges); err != nil {
		return err
	}
	if err := validateAcyclicComponentMap(m.Edges); err != nil {
		return err
	}
	if m.Root == "" {
		return errors.New("AVM component map root is required")
	}
	if m.Root != ComputeAVMComponentMapRoot(m) {
		return errors.New("AVM component map root mismatch")
	}
	return nil
}

func (c AVMComponent) Validate() error {
	if err := validateComponentToken("AVM component id", string(c.ID), MaxComponentIDLength); err != nil {
		return err
	}
	if !IsAVMComponentLayer(c.Layer) {
		return fmt.Errorf("unknown AVM component layer %q", c.Layer)
	}
	if len(c.Capabilities) == 0 {
		return fmt.Errorf("AVM component %q must declare capabilities", c.ID)
	}
	var previous string
	seen := make(map[string]struct{}, len(c.Capabilities))
	for i, capability := range c.Capabilities {
		if err := validateComponentToken("AVM component capability", capability, MaxComponentCapabilityLength); err != nil {
			return err
		}
		if _, found := seen[capability]; found {
			return fmt.Errorf("duplicate AVM component capability %q", capability)
		}
		seen[capability] = struct{}{}
		if i > 0 && previous >= capability {
			return errors.New("AVM component capabilities must be sorted canonically")
		}
		previous = capability
	}
	return nil
}

func (e AVMComponentEdge) Validate(components map[ComponentID]AVMComponent) error {
	if _, found := components[e.From]; !found {
		return fmt.Errorf("AVM component edge source %q is not declared", e.From)
	}
	if _, found := components[e.To]; !found {
		return fmt.Errorf("AVM component edge target %q is not declared", e.To)
	}
	if e.From == e.To {
		return fmt.Errorf("AVM component edge %q must not be self-referential", e.From)
	}
	fromLayer := components[e.From].Layer
	toLayer := components[e.To].Layer
	if componentLayerOrder(fromLayer) >= componentLayerOrder(toLayer) {
		return fmt.Errorf("AVM component edge %q -> %q must flow toward lower runtime layers", e.From, e.To)
	}
	return nil
}

func IsAVMComponentLayer(layer ComponentLayer) bool {
	switch layer {
	case ComponentLayerCore,
		ComponentLayerRouter,
		ComponentLayerEngine,
		ComponentLayerBackend,
		ComponentLayerZoneState:
		return true
	default:
		return false
	}
}

func ComputeAVMComponentMapRoot(componentMap AVMComponentMap) string {
	componentMap = CanonicalAVMComponentMap(componentMap)
	h := sha256.New()
	writeComponentPart(h, "aetra-avm-component-map-v1")
	writeComponentUint64(h, uint64(len(componentMap.Components)))
	for _, component := range componentMap.Components {
		writeComponentPart(h, string(component.ID))
		writeComponentPart(h, string(component.Layer))
		writeComponentUint64(h, uint64(len(component.Capabilities)))
		for _, capability := range component.Capabilities {
			writeComponentPart(h, capability)
		}
	}
	writeComponentUint64(h, uint64(len(componentMap.Edges)))
	for _, edge := range componentMap.Edges {
		writeComponentPart(h, string(edge.From))
		writeComponentPart(h, string(edge.To))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func validateRequiredComponents(components map[ComponentID]AVMComponent) error {
	required := map[ComponentID][]string{
		ComponentAetraCore:		{ComponentCapabilityConsensus, ComponentCapabilityBlockLifecycle, ComponentCapabilityGlobalRoots, ComponentCapabilityZoneRegistry},
		ComponentAVMRouter:		{ComponentCapabilityRouteMessages, ComponentCapabilityClassifyZone, ComponentCapabilityValidateBudget, ComponentCapabilityDispatch},
		ComponentSyncEngine:		{ComponentCapabilityMsgServer, ComponentCapabilityKeeperCalls},
		ComponentAsyncEngine:		{ComponentCapabilityQueues, ComponentCapabilityScheduling},
		ComponentActorRuntime:		{ComponentCapabilityMailboxes, ComponentCapabilityContinuations},
		ComponentNativeModules:		{ComponentCapabilityNativeModules},
		ComponentAVMActors:		{ComponentCapabilityAVMActors},
		ComponentWASMAdapter:		{ComponentCapabilityWASMAdapter},
		ComponentZoneStateRoots:	{ComponentCapabilityStoreV2, ComponentCapabilityZoneRoots},
	}
	for id, capabilities := range required {
		component, found := components[id]
		if !found {
			return fmt.Errorf("AVM component map missing required component %q", id)
		}
		for _, capability := range capabilities {
			if !componentHasCapability(component, capability) {
				return fmt.Errorf("AVM component %q missing capability %q", id, capability)
			}
		}
	}
	return nil
}

func validateRequiredEdges(edges map[string]struct{}) error {
	required := []AVMComponentEdge{
		{From: ComponentAetraCore, To: ComponentAVMRouter},
		{From: ComponentAVMRouter, To: ComponentSyncEngine},
		{From: ComponentAVMRouter, To: ComponentAsyncEngine},
		{From: ComponentAVMRouter, To: ComponentActorRuntime},
		{From: ComponentSyncEngine, To: ComponentNativeModules},
		{From: ComponentSyncEngine, To: ComponentAVMActors},
		{From: ComponentAsyncEngine, To: ComponentAVMActors},
		{From: ComponentActorRuntime, To: ComponentAVMActors},
		{From: ComponentNativeModules, To: ComponentZoneStateRoots},
		{From: ComponentAVMActors, To: ComponentZoneStateRoots},
		{From: ComponentWASMAdapter, To: ComponentZoneStateRoots},
	}
	for _, edge := range required {
		key := string(edge.From) + "->" + string(edge.To)
		if _, found := edges[key]; !found {
			return fmt.Errorf("AVM component map missing edge %s", key)
		}
	}
	return nil
}

func validateAcyclicComponentMap(edges []AVMComponentEdge) error {
	graph := make(map[ComponentID][]ComponentID, len(edges))
	for _, edge := range edges {
		graph[edge.From] = append(graph[edge.From], edge.To)
	}
	visiting := make(map[ComponentID]bool)
	visited := make(map[ComponentID]bool)
	var visit func(ComponentID) error
	visit = func(id ComponentID) error {
		if visiting[id] {
			return fmt.Errorf("AVM component map contains cycle at %q", id)
		}
		if visited[id] {
			return nil
		}
		visiting[id] = true
		for _, next := range graph[id] {
			if err := visit(next); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for from := range graph {
		if err := visit(from); err != nil {
			return err
		}
	}
	return nil
}

func cloneAVMComponents(components []AVMComponent) []AVMComponent {
	out := make([]AVMComponent, len(components))
	for i, component := range components {
		out[i] = component
		out[i].ID = ComponentID(strings.TrimSpace(string(component.ID)))
		out[i].Layer = ComponentLayer(strings.TrimSpace(string(component.Layer)))
		out[i].Capabilities = append([]string(nil), component.Capabilities...)
		for j, capability := range out[i].Capabilities {
			out[i].Capabilities[j] = strings.TrimSpace(capability)
		}
		sort.Strings(out[i].Capabilities)
	}
	return out
}

func compareComponentEdges(left, right AVMComponentEdge) int {
	if left.From < right.From {
		return -1
	}
	if left.From > right.From {
		return 1
	}
	if left.To < right.To {
		return -1
	}
	if left.To > right.To {
		return 1
	}
	return 0
}

func componentLayerOrder(layer ComponentLayer) int {
	switch layer {
	case ComponentLayerCore:
		return 0
	case ComponentLayerRouter:
		return 1
	case ComponentLayerEngine:
		return 2
	case ComponentLayerBackend:
		return 3
	case ComponentLayerZoneState:
		return 4
	default:
		return -1
	}
}

func componentHasCapability(component AVMComponent, capability string) bool {
	for _, item := range component.Capabilities {
		if item == capability {
			return true
		}
	}
	return false
}

func validateComponentToken(fieldName, value string, maxLen int) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxLen)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func writeComponentPart(w componentByteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func writeComponentUint64(w componentByteWriter, value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	_, _ = w.Write(out[:])
}

type componentByteWriter interface {
	Write([]byte) (int, error)
}
