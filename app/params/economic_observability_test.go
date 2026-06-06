package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultEconomicObservabilityReportIsComplete(t *testing.T) {
	report := BuildEconomicObservabilityReport(nil, nil)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Metrics, 18)
	require.Len(t, report.Events, 13)
	require.Equal(t, 18, report.RequiredMetrics)
	require.Equal(t, 13, report.RequiredEvents)
	require.Equal(t, 18, report.CoveredMetrics)
	require.Equal(t, 13, report.CoveredEvents)
	require.Equal(t, BasisPoints, report.MetricCoverageBps)
	require.Equal(t, BasisPoints, report.EventCoverageBps)
	require.Contains(t, report.GovernanceSummary, "required_metrics=18/18")
	require.Contains(t, report.GovernanceSummary, "required_events=13/13")

	for _, metric := range report.Metrics {
		require.Equal(t, EconomicObservabilityKindMetric, metric.Kind)
		require.True(t, metric.Required)
		require.True(t, metric.Queryable)
		require.True(t, metric.TelemetryEnabled)
		require.NotZero(t, metric.SchemaVersion)
		require.NotEmpty(t, metric.Source)
		require.NotEmpty(t, metric.Labels)
	}
	for _, event := range report.Events {
		require.Equal(t, EconomicObservabilityKindEvent, event.Kind)
		require.True(t, event.Required)
		require.True(t, event.TelemetryEnabled)
		require.True(t, event.Emitted)
		require.NotZero(t, event.SchemaVersion)
		require.NotEmpty(t, event.Source)
		require.NotEmpty(t, event.Labels)
	}
}

func TestEconomicObservabilityRejectsMissingAndDuplicateMetric(t *testing.T) {
	metrics := DefaultEconomicObservabilityMetrics()
	metrics = append(metrics[:1], metrics[2:]...)
	metrics = append(metrics, metrics[0])

	report := BuildEconomicObservabilityReport(metrics, nil)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicMetricGrossMintedPerEpoch+":missing_required_observability")
	require.Contains(t, report.Failed, EconomicMetricCurrentInflationRate+":duplicate_signal")
	require.Less(t, report.MetricCoverageBps, BasisPoints)
}

func TestEconomicObservabilityRequiresQueryableMetrics(t *testing.T) {
	metrics := DefaultEconomicObservabilityMetrics()
	for i := range metrics {
		if metrics[i].ID == EconomicMetricStorageRentWarnings {
			metrics[i].Queryable = false
			metrics[i].TelemetryEnabled = false
		}
	}

	report := BuildEconomicObservabilityReport(metrics, nil)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicMetricStorageRentWarnings+":metric_not_queryable")
	require.Contains(t, report.Failed, EconomicMetricStorageRentWarnings+":telemetry_disabled")
	require.Less(t, report.MetricCoverageBps, BasisPoints)
}

func TestEconomicObservabilityRequiresEmittedEvents(t *testing.T) {
	events := DefaultEconomicObservabilityEvents()
	for i := range events {
		if events[i].ID == EconomicEventDeflationGuard {
			events[i].Emitted = false
		}
	}

	report := BuildEconomicObservabilityReport(nil, events)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicEventDeflationGuard+":event_not_emitted")
	require.Less(t, report.EventCoverageBps, BasisPoints)
}

func TestEconomicObservabilityRejectsBadSignalMetadata(t *testing.T) {
	metrics := DefaultEconomicObservabilityMetrics()
	metrics[0].Kind = EconomicObservabilityKindEvent
	metrics[1].Source = ""
	metrics[2].Labels = []string{" "}

	events := DefaultEconomicObservabilityEvents()
	events[0].SchemaVersion = 0
	events = append(events, EconomicObservabilitySignal{Kind: EconomicObservabilityKindEvent, Required: true})

	report := BuildEconomicObservabilityReport(metrics, events)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicMetricCurrentInflationRate+":wrong_observability_kind")
	require.Contains(t, report.Failed, EconomicMetricGrossMintedPerEpoch+":source_missing")
	require.Contains(t, report.Failed, EconomicMetricBurnedPerEpoch+":label_0_blank")
	require.Contains(t, report.Failed, EconomicEventInflationUpdate+":schema_version_missing")
	require.Contains(t, report.Failed, EconomicObservabilityKindEvent+":signal_id_required")
}
