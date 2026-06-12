package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	MempoolSeparationSpecVersion	= uint64(1)
	LatencyStrategySpecVersion	= uint64(1)
)

type MempoolMessageClass string
type LatencyOperationClass string

const (
	MempoolClassAccount	MempoolMessageClass	= "account"
	MempoolClassContract	MempoolMessageClass	= "contract"
	MempoolClassIdentity	MempoolMessageClass	= "identity"
	MempoolClassPayment	MempoolMessageClass	= "payment"
	MempoolClassSystem	MempoolMessageClass	= "system"

	LatencySingleZoneLocal		LatencyOperationClass	= "single-zone-local"
	LatencySameZoneCrossShard	LatencyOperationClass	= "same-zone-cross-shard"
	LatencyCrossZoneAsync		LatencyOperationClass	= "cross-zone-async"
	LatencyContractPromiseResolve	LatencyOperationClass	= "contract-promise-resolution"
)

type MempoolSeparationParams struct {
	MaxPerSender		uint32
	MaxPerTargetObject	uint32
	SystemShardID		ShardID
	ParamsHash		string
}

type MempoolAdmissionTx struct {
	TxHash		string
	Sender		string
	TargetZoneID	ZoneID
	TargetShardID	ShardID
	RouteKey	string
	TargetObject	string
	MessageClass	MempoolMessageClass
	FeeNAET		uint64
	ExpiryHeight	uint64
	AdmissionHeight	uint64
	PriorityClass	uint32
	TargetKnown	bool
	PreResolved	bool
}

type MempoolLane struct {
	ZoneID		ZoneID
	ShardID		ShardID
	MessageClass	MempoolMessageClass
	Transactions	[]MempoolAdmissionTx
	LaneHash	string
}

type MempoolSeparationPlan struct {
	Height		uint64
	Params		MempoolSeparationParams
	Lanes		[]MempoolLane
	PlanHash	string
}

type LatencyTarget struct {
	OperationClass	LatencyOperationClass
	TargetBlocks	uint64
	Description	string
	TargetHash	string
}

type LatencyMetric struct {
	OperationClass	LatencyOperationClass
	ObservedBlocks	uint64
	SampleCount	uint64
	Height		uint64
	MetricHash	string
}

type CrossZoneMessageSLAParams struct {
	MaxFinalityDelayBlocks	uint64
	MaxQueueDelayBlocks	uint64
	MaxDeliveryBlocks	uint64
	ParamsHash		string
}

type CongestionForwardingFeePolicy struct {
	BaseFeeNAET		uint64
	CongestionWeightBps	uint64
	NearExpiryPremiumBps	uint64
	MaxForwardingFeeNAET	uint64
	CriticalLaneDiscountBps	uint64
	PolicyHash		string
}

type DeliveryPriorityItem struct {
	MessageID		string
	OperationClass		LatencyOperationClass
	DestinationZone		ZoneID
	DestinationShard	ShardID
	FeeNAET			uint64
	CongestionScore		uint64
	ExpiryHeight		uint64
	EnqueuedHeight		uint64
	Critical		bool
	ItemHash		string
}

type LatencyStrategySpec struct {
	Version		uint64
	Targets		[]LatencyTarget
	SLA		CrossZoneMessageSLAParams
	FeePolicy	CongestionForwardingFeePolicy
	MetricsRoot	string
	DeliveryRoot	string
	Root		string
}

func DefaultMempoolSeparationParams() MempoolSeparationParams {
	params := MempoolSeparationParams{
		MaxPerSender:		1024,
		MaxPerTargetObject:	2048,
		SystemShardID:		"0",
	}
	params.ParamsHash = ComputeMempoolSeparationParamsHash(params)
	return params
}

