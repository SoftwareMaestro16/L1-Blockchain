package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ImplementationRoadmapPhaseID string

const (
	RoadmapPhaseBaselineAudit   ImplementationRoadmapPhaseID = "phase-0-baseline-audit"
	RoadmapPhaseKernelRootModel ImplementationRoadmapPhaseID = "phase-1-kernel-root-model"
)

type ImplementationRoadmap struct {
	Phases      []ImplementationRoadmapPhase
	RoadmapHash string
}

type ImplementationRoadmapPhase struct {
	PhaseID      ImplementationRoadmapPhaseID
	PhaseNumber  uint32
	Name         string
	Tasks        []RoadmapChecklistItem
	ExitCriteria []RoadmapChecklistItem
	Evidence     RoadmapEvidence
	PhaseHash    string
}

type RoadmapChecklistItem struct {
	ID          string
	Description string
	Complete    bool
}

type RoadmapEvidence struct {
	ModuleInventory                 []RoadmapModuleInventoryEntry
	CrossModuleDirectWritesAudited  bool
	ExportImportTestsAdded          bool
	ModuleInvariantHarnessAdded     bool
	RootContributionInterfaceDesign bool
	CurrentStateReproducible        bool
	ModuleBoundariesDocumented      bool
	MigrationRiskListComplete       bool
	AetherCoreModuleImplemented     bool
	ZonesModuleImplemented          bool
	ZoneRegistryImplemented         bool
	GlobalStateRootImplemented      bool
	BlockCommitmentMetadataQueries  bool
	DefaultZoneRunnable             bool
	DefaultZoneRootIncluded         bool
	ExportImportPreservesRootMeta   bool
}

type RoadmapModuleInventoryEntry struct {
	ModuleName CosmosSDKModuleName
	ModulePath string
	StoreKey   string
	StateKeys  []string
	RootType   RootType
}

func DefaultImplementationRoadmap() (ImplementationRoadmap, error) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	if err != nil {
		return ImplementationRoadmap{}, err
	}
	inventory := BuildRoadmapModuleInventory(manifest)
	phases := []ImplementationRoadmapPhase{
		roadmapPhaseZero(inventory),
		roadmapPhaseOne(inventory),
	}
	return NewImplementationRoadmap(phases)
}

func NewImplementationRoadmap(phases []ImplementationRoadmapPhase) (ImplementationRoadmap, error) {
	roadmap := ImplementationRoadmap{Phases: normalizeRoadmapPhases(phases)}
	if err := roadmap.ValidateFormat(); err != nil {
		return ImplementationRoadmap{}, err
	}
	for i := range roadmap.Phases {
		roadmap.Phases[i].PhaseHash = ComputeRoadmapPhaseHash(roadmap.Phases[i])
	}
	roadmap.RoadmapHash = ComputeImplementationRoadmapHash(roadmap)
	return roadmap, roadmap.Validate()
}

func BuildRoadmapModuleInventory(manifest CosmosModuleRequirementManifest) []RoadmapModuleInventoryEntry {
	manifest.Modules = normalizeCosmosModuleSurfaces(manifest.Modules)
	inventory := make([]RoadmapModuleInventoryEntry, 0, len(manifest.Modules))
	for _, surface := range manifest.Modules {
		storeKey := surface.KeeperIsolation.StoreKey
		inventory = append(inventory, RoadmapModuleInventoryEntry{
			ModuleName: surface.ModuleName,
			ModulePath: surface.ModulePath,
			StoreKey:   storeKey,
			StateKeys:  []string{storeKey + "/params", storeKey + "/genesis", storeKey + "/root"},
			RootType:   surface.RootContribution.RootType,
		})
	}
	return normalizeRoadmapModuleInventory(inventory)
}

