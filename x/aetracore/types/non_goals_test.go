package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonGoalManifestCoversSection18(t *testing.T) {
	manifest, err := DefaultNonGoalManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.NonGoals, 6)
	require.Equal(t, ComputeNonGoalManifestHash(manifest), manifest.ManifestHash)

	for _, id := range RequiredNonGoalIDs() {
		nonGoal, found := NonGoalByID(manifest, id)
		require.True(t, found, id)
		require.True(t, nonGoal.Forbidden, id)
		require.True(t, nonGoal.ConsensusBoundary, id)
		require.NotEmpty(t, nonGoal.Boundary, id)
		require.NotEmpty(t, nonGoal.Modules, id)
		require.NotEmpty(t, nonGoal.Rationale, id)
		require.Equal(t, ComputeNonGoalHash(nonGoal), nonGoal.NonGoalHash)
		for _, moduleName := range nonGoal.Modules {
			require.True(t, IsRequiredCosmosSDKModule(moduleName), id)
		}
	}

	social, found := NonGoalByID(manifest, NonGoalNoMessagingSocialApplicationLayer)
	require.True(t, found)
	require.Equal(t, RequiredCosmosSDKModules(), social.Modules)
	require.Contains(t, social.Rationale, "social messaging products remain outside Aether Core scope")

	ui, found := NonGoalByID(manifest, NonGoalNoUIAssumptions)
	require.True(t, found)
	require.Equal(t, []CosmosSDKModuleName{CosmosModuleAetraCore, CosmosModuleServices, CosmosModuleZones}, ui.Modules)

	offchain, found := NonGoalByID(manifest, NonGoalNoCanonicalOffchainResultWithoutProof)
	require.True(t, found)
	require.True(t, offchain.CommitmentRequired)
	require.Equal(t, []CosmosSDKModuleName{CosmosModuleMessages, CosmosModuleServices, CosmosModuleStorage}, offchain.Modules)
	require.Contains(t, offchain.Rationale, "canonical result requires committed proof or receipt")

	synchronous, found := NonGoalByID(manifest, NonGoalNoDirectSynchronousCrossZoneFunctionCall)
	require.True(t, found)
	require.Equal(t, RequiredCosmosSDKModules(), synchronous.Modules)
	require.Contains(t, synchronous.Rationale, "synchronous cross-zone keeper calls cannot mutate state")
}

func TestNonGoalManifestRejectsMissingDuplicateAndMalformedNonGoals(t *testing.T) {
	manifest, err := DefaultNonGoalManifest()
	require.NoError(t, err)

	missing := manifest
	missing.NonGoals = append([]NonGoal(nil), manifest.NonGoals[1:]...)
	missing.ManifestHash = ComputeNonGoalManifestHash(missing)
	require.ErrorContains(t, missing.Validate(), "must include 6 required non-goals")

	duplicate := manifest
	duplicate.NonGoals = append([]NonGoal(nil), manifest.NonGoals...)
	duplicate.NonGoals[len(duplicate.NonGoals)-1] = duplicate.NonGoals[0]
	duplicate.ManifestHash = ComputeNonGoalManifestHash(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")

	unknownModule := manifest
	unknownModule.NonGoals = append([]NonGoal(nil), manifest.NonGoals...)
	unknownModule.NonGoals[0].Modules = []CosmosSDKModuleName{CosmosSDKModuleName("dex")}
	unknownModule.NonGoals[0].NonGoalHash = ComputeNonGoalHash(unknownModule.NonGoals[0])
	unknownModule.ManifestHash = ComputeNonGoalManifestHash(unknownModule)
	require.ErrorContains(t, unknownModule.Validate(), "unknown module")

	notForbidden := manifest
	notForbidden.NonGoals = append([]NonGoal(nil), manifest.NonGoals...)
	notForbidden.NonGoals[0].Forbidden = false
	notForbidden.NonGoals[0].NonGoalHash = ComputeNonGoalHash(notForbidden.NonGoals[0])
	notForbidden.ManifestHash = ComputeNonGoalManifestHash(notForbidden)
	require.ErrorContains(t, notForbidden.Validate(), "must be forbidden")

	noRationale := manifest
	noRationale.NonGoals = append([]NonGoal(nil), manifest.NonGoals...)
	noRationale.NonGoals[0].Rationale = nil
	noRationale.NonGoals[0].NonGoalHash = ComputeNonGoalHash(noRationale.NonGoals[0])
	noRationale.ManifestHash = ComputeNonGoalManifestHash(noRationale)
	require.ErrorContains(t, noRationale.Validate(), "rationale is required")

	noCommitment := manifest
	noCommitment.NonGoals = append([]NonGoal(nil), manifest.NonGoals...)
	offchain, found := NonGoalByID(noCommitment, NonGoalNoCanonicalOffchainResultWithoutProof)
	require.True(t, found)
	for i := range noCommitment.NonGoals {
		if noCommitment.NonGoals[i].ID == offchain.ID {
			noCommitment.NonGoals[i].CommitmentRequired = false
			noCommitment.NonGoals[i].NonGoalHash = ComputeNonGoalHash(noCommitment.NonGoals[i])
			break
		}
	}
	noCommitment.ManifestHash = ComputeNonGoalManifestHash(noCommitment)
	require.ErrorContains(t, noCommitment.Validate(), "committed proof or receipt")

	hashMismatch := manifest
	hashMismatch.NonGoals = append([]NonGoal(nil), manifest.NonGoals...)
	hashMismatch.NonGoals[0].Rationale = append([]string(nil), hashMismatch.NonGoals[0].Rationale...)
	hashMismatch.NonGoals[0].Rationale[0] = "tampered rationale"
	hashMismatch.ManifestHash = ComputeNonGoalManifestHash(hashMismatch)
	require.ErrorContains(t, hashMismatch.Validate(), "non-goal hash mismatch")
}

func TestNonGoalManifestHashIsCanonical(t *testing.T) {
	manifest, err := DefaultNonGoalManifest()
	require.NoError(t, err)

	reversed := reverseNonGoals(manifest.NonGoals)
	for i := range reversed {
		reversed[i].Modules = reverseModules(reversed[i].Modules)
		reversed[i].Rationale = reverseStrings(reversed[i].Rationale)
		reversed[i].NonGoalHash = ""
	}
	reordered, err := NewNonGoalManifest(reversed)
	require.NoError(t, err)
	require.Equal(t, manifest.ManifestHash, reordered.ManifestHash)
	require.Equal(t, manifest.NonGoals, reordered.NonGoals)

	tampered := manifest
	tampered.NonGoals = append([]NonGoal(nil), manifest.NonGoals...)
	tampered.NonGoals[0].Boundary = "tampered boundary"
	tampered.NonGoals[0].NonGoalHash = ComputeNonGoalHash(tampered.NonGoals[0])
	tampered.ManifestHash = ComputeNonGoalManifestHash(tampered)
	require.NoError(t, tampered.Validate())
	require.NotEqual(t, manifest.ManifestHash, tampered.ManifestHash)
}

func reverseNonGoals(nonGoals []NonGoal) []NonGoal {
	out := make([]NonGoal, len(nonGoals))
	for i := range nonGoals {
		out[i] = nonGoals[len(nonGoals)-1-i]
	}
	return out
}
