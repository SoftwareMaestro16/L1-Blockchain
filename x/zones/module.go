package zones

import (
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/zones/keeper"
	"github.com/sovereign-l1/l1/x/zones/types"
)

const ConsensusVersion = prototype.NextMigrationVersion

var (
	_	module.AppModuleBasic	= AppModule{}
	_	module.HasGenesis	= AppModule{}
	_	module.HasServices	= AppModule{}
	_	appmodule.AppModule	= AppModule{}
)

type AppModule struct {
	keeper *keeper.Keeper
}

func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

func (AppModule) IsOnePerModuleType()						{}
func (AppModule) IsAppModule()							{}
func (AppModule) Name() string							{ return types.ModuleName }
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino)			{}
func (AppModule) RegisterInterfaces(codectypes.InterfaceRegistry)		{}
func (AppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)	{}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	if err := cfg.RegisterMigration(types.ModuleName, 1, func(ctx sdk.Context) error {
		return am.keeper.Migrate1to2State(ctx)
	}); err != nil {
		panic(fmt.Sprintf("failed to register x/%s migration from version 1 to 2: %v", types.ModuleName, err))
	}
}

func (AppModule) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	return mustMarshalGenesis(types.ModuleName, keeper.DefaultGenesis())
}

func (AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs keeper.GenesisState
	if err := unmarshalGenesis(types.ModuleName, bz, &gs); err != nil {
		return err
	}
	return gs.Validate()
}

func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, bz json.RawMessage) {
	var gs keeper.GenesisState
	if err := unmarshalGenesis(types.ModuleName, bz, &gs); err != nil {
		panic(err)
	}
	if err := am.keeper.InitGenesisState(ctx, gs); err != nil {
		panic(fmt.Errorf("failed to initialize %s genesis: %w", types.ModuleName, err))
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesisState(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to export %s genesis: %w", types.ModuleName, err))
	}
	return mustMarshalGenesis(types.ModuleName, gs)
}

func (AppModule) ConsensusVersion() uint64	{ return ConsensusVersion }
func (AppModule) GetTxCmd() *cobra.Command	{ return nil }
func (AppModule) GetQueryCmd() *cobra.Command	{ return nil }

func mustMarshalGenesis(moduleName string, value any) json.RawMessage {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Errorf("failed to marshal %s genesis: %w", moduleName, err))
	}
	return bz
}

func unmarshalGenesis(moduleName string, bz json.RawMessage, target any) error {
	if len(bz) == 0 {
		return fmt.Errorf("missing %s genesis state", moduleName)
	}
	if err := json.Unmarshal(bz, target); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", moduleName, err)
	}
	return nil
}
