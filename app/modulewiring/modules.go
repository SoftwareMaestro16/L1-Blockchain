package modulewiring

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	coregenesis "cosmossdk.io/core/genesis"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/epochs"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/protocolpool"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"

	"github.com/sovereign-l1/l1/app/stakingpolicy"
	actorregistrymodule "github.com/sovereign-l1/l1/x/actor-registry"
	actorregistrykeeper "github.com/sovereign-l1/l1/x/actor-registry/keeper"
	aetraeconomicsmodule "github.com/sovereign-l1/l1/x/aetra-economics"
	aetraeconomicskeeper "github.com/sovereign-l1/l1/x/aetra-economics/keeper"
	aetrastakingpolicymodule "github.com/sovereign-l1/l1/x/aetra-staking-policy"
	aetrastakingpolicykeeper "github.com/sovereign-l1/l1/x/aetra-staking-policy/keeper"
	aetravalidatorscoremodule "github.com/sovereign-l1/l1/x/aetra-validator-score"
	aetravalidatorscorekeeper "github.com/sovereign-l1/l1/x/aetra-validator-score/keeper"
	aetracoremodule "github.com/sovereign-l1/l1/x/aetracore"
	aetracorekeeper "github.com/sovereign-l1/l1/x/aetracore/keeper"
	avmschedulermodule "github.com/sovereign-l1/l1/x/avm-scheduler"
	avmschedulerkeeper "github.com/sovereign-l1/l1/x/avm-scheduler/keeper"
	bridgehubmodule "github.com/sovereign-l1/l1/x/bridge-hub"
	bridgehubkeeper "github.com/sovereign-l1/l1/x/bridge-hub/keeper"
	burnmodule "github.com/sovereign-l1/l1/x/burn"
	burnkeeper "github.com/sovereign-l1/l1/x/burn/keeper"
	configmodule "github.com/sovereign-l1/l1/x/config"
	configvotingmodule "github.com/sovereign-l1/l1/x/config-voting"
	configvotingkeeper "github.com/sovereign-l1/l1/x/config-voting/keeper"
	configkeeper "github.com/sovereign-l1/l1/x/config/keeper"
	constitutionmodule "github.com/sovereign-l1/l1/x/constitution"
	constitutionkeeper "github.com/sovereign-l1/l1/x/constitution/keeper"
	contractsmodule "github.com/sovereign-l1/l1/x/contracts"
	contractskeeper "github.com/sovereign-l1/l1/x/contracts/keeper"
	crosschainregistrymodule "github.com/sovereign-l1/l1/x/cross-chain-registry"
	crosschainregistrykeeper "github.com/sovereign-l1/l1/x/cross-chain-registry/keeper"
	delegatorprotectionmodule "github.com/sovereign-l1/l1/x/delegator-protection"
	delegatorprotectionkeeper "github.com/sovereign-l1/l1/x/delegator-protection/keeper"
	dynamiccommissionmodule "github.com/sovereign-l1/l1/x/dynamic-commission"
	dynamiccommissionkeeper "github.com/sovereign-l1/l1/x/dynamic-commission/keeper"
	emissionsmodule "github.com/sovereign-l1/l1/x/emissions"
	emissionskeeper "github.com/sovereign-l1/l1/x/emissions/keeper"
	nativeevidencemodule "github.com/sovereign-l1/l1/x/evidence"
	nativeevidencekeeper "github.com/sovereign-l1/l1/x/evidence/keeper"
	feecollectormodule "github.com/sovereign-l1/l1/x/fee-collector"
	feecollectorkeeper "github.com/sovereign-l1/l1/x/fee-collector/keeper"
	feesmodule "github.com/sovereign-l1/l1/x/fees"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	identityrootmodule "github.com/sovereign-l1/l1/x/identity-root"
	identityrootkeeper "github.com/sovereign-l1/l1/x/identity-root/keeper"
	loadmodule "github.com/sovereign-l1/l1/x/load"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	meshmodule "github.com/sovereign-l1/l1/x/mesh"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	mintauthoritymodule "github.com/sovereign-l1/l1/x/mint-authority"
	mintauthoritykeeper "github.com/sovereign-l1/l1/x/mint-authority/keeper"
	nativeaccountmodule "github.com/sovereign-l1/l1/x/native-account"
	nativeaccountkeeper "github.com/sovereign-l1/l1/x/native-account/keeper"
	networkingmodule "github.com/sovereign-l1/l1/x/networking"
	networkingkeeper "github.com/sovereign-l1/l1/x/networking/keeper"
	nominatorpoolmodule "github.com/sovereign-l1/l1/x/nominator-pool"
	nominatorpoolkeeper "github.com/sovereign-l1/l1/x/nominator-pool/keeper"
	paymentsmodule "github.com/sovereign-l1/l1/x/payments"
	paymentskeeper "github.com/sovereign-l1/l1/x/payments/keeper"
	performancemodule "github.com/sovereign-l1/l1/x/performance"
	performancekeeper "github.com/sovereign-l1/l1/x/performance/keeper"
	reportermodule "github.com/sovereign-l1/l1/x/reporter"
	reporterkeeper "github.com/sovereign-l1/l1/x/reporter/keeper"
	reputationmodule "github.com/sovereign-l1/l1/x/reputation"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
	routingmodule "github.com/sovereign-l1/l1/x/routing"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	schedulermodule "github.com/sovereign-l1/l1/x/scheduler"
	schedulerkeeper "github.com/sovereign-l1/l1/x/scheduler/keeper"
	shardingcoordinatormodule "github.com/sovereign-l1/l1/x/sharding-coordinator"
	shardingcoordinatorkeeper "github.com/sovereign-l1/l1/x/sharding-coordinator/keeper"
	singlenominatorpoolmodule "github.com/sovereign-l1/l1/x/single-nominator-pool"
	singlenominatorpoolkeeper "github.com/sovereign-l1/l1/x/single-nominator-pool/keeper"
	stakeconcentrationmodule "github.com/sovereign-l1/l1/x/stake-concentration"
	stakeconcentrationkeeper "github.com/sovereign-l1/l1/x/stake-concentration/keeper"
	storagerentmodule "github.com/sovereign-l1/l1/x/storage-rent"
	storagerentkeeper "github.com/sovereign-l1/l1/x/storage-rent/keeper"
	systemregistrymodule "github.com/sovereign-l1/l1/x/system-registry"
	systemregistrykeeper "github.com/sovereign-l1/l1/x/system-registry/keeper"
	treasurymodule "github.com/sovereign-l1/l1/x/treasury"
	treasurykeeper "github.com/sovereign-l1/l1/x/treasury/keeper"
	validatorelectionmodule "github.com/sovereign-l1/l1/x/validator-election"
	validatorelectionkeeper "github.com/sovereign-l1/l1/x/validator-election/keeper"
	validatorinsurancemodule "github.com/sovereign-l1/l1/x/validator-insurance"
	validatorinsurancekeeper "github.com/sovereign-l1/l1/x/validator-insurance/keeper"
	validatorregistrymodule "github.com/sovereign-l1/l1/x/validator-registry"
	validatorregistrykeeper "github.com/sovereign-l1/l1/x/validator-registry/keeper"
	zonesmodule "github.com/sovereign-l1/l1/x/zones"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
)

