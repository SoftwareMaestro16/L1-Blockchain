package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/accounts"
)

type ReservedSystemModuleAccount = accounts.ReservedSystemModuleAccount

func GetMaccPerms() map[string][]string {
	return accounts.ModuleAccountPermissions()
}

func BlockedAddresses() map[string]bool {
	return accounts.BlockedAddresses()
}

func ReservedSystemModuleAccounts() []ReservedSystemModuleAccount {
	return accounts.ReservedSystemModuleAccounts()
}

func ReservedSystemModuleAccountByName(name string) (ReservedSystemModuleAccount, bool) {
	return accounts.ReservedSystemModuleAccountByName(name)
}

func ReservedSystemModuleAccountByModuleAccountName(moduleAccountName string) (ReservedSystemModuleAccount, bool) {
	return accounts.ReservedSystemModuleAccountByModuleAccountName(moduleAccountName)
}

func IsReservedSystemModuleAccountName(moduleAccountName string) bool {
	return accounts.IsReservedSystemModuleAccountName(moduleAccountName)
}

func ReservedSystemModuleAccountAddress(moduleAccountName string) (sdk.AccAddress, bool, error) {
	return accounts.ReservedSystemModuleAccountAddress(moduleAccountName)
}

func ValidateReservedSystemModuleAccountWiring(blocked map[string]bool) error {
	return accounts.ValidateReservedSystemModuleAccountWiring(blocked)
}
