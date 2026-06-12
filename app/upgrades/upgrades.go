package upgrades

import (
	"context"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const Name = "v053-to-v054"

type HandlerDependencies struct {
	UpgradeKeeper	*upgradekeeper.Keeper
	ModuleManager	*module.Manager
	Configurator	module.Configurator
	SetStoreLoader	func(baseapp.StoreLoader)
}

func RegisterHandlers(deps HandlerDependencies) {
	deps.UpgradeKeeper.SetUpgradeHandler(
		Name,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdk.UnwrapSDKContext(ctx).Logger().Debug("this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
			if err := ValidateVersionMap(fromVM, deps.ModuleManager.GetVersionMap()); err != nil {
				return nil, err
			}
			return deps.ModuleManager.RunMigrations(ctx, deps.Configurator, fromVM)
		},
	)

	upgradeInfo, err := deps.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == Name && !deps.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{},
		}
		deps.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func ValidateVersionMap(fromVM, currentVM module.VersionMap, allowedNewModules ...string) error {
	allowed := make(map[string]bool, len(allowedNewModules))
	for _, moduleName := range allowedNewModules {
		allowed[moduleName] = true
	}

	moduleNames := make([]string, 0, len(currentVM))
	for moduleName := range currentVM {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)

	for _, moduleName := range moduleNames {
		currentVersion := currentVM[moduleName]
		if currentVersion == 0 {
			return fmt.Errorf("invalid current module version for %s: 0", moduleName)
		}
		fromVersion, found := fromVM[moduleName]
		if !found {
			if allowed[moduleName] {
				continue
			}
			return fmt.Errorf("missing module version for %s", moduleName)
		}
		if fromVersion > currentVersion {
			return fmt.Errorf("module %s version %d is newer than current version %d", moduleName, fromVersion, currentVersion)
		}
	}

	return nil
}
