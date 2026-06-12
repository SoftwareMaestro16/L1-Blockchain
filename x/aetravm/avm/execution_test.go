package avm

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	"github.com/stretchr/testify/require"
)

func TestAVMExecutionPhases(t *testing.T) {
	state, _ := chunk.NewBuilder().SetData([]byte("initial_state"), 104).Build()
	msgPayload, _ := chunk.NewBuilder().SetData([]byte("message"), 56).Build()

	msg := Message{
		Type:		MessageExternal,
		Sender:		"user_a",
		Target:		"contract_a",
		Payload:	msgPayload,
		GasLimit:	10000,
	}

	blockCtx := BlockContext{Height: 1, ChainID: "test-chain"}

	engine := NewEngine()

	newState, actions, receipt, err := engine.Execute(state, msg, blockCtx, 10000, 256)
	require.NoError(t, err)
	require.Equal(t, uint32(contractstypes.ExitCodeOK), receipt.ExitCode)
	require.Equal(t, uint64(2100), receipt.GasUsed)
	require.Equal(t, 5, len(receipt.PhaseGas))

	require.Equal(t, uint64(500), receipt.PhaseGas[PhaseStorage])
	require.Equal(t, uint64(100), receipt.PhaseGas[PhaseCredit])
	require.Equal(t, uint64(1000), receipt.PhaseGas[PhaseCompute])
	require.Equal(t, uint64(200), receipt.PhaseGas[PhaseAction])
	require.Equal(t, uint64(300), receipt.PhaseGas[PhaseFinalization])

	require.NotNil(t, newState)
	require.Equal(t, state.Hash(), newState.Hash())
	require.NotEmpty(t, receipt.ExecutionTraceHash)

	_, _, receipt2, err := engine.Execute(state, msg, blockCtx, 1000, 256)
	require.NoError(t, err)
	require.Equal(t, uint32(contractstypes.ExitCodeOutOfGas), receipt2.ExitCode)
	require.Equal(t, uint64(1000), receipt2.GasUsed)
	require.Equal(t, receipt2.StateRootBefore, receipt2.StateRootAfter)
	require.Empty(t, actions)

	abortPayload, _ := chunk.NewBuilder().SetData([]byte("trigger_abort"), 13).Build()
	msg.Payload = abortPayload
	_, _, receipt3, err := engine.Execute(state, msg, blockCtx, 10000, 256)
	require.NoError(t, err)
	require.Equal(t, uint32(contractstypes.ExitCodeContractAbort), receipt3.ExitCode)
	require.Equal(t, receipt3.StateRootBefore, receipt3.StateRootAfter)

	forbiddenPayload, _ := chunk.NewBuilder().SetData([]byte("use_forbidden_opcode"), 20).Build()
	msg.Payload = forbiddenPayload
	_, _, receipt4, err := engine.Execute(state, msg, blockCtx, 10000, 256)
	require.NoError(t, err)
	require.Equal(t, uint32(contractstypes.ExitCodeCodeRejected), receipt4.ExitCode)
}

func TestDeterministicExecution(t *testing.T) {
	state, _ := chunk.NewBuilder().SetData([]byte("state"), 40).Build()
	msgPayload, _ := chunk.NewBuilder().SetData([]byte("emit_actions"), 12).Build()

	msg := Message{
		Type:		MessageInternal,
		Sender:		"contract_a",
		Target:		"contract_b",
		Payload:	msgPayload,
		GasLimit:	5000,
	}
	blockCtx := BlockContext{Height: 1, ChainID: "test-chain"}

	engine := NewEngine()

	res1_state, res1_actions, res1_receipt, _ := engine.Execute(state, msg, blockCtx, 5000, 256)
	res2_state, _, res2_receipt, _ := engine.Execute(state, msg, blockCtx, 5000, 256)

	require.Equal(t, res1_state.Hash(), res2_state.Hash())
	require.Equal(t, res1_receipt.GasUsed, res2_receipt.GasUsed)
	require.Equal(t, res1_receipt.StateRootAfter, res2_receipt.StateRootAfter)
	require.Equal(t, res1_receipt.EmittedActionsHash, res2_receipt.EmittedActionsHash)
	require.Equal(t, res1_receipt.ExecutionTraceHash, res2_receipt.ExecutionTraceHash)

	require.Equal(t, 2, len(res1_actions))

	require.Equal(t, ActionInternal, res1_actions[0].Type)
	require.Equal(t, ActionExternal, res1_actions[1].Type)
}

func TestSystemBounceOnRevert(t *testing.T) {
	state, _ := chunk.NewBuilder().SetData([]byte("state"), 40).Build()
	msgPayload, _ := chunk.NewBuilder().SetData([]byte("emit_with_bounce"), 16).Build()

	msg := Message{
		Type:		MessageExternal,
		Sender:		"user_a",
		Target:		"contract_a",
		Payload:	msgPayload,
		GasLimit:	1500,
	}
	blockCtx := BlockContext{Height: 1, ChainID: "test-chain"}

	engine := NewEngine()

	_, actions, receipt, err := engine.Execute(state, msg, blockCtx, 1500, 256)
	require.NoError(t, err)
	require.Equal(t, uint32(contractstypes.ExitCodeOutOfGas), receipt.ExitCode)

	require.Equal(t, 1, len(actions))
	require.Equal(t, ActionSystem, actions[0].Type)
	require.True(t, actions[0].SystemBounce)
}
