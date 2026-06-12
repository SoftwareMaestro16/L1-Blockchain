package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShardLocalMessageStoresSortAndBuildProofableEntries(t *testing.T) {
	msgA := shardMessage(t, "msg-a", ZoneIDApplication, "1", ZoneIDContract, "2", "contract/instance/a", 2, 20, 2, 3)
	msgB := shardMessage(t, "msg-b", ZoneIDApplication, "1", ZoneIDContract, "2", "contract/instance/b", 1, 19, 1, 3)
	store, err := NewShardMessageStore(ShardMessageStore{
		ZoneID:		ZoneIDApplication,
		ShardID:	"1",
		Height:		88,
		QueueKind:	ShardQueueOutbox,
		Messages:	[]ShardMessageEnvelope{msgA, msgB},
	})
	require.NoError(t, err)
	require.Equal(t, "msg-b", store.Messages[0].MsgID)
	require.Equal(t, ComputeShardMessageStoreRoot(store), store.StoreRoot)
	require.NoError(t, store.ValidateHash())

	entries, err := BuildShardMessageStoreEntries(store)
	require.NoError(t, err)
	require.Equal(t, "zones/APPLICATION_ZONE/shards/1/outbox/msg-b", entries[0].Key)
	require.Equal(t, msgB.MessageHash, entries[0].MessageHash)

	reordered, err := NewShardMessageStore(ShardMessageStore{
		ZoneID:		ZoneIDApplication,
		ShardID:	"1",
		Height:		88,
		QueueKind:	ShardQueueOutbox,
		Messages:	[]ShardMessageEnvelope{msgB, msgA},
	})
	require.NoError(t, err)
	require.Equal(t, store.StoreRoot, reordered.StoreRoot)
}

func TestShardRootsBindIntoZoneCommitment(t *testing.T) {
	rootA := shardRoot(t, ZoneIDFinancial, "0", 77)
	rootB := shardRoot(t, ZoneIDFinancial, "1", 77)
	shardRootsRoot, err := ComputeShardRootsRoot([]ShardRoot{rootB, rootA})
	require.NoError(t, err)

	commitment, err := BuildZoneCommitmentFromShardRoots(
		77,
		ZoneIDFinancial,
		[]ShardRoot{rootA, rootB},
		testHash("financial/state"),
		testHash("financial/inbox"),
		testHash("financial/outbox"),
		testHash("financial/receipts"),
		testHash("financial/events"),
		testHash("financial/params"),
		testHash("financial/summary"),
	)
	require.NoError(t, err)
	require.Equal(t, shardRootsRoot, commitment.ShardRootsRoot)
	require.NoError(t, commitment.ValidateHash())
}

func TestShardMigrationExecutorIsReplayable(t *testing.T) {
	taskA, err := NewShardMigrationTask(ShardMigrationTask{
		ZoneID:			ZoneIDContract,
		SourceShardID:		"1",
		DestinationShardID:	"3",
		SourceLayoutEpoch:	4,
		TargetLayoutEpoch:	5,
		KeyPrefix:		"contract/storage/a",
		DeliveryEpoch:		5,
	})
	require.NoError(t, err)
	taskB, err := NewShardMigrationTask(ShardMigrationTask{
		ZoneID:			ZoneIDContract,
		SourceShardID:		"2",
		DestinationShardID:	"3",
		SourceLayoutEpoch:	4,
		TargetLayoutEpoch:	5,
		KeyPrefix:		"contract/storage/b",
		DeliveryEpoch:		5,
	})
	require.NoError(t, err)

	before := testHash("contract/state/before-migration")
	receiptsAB, rootAB, err := ExecuteShardMigrationTasks([]ShardMigrationTask{taskA, taskB}, 90, before)
	require.NoError(t, err)
	receiptsBA, rootBA, err := ExecuteShardMigrationTasks([]ShardMigrationTask{taskB, taskA}, 90, before)
	require.NoError(t, err)
	require.Equal(t, rootAB, rootBA)
	require.Equal(t, receiptsAB[0].ReceiptHash, receiptsBA[0].ReceiptHash)
	require.NoError(t, receiptsAB[0].ValidateHash())
}

