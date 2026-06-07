package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredTestCoverageManifestCoversSection16(t *testing.T) {
	manifest, err := DefaultRequiredTestCoverageManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.UnitTests, 12)
	require.Len(t, manifest.IntegrationTests, 8)
	require.Len(t, manifest.InvariantTests, 9)
	require.Len(t, manifest.SimulationTests, 8)
	require.Len(t, manifest.PerformanceTests, 9)
	require.Equal(t, ComputeRequiredTestCoverageManifestHash(manifest), manifest.ManifestHash)

	for _, id := range RequiredUnitTestCoverageIDs() {
		spec, found := RequiredUnitTestCoverageByID(manifest, id)
		require.True(t, found, id)
		require.Equal(t, TestCoverageKindUnit, spec.Kind)
		require.True(t, IsRequiredCosmosSDKModule(spec.ModuleName), id)
		require.True(t, IsImplementationRoadmapPhaseID(spec.PhaseID), id)
		require.NotEmpty(t, spec.CoverageTarget, id)
		require.NotEmpty(t, spec.Assertions, id)
		require.Equal(t, ComputeRequiredTestCoverageSpecHash(spec), spec.SpecHash)
	}

	for _, id := range RequiredIntegrationTestCoverageIDs() {
		spec, found := RequiredIntegrationTestCoverageByID(manifest, id)
		require.True(t, found, id)
		require.Equal(t, TestCoverageKindIntegration, spec.Kind)
		require.True(t, IsRequiredCosmosSDKModule(spec.ModuleName), id)
		require.True(t, IsImplementationRoadmapPhaseID(spec.PhaseID), id)
		require.NotEmpty(t, spec.CoverageTarget, id)
		require.NotEmpty(t, spec.Assertions, id)
		require.Equal(t, ComputeRequiredTestCoverageSpecHash(spec), spec.SpecHash)
	}

	for _, id := range RequiredInvariantTestCoverageIDs() {
		spec, found := RequiredInvariantTestCoverageByID(manifest, id)
		require.True(t, found, id)
		require.Equal(t, TestCoverageKindInvariant, spec.Kind)
		require.True(t, IsRequiredCosmosSDKModule(spec.ModuleName), id)
		require.True(t, IsImplementationRoadmapPhaseID(spec.PhaseID), id)
		require.NotEmpty(t, spec.CoverageTarget, id)
		require.NotEmpty(t, spec.Assertions, id)
		require.Equal(t, ComputeRequiredTestCoverageSpecHash(spec), spec.SpecHash)
	}

	for _, id := range RequiredSimulationTestCoverageIDs() {
		spec, found := RequiredSimulationTestCoverageByID(manifest, id)
		require.True(t, found, id)
		require.Equal(t, TestCoverageKindSimulation, spec.Kind)
		require.True(t, IsRequiredCosmosSDKModule(spec.ModuleName), id)
		require.True(t, IsImplementationRoadmapPhaseID(spec.PhaseID), id)
		require.NotEmpty(t, spec.CoverageTarget, id)
		require.NotEmpty(t, spec.Assertions, id)
		require.Equal(t, ComputeRequiredTestCoverageSpecHash(spec), spec.SpecHash)
	}

	for _, id := range RequiredPerformanceTestCoverageIDs() {
		spec, found := RequiredPerformanceTestCoverageByID(manifest, id)
		require.True(t, found, id)
		require.Equal(t, TestCoverageKindPerformance, spec.Kind)
		require.True(t, IsRequiredCosmosSDKModule(spec.ModuleName), id)
		require.True(t, IsImplementationRoadmapPhaseID(spec.PhaseID), id)
		require.NotEmpty(t, spec.CoverageTarget, id)
		require.NotEmpty(t, spec.Assertions, id)
		require.Equal(t, ComputeRequiredTestCoverageSpecHash(spec), spec.SpecHash)
	}

	messageID, found := RequiredUnitTestCoverageByID(manifest, UnitCoverageMessageIDDerivation)
	require.True(t, found)
	require.Equal(t, CosmosModuleMessages, messageID.ModuleName)
	require.Equal(t, RoadmapPhaseCrossZoneMessages, messageID.PhaseID)
	require.Contains(t, messageID.Assertions, "same canonical envelope derives identical id")

	rootEncoding, found := RequiredUnitTestCoverageByID(manifest, UnitCoverageRootEncoding)
	require.True(t, found)
	require.Equal(t, CosmosModuleAetraCore, rootEncoding.ModuleName)

	payment, found := RequiredIntegrationTestCoverageByID(manifest, IntegrationCoverageCrossZoneIdentityBoundPayment)
	require.True(t, found)
	require.Equal(t, CosmosModulePayments, payment.ModuleName)
	require.Equal(t, RoadmapPhaseIdentityPaymentIntegration, payment.PhaseID)

	contractOutbound, found := RequiredIntegrationTestCoverageByID(manifest, IntegrationCoverageContractOutboundMessageFinancial)
	require.True(t, found)
	require.Equal(t, CosmosModuleContracts, contractOutbound.ModuleName)
	require.Equal(t, RoadmapPhaseVMRuntime, contractOutbound.PhaseID)

	globalRootInvariant, found := RequiredInvariantTestCoverageByID(manifest, InvariantCoverageGlobalRootIncludesEnabledZones)
	require.True(t, found)
	require.Equal(t, CosmosModuleAetraCore, globalRootInvariant.ModuleName)
	require.Equal(t, RoadmapPhaseKernelRootModel, globalRootInvariant.PhaseID)

	escrowInvariant, found := RequiredInvariantTestCoverageByID(manifest, InvariantCoveragePaymentSettlementWithinEscrow)
	require.True(t, found)
	require.Equal(t, CosmosModulePayments, escrowInvariant.ModuleName)
	require.Equal(t, RoadmapPhaseIdentityPaymentIntegration, escrowInvariant.PhaseID)

	blockSTMSimulation, found := RequiredSimulationTestCoverageByID(manifest, SimulationCoverageMixedZoneExecutionUnderBlockSTM)
	require.True(t, found)
	require.Equal(t, CosmosModuleZones, blockSTMSimulation.ModuleName)
	require.Equal(t, RoadmapPhasePerformanceHardening, blockSTMSimulation.PhaseID)

	rootAggregationPerf, found := RequiredPerformanceTestCoverageByID(manifest, PerformanceCoverageRootAggregationCostPerZone)
	require.True(t, found)
	require.Equal(t, CosmosModuleAetraCore, rootAggregationPerf.ModuleName)
	require.Equal(t, RoadmapPhasePerformanceHardening, rootAggregationPerf.PhaseID)

	serviceLookupPerf, found := RequiredPerformanceTestCoverageByID(manifest, PerformanceCoverageServiceLookupLatency)
	require.True(t, found)
	require.Equal(t, CosmosModuleServices, serviceLookupPerf.ModuleName)
	require.Equal(t, RoadmapPhasePerformanceHardening, serviceLookupPerf.PhaseID)
}

