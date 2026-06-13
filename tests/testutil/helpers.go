package testutil

import (
	"encoding/json"
	"math/rand"
	"testing"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
)

func NewInitializedApp(t *testing.T, chainID string) *l1app.L1App {
	t.Helper()
	app := l1app.NewL1App(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		sims.AppOptionsMap{flags.FlagHome: t.TempDir()},
		baseapp.SetChainID(chainID),
	)
	genesis := l1app.GenesisStateWithSingleValidator(t, app)
	_ = genesis
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	_, err = app.InitChain(&abci.RequestInitChain{
		ChainId:		chainID,
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)
	return app
}

func NewContext(app *l1app.L1App, height int64) sdk.Context {
	if app.LastBlockHeight() > 0 {
		return app.NewNextBlockContext(cmtproto.Header{ChainID: app.ChainID(), Height: height}).
			WithBlockHeight(height).
			WithChainID(app.ChainID())
	}
	return app.NewContext(false).WithBlockHeight(height).WithChainID(app.ChainID())
}

func AddFundedSigner(t *testing.T, app *l1app.L1App, ctx sdk.Context, amount sdkmath.Int) (cryptotypes.PrivKey, sdk.AccAddress) {
	t.Helper()
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)
	FundAccount(t, app, ctx, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amount)))
	return priv, addr
}

func FundAccount(t *testing.T, app *l1app.L1App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
}

func EncodeSignedTx(
	t *testing.T,
	app *l1app.L1App,
	ctx sdk.Context,
	priv cryptotypes.PrivKey,
	msgs []sdk.Msg,
	fee sdk.Coins,
	gas uint64,
) []byte {
	t.Helper()
	return EncodeSignedTxWithChainID(t, app, ctx, priv, msgs, fee, gas, app.ChainID())
}

func EncodeSignedTxWithChainID(
	t *testing.T,
	app *l1app.L1App,
	ctx sdk.Context,
	priv cryptotypes.PrivKey,
	msgs []sdk.Msg,
	fee sdk.Coins,
	gas uint64,
	chainID string,
) []byte {
	t.Helper()
	addr := sdk.AccAddress(priv.PubKey().Address())
	acc := app.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	tx, err := sims.GenSignedMockTx(
		rand.New(rand.NewSource(1)), // #nosec G404 -- deterministic PRNG for reproducible test data
		app.TxConfig(),
		msgs,
		fee,
		gas,
		chainID,
		[]uint64{acc.GetAccountNumber()},
		[]uint64{acc.GetSequence()},
		priv,
	)
	require.NoError(t, err)
	txBytes, err := app.TxConfig().TxEncoder()(tx)
	require.NoError(t, err)
	return txBytes
}

func FinalizeBlock(t *testing.T, app *l1app.L1App, height int64, txs ...[]byte) *abci.ResponseFinalizeBlock {
	t.Helper()
	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	height,
		Hash:	app.LastCommitID().Hash,
		Txs:	txs,
	})
	require.NoError(t, err)
	return res
}

func Commit(t *testing.T, app *l1app.L1App) {
	t.Helper()
	_, err := app.Commit()
	require.NoError(t, err)
}

func AccountNumberAndSequence(t *testing.T, app *l1app.L1App, ctx sdk.Context, addr sdk.AccAddress) (uint64, uint64) {
	t.Helper()
	acc := app.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	return acc.GetAccountNumber(), acc.GetSequence()
}

func NewBaseAccount(addr sdk.AccAddress, pubKey cryptotypes.PubKey) types.GenesisAccount {
	return types.NewBaseAccount(addr, pubKey, 0, 0)
}

func BankBalance(addr sdk.AccAddress, coins sdk.Coins) banktypes.Balance {
	return banktypes.Balance{Address: addr.String(), Coins: coins}
}
