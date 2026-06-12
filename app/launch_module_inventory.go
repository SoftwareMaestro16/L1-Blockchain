package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed launch_module_inventory.json
var launchModuleInventoryJSON []byte

type LaunchModuleClassification string

const (
	LaunchModuleLaunchCore		LaunchModuleClassification	= "launch_core"
	LaunchModuleLaunchSupport	LaunchModuleClassification	= "launch_support"
	LaunchModuleFutureAVMStandard	LaunchModuleClassification	= "future_avm_standard"
	LaunchModulePrototypeOnly	LaunchModuleClassification	= "prototype_only"
	LaunchModuleDisabled		LaunchModuleClassification	= "disabled"
)

type LaunchModuleInventoryEntry struct {
	XDir				string				`json:"x_dir"`
	ModuleName			string				`json:"module_name"`
	Classification			LaunchModuleClassification	`json:"classification"`
	AppWired			bool				`json:"app_wired"`
	PublicTestnetReason		string				`json:"public_testnet_reason"`
	OwnsConsensusState		bool				`json:"owns_consensus_state"`
	KVBackedRuntimeMutations	bool				`json:"kv_backed_runtime_mutations"`
	ExportImportStatus		string				`json:"export_import_status"`
	InvariantStatus			string				`json:"invariant_status"`
	BlockLifecycleScanningRisk	string				`json:"block_lifecycle_scanning_risk"`
}

func DefaultLaunchModuleInventory() []LaunchModuleInventoryEntry {
	var entries []LaunchModuleInventoryEntry
	if err := json.Unmarshal(launchModuleInventoryJSON, &entries); err != nil {
		panic(fmt.Errorf("invalid launch module inventory: %w", err))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].XDir < entries[j].XDir
	})
	return entries
}

func ValidateLaunchModuleInventory(entries []LaunchModuleInventoryEntry, xDirs []string) error {
	if len(entries) == 0 {
		return fmt.Errorf("launch module inventory is required")
	}
	classes := map[LaunchModuleClassification]struct{}{
		LaunchModuleLaunchCore:		{},
		LaunchModuleLaunchSupport:	{},
		LaunchModuleFutureAVMStandard:	{},
		LaunchModulePrototypeOnly:	{},
		LaunchModuleDisabled:		{},
	}
	xDirSet := make(map[string]struct{}, len(xDirs))
	for _, dir := range xDirs {
		dir = normalizeXDir(dir)
		if dir != "" {
			xDirSet[dir] = struct{}{}
		}
	}
	seenDirs := make(map[string]struct{}, len(entries))
	seenModules := make(map[string]string, len(entries))
	for _, entry := range entries {
		if _, found := classes[entry.Classification]; !found {
			return fmt.Errorf("%s has invalid classification %q", entry.XDir, entry.Classification)
		}
		if normalizeXDir(entry.XDir) != entry.XDir || !strings.HasPrefix(entry.XDir, "x/") {
			return fmt.Errorf("inventory x_dir must be normalized x/<name>, got %q", entry.XDir)
		}
		if _, duplicate := seenDirs[entry.XDir]; duplicate {
			return fmt.Errorf("duplicate launch inventory x_dir %s", entry.XDir)
		}
		seenDirs[entry.XDir] = struct{}{}
		if len(xDirSet) > 0 {
			if _, found := xDirSet[entry.XDir]; !found {
				return fmt.Errorf("inventory references missing directory %s", entry.XDir)
			}
		}
		if strings.TrimSpace(entry.PublicTestnetReason) == "" {
			return fmt.Errorf("%s missing public testnet reason", entry.XDir)
		}
		if strings.TrimSpace(entry.ExportImportStatus) == "" {
			return fmt.Errorf("%s missing export/import status", entry.XDir)
		}
		if strings.TrimSpace(entry.InvariantStatus) == "" {
			return fmt.Errorf("%s missing invariant status", entry.XDir)
		}
		if strings.TrimSpace(entry.BlockLifecycleScanningRisk) == "" {
			return fmt.Errorf("%s missing block lifecycle scanning risk", entry.XDir)
		}
		if entry.AppWired && strings.TrimSpace(entry.ModuleName) == "" {
			return fmt.Errorf("%s is app_wired but has no module_name", entry.XDir)
		}
		if entry.ModuleName != "" {
			moduleName := strings.TrimSpace(entry.ModuleName)
			if moduleName != entry.ModuleName {
				return fmt.Errorf("%s module_name must be trimmed", entry.XDir)
			}
			if previous, duplicate := seenModules[moduleName]; duplicate {
				return fmt.Errorf("duplicate module_name %s in %s and %s", moduleName, previous, entry.XDir)
			}
			seenModules[moduleName] = entry.XDir
		}
	}
	for dir := range xDirSet {
		if _, found := seenDirs[dir]; !found {
			return fmt.Errorf("launch module inventory missing %s", dir)
		}
	}
	return nil
}