func TestRequiredTestCoverageManifestRejectsMissingDuplicateAndMalformedCoverage(t *testing.T) {
	manifest, err := DefaultRequiredTestCoverageManifest()
	require.NoError(t, err)

	missingUnit := manifest
	missingUnit.UnitTests = append([]RequiredTestCoverageSpec(nil), manifest.UnitTests[1:]...)
	missingUnit.ManifestHash = ComputeRequiredTestCoverageManifestHash(missingUnit)
	require.ErrorContains(t, missingUnit.Validate(), "must include 12 required coverage areas")

	duplicateIntegration := manifest
	duplicateIntegration.IntegrationTests = append([]RequiredTestCoverageSpec(nil), manifest.IntegrationTests...)
	duplicateIntegration.IntegrationTests[len(duplicateIntegration.IntegrationTests)-1] = duplicateIntegration.IntegrationTests[0]
	duplicateIntegration.ManifestHash = ComputeRequiredTestCoverageManifestHash(duplicateIntegration)
	require.ErrorContains(t, duplicateIntegration.Validate(), "duplicate")

	unknownModule := manifest
	unknownModule.UnitTests = append([]RequiredTestCoverageSpec(nil), manifest.UnitTests...)
	unknownModule.UnitTests[0].ModuleName = CosmosSDKModuleName("dex")
	unknownModule.UnitTests[0].SpecHash = ComputeRequiredTestCoverageSpecHash(unknownModule.UnitTests[0])
	unknownModule.ManifestHash = ComputeRequiredTestCoverageManifestHash(unknownModule)
	require.ErrorContains(t, unknownModule.Validate(), "unknown module")

	unknownPhase := manifest
	unknownPhase.IntegrationTests = append([]RequiredTestCoverageSpec(nil), manifest.IntegrationTests...)
	unknownPhase.IntegrationTests[0].PhaseID = ImplementationRoadmapPhaseID("phase-99")
	unknownPhase.IntegrationTests[0].SpecHash = ComputeRequiredTestCoverageSpecHash(unknownPhase.IntegrationTests[0])
	unknownPhase.ManifestHash = ComputeRequiredTestCoverageManifestHash(unknownPhase)
	require.ErrorContains(t, unknownPhase.Validate(), "unknown roadmap phase")

	missingInvariant := manifest
	missingInvariant.InvariantTests = append([]RequiredTestCoverageSpec(nil), manifest.InvariantTests[1:]...)
	missingInvariant.ManifestHash = ComputeRequiredTestCoverageManifestHash(missingInvariant)
	require.ErrorContains(t, missingInvariant.Validate(), "must include 9 required coverage areas")

	duplicateSimulation := manifest
	duplicateSimulation.SimulationTests = append([]RequiredTestCoverageSpec(nil), manifest.SimulationTests...)
	duplicateSimulation.SimulationTests[len(duplicateSimulation.SimulationTests)-1] = duplicateSimulation.SimulationTests[0]
	duplicateSimulation.ManifestHash = ComputeRequiredTestCoverageManifestHash(duplicateSimulation)
	require.ErrorContains(t, duplicateSimulation.Validate(), "duplicate")

	missingPerformance := manifest
	missingPerformance.PerformanceTests = append([]RequiredTestCoverageSpec(nil), manifest.PerformanceTests[1:]...)
	missingPerformance.ManifestHash = ComputeRequiredTestCoverageManifestHash(missingPerformance)
	require.ErrorContains(t, missingPerformance.Validate(), "must include 9 required coverage areas")

	noAssertions := manifest
	noAssertions.UnitTests = append([]RequiredTestCoverageSpec(nil), manifest.UnitTests...)
	noAssertions.UnitTests[0].Assertions = nil
	noAssertions.UnitTests[0].SpecHash = ComputeRequiredTestCoverageSpecHash(noAssertions.UnitTests[0])
	noAssertions.ManifestHash = ComputeRequiredTestCoverageManifestHash(noAssertions)
	require.ErrorContains(t, noAssertions.Validate(), "assertions are required")

	hashMismatch := manifest
	hashMismatch.UnitTests = append([]RequiredTestCoverageSpec(nil), manifest.UnitTests...)
	hashMismatch.UnitTests[0].Assertions = append([]string(nil), hashMismatch.UnitTests[0].Assertions...)
	hashMismatch.UnitTests[0].Assertions[0] = "tampered assertion"
	hashMismatch.ManifestHash = ComputeRequiredTestCoverageManifestHash(hashMismatch)
	require.ErrorContains(t, hashMismatch.Validate(), "spec hash mismatch")
}

