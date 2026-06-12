package constitution

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/x/constitution/keeper"
	"github.com/sovereign-l1/l1/x/constitution/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestAppModuleRegistersRuntimeServicesAndCommands(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, queryRouter := registerConstitutionServices(t, &k)

	require.NotNil(t, msgRouter.Handler(&types.MsgProposeConstitutionAmendment{}))
	require.NotNil(t, msgRouter.Handler(&types.MsgExecuteConstitutionAmendment{}))
	require.NotNil(t, queryRouter.Route("/l1.constitution.v1.Query/Constitution"))
	require.NotNil(t, queryRouter.Route("/l1.constitution.v1.Query/ProtectedLimits"))
	require.NotNil(t, NewAppModule(&k).GetTxCmd())
	require.NotNil(t, NewAppModule(&k).GetQueryCmd())
}

func TestConstitutionServiceDelayAndBounds(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, _ := registerConstitutionServices(t, &k)
	ctx := sdk.Context{}.WithBlockHeight(10)
	proposed := types.DefaultConstitution().Normalize()
	proposed.MaxBlockGas = 2_000_000

	_, err := msgRouter.Handler(&types.MsgProposeConstitutionAmendment{})(ctx, &types.MsgProposeConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		Amendment: types.Amendment{
			ID:		"raise-gas-bound",
			Proposed:	proposed,
		},
	})
	require.NoError(t, err)

	_, err = msgRouter.Handler(&types.MsgVoteConstitutionAmendment{})(ctx.WithBlockHeight(11), &types.MsgVoteConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		AmendmentID:	"raise-gas-bound",
		Support:	types.VoteSupportYes,
		VotingPowerBps:	7_000,
	})
	require.NoError(t, err)

	_, err = msgRouter.Handler(&types.MsgExecuteConstitutionAmendment{})(ctx.WithBlockHeight(99), &types.MsgExecuteConstitutionAmendment{
		Authority:	prototype.DefaultAuthority,
		AmendmentID:	"raise-gas-bound",
	})
	require.ErrorContains(t, err, "delay")

	query := keeper.NewGRPCQueryServer(&k)
	pending, err := query.PendingAmendments(nil, &types.QueryPendingAmendmentsRequest{Limit: 1})
	require.NoError(t, err)
	require.Len(t, pending.Amendments, 1)
	require.Equal(t, uint64(110), pending.Amendments[0].ExecutableHeight)
}

func registerConstitutionServices(t *testing.T, k *keeper.Keeper) (*baseapp.MsgServiceRouter, *baseapp.GRPCQueryRouter) {
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
