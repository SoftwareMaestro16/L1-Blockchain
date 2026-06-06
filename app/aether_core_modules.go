package app

import aethercorewiring "github.com/sovereign-l1/l1/app/wiring/aethercore"

type RoutingExecutionPoint = aethercorewiring.RoutingExecutionPoint

const RoutingExecutionPointAnteAdmissionOnly = aethercorewiring.RoutingExecutionPointAnteAdmissionOnly

func AetherCoreRoutingExecutionPoint() RoutingExecutionPoint {
	return aethercorewiring.RoutingExecution()
}

func AetherCorePrototypeModuleNames() []string {
	return aethercorewiring.PrototypeModuleNames()
}

func AetherCorePrototypeStoreKeys() []string {
	return aethercorewiring.PrototypeStoreKeys()
}

func AetherCoreSystemModuleNames() []string {
	return aethercorewiring.SystemModuleNames()
}

func AetherCoreSystemStoreKeys() []string {
	return aethercorewiring.SystemStoreKeys()
}
