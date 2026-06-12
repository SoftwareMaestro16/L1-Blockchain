package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestNextDynamicBaseFeeIsDeterministicSmoothedAndBounded(t *testing.T) {
	params := DynamicFeeControlParams{
		TargetBlockUtilizationBps:	5_000,
		MaxAdjustmentBps:		1_000,
		SmoothingWindow:		4,
		MinBaseFeeNaet:			sdkmath.NewInt(100),
		MaxBaseFeeNaet:			sdkmath.NewInt(10_000),
	}
	state := FeeControlState{
		CurrentBaseFeeNaet:	sdkmath.NewInt(1_000),
		RecentUtilizations:	[]uint32{2_000, 3_000, 4_000},
	}
	burst := CongestionSignals{BlockGasUtilizationBps: 10_000}
	left, err := NextDynamicBaseFee(params, state, burst)
	if err != nil {
		t.Fatalf("dynamic fee should compute: %v", err)
	}
	right, err := NextDynamicBaseFee(params, state, burst)
	if err != nil {
		t.Fatalf("dynamic fee should compute twice: %v", err)
	}
	if !left.NextBaseFeeNaet.Equal(right.NextBaseFeeNaet) || left.SmoothedUtilizationBps != right.SmoothedUtilizationBps || left.AdjustmentBps != right.AdjustmentBps {
		t.Fatalf("dynamic base fee must be deterministic: left=%+v right=%+v", left, right)
	}
	if left.SmoothedUtilizationBps != 4_750 {
		t.Fatalf("expected smoothing to mute short burst to 4750 bps, got %d", left.SmoothedUtilizationBps)
	}
	if !left.NextBaseFeeNaet.Equal(sdkmath.NewInt(995)) {
		t.Fatalf("expected bounded decrease to 995, got %s", left.NextBaseFeeNaet)
	}

	spam := CongestionSignals{BlockGasUtilizationBps: 10_000}
	high, err := NextDynamicBaseFee(params, FeeControlState{
		CurrentBaseFeeNaet:	sdkmath.NewInt(1_000),
		RecentUtilizations:	[]uint32{8_000, 9_000, 9_500},
	}, spam)
	if err != nil {
		t.Fatalf("spam fee should compute: %v", err)
	}
	if high.AdjustmentBps != 825 || !high.NextBaseFeeNaet.Equal(sdkmath.NewInt(1_083)) {
		t.Fatalf("expected bounded spam adjustment, got %+v", high)
	}
	if absInt32(high.AdjustmentBps) > int32(params.MaxAdjustmentBps) {
		t.Fatalf("fee adjustment exceeded governance bound: %+v", high)
	}
}

func TestDynamicFeeEstimatorTracksActualCostsWithinTolerance(t *testing.T) {
	params := DynamicFeeControlParams{
		TargetBlockUtilizationBps:	5_000,
		MaxAdjustmentBps:		1_000,
		SmoothingWindow:		2,
		MinBaseFeeNaet:			sdkmath.NewInt(100),
		MaxBaseFeeNaet:			sdkmath.NewInt(10_000),
	}
	estimate, err := EstimateDynamicFee(FeeEstimateInput{
		ControlParams:		params,
		State:			FeeControlState{CurrentBaseFeeNaet: sdkmath.NewInt(1_000), RecentUtilizations: []uint32{6_000}},
		Signals:		CongestionSignals{BlockGasUtilizationBps: 6_000, MempoolPressureBps: 1_000},
		ResourceParams:		DefaultResourceMultiplierParams(),
		GasLimit:		100_000,
		ResourceClass:		ResourceCompute,
		ActualInclusionFeeNaet:	sdkmath.NewInt(1_600),
		ToleranceBps:		250,
	})
	if err != nil {
		t.Fatalf("estimate should compute: %v", err)
	}
	if !estimate.WithinTolerance {
		t.Fatalf("expected estimate to track actual fee within tolerance: %+v", estimate)
	}
	if estimate.ResourceMultiplierBps != 16_000 {
		t.Fatalf("expected compute multiplier from mempool pressure, got %d", estimate.ResourceMultiplierBps)
	}
}

