package async

import (
	"bytes"
	"reflect"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestContractEmitsInternalMessageAndRecipientExecutesInOrder(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	contractA := deployTestContract(t, executor, deployer, []byte("a"))
	contractB := deployTestContract(t, executor, deployer, []byte("b"))

	require.NoError(t, executor.RegisterHandler(contractA, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{
			NewState:	[]byte("a:sent"),
			Outgoing: []MessageEnvelope{{
				Destination:	contractB,
				Value:		naetCoin(7),
				Opcode:		20,
				QueryID:	msg.QueryID,
				Body:		[]byte("from-a"),
				Bounce:		true,
				GasLimit:	100_000,
				ForwardFee:	forwardFee(),
			}},
			ResultCode:	ResultOK,
		}
	}))
	require.NoError(t, executor.RegisterHandler(contractB, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{
			NewState:	append([]byte("b:"), msg.Body...),
			ResultCode:	ResultOK,
		}
	}))

	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{{
		Source:			testAddr(9),
		Destination:		contractA,
		Value:			naetCoin(1),
		Opcode:			10,
		QueryID:		99,
		Body:			[]byte("start"),
		Bounce:			true,
		CreatedLogicalTime:	1,
		GasLimit:		100_000,
		ForwardFee:		forwardFee(),
	}}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	require.Equal(t, uint64(0), receipts[0].Sequence)
	require.Equal(t, uint64(1), receipts[1].Sequence)
	require.Equal(t, ResultOK, receipts[0].ResultCode)
	require.Equal(t, ResultOK, receipts[1].ResultCode)

	a, ok := executor.Contract(contractA)
	require.True(t, ok)
	require.Equal(t, []byte("a:sent"), a.State)
	require.Equal(t, uint64(1), a.LogicalTime)
	b, ok := executor.Contract(contractB)
	require.True(t, ok)
	require.Equal(t, []byte("b:from-a"), b.State)
	require.Equal(t, uint64(1), b.LogicalTime)
	require.Equal(t, uint64(2), executor.Metrics().ProcessedMessages)
	queue := executor.Receipts()
	require.Equal(t, uint64(0), queue[0].Sequence)
	require.Equal(t, uint64(1), queue[1].Sequence)
}

func TestExecutionEconomyFeedsProtocolLoop(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	contract := deployTestContract(t, executor, deployer, []byte("economy"))
	params := DefaultParams()

	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{
			NewState:	[]byte("abcde"),
			ResultCode:	ResultOK,
		}
	}))
	msg := testMessage(testAddr(9), contract, 1)
	msg.Value = naetCoin(0)
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultOK, receipts[0].ResultCode)

	activity := executor.EconomicActivity()
	require.Equal(t, params.ContractDeploymentCost, activity.AVMDeploymentCostNaet)
	require.Equal(t, params.StorageFeePerByte.MulRaw(5), activity.AVMStorageFeeNaet)
	require.Equal(t, params.ForwardingFee, activity.AVMForwardingFeeNaet)

	control, err := appparams.BalanceController(appparams.BalanceControllerInput{
		CurrentInflationBps:	appparams.DefaultTargetInflationBps,
		StakeRatioBps:		appparams.DefaultTargetStakeBps,
		BlockLoadBps:		appparams.DefaultTargetLoadBps,
		AnnualMint:		sdkmath.NewInt(100),
		AnnualBurn:		sdkmath.NewInt(100),
		Activity:		activity,
	})
	require.NoError(t, err)
	require.True(t, control.DeflationGuardActive)

	flow, err := executor.EconomicFlow(control)
	require.NoError(t, err)
	require.Equal(t, activity.TotalCharges(), flow.TotalChargesNaet)
	require.Equal(t, flow.TotalChargesNaet, flow.BurnNaet.Add(flow.TreasuryNaet).Add(flow.ValidatorRewardsNaet))
}

