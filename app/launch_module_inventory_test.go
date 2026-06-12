package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLaunchModuleInventoryCoversEveryXDirectory(t *testing.T) {
	entries := DefaultLaunchModuleInventory()

	require.NoError(t, ValidateLaunchModuleInventory(entries, readXModuleDirs(t)))
	require.Contains(t, launchInventoryEntryByXDir(entries), "x/market")
	require.Contains(t, launchInventoryEntryByXDir(entries), "x/aetravm")
	require.Contains(t, launchInventoryEntryByXDir(entries), "x/contracts")
}

func TestLaunchModuleInventoryCoversWiredAetraModules(t *testing.T) {
	app, _ := setup(true, 5)
	entries := DefaultLaunchModuleInventory()
	byModule := launchInventoryEntryByModule(entries)

	for moduleName := range app.ModuleManager.Modules {
		if _, found := byModule[moduleName]; found {
			continue
		}
		require.Contains(t, cosmosSDKLaunchModuleNames(), moduleName, "wired module %s must be listed in launch inventory or SDK allowlist", moduleName)
	}
	for _, entry := range entries {
		if entry.AppWired {
			require.Contains(t, app.ModuleManager.Modules, entry.ModuleName, "inventory marks %s as wired", entry.XDir)
		}
	}
	require.NoError(t, ValidatePublicTestnetLaunchProfile(entries, moduleManagerNames(app)))
}

func TestPublicTestnetLaunchProfileRejectsForbiddenInventory(t *testing.T) {
	require.ErrorContains(t,
		ValidatePublicTestnetLaunchProfile([]LaunchModuleInventoryEntry{testInventoryEntry("x/proto", "proto", LaunchModulePrototypeOnly, true, false, false)}, []string{"proto"}),
		"prototype_only",
	)
	require.ErrorContains(t,
		ValidatePublicTestnetLaunchProfile([]LaunchModuleInventoryEntry{testInventoryEntry("x/off", "off", LaunchModuleDisabled, true, false, false)}, []string{"off"}),
		"disabled",
	)
	require.ErrorContains(t,
		ValidatePublicTestnetLaunchProfile([]LaunchModuleInventoryEntry{testInventoryEntry("x/futureavm", "futureavm", LaunchModuleFutureAVMStandard, true, false, false)}, []string{"futureavm"}),
		"future AVM standard",
	)
	require.ErrorContains(t,
		ValidatePublicTestnetLaunchProfile([]LaunchModuleInventoryEntry{testInventoryEntry("x/memory", "memory", LaunchModuleLaunchCore, true, true, false)}, []string{"memory"}),
		"owns consensus state but is not KV-backed",
	)
}

func TestLaunchModuleInventoryDocsMatchModuleBoundarySummary(t *testing.T) {
	entries := DefaultLaunchModuleInventory()
	expected := RenderLaunchModuleInventoryBoundarySummary(entries)
	boundaries, err := os.ReadFile(filepath.Join("..", "docs", "module-boundaries.md"))
	require.NoError(t, err)
	require.Contains(t, strings.ReplaceAll(string(boundaries), "\r\n", "\n"), expected)
}

func readXModuleDirs(t *testing.T) []string {
	t.Helper()
	items, err := os.ReadDir(filepath.Join("..", "x"))
	require.NoError(t, err)
	dirs := make([]string, 0, len(items))
	for _, item := range items {
		if item.IsDir() {
			dirs = append(dirs, "x/"+item.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}

func launchInventoryEntryByXDir(entries []LaunchModuleInventoryEntry) map[string]LaunchModuleInventoryEntry {
	out := make(map[string]LaunchModuleInventoryEntry, len(entries))
	for _, entry := range entries {
		out[entry.XDir] = entry
	}
	return out
}

func launchInventoryEntryByModule(entries []LaunchModuleInventoryEntry) map[string]LaunchModuleInventoryEntry {
	out := make(map[string]LaunchModuleInventoryEntry, len(entries))
	for _, entry := range entries {
		if entry.ModuleName != "" {
			out[entry.ModuleName] = entry
		}
	}
	return out
}

func moduleManagerNames(app *L1App) []string {
	names := make([]string, 0, len(app.ModuleManager.Modules))
	for moduleName := range app.ModuleManager.Modules {
		names = append(names, moduleName)
	}
	sort.Strings(names)
	return names
}

func cosmosSDKLaunchModuleNames() map[string]struct{} {
	return map[string]struct{}{
		"auth":		{},
		"authz":	{},
		"bank":		{},
		"consensus":	{},
		"distribution":	{},
		"epochs":	{},
		"evidence":	{},
		"feegrant":	{},
		"genutil":	{},
		"gov":		{},
		"mint":		{},
		"protocolpool":	{},
		"slashing":	{},
		"staking":	{},
		"upgrade":	{},
		"vesting":	{},
	}
}

func testInventoryEntry(xDir, moduleName string, class LaunchModuleClassification, wired, ownsConsensus, kvBacked bool) LaunchModuleInventoryEntry {
	return LaunchModuleInventoryEntry{
		XDir:				xDir,
		ModuleName:			moduleName,
		Classification:			class,
		AppWired:			wired,
		PublicTestnetReason:		"test entry",
		OwnsConsensusState:		ownsConsensus,
		KVBackedRuntimeMutations:	kvBacked,
		ExportImportStatus:		"test",
		InvariantStatus:		"test",
		BlockLifecycleScanningRisk:	"test",
	}
}
