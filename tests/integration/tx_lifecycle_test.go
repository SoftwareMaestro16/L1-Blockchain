package integration_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	testutil "github.com/sovereign-l1/l1/tests/testutil"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tfkeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	tftypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestSignedBankTxReplayIsRejectedAfterSequenceIncrement(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-integration-1")
	ctx := testutil.NewContext(app, 1)
	senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))

	msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("norb", 100)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, senderPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("norb", 10)), 200_000)

	first := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, first.TxResults, 1)
	require.Zero(t, first.TxResults[0].Code, first.TxResults[0].Log)
	testutil.Commit(t, app)

	ctxAfterFirst := testutil.NewContext(app, 2)
	recipientAfterFirst := app.BankKeeper.GetBalance(ctxAfterFirst, recipient, "norb")
	require.Equal(t, sdkmath.NewInt(1_000_100), recipientAfterFirst.Amount)

	replay := testutil.FinalizeBlock(t, app, 2, txBytes)
	require.Len(t, replay.TxResults, 1)
	require.NotZero(t, replay.TxResults[0].Code, "replayed tx with stale sequence must fail")
	testutil.Commit(t, app)

	ctxAfterReplay := testutil.NewContext(app, 3)
	require.Equal(t, recipientAfterFirst, app.BankKeeper.GetBalance(ctxAfterReplay, recipient, "norb"))
}

func TestInvalidSignerTxFailsBeforeBalanceMutation(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-integration-2")
	ctx := testutil.NewContext(app, 1)
	_, victim := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	attackerPriv, _ := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	before := app.BankKeeper.GetBalance(ctx, recipient, "norb")

	msg := banktypes.NewMsgSend(victim, recipient, sdk.NewCoins(sdk.NewInt64Coin("norb", 100)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, attackerPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("norb", 10)), 200_000)
	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code, "tx signed by non-msg signer must fail")

	after := app.BankKeeper.GetBalance(testutil.NewContext(app, 1), recipient, "norb")
	require.Equal(t, before, after)
}

func TestTokenfactoryDexFeesCrossModuleLifecycle(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-integration-3")
	ctx := testutil.NewContext(app, 1)
	adminPriv, admin := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(2_000_000))
	_, trader := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(2_000_000))
	tfMsgServer := tfkeeper.NewMsgServerImpl(app.TokenFactoryKeeper)
	dexMsgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	createDenom, err := tfMsgServer.CreateDenom(ctx, &tftypes.MsgCreateDenom{Creator: admin.String(), Subdenom: "silver"})
	require.NoError(t, err)
	denom := createDenom.NewTokenDenom
	_, err = tfMsgServer.Mint(ctx, &tftypes.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 2_000),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)

	createPool, err := dexMsgServer.CreatePool(ctx, &dextypes.MsgCreatePool{
		Creator: admin.String(),
		TokenA:  sdk.NewInt64Coin("norb", 1_000),
		TokenB:  sdk.NewInt64Coin(denom, 1_000),
	})
	require.NoError(t, err)
	_, err = dexMsgServer.SwapExactAmountIn(ctx, &dextypes.MsgSwapExactAmountIn{
		Trader:        admin.String(),
		PoolId:        createPool.PoolId,
		TokenIn:       sdk.NewInt64Coin("norb", 10),
		TokenOutDenom: denom,
		MinAmountOut:  "1",
	})
	require.NoError(t, err)

	denomQuery, err := app.TokenFactoryKeeper.Denom(ctx, &tftypes.QueryDenomRequest{Denom: denom})
	require.NoError(t, err)
	require.Equal(t, admin.String(), denomQuery.Metadata.Admin)
	poolQuery, err := app.DexKeeper.Pool(ctx, &dextypes.QueryPoolRequest{PoolId: createPool.PoolId})
	require.NoError(t, err)
	testutil.AssertPoolAccounting(t, app, ctx, poolQuery.Pool)

	msg := banktypes.NewMsgSend(admin, trader, sdk.NewCoins(sdk.NewInt64Coin("norb", 100)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, adminPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("norb", 10)), 200_000)
	block := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, block.TxResults, 1)
	require.Zero(t, block.TxResults[0].Code, block.TxResults[0].Log)
	testutil.Commit(t, app)

	accounting, err := app.FeesKeeper.Accounting(testutil.NewContext(app, 2), &feestypes.QueryAccountingRequest{})
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("norb", 10)), accounting.ProtocolFeeState.TotalCollected)
	require.NoError(t, accounting.ProtocolFeeState.Validate())
}
