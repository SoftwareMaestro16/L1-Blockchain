package app

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log/v2"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	configkeeper "github.com/sovereign-l1/l1/x/config/keeper"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
)

func TestConfigSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, configtypes.ModuleName)
	require.Contains(t, app.keys, configtypes.StoreKey)
	require.Contains(t, genesis, configtypes.ModuleName)
	require.Contains(t, GetMaccPerms(), configtypes.ModuleName)
	require.Nil(t, GetMaccPerms()[configtypes.ModuleName])

	var configGenesis configkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[configtypes.ModuleName], &configGenesis))
	require.NoError(t, configGenesis.Validate())
	require.Equal(t, configkeeper.DefaultGenesis().Params.Authority, configGenesis.Params.Authority)
	require.Empty(t, configGenesis.State.Entries)
}

func TestConfigSystemStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	configGenesis := configkeeper.DefaultGenesis()
	configGenesis.State.Entries = []configtypes.ConfigEntry{
		{
			Key:		"runtime/max_validators",
			Value:		"100",
			Owner:		configGenesis.Params.Authority,
			Version:	1,
			UpdatedHeight:	0,
		},
	}
	require.NoError(t, configGenesis.Validate())
	configGenesisBytes, err := json.Marshal(configGenesis)
	require.NoError(t, err)
	genesis[configtypes.ModuleName] = configGenesisBytes
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

	restarted := NewL1App(log.NewNopLogger(), db, true, appOptions)
	restartedCtx := restarted.NewUncachedContext(false, cmtproto.Header{Height: restarted.LastBlockHeight()})
	exported, err := restarted.ConfigKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Entries, 1)
	require.Equal(t, "runtime/max_validators", exported.State.Entries[0].Key)
	require.Equal(t, "100", exported.State.Entries[0].Value)
}

func TestConfigSystemPersistentUpsertExportsFromContext(t *testing.T) {
	app, _ := setup(false, 5)
	ctx := app.NewUncachedContext(false, cmtproto.Header{Height: 1})
	authority := configkeeper.DefaultGenesis().Params.Authority

	entry, err := app.ConfigKeeper.UpsertEntryState(ctx, authority, "runtime/block_gas", "20000000", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), entry.Version)

	exported, err := app.ConfigKeeper.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Len(t, exported.State.Entries, 1)
	require.Equal(t, "runtime/block_gas", exported.State.Entries[0].Key)
}
