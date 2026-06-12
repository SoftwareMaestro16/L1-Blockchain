package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultFeeSmoothingWindow	= uint32(4)
	DefaultMaxBaseFeeAdjustmentBps	= uint32(1_250)
	DefaultEstimatorToleranceBps	= uint32(250)
	DefaultSpamSurchargeStepBps	= uint32(500)
	DefaultMaxSpamSurchargeBps	= uint32(5_000)
	DefaultCriticalReserveGasBps	= uint32(500)
	DefaultResourceMultiplierMaxBps	= uint32(30_000)

	ResourceCompute			= "compute"
	ResourceStorageWrite		= "storage_write"
	ResourceMessageForwarding	= "message_forwarding"
	ResourceDeployment		= "deployment"

	MessageClassStandard	= "standard"
	MessageClassCritical	= "critical_protocol"
	MessageClassDeployment	= "deployment"
	MessageClassForwarding	= "message_forwarding"

	FeeScenarioLowLoad	= "low_load"
	FeeScenarioSteadyLoad	= "steady_load"
	FeeScenarioBurstLoad	= "burst_load"
	FeeScenarioSpamLoad	= "spam_load"
)

type DynamicFeeControlParams struct {
	TargetBlockUtilizationBps	uint32
	MaxAdjustmentBps		uint32
	SmoothingWindow			uint32
	MinBaseFeeNaet			sdkmath.Int
	MaxBaseFeeNaet			sdkmath.Int
}

type FeeControlState struct {
	CurrentBaseFeeNaet	sdkmath.Int
	RecentUtilizations	[]uint32
}

type CongestionSignals struct {
	BlockGasUtilizationBps		uint32
	MempoolPressureBps		uint32
	FailedExecutionRateBps		uint32
	RepeatedSenderActivityBps	uint32
	StateWritePressureBps		uint32
}

type DynamicFeeControlResult struct {
	PreviousBaseFeeNaet	sdkmath.Int
	NextBaseFeeNaet		sdkmath.Int
	RawUtilizationBps	uint32
	SmoothedUtilizationBps	uint32
	AdjustmentBps		int32
	BoundedByGovernance	bool
}

type ResourceMultiplierParams struct {
	ComputeMaxBps		uint32
	StorageWriteMaxBps	uint32
	MessageForwardingMaxBps	uint32
	DeploymentMaxBps	uint32
	SpamSurchargeStepBps	uint32
	MaxSpamSurchargeBps	uint32
}

type ResourceFeeMultipliers struct {
	ComputeBps		uint32
	StorageWriteBps		uint32
	MessageForwardingBps	uint32
	DeploymentBps		uint32
	SpamSurchargeBps	uint32
}

type FeeEstimateInput struct {
	ControlParams		DynamicFeeControlParams
	State			FeeControlState
	Signals			CongestionSignals
	ResourceParams		ResourceMultiplierParams
	GasLimit		uint64
	ResourceClass		string
	RepeatedFailedTxs	uint32
	ActualInclusionFeeNaet	sdkmath.Int
	ToleranceBps		uint32
}

type FeeEstimate struct {
	RequiredFee		sdk.Coin
	BaseFee			sdk.Coin
	ResourceMultiplierBps	uint32
	SpamSurchargeBps	uint32
	EstimatedFeeNaet	sdkmath.Int
	ActualInclusionFeeNaet	sdkmath.Int
	WithinTolerance		bool
	ToleranceBps		uint32
	FeeControl		DynamicFeeControlResult
}

type MessageGasLimitPolicy struct {
	StandardMaxGas		uint64
	CriticalMaxGas		uint64
	DeploymentMaxGas	uint64
	ForwardingMaxGas	uint64
	CriticalReserveGasBps	uint32
}

type MessageGasLimitDecision struct {
	MessageClass		string
	MaxGas			uint64
	RequestedGas		uint64
	Allowed			bool
	CriticalReserveGas	uint64
	Auditable		bool
}

type FeeMarketSimulationStep struct {
	Scenario		string
	Signals			CongestionSignals
	GasLimit		uint64
	ResourceClass		string
	RepeatedFailedTxs	uint32
	ActualFeeNaet		sdkmath.Int
}

type FeeMarketSimulationReport struct {
	ScenarioCount		int
	FinalBaseFeeNaet	sdkmath.Int
	MaxBaseFeeNaet		sdkmath.Int
	EstimatorMismatches	int
	SpamSurchargeMaxBps	uint32
	Passed			bool
	Risks			[]string
}

