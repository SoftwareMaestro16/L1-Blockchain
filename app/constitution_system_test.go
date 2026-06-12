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

	constitutionkeeper "github.com/sovereign-l1/l1/x/constitution/keeper"
	constitutiontypes "github.com/sovereign-l1/l1/x/constitution/types"
)

func TestConstitutionSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, constitutiontypes.ModuleName)
	require.Contains(t, app.keys, constitutiontypes.StoreKey)
	require.Contains(t, genesis, constitutiontypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), constitutiontypes.ModuleName)

	var constitutionGenesis constitutionkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[constitutiontypes.ModuleName], &constitutionGenesis))
	require.NoError(t, constitutionGenesis.Validate())
	require.NotEmpty(t, constitutionGenesis.State.Constitution.ProtectedModules)
}

func TestConstitutionSystemStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	constitutionGenesis := constitutionkeeper.DefaultGenesis()
	constitutionGenesis.State.Constitution.MaxBlockGas = 123456789
	require.NoError(t, constitutionGenesis.Validate())
	constitutionGenesisBytes, err := json.Marshal(constitutionGenesis)
	require.NoError(t, err)
	genesis[constitutiontypes.ModuleName] = constitutionGenesisBytes
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
	exported, err := restarted.ConstitutionKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Equal(t, uint64(123456789), exported.State.Constitution.MaxBlockGas)
}
