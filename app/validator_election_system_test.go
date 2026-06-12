package app

import (
	"bytes"
	"encoding/hex"
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
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, validatorelectiontypes.ModuleName)
	require.Contains(t, app.keys, validatorelectiontypes.StoreKey)
	require.Contains(t, genesis, validatorelectiontypes.ModuleName)
	require.Contains(t, GetMaccPerms(), validatorelectiontypes.ModuleName)
	require.Nil(t, GetMaccPerms()[validatorelectiontypes.ModuleName])

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
	consensusKey := electionConsensusKeyHex(0x31)
	electionGenesis.State.CurrentValidatorSet = []validatorelectiontypes.ValidatorPower{{
		OperatorAddress:	rawElectionAddress("11"),
		ConsensusPublicKey:	consensusKey,
		VotingPower:		100,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	}}
	electionGenesis.State = electionGenesis.State.Normalize(electionGenesis.Params)
	require.NoError(t, electionGenesis.Validate())
	electionGenesisBytes, err := json.Marshal(electionGenesis)
	require.NoError(t, err)
	genesis[validatorelectiontypes.ModuleName] = electionGenesisBytes
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
	exported, err := restarted.ValidatorElectionKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.CurrentValidatorSet, 1)
	require.Equal(t, consensusKey, exported.State.CurrentValidatorSet[0].ConsensusPublicKey)
}

func TestValidatorElectionCurrentSetControlsFinalizeBlockValidatorUpdates(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	app := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, app)
	electionKey := electionConsensusKeyHex(0x42)

	electionGenesis := validatorelectionkeeper.DefaultGenesis()
	electionGenesis.State.CurrentValidatorSet = []validatorelectiontypes.ValidatorPower{{
		OperatorAddress:	rawElectionAddress("42"),
		ConsensusPublicKey:	electionKey,
		VotingPower:		77,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	}}
	electionGenesis.State = electionGenesis.State.Normalize(electionGenesis.Params)
	require.NoError(t, electionGenesis.Validate())
	electionGenesisBytes, err := json.Marshal(electionGenesis)
	require.NoError(t, err)
	genesis[validatorelectiontypes.ModuleName] = electionGenesisBytes
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	1,
		Hash:	app.LastCommitID().Hash,
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.ValidatorUpdates)

	var positive []abci.ValidatorUpdate
	hasStakingRemoval := false
	for _, update := range res.ValidatorUpdates {
		if update.Power > 0 {
			positive = append(positive, update)
		}
		if update.Power == 0 {
			hasStakingRemoval = true
		}
	}
	require.Len(t, positive, 1)
	require.Equal(t, int64(77), positive[0].Power)
	require.Equal(t, bytes.Repeat([]byte{0x42}, 32), positive[0].PubKey.GetEd25519())
	require.True(t, hasStakingRemoval, "staking validator must be removed when absent from elector current set")
}

func TestValidatorElectionMalformedConsensusKeyRejectsFinalizeBlock(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	app := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, app)

	electionGenesis := validatorelectionkeeper.DefaultGenesis()
	electionGenesis.State.CurrentValidatorSet = []validatorelectiontypes.ValidatorPower{{
		OperatorAddress:	rawElectionAddress("55"),
		ConsensusPublicKey:	"ed25519:not-a-32-byte-key",
		VotingPower:		1,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	}}
	electionGenesis.State = electionGenesis.State.Normalize(electionGenesis.Params)
	require.NoError(t, electionGenesis.Validate())
	electionGenesisBytes, err := json.Marshal(electionGenesis)
	require.NoError(t, err)
	genesis[validatorelectiontypes.ModuleName] = electionGenesisBytes
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
	require.ErrorContains(t, err, "ed25519 public key must be exactly 32 bytes")
}

func rawElectionAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}

func electionConsensusKeyHex(fill byte) string {
	return "ed25519:" + hex.EncodeToString(bytes.Repeat([]byte{fill}, 32))
}
