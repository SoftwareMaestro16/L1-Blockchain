package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	FeeMarketEventBaseFeeUpdated		= "fee_market_base_fee_updated"
	FeeMarketEventCongestionDetected	= "fee_market_congestion_detected"
	FeeMarketEventSenderSurcharge		= "fee_market_sender_surcharge"
	FeeMarketEventAllocationAccounted	= "fee_market_allocation_accounted"

	FeeBucketValidatorDelegator	= "validator_delegator"
	FeeBucketCommunityPool		= "community_pool"
	FeeBucketBurn			= "burn"
	FeeBucketStateReserve		= "state_maintenance_reserve"
	FeeBucketSecurityReserve	= "security_reserve"

	DefaultFeeMarketBaseFeeNaet		= int64(10)
	DefaultFeeMarketMinBaseFeeNaet		= int64(1)
	DefaultFeeMarketMaxBaseFeeNaet		= int64(1_000_000)
	DefaultFeeMarketMaxBaseFeeAdjustmentBps	= int64(1_250)
	DefaultFeeMarketSmoothingWindow		= uint32(4)
	DefaultFeeMarketStateWriteThreshold	= int64(1_000)
	DefaultFeeMarketDeploymentThreshold	= int64(10_000)
	DefaultFeeMarketForwardingThreshold	= uint64(100)
	DefaultFeeMarketForwardingUnitGas	= uint64(50)
	DefaultFeeMarketSurchargeStepBps	= int64(500)
	DefaultFeeMarketMaxSurchargeBps		= int64(10_000)
	DefaultFeeMarketMaxMultiplierBps	= int64(40_000)
)

type FeeMarketOptimizerParams struct {
	MinBaseFeeNaet			sdkmath.Int
	MaxBaseFeeNaet			sdkmath.Int
	DefaultBaseFeeNaet		sdkmath.Int
	TargetBlockUtilizationBps	int64
	MaxBaseFeeAdjustmentBps		int64
	SmoothingWindow			uint32
	MinResourceMultiplierBps	int64
	MaxResourceMultiplierBps	int64
	StateWriteThresholdBytes	int64
	DeploymentThresholdBytes	int64
	ForwardingThresholdMessages	uint64
	ForwardingUnitGas		uint64
	SurchargeStepBps		int64
	MaxSenderSurchargeBps		int64
	ValidatorRewardBps		int64
	CommunityPoolBps		int64
	BurnBps				int64
	StateReserveBps			int64
	SecurityReserveBps		int64
	EpochBurnCapNaet		sdkmath.Int
	EpochStateReserveCapNaet	sdkmath.Int
	EpochSecurityReserveCapNaet	sdkmath.Int
}

type FeeMarketOptimizerInput struct {
	EpochID				uint64
	BlockHeight			uint64
	CurrentBaseFeeNaet		sdkmath.Int
	RecentBaseFeesNaet		[]sdkmath.Int
	BlockGasUsed			uint64
	BlockGasLimit			uint64
	TxGasUsed			uint64
	MempoolPressureBps		int64
	FailedExecutionRateBps		int64
	SenderFailedTxCount		uint64
	StateWriteBytes			int64
	DeploymentBytes			int64
	ForwardingMessages		uint64
	OfferedFeeNaet			sdkmath.Int
	CollectedFeesNaet		sdkmath.Int
	ExistingEpochBurnNaet		sdkmath.Int
	ExistingEpochStateReserve	sdkmath.Int
	ExistingEpochSecurityReserve	sdkmath.Int
	Params				FeeMarketOptimizerParams
}

type FeeResourceMultipliers struct {
	ComputeBps	int64
	StorageBps	int64
	DeploymentBps	int64
	ForwardingBps	int64
}

type SenderLocalSurcharge struct {
	FailedTxCount	uint64
	SurchargeBps	int64
	SurchargeNaet	sdkmath.Int
}

type FeeValidationResult struct {
	RequiredFeeNaet		sdkmath.Int
	OfferedFeeNaet		sdkmath.Int
	MempoolAccepted		bool
	ExecutionAccepted	bool
	MempoolExecutionAligned	bool
}

