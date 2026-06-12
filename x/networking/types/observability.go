package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingObservableMetric string

const (
	ObservableMetricActivePeers			NetworkingObservableMetric	= "active_peers"
	ObservableMetricPeersByRole			NetworkingObservableMetric	= "peers_by_role"
	ObservableMetricActiveSessions			NetworkingObservableMetric	= "active_sessions"
	ObservableMetricStreamsByChannelType		NetworkingObservableMetric	= "streams_by_channel_type"
	ObservableMetricPerChannelBandwidth		NetworkingObservableMetric	= "per_channel_bandwidth"
	ObservableMetricPeerScore			NetworkingObservableMetric	= "peer_score"
	ObservableMetricOverlaySize			NetworkingObservableMetric	= "overlay_size"
	ObservableMetricOverlayChurn			NetworkingObservableMetric	= "overlay_churn"
	ObservableMetricDiscoveryQueryLatency		NetworkingObservableMetric	= "discovery_query_latency"
	ObservableMetricBroadcastDedupHitRate		NetworkingObservableMetric	= "broadcast_dedup_hit_rate"
	ObservableMetricRL2TransferThroughput		NetworkingObservableMetric	= "rl2_transfer_throughput"
	ObservableMetricRL2ChunkRetryRate		NetworkingObservableMetric	= "rl2_chunk_retry_rate"
	ObservableMetricBlockPropagationLatency		NetworkingObservableMetric	= "block_propagation_latency"
	ObservableMetricCrossZoneMessageDeliveryLatency	NetworkingObservableMetric	= "cross_zone_message_delivery_latency"
	ObservableMetricServiceTrafficVolume		NetworkingObservableMetric	= "service_traffic_volume"
	ObservableMetricRoutingFailureCount		NetworkingObservableMetric	= "routing_failure_count"
)

type NetworkingObservableEvent string

const (
	ObservableEventNetworkNodeRegistered		NetworkingObservableEvent	= "network_node_registered"
	ObservableEventNetworkSessionOpened		NetworkingObservableEvent	= "network_session_opened"
	ObservableEventNetworkSessionClosed		NetworkingObservableEvent	= "network_session_closed"
	ObservableEventNetworkPeerScoreUpdated		NetworkingObservableEvent	= "network_peer_score_updated"
	ObservableEventNetworkOverlayJoined		NetworkingObservableEvent	= "network_overlay_joined"
	ObservableEventNetworkOverlayLeft		NetworkingObservableEvent	= "network_overlay_left"
	ObservableEventNetworkDiscoveryRecordStored	NetworkingObservableEvent	= "network_discovery_record_stored"
	ObservableEventNetworkDiscoveryRecordExpired	NetworkingObservableEvent	= "network_discovery_record_expired"
	ObservableEventNetworkRL2TransferStarted	NetworkingObservableEvent	= "network_rl2_transfer_started"
	ObservableEventNetworkRL2TransferCompleted	NetworkingObservableEvent	= "network_rl2_transfer_completed"
	ObservableEventNetworkInvalidChunk		NetworkingObservableEvent	= "network_invalid_chunk"
	ObservableEventNetworkBroadcastConflict		NetworkingObservableEvent	= "network_broadcast_conflict"
	ObservableEventNetworkRouteFailed		NetworkingObservableEvent	= "network_route_failed"
)

type NetworkingObservableAlert string

const (
	ObservableAlertConsensusChannelLatencyAboveThreshold	NetworkingObservableAlert	= "consensus_channel_latency_above_threshold"
	ObservableAlertBlockPropagationLatencySpike		NetworkingObservableAlert	= "block_propagation_latency_spike"
	ObservableAlertPeerScoreCollapse			NetworkingObservableAlert	= "peer_score_collapse"
	ObservableAlertOverlayPartitionSuspected		NetworkingObservableAlert	= "overlay_partition_suspected"
	ObservableAlertDiscoveryPoisoningAttempt		NetworkingObservableAlert	= "discovery_poisoning_attempt"
	ObservableAlertRL2InvalidChunkSpike			NetworkingObservableAlert	= "rl2_invalid_chunk_spike"
	ObservableAlertServiceTrafficExceedingQuota		NetworkingObservableAlert	= "service_traffic_exceeding_quota"
	ObservableAlertCrossZoneMessageDeliveryBacklog		NetworkingObservableAlert	= "cross_zone_message_delivery_backlog"
	ObservableAlertEclipseRiskPeerDiversityLow		NetworkingObservableAlert	= "eclipse_risk_peer_diversity_low"
)

type NetworkingAlertSeverity string

const (
	NetworkingAlertSeverityWarning	NetworkingAlertSeverity	= "warning"
	NetworkingAlertSeverityCritical	NetworkingAlertSeverity	= "critical"
)

type NetworkingAlertCondition string

const (
	NetworkingAlertConditionAboveThreshold	NetworkingAlertCondition	= "above_threshold"
	NetworkingAlertConditionBelowThreshold	NetworkingAlertCondition	= "below_threshold"
)

type NetworkingMetricSample struct {
	Metric	NetworkingObservableMetric
	Labels	[]string
	Value	uint64
	Height	uint64
}

