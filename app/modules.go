package app

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/epochs"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/protocolpool"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	actorregistrymodule "github.com/sovereign-l1/l1/x/actor-registry"
	aetracoremodule "github.com/sovereign-l1/l1/x/aetracore"
	avmschedulermodule "github.com/sovereign-l1/l1/x/avm-scheduler"
	bridgehubmodule "github.com/sovereign-l1/l1/x/bridge-hub"
	burnmodule "github.com/sovereign-l1/l1/x/burn"
	configmodule "github.com/sovereign-l1/l1/x/config"
	configvotingmodule "github.com/sovereign-l1/l1/x/config-voting"
	constitutionmodule "github.com/sovereign-l1/l1/x/constitution"
	crosschainregistrymodule "github.com/sovereign-l1/l1/x/cross-chain-registry"
	delegatorprotectionmodule "github.com/sovereign-l1/l1/x/delegator-protection"
	dynamiccommissionmodule "github.com/sovereign-l1/l1/x/dynamic-commission"
	emissionsmodule "github.com/sovereign-l1/l1/x/emissions"
	nativeevidencemodule "github.com/sovereign-l1/l1/x/evidence"
	feecollectormodule "github.com/sovereign-l1/l1/x/fee-collector"
	feesmodule "github.com/sovereign-l1/l1/x/fees"
	identityrootmodule "github.com/sovereign-l1/l1/x/identity-root"
	loadmodule "github.com/sovereign-l1/l1/x/load"
	meshmodule "github.com/sovereign-l1/l1/x/mesh"
	mintauthoritymodule "github.com/sovereign-l1/l1/x/mint-authority"
	networkingmodule "github.com/sovereign-l1/l1/x/networking"
	nominatorpoolmodule "github.com/sovereign-l1/l1/x/nominator-pool"
	paymentsmodule "github.com/sovereign-l1/l1/x/payments"
	performancemodule "github.com/sovereign-l1/l1/x/performance"
	reportermodule "github.com/sovereign-l1/l1/x/reporter"
	reputationmodule "github.com/sovereign-l1/l1/x/reputation"
	routingmodule "github.com/sovereign-l1/l1/x/routing"
	schedulermodule "github.com/sovereign-l1/l1/x/scheduler"
	shardingcoordinatormodule "github.com/sovereign-l1/l1/x/sharding-coordinator"
	singlenominatorpoolmodule "github.com/sovereign-l1/l1/x/single-nominator-pool"
	stakeconcentrationmodule "github.com/sovereign-l1/l1/x/stake-concentration"
	storagerentmodule "github.com/sovereign-l1/l1/x/storage-rent"
	systemregistrymodule "github.com/sovereign-l1/l1/x/system-registry"
	treasurymodule "github.com/sovereign-l1/l1/x/treasury"
	validatorelectionmodule "github.com/sovereign-l1/l1/x/validator-election"
	validatorinsurancemodule "github.com/sovereign-l1/l1/x/validator-insurance"
	validatorregistrymodule "github.com/sovereign-l1/l1/x/validator-registry"
	zonesmodule "github.com/sovereign-l1/l1/x/zones"
)

func (app *L1App) initModules(
	appCodec codec.Codec,
	legacyAmino *codec.LegacyAmino,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfig client.TxConfig,
) {
	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app, txConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, nil),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, nil),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil, app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, nil),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		epochs.NewAppModule(app.EpochsKeeper),
		protocolpool.NewAppModule(app.ProtocolPoolKeeper, app.AccountKeeper, app.BankKeeper),
		constitutionmodule.NewAppModule(&app.ConstitutionKeeper),
		systemregistrymodule.NewAppModule(&app.SystemRegistryKeeper),
		nativeevidencemodule.NewAppModule(&app.NativeEvidenceKeeper),
		reportermodule.NewAppModule(&app.ReporterKeeper),
		nominatorpoolmodule.NewAppModule(&app.NominatorPoolKeeper),
		singlenominatorpoolmodule.NewAppModule(&app.SingleNominatorPoolKeeper),
		validatorelectionmodule.NewAppModule(&app.ValidatorElectionKeeper),
		validatorinsurancemodule.NewAppModule(&app.ValidatorInsuranceKeeper),
		validatorregistrymodule.NewAppModule(&app.ValidatorRegistryKeeper),
		configmodule.NewAppModule(&app.ConfigKeeper),
		configvotingmodule.NewAppModule(&app.ConfigVotingKeeper),
		aetracoremodule.NewAppModule(&app.AetraCoreKeeper),
		loadmodule.NewAppModule(&app.LoadKeeper),
		routingmodule.NewAppModule(&app.RoutingKeeper),
		zonesmodule.NewAppModule(&app.ZonesKeeper),
		meshmodule.NewAppModule(&app.MeshKeeper),
		networkingmodule.NewAppModule(&app.NetworkingKeeper),
		paymentsmodule.NewAppModule(&app.PaymentsKeeper),
		schedulermodule.NewAppModule(&app.SchedulerKeeper),
		avmschedulermodule.NewAppModule(&app.AVMSchedulerKeeper),
		actorregistrymodule.NewAppModule(&app.ActorRegistryKeeper),
		storagerentmodule.NewAppModule(&app.StorageRentKeeper),
		identityrootmodule.NewAppModule(&app.IdentityRootKeeper),
		bridgehubmodule.NewAppModule(&app.BridgeHubKeeper),
		crosschainregistrymodule.NewAppModule(&app.CrossChainRegistryKeeper),
		shardingcoordinatormodule.NewAppModule(&app.ShardingCoordinatorKeeper),
		burnmodule.NewAppModule(appCodec, app.BurnKeeper),
		treasurymodule.NewAppModule(appCodec, app.TreasuryKeeper),
		emissionsmodule.NewAppModule(appCodec, app.EmissionsKeeper),
		mintauthoritymodule.NewAppModule(appCodec, app.MintAuthorityKeeper),
		delegatorprotectionmodule.NewAppModule(appCodec, app.DelegatorProtectionKeeper),
		reputationmodule.NewAppModule(appCodec, app.ReputationKeeper),
		performancemodule.NewAppModule(appCodec, app.PerformanceKeeper),
		dynamiccommissionmodule.NewAppModule(appCodec, app.DynamicCommissionKeeper),
		stakeconcentrationmodule.NewAppModule(appCodec, app.StakeConcentrationKeeper),
		feecollectormodule.NewAppModule(appCodec, app.FeeCollectorKeeper),
		feesmodule.NewAppModule(appCodec, app.FeesKeeper),
	)

	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName:     gov.NewAppModuleBasic([]govclient.ProposalHandler{}),
		},
	)
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	app.ModuleManager.SetOrderPreBlockers(aetherCorePreBlockerOrder()...)
	app.ModuleManager.SetOrderBeginBlockers(aetherCoreBeginBlockerOrder()...)
	app.ModuleManager.SetOrderEndBlockers(aetherCoreEndBlockerOrder()...)
	app.ModuleManager.SetOrderInitGenesis(aetherCoreInitGenesisOrder()...)
	app.ModuleManager.SetOrderExportGenesis(aetherCoreExportGenesisOrder()...)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	if err := app.ModuleManager.RegisterServices(app.configurator); err != nil {
		panic(err)
	}

	app.RegisterUpgradeHandlers()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)
	testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)
	app.sm.RegisterStoreDecoders()

	_ = runtime.EventService{}
}
