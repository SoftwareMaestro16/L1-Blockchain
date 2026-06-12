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

	nativeevidencekeeper "github.com/sovereign-l1/l1/x/evidence/keeper"
	nativeevidencetypes "github.com/sovereign-l1/l1/x/evidence/types"
)

func TestNativeEvidenceSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, nativeevidencetypes.ModuleName)
	require.Contains(t, app.keys, nativeevidencetypes.StoreKey)
	require.Contains(t, genesis, nativeevidencetypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), nativeevidencetypes.ModuleName)

	var evidenceGenesis nativeevidencekeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[nativeevidencetypes.ModuleName], &evidenceGenesis))
	require.NoError(t, evidenceGenesis.Validate())
}

func TestNativeEvidenceStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	evidenceGenesis := nativeevidencekeeper.DefaultGenesis()
	evidenceGenesis.State.Evidence = []nativeevidencetypes.EvidenceRecord{{
		EvidenceID:		"app-evidence-1",
		Status:			nativeevidencetypes.StatusPending,
		EvidenceType:		nativeevidencetypes.EvidenceTypeFraud,
		AccusedValidator:	nativeEvidenceRawAddress("11"),
		Reporter:		nativeEvidenceRawAddress("22"),
		ProofPayloadHash:	nativeEvidenceProofHash("app-evidence-1"),
		PayloadSizeBytes:	128,
		SlashDecision: nativeevidencetypes.SlashDecision{
			FractionBps:	evidenceGenesis.Params.CriticalFaultSlashFractionBps,
			Tombstone:	true,
		},
		RewardDecision: nativeevidencetypes.RewardDecision{
			Reporter:	nativeEvidenceRawAddress("22"),
			AmountNaet:	100,
		},
		SubmittedHeight:	1,
		UpdatedHeight:		1,
		ExpirationHeight:	100,
		RequiresReview:		true,
	}}
	evidenceGenesis.State = evidenceGenesis.State.Normalize(evidenceGenesis.Params)
	require.NoError(t, evidenceGenesis.Validate())
	evidenceGenesisBytes, err := json.Marshal(evidenceGenesis)
	require.NoError(t, err)
	genesis[nativeevidencetypes.ModuleName] = evidenceGenesisBytes
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
	exported, err := restarted.NativeEvidenceKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Evidence, 1)
	require.Equal(t, "app-evidence-1", exported.State.Evidence[0].EvidenceID)
}

func nativeEvidenceProofHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func nativeEvidenceRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}
