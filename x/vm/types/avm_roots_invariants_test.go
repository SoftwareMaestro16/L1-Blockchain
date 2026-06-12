package types

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMRootCommitsSection43RootSet(t *testing.T) {
	root, err := NewAVMRoot(AVMRoot{
		Height:			42,
		RouterRoot:		engineHash("router"),
		AsyncMessageRoot:	engineHash("async"),
		ActorRoot:		engineHash("actor"),
		ContractRoot:		engineHash("contract"),
		ContinuationRoot:	engineHash("continuation"),
		InterfaceRoot:		engineHash("interface"),
		ReceiptRoot:		engineHash("receipt"),
	})
	require.NoError(t, err)
	require.NoError(t, root.Validate())
	require.Equal(t, ComputeAVMRootHash(root), root.RootHash)

	mutated := root
	mutated.InterfaceRoot = engineHash("interface-mutated")
	require.NotEqual(t, root.RootHash, ComputeAVMRootHash(mutated))

	bad := root
	bad.ActorRoot = "not-a-root"
	bad.RootHash = ComputeAVMRootHash(bad)
	require.ErrorContains(t, bad.Validate(), "actor root")
}

func TestAVMZoneStateRootIncludesStateMessageExecutionAndContinuationRoots(t *testing.T) {
	root, err := NewAVMZoneStateRoot(AVMZoneStateRoot{
		ZoneID:			zonestypes.ZoneIDContract,
		Height:			42,
		StateRoot:		engineHash("state"),
		MessageRoot:		engineHash("message"),
		ExecutionRoot:		engineHash("execution"),
		ContinuationRoot:	engineHash("continuation"),
	})
	require.NoError(t, err)
	require.NoError(t, root.Validate())
	require.Equal(t, ComputeAVMZoneStateRootHash(root), root.RootHash)

	mutated := root
	mutated.ContinuationRoot = engineHash("continuation-mutated")
	require.NotEqual(t, root.RootHash, ComputeAVMZoneStateRootHash(mutated))
}

func TestAVMStateInvariantsAcceptConsistentState(t *testing.T) {
	set := validInvariantSet(t)
	require.NoError(t, set.Validate())

	shuffled := set
	shuffled.QueuedMessages = append([]AVMQueuedMessageRef(nil), set.QueuedMessages...)
	shuffled.QueuedMessages[0], shuffled.QueuedMessages[1] = shuffled.QueuedMessages[1], shuffled.QueuedMessages[0]
	require.Equal(t, ComputeAVMStateInvariantRoot(set), ComputeAVMStateInvariantRoot(shuffled))
	require.NoError(t, shuffled.Validate())
}

func TestAVMStateInvariantsRejectMissingMessageAndReceiptRules(t *testing.T) {
	set := validInvariantSet(t)
	set.QueuedMessages = append(set.QueuedMessages, AVMQueuedMessageRef{
		MessageID:	"missing-message",
		ZoneID:		zonestypes.ZoneIDContract,
		QueueID:	"default",
		SortKey:	"0009",
	})
	require.ErrorContains(t, set.Validate(), "no stored message")

	set = validInvariantSet(t)
	set.Receipts = set.Receipts[:1]
	require.ErrorContains(t, set.Validate(), "exactly one receipt")

	set = validInvariantSet(t)
	set.Receipts = append(set.Receipts, AVMReceiptRecord{ReceiptID: "receipt-duplicate", MessageID: "msg-executed", ResultCode: async.ResultOK})
	require.ErrorContains(t, set.Validate(), "exactly one receipt")
}

func TestAVMStateInvariantsRejectExpiredAndBouncedDrift(t *testing.T) {
	set := validInvariantSet(t)
	for i := range set.Receipts {
		if set.Receipts[i].MessageID == "msg-expired" {
			set.Receipts[i].ResultCode = async.ResultExecutionFailed
		}
	}
	require.ErrorContains(t, set.Validate(), "receipt must be expired")

	set = validInvariantSet(t)
	for i := range set.StoredMessages {
		if set.StoredMessages[i].MessageID == "msg-bounced" {
			set.StoredMessages[i].OriginalMessageID = ""
		}
	}
	require.ErrorContains(t, set.Validate(), "original message id")
}

