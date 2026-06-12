package observability

import (
	"sync"
	"time"
)

const (
	MetricTelemetryEnabled			= "aetra_telemetry_enabled"
	MetricBlockHeight			= "aetra_block_height"
	MetricBlockTimeSeconds			= "aetra_block_time_seconds"
	MetricFinalityLatencySeconds		= "aetra_finality_latency_seconds"
	MetricBlockProcessing			= "aetra_block_processing_seconds"
	MetricTxLatency				= "aetra_tx_latency_seconds"
	MetricModuleErrors			= "aetra_module_errors_total"
	MetricFailedTxReasons			= "aetra_failed_tx_reasons_total"
	MetricFeesAccepted			= "aetra_fees_accepted_total"
	MetricFeesRejected			= "aetra_fees_rejected_total"
	MetricEconomyInflationBps		= "aetra_economy_inflation_bps"
	MetricEconomyBondedRatioBps		= "aetra_economy_bonded_ratio_bps"
	MetricEconomyEstimatedAPRBps		= "aetra_economy_estimated_apr_bps"
	MetricEconomyBurnRatioBps		= "aetra_economy_burn_ratio_bps"
	MetricEconomyValidatorFeeRatioBps	= "aetra_economy_validator_fee_ratio_bps"
	MetricEconomyDeflationGuard		= "aetra_economy_deflation_guard"
	MetricEconomyQueueLimited		= "aetra_economy_queue_limited"
	MetricEconomyRateLimited		= "aetra_economy_rate_limited"
	MetricEconomyTotalChargesNaet		= "aetra_economy_total_charges_naet"
	MetricEconomyBurnNaet			= "aetra_economy_burn_naet"
	MetricEconomyBurnedFeesNaet		= "aetra_economy_burned_fees_naet"
	MetricEconomyTreasuryNaet		= "aetra_economy_treasury_naet"
	MetricEconomyTreasuryBalanceNaet	= "aetra_economy_treasury_balance_naet"
	MetricEconomyValidatorRewardsNaet	= "aetra_economy_validator_rewards_naet"
	MetricEconomyOptimalState		= "aetra_economy_optimal_state"
	MetricEconomyFailedConditions		= "aetra_economy_failed_conditions"
	MetricEconomyInvariantsSatisfied	= "aetra_economy_invariants_satisfied"
	MetricEconomyInvariantFailures		= "aetra_economy_invariant_failures"
	MetricEconomyWeaknessControlsReady	= "aetra_economy_weakness_controls_ready"
	MetricEconomyMissingControls		= "aetra_economy_missing_controls"
	MetricEconomyInflationRiskCount		= "aetra_economy_inflation_risk_count"
	MetricEconomyCircuitBreakerActive	= "aetra_economy_circuit_breaker_active"
	MetricEconomyCircuitBreakerReasons	= "aetra_economy_circuit_breaker_reasons"
	MetricValidatorIncentivesHealthy	= "aetra_validator_incentives_healthy"
	MetricValidatorIncentiveFindings	= "aetra_validator_incentive_findings"
	MetricStakingCentralizationHealthy	= "aetra_staking_centralization_healthy"
	MetricStakingCentralizationRisks	= "aetra_staking_centralization_risks"
	MetricFeeModelEfficiencyHealthy		= "aetra_fee_model_efficiency_healthy"
	MetricFeeModelEfficiencyRisks		= "aetra_fee_model_efficiency_risks"
	MetricValidatorRewardPerPowerNaet	= "aetra_validator_reward_per_power_naet"
	MetricValidatorProfitabilityBps		= "aetra_validator_profitability_bps"
	MetricSlashingPenaltyNaet		= "aetra_slashing_penalty_naet"
	MetricSlashingEventsTotal		= "aetra_slashing_events_total"
	MetricValidatorJailEventsTotal		= "aetra_validator_jail_events_total"
	MetricValidatorUnjailEventsTotal	= "aetra_validator_unjail_events_total"
	MetricSlashingBurnNaet			= "aetra_slashing_burn_naet"
	MetricSlashingTreasuryNaet		= "aetra_slashing_treasury_naet"
	MetricSlashingReporterNaet		= "aetra_slashing_reporter_naet"
	MetricValidatorMissedBlocks		= "aetra_validator_missed_blocks_total"
	MetricValidatorUptimeBps		= "aetra_validator_uptime_bps"
	MetricValidatorConcentrationBps		= "aetra_validator_concentration_bps"
	MetricValidatorTopNPowerBps		= "aetra_validator_top_n_power_bps"
	MetricValidatorConcentrationRisks	= "aetra_validator_concentration_risks"
	MetricContractExecutionGas		= "aetra_contract_execution_gas"
	MetricNodeSyncStatus			= "aetra_node_sync_status"
	MetricLocalnetHealth			= "aetra_localnet_health"
	MetricProcessUptimeSeconds		= "aetra_process_uptime_seconds"
	MetricProcessMemoryBytes		= "aetra_process_memory_bytes"
	MetricProcessGoroutines			= "aetra_process_goroutines"
)