type FeeAllocationBuckets struct {
	CollectedFeesNaet		sdkmath.Int
	ValidatorDelegatorNaet		sdkmath.Int
	CommunityPoolNaet		sdkmath.Int
	BurnNaet			sdkmath.Int
	StateReserveNaet		sdkmath.Int
	SecurityReserveNaet		sdkmath.Int
	BurnCapRemainderNaet		sdkmath.Int
	StateCapRemainderNaet		sdkmath.Int
	SecurityCapRemainderNaet	sdkmath.Int
	SumsExactly			bool
}

type FeeEstimateData struct {
	BaseFeeNaet		sdkmath.Int
	RequiredFeeNaet		sdkmath.Int
	ConservativeFeeNaet	sdkmath.Int
	ResourceMultipliers	FeeResourceMultipliers
	SenderSurchargeBps	int64
}

type FeeMarketCongestionEvent struct {
	Type		string
	EpochID		uint64
	BlockHeight	uint64
	Reason		string
	ValueBps	int64
	AmountNaet	sdkmath.Int
}

type FeeMarketOptimizerOutput struct {
	BaseFeeNaet		sdkmath.Int
	PreviousBaseFeeNaet	sdkmath.Int
	AppliedDeltaBps		int64
	BlockUtilizationBps	int64
	ResourceMultipliers	FeeResourceMultipliers
	SenderSurcharge		SenderLocalSurcharge
	Validation		FeeValidationResult
	Allocation		FeeAllocationBuckets
	Estimate		FeeEstimateData
	CongestionEvents	[]FeeMarketCongestionEvent
}

type FeeMarketSimulationStep struct {
	Name	string
	Input	FeeMarketOptimizerInput
}

type FeeMarketSimulationReport struct {
	Steps		[]FeeMarketOptimizerOutput
	MinBaseFeeNaet	sdkmath.Int
	MaxBaseFeeNaet	sdkmath.Int
	Passed		bool
	Failed		[]string
}

func DefaultFeeMarketOptimizerParams() FeeMarketOptimizerParams {
	return FeeMarketOptimizerParams{
		MinBaseFeeNaet:			sdkmath.NewInt(DefaultFeeMarketMinBaseFeeNaet),
		MaxBaseFeeNaet:			sdkmath.NewInt(DefaultFeeMarketMaxBaseFeeNaet),
		DefaultBaseFeeNaet:		sdkmath.NewInt(DefaultFeeMarketBaseFeeNaet),
		TargetBlockUtilizationBps:	DefaultTargetLoadBps,
		MaxBaseFeeAdjustmentBps:	DefaultFeeMarketMaxBaseFeeAdjustmentBps,
		SmoothingWindow:		DefaultFeeMarketSmoothingWindow,
		MinResourceMultiplierBps:	BasisPoints,
		MaxResourceMultiplierBps:	DefaultFeeMarketMaxMultiplierBps,
		StateWriteThresholdBytes:	DefaultFeeMarketStateWriteThreshold,
		DeploymentThresholdBytes:	DefaultFeeMarketDeploymentThreshold,
		ForwardingThresholdMessages:	DefaultFeeMarketForwardingThreshold,
		ForwardingUnitGas:		DefaultFeeMarketForwardingUnitGas,
		SurchargeStepBps:		DefaultFeeMarketSurchargeStepBps,
		MaxSenderSurchargeBps:		DefaultFeeMarketMaxSurchargeBps,
		ValidatorRewardBps:		6_000,
		CommunityPoolBps:		500,
		BurnBps:			2_000,
		StateReserveBps:		1_000,
		SecurityReserveBps:		500,
	}
}