func BuildMempoolSeparationPlan(height uint64, txs []MempoolAdmissionTx, params MempoolSeparationParams) (MempoolSeparationPlan, error) {
	if height == 0 {
		return MempoolSeparationPlan{}, errors.New("aetracore mempool separation height must be positive")
	}
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return MempoolSeparationPlan{}, err
	}
	ordered := normalizeMempoolAdmissionTxs(txs)
	senderCounts := make(map[string]uint32)
	targetCounts := make(map[string]uint32)
	laneMap := make(map[string]*MempoolLane)
	for _, tx := range ordered {
		if err := tx.Validate(params); err != nil {
			return MempoolSeparationPlan{}, err
		}
		senderCounts[tx.Sender]++
		if senderCounts[tx.Sender] > params.MaxPerSender {
			return MempoolSeparationPlan{}, errors.New("aetracore mempool sender DoS limit exceeded")
		}
		targetKey := tx.TargetZoneShardObject()
		targetCounts[targetKey]++
		if targetCounts[targetKey] > params.MaxPerTargetObject {
			return MempoolSeparationPlan{}, errors.New("aetracore mempool target object DoS limit exceeded")
		}
		laneKey := tx.LaneKey()
		lane, found := laneMap[laneKey]
		if !found {
			laneMap[laneKey] = &MempoolLane{
				ZoneID:		tx.TargetZoneID,
				ShardID:	tx.TargetShardID,
				MessageClass:	tx.MessageClass,
			}
			lane = laneMap[laneKey]
		}
		lane.Transactions = append(lane.Transactions, tx)
	}
	lanes := make([]MempoolLane, 0, len(laneMap))
	for _, lane := range laneMap {
		normalized := lane.Normalize()
		normalized.LaneHash = ComputeMempoolLaneHash(normalized)
		lanes = append(lanes, normalized)
	}
	plan := MempoolSeparationPlan{
		Height:	height,
		Params:	params,
		Lanes:	normalizeMempoolLanes(lanes),
	}
	plan.PlanHash = ComputeMempoolSeparationPlanHash(plan)
	return plan, plan.Validate()
}

func DefaultLatencyTargets() []LatencyTarget {
	return []LatencyTarget{
		latencyTarget(LatencySingleZoneLocal, 1, "Single-zone local tx finalizes in one block."),
		latencyTarget(LatencySameZoneCrossShard, 2, "Same-zone cross-shard message executes in the next eligible block."),
		latencyTarget(LatencyCrossZoneAsync, 3, "Cross-zone async message executes after source commitment and deterministic delivery."),
		latencyTarget(LatencyContractPromiseResolve, 3, "Contract promise resolution executes as a future message."),
	}
}

func DefaultCrossZoneMessageSLAParams() CrossZoneMessageSLAParams {
	sla := CrossZoneMessageSLAParams{
		MaxFinalityDelayBlocks:	1,
		MaxQueueDelayBlocks:	2,
		MaxDeliveryBlocks:	3,
	}
	sla.ParamsHash = ComputeCrossZoneMessageSLAParamsHash(sla)
	return sla
}

func DefaultCongestionForwardingFeePolicy() CongestionForwardingFeePolicy {
	policy := CongestionForwardingFeePolicy{
		BaseFeeNAET:			1,
		CongestionWeightBps:		100,
		NearExpiryPremiumBps:		250,
		MaxForwardingFeeNAET:		1_000_000,
		CriticalLaneDiscountBps:	500,
	}
	policy.PolicyHash = ComputeCongestionForwardingFeePolicyHash(policy)
	return policy
}

func BuildLatencyStrategySpec(metrics []LatencyMetric, queue []DeliveryPriorityItem) (LatencyStrategySpec, error) {
	targets := normalizeLatencyTargets(DefaultLatencyTargets())
	sla := DefaultCrossZoneMessageSLAParams()
	feePolicy := DefaultCongestionForwardingFeePolicy()
	metrics = normalizeLatencyMetrics(metrics)
	queue = NormalizeDeliveryPriorityQueue(queue)
	for _, metric := range metrics {
		if err := metric.Validate(); err != nil {
			return LatencyStrategySpec{}, err
		}
	}
	for _, item := range queue {
		if err := item.Validate(); err != nil {
			return LatencyStrategySpec{}, err
		}
	}
	spec := LatencyStrategySpec{
		Version:	LatencyStrategySpecVersion,
		Targets:	targets,
		SLA:		sla,
		FeePolicy:	feePolicy,
		MetricsRoot:	ComputeLatencyMetricsRoot(metrics),
		DeliveryRoot:	ComputeDeliveryPriorityQueueRoot(queue),
	}
	spec.Root = ComputeLatencyStrategySpecRoot(spec)
	return spec, spec.Validate()
}

