package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultImplementationRoadmapCoversPhaseZeroAndOne(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)
	require.NoError(t, roadmap.Validate())
	require.Len(t, roadmap.Phases, 2)
	require.Equal(t, RoadmapPhaseBaselineAudit, roadmap.Phases[0].PhaseID)
	require.Equal(t, uint32(0), roadmap.Phases[0].PhaseNumber)
	require.Equal(t, RoadmapPhaseKernelRootModel, roadmap.Phases[1].PhaseID)
	require.Equal(t, uint32(1), roadmap.Phases[1].PhaseNumber)
	require.Equal(t, ComputeImplementationRoadmapHash(roadmap), roadmap.RoadmapHash)

	phase0 := roadmap.Phases[0]
	requireRoadmapTask(t, phase0, "inventory-current-modules-state-keys")
	requireRoadmapTask(t, phase0, "identify-cross-module-direct-writes")
	requireRoadmapTask(t, phase0, "add-export-import-tests-current-state")
	requireRoadmapTask(t, phase0, "add-module-invariant-test-harness")
	requireRoadmapTask(t, phase0, "add-root-contribution-interface-design")
	requireRoadmapExit(t, phase0, "current-aetheris-state-reproducible")
	requireRoadmapExit(t, phase0, "current-module-boundaries-documented")
	requireRoadmapExit(t, phase0, "migration-risk-list-complete")
	require.Len(t, phase0.Evidence.ModuleInventory, len(RequiredCosmosSDKModules()))
	require.True(t, phase0.Evidence.CrossModuleDirectWritesAudited)
	require.True(t, phase0.Evidence.ExportImportTestsAdded)
	require.True(t, phase0.Evidence.ModuleInvariantHarnessAdded)
	require.True(t, phase0.Evidence.RootContributionInterfaceDesign)
	require.True(t, phase0.Evidence.CurrentStateReproducible)
	require.True(t, phase0.Evidence.ModuleBoundariesDocumented)
	require.True(t, phase0.Evidence.MigrationRiskListComplete)

	phase1 := roadmap.Phases[1]
	requireRoadmapTask(t, phase1, "implement-x-aethercore")
	requireRoadmapTask(t, phase1, "implement-x-zones")
	requireRoadmapTask(t, phase1, "add-zone-registry")
	requireRoadmapTask(t, phase1, "add-root-contribution-interface")
	requireRoadmapTask(t, phase1, "add-global-state-root")
	requireRoadmapTask(t, phase1, "add-block-commitment-metadata-queries")
	requireRoadmapExit(t, phase1, "existing-chain-runs-as-default-zone")
	requireRoadmapExit(t, phase1, "global-root-includes-default-zone-root")
	requireRoadmapExit(t, phase1, "export-import-preserves-root-metadata")
	require.True(t, phase1.Evidence.AetherCoreModuleImplemented)
	require.True(t, phase1.Evidence.ZonesModuleImplemented)
	require.True(t, phase1.Evidence.ZoneRegistryImplemented)
	require.True(t, phase1.Evidence.GlobalStateRootImplemented)
	require.True(t, phase1.Evidence.BlockCommitmentMetadataQueries)
	require.True(t, phase1.Evidence.DefaultZoneRunnable)
	require.True(t, phase1.Evidence.DefaultZoneRootIncluded)
	require.True(t, phase1.Evidence.ExportImportPreservesRootMeta)
}

func TestImplementationRoadmapInventoryIsDerivedFromModuleManifest(t *testing.T) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	require.NoError(t, err)

	inventory := BuildRoadmapModuleInventory(manifest)
	require.Len(t, inventory, len(RequiredCosmosSDKModules()))
	require.Equal(t, RequiredCosmosSDKModules(), roadmapInventoryModuleNames(inventory))

	for _, entry := range inventory {
		require.NotEmpty(t, entry.ModulePath)
		require.Equal(t, string(entry.ModuleName), entry.StoreKey)
		require.Contains(t, entry.StateKeys, string(entry.ModuleName)+"/params")
		require.Contains(t, entry.StateKeys, string(entry.ModuleName)+"/genesis")
		require.Contains(t, entry.StateKeys, string(entry.ModuleName)+"/root")
		require.NotEmpty(t, entry.RootType)
	}
}

