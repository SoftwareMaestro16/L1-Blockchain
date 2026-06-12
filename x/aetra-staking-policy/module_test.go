package aetrastakingpolicy

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/x/aetra-staking-policy/keeper"
	"github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestAppModuleRegistersRuntimeServicesAndCommands(t *testing.T) {
	k := keeper.NewKeeper(prototype.DefaultAuthority)
	msgRouter, queryRouter := registerStakingPolicyServices(t, &k)

	require.NotNil(t, msgRouter.Handler(&types.MsgRegisterValidatorIdentity{}))
	require.NotNil(t, queryRouter.Route("/l1.aetrastakingpolicy.v1.Query/Params"))
	require.NotNil(t, NewAppModule(&k).GetTxCmd())
	require.NotNil(t, NewAppModule(&k).GetQueryCmd())
}

func TestMsgServiceRejectsUnauthorizedIdentityRegistration(t *testing.T) {
	k := keeper.NewKeeper(prototype.DefaultAuthority)
	msgRouter, _ := registerStakingPolicyServices(t, &k)
	handler := msgRouter.Handler(&types.MsgRegisterValidatorIdentity{})
	require.NotNil(t, handler)

	_, err := handler(sdk.Context{}, &types.MsgRegisterValidatorIdentity{
		Authority:	"bad-authority",
		Identity:	types.ValidatorIdentityMetadata{OperatorAddress: "AEvalidator", Moniker: "validator"},
	})
	require.ErrorContains(t, err, "invalid authority")
}

func registerStakingPolicyServices(t *testing.T, k *keeper.Keeper) (*baseapp.MsgServiceRouter, *baseapp.GRPCQueryRouter) {
	t.Helper()
	registry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewProtoCodec(registry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(registry)
	queryRouter := baseapp.NewGRPCQueryRouter()
	appModule := NewAppModule(k)
	appModule.RegisterInterfaces(registry)
	appModule.RegisterServices(sdkmodule.NewConfigurator(appCodec, msgRouter, queryRouter))
	return msgRouter, queryRouter
}