func (roadmap ImplementationRoadmap) ValidateFormat() error {
	roadmap.Phases = normalizeRoadmapPhases(roadmap.Phases)
	if len(roadmap.Phases) != 2 {
		return errors.New("aethercore implementation roadmap must include phase 0 and phase 1")
	}
	if roadmap.Phases[0].PhaseID != RoadmapPhaseBaselineAudit || roadmap.Phases[0].PhaseNumber != 0 {
		return errors.New("aethercore implementation roadmap phase 0 baseline audit is required")
	}
	if roadmap.Phases[1].PhaseID != RoadmapPhaseKernelRootModel || roadmap.Phases[1].PhaseNumber != 1 {
		return errors.New("aethercore implementation roadmap phase 1 kernel root model is required")
	}
	for _, phase := range roadmap.Phases {
		if err := phase.ValidateFormat(); err != nil {
			return err
		}
	}
	if roadmap.RoadmapHash != "" {
		return ValidateHash("aethercore implementation roadmap hash", roadmap.RoadmapHash)
	}
	return nil
}

func (roadmap ImplementationRoadmap) Validate() error {
	roadmap.Phases = normalizeRoadmapPhases(roadmap.Phases)
	if err := roadmap.ValidateFormat(); err != nil {
		return err
	}
	for _, phase := range roadmap.Phases {
		if err := phase.Validate(); err != nil {
			return err
		}
	}
	if roadmap.RoadmapHash != ComputeImplementationRoadmapHash(roadmap) {
		return errors.New("aethercore implementation roadmap hash mismatch")
	}
	return nil
}

func (phase ImplementationRoadmapPhase) ValidateFormat() error {
	phase = normalizeRoadmapPhase(phase)
	if phase.PhaseID != RoadmapPhaseBaselineAudit && phase.PhaseID != RoadmapPhaseKernelRootModel {
		return fmt.Errorf("unknown aethercore implementation roadmap phase %q", phase.PhaseID)
	}
	if phase.PhaseID == RoadmapPhaseBaselineAudit && phase.PhaseNumber != 0 {
		return errors.New("aethercore baseline audit phase number must be 0")
	}
	if phase.PhaseID == RoadmapPhaseKernelRootModel && phase.PhaseNumber != 1 {
		return errors.New("aethercore kernel root model phase number must be 1")
	}
	if err := validateRoadmapText("aethercore implementation roadmap phase name", phase.Name); err != nil {
		return err
	}
	if err := validateRoadmapChecklist("aethercore implementation roadmap task", phase.Tasks); err != nil {
		return err
	}
	if err := validateRoadmapChecklist("aethercore implementation roadmap exit criterion", phase.ExitCriteria); err != nil {
		return err
	}
	if err := phase.Evidence.Validate(phase.PhaseID); err != nil {
		return err
	}
	if phase.PhaseHash != "" {
		return ValidateHash("aethercore implementation roadmap phase hash", phase.PhaseHash)
	}
	return nil
}

func (phase ImplementationRoadmapPhase) Validate() error {
	phase = normalizeRoadmapPhase(phase)
	if err := phase.ValidateFormat(); err != nil {
		return err
	}
	if !roadmapChecklistComplete(phase.Tasks) {
		return fmt.Errorf("aethercore implementation roadmap %s has incomplete tasks", phase.PhaseID)
	}
	if !roadmapChecklistComplete(phase.ExitCriteria) {
		return fmt.Errorf("aethercore implementation roadmap %s has incomplete exit criteria", phase.PhaseID)
	}
	if phase.PhaseHash != ComputeRoadmapPhaseHash(phase) {
		return errors.New("aethercore implementation roadmap phase hash mismatch")
	}
	return nil
}