const (
	kindCounter	= "counter"
	kindGauge	= "gauge"
	kindSummary	= "summary"
)

type Definition struct {
	Name	string
	Help	string
	Type	string
}

var Definitions = []Definition{
	{MetricTelemetryEnabled, "Whether Aetra process telemetry is enabled.", kindGauge},
	{MetricBlockHeight, "Last finalized block height observed by the app process.", kindGauge},
	{MetricBlockTimeSeconds, "Unix timestamp of the last finalized block time observed by the app process.", kindGauge},
	{MetricFinalityLatencySeconds, "Observed block finality latency from proposal time to commit.", kindSummary},
	{MetricBlockProcessing, "FinalizeBlock processing duration observed by the app process.", kindSummary},
	{MetricTxLatency, "Approximate per-transaction FinalizeBlock processing latency.", kindSummary},
	{MetricModuleErrors, "Custom module errors counted with bounded labels.", kindCounter},
	{MetricFailedTxReasons, "Failed transaction reasons counted with bounded reason labels.", kindCounter},
	{MetricFeesAccepted, "Transactions whose fees passed custom fee policy.", kindCounter},
	{MetricFeesRejected, "Transactions rejected by custom fee policy.", kindCounter},
	{MetricEconomyInflationBps, "Last economic controller inflation output in basis points.", kindGauge},
	{MetricEconomyBondedRatioBps, "Last bonded stake ratio in basis points.", kindGauge},
	{MetricEconomyEstimatedAPRBps, "Last estimated staking APR in basis points.", kindGauge},
	{MetricEconomyBurnRatioBps, "Last economic controller burn ratio output in basis points.", kindGauge},
	{MetricEconomyValidatorFeeRatioBps, "Last economic controller validator fee ratio output in basis points.", kindGauge},
	{MetricEconomyDeflationGuard, "Whether the deflation guard was active in the last economic controller output.", kindGauge},
	{MetricEconomyQueueLimited, "Whether queue limiting was active in the last economic controller output.", kindGauge},
	{MetricEconomyRateLimited, "Whether rate limiting was active in the last economic controller output.", kindGauge},
	{MetricEconomyTotalChargesNaet, "Last protocol economic flow total charges in naet.", kindGauge},
	{MetricEconomyBurnNaet, "Last protocol economic flow burn amount in naet.", kindGauge},
	{MetricEconomyBurnedFeesNaet, "Total fee amount routed to burn in naet.", kindGauge},
	{MetricEconomyTreasuryNaet, "Last protocol economic flow treasury amount in naet.", kindGauge},
	{MetricEconomyTreasuryBalanceNaet, "Protocol treasury balance in naet.", kindGauge},
	{MetricEconomyValidatorRewardsNaet, "Last protocol economic flow validator reward amount in naet.", kindGauge},
	{MetricEconomyOptimalState, "Whether the last evaluated economic state met all optimality conditions.", kindGauge},
	{MetricEconomyFailedConditions, "Number of failed optimal economic state conditions in the last evaluation.", kindGauge},
	{MetricEconomyInvariantsSatisfied, "Whether the last evaluated economic invariant set passed.", kindGauge},
	{MetricEconomyInvariantFailures, "Number of failed economic invariants in the last evaluation.", kindGauge},
	{MetricEconomyWeaknessControlsReady, "Whether all known economic weakness controls are production ready.", kindGauge},
	{MetricEconomyMissingControls, "Number of missing economic weakness controls in the last evaluation.", kindGauge},
	{MetricEconomyInflationRiskCount, "Number of inflation model risks in the last evaluation.", kindGauge},
	{MetricEconomyCircuitBreakerActive, "Whether the economic circuit breaker is active.", kindGauge},
	{MetricEconomyCircuitBreakerReasons, "Number of active economic circuit breaker reasons.", kindGauge},
	{MetricValidatorIncentivesHealthy, "Whether the last validator incentive evaluation found no weaknesses.", kindGauge},
	{MetricValidatorIncentiveFindings, "Number of validator incentive findings in the last evaluation.", kindGauge},
	{MetricStakingCentralizationHealthy, "Whether the last staking centralization evaluation found no risks.", kindGauge},
	{MetricStakingCentralizationRisks, "Number of staking centralization risks in the last evaluation.", kindGauge},
	{MetricFeeModelEfficiencyHealthy, "Whether the last fee model efficiency evaluation found no risks.", kindGauge},
	{MetricFeeModelEfficiencyRisks, "Number of fee model efficiency risks in the last evaluation.", kindGauge},
	{MetricValidatorRewardPerPowerNaet, "Validator reward per unit of voting power in naet, labeled by validator state.", kindGauge},
	{MetricValidatorProfitabilityBps, "Validator profitability margin in basis points, labeled by validator state.", kindGauge},
	{MetricSlashingPenaltyNaet, "Last slashing penalty amount in naet, labeled by bounded reason.", kindGauge},
	{MetricSlashingEventsTotal, "Slashing events counted with bounded reason labels.", kindCounter},
	{MetricValidatorJailEventsTotal, "Validator jail events counted with bounded reason labels.", kindCounter},
	{MetricValidatorUnjailEventsTotal, "Validator unjail events counted with bounded reason labels.", kindCounter},
	{MetricSlashingBurnNaet, "Last slashing burn routing amount in naet, labeled by bounded reason.", kindGauge},
	{MetricSlashingTreasuryNaet, "Last slashing treasury routing amount in naet, labeled by bounded reason.", kindGauge},
	{MetricSlashingReporterNaet, "Last slashing reporter reward amount in naet, labeled by bounded reason.", kindGauge},
	{MetricValidatorMissedBlocks, "Validator missed blocks counted with bounded validator state labels.", kindCounter},
	{MetricValidatorUptimeBps, "Validator uptime in basis points over the configured scoring window.", kindGauge},
	{MetricValidatorConcentrationBps, "Validator voting power concentration in basis points.", kindGauge},
	{MetricValidatorTopNPowerBps, "Last validator active-set top-N voting power share in basis points.", kindGauge},
	{MetricValidatorConcentrationRisks, "Number of validator concentration warnings in the last report.", kindGauge},
	{MetricContractExecutionGas, "Contract execution gas observed by VM and contract result labels.", kindSummary},
	{MetricNodeSyncStatus, "Node sync status where 1 means catching up and 0 means caught up.", kindGauge},
	{MetricLocalnetHealth, "Localnet metrics endpoint health marker.", kindGauge},
	{MetricProcessUptimeSeconds, "Aetra process uptime in seconds.", kindGauge},
	{MetricProcessMemoryBytes, "Go runtime memory allocation bytes.", kindGauge},
	{MetricProcessGoroutines, "Go runtime goroutine count.", kindGauge},
}

type Labels map[string]string

type Registry struct {
	mu		sync.RWMutex
	enabled		bool
	startedAt	time.Time
	counters	map[metricKey]sample
	gauges		map[metricKey]sample
	summaries	map[metricKey]observation
}

type metricKey struct {
	name		string
	labelKey	string
}

type sample struct {
	labels	Labels
	value	float64
}

type observation struct {
	labels	Labels
	count	uint64
	sum	float64
}

var DefaultRegistry = NewRegistry()
