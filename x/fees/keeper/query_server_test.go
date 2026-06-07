package keeper_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func TestParamsQueryReturnsCurrentParams(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	res, err := app.FeesKeeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), res.Params)
}

func TestParamsQueryRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.FeesKeeper.Params(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestModuleBalancesQueryReturnsFeeCollectorBalance(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	feeCollector := app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	require.NotNil(t, feeCollector)
	before := app.BankKeeper.GetAllBalances(ctx, feeCollector)
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 100))))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 100))))

	res, err := app.FeesKeeper.ModuleBalances(ctx, &types.QueryModuleBalancesRequest{})
	require.NoError(t, err)

	found := false
	for _, balance := range res.Balances {
		if balance.ModuleName != types.FeeCollectorModuleName {
			continue
		}
		found = true
		require.Equal(t, aetraaddress.FormatAccAddress(feeCollector), balance.Address)
		require.Equal(t, before.Add(sdk.NewInt64Coin(types.BondDenom, 100)), balance.Balance)
	}
	require.True(t, found, "fee collector balance must be returned")
}

func TestModuleBalancesQueryRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.FeesKeeper.ModuleBalances(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestNetworkLoadQueryReturnsCurrentBlockUtilization(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockGasMeter(storetypes.NewGasMeter(types.DefaultParams().MaxBlockGas))
	ctx.BlockGasMeter().ConsumeGas(1_000_000, "test load")

	res, err := app.FeesKeeper.NetworkLoad(ctx, &types.QueryNetworkLoadRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(1_000_000), res.BlockGasConsumed)
	require.Equal(t, types.DefaultParams().MaxBlockGas, res.MaxBlockGas)
	require.Equal(t, uint32(500), res.UtilizationBps)
	require.False(t, res.Congested)
}

func TestEstimateFeeQueryReturnsBoundedQuote(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockGasMeter(storetypes.NewGasMeter(types.DefaultParams().MaxBlockGas))

	res, err := app.FeesKeeper.EstimateFee(ctx, &types.QueryEstimateFeeRequest{GasLimit: types.DefaultParams().MaxTxGas})
	require.NoError(t, err)
	require.Equal(t, "1naet", res.RequiredFee)
	require.Equal(t, "1naet", res.BaseFee)
	require.Equal(t, "1000naet", res.MaxFee)
	require.False(t, res.Congested)
	require.False(t, res.AtHardCap)

	ctx.BlockGasMeter().ConsumeGas(types.DefaultParams().MaxBlockGas-types.DefaultParams().MaxTxGas, "test congestion")
	res, err = app.FeesKeeper.EstimateFee(ctx, &types.QueryEstimateFeeRequest{GasLimit: types.DefaultParams().MaxTxGas})
	require.NoError(t, err)
	require.Equal(t, "1000naet", res.RequiredFee)
	require.True(t, res.Congested)
	require.True(t, res.AtHardCap)
}

func TestFeeEstimateQueriesRejectNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.FeesKeeper.NetworkLoad(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = app.FeesKeeper.EstimateFee(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
