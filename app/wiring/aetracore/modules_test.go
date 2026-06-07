package aetracore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAetraCoreCatalogsAreAlignedAndDeterministic(t *testing.T) {
	require.Equal(t, RoutingExecutionPointAnteAdmissionOnly, RoutingExecution())

	prototypeNames := PrototypeModuleNames()
	prototypeKeys := PrototypeStoreKeys()
	systemNames := SystemModuleNames()
	systemKeys := SystemStoreKeys()

	require.Len(t, prototypeKeys, len(prototypeNames))
	require.Len(t, systemKeys, len(systemNames))
	requireNoDuplicates(t, prototypeNames)
	requireNoDuplicates(t, prototypeKeys)
	requireNoDuplicates(t, systemNames)
	requireNoDuplicates(t, systemKeys)
}

func TestAetraCoreLifecycleOrdersIncludeRegisteredCatalogModules(t *testing.T) {
	initOrder := InitGenesisOrder()
	exportOrder := ExportGenesisOrder()

	for _, moduleName := range append(PrototypeModuleNames(), SystemModuleNames()...) {
		require.Contains(t, initOrder, moduleName)
		require.Contains(t, exportOrder, moduleName)
	}
}

func requireNoDuplicates(t *testing.T, values []string) {
	t.Helper()
	seen := map[string]struct{}{}
	for _, value := range values {
		require.NotEmpty(t, value)
		require.NotContains(t, seen, value)
		seen[value] = struct{}{}
	}
}
