package app

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestPoSDelegationUpdatesValidatorPower(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Len(t, validators, 1)

	validator := validators[0]
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.False(t, validator.Jailed)

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	require.Equal(t, BondDenom, bondDenom)

	delegation := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	beforeTokens := validator.Tokens
	beforePower := validator.GetConsensusPower(sdk.DefaultPowerReduction)

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	_, err = msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, delegation),
	))
	require.NoError(t, err)

	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(t, err)
	updatedValidator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, beforeTokens.Add(delegation), updatedValidator.Tokens)
	require.Equal(t, beforePower+5, updatedValidator.GetConsensusPower(sdk.DefaultPowerReduction))

	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, updates)

	var found bool
	for _, update := range updates {
		if update.Power == updatedValidator.GetConsensusPower(sdk.DefaultPowerReduction) {
			found = true
			break
		}
	}
	require.True(t, found, "expected validator-set update with new voting power")
}

func TestSlashingParamsAndSigningInfoRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	params, err := app.SlashingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Positive(t, params.SignedBlocksWindow)
	require.True(t, params.MinSignedPerWindow.IsPositive())
	require.True(t, params.SlashFractionDoubleSign.IsPositive())
	require.True(t, params.SlashFractionDowntime.IsPositive())

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)

	consAddrBytes, err := validators[0].GetConsAddr()
	require.NoError(t, err)
	consAddr := sdk.ConsAddress(consAddrBytes)
	expected := slashingtypes.NewValidatorSigningInfo(consAddr, 7, 3, time.Unix(0, 0).UTC(), false, 2)

	require.NoError(t, app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, expected))
	actual, err := app.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestAddTestAddrsUsesBondDenom(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	addr := AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(123))[0]
	require.Equal(t, sdk.NewInt64Coin(BondDenom, 123), app.BankKeeper.GetBalance(ctx, addr, BondDenom))
}