func TestImplementationRoadmapRejectsIncompletePhaseZeroEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	missingInventory := roadmap
	missingInventory.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	missingInventory.Phases[0].Evidence.ModuleInventory = missingInventory.Phases[0].Evidence.ModuleInventory[1:]
	missingInventory.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(missingInventory.Phases[0])
	missingInventory.RoadmapHash = ComputeImplementationRoadmapHash(missingInventory)
	require.ErrorContains(t, missingInventory.Validate(), "must include 9 required modules")

	noExportImportTests := roadmap
	noExportImportTests.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noExportImportTests.Phases[0].Evidence.ExportImportTestsAdded = false
	noExportImportTests.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(noExportImportTests.Phases[0])
	noExportImportTests.RoadmapHash = ComputeImplementationRoadmapHash(noExportImportTests)
	require.ErrorContains(t, noExportImportTests.Validate(), "export/import tests")

	incompleteExit := roadmap
	incompleteExit.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	incompleteExit.Phases[0].ExitCriteria[0].Complete = false
	incompleteExit.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(incompleteExit.Phases[0])
	incompleteExit.RoadmapHash = ComputeImplementationRoadmapHash(incompleteExit)
	require.ErrorContains(t, incompleteExit.Validate(), "incomplete exit criteria")
}

func TestImplementationRoadmapRejectsIncompletePhaseOneEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noZones := roadmap
	noZones.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noZones.Phases[1].Evidence.ZonesModuleImplemented = false
	noZones.Phases[1].PhaseHash = ComputeRoadmapPhaseHash(noZones.Phases[1])
	noZones.RoadmapHash = ComputeImplementationRoadmapHash(noZones)
	require.ErrorContains(t, noZones.Validate(), "x/zones")

	noGlobalRoot := roadmap
	noGlobalRoot.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noGlobalRoot.Phases[1].Evidence.GlobalStateRootImplemented = false
	noGlobalRoot.Phases[1].PhaseHash = ComputeRoadmapPhaseHash(noGlobalRoot.Phases[1])
	noGlobalRoot.RoadmapHash = ComputeImplementationRoadmapHash(noGlobalRoot)
	require.ErrorContains(t, noGlobalRoot.Validate(), "GlobalStateRoot")

	noRootMetadataRoundTrip := roadmap
	noRootMetadataRoundTrip.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noRootMetadataRoundTrip.Phases[1].Evidence.ExportImportPreservesRootMeta = false
	noRootMetadataRoundTrip.Phases[1].PhaseHash = ComputeRoadmapPhaseHash(noRootMetadataRoundTrip.Phases[1])
	noRootMetadataRoundTrip.RoadmapHash = ComputeImplementationRoadmapHash(noRootMetadataRoundTrip)
	require.ErrorContains(t, noRootMetadataRoundTrip.Validate(), "root metadata preservation")
}

func TestImplementationRoadmapHashIsCanonical(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	reversed, err := NewImplementationRoadmap([]ImplementationRoadmapPhase{roadmap.Phases[1], roadmap.Phases[0]})
	require.NoError(t, err)
	require.Equal(t, roadmap.RoadmapHash, reversed.RoadmapHash)
	require.Equal(t, roadmap.Phases, reversed.Phases)

	tampered := roadmap
	tampered.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	tampered.Phases[0].Tasks[0].Description = "tampered baseline task"
	tampered.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(tampered.Phases[0])
	tampered.RoadmapHash = ComputeImplementationRoadmapHash(tampered)
	require.NotEqual(t, roadmap.RoadmapHash, tampered.RoadmapHash)
	require.NoError(t, tampered.Validate())
}

func requireRoadmapTask(t *testing.T, phase ImplementationRoadmapPhase, id string) {
	t.Helper()
	for _, task := range phase.Tasks {
		if task.ID == id {
			require.True(t, task.Complete)
			return
		}
	}
	t.Fatalf("missing roadmap task %s", id)
}

func requireRoadmapExit(t *testing.T, phase ImplementationRoadmapPhase, id string) {
	t.Helper()
	for _, criterion := range phase.ExitCriteria {
		if criterion.ID == id {
			require.True(t, criterion.Complete)
			return
		}
	}
	t.Fatalf("missing roadmap exit criterion %s", id)
}

func roadmapInventoryModuleNames(entries []RoadmapModuleInventoryEntry) []CosmosSDKModuleName {
	entries = normalizeRoadmapModuleInventory(entries)
	out := make([]CosmosSDKModuleName, len(entries))
	for i, entry := range entries {
		out[i] = entry.ModuleName
	}
	return out
}
