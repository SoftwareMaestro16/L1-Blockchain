package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestFeeMarketBaseFeeUpdateIsBounded(t *testing.T) {
	params := DefaultFeeMarketOptimizerParams()
	params.MaxBaseFeeAdjustmentBps = 1_250
	out, err := OptimizeFeeMarket(baseFeeMarketInput(params, FeeMarketOptimizerInput{
		CurrentBaseFeeNaet:	sdkmath.NewInt(1_000),
		BlockGasUsed:		9_500,
		BlockGasLimit:		10_000,
		TxGasUsed:		100,
		MempoolPressureBps:	3_000,
		FailedExecutionRateBps:	1_000,
		OfferedFeeNaet:		sdkmath.NewInt(1_000_000),
	}))
	require.NoError(t, err)
	require.Greater(t, out.BaseFeeNaet.Int64(), int64(1_000))
	require.LessOrEqual(t, out.AppliedDeltaBps, params.MaxBaseFeeAdjustmentBps)
	require.Equal(t, int64(9_500), out.BlockUtilizationBps)
	require.Contains(t, feeMarketEventTypes(out.CongestionEvents), FeeMarketEventBaseFeeUpdated)
	require.Contains(t, feeMarketEventTypes(out.CongestionEvents), FeeMarketEventCongestionDetected)
}

func TestFeeMarketResourceMultipliersAndSenderSurcharge(t *testing.T) {
	params := DefaultFeeMarketOptimizerParams()
	out, err := OptimizeFeeMarket(baseFeeMarketInput(params, FeeMarketOptimizerInput{
		CurrentBaseFeeNaet:	sdkmath.NewInt(100),
		BlockGasUsed:		7_000,
		BlockGasLimit:		10_000,
		TxGasUsed:		100,
		StateWriteBytes:	2_000,
		DeploymentBytes:	20_000,
		ForwardingMessages:	200,
		SenderFailedTxCount:	4,
		OfferedFeeNaet:		sdkmath.NewInt(10_000_000),
	}))
	require.NoError(t, err)
	require.Equal(t, int64(20_000), out.ResourceMultipliers.StorageBps)
	require.Equal(t, int64(20_000), out.ResourceMultipliers.DeploymentBps)
	require.Equal(t, int64(20_000), out.ResourceMultipliers.ForwardingBps)
	require.Equal(t, int64(2_000), out.SenderSurcharge.SurchargeBps)
	require.True(t, out.SenderSurcharge.SurchargeNaet.IsPositive())
	require.Contains(t, feeMarketEventTypes(out.CongestionEvents), FeeMarketEventSenderSurcharge)
}

func TestFeeMarketMempoolAndExecutionValidationAligned(t *testing.T) {
	params := DefaultFeeMarketOptimizerParams()
	estimate, err := OptimizeFeeMarket(baseFeeMarketInput(params, FeeMarketOptimizerInput{
		CurrentBaseFeeNaet:	sdkmath.NewInt(50),
		TxGasUsed:		1_000,
	}))
	require.NoError(t, err)

	out, err := OptimizeFeeMarket(baseFeeMarketInput(params, FeeMarketOptimizerInput{
		CurrentBaseFeeNaet:	sdkmath.NewInt(50),
		TxGasUsed:		1_000,
		OfferedFeeNaet:		estimate.Validation.RequiredFeeNaet,
	}))
	require.NoError(t, err)
	require.True(t, out.Validation.MempoolAccepted)
	require.True(t, out.Validation.ExecutionAccepted)
	require.True(t, out.Validation.MempoolExecutionAligned)
	require.Equal(t, out.Validation.RequiredFeeNaet, out.Estimate.RequiredFeeNaet)
}

func TestFeeMarketAllocationSumsExactlyWithCaps(t *testing.T) {
	params := DefaultFeeMarketOptimizerParams()
	params.EpochBurnCapNaet = sdkmath.NewInt(150)
	out, err := OptimizeFeeMarket(baseFeeMarketInput(params, FeeMarketOptimizerInput{
		CollectedFeesNaet:	sdkmath.NewInt(1_000),
		OfferedFeeNaet:		sdkmath.NewInt(1_000_000),
	}))
	require.NoError(t, err)
	require.True(t, out.Allocation.SumsExactly)
	require.Equal(t, sdkmath.NewInt(1_000), out.Allocation.CollectedFeesNaet)
	require.Equal(t, sdkmath.NewInt(150), out.Allocation.BurnNaet)
	require.Equal(t, sdkmath.NewInt(100), out.Allocation.StateReserveNaet)
	require.Equal(t, sdkmath.NewInt(50), out.Allocation.SecurityReserveNaet)
	require.Equal(t, sdkmath.NewInt(50), out.Allocation.CommunityPoolNaet)
	require.Equal(t, sdkmath.NewInt(650), out.Allocation.ValidatorDelegatorNaet)
	require.Contains(t, feeMarketEventTypes(out.CongestionEvents), FeeMarketEventAllocationAccounted)
}

