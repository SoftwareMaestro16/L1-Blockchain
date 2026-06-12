package app

import (
	"github.com/cosmos/cosmos-sdk/types/module"

	appupgrades "github.com/sovereign-l1/l1/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the sample L1App upgrade
// from v053 to v054.
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.53.x to v0.54.x.
const UpgradeName = appupgrades.Name

func (app L1App) RegisterUpgradeHandlers() {
	appupgrades.RegisterHandlers(appupgrades.HandlerDependencies{
		UpgradeKeeper:	app.UpgradeKeeper,
		ModuleManager:	app.ModuleManager,
		Configurator:	app.Configurator(),
		SetStoreLoader:	app.SetStoreLoader,
	})
}

func ValidateUpgradeVersionMap(fromVM, currentVM module.VersionMap, allowedNewModules ...string) error {
	return appupgrades.ValidateVersionMap(fromVM, currentVM, allowedNewModules...)
}
