package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SetupOptions defines arguments that are passed into `Simapp` constructor.
type SetupOptions struct {
	Logger	log.Logger
	DB	*dbm.MemDB
	AppOpts	servertypes.AppOptions
}

func setup(withGenesis bool, invCheckPeriod uint) (*L1App, GenesisState) {
	db := dbm.NewMemDB()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome

	app := NewL1App(log.NewNopLogger(), db, true, appOptions)
	if withGenesis {
		return app, app.DefaultGenesis()
	}
	return app, GenesisState{}
}

// NewSimappWithCustomOptions initializes a new L1App with custom options.
func NewSimappWithCustomOptions(t *testing.T, isCheckTx bool, options SetupOptions) *L1App {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address:	acc.GetAddress().String(),
		Coins:		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}

	app := NewL1App(options.Logger, options.DB, true, options.AppOpts)
	genesisState := app.DefaultGenesis()
	genesisState, err = simtestutil.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)
	require.NoError(t, err)
	genesisState = withNativeTokenMetadata(app.AppCodec(), genesisState)

	if !isCheckTx {

		stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		_, err = app.InitChain(&abci.RequestInitChain{
			Validators:		[]abci.ValidatorUpdate{},
			ConsensusParams:	simtestutil.DefaultConsensusParams,
			AppStateBytes:		stateBytes,
		})
		require.NoError(t, err)
	}

	return app
}

// Setup initializes a new L1App. A Nop logger is set in L1App.
func Setup(t testing.TB, isCheckTx bool) *L1App {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address:	acc.GetAddress().String(),
		Coins:		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}

	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)

	return app
}

// SetupWithGenesisValSet initializes a new L1App with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the l1app from first genesis
// account. A Nop logger is set in L1App.
func SetupWithGenesisValSet(t testing.TB, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *L1App {
	t.Helper()

	app, genesisState := setup(true, 5)
	genesisState, err := simtestutil.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
	require.NoError(t, err)
	genesisState = withNativeTokenMetadata(app.AppCodec(), genesisState)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	simtestutil.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	},
	)
	require.NoError(t, err)

	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:			app.LastBlockHeight() + 1,
		Hash:			app.LastCommitID().Hash,
		NextValidatorsHash:	valSet.Hash(),
	})
	require.NoError(t, err)

	return app
}

// GenesisStateWithSingleValidator initializes GenesisState with a single validator and genesis accounts
// that also act as delegators.
func GenesisStateWithSingleValidator(t *testing.T, app *L1App) GenesisState {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balances := []banktypes.Balance{
		{
			Address:	acc.GetAddress().String(),
			Coins:		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
		},
	}

	genesisState := app.DefaultGenesis()
	genesisState, err = simtestutil.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, []authtypes.GenesisAccount{acc}, balances...)
	require.NoError(t, err)
	genesisState = withNativeTokenMetadata(app.AppCodec(), genesisState)

	return genesisState
}

// AddTestAddrsIncremental constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrsIncremental(app *L1App, ctx sdk.Context, accNum int, accAmt sdkmath.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, accAmt, simtestutil.CreateIncrementalAccounts)
}

func GetBondedTestValidator(t *testing.T, app *L1App, ctx sdk.Context) stakingtypes.Validator {
	t.Helper()

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	for _, validator := range validators {
		if validator.Status == stakingtypes.Bonded {
			return validator
		}
	}
	require.FailNow(t, "expected at least one bonded validator")
	return stakingtypes.Validator{}
}

func AddTestAddrsWithCoins(t *testing.T, app *L1App, ctx sdk.Context, accNum int, coins sdk.Coins) []sdk.AccAddress {
	t.Helper()

	testAddrs := simtestutil.CreateIncrementalAccounts(accNum)
	for _, addr := range testAddrs {
		FundTestAddr(t, app, ctx, addr, coins)
	}

	return testAddrs
}

func FundTestAddr(t *testing.T, app *L1App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()

	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
}

func addTestAddrs(app *L1App, ctx sdk.Context, accNum int, accAmt sdkmath.Int, strategy simtestutil.GenerateAccountStrategy) []sdk.AccAddress {
	testAddrs := strategy(accNum)
	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	if err != nil {
		panic(err)
	}

	initCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, accAmt))

	for _, addr := range testAddrs {
		initAccountWithCoins(app, ctx, addr, initCoins)
	}

	return testAddrs
}

func initAccountWithCoins(app *L1App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	if err != nil {
		panic(err)
	}

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins)
	if err != nil {
		panic(err)
	}
}

// NewTestNetworkFixture returns a new l1app AppConstructor for network simulation tests
func NewTestNetworkFixture() network.TestFixture {
	dir, err := os.MkdirTemp("", "l1app")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	defer os.RemoveAll(dir)

	app := NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(dir))

	appCtr := func(val network.ValidatorI) servertypes.Application {
		return NewL1App(
			val.GetCtx().Logger, dbm.NewMemDB(), true,
			simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir),
			bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			bam.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			bam.SetChainID(val.GetCtx().Viper.GetString(flags.FlagChainID)),
		)
	}

	return network.TestFixture{
		AppConstructor:	appCtr,
		GenesisState:	app.DefaultGenesis(),
		EncodingConfig: testutil.TestEncodingConfig{
			InterfaceRegistry:	app.InterfaceRegistry(),
			Codec:			app.AppCodec(),
			TxConfig:		app.TxConfig(),
			Amino:			app.LegacyAmino(),
		},
	}
}
