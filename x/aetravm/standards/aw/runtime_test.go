package aw

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

func TestAW5RunsAsAsyncAVMContractAndBouncesFailedSend(t *testing.T) {
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	deployer := testAddr(10)
	walletAddress, err := executor.DeployContract(deployer, awCodeHash(1), []byte("wallet"), nil, sdkmath.NewInt(10_000))
	require.NoError(t, err)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	state, err := NewState(WalletState{
		Address:          walletAddress,
		SignatureAllowed: true,
		Seqno:            DefaultWalletSeqno,
		WalletID:         55,
		PublicKey:        publicKey,
		Owner:            testAddr(2),
		Extensions:       make(map[string]ExtensionState),
	})
	require.NoError(t, err)
	require.NoError(t, executor.RegisterHandler(walletAddress, state.AsyncHandler(testNow)))

	missingDestination, err := async.DeriveContractAddress(deployer, awCodeHash(2), []byte("missing"))
	require.NoError(t, err)
	cmd := signCommand(t, privateKey, ExternalCommand{
		WalletAddress: walletAddress,
		WalletID:      state.Wallet.WalletID,
		Seqno:         state.Wallet.Seqno,
		ValidUntil:    testNow + 60,
		Kind:          CommandSend,
		Messages: []OutboundMessage{{
			To:      missingDestination,
			Amount:  sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1)),
			Payload: []byte("call"),
		}},
	})
	body, err := EncodeExternalCommand(cmd)
	require.NoError(t, err)
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{awMessage(testAddr(9), walletAddress, OpcodeSignedExternal, 1, body)}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 3)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	require.Equal(t, async.ResultNoDestination, receipts[1].ResultCode)
	require.Equal(t, async.BounceOpcode, receipts[2].Opcode)
	require.Equal(t, uint64(1), state.Wallet.Seqno)
	require.Equal(t, uint64(1), executor.Metrics().BouncedMessages)

	replay := awMessage(testAddr(9), walletAddress, OpcodeSignedExternal, 2, body)
	replay.Bounce = false
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{replay}))
	receipts, err = executor.ProcessBlock(2)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, async.ResultExecutionFailed, receipts[0].ResultCode)
	require.Contains(t, receipts[0].Error, "seqno")
	require.Equal(t, uint64(1), state.Wallet.Seqno)
}

func TestAW5AsyncContractRejectsNonNaetForwardFeeBeforeMutation(t *testing.T) {
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	walletAddress, err := executor.DeployContract(testAddr(10), awCodeHash(1), []byte("wallet"), nil, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	msg := awMessage(testAddr(9), walletAddress, OpcodeSignedExternal, 1, []byte("bad"))
	msg.ForwardFee = sdk.NewInt64Coin("testtoken", 1)
	require.ErrorContains(t, executor.EnqueueTxMessages([]async.MessageEnvelope{msg}), "forward fee denom")
}

func awMessage(source, destination sdk.AccAddress, opcode uint32, queryID uint64, body []byte) async.MessageEnvelope {
	return async.MessageEnvelope{
		Source:             source,
		Destination:        destination,
		Value:              sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
		Opcode:             opcode,
		QueryID:            queryID,
		Body:               body,
		Bounce:             true,
		CreatedLogicalTime: queryID,
		GasLimit:           100_000,
		ForwardFee:         sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
	}
}

func awCodeHash(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, async.CodeHashLength)
}
