package app

import "fmt"

func (app *L1App) ValidateAetraCoreWiringGate() error {
	if app == nil || app.ModuleManager == nil {
		return fmt.Errorf("aether core wiring gate requires initialized app")
	}
	if AetraCoreRoutingExecutionPoint() != RoutingExecutionPointAnteAdmissionOnly {
		return fmt.Errorf("unsupported routing execution point %s", AetraCoreRoutingExecutionPoint())
	}
	prototypeModuleNames := AetraCorePrototypeModuleNames()
	prototypeStoreKeys := AetraCorePrototypeStoreKeys()
	if len(prototypeModuleNames) != len(prototypeStoreKeys) {
		return fmt.Errorf("prototype module/store key count mismatch")
	}
	moduleAccountPermissions := GetMaccPerms()
	for i, moduleName := range prototypeModuleNames {
		if _, found := app.ModuleManager.Modules[moduleName]; !found {
			return fmt.Errorf("prototype module %s is not registered", moduleName)
		}
		storeKey := prototypeStoreKeys[i]
		if _, found := app.keys[storeKey]; !found {
			return fmt.Errorf("prototype module %s store key %s is not mounted", moduleName, storeKey)
		}
		if _, found := moduleAccountPermissions[moduleName]; found && !IsReservedSystemModuleAccountName(moduleName) {
			return fmt.Errorf("prototype module %s must not have module account permissions", moduleName)
		}
	}
	systemModuleNames := AetraCoreSystemModuleNames()
	systemStoreKeys := AetraCoreSystemStoreKeys()
	if len(systemModuleNames) != len(systemStoreKeys) {
		return fmt.Errorf("system module/store key count mismatch")
	}
	for i, moduleName := range systemModuleNames {
		if _, found := app.ModuleManager.Modules[moduleName]; !found {
			return fmt.Errorf("system module %s is not registered", moduleName)
		}
		storeKey := systemStoreKeys[i]
		if _, found := app.keys[storeKey]; !found {
			return fmt.Errorf("system module %s store key %s is not mounted", moduleName, storeKey)
		}
		if _, found := moduleAccountPermissions[moduleName]; found && !IsReservedSystemModuleAccountName(moduleName) {
			return fmt.Errorf("system module %s must not have module account permissions", moduleName)
		}
	}
	return nil
}
