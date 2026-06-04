package adversarial_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	testutil "github.com/sovereign-l1/l1/tests/testutil"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tfkeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	tftypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestMalformedTxBytesFailSafely(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-adversarial-1")

	for _, bz := range [][]byte{{0xff}, {0x0a, 0x80, 0x80}, []byte("not-a-protobuf-tx")} {
		require.NotPanics(t, func() {
			_, err := app.TxConfig().TxDecoder()(bz)
			require.Error(t, err)
		})
	}
}

func TestTokenfactoryAdminTakeoverAndSupplyMismatchAttempts(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-adversarial-2")
	ctx := testutil.NewContext(app, 1)
	adminPriv, admin := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, attacker := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, nextAdmin := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	msgServer := tfkeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &tftypes.MsgCreateDenom{Creator: admin.String(), Subdenom: "gold"})
	require.NoError(t, err)
	denom := createRes.NewTokenDenom
	_, err = msgServer.Mint(ctx, &tftypes.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 100),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)
	supplyBefore := app.BankKeeper.GetSupply(ctx, denom)

	_, err = msgServer.Mint(ctx, &tftypes.MsgMint{
		Sender:        attacker.String(),
		Amount:        sdk.NewInt64Coin(denom, 50),
		MintToAddress: attacker.String(),
	})
	require.ErrorIs(t, err, tftypes.ErrUnauthorized)
	_, err = msgServer.ChangeAdmin(ctx, &tftypes.MsgChangeAdmin{
		Sender:   attacker.String(),
		Denom:    denom,
		NewAdmin: attacker.String(),
	})
	require.ErrorIs(t, err, tftypes.ErrUnauthorized)
	require.Equal(t, supplyBefore, app.BankKeeper.GetSupply(ctx, denom))

	_, err = msgServer.ChangeAdmin(ctx, &tftypes.MsgChangeAdmin{
		Sender:   admin.String(),
		Denom:    denom,
		NewAdmin: nextAdmin.String(),
	})
	require.NoError(t, err)
	_, err = msgServer.Mint(ctx, &tftypes.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 1),
		MintToAddress: admin.String(),
	})
	require.ErrorIs(t, err, tftypes.ErrUnauthorized)
	_, err = msgServer.Mint(ctx, &tftypes.MsgMint{
		Sender:        nextAdmin.String(),
		Amount:        sdk.NewInt64Coin(denom, 10),
		MintToAddress: nextAdmin.String(),
	})
	require.NoError(t, err)
	require.Equal(t, supplyBefore.AddAmount(sdkmath.NewInt(10)), app.BankKeeper.GetSupply(ctx, denom))
	require.NotNil(t, adminPriv)
}

func TestDexManipulationAndCorruptedStateDoNotMutateSuccessfulPool(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-adversarial-3")
	ctx := testutil.NewContext(app, 1)
	_, trader := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	testutil.FundAccount(t, app, ctx, trader, sdk.NewCoins(sdk.NewInt64Coin("uatom", 10_000)))
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	createRes, err := msgServer.CreatePool(ctx, &dextypes.MsgCreatePool{
		Creator: trader.String(),
		TokenA:  sdk.NewInt64Coin("norb", 1_000),
		TokenB:  sdk.NewInt64Coin("uatom", 1_000),
	})
	require.NoError(t, err)
	poolBefore, found, err := app.DexKeeper.GetPool(ctx, createRes.PoolId)
	require.NoError(t, err)
	require.True(t, found)
	testutil.AssertPoolAccounting(t, app, ctx, poolBefore)

	_, err = msgServer.CreatePool(ctx, &dextypes.MsgCreatePool{
		Creator: trader.String(),
		TokenA:  sdk.NewInt64Coin("norb", 1),
		TokenB:  sdk.NewInt64Coin("norb", 1),
	})
	require.ErrorIs(t, err, dextypes.ErrInvalidPool)
	_, err = msgServer.CreatePool(ctx, &dextypes.MsgCreatePool{
		Creator: trader.String(),
		TokenA:  sdk.NewInt64Coin("uatom", 1),
		TokenB:  sdk.NewInt64Coin("norb", 1),
	})
	require.ErrorIs(t, err, dextypes.ErrInvalidPool)
	_, err = msgServer.SwapExactAmountIn(ctx, &dextypes.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        createRes.PoolId,
		TokenIn:       sdk.NewInt64Coin("uatom", 10),
		TokenOutDenom: "norb",
		MinAmountOut:  "999999999999",
	})
	require.ErrorIs(t, err, dextypes.ErrSlippage)

	poolAfter, found, err := app.DexKeeper.GetPool(ctx, createRes.PoolId)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, poolBefore, poolAfter)
	testutil.AssertPoolAccounting(t, app, ctx, poolAfter)

	require.NoError(t, app.DexKeeper.SetPool(ctx, dextypes.Pool{
		Id:          99,
		Denom0:      "norb",
		Denom1:      "uatom",
		Reserve0:    "corrupted",
		Reserve1:    "1",
		TotalShares: "1",
		LpDenom:     "lp/99",
	}))
	require.NotPanics(t, func() {
		_, err = msgServer.SwapExactAmountIn(ctx, &dextypes.MsgSwapExactAmountIn{
			Trader:        trader.String(),
			PoolId:        99,
			TokenIn:       sdk.NewInt64Coin("uatom", 1),
			TokenOutDenom: "norb",
			MinAmountOut:  "1",
		})
	})
	require.Error(t, err)
}

