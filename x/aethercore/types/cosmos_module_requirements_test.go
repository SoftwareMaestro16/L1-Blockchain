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
