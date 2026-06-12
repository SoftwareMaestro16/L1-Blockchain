package types

import (
	"bytes"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentitySpecRegisterAETDomain(t *testing.T) {
	state, domain := registerSpecDomain(t, "Alice.AET", addr(1), "salt", 10)

	require.Equal(t, "alice.aet", domain.Name)
	require.Equal(t, addr(1), domain.Owner)
	require.Equal(t, DomainLifecycleActive, mustLifecycle(t, state, "alice.aet", 11))
	require.Len(t, state.DomainNFTs, 1)
	require.Equal(t, domain.NFTID, state.DomainNFTs[0].ID)
}

func TestIdentitySpecDuplicateNormalizedNameRejected(t *testing.T) {
	state, _ := registerSpecDomain(t, "Alice.AET", addr(1), "salt", 10)
	commitment, err := ComputeRegistrationCommitment(" alice.aet ", addr(2), "salt2")
	require.NoError(t, err)

	_, err = CommitDomainRegistration(state, "ALICE", addr(2), commitment, 20)
	require.ErrorContains(t, err, "not available")

	state.Domains = append(state.Domains, state.Domains[0])
	require.ErrorContains(t, state.Validate(), "duplicate")
}

func TestIdentitySpecExpiredDomainCannotResolve(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := SetIdentityResolver(state, "alice.aet", addr(1), ResolverUpdate{
		Primary: addr(2),
	}, 11)
	require.NoError(t, err)

	_, err = ResolveIdentityAddress(state, "alice.aet", domain.ExpiryHeight)
	require.ErrorContains(t, err, "expired")
	require.True(t, IsDomainAvailable(state, "alice.aet", domain.ExpiryHeight))
}

func TestIdentitySpecRenewalPreservesOwnership(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	renewAt := domain.ExpiryHeight - 1

	next, renewed, err := RenewIdentityDomain(state, "alice.aet", addr(1), renewAt)
	require.NoError(t, err)
	require.Equal(t, addr(1), renewed.Owner)
	require.Greater(t, renewed.ExpiryHeight, domain.ExpiryHeight)
	require.Equal(t, DomainLifecycleActive, mustLifecycle(t, next, "alice.aet", domain.ExpiryHeight+1))
}

func TestIdentitySpecResolverUpdateRequiresOwner(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)

	_, _, err := SetIdentityResolver(state, "alice.aet", addr(2), ResolverUpdate{Primary: addr(3)}, 11)
	require.ErrorContains(t, err, "requires owner")

	next, record, err := SetIdentityResolver(state, "alice.aet", addr(1), ResolverUpdate{
		Primary:	addr(3),
		Contract:	addr(4),
		ZoneEndpoint:	"contract-zone/0:1",
	}, 11)
	require.NoError(t, err)
	require.Equal(t, addr(3), record.Primary)
	require.Equal(t, addr(4), record.Contract)
	require.Equal(t, "contract-zone/0:1", record.ZoneEndpoint)
	require.Len(t, next.Resolvers, 1)
}

func TestIdentitySpecReverseLookupRequiresAddressOwner(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := SetIdentityResolver(state, "alice.aet", addr(1), ResolverUpdate{Primary: addr(2)}, 11)
	require.NoError(t, err)

	_, _, err = SetIdentityReverse(state, addr(3), addr(2), "alice.aet", 12)
	require.ErrorContains(t, err, "address owner")

	next, reverse, err := SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 12)
	require.NoError(t, err)
	require.Equal(t, "alice.aet", reverse.Domain)
	require.Len(t, next.ReverseRecords, 1)
}

func TestIdentitySpecSubdomainIssuanceRequiresParentOwner(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)

	_, _, err := IssueSubdomain(state, "alice.aet", "dex", addr(2), addr(3), false, 11)
	require.ErrorContains(t, err, "parent owner")

	next, subdomain, err := IssueSubdomain(state, "alice.aet", "dex", addr(1), addr(3), false, 11)
	require.NoError(t, err)
	require.Equal(t, "dex.alice.aet", subdomain.Name)

	_, _, err = SetIdentityResolver(next, "dex.alice.aet", addr(3), ResolverUpdate{Primary: addr(4)}, 12)
	require.NoError(t, err)
}