func OptimizeFeeMarket(input FeeMarketOptimizerInput) (FeeMarketOptimizerOutput, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return FeeMarketOptimizerOutput{}, err
	}
	if err := input.Validate(params); err != nil {
		return FeeMarketOptimizerOutput{}, err
	}

	previousBaseFee := normalizeInt(input.CurrentBaseFeeNaet)
	if previousBaseFee.IsZero() {
		previousBaseFee = normalizeInt(params.DefaultBaseFeeNaet)
	}
	utilizationBps := feeMarketBlockUtilizationBps(input.BlockGasUsed, input.BlockGasLimit)
	nextBaseFee, deltaBps := updateFeeMarketBaseFee(previousBaseFee, utilizationBps, input.MempoolPressureBps, input.FailedExecutionRateBps, input.RecentBaseFeesNaet, params)
	multipliers := feeResourceMultipliers(input, params)
	requiredWithoutSurcharge := requiredResourceFee(input, nextBaseFee, multipliers, params)
	surcharge := senderLocalSurcharge(input.SenderFailedTxCount, requiredWithoutSurcharge, params)
	required := requiredWithoutSurcharge.Add(surcharge.SurchargeNaet)
	offered := normalizeInt(input.OfferedFeeNaet)
	validation := FeeValidationResult{
		RequiredFeeNaet:		required,
		OfferedFeeNaet:			offered,
		MempoolAccepted:		offered.GTE(required),
		ExecutionAccepted:		offered.GTE(required),
		MempoolExecutionAligned:	true,
	}
	collected := normalizeInt(input.CollectedFeesNaet)
	if collected.IsZero() {
		collected = required
	}
	allocation := allocateFeeBuckets(collected, input, params)
	estimate := FeeEstimateData{
		BaseFeeNaet:		nextBaseFee,
		RequiredFeeNaet:	required,
		ConservativeFeeNaet:	ApplyBps(required, BasisPoints+params.MaxBaseFeeAdjustmentBps),
		ResourceMultipliers:	multipliers,
		SenderSurchargeBps:	surcharge.SurchargeBps,
	}
	events := feeMarketEvents(input, utilizationBps, deltaBps, nextBaseFee, multipliers, surcharge, allocation, params)
	return FeeMarketOptimizerOutput{
		BaseFeeNaet:		nextBaseFee,
		PreviousBaseFeeNaet:	previousBaseFee,
		AppliedDeltaBps:	deltaBps,
		BlockUtilizationBps:	utilizationBps,
		ResourceMultipliers:	multipliers,
		SenderSurcharge:	surcharge,
		Validation:		validation,
		Allocation:		allocation,
		Estimate:		estimate,
		CongestionEvents:	events,
	}, nil
}

func SimulateFeeMarket(params FeeMarketOptimizerParams, steps []FeeMarketSimulationStep) (FeeMarketSimulationReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return FeeMarketSimulationReport{}, err
	}
	if len(steps) == 0 {
		return FeeMarketSimulationReport{}, fmt.Errorf("simulation steps are required")
	}
	outputs := make([]FeeMarketOptimizerOutput, 0, len(steps))
	failed := make([]string, 0)
	minBase := normalizeInt(params.MaxBaseFeeNaet)
	maxBase := normalizeInt(params.MinBaseFeeNaet)
	current := sdkmath.ZeroInt()
	for i, step := range steps {
		if step.Name == "" {
			return FeeMarketSimulationReport{}, fmt.Errorf("step name is required")
		}
		input := step.Input
		input.Params = params
		if i > 0 && normalizeInt(input.CurrentBaseFeeNaet).IsZero() {
			input.CurrentBaseFeeNaet = current
		}
		out, err := OptimizeFeeMarket(input)
		if err != nil {
			return FeeMarketSimulationReport{}, err
		}
		outputs = append(outputs, out)
		current = out.BaseFeeNaet
		if out.BaseFeeNaet.LT(minBase) {
			minBase = out.BaseFeeNaet
		}
		if out.BaseFeeNaet.GT(maxBase) {
			maxBase = out.BaseFeeNaet
		}
		if !out.Validation.MempoolExecutionAligned {
			failed = append(failed, step.Name+":mempool_execution_fee_mismatch")
		}
		if !out.Allocation.SumsExactly {
			failed = append(failed, step.Name+":fee_allocation_mismatch")
		}
		if absInt64(out.AppliedDeltaBps) > params.MaxBaseFeeAdjustmentBps {
			failed = append(failed, step.Name+":base_fee_adjustment_exceeded")
		}
	}
	return FeeMarketSimulationReport{Steps: outputs, MinBaseFeeNaet: minBase, MaxBaseFeeNaet: maxBase, Passed: len(failed) == 0, Failed: failed}, nil
}

