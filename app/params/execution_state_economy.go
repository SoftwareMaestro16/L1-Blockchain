package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	ExecutionCostClassCompute		= "compute"
	ExecutionCostClassStorageRead		= "storage_read"
	ExecutionCostClassStorageWrite		= "storage_write"
	ExecutionCostClassMessageForwarding	= "message_forwarding"

	ExecutionTraceKindOperation	= "operation"
	ExecutionTraceKindDeployment	= "deployment"
	ExecutionTraceKindAsyncFlow	= "async_flow"

	StateGrowthAlertAbnormal	= "abnormal_state_growth"

	DefaultComputeGasUnit				= uint64(1)
	DefaultStorageReadGasUnit			= uint64(5)
	DefaultStorageWriteGasUnit			= uint64(20)
	DefaultMessageForwardingGasUnit			= uint64(100)
	DefaultDeploymentBaseGas			= uint64(10_000)
	DefaultDeploymentCodeByteGas			= uint64(2)
	DefaultDeploymentInitStateByteGas		= uint64(4)
	DefaultExecutionGasPriceNaet			= int64(1)
	DefaultForwardingWorkloadToleranceBps		= int64(500)
	DefaultBenchmarkToleranceBps			= int64(1_000)
	DefaultDeploymentEstimateToleranceBps		= int64(500)
	DefaultStateGrowthSurchargeThresholdBytes	= int64(10_000)
	DefaultStateGrowthSurchargeStepBps		= int64(1_000)
	DefaultStateGrowthSurchargeMaxBps		= int64(10_000)
	DefaultStateMaintenanceReserveBps		= int64(1_000)
)

type ExecutionGasTable struct {
	ComputeGasUnit			uint64
	StorageReadGasUnit		uint64
	StorageWriteGasUnit		uint64
	MessageForwardingGasUnit	uint64
	DeploymentBaseGas		uint64
	DeploymentCodeByteGas		uint64
	DeploymentInitStateByteGas	uint64
	GasPriceNaet			sdkmath.Int
	BenchmarkToleranceBps		int64
	DeploymentEstimateToleranceBps	int64
	ForwardingWorkloadToleranceBps	int64
}

type ExecutionOperation struct {
	Class	string
	Count	uint64
}

type ExecutionTraceEntry struct {
	Kind	string
	Class	string
	Count	uint64
	GasUnit	uint64
	GasUsed	uint64
	FeeNaet	sdkmath.Int
}

type ExecutionCostInput struct {
	Operations	[]ExecutionOperation
	GasTable	ExecutionGasTable
}

type ExecutionCostOutput struct {
	TotalGas	uint64
	TotalFeeNaet	sdkmath.Int
	Traces		[]ExecutionTraceEntry
}

type DeploymentFeeEstimateInput struct {
	CodeSizeBytes		uint64
	InitStateBytes		uint64
	ExpectedActualFeeNaet	sdkmath.Int
	GasTable		ExecutionGasTable
}

type DeploymentFeeEstimateOutput struct {
	GasUsed			uint64
	FeeNaet			sdkmath.Int
	ExpectedActualFeeNaet	sdkmath.Int
	DeltaBps		int64
	WithinTolerance		bool
	Trace			ExecutionTraceEntry
}

type AsyncMessageFeeEstimateInput struct {
	MessageCount		uint64
	RouteHops		uint64
	PayloadBytes		uint64
	ExpectedWorkloadFeeNaet	sdkmath.Int
	GasTable		ExecutionGasTable
}

type AsyncMessageFeeEstimateOutput struct {
	GasUsed			uint64
	FeeNaet			sdkmath.Int
	ExpectedWorkloadFeeNaet	sdkmath.Int
	DeltaBps		int64
	CoversWorkload		bool
	Trace			ExecutionTraceEntry
}

type GasBenchmarkSample struct {
	Name		string
	Class		string
	MeasuredGas	uint64
	TableGas	uint64
}

type GasBenchmarkReport struct {
	Passed	bool
	Failed	[]string
	Samples	[]GasBenchmarkSample
}

