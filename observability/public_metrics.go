package observability

import (
	"fmt"
	"sort"
)

const (
	RequiredMetricBlockTime			= "block_time"
	RequiredMetricFinalityLatency		= "finality_latency"
	RequiredMetricMissedBlocks		= "missed_blocks"
	RequiredMetricValidatorUptime		= "validator_uptime"
	RequiredMetricValidatorConcentration	= "validator_concentration"
	RequiredMetricTopNVotingPower		= "top_10_top_20_top_33_voting_power"
	RequiredMetricInflation			= "inflation"
	RequiredMetricBondedRatio		= "bonded_ratio"
	RequiredMetricEstimatedAPR		= "estimated_apr"
	RequiredMetricBurnedFees		= "burned_fees"
	RequiredMetricTreasuryBalance		= "treasury_balance"
	RequiredMetricSlashingEvents		= "slashing_events"
	RequiredMetricJailUnjailEvents		= "jail_unjail_events"
	RequiredMetricContractExecutionGas	= "contract_execution_gas"
	RequiredMetricFailedTxReasons		= "failed_tx_reasons"
	RequiredMetricNodeSyncStatus		= "node_sync_status"
)

const (
	RequiredSurfaceCLIQueries		= "cli_queries"
	RequiredSurfaceGRPCQueries		= "grpc_queries"
	RequiredSurfaceRESTQueries		= "rest_queries_where_applicable"
	RequiredSurfacePrometheusMetrics	= "prometheus_metrics"
	RequiredSurfaceIndexerEvents		= "explorer_indexer_compatibility_events"
	RequiredSurfacePublicDashboards		= "public_testnet_dashboards"
)

type PublicMetricSpec struct {
	ID			string
	PrometheusName		string
	Required		bool
	CLIQuery		bool
	GRPCQuery		bool
	RESTQuery		bool
	Prometheus		bool
	IndexerEvent		bool
	PublicDashboard		bool
	BoundedLabels		bool
	ExplorerCompatible	bool
	TestnetDashboardReady	bool
}

type PublicSurfaceSpec struct {
	ID		string
	Ready		bool
	Required	bool
}

type PublicMetricsReadinessReport struct {
	Metrics		[]PublicMetricSpec
	Surfaces	[]PublicSurfaceSpec
	RequiredCount	int
	ReadyCount	int
	SurfaceCount	int
	SurfacesReady	int
	PrometheusOnly	[]string
	Failed		[]string
	Ready		bool
}

func DefaultPublicMetricSpecs() []PublicMetricSpec {
	return []PublicMetricSpec{
		publicMetric(RequiredMetricBlockTime, MetricBlockTimeSeconds, true, true, true, true, true, true),
		publicMetric(RequiredMetricFinalityLatency, MetricFinalityLatencySeconds, true, true, true, true, true, true),
		publicMetric(RequiredMetricMissedBlocks, MetricValidatorMissedBlocks, true, true, true, true, true, true),
		publicMetric(RequiredMetricValidatorUptime, MetricValidatorUptimeBps, true, true, true, true, true, true),
		publicMetric(RequiredMetricValidatorConcentration, MetricValidatorConcentrationBps, true, true, true, true, true, true),
		publicMetric(RequiredMetricTopNVotingPower, MetricValidatorTopNPowerBps, true, true, true, true, true, true),
		publicMetric(RequiredMetricInflation, MetricEconomyInflationBps, true, true, true, true, true, true),
		publicMetric(RequiredMetricBondedRatio, MetricEconomyBondedRatioBps, true, true, true, true, true, true),
		publicMetric(RequiredMetricEstimatedAPR, MetricEconomyEstimatedAPRBps, true, true, true, true, true, true),
		publicMetric(RequiredMetricBurnedFees, MetricEconomyBurnedFeesNaet, true, true, true, true, true, true),
		publicMetric(RequiredMetricTreasuryBalance, MetricEconomyTreasuryBalanceNaet, true, true, true, true, true, true),
		publicMetric(RequiredMetricSlashingEvents, MetricSlashingEventsTotal, true, true, true, true, true, true),
		publicMetric(RequiredMetricJailUnjailEvents, MetricValidatorJailEventsTotal, true, true, true, true, true, true),
		publicMetric(RequiredMetricContractExecutionGas, MetricContractExecutionGas, true, true, true, true, true, true),
		publicMetric(RequiredMetricFailedTxReasons, MetricFailedTxReasons, true, true, true, true, true, true),
		publicMetric(RequiredMetricNodeSyncStatus, MetricNodeSyncStatus, true, true, true, true, true, true),
	}
}

