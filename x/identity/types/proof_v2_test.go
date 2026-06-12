package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentityLightClientProofV2VerifiesAllResolutionObjectives(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := IssueSubdomain(state, "alice.aet", "api", addr(1), addr(1), true, 11)
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "api.alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(3),
		},
	}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "api.alice.aet", 13)
	require.NoError(t, err)

	proof, err := BuildIdentityLightClientResolutionProofV2(state, IdentityLightClientProofRequestV2{
		Name:		"svc.api.alice.aet",
		TrustedHeight:	14,
		RecordTTL:	10,
		ReverseAddress:	addr(2),
		Objectives: []IdentityProofObjectiveV2{
			IdentityProofObjectiveDomainNonExistence,
			IdentityProofObjectiveDomainStatusExpiry,
			IdentityProofObjectiveNFTBinding,
			IdentityProofObjectiveResolverRecord,
			IdentityProofObjectiveReverseConsistency,
			IdentityProofObjectiveSubdomainDelegation,
			IdentityProofObjectiveRecursiveResolution,
			IdentityProofObjectiveVersionAndFreshness,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "svc.api.alice.aet", proof.QueryDomain)
	require.NotNil(t, proof.QueryDomainAbsence)
	require.NotNil(t, proof.ResolutionProof)
	require.NotNil(t, proof.ReverseProof)
	require.Len(t, proof.SubdomainChain, 1)
	require.Equal(t, uint64(12), proof.RecordVersion)

	root, err := IdentityStateRoot(state)
	require.NoError(t, err)
	require.NoError(t, VerifyIdentityLightClientResolutionProofV2(proof, root, 14, 20))

	proof.RecordVersion++
	require.ErrorContains(t, VerifyIdentityLightClientResolutionProofV2(proof, root, 14, 20), "proof hash mismatch")
}

func TestIdentityLightClientProofV2DomainExistenceAndFreshness(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)

	proof, err := BuildIdentityLightClientResolutionProofV2(state, IdentityLightClientProofRequestV2{
		Name:		"alice.aet",
		TrustedHeight:	13,
		RecordTTL:	5,
	})
	require.NoError(t, err)
	require.NotNil(t, proof.QueryDomainProof)
	require.Equal(t, DomainLifecycleActive, proof.DomainStatus)
	require.Equal(t, uint64(12), proof.RecordVersion)

	root, err := IdentityStateRoot(state)
	require.NoError(t, err)
	require.NoError(t, VerifyIdentityLightClientResolutionProofV2(proof, root, 13, 18))
	require.ErrorContains(t, VerifyIdentityLightClientResolutionProofV2(proof, root, 13, 19), "stale")
	require.ErrorContains(t, VerifyIdentityLightClientResolutionProofV2(proof, identityHash("wrong-root"), 13, 18), "root mismatch")
}

func TestIdentityLightClientProofV2NonExistenceOnly(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)

	proof, err := BuildIdentityLightClientResolutionProofV2(state, IdentityLightClientProofRequestV2{
		Name:		"missing.aet",
		TrustedHeight:	13,
		RecordTTL:	5,
		Objectives: []IdentityProofObjectiveV2{
			IdentityProofObjectiveDomainNonExistence,
			IdentityProofObjectiveVersionAndFreshness,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, proof.QueryDomainAbsence)
	require.Nil(t, proof.ResolutionProof)
	require.Equal(t, DomainLifecycleAvailable, proof.DomainStatus)

	root, err := IdentityStateRoot(state)
	require.NoError(t, err)
	require.NoError(t, VerifyIdentityLightClientResolutionProofV2(proof, root, 13, 13))
}

func TestIdentityLightClientProofV2RejectsReverseMismatch(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)

	proof, err := BuildIdentityLightClientResolutionProofV2(state, IdentityLightClientProofRequestV2{
		Name:		"alice.aet",
		TrustedHeight:	14,
		RecordTTL:	10,
		ReverseAddress:	addr(2),
		Objectives: []IdentityProofObjectiveV2{
			IdentityProofObjectiveReverseConsistency,
			IdentityProofObjectiveRecursiveResolution,
			IdentityProofObjectiveVersionAndFreshness,
		},
	})
	require.NoError(t, err)
	root, err := IdentityStateRoot(state)
	require.NoError(t, err)
	require.NoError(t, VerifyIdentityLightClientResolutionProofV2(proof, root, 14, 14))

	proof.ReverseRecord.Address = addr(9)
	proof.ProofHash = ComputeIdentityLightClientResolutionProofHashV2(proof)
	require.ErrorContains(t, VerifyIdentityLightClientResolutionProofV2(proof, root, 14, 14), "reverse proof value mismatch")
}
