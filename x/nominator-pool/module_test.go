package nominatorpool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/spf13/cobra"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/keeper"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestAppModuleRegistersMsgQueryServicesAndCommands(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, queryRouter := registerModuleServices(t, &k)

	for _, msg := range []sdk.Msg{
		&types.MsgCreateNominatorPool{},
		&types.MsgDepositToPool{},
		&types.MsgRequestPoolWithdrawal{},
		&types.MsgCancelPoolWithdrawal{},
		&types.MsgClaimPoolRewards{},
		&types.MsgSyncPoolRewards{},
		&types.MsgClaimStakingRewards{},
		&types.MsgUpdatePoolCommission{},
		&types.MsgChangePoolValidator{},
		&types.MsgDepositToStakingPool{},
		&types.MsgRequestPoolUnbond{},
		&types.MsgWithdrawPoolStake{},
		&types.MsgTopUpPoolReserve{},
		&types.MsgClaimStakeReputation{},
		&types.MsgDelegateToValidator{},
		&types.MsgRegisterValidator{},
		&types.MsgUpdateValidator{},
		&types.MsgUpdateStakingParams{},
		&types.MsgCreateOfficialLiquidStakingPool{},
	} {
		require.NotNil(t, msgRouter.Handler(msg))
	}
	for _, route := range []string{
		"/l1.nominatorpool.v1.Query/NominatorPool",
		"/l1.nominatorpool.v1.Query/NominatorPools",
		"/l1.nominatorpool.v1.Query/PoolDelegator",
		"/l1.nominatorpool.v1.Query/PoolRewards",
		"/l1.nominatorpool.v1.Query/PoolShare",
		"/l1.nominatorpool.v1.Query/PoolAllocations",
		"/l1.nominatorpool.v1.Query/StakeReputation",
		"/l1.nominatorpool.v1.Query/AccountReputation",
		"/l1.nominatorpool.v1.Query/StakingRewards",
		"/l1.nominatorpool.v1.Query/StakingProof",
		"/l1.nominatorpool.v1.Query/PoolUnbondingQueue",
	} {
		require.NotNil(t, queryRouter.Route(route))
	}

	module := NewAppModule(&k)
	require.NotNil(t, module.GetTxCmd())
	require.NotNil(t, module.GetQueryCmd())
	txCommands := commandNames(module.GetTxCmd().Commands())
	for _, name := range []string{
		"create-pool",
		"deposit-to-pool",
		"request-withdrawal",
		"cancel-withdrawal",
		"deposit",
		"request-unbond",
		"withdraw",
		"claim-rewards",
		"sync-rewards",
		"claim-staking-rewards",
		"claim-reputation",
		"top-up-reserve",
		"update-pool-commission",
		"change-pool-validator",
		"register-validator",
		"update-validator",
		"update-staking-params",
		"create-official-pool",
	} {
		require.Contains(t, txCommands, name)
	}
	queryCommands := commandNames(module.GetQueryCmd().Commands())
	for _, name := range []string{
		"pool",
		"pools",
		"pool-delegator",
		"pool-rewards",
		"pool-share",
		"pool-allocations",
		"stake-reputation",
		"account-reputation",
		"staking-rewards",
		"staking-proof",
		"pool-unbonding-queue",
	} {
		require.Contains(t, queryCommands, name)
	}
}

