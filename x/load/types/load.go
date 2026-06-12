package types

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
)

const (
	BasisPoints	= uint32(10_000)

	DefaultWindowBlocks	= uint64(60)
	DefaultAlphaNumerator	= uint64(2)
	DefaultAlphaDenominator	= DefaultWindowBlocks + 1
	DefaultMaxDeltaBps	= uint32(500)

	DefaultTargetMempoolSize	= uint64(10_000)
	DefaultTargetBlockGas		= uint64(20_000_000)
	DefaultTargetLatencyBlocks	= uint64(5)
	DefaultTargetExecutionSteps	= uint64(20_000_000)

	DefaultMempoolSizeWeightBps		= uint32(2_000)
	DefaultBlockUtilizationWeightBps	= uint32(3_000)
	DefaultTxLatencyWeightBps		= uint32(2_000)
	DefaultFailureRateWeightBps		= uint32(1_000)
	DefaultExecutionTimeWeightBps		= uint32(2_000)

	LowLoadUpperBps		= uint32(3_000)
	MediumLoadUpperBps	= uint32(7_000)

	maxAlphaDenominator	= uint64(1_000_000)
)

type LoadBand string

const (
	LoadBandLow	LoadBand	= "LOW"
	LoadBandMedium	LoadBand	= "MEDIUM"
	LoadBandHigh	LoadBand	= "HIGH"
)

type Params struct {
	WindowBlocks		uint64
	AlphaNumerator		uint64
	AlphaDenominator	uint64
	MaxDeltaBps		uint32

	TargetMempoolSize	uint64
	TargetBlockGas		uint64
	TargetLatencyBlocks	uint64
	TargetExecutionSteps	uint64

	MempoolSizeWeightBps		uint32
	BlockUtilizationWeightBps	uint32
	TxLatencyWeightBps		uint32
	FailureRateWeightBps		uint32
	ExecutionTimeWeightBps		uint32
}

type Metrics struct {
	CanonicalMempoolSize		uint64
	UsedBlockGas			uint64
	AverageInclusionDelayBlocks	uint64
	FailedTxCount			uint64
	TotalTxCount			uint64
	ExecutionStepCount		uint64
}

type Scores struct {
	MempoolSizeBps		uint32
	BlockUtilizationBps	uint32
	TxLatencyBps		uint32
	FailureRateBps		uint32
	ExecutionTimeBps	uint32
}

type EMAState struct {
	MempoolSizeBps		uint32
	BlockUtilizationBps	uint32
	TxLatencyBps		uint32
	FailureRateBps		uint32
	ExecutionTimeBps	uint32
	LoadScoreBps		uint32
	WindowHeight		uint64
}

type Result struct {
	Scores		Scores
	EMA		EMAState
	RawLoadScoreBps	uint32
	LoadScoreBps	uint32
	Band		LoadBand
}

func DefaultParams() Params {
	return Params{
		WindowBlocks:		DefaultWindowBlocks,
		AlphaNumerator:		DefaultAlphaNumerator,
		AlphaDenominator:	DefaultAlphaDenominator,
		MaxDeltaBps:		DefaultMaxDeltaBps,

		TargetMempoolSize:	DefaultTargetMempoolSize,
		TargetBlockGas:		DefaultTargetBlockGas,
		TargetLatencyBlocks:	DefaultTargetLatencyBlocks,
		TargetExecutionSteps:	DefaultTargetExecutionSteps,

		MempoolSizeWeightBps:		DefaultMempoolSizeWeightBps,
		BlockUtilizationWeightBps:	DefaultBlockUtilizationWeightBps,
		TxLatencyWeightBps:		DefaultTxLatencyWeightBps,
		FailureRateWeightBps:		DefaultFailureRateWeightBps,
		ExecutionTimeWeightBps:		DefaultExecutionTimeWeightBps,
	}
}

func (p Params) Validate() error {
	if p.WindowBlocks == 0 {
		return errors.New("load window blocks must be positive")
	}
	if p.AlphaNumerator == 0 {
		return errors.New("load alpha numerator must be positive")
	}
	if p.AlphaDenominator == 0 {
		return errors.New("load alpha denominator must be positive")
	}
	if p.AlphaNumerator > p.AlphaDenominator {
		return errors.New("load alpha numerator cannot exceed denominator")
	}
	if p.AlphaDenominator > maxAlphaDenominator {
		return fmt.Errorf("load alpha denominator must be <= %d", maxAlphaDenominator)
	}
	if p.MaxDeltaBps > BasisPoints {
		return fmt.Errorf("load max delta must be <= %d bps", BasisPoints)
	}
	if p.TargetMempoolSize == 0 {
		return errors.New("load target mempool size must be positive")
	}
	if p.TargetBlockGas == 0 {
		return errors.New("load target block gas must be positive")
	}
	if p.TargetLatencyBlocks == 0 {
		return errors.New("load target latency blocks must be positive")
	}
	if p.TargetExecutionSteps == 0 {
		return errors.New("load target execution steps must be positive")
	}
	weightSum := uint64(p.MempoolSizeWeightBps) +
		uint64(p.BlockUtilizationWeightBps) +
		uint64(p.TxLatencyWeightBps) +
		uint64(p.FailureRateWeightBps) +
		uint64(p.ExecutionTimeWeightBps)
	if weightSum != uint64(BasisPoints) {
		return fmt.Errorf("load metric weights must sum to %d bps", BasisPoints)
	}
	return nil
}