func TestFeeAndGovernanceAbuseRejected(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-adversarial-4")
	ctx := testutil.NewContext(app, 1)
	feesMsgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)
	dexMsgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)
	tfMsgServer := tfkeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	_, err := feesMsgServer.UpdateParams(ctx, &feestypes.MsgUpdateParams{
		Authority: "orb1unauthorized",
		Params:    feestypes.DefaultParams(),
	})
	require.ErrorIs(t, err, feestypes.ErrUnauthorized)
	feesParams := feestypes.DefaultParams()
	feesParams.AllowedFeeDenoms = []string{"uatom"}
	_, err = feesMsgServer.UpdateParams(ctx, &feestypes.MsgUpdateParams{
		Authority: app.FeesKeeper.Authority(),
		Params:    feesParams,
	})
	require.Error(t, err)

	dexParams := dextypes.DefaultParams()
	dexParams.SwapFeeBps = dexParams.MaxSwapFeeBps + 1
	_, err = dexMsgServer.UpdateParams(ctx, &dextypes.MsgUpdateParams{
		Authority: app.DexKeeper.Authority(),
		Params:    dexParams,
	})
	require.ErrorIs(t, err, dextypes.ErrInvalidParams)

	tfParams := tftypes.DefaultParams()
	tfParams.MinSubdenomLength = tfParams.MaxSubdenomLength + 1
	_, err = tfMsgServer.UpdateParams(ctx, &tftypes.MsgUpdateParams{
		Authority: app.TokenFactoryKeeper.Authority(),
		Params:    tfParams,
	})
	require.ErrorIs(t, err, tftypes.ErrInvalidParams)
}

func TestRepeatedInvalidFeeSpamDoesNotAdvanceProtocolAccounting(t *testing.T) {
	app := testutil.NewInitializedApp(t, "orbitalis-adversarial-5")
	ctx := testutil.NewContext(app, 1)
	before, err := app.FeesKeeper.GetProtocolFeeState(ctx)
	require.NoError(t, err)
	called := 0
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called++
		return ctx, nil
	}

	for i := 0; i < 100; i++ {
		_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin("uatom", 1))}, false)
		require.ErrorIs(t, err, feestypes.ErrInvalidFee)
	}
	after, err := app.FeesKeeper.GetProtocolFeeState(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, called)
	require.Equal(t, before, after)
}

type feeTx struct {
	fees sdk.Coins
}

func (tx feeTx) GetMsgs() []sdk.Msg                    { return nil }
func (tx feeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (tx feeTx) GetGas() uint64                        { return 100_000 }
func (tx feeTx) GetFee() sdk.Coins                     { return tx.fees }
func (tx feeTx) FeePayer() []byte                      { return nil }
func (tx feeTx) FeeGranter() []byte                    { return nil }
