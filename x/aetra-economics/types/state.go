package types

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
)

type Params struct {
	Authority			string	`json:"authority"`
	InflationMinBps			uint32	`json:"inflation_min_bps"`
	InflationMaxBps			uint32	`json:"inflation_max_bps"`
	InflationChangeRateBps		uint32	`json:"inflation_change_rate_bps"`
	TargetBondedRatioBps		uint32	`json:"target_bonded_ratio_bps"`
	BurnMinBps			uint32	`json:"burn_min_bps"`
	BurnMaxBps			uint32	`json:"burn_max_bps"`
	BurnCurrentBps			uint32	`json:"burn_current_bps"`
	ValidatorRewardMinBps		uint32	`json:"validator_reward_min_bps"`
	ValidatorRewardMaxBps		uint32	`json:"validator_reward_max_bps"`
	ValidatorRewardBps		uint32	`json:"validator_reward_bps"`
	TreasuryMinBps			uint32	`json:"treasury_min_bps"`
	TreasuryMaxBps			uint32	`json:"treasury_max_bps"`
	TreasuryBps			uint32	`json:"treasury_bps"`
	EmergencyAllowZeroRewardShare	bool	`json:"emergency_allow_zero_reward_share"`
	RewardSmoothingWindow		uint64	`json:"reward_smoothing_window"`
	APRTargetMinBps			uint32	`json:"apr_target_min_bps"`
	APRTargetMaxBps			uint32	`json:"apr_target_max_bps"`
	EpochsPerYear			uint64	`json:"epochs_per_year"`
}

type EconomicsState struct {
	CurrentInflationBps	uint32			`json:"current_inflation_bps"`
	CurrentBondedRatioBps	uint32			`json:"current_bonded_ratio_bps"`
	EstimatedAPRBps		uint32			`json:"estimated_apr_bps"`
	TotalSupply		uint64			`json:"total_supply"`
	BurnedSupply		uint64			`json:"burned_supply"`
	TreasuryBalance		uint64			`json:"treasury_balance"`
	RewardHistory		[]EpochRewardSummary	`json:"reward_history"`
}

type EpochEconomicsInput struct {
	Epoch		uint64	`json:"epoch"`
	TotalSupply	uint64	`json:"total_supply"`
	BondedTokens	uint64	`json:"bonded_tokens"`
	FeesCollected	uint64	`json:"fees_collected"`
}

type FeeSplit struct {
	FeesCollected			uint64	`json:"fees_collected"`
	BurnAmount			uint64	`json:"burn_amount"`
	TreasuryAmount			uint64	`json:"treasury_amount"`
	ValidatorDelegatorRewards	uint64	`json:"validator_delegator_rewards"`
	BurnBps				uint32	`json:"burn_bps"`
	TreasuryBps			uint32	`json:"treasury_bps"`
	ValidatorRewardBps		uint32	`json:"validator_reward_bps"`
}

type EpochRewardSummary struct {
	Epoch				uint64	`json:"epoch"`
	StartingSupply			uint64	`json:"starting_supply"`
	EndingSupply			uint64	`json:"ending_supply"`
	BondedTokens			uint64	`json:"bonded_tokens"`
	BondedRatioBps			uint32	`json:"bonded_ratio_bps"`
	InflationBps			uint32	`json:"inflation_bps"`
	EstimatedAPRBps			uint32	`json:"estimated_apr_bps"`
	FeesCollected			uint64	`json:"fees_collected"`
	BurnedAmount			uint64	`json:"burned_amount"`
	TreasuryAmount			uint64	`json:"treasury_amount"`
	ValidatorDelegatorRewards	uint64	`json:"validator_delegator_rewards"`
	MintedRewards			uint64	`json:"minted_rewards"`
	GrossRewards			uint64	`json:"gross_rewards"`
	SmoothedRewards			uint64	`json:"smoothed_rewards"`
	NetSupplyChange			int64	`json:"net_supply_change"`
	BurnedSupply			uint64	`json:"burned_supply"`
	TreasuryBalance			uint64	`json:"treasury_balance"`
}

