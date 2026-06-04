package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func TestParamsQueryRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.FeesKeeper.Params(ctx, nil)
	require.Error(t, err)
}

func TestAccountingQueryReturnsState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.FeesKeeper.RecordCollectedFees(ctx, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1000))))

	res, err := app.FeesKeeper.Accounting(ctx, &types.QueryAccountingRequest{})
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1000)), res.ProtocolFeeState.TotalCollected)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 980)), res.ProtocolFeeState.ValidatorRewards)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 20)), res.ProtocolFeeState.CommunityPool)
}

func TestModuleBalancesQueryReturnsFeeAccounts(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	feeCollector := app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	before := app.BankKeeper.GetBalance(ctx, feeCollector, types.BondDenom)
	coins := sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 25))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, coins))

	res, err := app.FeesKeeper.ModuleBalances(ctx, &types.QueryModuleBalancesRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, res.Balances)
	require.Contains(t, res.Balances, types.ModuleBalance{
		ModuleName: authtypes.FeeCollectorName,
		Address:    feeCollector.String(),
		Balance:    sdk.NewCoins(before.Add(sdk.NewInt64Coin(types.BondDenom, 25))),
	})
}
