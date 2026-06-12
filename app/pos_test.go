package app

import (
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/app/stakingpolicy"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestPoSCreateValidatorWithNaet(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	selfDelegation := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	valAddr, validator := createFundedValidator(t, app, ctx, "phase4-create-validator", selfDelegation)

	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.Equal(t, selfDelegation, validator.Tokens)
	require.Equal(t, sdkmath.OneInt(), validator.MinSelfDelegation)
	require.Equal(t, int64(10), validator.GetConsensusPower(sdk.DefaultPowerReduction))

	delegation, err := app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(valAddr), valAddr)
	require.NoError(t, err)
	require.Equal(t, validator.DelegatorShares, delegation.Shares)
}

func TestPoSDirectUserDelegationMsgRouteRejected(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	validator := GetBondedTestValidator(t, app, ctx)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.False(t, validator.Jailed)

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	require.Equal(t, BondDenom, bondDenom)
	require.Equal(t, appparams.DirectUserDelegationDisabled, directUserDelegationGovernanceValue(t))

	delegation := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	beforeTokens := validator.Tokens
	beforePower := validator.GetConsensusPower(sdk.DefaultPowerReduction)

	msg := stakingtypes.NewMsgDelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, delegation),
	)
	handler := app.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err = handler(ctx, msg)
	require.ErrorContains(t, err, stakingpolicy.DirectUserDelegationDisabledMessage)

	valAddr := parseValidatorAddress(t, app, validator.OperatorAddress)
	updatedValidator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, beforeTokens, updatedValidator.Tokens)
	require.Equal(t, beforePower, updatedValidator.GetConsensusPower(sdk.DefaultPowerReduction))
}

func TestPoSValidatorSelfBondMsgDelegateAllowed(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	selfDelegation := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	valAddr, validator := createFundedValidator(t, app, ctx, "self-bond-operator-path", selfDelegation)
	operator := sdk.AccAddress(valAddr)
	extraSelfBond := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	beforeTokens := validator.Tokens

	msg := stakingtypes.NewMsgDelegate(
		operator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, extraSelfBond),
	)
	handler := app.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err := handler(ctx, msg)
	require.NoError(t, err)

	updatedValidator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, beforeTokens.Add(extraSelfBond), updatedValidator.Tokens)

	delegation, err := app.StakingKeeper.GetDelegation(ctx, operator, valAddr)
	require.NoError(t, err)
	require.True(t, delegation.Shares.IsPositive())
}

func TestPoSOfficialPoolDepositPathWorksWhileDirectDelegationDisabled(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())
	require.Equal(t, appparams.DirectUserDelegationDisabled, directUserDelegationGovernanceValue(t))

	poolID := "pos-official-pool"
	contractRaw := posRawAddress("66")
	pool, err := app.NominatorPoolKeeper.CreateOfficialLiquidStakingPool(nominatorpooltypes.MsgCreateOfficialLiquidStakingPool{
		Authority:		nominatorpooltypes.DefaultParams().Authority,
		PoolID:			poolID,
		ContractAddressUser:	aeFromRawForPoSTest(t, contractRaw),
		ContractAddressRaw:	contractRaw,
		PoolOperator:		posRawAddress("11"),
		PoolCommissionBps:	100,
		Height:			1,
	})
	require.NoError(t, err)
	user := aeFromRawForPoSTest(t, posRawAddress("22"))

	msg := &nominatorpooltypes.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		nominatorpooltypes.DefaultMinPoolDeposit,
		Height:		2,
	}
	handler := app.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err = handler(ctx, msg)
	require.NoError(t, err)

	query, found := app.NominatorPoolKeeper.PoolShare(nominatorpooltypes.QueryPoolShareRequest{
		PoolID:		pool.PoolID,
		Delegator:	posRawAddress("22"),
	})
	require.True(t, found)
	require.Equal(t, nominatorpooltypes.DefaultMinPoolDeposit, query.Share.Shares)
	require.Zero(t, query.PendingRewards)
}

func TestPoSDirectUserUnbondingMsgRouteRejected(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	validator := GetBondedTestValidator(t, app, ctx)
	delegation := sdk.TokensFromConsensusPower(4, sdk.DefaultPowerReduction)
	unbond := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	delegateStakeFixture(t, app, ctx, delegator, validator, delegation)
	balanceBeforeUnbond := app.BankKeeper.GetBalance(ctx, delegator, BondDenom)

	msg := stakingtypes.NewMsgUndelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, unbond),
	)
	handler := app.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err := handler(ctx, msg)
	require.ErrorContains(t, err, stakingpolicy.DirectUserDelegationDisabledMessage)

	valAddr := parseValidatorAddress(t, app, validator.OperatorAddress)
	remaining, err := app.StakingKeeper.GetDelegation(ctx, delegator, valAddr)
	require.NoError(t, err)
	require.True(t, remaining.Shares.IsPositive())

	_, err = app.StakingKeeper.GetUnbondingDelegation(ctx, delegator, valAddr)
	require.Error(t, err)
	require.Equal(t, balanceBeforeUnbond, app.BankKeeper.GetBalance(ctx, delegator, BondDenom))
}