type GenesisState struct {
	Params	Params		`json:"params"`
	State	EconomicsState	`json:"state"`
}

type MsgUpdateEconomicsParams struct {
	Authority	string	`json:"authority"`
	Params		Params	`json:"params"`
}

type MsgApplyEpochEconomics struct {
	Authority	string			`json:"authority"`
	Input		EpochEconomicsInput	`json:"input"`
}

type QueryCurrentInflationRequest struct{}
type QueryCurrentInflationResponse struct{ InflationBps uint32 }

type QueryCurrentBondedRatioRequest struct{}
type QueryCurrentBondedRatioResponse struct{ BondedRatioBps uint32 }

type QueryEstimatedAPRRequest struct {
	ValidatorCommissionBps		uint32	`json:"validator_commission_bps"`
	ValidatorOperatingCostBps	uint32	`json:"validator_operating_cost_bps"`
}

type QueryEstimatedAPRResponse struct {
	IsEstimate			bool	`json:"is_estimate"`
	EstimateLabel			string	`json:"estimate_label"`
	InflationOnlyAPRBps		uint32	`json:"inflation_only_apr_bps"`
	FeeAdjustedAPRBps		uint32	`json:"fee_adjusted_apr_bps"`
	ValidatorCommissionImpactBps	uint32	`json:"validator_commission_impact_bps"`
	EstimatedDelegatorAPRBps	uint32	`json:"estimated_delegator_apr_bps"`
	EstimatedValidatorGrossAPRBps	uint32	`json:"estimated_validator_gross_apr_bps"`
	EstimatedValidatorNetAPRBps	uint32	`json:"estimated_validator_net_apr_bps"`
}

type QueryFeeSplitParamsRequest struct{}
type QueryFeeSplitParamsResponse struct {
	BurnMinBps			uint32	`json:"burn_min_bps"`
	BurnMaxBps			uint32	`json:"burn_max_bps"`
	BurnCurrentBps			uint32	`json:"burn_current_bps"`
	ValidatorRewardMinBps		uint32	`json:"validator_reward_min_bps"`
	ValidatorRewardMaxBps		uint32	`json:"validator_reward_max_bps"`
	ValidatorRewardBps		uint32	`json:"validator_reward_bps"`
	TreasuryMinBps			uint32	`json:"treasury_min_bps"`
	TreasuryMaxBps			uint32	`json:"treasury_max_bps"`
	TreasuryBps			uint32	`json:"treasury_bps"`
	EmergencyAllowZeroRewardShare	bool	`json:"emergency_allow_zero_reward_share"`
}

type QueryBurnedSupplyRequest struct{}
type QueryBurnedSupplyResponse struct{ BurnedSupply uint64 }

type QueryTreasuryBalanceRequest struct{}
type QueryTreasuryBalanceResponse struct{ TreasuryBalance uint64 }

type QueryEpochRewardSummaryRequest struct{ Epoch uint64 }
type QueryEpochRewardSummaryResponse struct{ Summary EpochRewardSummary }

func DefaultParams(authority string) Params {
	return Params{
		Authority:		authority,
		InflationMinBps:	200,
		InflationMaxBps:	500,
		InflationChangeRateBps:	25,
		TargetBondedRatioBps:	6_000,
		BurnMinBps:		3_000,
		BurnMaxBps:		6_000,
		BurnCurrentBps:		5_000,
		ValidatorRewardMinBps:	2_000,
		ValidatorRewardMaxBps:	4_000,
		ValidatorRewardBps:	3_500,
		TreasuryMinBps:		1_000,
		TreasuryMaxBps:		2_000,
		TreasuryBps:		1_500,
		RewardSmoothingWindow:	7,
		APRTargetMinBps:	500,
		APRTargetMaxBps:	800,
		EpochsPerYear:		6_307_200,
	}
}

