package accounts

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
)

func TestModuleAccountPermissionsAreCloned(t *testing.T) {
	perms := ModuleAccountPermissions()
	perms[mintauthoritytypes.ModuleName] = nil

	fresh := ModuleAccountPermissions()
	require.Equal(t, []string{authtypes.Minter}, fresh[mintauthoritytypes.ModuleName])
}

func TestReservedSystemModuleAccountWiringValidatesBlockedPolicy(t *testing.T) {
	blocked := BlockedAddresses()

	require.NoError(t, ValidateReservedSystemModuleAccountWiring(blocked))

	mint, found := ReservedSystemModuleAccountByName("AETMint")
	require.True(t, found)
	require.False(t, mint.CanReceiveUserFunds)

	addr, found, err := ReservedSystemModuleAccountAddress(mint.ModuleAccountName)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, blocked[addr.String()])
}