func DefaultDynamicFeeControlParams(params Params) (DynamicFeeControlParams, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return DynamicFeeControlParams{}, err
	}
	baseFee, err := params.BaseFeeInt()
	if err != nil {
		return DynamicFeeControlParams{}, err
	}
	maxFee, err := params.MaxFeeInt()
	if err != nil {
		return DynamicFeeControlParams{}, err
	}
	return DynamicFeeControlParams{
		TargetBlockUtilizationBps:	params.TargetBlockUtilizationBps,
		MaxAdjustmentBps:		DefaultMaxBaseFeeAdjustmentBps,
		SmoothingWindow:		DefaultFeeSmoothingWindow,
		MinBaseFeeNaet:			baseFee,
		MaxBaseFeeNaet:			maxFee,
	}, nil
}

func (p DynamicFeeControlParams) Validate() error {
	if p.TargetBlockUtilizationBps == 0 || p.TargetBlockUtilizationBps >= uint32(BasisPoints) {
		return fmt.Errorf("target block utilization must be between 1 and 9999 bps")
	}
	if p.MaxAdjustmentBps == 0 || p.MaxAdjustmentBps > uint32(BasisPoints) {
		return fmt.Errorf("max adjustment must be within 1..10000 bps")
	}
	if p.SmoothingWindow == 0 {
		return errors.New("smoothing window must be positive")
	}
	if normalizeFeeInt(p.MinBaseFeeNaet).IsNegative() || !normalizeFeeInt(p.MaxBaseFeeNaet).IsPositive() {
		return errors.New("base fee bounds must be non-negative with positive max")
	}
	if normalizeFeeInt(p.MinBaseFeeNaet).GT(normalizeFeeInt(p.MaxBaseFeeNaet)) {
		return errors.New("minimum base fee cannot exceed maximum base fee")
	}
	return nil
}

func DefaultResourceMultiplierParams() ResourceMultiplierParams {
	return ResourceMultiplierParams{
		ComputeMaxBps:			DefaultResourceMultiplierMaxBps,
		StorageWriteMaxBps:		DefaultResourceMultiplierMaxBps,
		MessageForwardingMaxBps:	DefaultResourceMultiplierMaxBps,
		DeploymentMaxBps:		DefaultResourceMultiplierMaxBps,
		SpamSurchargeStepBps:		DefaultSpamSurchargeStepBps,
		MaxSpamSurchargeBps:		DefaultMaxSpamSurchargeBps,
	}
}

func (p ResourceMultiplierParams) Validate() error {
	if p.ComputeMaxBps == 0 {
		p.ComputeMaxBps = DefaultResourceMultiplierMaxBps
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "compute max", value: p.ComputeMaxBps},
		{name: "storage write max", value: p.StorageWriteMaxBps},
		{name: "message forwarding max", value: p.MessageForwardingMaxBps},
		{name: "deployment max", value: p.DeploymentMaxBps},
		{name: "spam surcharge step", value: p.SpamSurchargeStepBps},
		{name: "max spam surcharge", value: p.MaxSpamSurchargeBps},
	} {
		if item.value == 0 || item.value > 100_000 {
			return fmt.Errorf("%s must be within 1..100000 bps", item.name)
		}
	}
	return nil
}

func NextDynamicBaseFee(params DynamicFeeControlParams, state FeeControlState, signals CongestionSignals) (DynamicFeeControlResult, error) {
	if err := params.Validate(); err != nil {
		return DynamicFeeControlResult{}, err
	}
	if err := signals.Validate(); err != nil {
		return DynamicFeeControlResult{}, err
	}
	current := normalizeFeeInt(state.CurrentBaseFeeNaet)
	if !current.IsPositive() {
		current = params.MinBaseFeeNaet
	}
	current = clampInt(current, params.MinBaseFeeNaet, params.MaxBaseFeeNaet)
	smoothed := smoothedUtilizationBps(state.RecentUtilizations, signals.BlockGasUtilizationBps, params.SmoothingWindow)
	adjustment := boundedBaseFeeAdjustmentBps(params.TargetBlockUtilizationBps, smoothed, params.MaxAdjustmentBps)
	next := applyFeeAdjustment(current, adjustment)
	bounded := false
	if next.LT(params.MinBaseFeeNaet) {
		next = params.MinBaseFeeNaet
		bounded = true
	}
	if next.GT(params.MaxBaseFeeNaet) {
		next = params.MaxBaseFeeNaet
		bounded = true
	}
	return DynamicFeeControlResult{
		PreviousBaseFeeNaet:	current,
		NextBaseFeeNaet:	next,
		RawUtilizationBps:	signals.BlockGasUtilizationBps,
		SmoothedUtilizationBps:	smoothed,
		AdjustmentBps:		adjustment,
		BoundedByGovernance:	bounded || absInt32(adjustment) == int32(params.MaxAdjustmentBps),
	}, nil
}