func (p FeeMarketOptimizerParams) Validate() error {
	for _, field := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "min_base_fee_naet", value: p.MinBaseFeeNaet},
		{name: "max_base_fee_naet", value: p.MaxBaseFeeNaet},
		{name: "default_base_fee_naet", value: p.DefaultBaseFeeNaet},
		{name: "epoch_burn_cap_naet", value: p.EpochBurnCapNaet},
		{name: "epoch_state_reserve_cap_naet", value: p.EpochStateReserveCapNaet},
		{name: "epoch_security_reserve_cap_naet", value: p.EpochSecurityReserveCapNaet},
	} {
		if normalizeInt(field.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	if normalizeInt(p.MinBaseFeeNaet).GT(normalizeInt(p.MaxBaseFeeNaet)) {
		return fmt.Errorf("min_base_fee_naet must be <= max_base_fee_naet")
	}
	if normalizeInt(p.DefaultBaseFeeNaet).LT(normalizeInt(p.MinBaseFeeNaet)) || normalizeInt(p.DefaultBaseFeeNaet).GT(normalizeInt(p.MaxBaseFeeNaet)) {
		return fmt.Errorf("default_base_fee_naet must be within min/max")
	}
	if err := validateBps("target_block_utilization_bps", p.TargetBlockUtilizationBps, 1, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_base_fee_adjustment_bps", p.MaxBaseFeeAdjustmentBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.SmoothingWindow == 0 {
		return fmt.Errorf("smoothing_window must be positive")
	}
	if err := validateBps("min_resource_multiplier_bps", p.MinResourceMultiplierBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("max_resource_multiplier_bps", p.MaxResourceMultiplierBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if p.MinResourceMultiplierBps > p.MaxResourceMultiplierBps {
		return fmt.Errorf("min_resource_multiplier_bps must be <= max_resource_multiplier_bps")
	}
	if p.StateWriteThresholdBytes < 0 || p.DeploymentThresholdBytes < 0 {
		return fmt.Errorf("resource thresholds must not be negative")
	}
	if p.ForwardingUnitGas == 0 {
		return fmt.Errorf("forwarding_unit_gas must be positive")
	}
	if err := validateBps("surcharge_step_bps", p.SurchargeStepBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_sender_surcharge_bps", p.MaxSenderSurchargeBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	total := p.ValidatorRewardBps + p.CommunityPoolBps + p.BurnBps + p.StateReserveBps + p.SecurityReserveBps
	if total != BasisPoints {
		return fmt.Errorf("fee allocation bps must sum to 100%%")
	}
	return nil
}

func (p FeeMarketOptimizerParams) withDefaults() FeeMarketOptimizerParams {
	defaults := DefaultFeeMarketOptimizerParams()
	if p.MinBaseFeeNaet.IsNil() {
		p.MinBaseFeeNaet = defaults.MinBaseFeeNaet
	}
	if p.MaxBaseFeeNaet.IsNil() {
		p.MaxBaseFeeNaet = defaults.MaxBaseFeeNaet
	}
	if p.DefaultBaseFeeNaet.IsNil() {
		p.DefaultBaseFeeNaet = defaults.DefaultBaseFeeNaet
	}
	if p.TargetBlockUtilizationBps == 0 {
		p.TargetBlockUtilizationBps = defaults.TargetBlockUtilizationBps
	}
	if p.MaxBaseFeeAdjustmentBps == 0 {
		p.MaxBaseFeeAdjustmentBps = defaults.MaxBaseFeeAdjustmentBps
	}
	if p.SmoothingWindow == 0 {
		p.SmoothingWindow = defaults.SmoothingWindow
	}
	if p.MinResourceMultiplierBps == 0 {
		p.MinResourceMultiplierBps = defaults.MinResourceMultiplierBps
	}
	if p.MaxResourceMultiplierBps == 0 {
		p.MaxResourceMultiplierBps = defaults.MaxResourceMultiplierBps
	}
	if p.StateWriteThresholdBytes == 0 {
		p.StateWriteThresholdBytes = defaults.StateWriteThresholdBytes
	}
	if p.DeploymentThresholdBytes == 0 {
		p.DeploymentThresholdBytes = defaults.DeploymentThresholdBytes
	}
	if p.ForwardingThresholdMessages == 0 {
		p.ForwardingThresholdMessages = defaults.ForwardingThresholdMessages
	}
	if p.ForwardingUnitGas == 0 {
		p.ForwardingUnitGas = defaults.ForwardingUnitGas
	}
	if p.SurchargeStepBps == 0 {
		p.SurchargeStepBps = defaults.SurchargeStepBps
	}
	if p.MaxSenderSurchargeBps == 0 {
		p.MaxSenderSurchargeBps = defaults.MaxSenderSurchargeBps
	}
	if p.ValidatorRewardBps == 0 && p.CommunityPoolBps == 0 && p.BurnBps == 0 && p.StateReserveBps == 0 && p.SecurityReserveBps == 0 {
		p.ValidatorRewardBps = defaults.ValidatorRewardBps
		p.CommunityPoolBps = defaults.CommunityPoolBps
		p.BurnBps = defaults.BurnBps
		p.StateReserveBps = defaults.StateReserveBps
		p.SecurityReserveBps = defaults.SecurityReserveBps
	}
	return p
}

func (input FeeMarketOptimizerInput) Validate(params FeeMarketOptimizerParams) error {
	if input.EpochID == 0 {
		return fmt.Errorf("epoch_id must be positive")
	}
	if input.BlockGasLimit == 0 {
		return fmt.Errorf("block_gas_limit must be positive")
	}
	if input.BlockGasUsed > input.BlockGasLimit {
		return fmt.Errorf("block_gas_used must not exceed block_gas_limit")
	}
	for _, field := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "current_base_fee_naet", value: input.CurrentBaseFeeNaet},
		{name: "offered_fee_naet", value: input.OfferedFeeNaet},
		{name: "collected_fees_naet", value: input.CollectedFeesNaet},
		{name: "existing_epoch_burn_naet", value: input.ExistingEpochBurnNaet},
		{name: "existing_epoch_state_reserve", value: input.ExistingEpochStateReserve},
		{name: "existing_epoch_security_reserve", value: input.ExistingEpochSecurityReserve},
	} {
		if normalizeInt(field.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	for i, recent := range input.RecentBaseFeesNaet {
		if normalizeInt(recent).IsNegative() {
			return fmt.Errorf("recent_base_fees_naet[%d] must not be negative", i)
		}
	}
	if err := validateBps("mempool_pressure_bps", input.MempoolPressureBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("failed_execution_rate_bps", input.FailedExecutionRateBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	if input.StateWriteBytes < 0 || input.DeploymentBytes < 0 {
		return fmt.Errorf("state/deployment volumes must not be negative")
	}
	return nil
}

func updateFeeMarketBaseFee(current sdkmath.Int, utilizationBps, mempoolPressureBps, failedRateBps int64, recent []sdkmath.Int, params FeeMarketOptimizerParams) (sdkmath.Int, int64) {
	utilizationPressure := utilizationBps - params.TargetBlockUtilizationBps
	pressure := utilizationPressure + mempoolPressureBps/2 + failedRateBps/2
	deltaBps := clampInt64(pressure*params.MaxBaseFeeAdjustmentBps/BasisPoints, -params.MaxBaseFeeAdjustmentBps, params.MaxBaseFeeAdjustmentBps)
	raw := normalizeInt(current).Add(ApplyBps(normalizeInt(current), deltaBps))
	if raw.LT(normalizeInt(params.MinBaseFeeNaet)) {
		raw = normalizeInt(params.MinBaseFeeNaet)
	}
	if raw.GT(normalizeInt(params.MaxBaseFeeNaet)) {
		raw = normalizeInt(params.MaxBaseFeeNaet)
	}
	smoothed := smoothBaseFee(raw, recent, params.SmoothingWindow)
	if smoothed.LT(normalizeInt(params.MinBaseFeeNaet)) {
		smoothed = normalizeInt(params.MinBaseFeeNaet)
	}
	if smoothed.GT(normalizeInt(params.MaxBaseFeeNaet)) {
		smoothed = normalizeInt(params.MaxBaseFeeNaet)
	}
	applied := int64(0)
	if normalizeInt(current).IsPositive() {
		applied = smoothed.Sub(normalizeInt(current)).MulRaw(BasisPoints).Quo(normalizeInt(current)).Int64()
		applied = clampInt64(applied, -params.MaxBaseFeeAdjustmentBps, params.MaxBaseFeeAdjustmentBps)
	}
	return smoothed, applied
}

func feeResourceMultipliers(input FeeMarketOptimizerInput, params FeeMarketOptimizerParams) FeeResourceMultipliers {
	computePressure := feeMarketBlockUtilizationBps(input.BlockGasUsed, input.BlockGasLimit) - params.TargetBlockUtilizationBps
	compute := resourceMultiplierFromPressure(computePressure+input.MempoolPressureBps/2, params)
	storage := resourceMultiplierFromVolume(input.StateWriteBytes, params.StateWriteThresholdBytes, params)
	deployment := resourceMultiplierFromVolume(input.DeploymentBytes, params.DeploymentThresholdBytes, params)
	forwarding := resourceMultiplierFromCount(input.ForwardingMessages, params.ForwardingThresholdMessages, params)
	return FeeResourceMultipliers{ComputeBps: compute, StorageBps: storage, DeploymentBps: deployment, ForwardingBps: forwarding}
}

func requiredResourceFee(input FeeMarketOptimizerInput, baseFee sdkmath.Int, multipliers FeeResourceMultipliers, params FeeMarketOptimizerParams) sdkmath.Int {
	compute := ApplyBps(normalizeInt(baseFee).MulRaw(int64(input.TxGasUsed)), multipliers.ComputeBps)
	storage := ApplyBps(normalizeInt(baseFee).MulRaw(maxInt64(input.StateWriteBytes, 0)), multipliers.StorageBps)
	deployment := ApplyBps(normalizeInt(baseFee).MulRaw(maxInt64(input.DeploymentBytes, 0)), multipliers.DeploymentBps)
	forwardingGas := int64(input.ForwardingMessages * params.ForwardingUnitGas)
	forwarding := ApplyBps(normalizeInt(baseFee).MulRaw(forwardingGas), multipliers.ForwardingBps)
	return compute.Add(storage).Add(deployment).Add(forwarding)
}

func senderLocalSurcharge(failedTxCount uint64, required sdkmath.Int, params FeeMarketOptimizerParams) SenderLocalSurcharge {
	surchargeBps := clampInt64(int64(failedTxCount)*params.SurchargeStepBps, 0, params.MaxSenderSurchargeBps)
	return SenderLocalSurcharge{
		FailedTxCount:	failedTxCount,
		SurchargeBps:	surchargeBps,
		SurchargeNaet:	ApplyBps(normalizeInt(required), surchargeBps),
	}
}

func allocateFeeBuckets(collected sdkmath.Int, input FeeMarketOptimizerInput, params FeeMarketOptimizerParams) FeeAllocationBuckets {
	fees := normalizeInt(collected)
	community := ApplyBps(fees, params.CommunityPoolBps)
	burn := cappedBucket(ApplyBps(fees, params.BurnBps), normalizeInt(input.ExistingEpochBurnNaet), normalizeInt(params.EpochBurnCapNaet))
	state := cappedBucket(ApplyBps(fees, params.StateReserveBps), normalizeInt(input.ExistingEpochStateReserve), normalizeInt(params.EpochStateReserveCapNaet))
	security := cappedBucket(ApplyBps(fees, params.SecurityReserveBps), normalizeInt(input.ExistingEpochSecurityReserve), normalizeInt(params.EpochSecurityReserveCapNaet))
	validator := fees.Sub(community).Sub(burn).Sub(state).Sub(security)
	if validator.IsNegative() {
		validator = sdkmath.ZeroInt()
	}
	total := validator.Add(community).Add(burn).Add(state).Add(security)
	return FeeAllocationBuckets{
		CollectedFeesNaet:		fees,
		ValidatorDelegatorNaet:		validator,
		CommunityPoolNaet:		community,
		BurnNaet:			burn,
		StateReserveNaet:		state,
		SecurityReserveNaet:		security,
		BurnCapRemainderNaet:		capRemainder(normalizeInt(input.ExistingEpochBurnNaet).Add(burn), normalizeInt(params.EpochBurnCapNaet)),
		StateCapRemainderNaet:		capRemainder(normalizeInt(input.ExistingEpochStateReserve).Add(state), normalizeInt(params.EpochStateReserveCapNaet)),
		SecurityCapRemainderNaet:	capRemainder(normalizeInt(input.ExistingEpochSecurityReserve).Add(security), normalizeInt(params.EpochSecurityReserveCapNaet)),
		SumsExactly:			total.Equal(fees),
	}
}

func feeMarketEvents(input FeeMarketOptimizerInput, utilizationBps, deltaBps int64, baseFee sdkmath.Int, multipliers FeeResourceMultipliers, surcharge SenderLocalSurcharge, allocation FeeAllocationBuckets, params FeeMarketOptimizerParams) []FeeMarketCongestionEvent {
	events := []FeeMarketCongestionEvent{{
		Type:		FeeMarketEventBaseFeeUpdated,
		EpochID:	input.EpochID,
		BlockHeight:	input.BlockHeight,
		Reason:		"base_fee_formula_applied",
		ValueBps:	deltaBps,
		AmountNaet:	baseFee,
	}}
	if utilizationBps > params.TargetBlockUtilizationBps || input.MempoolPressureBps > 0 || input.FailedExecutionRateBps > 0 {
		events = append(events, FeeMarketCongestionEvent{Type: FeeMarketEventCongestionDetected, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Reason: "congestion_signal_nonzero", ValueBps: maxInt64(utilizationBps, maxInt64(input.MempoolPressureBps, input.FailedExecutionRateBps))})
	}
	if surcharge.SurchargeBps > 0 {
		events = append(events, FeeMarketCongestionEvent{Type: FeeMarketEventSenderSurcharge, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Reason: "sender_failed_tx_history", ValueBps: surcharge.SurchargeBps, AmountNaet: surcharge.SurchargeNaet})
	}
	if allocation.SumsExactly {
		events = append(events, FeeMarketCongestionEvent{Type: FeeMarketEventAllocationAccounted, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Reason: "fee_buckets_sum_to_collected", AmountNaet: allocation.CollectedFeesNaet})
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].BlockHeight == events[j].BlockHeight {
			return events[i].Type < events[j].Type
		}
		return events[i].BlockHeight < events[j].BlockHeight
	})
	_ = multipliers
	return events
}

func feeMarketBlockUtilizationBps(used, limit uint64) int64 {
	if limit == 0 {
		return 0
	}
	return int64(used * uint64(BasisPoints) / limit)
}

func resourceMultiplierFromPressure(pressureBps int64, params FeeMarketOptimizerParams) int64 {
	if pressureBps <= 0 {
		return params.MinResourceMultiplierBps
	}
	return clampInt64(BasisPoints+pressureBps, params.MinResourceMultiplierBps, params.MaxResourceMultiplierBps)
}

func resourceMultiplierFromVolume(volume, threshold int64, params FeeMarketOptimizerParams) int64 {
	if threshold <= 0 || volume <= threshold {
		return params.MinResourceMultiplierBps
	}
	over := (volume - threshold) * BasisPoints / threshold
	return clampInt64(BasisPoints+over, params.MinResourceMultiplierBps, params.MaxResourceMultiplierBps)
}

func resourceMultiplierFromCount(count, threshold uint64, params FeeMarketOptimizerParams) int64 {
	if threshold == 0 || count <= threshold {
		return params.MinResourceMultiplierBps
	}
	over := int64((count - threshold) * uint64(BasisPoints) / threshold)
	return clampInt64(BasisPoints+over, params.MinResourceMultiplierBps, params.MaxResourceMultiplierBps)
}

func smoothBaseFee(raw sdkmath.Int, recent []sdkmath.Int, window uint32) sdkmath.Int {
	if window == 0 || len(recent) == 0 {
		return normalizeInt(raw)
	}
	limit := int(window)
	if limit > len(recent) {
		limit = len(recent)
	}
	total := normalizeInt(raw)
	count := int64(1)
	for i := len(recent) - limit; i < len(recent); i++ {
		total = total.Add(normalizeInt(recent[i]))
		count++
	}
	return total.QuoRaw(count)
}

func cappedBucket(proposed, existing, cap sdkmath.Int) sdkmath.Int {
	proposed = normalizeInt(proposed)
	if normalizeInt(cap).IsZero() {
		return proposed
	}
	remaining := normalizeInt(cap).Sub(normalizeInt(existing))
	if remaining.IsNegative() {
		return sdkmath.ZeroInt()
	}
	return minInt(proposed, remaining)
}

func capRemainder(used, cap sdkmath.Int) sdkmath.Int {
	if normalizeInt(cap).IsZero() {
		return sdkmath.ZeroInt()
	}
	remaining := normalizeInt(cap).Sub(normalizeInt(used))
	if remaining.IsNegative() {
		return sdkmath.ZeroInt()
	}
	return remaining
}
