package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/sovereign-l1/l1/app/modulewiring"
)

func (app *L1App) initModules(
	appCodec codec.Codec,
	legacyAmino *codec.LegacyAmino,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfig client.TxConfig,
) {
	app.ModuleManager = modulewiring.NewModuleManager(modulewiring.ModuleDeps{
		AppCodec:		appCodec,
		TxConfig:		txConfig,
		DeliverTx:		app,
		InterfaceRegistry:	interfaceRegistry,

		AccountKeeper:		app.AccountKeeper,
		BankKeeper:		app.BankKeeper,
		StakingKeeper:		app.StakingKeeper,
		SlashingKeeper:		app.SlashingKeeper,
		MintKeeper:		app.MintKeeper,
		DistrKeeper:		app.DistrKeeper,
		GovKeeper:		&app.GovKeeper,
		UpgradeKeeper:		app.UpgradeKeeper,
		EvidenceKeeper:		app.EvidenceKeeper,
		ConsensusParamsKeeper:	app.ConsensusParamsKeeper,
		FeeGrantKeeper:		app.FeeGrantKeeper,
		AuthzKeeper:		app.AuthzKeeper,
		EpochsKeeper:		app.EpochsKeeper,
		ProtocolPoolKeeper:	app.ProtocolPoolKeeper,

		ConfigKeeper:			&app.ConfigKeeper,
		ConfigVotingKeeper:		&app.ConfigVotingKeeper,
		ConstitutionKeeper:		&app.ConstitutionKeeper,
		BurnKeeper:			app.BurnKeeper,
		TreasuryKeeper:			app.TreasuryKeeper,
		EmissionsKeeper:		app.EmissionsKeeper,
		MintAuthorityKeeper:		app.MintAuthorityKeeper,
		DelegatorProtectionKeeper:	app.DelegatorProtectionKeeper,
		ReputationKeeper:		app.ReputationKeeper,
		PerformanceKeeper:		app.PerformanceKeeper,
		DynamicCommissionKeeper:	app.DynamicCommissionKeeper,
		StakeConcentrationKeeper:	app.StakeConcentrationKeeper,
		FeeCollectorKeeper:		app.FeeCollectorKeeper,
		FeesKeeper:			app.FeesKeeper,
		AetraCoreKeeper:		&app.AetraCoreKeeper,
		LoadKeeper:			&app.LoadKeeper,
		RoutingKeeper:			&app.RoutingKeeper,
		ZonesKeeper:			&app.ZonesKeeper,
		MeshKeeper:			&app.MeshKeeper,
		NetworkingKeeper:		&app.NetworkingKeeper,
		NativeAccountKeeper:		&app.NativeAccountKeeper,
		PaymentsKeeper:			&app.PaymentsKeeper,
		SchedulerKeeper:		&app.SchedulerKeeper,
		AVMSchedulerKeeper:		&app.AVMSchedulerKeeper,
		ActorRegistryKeeper:		&app.ActorRegistryKeeper,
		ContractsKeeper:		&app.ContractsKeeper,
		StorageRentKeeper:		&app.StorageRentKeeper,
		IdentityRootKeeper:		&app.IdentityRootKeeper,
		BridgeHubKeeper:		&app.BridgeHubKeeper,
		CrossChainRegistryKeeper:	&app.CrossChainRegistryKeeper,
		ShardingCoordinatorKeeper:	&app.ShardingCoordinatorKeeper,
		SystemRegistryKeeper:		&app.SystemRegistryKeeper,
		NativeEvidenceKeeper:		&app.NativeEvidenceKeeper,
		ReporterKeeper:			&app.ReporterKeeper,
		NominatorPoolKeeper:		&app.NominatorPoolKeeper,
		SingleNominatorPoolKeeper:	&app.SingleNominatorPoolKeeper,
		ValidatorElectionKeeper:	&app.ValidatorElectionKeeper,
		ValidatorInsuranceKeeper:	&app.ValidatorInsuranceKeeper,
		ValidatorRegistryKeeper:	&app.ValidatorRegistryKeeper,
		AetraStakingPolicyKeeper:	&app.AetraStakingPolicyKeeper,
		AetraEconomicsKeeper:		&app.AetraEconomicsKeeper,
		AetraValidatorScoreKeeper:	&app.AetraValidatorScoreKeeper,
	})
	app.BasicModuleManager = modulewiring.NewBasicManager(app.ModuleManager, legacyAmino, interfaceRegistry)
	app.ModuleManager.SetOrderPreBlockers(aetherCorePreBlockerOrder()...)
	app.ModuleManager.SetOrderBeginBlockers(aetherCoreBeginBlockerOrder()...)
	app.ModuleManager.SetOrderEndBlockers(aetherCoreEndBlockerOrder()...)
	app.ModuleManager.SetOrderInitGenesis(aetherCoreInitGenesisOrder()...)
	app.ModuleManager.SetOrderExportGenesis(aetherCoreExportGenesisOrder()...)

	app.configurator = modulewiring.RegisterModuleServices(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter(), app.ModuleManager)
	app.RegisterUpgradeHandlers()
	modulewiring.RegisterRuntimeQueryServices(app.GRPCQueryRouter(), app.ModuleManager)
	app.sm = modulewiring.NewSimulationManager(app.appCodec, app.ModuleManager, app.AccountKeeper)
}