func DefaultGenesisState(authority string) GenesisState {
	params := DefaultParams(authority)
	return GenesisState{
		Params:	params,
		State: EconomicsState{
			CurrentInflationBps:	midpointBps(params.InflationMinBps, params.InflationMaxBps),
			CurrentBondedRatioBps:	params.TargetBondedRatioBps,
			EstimatedAPRBps:	EstimateAPRBps(midpointBps(params.InflationMinBps, params.InflationMaxBps), params.TargetBondedRatioBps),
			RewardHistory:		[]EpochRewardSummary{},
		},
	}
}

func ComputeInflationBps(params Params, bondedRatioBps uint32) uint32 {
	midpoint := midpointBps(params.InflationMinBps, params.InflationMaxBps)
	if bondedRatioBps == params.TargetBondedRatioBps {
		return midpoint
	}
	if bondedRatioBps < params.TargetBondedRatioBps {
		if params.TargetBondedRatioBps == 0 {
			return params.InflationMaxBps
		}
		gap := uint64(params.TargetBondedRatioBps - bondedRatioBps)
		headroom := uint64(params.InflationMaxBps - midpoint)
		return clampBps(uint32(uint64(midpoint)+headroom*gap/uint64(params.TargetBondedRatioBps)), params.InflationMinBps, params.InflationMaxBps)
	}
	upperRange := BasisPoints - params.TargetBondedRatioBps
	if upperRange == 0 {
		return params.InflationMinBps
	}
	gap := uint64(bondedRatioBps - params.TargetBondedRatioBps)
	room := uint64(midpoint - params.InflationMinBps)
	return clampBps(uint32(uint64(midpoint)-room*gap/uint64(upperRange)), params.InflationMinBps, params.InflationMaxBps)
}

func ComputeNextInflationBps(params Params, currentInflationBps, bondedRatioBps uint32) uint32 {
	target := ComputeInflationBps(params, bondedRatioBps)
	if target == currentInflationBps {
		return currentInflationBps
	}
	if target > currentInflationBps {
		delta := target - currentInflationBps
		if delta > params.InflationChangeRateBps {
			delta = params.InflationChangeRateBps
		}
		return clampBps(currentInflationBps+delta, params.InflationMinBps, params.InflationMaxBps)
	}
	delta := currentInflationBps - target
	if delta > params.InflationChangeRateBps {
		delta = params.InflationChangeRateBps
	}
	return clampBps(currentInflationBps-delta, params.InflationMinBps, params.InflationMaxBps)
}

func EstimateAPRBps(inflationBps, bondedRatioBps uint32) uint32 {
	if bondedRatioBps == 0 {
		return 0
	}
	return uint32((uint64(inflationBps)*uint64(BasisPoints) + uint64(bondedRatioBps)/2) / uint64(bondedRatioBps))
}

func EstimateAPRBreakdown(params Params, state EconomicsState, req QueryEstimatedAPRRequest) (QueryEstimatedAPRResponse, error) {
	if err := params.Validate(); err != nil {
		return QueryEstimatedAPRResponse{}, ErrInvalidParams.Wrap(err.Error())
	}
	if req.ValidatorCommissionBps > BasisPoints {
		return QueryEstimatedAPRResponse{}, ErrInvalidParams.Wrap("validator commission cannot exceed basis points")
	}
	if req.ValidatorOperatingCostBps > BasisPoints {
		return QueryEstimatedAPRResponse{}, ErrInvalidParams.Wrap("validator operating cost cannot exceed basis points")
	}

	inflationOnly := EstimateAPRBps(state.CurrentInflationBps, state.CurrentBondedRatioBps)
	feeAdjusted := inflationOnly
	if len(state.RewardHistory) > 0 {
		latest := state.RewardHistory[len(state.RewardHistory)-1]
		if latest.BondedTokens > 0 {
			feeAdjusted = annualizedRewardAPRBps(latest.GrossRewards, params.EpochsPerYear, latest.BondedTokens)
		}
	}
	commissionImpact := applyBpsUint32(feeAdjusted, req.ValidatorCommissionBps)
	delegatorAPR := subtractUint32Floor(feeAdjusted, commissionImpact)
	validatorGrossAPR, err := checkedAddUint32(feeAdjusted, commissionImpact)
	if err != nil {
		return QueryEstimatedAPRResponse{}, err
	}
	validatorNetAPR := subtractUint32Floor(validatorGrossAPR, req.ValidatorOperatingCostBps)

	return QueryEstimatedAPRResponse{
		IsEstimate:			true,
		EstimateLabel:			"estimate_not_guaranteed_return",
		InflationOnlyAPRBps:		inflationOnly,
		FeeAdjustedAPRBps:		feeAdjusted,
		ValidatorCommissionImpactBps:	commissionImpact,
		EstimatedDelegatorAPRBps:	delegatorAPR,
		EstimatedValidatorGrossAPRBps:	validatorGrossAPR,
		EstimatedValidatorNetAPRBps:	validatorNetAPR,
	}, nil
}

