package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityZoneUpgradeManifestCoversRequiredCapabilities(t *testing.T) {
	artifacts := testIdentityZoneUpgradeArtifacts(t)
	manifest, roots, err := BuildDefaultIdentityZoneUpgradeManifest(artifacts)
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Equal(t, IdentityZoneID, manifest.ZoneID)
	require.Equal(t, "IDENTITY", manifest.ZoneType)
	require.Equal(t, IdentityStoreV2Prefix, manifest.StorePrefix)
	require.Equal(t, roots.StateRoot, manifest.StateRoot)
	require.Equal(t, ComputeIdentityZoneCapabilityRoot(DefaultIdentityZoneCapabilities()), manifest.CapabilityRoot)
	require.Len(t, manifest.RequiredCapabilities, len(DefaultIdentityZoneCapabilities()))
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityCommitRevealRegistration)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityNFTOwnershipBinding)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityResolverRecords)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityReverseLookup)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityAuctions)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityDeterministicResolverVMHook)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityMultiRecordResolutionGraph)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityCrossZoneIdentityBinding)
	require.Contains(t, manifest.RequiredCapabilities, IdentityCapabilityLightClientProofVerification)
}

func TestIdentityZoneUpgradeManifestIsDeterministic(t *testing.T) {
	artifacts := testIdentityZoneUpgradeArtifacts(t)
	first, _, err := BuildDefaultIdentityZoneUpgradeManifest(artifacts)
	require.NoError(t, err)

	artifacts.Hooks = append([]IdentityResolverVMHook(nil), artifacts.Hooks...)
	artifacts.Graphs = append([]IdentityResolutionGraph(nil), artifacts.Graphs...)
	artifacts.Bindings = append([]IdentityCrossZoneBinding(nil), artifacts.Bindings...)
	artifacts.Proofs = append([]IdentityZoneProofIndexEntry{artifacts.Proofs[1], artifacts.Proofs[0]}, artifacts.Proofs[2:]...)
	second, _, err := BuildDefaultIdentityZoneUpgradeManifest(artifacts)
	require.NoError(t, err)
	require.Equal(t, first.UpgradeHash, second.UpgradeHash)
	require.Equal(t, first.CapabilityRoot, second.CapabilityRoot)
}

func TestIdentityZoneUpgradeRejectsMissingCapabilityAndBadAEKDescriptor(t *testing.T) {
	artifacts := testIdentityZoneUpgradeArtifacts(t)
	roots, err := BuildIdentityZoneRoots(artifacts.Height, artifacts.State, artifacts.Hooks, artifacts.Graphs, artifacts.Bindings, artifacts.Proofs)
	require.NoError(t, err)

	capabilities := DefaultIdentityZoneCapabilities()
	_, err = BuildIdentityZoneUpgradeManifest(DefaultAetraCoreIdentityZoneDescriptor(), DefaultIdentityZoneStateMachineDescriptor(), DefaultIdentityV2Architecture(), roots, capabilities[:len(capabilities)-1])
	require.ErrorContains(t, err, "complete capability set")

	descriptor := DefaultAetraCoreIdentityZoneDescriptor()
	descriptor.Enabled = false
	_, err = BuildIdentityZoneUpgradeManifest(descriptor, DefaultIdentityZoneStateMachineDescriptor(), DefaultIdentityV2Architecture(), roots, capabilities)
	require.ErrorContains(t, err, "must be enabled")
}

func TestIdentityZoneUpgradeArtifactsRequireRuntimeSubsystems(t *testing.T) {
	artifacts := testIdentityZoneUpgradeArtifacts(t)
	artifacts.Hooks = nil
	_, _, err := BuildDefaultIdentityZoneUpgradeManifest(artifacts)
	require.ErrorContains(t, err, "resolver VM hooks")

	artifacts = testIdentityZoneUpgradeArtifacts(t)
	artifacts.Graphs = nil
	_, _, err = BuildDefaultIdentityZoneUpgradeManifest(artifacts)
	require.ErrorContains(t, err, "resolution graph")

	artifacts = testIdentityZoneUpgradeArtifacts(t)
	artifacts.Bindings = nil
	_, _, err = BuildDefaultIdentityZoneUpgradeManifest(artifacts)
	require.ErrorContains(t, err, "cross-zone")

	artifacts = testIdentityZoneUpgradeArtifacts(t)
	artifacts.Proofs = nil
	_, _, err = BuildDefaultIdentityZoneUpgradeManifest(artifacts)
	require.ErrorContains(t, err, "light-client proof")
}

func TestIdentityZoneAEKDescriptorRequiresIdentityZone(t *testing.T) {
	descriptor := DefaultAetraCoreIdentityZoneDescriptor()
	require.NoError(t, ValidateIdentityZoneAEKDescriptor(descriptor))

	descriptor.RootPrefix = "identity-v1"
	err := ValidateIdentityZoneAEKDescriptor(descriptor)
	require.ErrorContains(t, err, "root prefix")
}

func testIdentityZoneUpgradeArtifacts(t *testing.T) IdentityZoneUpgradeArtifacts {
	t.Helper()
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2), ZoneEndpoint: "identity/alice"}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 12)
	require.NoError(t, err)
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	stateRoot, err := IdentityStateRoot(state)
	require.NoError(t, err)

	hook, err := NewIdentityResolverVMHook(IdentityResolverVMHook{
		HookID:		"primary-hook",
		NameHash:	nameHash,
		RecordKey:	ResolverKeyPrimary,
		InputHash:	identityHash("resolver-input"),
		OutputHash:	identityHash("resolver-output"),
		GasLimit:	10000,
		Version:	1,
	})
	require.NoError(t, err)

	graph, err := NewIdentityResolutionGraph(IdentityResolutionGraph{
		Height:	13,
		Nodes: []IdentityResolutionGraphNode{
			{NodeID: "domain", NameHash: nameHash, TargetHash: stateRoot},
			{NodeID: "resolver", NameHash: nameHash, RecordKey: ResolverKeyPrimary, TargetHash: identityHash("resolver-target")},
			{NodeID: "reverse", NameHash: nameHash, TargetHash: identityHash("reverse-target")},
		},
		Edges: []IdentityResolutionGraphEdge{
			{FromNodeID: "domain", ToNodeID: "resolver", EdgeKind: "resolves"},
			{FromNodeID: "resolver", ToNodeID: "reverse", EdgeKind: "reverse"},
		},
	})
	require.NoError(t, err)

	binding, err := NewIdentityCrossZoneBinding(IdentityCrossZoneBinding{
		NameHash:	nameHash,
		ZoneID:		"APPLICATION_ZONE",
		BindingKey:	"apps/app/alice",
		BindingRoot:	identityHash("application-binding-root"),
		Height:		13,
	})
	require.NoError(t, err)

	domainProof, err := QueryIdentityZoneLightClientProof(state, IdentityProofDomain, "alice.aet", 13)
	require.NoError(t, err)
	resolverProof, err := QueryIdentityZoneLightClientProof(state, IdentityProofResolver, "alice.aet", 13)
	require.NoError(t, err)
	reverseProof, _, err := QueryIdentityZoneReverseLookupProof(state, addr(2), 13)
	require.NoError(t, err)

	return IdentityZoneUpgradeArtifacts{
		State:		state,
		Hooks:		[]IdentityResolverVMHook{hook},
		Graphs:		[]IdentityResolutionGraph{graph},
		Bindings:	[]IdentityCrossZoneBinding{binding},
		Proofs:		[]IdentityZoneProofIndexEntry{domainProof, resolverProof, reverseProof},
		Height:		13,
	}
}
