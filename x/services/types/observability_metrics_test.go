package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceObservabilityMetricsManifestCoversSection181(t *testing.T) {
	manifest, err := DefaultServiceObservabilityMetricsManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.Metrics, 18)

	required := map[ServiceObservabilityMetricID]bool{
		ServiceMetricActiveServices:			false,
		ServiceMetricServicesByType:			false,
		ServiceMetricServicesByTrustModel:		false,
		ServiceMetricRegisteredInterfaces:		false,
		ServiceMetricActiveProviders:			false,
		ServiceMetricProviderCollateralTotal:		false,
		ServiceMetricCallsSubmitted:			false,
		ServiceMetricCallsExecuted:			false,
		ServiceMetricCallsFailed:			false,
		ServiceMetricCallsExpired:			false,
		ServiceMetricReceiptsAnchored:			false,
		ServiceMetricPaymentEscrowTotal:		false,
		ServiceMetricDisputesOpened:			false,
		ServiceMetricDisputesResolved:			false,
		ServiceMetricAverageServiceLookupLatency:	false,
		ServiceMetricAverageInterfaceLookupLatency:	false,
		ServiceMetricReceiptProofGenerationLatency:	false,
		ServiceMetricBlockSTMConflictRateServiceCalls:	false,
	}
	for _, metric := range manifest.Metrics {
		_, found := required[metric.MetricID]
		require.Truef(t, found, "unexpected metric %s", metric.MetricID)
		required[metric.MetricID] = true
		require.NotEmpty(t, metric.Source)
		require.Equal(t, ComputeServiceObservabilityMetricHash(metric), metric.MetricHash)
	}
	for metricID, found := range required {
		require.Truef(t, found, "missing metric %s", metricID)
	}
	require.Equal(t, ComputeServiceObservabilityMetricsManifestHash(manifest), manifest.ManifestHash)
}

func TestServiceObservabilityMetricsManifestRejectsMissingMetric(t *testing.T) {
	manifest, err := DefaultServiceObservabilityMetricsManifest()
	require.NoError(t, err)

	_, err = NewServiceObservabilityMetricsManifest(manifest.Metrics[1:])
	require.ErrorContains(t, err, "must include 18 metrics")
}

func TestServiceObservabilityMetricsManifestRejectsDuplicateMetric(t *testing.T) {
	manifest, err := DefaultServiceObservabilityMetricsManifest()
	require.NoError(t, err)
	metrics := append([]ServiceObservabilityMetric(nil), manifest.Metrics...)
	metrics[len(metrics)-1] = metrics[0]

	_, err = NewServiceObservabilityMetricsManifest(metrics)
	require.ErrorContains(t, err, "duplicate services observability metric")
}

func TestServiceObservabilityMetricsManifestRejectsHashTampering(t *testing.T) {
	manifest, err := DefaultServiceObservabilityMetricsManifest()
	require.NoError(t, err)

	tamperedMetric := manifest
	tamperedMetric.Metrics = append([]ServiceObservabilityMetric(nil), manifest.Metrics...)
	tamperedMetric.Metrics[0].Source = "x/services:tampered"
	require.ErrorContains(t, tamperedMetric.Validate(), "hash mismatch")

	tamperedManifest := manifest
	tamperedManifest.ManifestHash = testDistributedHash("tampered-observability-manifest")
	require.ErrorContains(t, tamperedManifest.Validate(), "manifest hash mismatch")
}

func TestServiceObservabilityMetricClassifiers(t *testing.T) {
	require.True(t, IsServiceObservabilityMetricID(ServiceMetricCallsSubmitted))
	require.False(t, IsServiceObservabilityMetricID(ServiceObservabilityMetricID("unknown")))
	require.True(t, IsServiceObservabilityMetricCategory(ServiceMetricCategoryPerformance))
	require.False(t, IsServiceObservabilityMetricCategory(ServiceObservabilityMetricCategory("unknown")))
	require.True(t, IsServiceObservabilityMetricUnit(ServiceMetricUnitRatioPPM))
	require.False(t, IsServiceObservabilityMetricUnit(ServiceObservabilityMetricUnit("unknown")))
}
