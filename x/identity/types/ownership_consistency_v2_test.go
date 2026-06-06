package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityNFTTransferHooksUpdateOrRejectAtomicallyV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, domain.Name, addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state.PendingResolverUpdates = []ResolverUpdateIntent{{Domain: domain.Name, Actor: addr(1), Nonce: 1}}
	require.NoError(t, state.Validate())

	_, _, err = ApplyIdentityNFTTransferHookStateV2(state, domain.NFTID, addr(9), 20, IdentityNFTTransferHookRejectOnRegistryMismatchV2)
	require.ErrorContains(t, err, "rejected")

	next, transferred, err := ApplyIdentityNFTTransferHookStateV2(state, domain.NFTID, addr(9), 20, IdentityNFTTransferHookUpdateRegistryV2)
	require.NoError(t, err)
	require.Equal(t, addr(9), transferred.Owner)
	require.Equal(t, addr(9), next.DomainNFTs[0].Owner)
	require.Equal(t, addr(9), next.Resolvers[0].Owner)
	require.Empty(t, next.PendingResolverUpdates)

	registryNext, registryDomain, err := ApplyIdentityRegistryTransferHookStateV2(next, domain.Name, addr(9), addr(7), 21)
	require.NoError(t, err)
	require.Equal(t, addr(7), registryDomain.Owner)
	require.Equal(t, addr(7), registryNext.DomainNFTs[0].Owner)
}

func TestIdentityNFTTransferHooksV2RecordBindingAtomicity(t *testing.T) {
	record := validationDomainRecord(t, "alice.aet", addr(1), 10, 100)
	binding := validationBinding(record, addr(1), 10)

	_, _, err := ApplyIdentityNFTTransferHookV2(record, binding, addr(9), 20, IdentityNFTTransferHookRejectOnRegistryMismatchV2)
	require.ErrorContains(t, err, "rejected")

	nextRecord, nextBinding, err := ApplyIdentityNFTTransferHookV2(record, binding, addr(9), 20, IdentityNFTTransferHookUpdateRegistryV2)
	require.NoError(t, err)
	require.Equal(t, addr(9), nextRecord.Owner)
	require.Equal(t, addr(9), nextBinding.Owner)

	registryRecord, registryBinding, err := ApplyIdentityRegistryTransferHookV2(nextRecord, nextBinding, addr(9), addr(7), 21)
	require.NoError(t, err)
	require.Equal(t, addr(7), registryRecord.Owner)
	require.Equal(t, addr(7), registryBinding.Owner)
}

func TestBrokenNFTBindingBlocksResolverUntilRepairV2(t *testing.T) {
	record := validationDomainRecord(t, "alice.aet", addr(1), 10, 100)
	binding := validationBinding(record, addr(9), 10)
	restricted, err := RestrictDomainRecordV2ForBrokenBinding(record, binding, addr(9), 20)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateResolverUpdateAuthorizationV2(restricted, addr(1), nil, ResolverKeyPrimary, 21), "broken nft binding")

	restricted.Owner = addr(9)
	repaired, repairedBinding, err := RepairDomainNFTBinding(restricted, binding, addr(9), addr(9), 22)
	require.NoError(t, err)
	require.NoError(t, ValidateResolverUpdateAuthorizationV2(repaired, addr(9), nil, ResolverKeyPrimary, 23))
	require.Equal(t, addr(9), repairedBinding.Owner)
}

func TestIdentityConsistencyAuditDetectsInvariantsV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, domain.Name, addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state.DomainNFTs[0].Owner = addr(9)
	state.Resolvers = append(state.Resolvers, ResolverRecord{Domain: "ghost.aet", Owner: addr(4), Primary: addr(5), UpdatedAtUnix: 13})
	state.ReverseRecords = append(state.ReverseRecords, ReverseRecord{Address: addr(3), Domain: domain.Name, UpdatedAtUnix: 13})
	state.Subdomains = append(state.Subdomains, SubdomainRecord{ParentName: "missing.aet", Name: "api.missing.aet", Owner: addr(6), CreatedHeight: 14})
	state.Auctions = append(state.Auctions, Auction{Name: domain.Name, CommitStartHeight: 15, RevealStartHeight: 20, RevealEndHeight: 25, Phase: AuctionPhaseCommit})
	delegation, err := NewDelegationRecordV2(domain.Name, addr(7), DelegationScopeResolverUpdate, []string{ResolverKeyPrimary}, domain.ExpiryHeight+1, 0, ResolverKeyPrimary, 15)
	require.NoError(t, err)

	audit := QueryIdentityPeriodicConsistencyAuditV2(IdentityConsistencyAuditRequestV2{State: state, Height: 16, Delegations: []DelegationRecordV2{delegation}})
	require.False(t, audit.Valid)
	requireAuditIssueV2(t, audit, "active_domain_without_nft_binding")
	requireAuditIssueV2(t, audit, "resolver_without_domain")
	requireAuditIssueV2(t, audit, "verified_reverse_without_forward_consistency")
	requireAuditIssueV2(t, audit, "subdomain_without_parent")
	requireAuditIssueV2(t, audit, "active_auction_for_active_name")
	requireAuditIssueV2(t, audit, "delegation_outlives_domain")
}

func TestIdentityConsistencyAuditExportMigrationAndQueryV2(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	require.NoError(t, ValidateIdentityStateExportV2(state, 12, nil))
	require.NoError(t, ValidateIdentityStateMigrationV2(state, state.Export(), 12, nil))

	after := state.Export()
	after.Domains = nil
	after.DomainNFTs = nil
	require.ErrorContains(t, ValidateIdentityStateMigrationV2(state, after, 12, nil), "removed domains")

	service := NewIdentityQueryServiceV2(IdentityQueryContextV2{State: state, Height: 12})
	resp := service.QueryPeriodicConsistencyAudit()
	require.Equal(t, IdentityQueryOK, resp.Code)
	require.NotNil(t, resp.Consistency)
	require.True(t, resp.Consistency.Valid)

	broken := state.Export()
	broken.DomainNFTs[0].Owner = addr(9)
	service = NewIdentityQueryServiceV2(IdentityQueryContextV2{State: broken, Height: 12})
	resp = service.QueryPeriodicConsistencyAudit()
	require.Equal(t, IdentityQueryVerificationFailed, resp.Code)
	require.False(t, resp.Consistency.Valid)
}

func requireAuditIssueV2(t *testing.T, audit IdentityConsistencyAuditResultV2, code string) {
	t.Helper()
	for _, issue := range audit.Issues {
		if issue.Code == code {
			return
		}
	}
	require.Failf(t, "missing audit issue", "expected issue code %s in %#v", code, audit.Issues)
}
