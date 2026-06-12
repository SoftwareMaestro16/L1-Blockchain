package params

import (
	"fmt"
	"sort"
	"strings"
)

const (
	EconomicObservabilityKindMetric	= "metric"
	EconomicObservabilityKindEvent	= "event"
	EconomicObservabilityKindQuery	= "query"

	EconomicMetricCurrentInflationRate		= "current_inflation_rate"
	EconomicMetricGrossMintedPerEpoch		= "gross_minted_aet_per_epoch"
	EconomicMetricBurnedPerEpoch			= "burned_aet_per_epoch"
	EconomicMetricNetSupplyChangePerEpoch		= "net_supply_change_per_epoch"
	EconomicMetricTotalBondedStakeRatio		= "total_bonded_stake_ratio"
	EconomicMetricActiveValidatorCount		= "active_validator_count"
	EconomicMetricStandbyValidatorCount		= "standby_validator_count"
	EconomicMetricTopNVotingPowerShare		= "top_n_voting_power_share"
	EconomicMetricValidatorRewardPerVotingPower	= "per_validator_reward_per_voting_power"
	EconomicMetricFeeRevenueByBucket		= "fee_revenue_by_bucket"
	EconomicMetricBaseFeeCongestionMultiplier	= "base_fee_and_congestion_multiplier"
	EconomicMetricFailedTxSurchargeTotals		= "failed_transaction_surcharge_totals"
	EconomicMetricStateBytesDelta			= "state_bytes_added_removed_net"
	EconomicMetricStorageRentWarnings		= "storage_rent_balances_and_exhaustion_warnings"
	EconomicMetricSlashingBySeverity		= "slashing_amount_by_severity"
	EconomicMetricReporterRewardsPaid		= "reporter_rewards_paid"
	EconomicMetricDeflationGuardActivations		= "deflation_guard_activation_count"
	EconomicMetricCircuitBreakerActivations		= "circuit_breaker_activation_count"

	EconomicEventInflationUpdate			= "inflation_update_event"
	EconomicEventFeeAllocation			= "fee_allocation_event"
	EconomicEventBurn				= "burn_event"
	EconomicEventDeflationGuard			= "deflation_guard_event"
	EconomicEventValidatorScoreUpdate		= "validator_score_update_event"
	EconomicEventValidatorConcentrationWarning	= "validator_concentration_warning_event"
	EconomicEventCommissionChangeWarning		= "commission_change_warning_event"
	EconomicEventDelegationRiskWarning		= "delegation_risk_warning_event"
	EconomicEventSlashingRoute			= "slashing_route_event"
	EconomicEventReporterReward			= "reporter_reward_event"
	EconomicEventStorageFee				= "storage_fee_event"
	EconomicEventStateRentWarning			= "state_rent_warning_event"
	EconomicEventCircuitBreaker			= "circuit_breaker_event"

	EconomicQueryCurrentParameters			= "current_economic_parameters"
	EconomicQueryInflationState			= "current_and_historical_inflation_state"
	EconomicQueryNetIssuanceByEpoch			= "net_issuance_by_epoch"
	EconomicQueryFeeDistributionByEpoch		= "fee_distribution_by_epoch"
	EconomicQueryBurnHistory			= "burn_history"
	EconomicQueryValidatorScoreComponents		= "validator_score_and_score_components"
	EconomicQueryValidatorConcentration		= "validator_concentration_metrics"
	EconomicQueryDelegatorRiskAdjustedYield		= "delegator_risk_adjusted_yield_estimate"
	EconomicQueryStorageFootprint			= "storage_footprint_by_account_or_contract"
	EconomicQueryStateRentStatus			= "state_rent_status"
	EconomicQueryFeeEstimateTxClass			= "fee_estimate_for_transaction_class"
	EconomicQuerySupplyProjectionCurrentParams	= "supply_projection_under_current_parameters"
)

type EconomicObservabilitySignal struct {
	ID			string
	Kind			string
	Scope			string
	Source			string
	Required		bool
	Queryable		bool
	TelemetryEnabled	bool
	Emitted			bool
	SchemaVersion		uint32
	Labels			[]string
}