func (s CongestionSignals) Validate() error {
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "block gas utilization", value: s.BlockGasUtilizationBps},
		{name: "mempool pressure", value: s.MempoolPressureBps},
		{name: "failed execution rate", value: s.FailedExecutionRateBps},
		{name: "repeated sender activity", value: s.RepeatedSenderActivityBps},
		{name: "state write pressure", value: s.StateWritePressureBps},
	} {
		if item.value > uint32(BasisPoints) {
			return fmt.Errorf("%s must be <= 10000 bps", item.name)
		}
	}
	return nil
}

func ResourceMultipliers(params ResourceMultiplierParams, signals CongestionSignals, repeatedFailedTxs uint32) (ResourceFeeMultipliers, error) {
	if params.ComputeMaxBps == 0 {
		params = DefaultResourceMultiplierParams()
	}
	if err := params.Validate(); err != nil {
		return ResourceFeeMultipliers{}, err
	}
	if err := signals.Validate(); err != nil {
		return ResourceFeeMultipliers{}, err
	}
	spam := uint32(uint64(repeatedFailedTxs) * uint64(params.SpamSurchargeStepBps))
	if spam > params.MaxSpamSurchargeBps {
		spam = params.MaxSpamSurchargeBps
	}
	return ResourceFeeMultipliers{
		ComputeBps:		boundedResourceMultiplier(signals.BlockGasUtilizationBps, signals.MempoolPressureBps, params.ComputeMaxBps),
		StorageWriteBps:	boundedResourceMultiplier(signals.StateWritePressureBps, signals.BlockGasUtilizationBps, params.StorageWriteMaxBps),
		MessageForwardingBps:	boundedResourceMultiplier(signals.MempoolPressureBps, signals.RepeatedSenderActivityBps, params.MessageForwardingMaxBps),
		DeploymentBps:		boundedResourceMultiplier(signals.BlockGasUtilizationBps, signals.StateWritePressureBps, params.DeploymentMaxBps),
		SpamSurchargeBps:	spam,
	}, nil
}

func EstimateDynamicFee(input FeeEstimateInput) (FeeEstimate, error) {
	if input.GasLimit == 0 {
		return FeeEstimate{}, errors.New("gas limit must be positive")
	}
	tolerance := input.ToleranceBps
	if tolerance == 0 {
		tolerance = DefaultEstimatorToleranceBps
	}
	if tolerance > uint32(BasisPoints) {
		return FeeEstimate{}, fmt.Errorf("estimator tolerance must be <= 10000 bps")
	}
	control, err := NextDynamicBaseFee(input.ControlParams, input.State, input.Signals)
	if err != nil {
		return FeeEstimate{}, err
	}
	multipliers, err := ResourceMultipliers(input.ResourceParams, input.Signals, input.RepeatedFailedTxs)
	if err != nil {
		return FeeEstimate{}, err
	}
	resourceMultiplier, err := selectResourceMultiplier(multipliers, input.ResourceClass)
	if err != nil {
		return FeeEstimate{}, err
	}
	totalMultiplier := uint32(uint64(resourceMultiplier) + uint64(multipliers.SpamSurchargeBps))
	if totalMultiplier < uint32(BasisPoints) {
		totalMultiplier = uint32(BasisPoints)
	}
	estimated := control.NextBaseFeeNaet.MulRaw(int64(totalMultiplier)).QuoRaw(int64(BasisPoints))
	actual := normalizeFeeInt(input.ActualInclusionFeeNaet)
	withinTolerance := true
	if actual.IsPositive() {
		withinTolerance = withinFeeTolerance(estimated, actual, tolerance)
	}
	return FeeEstimate{
		RequiredFee:		sdk.NewCoin(BondDenom, estimated),
		BaseFee:		sdk.NewCoin(BondDenom, control.NextBaseFeeNaet),
		ResourceMultiplierBps:	resourceMultiplier,
		SpamSurchargeBps:	multipliers.SpamSurchargeBps,
		EstimatedFeeNaet:	estimated,
		ActualInclusionFeeNaet:	actual,
		WithinTolerance:	withinTolerance,
		ToleranceBps:		tolerance,
		FeeControl:		control,
	}, nil
}

