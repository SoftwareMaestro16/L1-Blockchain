package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FeeBucketValidatorDelegator	= "validator_delegator_rewards"
	FeeBucketCommunityPool		= "community_pool"
	FeeBucketBurn			= "burn"
	FeeBucketStateMaintenance	= "state_maintenance_reserve"
	FeeBucketSecurityReserve	= "security_reserve"

	AdmissionModeMempool		= "mempool"
	AdmissionModeBlockExecution	= "block_execution"
)

type FeeDistributionParams struct {
	ValidatorBaseBps		uint32
	CommunityBaseBps		uint32
	BurnBaseBps			uint32
	StateMaintenanceBaseBps		uint32
	SecurityReserveBaseBps		uint32
	ValidatorRewardFloorBps		uint32
	ActivityBurnMaxBps		uint32
	StateMaintenanceMaxBps		uint32
	SecurityReserveMaxBps		uint32
	HighActivityThresholdBps	uint32
	HighStateGrowthThresholdBps	uint32
}

type FeeAllocationBucket struct {
	Bucket	string
	Amount	sdk.Coin
	Bps	uint32
}

type FeeAllocationEvent struct {
	Type		string
	Bucket		string
	Denom		string
	Amount		sdkmath.Int
	Bps		uint32
	Attributes	[]EventAttribute
}

type EventAttribute struct {
	Key	string
	Value	string
}

type FeeDistributionResult struct {
	Collected			sdk.Coin
	ValidatorDelegatorRewards	sdk.Coin
	CommunityPool			sdk.Coin
	Burn				sdk.Coin
	StateMaintenanceReserve		sdk.Coin
	SecurityReserve			sdk.Coin
	Buckets				[]FeeAllocationBucket
	Events				[]FeeAllocationEvent
	ValidatorRewardFloorEnforced	bool
	BurnAllocationBps		uint32
	StateMaintenanceBps		uint32
	SecurityReserveBps		uint32
}

type AccountActivityWindow struct {
	Sender			string
	WindowID		uint64
	TxCount			uint32
	FailedTxCount		uint32
	StateWriteCount		uint32
	DeploymentCount		uint32
	ExecutableMessages	uint32
}

type AntiSpamParams struct {
	WindowTxSoftLimit		uint32
	FailedTxSoftLimit		uint32
	StateWriteSoftLimit		uint32
	DeploymentSoftLimit		uint32
	AccountActivityScoreMaxBps	uint32
	FailedTxSurchargeStepBps	uint32
	StateWriteSurchargeStepBps	uint32
	DeploymentSurchargeStepBps	uint32
	MaxTotalSurchargeBps		uint32
	MinExecutableFeeNaet		sdkmath.Int
}

type AccountActivityScore struct {
	Sender			string
	WindowID		uint64
	ScoreBps		uint32
	FailedTxScoreBps	uint32
	StateWriteScoreBps	uint32
	DeploymentScoreBps	uint32
	Bounded			bool
}

type AntiSpamFeeInput struct {
	Mode		string
	BaseRequiredFee	sdk.Coin
	PaidFee		sdk.Coin
	Activity	AccountActivityWindow
	Params		AntiSpamParams
	Executable	bool
}

type AntiSpamAdmissionDecision struct {
	Mode			string
	Sender			string
	RequiredFee		sdk.Coin
	PaidFee			sdk.Coin
	ActivityScoreBps	uint32
	SurchargeBps		uint32
	Accepted		bool
	Reason			string
	Deterministic		bool
	Bounded			bool
}

func DefaultFeeDistributionParams() FeeDistributionParams {
	return FeeDistributionParams{
		ValidatorBaseBps:		8_000,
		CommunityBaseBps:		1_000,
		BurnBaseBps:			500,
		StateMaintenanceBaseBps:	300,
		SecurityReserveBaseBps:		200,
		ValidatorRewardFloorBps:	6_000,
		ActivityBurnMaxBps:		1_500,
		StateMaintenanceMaxBps:		1_500,
		SecurityReserveMaxBps:		500,
		HighActivityThresholdBps:	7_000,
		HighStateGrowthThresholdBps:	6_000,
	}
}