func ValidatePublicTestnetLaunchProfile(entries []LaunchModuleInventoryEntry, appModules []string) error {
	byModule := make(map[string]LaunchModuleInventoryEntry, len(entries))
	for _, entry := range entries {
		if entry.ModuleName != "" {
			byModule[entry.ModuleName] = entry
		}
	}
	for _, moduleName := range appModules {
		entry, found := byModule[moduleName]
		if !found {
			continue
		}
		if !entry.AppWired {
			return fmt.Errorf("module %s is wired but inventory marks %s as not app_wired", moduleName, entry.XDir)
		}
		switch entry.Classification {
		case LaunchModulePrototypeOnly, LaunchModuleDisabled:
			return fmt.Errorf("public testnet profile must not wire %s module %s", entry.Classification, moduleName)
		case LaunchModuleFutureAVMStandard:
			return fmt.Errorf("public testnet profile must not wire future AVM standard %s as a native module", moduleName)
		}
		if entry.OwnsConsensusState && !entry.KVBackedRuntimeMutations {
			return fmt.Errorf("module %s owns consensus state but is not KV-backed", moduleName)
		}
		if isNativeApplicationAssetModuleName(moduleName) {
			return fmt.Errorf("native application-asset module %s is not allowed in public testnet profile", moduleName)
		}
	}
	return nil
}

func RenderLaunchModuleInventoryMarkdown(entries []LaunchModuleInventoryEntry) string {
	entries = append([]LaunchModuleInventoryEntry(nil), entries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].XDir < entries[j].XDir
	})
	var b strings.Builder
	b.WriteString("# Launch Module Inventory\n\n")
	b.WriteString("Generated from `app/launch_module_inventory.json`. Keep `docs/module-boundaries.md` linked to this inventory.\n\n")
	b.WriteString("| x directory | module | classification | wired | consensus state | KV runtime | export/import | invariants | block scan risk |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, entry := range entries {
		moduleName := entry.ModuleName
		if moduleName == "" {
			moduleName = "-"
		}
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %t | %t | %t | %s | %s | %s |\n",
			entry.XDir,
			moduleName,
			entry.Classification,
			entry.AppWired,
			entry.OwnsConsensusState,
			entry.KVBackedRuntimeMutations,
			entry.ExportImportStatus,
			entry.InvariantStatus,
			entry.BlockLifecycleScanningRisk,
		))
	}
	return b.String()
}

func RenderLaunchModuleInventoryBoundarySummary(entries []LaunchModuleInventoryEntry) string {
	counts := map[LaunchModuleClassification]int{}
	for _, entry := range entries {
		counts[entry.Classification]++
	}
	classes := []LaunchModuleClassification{
		LaunchModuleDisabled,
		LaunchModuleFutureAVMStandard,
		LaunchModuleLaunchCore,
		LaunchModuleLaunchSupport,
		LaunchModulePrototypeOnly,
	}
	var b strings.Builder
	b.WriteString("## Launch Module Inventory\n\n")
	b.WriteString("Machine-readable launch inventory: `app/launch_module_inventory.json`.\n")
	b.WriteString("Generated launch scope doc: `docs/TESTNET.md`.\n\n")
	b.WriteString("Classification counts:\n")
	for _, class := range classes {
		b.WriteString(fmt.Sprintf("- `%s`: %d\n", class, counts[class]))
	}
	b.WriteString("\nPublic testnet profile rejects `prototype_only` and `disabled` modules in app wiring, rejects memory-only consensus keepers, and rejects native application-asset modules.\n")
	return b.String()
}

func normalizeXDir(dir string) string {
	dir = strings.TrimSpace(strings.ReplaceAll(dir, "\\", "/"))
	dir = strings.TrimPrefix(dir, "./")
	if dir == "" {
		return ""
	}
	if strings.HasPrefix(dir, "x/") {
		return dir
	}
	return "x/" + dir
}

func isNativeApplicationAssetModuleName(moduleName string) bool {
	switch moduleName {
	case "asset", "assetfactory":
		return true
	default:
		return false
	}
}
