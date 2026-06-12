package types

import (
	"errors"
	"fmt"
	"sort"
)

type ServiceObservabilityMetricID string
type ServiceObservabilityMetricCategory string
type ServiceObservabilityMetricUnit string

const (
	ServiceMetricActiveServices			ServiceObservabilityMetricID	= "active_services"
	ServiceMetricServicesByType			ServiceObservabilityMetricID	= "services_by_type"
	ServiceMetricServicesByTrustModel		ServiceObservabilityMetricID	= "services_by_trust_model"
	ServiceMetricRegisteredInterfaces		ServiceObservabilityMetricID	= "registered_interfaces"
	ServiceMetricActiveProviders			ServiceObservabilityMetricID	= "active_providers"
	ServiceMetricProviderCollateralTotal		ServiceObservabilityMetricID	= "provider_collateral_total"
	ServiceMetricCallsSubmitted			ServiceObservabilityMetricID	= "calls_submitted"
	ServiceMetricCallsExecuted			ServiceObservabilityMetricID	= "calls_executed"
	ServiceMetricCallsFailed			ServiceObservabilityMetricID	= "calls_failed"
	ServiceMetricCallsExpired			ServiceObservabilityMetricID	= "calls_expired"
	ServiceMetricReceiptsAnchored			ServiceObservabilityMetricID	= "receipts_anchored"
	ServiceMetricPaymentEscrowTotal			ServiceObservabilityMetricID	= "payment_escrow_total"
	ServiceMetricDisputesOpened			ServiceObservabilityMetricID	= "disputes_opened"
	ServiceMetricDisputesResolved			ServiceObservabilityMetricID	= "disputes_resolved"
	ServiceMetricAverageServiceLookupLatency	ServiceObservabilityMetricID	= "average_service_lookup_latency"
	ServiceMetricAverageInterfaceLookupLatency	ServiceObservabilityMetricID	= "average_interface_lookup_latency"
	ServiceMetricReceiptProofGenerationLatency	ServiceObservabilityMetricID	= "receipt_proof_generation_latency"
	ServiceMetricBlockSTMConflictRateServiceCalls	ServiceObservabilityMetricID	= "blockstm_conflict_rate_for_service_calls"

	ServiceMetricCategoryRegistry		ServiceObservabilityMetricCategory	= "registry"
	ServiceMetricCategoryInterface		ServiceObservabilityMetricCategory	= "interface"
	ServiceMetricCategoryProvider		ServiceObservabilityMetricCategory	= "provider"
	ServiceMetricCategoryCall		ServiceObservabilityMetricCategory	= "call"
	ServiceMetricCategoryReceipt		ServiceObservabilityMetricCategory	= "receipt"
	ServiceMetricCategoryPayment		ServiceObservabilityMetricCategory	= "payment"
	ServiceMetricCategoryDispute		ServiceObservabilityMetricCategory	= "dispute"
	ServiceMetricCategoryLatency		ServiceObservabilityMetricCategory	= "latency"
	ServiceMetricCategoryPerformance	ServiceObservabilityMetricCategory	= "performance"

	ServiceMetricUnitCount		ServiceObservabilityMetricUnit	= "count"
	ServiceMetricUnitAmount		ServiceObservabilityMetricUnit	= "amount"
	ServiceMetricUnitNanoseconds	ServiceObservabilityMetricUnit	= "nanoseconds"
	ServiceMetricUnitRatioPPM	ServiceObservabilityMetricUnit	= "ratio_ppm"
)

type ServiceObservabilityMetric struct {
	MetricID	ServiceObservabilityMetricID
	Category	ServiceObservabilityMetricCategory
	Unit		ServiceObservabilityMetricUnit
	Source		string
	Bounded		bool
	MetricHash	string
}

type ServiceObservabilityMetricsManifest struct {
	Metrics		[]ServiceObservabilityMetric
	ManifestHash	string
}