func TestFailedSendProducesDeterministicBounceWithoutDestinationMutation(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	source := deployTestContract(t, executor, deployer, []byte("source"))
	missingDest, err := DeriveContractAddress(deployer, testCodeHash(9), []byte("missing"))
	require.NoError(t, err)

	require.NoError(t, executor.RegisterHandler(source, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		if msg.Bounced {
			return ExecutionResult{
				NewState:	[]byte("source:bounced"),
				ResultCode:	ResultOK,
			}
		}
		return ExecutionResult{
			NewState:	[]byte("source:original"),
			ResultCode:	ResultOK,
		}
	}))

	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{{
		Source:			source,
		Destination:		missingDest,
		Value:			naetCoin(5),
		Opcode:			30,
		QueryID:		123,
		Body:			[]byte("will-bounce"),
		Bounce:			true,
		CreatedLogicalTime:	1,
		GasLimit:		100_000,
		ForwardFee:		forwardFee(),
	}}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	require.Equal(t, ResultNoDestination, receipts[0].ResultCode)
	require.True(t, receipts[0].BounceCreated)
	require.True(t, receipts[0].Refunded)
	require.Equal(t, sdkmath.NewInt(4), receipts[0].RefundAmountNaet)
	require.Equal(t, sdkmath.NewInt(1), receipts[0].RefundFeeNaet)
	require.Equal(t, receipts[1].Sequence, receipts[0].RefundOfSequence)
	require.Equal(t, sdkmath.NewInt(5), receipts[0].RefundAmountNaet.Add(receipts[0].RefundFeeNaet))
	require.True(t, executor.Metrics().BouncedMessages > 0)
	require.Equal(t, BounceOpcode, receipts[1].Opcode)
	require.True(t, receipts[1].Bounced)
	require.True(t, receipts[1].Source.Equals(missingDest))
	require.True(t, receipts[1].Destination.Equals(source))

	contract, ok := executor.Contract(source)
	require.True(t, ok)
	require.Equal(t, []byte("source:bounced"), contract.State)
	_, exists := executor.Contract(missingDest)
	require.False(t, exists)
}

func TestUnderfundedStorageRentFreezesAsyncContractAndPreservesState(t *testing.T) {
	params := DefaultParams()
	params.StorageFeePerByte = sdkmath.NewInt(10)
	params.ContractDeploymentCost = sdkmath.NewInt(1)
	executor, err := NewExecutor(params)
	require.NoError(t, err)
	contract, err := executor.DeployContract(testAddr(1), testCodeHash(3), []byte("rent"), []byte("init"), sdkmath.NewInt(2))
	require.NoError(t, err)
	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{NewState: []byte("expensive-state"), ResultCode: ResultOK}
	}))

	msg := testMessage(testAddr(9), contract, 1)
	msg.Value = naetCoin(0)
	msg.Bounce = false
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultExecutionFailed, receipts[0].ResultCode)
	require.Contains(t, receipts[0].Error, "contract frozen by storage rent")

	frozen, found := executor.Contract(contract)
	require.True(t, found)
	require.Equal(t, ContractStatusFrozen, frozen.Status)
	require.Equal(t, []byte("init"), frozen.State)
	require.True(t, frozen.BalanceNaet.IsZero())
	require.True(t, frozen.StorageRentDebtNaet.IsPositive())

	msg.QueryID = 2
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))
	receipts, err = executor.ProcessBlock(2)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultExecutionFailed, receipts[0].ResultCode)
	require.Contains(t, receipts[0].Error, "frozen by storage rent")
}

