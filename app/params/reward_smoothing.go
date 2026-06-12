package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	RewardSmoothingEventEpochDistributed	= "reward_smoothing_epoch_distributed"
	RewardSmoothingEventBoundApplied	= "reward_smoothing_bound_applied"

	DefaultRewardSmoothingMaxEpochChangeBps	= int64(2_000)
	DefaultRewardSmoothingEpochLengthBlocks	= uint64(10_000)
)

type RewardSmoothingParams struct {
	MaxRewardChangeBps	int64
	EpochLengthBlocks	uint64
}

type DelegatorRewardParticipant struct {
	DelegatorID	string
	StakeNaet	sdkmath.Int
}

type ValidatorRewardParticipant struct {
	ValidatorID	string
	VotingPower	uint64
	CommissionBps	int64
	DelegatorStake	[]DelegatorRewardParticipant
}

type RewardSmoothingInput struct {
	EpochID				uint64
	GrossRewardsNaet		sdkmath.Int
	PreviousEpochRewardsNaet	sdkmath.Int
	Validators			[]ValidatorRewardParticipant
	Params				RewardSmoothingParams
}

type DelegatorRewardAllocation struct {
	DelegatorID	string
	RewardNaet	sdkmath.Int
}

type ValidatorRewardAllocation struct {
	ValidatorID		string
	GrossRewardNaet		sdkmath.Int
	CommissionNaet		sdkmath.Int
	DelegatorRewardPoolNaet	sdkmath.Int
	DelegatorRewards	[]DelegatorRewardAllocation
}

type RewardSmoothingState struct {
	EpochID			uint64
	EpochLengthBlocks	uint64
	TotalRewardsNaet	sdkmath.Int
	ValidatorRewards	[]ValidatorRewardAllocation
}

type RewardSmoothingEvent struct {
	Type			string
	EpochID			uint64
	GrossRewardsNaet	sdkmath.Int
	SmoothedRewardsNaet	sdkmath.Int
	BoundApplied		bool
}

type RewardSmoothingOutput struct {
	EpochID			uint64
	GrossRewardsNaet	sdkmath.Int
	SmoothedRewardsNaet	sdkmath.Int
	BoundApplied		bool
	TotalVotingPower	uint64
	ValidatorRewards	[]ValidatorRewardAllocation
	State			RewardSmoothingState
	Events			[]RewardSmoothingEvent
}

func DefaultRewardSmoothingParams() RewardSmoothingParams {
	return RewardSmoothingParams{
		MaxRewardChangeBps:	DefaultRewardSmoothingMaxEpochChangeBps,
		EpochLengthBlocks:	DefaultRewardSmoothingEpochLengthBlocks,
	}
}

func SmoothEpochRewards(input RewardSmoothingInput) (RewardSmoothingOutput, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return RewardSmoothingOutput{}, err
	}
	if input.EpochID == 0 {
		return RewardSmoothingOutput{}, fmt.Errorf("epoch_id must be positive")
	}
	grossRewards := normalizeInt(input.GrossRewardsNaet)
	previousRewards := normalizeInt(input.PreviousEpochRewardsNaet)
	if grossRewards.IsNegative() {
		return RewardSmoothingOutput{}, fmt.Errorf("gross_rewards_naet must not be negative")
	}
	if previousRewards.IsNegative() {
		return RewardSmoothingOutput{}, fmt.Errorf("previous_epoch_rewards_naet must not be negative")
	}
	validators := normalizeRewardParticipants(input.Validators)
	if len(validators) == 0 {
		return RewardSmoothingOutput{}, fmt.Errorf("validators are required")
	}
	totalPower := uint64(0)
	for _, validator := range validators {
		if err := validator.Validate(); err != nil {
			return RewardSmoothingOutput{}, err
		}
		if totalPower > ^uint64(0)-validator.VotingPower {
			return RewardSmoothingOutput{}, fmt.Errorf("total voting power overflow")
		}
		totalPower += validator.VotingPower
	}
	if totalPower == 0 {
		return RewardSmoothingOutput{}, fmt.Errorf("total voting power must be positive")
	}
	if totalPower > uint64(^uint64(0)>>1) {
		return RewardSmoothingOutput{}, fmt.Errorf("total voting power exceeds deterministic int64 bounds")
	}

	smoothed, boundApplied := boundEpochRewards(grossRewards, previousRewards, params.MaxRewardChangeBps)
	validatorRewards := allocateValidatorRewards(smoothed, validators, totalPower)
	state := RewardSmoothingState{
		EpochID:		input.EpochID,
		EpochLengthBlocks:	params.EpochLengthBlocks,
		TotalRewardsNaet:	smoothed,
		ValidatorRewards:	validatorRewards,
	}
	if err := state.Validate(); err != nil {
		return RewardSmoothingOutput{}, err
	}
	events := []RewardSmoothingEvent{{
		Type:			RewardSmoothingEventEpochDistributed,
		EpochID:		input.EpochID,
		GrossRewardsNaet:	grossRewards,
		SmoothedRewardsNaet:	smoothed,
		BoundApplied:		boundApplied,
	}}
	if boundApplied {
		events = append(events, RewardSmoothingEvent{
			Type:			RewardSmoothingEventBoundApplied,
			EpochID:		input.EpochID,
			GrossRewardsNaet:	grossRewards,
			SmoothedRewardsNaet:	smoothed,
			BoundApplied:		true,
		})
	}
	return RewardSmoothingOutput{
		EpochID:		input.EpochID,
		GrossRewardsNaet:	grossRewards,
		SmoothedRewardsNaet:	smoothed,
		BoundApplied:		boundApplied,
		TotalVotingPower:	totalPower,
		ValidatorRewards:	validatorRewards,
		State:			state,
		Events:			events,
	}, nil
}

