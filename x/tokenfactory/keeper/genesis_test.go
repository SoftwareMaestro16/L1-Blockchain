package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestGenesisExportImportRoundTrip(t *testing.T) {
	sourceApp := l1app.Setup(t, false)
	sourceCtx := sourceApp.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(sourceApp, sourceCtx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(sourceApp.TokenFactoryKeeper)

	_, err := msgServer.CreateDenom(sourceCtx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "gold",
	})
	require.NoError(t, err)

	exported, err := sourceApp.TokenFactoryKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	targetApp := l1app.Setup(t, false)
	targetCtx := targetApp.NewContext(false)
	require.NoError(t, targetApp.TokenFactoryKeeper.InitGenesis(targetCtx, *exported))

	reexported, err := targetApp.TokenFactoryKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, reexported)

	again, err := sourceApp.TokenFactoryKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.Equal(t, exported, again)
}

func TestInitGenesisRejectsCorruptedStateWithoutWrites(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	corrupted := types.GenesisState{
		Params: types.DefaultParams(),
		Denoms: []types.DenomAuthorityMetadata{
			{Denom: "factory/not-an-address/gold", Admin: "not-an-address"},
		},
	}

	err := app.TokenFactoryKeeper.InitGenesis(ctx, corrupted)
	require.Error(t, err)

	_, found, err := app.TokenFactoryKeeper.GetDenom(ctx, "factory/not-an-address/gold")
	require.NoError(t, err)
	require.False(t, found)
}

func TestMigrationRejectsCorruptedDenomMetadata(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.TokenFactoryKeeper.SetDenom(ctx, types.DenomAuthorityMetadata{
		Denom: "factory/not-an-address/gold",
		Admin: "not-an-address",
	}))

	err := tokenfactorykeeper.NewMigrator(app.TokenFactoryKeeper).Migrate1to2(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid admin")
}

func TestTokenfactoryMigrationSucceedsOnValidState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, tokenfactorykeeper.NewMigrator(app.TokenFactoryKeeper).Migrate1to2(ctx))
}
