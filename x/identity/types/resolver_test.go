package types

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestApplyResolverUpdateSetAndChange(t *testing.T) {
	now := int64(100)
	owner := addr(1)
	wallet := addr(2)
	contract := addr(3)
	domain := activeDomainRecord(owner, now+1000)

	record, event, err := ApplyResolverUpdate(nil, domain, owner, ResolverUpdate{
		Domain:		"alice.aet",
		Primary:	wallet,
		Records:	map[string]sdk.AccAddress{ResolverKeyContract: contract},
		Metadata:	[]byte("profile"),
		UpdatedAtUnix:	now,
	}, nil, now)
	require.NoError(t, err)
	require.Equal(t, ResolverEventSet, event.Type)
	require.Equal(t, "alice.aet", event.Domain)
	require.Equal(t, wallet, record.Primary)
	require.Equal(t, contract, record.Records[ResolverKeyContract])

	nextWallet := addr(4)
	changed, event, err := ApplyResolverUpdate(&record, domain, owner, ResolverUpdate{
		Domain:		"alice.aet",
		Primary:	nextWallet,
		Records:	map[string]sdk.AccAddress{ResolverKeyWallet: nextWallet},
		UpdatedAtUnix:	now + 1,
	}, nil, now)
	require.NoError(t, err)
	require.Equal(t, ResolverEventChanged, event.Type)
	require.Equal(t, nextWallet, changed.Primary)
}

func TestApplyResolverUpdateRejectsZeroAndUnauthorized(t *testing.T) {
	now := int64(100)
	owner := addr(1)
	manager := addr(2)
	domain := activeDomainRecord(owner, now+1000)

	_, _, err := ApplyResolverUpdate(nil, domain, manager, ResolverUpdate{
		Domain:		"alice.aet",
		Primary:	addr(3),
		UpdatedAtUnix:	now,
	}, nil, now)
	require.ErrorContains(t, err, "unauthorized")

	_, _, err = ApplyResolverUpdate(nil, domain, owner, ResolverUpdate{
		Domain:		"alice.aet",
		Primary:	bytes.Repeat([]byte{0}, 20),
		UpdatedAtUnix:	now,
	}, nil, now)
	require.ErrorContains(t, err, "resolver primary")
}

func TestDelegatedResolverManager(t *testing.T) {
	now := int64(100)
	owner := addr(1)
	manager := addr(2)
	domain := activeDomainRecord(owner, now+1000)
	grant := &ResolverGrant{
		Domain:		"alice.aet",
		Owner:		owner,
		Manager:	manager,
		Keys:		[]string{ResolverKeyPrimary, ResolverKeyWallet},
		ExpiresAtUnix:	now + 100,
	}

	record, _, err := ApplyResolverUpdate(nil, domain, manager, ResolverUpdate{
		Domain:		"alice.aet",
		Primary:	addr(3),
		Records:	map[string]sdk.AccAddress{ResolverKeyWallet: addr(3)},
		UpdatedAtUnix:	now,
	}, grant, now)
	require.NoError(t, err)
	require.Equal(t, addr(3), record.Primary)

	_, _, err = ApplyResolverUpdate(&record, domain, manager, ResolverUpdate{
		Domain:		"alice.aet",
		Records:	map[string]sdk.AccAddress{ResolverKeyDEX: addr(4)},
		UpdatedAtUnix:	now + 1,
	}, grant, now)
	require.ErrorContains(t, err, "does not allow")

	grant.ExpiresAtUnix = now
	_, _, err = ApplyResolverUpdate(&record, domain, manager, ResolverUpdate{
		Domain:		"alice.aet",
		Primary:	addr(5),
		UpdatedAtUnix:	now + 1,
	}, grant, now)
	require.ErrorContains(t, err, "expired")
}

func TestResolvePaymentTarget(t *testing.T) {
	now := int64(100)
	owner := addr(1)
	target := addr(2)
	domain := activeDomainRecord(owner, now+1000)
	record := resolverRecord(owner, "alice.aet", target)

	resolved, err := ResolvePaymentTarget(record, domain, now)
	require.NoError(t, err)
	require.Equal(t, target, resolved)

	record.Primary = nil
	_, err = ResolvePaymentTarget(record, domain, now)
	require.ErrorContains(t, err, "domain not resolved")

	record.Primary = target
	domain.ExpiryUnix = now
	_, err = ResolvePaymentTarget(record, domain, now)
	require.ErrorContains(t, err, "expired")
}

func TestSubdomainResolution(t *testing.T) {
	now := int64(100)
	owner := addr(1)
	target := addr(2)
	domain := activeDomainRecord(owner, now+1000)
	record := resolverRecord(owner, "dex.alice.aet", target)

	resolved, err := ResolvePaymentTarget(record, domain, now)
	require.NoError(t, err)
	require.Equal(t, target, resolved)

	record.Domain = "dex.bob.aet"
	_, err = ResolvePaymentTarget(record, domain, now)
	require.ErrorContains(t, err, "base does not match")
}

func TestReverseResolution(t *testing.T) {
	now := int64(100)
	owner := addr(1)
	target := addr(2)
	domain := activeDomainRecord(owner, now+1000)
	record := resolverRecord(owner, "alice.aet", target)

	reverse, event, err := SetReverseResolution(domain, record, target, target, now)
	require.NoError(t, err)
	require.Equal(t, ResolverEventReverse, event.Type)
	require.Equal(t, "alice.aet", reverse.Domain)
	require.Equal(t, target, reverse.Address)

	_, _, err = SetReverseResolution(domain, record, addr(3), target, now)
	require.ErrorContains(t, err, "must control")

	_, _, err = SetReverseResolution(domain, record, addr(3), addr(3), now)
	require.ErrorContains(t, err, "does not point")
}

func TestResolverRecordOwnerMatchesRegistryOwner(t *testing.T) {
	now := int64(100)
	domain := activeDomainRecord(addr(1), now+1000)
	record := resolverRecord(addr(9), "alice.aet", addr(2))
	err := ValidateResolverRecordForDomain(record, domain, now)
	require.ErrorContains(t, err, "owner must match")
}

func TestValidateResolverKeyAndBounds(t *testing.T) {
	require.NoError(t, ValidateResolverKey("app_endpoint"))
	require.NoError(t, ValidateResolverKey(ResolverKeyWallet))
	require.Error(t, ValidateResolverKey("Bad"))
	require.Error(t, ValidateResolverKey("bad key"))

	record := resolverRecord(addr(1), "alice.aet", addr(2))
	record.Metadata = bytes.Repeat([]byte{1}, MaxResolverMetadataBytes+1)
	require.ErrorContains(t, ValidateResolverRecord(record), "metadata")
}

func activeDomainRecord(owner sdk.AccAddress, expiry int64) DomainRecord {
	return DomainRecord{
		Name:		"alice",
		TLD:		DomainTLD,
		Owner:		owner,
		ExpiryUnix:	expiry,
		NFTItemID:	DomainNFTItemID("alice"),
		Status:		DomainStatusActive,
		CreatedAtUnix:	1,
		UpdatedAtUnix:	2,
	}
}

func resolverRecord(owner sdk.AccAddress, domain string, primary sdk.AccAddress) ResolverRecord {
	return ResolverRecord{
		Domain:		domain,
		Owner:		owner,
		Primary:	primary,
		Records:	map[string]sdk.AccAddress{ResolverKeyWallet: primary},
		UpdatedAtUnix:	1,
	}
}

func addr(seed byte) sdk.AccAddress {
	out := make([]byte, 20)
	out[19] = seed
	return sdk.AccAddress(out)
}
