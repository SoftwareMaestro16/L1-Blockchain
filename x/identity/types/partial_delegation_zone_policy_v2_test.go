package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPartialDelegationV2ScopesVersionAndPrefixBoundLabels(t *testing.T) {
	prefix, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeSubdomainCreate, []string{"prefix.svc"}, 100, 2, "svc", 10)
	require.NoError(t, err)
	require.Equal(t, DelegationRecordVersionV2, prefix.DelegationVersion)
	require.NoError(t, ValidatePartialDelegationAuthorizationV2(prefix, PartialDelegationAuthorizationV2{
		Scope:				DelegationScopeSubdomainCreate,
		ChildLabel:			"svcapi",
		SubtreeDepth:			1,
		Height:				20,
		ExpectedDelegationVersion:	DelegationRecordVersionV2,
	}))
	require.ErrorContains(t, ValidatePartialDelegationAuthorizationV2(prefix, PartialDelegationAuthorizationV2{
		Scope:				DelegationScopeSubdomainCreate,
		ChildLabel:			"api",
		SubtreeDepth:			1,
		Height:				20,
		ExpectedDelegationVersion:	DelegationRecordVersionV2,
	}), "child label prefix")
	require.ErrorContains(t, ValidatePartialDelegationAuthorizationV2(prefix, PartialDelegationAuthorizationV2{
		Scope:				DelegationScopeSubdomainCreate,
		ChildLabel:			"svcapi",
		SubtreeDepth:			1,
		Height:				20,
		ExpectedDelegationVersion:	2,
	}), "version conflict")

	specific, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeSubdomainCreate, []string{"label.api"}, 100, 2, "", 10)
	require.NoError(t, err)
	require.NoError(t, ValidatePartialDelegationAuthorizationV2(specific, PartialDelegationAuthorizationV2{
		Scope:				DelegationScopeSubdomainCreate,
		ChildLabel:			"api",
		SubtreeDepth:			1,
		Height:				20,
		ExpectedDelegationVersion:	DelegationRecordVersionV2,
	}))
	require.ErrorContains(t, ValidatePartialDelegationAuthorizationV2(specific, PartialDelegationAuthorizationV2{
		Scope:				DelegationScopeSubdomainCreate,
		ChildLabel:			"www",
		SubtreeDepth:			1,
		Height:				20,
		ExpectedDelegationVersion:	DelegationRecordVersionV2,
	}), "does not allow child label")

	service, err := NewDelegationRecordV2("alice.aet", addr(8), DelegationScopeServiceRecordUpdate, []string{"service.rpc"}, 100, 1, "service.", 10)
	require.NoError(t, err)
	require.NoError(t, ValidatePartialDelegationAuthorizationV2(service, PartialDelegationAuthorizationV2{
		Scope:				DelegationScopeServiceRecordUpdate,
		RecordKey:			"service.rpc",
		Height:				20,
		ExpectedDelegationVersion:	DelegationRecordVersionV2,
	}))
	require.ErrorContains(t, ValidatePartialDelegationAuthorizationV2(service, PartialDelegationAuthorizationV2{
		Scope:				DelegationScopeRoutingRecordUpdate,
		RecordKey:			"service.rpc",
		Height:				20,
		ExpectedDelegationVersion:	DelegationRecordVersionV2,
	}), "scope mismatch")
}

func TestPartialDelegationV2RejectsEscalationAndParentTransferWithoutGrant(t *testing.T) {
	parent, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeServiceRecordUpdate, []string{"service.rpc"}, 100, 1, "service.", 10)
	require.NoError(t, err)
	child := parent
	child.ScopeBits |= DelegationScopeBitRoutingRecordUpdateV2
	require.ErrorContains(t, ValidateDelegationDoesNotEscalateV2(parent, child), "scope bits")

	child = parent
	child.Permissions = []string{"service.admin", "service.rpc"}
	require.ErrorContains(t, ValidateDelegationDoesNotEscalateV2(parent, child), "permission")

	child = parent
	child.ExpiresAtHeight = parent.ExpiresAtHeight + 1
	require.ErrorContains(t, ValidateDelegationDoesNotEscalateV2(parent, child), "expiry")

	transfer, err := NewDelegationRecordV2("alice.aet", addr(9), DelegationScopeSubdomainTransfer, []string{DelegationPermissionTransferParentV2}, 100, 0, "", 10)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateParentTransferByDelegationV2(transfer, DelegationRecordVersionV2, 20), "cannot transfer parent")
	transfer.CanTransferParent = true
	require.NoError(t, ValidateParentTransferByDelegationV2(transfer, DelegationRecordVersionV2, 20))
	require.ErrorContains(t, ValidateParentTransferByDelegationV2(transfer, 2, 20), "version conflict")
}

