package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/fees/types"
)

type feeTx struct {
	fees  sdk.Coins
	payer sdk.AccAddress
}

func (tx feeTx) GetMsgs() []sdk.Msg {
	return nil
}

func (tx feeTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func (tx feeTx) GetGas() uint64 {
	return 100_000
}

func (tx feeTx) GetFee() sdk.Coins {
	return tx.fees
}

func (tx feeTx) FeePayer() []byte {
	return tx.payer
}

func (tx feeTx) FeeGranter() []byte {
	return nil
}

func TestAnteHandlerDecoratorRejectsNonNativeFeeDenom(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(1)

	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin("uatom", 1))}, false)
	require.ErrorIs(t, err, types.ErrInvalidFee)
	require.False(t, called)
	require.Contains(t, err.Error(), types.BondDenom)
}

func TestAnteHandlerDecoratorAcceptsNativeFeeDenom(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(1)

	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1))}, false)
	require.NoError(t, err)
	require.True(t, called)
}

func TestAnteHandlerDecoratorRejectsZeroFee(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(1)

	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.Coins{}}, false)
	require.ErrorIs(t, err, types.ErrInvalidFee)
	require.False(t, called)
}

func TestAnteHandlerDecoratorRejectsBelowMinimumFee(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(1)
	params := types.DefaultParams()
	params.MinFeeAmount = "100"
	require.NoError(t, app.FeesKeeper.SetParams(ctx, params))

	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 99))}, false)
	require.ErrorIs(t, err, types.ErrInvalidFee)
	require.False(t, called)
}

func TestAnteHandlerDecoratorRecordsFeesAfterDeduction(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(1)
	payer := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	fee := sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1000))
	feeCollector := app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	before := app.BankKeeper.GetBalance(ctx, feeCollector, types.BondDenom)

	next := func(ctx sdk.Context, tx sdk.Tx, _ bool) (sdk.Context, error) {
		feeTx := tx.(sdk.FeeTx)
		if err := app.BankKeeper.SendCoinsFromAccountToModule(ctx, sdk.AccAddress(feeTx.FeePayer()), authtypes.FeeCollectorName, feeTx.GetFee()); err != nil {
			return ctx, err
		}
		return ctx, nil
	}

	newCtx, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: fee, payer: payer}, false)
	require.NoError(t, err)
	require.Equal(t, before.Add(sdk.NewInt64Coin(types.BondDenom, 1000)), app.BankKeeper.GetBalance(newCtx, feeCollector, types.BondDenom))

	state, err := app.FeesKeeper.GetProtocolFeeState(newCtx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1000)), state.TotalCollected)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 980)), state.ValidatorRewards)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 20)), state.CommunityPool)
	require.NoError(t, state.Validate())
}
