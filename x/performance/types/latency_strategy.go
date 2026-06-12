package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultLatencyTargetLocalBlocks		= uint64(1)
	DefaultLatencyTargetCrossShardBlocks	= uint64(1)
	DefaultLatencyTargetCrossZoneBlocks	= uint64(2)
	DefaultLatencyTargetPromiseBlocks	= uint64(1)
	DefaultNearExpiryPriorityWindowBlocks	= uint64(2)
	DefaultLatencyQueueDepthStep		= uint32(64)
)

type LatencyOperationClass string

const (
	LatencyOpSingleZoneLocalTx		LatencyOperationClass	= "single_zone_local_tx"
	LatencyOpSameZoneCrossShard		LatencyOperationClass	= "same_zone_cross_shard_message"
	LatencyOpCrossZoneAsyncMessage		LatencyOperationClass	= "cross_zone_async_message"
	LatencyOpContractPromiseResolution	LatencyOperationClass	= "contract_promise_resolution"
)

type CrossZoneMessageSLAParams struct {
	LocalTxTargetBlocks		uint64
	CrossShardTargetBlocks		uint64
	CrossZoneTargetBlocks		uint64
	PromiseTargetBlocks		uint64
	MinForwardingFee		string
	CongestionMultiplierBps		uint32
	QueueDepthFeeStep		uint32
	QueueDepthFeeBps		uint32
	NearExpiryPriorityWindow	uint64
	NearExpiryPriorityBoost		uint64
	FeePriorityWeight		uint64
	CongestionPriorityWeight	uint64
	ExpiryPriorityWeight		uint64
	MaxDeliveryQueueDepth		uint32
	MaxCrossZoneCommitmentLag	uint64
}

type LatencyMetricInput struct {
	OperationClass		LatencyOperationClass
	ZoneID			string
	ShardID			string
	DestinationZoneID	string
	DestinationShardID	string
	CreatedHeight		uint64
	SourceCommitmentHeight	uint64
	EligibleHeight		uint64
	ExecutedHeight		uint64
}

type LatencyMetric struct {
	OperationClass		LatencyOperationClass
	ZoneID			string
	ShardID			string
	DestinationZoneID	string
	DestinationShardID	string
	CreatedHeight		uint64
	SourceCommitmentHeight	uint64
	EligibleHeight		uint64
	ExecutedHeight		uint64
	TargetBlocks		uint64
	ObservedBlocks		uint64
	Satisfied		bool
	MetricHash		string
}

type LatencyDeliveryMessage struct {
	MessageID		string
	OperationClass		LatencyOperationClass
	SourceZoneID		string
	SourceShardID		string
	DestinationZoneID	string
	DestinationShardID	string
	CreatedHeight		uint64
	SourceCommitmentHeight	uint64
	EligibleHeight		uint64
	ExpiryHeight		uint64
	BaseForwardingFee	string
	ForwardingFee		string
	CongestionBps		uint32
	QueueDepth		uint32
	PromiseResolution	bool
	PriorityScore		uint64
	MessageHash		string
}

type LatencyDeliveryQueue struct {
	Height		uint64
	Messages	[]LatencyDeliveryMessage
	QueueRoot	string
}

func DefaultCrossZoneMessageSLAParams() CrossZoneMessageSLAParams {
	return CrossZoneMessageSLAParams{
		LocalTxTargetBlocks:		DefaultLatencyTargetLocalBlocks,
		CrossShardTargetBlocks:		DefaultLatencyTargetCrossShardBlocks,
		CrossZoneTargetBlocks:		DefaultLatencyTargetCrossZoneBlocks,
		PromiseTargetBlocks:		DefaultLatencyTargetPromiseBlocks,
		MinForwardingFee:		"1",
		CongestionMultiplierBps:	10_000,
		QueueDepthFeeStep:		DefaultLatencyQueueDepthStep,
		QueueDepthFeeBps:		250,
		NearExpiryPriorityWindow:	DefaultNearExpiryPriorityWindowBlocks,
		NearExpiryPriorityBoost:	10_000_000,
		FeePriorityWeight:		100,
		CongestionPriorityWeight:	10,
		ExpiryPriorityWeight:		1_000,
		MaxDeliveryQueueDepth:		50_000,
		MaxCrossZoneCommitmentLag:	8,
	}
}

