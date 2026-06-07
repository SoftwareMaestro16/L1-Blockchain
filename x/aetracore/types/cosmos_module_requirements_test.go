package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCosmosSDKModuleRequirementManifestCoversRequiredModules(t *testing.T) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.Modules, 9)
	require.Equal(t, RequiredCosmosSDKModules(), moduleNamesFromManifest(manifest))

	for _, surface := range manifest.Modules {
		require.True(t, surface.MsgServer, surface.ModuleName)
		require.True(t, surface.QueryServer, surface.ModuleName)
		require.True(t, surface.Keeper, surface.ModuleName)
		require.True(t, surface.Params, surface.ModuleName)
		require.True(t, surface.GenesisExport, surface.ModuleName)
		require.True(t, surface.GenesisImport, surface.ModuleName)
		require.True(t, surface.Invariants, surface.ModuleName)
		require.NoError(t, surface.KeeperIsolation.Validate(surface.ModuleName))
		require.Equal(t, string(surface.ModuleName), surface.KeeperIsolation.StoreKey)
		require.Equal(t, []string{string(surface.ModuleName)}, surface.KeeperIsolation.WritableStoreKeys)
		require.True(t, surface.KeeperIsolation.CrossZoneWritesProhibited, surface.ModuleName)
		require.Equal(t, CosmosModuleMessages, surface.KeeperIsolation.CrossZoneMessagesModule)
		require.True(t, surface.KeeperIsolation.DirectCallsLimitedToLocalHelpers, surface.ModuleName)
		require.True(t, surface.KeeperIsolation.SharedStateReadOnlyOrProofBacked, surface.ModuleName)
		require.NoError(t, surface.IBCBoundary.Validate(surface.ModuleName))
		require.True(t, surface.IBCBoundary.StateExportable, surface.ModuleName)
		require.True(t, surface.IBCBoundary.ReceiptsProofVerifiable, surface.ModuleName)
		require.True(t, surface.IBCBoundary.CanonicalBoundaryMessages, surface.ModuleName)
		require.True(t, surface.IBCBoundary.TimeoutRulesExplicit, surface.ModuleName)
		require.True(t, surface.IBCBoundary.ReplayRulesExplicit, surface.ModuleName)
		require.True(t, surface.IBCBoundary.DeterministicChannelRouting, surface.ModuleName)
		require.NoError(t, surface.ABCICompatibility.Validate(surface.ModuleName))
		require.True(t, surface.ABCICompatibility.ProposalOptimizationValidityNeutral, surface.ModuleName)
		require.True(t, surface.ABCICompatibility.PrecheckDeterministic, surface.ModuleName)
		require.True(t, surface.ABCICompatibility.FinalizeBlockAuthoritative, surface.ModuleName)
		require.True(t, surface.ABCICompatibility.EndBlockCleanupBounded, surface.ModuleName)
		require.True(t, surface.ABCICompatibility.RootAggregationAfterExecution, surface.ModuleName)
		require.Equal(t, KernelPhaseFinalizeBlock, surface.ABCICompatibility.RootAggregationPhase)
		require.NotEmpty(t, surface.Events, surface.ModuleName)
		require.NotEmpty(t, surface.TypedErrors, surface.ModuleName)
		require.NoError(t, surface.RootContribution.Validate())
		require.Equal(t, ComputeCosmosModuleSurfaceHash(surface), surface.SurfaceHash)
	}
	require.Equal(t, ComputeCosmosModuleRequirementManifestHash(manifest), manifest.ManifestHash)
}

func TestCosmosSDKModuleRequirementManifestRejectsMissingDuplicateAndTamperedSurface(t *testing.T) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	require.NoError(t, err)

	_, err = NewCosmosModuleRequirementManifest(manifest.Modules[1:])
	require.ErrorContains(t, err, "must include 9 required modules")

	duplicate := append([]CosmosModuleSurface(nil), manifest.Modules...)
	duplicate[len(duplicate)-1] = duplicate[0]
	_, err = NewCosmosModuleRequirementManifest(duplicate)
	require.ErrorContains(t, err, "duplicate")

	tampered := manifest
	tampered.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	tampered.Modules[0].Keeper = false
	tampered.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(tampered.Modules[0])
	tampered.ManifestHash = ComputeCosmosModuleRequirementManifestHash(tampered)
	require.ErrorContains(t, tampered.Validate(), "missing required module surface")
}

func TestCosmosSDKModuleRequirementManifestRejectsKeeperIsolationViolations(t *testing.T) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	require.NoError(t, err)

	crossZoneWrite := manifest
	crossZoneWrite.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	crossZoneWrite.Modules[0].KeeperIsolation.WritableStoreKeys = []string{string(crossZoneWrite.Modules[0].ModuleName), "messages"}
	crossZoneWrite.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(crossZoneWrite.Modules[0])
	crossZoneWrite.ManifestHash = ComputeCosmosModuleRequirementManifestHash(crossZoneWrite)
	require.ErrorContains(t, crossZoneWrite.Validate(), "must write only its own store key")

	ungrantedRead := manifest
	ungrantedRead.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	ungrantedRead.Modules[0].KeeperIsolation.ReadableStoreKeys = []string{string(ungrantedRead.Modules[0].ModuleName), "storage"}
	ungrantedRead.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(ungrantedRead.Modules[0])
	ungrantedRead.ManifestHash = ComputeCosmosModuleRequirementManifestHash(ungrantedRead)
	require.ErrorContains(t, ungrantedRead.Validate(), "without explicit capability")

	wrongTransport := manifest
	wrongTransport.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	wrongTransport.Modules[0].KeeperIsolation.CrossZoneMessagesModule = CosmosModuleRouting
	wrongTransport.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(wrongTransport.Modules[0])
	wrongTransport.ManifestHash = ComputeCosmosModuleRequirementManifestHash(wrongTransport)
	require.ErrorContains(t, wrongTransport.Validate(), "through x/messages")
}

