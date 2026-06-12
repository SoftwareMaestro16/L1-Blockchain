package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	AVMMetricAsyncMessagesSubmitted		AVMObservabilityMetric	= "async_messages_submitted"
	AVMMetricAsyncMessagesExecuted		AVMObservabilityMetric	= "async_messages_executed"
	AVMMetricAsyncMessagesExpired		AVMObservabilityMetric	= "async_messages_expired"
	AVMMetricAsyncMessagesBounced		AVMObservabilityMetric	= "async_messages_bounced"
	AVMMetricDeadLetterCount		AVMObservabilityMetric	= "dead_letter_count"
	AVMMetricRetryQueueSize			AVMObservabilityMetric	= "retry_queue_size"
	AVMMetricDelayedQueueSize		AVMObservabilityMetric	= "delayed_queue_size"
	AVMMetricActorCount			AVMObservabilityMetric	= "actor_count"
	AVMMetricActorMailboxDepth		AVMObservabilityMetric	= "actor_mailbox_depth"
	AVMMetricContinuationsActive		AVMObservabilityMetric	= "continuations_active"
	AVMMetricContinuationsExpired		AVMObservabilityMetric	= "continuations_expired"
	AVMMetricContractExecutions		AVMObservabilityMetric	= "contract_executions"
	AVMMetricContractFailures		AVMObservabilityMetric	= "contract_failures"
	AVMMetricGasReserved			AVMObservabilityMetric	= "gas_reserved"
	AVMMetricGasConsumed			AVMObservabilityMetric	= "gas_consumed"
	AVMMetricZoneAsyncBudgetUsage		AVMObservabilityMetric	= "zone_async_budget_usage"
	AVMMetricQueueDrainLatency		AVMObservabilityMetric	= "queue_drain_latency"
	AVMMetricReceiptRootGenerationTime	AVMObservabilityMetric	= "receipt_root_generation_time"

	AVMEventMessageSubmitted	AVMObservabilityEvent	= "avm_message_submitted"
	AVMEventMessageScheduled	AVMObservabilityEvent	= "avm_message_scheduled"
	AVMEventMessageExecuted		AVMObservabilityEvent	= "avm_message_executed"
	AVMEventMessageFailed		AVMObservabilityEvent	= "avm_message_failed"
	AVMEventMessageRetried		AVMObservabilityEvent	= "avm_message_retried"
	AVMEventMessageExpired		AVMObservabilityEvent	= "avm_message_expired"
	AVMEventMessageBounced		AVMObservabilityEvent	= "avm_message_bounced"
	AVMEventDeadLettered		AVMObservabilityEvent	= "avm_dead_lettered"
	AVMEventActorCreated		AVMObservabilityEvent	= "avm_actor_created"
	AVMEventActorMessageHandled	AVMObservabilityEvent	= "avm_actor_message_handled"
	AVMEventContinuationCreated	AVMObservabilityEvent	= "avm_continuation_created"
	AVMEventContinuationResumed	AVMObservabilityEvent	= "avm_continuation_resumed"
	AVMEventContinuationExpired	AVMObservabilityEvent	= "avm_continuation_expired"
	AVMEventContractExecuted	AVMObservabilityEvent	= "avm_contract_executed"
	AVMEventInterfaceRegistered	AVMObservabilityEvent	= "avm_interface_registered"
	AVMEventRuntimeUpgradeScheduled	AVMObservabilityEvent	= "avm_runtime_upgrade_scheduled"

	AVMAlertDeadLetterSpike				AVMObservabilityAlert	= "dead_letter_spike"
	AVMAlertRetryQueueBacklog			AVMObservabilityAlert	= "retry_queue_backlog"
	AVMAlertDelayedQueueBacklog			AVMObservabilityAlert	= "delayed_queue_backlog"
	AVMAlertZoneAsyncBudgetSaturation		AVMObservabilityAlert	= "zone_async_budget_saturation"
	AVMAlertActorMailboxBacklog			AVMObservabilityAlert	= "actor_mailbox_backlog"
	AVMAlertContinuationExpirySpike			AVMObservabilityAlert	= "continuation_expiry_spike"
	AVMAlertContractFailureSpike			AVMObservabilityAlert	= "contract_failure_spike"
	AVMAlertReceiptGenerationLatencyThreshold	AVMObservabilityAlert	= "receipt_generation_latency_above_threshold"
	AVMAlertQueueRootGenerationLatencyThreshold	AVMObservabilityAlert	= "queue_root_generation_latency_above_threshold"

	MaxAVMObservabilityItems	= 64
	MaxAVMObservabilityNameBytes	= 128
)

type AVMObservabilityMetric string
type AVMObservabilityEvent string
type AVMObservabilityAlert string

type AVMObservabilitySpec struct {
	Metrics		[]AVMObservabilityMetric
	Events		[]AVMObservabilityEvent
	Alerts		[]AVMObservabilityAlert
	SpecHash	string
}