func TestRequiredTestCoverageManifestHashIsCanonical(t *testing.T) {
	manifest, err := DefaultRequiredTestCoverageManifest()
	require.NoError(t, err)

	reversedUnits := reverseCoverageSpecs(manifest.UnitTests)
	reversedIntegration := reverseCoverageSpecs(manifest.IntegrationTests)
	reversedInvariants := reverseCoverageSpecs(manifest.InvariantTests)
	reversedSimulations := reverseCoverageSpecs(manifest.SimulationTests)
	reversedPerformance := reverseCoverageSpecs(manifest.PerformanceTests)
	reordered, err := NewRequiredTestCoverageManifest(reversedUnits, reversedIntegration, reversedInvariants, reversedSimulations, reversedPerformance)
	require.NoError(t, err)
	require.Equal(t, manifest.ManifestHash, reordered.ManifestHash)
	require.Equal(t, manifest.UnitTests, reordered.UnitTests)
	require.Equal(t, manifest.IntegrationTests, reordered.IntegrationTests)
	require.Equal(t, manifest.InvariantTests, reordered.InvariantTests)
	require.Equal(t, manifest.SimulationTests, reordered.SimulationTests)
	require.Equal(t, manifest.PerformanceTests, reordered.PerformanceTests)

	tampered := manifest
	tampered.IntegrationTests = append([]RequiredTestCoverageSpec(nil), manifest.IntegrationTests...)
	tampered.IntegrationTests[0].CoverageTarget = "tampered target"
	tampered.IntegrationTests[0].SpecHash = ComputeRequiredTestCoverageSpecHash(tampered.IntegrationTests[0])
	tampered.ManifestHash = ComputeRequiredTestCoverageManifestHash(tampered)
	require.NoError(t, tampered.Validate())
	require.NotEqual(t, manifest.ManifestHash, tampered.ManifestHash)
}

func reverseCoverageSpecs(specs []RequiredTestCoverageSpec) []RequiredTestCoverageSpec {
	out := make([]RequiredTestCoverageSpec, len(specs))
	for i := range specs {
		out[i] = specs[len(specs)-1-i]
	}
	return out
}
