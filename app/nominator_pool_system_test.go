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

	nominatorpoolkeeper "github.com/sovereign-l1/l1/x/nominator-pool/keeper"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestNominatorPoolSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, nominatorpooltypes.ModuleName)
	require.Contains(t, app.keys, nominatorpooltypes.StoreKey)
	require.Contains(t, genesis, nominatorpooltypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), nominatorpooltypes.ModuleName)

	var poolGenesis nominatorpoolkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[nominatorpooltypes.ModuleName], &poolGenesis))
	require.NoError(t, poolGenesis.Validate())
}

func TestNominatorPoolStateSurvivesFinalizeBlockRestart(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	poolGenesis := nominatorpoolkeeper.DefaultGenesis()
	poolGenesis.State.Pools = []nominatorpooltypes.NominatorPool{{
		PoolID:            "app-pool-1",
		PoolOperator:      nominatorPoolRawAddress("11"),
		ValidatorTarget:   nominatorPoolRawAddress("12"),
		TotalShares:       1_000,
		TotalBondedStake:  1_100,
		RewardIndex:       100 * nominatorpooltypes.IndexScale / 1_000,
		PoolCommissionBps: 100,
		Status:            nominatorpooltypes.PoolStatusActive,
		DelegatorShares: []nominatorpooltypes.DelegatorShare{{
			Delegator:             nominatorPoolRawAddress("22"),
			Shares:                1_000,
			RewardIndexCheckpoint: 0,
		}},
	}}
	poolGenesis.State = poolGenesis.State.Normalize(poolGenesis.Params)
	require.NoError(t, poolGenesis.Validate())
	poolGenesisBytes, err := json.Marshal(poolGenesis)
	require.NoError(t, err)
	genesis[nominatorpooltypes.ModuleName] = poolGenesisBytes
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
	exported, err := restarted.NominatorPoolKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Pools, 1)
	require.Equal(t, poolGenesis.State.Pools[0].RewardIndex, exported.State.Pools[0].RewardIndex)
}

func nominatorPoolRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}