func TestZonePolicyV2InheritanceLimitsCacheInvalidationAndRecursiveProof(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, subdomain, err := IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:	"alice.aet",
		Label:		"api",
		Actor:		addr(1),
		ChildOwner:	addr(2),
		Height:		20,
		DelegationType:	SubdomainDelegationOwnerControlledV2,
	})
	require.NoError(t, err)
	require.NotEmpty(t, state.Domains)

	parentPolicy, err := NewZonePolicyV2("alice.aet", ZonePolicyV2{
		AllowedRecordTypes:		[]string{"primary", "service"},
		AllowedServiceTypes:		[]string{"rpc.v1"},
		SubdomainCreationPolicy:	ZoneSubdomainCreationDelegatedV2,
		ResolverUpdatePolicy:		ZoneResolverUpdateDelegatedV2,
		InterfacePolicy:		ZoneInterfacePolicyHashRequiredV2,
		RoutingPolicy:			ZoneRoutingPolicyExplicitTargetsV2,
		MaxChildDepth:			2,
		MaxChildRecords:		5,
		UpdatedAtHeight:		20,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateZonePolicyForSubdomainV2(parentPolicy, subdomain, 1, 2, "primary", "rpc.v1"))
	require.ErrorContains(t, ValidateZonePolicyForSubdomainV2(parentPolicy, subdomain, 3, 2, "primary", "rpc.v1"), "max_child_depth")
	require.ErrorContains(t, ValidateZonePolicyForSubdomainV2(parentPolicy, subdomain, 1, 6, "primary", "rpc.v1"), "max_child_records")
	require.ErrorContains(t, ValidateZonePolicyForSubdomainV2(parentPolicy, subdomain, 1, 2, "primary", "grpc.v1"), "disallows service")

	childPolicy, err := NewZonePolicyV2("api.alice.aet", ZonePolicyV2{
		AllowedRecordTypes:		[]string{"primary"},
		AllowedServiceTypes:		[]string{"graphql.v1"},
		SubdomainCreationPolicy:	ZoneSubdomainCreationOwnerOnlyV2,
		ResolverUpdatePolicy:		ZoneResolverUpdateOwnerOnlyV2,
		InterfacePolicy:		ZoneInterfacePolicyWalletPolicyV2,
		RoutingPolicy:			ZoneRoutingPolicyWalletPolicyV2,
		MaxChildDepth:			1,
		MaxChildRecords:		2,
		UpdatedAtHeight:		21,
		ParentPolicyHash:		parentPolicy.PolicyHash,
		OverrideParent:			true,
	})
	require.NoError(t, err)
	subdomain.Detached = true
	resolved, err := ResolveZonePolicyForChildV2(parentPolicy, &childPolicy, subdomain)
	require.NoError(t, err)
	require.Equal(t, childPolicy.PolicyHash, resolved.PolicyHash)

	path, err := CanonicalResolutionPathV2("api.alice.aet")
	require.NoError(t, err)
	commitment, err := BuildIdentityPathCommitmentV2(path, 7, parentPolicy.LifecycleEpoch, childPolicy.LifecycleEpoch)
	require.NoError(t, err)
	cache, err := NewResolutionCacheRecordV2("api.alice.aet", commitment.PathHash, identityHash("record"), 50, 7, parentPolicy.LifecycleEpoch, childPolicy.LifecycleEpoch)
	require.NoError(t, err)
	updatedPolicy, invalidated, err := ApplyZonePolicyChangeV2(parentPolicy, 30, []ResolutionCacheRecordV2{cache})
	require.NoError(t, err)
	require.Equal(t, parentPolicy.LifecycleEpoch+1, updatedPolicy.LifecycleEpoch)
	require.Len(t, invalidated, 1)
	require.ErrorContains(t, ValidateResolutionCacheRecordV2(invalidated[0]), "valid_until_height")

	proof, err := BuildRecursivePolicyProofV2("alice.aet", "api.alice.aet", commitment, []ZonePolicyV2{parentPolicy, childPolicy})
	require.NoError(t, err)
	require.NoError(t, ValidateRecursivePolicyProofV2(proof))
	proof.ZonePolicies[0].AllowedRecordTypes = append(proof.ZonePolicies[0].AllowedRecordTypes, "wallet")
	require.ErrorContains(t, ValidateRecursivePolicyProofV2(proof), "hash mismatch")
}
