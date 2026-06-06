package app

import (
	"fmt"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
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
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/tx/signing"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
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

const appName = appparams.ChainName

const (
	AccountAddressPrefix   = "ae"
	ValidatorAddressPrefix = "aevaloper"
	ConsensusAddressPrefix = "aevalcons"
	BondDenom              = appparams.BaseDenom
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:                     nil,
		distrtypes.ModuleName:                          nil,
		minttypes.ModuleName:                           {authtypes.Minter},
		stakingtypes.BondedPoolName:                    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:                 {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                            {authtypes.Burner},
		protocolpooltypes.ModuleName:                   nil,
		protocolpooltypes.ProtocolPoolEscrowAccount:    nil,
		tokenfactorytypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
		dextypes.ModuleName:                            {authtypes.Minter, authtypes.Burner},
		burntypes.ModuleName:                           {authtypes.Burner},
		feecollectortypes.CollectorModuleName:          {authtypes.Burner},
		feecollectortypes.TreasuryModuleName:           nil,
		feecollectortypes.ProtectionModuleName:         nil,
		feecollectortypes.ValidatorInsuranceModuleName: nil,
		feecollectortypes.EcosystemGrantsModuleName:    nil,
		feecollectortypes.StorageRentReserveModuleName: nil,
		feecollectortypes.BurnModuleName:               nil,
		feecollectortypes.ReporterRewardsModuleName:    nil,
		feestypes.ModuleName:                           nil,
	}
)

var (
	_ runtime.AppI            = (*L1App)(nil)
	_ servertypes.Application = (*L1App)(nil)
)

// L1App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type L1App struct {
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys map[string]*storetypes.KVStoreKey

	// essential keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	// supplementary keepers
	FeeGrantKeeper            feegrantkeeper.Keeper
	AuthzKeeper               authzkeeper.Keeper
	EpochsKeeper              *epochskeeper.Keeper
	ProtocolPoolKeeper        protocolpoolkeeper.Keeper
	ConfigKeeper              configkeeper.Keeper
	ConfigVotingKeeper        configvotingkeeper.Keeper
	ConstitutionKeeper        constitutionkeeper.Keeper
	TokenFactoryKeeper        tokenfactorykeeper.Keeper
	DexKeeper                 dexkeeper.Keeper
	BurnKeeper                burnkeeper.Keeper
	TreasuryKeeper            treasurykeeper.Keeper
	EmissionsKeeper           emissionskeeper.Keeper
	DynamicCommissionKeeper   dynamiccommissionkeeper.Keeper
	StakeConcentrationKeeper  stakeconcentrationkeeper.Keeper
	FeeCollectorKeeper        feecollectorkeeper.Keeper
	FeesKeeper                feeskeeper.Keeper
	AetherCoreKeeper          aethercorekeeper.Keeper
	LoadKeeper                loadkeeper.Keeper
	RoutingKeeper             routingkeeper.Keeper
	ZonesKeeper               zoneskeeper.Keeper
	MeshKeeper                meshkeeper.Keeper
	NetworkingKeeper          networkingkeeper.Keeper
	PaymentsKeeper            paymentskeeper.Keeper
	SchedulerKeeper           schedulerkeeper.Keeper
	AVMSchedulerKeeper        avmschedulerkeeper.Keeper
	ActorRegistryKeeper       actorregistrykeeper.Keeper
	StorageRentKeeper         storagerentkeeper.Keeper
	IdentityRootKeeper        identityrootkeeper.Keeper
	BridgeHubKeeper           bridgehubkeeper.Keeper
	CrossChainRegistryKeeper  crosschainregistrykeeper.Keeper
	ShardingCoordinatorKeeper shardingcoordinatorkeeper.Keeper
	SystemRegistryKeeper      systemregistrykeeper.Keeper
	NativeEvidenceKeeper      nativeevidencekeeper.Keeper
	ReporterKeeper            reporterkeeper.Keeper
	NominatorPoolKeeper       nominatorpoolkeeper.Keeper
	SingleNominatorPoolKeeper singlenominatorpoolkeeper.Keeper
	ValidatorElectionKeeper   validatorelectionkeeper.Keeper
	ValidatorInsuranceKeeper  validatorinsurancekeeper.Keeper
	ValidatorRegistryKeeper   validatorregistrykeeper.Keeper

	// the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".aetheris")
	if err != nil {
		panic(err)
	}
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(AccountAddressPrefix, AccountAddressPrefix+"pub")
	cfg.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorAddressPrefix+"pub")
	cfg.SetBech32PrefixForConsensusNode(ConsensusAddressPrefix, ConsensusAddressPrefix+"pub")
	sdk.DefaultBondDenom = BondDenom
}

// NewL1App returns a reference to an initialized L1App.
func NewL1App(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *L1App {
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          aetherisaddress.Codec{},
			ValidatorAddressCodec: aetherisaddress.Codec{},
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	if err := interfaceRegistry.SigningContext().Validate(); err != nil {
		panic(err)
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	// Below we could construct and set an application specific mempool and
	// ABCI 1.0 PrepareProposal and ProcessProposal handlers. These defaults are
	// already set in the SDK's BaseApp, this shows an example of how to override
	// them.
	//
	// Example:
	//
	// bApp := baseapp.NewBaseApp(...)
	// nonceMempool := mempool.NewSenderNonceMempool()
	// abciPropHandler := NewDefaultProposalHandler(nonceMempool, bApp)
	//
	// bApp.SetMempool(nonceMempool)
	// bApp.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// bApp.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
	//
	// Alternatively, you can construct BaseApp options, append those to
	// baseAppOptions and pass them to NewBaseApp.
	//
	// Example:
	//
	// prepareOpt = func(app *baseapp.BaseApp) {
	// 	abciPropHandler := baseapp.NewDefaultProposalHandler(nonceMempool, app)
	// 	app.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// }
	// baseAppOptions = append(baseAppOptions, prepareOpt)

	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution())

	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		consensusparamtypes.StoreKey,
		upgradetypes.StoreKey,
		feegrant.StoreKey,
		evidencetypes.StoreKey,
		authzkeeper.StoreKey,
		epochstypes.StoreKey,
		protocolpooltypes.StoreKey,
		constitutiontypes.StoreKey,
		configtypes.StoreKey,
		configvotingtypes.StoreKey,
		systemregistrytypes.StoreKey,
		nativeevidencetypes.StoreKey,
		reportertypes.StoreKey,
		nominatorpooltypes.StoreKey,
		singlenominatorpooltypes.StoreKey,
		validatorelectiontypes.StoreKey,
		validatorinsurancetypes.StoreKey,
		validatorregistrytypes.StoreKey,
		tokenfactorytypes.StoreKey,
		dextypes.StoreKey,
		burntypes.StoreKey,
		treasurytypes.StoreKey,
		emissionstypes.StoreKey,
		dynamiccommissiontypes.StoreKey,
		stakeconcentrationtypes.StoreKey,
		feecollectortypes.StoreKey,
		feestypes.StoreKey,
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
		crosschainregistrytypes.StoreKey,
		shardingcoordinatortypes.StoreKey,
	)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	app := &L1App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
	}

	txConfig = app.initKeepers(appCodec, legacyAmino, logger, appOpts, keys)
	app.initModules(appCodec, legacyAmino, interfaceRegistry, txConfig)
	if err := app.ValidateAetherCoreWiringGate(); err != nil {
		panic(err)
	}

	// initialize stores
	app.MountKVStores(keys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(txConfig)

	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default postHandlers chain
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	app.setPostHandler()

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}
