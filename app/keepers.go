package app

import (
	"context"

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

	"github.com/sovereign-l1/l1/app/accounts"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/app/keeperconfig"
	"github.com/sovereign-l1/l1/app/keeperwiring"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
)

// validatorRegistryReputationAdapter wraps the reputation keeper for validator-registry usage.
type validatorRegistryReputationAdapter struct {
	Keeper reputationkeeper.Keeper
}

func (a validatorRegistryReputationAdapter) GetValidatorTotalScore(ctx context.Context, addr string) (uint32, bool, error) {
	vs, err := a.Keeper.GetValidatorReputation(ctx, addr)
	if err != nil {
		return 0, false, err
	}
	if vs == nil {
		return 0, false, nil
	}
	return vs.TotalScore, vs.IsJailed || vs.IsSlashed, nil
}

func (app *L1App) initKeepers(
	appCodec codec.Codec,
	legacyAmino *codec.LegacyAmino,
	logger log.Logger,
	appOpts servertypes.AppOptions,
	keys map[string]*storetypes.KVStoreKey,
) client.TxConfig {
	govAuthority := aetraaddress.FormatAccAddress(authtypes.NewModuleAddress(govtypes.ModuleName))
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
		accounts.ModuleAccountPermissions(),
		aetraaddress.Codec{},
		SDKBech32AccountPrefix,
		govAuthority,
		authkeeper.WithUnorderedTransactions(true),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		accounts.BlockedAddresses(),
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
		aetraaddress.Codec{},
		aetraaddress.Codec{},
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

	persistentKeepers := keeperwiring.NewPersistentKeepers(keys, app.BankKeeper)
	app.ConstitutionKeeper = persistentKeepers.ConstitutionKeeper
	app.ConfigKeeper = persistentKeepers.ConfigKeeper
	app.ConfigVotingKeeper = persistentKeepers.ConfigVotingKeeper
	app.SystemRegistryKeeper = persistentKeepers.SystemRegistryKeeper
	app.NativeEvidenceKeeper = persistentKeepers.NativeEvidenceKeeper
	app.ReporterKeeper = persistentKeepers.ReporterKeeper
	app.NominatorPoolKeeper = persistentKeepers.NominatorPoolKeeper
	app.SingleNominatorPoolKeeper = persistentKeepers.SingleNominatorPoolKeeper
	app.ValidatorElectionKeeper = persistentKeepers.ValidatorElectionKeeper
	app.ValidatorInsuranceKeeper = persistentKeepers.ValidatorInsuranceKeeper
	app.ValidatorRegistryKeeper = persistentKeepers.ValidatorRegistryKeeper
	app.AetraCoreKeeper = persistentKeepers.AetraCoreKeeper
	app.LoadKeeper = persistentKeepers.LoadKeeper
	app.RoutingKeeper = persistentKeepers.RoutingKeeper
	app.ZonesKeeper = persistentKeepers.ZonesKeeper
	app.MeshKeeper = persistentKeepers.MeshKeeper
	app.NetworkingKeeper = persistentKeepers.NetworkingKeeper
	app.NativeAccountKeeper = persistentKeepers.NativeAccountKeeper
	app.PaymentsKeeper = persistentKeepers.PaymentsKeeper
	app.SchedulerKeeper = persistentKeepers.SchedulerKeeper
	app.AVMSchedulerKeeper = persistentKeepers.AVMSchedulerKeeper
	app.ActorRegistryKeeper = persistentKeepers.ActorRegistryKeeper
	app.ContractsKeeper = persistentKeepers.ContractsKeeper
	app.StorageRentKeeper = persistentKeepers.StorageRentKeeper
	app.IdentityRootKeeper = persistentKeepers.IdentityRootKeeper
	app.BridgeHubKeeper = persistentKeepers.BridgeHubKeeper
	app.CrossChainRegistryKeeper = persistentKeepers.CrossChainRegistryKeeper
	app.ShardingCoordinatorKeeper = persistentKeepers.ShardingCoordinatorKeeper

	nativeKeepers := keeperwiring.NewNativeKeepers(keeperwiring.NativeKeeperDeps{
		AppCodec:	appCodec,
		Keys:		keys,
		AccountKeeper:	app.AccountKeeper,
		BankKeeper:	app.BankKeeper,
		DistrKeeper:	app.DistrKeeper,
		GovAuthority:	govAuthority,
	})
	app.BurnKeeper = nativeKeepers.BurnKeeper
	app.TreasuryKeeper = nativeKeepers.TreasuryKeeper
	app.EmissionsKeeper = nativeKeepers.EmissionsKeeper
	app.MintAuthorityKeeper = nativeKeepers.MintAuthorityKeeper
	app.DelegatorProtectionKeeper = nativeKeepers.DelegatorProtectionKeeper
	app.ReputationKeeper = nativeKeepers.ReputationKeeper
	app.ValidatorRegistryKeeper = app.ValidatorRegistryKeeper.WithReputationKeeper(
		validatorRegistryReputationAdapter{Keeper: app.ReputationKeeper},
	)
	app.PerformanceKeeper = nativeKeepers.PerformanceKeeper
	app.DynamicCommissionKeeper = nativeKeepers.DynamicCommissionKeeper
	app.StakeConcentrationKeeper = nativeKeepers.StakeConcentrationKeeper
	app.FeeCollectorKeeper = nativeKeepers.FeeCollectorKeeper
	app.FeesKeeper = nativeKeepers.FeesKeeper
	app.AetraStakingPolicyKeeper = nativeKeepers.AetraStakingPolicyKeeper
	app.AetraEconomicsKeeper = nativeKeepers.AetraEconomicsKeeper
	app.AetraValidatorScoreKeeper = nativeKeepers.AetraValidatorScoreKeeper
	return txConfig
}
