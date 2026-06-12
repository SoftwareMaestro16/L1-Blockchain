package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatencyMetricsByOperationClass(t *testing.T) {
	params := DefaultCrossZoneMessageSLAParams()

	local, err := BuildLatencyMetric(LatencyMetricInput{
		OperationClass:	LatencyOpSingleZoneLocalTx,
		ZoneID:		"financial",
		ShardID:	"shard-a",
		CreatedHeight:	10,
		ExecutedHeight:	11,
	}, params)
	require.NoError(t, err)
	require.True(t, local.Satisfied)
	require.Equal(t, uint64(1), local.TargetBlocks)

	crossShard, err := BuildLatencyMetric(LatencyMetricInput{
		OperationClass:		LatencyOpSameZoneCrossShard,
		ZoneID:			"financial",
		ShardID:		"shard-a",
		DestinationZoneID:	"financial",
		DestinationShardID:	"shard-b",
		CreatedHeight:		10,
		ExecutedHeight:		12,
	}, params)
	require.NoError(t, err)
	require.False(t, crossShard.Satisfied)
	require.Equal(t, uint64(2), crossShard.ObservedBlocks)

	crossZone, err := BuildLatencyMetric(LatencyMetricInput{
		OperationClass:		LatencyOpCrossZoneAsyncMessage,
		ZoneID:			"identity",
		ShardID:		"shard-a",
		DestinationZoneID:	"contract",
		DestinationShardID:	"shard-c",
		CreatedHeight:		20,
		SourceCommitmentHeight:	21,
		ExecutedHeight:		23,
	}, params)
	require.NoError(t, err)
	require.True(t, crossZone.Satisfied)
	require.Equal(t, uint64(2), crossZone.ObservedBlocks)

	promise, err := BuildLatencyMetric(LatencyMetricInput{
		OperationClass:	LatencyOpContractPromiseResolution,
		ZoneID:		"contract",
		ShardID:	"shard-c",
		CreatedHeight:	30,
		EligibleHeight:	32,
		ExecutedHeight:	33,
	}, params)
	require.NoError(t, err)
	require.True(t, promise.Satisfied)
	require.Equal(t, uint64(1), promise.ObservedBlocks)
}

func TestLatencyRejectsCrossZoneBeforeSourceCommitmentAndNonFuturePromise(t *testing.T) {
	params := DefaultCrossZoneMessageSLAParams()

	_, err := BuildLatencyMetric(LatencyMetricInput{
		OperationClass:		LatencyOpCrossZoneAsyncMessage,
		ZoneID:			"identity",
		ShardID:		"shard-a",
		DestinationZoneID:	"contract",
		DestinationShardID:	"shard-c",
		CreatedHeight:		20,
		SourceCommitmentHeight:	22,
		ExecutedHeight:		21,
	}, params)
	require.ErrorContains(t, err, "before source commitment")

	_, err = BuildLatencyMetric(LatencyMetricInput{
		OperationClass:	LatencyOpContractPromiseResolution,
		ZoneID:		"contract",
		ShardID:	"shard-c",
		CreatedHeight:	30,
		EligibleHeight:	30,
		ExecutedHeight:	31,
	}, params)
	require.ErrorContains(t, err, "future message")
}

func TestCongestionAwareForwardingFeesIncreaseWithQueuePressure(t *testing.T) {
	params := DefaultCrossZoneMessageSLAParams()
	base, err := ComputeCongestionAwareForwardingFee("1", 0, 0, params)
	require.NoError(t, err)

	congested, err := ComputeCongestionAwareForwardingFee("1", 8_000, params.QueueDepthFeeStep*2, params)
	require.NoError(t, err)
	require.NotEqual(t, base, congested)

	baseInt, err := parsePerformanceNonNegativeInt("base", base)
	require.NoError(t, err)
	congestedInt, err := parsePerformanceNonNegativeInt("congested", congested)
	require.NoError(t, err)
	require.True(t, congestedInt.GT(baseInt))
}

func TestLatencyDeliveryQueuePrioritizesNearExpiryAndFee(t *testing.T) {
	params := DefaultCrossZoneMessageSLAParams()
	height := uint64(50)
	nearExpiry := latencyDeliveryMessage("near-expiry", LatencyOpCrossZoneAsyncMessage, "identity", "contract", 10, 12, 50, 51, "1", 100, 1)
	highFee := latencyDeliveryMessage("high-fee", LatencyOpCrossZoneAsyncMessage, "identity", "contract", 10, 12, 50, 80, "100", 0, 0)
	normal := latencyDeliveryMessage("normal", LatencyOpCrossZoneAsyncMessage, "identity", "contract", 10, 12, 50, 90, "2", 0, 0)

	queue, err := BuildLatencyDeliveryQueue(height, []LatencyDeliveryMessage{normal, highFee, nearExpiry}, params)
	require.NoError(t, err)
	require.NoError(t, queue.Validate(params.Normalize()))
	require.Len(t, queue.Messages, 3)
	require.Equal(t, nearExpiry.MessageID, queue.Messages[0].MessageID)
	require.Equal(t, highFee.MessageID, queue.Messages[1].MessageID)
	require.NotEmpty(t, queue.QueueRoot)
}

func TestLatencyDeliveryQueueValidatesPromiseFutureMessages(t *testing.T) {
	params := DefaultCrossZoneMessageSLAParams()
	msg := latencyDeliveryMessage("promise", LatencyOpContractPromiseResolution, "contract", "contract", 20, 0, 21, 80, "1", 0, 0)
	msg.PromiseResolution = true

	queue, err := BuildLatencyDeliveryQueue(21, []LatencyDeliveryMessage{msg}, params)
	require.NoError(t, err)
	require.Len(t, queue.Messages, 1)

	bad := msg
	bad.EligibleHeight = bad.CreatedHeight
	_, err = BuildLatencyDeliveryQueue(21, []LatencyDeliveryMessage{bad}, params)
	require.ErrorContains(t, err, "future eligible")
}

func latencyDeliveryMessage(seed string, class LatencyOperationClass, sourceZone, destinationZone string, created, committed, eligible, expiry uint64, fee string, congestion uint32, queueDepth uint32) LatencyDeliveryMessage {
	msg := LatencyDeliveryMessage{
		MessageID:		hashStrings("latency-message", seed),
		OperationClass:		class,
		SourceZoneID:		sourceZone,
		SourceShardID:		"shard-a",
		DestinationZoneID:	destinationZone,
		DestinationShardID:	"shard-b",
		CreatedHeight:		created,
		SourceCommitmentHeight:	committed,
		EligibleHeight:		eligible,
		ExpiryHeight:		expiry,
		BaseForwardingFee:	fee,
		CongestionBps:		congestion,
		QueueDepth:		queueDepth,
	}
	return msg
}