type ModuleDeps struct {
	AppCodec		codec.Codec
	TxConfig		client.TxConfig
	DeliverTx		coregenesis.TxHandler
	InterfaceRegistry	codectypes.InterfaceRegistry

	AccountKeeper		authkeeper.AccountKeeper
	BankKeeper		bankkeeper.BaseKeeper
	StakingKeeper		*stakingkeeper.Keeper
	SlashingKeeper		slashingkeeper.Keeper
	MintKeeper		mintkeeper.Keeper
	DistrKeeper		distrkeeper.Keeper
	GovKeeper		*govkeeper.Keeper
	UpgradeKeeper		*upgradekeeper.Keeper
	EvidenceKeeper		evidencekeeper.Keeper
	ConsensusParamsKeeper	consensusparamkeeper.Keeper
	FeeGrantKeeper		feegrantkeeper.Keeper
	AuthzKeeper		authzkeeper.Keeper
	EpochsKeeper		*epochskeeper.Keeper
	ProtocolPoolKeeper	protocolpoolkeeper.Keeper

	ConfigKeeper			*configkeeper.Keeper
	ConfigVotingKeeper		*configvotingkeeper.Keeper
	ConstitutionKeeper		*constitutionkeeper.Keeper
	BurnKeeper			burnkeeper.Keeper
	TreasuryKeeper			treasurykeeper.Keeper
	EmissionsKeeper			emissionskeeper.Keeper
	MintAuthorityKeeper		mintauthoritykeeper.Keeper
	DelegatorProtectionKeeper	delegatorprotectionkeeper.Keeper
	ReputationKeeper		reputationkeeper.Keeper
	PerformanceKeeper		performancekeeper.Keeper
	DynamicCommissionKeeper		dynamiccommissionkeeper.Keeper
	StakeConcentrationKeeper	stakeconcentrationkeeper.Keeper
	FeeCollectorKeeper		feecollectorkeeper.Keeper
	FeesKeeper			feeskeeper.Keeper
	AetraCoreKeeper			*aetracorekeeper.Keeper
	LoadKeeper			*loadkeeper.Keeper
	RoutingKeeper			*routingkeeper.Keeper
	ZonesKeeper			*zoneskeeper.Keeper
	MeshKeeper			*meshkeeper.Keeper
	NetworkingKeeper		*networkingkeeper.Keeper
	NativeAccountKeeper		*nativeaccountkeeper.Keeper
	PaymentsKeeper			*paymentskeeper.Keeper
	SchedulerKeeper			*schedulerkeeper.Keeper
	AVMSchedulerKeeper		*avmschedulerkeeper.Keeper
	ActorRegistryKeeper		*actorregistrykeeper.Keeper
	ContractsKeeper			*contractskeeper.Keeper
	StorageRentKeeper		*storagerentkeeper.Keeper
	IdentityRootKeeper		*identityrootkeeper.Keeper
	BridgeHubKeeper			*bridgehubkeeper.Keeper
	CrossChainRegistryKeeper	*crosschainregistrykeeper.Keeper
	ShardingCoordinatorKeeper	*shardingcoordinatorkeeper.Keeper
	SystemRegistryKeeper		*systemregistrykeeper.Keeper
	NativeEvidenceKeeper		*nativeevidencekeeper.Keeper
	ReporterKeeper			*reporterkeeper.Keeper
	NominatorPoolKeeper		*nominatorpoolkeeper.Keeper
	SingleNominatorPoolKeeper	*singlenominatorpoolkeeper.Keeper
	ValidatorElectionKeeper		*validatorelectionkeeper.Keeper
	ValidatorInsuranceKeeper	*validatorinsurancekeeper.Keeper
	ValidatorRegistryKeeper		*validatorregistrykeeper.Keeper
	AetraStakingPolicyKeeper	*aetrastakingpolicykeeper.Keeper
	AetraEconomicsKeeper		*aetraeconomicskeeper.Keeper
	AetraValidatorScoreKeeper	*aetravalidatorscorekeeper.Keeper
}

