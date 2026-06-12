package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDelegationRecordV2ValidatesScopePermissionsAndUse(t *testing.T) {
	record, err := NewDelegationRecordV2(
		"api.alice.aet",
		addr(7),
		DelegationScopeServiceRecordUpdate,
		[]string{"service.rpc", "service.grpc"},
		100,
		2,
		"service.",
		10,
	)
	require.NoError(t, err)
	expectedHash, err := DomainRecordV2NameHash("api.alice.aet")
	require.NoError(t, err)
	require.Equal(t, expectedHash, record.NameHash)
	require.Equal(t, []string{"service.grpc", "service.rpc"}, record.Permissions)
	require.NoError(t, ValidateDelegationRecordV2Use(record, DelegationScopeServiceRecordUpdate, "service.rpc", "service.rpc", 1, 50))
	require.ErrorContains(t, ValidateDelegationRecordV2Use(record, DelegationScopeRoutingRecordUpdate, "service.rpc", "service.rpc", 1, 50), "scope mismatch")
	require.ErrorContains(t, ValidateDelegationRecordV2Use(record, DelegationScopeServiceRecordUpdate, "service.rpc", "interface.aw5", 1, 50), "record prefix limit")
	require.ErrorContains(t, ValidateDelegationRecordV2Use(record, DelegationScopeServiceRecordUpdate, "service.rpc", "service.rpc", 3, 50), "subtree limit")
	require.ErrorContains(t, ValidateDelegationRecordV2Use(record, DelegationScopeServiceRecordUpdate, "service.rpc", "service.rpc", 1, 100), "expired")
}

func TestDelegationRecordV2RejectsInvalidCanonicalState(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	record := DelegationRecordV2{
		NameHash:		nameHash,
		Delegate:		addr(7),
		Scope:			DelegationScopeResolverUpdate,
		Permissions:		[]string{ResolverKeyWallet, ResolverKeyContract},
		ExpiresAtHeight:	100,
		CreatedAtHeight:	10,
	}
	require.ErrorContains(t, ValidateDelegationRecordV2(record), "permissions must be sorted")

	record.Permissions = []string{ResolverKeyContract, ResolverKeyWallet}
	record.Scope = DelegationScopeV2("bad_scope")
	require.ErrorContains(t, ValidateDelegationRecordV2(record), "invalid identity v2 delegation scope")

	record.Scope = DelegationScopeResolverUpdate
	record.ExpiresAtHeight = 10
	require.ErrorContains(t, ValidateDelegationRecordV2(record), "expires_at_height")
}

func TestAuctionRecordV2BuildsFromSealedAuctionLifecycle(t *testing.T) {
	state := EmptyIdentityState(DefaultIdentityParams())
	state, auction, err := StartSealedAuction(state, "market", 100)
	require.NoError(t, err)
	leftCommit, err := ComputeAuctionCommitment("market.aet", addr(1), 100, "left")
	require.NoError(t, err)
	rightCommit, err := ComputeAuctionCommitment("market.aet", addr(2), 150, "right")
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(2), rightCommit, 101)
	require.NoError(t, err)
	state, auction, err = CommitAuctionBid(state, auction.Name, addr(1), leftCommit, 102)
	require.NoError(t, err)

	commitRecord, err := BuildAuctionRecordV2(auction, 100, "domain.fees")
	require.NoError(t, err)
	require.Equal(t, AuctionRecordV2Commit, commitRecord.Status)
	require.Equal(t, auction.RevealStartHeight, commitRecord.CommitEndHeight)
	require.Equal(t, uint64(0), commitRecord.WinningBid)
	require.Empty(t, commitRecord.Winner)
	require.Equal(t, uint64(0), commitRecord.RevealedBidsCount)
	require.Equal(t, ComputeAuctionSealedCommitmentsRootV2(auction.Commitments), commitRecord.SealedCommitmentsRoot)

	revealHeight := auction.RevealStartHeight
	state, _, err = RevealAuctionBid(state, auction.Name, addr(1), 100, "left", revealHeight)
	require.NoError(t, err)
	state, _, err = RevealAuctionBid(state, auction.Name, addr(2), 150, "right", revealHeight+1)
	require.NoError(t, err)
	_, auction, err = FinalizeSealedAuction(state, auction.Name, auction.RevealEndHeight)
	require.NoError(t, err)

	finalRecord, err := BuildAuctionRecordV2(auction, 100, "domain.fees")
	require.NoError(t, err)
	expectedHash, err := DomainRecordV2NameHash("market.aet")
	require.NoError(t, err)
	require.Equal(t, identityHash("identity-v2-auction", "market.aet", "00000000000000000100"), finalRecord.AuctionID)
	require.Equal(t, expectedHash, finalRecord.NameHash)
	require.Equal(t, AuctionRecordV2Finalized, finalRecord.Status)
	require.Equal(t, uint64(150), finalRecord.WinningBid)
	require.Equal(t, addr(2), finalRecord.Winner)
	require.Equal(t, uint64(2), finalRecord.RevealedBidsCount)
	require.Equal(t, "domain.fees", finalRecord.FeeSplitID)
	require.NoError(t, ValidateAuctionRecordV2(finalRecord))
}

func TestAuctionRecordV2RejectsInvalidFinalizedState(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("market.aet")
	require.NoError(t, err)
	record := AuctionRecordV2{
		AuctionID:		identityHash("identity-v2-auction", "market.aet", "00000000000000000100"),
		NameHash:		nameHash,
		Status:			AuctionRecordV2Finalized,
		CommitStartHeight:	100,
		CommitEndHeight:	200,
		RevealStartHeight:	200,
		RevealEndHeight:	300,
		MinBid:			100,
		WinningBid:		50,
		SealedCommitmentsRoot:	identityHash("commitments"),
		RevealedBidsCount:	1,
		FeeSplitID:		"domain.fees",
	}
	require.ErrorContains(t, ValidateAuctionRecordV2(record), "auction winner")

	record.Winner = addr(1)
	require.ErrorContains(t, ValidateAuctionRecordV2(record), "winning_bid")
}
