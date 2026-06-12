package async

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecutionReceiptGeneratedForSuccess(t *testing.T) {
	params := DefaultParams()
	params.MaxMessagesPerBlock = 1
	executor, err := NewExecutor(params)
	require.NoError(t, err)
	contract := deployTestContract(t, executor, testAddr(1), []byte("receipt-success"))
	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{
			NewState:	[]byte("committed"),
			Outgoing: []MessageEnvelope{
				testMessage(contract.Address, contract.Address, 99),
			},
			ResultCode:	ResultOK,
		}
	}))

	msg := testMessage(testAddr(9), contract, 1)
	msg.Bounce = false
	msg.Value = naetCoin(0)
	require.NoError(t, executor.EnqueueMessage(msg))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)

	receipt := receipts[0]
	require.NoError(t, ValidateExecutionReceipt(receipt))
	require.NotEmpty(t, receipt.ReceiptID)
	require.NotEmpty(t, receipt.TxHash)
	require.Len(t, receipt.MessageID, MessageIDLength)
	require.Equal(t, ExecutionKindInternal, receipt.ExecutionKind)
	require.True(t, receipt.ContractAddress.Equals(contract))
	require.True(t, receipt.Caller.Equals(testAddr(9)))
	require.Equal(t, msg.GasLimit, receipt.GasLimit)
	require.Equal(t, ResultOK, receipt.ExitCode)
	require.Equal(t, "ok", receipt.ExitReason)
	require.True(t, receipt.StateCommitted)
	require.NotEqual(t, receipt.StateRootBefore, receipt.StateRootAfter)
	require.Len(t, receipt.EmittedMessageIDs, 1)
	require.Equal(t, EventInternalExecuted, receipt.Events[0].Type)
	require.Equal(t, EventMessageQueued, receipt.Events[1].Type)
}

func TestExecutionReceiptGeneratedForFailedExecutionAndRollback(t *testing.T) {
	executor := newTestExecutor(t)
	contract := deployTestContract(t, executor, testAddr(1), []byte("receipt-fail"))
	before, found := executor.Contract(contract)
	require.True(t, found)
	beforeRoot := ContractStateRoot(before)
	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{
			NewState:	[]byte("must-roll-back"),
			ResultCode:	ResultExecutionFailed,
			Error:		"handler failed",
		}
	}))

	msg := testMessage(testAddr(9), contract, 1)
	msg.Bounce = false
	msg.Value = naetCoin(0)
	require.NoError(t, executor.EnqueueMessage(msg))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)

	receipt := receipts[0]
	require.NoError(t, ValidateExecutionReceipt(receipt))
	require.Equal(t, ResultExecutionFailed, receipt.ExitCode)
	require.Equal(t, "handler failed", receipt.ExitReason)
	require.Equal(t, FailedPhaseExecution, receipt.FailedPhase)
	require.False(t, receipt.StateCommitted)
	require.Equal(t, beforeRoot, receipt.StateRootBefore)
	require.Equal(t, beforeRoot, receipt.StateRootAfter)
	after, found := executor.Contract(contract)
	require.True(t, found)
	require.Equal(t, before.State, after.State)
	require.Equal(t, beforeRoot, ContractStateRoot(after))
}

func TestExecutionReceiptEventsHaveStableAttributes(t *testing.T) {
	executor := newTestExecutor(t)
	contract := deployTestContract(t, executor, testAddr(1), []byte("receipt-events"))
	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{NewState: []byte("ok"), ResultCode: ResultOK}
	}))
	require.NoError(t, executor.EnqueueMessage(testMessage(testAddr(9), contract, 7)))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	event := receipts[0].Events[0]
	require.Equal(t, EventInternalExecuted, event.Type)
	require.Equal(t, []AVMEventAttribute{
		EventAttr("message_id", event.Attributes[0].Value),
		EventAttr("contract", queueAddressKey(contract)),
		EventAttr("caller", queueAddressKey(testAddr(9))),
		EventAttr("destination", queueAddressKey(contract)),
		EventAttr("opcode", "1"),
		EventAttr("query_id", "7"),
		EventAttr("exit_code", "0"),
		EventAttr("gas_used", "10000"),
		EventAttr("height", "1"),
		EventAttr("state_committed", "true"),
	}, event.Attributes)

	for _, eventType := range []string{
		EventCodeStored,
		EventContractDeployed,
		EventExternalExecuted,
		EventInternalExecuted,
		EventMessageQueued,
		EventMessageBounced,
		EventContractFrozen,
		EventContractUnfrozen,
		EventRentPaid,
	} {
		require.NoError(t, NewAVMEvent(eventType, EventAttr("stable", "true")).Validate())
	}
}

func TestReceiptsExportImportStableAndQueryPaginated(t *testing.T) {
	executor := newTestExecutor(t)
	contract := deployTestContract(t, executor, testAddr(1), []byte("receipt-query"))
	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{NewState: append([]byte("q:"), byte(msg.QueryID)), ResultCode: ResultOK}
	}))
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{
		testMessage(testAddr(9), contract, 1),
		testMessage(testAddr(9), contract, 2),
	}))
	_, err := executor.ProcessBlock(1)
	require.NoError(t, err)

	page, err := executor.QueryReceipts(contract, 1)
	require.NoError(t, err)
	require.Len(t, page, 1)
	require.Equal(t, uint64(1), page[0].QueryID)
	require.ErrorContains(t, func() error {
		_, err := executor.QueryReceipts(contract, 0)
		return err
	}(), "limit")

	exported := executor.ExportState()
	require.NoError(t, ValidateExportedState(exported))
	imported, err := ImportState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported.Receipts, imported.ExportState().Receipts))
	for _, receipt := range imported.Receipts() {
		require.NoError(t, ValidateExecutionReceipt(receipt))
	}
}
