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
	aethercoremodule "github.com/sovereign-l1/l1/x/aethercore"
	avmschedulermodule "github.com/sovereign-l1/l1/x/avm-scheduler"
	bridgehubmodule "github.com/sovereign-l1/l1/x/bridge-hub"
	configmodule "github.com/sovereign-l1/l1/x/config"
	constitutionmodule "github.com/sovereign-l1/l1/x/constitution"
	dexmodule "github.com/sovereign-l1/l1/x/dex"
	feesmodule "github.com/sovereign-l1/l1/x/fees"
	identityrootmodule "github.com/sovereign-l1/l1/x/identity-root"
	loadmodule "github.com/sovereign-l1/l1/x/load"
	meshmodule "github.com/sovereign-l1/l1/x/mesh"
	networkingmodule "github.com/sovereign-l1/l1/x/networking"
	paymentsmodule "github.com/sovereign-l1/l1/x/payments"
	routingmodule "github.com/sovereign-l1/l1/x/routing"
	schedulermodule "github.com/sovereign-l1/l1/x/scheduler"
	storagerentmodule "github.com/sovereign-l1/l1/x/storage-rent"
	systemregistrymodule "github.com/sovereign-l1/l1/x/system-registry"
	tokenfactorymodule "github.com/sovereign-l1/l1/x/tokenfactory"
	validatorelectionmodule "github.com/sovereign-l1/l1/x/validator-election"
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
		validatorelectionmodule.NewAppModule(&app.ValidatorElectionKeeper),
		validatorregistrymodule.NewAppModule(&app.ValidatorRegistryKeeper),
		configmodule.NewAppModule(&app.ConfigKeeper),
		aethercoremodule.NewAppModule(app.AetherCoreKeeper),
		loadmodule.NewAppModule(app.LoadKeeper),
		routingmodule.NewAppModule(app.RoutingKeeper),
		zonesmodule.NewAppModule(app.ZonesKeeper),
		meshmodule.NewAppModule(app.MeshKeeper),
		networkingmodule.NewAppModule(app.NetworkingKeeper),
		paymentsmodule.NewAppModule(app.PaymentsKeeper),
		schedulermodule.NewAppModule(&app.SchedulerKeeper),
		avmschedulermodule.NewAppModule(&app.AVMSchedulerKeeper),
		actorregistrymodule.NewAppModule(&app.ActorRegistryKeeper),
		storagerentmodule.NewAppModule(&app.StorageRentKeeper),
		identityrootmodule.NewAppModule(&app.IdentityRootKeeper),
		bridgehubmodule.NewAppModule(&app.BridgeHubKeeper),
		tokenfactorymodule.NewAppModule(appCodec, app.TokenFactoryKeeper),
		dexmodule.NewAppModule(appCodec, app.DexKeeper),
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
