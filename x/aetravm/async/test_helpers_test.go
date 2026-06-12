package async

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func newTestExecutor(t *testing.T) *Executor {
	t.Helper()
	executor, err := NewExecutor(DefaultParams())
	require.NoError(t, err)
	return executor
}

func deployTestContract(t *testing.T, executor *Executor, deployer sdk.AccAddress, salt []byte) sdk.AccAddress {
	t.Helper()
	address, err := executor.DeployContract(deployer, testCodeHash(salt[0]), salt, append([]byte("init:"), salt...), sdkmath.NewInt(10_000))
	require.NoError(t, err)
	return address
}

func testMessage(source, dest sdk.AccAddress, queryID uint64) MessageEnvelope {
	return MessageEnvelope{
		Source:			source,
		Destination:		dest,
		Value:			naetCoin(1),
		Opcode:			1,
		QueryID:		queryID,
		Body:			[]byte("body"),
		Bounce:			true,
		CreatedLogicalTime:	queryID,
		GasLimit:		100_000,
		ForwardFee:		forwardFee(),
	}
}

func forwardFee() sdk.Coin {
	return sdk.NewCoin(appparams.BaseDenom, DefaultParams().ForwardingFee)
}

func naetCoin(amount int64) sdk.Coin {
	return sdk.NewCoin(appparams.BaseDenom, sdkmath.NewInt(amount))
}

func testAddr(fill byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{fill}, 20))
}

func testCodeHash(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, CodeHashLength)
}

func cloneContracts(contracts []ContractAccount) []ContractAccount {
	out := make([]ContractAccount, len(contracts))
	copy(out, contracts)
	return out
}

func cloneQueuedMessagesForTest(messages []QueuedMessage) []QueuedMessage {
	out := make([]QueuedMessage, len(messages))
	copy(out, messages)
	return out
}
