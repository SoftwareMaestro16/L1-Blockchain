package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCosmosSDKNewModuleMapCoversSectionTenOne(t *testing.T) {
	require.NoError(t, ValidateCosmosSDKNewModuleMap())

	moduleMap, err := DefaultCosmosModuleMap()
	require.NoError(t, err)
	require.NoError(t, moduleMap.Validate())
	require.Len(t, moduleMap.Modules, 11)
	require.NotEmpty(t, moduleMap.Root)

	byModule := map[string]CosmosModuleDescriptor{}
	for _, module := range moduleMap.Modules {
		require.NoError(t, module.Validate())
		byModule[module.Module] = module
	}

	require.Equal(t, "Core", byModule["x/aetracore"].Zone)
	require.Contains(t, byModule["x/aetracore"].Responsibility, "Zone registry")
	require.Contains(t, byModule["x/aetracore"].AcceptanceSignal, "identical zone commitments")

	require.Equal(t, "Core + zones", byModule["x/msgbus"].Zone)
	require.True(t, slices.Contains(byModule["x/msgbus"].Dependencies, "x/aetracore"))
	require.True(t, slices.Contains(byModule["x/msgbus"].Dependencies, "x/proofregistry"))
	require.Contains(t, byModule["x/msgbus"].AcceptanceSignal, "message inclusion")

	require.Equal(t, "Financial Zone", byModule["x/payments"].Zone)
	require.Equal(t, []string{"x/msgbus", "x/proofregistry", "x/zonefees"}, byModule["x/payments"].Dependencies)
	require.Contains(t, byModule["x/payments"].AcceptanceSignal, "value-conserving receipts")

	require.Equal(t, "Node-side + core checks", byModule["x/zonemempool"].Zone)
	require.Contains(t, byModule["x/zonemempool"].AcceptanceSignal, "reject malformed schedules")
}

func TestCosmosModuleMapRootIsCanonicalAcrossInputOrder(t *testing.T) {
	modules := CosmosSDKNewModules()
	moduleMap, err := BuildCosmosModuleMap(modules)
	require.NoError(t, err)

	reordered := append([]CosmosModuleDescriptor(nil), modules...)
	slices.Reverse(reordered)
	reorderedMap, err := BuildCosmosModuleMap(reordered)
	require.NoError(t, err)
	require.Equal(t, moduleMap.Root, reorderedMap.Root)
	require.Equal(t, moduleMap.Modules, reorderedMap.Modules)
}

func TestCosmosModuleDescriptorValidationRejectsMalformedEntries(t *testing.T) {
	valid := CosmosSDKNewModules()[0]
	valid.DescriptorHash = ""
	require.NoError(t, valid.ValidateFormat())

	duplicate, err := BuildCosmosModuleMap([]CosmosModuleDescriptor{CosmosSDKNewModules()[0], CosmosSDKNewModules()[0]})
	require.ErrorContains(t, err, "duplicate cosmos module")
	require.Empty(t, duplicate.Root)

	noDeps := CosmosModuleDescriptor{
		Module:			"x/test",
		Responsibility:		"test module",
		Zone:			"Core",
		AcceptanceSignal:	"test acceptance",
	}
	_, err = BuildCosmosModuleDescriptor(noDeps)
	require.ErrorContains(t, err, "dependencies are required")

	tampered := CosmosSDKNewModules()[0]
	tampered.AcceptanceSignal = strings.ReplaceAll(tampered.AcceptanceSignal, "identical", "different")
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")
}
