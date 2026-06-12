package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceObservabilitySignalsManifestCoversEventsAndAlerts(t *testing.T) {
	manifest, err := DefaultServiceObservabilitySignalsManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.Events, 18)
	require.Len(t, manifest.Alerts, 8)

	requiredEvents := map[ServiceObservabilityEventID]bool{
		ServiceEventRegistered:			false,
		ServiceEventUpdated:			false,
		ServiceEventRenewed:			false,
		ServiceEventDisabled:			false,
		ServiceEventIdentityBound:		false,
		ServiceEventInterfaceRegistered:	false,
		ServiceEventInterfaceDeprecated:	false,
		ServiceEventCallSubmitted:		false,
		ServiceEventCallExecuted:		false,
		ServiceEventCallFailed:			false,
		ServiceEventResultAnchored:		false,
		ServiceEventReceiptCommitted:		false,
		ServiceEventPaymentEscrowed:		false,
		ServiceEventPaymentSettled:		false,
		ServiceEventProviderRegistered:		false,
		ServiceEventProviderCollateralStaked:	false,
		ServiceEventProviderFaultSubmitted:	false,
		ServiceEventProviderPenalized:		false,
	}
	for _, event := range manifest.Events {
		_, found := requiredEvents[event.EventID]
		require.Truef(t, found, "unexpected event %s", event.EventID)
		requiredEvents[event.EventID] = true
		require.NotEmpty(t, event.Source)
		require.Equal(t, ComputeServiceObservabilityEventHash(event), event.EventHash)
	}
	for eventID, found := range requiredEvents {
		require.Truef(t, found, "missing event %s", eventID)
	}

	requiredAlerts := map[ServiceObservabilityAlertID]bool{
		ServiceAlertCallFailureSpike:			false,
		ServiceAlertReceiptAnchoringBacklog:		false,
		ServiceAlertProviderFaultSpike:			false,
		ServiceAlertInterfaceHashMismatchAttempt:	false,
		ServiceAlertExpiredServiceReceivingCalls:	false,
		ServiceAlertPaymentEscrowSettlementBacklog:	false,
		ServiceAlertReceiptProofLatencyThreshold:	false,
		ServiceAlertRegistryLookupLatencyThreshold:	false,
	}
	for _, alert := range manifest.Alerts {
		_, found := requiredAlerts[alert.AlertID]
		require.Truef(t, found, "unexpected alert %s", alert.AlertID)
		requiredAlerts[alert.AlertID] = true
		require.NotZero(t, alert.WindowBlocks)
		require.NotZero(t, alert.Threshold)
		require.True(t, IsServiceObservabilityMetricID(alert.MetricID))
		require.True(t, IsServiceObservabilityEventID(alert.TriggerEventID))
		require.Equal(t, ComputeServiceObservabilityAlertHash(alert), alert.AlertHash)
	}
	for alertID, found := range requiredAlerts {
		require.Truef(t, found, "missing alert %s", alertID)
	}
	require.Equal(t, ComputeServiceObservabilitySignalsManifestHash(manifest), manifest.ManifestHash)
}

func TestServiceObservabilitySignalsManifestRejectsMissingSignals(t *testing.T) {
	manifest, err := DefaultServiceObservabilitySignalsManifest()
	require.NoError(t, err)

	_, err = NewServiceObservabilitySignalsManifest(manifest.Events[1:], manifest.Alerts)
	require.ErrorContains(t, err, "must include 18 events")

	_, err = NewServiceObservabilitySignalsManifest(manifest.Events, manifest.Alerts[1:])
	require.ErrorContains(t, err, "must include 8 alerts")
}

func TestServiceObservabilitySignalsManifestRejectsDuplicateSignals(t *testing.T) {
	manifest, err := DefaultServiceObservabilitySignalsManifest()
	require.NoError(t, err)

	events := append([]ServiceObservabilityEvent(nil), manifest.Events...)
	events[len(events)-1] = events[0]
	_, err = NewServiceObservabilitySignalsManifest(events, manifest.Alerts)
	require.ErrorContains(t, err, "duplicate services observability event")

	alerts := append([]ServiceObservabilityAlert(nil), manifest.Alerts...)
	alerts[len(alerts)-1] = alerts[0]
	_, err = NewServiceObservabilitySignalsManifest(manifest.Events, alerts)
	require.ErrorContains(t, err, "duplicate services observability alert")
}

func TestServiceObservabilitySignalsManifestRejectsHashTampering(t *testing.T) {
	manifest, err := DefaultServiceObservabilitySignalsManifest()
	require.NoError(t, err)

	tamperedEvent := manifest
	tamperedEvent.Events = append([]ServiceObservabilityEvent(nil), manifest.Events...)
	tamperedEvent.Events[0].Source = "x/services:tampered"
	require.ErrorContains(t, tamperedEvent.Validate(), "event")
	require.ErrorContains(t, tamperedEvent.Validate(), "hash mismatch")

	tamperedAlert := manifest
	tamperedAlert.Alerts = append([]ServiceObservabilityAlert(nil), manifest.Alerts...)
	tamperedAlert.Alerts[0].Threshold++
	require.ErrorContains(t, tamperedAlert.Validate(), "alert")
	require.ErrorContains(t, tamperedAlert.Validate(), "hash mismatch")

	tamperedManifest := manifest
	tamperedManifest.ManifestHash = testDistributedHash("tampered-observability-signals-manifest")
	require.ErrorContains(t, tamperedManifest.Validate(), "signals manifest hash mismatch")
}

func TestServiceObservabilitySignalsClassifiers(t *testing.T) {
	require.True(t, IsServiceObservabilityEventID(ServiceEventCallFailed))
	require.False(t, IsServiceObservabilityEventID(ServiceObservabilityEventID("unknown")))
	require.True(t, IsServiceObservabilityEventCategory(ServiceEventCategoryProvider))
	require.False(t, IsServiceObservabilityEventCategory(ServiceObservabilityEventCategory("unknown")))
	require.True(t, IsServiceObservabilityAlertID(ServiceAlertProviderFaultSpike))
	require.False(t, IsServiceObservabilityAlertID(ServiceObservabilityAlertID("unknown")))
	require.True(t, IsServiceObservabilityAlertSeverity(ServiceAlertSeverityCritical))
	require.False(t, IsServiceObservabilityAlertSeverity(ServiceObservabilityAlertSeverity("unknown")))
}