type NetworkingEventRecord struct {
	Event		NetworkingObservableEvent
	NodeID		string
	OverlayID	string
	Channel		ChannelClass
	TransferID	string
	MessageID	string
	EvidenceHash	string
	Height		uint64
	EventID		string
}

type NetworkingObservabilitySpec struct {
	Metrics		[]NetworkingObservableMetric
	Events		[]NetworkingObservableEvent
	Alerts		[]NetworkingObservableAlert
	SpecRoot	string
}

type NetworkingObservabilityReport struct {
	Spec		NetworkingObservabilitySpec
	Metrics		[]NetworkingMetricSample
	Events		[]NetworkingEventRecord
	MissingMetrics	[]NetworkingObservableMetric
	MissingEvents	[]NetworkingObservableEvent
	Ready		bool
	ReportHash	string
}

type NetworkingAlertRule struct {
	Alert		NetworkingObservableAlert
	Severity	NetworkingAlertSeverity
	Condition	NetworkingAlertCondition
	SourceMetrics	[]NetworkingObservableMetric
	SourceEvents	[]NetworkingObservableEvent
	Threshold	uint64
	WindowBlocks	uint64
	Description	string
}

type NetworkingAlertSignal struct {
	Alert		NetworkingObservableAlert
	Severity	NetworkingAlertSeverity
	Condition	NetworkingAlertCondition
	SourceMetric	NetworkingObservableMetric
	SourceEvent	NetworkingObservableEvent
	NodeID		string
	OverlayID	string
	Observed	uint64
	Threshold	uint64
	WindowBlocks	uint64
	Height		uint64
	TriggerID	string
}

type NetworkingAlertReport struct {
	Rules		[]NetworkingAlertRule
	Signals		[]NetworkingAlertSignal
	MissingAlerts	[]NetworkingObservableAlert
	Ready		bool
	ReportHash	string
}

func DefaultNetworkingObservabilitySpec() NetworkingObservabilitySpec {
	spec := NetworkingObservabilitySpec{
		Metrics: []NetworkingObservableMetric{
			ObservableMetricActivePeers,
			ObservableMetricPeersByRole,
			ObservableMetricActiveSessions,
			ObservableMetricStreamsByChannelType,
			ObservableMetricPerChannelBandwidth,
			ObservableMetricPeerScore,
			ObservableMetricOverlaySize,
			ObservableMetricOverlayChurn,
			ObservableMetricDiscoveryQueryLatency,
			ObservableMetricBroadcastDedupHitRate,
			ObservableMetricRL2TransferThroughput,
			ObservableMetricRL2ChunkRetryRate,
			ObservableMetricBlockPropagationLatency,
			ObservableMetricCrossZoneMessageDeliveryLatency,
			ObservableMetricServiceTrafficVolume,
			ObservableMetricRoutingFailureCount,
		},
		Events: []NetworkingObservableEvent{
			ObservableEventNetworkNodeRegistered,
			ObservableEventNetworkSessionOpened,
			ObservableEventNetworkSessionClosed,
			ObservableEventNetworkPeerScoreUpdated,
			ObservableEventNetworkOverlayJoined,
			ObservableEventNetworkOverlayLeft,
			ObservableEventNetworkDiscoveryRecordStored,
			ObservableEventNetworkDiscoveryRecordExpired,
			ObservableEventNetworkRL2TransferStarted,
			ObservableEventNetworkRL2TransferCompleted,
			ObservableEventNetworkInvalidChunk,
			ObservableEventNetworkBroadcastConflict,
			ObservableEventNetworkRouteFailed,
		},
		Alerts: []NetworkingObservableAlert{
			ObservableAlertConsensusChannelLatencyAboveThreshold,
			ObservableAlertBlockPropagationLatencySpike,
			ObservableAlertPeerScoreCollapse,
			ObservableAlertOverlayPartitionSuspected,
			ObservableAlertDiscoveryPoisoningAttempt,
			ObservableAlertRL2InvalidChunkSpike,
			ObservableAlertServiceTrafficExceedingQuota,
			ObservableAlertCrossZoneMessageDeliveryBacklog,
			ObservableAlertEclipseRiskPeerDiversityLow,
		},
	}
	spec = NormalizeNetworkingObservabilitySpec(spec)
	spec.SpecRoot = ComputeNetworkingObservabilitySpecRoot(spec)
	return spec
}

