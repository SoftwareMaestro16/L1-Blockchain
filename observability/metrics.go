package observability

import (
	"sync"
	"time"
)

const (
	MetricTelemetryEnabled     = "orbitalis_telemetry_enabled"
	MetricBlockHeight          = "orbitalis_block_height"
	MetricBlockTimeSeconds     = "orbitalis_block_time_seconds"
	MetricBlockProcessing      = "orbitalis_block_processing_seconds"
	MetricTxLatency            = "orbitalis_tx_latency_seconds"
	MetricModuleErrors         = "orbitalis_module_errors_total"
	MetricDexPoolCount         = "orbitalis_dex_pool_count"
	MetricDexLiquidityNorb     = "orbitalis_dex_liquidity_norb"
	MetricDexSwaps             = "orbitalis_dex_swaps_total"
	MetricFeesAccepted         = "orbitalis_fees_accepted_total"
	MetricFeesRejected         = "orbitalis_fees_rejected_total"
	MetricLocalnetHealth       = "orbitalis_localnet_health"
	MetricProcessUptimeSeconds = "orbitalis_process_uptime_seconds"
	MetricProcessMemoryBytes   = "orbitalis_process_memory_bytes"
	MetricProcessGoroutines    = "orbitalis_process_goroutines"
)

const (
	kindCounter = "counter"
	kindGauge   = "gauge"
	kindSummary = "summary"
)

type Definition struct {
	Name string
	Help string
	Type string
}

var Definitions = []Definition{
	{MetricTelemetryEnabled, "Whether Orbitalis process telemetry is enabled.", kindGauge},
	{MetricBlockHeight, "Last finalized block height observed by the app process.", kindGauge},
	{MetricBlockTimeSeconds, "Unix timestamp of the last finalized block time observed by the app process.", kindGauge},
	{MetricBlockProcessing, "FinalizeBlock processing duration observed by the app process.", kindSummary},
	{MetricTxLatency, "Approximate per-transaction FinalizeBlock processing latency.", kindSummary},
	{MetricModuleErrors, "Custom module errors counted with bounded labels.", kindCounter},
	{MetricDexPoolCount, "DEX pools observed by this process since startup.", kindGauge},
	{MetricDexLiquidityNorb, "DEX native norb liquidity observed by this process since startup.", kindGauge},
	{MetricDexSwaps, "Successful DEX swaps observed by this process.", kindCounter},
	{MetricFeesAccepted, "Transactions whose fees passed custom fee policy.", kindCounter},
	{MetricFeesRejected, "Transactions rejected by custom fee policy.", kindCounter},
	{MetricLocalnetHealth, "Localnet metrics endpoint health marker.", kindGauge},
	{MetricProcessUptimeSeconds, "Orbitalis process uptime in seconds.", kindGauge},
	{MetricProcessMemoryBytes, "Go runtime memory allocation bytes.", kindGauge},
	{MetricProcessGoroutines, "Go runtime goroutine count.", kindGauge},
}

type Labels map[string]string

type Registry struct {
	mu        sync.RWMutex
	enabled   bool
	startedAt time.Time
	counters  map[metricKey]sample
	gauges    map[metricKey]sample
	summaries map[metricKey]observation
}

type metricKey struct {
	name     string
	labelKey string
}

type sample struct {
	labels Labels
	value  float64
}

type observation struct {
	labels Labels
	count  uint64
	sum    float64
}

var DefaultRegistry = NewRegistry()
