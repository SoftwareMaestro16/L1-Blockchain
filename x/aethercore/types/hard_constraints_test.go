package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHardConstraintManifestCoversSection17(t *testing.T) {
	manifest, err := DefaultHardConstraintManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.Constraints, 10)
	require.Equal(t, ComputeHardConstraintManifestHash(manifest), manifest.ManifestHash)

	for _, id := range RequiredHardConstraintIDs() {
		constraint, found := HardConstraintByID(manifest, id)
		require.True(t, found, id)
		require.True(t, constraint.ConsensusCritical, id)
		require.True(t, constraint.StateTransitionGuard, id)
		require.True(t, constraint.CommitmentBacked, id)
		require.NotEmpty(t, constraint.EnforcementTarget, id)
		require.NotEmpty(t, constraint.Modules, id)
		require.NotEmpty(t, constraint.Evidence, id)
		require.Equal(t, ComputeHardConstraintHash(constraint), constraint.ConstraintHash)
		for _, moduleName := range constraint.Modules {
			require.True(t, IsRequiredCosmosSDKModule(moduleName), id)
		}
	}

	noNetwork, found := HardConstraintByID(manifest, HardConstraintNoExternalNetworkStateTransition)
	require.True(t, found)
	require.Equal(t, RequiredCosmosSDKModules(), noNetwork.Modules)
	require.Contains(t, noNetwork.Evidence, "external network calls are outside consensus")

	messageOnly, found := HardConstraintByID(manifest, HardConstraintCrossZoneInteractionsMessageOnly)
	require.True(t, found)
	require.Contains(t, messageOnly.Evidence, "cross-zone mutations route through x/messages")

	queueBound, found := HardConstraintByID(manifest, HardConstraintBoundedQueueDraining)
	require.True(t, found)
	require.Equal(t, []CosmosSDKModuleName{CosmosModuleMessages, CosmosModuleRouting, CosmosModuleZones}, queueBound.Modules)

	proofGas, found := HardConstraintByID(manifest, HardConstraintGasAccountedProofVerification)
	require.True(t, found)
	require.Equal(t, []CosmosSDKModuleName{CosmosModuleContracts, CosmosModuleMessages, CosmosModuleStorage}, proofGas.Modules)
}

func TestHardConstraintManifestRejectsMissingDuplicateAndMalformedConstraints(t *testing.T) {
	manifest, err := DefaultHardConstraintManifest()
	require.NoError(t, err)

	missing := manifest
	missing.Constraints = append([]HardConstraint(nil), manifest.Constraints[1:]...)
	missing.ManifestHash = ComputeHardConstraintManifestHash(missing)
	require.ErrorContains(t, missing.Validate(), "must include 10 required constraints")

	duplicate := manifest
	duplicate.Constraints = append([]HardConstraint(nil), manifest.Constraints...)
	duplicate.Constraints[len(duplicate.Constraints)-1] = duplicate.Constraints[0]
	duplicate.ManifestHash = ComputeHardConstraintManifestHash(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")

	unknownModule := manifest
	unknownModule.Constraints = append([]HardConstraint(nil), manifest.Constraints...)
	unknownModule.Constraints[0].Modules = []CosmosSDKModuleName{CosmosSDKModuleName("dex")}
	unknownModule.Constraints[0].ConstraintHash = ComputeHardConstraintHash(unknownModule.Constraints[0])
	unknownModule.ManifestHash = ComputeHardConstraintManifestHash(unknownModule)
	require.ErrorContains(t, unknownModule.Validate(), "unknown module")

	notConsensusCritical := manifest
	notConsensusCritical.Constraints = append([]HardConstraint(nil), manifest.Constraints...)
	notConsensusCritical.Constraints[0].ConsensusCritical = false
	notConsensusCritical.Constraints[0].ConstraintHash = ComputeHardConstraintHash(notConsensusCritical.Constraints[0])
	notConsensusCritical.ManifestHash = ComputeHardConstraintManifestHash(notConsensusCritical)
	require.ErrorContains(t, notConsensusCritical.Validate(), "consensus critical")

	noEvidence := manifest
	noEvidence.Constraints = append([]HardConstraint(nil), manifest.Constraints...)
	noEvidence.Constraints[0].Evidence = nil
	noEvidence.Constraints[0].ConstraintHash = ComputeHardConstraintHash(noEvidence.Constraints[0])
	noEvidence.ManifestHash = ComputeHardConstraintManifestHash(noEvidence)
	require.ErrorContains(t, noEvidence.Validate(), "evidence is required")

	hashMismatch := manifest
	hashMismatch.Constraints = append([]HardConstraint(nil), manifest.Constraints...)
	hashMismatch.Constraints[0].Evidence = append([]string(nil), hashMismatch.Constraints[0].Evidence...)
	hashMismatch.Constraints[0].Evidence[0] = "tampered evidence"
	hashMismatch.ManifestHash = ComputeHardConstraintManifestHash(hashMismatch)
	require.ErrorContains(t, hashMismatch.Validate(), "constraint hash mismatch")
}

func TestHardConstraintManifestHashIsCanonical(t *testing.T) {
	manifest, err := DefaultHardConstraintManifest()
	require.NoError(t, err)

	reversed := reverseHardConstraints(manifest.Constraints)
	for i := range reversed {
		reversed[i].Modules = reverseModules(reversed[i].Modules)
		reversed[i].Evidence = reverseStrings(reversed[i].Evidence)
		reversed[i].ConstraintHash = ""
	}
	reordered, err := NewHardConstraintManifest(reversed)
	require.NoError(t, err)
	require.Equal(t, manifest.ManifestHash, reordered.ManifestHash)
	require.Equal(t, manifest.Constraints, reordered.Constraints)

	tampered := manifest
	tampered.Constraints = append([]HardConstraint(nil), manifest.Constraints...)
	tampered.Constraints[0].EnforcementTarget = "tampered target"
	tampered.Constraints[0].ConstraintHash = ComputeHardConstraintHash(tampered.Constraints[0])
	tampered.ManifestHash = ComputeHardConstraintManifestHash(tampered)
	require.NoError(t, tampered.Validate())
	require.NotEqual(t, manifest.ManifestHash, tampered.ManifestHash)
}

func reverseHardConstraints(constraints []HardConstraint) []HardConstraint {
	out := make([]HardConstraint, len(constraints))
	for i := range constraints {
		out[i] = constraints[len(constraints)-1-i]
	}
	return out
}

func reverseModules(modules []CosmosSDKModuleName) []CosmosSDKModuleName {
	out := make([]CosmosSDKModuleName, len(modules))
	for i := range modules {
		out[i] = modules[len(modules)-1-i]
	}
	return out
}

func reverseStrings(values []string) []string {
	out := make([]string, len(values))
	for i := range values {
		out[i] = values[len(values)-1-i]
	}
	return out
}