func ValidateNetworkingObservabilitySpec(spec NetworkingObservabilitySpec) error {
	normalized := NormalizeNetworkingObservabilitySpec(spec)
	required := DefaultNetworkingObservabilitySpec()
	if len(normalized.Metrics) != len(required.Metrics) {
		return fmt.Errorf("networking observability spec must define %d metrics", len(required.Metrics))
	}
	if len(normalized.Events) != len(required.Events) {
		return fmt.Errorf("networking observability spec must define %d events", len(required.Events))
	}
	if len(normalized.Alerts) != len(required.Alerts) {
		return fmt.Errorf("networking observability spec must define %d alerts", len(required.Alerts))
	}
	seenMetrics := make(map[NetworkingObservableMetric]struct{}, len(normalized.Metrics))
	for _, metric := range normalized.Metrics {
		if !IsNetworkingObservableMetric(metric) {
			return fmt.Errorf("unknown networking observability metric %q", metric)
		}
		if _, found := seenMetrics[metric]; found {
			return errors.New("networking observability duplicate metric")
		}
		seenMetrics[metric] = struct{}{}
	}
	for _, metric := range required.Metrics {
		if _, found := seenMetrics[metric]; !found {
			return fmt.Errorf("networking observability missing metric %s", metric)
		}
	}
	seenEvents := make(map[NetworkingObservableEvent]struct{}, len(normalized.Events))
	for _, event := range normalized.Events {
		if !IsNetworkingObservableEvent(event) {
			return fmt.Errorf("unknown networking observability event %q", event)
		}
		if _, found := seenEvents[event]; found {
			return errors.New("networking observability duplicate event")
		}
		seenEvents[event] = struct{}{}
	}
	for _, event := range required.Events {
		if _, found := seenEvents[event]; !found {
			return fmt.Errorf("networking observability missing event %s", event)
		}
	}
	seenAlerts := make(map[NetworkingObservableAlert]struct{}, len(normalized.Alerts))
	for _, alert := range normalized.Alerts {
		if !IsNetworkingObservableAlert(alert) {
			return fmt.Errorf("unknown networking observability alert %q", alert)
		}
		if _, found := seenAlerts[alert]; found {
			return errors.New("networking observability duplicate alert")
		}
		seenAlerts[alert] = struct{}{}
	}
	for _, alert := range required.Alerts {
		if _, found := seenAlerts[alert]; !found {
			return fmt.Errorf("networking observability missing alert %s", alert)
		}
	}
	if normalized.SpecRoot == "" {
		return errors.New("networking observability spec root is required")
	}
	if normalized.SpecRoot != ComputeNetworkingObservabilitySpecRoot(normalized) {
		return errors.New("networking observability spec root mismatch")
	}
	return nil
}

func BuildNetworkingObservabilityReport(spec NetworkingObservabilitySpec, metrics []NetworkingMetricSample, events []NetworkingEventRecord) (NetworkingObservabilityReport, error) {
	spec = NormalizeNetworkingObservabilitySpec(spec)
	if err := ValidateNetworkingObservabilitySpec(spec); err != nil {
		return NetworkingObservabilityReport{}, err
	}
	normalizedMetrics := NormalizeNetworkingMetricSamples(metrics)
	normalizedEvents := NormalizeNetworkingEventRecords(events)
	report := NetworkingObservabilityReport{
		Spec:		spec,
		Metrics:	normalizedMetrics,
		Events:		normalizedEvents,
	}
	coveredMetrics := make(map[NetworkingObservableMetric]struct{}, len(normalizedMetrics))
	for _, sample := range normalizedMetrics {
		if err := sample.Validate(); err != nil {
			return NetworkingObservabilityReport{}, err
		}
		coveredMetrics[sample.Metric] = struct{}{}
	}
	for _, metric := range spec.Metrics {
		if _, found := coveredMetrics[metric]; !found {
			report.MissingMetrics = append(report.MissingMetrics, metric)
		}
	}
	coveredEvents := make(map[NetworkingObservableEvent]struct{}, len(normalizedEvents))
	for _, event := range normalizedEvents {
		if err := event.Validate(); err != nil {
			return NetworkingObservabilityReport{}, err
		}
		coveredEvents[event.Event] = struct{}{}
	}
	for _, event := range spec.Events {
		if _, found := coveredEvents[event]; !found {
			report.MissingEvents = append(report.MissingEvents, event)
		}
	}
	sortObservableMetrics(report.MissingMetrics)
	sortObservableEvents(report.MissingEvents)
	report.Ready = len(report.MissingMetrics) == 0 && len(report.MissingEvents) == 0
	report.ReportHash = ComputeNetworkingObservabilityReportHash(report)
	return report, nil
}

