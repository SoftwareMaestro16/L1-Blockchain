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

	schedulerkeeper "github.com/sovereign-l1/l1/x/scheduler/keeper"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
)

func TestSchedulerSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, schedulertypes.ModuleName)
	require.Contains(t, app.keys, schedulertypes.StoreKey)
	require.Contains(t, genesis, schedulertypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), schedulertypes.ModuleName)

	var schedulerGenesis schedulerkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[schedulertypes.ModuleName], &schedulerGenesis))
	require.NoError(t, schedulerGenesis.Validate())
	require.False(t, schedulerGenesis.Params.Enabled)
	require.Empty(t, schedulerGenesis.State.Jobs)
	require.Empty(t, schedulerGenesis.State.History)
}

func TestSchedulerSystemStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	schedulerGenesis := schedulerkeeper.DefaultGenesis()
	schedulerGenesis.State.Jobs = []schedulertypes.ScheduledJob{
		{
			ID:			"rent-collection",
			OwnerModule:		"aetracore",
			Type:			schedulertypes.JobTypePeriodic,
			NextExecutionHeight:	10,
			Interval:		10,
			MaxGas:			100,
		},
	}
	require.NoError(t, schedulerGenesis.Validate())
	schedulerGenesisBytes, err := json.Marshal(schedulerGenesis)
	require.NoError(t, err)
	genesis[schedulertypes.ModuleName] = schedulerGenesisBytes
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
	exported, err := restarted.SchedulerKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Jobs, 1)
	require.Equal(t, "rent-collection", exported.State.Jobs[0].ID)
	require.Equal(t, schedulertypes.JobTypePeriodic, exported.State.Jobs[0].Type)
}
