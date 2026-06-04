package app

import (
	"sort"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	dextypes "github.com/sovereign-l1/l1/x/dex/types"
)

func TestGetMaccPermsReturnsDefensiveCopy(t *testing.T) {
	perms := GetMaccPerms()
	require.Equal(t, []string{authtypes.Minter}, perms[minttypes.ModuleName])

	perms[minttypes.ModuleName][0] = authtypes.Burner
	delete(perms, minttypes.ModuleName)

	fresh := GetMaccPerms()
	require.Equal(t, []string{authtypes.Minter}, fresh[minttypes.ModuleName])
}

func TestBlockedAddressesExcludesGovAuthority(t *testing.T) {
	blocked := BlockedAddresses()

	require.False(t, blocked[authtypes.NewModuleAddress(govtypes.ModuleName).String()])
	require.True(t, blocked[authtypes.NewModuleAddress(dextypes.ModuleName).String()])
}

func TestGetStoreKeysReturnsStableOrder(t *testing.T) {
	app := &L1App{keys: newKVStoreKeys()}

	storeKeys := app.GetStoreKeys()
	names := make([]string, 0, len(storeKeys))
	for _, key := range storeKeys {
		names = append(names, key.Name())
	}

	require.True(t, sort.StringsAreSorted(names), names)
	require.Contains(t, names, dextypes.StoreKey)
}
