package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAVMComponentMapMatchesCoreComponentDiagram(t *testing.T) {
	componentMap := DefaultAVMComponentMap()
	require.NoError(t, componentMap.Validate())
	require.Len(t, componentMap.Components, 9)
	require.Len(t, componentMap.Edges, 12)
	require.Equal(t, ComputeAVMComponentMapRoot(componentMap), componentMap.Root)

	components := componentMapByID(componentMap)
	requireComponent(t, components, ComponentAetraCore, ComponentLayerCore,
		ComponentCapabilityConsensus,
		ComponentCapabilityBlockLifecycle,
		ComponentCapabilityGlobalRoots,
		ComponentCapabilityZoneRegistry,
	)
	requireComponent(t, components, ComponentAVMRouter, ComponentLayerRouter,
		ComponentCapabilityRouteMessages,
		ComponentCapabilityClassifyZone,
		ComponentCapabilityValidateBudget,
		ComponentCapabilityDispatch,
	)
	requireComponent(t, components, ComponentSyncEngine, ComponentLayerEngine,
		ComponentCapabilityMsgServer,
		ComponentCapabilityKeeperCalls,
	)
	requireComponent(t, components, ComponentAsyncEngine, ComponentLayerEngine,
		ComponentCapabilityQueues,
		ComponentCapabilityScheduling,
	)
	requireComponent(t, components, ComponentActorRuntime, ComponentLayerEngine,
		ComponentCapabilityMailboxes,
		ComponentCapabilityContinuations,
	)
	requireComponent(t, components, ComponentNativeModules, ComponentLayerBackend, ComponentCapabilityNativeModules)
	requireComponent(t, components, ComponentAVMActors, ComponentLayerBackend, ComponentCapabilityAVMActors)
	requireComponent(t, components, ComponentWASMAdapter, ComponentLayerBackend, ComponentCapabilityWASMAdapter)
	requireComponent(t, components, ComponentZoneStateRoots, ComponentLayerZoneState,
		ComponentCapabilityStoreV2,
		ComponentCapabilityZoneRoots,
	)

	requireComponentEdge(t, componentMap, ComponentAetraCore, ComponentAVMRouter)
	requireComponentEdge(t, componentMap, ComponentAVMRouter, ComponentSyncEngine)
	requireComponentEdge(t, componentMap, ComponentAVMRouter, ComponentAsyncEngine)
	requireComponentEdge(t, componentMap, ComponentAVMRouter, ComponentActorRuntime)
	requireComponentEdge(t, componentMap, ComponentSyncEngine, ComponentNativeModules)
	requireComponentEdge(t, componentMap, ComponentSyncEngine, ComponentAVMActors)
	requireComponentEdge(t, componentMap, ComponentAsyncEngine, ComponentAVMActors)
	requireComponentEdge(t, componentMap, ComponentActorRuntime, ComponentAVMActors)
	requireComponentEdge(t, componentMap, ComponentNativeModules, ComponentZoneStateRoots)
	requireComponentEdge(t, componentMap, ComponentAVMActors, ComponentZoneStateRoots)
	requireComponentEdge(t, componentMap, ComponentWASMAdapter, ComponentZoneStateRoots)
}

func TestAVMComponentMapCanonicalRootIsDeterministic(t *testing.T) {
	componentMap := DefaultAVMComponentMap()
	reordered := AVMComponentMap{
		Components:	append([]AVMComponent(nil), componentMap.Components...),
		Edges:		append([]AVMComponentEdge(nil), componentMap.Edges...),
	}
	reordered.Components[0], reordered.Components[len(reordered.Components)-1] = reordered.Components[len(reordered.Components)-1], reordered.Components[0]
	reordered.Edges[0], reordered.Edges[len(reordered.Edges)-1] = reordered.Edges[len(reordered.Edges)-1], reordered.Edges[0]
	reordered = CanonicalAVMComponentMap(reordered)
	reordered.Root = ComputeAVMComponentMapRoot(reordered)
	require.Equal(t, componentMap.Root, reordered.Root)
	require.NoError(t, reordered.Validate())

	mutated := componentMap
	mutated.Components = append([]AVMComponent(nil), componentMap.Components...)
	for i := range mutated.Components {
		if mutated.Components[i].ID == ComponentAVMRouter {
			mutated.Components[i].Capabilities = append([]string(nil), mutated.Components[i].Capabilities...)
			mutated.Components[i].Capabilities = append(mutated.Components[i].Capabilities, "observability")
			break
		}
	}
	mutated = CanonicalAVMComponentMap(mutated)
	mutated.Root = ComputeAVMComponentMapRoot(mutated)
	require.NotEqual(t, componentMap.Root, mutated.Root)
	require.NoError(t, mutated.Validate())
}

func TestAVMComponentMapRejectsMissingRequiredCapabilitiesEdgesAndInvalidFlow(t *testing.T) {
	missingCapability := DefaultAVMComponentMap()
	for i := range missingCapability.Components {
		if missingCapability.Components[i].ID == ComponentAVMRouter {
			missingCapability.Components[i].Capabilities = []string{
				ComponentCapabilityClassifyZone,
				ComponentCapabilityDispatch,
				ComponentCapabilityRouteMessages,
			}
			break
		}
	}
	missingCapability = CanonicalAVMComponentMap(missingCapability)
	missingCapability.Root = ComputeAVMComponentMapRoot(missingCapability)
	require.ErrorContains(t, missingCapability.Validate(), "validate_budget")

	missingEdge := DefaultAVMComponentMap()
	filtered := missingEdge.Edges[:0]
	for _, edge := range missingEdge.Edges {
		if edge.From == ComponentAVMRouter && edge.To == ComponentActorRuntime {
			continue
		}
		filtered = append(filtered, edge)
	}
	missingEdge.Edges = filtered
	missingEdge = CanonicalAVMComponentMap(missingEdge)
	missingEdge.Root = ComputeAVMComponentMapRoot(missingEdge)
	require.ErrorContains(t, missingEdge.Validate(), "missing edge")

	invalidFlow := DefaultAVMComponentMap()
	invalidFlow.Edges = append(invalidFlow.Edges, AVMComponentEdge{From: ComponentZoneStateRoots, To: ComponentAetraCore})
	invalidFlow = CanonicalAVMComponentMap(invalidFlow)
	invalidFlow.Root = ComputeAVMComponentMapRoot(invalidFlow)
	require.ErrorContains(t, invalidFlow.Validate(), "lower runtime layers")

	rootMismatch := DefaultAVMComponentMap()
	rootMismatch.Root = rootMismatch.Root[:len(rootMismatch.Root)-1] + "0"
	require.ErrorContains(t, rootMismatch.Validate(), "root mismatch")
}

func componentMapByID(componentMap AVMComponentMap) map[ComponentID]AVMComponent {
	out := make(map[ComponentID]AVMComponent, len(componentMap.Components))
	for _, component := range componentMap.Components {
		out[component.ID] = component
	}
	return out
}

func requireComponent(t *testing.T, components map[ComponentID]AVMComponent, id ComponentID, layer ComponentLayer, capabilities ...string) {
	t.Helper()
	component, found := components[id]
	require.True(t, found)
	require.Equal(t, layer, component.Layer)
	for _, capability := range capabilities {
		require.True(t, componentHasCapability(component, capability), capability)
	}
}

func requireComponentEdge(t *testing.T, componentMap AVMComponentMap, from ComponentID, to ComponentID) {
	t.Helper()
	for _, edge := range componentMap.Edges {
		if edge.From == from && edge.To == to {
			return
		}
	}
	t.Fatalf("missing edge %s -> %s", from, to)
}
