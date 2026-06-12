package types

import "testing"

func BenchmarkLoadScoreUpdate1k(b *testing.B) {
	benchmarkLoadScoreUpdate(b, 1_000)
}

func BenchmarkLoadScoreUpdate10k(b *testing.B) {
	benchmarkLoadScoreUpdate(b, 10_000)
}

func benchmarkLoadScoreUpdate(b *testing.B, updates int) {
	params := DefaultParams()
	metrics := make([]Metrics, updates)
	for i := 0; i < updates; i++ {
		metrics[i] = Metrics{
			CanonicalMempoolSize:		uint64((i * 37) % int(params.TargetMempoolSize+1)),
			UsedBlockGas:			uint64((i * 991) % int(params.TargetBlockGas+1)),
			AverageInclusionDelayBlocks:	uint64((i % int(params.TargetLatencyBlocks)) + 1),
			FailedTxCount:			uint64(i % 11),
			TotalTxCount:			uint64((i % 100) + 1),
			ExecutionStepCount:		uint64((i * 503) % int(params.TargetExecutionSteps+1)),
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var state EMAState
		for j := 0; j < updates; j++ {
			result, err := ComputeLoadScore(params, state, metrics[j])
			if err != nil {
				b.Fatal(err)
			}
			state = result.EMA
		}
	}
}
