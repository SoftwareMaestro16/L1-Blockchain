package observability

import (
	"sync"
	"time"
)

const (
	MetricTelemetryEnabled            = "aetheris_telemetry_enabled"
	MetricBlockHeight                 = "aetheris_block_height"
	MetricBlockTimeSeconds            = "aetheris_block_time_seconds"
	MetricBlockProcessing             = "aetheris_block_processing_seconds"
	MetricTxLatency                   = "aetheris_tx_latency_seconds"
	MetricModuleErrors                = "aetheris_module_errors_total"
	MetricDexPoolCount                = "aetheris_dex_pool_count"
	MetricDexLiquidityNaet            = "aetheris_dex_liquidity_naet"
	MetricDexSwaps                    = "aetheris_dex_swaps_total"
	MetricFeesAccepted                = "aetheris_fees_accepted_total"
	MetricFeesRejected                = "aetheris_fees_rejected_total"
	MetricEconomyInflationBps         = "aetheris_economy_inflation_bps"
	MetricEconomyBurnRatioBps         = "aetheris_economy_burn_ratio_bps"
	MetricEconomyValidatorFeeRatioBps = "aetheris_economy_validator_fee_ratio_bps"
	MetricEconomyDeflationGuard       = "aetheris_economy_deflation_guard"
	MetricEconomyQueueLimited         = "aetheris_economy_queue_limited"
	MetricEconomyRateLimited          = "aetheris_economy_rate_limited"
	MetricEconomyTotalChargesNaet     = "aetheris_economy_total_charges_naet"
	MetricEconomyBurnNaet             = "aetheris_economy_burn_naet"
	MetricEconomyTreasuryNaet         = "aetheris_economy_treasury_naet"
	MetricEconomyValidatorRewardsNaet = "aetheris_economy_validator_rewards_naet"
	MetricEconomyOptimalState         = "aetheris_economy_optimal_state"
	MetricEconomyFailedConditions     = "aetheris_economy_failed_conditions"
	MetricLocalnetHealth              = "aetheris_localnet_health"
	MetricProcessUptimeSeconds        = "aetheris_process_uptime_seconds"
	MetricProcessMemoryBytes          = "aetheris_process_memory_bytes"
	MetricProcessGoroutines           = "aetheris_process_goroutines"
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
	{MetricTelemetryEnabled, "Whether Aetheris process telemetry is enabled.", kindGauge},
	{MetricBlockHeight, "Last finalized block height observed by the app process.", kindGauge},
	{MetricBlockTimeSeconds, "Unix timestamp of the last finalized block time observed by the app process.", kindGauge},
	{MetricBlockProcessing, "FinalizeBlock processing duration observed by the app process.", kindSummary},
	{MetricTxLatency, "Approximate per-transaction FinalizeBlock processing latency.", kindSummary},
	{MetricModuleErrors, "Custom module errors counted with bounded labels.", kindCounter},
	{MetricDexPoolCount, "DEX pools observed by this process since startup.", kindGauge},
	{MetricDexLiquidityNaet, "DEX native naet liquidity observed by this process since startup.", kindGauge},
	{MetricDexSwaps, "Successful DEX swaps observed by this process.", kindCounter},
	{MetricFeesAccepted, "Transactions whose fees passed custom fee policy.", kindCounter},
	{MetricFeesRejected, "Transactions rejected by custom fee policy.", kindCounter},
	{MetricEconomyInflationBps, "Last economic controller inflation output in basis points.", kindGauge},
	{MetricEconomyBurnRatioBps, "Last economic controller burn ratio output in basis points.", kindGauge},
	{MetricEconomyValidatorFeeRatioBps, "Last economic controller validator fee ratio output in basis points.", kindGauge},
	{MetricEconomyDeflationGuard, "Whether the deflation guard was active in the last economic controller output.", kindGauge},
	{MetricEconomyQueueLimited, "Whether queue limiting was active in the last economic controller output.", kindGauge},
	{MetricEconomyRateLimited, "Whether rate limiting was active in the last economic controller output.", kindGauge},
	{MetricEconomyTotalChargesNaet, "Last protocol economic flow total charges in naet.", kindGauge},
	{MetricEconomyBurnNaet, "Last protocol economic flow burn amount in naet.", kindGauge},
	{MetricEconomyTreasuryNaet, "Last protocol economic flow treasury amount in naet.", kindGauge},
	{MetricEconomyValidatorRewardsNaet, "Last protocol economic flow validator reward amount in naet.", kindGauge},
	{MetricEconomyOptimalState, "Whether the last evaluated economic state met all optimality conditions.", kindGauge},
	{MetricEconomyFailedConditions, "Number of failed optimal economic state conditions in the last evaluation.", kindGauge},
	{MetricLocalnetHealth, "Localnet metrics endpoint health marker.", kindGauge},
	{MetricProcessUptimeSeconds, "Aetheris process uptime in seconds.", kindGauge},
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