func ComputeCongestionAwareForwardingFee(baseValue uint64, congestionScore uint64, blocksUntilExpiry uint64, critical bool, policy CongestionForwardingFeePolicy) (uint64, error) {
	policy = policy.Normalize()
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	fee := policy.BaseFeeNAET + baseValue
	fee += (congestionScore * policy.CongestionWeightBps) / 10_000
	if blocksUntilExpiry <= 1 {
		fee += (fee * policy.NearExpiryPremiumBps) / 10_000
	}
	if critical && policy.CriticalLaneDiscountBps > 0 {
		discount := (fee * policy.CriticalLaneDiscountBps) / 10_000
		fee -= discount
	}
	if fee > policy.MaxForwardingFeeNAET {
		return policy.MaxForwardingFeeNAET, nil
	}
	return fee, nil
}

func (p MempoolSeparationParams) Normalize() MempoolSeparationParams {
	if p.SystemShardID == "" {
		p.SystemShardID = "0"
	}
	p.ParamsHash = normalizePerformanceHash(p.ParamsHash)
	return p
}

func (p MempoolSeparationParams) Validate() error {
	p = p.Normalize()
	if p.MaxPerSender == 0 {
		return errors.New("aetracore mempool max per sender must be positive")
	}
	if p.MaxPerTargetObject == 0 {
		return errors.New("aetracore mempool max per target object must be positive")
	}
	if err := ValidateShardID(p.SystemShardID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mempool params hash", p.ParamsHash); err != nil {
		return err
	}
	if p.ParamsHash != ComputeMempoolSeparationParamsHash(p) {
		return errors.New("aetracore mempool params hash mismatch")
	}
	return nil
}

func (tx MempoolAdmissionTx) Normalize() MempoolAdmissionTx {
	tx.Sender = compactPerformanceText(tx.Sender)
	tx.RouteKey = compactPerformanceText(tx.RouteKey)
	tx.TargetObject = compactPerformanceText(tx.TargetObject)
	tx.MessageClass = MempoolMessageClass(compactPerformanceText(string(tx.MessageClass)))
	tx.TxHash = normalizePerformanceHash(tx.TxHash)
	return tx
}