func DefaultNetworkingAlertRules() []NetworkingAlertRule {
	return []NetworkingAlertRule{
		{
			Alert:		ObservableAlertConsensusChannelLatencyAboveThreshold,
			Severity:	NetworkingAlertSeverityCritical,
			Condition:	NetworkingAlertConditionAboveThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricBlockPropagationLatency},
			Threshold:	250,
			WindowBlocks:	3,
			Description:	"Consensus channel latency above threshold",
		},
		{
			Alert:		ObservableAlertBlockPropagationLatencySpike,
			Severity:	NetworkingAlertSeverityWarning,
			Condition:	NetworkingAlertConditionAboveThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricBlockPropagationLatency},
			Threshold:	500,
			WindowBlocks:	5,
			Description:	"Block propagation latency spike",
		},
		{
			Alert:		ObservableAlertPeerScoreCollapse,
			Severity:	NetworkingAlertSeverityCritical,
			Condition:	NetworkingAlertConditionBelowThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricPeerScore},
			Threshold:	2_500,
			WindowBlocks:	10,
			Description:	"Peer score collapse",
		},
		{
			Alert:		ObservableAlertOverlayPartitionSuspected,
			Severity:	NetworkingAlertSeverityCritical,
			Condition:	NetworkingAlertConditionAboveThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricRoutingFailureCount, ObservableMetricOverlayChurn},
			SourceEvents:	[]NetworkingObservableEvent{ObservableEventNetworkRouteFailed},
			Threshold:	3,
			WindowBlocks:	6,
			Description:	"Overlay partition suspected",
		},
		{
			Alert:		ObservableAlertDiscoveryPoisoningAttempt,
			Severity:	NetworkingAlertSeverityCritical,
			Condition:	NetworkingAlertConditionAboveThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricDiscoveryQueryLatency},
			SourceEvents:	[]NetworkingObservableEvent{ObservableEventNetworkDiscoveryRecordStored, ObservableEventNetworkDiscoveryRecordExpired},
			Threshold:	1,
			WindowBlocks:	10,
			Description:	"Discovery poisoning attempt",
		},
		{
			Alert:		ObservableAlertRL2InvalidChunkSpike,
			Severity:	NetworkingAlertSeverityCritical,
			Condition:	NetworkingAlertConditionAboveThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricRL2ChunkRetryRate},
			SourceEvents:	[]NetworkingObservableEvent{ObservableEventNetworkInvalidChunk},
			Threshold:	2,
			WindowBlocks:	4,
			Description:	"RL2 invalid chunk spike",
		},
		{
			Alert:		ObservableAlertServiceTrafficExceedingQuota,
			Severity:	NetworkingAlertSeverityWarning,
			Condition:	NetworkingAlertConditionAboveThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricServiceTrafficVolume, ObservableMetricPerChannelBandwidth},
			Threshold:	10 << 20,
			WindowBlocks:	5,
			Description:	"Service traffic exceeding quota",
		},
		{
			Alert:		ObservableAlertCrossZoneMessageDeliveryBacklog,
			Severity:	NetworkingAlertSeverityWarning,
			Condition:	NetworkingAlertConditionAboveThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricCrossZoneMessageDeliveryLatency},
			Threshold:	1_000,
			WindowBlocks:	8,
			Description:	"Cross-zone message delivery backlog",
		},
		{
			Alert:		ObservableAlertEclipseRiskPeerDiversityLow,
			Severity:	NetworkingAlertSeverityCritical,
			Condition:	NetworkingAlertConditionBelowThreshold,
			SourceMetrics:	[]NetworkingObservableMetric{ObservableMetricActivePeers, ObservableMetricPeersByRole, ObservableMetricOverlaySize},
			Threshold:	4,
			WindowBlocks:	20,
			Description:	"Eclipse risk peer diversity low",
		},
	}
}

func ValidateNetworkingAlertRules(rules []NetworkingAlertRule) error {
	normalized := NormalizeNetworkingAlertRules(rules)
	required := DefaultNetworkingObservabilitySpec().Alerts
	if len(normalized) != len(required) {
		return fmt.Errorf("networking observability alert rules must define %d alerts", len(required))
	}
	seen := make(map[NetworkingObservableAlert]NetworkingAlertRule, len(normalized))
	for _, rule := range normalized {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.Alert]; found {
			return errors.New("networking observability duplicate alert rule")
		}
		seen[rule.Alert] = rule
	}
	for _, alert := range required {
		if _, found := seen[alert]; !found {
			return fmt.Errorf("networking observability missing alert rule %s", alert)
		}
	}
	return nil
}

func BuildNetworkingAlertReport(rules []NetworkingAlertRule, signals []NetworkingAlertSignal) (NetworkingAlertReport, error) {
	normalizedRules := NormalizeNetworkingAlertRules(rules)
	if err := ValidateNetworkingAlertRules(normalizedRules); err != nil {
		return NetworkingAlertReport{}, err
	}
	normalizedSignals := NormalizeNetworkingAlertSignals(signals)
	report := NetworkingAlertReport{
		Rules:		normalizedRules,
		Signals:	normalizedSignals,
	}
	signaled := make(map[NetworkingObservableAlert]struct{}, len(normalizedSignals))
	for _, signal := range normalizedSignals {
		if err := signal.Validate(); err != nil {
			return NetworkingAlertReport{}, err
		}
		signaled[signal.Alert] = struct{}{}
	}
	for _, rule := range normalizedRules {
		if _, found := signaled[rule.Alert]; !found {
			report.MissingAlerts = append(report.MissingAlerts, rule.Alert)
		}
	}
	sortObservableAlerts(report.MissingAlerts)
	report.Ready = len(report.MissingAlerts) == 0
	report.ReportHash = ComputeNetworkingAlertReportHash(report)
	return report, nil
}

func NewNetworkingEventRecord(event NetworkingObservableEvent, nodeID, overlayID string, channel ChannelClass, transferID, messageID, evidenceHash string, height uint64) NetworkingEventRecord {
	record := NetworkingEventRecord{
		Event:		event,
		NodeID:		nodeID,
		OverlayID:	overlayID,
		Channel:	channel,
		TransferID:	transferID,
		MessageID:	messageID,
		EvidenceHash:	evidenceHash,
		Height:		height,
	}
	record = NormalizeNetworkingEventRecord(record)
	record.EventID = ComputeNetworkingEventID(record)
	return record
}

