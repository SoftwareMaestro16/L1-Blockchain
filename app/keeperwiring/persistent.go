package keeperwiring

import (
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	actorregistrykeeper "github.com/sovereign-l1/l1/x/actor-registry/keeper"
	actorregistrytypes "github.com/sovereign-l1/l1/x/actor-registry/types"
	aetracorekeeper "github.com/sovereign-l1/l1/x/aetracore/keeper"
	aetracoretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	avmschedulerkeeper "github.com/sovereign-l1/l1/x/avm-scheduler/keeper"
	avmschedulertypes "github.com/sovereign-l1/l1/x/avm-scheduler/types"
	bridgehubkeeper "github.com/sovereign-l1/l1/x/bridge-hub/keeper"
	bridgehubtypes "github.com/sovereign-l1/l1/x/bridge-hub/types"
	configvotingkeeper "github.com/sovereign-l1/l1/x/config-voting/keeper"
	configvotingtypes "github.com/sovereign-l1/l1/x/config-voting/types"
	configkeeper "github.com/sovereign-l1/l1/x/config/keeper"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	constitutionkeeper "github.com/sovereign-l1/l1/x/constitution/keeper"
	constitutiontypes "github.com/sovereign-l1/l1/x/constitution/types"
	contractskeeper "github.com/sovereign-l1/l1/x/contracts/keeper"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	crosschainregistrykeeper "github.com/sovereign-l1/l1/x/cross-chain-registry/keeper"
	crosschainregistrytypes "github.com/sovereign-l1/l1/x/cross-chain-registry/types"
	nativeevidencekeeper "github.com/sovereign-l1/l1/x/evidence/keeper"
	nativeevidencetypes "github.com/sovereign-l1/l1/x/evidence/types"
	identityrootkeeper "github.com/sovereign-l1/l1/x/identity-root/keeper"
	identityroottypes "github.com/sovereign-l1/l1/x/identity-root/types"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	nativeaccountkeeper "github.com/sovereign-l1/l1/x/native-account/keeper"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	networkingkeeper "github.com/sovereign-l1/l1/x/networking/keeper"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
	nominatorpoolkeeper "github.com/sovereign-l1/l1/x/nominator-pool/keeper"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	paymentskeeper "github.com/sovereign-l1/l1/x/payments/keeper"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
	reporterkeeper "github.com/sovereign-l1/l1/x/reporter/keeper"
	reportertypes "github.com/sovereign-l1/l1/x/reporter/types"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	schedulerkeeper "github.com/sovereign-l1/l1/x/scheduler/keeper"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
	shardingcoordinatorkeeper "github.com/sovereign-l1/l1/x/sharding-coordinator/keeper"
	shardingcoordinatortypes "github.com/sovereign-l1/l1/x/sharding-coordinator/types"
	singlenominatorpoolkeeper "github.com/sovereign-l1/l1/x/single-nominator-pool/keeper"
	singlenominatorpooltypes "github.com/sovereign-l1/l1/x/single-nominator-pool/types"
	storagerentkeeper "github.com/sovereign-l1/l1/x/storage-rent/keeper"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrykeeper "github.com/sovereign-l1/l1/x/system-registry/keeper"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	validatorelectionkeeper "github.com/sovereign-l1/l1/x/validator-election/keeper"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancekeeper "github.com/sovereign-l1/l1/x/validator-insurance/keeper"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrykeeper "github.com/sovereign-l1/l1/x/validator-registry/keeper"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

type PersistentKeepers struct {
	ConstitutionKeeper		constitutionkeeper.Keeper
	ConfigKeeper			configkeeper.Keeper
	ConfigVotingKeeper		configvotingkeeper.Keeper
	SystemRegistryKeeper		systemregistrykeeper.Keeper
	NativeEvidenceKeeper		nativeevidencekeeper.Keeper
	ReporterKeeper			reporterkeeper.Keeper
	NominatorPoolKeeper		nominatorpoolkeeper.Keeper
	SingleNominatorPoolKeeper	singlenominatorpoolkeeper.Keeper
	ValidatorElectionKeeper		validatorelectionkeeper.Keeper
	ValidatorInsuranceKeeper	validatorinsurancekeeper.Keeper
	ValidatorRegistryKeeper		validatorregistrykeeper.Keeper
	AetraCoreKeeper			aetracorekeeper.Keeper
	LoadKeeper			loadkeeper.Keeper
	RoutingKeeper			routingkeeper.Keeper
	ZonesKeeper			zoneskeeper.Keeper
	MeshKeeper			meshkeeper.Keeper
	NetworkingKeeper		networkingkeeper.Keeper
	NativeAccountKeeper		nativeaccountkeeper.Keeper
	PaymentsKeeper			paymentskeeper.Keeper
	SchedulerKeeper			schedulerkeeper.Keeper
	AVMSchedulerKeeper		avmschedulerkeeper.Keeper
	ActorRegistryKeeper		actorregistrykeeper.Keeper
	ContractsKeeper			contractskeeper.Keeper
	StorageRentKeeper		storagerentkeeper.Keeper
	IdentityRootKeeper		identityrootkeeper.Keeper
	BridgeHubKeeper			bridgehubkeeper.Keeper
	CrossChainRegistryKeeper	crosschainregistrykeeper.Keeper
	ShardingCoordinatorKeeper	shardingcoordinatorkeeper.Keeper
}

func NewPersistentKeepers(keys map[string]*storetypes.KVStoreKey, bankKeeper storagerentkeeper.BankKeeper) PersistentKeepers {
	nativeAccountKeeper := nativeaccountkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[nativeaccounttypes.StoreKey]))
	storageRentKeeper := storagerentkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[storagerenttypes.StoreKey])).WithBankKeeper(bankKeeper)

	contractsKeeper := contractskeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[contractstypes.StoreKey]))
	contractsKeeper = contractsKeeper.WithAccountStatusReader(nativeAccountKeeper).WithBankKeeper(bankKeeper).WithStorageRentRateProvider(storageRentKeeper)

	return PersistentKeepers{
		ConstitutionKeeper:		constitutionkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[constitutiontypes.StoreKey])),
		ConfigKeeper:			configkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[configtypes.StoreKey])),
		ConfigVotingKeeper:		configvotingkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[configvotingtypes.StoreKey])),
		SystemRegistryKeeper:		systemregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[systemregistrytypes.StoreKey])),
		NativeEvidenceKeeper:		nativeevidencekeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[nativeevidencetypes.StoreKey])),
		ReporterKeeper:			reporterkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[reportertypes.StoreKey])),
		NominatorPoolKeeper:		nominatorpoolkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[nominatorpooltypes.StoreKey])),
		SingleNominatorPoolKeeper:	singlenominatorpoolkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[singlenominatorpooltypes.StoreKey])),
		ValidatorElectionKeeper:	validatorelectionkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[validatorelectiontypes.StoreKey])),
		ValidatorInsuranceKeeper:	validatorinsurancekeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[validatorinsurancetypes.StoreKey])),
		ValidatorRegistryKeeper:	validatorregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[validatorregistrytypes.StoreKey])),
		AetraCoreKeeper:		aetracorekeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[aetracoretypes.StoreKey])),
		LoadKeeper:			loadkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[loadtypes.StoreKey])),
		RoutingKeeper:			routingkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[routingtypes.StoreKey])),
		ZonesKeeper:			zoneskeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[zonestypes.StoreKey])),
		MeshKeeper:			meshkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[meshtypes.StoreKey])),
		NetworkingKeeper:		networkingkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[networkingtypes.StoreKey])),
		NativeAccountKeeper:		nativeAccountKeeper,
		PaymentsKeeper:			paymentskeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[paymentstypes.StoreKey])),
		SchedulerKeeper:		schedulerkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[schedulertypes.StoreKey])),
		AVMSchedulerKeeper:		avmschedulerkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[avmschedulertypes.StoreKey])),
		ActorRegistryKeeper:		actorregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[actorregistrytypes.StoreKey])),
		ContractsKeeper:		contractsKeeper,
		StorageRentKeeper:		storageRentKeeper,
		IdentityRootKeeper:		identityrootkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[identityroottypes.StoreKey])),
		BridgeHubKeeper:		bridgehubkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[bridgehubtypes.StoreKey])),
		CrossChainRegistryKeeper:	crosschainregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[crosschainregistrytypes.StoreKey])),
		ShardingCoordinatorKeeper:	shardingcoordinatorkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[shardingcoordinatortypes.StoreKey])),
	}
}
