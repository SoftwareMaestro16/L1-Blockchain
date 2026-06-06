package params

import (
	"fmt"
	"sort"
	"strings"
)

const (
	EconomicObservabilityKindMetric = "metric"
	EconomicObservabilityKindEvent  = "event"

	EconomicMetricCurrentInflationRate          = "current_inflation_rate"
	EconomicMetricGrossMintedPerEpoch           = "gross_minted_aet_per_epoch"
	EconomicMetricBurnedPerEpoch                = "burned_aet_per_epoch"
	EconomicMetricNetSupplyChangePerEpoch       = "net_supply_change_per_epoch"
	EconomicMetricTotalBondedStakeRatio         = "total_bonded_stake_ratio"
	EconomicMetricActiveValidatorCount          = "active_validator_count"
	EconomicMetricStandbyValidatorCount         = "standby_validator_count"
	EconomicMetricTopNVotingPowerShare          = "top_n_voting_power_share"
	EconomicMetricValidatorRewardPerVotingPower = "per_validator_reward_per_voting_power"
	EconomicMetricFeeRevenueByBucket            = "fee_revenue_by_bucket"
	EconomicMetricBaseFeeCongestionMultiplier   = "base_fee_and_congestion_multiplier"
	EconomicMetricFailedTxSurchargeTotals       = "failed_transaction_surcharge_totals"
	EconomicMetricStateBytesDelta               = "state_bytes_added_removed_net"
	EconomicMetricStorageRentWarnings           = "storage_rent_balances_and_exhaustion_warnings"
	EconomicMetricSlashingBySeverity            = "slashing_amount_by_severity"
	EconomicMetricReporterRewardsPaid           = "reporter_rewards_paid"
	EconomicMetricDeflationGuardActivations     = "deflation_guard_activation_count"
	EconomicMetricCircuitBreakerActivations     = "circuit_breaker_activation_count"

	EconomicEventInflationUpdate               = "inflation_update_event"
	EconomicEventFeeAllocation                 = "fee_allocation_event"
	EconomicEventBurn                          = "burn_event"
	EconomicEventDeflationGuard                = "deflation_guard_event"
	EconomicEventValidatorScoreUpdate          = "validator_score_update_event"
	EconomicEventValidatorConcentrationWarning = "validator_concentration_warning_event"
	EconomicEventCommissionChangeWarning       = "commission_change_warning_event"
	EconomicEventDelegationRiskWarning         = "delegation_risk_warning_event"
	EconomicEventSlashingRoute                 = "slashing_route_event"
	EconomicEventReporterReward                = "reporter_reward_event"
	EconomicEventStorageFee                    = "storage_fee_event"
	EconomicEventStateRentWarning              = "state_rent_warning_event"
	EconomicEventCircuitBreaker                = "circuit_breaker_event"
)

type EconomicObservabilitySignal struct {
	ID               string
	Kind             string
	Scope            string
	Source           string
	Required         bool
	Queryable        bool
	TelemetryEnabled bool
	Emitted          bool
	SchemaVersion    uint32
	Labels           []string
}

type EconomicObservabilityReport struct {
	Metrics           []EconomicObservabilitySignal
	Events            []EconomicObservabilitySignal
	RequiredMetrics   int
	RequiredEvents    int
	CoveredMetrics    int
	CoveredEvents     int
	MetricCoverageBps int64
	EventCoverageBps  int64
	Passed            bool
	Failed            []string
	GovernanceSummary string
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

func BuildEconomicObservabilityReport(metrics, events []EconomicObservabilitySignal) EconomicObservabilityReport {
	if metrics == nil {
		metrics = DefaultEconomicObservabilityMetrics()
	}
	if events == nil {
		events = DefaultEconomicObservabilityEvents()
	}
	metricSignals, metricFailed, requiredMetrics, coveredMetrics := evaluateObservabilitySignals(EconomicObservabilityKindMetric, metrics, requiredEconomicMetricIDs())
	eventSignals, eventFailed, requiredEvents, coveredEvents := evaluateObservabilitySignals(EconomicObservabilityKindEvent, events, requiredEconomicEventIDs())
	failed := append(metricFailed, eventFailed...)
	sort.Strings(failed)

	metricCoverage := coverageBps(coveredMetrics, requiredMetrics)
	eventCoverage := coverageBps(coveredEvents, requiredEvents)
	return EconomicObservabilityReport{
		Metrics:           metricSignals,
		Events:            eventSignals,
		RequiredMetrics:   requiredMetrics,
		RequiredEvents:    requiredEvents,
		CoveredMetrics:    coveredMetrics,
		CoveredEvents:     coveredEvents,
		MetricCoverageBps: metricCoverage,
		EventCoverageBps:  eventCoverage,
		Passed:            len(failed) == 0 && metricCoverage == BasisPoints && eventCoverage == BasisPoints,
		Failed:            failed,
		GovernanceSummary: fmt.Sprintf("required_metrics=%d/%d required_events=%d/%d metric_coverage_bps=%d event_coverage_bps=%d", coveredMetrics, requiredMetrics, coveredEvents, requiredEvents, metricCoverage, eventCoverage),
	}
}

func requiredMetric(id, scope, source string, labels ...string) EconomicObservabilitySignal {
	return EconomicObservabilitySignal{
		ID:               id,
		Kind:             EconomicObservabilityKindMetric,
		Scope:            scope,
		Source:           source,
		Required:         true,
		Queryable:        true,
		TelemetryEnabled: true,
		SchemaVersion:    1,
		Labels:           append([]string{}, labels...),
	}
}

func requiredEvent(id, scope, source string, labels ...string) EconomicObservabilitySignal {
	return EconomicObservabilitySignal{
		ID:               id,
		Kind:             EconomicObservabilityKindEvent,
		Scope:            scope,
		Source:           source,
		Required:         true,
		TelemetryEnabled: true,
		Emitted:          true,
		SchemaVersion:    1,
		Labels:           append([]string{}, labels...),
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
			if !signal.TelemetryEnabled {
				failed = append(failed, signal.ID+":telemetry_disabled")
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
			if kind == EconomicObservabilityKindMetric && !signal.Queryable {
				failed = append(failed, signal.ID+":metric_not_queryable")
			}
			if kind == EconomicObservabilityKindEvent && !signal.Emitted {
				failed = append(failed, signal.ID+":event_not_emitted")
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
	if !signal.Required || signal.Kind != kind || !signal.TelemetryEnabled || signal.SchemaVersion == 0 || len(signal.Labels) == 0 || strings.TrimSpace(signal.Source) == "" || strings.TrimSpace(signal.Scope) == "" {
		return false
	}
	if kind == EconomicObservabilityKindMetric {
		return signal.Queryable
	}
	return signal.Emitted
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
