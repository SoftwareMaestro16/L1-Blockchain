package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityReadOnlyQueryAuditDoesNotMutateStateV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)
	before := state.Export()

	resp, audit, err := AuditIdentityReadOnlyQueryV2(state, 14, "resolve_primary", "alice.aet", func(q IdentityQueryServiceV2) IdentityQueryResponseV2 {
		return q.QueryResolvePrimary("alice.aet")
	})
	require.NoError(t, err)
	require.Equal(t, IdentityQueryOK, resp.Code)
	require.False(t, audit.MutationDetected)
	require.Empty(t, audit.ConsensusWrites)
	require.Equal(t, audit.BeforeRoot, audit.AfterRoot)
	require.Equal(t, before, state.Export())
	require.Len(t, audit.Observability, 1)
	require.True(t, audit.Observability[0].ConsensusFree)

	_, proofAudit, err := AuditIdentityReadOnlyQueryV2(state, 14, "resolution_proof", "alice.aet", func(q IdentityQueryServiceV2) IdentityQueryResponseV2 {
		return q.QueryResolutionProof("alice.aet")
	})
	require.NoError(t, err)
	require.False(t, proofAudit.MutationDetected)
	require.Equal(t, proofAudit.BeforeRoot, proofAudit.AfterRoot)

	_, reverseAudit, err := AuditIdentityReadOnlyQueryV2(state, 14, "verified_reverse", "alice.aet", func(q IdentityQueryServiceV2) IdentityQueryResponseV2 {
		return q.QueryVerifiedReverse(addr(2), nil)
	})
	require.NoError(t, err)
	require.False(t, reverseAudit.MutationDetected)
	require.Equal(t, reverseAudit.BeforeRoot, reverseAudit.AfterRoot)
}

func TestIdentityLookupObservabilityIsEventOnlyV2(t *testing.T) {
	event, err := NewIdentityLookupObservabilityEventV2("alice.aet", "direct", 20, 10)
	require.NoError(t, err)
	require.Equal(t, IdentityObservabilityLookupVolumeV2, event.Type)
	require.True(t, event.ConsensusFree)
	require.Equal(t, uint64(10), event.Count)

	telemetry, err := BuildIdentityVoteExtensionTelemetryV2(true, 20, 99)
	require.NoError(t, err)
	require.True(t, telemetry.Enabled)
	require.Equal(t, IdentityABCIEventVoteTelemetryV2, telemetry.Event.Type)
	disabled, err := BuildIdentityVoteExtensionTelemetryV2(false, 0, 99)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)
	require.Nil(t, disabled.Event)
}

func TestIdentityABCIPlusPrecheckAndProposalGroupingV2(t *testing.T) {
	aliceHash := mustDomainHashV2(t, "alice.aet")
	bobHash := mustDomainHashV2(t, "bob.aet")
	valid := MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 1), Name: "alice.aet", NameHash: aliceHash, Patch: ResolverPatch{Primary: addr(2)}, ExpectedRecordVersion: 1, RecordTTL: 30}
	require.True(t, PrecheckIdentityABCIPlusTxV2(valid, 20).Accepted)

	malformed := valid
	malformed.RecordTTL = 0
	rejected := PrecheckIdentityABCIPlusTxV2(malformed, 20)
	require.False(t, rejected.Accepted)
	require.NotNil(t, rejected.Event)
	require.Equal(t, IdentityABCIEventMalformedRejectedV2, rejected.Event.Type)

	grouping, err := GroupIdentityProposalUpdatesV2([]IdentityMsgV2{
		valid,
		MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 2), Name: "bob.aet", NameHash: bobHash, Patch: ResolverPatch{Primary: addr(3)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		MsgRenewDomainV2{Auth: txAuth(IdentitySignerScopeOwner, 3), Name: "alice.aet", NameHash: aliceHash, ExpectedRecordVersion: 1},
	}, 20)
	require.NoError(t, err)
	require.Equal(t, IdentityProposalGroupingOrderNameHashV2, grouping.Order)
	require.Len(t, grouping.Groups, 2)
	require.Equal(t, aliceHash, grouping.Groups[0].GroupKey)
	require.Equal(t, []uint32{0, 2}, grouping.Groups[0].Indexes)
	require.Equal(t, bobHash, grouping.Groups[1].GroupKey)
	require.Equal(t, []uint32{1}, grouping.Groups[1].Indexes)
	require.Len(t, grouping.Events, 2)
	require.Equal(t, IdentityABCIEventProposalGroupV2, grouping.Events[0].Type)
}

func TestIdentityABCIPlusFinalizeBoundedExpiryAndCacheEventsV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt-a", 10)
	state, _ = registerSpecDomainInState(t, state, "bob", addr(2), "salt-b", 10)
	state, _ = registerSpecDomainInState(t, state, "carol", addr(3), "salt-c", 10)
	for i := range state.Domains {
		switch state.Domains[i].Name {
		case "bob.aet":
			state.Domains[i].ExpiryHeight = 18
		case "alice.aet":
			state.Domains[i].ExpiryHeight = 19
		case "carol.aet":
			state.Domains[i].ExpiryHeight = 30
		}
	}
	state = state.Export()
	require.NoError(t, state.Validate())
	cacheAlice, err := NewResolutionCacheRecordV2("alice.aet", identityHash("path-a"), identityHash("record-a"), 100, 7, 2, 3)
	require.NoError(t, err)
	cacheBob, err := NewResolutionCacheRecordV2("bob.aet", identityHash("path-b"), identityHash("record-b"), 100, 7, 2, 3)
	require.NoError(t, err)

	resp, err := FinalizeIdentityABCIPlusV2(IdentityFinalizeRequestV2{
		State:		state,
		Height:		20,
		ExpiryLimit:	1,
		CacheRecords:	[]ResolutionCacheRecordV2{cacheAlice, cacheBob},
	})
	require.NoError(t, err)
	require.Equal(t, state, resp.State)
	require.Len(t, resp.ExpiredDomains, 1)
	require.Equal(t, "bob.aet", resp.ExpiredDomains[0].Name)
	require.Len(t, resp.Events, 2)
	require.Equal(t, IdentityABCIEventDomainExpiredV2, resp.Events[0].Type)
	require.Equal(t, IdentityABCIEventCacheInvalidatedV2, resp.Events[1].Type)
	require.Equal(t, uint64(100), resp.InvalidatedCaches[0].ValidUntilHeight)
	require.Equal(t, uint64(0), resp.InvalidatedCaches[1].ValidUntilHeight)

	resp2, err := FinalizeIdentityABCIPlusV2(IdentityFinalizeRequestV2{State: state, Height: 20, ExpiryLimit: 2, CacheRecords: []ResolutionCacheRecordV2{cacheAlice, cacheBob}})
	require.NoError(t, err)
	require.Equal(t, []string{"bob.aet", "alice.aet"}, []string{resp2.ExpiredDomains[0].Name, resp2.ExpiredDomains[1].Name})
	require.Equal(t, IdentityABCIEventDomainExpiredV2, resp2.Events[0].Type)
	require.Equal(t, "bob.aet", resp2.Events[0].Name)
	require.Equal(t, "alice.aet", resp2.Events[2].Name)
}
