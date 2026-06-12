package app

import (
	"crypto/sha256"
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

	reporterkeeper "github.com/sovereign-l1/l1/x/reporter/keeper"
	reportertypes "github.com/sovereign-l1/l1/x/reporter/types"
)

func TestReporterSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, reportertypes.ModuleName)
	require.Contains(t, app.keys, reportertypes.StoreKey)
	require.Contains(t, genesis, reportertypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), reportertypes.ModuleName)

	var reporterGenesis reporterkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[reportertypes.ModuleName], &reporterGenesis))
	require.NoError(t, reporterGenesis.Validate())
}

func TestReporterStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	reporterGenesis := reporterkeeper.DefaultGenesis()
	reporterAddress := reporterRawAddress("11")
	reporterGenesis.State.Reporters = []reportertypes.ReporterRecord{{
		ReporterAddress:	reporterAddress,
		BondedAmount:		reportertypes.DefaultMinBondAmount,
		ReporterScore:		reporterGenesis.Params.InitialScore,
		AcceptedReports:	1,
		Status:			reportertypes.StatusActive,
		RewardHistory: []reportertypes.ReporterReward{{
			ReportID:	"app-report-1",
			Amount:		77,
			CreatedAt:	1,
		}},
	}}
	reporterGenesis.State.Reports = []reportertypes.ReportRecord{{
		ReportID:		"app-report-1",
		ReporterAddress:	reporterAddress,
		ReportType:		reportertypes.ReportTypeAvailability,
		Subject:		"availability-window",
		PayloadHash:		reporterPayloadHash("app-report-1"),
		PayloadSizeBytes:	128,
		Status:			reportertypes.ReportStatusAccepted,
		SubmittedHeight:	1,
		FinalizedHeight:	1,
		RewardAmount:		77,
	}}
	reporterGenesis.State = reporterGenesis.State.Normalize(reporterGenesis.Params)
	require.NoError(t, reporterGenesis.Validate())
	reporterGenesisBytes, err := json.Marshal(reporterGenesis)
	require.NoError(t, err)
	genesis[reportertypes.ModuleName] = reporterGenesisBytes
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
	exported, err := restarted.ReporterKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Reporters, 1)
	require.Len(t, exported.State.Reporters[0].RewardHistory, 1)
	require.False(t, exported.State.Reporters[0].RewardHistory[0].Claimed)
}

func reporterPayloadHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func reporterRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}
