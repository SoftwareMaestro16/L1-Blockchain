package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const determinismChainID = "orbitalis-determinism-1"

func TestDeterministicVoteExtensionsAreConfigGated(t *testing.T) {
	require.False(t, deterministicVoteExtensionsEnabled(nil))
	require.False(t, deterministicVoteExtensionsEnabled(simtestutil.AppOptionsMap{}))
	require.False(t, deterministicVoteExtensionsEnabled(simtestutil.AppOptionsMap{
		DeterministicVoteExtensionsAppOption: "not-bool",
	}))
	require.True(t, deterministicVoteExtensionsEnabled(simtestutil.AppOptionsMap{
		DeterministicVoteExtensionsAppOption: true,
	}))
	require.True(t, deterministicVoteExtensionsEnabled(simtestutil.AppOptionsMap{
		DeterministicVoteExtensionsAppOption: "true",
	}))
}

func TestVoteExtensionPayloadIsDeterministicAndVerified(t *testing.T) {
	handler := NewVoteExtensionHandler()
	req := &abci.RequestExtendVote{
		Hash:   []byte("deterministic-block-hash"),
		Height: 42,
	}

	respA, err := handler.ExtendVote()(sdk.Context{}, req)
	require.NoError(t, err)
	respB, err := handler.ExtendVote()(sdk.Context{}, req)
	require.NoError(t, err)
	require.Equal(t, respA.VoteExtension, respB.VoteExtension)

	var ve VoteExtension
	require.NoError(t, json.Unmarshal(respA.VoteExtension, &ve))
	require.Equal(t, req.Hash, ve.Hash)
	require.Equal(t, req.Height, ve.Height)
	require.Equal(t, deterministicVoteExtensionData(req.Hash, req.Height), ve.Data)

	verifyResp, err := handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Hash:          req.Hash,
		Height:        req.Height,
		VoteExtension: respA.VoteExtension,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, verifyResp.Status)

	tampered := append([]byte(nil), respA.VoteExtension...)
	tampered[len(tampered)-2] ^= 0x01
	verifyResp, err = handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Hash:          req.Hash,
		Height:        req.Height,
		VoteExtension: tampered,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, verifyResp.Status)
}

func TestInitChainerRejectsMalformedGenesisWithoutPanic(t *testing.T) {
	l1, _ := setup(false, 5)

	require.NotPanics(t, func() {
		resp, err := l1.InitChainer(sdk.Context{}, &abci.RequestInitChain{AppStateBytes: []byte("{")})
		require.Nil(t, resp)
		require.Error(t, err)
	})
}

func TestSameGenesisAndTxSequenceProducesSameAppHashAndExport(t *testing.T) {
	genesisBytes, valSetHash, txBytes := deterministicGenesisAndTx(t)

	appHashA, exportedA := runDeterministicBlocks(t, genesisBytes, valSetHash, [][]byte{txBytes})
	appHashB, exportedB := runDeterministicBlocks(t, genesisBytes, valSetHash, [][]byte{txBytes})

	require.Equal(t, appHashA, appHashB)
	require.Equal(t, exportedA.Height, exportedB.Height)
	require.Equal(t, exportedA.ConsensusParams, exportedB.ConsensusParams)
	require.Equal(t, exportedA.Validators, exportedB.Validators)
	require.True(t, bytes.Equal(exportedA.AppState, exportedB.AppState))
}

func TestConsensusCriticalSourceRejectsNondeterministicPrimitives(t *testing.T) {
	for _, root := range []string{"../app", "../x"} {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			require.NoError(t, err)
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") ||
				strings.HasSuffix(path, ".pb.go") ||
				strings.HasSuffix(path, ".pb.gw.go") ||
				strings.HasSuffix(path, "test_helpers.go") {
				return nil
			}

			bz, err := os.ReadFile(path)
			require.NoError(t, err)
			src := string(bz)
			for _, forbidden := range []string{
				"crypto/rand",
				"math/rand",
				"rand.",
				"time.Now(",
				"go func",
				"select {",
			} {
				require.NotContains(t, src, forbidden, path)
			}
			return nil
		})
		require.NoError(t, err)
	}
}

