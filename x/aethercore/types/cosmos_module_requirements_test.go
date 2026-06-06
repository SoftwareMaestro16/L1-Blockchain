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