func DefaultAVMObservabilitySpec() (AVMObservabilitySpec, error) {
	spec := AVMObservabilitySpec{
		Metrics:	AllAVMObservabilityMetrics(),
		Events:		AllAVMObservabilityEvents(),
		Alerts:		AllAVMObservabilityAlerts(),
	}
	spec.SpecHash = ComputeAVMObservabilitySpecHash(spec)
	return spec, spec.Validate()
}

func (s AVMObservabilitySpec) Validate() error {
	s = canonicalAVMObservabilitySpec(s)
	if err := validateAVMObservabilityMetrics(s.Metrics); err != nil {
		return err
	}
	if err := validateAVMObservabilityEvents(s.Events); err != nil {
		return err
	}
	if err := validateAVMObservabilityAlerts(s.Alerts); err != nil {
		return err
	}
	if s.SpecHash == "" {
		return errors.New("AVM observability spec hash is required")
	}
	if err := validateAVMComparisonHash("AVM observability spec hash", s.SpecHash); err != nil {
		return err
	}
	if s.SpecHash != ComputeAVMObservabilitySpecHash(s) {
		return errors.New("AVM observability spec hash mismatch")
	}
	return nil
}

func AllAVMObservabilityMetrics() []AVMObservabilityMetric {
	metrics := []AVMObservabilityMetric{
		AVMMetricAsyncMessagesSubmitted,
		AVMMetricAsyncMessagesExecuted,
		AVMMetricAsyncMessagesExpired,
		AVMMetricAsyncMessagesBounced,
		AVMMetricDeadLetterCount,
		AVMMetricRetryQueueSize,
		AVMMetricDelayedQueueSize,
		AVMMetricActorCount,
		AVMMetricActorMailboxDepth,
		AVMMetricContinuationsActive,
		AVMMetricContinuationsExpired,
		AVMMetricContractExecutions,
		AVMMetricContractFailures,
		AVMMetricGasReserved,
		AVMMetricGasConsumed,
		AVMMetricZoneAsyncBudgetUsage,
		AVMMetricQueueDrainLatency,
		AVMMetricReceiptRootGenerationTime,
	}
	sort.Slice(metrics, func(i, j int) bool { return metrics[i] < metrics[j] })
	return metrics
}

func AllAVMObservabilityEvents() []AVMObservabilityEvent {
	events := []AVMObservabilityEvent{
		AVMEventMessageSubmitted,
		AVMEventMessageScheduled,
		AVMEventMessageExecuted,
		AVMEventMessageFailed,
		AVMEventMessageRetried,
		AVMEventMessageExpired,
		AVMEventMessageBounced,
		AVMEventDeadLettered,
		AVMEventActorCreated,
		AVMEventActorMessageHandled,
		AVMEventContinuationCreated,
		AVMEventContinuationResumed,
		AVMEventContinuationExpired,
		AVMEventContractExecuted,
		AVMEventInterfaceRegistered,
		AVMEventRuntimeUpgradeScheduled,
	}
	sort.Slice(events, func(i, j int) bool { return events[i] < events[j] })
	return events
}

func AllAVMObservabilityAlerts() []AVMObservabilityAlert {
	alerts := []AVMObservabilityAlert{
		AVMAlertDeadLetterSpike,
		AVMAlertRetryQueueBacklog,
		AVMAlertDelayedQueueBacklog,
		AVMAlertZoneAsyncBudgetSaturation,
		AVMAlertActorMailboxBacklog,
		AVMAlertContinuationExpirySpike,
		AVMAlertContractFailureSpike,
		AVMAlertReceiptGenerationLatencyThreshold,
		AVMAlertQueueRootGenerationLatencyThreshold,
	}
	sort.Slice(alerts, func(i, j int) bool { return alerts[i] < alerts[j] })
	return alerts
}

func IsAVMObservabilityMetric(metric AVMObservabilityMetric) bool {
	for _, required := range AllAVMObservabilityMetrics() {
		if metric == required {
			return true
		}
	}
	return false
}

func IsAVMObservabilityEvent(event AVMObservabilityEvent) bool {
	for _, required := range AllAVMObservabilityEvents() {
		if event == required {
			return true
		}
	}
	return false
}

func IsAVMObservabilityAlert(alert AVMObservabilityAlert) bool {
	for _, required := range AllAVMObservabilityAlerts() {
		if alert == required {
			return true
		}
	}
	return false
}

