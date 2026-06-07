package types

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aetravm/async"
)

func TestMessageEnqueueDeliveryAndDeterministicOrder(t *testing.T) {
	executor := newExecutor(t)
	dest := deployContract(t, executor, addr(1), []byte("dest"))
	require.NoError(t, executor.RegisterHandler(dest, func(contract async.ContractAccount, msg async.MessageEnvelope) async.ExecutionResult {
		return async.ExecutionResult{NewState: append(contract.State, msg.Body...), ResultCode: async.ResultOK}
	}))
	first := newMsg(t, addr(9), dest, 1, []byte("a"), true, 0)
	second := newMsg(t, addr(9), dest, 2, []byte("b"), true, 0)

	require.NoError(t, EnqueueMessage(executor, async.DefaultParams(), first))
	require.NoError(t, EnqueueMessage(executor, async.DefaultParams(), second))
	queue := executor.Queue()
	require.Len(t, queue, 2)
	require.Equal(t, uint64(0), queue[0].Sequence)
	require.Equal(t, uint64(1), queue[1].Sequence)

	receipts, err := DeliverMessages(executor, 1)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	require.Equal(t, async.ResultOK, receipts[0].ResultCode)
	require.Equal(t, async.ResultOK, receipts[1].ResultCode)
}

func TestBounceOnFailureExpiredMessageAndRefundNoDoubleSpend(t *testing.T) {
	t.Run("bounce missing destination", func(t *testing.T) {
		executor := newExecutor(t)
		source := deployContract(t, executor, addr(1), []byte("source"))
		missing := addr(8)
		require.NoError(t, executor.RegisterHandler(source, func(contract async.ContractAccount, msg async.MessageEnvelope) async.ExecutionResult {
			if msg.Bounced {
				return async.ExecutionResult{NewState: []byte("bounced"), ResultCode: async.ResultOK}
			}
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultOK}
		}))
		msg := newMsg(t, source, missing, 1, []byte("body"), true, 0)
		require.NoError(t, EnqueueMessage(executor, async.DefaultParams(), msg))
		receipts, err := DeliverMessages(executor, 1)
		require.NoError(t, err)
		require.Len(t, receipts, 2)
		require.Equal(t, async.ResultNoDestination, receipts[0].ResultCode)
		require.Equal(t, async.BounceOpcode, receipts[1].Opcode)
	})

	t.Run("expired", func(t *testing.T) {
		executor := newExecutor(t)
		dest := deployContract(t, executor, addr(1), []byte("dest"))
		msg := newMsg(t, addr(9), dest, 1, []byte("body"), false, 1)
		require.NoError(t, EnqueueMessage(executor, async.DefaultParams(), msg))
		receipts, err := DeliverMessages(executor, 2)
		require.NoError(t, err)
		require.Len(t, receipts, 2)
		require.Equal(t, async.ResultExpired, receipts[0].ResultCode)
		require.Equal(t, async.RefundOpcode, receipts[1].Opcode)
	})

	t.Run("refund no double spend", func(t *testing.T) {
		executor := newExecutor(t)
		msg := newMsg(t, addr(7), addr(8), 1, []byte("body"), false, 0)
		msg.Opcode = async.RefundOpcode
		require.NoError(t, EnqueueMessage(executor, async.DefaultParams(), msg))
		receipts, err := DeliverMessages(executor, 1)
		require.NoError(t, err)
		require.Len(t, receipts, 1)
		require.Equal(t, async.ResultNoDestination, receipts[0].ResultCode)
		require.Empty(t, executor.Queue())
		require.Zero(t, executor.Metrics().RefundMessages)
	})
}

func TestMessageValidationAndID(t *testing.T) {
	msg := newMsg(t, addr(1), addr(2), 1, []byte("body"), true, 0)
	require.Len(t, msg.ID, MessageIDBytes)
	require.Equal(t, msg.ID, MessageID(msg))

	msg.ID = []byte{1}
	require.ErrorContains(t, msg.Validate(async.DefaultParams()), "message id")
	msg = newMsg(t, addr(1), addr(2), 1, []byte("body"), true, 0)
	msg.GasLimit = 0
	require.ErrorContains(t, msg.Validate(async.DefaultParams()), "gas_limit")
}

func newExecutor(t *testing.T) *async.Executor {
	t.Helper()
	executor, err := async.NewExecutor(async.DefaultParams())
	require.NoError(t, err)
	return executor
}

func deployContract(t *testing.T, executor *async.Executor, deployer sdk.AccAddress, salt []byte) sdk.AccAddress {
	t.Helper()
	address, err := executor.DeployContract(deployer, bytes.Repeat([]byte{salt[0]}, async.CodeHashLength), salt, nil, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	return address
}

func newMsg(t *testing.T, source, dest sdk.AccAddress, queryID uint64, body []byte, bounce bool, deadline uint64) Message {
	t.Helper()
	msg, err := NewMessage(source, dest, sdkmath.NewInt(1), 1, queryID, body, bounce, deadline, 100_000, queryID)
	require.NoError(t, err)
	return msg
}

func addr(seed byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20))
}
