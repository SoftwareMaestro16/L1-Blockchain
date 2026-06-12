package keeper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/identity-root/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const authority = prototype.DefaultAuthority

var (
	ownerA	= mustAE("11")
	ownerB	= mustAE("22")
)

func setupKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	gs.IdentityParams.RegistrationPeriod = 100
	gs.IdentityParams.RenewalPeriod = 50
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func resolverRoot(seed string) string {
	return strings.Repeat(seed, 64)
}

func mustAE(hexByte string) string {
	bz, err := addressing.Parse("4:000000000000000000000000" + strings.Repeat(hexByte, 20))
	if err != nil {
		panic(err)
	}
	text, err := addressing.FormatUserFriendly(bz)
	if err != nil {
		panic(err)
	}
	return text
}

func TestDefaultGenesisDisabled(t *testing.T) {
	gs := DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.NotEmpty(t, gs.State.RootAuthorities)
}

func TestRegisterName(t *testing.T) {
	k := setupKeeper(t)

	record, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "Alice", Height: 10})
	require.NoError(t, err)
	require.Equal(t, "alice.aet", record.Name)
	require.Equal(t, ownerA, record.Owner)
	require.Equal(t, uint64(110), record.ExpiryHeight)
	require.Equal(t, types.DomainRentPayerOwner, record.RentPayerPolicy)
	require.Equal(t, uint64(10), record.LastStorageChargeHeight)

	_, found, err := k.NameRecord("ALICE.AET")
	require.NoError(t, err)
	require.True(t, found)
}

func TestTransferName(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)

	record, err := k.TransferName(types.MsgTransferName{Owner: ownerA, Name: "alice", NewOwner: ownerB, Height: 20})
	require.NoError(t, err)
	require.Equal(t, ownerB, record.Owner)

	_, err = k.TransferName(types.MsgTransferName{Owner: ownerA, Name: "alice", NewOwner: ownerA, Height: 21})
	require.ErrorContains(t, err, "requires owner")
}

func TestResolverUpdateAndResolve(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)

	resolver, err := k.SetResolver(types.MsgSetResolver{Owner: ownerA, Name: "alice", ResolverRoot: resolverRoot("a"), Height: 11})
	require.NoError(t, err)
	require.Equal(t, resolverRoot("a"), resolver.ResolverRoot)

	_, resolved, active, err := k.ResolveName("alice", 12)
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, resolverRoot("a"), resolved.ResolverRoot)
}

func TestReverseRecord(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)

	reverse, err := k.SetReverseRecord(types.MsgSetReverseRecord{Owner: ownerA, Address: "addr-1", Name: "alice", Height: 12})
	require.ErrorContains(t, err, "AE user-facing")

	reverse, err = k.SetReverseRecord(types.MsgSetReverseRecord{Owner: ownerA, Address: ownerA, Name: "alice", Height: 12})
	require.NoError(t, err)
	require.Equal(t, "alice.aet", reverse.Name)

	queried, found, err := k.ReverseRecord(ownerA)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, reverse, queried)
}

func TestExpiryAndRenewal(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)

	_, _, active, err := k.ResolveName("alice", 110)
	require.NoError(t, err)
	require.False(t, active)

	renewed, err := k.RenewName(types.MsgRenewName{Owner: ownerA, Name: "alice", Height: 111})
	require.NoError(t, err)
	require.Equal(t, uint64(161), renewed.ExpiryHeight)
	_, _, active, err = k.ResolveName("alice", 120)
	require.NoError(t, err)
	require.True(t, active)
}

func TestReservedNameRejectedForNormalUser(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.ReserveName(types.MsgReserveName{Authority: authority, Name: "admin", Reason: "root"})
	require.NoError(t, err)

	_, err = k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "admin", Height: 10})
	require.ErrorContains(t, err, "reserved")
}

func TestSubdomainOwnershipFollowsParentPolicy(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10, SubdomainPolicy: types.SubdomainPolicyOwnerOnly})
	require.NoError(t, err)

	_, err = k.CreateSubdomain(types.MsgCreateSubdomain{Owner: ownerA, ParentName: "alice", Label: "app", SubdomainOwner: ownerB, Height: 11})
	require.ErrorContains(t, err, "parent policy")
	record, err := k.CreateSubdomain(types.MsgCreateSubdomain{Owner: ownerA, ParentName: "alice", Label: "app", Height: 11})
	require.NoError(t, err)
	require.Equal(t, "app.alice.aet", record.Name)
	require.Equal(t, ownerA, record.Owner)
}

