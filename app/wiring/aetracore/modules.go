package aetracore

import (
	"slices"

	actorregistrytypes "github.com/sovereign-l1/l1/x/actor-registry/types"
	aetraeconomicstypes "github.com/sovereign-l1/l1/x/aetra-economics/types"
	aetrastakingpolicytypes "github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
	aetravalidatorscoretypes "github.com/sovereign-l1/l1/x/aetra-validator-score/types"
	aetracoretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	avmschedulertypes "github.com/sovereign-l1/l1/x/avm-scheduler/types"
	bridgehubtypes "github.com/sovereign-l1/l1/x/bridge-hub/types"
	configvotingtypes "github.com/sovereign-l1/l1/x/config-voting/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	constitutiontypes "github.com/sovereign-l1/l1/x/constitution/types"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	crosschainregistrytypes "github.com/sovereign-l1/l1/x/cross-chain-registry/types"
	nativeevidencetypes "github.com/sovereign-l1/l1/x/evidence/types"
	identityroottypes "github.com/sovereign-l1/l1/x/identity-root/types"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
	reportertypes "github.com/sovereign-l1/l1/x/reporter/types"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
	shardingcoordinatortypes "github.com/sovereign-l1/l1/x/sharding-coordinator/types"
	singlenominatorpooltypes "github.com/sovereign-l1/l1/x/single-nominator-pool/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

type RoutingExecutionPoint string

const (
	// Routing remains an admission/ante-level executable spec until a coordinated
	// upgrade adds public Msg services and production persistence semantics.
	RoutingExecutionPointAnteAdmissionOnly RoutingExecutionPoint = "ANTE_ADMISSION_ONLY"
)

var prototypeModules = []string{
	aetracoretypes.ModuleName,
	loadtypes.ModuleName,
	routingtypes.ModuleName,
	zonestypes.ModuleName,
	meshtypes.ModuleName,
	networkingtypes.ModuleName,
	paymentstypes.ModuleName,
	configvotingtypes.ModuleName,
	schedulertypes.ModuleName,
	avmschedulertypes.ModuleName,
	actorregistrytypes.ModuleName,
	contractstypes.ModuleName,
	storagerenttypes.ModuleName,
	identityroottypes.ModuleName,
	bridgehubtypes.ModuleName,
	crosschainregistrytypes.ModuleName,
	shardingcoordinatortypes.ModuleName,
}

var systemModules = []string{
	constitutiontypes.ModuleName,
	systemregistrytypes.ModuleName,
	nativeevidencetypes.ModuleName,
	reportertypes.ModuleName,
	nominatorpooltypes.ModuleName,
	singlenominatorpooltypes.ModuleName,
	validatorelectiontypes.ModuleName,
	validatorinsurancetypes.ModuleName,
	validatorregistrytypes.ModuleName,
	aetrastakingpolicytypes.ModuleName,
	aetraeconomicstypes.ModuleName,
	aetravalidatorscoretypes.ModuleName,
	configtypes.ModuleName,
	nativeaccounttypes.ModuleName,
}

func RoutingExecution() RoutingExecutionPoint {
	return RoutingExecutionPointAnteAdmissionOnly
}

func PrototypeModuleNames() []string {
	return slices.Clone(prototypeModules)
}

func PrototypeStoreKeys() []string {
	return []string{
		aetracoretypes.StoreKey,
		loadtypes.StoreKey,
		routingtypes.StoreKey,
		zonestypes.StoreKey,
		meshtypes.StoreKey,
		networkingtypes.StoreKey,
		paymentstypes.StoreKey,
		configvotingtypes.StoreKey,
		schedulertypes.StoreKey,
		avmschedulertypes.StoreKey,
		actorregistrytypes.StoreKey,
		contractstypes.StoreKey,
		storagerenttypes.StoreKey,
		identityroottypes.StoreKey,
		bridgehubtypes.StoreKey,
		crosschainregistrytypes.StoreKey,
		shardingcoordinatortypes.StoreKey,
	}
}

func SystemModuleNames() []string {
	return slices.Clone(systemModules)
}

func SystemStoreKeys() []string {
	return []string{
		constitutiontypes.StoreKey,
		systemregistrytypes.StoreKey,
		nativeevidencetypes.StoreKey,
		reportertypes.StoreKey,
		nominatorpooltypes.StoreKey,
		singlenominatorpooltypes.StoreKey,
		validatorelectiontypes.StoreKey,
		validatorinsurancetypes.StoreKey,
		validatorregistrytypes.StoreKey,
		aetrastakingpolicytypes.StoreKey,
		aetraeconomicstypes.StoreKey,
		aetravalidatorscoretypes.StoreKey,
		configtypes.StoreKey,
		nativeaccounttypes.StoreKey,
	}
}
