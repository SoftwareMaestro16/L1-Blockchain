package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func TestGenesisExportImportRoundTrip(t *testing.T) {
	sourceApp := l1app.Setup(t, false)
	sourceCtx := sourceApp.NewContext(false)
	require.NoError(t, sourceApp.FeesKeeper.RecordCollectedFees(sourceCtx, sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1_000))))

	exported, err := sourceApp.FeesKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	targetApp := l1app.Setup(t, false)
	targetCtx := targetApp.NewContext(false)
	require.NoError(t, targetApp.FeesKeeper.InitGenesis(targetCtx, *exported))

	reexported, err := targetApp.FeesKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, reexported)

	again, err := sourceApp.FeesKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.Equal(t, exported, again)
}

func TestInitGenesisRejectsCorruptedStateWithoutPartialWrites(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	before, err := app.FeesKeeper.ExportGenesis(ctx)
	require.NoError(t, err)

	corrupted := types.GenesisState{
		Params:	types.DefaultParams(),
		ProtocolFeeState: types.ProtocolFeeState{
			TotalCollected:		sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 100)),
			ValidatorRewards:	sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 90)),
			CommunityPool:		sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 5)),
		},
	}

	err = app.FeesKeeper.InitGenesis(ctx, corrupted)
	require.Error(t, err)

	after, err := app.FeesKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestMigrationRejectsCorruptedFeeState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	corrupted := types.ProtocolFeeState{
		TotalCollected:		sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 100)),
		ValidatorRewards:	sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 75)),
		CommunityPool:		sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 20)),
	}
	bz, err := app.AppCodec().Marshal(&corrupted)
	require.NoError(t, err)
	ctx.KVStore(app.GetKey(types.StoreKey)).Set(types.ProtocolFeeStateKey, bz)

	err = feeskeeper.NewMigrator(app.FeesKeeper).Migrate1to2(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fee accounting mismatch")
}

func TestFeesMigrationSucceedsOnValidState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, feeskeeper.NewMigrator(app.FeesKeeper).Migrate1to2(ctx))
}

// TestCongestionStateGenesisRoundTrip verifies that congestion state is exported,
// imported, and survives a genesis export→import→re-export round trip.
func TestCongestionStateGenesisRoundTrip(t *testing.T) {
	sourceApp := l1app.Setup(t, false)
	sourceCtx := sourceApp.NewContext(false)

	require.NoError(t, sourceApp.FeesKeeper.SetCongestionState(sourceCtx, 7_500))

	exported, err := sourceApp.FeesKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.Equal(t, uint32(7_500), exported.CongestionBps)

	targetApp := l1app.Setup(t, false)
	targetCtx := targetApp.NewContext(false)
	require.NoError(t, targetApp.FeesKeeper.InitGenesis(targetCtx, *exported))
	require.Equal(t, uint32(7_500), targetApp.FeesKeeper.GetCongestionState(targetCtx))

	reexported, err := targetApp.FeesKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, reexported)
}

// TestCongestionStateDeterminism verifies that SetCongestionState+GetCongestionState
// is deterministic: same input bps yields same output.
func TestCongestionStateDeterminism(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	cases := []uint32{0, 1, 5_000, 7_500, 10_000}
	for _, bps := range cases {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		require.NoError(t, app.FeesKeeper.SetCongestionState(ctx, bps))
		got := app.FeesKeeper.GetCongestionState(ctx)
		require.Equal(t, bps, got, "congestion bps determinism at block %d", ctx.BlockHeight())
	}
}

// TestCongestionStateClampedAtBasisPoints verifies that values above 10000 bps
// are clamped by InitGenesis.
func TestCongestionStateInitGenesisClampsOversizedValue(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	gs := &types.GenesisState{
		Params:			types.DefaultParams(),
		ProtocolFeeState:	types.DefaultProtocolFeeState(),
		CongestionBps:		99_999,
	}
	require.NoError(t, app.FeesKeeper.InitGenesis(ctx, *gs))
	require.Equal(t, uint32(0), app.FeesKeeper.GetCongestionState(ctx))
}