func TestPoSDirectUserRedelegationMsgRouteRejected(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	source := GetBondedTestValidator(t, app, ctx)
	dstValAddr, destination := createFundedValidator(t, app, ctx, "phase4-redelegate-dst", sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))
	require.Equal(t, formatValidatorAddress(t, app, dstValAddr), destination.OperatorAddress)

	delegation := sdk.TokensFromConsensusPower(4, sdk.DefaultPowerReduction)
	redelegate := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	sourceOperator := sdk.AccAddress(parseValidatorAddress(t, app, source.OperatorAddress)).String()
	destinationOperator := sdk.AccAddress(dstValAddr).String()
	delegator := sdk.AccAddress(nil)
	for _, candidate := range AddTestAddrsIncremental(app, ctx, 4, delegation.MulRaw(2)) {
		if candidate.String() != sourceOperator && candidate.String() != destinationOperator {
			delegator = candidate
			break
		}
	}
	require.NotNil(t, delegator)

	delegateStakeFixture(t, app, ctx, delegator, source, delegation)

	msg := stakingtypes.NewMsgBeginRedelegate(
		delegator.String(),
		source.OperatorAddress,
		destination.OperatorAddress,
		sdk.NewCoin(BondDenom, redelegate),
	)
	handler := app.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err := handler(ctx, msg)
	require.ErrorContains(t, err, stakingpolicy.DirectUserDelegationDisabledMessage)

	srcValAddr := parseValidatorAddress(t, app, source.OperatorAddress)
	sourceDelegation, err := app.StakingKeeper.GetDelegation(ctx, delegator, srcValAddr)
	require.NoError(t, err)
	require.True(t, sourceDelegation.Shares.IsPositive())

	_, err = app.StakingKeeper.GetDelegation(ctx, delegator, dstValAddr)
	require.Error(t, err)

	_, err = app.StakingKeeper.GetRedelegation(ctx, delegator, srcValAddr, dstValAddr)
	require.Error(t, err)
}

func TestPoSDirectUserDelegationRejectsBeforeSDKValidation(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	validator := GetBondedTestValidator(t, app, ctx)
	delegator := AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10_000_000), sdk.NewInt64Coin("uatom", 10_000_000)))[0]
	msg := stakingtypes.NewMsgDelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewInt64Coin("uatom", 5_000_000),
	)

	handler := app.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err := handler(ctx, msg)
	require.ErrorContains(t, err, stakingpolicy.DirectUserDelegationDisabledMessage)
}

func TestSlashingParamsAndSigningInfoRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	params, err := app.SlashingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Positive(t, params.SignedBlocksWindow)
	require.True(t, params.MinSignedPerWindow.IsPositive())
	require.Equal(t, appparams.BpsToLegacyDec(appparams.DoubleSignSlashDefaultBps), params.SlashFractionDoubleSign)
	require.True(t, params.SlashFractionDoubleSign.GTE(appparams.BpsToLegacyDec(appparams.DoubleSignSlashMinBps)))
	require.True(t, params.SlashFractionDoubleSign.LTE(appparams.BpsToLegacyDec(appparams.DoubleSignSlashMaxBps)))
	require.Equal(t, appparams.BpsToLegacyDec(appparams.DowntimeFirstSlashDefaultBps), params.SlashFractionDowntime)
	require.True(t, params.SlashFractionDowntime.GTE(appparams.BpsToLegacyDec(appparams.DowntimeFirstSlashMinBps)))
	require.True(t, params.SlashFractionDowntime.LTE(appparams.BpsToLegacyDec(appparams.DowntimeFirstSlashMaxBps)))
	require.Equal(t, time.Duration(appparams.DowntimeFirstJailDefaultMinutes)*time.Minute, params.DowntimeJailDuration)

	validator := GetBondedTestValidator(t, app, ctx)
	consAddrBytes, err := validator.GetConsAddr()
	require.NoError(t, err)
	consAddr := sdk.ConsAddress(consAddrBytes)
	expected := slashingtypes.NewValidatorSigningInfo(consAddr, 7, 3, time.Unix(0, 0).UTC(), true, 2)

	require.NoError(t, app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, expected))
	actual, err := app.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
	require.True(t, actual.Tombstoned)

	require.NoError(t, app.SlashingKeeper.SetMissedBlockBitmapValue(ctx, consAddr, 5, true))
	missed, err := app.SlashingKeeper.GetMissedBlockBitmapValue(ctx, consAddr, 5)
	require.NoError(t, err)
	require.True(t, missed)

	require.NoError(t, app.SlashingKeeper.SetMissedBlockBitmapValue(ctx, consAddr, 5, false))
	missed, err = app.SlashingKeeper.GetMissedBlockBitmapValue(ctx, consAddr, 5)
	require.NoError(t, err)
	require.False(t, missed)
}

