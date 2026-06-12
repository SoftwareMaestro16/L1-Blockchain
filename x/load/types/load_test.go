package types

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParamsValidate(t *testing.T) {
	params := DefaultParams()

	require.NoError(t, params.Validate())
	require.Equal(t, DefaultWindowBlocks, params.WindowBlocks)
	require.Equal(t, DefaultAlphaNumerator, params.AlphaNumerator)
	require.Equal(t, DefaultAlphaDenominator, params.AlphaDenominator)
	require.Equal(t, DefaultMaxDeltaBps, params.MaxDeltaBps)
	require.Equal(t, uint32(10_000), params.MempoolSizeWeightBps+
		params.BlockUtilizationWeightBps+
		params.TxLatencyWeightBps+
		params.FailureRateWeightBps+
		params.ExecutionTimeWeightBps)
}

func TestInvalidParamsRejected(t *testing.T) {
	tests := []struct {
		name	string
		mut	func(*Params)
		err	string
	}{
		{
			name:	"zero window",
			mut:	func(p *Params) { p.WindowBlocks = 0 },
			err:	"window",
		},
		{
			name:	"zero alpha numerator",
			mut:	func(p *Params) { p.AlphaNumerator = 0 },
			err:	"alpha numerator",
		},
		{
			name:	"zero alpha denominator",
			mut:	func(p *Params) { p.AlphaDenominator = 0 },
			err:	"alpha denominator",
		},
		{
			name:	"alpha numerator greater than denominator",
			mut:	func(p *Params) { p.AlphaNumerator = p.AlphaDenominator + 1 },
			err:	"cannot exceed",
		},
		{
			name:	"alpha denominator too large",
			mut:	func(p *Params) { p.AlphaDenominator = maxAlphaDenominator + 1 },
			err:	"alpha denominator",
		},
		{
			name:	"delta too large",
			mut:	func(p *Params) { p.MaxDeltaBps = BasisPoints + 1 },
			err:	"max delta",
		},
		{
			name:	"zero mempool target",
			mut:	func(p *Params) { p.TargetMempoolSize = 0 },
			err:	"mempool",
		},
		{
			name:	"zero block gas target",
			mut:	func(p *Params) { p.TargetBlockGas = 0 },
			err:	"block gas",
		},
		{
			name:	"zero latency target",
			mut:	func(p *Params) { p.TargetLatencyBlocks = 0 },
			err:	"latency",
		},
		{
			name:	"zero execution target",
			mut:	func(p *Params) { p.TargetExecutionSteps = 0 },
			err:	"execution",
		},
		{
			name:	"weights do not sum",
			mut:	func(p *Params) { p.ExecutionTimeWeightBps++ },
			err:	"weights",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			tt.mut(&params)
			require.ErrorContains(t, params.Validate(), tt.err)
		})
	}
}

func TestNormalizeMetricsBounds(t *testing.T) {
	params := DefaultParams()
	scores, err := NormalizeMetrics(params, Metrics{
		CanonicalMempoolSize:		params.TargetMempoolSize * 2,
		UsedBlockGas:			params.TargetBlockGas * 2,
		AverageInclusionDelayBlocks:	params.TargetLatencyBlocks * 2,
		FailedTxCount:			2,
		TotalTxCount:			1,
		ExecutionStepCount:		params.TargetExecutionSteps * 2,
	})
	require.NoError(t, err)
	require.Equal(t, Scores{
		MempoolSizeBps:		BasisPoints,
		BlockUtilizationBps:	BasisPoints,
		TxLatencyBps:		BasisPoints,
		FailureRateBps:		BasisPoints,
		ExecutionTimeBps:	BasisPoints,
	}, scores)
	require.NoError(t, scores.Validate())
}

func TestComputeLoadScoreDeterministic(t *testing.T) {
	params := DefaultParams()
	metrics := Metrics{
		CanonicalMempoolSize:		4_200,
		UsedBlockGas:			7_500_000,
		AverageInclusionDelayBlocks:	3,
		FailedTxCount:			7,
		TotalTxCount:			100,
		ExecutionStepCount:		8_250_000,
	}
	previous := EMAState{
		MempoolSizeBps:		100,
		BlockUtilizationBps:	200,
		TxLatencyBps:		300,
		FailureRateBps:		400,
		ExecutionTimeBps:	500,
		LoadScoreBps:		275,
		WindowHeight:		11,
	}

	left, err := ComputeLoadScore(params, previous, metrics)
	require.NoError(t, err)
	right, err := ComputeLoadScore(params, previous, metrics)
	require.NoError(t, err)

	require.Equal(t, left, right)
	require.Equal(t, uint64(12), left.EMA.WindowHeight)
	require.NoError(t, left.EMA.Validate())
}

func TestSpikeIsCappedByMaxDelta(t *testing.T) {
	params := DefaultParams()
	params.AlphaNumerator = 1
	params.AlphaDenominator = 1
	params.MaxDeltaBps = 500

	result, err := ComputeLoadScore(params, EMAState{LoadScoreBps: 1_000}, saturatedMetrics(params))
	require.NoError(t, err)

	require.Equal(t, uint32(10_000), result.RawLoadScoreBps)
	require.Equal(t, uint32(1_500), result.LoadScoreBps)
	require.Equal(t, result.LoadScoreBps, result.EMA.LoadScoreBps)
}