func TestQueueLimitsPreventDoS(t *testing.T) {
	params := DefaultParams()
	params.MaxMessagesPerTx = 2
	params.MaxMessagesPerBlock = 1
	params.MaxBodySize = 8
	executor, err := NewExecutor(params)
	require.NoError(t, err)
	contract := deployTestContract(t, executor, testAddr(1), []byte("limited"))
	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{NewState: contract.State, ResultCode: ResultOK}
	}))

	tooMany := []MessageEnvelope{
		testMessage(testAddr(9), contract, 1),
		testMessage(testAddr(9), contract, 2),
		testMessage(testAddr(9), contract, 3),
	}
	require.ErrorContains(t, executor.EnqueueTxMessages(tooMany), "messages per tx")

	large := testMessage(testAddr(9), contract, 1)
	large.Body = []byte("too-large")
	require.ErrorContains(t, executor.EnqueueTxMessages([]MessageEnvelope{large}), "message body size")

	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{
		testMessage(testAddr(9), contract, 1),
		testMessage(testAddr(9), contract, 2),
	}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Len(t, executor.Queue(), 1)
	require.Equal(t, uint64(1), executor.Metrics().QueueLag)
}

func TestDelayedMessagesWaitForReadyBlockAndSurviveExportImport(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	contract := deployTestContract(t, executor, deployer, []byte("delay"))
	require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{NewState: append([]byte("seen:"), msg.Body...), ResultCode: ResultOK}
	}))

	delayed := testMessage(testAddr(9), contract, 1)
	delayed.Body = []byte("delayed")
	delayed.DeliverAtBlock = 3
	immediate := testMessage(testAddr(9), contract, 2)
	immediate.Body = []byte("immediate")
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{delayed, immediate}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, uint64(2), receipts[0].QueryID)
	require.Len(t, executor.Queue(), 1)
	require.Equal(t, uint64(3), executor.Queue()[0].Envelope.DeliverAtBlock)
	require.Zero(t, executor.Metrics().QueueLag)

	exported := executor.ExportState()
	imported, err := ImportState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported, imported.ExportState()))
	require.NoError(t, imported.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{NewState: append([]byte("seen:"), msg.Body...), ResultCode: ResultOK}
	}))

	receipts, err = imported.ProcessBlock(2)
	require.NoError(t, err)
	require.Empty(t, receipts)
	require.Len(t, imported.Queue(), 1)

	receipts, err = imported.ProcessBlock(3)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, uint64(1), receipts[0].QueryID)
	updated, ok := imported.Contract(contract)
	require.True(t, ok)
	require.Equal(t, []byte("seen:delayed"), updated.State)
}

func TestRetryPolicySchedulesBoundedRetriesBeforeBounceAndDeadLetter(t *testing.T) {
	params := DefaultParams()
	params.MaxMessagesPerBlock = 1
	executor, err := NewExecutor(params)
	require.NoError(t, err)
	deployer := testAddr(1)
	source := deployTestContract(t, executor, deployer, []byte("source"))
	missingDest, err := DeriveContractAddress(deployer, testCodeHash(9), []byte("missing-retry"))
	require.NoError(t, err)
	require.NoError(t, executor.RegisterHandler(source, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		if msg.Bounced {
			return ExecutionResult{NewState: []byte("source:bounced-after-retry"), ResultCode: ResultOK}
		}
		return ExecutionResult{NewState: contract.State, ResultCode: ResultOK}
	}))

	msg := testMessage(source, missingDest, 77)
	msg.MaxRetries = 2
	msg.RetryDelayBlocks = 1
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultNoDestination, receipts[0].ResultCode)
	require.True(t, receipts[0].RetryScheduled)
	require.Len(t, executor.Queue(), 1)
	require.Equal(t, uint32(1), executor.Queue()[0].Envelope.RetryCount)
	require.Equal(t, uint64(2), executor.Queue()[0].Envelope.DeliverAtBlock)
	require.Empty(t, executor.DeadLetters())

	receipts, err = executor.ProcessBlock(2)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.True(t, receipts[0].RetryScheduled)
	require.Equal(t, uint32(2), executor.Queue()[0].Envelope.RetryCount)
	require.Equal(t, uint64(3), executor.Queue()[0].Envelope.DeliverAtBlock)

	receipts, err = executor.ProcessBlock(3)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.False(t, receipts[0].RetryScheduled)
	require.Len(t, executor.Queue(), 1)
	require.Equal(t, BounceOpcode, executor.Queue()[0].Envelope.Opcode)
	require.Len(t, executor.DeadLetters(), 1)
	require.Equal(t, receipts[0].Sequence, executor.DeadLetters()[0].FailedSequence)
	require.Equal(t, uint64(2), executor.Metrics().RetriedMessages)
	require.Equal(t, uint64(1), executor.Metrics().DeadLetterMessages)

	receipts, err = executor.ProcessBlock(4)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, BounceOpcode, receipts[0].Opcode)
	contract, ok := executor.Contract(source)
	require.True(t, ok)
	require.Equal(t, []byte("source:bounced-after-retry"), contract.State)
}

