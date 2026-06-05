package app

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	orbitaladdress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestStateExportImportPreservesPrototypeModuleData(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	admin := AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000_000))[0]

	tokenfactoryMsg := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)
	createDenomRes, err := tokenfactoryMsg.CreateDenom(ctx, &tokenfactorytypes.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "exportgold",
	})
	require.NoError(t, err)
	factoryDenom := createDenomRes.NewTokenDenom

	_, err = tokenfactoryMsg.Mint(ctx, &tokenfactorytypes.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(factoryDenom, 100_000_000),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)

	dexMsg := dexkeeper.NewMsgServerImpl(app.DexKeeper)
	createPoolRes, err := dexMsg.CreatePool(ctx, &dextypes.MsgCreatePool{
		Creator: admin.String(),
		TokenA:  sdk.NewInt64Coin(appparams.BaseDenom, 10_000_000),
		TokenB:  sdk.NewInt64Coin(factoryDenom, 10_000_000),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), createPoolRes.PoolId)

	exportedGenesis, exportedAppState := exportGenesisFromContext(t, app, ctx)
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis))

	adminText := orbitaladdress.FormatAccAddress(admin)
	requireExportedPrototypeState(t, app, exportedGenesis, adminText, factoryDenom)

	imported, _ := setup(false, 5)
	_, err = imported.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   exportedAppState,
	})
	require.NoError(t, err)

	importedCtx := imported.NewContext(false)
	meta, found, err := imported.TokenFactoryKeeper.GetDenom(importedCtx, factoryDenom)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, adminText, meta.Admin)

	pool, found, err := imported.DexKeeper.GetPool(importedCtx, 1)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, factoryDenom, pool.Denom0)
	require.Equal(t, appparams.BaseDenom, pool.Denom1)
	require.Equal(t, "10000000", pool.Reserve0)
	require.Equal(t, "10000000", pool.Reserve1)
	require.Equal(t, "10000000", pool.TotalShares)
	require.Equal(t, "lp/1", pool.LpDenom)

	require.Equal(t, sdk.NewInt64Coin(factoryDenom, 90_000_000), imported.BankKeeper.GetBalance(importedCtx, admin, factoryDenom))
	require.Equal(t, sdk.NewInt64Coin("lp/1", 10_000_000), imported.BankKeeper.GetBalance(importedCtx, admin, "lp/1"))
}

func TestStateImportRejectsCorruptedExportedPrototypeData(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	admin := AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000_000))[0]

	tokenfactoryMsg := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)
	createDenomRes, err := tokenfactoryMsg.CreateDenom(ctx, &tokenfactorytypes.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "corruptgold",
	})
	require.NoError(t, err)
	_, err = tokenfactoryMsg.Mint(ctx, &tokenfactorytypes.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(createDenomRes.NewTokenDenom, 100_000),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)

	dexMsg := dexkeeper.NewMsgServerImpl(app.DexKeeper)
	_, err = dexMsg.CreatePool(ctx, &dextypes.MsgCreatePool{
		Creator: admin.String(),
		TokenA:  sdk.NewInt64Coin(appparams.BaseDenom, 1_000),
		TokenB:  sdk.NewInt64Coin(createDenomRes.NewTokenDenom, 1_000),
	})
	require.NoError(t, err)

	exportedGenesis, _ := exportGenesisFromContext(t, app, ctx)

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[dextypes.ModuleName], &dexGenState)
	require.Len(t, dexGenState.Pools, 1)
	dexGenState.Pools[0].Reserve0 = "not-an-int"
	exportedGenesis[dextypes.ModuleName] = app.AppCodec().MustMarshalJSON(&dexGenState)

	err = app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reserve0")
}

func exportGenesisFromContext(t *testing.T, app *L1App, ctx sdk.Context) (GenesisState, []byte) {
	t.Helper()

	genesis, err := app.ModuleManager.ExportGenesisForModules(ctx, app.AppCodec(), nil)
	require.NoError(t, err)
	appState, err := json.MarshalIndent(genesis, "", "  ")
	require.NoError(t, err)
	return genesis, appState
}

func requireExportedPrototypeState(t *testing.T, app *L1App, genesis GenesisState, admin, factoryDenom string) {
	t.Helper()

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, []string{appparams.BaseDenom}, feesGenState.Params.AllowedFeeDenoms)

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Len(t, tokenfactoryGenState.Denoms, 1)
	require.Equal(t, factoryDenom, tokenfactoryGenState.Denoms[0].Denom)
	require.Equal(t, admin, tokenfactoryGenState.Denoms[0].Admin)

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, uint64(2), dexGenState.NextPoolId)
	require.Len(t, dexGenState.Pools, 1)
	require.Equal(t, factoryDenom, dexGenState.Pools[0].Denom0)
	require.Equal(t, appparams.BaseDenom, dexGenState.Pools[0].Denom1)
	require.Equal(t, "10000000", dexGenState.Pools[0].Reserve0)
	require.Equal(t, "10000000", dexGenState.Pools[0].Reserve1)
	require.Equal(t, "10000000", dexGenState.Pools[0].TotalShares)

	var bankGenState banktypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)
	require.Contains(t, bankGenState.Supply, sdk.NewInt64Coin(factoryDenom, 100_000_000))
	require.Contains(t, bankGenState.Supply, sdk.NewInt64Coin("lp/1", 10_000_000))
}
