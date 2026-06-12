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

	systemregistrykeeper "github.com/sovereign-l1/l1/x/system-registry/keeper"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
)

func TestSystemRegistrySystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, systemregistrytypes.ModuleName)
	require.Contains(t, app.keys, systemregistrytypes.StoreKey)
	require.Contains(t, genesis, systemregistrytypes.ModuleName)
	require.Contains(t, GetMaccPerms(), systemregistrytypes.ModuleName)
	require.Nil(t, GetMaccPerms()[systemregistrytypes.ModuleName])

	var registryGenesis systemregistrykeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[systemregistrytypes.ModuleName], &registryGenesis))
	require.NoError(t, registryGenesis.Validate())
	require.NotEmpty(t, registryGenesis.State.Entities)
}

func TestSystemRegistrySystemStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	registryGenesis := systemregistrykeeper.DefaultGenesis()
	registryGenesis.State.Entities = append(registryGenesis.State.Entities, systemregistrytypes.SystemEntity{
		ModuleName:		"state-metering",
		ModuleAccountAddress:	"4:0000000000000000000000005555555555555555555555555555555555555555",
		AuthorityAddress:	registryGenesis.Params.Authority,
		Status:			systemregistrytypes.StatusActive,
		Version:		1,
		Dependencies:		[]string{systemregistrytypes.ModuleConstitution},
	})
	registryGenesis.State = registryGenesis.State.Normalize(registryGenesis.Params)
	require.NoError(t, registryGenesis.Validate())
	registryGenesisBytes, err := json.Marshal(registryGenesis)
	require.NoError(t, err)
	genesis[systemregistrytypes.ModuleName] = registryGenesisBytes
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
	exported, err := restarted.SystemRegistryKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	entity, found := exported.State.Entity("state-metering")
	require.True(t, found)
	require.Equal(t, systemregistrytypes.StatusActive, entity.Status)
}
