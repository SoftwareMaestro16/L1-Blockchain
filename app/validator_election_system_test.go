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

	validatorelectionkeeper "github.com/sovereign-l1/l1/x/validator-election/keeper"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestValidatorElectionSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetherCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, validatorelectiontypes.ModuleName)
	require.Contains(t, app.keys, validatorelectiontypes.StoreKey)
	require.Contains(t, genesis, validatorelectiontypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), validatorelectiontypes.ModuleName)

	var electionGenesis validatorelectionkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[validatorelectiontypes.ModuleName], &electionGenesis))
	require.NoError(t, electionGenesis.Validate())
}

func TestValidatorElectionStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	electionGenesis := validatorelectionkeeper.DefaultGenesis()
	electionGenesis.State.CurrentValidatorSet = []validatorelectiontypes.ValidatorPower{{
		OperatorAddress:    rawElectionAddress("11"),
		ConsensusPublicKey: "ed25519:app-election",
		VotingPower:        100,
		ValidatorStatus:    validatorregistrytypes.StatusActive,
	}}
	electionGenesis.State = electionGenesis.State.Normalize(electionGenesis.Params)
	require.NoError(t, electionGenesis.Validate())
	electionGenesisBytes, err := json.Marshal(electionGenesis)
	require.NoError(t, err)
	genesis[validatorelectiontypes.ModuleName] = electionGenesisBytes
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

	restarted := NewL1App(log.NewNopLogger(), db, true, appOptions)
	restartedCtx := restarted.NewUncachedContext(false, cmtproto.Header{Height: restarted.LastBlockHeight()})
	exported, err := restarted.ValidatorElectionKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.CurrentValidatorSet, 1)
	require.Equal(t, "ed25519:app-election", exported.State.CurrentValidatorSet[0].ConsensusPublicKey)
}

func rawElectionAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}