func TestLoadScoreMovesAcrossBandsGradually(t *testing.T) {
	params := DefaultParams()
	params.AlphaNumerator = 1
	params.AlphaDenominator = 1
	params.MaxDeltaBps = 1_000

	var state EMAState
	for i := 0; i < 2; i++ {
		result, err := ComputeLoadScore(params, state, saturatedMetrics(params))
		require.NoError(t, err)
		state = result.EMA
	}
	require.Equal(t, uint32(2_000), state.LoadScoreBps)
	require.Equal(t, LoadBandLow, BandForScore(state.LoadScoreBps))

	result, err := ComputeLoadScore(params, state, saturatedMetrics(params))
	require.NoError(t, err)
	state = result.EMA
	require.Equal(t, uint32(3_000), state.LoadScoreBps)
	require.Equal(t, LoadBandMedium, result.Band)

	for i := 0; i < 4; i++ {
		result, err = ComputeLoadScore(params, state, saturatedMetrics(params))
		require.NoError(t, err)
		state = result.EMA
	}
	require.Equal(t, uint32(7_000), state.LoadScoreBps)
	require.Equal(t, LoadBandHigh, BandForScore(state.LoadScoreBps))
}

func TestBandForScoreThresholds(t *testing.T) {
	require.Equal(t, LoadBandLow, BandForScore(0))
	require.Equal(t, LoadBandLow, BandForScore(2_999))
	require.Equal(t, LoadBandMedium, BandForScore(3_000))
	require.Equal(t, LoadBandMedium, BandForScore(6_999))
	require.Equal(t, LoadBandHigh, BandForScore(7_000))
	require.Equal(t, LoadBandHigh, BandForScore(10_000))
}

func TestFailureRateHandlesZeroTotalTxCount(t *testing.T) {
	params := DefaultParams()
	scores, err := NormalizeMetrics(params, Metrics{FailedTxCount: 0, TotalTxCount: 0})
	require.NoError(t, err)
	require.Zero(t, scores.FailureRateBps)

	scores, err = NormalizeMetrics(params, Metrics{FailedTxCount: 1, TotalTxCount: 0})
	require.NoError(t, err)
	require.Equal(t, BasisPoints, scores.FailureRateBps)
}

func TestExtremeInputsDoNotOverflow(t *testing.T) {
	params := DefaultParams()
	params.TargetMempoolSize = math.MaxUint64
	params.TargetBlockGas = math.MaxUint64
	params.TargetLatencyBlocks = math.MaxUint64
	params.TargetExecutionSteps = math.MaxUint64

	scores, err := NormalizeMetrics(params, Metrics{
		CanonicalMempoolSize:		math.MaxUint64 - 1,
		UsedBlockGas:			math.MaxUint64 - 1,
		AverageInclusionDelayBlocks:	math.MaxUint64 - 1,
		FailedTxCount:			math.MaxUint64 - 1,
		TotalTxCount:			math.MaxUint64,
		ExecutionStepCount:		math.MaxUint64 - 1,
	})
	require.NoError(t, err)
	require.NoError(t, scores.Validate())
	require.LessOrEqual(t, scores.MempoolSizeBps, BasisPoints)
	require.LessOrEqual(t, scores.BlockUtilizationBps, BasisPoints)
	require.LessOrEqual(t, scores.TxLatencyBps, BasisPoints)
	require.LessOrEqual(t, scores.FailureRateBps, BasisPoints)
	require.LessOrEqual(t, scores.ExecutionTimeBps, BasisPoints)

	_, err = ComputeLoadScore(params, EMAState{}, Metrics{
		CanonicalMempoolSize:		math.MaxUint64,
		UsedBlockGas:			math.MaxUint64,
		AverageInclusionDelayBlocks:	math.MaxUint64,
		FailedTxCount:			math.MaxUint64,
		TotalTxCount:			math.MaxUint64,
		ExecutionStepCount:		math.MaxUint64,
	})
	require.NoError(t, err)
}

func TestCorruptedPreviousEMARejected(t *testing.T) {
	_, err := ComputeLoadScore(DefaultParams(), EMAState{LoadScoreBps: BasisPoints + 1}, Metrics{})
	require.ErrorContains(t, err, "load score")

	_, err = ComputeLoadScore(DefaultParams(), EMAState{WindowHeight: math.MaxUint64}, Metrics{})
	require.ErrorContains(t, err, "window height overflow")
}

func TestEMAStateJSONRoundTripPreservesExactScore(t *testing.T) {
	state := EMAState{
		MempoolSizeBps:		111,
		BlockUtilizationBps:	222,
		TxLatencyBps:		333,
		FailureRateBps:		444,
		ExecutionTimeBps:	555,
		LoadScoreBps:		666,
		WindowHeight:		777,
	}

	raw, err := json.Marshal(state)
	require.NoError(t, err)
	var imported EMAState
	require.NoError(t, json.Unmarshal(raw, &imported))

	require.Equal(t, state, imported)
	require.NoError(t, imported.Validate())
}

func saturatedMetrics(params Params) Metrics {
	return Metrics{
		CanonicalMempoolSize:		params.TargetMempoolSize,
		UsedBlockGas:			params.TargetBlockGas,
		AverageInclusionDelayBlocks:	params.TargetLatencyBlocks,
		FailedTxCount:			1,
		TotalTxCount:			1,
		ExecutionStepCount:		params.TargetExecutionSteps,
	}
}
