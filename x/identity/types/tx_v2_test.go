package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityTxV2CoreMessagesValidate(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	parentHash, err := DomainRecordV2NameHash("parent.aet")
	require.NoError(t, err)
	auctionID := identityHash("auction-id")
	commitmentHash := identityHash("commitment")
	reverse, err := NewReverseResolutionRecordV2(addr(2), "alice.aet", true, 10, 100)
	require.NoError(t, err)
	delegation, err := NewDelegationRecordV2("parent.aet", addr(7), DelegationScopeSubdomainCreate, []string{"create"}, 100, 1, "", 10)
	require.NoError(t, err)

	msgs := []IdentityMsgV2{
		MsgCommitRegistrationV2{
			Auth:			txAuth(IdentitySignerScopeRegistration, 1),
			Name:			"alice.aet",
			NameHash:		nameHash,
			CommitmentHash:		commitmentHash,
			CommitmentVersion:	1,
			SaltHashOptional:	identityHash("salt"),
		},
		MsgRevealRegistrationV2{
			Auth:			txAuth(IdentitySignerScopeRegistration, 2),
			Name:			"alice.aet",
			NameHash:		nameHash,
			CommitmentHash:		commitmentHash,
			CommitmentVersion:	1,
			Salt:			"salt",
		},
		MsgRegisterDirectV2{Auth: txAuth(IdentitySignerScopeRegistration, 3), Name: "alice.aet", NameHash: nameHash, Owner: addr(1), ExpectedRecordVersion: 1},
		MsgRenewDomainV2{Auth: txAuth(IdentitySignerScopeOwner, 4), Name: "alice.aet", NameHash: nameHash, ExpectedRecordVersion: 1},
		MsgTransferDomainV2{Auth: txAuth(IdentitySignerScopeOwner, 5), Name: "alice.aet", NameHash: nameHash, NewOwner: addr(3), ExpectedRecordVersion: 1},
		MsgSetResolverV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 6), Name: "alice.aet", NameHash: nameHash, Resolver: addr(4), ExpectedRecordVersion: 1},
		MsgUpdateResolverRecordV2{Auth: txAuth(IdentitySignerScopeResolverUpdate, 7), Name: "alice.aet", NameHash: nameHash, Patch: ResolverPatch{Primary: addr(2)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		MsgSetReverseRecordV2{Auth: txAuth(IdentitySignerScopeReverseUpdate, 8), Record: reverse, ExpectedRecordVersion: 1},
		MsgVerifyReverseRecordV2{Auth: txAuth(IdentitySignerScopeReverseUpdate, 9), Record: reverse, AuthorizedAliasKeys: []string{ResolverKeyWallet}, ExpectedRecordVersion: 1},
		MsgCreateSubdomainV2{Auth: txAuth(IdentitySignerScopeSubdomainAdmin, 10), ParentName: "parent.aet", ParentNameHash: parentHash, Label: "api", ChildOwner: addr(5), ExpectedParentVersion: 1},
		MsgCreateSubdomainV2{Auth: txAuth(IdentitySignerScopeSubdomainAdmin, 10), ParentName: "parent.aet", ParentNameHash: parentHash, Label: "paid", ChildOwner: addr(5), DelegationType: SubdomainDelegationDetachedPaidV2, ChildExpiryHeight: 200, DetachedPaid: true, IndependentPayment: true, ParentAuthorization: true, ExpectedParentVersion: 1},
		MsgDelegateSubdomainV2{Auth: txAuth(IdentitySignerScopeDelegationAdmin, 11), Delegation: delegation, ExpectedRecordVersion: 1},
		MsgRevokeDelegationV2{Auth: txAuth(IdentitySignerScopeDelegationAdmin, 12), Name: "parent.aet", NameHash: parentHash, Delegate: addr(7), Scope: DelegationScopeSubdomainCreate, ExpectedRecordVersion: 1},
		MsgStartAuctionV2{Auth: txAuth(IdentitySignerScopeAuctionAdmin, 13), Name: "alice.aet", NameHash: nameHash, MinBid: 100, FeeSplitID: "domain.fees"},
		MsgCommitBidV2{Auth: txAuth(IdentitySignerScopeAuctionBidder, 14), AuctionID: auctionID, NameHash: nameHash, CommitmentHash: commitmentHash},
		MsgRevealBidV2{Auth: txAuth(IdentitySignerScopeAuctionBidder, 15), AuctionID: auctionID, NameHash: nameHash, Bid: 100, Salt: "bid-salt", CommitmentHash: commitmentHash},
		MsgFinalizeAuctionV2{Auth: txAuth(IdentitySignerScopeAuctionAdmin, 16), AuctionID: auctionID, NameHash: nameHash, ExpectedAuctionVersion: 1},
		MsgExpireDomainV2{Auth: txAuth(IdentitySignerScopeOwner, 17), Name: "alice.aet", NameHash: nameHash, ExpectedRecordVersion: 1},
		MsgBatchUpdateResolversV2{Auth: txAuth(IdentitySignerScopeBatchAdmin, 18), Updates: []ResolverBatchUpdateV2{{Name: "alice.aet", NameHash: nameHash, Patch: ResolverPatch{Primary: addr(2)}, ExpectedRecordVersion: 1, RecordTTL: 30}}},
		MsgBatchRenewDomainsV2{Auth: txAuth(IdentitySignerScopeBatchAdmin, 19), Renewals: []RenewDomainBatchItemV2{{Name: "alice.aet", NameHash: nameHash, ExpectedRecordVersion: 1}}},
		MsgInvalidateResolutionCacheV2{Auth: txAuth(IdentitySignerScopeCacheAdmin, 20), NameHash: nameHash, ResolutionPathHash: identityHash("path"), SourceVersion: 1, ParentEpoch: 2, ChildEpoch: 3, ExpectedRecordVersion: 1},
	}
	require.Len(t, msgs, 21)
	for _, msg := range msgs {
		require.NoError(t, ValidateIdentityMsgV2(msg), msg.IdentityMessageName())
		require.NotEmpty(t, msg.IdentityMessageName())
		require.NotEmpty(t, msg.SignerAddress())
		require.NotEmpty(t, msg.SignerScope())
	}
}

