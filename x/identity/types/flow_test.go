package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityDataFlowPlansResolverUpdatePipeline(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	flow, err := PlanIdentityDataFlow(state, IdentityDataFlowRequest{
		Operation:	IdentityOperationResolverUpdate,
		Domain:		"api.alice.aet",
		Actor:		addr(1),
		ResolverPatch:	&ResolverPatch{Primary: addr(2)},
	}, 12)
	require.NoError(t, err)
	require.Equal(t, "api.alice.aet", flow.Domain)
	require.Equal(t, []IdentityFlowStep{
		IdentityFlowStepDeterministicValidation,
		IdentityFlowStepRegistryNFTCheck,
		IdentityFlowStepResolverStateUpdate,
		IdentityFlowStepStoreV2Writes,
		IdentityFlowStepEventsProofsRouting,
	}, flow.Steps)
	require.Contains(t, flow.EventTypes, "identity_resolver_updated")
	require.Equal(t, []IdentityV2Component{IdentityV2RoutingIntegration}, flow.RoutingHooks)
	require.Len(t, flow.ProofRoot, 64)

	resolverKey, err := IdentityResolverStoreKey("api.alice.aet")
	require.NoError(t, err)
	require.Equal(t, []string{resolverKey}, flow.StoreWrites)
	require.Contains(t, flow.AccessSet.Writes, resolverKey)
}

func TestIdentityDataFlowPlansTransferAndRejectsNonOwner(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	flow, err := PlanIdentityDataFlow(state, IdentityDataFlowRequest{
		Operation:	IdentityOperationTransfer,
		Domain:		"alice.aet",
		Actor:		addr(1),
		NewOwner:	addr(9),
	}, 12)
	require.NoError(t, err)
	require.Contains(t, flow.EventTypes, "identity_domain_transferred")
	require.GreaterOrEqual(t, len(flow.StoreWrites), 4)

	_, err = PlanIdentityDataFlow(state, IdentityDataFlowRequest{
		Operation:	IdentityOperationTransfer,
		Domain:		"alice.aet",
		Actor:		addr(2),
		NewOwner:	addr(9),
	}, 12)
	require.ErrorContains(t, err, "requires owner")
}

func TestIdentityTrustBoundariesValidateRegistryNFTAndResolverControl(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	domain, nft, err := ValidateRegistryNFTAuthority(state, "alice.aet", 12)
	require.NoError(t, err)
	require.Equal(t, domain.NFTID, nft.ID)

	authority, err := ValidateResolverControlBoundary(state, "api.alice.aet", addr(1), 12)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", authority.Name)

	_, err = ValidateResolverControlBoundary(state, "api.alice.aet", addr(2), 12)
	require.ErrorContains(t, err, "requires owner")

	state.DomainNFTs[0].Owner = addr(9)
	_, _, err = ValidateRegistryNFTAuthority(state, "alice.aet", 12)
	require.ErrorContains(t, err, "ownership mismatch")
}

func TestIdentityTrustBoundariesProofAndCacheFreshness(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof, err := BuildIdentityResolutionProof(state, "alice.aet", 13)
	require.NoError(t, err)
	root, err := IdentityStateRoot(state)
	require.NoError(t, err)

	resolution, err := ValidateResolutionProofBoundary(proof, root, 13)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", resolution.ResolverDomain)

	_, err = ValidateResolutionProofBoundary(proof, identityHash("wrong-root"), 13)
	require.ErrorContains(t, err, "trusted root")

	require.NoError(t, ValidateResolutionCacheFreshness(10, 13, 3))
	require.ErrorContains(t, ValidateResolutionCacheFreshness(10, 14, 3), "stale")

	report := EvaluateIdentityTrustBoundaries(state, "alice.aet", addr(1), 13, &proof, 12, 2)
	require.Equal(t, IdentityBoundaryTrusted, report.RegistryNFTAuthority)
	require.Equal(t, IdentityBoundaryTrusted, report.ResolverAuthority)
	require.Equal(t, IdentityBoundaryTrusted, report.LightClientProof)
	require.Equal(t, IdentityBoundaryTrusted, report.CacheFreshness)
	require.Equal(t, IdentityBoundaryAdvisory, report.ServiceMetadata)
	require.Equal(t, IdentityBoundaryAdvisory, report.InterfaceDescriptor)
}
