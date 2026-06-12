package types

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/stretchr/testify/require"
)

func TestActorRuntimePlanModelsMailboxIsolationEmissionsAndContinuations(t *testing.T) {
	msgA := actorMailboxMessage("service", "counter", 1)
	msgB := actorMailboxMessage("service", "counter", 2)
	emitted := engineAsyncMessage(9, 10, 25)
	plan, err := NewActorRuntimePlan(ActorRuntimePlan{
		Height:	12,
		Actors: []ActorRuntimeActor{{
			ActorID:	"counter",
			CodeRef:	"code/counter/v1",
			StateRoot:	engineHash("counter-state"),
			Mailbox:	[]ActorMailboxMessage{msgB, msgA},
		}},
		Executions: []ActorExecution{{
			ActorID:		"counter",
			MessageSequence:	1,
			Handler:		"receive",
			GasLimit:		100,
			GasUsed:		40,
			StateWrites: []ActorStateWrite{{
				ActorID:	"counter",
				Key:		ActorStateKeyPrefix("counter") + "balance",
				Hash:		engineHash("balance"),
			}},
			EmittedMessages:	[]async.MessageEnvelope{emitted},
			ResultCode:		async.ResultOK,
		}},
		Continuations: []ContinuationRecord{{
			ContinuationID:		"payment-timeout",
			ActorID:		"counter",
			StepIndex:		1,
			PartialStateHash:	engineHash("partial"),
			PartialStateBytes:	32,
			ResumeHeight:		13,
			ExpiryHeight:		20,
			GasReserved:		50,
			Status:			ContinuationStatusScheduled,
		}},
	})
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Equal(t, ComputeActorRuntimePlanRoot(plan), plan.PlanRoot)
	require.Equal(t, uint64(1), plan.Actors[0].Mailbox[0].Sequence)
	require.Equal(t, uint64(2), plan.Actors[0].Mailbox[1].Sequence)

	mutated := plan
	mutated.Executions = append([]ActorExecution(nil), plan.Executions...)
	mutated.Executions[0].GasUsed++
	require.NotEqual(t, plan.PlanRoot, ComputeActorRuntimePlanRoot(mutated))
}

func TestActorRuntimeRejectsCrossActorStateMutationAndDuplicateExecution(t *testing.T) {
	actor := ActorRuntimeActor{
		ActorID:	"counter",
		CodeRef:	"code/counter/v1",
		StateRoot:	engineHash("counter-state"),
		Mailbox:	[]ActorMailboxMessage{actorMailboxMessage("service", "counter", 1)},
	}
	_, err := NewActorRuntimePlan(ActorRuntimePlan{
		Height:	12,
		Actors:	[]ActorRuntimeActor{actor},
		Executions: []ActorExecution{{
			ActorID:		"counter",
			MessageSequence:	1,
			Handler:		"receive",
			GasLimit:		100,
			StateWrites: []ActorStateWrite{{
				ActorID:	"other",
				Key:		ActorStateKeyPrefix("other") + "balance",
				Hash:		engineHash("balance"),
			}},
		}},
	})
	require.ErrorContains(t, err, "another actor")

	_, err = NewActorRuntimePlan(ActorRuntimePlan{
		Height:	12,
		Actors:	[]ActorRuntimeActor{actor},
		Executions: []ActorExecution{
			{ActorID: "counter", MessageSequence: 1, Handler: "receive", GasLimit: 100},
			{ActorID: "counter", MessageSequence: 1, Handler: "receive", GasLimit: 100},
		},
	})
	require.ErrorContains(t, err, "duplicate actor runtime execution")
}

