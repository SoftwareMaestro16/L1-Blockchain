package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDomainRecordV2FromDomainValidatesCanonicalState(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)
	record.Resolver = addr(1)

	err = ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{
		CurrentHeight:	12,
		NFTOwner:	state.DomainNFTs[0].Owner,
	})
	require.NoError(t, err)
	require.Equal(t, record.Name, record.NormalizedName)
	require.Equal(t, DomainTLD, record.TLD)
	require.Len(t, record.NameHash, 64)
	require.Empty(t, record.ParentNameHash)
}

func TestDomainRecordV2ParentHashUsesImmediateParent(t *testing.T) {
	parentHash, err := DomainRecordV2NameHash("dex.alice.aet")
	require.NoError(t, err)
	childHash, err := DomainRecordV2ParentNameHash("api.dex.alice.aet")
	require.NoError(t, err)
	require.Equal(t, parentHash, childHash)

	parent, found, err := ImmediateParentAETDomain("api.dex.alice.aet")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "dex.alice.aet", parent)
}

func TestDomainRecordV2RejectsHashAndParentMismatch(t *testing.T) {
	_, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)
	record.NameHash = identityHash("wrong")
	require.ErrorContains(t, ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{CurrentHeight: 12, NFTOwner: addr(1)}), "name_hash")

	record, err = NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)
	record.ParentNameHash = identityHash("wrong-parent")
	require.ErrorContains(t, ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{CurrentHeight: 12, NFTOwner: addr(1)}), "parent_name_hash")
}

func TestDomainRecordV2EnforcesNFTOwnerAndExpiry(t *testing.T) {
	_, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)

	require.ErrorContains(t, ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{CurrentHeight: 12, NFTOwner: addr(9)}), "nft owner")
	require.ErrorContains(t, ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{CurrentHeight: record.ExpiryHeight, NFTOwner: addr(1)}), "expiry_height")
}

func TestDomainRecordV2ResolverMustBeOwnerOrDelegate(t *testing.T) {
	_, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 1_000, 12)
	require.NoError(t, err)
	record.Resolver = addr(7)

	err = ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{
		CurrentHeight:	12,
		NFTOwner:	addr(1),
	})
	require.ErrorContains(t, err, "authorized delegate")

	err = ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{
		CurrentHeight:		12,
		NFTOwner:		addr(1),
		ResolverDelegates:	[]sdk.AccAddress{addr(7)},
	})
	require.NoError(t, err)
}