func DefaultPublicSurfaceSpecs() []PublicSurfaceSpec {
	return []PublicSurfaceSpec{
		{ID: RequiredSurfaceCLIQueries, Ready: true, Required: true},
		{ID: RequiredSurfaceGRPCQueries, Ready: true, Required: true},
		{ID: RequiredSurfaceRESTQueries, Ready: true, Required: true},
		{ID: RequiredSurfacePrometheusMetrics, Ready: true, Required: true},
		{ID: RequiredSurfaceIndexerEvents, Ready: true, Required: true},
		{ID: RequiredSurfacePublicDashboards, Ready: true, Required: true},
	}
}

func ValidatePublicMetricsReadiness(metrics []PublicMetricSpec, surfaces []PublicSurfaceSpec) error {
	report := BuildPublicMetricsReadinessReport(metrics, surfaces)
	if !report.Ready {
		return fmt.Errorf("public metrics readiness failed: %v", report.Failed)
	}
	return nil
}

func BuildPublicMetricsReadinessReport(metrics []PublicMetricSpec, surfaces []PublicSurfaceSpec) PublicMetricsReadinessReport {
	if metrics == nil {
		metrics = DefaultPublicMetricSpecs()
	}
	if surfaces == nil {
		surfaces = DefaultPublicSurfaceSpecs()
	}
	metrics = normalizePublicMetrics(metrics)
	surfaces = normalizePublicSurfaces(surfaces)
	requiredMetrics := requiredPublicMetricIDs()
	requiredSurfaces := requiredPublicSurfaceIDs()
	knownPrometheus := prometheusDefinitionNames()
	seenMetrics := map[string]PublicMetricSpec{}
	seenSurfaces := map[string]PublicSurfaceSpec{}
	failed := make([]string, 0)
	prometheusOnly := make([]string, 0)
	requiredCount := 0
	readyCount := 0
	surfaceCount := 0
	surfacesReady := 0

	for _, metric := range metrics {
		if metric.ID == "" {
			failed = append(failed, "metric_id_required")
			continue
		}
		if _, duplicate := seenMetrics[metric.ID]; duplicate {
			failed = append(failed, metric.ID+":duplicate_metric")
		}
		seenMetrics[metric.ID] = metric
		if !requiredMetrics[metric.ID] {
			failed = append(failed, metric.ID+":unknown_metric")
		}
		if metric.Required {
			requiredCount++
		}
		if metric.Required && !metricReady(metric, knownPrometheus) {
			failed = append(failed, metric.ID+":missing_required_surface")
		}
		if metric.Required && metricReady(metric, knownPrometheus) {
			readyCount++
		}
		if metric.Prometheus && !(metric.CLIQuery || metric.GRPCQuery || metric.RESTQuery || metric.IndexerEvent || metric.PublicDashboard) {
			prometheusOnly = append(prometheusOnly, metric.ID)
		}
	}
	for id := range requiredMetrics {
		if _, ok := seenMetrics[id]; !ok {
			failed = append(failed, id+":missing_metric")
		}
	}

	for _, surface := range surfaces {
		if surface.ID == "" {
			failed = append(failed, "surface_id_required")
			continue
		}
		if _, duplicate := seenSurfaces[surface.ID]; duplicate {
			failed = append(failed, surface.ID+":duplicate_surface")
		}
		seenSurfaces[surface.ID] = surface
		if !requiredSurfaces[surface.ID] {
			failed = append(failed, surface.ID+":unknown_surface")
		}
		if surface.Required {
			surfaceCount++
		}
		if surface.Required && !surface.Ready {
			failed = append(failed, surface.ID+":surface_not_ready")
		}
		if surface.Required && surface.Ready {
			surfacesReady++
		}
	}
	for id := range requiredSurfaces {
		if _, ok := seenSurfaces[id]; !ok {
			failed = append(failed, id+":missing_surface")
		}
	}

	sort.Strings(failed)
	sort.Strings(prometheusOnly)
	return PublicMetricsReadinessReport{
		Metrics:	metrics,
		Surfaces:	surfaces,
		RequiredCount:	requiredCount,
		ReadyCount:	readyCount,
		SurfaceCount:	surfaceCount,
		SurfacesReady:	surfacesReady,
		PrometheusOnly:	prometheusOnly,
		Failed:		failed,
		Ready:		len(failed) == 0 && len(prometheusOnly) == 0,
	}
}