func DefaultMessageGasLimitPolicy(params Params) MessageGasLimitPolicy {
	params = NormalizeParams(params)
	return MessageGasLimitPolicy{
		StandardMaxGas:		params.MaxTxGas,
		CriticalMaxGas:		params.MaxTxGas / 2,
		DeploymentMaxGas:	params.MaxTxGas,
		ForwardingMaxGas:	params.MaxTxGas / 4,
		CriticalReserveGasBps:	DefaultCriticalReserveGasBps,
	}
}

func EvaluateMessageGasLimit(policy MessageGasLimitPolicy, messageClass string, requestedGas uint64, maxBlockGas uint64) (MessageGasLimitDecision, error) {
	if requestedGas == 0 {
		return MessageGasLimitDecision{}, errors.New("requested gas must be positive")
	}
	if maxBlockGas == 0 {
		return MessageGasLimitDecision{}, errors.New("max block gas must be positive")
	}
	if policy.CriticalReserveGasBps > uint32(BasisPoints) {
		return MessageGasLimitDecision{}, fmt.Errorf("critical reserve gas must be <= 10000 bps")
	}
	maxGas, err := maxGasForClass(policy, messageClass)
	if err != nil {
		return MessageGasLimitDecision{}, err
	}
	reserve := sdkmath.NewIntFromUint64(maxBlockGas).MulRaw(int64(policy.CriticalReserveGasBps)).QuoRaw(int64(BasisPoints)).Uint64()
	return MessageGasLimitDecision{
		MessageClass:		messageClass,
		MaxGas:			maxGas,
		RequestedGas:		requestedGas,
		Allowed:		requestedGas <= maxGas,
		CriticalReserveGas:	reserve,
		Auditable:		true,
	}, nil
}

func SimulateDynamicFeeMarket(params DynamicFeeControlParams, resourceParams ResourceMultiplierParams, initial FeeControlState, steps []FeeMarketSimulationStep) (FeeMarketSimulationReport, error) {
	if len(steps) == 0 {
		return FeeMarketSimulationReport{}, errors.New("fee market simulation requires steps")
	}
	state := initial
	maxFee := sdkmath.ZeroInt()
	mismatches := 0
	maxSpam := uint32(0)
	risks := make([]string, 0)
	scenarios := make(map[string]struct{})
	for i, step := range steps {
		if step.Scenario == "" {
			return FeeMarketSimulationReport{}, fmt.Errorf("simulation step %d scenario is required", i)
		}
		scenarios[step.Scenario] = struct{}{}
		estimate, err := EstimateDynamicFee(FeeEstimateInput{
			ControlParams:		params,
			State:			state,
			Signals:		step.Signals,
			ResourceParams:		resourceParams,
			GasLimit:		step.GasLimit,
			ResourceClass:		step.ResourceClass,
			RepeatedFailedTxs:	step.RepeatedFailedTxs,
			ActualInclusionFeeNaet:	step.ActualFeeNaet,
			ToleranceBps:		DefaultEstimatorToleranceBps,
		})
		if err != nil {
			return FeeMarketSimulationReport{}, err
		}
		if !estimate.WithinTolerance {
			mismatches++
		}
		if estimate.FeeControl.NextBaseFeeNaet.GT(maxFee) {
			maxFee = estimate.FeeControl.NextBaseFeeNaet
		}
		if estimate.SpamSurchargeBps > maxSpam {
			maxSpam = estimate.SpamSurchargeBps
		}
		state.CurrentBaseFeeNaet = estimate.FeeControl.NextBaseFeeNaet
		state.RecentUtilizations = appendSmoothedWindow(state.RecentUtilizations, step.Signals.BlockGasUtilizationBps, params.SmoothingWindow)
	}
	for _, required := range []string{FeeScenarioLowLoad, FeeScenarioSteadyLoad, FeeScenarioBurstLoad, FeeScenarioSpamLoad} {
		if _, found := scenarios[required]; !found {
			risks = append(risks, "missing_"+required)
		}
	}
	if mismatches > 0 {
		risks = append(risks, "fee_estimator_outside_tolerance")
	}
	sort.Strings(risks)
	return FeeMarketSimulationReport{
		ScenarioCount:		len(scenarios),
		FinalBaseFeeNaet:	state.CurrentBaseFeeNaet,
		MaxBaseFeeNaet:		maxFee,
		EstimatorMismatches:	mismatches,
		SpamSurchargeMaxBps:	maxSpam,
		Passed:			len(risks) == 0,
		Risks:			risks,
	}, nil
}

