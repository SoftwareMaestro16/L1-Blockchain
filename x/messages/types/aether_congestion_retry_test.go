package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestCongestionAwareRoutingUsesCommittedMetricsAndFairness(t *testing.T) {
	table := testRoutingTable(t)
	msg := testAetherMessageDraft(testAetherRoute(t, 170, 1))
	params := testRoutingParams()
	params.FailureRateWeight = 3
	params.GasUtilizationWeight = 2
	params.ExpiryRateWeight = 4
	params.FairnessCredit = 10
	params.CriticalPriorityCredit = 20
	params.NormalPriorityFloor = 5
	metrics := []AetherRoutingMetric{
		congestionMetric(t, zonestypes.ZoneIDApplication, "app-1", 169, 2, 2, 1, 1, 1, 1, 0, false),
		congestionMetric(t, zonestypes.ZoneIDIdentity, "identity-1", 169, 50, 50, 20, 5, 10, 3, 0, false),
		congestionMetric(t, zonestypes.ZoneIDContract, "contract-7", 169, 1, 1, 1, 0, 1, 0, 2, true),
	}
	adjacency := []AetherRoutingEdge{
		{FromZoneID: zonestypes.ZoneIDFinancial, FromShardID: "financial-1", ToZoneID: zonestypes.ZoneIDIdentity, ToShardID: "identity-1"},
		{FromZoneID: zonestypes.ZoneIDIdentity, FromShardID: "identity-1", ToZoneID: zonestypes.ZoneIDContract, ToShardID: "contract-7"},
		{FromZoneID: zonestypes.ZoneIDFinancial, FromShardID: "financial-1", ToZoneID: zonestypes.ZoneIDApplication, ToShardID: "app-1"},
		{FromZoneID: zonestypes.ZoneIDApplication, FromShardID: "app-1", ToZoneID: zonestypes.ZoneIDContract, ToShardID: "contract-7"},
	}

	_, plan, err := CommitAetherMessageDeterministicRoute(msg, table, metrics, adjacency, params)
	require.NoError(t, err)
	require.Equal(t, zonestypes.ZoneIDApplication, plan.SelectedPath.Path[1].ZoneID)
	require.NoError(t, plan.Validate())

	reorderedMetrics := []AetherRoutingMetric{metrics[2], metrics[0], metrics[1]}
	reorderedAdjacency := []AetherRoutingEdge{adjacency[3], adjacency[2], adjacency[1], adjacency[0]}
	_, again, err := CommitAetherMessageDeterministicRoute(msg, table, reorderedMetrics, reorderedAdjacency, params)
	require.NoError(t, err)
	require.Equal(t, plan.RouteCommitment, again.RouteCommitment)
}

func TestRetryScheduleExpiryAndBounceAreDeterministic(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 180, 1))
	failed, err := AetherReceiptFromMessage(msg, ReceiptStatusFailed, 90, 3, sdkmath.NewInt(1), nil, "ERR_QUEUE_LIMIT", EmptyHash(), testMessageHash("retry-writes"))
	require.NoError(t, err)
	policy := testRetryPolicy()
	state := AetherRetryState{
		MsgID:			msg.MsgID,
		RetryCount:		1,
		ForwardingFeeEscrow:	sdkmath.NewInt(10),
		LastAttemptHeight:	90,
	}
	state.StateHash = ComputeAetherRetryStateHash(state)
	decision, err := DecideAetherRetry(msg, failed, state, policy, 90, AetherFailureTransientQueueLimit)
	require.NoError(t, err)
	require.Equal(t, AetherRetrySchedule, decision.Kind)
	require.Equal(t, uint32(2), decision.RetryCount)
	require.Equal(t, uint64(94), decision.NextEligibleHeight)
	require.Equal(t, sdkmath.NewInt(8), decision.RemainingFeeEscrow)
	require.NoError(t, decision.Validate())

	noRetry, err := DecideAetherRetry(msg, failed, state, policy, 90, AetherFailureInvalidPayload)
	require.NoError(t, err)
	require.Equal(t, AetherRetryNone, noRetry.Kind)

	expired, err := ExpireAetherMessage(msg, msg.ExpiryHeight+1, EmptyHash(), testMessageHash("expired-writes"))
	require.NoError(t, err)
	require.Equal(t, ReceiptStatusExpired, expired.Status)
	expiredDecision, err := DecideAetherRetry(msg, expired, state, policy, msg.ExpiryHeight+1, AetherFailureExpired)
	require.NoError(t, err)
	require.Equal(t, AetherRetryExpired, expiredDecision.Kind)

	bounce, err := BuildAetherBounce(msg, expired, sdkmath.NewInt(5), sdkmath.NewInt(1), 99, msg.ExpiryHeight+1, policy)
	require.NoError(t, err)
	require.Equal(t, msg.MsgID, bounce.BounceMsg.ParentMsgID)
	require.Equal(t, msg.TraceID, bounce.BounceMsg.TraceID)
	require.Equal(t, sdkmath.NewInt(5), bounce.RemainingValueNAET)
	require.False(t, bounce.BounceMsg.ValueNAET.GT(msg.ValueNAET))
	require.False(t, bounce.RemainingFee.GT(msg.ForwardingFee))
	require.NoError(t, bounce.Validate())
}