func TestAVMStateInvariantsRejectContinuationMailboxStorageAndZoneRootDrift(t *testing.T) {
	set := validInvariantSet(t)
	set.Continuations[0].ActorID = "missing-actor"
	require.ErrorContains(t, set.Validate(), "missing actor")

	set = validInvariantSet(t)
	set.MailboxEntries[0], set.MailboxEntries[1] = set.MailboxEntries[1], set.MailboxEntries[0]
	require.NoError(t, set.Validate())
	set.MailboxEntries = append(set.MailboxEntries, AVMActorMailboxEntry{ActorID: "actor-1", SortKey: "0001", MessageID: "msg-executed"})
	require.ErrorContains(t, set.Validate(), "duplicate mailbox sort key")

	set = validInvariantSet(t)
	set.ContractStorage[0].Key = AVMContractStorageKey("contract-2", "balance")
	require.ErrorContains(t, set.Validate(), "scoped by contract address")

	set = validInvariantSet(t)
	set.ZoneRoots[0].ContinuationRoot = engineHash("wrong")
	require.ErrorContains(t, set.Validate(), "hash mismatch")
}

func validInvariantSet(t *testing.T) AVMStateInvariantSet {
	t.Helper()
	zoneRoot, err := NewAVMZoneStateRoot(AVMZoneStateRoot{
		ZoneID:			zonestypes.ZoneIDContract,
		Height:			42,
		StateRoot:		engineHash("zone-state"),
		MessageRoot:		engineHash("zone-message"),
		ExecutionRoot:		engineHash("zone-execution"),
		ContinuationRoot:	engineHash("zone-continuation"),
	})
	require.NoError(t, err)
	return AVMStateInvariantSet{
		StoredMessages: []AVMStoredMessageRecord{
			{MessageID: "msg-bounced", ZoneID: zonestypes.ZoneIDContract, Bounced: true, OriginalMessageID: "msg-original"},
			{MessageID: "msg-executed", ZoneID: zonestypes.ZoneIDContract},
			{MessageID: "msg-expired", ZoneID: zonestypes.ZoneIDContract, Expired: true},
			{MessageID: "msg-original", ZoneID: zonestypes.ZoneIDContract},
		},
		QueuedMessages: []AVMQueuedMessageRef{
			{MessageID: "msg-executed", ZoneID: zonestypes.ZoneIDContract, QueueID: "default", SortKey: "0001"},
			{MessageID: "msg-expired", ZoneID: zonestypes.ZoneIDContract, QueueID: "default", SortKey: "0002"},
		},
		ExecutedMessages: []AVMExecutedMessageRef{
			{MessageID: "msg-executed"},
			{MessageID: "msg-expired"},
		},
		Receipts: []AVMReceiptRecord{
			{ReceiptID: "receipt-executed", MessageID: "msg-executed", ResultCode: async.ResultOK},
			{ReceiptID: "receipt-expired", MessageID: "msg-expired", ResultCode: async.ResultExpired},
		},
		Continuations: []AVMContinuationRef{
			{ContinuationID: "cont-1", ActorID: "actor-1"},
			{ContinuationID: "cont-2", WorkflowID: "workflow-1"},
		},
		Actors: []AVMActorRef{
			{ActorID: "actor-1"},
		},
		Workflows: []AVMWorkflowRef{
			{WorkflowID: "workflow-1"},
		},
		MailboxEntries: []AVMActorMailboxEntry{
			{ActorID: "actor-1", SortKey: "0001", MessageID: "msg-executed"},
			{ActorID: "actor-1", SortKey: "0002", MessageID: "msg-expired"},
		},
		ContractStorage: []AVMContractStorageRef{
			{ContractAddress: "contract-1", Key: AVMContractStorageKey("contract-1", "balance")},
		},
		ZoneRoots:	[]AVMZoneStateRoot{zoneRoot},
	}
}
