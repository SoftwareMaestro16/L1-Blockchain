package keeper_test

import (
	"testing"

	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/fees/types"
)

type feeTx struct {
	fees sdk.Coins
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
	return nil
}

func (tx feeTx) FeeGranter() []byte {
	return nil
}

func TestAnteHandlerDecoratorRejectsNonNativeFeeDenom(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

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
	ctx := app.NewContext(false)

	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1))}, false)
	require.NoError(t, err)
	require.True(t, called)
}
