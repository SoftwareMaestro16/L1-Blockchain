package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCosmosModuleBoundarySpecCoversSectionsTenTwoAndTenThree(t *testing.T) {
	require.NoError(t, ValidateCosmosModuleBoundarySpec())

	spec, err := DefaultCosmosModuleBoundarySpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.Modifications, 7)
	require.Len(t, spec.BoundaryRules, 7)
	require.NotEmpty(t, spec.Root)

	byModule := map[string]ExistingCosmosModuleModification{}
	for _, modification := range spec.Modifications {
		require.NoError(t, modification.Validate())
		byModule[modification.Module] = modification
	}

	require.Contains(t, byModule["bank"].RequiredModification, "zone-aware account routing")
	require.Contains(t, byModule["bank"].RequiredModification, "cross-shard escrow")
	require.Contains(t, byModule["bank"].Boundary, "Financial Zone prefixes")
	require.Contains(t, byModule["staking"].RequiredModification, "validator set commitment")
	require.Contains(t, byModule["staking"].Boundary, "globally committed")
	require.Contains(t, byModule["slashing"].Boundary, "payment dispute penalties")
	require.Contains(t, byModule["mint/distribution"].Boundary, "committed fee roots")
	require.Contains(t, byModule["fees"].RequiredModification, "forwarding fee escrow")
	require.Contains(t, byModule["contract-assets"].Boundary, "Financial Zone state")
	require.Contains(t, byModule["avm-dex-contract"].RequiredModification, "pool shard placement")
	require.Contains(t, byModule["avm-dex-contract"].Boundary, "settle through receipts")

	byRule := map[string]CosmosModuleBoundaryRule{}
	for _, rule := range spec.BoundaryRules {
		require.NoError(t, rule.Validate())
		byRule[rule.Rule] = rule
	}
	require.Contains(t, byRule["Core modules commit roots and schedule work."].Enforcement, "cannot own zone-local")
	require.Contains(t, byRule["Zone modules own local state transitions."].Enforcement, "zone and shard prefixes")
	require.Contains(t, byRule["Message module connects zones and shards."].Enforcement, "committed messages and receipts")
	require.Contains(t, byRule["Proof module verifies committed state only."].Enforcement, "cannot read live mempool")
	require.Contains(t, byRule["VM module cannot mutate state outside Contract Zone except by message."].Enforcement, "AVM syscalls emit messages")
	require.Contains(t, byRule["Identity module cannot transfer funds except through Financial Zone messages."].Enforcement, "cannot directly debit or credit")
	require.Contains(t, byRule["Payments module cannot resolve names except through Identity Zone proof or message."].Enforcement, "verified .aet resolution")
}

func TestCosmosModuleBoundarySpecRootIsCanonicalAcrossInputOrder(t *testing.T) {
	spec, err := DefaultCosmosModuleBoundarySpec()
	require.NoError(t, err)

	modifications := append([]ExistingCosmosModuleModification(nil), ExistingCosmosModuleModifications()...)
	rules := append([]CosmosModuleBoundaryRule(nil), CosmosModuleBoundaryRules()...)
	slices.Reverse(modifications)
	slices.Reverse(rules)

	reordered, err := BuildCosmosModuleBoundarySpec(modifications, rules)
	require.NoError(t, err)
	require.Equal(t, spec.Root, reordered.Root)
	require.Equal(t, spec.Modifications, reordered.Modifications)
	require.Equal(t, spec.BoundaryRules, reordered.BoundaryRules)
}

func TestCosmosModuleBoundarySpecRejectsMalformedEntries(t *testing.T) {
	duplicate, err := BuildCosmosModuleBoundarySpec(
		[]ExistingCosmosModuleModification{ExistingCosmosModuleModifications()[0], ExistingCosmosModuleModifications()[0]},
		CosmosModuleBoundaryRules(),
	)
	require.ErrorContains(t, err, "duplicate existing cosmos module")
	require.Empty(t, duplicate.Root)

	noBoundary := ExistingCosmosModuleModification{
		Module:			"bank",
		RequiredModification:	"zone-aware account routing",
	}
	_, err = BuildExistingCosmosModuleModification(noBoundary)
	require.ErrorContains(t, err, "boundary is required")

	noEnforcement := CosmosModuleBoundaryRule{
		Rule: "Core modules commit roots and schedule work.",
	}
	_, err = BuildCosmosModuleBoundaryRule(noEnforcement)
	require.ErrorContains(t, err, "enforcement is required")

	tampered := ExistingCosmosModuleModifications()[0]
	tampered.Boundary = strings.ReplaceAll(tampered.Boundary, "Financial", "Application")
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")

	tamperedRule := CosmosModuleBoundaryRules()[0]
	tamperedRule.Enforcement = strings.ReplaceAll(tamperedRule.Enforcement, "cannot", "can")
	require.ErrorContains(t, tamperedRule.Validate(), "rule hash mismatch")
}
