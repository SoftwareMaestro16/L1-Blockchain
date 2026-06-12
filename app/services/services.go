package services

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	stakingv1beta "cosmossdk.io/api/cosmos/staking/v1beta1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func AutoCLIOptions(modules map[string]any) autocli.AppOptions {
	appModules := make(map[string]appmodule.AppModule)
	for _, m := range modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				appModules[moduleName] = appModule
			}
		}
	}

	moduleOptions := runtimeservices.ExtractAutoCLIOptions(modules)
	disableConflictingStakingHistoricalInfoCommand(moduleOptions)

	return autocli.AppOptions{
		Modules:		appModules,
		ModuleOptions:		moduleOptions,
		AddressCodec:		aetraaddress.Codec{},
		ValidatorAddressCodec:	aetraaddress.Codec{},
		ConsensusAddressCodec:	aetraaddress.Codec{},
	}
}

func disableConflictingStakingHistoricalInfoCommand(moduleOptions map[string]*autocliv1.ModuleOptions) {
	stakingOptions := moduleOptions[stakingtypes.ModuleName]
	if stakingOptions == nil {
		stakingOptions = &autocliv1.ModuleOptions{}
		moduleOptions[stakingtypes.ModuleName] = stakingOptions
	}
	if stakingOptions.Query == nil {
		stakingOptions.Query = &autocliv1.ServiceCommandDescriptor{Service: stakingv1beta.Query_ServiceDesc.ServiceName}
	}
	if stakingOptions.Query.Service == "" {
		stakingOptions.Query.Service = stakingv1beta.Query_ServiceDesc.ServiceName
	}
	stakingOptions.Query.RpcCommandOptions = append(stakingOptions.Query.RpcCommandOptions, &autocliv1.RpcCommandOptions{
		RpcMethod:	"HistoricalInfo",
		Skip:		true,
	})

	for _, options := range moduleOptions {
		if options == nil || options.Query == nil {
			continue
		}
		for _, rpc := range options.Query.RpcCommandOptions {
			if rpc.RpcMethod == "HistoricalInfo" {
				rpc.Skip = true
			}
		}
	}
}

func RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig, basicManager module.BasicManager) {
	clientCtx := apiSvr.ClientCtx
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	basicManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}