func NewNetworkingAlertSignal(rule NetworkingAlertRule, sourceMetric NetworkingObservableMetric, sourceEvent NetworkingObservableEvent, nodeID, overlayID string, observed uint64, height uint64) NetworkingAlertSignal {
	rule = NormalizeNetworkingAlertRule(rule)
	signal := NetworkingAlertSignal{
		Alert:		rule.Alert,
		Severity:	rule.Severity,
		Condition:	rule.Condition,
		SourceMetric:	sourceMetric,
		SourceEvent:	sourceEvent,
		NodeID:		nodeID,
		OverlayID:	overlayID,
		Observed:	observed,
		Threshold:	rule.Threshold,
		WindowBlocks:	rule.WindowBlocks,
		Height:		height,
	}
	signal = NormalizeNetworkingAlertSignal(signal)
	signal.TriggerID = ComputeNetworkingAlertTriggerID(signal)
	return signal
}

func (sample NetworkingMetricSample) Validate() error {
	sample = NormalizeNetworkingMetricSample(sample)
	if !IsNetworkingObservableMetric(sample.Metric) {
		return fmt.Errorf("unknown networking observability metric %q", sample.Metric)
	}
	if sample.Height == 0 {
		return errors.New("networking observability metric height must be positive")
	}
	return nil
}

func (record NetworkingEventRecord) Validate() error {
	record = NormalizeNetworkingEventRecord(record)
	if !IsNetworkingObservableEvent(record.Event) {
		return fmt.Errorf("unknown networking observability event %q", record.Event)
	}
	if record.Height == 0 {
		return errors.New("networking observability event height must be positive")
	}
	if record.NodeID != "" {
		if err := ValidateHash("networking observability event node id", record.NodeID); err != nil {
			return err
		}
	}
	if record.OverlayID != "" {
		if err := ValidateHash("networking observability event overlay id", record.OverlayID); err != nil {
			return err
		}
	}
	if record.Channel != "" && !IsChannelClass(record.Channel) {
		return fmt.Errorf("unknown networking observability event channel %q", record.Channel)
	}
	if record.TransferID != "" {
		if err := ValidateHash("networking observability event transfer id", record.TransferID); err != nil {
			return err
		}
	}
	if record.MessageID != "" {
		if err := ValidateHash("networking observability event message id", record.MessageID); err != nil {
			return err
		}
	}
	if record.EvidenceHash != "" {
		if err := ValidateHash("networking observability event evidence hash", record.EvidenceHash); err != nil {
			return err
		}
	}
	if err := ValidateHash("networking observability event id", record.EventID); err != nil {
		return err
	}
	if record.EventID != ComputeNetworkingEventID(record) {
		return errors.New("networking observability event id mismatch")
	}
	return nil
}

func (rule NetworkingAlertRule) Validate() error {
	rule = NormalizeNetworkingAlertRule(rule)
	if !IsNetworkingObservableAlert(rule.Alert) {
		return fmt.Errorf("unknown networking observability alert %q", rule.Alert)
	}
	if !IsNetworkingAlertSeverity(rule.Severity) {
		return fmt.Errorf("unknown networking observability alert severity %q", rule.Severity)
	}
	if !IsNetworkingAlertCondition(rule.Condition) {
		return fmt.Errorf("unknown networking observability alert condition %q", rule.Condition)
	}
	if rule.Threshold == 0 {
		return errors.New("networking observability alert threshold must be positive")
	}
	if rule.WindowBlocks == 0 {
		return errors.New("networking observability alert window blocks must be positive")
	}
	if len(rule.SourceMetrics) == 0 && len(rule.SourceEvents) == 0 {
		return errors.New("networking observability alert source is required")
	}
	for _, metric := range rule.SourceMetrics {
		if !IsNetworkingObservableMetric(metric) {
			return fmt.Errorf("unknown networking observability alert source metric %q", metric)
		}
	}
	for _, event := range rule.SourceEvents {
		if !IsNetworkingObservableEvent(event) {
			return fmt.Errorf("unknown networking observability alert source event %q", event)
		}
	}
	if strings.TrimSpace(rule.Description) == "" {
		return errors.New("networking observability alert description is required")
	}
	return nil
}