func ComputeFeeSplit(params Params, fees uint64) (FeeSplit, error) {
	if err := params.Validate(); err != nil {
		return FeeSplit{}, ErrInvalidParams.Wrap(err.Error())
	}
	burn, err := mulDivUint64(fees, uint64(params.BurnCurrentBps), uint64(BasisPoints))
	if err != nil {
		return FeeSplit{}, err
	}
	treasury, err := mulDivUint64(fees, uint64(params.TreasuryBps), uint64(BasisPoints))
	if err != nil {
		return FeeSplit{}, err
	}
	feeAllocated, err := checkedAddUint64(burn, treasury)
	if err != nil || feeAllocated > fees {
		return FeeSplit{}, ErrInvalidState.Wrap("fee split exceeds collected fees")
	}
	validatorRewards := fees - burn - treasury
	return FeeSplit{
		FeesCollected:			fees,
		BurnAmount:			burn,
		TreasuryAmount:			treasury,
		ValidatorDelegatorRewards:	validatorRewards,
		BurnBps:			params.BurnCurrentBps,
		TreasuryBps:			params.TreasuryBps,
		ValidatorRewardBps:		params.ValidatorRewardBps,
	}, nil
}

func ApplyEpoch(params Params, state EconomicsState, input EpochEconomicsInput) (EconomicsState, EpochRewardSummary, error) {
	if err := params.Validate(); err != nil {
		return EconomicsState{}, EpochRewardSummary{}, ErrInvalidParams.Wrap(err.Error())
	}
	if err := input.Validate(); err != nil {
		return EconomicsState{}, EpochRewardSummary{}, ErrInvalidState.Wrap(err.Error())
	}
	bondedRatio := ratioBps(input.BondedTokens, input.TotalSupply)
	inflation := ComputeNextInflationBps(params, state.CurrentInflationBps, bondedRatio)
	estimatedAPR := EstimateAPRBps(inflation, bondedRatio)
	annualMint, err := mulDivUint64(input.TotalSupply, uint64(inflation), uint64(BasisPoints))
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	minted, err := mulDivUint64(annualMint, 1, params.EpochsPerYear)
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	split, err := ComputeFeeSplit(params, input.FeesCollected)
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	grossRewards, err := checkedAddUint64(minted, split.ValidatorDelegatorRewards)
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	smoothed := SmoothReward(grossRewards, state.RewardHistory, params.RewardSmoothingWindow)
	increasedSupply, err := checkedAddUint64(input.TotalSupply, minted)
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	if split.BurnAmount > increasedSupply {
		return EconomicsState{}, EpochRewardSummary{}, ErrInvalidState.Wrap("burn amount exceeds post-mint supply")
	}
	endingSupply := increasedSupply - split.BurnAmount
	netSupplyChange, err := supplyDeltaInt64(minted, split.BurnAmount)
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	burnedSupply, err := checkedAddUint64(state.BurnedSupply, split.BurnAmount)
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	treasuryBalance, err := checkedAddUint64(state.TreasuryBalance, split.TreasuryAmount)
	if err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	summary := EpochRewardSummary{
		Epoch:				input.Epoch,
		StartingSupply:			input.TotalSupply,
		EndingSupply:			endingSupply,
		BondedTokens:			input.BondedTokens,
		BondedRatioBps:			bondedRatio,
		InflationBps:			inflation,
		EstimatedAPRBps:		estimatedAPR,
		FeesCollected:			input.FeesCollected,
		BurnedAmount:			split.BurnAmount,
		TreasuryAmount:			split.TreasuryAmount,
		ValidatorDelegatorRewards:	split.ValidatorDelegatorRewards,
		MintedRewards:			minted,
		GrossRewards:			grossRewards,
		SmoothedRewards:		smoothed,
		NetSupplyChange:		netSupplyChange,
		BurnedSupply:			burnedSupply,
		TreasuryBalance:		treasuryBalance,
	}
	if err := summary.Validate(params); err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	next := state
	next.CurrentInflationBps = inflation
	next.CurrentBondedRatioBps = bondedRatio
	next.EstimatedAPRBps = estimatedAPR
	next.TotalSupply = endingSupply
	next.BurnedSupply = burnedSupply
	next.TreasuryBalance = treasuryBalance
	next.RewardHistory = appendRewardSummary(next.RewardHistory, summary, params.RewardSmoothingWindow)
	if err := next.Validate(params); err != nil {
		return EconomicsState{}, EpochRewardSummary{}, err
	}
	return next, summary, nil
}

