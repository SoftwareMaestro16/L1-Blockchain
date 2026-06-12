package types

import (
	"errors"
	"fmt"
	"sort"
)

type PoolAllocationPlanInput struct {
	PoolID			string
	Epoch			uint64
	Height			uint64
	MaxTouchedAllocations	uint32
	Weights			[]AllocationWeight
}

func BuildPoolAllocationPlan(input PoolAllocationPlanInput) (PoolRebalanceReceipt, error) {
	if err := validateID("pool allocation plan pool id", input.PoolID, MaxPoolIDBytesV1); err != nil {
		return PoolRebalanceReceipt{}, err
	}
	if input.Epoch == 0 || input.Height == 0 {
		return PoolRebalanceReceipt{}, errors.New("pool allocation plan epoch and height must be positive")
	}
	if input.MaxTouchedAllocations == 0 {
		return PoolRebalanceReceipt{}, errors.New("pool allocation plan max touched allocations must be positive")
	}
	weights := append([]AllocationWeight(nil), input.Weights...)
	sort.SliceStable(weights, func(i, j int) bool { return weights[i].ValidatorAddress < weights[j].ValidatorAddress })

	receipt := PoolRebalanceReceipt{
		PoolID:	input.PoolID,
		Epoch:	input.Epoch,
		Height:	input.Height,
	}
	for _, weight := range weights {
		if weight.WeightBps == 0 {
			continue
		}
		if uint32(len(receipt.Allocations)) >= input.MaxTouchedAllocations {
			break
		}
		allocation := PoolValidatorAllocation{
			PoolID:			input.PoolID,
			Validator:		weight.ValidatorAddress,
			TargetWeightBps:	weight.WeightBps,
			UpdatedHeight:		input.Height,
		}
		receipt.Allocations = append(receipt.Allocations, allocation)
		receipt.InternalMetadata.TouchedKeys = append(receipt.InternalMetadata.TouchedKeys, string(PoolAllocationKey(input.PoolID, weight.ValidatorAddress)))
	}
	if len(receipt.Allocations) == 0 {
		return PoolRebalanceReceipt{}, errors.New("pool allocation plan requires at least one positive allocation")
	}
	return receipt, nil
}

func ApplyPoolAllocationPlan(params Params, existing []PoolValidatorAllocation, receipt PoolRebalanceReceipt) ([]PoolValidatorAllocation, []string, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, err
	}
	if err := validateID("pool allocation transition pool id", receipt.PoolID, params.MaxPoolIDBytes); err != nil {
		return nil, nil, err
	}
	if receipt.Epoch == 0 || receipt.Height == 0 {
		return nil, nil, errors.New("pool allocation transition epoch and height must be positive")
	}
	if len(receipt.Allocations) == 0 {
		return nil, nil, errors.New("pool allocation transition requires allocations")
	}

	byKey := make(map[string]PoolValidatorAllocation, len(existing)+len(receipt.Allocations))
	for _, allocation := range existing {
		key := allocation.PoolID + "\x00" + allocation.Validator
		byKey[key] = allocation
	}
	touched := make([]string, 0, len(receipt.Allocations))
	seenPlanKeys := map[string]struct{}{}
	for _, allocation := range SortPoolValidatorAllocations(receipt.Allocations) {
		if allocation.PoolID != receipt.PoolID {
			return nil, nil, errors.New("pool allocation transition allocation pool mismatch")
		}
		if err := ValidateUserFacingAEAddress("pool allocation transition validator", allocation.Validator); err != nil {
			return nil, nil, err
		}
		if allocation.TargetWeightBps == 0 {
			return nil, nil, errors.New("pool allocation transition target weight must be positive")
		}
		if allocation.TargetWeightBps > params.MaxPoolValidatorAllocationBps || allocation.TargetWeightBps > params.ValidatorPowerCapBps {
			return nil, nil, fmt.Errorf("pool allocation transition target weight %d exceeds configured cap", allocation.TargetWeightBps)
		}
		allocation.UpdatedHeight = receipt.Height
		key := allocation.PoolID + "\x00" + allocation.Validator
		if _, found := seenPlanKeys[key]; found {
			return nil, nil, errors.New("pool allocation transition duplicate validator")
		}
		seenPlanKeys[key] = struct{}{}
		byKey[key] = allocation
		touched = append(touched, string(PoolAllocationKey(allocation.PoolID, allocation.Validator)))
	}

	next := make([]PoolValidatorAllocation, 0, len(byKey))
	for _, allocation := range byKey {
		next = append(next, allocation)
	}
	next = SortPoolValidatorAllocations(next)
	sort.Strings(touched)
	return next, touched, nil
}
