package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubdomainCreationV2DetachedLifecycleAndExpiryRules(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	parent, found := findDomain(state, "alice.aet")
	require.True(t, found)

	_, _, err := IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:		"alice.aet",
		Label:			"api",
		Actor:			addr(1),
		ChildOwner:		addr(2),
		Height:			20,
		ChildExpiryHeight:	parent.ExpiryHeight + 10,
		DelegationType:		SubdomainDelegationOwnerControlledV2,
	})
	require.ErrorContains(t, err, "cannot exceed parent expiry")

	next, record, err := IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:		"alice.aet",
		Label:			"api",
		Actor:			addr(1),
		ChildOwner:		addr(2),
		Height:			20,
		ChildExpiryHeight:	parent.ExpiryHeight + 10,
		DelegationType:		SubdomainDelegationDetachedPaidV2,
		DetachedPaid:		true,
		IndependentPayment:	true,
		ParentAuthorization:	true,
	})
	require.NoError(t, err)
	require.True(t, record.Detached)
	require.True(t, record.ParentAuthorized)
	require.Equal(t, SubdomainDelegationDetachedPaidV2, record.DelegationType)
	child, found := findDomain(next, "api.alice.aet")
	require.True(t, found)
	require.Equal(t, parent.ExpiryHeight+10, child.ExpiryHeight)
}

func TestSubdomainCreationV2DelegateZoneAndEphemeralValidation(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	delegate, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeSubdomainCreate, []string{"create"}, 100, 2, "", 10)
	require.NoError(t, err)
	_, record, err := IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:	"alice.aet",
		Label:		"svc",
		Actor:		addr(7),
		ChildOwner:	addr(8),
		Height:		20,
		DelegationType:	SubdomainDelegationDelegateControlledV2,
		Delegation:	&delegate,
	})
	require.NoError(t, err)
	require.Equal(t, SubdomainDelegationDelegateControlledV2, record.DelegationType)

	zone, err := NewDelegationRecordV2("alice.aet", addr(9), DelegationScopeZoneAdmin, []string{"create", "resolve"}, 100, 2, "", 10)
	require.NoError(t, err)
	_, record, err = IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:	"alice.aet",
		Label:		"zone",
		Actor:		addr(9),
		ChildOwner:	addr(9),
		Height:		20,
		DelegationType:	SubdomainDelegationZoneManagedV2,
		Delegation:	&zone,
	})
	require.NoError(t, err)
	require.Equal(t, SubdomainDelegationZoneManagedV2, record.DelegationType)

	_, _, err = IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:	"alice.aet",
		Label:		"tmp",
		Actor:		addr(1),
		ChildOwner:	addr(1),
		Height:		20,
		DelegationType:	SubdomainDelegationEphemeralServiceV2,
	})
	require.ErrorContains(t, err, "ephemeral")
}

func TestDelegationScopeBitsAndTimeLockedRevocationV2(t *testing.T) {
	record, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeSubdomainCreate, []string{"create"}, 100, 2, "", 10)
	require.NoError(t, err)
	require.Equal(t, DelegationScopeBitSubdomainCreateV2, record.ScopeBits)

	record.ScopeBits = DelegationScopeBitResolverUpdateV2
	require.ErrorContains(t, ValidateDelegationRecordV2(record), "scope_bits")

	record.ScopeBits = DelegationScopeBitSubdomainCreateV2
	record.TimeLockedUntilHeight = 50
	require.NoError(t, ValidateDelegationRecordV2(record))
	_, revoked, err := RevokeDelegationV2([]DelegationRecordV2{record}, "alice.aet", addr(7), DelegationScopeSubdomainCreate, addr(1), addr(1), 40)
	require.ErrorContains(t, err, "time-locked")
	require.False(t, revoked)

	remaining, revoked, err := RevokeDelegationV2([]DelegationRecordV2{record}, "alice.aet", addr(7), DelegationScopeSubdomainCreate, addr(1), addr(1), 50)
	require.NoError(t, err)
	require.True(t, revoked)
	require.Empty(t, remaining)
}

func TestIdentityPathCommitmentAndOptimizedRecursiveProofV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := IssueSubdomain(state, "alice.aet", "api", addr(1), addr(2), false, 12)
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "api.alice.aet", addr(2), ResolverPatch{Primary: addr(3)}, 13)
	require.NoError(t, err)
	path, err := CanonicalResolutionPathV2("api.alice.aet")
	require.NoError(t, err)
	commitment, err := BuildIdentityPathCommitmentV2(path, 7, 2, 3)
	require.NoError(t, err)
	require.NoError(t, ValidateIdentityPathCommitmentV2(commitment))
	require.Equal(t, identityHash("identity-v2-resolution-path", "00000000000000000002", "alice.aet", "api.alice.aet"), commitment.PathHash)

	record, err := BuildUnifiedResolutionRecordV2(state, "api.alice.aet", 14, 30)
	require.NoError(t, err)
	recordHash, err := ComputeResolvedRecordHashV2(record)
	require.NoError(t, err)
	cache, err := NewResolutionCacheRecordV2("api.alice.aet", commitment.PathHash, recordHash, 40, 7, 2, 3)
	require.NoError(t, err)
	proof, optimizedCommitment, err := BuildOptimizedRecursiveResolutionProofV2(OptimizedRecursiveResolutionProofRequestV2{
		State:		state,
		ChainID:	"aetra-local-1",
		RootName:	"alice.aet",
		TargetName:	"api.alice.aet",
		Height:		14,
		TTL:		30,
		Cache:		&cache,
		SourceVersion:	7,
		ParentEpoch:	2,
		ChildEpoch:	3,
	})
	require.NoError(t, err)
	require.Equal(t, commitment.CommitmentHash, optimizedCommitment.CommitmentHash)
	require.NotNil(t, proof.CacheRecordOptional)

	query := NewIdentityQueryServiceV2(IdentityQueryContextV2{State: state, Height: 14, DefaultTTL: 30})
	resp := query.QueryOptimizedRecursiveResolutionProof("alice.aet", "api.alice.aet", &cache, 7, 2, 3, false, false)
	require.Equal(t, IdentityQueryOK, resp.Code)
	require.NotNil(t, resp.RecursiveProof)
	require.NotNil(t, resp.PathCommitment)

	staleParent := InvalidateResolutionCacheRecordV2ForParentEpochChange(cache, 4)
	require.ErrorContains(t, ValidateResolutionCacheRecordV2(staleParent), "valid_until_height")
	_, _, err = BuildOptimizedRecursiveResolutionProofV2(OptimizedRecursiveResolutionProofRequestV2{
		State:		state,
		ChainID:	"aetra-local-1",
		RootName:	"alice.aet",
		TargetName:	"api.alice.aet",
		Height:		14,
		TTL:		30,
		Cache:		&cache,
		SourceVersion:	7,
		ParentEpoch:	4,
		ChildEpoch:	3,
	})
	require.ErrorContains(t, err, "parent epoch")
}