func DefaultServiceObservabilityMetricsManifest() (ServiceObservabilityMetricsManifest, error) {
	return NewServiceObservabilityMetricsManifest([]ServiceObservabilityMetric{
		newServiceObservabilityMetric(ServiceMetricActiveServices, ServiceMetricCategoryRegistry, ServiceMetricUnitCount, "x/services:descriptors", true),
		newServiceObservabilityMetric(ServiceMetricServicesByType, ServiceMetricCategoryRegistry, ServiceMetricUnitCount, "x/services:index/type", true),
		newServiceObservabilityMetric(ServiceMetricServicesByTrustModel, ServiceMetricCategoryRegistry, ServiceMetricUnitCount, "x/services:index/trust_model", true),
		newServiceObservabilityMetric(ServiceMetricRegisteredInterfaces, ServiceMetricCategoryInterface, ServiceMetricUnitCount, "x/serviceinterface:interfaces", true),
		newServiceObservabilityMetric(ServiceMetricActiveProviders, ServiceMetricCategoryProvider, ServiceMetricUnitCount, "x/serviceproviders:providers", true),
		newServiceObservabilityMetric(ServiceMetricProviderCollateralTotal, ServiceMetricCategoryProvider, ServiceMetricUnitAmount, "x/serviceproviders:collateral", true),
		newServiceObservabilityMetric(ServiceMetricCallsSubmitted, ServiceMetricCategoryCall, ServiceMetricUnitCount, "x/servicecalls:calls", false),
		newServiceObservabilityMetric(ServiceMetricCallsExecuted, ServiceMetricCategoryCall, ServiceMetricUnitCount, "x/servicecalls:receipts/executed", false),
		newServiceObservabilityMetric(ServiceMetricCallsFailed, ServiceMetricCategoryCall, ServiceMetricUnitCount, "x/servicecalls:receipts/failed", false),
		newServiceObservabilityMetric(ServiceMetricCallsExpired, ServiceMetricCategoryCall, ServiceMetricUnitCount, "x/servicecalls:receipts/expired", false),
		newServiceObservabilityMetric(ServiceMetricReceiptsAnchored, ServiceMetricCategoryReceipt, ServiceMetricUnitCount, "x/servicereceipts:receipts", false),
		newServiceObservabilityMetric(ServiceMetricPaymentEscrowTotal, ServiceMetricCategoryPayment, ServiceMetricUnitAmount, "x/servicepayments:escrow", true),
		newServiceObservabilityMetric(ServiceMetricDisputesOpened, ServiceMetricCategoryDispute, ServiceMetricUnitCount, "x/services:disputes/opened", false),
		newServiceObservabilityMetric(ServiceMetricDisputesResolved, ServiceMetricCategoryDispute, ServiceMetricUnitCount, "x/services:disputes/resolved", false),
		newServiceObservabilityMetric(ServiceMetricAverageServiceLookupLatency, ServiceMetricCategoryLatency, ServiceMetricUnitNanoseconds, "x/services:query/service", false),
		newServiceObservabilityMetric(ServiceMetricAverageInterfaceLookupLatency, ServiceMetricCategoryLatency, ServiceMetricUnitNanoseconds, "x/serviceinterface:query/interface", false),
		newServiceObservabilityMetric(ServiceMetricReceiptProofGenerationLatency, ServiceMetricCategoryLatency, ServiceMetricUnitNanoseconds, "x/servicereceipts:query/proof", false),
		newServiceObservabilityMetric(ServiceMetricBlockSTMConflictRateServiceCalls, ServiceMetricCategoryPerformance, ServiceMetricUnitRatioPPM, "x/servicecalls:blockstm/conflicts", false),
	})
}