type StateGrowthAccountDelta struct {
	ID		string
	BytesAdded	int64
	BytesRemoved	int64
}

type StateGrowthTelemetryInput struct {
	BlockHeight			uint64
	EpochID				uint64
	AccountDeltas			[]StateGrowthAccountDelta
	PreviousEpochNetGrowthBytes	int64
	BaseStorageExpansionFeeNaet	sdkmath.Int
	DeleteOriginalCostNaet		sdkmath.Int
	DeleteRefundNaet		sdkmath.Int
	StorageAgePeriods		uint64
	Params				StateGrowthParams
}

type StateGrowthParams struct {
	HighGrowthThresholdBytes	int64
	SurchargeStepBps		int64
	MaxSurchargeBps			int64
	StateMaintenanceReserveBps	int64
	DeleteRefundDecayBpsPerPeriod	int64
	AbnormalGrowthAlertBps		int64
}

type StateGrowthTopAccount struct {
	ID		string
	BytesAdded	int64
	BytesRemoved	int64
	NetGrowthBytes	int64
}

type StateGrowthTelemetryOutput struct {
	BlockHeight			uint64
	EpochID				uint64
	BytesAddedPerBlock		int64
	BytesRemovedPerBlock		int64
	NetGrowthBytes			int64
	NetStateGrowthPerEpochBytes	int64
	TopStateGrowthAccounts		[]StateGrowthTopAccount
	SurchargeBps			int64
	StorageExpansionFeeNaet		sdkmath.Int
	StateMaintenanceReserveNaet	sdkmath.Int
	DeleteRefundAfterDecayNaet	sdkmath.Int
	Alerts				[]string
}

func DefaultExecutionGasTable() ExecutionGasTable {
	return ExecutionGasTable{
		ComputeGasUnit:			DefaultComputeGasUnit,
		StorageReadGasUnit:		DefaultStorageReadGasUnit,
		StorageWriteGasUnit:		DefaultStorageWriteGasUnit,
		MessageForwardingGasUnit:	DefaultMessageForwardingGasUnit,
		DeploymentBaseGas:		DefaultDeploymentBaseGas,
		DeploymentCodeByteGas:		DefaultDeploymentCodeByteGas,
		DeploymentInitStateByteGas:	DefaultDeploymentInitStateByteGas,
		GasPriceNaet:			sdkmath.NewInt(DefaultExecutionGasPriceNaet),
		BenchmarkToleranceBps:		DefaultBenchmarkToleranceBps,
		DeploymentEstimateToleranceBps:	DefaultDeploymentEstimateToleranceBps,
		ForwardingWorkloadToleranceBps:	DefaultForwardingWorkloadToleranceBps,
	}
}

func DefaultStateGrowthParams() StateGrowthParams {
	return StateGrowthParams{
		HighGrowthThresholdBytes:	DefaultStateGrowthSurchargeThresholdBytes,
		SurchargeStepBps:		DefaultStateGrowthSurchargeStepBps,
		MaxSurchargeBps:		DefaultStateGrowthSurchargeMaxBps,
		StateMaintenanceReserveBps:	DefaultStateMaintenanceReserveBps,
		DeleteRefundDecayBpsPerPeriod:	DefaultDeleteRefundDecayBpsPerPeriod,
		AbnormalGrowthAlertBps:		BasisPoints,
	}
}

func EstimateExecutionCost(input ExecutionCostInput) (ExecutionCostOutput, error) {
	table := input.GasTable.withDefaults()
	if err := table.Validate(); err != nil {
		return ExecutionCostOutput{}, err
	}
	totalGas := uint64(0)
	traces := make([]ExecutionTraceEntry, 0, len(input.Operations))
	for _, op := range input.Operations {
		if op.Count == 0 {
			return ExecutionCostOutput{}, fmt.Errorf("execution operation count must be positive")
		}
		gasUnit, err := table.gasUnitForClass(op.Class)
		if err != nil {
			return ExecutionCostOutput{}, err
		}
		gasUsed, overflow := safeMulU64(gasUnit, op.Count)
		if overflow {
			return ExecutionCostOutput{}, fmt.Errorf("execution gas overflow")
		}
		totalGas, overflow = safeAddU64Local(totalGas, gasUsed)
		if overflow {
			return ExecutionCostOutput{}, fmt.Errorf("execution gas overflow")
		}
		traces = append(traces, ExecutionTraceEntry{
			Kind:		ExecutionTraceKindOperation,
			Class:		op.Class,
			Count:		op.Count,
			GasUnit:	gasUnit,
			GasUsed:	gasUsed,
			FeeNaet:	table.feeForGas(gasUsed),
		})
	}
	return ExecutionCostOutput{
		TotalGas:	totalGas,
		TotalFeeNaet:	table.feeForGas(totalGas),
		Traces:		traces,
	}, nil
}

