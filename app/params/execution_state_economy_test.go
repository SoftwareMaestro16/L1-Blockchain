package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestExecutionCostClassesProducePredictableTraces(t *testing.T) {
	table := DefaultExecutionGasTable()
	table.GasPriceNaet = sdkmath.NewInt(2)

	out, err := EstimateExecutionCost(ExecutionCostInput{
		GasTable:	table,
		Operations: []ExecutionOperation{
			{Class: ExecutionCostClassCompute, Count: 10},
			{Class: ExecutionCostClassStorageRead, Count: 2},
			{Class: ExecutionCostClassStorageWrite, Count: 3},
			{Class: ExecutionCostClassMessageForwarding, Count: 1},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(180), out.TotalGas)
	require.Equal(t, sdkmath.NewInt(360), out.TotalFeeNaet)
	require.Len(t, out.Traces, 4)
	require.Equal(t, ExecutionCostClassStorageWrite, out.Traces[2].Class)
	require.Equal(t, uint64(60), out.Traces[2].GasUsed)
	require.Equal(t, sdkmath.NewInt(120), out.Traces[2].FeeNaet)
}

func TestGasBenchmarkReportBacksProtocolGasTable(t *testing.T) {
	table := DefaultExecutionGasTable()
	table.BenchmarkToleranceBps = 1_000
	report, err := ValidateGasBenchmarks(table, []GasBenchmarkSample{
		{Name: "compute/noop", Class: ExecutionCostClassCompute, MeasuredGas: 1},
		{Name: "storage/read", Class: ExecutionCostClassStorageRead, MeasuredGas: 5},
		{Name: "storage/write", Class: ExecutionCostClassStorageWrite, MeasuredGas: 19},
		{Name: "forward/message", Class: ExecutionCostClassMessageForwarding, MeasuredGas: 80},
	})
	require.NoError(t, err)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "forward/message")
	require.Len(t, report.Samples, 4)
	require.Equal(t, uint64(DefaultStorageWriteGasUnit), report.Samples[2].TableGas)
}

func TestDeploymentFeeEstimateWithinTolerance(t *testing.T) {
	table := DefaultExecutionGasTable()
	table.GasPriceNaet = sdkmath.NewInt(2)
	table.DeploymentEstimateToleranceBps = 500

	out, err := EstimateDeploymentFee(DeploymentFeeEstimateInput{
		CodeSizeBytes:		1_000,
		InitStateBytes:		250,
		ExpectedActualFeeNaet:	sdkmath.NewInt(26_100),
		GasTable:		table,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(13_000), out.GasUsed)
	require.Equal(t, sdkmath.NewInt(26_000), out.FeeNaet)
	require.True(t, out.WithinTolerance)
	require.LessOrEqual(t, out.DeltaBps, table.DeploymentEstimateToleranceBps)
	require.Equal(t, ExecutionTraceKindDeployment, out.Trace.Kind)
}

func TestAsyncMessageForwardingFeeCoversRoutedWorkload(t *testing.T) {
	table := DefaultExecutionGasTable()
	table.GasPriceNaet = sdkmath.NewInt(1)
	table.ForwardingWorkloadToleranceBps = 500

	out, err := EstimateAsyncMessageFlowFee(AsyncMessageFeeEstimateInput{
		MessageCount:			3,
		RouteHops:			2,
		PayloadBytes:			512,
		ExpectedWorkloadFeeNaet:	sdkmath.NewInt(790),
		GasTable:			table,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(800), out.GasUsed)
	require.Equal(t, sdkmath.NewInt(800), out.FeeNaet)
	require.True(t, out.CoversWorkload)
	require.Equal(t, uint64(8), out.Trace.Count)
	require.Equal(t, ExecutionCostClassMessageForwarding, out.Trace.Class)
}

func TestStateGrowthTelemetryReportsBlockEpochTopAccountsAndSurcharge(t *testing.T) {
	params := DefaultStateGrowthParams()
	params.HighGrowthThresholdBytes = 1_000
	params.SurchargeStepBps = 1_000
	params.MaxSurchargeBps = 5_000
	params.StateMaintenanceReserveBps = 2_000
	params.DeleteRefundDecayBpsPerPeriod = 1_000

	out, err := ComputeStateGrowthTelemetry(StateGrowthTelemetryInput{
		BlockHeight:			100,
		EpochID:			5,
		PreviousEpochNetGrowthBytes:	500,
		BaseStorageExpansionFeeNaet:	sdkmath.NewInt(10_000),
		DeleteOriginalCostNaet:		sdkmath.NewInt(1_000),
		DeleteRefundNaet:		sdkmath.NewInt(800),
		StorageAgePeriods:		3,
		Params:				params,
		AccountDeltas: []StateGrowthAccountDelta{
			{ID: "acct-a", BytesAdded: 1_500, BytesRemoved: 100},
			{ID: "contract-b", BytesAdded: 800, BytesRemoved: 0},
			{ID: "acct-c", BytesAdded: 0, BytesRemoved: 300},
		},
	})
	require.NoError(t, err)
	require.Equal(t, int64(2_300), out.BytesAddedPerBlock)
	require.Equal(t, int64(400), out.BytesRemovedPerBlock)
	require.Equal(t, int64(1_900), out.NetGrowthBytes)
	require.Equal(t, int64(2_400), out.NetStateGrowthPerEpochBytes)
	require.Equal(t, int64(1_000), out.SurchargeBps)
	require.Equal(t, sdkmath.NewInt(11_000), out.StorageExpansionFeeNaet)
	require.Equal(t, sdkmath.NewInt(2_200), out.StateMaintenanceReserveNaet)
	require.Equal(t, sdkmath.NewInt(560), out.DeleteRefundAfterDecayNaet)
	require.Contains(t, out.Alerts, StateGrowthAlertAbnormal)
	require.Len(t, out.TopStateGrowthAccounts, 3)
	require.Equal(t, "acct-a", out.TopStateGrowthAccounts[0].ID)
}

func TestStateGrowthRefundDecayResistsWriteDeleteExtraction(t *testing.T) {
	params := DefaultStateGrowthParams()
	params.DeleteRefundDecayBpsPerPeriod = 2_000
	out, err := ComputeStateGrowthTelemetry(StateGrowthTelemetryInput{
		BlockHeight:			200,
		EpochID:			6,
		BaseStorageExpansionFeeNaet:	sdkmath.NewInt(1_000),
		DeleteOriginalCostNaet:		sdkmath.NewInt(500),
		DeleteRefundNaet:		sdkmath.NewInt(900),
		StorageAgePeriods:		10,
		Params:				params,
		AccountDeltas: []StateGrowthAccountDelta{
			{ID: "contract-a", BytesAdded: 100, BytesRemoved: 100},
		},
	})
	require.NoError(t, err)
	require.True(t, out.DeleteRefundAfterDecayNaet.IsZero())
	require.True(t, len(out.Alerts) == 0)
}
