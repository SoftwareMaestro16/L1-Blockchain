package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	aetracoretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

type AetherMessageLifecycleRecord struct {
	MsgID		string
	Stage		MessageLifecycleStage
	Height		uint64
	RouteCommitment	string
	ReceiptHash	string
	RecordHash	string
}

type AetherRoutingParams struct {
	MaxHopCount		uint32
	BaseHopCost		uint64
	CongestionWeight	uint64
	QueueWeight		uint64
	LatencyWeight		uint64
	CapacityPenalty		uint64
	RequiredCapacity	uint64
	FailureRateWeight	uint64
	GasUtilizationWeight	uint64
	ExpiryRateWeight	uint64
	FairnessCredit		uint64
	CriticalPriorityCredit	uint64
	NormalPriorityFloor	uint64
	GovernanceHash		string
}

type AetherRoutingMetric struct {
	ZoneID			zonestypes.ZoneID
	ShardID			string
	CommittedHeight		uint64
	OutboxBacklog		uint64
	InboxBacklog		uint64
	AverageExecutionDelay	uint64
	FailedDeliveryRate	uint64
	ShardGasUtilization	uint64
	MessageExpiryRate	uint64
	Capacity		uint64
	CongestionScore		uint64
	QueueBacklog		uint64
	FairnessCredit		uint64
	CriticalPriorityLane	bool
	MetricHash		string
}

type AetherRoutingEdge struct {
	FromZoneID	zonestypes.ZoneID
	FromShardID	string
	ToZoneID	zonestypes.ZoneID
	ToShardID	string
}

type AetherRoutingHop struct {
	ZoneID		zonestypes.ZoneID
	ShardID		string
	Coordinate	string
}

type AetherRouteCandidate struct {
	Path			[]AetherRoutingHop
	HopCount		uint32
	TotalCost		uint64
	PathCommitment		string
	DestinationShardID	string
}

type AetherRoutePlan struct {
	RoutingEpoch		uint64
	RoutingTableHash	string
	MessageClassHash	string
	SelectedPath		AetherRouteCandidate
	RouteCommitment		string
}

func NewAetherMessageLifecycleRecord(record AetherMessageLifecycleRecord) (AetherMessageLifecycleRecord, error) {
	if record.RecordHash != "" {
		return AetherMessageLifecycleRecord{}, errors.New("aether lifecycle record hash must be empty before construction")
	}
	record = normalizeAetherLifecycleRecord(record)
	if err := record.ValidateForHash(); err != nil {
		return AetherMessageLifecycleRecord{}, err
	}
	record.RecordHash = ComputeAetherMessageLifecycleRecordHash(record)
	return record, record.Validate()
}

func ValidateAetherMessageLifecycle(records []AetherMessageLifecycleRecord) error {
	ordered := cloneAetherLifecycleRecords(records)
	sortAetherLifecycleRecords(ordered)
	expected := []MessageLifecycleStage{
		MessageLifecycleCreated,
		MessageLifecycleQueuedInSourceOutbox,
		MessageLifecycleCommittedInMessageRoot,
		MessageLifecycleEligibleForDelivery,
		MessageLifecycleQueuedInDestinationInbox,
		MessageLifecycleExecutedOrFailed,
		MessageLifecycleReceipt,
		MessageLifecycleBounceOrFinalize,
	}
	if len(ordered) != len(expected) {
		return errors.New("aether lifecycle requires all canonical stages")
	}
	var msgID string
	var previousHeight uint64
	for i, record := range ordered {
		if err := record.Validate(); err != nil {
			return err
		}
		if record.Stage != expected[i] {
			return fmt.Errorf("aether lifecycle stage %d must be %s", i, expected[i])
		}
		if msgID == "" {
			msgID = record.MsgID
		}
		if record.MsgID != msgID {
			return errors.New("aether lifecycle message id mismatch")
		}
		if i > 0 && record.Height < previousHeight {
			return errors.New("aether lifecycle heights must not decrease")
		}
		previousHeight = record.Height
	}
	return nil
}

func ComputeAetherLifecycleRoot(records []AetherMessageLifecycleRecord) (string, error) {
	if err := ValidateAetherMessageLifecycle(records); err != nil {
		return "", err
	}
	ordered := cloneAetherLifecycleRecords(records)
	sortAetherLifecycleRecords(ordered)
	parts := []string{"aetra-aether-message-lifecycle-root-v1", fmt.Sprint(len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.RecordHash)
	}
	return hashParts(parts...), nil
}