func (e RoadmapEvidence) Validate(phaseID ImplementationRoadmapPhaseID) error {
	e.ModuleInventory = normalizeRoadmapModuleInventory(e.ModuleInventory)
	if len(e.ModuleInventory) == 0 {
		return errors.New("aethercore implementation roadmap module inventory is required")
	}
	if err := validateRoadmapModuleInventory(e.ModuleInventory); err != nil {
		return err
	}
	switch phaseID {
	case RoadmapPhaseBaselineAudit:
		if !e.CrossModuleDirectWritesAudited {
			return errors.New("aethercore phase 0 must audit cross-module direct writes")
		}
		if !e.ExportImportTestsAdded {
			return errors.New("aethercore phase 0 must add export/import tests")
		}
		if !e.ModuleInvariantHarnessAdded {
			return errors.New("aethercore phase 0 must add module invariant test harness")
		}
		if !e.RootContributionInterfaceDesign {
			return errors.New("aethercore phase 0 must design root contribution interface")
		}
		if !e.CurrentStateReproducible {
			return errors.New("aethercore phase 0 exit requires reproducible current state")
		}
		if !e.ModuleBoundariesDocumented {
			return errors.New("aethercore phase 0 exit requires documented module boundaries")
		}
		if !e.MigrationRiskListComplete {
			return errors.New("aethercore phase 0 exit requires complete migration risk list")
		}
	case RoadmapPhaseKernelRootModel:
		if !e.AetherCoreModuleImplemented {
			return errors.New("aethercore phase 1 must implement x/aethercore")
		}
		if !e.ZonesModuleImplemented {
			return errors.New("aethercore phase 1 must implement x/zones")
		}
		if !e.ZoneRegistryImplemented {
			return errors.New("aethercore phase 1 must add zone registry")
		}
		if !e.RootContributionInterfaceDesign {
			return errors.New("aethercore phase 1 must add root contribution interface")
		}
		if !e.GlobalStateRootImplemented {
			return errors.New("aethercore phase 1 must add GlobalStateRoot")
		}
		if !e.BlockCommitmentMetadataQueries {
			return errors.New("aethercore phase 1 must add block commitment metadata queries")
		}
		if !e.DefaultZoneRunnable {
			return errors.New("aethercore phase 1 exit requires runnable default zone")
		}
		if !e.DefaultZoneRootIncluded {
			return errors.New("aethercore phase 1 exit requires global root to include default zone root")
		}
		if !e.ExportImportPreservesRootMeta {
			return errors.New("aethercore phase 1 exit requires export/import root metadata preservation")
		}
	default:
		return fmt.Errorf("unknown aethercore implementation roadmap phase %q", phaseID)
	}
	return nil
}

func ComputeRoadmapPhaseHash(phase ImplementationRoadmapPhase) string {
	phase = normalizeRoadmapPhase(phase)
	return hashRoot("aetheris-aek-implementation-roadmap-phase-v1", func(w byteWriter) {
		writePart(w, string(phase.PhaseID))
		writeUint64(w, uint64(phase.PhaseNumber))
		writePart(w, phase.Name)
		writeRoadmapChecklist(w, phase.Tasks)
		writeRoadmapChecklist(w, phase.ExitCriteria)
		writeRoadmapEvidence(w, phase.Evidence)
	})
}

func ComputeImplementationRoadmapHash(roadmap ImplementationRoadmap) string {
	roadmap.Phases = normalizeRoadmapPhases(roadmap.Phases)
	return hashRoot("aetheris-aek-implementation-roadmap-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(roadmap.Phases)))
		for _, phase := range roadmap.Phases {
			writePart(w, phase.PhaseHash)
		}
	})
}

func roadmapPhaseZero(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:     RoadmapPhaseBaselineAudit,
		PhaseNumber: 0,
		Name:        "Baseline Audit",
		Tasks: roadmapChecklist(
			"inventory-current-modules-state-keys",
			"identify-cross-module-direct-writes",
			"add-export-import-tests-current-state",
			"add-module-invariant-test-harness",
			"add-root-contribution-interface-design",
		),
		ExitCriteria: roadmapChecklist(
			"current-aetheris-state-reproducible",
			"current-module-boundaries-documented",
			"migration-risk-list-complete",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:                 inventory,
			CrossModuleDirectWritesAudited:  true,
			ExportImportTestsAdded:          true,
			ModuleInvariantHarnessAdded:     true,
			RootContributionInterfaceDesign: true,
			CurrentStateReproducible:        true,
			ModuleBoundariesDocumented:      true,
			MigrationRiskListComplete:       true,
		},
	}
}

func roadmapPhaseOne(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:     RoadmapPhaseKernelRootModel,
		PhaseNumber: 1,
		Name:        "Kernel and Root Model",
		Tasks: roadmapChecklist(
			"implement-x-aethercore",
			"implement-x-zones",
			"add-zone-registry",
			"add-root-contribution-interface",
			"add-global-state-root",
			"add-block-commitment-metadata-queries",
		),
		ExitCriteria: roadmapChecklist(
			"existing-chain-runs-as-default-zone",
			"global-root-includes-default-zone-root",
			"export-import-preserves-root-metadata",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:                 inventory,
			RootContributionInterfaceDesign: true,
			AetherCoreModuleImplemented:     true,
			ZonesModuleImplemented:          true,
			ZoneRegistryImplemented:         true,
			GlobalStateRootImplemented:      true,
			BlockCommitmentMetadataQueries:  true,
			DefaultZoneRunnable:             true,
			DefaultZoneRootIncluded:         true,
			ExportImportPreservesRootMeta:   true,
		},
	}
}

