package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestSyncExecutionPlanModelsCosmosBlockBoundPipeline(t *testing.T) {
	plan, err := NewSyncExecutionPlan(SyncExecutionPlan{
		Height:		10,
		Module:		SyncModuleBank,
		Route:		"bank.transfer",
		GasLimit:	1_000,
		GasUsed:	500,
		Atomic:		true,
		ReceiptPolicy:	SyncReceiptCommitted,
		Steps:		syncPipelineSteps(),
		StateWrites: []SyncStateWrite{
			{StoreKey: "bank", Key: "balances/alice"},
			{StoreKey: "bank", Key: "balances/bob"},
		},
		Events:		[]string{"coin_received", "coin_spent"},
		ResultCode:	async.ResultOK,
	})
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Equal(t, ComputeSyncExecutionReceiptRoot(plan), plan.ReceiptRoot)

	mutated := plan
	mutated.GasUsed++
	require.NotEqual(t, plan.ReceiptRoot, ComputeSyncExecutionReceiptRoot(mutated))
}

func TestSyncExecutionPlanRejectsNonAtomicFailedStateAndBadStageOrder(t *testing.T) {
	failed := SyncExecutionPlan{
		Height:		10,
		Module:		SyncModuleDEX,
		Route:		"dex.swap",
		GasLimit:	1_000,
		GasUsed:	100,
		Atomic:		true,
		ReceiptPolicy:	SyncReceiptCommitted,
		Steps:		syncPipelineSteps(),
		StateWrites:	[]SyncStateWrite{{StoreKey: "dex", Key: "pool/1"}},
		Error:		"deterministic slippage error",
	}
	failed.ReceiptRoot = ComputeSyncExecutionReceiptRoot(failed)
	require.ErrorContains(t, failed.Validate(), "must not commit")

	badOrder, err := NewSyncExecutionPlan(SyncExecutionPlan{
		Height:		10,
		Module:		SyncModuleStaking,
		Route:		"staking.delegate",
		GasLimit:	1_000,
		GasUsed:	100,
		Atomic:		true,
		ReceiptPolicy:	SyncReceiptCommitted,
		Steps: []SyncExecutionStep{
			{Stage: SyncStageTx, Detail: "tx"},
			{Stage: SyncStageMsgServer, Detail: "msgserver"},
			{Stage: SyncStageAnte, Detail: "ante"},
			{Stage: SyncStageKeeper, Detail: "keeper"},
			{Stage: SyncStageStore, Detail: "store"},
			{Stage: SyncStageEvents, Detail: "events"},
			{Stage: SyncStageReceipt, Detail: "receipt"},
		},
	})
	require.ErrorContains(t, err, "stage 1")
	require.NotEmpty(t, badOrder.ReceiptRoot)
}

func TestAsyncExecutionPlanModelsQueuesRetryDeadLettersAndContinuations(t *testing.T) {
	msgA := engineAsyncMessage(1, 2, 20)
	msgB := engineAsyncMessage(2, 3, 30)
	receipt := async.ExecutionReceipt{
		Sequence:	1,
		Source:		msgA.Source,
		Destination:	msgA.Destination,
		Opcode:		msgA.Opcode,
		QueryID:	msgA.QueryID,
		ResultCode:	async.ResultExecutionFailed,
		GasUsed:	20,
		RetryCount:	1,
		RetryScheduled:	true,
		Error:		"handler unavailable",
	}
	dead := async.DeadLetter{
		Sequence:	1,
		FailedSequence:	1,
		RecordedBlock:	12,
		Envelope:	msgA,
		Receipt:	receipt,
		Reason:		"handler unavailable",
	}
	plan, err := NewAsyncExecutionPlan(AsyncExecutionPlan{
		Height:	12,
		Queues: []AsyncZoneQueue{{
			ZoneID:		zonestypes.ZoneIDApplication,
			QueueID:	"application/async",
			Lane:		AsyncEngineLaneScheduled,
			Messages:	[]async.MessageEnvelope{msgB, msgA},
			MaxMessages:	10,
			MaxGas:		100,
		}},
		RetryPolicy: AsyncRetryPolicy{
			MaxRetries:		2,
			RetryDelayBlocks:	1,
			DeadlineBlock:		20,
			Bounce:			true,
			DeadLetter:		true,
		},
		Receipts:	[]async.ExecutionReceipt{receipt},
		DeadLetters:	[]async.DeadLetter{dead},
		Continuations: []AsyncContinuation{{
			Token:		"resume-counter",
			ZoneID:		zonestypes.ZoneIDApplication,
			Contract:	"counter",
			DeliverAtBlock:	13,
			DeadlineBlock:	20,
			StateRoot:	engineHash("continuation-state"),
		}},
	})
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Equal(t, ComputeAsyncExecutionPlanRoot(plan), plan.PlanRoot)
	require.Equal(t, uint64(2), plan.Queues[0].Messages[0].QueryID)
	require.Equal(t, uint64(3), plan.Queues[0].Messages[1].QueryID)

	mutatedReceipt := plan
	mutatedReceipt.Receipts = append([]async.ExecutionReceipt(nil), plan.Receipts...)
	mutatedReceipt.Receipts[0].QueryID++
	require.NotEqual(t, plan.PlanRoot, ComputeAsyncExecutionPlanRoot(mutatedReceipt))

	mutated := plan
	mutated.RetryPolicy.RetryDelayBlocks++
	require.NotEqual(t, plan.PlanRoot, ComputeAsyncExecutionPlanRoot(mutated))
}