func TestRetryBoundsPreventUnboundedQueueGrowth(t *testing.T) {
	msg := testAetherMessage(t, testAetherRoute(t, 190, 1))
	failed, err := AetherReceiptFromMessage(msg, ReceiptStatusFailed, 91, 1, sdkmath.ZeroInt(), nil, "ERR_QUEUE_LIMIT", EmptyHash(), testMessageHash("bounded-writes"))
	require.NoError(t, err)
	policy := testRetryPolicy()
	state := AetherRetryState{
		MsgID:			msg.MsgID,
		RetryCount:		policy.MaxRetryCount,
		ForwardingFeeEscrow:	sdkmath.NewInt(100),
		LastAttemptHeight:	91,
	}
	state.StateHash = ComputeAetherRetryStateHash(state)
	decision, err := DecideAetherRetry(msg, failed, state, policy, 91, AetherFailureTransientQueueLimit)
	require.NoError(t, err)
	require.Equal(t, AetherRetryNone, decision.Kind)

	lowFee := state
	lowFee.RetryCount = 0
	lowFee.ForwardingFeeEscrow = sdkmath.ZeroInt()
	lowFee.StateHash = ComputeAetherRetryStateHash(lowFee)
	decision, err = DecideAetherRetry(msg, failed, lowFee, policy, 91, AetherFailureTransientQueueLimit)
	require.NoError(t, err)
	require.Equal(t, AetherRetryNone, decision.Kind)
}

func congestionMetric(t *testing.T, zoneID zonestypes.ZoneID, shardID string, height uint64, outbox uint64, inbox uint64, delay uint64, failure uint64, gas uint64, expiry uint64, fairness uint64, critical bool) AetherRoutingMetric {
	t.Helper()
	metric := AetherRoutingMetric{
		ZoneID:			zoneID,
		ShardID:		shardID,
		CommittedHeight:	height,
		OutboxBacklog:		outbox,
		InboxBacklog:		inbox,
		AverageExecutionDelay:	delay,
		FailedDeliveryRate:	failure,
		ShardGasUtilization:	gas,
		MessageExpiryRate:	expiry,
		Capacity:		100,
		FairnessCredit:		fairness,
		CriticalPriorityLane:	critical,
	}
	metric = normalizeAetherRoutingMetric(metric)
	metric.MetricHash = ComputeAetherRoutingMetricHash(metric)
	require.NoError(t, metric.Validate())
	return metric
}

func testRetryPolicy() AetherRetryPolicy {
	policy := AetherRetryPolicy{
		MaxRetryCount:		3,
		BaseDelayHeights:	2,
		MaxDelayHeights:	10,
		RetryFee:		sdkmath.NewInt(2),
		BounceGasLimit:		20,
		MaxBouncePayloadLen:	256,
	}
	policy.PolicyHash = hashParts("aether-retry-policy-test", "v1")
	return policy
}
