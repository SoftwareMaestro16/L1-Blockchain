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

	validatorinsurancekeeper "github.com/sovereign-l1/l1/x/validator-insurance/keeper"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
)

func TestValidatorInsuranceSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, validatorinsurancetypes.ModuleName)
	require.Contains(t, app.keys, validatorinsurancetypes.StoreKey)
	require.Contains(t, genesis, validatorinsurancetypes.ModuleName)
	require.Contains(t, GetMaccPerms(), validatorinsurancetypes.ModuleName)
	require.Nil(t, GetMaccPerms()[validatorinsurancetypes.ModuleName])

	var insuranceGenesis validatorinsurancekeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[validatorinsurancetypes.ModuleName], &insuranceGenesis))
	require.NoError(t, insuranceGenesis.Validate())
}

func TestValidatorInsurancePendingClaimSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	insuranceGenesis := validatorinsurancekeeper.DefaultGenesis()
	insuranceGenesis.State.Insurances = []validatorinsurancetypes.ValidatorInsurance{{
		ValidatorAddress:	validatorInsuranceRawAddress("11"),
		Balance:		1_500,
		ValidatorStatus:	"candidate",
	}}
	insuranceGenesis.State.Claims = []validatorinsurancetypes.InsuranceClaim{{
		ClaimID:		"claim-1",
		ValidatorAddress:	validatorInsuranceRawAddress("11"),
		Claimant:		validatorInsuranceRawAddress("22"),
		Amount:			700,
		Status:			validatorinsurancetypes.ClaimStatusPending,
		Reason:			"pending review",
		SubmittedHeight:	1,
	}}
	insuranceGenesis.State = insuranceGenesis.State.Normalize(insuranceGenesis.Params)
	require.NoError(t, insuranceGenesis.Validate())
	insuranceGenesisBytes, err := json.Marshal(insuranceGenesis)
	require.NoError(t, err)
	genesis[validatorinsurancetypes.ModuleName] = insuranceGenesisBytes
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
	exported, err := restarted.ValidatorInsuranceKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Insurances, 1)
	require.Len(t, exported.State.Claims, 1)
	require.Equal(t, insuranceGenesis.State.Claims[0], exported.State.Claims[0])
}

func validatorInsuranceRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}