func (p FeeDistributionParams) Validate() error {
	baseTotal := uint64(p.ValidatorBaseBps) + uint64(p.CommunityBaseBps) + uint64(p.BurnBaseBps) +
		uint64(p.StateMaintenanceBaseBps) + uint64(p.SecurityReserveBaseBps)
	if baseTotal != BasisPoints {
		return fmt.Errorf("fee distribution base buckets must sum to 10000 bps")
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "validator reward floor", value: p.ValidatorRewardFloorBps},
		{name: "activity burn max", value: p.ActivityBurnMaxBps},
		{name: "state maintenance max", value: p.StateMaintenanceMaxBps},
		{name: "security reserve max", value: p.SecurityReserveMaxBps},
		{name: "high activity threshold", value: p.HighActivityThresholdBps},
		{name: "high state growth threshold", value: p.HighStateGrowthThresholdBps},
	} {
		if item.value > uint32(BasisPoints) {
			return fmt.Errorf("%s must be <= 10000 bps", item.name)
		}
	}
	if p.ValidatorRewardFloorBps > p.ValidatorBaseBps {
		return errors.New("validator reward floor cannot exceed validator base allocation")
	}
	if p.HighActivityThresholdBps >= uint32(BasisPoints) || p.HighStateGrowthThresholdBps >= uint32(BasisPoints) {
		return errors.New("activity and state growth thresholds must be below 10000 bps")
	}
	return nil
}

func DistributeFeeBuckets(collected sdk.Coin, params FeeDistributionParams, signals CongestionSignals, stateGrowthBps uint32) (FeeDistributionResult, error) {
	if params.ValidatorBaseBps == 0 {
		params = DefaultFeeDistributionParams()
	}
	if err := params.Validate(); err != nil {
		return FeeDistributionResult{}, err
	}
	if err := signals.Validate(); err != nil {
		return FeeDistributionResult{}, err
	}
	if stateGrowthBps > uint32(BasisPoints) {
		return FeeDistributionResult{}, fmt.Errorf("state growth must be <= 10000 bps")
	}
	if collected.Denom != BondDenom || !collected.IsValid() || !collected.IsPositive() {
		return FeeDistributionResult{}, fmt.Errorf("collected fee must be positive %s", BondDenom)
	}

	validatorBps := params.ValidatorBaseBps
	communityBps := params.CommunityBaseBps
	burnBps := params.BurnBaseBps
	stateBps := params.StateMaintenanceBaseBps
	securityBps := params.SecurityReserveBaseBps
	floorEnforced := false

	for _, request := range []struct {
		target	*uint32
		add	uint32
	}{
		{target: &burnBps, add: activityDependentBps(signals.BlockGasUtilizationBps, params.HighActivityThresholdBps, params.ActivityBurnMaxBps)},
		{target: &stateBps, add: activityDependentBps(stateGrowthBps, params.HighStateGrowthThresholdBps, params.StateMaintenanceMaxBps)},
		{target: &securityBps, add: activityDependentBps(maxUint32(signals.MempoolPressureBps, signals.FailedExecutionRateBps), params.HighActivityThresholdBps, params.SecurityReserveMaxBps)},
	} {
		if request.add == 0 {
			continue
		}
		available := uint32(0)
		if validatorBps > params.ValidatorRewardFloorBps {
			available = validatorBps - params.ValidatorRewardFloorBps
		}
		applied := minUint32Fee(request.add, available)
		*request.target += applied
		validatorBps -= applied
		if applied < request.add {
			floorEnforced = true
		}
	}

	allocations := []FeeAllocationBucket{
		{Bucket: FeeBucketCommunityPool, Bps: communityBps},
		{Bucket: FeeBucketBurn, Bps: burnBps},
		{Bucket: FeeBucketStateMaintenance, Bps: stateBps},
		{Bucket: FeeBucketSecurityReserve, Bps: securityBps},
	}
	allocated := sdkmath.ZeroInt()
	for i := range allocations {
		amount := collected.Amount.MulRaw(int64(allocations[i].Bps)).QuoRaw(int64(BasisPoints))
		allocations[i].Amount = sdk.NewCoin(collected.Denom, amount)
		allocated = allocated.Add(amount)
	}
	validatorAmount := collected.Amount.Sub(allocated)
	validator := FeeAllocationBucket{
		Bucket:	FeeBucketValidatorDelegator,
		Amount:	sdk.NewCoin(collected.Denom, validatorAmount),
		Bps:	validatorBps,
	}
	allocations = append([]FeeAllocationBucket{validator}, allocations...)
	sort.SliceStable(allocations, func(i, j int) bool {
		return allocations[i].Bucket < allocations[j].Bucket
	})

	result := FeeDistributionResult{
		Collected:			collected,
		Buckets:			allocations,
		ValidatorRewardFloorEnforced:	floorEnforced || validatorBps == params.ValidatorRewardFloorBps,
		BurnAllocationBps:		burnBps,
		StateMaintenanceBps:		stateBps,
		SecurityReserveBps:		securityBps,
	}
	for _, bucket := range allocations {
		switch bucket.Bucket {
		case FeeBucketValidatorDelegator:
			result.ValidatorDelegatorRewards = bucket.Amount
		case FeeBucketCommunityPool:
			result.CommunityPool = bucket.Amount
		case FeeBucketBurn:
			result.Burn = bucket.Amount
		case FeeBucketStateMaintenance:
			result.StateMaintenanceReserve = bucket.Amount
		case FeeBucketSecurityReserve:
			result.SecurityReserve = bucket.Amount
		}
		result.Events = append(result.Events, FeeAllocationEvent{
			Type:	"fee_allocation",
			Bucket:	bucket.Bucket,
			Denom:	bucket.Amount.Denom,
			Amount:	bucket.Amount.Amount,
			Bps:	bucket.Bps,
			Attributes: []EventAttribute{
				{Key: "activity_bps", Value: fmt.Sprintf("%d", signals.BlockGasUtilizationBps)},
				{Key: "state_growth_bps", Value: fmt.Sprintf("%d", stateGrowthBps)},
			},
		})
	}
	if !sumFeeBuckets(result.Buckets).Equal(collected.Amount) {
		return FeeDistributionResult{}, errors.New("fee distribution allocation mismatch")
	}
	return result, nil
}