func TestAsyncExecutionPlanRejectsUnboundedRetryExpiredMessagesAndInvalidReceipts(t *testing.T) {
	_, err := NewAsyncExecutionPlan(AsyncExecutionPlan{
		Height:	12,
		RetryPolicy: AsyncRetryPolicy{
			DeadLetter: true,
		},
	})
	require.ErrorContains(t, err, "bounded retry")

	expired := engineAsyncMessage(1, 2, 20)
	expired.DeadlineBlock = 11
	_, err = NewAsyncExecutionPlan(AsyncExecutionPlan{
		Height:	12,
		Queues: []AsyncZoneQueue{{
			ZoneID:		zonestypes.ZoneIDApplication,
			QueueID:	"application/async",
			Lane:		AsyncEngineLaneCrossZone,
			Messages:	[]async.MessageEnvelope{expired},
			MaxMessages:	10,
			MaxGas:		100,
		}},
		RetryPolicy:	AsyncRetryPolicy{MaxRetries: 1, RetryDelayBlocks: 1, DeadLetter: true},
	})
	require.ErrorContains(t, err, "expired")

	successRetry := async.ExecutionReceipt{
		Sequence:	1,
		ResultCode:	async.ResultOK,
		GasUsed:	1,
		RetryScheduled:	true,
	}
	_, err = NewAsyncExecutionPlan(AsyncExecutionPlan{
		Height:		12,
		RetryPolicy:	AsyncRetryPolicy{MaxRetries: 1, RetryDelayBlocks: 1},
		Receipts:	[]async.ExecutionReceipt{successRetry},
	})
	require.ErrorContains(t, err, "successful receipt")

	deadMsg := engineAsyncMessage(3, 4, 20)
	deadReceipt := async.ExecutionReceipt{
		Sequence:	1,
		Source:		deadMsg.Source,
		Destination:	deadMsg.Destination,
		Opcode:		deadMsg.Opcode,
		QueryID:	deadMsg.QueryID,
		ResultCode:	async.ResultExecutionFailed,
		GasUsed:	1,
		Error:		"failed",
	}
	_, err = NewAsyncExecutionPlan(AsyncExecutionPlan{
		Height:		12,
		RetryPolicy:	AsyncRetryPolicy{MaxRetries: 1, RetryDelayBlocks: 1},
		DeadLetters: []async.DeadLetter{{
			Sequence:	1,
			FailedSequence:	1,
			RecordedBlock:	12,
			Envelope:	deadMsg,
			Receipt:	deadReceipt,
			Reason:		"failed",
		}},
	})
	require.ErrorContains(t, err, "dead letter policy")
}

func syncPipelineSteps() []SyncExecutionStep {
	return []SyncExecutionStep{
		{Stage: SyncStageTx, Detail: "tx"},
		{Stage: SyncStageAnte, Detail: "ante"},
		{Stage: SyncStageMsgServer, Detail: "msgserver"},
		{Stage: SyncStageKeeper, Detail: "keeper"},
		{Stage: SyncStageStore, Detail: "store"},
		{Stage: SyncStageEvents, Detail: "events"},
		{Stage: SyncStageReceipt, Detail: "receipt"},
	}
}

func engineAsyncMessage(fill byte, queryID uint64, gas uint64) async.MessageEnvelope {
	return async.MessageEnvelope{
		Source:			sdk.AccAddress(bytes.Repeat([]byte{fill}, 20)),
		Destination:		sdk.AccAddress(bytes.Repeat([]byte{fill + 10}, 20)),
		Value:			sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
		Opcode:			uint32(fill),
		QueryID:		queryID,
		Body:			[]byte("body"),
		Bounce:			true,
		CreatedLogicalTime:	queryID,
		DeliverAtBlock:		queryID,
		MaxRetries:		2,
		RetryDelayBlocks:	1,
		DeadlineBlock:		20,
		GasLimit:		gas,
		ForwardFee:		sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
	}
}

func engineHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