func (signal NetworkingAlertSignal) Validate() error {
	signal = NormalizeNetworkingAlertSignal(signal)
	if !IsNetworkingObservableAlert(signal.Alert) {
		return fmt.Errorf("unknown networking observability alert signal %q", signal.Alert)
	}
	if !IsNetworkingAlertSeverity(signal.Severity) {
		return fmt.Errorf("unknown networking observability alert signal severity %q", signal.Severity)
	}
	if !IsNetworkingAlertCondition(signal.Condition) {
		return fmt.Errorf("unknown networking observability alert signal condition %q", signal.Condition)
	}
	if signal.SourceMetric != "" && !IsNetworkingObservableMetric(signal.SourceMetric) {
		return fmt.Errorf("unknown networking observability alert signal source metric %q", signal.SourceMetric)
	}
	if signal.SourceEvent != "" && !IsNetworkingObservableEvent(signal.SourceEvent) {
		return fmt.Errorf("unknown networking observability alert signal source event %q", signal.SourceEvent)
	}
	if signal.SourceMetric == "" && signal.SourceEvent == "" {
		return errors.New("networking observability alert signal source is required")
	}
	if signal.Height == 0 {
		return errors.New("networking observability alert signal height must be positive")
	}
	if signal.Threshold == 0 {
		return errors.New("networking observability alert signal threshold must be positive")
	}
	if signal.WindowBlocks == 0 {
		return errors.New("networking observability alert signal window blocks must be positive")
	}
	if signal.Condition == NetworkingAlertConditionAboveThreshold && signal.Observed < signal.Threshold {
		return errors.New("networking observability alert signal observed value is below threshold")
	}
	if signal.Condition == NetworkingAlertConditionBelowThreshold && signal.Observed > signal.Threshold {
		return errors.New("networking observability alert signal observed value is above threshold")
	}
	if signal.NodeID != "" {
		if err := ValidateHash("networking observability alert signal node id", signal.NodeID); err != nil {
			return err
		}
	}
	if signal.OverlayID != "" {
		if err := ValidateHash("networking observability alert signal overlay id", signal.OverlayID); err != nil {
			return err
		}
	}
	if err := ValidateHash("networking observability alert trigger id", signal.TriggerID); err != nil {
		return err
	}
	if signal.TriggerID != ComputeNetworkingAlertTriggerID(signal) {
		return errors.New("networking observability alert trigger id mismatch")
	}
	return nil
}

func ComputeNetworkingObservabilitySpecRoot(spec NetworkingObservabilitySpec) string {
	spec = NormalizeNetworkingObservabilitySpec(spec)
	parts := []string{"networking-observability-spec"}
	for _, metric := range spec.Metrics {
		parts = append(parts, "metric", string(metric))
	}
	for _, event := range spec.Events {
		parts = append(parts, "event", string(event))
	}
	for _, alert := range spec.Alerts {
		parts = append(parts, "alert", string(alert))
	}
	return HashParts(parts...)
}

func ComputeNetworkingObservabilityReportHash(report NetworkingObservabilityReport) string {
	parts := []string{"networking-observability-report", report.Spec.SpecRoot, fmt.Sprintf("%t", report.Ready)}
	for _, sample := range NormalizeNetworkingMetricSamples(report.Metrics) {
		parts = append(parts, "metric", string(sample.Metric), fmt.Sprintf("%d", sample.Value), fmt.Sprintf("%d", sample.Height))
		parts = append(parts, sample.Labels...)
	}
	for _, event := range NormalizeNetworkingEventRecords(report.Events) {
		parts = append(parts, "event", string(event.Event), event.EventID, event.NodeID, event.OverlayID, string(event.Channel), event.TransferID, event.MessageID, event.EvidenceHash, fmt.Sprintf("%d", event.Height))
	}
	for _, metric := range report.MissingMetrics {
		parts = append(parts, "missing_metric", string(metric))
	}
	for _, event := range report.MissingEvents {
		parts = append(parts, "missing_event", string(event))
	}
	return HashParts(parts...)
}

func ComputeNetworkingAlertReportHash(report NetworkingAlertReport) string {
	parts := []string{"networking-alert-report", fmt.Sprintf("%t", report.Ready)}
	for _, rule := range NormalizeNetworkingAlertRules(report.Rules) {
		parts = append(parts, "rule", string(rule.Alert), string(rule.Severity), string(rule.Condition), fmt.Sprintf("%d", rule.Threshold), fmt.Sprintf("%d", rule.WindowBlocks), rule.Description)
		for _, metric := range rule.SourceMetrics {
			parts = append(parts, "metric", string(metric))
		}
		for _, event := range rule.SourceEvents {
			parts = append(parts, "event", string(event))
		}
	}
	for _, signal := range NormalizeNetworkingAlertSignals(report.Signals) {
		parts = append(parts, "signal", string(signal.Alert), string(signal.Severity), string(signal.Condition), string(signal.SourceMetric), string(signal.SourceEvent), signal.NodeID, signal.OverlayID, fmt.Sprintf("%d", signal.Observed), fmt.Sprintf("%d", signal.Threshold), fmt.Sprintf("%d", signal.WindowBlocks), fmt.Sprintf("%d", signal.Height), signal.TriggerID)
	}
	for _, alert := range report.MissingAlerts {
		parts = append(parts, "missing_alert", string(alert))
	}
	return HashParts(parts...)
}

func ComputeNetworkingEventID(record NetworkingEventRecord) string {
	record = NormalizeNetworkingEventRecord(record)
	return HashParts(
		"networking-observability-event",
		string(record.Event),
		record.NodeID,
		record.OverlayID,
		string(record.Channel),
		record.TransferID,
		record.MessageID,
		record.EvidenceHash,
		fmt.Sprintf("%d", record.Height),
	)
}

func ComputeNetworkingAlertTriggerID(signal NetworkingAlertSignal) string {
	signal = NormalizeNetworkingAlertSignal(signal)
	return HashParts(
		"networking-observability-alert",
		string(signal.Alert),
		string(signal.Severity),
		string(signal.Condition),
		string(signal.SourceMetric),
		string(signal.SourceEvent),
		signal.NodeID,
		signal.OverlayID,
		fmt.Sprintf("%d", signal.Observed),
		fmt.Sprintf("%d", signal.Threshold),
		fmt.Sprintf("%d", signal.WindowBlocks),
		fmt.Sprintf("%d", signal.Height),
	)
}