func DefaultAntiSpamParams(params Params) AntiSpamParams {
	params = NormalizeParams(params)
	minFee, err := params.MinFeeInt()
	if err != nil {
		minFee = sdkmath.NewInt(1)
	}
	return AntiSpamParams{
		WindowTxSoftLimit:		uint32(params.MaxSenderTxsPerBlock),
		FailedTxSoftLimit:		3,
		StateWriteSoftLimit:		10,
		DeploymentSoftLimit:		2,
		AccountActivityScoreMaxBps:	uint32(BasisPoints),
		FailedTxSurchargeStepBps:	DefaultSpamSurchargeStepBps,
		StateWriteSurchargeStepBps:	250,
		DeploymentSurchargeStepBps:	750,
		MaxTotalSurchargeBps:		DefaultMaxSpamSurchargeBps,
		MinExecutableFeeNaet:		minFee,
	}
}

func (p AntiSpamParams) Validate() error {
	if p.WindowTxSoftLimit == 0 || p.FailedTxSoftLimit == 0 || p.StateWriteSoftLimit == 0 || p.DeploymentSoftLimit == 0 {
		return errors.New("anti-spam soft limits must be positive")
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "activity score max", value: p.AccountActivityScoreMaxBps},
		{name: "failed tx surcharge step", value: p.FailedTxSurchargeStepBps},
		{name: "state write surcharge step", value: p.StateWriteSurchargeStepBps},
		{name: "deployment surcharge step", value: p.DeploymentSurchargeStepBps},
		{name: "max total surcharge", value: p.MaxTotalSurchargeBps},
	} {
		if item.value == 0 || item.value > 100_000 {
			return fmt.Errorf("%s must be within 1..100000 bps", item.name)
		}
	}
	if !normalizeFeeInt(p.MinExecutableFeeNaet).IsPositive() {
		return errors.New("minimum executable fee must be positive")
	}
	return nil
}

func ScoreAccountActivity(window AccountActivityWindow, params AntiSpamParams) (AccountActivityScore, error) {
	if params.WindowTxSoftLimit == 0 {
		params = DefaultAntiSpamParams(DefaultParams())
	}
	if err := params.Validate(); err != nil {
		return AccountActivityScore{}, err
	}
	failed := excessScoreBps(window.FailedTxCount, params.FailedTxSoftLimit, params.AccountActivityScoreMaxBps)
	stateWrites := excessScoreBps(window.StateWriteCount, params.StateWriteSoftLimit, params.AccountActivityScoreMaxBps)
	deployments := excessScoreBps(window.DeploymentCount, params.DeploymentSoftLimit, params.AccountActivityScoreMaxBps)
	total := uint64(excessScoreBps(window.TxCount, params.WindowTxSoftLimit, params.AccountActivityScoreMaxBps)) +
		uint64(failed) + uint64(stateWrites) + uint64(deployments)
	score := uint32(total)
	bounded := false
	if score > params.AccountActivityScoreMaxBps {
		score = params.AccountActivityScoreMaxBps
		bounded = true
	}
	return AccountActivityScore{
		Sender:			window.Sender,
		WindowID:		window.WindowID,
		ScoreBps:		score,
		FailedTxScoreBps:	failed,
		StateWriteScoreBps:	stateWrites,
		DeploymentScoreBps:	deployments,
		Bounded:		bounded,
	}, nil
}

