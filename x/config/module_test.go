package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/x/config/keeper"
	"github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestAppModuleRegistersRuntimeServicesAndCommands(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, queryRouter := registerConfigServices(t, &k)

	require.NotNil(t, msgRouter.Handler(&types.MsgSubmitConfigChange{}))
	require.NotNil(t, msgRouter.Handler(&types.MsgExecuteConfigChange{}))
	require.NotNil(t, queryRouter.Route("/l1.config.v1.Query/Params"))
	require.NotNil(t, queryRouter.Route("/l1.config.v1.Query/EffectiveParams"))
	require.NotNil(t, NewAppModule(&k).GetTxCmd())
	require.NotNil(t, NewAppModule(&k).GetQueryCmd())
}

func TestConfigMsgServiceDelayEventsAndQueries(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, _ := registerConfigServices(t, &k)
	ctx := sdk.Context{}.WithBlockHeight(7).WithEventManager(sdk.NewEventManager())

	submit := msgRouter.Handler(&types.MsgSubmitConfigChange{})
	result, err := submit(ctx, &types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change: types.ConfigChange{
			ID:		"service-gas",
			Key:		types.KeyConsensusMaxBlockGas,
			Value:		"1000000",
			Operation:	types.OperationSet,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Events)

	change, found, err := k.ConfigChange("service-gas")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(150), change.ActivationHeight)

	approve := msgRouter.Handler(&types.MsgApproveConfigChange{})
	_, err = approve(ctx.WithBlockHeight(8), &types.MsgApproveConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "service-gas"})
	require.NoError(t, err)

	execute := msgRouter.Handler(&types.MsgExecuteConfigChange{})
	_, err = execute(ctx.WithBlockHeight(change.ActivationHeight-1), &types.MsgExecuteConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "service-gas"})
	require.ErrorContains(t, err, "activation height")

	query := keeper.NewGRPCQueryServer(&k)
	pending, err := query.PendingChanges(nil, &types.QueryPendingChangesRequest{Limit: 1})
	require.NoError(t, err)
	require.Len(t, pending.Changes, 1)

	effectiveBefore, err := query.EffectiveParams(ctx.WithBlockHeight(change.ActivationHeight-1), &types.QueryEffectiveParamsRequest{})
	require.NoError(t, err)
	require.Empty(t, effectiveBefore.Entries)

	_, err = execute(ctx.WithBlockHeight(change.ActivationHeight), &types.MsgExecuteConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "service-gas"})
	require.NoError(t, err)
	effectiveAfter, err := query.EffectiveParams(ctx.WithBlockHeight(change.ActivationHeight), &types.QueryEffectiveParamsRequest{})
	require.NoError(t, err)
	require.Len(t, effectiveAfter.Entries, 1)
}

func registerConfigServices(t *testing.T, k *keeper.Keeper) (*baseapp.MsgServiceRouter, *baseapp.GRPCQueryRouter) {
	t.Helper()
	registry := codectypes.NewInterfaceRegistry()
	NewAppModule(k).RegisterInterfaces(registry)
	appCodec := codec.NewProtoCodec(registry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(registry)
	queryRouter := baseapp.NewGRPCQueryRouter()
	NewAppModule(k).RegisterServices(sdkmodule.NewConfigurator(appCodec, msgRouter, queryRouter))
	return msgRouter, queryRouter
}
