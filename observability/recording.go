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

func RecordDexLiquidityNorbDelta(delta int64) {
	AddGauge(MetricDexLiquidityNorb, Labels{"denom": "norb"}, float64(delta))
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

func (r *Registry) collectRuntime() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	r.SetGauge(MetricProcessUptimeSeconds, nil, time.Since(r.startedAt).Seconds())
	r.SetGauge(MetricProcessMemoryBytes, Labels{"type": "alloc"}, float64(mem.Alloc))
	r.SetGauge(MetricProcessGoroutines, nil, float64(runtime.NumGoroutine()))
}
