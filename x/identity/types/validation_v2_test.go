package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNameNormalizationV2ValidAndInvalidVectors(t *testing.T) {
	valid := []string{
		"alice.aet",
		"api.alice.aet",
		"dex_api1.alice.aet",
		"l1.l2.l3.l4.l5.l6.l7.l8.aet",
	}
	for _, name := range valid {
		result, err := NormalizeAETDomainVersioned(name, NameNormalizationVersionV2)
		require.NoError(t, err, name)
		require.Equal(t, name, result.NormalizedName)
		require.Equal(t, NameNormalizationVersionV2, result.Version)
		require.NotEmpty(t, result.NameHash)
		require.LessOrEqual(t, len(result.Labels), MaxDomainLabels)
	}

	tooLongLabel := strings.Repeat("a", MaxDomainFullBytes)
	invalid := map[string]string{
		"Alice.aet":			"lowercase",
		"alice":			"must end",
		"алиса.aet":			"ASCII",
		".alice.aet":			"leading or trailing separator",
		"alice..aet":			"leading or trailing separator",
		"admin.aet":			"reserved",
		"xn--alice.aet":		"punycode",
		"bad--dash.aet":		"repeated hyphen",
		"-bad.aet":			"start or end",
		"bad-.aet":			"start or end",
		tooLongLabel + DomainTLD:	"<= 253",
		"a.b.c.d.e.f.g.h.i.aet":	"not exceed",
	}
	for name, want := range invalid {
		_, err := NormalizeAETDomainVersioned(name, NameNormalizationVersionV2)
		require.ErrorContains(t, err, want, name)
	}
}

func TestNameNormalizationV2MigrationAndTxVersionRejection(t *testing.T) {
	result, err := NormalizeAETDomainVersioned("alice.aet", NameNormalizationVersionV2)
	require.NoError(t, err)

	same, err := MigrateNameNormalizationVersionV2(result, NameNormalizationVersionV2)
	require.NoError(t, err)
	require.Equal(t, result.NormalizedName, same.NormalizedName)
	require.Equal(t, result.NameHash, same.NameHash)

	_, err = NormalizeAETDomainVersioned("alice.aet", NameNormalizationVersionV2+1)
	require.ErrorContains(t, err, "unsupported identity name normalization version")

	auth := txAuth(IdentitySignerScopeResolverUpdate, 9)
	auth.NameNormalizationVersion = NameNormalizationVersionV2 + 1
	msg := MsgUpdateResolverRecordV2{
		Auth:			auth,
		Name:			"alice.aet",
		NameHash:		result.NameHash,
		Patch:			ResolverPatch{Primary: addr(1)},
		ExpectedRecordVersion:	1,
		RecordTTL:		10,
	}
	require.ErrorContains(t, msg.ValidateBasic(), "unsupported identity name normalization version")
}

func TestOwnershipValidationV2TransferInvariantsAndRepairFlow(t *testing.T) {
	record := validationDomainRecord(t, "alice.aet", addr(1), 10, 100)
	binding := validationBinding(record, addr(1), 10)

	require.NoError(t, ValidatePreTransferOwnershipV2(record, binding, addr(1), 20))
	racedBinding := binding
	racedBinding.Owner = addr(8)
	require.ErrorContains(t, ValidatePreTransferOwnershipV2(record, racedBinding, addr(1), 20), "registry owner must equal nft owner")

	nextRecord, nextBinding, err := TransferDomainNFTBindingWithInvariantsV2(record, binding, addr(1), addr(2), 20)
	require.NoError(t, err)
	require.True(t, addressesEqual(nextRecord.Owner, addr(2)))
	require.True(t, addressesEqual(nextBinding.Owner, addr(2)))
	require.Equal(t, uint64(20), nextRecord.UpdatedAtHeight)
	require.Equal(t, uint64(20), nextBinding.LastVerifiedHeight)

	badPost := nextBinding
	badPost.Owner = addr(3)
	require.ErrorContains(t, ValidatePostTransferOwnershipV2(record, binding, nextRecord, badPost, addr(2), 20), "atomically update")

	restricted, err := RestrictDomainRecordV2ForBrokenBinding(record, racedBinding, addr(1), 21)
	require.NoError(t, err)
	repaired, repairedBinding, err := RepairDomainNFTBindingInternalFailureV2(restricted, racedBinding, addr(1), 22)
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Active, repaired.Status)
	require.Zero(t, repaired.Flags&DomainRecordV2FlagRestricted)
	require.True(t, addressesEqual(repairedBinding.Owner, record.Owner))
}

func TestOwnershipValidationV2ResolverAndSubdomainAuthorization(t *testing.T) {
	parent := validationDomainRecord(t, "alice.aet", addr(1), 10, 100)
	require.NoError(t, ValidateResolverUpdateAuthorizationV2(parent, addr(1), nil, ResolverKeyPrimary, 20))
	require.NoError(t, ValidateSubdomainCreationAuthorizationV2(parent, addr(1), nil, "api", 1, 20))

	delegate, err := NewDelegationRecordV2("alice.aet", addr(2), DelegationScopeResolverUpdate, []string{ResolverKeyPrimary}, 60, 0, ResolverKeyPrimary, 10)
	require.NoError(t, err)
	require.NoError(t, ValidateResolverUpdateAuthorizationV2(parent, addr(2), &delegate, ResolverKeyPrimary, 20))
	require.ErrorContains(t, ValidateResolverUpdateAuthorizationV2(parent, addr(3), &delegate, ResolverKeyPrimary, 20), "delegate mismatch")

	subDelegate, err := NewDelegationRecordV2("alice.aet", addr(4), DelegationScopeSubdomainCreate, []string{"create"}, 60, 2, "", 10)
	require.NoError(t, err)
	require.NoError(t, ValidateSubdomainCreationAuthorizationV2(parent, addr(4), &subDelegate, "api", 1, 20))
	require.ErrorContains(t, ValidateSubdomainCreationAuthorizationV2(parent, addr(5), &subDelegate, "api", 1, 20), "delegate mismatch")

	expired := parent
	expired.Status = DomainRecordV2Expired
	require.ErrorContains(t, ValidateResolverUpdateAuthorizationV2(expired, addr(1), nil, ResolverKeyPrimary, 120), "expired domain owner")
	require.NoError(t, ValidateResolverUpdateAuthorizationV2(expired, addr(1), nil, ResolverRecoveryMetadataKeyV2, 120))
	require.ErrorContains(t, ValidateSubdomainCreationAuthorizationV2(expired, addr(1), nil, "api", 1, 120), "active parent")
}

func validationDomainRecord(t *testing.T, name string, owner sdk.AccAddress, createdHeight uint64, expiryHeight uint64) DomainRecordV2 {
	t.Helper()
	record, err := NewDomainRecordV2FromDomain(Domain{
		Name:			name,
		Owner:			owner,
		NFTID:			"nft-" + name,
		RegisteredHeight:	createdHeight,
		ExpiryHeight:		expiryHeight,
		UpdatedHeight:		createdHeight,
	}, DomainRecordV2Active, 0, createdHeight)
	require.NoError(t, err)
	return record
}

func validationBinding(record DomainRecordV2, owner sdk.AccAddress, height uint64) DomainNFTBinding {
	return DomainNFTBinding{
		NameHash:		record.NameHash,
		NFTClassID:		record.NFTClassID,
		NFTItemID:		record.NFTItemID,
		Owner:			owner,
		LastVerifiedHeight:	height,
		BindingVersion:		1,
	}
}