type EconomicObservabilityReport struct {
	Metrics			[]EconomicObservabilitySignal
	Events			[]EconomicObservabilitySignal
	Queries			[]EconomicObservabilitySignal
	RequiredMetrics		int
	RequiredEvents		int
	RequiredQueries		int
	CoveredMetrics		int
	CoveredEvents		int
	CoveredQueries		int
	MetricCoverageBps	int64
	EventCoverageBps	int64
	QueryCoverageBps	int64
	Passed			bool
	Failed			[]string
	GovernanceSummary	string
}

func DefaultEconomicObservabilityMetrics() []EconomicObservabilitySignal {
	return []EconomicObservabilitySignal{
		requiredMetric(EconomicMetricCurrentInflationRate, "epoch", "adaptive_inflation.controller_state", "epoch_id"),
		requiredMetric(EconomicMetricGrossMintedPerEpoch, "epoch", "adaptive_inflation.net_issuance", "epoch_id"),
		requiredMetric(EconomicMetricBurnedPerEpoch, "epoch", "burn_deflation.accounting", "epoch_id", "source"),
		requiredMetric(EconomicMetricNetSupplyChangePerEpoch, "epoch", "adaptive_inflation.net_issuance", "epoch_id"),
		requiredMetric(EconomicMetricTotalBondedStakeRatio, "epoch", "staking_enhancements.score_inputs", "epoch_id"),
		requiredMetric(EconomicMetricActiveValidatorCount, "epoch", "staking_enhancements.active_set", "epoch_id"),
		requiredMetric(EconomicMetricStandbyValidatorCount, "epoch", "staking_enhancements.standby_set", "epoch_id"),
		requiredMetric(EconomicMetricTopNVotingPowerShare, "epoch", "staking_enhancements.concentration_metrics", "epoch_id", "top_n"),
		requiredMetric(EconomicMetricValidatorRewardPerVotingPower, "validator_epoch", "implementation_sequencing.reward_per_power", "epoch_id", "validator_id"),
		requiredMetric(EconomicMetricFeeRevenueByBucket, "epoch", "fee_market_optimizer.allocation", "epoch_id", "bucket"),
		requiredMetric(EconomicMetricBaseFeeCongestionMultiplier, "block", "fee_market_optimizer.base_fee_resource_multipliers", "block_height"),
		requiredMetric(EconomicMetricFailedTxSurchargeTotals, "block", "fee_market_optimizer.sender_surcharge", "block_height", "sender"),
		requiredMetric(EconomicMetricStateBytesDelta, "block_epoch", "execution_state_economy.state_growth_telemetry", "epoch_id", "block_height"),
		requiredMetric(EconomicMetricStorageRentWarnings, "account_contract_epoch", "storage_economy.rent_status", "epoch_id", "owner_id"),
		requiredMetric(EconomicMetricSlashingBySeverity, "epoch", "economic_security.penalty_routing", "epoch_id", "severity"),
		requiredMetric(EconomicMetricReporterRewardsPaid, "epoch", "economic_security.reporter_rewards", "epoch_id", "reporter"),
		requiredMetric(EconomicMetricDeflationGuardActivations, "epoch", "burn_deflation.deflation_guard", "epoch_id", "reason"),
		requiredMetric(EconomicMetricCircuitBreakerActivations, "epoch", "economic_security.circuit_breaker", "epoch_id", "reason"),
	}
}