func SmoothReward(current uint64, history []EpochRewardSummary, window uint64) uint64 {
	if window <= 1 || len(history) == 0 {
		return current
	}
	sum := current
	count := uint64(1)
	for i := len(history) - 1; i >= 0 && count < window; i-- {
		next, err := checkedAddUint64(sum, history[i].GrossRewards)
		if err != nil {
			return current
		}
		sum = next
		count++
		if i == 0 {
			break
		}
	}
	return sum / count
}

func (p Params) Validate() error {
	if strings.TrimSpace(p.Authority) == "" {
		return errors.New("authority must be non-empty")
	}
	if p.InflationMinBps > p.InflationMaxBps || p.InflationMaxBps > BasisPoints {
		return fmt.Errorf("inflation bounds are invalid")
	}
	if p.InflationChangeRateBps == 0 || p.InflationChangeRateBps > BasisPoints {
		return fmt.Errorf("inflation change rate must be between 1 and %d bps", BasisPoints)
	}
	if p.TargetBondedRatioBps == 0 || p.TargetBondedRatioBps >= BasisPoints {
		return fmt.Errorf("target bonded ratio must be between 1 and %d bps", BasisPoints-1)
	}
	if p.BurnMinBps > p.BurnMaxBps || p.BurnMaxBps > BasisPoints {
		return fmt.Errorf("burn bounds are invalid")
	}
	if p.BurnCurrentBps < p.BurnMinBps || p.BurnCurrentBps > p.BurnMaxBps {
		return fmt.Errorf("current burn percentage must stay inside burn bounds")
	}
	if p.ValidatorRewardMinBps > p.ValidatorRewardMaxBps || p.ValidatorRewardMaxBps > BasisPoints {
		return fmt.Errorf("validator reward bounds are invalid")
	}
	if p.ValidatorRewardBps == 0 {
		if !p.EmergencyAllowZeroRewardShare {
			return fmt.Errorf("validator reward share cannot be zero without emergency governance")
		}
	} else if p.ValidatorRewardBps < p.ValidatorRewardMinBps || p.ValidatorRewardBps > p.ValidatorRewardMaxBps {
		return fmt.Errorf("validator reward percentage must stay inside reward bounds")
	}
	if p.TreasuryMinBps > p.TreasuryMaxBps || p.TreasuryMaxBps > BasisPoints {
		return fmt.Errorf("treasury bounds are invalid")
	}
	if p.TreasuryBps < p.TreasuryMinBps || p.TreasuryBps > p.TreasuryMaxBps {
		return fmt.Errorf("treasury percentage must stay inside treasury bounds")
	}
	if uint64(p.BurnCurrentBps)+uint64(p.ValidatorRewardBps)+uint64(p.TreasuryBps) != uint64(BasisPoints) {
		return fmt.Errorf("fee split percentages must sum to %d bps", BasisPoints)
	}
	if p.RewardSmoothingWindow == 0 {
		return errors.New("reward smoothing window must be positive")
	}
	if p.APRTargetMinBps > p.APRTargetMaxBps || p.APRTargetMaxBps > BasisPoints {
		return fmt.Errorf("apr target bounds are invalid")
	}
	if p.EpochsPerYear == 0 {
		return errors.New("epochs per year must be positive")
	}
	return nil
}