func TestRetryDeadlinePreventsSchedulingAndRecordsDeadLetter(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	source := deployTestContract(t, executor, deployer, []byte("deadline"))
	missingDest, err := DeriveContractAddress(deployer, testCodeHash(8), []byte("deadline-missing"))
	require.NoError(t, err)

	msg := testMessage(source, missingDest, 88)
	msg.MaxRetries = 1
	msg.RetryDelayBlocks = 2
	msg.DeadlineBlock = 1
	msg.Bounce = false
	msg.Value = naetCoin(0)
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.False(t, receipts[0].RetryScheduled)
	require.Empty(t, executor.Queue())
	require.Len(t, executor.DeadLetters(), 1)
	require.Equal(t, "destination contract not found", executor.DeadLetters()[0].Reason)
}

func TestMailboxViewsAreCanonicalAndValidatedOnImport(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	destA := deployTestContract(t, executor, deployer, []byte("mba"))
	destB := deployTestContract(t, executor, deployer, []byte("mbb"))
	source := testAddr(9)
	delayed := testMessage(source, destA, 1)
	delayed.DeliverAtBlock = 5
	immediate := testMessage(source, destA, 2)
	other := testMessage(source, destB, 3)
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{delayed, immediate, other}))

	inbox := executor.Inbox(destA)
	require.Len(t, inbox, 2)
	require.Equal(t, uint64(2), inbox[0].Envelope.QueryID)
	require.Equal(t, uint64(1), inbox[1].Envelope.QueryID)
	outbox := executor.Outbox(source)
	require.Len(t, outbox, 3)
	require.Equal(t, uint64(2), outbox[0].Envelope.QueryID)

	exported := executor.ExportState()
	imported, err := ImportState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported, imported.ExportState()))

	corrupted := exported
	corrupted.Inbox = cloneQueuedMap(exported.Inbox)
	corrupted.Inbox[inboxKey(destA)][0].Envelope.Destination = destB
	_, err = ImportState(corrupted)
	require.ErrorContains(t, err, "owner key drift")
}

func TestMessageEnvelopeRejectsDeliveryAfterDeadline(t *testing.T) {
	msg := testMessage(testAddr(1), testAddr(2), 1)
	msg.DeliverAtBlock = 10
	msg.DeadlineBlock = 9
	require.ErrorContains(t, msg.Validate(DefaultParams()), "deliver block")
}

func TestInvalidOutgoingMessageDoesNotCommitRecipientState(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	dest := deployTestContract(t, executor, deployer, []byte("dest"))
	require.NoError(t, executor.RegisterHandler(dest, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		out := testMessage(contract.Address, dest, 99)
		out.DeliverAtBlock = 5
		out.DeadlineBlock = 4
		return ExecutionResult{
			NewState:	[]byte("mutated"),
			Outgoing:	[]MessageEnvelope{out},
			ResultCode:	ResultOK,
		}
	}))
	msg := testMessage(testAddr(9), dest, 1)
	msg.Bounce = false
	msg.Value = naetCoin(0)
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultLimitExceeded, receipts[0].ResultCode)
	require.Empty(t, executor.Queue())
	contract, ok := executor.Contract(dest)
	require.True(t, ok)
	require.Equal(t, []byte("init:dest"), contract.State)
}