func EstimateDeploymentFee(input DeploymentFeeEstimateInput) (DeploymentFeeEstimateOutput, error) {
	table := input.GasTable.withDefaults()
	if err := table.Validate(); err != nil {
		return DeploymentFeeEstimateOutput{}, err
	}
	codeGas, overflow := safeMulU64(table.DeploymentCodeByteGas, input.CodeSizeBytes)
	if overflow {
		return DeploymentFeeEstimateOutput{}, fmt.Errorf("deployment gas overflow")
	}
	initGas, overflow := safeMulU64(table.DeploymentInitStateByteGas, input.InitStateBytes)
	if overflow {
		return DeploymentFeeEstimateOutput{}, fmt.Errorf("deployment gas overflow")
	}
	totalGas, overflow := safeAddU64Local(table.DeploymentBaseGas, codeGas)
	if overflow {
		return DeploymentFeeEstimateOutput{}, fmt.Errorf("deployment gas overflow")
	}
	totalGas, overflow = safeAddU64Local(totalGas, initGas)
	if overflow {
		return DeploymentFeeEstimateOutput{}, fmt.Errorf("deployment gas overflow")
	}
	fee := table.feeForGas(totalGas)
	expected := normalizeInt(input.ExpectedActualFeeNaet)
	delta := estimateDeltaBps(fee, expected)
	return DeploymentFeeEstimateOutput{
		GasUsed:		totalGas,
		FeeNaet:		fee,
		ExpectedActualFeeNaet:	expected,
		DeltaBps:		delta,
		WithinTolerance:	expected.IsZero() || delta <= table.DeploymentEstimateToleranceBps,
		Trace: ExecutionTraceEntry{
			Kind:		ExecutionTraceKindDeployment,
			Class:		ExecutionTraceKindDeployment,
			Count:		input.CodeSizeBytes + input.InitStateBytes,
			GasUsed:	totalGas,
			FeeNaet:	fee,
		},
	}, nil
}

func EstimateAsyncMessageFlowFee(input AsyncMessageFeeEstimateInput) (AsyncMessageFeeEstimateOutput, error) {
	table := input.GasTable.withDefaults()
	if err := table.Validate(); err != nil {
		return AsyncMessageFeeEstimateOutput{}, err
	}
	if input.MessageCount == 0 {
		return AsyncMessageFeeEstimateOutput{}, fmt.Errorf("message_count must be positive")
	}
	workUnits := input.MessageCount
	if input.RouteHops > 0 {
		workUnits *= input.RouteHops
	}
	payloadUnits := (input.PayloadBytes + 255) / 256
	workUnits += payloadUnits
	gasUsed, overflow := safeMulU64(table.MessageForwardingGasUnit, workUnits)
	if overflow {
		return AsyncMessageFeeEstimateOutput{}, fmt.Errorf("message forwarding gas overflow")
	}
	fee := table.feeForGas(gasUsed)
	expected := normalizeInt(input.ExpectedWorkloadFeeNaet)
	delta := estimateDeltaBps(fee, expected)
	return AsyncMessageFeeEstimateOutput{
		GasUsed:			gasUsed,
		FeeNaet:			fee,
		ExpectedWorkloadFeeNaet:	expected,
		DeltaBps:			delta,
		CoversWorkload:			expected.IsZero() || fee.GTE(expected) || delta <= table.ForwardingWorkloadToleranceBps,
		Trace: ExecutionTraceEntry{
			Kind:		ExecutionTraceKindAsyncFlow,
			Class:		ExecutionCostClassMessageForwarding,
			Count:		workUnits,
			GasUnit:	table.MessageForwardingGasUnit,
			GasUsed:	gasUsed,
			FeeNaet:	fee,
		},
	}, nil
}