func (s EconomicsState) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if s.CurrentInflationBps < params.InflationMinBps || s.CurrentInflationBps > params.InflationMaxBps {
		return errors.New("current inflation outside configured bounds")
	}
	if s.CurrentBondedRatioBps > BasisPoints {
		return errors.New("current bonded ratio exceeds basis points")
	}
	if s.EstimatedAPRBps > BasisPoints {
		return errors.New("estimated apr exceeds basis points")
	}
	for i, summary := range s.RewardHistory {
		if err := summary.Validate(params); err != nil {
			return fmt.Errorf("reward summary %d: %w", i, err)
		}
		if i > 0 && s.RewardHistory[i-1].Epoch >= summary.Epoch {
			return errors.New("reward history must be sorted by increasing epoch")
		}
	}
	return nil
}

func (i EpochEconomicsInput) Validate() error {
	if i.Epoch == 0 {
		return errors.New("epoch must be positive")
	}
	if i.TotalSupply == 0 {
		return errors.New("total supply must be positive")
	}
	if i.BondedTokens > i.TotalSupply {
		return errors.New("bonded tokens cannot exceed total supply")
	}
	return nil
}

func (s EpochRewardSummary) Validate(params Params) error {
	if s.Epoch == 0 {
		return errors.New("epoch must be positive")
	}
	if s.StartingSupply == 0 || s.EndingSupply == 0 {
		return errors.New("supply fields must be positive")
	}
	if s.BondedTokens > s.StartingSupply {
		return errors.New("bonded tokens cannot exceed starting supply")
	}
	if s.BondedRatioBps > BasisPoints || s.InflationBps > BasisPoints || s.EstimatedAPRBps > BasisPoints {
		return errors.New("summary bps field exceeds basis points")
	}
	feeAllocated, err := checkedAddUint64(s.BurnedAmount, s.TreasuryAmount)
	if err != nil || feeAllocated > s.FeesCollected {
		return errors.New("fee split exceeds fees collected")
	}
	if s.ValidatorDelegatorRewards != s.FeesCollected-s.BurnedAmount-s.TreasuryAmount {
		return errors.New("validator rewards do not reconcile to fee split")
	}
	grossRewards, err := checkedAddUint64(s.MintedRewards, s.ValidatorDelegatorRewards)
	if err != nil || s.GrossRewards != grossRewards {
		return errors.New("gross rewards must equal minted plus validator fee rewards")
	}
	expectedEnding, err := checkedAddUint64(s.StartingSupply, s.MintedRewards)
	if err != nil || s.BurnedAmount > expectedEnding {
		return errors.New("ending supply arithmetic overflow")
	}
	expectedEnding -= s.BurnedAmount
	if s.EndingSupply != expectedEnding {
		return errors.New("ending supply must equal starting supply plus minted rewards minus burned amount")
	}
	if s.InflationBps < params.InflationMinBps || s.InflationBps > params.InflationMaxBps {
		return errors.New("summary inflation outside bounds")
	}
	delta, err := supplyDeltaInt64(s.MintedRewards, s.BurnedAmount)
	if err != nil || s.NetSupplyChange != delta {
		return errors.New("net supply change must equal minted rewards minus burned amount")
	}
	return nil
}

func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}
	return g.State.Validate(g.Params)
}

