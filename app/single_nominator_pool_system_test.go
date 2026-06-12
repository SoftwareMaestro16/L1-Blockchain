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

	singlenominatorpoolkeeper "github.com/sovereign-l1/l1/x/single-nominator-pool/keeper"
	singlenominatorpooltypes "github.com/sovereign-l1/l1/x/single-nominator-pool/types"
)

func TestSingleNominatorPoolSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, singlenominatorpooltypes.ModuleName)
	require.Contains(t, app.keys, singlenominatorpooltypes.StoreKey)
	require.Contains(t, genesis, singlenominatorpooltypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), singlenominatorpooltypes.ModuleName)

	var poolGenesis singlenominatorpoolkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[singlenominatorpooltypes.ModuleName], &poolGenesis))
	require.NoError(t, poolGenesis.Validate())
}

func TestSingleNominatorPoolPendingWithdrawalSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	poolGenesis := singlenominatorpoolkeeper.DefaultGenesis()
	poolGenesis.State.Pools = []singlenominatorpooltypes.SingleNominatorPool{{
		PoolAddress:	singlenominatorPoolRawAddress("10"),
		Owner:		singlenominatorPoolRawAddress("11"),
		Validator:	singlenominatorPoolRawAddress("12"),
		BondedStake:	700,
		PendingWithdrawal: singlenominatorpooltypes.PendingWithdrawal{
			Amount:		300,
			RequestHeight:	1,
			CompleteHeight:	1 + poolGenesis.Params.UnbondingBlocks,
			Status:		singlenominatorpooltypes.WithdrawalStatusPending,
		},
		RewardBalance:	55,
		Status:		singlenominatorpooltypes.StatusActive,
	}}
	poolGenesis.State = poolGenesis.State.Normalize(poolGenesis.Params)
	require.NoError(t, poolGenesis.Validate())
	poolGenesisBytes, err := json.Marshal(poolGenesis)
	require.NoError(t, err)
	genesis[singlenominatorpooltypes.ModuleName] = poolGenesisBytes
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
	exported, err := restarted.SingleNominatorPoolKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Pools, 1)
	require.Equal(t, poolGenesis.State.Pools[0].PendingWithdrawal, exported.State.Pools[0].PendingWithdrawal)
	require.Equal(t, uint64(55), exported.State.Pools[0].RewardBalance)
}

func singlenominatorPoolRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}