func (tx MempoolAdmissionTx) Validate(params MempoolSeparationParams) error {
	tx = tx.Normalize()
	params = params.Normalize()
	if err := ValidateHash("aetracore mempool tx hash", tx.TxHash); err != nil {
		return err
	}
	if err := validateToken("aetracore mempool sender", tx.Sender, MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateZoneID(tx.TargetZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(tx.TargetShardID); err != nil {
		return err
	}
	if !IsMempoolMessageClass(tx.MessageClass) {
		return fmt.Errorf("unknown aetracore mempool message class %q", tx.MessageClass)
	}
	if tx.AdmissionHeight == 0 {
		return errors.New("aetracore mempool admission height must be positive")
	}
	if tx.ExpiryHeight != 0 && tx.ExpiryHeight < tx.AdmissionHeight {
		return errors.New("aetracore mempool expiry precedes admission")
	}
	if tx.TargetObject == "" {
		return errors.New("aetracore mempool target object is required")
	}
	if tx.RouteKey == "" && tx.TargetKnown {
		return errors.New("aetracore mempool target route key is required when known")
	}
	if !tx.TargetKnown && !tx.PreResolved && tx.TargetShardID != params.SystemShardID {
		return errors.New("aetracore mempool unknown target must be pre-resolved or routed to system shard")
	}
	return nil
}

func (tx MempoolAdmissionTx) LaneKey() string {
	tx = tx.Normalize()
	return string(tx.TargetZoneID) + "/" + string(tx.TargetShardID) + "/" + string(tx.MessageClass)
}

func (tx MempoolAdmissionTx) TargetZoneShardObject() string {
	tx = tx.Normalize()
	return string(tx.TargetZoneID) + "/" + string(tx.TargetShardID) + "/" + tx.TargetObject
}

func (l MempoolLane) Normalize() MempoolLane {
	l.MessageClass = MempoolMessageClass(compactPerformanceText(string(l.MessageClass)))
	l.Transactions = normalizeMempoolAdmissionTxs(l.Transactions)
	l.LaneHash = normalizePerformanceHash(l.LaneHash)
	return l
}

func (l MempoolLane) Validate(params MempoolSeparationParams) error {
	l = l.Normalize()
	if err := ValidateZoneID(l.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(l.ShardID); err != nil {
		return err
	}
	if !IsMempoolMessageClass(l.MessageClass) {
		return fmt.Errorf("unknown aetracore mempool lane message class %q", l.MessageClass)
	}
	if len(l.Transactions) == 0 {
		return errors.New("aetracore mempool lane requires transactions")
	}
	for _, tx := range l.Transactions {
		if err := tx.Validate(params); err != nil {
			return err
		}
		if tx.TargetZoneID != l.ZoneID || tx.TargetShardID != l.ShardID || tx.MessageClass != l.MessageClass {
			return errors.New("aetracore mempool lane transaction route mismatch")
		}
	}
	if err := ValidateHash("aetracore mempool lane hash", l.LaneHash); err != nil {
		return err
	}
	if l.LaneHash != ComputeMempoolLaneHash(l) {
		return errors.New("aetracore mempool lane hash mismatch")
	}
	return nil
}

func (p MempoolSeparationPlan) Normalize() MempoolSeparationPlan {
	p.Params = p.Params.Normalize()
	p.Lanes = normalizeMempoolLanes(p.Lanes)
	p.PlanHash = normalizePerformanceHash(p.PlanHash)
	return p
}

func (p MempoolSeparationPlan) Validate() error {
	p = p.Normalize()
	if p.Height == 0 {
		return errors.New("aetracore mempool separation height must be positive")
	}
	if err := p.Params.Validate(); err != nil {
		return err
	}
	if len(p.Lanes) == 0 {
		return errors.New("aetracore mempool separation plan requires lanes")
	}
	var previous string
	for _, lane := range p.Lanes {
		if err := lane.Validate(p.Params); err != nil {
			return err
		}
		key := laneSortKey(lane)
		if previous != "" && previous >= key {
			return errors.New("aetracore mempool lanes must be sorted canonically")
		}
		previous = key
	}
	if err := ValidateHash("aetracore mempool separation plan hash", p.PlanHash); err != nil {
		return err
	}
	if p.PlanHash != ComputeMempoolSeparationPlanHash(p) {
		return errors.New("aetracore mempool separation plan hash mismatch")
	}
	return nil
}

func (t LatencyTarget) Normalize() LatencyTarget {
	t.Description = compactPerformanceText(t.Description)
	t.TargetHash = normalizePerformanceHash(t.TargetHash)
	return t
}

func (t LatencyTarget) Validate() error {
	t = t.Normalize()
	if !IsLatencyOperationClass(t.OperationClass) {
		return fmt.Errorf("unknown aetracore latency operation class %q", t.OperationClass)
	}
	if t.TargetBlocks == 0 {
		return errors.New("aetracore latency target blocks must be positive")
	}
	if t.Description == "" {
		return errors.New("aetracore latency target description is required")
	}
	if err := ValidateHash("aetracore latency target hash", t.TargetHash); err != nil {
		return err
	}
	if t.TargetHash != ComputeLatencyTargetHash(t) {
		return errors.New("aetracore latency target hash mismatch")
	}
	return nil
}

func (m LatencyMetric) Normalize() LatencyMetric {
	m.MetricHash = normalizePerformanceHash(m.MetricHash)
	return m
}

func (m LatencyMetric) Validate() error {
	m = m.Normalize()
	if !IsLatencyOperationClass(m.OperationClass) {
		return fmt.Errorf("unknown aetracore latency metric operation class %q", m.OperationClass)
	}
	if m.SampleCount == 0 {
		return errors.New("aetracore latency metric sample count must be positive")
	}
	if m.Height == 0 {
		return errors.New("aetracore latency metric height must be positive")
	}
	if err := ValidateHash("aetracore latency metric hash", m.MetricHash); err != nil {
		return err
	}
	if m.MetricHash != ComputeLatencyMetricHash(m) {
		return errors.New("aetracore latency metric hash mismatch")
	}
	return nil
}

func (s CrossZoneMessageSLAParams) Normalize() CrossZoneMessageSLAParams {
	s.ParamsHash = normalizePerformanceHash(s.ParamsHash)
	return s
}

func (s CrossZoneMessageSLAParams) Validate() error {
	s = s.Normalize()
	if s.MaxFinalityDelayBlocks == 0 || s.MaxQueueDelayBlocks == 0 || s.MaxDeliveryBlocks == 0 {
		return errors.New("aetracore cross-zone SLA bounds must be positive")
	}
	if s.MaxDeliveryBlocks < s.MaxFinalityDelayBlocks+s.MaxQueueDelayBlocks {
		return errors.New("aetracore cross-zone SLA delivery bound is too low")
	}
	if err := ValidateHash("aetracore cross-zone SLA hash", s.ParamsHash); err != nil {
		return err
	}
	if s.ParamsHash != ComputeCrossZoneMessageSLAParamsHash(s) {
		return errors.New("aetracore cross-zone SLA hash mismatch")
	}
	return nil
}

func (p CongestionForwardingFeePolicy) Normalize() CongestionForwardingFeePolicy {
	p.PolicyHash = normalizePerformanceHash(p.PolicyHash)
	return p
}

func (p CongestionForwardingFeePolicy) Validate() error {
	p = p.Normalize()
	if p.BaseFeeNAET == 0 {
		return errors.New("aetracore forwarding base fee must be positive")
	}
	if p.MaxForwardingFeeNAET < p.BaseFeeNAET {
		return errors.New("aetracore forwarding max fee must cover base fee")
	}
	if p.CriticalLaneDiscountBps > 10_000 {
		return errors.New("aetracore forwarding critical discount must be <= 10000 bps")
	}
	if err := ValidateHash("aetracore forwarding fee policy hash", p.PolicyHash); err != nil {
		return err
	}
	if p.PolicyHash != ComputeCongestionForwardingFeePolicyHash(p) {
		return errors.New("aetracore forwarding fee policy hash mismatch")
	}
	return nil
}

func (d DeliveryPriorityItem) Normalize() DeliveryPriorityItem {
	d.MessageID = normalizePerformanceHash(d.MessageID)
	d.ItemHash = normalizePerformanceHash(d.ItemHash)
	return d
}

func (d DeliveryPriorityItem) Validate() error {
	d = d.Normalize()
	if err := ValidateHash("aetracore delivery message id", d.MessageID); err != nil {
		return err
	}
	if !IsLatencyOperationClass(d.OperationClass) {
		return fmt.Errorf("unknown aetracore delivery operation class %q", d.OperationClass)
	}
	if err := ValidateZoneID(d.DestinationZone); err != nil {
		return err
	}
	if err := ValidateShardID(d.DestinationShard); err != nil {
		return err
	}
	if d.ExpiryHeight == 0 || d.EnqueuedHeight == 0 {
		return errors.New("aetracore delivery expiry and enqueue height must be positive")
	}
	if d.ExpiryHeight < d.EnqueuedHeight {
		return errors.New("aetracore delivery expiry precedes enqueue height")
	}
	if err := ValidateHash("aetracore delivery item hash", d.ItemHash); err != nil {
		return err
	}
	if d.ItemHash != ComputeDeliveryPriorityItemHash(d) {
		return errors.New("aetracore delivery item hash mismatch")
	}
	return nil
}

func (s LatencyStrategySpec) Normalize() LatencyStrategySpec {
	if s.Version == 0 {
		s.Version = LatencyStrategySpecVersion
	}
	s.Targets = normalizeLatencyTargets(s.Targets)
	s.SLA = s.SLA.Normalize()
	s.FeePolicy = s.FeePolicy.Normalize()
	s.MetricsRoot = normalizePerformanceHash(s.MetricsRoot)
	s.DeliveryRoot = normalizePerformanceHash(s.DeliveryRoot)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s LatencyStrategySpec) Validate() error {
	s = s.Normalize()
	if s.Version != LatencyStrategySpecVersion {
		return fmt.Errorf("aetracore latency strategy version must be %d", LatencyStrategySpecVersion)
	}
	if err := s.SLA.Validate(); err != nil {
		return err
	}
	if err := s.FeePolicy.Validate(); err != nil {
		return err
	}
	if len(s.Targets) != 4 {
		return errors.New("aetracore latency strategy requires all operation targets")
	}
	seen := make(map[LatencyOperationClass]struct{}, len(s.Targets))
	for _, target := range s.Targets {
		if err := target.Validate(); err != nil {
			return err
		}
		seen[target.OperationClass] = struct{}{}
	}
	for _, class := range []LatencyOperationClass{LatencySingleZoneLocal, LatencySameZoneCrossShard, LatencyCrossZoneAsync, LatencyContractPromiseResolve} {
		if _, found := seen[class]; !found {
			return fmt.Errorf("aetracore latency strategy missing target %s", class)
		}
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"aetracore latency metrics root", s.MetricsRoot},
		{"aetracore latency delivery root", s.DeliveryRoot},
		{"aetracore latency strategy root", s.Root},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if s.Root != ComputeLatencyStrategySpecRoot(s) {
		return errors.New("aetracore latency strategy root mismatch")
	}
	return nil
}

func IsMempoolMessageClass(class MempoolMessageClass) bool {
	switch class {
	case MempoolClassAccount, MempoolClassContract, MempoolClassIdentity, MempoolClassPayment, MempoolClassSystem:
		return true
	default:
		return false
	}
}

func IsLatencyOperationClass(class LatencyOperationClass) bool {
	switch class {
	case LatencySingleZoneLocal, LatencySameZoneCrossShard, LatencyCrossZoneAsync, LatencyContractPromiseResolve:
		return true
	default:
		return false
	}
}

func ComputeMempoolSeparationParamsHash(params MempoolSeparationParams) string {
	params = params.Normalize()
	return hashParts("aetra-mempool-separation-params-v1", fmt.Sprint(params.MaxPerSender), fmt.Sprint(params.MaxPerTargetObject), string(params.SystemShardID))
}

func ComputeMempoolAdmissionTxHash(tx MempoolAdmissionTx) string {
	tx = tx.Normalize()
	return hashParts(
		"aetra-mempool-admission-tx-v1",
		tx.TxHash,
		tx.Sender,
		string(tx.TargetZoneID),
		string(tx.TargetShardID),
		tx.RouteKey,
		tx.TargetObject,
		string(tx.MessageClass),
		fmt.Sprint(tx.FeeNAET),
		fmt.Sprint(tx.ExpiryHeight),
		fmt.Sprint(tx.AdmissionHeight),
		fmt.Sprint(tx.PriorityClass),
		fmt.Sprint(tx.TargetKnown),
		fmt.Sprint(tx.PreResolved),
	)
}

func ComputeMempoolLaneHash(lane MempoolLane) string {
	lane = lane.Normalize()
	parts := []string{"aetra-mempool-lane-v1", string(lane.ZoneID), string(lane.ShardID), string(lane.MessageClass), fmt.Sprint(len(lane.Transactions))}
	for _, tx := range lane.Transactions {
		parts = append(parts, ComputeMempoolAdmissionTxHash(tx))
	}
	return hashParts(parts...)
}

func ComputeMempoolSeparationPlanHash(plan MempoolSeparationPlan) string {
	plan = plan.Normalize()
	parts := []string{"aetra-mempool-separation-plan-v1", fmt.Sprint(plan.Height), plan.Params.ParamsHash, fmt.Sprint(len(plan.Lanes))}
	for _, lane := range plan.Lanes {
		parts = append(parts, lane.LaneHash)
	}
	return hashParts(parts...)
}

func ComputeLatencyTargetHash(target LatencyTarget) string {
	target = target.Normalize()
	return hashParts("aetra-latency-target-v1", string(target.OperationClass), fmt.Sprint(target.TargetBlocks), target.Description)
}

func ComputeLatencyMetricHash(metric LatencyMetric) string {
	metric = metric.Normalize()
	return hashParts("aetra-latency-metric-v1", string(metric.OperationClass), fmt.Sprint(metric.ObservedBlocks), fmt.Sprint(metric.SampleCount), fmt.Sprint(metric.Height))
}

func ComputeCrossZoneMessageSLAParamsHash(sla CrossZoneMessageSLAParams) string {
	sla = sla.Normalize()
	return hashParts("aetra-cross-zone-message-sla-v1", fmt.Sprint(sla.MaxFinalityDelayBlocks), fmt.Sprint(sla.MaxQueueDelayBlocks), fmt.Sprint(sla.MaxDeliveryBlocks))
}

func ComputeCongestionForwardingFeePolicyHash(policy CongestionForwardingFeePolicy) string {
	policy = policy.Normalize()
	return hashParts("aetra-congestion-forwarding-fee-policy-v1", fmt.Sprint(policy.BaseFeeNAET), fmt.Sprint(policy.CongestionWeightBps), fmt.Sprint(policy.NearExpiryPremiumBps), fmt.Sprint(policy.MaxForwardingFeeNAET), fmt.Sprint(policy.CriticalLaneDiscountBps))
}

func ComputeDeliveryPriorityItemHash(item DeliveryPriorityItem) string {
	item = item.Normalize()
	return hashParts("aetra-delivery-priority-item-v1", item.MessageID, string(item.OperationClass), string(item.DestinationZone), string(item.DestinationShard), fmt.Sprint(item.FeeNAET), fmt.Sprint(item.CongestionScore), fmt.Sprint(item.ExpiryHeight), fmt.Sprint(item.EnqueuedHeight), fmt.Sprint(item.Critical))
}

func ComputeLatencyMetricsRoot(metrics []LatencyMetric) string {
	ordered := normalizeLatencyMetrics(metrics)
	parts := []string{"aetra-latency-metrics-root-v1", fmt.Sprint(len(ordered))}
	for _, metric := range ordered {
		parts = append(parts, metric.MetricHash)
	}
	return hashParts(parts...)
}

func ComputeDeliveryPriorityQueueRoot(queue []DeliveryPriorityItem) string {
	ordered := NormalizeDeliveryPriorityQueue(queue)
	parts := []string{"aetra-delivery-priority-queue-root-v1", fmt.Sprint(len(ordered))}
	for _, item := range ordered {
		parts = append(parts, item.ItemHash)
	}
	return hashParts(parts...)
}

func ComputeLatencyStrategySpecRoot(spec LatencyStrategySpec) string {
	spec = spec.Normalize()
	parts := []string{"aetra-latency-strategy-spec-v1", fmt.Sprint(spec.Version), spec.SLA.ParamsHash, spec.FeePolicy.PolicyHash, spec.MetricsRoot, spec.DeliveryRoot}
	for _, target := range spec.Targets {
		parts = append(parts, target.TargetHash)
	}
	return hashParts(parts...)
}

func latencyTarget(class LatencyOperationClass, blocks uint64, description string) LatencyTarget {
	target := LatencyTarget{OperationClass: class, TargetBlocks: blocks, Description: description}.Normalize()
	target.TargetHash = ComputeLatencyTargetHash(target)
	return target
}

func latencyMetric(class LatencyOperationClass, observed uint64, samples uint64, height uint64) LatencyMetric {
	metric := LatencyMetric{OperationClass: class, ObservedBlocks: observed, SampleCount: samples, Height: height}.Normalize()
	metric.MetricHash = ComputeLatencyMetricHash(metric)
	return metric
}

func deliveryPriorityItem(messageID string, class LatencyOperationClass, zone ZoneID, shard ShardID, fee uint64, congestion uint64, expiry uint64, enqueued uint64, critical bool) DeliveryPriorityItem {
	item := DeliveryPriorityItem{MessageID: messageID, OperationClass: class, DestinationZone: zone, DestinationShard: shard, FeeNAET: fee, CongestionScore: congestion, ExpiryHeight: expiry, EnqueuedHeight: enqueued, Critical: critical}.Normalize()
	item.ItemHash = ComputeDeliveryPriorityItemHash(item)
	return item
}

func normalizeMempoolAdmissionTxs(values []MempoolAdmissionTx) []MempoolAdmissionTx {
	out := make([]MempoolAdmissionTx, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareMempoolAdmissionTx(out[i], out[j]) < 0
	})
	return out
}

func normalizeMempoolLanes(values []MempoolLane) []MempoolLane {
	out := make([]MempoolLane, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.LaneHash == "" {
			normalized.LaneHash = ComputeMempoolLaneHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return laneSortKey(out[i]) < laneSortKey(out[j])
	})
	return out
}

func normalizeLatencyTargets(values []LatencyTarget) []LatencyTarget {
	out := make([]LatencyTarget, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.TargetHash == "" {
			normalized.TargetHash = ComputeLatencyTargetHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].OperationClass < out[j].OperationClass
	})
	return out
}

func normalizeLatencyMetrics(values []LatencyMetric) []LatencyMetric {
	out := make([]LatencyMetric, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.MetricHash == "" {
			normalized.MetricHash = ComputeLatencyMetricHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].OperationClass != out[j].OperationClass {
			return out[i].OperationClass < out[j].OperationClass
		}
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].MetricHash < out[j].MetricHash
	})
	return out
}

