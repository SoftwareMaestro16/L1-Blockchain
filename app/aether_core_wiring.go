package app

import (
	"fmt"
	"slices"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	actorregistrytypes "github.com/sovereign-l1/l1/x/actor-registry/types"
	aethercoretypes "github.com/sovereign-l1/l1/x/aethercore/types"
	avmschedulertypes "github.com/sovereign-l1/l1/x/avm-scheduler/types"
	bridgehubtypes "github.com/sovereign-l1/l1/x/bridge-hub/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	constitutiontypes "github.com/sovereign-l1/l1/x/constitution/types"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	identityroottypes "github.com/sovereign-l1/l1/x/identity-root/types"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

type RoutingExecutionPoint string

const (
	// Routing remains an admission/ante-level executable spec until a coordinated
	// upgrade adds public Msg services and production persistence semantics.
	RoutingExecutionPointAnteAdmissionOnly RoutingExecutionPoint = "ANTE_ADMISSION_ONLY"
)

var aetherCorePrototypeModules = []string{
	aethercoretypes.ModuleName,
	loadtypes.ModuleName,
	routingtypes.ModuleName,
	zonestypes.ModuleName,
	meshtypes.ModuleName,
	networkingtypes.ModuleName,
	paymentstypes.ModuleName,
	schedulertypes.ModuleName,
	avmschedulertypes.ModuleName,
	actorregistrytypes.ModuleName,
	storagerenttypes.ModuleName,
	identityroottypes.ModuleName,
	bridgehubtypes.ModuleName,
}

var aetherCoreSystemModules = []string{
	constitutiontypes.ModuleName,
	systemregistrytypes.ModuleName,
	validatorelectiontypes.ModuleName,
	validatorregistrytypes.ModuleName,
	configtypes.ModuleName,
}

func AetherCoreRoutingExecutionPoint() RoutingExecutionPoint {
	return RoutingExecutionPointAnteAdmissionOnly
}

func AetherCorePrototypeModuleNames() []string {
	return slices.Clone(aetherCorePrototypeModules)
}

func AetherCorePrototypeStoreKeys() []string {
	return []string{
		aethercoretypes.StoreKey,
		loadtypes.StoreKey,
		routingtypes.StoreKey,
		zonestypes.StoreKey,
		meshtypes.StoreKey,
		networkingtypes.StoreKey,
		paymentstypes.StoreKey,
		schedulertypes.StoreKey,
		avmschedulertypes.StoreKey,
		actorregistrytypes.StoreKey,
		storagerenttypes.StoreKey,
		identityroottypes.StoreKey,
		bridgehubtypes.StoreKey,
	}
}

func AetherCoreSystemModuleNames() []string {
	return slices.Clone(aetherCoreSystemModules)
}

func AetherCoreSystemStoreKeys() []string {
	return []string{
		constitutiontypes.StoreKey,
		systemregistrytypes.StoreKey,
		validatorelectiontypes.StoreKey,
		validatorregistrytypes.StoreKey,
		configtypes.StoreKey,
	}
}

func aetherCorePreBlockerOrder() []string {
	return []string{upgradetypes.ModuleName, authtypes.ModuleName}
}

func aetherCoreBeginBlockerOrder() []string {
	return []string{
		minttypes.ModuleName,
		distrtypes.ModuleName,
		protocolpooltypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		epochstypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorregistrytypes.ModuleName,
		configtypes.ModuleName,
		aethercoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
	}
}

func aetherCoreEndBlockerOrder() []string {
	return []string{
		banktypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		protocolpooltypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorregistrytypes.ModuleName,
		configtypes.ModuleName,
		aethercoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
	}
}

func aetherCoreInitGenesisOrder() []string {
	return []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		epochstypes.ModuleName,
		protocolpooltypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorregistrytypes.ModuleName,
		configtypes.ModuleName,
		aethercoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
		feestypes.ModuleName,
		tokenfactorytypes.ModuleName,
		dextypes.ModuleName,
	}
}

func aetherCoreExportGenesisOrder() []string {
	return []string{
		consensusparamtypes.ModuleName,
		authtypes.ModuleName,
		protocolpooltypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		epochstypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorregistrytypes.ModuleName,
		configtypes.ModuleName,
		aethercoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
		feestypes.ModuleName,
		tokenfactorytypes.ModuleName,
		dextypes.ModuleName,
	}
}

func (app *L1App) ValidateAetherCoreWiringGate() error {
	if app == nil || app.ModuleManager == nil {
		return fmt.Errorf("aether core wiring gate requires initialized app")
	}
	if AetherCoreRoutingExecutionPoint() != RoutingExecutionPointAnteAdmissionOnly {
		return fmt.Errorf("unsupported routing execution point %s", AetherCoreRoutingExecutionPoint())
	}
	for _, moduleName := range AetherCorePrototypeModuleNames() {
		if _, found := app.ModuleManager.Modules[moduleName]; !found {
			return fmt.Errorf("prototype module %s is not registered", moduleName)
		}
		if _, found := app.keys[moduleName]; !found {
			return fmt.Errorf("prototype module %s store key is not mounted", moduleName)
		}
		if _, found := maccPerms[moduleName]; found {
			return fmt.Errorf("prototype module %s must not have module account permissions", moduleName)
		}
	}
	for _, moduleName := range AetherCoreSystemModuleNames() {
		if _, found := app.ModuleManager.Modules[moduleName]; !found {
			return fmt.Errorf("system module %s is not registered", moduleName)
		}
		if _, found := app.keys[moduleName]; !found {
			return fmt.Errorf("system module %s store key is not mounted", moduleName)
		}
		if _, found := maccPerms[moduleName]; found {
			return fmt.Errorf("system module %s must not have module account permissions", moduleName)
		}
	}
	return nil
}