func ComputeAVMObservabilitySpecHash(spec AVMObservabilitySpec) string {
	spec = canonicalAVMObservabilitySpec(spec)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-observability-spec-v1")
	writeEngineUint64(h, uint64(len(spec.Metrics)))
	for _, metric := range spec.Metrics {
		writeEnginePart(h, string(metric))
	}
	writeEngineUint64(h, uint64(len(spec.Events)))
	for _, event := range spec.Events {
		writeEnginePart(h, string(event))
	}
	writeEngineUint64(h, uint64(len(spec.Alerts)))
	for _, alert := range spec.Alerts {
		writeEnginePart(h, string(alert))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMObservabilitySpec(spec AVMObservabilitySpec) AVMObservabilitySpec {
	spec.Metrics = append([]AVMObservabilityMetric(nil), spec.Metrics...)
	for i := range spec.Metrics {
		spec.Metrics[i] = AVMObservabilityMetric(strings.TrimSpace(string(spec.Metrics[i])))
	}
	sort.SliceStable(spec.Metrics, func(i, j int) bool { return spec.Metrics[i] < spec.Metrics[j] })
	spec.Events = append([]AVMObservabilityEvent(nil), spec.Events...)
	for i := range spec.Events {
		spec.Events[i] = AVMObservabilityEvent(strings.TrimSpace(string(spec.Events[i])))
	}
	sort.SliceStable(spec.Events, func(i, j int) bool { return spec.Events[i] < spec.Events[j] })
	spec.Alerts = append([]AVMObservabilityAlert(nil), spec.Alerts...)
	for i := range spec.Alerts {
		spec.Alerts[i] = AVMObservabilityAlert(strings.TrimSpace(string(spec.Alerts[i])))
	}
	sort.SliceStable(spec.Alerts, func(i, j int) bool { return spec.Alerts[i] < spec.Alerts[j] })
	spec.SpecHash = strings.TrimSpace(spec.SpecHash)
	return spec
}

func validateAVMObservabilityMetrics(metrics []AVMObservabilityMetric) error {
	required := AllAVMObservabilityMetrics()
	if len(metrics) != len(required) || len(metrics) > MaxAVMObservabilityItems {
		return fmt.Errorf("AVM observability spec must contain every section 21 metric")
	}
	seen := make(map[AVMObservabilityMetric]struct{}, len(metrics))
	previous := ""
	for _, metric := range metrics {
		if err := validateAVMObservabilityName("AVM observability metric", string(metric)); err != nil {
			return err
		}
		if !IsAVMObservabilityMetric(metric) {
			return fmt.Errorf("invalid AVM observability metric %q", metric)
		}
		if _, found := seen[metric]; found {
			return fmt.Errorf("duplicate AVM observability metric %q", metric)
		}
		current := string(metric)
		if previous != "" && previous >= current {
			return errors.New("AVM observability metrics must be sorted canonically")
		}
		previous = current
		seen[metric] = struct{}{}
	}
	for _, metric := range required {
		if _, found := seen[metric]; !found {
			return fmt.Errorf("AVM observability spec missing metric %s", metric)
		}
	}
	return nil
}

func validateAVMObservabilityEvents(events []AVMObservabilityEvent) error {
	required := AllAVMObservabilityEvents()
	if len(events) != len(required) || len(events) > MaxAVMObservabilityItems {
		return fmt.Errorf("AVM observability spec must contain every section 21 event")
	}
	seen := make(map[AVMObservabilityEvent]struct{}, len(events))
	previous := ""
	for _, event := range events {
		if err := validateAVMObservabilityName("AVM observability event", string(event)); err != nil {
			return err
		}
		if !IsAVMObservabilityEvent(event) {
			return fmt.Errorf("invalid AVM observability event %q", event)
		}
		if _, found := seen[event]; found {
			return fmt.Errorf("duplicate AVM observability event %q", event)
		}
		current := string(event)
		if previous != "" && previous >= current {
			return errors.New("AVM observability events must be sorted canonically")
		}
		previous = current
		seen[event] = struct{}{}
	}
	for _, event := range required {
		if _, found := seen[event]; !found {
			return fmt.Errorf("AVM observability spec missing event %s", event)
		}
	}
	return nil
}

func validateAVMObservabilityAlerts(alerts []AVMObservabilityAlert) error {
	required := AllAVMObservabilityAlerts()
	if len(alerts) != len(required) || len(alerts) > MaxAVMObservabilityItems {
		return fmt.Errorf("AVM observability spec must contain every section 21 alert")
	}
	seen := make(map[AVMObservabilityAlert]struct{}, len(alerts))
	previous := ""
	for _, alert := range alerts {
		if err := validateAVMObservabilityName("AVM observability alert", string(alert)); err != nil {
			return err
		}
		if !IsAVMObservabilityAlert(alert) {
			return fmt.Errorf("invalid AVM observability alert %q", alert)
		}
		if _, found := seen[alert]; found {
			return fmt.Errorf("duplicate AVM observability alert %q", alert)
		}
		current := string(alert)
		if previous != "" && previous >= current {
			return errors.New("AVM observability alerts must be sorted canonically")
		}
		previous = current
		seen[alert] = struct{}{}
	}
	for _, alert := range required {
		if _, found := seen[alert]; !found {
			return fmt.Errorf("AVM observability spec missing alert %s", alert)
		}
	}
	return nil
}

func validateAVMObservabilityName(fieldName, value string) error {
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", fieldName)
	}
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > MaxAVMObservabilityNameBytes {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxAVMObservabilityNameBytes)
	}
	for _, r := range value {
		if r < 0x20 || r == '|' {
			return fmt.Errorf("%s contains invalid character", fieldName)
		}
	}
	return nil
}