func TestActorRuntimeRejectsInvalidMailboxAndFailedStateCommit(t *testing.T) {
	_, err := NewActorRuntimePlan(ActorRuntimePlan{
		Height:	12,
		Actors: []ActorRuntimeActor{{
			ActorID:	"counter",
			CodeRef:	"code/counter/v1",
			StateRoot:	engineHash("counter-state"),
			Mailbox:	[]ActorMailboxMessage{actorMailboxMessage("service", "other", 1)},
		}},
	})
	require.ErrorContains(t, err, "target must match actor")

	_, err = NewActorRuntimePlan(ActorRuntimePlan{
		Height:	12,
		Actors: []ActorRuntimeActor{{
			ActorID:	"counter",
			CodeRef:	"code/counter/v1",
			StateRoot:	engineHash("counter-state"),
			Mailbox: []ActorMailboxMessage{
				actorMailboxMessage("service", "counter", 1),
				actorMailboxMessage("service", "counter", 1),
			},
		}},
	})
	require.ErrorContains(t, err, "duplicate actor runtime mailbox sequence")

	_, err = NewActorRuntimePlan(ActorRuntimePlan{
		Height:	12,
		Actors: []ActorRuntimeActor{{
			ActorID:	"counter",
			CodeRef:	"code/counter/v1",
			StateRoot:	engineHash("counter-state"),
			Mailbox:	[]ActorMailboxMessage{actorMailboxMessage("service", "counter", 1)},
		}},
		Executions: []ActorExecution{{
			ActorID:		"counter",
			MessageSequence:	1,
			Handler:		"receive",
			GasLimit:		100,
			StateWrites: []ActorStateWrite{{
				ActorID:	"counter",
				Key:		ActorStateKeyPrefix("counter") + "balance",
				Hash:		engineHash("balance"),
			}},
			Error:	"deterministic handler error",
		}},
	})
	require.ErrorContains(t, err, "must not commit")
}

func TestContinuationStorageRequiresSchedulerResumeAndExpiryReceipt(t *testing.T) {
	actor := ActorRuntimeActor{
		ActorID:	"counter",
		CodeRef:	"code/counter/v1",
		StateRoot:	engineHash("counter-state"),
	}
	_, err := NewActorRuntimePlan(ActorRuntimePlan{
		Height:	12,
		Actors:	[]ActorRuntimeActor{actor},
		Continuations: []ContinuationRecord{{
			ContinuationID:		"resume-counter",
			ActorID:		"counter",
			StepIndex:		1,
			PartialStateHash:	engineHash("partial"),
			PartialStateBytes:	32,
			ResumeHeight:		10,
			ExpiryHeight:		20,
			GasReserved:		50,
			Status:			ContinuationStatusResumed,
			ResumeBy:		"direct-call",
		}},
	})
	require.ErrorContains(t, err, "only through scheduler")

	_, err = NewActorRuntimePlan(ActorRuntimePlan{
		Height:	21,
		Actors:	[]ActorRuntimeActor{actor},
		Continuations: []ContinuationRecord{{
			ContinuationID:		"expired-counter",
			ActorID:		"counter",
			StepIndex:		1,
			PartialStateHash:	engineHash("partial"),
			PartialStateBytes:	32,
			ResumeHeight:		10,
			ExpiryHeight:		20,
			GasReserved:		50,
			Status:			ContinuationStatusScheduled,
		}},
	})
	require.ErrorContains(t, err, "failure receipt")

	expired, err := NewActorRuntimePlan(ActorRuntimePlan{
		Height:	21,
		Actors:	[]ActorRuntimeActor{actor},
		Continuations: []ContinuationRecord{{
			ContinuationID:		"expired-counter",
			ActorID:		"counter",
			StepIndex:		1,
			PartialStateHash:	engineHash("partial"),
			PartialStateBytes:	32,
			ResumeHeight:		10,
			ExpiryHeight:		20,
			GasReserved:		50,
			Status:			ContinuationStatusExpired,
			FailureReceipt: async.ExecutionReceipt{
				Sequence:	1,
				ResultCode:	async.ResultExpired,
				GasUsed:	1,
				Error:		"continuation expired",
			},
		}},
	})
	require.NoError(t, err)
	require.NoError(t, expired.Validate())
}

func actorMailboxMessage(source, target string, sequence uint64) ActorMailboxMessage {
	msg := engineAsyncMessage(byte(sequence), sequence, 20)
	return ActorMailboxMessage{
		Sequence:		sequence,
		SourceActor:		source,
		TargetActor:		target,
		CreatedLogicalTime:	sequence,
		Envelope:		msg,
	}
}