func IsNetworkingObservableMetric(metric NetworkingObservableMetric) bool {
	switch metric {
	case ObservableMetricActivePeers,
		ObservableMetricPeersByRole,
		ObservableMetricActiveSessions,
		ObservableMetricStreamsByChannelType,
		ObservableMetricPerChannelBandwidth,
		ObservableMetricPeerScore,
		ObservableMetricOverlaySize,
		ObservableMetricOverlayChurn,
		ObservableMetricDiscoveryQueryLatency,
		ObservableMetricBroadcastDedupHitRate,
		ObservableMetricRL2TransferThroughput,
		ObservableMetricRL2ChunkRetryRate,
		ObservableMetricBlockPropagationLatency,
		ObservableMetricCrossZoneMessageDeliveryLatency,
		ObservableMetricServiceTrafficVolume,
		ObservableMetricRoutingFailureCount:
		return true
	default:
		return false
	}
}

func IsNetworkingObservableEvent(event NetworkingObservableEvent) bool {
	switch event {
	case ObservableEventNetworkNodeRegistered,
		ObservableEventNetworkSessionOpened,
		ObservableEventNetworkSessionClosed,
		ObservableEventNetworkPeerScoreUpdated,
		ObservableEventNetworkOverlayJoined,
		ObservableEventNetworkOverlayLeft,
		ObservableEventNetworkDiscoveryRecordStored,
		ObservableEventNetworkDiscoveryRecordExpired,
		ObservableEventNetworkRL2TransferStarted,
		ObservableEventNetworkRL2TransferCompleted,
		ObservableEventNetworkInvalidChunk,
		ObservableEventNetworkBroadcastConflict,
		ObservableEventNetworkRouteFailed:
		return true
	default:
		return false
	}
}

func IsNetworkingObservableAlert(alert NetworkingObservableAlert) bool {
	switch alert {
	case ObservableAlertConsensusChannelLatencyAboveThreshold,
		ObservableAlertBlockPropagationLatencySpike,
		ObservableAlertPeerScoreCollapse,
		ObservableAlertOverlayPartitionSuspected,
		ObservableAlertDiscoveryPoisoningAttempt,
		ObservableAlertRL2InvalidChunkSpike,
		ObservableAlertServiceTrafficExceedingQuota,
		ObservableAlertCrossZoneMessageDeliveryBacklog,
		ObservableAlertEclipseRiskPeerDiversityLow:
		return true
	default:
		return false
	}
}

func IsNetworkingAlertSeverity(severity NetworkingAlertSeverity) bool {
	switch severity {
	case NetworkingAlertSeverityWarning, NetworkingAlertSeverityCritical:
		return true
	default:
		return false
	}
}

func IsNetworkingAlertCondition(condition NetworkingAlertCondition) bool {
	switch condition {
	case NetworkingAlertConditionAboveThreshold, NetworkingAlertConditionBelowThreshold:
		return true
	default:
		return false
	}
}

func NormalizeNetworkingObservabilitySpec(spec NetworkingObservabilitySpec) NetworkingObservabilitySpec {
	spec.Metrics = normalizeObservableMetrics(spec.Metrics)
	spec.Events = normalizeObservableEvents(spec.Events)
	spec.Alerts = normalizeObservableAlerts(spec.Alerts)
	spec.SpecRoot = normalizeHashText(spec.SpecRoot)
	return spec
}

func NormalizeNetworkingMetricSamples(samples []NetworkingMetricSample) []NetworkingMetricSample {
	out := make([]NetworkingMetricSample, 0, len(samples))
	for _, sample := range samples {
		sample = NormalizeNetworkingMetricSample(sample)
		if sample.Metric == "" {
			continue
		}
		out = append(out, sample)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Metric != out[j].Metric {
			return out[i].Metric < out[j].Metric
		}
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return strings.Join(out[i].Labels, "\x00") < strings.Join(out[j].Labels, "\x00")
	})
	return out
}

func NormalizeNetworkingMetricSample(sample NetworkingMetricSample) NetworkingMetricSample {
	sample.Metric = NetworkingObservableMetric(strings.ToLower(strings.TrimSpace(string(sample.Metric))))
	labels := make([]string, 0, len(sample.Labels))
	seen := make(map[string]struct{}, len(sample.Labels))
	for _, label := range sample.Labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		if _, found := seen[label]; found {
			continue
		}
		seen[label] = struct{}{}
		labels = append(labels, label)
	}
	sort.Strings(labels)
	sample.Labels = labels
	return sample
}

func NormalizeNetworkingAlertRules(rules []NetworkingAlertRule) []NetworkingAlertRule {
	out := make([]NetworkingAlertRule, 0, len(rules))
	for _, rule := range rules {
		rule = NormalizeNetworkingAlertRule(rule)
		if rule.Alert == "" {
			continue
		}
		out = append(out, rule)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Alert < out[j].Alert
	})
	return out
}