func TestFailureRollsBackRecipientStateAndRefundsWhenBounceDisabled(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	source := deployTestContract(t, executor, deployer, []byte("source"))
	dest := deployTestContract(t, executor, deployer, []byte("dest"))

	require.NoError(t, executor.RegisterHandler(source, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		if msg.Opcode == RefundOpcode {
			return ExecutionResult{NewState: []byte("source:refund"), ResultCode: ResultOK}
		}
		return ExecutionResult{NewState: contract.State, ResultCode: ResultOK}
	}))
	require.NoError(t, executor.RegisterHandler(dest, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
		return ExecutionResult{
			NewState:	[]byte("dest:mutated-but-rolled-back"),
			ResultCode:	ResultExecutionFailed,
			Error:		"handler failure",
		}
	}))

	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{{
		Source:		source,
		Destination:	dest,
		Value:		naetCoin(3),
		Opcode:		40,
		QueryID:	1,
		Body:		[]byte("fail"),
		Bounce:		false,
		GasLimit:	100_000,
		ForwardFee:	forwardFee(),
	}}))
	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	require.Equal(t, ResultExecutionFailed, receipts[0].ResultCode)
	require.False(t, receipts[0].BounceCreated)
	require.True(t, receipts[0].RefundCreated)
	require.True(t, receipts[0].Refunded)
	require.Equal(t, sdkmath.NewInt(2), receipts[0].RefundAmountNaet)
	require.Equal(t, sdkmath.NewInt(1), receipts[0].RefundFeeNaet)
	require.Equal(t, RefundOpcode, receipts[1].Opcode)

	destContract, ok := executor.Contract(dest)
	require.True(t, ok)
	require.Equal(t, []byte("init:dest"), destContract.State)
	sourceContract, ok := executor.Contract(source)
	require.True(t, ok)
	require.Equal(t, []byte("source:refund"), sourceContract.State)
	require.Equal(t, uint64(1), executor.Metrics().RefundMessages)
	require.Equal(t, uint64(1), executor.Metrics().FailedExecutions)
}

func TestExportImportPreservesQueueStateExactly(t *testing.T) {
	executor := newTestExecutor(t)
	contract := deployTestContract(t, executor, testAddr(1), []byte("export"))
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{
		testMessage(testAddr(9), contract, 1),
		testMessage(testAddr(9), contract, 2),
	}))

	exported := executor.ExportState()
	imported, err := ImportState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported, imported.ExportState()))
	require.Equal(t, exported.Queue, imported.Queue())
	require.Equal(t, exported.NextTxIndex, imported.ExportState().NextTxIndex)
	require.Equal(t, uint64(0), exported.Queue[0].TxIndex)
	require.Equal(t, uint32(0), exported.Queue[0].MessageIndex)
	require.Equal(t, uint32(1), exported.Queue[1].MessageIndex)
	require.Equal(t, exported.Queue[0].Envelope.CreatedLogicalTime, exported.Queue[0].SourceLogicalTime)
	require.Equal(t, queueAddressKey(exported.Queue[0].Envelope.Destination), exported.Queue[0].DestinationKey)
}