func DefaultEconomicObservabilityEvents() []EconomicObservabilitySignal {
	return []EconomicObservabilitySignal{
		requiredEvent(EconomicEventInflationUpdate, "epoch", "adaptive_inflation.events", "epoch_id"),
		requiredEvent(EconomicEventFeeAllocation, "block", "fee_market_optimizer.events", "epoch_id", "block_height", "bucket"),
		requiredEvent(EconomicEventBurn, "block", "burn_deflation.events", "epoch_id", "block_height", "source"),
		requiredEvent(EconomicEventDeflationGuard, "block_epoch", "burn_deflation.deflation_guard", "epoch_id", "reason"),
		requiredEvent(EconomicEventValidatorScoreUpdate, "epoch_validator", "staking_enhancements.validator_events", "epoch_id", "validator_id"),
		requiredEvent(EconomicEventValidatorConcentrationWarning, "epoch_validator", "staking_enhancements.concentration_events", "epoch_id", "validator_id"),
		requiredEvent(EconomicEventCommissionChangeWarning, "validator", "validator_reputation.capture_risk", "validator_id"),
		requiredEvent(EconomicEventDelegationRiskWarning, "delegation", "validator_reputation.delegation_risk", "validator_id"),
		requiredEvent(EconomicEventSlashingRoute, "block", "economic_security.penalty_routing", "epoch_id", "validator_id"),
		requiredEvent(EconomicEventReporterReward, "block", "economic_security.reporter_reward", "epoch_id", "reporter"),
		requiredEvent(EconomicEventStorageFee, "block", "storage_economy.fee_events", "block_height", "owner_id"),
		requiredEvent(EconomicEventStateRentWarning, "epoch", "storage_economy.rent_status", "epoch_id", "owner_id"),
		requiredEvent(EconomicEventCircuitBreaker, "block_epoch", "economic_security.circuit_breaker", "epoch_id", "reason"),
	}
}

func DefaultEconomicObservabilityQueries() []EconomicObservabilitySignal {
	return []EconomicObservabilitySignal{
		requiredQuery(EconomicQueryCurrentParameters, "governance", "params.economic_control_surface", "module"),
		requiredQuery(EconomicQueryInflationState, "epoch", "adaptive_inflation.controller_state", "epoch_id"),
		requiredQuery(EconomicQueryNetIssuanceByEpoch, "epoch", "adaptive_inflation.net_issuance", "epoch_id"),
		requiredQuery(EconomicQueryFeeDistributionByEpoch, "epoch", "fee_market_optimizer.allocation", "epoch_id", "bucket"),
		requiredQuery(EconomicQueryBurnHistory, "epoch_window", "burn_deflation.supply_query", "from_epoch", "to_epoch"),
		requiredQuery(EconomicQueryValidatorScoreComponents, "validator_epoch", "staking_enhancements.score", "epoch_id", "validator_id"),
		requiredQuery(EconomicQueryValidatorConcentration, "epoch", "staking_enhancements.concentration_metrics", "epoch_id", "top_n"),
		requiredQuery(EconomicQueryDelegatorRiskAdjustedYield, "delegator_validator_epoch", "validator_reputation.yield_estimate", "epoch_id", "delegator", "validator_id"),
		requiredQuery(EconomicQueryStorageFootprint, "account_contract", "storage_economy.footprint", "owner_id"),
		requiredQuery(EconomicQueryStateRentStatus, "account_contract_epoch", "storage_economy.rent_status", "epoch_id", "owner_id"),
		requiredQuery(EconomicQueryFeeEstimateTxClass, "transaction_class", "fee_market_optimizer.estimate", "tx_class"),
		requiredQuery(EconomicQuerySupplyProjectionCurrentParams, "governance_projection", "supply_stabilization.projection", "projection_years"),
	}
}

