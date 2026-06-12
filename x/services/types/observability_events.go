package types

import (
	"errors"
	"fmt"
	"sort"
)

type ServiceObservabilityEventID string
type ServiceObservabilityEventCategory string
type ServiceObservabilityAlertID string
type ServiceObservabilityAlertSeverity string

const (
	ServiceEventRegistered			ServiceObservabilityEventID	= "service_registered"
	ServiceEventUpdated			ServiceObservabilityEventID	= "service_updated"
	ServiceEventRenewed			ServiceObservabilityEventID	= "service_renewed"
	ServiceEventDisabled			ServiceObservabilityEventID	= "service_disabled"
	ServiceEventIdentityBound		ServiceObservabilityEventID	= "service_identity_bound"
	ServiceEventInterfaceRegistered		ServiceObservabilityEventID	= "interface_registered"
	ServiceEventInterfaceDeprecated		ServiceObservabilityEventID	= "interface_deprecated"
	ServiceEventCallSubmitted		ServiceObservabilityEventID	= "service_call_submitted"
	ServiceEventCallExecuted		ServiceObservabilityEventID	= "service_call_executed"
	ServiceEventCallFailed			ServiceObservabilityEventID	= "service_call_failed"
	ServiceEventResultAnchored		ServiceObservabilityEventID	= "service_result_anchored"
	ServiceEventReceiptCommitted		ServiceObservabilityEventID	= "service_receipt_committed"
	ServiceEventPaymentEscrowed		ServiceObservabilityEventID	= "service_payment_escrowed"
	ServiceEventPaymentSettled		ServiceObservabilityEventID	= "service_payment_settled"
	ServiceEventProviderRegistered		ServiceObservabilityEventID	= "provider_registered"
	ServiceEventProviderCollateralStaked	ServiceObservabilityEventID	= "provider_collateral_staked"
	ServiceEventProviderFaultSubmitted	ServiceObservabilityEventID	= "provider_fault_submitted"
	ServiceEventProviderPenalized		ServiceObservabilityEventID	= "provider_penalized"

	ServiceEventCategoryRegistry	ServiceObservabilityEventCategory	= "registry"
	ServiceEventCategoryInterface	ServiceObservabilityEventCategory	= "interface"
	ServiceEventCategoryCall	ServiceObservabilityEventCategory	= "call"
	ServiceEventCategoryReceipt	ServiceObservabilityEventCategory	= "receipt"
	ServiceEventCategoryPayment	ServiceObservabilityEventCategory	= "payment"
	ServiceEventCategoryProvider	ServiceObservabilityEventCategory	= "provider"

	ServiceAlertCallFailureSpike			ServiceObservabilityAlertID	= "service_call_failure_spike"
	ServiceAlertReceiptAnchoringBacklog		ServiceObservabilityAlertID	= "receipt_anchoring_backlog"
	ServiceAlertProviderFaultSpike			ServiceObservabilityAlertID	= "provider_fault_spike"
	ServiceAlertInterfaceHashMismatchAttempt	ServiceObservabilityAlertID	= "interface_hash_mismatch_attempt"
	ServiceAlertExpiredServiceReceivingCalls	ServiceObservabilityAlertID	= "expired_service_receiving_calls"
	ServiceAlertPaymentEscrowSettlementBacklog	ServiceObservabilityAlertID	= "payment_escrow_settlement_backlog"
	ServiceAlertReceiptProofLatencyThreshold	ServiceObservabilityAlertID	= "receipt_proof_latency_above_threshold"
	ServiceAlertRegistryLookupLatencyThreshold	ServiceObservabilityAlertID	= "registry_lookup_latency_above_threshold"

	ServiceAlertSeverityWarning	ServiceObservabilityAlertSeverity	= "warning"
	ServiceAlertSeverityCritical	ServiceObservabilityAlertSeverity	= "critical"
)

type ServiceObservabilityEvent struct {
	EventID		ServiceObservabilityEventID
	Category	ServiceObservabilityEventCategory
	Source		string
	EventHash	string
}

