package delegatorprotection

import (
	"context"
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

	"github.com/sovereign-l1/l1/x/delegator-protection/keeper"
	"github.com/sovereign-l1/l1/x/delegator-protection/types"
	delegatorprotectionpb "github.com/sovereign-l1/l1/x/delegator-protection/types/delegatorprotectionpb"
)

const ConsensusVersion = 1

var (
	_	module.AppModuleBasic	= AppModule{}
	_	module.HasGenesis	= AppModule{}
	_	module.HasServices	= AppModule{}
	_	appmodule.AppModule	= AppModule{}
)

type AppModule struct {
	cdc	codec.Codec
	keeper	keeper.Keeper
}

func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule	{ return AppModule{cdc: cdc, keeper: k} }
func (AppModule) IsOnePerModuleType()				{}
func (AppModule) IsAppModule()					{}
func (AppModule) Name() string					{ return types.ModuleName }
func (AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino)	{}
func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	delegatorprotectionpb.RegisterInterfaces(registry)
}

func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := delegatorprotectionpb.RegisterQueryHandlerClient(context.Background(), mux, delegatorprotectionpb.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	delegatorprotectionpb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	delegatorprotectionpb.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

func (am AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(am.keeper.DefaultGenesis())
	if err != nil {
		panic(err)
	}
	return bz
}

func (am AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.DelegatorProtectionState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.NormalizeDelegatorProtectionState(gs).Validate()
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var gs types.DelegatorProtectionState
	if err := json.Unmarshal(bz, &gs); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err))
	}
	if err := am.keeper.InitGenesis(ctx, gs); err != nil {
		panic(fmt.Errorf("failed to initialize %s genesis: %w", types.ModuleName, err))
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to export %s genesis: %w", types.ModuleName, err))
	}
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am AppModule) ConsensusVersion() uint64		{ return ConsensusVersion }
func (am AppModule) GetTxCmd() *cobra.Command		{ return nil }
func (am AppModule) GetQueryCmd() *cobra.Command	{ return nil }