func BuildEconomicObservabilityReport(metrics, events []EconomicObservabilitySignal, queriesInput ...[]EconomicObservabilitySignal) EconomicObservabilityReport {
	if metrics == nil {
		metrics = DefaultEconomicObservabilityMetrics()
	}
	if events == nil {
		events = DefaultEconomicObservabilityEvents()
	}
	queries := DefaultEconomicObservabilityQueries()
	if len(queriesInput) > 0 && queriesInput[0] != nil {
		queries = queriesInput[0]
	}
	metricSignals, metricFailed, requiredMetrics, coveredMetrics := evaluateObservabilitySignals(EconomicObservabilityKindMetric, metrics, requiredEconomicMetricIDs())
	eventSignals, eventFailed, requiredEvents, coveredEvents := evaluateObservabilitySignals(EconomicObservabilityKindEvent, events, requiredEconomicEventIDs())
	querySignals, queryFailed, requiredQueries, coveredQueries := evaluateObservabilitySignals(EconomicObservabilityKindQuery, queries, requiredEconomicQueryIDs())
	failed := append(metricFailed, eventFailed...)
	failed = append(failed, queryFailed...)
	sort.Strings(failed)

	metricCoverage := coverageBps(coveredMetrics, requiredMetrics)
	eventCoverage := coverageBps(coveredEvents, requiredEvents)
	queryCoverage := coverageBps(coveredQueries, requiredQueries)
	return EconomicObservabilityReport{
		Metrics:		metricSignals,
		Events:			eventSignals,
		Queries:		querySignals,
		RequiredMetrics:	requiredMetrics,
		RequiredEvents:		requiredEvents,
		RequiredQueries:	requiredQueries,
		CoveredMetrics:		coveredMetrics,
		CoveredEvents:		coveredEvents,
		CoveredQueries:		coveredQueries,
		MetricCoverageBps:	metricCoverage,
		EventCoverageBps:	eventCoverage,
		QueryCoverageBps:	queryCoverage,
		Passed:			len(failed) == 0 && metricCoverage == BasisPoints && eventCoverage == BasisPoints && queryCoverage == BasisPoints,
		Failed:			failed,
		GovernanceSummary:	fmt.Sprintf("required_metrics=%d/%d required_events=%d/%d required_queries=%d/%d metric_coverage_bps=%d event_coverage_bps=%d query_coverage_bps=%d", coveredMetrics, requiredMetrics, coveredEvents, requiredEvents, coveredQueries, requiredQueries, metricCoverage, eventCoverage, queryCoverage),
	}
}

func requiredMetric(id, scope, source string, labels ...string) EconomicObservabilitySignal {
	return EconomicObservabilitySignal{
		ID:			id,
		Kind:			EconomicObservabilityKindMetric,
		Scope:			scope,
		Source:			source,
		Required:		true,
		Queryable:		true,
		TelemetryEnabled:	true,
		SchemaVersion:		1,
		Labels:			append([]string{}, labels...),
	}
}

func requiredEvent(id, scope, source string, labels ...string) EconomicObservabilitySignal {
	return EconomicObservabilitySignal{
		ID:			id,
		Kind:			EconomicObservabilityKindEvent,
		Scope:			scope,
		Source:			source,
		Required:		true,
		TelemetryEnabled:	true,
		Emitted:		true,
		SchemaVersion:		1,
		Labels:			append([]string{}, labels...),
	}
}

func requiredQuery(id, scope, source string, labels ...string) EconomicObservabilitySignal {
	return EconomicObservabilitySignal{
		ID:		id,
		Kind:		EconomicObservabilityKindQuery,
		Scope:		scope,
		Source:		source,
		Required:	true,
		Queryable:	true,
		SchemaVersion:	1,
		Labels:		append([]string{}, labels...),
	}
}