func smoothedUtilizationBps(recent []uint32, current uint32, window uint32) uint32 {
	values := appendSmoothedWindow(recent, current, window)
	sum := uint64(0)
	for _, value := range values {
		sum += uint64(value)
	}
	return uint32(sum / uint64(len(values)))
}

func appendSmoothedWindow(recent []uint32, current uint32, window uint32) []uint32 {
	if window == 0 {
		window = 1
	}
	values := make([]uint32, 0, int(window))
	keepRecent := int(window) - 1
	start := len(recent) - keepRecent
	if start < 0 {
		start = 0
	}
	for i, value := range recent {
		if i >= start && len(values) < keepRecent {
			values = append(values, value)
		}
	}
	values = append(values, current)
	return values
}

func boundedBaseFeeAdjustmentBps(targetBps, utilizationBps, maxAdjustmentBps uint32) int32 {
	if utilizationBps == targetBps {
		return 0
	}
	if utilizationBps > targetBps {
		denom := uint64(BasisPoints) - uint64(targetBps)
		if denom == 0 {
			return int32(maxAdjustmentBps)
		}
		adjustment := uint32(uint64(utilizationBps-targetBps) * uint64(maxAdjustmentBps) / denom)
		if adjustment > maxAdjustmentBps {
			adjustment = maxAdjustmentBps
		}
		return int32(adjustment)
	}
	adjustment := uint32(uint64(targetBps-utilizationBps) * uint64(maxAdjustmentBps) / uint64(targetBps))
	if adjustment > maxAdjustmentBps {
		adjustment = maxAdjustmentBps
	}
	return -int32(adjustment)
}

func applyFeeAdjustment(value sdkmath.Int, adjustmentBps int32) sdkmath.Int {
	if adjustmentBps == 0 {
		return value
	}
	if adjustmentBps > 0 {
		numerator := value.MulRaw(int64(BasisPoints) + int64(adjustmentBps))
		return ceilQuo(numerator, sdkmath.NewIntFromUint64(BasisPoints))
	}
	return value.MulRaw(int64(BasisPoints) + int64(adjustmentBps)).QuoRaw(int64(BasisPoints))
}

func boundedResourceMultiplier(left, right uint32, maxBps uint32) uint32 {
	pressure := left
	if right > pressure {
		pressure = right
	}
	multiplier := uint32(BasisPoints) + pressure
	if multiplier > maxBps {
		return maxBps
	}
	return multiplier
}

func selectResourceMultiplier(m ResourceFeeMultipliers, resourceClass string) (uint32, error) {
	switch resourceClass {
	case "", ResourceCompute:
		return m.ComputeBps, nil
	case ResourceStorageWrite:
		return m.StorageWriteBps, nil
	case ResourceMessageForwarding:
		return m.MessageForwardingBps, nil
	case ResourceDeployment:
		return m.DeploymentBps, nil
	default:
		return 0, fmt.Errorf("unsupported resource class %q", resourceClass)
	}
}

func maxGasForClass(policy MessageGasLimitPolicy, messageClass string) (uint64, error) {
	switch messageClass {
	case "", MessageClassStandard:
		return policy.StandardMaxGas, nil
	case MessageClassCritical:
		return policy.CriticalMaxGas, nil
	case MessageClassDeployment:
		return policy.DeploymentMaxGas, nil
	case MessageClassForwarding:
		return policy.ForwardingMaxGas, nil
	default:
		return 0, fmt.Errorf("unsupported message class %q", messageClass)
	}
}

func withinFeeTolerance(estimate sdkmath.Int, actual sdkmath.Int, toleranceBps uint32) bool {
	if !actual.IsPositive() {
		return true
	}
	diff := estimate.Sub(actual)
	if diff.IsNegative() {
		diff = diff.Neg()
	}
	return diff.MulRaw(int64(BasisPoints)).LTE(actual.MulRaw(int64(toleranceBps)))
}

func clampInt(value, min, max sdkmath.Int) sdkmath.Int {
	if value.LT(min) {
		return min
	}
	if value.GT(max) {
		return max
	}
	return value
}

func normalizeFeeInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}

func absInt32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}
