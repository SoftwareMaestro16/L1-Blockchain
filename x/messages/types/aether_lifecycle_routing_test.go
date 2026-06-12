package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	aetracoretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAetherMessageLifecycleRequiresCanonicalStages(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 140, 2))
	receipt, err := AetherReceiptFromMessage(msg, ReceiptStatusExecuted, 146, 10, sdkmath.OneInt(), nil, "", EmptyHash(), testMessageHash("life-writes"))
	require.NoError(t, err)
	records := []AetherMessageLifecycleRecord{
		lifecycleRecord(t, msg, MessageLifecycleBounceOrFinalize, 147, receipt.ReceiptHash),
		lifecycleRecord(t, msg, MessageLifecycleCreated, 140, ""),
		lifecycleRecord(t, msg, MessageLifecycleQueuedInSourceOutbox, 141, ""),
		lifecycleRecord(t, msg, MessageLifecycleCommittedInMessageRoot, 142, ""),
		lifecycleRecord(t, msg, MessageLifecycleEligibleForDelivery, 143, ""),
		lifecycleRecord(t, msg, MessageLifecycleQueuedInDestinationInbox, 144, ""),
		lifecycleRecord(t, msg, MessageLifecycleExecutedOrFailed, 145, ""),
		lifecycleRecord(t, msg, MessageLifecycleReceipt, 146, receipt.ReceiptHash),
	}
	require.NoError(t, ValidateAetherMessageLifecycle(records))
	rootA, err := ComputeAetherLifecycleRoot(records)
	require.NoError(t, err)
	rootB, err := ComputeAetherLifecycleRoot([]AetherMessageLifecycleRecord{records[1], records[2], records[3], records[4], records[5], records[6], records[7], records[0]})
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)

	missing := records[:len(records)-1]
	require.ErrorContains(t, ValidateAetherMessageLifecycle(missing), "all canonical stages")
}

func TestDeterministicAetherRoutingSelectsLowestCommittedCost(t *testing.T) {
	table := testRoutingTable(t)
	msg := testAetherMessageDraft(testAetherRoute(t, 150, 1))
	metrics := []AetherRoutingMetric{
		routingMetric(t, zonestypes.ZoneIDApplication, "app-1", 150, 100, 1, 1),
		routingMetric(t, zonestypes.ZoneIDIdentity, "identity-1", 150, 100, 50, 50),
		routingMetric(t, zonestypes.ZoneIDContract, "contract-7", 150, 100, 1, 1),
	}
	adjacency := []AetherRoutingEdge{
		{FromZoneID: zonestypes.ZoneIDFinancial, FromShardID: "financial-1", ToZoneID: zonestypes.ZoneIDIdentity, ToShardID: "identity-1"},
		{FromZoneID: zonestypes.ZoneIDIdentity, FromShardID: "identity-1", ToZoneID: zonestypes.ZoneIDContract, ToShardID: "contract-7"},
		{FromZoneID: zonestypes.ZoneIDFinancial, FromShardID: "financial-1", ToZoneID: zonestypes.ZoneIDApplication, ToShardID: "app-1"},
		{FromZoneID: zonestypes.ZoneIDApplication, FromShardID: "app-1", ToZoneID: zonestypes.ZoneIDContract, ToShardID: "contract-7"},
	}
	committed, plan, err := CommitAetherMessageDeterministicRoute(msg, table, metrics, adjacency, testRoutingParams())
	require.NoError(t, err)
	require.Equal(t, plan.RouteCommitment, committed.RouteCommitment)
	require.Equal(t, zonestypes.ZoneIDApplication, plan.SelectedPath.Path[1].ZoneID)
	require.Equal(t, uint32(2), plan.SelectedPath.HopCount)
	require.NoError(t, committed.Validate())
	require.NoError(t, plan.Validate())

	again, againPlan, err := CommitAetherMessageDeterministicRoute(msg, table, []AetherRoutingMetric{metrics[2], metrics[0], metrics[1]}, []AetherRoutingEdge{adjacency[3], adjacency[1], adjacency[0], adjacency[2]}, testRoutingParams())
	require.NoError(t, err)
	require.Equal(t, committed.MsgID, again.MsgID)
	require.Equal(t, plan.RouteCommitment, againPlan.RouteCommitment)
}

