package integration_test

import (
	"testing"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/app/stakingpolicy"
	testutil "github.com/sovereign-l1/l1/tests/testutil"
)

func TestDirectUserDelegationTxRejectedBeforeValidatorSetUpdates(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetra-integration-pos-updates")
	ctx := testutil.NewContext(app, 1)
	delegatorPriv, delegator := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(20_000_000))
	validator := bondedValidator(t, app, ctx)
	beforePower := validator.GetConsensusPower(sdk.DefaultPowerReduction)

	delegation := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	txBytes := testutil.EncodeSignedTx(
		t,
		app,
		ctx,
		delegatorPriv,
		[]sdk.Msg{stakingtypes.NewMsgDelegate(
			delegator.String(),
			validator.OperatorAddress,
			sdk.NewCoin("naet", delegation),
		)},
		sdk.NewCoins(sdk.NewInt64Coin("naet", 100)),
		250_000,
	)

	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code)
	require.Contains(t, res.TxResults[0].Log, stakingpolicy.DirectUserDelegationDisabledMessage)
	require.Empty(t, res.ValidatorUpdates)
	testutil.Commit(t, app)

	afterCtx := testutil.NewContext(app, 2)
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(t, err)
	updatedValidator, err := app.StakingKeeper.GetValidator(afterCtx, valAddr)
	require.NoError(t, err)
	require.Equal(t, beforePower, updatedValidator.GetConsensusPower(sdk.DefaultPowerReduction))
	_, err = app.StakingKeeper.GetDelegation(afterCtx, delegator, valAddr)
	require.Error(t, err)
}

func TestRejectedDirectUserDelegationDoesNotExportDelegationState(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetra-integration-pos-restart")
	ctx := testutil.NewContext(app, 1)
	delegatorPriv, delegator := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(20_000_000))
	validator := bondedValidator(t, app, ctx)
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(t, err)

	delegation := sdk.TokensFromConsensusPower(3, sdk.DefaultPowerReduction)
	txBytes := testutil.EncodeSignedTx(
		t,
		app,
		ctx,
		delegatorPriv,
		[]sdk.Msg{stakingtypes.NewMsgDelegate(
			delegator.String(),
			validator.OperatorAddress,
			sdk.NewCoin("naet", delegation),
		)},
		sdk.NewCoins(sdk.NewInt64Coin("naet", 100)),
		250_000,
	)

	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code)
	require.Contains(t, res.TxResults[0].Log, stakingpolicy.DirectUserDelegationDisabledMessage)
	testutil.Commit(t, app)

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	restarted := l1app.NewL1App(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		sims.AppOptionsMap{flags.FlagHome: t.TempDir()},
		baseapp.SetChainID(app.ChainID()),
	)
	_, err = restarted.InitChain(&abci.RequestInitChain{
		ChainId:		app.ChainID(),
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	&exported.ConsensusParams,
		AppStateBytes:		exported.AppState,
	})
	require.NoError(t, err)

	restartedCtx := testutil.NewContext(restarted, 1)
	_, err = restarted.StakingKeeper.GetDelegation(restartedCtx, delegator, valAddr)
	require.Error(t, err)

	restartedValidator, err := restarted.StakingKeeper.GetValidator(restartedCtx, valAddr)
	require.NoError(t, err)
	require.Equal(t, validator.Tokens, restartedValidator.Tokens)
}

func bondedValidator(t *testing.T, app *l1app.L1App, ctx sdk.Context) stakingtypes.Validator {
	t.Helper()
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	for _, validator := range validators {
		if validator.Status == stakingtypes.Bonded {
			return validator
		}
	}
	t.Fatal("expected at least one bonded validator")
	return stakingtypes.Validator{}
}
