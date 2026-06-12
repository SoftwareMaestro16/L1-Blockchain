package app

import (
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log/v2"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/observability"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestDefaultGenesisInitExportValidateAcceptanceChain(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis
	genesis = GenesisStateWithSingleValidator(t, app)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	1,
		Hash:	app.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	exportedA, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	exportedB, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	require.Equal(t, exportedA.AppState, exportedB.AppState)

	var exportedGenesis GenesisState
	require.NoError(t, json.Unmarshal(exportedA.AppState, &exportedGenesis))
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis))
}

func TestAppGenesisExportImportRoundTripAndDeterminism(t *testing.T) {
	source, _ := setup(true, 5)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	_, err = source.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)
	_, err = source.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	1,
		Hash:	source.LastCommitID().Hash,
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
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	&exportedA.ConsensusParams,
		AppStateBytes:		exportedA.AppState,
	})
	require.NoError(t, err)
	_, err = target.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	1,
		Hash:	target.LastCommitID().Hash,
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
	fromVM[loadtypes.ModuleName] = 1
	fromVM[routingtypes.ModuleName] = 1
	fromVM[zonestypes.ModuleName] = 1
	fromVM[meshtypes.ModuleName] = 1

	updated, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	require.NoError(t, err)
	require.Equal(t, uint64(2), updated[feestypes.ModuleName])
	require.Equal(t, uint64(2), updated[loadtypes.ModuleName])
	require.Equal(t, uint64(2), updated[routingtypes.ModuleName])
	require.Equal(t, uint64(2), updated[zonestypes.ModuleName])
	require.Equal(t, uint64(2), updated[meshtypes.ModuleName])
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
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	1,
		Hash:	app.LastCommitID().Hash,
		Time:	time.Unix(1_700_000_000, 0).UTC(),
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	return app.LastCommitID().Hash
}
