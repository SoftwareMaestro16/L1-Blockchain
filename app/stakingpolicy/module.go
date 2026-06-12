package stakingpolicy

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const consensusVersion uint64 = 5

var (
	_	module.AppModuleBasic		= AppModuleBasic{}
	_	module.AppModuleSimulation	= AppModule{}
	_	module.HasServices		= AppModule{}
	_	module.HasABCIGenesis		= AppModule{}
	_	module.HasABCIEndBlock		= AppModule{}
	_	appmodule.AppModule		= AppModule{}
	_	appmodule.HasBeginBlocker	= AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string	{ return stakingtypes.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	stakingtypes.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	stakingtypes.RegisterInterfaces(registry)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(stakingtypes.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data stakingtypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", stakingtypes.ModuleName, err)
	}
	return staking.ValidateGenesis(&data)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := stakingtypes.RegisterQueryHandlerClient(context.Background(), mux, stakingtypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (amb AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd(amb.cdc.InterfaceRegistry().SigningContext().ValidatorAddressCodec(), amb.cdc.InterfaceRegistry().SigningContext().AddressCodec())
}

type AppModule struct {
	AppModuleBasic

	keeper		*stakingkeeper.Keeper
	accountKeeper	stakingtypes.AccountKeeper
	bankKeeper	stakingtypes.BankKeeper
	legacySubspace	exported.Subspace
}

func NewAppModule(
	cdc codec.Codec,
	keeper *stakingkeeper.Keeper,
	accountKeeper stakingtypes.AccountKeeper,
	bankKeeper stakingtypes.BankKeeper,
	legacySubspace exported.Subspace,
) AppModule {
	return AppModule{
		AppModuleBasic:	AppModuleBasic{cdc: cdc},
		keeper:		keeper,
		accountKeeper:	accountKeeper,
		bankKeeper:	bankKeeper,
		legacySubspace:	legacySubspace,
	}
}

func (am AppModule) IsOnePerModuleType()	{}
func (am AppModule) IsAppModule()		{}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	inner := stakingkeeper.NewMsgServerImpl(am.keeper)
	stakingtypes.RegisterMsgServer(cfg.MsgServer(), NewPoolOnlyMsgServer(inner))
	querier := stakingkeeper.Querier{Keeper: am.keeper}
	stakingtypes.RegisterQueryServer(cfg.QueryServer(), querier)

	m := stakingkeeper.NewMigrator(am.keeper, am.legacySubspace)
	if err := cfg.RegisterMigration(stakingtypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", stakingtypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(stakingtypes.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 2 to 3: %v", stakingtypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(stakingtypes.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 3 to 4: %v", stakingtypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(stakingtypes.ModuleName, 4, m.Migrate4to5); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 4 to 5: %v", stakingtypes.ModuleName, err))
	}
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState stakingtypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	return am.keeper.InitGenesis(ctx, &genesisState)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(am.keeper.ExportGenesis(ctx))
}

func (AppModule) ConsensusVersion() uint64	{ return consensusVersion }

func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.keeper.BeginBlocker(ctx)
}

func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return am.keeper.EndBlocker(ctx)
}

func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

func (am AppModule) ProposalMsgsX(weights simsx.WeightSource, reg simsx.Registry) {
	reg.Add(weights.Get("msg_update_params", 100), simulation.MsgUpdateParamsFactory())
}

func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[stakingtypes.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

func (am AppModule) WeightedOperationsX(weights simsx.WeightSource, reg simsx.Registry) {
	reg.Add(weights.Get("msg_create_validator", 100), simulation.MsgCreateValidatorFactory(am.keeper))
	reg.Add(weights.Get("msg_edit_validator", 5), simulation.MsgEditValidatorFactory(am.keeper))
	reg.Add(weights.Get("msg_cancel_unbonding_delegation", 100), simulation.MsgCancelUnbondingDelegationFactory(am.keeper))
}
