package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	sdkmath "cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
)

func TestConsensusParamsAreLoadedFromGenesis(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	resp, err := app.ConsensusParamsKeeper.Params(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.NotNil(t, resp.Params.Block)
	require.Greater(t, resp.Params.Block.MaxBytes, int64(0))
	require.Greater(t, resp.Params.Block.MaxGas, int64(0))
	require.NotNil(t, resp.Params.Evidence)
	require.Len(t, resp.Params.Validator.PubKeyTypes, 1)
	require.Equal(t, "ed25519", resp.Params.Validator.PubKeyTypes[0])
}

func TestConsensusParamsAreNotHardcodedConstants(t *testing.T) {
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

	app, genesisState := setup(true, 5)
	genesisState, err = simtestutil.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)
	require.NoError(t, err)
	genesisState = withNativeTokenMetadata(app.AppCodec(), genesisState)
	stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	customParams := &cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes:	999999,
			MaxGas:		12345678,
		},
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks:	100,
			MaxAgeDuration:		1000000000,
			MaxBytes:		5000,
		},
		Validator: &cmtproto.ValidatorParams{
			PubKeyTypes: []string{cmttypes.ABCIPubKeyTypeEd25519},
		},
	}

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	customParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)

	ctx := app.NewContext(false)
	resp, err := app.ConsensusParamsKeeper.Params(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, int64(999999), resp.Params.Block.MaxBytes)
	require.Equal(t, int64(12345678), resp.Params.Block.MaxGas)
	require.Equal(t, int64(100), resp.Params.Evidence.MaxAgeNumBlocks)
}