func TestFeeMarketSimulationCoversLowSteadyBurstAndSpamLoad(t *testing.T) {
	params := DefaultFeeMarketOptimizerParams()
	report, err := SimulateFeeMarket(params, []FeeMarketSimulationStep{
		{
			Name:	"low_load",
			Input: baseFeeMarketInput(params, FeeMarketOptimizerInput{
				EpochID:	20,
				BlockGasUsed:	2_000,
				TxGasUsed:	100,
			}),
		},
		{
			Name:	"steady_load",
			Input: baseFeeMarketInput(params, FeeMarketOptimizerInput{
				EpochID:	21,
				BlockGasUsed:	7_000,
				TxGasUsed:	100,
			}),
		},
		{
			Name:	"burst_load",
			Input: baseFeeMarketInput(params, FeeMarketOptimizerInput{
				EpochID:		22,
				BlockGasUsed:		9_500,
				MempoolPressureBps:	4_000,
				TxGasUsed:		100,
				StateWriteBytes:	2_000,
			}),
		},
		{
			Name:	"spam_load",
			Input: baseFeeMarketInput(params, FeeMarketOptimizerInput{
				EpochID:		23,
				BlockGasUsed:		9_000,
				MempoolPressureBps:	5_000,
				FailedExecutionRateBps:	3_000,
				SenderFailedTxCount:	10,
				TxGasUsed:		100,
			}),
		},
	})
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Steps, 4)
	require.True(t, report.MinBaseFeeNaet.LTE(report.MaxBaseFeeNaet))
	require.True(t, report.Steps[3].SenderSurcharge.SurchargeNaet.IsPositive())
	for _, step := range report.Steps {
		require.True(t, step.Allocation.SumsExactly)
		require.True(t, step.Validation.MempoolExecutionAligned)
	}
}

func TestFeeMarketRejectsInvalidAllocationSplit(t *testing.T) {
	params := DefaultFeeMarketOptimizerParams()
	params.ValidatorRewardBps = 9_000
	_, err := OptimizeFeeMarket(baseFeeMarketInput(params, FeeMarketOptimizerInput{}))
	require.ErrorContains(t, err, "fee allocation bps must sum to 100%")
}

func baseFeeMarketInput(params FeeMarketOptimizerParams, override FeeMarketOptimizerInput) FeeMarketOptimizerInput {
	input := FeeMarketOptimizerInput{
		EpochID:		1,
		BlockHeight:		10,
		CurrentBaseFeeNaet:	params.DefaultBaseFeeNaet,
		BlockGasUsed:		7_000,
		BlockGasLimit:		10_000,
		TxGasUsed:		100,
		OfferedFeeNaet:		sdkmath.NewInt(1_000_000),
		Params:			params,
	}
	if override.EpochID != 0 {
		input.EpochID = override.EpochID
	}
	if override.BlockHeight != 0 {
		input.BlockHeight = override.BlockHeight
	}
	if !override.CurrentBaseFeeNaet.IsNil() {
		input.CurrentBaseFeeNaet = override.CurrentBaseFeeNaet
	}
	if override.RecentBaseFeesNaet != nil {
		input.RecentBaseFeesNaet = override.RecentBaseFeesNaet
	}
	if override.BlockGasUsed != 0 {
		input.BlockGasUsed = override.BlockGasUsed
	}
	if override.BlockGasLimit != 0 {
		input.BlockGasLimit = override.BlockGasLimit
	}
	if override.TxGasUsed != 0 {
		input.TxGasUsed = override.TxGasUsed
	}
	if override.MempoolPressureBps != 0 {
		input.MempoolPressureBps = override.MempoolPressureBps
	}
	if override.FailedExecutionRateBps != 0 {
		input.FailedExecutionRateBps = override.FailedExecutionRateBps
	}
	if override.SenderFailedTxCount != 0 {
		input.SenderFailedTxCount = override.SenderFailedTxCount
	}
	if override.StateWriteBytes != 0 {
		input.StateWriteBytes = override.StateWriteBytes
	}
	if override.DeploymentBytes != 0 {
		input.DeploymentBytes = override.DeploymentBytes
	}
	if override.ForwardingMessages != 0 {
		input.ForwardingMessages = override.ForwardingMessages
	}
	if !override.OfferedFeeNaet.IsNil() {
		input.OfferedFeeNaet = override.OfferedFeeNaet
	}
	if !override.CollectedFeesNaet.IsNil() {
		input.CollectedFeesNaet = override.CollectedFeesNaet
	}
	if !override.ExistingEpochBurnNaet.IsNil() {
		input.ExistingEpochBurnNaet = override.ExistingEpochBurnNaet
	}
	if !override.ExistingEpochStateReserve.IsNil() {
		input.ExistingEpochStateReserve = override.ExistingEpochStateReserve
	}
	if !override.ExistingEpochSecurityReserve.IsNil() {
		input.ExistingEpochSecurityReserve = override.ExistingEpochSecurityReserve
	}
	return input
}

func feeMarketEventTypes(events []FeeMarketCongestionEvent) []string {
	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.Type)
	}
	return types
}