func TestIdentitySpecParentControlledSubdomainResolver(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := IssueSubdomain(state, "alice.aet", "vault", addr(1), addr(3), true, 11)
	require.NoError(t, err)

	_, _, err = SetIdentityResolver(state, "vault.alice.aet", addr(3), ResolverUpdate{Primary: addr(4)}, 12)
	require.ErrorContains(t, err, "parent owner")

	_, _, err = SetIdentityResolver(state, "vault.alice.aet", addr(1), ResolverUpdate{Primary: addr(4)}, 12)
	require.NoError(t, err)
}

func TestIdentitySpecNFTTransferChangesOwnerAndInvalidatesPendingResolverUpdates(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state.PendingResolverUpdates = []ResolverUpdateIntent{{Domain: "alice.aet", Actor: addr(1), Nonce: 1}}
	require.NoError(t, state.Validate())

	next, domain, err := TransferDomainNFT(state, "alice.aet", addr(1), addr(9), 20)
	require.NoError(t, err)
	require.Equal(t, addr(9), domain.Owner)
	require.Empty(t, next.PendingResolverUpdates)
	require.Equal(t, addr(9), next.DomainNFTs[0].Owner)

	_, _, err = SetIdentityResolver(next, "alice.aet", addr(1), ResolverUpdate{Primary: addr(2)}, 21)
	require.ErrorContains(t, err, "requires owner")
}

func TestIdentitySpecAuctionTieBreakerDeterministic(t *testing.T) {
	state := EmptyIdentityState(DefaultIdentityParams())
	state, auction, err := StartSealedAuction(state, "market", 100)
	require.NoError(t, err)

	leftCommit, err := ComputeAuctionCommitment("market.aet", addr(1), 100, "left")
	require.NoError(t, err)
	rightCommit, err := ComputeAuctionCommitment("market.aet", addr(2), 100, "right")
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(2), rightCommit, 101)
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(1), leftCommit, 102)
	require.NoError(t, err)

	revealHeight := auction.RevealStartHeight
	state, _, err = RevealAuctionBid(state, auction.Name, addr(2), 100, "right", revealHeight)
	require.NoError(t, err)
	state, _, err = RevealAuctionBid(state, auction.Name, addr(1), 100, "left", revealHeight)
	require.NoError(t, err)

	next, finalized, err := FinalizeSealedAuction(state, auction.Name, auction.RevealEndHeight)
	require.NoError(t, err)
	expectedWinner := addr(1)
	if bytes.Compare([]byte(rightCommit), []byte(leftCommit)) < 0 {
		expectedWinner = addr(2)
	}
	require.Equal(t, expectedWinner, finalized.Winner)
	require.Len(t, finalized.Refunds, 1)
	require.Equal(t, AuctionPhaseFinalized, next.Auctions[0].Phase)
}

func TestIdentitySpecExportImportPreservesDomainLifecycle(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := SetIdentityResolver(state, domain.Name, addr(1), ResolverUpdate{Primary: addr(2)}, 11)
	require.NoError(t, err)

	exported := state.Export()
	imported, err := ImportIdentityState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported, imported))
	require.Equal(t, DomainLifecycleActive, mustLifecycle(t, imported, "alice.aet", 12))
}

func registerSpecDomain(t *testing.T, name string, owner sdk.AccAddress, salt string, height uint64) (IdentityState, Domain) {
	t.Helper()
	state := EmptyIdentityState(DefaultIdentityParams())
	commitment, err := ComputeRegistrationCommitment(name, owner, salt)
	require.NoError(t, err)
	state, err = CommitDomainRegistration(state, name, owner, commitment, height)
	require.NoError(t, err)
	state, domain, err := RevealRegisterDomain(state, name, owner, salt, height+1)
	require.NoError(t, err)
	return state, domain
}

func mustLifecycle(t *testing.T, state IdentityState, name string, height uint64) DomainLifecycleStatus {
	t.Helper()
	status, err := DomainLifecycle(state, name, height)
	require.NoError(t, err)
	return status
}