func (s Scores) Validate() error {
	if s.MempoolSizeBps > BasisPoints {
		return fmt.Errorf("mempool size score exceeds %d bps", BasisPoints)
	}
	if s.BlockUtilizationBps > BasisPoints {
		return fmt.Errorf("block utilization score exceeds %d bps", BasisPoints)
	}
	if s.TxLatencyBps > BasisPoints {
		return fmt.Errorf("tx latency score exceeds %d bps", BasisPoints)
	}
	if s.FailureRateBps > BasisPoints {
		return fmt.Errorf("failure rate score exceeds %d bps", BasisPoints)
	}
	if s.ExecutionTimeBps > BasisPoints {
		return fmt.Errorf("execution time score exceeds %d bps", BasisPoints)
	}
	return nil
}

func (s EMAState) Validate() error {
	scores := Scores{
		MempoolSizeBps:		s.MempoolSizeBps,
		BlockUtilizationBps:	s.BlockUtilizationBps,
		TxLatencyBps:		s.TxLatencyBps,
		FailureRateBps:		s.FailureRateBps,
		ExecutionTimeBps:	s.ExecutionTimeBps,
	}
	if err := scores.Validate(); err != nil {
		return err
	}
	if s.LoadScoreBps > BasisPoints {
		return fmt.Errorf("load score exceeds %d bps", BasisPoints)
	}
	return nil
}

func NormalizeMetrics(params Params, metrics Metrics) (Scores, error) {
	if err := params.Validate(); err != nil {
		return Scores{}, err
	}
	scores := Scores{
		MempoolSizeBps:		normalizeRatioBps(metrics.CanonicalMempoolSize, params.TargetMempoolSize),
		BlockUtilizationBps:	normalizeRatioBps(metrics.UsedBlockGas, params.TargetBlockGas),
		TxLatencyBps:		normalizeRatioBps(metrics.AverageInclusionDelayBlocks, params.TargetLatencyBlocks),
		FailureRateBps:		normalizeRatioBps(metrics.FailedTxCount, maxUint64(1, metrics.TotalTxCount)),
		ExecutionTimeBps:	normalizeRatioBps(metrics.ExecutionStepCount, params.TargetExecutionSteps),
	}
	return scores, nil
}

func ComputeLoadScore(params Params, previous EMAState, metrics Metrics) (Result, error) {
	if err := params.Validate(); err != nil {
		return Result{}, err
	}
	if err := previous.Validate(); err != nil {
		return Result{}, err
	}
	if previous.WindowHeight == math.MaxUint64 {
		return Result{}, errors.New("load EMA window height overflow")
	}
	scores, err := NormalizeMetrics(params, metrics)
	if err != nil {
		return Result{}, err
	}
	ema := EMAState{
		MempoolSizeBps:		updateEMA(params, previous.MempoolSizeBps, scores.MempoolSizeBps),
		BlockUtilizationBps:	updateEMA(params, previous.BlockUtilizationBps, scores.BlockUtilizationBps),
		TxLatencyBps:		updateEMA(params, previous.TxLatencyBps, scores.TxLatencyBps),
		FailureRateBps:		updateEMA(params, previous.FailureRateBps, scores.FailureRateBps),
		ExecutionTimeBps:	updateEMA(params, previous.ExecutionTimeBps, scores.ExecutionTimeBps),
		WindowHeight:		previous.WindowHeight + 1,
	}
	rawScore := weightedScore(params, Scores{
		MempoolSizeBps:		ema.MempoolSizeBps,
		BlockUtilizationBps:	ema.BlockUtilizationBps,
		TxLatencyBps:		ema.TxLatencyBps,
		FailureRateBps:		ema.FailureRateBps,
		ExecutionTimeBps:	ema.ExecutionTimeBps,
	})
	ema.LoadScoreBps = clampDelta(previous.LoadScoreBps, rawScore, params.MaxDeltaBps)
	return Result{
		Scores:			scores,
		EMA:			ema,
		RawLoadScoreBps:	rawScore,
		LoadScoreBps:		ema.LoadScoreBps,
		Band:			BandForScore(ema.LoadScoreBps),
	}, nil
}

func BandForScore(scoreBps uint32) LoadBand {
	switch {
	case scoreBps < LowLoadUpperBps:
		return LoadBandLow
	case scoreBps < MediumLoadUpperBps:
		return LoadBandMedium
	default:
		return LoadBandHigh
	}
}

func normalizeRatioBps(value, target uint64) uint32 {
	if target == 0 || value >= target {
		return BasisPoints
	}
	hi, lo := bits.Mul64(value, uint64(BasisPoints))
	quotient, _ := bits.Div64(hi, lo, target)
	if quotient > uint64(BasisPoints) {
		return BasisPoints
	}
	return uint32(quotient)
}

func updateEMA(params Params, previous, current uint32) uint32 {
	weightedCurrent := params.AlphaNumerator * uint64(current)
	weightedPrevious := (params.AlphaDenominator - params.AlphaNumerator) * uint64(previous)
	return uint32((weightedCurrent + weightedPrevious) / params.AlphaDenominator)
}

func weightedScore(params Params, scores Scores) uint32 {
	total := uint64(params.MempoolSizeWeightBps)*uint64(scores.MempoolSizeBps) +
		uint64(params.BlockUtilizationWeightBps)*uint64(scores.BlockUtilizationBps) +
		uint64(params.TxLatencyWeightBps)*uint64(scores.TxLatencyBps) +
		uint64(params.FailureRateWeightBps)*uint64(scores.FailureRateBps) +
		uint64(params.ExecutionTimeWeightBps)*uint64(scores.ExecutionTimeBps)
	score := total / uint64(BasisPoints)
	if score > uint64(BasisPoints) {
		return BasisPoints
	}
	return uint32(score)
}

func clampDelta(previous, current, maxDelta uint32) uint32 {
	if current > previous {
		delta := current - previous
		if delta > maxDelta {
			return previous + maxDelta
		}
		return current
	}
	delta := previous - current
	if delta > maxDelta {
		return previous - maxDelta
	}
	return current
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