func (p CrossZoneMessageSLAParams) Normalize() CrossZoneMessageSLAParams {
	defaults := DefaultCrossZoneMessageSLAParams()
	if p.LocalTxTargetBlocks == 0 {
		p.LocalTxTargetBlocks = defaults.LocalTxTargetBlocks
	}
	if p.CrossShardTargetBlocks == 0 {
		p.CrossShardTargetBlocks = defaults.CrossShardTargetBlocks
	}
	if p.CrossZoneTargetBlocks == 0 {
		p.CrossZoneTargetBlocks = defaults.CrossZoneTargetBlocks
	}
	if p.PromiseTargetBlocks == 0 {
		p.PromiseTargetBlocks = defaults.PromiseTargetBlocks
	}
	if strings.TrimSpace(p.MinForwardingFee) == "" {
		p.MinForwardingFee = defaults.MinForwardingFee
	}
	if p.CongestionMultiplierBps == 0 {
		p.CongestionMultiplierBps = defaults.CongestionMultiplierBps
	}
	if p.QueueDepthFeeStep == 0 {
		p.QueueDepthFeeStep = defaults.QueueDepthFeeStep
	}
	if p.QueueDepthFeeBps == 0 {
		p.QueueDepthFeeBps = defaults.QueueDepthFeeBps
	}
	if p.NearExpiryPriorityWindow == 0 {
		p.NearExpiryPriorityWindow = defaults.NearExpiryPriorityWindow
	}
	if p.NearExpiryPriorityBoost == 0 {
		p.NearExpiryPriorityBoost = defaults.NearExpiryPriorityBoost
	}
	if p.FeePriorityWeight == 0 {
		p.FeePriorityWeight = defaults.FeePriorityWeight
	}
	if p.CongestionPriorityWeight == 0 {
		p.CongestionPriorityWeight = defaults.CongestionPriorityWeight
	}
	if p.ExpiryPriorityWeight == 0 {
		p.ExpiryPriorityWeight = defaults.ExpiryPriorityWeight
	}
	if p.MaxDeliveryQueueDepth == 0 {
		p.MaxDeliveryQueueDepth = defaults.MaxDeliveryQueueDepth
	}
	if p.MaxCrossZoneCommitmentLag == 0 {
		p.MaxCrossZoneCommitmentLag = defaults.MaxCrossZoneCommitmentLag
	}
	p.MinForwardingFee = strings.TrimSpace(p.MinForwardingFee)
	return p
}

func (p CrossZoneMessageSLAParams) Validate() error {
	params := p.Normalize()
	for _, target := range []struct {
		name	string
		value	uint64
	}{
		{"local tx target", params.LocalTxTargetBlocks},
		{"cross-shard target", params.CrossShardTargetBlocks},
		{"cross-zone target", params.CrossZoneTargetBlocks},
		{"promise target", params.PromiseTargetBlocks},
	} {
		if target.value == 0 || target.value > 1_000 {
			return fmt.Errorf("latency SLA %s blocks is out of bounds", target.name)
		}
	}
	if _, err := parsePerformanceNonNegativeInt("latency SLA minimum forwarding fee", params.MinForwardingFee); err != nil {
		return err
	}
	if params.CongestionMultiplierBps < 10_000 || params.CongestionMultiplierBps > 100_000 {
		return errors.New("latency SLA congestion multiplier bps must be between 10000 and 100000")
	}
	if params.QueueDepthFeeBps > 100_000 {
		return errors.New("latency SLA queue depth fee bps is out of bounds")
	}
	if params.MaxDeliveryQueueDepth == 0 {
		return errors.New("latency SLA max delivery queue depth must be positive")
	}
	return nil
}

