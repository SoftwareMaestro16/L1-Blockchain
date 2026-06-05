package keeper_test

import (
	"errors"
	"testing"

	sdkmath "cosmossdk.io/math"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/fees/types"
)

type feeTx struct {
	fees  sdk.Coins
	payer sdk.AccAddress
	msgs  []sdk.Msg
}

func (tx feeTx) GetMsgs() []sdk.Msg {
	return tx.msgs
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

type noFeeTx struct{}

func (tx noFeeTx) GetMsgs() []sdk.Msg {
	return nil
}

func (tx noFeeTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func TestAnteHandlerDecoratorFeePolicy(t *testing.T) {
	tests := []struct {
		name         string
		tx           sdk.Tx
		wantErr      string
		wantNextCall bool
	}{
		{
			name:         "accepts native fee denom",
			tx:           feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1))},
			wantNextCall: true,
		},
		{
			name:    "rejects empty fee list",
			tx:      feeTx{fees: sdk.Coins{}},
			wantErr: "fee must be positive",
		},
		{
			name:    "rejects nil fee list",
			tx:      feeTx{},
			wantErr: "fee must be positive",
		},
		{
			name:    "rejects zero native fee coin",
			tx:      feeTx{fees: sdk.Coins{sdk.NewInt64Coin(types.BondDenom, 0)}},
			wantErr: "invalid fee coins",
		},
		{
			name:    "rejects non native fee denom",
			tx:      feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin("uatom", 1))},
			wantErr: "fee denom uatom not accepted; use norb",
		},
		{
			name:    "rejects mixed native and non native fee denoms",
			tx:      feeTx{fees: sdk.Coins{sdk.NewInt64Coin(types.BondDenom, 1), sdk.NewInt64Coin("testtoken", 1)}},
			wantErr: "fee denom testtoken not accepted; use norb",
		},
		{
			name:    "rejects malformed fee coin",
			tx:      feeTx{fees: sdk.Coins{{Denom: "!", Amount: sdkmath.NewInt(1)}}},
			wantErr: "invalid fee coins",
		},
		{
			name: "rejects duplicate fee denom entries",
			tx: feeTx{fees: sdk.Coins{
				sdk.NewInt64Coin(types.BondDenom, 1),
				sdk.NewInt64Coin(types.BondDenom, 2),
			}},
			wantErr: "invalid fee coins",
		},
		{
			name:    "rejects transaction without fee interface",
			tx:      noFeeTx{},
			wantErr: "transaction must expose fees",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := l1app.Setup(t, false)
			ctx := app.NewContext(false)

			called := false
			next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
				called = true
				return ctx, nil
			}

			_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, tc.tx, false)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, types.ErrInvalidFee)
				require.Contains(t, err.Error(), tc.wantErr)
			}
			require.Equal(t, tc.wantNextCall, called)
		})
	}
}

func TestAnteHandlerDecoratorPropagatesNextError(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	nextErr := errors.New("next failed")

	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nextErr
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1))}, false)
	require.ErrorIs(t, err, nextErr)
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

func TestAnteHandlerDecoratorAllowsGenesisCreateValidatorWithoutFee(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(0)

	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{
		fees: sdk.Coins{},
		msgs: []sdk.Msg{&stakingtypes.MsgCreateValidator{}},
	}, false)
	require.NoError(t, err)
	require.True(t, called)
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