func TestExportImportPreservesSortedRecords(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "zeta", Height: 10})
	require.NoError(t, err)
	_, err = k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alpha", Height: 10})
	require.NoError(t, err)

	exported := k.ExportGenesis()
	require.Equal(t, "alpha.aet", exported.State.Records[0].Name)
	var imported Keeper
	require.NoError(t, imported.InitGenesis(exported))
	reexported := imported.ExportGenesis()
	require.Equal(t, exported.State.Records, reexported.State.Records)
}

func TestNFTBindingRequiredWhenEnabled(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	gs.IdentityParams.NFTBindingEnabled = true
	require.NoError(t, k.InitGenesis(gs))

	_, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.ErrorContains(t, err, "NFT binding")
	record, err := k.RegisterName(types.MsgRegisterName{
		Owner:	ownerA,
		Name:	"alice",
		Height:	10,
		NFTBinding: types.IdentityNFTBindingReference{
			Enabled:	true,
			ClassID:	"identity",
			NFTID:		"alice",
			Owner:		ownerA,
		},
	})
	require.NoError(t, err)
	require.True(t, record.NFTBinding.Enabled)
	require.Equal(t, ownerA, record.NFTBinding.Owner)
}

func TestGetOwnerReturnsCorrectOwnerAfterRegisterAndTransfer(t *testing.T) {
	k := setupKeeper(t)
	record, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)
	require.Equal(t, ownerA, record.Owner)

	queried, found, err := k.NameRecord("alice")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ownerA, queried.Owner)

	transferred, err := k.TransferName(types.MsgTransferName{Owner: ownerA, Name: "alice", NewOwner: ownerB, Height: 20})
	require.NoError(t, err)
	require.Equal(t, ownerB, transferred.Owner)

	queried, found, err = k.NameRecord("alice")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ownerB, queried.Owner)
}

func TestDomainOwnerMustBeAEAddress(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.RegisterName(types.MsgRegisterName{Owner: "owner-a", Name: "alice", Height: 10})
	require.ErrorContains(t, err, "AE user-facing")
}

func TestDomainStorageRentAccruesByPolicy(t *testing.T) {
	k := setupKeeper(t)
	record, err := k.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)
	require.Zero(t, record.StorageRentDebt)

	renewed, err := k.RenewName(types.MsgRenewName{Owner: ownerA, Name: "alice", Height: 20})
	require.NoError(t, err)
	require.NotZero(t, renewed.StorageRentDebt)
	require.Equal(t, uint64(20), renewed.LastStorageChargeHeight)

	protocol := setupKeeper(t)
	gs := protocol.ExportGenesis()
	gs.IdentityParams.DefaultDomainRentPayerPolicy = types.DomainRentPayerProtocol
	require.NoError(t, protocol.InitGenesis(gs))
	record, err = protocol.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "bob", Height: 10})
	require.NoError(t, err)
	require.Equal(t, types.DomainRentPayerProtocol, record.RentPayerPolicy)
	renewed, err = protocol.RenewName(types.MsgRenewName{Owner: ownerA, Name: "bob", Height: 20})
	require.NoError(t, err)
	require.Zero(t, renewed.StorageRentDebt)
	require.Equal(t, uint64(20), renewed.LastStorageChargeHeight)
}

func TestDomainRegistryExportImportPreservesOwnerResolverAndRent(t *testing.T) {
	source := setupKeeper(t)
	_, err := source.RegisterName(types.MsgRegisterName{Owner: ownerA, Name: "alice", Height: 10})
	require.NoError(t, err)
	_, err = source.SetResolver(types.MsgSetResolver{Owner: ownerA, Name: "alice", ResolverRoot: resolverRoot("b"), Height: 12})
	require.NoError(t, err)
	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())

	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	record, found, err := target.NameRecord("alice.aet")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ownerA, record.Owner)
	require.NotZero(t, record.StorageRentDebt)
	_, resolver, active, err := target.ResolveName("alice", 13)
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, resolverRoot("b"), resolver.ResolverRoot)
}
