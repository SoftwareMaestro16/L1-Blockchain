package observability

import (
	"runtime"
	"time"
)

func RecordFinalizeBlock(height int64, blockTime time.Time, txCount int, duration time.Duration) {
	if height >= 0 {
		SetGauge(MetricBlockHeight, nil, float64(height))
	}
	if !blockTime.IsZero() {
		SetGauge(MetricBlockTimeSeconds, nil, float64(blockTime.Unix()))
	}
	if duration >= 0 {
		Observe(MetricBlockProcessing, Labels{"result": "finalized"}, duration.Seconds())
		if txCount > 0 {
			Observe(MetricTxLatency, Labels{"result": "finalized"}, duration.Seconds()/float64(txCount))
		}
	}
}

func RecordModuleError(module, action, reason string) {
	IncCounter(MetricModuleErrors, Labels{"module": module, "action": action, "reason": reason}, 1)
}

func RecordDexPoolCreated() {
	AddGauge(MetricDexPoolCount, nil, 1)
}

func RecordDexLiquidityNaetDelta(delta int64) {
	AddGauge(MetricDexLiquidityNaet, Labels{"denom": "naet"}, float64(delta))
}

func RecordDexSwap() {
	IncCounter(MetricDexSwaps, Labels{"result": "success"}, 1)
}

func RecordFeeAccepted() {
	IncCounter(MetricFeesAccepted, Labels{"result": "accepted"}, 1)
}

func RecordFeeRejected(reason string) {
	IncCounter(MetricFeesRejected, Labels{"reason": reason}, 1)
}

func RecordEconomicControl(inflationBps, burnRatioBps, validatorFeeRatioBps int64, deflationGuardActive, queueLimited, rateLimited bool) {
	SetGauge(MetricEconomyInflationBps, nil, float64(inflationBps))
	SetGauge(MetricEconomyBurnRatioBps, nil, float64(burnRatioBps))
	SetGauge(MetricEconomyValidatorFeeRatioBps, nil, float64(validatorFeeRatioBps))
	SetGauge(MetricEconomyDeflationGuard, nil, boolFloat(deflationGuardActive))
	SetGauge(MetricEconomyQueueLimited, nil, boolFloat(queueLimited))
	SetGauge(MetricEconomyRateLimited, nil, boolFloat(rateLimited))
}

func RecordEconomicFlow(totalChargesNaet, burnNaet, treasuryNaet, validatorRewardsNaet int64) {
	SetGauge(MetricEconomyTotalChargesNaet, Labels{"denom": "naet"}, float64(totalChargesNaet))
	SetGauge(MetricEconomyBurnNaet, Labels{"denom": "naet"}, float64(burnNaet))
	SetGauge(MetricEconomyTreasuryNaet, Labels{"denom": "naet"}, float64(treasuryNaet))
	SetGauge(MetricEconomyValidatorRewardsNaet, Labels{"denom": "naet"}, float64(validatorRewardsNaet))
}

func RecordOptimalEconomicState(optimal bool, failedConditionCount int) {
	if failedConditionCount < 0 {
		failedConditionCount = 0
	}
	SetGauge(MetricEconomyOptimalState, nil, boolFloat(optimal))
	SetGauge(MetricEconomyFailedConditions, nil, float64(failedConditionCount))
}

func (r *Registry) collectRuntime() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	r.SetGauge(MetricProcessUptimeSeconds, nil, time.Since(r.startedAt).Seconds())
	r.SetGauge(MetricProcessMemoryBytes, Labels{"type": "alloc"}, float64(mem.Alloc))
	r.SetGauge(MetricProcessGoroutines, nil, float64(runtime.NumGoroutine()))
}
