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

	validatorregistrykeeper "github.com/sovereign-l1/l1/x/validator-registry/keeper"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestValidatorRegistrySystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, validatorregistrytypes.ModuleName)
	require.Contains(t, app.keys, validatorregistrytypes.StoreKey)
	require.Contains(t, genesis, validatorregistrytypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), validatorregistrytypes.ModuleName)

	var registryGenesis validatorregistrykeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[validatorregistrytypes.ModuleName], &registryGenesis))
	require.NoError(t, registryGenesis.Validate())
}

func TestValidatorRegistrySystemStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	registryGenesis := validatorregistrykeeper.DefaultGenesis()
	registryGenesis.State.Validators = append(registryGenesis.State.Validators, validatorregistrytypes.ValidatorRecord{
		OperatorAddress:	rawAddress("11"),
		ConsensusPublicKey:	"ed25519:app-validator",
		TreasuryAddress:	rawAddress("12"),
		WithdrawalAddress:	rawAddress("13"),
		EmergencyAddress:	rawAddress("14"),
		CommissionPolicy:	validatorregistrytypes.DefaultCommissionPolicy(),
		Status:			validatorregistrytypes.StatusCandidate,
		SelfBond:		validatorregistrytypes.DefaultMinValidatorStake,
		History: []validatorregistrytypes.ValidatorHistoryEvent{
			{Height: 1, Type: validatorregistrytypes.HistoryRegistered, Detail: "genesis"},
		},
	})
	registryGenesis.State = registryGenesis.State.Normalize(registryGenesis.Params)
	require.NoError(t, registryGenesis.Validate())
	registryGenesisBytes, err := json.Marshal(registryGenesis)
	require.NoError(t, err)
	genesis[validatorregistrytypes.ModuleName] = registryGenesisBytes
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
	exported, err := restarted.ValidatorRegistryKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	validator, found := exported.State.Validator(rawAddress("11"))
	require.True(t, found)
	require.Equal(t, "ed25519:app-validator", validator.ConsensusPublicKey)
}

func rawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}