func AntiSpamSurchargeBps(window AccountActivityWindow, params AntiSpamParams) (uint32, AccountActivityScore, error) {
	score, err := ScoreAccountActivity(window, params)
	if err != nil {
		return 0, AccountActivityScore{}, err
	}
	failedExcess := excessCount(window.FailedTxCount, params.FailedTxSoftLimit)
	stateExcess := excessCount(window.StateWriteCount, params.StateWriteSoftLimit)
	deploymentExcess := excessCount(window.DeploymentCount, params.DeploymentSoftLimit)
	surcharge := uint64(failedExcess)*uint64(params.FailedTxSurchargeStepBps) +
		uint64(stateExcess)*uint64(params.StateWriteSurchargeStepBps) +
		uint64(deploymentExcess)*uint64(params.DeploymentSurchargeStepBps)
	if surcharge > uint64(params.MaxTotalSurchargeBps) {
		surcharge = uint64(params.MaxTotalSurchargeBps)
	}
	return uint32(surcharge), score, nil
}

func ValidateAntiSpamFee(input AntiSpamFeeInput) (AntiSpamAdmissionDecision, error) {
	mode := input.Mode
	if mode == "" {
		mode = AdmissionModeMempool
	}
	if mode != AdmissionModeMempool && mode != AdmissionModeBlockExecution {
		return AntiSpamAdmissionDecision{}, fmt.Errorf("unsupported admission mode %q", mode)
	}
	params := input.Params
	if params.WindowTxSoftLimit == 0 {
		params = DefaultAntiSpamParams(DefaultParams())
	}
	if err := params.Validate(); err != nil {
		return AntiSpamAdmissionDecision{}, err
	}
	if input.BaseRequiredFee.Denom != BondDenom || !input.BaseRequiredFee.IsValid() || input.BaseRequiredFee.IsNegative() {
		return AntiSpamAdmissionDecision{}, fmt.Errorf("base required fee must use %s", BondDenom)
	}
	if input.PaidFee.Denom != BondDenom || !input.PaidFee.IsValid() || input.PaidFee.IsNegative() {
		return AntiSpamAdmissionDecision{}, fmt.Errorf("paid fee must use %s", BondDenom)
	}
	base := input.BaseRequiredFee.Amount
	if input.Executable && base.LT(params.MinExecutableFeeNaet) {
		base = params.MinExecutableFeeNaet
	}
	surchargeBps, score, err := AntiSpamSurchargeBps(input.Activity, params)
	if err != nil {
		return AntiSpamAdmissionDecision{}, err
	}
	required := base.MulRaw(int64(BasisPoints + uint64(surchargeBps))).QuoRaw(int64(BasisPoints))
	if required.LT(base) {
		required = base
	}
	decision := AntiSpamAdmissionDecision{
		Mode:			mode,
		Sender:			input.Activity.Sender,
		RequiredFee:		sdk.NewCoin(BondDenom, required),
		PaidFee:		input.PaidFee,
		ActivityScoreBps:	score.ScoreBps,
		SurchargeBps:		surchargeBps,
		Accepted:		input.PaidFee.Amount.GTE(required),
		Deterministic:		true,
		Bounded:		surchargeBps <= params.MaxTotalSurchargeBps,
	}
	if !decision.Accepted {
		decision.Reason = "fee_below_anti_spam_requirement"
	}
	return decision, nil
}

func activityDependentBps(value, threshold, maxAdd uint32) uint32 {
	if value <= threshold || maxAdd == 0 {
		return 0
	}
	denom := uint64(BasisPoints) - uint64(threshold)
	if denom == 0 {
		return maxAdd
	}
	add := uint32(uint64(value-threshold) * uint64(maxAdd) / denom)
	if add > maxAdd {
		return maxAdd
	}
	return add
}

func sumFeeBuckets(buckets []FeeAllocationBucket) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, bucket := range buckets {
		total = total.Add(bucket.Amount.Amount)
	}
	return total
}

func excessScoreBps(value, softLimit, maxBps uint32) uint32 {
	if value <= softLimit {
		return 0
	}
	score := uint64(value-softLimit) * uint64(BasisPoints) / uint64(softLimit)
	if score > uint64(maxBps) {
		score = uint64(maxBps)
	}
	return uint32(score)
}

func excessCount(value, softLimit uint32) uint32 {
	if value <= softLimit {
		return 0
	}
	return value - softLimit
}

func maxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func minUint32Fee(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
