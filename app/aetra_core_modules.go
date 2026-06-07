package app

import aetracorewiring "github.com/sovereign-l1/l1/app/wiring/aetracore"

type RoutingExecutionPoint = aetracorewiring.RoutingExecutionPoint

const RoutingExecutionPointAnteAdmissionOnly = aetracorewiring.RoutingExecutionPointAnteAdmissionOnly

func AetraCoreRoutingExecutionPoint() RoutingExecutionPoint {
	return aetracorewiring.RoutingExecution()
}

func AetraCorePrototypeModuleNames() []string {
	return aetracorewiring.PrototypeModuleNames()
}

func AetraCorePrototypeStoreKeys() []string {
	return aetracorewiring.PrototypeStoreKeys()
}

func AetraCoreSystemModuleNames() []string {
	return aetracorewiring.SystemModuleNames()
}

func AetraCoreSystemStoreKeys() []string {
	return aetracorewiring.SystemStoreKeys()
}
