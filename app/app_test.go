package app

import (
	"encoding/json"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	require.Equal(t, "norb", BondDenom)
	require.Equal(t, "norb", sdk.DefaultBondDenom)
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

	require.Equal(t, appparams.NativeTokenMetadata(), native)
	require.NoError(t, native.Validate())
}

func TestDefaultGenesisValidatesAndSetsCustomModuleDefaults(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, []string{appparams.BaseDenom}, feesGenState.Params.AllowedFeeDenoms)
	require.Equal(t, "0.98", feesGenState.Params.ValidatorRewardsRatio)
	require.Equal(t, "0.02", feesGenState.Params.CommunityPoolRatio)

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Empty(t, tokenfactoryGenState.Denoms)

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, dextypes.DefaultNextPoolID, dexGenState.NextPoolId)
	require.Empty(t, dexGenState.Pools)
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
