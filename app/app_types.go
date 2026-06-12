package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"

	actorregistrykeeper "github.com/sovereign-l1/l1/x/actor-registry/keeper"
	aetraeconomicskeeper "github.com/sovereign-l1/l1/x/aetra-economics/keeper"
	aetrastakingpolicykeeper "github.com/sovereign-l1/l1/x/aetra-staking-policy/keeper"
	aetravalidatorscorekeeper "github.com/sovereign-l1/l1/x/aetra-validator-score/keeper"
	aetracorekeeper "github.com/sovereign-l1/l1/x/aetracore/keeper"
	avmschedulerkeeper "github.com/sovereign-l1/l1/x/avm-scheduler/keeper"
	bridgehubkeeper "github.com/sovereign-l1/l1/x/bridge-hub/keeper"
	burnkeeper "github.com/sovereign-l1/l1/x/burn/keeper"
	configvotingkeeper "github.com/sovereign-l1/l1/x/config-voting/keeper"
	configkeeper "github.com/sovereign-l1/l1/x/config/keeper"
	constitutionkeeper "github.com/sovereign-l1/l1/x/constitution/keeper"
	contractskeeper "github.com/sovereign-l1/l1/x/contracts/keeper"
	crosschainregistrykeeper "github.com/sovereign-l1/l1/x/cross-chain-registry/keeper"
	delegatorprotectionkeeper "github.com/sovereign-l1/l1/x/delegator-protection/keeper"
	dynamiccommissionkeeper "github.com/sovereign-l1/l1/x/dynamic-commission/keeper"
	emissionskeeper "github.com/sovereign-l1/l1/x/emissions/keeper"
	nativeevidencekeeper "github.com/sovereign-l1/l1/x/evidence/keeper"
	feecollectorkeeper "github.com/sovereign-l1/l1/x/fee-collector/keeper"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	identityrootkeeper "github.com/sovereign-l1/l1/x/identity-root/keeper"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	mintauthoritykeeper "github.com/sovereign-l1/l1/x/mint-authority/keeper"
	nativeaccountkeeper "github.com/sovereign-l1/l1/x/native-account/keeper"
	networkingkeeper "github.com/sovereign-l1/l1/x/networking/keeper"
	nominatorpoolkeeper "github.com/sovereign-l1/l1/x/nominator-pool/keeper"
	paymentskeeper "github.com/sovereign-l1/l1/x/payments/keeper"
	performancekeeper "github.com/sovereign-l1/l1/x/performance/keeper"
	reporterkeeper "github.com/sovereign-l1/l1/x/reporter/keeper"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	schedulerkeeper "github.com/sovereign-l1/l1/x/scheduler/keeper"
	shardingcoordinatorkeeper "github.com/sovereign-l1/l1/x/sharding-coordinator/keeper"
	singlenominatorpoolkeeper "github.com/sovereign-l1/l1/x/single-nominator-pool/keeper"
	stakeconcentrationkeeper "github.com/sovereign-l1/l1/x/stake-concentration/keeper"
	storagerentkeeper "github.com/sovereign-l1/l1/x/storage-rent/keeper"
	systemregistrykeeper "github.com/sovereign-l1/l1/x/system-registry/keeper"
	treasurykeeper "github.com/sovereign-l1/l1/x/treasury/keeper"
	validatorelectionkeeper "github.com/sovereign-l1/l1/x/validator-election/keeper"
	validatorinsurancekeeper "github.com/sovereign-l1/l1/x/validator-insurance/keeper"
	validatorregistrykeeper "github.com/sovereign-l1/l1/x/validator-registry/keeper"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
)

var (
	_	runtime.AppI		= (*L1App)(nil)
	_	servertypes.Application	= (*L1App)(nil)
)

// L1App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type L1App struct {
	*baseapp.BaseApp
	legacyAmino		*codec.LegacyAmino
	appCodec		codec.Codec
	txConfig		client.TxConfig
	interfaceRegistry	codectypes.InterfaceRegistry

	// keys to access the substores
	keys	map[string]*storetypes.KVStoreKey

	// essential keepers
	AccountKeeper		authkeeper.AccountKeeper
	BankKeeper		bankkeeper.BaseKeeper
	StakingKeeper		*stakingkeeper.Keeper
	SlashingKeeper		slashingkeeper.Keeper
	MintKeeper		mintkeeper.Keeper
	DistrKeeper		distrkeeper.Keeper
	GovKeeper		govkeeper.Keeper
	UpgradeKeeper		*upgradekeeper.Keeper
	EvidenceKeeper		evidencekeeper.Keeper
	ConsensusParamsKeeper	consensusparamkeeper.Keeper

	// supplementary keepers
	FeeGrantKeeper			feegrantkeeper.Keeper
	AuthzKeeper			authzkeeper.Keeper
	EpochsKeeper			*epochskeeper.Keeper
	ProtocolPoolKeeper		protocolpoolkeeper.Keeper
	ConfigKeeper			configkeeper.Keeper
	ConfigVotingKeeper		configvotingkeeper.Keeper
	ConstitutionKeeper		constitutionkeeper.Keeper
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
	SystemRegistryKeeper		systemregistrykeeper.Keeper
	NativeEvidenceKeeper		nativeevidencekeeper.Keeper
	ReporterKeeper			reporterkeeper.Keeper
	NominatorPoolKeeper		nominatorpoolkeeper.Keeper
	SingleNominatorPoolKeeper	singlenominatorpoolkeeper.Keeper
	ValidatorElectionKeeper		validatorelectionkeeper.Keeper
	ValidatorInsuranceKeeper	validatorinsurancekeeper.Keeper
	ValidatorRegistryKeeper		validatorregistrykeeper.Keeper
	AetraStakingPolicyKeeper	aetrastakingpolicykeeper.Keeper
	AetraEconomicsKeeper		aetraeconomicskeeper.Keeper
	AetraValidatorScoreKeeper	aetravalidatorscorekeeper.Keeper

	// the module manager
	ModuleManager		*module.Manager
	BasicModuleManager	module.BasicManager

	// simulation manager
	sm	*module.SimulationManager

	// invariant registry
	invariantRegistry	*AppInvariantRouteRegistry

	// module configurator
	configurator	module.Configurator
}