func BuildLatencyMetric(input LatencyMetricInput, params CrossZoneMessageSLAParams) (LatencyMetric, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return LatencyMetric{}, err
	}
	input = input.Normalize()
	if err := input.Validate(params); err != nil {
		return LatencyMetric{}, err
	}
	target := latencyTargetForClass(input.OperationClass, params)
	start := input.CreatedHeight
	if input.OperationClass == LatencyOpCrossZoneAsyncMessage {
		start = input.SourceCommitmentHeight
	}
	if input.OperationClass == LatencyOpContractPromiseResolution {
		start = input.EligibleHeight
	}
	observed := uint64(0)
	if input.ExecutedHeight >= start {
		observed = input.ExecutedHeight - start
	}
	metric := LatencyMetric{
		OperationClass:		input.OperationClass,
		ZoneID:			input.ZoneID,
		ShardID:		input.ShardID,
		DestinationZoneID:	input.DestinationZoneID,
		DestinationShardID:	input.DestinationShardID,
		CreatedHeight:		input.CreatedHeight,
		SourceCommitmentHeight:	input.SourceCommitmentHeight,
		EligibleHeight:		input.EligibleHeight,
		ExecutedHeight:		input.ExecutedHeight,
		TargetBlocks:		target,
		ObservedBlocks:		observed,
		Satisfied:		observed <= target,
	}
	metric.MetricHash = ComputeLatencyMetricHash(metric)
	return metric, metric.Validate()
}

func (i LatencyMetricInput) Normalize() LatencyMetricInput {
	i.ZoneID = strings.TrimSpace(i.ZoneID)
	i.ShardID = strings.TrimSpace(i.ShardID)
	i.DestinationZoneID = strings.TrimSpace(i.DestinationZoneID)
	i.DestinationShardID = strings.TrimSpace(i.DestinationShardID)
	i.OperationClass = LatencyOperationClass(strings.TrimSpace(string(i.OperationClass)))
	return i
}