func TestAetherRoutingRejectsHopLimitAndMissingCommittedZone(t *testing.T) {
	table := testRoutingTable(t)
	msg := testAetherMessageDraft(testAetherRoute(t, 160, 1))
	params := testRoutingParams()
	params.MaxHopCount = 1
	_, err := SelectDeterministicAetherRoute(msg, table, nil, []AetherRoutingEdge{
		{FromZoneID: zonestypes.ZoneIDFinancial, FromShardID: "financial-1", ToZoneID: zonestypes.ZoneIDApplication, ToShardID: "app-1"},
		{FromZoneID: zonestypes.ZoneIDApplication, FromShardID: "app-1", ToZoneID: zonestypes.ZoneIDContract, ToShardID: "contract-7"},
	}, params)
	require.ErrorContains(t, err, "candidate")

	badTable := table
	badTable.Entries = badTable.Entries[:1]
	badTable.TableHash = aetracoretypes.ComputeRoutingTableHash(badTable)
	_, err = SelectDeterministicAetherRoute(msg, badTable, nil, nil, testRoutingParams())
	require.ErrorContains(t, err, "routing table")
}

func lifecycleRecord(t *testing.T, msg AetherMessage, stage MessageLifecycleStage, height uint64, receiptHash string) AetherMessageLifecycleRecord {
	t.Helper()
	record, err := NewAetherMessageLifecycleRecord(AetherMessageLifecycleRecord{
		MsgID:			msg.MsgID,
		Stage:			stage,
		Height:			height,
		RouteCommitment:	msg.RouteCommitment,
		ReceiptHash:		receiptHash,
	})
	require.NoError(t, err)
	return record
}

func testAetherMessageDraft(route UnifiedMessageRoute) AetherMessage {
	msg := testAetherMessageNoT(nil, route)
	msg.MsgID = ""
	msg.RouteCommitment = EmptyHash()
	return msg
}

func testAetherMessageNoT(t *testing.T, route UnifiedMessageRoute) AetherMessage {
	msg := AetherMessage{
		Sender:			"account/alice",
		SenderZoneID:		zonestypes.ZoneIDFinancial,
		SenderShardID:		"financial-1",
		Receiver:		"contract/vault",
		ReceiverZoneID:		zonestypes.ZoneIDContract,
		ReceiverShardID:	"contract-7",
		ValueNAET:		sdkmath.NewInt(10),
		Payload:		[]byte("execute"),
		PayloadType:		"contract.execute",
		GasLimit:		100,
		GasPrice:		sdkmath.NewInt(2),
		ForwardingFee:		sdkmath.NewInt(3),
		ExpiryHeight:		180,
		Bounce:			true,
		ExecutionMode:		ExecutionModeAsync,
		OrderingClass:		OrderingClassSenderOrdered,
		RouteCommitment:	route.RouteCommitment,
		CreatedAtHeight:	150,
		Nonce:			9,
	}
	if t == nil {
		msg = normalizeAetherMessage(msg)
		msg.TraceID = ComputeAetherTraceID(msg)
		return msg
	}
	out, err := NewAetherMessage(msg)
	require.NoError(t, err)
	return out
}

func testRoutingTable(t *testing.T) aetracoretypes.RoutingTableCommitment {
	t.Helper()
	table, err := aetracoretypes.NewRoutingTableCommitment(7, 150, []aetracoretypes.RoutingZoneEntry{
		{ZoneID: aetracoretypes.ZoneID("FINANCIAL_ZONE"), LayoutEpoch: 1, ActiveShards: 2, LayoutHash: testCoreHash("financial-layout")},
		{ZoneID: aetracoretypes.ZoneID("APPLICATION_ZONE"), LayoutEpoch: 1, ActiveShards: 2, LayoutHash: testCoreHash("application-layout")},
		{ZoneID: aetracoretypes.ZoneID("IDENTITY_ZONE"), LayoutEpoch: 1, ActiveShards: 2, LayoutHash: testCoreHash("identity-layout")},
		{ZoneID: aetracoretypes.ZoneID("CONTRACT_ZONE"), LayoutEpoch: 1, ActiveShards: 2, LayoutHash: testCoreHash("contract-layout")},
	})
	require.NoError(t, err)
	return table
}

func routingMetric(t *testing.T, zoneID zonestypes.ZoneID, shardID string, height uint64, capacity uint64, congestion uint64, backlog uint64) AetherRoutingMetric {
	t.Helper()
	metric := AetherRoutingMetric{ZoneID: zoneID, ShardID: shardID, CommittedHeight: height, Capacity: capacity, CongestionScore: congestion, QueueBacklog: backlog}
	metric.MetricHash = ComputeAetherRoutingMetricHash(metric)
	require.NoError(t, metric.Validate())
	return metric
}

func testRoutingParams() AetherRoutingParams {
	return AetherRoutingParams{
		MaxHopCount:		3,
		BaseHopCost:		10,
		CongestionWeight:	2,
		QueueWeight:		1,
		LatencyWeight:		1,
		CapacityPenalty:	100,
		RequiredCapacity:	10,
		GovernanceHash:		EmptyHash(),
	}
}

func testCoreHash(seed string) string {
	return hashParts("aether-core-test-hash", seed)
}