func ValidateGasBenchmarks(table ExecutionGasTable, samples []GasBenchmarkSample) (GasBenchmarkReport, error) {
	table = table.withDefaults()
	if err := table.Validate(); err != nil {
		return GasBenchmarkReport{}, err
	}
	if len(samples) == 0 {
		return GasBenchmarkReport{}, fmt.Errorf("benchmark samples must not be empty")
	}
	failed := make([]string, 0)
	for i, sample := range samples {
		if sample.Name == "" {
			return GasBenchmarkReport{}, fmt.Errorf("benchmark sample name is required")
		}
		if sample.TableGas == 0 {
			gasUnit, err := table.gasUnitForClass(sample.Class)
			if err != nil {
				return GasBenchmarkReport{}, err
			}
			samples[i].TableGas = gasUnit
			sample.TableGas = gasUnit
		}
		if sample.MeasuredGas == 0 {
			return GasBenchmarkReport{}, fmt.Errorf("benchmark measured gas must be positive")
		}
		delta := absInt64(int64(sample.TableGas)*BasisPoints/int64(sample.MeasuredGas) - BasisPoints)
		if delta > table.BenchmarkToleranceBps {
			failed = append(failed, sample.Name)
		}
	}
	return GasBenchmarkReport{Passed: len(failed) == 0, Failed: failed, Samples: samples}, nil
}

func ComputeStateGrowthTelemetry(input StateGrowthTelemetryInput) (StateGrowthTelemetryOutput, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return StateGrowthTelemetryOutput{}, err
	}
	if input.BlockHeight == 0 || input.EpochID == 0 {
		return StateGrowthTelemetryOutput{}, fmt.Errorf("block_height and epoch_id must be positive")
	}
	added := int64(0)
	removed := int64(0)
	top := make([]StateGrowthTopAccount, 0, len(input.AccountDeltas))
	for _, delta := range input.AccountDeltas {
		if delta.ID == "" {
			return StateGrowthTelemetryOutput{}, fmt.Errorf("state growth account id is required")
		}
		if delta.BytesAdded < 0 || delta.BytesRemoved < 0 {
			return StateGrowthTelemetryOutput{}, fmt.Errorf("state growth bytes must not be negative")
		}
		added += delta.BytesAdded
		removed += delta.BytesRemoved
		top = append(top, StateGrowthTopAccount{
			ID:		delta.ID,
			BytesAdded:	delta.BytesAdded,
			BytesRemoved:	delta.BytesRemoved,
			NetGrowthBytes:	delta.BytesAdded - delta.BytesRemoved,
		})
	}
	sort.SliceStable(top, func(i, j int) bool {
		if top[i].NetGrowthBytes == top[j].NetGrowthBytes {
			return top[i].ID < top[j].ID
		}
		return top[i].NetGrowthBytes > top[j].NetGrowthBytes
	})
	if len(top) > 5 {
		top = top[:5]
	}
	net := added - removed
	epochNet := input.PreviousEpochNetGrowthBytes + net
	surchargeBps := stateGrowthSurchargeBps(net, params)
	baseExpansionFee := normalizeInt(input.BaseStorageExpansionFeeNaet)
	if baseExpansionFee.IsNegative() {
		return StateGrowthTelemetryOutput{}, fmt.Errorf("base_storage_expansion_fee_naet must not be negative")
	}
	surchargeFee := ApplyBps(baseExpansionFee, surchargeBps)
	expansionFee := baseExpansionFee.Add(surchargeFee)
	reserve := ApplyBps(expansionFee, params.StateMaintenanceReserveBps)
	refundAfterDecay := storageChurnResistantRefund(input, params)
	alerts := make([]string, 0)
	if params.HighGrowthThresholdBytes > 0 && net > params.HighGrowthThresholdBytes {
		alerts = append(alerts, StateGrowthAlertAbnormal)
	}
	return StateGrowthTelemetryOutput{
		BlockHeight:			input.BlockHeight,
		EpochID:			input.EpochID,
		BytesAddedPerBlock:		added,
		BytesRemovedPerBlock:		removed,
		NetGrowthBytes:			net,
		NetStateGrowthPerEpochBytes:	epochNet,
		TopStateGrowthAccounts:		top,
		SurchargeBps:			surchargeBps,
		StorageExpansionFeeNaet:	expansionFee,
		StateMaintenanceReserveNaet:	reserve,
		DeleteRefundAfterDecayNaet:	refundAfterDecay,
		Alerts:				alerts,
	}, nil
}