func NormalizeNetworkingAlertRule(rule NetworkingAlertRule) NetworkingAlertRule {
	rule.Alert = NetworkingObservableAlert(strings.ToLower(strings.TrimSpace(string(rule.Alert))))
	rule.Severity = NetworkingAlertSeverity(strings.ToLower(strings.TrimSpace(string(rule.Severity))))
	rule.Condition = NetworkingAlertCondition(strings.ToLower(strings.TrimSpace(string(rule.Condition))))
	rule.SourceMetrics = normalizeObservableMetrics(rule.SourceMetrics)
	rule.SourceEvents = normalizeObservableEvents(rule.SourceEvents)
	rule.Description = strings.TrimSpace(rule.Description)
	return rule
}

func NormalizeNetworkingAlertSignals(signals []NetworkingAlertSignal) []NetworkingAlertSignal {
	out := make([]NetworkingAlertSignal, 0, len(signals))
	for _, signal := range signals {
		signal = NormalizeNetworkingAlertSignal(signal)
		if signal.Alert == "" {
			continue
		}
		out = append(out, signal)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Alert != out[j].Alert {
			return out[i].Alert < out[j].Alert
		}
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].TriggerID < out[j].TriggerID
	})
	return out
}

func NormalizeNetworkingAlertSignal(signal NetworkingAlertSignal) NetworkingAlertSignal {
	signal.Alert = NetworkingObservableAlert(strings.ToLower(strings.TrimSpace(string(signal.Alert))))
	signal.Severity = NetworkingAlertSeverity(strings.ToLower(strings.TrimSpace(string(signal.Severity))))
	signal.Condition = NetworkingAlertCondition(strings.ToLower(strings.TrimSpace(string(signal.Condition))))
	signal.SourceMetric = NetworkingObservableMetric(strings.ToLower(strings.TrimSpace(string(signal.SourceMetric))))
	signal.SourceEvent = NetworkingObservableEvent(strings.ToLower(strings.TrimSpace(string(signal.SourceEvent))))
	signal.NodeID = normalizeHashText(signal.NodeID)
	signal.OverlayID = normalizeHashText(signal.OverlayID)
	signal.TriggerID = normalizeHashText(signal.TriggerID)
	return signal
}

func NormalizeNetworkingEventRecords(records []NetworkingEventRecord) []NetworkingEventRecord {
	out := make([]NetworkingEventRecord, 0, len(records))
	for _, record := range records {
		record = NormalizeNetworkingEventRecord(record)
		if record.Event == "" {
			continue
		}
		out = append(out, record)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Event != out[j].Event {
			return out[i].Event < out[j].Event
		}
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].EventID < out[j].EventID
	})
	return out
}

func NormalizeNetworkingEventRecord(record NetworkingEventRecord) NetworkingEventRecord {
	record.Event = NetworkingObservableEvent(strings.ToLower(strings.TrimSpace(string(record.Event))))
	record.NodeID = normalizeHashText(record.NodeID)
	record.OverlayID = normalizeHashText(record.OverlayID)
	record.Channel = ChannelClass(strings.ToUpper(strings.TrimSpace(string(record.Channel))))
	record.TransferID = normalizeHashText(record.TransferID)
	record.MessageID = normalizeHashText(record.MessageID)
	record.EvidenceHash = normalizeHashText(record.EvidenceHash)
	record.EventID = normalizeHashText(record.EventID)
	return record
}

func normalizeObservableMetrics(metrics []NetworkingObservableMetric) []NetworkingObservableMetric {
	out := make([]NetworkingObservableMetric, 0, len(metrics))
	seen := make(map[NetworkingObservableMetric]struct{}, len(metrics))
	for _, metric := range metrics {
		metric = NetworkingObservableMetric(strings.ToLower(strings.TrimSpace(string(metric))))
		if metric == "" {
			continue
		}
		if _, found := seen[metric]; found {
			continue
		}
		seen[metric] = struct{}{}
		out = append(out, metric)
	}
	sortObservableMetrics(out)
	return out
}

func normalizeObservableEvents(events []NetworkingObservableEvent) []NetworkingObservableEvent {
	out := make([]NetworkingObservableEvent, 0, len(events))
	seen := make(map[NetworkingObservableEvent]struct{}, len(events))
	for _, event := range events {
		event = NetworkingObservableEvent(strings.ToLower(strings.TrimSpace(string(event))))
		if event == "" {
			continue
		}
		if _, found := seen[event]; found {
			continue
		}
		seen[event] = struct{}{}
		out = append(out, event)
	}
	sortObservableEvents(out)
	return out
}

func normalizeObservableAlerts(alerts []NetworkingObservableAlert) []NetworkingObservableAlert {
	out := make([]NetworkingObservableAlert, 0, len(alerts))
	seen := make(map[NetworkingObservableAlert]struct{}, len(alerts))
	for _, alert := range alerts {
		alert = NetworkingObservableAlert(strings.ToLower(strings.TrimSpace(string(alert))))
		if alert == "" {
			continue
		}
		if _, found := seen[alert]; found {
			continue
		}
		seen[alert] = struct{}{}
		out = append(out, alert)
	}
	sortObservableAlerts(out)
	return out
}

func sortObservableMetrics(values []NetworkingObservableMetric) {
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
}

func sortObservableEvents(values []NetworkingObservableEvent) {
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
}

func sortObservableAlerts(values []NetworkingObservableAlert) {
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
}