func evaluateObservabilitySignals(kind string, signals []EconomicObservabilitySignal, expectedIDs []string) ([]EconomicObservabilitySignal, []string, int, int) {
	out := append([]EconomicObservabilitySignal{}, signals...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	failed := make([]string, 0)
	seen := make(map[string]EconomicObservabilitySignal, len(out))
	for _, signal := range out {
		if signal.ID == "" {
			failed = append(failed, kind+":signal_id_required")
			continue
		}
		if signal.Kind != kind {
			failed = append(failed, signal.ID+":wrong_observability_kind")
		}
		if _, ok := seen[signal.ID]; ok {
			failed = append(failed, signal.ID+":duplicate_signal")
		}
		seen[signal.ID] = signal
		if signal.Required {
			if strings.TrimSpace(signal.Scope) == "" {
				failed = append(failed, signal.ID+":scope_missing")
			}
			if strings.TrimSpace(signal.Source) == "" {
				failed = append(failed, signal.ID+":source_missing")
			}
			if signal.SchemaVersion == 0 {
				failed = append(failed, signal.ID+":schema_version_missing")
			}
			if len(signal.Labels) == 0 {
				failed = append(failed, signal.ID+":labels_missing")
			}
			for i, label := range signal.Labels {
				if strings.TrimSpace(label) == "" {
					failed = append(failed, fmt.Sprintf("%s:label_%d_blank", signal.ID, i))
				}
			}
			switch kind {
			case EconomicObservabilityKindMetric:
				if !signal.Queryable {
					failed = append(failed, signal.ID+":metric_not_queryable")
				}
				if !signal.TelemetryEnabled {
					failed = append(failed, signal.ID+":telemetry_disabled")
				}
			case EconomicObservabilityKindEvent:
				if !signal.Emitted {
					failed = append(failed, signal.ID+":event_not_emitted")
				}
				if !signal.TelemetryEnabled {
					failed = append(failed, signal.ID+":telemetry_disabled")
				}
			case EconomicObservabilityKindQuery:
				if !signal.Queryable {
					failed = append(failed, signal.ID+":query_not_queryable")
				}
			}
		}
	}

	required := len(expectedIDs)
	covered := 0
	for _, id := range expectedIDs {
		signal, ok := seen[id]
		if !ok {
			failed = append(failed, id+":missing_required_observability")
			continue
		}
		if observabilitySignalCovered(kind, signal) {
			covered++
		}
	}
	return out, failed, required, covered
}

func observabilitySignalCovered(kind string, signal EconomicObservabilitySignal) bool {
	if !signal.Required || signal.Kind != kind || signal.SchemaVersion == 0 || len(signal.Labels) == 0 || strings.TrimSpace(signal.Source) == "" || strings.TrimSpace(signal.Scope) == "" {
		return false
	}
	if kind == EconomicObservabilityKindMetric {
		return signal.Queryable && signal.TelemetryEnabled
	}
	if kind == EconomicObservabilityKindEvent {
		return signal.Emitted && signal.TelemetryEnabled
	}
	return signal.Queryable
}

func requiredEconomicMetricIDs() []string {
	return []string{
		EconomicMetricCurrentInflationRate,
		EconomicMetricGrossMintedPerEpoch,
		EconomicMetricBurnedPerEpoch,
		EconomicMetricNetSupplyChangePerEpoch,
		EconomicMetricTotalBondedStakeRatio,
		EconomicMetricActiveValidatorCount,
		EconomicMetricStandbyValidatorCount,
		EconomicMetricTopNVotingPowerShare,
		EconomicMetricValidatorRewardPerVotingPower,
		EconomicMetricFeeRevenueByBucket,
		EconomicMetricBaseFeeCongestionMultiplier,
		EconomicMetricFailedTxSurchargeTotals,
		EconomicMetricStateBytesDelta,
		EconomicMetricStorageRentWarnings,
		EconomicMetricSlashingBySeverity,
		EconomicMetricReporterRewardsPaid,
		EconomicMetricDeflationGuardActivations,
		EconomicMetricCircuitBreakerActivations,
	}
}

func requiredEconomicEventIDs() []string {
	return []string{
		EconomicEventInflationUpdate,
		EconomicEventFeeAllocation,
		EconomicEventBurn,
		EconomicEventDeflationGuard,
		EconomicEventValidatorScoreUpdate,
		EconomicEventValidatorConcentrationWarning,
		EconomicEventCommissionChangeWarning,
		EconomicEventDelegationRiskWarning,
		EconomicEventSlashingRoute,
		EconomicEventReporterReward,
		EconomicEventStorageFee,
		EconomicEventStateRentWarning,
		EconomicEventCircuitBreaker,
	}
}

func requiredEconomicQueryIDs() []string {
	return []string{
		EconomicQueryCurrentParameters,
		EconomicQueryInflationState,
		EconomicQueryNetIssuanceByEpoch,
		EconomicQueryFeeDistributionByEpoch,
		EconomicQueryBurnHistory,
		EconomicQueryValidatorScoreComponents,
		EconomicQueryValidatorConcentration,
		EconomicQueryDelegatorRiskAdjustedYield,
		EconomicQueryStorageFootprint,
		EconomicQueryStateRentStatus,
		EconomicQueryFeeEstimateTxClass,
		EconomicQuerySupplyProjectionCurrentParams,
	}
}