func TestMsgServiceDepositAndQuerySurface(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, _ := registerModuleServices(t, &k)
	pool := createServiceOfficialPool(t, &k)
	user := aeFromRawForServiceTest(t, serviceRawAddress("22"))

	handler := msgRouter.Handler(&types.MsgDepositToStakingPool{})
	require.NotNil(t, handler)
	_, err := handler(sdk.Context{}, &types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	query := keeper.NewQueryServerImpl(&k)
	share, err := query.PoolShare(context.Background(), &types.QueryPoolShareRequest{PoolID: pool.PoolID, Delegator: serviceRawAddress("22")})
	require.NoError(t, err)
	require.Equal(t, types.DefaultMinPoolDeposit, share.Share.Shares)

	poolRes, err := query.NominatorPool(context.Background(), &types.QueryNominatorPoolRequest{PoolID: pool.PoolID})
	require.NoError(t, err)
	require.Equal(t, pool.PoolID, poolRes.Pool.PoolID)

	poolsRes, err := query.NominatorPools(context.Background(), &types.QueryNominatorPoolsRequest{})
	require.NoError(t, err)
	require.Len(t, poolsRes.Pools, 1)

	delegatorRes, err := query.PoolDelegator(context.Background(), &types.QueryPoolDelegatorRequest{PoolID: pool.PoolID, Delegator: serviceRawAddress("22")})
	require.NoError(t, err)
	require.Equal(t, types.DefaultMinPoolDeposit, delegatorRes.Delegator.Shares)

	queueRes, err := query.PoolUnbondingQueue(context.Background(), &types.QueryPoolUnbondingQueueRequest{PoolID: pool.PoolID})
	require.NoError(t, err)
	require.Empty(t, queueRes.UnbondingQueue)
}

func TestMsgServiceDirectDelegationRejected(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, _ := registerModuleServices(t, &k)
	handler := msgRouter.Handler(&types.MsgDelegateToValidator{})
	require.NotNil(t, handler)

	_, err := handler(sdk.Context{}, &types.MsgDelegateToValidator{
		UserAddress:		aeFromRawForServiceTest(t, serviceRawAddress("22")),
		ValidatorAddress:	aeFromRawForServiceTest(t, serviceRawAddress("33")),
		Amount:			1,
		Height:			2,
	})
	require.ErrorContains(t, err, "direct user delegation to validators is disabled")
}

func TestMsgAndQueryCodecRoundTripPreservesSurfaceFields(t *testing.T) {
	k := keeper.NewKeeper()
	registry := codectypes.NewInterfaceRegistry()
	module := NewAppModule(&k)
	module.RegisterInterfaces(registry)
	appCodec := codec.NewProtoCodec(registry)
	user := aeFromRawForServiceTest(t, serviceRawAddress("22"))
	validator := aeFromRawForServiceTest(t, serviceRawAddress("33"))

	deposit := &types.MsgDepositToStakingPool{
		PoolID:		"pool-codec",
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		9,
	}
	depositBz, err := appCodec.Marshal(deposit)
	require.NoError(t, err)
	var depositOut types.MsgDepositToStakingPool
	require.NoError(t, appCodec.Unmarshal(depositBz, &depositOut))
	require.Equal(t, *deposit, depositOut)

	directDelegation := &types.MsgDelegateToValidator{
		Authority:		prototype.DefaultAuthority,
		UserAddress:		user,
		ValidatorAddress:	validator,
		Amount:			1,
		Height:			10,
	}
	delegationBz, err := appCodec.Marshal(directDelegation)
	require.NoError(t, err)
	var delegationOut types.MsgDelegateToValidator
	require.NoError(t, appCodec.Unmarshal(delegationBz, &delegationOut))
	require.Equal(t, *directDelegation, delegationOut)

	query := &types.QueryPoolShareRequest{PoolID: "pool-codec", Delegator: serviceRawAddress("22")}
	queryBz, err := appCodec.Marshal(query)
	require.NoError(t, err)
	var queryOut types.QueryPoolShareRequest
	require.NoError(t, appCodec.Unmarshal(queryBz, &queryOut))
	require.Equal(t, *query, queryOut)

	legacyPoolQuery := &types.QueryNominatorPoolRequest{PoolID: "pool-codec"}
	legacyPoolQueryBz, err := appCodec.Marshal(legacyPoolQuery)
	require.NoError(t, err)
	var legacyPoolQueryOut types.QueryNominatorPoolRequest
	require.NoError(t, appCodec.Unmarshal(legacyPoolQueryBz, &legacyPoolQueryOut))
	require.Equal(t, *legacyPoolQuery, legacyPoolQueryOut)
}

func registerModuleServices(t *testing.T, k *keeper.Keeper) (*baseapp.MsgServiceRouter, *baseapp.GRPCQueryRouter) {
	t.Helper()
	registry := codectypes.NewInterfaceRegistry()
	module := NewAppModule(k)
	module.RegisterInterfaces(registry)
	appCodec := codec.NewProtoCodec(registry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(registry)
	queryRouter := baseapp.NewGRPCQueryRouter()
	configurator := module2Configurator(appCodec, msgRouter, queryRouter)
	module.RegisterServices(configurator)
	return msgRouter, queryRouter
}

func module2Configurator(appCodec codec.Codec, msgRouter *baseapp.MsgServiceRouter, queryRouter *baseapp.GRPCQueryRouter) module.Configurator {
	return module.NewConfigurator(appCodec, msgRouter, queryRouter)
}

func createServiceOfficialPool(t *testing.T, k *keeper.Keeper) types.NominatorPool {
	t.Helper()
	contractRaw := serviceRawAddress("66")
	pool, err := k.CreateOfficialLiquidStakingPool(types.MsgCreateOfficialLiquidStakingPool{
		Authority:		prototype.DefaultAuthority,
		PoolID:			"service-pool",
		ContractAddressUser:	aeFromRawForServiceTest(t, contractRaw),
		ContractAddressRaw:	contractRaw,
		PoolOperator:		serviceRawAddress("11"),
		PoolCommissionBps:	100,
		Height:			1,
	})
	require.NoError(t, err)
	return pool
}

func commandNames(commands []*cobra.Command) []string {
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.Name())
	}
	return names
}

func serviceRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}

func aeFromRawForServiceTest(t *testing.T, raw string) string {
	t.Helper()
	bz, err := addressing.Parse(raw)
	require.NoError(t, err)
	user, err := addressing.FormatUserFriendly(bz)
	require.NoError(t, err)
	return user
}