type ServiceObservabilityAlert struct {
	AlertID		ServiceObservabilityAlertID
	Severity	ServiceObservabilityAlertSeverity
	MetricID	ServiceObservabilityMetricID
	TriggerEventID	ServiceObservabilityEventID
	WindowBlocks	uint64
	Threshold	uint64
	AlertHash	string
}

type ServiceObservabilitySignalsManifest struct {
	Events		[]ServiceObservabilityEvent
	Alerts		[]ServiceObservabilityAlert
	ManifestHash	string
}

func DefaultServiceObservabilitySignalsManifest() (ServiceObservabilitySignalsManifest, error) {
	return NewServiceObservabilitySignalsManifest(defaultServiceObservabilityEvents(), defaultServiceObservabilityAlerts())
}

func NewServiceObservabilitySignalsManifest(events []ServiceObservabilityEvent, alerts []ServiceObservabilityAlert) (ServiceObservabilitySignalsManifest, error) {
	manifest := ServiceObservabilitySignalsManifest{
		Events:	canonicalServiceObservabilityEvents(events),
		Alerts:	canonicalServiceObservabilityAlerts(alerts),
	}
	if err := manifest.ValidateFormat(); err != nil {
		return ServiceObservabilitySignalsManifest{}, err
	}
	for i := range manifest.Events {
		manifest.Events[i].EventHash = ComputeServiceObservabilityEventHash(manifest.Events[i])
	}
	for i := range manifest.Alerts {
		manifest.Alerts[i].AlertHash = ComputeServiceObservabilityAlertHash(manifest.Alerts[i])
	}
	manifest.ManifestHash = ComputeServiceObservabilitySignalsManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (manifest ServiceObservabilitySignalsManifest) ValidateFormat() error {
	manifest.Events = canonicalServiceObservabilityEvents(manifest.Events)
	manifest.Alerts = canonicalServiceObservabilityAlerts(manifest.Alerts)
	if len(manifest.Events) != len(requiredServiceObservabilityEventIDs()) {
		return fmt.Errorf("services observability signals manifest must include %d events", len(requiredServiceObservabilityEventIDs()))
	}
	if len(manifest.Alerts) != len(requiredServiceObservabilityAlertIDs()) {
		return fmt.Errorf("services observability signals manifest must include %d alerts", len(requiredServiceObservabilityAlertIDs()))
	}
	events := map[ServiceObservabilityEventID]struct{}{}
	for _, event := range manifest.Events {
		if err := event.ValidateFormat(); err != nil {
			return err
		}
		if _, found := events[event.EventID]; found {
			return fmt.Errorf("duplicate services observability event %q", event.EventID)
		}
		events[event.EventID] = struct{}{}
	}
	for _, eventID := range requiredServiceObservabilityEventIDs() {
		if _, found := events[eventID]; !found {
			return fmt.Errorf("missing services observability event %q", eventID)
		}
	}
	alerts := map[ServiceObservabilityAlertID]struct{}{}
	for _, alert := range manifest.Alerts {
		if err := alert.ValidateFormat(); err != nil {
			return err
		}
		if _, found := alerts[alert.AlertID]; found {
			return fmt.Errorf("duplicate services observability alert %q", alert.AlertID)
		}
		alerts[alert.AlertID] = struct{}{}
	}
	for _, alertID := range requiredServiceObservabilityAlertIDs() {
		if _, found := alerts[alertID]; !found {
			return fmt.Errorf("missing services observability alert %q", alertID)
		}
	}
	return nil
}

func (manifest ServiceObservabilitySignalsManifest) Validate() error {
	manifest.Events = canonicalServiceObservabilityEvents(manifest.Events)
	manifest.Alerts = canonicalServiceObservabilityAlerts(manifest.Alerts)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, event := range manifest.Events {
		if event.EventHash == "" {
			return fmt.Errorf("services observability event %q hash is required", event.EventID)
		}
		if expected := ComputeServiceObservabilityEventHash(event); event.EventHash != expected {
			return fmt.Errorf("services observability event %q hash mismatch: expected %s", event.EventID, expected)
		}
	}
	for _, alert := range manifest.Alerts {
		if alert.AlertHash == "" {
			return fmt.Errorf("services observability alert %q hash is required", alert.AlertID)
		}
		if expected := ComputeServiceObservabilityAlertHash(alert); alert.AlertHash != expected {
			return fmt.Errorf("services observability alert %q hash mismatch: expected %s", alert.AlertID, expected)
		}
	}
	if manifest.ManifestHash == "" {
		return errors.New("services observability signals manifest hash is required")
	}
	if expected := ComputeServiceObservabilitySignalsManifestHash(manifest); manifest.ManifestHash != expected {
		return fmt.Errorf("services observability signals manifest hash mismatch: expected %s", expected)
	}
	return nil
}

func (event ServiceObservabilityEvent) ValidateFormat() error {
	if !IsServiceObservabilityEventID(event.EventID) {
		return fmt.Errorf("unknown services observability event %q", event.EventID)
	}
	if !IsServiceObservabilityEventCategory(event.Category) {
		return fmt.Errorf("unknown services observability event category %q", event.Category)
	}
	if err := validateInterfaceToken("services observability event source", event.Source); err != nil {
		return err
	}
	return nil
}

func (alert ServiceObservabilityAlert) ValidateFormat() error {
	if !IsServiceObservabilityAlertID(alert.AlertID) {
		return fmt.Errorf("unknown services observability alert %q", alert.AlertID)
	}
	if !IsServiceObservabilityAlertSeverity(alert.Severity) {
		return fmt.Errorf("unknown services observability alert severity %q", alert.Severity)
	}
	if !IsServiceObservabilityMetricID(alert.MetricID) {
		return fmt.Errorf("unknown services observability alert metric %q", alert.MetricID)
	}
	if !IsServiceObservabilityEventID(alert.TriggerEventID) {
		return fmt.Errorf("unknown services observability alert trigger event %q", alert.TriggerEventID)
	}
	if alert.WindowBlocks == 0 {
		return fmt.Errorf("services observability alert %q window blocks must be positive", alert.AlertID)
	}
	if alert.Threshold == 0 {
		return fmt.Errorf("services observability alert %q threshold must be positive", alert.AlertID)
	}
	return nil
}

func ComputeServiceObservabilityEventHash(event ServiceObservabilityEvent) string {
	return servicesHashParts(
		"aetra-services-observability-event-v1",
		string(event.EventID),
		string(event.Category),
		event.Source,
	)
}

func ComputeServiceObservabilityAlertHash(alert ServiceObservabilityAlert) string {
	return servicesHashParts(
		"aetra-services-observability-alert-v1",
		string(alert.AlertID),
		string(alert.Severity),
		string(alert.MetricID),
		string(alert.TriggerEventID),
		fmt.Sprint(alert.WindowBlocks),
		fmt.Sprint(alert.Threshold),
	)
}

func ComputeServiceObservabilitySignalsManifestHash(manifest ServiceObservabilitySignalsManifest) string {
	manifest.Events = canonicalServiceObservabilityEvents(manifest.Events)
	manifest.Alerts = canonicalServiceObservabilityAlerts(manifest.Alerts)
	parts := []string{
		"aetra-services-observability-signals-manifest-v1",
		"events",
		fmt.Sprint(len(manifest.Events)),
	}
	for _, event := range manifest.Events {
		parts = append(parts, string(event.EventID), ComputeServiceObservabilityEventHash(event))
	}
	parts = append(parts, "alerts", fmt.Sprint(len(manifest.Alerts)))
	for _, alert := range manifest.Alerts {
		parts = append(parts, string(alert.AlertID), ComputeServiceObservabilityAlertHash(alert))
	}
	return servicesHashParts(parts...)
}

func IsServiceObservabilityEventID(eventID ServiceObservabilityEventID) bool {
	for _, required := range requiredServiceObservabilityEventIDs() {
		if eventID == required {
			return true
		}
	}
	return false
}

func IsServiceObservabilityEventCategory(category ServiceObservabilityEventCategory) bool {
	switch category {
	case ServiceEventCategoryRegistry, ServiceEventCategoryInterface, ServiceEventCategoryCall,
		ServiceEventCategoryReceipt, ServiceEventCategoryPayment, ServiceEventCategoryProvider:
		return true
	default:
		return false
	}
}

func IsServiceObservabilityAlertID(alertID ServiceObservabilityAlertID) bool {
	for _, required := range requiredServiceObservabilityAlertIDs() {
		if alertID == required {
			return true
		}
	}
	return false
}

func IsServiceObservabilityAlertSeverity(severity ServiceObservabilityAlertSeverity) bool {
	switch severity {
	case ServiceAlertSeverityWarning, ServiceAlertSeverityCritical:
		return true
	default:
		return false
	}
}

func defaultServiceObservabilityEvents() []ServiceObservabilityEvent {
	return []ServiceObservabilityEvent{
		newServiceObservabilityEvent(ServiceEventRegistered, ServiceEventCategoryRegistry, "x/services:register"),
		newServiceObservabilityEvent(ServiceEventUpdated, ServiceEventCategoryRegistry, "x/services:update"),
		newServiceObservabilityEvent(ServiceEventRenewed, ServiceEventCategoryRegistry, "x/services:renew"),
		newServiceObservabilityEvent(ServiceEventDisabled, ServiceEventCategoryRegistry, "x/services:disable"),
		newServiceObservabilityEvent(ServiceEventIdentityBound, ServiceEventCategoryRegistry, "x/services:identity/bind"),
		newServiceObservabilityEvent(ServiceEventInterfaceRegistered, ServiceEventCategoryInterface, "x/serviceinterface:register"),
		newServiceObservabilityEvent(ServiceEventInterfaceDeprecated, ServiceEventCategoryInterface, "x/serviceinterface:deprecate"),
		newServiceObservabilityEvent(ServiceEventCallSubmitted, ServiceEventCategoryCall, "x/servicecalls:submit"),
		newServiceObservabilityEvent(ServiceEventCallExecuted, ServiceEventCategoryCall, "x/servicecalls:execute"),
		newServiceObservabilityEvent(ServiceEventCallFailed, ServiceEventCategoryCall, "x/servicecalls:fail"),
		newServiceObservabilityEvent(ServiceEventResultAnchored, ServiceEventCategoryCall, "x/servicecalls:anchor_result"),
		newServiceObservabilityEvent(ServiceEventReceiptCommitted, ServiceEventCategoryReceipt, "x/servicereceipts:commit"),
		newServiceObservabilityEvent(ServiceEventPaymentEscrowed, ServiceEventCategoryPayment, "x/servicepayments:escrow"),
		newServiceObservabilityEvent(ServiceEventPaymentSettled, ServiceEventCategoryPayment, "x/servicepayments:settle"),
		newServiceObservabilityEvent(ServiceEventProviderRegistered, ServiceEventCategoryProvider, "x/serviceproviders:register"),
		newServiceObservabilityEvent(ServiceEventProviderCollateralStaked, ServiceEventCategoryProvider, "x/serviceproviders:stake_collateral"),
		newServiceObservabilityEvent(ServiceEventProviderFaultSubmitted, ServiceEventCategoryProvider, "x/serviceproviders:submit_fault"),
		newServiceObservabilityEvent(ServiceEventProviderPenalized, ServiceEventCategoryProvider, "x/serviceproviders:penalize"),
	}
}

func defaultServiceObservabilityAlerts() []ServiceObservabilityAlert {
	return []ServiceObservabilityAlert{
		newServiceObservabilityAlert(ServiceAlertCallFailureSpike, ServiceAlertSeverityCritical, ServiceMetricCallsFailed, ServiceEventCallFailed, 100, 10),
		newServiceObservabilityAlert(ServiceAlertReceiptAnchoringBacklog, ServiceAlertSeverityWarning, ServiceMetricReceiptsAnchored, ServiceEventResultAnchored, 100, 25),
		newServiceObservabilityAlert(ServiceAlertProviderFaultSpike, ServiceAlertSeverityCritical, ServiceMetricDisputesOpened, ServiceEventProviderFaultSubmitted, 100, 5),
		newServiceObservabilityAlert(ServiceAlertInterfaceHashMismatchAttempt, ServiceAlertSeverityCritical, ServiceMetricRegisteredInterfaces, ServiceEventInterfaceRegistered, 1, 1),
		newServiceObservabilityAlert(ServiceAlertExpiredServiceReceivingCalls, ServiceAlertSeverityCritical, ServiceMetricCallsExpired, ServiceEventCallSubmitted, 1, 1),
		newServiceObservabilityAlert(ServiceAlertPaymentEscrowSettlementBacklog, ServiceAlertSeverityWarning, ServiceMetricPaymentEscrowTotal, ServiceEventPaymentEscrowed, 100, 25),
		newServiceObservabilityAlert(ServiceAlertReceiptProofLatencyThreshold, ServiceAlertSeverityWarning, ServiceMetricReceiptProofGenerationLatency, ServiceEventReceiptCommitted, 10, 1_000_000_000),
		newServiceObservabilityAlert(ServiceAlertRegistryLookupLatencyThreshold, ServiceAlertSeverityWarning, ServiceMetricAverageServiceLookupLatency, ServiceEventRegistered, 10, 500_000_000),
	}
}

func newServiceObservabilityEvent(eventID ServiceObservabilityEventID, category ServiceObservabilityEventCategory, source string) ServiceObservabilityEvent {
	return ServiceObservabilityEvent{
		EventID:	eventID,
		Category:	category,
		Source:		source,
	}
}

func newServiceObservabilityAlert(alertID ServiceObservabilityAlertID, severity ServiceObservabilityAlertSeverity, metricID ServiceObservabilityMetricID, triggerEventID ServiceObservabilityEventID, windowBlocks uint64, threshold uint64) ServiceObservabilityAlert {
	return ServiceObservabilityAlert{
		AlertID:	alertID,
		Severity:	severity,
		MetricID:	metricID,
		TriggerEventID:	triggerEventID,
		WindowBlocks:	windowBlocks,
		Threshold:	threshold,
	}
}

func canonicalServiceObservabilityEvents(events []ServiceObservabilityEvent) []ServiceObservabilityEvent {
	canonical := append([]ServiceObservabilityEvent(nil), events...)
	sort.SliceStable(canonical, func(i, j int) bool {
		return canonical[i].EventID < canonical[j].EventID
	})
	return canonical
}

func canonicalServiceObservabilityAlerts(alerts []ServiceObservabilityAlert) []ServiceObservabilityAlert {
	canonical := append([]ServiceObservabilityAlert(nil), alerts...)
	sort.SliceStable(canonical, func(i, j int) bool {
		return canonical[i].AlertID < canonical[j].AlertID
	})
	return canonical
}

func requiredServiceObservabilityEventIDs() []ServiceObservabilityEventID {
	return []ServiceObservabilityEventID{
		ServiceEventCallExecuted,
		ServiceEventCallFailed,
		ServiceEventCallSubmitted,
		ServiceEventDisabled,
		ServiceEventIdentityBound,
		ServiceEventInterfaceDeprecated,
		ServiceEventInterfaceRegistered,
		ServiceEventPaymentEscrowed,
		ServiceEventPaymentSettled,
		ServiceEventProviderCollateralStaked,
		ServiceEventProviderFaultSubmitted,
		ServiceEventProviderPenalized,
		ServiceEventProviderRegistered,
		ServiceEventReceiptCommitted,
		ServiceEventRegistered,
		ServiceEventRenewed,
		ServiceEventResultAnchored,
		ServiceEventUpdated,
	}
}

func requiredServiceObservabilityAlertIDs() []ServiceObservabilityAlertID {
	return []ServiceObservabilityAlertID{
		ServiceAlertCallFailureSpike,
		ServiceAlertExpiredServiceReceivingCalls,
		ServiceAlertInterfaceHashMismatchAttempt,
		ServiceAlertPaymentEscrowSettlementBacklog,
		ServiceAlertProviderFaultSpike,
		ServiceAlertReceiptAnchoringBacklog,
		ServiceAlertReceiptProofLatencyThreshold,
		ServiceAlertRegistryLookupLatencyThreshold,
	}
}