func NewModuleManager(deps ModuleDeps) *module.Manager {
	return module.NewManager(
		genutil.NewAppModule(deps.AccountKeeper, deps.StakingKeeper, deps.DeliverTx, deps.TxConfig),
		auth.NewAppModule(deps.AppCodec, deps.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		vesting.NewAppModule(deps.AccountKeeper, deps.BankKeeper),
		bank.NewAppModule(deps.AppCodec, deps.BankKeeper, deps.AccountKeeper, nil),
		feegrantmodule.NewAppModule(deps.AppCodec, deps.AccountKeeper, deps.BankKeeper, deps.FeeGrantKeeper, deps.InterfaceRegistry),
		gov.NewAppModule(deps.AppCodec, deps.GovKeeper, deps.AccountKeeper, deps.BankKeeper, nil),
		mint.NewAppModule(deps.AppCodec, deps.MintKeeper, deps.AccountKeeper, nil, nil),
		slashing.NewAppModule(deps.AppCodec, deps.SlashingKeeper, deps.AccountKeeper, deps.BankKeeper, deps.StakingKeeper, nil, deps.InterfaceRegistry),
		distr.NewAppModule(deps.AppCodec, deps.DistrKeeper, deps.AccountKeeper, deps.BankKeeper, deps.StakingKeeper, nil),
		stakingpolicy.NewAppModule(deps.AppCodec, deps.StakingKeeper, deps.AccountKeeper, deps.BankKeeper, nil),
		upgrade.NewAppModule(deps.UpgradeKeeper, deps.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(deps.EvidenceKeeper),
		authzmodule.NewAppModule(deps.AppCodec, deps.AuthzKeeper, deps.AccountKeeper, deps.BankKeeper, deps.InterfaceRegistry),
		consensus.NewAppModule(deps.AppCodec, deps.ConsensusParamsKeeper),
		epochs.NewAppModule(deps.EpochsKeeper),
		protocolpool.NewAppModule(deps.ProtocolPoolKeeper, deps.AccountKeeper, deps.BankKeeper),
		constitutionmodule.NewAppModule(deps.ConstitutionKeeper),
		systemregistrymodule.NewAppModule(deps.SystemRegistryKeeper),
		nativeevidencemodule.NewAppModule(deps.NativeEvidenceKeeper),
		reportermodule.NewAppModule(deps.ReporterKeeper),
		nominatorpoolmodule.NewAppModule(deps.NominatorPoolKeeper),
		singlenominatorpoolmodule.NewAppModule(deps.SingleNominatorPoolKeeper),
		validatorelectionmodule.NewAppModule(deps.ValidatorElectionKeeper),
		validatorinsurancemodule.NewAppModule(deps.ValidatorInsuranceKeeper),
		validatorregistrymodule.NewAppModule(deps.ValidatorRegistryKeeper),
		aetrastakingpolicymodule.NewAppModule(deps.AetraStakingPolicyKeeper),
		aetraeconomicsmodule.NewAppModule(deps.AetraEconomicsKeeper),
		aetravalidatorscoremodule.NewAppModule(deps.AetraValidatorScoreKeeper),
		configmodule.NewAppModule(deps.ConfigKeeper),
		configvotingmodule.NewAppModule(deps.ConfigVotingKeeper),
		aetracoremodule.NewAppModule(deps.AetraCoreKeeper),
		loadmodule.NewAppModule(deps.LoadKeeper),
		routingmodule.NewAppModule(deps.RoutingKeeper),
		zonesmodule.NewAppModule(deps.ZonesKeeper),
		meshmodule.NewAppModule(deps.MeshKeeper),
		networkingmodule.NewAppModule(deps.NetworkingKeeper),
		nativeaccountmodule.NewAppModule(*deps.NativeAccountKeeper),
		paymentsmodule.NewAppModule(deps.PaymentsKeeper),
		schedulermodule.NewAppModule(deps.SchedulerKeeper),
		avmschedulermodule.NewAppModule(deps.AVMSchedulerKeeper),
		actorregistrymodule.NewAppModule(deps.ActorRegistryKeeper),
		contractsmodule.NewAppModule(deps.ContractsKeeper),
		storagerentmodule.NewAppModule(deps.StorageRentKeeper),
		identityrootmodule.NewAppModule(deps.IdentityRootKeeper),
		bridgehubmodule.NewAppModule(deps.BridgeHubKeeper),
		crosschainregistrymodule.NewAppModule(deps.CrossChainRegistryKeeper),
		shardingcoordinatormodule.NewAppModule(deps.ShardingCoordinatorKeeper),
		burnmodule.NewAppModule(deps.AppCodec, deps.BurnKeeper),
		treasurymodule.NewAppModule(deps.AppCodec, deps.TreasuryKeeper),
		emissionsmodule.NewAppModule(deps.AppCodec, deps.EmissionsKeeper),
		mintauthoritymodule.NewAppModule(deps.AppCodec, deps.MintAuthorityKeeper),
		delegatorprotectionmodule.NewAppModule(deps.AppCodec, deps.DelegatorProtectionKeeper),
		reputationmodule.NewAppModule(deps.AppCodec, deps.ReputationKeeper),
		performancemodule.NewAppModule(deps.AppCodec, deps.PerformanceKeeper),
		dynamiccommissionmodule.NewAppModule(deps.AppCodec, deps.DynamicCommissionKeeper),
		stakeconcentrationmodule.NewAppModule(deps.AppCodec, deps.StakeConcentrationKeeper),
		feecollectormodule.NewAppModule(deps.AppCodec, deps.FeeCollectorKeeper),
		feesmodule.NewAppModule(deps.AppCodec, deps.FeesKeeper),
	)
}

func NewBasicManager(manager *module.Manager, legacyAmino *codec.LegacyAmino, interfaceRegistry codectypes.InterfaceRegistry) module.BasicManager {
	basicManager := module.NewBasicManagerFromManager(
		manager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName:	genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName:		gov.NewAppModuleBasic([]govclient.ProposalHandler{}),
		},
	)
	basicManager.RegisterLegacyAminoCodec(legacyAmino)
	basicManager.RegisterInterfaces(interfaceRegistry)
	return basicManager
}

func RegisterModuleServices(
	appCodec codec.Codec,
	msgRouter *baseapp.MsgServiceRouter,
	queryRouter *baseapp.GRPCQueryRouter,
	manager *module.Manager,
) module.Configurator {
	configurator := module.NewConfigurator(appCodec, msgRouter, queryRouter)
	if err := manager.RegisterServices(configurator); err != nil {
		panic(err)
	}
	return configurator
}

func RegisterRuntimeQueryServices(queryRouter *baseapp.GRPCQueryRouter, manager *module.Manager) {
	autocliv1.RegisterQueryServer(queryRouter, runtimeservices.NewAutoCLIQueryService(manager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(queryRouter, reflectionSvc)
	testdata_pulsar.RegisterQueryServer(queryRouter, testdata_pulsar.QueryImpl{})

	_ = runtime.EventService{}
}

func NewSimulationManager(appCodec codec.Codec, manager *module.Manager, accountKeeper authkeeper.AccountKeeper) *module.SimulationManager {
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(appCodec, accountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	simulationManager := module.NewSimulationManagerFromAppModules(manager.Modules, overrideModules)
	simulationManager.RegisterStoreDecoders()
	return simulationManager
}