func (t ExecutionGasTable) Validate() error {
	for _, item := range []struct {
		name	string
		value	uint64
	}{
		{name: "compute_gas_unit", value: t.ComputeGasUnit},
		{name: "storage_read_gas_unit", value: t.StorageReadGasUnit},
		{name: "storage_write_gas_unit", value: t.StorageWriteGasUnit},
		{name: "message_forwarding_gas_unit", value: t.MessageForwardingGasUnit},
		{name: "deployment_base_gas", value: t.DeploymentBaseGas},
		{name: "deployment_code_byte_gas", value: t.DeploymentCodeByteGas},
		{name: "deployment_init_state_byte_gas", value: t.DeploymentInitStateByteGas},
	} {
		if item.value == 0 {
			return fmt.Errorf("%s must be positive", item.name)
		}
	}
	if normalizeInt(t.GasPriceNaet).IsNegative() {
		return fmt.Errorf("gas_price_naet must not be negative")
	}
	if err := validateBps("benchmark_tolerance_bps", t.BenchmarkToleranceBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("deployment_estimate_tolerance_bps", t.DeploymentEstimateToleranceBps, 0, BasisPoints); err != nil {
		return err
	}
	return validateBps("forwarding_workload_tolerance_bps", t.ForwardingWorkloadToleranceBps, 0, BasisPoints)
}

func (t ExecutionGasTable) withDefaults() ExecutionGasTable {
	defaults := DefaultExecutionGasTable()
	if t.ComputeGasUnit == 0 {
		t.ComputeGasUnit = defaults.ComputeGasUnit
	}
	if t.StorageReadGasUnit == 0 {
		t.StorageReadGasUnit = defaults.StorageReadGasUnit
	}
	if t.StorageWriteGasUnit == 0 {
		t.StorageWriteGasUnit = defaults.StorageWriteGasUnit
	}
	if t.MessageForwardingGasUnit == 0 {
		t.MessageForwardingGasUnit = defaults.MessageForwardingGasUnit
	}
	if t.DeploymentBaseGas == 0 {
		t.DeploymentBaseGas = defaults.DeploymentBaseGas
	}
	if t.DeploymentCodeByteGas == 0 {
		t.DeploymentCodeByteGas = defaults.DeploymentCodeByteGas
	}
	if t.DeploymentInitStateByteGas == 0 {
		t.DeploymentInitStateByteGas = defaults.DeploymentInitStateByteGas
	}
	if t.GasPriceNaet.IsNil() {
		t.GasPriceNaet = defaults.GasPriceNaet
	}
	if t.BenchmarkToleranceBps == 0 {
		t.BenchmarkToleranceBps = defaults.BenchmarkToleranceBps
	}
	if t.DeploymentEstimateToleranceBps == 0 {
		t.DeploymentEstimateToleranceBps = defaults.DeploymentEstimateToleranceBps
	}
	if t.ForwardingWorkloadToleranceBps == 0 {
		t.ForwardingWorkloadToleranceBps = defaults.ForwardingWorkloadToleranceBps
	}
	return t
}

func (t ExecutionGasTable) gasUnitForClass(class string) (uint64, error) {
	switch class {
	case ExecutionCostClassCompute:
		return t.ComputeGasUnit, nil
	case ExecutionCostClassStorageRead:
		return t.StorageReadGasUnit, nil
	case ExecutionCostClassStorageWrite:
		return t.StorageWriteGasUnit, nil
	case ExecutionCostClassMessageForwarding:
		return t.MessageForwardingGasUnit, nil
	default:
		return 0, fmt.Errorf("unknown execution cost class %q", class)
	}
}

func (t ExecutionGasTable) feeForGas(gas uint64) sdkmath.Int {
	return normalizeInt(t.GasPriceNaet).MulRaw(int64(gas))
}

func (p StateGrowthParams) Validate() error {
	if p.HighGrowthThresholdBytes < 0 {
		return fmt.Errorf("high_growth_threshold_bytes must not be negative")
	}
	if err := validateBps("surcharge_step_bps", p.SurchargeStepBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("max_surcharge_bps", p.MaxSurchargeBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("state_maintenance_reserve_bps", p.StateMaintenanceReserveBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("delete_refund_decay_bps_per_period", p.DeleteRefundDecayBpsPerPeriod, 0, BasisPoints); err != nil {
		return err
	}
	return validateBps("abnormal_growth_alert_bps", p.AbnormalGrowthAlertBps, 0, DefaultMaxLoadMultiplierBps)
}

func (p StateGrowthParams) withDefaults() StateGrowthParams {
	defaults := DefaultStateGrowthParams()
	if p.HighGrowthThresholdBytes == 0 {
		p.HighGrowthThresholdBytes = defaults.HighGrowthThresholdBytes
	}
	if p.SurchargeStepBps == 0 {
		p.SurchargeStepBps = defaults.SurchargeStepBps
	}
	if p.MaxSurchargeBps == 0 {
		p.MaxSurchargeBps = defaults.MaxSurchargeBps
	}
	if p.StateMaintenanceReserveBps == 0 {
		p.StateMaintenanceReserveBps = defaults.StateMaintenanceReserveBps
	}
	if p.DeleteRefundDecayBpsPerPeriod == 0 {
		p.DeleteRefundDecayBpsPerPeriod = defaults.DeleteRefundDecayBpsPerPeriod
	}
	if p.AbnormalGrowthAlertBps == 0 {
		p.AbnormalGrowthAlertBps = defaults.AbnormalGrowthAlertBps
	}
	return p
}

func stateGrowthSurchargeBps(netGrowthBytes int64, params StateGrowthParams) int64 {
	if params.HighGrowthThresholdBytes <= 0 || netGrowthBytes <= params.HighGrowthThresholdBytes {
		return 0
	}
	multiple := (netGrowthBytes - params.HighGrowthThresholdBytes + params.HighGrowthThresholdBytes - 1) / params.HighGrowthThresholdBytes
	return clampInt64(multiple*params.SurchargeStepBps, 0, params.MaxSurchargeBps)
}

func storageChurnResistantRefund(input StateGrowthTelemetryInput, params StateGrowthParams) sdkmath.Int {
	refund := normalizeInt(input.DeleteRefundNaet)
	originalCost := normalizeInt(input.DeleteOriginalCostNaet)
	if refund.IsNegative() || originalCost.IsNegative() {
		return sdkmath.ZeroInt()
	}
	decay := clampInt64(int64(input.StorageAgePeriods)*params.DeleteRefundDecayBpsPerPeriod, 0, BasisPoints)
	afterDecay := ApplyBps(refund, BasisPoints-decay)
	if originalCost.IsPositive() && afterDecay.GT(originalCost) {
		return originalCost
	}
	return afterDecay
}

func estimateDeltaBps(estimate, actual sdkmath.Int) int64 {
	estimate = normalizeInt(estimate)
	actual = normalizeInt(actual)
	if actual.IsZero() {
		if estimate.IsZero() {
			return 0
		}
		return BasisPoints
	}
	return absInt64(estimate.Sub(actual).MulRaw(BasisPoints).Quo(actual).Int64())
}

func safeMulU64(a, b uint64) (uint64, bool) {
	if a != 0 && b > ^uint64(0)/a {
		return 0, true
	}
	return a * b, false
}

func safeAddU64Local(a, b uint64) (uint64, bool) {
	if b > ^uint64(0)-a {
		return 0, true
	}
	return a + b, false
}