func TestRoutingStabilityAcrossLayoutEpochsAndInFlightMessages(t *testing.T) {
	layout1 := routingTestLayout(t, ZoneIDContract, 1, ShardAssignmentConsistentHash)
	layout1.PlacementOverrides = []ShardPlacementOverride{{ObjectKey: "contract/instance/alice", ShardID: "1"}}
	layout1.LayoutHash = ComputeShardLayoutHash(layout1)
	require.NoError(t, layout1.ValidateHash())

	layout2 := routingTestLayout(t, ZoneIDContract, 2, ShardAssignmentConsistentHash)
	layout2.PlacementOverrides = []ShardPlacementOverride{{ObjectKey: "contract/instance/alice", ShardID: "2"}}
	layout2.LayoutHash = ComputeShardLayoutHash(layout2)
	require.NoError(t, layout2.ValidateHash())

	input := ShardRoutingInput{ZoneID: ZoneIDContract, StateKey: "contract/instance/alice", ShardLayoutEpoch: 1}
	routeA, err := RouteKeyToShard(layout1, input)
	require.NoError(t, err)
	routeB, err := RouteKeyToShard(layout1, input)
	require.NoError(t, err)
	require.Equal(t, routeA.RouteHash, routeB.RouteHash)
	require.Equal(t, ShardID("1"), routeA.ShardID)

	epoch2Route, err := RouteKeyToShard(layout2, ShardRoutingInput{ZoneID: ZoneIDContract, StateKey: "contract/instance/alice", ShardLayoutEpoch: 2})
	require.NoError(t, err)
	require.Equal(t, ShardID("2"), epoch2Route.ShardID)

	inFlightBeforeSplit := shardMessage(t, "msg-before-split", ZoneIDApplication, "1", ZoneIDContract, "", "contract/instance/alice", 1, 30, 1, 1)
	inFlightAfterSplit := shardMessage(t, "msg-after-split", ZoneIDApplication, "1", ZoneIDContract, "", "contract/instance/alice", 1, 31, 2, 2)
	beforeRoute, err := ResolveShardMessageDeliveryRoute([]ShardLayout{layout2, layout1}, inFlightBeforeSplit)
	require.NoError(t, err)
	afterRoute, err := ResolveShardMessageDeliveryRoute([]ShardLayout{layout1, layout2}, inFlightAfterSplit)
	require.NoError(t, err)
	require.Equal(t, ShardID("1"), beforeRoute.ShardID)
	require.Equal(t, ShardID("2"), afterRoute.ShardID)
	require.Equal(t, ZoneID(ZoneIDApplication), inFlightBeforeSplit.SenderZoneID)
	require.Equal(t, ShardID("1"), inFlightBeforeSplit.SenderShardID)
}

func shardMessage(
	t *testing.T,
	msgID string,
	senderZone ZoneID,
	senderShard ShardID,
	receiverZone ZoneID,
	receiverShard ShardID,
	stateKey string,
	priority uint32,
	admissionHeight uint64,
	sourceEpoch uint64,
	deliveryEpoch uint64,
) ShardMessageEnvelope {
	t.Helper()
	msg, err := NewShardMessageEnvelope(ShardMessageEnvelope{
		MsgID:			msgID,
		TraceID:		"trace-" + msgID,
		Sender:			"acct/source-" + msgID,
		Receiver:		"acct/receiver-" + msgID,
		SenderZoneID:		senderZone,
		SenderShardID:		senderShard,
		ReceiverZoneID:		receiverZone,
		ReceiverShardID:	receiverShard,
		DestinationStateKey:	stateKey,
		PayloadType:		"test.payload",
		PayloadHash:		testHash(msgID + "/payload"),
		Priority:		priority,
		AdmissionHeight:	admissionHeight,
		CreatedAtHeight:	admissionHeight,
		ExpiryHeight:		admissionHeight + 100,
		MessageIndex:		priority,
		SourceLayoutEpoch:	sourceEpoch,
		DeliveryLayoutEpoch:	deliveryEpoch,
	})
	require.NoError(t, err)
	return msg
}