func deterministicGenesisAndTx(t *testing.T) ([]byte, []byte, []byte) {
	t.Helper()

	seedApp := NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.AppOptionsMap{
		flags.FlagHome: t.TempDir(),
	})

	senderPriv := secp256k1.GenPrivKeyFromSecret([]byte("orbitalis deterministic sender"))
	recipientPriv := secp256k1.GenPrivKeyFromSecret([]byte("orbitalis deterministic recipient"))
	sender := sdk.AccAddress(senderPriv.PubKey().Address())
	recipient := sdk.AccAddress(recipientPriv.PubKey().Address())

	genAccs := []authtypes.GenesisAccount{
		authtypes.NewBaseAccount(sender, senderPriv.PubKey(), 0, 0),
		authtypes.NewBaseAccount(recipient, recipientPriv.PubKey(), 1, 0),
	}
	balance := banktypes.Balance{
		Address: sender.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(BondDenom, sdkmath.NewInt(100_000_000_000_000))),
	}

	valPriv := cmted25519.GenPrivKeyFromSecret([]byte("orbitalis deterministic validator"))
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{cmttypes.NewValidator(valPriv.PubKey(), 1)})

	genesis := seedApp.DefaultGenesis()
	genesis, err := simtestutil.GenesisStateWithValSet(seedApp.AppCodec(), genesis, valSet, genAccs, balance)
	require.NoError(t, err)

	genesisBytes, err := cmtjson.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	txBytes := deterministicMsgSendTx(t, seedApp, senderPriv, sender, recipient, 0, 0)
	return genesisBytes, valSet.Hash(), txBytes
}

func deterministicMsgSendTx(
	t *testing.T,
	l1 *L1App,
	priv *secp256k1.PrivKey,
	from sdk.AccAddress,
	to sdk.AccAddress,
	accountNumber uint64,
	sequence uint64,
) []byte {
	t.Helper()

	txConfig := l1.TxConfig()
	txBuilder := txConfig.NewTxBuilder()
	msg := banktypes.NewMsgSend(from, to, sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 12345)))
	require.NoError(t, txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(200_000)
	txBuilder.SetMemo("")

	signMode, err := authsigning.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	sig := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: signMode,
		},
		Sequence: sequence,
	}
	require.NoError(t, txBuilder.SetSignatures(sig))

	signerData := authsigning.SignerData{
		Address:       from.String(),
		ChainID:       determinismChainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
		PubKey:        priv.PubKey(),
	}
	signBytes, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		signMode,
		signerData,
		txBuilder.GetTx(),
	)
	require.NoError(t, err)

	signature, err := priv.Sign(signBytes)
	require.NoError(t, err)
	sig.Data.(*signing.SingleSignatureData).Signature = signature
	require.NoError(t, txBuilder.SetSignatures(sig))

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	return txBytes
}

func runDeterministicBlocks(t *testing.T, genesisBytes []byte, valSetHash []byte, firstBlockTxs [][]byte) ([]byte, exportedState) {
	t.Helper()

	l1 := NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.AppOptionsMap{
		flags.FlagHome: t.TempDir(),
	}, bam.SetChainID(determinismChainID))
	_, err := l1.InitChain(&abci.RequestInitChain{
		ChainId:         determinismChainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   genesisBytes,
	})
	require.NoError(t, err)

	var appHash []byte
	blockTxs := [][][]byte{firstBlockTxs, nil, nil}
	for i, txs := range blockTxs {
		height := int64(i + 1)
		resp, err := l1.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:             height,
			Time:               time.Unix(height, 0).UTC(),
			Hash:               deterministicBlockHash(height),
			NextValidatorsHash: valSetHash,
			Txs:                txs,
		})
		require.NoError(t, err)
		require.Len(t, resp.TxResults, len(txs))
		for _, txResult := range resp.TxResults {
			require.Equal(t, uint32(0), txResult.Code, txResult.Log)
		}
		appHash = append(appHash[:0], resp.AppHash...)
		_, err = l1.Commit()
		require.NoError(t, err)
	}

	exported, err := l1.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	return appHash, exportedState{
		AppState:        append([]byte(nil), exported.AppState...),
		Validators:      exported.Validators,
		Height:          exported.Height,
		ConsensusParams: exported.ConsensusParams,
	}
}

func deterministicBlockHash(height int64) []byte {
	sum := sha256.Sum256([]byte(fmt.Sprintf("orbitalis-deterministic-block-%d", height)))
	return sum[:]
}

type exportedState struct {
	AppState        []byte
	Validators      []cmttypes.GenesisValidator
	Height          int64
	ConsensusParams cmtproto.ConsensusParams
}
