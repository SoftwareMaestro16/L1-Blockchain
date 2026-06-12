package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	l1app "github.com/sovereign-l1/l1/app"
	emissionskeeper "github.com/sovereign-l1/l1/x/emissions/keeper"
	"github.com/sovereign-l1/l1/x/emissions/types"
)

func TestStakingRatioBelowTargetIncreasesRewardsWithinBounds(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	params := testParams()
	require.NoError(t, app.EmissionsKeeper.SetParams(ctx, params))
	msgServer := emissionskeeper.NewMsgServerImpl(app.EmissionsKeeper)

	res, err := msgServer.FinalizeEmissionEpoch(ctx, &types.MsgFinalizeEmissionEpoch{
		Authority:		app.EmissionsKeeper.Authority(),
		Epoch:			1,
		StakingRatioBps:	5_000,
	})
	require.NoError(t, err)

	require.Equal(t, uint32(400), res.EmissionEpoch.InflationBps)
	require.Equal(t, sdkmath.NewInt(400), res.EmissionEpoch.EmissionAmount.Amount)
	require.Equal(t, sdkmath.NewInt(280), res.EmissionEpoch.ValidatorReward.Amount)
	stored, err := app.EmissionsKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint32(400), stored.CurrentInflationBps)
}

func TestStakingRatioAboveTargetDecreasesRewardsWithinBounds(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	params := testParams()
	require.NoError(t, app.EmissionsKeeper.SetParams(ctx, params))
	msgServer := emissionskeeper.NewMsgServerImpl(app.EmissionsKeeper)

	res, err := msgServer.FinalizeEmissionEpoch(ctx, &types.MsgFinalizeEmissionEpoch{
		Authority:		app.EmissionsKeeper.Authority(),
		Epoch:			1,
		StakingRatioBps:	8_000,
	})
	require.NoError(t, err)

	require.Equal(t, uint32(100), res.EmissionEpoch.InflationBps)
	require.Equal(t, sdkmath.NewInt(100), res.EmissionEpoch.EmissionAmount.Amount)
	require.Equal(t, sdkmath.NewInt(70), res.EmissionEpoch.ValidatorReward.Amount)
}

func TestDistributionWeightsSumValidation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := emissionskeeper.NewMsgServerImpl(app.EmissionsKeeper)
	params := testParams()
	params.DistributionWeights.EcosystemBps = 499

	_, err := msgServer.UpdateEmissionsParams(ctx, &types.MsgUpdateEmissionsParams{
		Authority:	app.EmissionsKeeper.Authority(),
		Params:		params,
	})
	require.ErrorIs(t, err, types.ErrInvalidParams)
}

func TestDeterministicRoundingRemainder(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	params := testParams()
	params.CurrentInflationBps = 10_000
	params.MinAnnualInflationBps = 1
	params.MaxAnnualInflationBps = 10_000
	params.ConstitutionalMaxInflationBps = 10_000
	params.TargetStakingRatioBps = 5_000
	params.ResponsivenessBps = 1_000
	params.AnnualReferenceSupply = sdk.NewInt64Coin(types.BaseDenom, 101)
	params.EpochsPerYear = 1
	params.DistributionWeights = types.DistributionWeights{
		ValidatorRewardBps:	3_333,
		TreasuryBps:		3_333,
		ProtectionBps:		3_333,
		BurnBps:		1,
		EcosystemBps:		0,
	}
	require.NoError(t, app.EmissionsKeeper.SetParams(ctx, params))

	first, err := app.EmissionsKeeper.FinalizeEmissionEpoch(ctx, 1, 5_000)
	require.NoError(t, err)
	second, err := types.ComputeEpochEmission(params, 2, 5_000, 1)
	require.NoError(t, err)

	require.Equal(t, sdkmath.NewInt(101), first.EmissionAmount.Amount)
	require.Equal(t, sdkmath.NewInt(33), first.ValidatorReward.Amount)
	require.Equal(t, sdkmath.NewInt(33), first.Treasury.Amount)
	require.Equal(t, sdkmath.NewInt(33), first.ProtectionFund.Amount)
	require.Equal(t, sdkmath.NewInt(2), first.RoundingRemainder.Amount)
	require.Equal(t, first.ValidatorReward.Amount, second.ValidatorReward.Amount)
	require.Equal(t, first.RoundingRemainder.Amount, second.RoundingRemainder.Amount)
}

func TestExportImportPreservesEmissionEpoch(t *testing.T) {
	source := l1app.Setup(t, false)
	sourceCtx := source.NewContext(false)
	require.NoError(t, source.EmissionsKeeper.SetParams(sourceCtx, testParams()))
	_, err := source.EmissionsKeeper.FinalizeEmissionEpoch(sourceCtx, 7, 5_000)
	require.NoError(t, err)

	exported, err := source.EmissionsKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := l1app.Setup(t, false)
	targetCtx := target.NewContext(false)
	require.NoError(t, target.EmissionsKeeper.InitGenesis(targetCtx, *exported))
	imported, err := target.EmissionsKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func testParams() types.Params {
	params := types.DefaultParams()
	params.CurrentInflationBps = 300
	params.TargetStakingRatioBps = 6_000
	params.MinAnnualInflationBps = 100
	params.MaxAnnualInflationBps = 500
	params.ConstitutionalMaxInflationBps = 500
	params.ResponsivenessBps = 1_000
	params.AnnualReferenceSupply = sdk.NewInt64Coin(types.BaseDenom, 3_650_000)
	params.EpochsPerYear = 365
	params.DistributionWeights = types.DefaultDistributionWeights()
	return params
}