func TestIdentityTxV2RejectsDomainSeparationPaymentScopeAndVersion(t *testing.T) {
	msg := MsgRenewDomainV2{Auth: txAuth(IdentitySignerScopeOwner, 1), Name: "alice.aet", ExpectedRecordVersion: 1}
	require.NoError(t, msg.ValidateBasic())

	msg.Auth.ChainID = ""
	require.ErrorContains(t, msg.ValidateBasic(), "chain_id")
	msg.Auth = txAuth(IdentitySignerScopeResolverUpdate, 1)
	require.ErrorContains(t, msg.ValidateBasic(), "signer scope")
	msg.Auth = txAuth(IdentitySignerScopeOwner, 0)
	require.ErrorContains(t, msg.ValidateBasic(), "nonce")
	msg.Auth = txAuth(IdentitySignerScopeOwner, 1)
	msg.Auth.Fee = 0
	require.ErrorContains(t, msg.ValidateBasic(), "fee payment")
	msg.Auth = txAuth(IdentitySignerScopeOwner, 1)
	msg.Auth.StorageCost = 0
	require.ErrorContains(t, msg.ValidateBasic(), "storage-cost")
	msg.Auth = txAuth(IdentitySignerScopeOwner, 1)
	msg.ExpectedRecordVersion = 0
	require.ErrorContains(t, msg.ValidateBasic(), "expected record version")
}

func TestIdentityTxV2RejectsReplayAndCommitRevealFields(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	commit := MsgCommitRegistrationV2{
		Auth:			txAuth(IdentitySignerScopeRegistration, 1),
		Name:			"alice.aet",
		NameHash:		nameHash,
		CommitmentHash:		identityHash("commitment"),
		CommitmentVersion:	1,
	}
	require.NoError(t, commit.ValidateBasic())

	commit.Auth.Nonce = 0
	require.ErrorContains(t, commit.ValidateBasic(), "nonce")
	commit.Auth = txAuth(IdentitySignerScopeRegistration, 1)
	commit.CommitmentHash = "bad"
	require.ErrorContains(t, commit.ValidateBasic(), "commitment hash")

	reveal := MsgRevealBidV2{Auth: txAuth(IdentitySignerScopeAuctionBidder, 2), AuctionID: identityHash("auction"), Bid: 100, Salt: "salt", CommitmentHash: identityHash("commitment")}
	require.NoError(t, reveal.ValidateBasic())
	reveal.Salt = ""
	require.ErrorContains(t, reveal.ValidateBasic(), "salt")
	reveal.Salt = "salt"
	reveal.Bid = 0
	require.ErrorContains(t, reveal.ValidateBasic(), "bid amount")
}

func TestIdentityTxV2RejectsNameHashMismatchAndDuplicateBatches(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	wrongHash, err := DomainRecordV2NameHash("bob.aet")
	require.NoError(t, err)
	msg := MsgTransferDomainV2{Auth: txAuth(IdentitySignerScopeOwner, 1), Name: "alice.aet", NameHash: wrongHash, NewOwner: addr(3), ExpectedRecordVersion: 1}
	require.ErrorContains(t, msg.ValidateBasic(), "name_hash mismatch")

	batch := MsgBatchUpdateResolversV2{
		Auth:	txAuth(IdentitySignerScopeBatchAdmin, 2),
		Updates: []ResolverBatchUpdateV2{
			{Name: "alice.aet", NameHash: nameHash, Patch: ResolverPatch{Primary: addr(2)}, ExpectedRecordVersion: 1, RecordTTL: 30},
			{Name: "alice.aet", NameHash: nameHash, Patch: ResolverPatch{Contract: addr(3)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		},
	}
	require.ErrorContains(t, batch.ValidateBasic(), "duplicate domain")
}

func txAuth(scope IdentitySignerScopeV2, nonce uint64) IdentityTxAuthV2 {
	return IdentityTxAuthV2{
		ChainID:			"aetra-local-1",
		Signer:				addr(byte(nonce + 10)),
		Scope:				scope,
		NameNormalizationVersion:	NameNormalizationVersionV2,
		Nonce:				nonce,
		Fee:				1,
		StorageCost:			1,
	}
}