func TestImportStateRejectsDuplicateAndMalformedContractQueueState(t *testing.T) {
	executor := newTestExecutor(t)
	contract := deployTestContract(t, executor, testAddr(1), []byte("export"))
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{
		testMessage(testAddr(9), contract, 1),
		testMessage(testAddr(9), contract, 2),
	}))
	exported := executor.ExportState()

	t.Run("duplicate contract address", func(t *testing.T) {
		corrupted := exported
		corrupted.Contracts = append(cloneContracts(exported.Contracts), exported.Contracts[0])
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "duplicate contract address")
	})

	t.Run("malformed contract state", func(t *testing.T) {
		corrupted := exported
		corrupted.Contracts = cloneContracts(exported.Contracts)
		corrupted.Contracts[0].State = bytes.Repeat([]byte{1}, int(DefaultParams().MaxStateSize)+1)
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "contract state size")
	})

	t.Run("duplicate queued sequence", func(t *testing.T) {
		corrupted := exported
		corrupted.Queue = append(cloneQueuedMessagesForTest(exported.Queue), exported.Queue[0])
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "duplicate queued message sequence")
	})

	t.Run("queue sequence drift", func(t *testing.T) {
		corrupted := exported
		corrupted.NextSequence = exported.Queue[0].Sequence
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "must be less than next_sequence")
	})

	t.Run("malformed queued message", func(t *testing.T) {
		corrupted := exported
		corrupted.Queue = cloneQueuedMessagesForTest(exported.Queue)
		corrupted.Queue[0].Envelope.Source = sdk.AccAddress(make([]byte, 20))
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "must not be zero")
	})

	t.Run("queue ordering drift", func(t *testing.T) {
		corrupted := exported
		corrupted.Queue = cloneQueuedMessagesForTest(exported.Queue)
		corrupted.Queue[0], corrupted.Queue[1] = corrupted.Queue[1], corrupted.Queue[0]
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "sorted")
	})

	t.Run("queue logical time drift", func(t *testing.T) {
		corrupted := exported
		corrupted.Queue = cloneQueuedMessagesForTest(exported.Queue)
		corrupted.Queue[0].SourceLogicalTime++
		corrupted.Queue[0].MessageID = QueueMessageID(corrupted.Queue[0])
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "source logical time drift")
	})

	t.Run("queue destination key drift", func(t *testing.T) {
		corrupted := exported
		corrupted.Queue = cloneQueuedMessagesForTest(exported.Queue)
		corrupted.Queue[0].DestinationKey = "wrong"
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "destination key drift")
	})

	t.Run("runtime execution block height is not importable queue state", func(t *testing.T) {
		corrupted := exported
		corrupted.Queue = cloneQueuedMessagesForTest(exported.Queue)
		corrupted.Queue[0].Envelope.ExecutionBlockHeight = 99
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "execution block height")
	})

	t.Run("tx index drift", func(t *testing.T) {
		corrupted := exported
		corrupted.Queue = cloneQueuedMessagesForTest(exported.Queue)
		corrupted.Queue[0].TxIndex = exported.NextTxIndex
		_, err := ImportState(corrupted)
		require.ErrorContains(t, err, "next_tx_index")
	})
}

func TestMessageEnvelopeRequiresGasLimitAndNaetForwardFee(t *testing.T) {
	params := DefaultParams()
	msg := testMessage(testAddr(1), testAddr(2), 1)
	msg.GasLimit = 0
	require.ErrorContains(t, msg.Validate(params), "gas limit")

	msg = testMessage(testAddr(1), testAddr(2), 1)
	msg.ForwardFee = sdk.NewInt64Coin("uatom", 1)
	require.ErrorContains(t, msg.Validate(params), "forward fee denom")

	msg = testMessage(testAddr(1), testAddr(2), 1)
	msg.MaxRetries = params.MaxRetriesPerMessage + 1
	require.ErrorContains(t, msg.Validate(params), "max retries")

	msg = testMessage(testAddr(1), testAddr(2), 1)
	msg.RetryDelayBlocks = params.MaxRetryDelayBlocks + 1
	require.ErrorContains(t, msg.Validate(params), "retry delay")
}