func (i LatencyMetricInput) Validate(params CrossZoneMessageSLAParams) error {
	input := i.Normalize()
	if !IsLatencyOperationClass(input.OperationClass) {
		return errors.New("latency metric operation class is unsupported")
	}
	if err := validateExecutionToken("latency metric zone id", input.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("latency metric shard id", input.ShardID); err != nil {
		return err
	}
	if input.CreatedHeight == 0 || input.ExecutedHeight == 0 {
		return errors.New("latency metric created and executed heights must be positive")
	}
	if input.ExecutedHeight < input.CreatedHeight {
		return errors.New("latency metric executed height precedes created height")
	}
	switch input.OperationClass {
	case LatencyOpSingleZoneLocalTx:
		if input.ExecutedHeight > input.CreatedHeight+params.LocalTxTargetBlocks {
			return nil
		}
	case LatencyOpSameZoneCrossShard:
		if input.DestinationZoneID != "" && input.DestinationZoneID != input.ZoneID {
			return errors.New("latency same-zone cross-shard destination zone mismatch")
		}
		if input.DestinationShardID == "" || input.DestinationShardID == input.ShardID {
			return errors.New("latency same-zone cross-shard requires different destination shard")
		}
	case LatencyOpCrossZoneAsyncMessage:
		if err := validateExecutionToken("latency cross-zone destination zone id", input.DestinationZoneID); err != nil {
			return err
		}
		if err := validateExecutionToken("latency cross-zone destination shard id", input.DestinationShardID); err != nil {
			return err
		}
		if input.DestinationZoneID == input.ZoneID {
			return errors.New("latency cross-zone message requires different destination zone")
		}
		if input.SourceCommitmentHeight == 0 || input.SourceCommitmentHeight < input.CreatedHeight {
			return errors.New("latency cross-zone message requires source commitment")
		}
		if input.SourceCommitmentHeight-input.CreatedHeight > params.MaxCrossZoneCommitmentLag {
			return errors.New("latency cross-zone source commitment lag exceeds SLA")
		}
		if input.ExecutedHeight < input.SourceCommitmentHeight {
			return errors.New("latency cross-zone message executed before source commitment")
		}
	case LatencyOpContractPromiseResolution:
		if input.EligibleHeight <= input.CreatedHeight {
			return errors.New("latency contract promise resolution must execute as future message")
		}
		if input.ExecutedHeight < input.EligibleHeight {
			return errors.New("latency contract promise resolution executed before eligibility")
		}
	}
	return nil
}

func (m LatencyMetric) Validate() error {
	metric := m.Normalize()
	if !IsLatencyOperationClass(metric.OperationClass) {
		return errors.New("latency metric operation class is unsupported")
	}
	if err := validateExecutionToken("latency metric zone id", metric.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("latency metric shard id", metric.ShardID); err != nil {
		return err
	}
	if metric.TargetBlocks == 0 || metric.CreatedHeight == 0 || metric.ExecutedHeight == 0 {
		return errors.New("latency metric target and heights must be positive")
	}
	if metric.MetricHash != ComputeLatencyMetricHash(metric) {
		return errors.New("latency metric hash mismatch")
	}
	return nil
}

func (m LatencyMetric) Normalize() LatencyMetric {
	m.ZoneID = strings.TrimSpace(m.ZoneID)
	m.ShardID = strings.TrimSpace(m.ShardID)
	m.DestinationZoneID = strings.TrimSpace(m.DestinationZoneID)
	m.DestinationShardID = strings.TrimSpace(m.DestinationShardID)
	m.MetricHash = normalizeLowerHex(m.MetricHash)
	return m
}

func ComputeCongestionAwareForwardingFee(baseFee string, congestionBps uint32, queueDepth uint32, params CrossZoneMessageSLAParams) (string, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return "", err
	}
	if congestionBps > 10_000 {
		return "", errors.New("latency forwarding fee congestion bps must be <= 10000")
	}
	base, err := parsePerformanceNonNegativeInt("latency forwarding base fee", baseFee)
	if err != nil {
		return "", err
	}
	minFee, err := parsePerformanceNonNegativeInt("latency forwarding minimum fee", params.MinForwardingFee)
	if err != nil {
		return "", err
	}
	if base.LT(minFee) {
		base = minFee
	}
	congestionBump := base.MulRaw(int64(congestionBps)).MulRaw(int64(params.CongestionMultiplierBps)).QuoRaw(100_000_000)
	queueSteps := queueDepth / params.QueueDepthFeeStep
	queueBump := base.MulRaw(int64(queueSteps)).MulRaw(int64(params.QueueDepthFeeBps)).QuoRaw(10_000)
	total := base.Add(congestionBump).Add(queueBump)
	if (congestionBps > 0 || queueSteps > 0) && total.Equal(base) {
		total = total.AddRaw(1)
	}
	return total.String(), nil
}

func BuildLatencyDeliveryQueue(height uint64, messages []LatencyDeliveryMessage, params CrossZoneMessageSLAParams) (LatencyDeliveryQueue, error) {
	if height == 0 {
		return LatencyDeliveryQueue{}, errors.New("latency delivery queue height must be positive")
	}
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return LatencyDeliveryQueue{}, err
	}
	if uint32(len(messages)) > params.MaxDeliveryQueueDepth {
		return LatencyDeliveryQueue{}, errors.New("latency delivery queue exceeds max depth")
	}
	out := make([]LatencyDeliveryMessage, 0, len(messages))
	for _, msg := range messages {
		msg = msg.Normalize()
		if err := msg.ValidateForHeight(height, params); err != nil {
			return LatencyDeliveryQueue{}, err
		}
		fee, err := ComputeCongestionAwareForwardingFee(msg.BaseForwardingFee, msg.CongestionBps, msg.QueueDepth, params)
		if err != nil {
			return LatencyDeliveryQueue{}, err
		}
		msg.ForwardingFee = fee
		msg.PriorityScore, err = ComputeLatencyDeliveryPriority(msg, height, params)
		if err != nil {
			return LatencyDeliveryQueue{}, err
		}
		msg.MessageHash = ComputeLatencyDeliveryMessageHash(msg)
		out = append(out, msg.Normalize())
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareLatencyDeliveryMessages(out[i], out[j]) < 0
	})
	queue := LatencyDeliveryQueue{Height: height, Messages: out}
	queue.QueueRoot = ComputeLatencyDeliveryQueueRoot(queue)
	return queue, queue.Validate(params)
}

func (m LatencyDeliveryMessage) Normalize() LatencyDeliveryMessage {
	m.MessageID = normalizeLowerHex(m.MessageID)
	m.SourceZoneID = strings.TrimSpace(m.SourceZoneID)
	m.SourceShardID = strings.TrimSpace(m.SourceShardID)
	m.DestinationZoneID = strings.TrimSpace(m.DestinationZoneID)
	m.DestinationShardID = strings.TrimSpace(m.DestinationShardID)
	m.BaseForwardingFee = strings.TrimSpace(m.BaseForwardingFee)
	m.ForwardingFee = strings.TrimSpace(m.ForwardingFee)
	m.MessageHash = normalizeLowerHex(m.MessageHash)
	m.OperationClass = LatencyOperationClass(strings.TrimSpace(string(m.OperationClass)))
	return m
}

