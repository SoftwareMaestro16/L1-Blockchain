package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverModuleBreakdownV2CoversSection132(t *testing.T) {
	breakdown, err := DefaultResolverModuleBreakdownV2()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.NotEmpty(t, breakdown.BreakdownHash)
	require.ElementsMatch(t, requiredResolverModuleStateObjectsV2(), breakdown.StateObjects)
	require.ElementsMatch(t, requiredResolverModuleMessagesV2(), breakdown.Messages)
	require.ElementsMatch(t, requiredResolverModuleQueriesV2(), breakdown.Queries)
	require.ElementsMatch(t, requiredResolverModuleIntegrationPointsV2(), breakdown.IntegrationPoints)
	require.Contains(t, breakdown.Messages, ResolverModuleMsgClearResolverRecord)
	require.Contains(t, breakdown.BackingPrimitives, "ValidateUnifiedResolutionRecordV2")
	require.Contains(t, breakdown.BackingPrimitives, "ExecuteBatchResolverUpdatesV2")
	require.Contains(t, breakdown.BackingPrimitives, "BuildIdentityResolutionProofFormatV2")
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecResolversPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecReversePrefix)
}

func TestResolverModuleValidationV2RecordTTLReverseBatchAndProof(t *testing.T) {
	state := resolverModuleState(t)
	record, err := BuildUnifiedResolutionRecordV2(state, "alice.aet", 14, 30)
	require.NoError(t, err)
	require.NoError(t, ValidateResolverModuleRecordV2(state, "alice.aet", record, 14))

	tooLong := record
	tooLong.RecordTTL = state.Domains[0].ExpiryHeight
	require.ErrorContains(t, ValidateResolverModuleRecordV2(state, "alice.aet", tooLong, 14), string(ResolverModuleFailureTTLExceedsDomainExpiry))

	badInterface := record
	badInterface.InterfaceDescriptors = []InterfaceDescriptorV2{{
		InterfaceID:		"wallet",
		SchemaHash:		InterfaceDescriptorHashPrefixV2 + identityHash("wrong"),
		SchemaInlineOptional:	`{"kind":"wallet"}`,
		Version:		"v1",
		RenderPolicy:		"wallet_confirm",
	}}
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badInterface), "inline schema hash mismatch")

	verified, err := NewReverseResolutionRecordV2(addr(2), "alice.aet", true, 15, state.Domains[0].ExpiryHeight)
	require.NoError(t, err)
	require.NoError(t, ValidateReverseResolutionRecordV2(state, verified, 16, nil))
	reverseKey, err := ResolverModuleReverseStoreKeyV2(verified)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(reverseKey, IdentityStoreV2SpecReversePrefix+"/"))

	mismatch, err := NewReverseResolutionRecordV2(addr(9), "alice.aet", true, 15, state.Domains[0].ExpiryHeight)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateReverseResolutionRecordV2(state, mismatch, 16, nil), "forward primary or authorized alias")

	msg := MsgBatchUpdateResolversV2{
		Auth:	resolverModuleAuth(addr(1), IdentitySignerScopeBatchAdmin, 1),
		Updates: []ResolverBatchUpdateV2{
			{Name: "alice.aet", NameHash: mustDomainHashV2(t, "alice.aet"), Patch: ResolverPatch{Primary: addr(3)}, ExpectedRecordVersion: ResolverRecordVersionV2(state.Resolvers[0]), RecordTTL: 30},
		},
	}
	next, response, err := ExecuteResolverModuleBatchUpdateV2(state, msg, IdentityBatchResolverUpdateOptionsV2{
		Mode:		IdentityBatchFailureAtomicV2,
		Height:		16,
		GasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
		GasLimit:	MinIdentityBatchResolverUpdateGasV2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(1), response.Successes)
	require.NoError(t, next.Validate())

	_, proofErr := BuildResolverModuleProofQueryV2(next, "aetra-test", identityHash("app"), "alice.aet", IdentityProofQueryResolveRecord, 17, 30, nil)
	require.NoError(t, proofErr)
}

func TestResolverModuleBreakdownV2RejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultResolverModuleBreakdownV2()
	require.NoError(t, err)

	missing := breakdown
	missing.Messages = missing.Messages[:len(missing.Messages)-1]
	_, err = NewResolverModuleBreakdownV2(missing)
	require.ErrorContains(t, err, "message entries")

	duplicate := breakdown
	duplicate.Queries[0] = duplicate.Queries[1]
	_, err = NewResolverModuleBreakdownV2(duplicate)
	require.ErrorContains(t, err, "duplicate query")
}

func TestSubdomainModuleBreakdownV2CoversSection133(t *testing.T) {
	breakdown, err := DefaultSubdomainModuleBreakdownV2()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.NotEmpty(t, breakdown.BreakdownHash)
	require.ElementsMatch(t, requiredSubdomainModuleStateObjectsV2(), breakdown.StateObjects)
	require.ElementsMatch(t, requiredSubdomainModuleMessagesV2(), breakdown.Messages)
	require.ElementsMatch(t, requiredSubdomainModuleQueriesV2(), breakdown.Queries)
	require.ElementsMatch(t, requiredSubdomainModuleIntegrationPointsV2(), breakdown.IntegrationPoints)
	require.Contains(t, breakdown.Messages, SubdomainModuleMsgUpdateZonePolicy)
	require.Contains(t, breakdown.Messages, SubdomainModuleMsgDetachSubdomain)
	require.Contains(t, breakdown.Messages, SubdomainModuleMsgRenewSubdomain)
	require.Contains(t, breakdown.BackingPrimitives, "IssueSubdomainV2")
	require.Contains(t, breakdown.BackingPrimitives, "BuildRecursivePolicyProofV2")
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecSubdomainsPrefix)
}