func (p RewardSmoothingParams) Validate() error {
	if err := validateBps("max_reward_change_bps", p.MaxRewardChangeBps, 1, BasisPoints); err != nil {
		return err
	}
	if p.EpochLengthBlocks == 0 {
		return fmt.Errorf("epoch_length_blocks must be positive")
	}
	return nil
}

func (p RewardSmoothingParams) withDefaults() RewardSmoothingParams {
	defaults := DefaultRewardSmoothingParams()
	if p.MaxRewardChangeBps == 0 {
		p.MaxRewardChangeBps = defaults.MaxRewardChangeBps
	}
	if p.EpochLengthBlocks == 0 {
		p.EpochLengthBlocks = defaults.EpochLengthBlocks
	}
	return p
}

func (p ValidatorRewardParticipant) Validate() error {
	if p.ValidatorID == "" {
		return fmt.Errorf("validator_id must be non-empty")
	}
	if p.VotingPower == 0 {
		return fmt.Errorf("validator voting power must be positive")
	}
	if err := ValidateCommissionBounds(p.CommissionBps, 0); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, delegator := range p.DelegatorStake {
		if delegator.DelegatorID == "" {
			return fmt.Errorf("delegator_id must be non-empty")
		}
		if _, found := seen[delegator.DelegatorID]; found {
			return fmt.Errorf("duplicate delegator %s", delegator.DelegatorID)
		}
		seen[delegator.DelegatorID] = struct{}{}
		if normalizeInt(delegator.StakeNaet).IsNegative() {
			return fmt.Errorf("delegator stake must not be negative")
		}
	}
	return nil
}

func (s RewardSmoothingState) Validate() error {
	if s.EpochID == 0 {
		return fmt.Errorf("epoch_id must be positive")
	}
	if s.EpochLengthBlocks == 0 {
		return fmt.Errorf("epoch_length_blocks must be positive")
	}
	if normalizeInt(s.TotalRewardsNaet).IsNegative() {
		return fmt.Errorf("total_rewards_naet must not be negative")
	}
	if !validatorRewardAllocationsSorted(s.ValidatorRewards) {
		return fmt.Errorf("validator rewards must be sorted by validator_id")
	}
	total := sdkmath.ZeroInt()
	seen := map[string]struct{}{}
	for _, allocation := range s.ValidatorRewards {
		if allocation.ValidatorID == "" {
			return fmt.Errorf("validator_id must be non-empty")
		}
		if _, found := seen[allocation.ValidatorID]; found {
			return fmt.Errorf("duplicate validator %s", allocation.ValidatorID)
		}
		seen[allocation.ValidatorID] = struct{}{}
		if !delegatorRewardAllocationsSorted(allocation.DelegatorRewards) {
			return fmt.Errorf("delegator rewards must be sorted by delegator_id")
		}
		validatorTotal := normalizeInt(allocation.CommissionNaet).Add(normalizeInt(allocation.DelegatorRewardPoolNaet))
		if !validatorTotal.Equal(normalizeInt(allocation.GrossRewardNaet)) {
			return fmt.Errorf("validator reward allocation must conserve rewards")
		}
		delegatorTotal := sdkmath.ZeroInt()
		delegatorSeen := map[string]struct{}{}
		for _, delegator := range allocation.DelegatorRewards {
			if delegator.DelegatorID == "" {
				return fmt.Errorf("delegator_id must be non-empty")
			}
			if _, found := delegatorSeen[delegator.DelegatorID]; found {
				return fmt.Errorf("duplicate delegator %s", delegator.DelegatorID)
			}
			delegatorSeen[delegator.DelegatorID] = struct{}{}
			if normalizeInt(delegator.RewardNaet).IsNegative() {
				return fmt.Errorf("delegator reward must not be negative")
			}
			delegatorTotal = delegatorTotal.Add(normalizeInt(delegator.RewardNaet))
		}
		if !delegatorTotal.Equal(normalizeInt(allocation.DelegatorRewardPoolNaet)) {
			return fmt.Errorf("delegator reward allocation must conserve rewards")
		}
		total = total.Add(normalizeInt(allocation.GrossRewardNaet))
	}
	if !total.Equal(normalizeInt(s.TotalRewardsNaet)) {
		return fmt.Errorf("epoch reward state must conserve total rewards")
	}
	return nil
}