func CommitAetherMessageDeterministicRoute(msg AetherMessage, table aetracoretypes.RoutingTableCommitment, metrics []AetherRoutingMetric, adjacency []AetherRoutingEdge, params AetherRoutingParams) (AetherMessage, AetherRoutePlan, error) {
	if msg.MsgID != "" {
		return AetherMessage{}, AetherRoutePlan{}, errors.New("aether route planning requires message id to be empty")
	}
	msg = normalizeAetherMessage(msg)
	if msg.TraceID == "" {
		msg.TraceID = ComputeAetherTraceID(msg)
	}
	msg.RouteCommitment = EmptyHash()
	if err := msg.ValidateForID(); err != nil {
		return AetherMessage{}, AetherRoutePlan{}, err
	}
	plan, err := SelectDeterministicAetherRoute(msg, table, metrics, adjacency, params)
	if err != nil {
		return AetherMessage{}, AetherRoutePlan{}, err
	}
	msg.RouteCommitment = plan.RouteCommitment
	committed, err := NewAetherMessage(msg)
	if err != nil {
		return AetherMessage{}, AetherRoutePlan{}, err
	}
	return committed, plan, nil
}

func SelectDeterministicAetherRoute(msg AetherMessage, table aetracoretypes.RoutingTableCommitment, metrics []AetherRoutingMetric, adjacency []AetherRoutingEdge, params AetherRoutingParams) (AetherRoutePlan, error) {
	msg = normalizeAetherMessage(msg)
	if err := msg.ValidateForID(); err != nil {
		return AetherRoutePlan{}, err
	}
	if err := table.ValidateHash(); err != nil {
		return AetherRoutePlan{}, err
	}
	if err := params.Validate(); err != nil {
		return AetherRoutePlan{}, err
	}
	if !routingTableHasZone(table, msg.SenderZoneID) || !routingTableHasZone(table, msg.ReceiverZoneID) {
		return AetherRoutePlan{}, errors.New("aether route missing source or destination zone in committed routing table")
	}
	metricMap, err := routingMetricMap(metrics)
	if err != nil {
		return AetherRoutePlan{}, err
	}
	source := AetherRoutingHop{
		ZoneID:		msg.SenderZoneID,
		ShardID:	msg.SenderShardID,
		Coordinate:	ComputeAetherRouteCoordinate(table, msg.SenderZoneID, msg.SenderShardID),
	}
	destination := AetherRoutingHop{
		ZoneID:		msg.ReceiverZoneID,
		ShardID:	msg.ReceiverShardID,
		Coordinate:	ComputeAetherRouteCoordinate(table, msg.ReceiverZoneID, msg.ReceiverShardID),
	}
	paths := generateAetherCandidatePaths(source, destination, adjacency, params.MaxHopCount)
	if len(paths) == 0 {
		return AetherRoutePlan{}, errors.New("aether route has no candidate path within hop limit")
	}
	candidates := make([]AetherRouteCandidate, 0, len(paths))
	for _, path := range paths {
		candidate := buildAetherRouteCandidate(path, table, metricMap, params)
		if candidate.HopCount <= params.MaxHopCount {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return AetherRoutePlan{}, errors.New("aether route candidates exceed hop limit")
	}
	sortAetherRouteCandidates(candidates)
	plan := AetherRoutePlan{
		RoutingEpoch:		table.RoutingEpoch,
		RoutingTableHash:	table.TableHash,
		MessageClassHash:	ComputeAetherMessageClassHash(msg),
		SelectedPath:		candidates[0],
	}
	plan.RouteCommitment = ComputeAetherRoutePlanCommitment(plan)
	return plan, plan.Validate()
}

func (p AetherRoutingParams) Validate() error {
	if p.MaxHopCount == 0 {
		return errors.New("aether routing max hop count must be positive")
	}
	if p.BaseHopCost == 0 {
		return errors.New("aether routing base hop cost must be positive")
	}
	if p.GovernanceHash != "" {
		return zonestypes.ValidateHash("aether routing governance hash", p.GovernanceHash)
	}
	if p.NormalPriorityFloor > p.CriticalPriorityCredit {
		return errors.New("aether routing normal priority floor must not exceed critical priority credit")
	}
	return nil
}

func (m AetherRoutingMetric) Validate() error {
	if err := zonestypes.ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	if err := validateToken("aether routing metric shard id", m.ShardID, MaxShardIDLength); err != nil {
		return err
	}
	if m.CommittedHeight == 0 {
		return errors.New("aether routing metric height must be positive")
	}
	if m.MetricHash != "" {
		if err := zonestypes.ValidateHash("aether routing metric hash", m.MetricHash); err != nil {
			return err
		}
		if m.MetricHash != ComputeAetherRoutingMetricHash(m) {
			return errors.New("aether routing metric hash mismatch")
		}
	}
	return nil
}

func (p AetherRoutePlan) Validate() error {
	if p.RoutingEpoch == 0 {
		return errors.New("aether route plan routing epoch must be positive")
	}
	if err := zonestypes.ValidateHash("aether route plan table hash", p.RoutingTableHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("aether route plan message class hash", p.MessageClassHash); err != nil {
		return err
	}
	if err := p.SelectedPath.Validate(); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("aether route plan commitment", p.RouteCommitment); err != nil {
		return err
	}
	if p.RouteCommitment != ComputeAetherRoutePlanCommitment(p) {
		return errors.New("aether route plan commitment mismatch")
	}
	return nil
}

func (c AetherRouteCandidate) Validate() error {
	if len(c.Path) < 2 {
		return errors.New("aether route candidate requires at least source and destination")
	}
	if c.HopCount != uint32(len(c.Path)-1) {
		return errors.New("aether route candidate hop count mismatch")
	}
	if err := zonestypes.ValidateHash("aether route candidate path commitment", c.PathCommitment); err != nil {
		return err
	}
	if c.DestinationShardID == "" {
		return errors.New("aether route candidate destination shard is required")
	}
	return nil
}

func (r AetherMessageLifecycleRecord) Validate() error {
	r = normalizeAetherLifecycleRecord(r)
	if err := r.ValidateForHash(); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("aether lifecycle record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAetherMessageLifecycleRecordHash(r) {
		return errors.New("aether lifecycle record hash mismatch")
	}
	return nil
}

func (r AetherMessageLifecycleRecord) ValidateForHash() error {
	if err := zonestypes.ValidateHash("aether lifecycle message id", r.MsgID); err != nil {
		return err
	}
	if !IsMessageLifecycleStage(r.Stage) {
		return fmt.Errorf("unknown aether lifecycle stage %q", r.Stage)
	}
	if r.Height == 0 {
		return errors.New("aether lifecycle height must be positive")
	}
	if r.RouteCommitment != "" {
		if err := zonestypes.ValidateHash("aether lifecycle route commitment", r.RouteCommitment); err != nil {
			return err
		}
	}
	if r.ReceiptHash != "" {
		if err := zonestypes.ValidateHash("aether lifecycle receipt hash", r.ReceiptHash); err != nil {
			return err
		}
	}
	return nil
}

func ComputeAetherMessageLifecycleRecordHash(record AetherMessageLifecycleRecord) string {
	record = normalizeAetherLifecycleRecord(record)
	return hashParts("aetra-aether-message-lifecycle-record-v1", record.MsgID, string(record.Stage), fmt.Sprint(record.Height), record.RouteCommitment, record.ReceiptHash)
}

func ComputeAetherRouteCoordinate(table aetracoretypes.RoutingTableCommitment, zoneID zonestypes.ZoneID, shardID string) string {
	return hashParts("aetra-aether-route-coordinate-v1", table.TableHash, fmt.Sprint(table.RoutingEpoch), string(zoneID), shardID)
}

func ComputeAetherMessageClassHash(msg AetherMessage) string {
	msg = normalizeAetherMessage(msg)
	return hashParts("aetra-aether-message-class-v1", msg.PayloadType, string(msg.ExecutionMode), string(msg.OrderingClass))
}

func ComputeAetherRoutingMetricHash(metric AetherRoutingMetric) string {
	metric = normalizeAetherRoutingMetric(metric)
	return hashParts(
		"aetra-aether-routing-metric-v2",
		string(metric.ZoneID),
		metric.ShardID,
		fmt.Sprint(metric.CommittedHeight),
		fmt.Sprint(metric.OutboxBacklog),
		fmt.Sprint(metric.InboxBacklog),
		fmt.Sprint(metric.AverageExecutionDelay),
		fmt.Sprint(metric.FailedDeliveryRate),
		fmt.Sprint(metric.ShardGasUtilization),
		fmt.Sprint(metric.MessageExpiryRate),
		fmt.Sprint(metric.Capacity),
		fmt.Sprint(metric.CongestionScore),
		fmt.Sprint(metric.QueueBacklog),
		fmt.Sprint(metric.FairnessCredit),
		fmt.Sprint(metric.CriticalPriorityLane),
	)
}

func ComputeAetherRoutePlanCommitment(plan AetherRoutePlan) string {
	return hashParts("aetra-aether-route-plan-v1", fmt.Sprint(plan.RoutingEpoch), plan.RoutingTableHash, plan.MessageClassHash, plan.SelectedPath.PathCommitment, fmt.Sprint(plan.SelectedPath.TotalCost), fmt.Sprint(plan.SelectedPath.HopCount), plan.SelectedPath.DestinationShardID)
}

func ComputeAetherPathCommitment(path []AetherRoutingHop) string {
	parts := []string{"aetra-aether-route-path-v1", fmt.Sprint(len(path))}
	for _, hop := range path {
		parts = append(parts, string(hop.ZoneID), hop.ShardID, hop.Coordinate)
	}
	return hashParts(parts...)
}

func buildAetherRouteCandidate(path []AetherRoutingHop, table aetracoretypes.RoutingTableCommitment, metricMap map[string]AetherRoutingMetric, params AetherRoutingParams) AetherRouteCandidate {
	hops := cloneAetherRoutingHops(path)
	hopCount := uint32(len(hops) - 1)
	total := uint64(hopCount)*params.BaseHopCost + uint64(hopCount)*params.LatencyWeight
	for i := 1; i < len(hops); i++ {
		metric := normalizeAetherRoutingMetric(metricMap[routingNodeKey(hops[i].ZoneID, hops[i].ShardID)])
		total += params.CongestionWeight * metric.CongestionScore
		total += params.QueueWeight * (metric.OutboxBacklog + metric.InboxBacklog + metric.QueueBacklog)
		total += params.LatencyWeight * metric.AverageExecutionDelay
		total += params.FailureRateWeight * metric.FailedDeliveryRate
		total += params.GasUtilizationWeight * metric.ShardGasUtilization
		total += params.ExpiryRateWeight * metric.MessageExpiryRate
		if params.RequiredCapacity > 0 && metric.Capacity < params.RequiredCapacity {
			total += params.CapacityPenalty
		}
		if metric.FairnessCredit > 0 && params.FairnessCredit > 0 {
			total = subtractCostFloor(total, params.FairnessCredit*metric.FairnessCredit, params.NormalPriorityFloor)
		}
		if metric.CriticalPriorityLane && params.CriticalPriorityCredit > 0 {
			total = subtractCostFloor(total, params.CriticalPriorityCredit, params.NormalPriorityFloor)
		}
	}
	for i := range hops {
		hops[i].Coordinate = ComputeAetherRouteCoordinate(table, hops[i].ZoneID, hops[i].ShardID)
	}
	return AetherRouteCandidate{
		Path:			hops,
		HopCount:		hopCount,
		TotalCost:		total,
		PathCommitment:		ComputeAetherPathCommitment(hops),
		DestinationShardID:	hops[len(hops)-1].ShardID,
	}
}

func generateAetherCandidatePaths(source AetherRoutingHop, destination AetherRoutingHop, adjacency []AetherRoutingEdge, maxHopCount uint32) [][]AetherRoutingHop {
	edges := normalizeAetherRoutingEdges(adjacency)
	if len(edges) == 0 {
		return [][]AetherRoutingHop{{source, destination}}
	}
	paths := make([][]AetherRoutingHop, 0)
	visited := map[string]struct{}{routingNodeKey(source.ZoneID, source.ShardID): {}}
	var walk func(AetherRoutingHop, []AetherRoutingHop)
	walk = func(current AetherRoutingHop, path []AetherRoutingHop) {
		if uint32(len(path)-1) > maxHopCount {
			return
		}
		if current.ZoneID == destination.ZoneID && current.ShardID == destination.ShardID {
			paths = append(paths, cloneAetherRoutingHops(path))
			return
		}
		for _, edge := range edges {
			if edge.FromZoneID != current.ZoneID || edge.FromShardID != current.ShardID {
				continue
			}
			next := AetherRoutingHop{ZoneID: edge.ToZoneID, ShardID: edge.ToShardID}
			key := routingNodeKey(next.ZoneID, next.ShardID)
			if _, found := visited[key]; found {
				continue
			}
			visited[key] = struct{}{}
			walk(next, append(path, next))
			delete(visited, key)
		}
	}
	walk(source, []AetherRoutingHop{source})
	return paths
}

func routingMetricMap(metrics []AetherRoutingMetric) (map[string]AetherRoutingMetric, error) {
	out := make(map[string]AetherRoutingMetric, len(metrics))
	for _, metric := range metrics {
		metric = normalizeAetherRoutingMetric(metric)
		if metric.MetricHash == "" {
			metric.MetricHash = ComputeAetherRoutingMetricHash(metric)
		}
		if err := metric.Validate(); err != nil {
			return nil, err
		}
		out[routingNodeKey(metric.ZoneID, metric.ShardID)] = metric
	}
	return out, nil
}

func routingTableHasZone(table aetracoretypes.RoutingTableCommitment, zoneID zonestypes.ZoneID) bool {
	for _, entry := range table.Entries {
		if string(entry.ZoneID) == string(zoneID) {
			return true
		}
	}
	return false
}

func routingNodeKey(zoneID zonestypes.ZoneID, shardID string) string {
	return string(zoneID) + "/" + shardID
}

func normalizeAetherRoutingMetric(metric AetherRoutingMetric) AetherRoutingMetric {
	metric.ShardID = strings.TrimSpace(metric.ShardID)
	if metric.QueueBacklog == 0 {
		metric.QueueBacklog = metric.OutboxBacklog + metric.InboxBacklog
	}
	if metric.CongestionScore == 0 {
		metric.CongestionScore = metric.AverageExecutionDelay + metric.FailedDeliveryRate + metric.ShardGasUtilization + metric.MessageExpiryRate
	}
	metric.MetricHash = strings.ToLower(strings.TrimSpace(metric.MetricHash))
	return metric
}

func subtractCostFloor(total uint64, credit uint64, floor uint64) uint64 {
	if total <= floor {
		return total
	}
	available := total - floor
	if credit >= available {
		return floor
	}
	return total - credit
}

func normalizeAetherLifecycleRecord(record AetherMessageLifecycleRecord) AetherMessageLifecycleRecord {
	record.MsgID = strings.ToLower(strings.TrimSpace(record.MsgID))
	record.RouteCommitment = strings.ToLower(strings.TrimSpace(record.RouteCommitment))
	record.ReceiptHash = strings.ToLower(strings.TrimSpace(record.ReceiptHash))
	record.RecordHash = strings.ToLower(strings.TrimSpace(record.RecordHash))
	return record
}

func normalizeAetherRoutingEdges(edges []AetherRoutingEdge) []AetherRoutingEdge {
	out := append([]AetherRoutingEdge(nil), edges...)
	sort.SliceStable(out, func(i, j int) bool {
		left := routingNodeKey(out[i].FromZoneID, out[i].FromShardID) + "/" + routingNodeKey(out[i].ToZoneID, out[i].ToShardID)
		right := routingNodeKey(out[j].FromZoneID, out[j].FromShardID) + "/" + routingNodeKey(out[j].ToZoneID, out[j].ToShardID)
		return left < right
	})
	return out
}

func sortAetherRouteCandidates(candidates []AetherRouteCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].TotalCost != candidates[j].TotalCost {
			return candidates[i].TotalCost < candidates[j].TotalCost
		}
		if candidates[i].HopCount != candidates[j].HopCount {
			return candidates[i].HopCount < candidates[j].HopCount
		}
		if candidates[i].PathCommitment != candidates[j].PathCommitment {
			return candidates[i].PathCommitment < candidates[j].PathCommitment
		}
		return candidates[i].DestinationShardID < candidates[j].DestinationShardID
	})
}

func sortAetherLifecycleRecords(records []AetherMessageLifecycleRecord) {
	order := map[MessageLifecycleStage]int{
		MessageLifecycleCreated:			0,
		MessageLifecycleQueuedInSourceOutbox:		1,
		MessageLifecycleCommittedInMessageRoot:		2,
		MessageLifecycleEligibleForDelivery:		3,
		MessageLifecycleQueuedInDestinationInbox:	4,
		MessageLifecycleExecutedOrFailed:		5,
		MessageLifecycleReceipt:			6,
		MessageLifecycleBounceOrFinalize:		7,
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].MsgID != records[j].MsgID {
			return records[i].MsgID < records[j].MsgID
		}
		return order[records[i].Stage] < order[records[j].Stage]
	})
}

func cloneAetherRoutingHops(hops []AetherRoutingHop) []AetherRoutingHop {
	out := make([]AetherRoutingHop, len(hops))
	copy(out, hops)
	return out
}

func cloneAetherLifecycleRecords(records []AetherMessageLifecycleRecord) []AetherMessageLifecycleRecord {
	out := make([]AetherMessageLifecycleRecord, len(records))
	copy(out, records)
	return out
}
