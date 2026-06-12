package types

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMempoolSeparationPlanBuildsZoneShardMessageClassLanes(t *testing.T) {
	params := DefaultMempoolSeparationParams()
	txs := []MempoolAdmissionTx{
		mempoolTx("contract-low-fee", "alice", ZoneIDContract, "1", "contract:abc", MempoolClassContract, 5, 90, 10, 2, true, false),
		mempoolTx("financial-high-fee", "bob", ZoneIDFinancial, "0", "account:bob", MempoolClassPayment, 50, 80, 10, 1, true, false),
		mempoolTx("contract-high-fee", "carol", ZoneIDContract, "1", "contract:abc", MempoolClassContract, 70, 70, 10, 2, true, false),
	}

	plan, err := BuildMempoolSeparationPlan(11, txs, params)
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.Len(t, plan.Lanes, 2)
	require.NotEmpty(t, plan.PlanHash)

	contractLane := findMempoolLane(t, plan, ZoneIDContract, "1", MempoolClassContract)
	require.Len(t, contractLane.Transactions, 2)
	require.Equal(t, hashParts("mempool-tx", "contract-high-fee"), contractLane.Transactions[0].TxHash)
	require.Equal(t, hashParts("mempool-tx", "contract-low-fee"), contractLane.Transactions[1].TxHash)
}

func TestMempoolSeparationPlanRootIsCanonicalAcrossInputOrder(t *testing.T) {
	params := DefaultMempoolSeparationParams()
	txs := []MempoolAdmissionTx{
		mempoolTx("a", "alice", ZoneIDIdentity, "0", "name:alice.aet", MempoolClassIdentity, 10, 40, 12, 1, true, false),
		mempoolTx("b", "bob", ZoneIDFinancial, "0", "account:bob", MempoolClassPayment, 11, 39, 12, 1, true, false),
		mempoolTx("c", "carol", ZoneIDContract, "2", "contract:carol", MempoolClassContract, 12, 38, 12, 1, true, false),
	}
	planA, err := BuildMempoolSeparationPlan(12, txs, params)
	require.NoError(t, err)

	slices.Reverse(txs)
	planB, err := BuildMempoolSeparationPlan(12, txs, params)
	require.NoError(t, err)
	require.Equal(t, planA.PlanHash, planB.PlanHash)
	require.Equal(t, planA.Lanes, planB.Lanes)
}

func TestMempoolSeparationPlanAppliesDoSLimitsAndUnknownTargetRules(t *testing.T) {
	params := DefaultMempoolSeparationParams()
	params.MaxPerSender = 1
	params.ParamsHash = ComputeMempoolSeparationParamsHash(params)
	_, err := BuildMempoolSeparationPlan(13, []MempoolAdmissionTx{
		mempoolTx("sender-1", "alice", ZoneIDFinancial, "0", "account:alice", MempoolClassPayment, 10, 50, 13, 1, true, false),
		mempoolTx("sender-2", "alice", ZoneIDFinancial, "0", "account:bob", MempoolClassPayment, 10, 50, 13, 1, true, false),
	}, params)
	require.ErrorContains(t, err, "sender DoS")

	params = DefaultMempoolSeparationParams()
	params.MaxPerTargetObject = 1
	params.ParamsHash = ComputeMempoolSeparationParamsHash(params)
	_, err = BuildMempoolSeparationPlan(13, []MempoolAdmissionTx{
		mempoolTx("target-1", "alice", ZoneIDContract, "0", "contract:shared", MempoolClassContract, 10, 50, 13, 1, true, false),
		mempoolTx("target-2", "bob", ZoneIDContract, "0", "contract:shared", MempoolClassContract, 10, 50, 13, 1, true, false),
	}, params)
	require.ErrorContains(t, err, "target object DoS")

	unknownRemoteShard := mempoolTx("unknown", "alice", ZoneIDIdentity, "2", "name:unknown.aet", MempoolClassIdentity, 10, 50, 13, 1, false, false)
	unknownRemoteShard.RouteKey = ""
	_, err = BuildMempoolSeparationPlan(13, []MempoolAdmissionTx{unknownRemoteShard}, DefaultMempoolSeparationParams())
	require.ErrorContains(t, err, "unknown target")

	systemRouted := unknownRemoteShard
	systemRouted.TargetShardID = "0"
	plan, err := BuildMempoolSeparationPlan(13, []MempoolAdmissionTx{systemRouted}, DefaultMempoolSeparationParams())
	require.NoError(t, err)
	require.Len(t, plan.Lanes, 1)
}

