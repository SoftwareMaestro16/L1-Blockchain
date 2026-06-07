package anft

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

func TestANFT66AndASBT67RunAsAsyncAVMContracts(t *testing.T) {
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	deployer := testAddr(10)
	collectionAddress, err := executor.DeployContract(deployer, anftCodeHash(1), []byte("collection"), nil, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	admin := testAddr(1)
	alice := testAddr(2)
	bob := testAddr(3)
	authority := testAddr(4)
	collection := testCollection(admin)
	collection.Address = collectionAddress
	state, err := NewState(collection)
	require.NoError(t, err)
	require.NoError(t, executor.RegisterHandler(collectionAddress, state.AsyncHandler()))

	mintNFT := EncodeMintNFTMessage(MintNFTMessage{Caller: admin, Owner: alice, Metadata: testMetadata("Runtime NFT")})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{anftMessage(admin, collectionAddress, OpcodeMintNFT, 1, mintNFT)}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	itemAddress, err := state.ItemAddress(0)
	require.NoError(t, err)

	transfer := EncodeTransferMessage(TransferMessage{Caller: alice, ItemAddress: itemAddress, NewOwner: bob})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{anftMessage(alice, collectionAddress, OpcodeTransfer, 2, transfer)}))
	receipts, err = executor.ProcessBlock(2)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	owner, err := state.RequestCurrentOwner(itemAddress)
	require.NoError(t, err)
	require.Equal(t, bob, owner)

	mintSBT := EncodeMintSBTMessage(MintSBTMessage{Caller: admin, Owner: alice, Authority: authority, Metadata: testMetadata("Runtime SBT")})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{anftMessage(admin, collectionAddress, OpcodeMintSBT, 3, mintSBT)}))
	receipts, err = executor.ProcessBlock(3)
	require.NoError(t, err)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	sbtAddress, err := state.ItemAddress(1)
	require.NoError(t, err)

	badTransfer := EncodeTransferMessage(TransferMessage{Caller: alice, ItemAddress: sbtAddress, NewOwner: bob})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{anftMessage(alice, collectionAddress, OpcodeTransfer, 4, badTransfer)}))
	receipts, err = executor.ProcessBlock(4)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	require.Equal(t, async.ResultExecutionFailed, receipts[0].ResultCode)
	require.Equal(t, async.BounceOpcode, receipts[1].Opcode)
	owner, err = state.RequestCurrentOwner(sbtAddress)
	require.NoError(t, err)
	require.Equal(t, alice, owner)

	revoke := EncodeRevokeSBTMessage(RevokeSBTMessage{Caller: authority, ItemAddress: sbtAddress, RevokedAt: 10, Reason: "done"})
	require.NoError(t, executor.EnqueueTxMessages([]async.MessageEnvelope{anftMessage(authority, collectionAddress, OpcodeRevokeSBT, 5, revoke)}))
	receipts, err = executor.ProcessBlock(5)
	require.NoError(t, err)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	proof, err := state.ProveSBTOwnership(sbtAddress, alice)
	require.NoError(t, err)
	require.True(t, proof.Revoked)
}

func TestANFT66AsyncContractRejectsNonNaetForwardFee(t *testing.T) {
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	collectionAddress, err := executor.DeployContract(testAddr(10), anftCodeHash(1), []byte("collection"), nil, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	msg := anftMessage(testAddr(1), collectionAddress, OpcodeMintNFT, 1, []byte("bad"))
	msg.ForwardFee = sdk.NewInt64Coin("testtoken", 1)
	require.ErrorContains(t, executor.EnqueueTxMessages([]async.MessageEnvelope{msg}), "forward fee denom")
}

func anftMessage(source, destination sdk.AccAddress, opcode uint32, queryID uint64, body []byte) async.MessageEnvelope {
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

func anftCodeHash(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, async.CodeHashLength)
}
