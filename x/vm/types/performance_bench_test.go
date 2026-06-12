package types

import (
	"fmt"
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func BenchmarkAVMQueueInsertAndPop(b *testing.B) {
	messages := make([]AVMAsyncMessage, 128)
	for i := range messages {
		messages[i] = benchAVMMessage(fmt.Sprintf("bench-sender-%03d", i), uint64(i+1), 10, 0, 100, 20)
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		queue, _ := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
		for _, msg := range messages {
			var err error
			queue, _, err = AdmitAVMZoneQueueMessage(queue, msg, 10, 256)
			if err != nil {
				b.Fatal(err)
			}
		}
		if _, err := SelectAVMZoneQueueWork(queue, messages, 11, zonestypes.ZoneExecutionBudget{MaxGas: 1_000_000, MaxMessages: 128}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAVMActorExecution(b *testing.B) {
	actor := ActorRuntimeActor{
		ActorID:	"bench-actor",
		CodeRef:	"code/bench/v1",
		StateRoot:	engineHash("bench-actor-state"),
		Mailbox:	[]ActorMailboxMessage{benchActorMailboxMessage("service", "bench-actor", 1)},
	}
	execution := ActorExecution{
		ActorID:		"bench-actor",
		MessageSequence:	1,
		Handler:		"handle",
		GasLimit:		1000,
		GasUsed:		100,
		StateWrites: []ActorStateWrite{{
			ActorID:	"bench-actor",
			Key:		ActorStateKeyPrefix("bench-actor") + "counter",
			Hash:		engineHash("counter"),
		}},
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		if _, err := NewActorRuntimePlan(ActorRuntimePlan{Height: 20, Actors: []ActorRuntimeActor{actor}, Executions: []ActorExecution{execution}}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAVMContinuationResume(b *testing.B) {
	actor := ActorRuntimeActor{
		ActorID:	"bench-actor",
		CodeRef:	"code/bench/v1",
		StateRoot:	engineHash("bench-actor-state"),
	}
	continuation := ContinuationRecord{
		ContinuationID:		"bench-continuation",
		ActorID:		actor.ActorID,
		StepIndex:		1,
		PartialStateHash:	engineHash("partial"),
		PartialStateBytes:	64,
		ResumeHeight:		21,
		ExpiryHeight:		30,
		GasReserved:		1000,
		Status:			ContinuationStatusResumed,
		ResumeBy:		ContinuationResumeByScheduler,
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		if _, err := NewActorRuntimePlan(ActorRuntimePlan{Height: 21, Actors: []ActorRuntimeActor{actor}, Continuations: []ContinuationRecord{continuation}}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAVMCrossZoneMessages(b *testing.B) {
	messages := make([]AVMAsyncMessage, 64)
	for i := range messages {
		messages[i] = benchAVMMessage(fmt.Sprintf("cross-zone-%03d", i), uint64(i+1), 10, 0, 100, 50)
	}
	queue, _ := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	policy := AVMCrossZoneRoutePolicy{
		SourceZone:		zonestypes.ZoneIDApplication,
		DestinationZone:	zonestypes.ZoneIDContract,
		GasPolicy:		zonestypes.DefaultZoneGasPolicy(),
		ExecutionBudget:	zonestypes.ZoneExecutionBudget{MaxGas: 1_000_000, MaxMessages: 128},
		MessageFilter:		zonestypes.ZoneMessageFilter{AllowedMessageTypes: []string{"contract.call"}},
		AllowedOpcodes:		[]string{"contract.call"},
		BounceBehavior:		AVMCrossZoneBounceAllowed,
		ProofRequirement:	AVMCrossZoneProofNone,
		ValueAccounting:	AVMCrossZoneValueMessage,
	}
	policy.PolicyHash = ComputeAVMCrossZoneRoutePolicyHash(policy)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		next := queue
		for _, msg := range messages {
			var err error
			next, _, err = AdmitAVMCrossZoneMessage(next, msg, 10, 128, policy)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkAVMStoreV2ReadWriteLatency(b *testing.B) {
	messages := make([]AVMAsyncMessage, 256)
	for i := range messages {
		messages[i] = benchAVMMessage(fmt.Sprintf("store-%03d", i), uint64(i+1), 10, 0, 100, 10)
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		store := make(map[string]AVMStoreV2MessageRecord, len(messages))
		for _, msg := range messages {
			record, _, err := NewAVMStoreV2MessageRecord(msg)
			if err != nil {
				b.Fatal(err)
			}
			store[record.MessageKey] = record
		}
		for _, msg := range messages {
			if _, found := store[AVMAsyncMessageKey(msg.ID)]; !found {
				b.Fatal("missing Store v2 message record")
			}
		}
	}
}

func BenchmarkAVMRootGeneration(b *testing.B) {
	root := AVMRoot{
		Height:			100,
		RouterRoot:		engineHash("router"),
		AsyncMessageRoot:	engineHash("async"),
		ActorRoot:		engineHash("actor"),
		ContractRoot:		engineHash("contract"),
		ContinuationRoot:	engineHash("continuation"),
		InterfaceRoot:		engineHash("interface"),
		ReceiptRoot:		engineHash("receipt"),
	}
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		root.Height = uint64(n + 1)
		if _, err := NewAVMRoot(root); err != nil {
			b.Fatal(err)
		}
	}
}

func benchAVMMessage(source string, nonce, createdHeight, delayHeight, expiryHeight, gasLimit uint64) AVMAsyncMessage {
	msg := AVMAsyncMessage{
		ChainID:		"aetra-1",
		Source:			source,
		Destination:		"contract",
		Payload:		[]byte("payload"),
		GasLimit:		gasLimit,
		DelayHeight:		delayHeight,
		ExpiryHeight:		expiryHeight,
		RetryPolicy:		DefaultAVMRetryPolicy(expiryHeight),
		BounceFlag:		true,
		SourceZone:		zonestypes.ZoneIDApplication,
		DestinationZone:	zonestypes.ZoneIDContract,
		SenderNonce:		nonce,
		PayloadType:		"contract.call",
		ValueNAET:		1,
		ForwardingFee:		1,
		Priority:		1,
		CreatedHeight:		createdHeight,
	}
	built, err := NewAVMAsyncMessage(msg)
	if err != nil {
		panic(err)
	}
	return built
}

func benchActorMailboxMessage(source, target string, sequence uint64) ActorMailboxMessage {
	return ActorMailboxMessage{
		Sequence:		sequence,
		SourceActor:		source,
		TargetActor:		target,
		CreatedLogicalTime:	sequence,
		Envelope:		async.MessageEnvelope{GasLimit: 100},
	}
}