func roadmapChecklist(ids ...string) []RoadmapChecklistItem {
	out := make([]RoadmapChecklistItem, len(ids))
	for i, id := range ids {
		out[i] = RoadmapChecklistItem{ID: id, Description: strings.ReplaceAll(id, "-", " "), Complete: true}
	}
	return out
}

func normalizeRoadmapPhases(phases []ImplementationRoadmapPhase) []ImplementationRoadmapPhase {
	out := make([]ImplementationRoadmapPhase, len(phases))
	for i, phase := range phases {
		phase = normalizeRoadmapPhase(phase)
		if phase.PhaseHash == "" {
			phase.PhaseHash = ComputeRoadmapPhaseHash(phase)
		}
		out[i] = phase
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].PhaseNumber < out[j].PhaseNumber
	})
	return out
}

func normalizeRoadmapPhase(phase ImplementationRoadmapPhase) ImplementationRoadmapPhase {
	phase.PhaseID = ImplementationRoadmapPhaseID(strings.TrimSpace(string(phase.PhaseID)))
	phase.Name = strings.TrimSpace(phase.Name)
	phase.Tasks = normalizeRoadmapChecklist(phase.Tasks)
	phase.ExitCriteria = normalizeRoadmapChecklist(phase.ExitCriteria)
	phase.Evidence.ModuleInventory = normalizeRoadmapModuleInventory(phase.Evidence.ModuleInventory)
	phase.PhaseHash = strings.ToLower(strings.TrimSpace(phase.PhaseHash))
	return phase
}

func normalizeRoadmapChecklist(items []RoadmapChecklistItem) []RoadmapChecklistItem {
	out := make([]RoadmapChecklistItem, len(items))
	for i, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		item.Description = strings.TrimSpace(item.Description)
		out[i] = item
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func normalizeRoadmapModuleInventory(entries []RoadmapModuleInventoryEntry) []RoadmapModuleInventoryEntry {
	out := make([]RoadmapModuleInventoryEntry, len(entries))
	for i, entry := range entries {
		entry.ModuleName = CosmosSDKModuleName(strings.TrimSpace(string(entry.ModuleName)))
		entry.ModulePath = strings.TrimSpace(entry.ModulePath)
		entry.StoreKey = strings.TrimSpace(entry.StoreKey)
		entry.StateKeys = append([]string(nil), entry.StateKeys...)
		for j := range entry.StateKeys {
			entry.StateKeys[j] = strings.TrimSpace(entry.StateKeys[j])
		}
		sort.Strings(entry.StateKeys)
		entry.RootType = RootType(strings.TrimSpace(string(entry.RootType)))
		out[i] = entry
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ModuleName < out[j].ModuleName
	})
	return out
}

func validateRoadmapChecklist(field string, items []RoadmapChecklistItem) error {
	if len(items) == 0 {
		return fmt.Errorf("%s is required", field)
	}
	seen := make(map[string]struct{}, len(items))
	var previous string
	for i, item := range items {
		if err := validatePolicyID(field+" id", item.ID); err != nil {
			return err
		}
		if err := validateRoadmapText(field+" description", item.Description); err != nil {
			return err
		}
		if _, found := seen[item.ID]; found {
			return fmt.Errorf("duplicate %s %s", field, item.ID)
		}
		seen[item.ID] = struct{}{}
		if i > 0 && previous >= item.ID {
			return fmt.Errorf("%s must be sorted canonically", field)
		}
		previous = item.ID
	}
	return nil
}