func NormalizeDeliveryPriorityQueue(values []DeliveryPriorityItem) []DeliveryPriorityItem {
	out := make([]DeliveryPriorityItem, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.ItemHash == "" {
			normalized.ItemHash = ComputeDeliveryPriorityItemHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareDeliveryPriorityItems(out[i], out[j]) < 0
	})
	return out
}

func compareMempoolAdmissionTx(left, right MempoolAdmissionTx) int {
	for _, pair := range [][2]string{
		{string(left.TargetZoneID), string(right.TargetZoneID)},
		{string(left.TargetShardID), string(right.TargetShardID)},
		{string(left.MessageClass), string(right.MessageClass)},
	} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	if left.PriorityClass != right.PriorityClass {
		if left.PriorityClass < right.PriorityClass {
			return -1
		}
		return 1
	}
	if left.ExpiryHeight != right.ExpiryHeight {
		if left.ExpiryHeight == 0 {
			return 1
		}
		if right.ExpiryHeight == 0 || left.ExpiryHeight < right.ExpiryHeight {
			return -1
		}
		return 1
	}
	if left.FeeNAET != right.FeeNAET {
		if left.FeeNAET > right.FeeNAET {
			return -1
		}
		return 1
	}
	if left.AdmissionHeight != right.AdmissionHeight {
		if left.AdmissionHeight < right.AdmissionHeight {
			return -1
		}
		return 1
	}
	if left.TxHash < right.TxHash {
		return -1
	}
	if left.TxHash > right.TxHash {
		return 1
	}
	return 0
}

func compareDeliveryPriorityItems(left, right DeliveryPriorityItem) int {
	if left.Critical != right.Critical {
		if left.Critical {
			return -1
		}
		return 1
	}
	if left.ExpiryHeight != right.ExpiryHeight {
		if left.ExpiryHeight < right.ExpiryHeight {
			return -1
		}
		return 1
	}
	if left.CongestionScore != right.CongestionScore {
		if left.CongestionScore > right.CongestionScore {
			return -1
		}
		return 1
	}
	if left.FeeNAET != right.FeeNAET {
		if left.FeeNAET > right.FeeNAET {
			return -1
		}
		return 1
	}
	if left.EnqueuedHeight != right.EnqueuedHeight {
		if left.EnqueuedHeight < right.EnqueuedHeight {
			return -1
		}
		return 1
	}
	if left.MessageID < right.MessageID {
		return -1
	}
	if left.MessageID > right.MessageID {
		return 1
	}
	return 0
}

func laneSortKey(lane MempoolLane) string {
	return string(lane.ZoneID) + "/" + string(lane.ShardID) + "/" + string(lane.MessageClass)
}