func TestRefundMessageFailureDoesNotCreateDoubleRefundCycle(t *testing.T) {
	executor := newTestExecutor(t)
	missingSource := testAddr(8)
	missingDestination := testAddr(9)
	msg := testMessage(missingSource, missingDestination, 1)
	msg.Bounce = false
	msg.Opcode = RefundOpcode
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultRefundSuppressed, receipts[0].ResultCode)
	require.False(t, receipts[0].Refunded)
	require.Empty(t, executor.Queue())
	require.Zero(t, executor.Metrics().RefundMessages)
}

func TestBouncedMessageFailureDoesNotCreateBounceLoop(t *testing.T) {
	executor := newTestExecutor(t)
	missingSource := testAddr(8)
	missingDestination := testAddr(9)
	msg := testMessage(missingSource, missingDestination, 1)
	msg.Opcode = BounceOpcode
	msg.Bounce = false
	msg.Bounced = true
	msg.Value = naetCoin(5)
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Equal(t, ResultBounceSuppressed, receipts[0].ResultCode)
	require.False(t, receipts[0].Refunded)
	require.Empty(t, executor.Queue())
	require.Zero(t, executor.Metrics().BouncedMessages)
	require.Zero(t, executor.Metrics().RefundMessages)
}

func TestMarkRefundedRejectsDoubleRefund(t *testing.T) {
	receipt := ExecutionReceipt{Sequence: 7}
	refund := RefundCalculation{Amount: sdkmath.NewInt(4), Fee: sdkmath.NewInt(1)}
	require.NoError(t, MarkRefunded(&receipt, refund, "bounce", 8))
	require.True(t, receipt.Refunded)
	require.ErrorContains(t, MarkRefunded(&receipt, refund, "refund", 9), "already refunded")
}

func TestExportImportPreservesBounceRefundStatus(t *testing.T) {
	params := DefaultParams()
	params.MaxMessagesPerBlock = 1
	executor, err := NewExecutor(params)
	require.NoError(t, err)
	deployer := testAddr(1)
	source := deployTestContract(t, executor, deployer, []byte("source"))
	missingDest, err := DeriveContractAddress(deployer, testCodeHash(10), []byte("missing-export-bounce"))
	require.NoError(t, err)
	msg := testMessage(source, missingDest, 42)
	msg.Value = naetCoin(5)
	msg.Bounce = true
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))

	receipts, err := executor.ProcessBlock(1)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.True(t, receipts[0].BounceCreated)
	require.True(t, receipts[0].Refunded)
	require.Equal(t, sdkmath.NewInt(4), receipts[0].RefundAmountNaet)
	require.Len(t, executor.Queue(), 1)
	require.True(t, executor.Queue()[0].Envelope.Bounced)
	require.False(t, executor.Queue()[0].Envelope.Bounce)
	require.Equal(t, receipts[0].Sequence, executor.Queue()[0].Envelope.RefundOfSequence)

	exported := executor.ExportState()
	imported, err := ImportState(exported)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(exported.Receipts, imported.ExportState().Receipts))
	require.True(t, imported.Queue()[0].Envelope.Bounced)
	require.Equal(t, receipts[0].Sequence, imported.Queue()[0].Envelope.RefundOfSequence)
}