func publicMetric(id, prometheusName string, cli, grpc, rest, prometheus, indexer, dashboard bool) PublicMetricSpec {
	return PublicMetricSpec{
		ID:			id,
		PrometheusName:		prometheusName,
		Required:		true,
		CLIQuery:		cli,
		GRPCQuery:		grpc,
		RESTQuery:		rest,
		Prometheus:		prometheus,
		IndexerEvent:		indexer,
		PublicDashboard:	dashboard,
		BoundedLabels:		true,
		ExplorerCompatible:	true,
		TestnetDashboardReady:	true,
	}
}

func metricReady(metric PublicMetricSpec, knownPrometheus map[string]bool) bool {
	if metric.PrometheusName == "" || !knownPrometheus[metric.PrometheusName] {
		return false
	}
	return metric.CLIQuery &&
		metric.GRPCQuery &&
		metric.RESTQuery &&
		metric.Prometheus &&
		metric.IndexerEvent &&
		metric.PublicDashboard &&
		metric.BoundedLabels &&
		metric.ExplorerCompatible &&
		metric.TestnetDashboardReady
}

func normalizePublicMetrics(metrics []PublicMetricSpec) []PublicMetricSpec {
	out := append([]PublicMetricSpec{}, metrics...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func normalizePublicSurfaces(surfaces []PublicSurfaceSpec) []PublicSurfaceSpec {
	out := append([]PublicSurfaceSpec{}, surfaces...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func prometheusDefinitionNames() map[string]bool {
	out := make(map[string]bool, len(Definitions))
	for _, def := range Definitions {
		out[def.Name] = true
	}
	return out
}

func requiredPublicMetricIDs() map[string]bool {
	return map[string]bool{
		RequiredMetricBlockTime:		true,
		RequiredMetricFinalityLatency:		true,
		RequiredMetricMissedBlocks:		true,
		RequiredMetricValidatorUptime:		true,
		RequiredMetricValidatorConcentration:	true,
		RequiredMetricTopNVotingPower:		true,
		RequiredMetricInflation:		true,
		RequiredMetricBondedRatio:		true,
		RequiredMetricEstimatedAPR:		true,
		RequiredMetricBurnedFees:		true,
		RequiredMetricTreasuryBalance:		true,
		RequiredMetricSlashingEvents:		true,
		RequiredMetricJailUnjailEvents:		true,
		RequiredMetricContractExecutionGas:	true,
		RequiredMetricFailedTxReasons:		true,
		RequiredMetricNodeSyncStatus:		true,
	}
}

func requiredPublicSurfaceIDs() map[string]bool {
	return map[string]bool{
		RequiredSurfaceCLIQueries:		true,
		RequiredSurfaceGRPCQueries:		true,
		RequiredSurfaceRESTQueries:		true,
		RequiredSurfacePrometheusMetrics:	true,
		RequiredSurfaceIndexerEvents:		true,
		RequiredSurfacePublicDashboards:	true,
	}
}
