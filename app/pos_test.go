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

	validator := GetBondedTestValidator(t, app, ctx)
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

func TestPoSRejectsInvalidDelegations(t *testing.T) {
	tests := []struct {
		name             string
		fundedCoins      sdk.Coins
		delegationAmount sdk.Coin
		validatorAddress string
	}{
		{
			name:             "wrong denom",
			fundedCoins:      sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10_000_000), sdk.NewInt64Coin("uatom", 10_000_000)),
			delegationAmount: sdk.NewInt64Coin("uatom", 5_000_000),
		},
		{
			name:             "insufficient funds",
			fundedCoins:      sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1)),
			delegationAmount: sdk.NewInt64Coin(BondDenom, 5_000_000),
		},
		{
			name:             "invalid validator address",
			fundedCoins:      sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10_000_000)),
			delegationAmount: sdk.NewInt64Coin(BondDenom, 5_000_000),
			validatorAddress: "not-a-validator-address",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := Setup(t, false)
			ctx := app.NewContext(false)
			validator := GetBondedTestValidator(t, app, ctx)
			validatorAddress := validator.OperatorAddress
			if tc.validatorAddress != "" {
				validatorAddress = tc.validatorAddress
			}

			delegator := AddTestAddrsWithCoins(t, app, ctx, 1, tc.fundedCoins)[0]
			msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
			_, err := msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(
				delegator.String(),
				validatorAddress,
				tc.delegationAmount,
			))
			require.Error(t, err)
		})
	}
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

	validator := GetBondedTestValidator(t, app, ctx)
	consAddrBytes, err := validator.GetConsAddr()
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