func TestCongestionResponseSeparatesSignalsAndEscalatesFailedSpam(t *testing.T) {
	params := DefaultResourceMultiplierParams()
	signals := CongestionSignals{
		BlockGasUtilizationBps:		6_000,
		MempoolPressureBps:		9_000,
		FailedExecutionRateBps:		8_000,
		RepeatedSenderActivityBps:	9_500,
		StateWritePressureBps:		7_000,
	}
	low, err := ResourceMultipliers(params, signals, 1)
	if err != nil {
		t.Fatalf("resource multipliers should compute: %v", err)
	}
	high, err := ResourceMultipliers(params, signals, 20)
	if err != nil {
		t.Fatalf("resource multipliers should compute with spam: %v", err)
	}
	if low.ComputeBps != 19_000 {
		t.Fatalf("expected early compute congestion from mempool pressure, got %d", low.ComputeBps)
	}
	if low.StorageWriteBps != 17_000 {
		t.Fatalf("expected storage-write pressure multiplier, got %d", low.StorageWriteBps)
	}
	if high.SpamSurchargeBps <= low.SpamSurchargeBps || high.SpamSurchargeBps != params.MaxSpamSurchargeBps {
		t.Fatalf("expected failed spam surcharge to escalate to cap: low=%d high=%d", low.SpamSurchargeBps, high.SpamSurchargeBps)
	}
}

func TestMessageClassGasLimitsReserveCriticalProtocolOperations(t *testing.T) {
	policy := MessageGasLimitPolicy{
		StandardMaxGas:		1_000_000,
		CriticalMaxGas:		400_000,
		DeploymentMaxGas:	800_000,
		ForwardingMaxGas:	250_000,
		CriticalReserveGasBps:	500,
	}
	critical, err := EvaluateMessageGasLimit(policy, MessageClassCritical, 350_000, 20_000_000)
	if err != nil {
		t.Fatalf("critical gas decision should compute: %v", err)
	}
	if !critical.Allowed || critical.CriticalReserveGas != 1_000_000 || !critical.Auditable {
		t.Fatalf("expected bounded auditable critical reserve, got %+v", critical)
	}
	deployment, err := EvaluateMessageGasLimit(policy, MessageClassDeployment, 900_000, 20_000_000)
	if err != nil {
		t.Fatalf("deployment gas decision should compute: %v", err)
	}
	if deployment.Allowed {
		t.Fatalf("expected deployment above class limit to be rejected")
	}
}

func TestSimulateDynamicFeeMarketCoversLoadScenarios(t *testing.T) {
	params := DynamicFeeControlParams{
		TargetBlockUtilizationBps:	5_000,
		MaxAdjustmentBps:		1_000,
		SmoothingWindow:		4,
		MinBaseFeeNaet:			sdkmath.NewInt(100),
		MaxBaseFeeNaet:			sdkmath.NewInt(10_000),
	}
	steps := []FeeMarketSimulationStep{
		{Scenario: FeeScenarioLowLoad, Signals: CongestionSignals{BlockGasUtilizationBps: 2_000}, GasLimit: 100_000, ResourceClass: ResourceCompute},
		{Scenario: FeeScenarioSteadyLoad, Signals: CongestionSignals{BlockGasUtilizationBps: 5_000}, GasLimit: 100_000, ResourceClass: ResourceCompute},
		{Scenario: FeeScenarioBurstLoad, Signals: CongestionSignals{BlockGasUtilizationBps: 10_000}, GasLimit: 100_000, ResourceClass: ResourceCompute},
		{Scenario: FeeScenarioSpamLoad, Signals: CongestionSignals{BlockGasUtilizationBps: 9_500, MempoolPressureBps: 9_000, RepeatedSenderActivityBps: 9_000}, GasLimit: 100_000, ResourceClass: ResourceCompute, RepeatedFailedTxs: 20},
	}
	report, err := SimulateDynamicFeeMarket(params, DefaultResourceMultiplierParams(), FeeControlState{CurrentBaseFeeNaet: sdkmath.NewInt(100)}, steps)
	if err != nil {
		t.Fatalf("simulation should compute: %v", err)
	}
	if !report.Passed {
		t.Fatalf("expected fee simulation to pass, got %+v", report)
	}
	if report.ScenarioCount != 4 || report.SpamSurchargeMaxBps != DefaultMaxSpamSurchargeBps {
		t.Fatalf("expected full scenario coverage and capped spam surcharge, got %+v", report)
	}
}
