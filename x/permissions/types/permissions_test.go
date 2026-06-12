package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPermissionGrantCheckExpiryAndRevocation(t *testing.T) {
	registry := NewRegistry()
	permission := permission("p1", ScopeResolverDelegate, "alice.aet", 10, 20)
	require.NoError(t, registry.Grant(permission))

	require.False(t, registry.Check(permission.Grantee, ScopeResolverDelegate, "alice.aet", 9))
	require.True(t, registry.Check(permission.Grantee, ScopeResolverDelegate, "alice.aet", 10))
	require.False(t, registry.Check(permission.Grantee, ScopeResolverDelegate, "alice.aet", 20))

	require.NoError(t, registry.Revoke("p1", permission.Owner, 15))
	require.True(t, registry.Check(permission.Grantee, ScopeResolverDelegate, "alice.aet", 14))
	require.False(t, registry.Check(permission.Grantee, ScopeResolverDelegate, "alice.aet", 15))
}

func TestPermissionValidation(t *testing.T) {
	require.ErrorContains(t, permission("", ScopeDomainManager, "alice.aet", 1, 10).Validate(), "id")
	require.ErrorContains(t, permission("p1", "unknown", "alice.aet", 1, 10).Validate(), "scope")
	require.ErrorContains(t, permission("p1", ScopeDomainManager, "", 1, 10).Validate(), "resource")
	require.ErrorContains(t, permission("p1", ScopeDomainManager, "alice.aet", 10, 10).Validate(), "after grant")

	p := permission("p1", ScopeDomainManager, "alice.aet", 1, 10)
	p.Owner = permAddr(0)
	require.ErrorContains(t, p.Validate(), "must not be zero address")
}

func TestPermissionChecksAreDeterministicAndNoHiddenSuperuser(t *testing.T) {
	registry := NewRegistry()
	p2 := permission("b", ScopeContractExtension, "wallet-ext", 1, 10)
	p1 := permission("a", ScopeGovernance, "fees.params", 1, 10)
	require.NoError(t, registry.Grant(p2))
	require.NoError(t, registry.Grant(p1))

	listed := registry.List()
	require.Equal(t, "a", listed[0].ID)
	require.Equal(t, "b", listed[1].ID)
	require.True(t, registry.Check(p1.Grantee, ScopeGovernance, "fees.params", 5))
	require.False(t, registry.Check(p1.Grantee, ScopeEmergency, "fees.params", 5))
	require.False(t, HasHiddenSuperuserBypass())
}

func permission(id, scope, resource string, grant uint64, expiry uint64) Permission {
	return Permission{
		ID:			id,
		Owner:			permAddr(0x11),
		Grantee:		permAddr(0x22),
		Scope:			scope,
		Resource:		resource,
		GrantedAtHeight:	grant,
		ExpiresAtHeight:	expiry,
	}
}

func permAddr(fill byte) sdk.AccAddress {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return sdk.AccAddress(out)
}
