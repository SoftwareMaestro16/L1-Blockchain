package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	"github.com/sovereign-l1/l1/x/dex/types"
)

func TestGenesisExportImportRoundTrip(t *testing.T) {
	sourceApp, sourceCtx, _, _, _ := setupDexPool(t)

	exported, err := sourceApp.DexKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	targetApp := l1app.Setup(t, false)
	targetCtx := targetApp.NewContext(false)
	require.NoError(t, targetApp.DexKeeper.InitGenesis(targetCtx, *exported))

	reexported, err := targetApp.DexKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, reexported)

	again, err := sourceApp.DexKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.Equal(t, exported, again)
}

func TestInitGenesisRejectsCorruptedStateWithoutWrites(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	corrupted := types.GenesisState{
		NextPoolId: 2,
		Params:     types.DefaultParams(),
		Pools: []types.Pool{
			{
				Id:          1,
				Denom0:      "norb",
				Denom1:      "uatom",
				Reserve0:    "not-an-int",
				Reserve1:    "100",
				TotalShares: "100",
				LpDenom:     "lp/1",
			},
		},
	}

	err := app.DexKeeper.InitGenesis(ctx, corrupted)
	require.Error(t, err)

	_, found, err := app.DexKeeper.GetPool(ctx, 1)
	require.NoError(t, err)
	require.False(t, found)
}

func TestMigrationRejectsCorruptedPoolState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
		Id:          77,
		Denom0:      "norb",
		Denom1:      "uatom",
		Reserve0:    "0",
		Reserve1:    "100",
		TotalShares: "100",
		LpDenom:     "lp/77",
	}))
	require.NoError(t, app.DexKeeper.SetNextPoolID(ctx, 78))

	err := dexkeeper.NewMigrator(app.DexKeeper).Migrate1to2(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reserve0")
}

func TestInitGenesisRejectsDuplicatePoolID(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	pools := []types.Pool{
		{
			Id:          1,
			Denom0:      "norb",
			Denom1:      "uatom",
			Reserve0:    "100",
			Reserve1:    "100",
			TotalShares: "100",
			LpDenom:     "lp/1",
		},
		{
			Id:          1,
			Denom0:      "factory/orb1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp3h70a/gold",
			Denom1:      "norb",
			Reserve0:    "100",
			Reserve1:    "100",
			TotalShares: "100",
			LpDenom:     "lp/1",
		},
	}

	err := app.DexKeeper.InitGenesis(ctx, types.GenesisState{NextPoolId: 3, Params: types.DefaultParams(), Pools: pools})
	require.Error(t, err)
}

func TestDexMigrationSucceedsOnValidState(t *testing.T) {
	app, ctx, _, _, _ := setupDexPool(t)
	require.NoError(t, dexkeeper.NewMigrator(app.DexKeeper).Migrate1to2(ctx))
}