func TestCosmosSDKModuleRequirementManifestRejectsIBCReadyBoundaryViolations(t *testing.T) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	require.NoError(t, err)

	notExportable := manifest
	notExportable.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	notExportable.Modules[0].IBCBoundary.StateExportable = false
	notExportable.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(notExportable.Modules[0])
	notExportable.ManifestHash = ComputeCosmosModuleRequirementManifestHash(notExportable)
	require.ErrorContains(t, notExportable.Validate(), "must export module state")

	noReceiptProof := manifest
	noReceiptProof.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	noReceiptProof.Modules[0].IBCBoundary.ReceiptsProofVerifiable = false
	noReceiptProof.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(noReceiptProof.Modules[0])
	noReceiptProof.ManifestHash = ComputeCosmosModuleRequirementManifestHash(noReceiptProof)
	require.ErrorContains(t, noReceiptProof.Validate(), "must make receipts proof-verifiable")

	nondeterministicRouting := manifest
	nondeterministicRouting.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	nondeterministicRouting.Modules[0].IBCBoundary.DeterministicChannelRouting = false
	nondeterministicRouting.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(nondeterministicRouting.Modules[0])
	nondeterministicRouting.ManifestHash = ComputeCosmosModuleRequirementManifestHash(nondeterministicRouting)
	require.ErrorContains(t, nondeterministicRouting.Validate(), "nondeterministic node routing state")

	noReplayPolicy := manifest
	noReplayPolicy.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	noReplayPolicy.Modules[0].IBCBoundary.ReplayRulesExplicit = false
	noReplayPolicy.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(noReplayPolicy.Modules[0])
	noReplayPolicy.ManifestHash = ComputeCosmosModuleRequirementManifestHash(noReplayPolicy)
	require.ErrorContains(t, noReplayPolicy.Validate(), "explicit replay rules")
}

func TestCosmosSDKModuleRequirementManifestRejectsABCICompatibilityViolations(t *testing.T) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	require.NoError(t, err)

	optimizationChangesValidity := manifest
	optimizationChangesValidity.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	optimizationChangesValidity.Modules[0].ABCICompatibility.ProposalOptimizationValidityNeutral = false
	optimizationChangesValidity.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(optimizationChangesValidity.Modules[0])
	optimizationChangesValidity.ManifestHash = ComputeCosmosModuleRequirementManifestHash(optimizationChangesValidity)
	require.ErrorContains(t, optimizationChangesValidity.Validate(), "proposal optimization validity-neutral")

	nondeterministicPrecheck := manifest
	nondeterministicPrecheck.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	nondeterministicPrecheck.Modules[0].ABCICompatibility.PrecheckDeterministic = false
	nondeterministicPrecheck.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(nondeterministicPrecheck.Modules[0])
	nondeterministicPrecheck.ManifestHash = ComputeCosmosModuleRequirementManifestHash(nondeterministicPrecheck)
	require.ErrorContains(t, nondeterministicPrecheck.Validate(), "precheck deterministic")

	nonAuthoritativeFinalize := manifest
	nonAuthoritativeFinalize.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	nonAuthoritativeFinalize.Modules[0].ABCICompatibility.FinalizeBlockAuthoritative = false
	nonAuthoritativeFinalize.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(nonAuthoritativeFinalize.Modules[0])
	nonAuthoritativeFinalize.ManifestHash = ComputeCosmosModuleRequirementManifestHash(nonAuthoritativeFinalize)
	require.ErrorContains(t, nonAuthoritativeFinalize.Validate(), "FinalizeBlock authoritative")

	unboundedCleanup := manifest
	unboundedCleanup.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	unboundedCleanup.Modules[0].ABCICompatibility.EndBlockCleanupBounded = false
	unboundedCleanup.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(unboundedCleanup.Modules[0])
	unboundedCleanup.ManifestHash = ComputeCosmosModuleRequirementManifestHash(unboundedCleanup)
	require.ErrorContains(t, unboundedCleanup.Validate(), "bound end-block cleanup")

	wrongRootPhase := manifest
	wrongRootPhase.Modules = append([]CosmosModuleSurface(nil), manifest.Modules...)
	wrongRootPhase.Modules[0].ABCICompatibility.RootAggregationPhase = KernelPhaseCommit
	wrongRootPhase.Modules[0].SurfaceHash = ComputeCosmosModuleSurfaceHash(wrongRootPhase.Modules[0])
	wrongRootPhase.ManifestHash = ComputeCosmosModuleRequirementManifestHash(wrongRootPhase)
	require.ErrorContains(t, wrongRootPhase.Validate(), "aggregate roots in FinalizeBlock")
}

func TestCosmosSDKModuleRequirementClassifiers(t *testing.T) {
	require.True(t, IsRequiredCosmosSDKModule(CosmosModuleContracts))
	require.False(t, IsRequiredCosmosSDKModule(CosmosSDKModuleName("dex")))
}

func moduleNamesFromManifest(manifest CosmosModuleRequirementManifest) []CosmosSDKModuleName {
	out := make([]CosmosSDKModuleName, len(manifest.Modules))
	for i, module := range manifest.Modules {
		out[i] = module.ModuleName
	}
	return out
}
