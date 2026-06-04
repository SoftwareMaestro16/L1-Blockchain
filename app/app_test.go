package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log/v2"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/observability"
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

func TestAppGenesisExportImportRoundTripAndDeterminism(t *testing.T) {
	source, _ := setup(true, 5)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	_, err = source.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	_, err = source.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   source.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = source.Commit()
	require.NoError(t, err)

	exportedA, err := source.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	exportedB, err := source.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	require.Equal(t, exportedA.AppState, exportedB.AppState)

	var exportedState GenesisState
	require.NoError(t, json.Unmarshal(exportedA.AppState, &exportedState))
	require.NoError(t, source.BasicModuleManager.ValidateGenesis(source.AppCodec(), source.TxConfig(), exportedState))

	target := NewL1App(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome},
	)
	_, err = target.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: &exportedA.ConsensusParams,
		AppStateBytes:   exportedA.AppState,
	})
	require.NoError(t, err)
	_, err = target.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   target.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = target.Commit()
	require.NoError(t, err)

	reexported, err := target.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	var reexportedState GenesisState
	require.NoError(t, json.Unmarshal(reexported.AppState, &reexportedState))
	require.NoError(t, target.BasicModuleManager.ValidateGenesis(target.AppCodec(), target.TxConfig(), reexportedState))
}

func TestTelemetryDoesNotChangeAppHash(t *testing.T) {
	source, _ := setup(true, 5)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	t.Cleanup(func() {
		observability.ResetForTesting()
	})

	enabledHash := runSingleBlockForTelemetryTest(t, stateBytes, true)
	disabledHash := runSingleBlockForTelemetryTest(t, stateBytes, false)

	require.Equal(t, enabledHash, disabledHash)
}

func TestCustomModuleMigrationsFromV1ToCurrent(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	fromVM := app.ModuleManager.GetVersionMap()
	fromVM[feestypes.ModuleName] = 1
	fromVM[tokenfactorytypes.ModuleName] = 1
	fromVM[dextypes.ModuleName] = 1

	updated, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	require.NoError(t, err)
	require.Equal(t, uint64(2), updated[feestypes.ModuleName])
	require.Equal(t, uint64(2), updated[tokenfactorytypes.ModuleName])
	require.Equal(t, uint64(2), updated[dextypes.ModuleName])
}

func runSingleBlockForTelemetryTest(t *testing.T, stateBytes []byte, telemetryEnabled bool) []byte {
	t.Helper()
	observability.ResetForTesting()
	observability.SetEnabled(telemetryEnabled)
	app := NewL1App(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome},
	)
	_, err := app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   app.LastCommitID().Hash,
		Time:   time.Unix(1_700_000_000, 0).UTC(),
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	return app.LastCommitID().Hash
}