func boundEpochRewards(gross, previous sdkmath.Int, maxChangeBps int64) (sdkmath.Int, bool) {
	if previous.IsZero() {
		return gross, false
	}
	maxDelta := previous.MulRaw(maxChangeBps).QuoRaw(BasisPoints)
	upper := previous.Add(maxDelta)
	lower := previous.Sub(maxDelta)
	if lower.IsNegative() {
		lower = sdkmath.ZeroInt()
	}
	if gross.GT(upper) {
		return upper, true
	}
	if gross.LT(lower) {
		return lower, true
	}
	return gross, false
}

func allocateValidatorRewards(total sdkmath.Int, validators []ValidatorRewardParticipant, totalPower uint64) []ValidatorRewardAllocation {
	out := make([]ValidatorRewardAllocation, 0, len(validators))
	allocated := sdkmath.ZeroInt()
	for i, validator := range validators {
		gross := total.MulRaw(int64(validator.VotingPower)).QuoRaw(int64(totalPower))
		if i == len(validators)-1 {
			gross = total.Sub(allocated)
		}
		allocated = allocated.Add(gross)
		commission := ApplyBps(gross, validator.CommissionBps)
		delegatorPool := gross.Sub(commission)
		out = append(out, ValidatorRewardAllocation{
			ValidatorID:			validator.ValidatorID,
			GrossRewardNaet:		gross,
			CommissionNaet:			commission,
			DelegatorRewardPoolNaet:	delegatorPool,
			DelegatorRewards:		allocateDelegatorRewards(delegatorPool, validator.DelegatorStake),
		})
	}
	return out
}

func allocateDelegatorRewards(total sdkmath.Int, delegators []DelegatorRewardParticipant) []DelegatorRewardAllocation {
	delegators = normalizeDelegatorParticipants(delegators)
	if len(delegators) == 0 {
		return []DelegatorRewardAllocation{}
	}
	totalStake := sdkmath.ZeroInt()
	for _, delegator := range delegators {
		totalStake = totalStake.Add(normalizeInt(delegator.StakeNaet))
	}
	if !totalStake.IsPositive() {
		return []DelegatorRewardAllocation{}
	}
	out := make([]DelegatorRewardAllocation, 0, len(delegators))
	allocated := sdkmath.ZeroInt()
	for i, delegator := range delegators {
		reward := total.Mul(normalizeInt(delegator.StakeNaet)).Quo(totalStake)
		if i == len(delegators)-1 {
			reward = total.Sub(allocated)
		}
		allocated = allocated.Add(reward)
		out = append(out, DelegatorRewardAllocation{DelegatorID: delegator.DelegatorID, RewardNaet: reward})
	}
	return out
}

func normalizeRewardParticipants(validators []ValidatorRewardParticipant) []ValidatorRewardParticipant {
	out := append([]ValidatorRewardParticipant(nil), validators...)
	for i := range out {
		out[i].DelegatorStake = normalizeDelegatorParticipants(out[i].DelegatorStake)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ValidatorID < out[j].ValidatorID })
	return out
}

func normalizeDelegatorParticipants(delegators []DelegatorRewardParticipant) []DelegatorRewardParticipant {
	out := append([]DelegatorRewardParticipant(nil), delegators...)
	sort.Slice(out, func(i, j int) bool { return out[i].DelegatorID < out[j].DelegatorID })
	return out
}

func validatorRewardAllocationsSorted(values []ValidatorRewardAllocation) bool {
	return sort.SliceIsSorted(values, func(i, j int) bool { return values[i].ValidatorID < values[j].ValidatorID })
}

func delegatorRewardAllocationsSorted(values []DelegatorRewardAllocation) bool {
	return sort.SliceIsSorted(values, func(i, j int) bool { return values[i].DelegatorID < values[j].DelegatorID })
}
