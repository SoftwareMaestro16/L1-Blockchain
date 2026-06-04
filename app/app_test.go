package app

import (
	"encoding/json"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestOrbitalisChainConstants(t *testing.T) {
	require.Equal(t, "Orbitalis", appName)
	require.Equal(t, "orb", AccountAddressPrefix)
	require.Equal(t, "orbvaloper", ValidatorAddressPrefix)
	require.Equal(t, "orbvalcons", ConsensusAddressPrefix)
	require.Equal(t, appparams.BaseDenom, BondDenom)
	require.Equal(t, appparams.BaseDenom, sdk.DefaultBondDenom)
	require.Equal(t, int64(1_000_000_000), appparams.BaseUnitsPerDisplay)
	require.True(t, strings.HasSuffix(DefaultNodeHome, ".orbitalis"), DefaultNodeHome)
}

func TestDefaultGenesisIncludesNativeTokenMetadata(t *testing.T) {
	app, genesis := setup(true, 5)

	var bankGenState banktypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)

	var native banktypes.Metadata
	for _, metadata := range bankGenState.DenomMetadata {
		if metadata.Base == appparams.BaseDenom {
			native = metadata
			break
		}
	}

	requireNativeTokenMetadata(t, native)
}

func requireNativeTokenMetadata(t *testing.T, native banktypes.Metadata) {
	t.Helper()

	require.NoError(t, native.Validate())
	require.Equal(t, appparams.BaseDenom, native.Base)
	require.Equal(t, appparams.DisplayDenom, native.Display)
	require.Equal(t, appparams.TokenSymbol, native.Symbol)
	require.Equal(t, appparams.TokenName, native.Name)
	requireDenomUnit(t, native, appparams.BaseDenom, 0)
	requireDenomUnit(t, native, appparams.DisplayDenom, appparams.DisplayDenomExponent)
}

func requireDenomUnit(t *testing.T, metadata banktypes.Metadata, denom string, exponent uint32) {
	t.Helper()

	for _, unit := range metadata.DenomUnits {
		if unit.Denom == denom {
			require.Equal(t, exponent, unit.Exponent)
			return
		}
	}
	require.Failf(t, "missing denom unit", "denom %s", denom)
}

func TestDefaultGenesisValidatesAndSetsCustomModuleDefaults(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, []string{appparams.BaseDenom}, feesGenState.Params.AllowedFeeDenoms)
	require.Equal(t, "0.98", feesGenState.Params.ValidatorRewardsRatio)
	require.Equal(t, "0.02", feesGenState.Params.CommunityPoolRatio)

	stakingGenState := stakingtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
	require.Equal(t, appparams.BaseDenom, stakingGenState.Params.BondDenom)

	var mintGenState minttypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenState)
	require.Equal(t, appparams.BaseDenom, mintGenState.Params.MintDenom)

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Empty(t, tokenfactoryGenState.Denoms)

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, dextypes.DefaultNextPoolID, dexGenState.NextPoolId)
	require.Empty(t, dexGenState.Pools)
}

func TestNativeTokenRuntimeMetadataSendAndSupply(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	metadata, found := app.BankKeeper.GetDenomMetaData(ctx, appparams.BaseDenom)
	require.True(t, found)
	requireNativeTokenMetadata(t, metadata)

	addrs := AddTestAddrsIncremental(app, ctx, 2, sdkmath.NewInt(1_000_000))
	sender, recipient := addrs[0], addrs[1]
	beforeSupply := app.BankKeeper.GetSupply(ctx, appparams.BaseDenom)
	sendAmount := sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 123_456))

	require.NoError(t, app.BankKeeper.SendCoins(ctx, sender, recipient, sendAmount))

	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 876_544), app.BankKeeper.GetBalance(ctx, sender, appparams.BaseDenom))
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 1_123_456), app.BankKeeper.GetBalance(ctx, recipient, appparams.BaseDenom))
	require.Equal(t, beforeSupply, app.BankKeeper.GetSupply(ctx, appparams.BaseDenom))
}

func TestCustomModuleGenesisInitExportRoundTrip(t *testing.T) {
	app := Setup(t, false)

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	var exportedGenesis GenesisState
	require.NoError(t, json.Unmarshal(exported.AppState, &exportedGenesis))
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis))

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, feestypes.DefaultGenesisState(), &feesGenState)

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Empty(t, tokenfactoryGenState.Denoms)
	require.NoError(t, tokenfactoryGenState.Validate())

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, dextypes.DefaultNextPoolID, dexGenState.NextPoolId)
	require.Empty(t, dexGenState.Pools)
	require.NoError(t, dexGenState.Validate())
}