func TestExecutionLimitsRejectGasEmittedMessagesAndStorageWrites(t *testing.T) {
	params := DefaultParams()
	params.MaxEmittedMessagesPerExec = 1
	params.MaxStorageWritesPerExec = 1

	t.Run("gas limit", func(t *testing.T) {
		executor, err := NewExecutor(params)
		require.NoError(t, err)
		contract := deployTestContract(t, executor, testAddr(1), []byte("gas"))
		require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
			return ExecutionResult{NewState: contract.State, GasUsed: msg.GasLimit + 1, ResultCode: ResultOK}
		}))
		msg := testMessage(testAddr(9), contract, 1)
		msg.GasLimit = 10
		msg.Bounce = false
		msg.Value = naetCoin(0)
		require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))
		receipts, err := executor.ProcessBlock(1)
		require.NoError(t, err)
		require.Len(t, receipts, 1)
		require.Equal(t, ResultLimitExceeded, receipts[0].ResultCode)
		require.Contains(t, receipts[0].Error, "gas limit")
	})

	t.Run("emitted messages", func(t *testing.T) {
		executor, err := NewExecutor(params)
		require.NoError(t, err)
		source := deployTestContract(t, executor, testAddr(1), []byte("emit"))
		dest := deployTestContract(t, executor, testAddr(1), []byte("dest"))
		require.NoError(t, executor.RegisterHandler(source, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
			return ExecutionResult{
				NewState:	contract.State,
				ResultCode:	ResultOK,
				Outgoing: []MessageEnvelope{
					testMessage(contract.Address, dest, 10),
					testMessage(contract.Address, dest, 11),
				},
			}
		}))
		msg := testMessage(testAddr(9), source, 1)
		msg.Bounce = false
		msg.Value = naetCoin(0)
		require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))
		receipts, err := executor.ProcessBlock(1)
		require.NoError(t, err)
		require.Len(t, receipts, 1)
		require.Equal(t, ResultLimitExceeded, receipts[0].ResultCode)
		require.Contains(t, receipts[0].Error, "emitted message limit")
	})

	t.Run("storage writes", func(t *testing.T) {
		executor, err := NewExecutor(params)
		require.NoError(t, err)
		contract := deployTestContract(t, executor, testAddr(1), []byte("stor"))
		require.NoError(t, executor.RegisterHandler(contract, func(contract ContractAccount, msg MessageEnvelope) ExecutionResult {
			return ExecutionResult{NewState: contract.State, StorageWrites: 2, ResultCode: ResultOK}
		}))
		msg := testMessage(testAddr(9), contract, 1)
		msg.Bounce = false
		msg.Value = naetCoin(0)
		require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{msg}))
		receipts, err := executor.ProcessBlock(1)
		require.NoError(t, err)
		require.Len(t, receipts, 1)
		require.Equal(t, ResultLimitExceeded, receipts[0].ResultCode)
		require.Contains(t, receipts[0].Error, "storage write limit")
	})
}

func TestQueueOrderingUsesTxMessageLogicalDestinationAndSequence(t *testing.T) {
	executor := newTestExecutor(t)
	deployer := testAddr(1)
	destA := deployTestContract(t, executor, deployer, []byte("a"))
	destB := deployTestContract(t, executor, deployer, []byte("b"))
	require.NoError(t, executor.EnqueueTxMessages([]MessageEnvelope{
		testMessage(testAddr(9), destB, 2),
		testMessage(testAddr(9), destA, 1),
	}))
	queue := executor.Queue()
	require.Len(t, queue, 2)
	require.Equal(t, uint64(0), queue[0].TxIndex)
	require.Equal(t, uint32(0), queue[0].MessageIndex)
	require.Equal(t, uint64(0), queue[1].TxIndex)
	require.Equal(t, uint32(1), queue[1].MessageIndex)
	require.Equal(t, uint64(0), queue[0].Sequence)
	require.Equal(t, uint64(1), queue[1].Sequence)
}

func TestContractAddressDerivationIsStableAndRejectsMalformedInput(t *testing.T) {
	deployer := testAddr(1)
	codeHash := testCodeHash(1)
	a, err := DeriveContractAddress(deployer, codeHash, []byte("salt"))
	require.NoError(t, err)
	b, err := DeriveContractAddress(deployer, codeHash, []byte("salt"))
	require.NoError(t, err)
	require.Equal(t, a, b)

	_, err = DeriveContractAddress(sdk.AccAddress(make([]byte, 20)), codeHash, []byte("salt"))
	require.ErrorContains(t, err, "must not be zero")
	_, err = DeriveContractAddress(deployer, []byte{1}, []byte("salt"))
	require.ErrorContains(t, err, "contract code hash")
}
