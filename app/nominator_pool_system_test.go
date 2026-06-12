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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	nominatorpoolkeeper "github.com/sovereign-l1/l1/x/nominator-pool/keeper"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestNominatorPoolSystemModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

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
		PoolID:			"app-pool-1",
		PoolOperator:		nominatorPoolRawAddress("11"),
		ValidatorTarget:	nominatorPoolRawAddress("12"),
		TotalShares:		1_000,
		TotalBondedStake:	1_100,
		RewardIndex:		100 * nominatorpooltypes.IndexScale / 1_000,
		PoolCommissionBps:	100,
		Status:			nominatorpooltypes.PoolStatusActive,
		DelegatorShares: []nominatorpooltypes.DelegatorShare{{
			Delegator:		nominatorPoolRawAddress("22"),
			Shares:			1_000,
			RewardIndexCheckpoint:	0,
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
	exported, err := restarted.NominatorPoolKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Pools, 1)
	require.Equal(t, poolGenesis.State.Pools[0].RewardIndex, exported.State.Pools[0].RewardIndex)
}

func TestNominatorPoolRuntimeMutationPersistsToKVStore(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = source.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)

	sourceCtx := source.NewNextBlockContext(cmtproto.Header{Height: 1})
	initial, err := source.NominatorPoolKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	contractUser, contractRaw := nominatorPoolAddressPair(t, "77")
	userAddress, _ := nominatorPoolAddressPair(t, "44")

	poolID := "runtime-kv-official-pool"
	nominatorPoolMsg(t, source, sourceCtx, &nominatorpooltypes.MsgCreateOfficialLiquidStakingPool{
		Authority:		initial.Params.Authority,
		PoolID:			poolID,
		ContractAddressUser:	contractUser,
		ContractAddressRaw:	contractRaw,
		PoolOperator:		nominatorPoolRawAddress("11"),
		PoolCommissionBps:	100,
		Height:			2,
	})
	nominatorPoolMsg(t, source, sourceCtx, &nominatorpooltypes.MsgDepositToStakingPool{
		PoolID:		poolID,
		WalletAddress:	userAddress,
		Amount:		nominatorpooltypes.DefaultMinPoolDeposit,
		Height:		3,
	})

	source.SimWriteState()
	_, err = source.Commit()
	require.NoError(t, err)

	restarted := NewL1App(log.NewNopLogger(), db, true, appOptions)
	restartedCtx := restarted.NewUncachedContext(false, cmtproto.Header{Height: restarted.LastBlockHeight()})
	exported, err := restarted.NominatorPoolKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.Pools, 1)
	require.Equal(t, poolID, exported.State.Pools[0].PoolID)
	require.Len(t, exported.State.PoolShares, 1)
	require.Equal(t, userAddress, exported.State.PoolShares[0].Owner)
	require.Equal(t, nominatorpooltypes.DefaultMinPoolDeposit, exported.State.PoolShares[0].Shares)
}

func TestFinalAppWiringOfficialStakingPoolFlowExportImportRestart(t *testing.T) {
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, appOptions)
	sourceCtx := source.NewNextBlockContext(cmtproto.Header{Height: 1})

	initial, err := source.NominatorPoolKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	contractUser, contractRaw := nominatorPoolAddressPair(t, "66")
	userAddress, userRaw := nominatorPoolAddressPair(t, "33")

	poolID := "final-app-official-pool"
	nominatorPoolMsg(t, source, sourceCtx, &nominatorpooltypes.MsgCreateOfficialLiquidStakingPool{
		Authority:		initial.Params.Authority,
		PoolID:			poolID,
		ContractAddressUser:	contractUser,
		ContractAddressRaw:	contractRaw,
		PoolOperator:		nominatorPoolRawAddress("11"),
		PoolCommissionBps:	100,
		Height:			2,
	})

	nominatorPoolMsg(t, source, sourceCtx, &nominatorpooltypes.MsgDepositToStakingPool{
		PoolID:		poolID,
		WalletAddress:	userAddress,
		Amount:		nominatorpooltypes.DefaultMinPoolDeposit,
		Height:		3,
	})
	pool, found := source.NominatorPoolKeeper.NominatorPool(poolID)
	require.True(t, found)
	require.True(t, pool.OfficialLiquidStaking)
	require.Equal(t, contractUser, pool.ContractAddressUser)
	require.Equal(t, contractRaw, pool.ContractAddressRaw)
	share, found := source.NominatorPoolKeeper.PoolShare(nominatorpooltypes.QueryPoolShareRequest{PoolID: poolID, Delegator: userRaw})
	require.True(t, found)
	require.Equal(t, nominatorpooltypes.DefaultMinPoolDeposit, share.Share.Shares)

	exported, err := source.NominatorPoolKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	require.Len(t, exported.State.LiquidStakingPools, 1)
	require.Len(t, exported.State.PoolShares, 1)
	require.Equal(t, userAddress, exported.State.PoolShares[0].Owner)
	require.Equal(t, nominatorpooltypes.DefaultMinPoolDeposit, exported.State.PoolShares[0].Shares)

	restarted := NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, appOptions)
	restartedCtx := restarted.NewUncachedContext(false, cmtproto.Header{Height: 4})
	require.NoError(t, restarted.NominatorPoolKeeper.InitGenesisState(restartedCtx, exported))
	reexported, err := restarted.NominatorPoolKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	require.Equal(t, exported, reexported)
}

func nominatorPoolRawAddress(hexByte string) string {
	return "4:000000000000000000000000" + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte + hexByte
}

func nominatorPoolAddressPair(t *testing.T, hexByte string) (string, string) {
	t.Helper()
	raw := nominatorPoolRawAddress(hexByte)
	bz, err := addressing.Parse(raw)
	require.NoError(t, err)
	user := addressing.FormatAccAddress(sdk.AccAddress(bz))
	return user, raw
}

func nominatorPoolMsg(t *testing.T, app *L1App, ctx sdk.Context, msg sdk.Msg) interface{} {
	t.Helper()
	handler := app.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	res, err := handler(ctx, msg)
	require.NoError(t, err)
	return res
}
