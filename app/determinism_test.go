package app

import (
	"crypto/sha256"
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestDefaultGenesisJSONIsDeterministic(t *testing.T) {
	_, firstGenesis := setup(true, 5)
	firstBytes, err := json.Marshal(firstGenesis)
	require.NoError(t, err)
	firstHash := sha256.Sum256(firstBytes)

	for i := 0; i < 5; i++ {
		app, genesis := setup(true, 5)
		_ = genesis
		genesisBytes, err := json.Marshal(genesis)
		require.NoError(t, err)
		require.Equal(t, firstBytes, genesisBytes, "default genesis JSON changed on iteration %d", i)
		require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

		hash := sha256.Sum256(genesisBytes)
		require.Equal(t, firstHash, hash)
	}
}

func TestExportAppStateIsDeterministicForSameState(t *testing.T) {
	app := Setup(t, false)

	first, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	firstHash := sha256.Sum256(first.AppState)

	for i := 0; i < 3; i++ {
		exported, err := app.ExportAppStateAndValidators(false, nil, nil)
		require.NoError(t, err)
		require.Equal(t, first.AppState, exported.AppState, "export app state changed on iteration %d", i)
		require.Equal(t, firstHash, sha256.Sum256(exported.AppState))
	}
}

func TestSameGenesisAndEmptyBlockSequenceExportDeterministically(t *testing.T) {
	baseApp, genesis, valSet := deterministicGenesisWithValidator(t)
	genesisBytes, err := cmtjson.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	require.NoError(t, baseApp.BasicModuleManager.ValidateGenesis(baseApp.AppCodec(), baseApp.TxConfig(), genesis))

	first := runEmptyBlocksAndExportAppState(t, genesisBytes, valSet.Hash(), 3)
	firstHash := sha256.Sum256(first)

	for i := 0; i < 3; i++ {
		next := runEmptyBlocksAndExportAppState(t, genesisBytes, valSet.Hash(), 3)
		require.Equal(t, first, next, "same genesis and empty block sequence exported different app state on iteration %d", i)
		require.Equal(t, firstHash, sha256.Sum256(next))
	}
}

func deterministicGenesisWithValidator(t testing.TB) (*L1App, GenesisState, *cmttypes.ValidatorSet) {
	t.Helper()

	app, genesis := setup(true, 5)
	validatorPrivKey := cmted25519.GenPrivKeyFromSecret([]byte("aetra-deterministic-validator"))
	validator := cmttypes.NewValidator(validatorPrivKey.PubKey(), 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	accountPrivKey := secp256k1.GenPrivKeyFromSecret([]byte("aetra-deterministic-account"))
	account := authtypes.NewBaseAccount(accountPrivKey.PubKey().Address().Bytes(), accountPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address:	account.GetAddress().String(),
		Coins: sdk.NewCoins(
			sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000)),
		),
	}

	genesis, err := simtestutil.GenesisStateWithValSet(app.AppCodec(), genesis, valSet, []authtypes.GenesisAccount{account}, balance)
	require.NoError(t, err)
	genesis = withNativeTokenMetadata(app.AppCodec(), genesis)
	return app, genesis, valSet
}

func runEmptyBlocksAndExportAppState(t testing.TB, genesisBytes []byte, nextValidatorsHash []byte, blocks int64) []byte {
	t.Helper()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	app := NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, appOptions)

	_, err := app.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	simtestutil.DefaultConsensusParams,
		AppStateBytes:		genesisBytes,
	})
	require.NoError(t, err)

	for height := int64(1); height <= blocks; height++ {
		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:			height,
			Hash:			app.LastCommitID().Hash,
			NextValidatorsHash:	nextValidatorsHash,
		})
		require.NoError(t, err)
		_, err = app.Commit()
		require.NoError(t, err)
	}

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	return exported.AppState
}