func validateRoadmapModuleInventory(entries []RoadmapModuleInventoryEntry) error {
	required := RequiredCosmosSDKModules()
	if len(entries) != len(required) {
		return fmt.Errorf("aethercore roadmap inventory must include %d required modules", len(required))
	}
	seen := make(map[CosmosSDKModuleName]struct{}, len(entries))
	var previous CosmosSDKModuleName
	for i, entry := range entries {
		if !IsRequiredCosmosSDKModule(entry.ModuleName) {
			return fmt.Errorf("aethercore roadmap inventory has unknown module %q", entry.ModuleName)
		}
		if err := validateToken("aethercore roadmap inventory module path", entry.ModulePath, MaxScopeLength); err != nil {
			return err
		}
		if err := validateToken("aethercore roadmap inventory store key", entry.StoreKey, MaxScopeLength); err != nil {
			return err
		}
		if len(entry.StateKeys) == 0 {
			return fmt.Errorf("aethercore roadmap inventory %s state keys are required", entry.ModuleName)
		}
		for _, stateKey := range entry.StateKeys {
			if err := validateToken("aethercore roadmap inventory state key", stateKey, MaxScopeLength); err != nil {
				return err
			}
		}
		if err := validateToken("aethercore roadmap inventory root type", string(entry.RootType), MaxScopeLength); err != nil {
			return err
		}
		if _, found := seen[entry.ModuleName]; found {
			return fmt.Errorf("duplicate aethercore roadmap inventory module %s", entry.ModuleName)
		}
		seen[entry.ModuleName] = struct{}{}
		if i > 0 && previous >= entry.ModuleName {
			return errors.New("aethercore roadmap inventory modules must be sorted canonically")
		}
		previous = entry.ModuleName
	}
	for _, moduleName := range required {
		if _, found := seen[moduleName]; !found {
			return fmt.Errorf("missing aethercore roadmap inventory module %s", moduleName)
		}
	}
	return nil
}

func validateRoadmapText(field string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > 256 {
		return fmt.Errorf("%s must be <= 256 bytes", field)
	}
	return nil
}

func roadmapChecklistComplete(items []RoadmapChecklistItem) bool {
	for _, item := range items {
		if !item.Complete {
			return false
		}
	}
	return true
}

func writeRoadmapChecklist(w byteWriter, items []RoadmapChecklistItem) {
	items = normalizeRoadmapChecklist(items)
	writeUint64(w, uint64(len(items)))
	for _, item := range items {
		writePart(w, item.ID)
		writePart(w, item.Description)
		writePart(w, fmt.Sprint(item.Complete))
	}
}

func writeRoadmapEvidence(w byteWriter, evidence RoadmapEvidence) {
	inventory := normalizeRoadmapModuleInventory(evidence.ModuleInventory)
	writeUint64(w, uint64(len(inventory)))
	for _, entry := range inventory {
		writePart(w, string(entry.ModuleName))
		writePart(w, entry.ModulePath)
		writePart(w, entry.StoreKey)
		for _, stateKey := range entry.StateKeys {
			writePart(w, stateKey)
		}
		writePart(w, string(entry.RootType))
	}
	writePart(w, fmt.Sprint(evidence.CrossModuleDirectWritesAudited))
	writePart(w, fmt.Sprint(evidence.ExportImportTestsAdded))
	writePart(w, fmt.Sprint(evidence.ModuleInvariantHarnessAdded))
	writePart(w, fmt.Sprint(evidence.RootContributionInterfaceDesign))
	writePart(w, fmt.Sprint(evidence.CurrentStateReproducible))
	writePart(w, fmt.Sprint(evidence.ModuleBoundariesDocumented))
	writePart(w, fmt.Sprint(evidence.MigrationRiskListComplete))
	writePart(w, fmt.Sprint(evidence.AetherCoreModuleImplemented))
	writePart(w, fmt.Sprint(evidence.ZonesModuleImplemented))
	writePart(w, fmt.Sprint(evidence.ZoneRegistryImplemented))
	writePart(w, fmt.Sprint(evidence.GlobalStateRootImplemented))
	writePart(w, fmt.Sprint(evidence.BlockCommitmentMetadataQueries))
	writePart(w, fmt.Sprint(evidence.DefaultZoneRunnable))
	writePart(w, fmt.Sprint(evidence.DefaultZoneRootIncluded))
	writePart(w, fmt.Sprint(evidence.ExportImportPreservesRootMeta))
}