func TestStakingRewardsDistributionCanBeWithdrawn(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	validator := GetBondedTestValidator(t, app, ctx)
	valAddr := parseValidatorAddress(t, app, validator.OperatorAddress)
	delegation := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	delegateStakeFixture(t, app, ctx, delegator, validator, delegation)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(time.Second))
	updatedValidator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	distrMsgServer := distrkeeper.NewMsgServerImpl(app.DistrKeeper)
	depositor := AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	_, err = distrMsgServer.DepositValidatorRewardsPool(ctx, distrtypes.NewMsgDepositValidatorRewardsPool(
		depositor.String(),
		updatedValidator.OperatorAddress,
		sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 100_000)),
	))
	require.NoError(t, err)

	balanceBefore := app.BankKeeper.GetBalance(ctx, delegator, BondDenom)
	_, err = distrMsgServer.WithdrawDelegatorReward(ctx, distrtypes.NewMsgWithdrawDelegatorReward(
		delegator.String(),
		validator.OperatorAddress,
	))
	require.NoError(t, err)

	balanceAfter := app.BankKeeper.GetBalance(ctx, delegator, BondDenom)
	require.True(t, balanceAfter.Amount.GT(balanceBefore.Amount), "delegator must receive naet staking rewards")
}

func TestPoSMintPolicyIsNaetAndUncappedWithBoundedInflation(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	params, err := app.MintKeeper.Params.Get(ctx)
	require.NoError(t, err)
	expected := appparams.AetraMintParams()
	require.Equal(t, BondDenom, params.MintDenom)
	require.True(t, params.MaxSupply.IsZero(), "zero max supply means uncapped PoS issuance in Cosmos SDK mint params")
	require.NoError(t, params.Validate())
	require.Equal(t, expected.InflationRateChange, params.InflationRateChange)
	require.Equal(t, expected.InflationMin, params.InflationMin)
	require.Equal(t, expected.InflationMax, params.InflationMax)
	require.Equal(t, expected.GoalBonded, params.GoalBonded)
	require.Positive(t, params.BlocksPerYear)

	minter, err := app.MintKeeper.Minter.Get(ctx)
	require.NoError(t, err)
	require.True(t, minter.Inflation.GTE(params.InflationMin))
	require.True(t, minter.Inflation.LTE(params.InflationMax))
}

func TestAddTestAddrsUsesBondDenom(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	addr := AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(123))[0]
	require.Equal(t, sdk.NewInt64Coin(BondDenom, 123), app.BankKeeper.GetBalance(ctx, addr, BondDenom))
}

func createFundedValidator(t *testing.T, app *L1App, ctx sdk.Context, moniker string, selfDelegation sdkmath.Int) (sdk.ValAddress, stakingtypes.Validator) {
	t.Helper()
	operator := AddTestAddrsIncremental(app, ctx, 1, selfDelegation.MulRaw(2))[0]
	valAddr := sdk.ValAddress(operator)
	valText := formatValidatorAddress(t, app, valAddr)
	msg, err := stakingtypes.NewMsgCreateValidator(
		valText,
		ed25519.GenPrivKey().PubKey(),
		sdk.NewCoin(BondDenom, selfDelegation),
		stakingtypes.Description{Moniker: moniker},
		stakingtypes.NewCommissionRates(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
		sdkmath.OneInt(),
	)
	require.NoError(t, err)

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(t, err)

	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)

	validator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	return valAddr, validator
}

func delegateStakeFixture(t *testing.T, app *L1App, ctx sdk.Context, delegator sdk.AccAddress, validator stakingtypes.Validator, amount sdkmath.Int) {
	t.Helper()

	_, err := app.StakingKeeper.Delegate(ctx, delegator, amount, stakingtypes.Unbonded, validator, true)
	require.NoError(t, err)
}

func parseValidatorAddress(t *testing.T, app *L1App, text string) sdk.ValAddress {
	t.Helper()

	bz, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(text)
	require.NoError(t, err)
	return sdk.ValAddress(bz)
}

func formatValidatorAddress(t *testing.T, app *L1App, addr sdk.ValAddress) string {
	t.Helper()

	text, err := app.StakingKeeper.ValidatorAddressCodec().BytesToString(addr)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(text, ValidatorAddressPrefix), text)
	require.NotRegexp(t, `^[a-z]+1`, text)
	return text
}

func directUserDelegationGovernanceValue(t *testing.T) string {
	t.Helper()
	for _, value := range appparams.DefaultGovernanceGenesisParams() {
		if value.Key == appparams.GovernanceParamDirectUserDelegation {
			return value.StringValue
		}
	}
	t.Fatalf("%s missing from default governance genesis params", appparams.GovernanceParamDirectUserDelegation)
	return ""
}

func posRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + strings.Repeat(hexByte, 20)
}

func aeFromRawForPoSTest(t *testing.T, raw string) string {
	t.Helper()
	bz, err := addressing.Parse(raw)
	require.NoError(t, err)
	text, err := addressing.FormatUserFriendly(bz)
	require.NoError(t, err)
	return text
}