func NewServiceObservabilityMetricsManifest(metrics []ServiceObservabilityMetric) (ServiceObservabilityMetricsManifest, error) {
	manifest := ServiceObservabilityMetricsManifest{
		Metrics: canonicalServiceObservabilityMetrics(metrics),
	}
	if err := manifest.ValidateFormat(); err != nil {
		return ServiceObservabilityMetricsManifest{}, err
	}
	for i := range manifest.Metrics {
		manifest.Metrics[i].MetricHash = ComputeServiceObservabilityMetricHash(manifest.Metrics[i])
	}
	manifest.ManifestHash = ComputeServiceObservabilityMetricsManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (manifest ServiceObservabilityMetricsManifest) ValidateFormat() error {
	manifest.Metrics = canonicalServiceObservabilityMetrics(manifest.Metrics)
	if len(manifest.Metrics) != len(requiredServiceObservabilityMetricIDs()) {
		return fmt.Errorf("services observability manifest must include %d metrics", len(requiredServiceObservabilityMetricIDs()))
	}
	seen := map[ServiceObservabilityMetricID]struct{}{}
	for _, metric := range manifest.Metrics {
		if err := metric.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[metric.MetricID]; found {
			return fmt.Errorf("duplicate services observability metric %q", metric.MetricID)
		}
		seen[metric.MetricID] = struct{}{}
	}
	for _, metricID := range requiredServiceObservabilityMetricIDs() {
		if _, found := seen[metricID]; !found {
			return fmt.Errorf("missing services observability metric %q", metricID)
		}
	}
	return nil
}

func (manifest ServiceObservabilityMetricsManifest) Validate() error {
	manifest.Metrics = canonicalServiceObservabilityMetrics(manifest.Metrics)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, metric := range manifest.Metrics {
		if metric.MetricHash == "" {
			return fmt.Errorf("services observability metric %q hash is required", metric.MetricID)
		}
		if expected := ComputeServiceObservabilityMetricHash(metric); metric.MetricHash != expected {
			return fmt.Errorf("services observability metric %q hash mismatch: expected %s", metric.MetricID, expected)
		}
	}
	if manifest.ManifestHash == "" {
		return errors.New("services observability manifest hash is required")
	}
	if expected := ComputeServiceObservabilityMetricsManifestHash(manifest); manifest.ManifestHash != expected {
		return fmt.Errorf("services observability manifest hash mismatch: expected %s", expected)
	}
	return nil
}

func (metric ServiceObservabilityMetric) ValidateFormat() error {
	if !IsServiceObservabilityMetricID(metric.MetricID) {
		return fmt.Errorf("unknown services observability metric %q", metric.MetricID)
	}
	if !IsServiceObservabilityMetricCategory(metric.Category) {
		return fmt.Errorf("unknown services observability metric category %q", metric.Category)
	}
	if !IsServiceObservabilityMetricUnit(metric.Unit) {
		return fmt.Errorf("unknown services observability metric unit %q", metric.Unit)
	}
	if err := validateInterfaceToken("services observability metric source", metric.Source); err != nil {
		return err
	}
	return nil
}

func ComputeServiceObservabilityMetricHash(metric ServiceObservabilityMetric) string {
	return servicesHashParts(
		"aetra-services-observability-metric-v1",
		string(metric.MetricID),
		string(metric.Category),
		string(metric.Unit),
		metric.Source,
		fmt.Sprint(metric.Bounded),
	)
}

func ComputeServiceObservabilityMetricsManifestHash(manifest ServiceObservabilityMetricsManifest) string {
	manifest.Metrics = canonicalServiceObservabilityMetrics(manifest.Metrics)
	parts := []string{
		"aetra-services-observability-manifest-v1",
		fmt.Sprint(len(manifest.Metrics)),
	}
	for _, metric := range manifest.Metrics {
		parts = append(parts, string(metric.MetricID), ComputeServiceObservabilityMetricHash(metric))
	}
	return servicesHashParts(parts...)
}

func IsServiceObservabilityMetricID(metricID ServiceObservabilityMetricID) bool {
	for _, required := range requiredServiceObservabilityMetricIDs() {
		if metricID == required {
			return true
		}
	}
	return false
}

func IsServiceObservabilityMetricCategory(category ServiceObservabilityMetricCategory) bool {
	switch category {
	case ServiceMetricCategoryRegistry, ServiceMetricCategoryInterface, ServiceMetricCategoryProvider, ServiceMetricCategoryCall,
		ServiceMetricCategoryReceipt, ServiceMetricCategoryPayment, ServiceMetricCategoryDispute, ServiceMetricCategoryLatency,
		ServiceMetricCategoryPerformance:
		return true
	default:
		return false
	}
}

func IsServiceObservabilityMetricUnit(unit ServiceObservabilityMetricUnit) bool {
	switch unit {
	case ServiceMetricUnitCount, ServiceMetricUnitAmount, ServiceMetricUnitNanoseconds, ServiceMetricUnitRatioPPM:
		return true
	default:
		return false
	}
}

func newServiceObservabilityMetric(metricID ServiceObservabilityMetricID, category ServiceObservabilityMetricCategory, unit ServiceObservabilityMetricUnit, source string, bounded bool) ServiceObservabilityMetric {
	return ServiceObservabilityMetric{
		MetricID:	metricID,
		Category:	category,
		Unit:		unit,
		Source:		source,
		Bounded:	bounded,
	}
}

func canonicalServiceObservabilityMetrics(metrics []ServiceObservabilityMetric) []ServiceObservabilityMetric {
	canonical := append([]ServiceObservabilityMetric(nil), metrics...)
	sort.SliceStable(canonical, func(i, j int) bool {
		return canonical[i].MetricID < canonical[j].MetricID
	})
	return canonical
}

func requiredServiceObservabilityMetricIDs() []ServiceObservabilityMetricID {
	return []ServiceObservabilityMetricID{
		ServiceMetricActiveProviders,
		ServiceMetricActiveServices,
		ServiceMetricAverageInterfaceLookupLatency,
		ServiceMetricAverageServiceLookupLatency,
		ServiceMetricBlockSTMConflictRateServiceCalls,
		ServiceMetricCallsExecuted,
		ServiceMetricCallsExpired,
		ServiceMetricCallsFailed,
		ServiceMetricCallsSubmitted,
		ServiceMetricDisputesOpened,
		ServiceMetricDisputesResolved,
		ServiceMetricPaymentEscrowTotal,
		ServiceMetricProviderCollateralTotal,
		ServiceMetricReceiptProofGenerationLatency,
		ServiceMetricReceiptsAnchored,
		ServiceMetricRegisteredInterfaces,
		ServiceMetricServicesByTrustModel,
		ServiceMetricServicesByType,
	}
}