func TestSubdomainModuleValidationV2DelegationZoneIndexAndPath(t *testing.T) {
	state := validIdentityCoreState(t)
	next, subdomain, err := IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:	"alice.aet",
		Label:		"api",
		Actor:		addr(1),
		ChildOwner:	addr(2),
		Height:		20,
		DelegationType:	SubdomainDelegationOwnerControlledV2,
	})
	require.NoError(t, err)
	report, err := BuildSubdomainModuleIndexV2(next)
	require.NoError(t, err)
	require.True(t, report.Valid, report.Issues)
	require.Len(t, report.Index, 1)
	require.Equal(t, "api", report.Index[0].ChildLabel)
	require.True(t, strings.HasPrefix(report.Index[0].StoreKey, IdentityStoreV2SpecSubdomainsPrefix+"/"))

	tooLong := SubdomainCreationPolicyV2{
		ParentName:		"alice.aet",
		Label:			"paid",
		Actor:			addr(1),
		ChildOwner:		addr(2),
		Height:			21,
		ChildExpiryHeight:	next.Domains[0].ExpiryHeight + 10,
		DelegationType:		SubdomainDelegationOwnerControlledV2,
	}
	_, err = ValidateSubdomainCreationV2(next, tooLong)
	require.ErrorContains(t, err, "child expiry")

	detachedMissingPayment := tooLong
	detachedMissingPayment.DelegationType = SubdomainDelegationDetachedPaidV2
	detachedMissingPayment.DetachedPaid = true
	detachedMissingPayment.ParentAuthorization = true
	_, err = ValidateSubdomainCreationV2(next, detachedMissingPayment)
	require.ErrorContains(t, err, "independent payment")

	parentDelegate, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeServiceRecordUpdate, []string{"service.rpc"}, 100, 1, "service.", 10)
	require.NoError(t, err)
	childDelegate := parentDelegate
	childDelegate.ScopeBits |= DelegationScopeBitRoutingRecordUpdateV2
	require.ErrorContains(t, ValidateDelegationDoesNotEscalateV2(parentDelegate, childDelegate), "scope bits")
	delegationKey, err := SubdomainModuleDelegationStoreKeyV2("alice.aet", addr(7), DelegationScopeServiceRecordUpdate)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(delegationKey, IdentityStoreV2SpecDelegationsPrefix+"/"))

	parentPolicy, err := NewZonePolicyV2("alice.aet", ZonePolicyV2{
		AllowedRecordTypes:		[]string{"primary", "service"},
		AllowedServiceTypes:		[]string{"rpc.v1"},
		SubdomainCreationPolicy:	ZoneSubdomainCreationDelegatedV2,
		ResolverUpdatePolicy:		ZoneResolverUpdateDelegatedV2,
		InterfacePolicy:		ZoneInterfacePolicyHashRequiredV2,
		RoutingPolicy:			ZoneRoutingPolicyExplicitTargetsV2,
		MaxChildDepth:			1,
		MaxChildRecords:		2,
		UpdatedAtHeight:		20,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateZonePolicyForSubdomainV2(parentPolicy, subdomain, 1, 2, "primary", "rpc.v1"))
	require.ErrorContains(t, ValidateZonePolicyForSubdomainV2(parentPolicy, subdomain, 2, 2, "primary", "rpc.v1"), "max_child_depth")

	commitment, proof, err := ValidateSubdomainModulePathPolicyV2("alice.aet", "api.alice.aet", 1, parentPolicy.LifecycleEpoch, parentPolicy.LifecycleEpoch, []ZonePolicyV2{parentPolicy})
	require.NoError(t, err)
	require.NotEmpty(t, commitment.CommitmentHash)
	require.NoError(t, ValidateRecursivePolicyProofV2(proof))
}

func TestSubdomainModuleBreakdownV2RejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultSubdomainModuleBreakdownV2()
	require.NoError(t, err)

	missing := breakdown
	missing.StateObjects = missing.StateObjects[:len(missing.StateObjects)-1]
	_, err = NewSubdomainModuleBreakdownV2(missing)
	require.ErrorContains(t, err, "state object entries")
}

func resolverModuleState(t *testing.T) IdentityState {
	t.Helper()
	state := validIdentityCoreState(t)
	var err error
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Contract:	addr(4),
	}, 12)
	require.NoError(t, err)
	require.NoError(t, state.Validate())
	return state
}

func resolverModuleAuth(signer []byte, scope IdentitySignerScopeV2, nonce uint64) IdentityTxAuthV2 {
	return IdentityTxAuthV2{
		ChainID:			"aetra-local-1",
		Signer:				signer,
		Scope:				scope,
		NameNormalizationVersion:	NameNormalizationVersionV2,
		Nonce:				nonce,
		Fee:				1,
		StorageCost:			1,
	}
}