func (m LatencyDeliveryMessage) ValidateForHeight(height uint64, params CrossZoneMessageSLAParams) error {
	msg := m.Normalize()
	if err := validateHexHash("latency delivery message id", msg.MessageID); err != nil {
		return err
	}
	if !IsLatencyOperationClass(msg.OperationClass) {
		return errors.New("latency delivery operation class is unsupported")
	}
	if err := validateExecutionToken("latency delivery source zone id", msg.SourceZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("latency delivery source shard id", msg.SourceShardID); err != nil {
		return err
	}
	if err := validateExecutionToken("latency delivery destination zone id", msg.DestinationZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("latency delivery destination shard id", msg.DestinationShardID); err != nil {
		return err
	}
	if msg.CreatedHeight == 0 || msg.EligibleHeight == 0 || msg.ExpiryHeight == 0 {
		return errors.New("latency delivery heights must be positive")
	}
	if height > msg.ExpiryHeight {
		return errors.New("latency delivery message is expired")
	}
	if height < msg.EligibleHeight {
		return errors.New("latency delivery message is not eligible")
	}
	if msg.OperationClass == LatencyOpCrossZoneAsyncMessage {
		if msg.SourceCommitmentHeight == 0 {
			return errors.New("latency cross-zone delivery requires source commitment")
		}
		if msg.EligibleHeight < msg.SourceCommitmentHeight {
			return errors.New("latency cross-zone delivery eligibility precedes source commitment")
		}
		if msg.SourceCommitmentHeight-msg.CreatedHeight > params.MaxCrossZoneCommitmentLag {
			return errors.New("latency cross-zone delivery source commitment lag exceeds SLA")
		}
	}
	if msg.OperationClass == LatencyOpContractPromiseResolution {
		if !msg.PromiseResolution {
			return errors.New("latency contract promise delivery requires promise marker")
		}
		if msg.EligibleHeight <= msg.CreatedHeight {
			return errors.New("latency contract promise delivery must be future eligible")
		}
	}
	if _, err := parsePerformanceNonNegativeInt("latency delivery base forwarding fee", msg.BaseForwardingFee); err != nil {
		return err
	}
	if msg.ForwardingFee != "" {
		if _, err := parsePerformanceNonNegativeInt("latency delivery forwarding fee", msg.ForwardingFee); err != nil {
			return err
		}
	}
	if msg.CongestionBps > 10_000 {
		return errors.New("latency delivery congestion bps must be <= 10000")
	}
	return nil
}

func ComputeLatencyDeliveryPriority(msg LatencyDeliveryMessage, height uint64, params CrossZoneMessageSLAParams) (uint64, error) {
	msg = msg.Normalize()
	fee, err := parsePerformanceNonNegativeInt("latency delivery forwarding fee", msg.ForwardingFee)
	if err != nil {
		return 0, err
	}
	if !fee.IsUint64() {
		return 0, errors.New("latency delivery forwarding fee is too large")
	}
	urgency := uint64(0)
	if msg.ExpiryHeight <= height+params.NearExpiryPriorityWindow {
		urgency += params.NearExpiryPriorityBoost
	}
	if msg.ExpiryHeight > height {
		urgency += params.ExpiryPriorityWeight / (msg.ExpiryHeight - height)
	}
	feeScore := fee.Uint64() * params.FeePriorityWeight
	congestionScore := uint64(msg.CongestionBps) * params.CongestionPriorityWeight
	return urgency + feeScore + congestionScore, nil
}

func (q LatencyDeliveryQueue) Validate(params CrossZoneMessageSLAParams) error {
	if q.Height == 0 {
		return errors.New("latency delivery queue height must be positive")
	}
	if uint32(len(q.Messages)) > params.MaxDeliveryQueueDepth {
		return errors.New("latency delivery queue exceeds max depth")
	}
	for i, msg := range q.Messages {
		if err := msg.ValidateForHeight(q.Height, params); err != nil {
			return err
		}
		if msg.ForwardingFee == "" || msg.PriorityScore == 0 {
			return errors.New("latency delivery message requires computed fee and priority")
		}
		if msg.MessageHash != ComputeLatencyDeliveryMessageHash(msg) {
			return errors.New("latency delivery message hash mismatch")
		}
		if i > 0 && compareLatencyDeliveryMessages(q.Messages[i-1], msg) > 0 {
			return errors.New("latency delivery messages must be sorted canonically")
		}
	}
	if q.QueueRoot != ComputeLatencyDeliveryQueueRoot(q) {
		return errors.New("latency delivery queue root mismatch")
	}
	return nil
}

func IsLatencyOperationClass(class LatencyOperationClass) bool {
	switch class {
	case LatencyOpSingleZoneLocalTx, LatencyOpSameZoneCrossShard, LatencyOpCrossZoneAsyncMessage, LatencyOpContractPromiseResolution:
		return true
	default:
		return false
	}
}

func ComputeLatencyMetricHash(metric LatencyMetric) string {
	metric = metric.Normalize()
	return hashStrings(
		"latency-metric",
		string(metric.OperationClass),
		metric.ZoneID,
		metric.ShardID,
		metric.DestinationZoneID,
		metric.DestinationShardID,
		fmt.Sprintf("%020d", metric.CreatedHeight),
		fmt.Sprintf("%020d", metric.SourceCommitmentHeight),
		fmt.Sprintf("%020d", metric.EligibleHeight),
		fmt.Sprintf("%020d", metric.ExecutedHeight),
		fmt.Sprintf("%020d", metric.TargetBlocks),
		fmt.Sprintf("%020d", metric.ObservedBlocks),
		fmt.Sprintf("%t", metric.Satisfied),
	)
}

func ComputeLatencyDeliveryMessageHash(msg LatencyDeliveryMessage) string {
	msg = msg.Normalize()
	return hashStrings(
		"latency-delivery-message",
		msg.MessageID,
		string(msg.OperationClass),
		msg.SourceZoneID,
		msg.SourceShardID,
		msg.DestinationZoneID,
		msg.DestinationShardID,
		fmt.Sprintf("%020d", msg.CreatedHeight),
		fmt.Sprintf("%020d", msg.SourceCommitmentHeight),
		fmt.Sprintf("%020d", msg.EligibleHeight),
		fmt.Sprintf("%020d", msg.ExpiryHeight),
		msg.BaseForwardingFee,
		msg.ForwardingFee,
		fmt.Sprintf("%020d", uint64(msg.CongestionBps)),
		fmt.Sprintf("%020d", uint64(msg.QueueDepth)),
		fmt.Sprintf("%t", msg.PromiseResolution),
		fmt.Sprintf("%020d", msg.PriorityScore),
	)
}

func ComputeLatencyDeliveryQueueRoot(queue LatencyDeliveryQueue) string {
	parts := []string{"latency-delivery-queue", fmt.Sprintf("%020d", queue.Height)}
	for _, msg := range queue.Messages {
		parts = append(parts, msg.MessageHash)
	}
	return hashStrings(parts...)
}

func latencyTargetForClass(class LatencyOperationClass, params CrossZoneMessageSLAParams) uint64 {
	switch class {
	case LatencyOpSingleZoneLocalTx:
		return params.LocalTxTargetBlocks
	case LatencyOpSameZoneCrossShard:
		return params.CrossShardTargetBlocks
	case LatencyOpCrossZoneAsyncMessage:
		return params.CrossZoneTargetBlocks
	case LatencyOpContractPromiseResolution:
		return params.PromiseTargetBlocks
	default:
		return 0
	}
}

func compareLatencyDeliveryMessages(left, right LatencyDeliveryMessage) int {
	if left.PriorityScore != right.PriorityScore {
		if left.PriorityScore > right.PriorityScore {
			return -1
		}
		return 1
	}
	leftFee, _ := parsePerformanceNonNegativeInt("left latency fee", left.ForwardingFee)
	rightFee, _ := parsePerformanceNonNegativeInt("right latency fee", right.ForwardingFee)
	if !leftFee.Equal(rightFee) {
		if leftFee.GT(rightFee) {
			return -1
		}
		return 1
	}
	if left.ExpiryHeight != right.ExpiryHeight {
		return compareUint64(left.ExpiryHeight, right.ExpiryHeight)
	}
	if left.EligibleHeight != right.EligibleHeight {
		return compareUint64(left.EligibleHeight, right.EligibleHeight)
	}
	if left.MessageID < right.MessageID {
		return -1
	}
	if left.MessageID > right.MessageID {
		return 1
	}
	return 0
}