func appendRewardSummary(history []EpochRewardSummary, summary EpochRewardSummary, window uint64) []EpochRewardSummary {
	next := append([]EpochRewardSummary(nil), history...)
	next = append(next, summary)
	sort.Slice(next, func(i, j int) bool {
		return next[i].Epoch < next[j].Epoch
	})
	if window > 0 && uint64(len(next)) > window {
		next = next[len(next)-int(window):]
	}
	return next
}

func midpointBps(minBps, maxBps uint32) uint32 {
	return uint32((uint64(minBps) + uint64(maxBps)) / 2)
}

func clampBps(value, minBps, maxBps uint32) uint32 {
	if value < minBps {
		return minBps
	}
	if value > maxBps {
		return maxBps
	}
	return value
}

func ratioBps(numerator, denominator uint64) uint32 {
	if denominator == 0 {
		return 0
	}
	if numerator >= denominator {
		return BasisPoints
	}
	product := new(big.Int).Mul(new(big.Int).SetUint64(numerator), new(big.Int).SetUint64(uint64(BasisPoints)))
	product.Add(product, new(big.Int).SetUint64(denominator/2))
	product.Quo(product, new(big.Int).SetUint64(denominator))
	if product.Cmp(new(big.Int).SetUint64(uint64(BasisPoints))) > 0 {
		return BasisPoints
	}
	return uint32(product.Uint64())
}

func mulDivUint64(value, multiplier, denominator uint64) (uint64, error) {
	if denominator == 0 {
		return 0, ErrInvalidState.Wrap("division by zero")
	}
	product := new(big.Int).Mul(new(big.Int).SetUint64(value), new(big.Int).SetUint64(multiplier))
	product.Quo(product, new(big.Int).SetUint64(denominator))
	if !product.IsUint64() {
		return 0, ErrInvalidState.Wrap("uint64 accounting overflow")
	}
	return product.Uint64(), nil
}

func checkedAddUint64(a, b uint64) (uint64, error) {
	sum := a + b
	if sum < a {
		return 0, ErrInvalidState.Wrap("uint64 accounting overflow")
	}
	return sum, nil
}

func checkedAddUint32(a, b uint32) (uint32, error) {
	sum := uint64(a) + uint64(b)
	if sum > uint64(^uint32(0)) {
		return 0, ErrInvalidState.Wrap("uint32 accounting overflow")
	}
	return uint32(sum), nil
}

func subtractUint32Floor(value, decrement uint32) uint32 {
	if decrement >= value {
		return 0
	}
	return value - decrement
}

func applyBpsUint32(value, bps uint32) uint32 {
	return uint32((uint64(value)*uint64(bps) + uint64(BasisPoints)/2) / uint64(BasisPoints))
}

func annualizedRewardAPRBps(epochRewards, epochsPerYear, bondedTokens uint64) uint32 {
	if epochRewards == 0 || epochsPerYear == 0 || bondedTokens == 0 {
		return 0
	}
	numerator := new(big.Int).Mul(new(big.Int).SetUint64(epochRewards), new(big.Int).SetUint64(epochsPerYear))
	numerator.Mul(numerator, new(big.Int).SetUint64(uint64(BasisPoints)))
	numerator.Add(numerator, new(big.Int).SetUint64(bondedTokens/2))
	numerator.Quo(numerator, new(big.Int).SetUint64(bondedTokens))
	max := new(big.Int).SetUint64(uint64(^uint32(0)))
	if numerator.Cmp(max) > 0 {
		return ^uint32(0)
	}
	return uint32(numerator.Uint64())
}

func supplyDeltaInt64(minted, burned uint64) (int64, error) {
	if minted >= burned {
		delta := minted - burned
		if delta > uint64(math.MaxInt64) {
			return 0, ErrInvalidState.Wrap("net supply increase exceeds int64")
		}
		return int64(delta), nil
	}
	delta := burned - minted
	if delta > uint64(math.MaxInt64) {
		return 0, ErrInvalidState.Wrap("net supply decrease exceeds int64")
	}
	return -int64(delta), nil
}
