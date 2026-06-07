package aft

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

func TestAFT44RunsAsAsyncAVMContractAndBouncesTransfer(t *testing.T) {
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	deployer := testAddr(10)
	masterAddress, err := executor.DeployContract(deployer, aftCodeHash(1), []byte("master"), nil, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	admin := testAddr(1)
	alice := testAddr(2)
	bob := testAddr(3)
	master := testMaster(admin)
	master.Address = masterAddress
	state, err := NewState(master)
	require.NoError(t, err)
	require.NoError(t, executor.RegisterHandler(masterAddress, state.AsyncHandler()))

	mint := EncodeMintMessage(MintMessage{Caller: admin, Recipient: alice, Amount: sdkmath.NewInt(100)})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{aftMessage(admin, masterAddress, OpcodeMint, 1, mint)}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	require.Equal(t, sdkmath.NewInt(100), state.Master.TotalSupply)

	metadata := EncodeChangeMetadataMessage(admin, TokenMetadata{Name: "Runtime Token", Symbol: "RTK", Decimals: 6, ContentRef: "ipfs://runtime"})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{aftMessage(admin, masterAddress, OpcodeMetadata, 10, metadata)}))
	receipts, err = executor.ProcessBlock(10)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	require.Equal(t, "RTK", state.Master.Metadata.Symbol)

	transfer := EncodeTransferMessage(TransferMessage{Owner: alice, Recipient: bob, Amount: sdkmath.NewInt(25)})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{aftMessage(alice, masterAddress, OpcodeTransfer, 2, transfer)}))
	receipts, err = executor.ProcessBlock(2)
	require.NoError(t, err)
	require.Len(t, receipts, 3)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	require.Equal(t, async.ResultNoDestination, receipts[1].ResultCode)
	require.Equal(t, async.BounceOpcode, receipts[2].Opcode)
	require.Equal(t, uint64(1), executor.Metrics().BouncedMessages)
	require.NoError(t, state.ValidateAccounting())

	aliceWallet, ok, err := state.Wallet(alice)
	require.NoError(t, err)
	require.True(t, ok)
	_, pending := aliceWallet.PendingQueryIDs[2]
	_, processed := aliceWallet.ProcessedQueryIDs[2]
	require.False(t, pending)
	require.True(t, processed)
}

func TestAFT44AsyncContractRejectsNonNaetForwardFee(t *testing.T) {
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	masterAddress, err := executor.DeployContract(testAddr(10), aftCodeHash(1), []byte("master"), nil, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	msg := aftMessage(testAddr(1), masterAddress, OpcodeMint, 1, []byte("bad"))
	msg.ForwardFee = sdk.NewInt64Coin("testtoken", 1)
	require.ErrorContains(t, executor.EnqueueTxMessages([]async.MessageEnvelope{msg}), "forward fee denom")
}

func aftMessage(source, destination sdk.AccAddress, opcode uint32, queryID uint64, body []byte) async.MessageEnvelope {
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

func aftCodeHash(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, async.CodeHashLength)
}
