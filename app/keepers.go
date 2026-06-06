package app

import (
	"cosmossdk.io/log/v2"
	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/app/keeperconfig"
	actorregistrykeeper "github.com/sovereign-l1/l1/x/actor-registry/keeper"
	actorregistrytypes "github.com/sovereign-l1/l1/x/actor-registry/types"
	aethercorekeeper "github.com/sovereign-l1/l1/x/aethercore/keeper"
	aethercoretypes "github.com/sovereign-l1/l1/x/aethercore/types"
	avmschedulerkeeper "github.com/sovereign-l1/l1/x/avm-scheduler/keeper"
	avmschedulertypes "github.com/sovereign-l1/l1/x/avm-scheduler/types"
	bridgehubkeeper "github.com/sovereign-l1/l1/x/bridge-hub/keeper"
	bridgehubtypes "github.com/sovereign-l1/l1/x/bridge-hub/types"
	burnkeeper "github.com/sovereign-l1/l1/x/burn/keeper"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	configvotingkeeper "github.com/sovereign-l1/l1/x/config-voting/keeper"
	configvotingtypes "github.com/sovereign-l1/l1/x/config-voting/types"
	configkeeper "github.com/sovereign-l1/l1/x/config/keeper"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	constitutionkeeper "github.com/sovereign-l1/l1/x/constitution/keeper"
	constitutiontypes "github.com/sovereign-l1/l1/x/constitution/types"
	crosschainregistrykeeper "github.com/sovereign-l1/l1/x/cross-chain-registry/keeper"
	crosschainregistrytypes "github.com/sovereign-l1/l1/x/cross-chain-registry/types"
	delegatorprotectionkeeper "github.com/sovereign-l1/l1/x/delegator-protection/keeper"
	delegatorprotectiontypes "github.com/sovereign-l1/l1/x/delegator-protection/types"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	dynamiccommissionkeeper "github.com/sovereign-l1/l1/x/dynamic-commission/keeper"
	dynamiccommissiontypes "github.com/sovereign-l1/l1/x/dynamic-commission/types"
	emissionskeeper "github.com/sovereign-l1/l1/x/emissions/keeper"
	emissionstypes "github.com/sovereign-l1/l1/x/emissions/types"
	nativeevidencekeeper "github.com/sovereign-l1/l1/x/evidence/keeper"
	nativeevidencetypes "github.com/sovereign-l1/l1/x/evidence/types"
	feecollectorkeeper "github.com/sovereign-l1/l1/x/fee-collector/keeper"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	identityrootkeeper "github.com/sovereign-l1/l1/x/identity-root/keeper"
	identityroottypes "github.com/sovereign-l1/l1/x/identity-root/types"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	mintauthoritykeeper "github.com/sovereign-l1/l1/x/mint-authority/keeper"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	networkingkeeper "github.com/sovereign-l1/l1/x/networking/keeper"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
	nominatorpoolkeeper "github.com/sovereign-l1/l1/x/nominator-pool/keeper"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	paymentskeeper "github.com/sovereign-l1/l1/x/payments/keeper"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
	performancekeeper "github.com/sovereign-l1/l1/x/performance/keeper"
	performancetypes "github.com/sovereign-l1/l1/x/performance/types"
	reporterkeeper "github.com/sovereign-l1/l1/x/reporter/keeper"
	reportertypes "github.com/sovereign-l1/l1/x/reporter/types"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	schedulerkeeper "github.com/sovereign-l1/l1/x/scheduler/keeper"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
	shardingcoordinatorkeeper "github.com/sovereign-l1/l1/x/sharding-coordinator/keeper"
	shardingcoordinatortypes "github.com/sovereign-l1/l1/x/sharding-coordinator/types"
	singlenominatorpoolkeeper "github.com/sovereign-l1/l1/x/single-nominator-pool/keeper"
	singlenominatorpooltypes "github.com/sovereign-l1/l1/x/single-nominator-pool/types"
	stakeconcentrationkeeper "github.com/sovereign-l1/l1/x/stake-concentration/keeper"
	stakeconcentrationtypes "github.com/sovereign-l1/l1/x/stake-concentration/types"
	storagerentkeeper "github.com/sovereign-l1/l1/x/storage-rent/keeper"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrykeeper "github.com/sovereign-l1/l1/x/system-registry/keeper"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
	treasurykeeper "github.com/sovereign-l1/l1/x/treasury/keeper"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
	validatorelectionkeeper "github.com/sovereign-l1/l1/x/validator-election/keeper"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancekeeper "github.com/sovereign-l1/l1/x/validator-insurance/keeper"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrykeeper "github.com/sovereign-l1/l1/x/validator-registry/keeper"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func (app *L1App) initKeepers(
	appCodec codec.Codec,
	legacyAmino *codec.LegacyAmino,
	logger log.Logger,
	appOpts servertypes.AppOptions,
	keys map[string]*storetypes.KVStoreKey,
) client.TxConfig {
	govAuthority := aetherisaddress.FormatAccAddress(authtypes.NewModuleAddress(govtypes.ModuleName))
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		govAuthority,
		runtime.EventService{},
	)
	app.BaseApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		GetMaccPerms(),
		aetherisaddress.Codec{},
		AccountAddressPrefix,
		govAuthority,
		authkeeper.WithUnorderedTransactions(true),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		govAuthority,
		logger,
	)

	txConfig := keeperconfig.NewTxConfig(appCodec, app.BankKeeper)
	app.txConfig = txConfig

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		govAuthority,
		aetherisaddress.Codec{},
		aetherisaddress.Codec{},
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		govAuthority,
	)
	app.ProtocolPoolKeeper = protocolpoolkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[protocolpooltypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		govAuthority,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		govAuthority,
		distrkeeper.WithExternalCommunityPool(app.ProtocolPoolKeeper),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		app.StakingKeeper,
		govAuthority,
	)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feegrant.StoreKey]),
		app.AccountKeeper,
	)
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		keeperconfig.UnsafeSkipUpgradeHeights(appOpts),
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		app.BaseApp,
		govAuthority,
	)

	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler)
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		govAuthority,
		govkeeper.NewDefaultCalculateVoteResultsAndVotingPower(app.StakingKeeper),
	)
	govKeeper.SetLegacyRouter(govRouter)
	app.GovKeeper = *govKeeper.SetHooks(govtypes.NewMultiGovHooks())

	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	app.EvidenceKeeper = *evidenceKeeper

	epochsKeeper := epochskeeper.NewKeeper(runtime.NewKVStoreService(keys[epochstypes.StoreKey]), appCodec)
	app.EpochsKeeper = &epochsKeeper
	app.EpochsKeeper.SetHooks(epochstypes.NewMultiEpochHooks())
	app.ConstitutionKeeper = constitutionkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[constitutiontypes.StoreKey]))
	app.ConfigKeeper = configkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[configtypes.StoreKey]))
	app.ConfigVotingKeeper = configvotingkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[configvotingtypes.StoreKey]))
	app.SystemRegistryKeeper = systemregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[systemregistrytypes.StoreKey]))
	app.NativeEvidenceKeeper = nativeevidencekeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[nativeevidencetypes.StoreKey]))
	app.ReporterKeeper = reporterkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[reportertypes.StoreKey]))
	app.NominatorPoolKeeper = nominatorpoolkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[nominatorpooltypes.StoreKey]))
	app.SingleNominatorPoolKeeper = singlenominatorpoolkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[singlenominatorpooltypes.StoreKey]))
	app.ValidatorElectionKeeper = validatorelectionkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[validatorelectiontypes.StoreKey]))
	app.ValidatorInsuranceKeeper = validatorinsurancekeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[validatorinsurancetypes.StoreKey]))
	app.ValidatorRegistryKeeper = validatorregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[validatorregistrytypes.StoreKey]))
	app.TokenFactoryKeeper = tokenfactorykeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[tokenfactorytypes.StoreKey]),
		app.BankKeeper,
		govAuthority,
	)
	app.DexKeeper = dexkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[dextypes.StoreKey]),
		app.BankKeeper,
		govAuthority,
	)
	app.BurnKeeper = burnkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[burntypes.StoreKey]),
		app.BankKeeper,
		govAuthority,
	)
	app.TreasuryKeeper = treasurykeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[treasurytypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		govAuthority,
	)
	app.EmissionsKeeper = emissionskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[emissionstypes.StoreKey]),
		govAuthority,
	)
	app.MintAuthorityKeeper = mintauthoritykeeper.NewKeeper(
		runtime.NewKVStoreService(keys[mintauthoritytypes.StoreKey]),
		app.BankKeeper,
		govAuthority,
	)
	app.DelegatorProtectionKeeper = delegatorprotectionkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[delegatorprotectiontypes.StoreKey]),
		govAuthority,
	)
	app.ReputationKeeper = reputationkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[reputationtypes.StoreKey]),
		govAuthority,
	)
	app.PerformanceKeeper = performancekeeper.NewKeeper(
		runtime.NewKVStoreService(keys[performancetypes.StoreKey]),
		govAuthority,
	)
	app.DynamicCommissionKeeper = dynamiccommissionkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[dynamiccommissiontypes.StoreKey]),
		govAuthority,
	)
	app.StakeConcentrationKeeper = stakeconcentrationkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakeconcentrationtypes.StoreKey]),
		govAuthority,
	)
	app.FeeCollectorKeeper = feecollectorkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feecollectortypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		govAuthority,
	)
	app.FeesKeeper = feeskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feestypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		govAuthority,
	)
	app.AetherCoreKeeper = aethercorekeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[aethercoretypes.StoreKey]))
	app.LoadKeeper = loadkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[loadtypes.StoreKey]))
	app.RoutingKeeper = routingkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[routingtypes.StoreKey]))
	app.ZonesKeeper = zoneskeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[zonestypes.StoreKey]))
	app.MeshKeeper = meshkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[meshtypes.StoreKey]))
	app.NetworkingKeeper = networkingkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[networkingtypes.StoreKey]))
	app.PaymentsKeeper = paymentskeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[paymentstypes.StoreKey]))
	app.SchedulerKeeper = schedulerkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[schedulertypes.StoreKey]))
	app.AVMSchedulerKeeper = avmschedulerkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[avmschedulertypes.StoreKey]))
	app.ActorRegistryKeeper = actorregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[actorregistrytypes.StoreKey]))
	app.StorageRentKeeper = storagerentkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[storagerenttypes.StoreKey]))
	app.IdentityRootKeeper = identityrootkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[identityroottypes.StoreKey]))
	app.BridgeHubKeeper = bridgehubkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[bridgehubtypes.StoreKey]))
	app.CrossChainRegistryKeeper = crosschainregistrykeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[crosschainregistrytypes.StoreKey]))
	app.ShardingCoordinatorKeeper = shardingcoordinatorkeeper.NewPersistentKeeper(runtime.NewKVStoreService(keys[shardingcoordinatortypes.StoreKey]))
	return txConfig
}