func TestLatencyStrategySpecCoversTargetsMetricsSLAFeesAndQueues(t *testing.T) {
	metrics := []LatencyMetric{
		latencyMetric(LatencySingleZoneLocal, 1, 10, 20),
		latencyMetric(LatencyCrossZoneAsync, 3, 5, 20),
	}
	queue := []DeliveryPriorityItem{
		deliveryPriorityItem(hashParts("msg", "slow"), LatencyCrossZoneAsync, ZoneIDContract, "1", 30, 10, 40, 20, false),
		deliveryPriorityItem(hashParts("msg", "near-expiry"), LatencySameZoneCrossShard, ZoneIDFinancial, "0", 5, 1, 21, 20, false),
		deliveryPriorityItem(hashParts("msg", "critical"), LatencyContractPromiseResolve, ZoneIDContract, "1", 1, 1, 35, 20, true),
	}

	spec, err := BuildLatencyStrategySpec(metrics, queue)
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.Targets, 4)
	require.NotEmpty(t, spec.MetricsRoot)
	require.NotEmpty(t, spec.DeliveryRoot)
	require.NotEmpty(t, spec.Root)

	orderedQueue := NormalizeDeliveryPriorityQueue(queue)
	require.True(t, orderedQueue[0].Critical)
	require.Equal(t, hashParts("msg", "near-expiry"), orderedQueue[1].MessageID)

	base, err := ComputeCongestionAwareForwardingFee(1_000, 1_000, 10, false, spec.FeePolicy)
	require.NoError(t, err)
	nearExpiry, err := ComputeCongestionAwareForwardingFee(1_000, 1_000, 1, false, spec.FeePolicy)
	require.NoError(t, err)
	critical, err := ComputeCongestionAwareForwardingFee(1_000, 1_000, 1, true, spec.FeePolicy)
	require.NoError(t, err)
	require.Greater(t, nearExpiry, base)
	require.Less(t, critical, nearExpiry)
}

func TestLatencyStrategySpecRejectsInvalidHashesAndBounds(t *testing.T) {
	badSLA := DefaultCrossZoneMessageSLAParams()
	badSLA.MaxDeliveryBlocks = 1
	badSLA.ParamsHash = ComputeCrossZoneMessageSLAParamsHash(badSLA)
	require.ErrorContains(t, badSLA.Validate(), "delivery bound")

	item := deliveryPriorityItem(hashParts("msg", "bad"), LatencyCrossZoneAsync, ZoneIDContract, "0", 1, 1, 30, 20, false)
	item.ExpiryHeight = 19
	item.ItemHash = ComputeDeliveryPriorityItemHash(item)
	require.ErrorContains(t, item.Validate(), "expiry precedes")

	metric := latencyMetric(LatencySingleZoneLocal, 1, 1, 20)
	metric.ObservedBlocks = 2
	require.ErrorContains(t, metric.Validate(), "hash mismatch")
}

func mempoolTx(seed string, sender string, zoneID ZoneID, shardID ShardID, object string, class MempoolMessageClass, fee uint64, expiry uint64, admission uint64, priority uint32, known bool, preResolved bool) MempoolAdmissionTx {
	return MempoolAdmissionTx{
		TxHash:			hashParts("mempool-tx", seed),
		Sender:			sender,
		TargetZoneID:		zoneID,
		TargetShardID:		shardID,
		RouteKey:		object + "/route",
		TargetObject:		object,
		MessageClass:		class,
		FeeNAET:		fee,
		ExpiryHeight:		expiry,
		AdmissionHeight:	admission,
		PriorityClass:		priority,
		TargetKnown:		known,
		PreResolved:		preResolved,
	}
}

func findMempoolLane(t *testing.T, plan MempoolSeparationPlan, zoneID ZoneID, shardID ShardID, class MempoolMessageClass) MempoolLane {
	t.Helper()
	for _, lane := range plan.Lanes {
		if lane.ZoneID == zoneID && lane.ShardID == shardID && lane.MessageClass == class {
			return lane
		}
	}
	t.Fatalf("missing lane %s/%s/%s", zoneID, shardID, class)
	return MempoolLane{}
}
